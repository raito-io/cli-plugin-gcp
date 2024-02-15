package bigquery

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/raito-io/bexpression"
	"github.com/raito-io/bexpression/base"
	"github.com/raito-io/bexpression/datacomparison"
	"github.com/raito-io/cli/base/access_provider/sync_from_target"
	"github.com/raito-io/cli/base/access_provider/sync_to_target"
	ds "github.com/raito-io/cli/base/data_source"
	"github.com/raito-io/cli/base/wrappers"
	"github.com/raito-io/golang-set/set"
	"google.golang.org/api/bigquery/v2"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

//go:generate go run github.com/vektra/mockery/v2 --name=filteringRepository --with-expecter --inpackage
type filteringRepository interface {
	ListFilters(ctx context.Context, table *org.GcpOrgEntity, fn func(ctx context.Context, rap *bigquery.RowAccessPolicy, users []string, groups []string, internalizable bool) error) error
	CreateOrUpdateFilter(ctx context.Context, filter *BQFilter) error
	DeleteFilter(ctx context.Context, table *BQReferencedTable, filterName string) error
}

//go:generate go run github.com/vektra/mockery/v2 --name=filteringDataObjectIterator --with-expecter --inpackage
type filteringDataObjectIterator interface {
	Sync(ctx context.Context, config *ds.DataSourceSyncConfig, skipColumns bool, fn func(ctx context.Context, object *org.GcpOrgEntity) error) error
}

type BqFilteringService struct {
	filteringRepository filteringRepository
	dataObjectIterator  filteringDataObjectIterator
}

func NewBqFilteringService(filteringRepository filteringRepository, dataObjectIterator filteringDataObjectIterator) *BqFilteringService {
	return &BqFilteringService{
		filteringRepository: filteringRepository,
		dataObjectIterator:  dataObjectIterator,
	}
}

func (s *BqFilteringService) ImportFilters(ctx context.Context, config *ds.DataSourceSyncConfig, accessProviderHandler wrappers.AccessProviderHandler, raitoFilters set.Set[string]) error {
	err := s.dataObjectIterator.Sync(ctx, config, true, func(ctx context.Context, object *org.GcpOrgEntity) error {
		if object.Type != ds.Table {
			return nil
		}

		err := s.filteringRepository.ListFilters(ctx, object, func(ctx context.Context, rap *bigquery.RowAccessPolicy, users []string, groups []string, internalizable bool) error {
			externalId := fmt.Sprintf("%s.%s.%s.%s", rap.RowAccessPolicyReference.ProjectId, rap.RowAccessPolicyReference.DatasetId, rap.RowAccessPolicyReference.TableId, rap.RowAccessPolicyReference.PolicyId)

			if raitoFilters.Contains(externalId) {
				return nil
			}

			err := accessProviderHandler.AddAccessProviders(&sync_from_target.AccessProvider{
				ExternalId:        externalId,
				Name:              rap.RowAccessPolicyReference.PolicyId,
				NamingHint:        rap.RowAccessPolicyReference.PolicyId,
				Action:            sync_from_target.Filtered,
				Policy:            rap.FilterPredicate,
				NotInternalizable: !internalizable,
				What: []sync_from_target.WhatItem{
					{
						DataObject: &ds.DataObjectReference{
							FullName: fmt.Sprintf("%s.%s.%s", rap.RowAccessPolicyReference.ProjectId, rap.RowAccessPolicyReference.DatasetId, rap.RowAccessPolicyReference.TableId),
							Type:     ds.Table,
						},
					},
				},
				Who: &sync_from_target.WhoItem{
					Users:  users,
					Groups: groups,
				},
				ActualName: rap.RowAccessPolicyReference.PolicyId,
			})

			if err != nil {
				return fmt.Errorf("add access provider filter: %w", err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("list filters for %s: %w", object.Id, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("iterate dataobject for filters: %w", err)
	}

	return nil
}

func (s *BqFilteringService) ExportFilter(ctx context.Context, accessProvider *sync_to_target.AccessProvider, accessProviderFeedbackHandler wrappers.AccessProviderFeedbackHandler) (*string, error) {
	if len(accessProvider.What) == 0 { // Probably a filter on a deleted data object. We can safely ignore this.
		err := accessProviderFeedbackHandler.AddAccessProviderFeedback(sync_to_target.AccessProviderSyncFeedback{
			AccessProvider: accessProvider.Id,
		})

		if err != nil {
			return nil, fmt.Errorf("add access provider feedback: %w", err)
		}

		return nil, nil
	}

	if !(len(accessProvider.What) == 1 && accessProvider.What[0].DataObject.Type == ds.Table) {
		err := accessProviderFeedbackHandler.AddAccessProviderFeedback(sync_to_target.AccessProviderSyncFeedback{
			AccessProvider: accessProvider.Id,
			Errors:         []string{"filter what is not a single table"},
		})

		if err != nil {
			return nil, fmt.Errorf("add access provider feedback: %w", err)
		}

		return nil, nil
	}

	var actualNamePtr *string
	var externalId *string
	var err error

	if accessProvider.Delete {
		if accessProvider.ExternalId != nil {
			externalId = accessProvider.ExternalId

			externalIdSplit := strings.SplitN(*accessProvider.ExternalId, ".", 4)
			actualNamePtr = &externalIdSplit[3]

			err = s.filteringRepository.DeleteFilter(ctx, &BQReferencedTable{Project: externalIdSplit[0], Dataset: externalIdSplit[1], Table: externalIdSplit[2]}, externalIdSplit[3])
		}
	} else {
		table := strings.SplitN(accessProvider.What[0].DataObject.FullName, ".", 3)
		tableReference := BQReferencedTable{
			Project: table[0],
			Dataset: table[1],
			Table:   table[2],
		}

		actualNamePtr, externalId, err = s.createOrUpdateMask(ctx, tableReference, accessProvider)
	}

	var errors []string
	if err != nil {
		errors = append(errors, err.Error())
	}

	actualName := ""
	if actualNamePtr != nil {
		actualName = *actualNamePtr
	}

	err = accessProviderFeedbackHandler.AddAccessProviderFeedback(sync_to_target.AccessProviderSyncFeedback{
		AccessProvider: accessProvider.Id,
		ActualName:     actualName,
		ExternalId:     externalId,
		Errors:         errors,
	})
	if err != nil {
		return externalId, fmt.Errorf("add access provider feedback: %w", err)
	}

	return externalId, nil
}

func (s *BqFilteringService) createOrUpdateMask(ctx context.Context, table BQReferencedTable, ap *sync_to_target.AccessProvider) (*string, *string, error) {
	var filterExpression string

	if ap.PolicyRule != nil {
		filterExpression = *ap.PolicyRule
	} else if ap.FilterCriteria != nil {
		var err error

		filterExpression, err = createFilterExpression(ctx, ap.FilterCriteria)
		if err != nil {
			return nil, nil, fmt.Errorf("create filter expression: %w", err)
		}
	} else {
		return nil, nil, fmt.Errorf("access provider policy rule or filter criteria is required")
	}

	var filterName string

	if ap.ExternalId != nil {
		filterName = strings.SplitN(*ap.ExternalId, ".", 4)[3]
	} else {
		filterName = validSqlName(ap.NamingHint)
	}

	externalId := fmt.Sprintf("%s.%s.%s.%s", table.Project, table.Dataset, table.Table, filterName)

	filter := BQFilter{
		FilterName:       filterName,
		Table:            table,
		Users:            ap.Who.Users,
		Groups:           ap.Who.Groups,
		FilterExpression: filterExpression,
	}

	common.Logger.Info(fmt.Sprintf("create or update filter %+v", filter))

	err := s.filteringRepository.CreateOrUpdateFilter(ctx, &filter)
	if err != nil {
		return nil, nil, fmt.Errorf("create or update filter: %w", err)
	}

	return &filterName, &externalId, nil
}

func createFilterExpression(ctx context.Context, filterCriteria *bexpression.DataComparisonExpression) (string, error) {
	filterVisitor := NewFilterExpressionVisitor()

	err := filterCriteria.Accept(ctx, filterVisitor)
	if err != nil {
		return "", fmt.Errorf("building filter expression: %w", err)
	}

	return filterVisitor.GetExpression(), nil
}

var _ base.Visitor = (*FilterExpressionVisitor)(nil)

type FilterExpressionVisitor struct {
	stringBuilder strings.Builder

	binaryExpressionLevel int
}

func NewFilterExpressionVisitor() *FilterExpressionVisitor {
	return &FilterExpressionVisitor{}
}

func (f *FilterExpressionVisitor) GetExpression() string {
	return f.stringBuilder.String()
}

func (f *FilterExpressionVisitor) EnterExpressionElement(_ context.Context, element base.VisitableElement) error {
	if node, ok := element.(*bexpression.DataComparisonExpression); ok {
		if f.binaryExpressionLevel > 0 && node.Literal == nil {
			f.stringBuilder.WriteString("(")
		}

		f.binaryExpressionLevel++
	}

	return nil
}

func (f *FilterExpressionVisitor) LeaveExpressionElement(_ context.Context, element base.VisitableElement) {
	if node, ok := element.(*bexpression.DataComparisonExpression); ok {
		f.binaryExpressionLevel--
		if f.binaryExpressionLevel > 0 && node.Literal == nil {
			f.stringBuilder.WriteString(")")
		}
	}
}

func (f *FilterExpressionVisitor) Literal(_ context.Context, l interface{}) error {
	switch node := l.(type) {
	case bool:
		f.stringBuilder.WriteString(strings.ToUpper(strconv.FormatBool(node)))
	case int:
		f.stringBuilder.WriteString(fmt.Sprintf("%d", node))
	case float64:
		f.stringBuilder.WriteString(fmt.Sprintf("%f", node))
	case string:
		f.stringBuilder.WriteString(fmt.Sprintf("%q", node))
	case time.Time:
		f.stringBuilder.WriteString(fmt.Sprintf("DATETIME(%d, %d, %d, %d, %d, %d)", node.Year(), node.Month(), node.Day(), node.Hour(), node.Minute(), node.Second()))
	case datacomparison.ComparisonOperator:
		switch node {
		case datacomparison.ComparisonOperatorEqual:
			f.stringBuilder.WriteString(" = ")
		case datacomparison.ComparisonOperatorNotEqual:
			f.stringBuilder.WriteString(" != ")
		case datacomparison.ComparisonOperatorLessThan:
			f.stringBuilder.WriteString(" < ")
		case datacomparison.ComparisonOperatorLessThanOrEqual:
			f.stringBuilder.WriteString(" <= ")
		case datacomparison.ComparisonOperatorGreaterThan:
			f.stringBuilder.WriteString(" > ")
		case datacomparison.ComparisonOperatorGreaterThanOrEqual:
			f.stringBuilder.WriteString(" >= ")
		}
	case *datacomparison.Reference:
		err := f.visitReference(node)
		if err != nil {
			return err
		}
	case base.AggregatorOperator:
		switch node {
		case base.AggregatorOperatorAnd:
			f.stringBuilder.WriteString(" AND ")
		case base.AggregatorOperatorOr:
			f.stringBuilder.WriteString(" OR ")
		default:
			return fmt.Errorf("unsupported aggregator operator: %s", node)
		}
	case base.UnaryOperator:
		if node != base.UnaryOperatorNot {
			return fmt.Errorf("unsupported unary operator: %s", node)
		}

		f.stringBuilder.WriteString("NOT ")
	}

	return nil
}

func (f *FilterExpressionVisitor) visitReference(ref *datacomparison.Reference) error {
	switch ref.EntityType {
	case datacomparison.EntityTypeDataObject:
		var object ds.DataObjectReference

		err := json.Unmarshal([]byte(ref.EntityID), &object)
		if err != nil {
			return fmt.Errorf("unmarshal reference entity id: %w", err)
		}

		if object.Type != ds.Column {
			return fmt.Errorf("unsupported reference entity type: %s", object.Type)
		}

		parsedDataObject := strings.SplitN(object.FullName, ".", 4)
		if len(parsedDataObject) != 4 {
			return fmt.Errorf("unsupported reference entity id: %s", object.FullName)
		}

		f.stringBuilder.WriteString(parsedDataObject[3])

		return nil
	case datacomparison.EntityTypeColumnReferenceByName:
		f.stringBuilder.WriteString(ref.EntityID)

		return nil
	default:
		return fmt.Errorf("unsupported reference entity type: %s", ref.EntityType.String())
	}
}

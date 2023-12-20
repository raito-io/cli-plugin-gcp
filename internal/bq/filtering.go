package bigquery

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/raito-io/bexpression"
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

		filterExpression, err = createFilterExpression(ap.FilterCriteria)
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
		filterName = strings.ReplaceAll(strings.ReplaceAll(ap.NamingHint, " ", "_"), "-", "_")
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

func createFilterExpression(expr *bexpression.BinaryExpression) (string, error) {
	var queryBuilder strings.Builder

	var aggregatorOperandStack []string
	binaryExpressionLevel := 0

	traverser := bexpression.NewTraverser(
		bexpression.WithEnterBinaryExpressionFn(func(node *bexpression.BinaryExpression) error {
			if binaryExpressionLevel > 0 && node.Literal == nil {
				queryBuilder.WriteString("(")
			}

			binaryExpressionLevel++

			return nil
		}),
		bexpression.WithLeaveBinaryExpressionFn(func(node *bexpression.BinaryExpression) {
			binaryExpressionLevel--

			if binaryExpressionLevel > 0 && node.Literal == nil {
				queryBuilder.WriteString(")")
			}
		}),
		bexpression.WithLiteralBoolFn(func(b bool) error {
			if b {
				queryBuilder.WriteString("TRUE")
			} else {
				queryBuilder.WriteString("FALSE")
			}

			return nil
		}),
		bexpression.WithLiteralIntFn(func(i int) error {
			queryBuilder.WriteString(fmt.Sprintf("%d", i))

			return nil
		}),
		bexpression.WithLiteralFloatFn(func(f float64) error {
			queryBuilder.WriteString(fmt.Sprintf("%f", f))

			return nil
		}),
		bexpression.WithLiteralStringFn(func(value string) error {
			queryBuilder.WriteString(fmt.Sprintf("%q", value))

			return nil
		}),

		bexpression.WithLiteralTimestampFn(func(value time.Time) error {
			queryBuilder.WriteString(fmt.Sprintf("DATETIME(%d, %d, %d, %d, %d, %d)", value.Year(), value.Month(), value.Day(), value.Hour(), value.Minute(), value.Second()))

			return nil
		}),
		bexpression.WithComparisonOperatorFn(func(node bexpression.ComparisonOperator) error {
			switch node {
			case bexpression.ComparisonOperatorEqual:
				queryBuilder.WriteString(" = ")
			case bexpression.ComparisonOperatorNotEqual:
				queryBuilder.WriteString(" != ")
			case bexpression.ComparisonOperatorLessThan:
				queryBuilder.WriteString(" < ")
			case bexpression.ComparisonOperatorLessThanOrEqual:
				queryBuilder.WriteString(" <= ")
			case bexpression.ComparisonOperatorGreaterThan:
				queryBuilder.WriteString(" > ")
			case bexpression.ComparisonOperatorGreaterThanOrEqual:
				queryBuilder.WriteString(" >= ")
			default:
				return fmt.Errorf("unsupported comparison operator: %s", node)
			}

			return nil
		}),
		bexpression.WithReferenceFn(func(node *bexpression.Reference) error {
			if node.EntityType != bexpression.EntityTypeDataObject {
				return fmt.Errorf("unsupported reference entity type: %s", node.EntityType)
			}

			var object ds.DataObjectReference

			err := json.Unmarshal([]byte(node.EntityId), &object)
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

			queryBuilder.WriteString(parsedDataObject[3])

			return nil
		}),
		bexpression.WithEnterAggregatorFn(func(node *bexpression.Aggregator) error {
			var operand string

			switch node.Operator {
			case bexpression.AggregatorOperatorAnd:
				operand = "AND"
			case bexpression.AggregatorOperatorOr:
				operand = "OR"
			default:
				return fmt.Errorf("unsupported aggregation operator: %s", node.Operator)
			}

			aggregatorOperandStack = append(aggregatorOperandStack, operand)

			return nil
		}),
		bexpression.WithNextAggregatorOperand(func() error {
			queryBuilder.WriteString(fmt.Sprintf(" %s ", aggregatorOperandStack[len(aggregatorOperandStack)-1]))

			return nil
		}),
		bexpression.WithLeaveAggregatorFn(func(node *bexpression.Aggregator) {
			aggregatorOperandStack = aggregatorOperandStack[:len(aggregatorOperandStack)-1]
		}),
		bexpression.WithEnterUnaryExpressionFn(func(node *bexpression.UnaryExpression) error {
			if node.Operator != bexpression.UnaryOperatorNot {
				return fmt.Errorf("unsupported unary operator: %s", node.Operator)
			}

			queryBuilder.WriteString("NOT ")

			return nil
		}),
		bexpression.WithLeaveUnaryExpressionFn(func(node *bexpression.UnaryExpression) {
			queryBuilder.WriteString("")
		}),
	)

	err := traverser.TraverseBinaryExpression(expr)
	if err != nil {
		return "", fmt.Errorf("expression traversal: %w", err)
	}

	return queryBuilder.String(), nil
}

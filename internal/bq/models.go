package bigquery

import (
	"strings"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/bigquery/datapolicies/apiv1/datapoliciespb"
)

type GroupEntity struct {
	ExternalId string
	Email      string
	Members    []string
}

type UserEntity struct {
	ExternalId string
	Name       string
	Email      string
}

type BQEntity struct {
	ID         string
	Type       string
	Name       string
	FullName   string
	ParentId   string
	Location   string
	PolicyTags []string
}

type BQInformationSchemaEntity struct {
	CachedQuery   bool                                 `bigquery:"cache_hit"`
	User          string                               `bigquery:"user_email"`
	Query         string                               `bigquery:"query"`
	StatementType string                               `bigquery:"statement_type"`
	Tables        []BQInformationSchemaReferencedTable `bigquery:"referenced_tables"`
	StartTime     int64                                `bigquery:"start_time"`
	EndTime       int64                                `bigquery:"end_time"`
}

type BQInformationSchemaReferencedTable struct {
	Project bigquery.NullString `bigquery:"project_id"`
	Dataset bigquery.NullString `bigquery:"dataset_id"`
	Table   bigquery.NullString `bigquery:"table_id"`
}

type BQReferencedTable struct {
	Project string `bigquery:"project_id"`
	Dataset string `bigquery:"dataset_id"`
	Table   string `bigquery:"table_id"`
}

type BQFilter struct {
	FilterName       string
	Table            BQReferencedTable
	Users            []string
	Groups           []string
	FilterExpression string
}

type BQMaskingInformation struct {
	DataPolicy BQDataPolicy
	PolicyTag  BQPolicyTag
}

type BQDataPolicy struct {
	FullName   string
	PolicyType datapoliciespb.DataMaskingPolicy_PredefinedExpression
}

type BQPolicyTag struct {
	Name        string
	Description string
	FullName    string
	ParentTag   string
}

func (t *BQPolicyTag) Taxonomy() string {
	return strings.Join(strings.Split(t.FullName, "/")[0:6], "/")
}

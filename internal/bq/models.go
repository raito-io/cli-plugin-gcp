package bigquery

import (
	"strings"

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
	CachedQuery   bool                `bigquery:"cache_hit"`
	User          string              `bigquery:"user_email"`
	Query         string              `bigquery:"query"`
	StatementType string              `bigquery:"statement_type"`
	Tables        []BQReferencedTable `bigquery:"referenced_tables"`
	StartTime     int64               `bigquery:"start_time"`
	EndTime       int64               `bigquery:"end_time"`
}

type BQReferencedTable struct {
	Project string `bigquery:"project_id"`
	Dataset string `bigquery:"dataset_id"`
	Table   string `bigquery:"table_id"`
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

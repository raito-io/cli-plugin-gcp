package it

import (
	bigquery2 "cloud.google.com/go/bigquery"

	bigquery "github.com/raito-io/cli-plugin-gcp/internal/bq"
)

type TestRepositoryAndClient struct {
	Repository *bigquery.Repository
	Client     *bigquery2.Client
}

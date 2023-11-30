package bigquery

import (
	bigquery2 "cloud.google.com/go/bigquery"
)

type TestRepositoryAndClient struct {
	Repository *Repository
	Client     *bigquery2.Client
}

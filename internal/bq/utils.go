package bigquery

import (
	"cloud.google.com/go/bigquery"
)

func getRoleForBQEntity(t bigquery.AccessRole) string {
	if t == bigquery.OwnerRole {
		return "roles/bigquery.dataOwner"
	} else if t == bigquery.ReaderRole {
		return "roles/bigquery.dataViewer"
	} else if t == bigquery.WriterRole {
		return "roles/bigquery.dataEditor"
	}

	return string(t)
}

func GetValueIfExists[T any](p *T, defaultValue T) T {
	if p == nil {
		return defaultValue
	}

	return *p
}

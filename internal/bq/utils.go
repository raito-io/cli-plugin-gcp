package bigquery

import (
	"regexp"

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

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9_]+`)
var nonAlphaNumericToUnderscore = regexp.MustCompile(`[+|\\/\- ]+`)

func validSqlName(originalName string) string {
	return nonAlphanumericRegex.ReplaceAllString(nonAlphaNumericToUnderscore.ReplaceAllString(originalName, "_"), "")
}

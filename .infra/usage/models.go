package main

type ServiceAccountList struct {
	ProjectId     string        `json:"project_id"`
	EmailSuffix   string        `json:"email_suffix"`
	IamIdentities []IamIdentity `json:"iam_identities"`
}

type IamIdentity struct {
	Name         string              `json:"name"`
	Email        string              `json:"email"`
	Type         string              `json:"type"`
	Update       *UpdateUsageDetails `json:"update,omitempty"`
	DataObjects  []DatasetTables     `json:"data_objects"`
}

type DatasetTables struct {
	Dataset string   `json:"dataset"`
	Tables  []string `json:"tables"`
}

type DatasetTable struct {
	Dataset string `json:"dataset"`
	Table   string `json:"table"`
}

type UpdateUsageDetails struct {
	P        float64  `json:"p"`
	Queries []string `json:"queries"`
}

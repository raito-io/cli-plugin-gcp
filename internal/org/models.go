package org

type GcpOrgEntity struct {
	// GcpDetails
	EntryName string

	// RaitoDetails
	Id          string
	Name        string
	FullName    string
	Type        string
	Location    string
	Description string
	PolicyTags  []string
	Parent      *GcpOrgEntity
}

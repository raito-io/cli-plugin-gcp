package org

type GcpOrgEntity struct {
	// GcpDetails
	EntryName string

	// RaitoDetails
	Id     string
	Name   string
	Type   string
	Parent *GcpOrgEntity
}

package org

type GcpOrgEntity struct {
	Id     string
	Name   string
	Type   string
	Parent *GcpOrgEntity
}

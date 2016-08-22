package models

import "github.com/cf-furnace/route-emitter/cfroutes"

type DesiredLRP struct {
	ProcessGuid string
	LogGuid     string

	Routes          cfroutes.CFRoutes
	ModificationTag *ModificationTag

	Instances int32

	Ports []uint32
}

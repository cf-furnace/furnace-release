package routing_table

import "github.com/cf-furnace/route-emitter/models"

type ActualLRPRoutingInfo struct {
	ActualLRP  *models.ActualLRP
	Evacuating bool
}

func NewActualLRPRoutingInfo(lrp *models.ActualLRP) *ActualLRPRoutingInfo {
	return &ActualLRPRoutingInfo{
		ActualLRP:  lrp,
		Evacuating: false,
	}
}

package models

const (
	ActualLRPStateUnclaimed = "UNCLAIMED"
	ActualLRPStateClaimed   = "CLAIMED"
	ActualLRPStateRunning   = "RUNNING"
	ActualLRPStateCrashed   = "CRASHED"
)

type PortMapping struct {
	HostPort      uint32
	ContainerPort uint32
}

type ActualLRP struct {
	ProcessGuid  string
	InstanceGuid string
	Index        int32

	State string

	Address string
	Ports   []PortMapping
}

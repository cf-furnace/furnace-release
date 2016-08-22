package models

const (
	EventTypeInvalid = ""

	EventTypeDesiredLRPCreated = "desired_lrp_created"
	EventTypeDesiredLRPChanged = "desired_lrp_changed"
	EventTypeDesiredLRPRemoved = "desired_lrp_removed"

	EventTypeActualLRPCreated = "actual_lrp_created"
	EventTypeActualLRPChanged = "actual_lrp_changed"
	EventTypeActualLRPRemoved = "actual_lrp_removed"
)

type Event interface {
	Key() string
	EventType() string
}

type ActualLRPCreatedEvent struct {
	ActualLRP *ActualLRP
}

func (e *ActualLRPCreatedEvent) Key() string {
	return e.ActualLRP.ProcessGuid
}

func (e *ActualLRPCreatedEvent) EventType() string {
	return EventTypeActualLRPCreated
}

type ActualLRPChangedEvent struct {
	Before *ActualLRP
	After  *ActualLRP
}

func (e *ActualLRPChangedEvent) Key() string {
	return e.Before.ProcessGuid
}

func (e *ActualLRPChangedEvent) EventType() string {
	return EventTypeActualLRPChanged
}

type ActualLRPRemovedEvent struct {
	ActualLRP *ActualLRP
}

func (e *ActualLRPRemovedEvent) Key() string {
	return e.ActualLRP.ProcessGuid
}

func (e *ActualLRPRemovedEvent) EventType() string {
	return EventTypeActualLRPRemoved
}

type DesiredLRPCreatedEvent struct {
	DesiredLRP *DesiredLRP
}

func (e *DesiredLRPCreatedEvent) Key() string {
	return e.DesiredLRP.ProcessGuid
}

func (e *DesiredLRPCreatedEvent) EventType() string {
	return EventTypeDesiredLRPCreated
}

type DesiredLRPChangedEvent struct {
	Before *DesiredLRP
	After  *DesiredLRP
}

func (e *DesiredLRPChangedEvent) Key() string {
	return e.Before.ProcessGuid
}

func (e *DesiredLRPChangedEvent) EventType() string {
	return EventTypeDesiredLRPChanged
}

type DesiredLRPRemovedEvent struct {
	DesiredLRP *DesiredLRP
}

func (e *DesiredLRPRemovedEvent) Key() string {
	return e.DesiredLRP.ProcessGuid
}

func (e *DesiredLRPRemovedEvent) EventType() string {
	return EventTypeDesiredLRPRemoved
}

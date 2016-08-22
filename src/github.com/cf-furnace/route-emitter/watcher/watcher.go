package watcher

import (
	"os"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"github.com/cf-furnace/route-emitter/metric"
	"github.com/cf-furnace/route-emitter/models"
	"github.com/cf-furnace/route-emitter/nats_emitter"
	"github.com/cf-furnace/route-emitter/routing_table"
	"github.com/cf-furnace/route-emitter/syncer"
)

var (
	routesTotal  = metric.Metric("RoutesTotal")
	routesSynced = metric.Counter("RoutesSynced")

	routeSyncDuration = metric.Duration("RouteEmitterSyncDuration")

	routesRegistered   = metric.Counter("RoutesRegistered")
	routesUnregistered = metric.Counter("RoutesUnregistered")
)

type Watcher struct {
	clock      clock.Clock
	table      routing_table.RoutingTable
	emitter    nats_emitter.NATSEmitter
	syncEvents syncer.Events
	events     <-chan models.Event
	logger     lager.Logger
}

func NewWatcher(
	clock clock.Clock,
	table routing_table.RoutingTable,
	emitter nats_emitter.NATSEmitter,
	syncEvents syncer.Events,
	events <-chan models.Event,
	logger lager.Logger,
) *Watcher {
	return &Watcher{
		clock:      clock,
		table:      table,
		emitter:    emitter,
		syncEvents: syncEvents,
		events:     events,
		logger:     logger.Session("watcher"),
	}
}

func (watcher *Watcher) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	watcher.logger.Info("starting")

	close(ready)
	watcher.logger.Info("started")
	defer watcher.logger.Info("finished")

	for {
		select {
		case <-watcher.syncEvents.Emit:
			logger := watcher.logger.Session("emit")
			watcher.emit(logger)

		case event := <-watcher.events:
			watcher.handleEvent(watcher.logger, event)

		case <-signals:
			watcher.logger.Info("stopping")
			return nil
		}
	}
}

func (watcher *Watcher) emit(logger lager.Logger) {
	messagesToEmit := watcher.table.MessagesToEmit()

	logger.Debug("emitting-messages", lager.Data{"messages": messagesToEmit})
	err := watcher.emitter.Emit(messagesToEmit)
	if err != nil {
		logger.Error("failed-to-emit-routes", err)
	}

	routesSynced.Add(messagesToEmit.RouteRegistrationCount())
	err = routesTotal.Send(watcher.table.RouteCount())
	if err != nil {
		logger.Error("failed-to-send-routes-total-metric", err)
	}
}

func (watcher *Watcher) handleEvent(logger lager.Logger, event models.Event) {
	switch event := event.(type) {
	case *models.DesiredLRPCreatedEvent:
		watcher.handleDesiredCreate(logger, event.DesiredLRP)
	case *models.DesiredLRPChangedEvent:
		watcher.handleDesiredUpdate(logger, event.Before, event.After)
	case *models.DesiredLRPRemovedEvent:
		watcher.handleDesiredDelete(logger, event.DesiredLRP)
	case *models.ActualLRPCreatedEvent:
		watcher.handleActualCreate(logger, routing_table.NewActualLRPRoutingInfo(event.ActualLRP))
	case *models.ActualLRPChangedEvent:
		watcher.handleActualUpdate(logger,
			routing_table.NewActualLRPRoutingInfo(event.Before),
			routing_table.NewActualLRPRoutingInfo(event.After),
		)
	case *models.ActualLRPRemovedEvent:
		watcher.handleActualDelete(logger, routing_table.NewActualLRPRoutingInfo(event.ActualLRP))
	default:
		logger.Info("did-not-handle-unrecognizable-event", lager.Data{"event-type": event.EventType()})
	}
}

func (watcher *Watcher) handleDesiredCreate(logger lager.Logger, d *models.DesiredLRP) {
	logger = logger.Session("handle-desired-create", desiredLRPData(d))
	logger.Info("starting")
	defer logger.Info("complete")

	watcher.setRoutesForDesired(logger, d)
}

func (watcher *Watcher) handleDesiredUpdate(logger lager.Logger, before, after *models.DesiredLRP) {
	logger = logger.Session("handling-desired-update", lager.Data{
		"before": desiredLRPData(before),
		"after":  desiredLRPData(after),
	})
	logger.Info("starting")
	defer logger.Info("complete")

	afterKeysSet := watcher.setRoutesForDesired(logger, after)

	beforeRoutingKeys := routing_table.RoutingKeysFromDesired(before)

	afterContainerPorts := map[uint32]struct{}{}
	for _, route := range after.Routes {
		afterContainerPorts[route.Port] = struct{}{}
	}

	requestedInstances := after.Instances - before.Instances

	for _, key := range beforeRoutingKeys {
		_, hasKey := afterKeysSet[key]
		_, hasPort := afterContainerPorts[key.ContainerPort]
		if !hasKey || !hasPort {
			messagesToEmit := watcher.table.RemoveRoutes(key, after.ModificationTag)
			watcher.emitMessages(logger, messagesToEmit)
		}

		// in case of scale down, remove endpoints
		if requestedInstances < 0 {
			logger.Info("removing-endpoints", lager.Data{"removal_count": -1 * requestedInstances, "routing_key": key})

			for index := before.Instances - 1; index >= after.Instances; index-- {
				endpoints := watcher.table.EndpointsForIndex(key, index)

				for i, _ := range endpoints {
					messagesToEmit := watcher.table.RemoveEndpoint(key, endpoints[i])
					watcher.emitMessages(logger, messagesToEmit)
				}
			}
		}
	}
}

func (watcher *Watcher) setRoutesForDesired(logger lager.Logger, d *models.DesiredLRP) map[routing_table.RoutingKey]struct{} {
	routingKeySet := map[routing_table.RoutingKey]struct{}{}

	routeEntries := make(map[routing_table.RoutingKey][]routing_table.Route)
	for _, route := range d.Routes {
		key := routing_table.RoutingKey{ProcessGuid: d.ProcessGuid, ContainerPort: route.Port}
		routingKeySet[key] = struct{}{}

		routes := []routing_table.Route{}
		for _, hostname := range route.Hostnames {
			routes = append(routes, routing_table.Route{
				Hostname:        hostname,
				LogGuid:         d.LogGuid,
				RouteServiceUrl: route.RouteServiceUrl,
			})
		}
		routeEntries[key] = append(routeEntries[key], routes...)
	}
	for key := range routeEntries {
		messagesToEmit := watcher.table.SetRoutes(key, routeEntries[key], nil)
		watcher.emitMessages(logger, messagesToEmit)
	}

	return routingKeySet
}

func (watcher *Watcher) handleDesiredDelete(logger lager.Logger, d *models.DesiredLRP) {
	logger = logger.Session("handling-desired-delete", desiredLRPData(d))
	logger.Info("starting")
	defer logger.Info("complete")

	for _, key := range routing_table.RoutingKeysFromDesired(d) {
		messagesToEmit := watcher.table.RemoveRoutes(key, d.ModificationTag)

		watcher.emitMessages(logger, messagesToEmit)
	}
}

func (watcher *Watcher) handleActualCreate(logger lager.Logger, actualLRPInfo *routing_table.ActualLRPRoutingInfo) {
	logger = logger.Session("handling-actual-create", actualLRPData(actualLRPInfo))
	logger.Info("starting")
	defer logger.Info("complete")

	if actualLRPInfo.ActualLRP.State == models.ActualLRPStateRunning {
		watcher.addAndEmit(logger, actualLRPInfo)
	}
}

func (watcher *Watcher) handleActualUpdate(logger lager.Logger, before, after *routing_table.ActualLRPRoutingInfo) {
	logger = logger.Session("handling-actual-update", lager.Data{
		"before": actualLRPData(before),
		"after":  actualLRPData(after),
	})
	logger.Info("starting")
	defer logger.Info("complete")

	switch {
	case after.ActualLRP.State == models.ActualLRPStateRunning:
		watcher.addAndEmit(logger, after)
	case after.ActualLRP.State != models.ActualLRPStateRunning && before.ActualLRP.State == models.ActualLRPStateRunning:
		watcher.removeAndEmit(logger, before)
	}
}

func (watcher *Watcher) handleActualDelete(logger lager.Logger, actualLRPInfo *routing_table.ActualLRPRoutingInfo) {
	logger = logger.Session("handling-actual-delete", actualLRPData(actualLRPInfo))
	logger.Info("starting")
	defer logger.Info("complete")

	if actualLRPInfo.ActualLRP.State == models.ActualLRPStateRunning {
		watcher.removeAndEmit(logger, actualLRPInfo)
	}
}

func (watcher *Watcher) addAndEmit(logger lager.Logger, actualLRPInfo *routing_table.ActualLRPRoutingInfo) {
	logger.Info("watcher-add-and-emit", lager.Data{"net_info": lager.Data{"address": actualLRPInfo.ActualLRP.Address, "ports": actualLRPInfo.ActualLRP.Ports}})
	endpoints, err := routing_table.EndpointsFromActual(actualLRPInfo)
	if err != nil {
		logger.Error("failed-to-extract-endpoint-from-actual", err)
		return
	}

	for _, endpoint := range endpoints {
		key := routing_table.RoutingKey{ProcessGuid: actualLRPInfo.ActualLRP.ProcessGuid, ContainerPort: uint32(endpoint.ContainerPort)}

		messagesToEmit := watcher.table.AddEndpoint(key, endpoint)
		watcher.emitMessages(logger, messagesToEmit)
	}
}

func (watcher *Watcher) removeAndEmit(logger lager.Logger, actualLRPInfo *routing_table.ActualLRPRoutingInfo) {
	logger.Info("watcher-remove-and-emit", lager.Data{"net_info": lager.Data{"address": actualLRPInfo.ActualLRP.Address, "ports": actualLRPInfo.ActualLRP.Ports}})
	endpoints, err := routing_table.EndpointsFromActual(actualLRPInfo)
	if err != nil {
		logger.Error("failed-to-extract-endpoint-from-actual", err)
		return
	}

	for _, key := range routing_table.RoutingKeysFromActual(actualLRPInfo.ActualLRP) {
		for _, endpoint := range endpoints {
			if key.ContainerPort == endpoint.ContainerPort {
				messagesToEmit := watcher.table.RemoveEndpoint(key, endpoint)
				watcher.emitMessages(logger, messagesToEmit)
			}
		}
	}
}

func (watcher *Watcher) emitMessages(logger lager.Logger, messagesToEmit routing_table.MessagesToEmit) {
	if watcher.emitter != nil {
		logger.Debug("emit-messages", lager.Data{"messages": messagesToEmit})
		watcher.emitter.Emit(messagesToEmit)
		routesRegistered.Add(messagesToEmit.RouteRegistrationCount())
		routesUnregistered.Add(messagesToEmit.RouteUnregistrationCount())
	}
}

func desiredLRPData(d *models.DesiredLRP) lager.Data {
	return lager.Data{
		"process-guid": d.ProcessGuid,
		"routes":       d.Routes,
		"instances":    d.Instances,
	}
}

func actualLRPData(lrpRoutingInfo *routing_table.ActualLRPRoutingInfo) lager.Data {
	lrp := lrpRoutingInfo.ActualLRP
	return lager.Data{
		"process-guid": lrp.ProcessGuid,
		"index":        lrp.Index,
		// "domain":        lrp.Domain,
		"instance-guid": lrp.InstanceGuid,
		// "cell-id":       lrp.ActualLRPInstanceKey.CellId,
		"address":    lrp.Address,
		"evacuating": lrpRoutingInfo.Evacuating,
	}
}

package endpoint

import (
	"endpoint-visualiser-server/pkg/clienthandler/websocket"
	"endpoint-visualiser-server/pkg/event"
	"fmt"
	"sync"
)

type Manager struct {
	config           []ManagableEndpoint
	websocketManager *websocket.Manager
	eventInChan      <-chan event.Event
}

type ManagerOption func(*Manager)

type ManagableEndpoint struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	MaxConns int    `json:"maxConns"`
}

func WithConfig(config []ManagableEndpoint) ManagerOption {
	return func(m *Manager) {
		m.config = config
	}
}

func WithWebSocketTarget(target *websocket.Manager) ManagerOption {
	return func(m *Manager) {
		m.websocketManager = target
	}
}

func NewManager(eventChan <-chan event.Event, opts ...ManagerOption) *Manager {
	manager := &Manager{eventInChan: eventChan}
	for _, opt := range opts {
		opt(manager)
	}
	return manager
}

func (m *Manager) Start(synchStart *sync.WaitGroup) {
	synchStart.Add(len(m.config) + 1) //  One for each endpoint and the router
	routingMap := make(map[int]chan<- interface{})

	for _, endpoint := range m.config {
		endpointEventInChan := make(chan interface{})
		routingMap[endpoint.ID] = endpointEventInChan
		go endpointProcessor(endpoint, m.websocketManager.GetSingleRequestSender(endpoint.ID), endpointEventInChan)
		synchStart.Done()
	}

	go routeEvents(m.eventInChan, routingMap)
	synchStart.Done()
}

func routeEvents(inChan <-chan event.Event, routeMap map[int]chan<- interface{}) {
	for e := range inChan {
		fmt.Printf("\nEndpoint Manager Router Recevived event %T on inputChan, sending to endpoint %d", e.Event, e.Destination)

		if routeChan := routeMap[e.Destination]; routeChan != nil {
			routeChan <- e
			continue
		}
		fmt.Printf("\nYou cannot send messages to endpoint %d, it doesn't exist in your config!", e.Destination)
	}
}

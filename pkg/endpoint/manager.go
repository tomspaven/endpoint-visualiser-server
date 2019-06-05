package endpoint

import (
	"endpoint-visualiser-server/pkg/clienthandler/websocket"
	"endpoint-visualiser-server/pkg/event"
	"io/ioutil"
	"log"
	"sync"
)

type Manager struct {
	config           []ManagableEndpoint
	websocketManager *websocket.Manager
	eventInChan      <-chan event.Event
	logger           *log.Logger
	cntl             controlStructures
	handlerMap       map[string]eventHandler
}

type controlStructures struct {
	stopChan        chan struct{}
	changeDelayChan chan int
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

func WithLogger(l *log.Logger) ManagerOption {
	return func(m *Manager) {
		m.logger = l
	}
}

func WithWebSocketTarget(target *websocket.Manager) ManagerOption {
	return func(m *Manager) {
		m.websocketManager = target
	}
}

func NewManager(eventChan <-chan event.Event, opts ...ManagerOption) *Manager {
	defaultDiscardLogger := log.New(ioutil.Discard, "", 0)
	manager := &Manager{
		eventInChan: eventChan,
		cntl: controlStructures{
			stopChan:        make(chan struct{}),
			changeDelayChan: make(chan int),
		},
		logger: defaultDiscardLogger,
	}
	manager.setupHandlerMap()

	for _, opt := range opts {
		opt(manager)
	}
	return manager
}

func (m *Manager) setupHandlerMap() {
	handlerMap := make(map[string]eventHandler)
	handlerMap[event.ConnectEvent{}.String()] = eventHandler{connectEventPredicate, m.connectEventAction}
	handlerMap[event.DisconnectEvent{}.String()] = eventHandler{disconnectEventPredicate, m.disconnectEventAction}
	handlerMap[event.StartTrafficEvent{}.String()] = eventHandler{startTrafficEventPredicate, m.startTrafficEventAction}
	handlerMap[event.StopTrafficEvent{}.String()] = eventHandler{stopTrafficEventPredicate, m.stopTrafficEventAction}
	handlerMap[event.DelayShortEvent{}.String()] = eventHandler{delayShortEventPredicate, m.delayShortEventAction}
	handlerMap[event.DelayMediumEvent{}.String()] = eventHandler{delayMediumEventPredicate, m.delayMediumEventAction}
	handlerMap[event.DelayLongEvent{}.String()] = eventHandler{delayLongEventPredicate, m.delayLongEventAction}
	handlerMap[event.StopRespondingEvent{}.String()] = eventHandler{delayStopRespondingEventPredicate, m.delayStopRespondingEventAction}
	handlerMap[event.StartRespondingEvent{}.String()] = eventHandler{delayStartRespondingEventPredicate, m.delayStartRespondingEventAction}
	m.handlerMap = handlerMap
	return
}

func (m *Manager) Start(synchStart *sync.WaitGroup) {
	synchStart.Add(len(m.config) + 1) //  One for each endpoint and the router
	routingMap := make(map[int]chan<- interface{})

	for _, endpoint := range m.config {
		endpointEventInChan := make(chan interface{})
		routingMap[endpoint.ID] = endpointEventInChan
		go m.endpointProcessor(endpoint, endpointEventInChan)
		synchStart.Done()
	}

	go m.routeEvents(m.eventInChan, routingMap)
	synchStart.Done()
}

func (m *Manager) routeEvents(inChan <-chan event.Event, routeMap map[int]chan<- interface{}) {
	for e := range inChan {
		m.logger.Printf("\nEndpoint Manager Router Recevived event %T on inputChan, sending to endpoint %d", e.Event, e.Destination)

		if routeChan := routeMap[e.Destination]; routeChan != nil {
			routeChan <- e
			continue
		}
		m.logger.Printf("\nYou cannot send messages to endpoint %d, it doesn't exist in your config!", e.Destination)
	}
}

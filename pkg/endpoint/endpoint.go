package endpoint

import (
	"endpoint-visualiser-server/pkg/event"
)

type ClientSender func(interface{}) error

type epState int

const (
	epStateDown      = 0
	epStateUpWaiting = iota
	epStateUpReceiving
	epStateImpared
)

type endpointProcessingState struct {
	endpointState     epState
	currentDelayState int
}

func (m *Manager) endpointProcessor(epConfig ManagableEndpoint, eventInChan <-chan interface{}) {

	m.logger.Printf("\nEndpoint Processor %d started!", epConfig.ID)

	sender := m.websocketManager.GetSingleRequestSender(epConfig.ID)
	state := endpointProcessingState{
		endpointState:     epStateDown,
		currentDelayState: noResponseDelayMS,
	}

	for eRaw := range eventInChan {
		m.logger.Printf("\nEndpoint Processor %d received message on inChan!", epConfig.ID)
		if e, ok := eRaw.(event.Event); ok {
			var payload interface{}
			payload, state = m.handlerMap[e.String()].handleEvent(state, sender)
			if payload != nil {
				m.logger.Printf("\n Sending message to client: %s", payload)
				err := sender(payload)
				if err != nil {
					m.logger.Printf("\nError sending message to client: %s", err.Error())
				}
			}
			continue
		}
		m.logger.Printf("\nType assert error on event received by endpoint %d", epConfig.ID)
	}
}

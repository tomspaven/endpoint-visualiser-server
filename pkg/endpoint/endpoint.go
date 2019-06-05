package endpoint

import (
	"endpoint-visualiser-server/pkg/event"
	"math/rand"
	"time"
)

type messageID string

const (
	endpointConnected    messageID = "EndpointConnected"
	endpointDisconnected messageID = "EndpointDisconnected"
	endpointImpaired     messageID = "EndpointImpaired"
	trafficRequest       messageID = "TrafficReqeust"
	trafficResponse      messageID = "TrafficResponse"
)

const (
	noResponseDelayMS     int = 0
	shortResponseDelayMS  int = 500
	mediumResponseDelayMS int = 1000
	longResponseDelayMS   int = 3000
	stopRespondingMS      int = -1
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

type characterGenerator func() (string, func() int)

const (
	heartbeatIntervalMS     int = 3000
	maxWaitForNextMessageMS int = 2000
)

func heartBeatGenerator() (character string, delayFunc func() int) {
	return "❤️", func() int { return heartbeatIntervalMS }
}

func randomMessageGenerator() (character string, delayFunc func() int) {
	genRand := func(min, max int) int { return rand.Intn(max-min) + min }
	messages := []string{"🐷", "🐤", "🏈", "⚽", "🍋", "🍌"}
	nextMessageDelayFunc := func() int {
		return genRand(0, maxWaitForNextMessageMS)
	}

	return messages[genRand(0, len(messages))], nextMessageDelayFunc
}

func (m *Manager) trafficInitiator(sender ClientSender, generateCharacter characterGenerator) {

	responseDelay := noResponseDelayMS

goRoutineLoop:
	for {
		select {
		case <-m.cntl.stopChan:
			break goRoutineLoop
		case responseDelay = <-m.cntl.changeDelayChan:
			continue
		default:
			char, getNextMessageDelay := generateCharacter()
			errChan := make(chan error)
			go sendMessage(sender, errChan, char, responseDelay)
			nextMessageTimer := time.NewTimer(time.Duration(getNextMessageDelay()) * time.Millisecond)
			select {
			case <-nextMessageTimer.C:
				continue
			case <-m.cntl.stopChan:
				break goRoutineLoop
			case responseDelay = <-m.cntl.changeDelayChan:
				continue
			case err := <-errChan:
				m.logger.Printf("\n%s", err.Error())
				break goRoutineLoop
			}
		}
	}
	return
}

type TrafficMessage struct {
	ID        string `json:"id"`
	Character string `json:"character"`
}

const clientRenderLatencyMS int = 400

func sendMessage(clientSender ClientSender, errChan chan<- error, char string, responseDelay int) {

	request := TrafficMessage{ID: "TrafficRequest", Character: char}
	if err := clientSender(request); err != nil {
		errChan <- err
		return
	}

	if responseDelay != stopRespondingMS {
		responseDelay += clientRenderLatencyMS
		responseTimer := time.NewTimer((time.Duration(responseDelay) * time.Millisecond))
		<-responseTimer.C
		response := TrafficMessage{ID: "TrafficResponse", Character: char}

		if err := clientSender(response); err != nil {
			errChan <- err
			return
		}
	}
}

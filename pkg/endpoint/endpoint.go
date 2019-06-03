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

type EndpointConnectedMessage struct {
	RequestID      messageID `json:"id"`
	NumConnections int       `json:"numConnections"`
}

type EndpointDisconnectedMessage struct {
	RequestID messageID `json:"id"`
}

type EndpointImpairmentMessage struct {
	RequestID                messageID `json:"id"`
	WorstImparedResponseTime int       `json:"worstResponse"`
	ImparedResponseTime      int       `json:"time"`
}

type ClientSender func(interface{}) error

func (m *Manager) endpointProcessor(epConfig ManagableEndpoint, clientSender ClientSender, eventInChan <-chan interface{}) {
	m.logger.Printf("\nEndpoint Processor %d started!", epConfig.ID)

	stopTrafficChan := make(chan struct{})
	changeDelayChan := make(chan int)

	for eRaw := range eventInChan {
		m.logger.Printf("\nEndpoint Processor %d received message on inChan!", epConfig.ID)
		if e, ok := eRaw.(event.Event); ok {
			var payload interface{}

			switch e.Event.(type) {
			case event.ConnectEvent:
				payload = EndpointConnectedMessage{endpointConnected, 16}
				go m.trafficInitiator(clientSender, heartBeatGenerator, stopTrafficChan, changeDelayChan) // Start heartbeat
				break
			case event.DisconnectEvent:
				stopTrafficChan <- struct{}{}
				payload = EndpointDisconnectedMessage{endpointDisconnected}
				break
			case event.StartTrafficEvent:
				stopTrafficChan <- struct{}{}                                                                 // Stop heartbeat
				go m.trafficInitiator(clientSender, randomMessageGenerator, stopTrafficChan, changeDelayChan) // Start traffic
				continue
			case event.StopTrafficEvent:
				stopTrafficChan <- struct{}{}                                                             // Stop traffic
				go m.trafficInitiator(clientSender, heartBeatGenerator, stopTrafficChan, changeDelayChan) // Start heartbeat
				continue
			case event.DelayShortEvent:
				changeDelayChan <- shortResponseDelayMS
				payload = EndpointImpairmentMessage{
					RequestID:                endpointImpaired,
					WorstImparedResponseTime: longResponseDelayMS,
					ImparedResponseTime:      shortResponseDelayMS,
				}
				break
			case event.DelayMediumEvent:
				changeDelayChan <- mediumResponseDelayMS
				payload = EndpointImpairmentMessage{
					RequestID:                endpointImpaired,
					WorstImparedResponseTime: longResponseDelayMS,
					ImparedResponseTime:      mediumResponseDelayMS,
				}
				break
			case event.DelayLongEvent:
				changeDelayChan <- longResponseDelayMS
				payload = EndpointImpairmentMessage{
					RequestID:                endpointImpaired,
					WorstImparedResponseTime: longResponseDelayMS,
					ImparedResponseTime:      longResponseDelayMS,
				}
				break
			case event.StopRespondingEvent:
				changeDelayChan <- stopRespondingMS
				payload = EndpointImpairmentMessage{
					RequestID:                endpointImpaired,
					WorstImparedResponseTime: longResponseDelayMS,
					ImparedResponseTime:      stopRespondingMS,
				}
				break
			case event.StartRespondingEvent:
				changeDelayChan <- noResponseDelayMS
				payload = EndpointImpairmentMessage{
					RequestID:                endpointImpaired,
					WorstImparedResponseTime: longResponseDelayMS,
					ImparedResponseTime:      noResponseDelayMS,
				}
				break
			default:
				break
			}

			m.logger.Printf("\n Sending message to client: %s", payload)
			err := clientSender(payload)
			if err != nil {
				m.logger.Printf("\nError sending message to client: %s", err.Error())
			}

			continue
		}
		//  Type assert error
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

func (m *Manager) trafficInitiator(clientSender ClientSender, generateCharacter characterGenerator, stopChan <-chan struct{}, changeDelayChan <-chan int) {

	responseDelay := noResponseDelayMS

goRoutineLoop:
	for {
		select {
		case <-stopChan:
			break goRoutineLoop
		case responseDelay = <-changeDelayChan:
			continue
		default:
			char, getNextMessageDelay := generateCharacter()
			errChan := make(chan error)
			go sendMessage(clientSender, errChan, char, responseDelay)
			nextMessageTimer := time.NewTimer(time.Duration(getNextMessageDelay()) * time.Millisecond)
			select {
			case <-nextMessageTimer.C:
				continue
			case <-stopChan:
				break goRoutineLoop
			case responseDelay = <-changeDelayChan:
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

	responseDelay += clientRenderLatencyMS
	responseTimer := time.NewTimer((time.Duration(responseDelay) * time.Millisecond))
	<-responseTimer.C
	response := TrafficMessage{ID: "TrafficResponse", Character: char}

	if err := clientSender(response); err != nil {
		errChan <- err
		return
	}
}

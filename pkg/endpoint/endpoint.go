package endpoint

import (
	"endpoint-visualiser-server/pkg/event"
	"fmt"
	"math/rand"
	"time"
)

type messageID string

const (
	endpointConnected    messageID = "EndpointConnected"
	endpointDisconnected messageID = "EndpointDisconnected"
	endpointImpared      messageID = "EndpointImpared"
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
	ImparedResponseTime      int       `json:"worstResponse"`
}

type ClientSender func(interface{}) error

func endpointProcessor(epConfig ManagableEndpoint, clientSender ClientSender, eventInChan <-chan interface{}) {
	fmt.Printf("\nEndpoint Processor %d started!", epConfig.ID)

	stopTrafficChan := make(chan struct{})
	changeDelayChan := make(chan int)

	for eRaw := range eventInChan {

		fmt.Printf("\nEndpoint Processor %d received message on inChan!", epConfig.ID)
		if e, ok := eRaw.(event.Event); ok {
			var payload interface{}

			switch e.Event.(type) {
			case event.ConnectEvent:
				payload = EndpointConnectedMessage{endpointConnected, 16}
				go trafficInitiator(clientSender, heartBeatGenerator, stopTrafficChan, changeDelayChan) // Start heartbeat
				break
			case event.DisconnectEvent:
				stopTrafficChan <- struct{}{}
				payload = EndpointDisconnectedMessage{endpointDisconnected}
				break
			case event.StartTrafficEvent:
				stopTrafficChan <- struct{}{}                                                               // Stop heartbeat
				go trafficInitiator(clientSender, randomMessageGenerator, stopTrafficChan, changeDelayChan) // Start traffic
				continue
			case event.StopTrafficEvent:
				stopTrafficChan <- struct{}{}                                                           // Stop traffic
				go trafficInitiator(clientSender, heartBeatGenerator, stopTrafficChan, changeDelayChan) // Start heartbeat
				continue
			case event.DelayShortEvent:
				changeDelayChan <- shortResponseDelayMS
				payload = EndpointImpairmentMessage{
					RequestID:                endpointImpared,
					WorstImparedResponseTime: longResponseDelayMS,
					ImparedResponseTime:      shortResponseDelayMS,
				}
				break
			case event.DelayMediumEvent:
				changeDelayChan <- mediumResponseDelayMS
				payload = EndpointImpairmentMessage{
					RequestID:                endpointImpared,
					WorstImparedResponseTime: longResponseDelayMS,
					ImparedResponseTime:      mediumResponseDelayMS,
				}
				break
			case event.DelayLongEvent:
				changeDelayChan <- longResponseDelayMS
				payload = EndpointImpairmentMessage{
					RequestID:                endpointImpared,
					WorstImparedResponseTime: longResponseDelayMS,
					ImparedResponseTime:      longResponseDelayMS,
				}
				break
			case event.StopRespondingEvent:
				changeDelayChan <- stopRespondingMS
				payload = EndpointImpairmentMessage{
					RequestID:                endpointImpared,
					WorstImparedResponseTime: longResponseDelayMS,
					ImparedResponseTime:      stopRespondingMS,
				}
				break
			case event.StartRespondingEvent:
				changeDelayChan <- noResponseDelayMS
				payload = EndpointImpairmentMessage{
					RequestID:                endpointImpared,
					WorstImparedResponseTime: longResponseDelayMS,
					ImparedResponseTime:      noResponseDelayMS,
				}
				break
			default:
				break
			}

			fmt.Printf("\n Sending message to client: %s", payload)
			err := clientSender(payload)
			if err != nil {
				fmt.Printf("\nError sending message to client: %s", err.Error())
			}

			continue
		}
		//  Type assert error
		fmt.Printf("\nType assert error on event received by endpoint %d", epConfig.ID)
	}
}

type characterGenerator func() (string, func() int)

const heartbeatIntervalMS = 3000

func heartBeatGenerator() (character string, delayFunc func() int) {
	return "❤️", func() int { return heartbeatIntervalMS }
}

const maxWaitForNextMessageMS int = 2000

func randomMessageGenerator() (character string, delayFunc func() int) {
	genRand := func(min, max int) int { return rand.Intn(max-min) + min }
	messages := []string{"🐷", "🐤", "🏈", "⚽", "🍋", "🍌"}
	nextMessageDelayFunc := func() int {
		return genRand(0, maxWaitForNextMessageMS)
	}

	return messages[genRand(0, len(messages))], nextMessageDelayFunc
}

func trafficInitiator(clientSender ClientSender, generateCharacter characterGenerator, stopChan <-chan struct{}, changeDelayChan <-chan int) {

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
			go sendMessage(clientSender, char, responseDelay)
			nextMessageTimer := time.NewTimer(time.Duration(getNextMessageDelay()) * time.Millisecond)
			select {
			case <-nextMessageTimer.C:
				continue
			case <-stopChan:
				break goRoutineLoop
			}
		}
	}
}

type TrafficMessage struct{ id, character string }

func sendMessage(clientSender ClientSender, char string, responseDelay int) {

	request := TrafficMessage{id: "TrafficReqeust", character: char}
	if err := clientSender(request); err != nil {
		fmt.Printf("\nError sending request %s to client", char)
		return
	}

	responseTimer := time.NewTimer(time.Duration(responseDelay) * time.Millisecond)
	<-responseTimer.C
	response := TrafficMessage{id: "TrafficResponse", character: char}

	if err := clientSender(response); err != nil {
		fmt.Printf("\nError sending response %s to client", char)
		return
	}
}

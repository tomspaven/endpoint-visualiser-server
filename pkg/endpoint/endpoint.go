package endpoint

import (
	"endpoint-visualiser-server/pkg/event"
	"fmt"
)

type EndpointConnectedMessage struct {
	ID             int `json:"id"`
	NumConnections int `json:"numConnections"`
}

func endpointProcessor(epConfig ManagableEndpoint, clientSender func(interface{}) error, eventInChan <-chan interface{}) {

	fmt.Printf("\nEndpoint Processor %d started!", epConfig.ID)
	for eRaw := range eventInChan {

		fmt.Printf("\nEndpoint Processor %d received message on inChan!", epConfig.ID)
		if e, ok := eRaw.(event.Event); ok {
			var payload interface{}

			switch e.Event.(type) {
			case event.ConnectEvent:
				payload = EndpointConnectedMessage{
					epConfig.ID,
					epConfig.MaxConns,
				}
				break
			default:
				break
			}

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

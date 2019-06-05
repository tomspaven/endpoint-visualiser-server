package endpoint_test

import (
	"endpoint-visualiser-server/pkg/clienthandler/websocket"
	"endpoint-visualiser-server/pkg/endpoint"
	"endpoint-visualiser-server/pkg/event"
	"math/rand"
	"testing"
)

func TestEndpointProcessor(t *testing.T) {

	const numEndpoints = 10
	eventChan := make(chan event.Event)

	mockWebSocketManager := websocket.New(
		websocket.GetMockWithClientRegisterer(t, numEndpoints),
	)

	_ = endpoint.NewManager(eventChan,
		endpoint.WithConfig(generateConfig(numEndpoints, configGenerators{randomMaxConnsGenerator, randomCharStringGenerator})),
		endpoint.WithWebSocketTarget(mockWebSocketManager))
}

type maxConnsGenerator func(int) int
type titleGenerator func(int) string
type configGenerators struct {
	maxConnsGenerator
	titleGenerator
}

func randomMaxConnsGenerator(maxConns int) int {
	return rand.Intn(maxConns)
}

func randomCharStringGenerator(numChars int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, numChars)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func generateConfig(numEndpoints int, gens configGenerators) (config []endpoint.ManagableEndpoint) {
	const (
		maxConns     = 64
		numNameChars = 12
	)

	config = make([]endpoint.ManagableEndpoint, numEndpoints)
	for i := 0; i < numEndpoints; i++ {
		config[i] = endpoint.ManagableEndpoint{
			ID:       i,
			Title:    gens.titleGenerator(numNameChars),
			MaxConns: gens.maxConnsGenerator(maxConns),
		}
	}
	return
}

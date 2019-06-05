package websocket

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gorilla/websocket"
)

func GetMockWithClientRegisterer(t *testing.T, numEndpoints int) func(*Manager) {
	return func(m *Manager) {
		// Dont need to register anything.
		m.registrationHandler = func(w http.ResponseWriter, r *http.Request) { return }

		for i := 0; i < numEndpoints; i++ {
			m.clientLock.Lock()
			m.clients[i] = generateMockWSClient(t)
			m.clientLock.Unlock()
		}
		return
	}

}

func generateMockWSClient(t *testing.T) *websocket.Conn {
	mockURL := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/echo"}
	mockClient, _, err := websocket.DefaultDialer.Dial(mockURL.String(), nil)
	if err != nil {
		t.Fail()
	}
	return mockClient
}

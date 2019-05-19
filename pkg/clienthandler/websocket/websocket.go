package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Manager struct {
	registrationHandler func(w http.ResponseWriter, r *http.Request)
	clients             map[int]*websocket.Conn
}

type ManagerOption func(*Manager)

func WithClientRegisterer(m *Manager) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	m.registrationHandler = func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			buildErrorResponse(w, err)
			return
		}

		fmt.Printf("\nReceived Registration Request for endpoint %d!", id)
		websocket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			buildErrorResponse(w, err)
			return
		}
		m.clients[id] = websocket
		buildResponse(w, nil)
	}
}

func buildErrorResponse(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	buildResponse(w, struct {
		Error string `json:"error"`
	}{err.Error()})

}

func buildResponse(w http.ResponseWriter, payload interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if payload != nil {
		prettyJSON, _ := json.MarshalIndent(payload, "", "    ")
		fmt.Printf("Sending %s", prettyJSON)
		json.NewEncoder(w).Encode(payload)
	}
}

func New(opts ...ManagerOption) *Manager {
	manager := &Manager{clients: make(map[int]*websocket.Conn)}
	for _, opt := range opts {
		opt(manager)
	}
	return manager
}

func (m *Manager) Handler() func(w http.ResponseWriter, r *http.Request) {
	return m.registrationHandler
}

func (m *Manager) GetSingleRequestSender(id int) func(interface{}) error {
	return func(event interface{}) error {
		return m.sendRequestToSingleClient(id, event)
	}
}

func (m *Manager) sendRequestToSingleClient(id int, event interface{}) error {

	if m.clients[id] == nil {
		return fmt.Errorf("No client has registered to receive websocket events for endpoint %d", id)
	}

	bytes, err := json.Marshal(event)
	if err == nil {
		err = m.clients[id].WriteMessage(websocket.TextMessage, bytes)
	}
	return err
}

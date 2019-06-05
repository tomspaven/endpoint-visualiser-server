package rest

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type RestManager struct {
	config []DiscoverableEndpoint
	logger *log.Logger
}

func New(opts ...ManagerOption) *RestManager {
	defaultDiscardLogger := log.New(ioutil.Discard, "", 0)
	manager := &RestManager{logger: defaultDiscardLogger}
	for _, opt := range opts {
		opt(manager)
	}
	return manager
}

type DiscoverableEndpoint struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	MaxConns int    `json:"maxConns"`
}

type ManagerOption func(*RestManager)

func WithConfig(config []DiscoverableEndpoint) ManagerOption {
	return func(m *RestManager) {
		m.config = config
	}
}

func WithLogger(l *log.Logger) ManagerOption {
	return func(m *RestManager) {
		m.logger = l
	}
}

func (m *RestManager) EndpointDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	m.buildResponse(w, m.config)
	return
}

func (m *RestManager) buildErrorResponse(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	m.buildResponse(w, struct {
		Error string `json:"error"`
	}{err.Error()})

}

func (m *RestManager) buildResponse(w http.ResponseWriter, payload interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if payload != nil {
		prettyJSON, _ := json.MarshalIndent(payload, "", "    ")
		m.logger.Printf("\nSending discovery response %s!", prettyJSON)
		json.NewEncoder(w).Encode(payload)
	}
}

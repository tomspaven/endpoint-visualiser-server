package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type RestManager struct {
	config []DiscoverableEndpoint
}

func New(opts ...ManagerOption) *RestManager {
	manager := &RestManager{}
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

func (m *RestManager) EndpointDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	buildResponse(w, m.config)
	return
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
		fmt.Printf("\nSending discovery response %s!", prettyJSON)
		json.NewEncoder(w).Encode(payload)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"endpoint-visualiser-server/pkg/clienthandler/rest"
	"endpoint-visualiser-server/pkg/clienthandler/websocket"
	"endpoint-visualiser-server/pkg/endpoint"
	"endpoint-visualiser-server/pkg/event"
	"endpoint-visualiser-server/pkg/keyboard"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	logfile, err := os.OpenFile("log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Couldn't initialise log file: %s", err.Error())
		os.Exit(1)
	}
	logger := log.New(logfile, "", 0)

	config, err := ReadConfig()
	if err != nil {
		fmt.Printf("Failed To Read Config")
		os.Exit(1)
	}

	eventChan := make(chan event.Event)

	webSocketManager := websocket.New(
		websocket.WithClientRegisterer,
		websocket.WithLogger(logger),
	)

	restManager := rest.New(
		rest.WithConfig(config.Endpoints),
		rest.WithLogger(logger),
	)

	endpointManager := endpoint.NewManager(eventChan,
		endpoint.WithConfig(copyEnpointConfig(config.Endpoints)),
		endpoint.WithWebSocketTarget(webSocketManager),
		endpoint.WithLogger(logger),
	)

	keyListener, err := keyboard.NewListener(eventChan,
		keyboard.WithConfig(config.KeyProfiles),
		keyboard.WithLogger(logger),
	)

	if err != nil {
		log.Printf("KeyListener Startup Failed. Error: %s", err.Error())
	}

	router := mux.NewRouter()
	router.HandleFunc("/endpoints", restManager.EndpointDiscoveryHandler).Methods("GET")
	router.Path("/websocketRegistration/{id:[0-9]+}").HandlerFunc(webSocketManager.Handler())

	synchStart := &sync.WaitGroup{}
	endpointManager.Start(synchStart)
	keyListener.Start(synchStart)
	synchStart.Wait()

	fmt.Printf("\nListening on port 3031")
	log.Fatal(http.ListenAndServe(":3031", handlers.CORS(handlers.AllowedMethods([]string{"GET", "POST"}), handlers.AllowedOrigins([]string{"*"}))(router)))
}

type Config struct {
	Endpoints   []rest.DiscoverableEndpoint `json:"endpoints"`
	KeyProfiles []keyboard.KeyPressProfile  `json:"keypressProfiles"`
}

func copyEnpointConfig(deps []rest.DiscoverableEndpoint) []endpoint.ManagableEndpoint {
	managableEndpoints := make([]endpoint.ManagableEndpoint, len(deps))
	for i, dep := range deps {
		managableEndpoints[i] = endpoint.ManagableEndpoint{
			ID:       dep.ID,
			Title:    dep.Title,
			MaxConns: dep.MaxConns,
		}
	}
	return managableEndpoints
}

func ReadConfig() (Config, error) {
	configFile, err := os.Open("config.json")
	if err != nil {
		return Config{}, err
	}
	defer configFile.Close()

	fileDataBytes, _ := ioutil.ReadAll(configFile)
	var config Config
	err = json.Unmarshal(fileDataBytes, &config)
	return config, err
}

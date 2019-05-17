package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"endpoint-visualiser-server/pkg/clienthandler/rest"
	"endpoint-visualiser-server/pkg/clienthandler/websocket"
	"endpoint-visualiser-server/pkg/config"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	config, err := config.ReadConfig()
	if err != nil {
		fmt.Printf("Failed To Read Config")
		os.Exit(1)
	}

	//eventChan := make(chan event.Event)

	webSocketManager := websocket.New(websocket.WithClientRegisterer)
	restManager := rest.New(config)

	//endpointManager := endpoint.NewManager(config, webSocketManager, eventChan)
	//keyListener, err := keyboard.NewListener(config, eventChan)
	if err != nil {
		fmt.Printf("KeyListener Startup Failed. Error: %s", err.Error())
	}

	router := mux.NewRouter()
	router.HandleFunc("/endpoints", restManager.EndpointDiscoveryHandler).Methods("GET")
	router.Path("/websocketRegistration/{id:[0-9]+}").HandlerFunc(webSocketManager.Handler())

	//synchStart := &sync.WaitGroup{}
	//endpointManager.Start(synchStart)
	//keyListener.Start(synchStart)
	//synchStart.Wait()

	fmt.Printf("\nListening on port 3031")
	log.Fatal(http.ListenAndServe(":3031", handlers.CORS(handlers.AllowedMethods([]string{"GET", "POST"}), handlers.AllowedOrigins([]string{"*"}))(router)))
}

package main

import (
	"fmt"
	"log"
	"net/http"

	"endpoint-visualiser-server/pkg/clienthandler/rest"
	"endpoint-visualiser-server/pkg/keyboard"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/endpoints", rest.EndpointDiscoveryHandler).Methods("GET")
	router.HandleFunc("/ws", websocketsHandler)
	_, _, _ = keyboard.GetChar()

	fmt.Printf("\nListening on port 3031")
	log.Fatal(http.ListenAndServe(":3031", handlers.CORS(handlers.AllowedMethods([]string{"GET"}), handlers.AllowedOrigins([]string{"*"}))(router)))
}

func websocketsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	// register client
	clients[ws] = true
}

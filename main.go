package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

var (
	hub *Hub
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func connectionHandler(w http.ResponseWriter, r *http.Request) {

	connection, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		_, _ = fmt.Fprintf(w, "%+v\n", err)
		return
	}

	socketID, err := hub.Join(connection)

	fmt.Println(socketID)
}

func main() {

	hub = NewHub()

	fmt.Println(uuid.New().String()[:4] + uuid.New().String()[:4])

	router := mux.NewRouter()

	router.HandleFunc("/ws", connectionHandler)

	handler := cors.AllowAll().Handler(router)

	log.Println("> starting server on port: 5000")

	log.Fatal(http.ListenAndServe(":5000", handler))
}

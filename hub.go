package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub ...
type Hub struct {
	Sockets map[string]*Socket

	leaveChan  chan *Socket
	routerChan chan Message

	mutex sync.Mutex
}

// GetSockets ...
func (h *Hub) GetSockets() (sockets []*Socket) {

	for _, socket := range h.Sockets {
		sockets = append(sockets, socket)
	}

	return

}

// GetSocketCount ...
func (h *Hub) GetSocketCount() int {
	return len(h.Sockets)
}

// GetSocket ...
func (h *Hub) GetSocket(socketID string) (*Socket, bool) {

	socket, ok := h.Sockets[socketID]

	return socket, ok

}

// NewHub ...
func NewHub() *Hub {
	hub := Hub{}

	hub.Sockets = make(map[string]*Socket)
	hub.routerChan = make(chan Message)
	hub.leaveChan = make(chan *Socket)

	go hub.routerRoutine()
	go hub.routerRoutine()
	go hub.routerRoutine()
	go hub.leaveRoutine()

	return &hub
}

// SpawnRouterRoutines ...
func (h *Hub) SpawnRouterRoutines(count int) {
	for i := 0; i < count; i++ {
		go h.routerRoutine()
	}
}

// Join ...
func (h *Hub) Join(conn *websocket.Conn) (string, error) {

	socket := NewSocket(conn, h)

	h.mutex.Lock()

	h.Sockets[socket.ID] = socket

	h.mutex.Unlock()

	initPayload := struct {
		Socket *Socket   `json:"socket"`
		Users  []*Socket `json:"users"`
	}{
		Socket: socket,
		Users:  h.GetSockets(),
	}

	initMessage := Message{
		Source: "",
		Header: Header{
			Emit:            "broadcast",
			DestinationType: "socket",
			Destination:     socket.ID,
			Event:           "INIT",
		},
		Payload: initPayload,
	}

	newMessage := Message{
		Source: socket.ID,
		Header: Header{
			Emit:            "emit",
			DestinationType: "hub",
			Event:           "JOIN_USER",
		},
		Payload: socket,
	}

	h.routerChan <- newMessage
	h.routerChan <- initMessage

	return socket.ID, nil

}

func (h *Hub) routerRoutine() {

	// Used to relay messages accross hub as per the Header

	fmt.Println("hub router spawned")

	for {
		select {
		case msg := <-h.routerChan:

			if msg.Header.DestinationType == "hub" {

				for _, socket := range h.Sockets {

					if msg.Header.Emit == "emit" && socket.ID == msg.Source {
						// If message type is emit, skip the current socket
						continue
					}

					socket.WriteChan <- msg
				}

			} else if msg.Header.DestinationType == "socket" {

				if socket, ok := h.Sockets[msg.Header.Destination]; ok {

					socket.WriteChan <- msg

					if msg.Header.Emit == "broadcast" {

						if sourceSocket, ok := h.GetSocket(msg.Source); ok {
							sourceSocket.WriteChan <- msg
						}

					}

				} else {
					log.Println("invalid destination")
				}

			}

		}
	}

}

func (h *Hub) leaveRoutine() {

	fmt.Println("Leave routine spawned")

	for {

		select {

		case socket := <-h.leaveChan:

			fmt.Println(socket.ID, "is leaving")

			h.mutex.Lock()

			delete(h.Sockets, socket.ID)

			h.mutex.Unlock()

			socket.QuitChan <- true

		}

	}

}

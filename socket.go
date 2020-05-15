package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Socket ...
type Socket struct {
	// ID ...
	ID string `json:"id"`

	Conn *websocket.Conn `json:"-"`
	Hub  *Hub            `json:"-"`

	Data map[string]interface{} `json:"data"`

	WriteChan chan Message `json:"-"`
	QuitChan  chan bool    `json:"-"`
}

func generateSocketID() string {
	return uuid.New().String()[:4] + uuid.New().String()[:4]
}

// NewSocket ...
func NewSocket(conn *websocket.Conn, hub *Hub) *Socket {

	socket := Socket{
		Conn:      conn,
		ID:        generateSocketID(),
		Hub:       hub,
		WriteChan: make(chan Message),
		QuitChan:  make(chan bool),
		Data:      make(map[string]interface{}),
	}

	go socket.writeRoutine()
	go socket.readRoutine()
	go socket.quitRoutine()

	return &socket
}

func (s *Socket) quitRoutine() {

	// Does the shutdown jobs when quit recieved

	<-s.QuitChan

	log.Println("Closing connection for", s.ID)
	s.Conn.Close()

	leaveMessage := Message{
		Source: s.ID,
		Header: Header{
			Emit:            "emit",
			DestinationType: "hub",
			Event:           "LEAVE_USER",
		},
		Payload: s,
	}

	s.Hub.routerChan <- leaveMessage

}

func (s *Socket) writeRoutine() {
	fmt.Println("spawn writer for", s.ID)
	for {
		select {
		case <-s.QuitChan:
			return
		case message := <-s.WriteChan:
			// fmt.Println("writing to", s.ID, message)

			if err := s.Conn.WriteJSON(message); err != nil {
				log.Println("err write", s.ID, err)
				s.Hub.leaveChan <- s
				return
			}

		}
	}
}

func (s *Socket) readRoutine() {

	fmt.Println("spawn reader for", s.ID)

	for {
		var parseMessage Message

		parseMessage.Source = s.ID

		if err := s.Conn.ReadJSON(&parseMessage); err != nil {
			// log.Println("err read", s.ID, err)
			s.Hub.leaveChan <- s
			return
		}

		// fmt.Printf("Read message %+v\n", parseMessage)

		s.Hub.routerChan <- parseMessage
	}

}

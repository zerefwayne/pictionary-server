package main

import "encoding/json"

// Header ...
type Header struct {
	Emit            string `json:"emit"`
	DestinationType string `json:"destinationType"`
	Destination     string `json:"destination"`
	Event           string `json:"event"`
}

// Message ...
type Message struct {
	// Source ...
	Source string `json:"source"`
	// Header ...
	Header Header `json:"header"`
	// Payload ...
	Payload interface{} `json:"payload"`
}

// JSON returns the JSON form of Message
func (m *Message) JSON() string {
	message, _ := json.Marshal(m)
	return string(message)
}

package websocket

import (
	"log"
)

type Message struct {
	Type string `json:"type"`
}

type Hub struct {
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			log.Printf("Received client: %s", client.Conn.LocalAddr().String())
			h.Clients[client] = true
		case client := <-h.Unregister:
			log.Printf("Client %s dropped", client.Conn.LocalAddr().String())
			delete(h.Clients, client)
		case message := <-h.Broadcast:
			log.Println("Received broadcast: ", message.Type)
		}
	}
}

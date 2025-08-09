package websocket

import (
	"log"

	"github.com/dimfu/castaway/internal/store"
)

type Message struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type Hub struct {
	Clients    map[string]map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message
	Store      *store.Store
}

func NewHub(store *store.Store) *Hub {
	return &Hub{
		Clients:    make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message),
		Store:      store,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			connections := h.Clients[client.WhichStore]
			if connections == nil {
				connections = make(map[*Client]bool)
				h.Clients[client.WhichStore] = connections
			}
			h.Clients[client.WhichStore][client] = true
			log.Printf("Received client: %s", client.Conn.LocalAddr().String())
		case client := <-h.Unregister:
			log.Printf("Client %s dropped", client.Conn.LocalAddr().String())
			delete(h.Clients[client.WhichStore], client)
		case message := <-h.Broadcast:
			log.Println("Received broadcast: ", message.Type)
			if message.Type == "start_download" {
				clients := h.Clients[message.Payload]
				for client := range clients {
					client.Send <- Message{Type: message.Type, Payload: message.Payload}
				}
			}
		}
	}
}

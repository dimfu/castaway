package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn       *websocket.Conn
	Hub        *Hub
	Send       chan Message
	WhichStore string
}

func NewClient(conn *websocket.Conn, hub *Hub, s string) *Client {
	return &Client{
		Conn:       conn,
		Hub:        hub,
		Send:       make(chan Message),
		WhichStore: s,
	}
}

func (c *Client) Read() {
	// unregister client';s connection when they no longer connected to the websocket
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		var msg Message
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if closeErr, ok := err.(*websocket.CloseError); ok {
				if closeErr.Code == websocket.CloseGoingAway {
					break // ignore client disconnect error
				}
			}
			log.Println("Error: ", err)
			break
		}
		c.Hub.Broadcast <- msg
	}
}

func (c *Client) Write() {
	ticker := time.NewTimer(5 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			} else {
				err := c.Conn.WriteJSON(message)
				if err != nil {
					if closeErr, ok := err.(*websocket.CloseError); ok {
						if closeErr.Code == websocket.CloseGoingAway {
							break // ignore client disconnect error
						}
					}
					log.Println("Error: ", err)
					break
				}
			}
		case <-ticker.C:
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) Close() {
	close(c.Send)
}

package websocket

import (
	"encoding/json"
	"fmt"
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
		msgtype, data, err := c.Conn.ReadMessage()
		if err != nil {
			c.sendErr(fmt.Sprintf("Error while reading Blob from client: %s", err.Error()))
			break
		}

		if msgtype == websocket.BinaryMessage {
			r, err := c.Hub.Store.FindRegistry(c.WhichStore)
			if err != nil {
				break
			}

			if msgtype != websocket.BinaryMessage {
				continue
			}

			// write chunks to registry buffer, and wait for the client to download the buffered chunk
			// so that the chunk buffer can be cleared for the next chunk
			r.WriteChunks(data)

			log.Printf("Wrote bytes (total %d/%d)\n", len(data), cap(r.Buffer))
			if err := c.Conn.WriteJSON(Message{Type: "upload_chunk_success"}); err != nil {
				c.sendErr(fmt.Sprintf("Error while reading Blob from client: %s", err.Error()))
				break
			}
		} else {
			jsonMsg := new(Message)
			err := json.Unmarshal(data, jsonMsg)
			if err != nil {
				if closeErr, ok := err.(*websocket.CloseError); ok {
					if closeErr.Code == websocket.CloseGoingAway {
						break // ignore client disconnect error
					}
				}
				break
			}
			c.Hub.Broadcast <- *jsonMsg
		}
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
				c.sendErr("Unexpected error")
				return
			} else {
				err := c.Conn.WriteJSON(message)
				if err != nil {
					if closeErr, ok := err.(*websocket.CloseError); ok {
						if closeErr.Code == websocket.CloseGoingAway {
							break // ignore client disconnect error
						}
					}
					c.sendErr(fmt.Sprintf("Error while writing message into JSON: %s", err.Error()))
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

func (c *Client) sendErr(status string) {
	if msg, err := json.Marshal(Message{Type: "error", Payload: status}); err == nil {
		c.Conn.WriteMessage(websocket.TextMessage, msg)
	}
}

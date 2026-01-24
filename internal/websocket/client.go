package websocket

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn      *websocket.Conn
	Send      chan []byte
	UserID    string
	ChannelID string
	Hub       *Hub
}

func (c *Client) ReadPump(handle func(*Client, ClientEvent)) {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		var event ClientEvent
		if err := c.Conn.ReadJSON(&event); err != nil {
			log.Println("read error:", err)
			break
		}
		handle(c, event)
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()

	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

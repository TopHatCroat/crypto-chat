package protocol

import (
	"github.com/gorilla/websocket"
	"time"
	"log"
)

const (
	// Time allowed to write a message to the peer.
	WriteWait = 1 * time.Second
	// Time allowed to read the next pong message from the peer.
	PongWait = 6 * time.Second
	// Send pings to peer with this period. Must be less than PongWait.
	PingPeriod = (PongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Client struct {
	UserID int64
	Pool   *ClientPool
	Conn   *websocket.Conn
	Send   chan Message
}

func (c *Client) ReceiveContent() {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(PongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	for {
		status := &StatusUpdate{}
		err := c.Conn.ReadJSON(status)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		log.Printf("websocket: %s", status.Message)
	}
}

func (c *Client) SendContent() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		c.Conn.Close()
		ticker.Stop()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteJSON(StatusUpdate{Message:"not working bro"}) //websocket.CloseMessage
			}

			err := c.Conn.WriteJSON(message)
			if err != nil {
				c.Conn.WriteJSON(StatusUpdate{Message:"not working bro 2"})
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.Conn.WriteJSON(&StatusUpdate{Message:"ping"}); err != nil {
				return
			}
		}
	}
}

type ClientPool struct {
	clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Message    chan *Message
}

func (cp *ClientPool) Start() {
	for {
		select {
		case client := <-cp.Register:
			cp.clients[client] = true
		case client := <-cp.Unregister:
			if _, ok := cp.clients[client]; ok {
				delete(cp.clients, client)
				close(client.Send)
			}
		case message := <-cp.Message:
			if client := cp.findClient(message.Receiver); client != nil {
				client.Send <- *message
			}
		}
	}
}

func NewClientPool() *ClientPool {
	return &ClientPool{
		Message:    make(chan *Message, 10),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (cp *ClientPool) findClient(id int64) *Client {
	for client := range cp.clients {
		if client.UserID == id {
			return client
		}
	}
	return nil
}

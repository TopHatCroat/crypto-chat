package protocol

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	WriteWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	PongWait = 6 * time.Second
	// Send pings to peer with this period. Must be less than PongWait.
	PingPeriod = (PongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Client struct {
	ID   int64
	Pool *ClientPool
	Conn *websocket.Conn
	Send chan MessageData
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
				c.Conn.WriteJSON(StatusUpdate{Message: "not working bro"}) //websocket.CloseMessage
			}

			err := c.Conn.WriteJSON(message)
			if err != nil {
				log.Printf("sdn err: %s \n", err)
				c.Conn.WriteJSON(StatusUpdate{Message: "not working bro 2"})
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

type ClientPool struct {
	clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Message    chan *MessageData
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
			if client := cp.findClient(message.RecieverID); client != nil {
				client.Send <- *message
			}
		}
	}
}

func NewClientPool() *ClientPool {
	return &ClientPool{
		Message:    make(chan *MessageData, 10),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (cp *ClientPool) findClient(id int64) *Client {
	for client := range cp.clients {
		log.Println(client.ID, id)
		if client.ID == id {
			return client
		}
	}
	return nil
}

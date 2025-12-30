package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development
		// In production, you should check the origin
		return true
	},
}

// Connection is a wrapper around websocket.Conn
type Connection struct {
	Ws   *websocket.Conn
	Send chan []byte
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Ws.Close()
	}()

	c.Conn.Ws.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.Ws.SetReadLimit(maxMessageSize)
	c.Conn.Ws.SetPongHandler(func(string) error {
		c.Conn.Ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.Ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var message Message
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Handle different message types
		switch message.Type {
		case "subscribe":
			// Subscribe to a chat
			if chatID := message.ChatID; chatID != "" {
				c.Mu.Lock()
				c.Chats[chatID] = true
				c.Mu.Unlock()
				log.Printf("Client %s subscribed to chat %s", c.UserID, chatID)
			}
			if groupID := message.GroupID; groupID != "" {
				c.Mu.Lock()
				c.Chats[groupID] = true
				c.Mu.Unlock()
				log.Printf("Client %s subscribed to group %s", c.UserID, groupID)
			}
		case "unsubscribe":
			// Unsubscribe from a chat
			if chatID := message.ChatID; chatID != "" {
				c.Mu.Lock()
				delete(c.Chats, chatID)
				c.Mu.Unlock()
				log.Printf("Client %s unsubscribed from chat %s", c.UserID, chatID)
			}
			if groupID := message.GroupID; groupID != "" {
				c.Mu.Lock()
				delete(c.Chats, groupID)
				c.Mu.Unlock()
				log.Printf("Client %s unsubscribed from group %s", c.UserID, groupID)
			}
		case "typing":
			// Broadcast typing indicator
			c.Hub.Broadcast <- &message
		default:
			// Broadcast other messages
			c.Hub.Broadcast <- &message
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.Ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.Ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.Ws.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.Ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.Ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}


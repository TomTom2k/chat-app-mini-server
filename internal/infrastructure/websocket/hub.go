package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/TomTom2k/chat-app/server/internal/domain"
)

// Client represents a WebSocket client
type Client struct {
	Hub      *Hub
	Conn     *Connection
	UserID   string
	Send     chan []byte
	Chats    map[string]bool // Track which chats this client is subscribed to
	Mu       sync.RWMutex
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[string]*Client // userID -> Client

	// Inbound messages from the clients
	Broadcast chan *Message

	// Register requests from the clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// User repository for updating online status
	UserRepo domain.UserRepository

	mu sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type      string                 `json:"type"`      // message, typing, online, offline
	ChatID    string                 `json:"chatId,omitempty"`
	GroupID   string                 `json:"groupId,omitempty"`
	SenderID  string                 `json:"senderId,omitempty"`
	Content   string                 `json:"content,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
}

func NewHub(userRepo domain.UserRepository) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		Broadcast:  make(chan *Message, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		UserRepo:   userRepo,
		mu:         sync.RWMutex{},
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			
			// Update user online status
			user, err := h.UserRepo.GetByID(client.UserID)
			if err == nil {
				user.Online = true
				h.UserRepo.UpdateUser(user)
			}
			
			// Broadcast online status to friends
			h.broadcastOnlineStatus(client.UserID, true)
			
			log.Printf("Client registered: %s (Total: %d)", client.UserID, len(h.clients))

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()
			
			// Update user offline status
			user, err := h.UserRepo.GetByID(client.UserID)
			if err == nil {
				user.Online = false
				h.UserRepo.UpdateUser(user)
			}
			
			// Broadcast offline status to friends
			h.broadcastOnlineStatus(client.UserID, false)
			
			log.Printf("Client unregistered: %s (Total: %d)", client.UserID, len(h.clients))

		case message := <-h.Broadcast:
			h.handleBroadcast(message)
		}
	}
}

func (h *Hub) handleBroadcast(message *Message) {
	switch message.Type {
	case "message":
		h.broadcastToChat(message)
	case "typing":
		h.broadcastToChat(message)
	case "online", "offline":
		h.broadcastToFriends(message)
	default:
		h.broadcastToAll(message)
	}
}

func (h *Hub) broadcastToChat(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

		// Send to all clients subscribed to this chat
		for _, client := range h.clients {
			client.Mu.RLock()
			chatID := message.ChatID
			if chatID == "" {
				chatID = message.GroupID
			}
			subscribed := client.Chats[chatID]
			client.Mu.RUnlock()

		if subscribed {
			select {
			case client.Send <- h.messageToBytes(message):
			default:
				close(client.Send)
				delete(h.clients, client.UserID)
			}
		}
	}
}

func (h *Hub) broadcastToFriends(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Get user's friends
	user, err := h.UserRepo.GetByID(message.SenderID)
	if err != nil {
		return
	}

	// Send to all online friends
	for _, friendID := range user.Friends {
		if client, ok := h.clients[friendID]; ok {
			select {
			case client.Send <- h.messageToBytes(message):
			default:
				close(client.Send)
				delete(h.clients, friendID)
			}
		}
	}
}

func (h *Hub) broadcastToAll(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.Send <- h.messageToBytes(message):
		default:
			close(client.Send)
			delete(h.clients, client.UserID)
		}
	}
}

func (h *Hub) broadcastOnlineStatus(userID string, online bool) {
	_, err := h.UserRepo.GetByID(userID)
	if err != nil {
		return
	}

	statusType := "offline"
	if online {
		statusType = "online"
	}

	message := &Message{
		Type:     statusType,
		SenderID: userID,
		Data: map[string]interface{}{
			"userId": userID,
			"online": online,
		},
	}

	h.broadcastToFriends(message)
}

func (h *Hub) messageToBytes(message *Message) []byte {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return nil
	}
	return data
}

func (h *Hub) GetClient(userID string) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[userID]
}

func (h *Hub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}


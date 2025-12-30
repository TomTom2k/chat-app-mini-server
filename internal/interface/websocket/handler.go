package websocket

import (
	"log"
	"net/http"
	"strings"

	"github.com/TomTom2k/chat-app/server/internal/config"
	ws "github.com/TomTom2k/chat-app/server/internal/infrastructure/websocket"
	"github.com/TomTom2k/chat-app/server/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	Hub    *ws.Hub
	Config *config.Config
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get token from query parameter or Authorization header
	var token string
	
	// Try query parameter first (for WebSocket)
	if tokenParam := c.Query("token"); tokenParam != "" {
		token = tokenParam
	} else {
		// Try Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}
	}

	if token == "" {
		log.Printf("WebSocket: Token required")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}

	// Validate token
	claims, err := jwt.ValidateToken(token, h.Config.JWTSecret)
	if err != nil {
		log.Printf("WebSocket: Invalid token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	userID := claims.UserID
	log.Printf("WebSocket: User %s connecting", userID)

	// Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket: Upgrade error: %v", err)
		return
	}

	log.Printf("WebSocket: User %s connected successfully", userID)

	// Create connection wrapper
	wsConn := &ws.Connection{
		Ws:   conn,
		Send: make(chan []byte, 256),
	}

	// Create client
	client := &ws.Client{
		Hub:    h.Hub,
		Conn:   wsConn,
		UserID: userID,
		Send:   make(chan []byte, 256),
		Chats:  make(map[string]bool),
	}

	// Register client
	h.Hub.Register <- client

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}



package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/TomTom2k/chat-app/server/internal/infrastructure/websocket"
	"github.com/TomTom2k/chat-app/server/internal/usecase"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	ChatUseCase usecase.ChatUseCase
	Hub         *websocket.Hub
}

// GetChats godoc
// @Summary      Lấy danh sách chats
// @Description  Lấy danh sách tất cả chats của user hiện tại
// @Tags         Chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /chats [get]
func (h *ChatHandler) GetChats(c *gin.Context) {
	userID, _ := c.Get("userID")

	chats, err := h.ChatUseCase.GetChats(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chats)
}

// GetChat godoc
// @Summary      Lấy thông tin một chat
// @Description  Lấy thông tin chi tiết của một chat
// @Tags         Chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chatId  path  string  true  "Chat ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /chats/{chatId} [get]
func (h *ChatHandler) GetChat(c *gin.Context) {
	chatID := c.Param("chatId")
	userID, _ := c.Get("userID")

	chat, err := h.ChatUseCase.GetChat(chatID, userID.(string))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chat)
}

// CreateChat godoc
// @Summary      Tạo chat mới
// @Description  Tạo một chat mới với user khác
// @Tags         Chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object true "Create Chat Request" example({"userId":"507f1f77bcf86cd799439012"})
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /chats [post]
func (h *ChatHandler) CreateChat(c *gin.Context) {
	type Req struct {
		UserID string `json:"userId" binding:"required" example:"507f1f77bcf86cd799439012"`
	}
	var req Req

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	result, err := h.ChatUseCase.CreateChat(userID.(string), req.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetMessages godoc
// @Summary      Lấy danh sách messages trong chat
// @Description  Lấy tất cả messages trong một chat
// @Tags         Chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chatId  path  string  true  "Chat ID"
// @Success      200  {array}   map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /chats/{chatId}/messages [get]
func (h *ChatHandler) GetMessages(c *gin.Context) {
	chatID := c.Param("chatId")
	userID, _ := c.Get("userID")

	messages, err := h.ChatUseCase.GetMessages(chatID, userID.(string))
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// SendMessage godoc
// @Summary      Gửi message trong chat
// @Description  Gửi một message mới trong chat
// @Tags         Chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chatId  path  string  true  "Chat ID"
// @Param        request body object true "Send Message Request" example({"content":"Tin nhắn mới"})
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /chats/{chatId}/messages [post]
func (h *ChatHandler) SendMessage(c *gin.Context) {
	type Req struct {
		Content string `json:"content" binding:"required" example:"Tin nhắn mới"`
	}
	var req Req

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chatID := c.Param("chatId")
	userID, _ := c.Get("userID")

	result, err := h.ChatUseCase.SendMessage(chatID, userID.(string), req.Content)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast message via WebSocket
	if h.Hub != nil {
		message := &websocket.Message{
			Type:      "message",
			ChatID:    chatID,
			SenderID:  userID.(string),
			Content:   req.Content,
			Timestamp: time.Now().Format(time.RFC3339),
			Data: map[string]interface{}{
				"id":       result["id"],
				"sender":   result["sender"],
				"senderId": result["senderId"],
				"content":  result["content"],
				"time":     result["time"],
			},
		}
		select {
		case h.Hub.Broadcast <- message:
			// Message queued for broadcast
		default:
			// Channel is full, skip broadcast
		}
	}

	c.JSON(http.StatusOK, result)
}


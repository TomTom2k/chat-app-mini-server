package http

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TomTom2k/chat-app/server/internal/domain"
	"github.com/TomTom2k/chat-app/server/internal/domain/entity"
	"github.com/TomTom2k/chat-app/server/internal/infrastructure/websocket"
	"github.com/TomTom2k/chat-app/server/internal/usecase"
	"github.com/gin-gonic/gin"
)

type ConversationHandler struct {
	ConversationUseCase usecase.ConversationUseCase
	Hub                 *websocket.Hub
	MessageRepo         domain.MessageRepository
}

// GetConversations godoc
// @Summary      Lấy danh sách conversations
// @Description  Lấy danh sách tất cả conversations (chats và groups) của user hiện tại
// @Tags         Conversations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /conversations [get]
func (h *ConversationHandler) GetConversations(c *gin.Context) {
	userID, _ := c.Get("userID")

	conversations, err := h.ConversationUseCase.GetConversations(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conversations)
}

// GetConversation godoc
// @Summary      Lấy thông tin một conversation
// @Description  Lấy thông tin chi tiết của một conversation
// @Tags         Conversations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        conversationId  path  string  true  "Conversation ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /conversations/{conversationId} [get]
func (h *ConversationHandler) GetConversation(c *gin.Context) {
	conversationID := c.Param("conversationId")
	userID, _ := c.Get("userID")

	conversation, err := h.ConversationUseCase.GetConversation(conversationID, userID.(string))
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

	c.JSON(http.StatusOK, conversation)
}

// CreateDirectConversation godoc
// @Summary      Tạo direct conversation (chat đơn)
// @Description  Tạo một direct conversation mới với user khác
// @Tags         Conversations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object true "Create Direct Conversation Request" example({"userId":"507f1f77bcf86cd799439012"})
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /conversations/direct [post]
func (h *ConversationHandler) CreateDirectConversation(c *gin.Context) {
	type Req struct {
		UserID string `json:"userId" binding:"required" example:"507f1f77bcf86cd799439012"`
	}
	var req Req

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	result, err := h.ConversationUseCase.CreateDirectConversation(userID.(string), req.UserID)
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

// CreateGroupConversation godoc
// @Summary      Tạo group conversation
// @Description  Tạo một group conversation mới với các thành viên
// @Tags         Conversations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object true "Create Group Conversation Request" example({"name":"Executive Suite","description":"Nhóm quản lý","userIds":["507f1f77bcf86cd799439012","507f1f77bcf86cd799439013"]})
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /conversations/group [post]
func (h *ConversationHandler) CreateGroupConversation(c *gin.Context) {
	type Req struct {
		Name        string   `json:"name" binding:"required" example:"Executive Suite"`
		Description string   `json:"description" example:"Nhóm quản lý"`
		UserIDs     []string `json:"userIds" binding:"required" example:"[\"507f1f77bcf86cd799439012\",\"507f1f77bcf86cd799439013\"]"`
	}
	var req Req

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	result, err := h.ConversationUseCase.CreateGroupConversation(req.Name, req.Description, userID.(string), req.UserIDs)
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
// @Summary      Lấy danh sách messages trong conversation
// @Description  Lấy tất cả messages trong một conversation
// @Tags         Conversations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        conversationId  path  string  true  "Conversation ID"
// @Success      200  {array}   map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /conversations/{conversationId}/messages [get]
func (h *ConversationHandler) GetMessages(c *gin.Context) {
	conversationID := c.Param("conversationId")
	userID, _ := c.Get("userID")

	messages, err := h.ConversationUseCase.GetMessages(conversationID, userID.(string))
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
// @Summary      Gửi message trong conversation
// @Description  Gửi một message mới trong conversation (text, file, image, video, audio)
// @Tags         Conversations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        conversationId  path  string  true  "Conversation ID"
// @Param        request body object true "Send Message Request" example({"content":"Tin nhắn mới","replyToId":"","type":"text"})
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /conversations/{conversationId}/messages [post]
func (h *ConversationHandler) SendMessage(c *gin.Context) {
	type Req struct {
		Content    string                   `json:"content"`
		ReplyToID  string                   `json:"replyToId,omitempty"`
		Type       string                   `json:"type,omitempty"`
		Attachments []entity.MessageAttachment `json:"attachments,omitempty"`
	}
	var req Req

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate content or attachments
	if req.Content == "" && (req.Attachments == nil || len(req.Attachments) == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content or attachments required"})
		return
	}

	conversationID := c.Param("conversationId")
	userID, _ := c.Get("userID")

	messageType := entity.MessageTypeText
	if req.Type != "" {
		messageType = entity.MessageType(req.Type)
	}

	result, err := h.ConversationUseCase.SendMessage(conversationID, userID.(string), req.Content, req.ReplyToID, messageType, req.Attachments)
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
			ChatID:    conversationID,
			SenderID:  userID.(string),
			Content:   req.Content,
			Timestamp: time.Now().Format(time.RFC3339),
			Data:      result,
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

// UploadFile godoc
// @Summary      Upload file (image, video, file, audio)
// @Description  Upload file for conversation message with size limits
// @Tags         Conversations
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "File to upload"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /conversations/upload [post]
func (h *ConversationHandler) UploadFile(c *gin.Context) {
	// File size limits (in bytes)
	const (
		MaxImageSize = 10 * 1024 * 1024  // 10MB
		MaxVideoSize = 50 * 1024 * 1024  // 50MB
		MaxFileSize  = 20 * 1024 * 1024  // 20MB
		MaxAudioSize = 10 * 1024 * 1024  // 10MB
	)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Determine file type
	contentType := file.Header.Get("Content-Type")
	var fileType string
	var maxSize int64

	if strings.HasPrefix(contentType, "image/") {
		fileType = "image"
		maxSize = MaxImageSize
	} else if strings.HasPrefix(contentType, "video/") {
		fileType = "video"
		maxSize = MaxVideoSize
	} else if strings.HasPrefix(contentType, "audio/") {
		fileType = "audio"
		maxSize = MaxAudioSize
	} else {
		fileType = "file"
		maxSize = MaxFileSize
	}

	// Check file size
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("file size exceeds limit: %d bytes (max: %d bytes)", file.Size, maxSize),
		})
		return
	}

	// Create uploads directory if not exists
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
	filepath := filepath.Join(uploadDir, filename)

	// Save file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	// Return file info
	c.JSON(http.StatusOK, gin.H{
		"type":      fileType,
		"url":       "/uploads/" + filename,
		"file_name": file.Filename,
		"file_size": file.Size,
		"mime_type": contentType,
	})
}

// AddReaction godoc
// @Summary      Thêm reaction vào message
// @Description  Thêm emoji reaction vào một message
// @Tags         Conversations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        messageId  path  string  true  "Message ID"
// @Param        request body object true "Add Reaction Request" example({"emoji":"👍"})
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /conversations/messages/{messageId}/reactions [post]
func (h *ConversationHandler) AddReaction(c *gin.Context) {
	type Req struct {
		Emoji string `json:"emoji" binding:"required"`
	}
	var req Req

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	messageID := c.Param("messageId")
	userID, _ := c.Get("userID")

	err := h.ConversationUseCase.AddReaction(messageID, userID.(string), req.Emoji)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get message to find conversationID
		msg, err := h.MessageRepo.GetMessageByID(messageID)
		if err == nil {
			// Broadcast reaction via WebSocket
			if h.Hub != nil {
				message := &websocket.Message{
					Type:      "reaction",
					ChatID:    msg.GetConversationID(),
					SenderID:  userID.(string),
					Timestamp: time.Now().Format(time.RFC3339),
					Data: map[string]interface{}{
						"messageId": messageID,
						"emoji":     req.Emoji,
						"action":    "add",
					},
				}
				select {
				case h.Hub.Broadcast <- message:
				default:
				}
			}
		}

	c.JSON(http.StatusOK, gin.H{"message": "reaction added"})
}

// RemoveReaction godoc
// @Summary      Xóa reaction khỏi message
// @Description  Xóa emoji reaction khỏi một message
// @Tags         Conversations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        messageId  path  string  true  "Message ID"
// @Param        request body object true "Remove Reaction Request" example({"emoji":"👍"})
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /conversations/messages/{messageId}/reactions [delete]
func (h *ConversationHandler) RemoveReaction(c *gin.Context) {
	type Req struct {
		Emoji string `json:"emoji" binding:"required"`
	}
	var req Req

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	messageID := c.Param("messageId")
	userID, _ := c.Get("userID")

	err := h.ConversationUseCase.RemoveReaction(messageID, userID.(string), req.Emoji)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get message to find conversationID
		msg, err := h.MessageRepo.GetMessageByID(messageID)
		if err == nil {
			// Broadcast reaction removal via WebSocket
			if h.Hub != nil {
				message := &websocket.Message{
					Type:      "reaction",
					ChatID:    msg.GetConversationID(),
					SenderID:  userID.(string),
					Timestamp: time.Now().Format(time.RFC3339),
					Data: map[string]interface{}{
						"messageId": messageID,
						"emoji":     req.Emoji,
						"action":    "remove",
					},
				}
				select {
				case h.Hub.Broadcast <- message:
				default:
				}
			}
		}

	c.JSON(http.StatusOK, gin.H{"message": "reaction removed"})
}

// MarkAsRead godoc
// @Summary      Đánh dấu message đã đọc
// @Description  Đánh dấu một message là đã đọc bởi user hiện tại
// @Tags         Conversations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        messageId  path  string  true  "Message ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /conversations/messages/{messageId}/read [post]
func (h *ConversationHandler) MarkAsRead(c *gin.Context) {
	messageID := c.Param("messageId")
	userID, _ := c.Get("userID")

	err := h.ConversationUseCase.MarkMessageAsRead(messageID, userID.(string))
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast read receipt via WebSocket
	if h.Hub != nil {
		msg, err := h.MessageRepo.GetMessageByID(messageID)
		if err == nil {
			message := &websocket.Message{
				Type:      "read_receipt",
				ChatID:    msg.GetConversationID(),
				SenderID:  userID.(string),
				Timestamp: time.Now().Format(time.RFC3339),
				Data: map[string]interface{}{
					"messageId": messageID,
				},
			}
			select {
			case h.Hub.Broadcast <- message:
			default:
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}


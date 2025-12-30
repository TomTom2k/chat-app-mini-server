package http

import (
	"net/http"
	"strings"

	"github.com/TomTom2k/chat-app/server/internal/usecase"
	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	GroupUseCase usecase.GroupUseCase
}

// GetGroups godoc
// @Summary      Lấy danh sách groups
// @Description  Lấy danh sách tất cả groups mà user là thành viên
// @Tags         Groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /groups [get]
func (h *GroupHandler) GetGroups(c *gin.Context) {
	userID, _ := c.Get("userID")

	groups, err := h.GroupUseCase.GetGroups(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}

// GetGroup godoc
// @Summary      Lấy thông tin một group
// @Description  Lấy thông tin chi tiết của một group
// @Tags         Groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        groupId  path  string  true  "Group ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /groups/{groupId} [get]
func (h *GroupHandler) GetGroup(c *gin.Context) {
	groupID := c.Param("groupId")
	userID, _ := c.Get("userID")

	group, err := h.GroupUseCase.GetGroup(groupID, userID.(string))
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

	c.JSON(http.StatusOK, group)
}

// CreateGroup godoc
// @Summary      Tạo group mới
// @Description  Tạo một group mới với các thành viên
// @Tags         Groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object true "Create Group Request" example({"name":"Executive Suite","description":"Nhóm quản lý","userIds":["507f1f77bcf86cd799439012","507f1f77bcf86cd799439013"]})
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /groups [post]
func (h *GroupHandler) CreateGroup(c *gin.Context) {
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

	result, err := h.GroupUseCase.CreateGroup(req.Name, req.Description, userID.(string), req.UserIDs)
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
// @Summary      Lấy danh sách messages trong group
// @Description  Lấy tất cả messages trong một group
// @Tags         Groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        groupId  path  string  true  "Group ID"
// @Success      200  {array}   map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /groups/{groupId}/messages [get]
func (h *GroupHandler) GetMessages(c *gin.Context) {
	groupID := c.Param("groupId")
	userID, _ := c.Get("userID")

	messages, err := h.GroupUseCase.GetMessages(groupID, userID.(string))
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
// @Summary      Gửi message trong group
// @Description  Gửi một message mới trong group
// @Tags         Groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        groupId  path  string  true  "Group ID"
// @Param        request body object true "Send Message Request" example({"content":"Tin nhắn mới trong nhóm"})
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /groups/{groupId}/messages [post]
func (h *GroupHandler) SendMessage(c *gin.Context) {
	type Req struct {
		Content string `json:"content" binding:"required" example:"Tin nhắn mới trong nhóm"`
	}
	var req Req

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	groupID := c.Param("groupId")
	userID, _ := c.Get("userID")

	result, err := h.GroupUseCase.SendMessage(groupID, userID.(string), req.Content)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}


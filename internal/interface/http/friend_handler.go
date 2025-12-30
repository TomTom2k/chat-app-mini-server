package http

import (
	"net/http"
	"strings"

	"github.com/TomTom2k/chat-app/server/internal/usecase"
	"github.com/gin-gonic/gin"
)

type FriendHandler struct {
	FriendUseCase usecase.FriendUseCase
}

// GetFriends godoc
// @Summary      Lấy danh sách bạn bè
// @Description  Lấy danh sách tất cả bạn bè của user hiện tại
// @Tags         Friends
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /friends [get]
func (h *FriendHandler) GetFriends(c *gin.Context) {
	userID, _ := c.Get("userID")

	friends, err := h.FriendUseCase.GetFriends(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, friends)
}

// AddFriend godoc
// @Summary      Thêm bạn bè
// @Description  Thêm một user vào danh sách bạn bè
// @Tags         Friends
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object true "Add Friend Request" example({"userId":"507f1f77bcf86cd799439012"})
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /friends [post]
func (h *FriendHandler) AddFriend(c *gin.Context) {
	type Req struct {
		UserID string `json:"userId" binding:"required" example:"507f1f77bcf86cd799439012"`
	}
	var req Req

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	err := h.FriendUseCase.AddFriend(userID.(string), req.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend added successfully"})
}

// DeleteFriend godoc
// @Summary      Xóa bạn bè
// @Description  Xóa một user khỏi danh sách bạn bè
// @Tags         Friends
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        friendId  path  string  true  "Friend ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /friends/{friendId} [delete]
func (h *FriendHandler) DeleteFriend(c *gin.Context) {
	friendID := c.Param("friendId")
	userID, _ := c.Get("userID")

	err := h.FriendUseCase.DeleteFriend(friendID, userID.(string))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend removed successfully"})
}

// SearchUsers godoc
// @Summary      Tìm kiếm users
// @Description  Tìm kiếm users theo tên hoặc email
// @Tags         Friends
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        q  query  string  true  "Search query"
// @Success      200  {array}   map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/search [get]
func (h *FriendHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	users, err := h.FriendUseCase.SearchUsers(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}


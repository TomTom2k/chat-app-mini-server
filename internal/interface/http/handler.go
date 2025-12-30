package http

import (
	"net/http"
	"strings"

	"github.com/TomTom2k/chat-app/server/internal/domain/entity"
	"github.com/TomTom2k/chat-app/server/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	UserUseCase usecase.UserUseCase
}

// Register godoc
// @Summary      Đăng ký user mới
// @Description  Tạo tài khoản user mới
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body object true "Register Request" example({"name":"Nguyễn Văn A","email":"user@example.com","password":"password123"})
// @Success      200  {object}  usecase.RegisterResult
// @Failure      400  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	type Req struct {
		Name     string `json:"name" validate:"required,min=1,max=100" example:"Nguyễn Văn A"`
		Email    string `json:"email" validate:"required,email" example:"user@example.com"`
		Password string `json:"password" validate:"required,min=6" example:"password123"`
	}
	var req Req

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, err.Error())
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": strings.Join(errors, ", ")})
		return
	}

	result, err := h.UserUseCase.Register(entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		// Check if it's a duplicate user error
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Login godoc
// @Summary      Đăng nhập
// @Description  Đăng nhập với email và password
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body object true "Login Request" example({"email":"user@example.com","password":"password123"})
// @Success      200  {object}  usecase.RegisterResult
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	type Req struct {
		Email    string `json:"email" validate:"required,email" example:"user@example.com"`
		Password string `json:"password" validate:"required" example:"password123"`
	}
	var req Req

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, err.Error())
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": strings.Join(errors, ", ")})
		return
	}

	result, err := h.UserUseCase.Login(req.Email, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "invalid credentials") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetMe godoc
// @Summary      Lấy thông tin user hiện tại
// @Description  Lấy thông tin của user đang đăng nhập
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  entity.User
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /auth/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	user, err := h.UserUseCase.GetMe(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

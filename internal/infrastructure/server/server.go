package server

import (
	"log"

	"github.com/TomTom2k/chat-app/server/internal/config"
	"github.com/TomTom2k/chat-app/server/internal/infrastructure/di"
	"github.com/TomTom2k/chat-app/server/internal/interface/http"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	router *gin.Engine
	config *config.Config
}

func NewServer(cfg *config.Config, container *di.Container) *Server {
	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Setup middleware
	setupMiddleware(router, cfg)

	// Setup routes
	setupRoutes(router, container)

	return &Server{
		router: router,
		config: cfg,
	}
}

func setupMiddleware(router *gin.Engine, cfg *config.Config) {
	// CORS middleware
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	// Request logging (only in development)
	if cfg.Environment == "development" {
		router.Use(gin.Logger())
	}

	// Recovery middleware
	router.Use(gin.Recovery())
}

func setupRoutes(router *gin.Engine, container *di.Container) {
	// Health check
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Chat App API is running",
			"status":  "ok",
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := router.Group("/api")
	{
		setupAuthRoutes(api, container)
		setupChatRoutes(api, container)
		setupFriendRoutes(api, container)
		setupGroupRoutes(api, container)
		setupUserRoutes(api, container)
	}
}

func setupAuthRoutes(api *gin.RouterGroup, container *di.Container) {
	auth := api.Group("/auth")
	{
		auth.POST("/register", container.UserHandler.Register)
		auth.POST("/login", container.UserHandler.Login)
		auth.GET("/me", http.AuthMiddleware(container.Config), container.UserHandler.GetMe)
	}
}

func setupChatRoutes(api *gin.RouterGroup, container *di.Container) {
	chats := api.Group("/chats")
	chats.Use(http.AuthMiddleware(container.Config))
	{
		chats.GET("", container.ChatHandler.GetChats)
		chats.POST("", container.ChatHandler.CreateChat)
		chats.GET("/:chatId", container.ChatHandler.GetChat)
		chats.GET("/:chatId/messages", container.ChatHandler.GetMessages)
		chats.POST("/:chatId/messages", container.ChatHandler.SendMessage)
	}
}

func setupFriendRoutes(api *gin.RouterGroup, container *di.Container) {
	friends := api.Group("/friends")
	friends.Use(http.AuthMiddleware(container.Config))
	{
		friends.GET("", container.FriendHandler.GetFriends)
		friends.POST("", container.FriendHandler.AddFriend)
		friends.DELETE("/:friendId", container.FriendHandler.DeleteFriend)
	}

	users := api.Group("/users")
	users.Use(http.AuthMiddleware(container.Config))
	{
		users.GET("/search", container.FriendHandler.SearchUsers)
	}
}

func setupGroupRoutes(api *gin.RouterGroup, container *di.Container) {
	groups := api.Group("/groups")
	groups.Use(http.AuthMiddleware(container.Config))
	{
		groups.GET("", container.GroupHandler.GetGroups)
		groups.POST("", container.GroupHandler.CreateGroup)
		groups.GET("/:groupId", container.GroupHandler.GetGroup)
		groups.GET("/:groupId/messages", container.GroupHandler.GetMessages)
		groups.POST("/:groupId/messages", container.GroupHandler.SendMessage)
	}
}

func setupUserRoutes(api *gin.RouterGroup, container *di.Container) {
	// User routes can be added here if needed
}

func (s *Server) Start() error {
	addr := ":" + s.config.ServerPort
	log.Printf("Server starting on %s", addr)
	return s.router.Run(addr)
}


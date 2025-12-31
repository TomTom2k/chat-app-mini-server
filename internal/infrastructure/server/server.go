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

	// Static file serving for uploads
	router.Static("/uploads", "./uploads")

	// API routes
	api := router.Group("/api")
	{
		setupAuthRoutes(api, container)
		setupConversationRoutes(api, container)
		setupFriendRoutes(api, container)
		setupUserRoutes(api, container)
		setupWebSocketRoutes(router, container)
	}
}

func setupWebSocketRoutes(router *gin.Engine, container *di.Container) {
	// WebSocket route doesn't use AuthMiddleware because we handle auth in the handler
	// to support token in query parameter
	router.GET("/ws", container.WebSocketHandler.HandleWebSocket)
}

func setupAuthRoutes(api *gin.RouterGroup, container *di.Container) {
	auth := api.Group("/auth")
	{
		auth.POST("/register", container.UserHandler.Register)
		auth.POST("/login", container.UserHandler.Login)
		auth.GET("/me", http.AuthMiddleware(container.Config), container.UserHandler.GetMe)
	}
}

func setupConversationRoutes(api *gin.RouterGroup, container *di.Container) {
	conversations := api.Group("/conversations")
	conversations.Use(http.AuthMiddleware(container.Config))
	{
		conversations.GET("", container.ConversationHandler.GetConversations)
		conversations.POST("/direct", container.ConversationHandler.CreateDirectConversation)
		conversations.POST("/group", container.ConversationHandler.CreateGroupConversation)
		conversations.GET("/:conversationId", container.ConversationHandler.GetConversation)
		conversations.GET("/:conversationId/messages", container.ConversationHandler.GetMessages)
		conversations.POST("/:conversationId/messages", container.ConversationHandler.SendMessage)
		conversations.POST("/upload", container.ConversationHandler.UploadFile)
		conversations.POST("/messages/:messageId/reactions", container.ConversationHandler.AddReaction)
		conversations.DELETE("/messages/:messageId/reactions", container.ConversationHandler.RemoveReaction)
		conversations.POST("/messages/:messageId/read", container.ConversationHandler.MarkAsRead)
	}
	
	// Keep backward compatibility with /chats routes
	chats := api.Group("/chats")
	chats.Use(http.AuthMiddleware(container.Config))
	{
		chats.GET("", container.ConversationHandler.GetConversations)
		chats.POST("", container.ConversationHandler.CreateDirectConversation)
		chats.GET("/:chatId", func(c *gin.Context) {
			c.Params[0].Key = "conversationId"
			container.ConversationHandler.GetConversation(c)
		})
		chats.GET("/:chatId/messages", func(c *gin.Context) {
			c.Params[0].Key = "conversationId"
			container.ConversationHandler.GetMessages(c)
		})
		chats.POST("/:chatId/messages", func(c *gin.Context) {
			c.Params[0].Key = "conversationId"
			container.ConversationHandler.SendMessage(c)
		})
		chats.POST("/upload", container.ConversationHandler.UploadFile)
		chats.POST("/messages/:messageId/reactions", container.ConversationHandler.AddReaction)
		chats.DELETE("/messages/:messageId/reactions", container.ConversationHandler.RemoveReaction)
		chats.POST("/messages/:messageId/read", container.ConversationHandler.MarkAsRead)
	}
}

func setupFriendRoutes(api *gin.RouterGroup, container *di.Container) {
	friends := api.Group("/friends")
	friends.Use(http.AuthMiddleware(container.Config))
	{
		friends.GET("", container.FriendHandler.GetFriends)
		friends.POST("", container.FriendHandler.AddFriend)
		friends.DELETE("/:friendId", container.FriendHandler.DeleteFriend)
		
		// Friend requests
		friends.GET("/requests/pending", container.FriendHandler.GetPendingRequests)
		friends.GET("/requests/sent", container.FriendHandler.GetSentRequests)
		friends.POST("/requests/:requestId/accept", container.FriendHandler.AcceptFriendRequest)
		friends.POST("/requests/:requestId/reject", container.FriendHandler.RejectFriendRequest)
	}

	users := api.Group("/users")
	users.Use(http.AuthMiddleware(container.Config))
	{
		users.GET("/search", container.FriendHandler.SearchUsers)
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


package di

import (
	"github.com/TomTom2k/chat-app/server/internal/config"
	"github.com/TomTom2k/chat-app/server/internal/domain"
	"github.com/TomTom2k/chat-app/server/internal/infrastructure/repository"
	"github.com/TomTom2k/chat-app/server/internal/infrastructure/websocket"
	"github.com/TomTom2k/chat-app/server/internal/interface/http"
	wsHandler "github.com/TomTom2k/chat-app/server/internal/interface/websocket"
	"github.com/TomTom2k/chat-app/server/internal/usecase"
)

// Container holds all dependencies
type Container struct {
	Config              *config.Config
	UserRepository      domain.UserRepository
	ConversationRepository domain.ConversationRepository
	MessageRepository   domain.MessageRepository
	FriendRepository    domain.FriendRepository
	
	UserUseCase         *usecase.UserUseCase
	ConversationUseCase *usecase.ConversationUseCase
	FriendUseCase       *usecase.FriendUseCase
	
	UserHandler         *http.UserHandler
	ConversationHandler *http.ConversationHandler
	FriendHandler       *http.FriendHandler
	
	Hub                 *websocket.Hub
	WebSocketHandler    *wsHandler.WebSocketHandler
}

// NewContainer initializes all dependencies
func NewContainer(cfg *config.Config) *Container {
	// Initialize repositories
	userRepo := repository.NewUserRepository()
	conversationRepo := repository.NewConversationRepository()
	messageRepo := repository.NewMessageRepository()
	friendRepo := repository.NewFriendRepository()

	// Initialize usecases
	userUseCase := &usecase.UserUseCase{
		Repo:      userRepo,
		JWTSecret: cfg.JWTSecret,
	}

	// Initialize WebSocket Hub first (needed by use cases)
	hub := websocket.NewHub(userRepo, conversationRepo, messageRepo)
	go hub.Run()

	conversationUseCase := &usecase.ConversationUseCase{
		ConversationRepo: conversationRepo,
		UserRepo:         userRepo,
		MessageRepo:      messageRepo,
		Hub:              hub,
	}

	friendUseCase := &usecase.FriendUseCase{
		UserRepo: userRepo,
		Hub:      hub,
	}

	// Initialize handlers
	userHandler := &http.UserHandler{
		UserUseCase: *userUseCase,
	}

	conversationHandler := &http.ConversationHandler{
		ConversationUseCase: *conversationUseCase,
		Hub:                 hub,
		MessageRepo:         messageRepo,
	}

	friendHandler := &http.FriendHandler{
		FriendUseCase: *friendUseCase,
	}

	// Initialize WebSocket Handler
	wsHandler := &wsHandler.WebSocketHandler{
		Hub:    hub,
		Config: cfg,
	}

	return &Container{
		Config:                cfg,
		UserRepository:        userRepo,
		ConversationRepository: conversationRepo,
		MessageRepository:     messageRepo,
		FriendRepository:      friendRepo,
		UserUseCase:           userUseCase,
		ConversationUseCase:    conversationUseCase,
		FriendUseCase:          friendUseCase,
		UserHandler:            userHandler,
		ConversationHandler:    conversationHandler,
		FriendHandler:          friendHandler,
		Hub:                    hub,
		WebSocketHandler:       wsHandler,
	}
}


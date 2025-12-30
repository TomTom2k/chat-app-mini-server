package di

import (
	"github.com/TomTom2k/chat-app/server/internal/config"
	"github.com/TomTom2k/chat-app/server/internal/domain"
	"github.com/TomTom2k/chat-app/server/internal/infrastructure/repository"
	"github.com/TomTom2k/chat-app/server/internal/interface/http"
	"github.com/TomTom2k/chat-app/server/internal/usecase"
)

// Container holds all dependencies
type Container struct {
	Config              *config.Config
	UserRepository      domain.UserRepository
	ChatRepository      domain.ChatRepository
	MessageRepository   domain.MessageRepository
	FriendRepository    domain.FriendRepository
	GroupRepository     domain.GroupRepository
	GroupMemberRepository domain.GroupMemberRepository
	
	UserUseCase         *usecase.UserUseCase
	ChatUseCase         *usecase.ChatUseCase
	FriendUseCase       *usecase.FriendUseCase
	GroupUseCase        *usecase.GroupUseCase
	
	UserHandler         *http.UserHandler
	ChatHandler         *http.ChatHandler
	FriendHandler       *http.FriendHandler
	GroupHandler        *http.GroupHandler
}

// NewContainer initializes all dependencies
func NewContainer(cfg *config.Config) *Container {
	// Initialize repositories
	userRepo := repository.NewUserRepository()
	chatRepo := repository.NewChatRepository()
	messageRepo := repository.NewMessageRepository()
	friendRepo := repository.NewFriendRepository()
	groupRepo := repository.NewGroupRepository()
	groupMemberRepo := repository.NewGroupMemberRepository()

	// Initialize usecases
	userUseCase := &usecase.UserUseCase{
		Repo:      userRepo,
		JWTSecret: cfg.JWTSecret,
	}

	chatUseCase := &usecase.ChatUseCase{
		ChatRepo:    chatRepo,
		UserRepo:    userRepo,
		MessageRepo: messageRepo,
	}

	friendUseCase := &usecase.FriendUseCase{
		UserRepo: userRepo,
	}

	groupUseCase := &usecase.GroupUseCase{
		GroupRepo:       groupRepo,
		GroupMemberRepo: groupMemberRepo,
		UserRepo:        userRepo,
		MessageRepo:     messageRepo,
	}

	// Initialize handlers
	userHandler := &http.UserHandler{
		UserUseCase: *userUseCase,
	}

	chatHandler := &http.ChatHandler{
		ChatUseCase: *chatUseCase,
	}

	friendHandler := &http.FriendHandler{
		FriendUseCase: *friendUseCase,
	}

	groupHandler := &http.GroupHandler{
		GroupUseCase: *groupUseCase,
	}

	return &Container{
		Config:              cfg,
		UserRepository:      userRepo,
		ChatRepository:      chatRepo,
		MessageRepository:   messageRepo,
		FriendRepository:    friendRepo,
		GroupRepository:     groupRepo,
		GroupMemberRepository: groupMemberRepo,
		UserUseCase:         userUseCase,
		ChatUseCase:         chatUseCase,
		FriendUseCase:       friendUseCase,
		GroupUseCase:        groupUseCase,
		UserHandler:         userHandler,
		ChatHandler:         chatHandler,
		FriendHandler:       friendHandler,
		GroupHandler:        groupHandler,
	}
}


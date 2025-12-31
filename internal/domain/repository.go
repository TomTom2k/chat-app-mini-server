package domain

import "github.com/TomTom2k/chat-app/server/internal/domain/entity"

type UserRepository interface {
	CreateUser(user entity.User) error
	GetByEmail(email string) (entity.User, error)
	GetByUsername(username string) (entity.User, error)
	GetByID(id string) (entity.User, error)
	SearchUsers(query string) ([]entity.User, error)
	UpdateUser(user entity.User) error
	AddFriend(userID1, userID2 string) error
	RemoveFriend(userID1, userID2 string) error
	AddSentRequest(userID, targetUserID string) error
	RemoveSentRequest(userID, targetUserID string) error
	AddPendingRequest(userID, senderUserID string) error
	RemovePendingRequest(userID, senderUserID string) error
	GetUsersByIDs(userIDs []string) ([]entity.User, error)
}

type ConversationRepository interface {
	CreateConversation(conversation entity.Conversation) error
	GetConversationByID(conversationID string) (entity.Conversation, error)
	GetConversationsByUserID(userID string) ([]entity.Conversation, error)
	GetDirectConversationByUserIDs(userID1, userID2 string) (entity.Conversation, error)
	UpdateConversation(conversation entity.Conversation) error
	AddMember(conversationID, userID, role string) error
	RemoveMember(conversationID, userID string) error
}

type MessageRepository interface {
	CreateMessage(message entity.Message) error
	GetMessagesByConversationID(conversationID string) ([]entity.Message, error)
	GetMessageByID(messageID string) (entity.Message, error)
	UpdateMessage(message entity.Message) error
	AddReaction(messageID, userID, emoji string) error
	RemoveReaction(messageID, userID, emoji string) error
	MarkAsRead(messageID, userID string) error
	MarkAsDelivered(messageID string) error
}

type FriendRepository interface {
	CreateFriend(friend entity.Friend) error
	GetFriendsByUserID(userID string) ([]entity.Friend, error)
	GetFriendByID(friendID string) (entity.Friend, error)
	GetFriendByUserIDs(userID1, userID2 string) (entity.Friend, error)
	GetPendingRequestsByUserID(userID string) ([]entity.Friend, error) // Lời mời nhận được (userID là UserID2)
	GetSentRequestsByUserID(userID string) ([]entity.Friend, error)     // Lời mời đã gửi (userID là UserID1)
	DeleteFriend(friendID string) error
	UpdateFriend(friend entity.Friend) error
}


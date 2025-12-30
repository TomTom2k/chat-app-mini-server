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

type ChatRepository interface {
	CreateChat(chat entity.Chat) error
	GetChatByID(chatID string) (entity.Chat, error)
	GetChatsByUserID(userID string) ([]entity.Chat, error)
	GetChatByUserIDs(userID1, userID2 string) (entity.Chat, error)
	UpdateChat(chat entity.Chat) error
}

type MessageRepository interface {
	CreateMessage(message entity.Message) error
	GetMessagesByChatID(chatID string) ([]entity.Message, error)
	GetMessagesByGroupID(groupID string) ([]entity.Message, error)
	GetMessageByID(messageID string) (entity.Message, error)
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

type GroupRepository interface {
	CreateGroup(group entity.Group) error
	GetGroupByID(groupID string) (entity.Group, error)
	GetGroupsByUserID(userID string) ([]entity.Group, error)
	UpdateGroup(group entity.Group) error
}

type GroupMemberRepository interface {
	CreateGroupMember(member entity.GroupMember) error
	GetGroupMembersByGroupID(groupID string) ([]entity.GroupMember, error)
	GetGroupMembersByUserID(userID string) ([]entity.GroupMember, error)
	DeleteGroupMember(groupID, userID string) error
}

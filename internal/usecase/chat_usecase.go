package usecase

import (
	"errors"
	"time"

	"github.com/TomTom2k/chat-app/server/internal/domain"
	"github.com/TomTom2k/chat-app/server/internal/domain/entity"
)

type ChatUseCase struct {
	ChatRepo    domain.ChatRepository
	UserRepo    domain.UserRepository
	MessageRepo domain.MessageRepository
	Hub         interface {
		IsUserOnline(userID string) bool
	}
}

func (uc *ChatUseCase) GetChats(userID string) ([]map[string]interface{}, error) {
	chats, err := uc.ChatRepo.GetChatsByUserID(userID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, chat := range chats {
		// Get the other user
		var otherUserID string
		if chat.UserID1 == userID {
			otherUserID = chat.UserID2
		} else {
			otherUserID = chat.UserID1
		}

		otherUser, err := uc.UserRepo.GetByID(otherUserID)
		if err != nil {
			continue
		}

		// Check online status from Hub if available, otherwise use database status
		isOnline := otherUser.Online
		if uc.Hub != nil {
			isOnline = uc.Hub.IsUserOnline(otherUserID)
		}

		chatData := map[string]interface{}{
			"id":          chat.ID,
			"name":        otherUser.FullName,
			"lastMessage": chat.LastMessage,
			"time":        chat.LastMessageTime,
			"unread":      chat.Unread,
			"online":      isOnline,
			"avatar":      otherUser.Avatar,
			"userId1":     chat.UserID1,
			"userId2":     chat.UserID2,
		}
		result = append(result, chatData)
	}

	return result, nil
}

func (uc *ChatUseCase) GetChat(chatID, userID string) (map[string]interface{}, error) {
	chat, err := uc.ChatRepo.GetChatByID(chatID)
	if err != nil {
		return nil, err
	}

	// Verify user is part of this chat
	if chat.UserID1 != userID && chat.UserID2 != userID {
		return nil, errors.New("unauthorized")
	}

	// Get the other user
	var otherUserID string
	if chat.UserID1 == userID {
		otherUserID = chat.UserID2
	} else {
		otherUserID = chat.UserID1
	}

	otherUser, err := uc.UserRepo.GetByID(otherUserID)
	if err != nil {
		return nil, err
	}

	// Check online status from Hub if available, otherwise use database status
	isOnline := otherUser.Online
	if uc.Hub != nil {
		isOnline = uc.Hub.IsUserOnline(otherUserID)
	}

	result := map[string]interface{}{
		"id":      chat.ID,
		"name":    otherUser.FullName,
		"online":  isOnline,
		"avatar":  otherUser.Avatar,
		"userId1": chat.UserID1,
		"userId2": chat.UserID2,
		"users": []map[string]interface{}{
			{
				"id":     otherUser.ID,
				"name":   otherUser.FullName,
				"email":  otherUser.Email,
				"avatar": otherUser.Avatar,
				"online": isOnline,
			},
		},
	}

	return result, nil
}

func (uc *ChatUseCase) CreateChat(userID1, userID2 string) (map[string]interface{}, error) {
	// Check if chat already exists
	existingChat, _ := uc.ChatRepo.GetChatByUserIDs(userID1, userID2)
	if existingChat.ID != "" {
		return map[string]interface{}{"id": existingChat.ID}, nil
	}

	// Verify both users exist
	user1, err := uc.UserRepo.GetByID(userID1)
	if err != nil {
		return nil, errors.New("user not found")
	}

	_, err = uc.UserRepo.GetByID(userID2)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if users are friends (accepted)
	isFriend := false
	for _, friendID := range user1.Friends {
		if friendID == userID2 {
			isFriend = true
			break
		}
	}

	if !isFriend {
		return nil, errors.New("you can only chat with accepted friends")
	}

	chat := entity.Chat{
		UserID1: userID1,
		UserID2: userID2,
		Unread:  0,
	}

	err = uc.ChatRepo.CreateChat(chat)
	if err != nil {
		return nil, err
	}

	// Get created chat
	createdChat, err := uc.ChatRepo.GetChatByUserIDs(userID1, userID2)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"id": createdChat.ID}, nil
}

func (uc *ChatUseCase) GetMessages(chatID, userID string) ([]map[string]interface{}, error) {
	// Verify user is part of this chat
	chat, err := uc.ChatRepo.GetChatByID(chatID)
	if err != nil {
		return nil, err
	}

	if chat.UserID1 != userID && chat.UserID2 != userID {
		return nil, errors.New("unauthorized")
	}

	messages, err := uc.MessageRepo.GetMessagesByChatID(chatID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, msg := range messages {
		sender, _ := uc.UserRepo.GetByID(msg.SenderID)
		result = append(result, map[string]interface{}{
			"id":       msg.ID,
			"sender":   sender.FullName,
			"senderId": msg.SenderID,
			"content":  msg.Content,
			"time":     msg.CreatedAt,
			"isMe":     msg.SenderID == userID,
		})
	}

	return result, nil
}

func (uc *ChatUseCase) SendMessage(chatID, senderID, content string) (map[string]interface{}, error) {
	// Verify user is part of this chat
	chat, err := uc.ChatRepo.GetChatByID(chatID)
	if err != nil {
		return nil, err
	}

	if chat.UserID1 != senderID && chat.UserID2 != senderID {
		return nil, errors.New("unauthorized")
	}

	message := entity.Message{
		ChatID:   chatID,
		SenderID: senderID,
		Content:  content,
	}

	err = uc.MessageRepo.CreateMessage(message)
	if err != nil {
		return nil, err
	}

	// Update chat last message
	now := time.Now()
	chat.LastMessage = content
	chat.LastMessageTime = &now
	if chat.UserID1 == senderID {
		chat.Unread++
	} else {
		chat.Unread++
	}
	uc.ChatRepo.UpdateChat(chat)

	// Get created message
	messages, _ := uc.MessageRepo.GetMessagesByChatID(chatID)
	if len(messages) > 0 {
		lastMsg := messages[len(messages)-1]
		sender, _ := uc.UserRepo.GetByID(senderID)
		return map[string]interface{}{
			"id":       lastMsg.ID,
			"sender":   sender.FullName,
			"senderId": lastMsg.SenderID,
			"content":  lastMsg.Content,
			"time":     lastMsg.CreatedAt,
		}, nil
	}

	return nil, errors.New("failed to create message")
}





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
		result = append(result, uc.messageToMap(msg, userID, sender))
	}

	return result, nil
}

func (uc *ChatUseCase) SendMessage(chatID, senderID, content string, replyToID string, messageType entity.MessageType, attachments []entity.MessageAttachment) (map[string]interface{}, error) {
	// Verify user is part of this chat
	chat, err := uc.ChatRepo.GetChatByID(chatID)
	if err != nil {
		return nil, err
	}

	if chat.UserID1 != senderID && chat.UserID2 != senderID {
		return nil, errors.New("unauthorized")
	}

	// Verify replyToID if provided
	if replyToID != "" {
		_, err := uc.MessageRepo.GetMessageByID(replyToID)
		if err != nil {
			return nil, errors.New("reply to message not found")
		}
	}

	if messageType == "" {
		messageType = entity.MessageTypeText
	}

	message := entity.Message{
		ChatID:      chatID,
		SenderID:    senderID,
		Type:        messageType,
		Content:     content,
		ReplyToID:   replyToID,
		Attachments: attachments,
		Reactions:   []entity.MessageReaction{},
		ReadReceipts: []entity.ReadReceipt{},
		Status:      entity.MessageStatusSent,
	}

	err = uc.MessageRepo.CreateMessage(message)
	if err != nil {
		return nil, err
	}

	// Update chat last message
	now := time.Now()
	lastMessageText := content
	if len(attachments) > 0 {
		switch attachments[0].Type {
		case "image":
			lastMessageText = "📷 Hình ảnh"
		case "video":
			lastMessageText = "🎥 Video"
		case "file":
			lastMessageText = "📎 " + attachments[0].FileName
		case "audio":
			lastMessageText = "🎤 Tin nhắn thoại"
		}
	}
	chat.LastMessage = lastMessageText
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
		return uc.messageToMap(lastMsg, senderID, sender), nil
	}

	return nil, errors.New("failed to create message")
}

func (uc *ChatUseCase) AddReaction(messageID, userID, emoji string) error {
	// Verify user has access to the message
	message, err := uc.MessageRepo.GetMessageByID(messageID)
	if err != nil {
		return err
	}

	// Verify user is part of the chat
	chat, err := uc.ChatRepo.GetChatByID(message.ChatID)
	if err != nil {
		return err
	}

	if chat.UserID1 != userID && chat.UserID2 != userID {
		return errors.New("unauthorized")
	}

	return uc.MessageRepo.AddReaction(messageID, userID, emoji)
}

func (uc *ChatUseCase) RemoveReaction(messageID, userID, emoji string) error {
	// Verify user has access to the message
	message, err := uc.MessageRepo.GetMessageByID(messageID)
	if err != nil {
		return err
	}

	// Verify user is part of the chat
	chat, err := uc.ChatRepo.GetChatByID(message.ChatID)
	if err != nil {
		return err
	}

	if chat.UserID1 != userID && chat.UserID2 != userID {
		return errors.New("unauthorized")
	}

	return uc.MessageRepo.RemoveReaction(messageID, userID, emoji)
}

func (uc *ChatUseCase) MarkMessageAsRead(messageID, userID string) error {
	// Verify user has access to the message
	message, err := uc.MessageRepo.GetMessageByID(messageID)
	if err != nil {
		return err
	}

	// Verify user is part of the chat
	chat, err := uc.ChatRepo.GetChatByID(message.ChatID)
	if err != nil {
		return err
	}

	if chat.UserID1 != userID && chat.UserID2 != userID {
		return errors.New("unauthorized")
	}

	// Don't mark own messages as read
	if message.SenderID == userID {
		return nil
	}

	return uc.MessageRepo.MarkAsRead(messageID, userID)
}

func (uc *ChatUseCase) messageToMap(msg entity.Message, currentUserID string, sender entity.User) map[string]interface{} {
	// Get reply message if exists
	var replyTo map[string]interface{} = nil
	if msg.ReplyToID != "" {
		replyMsg, err := uc.MessageRepo.GetMessageByID(msg.ReplyToID)
		if err == nil {
			replySender, _ := uc.UserRepo.GetByID(replyMsg.SenderID)
			replyTo = map[string]interface{}{
				"id":       replyMsg.ID,
				"content":  replyMsg.Content,
				"sender":   replySender.FullName,
				"senderId": replyMsg.SenderID,
			}
		}
	}

	// Get read receipts with user info
	readReceipts := make([]map[string]interface{}, 0)
	for _, receipt := range msg.ReadReceipts {
		user, _ := uc.UserRepo.GetByID(receipt.UserID)
		readReceipts = append(readReceipts, map[string]interface{}{
			"user_id": receipt.UserID,
			"user_name": user.FullName,
			"user_avatar": user.Avatar,
			"read_at": receipt.ReadAt,
		})
	}

	return map[string]interface{}{
		"id":            msg.ID,
		"sender":        sender.FullName,
		"senderId":      msg.SenderID,
		"type":          msg.Type,
		"content":       msg.Content,
		"replyTo":       replyTo,
		"attachments":   msg.Attachments,
		"reactions":     msg.Reactions,
		"readReceipts":  readReceipts,
		"status":        msg.Status,
		"time":          msg.CreatedAt,
		"isMe":          msg.SenderID == currentUserID,
	}
}





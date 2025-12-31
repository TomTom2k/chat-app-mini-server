package usecase

import (
	"errors"
	"time"

	"github.com/TomTom2k/chat-app/server/internal/domain"
	"github.com/TomTom2k/chat-app/server/internal/domain/entity"
)

type ConversationUseCase struct {
	ConversationRepo domain.ConversationRepository
	UserRepo         domain.UserRepository
	MessageRepo      domain.MessageRepository
	Hub              interface {
		IsUserOnline(userID string) bool
	}
}

func (uc *ConversationUseCase) GetConversations(userID string) ([]map[string]interface{}, error) {
	conversations, err := uc.ConversationRepo.GetConversationsByUserID(userID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, conv := range conversations {
		convData := map[string]interface{}{
			"id":          conv.ID,
			"type":        conv.Type,
			"name":        conv.Name,
			"lastMessage": conv.LastMessage,
			"time":        conv.LastMessageTime,
			"unread":      conv.Unread,
			"avatar":      conv.Avatar,
		}

		// For direct conversations, get the other user's info
		if conv.Type == entity.ConversationTypeDirect {
			var otherUserID string
			for _, member := range conv.Members {
				if member.UserID != userID {
					otherUserID = member.UserID
					break
				}
			}

			if otherUserID != "" {
				otherUser, err := uc.UserRepo.GetByID(otherUserID)
				if err == nil {
					isOnline := otherUser.Online
					if uc.Hub != nil {
						isOnline = uc.Hub.IsUserOnline(otherUserID)
					}
					convData["name"] = otherUser.FullName
					convData["avatar"] = otherUser.Avatar
					convData["online"] = isOnline
					convData["userId1"] = userID
					convData["userId2"] = otherUserID
				}
			}
		} else {
			// For group conversations
			convData["members"] = len(conv.Members)
			convData["online"] = false // Groups don't have online status
		}

		result = append(result, convData)
	}

	return result, nil
}

func (uc *ConversationUseCase) GetConversation(conversationID, userID string) (map[string]interface{}, error) {
	conv, err := uc.ConversationRepo.GetConversationByID(conversationID)
	if err != nil {
		return nil, err
	}

	// Verify user is a member
	isMember := false
	for _, member := range conv.Members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		return nil, errors.New("unauthorized")
	}

	result := map[string]interface{}{
		"id":      conv.ID,
		"type":    conv.Type,
		"name":    conv.Name,
		"avatar":  conv.Avatar,
		"members": len(conv.Members),
	}

	// For direct conversations, get the other user's info
	if conv.Type == entity.ConversationTypeDirect {
		var otherUserID string
		for _, member := range conv.Members {
			if member.UserID != userID {
				otherUserID = member.UserID
				break
			}
		}

		if otherUserID != "" {
			otherUser, err := uc.UserRepo.GetByID(otherUserID)
			if err == nil {
				isOnline := otherUser.Online
				if uc.Hub != nil {
					isOnline = uc.Hub.IsUserOnline(otherUserID)
				}
				result["name"] = otherUser.FullName
				result["avatar"] = otherUser.Avatar
				result["online"] = isOnline
				result["userId1"] = userID
				result["userId2"] = otherUserID
				result["users"] = []map[string]interface{}{
					{
						"id":     otherUser.ID,
						"name":   otherUser.FullName,
						"email":  otherUser.Email,
						"avatar": otherUser.Avatar,
						"online": isOnline,
					},
				}
			}
		}
	} else {
		// For group conversations, get all members
		membersList := make([]map[string]interface{}, 0)
		for _, member := range conv.Members {
			user, err := uc.UserRepo.GetByID(member.UserID)
			if err == nil {
				isOnline := user.Online
				if uc.Hub != nil {
					isOnline = uc.Hub.IsUserOnline(member.UserID)
				}
				membersList = append(membersList, map[string]interface{}{
					"id":     user.ID,
					"name":   user.FullName,
					"email":  user.Email,
					"avatar": user.Avatar,
					"online": isOnline,
					"role":   member.Role,
				})
			}
		}
		result["users"] = membersList
		result["online"] = false
	}

	return result, nil
}

func (uc *ConversationUseCase) CreateDirectConversation(userID1, userID2 string) (map[string]interface{}, error) {
	// Check if conversation already exists
	existingConv, _ := uc.ConversationRepo.GetDirectConversationByUserIDs(userID1, userID2)
	if existingConv.ID != "" {
		return map[string]interface{}{"id": existingConv.ID}, nil
	}

	// Verify both users exist
	user1, err := uc.UserRepo.GetByID(userID1)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user2, err := uc.UserRepo.GetByID(userID2)
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

	now := time.Now()
	conversation := entity.Conversation{
		Type:    entity.ConversationTypeDirect,
		Name:    user2.FullName, // Default name is other user's name
		Unread:  0,
		Members: []entity.ConversationMember{
			{UserID: userID1, JoinedAt: now},
			{UserID: userID2, JoinedAt: now},
		},
	}

	err = uc.ConversationRepo.CreateConversation(conversation)
	if err != nil {
		return nil, err
	}

	// Get created conversation
	createdConv, err := uc.ConversationRepo.GetDirectConversationByUserIDs(userID1, userID2)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"id": createdConv.ID}, nil
}

func (uc *ConversationUseCase) CreateGroupConversation(name, description, createdBy string, userIds []string) (map[string]interface{}, error) {
	// Verify all users exist
	for _, userID := range userIds {
		_, err := uc.UserRepo.GetByID(userID)
		if err != nil {
			return nil, errors.New("user not found: " + userID)
		}
	}

	now := time.Now()
	members := []entity.ConversationMember{
		{UserID: createdBy, Role: "admin", JoinedAt: now},
	}

	// Add other users as members
	for _, userID := range userIds {
		if userID != createdBy {
			members = append(members, entity.ConversationMember{
				UserID:   userID,
				Role:     "member",
				JoinedAt: now,
			})
		}
	}

	conversation := entity.Conversation{
		Type:        entity.ConversationTypeGroup,
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		Unread:      0,
		Members:     members,
	}

	err := uc.ConversationRepo.CreateConversation(conversation)
	if err != nil {
		return nil, err
	}

	// Return the created conversation
	return map[string]interface{}{
		"id":          conversation.ID,
		"name":        conversation.Name,
		"description": conversation.Description,
		"members":     len(members),
		"lastActive":  &now,
		"createdAt":   conversation.CreatedAt,
		"updatedAt":   conversation.UpdatedAt,
	}, nil
}

func (uc *ConversationUseCase) GetMessages(conversationID, userID string) ([]map[string]interface{}, error) {
	// Verify user is a member
	conv, err := uc.ConversationRepo.GetConversationByID(conversationID)
	if err != nil {
		return nil, err
	}

	isMember := false
	for _, member := range conv.Members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		return nil, errors.New("unauthorized")
	}

	messages, err := uc.MessageRepo.GetMessagesByConversationID(conversationID)
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

func (uc *ConversationUseCase) SendMessage(conversationID, senderID, content string, replyToID string, messageType entity.MessageType, attachments []entity.MessageAttachment) (map[string]interface{}, error) {
	// Verify user is a member
	conv, err := uc.ConversationRepo.GetConversationByID(conversationID)
	if err != nil {
		return nil, err
	}

	isMember := false
	for _, member := range conv.Members {
		if member.UserID == senderID {
			isMember = true
			break
		}
	}

	if !isMember {
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
		ConversationID: conversationID,
		SenderID:       senderID,
		Type:           messageType,
		Content:        content,
		ReplyToID:      replyToID,
		Attachments:    attachments,
		Reactions:      []entity.MessageReaction{},
		ReadReceipts:   []entity.ReadReceipt{},
		Status:         entity.MessageStatusSent,
	}

	err = uc.MessageRepo.CreateMessage(message)
	if err != nil {
		return nil, err
	}

	// Update conversation last message
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
	conv.LastMessage = lastMessageText
	conv.LastMessageTime = &now
	// Increment unread for other members
	for _, member := range conv.Members {
		if member.UserID != senderID {
			conv.Unread++
		}
	}
	uc.ConversationRepo.UpdateConversation(conv)

	// Get created message
	messages, _ := uc.MessageRepo.GetMessagesByConversationID(conversationID)
	if len(messages) > 0 {
		lastMsg := messages[len(messages)-1]
		sender, _ := uc.UserRepo.GetByID(senderID)
		return uc.messageToMap(lastMsg, senderID, sender), nil
	}

	return nil, errors.New("failed to create message")
}

func (uc *ConversationUseCase) AddReaction(messageID, userID, emoji string) error {
	// Verify user has access to the message
	message, err := uc.MessageRepo.GetMessageByID(messageID)
	if err != nil {
		return err
	}

	// Verify user is a member of the conversation
	conv, err := uc.ConversationRepo.GetConversationByID(message.GetConversationID())
	if err != nil {
		return err
	}

	isMember := false
	for _, member := range conv.Members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		return errors.New("unauthorized")
	}

	return uc.MessageRepo.AddReaction(messageID, userID, emoji)
}

func (uc *ConversationUseCase) RemoveReaction(messageID, userID, emoji string) error {
	// Verify user has access to the message
	message, err := uc.MessageRepo.GetMessageByID(messageID)
	if err != nil {
		return err
	}

	// Verify user is a member of the conversation
	conv, err := uc.ConversationRepo.GetConversationByID(message.GetConversationID())
	if err != nil {
		return err
	}

	isMember := false
	for _, member := range conv.Members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		return errors.New("unauthorized")
	}

	return uc.MessageRepo.RemoveReaction(messageID, userID, emoji)
}

func (uc *ConversationUseCase) MarkMessageAsRead(messageID, userID string) error {
	// Verify user has access to the message
	message, err := uc.MessageRepo.GetMessageByID(messageID)
	if err != nil {
		return err
	}

	// Verify user is a member of the conversation
	conv, err := uc.ConversationRepo.GetConversationByID(message.GetConversationID())
	if err != nil {
		return err
	}

	isMember := false
	for _, member := range conv.Members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		return errors.New("unauthorized")
	}

	// Don't mark own messages as read
	if message.SenderID == userID {
		return nil
	}

	return uc.MessageRepo.MarkAsRead(messageID, userID)
}

func (uc *ConversationUseCase) messageToMap(msg entity.Message, currentUserID string, sender entity.User) map[string]interface{} {
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
			"user_id":    receipt.UserID,
			"user_name":  user.FullName,
			"user_avatar": user.Avatar,
			"read_at":    receipt.ReadAt,
		})
	}

	return map[string]interface{}{
		"id":           msg.ID,
		"sender":       sender.FullName,
		"senderId":     msg.SenderID,
		"type":         msg.Type,
		"content":      msg.Content,
		"replyTo":      replyTo,
		"attachments":  msg.Attachments,
		"reactions":    msg.Reactions,
		"readReceipts": readReceipts,
		"status":       msg.Status,
		"time":         msg.CreatedAt,
		"isMe":         msg.SenderID == currentUserID,
	}
}


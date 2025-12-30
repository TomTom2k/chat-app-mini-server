package usecase

import (
	"errors"
	"time"

	"github.com/TomTom2k/chat-app/server/internal/domain"
	"github.com/TomTom2k/chat-app/server/internal/domain/entity"
)

type GroupUseCase struct {
	GroupRepo       domain.GroupRepository
	GroupMemberRepo domain.GroupMemberRepository
	UserRepo        domain.UserRepository
	MessageRepo     domain.MessageRepository
}

func (uc *GroupUseCase) GetGroups(userID string) ([]map[string]interface{}, error) {
	// Get all groups user is member of
	members, err := uc.GroupMemberRepo.GetGroupMembersByUserID(userID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, member := range members {
		group, err := uc.GroupRepo.GetGroupByID(member.GroupID)
		if err != nil {
			continue
		}

		// Get member count
		allMembers, _ := uc.GroupMemberRepo.GetGroupMembersByGroupID(group.ID)

		result = append(result, map[string]interface{}{
			"id":            group.ID,
			"name":          group.Name,
			"description":   group.Description,
			"members":       len(allMembers),
			"lastActive":    group.LastActive,
			"avatar":        group.Avatar,
			"memberAvatars": []string{}, // TODO: implement
			"createdAt":     group.CreatedAt,
			"updatedAt":     group.UpdatedAt,
		})
	}

	return result, nil
}

func (uc *GroupUseCase) GetGroup(groupID, userID string) (map[string]interface{}, error) {
	// Verify user is member
	members, err := uc.GroupMemberRepo.GetGroupMembersByGroupID(groupID)
	if err != nil {
		return nil, err
	}

	isMember := false
	for _, member := range members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		return nil, errors.New("unauthorized")
	}

	group, err := uc.GroupRepo.GetGroupByID(groupID)
	if err != nil {
		return nil, err
	}

	// Get all members with user details
	memberDetails := make([]map[string]interface{}, 0)
	for _, member := range members {
		user, _ := uc.UserRepo.GetByID(member.UserID)
		memberDetails = append(memberDetails, map[string]interface{}{
			"id":     user.ID,
			"name":   user.FullName,
			"email":  user.Email,
			"avatar": "",
			"online": false, // TODO: implement
		})
	}

	result := map[string]interface{}{
		"id":            group.ID,
		"name":          group.Name,
		"description":   group.Description,
		"members":       memberDetails,
		"memberCount":   len(members),
		"lastActive":    group.LastActive,
		"avatar":        group.Avatar,
		"memberAvatars": []string{}, // TODO: implement
		"createdAt":     group.CreatedAt,
		"updatedAt":     group.UpdatedAt,
	}

	return result, nil
}

func (uc *GroupUseCase) CreateGroup(name, description, createdBy string, userIds []string) (map[string]interface{}, error) {
	// Verify all users exist
	for _, userID := range userIds {
		_, err := uc.UserRepo.GetByID(userID)
		if err != nil {
			return nil, errors.New("user not found: " + userID)
		}
	}

	group := entity.Group{
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
	}

	err := uc.GroupRepo.CreateGroup(group)
	if err != nil {
		return nil, err
	}

	// Get created group
	createdGroup, err := uc.GroupRepo.GetGroupByID(group.ID)
	if err != nil {
		return nil, err
	}

	// Add creator as admin
	member := entity.GroupMember{
		GroupID: createdGroup.ID,
		UserID:  createdBy,
		Role:    "admin",
	}
	uc.GroupMemberRepo.CreateGroupMember(member)

	// Add other users as members
	for _, userID := range userIds {
		if userID != createdBy {
			member := entity.GroupMember{
				GroupID: createdGroup.ID,
				UserID:  userID,
				Role:    "member",
			}
			uc.GroupMemberRepo.CreateGroupMember(member)
		}
	}

	now := time.Now()
	return map[string]interface{}{
		"id":          createdGroup.ID,
		"name":        createdGroup.Name,
		"description": createdGroup.Description,
		"members":     len(userIds) + 1, // +1 for creator
		"lastActive":  &now,
		"createdAt":   createdGroup.CreatedAt,
		"updatedAt":   createdGroup.UpdatedAt,
	}, nil
}

func (uc *GroupUseCase) GetMessages(groupID, userID string) ([]map[string]interface{}, error) {
	// Verify user is member
	members, err := uc.GroupMemberRepo.GetGroupMembersByGroupID(groupID)
	if err != nil {
		return nil, err
	}

	isMember := false
	for _, member := range members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		return nil, errors.New("unauthorized")
	}

	messages, err := uc.MessageRepo.GetMessagesByGroupID(groupID)
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

func (uc *GroupUseCase) SendMessage(groupID, senderID, content string) (map[string]interface{}, error) {
	// Verify user is member
	members, err := uc.GroupMemberRepo.GetGroupMembersByGroupID(groupID)
	if err != nil {
		return nil, err
	}

	isMember := false
	for _, member := range members {
		if member.UserID == senderID {
			isMember = true
			break
		}
	}

	if !isMember {
		return nil, errors.New("unauthorized")
	}

	message := entity.Message{
		GroupID:  groupID,
		SenderID: senderID,
		Content:  content,
	}

	err = uc.MessageRepo.CreateMessage(message)
	if err != nil {
		return nil, err
	}

	// Update group last active
	group, _ := uc.GroupRepo.GetGroupByID(groupID)
	now := time.Now()
	group.LastActive = &now
	uc.GroupRepo.UpdateGroup(group)

	// Get created message
	messages, _ := uc.MessageRepo.GetMessagesByGroupID(groupID)
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

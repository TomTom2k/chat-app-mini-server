package usecase

import (
	"errors"

	"github.com/TomTom2k/chat-app/server/internal/domain"
	"github.com/TomTom2k/chat-app/server/internal/domain/entity"
)

type FriendUseCase struct {
	FriendRepo domain.FriendRepository
	UserRepo   domain.UserRepository
}

func (uc *FriendUseCase) GetFriends(userID string) ([]map[string]interface{}, error) {
	friends, err := uc.FriendRepo.GetFriendsByUserID(userID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, friend := range friends {
		var friendUserID string
		if friend.UserID1 == userID {
			friendUserID = friend.UserID2
		} else {
			friendUserID = friend.UserID1
		}

		friendUser, err := uc.UserRepo.GetByID(friendUserID)
		if err != nil {
			continue
		}

		result = append(result, map[string]interface{}{
			"id":     friendUser.ID,
			"name":   friendUser.FullName,
			"status": "Online", // TODO: implement status
			"avatar": "",
			"email":  friendUser.Email,
		})
	}

	return result, nil
}

func (uc *FriendUseCase) AddFriend(userID1, userID2 string) error {
	// Check if already friends
	existing, _ := uc.FriendRepo.GetFriendByUserIDs(userID1, userID2)
	if existing.ID != "" {
		if existing.Status == "accepted" {
			return errors.New("friendship already exists")
		}
		// Update existing pending request
		existing.Status = "accepted"
		return uc.FriendRepo.UpdateFriend(existing)
	}

	// Verify user exists
	_, err := uc.UserRepo.GetByID(userID2)
	if err != nil {
		return errors.New("user not found")
	}

	friend := entity.Friend{
		UserID1: userID1,
		UserID2: userID2,
		Status:  "accepted",
	}

	return uc.FriendRepo.CreateFriend(friend)
}

func (uc *FriendUseCase) DeleteFriend(friendID, userID string) error {
	friend, err := uc.FriendRepo.GetFriendByID(friendID)
	if err != nil {
		return err
	}

	// Verify user is part of this friendship
	if friend.UserID1 != userID && friend.UserID2 != userID {
		return errors.New("unauthorized")
	}

	return uc.FriendRepo.DeleteFriend(friendID)
}

func (uc *FriendUseCase) SearchUsers(query string) ([]map[string]interface{}, error) {
	users, err := uc.UserRepo.SearchUsers(query)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, user := range users {
		result = append(result, map[string]interface{}{
			"id":        user.ID,
			"name":      user.FullName,
			"email":     user.Email,
			"avatar":    "",
			"online":    false, // TODO: implement online status
			"createdAt": user.CreatedAt,
			"updatedAt": user.UpdatedAt,
		})
	}

	return result, nil
}





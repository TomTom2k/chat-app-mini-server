package usecase

import (
	"errors"

	"github.com/TomTom2k/chat-app/server/internal/domain"
)

type FriendUseCase struct {
	UserRepo domain.UserRepository
	Hub      interface {
		IsUserOnline(userID string) bool
	}
}

func (uc *FriendUseCase) GetFriends(userID string) ([]map[string]interface{}, error) {
	user, err := uc.UserRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if len(user.Friends) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Get all friend users
	friendUsers, err := uc.UserRepo.GetUsersByIDs(user.Friends)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, friendUser := range friendUsers {
		// Check online status from Hub if available, otherwise use database status
		isOnline := friendUser.Online
		if uc.Hub != nil {
			isOnline = uc.Hub.IsUserOnline(friendUser.ID)
		}
		
		status := "Offline"
		if isOnline {
			status = "Online"
		}
		
		result = append(result, map[string]interface{}{
			"id":     friendUser.ID,
			"name":   friendUser.FullName,
			"status": status,
			"online": isOnline,
			"avatar": friendUser.Avatar,
			"email":  friendUser.Email,
		})
	}

	return result, nil
}

func (uc *FriendUseCase) AddFriend(userID1, userID2 string) error {
	// Verify user exists
	_, err := uc.UserRepo.GetByID(userID2)
	if err != nil {
		return errors.New("user not found")
	}

	// Get current user to check existing relationships
	user1, err := uc.UserRepo.GetByID(userID1)
	if err != nil {
		return err
	}

	// Check if already friends
	for _, friendID := range user1.Friends {
		if friendID == userID2 {
			return errors.New("friendship already exists")
		}
	}

	// Check if already sent request
	for _, sentID := range user1.SentRequests {
		if sentID == userID2 {
			return errors.New("friend request already sent")
		}
	}

	// Check if there's a pending request from userID2
	for _, pendingID := range user1.PendingRequests {
		if pendingID == userID2 {
			// Accept the request: add to friends and remove from pending/sent
			err = uc.UserRepo.AddFriend(userID1, userID2)
			if err != nil {
				return err
			}
			// Remove from pending/sent requests
			err = uc.UserRepo.RemovePendingRequest(userID1, userID2)
			if err != nil {
				return err
			}
			err = uc.UserRepo.RemoveSentRequest(userID2, userID1)
			if err != nil {
				return err
			}
			return nil
		}
	}

	// Create new friend request
	return uc.UserRepo.AddSentRequest(userID1, userID2)
}

func (uc *FriendUseCase) DeleteFriend(friendID, userID string) error {
	// Verify friend exists
	_, err := uc.UserRepo.GetByID(friendID)
	if err != nil {
		return errors.New("friend not found")
	}

	// Verify user is friend
	user, err := uc.UserRepo.GetByID(userID)
	if err != nil {
		return err
	}

	isFriend := false
	for _, friendIDInList := range user.Friends {
		if friendIDInList == friendID {
			isFriend = true
			break
		}
	}

	if !isFriend {
		return errors.New("unauthorized")
	}

	return uc.UserRepo.RemoveFriend(userID, friendID)
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
			"avatar":    user.Avatar,
			"online":    user.Online,
			"createdAt": user.CreatedAt,
			"updatedAt": user.UpdatedAt,
		})
	}

	return result, nil
}

func (uc *FriendUseCase) GetPendingRequests(userID string) ([]map[string]interface{}, error) {
	user, err := uc.UserRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if len(user.PendingRequests) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Get all users who sent requests
	requestUsers, err := uc.UserRepo.GetUsersByIDs(user.PendingRequests)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, requestUser := range requestUsers {
		result = append(result, map[string]interface{}{
			"id":        requestUser.ID, // Use senderUserID as ID for accept/reject
			"userId":    requestUser.ID,
			"name":      requestUser.FullName,
			"email":     requestUser.Email,
			"avatar":    requestUser.Avatar,
			"createdAt": user.UpdatedAt, // Use updated_at as approximate time
		})
	}

	return result, nil
}

func (uc *FriendUseCase) GetSentRequests(userID string) ([]map[string]interface{}, error) {
	user, err := uc.UserRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if len(user.SentRequests) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Get all users who received requests
	requestUsers, err := uc.UserRepo.GetUsersByIDs(user.SentRequests)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, requestUser := range requestUsers {
		result = append(result, map[string]interface{}{
			"id":        requestUser.ID, // Use receiverUserID as ID
			"userId":    requestUser.ID,
			"name":      requestUser.FullName,
			"email":     requestUser.Email,
			"avatar":    requestUser.Avatar,
			"createdAt": user.UpdatedAt, // Use updated_at as approximate time
		})
	}

	return result, nil
}

func (uc *FriendUseCase) AcceptFriendRequest(senderUserID, userID string) error {
	user, err := uc.UserRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Verify this is a pending request
	isPending := false
	for _, pendingID := range user.PendingRequests {
		if pendingID == senderUserID {
			isPending = true
			break
		}
	}

	if !isPending {
		return errors.New("request is not pending")
	}

	// Accept: add to friends and remove from pending/sent
	err = uc.UserRepo.AddFriend(userID, senderUserID)
	if err != nil {
		return err
	}

	// Remove from pending/sent requests
	err = uc.UserRepo.RemovePendingRequest(userID, senderUserID)
	if err != nil {
		return err
	}
	err = uc.UserRepo.RemoveSentRequest(senderUserID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (uc *FriendUseCase) RejectFriendRequest(senderUserID, userID string) error {
	user, err := uc.UserRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Verify this is a pending request
	isPending := false
	for _, pendingID := range user.PendingRequests {
		if pendingID == senderUserID {
			isPending = true
			break
		}
	}

	if !isPending {
		return errors.New("request is not pending")
	}

	// Reject: remove from pending/sent requests
	err = uc.UserRepo.RemovePendingRequest(userID, senderUserID)
	if err != nil {
		return err
	}
	err = uc.UserRepo.RemoveSentRequest(senderUserID, userID)
	return err
}

package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/TomTom2k/chat-app/server/internal/domain"
	"github.com/TomTom2k/chat-app/server/internal/domain/entity"
	"github.com/TomTom2k/chat-app/server/internal/infrastructure/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository() domain.UserRepository {
	return &userRepository{
		collection: mongodb.OpenCollection("users"),
	}
}

func (r *userRepository) CreateUser(user entity.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user with email already exists
	existingUser, _ := r.GetByEmail(user.Email)
	if existingUser.Email != "" {
		return errors.New("user with this email already exists")
	}

	// Check if user with username already exists
	existingUser, _ = r.GetByUsername(user.Username)
	if existingUser.Username != "" {
		return errors.New("user with this username already exists")
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Generate ID (24 character hex string, similar to MongoDB ObjectID)
	if user.ID == "" {
		user.ID = generateID()
	}

	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *userRepository) GetByEmail(email string) (entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user entity.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.User{}, nil
		}
		return entity.User{}, err
	}
	return user, nil
}

func (r *userRepository) GetByUsername(username string) (entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user entity.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.User{}, nil
		}
		return entity.User{}, err
	}
	return user, nil
}

func (r *userRepository) GetByID(id string) (entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user entity.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.User{}, errors.New("user not found")
		}
		return entity.User{}, err
	}
	return user, nil
}

func (r *userRepository) SearchUsers(query string) ([]entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"email": bson.M{"$regex": query, "$options": "i"}},
			{"full_name": bson.M{"$regex": query, "$options": "i"}},
			{"username": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []entity.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	// Remove passwords from results
	for i := range users {
		users[i].Password = ""
	}

	return users, nil
}

func (r *userRepository) UpdateUser(user entity.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": user})
	return err
}

func (r *userRepository) AddFriend(userID1, userID2 string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()

	// Add userID2 to userID1's friends list
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID1},
		bson.M{
			"$addToSet": bson.M{"friends": userID2},
			"$set":      bson.M{"updated_at": now},
		},
	)
	if err != nil {
		return err
	}

	// Add userID1 to userID2's friends list
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID2},
		bson.M{
			"$addToSet": bson.M{"friends": userID1},
			"$set":      bson.M{"updated_at": now},
		},
	)
	return err
}

func (r *userRepository) RemoveFriend(userID1, userID2 string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()

	// Remove userID2 from userID1's friends list
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID1},
		bson.M{
			"$pull": bson.M{"friends": userID2},
			"$set":  bson.M{"updated_at": now},
		},
	)
	if err != nil {
		return err
	}

	// Remove userID1 from userID2's friends list
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID2},
		bson.M{
			"$pull": bson.M{"friends": userID1},
			"$set":  bson.M{"updated_at": now},
		},
	)
	return err
}

func (r *userRepository) AddSentRequest(userID, targetUserID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()

	// Add targetUserID to userID's sent_requests
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{
			"$addToSet": bson.M{"sent_requests": targetUserID},
			"$set":      bson.M{"updated_at": now},
		},
	)
	if err != nil {
		return err
	}

	// Add userID to targetUserID's pending_requests
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": targetUserID},
		bson.M{
			"$addToSet": bson.M{"pending_requests": userID},
			"$set":      bson.M{"updated_at": now},
		},
	)
	return err
}

func (r *userRepository) RemoveSentRequest(userID, targetUserID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()

	// Remove targetUserID from userID's sent_requests
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{
			"$pull": bson.M{"sent_requests": targetUserID},
			"$set":  bson.M{"updated_at": now},
		},
	)
	if err != nil {
		return err
	}

	// Remove userID from targetUserID's pending_requests
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": targetUserID},
		bson.M{
			"$pull": bson.M{"pending_requests": userID},
			"$set":  bson.M{"updated_at": now},
		},
	)
	return err
}

func (r *userRepository) AddPendingRequest(userID, senderUserID string) error {
	// This is handled by AddSentRequest, but kept for consistency
	return r.AddSentRequest(senderUserID, userID)
}

func (r *userRepository) RemovePendingRequest(userID, senderUserID string) error {
	// This is handled by RemoveSentRequest, but kept for consistency
	return r.RemoveSentRequest(senderUserID, userID)
}

func (r *userRepository) GetUsersByIDs(userIDs []string) ([]entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id": bson.M{"$in": userIDs},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []entity.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	// Remove passwords from results
	for i := range users {
		users[i].Password = ""
	}

	return users, nil
}

// generateID generates a 24-character hex string (similar to MongoDB ObjectID)
func generateID() string {
	bytes := make([]byte, 12)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}


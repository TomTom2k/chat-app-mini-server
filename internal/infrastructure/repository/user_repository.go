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

// generateID generates a 24-character hex string (similar to MongoDB ObjectID)
func generateID() string {
	bytes := make([]byte, 12)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}


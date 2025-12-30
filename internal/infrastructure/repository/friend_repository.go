package repository

import (
	"context"
	"errors"
	"time"

	"github.com/TomTom2k/chat-app/server/internal/domain"
	"github.com/TomTom2k/chat-app/server/internal/domain/entity"
	"github.com/TomTom2k/chat-app/server/internal/infrastructure/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type friendRepository struct {
	collection *mongo.Collection
}

func NewFriendRepository() domain.FriendRepository {
	return &friendRepository{
		collection: mongodb.OpenCollection("friends"),
	}
}

func (r *friendRepository) CreateFriend(friend entity.Friend) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	friend.CreatedAt = now
	friend.UpdatedAt = now

	if friend.ID == "" {
		friend.ID = generateID()
	}

	_, err := r.collection.InsertOne(ctx, friend)
	return err
}

func (r *friendRepository) GetFriendsByUserID(userID string) ([]entity.Friend, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"user_id_1": userID},
			{"user_id_2": userID},
		},
		"status": "accepted",
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var friends []entity.Friend
	if err := cursor.All(ctx, &friends); err != nil {
		return nil, err
	}

	return friends, nil
}

func (r *friendRepository) GetFriendByID(friendID string) (entity.Friend, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var friend entity.Friend
	err := r.collection.FindOne(ctx, bson.M{"_id": friendID}).Decode(&friend)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Friend{}, errors.New("friend not found")
		}
		return entity.Friend{}, err
	}
	return friend, nil
}

func (r *friendRepository) GetFriendByUserIDs(userID1, userID2 string) (entity.Friend, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"user_id_1": userID1, "user_id_2": userID2},
			{"user_id_1": userID2, "user_id_2": userID1},
		},
	}

	var friend entity.Friend
	err := r.collection.FindOne(ctx, filter).Decode(&friend)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Friend{}, nil
		}
		return entity.Friend{}, err
	}
	return friend, nil
}

func (r *friendRepository) DeleteFriend(friendID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": friendID})
	return err
}

func (r *friendRepository) UpdateFriend(friend entity.Friend) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	friend.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": friend.ID}, bson.M{"$set": friend})
	return err
}





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

type chatRepository struct {
	collection *mongo.Collection
}

func NewChatRepository() domain.ChatRepository {
	return &chatRepository{
		collection: mongodb.OpenCollection("chats"),
	}
}

func (r *chatRepository) CreateChat(chat entity.Chat) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	chat.CreatedAt = now
	chat.UpdatedAt = now

	if chat.ID == "" {
		chat.ID = generateID()
	}

	_, err := r.collection.InsertOne(ctx, chat)
	return err
}

func (r *chatRepository) GetChatByID(chatID string) (entity.Chat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var chat entity.Chat
	err := r.collection.FindOne(ctx, bson.M{"_id": chatID}).Decode(&chat)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Chat{}, errors.New("chat not found")
		}
		return entity.Chat{}, err
	}
	return chat, nil
}

func (r *chatRepository) GetChatsByUserID(userID string) ([]entity.Chat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"user_id_1": userID},
			{"user_id_2": userID},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var chats []entity.Chat
	if err := cursor.All(ctx, &chats); err != nil {
		return nil, err
	}

	return chats, nil
}

func (r *chatRepository) GetChatByUserIDs(userID1, userID2 string) (entity.Chat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"user_id_1": userID1, "user_id_2": userID2},
			{"user_id_1": userID2, "user_id_2": userID1},
		},
	}

	var chat entity.Chat
	err := r.collection.FindOne(ctx, filter).Decode(&chat)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Chat{}, nil
		}
		return entity.Chat{}, err
	}
	return chat, nil
}

func (r *chatRepository) UpdateChat(chat entity.Chat) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chat.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": chat.ID}, bson.M{"$set": chat})
	return err
}





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
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type messageRepository struct {
	collection *mongo.Collection
}

func NewMessageRepository() domain.MessageRepository {
	return &messageRepository{
		collection: mongodb.OpenCollection("messages"),
	}
}

func (r *messageRepository) CreateMessage(message entity.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message.CreatedAt = time.Now()
	if message.ID == "" {
		message.ID = generateID()
	}

	_, err := r.collection.InsertOne(ctx, message)
	return err
}

func (r *messageRepository) GetMessagesByChatID(chatID string) ([]entity.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"chat_id": chatID}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []entity.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *messageRepository) GetMessagesByGroupID(groupID string) ([]entity.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"group_id": groupID}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []entity.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *messageRepository) GetMessageByID(messageID string) (entity.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var message entity.Message
	err := r.collection.FindOne(ctx, bson.M{"_id": messageID}).Decode(&message)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Message{}, errors.New("message not found")
		}
		return entity.Message{}, err
	}
	return message, nil
}





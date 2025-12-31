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
	message.UpdatedAt = time.Now()
	if message.ID == "" {
		message.ID = generateID()
	}
	if message.Type == "" {
		message.Type = entity.MessageTypeText
	}
	if message.Status == "" {
		message.Status = entity.MessageStatusSent
	}
	if message.Reactions == nil {
		message.Reactions = []entity.MessageReaction{}
	}
	if message.ReadReceipts == nil {
		message.ReadReceipts = []entity.ReadReceipt{}
	}
	if message.Attachments == nil {
		message.Attachments = []entity.MessageAttachment{}
	}
	// Set ConversationID from ChatID or GroupID if not set (backward compatibility)
	if message.ConversationID == "" {
		if message.ChatID != "" {
			message.ConversationID = message.ChatID
		} else if message.GroupID != "" {
			message.ConversationID = message.GroupID
		}
	}

	_, err := r.collection.InsertOne(ctx, message)
	return err
}

func (r *messageRepository) GetMessagesByConversationID(conversationID string) ([]entity.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Support both old format (chat_id/group_id) and new format (conversation_id)
	filter := bson.M{
		"$or": []bson.M{
			{"conversation_id": conversationID},
			{"chat_id": conversationID},
			{"group_id": conversationID},
		},
	}
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

func (r *messageRepository) UpdateMessage(message entity.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": message.ID},
		bson.M{"$set": message},
	)
	return err
}

func (r *messageRepository) AddReaction(messageID, userID, emoji string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if reaction already exists
	var message entity.Message
	err := r.collection.FindOne(ctx, bson.M{"_id": messageID}).Decode(&message)
	if err != nil {
		return err
	}

	// Check if user already reacted with this emoji
	for _, reaction := range message.Reactions {
		if reaction.UserID == userID && reaction.Emoji == emoji {
			return nil // Already reacted
		}
	}

	// Add reaction
	reaction := entity.MessageReaction{
		UserID: userID,
		Emoji:  emoji,
	}
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": messageID},
		bson.M{
			"$push": bson.M{"reactions": reaction},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

func (r *messageRepository) RemoveReaction(messageID, userID, emoji string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": messageID},
		bson.M{
			"$pull": bson.M{"reactions": bson.M{"user_id": userID, "emoji": emoji}},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

func (r *messageRepository) MarkAsRead(messageID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if already read by this user
	var message entity.Message
	err := r.collection.FindOne(ctx, bson.M{"_id": messageID}).Decode(&message)
	if err != nil {
		return err
	}

	// Check if already in read receipts
	for _, receipt := range message.ReadReceipts {
		if receipt.UserID == userID {
			return nil // Already read
		}
	}

	// Add read receipt
	receipt := entity.ReadReceipt{
		UserID: userID,
		ReadAt: time.Now(),
	}
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": messageID},
		bson.M{
			"$push": bson.M{"read_receipts": receipt},
			"$set": bson.M{
				"status":     entity.MessageStatusRead,
				"updated_at": time.Now(),
			},
		},
	)
	return err
}

func (r *messageRepository) MarkAsDelivered(messageID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": messageID},
		bson.M{
			"$set": bson.M{
				"status":     entity.MessageStatusDelivered,
				"updated_at": time.Now(),
			},
		},
	)
	return err
}





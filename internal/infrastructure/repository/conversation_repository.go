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

type conversationRepository struct {
	collection *mongo.Collection
}

func NewConversationRepository() domain.ConversationRepository {
	return &conversationRepository{
		collection: mongodb.OpenCollection("conversations"),
	}
}

func (r *conversationRepository) CreateConversation(conversation entity.Conversation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	conversation.CreatedAt = now
	conversation.UpdatedAt = now
	if conversation.ID == "" {
		conversation.ID = generateID()
	}

	// Set joined_at for all members
	for i := range conversation.Members {
		if conversation.Members[i].JoinedAt.IsZero() {
			conversation.Members[i].JoinedAt = now
		}
	}

	_, err := r.collection.InsertOne(ctx, conversation)
	return err
}

func (r *conversationRepository) GetConversationByID(conversationID string) (entity.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var conversation entity.Conversation
	err := r.collection.FindOne(ctx, bson.M{"_id": conversationID}).Decode(&conversation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Conversation{}, errors.New("conversation not found")
		}
		return entity.Conversation{}, err
	}
	return conversation, nil
}

func (r *conversationRepository) GetConversationsByUserID(userID string) ([]entity.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find conversations where user is a member
	filter := bson.M{
		"members.user_id": userID,
	}
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var conversations []entity.Conversation
	if err := cursor.All(ctx, &conversations); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *conversationRepository) GetDirectConversationByUserIDs(userID1, userID2 string) (entity.Conversation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find direct conversation with exactly these 2 members
	filter := bson.M{
		"type": entity.ConversationTypeDirect,
		"members.user_id": bson.M{
			"$all": []string{userID1, userID2},
		},
		"members": bson.M{
			"$size": 2,
		},
	}

	var conversation entity.Conversation
	err := r.collection.FindOne(ctx, filter).Decode(&conversation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Conversation{}, errors.New("conversation not found")
		}
		return entity.Conversation{}, err
	}
	return conversation, nil
}

func (r *conversationRepository) UpdateConversation(conversation entity.Conversation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conversation.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": conversation.ID},
		bson.M{"$set": conversation},
	)
	return err
}

func (r *conversationRepository) AddMember(conversationID, userID, role string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	member := entity.ConversationMember{
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
	}

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": conversationID},
		bson.M{
			"$push": bson.M{"members": member},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

func (r *conversationRepository) RemoveMember(conversationID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": conversationID},
		bson.M{
			"$pull": bson.M{"members": bson.M{"user_id": userID}},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	return err
}


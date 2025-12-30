package repository

import (
	"context"
	"time"

	"github.com/TomTom2k/chat-app/server/internal/domain"
	"github.com/TomTom2k/chat-app/server/internal/domain/entity"
	"github.com/TomTom2k/chat-app/server/internal/infrastructure/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type groupMemberRepository struct {
	collection *mongo.Collection
}

func NewGroupMemberRepository() domain.GroupMemberRepository {
	return &groupMemberRepository{
		collection: mongodb.OpenCollection("group_members"),
	}
}

func (r *groupMemberRepository) CreateGroupMember(member entity.GroupMember) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	member.JoinedAt = time.Now()
	if member.ID == "" {
		member.ID = generateID()
	}

	_, err := r.collection.InsertOne(ctx, member)
	return err
}

func (r *groupMemberRepository) GetGroupMembersByGroupID(groupID string) ([]entity.GroupMember, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"group_id": groupID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []entity.GroupMember
	if err := cursor.All(ctx, &members); err != nil {
		return nil, err
	}

	return members, nil
}

func (r *groupMemberRepository) GetGroupMembersByUserID(userID string) ([]entity.GroupMember, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []entity.GroupMember
	if err := cursor.All(ctx, &members); err != nil {
		return nil, err
	}

	return members, nil
}

func (r *groupMemberRepository) DeleteGroupMember(groupID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"group_id": groupID, "user_id": userID})
	return err
}


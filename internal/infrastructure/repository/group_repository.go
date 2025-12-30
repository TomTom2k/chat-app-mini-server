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

type groupRepository struct {
	collection *mongo.Collection
}

func NewGroupRepository() domain.GroupRepository {
	return &groupRepository{
		collection: mongodb.OpenCollection("groups"),
	}
}

func (r *groupRepository) CreateGroup(group entity.Group) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	group.CreatedAt = now
	group.UpdatedAt = now

	if group.ID == "" {
		group.ID = generateID()
	}

	_, err := r.collection.InsertOne(ctx, group)
	return err
}

func (r *groupRepository) GetGroupByID(groupID string) (entity.Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var group entity.Group
	err := r.collection.FindOne(ctx, bson.M{"_id": groupID}).Decode(&group)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return entity.Group{}, errors.New("group not found")
		}
		return entity.Group{}, err
	}
	return group, nil
}

func (r *groupRepository) GetGroupsByUserID(userID string) ([]entity.Group, error) {
	// This will be joined with group_members collection
	// For now, we'll get all groups and filter in usecase
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groups []entity.Group
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func (r *groupRepository) UpdateGroup(group entity.Group) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	group.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": group.ID}, bson.M{"$set": group})
	return err
}





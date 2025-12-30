package entity

import "time"

type Group struct {
	ID            string    `json:"id" bson:"_id"`
	Name          string    `json:"name" bson:"name"`
	Description   string    `json:"description,omitempty" bson:"description,omitempty"`
	Avatar        string    `json:"avatar,omitempty" bson:"avatar,omitempty"`
	CreatedBy     string    `json:"created_by" bson:"created_by"`
	LastActive    *time.Time `json:"last_active,omitempty" bson:"last_active,omitempty"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" bson:"updated_at"`
}

type GroupMember struct {
	ID        string    `json:"id" bson:"_id"`
	GroupID   string    `json:"group_id" bson:"group_id"`
	UserID    string    `json:"user_id" bson:"user_id"`
	Role      string    `json:"role" bson:"role"` // "admin", "member"
	JoinedAt  time.Time `json:"joined_at" bson:"joined_at"`
}





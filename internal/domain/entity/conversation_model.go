package entity

import "time"

type ConversationType string

const (
	ConversationTypeDirect ConversationType = "direct" // Chat đơn (2 members)
	ConversationTypeGroup  ConversationType = "group"  // Chat nhóm (>2 members)
)

type ConversationMember struct {
	UserID    string    `json:"user_id" bson:"user_id"`
	Role      string    `json:"role,omitempty" bson:"role,omitempty"` // "admin", "member" (chỉ cho group)
	JoinedAt  time.Time `json:"joined_at" bson:"joined_at"`
}

type Conversation struct {
	ID              string               `json:"id" bson:"_id"`
	Type            ConversationType     `json:"type" bson:"type"` // "direct" or "group"
	Name            string               `json:"name,omitempty" bson:"name,omitempty"` // Tên nhóm (chỉ cho group), hoặc tên chat đơn
	Description     string               `json:"description,omitempty" bson:"description,omitempty"` // Mô tả (chỉ cho group)
	Avatar          string               `json:"avatar,omitempty" bson:"avatar,omitempty"`
	Members         []ConversationMember `json:"members" bson:"members"` // Danh sách thành viên
	LastMessage     string               `json:"last_message,omitempty" bson:"last_message,omitempty"`
	LastMessageTime *time.Time           `json:"last_message_time,omitempty" bson:"last_message_time,omitempty"`
	Unread          int                  `json:"unread" bson:"unread"`
	CreatedBy       string               `json:"created_by,omitempty" bson:"created_by,omitempty"` // Người tạo (chỉ cho group)
	CreatedAt       time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at" bson:"updated_at"`
}


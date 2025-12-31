package entity

import "time"

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeVideo MessageType = "video"
	MessageTypeFile  MessageType = "file"
	MessageTypeAudio MessageType = "audio"
)

type MessageStatus string

const (
	MessageStatusSent     MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead     MessageStatus = "read"
)

type MessageAttachment struct {
	Type     string `json:"type" bson:"type"`         // image, video, file, audio
	URL      string `json:"url" bson:"url"`
	FileName string `json:"file_name" bson:"file_name"`
	FileSize int64  `json:"file_size" bson:"file_size"` // in bytes
	MimeType string `json:"mime_type" bson:"mime_type"`
}

type MessageReaction struct {
	UserID string `json:"user_id" bson:"user_id"`
	Emoji  string `json:"emoji" bson:"emoji"` // 👍, ❤️, 😂, etc.
}

type ReadReceipt struct {
	UserID    string    `json:"user_id" bson:"user_id"`
	ReadAt    time.Time `json:"read_at" bson:"read_at"`
}

type Message struct {
	ID             string             `json:"id" bson:"_id"`
	ConversationID string             `json:"conversation_id" bson:"conversation_id"` // Mới: thay thế chat_id và group_id
	ChatID         string             `json:"chat_id,omitempty" bson:"chat_id,omitempty"` // Cũ: backward compatibility
	GroupID        string             `json:"group_id,omitempty" bson:"group_id,omitempty"` // Cũ: backward compatibility
	SenderID       string             `json:"sender_id" bson:"sender_id"`
	Type           MessageType        `json:"type" bson:"type"`
	Content        string             `json:"content" bson:"content"`
	ReplyToID      string             `json:"reply_to_id,omitempty" bson:"reply_to_id,omitempty"` // ID of message being replied to
	Attachments    []MessageAttachment `json:"attachments,omitempty" bson:"attachments,omitempty"`
	Reactions      []MessageReaction   `json:"reactions,omitempty" bson:"reactions,omitempty"`
	ReadReceipts   []ReadReceipt      `json:"read_receipts,omitempty" bson:"read_receipts,omitempty"`
	Status         MessageStatus      `json:"status" bson:"status"`
	CreatedAt      time.Time          `json:"time" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

// GetConversationID returns conversation_id, or falls back to chat_id/group_id for backward compatibility
func (m *Message) GetConversationID() string {
	if m.ConversationID != "" {
		return m.ConversationID
	}
	if m.ChatID != "" {
		return m.ChatID
	}
	return m.GroupID
}





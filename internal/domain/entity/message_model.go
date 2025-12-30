package entity

import "time"

type Message struct {
	ID        string    `json:"id" bson:"_id"`
	ChatID    string    `json:"chat_id" bson:"chat_id"`
	GroupID   string    `json:"group_id,omitempty" bson:"group_id,omitempty"`
	SenderID  string    `json:"sender_id" bson:"sender_id"`
	Content   string    `json:"content" bson:"content"`
	CreatedAt time.Time `json:"time" bson:"created_at"`
}





package entity

import "time"

type Chat struct {
	ID         string    `json:"id" bson:"_id"`
	UserID1    string    `json:"user_id_1" bson:"user_id_1"`
	UserID2    string    `json:"user_id_2" bson:"user_id_2"`
	LastMessage string   `json:"last_message,omitempty" bson:"last_message,omitempty"`
	LastMessageTime *time.Time `json:"time,omitempty" bson:"last_message_time,omitempty"`
	Unread     int       `json:"unread" bson:"unread"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
}





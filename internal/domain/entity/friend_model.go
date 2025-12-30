package entity

import "time"

type Friend struct {
	ID        string    `json:"id" bson:"_id"`
	UserID1   string    `json:"user_id_1" bson:"user_id_1"`
	UserID2   string    `json:"user_id_2" bson:"user_id_2"`
	Status    string    `json:"status" bson:"status"` // "pending", "accepted", "blocked"
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}





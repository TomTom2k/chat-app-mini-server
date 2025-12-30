package entity

import "time"

type User struct {
	ID        string    `json:"id" bson:"_id"`
	Username  string    `json:"username,omitempty" bson:"username,omitempty"`
	Email     string    `json:"email" bson:"email" validate:"required,email"`
	Name      string    `json:"name" bson:"name" validate:"required"`
	FullName  string    `json:"-" bson:"full_name"` // Internal field, map to Name
	Password  string    `json:"password,omitempty" bson:"password" validate:"required"`
	Avatar    string    `json:"avatar,omitempty" bson:"avatar,omitempty"`
	Online    bool      `json:"online,omitempty" bson:"online,omitempty"`
	CreatedAt time.Time `json:"createdAt" bson:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updated_at"`
}

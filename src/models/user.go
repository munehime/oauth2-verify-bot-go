package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OAuth struct {
	UserID       string    `bson:"userId,omitempty" json:"userId,omitempty"`
	Username     string    `bson:"username,omitempty" json:"username,omitempty"`
	AvatarURL    string    `bson:"avatarUrl,omitempty" json:"avatarUrl,omitempty"`
	DateAdded    time.Time `bson:"dateAdded,omitempty" json:"dateAdded,omitempty"`
	LastVerified time.Time `bson:"lastVerified,omitempty" json:"lastVerified,omitempty"`
}

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Osu       OAuth              `bson:"osu,omitempty" json:"osu,omitempty"`
	Discord   OAuth              `bson:"discord,omitempty" json:"discord,omitempty"`
	Country   string             `bson:"country,omitempty" json:"country,omitempty"`
	Roles     []string           `bson:"roles,omitempty" json:"roles,omitempty"`
	LastLogin time.Time          `bson:"lastLogin,omitempty" json:"lastLogin,omitempty"`
	CreatedAt time.Time          `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt time.Time          `bson:"updateddAt,omitempty" json:"updateddAt,omitempty"`
}

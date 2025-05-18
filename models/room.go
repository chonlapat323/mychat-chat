package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Room struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name" validate:"required"`
	Type      string             `bson:"type" json:"type" validate:"required,oneof=public private"`
	Members   []SafeUser         `bson:"members" json:"members"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoomID    primitive.ObjectID `bson:"room_id" json:"room_id"`
	SenderID  primitive.ObjectID `bson:"sender_id" json:"sender_id"`
	Sender    string             `bson:"sender" json:"sender"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

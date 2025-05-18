package database

import (
	"context"
	"log"
	"os"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client            *mongo.Client
	RoomCollection    *mongo.Collection
	UserCollection    *mongo.Collection
	MessageCollection *mongo.Collection

	once sync.Once
)

func InitMongo() {
	once.Do(func() {
		uri := os.Getenv("MONGO_URI")
		clientOptions := options.Client().ApplyURI(uri)
		client, err := mongo.Connect(context.Background(), clientOptions)
		if err != nil {
			log.Fatal("MongoDB connection error:", err)
		}
		Client = client

		db := client.Database("mychat") // ชื่อฐานข้อมูล

		// กำหนด collection ต่างๆ
		RoomCollection = db.Collection("rooms")
		UserCollection = db.Collection("users")
		MessageCollection = db.Collection("messages")

		log.Println("MongoDB connected and collections set")
	})
}

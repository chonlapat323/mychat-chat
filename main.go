package main

import (
	"log"
	"mychat-chat/database"
	"mychat-chat/handlers"
	"mychat-chat/middleware"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// ✅ สำคัญมาก!
	database.InitMongo()

	http.Handle("/ws", middleware.CORSMiddleware(http.HandlerFunc(handlers.WebSocketHandler)))

	log.Println("WebSocket service running at :5001")
	log.Fatal(http.ListenAndServe(":5001", nil))
}

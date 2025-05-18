package main

import (
	"log"
	"mychat-chat/database"
	"mychat-chat/handlers"
	"mychat-chat/middleware"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// โหลดค่าจาก .env
	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load()

		if err != nil {
			log.Fatal("ไม่พบไฟล์ .env หรือโหลดไม่สำเร็จ")
		}
	}

	database.InitMongo()

	http.Handle("/ws", middleware.CORSMiddleware(http.HandlerFunc(handlers.WebSocketHandler)))

	log.Println("WebSocket service running at :5001")
	log.Fatal(http.ListenAndServe(":5001", nil))
}

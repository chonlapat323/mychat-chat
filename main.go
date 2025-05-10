package main

import (
	"log"
	"net/http"

	"mychat-chat/handlers"
)

func main() {
	http.HandleFunc("/ws", handlers.WebSocketHandler)

	log.Println("WebSocket service running at :4003")
	err := http.ListenAndServe(":4003", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

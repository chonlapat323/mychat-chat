package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"mychat-chat/utils"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🌐 Incoming WebSocket request")

	rawToken := r.URL.Query().Get("token")
	token := strings.TrimPrefix(rawToken, "Bearer ")
	log.Println("📥 Token from query:", token)

	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	claims, err := utils.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	defer conn.Close()
	log.Printf("User %s connected\n", claims.UserID)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		log.Printf("From %s: %s\n", claims.UserID, msg)

		// ส่งกลับข้อความให้ client
		reply := fmt.Sprintf("📨 Server received your message: %s", msg)
		err = conn.WriteMessage(websocket.TextMessage, []byte(reply))
		if err != nil {
			log.Println("❌ Write error:", err)
		} else {
			log.Println("✅ Sent reply to client:", reply)
		}
	}
}

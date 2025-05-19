package handlers

import (
	"context"
	"encoding/json"
	"log"
	"mychat-chat/database"
	"mychat-chat/models"
	"mychat-chat/utils"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// roomID -> map[connection]userID
var roomConnections = make(map[string]map[*websocket.Conn]string)
var mu sync.Mutex

// MessageEvent represents incoming WebSocket messages from the client
type MessageEvent struct {
	Type   string `json:"type"`
	RoomID string `json:"room_id"`
	Text   string `json:"text,omitempty"`
}

func parseTokenFromCookie(cookieHeader string) string {
	cookies := strings.Split(cookieHeader, ";")
	for _, c := range cookies {
		c = strings.TrimSpace(c)
		if strings.HasPrefix(c, "token=") {
			return strings.TrimPrefix(c, "token=")
		}
	}
	return ""
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	cookieHeader := r.Header.Get("Cookie")
	token := parseTokenFromCookie(cookieHeader)
	if token == "" {
		http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
		return
	}

	claims, err := utils.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	userID := claims.UserID
	userName := claims.Email
	ImageURL := claims.ImageURL
	log.Printf("✅ WebSocket connected: user %s (%s)", userID, userName)

	defer func() {
		conn.Close()
		removeConnectionFromAllRooms(conn)
	}()

	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	go func() {
		ticker := time.NewTicker(54 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	for {
		var msg MessageEvent
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("🔌 WebSocket closed: %v", err)
			break
		}

		switch msg.Type {
		case "join":
			mu.Lock()
			if _, ok := roomConnections[msg.RoomID]; !ok {
				roomConnections[msg.RoomID] = make(map[*websocket.Conn]string)
			}
			roomConnections[msg.RoomID][conn] = userID
			mu.Unlock()

			log.Printf("👥 User %s joined room %s", userID, msg.RoomID)

			// ✅ สร้าง user แบบเต็ม
			user := struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Email    string `json:"email"`
				ImageURL string `json:"image_url"`
			}{
				ID:       userID,
				Name:     userName,
				Email:    userName,
				ImageURL: ImageURL,
			}

			// ✅ เตรียม event
			userJoined := struct {
				Type    string      `json:"type"`
				Payload interface{} `json:"payload"`
			}{
				Type:    "user_joined",
				Payload: user,
			}

			data, _ := json.Marshal(userJoined)

			// ✅ broadcast ไปยังทุกคนในห้อง
			mu.Lock()
			for c := range roomConnections[msg.RoomID] {
				if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
					log.Println("Write error (user_joined):", err)
					c.Close()
					removeConnectionFromAllRooms(c)
				}
			}
			mu.Unlock()

		case "message":
			// ✅ เช็คก่อนว่า connection อยู่ในห้องแล้วหรือยัง
			mu.Lock()
			_, joined := roomConnections[msg.RoomID][conn]
			mu.Unlock()

			if !joined {
				log.Printf("⚠️ Message ignored: user %s has not joined room %s", userID, msg.RoomID)
				continue
			}

			log.Printf("📩 Message from user %s in room %s: %s", userID, msg.RoomID, msg.Text)

			if err := SaveMessageToMongo(msg.RoomID, userID, userName, msg.Text); err != nil {
				log.Println("❌ Failed to save message to MongoDB:", err)
			} else {
				log.Println("💾 Message saved to MongoDB")
			}

			broadcastToRoom(msg.RoomID, userID, userName, msg.Text)

		default:
			log.Printf("⚠️ Unknown message type: %s", msg.Type)
		}
	}
}

func removeConnectionFromAllRooms(conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()
	for roomID, conns := range roomConnections {
		if _, ok := conns[conn]; ok {
			delete(conns, conn)
			log.Printf("❌ Disconnected from room %s", roomID)
		}
	}
}

func broadcastToRoom(roomID, senderID, senderName, text string) {
	mu.Lock()
	conns := roomConnections[roomID]
	mu.Unlock()

	roomObjID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		log.Println("❌ Invalid roomID for broadcast:", err)
		return
	}

	senderObjID, err := primitive.ObjectIDFromHex(senderID)
	if err != nil {
		log.Println("❌ Invalid senderID for broadcast:", err)
		return
	}

	message := models.Message{
		ID:        primitive.NewObjectID(),
		RoomID:    roomObjID,
		SenderID:  senderObjID,
		Sender:    senderName,
		Content:   text,
		CreatedAt: time.Now(),
	}

	data, _ := json.Marshal(struct {
		Type      string    `json:"type"`
		ID        string    `json:"id"`
		RoomID    string    `json:"room_id"`
		SenderID  string    `json:"sender_id"`
		Sender    string    `json:"sender"`
		Content   string    `json:"content"`
		CreatedAt time.Time `json:"created_at"`
	}{
		Type:      "message",
		ID:        message.ID.Hex(),
		RoomID:    message.RoomID.Hex(),
		SenderID:  message.SenderID.Hex(),
		Sender:    message.Sender,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	})

	log.Printf("📢 Broadcasting to %d connections in room %s", len(conns), roomID)
	for conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Println("Write error:", err)
			conn.Close()
			removeConnectionFromAllRooms(conn)
		}
	}
}

func SaveMessageToMongo(roomIDStr, userIDStr, senderName, content string) error {
	roomID, err := primitive.ObjectIDFromHex(roomIDStr)
	if err != nil {
		log.Println("❌ Invalid roomID:", err)
		return err
	}

	senderID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		log.Println("❌ Invalid senderID:", err)
		return err
	}

	message := models.Message{
		RoomID:    roomID,
		SenderID:  senderID,
		Sender:    senderName,
		Content:   content,
		CreatedAt: time.Now(),
	}

	_, err = database.MessageCollection.InsertOne(context.TODO(), message)
	if err != nil {
		log.Println("❌ MongoDB insert error:", err)
	}
	return err
}

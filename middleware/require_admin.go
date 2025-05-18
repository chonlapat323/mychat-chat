package middleware

import (
	"context"
	"mychat-chat/contextkey"
	"mychat-chat/utils"
	"net/http"
)

// RequireAdmin เป็น middleware ที่ตรวจว่า token มี role เป็น admin หรือไม่
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("token")
		if err != nil || tokenCookie.Value == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}
		claims, err := utils.ValidateToken(tokenCookie.Value)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		if claims.Role != "admin" {
			http.Error(w, "Forbidden: admin only", http.StatusForbidden)
			return
		}
		// ส่ง user_id และ role เข้า context
		ctx := context.WithValue(r.Context(), contextkey.UserID, claims.UserID)
		ctx = context.WithValue(ctx, contextkey.Role, claims.Role) // 🔧 แก้ให้ถูก key ด้วย
		next(w, r.WithContext(ctx))
	}
}

package middleware

import (
	"context"
	"mychat-chat/contextkey"
	"mychat-chat/utils"
	"net/http"
)

type contextKey string

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil || cookie.Value == "" {
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}

		tokenString := cookie.Value

		// ✅ ตรวจว่าถูก blacklist หรือไม่
		isBlacklisted, err := utils.IsTokenBlacklisted(tokenString)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		if isBlacklisted {
			http.Error(w, "Token revoked", http.StatusUnauthorized)
			return
		}

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextkey.UserID, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

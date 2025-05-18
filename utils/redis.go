package utils

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var RedisClient *redis.Client

func InitRedis() {
	url := os.Getenv("REDIS_URL") // จาก docker-compose
	RedisClient = redis.NewClient(&redis.Options{
		Addr: url,
	})
}

func IsTokenBlacklisted(token string) (bool, error) {
	_, err := RedisClient.Get(ctx, "blacklist:"+token).Result()
	if err == redis.Nil {
		return false, nil // ยังไม่ถูก block
	}
	if err != nil {
		return false, err
	}
	return true, nil // เจอ → แปลว่าเคย logout แล้ว
}

func BlacklistToken(token string, exp time.Time) error {
	ttl := time.Until(exp)
	if ttl <= 0 {
		ttl = time.Hour // fallback กันไว้ 1 ชม.
	}
	return RedisClient.Set(ctx, "blacklist:"+token, "1", ttl).Err()
}

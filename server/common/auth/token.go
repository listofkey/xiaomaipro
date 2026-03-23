package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

const tokenBlacklistPrefix = "user:token:blacklist:"

func TokenBlacklistKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return tokenBlacklistPrefix + hex.EncodeToString(sum[:])
}

func BlacklistToken(ctx context.Context, rdb *redis.Client, token string, ttl time.Duration) error {
	if rdb == nil || token == "" || ttl <= 0 {
		return nil
	}

	return rdb.Set(ctx, TokenBlacklistKey(token), "1", ttl).Err()
}

func IsTokenBlacklisted(ctx context.Context, rdb *redis.Client, token string) (bool, error) {
	if rdb == nil || token == "" {
		return false, nil
	}

	exists, err := rdb.Exists(ctx, TokenBlacklistKey(token)).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

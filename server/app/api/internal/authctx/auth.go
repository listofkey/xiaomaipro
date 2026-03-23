package authctx

import "context"

type contextKey string

const (
	userIDKey      contextKey = "auth.userID"
	userStatusKey  contextKey = "auth.userStatus"
	accessTokenKey contextKey = "auth.accessToken"
)

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserID(ctx context.Context) int64 {
	value, _ := ctx.Value(userIDKey).(int64)
	return value
}

func WithUserStatus(ctx context.Context, status int32) context.Context {
	return context.WithValue(ctx, userStatusKey, status)
}

func UserStatus(ctx context.Context) int32 {
	value, _ := ctx.Value(userStatusKey).(int32)
	return value
}

func WithAccessToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, accessTokenKey, token)
}

func AccessToken(ctx context.Context) string {
	value, _ := ctx.Value(accessTokenKey).(string)
	return value
}

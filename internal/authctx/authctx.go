package authctx

import (
	"context"

	"github.com/google/uuid"
)

type key int

const (
	userIDKey key = iota
	isAdminKey
)

func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(userIDKey).(uuid.UUID)
	return v, ok
}

func WithIsAdmin(ctx context.Context, isAdmin bool) context.Context {
	return context.WithValue(ctx, isAdminKey, isAdmin)
}

func IsAdmin(ctx context.Context) bool {
	v, _ := ctx.Value(isAdminKey).(bool)
	return v
}

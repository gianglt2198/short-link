package utils

import (
	"context"

	"github.com/gianglt1/short-link/internal/common"
)

func NewRequestID() string {
	return NewID(32, "req")
}

func GetRequestIDFromCtx(ctx context.Context) string {
	requestID, ok := ctx.Value(common.KEY_REQUEST_ID).(string)
	if !ok {
		return ""
	}
	return requestID
}

func GetValueByKeyFromCtx[T any](ctx context.Context, k common.KeyType) *T {
	if v, ok := ctx.Value(common.KeyType(k)).(*T); ok {
		return v
	}

	return nil
}

func ApplyRequestIDWithContext(ctx context.Context) (context.Context, string) {
	if requestID := GetRequestIDFromCtx(ctx); requestID != "" {
		return ctx, requestID
	}

	requestID := NewRequestID()
	return context.WithValue(ctx, common.KEY_REQUEST_ID, requestID), requestID
}

func ApplyValueByKeyWithCtx[T any](ctx context.Context, k common.KeyType, v *T) context.Context {
	return context.WithValue(ctx, k, v)
}

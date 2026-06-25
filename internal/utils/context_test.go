package utils

import (
	"context"
	"testing"

	"github.com/gianglt1/short-link/internal/common"
)

func TestGetRequestIDFromCtx_Present(t *testing.T) {
	ctx := context.WithValue(context.Background(), common.KEY_REQUEST_ID, "req-abc")
	if got := GetRequestIDFromCtx(ctx); got != "req-abc" {
		t.Errorf("got %q, want %q", got, "req-abc")
	}
}

func TestGetRequestIDFromCtx_Missing(t *testing.T) {
	if got := GetRequestIDFromCtx(context.Background()); got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestApplyRequestIDWithContext_GeneratesID(t *testing.T) {
	ctx, id := ApplyRequestIDWithContext(context.Background())
	if id == "" {
		t.Fatal("expected a generated request ID, got empty string")
	}
	if got := GetRequestIDFromCtx(ctx); got != id {
		t.Errorf("context has %q, want %q", got, id)
	}
}

func TestApplyRequestIDWithContext_Idempotent(t *testing.T) {
	ctx := context.WithValue(context.Background(), common.KEY_REQUEST_ID, "existing-id")
	_, id := ApplyRequestIDWithContext(ctx)
	if id != "existing-id" {
		t.Errorf("got %q, want existing-id", id)
	}
}

package auth_test

import (
	"context"
	"testing"

	"github.com/Billy-Davies-2/llm-test/pkg/auth"
)

func TestPerRPCCredentials_GetRequestMetadata(t *testing.T) {
	token := "test-token-123"
	pc := auth.PerRPCCredentials(token)
	md, err := pc.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	authHeader, ok := md["authorization"]
	if !ok {
		t.Fatal("authorization metadata missing")
	}
	if authHeader != "Bearer test-token-123" {
		t.Errorf("got %q, want %q", authHeader, "Bearer test-token-123")
	}
}

func TestUnaryServerInterceptor_MissingToken(t *testing.T) {
	interceptor, err := auth.UnaryServerInterceptor(nil, "client-id")
	if err != nil {
		t.Fatalf("failed to create interceptor: %v", err)
	}
	_, err = interceptor(context.Background(), nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	})
	if err == nil {
		t.Error("expected error for missing metadata, got nil")
	}
}

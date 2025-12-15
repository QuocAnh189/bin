package auth

import (
	"context"
	"testing"
	"time"

	"github.com/aq189/bin/internal/domain/token"
	"github.com/aq189/bin/pkg/jwt"
	"github.com/aq189/bin/pkg/logger"
)

func TestService_IssueToken(t *testing.T) {
	// Setup
	jwtService, err := jwt.NewService(jwt.Config{
		Secret:          "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		Issuer:          "root-server",
	})
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}

	log := logger.New(logger.Config{
		Level:  "debug",
		Format: "json",
	})

	service := New(Config{
		JWTService: jwtService,
		Logger:     log,
	})

	ctx := context.Background()

	t.Run("successfully issues token", func(t *testing.T) {
		claims := token.Claims{
			Subject: "user-123",
			Roles:   []string{"admin"},
		}

		tok, err := service.IssueToken(ctx, claims)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if tok == nil {
			t.Fatal("expected token, got nil")
		}

		if tok.Type != token.TypeAccess {
			t.Errorf("expected token type %s, got %s", token.TypeAccess, tok.Type)
		}

		if tok.Value == "" {
			t.Error("expected non-empty token value")
		}
	})

	t.Run("issued token can be validated", func(t *testing.T) {
		claims := token.Claims{
			Subject: "user-456",
			Roles:   []string{"user"},
		}

		tok, err := service.IssueToken(ctx, claims)
		if err != nil {
			t.Fatalf("failed to issue token: %v", err)
		}

		validatedClaims, err := service.ValidateToken(ctx, tok.Value)
		if err != nil {
			t.Fatalf("failed to validate token: %v", err)
		}

		if validatedClaims.Subject != claims.Subject {
			t.Errorf("expected subject %s, got %s", claims.Subject, validatedClaims.Subject)
		}
	})
}

func TestService_ValidateToken(t *testing.T) {
	jwtService, _ := jwt.NewService(jwt.Config{
		Secret:          "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		Issuer:          "root-server",
	})

	log := logger.New(logger.Config{Level: "error", Format: "json"})
	service := New(Config{JWTService: jwtService, Logger: log})
	ctx := context.Background()

	t.Run("rejects invalid token", func(t *testing.T) {
		_, err := service.ValidateToken(ctx, "invalid-token")
		if err == nil {
			t.Error("expected error for invalid token, got nil")
		}
	})

	t.Run("rejects empty token", func(t *testing.T) {
		_, err := service.ValidateToken(ctx, "")
		if err == nil {
			t.Error("expected error for empty token, got nil")
		}
	})
}

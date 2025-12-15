package auth

import (
	"context"
	"fmt"

	"github.com/aq189/bin/internal/domain/token"
	"github.com/aq189/bin/pkg/jwt"
	"github.com/aq189/bin/pkg/logger"
)

// Service handles authentication operations
type Service struct {
	jwtService jwt.Service
	logger     logger.Logger
}

// Config holds auth service configuration
type Config struct {
	JWTService jwt.Service
	Logger     logger.Logger
}

// New creates a new auth service
func New(cfg Config) *Service {
	return &Service{
		jwtService: cfg.JWTService,
		logger:     cfg.Logger,
	}
}

// IssueToken generates a new JWT token
func (s *Service) IssueToken(ctx context.Context, claims token.Claims) (*token.Token, error) {
	tok, err := s.jwtService.Generate(claims)
	if err != nil {
		s.logger.Error("failed to generate token", map[string]any{"error": err})
		return nil, fmt.Errorf("generate token: %w", err)
	}

	s.logger.Info("token issued", map[string]any{
		"subject": claims.Subject,
		"type":    tok.Type,
	})

	return tok, nil
}

// ValidateToken validates a JWT token and returns its claims
func (s *Service) ValidateToken(ctx context.Context, tokenString string) (*token.Claims, error) {
	claims, err := s.jwtService.Validate(tokenString)
	if err != nil {
		s.logger.Warn("token validation failed", map[string]any{"error": err})
		return nil, fmt.Errorf("validate token: %w", err)
	}

	return claims, nil
}

// RefreshToken generates a new access token from a refresh token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*token.Token, error) {
	claims, err := s.jwtService.Validate(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Issue new access token
	newToken, err := s.IssueToken(ctx, *claims)
	if err != nil {
		return nil, fmt.Errorf("issue new token: %w", err)
	}

	return newToken, nil
}

// RevokeToken revokes a token (implementation depends on token blacklist)
func (s *Service) RevokeToken(ctx context.Context, tokenString string) error {
	// TODO: Implement token blacklist with Redis
	s.logger.Info("token revoked", map[string]any{"token": tokenString})
	return nil
}

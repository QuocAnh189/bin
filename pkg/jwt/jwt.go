package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aq189/bin/internal/domain/token"
)

// Service handles JWT operations
type Service interface {
	Generate(claims token.Claims) (*token.Token, error)
	Validate(tokenString string) (*token.Claims, error)
}

// Config holds JWT service configuration
type Config struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
}

// jwtService implements JWT service
type jwtService struct {
	config Config
}

// NewService creates a new JWT service
func NewService(config Config) (Service, error) {
	if config.Secret == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}

	return &jwtService{
		config: config,
	}, nil
}

// Generate creates a new JWT token
func (s *jwtService) Generate(claims token.Claims) (*token.Token, error) {
	now := time.Now()
	expiresAt := now.Add(s.config.AccessTokenTTL)

	claims.Issuer = s.config.Issuer
	claims.IssuedAt = now
	claims.ExpiresAt = expiresAt

	// Create header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return nil, fmt.Errorf("marshal header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return nil, fmt.Errorf("marshal claims: %w", err)
	}

	// Encode header and claims
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	message := headerEncoded + "." + claimsEncoded
	signature := s.sign(message)

	tokenString := message + "." + signature

	return &token.Token{
		Value:     tokenString,
		Type:      token.TypeAccess,
		ExpiresAt: expiresAt,
		IssuedAt:  now,
	}, nil
}

// Validate validates a JWT token and returns its claims
func (s *jwtService) Validate(tokenString string) (*token.Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerEncoded, claimsEncoded, signature := parts[0], parts[1], parts[2]

	// Verify signature
	message := headerEncoded + "." + claimsEncoded
	expectedSignature := s.sign(message)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsEncoded)
	if err != nil {
		return nil, fmt.Errorf("decode claims: %w", err)
	}

	var claims token.Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal claims: %w", err)
	}

	// Validate expiration
	if time.Now().After(claims.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

// sign creates an HMAC signature
func (s *jwtService) sign(message string) string {
	h := hmac.New(sha256.New, []byte(s.config.Secret))
	h.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

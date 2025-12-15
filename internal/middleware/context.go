package middleware

import (
	"context"

	"github.com/aq189/bin/internal/domain/token"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	claimsKey    contextKey = "claims"
)

// contextWithRequestID adds a request ID to the context
func contextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext retrieves the request ID from the context
func RequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// contextWithClaims adds JWT claims to the context
func contextWithClaims(ctx context.Context, claims *token.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// ClaimsFromContext retrieves JWT claims from the context
func ClaimsFromContext(ctx context.Context) *token.Claims {
	if claims, ok := ctx.Value(claimsKey).(*token.Claims); ok {
		return claims
	}
	return nil
}

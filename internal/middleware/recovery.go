package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/aq189/bin/pkg/logger"
)

// Recovery creates a middleware that recovers from panics
func Recovery(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("panic recovered", map[string]any{
						"error":      fmt.Sprintf("%v", err),
						"stack":      string(debug.Stack()),
						"request_id": RequestIDFromContext(r.Context()),
						"path":       r.URL.Path,
					})

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

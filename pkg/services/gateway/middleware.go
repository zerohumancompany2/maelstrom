package gateway

import (
	"net/http"
)

// AuthMiddleware applies authentication to HTTP handlers
type AuthMiddleware struct{}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// Apply applies auth middleware to a handler
func (m *AuthMiddleware) Apply(handler http.Handler) http.Handler {
	return nil
}

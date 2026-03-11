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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || len(authHeader) < 7 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		prefix := authHeader[:7]
		if prefix != "Bearer " {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		token := authHeader[7:]
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

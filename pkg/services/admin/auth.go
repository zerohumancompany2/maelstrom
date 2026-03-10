package admin

import (
	"errors"
	"sync"
	"time"
)

type AuthManager interface {
	Verify2FA(token string) error
	CreateToken(userID string, duration time.Duration) (string, error)
}

type authManager struct {
	tokens map[string]*tokenEntry
	mu     sync.RWMutex
}

type tokenEntry struct {
	userID    string
	expiresAt time.Time
}

func NewAuthManager() AuthManager {
	return &authManager{
		tokens: make(map[string]*tokenEntry),
	}
}

func (am *authManager) Verify2FA(token string) error {
	am.mu.RLock()
	defer am.mu.RUnlock()

	entry, ok := am.tokens[token]
	if !ok {
		return errors.New("invalid 2FA token")
	}

	if time.Now().After(entry.expiresAt) {
		return errors.New("2FA token expired")
	}

	return nil
}

func (am *authManager) CreateToken(userID string, duration time.Duration) (string, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	token := generateToken(userID, time.Now())
	am.tokens[token] = &tokenEntry{
		userID:    userID,
		expiresAt: time.Now().Add(duration),
	}

	return token, nil
}

func generateToken(userID string, timestamp time.Time) string {
	return userID + ":" + timestamp.Format("20060102150405")
}

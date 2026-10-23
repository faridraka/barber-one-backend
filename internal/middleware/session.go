package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"
)

const SessionLifetime = 24 * time.Hour

type Session struct {
	UserID    int64
	ExpiresAt time.Time
}

type SessionStore struct {
	mu       sync.Mutex
	sessions map[string]Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]Session)}
}

func (s *SessionStore) Create(userID int64) (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(randomBytes)
	now := time.Now()
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Clean up expired sessions
	for existingToken, existingSession := range s.sessions {
		if !now.Before(existingSession.ExpiresAt) {
			delete(s.sessions, existingToken)
		}
	}
	
	s.sessions[token] = Session{
		UserID:    userID,
		ExpiresAt: now.Add(SessionLifetime),
	}
	return token, nil
}

func (s *SessionStore) UserIDFromRequest(r *http.Request) (int64, bool) {
	token := BearerToken(r)
	if token == "" {
		return 0, false
	}
	
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	
	storedSession, ok := s.sessions[token]
	if !ok {
		return 0, false
	}
	
	if !now.Before(storedSession.ExpiresAt) {
		delete(s.sessions, token)
		return 0, false
	}
	
	return storedSession.UserID, true
}

func (s *SessionStore) RevokeFromRequest(r *http.Request) {
	if token := BearerToken(r); token != "" {
		s.mu.Lock()
		delete(s.sessions, token)
		s.mu.Unlock()
	}
}

func BearerToken(r *http.Request) string {
	const prefix = "Bearer "
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token
}

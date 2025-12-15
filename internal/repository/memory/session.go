package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aq189/bin/internal/domain/session"
)

// SessionRepository implements in-memory session storage
type SessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*session.Session
}

// NewSessionRepository creates a new in-memory session repository
func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		sessions: make(map[string]*session.Session),
	}
}

// Create stores a new session
func (r *SessionRepository) Create(ctx context.Context, sess *session.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[sess.ID]; exists {
		return fmt.Errorf("session already exists")
	}

	r.sessions[sess.ID] = sess
	return nil
}

// Get retrieves a session by ID
func (r *SessionRepository) Get(ctx context.Context, id string) (*session.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sess, exists := r.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	return sess, nil
}

// Update updates an existing session
func (r *SessionRepository) Update(ctx context.Context, sess *session.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[sess.ID]; !exists {
		return fmt.Errorf("session not found")
	}

	r.sessions[sess.ID] = sess
	return nil
}

// Delete removes a session
func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.sessions, id)
	return nil
}

// DeleteExpired removes all expired sessions
func (r *SessionRepository) DeleteExpired(ctx context.Context) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	now := time.Now()

	for id, sess := range r.sessions {
		if sess.ExpiresAt.Before(now) {
			delete(r.sessions, id)
			count++
		}
	}

	return count, nil
}

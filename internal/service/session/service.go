package session

import (
	"context"
	"fmt"
	"time"

	"root/internal/domain/session"
	"root/pkg/logger"
)

// SessionRepository defines the interface for session storage
type SessionRepository interface {
	Create(ctx context.Context, session *session.Session) error
	Get(ctx context.Context, id string) (*session.Session, error)
	Update(ctx context.Context, session *session.Session) error
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) (int, error)
}

// Service handles session management
type Service struct {
	repo          SessionRepository
	defaultTTL    time.Duration
	cleanupPeriod time.Duration
	logger        logger.Logger
}

// Config holds session service configuration
type Config struct {
	Repository    SessionRepository
	DefaultTTL    time.Duration
	CleanupPeriod time.Duration
	Logger        logger.Logger
}

// New creates a new session service
func New(cfg Config) *Service {
	return &Service{
		repo:          cfg.Repository,
		defaultTTL:    cfg.DefaultTTL,
		cleanupPeriod: cfg.CleanupPeriod,
		logger:        cfg.Logger,
	}
}

// Create creates a new session
func (s *Service) Create(ctx context.Context, userID, serviceID string, data map[string]any, ttl time.Duration) (*session.Session, error) {
	if ttl == 0 {
		ttl = s.defaultTTL
	}

	now := time.Now()
	sess := &session.Session{
		ID:        generateSessionID(),
		UserID:    userID,
		ServiceID: serviceID,
		Data:      data,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(ttl),
	}

	if err := s.repo.Create(ctx, sess); err != nil {
		s.logger.Error("failed to create session", map[string]any{"error": err})
		return nil, fmt.Errorf("create session: %w", err)
	}

	s.logger.Info("session created", map[string]any{
		"session_id": sess.ID,
		"user_id":    userID,
		"service_id": serviceID,
	})

	return sess, nil
}

// Get retrieves a session by ID
func (s *Service) Get(ctx context.Context, id string) (*session.Session, error) {
	sess, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	if sess.IsExpired() {
		s.logger.Warn("attempted to access expired session", map[string]any{"session_id": id})
		return nil, fmt.Errorf("session expired")
	}

	return sess, nil
}

// Update updates session data
func (s *Service) Update(ctx context.Context, id string, data map[string]any) error {
	sess, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	sess.Data = data
	sess.Touch()

	if err := s.repo.Update(ctx, sess); err != nil {
		s.logger.Error("failed to update session", map[string]any{"error": err, "session_id": id})
		return fmt.Errorf("update session: %w", err)
	}

	return nil
}

// Delete deletes a session
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete session", map[string]any{"error": err, "session_id": id})
		return fmt.Errorf("delete session: %w", err)
	}

	s.logger.Info("session deleted", map[string]any{"session_id": id})
	return nil
}

// StartCleanup starts a background goroutine to clean up expired sessions
func (s *Service) StartCleanup(ctx context.Context) {
	ticker := time.NewTicker(s.cleanupPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("stopping session cleanup", map[string]any{})
			return
		case <-ticker.C:
			deleted, err := s.repo.DeleteExpired(ctx)
			if err != nil {
				s.logger.Error("failed to delete expired sessions", map[string]any{"error": err})
				continue
			}
			if deleted > 0 {
				s.logger.Info("cleaned up expired sessions", map[string]any{"count": deleted})
			}
		}
	}
}

func generateSessionID() string {
	// TODO: Implement secure session ID generation
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}

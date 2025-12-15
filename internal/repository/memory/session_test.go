package memory

import (
	"context"
	"testing"
	"time"

	"root/internal/domain/session"
)

func TestSessionRepository_Create(t *testing.T) {
	repo := NewSessionRepository()
	ctx := context.Background()

	sess := &session.Session{
		ID:        "sess-123",
		UserID:    "user-123",
		ServiceID: "service-1",
		Data:      map[string]any{"key": "value"},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UpdatedAt: time.Now(),
	}

	t.Run("creates session successfully", func(t *testing.T) {
		err := repo.Create(ctx, sess)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("rejects duplicate session", func(t *testing.T) {
		err := repo.Create(ctx, sess)
		if err == nil {
			t.Error("expected error for duplicate session, got nil")
		}
	})
}

func TestSessionRepository_Get(t *testing.T) {
	repo := NewSessionRepository()
	ctx := context.Background()

	sess := &session.Session{
		ID:        "sess-456",
		UserID:    "user-456",
		ServiceID: "service-2",
		Data:      map[string]any{"theme": "dark"},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UpdatedAt: time.Now(),
	}

	repo.Create(ctx, sess)

	t.Run("retrieves existing session", func(t *testing.T) {
		retrieved, err := repo.Get(ctx, sess.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if retrieved.ID != sess.ID {
			t.Errorf("expected ID %s, got %s", sess.ID, retrieved.ID)
		}

		if retrieved.UserID != sess.UserID {
			t.Errorf("expected UserID %s, got %s", sess.UserID, retrieved.UserID)
		}
	})

	t.Run("returns error for non-existent session", func(t *testing.T) {
		_, err := repo.Get(ctx, "non-existent")
		if err == nil {
			t.Error("expected error for non-existent session, got nil")
		}
	})
}

func TestSessionRepository_Update(t *testing.T) {
	repo := NewSessionRepository()
	ctx := context.Background()

	sess := &session.Session{
		ID:        "sess-789",
		UserID:    "user-789",
		ServiceID: "service-3",
		Data:      map[string]any{"count": 1},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UpdatedAt: time.Now(),
	}

	repo.Create(ctx, sess)

	t.Run("updates existing session", func(t *testing.T) {
		sess.Data["count"] = 2
		err := repo.Update(ctx, sess)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		retrieved, _ := repo.Get(ctx, sess.ID)
		if retrieved.Data["count"] != 2 {
			t.Errorf("expected count 2, got %v", retrieved.Data["count"])
		}
	})

	t.Run("returns error for non-existent session", func(t *testing.T) {
		nonExistent := &session.Session{ID: "non-existent"}
		err := repo.Update(ctx, nonExistent)
		if err == nil {
			t.Error("expected error for non-existent session, got nil")
		}
	})
}

func TestSessionRepository_Delete(t *testing.T) {
	repo := NewSessionRepository()
	ctx := context.Background()

	sess := &session.Session{
		ID:        "sess-delete",
		UserID:    "user-delete",
		ServiceID: "service-delete",
		Data:      map[string]any{},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UpdatedAt: time.Now(),
	}

	repo.Create(ctx, sess)

	t.Run("deletes existing session", func(t *testing.T) {
		err := repo.Delete(ctx, sess.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, err = repo.Get(ctx, sess.ID)
		if err == nil {
			t.Error("expected error when getting deleted session, got nil")
		}
	})

	t.Run("deleting non-existent session does not error", func(t *testing.T) {
		err := repo.Delete(ctx, "non-existent")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func TestSessionRepository_DeleteExpired(t *testing.T) {
	repo := NewSessionRepository()
	ctx := context.Background()

	// Create expired session
	expiredSess := &session.Session{
		ID:        "sess-expired",
		UserID:    "user-expired",
		ServiceID: "service-expired",
		Data:      map[string]any{},
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-2 * time.Hour),
	}

	// Create valid session
	validSess := &session.Session{
		ID:        "sess-valid",
		UserID:    "user-valid",
		ServiceID: "service-valid",
		Data:      map[string]any{},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		UpdatedAt: time.Now(),
	}

	repo.Create(ctx, expiredSess)
	repo.Create(ctx, validSess)

	t.Run("deletes only expired sessions", func(t *testing.T) {
		count, err := repo.DeleteExpired(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if count != 1 {
			t.Errorf("expected 1 deleted session, got %d", count)
		}

		// Verify expired session is deleted
		_, err = repo.Get(ctx, expiredSess.ID)
		if err == nil {
			t.Error("expected expired session to be deleted")
		}

		// Verify valid session still exists
		_, err = repo.Get(ctx, validSess.ID)
		if err != nil {
			t.Error("expected valid session to still exist")
		}
	})
}

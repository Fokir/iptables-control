package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo          *Repository
	sessionMaxAge time.Duration
}

func NewService(repo *Repository, sessionMaxAgeSec int) *Service {
	return &Service{
		repo:          repo,
		sessionMaxAge: time.Duration(sessionMaxAgeSec) * time.Second,
	}
}

func (s *Service) EnsureAdminUser(username, password string) error {
	count, err := s.repo.UserCount()
	if err != nil {
		return fmt.Errorf("count users: %w", err)
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	_, err = s.repo.CreateUser(username, string(hash))
	if err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}

	slog.Info("admin user created", "username", username)
	return nil
}

func (s *Service) Login(username, password string) (*Session, *User, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fmt.Errorf("invalid credentials")
		}
		return nil, nil, fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	session := &Session{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(s.sessionMaxAge),
	}

	if err := s.repo.CreateSession(session); err != nil {
		return nil, nil, fmt.Errorf("create session: %w", err)
	}

	return session, user, nil
}

func (s *Service) ValidateSession(sessionID string) (*User, error) {
	session, err := s.repo.GetSession(sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("get session: %w", err)
	}

	user, err := s.repo.GetUserByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	// Sliding window: extend session if more than half expired
	halfLife := s.sessionMaxAge / 2
	if time.Until(session.ExpiresAt) < halfLife {
		_ = s.repo.ExtendSession(session.ID, time.Now().Add(s.sessionMaxAge))
	}

	return user, nil
}

func (s *Service) Logout(sessionID string) error {
	return s.repo.DeleteSession(sessionID)
}

func (s *Service) CleanupExpiredSessions() error {
	return s.repo.DeleteExpiredSessions()
}

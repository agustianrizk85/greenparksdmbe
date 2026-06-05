// Package auth provides token-based authentication for the SDM dashboard:
// password login against the user store and in-memory bearer-token sessions.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"greenpark/sdm/internal/domain"
	"greenpark/sdm/internal/passwd"
	"greenpark/sdm/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("username atau password salah")
	ErrUnauthorized       = errors.New("sesi tidak valid atau kedaluwarsa")
)

type session struct {
	userID  string
	expires time.Time
}

// Service authenticates users and tracks active sessions.
type Service struct {
	repo repository.Repository
	ttl  time.Duration

	mu       sync.Mutex
	sessions map[string]session
}

// New returns an auth Service backed by the repository's user store.
func New(repo repository.Repository, ttl time.Duration) *Service {
	if ttl <= 0 {
		ttl = 12 * time.Hour
	}
	return &Service{repo: repo, ttl: ttl, sessions: make(map[string]session)}
}

// Login verifies credentials and returns a new session token + the user.
func (s *Service) Login(username, password string) (string, domain.User, error) {
	u, err := s.repo.UserByUsername(username)
	if err != nil || !passwd.Verify(password, u.Salt, u.PasswordHash) {
		return "", domain.User{}, ErrInvalidCredentials
	}
	token, err := newToken()
	if err != nil {
		return "", domain.User{}, err
	}
	s.mu.Lock()
	s.sessions[token] = session{userID: u.ID, expires: time.Now().Add(s.ttl)}
	s.mu.Unlock()
	return token, u, nil
}

// Validate resolves a bearer token to its user, enforcing expiry.
func (s *Service) Validate(token string) (domain.User, error) {
	s.mu.Lock()
	sess, ok := s.sessions[token]
	if ok && time.Now().After(sess.expires) {
		delete(s.sessions, token)
		ok = false
	}
	s.mu.Unlock()
	if !ok {
		return domain.User{}, ErrUnauthorized
	}
	u, err := s.repo.UserByID(sess.userID)
	if err != nil {
		return domain.User{}, ErrUnauthorized
	}
	return u, nil
}

// Logout invalidates a token (no-op if unknown).
func (s *Service) Logout(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

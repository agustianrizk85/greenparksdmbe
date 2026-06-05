// Package service holds the business logic of the SDM Legal-Permit dashboard:
// generic CRUD over the data collections plus the authentication use-cases. The
// transport layer stays thin.
package service

import (
	"greenpark/sdm/internal/auth"
	"greenpark/sdm/internal/domain"
	"greenpark/sdm/internal/repository"
)

// Service exposes the data read/write and auth use-cases.
type Service struct {
	repo repository.Repository
	auth *auth.Service
}

// New builds a Service from the repository and auth service.
func New(repo repository.Repository, authSvc *auth.Service) *Service {
	return &Service{repo: repo, auth: authSvc}
}

/* ---- data ---- */

// Data returns every collection in one map (used by the dashboard initial load).
func (s *Service) Data() map[string][]domain.Record { return s.repo.Data() }

func (s *Service) List(collection string) ([]domain.Record, error) {
	return s.repo.List(collection)
}
func (s *Service) Create(collection string, rec domain.Record) (domain.Record, error) {
	return s.repo.Create(collection, rec)
}
func (s *Service) Update(collection, id string, patch domain.Record) (domain.Record, error) {
	return s.repo.Update(collection, id, patch)
}
func (s *Service) Delete(collection, id string) (bool, error) {
	return s.repo.Delete(collection, id)
}

/* ---- auth ---- */

func (s *Service) Login(username, password string) (string, domain.User, error) {
	return s.auth.Login(username, password)
}
func (s *Service) Validate(token string) (domain.User, error) { return s.auth.Validate(token) }
func (s *Service) Logout(token string)                        { s.auth.Logout(token) }

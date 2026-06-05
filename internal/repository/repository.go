// Package repository defines storage access for the SDM Legal-Permit dashboard
// and ships a file-backed, in-memory implementation seeded from an embedded
// snapshot. Collections are generic JSON record lists; writes are mutex-guarded
// and persisted to a JSON file so edits survive restarts.
package repository

import (
	"errors"

	"greenpark/sdm/internal/domain"
)

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("resource not found")

// ErrUnknownCollection is returned for a collection name not in the allow-list.
var ErrUnknownCollection = errors.New("unknown collection")

// Repository is the persistence boundary for the dashboard data set.
type Repository interface {
	// ---- reads ----
	Data() map[string][]domain.Record
	List(collection string) ([]domain.Record, error)

	// ---- writes ----
	Create(collection string, rec domain.Record) (domain.Record, error)
	Update(collection, id string, patch domain.Record) (domain.Record, error)
	Delete(collection, id string) (bool, error)

	// ---- users (auth) ----
	UserByUsername(username string) (domain.User, error)
	UserByID(id string) (domain.User, error)
}

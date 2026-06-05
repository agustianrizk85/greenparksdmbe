package repository

import (
	"strings"
	"sync"

	"greenpark/sdm/internal/domain"
)

// fileRepository is a mutex-guarded Repository. The full state is held in memory
// for fast reads and flushed to disk on every write.
type fileRepository struct {
	mu   sync.RWMutex
	path string
	st   *state
}

// NewRepository returns a Repository persisted to the given JSON file path.
// An empty path keeps everything in memory only (handy for tests).
func NewRepository(path string) (Repository, error) {
	st, err := load(path)
	if err != nil {
		return nil, err
	}
	return &fileRepository{path: path, st: st}, nil
}

// persist flushes the current state. Callers must hold the write lock.
func (r *fileRepository) persist() error { return save(r.path, r.st) }

/* ---------------------------- reads ---------------------------- */

func (r *fileRepository) Data() map[string][]domain.Record {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string][]domain.Record, len(r.st.Collections))
	for _, c := range domain.Collections {
		out[c] = clone(r.st.Collections[c])
	}
	return out
}

func (r *fileRepository) List(collection string) ([]domain.Record, error) {
	if !domain.IsCollection(collection) {
		return nil, ErrUnknownCollection
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return clone(r.st.Collections[collection]), nil
}

/* ---------------------------- writes ---------------------------- */

func (r *fileRepository) Create(collection string, rec domain.Record) (domain.Record, error) {
	if !domain.IsCollection(collection) {
		return nil, ErrUnknownCollection
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if recordID(rec) == "" {
		rec["id"] = newID(string(collection[0]))
	}
	r.st.Collections[collection] = append(r.st.Collections[collection], rec)
	return rec, r.persist()
}

func (r *fileRepository) Update(collection, id string, patch domain.Record) (domain.Record, error) {
	if !domain.IsCollection(collection) {
		return nil, ErrUnknownCollection
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	xs := r.st.Collections[collection]
	for i, rec := range xs {
		if recordID(rec) == id {
			for k, v := range patch {
				if k == "id" {
					continue
				}
				rec[k] = v
			}
			xs[i] = rec
			return rec, r.persist()
		}
	}
	return nil, ErrNotFound
}

func (r *fileRepository) Delete(collection, id string) (bool, error) {
	if !domain.IsCollection(collection) {
		return false, ErrUnknownCollection
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	xs := r.st.Collections[collection]
	for i, rec := range xs {
		if recordID(rec) == id {
			r.st.Collections[collection] = append(xs[:i:i], xs[i+1:]...)
			return true, r.persist()
		}
	}
	return false, nil
}

/* ---------------------------- users ---------------------------- */

func (r *fileRepository) UserByUsername(username string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	username = strings.ToLower(strings.TrimSpace(username))
	for _, u := range r.st.Users {
		if strings.ToLower(u.Username) == username {
			return u.toDomain(), nil
		}
	}
	return domain.User{}, ErrNotFound
}

func (r *fileRepository) UserByID(id string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.st.Users {
		if u.ID == id {
			return u.toDomain(), nil
		}
	}
	return domain.User{}, ErrNotFound
}

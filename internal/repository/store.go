package repository

import (
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"greenpark/sdm/internal/domain"
	"greenpark/sdm/internal/passwd"
)

// seedJSON is the initial Legal-Permit data snapshot (13 collections), exported
// from the design's seedData and embedded so it ships inside the binary.
//
//go:embed seed.json
var seedJSON []byte

// state is the full persisted snapshot: every collection plus user accounts.
type state struct {
	Collections map[string][]domain.Record `json:"collections"`
	Users       []storeUser                `json:"users"`
}

// storeUser is the persisted user shape (keeps password material on disk).
type storeUser struct {
	ID           string      `json:"id"`
	Username     string      `json:"username"`
	Name         string      `json:"name"`
	Role         domain.Role `json:"role"`
	PasswordHash string      `json:"passwordHash"`
	Salt         string      `json:"salt"`
}

func (u storeUser) toDomain() domain.User {
	return domain.User{
		ID:           u.ID,
		Username:     u.Username,
		Name:         u.Name,
		Role:         u.Role,
		PasswordHash: u.PasswordHash,
		Salt:         u.Salt,
	}
}

// seedState builds the default state from the embedded snapshot + default users.
func seedState() (*state, error) {
	cols := map[string][]domain.Record{}
	if err := json.Unmarshal(seedJSON, &cols); err != nil {
		return nil, err
	}
	// Ensure every known collection exists (so empty ones are still addressable).
	for _, c := range domain.Collections {
		if cols[c] == nil {
			cols[c] = []domain.Record{}
		}
	}
	return &state{Collections: cols, Users: seedUsers()}, nil
}

// seedUsers creates the default accounts. Change these in any real deployment.
func seedUsers() []storeUser {
	mk := func(id, username, name string, role domain.Role, password string) storeUser {
		salt := passwd.NewSalt()
		return storeUser{ID: id, Username: username, Name: name, Role: role, Salt: salt, PasswordHash: passwd.Hash(password, salt)}
	}
	return []storeUser{
		mk("usr-admin", "admin", "Administrator SDM & Legal", domain.RoleAdmin, "admin123"),
		mk("usr-viewer", "viewer", "Viewer", domain.RoleViewer, "viewer123"),
	}
}

// load reads the state from disk; seeds + writes a fresh one if missing.
func load(path string) (*state, error) {
	if path == "" {
		return seedState()
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			s, err := seedState()
			if err != nil {
				return nil, err
			}
			if err := save(path, s); err != nil {
				return nil, err
			}
			return s, nil
		}
		return nil, err
	}
	s := &state{}
	if err := json.Unmarshal(b, s); err != nil {
		return nil, err
	}
	if s.Collections == nil {
		s.Collections = map[string][]domain.Record{}
	}
	for _, c := range domain.Collections {
		if s.Collections[c] == nil {
			s.Collections[c] = []domain.Record{}
		}
	}
	return s, nil
}

// save atomically writes the state to disk (temp + rename).
func save(path string, s *state) error {
	if path == "" {
		return nil
	}
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// newID returns a short, collision-resistant identifier with the given prefix.
func newID(prefix string) string {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return prefix + "-0000000000"
	}
	return prefix + "-" + hex.EncodeToString(b)
}

// recordID extracts the "id" string from a record (empty if absent).
func recordID(r domain.Record) string {
	if v, ok := r["id"].(string); ok {
		return v
	}
	return ""
}

// clone returns a shallow copy of a record slice (reads never alias the store).
func clone(xs []domain.Record) []domain.Record {
	out := make([]domain.Record, len(xs))
	copy(out, xs)
	return out
}

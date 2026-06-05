// Package domain holds the core entities of the SDM & Legal "Legal-Permit War
// Room". Business records are heterogeneous per collection, so they are modelled
// as generic JSON objects (Record). Each record carries an "id".
package domain

// Record is one generic JSON object in a collection (has an "id" field).
type Record = map[string]any

// Collections is the fixed set of editable collections (mirrors the front-end).
var Collections = []string{
	"projects", "units", "bpn", "pbg", "pks", "documents",
	"risks", "bottlenecks", "escalations", "evidence", "actions", "daily", "weekly",
}

// IsCollection reports whether name is a known collection.
func IsCollection(name string) bool {
	for _, c := range Collections {
		if c == name {
			return true
		}
	}
	return false
}

// Role enumerates the access levels for a dashboard user.
type Role string

const (
	RoleAdmin  Role = "admin"  // full CRUD access
	RoleViewer Role = "viewer" // read-only access
)

// User is a dashboard account. Password material is never serialised to clients.
type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	Role         Role   `json:"role"`
	PasswordHash string `json:"-"`
	Salt         string `json:"-"`
}

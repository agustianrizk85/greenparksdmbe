package http

import "net/http"

// NewRouter wires all routes and applies global middleware.
//
// Access tiers:
//   - public: health + login
//   - authenticated: reads (data + list)
//   - admin: every write (create / update / delete)
func NewRouter(h *Handler, allowOrigin string) http.Handler {
	mux := http.NewServeMux()

	// public
	mux.HandleFunc("GET /api/health", h.health)

	// authenticated session — login/logout/refresh owned by master auth (SSO)
	mux.HandleFunc("GET /api/auth/me", h.requireAuth(h.me))

	// reads
	mux.HandleFunc("GET /api/data", h.requireAuth(h.data))
	mux.HandleFunc("GET /api/{col}", h.requireAuth(h.list))

	// writes (admin) — generic CRUD per collection
	mux.HandleFunc("POST /api/{col}", h.requireAdmin(h.create))
	mux.HandleFunc("PATCH /api/{col}/{id}", h.requireAdmin(h.update))
	mux.HandleFunc("PUT /api/{col}/{id}", h.requireAdmin(h.update))
	mux.HandleFunc("DELETE /api/{col}/{id}", h.requireAdmin(h.remove))

	return chain(mux, logger, cors(allowOrigin))
}

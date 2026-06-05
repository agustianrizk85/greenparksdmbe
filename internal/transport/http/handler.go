package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"greenpark/sdm/internal/domain"
	"greenpark/sdm/internal/repository"
	"greenpark/sdm/internal/service"
)

// Handler holds the dependencies for the HTTP handlers.
type Handler struct {
	svc *service.Service
}

// NewHandler creates a Handler bound to the service.
func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

/* ---------------------------- auth plumbing ---------------------------- */

type ctxKey int

const userCtxKey ctxKey = 0

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimSpace(h[len("Bearer "):])
	}
	return ""
}

func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := h.svc.Validate(bearer(r))
		if err != nil {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userCtxKey, u)))
	}
}

func (h *Handler) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return h.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		if u, ok := r.Context().Value(userCtxKey).(domain.User); !ok || u.Role != domain.RoleAdmin {
			writeError(w, http.StatusForbidden, "butuh akses admin")
			return
		}
		next(w, r)
	})
}

func decodeRecord(w http.ResponseWriter, r *http.Request) (domain.Record, bool) {
	var v domain.Record
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON tidak valid: "+err.Error())
		return nil, false
	}
	return v, true
}

/* ---------------------------- auth handlers ---------------------------- */

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "body JSON tidak valid")
		return
	}
	token, user, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": user})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	h.svc.Logout(bearer(r))
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	u, _ := r.Context().Value(userCtxKey).(domain.User)
	writeJSON(w, http.StatusOK, u)
}

/* ---------------------------- data handlers ---------------------------- */

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "sdm"})
}

func (h *Handler) data(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.svc.Data())
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	recs, err := h.svc.List(r.PathValue("col"))
	if err != nil {
		writeCollectionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, recs)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	rec, ok := decodeRecord(w, r)
	if !ok {
		return
	}
	saved, err := h.svc.Create(r.PathValue("col"), rec)
	if err != nil {
		writeCollectionError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, saved)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	patch, ok := decodeRecord(w, r)
	if !ok {
		return
	}
	saved, err := h.svc.Update(r.PathValue("col"), r.PathValue("id"), patch)
	if err != nil {
		writeCollectionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

func (h *Handler) remove(w http.ResponseWriter, r *http.Request) {
	ok, err := h.svc.Delete(r.PathValue("col"), r.PathValue("id"))
	if err != nil {
		writeCollectionError(w, err)
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "data tidak ditemukan")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeCollectionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrUnknownCollection):
		writeError(w, http.StatusNotFound, "koleksi tidak dikenal")
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "data tidak ditemukan")
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

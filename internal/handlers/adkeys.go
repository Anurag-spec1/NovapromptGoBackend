package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yourname/novapromptgobackend/internal/httputil"
	"github.com/yourname/novapromptgobackend/internal/models"
	"github.com/yourname/novapromptgobackend/internal/repository"
)

type AdKeyHandler struct{ Repo *repository.AdKeyRepo }

func (h *AdKeyHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var value map[string]any
	if err := httputil.Decode(r, &value); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	k := &models.AdKey{Key: key, Value: value}
	if err := h.Repo.Upsert(r.Context(), k); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, k)
}

func (h *AdKeyHandler) Get(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	k, err := h.Repo.Get(r.Context(), key)
	if errors.Is(err, repository.ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "not_found", "key not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, k)
}

func (h *AdKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.Repo.List(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]any{"items": items})
}
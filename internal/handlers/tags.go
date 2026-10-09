package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yourname/novapromptgobackend/internal/httputil"
	"github.com/yourname/novapromptgobackend/internal/models"
	"github.com/yourname/novapromptgobackend/internal/repository"
)

type TagHandler struct{ Repo *repository.TagRepo }

func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	var t models.Tag
	if err := httputil.Decode(r, &t); err != nil || t.Name == "" {
		httputil.Error(w, http.StatusBadRequest, "invalid_body", "name is required")
		return
	}
	if err := h.Repo.Create(r.Context(), &t); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, t)
}

func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.Repo.List(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *TagHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var t models.Tag
	if err := httputil.Decode(r, &t); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	t.ID = id
	if err := h.Repo.Update(r.Context(), &t); errors.Is(err, repository.ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "not_found", "tag not found")
		return
	} else if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, t)
}

func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.Repo.Delete(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "not_found", "tag not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
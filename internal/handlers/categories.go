package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yourname/novapromptgobackend/internal/httputil"
	"github.com/yourname/novapromptgobackend/internal/models"
	"github.com/yourname/novapromptgobackend/internal/repository"
)

type CategoryHandler struct{ Repo *repository.CategoryRepo }

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var c models.Category
	if err := httputil.Decode(r, &c); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if c.Name == "" || c.Slug == "" {
		httputil.Error(w, http.StatusBadRequest, "validation", "name and slug are required")
		return
	}
	if err := h.Repo.Create(r.Context(), &c); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, c)
}

func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := h.Repo.GetByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "not_found", "category not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, c)
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.Repo.List(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var c models.Category
	if err := httputil.Decode(r, &c); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	c.ID = id
	if err := h.Repo.Update(r.Context(), &c); errors.Is(err, repository.ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "not_found", "category not found")
		return
	} else if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, c)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.Repo.Delete(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "not_found", "category not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yourname/novapromptgobackend/internal/httputil"
	"github.com/yourname/novapromptgobackend/internal/models"
	"github.com/yourname/novapromptgobackend/internal/repository"
)

type DescriptionHandler struct{ Repo *repository.DescriptionRepo }

func (h *DescriptionHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	imageID := chi.URLParam(r, "imageID")
	var body struct {
		Body string `json:"body"`
	}
	if err := httputil.Decode(r, &body); err != nil || body.Body == "" {
		httputil.Error(w, http.StatusBadRequest, "invalid_body", "body is required")
		return
	}
	d := &models.Description{ImageID: imageID, Body: body.Body}
	if err := h.Repo.Upsert(r.Context(), d); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, d)
}

func (h *DescriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	imageID := chi.URLParam(r, "imageID")
	d, err := h.Repo.GetByImageID(r.Context(), imageID)
	if errors.Is(err, repository.ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "not_found", "description not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, d)
}

func (h *DescriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	imageID := chi.URLParam(r, "imageID")
	err := h.Repo.DeleteByImageID(r.Context(), imageID)
	if errors.Is(err, repository.ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "not_found", "description not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
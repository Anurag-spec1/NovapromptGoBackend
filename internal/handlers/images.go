package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/yourname/novapromptgobackend/internal/httputil"
	"github.com/yourname/novapromptgobackend/internal/models"
	"github.com/yourname/novapromptgobackend/internal/repository"
)

type ImageHandler struct {
	Repo *repository.ImageRepo
}

func (h *ImageHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title         string  `json:"title"`
		CloudinaryURL string  `json:"cloudinary_url"`
		CategoryID    *string `json:"category_id,omitempty"`
	}
	if err := httputil.Decode(r, &body); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if body.Title == "" || body.CloudinaryURL == "" {
		httputil.Error(w, http.StatusBadRequest, "validation", "title and cloudinary_url are required")
		return
	}
	img := &models.Image{
		Title:         body.Title,
		CloudinaryURL: body.CloudinaryURL,
		CategoryID:    body.CategoryID,
	}
	if err := h.Repo.Create(r.Context(), img); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, img)
}

func (h *ImageHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	img, err := h.Repo.GetByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "not_found", "image not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, img)
}

func (h *ImageHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	images, err := h.Repo.List(r.Context(), limit, offset)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]any{"items": images, "limit": limit, "offset": offset})
}

func (h *ImageHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Title         string  `json:"title"`
		CloudinaryURL string  `json:"cloudinary_url"`
		CategoryID    *string `json:"category_id,omitempty"`
	}
	if err := httputil.Decode(r, &body); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	img := &models.Image{ID: id, Title: body.Title, CloudinaryURL: body.CloudinaryURL, CategoryID: body.CategoryID}
	if err := h.Repo.Update(r.Context(), img); errors.Is(err, repository.ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "not_found", "image not found")
		return
	} else if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, img)
}

func (h *ImageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.Repo.Delete(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "not_found", "image not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Anurag-spec1/novapromptgobackend/internal/cache"
	"github.com/Anurag-spec1/novapromptgobackend/internal/httputil"
	"github.com/Anurag-spec1/novapromptgobackend/internal/models"
	"github.com/Anurag-spec1/novapromptgobackend/internal/repository"
)

type CategoryHandler struct {
	Repo  *repository.CategoryRepo
	Cache *cache.Client
}

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
	
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := h.Cache.Del(ctx, cache.KeyCategoriesList()); err != nil {
			cache.LogSafe("invalidate.categories_list", err)
		}
	}()
	httputil.JSON(w, http.StatusCreated, c)
}

func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	key := cache.KeyCategory(id)

	var cached models.Category
	if hit, err := h.Cache.GetJSON(r.Context(), key, &cached); err != nil {
		cache.LogSafe("cache.get.category", err, "id", id)
	} else if hit {
		httputil.JSON(w, http.StatusOK, cached)
		return
	}

	c, err := h.Repo.GetByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "not_found", "category not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	
	if err := h.Cache.SetJSON(r.Context(), key, c, cache.TTLCategories*time.Second); err != nil {
		cache.LogSafe("cache.set.category", err, "id", id)
	}
	httputil.JSON(w, http.StatusOK, c)
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	key := cache.KeyCategoriesList()

	var cached struct {
		Items []models.Category `json:"items"`
	}
	if hit, err := h.Cache.GetJSON(r.Context(), key, &cached); err != nil {
		cache.LogSafe("cache.get.category_list", err)
	} else if hit {
		slog.Debug("cache_hit", "key", key)
		httputil.JSON(w, http.StatusOK, cached)
		return
	}
	slog.Debug("cache_miss", "key", key)

	items, err := h.Repo.List(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	resp := struct {
		Items []models.Category `json:"items"`
	}{Items: items}

	if err := h.Cache.SetJSON(r.Context(), key, resp, cache.TTLCategories*time.Second); err != nil {
		cache.LogSafe("cache.set.category_list", err)
	}
	httputil.JSON(w, http.StatusOK, resp)
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
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := h.Cache.Del(ctx, cache.KeyCategory(id), cache.KeyCategoriesList()); err != nil {
			cache.LogSafe("invalidate.category", err, "id", id)
		}
	}()
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
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := h.Cache.Del(ctx, cache.KeyCategory(id), cache.KeyCategoriesList()); err != nil {
			cache.LogSafe("invalidate.category", err, "id", id)
		}
	}()
	w.WriteHeader(http.StatusNoContent)
}
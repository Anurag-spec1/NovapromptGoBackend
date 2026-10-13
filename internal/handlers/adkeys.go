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

type AdKeyHandler struct {
	Repo  *repository.AdKeyRepo
	Cache *cache.Client
}

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
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := h.Cache.Del(ctx, cache.KeyAdKey(key), cache.KeyAdKeysList()); err != nil {
			cache.LogSafe("invalidate.adkey", err, "key", key)
		}
	}()
	httputil.JSON(w, http.StatusOK, k)
}

func (h *AdKeyHandler) Get(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	cacheKey := cache.KeyAdKey(key)

	var cached models.AdKey
	if hit, err := h.Cache.GetJSON(r.Context(), cacheKey, &cached); err != nil {
		cache.LogSafe("cache.get.adkey", err, "key", key)
	} else if hit {
		httputil.JSON(w, http.StatusOK, cached)
		return
	}

	k, err := h.Repo.Get(r.Context(), key)
	if errors.Is(err, repository.ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "not_found", "key not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	if err := h.Cache.SetJSON(r.Context(), cacheKey, k, cache.TTLAdKeys*time.Second); err != nil {
		cache.LogSafe("cache.set.adkey", err, "key", key)
	}
	httputil.JSON(w, http.StatusOK, k)
}

func (h *AdKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	cacheKey := cache.KeyAdKeysList()
	var cached struct {
		Items []models.AdKey `json:"items"`
	}
	if hit, err := h.Cache.GetJSON(r.Context(), cacheKey, &cached); err != nil {
		cache.LogSafe("cache.get.adkeys_list", err)
	} else if hit {
		slog.Debug("cache_hit", "key", cacheKey)
		httputil.JSON(w, http.StatusOK, cached)
		return
	}
	slog.Debug("cache_miss", "key", cacheKey)

	items, err := h.Repo.List(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	resp := struct {
		Items []models.AdKey `json:"items"`
	}{Items: items}
	if err := h.Cache.SetJSON(r.Context(), cacheKey, resp, cache.TTLAdKeys*time.Second); err != nil {
		cache.LogSafe("cache.set.adkeys_list", err)
	}
	httputil.JSON(w, http.StatusOK, resp)
}
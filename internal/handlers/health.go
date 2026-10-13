package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Anurag-spec1/novapromptgobackend/internal/cache"
	"github.com/Anurag-spec1/novapromptgobackend/internal/httputil"
)

type HealthHandler struct {
	Pool  *pgxpool.Pool
	Cache *cache.Client
}

func (h *HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "ok"
	redisStatus := "ok"

	if err := h.Pool.Ping(ctx); err != nil {
		dbStatus = "down"
	}
	if err := h.Cache.Ping(ctx); err != nil {
		redisStatus = "down"
	}

	overall := http.StatusOK
	if dbStatus != "ok" {
		overall = http.StatusServiceUnavailable
	}

	httputil.JSON(w, overall, map[string]any{
		"status":  statusText(overall),
		"postgres": dbStatus,
		"redis":    redisStatus,
	})
}

func statusText(code int) string {
	if code == http.StatusOK {
		return "ready"
	}
	return "degraded"
}
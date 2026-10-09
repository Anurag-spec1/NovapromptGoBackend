package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourname/novapromptgobackend/internal/httputil"
)

type HealthHandler struct {
	Pool *pgxpool.Pool
}

func (h *HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.Pool.Ping(ctx); err != nil {
		httputil.Error(w, http.StatusServiceUnavailable, "db_unavailable", "database not reachable")
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
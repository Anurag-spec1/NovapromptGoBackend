package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb      *redis.Client
	capacity int
	refill   float64
	prefix   string
}

func NewRateLimiter(rdb *redis.Client, prefix string, capacity int, refill float64) *RateLimiter {
	return &RateLimiter{rdb: rdb, prefix: prefix, capacity: capacity, refill: refill}
}

var tokenBucketScript = redis.NewScript(`
local key       = KEYS[1]
local capacity  = tonumber(ARGV[1])
local refill    = tonumber(ARGV[2])
local now       = tonumber(ARGV[3])
local ttl       = tonumber(ARGV[4])

local data = redis.call("HMGET", key, "tokens", "ts")
local tokens = tonumber(data[1])
local ts     = tonumber(data[2])

if tokens == nil then
  tokens = capacity
  ts = now
end

local elapsed = math.max(0, now - ts)
tokens = math.min(capacity, tokens + elapsed * refill)

local allowed = 0
if tokens >= 1 then
  tokens = tokens - 1
  allowed = 1
end

redis.call("HMSET", key, "tokens", tokens, "ts", now)
redis.call("EXPIRE", key, ttl)
return {allowed, math.floor(tokens)}
`)

func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			key := fmt.Sprintf("%s:%s", rl.prefix, ip)

			ctx, cancel := context.WithTimeout(r.Context(), 100*time.Millisecond)
			defer cancel()

			now := time.Now().Unix()
			ttl := int(rl.capacity/2) + 60
			if ttl < 60 {
				ttl = 60
			}

			res, err := tokenBucketScript.Run(ctx, rl.rdb,
				[]string{key},
				rl.capacity, rl.refill, now, ttl,
			).Result()

			if err != nil {
				slog.Warn("ratelimit_redis_error", "err", err, "ip", ip)
				next.ServeHTTP(w, r)
				return
			}

			arr, ok := res.([]interface{})
			if !ok || len(arr) < 1 {
				slog.Warn("ratelimit_unexpected_result", "res", res)
				next.ServeHTTP(w, r)
				return
			}
			allowed, _ := arr[0].(int64)

			if allowed == 0 {
				w.Header().Set("Retry-After", "1")
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.capacity))
				w.Header().Set("X-RateLimit-Remaining", "0")
				http.Error(w, `{"error":{"code":"rate_limited","message":"too many requests"}}`,
					http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return trimSpace(xff[:i])
			}
		}
		return trimSpace(xff)
	}
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}
	return addr
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
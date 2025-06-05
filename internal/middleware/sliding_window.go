package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/redis/go-redis/v9"
)

type SlidingWindowLimiter struct {
	client    *redis.Client
	limit     int
	window    time.Duration
}

func NewRateLimiter(client *redis.Client) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		client:    client,
	}
}

func (l *SlidingWindowLimiter) SetLimit(limit int, window time.Duration) {
	l.limit = limit
	l.window = window

	if l.limit <= 0 {
		panic("Rate limit must be greater than 0")
	}
	if l.window <= 0 {
		panic("Window duration must be greater than 0")
	}

	if l.client == nil {
		panic("Redis client is not set")
	}
}

func (l *SlidingWindowLimiter) Apply(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ip := r.RemoteAddr
		key := "rate:" + ip + ":" + r.URL.Path

		now := time.Now().Unix()
		windowStart := now - int64(l.window.Seconds())

		pipe := l.client.Pipeline()
		pipe.ZAdd(r.Context(), key, redis.Z{Score: float64(now), Member: now})
		pipe.ZRemRangeByScore(r.Context(), key, "0", fmt.Sprintf("%d", windowStart))
		countCmd := pipe.ZCard(r.Context(), key)
		pipe.Expire(r.Context(), key, l.window)
		_, err := pipe.Exec(r.Context())

		if err != nil {
			log.Printf("Error rate limiter: "+err.Error())
			http.Error(w, "Rate limiter error", http.StatusInternalServerError)
			return
		}

		if countCmd.Val() > int64(l.limit) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next(w, r, ps)
	}
}

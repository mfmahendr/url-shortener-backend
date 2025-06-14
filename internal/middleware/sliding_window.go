package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	nanoid "github.com/matoous/go-nanoid/v2"
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
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Printf("Unable to parse RemoteAddr: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		key := "rate:" + host + ":" + r.URL.Path

		now := time.Now().Unix()
		uniqueID, err := nanoid.Generate("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 6)
		if err != nil {
			log.Println("Error generating ID:", err)
		}
		uniqueMember := fmt.Sprintf("%d.%s", now, uniqueID)
		windowStart := now - int64(l.window.Seconds())

		pipe := l.client.Pipeline()
		pipe.ZAdd(r.Context(), key, redis.Z{Score: float64(now), Member: uniqueMember})
		pipe.ZRemRangeByScore(r.Context(), key, "0", fmt.Sprintf("%d", windowStart))
		countCmd := pipe.ZCard(r.Context(), key)
		pipe.Expire(r.Context(), key, l.window)
		_, err = pipe.Exec(r.Context())

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

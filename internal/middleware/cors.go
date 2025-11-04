package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func CORS(router *httprouter.Router) http.Handler {
	allowed := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	allowedMap := make(map[string]bool)
	for _, origin := range allowed {
		allowedMap[strings.TrimSpace(origin)] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedMap[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		router.ServeHTTP(w, r)
	})
}

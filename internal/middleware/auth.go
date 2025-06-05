package middleware

import (
	"context"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	auth "firebase.google.com/go/v4/auth"
	"github.com/julienschmidt/httprouter"
)

type AuthMiddleware struct {
	AuthClient *auth.Client
}

func NewAuthMiddleware(app *firebase.App) *AuthMiddleware {
	authClient, err := app.Auth(context.Background())
	if err != nil {
		return nil
	}
	return &AuthMiddleware{AuthClient: authClient}
}

func (m *AuthMiddleware) RequireAuth(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// get the Authorization header
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		// extract the token from the header
		idToken := strings.TrimPrefix(header, "Bearer ")
		token, err := m.AuthClient.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user", token.UID)
		r = r.WithContext(ctx)
		next(w, r, p)
	}
}

func (m *AuthMiddleware) RequireAdminAuth(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		ctx := r.Context()

		token, err := m.AuthClient.VerifyIDToken(ctx, idToken)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if isAdmin, ok := token.Claims["admin"].(bool); !ok || !isAdmin {
			http.Error(w, "Forbidden: admin access only", http.StatusForbidden)
			return
		}

		ctx = context.WithValue(ctx, "user", token.UID)
		next(w, r.WithContext(ctx), ps)
	}
}

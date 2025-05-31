//go:build wireinject
// +build wireinject

package di

import (
	"context"

	"github.com/mfmahendr/url-shortener-backend/config"
	"github.com/mfmahendr/url-shortener-backend/internal/controllers"
	"github.com/mfmahendr/url-shortener-backend/internal/middleware"
	firestore_service "github.com/mfmahendr/url-shortener-backend/internal/services/firestore"
	"github.com/mfmahendr/url-shortener-backend/internal/services/url_service"

	"github.com/google/wire"
)

var firebaseAppSet = wire.NewSet(
	config.InitFirebase,
)

func InitializeController(ctx context.Context) (*controllers.URLController, error) {
    wire.Build(
        firebaseAppSet,
        firestore_service.New,
        url_service.New,
        controllers.New,
    )
    return &controllers.URLController{}, nil
}

func InitializeAuthMiddleware(ctx context.Context) (*middleware.AuthMiddleware, error) {
	wire.Build(
		firebaseAppSet,
		middleware.NewAuthMiddleware,
	)
	return &middleware.AuthMiddleware{}, nil
}

//go:build wireinject
// +build wireinject

package di

import (
	"context"

	"github.com/mfmahendr/url-shortener-backend/config"
	"github.com/mfmahendr/url-shortener-backend/internal/controllers"
	"github.com/mfmahendr/url-shortener-backend/internal/middleware"
	firestore_service "github.com/mfmahendr/url-shortener-backend/internal/services/firestore"
	"github.com/mfmahendr/url-shortener-backend/internal/services/tracking_service"
	"github.com/mfmahendr/url-shortener-backend/internal/services/url_service"

	"github.com/google/wire"
)

var firebaseAppSet = wire.NewSet(
    config.InitFirebase,
)

var firestoreServiceSet = wire.NewSet(
    firestore_service.New,
    wire.Bind(new(firestore_service.FirestoreService), new(*firestore_service.FirestoreServiceImpl)),
    wire.Bind(new(firestore_service.Shortlink), new(*firestore_service.FirestoreServiceImpl)),
    wire.Bind(new(firestore_service.ClickLog), new(*firestore_service.FirestoreServiceImpl)),
)

func InitializeController(ctx context.Context) (*controllers.URLController, error) {
    wire.Build(
        firebaseAppSet,
        config.NewRedisClient,
        firestoreServiceSet,
        tracking_service.New,
        url_service.New,
        controllers.New,
    )
    return nil, nil
}

func InitializeAuthMiddleware(ctx context.Context) (*middleware.AuthMiddleware, error) {
    wire.Build(
        firebaseAppSet,
        middleware.NewAuthMiddleware,
    )
    return nil, nil
}
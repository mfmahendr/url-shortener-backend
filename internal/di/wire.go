//go:build wireinject
// +build wireinject

package di

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"github.com/mfmahendr/url-shortener-backend/config"
	"github.com/mfmahendr/url-shortener-backend/internal/controllers"
	"github.com/mfmahendr/url-shortener-backend/internal/middleware"
	firestore_service "github.com/mfmahendr/url-shortener-backend/internal/services/firestore"
	// "google.golang.org/api/safebrowsing/v4"

	safebrowsing_service "github.com/mfmahendr/url-shortener-backend/internal/services/safebrowsing"

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
	wire.Bind(new(firestore_service.BlacklistManager), new(*firestore_service.FirestoreServiceImpl)),
	wire.Bind(new(firestore_service.BlacklistChecker), new(*firestore_service.FirestoreServiceImpl)),
)

func InitializeController(ctx context.Context, app *firebase.App, safeBrowsingKey string) (*controllers.URLController, error) {
	wire.Build(
		firestoreServiceSet,
		config.NewRedisClient,
		tracking_service.New,
		safebrowsing_service.New,
        // safebrowsing.NewService,
		url_service.New,
		middleware.NewRateLimiter,
		controllers.New,
	)
	return nil, nil
}

func InitializeAuthMiddleware(app *firebase.App) (*middleware.AuthMiddleware, error) {
	wire.Build(
		middleware.NewAuthMiddleware,
	)
	return nil, nil
}

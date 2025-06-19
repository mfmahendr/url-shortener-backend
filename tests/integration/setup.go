package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	firebase "firebase.google.com/go/v4"
	"github.com/mfmahendr/url-shortener-backend/tests/integration/containers"
	redis "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	firestore_service "github.com/mfmahendr/url-shortener-backend/internal/services/firestore"
)

type TestContainerEnv struct {
	rdClient    *redis.Client
	FsApp       *firebase.App
	rdContainer *tcredis.RedisContainer
	fbContainer testcontainers.Container
	fbConn      *grpc.ClientConn
}

var (
	tcEnv     *TestContainerEnv
	fsService *firestore_service.FirestoreServiceImpl // firestore service (global because we don't want firebase App to create so many firestore client)
	initErr   error
)

func GetSharedTestContainerEnv(ctx context.Context, t *testing.T) *TestContainerEnv {
	if tcEnv == nil {
		tcEnv, initErr = initializeTestContainerEnvironment(ctx)
	}
	if t != nil && initErr != nil {
		t.Fatalf("failed to initialize test container env:\n%v", initErr)
	}
	return tcEnv
}

// set redis client and firebase (auth, firestore) app
func initializeTestContainerEnvironment(ctx context.Context) (*TestContainerEnv, error) {
	// redis container & client
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %v", err)
	}

	rdsConfPath := filepath.Join(cwd, "testdata", "redis7.conf")
	redisContainer, redisURL, err := containers.StartRedis(ctx, rdsConfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to start redis: %v", err)
	}
	rdClient := redis.NewClient(&redis.Options{Addr: redisURL})

	// Firebase emulator container
	firebaseContainer, authHost, fsHost, err := containers.StartFirebaseEmulator(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start firestore emulator: %v", err)
	}

	// gRPC Connection to Firestore Emulator
	conn, err := grpc.NewClient(
		fsHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(emulatorCreds{}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %v", err)
	}

	os.Setenv("FIREBASE_AUTH_EMULATOR_HOST", authHost)
	firebaseApp, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: "dummy-project"}, option.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("firebase.NewApp: %v", err)
	}

	fsService, err = firestore_service.New(ctx, firebaseApp)
	if err != nil {
		return nil, err
	}

	return &TestContainerEnv{
		rdClient:    rdClient,
		FsApp:       firebaseApp,
		rdContainer: redisContainer,
		fbContainer: firebaseContainer,
		fbConn:      conn,
	}, nil
}

func (tce *TestContainerEnv) Cleanup(ctx context.Context) (err error) {
	err = tce.rdClient.Close()
	if err != nil {
		return
	}

	err = tce.fbContainer.Terminate(ctx)
	if err != nil {
		return
	}

	err = tce.rdContainer.Terminate(ctx)
	if err != nil {
		return
	}

	err = tce.fbConn.Close()
	if err != nil {
		return
	}

	return
}

// emulatorCreds implements grpc.PerRPCCredentials for Firestore Emulator
type emulatorCreds struct{}

func (emulatorCreds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{}, nil
}
func (emulatorCreds) RequireTransportSecurity() bool {
	return false
}

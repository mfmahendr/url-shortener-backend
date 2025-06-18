package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	firebase "firebase.google.com/go/v4"
	"github.com/mfmahendr/url-shortener-backend/tests/integration/containers"
	redis "github.com/redis/go-redis/v9"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TestContainerEnv struct {
	RdClient *redis.Client
	FsApp    *firebase.App
}

func InitializeTestContainerEnvironment(t *testing.T) *TestContainerEnv {
	ctx := context.Background()

	// redis container & client
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	rdsConfPath := filepath.Join(cwd, "testdata", "redis7.conf")
	redisContainer, redisURL, err := containers.StartRedis(ctx, rdsConfPath)
	if err != nil {
		t.Fatalf("failed to start redis: %v", err)
	}
	rdClient := redis.NewClient(&redis.Options{Addr: redisURL})

	// Firebase emulator container
	firebaseContainer, authHost, fsHost, err := containers.StartFirebaseEmulator(ctx)
	if err != nil {
		t.Fatalf("failed to start firestore emulator: %v", err)
	}

	// gRPC Connection to Firestore Emulator
	conn, err := grpc.NewClient(
		fsHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(emulatorCreds{}),
	)
	if err != nil {
		t.Fatalf("failed to create gRPC client: %v", err)
	}

	os.Setenv("FIREBASE_AUTH_EMULATOR_HOST", authHost)
	firebaseApp, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: "dummy-project"}, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("firebase.NewApp: %v", err)
	}

	t.Cleanup(func() {
		if err := rdClient.Close(); err != nil {
			t.Logf("error closing redis client: %v", err)
		}
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Logf("error terminating redis container: %v", err)
		}
		if err := firebaseContainer.Terminate(ctx); err != nil {
			t.Logf("error terminating firebase container: %v", err)
		}
		if err := conn.Close(); err != nil {
			t.Logf("error closing firestore connection: %v", err)
		}
	})

	return &TestContainerEnv{
		RdClient: rdClient,
		FsApp:    firebaseApp,
	}
}

// emulatorCreds implements grpc.PerRPCCredentials for Firestore Emulator
type emulatorCreds struct{}

func (emulatorCreds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{}, nil
}
func (emulatorCreds) RequireTransportSecurity() bool {
	return false
}

package containers

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// only auth and firestore
func StartFirebaseEmulator(ctx context.Context) (tc.Container, string, string, error) {
	req := tc.ContainerRequest{
		Image:        "andreyka26/firebase-emulator:latest",
		Env:          map[string]string{"ENABLE_UI": "false", "EMULATORS": "auth"},
		ExposedPorts: []string{"9099/tcp", "8080/tcp"},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.PortBindings = nat.PortMap{
				"8080/tcp": {{HostPort: "8083"}},
				"9099/tcp": {{HostPort: "9099"}},
			}
		},
		WaitingFor: wait.ForLog("All emulators ready!").WithStartupTimeout(2 * time.Minute),
	}
	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", "", err
	}

	authHost, err := extractEndpoint(ctx, container, "9099/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, "", "", err
	}
	firestoreHost, err := extractEndpoint(ctx, container, "8080/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, "", "", err
	}

	return container, authHost, firestoreHost, nil
}

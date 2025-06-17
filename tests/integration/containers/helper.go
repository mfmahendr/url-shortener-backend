package containers

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
)

func extractEndpoint(ctx context.Context, container testcontainers.Container, port string) (string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return "", err
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(port))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s", host, mappedPort.Port()), nil
}

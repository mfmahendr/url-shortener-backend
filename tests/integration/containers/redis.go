package containers

import (
	"context"

	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func StartRedis(ctx context.Context, confPath string) (*tcredis.RedisContainer, string, error) {
	container, err := tcredis.Run(ctx,
		"redis:7.2-alpine",					// image
		tcredis.WithSnapshotting(10, 1),
		tcredis.WithLogLevel(tcredis.LogLevelVerbose),
		tcredis.WithConfigFile(confPath),
	)
	if err != nil {
		return nil, "", err
	}

	endpoint, err := extractEndpoint(ctx, container.Container, "6379/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, "", err
	}

	return container, endpoint, nil
}

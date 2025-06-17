package containers

import (
	"context"

	tcfirestore "github.com/testcontainers/testcontainers-go/modules/gcloud/firestore"
)

func StartFirestoreEmulator(ctx context.Context) (*tcfirestore.Container, string, error) {
	container, err := tcfirestore.Run(
		ctx,
		"gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
		tcfirestore.WithProjectID("a-fake-firestore-project-id"),
	)
	if err != nil {
		return nil, "", err
	}

	endpoint, err := extractEndpoint(ctx, container.Container, "8080/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, "", err
	}

	return container, endpoint, nil
}

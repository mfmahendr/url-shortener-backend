package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	var err error
	tcEnv, err = initializeTestContainerEnvironment(ctx)
	if err != nil {
		fmt.Printf("failed to initialize test env: %v\n", err)
		os.Exit(1)
	}

	validators.Init()

	exitCode := m.Run()

	err = tcEnv.Cleanup(ctx)
	if err != nil {
		fmt.Printf("failed to cleanup test env: %v\n", err)
	}

	os.Exit(exitCode)
}

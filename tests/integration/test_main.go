package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	tcEnv, err = initializeTestContainerEnvironment(ctx)
	if err != nil {
		fmt.Printf("failed to initialize test env: %v\n", err)
		os.Exit(1)
	}

	exitCode := m.Run()

	tcEnv.Cleanup(ctx)

	os.Exit(exitCode)
}

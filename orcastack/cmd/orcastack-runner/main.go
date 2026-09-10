package main

import (
	"context"
	"log"

	"github.com/orcastack/orcastackapi/internal/platform/app"
	"github.com/orcastack/orcastackapi/internal/platform/config"
	"github.com/orcastack/orcastackapi/internal/runnerapi"
)

func main() {
	err := app.Run(context.Background(), app.WithServiceSecurity(app.Config{
		Name:               "orcastack-runner",
		Role:               "pipeline-runner",
		Summary:            "Coordinates pipeline execution, live job events, artifacts, and workload dispatch.",
		RegisterHTTPRoutes: runnerapi.Register,
		HTTPPort:           config.String("ORCASTACK_RUNNER_HTTP_PORT", "8086"),
		GRPCPort:           config.String("ORCASTACK_RUNNER_GRPC_PORT", "9086"),
	}, "ORCASTACK_RUNNER_IDENTITY", app.DefaultRunnerIdentity))
	if err != nil {
		log.Fatal(err)
	}
}
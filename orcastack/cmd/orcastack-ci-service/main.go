package main

import (
	"context"
	"log"

	"github.com/orcastack/orcastackapi/internal/platform/app"
	"github.com/orcastack/orcastackapi/internal/platform/config"
)

func main() {
	err := app.Run(context.Background(), app.WithServiceSecurity(app.Config{
		Name:               "orcastack-ci-service",
		Role:               "ci-engine",
		Summary:            "Schedules pipeline jobs, streams logs to HBase, and stores artifacts in HDFS.",
		HTTPPort:           config.String("ORCASTACK_CI_HTTP_PORT", "8083"),
		GRPCPort:           config.String("ORCASTACK_CI_GRPC_PORT", "9083"),
	}, "ORCASTACK_CI_IDENTITY", app.DefaultCIIdentity))
	if err != nil {
		log.Fatal(err)
	}
}

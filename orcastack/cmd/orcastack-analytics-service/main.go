package main

import (
	"context"
	"log"

	"github.com/orcastack/orcastackapi/internal/platform/app"
	"github.com/orcastack/orcastackapi/internal/platform/config"
)

func main() {
	err := app.Run(context.Background(), app.WithServiceSecurity(app.Config{
		Name:               "orcastack-analytics-service",
		Role:               "analytics",
		Summary:            "Computes failure patterns, risky modules, branch health, and developer activity metrics.",
		HTTPPort:           config.String("ORCASTACK_ANALYTICS_HTTP_PORT", "8085"),
		GRPCPort:           config.String("ORCASTACK_ANALYTICS_GRPC_PORT", "9085"),
	}, "ORCASTACK_ANALYTICS_IDENTITY", app.DefaultAnalyticsIdentity))
	if err != nil {
		log.Fatal(err)
	}
}

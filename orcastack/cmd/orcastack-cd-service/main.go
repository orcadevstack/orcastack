package main

import (
	"context"
	"log"

	"github.com/orcastack/orcastackapi/internal/platform/app"
	"github.com/orcastack/orcastackapi/internal/platform/config"
)

func main() {
	err := app.Run(context.Background(), app.WithServiceSecurity(app.Config{
		Name:               "orcastack-cd-service",
		Role:               "cd-engine",
		Summary:            "Deploys verified artifacts to Kubernetes, Docker, OpenStack, or bare metal targets.",
		HTTPPort:           config.String("ORCASTACK_CD_HTTP_PORT", "8084"),
		GRPCPort:           config.String("ORCASTACK_CD_GRPC_PORT", "9084"),
	}, "ORCASTACK_CD_IDENTITY", app.DefaultCDIdentity))
	if err != nil {
		log.Fatal(err)
	}
}

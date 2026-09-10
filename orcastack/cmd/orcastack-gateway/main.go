package main

import (
	"context"
	"log"

	"github.com/orcastack/orcastackapi/internal/gatewayapi"
	"github.com/orcastack/orcastackapi/internal/platform/app"
	"github.com/orcastack/orcastackapi/internal/platform/config"
)

func main() {
	err := app.Run(context.Background(), app.WithServiceSecurity(app.Config{
		Name:               "orcastack-gateway",
		Role:               "api-gateway",
		Summary:            "Single entrypoint for projects, reviews, pipelines, deployments, and analytics.",
		RegisterHTTPRoutes: gatewayapi.Register,
		HTTPPort:           config.String("ORCASTACK_GATEWAY_HTTP_PORT", "8080"),
		GRPCPort:           config.String("ORCASTACK_GATEWAY_GRPC_PORT", "9080"),
	}, "ORCASTACK_GATEWAY_IDENTITY", app.DefaultGatewayIdentity))
	if err != nil {
		log.Fatal(err)
	}
}

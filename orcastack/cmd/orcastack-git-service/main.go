package main

import (
	"context"
	"log"

	"github.com/orcastack/orcastackapi/internal/gitapi"
	"github.com/orcastack/orcastackapi/internal/platform/app"
	"github.com/orcastack/orcastackapi/internal/platform/config"
)

func main() {
	err := app.Run(context.Background(), app.WithServiceSecurity(app.Config{
		Name:               "orcastack-git-service",
		Role:               "git-rpc",
		Summary:            "Owns repository storage, refs, packfile ingestion, and Git metadata queries.",
		RegisterHTTPRoutes: gitapi.Register,
		HTTPPort:           config.String("ORCASTACK_GIT_HTTP_PORT", "8081"),
		GRPCPort:           config.String("ORCASTACK_GIT_GRPC_PORT", "9081"),
	}, "ORCASTACK_GIT_IDENTITY", app.DefaultGitIdentity))
	if err != nil {
		log.Fatal(err)
	}
}

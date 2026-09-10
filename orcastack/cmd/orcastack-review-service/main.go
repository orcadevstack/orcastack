package main

import (
	"context"
	"log"

	"github.com/orcastack/orcastackapi/internal/platform/app"
	"github.com/orcastack/orcastackapi/internal/platform/config"
)

func main() {
	err := app.Run(context.Background(), app.WithServiceSecurity(app.Config{
		Name:               "orcastack-review-service",
		Role:               "code-review",
		Summary:            "Manages changes, patchsets, comments, approvals, and merge rules.",
		HTTPPort:           config.String("ORCASTACK_REVIEW_HTTP_PORT", "8082"),
		GRPCPort:           config.String("ORCASTACK_REVIEW_GRPC_PORT", "9082"),
	}, "ORCASTACK_REVIEW_IDENTITY", app.DefaultReviewIdentity))
	if err != nil {
		log.Fatal(err)
	}
}

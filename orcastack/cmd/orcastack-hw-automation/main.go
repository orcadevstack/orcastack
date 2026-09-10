package main

import (
	"context"
	"log"

	"github.com/orcastack/orcastackapi/internal/hwautomation"
	"github.com/orcastack/orcastackapi/internal/platform/app"
	"github.com/orcastack/orcastackapi/internal/platform/config"
)

func main() {
	err := app.Run(context.Background(), app.WithServiceSecurity(app.Config{
		Name:               "orcastack-hw-automation",
		Role:               "hardware-automation",
		Summary:            "Executes firmware flashing, device pool management, hardware validation, and telemetry collection workflows.",
		RegisterHTTPRoutes: hwautomation.Register,
		HTTPPort:           config.String("ORCASTACK_HW_AUTOMATION_HTTP_PORT", "8087"),
		GRPCPort:           config.String("ORCASTACK_HW_AUTOMATION_GRPC_PORT", "9087"),
	}, "ORCASTACK_HW_AUTOMATION_IDENTITY", app.DefaultHWAutoIdentity))
	if err != nil {
		log.Fatal(err)
	}
}
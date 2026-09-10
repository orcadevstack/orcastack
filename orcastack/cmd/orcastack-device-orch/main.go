package main

import (
	"context"
	"log"

	"github.com/orcastack/orcastackapi/internal/deviceorch"
	"github.com/orcastack/orcastackapi/internal/platform/app"
	"github.com/orcastack/orcastackapi/internal/platform/config"
)

func main() {
	err := app.Run(context.Background(), app.WithServiceSecurity(app.Config{
		Name:               "orcastack-device-orch",
		Role:               "device-orchestration",
		Summary:            "Maintains device inventory, health, reservation state, and allocation data for hardware-aware pipelines.",
		RegisterHTTPRoutes: deviceorch.Register,
		HTTPPort:           config.String("ORCASTACK_DEVICE_ORCH_HTTP_PORT", "8089"),
		GRPCPort:           config.String("ORCASTACK_DEVICE_ORCH_GRPC_PORT", "9089"),
	}, "ORCASTACK_DEVICE_ORCH_IDENTITY", app.DefaultDeviceOrchIdentity))
	if err != nil {
		log.Fatal(err)
	}
}
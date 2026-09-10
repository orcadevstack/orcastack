package hwautomation

import (
	"encoding/json"
	"net/http"
)

type Workflow struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Status        string   `json:"status"`
	TargetPool    string   `json:"target_pool"`
	RequiredTags  []string `json:"required_tags"`
	LastRun       string   `json:"last_run"`
	FirmwareImage string   `json:"firmware_image"`
}

func Register(mux *http.ServeMux) {
	workflows := []Workflow{}

	mux.HandleFunc("/workflows", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]any{
			"engine":    "hardware-automation",
			"workflows": workflows,
			"capabilities": []string{"flash-firmware", "reserve-device", "capture-telemetry", "collect-test-results"},
		})
	})

	mux.HandleFunc("/device-pools", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]any{
			"pools": []map[string]any{},
		})
	})
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
package deviceorch

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Device struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Status       string   `json:"status"`
	Health       string   `json:"health"`
	Tags         []string `json:"tags"`
	Location     string   `json:"location"`
	Capabilities []string `json:"capabilities"`
	AssignedRun  string   `json:"assigned_run,omitempty"`
}

func Register(mux *http.ServeMux) {
	devices := []Device{}

	mux.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/devices" {
			http.NotFound(w, r)
			return
		}
		respondJSON(w, http.StatusOK, map[string]any{
			"devices": devices,
			"summary": map[string]int{"total": len(devices), "healthy": 0, "busy": 0, "offline": 0},
		})
	})

	mux.HandleFunc("/devices/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/devices/")
		for _, device := range devices {
			if device.ID == id {
				respondJSON(w, http.StatusOK, device)
				return
			}
		}
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "device not found"})
	})

	mux.HandleFunc("/allocations", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]any{
			"allocations": []map[string]string{},
		})
	})
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
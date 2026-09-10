package swautomation

import (
	"encoding/json"
	"net/http"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("/workflows", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]any{
			"engine":    "software-automation",
			"workflows": []map[string]any{},
		})
	})

	mux.HandleFunc("/integrations", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]any{
			"integrations": []map[string]any{},
		})
	})
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
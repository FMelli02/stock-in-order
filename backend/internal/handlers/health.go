package handlers

import (
	"encoding/json"
	"net/http"
)

// Health returns 200 OK to indicate the server is reachable.
func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

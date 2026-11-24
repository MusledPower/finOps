package handler

import (
	"encoding/json"
	"net/http"
)

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
func Capabilities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(
		map[string]any{
			"schemaVersion": "1.0",
			"limits": map[string]any{
				"epicChars":      10000,
				"otherChars":     6000,
				"contextSources": 8,
				"contextChars":   36000,
				"requestBytes":   262144,
			},
			"policyVersion": "policy-v1",
		},
	)
}

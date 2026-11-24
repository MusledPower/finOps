package handler

import (
	"encoding/json"
	"net/http"
)

// Health godoc
// @Summary      Health check
// @Description  Checks whether the service is running
// @Tags         System
// @Produce      json
// @Success      200 {object} map[string]string "Service is healthy"
// @Router       /health [get]
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Capabilities godoc
// @Summary      Get API capabilities
// @Description  Returns API schema version, limits and policy version
// @Tags         System
// @Produce      json
// @Success      200 {object} CapabilitiesResponse
// @Router       /capabilities [get]
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

// CapabilitiesResponse represents API capabilities
type CapabilitiesResponse struct {
	SchemaVersion string             `json:"schemaVersion" example:"1.0"`
	Limits        CapabilitiesLimits `json:"limits"`
	PolicyVersion string             `json:"policyVersion" example:"policy-v1"`
}

// CapabilitiesLimits represents API limits
type CapabilitiesLimits struct {
	EpicChars      int `json:"epicChars" example:"10000"`
	OtherChars     int `json:"otherChars" example:"6000"`
	ContextSources int `json:"contextSources" example:"8"`
	ContextChars   int `json:"contextChars" example:"36000"`
	RequestBytes   int `json:"requestBytes" example:"262144"`
}

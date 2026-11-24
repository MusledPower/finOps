package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"task-decomposer-backend/internal/domain"
	"task-decomposer-backend/internal/provider"
	"task-decomposer-backend/internal/repository"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PlanService struct {
	repo     *repository.Repository
	planner  provider.Planner
	maxBytes int64
}

func NewPlanService(r *repository.Repository, p provider.Planner, max int64) *PlanService {
	return &PlanService{r, p, max}
}

// Create godoc
// @Summary      Create plan
// @Description  Generates a plan based on the provided epic, preferences and context
// @Tags         Plans
// @Accept       json
// @Produce      json
// @Param        request body domain.PlanRequest true "Plan creation request"
// @Success      200 {object} domain.PlanResponse
// @Failure      400 {object} ErrorResponse "Validation error"
// @Failure      422 {object} ErrorResponse "Model output validation failed"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Failure      502 {object} ErrorResponse "Planner provider unavailable"
// @Router       /plans [post]
func (s *PlanService) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, s.maxBytes)
	var req domain.PlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "VALIDATION_ERROR", "Invalid JSON request")
		return
	}
	req.Epic.Text = strings.TrimSpace(req.Epic.Text)
	if req.Epic.Text == "" || len(req.Epic.Text) > 10000 {
		writeError(w, 400, "VALIDATION_ERROR", "Epic text is required and must not exceed 10000 characters")
		return
	}
	other := 0
	for _, v := range append(append(req.Epic.AcceptanceCriteria, req.Epic.Constraints...), req.Epic.NonGoals...) {
		other += len(v)
	}
	if other > 6000 {
		writeError(w, 400, "VALIDATION_ERROR", "Acceptance criteria, constraints and non-goals exceed 6000 characters")
		return
	}
	if req.RequestID == "" {
		writeError(w, 400, "VALIDATION_ERROR", "requestId is required")
		return
	}
	if req.Preference == "" {
		req.Preference = "balanced"
	}
	if req.Preference != "balanced" && req.Preference != "quality" && req.Preference != "cost" {
		writeError(w, 400, "VALIDATION_ERROR", "Unknown preference")
		return
	}
	raw, _ := json.Marshal(req)
	sum := sha256.Sum256(raw)
	hash := hex.EncodeToString(sum[:])
	start := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()
	out, model, err := s.planner.Generate(ctx, req)
	if err != nil {
		writeError(w, 502, "PROVIDER_UNAVAILABLE", "Planner provider failed")
		return
	}
	if out.Kind == "plan" {
		if err := ValidatePlan(out.Plan, req.AllowedModels, req.Context); err != nil {
			writeError(w, 422, "MODEL_OUTPUT_INVALID", err.Error())
			return
		}
		if out.Plan.ID == "" {
			out.Plan.ID = uuid.NewString()
		}
		if out.Plan.Version == 0 {
			out.Plan.Version = 1
		}
	}
	out.RequestID = req.RequestID
	out.Meta.LatencyMs = time.Since(start).Milliseconds()
	out.Meta.PlannerModel = model
	if err := s.repo.Save(ctx, req.RequestID, req, out, hash); err != nil {
		writeError(w, 500, "INTERNAL_ERROR", "Could not persist plan response")
		return
	}
	writeJSON(w, 200, out)
}

// Get godoc
// @Summary      Get plan
// @Description  Retrieves a previously generated plan by ID
// @Tags         Plans
// @Produce      json
// @Param        id path string true "Plan ID"
// @Success      200 {object} domain.PlanResponse
// @Failure      404 {object} ErrorResponse "Plan not found"
// @Router       /plans/{id} [get]
func (s *PlanService) Get(w http.ResponseWriter, r *http.Request) {
	b, err := s.repo.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, 404, "NOT_FOUND", "Plan not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_, _ = w.Write(b)
}

// ErrorResponse represents API error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code      string `json:"code" example:"VALIDATION_ERROR"`
	Message   string `json:"message" example:"Invalid JSON request"`
	Retryable bool   `json:"retryable" example:"false"`
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(
		w,
		status,
		map[string]any{"error": map[string]any{"code": code, "message": msg, "retryable": status >= 500}},
	)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

var _ = fmt.Sprintf

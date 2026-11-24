package provider

import (
	"context"
	"task-decomposer-backend/internal/domain"
)

type Planner interface {
	Generate(context.Context, domain.PlanRequest) (domain.PlanResponse, string, error)
}

package service

import (
	"fmt"
	"strings"
	"task-decomposer-backend/internal/domain"
)

func ValidatePlan(
	p *domain.Plan,
	allowed []domain.AllowedModel,
	contextRefs []domain.ContextRef,
) error {
	if p == nil {
		return fmt.Errorf("plan is required")
	}

	if len(p.Tasks) < 1 || len(p.Tasks) > 7 {
		return fmt.Errorf("tasks count must be between 1 and 7")
	}

	ids := map[string]bool{}
	models := map[string]bool{}

	// "epic" — специальный системный источник контекста.
	refs := map[string]bool{
		"epic": true,
	}

	for _, m := range allowed {
		models[m.ID] = true
	}

	for _, c := range contextRefs {
		if c.ID != "" {
			refs[c.ID] = true
		}
	}

	for _, t := range p.Tasks {
		if strings.TrimSpace(t.ID) == "" || ids[t.ID] {
			return fmt.Errorf("task ids must be non-empty and unique")
		}

		ids[t.ID] = true

		if strings.TrimSpace(t.Title) == "" ||
			strings.TrimSpace(t.Goal) == "" ||
			len(t.AcceptanceCriteria) == 0 {
			return fmt.Errorf("task %s has missing required fields", t.ID)
		}

		if t.RecommendedModelID != nil &&
			!models[*t.RecommendedModelID] {
			return fmt.Errorf(
				"task %s recommends a model outside allowedModels",
				t.ID,
			)
		}

		for _, ref := range t.ContextRefs {
			if !refs[ref] {
				return fmt.Errorf(
					"task %s references unknown context %s",
					t.ID,
					ref,
				)
			}
		}
	}

	// Проверяем зависимости между задачами.
	edges := map[string][]string{}

	for _, t := range p.Tasks {
		for _, d := range t.DependsOn {
			if !ids[d] || d == t.ID {
				return fmt.Errorf(
					"task %s has invalid dependency %s",
					t.ID,
					d,
				)
			}

			edges[t.ID] = append(edges[t.ID], d)
		}
	}

	// Проверяем отсутствие циклических зависимостей.
	state := map[string]int{}

	var visit func(string) bool

	visit = func(id string) bool {
		if state[id] == 1 {
			return false
		}

		if state[id] == 2 {
			return true
		}

		state[id] = 1

		for _, d := range edges[id] {
			if !visit(d) {
				return false
			}
		}

		state[id] = 2
		return true
	}

	for id := range ids {
		if !visit(id) {
			return fmt.Errorf("dependency cycle detected")
		}
	}

	return nil
}

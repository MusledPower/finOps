package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"task-decomposer-backend/internal/domain"
)

type DeepSeek struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

func NewDeepSeek(
	baseURL, apiKey, model string,
	timeout time.Duration,
) *DeepSeek {
	return &DeepSeek{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *DeepSeek) Generate(
	ctx context.Context,
	input domain.PlanRequest,
) (domain.PlanResponse, string, error) {

	if p.apiKey == "" || p.model == "" {
		return domain.PlanResponse{}, "", errors.New(
			"DeepSeek provider is not configured",
		)
	}

	system := `You are a task decomposition planner.

Return ONLY valid JSON.

For a plan, use this structure:
{
  "kind": "plan",
  "plan": {
    "summary": "",
    "assumptions": [],
    "warnings": [],
    "tasks": [
      {
        "id": "S1",
        "title": "",
        "goal": "",
        "expectedOutputs": [],
        "acceptanceCriteria": [],
        "dependsOn": [],
        "contextRefs": [],
        "assumptions": [],
        "risk": "low",
        "recommendedModelId": null,
        "recommendedParams": [],
        "recommendationReasons": [],
        "recommendationSource": "policy-v1"
      }
    ]
  }
}

If essential information is missing, return:
{
  "kind": "clarification",
  "questions": [
    {
      "id": "q1",
      "text": "",
      "reason": ""
    }
  ]
}

Rules:
- Create 1-7 meaningful tasks.
- Every task needs observable acceptance criteria.
- Dependencies must reference existing task IDs.
- Do not create circular dependencies.
- Do not invent project facts or model IDs.

Context reference rules:
- Use only IDs explicitly provided in contextSources.
- Never use "epic" as a contextRef.
- If a task is based on the epic description, use an empty contextRefs array.
- Never invent context source IDs.

Return JSON only, without Markdown or explanations.`

	user, err := json.Marshal(input)
	if err != nil {
		return domain.PlanResponse{}, "", fmt.Errorf(
			"marshal plan request: %w", err,
		)
	}

	payload := map[string]any{
		"model": p.model,
		"messages": []any{
			map[string]string{
				"role":    "system",
				"content": system,
			},
			map[string]string{
				"role":    "user",
				"content": string(user),
			},
		},
		"response_format": map[string]string{
			"type": "json_object",
		},
		"thinking": map[string]string{
			"type": "disabled",
		},
		"max_tokens": 8192,
		"stream":     false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return domain.PlanResponse{}, "", fmt.Errorf(
			"marshal DeepSeek payload: %w", err,
		)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.baseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return domain.PlanResponse{}, "", err
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return domain.PlanResponse{}, "", fmt.Errorf(
			"DeepSeek request failed: %w", err,
		)
	}
	defer resp.Body.Close()

	const maxResponseSize = 2 << 20 // 2 MiB

	responseBody, err := io.ReadAll(
		io.LimitReader(resp.Body, maxResponseSize+1),
	)
	if err != nil {
		return domain.PlanResponse{}, "", fmt.Errorf(
			"read DeepSeek response: %w", err,
		)
	}

	if len(responseBody) > maxResponseSize {
		return domain.PlanResponse{}, "", errors.New(
			"DeepSeek response exceeds size limit",
		)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.PlanResponse{}, "", fmt.Errorf(
			"DeepSeek returned HTTP %d: %s",
			resp.StatusCode,
			string(responseBody),
		)
	}

	var envelope struct {
		Model string `json:"model"`

		Choices []struct {
			FinishReason string `json:"finish_reason"`

			Message struct {
				Content *string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		return domain.PlanResponse{}, "", fmt.Errorf(
			"decode DeepSeek response: %w", err,
		)
	}

	if len(envelope.Choices) == 0 {
		return domain.PlanResponse{}, "", errors.New(
			"DeepSeek returned no choices",
		)
	}

	choice := envelope.Choices[0]

	if choice.FinishReason == "length" {
		return domain.PlanResponse{}, envelope.Model, errors.New(
			"DeepSeek response was truncated: max_tokens reached",
		)
	}

	if choice.FinishReason != "stop" {
		return domain.PlanResponse{}, envelope.Model, fmt.Errorf(
			"unexpected DeepSeek finish reason: %s",
			choice.FinishReason,
		)
	}

	if choice.Message.Content == nil ||
		strings.TrimSpace(*choice.Message.Content) == "" {
		return domain.PlanResponse{}, envelope.Model, errors.New(
			"DeepSeek returned empty content",
		)
	}

	var out domain.PlanResponse

	if err := json.Unmarshal(
		[]byte(*choice.Message.Content),
		&out,
	); err != nil {
		return domain.PlanResponse{}, envelope.Model, fmt.Errorf(
			"decode plan JSON: %w", err,
		)
	}

	out.SchemaVersion = "1.0"
	out.Meta.PlannerModel = envelope.Model
	out.Meta.PromptVersion = "1.0"
	out.Meta.ProviderCalls = 1

	return out, envelope.Model, nil
}

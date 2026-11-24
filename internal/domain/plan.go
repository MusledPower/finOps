package domain

import "time"

type PlanRequest struct {
	SchemaVersion        string                `json:"schemaVersion"`
	RequestID            string                `json:"requestId"`
	Epic                 Epic                  `json:"epic"`
	Preference           string                `json:"preference"`
	ProjectProfile       ProjectProfile        `json:"projectProfile"`
	Context              []ContextRef          `json:"context"`
	AllowedModels        []AllowedModel        `json:"allowedModels"`
	Catalog              Catalog               `json:"catalog"`
	ClarificationAnswers []ClarificationAnswer `json:"clarificationAnswers"`
}
type Epic struct {
	Text               string   `json:"text"`
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	Constraints        []string `json:"constraints"`
	NonGoals           []string `json:"nonGoals"`
}
type ProjectProfile struct {
	Stack []string `json:"stack"`
	Text  string   `json:"text"`
}
type ContextRef struct {
	ID          string `json:"id"`
	Path        string `json:"path"`
	StartLine   int    `json:"startLine"`
	EndLine     int    `json:"endLine"`
	ContentHash string `json:"contentHash"`
	Content     string `json:"content"`
}
type AllowedModel struct {
	ID          string   `json:"id"`
	DisplayName string   `json:"displayName"`
	Params      []string `json:"params"`
}
type Catalog struct {
	Source    string    `json:"source"`
	FetchedAt time.Time `json:"fetchedAt"`
}
type ClarificationAnswer struct {
	ID     string `json:"id"`
	Answer string `json:"answer"`
}

type PlanResponse struct {
	SchemaVersion string     `json:"schemaVersion"`
	RequestID     string     `json:"requestId"`
	Kind          string     `json:"kind"`
	Plan          *Plan      `json:"plan,omitempty"`
	Questions     []Question `json:"questions,omitempty"`
	Meta          Meta       `json:"meta,omitempty"`
}
type Plan struct {
	ID          string   `json:"id"`
	Version     int      `json:"version"`
	Summary     string   `json:"summary"`
	Assumptions []string `json:"assumptions"`
	Warnings    []string `json:"warnings"`
	Tasks       []Task   `json:"tasks"`
}
type Task struct {
	ID                    string   `json:"id"`
	Title                 string   `json:"title"`
	Goal                  string   `json:"goal"`
	ExpectedOutputs       []string `json:"expectedOutputs"`
	AcceptanceCriteria    []string `json:"acceptanceCriteria"`
	DependsOn             []string `json:"dependsOn"`
	ContextRefs           []string `json:"contextRefs"`
	Assumptions           []string `json:"assumptions"`
	Risk                  string   `json:"risk"`
	RecommendedModelID    *string  `json:"recommendedModelId"`
	RecommendedParams     []string `json:"recommendedParams"`
	RecommendationReasons []string `json:"recommendationReasons"`
	RecommendationSource  string   `json:"recommendationSource"`
}
type Question struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Reason string `json:"reason"`
}
type Meta struct {
	PlannerModel  string   `json:"plannerModel"`
	PromptVersion string   `json:"promptVersion"`
	PolicyVersion string   `json:"policyVersion"`
	LatencyMs     int64    `json:"latencyMs"`
	ProviderCalls int      `json:"providerCalls"`
	Usage         Usage    `json:"usage"`
	Cost          *float64 `json:"cost"`
}
type Usage struct {
	InputTokens  *int `json:"inputTokens"`
	OutputTokens *int `json:"outputTokens"`
}

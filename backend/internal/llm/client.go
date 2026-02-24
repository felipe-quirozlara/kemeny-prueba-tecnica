package llm

import "context"

// TaskClassification represents the result of an LLM classifying a task.
type TaskClassification struct {
	Tags     []string `json:"tags"`
	Priority string   `json:"priority"` // "high", "medium", "low"
	Category string   `json:"category"` // "bug", "feature", "improvement", "research"
	Summary  string   `json:"summary"`  // One-line summary
}

// LLMClient defines the interface for AI-powered task classification.
// Implementations should handle timeouts, retries, and malformed responses.
type LLMClient interface {
	ClassifyTask(ctx context.Context, title string, description string) (*TaskClassification, error)
}

package tests

import (
	"context"
	"testing"

	"github.com/KemenyStudio/task-manager/internal/llm"
)

// TestMockClient verifies the mock LLM client returns expected classifications.
func TestMockClientBugClassification(t *testing.T) {
	client := llm.NewMockClient()

	result, err := client.ClassifyTask(context.Background(),
		"Fix: Login page crashes on invalid email",
		"When a user enters an email without @ symbol, the login page throws an unhandled exception and shows a white screen.",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Category != "bug" {
		t.Errorf("expected category 'bug', got '%s'", result.Category)
	}

	if result.Priority != "high" {
		t.Errorf("expected priority 'high' for crash bug, got '%s'", result.Priority)
	}

	if !containsTag(result.Tags, "bug") {
		t.Errorf("expected tags to contain 'bug', got %v", result.Tags)
	}
}

func TestMockClientFeatureClassification(t *testing.T) {
	client := llm.NewMockClient()

	result, err := client.ClassifyTask(context.Background(),
		"Implement OAuth with Google",
		"Add Google OAuth as an alternative authentication method. Support redirect flow, callback, and user creation.",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Category != "feature" {
		t.Errorf("expected category 'feature', got '%s'", result.Category)
	}

	if !containsTag(result.Tags, "security") {
		t.Errorf("expected tags to contain 'security' for OAuth task, got %v", result.Tags)
	}
}

func TestMockClientResearchClassification(t *testing.T) {
	client := llm.NewMockClient()

	result, err := client.ClassifyTask(context.Background(),
		"Investigate rate limiting options for the API",
		"Research token bucket vs sliding window, implementation options, Redis vs in-memory.",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Category != "research" {
		t.Errorf("expected category 'research', got '%s'", result.Category)
	}

	if !containsTag(result.Tags, "backend") {
		t.Errorf("expected tags to contain 'backend' for API task, got %v", result.Tags)
	}
}

func TestMockClientImprovementClassification(t *testing.T) {
	client := llm.NewMockClient()

	result, err := client.ClassifyTask(context.Background(),
		"Refactor error handling in backend",
		"Clean up inconsistent error responses. Create centralized error handler with proper types.",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Category != "improvement" {
		t.Errorf("expected category 'improvement', got '%s'", result.Category)
	}
}

func TestMockClientDevOpsClassification(t *testing.T) {
	client := llm.NewMockClient()

	result, err := client.ClassifyTask(context.Background(),
		"Configure CI/CD pipeline with GitHub Actions",
		"Set up pipeline for tests, Docker build, and deploy to staging.",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !containsTag(result.Tags, "devops") {
		t.Errorf("expected tags to contain 'devops', got %v", result.Tags)
	}
}

// TestMockClientImplementsInterface verifies MockClient satisfies LLMClient.
func TestMockClientImplementsInterface(t *testing.T) {
	var _ llm.LLMClient = (*llm.MockClient)(nil)
}

func containsTag(tags []string, target string) bool {
	for _, tag := range tags {
		if tag == target {
			return true
		}
	}
	return false
}

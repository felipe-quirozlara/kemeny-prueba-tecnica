package llm

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os"

    "github.com/openai/openai-go/v3"
    "github.com/openai/openai-go/v3/option"
    "github.com/openai/openai-go/v3/responses"
)

// OpenAIClient uses the official openai-go SDK.
type OpenAIClient struct {
    client openai.Client
    model  string
}

// NewOpenAIClient returns a configured OpenAI client or error if API key missing.
func NewOpenAIClient() (*OpenAIClient, error) {
    key := os.Getenv("OPENAI_API_KEY")
    if key == "" {
        return nil, errors.New("OPENAI_API_KEY not set")
    }
    // allow override model via env var
    model := os.Getenv("OPENAI_MODEL")
    if model == "" {
        model = "gpt-3.5-turbo"
    }
    cli := openai.NewClient(option.WithAPIKey(key))
    return &OpenAIClient{client: cli, model: model}, nil
}

func (c *OpenAIClient) ClassifyTask(ctx context.Context, title, description string) (*TaskClassification, error) {
    prompt := buildPrompt(title, description)

    resp, err := c.client.Responses.New(ctx, responses.ResponseNewParams{
        Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(prompt)},
        Model: c.model,
    })
    if err != nil {
        return nil, fmt.Errorf("openai sdk error: %w", err)
    }

    out := resp.OutputText()
    if out == "" {
        return nil, errors.New("empty response from openai")
    }

    var tc TaskClassification
    if err := parseJSONFromString(out, &tc); err != nil {
        return nil, fmt.Errorf("failed parsing openai output: %w", err)
    }
    return &tc, nil
}

// Helper: build prompt
func buildPrompt(title, description string) string {
    return fmt.Sprintf(`Classify the following task. Reply with JSON only.
{
  "tags": ["tag1","tag2"],
  "priority": "low|medium|high|urgent",
  "category": "bug|feature|improvement|research",
  "summary": "one-line summary"
}

Title: %s

Description: %s
`, title, description)
}

// parseJSONFromString finds the first JSON object in s and unmarshals it into v.
func parseJSONFromString(s string, v interface{}) error {
    // naive approach: find first '{' and last '}'
    start := -1
    end := -1
    for i, ch := range s {
        if ch == '{' && start == -1 {
            start = i
        }
        if ch == '}' {
            end = i
        }
    }
    if start == -1 || end == -1 || end <= start {
        return errors.New("no JSON object found in response")
    }
    sub := s[start:end+1]
    return json.Unmarshal([]byte(sub), v)
}

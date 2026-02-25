package llm


import (
    "context"
    "errors"
    "fmt"
    "os"
    "strings"

    "github.com/anthropics/anthropic-sdk-go"
    "github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicClient uses the official Anthropics Go SDK.
type AnthropicClient struct {
    client anthropic.Client
    model  string
}

// NewAnthropicClient returns a configured Anthropics client or error if API key missing.
func NewAnthropicClient() (*AnthropicClient, error) {
    key := os.Getenv("ANTHROPIC_API_KEY")
    if key == "" {
        return nil, errors.New("ANTHROPIC_API_KEY not set")
    }
    model := os.Getenv("ANTHROPIC_MODEL")
    if model == "" {
        model = "claude-2.1"
    }
    cli := anthropic.NewClient(option.WithAPIKey(key))
    return &AnthropicClient{client: cli, model: model}, nil
}

func (c *AnthropicClient) ClassifyTask(ctx context.Context, title, description string) (*TaskClassification, error) {
    prompt := buildPrompt(title, description)

    msg, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
        MaxTokens: 300,
        Messages: []anthropic.MessageParam{
            anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
        },
        Model: anthropic.Model(c.model),
    })
    if err != nil {
        return nil, fmt.Errorf("anthropic sdk error: %w", err)
    }
    out := contentToString(msg.Content)
    if strings.TrimSpace(out) == "" {
        return nil, errors.New("empty response from anthropic")
    }

    var tc TaskClassification
    if err := parseJSONFromString(out, &tc); err != nil {
        return nil, fmt.Errorf("failed parsing anthropic output: %w", err)
    }
    return &tc, nil
}

// contentToString converts the SDK's ContentBlockUnion slice into a plain string
// by concatenating each element's default string representation. This is a
// conservative approach that extracts visible text regardless of the union
// internals; if you prefer a stricter extraction (e.g. only text blocks) we
// can refine it after inspecting the SDK types.
func contentToString(content []anthropic.ContentBlockUnion) string {
    var parts []string
    for _, cb := range content {
        parts = append(parts, fmt.Sprintf("%v", cb))
    }
    return strings.Join(parts, " ")
}

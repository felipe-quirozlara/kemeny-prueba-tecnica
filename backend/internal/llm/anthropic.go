package llm

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "time"
)

// AnthropicClient is a minimal HTTP client for Anthropic completions.
type AnthropicClient struct {
    apiKey     string
    httpClient *http.Client
    model      string
}

func NewAnthropicClient() (*AnthropicClient, error) {
    key := os.Getenv("ANTHROPIC_API_KEY")
    if key == "" {
        return nil, errors.New("ANTHROPIC_API_KEY not set")
    }
    model := os.Getenv("ANTHROPIC_MODEL")
    if model == "" {
        model = "claude-2.1"
    }
    return &AnthropicClient{
        apiKey: key,
        httpClient: &http.Client{ Timeout: 15 * time.Second },
        model: model,
    }, nil
}

func (c *AnthropicClient) ClassifyTask(ctx context.Context, title, description string) (*TaskClassification, error) {
    prompt := buildPrompt(title, description)
    reqBody := map[string]interface{}{
        "model": c.model,
        "prompt": prompt,
        "max_tokens": 300,
        "temperature": 0.0,
    }
    b, err := json.Marshal(reqBody)
    if err != nil { return nil, err }

    req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/complete", io.NopCloser(bytesReader(b)))
    if err != nil { return nil, err }
    req.Header.Set("x-api-key", c.apiKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil { return nil, err }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("anthropic error: status=%d body=%s", resp.StatusCode, string(body))
    }

    var respBody struct{
        Completion string `json:"completion"`
    }
    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    if err := json.Unmarshal(bodyBytes, &respBody); err != nil {
        return nil, err
    }

    var tc TaskClassification
    if err := parseJSONFromString(respBody.Completion, &tc); err != nil {
        return nil, fmt.Errorf("failed parsing anthropic output: %w", err)
    }
    return &tc, nil
}

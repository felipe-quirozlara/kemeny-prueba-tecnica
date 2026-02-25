package llm

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "os"
    "time"
)

// OpenAIClient uses the OpenAI REST Chat Completions API.
type OpenAIClient struct {
    apiKey     string
    httpClient *http.Client
    model      string
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
    return &OpenAIClient{
        apiKey: key,
        httpClient: &http.Client{
            Timeout: 15 * time.Second,
        },
        model: model,
    }, nil
}

func (c *OpenAIClient) ClassifyTask(ctx context.Context, title, description string) (*TaskClassification, error) {
    prompt := buildPrompt(title, description)

    reqBody := map[string]interface{}{
        "model": c.model,
        "messages": []map[string]string{
            {"role": "system", "content": "You are a helpful task classifier. Respond with JSON only."},
            {"role": "user", "content": prompt},
        },
        "max_tokens": 300,
        "temperature": 0.0,
    }

    b, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", io.NopCloser(bytesReader(b)))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", "Bearer "+c.apiKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("openai error: status=%d body=%s", resp.StatusCode, string(body))
    }

    var respBody struct {
        Choices []struct{
            Message struct{
                Content string `json:"content"`
            } `json:"message"`
        } `json:"choices"`
    }
    bodyBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    if err := json.Unmarshal(bodyBytes, &respBody); err != nil {
        return nil, err
    }
    if len(respBody.Choices) == 0 {
        return nil, errors.New("no choices from openai")
    }

    raw := respBody.Choices[0].Message.Content
    // Try to unmarshal the content as JSON into TaskClassification
    var tc TaskClassification
    if err := parseJSONFromString(raw, &tc); err != nil {
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

// bytesReader returns an io.Reader from bytes without importing bytes at top-level to keep file simple
func bytesReader(b []byte) *reader { return &reader{b:b} }

type reader struct{ b []byte; i int }
func (r *reader) Read(p []byte) (n int, err error) {
    if r.i >= len(r.b) { return 0, io.EOF }
    n = copy(p, r.b[r.i:])
    r.i += n
    return n, nil
}

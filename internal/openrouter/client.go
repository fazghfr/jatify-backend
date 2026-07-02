package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultBaseURL = "https://openrouter.ai/api/v1"

// ORClient is the interface the service depends on.
type ORClient interface {
	AnalyzeResume(ctx context.Context, resumeText string) (string, error)
	AnalyzeJobMatch(ctx context.Context, resumeText string, jobDescription string) (string, error)
}

// Client is the concrete OpenRouter HTTP client.
type Client struct {
	apiKey  string
	model   string
	baseURL string
	http    *http.Client
}

// New creates a Client. model defaults to "openai/gpt-oss-120b:free" if empty.
func New(apiKey, model string) *Client {
	if model == "" {
		model = "openai/gpt-oss-120b:free"
	}
	return &Client{
		apiKey:  apiKey,
		model:   model,
		baseURL: defaultBaseURL,
		http:    &http.Client{},
	}
}

// --- request / response shapes ---

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// systemprompt is hardcoded.
// TODO: make this configurable by an authorized user
const systemPrompt = `You are an expert resume reviewer. Analyze the provided resume text and
return a JSON object with the following fields:
  summary (string), skills ([]string), experience_years (int),
  strengths ([]string), weaknesses ([]string),
  improvement_suggestions ([]string), overall_score (int, 1-10).
Return only the JSON object, no extra prose.`

const jobMatchSystemPrompt = `You are an expert recruiter. Compare the provided resume against the job description and return a JSON object with exactly these fields:
  score (int, 0-100), strengths ([]string), skills_match ([]string), gaps ([]string), suggestions ([]string).
Return only the JSON object, no extra prose.`

// AnalyzeJobMatch sends resume text and job description to OpenRouter and returns raw JSON.
func (c *Client) AnalyzeJobMatch(ctx context.Context, resumeText string, jobDescription string) (string, error) {
	userMsg := "Resume:\n" + resumeText + "\n\nJob Description:\n" + jobDescription
	body, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []message{
			{Role: "system", Content: jobMatchSystemPrompt},
			{Role: "user", Content: userMsg},
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openrouter: unexpected status %d", resp.StatusCode)
	}

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("openrouter: no choices in response")
	}
	return result.Choices[0].Message.Content, nil
}

// AnalyzeResume sends the resume text to OpenRouter and returns the raw JSON string.
func (c *Client) AnalyzeResume(ctx context.Context, resumeText string) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: resumeText},
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openrouter: status %d: %s", resp.StatusCode, string(b))
	}

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("openrouter: no choices in response")
	}
	return result.Choices[0].Message.Content, nil
}

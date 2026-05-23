// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package brain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rivernova/hosomaki/internal/config"
)

// Client talks to a local AI model via HTTP.
type Client struct {
	endpoint string
	model    string
	http     *http.Client
}

// New creates a Client. If model is empty, the config default is used.
func New(model string) *Client {
	cfg := config.C.AI
	if model == "" {
		model = cfg.Model
	}
	return &Client{
		endpoint: cfg.Endpoint,
		model:    model,
		http:     &http.Client{Timeout: 120 * time.Second},
	}
}

// Explain asks the model to explain a system message in plain language.
func (c *Client) Explain(input string) (string, error) {
	prompt := buildExplainPrompt(input)
	return c.generate(prompt)
}

// Status asks the model to summarize system health from a collected snapshot.
func (c *Client) Status(payload string, brief bool) (string, error) {
	prompt := buildStatusPrompt(payload, brief)
	return c.generate(prompt)
}

// ── Ollama API ────────────────────────────────────────────────────────────────

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func (c *Client) generate(prompt string) (string, error) {
	body, err := json.Marshal(ollamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	url := c.endpoint + "/api/generate"
	resp, err := c.http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("could not reach Ollama at %s — is it running? (ollama serve): %w", c.endpoint, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama returned %d: %s", resp.StatusCode, string(raw))
	}

	var result ollamaResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("Ollama error: %s", result.Error)
	}

	return result.Response, nil
}

// ── Prompts ───────────────────────────────────────────────────────────────────

func buildExplainPrompt(input string) string {
	return fmt.Sprintf(`You are a Linux system expert. A user has given you a system message, error, or log output.

Explain clearly and concisely what it means, why it happened, and what (if anything) the user should do about it.
Be direct. Use plain language. No markdown. No bullet points. Max 5 sentences.

System message:
%s`, input)
}

func buildStatusPrompt(payload string, brief bool) string {
	style := "Write a clear, concise paragraph (5–8 sentences) summarizing system health. Mention any anomalies or points of attention."
	if brief {
		style = "Summarize system health in a single sentence. Mention the most critical issue if any."
	}
	return fmt.Sprintf(`You are a Linux system expert. Here is a snapshot of the current system state.

%s

No markdown. No bullet points. Plain text only.

System snapshot:
%s`, style, payload)
}

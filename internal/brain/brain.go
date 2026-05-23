// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// handles communication with the local AI model
package brain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rivernova/hosomaki/internal/config"
)

type Client struct {
	endpoint   string
	model      string
	httpClient *http.Client
}

func New(model string) *Client {
	cfg := config.C.AI
	if model == "" {
		model = cfg.Model
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = config.DefaultTimeout
	}
	return &Client{
		endpoint:   cfg.Endpoint,
		model:      model,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) Explain(input string) (string, error) {
	return c.generate(buildExplainPrompt(input))
}

func (c *Client) Status(payload string, brief bool) (string, error) {
	return c.generate(buildStatusPrompt(payload, brief))
}

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
		return "", fmt.Errorf("brain: failed to build request: %w", err)
	}

	url := c.endpoint + "/api/generate"
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("brain: could not reach Ollama at %s — is it running? (ollama serve): %w", c.endpoint, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("brain: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("brain: Ollama returned HTTP %d: %s", resp.StatusCode, string(raw))
	}

	var result ollamaResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("brain: failed to parse response: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("brain: Ollama error: %s", result.Error)
	}

	return result.Response, nil
}

func buildExplainPrompt(input string) string {
	return fmt.Sprintf(`You are a Linux system expert. A user has piped log output or an error message to you.

Explain clearly and concisely:
1. What it means.
2. Why it likely happened.
3. What the user should do about it (if anything).

Rules: plain text only, no markdown, no bullet points, max 5 sentences, be direct.

Input:
%s`, input)
}

func buildStatusPrompt(payload string, brief bool) string {
	style := "Write a clear, concise paragraph (5–8 sentences) summarising system health. Highlight any anomalies or points of attention."
	if brief {
		style = "Summarise system health in a single sentence. Mention the most critical issue if any."
	}
	return fmt.Sprintf(`You are a Linux system expert. Here is a snapshot of the current system state.

%s

Rules: plain text only, no markdown, no bullet points.

System snapshot:
%s`, style, payload)
}

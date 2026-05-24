// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package ollama implements the ai.Provider interface for local Ollama models.
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client sends prompts to a local Ollama instance.
// It implements ai.Provider.
type Client struct {
	endpoint   string
	model      string
	httpClient *http.Client
}

// New returns a Client configured with the given endpoint, model and timeout.
// All three values are required; callers (typically cmd/) read them from config
// and pass them in explicitly — no global state is read here.
func New(endpoint, model string, timeout time.Duration) *Client {
	return &Client{
		endpoint:   endpoint,
		model:      model,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// Generate implements ai.Provider.
func (c *Client) Generate(_ context.Context, prompt string) (string, error) {
	body, err := json.Marshal(request{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("ollama: marshal request: %w", err)
	}

	url := c.endpoint + "/api/generate"
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama: could not reach %s — is Ollama running? (ollama serve): %w", c.endpoint, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ollama: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama: HTTP %d: %s", resp.StatusCode, string(raw))
	}

	var result response
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("ollama: parse response: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("ollama: %s", result.Error)
	}

	return result.Response, nil
}

// request is the JSON body sent to /api/generate.
type request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// response is the JSON body returned by /api/generate.
type response struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// minimal ollama client

type Client struct {
	endpoint   string
	model      string
	httpClient *http.Client
}

func New(endpoint, model string, timeout time.Duration) *Client {
	return &Client{
		endpoint:   endpoint,
		model:      model,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	return c.GenerateStream(ctx, prompt, nil, nil)
}

func (c *Client) GenerateStream(ctx context.Context, prompt string, onFirstToken func(), w io.Writer) (string, error) {
	return c.generate(ctx, prompt, "", onFirstToken, w)
}

func (c *Client) GenerateJSON(ctx context.Context, prompt string, onFirstToken func()) (string, error) {
	return c.generate(ctx, prompt, "json", onFirstToken, nil)
}

func (c *Client) generate(ctx context.Context, prompt, format string, onFirstToken func(), w io.Writer) (string, error) {
	body, err := json.Marshal(request{
		Model:  c.model,
		Prompt: prompt,
		Format: format,
		Stream: true,
	})
	if err != nil {
		return "", fmt.Errorf("ollama: marshal request: %w", err)
	}

	url := c.endpoint + "/api/generate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama: could not reach %s — is Ollama running? (ollama serve): %w", c.endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama: HTTP %d: %s", resp.StatusCode, string(raw))
	}

	var (
		full      strings.Builder
		firstSeen bool
	)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var chunk response
		if err := json.Unmarshal(line, &chunk); err != nil {
			return "", fmt.Errorf("ollama: parse chunk: %w", err)
		}
		if chunk.Error != "" {
			return "", fmt.Errorf("ollama: %s", chunk.Error)
		}

		if !firstSeen && chunk.Response != "" {
			firstSeen = true
			if onFirstToken != nil {
				onFirstToken()
			}
		}

		if w != nil {
			fmt.Fprint(w, chunk.Response)
		}
		full.WriteString(chunk.Response)

		if chunk.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("ollama: read stream: %w", err)
	}

	return full.String(), nil
}

type request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Format string `json:"format,omitempty"`
	Stream bool   `json:"stream"`
}

type response struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

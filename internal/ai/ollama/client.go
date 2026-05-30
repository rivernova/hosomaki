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
	"net"
	"net/http"
	"strings"
	"time"
)

const systemInstruction = `You are a structured data emitter. You output raw XML only.
Your response must start with <analysis> and end with </analysis>.
No text before <analysis>. No text after </analysis>.
No markdown. No explanation. No preamble. No postamble.`

type Client struct {
	endpoint    string
	model       string
	httpClient  *http.Client
	temperature float64
	numPredict  int
}

func New(endpoint, model string, timeout time.Duration, temperature float64, numPredict int) *Client {

	dialTimeout := timeout
	if dialTimeout == 0 {
		dialTimeout = 30 * time.Second
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   dialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ResponseHeaderTimeout: dialTimeout,
	}

	return &Client{
		endpoint:    endpoint,
		model:       model,
		httpClient:  &http.Client{Timeout: 0, Transport: transport},
		temperature: temperature,
		numPredict:  numPredict,
	}
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	return c.generateStream(ctx, prompt, nil, nil)
}

func (c *Client) GenerateStream(ctx context.Context, prompt string, onFirstToken func(), w io.Writer) (string, error) {
	return c.generateStream(ctx, prompt, onFirstToken, w)
}

func (c *Client) generateStream(ctx context.Context, prompt string, onFirstToken func(), w io.Writer) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model:  c.model,
		Stream: true,
		Messages: []chatMessage{
			{Role: "system", Content: systemInstruction},
			{Role: "user", Content: prompt},
		},
		Options: &requestOptions{
			Temperature:   c.temperature,
			NumPredict:    c.numPredict,
			TopP:          0.9,
			RepeatPenalty: 1.1,
		},
	})
	if err != nil {
		return "", fmt.Errorf("ollama: marshal request: %w", err)
	}

	url := c.endpoint + "/api/chat"
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

		var chunk chatResponse
		if err := json.Unmarshal(line, &chunk); err != nil {
			return "", fmt.Errorf("ollama: parse chunk: %w", err)
		}
		if chunk.Error != "" {
			return "", fmt.Errorf("ollama: %s", chunk.Error)
		}

		token := chunk.Message.Content
		if !firstSeen && token != "" {
			firstSeen = true
			if onFirstToken != nil {
				onFirstToken()
			}
		}

		if w != nil {
			fmt.Fprint(w, token)
		}
		full.WriteString(token)

		if chunk.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("ollama: read stream: %w", err)
	}

	return extractXML(full.String()), nil
}

func extractXML(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return s
	}

	s = stripMarkdownFence(s)

	if idx := strings.Index(s, "<analysis"); idx > 0 {
		s = s[idx:]
	}

	if idx := strings.LastIndex(s, "</analysis>"); idx >= 0 {
		s = s[:idx+len("</analysis>")]
	}

	return strings.TrimSpace(s)
}

func stripMarkdownFence(s string) string {
	for _, fence := range []string{"```xml", "```XML", "```"} {
		if strings.HasPrefix(s, fence) {
			s = s[len(fence):]
			s = strings.TrimPrefix(s, "\n")
			if idx := strings.LastIndex(s, "```"); idx >= 0 {
				s = s[:idx]
			}
			return strings.TrimSpace(s)
		}
	}
	return s
}

type chatRequest struct {
	Model    string          `json:"model"`
	Messages []chatMessage   `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  *requestOptions `json:"options,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Message chatMessage `json:"message"`
	Done    bool        `json:"done"`
	Error   string      `json:"error,omitempty"`
}

type requestOptions struct {
	Temperature   float64 `json:"temperature"`
	NumPredict    int     `json:"num_predict"`
	TopP          float64 `json:"top_p"`
	RepeatPenalty float64 `json:"repeat_penalty"`
}

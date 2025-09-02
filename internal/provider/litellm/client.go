package litellm

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a LiteLLM API client.
type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), APIKey: apiKey, HTTP: &http.Client{Timeout: 60 * time.Second}}
}

// ListModels fetches available models.
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list models: %s: %s", resp.Status, string(b))
	}
	var out struct {
		Data []Model `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// StreamChat starts a streaming chat completion.
// Returns a channel of string deltas and a channel for errors.
func (c *Client) StreamChat(ctx context.Context, reqPayload ChatRequest) (<-chan string, <-chan error) {
	deltas := make(chan string)
	errs := make(chan error, 1)
	go func() {
		defer close(deltas)
		defer close(errs)

		bodyBytes, err := json.Marshal(reqPayload)
		if err != nil {
			errs <- err
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/chat/completions", strings.NewReader(string(bodyBytes)))
		if err != nil {
			errs <- err
			return
		}
		req.Header.Set("Content-Type", "application/json")
		if c.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.APIKey)
		}

		resp, err := c.HTTP.Do(req)
		if err != nil {
			errs <- err
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			b, _ := io.ReadAll(resp.Body)
			errs <- errors.New(string(b))
			return
		}

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				// Expect lines like: "data: {\"choices\":[{\"delta\":{\"content\":\"...\"}}]}"
				if strings.HasPrefix(line, "data: ") {
					data := strings.TrimPrefix(line, "data: ")
					data = strings.TrimSpace(data)
					if data == "[DONE]" {
						return
					}
					var chunk struct {
						Choices []struct {
							Delta struct {
								Content string `json:"content"`
							} `json:"delta"`
						} `json:"choices"`
					}
					if err := json.Unmarshal([]byte(data), &chunk); err == nil {
						if len(chunk.Choices) > 0 {
							deltas <- chunk.Choices[0].Delta.Content
						}
					}
				}
			}
			if err != nil {
				if err == io.EOF {
					return
				}
				errs <- err
				return
			}
		}
	}()
	return deltas, errs
}

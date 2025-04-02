package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const claudeAPIEndpoint = "https://api.anthropic.com/v1/messages"
const claudeStreamAPIEndpoint = "https://api.anthropic.com/v1/messages"

type ClaudeProvider struct {
	apiKey string
	client *http.Client
}

func NewClaudeProvider(apiKey string) (*ClaudeProvider, error) {
	if apiKey == "" {
		return nil, errors.New("API key is required for Claude provider")
	}

	return &ClaudeProvider{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: time.Second * 120,
		},
	}, nil
}

func (p *ClaudeProvider) Name() ProviderType {
	return ProviderClaude
}

func (p *ClaudeProvider) AvailableModels() []string {
	return []string{
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
		"claude-2.1",
		"claude-2.0",
		"claude-instant-1.2",
	}
}

func (p *ClaudeProvider) DefaultModel() string {
	return "claude-3-haiku-20240307"
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeRequest struct {
	Model       string          `json:"model"`
	Messages    []claudeMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type claudeResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Role      string `json:"role"`
	Content   []struct {
		Type  string `json:"type"`
		Text  string `json:"text"`
	} `json:"content"`
	Model     string `json:"model"`
	StopReason string `json:"stop_reason"`
}

type claudeStreamResponse struct {
	Type      string `json:"type"`
	Delta     struct {
		Type  string `json:"type"`
		Text  string `json:"text"`
	} `json:"delta"`
}

func (p *ClaudeProvider) convertMessagesToClaudeMessages(messages []Message) []claudeMessage {
	result := make([]claudeMessage, 0)

	for _, msg := range messages {
		var role string
		switch strings.ToLower(msg.Role) {
		case "user":
			role = "user"
		case "assistant":
			role = "assistant"
		case "system":
			// Claude doesn't have system messages in the same way
			// We'll prepend this to the first user message
			continue
		}

		if role != "" {
			result = append(result, claudeMessage{
				Role:    role,
				Content: msg.Content,
			})
		}
	}

	// Handle system message - find it and prepend to first user message if found
	var systemContent string
	for _, msg := range messages {
		if strings.ToLower(msg.Role) == "system" {
			systemContent = msg.Content
			break
		}
	}

	if systemContent != "" && len(result) > 0 {
		for i, msg := range result {
			if msg.Role == "user" {
				// Prepend system message to first user message
				result[i].Content = fmt.Sprintf("%s\n\n%s", systemContent, msg.Content)
				break
			}
		}
	}

	return result
}

func (p *ClaudeProvider) CreateCompletion(ctx context.Context, req CompletionRequest) (string, error) {
	claudeMessages := p.convertMessagesToClaudeMessages(req.Messages)
	
	if len(claudeMessages) == 0 {
		return "", errors.New("no valid messages to send to Claude")
	}

	claudeReq := claudeRequest{
		Model:       req.Model,
		Messages:    claudeMessages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      false,
	}

	jsonData, err := json.Marshal(claudeReq)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", claudeAPIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Claude API returned error: %s - %s", resp.Status, string(bodyBytes))
	}

	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return "", err
	}

	// Extract content text from the response
	var contentText string
	for _, content := range claudeResp.Content {
		if content.Type == "text" {
			contentText += content.Text
		}
	}

	return contentText, nil
}

func (p *ClaudeProvider) CreateCompletionStream(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, error) {
	claudeMessages := p.convertMessagesToClaudeMessages(req.Messages)
	
	if len(claudeMessages) == 0 {
		return nil, errors.New("no valid messages to send to Claude")
	}

	claudeReq := claudeRequest{
		Model:       req.Model,
		Messages:    claudeMessages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      true,
	}

	jsonData, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", claudeStreamAPIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Claude API returned error: %s - %s", resp.Status, string(bodyBytes))
	}

	responseChan := make(chan CompletionResponse)

	go func() {
		defer resp.Body.Close()
		defer close(responseChan)

		reader := bufio.NewReader(resp.Body)
		
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					// Error occurred, make sure to send a done signal before returning
					responseChan <- CompletionResponse{
						Content: "",
						Done:    true,
					}
					return
				}
				// EOF reached
				break
			}

			// Skip empty lines
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// SSE messages start with "data: "
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			// Extract the JSON data
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				responseChan <- CompletionResponse{
					Content: "",
					Done:    true,
				}
				return
			}

			var streamResp claudeStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				// Skip malformed messages, but don't close the connection
				continue
			}

			// Only process content stream events
			if streamResp.Type != "content_block_delta" {
				continue
			}

			responseChan <- CompletionResponse{
				Content: streamResp.Delta.Text,
				Done:    false,
			}
		}

		// In case we didn't get a [DONE] message
		responseChan <- CompletionResponse{
			Content: "",
			Done:    true,
		}
	}()

	return responseChan, nil
}
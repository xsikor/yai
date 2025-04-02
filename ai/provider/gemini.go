package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	client *genai.Client
}

func NewGeminiProvider(apiKey string) (*GeminiProvider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	return &GeminiProvider{
		client: client,
	}, nil
}

func (p *GeminiProvider) Name() ProviderType {
	return ProviderGemini
}

func (p *GeminiProvider) AvailableModels() []string {
	return []string{
		"gemini-2.5-pro-exp-03-25",
		"gemini-2.0-flash",
		"gemini-2.0-flash-lite",
		"gemini-2.0-pro",
		"gemini-1.5-flash",
		"gemini-1.5-flash-8b", 
		"gemini-1.5-pro",
		"gemini-embedding-exp",
		"imagen-3.0-generate-002",
	}
}

func (p *GeminiProvider) DefaultModel() string {
	return "gemini-2.0-flash"
}

func (p *GeminiProvider) convertMessagesToGeminiParts(messages []Message) ([]genai.Part, error) {
	// For Gemini, we'll combine all messages into a single prompt
	var systemContent string
	var prompt strings.Builder

	// Extract system message if present
	for _, msg := range messages {
		if strings.ToLower(msg.Role) == "system" {
			systemContent = msg.Content
			break
		}
	}

	// If we have a system message, add it to the beginning with a separator
	if systemContent != "" {
		prompt.WriteString("System Instructions: ")
		prompt.WriteString(systemContent)
		prompt.WriteString("\n\n")
	}

	// Add the conversation history
	for _, msg := range messages {
		role := strings.ToLower(msg.Role)
		if role != "system" {
			if role == "user" {
				prompt.WriteString("User: ")
			} else if role == "assistant" {
				prompt.WriteString("Assistant: ")
			}
			prompt.WriteString(msg.Content)
			prompt.WriteString("\n\n")
		}
	}

	// Add a final prompt for the AI to continue
	prompt.WriteString("Assistant: ")

	return []genai.Part{genai.Text(prompt.String())}, nil
}

func (p *GeminiProvider) CreateCompletion(ctx context.Context, req CompletionRequest) (string, error) {
	model := p.client.GenerativeModel(req.Model)
	model.SetTemperature(float32(req.Temperature))
	if req.MaxTokens > 0 {
		model.SetMaxOutputTokens(int32(req.MaxTokens))
	}

	parts, err := p.convertMessagesToGeminiParts(req.Messages)
	if err != nil {
		return "", err
	}

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no content generated")
	}

	// Extract the text from the response
	content, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response format from Gemini API")
	}

	return string(content), nil
}

func (p *GeminiProvider) CreateCompletionStream(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, error) {
	model := p.client.GenerativeModel(req.Model)
	model.SetTemperature(float32(req.Temperature))
	if req.MaxTokens > 0 {
		model.SetMaxOutputTokens(int32(req.MaxTokens))
	}

	parts, err := p.convertMessagesToGeminiParts(req.Messages)
	if err != nil {
		return nil, err
	}

	iter := model.GenerateContentStream(ctx, parts...)
	responseChan := make(chan CompletionResponse)

	go func() {
		defer close(responseChan)

		for {
			resp, err := iter.Next()
			if err != nil {
				if err == io.EOF {
					// Send final token with done flag
					responseChan <- CompletionResponse{
						Content: "",
						Done:    true,
					}
					return
				}
				// Other error occurred, but still send a done signal to prevent hanging
				responseChan <- CompletionResponse{
					Content: "",
					Done:    true,
				}
				return
			}

			if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
				continue
			}

			// Extract the text from the response
			content, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
			if !ok {
				// Skip non-text parts
				continue
			}

			responseChan <- CompletionResponse{
				Content: string(content),
				Done:    false,
			}
		}
	}()

	return responseChan, nil
}
package provider

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	client *openai.Client
}

func NewOpenAIProvider(apiKey string, proxyURL string) (*OpenAIProvider, error) {
	var client *openai.Client

	if proxyURL != "" {
		clientConfig := openai.DefaultConfig(apiKey)

		proxyUrl, err := url.Parse(proxyURL)
		if err != nil {
			return nil, err
		}

		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}

		clientConfig.HTTPClient = &http.Client{
			Transport: transport,
		}

		client = openai.NewClientWithConfig(clientConfig)
	} else {
		client = openai.NewClient(apiKey)
	}

	return &OpenAIProvider{
		client: client,
	}, nil
}

func (p *OpenAIProvider) Name() ProviderType {
	return ProviderOpenAI
}

func (p *OpenAIProvider) AvailableModels() []string {
	return []string{
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
		"gpt-4",
		"gpt-4-32k",
		"gpt-4-turbo",
	}
}

func (p *OpenAIProvider) DefaultModel() string {
	return "gpt-3.5-turbo"
}

func (p *OpenAIProvider) CreateCompletion(ctx context.Context, req CompletionRequest) (string, error) {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	resp, err := p.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       req.Model,
			MaxTokens:   req.MaxTokens,
			Temperature: float32(req.Temperature),
			Messages:    messages,
		},
	)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (p *OpenAIProvider) CreateCompletionStream(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, error) {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	stream, err := p.client.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:       req.Model,
			MaxTokens:   req.MaxTokens,
			Temperature: float32(req.Temperature),
			Messages:    messages,
			Stream:      true,
		},
	)
	if err != nil {
		return nil, err
	}

	responseChan := make(chan CompletionResponse)

	go func() {
		defer stream.Close()
		defer close(responseChan)

		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				responseChan <- CompletionResponse{
					Content: "",
					Done:    true,
				}
				return
			}

			if err != nil {
				// Always send a done signal on error to prevent hanging
				responseChan <- CompletionResponse{
					Content: "",
					Done:    true,
				}
				return
			}

			delta := resp.Choices[0].Delta.Content
			responseChan <- CompletionResponse{
				Content: delta,
				Done:    false,
			}
		}
	}()

	return responseChan, nil
}
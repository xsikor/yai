package provider

import (
	"context"
)

type ProviderType string

const (
	ProviderOpenAI ProviderType = "openai"
	ProviderClaude ProviderType = "claude"
	ProviderGemini ProviderType = "gemini"
)

type Message struct {
	Role    string
	Content string
}

type CompletionRequest struct {
	Model       string
	Messages    []Message
	MaxTokens   int
	Temperature float64
	Stream      bool
}

type CompletionResponse struct {
	Content    string
	Done       bool
	Executable bool
}

type Provider interface {
	Name() ProviderType
	AvailableModels() []string
	DefaultModel() string
	CreateCompletion(ctx context.Context, req CompletionRequest) (string, error)
	CreateCompletionStream(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, error)
}
# Action Plan for Adding Anthropic Claude and Google Gemini Support

## Overview
This plan outlines the steps needed to add support for additional AI providers (Anthropic Claude and Google Gemini) to the Yai terminal assistant. Currently, Yai only supports OpenAI's GPT models through the go-openai client.

## Phase 1: Refactor Existing Code

1. Create a new provider interface and common abstractions:
   - Create `ai/provider/interface.go` with a common Provider interface
   - Define message and model structs that are provider-agnostic
   - Refactor `ai/engine.go` to use the provider interface

2. Update configuration system:
   - Modify `config/ai.go` to support provider type selection
   - Add provider-specific configuration options
   - Update configuration loading/saving in `config/config.go`

## Phase 2: Add Provider Implementations

3. Implement Anthropic Claude provider:
   - Add Go client for Anthropic API (possibly github.com/anthropics/anthropic-sdk-golang)
   - Create `ai/provider/claude.go` implementing the Provider interface
   - Support Claude's message format and streaming capabilities
   - Map Claude-specific models

4. Implement Google Gemini provider:
   - Add Go client for Google Gemini API using "github.com/google/generative-ai-go/genai" and "google.golang.org/api/option"
   - Create `ai/provider/gemini.go` implementing the Provider interface
   - Support Gemini's message format and streaming capabilities
   - Map Gemini-specific models

5. Refactor OpenAI provider:
   - Move OpenAI-specific code to `ai/provider/openai.go`
   - Make it implement the Provider interface

## Phase 3: Update User Interface

6. Enhance configuration management:
   - Add provider selection to first-run experience
   - Update configuration file format
   - Ensure backward compatibility

7. Update UI components:
   - Modify model selection to be provider-aware
   - Add provider-specific settings
   - Update help information to include new providers

## Phase 4: Testing and Documentation

8. Create unit tests:
   - Test provider implementations
   - Test configuration handling
   - Mock API responses for testing

9. Update documentation:
   - Update README.md with new provider information
   - Add configuration examples for each provider
   - Document model capabilities and limitations

## Implementation Details

### Provider Interface (Draft)
```go
// ai/provider/interface.go
package provider

type ProviderType string

const (
    ProviderOpenAI    ProviderType = "openai"
    ProviderClaude    ProviderType = "claude"
    ProviderGemini    ProviderType = "gemini"
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
    Content string
    Done    bool
}

type Provider interface {
    Name() ProviderType
    AvailableModels() []string
    DefaultModel() string
    CreateCompletion(ctx context.Context, req CompletionRequest) (string, error)
    CreateCompletionStream(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, error)
}
```

### Configuration Updates (Draft)
```go
// config/ai.go
type AiConfig struct {
    provider    ProviderType
    key         string
    model       string
    proxy       string
    temperature float64
    maxTokens   int
}

func (c AiConfig) GetProvider() ProviderType {
    return c.provider
}
```

### Estimated Timeline
- Phase 1: 2-3 days
- Phase 2: 4-5 days
- Phase 3: 2-3 days
- Phase 4: 2-3 days

Total: Approximately 2 weeks for full implementation
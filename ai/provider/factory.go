package provider

import "fmt"

// CreateProvider creates and returns the appropriate provider based on the type
func CreateProvider(providerType ProviderType, apiKey string, proxyURL string) (Provider, error) {
	switch providerType {
	case ProviderOpenAI:
		return NewOpenAIProvider(apiKey, proxyURL)
	case ProviderClaude:
		return NewClaudeProvider(apiKey)
	case ProviderGemini:
		return NewGeminiProvider(apiKey)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}
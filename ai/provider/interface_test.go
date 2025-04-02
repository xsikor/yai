package provider

import (
	"testing"
)

func TestProviderNameMethod(t *testing.T) {
	// Create mock providers
	openaiProvider, _ := NewOpenAIProvider("fake-key", "")
	claudeProvider, _ := NewClaudeProvider("fake-key")
	geminiProvider, _ := NewGeminiProvider("fake-key")

	// Test each provider's Name() method
	if openaiProvider.Name() != ProviderOpenAI {
		t.Errorf("Expected openai provider name to be %s, got %s", ProviderOpenAI, openaiProvider.Name())
	}

	if claudeProvider.Name() != ProviderClaude {
		t.Errorf("Expected claude provider name to be %s, got %s", ProviderClaude, claudeProvider.Name())
	}

	if geminiProvider.Name() != ProviderGemini {
		t.Errorf("Expected gemini provider name to be %s, got %s", ProviderGemini, geminiProvider.Name())
	}
}

func TestProviderDefaultModel(t *testing.T) {
	// Create mock providers
	openaiProvider, _ := NewOpenAIProvider("fake-key", "")
	claudeProvider, _ := NewClaudeProvider("fake-key")
	geminiProvider, _ := NewGeminiProvider("fake-key")

	// Each provider should return a non-empty default model
	if openaiProvider.DefaultModel() == "" {
		t.Error("OpenAI provider returned empty default model")
	}

	if claudeProvider.DefaultModel() == "" {
		t.Error("Claude provider returned empty default model")
	}

	if geminiProvider.DefaultModel() == "" {
		t.Error("Gemini provider returned empty default model")
	}
}

func TestProviderAvailableModels(t *testing.T) {
	// Create mock providers
	openaiProvider, _ := NewOpenAIProvider("fake-key", "")
	claudeProvider, _ := NewClaudeProvider("fake-key")
	geminiProvider, _ := NewGeminiProvider("fake-key")

	// Each provider should return a non-empty list of models
	if len(openaiProvider.AvailableModels()) == 0 {
		t.Error("OpenAI provider returned no available models")
	}

	if len(claudeProvider.AvailableModels()) == 0 {
		t.Error("Claude provider returned no available models")
	}

	if len(geminiProvider.AvailableModels()) == 0 {
		t.Error("Gemini provider returned no available models")
	}
}

func TestProviderFactory(t *testing.T) {
	// Test factory with each provider type
	openaiProvider, err := CreateProvider(ProviderOpenAI, "fake-key", "")
	if err != nil {
		t.Fatalf("Failed to create OpenAI provider: %v", err)
	}
	if openaiProvider.Name() != ProviderOpenAI {
		t.Errorf("Factory returned wrong provider type for OpenAI")
	}

	claudeProvider, err := CreateProvider(ProviderClaude, "fake-key", "")
	if err != nil {
		t.Fatalf("Failed to create Claude provider: %v", err)
	}
	if claudeProvider.Name() != ProviderClaude {
		t.Errorf("Factory returned wrong provider type for Claude")
	}

	geminiProvider, err := CreateProvider(ProviderGemini, "fake-key", "")
	if err != nil {
		t.Fatalf("Failed to create Gemini provider: %v", err)
	}
	if geminiProvider.Name() != ProviderGemini {
		t.Errorf("Factory returned wrong provider type for Gemini")
	}

	// Test with an invalid provider type
	_, err = CreateProvider("invalid-provider", "fake-key", "")
	if err == nil {
		t.Error("Factory did not return error for invalid provider type")
	}
}
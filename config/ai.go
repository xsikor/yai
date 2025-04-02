package config

import "github.com/xsikor/yai/ai/provider"

const (
	// Keys for configuration
	ai_provider    = "AI_PROVIDER"
	ai_key         = "AI_KEY"
	ai_model       = "AI_MODEL"
	ai_proxy       = "AI_PROXY"
	ai_temperature = "AI_TEMPERATURE"
	ai_max_tokens  = "AI_MAX_TOKENS"

	// Legacy keys for backward compatibility
	openai_key         = "OPENAI_KEY"
	openai_model       = "OPENAI_MODEL"
	openai_proxy       = "OPENAI_PROXY"
	openai_temperature = "OPENAI_TEMPERATURE"
	openai_max_tokens  = "OPENAI_MAX_TOKENS"
)

type AiConfig struct {
	providerType provider.ProviderType
	key          string
	model        string
	proxy        string
	temperature  float64
	maxTokens    int
}

func (c AiConfig) GetProviderType() provider.ProviderType {
	return c.providerType
}

func (c AiConfig) GetKey() string {
	return c.key
}

func (c AiConfig) GetModel() string {
	return c.model
}

func (c AiConfig) GetProxy() string {
	return c.proxy
}

func (c AiConfig) GetTemperature() float64 {
	return c.temperature
}

func (c AiConfig) GetMaxTokens() int {
	return c.maxTokens
}

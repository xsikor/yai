package config

import (
	"fmt"
	"strings"

	"github.com/ekkinox/yai/ai/provider"
	"github.com/ekkinox/yai/system"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

type Config struct {
	ai     AiConfig
	user   UserConfig
	system *system.Analysis
}

func (c *Config) GetAiConfig() AiConfig {
	return c.ai
}

func (c *Config) GetUserConfig() UserConfig {
	return c.user
}

func (c *Config) GetSystemConfig() *system.Analysis {
	return c.system
}

func NewConfig() (*Config, error) {
	system := system.Analyse()

	viper.SetConfigName(strings.ToLower(system.GetApplicationName()))
	viper.AddConfigPath(fmt.Sprintf("%s/.config/", system.GetHomeDirectory()))

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Check for provider configuration
	var providerType provider.ProviderType
	if viper.IsSet(ai_provider) {
		providerType = provider.ProviderType(viper.GetString(ai_provider))
	} else {
		// Default to OpenAI for backward compatibility
		providerType = provider.ProviderOpenAI
	}

	// Get API key based on provider
	var apiKey string
	if viper.IsSet(ai_key) {
		apiKey = viper.GetString(ai_key)
	} else {
		// Fall back to legacy OpenAI key for backward compatibility
		apiKey = viper.GetString(openai_key)
	}

	// Get model based on provider
	var model string
	if viper.IsSet(ai_model) {
		model = viper.GetString(ai_model)
	} else {
		// Fall back to legacy OpenAI model for backward compatibility
		model = viper.GetString(openai_model)
	}

	// Get other settings with new keys, falling back to legacy keys
	var proxy string
	if viper.IsSet(ai_proxy) {
		proxy = viper.GetString(ai_proxy)
	} else {
		proxy = viper.GetString(openai_proxy)
	}

	var temperature float64
	if viper.IsSet(ai_temperature) {
		temperature = viper.GetFloat64(ai_temperature)
	} else {
		temperature = viper.GetFloat64(openai_temperature)
	}

	var maxTokens int
	if viper.IsSet(ai_max_tokens) {
		maxTokens = viper.GetInt(ai_max_tokens)
	} else {
		maxTokens = viper.GetInt(openai_max_tokens)
	}

	return &Config{
		ai: AiConfig{
			providerType: providerType,
			key:          apiKey,
			model:        model,
			proxy:        proxy,
			temperature:  temperature,
			maxTokens:    maxTokens,
		},
		user: UserConfig{
			defaultPromptMode: viper.GetString(user_default_prompt_mode),
			preferences:       viper.GetString(user_preferences),
		},
		system: system,
	}, nil
}

func WriteConfig(providerType provider.ProviderType, key string, model string, write bool) (*Config, error) {
	system := system.Analyse()

	// Set provider type
	viper.Set(ai_provider, string(providerType))
	
	// Set AI config values
	viper.Set(ai_key, key)
	viper.Set(ai_model, model)
	viper.SetDefault(ai_proxy, "")
	viper.SetDefault(ai_temperature, 0.2)
	viper.SetDefault(ai_max_tokens, 1000)
	
	// Set legacy config for backward compatibility
	if providerType == provider.ProviderOpenAI {
		viper.Set(openai_key, key)
		viper.Set(openai_model, model)
	}

	// user defaults
	viper.SetDefault(user_default_prompt_mode, "exec")
	viper.SetDefault(user_preferences, "")

	if write {
		err := viper.SafeWriteConfigAs(system.GetConfigFile())
		if err != nil {
			return nil, err
		}
	}

	return NewConfig()
}

// Helper method to get default model for a provider
func GetDefaultModelForProvider(providerType provider.ProviderType) string {
	switch providerType {
	case provider.ProviderOpenAI:
		return openai.GPT3Dot5Turbo
	case provider.ProviderClaude:
		return "claude-3-haiku-20240307"
	case provider.ProviderGemini:
		return "gemini-2.0-flash"
	default:
		return openai.GPT3Dot5Turbo
	}
}

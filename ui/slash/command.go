package slash

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ekkinox/yai/ai/provider"
	"github.com/ekkinox/yai/config"
)

// SlashCommand represents a command that can be executed with a / prefix
type SlashCommand struct {
	Name        string
	Description string
	Execute     func(config *config.Config, args string) string
}

// SlashCommands is a collection of all available slash commands
var SlashCommands []SlashCommand

// InitSlashCommands initializes all slash commands
func InitSlashCommands() {
	SlashCommands = []SlashCommand{
		{
			Name:        "help",
			Description: "Show available slash commands",
			Execute: func(config *config.Config, args string) string {
				return formatHelpOutput()
			},
		},
		{
			Name:        "config",
			Description: "Show current configuration",
			Execute: func(config *config.Config, args string) string {
				return formatConfigOutput(config)
			},
		},
		{
			Name:        "models",
			Description: "Show available AI models for current provider",
			Execute: func(config *config.Config, args string) string {
				return formatModelsOutput(config)
			},
		},
		{
			Name:        "providers",
			Description: "Show available AI providers",
			Execute: func(config *config.Config, args string) string {
				return formatProvidersOutput()
			},
		},
		{
			Name:        "clear",
			Description: "Clear the screen",
			Execute: func(config *config.Config, args string) string {
				return "[clear]"
			},
		},
		{
			Name:        "reset",
			Description: "Reset conversation history",
			Execute: func(config *config.Config, args string) string {
				return "[reset]"
			},
		},
		{
			Name:        "mode",
			Description: "Switch between chat and exec modes",
			Execute: func(config *config.Config, args string) string {
				return "[mode]"
			},
		},
	}
}

// IsSlashCommand checks if the input is a slash command
func IsSlashCommand(input string) bool {
	return strings.HasPrefix(input, "/")
}

// GetCommand returns the slash command if found
func GetCommand(input string) *SlashCommand {
	if !IsSlashCommand(input) {
		return nil
	}

	// Extract command name and arguments
	parts := strings.SplitN(strings.TrimPrefix(input, "/"), " ", 2)
	cmdName := parts[0]

	for _, cmd := range SlashCommands {
		if cmd.Name == cmdName {
			return &cmd
		}
	}

	return nil
}

// ExecuteCommand executes a slash command and returns the output
func ExecuteCommand(config *config.Config, input string) string {
	cmd := GetCommand(input)
	if cmd == nil {
		return fmt.Sprintf("Unknown command: %s\nType /help for available commands.", input)
	}

	// Extract arguments if any
	parts := strings.SplitN(strings.TrimPrefix(input, "/"), " ", 2)
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	return cmd.Execute(config, args)
}

// GetCompletions returns possible command completions for the current input
func GetCompletions(input string) []string {
	if !IsSlashCommand(input) {
		return nil
	}

	// Get the command name being typed
	prefix := strings.TrimPrefix(input, "/")
	
	// If the user is typing arguments, return empty
	if strings.Contains(prefix, " ") {
		return nil
	}

	// Find matching commands
	var matches []string
	for _, cmd := range SlashCommands {
		if strings.HasPrefix(cmd.Name, prefix) {
			matches = append(matches, "/"+cmd.Name)
		}
	}

	// Sort alphabetically
	sort.Strings(matches)
	return matches
}

// Format helpers
func formatHelpOutput() string {
	var sb strings.Builder
	
	sb.WriteString("## Available Commands\n\n")
	
	for _, cmd := range SlashCommands {
		sb.WriteString(fmt.Sprintf("- `/%s`: %s\n", cmd.Name, cmd.Description))
	}
	
	sb.WriteString("\nType any command to execute it.")
	
	return sb.String()
}

func formatConfigOutput(cfg *config.Config) string {
	var sb strings.Builder
	
	sb.WriteString("## Current Configuration\n\n")
	
	// AI Provider Info
	sb.WriteString(fmt.Sprintf("**Provider**: %s\n", cfg.GetAiConfig().GetProviderType()))
	sb.WriteString(fmt.Sprintf("**Model**: %s\n", cfg.GetAiConfig().GetModel()))
	sb.WriteString(fmt.Sprintf("**Temperature**: %.2f\n", cfg.GetAiConfig().GetTemperature()))
	sb.WriteString(fmt.Sprintf("**Max Tokens**: %d\n", cfg.GetAiConfig().GetMaxTokens()))
	
	// System Info
	sb.WriteString("\n**System Information**\n")
	sb.WriteString(fmt.Sprintf("- OS: %s\n", cfg.GetSystemConfig().GetOperatingSystem().String()))
	sb.WriteString(fmt.Sprintf("- Distribution: %s\n", cfg.GetSystemConfig().GetDistribution()))
	sb.WriteString(fmt.Sprintf("- Shell: %s\n", cfg.GetSystemConfig().GetShell()))
	sb.WriteString(fmt.Sprintf("- Editor: %s\n", cfg.GetSystemConfig().GetEditor()))
	
	// User Preferences
	sb.WriteString("\n**User Preferences**\n")
	sb.WriteString(fmt.Sprintf("- Default Mode: %s\n", cfg.GetUserConfig().GetDefaultPromptMode()))
	
	if cfg.GetUserConfig().GetPreferences() != "" {
		sb.WriteString(fmt.Sprintf("- Custom Preferences: %s\n", cfg.GetUserConfig().GetPreferences()))
	}
	
	return sb.String()
}

func formatModelsOutput(cfg *config.Config) string {
	var sb strings.Builder
	
	providerType := cfg.GetAiConfig().GetProviderType()
	currentModel := cfg.GetAiConfig().GetModel()
	
	sb.WriteString(fmt.Sprintf("## Available Models for %s\n\n", providerType))
	
	// Create provider instance to get available models
	p, err := provider.CreateProvider(
		providerType,
		cfg.GetAiConfig().GetKey(),
		cfg.GetAiConfig().GetProxy(),
	)
	
	if err != nil {
		return fmt.Sprintf("Error getting models: %s", err.Error())
	}
	
	models := p.AvailableModels()
	
	for i, model := range models {
		if model == currentModel {
			sb.WriteString(fmt.Sprintf("- **%s** (current)\n", model))
		} else {
			sb.WriteString(fmt.Sprintf("- %s\n", model))
		}
		
		// Add description based on model name
		if i < len(models)-1 {
			sb.WriteString("\n")
		}
	}
	
	return sb.String()
}

func formatProvidersOutput() string {
	var sb strings.Builder
	
	sb.WriteString("## Available AI Providers\n\n")
	
	// List all providers
	sb.WriteString("- **OpenAI** (GPT models)\n")
	sb.WriteString("  - Use `-p openai` flag to select\n")
	sb.WriteString("  - Default model: gpt-3.5-turbo\n\n")
	
	sb.WriteString("- **Google Gemini** (Gemini models)\n")
	sb.WriteString("  - Use `-p gemini` flag to select\n")
	sb.WriteString("  - Default model: gemini-2.0-flash\n\n")
	
	sb.WriteString("- **Anthropic Claude** (Claude models)\n")
	sb.WriteString("  - Use `-p claude` flag to select\n")
	sb.WriteString("  - Default model: claude-3-haiku-20240307\n")
	
	return sb.String()
}
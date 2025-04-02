package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/xsikor/yai/config"
	"github.com/xsikor/yai/ui/slash"
)

const (
	exec_icon            = "🚀 > "
	exec_placeholder     = "Execute something..."
	config_icon          = "🔒 > "
	config_placeholder   = "Enter your API key..."
	provider_icon        = "🤖 > "
	provider_placeholder = "Select provider (openai, claude, gemini)..."
	model_icon           = "📦 > "
	model_placeholder    = "Select model (press Enter for default)..."
	chat_icon            = "💬 > "
	chat_placeholder     = "Ask me something..."
)

type Prompt struct {
	mode         PromptMode
	input        textinput.Model
	autocomplete *slash.AutocompleteState
}

func NewPrompt(mode PromptMode) *Prompt {
	input := textinput.New()
	input.Placeholder = getPromptPlaceholder(mode)
	input.TextStyle = getPromptStyle(mode)
	input.Prompt = getPromptIcon(mode)

	if mode == ConfigPromptMode {
		input.EchoMode = textinput.EchoPassword
	}

	input.Focus()

	// Initialize slash commands
	slash.InitSlashCommands()

	return &Prompt{
		mode:         mode,
		input:        input,
		autocomplete: slash.NewAutocompleteState(),
	}
}

func (p *Prompt) GetMode() PromptMode {
	return p.mode
}

func (p *Prompt) SetMode(mode PromptMode) *Prompt {
	p.mode = mode

	p.input.TextStyle = getPromptStyle(mode)
	p.input.Prompt = getPromptIcon(mode)
	p.input.Placeholder = getPromptPlaceholder(mode)

	return p
}

func (p *Prompt) SetValue(value string) *Prompt {
	p.input.SetValue(value)

	return p
}

func (p *Prompt) GetValue() string {
	return p.input.Value()
}

func (p *Prompt) Blur() *Prompt {
	p.input.Blur()

	return p
}

func (p *Prompt) Focus() *Prompt {
	p.input.Focus()

	return p
}

func (p *Prompt) Update(msg tea.Msg) (*Prompt, tea.Cmd) {
	var updateCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			// Handle autocomplete
			currentValue := p.input.Value()
			if strings.HasPrefix(currentValue, "/") {
				if !p.autocomplete.Active {
					p.autocomplete.StartAutocomplete(currentValue)
				}

				if p.autocomplete.Active {
					suggestion := p.autocomplete.GetCurrentSuggestion()
					if suggestion != "" {
						p.input.SetValue(suggestion)
						return p, nil
					}
				}
			}
		case tea.KeyShiftTab:
			// Cycle through previous suggestions
			if p.autocomplete.Active {
				suggestion := p.autocomplete.PrevSuggestion()
				if suggestion != "" {
					p.input.SetValue(suggestion)
					return p, nil
				}
			}
		case tea.KeyEnter, tea.KeyEsc, tea.KeyCtrlC:
			// Reset autocomplete
			p.autocomplete.Reset()
		default:
			// If the user types, update autocomplete suggestions
			if p.mode == ChatPromptMode || p.mode == ExecPromptMode {
				currentValue := p.input.Value()

				// Special handling for the first slash character
				if currentValue == "/" {
					p.autocomplete.StartAutocomplete(currentValue)
				} else if strings.HasPrefix(currentValue, "/") {
					p.autocomplete.StartAutocomplete(currentValue)
				} else {
					p.autocomplete.Reset()
				}
			}
		}
	}

	p.input, updateCmd = p.input.Update(msg)
	return p, updateCmd
}

// HasActiveAutocomplete returns true if autocomplete is active
func (p *Prompt) HasActiveAutocomplete() bool {
	return p.autocomplete.Active
}

// GetAutocompleteSuggestions returns formatted autocomplete suggestions
func (p *Prompt) GetAutocompleteSuggestions() string {
	return p.autocomplete.FormatSuggestions()
}

// IsSlashCommand checks if the current input is a slash command
func (p *Prompt) IsSlashCommand() bool {
	return slash.IsSlashCommand(p.input.Value())
}

// ExecuteSlashCommand executes the current slash command with the given config
func (p *Prompt) ExecuteSlashCommand(config *config.Config) string {
	return slash.ExecuteCommand(config, p.input.Value())
}

func (p *Prompt) View() string {
	return p.input.View()
}

func (p *Prompt) AsString() string {
	style := getPromptStyle(p.mode)

	return fmt.Sprintf("%s%s", style.Render(getPromptIcon(p.mode)), style.Render(p.input.Value()))
}

func getPromptStyle(mode PromptMode) lipgloss.Style {
	switch mode {
	case ExecPromptMode:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(exec_color))
	case ConfigPromptMode:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(config_color))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(chat_color))
	}
}

func getPromptIcon(mode PromptMode) string {
	style := getPromptStyle(mode)

	switch mode {
	case ExecPromptMode:
		return style.Render(exec_icon)
	case ConfigPromptMode:
		return style.Render(config_icon)
	case ProviderPromptMode:
		return style.Render(provider_icon)
	case ModelPromptMode:
		return style.Render(model_icon)
	default:
		return style.Render(chat_icon)
	}
}

func getPromptPlaceholder(mode PromptMode) string {
	switch mode {
	case ExecPromptMode:
		return exec_placeholder
	case ConfigPromptMode:
		return config_placeholder
	case ProviderPromptMode:
		return provider_placeholder
	case ModelPromptMode:
		return model_placeholder
	default:
		return chat_placeholder
	}
}

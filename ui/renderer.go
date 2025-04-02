package ui

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

const (
	exec_color    = "#ffa657"
	config_color  = "#ffffff"
	chat_color    = "#66b3ff"
	help_color    = "#aaaaaa"
	error_color   = "#cc3333"
	warning_color = "#ffcc00"
	success_color = "#46b946"
)

type Renderer struct {
	contentRenderer *glamour.TermRenderer
	successRenderer lipgloss.Style
	warningRenderer lipgloss.Style
	errorRenderer   lipgloss.Style
	helpRenderer    lipgloss.Style
}

func NewRenderer(options ...glamour.TermRendererOption) *Renderer {
	contentRenderer, err := glamour.NewTermRenderer(options...)
	if err != nil {
		return nil
	}

	successRenderer := lipgloss.NewStyle().Foreground(lipgloss.Color(success_color))
	warningRenderer := lipgloss.NewStyle().Foreground(lipgloss.Color(warning_color))
	errorRenderer := lipgloss.NewStyle().Foreground(lipgloss.Color(error_color))
	helpRenderer := lipgloss.NewStyle().Foreground(lipgloss.Color(help_color)).Italic(true)

	return &Renderer{
		contentRenderer: contentRenderer,
		successRenderer: successRenderer,
		warningRenderer: warningRenderer,
		errorRenderer:   errorRenderer,
		helpRenderer:    helpRenderer,
	}
}

func (r *Renderer) RenderContent(in string) string {
	out, _ := r.contentRenderer.Render(in)

	return out
}

func (r *Renderer) RenderSuccess(in string) string {
	return r.successRenderer.Render(in)
}

func (r *Renderer) RenderWarning(in string) string {
	return r.warningRenderer.Render(in)
}

func (r *Renderer) RenderError(in string) string {
	return r.errorRenderer.Render(in)
}

func (r *Renderer) RenderHelp(in string) string {
	return r.helpRenderer.Render(in)
}

func (r *Renderer) RenderConfigMessage() string {
	welcome := "Welcome! 👋  \n\n"
	welcome += "I cannot find a configuration file, please first select an AI provider:\n\n"
	welcome += "1. OpenAI (GPT models)\n"
	welcome += "2. Google Gemini\n"
	welcome += "3. Anthropic Claude\n\n"
	welcome += "Enter a number (1-3, default: 1): "

	return welcome
}

func (r *Renderer) RenderApiKeyMessage() string {
	welcome := "Please enter an API key for your selected provider.\n\n"
	welcome += "For OpenAI, get a key from https://platform.openai.com/account/api-keys\n"
	welcome += "For Google Gemini, get a key from https://ai.google.dev/\n"
	welcome += "For Anthropic Claude, get a key from https://console.anthropic.com/\n"

	return welcome
}

func (r *Renderer) RenderModelMessage(provider string) string {
	message := "Select a model for " + provider + ":\n\n"
	
	if provider == "openai" {
		message += "1. gpt-3.5-turbo (Default, fast & cost-effective)\n"
		message += "2. gpt-4 (More powerful reasoning)\n"
		message += "3. gpt-4-turbo (Latest model, improved capabilities)\n\n"
		message += "Enter a number (1-3, default: 1): "
	} else if provider == "gemini" {
		message += "1. gemini-2.0-flash (Default, fast, real-time streaming)\n"
		message += "2. gemini-2.5-pro-exp-03-25 (Experimental, advanced reasoning)\n"
		message += "3. gemini-2.0-pro (Advanced features)\n"
		message += "4. gemini-2.0-flash-lite (Cost efficient, low latency)\n"
		message += "5. gemini-1.5-pro (Complex reasoning)\n"
		message += "6. gemini-1.5-flash (Fast, versatile performance)\n"
		message += "7. gemini-1.5-flash-8b (High volume tasks)\n\n"
		message += "Enter a number (1-7, default: 1): "
	} else if provider == "claude" {
		message += "1. claude-3-haiku-20240307 (Default, fast & cost-effective)\n"
		message += "2. claude-3-sonnet-20240229 (Balanced power & speed)\n"
		message += "3. claude-3-opus-20240229 (Most powerful model)\n\n"
		message += "Enter a number (1-3, default: 1): "
	}
	
	return message
}

func (r *Renderer) RenderHelpMessage() string {
	help := "**Keyboard Shortcuts**\n"
	help += "- `↑`/`↓` : navigate in history\n"
	help += "- `tab`   : switch between `🚀 exec` and `💬 chat` prompt modes\n"
	help += "- `ctrl+h`: show help\n"
	help += "- `ctrl+s`: edit settings\n"
	help += "- `ctrl+r`: clear terminal and reset discussion history\n"
	help += "- `ctrl+l`: clear terminal but keep discussion history\n"
	help += "- `ctrl+c`: exit or interrupt command execution\n\n"
	
	help += "**Slash Commands**\n"
	help += "- `/help`: show available slash commands\n" 
	help += "- `/config`: show current configuration\n"
	help += "- `/models`: show available AI models\n"
	help += "- `/providers`: show available AI providers\n"
	help += "- `/mode`: toggle between chat and execute modes\n"
	help += "- `/clear`: clear the screen\n"
	help += "- `/reset`: reset conversation history\n\n"
	help += "Type a slash command and use Tab to autocomplete.\n\n"
	
	help += "**CLI Options**\n"
	help += "- `-e`: use exec prompt mode\n"
	help += "- `-c`: use chat prompt mode\n"
	help += "- `-p`: select AI provider (openai, claude, gemini)\n"
	help += "- `-model`: specify AI model to use\n"
	help += "- `-m`: show current AI model and provider\n"

	return help
}

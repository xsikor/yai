package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/spf13/viper"

	"github.com/xsikor/yai/ai"
	"github.com/xsikor/yai/ai/provider"
	"github.com/xsikor/yai/config"
	"github.com/xsikor/yai/history"
	"github.com/xsikor/yai/run"
)

type UiState struct {
	error        error
	runMode      RunMode
	promptMode   PromptMode
	providerType provider.ProviderType
	modelName    string
	configuring  bool
	querying     bool
	confirming   bool
	executing    bool
	args         string
	pipe         string
	buffer       string
	command      string
}

type UiDimensions struct {
	width  int
	height int
}

type UiComponents struct {
	prompt   *Prompt
	renderer *Renderer
	spinner  *Spinner
}

type Ui struct {
	state      UiState
	dimensions UiDimensions
	components UiComponents
	config     *config.Config
	engine     *ai.Engine
	history    *history.History
}

func NewUi(input *UiInput) *Ui {
	return &Ui{
		state: UiState{
			error:        nil,
			runMode:      input.GetRunMode(),
			promptMode:   input.GetPromptMode(),
			providerType: input.GetProviderType(),
			modelName:    input.GetModelName(),
			configuring:  false,
			querying:     false,
			confirming:   false,
			executing:    false,
			args:         input.GetArgs(),
			pipe:         input.GetPipe(),
			buffer:       "",
			command:      "",
		},
		dimensions: UiDimensions{
			150,
			150,
		},
		components: UiComponents{
			prompt: NewPrompt(input.GetPromptMode()),
			renderer: NewRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(150),
			),
			spinner: NewSpinner(),
		},
		history: history.NewHistory(),
	}
}

func (u *Ui) Init() tea.Cmd {
	config, err := config.NewConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if u.state.runMode == ReplMode {
				return tea.Sequence(
					tea.ClearScreen,
					u.startConfig(),
				)
			} else {
				return u.startConfig()
			}
		} else {
			return tea.Sequence(
				tea.Println(u.components.renderer.RenderError(err.Error())),
				tea.Quit,
			)
		}
	}

	if u.state.runMode == ReplMode {
		return u.startRepl(config)
	} else {
		return u.startCli(config)
	}
}

func (u *Ui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds       []tea.Cmd
		promptCmd  tea.Cmd
		spinnerCmd tea.Cmd
	)

	switch msg := msg.(type) {
	// spinner
	case spinner.TickMsg:
		if u.state.querying {
			u.components.spinner, spinnerCmd = u.components.spinner.Update(msg)
			cmds = append(
				cmds,
				spinnerCmd,
			)
		}
	// size
	case tea.WindowSizeMsg:
		u.dimensions.width = msg.Width
		u.dimensions.height = msg.Height
		u.components.renderer = NewRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(u.dimensions.width),
		)
	// keyboard
	case tea.KeyMsg:
		switch msg.Type {
		// quit
		case tea.KeyCtrlC:
			return u, tea.Quit
		// history
		case tea.KeyUp, tea.KeyDown:
			if !u.state.querying && !u.state.confirming {
				var input *string
				if msg.Type == tea.KeyUp {
					input = u.history.GetPrevious()
				} else {
					input = u.history.GetNext()
				}
				if input != nil {
					u.components.prompt.SetValue(*input)
					u.components.prompt, promptCmd = u.components.prompt.Update(msg)
					cmds = append(
						cmds,
						promptCmd,
					)
				}
			}
		// switch mode
		case tea.KeyTab:
			if !u.state.querying && !u.state.confirming {
				var modeChangeMessage string

				if u.state.promptMode == ChatPromptMode {
					u.state.promptMode = ExecPromptMode
					u.components.prompt.SetMode(ExecPromptMode)
					u.engine.SetMode(ai.ExecEngineMode)
					modeChangeMessage = u.components.renderer.RenderSuccess("\n[Switched to command mode with context preservation]\n")
				} else {
					u.state.promptMode = ChatPromptMode
					u.components.prompt.SetMode(ChatPromptMode)
					u.engine.SetMode(ai.ChatEngineMode)
					modeChangeMessage = u.components.renderer.RenderSuccess("\n[Switched to chat mode with context preservation]\n")
				}

				// Don't call engine.Reset() to preserve context between modes
				u.components.prompt, promptCmd = u.components.prompt.Update(msg)

				// Add the mode switch information to terminal outputs for better context
				var oldMode, newMode string
				if u.state.promptMode == ChatPromptMode {
					oldMode = "command"
					newMode = "chat"
				} else {
					oldMode = "chat"
					newMode = "command"
				}
				u.engine.AddTerminalOutput(fmt.Sprintf("Switched from %s mode to %s mode. Context from previous conversation was preserved.", oldMode, newMode))

				cmds = append(
					cmds,
					promptCmd,
					tea.Println(modeChangeMessage),
					textinput.Blink,
				)
			}
		// enter
		case tea.KeyEnter:
			return u.handleEnterKey(msg)

		// help
		case tea.KeyCtrlH:
			if !u.state.configuring && !u.state.querying && !u.state.confirming {
				u.components.prompt, promptCmd = u.components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					tea.Println(u.components.renderer.RenderContent(u.components.renderer.RenderHelpMessage())),
					textinput.Blink,
				)
			}

		// clear
		case tea.KeyCtrlL:
			if !u.state.querying && !u.state.confirming {
				u.components.prompt, promptCmd = u.components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					tea.ClearScreen,
					textinput.Blink,
				)
			}

		// reset
		case tea.KeyCtrlR:
			if !u.state.querying && !u.state.confirming {
				u.history.Reset()
				u.engine.Reset()
				u.components.prompt.SetValue("")
				u.components.prompt, promptCmd = u.components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					tea.ClearScreen,
					textinput.Blink,
				)
			}

		// edit settings
		case tea.KeyCtrlS:
			if !u.state.querying && !u.state.confirming && !u.state.configuring && !u.state.executing {
				u.state.executing = true
				u.state.buffer = ""
				u.state.command = ""
				u.components.prompt.Blur()
				u.components.prompt, promptCmd = u.components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					u.editSettings(),
				)
			}

		default:
			if u.state.confirming {
				if strings.ToLower(msg.String()) == "y" {
					u.state.confirming = false
					u.state.executing = true
					u.state.buffer = ""
					u.components.prompt.SetValue("")
					return u, tea.Sequence(
						promptCmd,
						u.execCommand(u.state.command),
					)
				} else {
					u.state.confirming = false
					u.state.executing = false
					u.state.buffer = ""
					u.components.prompt, promptCmd = u.components.prompt.Update(msg)
					u.components.prompt.SetValue("")
					u.components.prompt.Focus()
					if u.state.runMode == ReplMode {
						cmds = append(
							cmds,
							promptCmd,
							tea.Println(fmt.Sprintf("\n%s\n", u.components.renderer.RenderWarning("[cancel]"))),
							textinput.Blink,
						)
					} else {
						return u, tea.Sequence(
							promptCmd,
							tea.Println(fmt.Sprintf("\n%s\n", u.components.renderer.RenderWarning("[cancel]"))),
							tea.Quit,
						)
					}
				}
				u.state.command = ""
			} else {
				u.components.prompt.Focus()
				u.components.prompt, promptCmd = u.components.prompt.Update(msg)
				cmds = append(
					cmds,
					promptCmd,
					textinput.Blink,
				)
			}
		}
	// engine exec feedback
	case ai.EngineExecOutput:
		var output string
		if msg.IsExecutable() {
			// Check for information queries that should run automatically
			if strings.Contains(strings.ToLower(u.state.args), "what") ||
				strings.Contains(strings.ToLower(u.state.args), "how") ||
				strings.Contains(strings.ToLower(u.state.args), "show") ||
				strings.Contains(strings.ToLower(u.state.args), "display") ||
				strings.Contains(strings.ToLower(u.state.args), "print") {
				// Auto-execute basic info commands
				u.state.confirming = false
				u.state.executing = true
				output = u.components.renderer.RenderContent(fmt.Sprintf("Executing: `%s`", msg.GetCommand()))
				// Save output to engine context
				u.engine.AddTerminalOutput(output)
				u.components.prompt, promptCmd = u.components.prompt.Update(msg)
				return u, tea.Sequence(
					promptCmd,
					tea.Println(output),
					u.execCommand(msg.GetCommand()),
				)
			}

			// Regular confirmation flow
			u.state.confirming = true
			u.state.command = msg.GetCommand()
			output = u.components.renderer.RenderContent(fmt.Sprintf("`%s`", u.state.command))
			output += fmt.Sprintf("  %s\n\n  confirm execution? [y/N]", u.components.renderer.RenderHelp(msg.GetExplanation()))
			u.components.prompt.Blur()
		} else {
			output = u.components.renderer.RenderContent(msg.GetExplanation())
			u.components.prompt.Focus()
			if u.state.runMode == CliMode {
				// Save output to engine context before quitting
				u.engine.AddTerminalOutput(output)
				return u, tea.Sequence(
					tea.Println(output),
					tea.Quit,
				)
			}
		}
		// Save output to engine context
		u.engine.AddTerminalOutput(output)
		u.components.prompt, promptCmd = u.components.prompt.Update(msg)
		return u, tea.Sequence(
			promptCmd,
			textinput.Blink,
			tea.Println(output),
		)
	// engine chat stream feedback
	case ai.EngineChatStreamOutput:
		if msg.IsLast() {
			output := u.components.renderer.RenderContent(u.state.buffer)
			u.state.buffer = ""
			u.components.prompt.Focus()
			if u.state.runMode == CliMode {
				return u, tea.Sequence(
					tea.Println(output),
					tea.Quit,
				)
			} else {
				return u, tea.Sequence(
					tea.Println(output),
					textinput.Blink,
				)
			}
		} else {
			return u, u.awaitChatStream()
		}
	// runner feedback
	case run.RunOutput:
		u.state.querying = false
		u.components.prompt, promptCmd = u.components.prompt.Update(msg)
		u.components.prompt.Focus()
		output := u.components.renderer.RenderSuccess(fmt.Sprintf("\n%s\n", msg.GetSuccessMessage()))
		if msg.HasError() {
			output = u.components.renderer.RenderError(fmt.Sprintf("\n%s\n", msg.GetErrorMessage()))
		}
		if u.state.runMode == CliMode {
			return u, tea.Sequence(
				tea.Println(output),
				tea.Quit,
			)
		} else {
			return u, tea.Sequence(
				tea.Println(output),
				promptCmd,
				textinput.Blink,
			)
		}
	// errors
	case error:
		u.state.error = msg
		return u, nil
	}

	return u, tea.Batch(cmds...)
}

func (u *Ui) View() string {
	if u.state.error != nil {
		return u.components.renderer.RenderError(fmt.Sprintf("[error] %s", u.state.error))
	}

	if u.state.configuring {
		return fmt.Sprintf(
			"%s\n%s",
			u.components.renderer.RenderContent(u.state.buffer),
			u.components.prompt.View(),
		)
	}

	if !u.state.querying && !u.state.confirming && !u.state.executing {
		// If we have active autocomplete, show suggestions
		if u.components.prompt.HasActiveAutocomplete() {
			return fmt.Sprintf(
				"%s\n\n%s",
				u.components.prompt.View(),
				u.components.prompt.GetAutocompleteSuggestions(),
			)
		}
		return u.components.prompt.View()
	}

	if u.state.promptMode == ChatPromptMode {
		return u.components.renderer.RenderContent(u.state.buffer)
	} else {
		if u.state.querying {
			return u.components.spinner.View()
		} else {
			if !u.state.executing {
				return u.components.renderer.RenderContent(u.state.buffer)
			}
		}
	}

	return ""
}

func (u *Ui) startRepl(config *config.Config) tea.Cmd {
	return tea.Sequence(
		tea.ClearScreen,
		tea.Println(u.components.renderer.RenderContent(u.components.renderer.RenderHelpMessage())),
		textinput.Blink,
		func() tea.Msg {
			u.config = config

			// Use chat mode as default if no preference in config
			if u.state.promptMode == DefaultPromptMode {
				configMode := config.GetUserConfig().GetDefaultPromptMode()
				if configMode != "" {
					u.state.promptMode = GetPromptModeFromString(configMode)
				} else {
					u.state.promptMode = ChatPromptMode
				}
			}

			engineMode := ai.ExecEngineMode
			if u.state.promptMode == ChatPromptMode {
				engineMode = ai.ChatEngineMode
			}

			engine, err := ai.NewEngine(engineMode, config)
			if err != nil {
				return err
			}

			if u.state.pipe != "" {
				engine.SetPipe(u.state.pipe)
			}

			u.engine = engine
			u.state.buffer = "Welcome \n\n"
			u.state.command = ""
			u.components.prompt = NewPrompt(u.state.promptMode)

			return nil
		},
	)
}

func (u *Ui) startCli(config *config.Config) tea.Cmd {
	u.config = config

	// Use chat mode as default if no preference in config
	if u.state.promptMode == DefaultPromptMode {
		configMode := config.GetUserConfig().GetDefaultPromptMode()
		if configMode != "" {
			u.state.promptMode = GetPromptModeFromString(configMode)
		} else {
			u.state.promptMode = ChatPromptMode
		}
	}

	engineMode := ai.ExecEngineMode
	if u.state.promptMode == ChatPromptMode {
		engineMode = ai.ChatEngineMode
	}

	engine, err := ai.NewEngine(engineMode, config)
	if err != nil {
		u.state.error = err
		return nil
	}

	if u.state.pipe != "" {
		engine.SetPipe(u.state.pipe)
	}

	u.engine = engine
	u.state.querying = true
	u.state.confirming = false
	u.state.buffer = ""
	u.state.command = ""

	if u.state.promptMode == ExecPromptMode {
		return tea.Batch(
			u.components.spinner.Tick,
			func() tea.Msg {
				output, err := u.engine.ExecCompletion(u.state.args)
				u.state.querying = false
				if err != nil {
					return err
				}

				return *output
			},
		)
	} else {
		return tea.Batch(
			u.startChatStream(u.state.args),
			u.awaitChatStream(),
		)
	}
}

func (u *Ui) startConfig() tea.Cmd {
	return func() tea.Msg {
		u.state.configuring = true
		u.state.querying = false
		u.state.confirming = false
		u.state.executing = false

		u.state.buffer = u.components.renderer.RenderConfigMessage()
		u.state.command = ""
		u.components.prompt = NewPrompt(ProviderPromptMode)

		return nil
	}
}

func (u *Ui) startModelConfig(providerType provider.ProviderType) tea.Cmd {
	return func() tea.Msg {
		u.state.providerType = providerType
		u.state.buffer = u.components.renderer.RenderModelMessage(string(providerType))
		u.components.prompt = NewPrompt(ModelPromptMode)

		return nil
	}
}

func (u *Ui) startApiKeyConfig(model string) tea.Cmd {
	return func() tea.Msg {
		if model != "" {
			u.state.modelName = model
		} else {
			// Use default model if none specified
			u.state.modelName = config.GetDefaultModelForProvider(u.state.providerType)
		}

		u.state.buffer = u.components.renderer.RenderApiKeyMessage()
		u.components.prompt = NewPrompt(ConfigPromptMode)

		return nil
	}
}

func (u *Ui) finishConfig(input string) tea.Cmd {
	// Step 1: Provider selection
	if u.components.prompt.GetMode() == ProviderPromptMode {
		var providerType provider.ProviderType

		// Handle empty input (default to OpenAI)
		if input == "" {
			input = "1"
		}

		switch input {
		case "1", "openai":
			providerType = provider.ProviderOpenAI
		case "2", "gemini":
			providerType = provider.ProviderGemini
		case "3", "claude":
			providerType = provider.ProviderClaude
		default:
			// Default to OpenAI if input is invalid
			providerType = provider.ProviderOpenAI
		}

		return u.startModelConfig(providerType)
	}

	// Step 2: Model selection
	if u.components.prompt.GetMode() == ModelPromptMode {
		var modelName string

		// Handle empty input (use default)
		if input == "" {
			input = "1" // Default to first option
		}

		// Select model based on provider and input
		switch string(u.state.providerType) {
		case "openai":
			switch input {
			case "1":
				modelName = "gpt-3.5-turbo"
			case "2":
				modelName = "gpt-4"
			case "3":
				modelName = "gpt-4-turbo"
			default:
				modelName = "gpt-3.5-turbo" // Default
			}
		case "gemini":
			switch input {
			case "1":
				modelName = "gemini-2.0-flash"
			case "2":
				modelName = "gemini-2.5-pro-exp-03-25"
			case "3":
				modelName = "gemini-2.0-pro"
			case "4":
				modelName = "gemini-2.0-flash-lite"
			case "5":
				modelName = "gemini-1.5-pro"
			case "6":
				modelName = "gemini-1.5-flash"
			case "7":
				modelName = "gemini-1.5-flash-8b"
			default:
				modelName = "gemini-2.0-flash" // Default
			}
		case "claude":
			switch input {
			case "1":
				modelName = "claude-3-haiku-20240307"
			case "2":
				modelName = "claude-3-sonnet-20240229"
			case "3":
				modelName = "claude-3-opus-20240229"
			default:
				modelName = "claude-3-haiku-20240307" // Default
			}
		default:
			// Fallback to default for selected provider
			modelName = config.GetDefaultModelForProvider(u.state.providerType)
		}

		return u.startApiKeyConfig(modelName)
	}

	// Step 3: API key input
	u.state.configuring = false

	// API Key validation - don't allow empty key
	if input == "" {
		u.state.error = fmt.Errorf("API key cannot be empty. Please provide a valid API key.")
		// Go back to the API key input
		return u.startApiKeyConfig(u.state.modelName)
	}

	// Model already selected in previous step
	model := u.state.modelName
	if model == "" {
		// Safety check - use default if somehow we got here with no model
		model = config.GetDefaultModelForProvider(u.state.providerType)
	}

	config, err := config.WriteConfig(u.state.providerType, input, model, true)
	if err != nil {
		u.state.error = err
		return nil
	}

	u.config = config
	engine, err := ai.NewEngine(ai.ExecEngineMode, config)
	if err != nil {
		u.state.error = err
		return nil
	}

	if u.state.pipe != "" {
		engine.SetPipe(u.state.pipe)
	}

	u.engine = engine

	if u.state.runMode == ReplMode {
		return tea.Sequence(
			tea.ClearScreen,
			tea.Println(u.components.renderer.RenderSuccess("\n[settings ok]\n")),
			textinput.Blink,
			func() tea.Msg {
				u.state.buffer = ""
				u.state.command = ""
				u.components.prompt = NewPrompt(ExecPromptMode)

				return nil
			},
		)
	} else {
		if u.state.promptMode == ExecPromptMode {
			u.state.querying = true
			u.state.configuring = false
			u.state.buffer = ""
			return tea.Sequence(
				tea.Println(u.components.renderer.RenderSuccess("\n[settings ok]")),
				u.components.spinner.Tick,
				func() tea.Msg {
					output, err := u.engine.ExecCompletion(u.state.args)
					u.state.querying = false
					if err != nil {
						return err
					}

					return *output
				},
			)
		} else {
			return tea.Batch(
				u.startChatStream(u.state.args),
				u.awaitChatStream(),
			)
		}
	}
}

func (u *Ui) startExec(input string) tea.Cmd {
	return func() tea.Msg {
		u.state.querying = true
		u.state.confirming = false
		u.state.buffer = ""
		u.state.command = ""

		output, err := u.engine.ExecCompletion(input)
		u.state.querying = false
		if err != nil {
			return err
		}

		return *output
	}
}

func (u *Ui) startChatStream(input string) tea.Cmd {
	return func() tea.Msg {
		u.state.querying = true
		u.state.executing = false
		u.state.confirming = false
		u.state.buffer = ""
		u.state.command = ""

		err := u.engine.ChatStreamCompletion(input)
		if err != nil {
			return err
		}

		return nil
	}
}

func (u *Ui) awaitChatStream() tea.Cmd {
	return func() tea.Msg {
		output := <-u.engine.GetChannel()
		u.state.buffer += output.GetContent()
		u.state.querying = !output.IsLast()

		return output
	}
}

func (u *Ui) execCommand(input string) tea.Cmd {
	u.state.querying = false
	u.state.confirming = false
	u.state.executing = true

	c := run.PrepareInteractiveCommand(input)

	return tea.ExecProcess(c, func(error error) tea.Msg {
		u.state.executing = false
		u.state.command = ""

		output := "[ok]"
		if error != nil {
			output = "[error]"
		}

		// Capture command execution result to engine context
		result := run.NewRunOutput(error, "[error]", "[ok]")
		u.engine.AddTerminalOutput(fmt.Sprintf("$ %s\n%s", input, output))

		return result
	})
}

func (u *Ui) editSettings() tea.Cmd {
	u.state.querying = false
	u.state.confirming = false
	u.state.executing = true

	c := run.PrepareEditSettingsCommand(fmt.Sprintf(
		"%s %s",
		u.config.GetSystemConfig().GetEditor(),
		u.config.GetSystemConfig().GetConfigFile(),
	))

	return tea.ExecProcess(c, func(error error) tea.Msg {
		u.state.executing = false
		u.state.command = ""

		if error != nil {
			return run.NewRunOutput(error, "[settings error]", "")
		}

		config, error := config.NewConfig()
		if error != nil {
			return run.NewRunOutput(error, "[settings error]", "")
		}

		u.config = config
		engine, error := ai.NewEngine(ai.ExecEngineMode, config)
		if u.state.pipe != "" {
			engine.SetPipe(u.state.pipe)
		}
		if error != nil {
			return run.NewRunOutput(error, "[settings error]", "")
		}
		u.engine = engine

		return run.NewRunOutput(nil, "", "[settings ok]")
	})
}

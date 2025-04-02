package ui

// This file contains the implementation for handling Enter key in the UI

import (
	"fmt"

	"github.com/xsikor/yai/ai"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// handleEnterKey handles the Enter key press in the UI
func (u *Ui) handleEnterKey(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds      []tea.Cmd
		promptCmd tea.Cmd
	)

	if u.state.configuring {
		return u, u.finishConfig(u.components.prompt.GetValue())
	}

	if !u.state.querying && !u.state.confirming {
		input := u.components.prompt.GetValue()
		if input != "" {
			// Check if this is a slash command
			if u.components.prompt.IsSlashCommand() {
				cmdOutput := u.components.prompt.ExecuteSlashCommand(u.config)
				inputPrint := u.components.prompt.AsString()
				u.history.Add(input)
				u.components.prompt.SetValue("")
				u.components.prompt, promptCmd = u.components.prompt.Update(msg)

				// Handle special commands
				if cmdOutput == "[clear]" {
					return u, tea.Sequence(
						promptCmd,
						tea.ClearScreen,
						textinput.Blink,
					)
				} else if cmdOutput == "[reset]" {
					u.engine.Reset()
					u.history.Reset()
					return u, tea.Sequence(
						promptCmd,
						tea.Println(u.components.renderer.RenderSuccess("\n[History cleared]\n")),
						textinput.Blink,
					)
				} else if cmdOutput == "[mode]" {
					// Toggle mode with context preservation
					var modeChangeMessage string

					if u.state.promptMode == ChatPromptMode {
						u.state.promptMode = ExecPromptMode
						u.components.prompt.SetMode(ExecPromptMode)
						u.engine.SetMode(ai.ExecEngineMode)
						modeChangeMessage = fmt.Sprintf("\n[Switched to %s mode with context preservation]\n", u.state.promptMode.String())
						// Add the mode switch information to terminal outputs for better context
						u.engine.AddTerminalOutput("Switched from chat mode to command mode. Context from previous conversation was preserved.")

						modeChangeMessage = fmt.Sprintf("\n[Switched to %s mode with context preservation]\n", u.state.promptMode.String())
						u.engine.SetMode(ai.ChatEngineMode)
						modeChangeMessage = fmt.Sprintf("\n[Switched to %s mode with context preservation]\n", u.state.promptMode.String())
						// Add the mode switch information to terminal outputs for better context
						u.engine.AddTerminalOutput("Switched from command mode to chat mode. Context from previous conversation was preserved.")

						u.state.promptMode = ChatPromptMode
						u.components.prompt.SetMode(ChatPromptMode)
						u.engine.SetMode(ai.ChatEngineMode)
						modeChangeMessage = fmt.Sprintf("\n[Switched to %s mode with context preservation]\n", u.state.promptMode.String())
					}
					return u, tea.Sequence(
						promptCmd,
						tea.Println(u.components.renderer.RenderSuccess(modeChangeMessage)),
						textinput.Blink,
					)
				}

				// Regular command output
				return u, tea.Sequence(
					promptCmd,
					tea.Println(inputPrint),
					tea.Println(u.components.renderer.RenderContent(cmdOutput)),
					textinput.Blink,
				)
			}

			// Regular input handling
			inputPrint := u.components.prompt.AsString()
			u.history.Add(input)
			u.components.prompt.SetValue("")
			u.components.prompt.Blur()
			u.components.prompt, promptCmd = u.components.prompt.Update(msg)

			// Store the input as args for auto-execution detection
			u.state.args = input
			if u.state.promptMode == ChatPromptMode {
				cmds = append(
					cmds,
					promptCmd,
					tea.Println(inputPrint),
					u.startChatStream(input),
					u.awaitChatStream(),
				)
			} else {
				cmds = append(
					cmds,
					promptCmd,
					tea.Println(inputPrint),
					u.startExec(input),
					u.components.spinner.Tick,
				)
			}
		}
	}

	return u, tea.Batch(cmds...)
}

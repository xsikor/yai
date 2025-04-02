package ui

// This file contains the implementation for handling Enter key in the UI

import (
	"fmt"
	
	"github.com/ekkinox/yai/ai"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
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
					// Toggle mode
					if u.state.promptMode == ChatPromptMode {
						u.state.promptMode = ExecPromptMode
						u.components.prompt.SetMode(ExecPromptMode)
						u.engine.SetMode(ai.ExecEngineMode)
					} else {
						u.state.promptMode = ChatPromptMode
						u.components.prompt.SetMode(ChatPromptMode)
						u.engine.SetMode(ai.ChatEngineMode)
					}
					return u, tea.Sequence(
						promptCmd,
						tea.Println(u.components.renderer.RenderSuccess(fmt.Sprintf("\n[Switched to %s mode]\n", u.state.promptMode.String()))),
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
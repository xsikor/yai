package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ekkinox/yai/ai/provider"
	"github.com/ekkinox/yai/config"
	"github.com/ekkinox/yai/system"
)

const noexec = "[noexec]"

type Engine struct {
	mode              EngineMode
	config            *config.Config
	provider          provider.Provider
	execMessages      []provider.Message
	chatMessages      []provider.Message
	sharedHistory     []provider.Message // Shared context between modes
	terminalOutputs   []string           // History of terminal outputs for context
	maxSharedHistory  int                // Maximum number of messages to keep in shared history
	maxTerminalOutput int                // Maximum number of terminal outputs to keep
	channel           chan EngineChatStreamOutput
	pipe              string
	running           bool
}

func NewEngine(mode EngineMode, config *config.Config) (*Engine, error) {
	providerInstance, err := provider.CreateProvider(
		config.GetAiConfig().GetProviderType(),
		config.GetAiConfig().GetKey(),
		config.GetAiConfig().GetProxy(),
	)
	if err != nil {
		return nil, err
	}

	return &Engine{
		mode:              mode,
		config:            config,
		provider:          providerInstance,
		execMessages:      make([]provider.Message, 0),
		chatMessages:      make([]provider.Message, 0),
		sharedHistory:     make([]provider.Message, 0),
		terminalOutputs:   make([]string, 0),
		maxSharedHistory:  5, // Store the last 5 messages for context
		maxTerminalOutput: 5, // Store the last 5 terminal outputs
		channel:           make(chan EngineChatStreamOutput),
		pipe:              "",
		running:           false,
	}, nil
}

func (e *Engine) SetMode(mode EngineMode) *Engine {
	// If mode is changing, save current context before switching
	if e.mode != mode {
		e.updateSharedHistory()
	}
	
	e.mode = mode

	return e
}

// updateSharedHistory saves recent messages from current mode to shared history
func (e *Engine) updateSharedHistory() {
	var currentMessages []provider.Message
	
	// Get messages from current mode
	if e.mode == ExecEngineMode {
		currentMessages = e.execMessages
	} else {
		currentMessages = e.chatMessages
	}
	
	// Only process if we have messages
	if len(currentMessages) > 0 {
		// Get the latest messages (up to maxSharedHistory)
		start := 0
		if len(currentMessages) > e.maxSharedHistory {
			start = len(currentMessages) - e.maxSharedHistory
		}
		
		// Update shared history with the latest messages
		e.sharedHistory = make([]provider.Message, len(currentMessages[start:]))
		copy(e.sharedHistory, currentMessages[start:])
	}
}

func (e *Engine) GetMode() EngineMode {
	return e.mode
}

func (e *Engine) GetChannel() chan EngineChatStreamOutput {
	return e.channel
}

// AddTerminalOutput adds a terminal output to history
func (e *Engine) AddTerminalOutput(output string) *Engine {
	// Skip empty outputs
	if strings.TrimSpace(output) == "" {
		return e
	}
	
	// Add new output to the terminal outputs
	e.terminalOutputs = append(e.terminalOutputs, output)
	
	// Trim history if it exceeds the maximum
	if len(e.terminalOutputs) > e.maxTerminalOutput {
		e.terminalOutputs = e.terminalOutputs[len(e.terminalOutputs)-e.maxTerminalOutput:]
	}
	
	return e
}

// GetTerminalOutputs returns all terminal outputs
func (e *Engine) GetTerminalOutputs() []string {
	return e.terminalOutputs
}

func (e *Engine) SetPipe(pipe string) *Engine {
	e.pipe = pipe
	
	// Auto-detect mode based on piped content if no mode was explicitly set
	if pipe != "" {
		// Use isProbablyCommand from ui package to detect if the pipe content is a command
		if e.detectPipeContentMode(pipe) {
			e.mode = ExecEngineMode
		} else {
			e.mode = ChatEngineMode
		}
	}

	return e
}

// detectPipeContentMode determines if piped content is likely a command or chat
// Returns true if content is likely a command, false otherwise
func (e *Engine) detectPipeContentMode(input string) bool {
	// If empty, can't tell
	if len(input) == 0 {
		return false
	}
	
	// Common command patterns
	commandRegexps := []string{
		`^(ls|cd|grep|find|git|docker|kubectl|npm|go|python|pip)\s`,
		`^[./]`, // Starts with ./ or /
		`\|\s*(\w+)`, // Contains pipe operator
		`sudo\s`, // Contains sudo
		`^(cat|less|more|head|tail|vim|nano|mkdir|rmdir|touch|chmod|chown)\s`,
		`^(apt|yum|brew)\s`,
		`\s(>|>>|<)\s`, // Contains redirection operators
	}
	
	// Check if input matches any command patterns
	for _, pattern := range commandRegexps {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}
	
	// Special case: Command-like queries
	commandQueryPatterns := []string{
		`(?i)^(what|how|show|get|find|list|display).*(command|run|execute)`,
		`(?i)^(what|how).*(ip|address|port|url|endpoint)`,
		`(?i)^(what|how).*(container|pod|instance|server|service)`,
		`(?i)^(how to|how do I) `,
		`(?i)command for`,
		`(?i)command to `,
	}
	
	// Check for command-like queries
	for _, pattern := range commandQueryPatterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}
	
	// If text is very short (1-3 words) and doesn't contain question mark, likely a command
	words := strings.Fields(input)
	if len(words) <= 3 && !strings.Contains(input, "?") {
		return true
	}
	
	// Default to false if no command patterns matched
	return false
}

func (e *Engine) Interrupt() *Engine {
	e.channel <- EngineChatStreamOutput{
		content:    "[Interrupt]",
		last:       true,
		interrupt:  true,
		executable: false,
	}

	e.running = false

	return e
}

func (e *Engine) Clear() *Engine {
	if e.mode == ExecEngineMode {
		e.execMessages = []provider.Message{}
	} else {
		e.chatMessages = []provider.Message{}
	}

	return e
}

func (e *Engine) Reset() *Engine {
	// Save current context before reset
	e.updateSharedHistory()
	
	// Clear both message histories
	e.execMessages = []provider.Message{}
	e.chatMessages = []provider.Message{}

	return e
}

func (e *Engine) ExecCompletion(input string) (*EngineExecOutput, error) {
	ctx := context.Background()

	e.running = true

	e.appendUserMessage(input)

	content, err := e.provider.CreateCompletion(
		ctx,
		provider.CompletionRequest{
			Model:       e.config.GetAiConfig().GetModel(),
			MaxTokens:   e.config.GetAiConfig().GetMaxTokens(),
			Temperature: e.config.GetAiConfig().GetTemperature(),
			Messages:    e.prepareCompletionMessages(),
		},
	)
	if err != nil {
		return nil, err
	}

	e.appendAssistantMessage(content)

	var output EngineExecOutput
	err = json.Unmarshal([]byte(content), &output)
	if err != nil {
		re := regexp.MustCompile(`\{.*?\}`)
		match := re.FindString(content)
		if match != "" {
			err = json.Unmarshal([]byte(match), &output)
			if err != nil {
				return nil, err
			}
		} else {
			output = EngineExecOutput{
				Command:     "",
				Explanation: content,
				Executable:  false,
			}
		}
	}

	return &output, nil
}

func (e *Engine) ChatStreamCompletion(input string) error {
	ctx := context.Background()

	e.running = true

	e.appendUserMessage(input)

	completionReq := provider.CompletionRequest{
		Model:       e.config.GetAiConfig().GetModel(),
		MaxTokens:   e.config.GetAiConfig().GetMaxTokens(),
		Temperature: e.config.GetAiConfig().GetTemperature(),
		Messages:    e.prepareCompletionMessages(),
		Stream:      true,
	}

	stream, err := e.provider.CreateCompletionStream(ctx, completionReq)
	if err != nil {
		return err
	}

	var output string

	for resp := range stream {
		if !e.running {
			break
		}

		output += resp.Content

		if resp.Done {
			executable := false
			if e.mode == ExecEngineMode {
				if !strings.HasPrefix(output, noexec) && !strings.Contains(output, "\n") {
					executable = true
				}
			}

			e.channel <- EngineChatStreamOutput{
				content:    "",
				last:       true,
				executable: executable,
			}
			e.running = false
			e.appendAssistantMessage(output)

			return nil
		}

		e.channel <- EngineChatStreamOutput{
			content: resp.Content,
			last:    false,
		}
	}

	// In case the stream closes without a done flag
	executable := false
	if e.mode == ExecEngineMode {
		if !strings.HasPrefix(output, noexec) && !strings.Contains(output, "\n") {
			executable = true
		}
	}

	// Always send a final message to signal completion
	e.channel <- EngineChatStreamOutput{
		content:    "",
		last:       true,
		executable: executable,
	}
	
	e.running = false
	e.appendAssistantMessage(output)

	return nil
}

func (e *Engine) appendUserMessage(content string) *Engine {
	msg := provider.Message{
		Role:    "user",
		Content: content,
	}

	if e.mode == ExecEngineMode {
		e.execMessages = append(e.execMessages, msg)
	} else {
		e.chatMessages = append(e.chatMessages, msg)
	}

	return e
}

func (e *Engine) appendAssistantMessage(content string) *Engine {
	msg := provider.Message{
		Role:    "assistant",
		Content: content,
	}

	if e.mode == ExecEngineMode {
		e.execMessages = append(e.execMessages, msg)
	} else {
		e.chatMessages = append(e.chatMessages, msg)
	}

	return e
}

func (e *Engine) prepareCompletionMessages() []provider.Message {
	messages := []provider.Message{
		{
			Role:    "system",
			Content: e.prepareSystemPrompt(),
		},
	}

	if e.pipe != "" {
		messages = append(
			messages,
			provider.Message{
				Role:    "user",
				Content: e.preparePipePrompt(),
			},
		)
		}
		
		// Add terminal outputs as context if available
		if len(e.terminalOutputs) > 0 {
			var terminalContext strings.Builder
			
			terminalContext.WriteString("Recent terminal outputs for context:\n\n")
			for i, output := range e.terminalOutputs {
				terminalContext.WriteString(fmt.Sprintf("Terminal output %d:\n```\n%s\n```\n\n", i+1, output))
			}
			
			messages = append(
				messages,
				provider.Message{
					Role:    "system",
					Content: terminalContext.String(),
				},
			)
		}
		
		// Add shared history context if available and we're in a new mode with no messages yet
	currentModeMessages := e.execMessages
	if e.mode == ChatEngineMode {
		currentModeMessages = e.chatMessages
	}
	
	if len(currentModeMessages) == 0 && len(e.sharedHistory) > 0 {
		// If we have no messages in the current mode but have shared history,
		// add a context message explaining we're continuing with context from the other mode
		contextModeStr := "chat"
		if e.mode == ChatEngineMode {
			contextModeStr = "command"
		}
		
		// Add context reminder
		messages = append(
			messages,
			provider.Message{
				Role:    "system",
				Content: fmt.Sprintf("Here is recent context from %s mode that might be relevant:", contextModeStr),
			},
		)
		
		// Add shared history
		messages = append(messages, e.sharedHistory...)
		
		// Add separator
		messages = append(
			messages,
			provider.Message{
				Role:    "system",
				Content: fmt.Sprintf("Now continuing in %s mode:", e.mode.String()),
			},
		)
	}

	// Add current mode messages
	if e.mode == ExecEngineMode {
		messages = append(messages, e.execMessages...)
	} else {
		messages = append(messages, e.chatMessages...)
	}

	return messages
}

func (e *Engine) preparePipePrompt() string {
	return fmt.Sprintf("I will work on the following input: %s", e.pipe)
}

func (e *Engine) prepareSystemPrompt() string {
	var bodyPart string
	if e.mode == ExecEngineMode {
		bodyPart = e.prepareSystemPromptExecPart()
	} else {
		bodyPart = e.prepareSystemPromptChatPart()
	}

	return fmt.Sprintf("%s\n%s", bodyPart, e.prepareSystemPromptContextPart())
}

func (e *Engine) prepareSystemPromptExecPart() string {
	return "Your are Yai, a powerful terminal assistant generating a JSON containing a command line for my input.\n" +
		"You will always reply using the following json structure: {\"cmd\":\"the command\", \"exp\": \"some explanation\", \"exec\": true}.\n" +
		"Your answer will always only contain the json structure, never add any advice or supplementary detail or information, even if I asked the same question before.\n" +
		"The field cmd will contain a single line command (don't use new lines, use separators like && and ; instead).\n" +
		"The field exp will contain an short explanation of the command if you managed to generate an executable command, otherwise it will contain the reason of your failure.\n" +
		"The field exec will contain true if you managed to generate an executable command, false otherwise." +
		"\n" +
		"Examples:\n" +
		"Me: list all files in my home dir\n" +
		"Yai: {\"cmd\":\"ls ~\", \"exp\": \"list all files in your home dir\", \"exec\\: true}\n" +
		"Me: list all pods of all namespaces\n" +
		"Yai: {\"cmd\":\"kubectl get pods --all-namespaces\", \"exp\": \"list pods form all k8s namespaces\", \"exec\": true}\n" +
		"Me: how are you ?\n" +
		"Yai: {\"cmd\":\"\", \"exp\": \"I'm good thanks but I cannot generate a command for this. Use the chat mode to discuss.\", \"exec\": false}"
}

func (e *Engine) prepareSystemPromptChatPart() string {
	return "You are Yai a powerful terminal assistant created by github.com/ekkinox.\n" +
		"You will answer in the most helpful possible way.\n" +
		"Always format your answer in markdown format.\n\n" +
		"For example:\n" +
		"Me: What is 2+2 ?\n" +
		"Yai: The answer for `2+2` is `4`\n" +
		"Me: +2 again ?\n" +
		"Yai: The answer is `6`\n"
}

func (e *Engine) prepareSystemPromptContextPart() string {
	part := "My context: "

	if e.config.GetSystemConfig().GetOperatingSystem() != system.UnknownOperatingSystem {
		part += fmt.Sprintf("my operating system is %s, ", e.config.GetSystemConfig().GetOperatingSystem().String())
	}
	if e.config.GetSystemConfig().GetDistribution() != "" {
		part += fmt.Sprintf("my distribution is %s, ", e.config.GetSystemConfig().GetDistribution())
	}
	if e.config.GetSystemConfig().GetHomeDirectory() != "" {
		part += fmt.Sprintf("my home directory is %s, ", e.config.GetSystemConfig().GetHomeDirectory())
	}
	if e.config.GetSystemConfig().GetShell() != "" {
		part += fmt.Sprintf("my shell is %s, ", e.config.GetSystemConfig().GetShell())
	}
	if e.config.GetSystemConfig().GetShell() != "" {
		part += fmt.Sprintf("my editor is %s, ", e.config.GetSystemConfig().GetEditor())
	}
	part += "take this into account. "

	if e.config.GetUserConfig().GetPreferences() != "" {
		part += fmt.Sprintf("Also, %s.", e.config.GetUserConfig().GetPreferences())
	}

	return part
}

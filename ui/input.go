package ui

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/xsikor/yai/ai/provider"
)

type UiInput struct {
	runMode      RunMode
	promptMode   PromptMode
	providerType provider.ProviderType
	modelName    string
	showModel    bool
	args         string
	pipe         string
}

func NewUIInput() (*UiInput, error) {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	var exec, chat, showModel bool
	var providerFlag, modelFlag string
	flagSet.BoolVar(&exec, "e", false, "exec prompt mode")
	flagSet.BoolVar(&chat, "c", false, "chat prompt mode")
	flagSet.BoolVar(&showModel, "m", false, "show current AI model and provider")
	flagSet.StringVar(&providerFlag, "p", "", "AI provider (openai, claude, gemini)")
	flagSet.StringVar(&modelFlag, "model", "", "specific model to use")
	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		fmt.Println("Error parsing flags:", err)
		return nil, err
	}

	args := flagSet.Args()

	stat, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("Error getting stat:", err)
		return nil, err
	}

	pipe := ""
	hasPipe := !(stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0)

	if hasPipe {
		reader := bufio.NewReader(os.Stdin)
		var builder strings.Builder

		for {
			r, _, err := reader.ReadRune()
			if err != nil && err == io.EOF {
				break
			}
			_, err = builder.WriteRune(r)
			if err != nil {
				fmt.Println("Error getting input:", err)
				return nil, err
			}
		}

		pipe = strings.TrimSpace(builder.String())
	}

	runMode := ReplMode
	if len(args) > 0 {
		runMode = CliMode
	} else if hasPipe && pipe != "" {
		runMode = CliMode // If we have piped input, run in CLI mode
	}

	// Setup prompt mode based on flags and/or pipe content
	promptMode := ChatPromptMode // Default to chat mode

	if exec && !chat {
		// Explicit exec mode requested
		promptMode = ExecPromptMode
	} else if chat && !exec {
		// Explicit chat mode requested
		promptMode = ChatPromptMode
	} else if hasPipe && pipe != "" && !exec && !chat {
		// Let the AI engine detect the mode automatically
		// Don't set promptMode here, it will be handled by the engine
		promptMode = DefaultPromptMode
	}

	// Set provider type based on flag
	var providerType provider.ProviderType
	switch strings.ToLower(providerFlag) {
	case "openai":
		providerType = provider.ProviderOpenAI
	case "claude":
		providerType = provider.ProviderClaude
	case "gemini":
		providerType = provider.ProviderGemini
	default:
		// Default to OpenAI if not specified
		providerType = provider.ProviderOpenAI
	}

	return &UiInput{
		runMode:      runMode,
		promptMode:   promptMode,
		providerType: providerType,
		modelName:    modelFlag,
		showModel:    showModel,
		args:         strings.Join(args, " "),
		pipe:         pipe,
	}, nil
}

func (i *UiInput) GetRunMode() RunMode {
	return i.runMode
}

func (i *UiInput) GetPromptMode() PromptMode {
	return i.promptMode
}

func (i *UiInput) GetArgs() string {
	return i.args
}

func (i *UiInput) GetPipe() string {
	return i.pipe
}

func (i *UiInput) GetProviderType() provider.ProviderType {
	return i.providerType
}

func (i *UiInput) GetShowModel() bool {
	return i.showModel
}

func (i *UiInput) GetModelName() string {
	return i.modelName
}

// isProbablyCommand determines if the input text is likely a shell command
// It uses heuristics to detect command patterns
func isProbablyCommand(input string) bool {
	// If empty, can't tell
	if len(input) == 0 {
		return false
	}

	// Common command patterns
	commandPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^(ls|cd|grep|find|git|docker|kubectl|npm|go|python|pip)\s`),
		regexp.MustCompile(`^[./]`),      // Starts with ./ or /
		regexp.MustCompile(`\|\s*(\w+)`), // Contains pipe operator
		regexp.MustCompile(`sudo\s`),     // Contains sudo
		regexp.MustCompile(`^(cat|less|more|head|tail|vim|nano|mkdir|rmdir|touch|chmod|chown)\s`),
		regexp.MustCompile(`^(apt|yum|brew)\s`),
		regexp.MustCompile(`\s(>|>>|<)\s`), // Contains redirection operators
	}

	// Check if input matches any command patterns
	for _, pattern := range commandPatterns {
		if pattern.MatchString(input) {
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

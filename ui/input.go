package ui

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	
	"github.com/ekkinox/yai/ai/provider"
)

type UiInput struct {
	runMode       RunMode
	promptMode    PromptMode
	providerType  provider.ProviderType
	modelName     string
	showModel     bool
	args          string
	pipe          string
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
	if !(stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0) {
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
	}

	promptMode := DefaultPromptMode
	if exec && !chat {
		promptMode = ExecPromptMode
	} else if !exec && chat {
		promptMode = ChatPromptMode
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

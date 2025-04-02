package ui

type PromptMode int

const (
	ExecPromptMode PromptMode = iota
	ConfigPromptMode
	ProviderPromptMode
	ModelPromptMode
	ChatPromptMode
	DefaultPromptMode
)

func (m PromptMode) String() string {
	switch m {
	case ExecPromptMode:
		return "exec"
	case ConfigPromptMode:
		return "config"
	case ProviderPromptMode:
		return "provider"
	case ModelPromptMode:
		return "model"
	case ChatPromptMode:
		return "chat"
	default:
		return "default"
	}
}

func GetPromptModeFromString(s string) PromptMode {
	switch s {
	case "exec":
		return ExecPromptMode
	case "config":
		return ConfigPromptMode
	case "provider":
		return ProviderPromptMode
	case "model":
		return ModelPromptMode
	case "chat":
		return ChatPromptMode
	default:
		return DefaultPromptMode
	}
}

type RunMode int

const (
	CliMode RunMode = iota
	ReplMode
)

func (m RunMode) String() string {
	if m == CliMode {
		return "cli"
	} else {
		return "repl"
	}
}

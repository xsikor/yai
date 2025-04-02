package slash

import (
	"sort"
	"strings"
)

// AutocompleteState represents the current state of autocompletion
type AutocompleteState struct {
	Active      bool
	Suggestions []string
	Index       int
	OriginalInput string
}

// NewAutocompleteState creates a new autocomplete state
func NewAutocompleteState() *AutocompleteState {
	return &AutocompleteState{
		Active:      false,
		Suggestions: []string{},
		Index:       0,
		OriginalInput: "",
	}
}

// StartAutocomplete initiates autocompletion for the given input
func (a *AutocompleteState) StartAutocomplete(input string) bool {
	if !IsSlashCommand(input) {
		a.Reset()
		return false
	}

	// Special case: if it's just a slash, show all commands
	if input == "/" {
		var allCommands []string
		for _, cmd := range SlashCommands {
			allCommands = append(allCommands, "/"+cmd.Name)
		}
		
		// Sort commands alphabetically
		sort.Strings(allCommands)
		
		a.Active = true
		a.Suggestions = allCommands
		a.Index = 0
		a.OriginalInput = input
		return true
	}

	// Get potential completions for partial commands
	suggestions := GetCompletions(input)
	if len(suggestions) == 0 {
		a.Reset()
		return false
	}

	a.Active = true
	a.Suggestions = suggestions
	a.Index = 0
	a.OriginalInput = input

	return true
}

// NextSuggestion cycles to the next suggestion
func (a *AutocompleteState) NextSuggestion() string {
	if !a.Active || len(a.Suggestions) == 0 {
		return ""
	}

	a.Index = (a.Index + 1) % len(a.Suggestions)
	return a.Suggestions[a.Index]
}

// PrevSuggestion cycles to the previous suggestion
func (a *AutocompleteState) PrevSuggestion() string {
	if !a.Active || len(a.Suggestions) == 0 {
		return ""
	}

	a.Index = (a.Index - 1 + len(a.Suggestions)) % len(a.Suggestions)
	return a.Suggestions[a.Index]
}

// GetCurrentSuggestion returns the current suggestion
func (a *AutocompleteState) GetCurrentSuggestion() string {
	if !a.Active || len(a.Suggestions) == 0 {
		return ""
	}

	return a.Suggestions[a.Index]
}

// Reset clears the autocomplete state
func (a *AutocompleteState) Reset() {
	a.Active = false
	a.Suggestions = []string{}
	a.Index = 0
	a.OriginalInput = ""
}

// FormatSuggestions returns a formatted string of suggestions for display
func (a *AutocompleteState) FormatSuggestions() string {
	if !a.Active || len(a.Suggestions) == 0 {
		return ""
	}

	var sb strings.Builder
	
	// Add heading for slash commands
	if a.OriginalInput == "/" {
		sb.WriteString("Available commands:\n")
	}
	
	// Format all suggestions
	for i, suggestion := range a.Suggestions {
		if i == a.Index {
			sb.WriteString("► " + suggestion + " ")
		} else {
			sb.WriteString(suggestion + " ")
		}
	}

	// Add hint for navigation and completion
	sb.WriteString("\n\nUse Up/Down to navigate, Tab or Enter to complete")

	return sb.String()
}

package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/ekkinox/yai/config"
	"github.com/ekkinox/yai/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	input, err := ui.NewUIInput()
	if err != nil {
		log.Fatal(err)
	}

	// Check if we should show model info
	if input.GetShowModel() {
		showModelInfo()
		return
	}

	if _, err := tea.NewProgram(ui.NewUi(input)).Run(); err != nil {
		log.Fatal(err)
	}
}

func showModelInfo() {
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Println("Config not found or invalid.")
		return
	}
	
	fmt.Printf("Current provider: %s\n", cfg.GetAiConfig().GetProviderType())
	fmt.Printf("Current model: %s\n", cfg.GetAiConfig().GetModel())
}

package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fredrikmwold/branch-picker-tui/internal/tui"
)

var _ tea.Program // keep bubbletea required in go.mod

func main() {
	p := tui.NewProgram()
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

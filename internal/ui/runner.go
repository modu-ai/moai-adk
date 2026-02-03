package ui

import tea "github.com/charmbracelet/bubbletea"

// programRunner runs a bubbletea Model and returns the final Model.
// It is a package-level variable so tests can replace it with a mock.
var programRunner = defaultProgramRunner

// defaultProgramRunner creates a real tea.Program and runs it.
func defaultProgramRunner(m tea.Model) (tea.Model, error) {
	p := tea.NewProgram(m)
	return p.Run()
}

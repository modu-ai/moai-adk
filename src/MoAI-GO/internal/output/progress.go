package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Phase names for initialization
var PhaseNames = []string{
	"Phase 1: Preparation and backup...",
	"Phase 2: Creating directory structure...",
	"Phase 3: Installing resources...",
	"Phase 4: Generating configurations...",
	"Phase 5: Validation and finalization...",
}

// progressMsg is sent when progress should be updated
type progressMsg struct {
	phase int
}

// finalMsg is sent when progress is complete
type finalMsg struct{}

// progressModel is the Bubble Tea model for progress display
type progressModel struct {
	progress progress.Model
	quitting bool
	phase    int
	total    int
}

// initialModel creates the initial progress model
func initialModel() progressModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	return progressModel{
		progress: p,
		total:    len(PhaseNames),
	}
}

// Init implements tea.Model
func (m progressModel) Init() tea.Cmd {
	return tickCmd()
}

// Update implements tea.Model
func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	case progressMsg:
		if msg.phase < m.total {
			m.phase = msg.phase
			// Increment progress
			return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
				return progressFrameMsg{}
			})
		} else {
			// All phases complete
			m.quitting = true
			return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
				return finalMsg{}
			})
		}
	case progressFrameMsg:
		newProgress, cmd := m.progress.Update(msg)
		m.progress = newProgress.(progress.Model)
		return m, cmd
	case tickMsg:
		// Auto-advance progress
		if m.phase < m.total {
			return m, func() tea.Msg {
				return progressMsg{phase: m.phase + 1}
			}
		}
	case finalMsg:
		return m, tea.Quit
	}
	return m, nil
}

// View implements tea.Model
func (m progressModel) View() string {
	if m.quitting {
		return ""
	}

	// Build the progress view
	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("\n%s\n\n", InfoStyle.Render("Initializing MoAI-ADK Project...")))

	// Progress bar
	b.WriteString(m.progress.View())

	// Current phase text
	if m.phase > 0 && m.phase <= len(PhaseNames) {
		b.WriteString(fmt.Sprintf("\n\n%s", MutedStyle.Render(PhaseNames[m.phase-1])))
	}

	return b.String()
}

// tickMsg is sent for timing
type tickMsg time.Time

// tickCmd creates a tick command
func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// progressFrameMsg is sent to update the progress frame
type progressFrameMsg struct{}

// ProgressWriter handles progress updates without running a full Bubble Tea program
type ProgressWriter struct {
	phase     int
	total     int
	startTime time.Time
	barWidth  int
	style     lipgloss.Style
}

// NewProgressWriter creates a new progress writer
func NewProgressWriter() *ProgressWriter {
	return &ProgressWriter{
		phase:     0,
		total:     len(PhaseNames),
		startTime: time.Time{},
		barWidth:  40,
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSuccess)).
			Bold(true),
	}
}

// Start starts the progress timer
func (pw *ProgressWriter) Start() {
	pw.startTime = time.Now()
}

// Update updates the progress to the next phase
func (pw *ProgressWriter) Update(message string) {
	if pw.phase < pw.total {
		pw.phase++
		pw.printProgress(message)
		time.Sleep(300 * time.Millisecond) // Simulate work
	}
}

// printProgress prints the current progress bar
func (pw *ProgressWriter) printProgress(message string) {
	// Calculate progress
	percentage := float64(pw.phase) / float64(pw.total)
	filled := int(percentage * float64(pw.barWidth))

	// Build the progress bar
	var bar strings.Builder
	bar.WriteString("[")

	for i := 0; i < pw.barWidth; i++ {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}

	bar.WriteString("]")
	bar.WriteString(fmt.Sprintf(" %d%%", int(percentage*100)))

	// Print progress with message
	fmt.Printf("\r%s %s", pw.style.Render(bar.String()), MutedStyle.Render(message))

	// Add newline when complete
	if pw.phase >= pw.total {
		fmt.Println()
	}
}

// Complete marks the progress as complete
func (pw *ProgressWriter) Complete() {
	pw.phase = pw.total
	duration := time.Since(pw.startTime)
	fmt.Printf("\n%s\n", SuccessStyle.Render("✅ Initialization Completed Successfully!"))
	fmt.Printf("%s\n", MutedStyle.Render(fmt.Sprintf("Duration: %dms", duration.Milliseconds())))
	fmt.Println()
}

// CreateProgressCallback creates a callback function for progress updates
func CreateProgressCallback() (*ProgressWriter, func(string, int, int)) {
	pw := NewProgressWriter()

	callback := func(message string, current, total int) {
		pw.Update(message)
	}

	return pw, callback
}

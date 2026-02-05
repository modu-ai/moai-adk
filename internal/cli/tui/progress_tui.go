package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/modu-ai/moai-adk/internal/core/project"
)

// ProgressStep represents a single step in the initialization/update process
type ProgressStep struct {
	Name    string
	Percent float64
	Message string
	Status  string // "pending", "running", "done", "error"
}

// ProgressModel is a Bubble Tea model with modern progress bars
type ProgressModel struct {
	steps       []ProgressStep
	currentStep int
	progress    []progress.Model // Individual progress bars for each step
	quitting    bool
	width       int
	height      int
	startTime   time.Time
	totalTime   time.Duration
	title       string
	result      *project.InitResult
	err         error
	done        bool

	// initFn is the function to run for initialization
	initFn func(context.Context, project.ProgressReporter) (*project.InitResult, error)

	// msgCh receives messages from the init goroutine (chan is reference type)
	msgCh chan combinedMsg
}

// Message types for progress updates
type (
	progressTickMsg time.Time
	initStartedMsg  struct {
		msgCh chan combinedMsg
	}
	progressCompleteMsg struct {
		Result *project.InitResult
	}
	progressErrorMsg struct {
		Err error
	}
	stepUpdateMsg struct {
		Index   int
		Percent float64
		Message string
	}
	stepCompleteMsg struct {
		Index   int
		Message string
	}
	stepErrorMsg struct {
		Index int
		Err   error
	}
)

// InitFunc is the type for initialization functions
type InitFunc func(context.Context, project.ProgressReporter) (*project.InitResult, error)

// NewProgressModel creates a new progress TUI model
func NewProgressModel(title string, stepNames []string, initFn InitFunc) ProgressModel {
	// Create individual progress bars for each step
	progressBars := make([]progress.Model, len(stepNames))
	for i := range progressBars {
		// Use different gradient colors for each step
		colors := getStepColors(i)
		progressBars[i] = progress.New(
			progress.WithScaledGradient(colors[0], colors[1]),
		)
	}

	steps := make([]ProgressStep, len(stepNames))
	for i, name := range stepNames {
		status := "pending"
		steps[i] = ProgressStep{
			Name:    name,
			Status:  status,
			Percent: 0.0,
		}
	}

	return ProgressModel{
		steps:       steps,
		currentStep: 0,
		progress:    progressBars,
		startTime:   time.Now(),
		title:       title,
		initFn:      initFn,
	}
}

// getStepColors returns gradient colors for each step
func getStepColors(index int) [2]string {
	// Cycle through different color schemes
	colors := [][2]string{
		{"#FF7CCB", "#FDFF8C"}, // Pink to yellow
		{"#7D56F4", "#00F2EA"}, // Purple to cyan
		{"#FA7921", "#FAE521"}, // Orange to yellow
		{"#0475FF", "#00D9FA"}, // Blue to cyan
		{"#B967FF", "#F9A8FF"}, // Purple to light purple
		{"#FF6B6B", "#4ECDC4"}, // Red to teal
		{"#A8E6CF", "#FFD3B6"}, // Green to peach
		{"#FF8B94", "#FF6E9D"}, // Pink variations
	}
	return colors[index%len(colors)]
}

// waitForProgressMsg returns a command that polls the message channel
func (m ProgressModel) waitForProgressMsg() tea.Cmd {
	return func() tea.Msg {
		if m.msgCh == nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] waitForProgressMsg: msgCh is nil\n")
			return nil
		}
		msg, ok := <-m.msgCh
		if !ok {
			fmt.Fprintf(os.Stderr, "[DEBUG] waitForProgressMsg: channel closed\n")
			return nil
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] waitForProgressMsg: received msg, isResult=%v\n", msg.isResult)
		if msg.isResult {
			if msg.resultErr != nil {
				return progressErrorMsg{Err: msg.resultErr}
			}
			return progressCompleteMsg{Result: msg.result}
		}
		// Map stepProgress to appropriate message
		switch msg.step.status {
		case "running":
			return stepUpdateMsg{
				Index:   msg.step.index,
				Percent: msg.step.percent,
				Message: msg.step.message,
			}
		case "done":
			return stepCompleteMsg{
				Index:   msg.step.index,
				Message: msg.step.message,
			}
		case "error":
			return stepErrorMsg{
				Index: msg.step.index,
				Err:   msg.step.err,
			}
		}
		return nil
	}
}

// Init implements tea.Model (must use value receiver for Bubble Tea)
func (m ProgressModel) Init() tea.Cmd {
	// Create channel immediately
	msgCh := make(chan combinedMsg, 100)

	// Start init in background immediately
	go func() {
		fmt.Fprintf(os.Stderr, "[DEBUG] Init goroutine started\n")
		reporter := newChannelProgressReporter(msgCh)

		// Start the forwarder goroutine
		go func() {
			fmt.Fprintf(os.Stderr, "[DEBUG] Forwarder goroutine started\n")
			for sp := range reporter.stepCh {
				fmt.Fprintf(os.Stderr, "[DEBUG] Forwarder sending: index=%d status=%s\n", sp.index, sp.status)
				msgCh <- combinedMsg{step: sp, isResult: false}
			}
			fmt.Fprintf(os.Stderr, "[DEBUG] Forwarder goroutine ended\n")
		}()

		// Run init and collect result
		fmt.Fprintf(os.Stderr, "[DEBUG] Calling initFn...\n")
		res, err := m.initFn(context.Background(), reporter)
		reporter.close()
		fmt.Fprintf(os.Stderr, "[DEBUG] initFn completed: res=%v err=%v\n", res != nil, err)

		// Send final result
		msgCh <- combinedMsg{result: res, resultErr: err, isResult: true}
		fmt.Fprintf(os.Stderr, "[DEBUG] Final result sent\n")
	}()

	// Send initStartedMsg with msgCh so Update can set it
	return tea.Batch(
		func() tea.Msg {
			fmt.Fprintf(os.Stderr, "[DEBUG] Sending initStartedMsg\n")
			return initStartedMsg{msgCh: msgCh}
		},
		tea.Tick(time.Second, func(t time.Time) tea.Msg { return progressTickMsg(t) }),
	)
}

// Update implements tea.Model (must use value receiver for Bubble Tea)
func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	case initStartedMsg:
		// Set msgCh on model (this is the key fix!)
		m.msgCh = msg.msgCh
		fmt.Fprintf(os.Stderr, "[DEBUG] initStartedMsg received, msgCh set\n")
		// Mark first step as running
		if len(m.steps) > 0 {
			m.steps[0].Status = "running"
		}
		// Start waiting for progress messages
		return m, m.waitForProgressMsg()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update progress bar widths
		barWidth := msg.Width - 20
		if barWidth > 60 {
			barWidth = 60
		}
		for i := range m.progress {
			m.progress[i].Width = barWidth
		}
	case progressTickMsg:
		m.totalTime = time.Since(m.startTime)
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return progressTickMsg(t) })
	case stepUpdateMsg:
		if msg.Index < len(m.steps) {
			m.steps[msg.Index].Percent = msg.Percent
			m.steps[msg.Index].Message = msg.Message
			m.steps[msg.Index].Status = "running"
		}
		// Continue waiting for more messages
		return m, m.waitForProgressMsg()
	case stepCompleteMsg:
		if msg.Index < len(m.steps) {
			m.steps[msg.Index].Status = "done"
			m.steps[msg.Index].Percent = 1.0
			m.steps[msg.Index].Message = msg.Message
			// Start next step
			if msg.Index+1 < len(m.steps) {
				m.steps[msg.Index+1].Status = "running"
			}
		}
		// Continue waiting for more messages
		return m, m.waitForProgressMsg()
	case stepErrorMsg:
		if msg.Index < len(m.steps) {
			m.steps[msg.Index].Status = "error"
			m.steps[msg.Index].Message = msg.Err.Error()
			m.err = msg.Err
			m.quitting = true
			return m, tea.Quit
		}
	case progressCompleteMsg:
		m.result = msg.Result
		m.done = true
		m.quitting = true
		// Mark all steps as done
		for i := range m.steps {
			if m.steps[i].Status != "done" {
				m.steps[i].Status = "done"
				m.steps[i].Percent = 1.0
			}
		}
		return m, tea.Quit
	case progressErrorMsg:
		m.err = msg.Err
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// View implements tea.Model
func (m ProgressModel) View() string {
	if m.quitting {
		if m.err != nil {
			return m.renderError()
		}
		return m.renderComplete()
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF7CCB")).
		Bold(true).
		MarginBottom(1)

	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n")

	// Render each step with its progress bar
	for i, step := range m.steps {
		// Step status icon
		var icon string
		var iconColor lipgloss.Color

		switch step.Status {
		case "pending":
			icon = "○"
			iconColor = lipgloss.Color("#626262")
		case "running":
			icon = "⟳"
			iconColor = lipgloss.Color("#0475FF")
		case "done":
			icon = "✓"
			iconColor = lipgloss.Color("#00F2EA")
		case "error":
			icon = "✗"
			iconColor = lipgloss.Color("#FF6B6B")
		default:
			icon = "○"
			iconColor = lipgloss.Color("#626262")
		}

		// Step name with icon
		iconStyle := lipgloss.NewStyle().Foreground(iconColor).Bold(true)
		nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

		if step.Status == "running" {
			nameStyle = nameStyle.Bold(true)
		}

		b.WriteString(fmt.Sprintf("%s %s", iconStyle.Render(icon), nameStyle.Render(step.Name)))

		// Progress bar for this step
		if step.Status == "running" || step.Status == "done" {
			b.WriteString("\n")
			b.WriteString("  " + m.progress[i].ViewAs(step.Percent))
		}

		// Message if available
		if step.Message != "" {
			msgStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262")).
				Faint(true).
				MarginLeft(2)
			b.WriteString("\n")
			b.WriteString(msgStyle.Render("  " + step.Message))
		}

		b.WriteString("\n\n")
	}

	// Overall progress
	overallPercent := m.calculateOverallProgress()
	overallStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FDFF8C")).
		Bold(true)
	b.WriteString(overallStyle.Render(fmt.Sprintf("Overall Progress: %.0f%%", overallPercent*100)))
	b.WriteString("\n")

	// Footer with timing
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Faint(true).
		MarginTop(1)

	duration := m.totalTime
	if duration > 0 {
		b.WriteString(footerStyle.Render(fmt.Sprintf("⏱ %s elapsed", duration.Round(time.Millisecond))))
		b.WriteString("\n")
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Faint(true).
		MarginTop(1)
	b.WriteString(helpStyle.Render("Press Ctrl+C to quit"))

	return b.String()
}

// calculateOverallProgress calculates the overall progress percentage
func (m ProgressModel) calculateOverallProgress() float64 {
	if len(m.steps) == 0 {
		return 0.0
	}
	total := 0.0
	for _, step := range m.steps {
		total += step.Percent
	}
	return total / float64(len(m.steps))
}

// renderComplete renders the completion screen
func (m ProgressModel) renderComplete() string {
	var b strings.Builder

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00F2EA")).
		Bold(true).
		MarginBottom(1)

	b.WriteString(successStyle.Render("✓ Initialization Complete"))
	b.WriteString("\n\n")

	if m.result != nil {
		resultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		b.WriteString(resultStyle.Render(fmt.Sprintf("Development mode: %s\n", m.result.DevelopmentMode)))
		b.WriteString(resultStyle.Render(fmt.Sprintf("Created %d directories and %d files\n",
			len(m.result.CreatedDirs), len(m.result.CreatedFiles))))

		if len(m.result.Warnings) > 0 {
			b.WriteString("\n")
			warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FDFF8C"))
			b.WriteString(warningStyle.Render("Warnings:"))
			b.WriteString("\n")
			for _, w := range m.result.Warnings {
				b.WriteString(warningStyle.Render("  • " + w))
				b.WriteString("\n")
			}
		}
	}

	durationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Faint(true).
		MarginTop(1)
	b.WriteString(durationStyle.Render(fmt.Sprintf("⏱ Total time: %s", m.totalTime.Round(time.Millisecond))))
	b.WriteString("\n")

	return b.String()
}

// renderError renders the error screen
func (m ProgressModel) renderError() string {
	var b strings.Builder

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true).
		MarginBottom(1)

	b.WriteString(errorStyle.Render("✗ Initialization Failed"))
	b.WriteString("\n\n")

	if m.err != nil {
		msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		b.WriteString(msgStyle.Render(m.err.Error()))
		b.WriteString("\n")
	}

	return b.String()
}

// GetResult returns the initialization result
func (m *ProgressModel) GetResult() *project.InitResult {
	return m.result
}

// GetError returns any error that occurred
func (m *ProgressModel) GetError() error {
	return m.err
}

// channelProgressReporter sends progress updates via channels
type channelProgressReporter struct {
	stepCh    chan stepProgress
	mu        sync.Mutex
	closed    bool
	stepIndex int
}

type stepProgress struct {
	index   int
	percent float64
	message string
	status  string
	err     error
}

// combinedMsg represents either a step progress update or a final result
type combinedMsg struct {
	step      stepProgress
	result    *project.InitResult
	resultErr error
	isResult  bool
}

func newChannelProgressReporter(msgCh chan combinedMsg) *channelProgressReporter {
	return &channelProgressReporter{
		stepCh: make(chan stepProgress, 100),
	}
}

func (r *channelProgressReporter) StepStart(name, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] StepStart called: name=%s message=%s index=%d\n", name, message, r.stepIndex)
	r.stepCh <- stepProgress{
		index:   r.stepIndex,
		percent: 0.0,
		message: message,
		status:  "running",
	}
}

func (r *channelProgressReporter) StepUpdate(message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	r.stepCh <- stepProgress{
		index:   r.stepIndex,
		percent: 0.5,
		message: message,
		status:  "running",
	}
}

func (r *channelProgressReporter) StepComplete(message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	r.stepCh <- stepProgress{
		index:   r.stepIndex,
		percent: 1.0,
		message: message,
		status:  "done",
	}
	r.stepIndex++
}

func (r *channelProgressReporter) StepError(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	r.stepCh <- stepProgress{
		index:   r.stepIndex,
		percent: 0,
		message: "",
		status:  "error",
		err:     err,
	}
}

func (r *channelProgressReporter) close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.closed {
		r.closed = true
		close(r.stepCh)
	}
}

// RunProgressTUI starts the TUI with the given initialization function
func RunProgressTUI(title string, stepNames []string, initFn InitFunc) (*project.InitResult, error) {
	model := NewProgressModel(title, stepNames, initFn)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	m, ok := finalModel.(ProgressModel)
	if !ok {
		return nil, fmt.Errorf("unexpected model type")
	}

	return m.GetResult(), m.GetError()
}

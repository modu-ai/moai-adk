package ui

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// progressImpl implements the Progress interface.
type progressImpl struct {
	theme    *Theme
	headless *HeadlessManager
	writer   io.Writer
}

// NewProgress creates a Progress backed by the given theme and headless manager.
// Output goes to os.Stdout.
func NewProgress(theme *Theme, hm *HeadlessManager) Progress {
	return &progressImpl{theme: theme, headless: hm, writer: os.Stdout}
}

// newProgressImpl creates a progressImpl with a custom writer (for testing).
func newProgressImpl(theme *Theme, hm *HeadlessManager, w io.Writer) *progressImpl {
	return &progressImpl{theme: theme, headless: hm, writer: w}
}

// Start creates a determinate progress bar with the given total.
// In headless mode it returns a log-based progress bar.
func (p *progressImpl) Start(title string, total int) ProgressBar {
	if p.headless.IsHeadless() || p.theme.NoColor {
		return newHeadlessProgressBar(p.theme, title, total, p.writer)
	}
	return newInteractiveProgressBar(p.theme, title, total)
}

// Spinner creates an indeterminate spinner.
// In headless mode it prints the title as a log line.
func (p *progressImpl) Spinner(title string) Spinner {
	if p.headless.IsHeadless() || p.theme.NoColor {
		return newHeadlessSpinner(p.theme, title, p.writer)
	}
	return newInteractiveSpinner(p.theme, title)
}

// --- interactiveSpinner ---

// spinnerTitleMsg is sent to update the spinner title.
type spinnerTitleMsg string

// spinnerStopMsg is sent to stop the spinner.
type spinnerStopMsg struct{}

// spinnerModel is the bubbletea Model for the animated spinner.
type spinnerModel struct {
	spinner spinner.Model
	title   string
	done    bool
}

func newSpinnerModel(theme *Theme, title string) spinnerModel {
	s := spinner.New(spinner.WithSpinner(spinner.Dot))
	if !theme.NoColor {
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.Primary))
	}
	return spinnerModel{spinner: s, title: title}
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinnerTitleMsg:
		m.title = string(msg)
		return m, nil
	case spinnerStopMsg:
		m.done = true
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return m.spinner.View() + " " + m.title + "\n"
}

// interactiveSpinner implements Spinner with an animated bubbles spinner.
type interactiveSpinner struct {
	program *tea.Program
	once    sync.Once
}

// @MX:WARN: [AUTO] 고루틴에 채널 전송으로 완료 신호를 보냅니다. 수신자가 없으면 영구 차단될 수 있습니다.
// @MX:REASON: [AUTO] 고루틴 수명 주기가 테아 프로그램의 수명 주기에 종속됩니다
func newInteractiveSpinner(theme *Theme, title string) *interactiveSpinner {
	m := newSpinnerModel(theme, title)
	p := tea.NewProgram(m)

	s := &interactiveSpinner{program: p}

	go func() {
		_, _ = p.Run()
	}()

	return s
}

// SetTitle updates the spinner title.
func (s *interactiveSpinner) SetTitle(title string) {
	s.program.Send(spinnerTitleMsg(title))
}

// Stop halts the spinner.
func (s *interactiveSpinner) Stop() {
	s.once.Do(func() {
		s.program.Send(spinnerStopMsg{})
		s.program.Wait()
	})
}

// --- interactiveProgressBar ---

// progressIncrMsg is sent to increment the progress bar.
type progressIncrMsg int

// progressTitleMsg is sent to update the progress bar title.
type progressTitleMsg string

// progressDoneMsg is sent to complete the progress bar.
type progressDoneMsg struct{}

// progressModel is the bubbletea Model for the animated progress bar.
type progressModel struct {
	bar     progress.Model
	title   string
	current int
	total   int
	done    bool
}

func newProgressModel(theme *Theme, title string, total int) progressModel {
	bar := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)
	if !theme.NoColor {
		bar = progress.New(
			progress.WithGradient(theme.Colors.Primary, theme.Colors.Secondary),
			progress.WithWidth(40),
		)
	}
	return progressModel{bar: bar, title: title, total: total}
}

func (m progressModel) Init() tea.Cmd {
	return nil
}

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case progressIncrMsg:
		m.current += int(msg)
		if m.current > m.total {
			m.current = m.total
		}
		return m, nil
	case progressTitleMsg:
		m.title = string(msg)
		return m, nil
	case progressDoneMsg:
		m.current = m.total
		m.done = true
		return m, tea.Quit
	case progress.FrameMsg:
		pm, cmd := m.bar.Update(msg)
		m.bar = pm.(progress.Model)
		return m, cmd
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m progressModel) View() string {
	if m.done {
		return ""
	}
	pct := 0.0
	if m.total > 0 {
		pct = float64(m.current) / float64(m.total)
	}
	return m.bar.ViewAs(pct) + " " + fmt.Sprintf("[%d/%d] %s\n", m.current, m.total, m.title)
}

// interactiveProgressBar implements ProgressBar with an animated bubbles progress bar.
type interactiveProgressBar struct {
	program *tea.Program
	once    sync.Once
}

// @MX:WARN: [AUTO] 고루틴에 채널 전송으로 완료 신호를 보냅니다. 수신자가 없으면 영구 차단될 수 있습니다.
// @MX:REASON: [AUTO] 고루틴 수명 주기가 테아 프로그램의 수명 주기에 종속됩니다
func newInteractiveProgressBar(theme *Theme, title string, total int) *interactiveProgressBar {
	m := newProgressModel(theme, title, total)
	p := tea.NewProgram(m)

	pb := &interactiveProgressBar{program: p}

	go func() {
		_, _ = p.Run()
	}()

	return pb
}

// Increment advances the progress by n.
func (b *interactiveProgressBar) Increment(n int) {
	b.program.Send(progressIncrMsg(n))
}

// SetTitle updates the progress bar title.
func (b *interactiveProgressBar) SetTitle(title string) {
	b.program.Send(progressTitleMsg(title))
}

// Done completes the progress bar at 100%.
func (b *interactiveProgressBar) Done() {
	b.once.Do(func() {
		b.program.Send(progressDoneMsg{})
		b.program.Wait()
	})
}

// --- headlessProgressBar ---

// headlessProgressBar implements ProgressBar with plain text log output.
type headlessProgressBar struct {
	theme   *Theme
	title   string
	total   int
	current int
	writer  io.Writer
}

// newHeadlessProgressBar creates a headless progress bar that writes log lines.
func newHeadlessProgressBar(theme *Theme, title string, total int, w io.Writer) *headlessProgressBar {
	return &headlessProgressBar{
		theme:  theme,
		title:  title,
		total:  total,
		writer: w,
	}
}

// Increment advances the progress by n and writes a log line.
func (b *headlessProgressBar) Increment(n int) {
	b.current += n
	if b.current > b.total {
		b.current = b.total
	}
	_, _ = fmt.Fprintf(b.writer, "[%d/%d] %s\n", b.current, b.total, b.title)
}

// SetTitle updates the progress bar title.
func (b *headlessProgressBar) SetTitle(title string) {
	b.title = title
}

// Done completes the progress bar at 100%.
func (b *headlessProgressBar) Done() {
	b.current = b.total
	_, _ = fmt.Fprintf(b.writer, "[%d/%d] %s\n", b.current, b.total, b.title)
}

// --- headlessSpinner ---

// headlessSpinner implements Spinner with plain text log output.
type headlessSpinner struct {
	theme   *Theme
	title   string
	writer  io.Writer
	stopped bool
}

// newHeadlessSpinner creates a headless spinner that prints the title.
func newHeadlessSpinner(theme *Theme, title string, w io.Writer) *headlessSpinner {
	s := &headlessSpinner{
		theme:  theme,
		title:  title,
		writer: w,
	}
	_, _ = fmt.Fprintf(w, "%s\n", title)
	return s
}

// SetTitle updates the spinner title and prints a log line.
func (s *headlessSpinner) SetTitle(title string) {
	s.title = title
	_, _ = fmt.Fprintf(s.writer, "%s\n", title)
}

// Stop halts the spinner.
func (s *headlessSpinner) Stop() {
	s.stopped = true
}

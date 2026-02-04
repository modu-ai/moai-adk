package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
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
	return newHeadlessProgressBar(p.theme, title, total, p.writer)
}

// Spinner creates an indeterminate spinner.
// In headless mode it prints the title as a log line.
func (p *progressImpl) Spinner(title string) Spinner {
	return newHeadlessSpinner(p.theme, title, p.writer)
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

// --- ASCII progress bar rendering ---

// renderASCIIBar renders a progress bar using ASCII characters: [=====>     ]
func renderASCIIBar(current, total, width int) string {
	if total <= 0 {
		return "[" + strings.Repeat(" ", width) + "]"
	}

	ratio := float64(current) / float64(total)
	if ratio > 1.0 {
		ratio = 1.0
	}

	filled := int(ratio * float64(width))
	empty := width - filled

	var bar strings.Builder
	bar.WriteString("[")
	if filled > 0 {
		bar.WriteString(strings.Repeat("=", filled))
	}
	if empty > 0 {
		bar.WriteString(strings.Repeat(" ", empty))
	}
	bar.WriteString("]")

	return bar.String()
}

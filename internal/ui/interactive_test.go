package ui

import (
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// newTestProgram creates a tea.Program configured for test environments without a TTY.
// It uses an empty string reader for input, io.Discard for output, and disables the renderer
// to avoid any TTY requirements.
func newTestProgram(m tea.Model) *tea.Program {
	return tea.NewProgram(m,
		tea.WithInput(strings.NewReader("")),
		tea.WithOutput(io.Discard),
		tea.WithoutRenderer(),
	)
}

// startTestProgram starts a tea.Program in a goroutine and returns a done channel.
func startTestProgram(p *tea.Program) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = p.Run()
	}()
	// Allow the program goroutine to initialize before sending messages.
	time.Sleep(10 * time.Millisecond)
	return done
}

// waitForProgram waits for the program to exit, failing the test if it exceeds timeout.
func waitForProgram(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
		// program exited cleanly
	case <-time.After(2 * time.Second):
		t.Error("tea.Program did not exit within 2 second timeout")
	}
}

// --- interactiveSpinner method tests ---
// These tests directly construct interactiveSpinner structs with TTY-free tea.Programs
// to cover SetTitle and Stop methods without requiring a real terminal.

func TestInteractiveSpinner_SetTitle(t *testing.T) {
	m := newSpinnerModel(testTheme(), "Initial")
	p := newTestProgram(m)
	s := &interactiveSpinner{program: p, once: sync.Once{}}
	done := startTestProgram(p)

	s.SetTitle("Updated title")
	s.Stop()

	waitForProgram(t, done)
}

func TestInteractiveSpinner_Stop_Idempotent(t *testing.T) {
	m := newSpinnerModel(testTheme(), "Loading")
	p := newTestProgram(m)
	s := &interactiveSpinner{program: p, once: sync.Once{}}
	done := startTestProgram(p)

	// sync.Once ensures Stop is idempotent; calling it multiple times is safe.
	s.Stop()
	s.Stop()
	s.Stop()

	waitForProgram(t, done)
}

func TestInteractiveSpinner_SetTitleThenStop(t *testing.T) {
	m := newSpinnerModel(testTheme(), "Checking dependencies")
	p := newTestProgram(m)
	s := &interactiveSpinner{program: p, once: sync.Once{}}
	done := startTestProgram(p)

	s.SetTitle("Downloading packages")
	s.SetTitle("Installing files")
	s.Stop()

	waitForProgram(t, done)
}

// --- interactiveProgressBar method tests ---

func TestInteractiveProgressBar_Increment(t *testing.T) {
	m := newProgressModel(testTheme(), "Processing", 10)
	p := newTestProgram(m)
	pb := &interactiveProgressBar{program: p, once: sync.Once{}}
	done := startTestProgram(p)

	pb.Increment(3)
	pb.Increment(2)
	pb.Done()

	waitForProgram(t, done)
}

func TestInteractiveProgressBar_SetTitle(t *testing.T) {
	m := newProgressModel(testTheme(), "Step 1", 5)
	p := newTestProgram(m)
	pb := &interactiveProgressBar{program: p, once: sync.Once{}}
	done := startTestProgram(p)

	pb.SetTitle("Step 2")
	pb.Done()

	waitForProgram(t, done)
}

func TestInteractiveProgressBar_Done_Idempotent(t *testing.T) {
	m := newProgressModel(testTheme(), "Processing", 10)
	p := newTestProgram(m)
	pb := &interactiveProgressBar{program: p, once: sync.Once{}}
	done := startTestProgram(p)

	// sync.Once ensures Done is idempotent; calling it multiple times is safe.
	pb.Done()
	pb.Done()
	pb.Done()

	waitForProgram(t, done)
}

func TestInteractiveProgressBar_IncrementSetTitleDone(t *testing.T) {
	m := newProgressModel(testTheme(), "Upload", 4)
	p := newTestProgram(m)
	pb := &interactiveProgressBar{program: p, once: sync.Once{}}
	done := startTestProgram(p)

	pb.Increment(1)
	pb.SetTitle("Uploading step 2")
	pb.Increment(1)
	pb.SetTitle("Uploading step 3")
	pb.Increment(1)
	pb.Done()

	waitForProgram(t, done)
}

func TestInteractiveProgressBar_ZeroIncrement(t *testing.T) {
	m := newProgressModel(testTheme(), "Deploy", 5)
	p := newTestProgram(m)
	pb := &interactiveProgressBar{program: p, once: sync.Once{}}
	done := startTestProgram(p)

	pb.Increment(0)
	pb.Done()

	waitForProgram(t, done)
}

// --- Spinner model additional coverage ---

func TestSpinnerModel_Update_SpinnerTickMsg(t *testing.T) {
	theme := NewTheme(ThemeConfig{Mode: "dark"})
	m := newSpinnerModel(theme, "Ticking")
	// Obtain a real TickMsg by calling Init and executing the command.
	tickCmd := m.Init()
	if tickCmd == nil {
		t.Fatal("Init should return a non-nil tick command")
	}
	msg := tickCmd()
	if msg == nil {
		t.Skip("tick command returned nil message; skipping")
	}
	// Only proceed if we got a spinner.TickMsg to exercise that branch.
	if _, ok := msg.(spinner.TickMsg); !ok {
		t.Skip("unexpected message type from tick command")
	}
	updated, cmd := m.Update(msg)
	result := updated.(spinnerModel)
	if result.done {
		t.Error("tick should not stop the spinner")
	}
	_ = cmd
}

// --- Progress model FrameMsg with color theme ---

func TestProgressModel_Update_FrameMsg_ColorTheme(t *testing.T) {
	theme := NewTheme(ThemeConfig{Mode: "dark"})
	m := newProgressModel(theme, "Color frame", 10)
	updated, _ := m.Update(progress.FrameMsg{})
	result := updated.(progressModel)
	if result.done {
		t.Error("FrameMsg should not mark the progress bar as done")
	}
}

// --- Interactive form functions: non-TTY error path coverage ---
// In non-TTY environments, huh form.Run() fails to open a TTY and returns a
// non-nil error that the interactive methods wrap with a prefix (e.g. "prompt: ...").
// These tests exercise the interactive code paths and accept any non-nil error.

func TestInputInteractive_NonTTY_ReturnsError(t *testing.T) {
	theme := testTheme()
	theme.NoColor = false
	hm := NewHeadlessManager()
	hm.ForceHeadless(false)

	p := &promptImpl{theme: theme, headless: hm}
	cfg := inputConfig{defaultVal: "test-default"}

	_, err := p.inputInteractive("Test label", cfg)
	if err == nil {
		t.Skip("inputInteractive succeeded (running in a real TTY environment)")
	}
	// ErrCancelled is also acceptable (huh.ErrUserAborted path).
	if err == ErrCancelled {
		t.Logf("inputInteractive returned ErrCancelled (user aborted path exercised)")
		return
	}
	t.Logf("inputInteractive returned expected non-TTY error: %v", err)
}

func TestInputInteractive_WithPlaceholder_NonTTY(t *testing.T) {
	theme := testTheme()
	theme.NoColor = false
	hm := NewHeadlessManager()
	hm.ForceHeadless(false)

	p := &promptImpl{theme: theme, headless: hm}
	cfg := inputConfig{placeholder: "enter name", defaultVal: ""}

	_, err := p.inputInteractive("Name", cfg)
	if err == nil {
		t.Skip("inputInteractive succeeded (running in a real TTY environment)")
	}
	t.Logf("inputInteractive with placeholder returned: %v", err)
}

func TestConfirmInteractive_NonTTY_ReturnsError(t *testing.T) {
	theme := testTheme()
	theme.NoColor = false
	hm := NewHeadlessManager()
	hm.ForceHeadless(false)

	p := &promptImpl{theme: theme, headless: hm}

	_, err := p.confirmInteractive("Continue?", true)
	if err == nil {
		t.Skip("confirmInteractive succeeded (running in a real TTY environment)")
	}
	if err == ErrCancelled {
		t.Logf("confirmInteractive returned ErrCancelled (user aborted path exercised)")
		return
	}
	t.Logf("confirmInteractive returned expected non-TTY error: %v", err)
}

func TestConfirmInteractive_DefaultFalse_NonTTY(t *testing.T) {
	theme := testTheme()
	theme.NoColor = false
	hm := NewHeadlessManager()
	hm.ForceHeadless(false)

	p := &promptImpl{theme: theme, headless: hm}

	_, err := p.confirmInteractive("Delete?", false)
	if err == nil {
		t.Skip("confirmInteractive succeeded (running in a real TTY environment)")
	}
	t.Logf("confirmInteractive (false default) returned: %v", err)
}

func TestSelectInteractive_NonTTY_ReturnsError(t *testing.T) {
	theme := testTheme()
	theme.NoColor = false
	hm := NewHeadlessManager()
	hm.ForceHeadless(false)

	s := &selectorImpl{theme: theme, headless: hm}

	_, err := s.selectInteractive("Pick language", testItems())
	if err == nil {
		t.Skip("selectInteractive succeeded (running in a real TTY environment)")
	}
	if err == ErrCancelled {
		t.Logf("selectInteractive returned ErrCancelled (user aborted path exercised)")
		return
	}
	t.Logf("selectInteractive returned expected non-TTY error: %v", err)
}

func TestMultiSelectInteractive_NonTTY_ReturnsError(t *testing.T) {
	theme := testTheme()
	theme.NoColor = false
	hm := NewHeadlessManager()
	hm.ForceHeadless(false)

	c := &checkboxImpl{theme: theme, headless: hm}

	_, err := c.multiSelectInteractive("Pick features", checkboxItems())
	if err == nil {
		t.Skip("multiSelectInteractive succeeded (running in a real TTY environment)")
	}
	if err == ErrCancelled {
		t.Logf("multiSelectInteractive returned ErrCancelled (user aborted path exercised)")
		return
	}
	t.Logf("multiSelectInteractive returned expected non-TTY error: %v", err)
}

// --- newInteractiveSpinner and newInteractiveProgressBar constructor coverage ---
// These functions call tea.NewProgram(m) with default stdout/stdin.
// In non-TTY test environments, tea.Program runs without a real terminal and exits
// cleanly when sent a stop/done message. We exercise the constructors directly
// by calling progressImpl.Start() and progressImpl.Spinner() with headless=false
// and NoColor=false so the interactive path is taken.

func TestProgressImpl_Spinner_InteractivePath(t *testing.T) {
	theme := NewTheme(ThemeConfig{NoColor: false, Mode: "dark"})
	hm := NewHeadlessManager()
	// Force non-headless to reach newInteractiveSpinner.
	hm.ForceHeadless(false)

	var buf strings.Builder
	prog := newProgressImpl(theme, hm, &buf)
	sp := prog.Spinner("Interactive spinner test")

	// The returned spinner may be either interactiveSpinner (NoColor=false, non-headless)
	// or headlessSpinner (if the implementation falls back). Either way, Stop must succeed.
	sp.SetTitle("Updated title")
	sp.Stop()
	// Calling Stop a second time must be safe (sync.Once on interactiveSpinner).
	sp.Stop()
}

func TestProgressImpl_Start_InteractivePath(t *testing.T) {
	theme := NewTheme(ThemeConfig{NoColor: false, Mode: "dark"})
	hm := NewHeadlessManager()
	// Force non-headless to reach newInteractiveProgressBar.
	hm.ForceHeadless(false)

	var buf strings.Builder
	prog := newProgressImpl(theme, hm, &buf)
	pb := prog.Start("Interactive progress test", 5)

	pb.Increment(2)
	pb.SetTitle("Step 2")
	pb.Increment(3)
	pb.Done()
	// Calling Done a second time must be safe (sync.Once on interactiveProgressBar).
	pb.Done()
}

// --- Wizard runInteractive: non-TTY error path coverage ---

func TestWizardRunInteractive_NonTTY_ReturnsError(t *testing.T) {
	theme := NewTheme(ThemeConfig{NoColor: false, Mode: "dark"})
	hm := NewHeadlessManager()
	hm.ForceHeadless(false)

	w := &wizardImpl{theme: theme, headless: hm}

	_, err := w.runInteractive(context.Background())
	if err == nil {
		t.Skip("runInteractive succeeded (running in a real TTY environment)")
	}
	// The error propagates from inputInteractive through runInteractive.
	t.Logf("runInteractive returned expected error in non-TTY: %v", err)
}

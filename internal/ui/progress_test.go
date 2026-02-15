package ui

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// --- headlessProgressBar unit tests ---

func TestNewProgressBar(t *testing.T) {
	var buf bytes.Buffer
	pb := newHeadlessProgressBar(testTheme(), "Deploying", 10, &buf)
	if pb.title != "Deploying" {
		t.Errorf("expected title 'Deploying', got %q", pb.title)
	}
	if pb.total != 10 {
		t.Error("expected total 10")
	}
	if pb.current != 0 {
		t.Error("expected current 0")
	}
}

func TestProgressBar_Increment(t *testing.T) {
	var buf bytes.Buffer
	pb := newHeadlessProgressBar(testTheme(), "Processing", 10, &buf)
	pb.Increment(3)
	if pb.current != 3 {
		t.Errorf("expected current 3, got %d", pb.current)
	}
	output := buf.String()
	if !strings.Contains(output, "[3/10]") {
		t.Errorf("expected '[3/10]' in output, got %q", output)
	}
	if !strings.Contains(output, "Processing") {
		t.Errorf("expected 'Processing' in output, got %q", output)
	}
}

func TestProgressBar_IncrementMultiple(t *testing.T) {
	var buf bytes.Buffer
	pb := newHeadlessProgressBar(testTheme(), "Processing", 5, &buf)
	pb.Increment(1)
	pb.Increment(1)
	pb.Increment(1)
	if pb.current != 3 {
		t.Errorf("expected current 3, got %d", pb.current)
	}
	output := buf.String()
	if !strings.Contains(output, "[1/5]") {
		t.Error("expected '[1/5]' in output")
	}
	if !strings.Contains(output, "[2/5]") {
		t.Error("expected '[2/5]' in output")
	}
	if !strings.Contains(output, "[3/5]") {
		t.Error("expected '[3/5]' in output")
	}
}

func TestProgressBar_IncrementBeyondTotal(t *testing.T) {
	var buf bytes.Buffer
	pb := newHeadlessProgressBar(testTheme(), "Processing", 3, &buf)
	pb.Increment(5)
	if pb.current != 3 {
		t.Errorf("expected current capped at 3, got %d", pb.current)
	}
}

func TestProgressBar_SetTitle(t *testing.T) {
	var buf bytes.Buffer
	pb := newHeadlessProgressBar(testTheme(), "Step 1", 10, &buf)
	pb.SetTitle("Step 2")
	if pb.title != "Step 2" {
		t.Errorf("expected title 'Step 2', got %q", pb.title)
	}
}

func TestProgressBar_Done(t *testing.T) {
	var buf bytes.Buffer
	pb := newHeadlessProgressBar(testTheme(), "Processing", 10, &buf)
	pb.Done()
	if pb.current != 10 {
		t.Errorf("expected current 10 after Done, got %d", pb.current)
	}
	output := buf.String()
	if !strings.Contains(output, "[10/10]") {
		t.Errorf("expected '[10/10]' in output, got %q", output)
	}
}

// --- headlessSpinner unit tests ---

func TestNewSpinner(t *testing.T) {
	var buf bytes.Buffer
	sp := newHeadlessSpinner(testTheme(), "Loading...", &buf)
	if sp.title != "Loading..." {
		t.Errorf("expected title 'Loading...', got %q", sp.title)
	}
	output := buf.String()
	if !strings.Contains(output, "Loading...") {
		t.Errorf("expected 'Loading...' in output, got %q", output)
	}
}

func TestSpinner_SetTitle(t *testing.T) {
	var buf bytes.Buffer
	sp := newHeadlessSpinner(testTheme(), "Loading...", &buf)
	sp.SetTitle("Downloading...")
	if sp.title != "Downloading..." {
		t.Errorf("expected title 'Downloading...', got %q", sp.title)
	}
	output := buf.String()
	if !strings.Contains(output, "Downloading...") {
		t.Errorf("expected 'Downloading...' in output, got %q", output)
	}
}

func TestSpinner_Stop(t *testing.T) {
	var buf bytes.Buffer
	sp := newHeadlessSpinner(testTheme(), "Loading...", &buf)
	sp.Stop()
	if !sp.stopped {
		t.Error("expected stopped to be true after Stop()")
	}
}

// --- Headless progress integration tests ---

func TestProgressHeadless_Start(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)

	var buf bytes.Buffer
	prog := newProgressImpl(theme, hm, &buf)
	pb := prog.Start("Deploying templates", 10)
	pb.Increment(1)
	pb.Increment(1)
	pb.Increment(1)

	output := buf.String()
	if !strings.Contains(output, "[1/10]") {
		t.Error("expected '[1/10]' in headless output")
	}
	if !strings.Contains(output, "[2/10]") {
		t.Error("expected '[2/10]' in headless output")
	}
	if !strings.Contains(output, "[3/10]") {
		t.Error("expected '[3/10]' in headless output")
	}
}

func TestProgressHeadless_Spinner(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)

	var buf bytes.Buffer
	prog := newProgressImpl(theme, hm, &buf)
	sp := prog.Spinner("Checking for updates...")

	output := buf.String()
	if !strings.Contains(output, "Checking for updates...") {
		t.Error("expected spinner title in headless output")
	}

	sp.SetTitle("Downloading...")
	output = buf.String()
	if !strings.Contains(output, "Downloading...") {
		t.Error("expected updated title in headless output")
	}

	sp.Stop()
}

func TestProgressNoColor_HeadlessBar(t *testing.T) {
	theme := NewTheme(ThemeConfig{NoColor: true})
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)

	var buf bytes.Buffer
	prog := newProgressImpl(theme, hm, &buf)
	pb := prog.Start("Processing", 10)
	pb.Increment(5)

	output := buf.String()
	if !strings.Contains(output, "[5/10]") {
		t.Errorf("expected '[5/10]' in output, got %q", output)
	}
}

// --- spinnerModel bubbletea model unit tests ---

func TestSpinnerModel_Init(t *testing.T) {
	m := newSpinnerModel(testTheme(), "Loading...")
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return spinner.Tick cmd")
	}
}

func TestSpinnerModel_Update_TitleMsg(t *testing.T) {
	m := newSpinnerModel(testTheme(), "Loading...")
	updated, cmd := m.Update(spinnerTitleMsg("New title"))
	result := updated.(spinnerModel)
	if result.title != "New title" {
		t.Errorf("expected 'New title', got %q", result.title)
	}
	if cmd != nil {
		t.Error("expected nil cmd for title update")
	}
}

func TestSpinnerModel_Update_StopMsg(t *testing.T) {
	m := newSpinnerModel(testTheme(), "Loading...")
	updated, cmd := m.Update(spinnerStopMsg{})
	result := updated.(spinnerModel)
	if !result.done {
		t.Error("expected done=true after stop")
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd")
	}
}

func TestSpinnerModel_Update_CtrlC(t *testing.T) {
	m := newSpinnerModel(testTheme(), "Loading...")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	result := updated.(spinnerModel)
	if !result.done {
		t.Error("expected done=true after Ctrl+C")
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd")
	}
}

func TestSpinnerModel_Update_OtherKey_Ignored(t *testing.T) {
	m := newSpinnerModel(testTheme(), "Loading...")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(spinnerModel)
	if result.done {
		t.Error("expected done=false for non-CtrlC key")
	}
	if cmd != nil {
		t.Error("expected nil cmd for non-CtrlC key")
	}
}

func TestSpinnerModel_View_Running(t *testing.T) {
	m := newSpinnerModel(testTheme(), "Loading...")
	view := m.View()
	if !strings.Contains(view, "Loading...") {
		t.Error("view should contain title")
	}
}

func TestSpinnerModel_View_Done(t *testing.T) {
	m := newSpinnerModel(testTheme(), "Loading...")
	m.done = true
	view := m.View()
	if view != "" {
		t.Error("done view should be empty")
	}
}

// --- progressModel bubbletea model unit tests ---

func TestProgressModel_Init(t *testing.T) {
	m := newProgressModel(testTheme(), "Processing", 10)
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should return nil cmd")
	}
}

func TestProgressModel_Update_IncrMsg(t *testing.T) {
	m := newProgressModel(testTheme(), "Processing", 10)
	updated, cmd := m.Update(progressIncrMsg(3))
	result := updated.(progressModel)
	if result.current != 3 {
		t.Errorf("expected current 3, got %d", result.current)
	}
	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestProgressModel_Update_IncrMsg_Accumulates(t *testing.T) {
	m := newProgressModel(testTheme(), "Processing", 10)
	updated, _ := m.Update(progressIncrMsg(3))
	m = updated.(progressModel)
	updated, _ = m.Update(progressIncrMsg(4))
	result := updated.(progressModel)
	if result.current != 7 {
		t.Errorf("expected current 7, got %d", result.current)
	}
}

func TestProgressModel_Update_IncrMsg_Caps(t *testing.T) {
	m := newProgressModel(testTheme(), "Processing", 5)
	updated, _ := m.Update(progressIncrMsg(10))
	result := updated.(progressModel)
	if result.current != 5 {
		t.Errorf("expected current capped at 5, got %d", result.current)
	}
}

func TestProgressModel_Update_TitleMsg(t *testing.T) {
	m := newProgressModel(testTheme(), "Step 1", 10)
	updated, _ := m.Update(progressTitleMsg("Step 2"))
	result := updated.(progressModel)
	if result.title != "Step 2" {
		t.Errorf("expected 'Step 2', got %q", result.title)
	}
}

func TestProgressModel_Update_DoneMsg(t *testing.T) {
	m := newProgressModel(testTheme(), "Processing", 10)
	updated, cmd := m.Update(progressDoneMsg{})
	result := updated.(progressModel)
	if result.current != 10 {
		t.Errorf("expected current 10, got %d", result.current)
	}
	if !result.done {
		t.Error("expected done=true")
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd")
	}
}

func TestProgressModel_Update_CtrlC(t *testing.T) {
	m := newProgressModel(testTheme(), "Processing", 10)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	result := updated.(progressModel)
	if !result.done {
		t.Error("expected done=true after Ctrl+C")
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd")
	}
}

func TestProgressModel_Update_OtherKey_Ignored(t *testing.T) {
	m := newProgressModel(testTheme(), "Processing", 10)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(progressModel)
	if result.done {
		t.Error("expected done=false for non-CtrlC key")
	}
	if cmd != nil {
		t.Error("expected nil cmd for non-CtrlC key")
	}
}

func TestProgressModel_View_Running(t *testing.T) {
	m := newProgressModel(testTheme(), "Processing", 10)
	m.current = 5
	view := m.View()
	if !strings.Contains(view, "Processing") {
		t.Error("view should contain title")
	}
	if !strings.Contains(view, fmt.Sprintf("[%d/%d]", 5, 10)) {
		t.Error("view should contain progress [5/10]")
	}
}

func TestProgressModel_View_Done(t *testing.T) {
	m := newProgressModel(testTheme(), "Processing", 10)
	m.done = true
	view := m.View()
	if view != "" {
		t.Error("done view should be empty")
	}
}

func TestProgressModel_View_ZeroTotal(t *testing.T) {
	m := newProgressModel(testTheme(), "Processing", 0)
	view := m.View()
	if !strings.Contains(view, "[0/0]") {
		t.Error("view should contain [0/0]")
	}
}

func TestProgressModel_NoColor(t *testing.T) {
	theme := NewTheme(ThemeConfig{NoColor: true})
	m := newProgressModel(theme, "Processing", 10)
	// No-color mode should use default gradient
	view := m.View()
	if !strings.Contains(view, "Processing") {
		t.Error("view should contain title even in no-color mode")
	}
}

// --- Model branch coverage tests ---

func TestSpinnerModel_WithColor(t *testing.T) {
	// Use a theme with NoColor=false to cover the color style branch
	theme := NewTheme(ThemeConfig{Mode: "dark"})
	m := newSpinnerModel(theme, "Color spinner")
	if m.title != "Color spinner" {
		t.Errorf("expected title 'Color spinner', got %q", m.title)
	}
	view := m.View()
	if !strings.Contains(view, "Color spinner") {
		t.Error("view should contain title in color mode")
	}
}

func TestProgressModel_WithColor(t *testing.T) {
	// Use a theme with NoColor=false to cover the color gradient branch
	theme := NewTheme(ThemeConfig{Mode: "dark"})
	m := newProgressModel(theme, "Color processing", 10)
	m.current = 5
	view := m.View()
	if !strings.Contains(view, "Color processing") {
		t.Error("view should contain title in color mode")
	}
	if !strings.Contains(view, "[5/10]") {
		t.Error("view should contain progress in color mode")
	}
}

func TestSpinnerModel_Update_GenericMsg(t *testing.T) {
	m := newSpinnerModel(testTheme(), "Loading...")
	// Send an unknown message type (string); should return model unchanged
	updated, cmd := m.Update("unknown message")
	result := updated.(spinnerModel)
	if result.done {
		t.Error("generic message should not set done=true")
	}
	if result.title != "Loading..." {
		t.Errorf("title should remain 'Loading...', got %q", result.title)
	}
	if cmd != nil {
		t.Error("expected nil cmd for unknown message type")
	}
}

func TestProgressModel_Update_GenericMsg(t *testing.T) {
	m := newProgressModel(testTheme(), "Processing", 10)
	// Send an unknown message type (int); should return model unchanged
	updated, cmd := m.Update(42)
	result := updated.(progressModel)
	if result.done {
		t.Error("generic message should not set done=true")
	}
	if result.current != 0 {
		t.Errorf("current should remain 0, got %d", result.current)
	}
	if cmd != nil {
		t.Error("expected nil cmd for unknown message type")
	}
}

func TestProgressModel_Update_FrameMsg(t *testing.T) {
	m := newProgressModel(testTheme(), "Processing", 10)
	// Send an empty progress.FrameMsg to cover the FrameMsg branch
	updated, _ := m.Update(progress.FrameMsg{})
	result := updated.(progressModel)
	if result.done {
		t.Error("FrameMsg should not set done=true")
	}
	if result.current != 0 {
		t.Errorf("current should remain 0 after FrameMsg, got %d", result.current)
	}
}

// --- NoColor dispatch for Start/Spinner ---

func TestProgress_NoColor_NonHeadless_Start(t *testing.T) {
	theme := NewTheme(ThemeConfig{NoColor: true})
	hm := NewHeadlessManager()
	// Do NOT force headless; NoColor should trigger headless path via ||

	var buf bytes.Buffer
	prog := newProgressImpl(theme, hm, &buf)
	pb := prog.Start("NoColor deploy", 5)
	pb.Increment(2)

	output := buf.String()
	if !strings.Contains(output, "[2/5]") {
		t.Errorf("expected '[2/5]' in NoColor output, got %q", output)
	}
	if !strings.Contains(output, "NoColor deploy") {
		t.Errorf("expected 'NoColor deploy' in output, got %q", output)
	}
}

func TestProgress_NoColor_NonHeadless_Spinner(t *testing.T) {
	theme := NewTheme(ThemeConfig{NoColor: true})
	hm := NewHeadlessManager()
	// Do NOT force headless; NoColor should trigger headless path via ||

	var buf bytes.Buffer
	prog := newProgressImpl(theme, hm, &buf)
	sp := prog.Spinner("NoColor spinner")

	output := buf.String()
	if !strings.Contains(output, "NoColor spinner") {
		t.Error("expected 'NoColor spinner' in output")
	}

	sp.SetTitle("Updated")
	output = buf.String()
	if !strings.Contains(output, "Updated") {
		t.Error("expected 'Updated' in output")
	}

	sp.Stop()
}

func TestSpinnerModel_Update_TickMsg(t *testing.T) {
	m := newSpinnerModel(testTheme(), "Loading...")
	// Send a spinner.TickMsg to exercise the tick handler
	tickCmd := m.Init()
	if tickCmd == nil {
		t.Fatal("Init should return a tick command")
	}
	// Execute the tick command to get a TickMsg
	msg := tickCmd()
	if msg == nil {
		t.Skip("tick command returned nil message")
	}
	updated, cmd := m.Update(msg)
	result := updated.(spinnerModel)
	if result.done {
		t.Error("tick message should not set done=true")
	}
	// The tick handler should return a new tick command
	_ = cmd
}

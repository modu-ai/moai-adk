package ui

import (
	"bytes"
	"strings"
	"testing"
)

// --- progressBar unit tests ---

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

// --- spinner unit tests ---

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
	// Stop should not panic and should mark as stopped
	if !sp.stopped {
		t.Error("expected stopped to be true after Stop()")
	}
}

// --- ASCII progress bar rendering ---

func TestRenderASCIIBar(t *testing.T) {
	tests := []struct {
		name     string
		current  int
		total    int
		width    int
		contains string
	}{
		{"empty", 0, 10, 20, "["},
		{"half", 5, 10, 20, "="},
		{"full", 10, 10, 20, "="},
		{"zero_total", 0, 0, 20, "["},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderASCIIBar(tt.current, tt.total, tt.width)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("expected %q to contain %q", result, tt.contains)
			}
			if !strings.HasPrefix(result, "[") || !strings.HasSuffix(result, "]") {
				t.Errorf("ASCII bar should be wrapped in brackets, got %q", result)
			}
		})
	}
}

func TestRenderASCIIBar_Proportional(t *testing.T) {
	bar := renderASCIIBar(5, 10, 20)
	// 50% of 20 = 10 filled characters
	filled := strings.Count(bar, "=")
	if filled < 8 || filled > 12 {
		t.Errorf("expected roughly 10 '=' chars for 50%%, got %d in %q", filled, bar)
	}
}

// --- Headless progress tests ---

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

// --- NoColor progress bar rendering ---

func TestProgressNoColor_ASCIIBar(t *testing.T) {
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

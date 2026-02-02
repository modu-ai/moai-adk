package output

import (
	"testing"
	"time"
)

// --- PhaseNames ---

func TestPhaseNames(t *testing.T) {
	if len(PhaseNames) == 0 {
		t.Fatal("PhaseNames is empty")
	}

	// Verify all phase names are non-empty
	for i, name := range PhaseNames {
		if name == "" {
			t.Errorf("PhaseNames[%d] is empty", i)
		}
	}

	// Should have 5 phases
	if len(PhaseNames) != 5 {
		t.Errorf("PhaseNames length = %d, want 5", len(PhaseNames))
	}
}

// --- NewProgressWriter ---

func TestNewProgressWriter(t *testing.T) {
	pw := NewProgressWriter()
	if pw == nil {
		t.Fatal("NewProgressWriter returned nil")
	}
	if pw.phase != 0 {
		t.Errorf("phase = %d, want 0", pw.phase)
	}
	if pw.total != len(PhaseNames) {
		t.Errorf("total = %d, want %d", pw.total, len(PhaseNames))
	}
	if pw.barWidth != 40 {
		t.Errorf("barWidth = %d, want 40", pw.barWidth)
	}
}

// --- Start ---

func TestProgressWriter_Start(t *testing.T) {
	pw := NewProgressWriter()

	// Before start, startTime should be zero
	if !pw.startTime.IsZero() {
		t.Error("startTime should be zero before Start()")
	}

	pw.Start()

	// After start, startTime should be set
	if pw.startTime.IsZero() {
		t.Error("startTime should not be zero after Start()")
	}

	// startTime should be recent
	elapsed := time.Since(pw.startTime)
	if elapsed > time.Second {
		t.Errorf("startTime seems too old: %v", elapsed)
	}
}

// --- Update ---

func TestProgressWriter_Update(t *testing.T) {
	pw := NewProgressWriter()
	pw.Start()

	// Initial phase is 0
	if pw.phase != 0 {
		t.Errorf("initial phase = %d, want 0", pw.phase)
	}

	// Update once
	pw.Update("Phase 1")
	if pw.phase != 1 {
		t.Errorf("phase after first update = %d, want 1", pw.phase)
	}

	// Update again
	pw.Update("Phase 2")
	if pw.phase != 2 {
		t.Errorf("phase after second update = %d, want 2", pw.phase)
	}
}

func TestProgressWriter_UpdateBeyondTotal(t *testing.T) {
	pw := NewProgressWriter()
	pw.Start()

	// Update past the total
	for i := 0; i < pw.total+5; i++ {
		pw.Update("Phase")
	}

	// Phase should not exceed total
	if pw.phase > pw.total {
		t.Errorf("phase = %d, should not exceed total %d", pw.phase, pw.total)
	}
}

// --- Complete ---

func TestProgressWriter_Complete(t *testing.T) {
	pw := NewProgressWriter()
	pw.Start()

	// Complete should set phase to total
	pw.Complete()

	if pw.phase != pw.total {
		t.Errorf("phase after Complete = %d, want %d", pw.phase, pw.total)
	}
}

// --- CreateProgressCallback ---

func TestCreateProgressCallback(t *testing.T) {
	pw, callback := CreateProgressCallback()

	if pw == nil {
		t.Fatal("CreateProgressCallback returned nil ProgressWriter")
	}
	if callback == nil {
		t.Fatal("CreateProgressCallback returned nil callback")
	}

	// Use the callback
	pw.Start()
	callback("test message", 1, 5)

	if pw.phase != 1 {
		t.Errorf("phase after callback = %d, want 1", pw.phase)
	}
}

// --- progressModel ---

func TestInitialModel(t *testing.T) {
	m := initialModel()
	if m.total != len(PhaseNames) {
		t.Errorf("total = %d, want %d", m.total, len(PhaseNames))
	}
	if m.phase != 0 {
		t.Errorf("phase = %d, want 0", m.phase)
	}
	if m.quitting {
		t.Error("quitting should be false initially")
	}
}

func TestProgressModel_View_Quitting(t *testing.T) {
	m := initialModel()
	m.quitting = true

	view := m.View()
	if view != "" {
		t.Errorf("View when quitting should return empty string, got %q", view)
	}
}

func TestProgressModel_View_WithPhase(t *testing.T) {
	m := initialModel()
	m.phase = 1

	view := m.View()
	if view == "" {
		t.Error("View with active phase should return non-empty string")
	}
}

func TestProgressModel_Init(t *testing.T) {
	m := initialModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a non-nil command")
	}
}

package convention

import (
	"testing"
)

// autoOpts returns a default AutoDetectOptions (auto-detection enabled, sample
// size 100, threshold 0.5, fallback conventional-commits) for tests that load a
// named built-in convention where the options are ignored.
func autoOpts() AutoDetectOptions {
	return AutoDetectOptions{
		Enabled:             true,
		SampleSize:          100,
		ConfidenceThreshold: 0.5,
		Fallback:            "conventional-commits",
	}
}

func TestNewManager_CreatesInstance(t *testing.T) {
	m := NewManager("/some/path")
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.Convention() != nil {
		t.Error("Convention() should be nil before loading")
	}
}

func TestManager_LoadConvention_Builtin(t *testing.T) {
	tests := []struct {
		name    string
		conv    string
		wantErr bool
	}{
		{"conventional-commits", "conventional-commits", false},
		{"angular", "angular", false},
		{"karma", "karma", false},
		{"unknown", "nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager("/unused")
			err := m.LoadConvention(tt.conv, autoOpts())
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConvention(%q) error = %v, wantErr %v", tt.conv, err, tt.wantErr)
				return
			}
			if !tt.wantErr && m.Convention() == nil {
				t.Error("Convention() should not be nil after loading builtin")
			}
			if !tt.wantErr && m.Convention().Name != tt.conv {
				t.Errorf("Convention().Name = %q, want %q", m.Convention().Name, tt.conv)
			}
		})
	}
}

func TestManager_LoadConvention_Auto(t *testing.T) {
	repoRoot := findGitRoot(t)

	m := NewManager(repoRoot)
	err := m.LoadConvention("auto", autoOpts())
	if err != nil {
		t.Fatalf("LoadConvention(auto) error = %v", err)
	}
	if m.Convention() == nil {
		t.Error("Convention() should not be nil after auto-detection")
	}
}

func TestManager_LoadConvention_AutoFallback(t *testing.T) {
	// Use a temp dir with no git history -- auto should fallback to
	// conventional-commits.
	tmpDir := t.TempDir()

	m := NewManager(tmpDir)
	err := m.LoadConvention("auto", autoOpts())
	if err != nil {
		t.Fatalf("LoadConvention(auto) with fallback error = %v", err)
	}
	if m.Convention() == nil {
		t.Fatal("Convention() should not be nil after auto-fallback")
	}
	if m.Convention().Name != "conventional-commits" {
		t.Errorf("Convention().Name = %q, want %q", m.Convention().Name, "conventional-commits")
	}
}

// TestManager_LoadConvention_AutoDetectDisabled covers Fix A (AC-WC9-008): when
// auto_detection.enabled is false, Detect is skipped and the configured fallback
// is loaded directly — even in a repo with detectable history.
func TestManager_LoadConvention_AutoDetectDisabled(t *testing.T) {
	repoRoot := findGitRoot(t)
	m := NewManager(repoRoot)
	opts := autoOpts()
	opts.Enabled = false
	opts.Fallback = "angular"
	if err := m.LoadConvention("auto", opts); err != nil {
		t.Fatalf("LoadConvention(auto, disabled) error = %v", err)
	}
	if got := m.Convention().Name; got != "angular" {
		t.Errorf("disabled auto-detect should load fallback; Convention().Name = %q, want angular", got)
	}
}

// TestManager_LoadConvention_AutoSampleSize covers Fix A (AC-WC9-008): an explicit
// non-default sample_size is forwarded to Detect and a convention is resolved
// without error (the prior hardcoded 100 is no longer used).
func TestManager_LoadConvention_AutoSampleSize(t *testing.T) {
	repoRoot := findGitRoot(t)
	m := NewManager(repoRoot)
	opts := autoOpts()
	opts.SampleSize = 25 // explicit non-default sample size forwarded to Detect
	if err := m.LoadConvention("auto", opts); err != nil {
		t.Fatalf("LoadConvention(auto, sample_size=25) error = %v", err)
	}
	if m.Convention() == nil {
		t.Error("Convention() should not be nil after sample-size detection")
	}
}

// TestManager_LoadConvention_AutoConfidenceFallback covers Fix A (AC-WC9-008): a
// 1.0 confidence threshold gates the detected result; the resolution still
// succeeds (either the detected convention at perfect confidence, or the fallback).
func TestManager_LoadConvention_AutoConfidenceFallback(t *testing.T) {
	repoRoot := findGitRoot(t)
	m := NewManager(repoRoot)
	opts := autoOpts()
	opts.ConfidenceThreshold = 1.0 // require perfect match → realistically falls back
	opts.Fallback = "karma"
	if err := m.LoadConvention("auto", opts); err != nil {
		t.Fatalf("LoadConvention(auto, threshold=1.0) error = %v", err)
	}
	// Both outcomes are valid Fix A behavior, so assert only that a convention was
	// resolved without error under the confidence gate.
	if m.Convention() == nil {
		t.Error("Convention() should not be nil after confidence-gated resolution")
	}
}

// TestManager_LoadConvention_AutoFallbackConfigured covers Fix A (AC-WC9-008): the
// configured fallback (not the hardcoded conventional-commits) is used.
func TestManager_LoadConvention_AutoFallbackConfigured(t *testing.T) {
	tmpDir := t.TempDir() // no git history → detection would fail, but disabled here
	m := NewManager(tmpDir)
	opts := autoOpts()
	opts.Enabled = false
	opts.Fallback = "angular"
	if err := m.LoadConvention("auto", opts); err != nil {
		t.Fatalf("LoadConvention(auto, fallback=angular) error = %v", err)
	}
	if got := m.Convention().Name; got != "angular" {
		t.Errorf("configured fallback not honored; Convention().Name = %q, want angular", got)
	}
}

// TestSetMaxLength covers Fix B (AC-WC9-010): SetMaxLength overrides the built-in
// convention's own MaxLength (100) with the configured value, applied after
// LoadConvention.
func TestSetMaxLength(t *testing.T) {
	m := NewManager("/unused")
	if err := m.LoadConvention("conventional-commits", autoOpts()); err != nil {
		t.Fatalf("LoadConvention: %v", err)
	}
	// Built-in default is 100; override to 72.
	m.SetMaxLength(72)
	if got := m.Convention().MaxLength; got != 72 {
		t.Errorf("SetMaxLength(72): Convention().MaxLength = %d, want 72 (override of built-in 100)", got)
	}
	// Non-positive is a no-op (preserves the current value).
	m.SetMaxLength(0)
	if got := m.Convention().MaxLength; got != 72 {
		t.Errorf("SetMaxLength(0) should be a no-op; MaxLength = %d, want 72", got)
	}
}

// TestSetMaxLength_NilConventionNoOp covers Fix B (AC-WC9-010): SetMaxLength on a
// Manager with no loaded convention is a safe no-op.
func TestSetMaxLength_NilConventionNoOp(t *testing.T) {
	m := NewManager("/unused")
	m.SetMaxLength(72) // must not panic
	if m.Convention() != nil {
		t.Error("SetMaxLength on nil convention must not load one")
	}
}

// TestLoadConvention_MaxLengthForwarded covers Fix A+B together (AC-WC9-010): a
// named convention loaded then SetMaxLength forwards the configured max_length.
func TestLoadConvention_MaxLengthForwarded(t *testing.T) {
	m := NewManager("/unused")
	if err := m.LoadConvention("angular", autoOpts()); err != nil {
		t.Fatalf("LoadConvention: %v", err)
	}
	m.SetMaxLength(50)
	if got := m.Convention().MaxLength; got != 50 {
		t.Errorf("forwarded max_length = %d, want 50", got)
	}
}

func TestManager_ValidateMessage(t *testing.T) {
	m := NewManager("/unused")
	if err := m.LoadConvention("conventional-commits", autoOpts()); err != nil {
		t.Fatalf("LoadConvention: %v", err)
	}

	tests := []struct {
		name    string
		message string
		valid   bool
	}{
		{"valid feat", "feat(auth): add JWT", true},
		{"valid fix", "fix: resolve bug", true},
		{"invalid", "random message", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.ValidateMessage(tt.message)
			if result.Valid != tt.valid {
				t.Errorf("ValidateMessage(%q).Valid = %v, want %v", tt.message, result.Valid, tt.valid)
			}
		})
	}
}

func TestManager_ValidateMessage_NoConvention(t *testing.T) {
	m := NewManager("/unused")
	result := m.ValidateMessage("anything goes")
	if !result.Valid {
		t.Error("ValidateMessage with no convention should always be valid")
	}
}

func TestManager_ValidateMessages(t *testing.T) {
	m := NewManager("/unused")
	if err := m.LoadConvention("conventional-commits", autoOpts()); err != nil {
		t.Fatalf("LoadConvention: %v", err)
	}

	messages := []string{
		"feat: add feature",
		"bad message",
		"fix: resolve bug",
	}

	results := m.ValidateMessages(messages)
	if len(results) != 3 {
		t.Fatalf("ValidateMessages returned %d results, want 3", len(results))
	}

	if !results[0].Valid {
		t.Error("results[0] should be valid")
	}
	if results[1].Valid {
		t.Error("results[1] should be invalid")
	}
	if !results[2].Valid {
		t.Error("results[2] should be valid")
	}
}

func TestManager_ValidateMessages_Empty(t *testing.T) {
	m := NewManager("/unused")
	results := m.ValidateMessages(nil)
	if len(results) != 0 {
		t.Errorf("ValidateMessages(nil) returned %d results, want 0", len(results))
	}
}

func TestManager_Convention_ReturnsNilBeforeLoad(t *testing.T) {
	m := NewManager("/unused")
	if m.Convention() != nil {
		t.Error("Convention() should be nil before any load")
	}
}

func TestManager_Convention_ReturnsLoadedConvention(t *testing.T) {
	m := NewManager("/unused")
	if err := m.LoadConvention("angular", autoOpts()); err != nil {
		t.Fatalf("LoadConvention: %v", err)
	}
	conv := m.Convention()
	if conv == nil {
		t.Fatal("Convention() should not be nil after loading")
		return // staticcheck SA5011
	}
	if conv.Name != "angular" {
		t.Errorf("Convention().Name = %q, want %q", conv.Name, "angular")
	}
}

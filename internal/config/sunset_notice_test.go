package config

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// bufferHandler is a minimal slog.Handler that captures log records.
type bufferHandler struct {
	buf *bytes.Buffer
}

func newBufferHandler() (*bufferHandler, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return &bufferHandler{buf: buf}, buf
}

func (h *bufferHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *bufferHandler) Handle(_ context.Context, r slog.Record) error {
	h.buf.WriteString(r.Message)
	h.buf.WriteByte('\n')
	return nil
}

func (h *bufferHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *bufferHandler) WithGroup(_ string) slog.Handler      { return h }

// TestSunsetNotice_FiresOnce verifies that SUNSET_CONFIG_DORMANT_NOTICE is emitted
// exactly once even when Loader.Load() is called twice with a sunset.yaml present.
// REQ-MIG003-018, AC-MIG003-15.
func TestSunsetNotice_FiresOnce(t *testing.T) {
	// Reset the sync.Once guard so this test runs independently.
	resetSunsetNoticeOnce()

	// Set up a sections dir that contains sunset.yaml.
	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, "moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	sunsetPath := filepath.Join(sectionsDir, "sunset.yaml")
	if err := os.WriteFile(sunsetPath, []byte("sunset:\n  enabled: true\n"), 0o644); err != nil {
		t.Fatalf("WriteFile sunset.yaml: %v", err)
	}

	// Install a capturing logger.
	handler, buf := newBufferHandler()
	origLogger := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(origLogger)

	moaiDir := filepath.Join(tmpDir, "moai")

	// First Loader.Load() call.
	loader1 := NewLoader()
	resetSunsetNoticeOnce() // ensure the once is fresh before first call
	_, _ = loader1.Load(moaiDir)

	// Second Loader.Load() call — notice must NOT fire again.
	loader2 := NewLoader()
	_, _ = loader2.Load(moaiDir)

	// Count occurrences.
	output := buf.String()
	count := strings.Count(output, "SUNSET_CONFIG_DORMANT_NOTICE")
	if count != 1 {
		t.Errorf("SUNSET_CONFIG_DORMANT_NOTICE: expected exactly 1 occurrence, got %d\nlog output:\n%s",
			count, output)
	}
}

// TestSunsetNotice_AbsentFileNoNotice verifies that when sunset.yaml is absent,
// no SUNSET_CONFIG_DORMANT_NOTICE is emitted.
// REQ-MIG003-018, AC-MIG003-15.
func TestSunsetNotice_AbsentFileNoNotice(t *testing.T) {
	// Reset the sync.Once guard.
	resetSunsetNoticeOnce()

	// Set up a sections dir WITHOUT sunset.yaml.
	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, "moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Install a capturing logger.
	handler, buf := newBufferHandler()
	origLogger := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(origLogger)

	moaiDir := filepath.Join(tmpDir, "moai")
	loader := NewLoader()
	_, _ = loader.Load(moaiDir)

	output := buf.String()
	if strings.Contains(output, "SUNSET_CONFIG_DORMANT_NOTICE") {
		t.Errorf("expected no SUNSET_CONFIG_DORMANT_NOTICE when sunset.yaml is absent, got:\n%s", output)
	}
}

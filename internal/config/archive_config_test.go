package config

import (
	"os"
	"path/filepath"
	"testing"
)

// archive_config_test.go — SPEC-SESSIONSTART-PERF-001 M2 (REQ-SSP-012 / REQ-SSP-018).
//
// The archive grace window is a tunable threshold, so it lives in config — never as
// an inline literal in business logic. The compiled default is the fallback when
// archive.yaml is absent.

func TestDefaultArchiveGraceDays(t *testing.T) {
	t.Parallel()

	if DefaultArchiveGraceDays != 90 {
		t.Errorf("DefaultArchiveGraceDays = %d, want 90", DefaultArchiveGraceDays)
	}
}

func TestNewDefaultArchiveConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultArchiveConfig()
	if cfg.GraceDays != DefaultArchiveGraceDays {
		t.Errorf("GraceDays = %d, want %d", cfg.GraceDays, DefaultArchiveGraceDays)
	}
}

// TestArchiveGraceDays_AbsentFileFallsBackToDefault covers the "no archive.yaml"
// path: the accessor must still resolve to the compiled default.
func TestArchiveGraceDays_AbsentFileFallsBackToDefault(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	if got := cfg.ArchiveGraceDays(); got != DefaultArchiveGraceDays {
		t.Errorf("ArchiveGraceDays() = %d, want %d", got, DefaultArchiveGraceDays)
	}
}

// TestArchiveGraceDays_ZeroFallsBackToDefault covers EC: an archive.yaml that
// declares the section but omits (or zeroes) grace_days must not degenerate into
// "no grace at all".
func TestArchiveGraceDays_ZeroFallsBackToDefault(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	cfg.Archive.GraceDays = 0

	if got := cfg.ArchiveGraceDays(); got != DefaultArchiveGraceDays {
		t.Errorf("ArchiveGraceDays() with zero value = %d, want the default %d", got, DefaultArchiveGraceDays)
	}
}

func TestLoadArchiveSection(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	yaml := `archive:
  grace_days: 45
`
	if err := os.WriteFile(filepath.Join(dir, "archive.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatalf("write archive.yaml: %v", err)
	}

	cfg := NewDefaultConfig()
	l := &Loader{loadedSections: map[string]bool{}}
	l.loadArchiveSection(dir, cfg)

	if cfg.Archive.GraceDays != 45 {
		t.Errorf("GraceDays = %d, want 45 from archive.yaml", cfg.Archive.GraceDays)
	}
	if got := cfg.ArchiveGraceDays(); got != 45 {
		t.Errorf("ArchiveGraceDays() = %d, want 45", got)
	}
}

package lifecycle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestActivatePersistentMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		workflow   string
		specID     string
		maxMinutes int
	}{
		{
			name:       "basic activation",
			workflow:   "run",
			specID:     "SPEC-001",
			maxMinutes: 60,
		},
		{
			name:       "no time limit",
			workflow:   "loop",
			specID:     "SPEC-002",
			maxMinutes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			projectDir := t.TempDir()
			if err := ActivatePersistentMode(projectDir, tt.workflow, tt.specID, tt.maxMinutes); err != nil {
				t.Fatalf("ActivatePersistentMode() error = %v", err)
			}

			// Verify file was created with correct content
			filePath := filepath.Join(projectDir, persistentModeFile)
			data, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("file not created: %v", err)
			}

			var mode PersistentMode
			if err := json.Unmarshal(data, &mode); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}

			if !mode.Active {
				t.Error("Active should be true after activation")
			}
			if mode.Workflow != tt.workflow {
				t.Errorf("Workflow = %q, want %q", mode.Workflow, tt.workflow)
			}
			if mode.SpecID != tt.specID {
				t.Errorf("SpecID = %q, want %q", mode.SpecID, tt.specID)
			}
			if mode.MaxDurationMinutes != tt.maxMinutes {
				t.Errorf("MaxDurationMinutes = %d, want %d", mode.MaxDurationMinutes, tt.maxMinutes)
			}
			if mode.StartedAt.IsZero() {
				t.Error("StartedAt should not be zero")
			}
		})
	}
}

func TestActivatePersistentMode_CreatesParentDirectories(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	// The .moai/state/ directory does not exist yet — ActivatePersistentMode must create it.
	if err := ActivatePersistentMode(projectDir, "run", "SPEC-001", 60); err != nil {
		t.Fatalf("ActivatePersistentMode() error = %v", err)
	}

	filePath := filepath.Join(projectDir, persistentModeFile)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("file should exist after activation")
	}
}

func TestDeactivatePersistentMode(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	// First activate
	if err := ActivatePersistentMode(projectDir, "run", "SPEC-001", 60); err != nil {
		t.Fatalf("ActivatePersistentMode() error = %v", err)
	}

	// Then deactivate
	if err := DeactivatePersistentMode(projectDir); err != nil {
		t.Fatalf("DeactivatePersistentMode() error = %v", err)
	}

	// Verify active=false
	mode, err := CheckPersistentMode(projectDir)
	if err != nil {
		t.Fatalf("CheckPersistentMode() error = %v", err)
	}
	if mode == nil {
		t.Fatal("mode should not be nil after deactivation")
	}
	if mode.Active {
		t.Error("Active should be false after deactivation")
	}
}

func TestDeactivatePersistentMode_MissingFile(t *testing.T) {
	t.Parallel()

	// Should not error when file does not exist
	projectDir := t.TempDir()
	if err := DeactivatePersistentMode(projectDir); err != nil {
		t.Errorf("DeactivatePersistentMode() should not error for missing file, got: %v", err)
	}
}

func TestCheckPersistentMode_MissingFile(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	mode, err := CheckPersistentMode(projectDir)
	if err != nil {
		t.Fatalf("CheckPersistentMode() error = %v", err)
	}
	if mode != nil {
		t.Error("mode should be nil when file does not exist")
	}
}

func TestCheckPersistentMode_ReturnsCorrectData(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	if err := ActivatePersistentMode(projectDir, "ddd", "SPEC-099", 120); err != nil {
		t.Fatalf("ActivatePersistentMode() error = %v", err)
	}

	mode, err := CheckPersistentMode(projectDir)
	if err != nil {
		t.Fatalf("CheckPersistentMode() error = %v", err)
	}
	if mode == nil {
		t.Fatal("mode should not be nil")
	}

	if !mode.Active {
		t.Error("Active should be true")
	}
	if mode.Workflow != "ddd" {
		t.Errorf("Workflow = %q, want %q", mode.Workflow, "ddd")
	}
	if mode.SpecID != "SPEC-099" {
		t.Errorf("SpecID = %q, want %q", mode.SpecID, "SPEC-099")
	}
	if mode.MaxDurationMinutes != 120 {
		t.Errorf("MaxDurationMinutes = %d, want 120", mode.MaxDurationMinutes)
	}
}

func TestIsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		maxMinutes     int
		startedAgo     time.Duration
		wantExpired    bool
	}{
		{
			name:        "no limit, never expires",
			maxMinutes:  0,
			startedAgo:  24 * time.Hour,
			wantExpired: false,
		},
		{
			name:        "within duration",
			maxMinutes:  60,
			startedAgo:  30 * time.Minute,
			wantExpired: false,
		},
		{
			name:        "exactly at limit (past)",
			maxMinutes:  60,
			startedAgo:  61 * time.Minute,
			wantExpired: true,
		},
		{
			name:        "2 hours ago, max 60 minutes",
			maxMinutes:  60,
			startedAgo:  2 * time.Hour,
			wantExpired: true,
		},
		{
			name:        "just started",
			maxMinutes:  60,
			startedAgo:  1 * time.Second,
			wantExpired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mode := &PersistentMode{
				Active:             true,
				MaxDurationMinutes: tt.maxMinutes,
				StartedAt:          time.Now().Add(-tt.startedAgo),
			}

			if got := mode.IsExpired(); got != tt.wantExpired {
				t.Errorf("IsExpired() = %v, want %v", got, tt.wantExpired)
			}
		})
	}
}

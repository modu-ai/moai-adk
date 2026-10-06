package config

// deploy_mode_test.go — SPEC-INIT-SHRINK-001 REQ-009 (OD-5 settled (a)):
// the deployment_mode reader beside ReadHarness, same llm.yaml section file,
// same loadYAMLFile seam.

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSectionsLLM(t *testing.T, dir, content string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	path := filepath.Join(dir, "llm.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}
	return path
}

// TestReadDeployModeValues reads both closed-set values from a section file
// that also carries the harness key (the real file shape).
func TestReadDeployModeValues(t *testing.T) {
	for _, tc := range []struct{ value, want string }{
		{"plugin", "plugin"},
		{"local", "local"},
	} {
		t.Run(tc.value, func(t *testing.T) {
			dir := t.TempDir()
			writeSectionsLLM(t, dir, "llm:\n  harness: claude\n  deployment_mode: "+tc.value+"\n")
			if got := ReadDeployModeFrom(dir); got != tc.want {
				t.Fatalf("ReadDeployModeFrom = %q, want %q", got, tc.want)
			}
			root := t.TempDir()
			if err := os.MkdirAll(filepath.Join(root, ".moai", "config", "sections"), 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			if err := os.WriteFile(filepath.Join(root, ".moai", "config", "sections", "llm.yaml"),
				[]byte("llm:\n  harness: claude\n  deployment_mode: "+tc.value+"\n"), 0o644); err != nil {
				t.Fatalf("write: %v", err)
			}
			if got := ReadDeployMode(root); got != tc.want {
				t.Fatalf("ReadDeployMode = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestReadDeployModeAbsentOrInvalidIsRecordless pins the record-less
// contract: an absent file, an absent key, and an out-of-set value all read
// as "" so the update path routes to the migration (REQ-015), never to an
// invented mode.
func TestReadDeployModeAbsentOrInvalidIsRecordless(t *testing.T) {
	t.Run("absent file", func(t *testing.T) {
		if got := ReadDeployModeFrom(t.TempDir()); got != "" {
			t.Fatalf("absent file = %q, want \"\"", got)
		}
	})
	t.Run("absent key", func(t *testing.T) {
		dir := t.TempDir()
		writeSectionsLLM(t, dir, "llm:\n  harness: claude\n")
		if got := ReadDeployModeFrom(dir); got != "" {
			t.Fatalf("absent key = %q, want \"\"", got)
		}
	})
	t.Run("invalid value", func(t *testing.T) {
		dir := t.TempDir()
		writeSectionsLLM(t, dir, "llm:\n  harness: claude\n  deployment_mode: banana\n")
		if got := ReadDeployModeFrom(dir); got != "" {
			t.Fatalf("invalid value = %q, want \"\"", got)
		}
	})
}

// TestIsValidDeployMode pins the closed set.
func TestIsValidDeployMode(t *testing.T) {
	for _, v := range []string{"plugin", "local"} {
		if !IsValidDeployMode(v) {
			t.Errorf("IsValidDeployMode(%q) = false, want true", v)
		}
	}
	for _, v := range []string{"", "banana", "PLUGIN", "Plugin"} {
		if IsValidDeployMode(v) {
			t.Errorf("IsValidDeployMode(%q) = true, want false", v)
		}
	}
}

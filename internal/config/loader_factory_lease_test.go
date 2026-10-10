package config

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// SPEC-FACTORY-LEASE-RENEW-001 (card t1663), milestone M1. These loader and default
// tests cover both keys, workflow.factory_lease.duration and
// workflow.factory_lease.renew_interval. The workflow-section symmetry case is
// withheld by ruling d-20261010T104937Z-2161 (option c), so these tests and the
// anti-rot inventory entries are the checks of the keys.

// writeFactoryLeaseWorkflow writes body to <root>/.moai/config/sections/workflow.yaml.
// An empty body writes no file.
func writeFactoryLeaseWorkflow(t *testing.T, root, body string) {
	t.Helper()
	if body == "" {
		return
	}
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir config sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
}

func TestFactoryLeaseDefaultsAreSeeded(t *testing.T) {
	if DefaultFactoryLeaseDuration != 15*time.Minute {
		t.Fatalf("DefaultFactoryLeaseDuration = %s, want 15m0s", DefaultFactoryLeaseDuration)
	}
	if DefaultFactoryLeaseRenewInterval != 5*time.Minute {
		t.Fatalf("DefaultFactoryLeaseRenewInterval = %s, want 5m0s", DefaultFactoryLeaseRenewInterval)
	}
	if DefaultFactoryLeaseRenewInterval > DefaultFactoryLeaseDuration/2 {
		t.Fatalf("default renew interval %s exceeds half the default duration %s",
			DefaultFactoryLeaseRenewInterval, DefaultFactoryLeaseDuration)
	}
	got := NewDefaultWorkflowConfig().FactoryLease
	want := FactoryLeaseConfig{Duration: "15m0s", RenewInterval: "5m0s"}
	if got != want {
		t.Fatalf("NewDefaultWorkflowConfig().FactoryLease = %+v, want %+v", got, want)
	}
}

func TestLoadFactoryLeaseReadsBothKeys(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	root := t.TempDir()
	writeFactoryLeaseWorkflow(t, root, "workflow:\n  factory_lease:\n    duration: 20m\n    renew_interval: 4m\n")
	got, err := LoadFactoryLease(root)
	if err != nil {
		t.Fatalf("LoadFactoryLease: %v", err)
	}
	want := FactoryLeaseConfig{Duration: "20m", RenewInterval: "4m"}
	if got != want {
		t.Fatalf("LoadFactoryLease = %+v, want %+v", got, want)
	}
}

func TestLoadFactoryLeaseDefaultsForOmittedKeys(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	cases := []struct {
		name string
		body string
		want FactoryLeaseConfig
	}{
		{"duration-only", "workflow:\n  factory_lease:\n    duration: 20m\n", FactoryLeaseConfig{Duration: "20m", RenewInterval: "5m0s"}},
		{"renew-only", "workflow:\n  factory_lease:\n    renew_interval: 3m\n", FactoryLeaseConfig{Duration: "15m0s", RenewInterval: "3m"}},
		{"no-factory-lease-block", "workflow:\n  execution_mode: auto\n", FactoryLeaseConfig{Duration: "15m0s", RenewInterval: "5m0s"}},
		{"no-file", "", FactoryLeaseConfig{Duration: "15m0s", RenewInterval: "5m0s"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			writeFactoryLeaseWorkflow(t, root, c.body)
			got, err := LoadFactoryLease(root)
			if err != nil {
				t.Fatalf("LoadFactoryLease: %v", err)
			}
			if got != c.want {
				t.Fatalf("LoadFactoryLease = %+v, want %+v", got, c.want)
			}
		})
	}
}

// Ruling d-20261010T111725Z-3ef9 (part a): a workflow.yaml that cannot be decoded is not
// a refusal. The keys take their defaults, and a slog warning names the file and the
// error, which is the config fallback path the other loaders use.
func TestLoadFactoryLeaseMalformedFileFallsBackWithWarning(t *testing.T) {
	t.Setenv(EnvConfigDir, "")
	root := t.TempDir()
	writeFactoryLeaseWorkflow(t, root, "workflow:\n  factory_lease: [unclosed\n")
	var logs bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	got, err := LoadFactoryLease(root)
	if err != nil {
		t.Fatalf("LoadFactoryLease on a malformed file: err = %v, want the defaults and a warning", err)
	}
	if want := NewDefaultWorkflowConfig().FactoryLease; got != want {
		t.Fatalf("LoadFactoryLease on a malformed file = %+v, want the defaults %+v", got, want)
	}
	needle := filepath.Join(".moai", "config", "sections", "workflow.yaml")
	if !strings.Contains(logs.String(), needle) || !strings.Contains(logs.String(), ErrInvalidYAML.Error()) {
		t.Fatalf("warning does not name the file %s and the error %q:\n%s", needle, ErrInvalidYAML, logs.String())
	}
}

// TestLoadFactoryLeaseIgnoresEnvConfigDir records the behaviour of the entry point:
// it reads <projectRoot>/.moai/config/sections and does not consult MOAI_CONFIG_DIR.
func TestLoadFactoryLeaseIgnoresEnvConfigDir(t *testing.T) {
	root := t.TempDir()
	other := t.TempDir()
	writeFactoryLeaseWorkflow(t, root, "workflow:\n  factory_lease:\n    duration: 20m\n")
	writeFactoryLeaseWorkflow(t, filepath.Join(other, ".moai"), "workflow:\n  factory_lease:\n    duration: 7m\n")
	t.Setenv(EnvConfigDir, filepath.Join(other, ".moai"))
	got, err := LoadFactoryLease(root)
	if err != nil {
		t.Fatalf("LoadFactoryLease: %v", err)
	}
	if got.Duration != "20m" {
		t.Fatalf("LoadFactoryLease honoured MOAI_CONFIG_DIR: duration = %q, want the project's %q", got.Duration, "20m")
	}
}

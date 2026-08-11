package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestPrepareFactorySettingsWritesTransientFile is AC-FB-010. When the operator
// did NOT pass --settings, the launcher writes a session-private file to
// os.TempDir() containing {"crossSessionInbound": "accept"} and returns the
// --settings flag pair pointing at it.
func TestPrepareFactorySettingsWritesTransientFile(t *testing.T) {
	for _, key := range []string{config.EnvMoaiFactorySettingsInjected} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}

	flag, cleanup := prepareFactorySettings([]string{"-p", "dev"})
	t.Cleanup(cleanup)

	if len(flag) != 2 || flag[0] != "--settings" {
		t.Fatalf("expected [\"--settings\", <path>], got %v", flag)
	}
	path := flag[1]
	if !filepath.IsAbs(path) {
		t.Errorf("settings path is not absolute: %q", path)
	}
	dir := filepath.Clean(filepath.Dir(path))
	wantDir := filepath.Clean(os.TempDir())
	if dir != wantDir {
		t.Errorf("settings file in %q, want os.TempDir() = %q", dir, wantDir)
	}
	base := filepath.Base(path)
	if !strings.HasPrefix(base, "moai-factory-") || !strings.HasSuffix(base, ".json") {
		t.Errorf("settings filename %q does not match moai-factory-*.json", base)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read the transient settings file: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("transient settings file is not valid JSON: %v", err)
	}
	if got["crossSessionInbound"] != "accept" {
		t.Errorf("crossSessionInbound = %q, want \"accept\"", got["crossSessionInbound"])
	}

	// The injected signal env var MUST be set so the hook knows auto-accept is active.
	if os.Getenv(config.EnvMoaiFactorySettingsInjected) != "1" {
		t.Errorf("%s not set after injection", config.EnvMoaiFactorySettingsInjected)
	}

	// Cleanup removes the file and restores the env var.
	cleanup()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("settings file survived cleanup: %v", err)
	}
	if _, present := os.LookupEnv(config.EnvMoaiFactorySettingsInjected); present {
		t.Errorf("%s still present after cleanup", config.EnvMoaiFactorySettingsInjected)
	}
}

// TestPrepareFactorySettingsHonorsOperatorSupplied is AC-FB-011. When the
// operator passed --settings <file> on the command line, the launcher SHALL NOT
// inject its own — no transient file is created, no flag is returned, and the
// injected env var stays unset (the hook will print the verify advisory).
func TestPrepareFactorySettingsHonorsOperatorSupplied(t *testing.T) {
	for _, key := range []string{config.EnvMoaiFactorySettingsInjected} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}

	// Long form.
	flag, cleanup := prepareFactorySettings([]string{"--settings", "/tmp/operator.json"})
	t.Cleanup(cleanup)
	if len(flag) != 0 {
		t.Errorf("long form: expected no injection, got %v", flag)
	}
	if os.Getenv(config.EnvMoaiFactorySettingsInjected) == "1" {
		t.Errorf("long form: injected env var set despite operator --settings")
	}

	// Equals form.
	flag2, cleanup2 := prepareFactorySettings([]string{"--settings=/tmp/op2.json"})
	t.Cleanup(cleanup2)
	if len(flag2) != 0 {
		t.Errorf("equals form: expected no injection, got %v", flag2)
	}

	// --settings before the pass-through marker is honored.
	flag3, cleanup3 := prepareFactorySettings([]string{"--settings", "/tmp/op3.json", "--", "--settings", "/tmp/decoy.json"})
	t.Cleanup(cleanup3)
	if len(flag3) != 0 {
		t.Errorf("pre-marker form: expected no injection, got %v", flag3)
	}

	// --settings AFTER the pass-through marker is NOT an operator supply (it is
	// a passthrough arg to claude, not a moai-level flag). The launcher DOES
	// inject here — the operator's intent to supply --settings to moai's
	// launcher is expressed before the marker only.
	flag4, cleanup4 := prepareFactorySettings([]string{"--", "--settings", "/tmp/post.json"})
	t.Cleanup(cleanup4)
	if len(flag4) != 2 {
		t.Errorf("post-marker form: expected injection (operator --settings is passthrough), got %v", flag4)
	}
}

// TestPrepareFactorySettingsFailsOpenOnWriteError is EC-4. When the transient
// file write fails, the launcher degrades to launching without the injected
// --settings (never blocks the launch). The flag is empty and cleanup is safe.
func TestPrepareFactorySettingsFailsOpenOnWriteError(t *testing.T) {
	// Drive TempDir to an unwritable location by setting TMPDIR to a path under
	// a read-only directory we create.
	roDir := t.TempDir()
	unwritable := filepath.Join(roDir, "no-write")
	if err := os.Mkdir(unwritable, 0o444); err != nil {
		t.Skipf("could not create read-only dir: %v", err)
	}
	t.Setenv("TMPDIR", unwritable)

	flag, cleanup := prepareFactorySettings([]string{"-p", "dev"})
	t.Cleanup(cleanup)

	if len(flag) != 0 {
		t.Errorf("fail-open: expected no flag on write failure, got %v", flag)
	}
	// cleanup is a no-op; calling it must not panic.
	cleanup()
}

// TestOperatorSuppliedSettingsDetectsAllForms covers the detection helper
// across flag shapes and the pass-through marker boundary.
func TestOperatorSuppliedSettingsDetectsAllForms(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{"long form", []string{"--settings", "/tmp/a.json"}, true},
		{"equals form", []string{"--settings=/tmp/a.json"}, true},
		{"among other flags", []string{"-p", "dev", "--settings", "/tmp/a.json", "--verbose"}, true},
		{"absent", []string{"-p", "dev"}, false},
		{"after marker (passthrough)", []string{"--", "--settings", "/tmp/a.json"}, false},
		{"empty args", nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := operatorSuppliedSettings(c.args)
			if got != c.want {
				t.Errorf("operatorSuppliedSettings(%v) = %v, want %v", c.args, got, c.want)
			}
		})
	}
}

package cli

// resolution_gate_test.go — AC-008 (b) and AC-020 (a) M1 half
// (SPEC-INIT-SHRINK-001): the resolution-gate harness shape and the
// hermeticity guards. The full measurement (scripts/check-bare-name-resolution.sh
// without --self-check) drives the real tool runtimes and is NEVER run from
// a Go test; only the self-check mode (no real runtime) is exercised here.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// resolutionGateScriptDir resolves the repo-root scripts/ directory from
// this package's directory (internal/cli → ../../scripts).
func resolutionGateScriptDir(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", "scripts"))
	if err != nil {
		t.Fatalf("resolve scripts dir: %v", err)
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		t.Fatalf("scripts directory missing at %s", root)
	}
	return root
}

// TestResolutionGateHarness is AC-008 (b): the harness's self-check mode
// runs its isolation cases, prints PASS lines and the RESULT verdict line,
// and exits 0 with fail=0. The negative control inside proves the scrub is
// load-bearing (a planted unscrubbed variable FAILs its case).
func TestResolutionGateHarness(t *testing.T) {
	scriptDir := resolutionGateScriptDir(t)
	measure := filepath.Join(scriptDir, "check-bare-name-resolution.sh")
	selfTest := filepath.Join(scriptDir, "test-bare-name-resolution.sh")
	for _, p := range []string{measure, selfTest} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("harness script missing: %s", p)
		}
	}

	cmd := exec.Command("sh", selfTest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("self-test failed: %v\n%s", err, out)
	}
	text := string(out)
	if !strings.Contains(text, "RESULT pass=") {
		t.Fatalf("self-test printed no RESULT line:\n%s", text)
	}
	if !strings.Contains(text, "fail=0") {
		t.Fatalf("self-test RESULT carries failures:\n%s", text)
	}
	if !strings.Contains(text, "PASS isolation-scrub-negative-control") {
		t.Fatalf("the scrub negative control did not pass (the gate missed a planted leak):\n%s", text)
	}
	if !strings.Contains(text, "PASS protected-set-hash") {
		t.Fatalf("the protected-set hash case did not pass:\n%s", text)
	}
	// Card t1438 review finding 5: the hash must be non-vacuous — a REAL
	// tamper on a scratch fake home must flip it, and the self-check records
	// that the caught change happened. Without this control a name-only hash
	// reads green while rewriting every protected file.
	if !strings.Contains(text, "PASS protected-set-hash-negative-control") {
		t.Fatalf("the protected-set hash negative control did not pass (a modified protected file did not flip the hash — the tamper check is vacuous):\n%s", text)
	}
}

// TestShrinkVerificationNeverReachesRealHome is AC-020 (a), M1 half: the
// gate scripts never assign HOME, the scrub is live-enumerated (static
// shape), and the default plugin runner refuses to exec under a test binary
// (the t1435 REQ-017 seam this SPEC's probe reuses — positive control).
func TestShrinkVerificationNeverReachesRealHome(t *testing.T) {
	t.Run("no-script-assigns-home", func(t *testing.T) {
		scriptDir := resolutionGateScriptDir(t)
		for _, name := range []string{"check-bare-name-resolution.sh", "test-bare-name-resolution.sh"} {
			data, err := os.ReadFile(filepath.Join(scriptDir, name))
			if err != nil {
				t.Fatalf("read %s: %v", name, err)
			}
			for i, line := range strings.Split(string(data), "\n") {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "#") {
					continue
				}
				// An ASSIGNMENT TO HOME: the token HOME at a word start,
				// followed by = — CODEX_HOME= and CLAUDE_CONFIG_DIR= are
				// the scratch-home writes and must not match (their prefix
				// ends in a word character).
				if homeAssignAt(line) {
					t.Errorf("%s:%d assigns HOME: %s", name, i+1, trimmed)
				}
			}
		}
	})

	t.Run("scrub-is-live-enumerated", func(t *testing.T) {
		scriptDir := resolutionGateScriptDir(t)
		data, err := os.ReadFile(filepath.Join(scriptDir, "check-bare-name-resolution.sh"))
		if err != nil {
			t.Fatalf("read script: %v", err)
		}
		text := string(data)
		// Static shape check (labelled static): the scrub enumerates the
		// live environment rather than carrying a hand list, and covers all
		// three variable families REQ-020 names.
		if !strings.Contains(text, "env") {
			t.Error("the script never reads the live environment — the scrub cannot be live-enumerated")
		}
		for _, family := range []string{"MOAI", "CLAUDE", "CODEX"} {
			if !strings.Contains(text, family+"_") {
				t.Errorf("the script's scrub does not mention the %s_ variable family", family)
			}
		}
	})

	t.Run("default-runner-refuses-under-test-binary", func(t *testing.T) {
		// Positive control for the REQ-017 seam: under a Go test binary the
		// default runner refuses, and the install step stays silent and
		// nil-returning even with a non-empty tool list and no opt-out.
		if !isPluginTestBinary() {
			t.Skip("not a test binary (unexpected under go test)")
		}
		var buf strings.Builder
		opts := newPluginInstallOptions([]pluginTool{pluginToolClaude}, t.TempDir(), false)
		if err := runPluginInstallStep(&buf, opts); err != nil {
			t.Fatalf("runPluginInstallStep under a test binary returned %v, want nil", err)
		}
		if buf.Len() != 0 {
			t.Fatalf("the refusing path wrote output: %q", buf.String())
		}
	})
}

// homeAssignAt reports whether line assigns the HOME variable: the token
// HOME at a word start (line start or a non-word character before it)
// followed by '='. CODEX_HOME= does not match — the character before HOME
// there is an underscore, a word character.
func homeAssignAt(line string) bool {
	isWordChar := func(c byte) bool {
		return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
	}
	for i := 0; i+5 <= len(line); i++ {
		if line[i:i+4] != "HOME" || line[i+4] != '=' {
			continue
		}
		if i == 0 || !isWordChar(line[i-1]) {
			return true
		}
	}
	return false
}

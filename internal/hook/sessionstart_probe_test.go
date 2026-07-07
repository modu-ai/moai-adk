// Package hook — SPEC-HOOK-SESSIONSTART-PROBE-001 acceptance-criteria tests.
//
// These tests are the TDD harness for the SessionStart probe wrapper change
// (Branch A: wrapper-only, self-contained bash). Each test function maps 1:1
// to an acceptance criterion in .moai/specs/SPEC-HOOK-SESSIONSTART-PROBE-001/acceptance.md
// and executes the exact AC verification command from that document.
//
// AC-001 (success-path byte-identical) and AC-005 (30 wrappers untouched) rely
// on `git show HEAD~1:` and therefore can only be mechanically verified AFTER
// the run-phase commit lands; they are covered by the post-commit §E.2 evidence
// batch (bash, verbatim) rather than by these Go tests.
package hook

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// sessionstartProbeProjectRoot returns the absolute project-root path derived
// from this test file's location (internal/hook/ → two levels up). Used to
// resolve the real wrapper script path and the template mirror path so the
// tests exercise the on-disk artifacts that ship to users, NOT a fixture copy.
func sessionstartProbeProjectRoot(t *testing.T) string {
	t.Helper()
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed — cannot locate test file")
	}
	// testFile = .../internal/hook/sessionstart_probe_test.go
	// project root = two dirs up from internal/hook/
	root := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", ".."))
	return root
}

// sessionstartProbeWrapperPath returns the absolute path to the local wrapper.
func sessionstartProbeWrapperPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(sessionstartProbeProjectRoot(t),
		".claude", "hooks", "moai", "handle-session-start.sh")
}

// sessionstartProbeTemplatePath returns the absolute path to the template mirror.
func sessionstartProbeTemplatePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(sessionstartProbeProjectRoot(t),
		"internal", "template", "templates", ".claude", "hooks", "moai",
		"handle-session-start.sh.tmpl")
}

// TestSessionStartProbe_AC002_All3TiersAbsentEmitsWarning verifies AC-HOOK-002
// (traceable to REQ-HOOK-002): when all 3 moai-binary resolution tiers are
// absent AND source=startup, the wrapper fallback emits a warning to BOTH
// (a) the stderr log at $HOME/.moai/logs/hook-stderr.log containing 'moai' +
//     at least one of {PATH, go/bin, .local/bin}, AND
// (b) stdout JSON carrying hookSpecificOutput + additionalContext.
func TestSessionStartProbe_AC002_All3TiersAbsentEmitsWarning(t *testing.T) {
	skipOnWindows(t)
	wrapper := sessionstartProbeWrapperPath(t)
	if _, err := os.Stat(wrapper); err != nil {
		t.Fatalf("wrapper not found: %v", err)
	}

	tmpHome := t.TempDir()

	cmd := exec.Command("bash", wrapper)
	// Stubbed env: PATH has no moai, HOME is empty tempdir (no go/bin, no .local/bin).
	cmd.Env = []string{"PATH=/usr/bin:/bin", "HOME=" + tmpHome}
	cmd.Stdin = strings.NewReader(`{"source":"startup"}`)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("wrapper execution failed: %v (stderr: %s)", err, stderr.String())
	}

	out := stdout.String()

	// (a) stderr log written with 'moai' substring.
	stderrLogPath := filepath.Join(tmpHome, ".moai", "logs", "hook-stderr.log")
	stderrLog, err := os.ReadFile(stderrLogPath)
	if err != nil {
		t.Fatalf("STDERR-FAIL: hook-stderr.log not created at %s: %v", stderrLogPath, err)
	}
	if !strings.Contains(string(stderrLog), "moai") {
		t.Errorf("STDERR-FAIL: log missing 'moai' substring. log=%q", string(stderrLog))
	}
	// At least one of {PATH, go/bin, .local/bin} must be named.
	if !strings.Contains(string(stderrLog), "PATH") &&
		!strings.Contains(string(stderrLog), "go/bin") &&
		!strings.Contains(string(stderrLog), ".local/bin") {
		t.Errorf("STDERR-FAIL: log names none of PATH/go/bin/.local/bin. log=%q", string(stderrLog))
	}

	// (b) stdout JSON carries hookSpecificOutput + additionalContext.
	if !strings.Contains(out, "hookSpecificOutput") {
		t.Errorf("STDOUT-FAIL: missing 'hookSpecificOutput'. stdout=%q", out)
	}
	if !strings.Contains(out, "additionalContext") {
		t.Errorf("STDOUT-FAIL: missing 'additionalContext'. stdout=%q", out)
	}
}

// TestSessionStartProbe_AC003_NonBlockingExit0 verifies AC-HOOK-003
// (traceable to REQ-HOOK-003): the wrapper fallback exits 0 (NOT 1, NOT 2).
// Session start proceeds uninterrupted.
func TestSessionStartProbe_AC003_NonBlockingExit0(t *testing.T) {
	skipOnWindows(t)
	wrapper := sessionstartProbeWrapperPath(t)

	tmpHome := t.TempDir()

	cmd := exec.Command("bash", wrapper)
	cmd.Env = []string{"PATH=/usr/bin:/bin", "HOME=" + tmpHome}
	cmd.Stdin = strings.NewReader(`{"source":"startup"}`)

	if err := cmd.Run(); err != nil {
		t.Fatalf("expected exit 0, got error: %v (exit code: %d)",
			err, cmd.ProcessState.ExitCode())
	}
	if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != 0 {
		t.Fatalf("expected exit code 0, got: %v",
			cmd.ProcessState.ExitCode())
	}
}

// TestSessionStartProbe_AC004_OncePerSessionDedup verifies AC-HOOK-004
// (traceable to REQ-HOOK-004): the warning is emitted exactly ONCE (on the
// startup invocation); the resume/clear/compact invocations emit NO warning.
func TestSessionStartProbe_AC004_OncePerSessionDedup(t *testing.T) {
	skipOnWindows(t)
	wrapper := sessionstartProbeWrapperPath(t)

	for _, src := range []string{"startup", "resume", "clear", "compact"} {
		t.Run(src, func(t *testing.T) {
			tmpHome := t.TempDir()

			cmd := exec.Command("bash", wrapper)
			cmd.Env = []string{"PATH=/usr/bin:/bin", "HOME=" + tmpHome}
			cmd.Stdin = strings.NewReader(`{"source":"` + src + `"}`)

			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &bytes.Buffer{}

			if err := cmd.Run(); err != nil {
				t.Fatalf("wrapper failed for source=%s: %v", src, err)
			}

			warnCount := strings.Count(stdout.String(), "additionalContext")
			if src == "startup" {
				if warnCount != 1 {
					t.Errorf("source=%s: expected WARN_COUNT=1 (EMITTED), got %d. stdout=%q",
						src, warnCount, stdout.String())
				}
			} else {
				if warnCount != 0 {
					t.Errorf("source=%s: expected WARN_COUNT=0 (SUPPRESSED), got %d. stdout=%q",
						src, warnCount, stdout.String())
				}
			}
		})
	}
}

// TestSessionStartProbe_AC006_TemplateFirstMirrorParity verifies AC-HOOK-006
// (traceable to REQ-HOOK-006): the fallback branch of the local .sh and the
// template .sh.tmpl are byte-identical (this wrapper uses no template vars).
func TestSessionStartProbe_AC006_TemplateFirstMirrorParity(t *testing.T) {
	skipOnWindows(t)
	local := sessionstartProbeWrapperPath(t)
	tmpl := sessionstartProbeTemplatePath(t)

	// awk '/^# Not found/,/^exit 0/' on both files; diff must be empty.
	awkRange := func(path string) ([]byte, error) {
		c := exec.Command("awk", "/^# Not found/,/^exit 0/", path)
		var out bytes.Buffer
		c.Stdout = &out
		if err := c.Run(); err != nil {
			return nil, err
		}
		return out.Bytes(), nil
	}

	localBranch, err := awkRange(local)
	if err != nil {
		t.Fatalf("awk on local failed: %v", err)
	}
	tmplBranch, err := awkRange(tmpl)
	if err != nil {
		t.Fatalf("awk on template failed: %v", err)
	}

	if !bytes.Equal(localBranch, tmplBranch) {
		t.Errorf("PARITY-FAIL: local and template fallback branches differ.\n--- local ---\n%s\n--- template ---\n%s",
			localBranch, tmplBranch)
	}
}

// TestSessionStartProbe_AC007_WarningContentActionability verifies AC-HOOK-007
// (traceable to REQ-HOOK-007): the warning text contains all 4 mandated elements
// — (a) tier enumeration PATH + go/bin + .local/bin, (b) consequence '31 wrappers',
// (c) non-blocking framing ('non-blocking' or 'advisory'), (d) remediation hint
// ('reinstall' / 'restore PATH' / 'rebuild' / 'make build').
func TestSessionStartProbe_AC007_WarningContentActionability(t *testing.T) {
	skipOnWindows(t)
	wrapper := sessionstartProbeWrapperPath(t)

	tmpHome := t.TempDir()

	cmd := exec.Command("bash", wrapper)
	cmd.Env = []string{"PATH=/usr/bin:/bin", "HOME=" + tmpHome}
	cmd.Stdin = strings.NewReader(`{"source":"startup"}`)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {
		t.Fatalf("wrapper failed: %v", err)
	}

	// Extract additionalContext value via the AC-007 grep pipeline.
	extract := exec.Command("bash", "-c",
		`echo "$0" | grep -oE '"additionalContext":"[^"]*"' | sed 's/"additionalContext":"//;s/"$//'`,
		stdout.String())
	var warn bytes.Buffer
	extract.Stdout = &warn
	if err := extract.Run(); err != nil {
		t.Fatalf("failed to extract additionalContext: %v (stdout=%q)", err, stdout.String())
	}
	warnText := warn.String()
	if warnText == "" {
		t.Fatalf("extracted warning text is empty. stdout=%q", stdout.String())
	}

	checks := []struct {
		label string
		ok    bool
	}{
		{"(a) PATH", strings.Contains(warnText, "PATH")},
		{"(a) go/bin", strings.Contains(warnText, "go/bin")},
		{"(a) .local/bin", strings.Contains(warnText, ".local/bin")},
		{"(b) 31 wrappers", strings.Contains(warnText, "31 wrappers")},
		{"(c) non-blocking|advisory",
			strings.Contains(warnText, "non-blocking") || strings.Contains(warnText, "advisory")},
		{"(d) reinstall|restore PATH|rebuild|make build",
			strings.Contains(warnText, "reinstall") ||
				strings.Contains(warnText, "restore PATH") ||
				strings.Contains(warnText, "rebuild") ||
				strings.Contains(warnText, "make build")},
	}

	var failed []string
	for _, c := range checks {
		if !c.ok {
			failed = append(failed, c.label)
		}
	}
	if len(failed) > 0 {
		t.Errorf("WARN-CONTENT-FAIL: missing %v. warn=%q", failed, warnText)
	}
}

// TestSessionStartProbe_EdgeCases_EmptyAndMalformedStdin verifies the §D.7
// edge-case contract: empty stdin / malformed JSON → emit warning (treat as
// startup — safer to warn than to suppress). This is NOT one of the 7 ACs but
// is an explicit manager-develop obligation per acceptance.md §D.7.
func TestSessionStartProbe_EdgeCases_EmptyAndMalformedStdin(t *testing.T) {
	skipOnWindows(t)
	wrapper := sessionstartProbeWrapperPath(t)

	cases := []struct {
		name  string
		stdin string
	}{
		{"empty", ""},
		{"malformed", `{malformed}`},
		{"not-json", `not a json at all`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tmpHome := t.TempDir()

			cmd := exec.Command("bash", wrapper)
			cmd.Env = []string{"PATH=/usr/bin:/bin", "HOME=" + tmpHome}
			cmd.Stdin = strings.NewReader(c.stdin)

			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &bytes.Buffer{}

			if err := cmd.Run(); err != nil {
				t.Fatalf("wrapper failed: %v", err)
			}
			if cmd.ProcessState.ExitCode() != 0 {
				t.Fatalf("exit code: %d (expected 0)", cmd.ProcessState.ExitCode())
			}
			if !strings.Contains(stdout.String(), "additionalContext") {
				t.Errorf("edge-case %s: expected warning emitted (safer-to-warn), stdout=%q",
					c.name, stdout.String())
			}
		})
	}
}

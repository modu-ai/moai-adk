package cli

// SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2d, Q4: the sync gate's decision core in
// Go is a parallel implementation of the distributed Claude script, which stays
// byte-identical. These tests hold the Go predicates to the script itself —
// they read or source the unmodified script rather than restating its rules —
// and pin the out-of-hook receipt producer `moai verify sync-gate`.

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/verify"
)

func syncGateScriptPath(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("..", "template", "templates", ".claude", "hooks", "moai", "sync-phase-quality-gate.sh"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// TestSyncGateSubjectPredicateMatchesScript runs the script's own subject case
// arm (read out of the script, not restated) against each subject and compares
// it with isSyncPhaseSubject.
func TestSyncGateSubjectPredicateMatchesScript(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not on PATH")
	}
	body, err := os.ReadFile(syncGateScriptPath(t))
	if err != nil {
		t.Fatal(err)
	}
	arm := regexp.MustCompile(`(?m)^\s*(\*"docs\(".*sync.*)\)\s*$`).FindSubmatch(body)
	if arm == nil {
		t.Fatal("the script's sync-phase subject case arm was not found; the parity test has nothing to compare against")
	}
	pattern := string(arm[1])
	subjects := []string{
		"docs(SPEC-X-001): sync-phase close",
		"chore(SPEC-X-001): sync-phase artifacts",
		"docs: sync readme",
		"chore: sync lockfile",
		"feat(cli): add a thing",
		"docs(SPEC-X-001): sync phase close",
		"fix: resync docs: sync later",
		"docs(a)(b): sync-phase",
		"docs(x) : sync-phase",
		"Docs: sync",
		"",
		"chore(x): sync-phaseX",
		"prefix docs(x): sync-phase suffix",
	}
	for _, s := range subjects {
		out, err := exec.Command("bash", "-c", `case "$1" in `+pattern+`) echo yes ;; *) echo no ;; esac`, "_", s).Output()
		if err != nil {
			t.Fatalf("bash case over %q: %v", s, err)
		}
		script := strings.TrimSpace(string(out)) == "yes"
		if got := isSyncPhaseSubject(s); got != script {
			t.Errorf("subject %q: Go predicate = %v, script = %v", s, got, script)
		}
	}
}

// TestSyncGateLanguageDetectionMatchesScript sources the script and compares
// its detect_languages output with detectSyncGateLanguages on the same trees.
func TestSyncGateLanguageDetectionMatchesScript(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not on PATH")
	}
	script := syncGateScriptPath(t)
	trees := map[string][]string{
		"go module":             {"go.mod"},
		"nested java":           {"src/main/java/com/acme/App.java"},
		"kotlin source":         {"app/src/Main.kt", "build.gradle.kts"},
		"pruned node":           {"node_modules/x/index.js", "README.md"},
		"python and go":         {"tool.py", "cmd/main.go"},
		"visual studio dir":     {".vs/settings.json"},
		"cpp header only":       {"include/x.h"},
		"r lower":               {"analysis/a.r"},
		"docs only":             {"docs/a.md"},
		"swift package":         {"Package.swift"},
		"elixir script":         {"lib/a.exs"},
		"gradle java kts plain": {"build.gradle.kts"},
	}
	for name, files := range trees {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			for _, rel := range files {
				p := filepath.Join(root, rel)
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			out, err := exec.Command("bash", "-c", `source "$1" && detect_languages "$2"`, "_", script, root).Output()
			if err != nil {
				t.Fatalf("source script: %v", err)
			}
			want := strings.Fields(string(out))
			got := detectSyncGateLanguages(root)
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Fatalf("languages: Go = %v, script = %v", got, want)
			}
		})
	}
}

// TestVerifySyncGateRecordsReceipt: `moai verify sync-gate` runs the checks out
// of hook and records a receipt the Stop chain reads (operator M2d decision
// (a): a verify-snapshot receipt; .moai/state/sync-quality-gate.last is not
// written).
func TestVerifySyncGateRecordsReceipt(t *testing.T) {
	f := newStopFixture(t)
	ctx := context.Background()

	run := func(t *testing.T) map[string]any {
		t.Helper()
		cmd := newVerifyCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs([]string{"sync-gate", "--project-root", f.root})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("moai verify sync-gate: %v", err)
		}
		var got map[string]any
		if err := json.Unmarshal(out.Bytes(), &got); err != nil {
			t.Fatalf("output is not JSON: %q", out.String())
		}
		return got
	}
	current := func(t *testing.T) *verify.Receipt {
		t.Helper()
		key, err := verify.Key(ctx, f.root)
		if err != nil {
			t.Fatal(err)
		}
		state, _ := syncGateReceiptState(key, detectSyncGateLanguages(f.root))
		return verify.LoadReceipt(f.root, state)
	}

	f.setFakeGo(t, 1)
	if got := run(t); got["verdict"] != "fail" {
		t.Fatalf("failing vet: verdict = %v, want fail (%v)", got["verdict"], got)
	}
	r := current(t)
	if r == nil || r.Verdict != "fail" || r.ExitCode&syncGateC1Failed == 0 {
		t.Fatalf("receipt after a failing vet = %+v, want verdict fail with the c1 bit", r)
	}

	f.setFakeGo(t, 0)
	if got := run(t); got["verdict"] != "pass" {
		t.Fatalf("passing vet: verdict = %v, want pass", got["verdict"])
	}
	if r := current(t); r == nil || r.Verdict != "pass" || r.ExitCode != 0 {
		t.Fatalf("receipt after a passing vet = %+v, want verdict pass", r)
	}
	if _, err := os.Stat(filepath.Join(f.root, ".moai", "state", "sync-quality-gate.last")); err == nil {
		t.Fatal("the Codex producer wrote the Claude-only record .moai/state/sync-quality-gate.last")
	}
}

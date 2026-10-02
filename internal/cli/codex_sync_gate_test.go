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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
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
		"cxx source":            {"src/kernel.cxx"},
		"hpp source":            {"include/x.hpp"},
		"hxx source":            {"include/x.hxx"},
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

// TestSyncGateCxxSourceDetectsCpp pins the semantics the parity test cannot
// see: both implementations share the same detection line, so both can miss
// the same case together and still agree. A tree whose only C++ evidence is a
// .cxx file must resolve to cpp on each side independently (card t1420) — the
// compile step and code_delta_pattern already cover .cxx, so a missed
// detection let a .cxx-only sync commit pass without any checker running.
func TestSyncGateCxxSourceDetectsCpp(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not on PATH")
	}
	root := t.TempDir()
	p := filepath.Join(root, "src", "kernel.cxx")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Run("go port", func(t *testing.T) {
		got := detectSyncGateLanguages(root)
		if !slices.Contains(got, "cpp") {
			t.Fatalf("languages: Go = %v, want cpp present", got)
		}
	})
	t.Run("script", func(t *testing.T) {
		out, err := exec.Command("bash", "-c", `source "$1" && detect_languages "$2"`, "_", syncGateScriptPath(t), root).Output()
		if err != nil {
			t.Fatalf("source script: %v", err)
		}
		if !slices.Contains(strings.Fields(string(out)), "cpp") {
			t.Fatalf("languages: script = %q, want cpp present", string(out))
		}
	})
}

// TestSyncGateCppHeaderSuffixesDetectCpp extends the t1420 semantics pin to
// the sibling header suffixes: .hpp and .hxx-only trees must resolve to cpp
// on each side independently (card t1412 batch, t1420 residual) — the compile
// step, code_delta_pattern, and changed-file filter cover them, so a missed
// detection let a header-only sync commit pass without any checker running.
func TestSyncGateCppHeaderSuffixesDetectCpp(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not on PATH")
	}
	for _, ext := range []string{"hpp", "hxx"} {
		t.Run(ext, func(t *testing.T) {
			root := t.TempDir()
			p := filepath.Join(root, "include", "x."+ext)
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			got := detectSyncGateLanguages(root)
			if !slices.Contains(got, "cpp") {
				t.Fatalf("languages: Go = %v, want cpp present", got)
			}
			out, err := exec.Command("bash", "-c", `source "$1" && detect_languages "$2"`, "_", syncGateScriptPath(t), root).Output()
			if err != nil {
				t.Fatalf("source script: %v", err)
			}
			if !slices.Contains(strings.Fields(string(out)), "cpp") {
				t.Fatalf("languages: script = %q, want cpp present", string(out))
			}
		})
	}
}

// TestSyncGateGoRootsSurviveWholeModuleDeletion drives the gate's go module
// root resolution through the deletion scenario the card premise describes
// (card t1412): a multi-module tree without a root go.mod whose changed files
// belong to a module deleted whole. The script's source guard stops sourcing
// before these functions are defined, so the test extracts their exact
// committed bytes by function-name range and runs them in bash.
func TestSyncGateGoRootsSurviveWholeModuleDeletion(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not on PATH")
	}
	script := syncGateScriptPath(t)
	root := t.TempDir()
	for rel, body := range map[string]string{
		"modules/b/go.mod": "module example.com/b\n\ngo 1.21\n",
		"modules/b/b.go":   "package b\n",
	} {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	extract := func(fn string) string {
		t.Helper()
		out, err := exec.Command("bash", "-c",
			"awk '/^"+fn+"\\(\\) \\{/,/^\\}/' \"$1\"", "_", script).Output()
		if err != nil {
			t.Fatalf("extract %s: %v", fn, err)
		}
		if len(out) == 0 {
			t.Fatalf("extract %s: empty range", fn)
		}
		return string(out)
	}
	frag := filepath.Join(t.TempDir(), "goroots.sh")
	if err := os.WriteFile(frag, []byte(
		extract("find_go_module_roots")+
			extract("find_surviving_go_module_roots")+
			extract("resolve_go_roots")), 0o644); err != nil {
		t.Fatal(err)
	}
	probe := func(projectRoot, delta string) string {
		t.Helper()
		payload := `source "$1"
PROJECT_ROOT=$2
SYNC_DELTA_FILES=$3
printf 'walk=[%s]\n' "$(find_go_module_roots)"
printf 'surviving=[%s]\n' "$(find_surviving_go_module_roots)"
printf 'resolved=[%s]\n' "$(resolve_go_roots)"
`
		out, err := exec.Command("bash", "-c", payload, "_", frag, projectRoot, delta).Output()
		if err != nil {
			t.Fatalf("probe: %v", err)
		}
		return string(out)
	}
	// Deletion scenario: the delta holds only the deleted module's paths, so
	// the walk resolves nothing, surviving discovery finds modules/b, and the
	// ladder picks it — pre-fix this fell to the root anchor and blocked
	// "does not contain main module" on a healthy tree.
	wantMod := filepath.Join(root, "modules/b")
	if got := probe(root, "modules/a/a.go\nmodules/a/go.mod"); got != fmt.Sprintf("walk=[]\nsurviving=[%s]\nresolved=[%s]\n", wantMod, wantMod) {
		t.Fatalf("deletion ladder: got %q, want walk=[] surviving/resolved=%s", got, wantMod)
	}
	// Delta rooted in the surviving module: the walk still wins.
	if got := probe(root, "modules/b/b.go"); got != fmt.Sprintf("walk=[%s]\nsurviving=[%s]\nresolved=[%s]\n", wantMod, wantMod, wantMod) {
		t.Fatalf("delta ladder: got %q, want walk/surviving/resolved=%s", got, wantMod)
	}
	// No go.mod anywhere: the bare root anchor survives as the final fallback
	// — its failure there is a real one (design intent, unchanged).
	bare := t.TempDir()
	if got := probe(bare, ""); got != fmt.Sprintf("walk=[]\nsurviving=[]\nresolved=[%s]\n", bare) {
		t.Fatalf("bare ladder: got %q, want resolved=%s", got, bare)
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

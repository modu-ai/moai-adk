package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// ─── The codex live-observation axis — an explicit declaration ─────────────
//
// The codex live tests in this package are opt-in: they spend real codex /
// z.ai quota, so they are gated behind an environment switch or the presence
// of a working codex binary. CI installs no codex binary and sets no switch,
// so EVERY test listed below skips there — and a skip is not a pass.
//
// That fact used to live nowhere. The package reported `ok` and the axis went
// unobserved in silence. This file is the declaration that replaces the
// silence, and the guards below are what keep the declaration true:
//
//   - codexLiveAxis lists every live-gated file, its enabling switches, and
//     what it would observe if it ran.
//   - TestCodexLiveAxis_DeclarationCoversEveryLiveGatedFile fails when a new
//     live-gated test file is added without being declared here.
//   - TestCodexLiveAxis_DeclaredSwitchesExistInSource fails when a declared
//     switch drifts from the switch the file actually reads.
//   - TestCodexLiveAxis_NotObservedInCI fails when CI DOES wire one of these
//     switches — at which point the "unobserved in CI" claim below is no
//     longer true and this declaration must be rewritten to say so.
//
// To observe the axis by hand (spends real quota):
//
//	MOAI_CODEX_LIVE_PROBE=1 MOAI_AUDIT_PIN_LIVE=1 go test ./internal/cli/ -run 'Live' -v
//
// Provisioning codex + credentials in CI is an operator decision (cost and
// secret exposure), deliberately NOT taken here: a credential-less CI job
// would skip exactly as it does today, relocating the silence rather than
// ending it.

// codexLiveAxisEntry declares one live-gated test file.
type codexLiveAxisEntry struct {
	file string // test file in this package
	// switches are the env vars / conditions that must hold for the file's
	// live tests to run. Each must literally occur in the file's source.
	switches []string
	// tests is the number of top-level Test functions in the file. Declaring
	// it keeps "how many tests are unobserved" honest: add one without
	// updating this and the guard fails.
	tests int
	// observes names what stays unobserved while the switches are unset.
	observes string
}

var codexLiveAxis = []codexLiveAxisEntry{
	{
		file:     "codex_live_protocol_probe_test.go",
		switches: []string{"MOAI_CODEX_LIVE_PROBE", "MOAI_CODEX_LIVE_BIN"},
		tests:    5,
		observes: "real app-server protocol behaviour: thread reuse, turn interrupt, sandbox-policy stickiness, approval stall, review turn.started",
	},
	{
		file:     "codex_review_gate_live_test.go",
		switches: []string{"MOAI_SKIP_LIVE_CODEX", "codexBinaryName"},
		tests:    1,
		observes: "the review gate blocking a real command-injection sink and a hardcoded key against a live codex",
	},
	{
		file:     "codex_review_target_live_test.go",
		switches: []string{"MOAI_SKIP_LIVE_CODEX", "codexBinaryName"},
		tests:    1,
		observes: "AC-CRT-010 — a baseBranch review target is not rejected by a live codex",
	},
	{
		file:     "audit_pin_live_test.go",
		switches: []string{"MOAI_AUDIT_PIN_LIVE", "codexBinaryName"},
		tests:    2,
		observes: "AC-AMP-006 GLM reasoning-delivery differential and AC-AMP-007 codex pin confirmation",
	},
}

// codexLiveAxisMinFiles is a floor guarding against a silently emptied
// registry: the guards below compare sets, and a comparison of two empty sets
// asserts nothing.
const codexLiveAxisMinFiles = 4

// codexLiveGatedFiles finds the live-gated test files by their shared idiom:
// resolving the codex binary at run time, or reading a live-only env switch.
func codexLiveGatedFiles(t *testing.T) []string {
	t.Helper()
	lookPath := regexp.MustCompile(`(?:exec\.LookPath|codexLookPath)\(codexBinaryName\)`)
	liveEnv := regexp.MustCompile(`"MOAI_(?:CODEX_LIVE_PROBE|AUDIT_PIN_LIVE)"`)

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	var found []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, "_test.go") {
			continue
		}
		if name == "codex_live_axis_declaration_test.go" {
			continue // this file names the switches in order to declare them
		}
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		src := string(b)
		if lookPath.MatchString(src) || liveEnv.MatchString(src) {
			found = append(found, name)
		}
	}
	sort.Strings(found)
	return found
}

func codexLiveAxisDeclaredFiles() []string {
	files := make([]string, 0, len(codexLiveAxis))
	for _, e := range codexLiveAxis {
		files = append(files, e.file)
	}
	sort.Strings(files)
	return files
}

// TestCodexLiveAxis_DeclarationCoversEveryLiveGatedFile — the declaration and
// the package agree on which files are live-gated, in both directions.
func TestCodexLiveAxis_DeclarationCoversEveryLiveGatedFile(t *testing.T) {
	if len(codexLiveAxis) < codexLiveAxisMinFiles {
		t.Fatalf("codexLiveAxis declares %d files, floor is %d — an emptied registry would make every guard in this file vacuous",
			len(codexLiveAxis), codexLiveAxisMinFiles)
	}

	found := codexLiveGatedFiles(t)
	declared := codexLiveAxisDeclaredFiles()

	if len(found) == 0 {
		t.Fatal("detected zero live-gated files — the detection idiom in codexLiveGatedFiles is stale, so this guard asserted nothing")
	}

	inDeclared := map[string]bool{}
	for _, f := range declared {
		inDeclared[f] = true
	}
	for _, f := range found {
		if !inDeclared[f] {
			t.Errorf("%s is live-gated but undeclared — add it to codexLiveAxis, or its skips stay silent", f)
		}
	}
	inFound := map[string]bool{}
	for _, f := range found {
		inFound[f] = true
	}
	for _, f := range declared {
		if !inFound[f] {
			t.Errorf("codexLiveAxis declares %s, but it is not live-gated (renamed, deleted, or no longer gated) — update the declaration", f)
		}
	}
}

// TestCodexLiveAxis_DeclaredSwitchesExistInSource — each declared switch is
// really the switch that file reads, and the declared test count is real.
func TestCodexLiveAxis_DeclaredSwitchesExistInSource(t *testing.T) {
	testFunc := regexp.MustCompile(`(?m)^func (Test\w+)\(t \*testing\.T\)`)
	for _, e := range codexLiveAxis {
		b, err := os.ReadFile(e.file)
		if err != nil {
			t.Errorf("declared file %s: %v", e.file, err)
			continue
		}
		src := string(b)

		if len(e.switches) == 0 {
			t.Errorf("%s: declares no switches — an unswitched entry claims nothing", e.file)
		}
		for _, sw := range e.switches {
			if !strings.Contains(src, sw) {
				t.Errorf("%s: declared switch %q does not occur in the file — the declaration drifted from the code", e.file, sw)
			}
		}
		if got := len(testFunc.FindAllString(src, -1)); got != e.tests {
			t.Errorf("%s: declares %d live tests, source has %d — update the declaration so the unobserved count stays honest", e.file, e.tests, got)
		}
		if strings.TrimSpace(e.observes) == "" {
			t.Errorf("%s: declares no `observes` text — what goes unobserved must be stated, not left blank", e.file)
		}
	}
}

// TestCodexLiveAxis_NotObservedInCI — the standing claim of this file is that
// CI does not observe this axis. That claim is checked, not asserted: if a
// workflow ever wires one of the declared switches, this fails and the
// declaration above must be rewritten to describe the new reality.
func TestCodexLiveAxis_NotObservedInCI(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	dir := filepath.Join(root, ".github", "workflows")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skipf("no workflow directory at %s: %v", dir, err)
	}

	// Only env-var switches are checked here: codexBinaryName is a Go
	// identifier, not something a workflow would ever contain.
	var envSwitches []string
	for _, e := range codexLiveAxis {
		for _, sw := range e.switches {
			if strings.HasPrefix(sw, "MOAI_") {
				envSwitches = append(envSwitches, sw)
			}
		}
	}
	if len(envSwitches) == 0 {
		t.Fatal("no env switches declared — this guard would pass without checking anything")
	}

	scanned, wired := 0, 0
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read workflow %s: %v", name, err)
		}
		scanned++
		src := string(b)
		for _, sw := range envSwitches {
			if strings.Contains(src, sw) {
				wired++
				t.Errorf(".github/workflows/%s wires %s — the codex live axis IS observed in CI now, so the declaration in codex_live_axis_declaration_test.go is stale and must be rewritten",
					name, sw)
			}
		}
	}
	if scanned == 0 {
		t.Fatal("scanned zero workflow files — this guard asserted nothing")
	}
	// Logged only when the claim actually holds: an unconditional "none wires"
	// line would still print beside the failures above and contradict them.
	if wired == 0 {
		t.Logf("codex live axis unobserved in CI: %d workflow files scanned, none wires any of %v", scanned, envSwitches)
	}
}

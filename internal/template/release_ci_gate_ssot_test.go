// release_ci_gate_ssot_test.go: the CI gate in scripts/release.sh reads the
// check names it demands from the release_gate block of
// .github/required-checks.yml, resolved relative to the SCRIPT FILE (card
// t1858). An empty or unreadable list would make the gate demand nothing, so
// the script must refuse. These tests run a COPY of the real script from a
// scratch layout, <tmp>/scripts/release.sh with <tmp>/.github/required-checks.yml
// set per case, with `--dry-run` and the stub gh; nothing here tags or pushes.
package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReleaseScriptCIGateSSoTFailsClosed: a missing SSoT file and a file with no
// release_gate names both refuse before any check run is read; the control copy
// with the real SSoT passes, so the refusals are not an artifact of the layout.
func TestReleaseScriptCIGateSSoTFailsClosed(t *testing.T) {
	rlsRequireTools(t, "jq")

	root := findProjectRootForMirrorTest(t)
	realScript, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rlsReleaseScriptRel)))
	if err != nil {
		t.Fatalf("read real release.sh: %v", err)
	}
	realSSoT, err := os.ReadFile(filepath.Join(root, ".github", "required-checks.yml"))
	if err != nil {
		t.Fatalf("read real required-checks.yml: %v", err)
	}

	const mainOnly = "version: 1\nbranches:\n  main:\n    contexts:\n      - Lint\n"
	rows := []struct {
		name     string
		ssot     *string // nil: the file is absent
		wantExit int
		wantOut  []string
	}{
		{name: "ssot_absent_is_refused", ssot: nil, wantExit: 1,
			wantOut: []string{"[FAIL] CI gate: the required-checks file", "is missing"}},
		{name: "no_release_gate_block_is_refused", ssot: ptr(mainOnly), wantExit: 1,
			wantOut: []string{"[FAIL] CI gate: no release_gate check names could be read"}},
		{name: "empty_release_gate_block_is_refused", ssot: ptr(mainOnly + "release_gate:\nauxiliary:\n  - x\n"), wantExit: 1,
			wantOut: []string{"[FAIL] CI gate: no release_gate check names could be read"}},
		{name: "trailing_comment_is_malformed", ssot: ptr(gateSSoT("  - Lint # keep\n  - \"Test (ubuntu-latest)\"\n  - \"Build (linux/amd64)\"\n")), wantExit: 1,
			wantOut: []string{"[FAIL] CI gate: malformed release_gate in", "trailing comment"}},
		{name: "single_quoted_entry_is_malformed", ssot: ptr(gateSSoT("  - Lint\n  - 'Test (ubuntu-latest)'\n  - \"Build (linux/amd64)\"\n")), wantExit: 1,
			wantOut: []string{"[FAIL] CI gate: malformed release_gate in", "single-quoted"}},
		{name: "duplicate_name_is_malformed", ssot: ptr(gateSSoT("  - Lint\n  - Lint\n  - Lint\n")), wantExit: 1,
			wantOut: []string{"[FAIL] CI gate: malformed release_gate in", "duplicate name"}},
		{name: "duplicate_key_is_malformed", ssot: ptr(gateSSoT("  - Lint\n") + "release_gate:\n  - \"Test (ubuntu-latest)\"\n"), wantExit: 1,
			wantOut: []string{"[FAIL] CI gate: malformed release_gate in", "duplicate release_gate key"}},
		{name: "bare_dash_entry_is_malformed", ssot: ptr(gateSSoT("  - Lint\n  -\n  - \"Test (ubuntu-latest)\"\n  - \"Build (linux/amd64)\"\n")), wantExit: 1,
			wantOut: []string{"[FAIL] CI gate: malformed release_gate in", "empty entry"}},
		{name: "dash_space_entry_is_malformed", ssot: ptr(gateSSoT("  - Lint\n  - \n  - \"Test (ubuntu-latest)\"\n  - \"Build (linux/amd64)\"\n")), wantExit: 1,
			wantOut: []string{"[FAIL] CI gate: malformed release_gate in", "empty entry"}},
		{name: "empty_quoted_entry_is_malformed", ssot: ptr(gateSSoT("  - Lint\n  - \"\"\n  - \"Build (linux/amd64)\"\n")), wantExit: 1,
			wantOut: []string{"[FAIL] CI gate: malformed release_gate in", "empty entry"}},
		{name: "blank_quoted_entry_is_malformed", ssot: ptr(gateSSoT("  - Lint\n  - \"  \"\n  - \"Build (linux/amd64)\"\n")), wantExit: 1,
			wantOut: []string{"[FAIL] CI gate: malformed release_gate in", "empty entry"}},
		{name: "space_padded_quoted_name_is_malformed", ssot: ptr(gateSSoT("  - \" Lint\"\n  - \"Test (ubuntu-latest)\"\n  - \"Build (linux/amd64)\"\n")), wantExit: 1,
			wantOut: []string{"[FAIL] CI gate: malformed release_gate in", "space-padded"}},
		{name: "unrecognized_line_is_malformed", ssot: ptr(gateSSoT("  - Lint\n  -\"Test (ubuntu-latest)\"\n  - \"Build (linux/amd64)\"\n")), wantExit: 1,
			wantOut: []string{"[FAIL] CI gate: malformed release_gate in", "unrecognized line"}},
		{name: "tab_indented_list_passes", ssot: ptr(gateSSoT("\t- Lint\n\t- \"Test (ubuntu-latest)\"\n\t- \"Build (linux/amd64)\"\n")), wantExit: 0,
			wantOut: []string{"CI gate:     verified"}},
		{name: "crlf_real_ssot_copy_passes", ssot: ptr(strings.ReplaceAll(string(realSSoT), "\n", "\r\n")), wantExit: 0,
			wantOut: []string{"CI gate:     verified"}},
		{name: "control_real_ssot_copy_passes", ssot: ptr(string(realSSoT)), wantExit: 0,
			wantOut: []string{"CI gate:     verified"}},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			r := rlsBuildReleaseRepo(t, rlsRcChangelog)
			s := rlsNewStub(t)

			layout := t.TempDir()
			scriptDir := filepath.Join(layout, "scripts")
			if err := os.MkdirAll(scriptDir, 0o755); err != nil {
				t.Fatalf("mkdir scripts: %v", err)
			}
			script := filepath.Join(scriptDir, "release.sh")
			if err := os.WriteFile(script, realScript, 0o755); err != nil {
				t.Fatalf("write script copy: %v", err)
			}
			if row.ssot != nil {
				ghDir := filepath.Join(layout, ".github")
				if err := os.MkdirAll(ghDir, 0o755); err != nil {
					t.Fatalf("mkdir .github: %v", err)
				}
				if err := os.WriteFile(filepath.Join(ghDir, "required-checks.yml"), []byte(*row.ssot), 0o644); err != nil {
					t.Fatalf("write ssot: %v", err)
				}
			}

			res := rlsRun(t, r.work, s.env, "bash", script, rlsRcTag, "--dry-run")
			out := rlsNorm(res.out)
			t.Logf("exit=%d\n%s", res.exit, out)
			if res.exit != row.wantExit {
				t.Errorf("exit code = %d, want %d", res.exit, row.wantExit)
			}
			rlsMustContain(t, out, row.wantOut...)
			// A refusal must come from the guard itself, before any verdict is
			// read: a warning that merely falls through to a later failure would
			// still print the text, so the check-runs endpoint must stay untouched.
			if row.wantExit == 1 {
				for _, c := range s.calls(t) {
					if strings.Contains(c, "check-runs") {
						t.Errorf("a refused gate must not read the check runs; gh calls: %q", s.calls(t))
					}
				}
			}
		})
	}
}

// gateSSoT builds a minimal SSoT whose release_gate block holds the given
// pre-indented list lines.
func gateSSoT(items string) string {
	return "version: 1\nbranches:\n  main:\n    contexts:\n      - Lint\nrelease_gate:\n" + items
}

func ptr(s string) *string { return &s }

// TestReleaseScriptRealSSoTDemandsEveryPinnedCheck drives the REAL script against
// the REAL .github/required-checks.yml (rlsRelease), with check-run pages from
// which exactly one pinned name is absent, and a page holding only Lint. The
// script names no check itself, so these rows are what fail the moment someone
// shortens the real release_gate list: the gate would then stop demanding the
// dropped name and the corresponding row would exit 0.
func TestReleaseScriptRealSSoTDemandsEveryPinnedCheck(t *testing.T) {
	rlsRequireTools(t, "jq")

	run := func(name, conclusion string) string {
		return `{"name":"` + name + `","status":"completed","conclusion":"` + conclusion + `","app":{"slug":"github-actions"}}`
	}
	page := func(runs ...string) string {
		return `{"total_count":` + string(rune('0'+len(runs))) + `,"check_runs":[` + strings.Join(runs, ",") + `]}`
	}
	lint, test, build := run("Lint", "success"), run("Test (ubuntu-latest)", "success"), run("Build (linux/amd64)", "success")
	const tail = "): "

	rows := []struct {
		name    string
		runs    string
		missing string
	}{
		{name: "lint_missing", runs: page(test, build), missing: "Lint. The"},
		{name: "test_missing", runs: page(lint, build), missing: "Test (ubuntu-latest). The"},
		{name: "build_missing", runs: page(lint, test), missing: "Build (linux/amd64). The"},
		{name: "only_lint_green", runs: page(lint), missing: "Test (ubuntu-latest); Build (linux/amd64). The"},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			r := rlsBuildReleaseRepo(t, rlsRcChangelog)
			s := rlsNewStub(t)
			s.setCheckRuns(t, row.runs)
			res := rlsRelease(t, r, s, rlsRcTag, "--dry-run")
			out := rlsNorm(res.out)
			t.Logf("exit=%d\n%s", res.exit, out)
			if res.exit != 1 {
				t.Errorf("exit code = %d, want 1", res.exit)
			}
			rlsMustContain(t, out, "CI gate: expected check run(s) not reported", "github-actions app"+tail+row.missing)
		})
	}
}

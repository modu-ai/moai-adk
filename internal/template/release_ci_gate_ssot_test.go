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

func ptr(s string) *string { return &s }

package cli

// init_resume_record_test.go — card t1591 repair round (SPEC-USERASSET-DEPLOY-GUARD-001),
// the two confirmed defects on the init resume path:
//
//   - Finding 1: the harness selection record was written only in the error
//     branch of the user-asset ensure. A run that died inside the ensure (a
//     crash or a kill, which no error branch observes) left no record, so the
//     resume could not tell --llm both from claude. The record is now written
//     before the ensure runs, and it retires with the checkpoint marker.
//   - Finding 2: a Codex wiring failure on the resume path was a warning, so the
//     resume retired its marker and record and reported success. On the resume
//     path the failure is now an error, returned before any retirement.
//
// TestInitWiringFailureWarnsOutsideResume and TestInitResumeRecordRetirementFollowsMarker
// are guards for behavior the repair must keep or establish: the healthy init path
// stays warn-only on a Codex failure, and the record retires only once no marker stands.

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func stubEnsure(t *testing.T, fn func(homeDir string, selection []string, out io.Writer) error) {
	t.Helper()
	orig := ensureUserAssetsLockedFn
	t.Cleanup(func() { ensureUserAssetsLockedFn = orig })
	ensureUserAssetsLockedFn = fn
}

// writeResumeMarker plants the resume checkpoint marker a failed resume leaves behind.
func writeResumeMarker(t *testing.T, projectDir string) {
	t.Helper()
	marker := filepath.Join(projectDir, ".moai", "resume-pending")
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(marker, []byte("resume in progress\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestInitResumeRecordSurvivesInterruptedEnsure — Finding 1 (re-aimed by card t1619:
// the record opens after the executor, so the interruption is the post-executor
// window). The run dies there. A panic stands in for the process death, so no error
// branch runs. The selection record must already be on disk for the resume to read.
func TestInitResumeRecordSurvivesInterruptedEnsure(t *testing.T) {
	isolatedInitHome(t)
	projectDir := filepath.Join(t.TempDir(), "interrupted-ensure-proj")
	origWindow := initWindowFn
	t.Cleanup(func() { initWindowFn = origWindow })
	initWindowFn = func(string) error {
		panic("simulated process death inside the post-executor window")
	}

	var interrupted any
	func() {
		defer func() { interrupted = recover() }()
		_ = resumeRun(t, projectDir, "")
	}()
	if interrupted == nil {
		t.Fatal("the injected interruption did not bite: the ensure was never reached")
	}
	if got, ok := readResumeWiring(projectDir); !ok || got != agentWiringClaude {
		t.Fatalf("the interrupted run left no selection record for its resume (ok=%v, got=%q, want claude)", ok, got)
	}
}

// TestInitResumeRecordRetirementFollowsMarker — guard for the retirement rule
// (Finding 1). A healthy completion retires the selection record once no checkpoint
// marker stands. While a marker still stands, the record stays for the resume the
// marker names.
func TestInitResumeRecordRetirementFollowsMarker(t *testing.T) {
	cases := []struct {
		name       string
		markerKept bool
	}{
		{name: "no marker: the record retires with the healthy completion", markerKept: false},
		{name: "marker stands: the record stays for the resume", markerKept: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isolatedInitHome(t)
			projectDir := filepath.Join(t.TempDir(), "retirement-proj")
			origWindow := initWindowFn
			t.Cleanup(func() { initWindowFn = origWindow })
			initWindowFn = func(root string) error {
				if tc.markerKept {
					writeResumeMarker(t, root)
				}
				return nil
			}

			cmd := newInitTestCmd()
			if err := cmd.Flags().Set("llm", "claude"); err != nil {
				t.Fatal(err)
			}
			var out, errOut strings.Builder
			cmd.SetOut(&out)
			cmd.SetErr(&errOut)
			if err := runInit(cmd, []string{projectDir}); err != nil {
				t.Fatalf("the healthy init failed: %v", err)
			}

			got, ok := readResumeWiring(projectDir)
			if tc.markerKept {
				if !ok || got != agentWiringClaude {
					t.Fatalf("a healthy completion with a standing checkpoint marker left no selection record for the resume the marker names (ok=%v, got=%q, want claude)", ok, got)
				}
				return
			}
			if ok {
				t.Fatalf("a healthy completion with no checkpoint marker left the selection record behind (got %q)", got)
			}
		})
	}
}

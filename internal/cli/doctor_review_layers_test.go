// doctor_review_layers_test.go — card t1683 (SPEC-REVIEW-GATE-REPAIR-001
// follow-up F4): the doctor warning for the two review layers being active
// together, so the same card diff is reviewed twice.
//
// Warn only when BOTH hold: the turn-end codex review gate is enabled and a
// factory lane session record exists (the card-review stage is a mandatory
// lane step with no config switch of its own). Every other state is OK, and
// an unreadable or absent config never fails the advisory check.
package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// writeReviewLayersWorkflow writes workflow.yaml under root with the given body.
func writeReviewLayersWorkflow(t *testing.T, root, body string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

const reviewGateOnYAML = "workflow:\n  codex:\n    review_gate:\n      enabled: true\n"
const reviewGateOffYAML = "workflow:\n  codex:\n    review_gate:\n      enabled: false\n"

func TestDoctorReviewLayerOverlap(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(t *testing.T, root string)
		want    uikit.CheckStatus
		wantMsg []string
	}{
		{
			name: "gate-on-and-lane-record",
			setup: func(t *testing.T, root string) {
				writeReviewLayersWorkflow(t, root, reviewGateOnYAML)
				seedFactoryRecord(t, root, "lane-session", "lane")
			},
			want:    uikit.CheckWarn,
			wantMsg: []string{"review_gate", "card-review", "record is present", "may review", "stale", "workflow.codex.review_gate.enabled"},
		},
		{
			name: "gate-off-and-lane-record",
			setup: func(t *testing.T, root string) {
				writeReviewLayersWorkflow(t, root, reviewGateOffYAML)
				seedFactoryRecord(t, root, "lane-session", "lane")
			},
			want: uikit.CheckOK,
		},
		{
			name: "gate-on-without-lane-record",
			setup: func(t *testing.T, root string) {
				writeReviewLayersWorkflow(t, root, reviewGateOnYAML)
				seedFactoryRecord(t, root, "leader-session", "leader")
			},
			want: uikit.CheckOK,
		},
		{
			name:  "config-absent",
			setup: func(t *testing.T, root string) { seedFactoryRecord(t, root, "lane-session", "lane") },
			want:  uikit.CheckOK,
		},
		{
			name: "config-unparseable",
			setup: func(t *testing.T, root string) {
				writeReviewLayersWorkflow(t, root, "workflow: [unclosed\n")
				seedFactoryRecord(t, root, "lane-session", "lane")
			},
			want: uikit.CheckOK,
		},
		{
			name:  "empty-root",
			setup: func(t *testing.T, root string) {},
			want:  uikit.CheckOK,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			tc.setup(t, root)

			check := checkReviewLayerOverlap(root, false)

			if check.Name != reviewLayerOverlapCheckName {
				t.Errorf("name = %q, want %q", check.Name, reviewLayerOverlapCheckName)
			}
			if check.Status != tc.want {
				t.Errorf("status = %v, want %v (message %q)", check.Status, tc.want, check.Message)
			}
			joined := check.Message + "\n" + check.Detail
			for _, want := range tc.wantMsg {
				if !strings.Contains(joined, want) {
					t.Errorf("message does not carry %q:\n%s", want, joined)
				}
			}
		})
	}
}

// TestDoctorReviewLayerOverlap_ConfigOrphanedWorktree — a linked worktree with
// no .moai of its own reads the gate flag from its primary checkout, exactly as
// the Stop hook does, so a lane record in that worktree still warns.
func TestDoctorReviewLayerOverlap_ConfigOrphanedWorktree(t *testing.T) {
	fx := newUntrackedFixture(t, reviewGateOnYAML)
	seedFactoryRecord(t, fx.W, "lane-session", "lane")

	check := checkReviewLayerOverlap(fx.W, false)

	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v, want warn (message %q)", check.Status, check.Message)
	}
}

// TestDoctorReviewLayerOverlap_RecordAtPrimaryRoot — the session-start writer
// puts a lane record in the primary checkout, so a doctor run from a linked
// worktree whose .moai is tracked (not config-orphaned) must look there too.
func TestDoctorReviewLayerOverlap_RecordAtPrimaryRoot(t *testing.T) {
	fx := newUntrackedFixture(t, reviewGateOnYAML)
	writeReviewLayersWorkflow(t, fx.W, reviewGateOnYAML)
	seedFactoryRecord(t, fx.P, "lane-session", "lane")

	check := checkReviewLayerOverlap(fx.W, false)

	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v, want warn (message %q)", check.Status, check.Message)
	}
}

// TestDoctorReviewLayerOverlap_UnreadablePrimaryRecords — when the primary
// checkout's record directory cannot be read, the check must not state an
// absence it never observed: it reports "unreadable, overlap not judged".
func TestDoctorReviewLayerOverlap_UnreadablePrimaryRecords(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("directory permission denial is not enforceable here")
	}
	fx := newUntrackedFixture(t, reviewGateOnYAML)
	writeReviewLayersWorkflow(t, fx.W, reviewGateOnYAML)
	seedFactoryRecord(t, fx.P, "lane-session", "lane")
	dir := factory.RuntimeStateDirForRoot(fx.P)
	if err := os.Chmod(dir, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	check := checkReviewLayerOverlap(fx.W, false)

	if check.Status != uikit.CheckOK || !strings.Contains(check.Message, "unreadable") {
		t.Errorf("status = %v message = %q, want ok with an unreadable-records message", check.Status, check.Message)
	}
}

// TestDoctorReviewLayerOverlap_SkippedRecordIsNotAbsence — a record file the
// shared reader skips (here: unreadable) must not read as "no lane record".
func TestDoctorReviewLayerOverlap_SkippedRecordIsNotAbsence(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("file permission denial is not enforceable here")
	}
	root := t.TempDir()
	writeReviewLayersWorkflow(t, root, reviewGateOnYAML)
	seedFactoryRecord(t, root, "lane-session", "lane")
	path := factory.RecordPath(root, "lane-session")
	if err := os.Chmod(path, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

	check := checkReviewLayerOverlap(root, false)

	if check.Status != uikit.CheckOK || !strings.Contains(check.Message, "unreadable") {
		t.Errorf("status = %v message = %q, want ok with an unreadable-records message", check.Status, check.Message)
	}
}

// TestDoctorReviewLayerOverlap_LaneRecordBesideMalformedFile — a malformed
// record file next to a valid lane record must not hide the lane: the lane was
// observed, so the warning stands.
func TestDoctorReviewLayerOverlap_LaneRecordBesideMalformedFile(t *testing.T) {
	root := t.TempDir()
	writeReviewLayersWorkflow(t, root, reviewGateOnYAML)
	seedFactoryRecord(t, root, "lane-session", "lane")
	bad := filepath.Join(filepath.Dir(factory.RecordPath(root, "lane-session")), "broken.json")
	if err := os.WriteFile(bad, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	check := checkReviewLayerOverlap(root, false)

	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v message = %q, want warn", check.Status, check.Message)
	}
}

// TestDoctorReviewLayerOverlap_UnidentifiablePrimary — a linked worktree whose
// primary checkout cannot be identified leaves the primary record directory
// unobserved, so the check must not state an absence.
func TestDoctorReviewLayerOverlap_UnidentifiablePrimary(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("file permission denial is not enforceable here")
	}
	fx := newUntrackedFixture(t, reviewGateOnYAML)
	writeReviewLayersWorkflow(t, fx.W, reviewGateOnYAML)
	seedFactoryRecord(t, fx.P, "lane-session", "lane")
	dotGit := filepath.Join(fx.W, ".git")
	if err := os.Chmod(dotGit, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dotGit, 0o644) })

	check := checkReviewLayerOverlap(fx.W, false)

	if check.Status != uikit.CheckOK || !strings.Contains(check.Message, "unreadable") {
		t.Errorf("status = %v message = %q, want ok with an unreadable-records message", check.Status, check.Message)
	}
}

// TestDoctorReviewLayerOverlap_Registered — the check is reachable through the
// grouped table (`moai doctor --check "Review Layer Overlap"`).
func TestDoctorReviewLayerOverlap_Registered(t *testing.T) {
	groups := runGroupedChecks(false, reviewLayerOverlapCheckName)
	for _, g := range groups {
		for _, c := range g.checks {
			if c.Name == reviewLayerOverlapCheckName {
				return
			}
		}
	}
	t.Errorf("%q not reachable via runGroupedChecks — not registered", reviewLayerOverlapCheckName)
}

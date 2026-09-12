package merge

// settings_snapshot_flow_test.go drives whole flows through the snapshot
// seam (SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001, plan.md D8): leftover
// judgement → deploy → staging → merge → settle. It pins AC-USB-006 (the base
// of a flow is the previous flow's render, in both call shapes),
// AC-USB-016 (the operator's D5 promotion outcomes), and the seam cell of
// AC-USB-008 (a failed promotion does not block).
//
// Only states between flows, or after a flow, are asserted (acceptance.md §A
// observation rule).

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

const pendingSnapshotRel = ".moai/cache/template-snapshot/claude/settings.json.pending"

type snapshotFlow struct {
	t    *testing.T
	root string
	warn strings.Builder
	out  strings.Builder
}

func newSnapshotFlowProject(t *testing.T) *snapshotFlow {
	t.Helper()
	return &snapshotFlow{t: t, root: newSnapshotProject(t)}
}

// begin is step ①: the leftover judgement that runs before any step of the
// flow removes or rewrites the live file.
func (f *snapshotFlow) begin() {
	backup.JudgeLeftoverSettingsSnapshot(f.root, &f.warn)
}

// backupUser reads the live settings.json the way the update flows do before
// their destructive steps; an absent file yields no backup.
func (f *snapshotFlow) backupUser() []FileBackup {
	data, err := os.ReadFile(filepath.Join(f.root, settingsRel))
	if err != nil {
		return nil
	}
	return []FileBackup{{Path: settingsRel, Data: data}}
}

// deploy writes render the way template.deployer does after a real write, and
// is followed by step ②, staging.
func (f *snapshotFlow) deploy(render string) {
	f.t.Helper()
	m := manifest.NewManager()
	if _, err := m.Load(f.root); err != nil {
		f.t.Fatalf("load manifest: %v", err)
	}
	writeRel(f.t, f.root, settingsRel, render)
	if err := m.Track(settingsRel, manifest.TemplateManaged, manifest.HashBytes([]byte(render))); err != nil {
		f.t.Fatalf("track: %v", err)
	}
	backup.StageDeployedSettingsSnapshot(f.root, m, &f.warn)
}

// restore is step ③ together with the merge it follows.
func (f *snapshotFlow) restore(backups []FileBackup) {
	f.t.Helper()
	if err := MergeUserFilesAndSettleSnapshot(f.root, backups, &f.out, &f.warn); err != nil {
		f.t.Fatalf("MergeUserFilesAndSettleSnapshot returned %v, want nil", err)
	}
}

// update runs a complete update flow in one call (the clean-reinstall shape).
func (f *snapshotFlow) update(render string) {
	f.begin()
	backups := f.backupUser()
	f.deploy(render)
	f.restore(backups)
}

// abort runs an update flow that stops after the deploy, before the merge.
func (f *snapshotFlow) abort(render string) {
	f.begin()
	_ = f.backupUser()
	f.deploy(render)
}

func (f *snapshotFlow) live() map[string]any {
	return decodeObject(f.t, readRel(f.t, f.root, settingsRel))
}

func (f *snapshotFlow) assertCanonical(want string) {
	f.t.Helper()
	got, err := os.ReadFile(filepath.Join(f.root, filepath.FromSlash(canonicalSnapshotRel)))
	if err != nil {
		f.t.Errorf("canonical snapshot unreadable (%v), want %s", err, want)
		return
	}
	if !bytes.Equal(got, []byte(want)) {
		f.t.Errorf("canonical snapshot = %s, want %s", got, want)
	}
}

func (f *snapshotFlow) assertPending(want string) {
	f.t.Helper()
	got, err := os.ReadFile(filepath.Join(f.root, filepath.FromSlash(pendingSnapshotRel)))
	if err != nil {
		f.t.Errorf("staging copy unreadable (%v), want %s", err, want)
		return
	}
	if !bytes.Equal(got, []byte(want)) {
		f.t.Errorf("staging copy = %s, want %s", got, want)
	}
}

func (f *snapshotFlow) assertNoPending() {
	f.t.Helper()
	if _, err := os.Stat(filepath.Join(f.root, filepath.FromSlash(pendingSnapshotRel))); err == nil {
		f.t.Errorf("staging copy still present, want none")
	}
}

func assertKeys(t *testing.T, doc map[string]any, want map[string]float64) {
	t.Helper()
	for key, value := range want {
		if got, ok := doc[key].(float64); !ok || got != value {
			t.Errorf("%s = %v, want %v (doc %v)", key, doc[key], value, doc)
		}
	}
}

// AC-USB-006 — across two cycles the second flow merges against the first
// flow's render, whether the flow runs in one call or split across a deploy
// call and a restore call (the template-sync shape).
func TestSettingsSnapshotFlow_TwoCycles_BaseIsPreviousRender(t *testing.T) {
	const r1 = `{"model":"sonnet","userOnly":false}`
	const user = `{"model":"sonnet","userOnly":true}`
	const r2 = `{"model":"haiku","userOnly":false,"newKey":1}`

	shapes := map[string]func(f *snapshotFlow, render string){
		"single_call_order": func(f *snapshotFlow, render string) { f.update(render) },
		"split_deploy_then_restore_order": func(f *snapshotFlow, render string) {
			deployPhase := func() []FileBackup {
				f.begin()
				backups := f.backupUser()
				f.deploy(render)
				return backups
			}
			restorePhase := func(backups []FileBackup) { f.restore(backups) }
			restorePhase(deployPhase())
		},
	}
	for name, run := range shapes {
		t.Run(name, func(t *testing.T) {
			f := newSnapshotFlowProject(t)
			writeRel(t, f.root, settingsRel, user)

			run(f, r1) // cycle 1: merge writes a result (D5 case 1)
			run(f, r2) // cycle 2: the user did not edit the cycle-1 result

			doc := f.live()
			if got := doc["model"]; got != "haiku" {
				t.Errorf("model = %v, want haiku", got)
			}
			if got := doc["newKey"]; got != float64(1) {
				t.Errorf("newKey = %v, want 1", got)
			}
			if got := doc["userOnly"]; got != true {
				t.Errorf("userOnly = %v, want true", got)
			}
			f.assertCanonical(r2)
		})
	}
}

// AC-USB-016 — the operator's D5 outcomes, numbered by plan.md D5 case.
func TestSettingsSnapshotFlow_PromotionRule(t *testing.T) {
	const (
		r1 = `{"a":1}`
		r2 = `{"a":2,"K":1}`
		r3 = `{"a":3,"K":1,"L":1}`
	)
	seed := func(t *testing.T, liveUser string) *snapshotFlow {
		f := newSnapshotFlowProject(t)
		writeRel(t, f.root, canonicalSnapshotRel, r1)
		if liveUser != "" {
			writeRel(t, f.root, settingsRel, liveUser)
		}
		return f
	}

	t.Run("c1_merge_writes_result_promotes", func(t *testing.T) {
		f := seed(t, `{"a":1,"u":1}`)
		f.update(r2)
		f.assertCanonical(r2)
		f.assertNoPending()
	})
	t.Run("c2_preserve_path_keeps_prior_base", func(t *testing.T) {
		f := seed(t, `{"a":`)
		f.update(r2)
		f.assertCanonical(r1)
		f.assertNoPending()

		writeRel(t, f.root, settingsRel, `{"a":1}`) // the user repairs the file
		f.update(r3)
		assertKeys(t, f.live(), map[string]float64{"a": 3, "K": 1, "L": 1})
	})
	t.Run("c3_merge_skipped_identical_promotes", func(t *testing.T) {
		f := seed(t, r2)
		f.update(r2)
		f.assertCanonical(r2)
	})
	t.Run("c4_abort_without_revert_promotes_next_flow", func(t *testing.T) {
		f := seed(t, `{"a":1}`)
		f.abort(r2)
		f.assertCanonical(r1)
		f.assertPending(r2)

		f.update(r3)
		assertKeys(t, f.live(), map[string]float64{"a": 3, "K": 1, "L": 1})
		f.assertCanonical(r3)
	})
	t.Run("c5_abort_then_revert_keeps_prior_base", func(t *testing.T) {
		f := seed(t, `{"a":1}`)
		f.abort(r2)
		writeRel(t, f.root, settingsRel, `{"a":1}`) // hand revert before the next flow
		f.assertCanonical(r1)
		f.assertPending(r2)

		f.update(r3)
		assertKeys(t, f.live(), map[string]float64{"a": 3, "K": 1, "L": 1})
		f.assertCanonical(r3)
	})
	t.Run("c6_init_writes_promotes_despite_bundle_rewrite", func(t *testing.T) {
		f := seed(t, "")
		f.begin()
		f.deploy(r2)
		writeRel(t, f.root, settingsRel, `{"a":9,"K":1}`) // autonomy bundle stand-in
		backup.SettleSettingsSnapshot(f.root, false, &f.warn)
		f.assertCanonical(r2)
	})
	t.Run("c8_update_without_user_file_promotes", func(t *testing.T) {
		f := seed(t, "")
		f.update(r2)
		f.assertCanonical(r2)
	})
}

// AC-USB-008 `promote_failure` — the flow wrote its staging copy, the
// promotion cannot replace a non-empty directory squatting the canonical path,
// and the flow still returns nil with exactly one promote-failed line.
func TestSettingsSnapshotFlow_PromoteFailureDoesNotBlock(t *testing.T) {
	f := newSnapshotFlowProject(t)
	writeRel(t, f.root, canonicalSnapshotRel+"/squatter", "x")
	writeRel(t, f.root, settingsRel, `{"a":1,"u":1}`)

	f.update(`{"a":2,"K":1}`) // f.restore fails the test on a non-nil return

	out := f.warn.String()
	if n := countPrefixedLines(out, backup.SettingsSnapshotPromoteFailedPrefix); n != 1 {
		t.Errorf("promote-failed lines = %d, want 1:\n%s", n, out)
	}
	if n := countPrefixedLines(out, backup.SettingsSnapshotWriteFailedPrefix); n != 0 {
		t.Errorf("write-failed lines = %d, want 0:\n%s", n, out)
	}
	if info, err := os.Stat(filepath.Join(f.root, filepath.FromSlash(canonicalSnapshotRel))); err != nil || !info.IsDir() {
		t.Errorf("the squatting directory at the canonical path was disturbed (err %v)", err)
	}
}

func countPrefixedLines(out, prefix string) int {
	n := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, prefix) {
			n++
		}
	}
	return n
}

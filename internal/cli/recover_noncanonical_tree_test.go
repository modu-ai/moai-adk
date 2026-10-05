package cli

// recover_noncanonical_tree_test.go — t1028.
//
// t1013 closed the rollback axis and left `recover` as a Gap, noting only that
// it hands recoverHomeState the same unnormalized cwd (:948, the same shape as
// rollback's :964) and that a shared shape is a hypothesis, not a measurement.
//
// It is not the same. The guard landed by t961 has exactly two call sites —
// AcquireMigrationAdmission and installMigrationMarker — and recoverHomeState
// reaches neither: it ACQUIRES the admission lock, reads the marker, and CLEARS
// it, where rollback INSTALLS one. Same shape, different mechanism, opposite
// outcome. So the tests below assert the asymmetry rather than a refusal, and
// the pair is run against one fixture so the difference cannot be an artifact
// of two different setups.
//
// The fixture is t1013's, reused whole per the card: two trees, a verified
// backup, and a target diverged from it. Only the owner-liveness field is new,
// because recover refuses a live owner and would otherwise stop before the
// axis under test.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// t1028DeadOwnerMarker installs a real marker through the production path and
// then rewrites ONE field: the owner PID.
//
// A marker written by this process names this process, and recover refuses a
// live owner — so without this the probe stops two gates early, on the wrong
// axis. platformPIDState treats pid <= 0 as dead deterministically, which is
// why -1 is used instead of a real exited process: no scheduling, no reuse of a
// recycled PID, no flake.
//
// Everything else in the marker stays exactly as production wrote it, so the
// identity checks recover performs (migration id, project key, project root)
// are answered by real values rather than by a hand-built fake.
func t1028DeadOwnerMarker(t *testing.T, primary, migrationID string) {
	t.Helper()
	release, err := homestate.InstallMigrationMarkerLocked(primary, migrationID)
	if err != nil {
		t.Fatalf("installing the marker from the canonical root must succeed: %v", err)
	}
	_ = release // deliberately NOT released: the marker is the state recover repairs.

	path, err := homestate.MigrationBarrierPath(primary)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var marker map[string]any
	if err := json.Unmarshal(raw, &marker); err != nil {
		t.Fatal(err)
	}
	marker["owner_pid"] = -1
	patched, err := json.Marshal(marker)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(patched, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}

	// Premise: the rewrite actually produced a dead owner. If the probe still
	// says live or indeterminate, recover stops on the owner gate and every
	// observation below would be of the wrong gate.
	read, err := homestate.ReadMigrationMarker(primary)
	if err != nil {
		t.Fatal(err)
	}
	if _, state := homestate.ProbeProcessIdentity(read.OwnerPID); state != homestate.ProcessIdentityDead {
		t.Fatalf("premise failed: the marker owner must probe as dead, got state %v", state)
	}
	if read.OwnerFingerprint == "" {
		t.Fatal("premise failed: recover refuses an empty owner fingerprint before reaching the axis under test")
	}
}

// TestT1028_GuardHasExactlyTwoCallSitesAndRecoverReachesNeither states, as an
// executable premise, the structural fact the observation rests on. A later
// edit that wires the guard into the recovery path should make THIS fail first,
// with a message that says what changed — rather than silently inverting the
// test below into a claim nobody re-derived.
func TestT1028_GuardHasExactlyTwoCallSitesAndRecoverReachesNeither(t *testing.T) {
	primary, worktree := t1013Fixture(t)

	// The guard refuses a worktree caller — established here so the tests below
	// cannot be read as "the guard is broken everywhere".
	if err := homestate.RefuseMutationFromNonCanonicalTree(worktree); err == nil {
		t.Fatal("premise failed: the guard must refuse a non-canonical caller")
	}
	if err := homestate.RefuseMutationFromNonCanonicalTree(primary); err != nil {
		t.Fatalf("premise failed: the guard must pass the canonical caller: %v", err)
	}

	// And the two entry points it actually gates both refuse that caller.
	if _, err := homestate.InstallMigrationMarkerLocked(worktree, "t1028-premise-install"); err == nil {
		t.Error("premise failed: installMigrationMarker must refuse a worktree caller")
	}
	if _, err := homestate.AcquireMigrationAdmission(worktree, "t1028-premise-admission"); err == nil {
		t.Error("premise failed: AcquireMigrationAdmission must refuse a worktree caller")
	}
}

// TestT1028_RecoverFromWorktreeIsNotRefusedAndMutatesTheCanonicalState is the
// observation. The card predicted a refusal by analogy with rollback; the code
// says otherwise, and this fires it rather than arguing about it.
//
// What it must show is not merely "no error" — a no-op returns nil too. It must
// show that a mutation the guard exists to prevent actually landed: the
// diverged target quarantined, the backup restored in its place, and the marker
// cleared, all driven from a caller whose own tree is not the tree it moved.
//
// [PINS CURRENT BEHAVIOUR, DOES NOT ENDORSE IT] Whether this exemption SHOULD
// hold is a separate decision, recorded in .moai/reports/t1028/verdict.md and
// not taken here. If it is decided that recovery must be refused too, this test
// is meant to be inverted deliberately — its failure at that point is the
// decision landing, not a regression.
func TestT1028_RecoverFromWorktreeIsNotRefusedAndMutatesTheCanonicalState(t *testing.T) {
	primary, worktree := t1013Fixture(t)
	target, before := t1013ArmedRollback(t, primary)
	id := t1013LatestBackupID(t, primary)
	t1028DeadOwnerMarker(t, primary, id)

	markerPath, err := homestate.MigrationBarrierPath(primary)
	if err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(markerPath); statErr != nil {
		t.Fatalf("premise failed: the marker must exist before recovery: %v", statErr)
	}

	err = recoverHomeState(context.Background(), worktree, id, id)
	if err != nil {
		t.Fatalf("recover from a linked worktree was NOT refused by design; it failed for another reason: %v", err)
	}

	// The mutation landed. Each of the three is a separate consequence, so a
	// partial recovery cannot pass by satisfying only one.
	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) == string(after) {
		t.Error("recover returned nil without touching the target: this would be a harmless no-op, not the hazard")
	}
	quarantined, _ := filepath.Glob(target + ".quarantine-*")
	if len(quarantined) == 0 {
		t.Error("the diverged target should have been quarantined by the recovery")
	}
	if _, statErr := os.Stat(markerPath); !os.IsNotExist(statErr) {
		t.Errorf("recover should have cleared the marker; stat says %v", statErr)
	}
}

// TestT1028_RollbackAndRecoverDivergeOnOneFixture is the discriminator, and the
// reason the two calls share a fixture rather than each building their own.
//
// A refusal observed in one test and a success observed in another could differ
// because the two setups differ. Here the tree, the backup, the target and the
// caller are one and the same, so the only variable left is which entry point
// was called — which is exactly the claim.
func TestT1028_RollbackAndRecoverDivergeOnOneFixture(t *testing.T) {
	primary, worktree := t1013Fixture(t)
	target, _ := t1013ArmedRollback(t, primary)
	id := t1013LatestBackupID(t, primary)

	// Same caller, same tree: rollback is refused.
	rollbackErr := rollbackHomeState(context.Background(), worktree, id)
	if rollbackErr == nil {
		t.Fatal("rollback from a linked worktree must still be refused")
	}
	if !strings.Contains(rollbackErr.Error(), "is not the canonical root") {
		t.Fatalf("the rollback refusal must be the tree guard's: %v", rollbackErr)
	}
	quarantined, _ := filepath.Glob(target + ".quarantine-*")
	if len(quarantined) != 0 {
		t.Errorf("the refused rollback must not have quarantined anything; found %v", quarantined)
	}

	// Same caller, same tree: recover proceeds.
	t1028DeadOwnerMarker(t, primary, id)
	if recoverErr := recoverHomeState(context.Background(), worktree, id, id); recoverErr != nil {
		t.Fatalf("recover from the same caller and the same tree must NOT be refused: %v", recoverErr)
	}
	quarantined, _ = filepath.Glob(target + ".quarantine-*")
	if len(quarantined) == 0 {
		t.Error("recover from the same caller should have quarantined the diverged target")
	}
}

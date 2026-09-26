package contract

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/mission"
)

// signedValidMissionFixture returns a decoded-equivalent contract fixture
// (AC-AP-014 Given): signed-valid per the seal consistency table, carrying
// ownership.write ["internal/foo/**"], a positive budget.operations, and the
// required top-level card field.
func signedValidMissionFixture() *Contract {
	acceptanceSHA := strings.Repeat("ab", 32)
	acCount := 18
	c := &Contract{
		SchemaVersion: SchemaVersion,
		SpecID:        "SPEC-AUTONOMY-PRECONDITION-001",
		Card:          "t1245",
		Acceptance:    &Acceptance{File: AcceptanceFile, SHA256: &acceptanceSHA, ACCount: &acCount},
		Invariants:    []string{"existing tests pass"},
		Ownership:     &Ownership{Write: []string{"internal/foo/**"}, Never: []string{"main"}},
		Approach:      "project signed contracts onto the mission validator",
		Actions:       []string{"commit", "worktree", ActionPushDevelop, "local-merge-develop"},
		Reobserve:     []string{"after acceptance change"},
		Review:        &Review{SecondModel: SecondModelNone, Human: ReviewClosureReport},
		Budget:        &Budget{Turns: 30, Operations: 40, AuditRetries: 2},
		EscalateOn: []string{
			"acceptance-change", "invariant-violation", "ownership-move",
			"new-architecture-or-api", "contradictory-evidence", "irreversible-action",
		},
		PlanAudit: &PlanAudit{Verdict: "PASS"},
		Signature: &Signature{
			SignerKind:       SignerHuman,
			Operator:         Operator{Name: "GOOS", Email: "goos@example.com"},
			SignedAt:         "2026-09-26T00:00:00Z",
			HeadSHA:          strings.Repeat("c", 40),
			ContractSHA256:   strings.Repeat("1", 64),
			AcceptanceSHA256: acceptanceSHA,
			Method:           MethodInteractiveTTY,
		},
	}
	seal, err := ComputeSeal(*c.Signature)
	if err != nil {
		panic("fixture seal: " + err.Error())
	}
	c.Signature.Seal = seal
	return c
}

// projectFixture projects the fixture and seals the result through mission's
// own sealer; the sealer is where incomplete_contract lives, so a sealed
// return is the proof limb 1 asks for.
func projectFixture(t *testing.T, c *Contract) mission.SealedContract {
	t.Helper()
	mc, err := ProjectToMission(c)
	if err != nil {
		t.Fatalf("ProjectToMission: unexpected error: %v", err)
	}
	sealed, err := mission.SealMissionContract(mc)
	if err != nil {
		t.Fatalf("mission.SealMissionContract: %v", err)
	}
	return sealed
}

// validDecisionInputs builds the snapshot/decision pair that passes every
// check in mission.ValidateMissionDecision, so each limb-2 fixture differs
// from it in exactly one field mission alone rejects.
func validDecisionInputs(sealed mission.SealedContract, now time.Time) (mission.MissionSnapshot, mission.Decision) {
	testSHA := strings.Repeat("cd", 32)
	evidence := map[string]string{
		"fresh_snapshot": testSHA,
		"scope":          "internal/foo",
		"tests":          testSHA,
	}
	snapshot := mission.MissionSnapshot{
		MissionID:        sealed.Contract.MissionID,
		ContractHash:     sealed.Hash,
		PolicyVersion:    sealed.Version,
		SnapshotHash:     strings.Repeat("e", 64),
		EvidenceRevision: 7,
		Evidence:         evidence,
		State:            mission.StateRunning,
		OperationsUsed:   0,
	}
	decision := mission.Decision{
		DecisionID:       "decision-1",
		MissionID:        sealed.Contract.MissionID,
		ContractHash:     sealed.Hash,
		PolicyVersion:    sealed.Version,
		SnapshotHash:     snapshot.SnapshotHash,
		EvidenceRevision: 7,
		Action:           mission.ActionCommit,
		Targets:          []string{"internal/foo/x.go"},
		RequiredEvidence: mission.RequiredEvidenceForAction(mission.ActionCommit),
		Evidence:         evidence,
		ExpiresAt:        now.Add(time.Hour),
	}
	return snapshot, decision
}

// TestContractProjectsOntoMissionValidator is the AC-AP-014 test. Its five
// subtests are the criterion's five limbs, in order.
func TestContractProjectsOntoMissionValidator(t *testing.T) {
	fixture := signedValidMissionFixture()
	projected, err := ProjectToMission(fixture)

	t.Run("limb 1 signed-valid contract is not refused incomplete_contract", func(t *testing.T) {
		if err != nil {
			t.Fatalf("ProjectToMission: unexpected error: %v", err)
		}
		if !projected.Approved {
			t.Fatalf("projected contract Approved = false; the signature state must populate it")
		}
		if projected.ResourceLimits.MaxOperations <= 0 {
			t.Fatalf("projected MaxOperations = %d; must be positive", projected.ResourceLimits.MaxOperations)
		}
		sealed, sealErr := mission.SealMissionContract(projected)
		if sealErr != nil {
			t.Fatalf("mission.SealMissionContract refused the projected contract: %v", sealErr)
		}
		if sealed.Hash == "" {
			t.Fatalf("sealed contract carries an empty hash")
		}
	})

	t.Run("limb 2 five refusal reasons are produced by mission.ValidateMissionDecision", func(t *testing.T) {
		sealed := projectFixture(t, fixture)
		now := time.Now()
		baseSnapshot, baseDecision := validDecisionInputs(sealed, now)

		receipt, err := mission.ValidateMissionDecision(sealed, baseSnapshot, baseDecision, now)
		if err != nil {
			t.Fatalf("baseline decision refused: %v", err)
		}
		if receipt.State != mission.ReceiptPrepared {
			t.Fatalf("baseline verdict State = %q, want prepared", receipt.State)
		}

		cases := []struct {
			name    string
			mutate  func(s *mission.MissionSnapshot, d *mission.Decision)
			refusal string
		}{
			{"mission_not_running", func(s *mission.MissionSnapshot, _ *mission.Decision) { s.State = mission.StateDraft }, "mission decision: mission_not_running"},
			{"mission_mismatch", func(s *mission.MissionSnapshot, _ *mission.Decision) { s.MissionID = "SPEC-OTHER-001" }, "mission decision: mission_mismatch"},
			{"policy_mismatch", func(_ *mission.MissionSnapshot, d *mission.Decision) { d.PolicyVersion = "contract-v0" }, "mission decision: policy_mismatch"},
			{"stale_snapshot", func(_ *mission.MissionSnapshot, d *mission.Decision) { d.SnapshotHash = strings.Repeat("f", 64) }, "mission decision: stale_snapshot"},
			{"expired_decision", func(_ *mission.MissionSnapshot, d *mission.Decision) { d.ExpiresAt = now.Add(-time.Minute) }, "mission decision: expired_decision"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				snapshot, decision := baseSnapshot, baseDecision
				tc.mutate(&snapshot, &decision)
				_, err := mission.ValidateMissionDecision(sealed, snapshot, decision, now)
				if err == nil {
					t.Fatalf("%s: expected a refusal, got a prepared verdict", tc.name)
				}
				if err.Error() != tc.refusal {
					t.Fatalf("refusal = %q, want %q", err.Error(), tc.refusal)
				}
			})
		}
	})

	t.Run("limb 3 slash-star-star translates to a prefix that neither narrows nor widens", func(t *testing.T) {
		sealed := projectFixture(t, fixture)
		now := time.Now()
		baseSnapshot, baseDecision := validDecisionInputs(sealed, now)

		inside := baseDecision
		inside.Targets = []string{"internal/foo/x.go"}
		if _, err := mission.ValidateMissionDecision(sealed, baseSnapshot, inside, now); err != nil {
			t.Fatalf("target internal/foo/x.go refused: %v (the /** to prefix translation narrowed the scope)", err)
		}

		outside := baseDecision
		outside.Targets = []string{"internal/foobar/x.go"}
		_, err := mission.ValidateMissionDecision(sealed, baseSnapshot, outside, now)
		if err == nil || err.Error() != "mission decision: scope_expansion" {
			t.Fatalf("target internal/foobar/x.go: err = %v, want scope_expansion (the translation widened the scope)", err)
		}
	})

	t.Run("limb 4 inner wildcard and unmapped fields fail closed naming the field; card does not trigger", func(t *testing.T) {
		inner := signedValidMissionFixture()
		inner.Ownership.Write = []string{"internal/*/x.go"}
		_, err := ProjectToMission(inner)
		if err == nil {
			t.Fatalf("inner-wildcard glob internal/*/x.go: expected a fail-closed error")
		}
		for _, want := range []string{"ownership.write", "internal/*/x.go"} {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error %q does not name %q", err.Error(), want)
			}
		}

		// The projection walks the live struct's field inventory and fails
		// closed on any key neither projected nor enumerated. Go structs are
		// closed, so a genuinely unmapped field is only constructible at the
		// key-list level; the projection's own inventory check is exercised
		// through that list with an unmapped key appended.
		keys := contractInventoryKeys(fixture)
		if unmapped := unmappedProjectionKeys(keys); len(unmapped) != 0 {
			t.Fatalf("schema_version 1 inventory has unmapped keys: %v", unmapped)
		}
		withUnmapped := append(slices.Clone(keys), "future_field")
		unmapped := unmappedProjectionKeys(withUnmapped)
		if !slices.Equal(unmapped, []string{"future_field"}) {
			t.Fatalf("unmappedProjectionKeys = %v, want [future_field]", unmapped)
		}

		// card is required by the schema and has no mission counterpart; it is
		// on the deliberately-not-projected list and must not fail anything.
		if fixture.Card == "" {
			t.Fatalf("fixture must carry the required card field")
		}
		if !slices.Contains(DeliberatelyNotProjected, "card") {
			t.Fatalf("card is missing from the deliberately-not-projected list %v", DeliberatelyNotProjected)
		}
		if _, err := ProjectToMission(fixture); err != nil {
			t.Fatalf("contract carrying card: unexpected error: %v", err)
		}
	})

	t.Run("limb 5 internal/mission exported surface unchanged from the M4-start baseline", func(t *testing.T) {
		_, thisFile, _, ok := runtime.Caller(0)
		if !ok {
			t.Fatalf("runtime.Caller failed; cannot locate the repository root")
		}
		root := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))

		cmd := exec.Command("go", "doc", "-all", "./internal/mission")
		cmd.Dir = root
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("go doc -all ./internal/mission: %v (stderr: %s)", err, stderr.String())
		}

		baseline, err := os.ReadFile(filepath.Join(root, "internal", "contract", "testdata", "mission_surface_baseline.txt"))
		if err != nil {
			t.Fatalf("read baseline: %v", err)
		}
		if !bytes.Equal(baseline, stdout.Bytes()) {
			t.Fatalf("internal/mission exported surface changed since the M4-start baseline (sha256 24d4fdf3...); regenerate no golden files by hand — internal/mission is a sealed validator. First divergence: %s", firstDivergence(baseline, stdout.Bytes()))
		}
	})
}

// firstDivergence returns the first line where got leaves want, for a bounded
// failure message.
func firstDivergence(want, got []byte) string {
	wantLines := strings.SplitAfter(string(want), "\n")
	gotLines := strings.SplitAfter(string(got), "\n")
	for i := 0; i < len(wantLines) && i < len(gotLines); i++ {
		if wantLines[i] != gotLines[i] {
			return "line " + reflect.ValueOf(i).String() + ": want " + strings.TrimSpace(wantLines[i]) + ", got " + strings.TrimSpace(gotLines[i])
		}
	}
	if len(wantLines) != len(gotLines) {
		return "line count differs"
	}
	return "unknown"
}

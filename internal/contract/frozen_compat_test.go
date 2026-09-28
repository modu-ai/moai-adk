package contract

import (
	"slices"
	"testing"
)

// TestVerify_SignedUnderPriorFrozenSetStillVerifies measures the backward
// compatibility of SPEC-INSTRUCTION-FILES-UNIFY-001's frozen-set growth
// (REQ-IFU-013 cascade): FrozenInstructionFiles gained AGENTS.md and
// AGENTS.local.md. A contract signed while the set held only the two CLAUDE
// basenames must still verify signed-valid under the grown set — the set feeds
// the DERIVED frozen_files report, never the signed digest — and the only
// observable change is that frozen_files gains exactly the two new globs.
func TestVerify_SignedUnderPriorFrozenSetStillVerifies(t *testing.T) {
	grown := slices.Clone(FrozenInstructionFiles)
	prior := []string{"CLAUDE.md", "CLAUDE.local.md"}
	t.Cleanup(func() { FrozenInstructionFiles = grown })

	body := renderFixture(derivedOpts())

	// Sign and verify while the prior set is in force.
	FrozenInstructionFiles = prior
	signed := signFixture(body)
	before := Verify(derivedInputs(signed))

	// Verify the SAME signed bytes under the grown set.
	FrozenInstructionFiles = grown
	after := Verify(derivedInputs(signed))

	for label, r := range map[string]Report{"prior set": before, "grown set": after} {
		if !r.Valid || r.State != StateSignedValid || len(r.Reasons) != 0 {
			t.Fatalf("%s: state=%s valid=%v reasons=%v, want signed-valid with no reasons", label, r.State, r.Valid, r.Reasons)
		}
	}
	if before.ContractSHA256 != after.ContractSHA256 || before.RecordedContractSHA256 != after.RecordedContractSHA256 {
		t.Errorf("digest moved with the frozen set: before=%s/%s after=%s/%s",
			before.ContractSHA256, before.RecordedContractSHA256, after.ContractSHA256, after.RecordedContractSHA256)
	}

	var added []string
	for _, g := range after.FrozenFiles {
		if !slices.Contains(before.FrozenFiles, g) {
			added = append(added, g)
		}
	}
	for _, g := range before.FrozenFiles {
		if !slices.Contains(after.FrozenFiles, g) {
			t.Errorf("frozen_files lost %q under the grown set", g)
		}
	}
	if want := []string{"**/AGENTS.local.md", "**/AGENTS.md"}; !slices.Equal(added, want) {
		t.Errorf("frozen_files gained %v, want exactly %v (before=%v after=%v)", added, want, before.FrozenFiles, after.FrozenFiles)
	}
	t.Logf("prior-set frozen_files = %v", before.FrozenFiles)
	t.Logf("grown-set frozen_files = %v", after.FrozenFiles)
}

package tiers

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDrift_EmptyValidator_StatusUnknown exercises REQ-NS3-002: a contract
// with no validator_command keeps DriftStatus=unknown after running drift
// checks (no command to run → no signal).
func TestDrift_EmptyValidator_StatusUnknown(t *testing.T) {
	root := t.TempDir()
	c := ContractNode{
		Identifier:       "no-validator",
		ContractKind:     ContractKindSchema,
		ContractPath:     "x",
		ValidatorCommand: "",
	}
	status := checkContractDrift(root, c)
	if status != DriftUnknown {
		t.Errorf("status = %q; want %q", status, DriftUnknown)
	}
}

// TestDrift_PassingValidator_Aligned exercises REQ-NS3-002: a validator
// that exits 0 yields DriftStatus=aligned.
func TestDrift_PassingValidator_Aligned(t *testing.T) {
	root := t.TempDir()
	c := ContractNode{
		Identifier:       "ok",
		ContractKind:     ContractKindSchema,
		ContractPath:     "x",
		ValidatorCommand: "true",
	}
	status := checkContractDrift(root, c)
	if status != DriftAligned {
		t.Errorf("status = %q; want %q", status, DriftAligned)
	}
}

// TestDrift_FailingValidator_Drifted exercises REQ-NS3-002: a validator
// that exits non-zero yields DriftStatus=drifted.
func TestDrift_FailingValidator_Drifted(t *testing.T) {
	root := t.TempDir()
	c := ContractNode{
		Identifier:       "bad",
		ContractKind:     ContractKindSchema,
		ContractPath:     "x",
		ValidatorCommand: "false",
	}
	status := checkContractDrift(root, c)
	if status != DriftCollapsed {
		t.Errorf("status = %q; want %q", status, DriftCollapsed)
	}
}

// TestDrift_RunDriftChecks_MutatesStatuses verifies runDriftChecks writes
// aligned/drifted statuses back onto a copy of each ContractNode.
func TestDrift_RunDriftChecks_MutatesStatuses(t *testing.T) {
	root := t.TempDir()
	nodes := []ContractNode{
		{Identifier: "a", ValidatorCommand: "true"},
		{Identifier: "b", ValidatorCommand: "false"},
		{Identifier: "c", ValidatorCommand: ""},
	}
	anyDrifted := runDriftChecks(root, nodes)
	if !anyDrifted {
		t.Errorf("anyDrifted = false; want true (b drifted)")
	}
	want := map[string]DriftStatus{"a": DriftAligned, "b": DriftCollapsed, "c": DriftUnknown}
	for _, n := range nodes {
		if got := want[n.Identifier]; got != n.DriftStatus {
			t.Errorf("node %q status = %q; want %q", n.Identifier, n.DriftStatus, got)
		}
	}
}

// TestDrift_CIMode_Indicator exercises REQ-NS3-002 CI gate: with
// TIER_CONTRACT_CI=1 set, driftCIMode returns true; unset returns false.
// The non-zero exit decision is at the CLI layer; this helper just exposes
// the env flag.
func TestDrift_CIMode_Indicator(t *testing.T) {
	t.Setenv("TIER_CONTRACT_CI", "1")
	if !driftCIMode() {
		t.Errorf("driftCIMode = false; want true with TIER_CONTRACT_CI=1")
	}
	t.Setenv("TIER_CONTRACT_CI", "")
	if driftCIMode() {
		t.Errorf("driftCIMode = true; want false with TIER_CONTRACT_CI unset")
	}
	t.Setenv("TIER_CONTRACT_CI", "0")
	if driftCIMode() {
		t.Errorf("driftCIMode = true; want false with TIER_CONTRACT_CI=0")
	}
}

// TestDrift_ValidatorCommand_MissingBinary_FailOpen ensures a missing
// validator binary (or any exec error) is treated as drifted, not a panic
// or propagated error (REQ-NS3-020 fail-open).
func TestDrift_ValidatorCommand_MissingBinary_FailOpen(t *testing.T) {
	root := t.TempDir()
	c := ContractNode{
		Identifier:       "missing-binary",
		ValidatorCommand: "/this/binary/does/not/exist",
	}
	// Should not panic; status collapses to drifted (validator unavailable
	// ≡ drifted per design.md §8 open question).
	status := checkContractDrift(root, c)
	if status != DriftCollapsed {
		t.Errorf("status = %q; want %q (fail-open: missing binary → drifted)", status, DriftCollapsed)
	}
}

// TestDrift_NeverBlocksEmission verifies that drift findings do NOT block
// tier emission (REQ-NS3-002): Enrich completes and emits tiers.json even
// when a contract's drift check fails. The CI exit decision is at the CLI
// layer, not the emission layer.
func TestDrift_NeverBlocksEmission(t *testing.T) {
	root := t.TempDir()
	registry := "contracts:\n  - identifier: drift-contract\n    contract_kind: schema\n    contract_path: x\n    validator_command: 'false'\n"
	writeRegistry(t, root, []byte(registry))
	// Set CI mode — even so, emission MUST proceed.
	t.Setenv("TIER_CONTRACT_CI", "1")
	if err := Enrich(root); err != nil {
		t.Fatalf("Enrich returned error under drift+CI: %v (REQ-NS3-002 fail-open to graph)", err)
	}
	tiersPath := filepath.Join(root, ".moai", "project", "navigator", "tiers.json")
	if _, err := os.Stat(tiersPath); err != nil {
		t.Fatalf("tiers.json not emitted under drift: %v (REQ-NS3-002 graph fail-open)", err)
	}
}

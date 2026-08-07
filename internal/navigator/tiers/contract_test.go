package tiers

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestContract_RegistryEmpty_ZeroNodes exercises AC-NS3-003: an absent OR
// empty contracts.yaml registry degrades gracefully to zero contract nodes.
// The run exits 0 and surfaces no user-facing error.
func TestContract_RegistryEmpty_ZeroNodes(t *testing.T) {
	t.Run("absent registry", func(t *testing.T) {
		root := t.TempDir()
		nodes, err := enumerateContracts(root)
		if err != nil {
			t.Fatalf("absent registry returned error: %v", err)
		}
		if len(nodes) != 0 {
			t.Errorf("absent registry emitted %d nodes; want 0", len(nodes))
		}
	})

	t.Run("empty registry", func(t *testing.T) {
		root := t.TempDir()
		writeRegistry(t, root, []byte("# empty registry\n\n"))
		nodes, err := enumerateContracts(root)
		if err != nil {
			t.Fatalf("empty registry returned error: %v", err)
		}
		if len(nodes) != 0 {
			t.Errorf("empty registry emitted %d nodes; want 0", len(nodes))
		}
	})
}

// TestContract_RegistryEntries_Emitted exercises AC-NS3-001 + AC-NS3-003:
// each registry triple `{contract_kind, contract_path, validator_command}`
// produces a ContractNode whose kind is recognized from the additive enum
// (schema / allowlist / openapi). DriftStatus starts at unknown.
func TestContract_RegistryEntries_Emitted(t *testing.T) {
	root := t.TempDir()
	registry := `# Tier 0 contract registry (test fixture)
contracts:
  - identifier: nav-graph-schema
    contract_kind: schema
    contract_path: .moai/project/navigator/nav-graph.json
    validator_command: go test -run ByteIdentical ./internal/navigator/sync/...
  - identifier: hook-input-schema
    contract_kind: schema
    contract_path: internal/hook/hook_input.schema.json
    validator_command: go test ./internal/hook/...
  - identifier: cli-flag-schema
    contract_kind: schema
    contract_path: internal/cli/flags.go
    validator_command: go test ./internal/cli/...
  - identifier: tauri-allowlist
    contract_kind: allowlist
    contract_path: src-tauri/tauri.conf.json
    validator_command: npm run check:allowlist
  - identifier: openapi-pets
    contract_kind: openapi
    contract_path: openapi.yaml
    validator_command: npx @redocly/cli lint openapi.yaml
`
	writeRegistry(t, root, []byte(registry))

	nodes, err := enumerateContracts(root)
	if err != nil {
		t.Fatalf("registry returned error: %v", err)
	}
	if got, want := len(nodes), 5; got != want {
		t.Fatalf("emitted %d nodes; want %d", got, want)
	}

	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Identifier < nodes[j].Identifier })

	if nodes[0].Identifier != "cli-flag-schema" {
		t.Errorf("first sorted node identifier = %q; want cli-flag-schema", nodes[0].Identifier)
	}
	// DriftStatus starts at unknown (REQ-NS3-001); drift check runs separately.
	if nodes[0].DriftStatus != DriftUnknown {
		t.Errorf("DriftStatus = %q; want %q", nodes[0].DriftStatus, DriftUnknown)
	}
	// All 3 additive kinds are recognized (AC-NS3-003).
	kinds := map[ContractKind]bool{}
	for _, n := range nodes {
		kinds[n.ContractKind] = true
	}
	for _, want := range []ContractKind{ContractKindSchema, ContractKindAllowlist, ContractKindOpenAPI} {
		if !kinds[want] {
			t.Errorf("contract_kind %q not present in registry output", want)
		}
	}
}

// TestContract_RegistryMalformed_FailOpen exercises AC-NS3-020 fail-open: a
// malformed registry yields 0 nodes and no error.
func TestContract_RegistryMalformed_FailOpen(t *testing.T) {
	root := t.TempDir()
	// Unparseable YAML-ish content with a tab indent (YAML rejects).
	writeRegistry(t, root, []byte("contracts:\n\t- identifier: bad\n\tcontract_kind: schema\n"))
	nodes, err := enumerateContracts(root)
	if err != nil {
		t.Fatalf("malformed registry returned error: %v (fail-open: should be nil)", err)
	}
	if len(nodes) != 0 {
		t.Errorf("malformed registry emitted %d nodes; want 0 (fail-open)", len(nodes))
	}
}

// writeRegistry writes the contracts.yaml registry under root's
// .moai/project/blueprint/ directory.
func writeRegistry(t *testing.T, root string, content []byte) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "project", "blueprint")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "contracts.yaml"), content, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestContract_KindsAreAdditive ensures the 3 ContractKind values are stable
// and the registry does NOT emit a node whose kind is outside the additive
// enum (forward-compat: an unknown kind degrades to the raw string but the
// enum is the SSOT — registry entries with unknown kinds are still emitted
// with their raw kind preserved).
func TestContract_KindsAreAdditive(t *testing.T) {
	root := t.TempDir()
	registry := "contracts:\n  - identifier: x\n    contract_kind: futurekind\n    contract_path: a\n    validator_command: b\n"
	writeRegistry(t, root, []byte(registry))
	nodes, err := enumerateContracts(root)
	if err != nil {
		t.Fatalf("registry returned error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("emitted %d nodes; want 1 (unknown kinds degrade but still emit)", len(nodes))
	}
	// Unknown kinds preserve the raw string (forward-compat).
	if string(nodes[0].ContractKind) != "futurekind" {
		t.Errorf("ContractKind = %q; want raw \"futurekind\"", nodes[0].ContractKind)
	}
	// The 3 known enum values do NOT collide.
	for _, known := range []string{"schema", "allowlist", "openapi"} {
		if strings.Contains(string(nodes[0].ContractKind), known) {
			t.Errorf("raw futurekind collided with enum %q", known)
		}
	}
}

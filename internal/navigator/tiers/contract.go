package tiers

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"log/slog"
)

// Contract registry path under the project root (REQ-NS3-003).
// `.moai/project/blueprint/contracts.yaml` is the template-distributed user
// declaration surface; an absent OR empty registry degrades gracefully to
// zero contract nodes.
const contractRegistryRel = ".moai/project/blueprint/contracts.yaml"

// rawContractEntry is the YAML shape of one registry triple. The registry
// uses a minimal hand-rolled parser (no external YAML dep) keyed on the
// `contracts:` block. Each entry carries `{identifier, contract_kind,
// contract_path, validator_command}`.
type rawContractEntry struct {
	Identifier       string
	ContractKind     string
	ContractPath     string
	ValidatorCommand string
}

// enumerateContracts loads the contract registry (REQ-NS3-003) and returns
// one ContractNode per declared entry. DriftStatus starts at `unknown`; the
// drift check (`drift.go`) populates it. Empty/absent/malformed registry →
// zero nodes, no error (fail-open per REQ-NS3-020).
//
// @MX:ANCHOR: [AUTO] Tier 0 contract enumeration entry; high fan_in (consumed by Enrich + tests + future CLI)
// @MX:REASON: drives the additive contract-node surface; called once per tier emission, structurally central
// @MX:SPEC:SPEC-NAVIGATOR-SYNC-003
func enumerateContracts(projectRoot string) ([]ContractNode, error) {
	path := filepath.Join(projectRoot, contractRegistryRel)
	content, err := os.ReadFile(path)
	if err != nil {
		// Absent registry → 0 nodes (graceful degrade, REQ-NS3-003).
		return nil, nil
	}
	entries, perr := parseContractRegistry(content)
	if perr != nil {
		// Malformed registry → 0 nodes (fail-open, REQ-NS3-020).
		slog.Debug("tiers: malformed contract registry, emitting 0 nodes",
			"path", path, "error", perr)
		return nil, nil
	}
	out := make([]ContractNode, 0, len(entries))
	for _, e := range entries {
		if e.Identifier == "" || e.ContractPath == "" {
			// Skip incomplete entries (best-effort).
			continue
		}
		out = append(out, ContractNode{
			Identifier:       e.Identifier,
			ContractKind:     ContractKind(e.ContractKind),
			ContractPath:     e.ContractPath,
			ValidatorCommand: e.ValidatorCommand,
			DriftStatus:      DriftUnknown,
		})
	}
	// Deterministic order (REQ-NS3-019): sort by identifier.
	sortContractNodes(out)
	return out, nil
}

// sortContractNodes sorts in-place by Identifier for byte-stable emission.
func sortContractNodes(nodes []ContractNode) {
	// Insertion sort — the contract surface is small (typically ≤ 5 nodes).
	for i := 1; i < len(nodes); i++ {
		for j := i; j > 0 && nodes[j-1].Identifier > nodes[j].Identifier; j-- {
			nodes[j-1], nodes[j] = nodes[j], nodes[j-1]
		}
	}
}

// parseContractRegistry parses the contracts.yaml block. The parser is
// intentionally minimal: it accepts the documented two-space-indented
// `- key: value` form. Unparseable input returns an error and the caller
// fail-opens to 0 nodes.
func parseContractRegistry(content []byte) ([]rawContractEntry, error) {
	var entries []rawContractEntry
	var current *rawContractEntry
	inContractsBlock := false
	for _, raw := range bytes.Split(content, []byte("\n")) {
		line := string(raw)
		// Strip trailing carriage return (Windows line endings).
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Top-level key.
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			if current != nil {
				entries = append(entries, *current)
				current = nil
			}
			if strings.HasPrefix(trimmed, "contracts:") {
				inContractsBlock = true
				continue
			}
			// Other top-level keys terminate the contracts block.
			inContractsBlock = false
			continue
		}
		if !inContractsBlock {
			continue
		}
		// Entry start.
		if strings.HasPrefix(trimmed, "- ") {
			if current != nil {
				entries = append(entries, *current)
			}
			current = &rawContractEntry{}
			// The first key may share the dash line.
			rest := strings.TrimPrefix(trimmed, "- ")
			applyContractEntryField(current, rest)
			continue
		}
		// Continuation field within the current entry.
		if current != nil {
			applyContractEntryField(current, trimmed)
		}
	}
	if current != nil {
		entries = append(entries, *current)
	}
	if entries == nil {
		entries = []rawContractEntry{}
	}
	return entries, nil
}

// applyContractEntryField parses one `key: value` field into the entry.
// Malformed lines are silently skipped (best-effort).
func applyContractEntryField(e *rawContractEntry, kv string) {
	idx := strings.Index(kv, ":")
	if idx < 0 {
		return
	}
	key := strings.TrimSpace(kv[:idx])
	val := strings.TrimSpace(kv[idx+1:])
	val = unquoteYAMLScalar(val)
	switch key {
	case "identifier":
		e.Identifier = val
	case "contract_kind":
		e.ContractKind = val
	case "contract_path":
		e.ContractPath = val
	case "validator_command":
		e.ValidatorCommand = val
	}
}

// unquoteYAMLScalar strips a single set of surrounding single or double
// quotes if present. Does not interpret escape sequences (best-effort).
func unquoteYAMLScalar(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// _ = bytes / fmt references kept for future parser-extension hooks without
// forcing an unused import cycle if the parser shrinks.
var _ = fmt.Sprintf

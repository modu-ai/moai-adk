package delegationmap

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// delegationMapRel is the delegation map's path relative to a project root.
// This file is the ONLY place in the package that names it, and it names it
// exactly once: every other reference resolves through DefaultMapPath, so a
// reader can confirm the read-only claim by looking here alone.
const delegationMapRel = ".moai/config/sections/delegation.yaml"

// DefaultMapPath returns the delegation map path under a project root.
func DefaultMapPath(projectRoot string) string {
	return filepath.Join(projectRoot, delegationMapRel)
}

// DefaultLedgerPath returns the routing ledger path under a project root. The
// file name comes from the producer package rather than a literal declared
// here — an independently-declared path would be free to drift from the file
// the producer actually writes (REQ-HLA-001).
func DefaultLedgerPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "state", routingLedgerFileName)
}

// delegationFile is the minimal shape this analyzer needs from the map. Only
// the per-subcommand agent designations are decoded; skills, domain skills, and
// the learning block are deliberately left out, because decoding a field is the
// first step toward writing it back.
type delegationFile struct {
	Delegation struct {
		Subcommands map[string]struct {
			Agents []string `yaml:"agents"`
		} `yaml:"subcommands"`
	} `yaml:"delegation"`
}

// ReadDelegationMap loads the designated agent list per subcommand.
//
// The file is opened READ-ONLY and is never written — not to reformat it, not
// to add a comment, not to apply a proposal. Applying an amendment is the
// Tier-4 approval gate's job (spec.md §G); this package produces the proposal
// that gate consumes and stops there.
//
// An absent map is reported as os.ErrNotExist so the caller can distinguish
// "no map" from "unreadable map" — the two warrant different result reasons.
func ReadDelegationMap(path string) (map[string][]string, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- caller-supplied config path, read-only
	if err != nil {
		return nil, err
	}
	var f delegationFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("delegationmap: parse %q: %w", path, err)
	}
	out := make(map[string][]string, len(f.Delegation.Subcommands))
	for name, entry := range f.Delegation.Subcommands {
		agents := make([]string, len(entry.Agents))
		copy(agents, entry.Agents)
		out[name] = agents
	}
	return out, nil
}

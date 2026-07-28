// SPEC: SPEC-HARNESS-LOOP-REPAIR-001 REQ-HLR-001 (proposal layout contract).
//
// layout.go is the single accessor every producer and consumer resolves
// proposals through. It lives next to the producer (scaffolder.go) so the
// on-disk layout is defined by the code that writes it.
//
// The layout is one DIRECTORY per draft:
//
//	<projectRoot>/.moai/harness/proposals/<DRAFT-ID>/
//	  ├── spec.md        # human body, downstream manager-spec authoring target
//	  └── proposal.json  # structured metadata (this file defines a draft)
//
// Consumers previously re-derived a flat `<DRAFT-ID>.json` path and filtered
// directory entries with `!e.IsDir()`, which excluded every generated draft by
// construction — silently, with no error surfaced. Routing all resolution
// through this file makes that class of drift a compile-time concern rather
// than a runtime no-op.
package proposalgen

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ProposalDirRel is the canonical proposals directory, relative to a project
// root. This is the ONE declaration of the path literal; consumers reach it via
// ProposalDir rather than redeclaring their own copy.
const ProposalDirRel = ".moai/harness/proposals"

// MetadataFileName is the per-draft metadata file. A draft directory is only a
// draft when this file is present — a bare directory is an incomplete artifact,
// not a pending proposal.
const MetadataFileName = "proposal.json"

// ProposalDir returns the absolute proposals directory for a project root.
func ProposalDir(projectRoot string) string {
	return filepath.Join(projectRoot, ProposalDirRel)
}

// ListDraftIDs returns the draft IDs present in proposalDir, sorted so callers
// get a stable selection order (the apply verb takes the first entry).
//
// A missing proposals directory yields an empty slice and no error — this
// preserves the producer's graceful no-op contract, under which the directory
// is not created until the first draft is written. Genuine read failures
// (permissions, an I/O fault) are returned so a caller can distinguish "no
// drafts" from "could not look".
func ListDraftIDs(proposalDir string) ([]string, error) {
	entries, err := os.ReadDir(proposalDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("proposalgen layout: read %q: %w", proposalDir, err)
	}

	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		metadata := filepath.Join(proposalDir, e.Name(), MetadataFileName)
		if _, statErr := os.Stat(metadata); statErr != nil {
			continue // incomplete draft — a directory without its metadata file
		}
		ids = append(ids, e.Name())
	}
	sort.Strings(ids)
	return ids, nil
}

// ProposalPath returns the proposal.json path for draftID under proposalDir.
//
// draftID is validated as a plain identifier: a path separator or a parent
// reference would let a caller-supplied ID escape the proposals directory, so
// both are rejected rather than cleaned.
// Validation errors carry no package-name prefix: they surface directly in CLI
// diagnostics, and the calling verb supplies its own context.
func ProposalPath(proposalDir, draftID string) (string, error) {
	if draftID == "" {
		return "", errors.New("empty proposal ID")
	}
	if strings.ContainsAny(draftID, `/\`) || strings.Contains(draftID, "..") {
		return "", fmt.Errorf("invalid proposal ID %q (path traversal not allowed)", draftID)
	}
	return filepath.Join(proposalDir, draftID, MetadataFileName), nil
}

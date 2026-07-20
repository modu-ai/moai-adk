// Package curator provides the typed managed-block writer for the self-
// evolving harness Curator (Loop 2 / write layer).
//
// The typed writer generalizes the mechanism previously hard-coded in
// internal/harness/layer3.go InjectMarker into a BlockType-keyed registry.
// Each BlockType maps to its heading + HTML-comment start/end marker pair.
// WriteManagedBlock locates the marker block for the given type, replaces
// its content atomically (or appends a fresh block when absent), and writes
// the file back with pre-block and post-block bytes preserved verbatim.
//
// Subagent boundary (C-HRA-008 / REQ-HEV2-035): this package MUST NOT call
// AskUserQuestion or reference mcp__askuser. L5 approval is routed through
// the orchestrator; the Curator returns proposal artifacts only.
package curator

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// BlockType enumerates the managed-block types the typed writer can
// manipulate. Each value maps to a markerRegistry entry that carries the
// heading + marker contract for that type.
type BlockType int

const (
	// BlockTypeLearnedWorkflow is the MOAI:LEARNED-WORKFLOW digest block
	// (Tier 4, CLAUDE.md always-loaded summary surface).
	BlockTypeLearnedWorkflow BlockType = iota
	// BlockTypeHarnessGenerated is the legacy
	// "## Project-Specific Configuration (Harness-Generated)" block produced
	// by the original InjectMarker. Preserved for backward compatibility so
	// existing callers continue to work through the typed writer byte-identical.
	BlockTypeHarnessGenerated
	// BlockTypeLearnedLocal is the MOAI:LEARNED-WORKFLOW-LOCAL append-only
	// section (Tier 3, CLAUDE.local.md permanent-record surface). Distinct from
	// the digest block: this surface appends only (no update/delete of existing
	// entries) and deduplicates on ledger_key (REQ-HEV2-012..014).
	BlockTypeLearnedLocal
)

// blockSpec is the per-BlockType marker registry entry.
type blockSpec struct {
	heading     string // the ## heading line (without trailing newline)
	startPrefix string // the start marker prefix (before attrs / closing -->)
	endMarker   string // the end marker comment
}

// markerRegistry maps each BlockType to its heading + marker contract.
// This generalizes the single markerBlockPattern from layer3.go into a
// per-type registry (spec.md §D.2).
var markerRegistry = map[BlockType]blockSpec{
	BlockTypeLearnedWorkflow: {
		heading:     "## MOAI:LEARNED-WORKFLOW",
		startPrefix: "<!-- moai:learned-start",
		endMarker:   "<!-- moai:learned-end -->",
	},
	BlockTypeHarnessGenerated: {
		heading:     "## Project-Specific Configuration (Harness-Generated)",
		startPrefix: "<!-- moai:harness-start",
		endMarker:   "<!-- moai:harness-end -->",
	},
	BlockTypeLearnedLocal: {
		heading:     "## MOAI:LEARNED-WORKFLOW-LOCAL",
		startPrefix: "<!-- moai:learned-local-start",
		endMarker:   "<!-- moai:learned-local-end -->",
	},
}

// compiledPatterns are the pre-compiled atomic-match regexes for each
// BlockType. The pattern matches heading + newline + startMarker[attrs] +
// body + endMarker as one atomic group (dotall mode), mirroring the
// original layer3.go markerBlockPattern.
var compiledPatterns = func() map[BlockType]*regexp.Regexp {
	m := make(map[BlockType]*regexp.Regexp, len(markerRegistry))
	for bt, spec := range markerRegistry {
		// (?s) dotall; [^>]* matches optional attrs in the start marker;
		// .*? non-greedy body.
		pattern := "(?s)" +
			regexp.QuoteMeta(spec.heading) + `\n` +
			regexp.QuoteMeta(spec.startPrefix) + `[^>]*-->` +
			`.*?` +
			regexp.QuoteMeta(spec.endMarker)
		re, err := regexp.Compile(pattern)
		if err != nil {
			panic(fmt.Sprintf("curator: invalid pattern for BlockType %d: %v", bt, err))
		}
		m[bt] = re
	}
	return m
}()

// BlockContent is the typed payload for a managed-block write.
type BlockContent struct {
	// Bullets is the ordered list of digest-layer entries. When RawBody is
	// empty, the writer renders the bullets into the block body.
	Bullets []Bullet
	// Tier is the evidence tier (3 = CLAUDE.local.md, 4 = CLAUDE.md). The
	// writer uses this for tier/surface validation (M2+).
	Tier int
	// RawBody is a pre-rendered body string that overrides bullet rendering
	// when non-empty. This is used by BlockTypeHarnessGenerated (via
	// InjectMarker) to pass the structured Domain/Level/Updated body through
	// the typed API without forcing it into the bullet model.
	RawBody string
	// StartAttrs carries attributes appended to the start marker (e.g.
	// ` id="..." generated="..."` for HarnessGenerated). Empty for types
	// whose start marker has no attributes (LearnedWorkflow).
	StartAttrs string
}

// Bullet is a single digest-layer entry within a managed block.
type Bullet struct {
	// LedgerKey is the cross-layer linkage to a ledger-layer entry. Empty
	// when the bullet is provisional (early-tier observation, evidence-or-null).
	LedgerKey string
	// Text is the distilled generic workflow knowledge. Anti-fabrication
	// binding (REQ-HEV2-011): MUST NOT carry internal SPEC IDs, REQ/AC
	// tokens, internal dates, or commit SHAs.
	Text string
	// Provisional is true when LedgerKey is empty (early-tier observation).
	Provisional bool
}

// Sentinel errors for the typed writer.
var (
	// ErrBlockNotFound is returned by CRUD operations when the target file
	// does not contain a marker block of the requested BlockType.
	ErrBlockNotFound = errors.New("curator: managed block not found")
	// ErrBulletNotFound is returned by UpdateBullet/DeleteBullet when no
	// bullet with the given ledger_key exists in the block.
	ErrBulletNotFound = errors.New("curator: bullet not found")
	// ErrDuplicateAppend is returned by AppendLearnedLocal when an existing
	// entry in the LOCAL section already carries the same ledger_key as the
	// proposed append (REQ-HEV2-014). The file is NOT modified.
	ErrDuplicateAppend = errors.New("curator: duplicate append (ledger_key already exists)")
)

// WriteManagedBlock reads the file at path, locates the marker block matching
// blockType, replaces the block content atomically (or appends a fresh block
// when absent), and writes the file back with pre-block and post-block bytes
// preserved verbatim.
//
// For BlockTypeLearnedWorkflow the writer enforces the three digest-layer
// gates BEFORE touching the file (REQ-HEV2-008 budget, REQ-HEV2-009 bullet
// cap, REQ-HEV2-011 anti-fabrication): on violation the file is NOT modified
// and a typed error (ErrDigestBudgetExceeded / ErrBulletCapExceeded /
// ErrForbiddenContent) is returned.
//
// The function is idempotent: re-invoking with content that renders to the
// same block text as the existing block produces zero byte-diff.
//
// Returns an error (without touching the file) when path is empty, the file
// cannot be read/written, the blockType is not in the marker registry, or a
// digest-layer gate is violated.
func WriteManagedBlock(path string, blockType BlockType, content BlockContent) error {
	if path == "" {
		return errors.New("curator: empty path")
	}
	// Digest-layer enforcement gates (LearnedWorkflow only — the legacy
	// HarnessGenerated block uses RawBody and is exempt to preserve its D1
	// byte-identical contract). These run before any file I/O so a rejected
	// write leaves the file untouched.
	if blockType == BlockTypeLearnedWorkflow {
		if err := validateDigestBlock(content); err != nil {
			return err
		}
	}
	block, err := renderBlock(blockType, content)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("curator: read %s: %w", path, err)
	}
	existing := string(data)

	re, ok := compiledPatterns[blockType]
	if !ok {
		return fmt.Errorf("curator: unknown BlockType %d", blockType)
	}

	var updated string
	if re.MatchString(existing) {
		// Atomic replace of the existing block.
		updated = re.ReplaceAllString(existing, block)
	} else {
		// Append a fresh block with a separating blank line. If the file
		// does not end with a newline, insert one first to avoid
		// concatenating the last existing line with the block heading.
		sep := ""
		if !strings.HasSuffix(existing, "\n") {
			sep = "\n"
		}
		updated = existing + sep + "\n" + block + "\n"
	}

	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("curator: write %s: %w", path, err)
	}
	return nil
}

// renderBlock assembles the full block text (heading through end marker)
// from the BlockType registry entry + the BlockContent payload.
func renderBlock(blockType BlockType, content BlockContent) (string, error) {
	spec, ok := markerRegistry[blockType]
	if !ok {
		return "", fmt.Errorf("curator: unknown BlockType %d", blockType)
	}
	body := content.RawBody
	if body == "" {
		body = renderBullets(content.Bullets)
	}
	startMarker := spec.startPrefix + content.StartAttrs + " -->"
	return spec.heading + "\n" + startMarker + "\n" + body + spec.endMarker, nil
}

// renderBullets renders a slice of Bullets into the block body text. Each
// bullet occupies one line: `- <text>` optionally followed by
// ` <!-- key: <ledger_key> -->` when LedgerKey is non-empty.
func renderBullets(bullets []Bullet) string {
	if len(bullets) == 0 {
		return ""
	}
	var b strings.Builder
	for _, bullet := range bullets {
		b.WriteString("- ")
		b.WriteString(bullet.Text)
		if bullet.LedgerKey != "" {
			b.WriteString(" <!-- key: ")
			b.WriteString(bullet.LedgerKey)
			b.WriteString(" -->")
		}
		b.WriteString("\n")
	}
	return b.String()
}

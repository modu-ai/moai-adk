// Package harness — `moai harness promote` CLI surface
// (SPEC-HARNESS-LOOP-REPAIR-001 M2-1, REQ-HLR-004).
//
// A proposalgen discovery draft is a pattern-discovery report whose designed
// consumer is manager-spec SPEC authoring, not Applier.Apply() (spec.md §A.4).
// The promote verb routes a draft to that consumer: it materialises a SPEC
// skeleton under .moai/specs/ carrying the draft ID as provenance, records a
// durable promotion audit record, and moves the draft out of the pending queue
// so `moai harness status` no longer counts it.
//
// HARD subagent boundary (C-HRA-008 / REQ-PGN-012): no source file in this
// package invokes AskUserQuestion. Promotion is explicit and per-draft; the
// verb takes flags (--id, --project-root) and emits structured output/errors.
package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/harness/proposalgen"
)

// promoteSpecsDirRel is the SPEC authoring root (relative to projectRoot) where
// the promoted skeleton lands.
const promoteSpecsDirRel = ".moai/specs"

// promotePromotedDirRel is where a promoted draft is relocated so it leaves the
// pending queue (proposalgen.ListDraftIDs reads proposals/, not promoted/).
// The draft's spec.md body + proposal.json are preserved here as the authoring
// source material — the verb never deletes a draft.
const promotePromotedDirRel = ".moai/harness/promoted"

// promoteAuditLogRel is the append-only promotion audit log (AC-HLR-005). Each
// successful promotion appends exactly one JSON record linking draft → SPEC.
const promoteAuditLogRel = ".moai/harness/promotions.jsonl"

// PromoteOptions carries the promote verb inputs.
type PromoteOptions struct {
	// ID is the draft ID to promote (required). It addresses
	// .moai/harness/proposals/<ID>/proposal.json via the shared accessor.
	ID string
	// ProjectRoot is the project root (absolute). Empty falls back to getwd.
	ProjectRoot string
}

// promotionRecord is the single audit-log entry (AC-HLR-005). Each field is
// derived from the promotion input so the record is a durable, machine-readable
// link from draft to SPEC.
type promotionRecord struct {
	DraftID    string `json:"draft_id"`
	SpecDir    string `json:"spec_dir"`
	PromotedAt string `json:"promoted_at"`
}

// RunPromote is the promote verb production entry point.
//
// Steps (REQ-HLR-004):
//  1. resolve + validate the draft via the shared proposalgen accessor;
//  2. materialise a SPEC skeleton under .moai/specs/<draftID>/ whose spec.md
//     frontmatter records <draftID> as provenance (AC-HLR-004 clause 2);
//  3. append exactly one audit record to promotions.jsonl (AC-HLR-005);
//  4. move the draft directory to .moai/harness/promoted/<draftID>/ so it
//     leaves the pending queue (AC-HLR-004 clause 3).
//
// The verb never invokes AskUserQuestion (C-HRA-008). Errors are wrapped with
// context; the cobra factory maps them to exit codes.
func RunPromote(opts PromoteOptions) error {
	root := opts.ProjectRoot
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("harness promote: resolve project root: %w", err)
		}
		root = wd
	} else {
		// Absolute-path rule (internal/cli/CLAUDE.md): filepath.Abs, never
		// filepath.Join(cwd, userPath).
		abs, err := filepath.Abs(root)
		if err != nil {
			return fmt.Errorf("harness promote: resolve project root: %w", err)
		}
		root = abs
	}

	// Step 1: resolve + validate the draft path through the shared accessor so
	// this consumer stays aligned with the producer's nested layout (REQ-HLR-001).
	draftJSONPath, err := proposalgen.ProposalPath(proposalgen.ProposalDir(root), opts.ID)
	if err != nil {
		return fmt.Errorf("harness promote: %w", err)
	}
	if _, statErr := os.Stat(draftJSONPath); statErr != nil {
		if os.IsNotExist(statErr) {
			return fmt.Errorf("harness promote: draft not found: %s", opts.ID)
		}
		return fmt.Errorf("harness promote: stat draft %s: %w", opts.ID, statErr)
	}
	draftDir := filepath.Dir(draftJSONPath)

	// Step 2: materialise the SPEC skeleton. The directory name is the draft ID
	// (deterministic, traceable, and a placeholder a human/manager-spec later
	// renames to SPEC-{DOMAIN}-{NUM}). spec-lint scans SPEC-*/ only, so a
	// PROPOSAL-* skeleton is invisible to the linter until finalised.
	specDir := filepath.Join(root, promoteSpecsDirRel, opts.ID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		return fmt.Errorf("harness promote: create SPEC dir %s: %w", specDir, err)
	}
	patternKey := readPatternKeyBestEffort(draftJSONPath)
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"),
		[]byte(specSkeleton(opts.ID, patternKey, time.Now().UTC())), 0o644); err != nil {
		return fmt.Errorf("harness promote: write spec.md: %w", err)
	}

	// Step 3: append the audit record (AC-HLR-005).
	record := promotionRecord{
		DraftID:    opts.ID,
		SpecDir:    opts.ID,
		PromotedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	if err := appendPromotionRecord(filepath.Join(root, promoteAuditLogRel), record); err != nil {
		return fmt.Errorf("harness promote: write audit record: %w", err)
	}

	// Step 4: move the draft out of the pending queue. The draft's spec.md body
	// and proposal.json are preserved under promoted/ as authoring source material.
	promotedTarget := filepath.Join(root, promotePromotedDirRel, opts.ID)
	if err := os.MkdirAll(filepath.Dir(promotedTarget), 0o755); err != nil {
		return fmt.Errorf("harness promote: create promoted dir: %w", err)
	}
	// A pre-existing target (re-promotion) is replaced; the audit record above
	// is the authoritative history, so replacing the relocated dir is safe.
	if _, statErr := os.Stat(promotedTarget); statErr == nil {
		_ = os.RemoveAll(promotedTarget)
	}
	if err := os.Rename(draftDir, promotedTarget); err != nil {
		return fmt.Errorf("harness promote: move draft to promoted: %w", err)
	}

	return nil
}

// readPatternKeyBestEffort extracts the pattern_key from the draft's
// proposal.json for the skeleton title hint. Failures are non-fatal — the
// provenance (draft ID) is the load-bearing field, not the pattern key.
func readPatternKeyBestEffort(draftJSONPath string) string {
	data, err := os.ReadFile(draftJSONPath)
	if err != nil {
		return ""
	}
	var probe struct {
		PatternKey string `json:"pattern_key"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return ""
	}
	return probe.PatternKey
}

// appendPromotionRecord appends one JSON record to the audit log, creating the
// parent directory and the file if absent. O_APPEND keeps concurrent promoters
// from clobbering each other's records.
func appendPromotionRecord(logPath string, record promotionRecord) error {
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return fmt.Errorf("mkdir audit log parent: %w", err)
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}
	defer func() { _ = f.Close() }()
	line, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal audit record: %w", err)
	}
	line = append(line, '\n')
	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("write audit record: %w", err)
	}
	return nil
}

// specSkeleton renders the placeholder spec.md body. The frontmatter records
// the draft ID as provenance (AC-HLR-004 clause 2); the body is a stub for a
// human/manager-spec to replace — the verb materialises the skeleton, not the
// requirements.
func specSkeleton(draftID, patternKey string, now time.Time) string {
	dateStr := now.Format("2006-01-02")
	title := "Promoted discovery draft " + draftID
	if patternKey != "" {
		title = "Promoted pattern " + patternKey + " (draft " + draftID + ")"
	}
	return fmt.Sprintf(`---
id: SPEC-HARNESS-TBD
title: %q
status: draft
provenance: %s
created: %s
updated: %s
---

# SPEC skeleton — promoted from draft %s

This skeleton was materialised by `+"`moai harness promote --id %s`"+`.
The `+`body`+` is a placeholder; a human or manager-spec must fill in the
requirements, acceptance criteria, and finalise the SPEC ID (the `+`id:`+`
field above is a TBD placeholder).

Source draft (spec.md body + proposal.json): `+"`.moai/harness/promoted/%s/`"+`.

Promotion audit log: `+"`.moai/harness/promotions.jsonl`"+` (AC-HLR-005).
`, title, draftID, dateStr, dateStr, draftID, draftID, draftID)
}

// NewPromoteCmd is the `moai harness promote` cobra factory.
//
// The function is exported so newHarnessRouterCmd() (internal/cli/harness_route.go,
// package cli) can register it under the active `moai harness` tree. It is also
// registered in the deprecation-marker newHarnessCmd() (internal/cli/harness.go)
// to preserve the shared-factory parity between the two trees.
//
// This subcommand never invokes AskUserQuestion. User interaction is owned
// exclusively by the orchestrator per the C-HRA-008 subagent boundary HARD contract.
func NewPromoteCmd() *cobra.Command {
	var (
		id          string
		projectRoot string
	)

	cmd := &cobra.Command{
		Use:   "promote",
		Short: "Promote a discovery draft to a SPEC skeleton (routes draft to manager-spec authoring)",
		Long: `Promote a pending harness discovery draft to a SPEC authoring skeleton.

A proposalgen draft is a pattern-discovery report, not an apply input (it
carries no target_path/field_key/new_value — spec.md §A.4). Its designed
consumer is manager-spec SPEC authoring. This verb materialises a SPEC
skeleton under .moai/specs/ carrying the draft ID as provenance, records a
promotion audit entry, and moves the draft out of the pending queue.

The skeleton body is a placeholder — a human or manager-spec fills in the
requirements. The draft's spec.md + proposal.json are preserved under
.moai/harness/promoted/ as authoring source material.

Examples:
  moai harness promote --id PROPOSAL-20260728-aaaaaaaa
  moai harness promote --id PROPOSAL-X --project-root /path/to/proj`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			err := RunPromote(PromoteOptions{ID: id, ProjectRoot: projectRoot})
			if err != nil {
				cmd.SilenceUsage = true
				cmd.SilenceErrors = true
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(),
				"promoted %s → .moai/specs/%s/ (draft moved to .moai/harness/promoted/%s/)\n",
				id, id, id)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "",
		"Draft ID to promote (addresses .moai/harness/proposals/<id>/proposal.json) (required)")
	cmd.Flags().StringVar(&projectRoot, "project-root", "",
		"Project root path (default: current directory)")

	if err := cmd.MarkFlagRequired("id"); err != nil {
		panic(fmt.Sprintf("harness promote: MarkFlagRequired: %v", err))
	}

	return cmd
}

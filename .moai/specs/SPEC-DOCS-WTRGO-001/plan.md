---
id: SPEC-DOCS-WTRGO-001
title: "Fix inaccurate moai worktree go navigation examples — implementation plan"
version: "0.1.0"
status: completed
created: 2026-07-19
updated: 2026-07-20
author: manager-spec
priority: P1
phase: "docs-site maintenance"
module: "docs-site/content/en/worktree/"
lifecycle: spec-anchored
tier: S
tags: "docs-site, worktree, cli-usage, fix"
---

# Implementation Plan — SPEC-DOCS-WTRGO-001

## §A Context

Documentation-only fix. The `moai worktree go` command prints a path to stdout and does not change directory; four docs pages incorrectly present it as a navigation command. The correct idiom is `cd "$(moai worktree go SPEC-ID)"`.

## §B Known Issues / Clarifications

- No unresolved clarification markers. The discrepancy inventory and fix patterns are fully enumerated (§F below); the correct idiom is confirmed from the CLI source and its `Long` description.

## §C Pre-flight

- Confirm the four target files exist under `docs-site/content/en/worktree/`.
- Confirm the six already-correct usages before editing so they are not touched (AC-05 / Scenario 4).
- Establish a clean docs-site build baseline (`hugo` from the docs-site root) before edits, to attribute any post-edit warning to this change.

## §D Constraints

- Markdown-only edits under `docs-site/content/en/worktree/`.
- No Go source, template, or CLI change (REQ-WTRGO-006).
- Preserve surrounding prose comments that describe the (now-corrected) behavior — e.g., "A new terminal opens and moves into the Worktree" — so the prose no longer contradicts the corrected command.

## §E Self-Verification

- Grep each edited file for residual bare `moai worktree go` (not preceded by `cd "$(` and not inside an already-correct substitution).
- Grep for the six preserved correct usages to confirm they are unchanged.
- Confirm zero matches under `internal/` / `pkg/` / `cmd/` in the change diff (docs-only guarantee).

## §F Fix Approach and Discrepancy Inventory (single milestone — sequential edits)

The fix is one logical milestone (M1) with four sequential file passes. The highest-change-likelihood decision — the exact rewrite form of each pattern — is settled first (§F.0), then applied mechanically per file.

### §F.0 Fix patterns (decision-first)

- **Pattern A (bare navigation)**: `moai worktree go X` → `cd "$(moai worktree go X)"`
- **Pattern B (chained)**: `moai worktree go X && Y` → `cd "$(moai worktree go X)" && Y`
- **Pattern C (tmux)**: `tmux new-session -d -s name 'moai worktree go X'` → `tmux new-session -d -s name -c "$(moai worktree go X)"`
- **Diagram labels**: revise node/sequence labels so they do not imply the bare command navigates.

### §F.1 Phase 1 — `_index.md`

Shell code (Pattern A): lines ~106, ~220, ~224, ~229.
Mermaid labels (REQ-WTRGO-004): line ~60 node `B1`, lines ~269-281 three `moai worktree go<br/>SPEC-AUTH-00X` node labels.
Preserve: line ~135 command-reference table (`cd "$(moai worktree go SPEC-AUTH-001)"`).

### §F.2 Phase 2 — `guide.md`

Shell code (Pattern A): lines ~387, ~396, ~501.
Shell code (Pattern B): lines ~489, ~490, ~491.
Shell code (Pattern C): lines ~578-579.
Mermaid labels (REQ-WTRGO-004): line ~417 sequence label, lines ~471-473 three `I1/I2/I3` node labels.
Preserve: lines ~114-151 command-reference table.

### §F.3 Phase 3 — `examples.md`

Shell code (Pattern A): lines ~52 (Next-steps text), ~67 (+ correct the "A new terminal opens and moves into the Worktree" prose on ~69-71), ~223, ~233, ~243, ~315, ~340, ~576 (automation script `$SPEC_ID`).
Shell code (Pattern B): lines ~511-515.
Shell code (Pattern C): lines ~539-540.
Mermaid labels (REQ-WTRGO-004): lines ~182, ~186, ~190 (flowchart node labels); lines ~478, ~483 (sequence diagram labels).
Preserve: line ~429 (`$ moai worktree go SPEC-AUTH-001` followed by `✗ The Worktree is corrupted.` — demonstrates CLI error output, NOT navigation; this usage must NOT be rewritten).
Preserve: lines ~557-560 Tip 2 for-loop (`cd "$(moai worktree go $spec)"`).

### §F.4 Phase 4 — `faq.md`

Shell code (Pattern A): lines ~150-151, ~154-155, ~158-159, ~355-356, ~361-362, ~521-522.
Preserve: lines ~122-127 flowchart usage, lines ~525-527 for-loop usage, line ~553 (standalone correct usage).

## §G Anti-Patterns to Avoid

- Blind `sed` substitution — the four correct usages and the prose comments require file-by-file judgment (per docs-site doctrine, blind bulk edits are prohibited).
- Rewriting mermaid diagrams beyond the minimal label change needed for accuracy.
- Touching non-English locales (out of scope).

## §H Cross-References

- `internal/cli/worktree/go.go` — CLI source confirming print-only behavior.
- `spec.md` §B — GEARS requirements REQ-WTRGO-001..007.
- `acceptance.md` §D — AC matrix.
- docs-site i18n rules — 4-locale sync obligation (deferred, out of scope here).

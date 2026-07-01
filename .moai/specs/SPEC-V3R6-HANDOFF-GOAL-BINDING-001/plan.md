# Implementation Plan — SPEC-V3R6-HANDOFF-GOAL-BINDING-001

## §A. Context

Documentation-rule alignment: bind Claude Code native `/goal` into the paste-ready resume Block 1 as a purpose-conditional re-set line, mirroring the established `/effort ultracode` conditional-emit mechanism. No Go source change; the only build interaction is `make build` re-embedding the edited template mirror files.

## §B. Tier Classification

**Tier M (standard).** Rationale:

- **3-artifact shape.** This SPEC produces a standalone `acceptance.md` (spec.md + plan.md + acceptance.md), which is the Tier M artifact set — Tier S inlines AC into spec.md §3. The standalone AC file alone signals Tier M per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier.
- **6-edited-file footprint (guaranteed).** 3 logical rule/style edits (session-handoff.md + moai.md + goal-directive.md), each mirrored LIVE + TMPL = 6 files, plus a conditional 7th/8th (context-window-management.md LIVE + TMPL, edited only if a contradiction is found — D.4). This 5-15-file footprint sits in the Tier M band.
- **PASS-threshold honesty.** Declaring `tier: M` encodes the **0.80** plan-auditor PASS threshold. Leaving `tier:` absent would force the backward-compat **Tier L default (0.85)**, over-stating the review bar for a doc-only change; classifying Tier S (0.75) would under-state it given the 3-artifact + 6-file shape.
- **Low risk, mechanical.** Despite the Tier M shape, the change is doc-only: zero Go code, zero behavior change to any binary; the `/effort ultracode` conditional line is the exact pattern mirrored. Harness stays light (spec-lint + doc parity greps + `make build` + `go build ./...`); no test-suite or coverage delta.

## §C. Pre-flight Verification (completed at plan-phase)

- [x] SPEC ID regex self-check → PASS (`SPEC-V3R6-HANDOFF-GOAL-BINDING-001`).
- [x] No duplicate SPEC ID under `.moai/specs/`.
- [x] All 4 candidate target files verified MIRRORED (live + template both present) — Template-First parity applies to all 4 (no live-only exemption).
- [x] `/effort ultracode` parity baseline captured: SSOT `session-handoff.md` count = 3, render `moai.md` count = 2 (the `/goal` binding should reach comparable presence).

## §D. Edit Targets (run-phase — manager-develop)

Template-First discipline: every edited `.claude/**` file also exists under `internal/template/templates/**` (all 4 verified MIRRORED). Edit the local copy AND the template copy, then `make build`.

### D.1 — SSOT: `session-handoff.md` (primary)

| Copy | Path |
|------|------|
| LIVE | `.claude/rules/moai/workflow/session-handoff.md` |
| TMPL | `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` |

Edits (REQ-HGB-001/002/003/005/006/007):
- **Canonical Format 6-block skeleton** (~line 30-45): add a `/goal <completion-condition>` conditional line to Block 1, positioned parallel to the `# /effort ultracode` comment line.
- **Block 1 Field-by-Field Specification** (~line 77-78): add a purpose-conditional `/goal` re-set bullet parallel to the `/effort ultracode` bullet — state the run-phase AND mechanically-verifiable emit trigger, the "Default on ambiguity: omit" rule, and the "NOT restored by `ultrathink.`" rationale.
- **Example** (Illustrative section, ~line 100-120): show the `/goal` line in a run-phase resume example with a verifiable condition. [HARD — D6 template neutrality] In the `session-handoff.md` **template** copy (`internal/template/templates/...`), the Example completion-condition MUST use language-neutral phrasing (e.g. "the SPEC's test suite passes AND lint is clean, or stop after N turns") — NOT `go test ./... exits 0` or any language-specific token — to preserve that template file's currently-clean neutrality baseline (verified 0 `go test` occurrences) per `CLAUDE.local.md §15/§25`. The LIVE `session-handoff.md` copy MAY likewise stay neutral for parity; `moai.md` / `goal-directive.md` already carry `go test` verbatim and are unaffected.
- **Diet Constraints — Pre-emit self-check (paste-ready budget)** (~line 284, currently "8 items"): add a `/goal` single-line diet item (single conditional line, not a multi-line block); update the item count label accordingly.
- **Anti-Patterns** general list (~line 165 region): add the `/goal` omission anti-pattern mirroring the `/effort ultracode` omission line.
- **Implementation Kickoff Approval invariant** (REQ-HGB-004): add explicit text that a `/goal` line does NOT authorize autonomous run-phase entry.

### D.2 — Render surface: `moai.md §8` (primary)

| Copy | Path |
|------|------|
| LIVE | `.claude/output-styles/moai/moai.md` |
| TMPL | `internal/template/templates/.claude/output-styles/moai/moai.md` |

Edits (REQ-HGB-001/005/006):
- **6-block skeleton** (~line 677-695): add the `/goal` conditional comment line to Block 1 parallel to the `# /effort ultracode` line.
- **Pre-emit self-check (session-handoff template completeness)** (~line 718, currently "9 items"): add a `/goal` conditional-emit item; update the "9 items" count label to reflect the addition.
- Preserve the drift-mitigation sentinel and the 3 concern-name qualifiers verbatim.

### D.3 — Cross-reference: `goal-directive.md`

| Copy | Path |
|------|------|
| LIVE | `.claude/rules/moai/workflow/goal-directive.md` |
| TMPL | `internal/template/templates/.claude/rules/moai/workflow/goal-directive.md` |

Edit (REQ-HGB-008): strengthen the MoAI Integration Notes `ultrathink.` resume-pairing bullet to explicitly state the Block 1 conditional `/goal` re-set binding and cross-reference `session-handoff.md`.

### D.4 — Non-contradiction target: `context-window-management.md` (conditional edit)

| Copy | Path |
|------|------|
| LIVE | `.claude/rules/moai/workflow/context-window-management.md` |
| TMPL | `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` |

Action (REQ-HGB-010): confirm the Resume message format section references `session-handoff.md` as SSOT and contains no `/goal`-contradicting inline claim. Edit ONLY if a contradiction is found; otherwise this is a verify-only (no-edit) target.

### D.5 — Re-embed

- Run `make build` after all template edits to recompile the embedded template FS (`//go:embed all:templates` in `internal/template/embed.go` — no generated `embedded.go`).
- Verify `go build ./...` passes.

## §E. Self-Verification (plan-phase audit-ready signal)

See `progress.md` §E.1. Plan-phase artifact set complete: spec.md (10 REQ, GEARS), plan.md (this file), acceptance.md (Given-When-Then + mechanical AC), progress.md (§E skeleton).

## §F. Milestones (priority-ordered, no time estimates)

- **M1** — SSOT + render live edits (D.1 + D.2 LIVE copies): add `/goal` line, Field-by-Field bullet, Example, Diet self-check item, Anti-pattern, Kickoff-Approval invariant, render skeleton + self-check item.
- **M2** — Cross-reference live edits (D.3 LIVE); D.4 non-contradiction check (edit only if needed).
- **M3** — Template-First mirror: apply identical edits to all TMPL copies (D.1-D.4 template paths); run `make build`; verify `go build ./...`.
- **M4** — Verification: run acceptance.md mechanical AC greps (SSOT↔render parity, Template-First parity, localization column parity, concern-name-qualifier count, anti-pattern presence); confirm all PASS.

## §G. Anti-Patterns to avoid during run-phase

- Editing the LIVE copy but forgetting the TMPL mirror (or vice versa) → Template-First parity AC fails. All 4 files are MIRRORED; edit both copies.
- Introducing a new pre-emit self-check concern-name qualifier → violates REQ-HGB-005 parity. Fit the `/goal` item within the existing 3 qualifiers.
- Making the `/goal` line multi-line → violates REQ-HGB-003 Diet. Single conditional line only.
- Translating the `/goal` literal or the `✂`/`─` symbols → violates REQ-HGB-006 localization contract. Command literals + marker symbols are verbatim across locales.
- Wording the `/goal` binding so it reads as authorizing autonomous run-phase → violates REQ-HGB-004. State the Kickoff-Approval invariant explicitly.
- Forgetting to update the self-check item-count labels ("8 items" / "9 items") after adding the `/goal` items → count-label drift.

## §H. Cross-References

Same as spec.md §D. Primary SSOT: `session-handoff.md`. Render surface: `moai.md §8`. The `/effort ultracode` conditional line is the canonical pattern to mirror.

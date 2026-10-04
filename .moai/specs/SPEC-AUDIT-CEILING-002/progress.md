# SPEC-AUDIT-CEILING-002 — Progress

## §A Status

Plan phase complete (2026-10-04, card t1500, worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1500`, branch `WT-audit-ceiling-guard`, HEAD `e497f6936`). Tier M: spec.md + plan.md + acceptance.md + progress.md (no design.md, no research.md). 6 REQ / 8 AC. This SPEC is the operator-decision-D9 narrowing of the never-landed SPEC-AUDIT-CEILING-001 (old-card draft, 3 audit iterations, FAIL 0.81 STOP); its final-round lessons are folded in per spec.md §G History.

## §B Plan-Phase Decisions

Recorded in spec.md §D (D1-D7): counter input family (card-scoped plan-audit family + SPEC-scoped dir; the forbidden `.moai/reports/plan-audit/` legacy stream is not an input), outcome enum exactly {debt-proceed, split, hold} with `hold-and-split` mapping to hold + split reference, the explicit clean-PASS-at-ceiling arm (old D31), absence-is-not-a-pass (old D32), the fail-closed resolver (old D21), the collision-free `REQ-ACR` prefix, and the CLI verb as the recording path's only caller (gate/agent-body wiring deferred).

## §C Scope Boundary

Out of scope per operator D9 and spec.md §E: cross-card dedupe, receipt formats beyond the one line, override/audit-trail surfaces, gate-text reconciliation, agent-body/gate wiring, in-code `auto_delta_rounds` eligibility. The old card tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1500` is read-only reference; nothing was written there.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-04
audit_ready: true

Evidence ledger (all measurements this plan phase, this tree, HEAD `e497f6936`, from the worktree root):

1. Code anchors re-verified live (dated pointers, plan.md §B): `internal/auditverdict/verdict.go` — Parse :69, AdmitLabel :157, SpecTier :167, PlanThreshold :184, Admit :195-233; `internal/contract/kickoff/decide.go` — auditverdict.Parse :373, Admit :376; `internal/homestate/card_audit_kickoff.go` — auditKickoffRefusal :16-36, §E.1 audit_ready reader :40-70; `internal/homestate/card_evidence_readers.go` — readAuditVerdict :93, admitVerdictFile :176-192; `internal/cli/mcp_worktree_root.go` — resolveAuditGates :122-132 with the fail-open comment at :121; `internal/config/types.go` — HarnessConfig :1328, PlanAudit :1401; `.moai/config/sections/harness.yaml` — plan_audit_tier_ceilings :75, plan_audit_ceiling_policy :82.
2. RED-now cells: nine ledger entries (LEDGER-ACR-A/B/B2/C/D/E/F/G/H) with commands, verbatim outputs, exit codes, and tree pin in acceptance.md §B — 3 absence greps (exit 1), 2 convention-doc zero counts (exit 1), outcome-type + record-directory absence (exit 1), the fail-open presence (exit 0, red because the defect exists), and 2 package-wide `go test -list` corroborations (exit 0, no match listed).
3. Report-tree population (primary checkout, worktrees excluded, 2026-10-04): 327 `plan-audit-iter*.md` files; max 6 in one card directory (`.moai/reports/t1152/`); 169 `SPEC-*-review-*.md` files, all under the convention-forbidden `.moai/reports/plan-audit/`. The operator's 405/160/max-11 figure is cited as the operator's measurement, not re-derived here.
4. SPEC id uniqueness: `SPEC-AUDIT-CEILING-002` matched nothing under `.moai/specs/` before adoption (grep exit 1). `related_specs` names only SPECs present on this tree (`SPEC-WF-AUDIT-GATE-001`, `SPEC-AUDIT-MODEL-CONVERGE-001` — both verified present).
5. Lint: `go run ./cmd/moai spec lint SPEC-AUDIT-CEILING-002 --strict` from this tree — see §G item 1 for the observed result. From-tree invocation is deliberate: the installed `moai` binary (build 0732cc699) predates this tree, and a tree-built run avoids attributing a measurement to a lagging build (verification-claim-integrity §2.2).

## §G Plan-Phase Notes

1. Lint corrections applied during authoring: (a) the OutOfScopeRule shape — the conforming form is an H3 `### Out of Scope — …` heading with list items (the `## Out of Scope`-heading-only shape triggers MissingExclusions); (b) every `-run` selector anchored (`^…$` per test name), and the symmetry-suite selector replaced by the measured harness surface — `TestStructYAMLSymmetry_*` has no Harness case (`go test -list` enumeration), so coverage lands in `loader_harness_extended_test.go` with a named new test whose absence is pinned as LEDGER-ACR-I. Final lint result is recorded in the plan-phase commit message body.
2. Binary lag: the running moai MCP server is build 0732cc699 (ancestor of `e497f6936`); no plan-phase finding rests on the lagging binary's behavior — every cited measurement is a from-tree grep, `go test -list`, or from-tree `go run`.
3. The `✂` handoff, mode selection, and run-phase logs (§E.2-§E.4, §F) are run/sync-phase concerns and intentionally absent at plan close.

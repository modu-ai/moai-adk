# SPEC-AUDIT-CEILING-002 — Progress

## §A Status

Plan phase at v0.2.0 (2026-10-04, card t1500, worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1500`, branch `WT-audit-ceiling-guard`; v0.1.0 at `e497f6936`, iter1 audit FAIL 0.75 at `58282d5ac`, repair commit follows). Tier M: spec.md + plan.md + acceptance.md + progress.md (no design.md, no research.md). 6 REQ / 11 AC (v0.2.0 added AC-ACR-009..011 and extended AC-ACR-003). This SPEC is the operator-decision-D9 narrowing of the never-landed SPEC-AUDIT-CEILING-001 (old-card draft, 3 audit iterations, FAIL 0.81 STOP); its final-round lessons are folded in per spec.md §G History.

## §B Plan-Phase Decisions

Recorded in spec.md §D (D1-D7): counter input family (card-scoped plan-audit family + SPEC-scoped dir; the forbidden `.moai/reports/plan-audit/` legacy stream is not an input), outcome enum exactly {debt-proceed, split, hold} with `hold-and-split` mapping to hold + split reference, the explicit clean-PASS-at-ceiling arm (old D31), absence-is-not-a-pass (old D32), the fail-closed resolver (old D21), the collision-free `REQ-ACR` prefix, and the CLI verb as the recording path's only caller (gate/agent-body wiring deferred).

## §C Scope Boundary

Out of scope per operator D9 and spec.md §E: cross-card dedupe, receipt formats beyond the one line, override/audit-trail surfaces, gate-text reconciliation, agent-body/gate wiring, in-code `auto_delta_rounds` eligibility. The old card tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1500` is read-only reference; nothing was written there.

## §D Iter1 Repair Polarity Table (v0.1.0 → v0.2.0)

Every normative sentence rewritten in the iter1 repair, old and new quoted side by side, with the polarity relation in one line. No rewrite reverses a direction; each preserves or strengthens the fail-closed posture.

| # | Surface | Old sentence (v0.1.0) | New sentence (v0.2.0) | Polarity relation |
|---|---------|----------------------|----------------------|-------------------|
| P1 | spec.md REQ-ACR-003 trigger | "When the system computes a SPEC's round count at or above its resolved ceiling:" | "When the system computes a SPEC's round count at or above its resolved ceiling and its latest verdict is not an admitted `PASS` — the REQ-ACR-004 arm:" | Trigger NARROWS to exclude the case REQ-ACR-004 owns; the two REQs now compose instead of contradicting (D4). No direction flip. |
| P2 | spec.md REQ-ACR-003 policy mapping | "a non-admitted verdict under any other or unreadable policy value records `hold`" | "a non-admitted verdict under a policy value of `split` records `split`, a non-admitted verdict under any other or unreadable policy value records `hold`" | ADDS one selectable arm; `hold` remains the fail-closed fallback for everything unnameable (D5). `split` becomes reachable, `hold` unchanged. |
| P3 | spec.md REQ-ACR-006 | "When the audit-gates resolution of a tree errors): The resolver (`resolveAuditGates` …) shall keep the error distinct from an empty not-configured result…" | "When the audit-gates resolution path of a tree errors — the workflow audit pins loader or the audit-gates resolver): The resolution path shall keep the error distinct from an empty not-configured result end to end — `workflowAuditPins` … shall not fold a read or parse error into a zero configuration, and `resolveAuditGates` … shall not fold a resolution error into an empty gate set —" | BROADENS the same fail-closed guarantee from one surface to both error classes on two surfaces (D3). Nothing weakened. |
| P4 | spec.md REQ-ACR-002 tail | "…comparing the computed round count against that ceiling." | "…and shall treat a resolved ceiling that is missing from the map or non-positive as a configuration error rather than as a ceiling of zero, comparing the computed round count against that ceiling only when it resolved." | STRENGTHENS — a mis-configured ceilings map can no longer fire the ceiling path with a zero ceiling (D10). Comparison itself unchanged. |
| P5 | spec.md C5 | "C5 — cross-platform. The new code paths use no `syscall` and no OS-conditional builds." | "C5 — cross-platform exemption: no syscall is used. The new code paths use no `syscall`, so no `//go:build` constraint applies and no build-tag literal exists in this SPEC's scope…" | Same prohibition restated in the explicit exemption-clause form the D8 gate matches (D9). Prohibition content identical. |
| P6 | acceptance.md §E edge 1 | "0 is below every ceiling (all configured ceilings ≥ 1), so the ceiling path never fires on an unaudited SPEC." | "0 is below every resolved ceiling, and a ceiling that resolves missing or non-positive is a configuration error (REQ-ACR-002), so the ceiling path never fires on an unaudited SPEC or on a mis-configured ceilings map." | The unenforced promise now cites the requirement that enforces it (D10). Promise kept, scope widened to the mis-configured case. |

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-04
audit_ready: true

Evidence ledger (all measurements this plan phase, this tree, HEAD `e497f6936`, from the worktree root):

1. Code anchors re-verified live (dated pointers, plan.md §B): `internal/auditverdict/verdict.go` — Parse :69, AdmitLabel :157, SpecTier :167, PlanThreshold :184, Admit :195-233; `internal/contract/kickoff/decide.go` — auditverdict.Parse :373, Admit :376; `internal/homestate/card_audit_kickoff.go` — auditKickoffRefusal :16-36, §E.1 audit_ready reader :40-70; `internal/homestate/card_evidence_readers.go` — readAuditVerdict :93, admitVerdictFile :176-192; `internal/cli/mcp_worktree_root.go` — resolveAuditGates :122-132 with the fail-open comment at :121, `auditSectionForRoot` feeding `workflowAuditPins(primary)` at :109; `internal/cli/audit_pin.go` — the "(N3)" fail-open comment :54-57 and the zero-config fold :58-63 (both error classes produced at :38-40 read / :47-49 parse); `internal/config/types.go` — HarnessConfig :1328-1353 with PlanAuditGlobal at :1349-1350 (the v0.1.0 pointer naming ":1401 PlanAudit" was `LevelConfig`'s member — corrected, iter1 D16); `.moai/config/sections/harness.yaml` — plan_audit_tier_ceilings :75, plan_audit_ceiling_policy :82; shipped-key inventory leaves at `internal/config/testdata/shipped_key_inventory.yaml:807-819`.
2. RED-now cells: TWELVE ledger entries (LEDGER-ACR-A, B, B2, C, D, E, F, G, H, I, J, K) with commands, verbatim outputs, exit codes, and tree pins in acceptance.md §B — the v0.1.0 eight-cell batch (A-I) measured at `e497f6936`, the v0.2.0 repair cells J and K measured at `58282d5ac`, and I re-run narrowed at `58282d5ac` (D11). Composition: 4 absence greps (A, B, F at exit 1; D at exit 0 — red because the defect exists), 2 convention-doc zero counts (B2, exit 1), outcome-type + record-directory absence (C, exit 1), 2 output/verb observations (H, J), and 4 package-wide `go test -list` corroborations (E, G, I, K — exit 0, no test listed). The v0.1.0 text of this item miscounted nine entries and two corroborations; corrected per iter1 defect D8.
3. Report-tree population (primary checkout, worktrees excluded, 2026-10-04): 327 `plan-audit-iter*.md` files; max 6 in one card directory (`.moai/reports/t1152/`); 169 `SPEC-*-review-*.md` files, all under the convention-forbidden `.moai/reports/plan-audit/`. The operator's 405/160/max-11 figure is cited as the operator's measurement, not re-derived here.
4. SPEC id uniqueness: `SPEC-AUDIT-CEILING-002` matched nothing under `.moai/specs/` before adoption (grep exit 1). `related_specs` names only SPECs present on this tree (`SPEC-WF-AUDIT-GATE-001`, `SPEC-AUDIT-MODEL-CONVERGE-001` — both verified present).
5. Lint: `go run ./cmd/moai spec lint SPEC-AUDIT-CEILING-002 --strict` from this tree — see §G item 1 for the observed result. From-tree invocation is deliberate: the installed `moai` binary (build 0732cc699) predates this tree, and a tree-built run avoids attributing a measurement to a lagging build (verification-claim-integrity §2.2).

## §G Plan-Phase Notes

1. Lint record, inline (iter1 D15): v0.1.0 — corrections applied during authoring were (a) the OutOfScopeRule shape (the conforming form is an H3 `### Out of Scope — …` heading with list items; the `## Out of Scope`-heading-only shape triggers MissingExclusions) and (b) anchored `-run` selectors; final result "✓ No findings — all SPEC documents are valid", exit 0, from this tree. v0.2.0 repair — re-run after the D1-D10 edits: "✓ No findings — all SPEC documents are valid", exit 0 (measured this repair round, from this tree, at the repair commit).
2. Binary lag: the running moai MCP server is build 0732cc699 (ancestor of `e497f6936`); no plan-phase finding rests on the lagging binary's behavior — every cited measurement is a from-tree grep, `go test -list`, or from-tree `go run`.
3. The `✂` handoff, mode selection, and run-phase logs (§E.2-§E.4, §F) are run/sync-phase concerns and intentionally absent at plan close.

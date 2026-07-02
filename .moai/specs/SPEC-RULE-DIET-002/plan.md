---
id: SPEC-RULE-DIET-002
title: "Rule diet: scope 6 reference-doctrine rules out of the always-loaded context surface — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-07-02
updated: 2026-07-02
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai + internal/template/templates/.claude/rules/moai"
lifecycle: spec-anchored
tags: "rule-diet, always-loaded, context-budget, paths-scoping, template-first, steering-align"
tier: M
era: V3R6
---

## A. Context

Implementation plan for SPEC-RULE-DIET-002. The change is frontmatter-only (`paths:` additions) plus two additive one-line retrieval summaries in already-always-loaded KEEP surfaces. It mirrors the mechanics validated by SPEC-STEERING-ALIGN-RULE-SCOPING-001: template-first → `make build` → byte-identical parity, verified by re-runnable commands and the existing token-budget guard.

## B. Known Issues / Risks (with mitigations)

| Risk | Severity | Mitigation |
|------|----------|------------|
| Over-scoping a load-bearing rule (the anti-goal) breaks per-turn orchestrator behavior | High | KEEP set fixed by REQ-RD2-006/007 with §A.3.1 evidence; the 6-file SCOPE set each has a pointer in a KEEP surface (§A.3.2). Plan-auditor scrutiny expected at this boundary. |
| `runtime-recovery-doctrine.md` scoped but an agent hits a withheld-recoverable error with the doctrine absent | Medium | REQ-RD2-004: verify the recovery pointer in `agent-common-protocol.md § Recovery-Signal Carve-Out` + actionable rungs in `session-handoff.md` / `context-window-management.md` (all KEEP) BEFORE scoping. M1 gate. |
| `sprint-round-naming.md` scoped → orchestrator drifts to legacy terms (Sprint/Round/Wave) in banners | Medium | REQ-RD2-005: verify/add a compact 4-term + retired-alias note in `moai.md § Banner Localization` (KEEP) BEFORE scoping. M1 gate. |
| Template/live drift after `paths:` addition | Medium | REQ-RD2-009: template-first + `make build` + byte-identical parity check per file (M3). |
| Token-budget guard count assertion breaks | Low | The guard recomputes `countNoPathsRuleFiles` relatively (no hardcoded literal); it auto-adjusts. REQ-RD2-011 forbids editing the guard. |
| A cross-reference to a scoped rule silently stops resolving | Low | Self-referential `paths:` (REQ-RD2-002) keeps files at their existing `.claude/rules/moai/...` path; all pointers still resolve. |

## C. Pre-flight (run-phase entry checks)

Before any edit, re-run the ground-truth commands (acceptance.md AC-RD2-001) to confirm LIVE=11 and TEMPLATE=11 always-loaded and that the 6 SCOPE files are MIRRORED. If the counts have drifted since plan-phase, return a blocker report rather than proceeding on a stale baseline.

## D. Technical Approach

- **Mechanism**: uniform self-referential `paths: "**/<filename>.md"` on each of the 6 SCOPE files (REQ-RD2-002). Frontmatter-only (REQ-RD2-008). Mechanism (b) `.moai/docs/` relocation is rejected (spec.md §C).
- **Exact frontmatter to add per SCOPE file** (top of file, before the first `#` heading; if a `description:` already exists, add `paths:` as a sibling key):

  | File (both trees) | `paths:` value |
  |-------------------|----------------|
  | `workflow/runtime-recovery-doctrine.md` | `"**/runtime-recovery-doctrine.md"` |
  | `workflow/dynamic-workflows.md` | `"**/dynamic-workflows.md"` |
  | `workflow/native-invocation-model.md` | `"**/native-invocation-model.md"` |
  | `development/sprint-round-naming.md` | `"**/sprint-round-naming.md"` |
  | `workflow/goal-directive.md` | `"**/goal-directive.md"` |
  | `workflow/verification-batch-pattern.md` | `"**/verification-batch-pattern.md"` |

- **KEEP set untouched**: the 5 KEEP files receive no frontmatter change.
- **Two additive retrieval summaries** (only if verification finds them absent):
  - runtime-recovery pointer — confirm `agent-common-protocol.md § Recovery-Signal Carve-Out` already names `runtime-recovery-doctrine.md §4` as SSOT (it does, per plan-phase check); no add expected.
  - sprint-round-naming 4-term note — confirm `moai.md § Banner Localization` carries the canonical Epic/SPEC/Milestone/Constitution set + retired-alias note; add a one-line note if absent.

## E. Self-Verification

See progress.md §E.1 for the plan-phase audit-ready signal.

## F. Milestones (priority-ordered, no time estimates)

- **M1 — Mitigation-precondition verification (gate for the 2 mitigated files)**
  - Verify the recovery pointer + actionable rungs exist in KEEP surfaces (REQ-RD2-004).
  - Verify/add the 4-term + retired-alias note in `moai.md § Banner Localization` (REQ-RD2-005).
  - Exit condition: both retrieval paths confirmed in an always-loaded surface. If not confirmable, drop `runtime-recovery-doctrine.md` and/or `sprint-round-naming.md` from the SCOPE set (deliver the 4 HIGH-confidence files only) and record the descope in progress.md.

- **M2 — Template-tree `paths:` additions (SSOT first)**
  - Add the self-referential `paths:` frontmatter to all 6 SCOPE files under `internal/template/templates/.claude/rules/moai/…` (REQ-RD2-001/002/008/009).

- **M3 — Re-embed + live-tree parity**
  - Run `make build` to re-embed the template FS into the binary.
  - Apply the identical `paths:` frontmatter to the 6 live-tree files.
  - Verify byte-identical template/live parity per file (REQ-RD2-009).

- **M4 — Measurable-delta + guard verification**
  - Re-run the count commands: LIVE 11→5, TEMPLATE 11→5 (REQ-RD2-010).
  - Run `go test ./internal/config/...` — the token-budget guard + enumeration test pass with the reduced surface (REQ-RD2-011).
  - Run `go build ./...` and `go vet ./...` (no code changed, but embed regeneration must compile).

- **M5 — Anti-goal audit (KEEP integrity)**
  - Confirm the 5 KEEP files still carry NO `paths:` (REQ-RD2-006/007).
  - Confirm no SCOPE-file body content changed (frontmatter-only diff, REQ-RD2-008).
  - Confirm all inbound cross-references to the 6 scoped files still resolve to `.claude/rules/moai/...` paths.

## G. Anti-Patterns (do NOT)

- Do NOT add `paths:` to a KEEP file (over-scoping — the anti-goal).
- Do NOT edit `token_budget_guard.go` / its test to make counts pass.
- Do NOT relocate any scoped rule body to `.moai/docs/`.
- Do NOT modify any SCOPE-rule body (frontmatter-only).
- Do NOT skip `make build` after a template edit (embed drift).
- Do NOT proceed if M1 mitigation preconditions fail — descope instead.

## H. Cross-References

- SPEC-STEERING-ALIGN-RULE-SCOPING-001 (predecessor; MIRRORED template-first precedent).
- `internal/config/token_budget_guard.go` (verification oracle).
- CLAUDE.local.md §2 (Template-First), §15 (language neutrality).

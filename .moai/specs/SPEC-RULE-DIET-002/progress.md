---
id: SPEC-RULE-DIET-002
title: "Rule diet: scope 6 reference-doctrine rules out of the always-loaded context surface — Progress"
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

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored: spec.md + plan.md + acceptance.md + progress.md (this file).

- **SPEC ID self-check**: `SPEC-RULE-DIET-002` — decomposition: SPEC ✓ | RULE ✓ | DIET ✓ | 002 ✓ → PASS (regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- **Frontmatter**: 12 canonical fields + `tier: M` + `era: V3R6` present in all 4 files.
- **Tier**: M (2-tree frontmatter change across 12 files + `make build` re-embed + guard verification; more than a single-file Tier-S edit, less than an L architecture change).
- **KEEP/SCOPE classification (evidence-based, spec.md §A.3)**:
  - KEEP (5, remain always-loaded): session-handoff.md, agent-common-protocol.md, askuser-protocol.md, verification-claim-integrity.md, context-window-management.md — 100.5 KB.
  - SCOPE (6, add self-referential `paths:`): runtime-recovery-doctrine.md, dynamic-workflows.md, native-invocation-model.md, sprint-round-naming.md, goal-directive.md, verification-batch-pattern.md — 71.7 KB.
- **Estimated always-loaded reduction**: LIVE 11→5 files, −73,466 bytes ≈ −71.7 KB ≈ −18,367 tokens (char/4). TEMPLATE 11→5 likewise.
- **Anti-goal guard**: REQ-RD2-006/007 forbid `paths:` on the 5 KEEP files; context-window-management.md KEEP is grounded in session-handoff.md's inline-number delegation.
- **Mitigation gates (observed at plan-phase)**:
  - REQ-RD2-004 (recovery pointer): **fully present** — `agent-common-protocol.md § Recovery-Signal Carve-Out` (always-loaded KEEP) names `runtime-recovery-doctrine.md §4` as SSOT (observed lines 60-72); actionable rungs confirmed in session-handoff.md + context-window-management.md (KEEP). No add needed.
  - REQ-RD2-005 (4-term naming summary): **partially present** — `moai.md` line 594 carries "Epic = multi-SPEC grouping, distinct from Milestone = within-SPEC ordered step" (always-loaded KEEP), but the **retired-alias note (Sprint/Round/Wave/cohort retired) is NOT present**. Run-phase M1 MUST add a one-line retired-alias note to `moai.md § Banner Localization` before scoping sprint-round-naming.md, else descope that file.
- **AC coverage**: 12 ACs (AC-RD2-001..012), all command-verifiable; 6 MUST-FIX, 2 SHOULD-FIX (mitigation), 4 NICE/support.
- **Exclusions**: 5 `### Out of Scope — <topic>` H3 sub-headings present (spec.md §C).

_Ready for plan-auditor review — the KEEP/SCOPE boundary at §A.3 is the focus of scrutiny._

## §E.2 Run-phase Evidence

- **M1 (mitigation gate)**: REQ-RD2-004 recovery pointer confirmed present in `agent-common-protocol.md` (KEEP, count 2). REQ-RD2-005 retired-alias note was ABSENT in `moai.md` → added a one-line canonical-taxonomy anchor bullet to `moai.md § Banner Localization` (both trees, byte-identical): "the four canonical terms are Epic / SPEC / Milestone / Constitution; the legacy aliases Sprint / cohort / Round / Wave are RETIRED". This keeps sprint-round-naming.md safely scopable (no descope needed).
- **M2/M3 (paths: additions)**: self-referential `paths: "**/<filename>.md"` frontmatter prepended to all 6 SCOPE files in BOTH trees (12 files): runtime-recovery-doctrine, dynamic-workflows, native-invocation-model, sprint-round-naming, goal-directive, verification-batch-pattern. Frontmatter-only (no body change).
- **Baseline (pre-flight AC-RD2-001, orchestrator-measured this run)**: LIVE 11 / TEMPLATE 11 always-loaded (matched plan baseline; the P2 runtime-recovery-doctrine template deployment had brought TEMPLATE to 11, so both trees reconciled at 11).
- Parity note (AC-RD2-008 adaptation): 3 SCOPE files are byte-parity mirrored (dynamic-workflows / native-invocation-model / goal-directive — verified byte-IDENTICAL after identical paths: add); 3 are sanitized/diverged pairs (runtime-recovery-doctrine + verification-batch-pattern in sanitized_pair registry; sprint-round-naming carries pre-existing template-lag divergence) — for those the sanitized_pair_parity test governs, not byte-identity. Identical paths: added to both trees regardless.

## §E.3 Run-phase Audit-Ready Signal

- **AC-RD2-009 (measurable delta) PASS**: LIVE always-loaded 11→**5**, TEMPLATE 11→**5** (orchestrator command-verified).
- **AC-RD2-002 PASS**: each of the 6 SCOPE files prints `paths: "**/<f>.md"` (both trees).
- **AC-RD2-006 PASS**: all 5 KEEP files print OK (no `paths:`).
- **AC-RD2-008 PASS (adapted)**: 3 byte-parity SCOPE files byte-IDENTICAL; 3 sanitized-pair files governed by sanitized_pair_parity test (GREEN).
- **AC-RD2-010 PASS**: `go test ./internal/config/...` GREEN (TestAlwaysLoadedTokenBudget + enumeration auto-adjust to reduced surface); `token_budget_guard.go` unchanged.
- **AC-RD2-004/005 PASS (SHOULD)**: recovery pointer present; retired-alias note added to moai.md.
- **AC-RD2-011/012 PASS**: inbound cross-refs still resolve (self-referential paths keep files in place); `go build ./...` + `go vet` exit 0; `go test ./internal/template/...` GREEN (mirror/leak/neutrality/embed).

## §E.4 Sync-phase Audit-Ready Signal

- 3-phase close (Tier M, orchestrator-direct sync): status `draft → completed`, era `V3R6`, updated `2026-07-02`.
- CHANGELOG `### Changed` 엔트리 추가.
- sync_commit_sha: <backfill>
- MX Tag: Tier M — MX는 sync sub-step(3-phase close, 별도 Mx 커밋 없음).

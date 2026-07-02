---
id: SPEC-RULE-DIET-002
title: "Rule diet: scope 6 reference-doctrine rules out of the always-loaded context surface — Acceptance Criteria"
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

## D. Acceptance Criteria Matrix

| AC | REQ | Verification command (re-runnable) | Pass condition |
|----|-----|-------------------------------------|----------------|
| AC-RD2-001 | context/pre-flight | `for f in $(find .claude/rules/moai -name '*.md'); do grep -q '^paths:' "$f" \|\| echo x; done \| wc -l` | Baseline reads **11** (pre-change) / **5** (post-change) |
| AC-RD2-002 | REQ-RD2-001/002 | `for f in runtime-recovery-doctrine dynamic-workflows native-invocation-model sprint-round-naming goal-directive verification-batch-pattern; do p=$(find .claude/rules/moai -name "$f.md"); grep -m1 '^paths:' "$p"; done` | Each prints a self-referential `paths: "**/<f>.md"` line |
| AC-RD2-003 | REQ-RD2-008 | `git diff --stat` on each scoped file (run-phase) | Only frontmatter lines added; 0 body lines changed |
| AC-RD2-004 | REQ-RD2-004 | `grep -n 'runtime-recovery-doctrine.md' .claude/rules/moai/core/agent-common-protocol.md` + `grep -n 'context-window-management' .claude/rules/moai/workflow/session-handoff.md` | Recovery pointer present in agent-common-protocol.md (KEEP); rungs/ /clear delegation present in session-handoff.md (KEEP) |
| AC-RD2-005 | REQ-RD2-005 | `grep -niE 'Epic.*SPEC.*Milestone\|Sprint.*retired\|Round.*retired' .claude/output-styles/moai/moai.md` | Canonical 4-term set + retired-alias note present in an always-loaded surface |
| AC-RD2-006 | REQ-RD2-006 | `for f in session-handoff agent-common-protocol askuser-protocol verification-claim-integrity context-window-management; do p=$(find .claude/rules/moai -name "$f.md"); grep -q '^paths:' "$p" && echo "FAIL $f" \|\| echo "OK $f"; done` | All 5 print `OK` (no `paths:` on any KEEP file) |
| AC-RD2-007 | REQ-RD2-007 | `grep -n 'carries no inline model-class' .claude/rules/moai/workflow/session-handoff.md` + confirm context-window-management.md has no `paths:` | The SSOT-delegation line exists AND context-window-management.md remains always-loaded |
| AC-RD2-008 | REQ-RD2-009 | `diff <(cat internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md) <(cat .claude/rules/moai/workflow/dynamic-workflows.md)` (repeat per scoped file) | Byte-identical (exit 0) for all 6 scoped files |
| AC-RD2-009 | REQ-RD2-010 | LIVE: command in AC-RD2-001 → 5. TEMPLATE: same over `internal/template/templates/.claude/rules/moai` → 5 | LIVE 11→5 AND TEMPLATE 11→5 |
| AC-RD2-010 | REQ-RD2-011 | `go test ./internal/config/...` | `TestAlwaysLoadedTokenBudget`, `TestAlwaysLoadedSurfaceEnumeration` PASS; guard source unchanged (`git diff --quiet internal/config/token_budget_guard.go`) |
| AC-RD2-011 | REQ-RD2-008 (cross-ref integrity) | `grep -rn 'runtime-recovery-doctrine\|dynamic-workflows\|native-invocation-model\|sprint-round-naming\|goal-directive\|verification-batch-pattern' .claude/rules/moai .claude/output-styles CLAUDE.md` | All inbound refs still resolve to `.claude/rules/moai/...` paths (self-referential paths preserved the location) |
| AC-RD2-012 | REQ-RD2-009 | `go build ./... && go vet ./...` | exit 0 (embed regeneration compiles) |

### D.1 Severity

- **MUST-FIX (blocking)**: AC-RD2-002, AC-RD2-006, AC-RD2-007, AC-RD2-008, AC-RD2-009, AC-RD2-010. (Scoping applied correctly; KEEP integrity preserved; parity + guard green.)
- **SHOULD-FIX**: AC-RD2-004, AC-RD2-005 (mitigation preconditions — if unverifiable, descope the affected file per plan M1, not a hard fail of the SPEC).
- **NICE**: AC-RD2-011, AC-RD2-012, AC-RD2-003.

### D.2 Traceability

Every REQ-RD2-001..011 maps to ≥1 AC above. The anti-goal (REQ-RD2-006/007) is guarded by AC-RD2-006/007. The measurable delta (REQ-RD2-010) is AC-RD2-009. The guard non-regression (REQ-RD2-011) is AC-RD2-010.

## Given-When-Then Scenarios

### Scenario 1 — SCOPE file removed from always-loaded surface (happy path)

- **Given** `workflow/dynamic-workflows.md` has no `paths:` frontmatter (always-loaded) in both trees.
- **When** the run-phase adds `paths: "**/dynamic-workflows.md"` template-first, runs `make build`, and mirrors to the live tree.
- **Then** `hasPathsRestriction` returns true for the file, it is excluded from `alwaysLoadedSurface`, the LIVE/TEMPLATE always-loaded count drops by 1 for this file, template/live are byte-identical, and every "See dynamic-workflows.md" pointer in CLAUDE.md §15 still resolves.

### Scenario 2 — KEEP file protected against over-scoping (anti-goal guard)

- **Given** `workflow/context-window-management.md` is a KEEP file that holds the numeric `/clear` thresholds session-handoff.md (KEEP) delegates to it.
- **When** the anti-goal audit (M5) runs `grep '^paths:'` against all 5 KEEP files.
- **Then** none carries a `paths:` field, context-window-management.md remains always-loaded, and the always-loaded surface still contains the model-class threshold numbers.

### Scenario 3 — Mitigation precondition fails → descope, not silent drop (edge case)

- **Given** the run-phase cannot confirm the recovery pointer in `agent-common-protocol.md § Recovery-Signal Carve-Out`.
- **When** M1 evaluates the mitigation precondition for `runtime-recovery-doctrine.md`.
- **Then** `runtime-recovery-doctrine.md` is DROPPED from the SCOPE set (delivered as the 4 HIGH-confidence + `sprint-round-naming` subset), the descope is recorded in progress.md, and no rule is left both scoped AND unreachable.

### Scenario 4 — Token-budget guard confirms the diet is real (edge case)

- **Given** the always-loaded surface measured ≈ 44,101 rule-tokens (11 files) at baseline.
- **When** `go test ./internal/config/...` runs after scoping 6 files (≈ −18,367 tokens).
- **Then** `TestAlwaysLoadedTokenBudget` passes with a lower total, `TestAlwaysLoadedSurfaceEnumeration` passes with the recomputed reduced count, and `token_budget_guard.go` shows no diff.

## Edge Cases

- A SCOPE file already carries a `description:` frontmatter key → add `paths:` as a sibling key (do not clobber `description:`).
- A cross-reference uses a bare filename (no path) → still resolves; self-referential paths keep the file in place.
- Windows path globs → the `**/<filename>.md` form is platform-neutral (matches the in-repo precedent).

## Quality Gate Criteria

- `go build ./...`, `go vet ./...`, `go test ./internal/config/...` all exit 0.
- Byte-identical template/live parity for all 6 scoped files.
- 5 KEEP files unchanged; guard source unchanged.

## Definition of Done

- [ ] 6 SCOPE files carry self-referential `paths:` in BOTH trees (byte-identical).
- [ ] 5 KEEP files carry NO `paths:`.
- [ ] LIVE 11→5 and TEMPLATE 11→5 always-loaded (command-verified).
- [ ] Mitigation retrieval summaries confirmed present in KEEP surfaces (or affected file descoped).
- [ ] Token-budget guard + enumeration test PASS; guard source untouched.
- [ ] All inbound cross-references still resolve.
- [ ] `make build` re-embed done; `go build`/`go vet`/config tests green.

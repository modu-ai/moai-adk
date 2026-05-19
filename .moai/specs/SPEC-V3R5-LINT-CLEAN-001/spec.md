---
id: SPEC-V3R5-LINT-CLEAN-001
title: "Agent Lint Baseline Cleanup (v3.5.0 Quality Debt Reduction)"
version: "0.2.0"
status: draft
created: 2026-05-19
updated: 2026-05-19
author: GOOS Kim
priority: P1
phase: "v3.5.0"
module: "internal/template/templates/.claude/agents/moai, .claude/agents/moai"
lifecycle: spec-anchored
tags: "lint, quality-debt, v3.5.0, baseline-reduction, agent-lint, moai-agent-lint"
---

# SPEC-V3R5-LINT-CLEAN-001 — Agent Lint Baseline Cleanup

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-19 | GOOS Kim | Initial draft — v3.5.0 Mega-Sprint quality-debt cleanup SPEC. Targets `moai agent lint --strict` baseline reduction from 176 (post-W0) → 0 across 4 phases, while preserving W0 delta-only D6 NEW=0 semantics. Out-of-band from Mega-Sprint W1/W2/W3/W4 critical-path; mergeable in parallel where file-scope-safe. |
| 0.2.0 | 2026-05-19 | GOOS Kim | plan-auditor iter 1 REVISE (harmonic mean 0.7229, D5 Cross-SPEC Traceability=0.45). Findings 1-14 resolved: (1) W2-deferred count fixed to canonical **13** (set size) / **12** (post-Phase-2 residual); (2-3) arithmetic recomputed with sum-check (60+4+30+70=164, residual=12); (4) intra-SPEC waves renamed to LCLN Phase 1-4 to disambiguate from Mega-Sprint Waves W0-W4; (5) design.md §1.1 FROZEN provenance reframed as superset with per-path rationale; (6) AC-LCLN-007.1 added for template-first invariant; (7) AC-005.2 bound tightened to [11, 16]; (8-14) minor strengthenings (exact equality checks, git worktree coverage, admin override scoping, golangci-lint conditional baseline, per-rule contradiction check, moai update --dry-run post-condition, REQ-LCLN-013 EARS clarification). |

## 1. Goal

Reduce the `moai agent lint --strict` baseline from **176 findings** (93 errors + 83 warnings, captured on `main` HEAD `02b2bb0a3`, 2026-05-19) to **0 findings**, while:

1. **Preserving delta-only D6 NEW=0 semantics** — every LCLN-Phase PR introduces zero new lint findings vs the prior-phase baseline.
2. **Avoiding Mega-Sprint W2 entanglement** — explicitly defer **13** LR-08 findings (the canonical W2-deferred set, see §3 Glossary and `research.md` §2.2) that will dissolve when SPEC-V3R5-CORE-SLIM-001 retires `expert-{backend,frontend,mobile}` and `moai-domain-{backend,frontend,database}`. After LCLN-Phase 2 deletes live `expert-mobile.md`, **12** of these 13 remain as the post-this-SPEC residual; the 13th (LR-08 on expert-mobile.md) is resolved in-flight by Phase 2.
3. **Respecting FROZEN zones** — no edits to zone-registry-enumerated paths nor to operational invariant files documented in `design.md` §1.1 Appendix A.
4. **Holding orthogonal lint surfaces at 0** — `golangci-lint run ./...`, `moai spec lint --strict`, and `moai workflow lint` remain at 0 throughout the SPEC lifecycle.

The cleanup is a **prerequisite for high-fidelity delta verification** across the remainder of the v3.5.0 Mega-Sprint. With 176 pre-existing findings, downstream Mega-Sprint SPECs (W1 CONSTITUTION-DUAL, W2 CORE-SLIM, W3 HARNESS-AUTONOMY, W4 PROJECT-MEGA) cannot cleanly assert "NEW_COUNT=0" without inheriting noise. This SPEC clears the slate.

### Note on baseline source

Memory `project_v3r5_w0_lifecycle_complete` referenced "321 findings (237 ERROR + 84 WARN)". Empirical re-measurement on plan-phase entry shows **176** (post-W0). The original 321 is pre-W0; W0 PR #1005 hotfix reduced LR-07 v2 fingerprint by 141. The relevant linter is **`moai agent lint --strict`**, not `golangci-lint` (which is already at 0 under the project's default config). See `research.md` §0 for full disambiguation.

## 2. Scope

### 2.0 Glossary — Disambiguation of "Wave" Terminology

This SPEC uses two distinct wave concepts that share the letter "W". To prevent collision (per plan-auditor iter-1 Finding 4):

| Term | Meaning | Examples |
|------|---------|----------|
| **Mega-Sprint Wave (W0, W1, W2, W3, W4)** | The five v3.5.0 release SPECs defined in `.moai/research/harness-autonomy-vision-2026-05-18.md` §5 | W0 = SPEC-V3R5-CLAUDE-REFRESH-001 (COMPLETE), W1 = CONSTITUTION-DUAL-001, W2 = CORE-SLIM-001, W3 = HARNESS-AUTONOMY-001, W4 = PROJECT-MEGA-001 |
| **LCLN Phase (Phase 1, Phase 2, Phase 3, Phase 4)** | This SPEC's internal cleanup phases, executed sequentially within this SPEC's run-phase | Phase 1 = D1+D2+D4 frontmatter & description hygiene; Phase 2 = D7 live `expert-mobile.md` deletion; Phase 3 = D3 tool-boundary cleanup; Phase 4 = D5 preload drift W4-resolvable subset |

Throughout this SPEC's documents (`spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`):

- Phrases like "**W2-deferred**" / "**Mega-Sprint W2**" refer to the Mega-Sprint Wave W2 (CORE-SLIM-001).
- Phrases like "**LCLN-Phase 2**" / "**Phase 2**" refer to this SPEC's internal cleanup phase 2.
- The string "W1-LCLN", "W2-LCLN", etc. (used in plan-auditor iter-1 artifacts) is REJECTED in iter-2; replaced with explicit "LCLN-Phase N" or "Phase N".
- The token "Wave" alone is ambiguous and is NOT used standalone; always paired with a qualifier ("Mega-Sprint Wave" or "LCLN Phase").

The 4 LCLN-Phases map 1:1 to the four LCLN-Phase PRs that this SPEC will produce. Mega-Sprint Wave references appear only in dependency discussions (§7 Dependencies) and W2-deferred bookkeeping.

### In Scope

- `internal/template/templates/.claude/agents/moai/*.md` — canonical agent definitions (template-first per CLAUDE.local.md §2)
- `.claude/agents/moai/*.md` — live agent definitions (sync target via `make build` + `moai update`)
- LR-01 (28), LR-02 (3), LR-03 (25), LR-05 (2), LR-06 (29), LR-12 (6) → 93 errors. Of these, 4 reside on `.claude/agents/moai/expert-mobile.md` (live, deleted in LCLN-Phase 2) and the remaining 89 are addressed by LCLN-Phases 1 + 3.
- LR-08 (83) → split into **W4-resolvable subset (70)** addressed by LCLN-Phase 4 + **W2-deferred set (13)** explicitly deferred to Mega-Sprint W2 SPEC-V3R5-CORE-SLIM-001. After LCLN-Phase 2 deletes live `expert-mobile.md`, 1 of the 13 W2-deferred findings (LR-08 on expert-mobile.md) is cleared, leaving 12 as the post-this-SPEC residual.

### Out of Scope

- **Go source code** (`internal/`, `pkg/`, `cmd/`) — `golangci-lint` baseline already 0; no action needed. Future Go-lint expansion belongs to a separate SPEC (e.g., `SPEC-V3R5-GO-LINT-EXPAND-001`) that proposes `.golangci.yml` adoption. See C8 below.
- **SPEC frontmatter under `.moai/specs/`** — `moai spec lint --strict` already 0; this SPEC must NOT introduce SPEC drift.
- **W2-deferred LR-08 findings** — 13 LR-08 drift items referencing skills/agents that Mega-Sprint W2 SPEC-V3R5-CORE-SLIM-001 will retire. Explicitly documented in `research.md` §2.2 as W2-DEFERRED. Resolving them here is wasted effort.
- **FROZEN zones** — guarded paths superset, enumerated in `design.md` §1.1 + Appendix A. Includes zone-registry `Frozen` entries plus operational invariants (askuser-protocol.md, zone-registry.md, agent-authoring.md, agent_lint.go).
- **`.claude/rules/moai/` rule files** in general — this SPEC does NOT modify rule documents
- **CLAUDE.md, CLAUDE.local.md, README.md** — not in cleanup scope
- **Source `internal/cli/agent_lint.go`** — the linter itself is correct; this SPEC fixes the agents it lints, not the linter

## 3. Requirements (EARS)

### Ubiquitous Requirements

- **REQ-LCLN-001**: The system SHALL maintain `moai agent lint --strict` exit code 0 on `main` HEAD after all LCLN-Phases merge AND Mega-Sprint W2 SPEC-V3R5-CORE-SLIM-001 lifecycle COMPLETEs.
- **REQ-LCLN-002**: The system SHALL preserve `golangci-lint run ./...` exit code 0 on `main` HEAD throughout the SPEC lifecycle.
- **REQ-LCLN-003**: The system SHALL preserve `moai spec lint --strict` exit code 0 on `main` HEAD throughout the SPEC lifecycle.
- **REQ-LCLN-004**: The system SHALL preserve `moai workflow lint` exit code 0 on `main` HEAD throughout the SPEC lifecycle.

### Event-Driven Requirements

- **REQ-LCLN-005** (delta semantics): WHEN any LCLN-Phase PR opens, the phase's pre-merge `moai agent lint --strict --format=json` output SHALL be diffed against the prior-phase's baseline JSON, AND the count of NEW findings (findings present post-phase but absent pre-phase, keyed by `file+line+rule`) SHALL equal 0.
- **REQ-LCLN-006** (forward progress): WHEN any LCLN-Phase merges to `main`, the post-merge total finding count SHALL equal the pre-phase total finding count minus the phase's sum-check reduction target (60/4/30/70 per plan.md §3).
- **REQ-LCLN-007** (template-first): WHEN a fix targets an agent definition, the edit SHALL be applied to `internal/template/templates/.claude/agents/moai/<agent>.md` first, then propagated via `make build` + sync, NOT directly to the live `.claude/agents/moai/<agent>.md`.

### State-Driven Requirements

- **REQ-LCLN-008** (FROZEN guard): WHILE any LCLN-Phase is in progress, the system SHALL NOT modify any file in the guarded-paths set (design.md §1.1 + Appendix A; superset of zone-registry-enumerated FROZEN paths).
- **REQ-LCLN-009** (Mega-Sprint W2 deferral): WHILE the Mega-Sprint W2 SPEC-V3R5-CORE-SLIM-001 is in `draft`, `planned`, or `in-progress` status, this SPEC SHALL NOT fix LR-08 findings whose drift skill is in {`moai-domain-backend`, `moai-domain-frontend`, `moai-domain-database`} OR whose affected agent is in {`expert-backend`, `expert-frontend`, `expert-mobile`}. The empirically-measured W2-deferred set has exactly **13** elements (canonical, captured 2026-05-19 on `main` HEAD `02b2bb0a3`); after LCLN-Phase 2 deletes live `expert-mobile.md` (clearing 1 of the 13), 12 remain as the post-Phase-4 residual.

### Unwanted Requirements

- **REQ-LCLN-010**: The system SHALL NOT introduce new SPEC frontmatter drift (i.e., `moai spec lint --strict` MUST NOT regress from 0).
- **REQ-LCLN-011**: The system SHALL NOT use `gofmt`, `goimports`, or any automated formatter/fixer that modifies files outside the SPEC's declared scope. All edits are explicit `Edit`-tool operations with reviewed diffs.
- **REQ-LCLN-012**: The system SHALL NOT modify `internal/cli/agent_lint.go` to weaken or suppress any LR-XX rule. The linter is the contract; agents conform to the linter.

### Optional Requirements

- **REQ-LCLN-013** (advisory safety net): WHERE per-package Go test coverage exists prior to a phase, the phase SHOULD NOT cause coverage to drop by more than 1.0 percentage point on any touched package. **Status: ADVISORY** — this is a `SHOULD`-strength requirement because agent-definition edits do not affect Go test coverage in normal operation. AC-LCLN-004.2 is correspondingly classified as **ADVISORY** (NOT in the must-pass Definition of Done list); coverage regression triggers a warning + acknowledgment, not a merge block. See `acceptance.md` § Definition of Done for the must-pass set.

## 4. Acceptance Criteria (EARS Hierarchical)

See `acceptance.md` for the full hierarchical AC tree with concrete shell commands. Summary structure:

- **AC-LCLN-001** (top-level): `moai agent lint --strict` exit 0 on `main` HEAD after all phases AND Mega-Sprint W2 merge.
  - AC-LCLN-001.1: Baseline captured before each LCLN-Phase.
  - AC-LCLN-001.2: Each LCLN-Phase reduces a contiguous subset of findings (per-phase reduction targets per plan.md §3 sum-check).
  - AC-LCLN-001.3: Final post-merge `./bin/moai agent lint --strict --format=json | jq '.summary.total == 0'` returns `true` (after Mega-Sprint W2 dissolves the 12-finding residual).
- **AC-LCLN-002** (delta semantics): NEW finding count vs prior baseline = 0 on every LCLN-Phase PR.
  - AC-LCLN-002.1: Pre-phase baseline JSON committed to `.moai/state/lint-baseline-pre-LCLN-P<N>.json`.
  - AC-LCLN-002.2: Diff command emits NEW count keyed by `file+line+rule`.
  - AC-LCLN-002.3: NEW count = 0 enforced as LCLN-Phase PR merge gate.
- **AC-LCLN-003** (scope discipline): No FROZEN zone touched, no orthogonal lint regression.
  - AC-LCLN-003.1: `moai spec lint --strict` exit 0 post-phase.
  - AC-LCLN-003.2: `golangci-lint run ./...` exit 0 post-phase (conditional on default linter set per C8).
  - AC-LCLN-003.3: `moai workflow lint` exit 0 post-phase.
  - AC-LCLN-003.4: No diff in design.md §1.1 + Appendix A guarded paths (automated Frozen Guard script audit).
- **AC-LCLN-004** (advisory) — coverage preservation: `go test ./...` passes on every LCLN-Phase PR (AC-LCLN-004.1 is must-pass; AC-LCLN-004.2 coverage delta is advisory only).
- **AC-LCLN-005** (Mega-Sprint W2 deferral hygiene): W2-deferred set documented (13 elements), audited bound `[11, 16]` (= 13 ± 3 upstream drift tolerance), and not modified by this SPEC's edits.
- **AC-LCLN-007** (template-first invariant): For every LCLN-Phase PR, edits to `internal/template/templates/.claude/agents/moai/*.md` mirror edits to `.claude/agents/moai/*.md` (or vice-versa for Phase 2 live-only deletion), AND `internal/template/embedded.go` is regenerated in the same PR.

## 5. Constraints

- **C1 (Template-First)** — All edits MUST flow through `internal/template/templates/.claude/agents/moai/`. After each edit, run `make build` to regenerate `internal/template/embedded.go`. Live `.claude/agents/moai/` updates via `moai update` (or are committed alongside template changes when the live copy is out of sync with template — current state for `expert-mobile.md`).
- **C2 (No auto-fix)** — No use of `gofmt`, `goimports`, `golangci-lint --fix`, or similar batch-rewriters. Each finding fixed via explicit `Edit`-tool operation with reviewed before/after.
- **C3 (No FROZEN modification)** — Frozen Guard pattern (design.md §1) MUST be enforced at the start of every LCLN-Phase. Any edit to a path in the guarded-paths set (design.md §1.1 + Appendix A) aborts the LCLN-Phase.
- **C4 (Mega-Sprint W2 deferral)** — LR-08 findings whose drift skill matches `moai-domain-{backend,frontend,database}` OR whose affected agent is `expert-{backend,frontend,mobile}` MUST be tagged W2-DEFERRED in the LCLN-Phase PR description, NOT fixed in this SPEC (with the in-flight exception of LCLN-Phase 2's expert-mobile.md deletion).
- **C5 (Orthogonal-lint preservation)** — Before opening each LCLN-Phase PR, run all four lints (`moai agent lint --strict`, `moai spec lint --strict`, `moai workflow lint`, `golangci-lint run ./...`). Three must remain at 0; only `moai agent lint --strict` should be decreasing.
- **C6 (Branch and PR naming)** — Per CLAUDE.local.md §18.2: branches named `chore/SPEC-V3R5-LINT-CLEAN-001-phase-<N>` or `fix/SPEC-V3R5-LINT-CLEAN-001-phase-<N>`. PRs use Conventional Commits format.
- **C7 (No time estimates)** — Per CLAUDE.local.md §6 and agent-common-protocol Time Estimation rule, no calendar/duration estimates in plan.md or acceptance.md. Priority labels and phase ordering only.
- **C8 (golangci-lint baseline conditionality)** — Adoption of `.golangci.yml` all-linters config is OUT OF SCOPE of this SPEC. The current 0-baseline assumption for `golangci-lint run ./...` (REQ-LCLN-002, AC-LCLN-003.2) is conditional on the project-default linter set (6 linters: errcheck, govet, ineffassign, staticcheck, typecheck, unused — see `research.md` §0). IF any concurrent SPEC introduces `.golangci.yml` with expanded linter selection during LCLN-Phases 1-4, this SPEC MUST re-baseline `golangci-lint` and capture the new baseline alongside `moai agent lint` baselines. Detection: at each LCLN-Phase entry, the orchestrator runs `[ -f .golangci.yml ] && [ -f .golangci.yaml ] && exit 0 || (echo "still default-config"; exit 0)`. If the file appears, halt and re-plan.

## 6. Exclusions (What NOT to Build)

- **Not a Go lint expansion** — this SPEC does NOT add a `.golangci.yml` config nor enable additional Go linters. Go source remains at the project-default linter set.
- **Not a rule weakening** — this SPEC does NOT modify any LR-XX rule in `internal/cli/agent_lint.go`, nor add `// nolint` directives, nor add per-rule allowlists.
- **Not a refactor** — agent definitions are edited only to the minimal extent needed to satisfy LR-XX rules. No restructuring of skill preload lists beyond drift-resolution, no description rewrites beyond `--deepthink` boilerplate removal.
- **Not a SPEC document rewrite** — `.moai/specs/` and `moai spec lint --strict` are out of scope; no SPEC frontmatter or content edits.
- **Not a documentation update** — `.claude/rules/moai/`, `CLAUDE.md`, `CLAUDE.local.md` are out of scope.
- **Not a W2 substitute** — this SPEC does NOT retire any agent or skill. W2-deferred findings remain.
- **Not a CI gate change** — `.github/workflows/` are not modified. The existing `moai agent lint --strict` step in CI will naturally start passing as findings drop to 0.

## 7. Dependencies

- **Upstream (blocking)**: Mega-Sprint W0 (SPEC-V3R5-CLAUDE-REFRESH-001) — already COMPLETE. Provides the post-W0 baseline of 176 that this SPEC reduces.
- **Parallel (non-blocking)**: Mega-Sprint W1 (CONSTITUTION-DUAL-001), W3 (HARNESS-AUTONOMY-001), W4 (PROJECT-MEGA-001) — file-scope-orthogonal. This SPEC's LCLN-Phases 1-4 can proceed in parallel branches with these Mega-Sprint W-series SPECs; rebase-onto-main before each LCLN-Phase PR opens (per AC-LCLN-005.1).
- **Downstream entanglement**: Mega-Sprint W2 (CORE-SLIM-001) — owns **13** W2-deferred LR-08 findings (canonical, see §2 Glossary and `research.md` §2.2). After this SPEC's LCLN-Phase 2 deletes live `expert-mobile.md` (clearing 1 of the 13), 12 remain as the post-this-SPEC residual. When Mega-Sprint W2 lifecycle COMPLETEs, those 12 will auto-resolve. This SPEC reaches "final 0/0 baseline" only after both this SPEC and Mega-Sprint W2 merge.

## 8. References

- `research.md` (this SPEC) — baseline analysis, methodology, discrepancies-vs-memory
- `plan.md` (this SPEC) — LCLN-Phase decomposition, sequencing decision, technical approach per phase
- `acceptance.md` (this SPEC) — testable AC leaves with shell commands and pass conditions
- `design.md` (this SPEC) — 5-Layer Safety Pipeline compatibility
- `.moai/specs/SPEC-V3R5-CLAUDE-REFRESH-001/` — W0 precedent for delta-only D6 NEW=0 semantics (AC-CLR-008)
- `.moai/research/harness-autonomy-vision-2026-05-18.md` §5 — Mega-Sprint W0..W4 roadmap
- `.claude/rules/moai/core/zone-registry.md` — FROZEN zone enumeration (CONST-V3R2-001..152; this SPEC's guarded paths are a superset, see `design.md` §1.1 Appendix A)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — canonical 12-field frontmatter schema
- `.claude/rules/moai/workflow/spec-workflow.md` — plan-phase workflow contract
- `CLAUDE.local.md` §2 (Template-First), §18 (Git Workflow), §22 (Dev settings intent)
- `internal/cli/agent_lint.go` — LR-01..LR-14 rule definitions
- Memory entry `project_v3r5_w0_lifecycle_complete` — Goal Stop hook 10/10 reference for v3.5.0 release-readiness criteria

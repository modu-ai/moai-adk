---
id: SPEC-V3R6-DOCS-RC2-README-001
title: "Progress — v3.0.0-rc2 README + CHANGELOG factual-alignment"
version: "0.3.0"
status: in-progress
created: 2026-06-19
updated: 2026-06-19
author: manager-spec
priority: P1
phase: "v3.0.0-rc2"
module: "repo-root docs"
lifecycle: spec-anchored
tags: "docs, readme, changelog, progress, v3r6, rc2"
era: V3R6
tier: M
---

# Progress — SPEC-V3R6-DOCS-RC2-README-001

## §A. Current Phase

**Run-phase** (status: draft → in-progress on M1 commit). Plan-phase complete; spec.md v0.3.0, plan.md, acceptance.md authored + iter-3 plan-audit PASS-WITH-DEBT 0.88.

Implementation Kickoff Approval (§19.1): **GRANTED** 2026-06-19 (plan-auditor verdict surfaced + plan summary presented; user approved run-phase entry; skip-eligible NO since 0.88 < 0.90 → explicit approval required).
Phase 0.95 Mode Selection: **sub-agent (Mode 5)** — see §E.0 below.
Rebase onto origin/main: divergence `13 2` → `0 2` (DSR 13 commits absorbed, scope independent). Backup: `backup/pre-rebase-docs-rc2-e8771c318`.

## §B. Milestone Status

| Milestone | Scope | Status | Commit SHA |
|-----------|-------|--------|------------|
| M1 | README.md (EN) quantitative fixes | pending (plan-phase) | — |
| M2 | README.md Mermaid + /moai db + expert-frontend + lifecycle | pending (plan-phase) | — |
| M3 | README.ko.md mirror + What's New rewrite | pending (plan-phase) | — |
| M4 | CHANGELOG rc2 promotion + mis-label cleanup | pending (plan-phase) | — |
| M5 | CLAUDE.md builder-harness path fix | pending (plan-phase) | — |
| M6 | Self-verification (grep AC suite) | pending (plan-phase) | — |

## §C. Decision Log

- **2026-06-19** — SPEC created at plan-phase. Decomposition self-check PASS. Ground-truth baseline verified against filesystem (`v3.0.0-rc2` confirmed, builder-harness path confirmed, 7 MoAI agents confirmed).
- **2026-06-19** — `era: V3R6` set explicitly in frontmatter to avoid H-6 unclassified fallback (per lifecycle-sync-gate.md § Frontmatter Era Field Semantics scenario 3).
- **2026-06-19** — Template counterpart decision (CLAUDE.md template source) DEFERRED to run-phase per plan.md §G.1.
- **2026-06-19 (iter-3 plan-audit fix)** — iter-3 plan-audit fix applied per plan-auditor iter-2 FAIL 0.74 verdict (2 BLOCKING + 2 SHOULD-FIX). D1' skills 32→31 LIVE (root cause = DSR rebase `SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001` deleted `moai-design-system` post-authoring → the iter-1 count of 32 became stale carry-over in spec.md §B Ground Truth, a `verification-claim-integrity.md` §2 attribution violation). D2' AC-KO-001b line-anchored on `README.ko.md:111` (KO D6 discipline parity — L58/126/560/1215 carry the CORRECT `16개` token, so the KO whole-file grep was vacuously satisfied). D3 precise Go LOC dropped from §B (pipeline over-counts non-deterministically; 191,248 LIVE). D5 AC-CL-001 internal contradiction split into two binary ACs (AC-CL-001a + AC-CL-001b). Frontmatter version 0.2.0 → 0.3.0; status unchanged (draft). spec.md / plan.md / acceptance.md updated in lockstep; README.md / README.ko.md / CHANGELOG.md / CLAUDE.md NOT touched (run-phase deliverables).
- **2026-06-19** — rebase onto origin/main: divergence `13 2` → `0 2` (clean ahead). DSR 13 commits (SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001, e453bcc86~935b2fc07) absorbed via `git rebase origin/main`; scope independent (DSR touched README.md/README.ko.md/CHANGELOG.md/CLAUDE.md 0 times). Backup branch `backup/pre-rebase-docs-rc2-e8771c318`; reflog preserves pre-rebase HEAD (ecd270369). Root cause of D1' drift traced here: DSR M1 deleted moai-design-system → skills 32→31 between spec authoring and now.
- **2026-06-19** — plan-auditor iter-3 re-audit: **PASS-WITH-DEBT 0.88** (Tier M 0.80 cleared; Clarity 0.90 / Completeness 0.95 / Testability 0.88 / Traceability 0.95). All 4 iter-2 defects CONFIRMED FIXED via LIVE re-derivation. 2 MINOR nits remain (AC-EN-010 body H4 sequence; AC-KO-001b commentary token) — non-blocking debt. skip-eligible NO (0.88 < 0.90). Report: `.moai/reports/plan-audit/SPEC-V3R6-DOCS-RC2-README-001-2026-06-19-iter3.md`.
- **2026-06-19** — Implementation Kickoff Approval (§19.1) **GRANTED** by user. plan-auditor iter-3 verdict + plan summary surfaced via AskUserQuestion; user selected "run-phase 진입 승인". Phase 0.5 arc: iter-1 FAIL 0.74 → iter-2 FAIL 0.74 → iter-3 PASS-WITH-DEBT 0.88 (monotonic +0.14).
- **2026-06-19** — Phase 0.95 Mode Selection: **sub-agent (Mode 5)** — sequential manager-develop M1→M6 (single coupled repo-root docs domain, 4 files sharing one §B baseline). Full evaluation: §E.0.

## §D. Preserved List (Abort Recovery)

_(populated at run-phase if a run is aborted mid-milestone; empty at plan-phase)_

## §E.0 — Phase 0.95 Mode Selection

**Decision: sub-agent (Mode 5)**

Input parameters:
- tier: M
- scope: 4 files (README.md, README.ko.md, CHANGELOG.md, CLAUDE.md §4)
- domain count: 1 (repo-root docs factual-alignment)
- file language mix: 100% markdown (zero Go code)
- concurrency benefit: LOW (single tightly-coupled doc domain; EN/KO/CHANGELOG/CLAUDE.md share one §B ground-truth baseline — 8 agents, 31 skills, 100K+ LOC, 16 langs, 17 commands)
- Agent Teams prereqs: NOT all met (single domain, 4 files — below the 3-domain OR 10-file scope gate)

Mode evaluation:
| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | 20 REQs / 30 ACs / 6 milestones — non-trivial |
| 2 background | no | Write/Edit required — not read-only |
| 3 agent-team | no | 1 domain, 4 files — below scope gate |
| 4 parallel | no | doc editing, not multi-domain research; coding-task parallelism caveat applies |
| 5 sub-agent | **yes** | sequential manager-develop M1→M6; single coupled doc domain favors sequential over parallel |
| 6 workflow | no | 4 files, not ≥30 mechanical-uniform |

Justification: SPEC A is a single-domain repo-root docs factual-alignment task across 4 tightly-coupled markdown files sharing one §B baseline. Sequential manager-develop per-milestone delegation (M1→M6) is the correct fit: the EN/KO mirror + CHANGELOG + CLAUDE.md edits must stay consistent with one baseline, so parallel fan-out would risk cross-file ground-truth drift (the very hazard this SPEC exists to fix). Mode 5 is the Anthropic-recommended default for non-research-parallel work.

## §E.1 Plan-phase Audit-Ready Signal

- [x] SPEC ID decomposition self-check printed (spec.md §J): `SPEC ✓ | V3R6 ✓ | DOCS ✓ | RC2 ✓ | README ✓ | 001 ✓ → PASS`
- [x] All 12 canonical frontmatter fields present in spec.md.
- [x] `era: V3R6` set in frontmatter.
- [x] §H Exclusions contains 6 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (satisfies `OutOfScopeRule`).
- [x] Every REQ cites a DRIFT-ID; every DRIFT-ID cites a file:line.
- [x] No Go code edits in any milestone.
- [x] No template edits in plan-phase.
- [x] `v3.0.0-rc2` wording consistent; no `v3.0.0 stable` claim.
- [x] No superseded-SPEC citation (REQ-X-002).
- [x] Draft SPECs not asserted finalized (REQ-X-003).
- [x] 4 plan-phase artifacts present: spec.md, plan.md, acceptance.md, progress.md.

## §E.2 Run-phase Evidence

### Milestone Status (post-run)

| Milestone | Scope | Status | Commit SHA |
|-----------|-------|--------|------------|
| M1 | README.md (EN) quantitative fixes | done | `93b195fc4` |
| M2 | README.md Mermaid + /moai db + expert-frontend + lifecycle | done | `1dbfc0a4c` |
| M3 | README.ko.md mirror + What's New rewrite | done | `1d574bd05` |
| M4 | CHANGELOG rc2 promotion + mis-label cleanup | done | `3313364ed` |
| M5 | CLAUDE.md builder-harness path fix (project + template) | done | `311d980d1` |
| M6 | Self-verification (grep AC suite) | done | (this commit) |

Commit SHAs are post-rebase values (worktree rebased onto advanced origin/main before M6; pre-rebase backup ref `backup/pre-rebase-docs-rc2-m6`).

### AC Binary PASS/FAIL Matrix (E1) — 30 ACs

| AC | Verdict | Grep command | Observed |
|----|---------|--------------|----------|
| AC-EN-001a | PASS | `grep -c "24 specialized" README.md` | `0` |
| AC-EN-001b | PASS | `sed -n '40p' README.md \| grep -cE "8 retained\|8 (AI )?agents"` | `1` |
| AC-EN-002a | PASS | `grep -c "26 agents\|26\b.*specialized" README.md` | `0` |
| AC-EN-002b | PASS | `grep -c "47 skills" README.md` | `0` |
| AC-EN-002c | PASS | `sed -n '60,70p' README.md \| grep -c "30"` + moai-* qual | `1`, `1` |
| AC-EN-003a | PASS | `grep -c "delegates to 24 specialized agents" README.md` | `0` |
| AC-EN-003b | PASS | `sed -n '262p' README.md \| grep -cE "8 retained\|8 (AI )?agents"` | `1` |
| AC-EN-004 | PASS | `grep -c "38,700" README.md` + `grep -cE "100K\+ lines"` | `0`, `1` |
| AC-EN-005 | PASS | `grep -c "18 languages\|18 programming"` + `sed -n '60,70p' \| grep -c "16"` | `0`, `1` |
| AC-EN-006a | PASS | `sed -n '264,286p' \| grep -cE 'Manager \(8\)\|...'` + archived body | `0`, `0` |
| AC-EN-006b | PASS | `sed -n '264,286p' \| grep -cE "manager-spec\|..."` | `8` |
| AC-EN-007a | PASS | `grep -cE "^#+ .*/moai db"` + `grep -cE "/moai db"` | `0`, `0` |
| AC-EN-007b | PASS | `grep -c "moai hook db-schema-sync"` | `1` |
| AC-EN-008 | PASS | `grep -n "expert-frontend"` + `sed -n '915,975p' \| grep -cE "manager-develop\|harness\|cycle_type"` | `0 lines`, `3` |
| AC-EN-009 | PASS | `sed -n '445,485p' \| grep -cE "3-phase\|..."` + SPEC citation | `1`, `1` |
| AC-EN-010 | PASS | `grep -cE "Opus 4\.[78]\|CC 2\.1\.17"` (option a) | `2` |
| AC-KO-001a | PASS | stale tokens each `grep -c` | all `0` |
| AC-KO-001b | PASS | drift line re-resolved (gone) + `16개`/db-schema-sync/30-스킬 | `4`, `1`, `1` |
| AC-KO-002a | PASS | `grep -cE "^#+ .*v2\.17\.0"` | `0` |
| AC-KO-002b | PASS | v3.0.0-rc2/V3R6/glm-5.2/8-agents/3-phase | `9`,`7`,`6`,`5` |
| AC-KO-003 | PASS | `my-harness-` + canonical namespace predicate | `0`, `1` |
| AC-CL-001a | PASS | `grep -cE "^## \[v3\.0\.0-rc2\]"` | `1` |
| AC-CL-001b | PASS-WITH-JUSTIFICATION (blocker) | `grep -cE "^## \[Unreleased\]"` | `30` (see §E.7) |
| AC-CL-002 | PASS | `grep -cE '^##.*v2\.20\.0-rc1'` | `0` |
| AC-CL-003 | PASS | `grep -cE "^## \[v3\.0\.0\]"` | `0` |
| AC-CLAUDE-001a | PASS | `grep -c "agents/builder/builder-harness"` | `0` |
| AC-CLAUDE-001b | PASS | `grep -c "agents/moai/builder-harness"` | `1` |
| AC-X-001 | PASS | `grep -rn "v3\.0\.0 stable"` (4 files) | `0` lines |
| AC-X-002 | PASS | `grep -rn "SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001"` | `0` lines |
| AC-X-003 | PASS (mooted) | RULES-SSOT-DEDUP-001 CHANGELOG entry | SPEC is `completed`; truthful phrasing correct (see §E.7) |

**Summary**: 29 PASS + 1 PASS-WITH-JUSTIFICATION (AC-CL-001b blocker) + 0 FAIL.

**Integration re-verification (2026-06-22)**: worktree rebased onto main (50-commit advance since base `9d9df059f`); all 30 ACs re-verified LIVE post-rebase. One stale value corrected: `moai-*` skills count **31 → 30** (1 moai-* skill removed during the 50-commit advance; re-verified `ls -d .claude/skills/moai-* | wc -l` = 30). Updated surfaces: README.md L40/L64, README.ko.md L99, acceptance.md AC-EN-002c + AC-KO-001b. All other ACs (agents 8 retained = 7 custom + Explore, version v3.0.0-rc2, builder-harness path, 16 languages, archived-mention context, Unreleased-30 justification) re-verified PASS unchanged.

### E2. Cross-platform build
`go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0. Zero Go source changed (docs-only).

### E3. Coverage
N/A — zero Go code changed.

### E4. Subagent boundary (C-HRA-008)
N/A — no Go source under `internal/harness/` or `internal/hook/` touched.

### E5. spec-lint
`moai spec lint .../spec.md` → `0 error(s), 1 warning(s)` (StatusGitConsistency — expected during run, resolves at close). F.4 bar (0 errors) met.

### AC-EN-006a companion — archived agent whole-file sweep
`grep -cE "manager-(strategy|quality|brain|project)" README.md` = `3`. These are Status-Transition-Ownership-Matrix / archived-agents enumeration mentions (legitimate "archived, MUST NOT be spawned" context), NOT active-catalog surfaceings — distinguished per acceptance.md edge case E.1. `grep -c "claude-code-guide" README.md` = `0`.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-19
run_commit_sha: "311d980d1"
run_status: audit-ready
ac_pass_count: 29
ac_fail_count: 0
ac_pass_with_justification_count: 1
preserve_list_post_run_count: 0
l44_pre_commit_fetch: true
l44_post_push_fetch: true             # post-push divergence 0 6 (clean); branch pushed to origin
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: exit_0
  windows: exit_0
total_run_phase_files: 6
m1_to_mN_commit_strategy: per-milestone commits; rebased onto origin/main before M6 (parallel RULES-* session 8 commits disjoint, clean rebase)
```

**Multi-session race note**: parallel session landed 8 commits (4 RULES-* SPEC sync closes + §E.4 backfills) mid-run; file sets disjoint; clean rebase absorbed. Backup ref `backup/pre-rebase-docs-rc2-m6`.

## §E.7 Run-phase Blocker Report

### Blocker 1 — AC-CL-001b over-broad predicate vs REQ-CL-001 scope

**The literal AC predicate cannot be satisfied without scope expansion the SPEC never authorized.**

- AC-CL-001b requires `grep -cE "^## \[Unreleased\]" CHANGELOG.md` = `0`; observed = `30`.
- The CHANGELOG uses `## [Unreleased] — <SPEC-ID>` as its established per-SPEC historical-cohort marker convention (30 headings, e.g. `## [Unreleased] — SPEC-V3R4-SPECLINT-DEBT-002`).
- REQ-CL-001 scoped ONLY the top block (now promoted → `## [v3.0.0-rc2]`, AC-CL-001a PASS). DRIFT-CL-01/02 enumerated only the top block + 4 v2.20.0-rc1 headings — NOT the 30 historical markers.
- Edge case E.2 contemplates a SINGLE top placeholder, not 30 markers that ARE the repo's changelog convention.

**Decision options** (surface to user via orchestrator):

| Option | Effect | Risk |
|--------|--------|------|
| A | AC-CL-001b = PASS-with-justification; SPEC closes | 30 markers remain (repo convention); future sweep may re-flag |
| B | Amend SPEC to scope AC-CL-001b to top-block-only (manager-spec re-delegation) | Lowest-risk; honors REQ intent |
| C | Full 30-heading demotion (scope expansion) | High-risk structural rewrite beyond factual-alignment; not recommended |

**Recommendation**: A or B. The factual-alignment intent (promote rc2 cohort, remove 4 mis-labeled v2.20.0-rc1 headings) is fully satisfied. The literal predicate is an iter-3 D5 split artifact that over-generalized "no top Unreleased" into "no Unreleased anywhere".

### Blocker 2 — Worktree isolation (informational)

Run executed in worktree `agent-a1651acae861e7fd2` (branch `worktree-agent-a1651acae861e7fd2`), NOT on `main` as spawn prompt Section A stated. Worktree HEAD was synced `0 0` at start. The 5 milestone commits are on the worktree branch; host checkout (`main`) must fast-forward/cherry-pick to absorb. Environment constraint (git dual-main-checkout prevention), not a SPEC defect.

### Blocker 3 — SPEC artifact reconciliation (resolved, informational)

At run start, worktree committed SPEC artifacts were v0.2.0 (iter-2) while shared checkout carried uncommitted v0.3.0 (iter-3, the PASS-WITH-DEBT 0.88 baseline). v0.3.0 artifacts were copied into the worktree so M6 ran against the correct iter-3 acceptance.md (AC-KO-001b L111 anchor, AC-CL-001 split, skills count 31). Resolved; no action needed.

### AC-X-003 note (REQ premise mooted, not a blocker)

REQ-X-003 required RULES-SSOT-DEDUP-001 / RULES-VERSION-FORMAT-001 to be phrased in-flight (draft). Post-rebase, RULES-SSOT-DEDUP-001 is `status: completed`. The CHANGELOG L12 entry truthfully describes it as completed (run+sync done, 15 ACs PASS). Factual-truth takes precedence over the stale REQ premise (asserting "draft" for a completed SPEC would violate verification-claim-integrity §1.1). AC-X-003 = PASS (mooted).

## §E.4 Sync-phase Audit-Ready Signal

_<pending run-phase>_

sync_commit_sha: ""

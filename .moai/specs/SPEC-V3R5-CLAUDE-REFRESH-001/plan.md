# Implementation Plan — SPEC-V3R5-CLAUDE-REFRESH-001

## 1. Strategy Overview

This SPEC is a **template-first refactor** with three logically-separable bundles:

1. **Bundle A (Settings Matchers)** — Pure template edit, 2 LOC, no downstream prompt impact.
2. **Bundle B (CLAUDE.md Architecture Truth)** — Documentation rewrite + small file deletion; downstream prompt impact on agents that read CLAUDE.md.
3. **Lint Verification** — Final gate via `moai agent lint --strict` (REQ-CLR-008).

The work is decomposed into 8 tasks (T1-T8), each mapped 1:1 to an AC for trace traceability. Tasks are executed in order: matchers first (lowest risk), then CLAUDE.md surgical edits, then file deletion, then lint verification. This ordering allows early CI signal on Bundle A while Bundle B work proceeds.

## 2. Task Decomposition

| Task | Maps to AC | File(s) Touched | Risk | Reversibility |
|------|-----------|-----------------|------|---------------|
| T0 | AC-CLR-008 (baseline) | None — baseline capture only (`moai agent lint --strict --format=json > /tmp/lint-baseline-w0.json`) | None | N/A |
| T1 | AC-CLR-001 | `settings.json.tmpl` line 6 | Low | High (1 LOC) |
| T2 | AC-CLR-002 | `settings.json.tmpl` line 81 | Low | High (1 LOC) |
| T3 | AC-CLR-003 | `CLAUDE.md` §5 (~30 LOC rewrite) | Medium | High (text only) |
| T4 | AC-CLR-004 | Delete `expert-mobile.md`; update CLAUDE.md §4 catalog | Low | Medium (file deletion) |
| T5 | AC-CLR-005 | `CLAUDE.md` §8 line 303 (1 LOC) | Low | High (1 LOC) |
| T6 | AC-CLR-006 | `CLAUDE.md` §1 + §8, `moai-constitution.md`, `agent-common-protocol.md` (~40 LOC removed) | Medium | High (text only) |
| T7 | AC-CLR-007 | `CLAUDE.md` footer + changelog entry (~10 LOC) | Low | High (text only) |
| T8 | AC-CLR-008 (delta verification) | None — verification only (`diff-lint-findings` against T0 baseline) | None | N/A |

**T0 (NEW, P0)** — Baseline capture (precondition for AC-CLR-008 delta verification):

```bash
mkdir -p /tmp
moai agent lint --strict --format=json > /tmp/lint-baseline-w0.json
# Commit baseline count to working memory (record in progress.md under "## Pre-Run Baseline" as the lint-baseline-w0 count, including the total findings, ERROR count, and WARN count).
```

AC-CLR-008 verification (T8) compares post-run lint output against `/tmp/lint-baseline-w0.json` and requires 0 NEW findings (pre-existing 321 findings preserved; no new findings introduced).

Task ordering rationale:
- T1, T2 first — Bundle A standalone, low blast radius, builds confidence.
- T3, T4 next — biggest architecture truth bundle, executed before AskUserQuestion compression so that §5 and §4 are stable when T6 references them.
- T5 before T6 — ToolSearch line is inside §8; fix syntax first so that compressed §8 does not preserve incorrect example.
- T6 last among edits — compression depends on T3-T5 being stable, so that the SSOT reference compression does not accidentally drop a now-corrected line.
- T7 last edit — version bump should reflect ALL completed edits; deferring to last ensures footer accurately describes shipped scope.
- T8 — pure verification gate after all edits, runs once at the end.

## 3. Milestones

| Milestone | Priority | Trigger Condition | Tasks Included |
|-----------|----------|-------------------|----------------|
| M1 — Settings Matchers Fixed | P0 | Bundle A complete | T1, T2 |
| M2 — Architecture Truth Restored | P0 | Bundle B §5 + agent retirement complete | T3, T4 |
| M3 — SSOT Compression Complete | P1 | ToolSearch + AskUserQuestion documentation cleaned | T5, T6 |
| M4 — Version Reconciled | P1 | CLAUDE.md footer + changelog updated | T7 |
| M5 — Lint Verified | P0 | `moai agent lint --strict` returns 0 ERROR / 0 WARN | T8 |

No time estimates per CLAUDE.md §Time Estimation rule — priority labels and phase ordering only.

## 4. Technical Approach

### Bundle A (T1 + T2) — Settings Matchers

**T1** — Edit `internal/template/templates/.claude/settings.json.tmpl` line 6:
- Before: `"matcher": "startup|resume",`
- After: `"matcher": "startup|resume|clear|compact",`
- Tool: `Edit` with unique surrounding context to disambiguate from other matcher lines.

**T2** — Edit same file, line 81:
- Before: `"matcher": "Write|Edit"`
- After: `"matcher": "Write|Edit|MultiEdit"`
- Tool: `Edit`. Sole `Write|Edit` matcher in the file (PostToolUse block).

Post-edit: regenerate `internal/template/embedded.go` via `make build` so that the change ships in the binary.

### Bundle B (T3-T7) — CLAUDE.md Architecture Truth

**T3** — CLAUDE.md §5 "Agent Chain for SPEC Execution" rewrite:
- Remove the 6-phase pipeline list (lines 152-159).
- Replace with a truth statement: "/moai run dispatches `manager-develop` as the sole implementer per `quality.yaml` `development_mode` (`ddd` or `tdd`). The agent applies the selected cycle (Red-Green-Refactor for TDD, or Analyze-Preserve-Improve for DDD) and may consult domain skills directly without spawning expert subagents."
- Add explicit dormancy footnote: "`expert-{backend,frontend}` are dormant in the auto-workflow but active in utility commands (`/moai fix`, `/moai loop`, `/moai mx`, `/moai review`, `/moai design`, `/moai e2e`). Use explicit invocation when needed: `Use the expert-backend subagent to ...`"
- Retain the section heading and any introductory paragraph so that anchors and cross-references remain stable.

**T4** — `expert-mobile` retirement:
- Delete `internal/template/templates/.claude/agents/moai/expert-mobile.md` via `Bash rm`.
- Update CLAUDE.md §4 Agent Catalog Expert Agents line to reflect the count change (currently `Backend, frontend, security, devops, performance, refactoring` — 6 items, no `mobile` listed; the line is already truthful but a footnote needs to be added: "expert-mobile retired in v3.5.0 (SPEC-V3R5-CLAUDE-REFRESH-001). Mobile development is handled by `my-harness-mobile-*` agents generated by `/moai project` for mobile-type projects.").
- Verification: `grep -rn moai-domain-mobile internal/template/templates/` MUST return zero lines (all 3 dangling refs were inside `expert-mobile.md` body — deleted with the file).

**T5** — CLAUDE.md §8 ToolSearch syntax:
- Locate the line containing `ToolSearch(query: "select:AskUserQuestion,TaskCreate,TaskUpdate,TaskList,TaskGet", max_results: 5)`.
- Remove `, max_results: 5` so the line matches the SSOT exactly: `ToolSearch(query: "select:AskUserQuestion,TaskCreate,TaskUpdate,TaskList,TaskGet")`.

**T6** — AskUserQuestion 5-way duplication compression:
- **T6.0 (baseline measurement, mandatory pre-edit step)**: Run `grep -nE 'AskUserQuestion|deferred tool|Deferred Tool' internal/template/templates/CLAUDE.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md | wc -l` to confirm baseline N at run-phase start. As of iter2 plan fix (2026-05-18), measured N=29. AC-CLR-006 target = ceil(N * 0.3) = ceil(29 * 0.3) = 9 lines combined across the 3 non-SSOT documents.
- **In-scope §1 HARD rules constraint**: ONLY the bullets mentioning 'AskUserQuestion-Only Interaction' or 'Deferred Tool Preload' (estimated 2 HARD bullets in §1) are in T6 scope. The other 9 HARD bullets in §1 (Language-Aware, Parallel Execution, No XML, Markdown Output, Context-First Discovery, Approach-First, Multi-File Decomposition, Post-Implementation Review, Reproduction-First) MUST NOT be modified by T6.
- SSOT (retain as-is): `.claude/rules/moai/core/askuser-protocol.md`
- CLAUDE.md §1 HARD rules: replace any procedural detail with a single line: "AskUserQuestion-Only Interaction (see `.claude/rules/moai/core/askuser-protocol.md`): ALL user-facing questions MUST go through `AskUserQuestion`."
- CLAUDE.md §8 User Interaction Architecture: keep the section heading and a 3-5 line summary; replace deeper procedural prose with: "Full protocol specification including preload sequence, Socratic interview structure, and anti-patterns: see `.claude/rules/moai/core/askuser-protocol.md` (SSOT)."
- `moai-constitution.md` §MoAI Orchestrator: replace the AskUserQuestion paraphrase with a single line: "AskUserQuestion is the orchestrator's exclusive user-question channel. See `.claude/rules/moai/core/askuser-protocol.md` for canonical procedure."
- `agent-common-protocol.md` §User Interaction Boundary: retain the brief asymmetric-boundary summary (subagent prohibition is load-bearing for that file's audience), but add an explicit SSOT citation: "Full preload sequence and Socratic interview structure: see `.claude/rules/moai/core/askuser-protocol.md`."
- Target reduction: combined paraphrase line count across the 3 non-SSOT documents from N=29 lines (measured iter2, 2026-05-18) to ≤ ceil(N * 0.3) = 9 lines (70% reduction). See T6.0 for baseline measurement command.

**T7** — Version bump and changelog:
- CLAUDE.md footer: change `Version: 14.0.0 (Agency v3.2 + Harness Design Integration)` (verified template HEAD at iter2 fix, 2026-05-18) to `Version: 14.2.0 (Architecture Truth + W0 Bundle A+B)`.
- CLAUDE.md footer: update the existing `Last Updated:` line to `Last Updated: 2026-05-18`.
- Add changelog entry after the existing v14.0.0 description block: "Changes in v14.2.0 (from v14.0.0): §5 Agent Chain rewritten to reflect `/moai run` runtime truth (manager-develop sole implementer; expert-{backend,frontend} dormant in auto-workflow); §4 Agent Catalog updated to note expert-mobile retirement; §8 ToolSearch syntax corrected to match askuser-protocol.md SSOT; §1 + §8 + cross-referenced rules compressed AskUserQuestion duplication to ≤2 locations (SSOT + brief references); settings.json.tmpl matchers extended (SessionStart += clear|compact; PostToolUse += MultiEdit)."
- **Critical**: T7 Edit `old_string` MUST be the literal current footer text starting with "Version: 14.0.0" (verified via `grep -n "^Version:" internal/template/templates/CLAUDE.md` at iter2 fix). If `make` or `git pull` between this plan write and run-phase has bumped the template version, re-verify before T7 Edit — Edit will FAIL on old_string mismatch otherwise.

### T8 — Lint Verification

- Run `moai agent lint --strict` from `/Users/goos/MoAI/moai-adk-go`.
- Required: exit code 0, stdout contains `✓ No findings`.
- If failures appear, diagnose each finding against the 8-rule lint engine (`internal/cli/agent_lint.go`) and fix root cause without weakening any rule.

## 5. Risks and Mitigations

| ID | Risk | Likelihood | Impact | Mitigation |
|----|------|------------|--------|------------|
| R1 | §5 Agent Chain rewrite confuses existing users who memorized the 6-phase narrative | Medium | Medium | Explicit migration note in changelog (T7) plus the dormancy table in §5 makes the new model self-explanatory. The 6-phase pipeline can optionally be re-framed as "Optional manual Agent Chain pattern" rather than fully removed, preserving familiarity. |
| R2 | AskUserQuestion duplication compression accidentally removes a nuance present in a paraphrase but missing from SSOT | Medium | High | Before each removal in T6, diff the paraphrase against the SSOT line-by-line and confirm full semantic coverage in SSOT. If a paraphrase contains a unique nuance, ADD it to the SSOT first (separate PR or pre-T6 commit), then compress. |
| R3 | CLAUDE.md edits conflict with parallel work on `main` | Low | Medium | Execute the SPEC run-phase when `main` is quiet (no other CLAUDE.md PRs OPEN). If conflict arises, rebase and re-apply Edits; the changes are localized to specific sections so conflict resolution is mechanical. |
| R4 | `expert-mobile` retirement breaks an undocumented user script that invokes the agent | Low | Low | No stub agent created (unlike planned W2 stubs for backend/frontend). Rationale: F-102 confirms zero workflow invocation and zero CLAUDE.md catalog reference; the agent was already silently dead. Mitigation if a user does invoke it after the fix: Claude Code returns a clear "agent not found" error rather than silent skill-load failure (current state), which is strictly better UX. |
| R5 | `moai agent lint --strict` fails on existing pre-conditions outside this SPEC's scope | Medium | Medium | Before starting T1, baseline the lint state from `main` HEAD to distinguish pre-existing failures from new ones introduced by this SPEC. If pre-existing failures block T8, document them in `progress.md` as out-of-scope and either (a) fix them inline if trivial, or (b) defer to a follow-up SPEC and adjust AC-CLR-008 verification to "no NEW findings introduced." |
| R6 | `make build` fails to regenerate `embedded.go` after settings.json.tmpl edit | Low | Medium | Run `make build` immediately after T2 to catch failures early. The build is required before T8 lint, since lint inspects template-rendered artifacts. |

## 6. Decision Log

| Decision | Rationale |
|----------|-----------|
| Hard delete `expert-mobile.md` with no stub (vs. W2's stub-redirect plan for backend/frontend) | F-102 confirms zero workflow invocations and CLAUDE.md catalog already excludes mobile (6 experts listed, no mobile). Stub redirect adds maintenance overhead for an agent with no documented users. |
| Version bump v14.0.0 → v14.2.0 (skip v14.1.0) | v14.1.0 is the currently-committed version (per the template footer at HEAD). v14.2.0 reflects the W0 cumulative refresh. This SPEC SHIPS v14.2.0; an interim v14.1.x is not produced. |
| Preserve `agent-common-protocol.md` §User Interaction Boundary summary (do not full-collapse to SSOT reference) | This file is the canonical agent-level protocol; subagent authors read it. The asymmetric-boundary rule (subagents MUST NOT prompt) is the LOAD-BEARING content for that audience and removing it would force agents to chase a cross-reference for a critical safety rule. Therefore we trim duplicated procedural detail but keep the boundary statement. |
| Tasks T1, T2 grouped as "Bundle A" but executed as separate atomic edits | Each is one LOC; separate Edits keep PR diff readable and allow cherry-pick if needed. Both edits target the same file in different sections so there is no Edit collision risk. |
| §4 Agent Catalog footnote (vs. complete catalog rewrite) for expert-mobile | The catalog already does not list mobile (it lists 6 experts); only a footnote is needed to document why one entry was removed from the count. Full rewrite is W2 scope. |
| Defer Frozen Guard / harness-learner / Core Slim to W1-W4 | This SPEC is a pure surgical refresh; runtime changes (W1+) require separate design + test scaffolding and should not be conflated with documentation truth restoration. |
| D5: ACs placed in `design.md` §2 instead of separate `acceptance.md` file | User prompt explicitly specified 3-file SPEC structure (spec.md/plan.md/design.md). Each AC retains hierarchical EARS format (Given/When/Then) with binary verification command. design.md §2 serves as canonical AC location for this SPEC; spec.md REQs cross-reference it. Resolves plan-auditor iter1 D2 (acceptance.md reference broken). |
| D6: AC-CLR-008 scoped to delta-only (no NEW findings) | Resolves plan-auditor iter1 D1. Rationale: 321 pre-existing findings (237 ERROR + 84 WARN as of 2026-05-18 baseline) are W2 scope — expert-{backend,frontend,mobile} retire dissolves the majority of LR-08 preload drift findings. W0 must not block on out-of-scope cleanup. AC-CLR-008 requires capturing a lint baseline at T0 and verifying 0 NEW findings post-run. |
| D7: Template CLAUDE.md is v14.0.0 (verified at iter2 fix via `grep -n "^Version:" internal/template/templates/CLAUDE.md`) | Resolves plan-auditor iter1 D3 (version baseline wrong). Project-root live CLAUDE.md may be at v14.1.0 or different version; SPEC scope is template only per Constraint C1. T7 Edit old_string sourced from template HEAD. Original plan assumed v14.1.0 baseline incorrectly; v14.0.0 → v14.2.0 bump skips v14.1.0 entirely. |

## 7. Validation Plan

### Pre-Run-Phase (Baseline)

1. **T0 (MANDATORY for AC-CLR-008)**: Capture lint baseline in JSON:
   ```bash
   mkdir -p /tmp
   moai agent lint --strict --format=json > /tmp/lint-baseline-w0.json
   moai agent lint --strict 2>&1 | tee /tmp/lint-baseline-w0.txt   # human-readable backup
   ```
   Expected at iter2 fix (2026-05-18): 321 total findings (237 ERROR + 84 WARN). AC-CLR-008 will compare post-run JSON against this baseline.
2. Confirm `expert-mobile.md` exists: `test -f internal/template/templates/.claude/agents/moai/expert-mobile.md && echo OK`
3. Capture AskUserQuestion paraphrase line count baseline: `grep -nE 'AskUserQuestion|deferred tool|Deferred Tool' internal/template/templates/CLAUDE.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md | wc -l`. Expected: N=29 at iter2 fix. AC-CLR-006 target = ceil(N * 0.3) = 9.
4. Confirm CLAUDE.md footer version: `grep "^Version:" internal/template/templates/CLAUDE.md`. Expected: `Version: 14.0.0 (Agency v3.2 + Harness Design Integration)` at iter2 fix.

### Per-Task Validation

After each task, run the AC verification command (see `design.md` §2 — each AC includes its verification command in the "Then" clause) for the corresponding AC.

### Post-Run-Phase (Gate)

1. AC-CLR-008 delta verification (REQ-CLR-008): `moai agent lint --strict --format=json | diff-lint-findings /tmp/lint-baseline-w0.json` SHALL report 0 NEW findings. Pre-existing 321 findings are preserved per Decision D6. If `diff-lint-findings` helper is unavailable, use `jq` to compute set difference: `jq -s '.[1].findings - .[0].findings | length' /tmp/lint-baseline-w0.json <(moai agent lint --strict --format=json)` SHALL equal 0.
2. `make build` succeeds (template regeneration).
3. `grep -rn moai-domain-mobile internal/template/templates/` returns empty (AC-CLR-004).
4. `grep "Version: 14.2.0" internal/template/templates/CLAUDE.md` returns 1 match (AC-CLR-007).

## 8. Out of Scope (Cross-Reference)

See `spec.md` §2 "Out of Scope" for the full deferred list. Notably:
- No runtime hook code change (W1).
- No `expert-{backend,frontend}` retirement (W2).
- No `manager-develop` Delegation Protocol edit (W2).
- No `harness-learner` mechanism (W3).
- No `meta-harness` workflow (W4).

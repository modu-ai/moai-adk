---
description: >
  Synchronizes documentation with code changes, verifies project quality,
  and finalizes pull requests. Third step of the Plan-Run-Sync workflow.
  Includes deep code review with auto-fix, coverage analysis with test generation,
  SPEC divergence analysis, project document updates, and Context Memory generation.
  Use when documentation sync, PR creation, or quality verification is needed.
user-invocable: false
metadata:
  version: "3.8.0"
  category: "workflow"
  status: "active"
  updated: "2026-05-17"
  tags: "sync, documentation, pull-request, quality, verification, pr, context-memory"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["sync", "docs", "pr", "documentation", "pull request", "changelog", "readme"]
  agents: ["manager-docs", "manager-git"]
  phases: ["sync"]
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->
<!-- Emits one line per Phase entry/exit to stderr in format: [trace] /moai sync Phase <N> <enter|exit> -->

# Sync Workflow Entry Router

이 파일은 `/moai sync` 진입점 라우터입니다. 각 Phase 그룹은 on-demand로 해당 sub-skill을 `Read`하여 로드합니다.

## Phase Owners (per the canonical agent catalog policy)

Skill injection: at each `manager-docs` spawn the orchestrator injects `At start, invoke Skill("moai-workflow-project") for the sync-phase documentation cycle.` (per `.claude/rules/moai/workflow/skill-routing.md` §1 and the delegation map `.moai/config/sections/delegation.yaml`).

Phase Owners: `manager-docs` (sync-phase artifact authoring — CHANGELOG.md + README.md + docs-site + progress.md §E.4 (lettered per `spec-frontmatter-schema.md` § progress.md Section Map) + the frontmatter status transition that rides the sync commit for all SPEC artifacts, states per `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix; MUST NOT modify spec.md/plan.md/acceptance.md body content per the same matrix) + `manager-git` (PR creation per branching strategy when Tier L OR `--pr` flag per the canonical Tier-based PR routing policy).

Sync-phase quality gate (per the canonical sync-phase quality gate policy) is enforced by the `.claude/hooks/moai/sync-phase-quality-gate.sh` Stop hook — lint + test + coverage delta verification + dependency manifest audit. The hook exits 0 always; in blocking mode (MOAI_SYNC_GATE_BLOCKING=1) it emits stdout JSON {"decision":"block"} on lint/test failure or coverage regression > 5pp. Per Claude Code hook semantics, stdout JSON is honored only on exit 0. The hook replaces the prior pattern of spawning an inline quality agent for coverage and security analysis during sync (that agent is archived per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 2; the Stop hook is its canonical replacement).

## Phase Routing Table

| Phase Group | Sub-skill 경로 | 내용 |
|------------|----------------|------|
| Phase 1: Pre-Sync Context + Deployment Readiness | `Read workflows/sync/quality-gates-context.md` | Purpose/Scope/Input/Mode/Flags/Context Loading, Phase 1 HUMAN GATE 1, Phase 2 DB Check, Phase 3 Deployment Readiness |
| Phase 7~Phase 10: Quality Verification | `Read workflows/sync/quality-gates-quality.md` | Phase 7 Quality Check, Phase 8 Security Scan, Phase 9 MX Tag Validation, Phase 10 Coverage Analysis |
| Phase 11~Phase 12: Analysis + Doc Sync | `Read workflows/sync/doc-execution.md` | Phase 11 Analysis, HUMAN GATE 2 Documentation Scope, Phase 12 Execute Doc Synchronization |
| Phase 13~Phase 14: Git Delivery + Completion | `Read workflows/sync/delivery.md` | Phase 13 Git Operations, Phase 14 Completion, GitStrategy PR-ready transition, Graceful Exit, Test Scenarios |

## Fan-Out Index

Every sync-phase fan-out site, listed here rather than only at the site itself. Three of the four live in sub-skills that are `Read` on demand, so without this index the orchestrator cannot know they exist until it has already entered the phase serially.

| Fan-Out ID | Trigger condition | Target file | What is parallelised |
|---|---|---|---|
| `FO-SYNC-1` | the quality-evidence fan-out script is on disk AND the runtime supports dynamic workflows | `workflows/sync.md` (below) | Phase 7 quality check — four quality dimensions in one parallel read-only pass |
| `FO-SYNC-2` | the modified files span several languages or packages | `workflows/sync/quality-gates-quality.md` | Phase 9 MX tag scan — one read-only shard per language or package |
| `FO-SYNC-3` | the coverage gaps span several independent packages | `workflows/sync/quality-gates-quality.md` | Phase 10 test drafting — one read-only drafter per package |
| `FO-SYNC-4` | the sync scope spans several independent document families | `workflows/sync/doc-execution.md` | Phase 12 document drafting — five read-only drafters, one applier |

## Docs ∥ Audit Concurrent Scheduling

The docs drafter fan-out (`FO-SYNC-4`, five read-only drafters D1-D5) launches CONCURRENTLY with the Phase 7-10 audit fan-out — in the SAME turn Phase 7 is entered, NOT serially after the audit completes. Reading the Phase Routing Table strictly top-to-bottom (Phase 7 → Phase 11 → Phase 12) would serialize the docs draft behind the full quality / security / MX / coverage pipeline. The orchestrator instead spawns the `FO-SYNC-4` drafter fan-out in the same single-turn multi-`Agent()` batch that enters Phase 7, so the docs draft is ready by the time the audit returns.

**Input independence.** Each docs drafter (D1-D5) reads its input from SPEC artifacts + git diff + the Phase 11 Step 1.5 divergence report — NOT from the concurrent audit's quality report, verdict, or per-dimension scores. The docs draft and the audit are input-independent; a drafter that read "the audit's functionality score" to decide CHANGELOG tone would create a hidden serial dependency that defeats the concurrency.

**Single-writer applier at gate-sync-2.** The concurrency is bought entirely by making the D1-D5 drafters read-only. `manager-docs` remains the sole write-capable agent: it applies the five drafts sequentially AFTER both fan-outs return, at the existing `gate-sync-2` HUMAN GATE 2 (Documentation Scope). The `[HARD]` concurrency guard (`agent-common-protocol.md` § Background Agent Execution — no two write-capable agents run concurrently) holds throughout; at no point during the concurrent fan-out do two write-capable agents run simultaneously. The audit verdict is surfaced to the user at the same gate-sync-2 round, with no extra human round-trip introduced.

**Concurrency guard codification.** The docs-drafter fan-out, the MX shard fan-out, the FO-SYNC-1 4-dim judges, and any sync-auditor fallback are ALL read-only. The single write-capable pass (`manager-docs` applying the docs drafts at gate-sync-2) runs AFTER both fan-outs return. This is the concurrency invariant: every concurrent agent is read-only (drafters return draft text; shards return findings; judges return structured verdicts; auditors read tree state), and the sole writer is sequential. The `[HARD]` concurrency guard (`agent-common-protocol.md` § Background Agent Execution) holds throughout every fan-out path — verified by inspection of every concurrent agent's read-only contract.

## Parallel Quality-Evidence Fan-Out (capability-gated)

**`FO-SYNC-1`.** **Where** `.claude/workflows/sync-audit-4dim.js` exists on disk **AND** the runtime supports dynamic workflows, the orchestrator shall launch it at Phase 7 (Quality Check) to gather the four quality dimensions in one parallel read-only pass. **Where** either condition is absent — the script was removed, or the runtime predates dynamic-workflow support — Phase 7 proceeds on its existing path with no error, no warning, and no interruption.

The orchestrator launches the script itself; this is scaling, not subagent nesting, so the flat agent hierarchy is preserved. Every judge is read-only and reports an `evidence_gaps` entry or a structured blocker report rather than prompting the user. On a clean PASS the verdict's binding additionally depends on the cross-model gate described under Binding promotion below.

**Binding promotion.** On the **happy path** — the workflow verdict is `PASS`, no dimension scored 0, the verdict is not `INCOMPLETE`, and no contested finding is present — the orchestrator SHALL treat the workflow's harmonic-mean verdict as **BINDING** for the sync-phase quality decision and SHALL NOT spawn the cold `sync-auditor` subagent. The four parallel xhigh judges subsume the one serial judge on the clean path (attributable diff-check, not a deletion of the auditor role). The mechanical binding predicate is codified in `internal/runtime.FourDimVerdict.IsBinding()`; the orchestrator constructs a `FourDimVerdict` from the workflow run output and consults `IsBinding()`. On any of the fallback triggers — (a) verdict `INCOMPLETE`, (b) any must-pass dimension scoring 0 (`zero_scored` array non-empty), or (c) a **contested finding** (any one judge reports `critical` severity, OR two or more judges return conflicting severity classifications for the same dimension) — the orchestrator SHALL spawn the cold `sync-auditor` subagent as the fallback binding-verdict owner, and the auditor's PASS/FAIL is treated as binding for that cycle. Neither `gate-sync-1` nor `gate-sync-2` is bypassed or auto-passed by either path; the cold auditor remains the fallback verdict owner under trigger (a)/(b)/(c).

**Cross-model gate on a clean PASS.** The four-dimension verdict is the Claude input. Where the tree requires a non-Claude audit backend, it is binding only together with a passing `audit_multi`; the orchestrator proceeds as follows after the script returns.

1. On a clean PASS, learn the orchestrator's own toplevel (`git rev-parse --show-toplevel`; `CLAUDE_PROJECT_DIR` names the primary checkout in a worktree session) and run `moai verify audit-plan --project-root <own toplevel>`. A plan (`config_status` `ok` or `absent`) with `cross_model_required: false` leaves the binding rule above unchanged and calls nothing. Output that is the `verify` group's help text (it contains `Shared diagnostic snapshot contract` and neither `config_status` nor an `audit-plan:` line) is the legacy path: the binding rule above applies unchanged, a cold `sync-auditor` spawned on a fallback trigger calls `audit_multi` without a `gates` argument under its own legacy path, and "plan surface unreachable, legacy path used" is recorded as a named Gap. `config_status: unreadable`, an `audit-plan:` error, or any other failure to run the verb (refused, crashed, timed out, malformed output) is not the legacy path: it is a PASS-blocking Gap, the verdict is not binding, and a configured required backend that does not answer stays fail-closed by name.
2. With `cross_model_required: true`, call `audit_multi` without a `gates` argument, with `project_root` set to the orchestrator's toplevel, the same `target` the cold auditor would use for this sync, and as `claude_verdict` an object synthesized from the workflow output: verdict `pass`, a one-line summary naming the harmonic mean, threshold and tier, and the workflow's findings — never judge reasoning text. The call is sequential, after the workflow returns; it writes only audit receipts and convergence state, never audited source or docs, so the single-writer invariant above holds. In a GPT- or GLM-launched session the supplied verdict is ignored and the independent Claude backend runs.
3. Write the digest of the result (`overall_verdict`, `gate_unmet`, `plan_source` and, per `per_backend_verdicts` entry, `backend`, `gate`, `verdict` — never summary or finding text) with the orchestrator's own `Write` tool to `<toplevel>/.moai/state/audit-plan-result.json`, fresh and immediately before the check: overwrite any earlier file at that name and never reuse one from an earlier audit. Then run `moai verify audit-plan --result-file <that path> --project-root <toplevel>`, passing only the path — never the JSON on the command line (the worktree guard refuses braces and quotes).
4. The sync verdict is binding only when both must pass: the four-dimension PASS, the `audit_multi` verdict `pass` with an empty `gate_unmet`, and `convergence_check.ok: true`. Otherwise it is not binding, for a named reason: the convergence verdict is `fail` (a reviewed defect; its findings go to the report); `gate_unmet` names the missing backend(s); the check names an unmet backend (including a result from a server that predates the plan); or `audit_multi` cannot be called while the plan is reachable and requires it (`audit_multi unreachable`, a Gap). The four-dimension verdict is never used alone in place of the missing half, and the cold `sync-auditor` is not spawned for these reasons — it would repeat the same call.
5. On a fallback trigger (a)/(b)/(c) the cold `sync-auditor` runs the verb and `audit_multi` itself and returns the result members in its report; its verdict is not final until the orchestrator's check passes. The orchestrator writes the digest file and runs the check as in step 3, and makes no other extra call on this branch.

The orchestrator's statement of which verdict is binding carries `cross_model: required (<backends>)`, `audit_multi: pass|fail|unavailable`, `gate_unmet: <backends|none>`, `audit_receipt: <ids|none>` and `plan_check: ok|unmet (<backends>)`. When the verdict is not binding it reads `binding: no — cross-model gate unmet: <backend>` (or `audit_multi unreachable`), never a bare four-dimension PASS.

## HUMAN GATE Map

| GATE | Location | Formal ID | Trigger |
|------|----------|-----------|---------|
| GATE 1: Pre-Sync Quality | `sync/quality-gates-context.md` | `gate-sync-1` | Working tree + all tests pass check before entering Phase 3 |
| GATE 2: Documentation Scope | `sync/doc-execution.md` | `gate-sync-2` | User reviews divergence report and approves doc regeneration scope |

> Note: Additional AskUserQuestion decision points exist in Phase 1 (gate failure), Phase 3 (test failure), Phase 6 (breaking changes), Phase 7 (test failure), Phase 8 (security critical), Phase 13 (CI mirror failure), and Phase 14 (next steps). These are inline decision gates, not named evolvable GATEs.

<!-- moai:contract-mode-start id="contract-sync-gates" -->
Where `workflow.autonomy.mode: contract` — the sync phase carries the last three lifecycle stages: Closure (status transition and verdict), Integration (re-measure on the merged tree), Push (inactive until the stop-before-push on a missing second review exists). No documentation-scope approval (`gate-sync-2`), next-step, or current-branch question is asked, and the failure decision points — Phase 1 gate failure, Phase 3 and Phase 7 test failure, Phase 6 compatibility break, Phase 8 critical security finding, Phase 13 local CI mirror failure — are routed to an escalation report instead of a question. The CI auto-fix loop after a pull request is unchanged. See `.claude/rules/moai/workflow/contract-autonomy.md` § Gate disposition.

<!-- moai:contract-mode-end -->
## Invocation Flow

```
/moai sync [mode] [--pr] [--auto-merge] [--skip-mx]
  ├── [trace] /moai sync Phase 1 enter
  │   Read workflows/sync/quality-gates-context.md  → HUMAN GATE 1 + Deployment Readiness
  ├── [trace] /moai sync Phase 7 enter
  │   Read workflows/sync/quality-gates-quality.md  → Quality/Security/MX/Coverage
  ├── [trace] /moai sync Phase 11 enter
  │   Read workflows/sync/doc-execution.md          → Divergence Analysis + HUMAN GATE 2 + Doc Sync
  └── [trace] /moai sync Phase 13 enter
      Read workflows/sync/delivery.md               → Git Ops + CI Mirror + PR + Auto-Merge + Completion
```

## Quick Reference

**Purpose**: 코드 변경사항과 문서를 동기화하고, PR을 생성하여 SPEC 사이클을 완료합니다.

**Input**: `$ARGUMENTS` = `[mode] [path] [flags]`

**Modes**: `auto` (기본) | `force` | `status` | `project`

**Flags**: `--pr` (PR 생성) | `--auto-merge` (auto-merge 옵트인) | `--merge` (deprecated alias of `--auto-merge`) | `--skip-mx` (MX 검증 스킵)

**HUMAN GATEs**: GATE 1 (quality-gates-context.md §Phase 1) → GATE 2 (doc-execution.md §Step 1.6)

**Status mode early exit**: Phase 9 완료 후 보고서 출력 및 종료 (Phase 11+ 실행 안 함)

**Chain entry**: sync may be entered via auto-chain from run-phase completion (a `full-pipeline` contract — announced in the transcript, no extra approval round at the run→sync phase boundary) or via explicit `/moai sync` invocation (a `single-phase` contract). Either way the sync-internal HUMAN GATEs (`gate-sync-1`, `gate-sync-2`) fire unchanged, and a FAIL/INCONCLUSIVE sync-audit verdict or a blocking sync quality gate HALTS the chain — no auto-completion past a failing gate.

## On-Demand Sub-skill Loading

각 Phase 진입 시점에 해당 sub-skill을 로드합니다:

```
# Phase 1 (Pre-Sync Quality Gate) 진입 시
Read ~/.claude/skills/moai/workflows/sync/quality-gates-context.md

# Phase 7 (Quality Verification) 진입 시
Read ~/.claude/skills/moai/workflows/sync/quality-gates-quality.md

# Phase 11 (Analysis and Planning) 진입 시
Read ~/.claude/skills/moai/workflows/sync/doc-execution.md

# Phase 13 (Git Operations) 진입 시
Read ~/.claude/skills/moai/workflows/sync/delivery.md
```

## Custom Harness Extension

@.moai/harness/sync-extension.md

*(이 파일은 `/moai project --harness`로 생성됩니다. 파일이 없으면 자동으로 skip됩니다.)*

## Sentinel Error Keys

A CI audit verifies the literal `MODE_PIPELINE_ONLY_UTILITY` sentinel remains present in this skill body (shared with `design.md`). Passing `--mode pipeline` to `/moai sync` is rejected because sync is a Multi-Agent subcommand; pipeline mode is reserved for utility subcommands.

## Routing Ledger Recording

At sync dispatch, the orchestrator records the routing decision to the routing-ledger via `moai harness ledger record` (per the SKILL.md router recording obligation). At the sync-phase quality gate, it appends the gate result as machine evidence via `moai harness ledger evidence --kind gate_exit --value <exit> --terminal --ref "sync-phase-quality-gate"`. Outcome is derived from machine evidence only — never supplied as an input. The recording is opt-in and fail-open; it never blocks the sync phase.

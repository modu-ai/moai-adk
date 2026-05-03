# SPEC-V3R3-RETIRED-AGENT-001 Progress

- plan_complete_at: 2026-05-04T07:00:00Z
- plan_status: audit-ready

## Artifacts

- spec.md — v0.1.0 (frontmatter v0.2.0 정합화 완료, 9 required fields; 16 REQs / 18 ACs; no BC, `breaking: false`)
- research.md — Phase 0.5 deep research (5-layer defect chain decomposition with mo.ai.kr 21:14:54 timeline; 30+ file:line citations; codebase grounded scan of `internal/template/templates/.claude/agents/moai/` + `internal/hook/agents/` + `internal/template/templates/.claude/hooks/moai/handle-subagent-start.sh.tmpl` + `internal/cli/launcher.go` + `mo.ai.kr` comparative file size evidence)
- plan.md — Phase 1B implementation plan (5 milestones M1-M5; 25+ file:line anchors; mx_plan with 10 tags / 7 files; REQ↔AC traceability matrix 16→18)
- acceptance.md — Given/When/Then for 18 ACs (happy path + edge cases per AC; integration test scaffold names declared for `TestAgentFrontmatterAudit`, `TestManagerCyclePresent`, `TestAgentStartBlocksRetiredAgent`, `TestValidateWorktreeReturnRejectsEmptyObject`, etc.)
- progress.md — this file (Phase 1B status + plan-audit-ready summary)

Optional artifacts (not produced in this plan-stage; deferred to /moai run if needed):
- tasks.md — task-level decomposition T-RA-01..T-RA-NN with TDD owner roles
- spec-compact.md — auto-extracted REQ + AC + Files-to-modify + Exclusions reference for run-phase token saving

## Branch

- Branch: feature/SPEC-V3R3-RETIRED-AGENT-001
- Mode: solo, no worktree (per CLAUDE.local.md §15 + user directive)
- Working directory: /Users/goos/MoAI/moai-adk-go
- Base: origin/main HEAD 145eb59a9 (SPEC-CI-MULTI-LLM-001 source_report 보강 #771)
- Parent commits visible:
  - 145eb59a9 docs(spec): SPEC-CI-MULTI-LLM-001 source_report 보강 (#771)
  - 5270d7f82 docs(spec): SPEC-V3R3-HYBRID-001 plan 4종 산출물 (#770)
  - cbc46c9b4 docs(spec): SPEC-GLM-MCP-001 plan (#769)

## Key Plan Decisions

- **5-layer defect chain root cause**: identified via mo.ai.kr 2026-05-04 21:14:54 timeline (sourced verbatim from user prompt). Layer 1 (retired stub frontmatter invalid) is the primary fix surface; layers 2-4 (worktree allocation, fallback propagation, path interpolation) are in-depth defense; layer 5 (stream_idle_partial) explicitly out of scope per spec.md §1.3.
- **manager-cycle.md is absent in moai-adk-go template** (verified via `ls`). mo.ai.kr deployment has 10245-byte version dated 2026-05-01 13:51 — same date/time as retired stubs. SPEC-V3R2-ORC-001 retirement decision was applied incompletely (retired stubs deployed without prerequisite manager-cycle.md being part of moai-adk-go ship).
- **manager-tdd retired stub**: standardize frontmatter with `retired: true`, `retired_replacement: manager-cycle`, `retired_param_hint: "cycle_type=tdd"`, `tools: []`, `skills: []`. Remove legacy `status: retired` custom field. Body content retains migration notes pattern from mo.ai.kr stub.
- **manager-ddd retired stub**: explicitly OUT OF SCOPE per spec.md §1.3 — same pattern but separate SPEC `SPEC-V3R3-RETIRED-DDD-001` (가칭) for future. Audit test (`agent_frontmatter_audit_test.go`) scoped to manager-tdd only to avoid drive-by drift.
- **SubagentStart hook guard**: new `internal/hook/agent_start.go` handler dispatched via `internal/hook/agents/factory.go` extension. Reads agent file frontmatter, detects `retired: true`, returns block decision with exit 2 + JSON. Wrapper script `handle-subagent-start.sh.tmpl` updated to propagate exit code (replace `exec` form with `exit $?`).
- **worktreePath validation guard**: new `validateWorktreeReturn()` helper in `internal/cli/launcher.go` (or new `internal/cli/agent_wrapper.go`). Rejects empty-object / null / non-string worktreePath with sentinel `WORKTREE_PATH_INVALID`. Skips validation when isolation mode is not `worktree` (no false-positive on no-worktree calls).
- **Path interpolation refactor**: 3-5 callsites (estimated; verified at M3 implementation) migrate from `fmt.Sprintf("...%s/%s/%s...", ...)` to Go `text/template` with typed data struct. Type-safe templates produce errors for non-string values instead of silently substituting `[object Object]` or `{}`.
- **TDD methodology**: M1 RED (4 test files: agent_frontmatter_audit + manager_cycle_present + agent_start + launcher_worktree_validation). M2-M4 GREEN (manager-cycle add + retired stub standardize + hook handler + worktree validation + path refactor). M5 REFACTOR (7 documentation substitutions + lessons.md #11 + final make build + full go test ./...).
- **Documentation substitution scope**: 7 references across 6 files (CLAUDE.md §4, §5, agent-authoring.md, agent-hooks.md, spec-workflow.md, manager-strategy.md, manager-ddd.md inline). Manager Agents count documentation: "8" remains (active manager-cycle replaces retired manager-tdd, net 8 effective).
- **No BC**: backward-compatible fix. Existing `manager-tdd` callers receive retirement message + migration hint (same as mo.ai.kr current behavior); standardized frontmatter improves hook detection but does not change observable behavior for end-user. `breaking: false`, `bc_id: []`.
- **mo.ai.kr propagation**: user-side fix is `moai update` (template sync after this SPEC merges). User data (`.moai/specs/`, `.moai/project/`) preserved; `.claude/agents/` synced.

## Frontmatter Migration Verification (spec.md v0.1.0)

- Required fields present (9/9): `id`, `version`, `status`, `created_at`, `updated_at`, `author`, `priority`, `labels`, `issue_number` ✅
- Rejected aliases absent (0): `created`, `updated`, `spec_id`, `title:` H1-alias ✅
- `version` quoted string: `"0.1.0"` ✅
- `priority` enum: `P0` (bare uppercase, no descriptor) ✅
- `labels` YAML array: `[agent-runtime, templates, retired-stub, manager-cycle, manager-tdd, hooks, bug-fix, v3r3]` ✅
- `created_at` / `updated_at` ISO date: `2026-05-04` / `2026-05-04` ✅
- `issue_number: null` ✅
- Optional BC fields: `breaking: false` + `bc_id: []` + `lifecycle: spec-anchored` ✅
- Optional related_specs: `[SPEC-V3R3-HYBRID-001]` ✅
- Optional dependencies: `[SPEC-V3R2-ORC-001]` ✅

## Codebase Scan Summary (research.md grounded)

### `manager-cycle.md` absence (P0 #1 root cause)

```bash
$ ls /Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/agents/moai/manager-cycle.md
ls: ... No such file or directory

$ ls -la /Users/goos/MoAI/mo.ai.kr/.claude/agents/moai/manager-cycle.md
-rw-r--r-- 1 goos staff 10245 May 1 13:51 ...
```

### Retired stub size deltas (P0 #2 evidence)

| Location | Size | Date |
|---|---|---|
| `internal/template/templates/.claude/agents/moai/manager-tdd.md` (active) | 6407 bytes | 2026-04-30 |
| `mo.ai.kr/.claude/agents/moai/manager-tdd.md` (retired stub) | 976 bytes | 2026-05-01 |
| `internal/template/templates/.claude/agents/moai/manager-ddd.md` (active, OUT OF SCOPE) | 7628 bytes | 2026-04-30 |
| `mo.ai.kr/.claude/agents/moai/manager-ddd.md` (retired stub, OUT OF SCOPE) | 1000 bytes | 2026-05-01 |

### SubagentStart hook current state (P0 #3 modification target)

- `internal/template/templates/.claude/hooks/moai/handle-subagent-start.sh.tmpl` (1050 bytes) — currently no-op pass-through to `moai hook subagent-start`
- `internal/hook/agents/` — 11 handler files exist but **none registered for SubagentStart event**
- New file: `internal/hook/agent_start.go` — dispatched via `internal/hook/agents/factory.go` switch case extension

### Documentation substitution targets (REQ-RA-013 scope, M5)

7 references across 6 files (research.md §3.6):
1. `manager-strategy.md` line: `manager-ddd or manager-tdd` → `manager-cycle`
2. `manager-ddd.md` 2 inline references → `manager-cycle with cycle_type=tdd`
3. `CLAUDE.md §4 Manager Agents (8)` — active list update
4. `CLAUDE.md §5 Agent Chain` — Phase 3 reference + MoAI Command Flow
5. `agent-authoring.md` Manager Agents listing
6. `agent-hooks.md` Agent Hook Actions table — manager-tdd row
7. `spec-workflow.md` Phase Overview table

## Next Phase

- Phase 0.5 Plan Audit Gate (plan-auditor) at `/moai run SPEC-V3R3-RETIRED-AGENT-001` entry — see `.claude/rules/moai/workflow/spec-workflow.md:172-204`.
- Implementation Methodology: TDD (per `.moai/config/sections/quality.yaml`).
- Run-phase command: `/moai run SPEC-V3R3-RETIRED-AGENT-001` (executed from `/Users/goos/MoAI/moai-adk-go` on branch `feature/SPEC-V3R3-RETIRED-AGENT-001`).
- Post-implementation: `/moai sync SPEC-V3R3-RETIRED-AGENT-001` for documentation sync (docs-site 4-locale per CLAUDE.local.md §17 if user-facing) + PR creation.

## Plan-Audit-Ready Checklist Summary

All 18 criteria PASS per plan.md §8:

- C1: Frontmatter v0.2.0 (9 required fields) ✅
- C2: HISTORY v0.1.0 entry ✅
- C3: 16 EARS REQs across 5 categories (Ubiquitous 6, Event-Driven 4, State-Driven 3, Optional 1, Unwanted 2) ✅
- C4: 18 ACs with 100% REQ mapping (16/16 REQ → AC traceability matrix in plan.md §1.4) ✅
- C5: BC scope clarity (no BC; `breaking: false`, `bc_id: []`) ✅
- C6: File:line anchors ≥10 (research.md: 30+, plan.md: 25+) ✅
- C7: Exclusions section present (spec.md §1.3 Non-Goals + §2.2 Out of Scope, 11 items) ✅
- C8: TDD methodology declared ✅
- C9: mx_plan section (10 tags / 7 files; 3 ANCHOR + 3 NOTE + 2 WARN + 1 TODO + 1 LEGACY) ✅
- C10: Risk table with file-anchored mitigations (spec.md §8: 13 risks; plan.md §5: 14 risks) ✅
- C11: Solo mode path discipline (4 HARD rules, no worktree per user directive) ✅
- C12: No implementation code in plan documents ✅
- C13: Acceptance.md G/W/T format with edge cases (18 ACs covered) ✅
- C14: Owner roles aligned with TDD methodology (M1-M5 declares expert-backend / manager-cycle owner roles) ✅
- C15: Cross-SPEC consistency (SPEC-V3R2-ORC-001 dependency declared as completed; SPEC-V3R3-HYBRID-001 PR #770 merged related; SPEC-V3R2-WF-005 PR #768 merged 16-language neutrality pattern applied) ✅
- C16: BC migration completeness (spec.md §10: no BC, backward-compatible fix per Bug Fixes section in CHANGELOG) ✅
- C17: 5-layer defect chain documented (research.md §2 layer-by-layer + spec.md §1.1 background) ✅
- C18: External evidence verified (mo.ai.kr file size diff via `ls -la`, manager-cycle absence via `ls`, SubagentStart hook spec via hooks-system.md) ✅

## Open Items for plan-auditor Review

- Confirm SubagentStart hook actually blocks spawn on exit code 2 + JSON `{"decision":"block"}` — `hooks-system.md` table says "Can Block: No" for SubagentStart but exit codes 0/1/2 documentation suggests blocking semantic. Empirical test at M3 implementation phase will resolve. Fallback: PreToolUse hook on Agent tool with matcher.
- Confirm `retired: true` custom frontmatter field is silently ignored by Claude Code agent runtime (not raised as YAML schema error). M2 verify with single test agent spawn. Fallback: encode retirement metadata in `description:` field.
- Validate `text/template` migration scope is ≤5 callsites. M3 grep measurement at implementation; if >5, scope-cut to validation-only or escalate to user.
- Confirm adding `manager-cycle.md` to template does NOT change Manager Agents count documentation (8 active = manager-cycle replaces retired manager-tdd, so net 8 unchanged).
- Verify mo.ai.kr's `manager-cycle.md` (10245 bytes) passes 16-language neutrality + anti-bias check before importing as template baseline. M2 manual review before commit.
- Confirm `worktreePath` empty-object validation does NOT false-positive on legitimate "no worktree" cases (when `isolation` is not `"worktree"`). Test at M1: `TestValidateWorktreeReturnSkipsWhenIsolationNotWorktree`.
- Decide whether `moai agents list --retired` (REQ-RA-014) is in scope for v3R3 first minor release or deferred to follow-up SPEC. AskUserQuestion at M5 decision point.
- Confirm scope discipline: `manager-ddd` retired stub (mo.ai.kr 1000 bytes evidence) is OUT OF SCOPE per spec.md §1.3. Audit test scoped to manager-tdd only. If audit test discovers manager-ddd retired stub at execution time, treat as observation only, not failure.

---

End of progress.md.

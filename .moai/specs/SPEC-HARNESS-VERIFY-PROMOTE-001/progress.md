# Progress — SPEC-HARNESS-VERIFY-PROMOTE-001

> Canonical §E lifecycle section skeleton. §E.1 (plan-phase) populated by
> manager-spec; §E.2 / §E.3 owned by manager-develop (run-phase); §E.4 owned by
> manager-docs (sync-phase). The §F Phase 0.95 Mode Selection section is written by
> the orchestrator before the first run-phase Agent() spawn (not by manager-spec).
> The literal `§E.2` / `§E.3` / `§E.4` headings are parser-load-bearing (era.go
> era classification) — do NOT rename.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-11
plan_audit_fix_at: 2026-07-11 (D1-D6; version 0.1.1)
tier: S
artifact_set: spec.md + plan.md + acceptance.md + progress.md
req_count: 9
ac_count: 12
depends_on: SPEC-PROJECT-HARNESS-BRIDGE-001 (must be status: completed before run-phase entry — Phase 0.5 Depends_on Pre-flight gate)
open_clarifications: 0 (both resolved — see below)

### Resolved clarifications (plan-audit D1)

1. **Promoted-offer placement seam — RESOLVED.** The harness-generation offer is
   placed in `project/meta-harness.md` as a post-project-type-confirmation harness
   proposal — meta-harness.md is the Phase 5.1 handoff module, NOT the interview's
   final question. The offer fires after the adaptive interview confirms the project
   type. Scope is NOT extended into `project/mode-detection.md` /
   `project/codebase-analysis.md` (SPEC-PROJECT-HARNESS-BRIDGE-001 scope). The
   Phase 4.2 "Generate harness" menu — hosted in `project/doc-generation.md` (read-only
   for THIS SPEC) — is RETAINED as a fallback (both entry points reachable).
2. **Verify-skill enforcement surface — RESOLVED.** The `harness-<name>-verify` skill
   is ALWAYS mandatory for every generated harness (not gated on any `harness-spec.yaml`
   field). When no build / launch / test recipe is discoverable, a documented STUB
   verify skill ("no recipe found") is generated rather than skipping the skill.

> Placement note: per the manager-spec ownership matrix, this agent populates ONLY
> §E.1 (plan-phase) and leaves §E.2/§E.3 (run-phase, manager-develop) and §E.4
> (sync-phase, manager-docs) as empty placeholders. The two resolution records above
> are plan-phase decisions, so they live here in §E.1.

## §E.2 Run-phase Evidence

12/12 AC PASS (0 FAIL). Every grep/diff run against the LOCAL worktree tree; verbatim outputs captured this run.

| AC | Status | Verification Command | Actual Output |
|----|--------|----------------------|---------------|
| AC-HVP-001 | PASS | `grep -c -i "생성할까요\|generate.*harness.*for this project\|post-project-type-confirmation\|promoted offer" meta-harness.md` | `5` (≥1) |
| AC-HVP-002 | PASS | `grep -c -i "post-project-type-confirmation\|하네스를 생성할까요" meta-harness.md` | `4` (≥1 NEW insert) |
| AC-HVP-003 | PASS | baseline `git show HEAD:harness-build-entry.md \| grep -c` = `0`; post-insert `grep -c -i "final.round\|final question\|생성할까요\|harness.*proposal"` | baseline `0` → `4` (≥1) |
| AC-HVP-004 | PASS | `grep -c "harness-.*-verify\|harness-<name>-verify"` ; `grep -c -i "run-skill-generator\|build.*launch.*test\|verification loop\|runnable check"` (harness-builder.md) | `4` / `4` (≥1 each) |
| AC-HVP-005 | PASS | `grep -c -i "tool priority\|category fit\|category-fit\|style preference" harness-builder.md` | `3` (≥1) |
| AC-HVP-006 | PASS | `grep -c -i "skill-first\|read the relevant.*SKILL.md\|before.*file.*code work" harness-builder.md` | `3` (≥1) |
| AC-HVP-007 | PASS | `grep -c -i "3-7\|3 to 7\|3–7\|recurring same-instruction\|emit.*skill instead\|over-generation" harness-builder.md` (in PLAN section) | `1` (≥1) |
| AC-HVP-008 | PASS | `grep -c -i "harness-\*\|harness- namespace\|harness-<name>"` ; `grep -c "\.claude/agents/moai/\|\.claude/skills/moai-\|\.claude/rules/moai/"` (harness-builder.md) | `12` / `18` (≥1 each — namespace + FROZEN reject-paths intact) |
| AC-HVP-009 | PASS | `grep -rn "\.moai/specs/" meta-harness.md harness-build-entry.md \| wc -l` | `0` |
| AC-HVP-010 | PASS | `diff -q local template-mirror` for all 3 touched files | `PARITY OK` × 3 |
| AC-HVP-011 | PASS | `make build` → exit 0 ; `go test ./internal/template/...` → `ok internal/template 1.318s` ; `grep -rn "SPEC-HARNESS-VERIFY-PROMOTE-001" internal/template/templates/ \| wc -l` → `0` | exit 0 / ok / 0 |
| AC-HVP-012 | PASS | `sed -n '/^## Tool Priority/,/^$/p' \| grep -c '.'` ; `sed -n '/^## Skill-First Execution/,/^$/p' \| grep -c '.'` | `5` / `3` (both 1..8 — bounded) |

Preservation / non-regression:

| Invariant | Status | Evidence |
|-----------|--------|----------|
| `harness-*` namespace-only intact | PASS | AC-HVP-008 first grep = 12 |
| FROZEN reject-path guard intact | PASS | AC-HVP-008 second grep = 18 |
| NO-SPEC scope guard (no `.moai/specs/` write path) | PASS | AC-HVP-009 = 0 |
| Phase 4.2 fallback (doc-generation.md) retained | PASS | doc-generation.md NOT in touched set (`git status --short`) — retained by non-modification |
| Existing 5 base artifact types preserved | PASS | Artifacts 1-5 unchanged; verify skill added as additive artifact 6 |
| Cross-platform + whole-repo build | PASS | `go build ./...` exit 0 ; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 |
| Whole-repo test non-regression (doc-only) | PASS | `go test ./...` exit 0 — 96 packages ok / 0 FAIL |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: PASS
run_complete_at: 2026-07-11
run_commit_sha: 75fa9217a   # M2 = final implementation commit; M3 (this) records run evidence
ac_pass_count: 12
ac_fail_count: 0
preserve_list_post_run_count: 5   # harness-* namespace; FROZEN reject-paths; NO-SPEC guard; Phase 4.2 fallback (doc-generation.md); 5 base artifact types
l44_pre_commit_fetch: "0 0 (synced with origin/main at run-phase start, HEAD 073e2fd51)"
l44_post_push_fetch: "worktree branch worktree-agent-a4863a2c5733ae8e5; L1 integration to main handled at completion"
new_warnings_or_lints_introduced: 0   # doc-only; go build + go test ./... clean (96 ok / 0 FAIL)
cross_platform_build:
  darwin_amd64: "exit 0"
  windows_amd64: "exit 0 (GOOS=windows GOARCH=amd64 go build ./...)"
total_run_phase_files: 6   # 3 workflow markdown files × 2 trees (local + template mirror); + spec.md frontmatter + progress.md evidence
m1_to_mN_commit_strategy: "3 milestones, Route A Hybrid Trunk direct-to-main. M1 2a042bef1 (promote offer + draft→in-progress), M2 75fa9217a (verify skill + specialist rule blocks + 3-7 guardrail), M3 (this — run evidence). L1 worktree; specific-path git add only."
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: PASS
sync_complete_at: 2026-07-12
sync_commit_sha: 09aec212ea7791a1401734f08d5e34954b593f61
changelog_entry_position: "[Unreleased] > ### Added (first entry, Epic 'Project-Harness Pipeline' SPEC 3/3)"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (this commit)"
  updated_field_refreshed: "2026-07-12"
template_parity_reverified: "PARITY OK x3 (project/meta-harness.md, harness-build-entry.md, harness-builder.md)"
neutrality_reverified: "grep -rn SPEC-HARNESS-VERIFY-PROMOTE-001 internal/template/templates/ -> 0"
epic_closure: "3-SPEC Project-Harness Pipeline Epic CLOSED (SPEC-PROJECT-HARNESS-BRIDGE-001 -> SPEC-HARNESS-MCP-PROVISION-001 -> SPEC-HARNESS-VERIFY-PROMOTE-001)"
```

## §F Phase 0.95 Mode Selection

Input parameters:
- tier: S (doc-only, 2-artifact LEAN — here 4 files incl. acceptance.md + progress.md)
- scope (file count): 3 touched files (meta-harness.md, harness-build-entry.md, harness-builder.md) + template mirrors
- domain count: 1 (workflow-skill markdown edits only)
- file language mix: 100% markdown (no Go / shell)
- concurrency benefit: LOW (sequential doc edits with a Template-First mirror step + make build between)

Mode evaluation:
| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | multi-file, multi-milestone semantic doc insertion — not a single-line change |
| 2 background | no | writes tracked files (background write restriction) |
| 3 agent-team | no | RETIRED |
| 4 parallel | no | single domain, not research-heavy; parallel writes to shared workflow tree risk conflict |
| 5 sub-agent | YES | coding/doc-heavy single-domain sequential work — default fallback fits |
| 6 workflow | no | <30 files, not a uniform mechanical transform |

Decision: sub-agent

Justification: Tier S doc-only single-domain work with a Template-First mirror + make build cadence between milestones is inherently sequential; Mode 5 (single manager-develop, sequential M1→M2→M3) is the correct default per Anthropic's coding-task parallelism caveat. No fan-out benefit.

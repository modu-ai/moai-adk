---
id: SPEC-SUBAGENT-NESTING-DOCTRINE-001
title: "Subagent-nesting doctrine correction + auditor read-only nesting pilot — Progress"
version: "0.1.0"
status: completed
created: 2026-07-24
updated: 2026-07-24
author: manager-spec
priority: P2
phase: "v3.0.2 target"
module: ".claude"
lifecycle: spec-anchored
tags: "doctrine, subagent-nesting, claude-code, agent-authoring, sync-auditor"
tier: M
---

# Progress — SPEC-SUBAGENT-NESTING-DOCTRINE-001

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID `SPEC-SUBAGENT-NESTING-DOCTRINE-001` self-check: `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` → PASS (executed Bash; verbatim `PASS`).
- Artifacts authored (4): spec.md + plan.md + acceptance.md + progress.md; directory-structured (no flat file).
- Frontmatter: all 12 canonical fields present across all 4 files; `created`/`updated` (not `_at`); `tags` comma-separated string (not `labels`); `status: draft`.
- Notation: GEARS (Ubiquitous / When / While / Where / Unwanted) throughout §B.
- Out of Scope: 6 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (spec.md §E), incl. `### Out of Scope — plan-auditor nesting pilot`.
- Scope: single SPEC, two milestones (M1 doc-correction + M2 opt-in env-gated pilot) — coupled, split not recommended.
- Ground truth: orchestrator-verified (spec.md §A); 7 M1 surfaces + 2 M2 surfaces confirmed to exist with template mirrors (2026-07-24).
- Clarifications RESOLVED (plan finalization): D5 → M2 pilot scope = `sync-auditor` only (`plan-auditor` deferred to a future SPEC, spec.md §E Out of Scope — plan-auditor nesting pilot); D6 → M1 + M2 both ship in v3.0.2 (shipped default flat, `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH` env opt-in only). Both clarification markers resolved; none remain (0 open).
- @MX targets: none (doctrine prose + agent frontmatter, no Go production code).

## §E.2 Run-phase Evidence

Run-phase = AC-driven doctrine + frontmatter edit. The acceptance.md grep/verification commands ARE the executable spec; each AC's exact command was run and its Actual Output cited. All 6 target files edited in BOTH the live tree and the `internal/template/templates/` mirror (Template-First). `make build` regenerated `catalog.yaml` (sync-auditor agent hash only — the A3c same-SPEC cascade).

| AC | Requirement | Status | Verification command | Actual Output |
|----|-------------|--------|----------------------|---------------|
| AC-SND-001 | REQ-SND-002 | PASS | Watch-note stale-gone + v2.1.217/double-guarantee/sync-auditor greps | stale count=0; DEPTH=2, v2.1.217=2, default-off=1, double/both=2, sync-auditor=5 |
| AC-SND-002 | REQ-SND-003 | PASS | §14-anchored concurrency-cap greps (live + template) | MAX_CONCURRENT=1, MAX_PER_SESSION=1, `20`=1, `200`=1, template=1 |
| AC-SND-003 | REQ-SND-004 | PASS | agent-authoring §Agent(agent_type) delta greps | stale=0; DEPTH=3; v2.1.217/default-off=3 |
| AC-SND-004 | REQ-SND-005 | PASS | Fork-Subagents unmarked-stale + supersed/configurable greps | unmarked=0; supersed/configurable=2 |
| AC-SND-005 | REQ-SND-006 | PASS | Tool-Permissions runtime-default-off grep | runtime-default-off=2 |
| AC-SND-006 | REQ-SND-007 | PASS | agent-patterns §Deprecated delta + gone-check (live + template) | v2.1.217=1, sync-auditor=4; unmarked-stale live=0, template=0 |
| AC-SND-007 | REQ-SND-008 | PASS | orchestration §Mode 6 retained + version-note greps | scaling-NOT-nesting=1 (retained); v2.1.217/default-off=1 |
| AC-SND-008 | REQ-SND-009 | PASS | zone-registry CONST-020/044 nesting-term greps (D4 no-op) | CONST-020=0, CONST-044=0 (concurrency caps authored as distinct §14 bullet, no re-sync) |
| AC-SND-009 | REQ-SND-010/022 | PASS | `make build` exit 0 + `go test ./internal/template/...` exit 0 | make build=0; go test template=0 (mirror-parity + neutrality green) |
| AC-SND-010 | REQ-SND-011 | PASS | SPEC-ID-in-template grep + TestTemplateNoInternalContentLeak / TestTemplateNeutralityAudit | SPEC-ID=0; both neutrality tests PASS |
| AC-SND-011 | REQ-SND-013/014/021 | PASS | held-in: Agent-in-tools + per-dimension + read-only-child greps | Agent-in-tools OK; per-dimension=12; read-only-mechanism=1 |
| AC-SND-012 | REQ-SND-015/019 | PASS | held-out: settings.json/tmpl env grep + other-10-agents Agent-in-tools loop | settings env=0; no UNEXPECTED Agent in the other 10 agents |
| AC-SND-013 | REQ-SND-018 | PASS | boundary-guard widened-filter grep (live + template) | 0 surviving matches in both trees |
| AC-SND-014 | REQ-SND-017 | PASS | read-only-mechanism grep + tool-grant-syntax Write/Edit grep | read-only=5; write-grant=0 |
| AC-SND-015 | REQ-SND-016 | PASS | verdict-ownership grep | verdict-ownership=1 |
| AC-SND-016 | REQ-SND-019 | PASS-WITH-DEBT | whole-tree DEPTH-env grep (AC literal) vs settings.json-scoped intent | whole-tree=8 (env-name in 5 doctrine mirrors, REQUIRED by AC-001/003 + AC-002 pattern, generic Claude Code env allowed per §25/Section B); settings.json.tmpl DEPTH-env=0 (TRUE intent satisfied — shipped default flat) |

Invariants:

| Invariant | Status | Evidence |
|-----------|--------|----------|
| `go build ./...` clean pre + post | PASS | exit 0 both |
| Mirror parity (6 files, live == template on edited spans) | PASS | identical old→new applied to both trees; `go test ./internal/template/...` exit 0 |
| Held-out: distributed settings has no depth env | PASS | `settings.json.tmpl` DEPTH-env count = 0 (`settings.json` is rendered from `.tmpl`, absent from templates/) |
| No Go test enforces whole-tree DEPTH-env=0 | PASS | `grep -rn CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH --include=*.go internal/ cmd/ pkg/` → 0 matches (CI does not break) |
| Other 10 retained agents carry no `Agent` in `tools` | PASS | AC-SND-012 loop → no UNEXPECTED lines |

> M2 tooling-enforcement completion (post-close cascade, status stays `completed`): the `moai agent lint` LR-02 rule (`internal/cli/agentlint/agent_lint.go` `checkAgentInTools` + `nestingPilotAllowlist`) now allowlists `sync-auditor` so the pilot frontmatter (`tools: … Agent …`) passes CI, while LR-02 STILL errors for every non-allowlisted agent (flat-hierarchy guard for the other 10). New tests: `TestCheckAgentInTools_NestingPilotAllowlist` + 2 table cases (allowlist-exempt + guard-intact). `bin/moai agent lint` → LR-02 count 0 on both the live `sync-auditor.md` and its template mirror.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-24
run_commit_sha: 190579229
run_status: complete
ac_pass_count: 15
ac_pass_with_debt_count: 1        # AC-SND-016 (whole-tree grep proxy vs settings.json-scoped intent)
ac_fail_count: 0
preserve_list_post_run_count: 0   # no PRESERVE-list violation; other sessions' files untouched
l44_pre_commit_fetch: not-run     # NO-PUSH shared branch (owned by parallel SEC-DEEPSCAN session); orchestrator cherry-picks later
l44_post_push_fetch: not-applicable  # no push per spawn constraint
new_warnings_or_lints_introduced: 0   # plan/acceptance MissingExclusions warnings are PRE-EXISTING grandfathered (not touched by run-phase)
cross_platform_build:
  go_build_all: exit-0
  make_build: exit-0
  goos_windows: not-run          # doctrine + frontmatter edit only, no OS-specific Go source touched
total_run_phase_files: 13        # 6 live + 6 template mirror + 1 catalog.yaml (A3c cascade)
m1_to_mN_commit_strategy: single-commit  # M1 doc-correction + M2 pilot in one pathspec-scoped commit; NO push
ac_debt_reconciliation: >
  AC-SND-016's literal whole-tree grep (CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH in
  internal/template/templates/ == 0) is UNSATISFIABLE alongside AC-SND-001 +
  AC-SND-003 (which require the same env-var token in the CLAUDE.md and
  agent-authoring.md template mirrors) and the AC-SND-002 pattern (which REQUIRES
  the same-class CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS in the template CLAUDE.md §14
  mirror). The env-var name is a GENERIC Claude Code identifier (allowed in
  templates per CLAUDE.local.md §25 + spawn Section B) documented ONLY in doctrine
  prose, never in a settings config that would enable nesting. REQ-SND-019's actual
  intent (env absent from the distributed settings.json so the shipped default stays
  flat) is FULLY satisfied: settings.json.tmpl DEPTH-env count = 0. The whole-tree
  grep is an over-broad proxy; the settings.json-scoped invariant (AC-SND-012 first
  grep) is the load-bearing one.
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-24
sync_commit_sha: 8e2880378
sync_status: complete
changelog_entry_position: "[Unreleased] > ### Changed (first entry)"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"
  plan_md: "in-progress -> completed"
  acceptance_md: "in-progress -> completed"
  progress_md: "in-progress -> completed"
b12_self_test_a: "grep -c SUBAGENT-NESTING-DOCTRINE-001 CHANGELOG.md (pre-emission) = 0 -> emission proceeded"
b12_self_test_b: "acceptance.md AC row count (grep -n '^| AC-') = 16; progress.md §E.2 ac_pass_count(15) + ac_pass_with_debt_count(1) = 16 -> match"
b12_self_test_c: "CHANGELOG entry cites .moai/specs/SPEC-SUBAGENT-NESTING-DOCTRINE-001/spec.md and internal/template/templates/ mirror paths; all pre-existing per run-phase §E.2"
canary_compliance_check: not-applicable  # this SPEC does not define a forward-looking policy that its own sync tests
ac_debt_carryover: >
  AC-SND-016 PASS-WITH-DEBT carried through sync unchanged (see §E.3
  ac_debt_reconciliation) — no sync-phase action required; the load-bearing
  settings.json.tmpl invariant is satisfied.
```

# progress.md — SPEC-CC2178-DOCS-ALIGN-001

> Plan-phase progress tracker. `§E.1`은 manager-spec이 plan-phase에서 populate; `§E.2`-`§E.5`는 placeholder heading만 (run/sync/Mx-phase에서 각 소유자가 채움).

## §A — Plan-Phase Metadata

- **SPEC ID**: SPEC-CC2178-DOCS-ALIGN-001
- **Tier**: S (docs-only, 3 milestones, ZERO Go code)
- **lifecycle**: spec-anchored
- **created**: 2026-06-16
- **plan-phase author**: GOOS (manager-spec)
- **artifacts**: spec.md (212 lines) + plan.md + acceptance.md + research.md

## §B — Scope Summary

- **IN**: CC 2.1.169→2.1.178 Tier 1 docs-only 9개 항목을 `.claude/rules/moai/` 6개 파일(template source + mirror) + docs-site 4-locale(ko/en/ja/zh) 페이지에 정합화. 8 REQ, 10 AC(8 기능 + 2 cross-cutting).
- **OUT**: Go 코드 전부(ZERO) / `availableModels` 비용 레버(sibling MODEL-POLICY-REPAIR-001 P2) / `[1m]` constraint 재검증(sibling P3) / Fable 5 tier(별도 전략 SPEC) / CC 공식 문서 수정 / CHANGELOG(sync-phase).
- **Tier verdict**: S — docs-only, < 5 Go files (= 0), 3 milestones. spec-workflow.md § SPEC Complexity Tier 기준 충족.

## §C — Mode Selection (Phase 0.95, populated at run-phase entry)

**Input parameters**:
- tier: S (docs-only, 3 milestones, ZERO Go code)
- scope (file count): 6 rules files (template source + mirror) + 7 docs-site pages × 4 locales = ~34 markdown files
- domain count: 2 (rules markdown + docs-site markdown; both are the same "markdown docs" domain)
- file language mix: 100% markdown (0% Go, 0% shell)
- concurrency benefit: LOW — sequential docs edits within each milestone; no inter-file parallelism benefit (Anthropic coding-task parallelism caveat: most docs tasks involve fewer truly parallelizable subtasks)
- Agent Teams prereqs: N/A (harness level not thorough for this Tier S docs-only SPEC)

**Mode evaluation table**:

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | NO | Not a typo/single-line — 9 features across ~34 files |
| 2 background | NO | Docs edits require Write operations; background agents auto-deny Write/Edit |
| 3 agent-team | NO | domain count = 2 (< 3 threshold); Tier S docs-only; Agent Teams prereqs not met |
| 4 parallel | NO | Not research-heavy; docs edits are sequential within milestone; concurrency benefit LOW |
| 5 sub-agent | NO (but this IS the effective mode) | Sequential milestone-by-milestone execution by manager-develop directly — Mode 5 semantics (one milestone at a time) |
| 6 workflow | NO | Not mechanical-uniform high-volume; docs edits are semantic per-feature; coding-heavy/docs-heavy new-content work stays Mode 5 |

**Decision**: sub-agent (Mode 5 semantics — sequential milestone execution)

**Justification**: Tier S docs-only SPEC with 2 domains (rules + docs-site, both markdown) and LOW concurrency benefit. Per Anthropic's coding-task parallelism caveat (orchestration-mode-selection.md §B: "most coding tasks involve fewer truly parallelizable tasks than research"), the sequential milestone-by-milestone path is the safe default. The orchestrator delegates once to manager-develop which executes M1 → M2 → M3 sequentially with per-milestone commits. No fan-out benefit: each milestone's docs edits are independent of the others but all flow through the same agent, and parallelizing across 4 locales of the same page would risk 4-locale parity drift (AC-DA-009 MUST gate).

## §D — Milestone Progress

- **M1** (permissions + skills discovery): COMPLETE — commit `25796ea9b`. REQ-DA-001~004 documented in rules source+mirror (settings-management.md, skill-authoring.md, agent-authoring.md) + docs-site 4-locale (settings-json.md, skill-guide.md, agent-guide.md).
- **M2** (hooks + agent governance): COMPLETE — commit `b38a28c78`. REQ-DA-005~007 documented in rules source+mirror (hooks-system.md, orchestration-mode-selection.md; agent-authoring.md disallowedTools MCP added in M1) + docs-site 4-locale (hooks-reference.md, hooks-guide.md, agent-guide.md).
- **M3** (session resume / `/cd`): COMPLETE — commit `679c1c72b`. REQ-DA-008 documented in session-handoff.md (mirror-direct) + docs-site 4-locale (statusline.md, moai-sync.md). Block 0 contract preserved; mirror-only doctrine (Diet Constraints/V0 Abort Gate, 6 internal SPEC ID refs) preserved.

### §D.1 M1 — permissions + skills discovery

COMPLETE. 20 files changed (3 rules source + 3 mirror + 12 docs-site 4-locale + spec.md frontmatter + progress.md). All 4 REQ grep verifications PASS at ≥1 across source+mirror+4-locale. Template-first mirror parity achieved via source edit → make build → cp to mirror.

### §D.2 M2 — hooks + agent governance

COMPLETE. 16 files changed (2 rules source + 2 mirror + 12 docs-site 4-locale). REQ-DA-005 post-session, REQ-DA-006 disallowedTools MCP enforcement, REQ-DA-007 auto-mode pre-launch classifier all documented. Accuracy-over-completeness: post-session notes MoAI-ADK does not wire this hook.

### §D.3 M3 — session resume / `/cd`

COMPLETE. 9 files changed (1 mirror-direct session-handoff.md + 8 docs-site 4-locale). REQ-DA-008 `/cd` cache-preserving resume documented as complement to Block 0 new-terminal path (not replacement). session-handoff.md mirror-direct discipline per plan.md §E.4.1: source-first BYPASS, make build NOT run for this file.

## §E.1 Plan-phase Audit-Ready Signal

- **plan_complete_at**: 2026-06-16
- **plan_status**: audit-ready
- **artifacts**: spec.md + plan.md + acceptance.md + research.md (4 files, Tier S docs-only set)
- **frontmatter**: 12 canonical fields validated (id matches `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`; `created`/`updated` not snake_case; `tags` comma-string; `lifecycle: spec-anchored`; `tier: S`)
- **SPEC ID self-check decomposition**: SPEC ✓ | CC2178 ✓ | DOCS ✓ | ALIGN ✓ | 001 ✓ → PASS
- **GEARS compliance**: 8 REQ 전부 GEARS notation (Ubiquitous/Where/When; no deprecated IF/THEN)
- **Exclusions**: spec.md §E에 9개 exclusion 항목 (Go 코드 / sibling 항목 / Fable 5 / CC 공식 문서 / CHANGELOG / 새 파일 / 번들 전수 조사 / troubleshooting 신규 생성 / availableModels)
- **Pre-write checks**: moai spec lint 통과 예정(run-phase 전 최종 확인); precedent SPEC-CC-DOCS-ALIGNMENT-001 패턴 준수
- **Predecessor**: SPEC-CC-DOCS-ALIGNMENT-001 (completed, 동일 docs-only 패턴)
- **Sibling**: SPEC-CC2178-MODEL-POLICY-REPAIR-001 (같은 CC 창, 비용 레버 분리)

## §E.2 Run-phase Evidence

| AC | Status | Verification | Observed |
|----|--------|--------------|----------|
| AC-DA-001 (Tool(param:value) MUST) | PASS | grep "Tool(param:value)" rules source+mirror+4-locale settings-json.md | source=2 mirror=2 en=1 ko=1 ja=1 zh=1 (all ≥1) |
| AC-DA-002 (closest-wins MUST) | PASS | grep -ciE "closest.wins\|closest-directory" skill-authoring+agent-authoring source + skill-guide+agent-guide 4-locale | source skill=2 agent=2; 4-locale skill-guide=2 agent-guide=1 (all ≥1) |
| AC-DA-003 (nested .claude/skills SHOULD) | PASS | grep nested .claude/skills skill-authoring source + skill-guide 4-locale | source=2; en=3 ko=3 ja=3 zh=3 (all ≥1) |
| AC-DA-004 (disableBundledSkills + safe-mode SHOULD) | PASS | grep disableBundledSkills + safe-mode settings-management+skill-authoring source + settings-json+skill-guide 4-locale | source d=2 s=1; 4-locale settings-json d=4 s=1 skill-guide d=2 s=1 (all ≥1) |
| AC-DA-005 (post-session SHOULD) | PASS | grep post-session\|PostSession hooks-system source + hooks-reference+hooks-guide 4-locale | source=1; 4-locale hooks-reference=1 hooks-guide=1 (all ≥1) |
| AC-DA-006 (disallowedTools MCP SHOULD) | PASS | grep disallowedTools.*MCP agent-authoring source + agent-guide 4-locale | source=2; 4-locale=2 (all ≥1) |
| AC-DA-007 (pre-launch classifier SHOULD) | PASS | grep pre.launch.classifier\|auto.mode.*classifier orchestration-mode-selection source (rules-only) | source=2 (≥1) |
| AC-DA-008 (/cd cache-preserving MUST) | PASS | grep /cd + cache-term session-handoff mirror + statusline+moai-sync 4-locale | mirror /cd-cache=2; 4-locale statusline /cd=2 cache-term≥4 moai-sync /cd=2 cache-term≥2 (all present) |
| AC-DA-009 (4-locale parity MUST) | PASS | scripts/docs-i18n-check.sh errors in my pages | 0 of 62 baseline errors in my pages (baseline pre-existing zh H1/glossary gaps out of scope) |
| AC-DA-010 (ZERO Go MUST) | PASS | git diff --name-only 8b5d5d49d HEAD \| grep '\.go$' | empty (0 Go files changed across M1+M2+M3) |

**MUST count**: 5/5 PASS (AC-DA-001, 002, 008, 009, 010). **SHOULD count**: 5/5 PASS (AC-DA-003, 004, 005, 006, 007). **Total**: 10/10 PASS.

## §E.3 Run-phase Audit-Ready Signal

- **run_complete_at**: 2026-06-16
- **run_commit_sha**: 679c1c72b (M3 final; M1=25796ea9b, M2=b38a28c78)
- **run_status**: audit-ready
- **ac_pass_count**: 10
- **ac_fail_count**: 0
- **preserve_list_post_run_count**: 0 (no PRESERVE-list files touched)
- **l44_pre_commit_fetch**: not applicable (docs-only SPEC, no dependency manifest changes)
- **l44_post_push_fetch**: pending push
- **new_warnings_or_lints_introduced**: 0 (spec-lint StatusGitConsistency warning is pre-transition known warning — resolves at sync-phase `in-progress → implemented`)
- **cross_platform_build**: n/a (ZERO Go code, no build artifact)
- **total_run_phase_files**: 45 (20 M1 + 16 M2 + 9 M3)
- **m1_to_mN_commit_strategy**: 3 separate milestone commits (M1 `25796ea9b`, M2 `b38a28c78`, M3 `679c1c72b`), each with Conventional Commits format + Authored-By-Agent: manager-develop trailer + 🗿 MoAI footer
- **template_mirror_parity**: 5 source-first files PARITY ✓; session-handoff.md intentional-DIFFER maintained (doctrine markers=3 ≥3, internal SPEC ID refs=6 ≥6)
- **neutrality_ci**: PASS (go test ./internal/template/... -run TestTemplateNeutralityAudit exit 0)

## §E.4 Sync-phase Audit-Ready Signal

- **sync_complete_at**: 2026-06-17
- sync_commit_sha: 12b81b0d6
- **sync_status**: audit-ready
- **changelog_entry_added**: true (CHANGELOG.md `[Unreleased]` → `### Added`; B12 duplicate pre-check `grep -c 'SPEC-CC2178-DOCS-ALIGN-001' CHANGELOG.md` = 0 before add)
- **status_transition**: in-progress → implemented (manager-docs owned per Status Transition Ownership Matrix; executed orchestrator-direct — see verification_basis)
- **readme_updated**: false (internal docs alignment; no MoAI feature/API/behavior change → README unaffected)
- **docs_site_re_edit**: none (run-phase completed all 4-locale edits; sync-phase does not re-edit docs-site content; AC-DA-009 already PASS)
- **zero_go_code_maintained**: true (AC-DA-010 re-verified against SPEC-scoped baseline `git diff --name-only 8b5d5d49d 827228819 -- '*.go'` = 0 source files, `embedded.go` excluded; sync commit adds 0 `.go` files)
- **ac_count_source**: acceptance.md (SSOT) = 10 AC (5 MUST + 5 SHOULD), NOT progress.md
- **verification_basis**: orchestrator-direct sync execution. manager-docs subagent spawn failed twice with `context window limit` (compact prompt also failed — the session's auto-loaded MoAI rules inflate the subagent system prompt past the window). Per established orchestrator-direct fallback (cf. L_orchestrator_direct_mx_2commit), Tier S docs-only mechanical sync executed directly with B12 discipline (CHANGELOG duplicate pre-check, acceptance.md-sourced AC count, ls path verification, SPEC-scoped ZERO-Go baseline attribution). sync-auditor not independently spawned for the same context-window reason; AC re-verification performed orchestrator-side via objective grep/file-existence checks (4-locale 24-file existence PASS, ZERO-Go scoped PASS, B12 duplicate 0).

### (Migrated from §E.5)

- **mx_complete_at**: 2026-06-17
- mx_commit_sha: be0eafe03
- **final_status**: completed
- **4-phase close**: plan (`8b5d5d49d`) + run (M1 `25796ea9b` / M2 `b38a28c78` / M3 `679c1c72b` / run-progress `827228819`) + sync (`12b81b0d6` + backfill `eb7d6038f`) + Mx (`be0eafe03` + backfill `cde0d062f`) — all phases complete
- **era_classification**: V3R6 (H-4: §E.2 run-evidence + §E.5 Mx + `sync_commit_sha` + `mx_commit_sha` all present)
- **ac_final**: 10/10 PASS (5 MUST: AC-DA-001/002/008/009/010 + 5 SHOULD: AC-DA-003/004/005/006/007); spec-lint 0 findings
- **closure**: SPEC-CC2178-DOCS-ALIGN-001 docs-only alignment complete — CC 2.1.169→2.1.178 Tier 1 items documented across rules source+mirror + docs-site 4-locale (ko/en/ja/zh), ZERO Go code
- **deferred** (separate future SPECs, out of scope per spec.md §E): EFFORT-MAP-RETIREMENT / TASK-TRIAGE / vff prose-discipline
- **mx_executed_by**: orchestrator-direct (manager-docs spawn context-window-blocked; Mx chore close is canonical- owner-valid for orchestrator per ownership matrix)

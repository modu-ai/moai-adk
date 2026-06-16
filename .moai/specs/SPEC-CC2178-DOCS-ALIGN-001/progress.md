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

_<pending run-phase>_

### §D.1 M1 — permissions + skills discovery

_<pending run-phase>_

### §D.2 M2 — hooks + agent governance

_<pending run-phase>_

### §D.3 M3 — session resume / `/cd`

_<pending run-phase>_

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

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending run-phase>_

## §E.5 Mx-phase Audit-Ready Signal

_<pending run-phase>_

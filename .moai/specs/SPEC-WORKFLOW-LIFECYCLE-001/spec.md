---
id: SPEC-WORKFLOW-LIFECYCLE-001
title: "SPEC 워크플로 P1 개선 — delta-spec 수명주기(completed 재개정 전이), depends_on run-phase 집행(pre-flight 차단), Tier L 산출물 집합 + plan-auditor 입력 계약 확정"
version: "0.1.0"
status: in-progress
created: 2026-07-09
updated: 2026-07-09
author: GOOS행님
priority: P1
phase: "v3.1.x"
module: ".claude/agents/moai + .claude/rules/moai/workflow + internal/template/templates"
lifecycle: spec-anchored
tier: L
tags: "workflow, lifecycle, delta-spec, depends-on, tier-l, plan-auditor-input, doc-only, template-mirror"
---

# SPEC-WORKFLOW-LIFECYCLE-001 — SPEC 워크플로 P1 개선

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-07-09 | GOOS행님 (via manager-spec) | 초안 — SPEC-AUDIT-GATE-INTEGRITY-001 P1 백로그 3건 통합 (delta-spec 수명주기 + depends_on 집행 + Tier L 산출물/plan-auditor 입력 계약). 감사 원 보고서 review-{1,3}.md + sync-audit 2026-07-09 + memory `audit-gate-integrity-001-completed`의 P1 백로그 근거. |

## §A Context

### A.1 문제 정의

SPEC-AUDIT-GATE-INTEGRITY-001 (sync 61b7bcc0a, sync-audit 0.99 PASS)의 3기 병렬 감사는 P0 4건(감사 게이트 무결성)을 완결했고, P1 3건을 후속 SPEC 백로그로 남겼다. 본 SPEC은 그 P1 3건을 단일 Tier L SPEC으로 통합한다. 공통 병리는 **"SPEC 워크플로의 수명주기 골격이 불완전하다"** — completed 이후 개정 전이가 없고, SPEC 간 의존 선언이 집행되지 않으며, Tier L 산물 집합과 plan-auditor 입력 계약이 명시적 SSOT로 확정되지 않았다.

3개 항목은 워크플로 수명주기의 서로 다른 축을 커버한다:
- **R1 (delta-spec)**: 시간 축 — completed SPEC의 사후 개정 메커니즘 부재
- **R2 (depends_on)**: 의존성 축 — 선언적 의존 관계의 강제 부재
- **R3 (Tier L 산물 + plan-auditor 입력 계약)**: 산물 축 — 5-file 집합과 plan-auditor 소비 계약의 SSOT 부재

### A.2 SDD 근거 (외부 표준 정합)

- **OpenSpec propose→apply→archive 패턴**: SDD 웹 리서치(Spec Kit / Kiro / EARS / OpenSpec, 2026-07-09 수행)에서 도출. OpenSpec의 3단계 개정 모델은 "완료된 명세의 사후 개정"을 정규화한다 — 본 SPEC R1은 이를 MoAI의 3-phase lifecycle에 맞게 재단한다: `propose` = `completed → in-progress` 전이 + HISTORY 행, `apply` = run→sync 재착수, `archive` = sync close 시 prior_completed_sha 보존.
- **GitHub Spec Kit cross-artifact consistency**: 의존성 선언은 게이트에서 강제되어야 한다 (advisory가 아니라 block). 본 SPEC R2가 이 원칙을 `depends_on`에 적용한다.
- **Kiro 3-artifact 모델 + Tier L 5-artifact 확장**: Kiro는 requirements/design/tasks 3-artifact를 표준으로 사용; MoAI는 LEAN 도입 시 이를 5-artifact(Tier L)로 확장했으나 plan-auditor 입력 계약은 3-artifact 시절의 "spec.md primary + acceptance/plan cross-ref"에 머물러 있다. 본 SPEC R3가 입력 계약을 Tier-differentiated로 재단한다.
- **`verification-claim-integrity.md` §1.1 surface 3**: 결함 주장은 도구 검증 전까지 가설이다. 본 SPEC의 research.md는 모든 gap 주장을 grep/Bash 실측으로 뒷받침한다 (AC-001를 R3 설계의 일부로 명시).

### A.3 결함 증거 (실측 anchor, 2026-07-09 기준 — research.md §B 전수 인용)

| # | 결함 | 증거 위치 (실측) |
|---|------|------------------|
| R1 | completed SPEC의 사후 개정 메커니즘 부재 | `internal/spec/status.go` `ValidStatuses` (8값, `amended` 부재); `internal/spec/audit.go:355-356` "If status is already completed, no drift"; `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (`completed → completed` 경로 부재, `* → superseded` 만 존재) |
| R2 | `depends_on`의 run-phase 강제 부재 | `internal/bodp/relatedness.go:177-180` (`DependsOn []string` 파싱 → Signal A branch-origin 결정에만 사용); `internal/runtime/audit_cache.go` + `audit_gate.go` + `audit_report.go` 전수 grep: `depends_on` 참조 0건 → Phase 0.5 plan-audit gate가 depends_on을 소비하지 않는다 |
| R3a | Tier L 산물 집합이 SSOT에 명시적이지 않음 | `.claude/rules/moai/workflow/spec-workflow.md:137` (5-file 집합 서술은 존재하나 SSOT인 spec-frontmatter-schema.md의 `tier:` optional 필드 설명에는 산물 목록 부재) |
| R3b | plan-auditor 입력 계약이 Tier-differentiated가 아님 | `.claude/agents/moai/plan-auditor.md:470-476` § Input Contract: "reads `spec.md` as primary. It may also read `acceptance.md` and `plan.md` for cross-reference" — design.md/research.md 언급 0건. Tier L 5-file 집합 중 2개가 입력 계약에서 누락 |
| R3c | plan-artifact hash subject list의 doctrine-vs-Go drift | `.claude/rules/moai/workflow/spec-workflow.md:313-319` (doctrine: 5-file prose: "spec.md / plan.md / acceptance.md / research.md / design.md") vs `internal/runtime/audit_cache.go:63-68` `planArtifactNames = ["acceptance.md", "plan.md", "spec.md", "tasks.md"]` (Go: 4-file, **tasks.md는 Tier L 산물 아님**, design/research 제외) |

### A.4 수정 방침

전면 doc-only (agent 정의 + rule/doctrine 편집 + template mirror). Go 코드 동작 무변경. R1은 frontmatter schema 확장(Optional 필드) + audit 동작 서술(구현은 Out of Scope); R2는 Phase 0.5 sub-step 서술(Go 구현은 Out of Scope); R3는 SSOT 명문화 + 입력 계약 확장. 모든 편집은 live → template mirror(Template Content Neutrality strip) → `make build` 순서.

## §B Requirements (GEARS)

### R1 — delta-spec 수명주기: completed 사후 개정 전이

채택 설계: **completed → in-progress 재전이 + HISTORY 기반 개정 기록** (새 `amended` 상태 추가 안함; `amendment_of:` Optional 필드 도입; 사후 개정은 git history의 `completed` 선행 상태로 감지). 기각 대안: (a) 9번째 `amended` 상태 추가 — enum 인플레이션 + 감사/era 분류 호환성 비용; (b) completed 상태에서 직접 새 SPEC 으로 supersede — 기존 `superseded_by` 패턴의 재탕이며 사후 개정(not replacement) 시나리오를 커버 못함.

#### REQ-WFL-001
The SPEC lifecycle (`internal/spec/status.go` `ValidStatuses` + `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum) shall support an in-place amendment transition where a SPEC with `status: completed` may transition back to `status: in-progress` **only when** the SPEC's HISTORY section records the prior completed version row (version, date, author, prior-completed rationale) AND the frontmatter carries an `amendment_of:` optional field whose value is either the SPEC's own ID (self-referential in-place amendment) or a parent SPEC ID (successor amendment pattern).

#### REQ-WFL-002
**When** a SPEC is amended in-place (completed → in-progress per REQ-WFL-001), the spec.md body shall carry an `## Amendments` sub-section under HISTORY recording: (a) the prior completed version string, (b) the prior completed commit SHA (or `unknown` if pre-git), (c) the amendment rationale (one paragraph), (d) the amendment scope (list of affected §B REQ IDs). The `## Amendments` sub-section is additive — the original HISTORY rows are preserved verbatim, and amendment rows append below them with monotonically increasing version.

#### REQ-WFL-003
The Status Transition Ownership Matrix (`.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix) shall carry a new row for the `completed → in-progress (amendment)` transition owned by manager-spec (re-delegation per the D-NEW-1 inline-fix pattern), with canonical commit subject pattern `feat(SPEC-{ID}): in-place amendment <rationale-summary>` — distinct from the existing `(none) → draft` plan-phase creation row and the `* → superseded` row.

#### REQ-WFL-004
**While** a SPEC is in the amendment transition (status: in-progress, `amendment_of:` set, HISTORY `## Amendments` populated), `moai spec audit` SHALL treat the SPEC as V3R6 modern era (subject to drift detection) — the `era:` frontmatter field, if present, is preserved verbatim; the `prior_completed_sha:` recorded in HISTORY enables drift detectors to distinguish "amendment in progress" from "stale implemented/completed drift" (the existing `internal/spec/drift.go:121` predicate for `frontmatterStatus == "completed" && gitStatus != "completed"` does NOT fire because frontmatter is `in-progress` during amendment). This requirement is doc-only — Go drift-detector enhancement is Out of Scope.

### R2 — depends_on run-phase 집행: pre-flight 차단

채택 설계: **Phase 0.5 sub-step "Depends_on Pre-flight Check" 확장** (plan-auditor 호출 직전에 실행; 실패 시 blocker report). 기각 대안: (a) 별도 Phase 0.6 신설 — 단계 인플레이션 + 기존 0.5를 건너뛰는 우회 경로; (b) BODP Signal A 확장 — BODP는 branch-origin 결정(branch 만들까 말까)이지 run 진입 허가(run 허락할까 말까)가 아니라 의미 불일치.

#### REQ-WFL-005
**When** `/moai run <SPEC-ID>` is invoked, the orchestrator SHALL execute a depends_on pre-flight check as the first sub-step of Phase 0.5 (Plan Audit Gate), BEFORE the plan-auditor subagent invocation. The pre-flight loads the SPEC's frontmatter `depends_on:` list (Optional field per spec-frontmatter-schema.md § Optional Fields); **Where** `depends_on` is absent or empty, the pre-flight trivially PASSes and proceeds to the plan-auditor step; **Where** `depends_on` lists one or more SPEC IDs, the pre-flight SHALL resolve each dependency's current `status:` frontmatter field by reading `.moai/specs/<dep-ID>/spec.md`.

#### REQ-WFL-006
For depends_on pre-flight enforcement, dependency fulfillment SHALL be defined strictly as the dependency SPEC's `status: completed` — all other 7 status values (draft, planned, in-progress, implemented, superseded, archived, rejected) SHALL be considered unfulfilled. The evaluation is deterministic per-status: no partial credit, no "near-completed" interpretation, no score-based bypass.

#### REQ-WFL-007
**When** one or more `depends_on` entries are unfulfilled (per REQ-WFL-006), the run-phase pre-flight SHALL NOT proceed to the plan-auditor step; the orchestrator SHALL surface a structured blocker via `AskUserQuestion` with three options: (a) **wait** (abort run; re-invoke after deps complete), (b) **override** (proceed with `--ignore-deps` flag; logged to `.moai/logs/depends-on-override.log` per audit trail), (c) **abort** (cancel run). The override path MUST record the unfulfilled dependency IDs + override rationale in the log; a bare `--ignore-deps` without the logged rationale is prohibited.

### R3 — Tier L 산물 집합 + plan-auditor 입력 계약 확정

채택 설계: **3면 SSOT 동기화** — (a) `spec-frontmatter-schema.md` `tier:` Optional 필드에 산물 목록 명시, (b) `plan-auditor.md` § Input Contract를 Tier-differentiated로 재작성, (c) plan-artifact hash subject list를 Go 구현 정합으로 명문화 (design/research는 manual skip 입력으로 명시, tasks.md는 V3R4 잔재로 서술). 기각 대안: Go의 `planArtifactNames`를 5-file로 확장 — Go 변경은 본 SPEC 범위 밖 (doc-only 원칙 위반).

#### REQ-WFL-008
The canonical Tier L artifact set SHALL be codified as exactly 5 files: `spec.md` + `plan.md` + `acceptance.md` + `design.md` + `research.md`. The `tier:` Optional field documentation in `.claude/rules/moai/development/spec-frontmatter-schema.md` § Optional Fields SHALL enumerate this 5-file set explicitly (cross-referencing `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier, which is the existing authoritative source). Tier S = 2 files (spec.md + plan.md, AC inline); Tier M = 3 files (spec.md + plan.md + acceptance.md); Tier L = 5 files (the canonical set above). The progress.md file is a lifecycle-tracking artifact owned by manager-spec/develop/docs per the phase — it is NOT counted toward the Tier artifact count.

#### REQ-WFL-009
The plan-auditor agent definition (`.claude/agents/moai/plan-auditor.md` § Input Contract) SHALL be Tier-differentiated: **Where** the SPEC under audit declares `tier: L` (or tier is absent, defaulting to Tier L for backward compat), the plan-auditor SHALL read `design.md` AND `research.md` in addition to the primary trio (spec.md / plan.md / acceptance.md); **Where** `tier: M`, the plan-auditor reads the primary trio only; **Where** `tier: S`, the plan-auditor reads spec.md + plan.md (AC inline in spec.md). The Input Contract section shall enumerate this Tier-differentiated artifact set with literal tokens **"Tier L: design.md + research.md are required inputs"** and **"Tier-differentiated input contract"**.

#### REQ-WFL-010
The plan-artifact hash subject list SHALL be codified in `.claude/rules/moai/workflow/spec-workflow.md` § Report Persistence as the 4-file set `{spec.md, plan.md, acceptance.md, tasks.md}` — matching `internal/runtime/audit_cache.go` `planArtifactNames` verbatim (Go 구현 정합). The doctrine shall additionally document: (a) `tasks.md` is a V3R4-era plan artifact name retained in the hash subject list for backward compatibility with grandfathered SPECs (V3R6 Tier L replaces it with design.md + research.md, which are NOT hash subjects); (b) `design.md` and `research.md` are **conservative manual-skip judgment inputs** (리터럴 토큰 `manual-skip judgment inputs`) — changes to them do NOT mechanically invalidate a cached skip verdict but MUST be considered by the orchestrator's manual skip decision alongside the 4-file hash. The skip-eligibility 4-condition predicate (§ Phase Transitions Plan to Run) is otherwise unchanged.

#### REQ-WFL-011
**When** a SPEC is amended in-place per REQ-WFL-001 (completed → in-progress, `## Amendments` HISTORY row added), the plan-artifact hash SHALL change because `spec.md` is modified — this invalidates any cached plan-auditor PASS verdict for the SPEC, forcing Phase 0.5 plan-audit re-execution on the next `/moai run`. The doctrine shall state this invalidation explicitly in § Report Persistence so orchestrators understand amendment is a cache-invalidating event (리터럴 토큰 `cache-invalidating event`).

### Cross-cutting — Template Mirror 동기화

#### REQ-WFL-012
**When** any live file touched by REQ-WFL-001..011 has a template mirror under `internal/template/templates/` (실측 확인 대상: `plan-auditor.md`, `manager-spec.md` — `internal/template/templates/.claude/agents/moai/`; `spec-frontmatter-schema.md`, `spec-workflow.md` — `internal/template/templates/.claude/rules/moai/`), the same edit shall be applied to the mirror with the Template Content Neutrality strip (내부 SPEC ID, REQ-WFL 토큰, 감사 인용, 내부 날짜/SHA 제거 — CI guard `template-neutrality-check.yaml`), followed by `make build` exiting 0.

## §C Constraints

- **doc-only**: `internal/spec/status.go`, `internal/spec/audit.go`, `internal/spec/drift.go`, `internal/runtime/audit_cache.go`, `internal/runtime/audit_gate.go`, `internal/runtime/audit_report.go`, `internal/bodp/relatedness.go` 등 Go 코드 동작 무변경. `make build`는 임베드 재컴파일 목적일 뿐 동작 변경이 아니다.
- **AC 기계 검증 의무**: 모든 AC는 정확한 명령 + 기대 출력(카운트/exit code)을 명기 (`verification-claim-integrity.md` §3.2).
- **Template Content Neutrality**: mirror 편집 시 `SPEC-WORKFLOW-LIFECYCLE` 토큰·REQ-WFL 토큰·내부 날짜가 template 트리에 유입 금지 (CLAUDE.local.md §2.1 + §25).
- **grep-stable 토큰 설계**: 수정으로 도입되는 문구는 AC grep이 안정적으로 잡을 리터럴 토큰(`amendment_of:`, `## Amendments`, `prior_completed_sha`, `Depends_on Pre-flight`, `--ignore-deps`, `depends-on-override.log`, `fulfillment`, `Tier-differentiated input contract`, `manual-skip judgment inputs`, `cache-invalidating event`)을 포함해야 한다.
- **언어 중립(템플릿 배포 자산)**: REQ-WFL-009의 Tier-differentiated 입력 계약 설명은 어떤 언어도 PRIMARY로 승격하지 않는다 (CLAUDE.local.md §15 16-언어 동등 [HARD]).
- **frontmatter 12 canonical fields** + GEARS compound-clause 형식 준수 (plan-auditor MP-2/MP-3 통과 요건).
- **이미 존재하는 산물에 대한 신규 발견 금지** (feedback_template_tree_is_subset_of_live, feedback_hypothesis_as_defect): 본 SPEC의 모든 gap 주장은 research.md에서 grep/Bash 실측으로 뒷받침된다.

## Out of Scope (제외 범위)

### Out of Scope — Go 코드 동작 변경
- `internal/spec/status.go` `ValidStatuses` 배열 확장 (새 `amended` 상태 추가 안함 — REQ-WFL-001는 기존 `in-progress` 재전이로 해결)
- `internal/spec/audit.go` / `drift.go`의 completed 상태 감사 로직 변경 — REQ-WFL-004는 서술만, 구현은 후속 SPEC
- `internal/runtime/audit_cache.go` `planArtifactNames` 4-file → 5-file 확장 — REQ-WFL-010는 현행 Go 정합을 그대로 명문화
- `internal/bodp/relatedness.go` `parseDependsOn` 재사용 여부 — REQ-WFL-005는 orchestrator-side pre-flight 서술이지 BODP 확장이 아님
- Phase 0.5 plan-audit gate의 Go 구현 확장 (depends_on pre-flight Go 코드) — REQ-WFL-005..007는 doctrine 서술만

### Out of Scope — P2 위생 백로그
- SPEC-AUDIT-GATE-INTEGRITY-001 P2 백록 5건 (plan 단계 번호 재정비, moai-workflow-spec/SKILL.md stale 표기 2건, sprint-round-naming.md:159 앵커 오류, planned legacy 경고 전파, [NEEDS CLARIFICATION] 마커 규약 도입 검토)은 본 SPEC 범위 밖 — 별도 위생 SPEC 소관

### Out of Scope — depends_on cycle detection / topological sort
- `depends_on` 그래프의 사이클 탐지 / 위상 정렬은 본 SPEC 범위 밖 — pre-flight는 각 dep의 status를 개별 조회할 뿐, 그래프 순회 로직은 포함하지 않는다 (후속 SPEC이 필요한 경우 별도로 설계)

### Out of Scope — amendment의 자동 감지 훅
- git history 기반 "이 SPEC은 과거 completed였는가?" 자동 감지 PostToolUse hook은 본 SPEC에서 만들지 않는다 — REQ-WFL-001는 `amendment_of:` 필드와 HISTORY `## Amendments` 섹션으로 명시적 선언을 요구한다 (자동 감지는 Out of Scope)

### Out of Scope — Tier S/M에 대한 design/research Optional 허용
- Tier S/M SPEC이 design.md/research.md를 가져도 되는지 여부는 본 SPEC이 다루지 않는다 — Tier 판정은 spec-assembly.md Socratic question + 작성자 판단 소관이며, 본 SPEC은 "Tier L은 5개"를 확정할 뿐 나머지 Tier의 산물 확장 금지를 서술하지 않는다

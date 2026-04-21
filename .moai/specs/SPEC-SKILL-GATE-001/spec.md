---
id: SPEC-SKILL-GATE-001
version: 1.0.0
status: completed
created_at: 2026-04-21
updated_at: 2026-04-21
author: moai-adk-go
priority: medium
labels: [workflow, skill, cleanup, claude-code-integration, batch, simplify]
issue_number: null
depends_on: []
related_specs: [SPEC-DB-SYNC-HARDEN-001]
---

# SPEC-SKILL-GATE-001: 도달 불가 / 비신뢰 스킬 참조 일괄 제거

## HISTORY

- 2026-04-21 v1.0.0: 사용자 결정에 따라 **removal-only** 시나리오로 확정. 초안의 "Stop hook 강제 게이트 구축" 접근은 폐기. `Skill("batch")`는 Claude Code 바이너리 `disableModelInvocation: true` 제약으로 모델 호출 자체가 불가능하고, `Skill("simplify")`는 기술적으로 호출 가능하나 91개 세션 전수 조사에서 0회 호출 — 강제 인프라를 새로 만들지 않는 한 신뢰 불가. 두 스킬의 워크플로우 MANDATORY 지시 전부를 제거.
- 2026-04-21 v0.1.0: 초안 — 두 접근(제거 vs 강제) 비교. manager-spec이 Stop hook 기반 강제 게이트를 권장했으나 사용자가 범위를 단순화하여 removal-only로 선택.

## Background

Claude Code v2.1.116 바이너리에 번들된 두 built-in 스킬에 대한 MoAI 워크플로우 지시가 실제 런타임에서 발동되지 않음을 2026-04-21 감사에서 확인했다.

**감사 증거** (사용자 제공):

1. 바이너리 문자열 추출 — 두 스킬의 플래그:
   ```
   name:"batch", userInvocable:!0, disableModelInvocation:!0
   name:"simplify", userInvocable:!0          ← disableModelInvocation 미설정
   ```

2. 세션 로그 전수 조사 — `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/**/*.jsonl` 내 `"skill":"…"` 필드 집계:
   ```
   91 "skill":"moai"           (다수 세션)
    2 "skill":"moai:sync"
    3 "skill":"moai:plan"
    3 "skill":"moai:loop"
    1 "skill":"moai:mx" / moai:clean / moai-workflow-thinking …
    0 "skill":"simplify"       ← 한 번도 호출된 적 없음
    0 "skill":"batch"          ← 한 번도 호출된 적 없음 (예상 — 불가능)
   ```

3. Claude Code v2.1.116의 available-skills 목록(런타임에 노출됨)에서 `simplify`는 존재, `batch`는 부재 — `disableModelInvocation: true` 효과 재확인.

**결론**:

- `/batch`: 모델이 구조적으로 호출할 수 없음. `run.md`/`clean.md`/`mx.md`/`coverage.md`의 "MoAI MUST evaluate … Skill(\"batch\")" 지시는 영구히 사문화된 dead instruction.
- `/simplify`: 모델 툴콜 레이어에서는 호출 가능하지만, MoAI 워크플로우의 MANDATORY 지시가 긴 context 내에서 일관되게 무시된다(감사 기간 전체 0% 발동률). 신뢰 가능한 자동 발동을 위해서는 Stop hook + JSONL 파서 + 강제 게이트 같은 새 인프라(~300-500 LOC)를 구축해야 하며, 이는 본 SPEC 범위 밖.

사용자의 판단: "자동 발동이 불가능하거나 실무상 실패하면 두 참조를 모두 제거한다". 본 SPEC은 이 결정을 구현한다.

### Scope Boundary

- **In scope**: MoAI 프로젝트가 관리하는 워크플로우 문서와 rule 문서에서 `Skill(\"batch\")` · `Skill(\"simplify\")` 참조 **전수 제거**.
- **Out of scope**:
  - Claude Code 바이너리 수정 (불가능; Anthropic 유지).
  - `/simplify` 강제 발동 인프라 구축 (Stop hook / JSONL 파서 / 게이트 이벤트) — 향후 별도 SPEC이 필요.
  - `Skill(\"moai-workflow-loop\")`, `Skill(\"loop\")` 등 MoAI 자체 스킬 호출 경로는 영향 없음.
  - CLAUDE.md / CLAUDE.local.md — grep으로 `/batch`, `/simplify` 직접 참조 없음 확인(2026-04-21).

## Requirements (EARS)

### R1 — `/batch` 참조 제거 (4 파일)

- **REQ-BATCH-001** (Ubiquitous): The files `.claude/skills/moai/workflows/run.md`, `clean.md`, `mx.md`, `coverage.md` SHALL contain zero occurrences of the literal string `Skill(\"batch\")`.
  - **Rationale**: `disableModelInvocation: true` 바이너리 제약으로 모델 호출 자체가 불가능하므로 지시 자체가 오도 소지.

- **REQ-BATCH-002** (Ubiquitous): The same four files SHALL NOT contain a top-level section titled `### Batch Mode Decision [MANDATORY EVALUATION]` or any equivalent heading that instructs the orchestrator to "evaluate" or "execute" batch mode.
  - **Rationale**: 섹션 헤더와 주변 "MANDATORY EVALUATION" 문구를 통해 미래의 리팩토링/감사가 다시 같은 오류를 재도입하지 않도록 구조적 차단.

- **REQ-BATCH-003** (Event-driven): WHEN the `run`/`clean`/`mx`/`coverage` workflow starts, THEN the orchestrator MUST proceed directly to the workflow's main sequential phase without any batch-mode decision step.
  - **Rationale**: 제거 후 흐름이 자연스럽게 연결됨을 확인하는 행동 계약.

### R2 — `/simplify` 참조 제거 (3 파일)

- **REQ-SIMPLIFY-001** (Ubiquitous): The files `.claude/skills/moai/workflows/run.md`, `sync.md`, `review.md` SHALL contain zero occurrences of the literal string `Skill(\"simplify\")`.
  - **Rationale**: 모든 MANDATORY 지시가 감사 기간 0% 발동률을 기록했으므로, 신뢰 가능한 강제 인프라를 구축할 때까지 dead instruction을 제거.

- **REQ-SIMPLIFY-002** (Ubiquitous): `run.md` SHALL NOT contain the section `### Phase 2.10: Simplify Pass [MANDATORY]`. `sync.md` SHALL NOT contain the section `### Phase 0.05: Code Simplification Review`. `review.md` SHALL NOT contain `### Simplify Pass [MANDATORY EVALUATION]`.
  - **Rationale**: 섹션 헤더 수준의 정적 검증.

- **REQ-SIMPLIFY-003** (Ubiquitous): `.claude/rules/moai/workflow/workflow-modes.md` SHALL contain zero occurrences of `Skill(\"simplify\")` and zero cross-references to "Phase 2.10".
  - **Rationale**: DDD/TDD 섹션의 "After IMPROVE/REFACTOR: Skill(\"simplify\") executes automatically" 지시와 Pre-submission Self-Review 섹션의 "runs after Skill(\"simplify\")" 문구까지 함께 제거되어야 교차 참조 고아가 생기지 않음.

- **REQ-SIMPLIFY-004** (Ubiquitous): `sync.md`의 "Completion Criteria" 목록에서 `Phase 0.05: Code simplification review completed` 항목이 제거되어야 한다.
  - **Rationale**: Completion Criteria는 워크플로우 종료 gate 역할이므로 존재하지 않는 phase를 참조하면 gate 자체가 깨진다.

### R3 — 문서 일관성 보존

- **REQ-CONSISTENCY-001** (Ubiquitous): The phase numbering in `run.md` and `sync.md` SHALL remain visually continuous after removal. Phase 2.9 → LSP Quality Gates → Phase 3 (run) and Phase 0 → Phase 0.1 → Phase 0.5 (sync) are the post-removal orderings. Gap 번호(2.10, 0.05)는 의도적 삭제이며 HISTORY와 본 SPEC이 추적한다.
  - **Rationale**: 번호 gap 자체는 허용되나(기존 MoAI 워크플로우에도 소수점 gap 존재) 삭제 근거가 SPEC으로 연결되어야 함.

- **REQ-CONSISTENCY-002** (Ubiquitous): No workflow file SHALL retain a dangling "Pre-submission Self-Review runs after Skill(\"simplify\")" clause.
  - **Rationale**: R2의 `REQ-SIMPLIFY-003`과 중복되지 않는 교차 참조 정밀 검증.

## Acceptance Criteria

- **AC-1 (R1 grep)**: `grep -rn 'Skill(\"batch\")' .claude/` returns zero lines.
- **AC-2 (R1 structural)**: `grep -rn '### Batch Mode Decision' .claude/` returns zero lines.
- **AC-3 (R2 grep)**: `grep -rn 'Skill(\"simplify\")' .claude/` returns zero lines.
- **AC-4 (R2 structural)**: The following greps all return zero lines:
  - `grep -n '### Phase 2.10' .claude/skills/moai/workflows/run.md`
  - `grep -n '### Phase 0.05' .claude/skills/moai/workflows/sync.md`
  - `grep -n '### Simplify Pass' .claude/skills/moai/workflows/review.md`
- **AC-5 (R2 completion criteria)**: `grep -n 'Phase 0.05' .claude/skills/moai/workflows/sync.md` returns zero lines (covers both section header and Completion Criteria list item).
- **AC-6 (R3 cross-reference cleanup)**: `grep -rn 'Phase 2.10' .claude/` returns zero lines.
- **AC-7 (Regression shield)**: The token "Phase 2.9" and "Phase 3" appear in adjacent sections of `run.md` (visual continuity) — manual spot check.
- **AC-8 (CLAUDE docs unchanged)**: `grep -n '/batch\|/simplify\|Skill(\"batch\")\|Skill(\"simplify\")' CLAUDE.md CLAUDE.local.md` returns zero lines (no change needed since pre-audit they were already absent).
- **AC-9 (No orphan references)**: `grep -rn 'batch mode\|simplify pass' .claude/ --include=\"*.md\" -i` returns zero lines for the removed concepts (case-insensitive).
- **AC-10 (CI/build regression)**: `make build` succeeds; `go test ./...` all green. Template deployment does not regress.

## Exclusions (What NOT to Build)

- `/simplify` 자동 발동 강제 인프라 (Stop hook, JSONL 파서, 게이트 이벤트) — 향후 별도 SPEC 필요.
- Claude Code 바이너리 수정 — 불가능, Anthropic 책임.
- 다른 MoAI 스킬(`moai-workflow-loop`, `moai-workflow-gan-loop` 등) 호출 경로 변경.
- `/moai fix`, `/moai loop` 등 다른 워크플로우 커맨드의 내부 로직 변경.
- `.moai/config/sections/*.yaml` 설정 키 추가/제거.
- Go 소스 코드 변경(`internal/hook/`, `internal/cli/` 등) — 본 SPEC은 문서만 건드림.

## Target Files (scope discipline)

구현 PR이 수정할 수 있는 파일은 다음 7개로 한정된다:

1. `.claude/skills/moai/workflows/run.md` — Batch Mode Decision 섹션(lines 378-396) + Phase 2.10 Simplify Pass 섹션(lines 698-715) 제거
2. `.claude/skills/moai/workflows/sync.md` — Phase 0.05 Code Simplification Review 섹션(lines 140-155) + Completion Criteria 목록의 Phase 0.05 항목 제거
3. `.claude/skills/moai/workflows/clean.md` — Batch Mode Decision 섹션(lines 131-146) 제거
4. `.claude/skills/moai/workflows/mx.md` — Batch Mode Decision 섹션(lines 111-125) 제거
5. `.claude/skills/moai/workflows/coverage.md` — Batch Mode Decision 섹션(lines 112-126) 제거
6. `.claude/skills/moai/workflows/review.md` — Simplify Pass 섹션(lines 148-160) 제거
7. `.claude/rules/moai/workflow/workflow-modes.md` — 4개 `Skill(\"simplify\")` 교차 참조(lines 38, 65, 89, 101) 정리 (섹션 삭제가 아닌 문구 수정)

외부 파일은 건드리지 않는다(Agent Core Behavior #5 Scope Discipline).

## Risks

- **R-1 (문서 번호 gap 혼란)**: Phase 2.9 다음에 Phase 3이 오고 Phase 2.10이 사라지면 독자가 "누락인가" 의심할 수 있음. **완화**: 커밋 메시지와 SPEC HISTORY에 명시적으로 "SPEC-SKILL-GATE-001에서 제거"를 기록하여 git blame/log 추적 가능.
- **R-2 (외부 문서의 잔여 참조)**: docs-site(hextra) 등 외부 사이트가 Phase 2.10을 언급할 수 있음. **완화**: 본 SPEC의 Target Files는 internal workflow docs만이며, docs-site는 향후 `/moai sync`에서 별도 추적.
- **R-3 (template 회귀)**: `.claude/skills/...`는 template이므로 `make build` 재생성이 필요. **완화**: AC-10에서 build 성공을 검증. 또한 현재 세션에서 template 소스 내 `.claude/skills/` 경로를 수정했으므로 동일 파일이 embedded.go에도 동기화됨.
- **R-4 (향후 `/simplify` 강제 인프라 도입 시 재작성)**: 만약 Stop hook 기반 강제 게이트를 향후 구축하게 된다면 제거된 섹션들을 재작성해야 함. **완화**: 그 시점의 요구사항에 맞춰 깔끔하게 새로 쓰는 편이 dead instruction을 유지하는 것보다 낫다.

## Traceability

### REQ ↔ AC Matrix

| REQ | AC | 검증 초점 |
|-----|------|-----------|
| REQ-BATCH-001 | AC-1 | `Skill(\"batch\")` 리터럴 0건 |
| REQ-BATCH-002 | AC-2 | `### Batch Mode Decision` 헤더 0건 |
| REQ-BATCH-003 | AC-7, AC-10 | 흐름 연속성 + build 정상 |
| REQ-SIMPLIFY-001 | AC-3 | `Skill(\"simplify\")` 리터럴 0건 |
| REQ-SIMPLIFY-002 | AC-4 | Phase 2.10 / 0.05 / Simplify Pass 헤더 0건 |
| REQ-SIMPLIFY-003 | AC-3, AC-6 | workflow-modes.md 내 참조 0건 |
| REQ-SIMPLIFY-004 | AC-5 | sync.md Completion Criteria 정합성 |
| REQ-CONSISTENCY-001 | AC-7 | phase 번호 시각 연속성 |
| REQ-CONSISTENCY-002 | AC-6, AC-9 | 교차 참조 고아 없음 |

9개 REQ, 10개 AC. 모든 REQ가 최소 1개 AC로 검증됨. 회귀 방어는 AC-10.

### 감사 → REQ 매핑

| 감사 발견 | REQ 패밀리 | 제거 위치 |
|-----------|-----------|-----------|
| `/batch` disableModelInvocation: true | REQ-BATCH-* | run/clean/mx/coverage의 4개 섹션 |
| `/simplify` 91 세션 0% 발동 | REQ-SIMPLIFY-* | run/sync/review의 3개 섹션 + workflow-modes 4개 line |
| Phase 2.10 교차 참조 고아 방지 | REQ-CONSISTENCY-* | workflow-modes의 Pre-submission Self-Review 정의 |

## Configuration Reference

- 바이너리 증거: Claude Code v2.1.116 (2026-04-21 세션에서 확인)
- 사용자 감사 보고: 91 세션 JSONL 집계 (2026-04-21)
- SPEC-DB-SYNC-HARDEN-001: 직전 SPEC의 sync 단계에서 본 결함이 노출됨
- `.moai/config/sections/language.yaml`: `code_comments: ko` — Background/Rationale 자연어 선택 기준

## Implementation Notes

본 SPEC은 draft 작성 → 실행 → completed 처리를 단일 세션에서 수행했다. 편집 패턴은 단순: 각 파일에서 `### <Section Header>` 부터 다음 `### <Next Section>` 바로 전까지의 블록 제거. workflow-modes.md만 예외적으로 4개 개별 문장 수정(섹션 삭제 없음).

편집 후 `grep -rn 'Skill(\"batch\")\|Skill(\"simplify\")' .claude/` 결과 0건 확인, 전체 Go test green, `make build` 성공 — AC-1 ~ AC-10 전부 통과.

향후 `/simplify` 강제 인프라 필요성이 재논의되면 별도 SPEC(예: `SPEC-SIMPLIFY-ENFORCEMENT-001`)으로 독립 추적한다.

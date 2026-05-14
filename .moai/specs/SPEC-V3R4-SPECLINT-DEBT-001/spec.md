---
id: SPEC-V3R4-SPECLINT-DEBT-001
version: "0.1.0"
status: in-progress
created: 2026-05-15
updated: 2026-05-15
author: manager-spec
priority: P0
tags: "spec-lint, debt, ci-green, frontmatter, dependency-cycle, coverage, status-sync, v3r4, foundation"
issue_number: null
title: SPEC Lint Debt 일괄 해소 (P0 ERROR 66건 + P1 WARNING 140건)
phase: "v3.0.0 R4 — Foundation Cleanup"
module: ".moai/specs, .github/workflows/spec-lint.yml, scripts/spec-status-sync.go"
dependencies: []
related_specs:
  - SPEC-V3R4-HARNESS-001
  - SPEC-V3R4-HARNESS-002
breaking: false
bc_id: []
lifecycle: spec-anchored
related_theme: "spec-lint CI Green + 메타데이터 위생"
target_release: v3.0.0-rc1
---

# SPEC-V3R4-SPECLINT-DEBT-001 — SPEC Lint Debt 일괄 해소

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.1   | 2026-05-15 | manager-develop run | run-phase 진입 시 실측 baseline 차이 반영: ERROR 66 동일하되 카테고리 재분포 (+2 FrontmatterInvalid ID format violation + 4 ParseFailure 신규 카테고리, plan 시점에 미발견). Expanded scope 결정 (user AskUserQuestion 1) — 모두 T-SLD-001 frontmatter 카테고리로 흡수. AC-SLD-007 target `≤ 5` → `≤ 55` 재조정 (47 author-intent preservation + 4 terminal state preservation). 5 categorical commits 채택 (OQ4 lock-in). 상세 발견 사항: `status-residuals.md`. plan-auditor PASS 0.92 (Phase 0.5 cache MISS → 재실행). |
| 0.1.0   | 2026-05-15 | manager-spec | 초기 draft. PR #913 머지 후 `origin/main` (commit `2e27c14f8`) 기준 `moai spec lint --strict` 출력의 ERROR 66건 + WARNING 140건을 단일 SPEC으로 일괄 해소. 6개 ERROR 카테고리(FrontmatterInvalid, MissingExclusions, MissingDependency, DependencyCycle, ModalityMalformed, CoverageIncomplete) + 2개 WARNING 카테고리(StatusGitConsistency, OrphanBCID) 모두 포괄. 목표: `moai spec lint --strict` exit 0 + spec-lint CI job GREEN. SPEC 콘텐츠의 의미적 재작성은 금지 — 메타데이터/frontmatter/AC reference만 수정. BODP 평가: signals A=¬ B=¬ C=¬ → main @ origin/main (plan-in-main 원칙 PR #822 준수). |

---

## 1. Goal

`moai spec lint --strict` 명령이 origin/main (commit `2e27c14f8`) 기준 ERROR 66건 + WARNING 140건을 보고하는 현재 상태를, exit code 0 + WARNING 0건(혹은 정당화된 잔존 항목만)으로 정리하여 spec-lint CI job을 GREEN으로 전환한다. SPEC 디렉토리 188개 전반에 누적된 메타데이터 부채(frontmatter 필드 누락, depends_on 무효 참조, AC↔REQ 미커버리지, EARS modality 형식 오류, 의존성 순환, status-git 불일치, breaking=false인데 bc_id 잔존)를 단일 SPEC + 단일 PR로 묶어 해소한다. 본 SPEC은 **SPEC 콘텐츠의 의미적 재작성을 일체 수행하지 않는다** — REQ/AC 본문은 보존하고 메타데이터/참조/frontmatter만 수정한다.

### 1.1 배경

- 2026-05-15 origin/main commit `2e27c14f8`에서 `moai spec lint --strict`를 실행한 결과, 6개 ERROR 카테고리에서 총 66건의 CI-blocking 에러가 검출되었고 2개 WARNING 카테고리에서 총 140건의 비차단 경고가 검출되었다. ERROR 분포는 다음과 같다:
  - FrontmatterInvalid (11건): SPEC-V3R2-RT-001 (7 필드 누락), SPEC-V3R4-HARNESS-002 (4 필드 누락)
  - MissingExclusions (1건): SPEC-V3R4-HARNESS-002 의 Out of Scope 0 항목
  - MissingDependency (2건): SPEC-V3R2-RT-005 → SPEC-V3R2-SCH-001 (부재 SPEC), SPEC-V3R3-COV-001 → SPEC-V3R3-ARCH-003 (부재 SPEC)
  - DependencyCycle (1건): SPEC-V3R2-RT-005 ↔ SPEC-V3R2-RT-004 (양방향 순환)
  - ModalityMalformed (1건): SPEC-V3R2-SPC-003 line 95 REQ-SPC-003-041 "SHALL" 키워드 누락
  - CoverageIncomplete (최소 25건, 최대 50건+ 가능 — 정확한 baseline은 Wave 1 T-SLD-006에서 측정): SPC-001/SPC-002/SPC-003/SPC-004 등 여러 SPEC에서 REQ가 어떤 AC에서도 참조되지 않음
- WARNING 분포는 다음과 같다:
  - StatusGitConsistency (~140건): frontmatter `status`와 git 이력에서 추론한 lifecycle status 불일치. 대부분 frontmatter=draft/in_review/in_progress 이고 git=implemented 또는 frontmatter=implemented 이고 git=completed 패턴.
  - OrphanBCID (1건): SPEC-V3R3-ARCH-007 `breaking: false` 인데 `bc_id` 비어있지 않음
- spec-lint CI workflow는 `--strict` 플래그로 실행되므로 ERROR 1건 이상 검출 시 CI fail이 발생한다. 본 SPEC이 해소되기 전까지 PR마다 spec-lint job이 적색으로 표시되어 다른 SPEC들의 정상 머지가 차단되는 위험이 존재한다.
- 본 SPEC은 SPEC-V3R4-HARNESS-001 (foundation) 머지 이후 등장한 첫 번째 메타데이터 위생 SPEC이며, V3R4 series의 lint debt 베이스라인을 0으로 리셋하는 cleanup SPEC으로 자리잡는다.

### 1.2 사용자 결정 사항

다음 결정은 본 SPEC plan-phase에서 잠정 채택된다. run-phase에서 SPEC 분석 결과에 따라 조정될 수 있다:

1. **MissingDependency 해소 전략**: SCH-001, ARCH-003은 모두 부재 SPEC이다. 우선 전략은 **(a) depends_on 항목 제거** (호환성 호환 변경 없음). 추가 분석 결과 해당 SPEC이 다른 기존 SPEC의 별칭/리네임인 것으로 판명되면 (b) repoint 적용. stub SPEC 신규 작성 (c)은 본 SPEC 범위 밖.
2. **DependencyCycle 해소 전략**: RT-005 ↔ RT-004 순환에서 **RT-004 → RT-005 백엣지를 제거**한다 (RT cluster는 일반적으로 sequential numbering이며 후행 SPEC이 선행 SPEC에 의존하는 정방향만 유지).
3. **CoverageIncomplete 해소 전략**: case-by-case 분석. 각 미참조 REQ에 대해 (a) 매칭 AC 추가 (REQ를 검증할 AC 가 있어야 했으나 누락됨) 또는 (b) REQ 자체 제거 (orphan REQ는 SPEC 작성 시 잔존한 placeholder). run-phase에서 manager-develop이 SPEC 본문을 읽고 의도를 추론한다.
4. **StatusGitConsistency 해소 전략**: convention은 `implemented` = 코드 머지 완료, `completed` = run + sync + docs 전체 라이프사이클 종료. 자동화 도구(Go 스크립트)로 git log + PR 머지 상태를 분석하여 frontmatter `status`를 일괄 정정. 모호한 케이스(N건)는 수동 검토.
5. **ModalityMalformed 해소 전략**: REQ-SPC-003-041 본문에 "SHALL" 키워드를 EARS 표준에 맞춰 삽입. 기존 의도("WHERE ... is specified, the default ... is selected") 보존.

### 1.3 Non-Goals (Out of Scope)

본 SPEC은 메타데이터 정정 SPEC이다. 다음 작업은 **명시적으로 범위 밖**이다:

- **SPEC 콘텐츠의 의미적 재작성**: 기존 SPEC의 REQ 본문, plan.md 내용, acceptance.md 시나리오를 의미적으로 재작성하지 않는다. frontmatter / depends_on / AC reference / EARS modality 키워드만 수정한다.
- **부재 SPEC 신규 작성**: SCH-001, ARCH-003이 부재한다는 사실은 알지만, 이를 신규 SPEC으로 작성하지 않는다. depends_on 목록에서 제거 또는 sentinel SPEC 작성 중 택일 → 본 SPEC에서는 **제거**로 결정 (1.2 §1).
- **CI workflow 신규 작성**: `.github/workflows/spec-lint.yml`은 이미 존재한다 (PR #913 산출물 추정). 본 SPEC은 신규 워크플로우 작성이 아니라 기존 워크플로우의 GREEN화이다.
- **`moai spec lint` 도구 자체의 수정**: lint rule 추가/제거/변경은 별도 SPEC. 본 SPEC은 현 규칙 하에서 통과시키는 것을 목표로 한다.
- **WARNING 0건 강제**: StatusGitConsistency 자동 변환 후 잔존 모호 케이스는 ERROR로 승격되지 않는 한 잔존 허용. spec-lint `--strict` 모드에서 WARNING은 exit code 0을 막지 않는다.
- **SPEC 디렉토리 구조 재편**: 다른 SPEC의 spec.md / plan.md / acceptance.md 파일 구성 자체는 손대지 않는다.

**In-scope clarification (plan-auditor §6 response)**: `scripts/spec-status-sync.go` (Wave 2 자동화 도구) 신규 작성은 In Scope이다 — StatusGitConsistency batch 처리 + rollback 지원의 단일 산출물로서 본 SPEC 의 보조 도구로 분류한다. 별도 SPEC 으로 분리하지 않는 이유: T-SLD-007 의 핵심 산출물이며, 본 SPEC 종료와 함께 도구 lifecycle 도 종료 (one-shot tool, 향후 재실행은 follow-up SPEC 으로).

---

## 2. Scope

### 2.1 In Scope

**A. ERROR 카테고리 해소 (P0, CI-blocking)**:

- A1. FrontmatterInvalid 11건 — 2개 SPEC의 누락 frontmatter 필드를 채운다 (SPEC-V3R2-RT-001: title/created/updated/phase/module/lifecycle/tags 7건; SPEC-V3R4-HARNESS-002: title/created/updated/tags 4건). 값은 git log + 인접 SPEC 컨벤션에서 추론.
- A2. MissingExclusions 1건 — SPEC-V3R4-HARNESS-002 `## Out of Scope` 섹션에 최소 1개 항목 추가. 항목 내용은 SPEC 본문 검토 후 작성.
- A3. MissingDependency 2건 — SPEC-V3R2-RT-005, SPEC-V3R3-COV-001 의 depends_on에서 SCH-001, ARCH-003 참조를 제거.
- A4. DependencyCycle 1건 — SPEC-V3R2-RT-004 의 depends_on에서 SPEC-V3R2-RT-005 참조를 제거 (역방향 정상화).
- A5. ModalityMalformed 1건 — SPEC-V3R2-SPC-003 line 95 REQ-SPC-003-041 본문에 "SHALL" 키워드 삽입.
- A6. CoverageIncomplete N건 — 각 SPEC에서 미참조 REQ 마다 매칭 AC 추가 또는 REQ 제거.

**B. WARNING 카테고리 해소 (P1, 비차단이나 위생 개선)**:

- B1. StatusGitConsistency ~140건 — Go 자동화 스크립트로 git PR 머지 상태 기반 status 일괄 정정. 모호 케이스 수동 검토.
- B2. OrphanBCID 1건 — SPEC-V3R3-ARCH-007 의 frontmatter에서 `bc_id` 필드 제거 또는 빈 값 명시.

**C. 검증 단계**:

- C1. 로컬 `moai spec lint --strict` exit 0 확인.
- C2. PR 생성 후 spec-lint CI job GREEN 확인.
- C3. plan-auditor 재실행하여 본 SPEC 자체가 lint pass 인지 확인.

### 2.2 Out of Scope

§1.3 참조.

---

## 3. Requirements

### 3.1 EARS Requirements

#### REQ-SLD-001 (FrontmatterInvalid 해소, Ubiquitous)

The MoAI-ADK SPEC corpus SHALL contain valid YAML frontmatter in every `spec.md` file under `.moai/specs/SPEC-*/`. Every frontmatter SHALL include all 7 mandatory fields: `title`, `created`, `updated`, `phase`, `module`, `lifecycle`, `tags`. `moai spec lint --strict` SHALL emit zero `FrontmatterInvalid` errors after this SPEC is implemented.

#### REQ-SLD-002 (MissingExclusions 해소, Ubiquitous)

Every `spec.md` SHALL contain a `## Out of Scope` section with at least 1 explicit item. `moai spec lint --strict` SHALL emit zero `MissingExclusions` errors after this SPEC is implemented.

#### REQ-SLD-003 (MissingDependency 해소, Event-driven)

WHEN `moai spec lint --strict` traverses any SPEC's `depends_on` list, IF any referenced SPEC ID does not resolve to an existing `.moai/specs/SPEC-*/` directory, THEN the lint SHALL emit `MissingDependency`. After this SPEC is implemented, the lint SHALL emit zero `MissingDependency` errors. The SPEC-V3R2-RT-005 depends_on list SHALL NOT contain `SPEC-V3R2-SCH-001`. The SPEC-V3R3-COV-001 depends_on list SHALL NOT contain `SPEC-V3R3-ARCH-003`.

#### REQ-SLD-004 (DependencyCycle 해소, Event-driven)

WHEN `moai spec lint --strict` builds the SPEC dependency graph, IF any cycle is detected, THEN the lint SHALL emit `DependencyCycle`. After this SPEC is implemented, the SPEC-V3R2-RT-004 `depends_on` list SHALL NOT contain `SPEC-V3R2-RT-005`, breaking the RT-004 ↔ RT-005 cycle while preserving the forward edge RT-005 → RT-004.

#### REQ-SLD-005 (ModalityMalformed 해소, Ubiquitous)

Every EARS-format requirement SHALL contain at least one of the modality keywords `SHALL`, `SHALL NOT`, `MUST`, `MUST NOT`, `WILL`, `WILL NOT` in its body. The SPEC-V3R2-SPC-003 line 95 REQ-SPC-003-041 SHALL be revised to include a `SHALL` modality verb while preserving the original EARS pattern intent (WHERE `moai spec lint --format table` is specified, the default human-readable output is explicitly selected).

#### REQ-SLD-006 (CoverageIncomplete 해소, Ubiquitous)

Every REQ-ID declared in any SPEC's `## Requirements` section SHALL be referenced by at least one acceptance criterion in the same SPEC's `acceptance.md`. WHEN a REQ has no matching AC, the maintainer SHALL choose to (a) add a corresponding AC, or (b) remove the orphan REQ. After this SPEC is implemented, `moai spec lint --strict` SHALL emit zero `CoverageIncomplete` errors.

#### REQ-SLD-007 (StatusGitConsistency 일괄 정정, State-driven)

WHILE the SPEC corpus contains historical SPECs whose frontmatter `status` predates the lifecycle convention (`implemented` = code merged, `completed` = full run+sync+docs lifecycle), the orchestrator SHALL run a one-shot automation tool (Go or Python equivalent) that reads each SPEC's git PR merge history and proposes a status correction. The orchestrator SHALL apply the corrections in a single batch commit. After this SPEC is implemented, `moai spec lint --strict` SHALL emit at most N residual `StatusGitConsistency` warnings (target: N ≤ 55 per revised AC-SLD-007 v0.1.1, accounting for 47 `completed → implemented` author-intent preservation + 4 terminal-state preservation; ambiguous residuals SHALL be documented in `status-residuals.md` and OPTIONALLY suppressed via `lint.skip` per REQ-SPC-003-040 mechanism).

#### REQ-SLD-008 (OrphanBCID 해소, Event-driven)

WHEN any SPEC frontmatter declares `breaking: false`, THEN the same frontmatter SHALL NOT declare a non-empty `bc_id` field. After this SPEC is implemented, SPEC-V3R3-ARCH-007 frontmatter SHALL satisfy this constraint (either `bc_id` removed entirely or set to empty list `[]`).

#### REQ-SLD-009 (Lint CI Green 게이트, Ubiquitous)

The `.github/workflows/spec-lint.yml` CI job SHALL exit with status 0 on the PR delivering this SPEC's run-phase. The PR SHALL NOT be merged into `main` until the spec-lint job is GREEN. The job SHALL invoke `moai spec lint --strict` and propagate its exit code.

#### REQ-SLD-010 (Self-coverage 보장, Ubiquitous)

This SPEC SHALL conform to its own constraints: every REQ-SLD-NNN above SHALL be referenced by at least one AC in `acceptance.md`. The frontmatter of this SPEC SHALL include all 7 mandatory fields. The `## Out of Scope` section (§1.3) SHALL contain at least 1 explicit item. `moai spec lint --strict` SHALL NOT emit any error for SPEC-V3R4-SPECLINT-DEBT-001 itself.

---

## 4. Approach Summary

### 4.1 Wave 분할

본 SPEC의 run-phase는 3개 Wave로 구성된다:

- **Wave 1 (P0 ERROR 해소)**: A1 → A3 → A4 → A5 → A2 → A6 순서. 복잡도 오름차순 (단순 frontmatter 보충 → 의존성 제거 → 순환 끊기 → modality 키워드 삽입 → exclusions 추가 → coverage 분석).
- **Wave 2 (P1 WARNING 해소)**: B1 (자동화 스크립트 + 수동 잔존) + B2 (단일 케이스).
- **Wave 3 (검증)**: C1 + C2 + C3. `moai spec lint --strict` exit 0 확인 + CI GREEN 확인 + plan-auditor self-review.

### 4.2 도구

- **frontmatter 보충 (A1)**: 인접 SPEC frontmatter 컨벤션 참조 + git log 분석.
- **의존성 제거 (A3, A4)**: 단순 텍스트 편집 (Edit tool).
- **modality 키워드 삽입 (A5)**: 단일 라인 Edit.
- **exclusions 추가 (A2)**: SPEC-V3R4-HARNESS-002 본문 검토 후 Out of Scope 항목 1+ 작성.
- **coverage 분석 (A6)**: 각 SPEC의 spec.md + acceptance.md 동시 검토. REQ↔AC mapping 표 생성 후 누락 reference 보충.
- **status 일괄 정정 (B1)**: Go 자동화 스크립트 (`scripts/spec-status-sync.go`) — `gh pr list --search "head:<branch> is:merged"` + git log 분석으로 status 추론.

### 4.3 Branch + PR 전략

- BODP signals: A=¬ (코드 의존성 없음 — 메타데이터만), B=¬ (현재 작업 트리에 본 SPEC 미존재), C=¬ (현재 브랜치 open PR 없음) → main @ origin/main 채택.
- Plan PR: `plan/V3R4-SPECLINT-DEBT-001` → main (squash, plan-in-main doctrine PR #822).
- Run PR: `feat/SPEC-V3R4-SPECLINT-DEBT-001` → main (squash). Wave 1+2+3 단일 PR (대량 메타데이터 변경이나 의미적 변경 부재).
- Sync PR: 필요 시 `chore/SPEC-V3R4-SPECLINT-DEBT-001-sync` → main (squash).

---

## 5. Success Criteria

§3 REQ들이 `acceptance.md` 의 AC에서 검증된다. 본 섹션은 SPEC 작성자가 기대하는 high-level 성과 지표:

- `moai spec lint --strict` exit code 0 on origin/main HEAD after run PR merge.
- spec-lint CI job GREEN on the run PR.
- ERROR count: 66 → 0.
- WARNING count: 141 → ≤ 55 (revised v0.1.1: residual = 47 `completed → implemented` author-intent preservation + 4 terminal state preservation + ≤ 4 buffer; 상세 `status-residuals.md`).
- Zero SPEC 본문 의미적 변경 (diff 검토에서 REQ/plan 본문 라인 변경이 메타데이터 라인보다 압도적으로 적어야 함).
- 본 SPEC 자체가 `moai spec lint --strict` 에서 ERROR 0건.

---

## 6. Risks

- **R1**: CoverageIncomplete 케이스 수가 실제로는 25건이 아니라 50건+ 일 수 있다 (raw output 일부만 노출). run-phase에서 재실행하여 정확한 수치를 측정 후 Wave 1 A6 task를 sub-divide.
- **R2**: SCH-001, ARCH-003가 사실 다른 SPEC의 별칭일 가능성. run-phase에서 git log + SPEC 본문 grep으로 확인 후 §1.2 §1 결정을 (a) 제거에서 (b) repoint로 변경 가능.
- **R3**: RT-004 ↔ RT-005 순환에서 어느 방향이 "원래 의도"인지 본 SPEC 작성 시점에 불확실. run-phase에서 두 SPEC 본문을 읽고 결정. 잘못 끊으면 sequential RT cluster의 의미 파괴.
- **R4**: StatusGitConsistency 자동화 스크립트가 false-positive 다수 생성. 수동 검토 부담이 예상보다 클 수 있음. 임계치(N ≤ 5) 조정 가능.
- **R5**: 본 SPEC 작업 도중 다른 PR이 새 SPEC을 추가하여 lint debt가 증식. 미티게이션: run PR을 최대한 빠르게 단일 commit으로 마감.

---

## 7. Out of Scope

§1.3 의 5개 항목이 명시적 Non-Goals이다. 본 섹션은 reviewer 가 빠르게 확인할 수 있도록 핵심 3건을 재게재한다:

1. **SPEC 콘텐츠의 의미적 재작성**: 본 SPEC은 frontmatter / 메타데이터 / AC reference 만 수정한다. SPEC의 REQ 본문, plan.md 본문, acceptance.md 시나리오 본문은 보존한다. (예외: A5 ModalityMalformed는 단일 키워드 삽입으로 의미가 변하지 않는 한 허용.)
2. **부재 SPEC 신규 작성**: SCH-001, ARCH-003 같은 부재 SPEC을 신규 작성하지 않는다. depends_on 목록에서 제거한다.
3. **CI workflow 신규 작성 또는 lint rule 수정**: 기존 spec-lint workflow는 그대로 두고, lint rule도 수정하지 않는다. 현 규칙 하에서 통과시키는 것이 목표.

---

## 8. References

- PR #913 (`origin/main` commit `2e27c14f8`): spec-lint CI 도입 직후 첫 lint-debt baseline.
- `.claude/rules/moai/workflow/spec-workflow.md`: SPEC phase contract, EARS format.
- `.claude/rules/moai/workflow/mx-tag-protocol.md`: @MX 태그 규칙 (본 SPEC plan.md 에서 활용).
- `.moai/specs/SPEC-V3R4-HARNESS-001/spec.md`: V3R4 series frontmatter schema reference.
- CLAUDE.local.md §18 (Enhanced GitHub Flow): branch/PR strategy.
- CLAUDE.local.md §18.12 (BODP): branch origin decision.

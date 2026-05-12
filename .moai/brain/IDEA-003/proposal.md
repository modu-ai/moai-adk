# IDEA-003 Proposal — Catalog Slimming + Harness-First Distribution

> Generated: 2026-05-12 · Phase 6 of `/moai brain` workflow
> Primary input: ideation.md · Market grounding: research.md

## Executive Summary

moai-adk의 skill/agent 카탈로그를 3-tier 구조 (Core / Optional Packs / Harness-Generated) 로 재편하고, `moai update`를 cruft-style drift detection 기반 안전 동기화로 재설계한다. 기존 사용자의 손실 0을 절대 목표로 한다.

핵심 산출물:
- **코어 카탈로그 슬림화** (36 skills → 18 skills, 29 agents → 15 agents, 약 -50%)
- **옵션 팩 시스템** (`moai pack add <name>`)
- **harness auto-trigger** (`/moai project` 인터뷰 끝에서 opt-out 제안)
- **안전 동기화** (`moai update` 가 drift 감지 + 백업 + 사용자 확인 + 자동 rollback)

## Capability Scope (NOT Implementation)

> [HARD] 이 섹션은 능력 (capability) 만 기술. 구현 세부 (Go 코드, 라이브러리 선택 등) 는 후속 SPEC에서 결정. (REQ-BRAIN-011 준수)

### Capability 1: 카탈로그를 명시적 3-tier로 분류한다

- 모든 skill과 agent는 `core` / `optional-pack:<name>` / `harness-generated` 중 정확히 하나의 tier에 속한다
- tier 정보는 카탈로그 manifest 파일에 선언적으로 기록된다 (`internal/template/catalog.yaml` 형식 등 후속 결정)
- `moai init`은 기본적으로 core만 deploy
- `moai update`는 manifest를 기준으로 사용자 환경과의 drift를 계산

### Capability 2: 옵션 팩을 사용자 명령으로 추가/제거할 수 있다

- 신규 명령 `moai pack {add|remove|list|available}`
- 팩은 "skills + agents + rules + commands" 의 묶음 (Anthropic plugin과 동등한 단위)
- 초기 팩 후보: backend, frontend, mobile, chrome-extension, auth, deployment, design, devops, testing
- 팩 install 시 의존 관계 자동 해결 (예: design 팩은 frontend 팩을 권장)

### Capability 3: `/moai project`가 harness 활성화를 자동 제안한다

- 인터뷰 끝에서 AskUserQuestion으로 harness opt-out 선택지 제시 (3 옵션: "예 권장" / "코어만" / "팩만 install")
- 기본값은 "예". 거부 시 코어만 또는 사용자가 선택한 팩만 install
- harness 실행 결과는 `.claude/skills/my-harness-*` 와 `.claude/agents/my-harness/` 에 격리

### Capability 4: `moai update`가 catalog drift를 안전하게 동기화한다

- 실행 시 자동 snapshot (`.moai/cache/catalog-snapshot-<timestamp>/`)
- Drift 카테고리:
  - **Core 자산의 사용자 직접 수정**: hash 비교로 감지
  - **사용자 추가 자산** (my-harness-*, custom skill): 절대 보존
  - **코어에서 제거된 자산**: 사용처 grep 후 사용 중이면 경고
  - **신규 코어 자산**: 자동 추가
  - **옵션 팩의 upstream 변경**: install된 팩만 동기화
- 모든 변경은 AskUserQuestion으로 사용자 승인
- 적용 실패 시 snapshot으로 자동 rollback
- Audit log: `.moai/cache/catalog-sync-<timestamp>.log`

### Capability 5: 카탈로그 상태를 진단할 수 있다

- 신규 명령 `moai doctor catalog` (또는 기존 `moai doctor` 확장)
- 출력: tier별 install 자산 수, context budget 사용량, idle skill (6개월+ 미사용) 목록, drift 여부

### Capability 6: 마이그레이션 문서를 4개국어로 제공한다

- CHANGELOG에 breaking-vs-non-breaking 명시
- docs-site에 마이그레이션 가이드 (ko/en/ja/zh 4개국어, §17 docs-site 규칙 준수)
- `moai update`의 첫 실행 시 마이그레이션 안내 출력

## SPEC Decomposition Candidates

> [HARD] 후속 `/moai plan` 워크플로우의 입력. 각 항목은 단일 SPEC ID로 진행 가능한 응집된 scope.

- SPEC-V3R4-CATALOG-001: 3-tier manifest 스키마 정의 + 현재 36 skills / 29 agents의 tier 분류 lock-in (코어/옵션-팩별/harness-gen). 산출: `internal/template/catalog.yaml`, 분류 테이블, lang_boundary_audit_test 와 동급의 `catalog_tier_audit_test.go`
- SPEC-V3R4-CATALOG-002: 카탈로그 디렉토리 재배치 — `internal/template/templates/.claude/skills/` 와 `agents/`를 코어/팩별 디렉토리로 분리 (`internal/template/packs/<pack>/`). 기존 build 파이프라인 (`make build`) 호환 유지. Go embed 패스 + deploy 로직 분기 반영
- SPEC-V3R4-CATALOG-003: `moai pack` CLI 명령 (add/remove/list/available) — manifest 기반 install + 의존성 해결 + 충돌 감지. 단위 테스트 + integration 테스트 (`/tmp/test-project` 환경)
- SPEC-V3R4-CATALOG-004: `moai update --catalog-sync` 안전 동기화 — cruft-style drift detection, 3-way merge, snapshot 백업, AskUserQuestion 흐름, rollback, audit log. **가장 위험**. evaluator-active strict mode + 광범위 시나리오 통합 테스트 필수
- SPEC-V3R4-CATALOG-005: `/moai project` 인터뷰 확장 — harness opt-out AskUserQuestion 추가, 거부 시 팩 install fallback. 기존 `manager-project` agent + project workflow skill 업데이트
- SPEC-V3R4-CATALOG-006: `moai doctor catalog` 진단 명령 — tier별 카운트 + context budget 표시 + idle skill 감지 (사용 빈도 추적 hook 검토). README에 사용법 추가
- SPEC-V3R4-CATALOG-007: 마이그레이션 docs + CHANGELOG + 4개국어 docs-site sync — `adk.mo.ai.kr/migration/v3r4-catalog/` (4개 locale) + 영문/한국어 CHANGELOG + `moai update` 첫 실행 시 인라인 안내

### 권장 실행 순서

```
Wave 1 (Foundation)
  ├─ SPEC-V3R4-CATALOG-001 (manifest 스키마)
  └─ SPEC-V3R4-CATALOG-002 (디렉토리 재배치)
         │
Wave 2 (Distribution)
  ├─ SPEC-V3R4-CATALOG-003 (moai pack)
  └─ SPEC-V3R4-CATALOG-005 (/moai project 인터뷰)
         │
Wave 3 (Safety — 가장 위험)
  └─ SPEC-V3R4-CATALOG-004 (moai update 안전 동기화)
         │
Wave 4 (Polish)
  ├─ SPEC-V3R4-CATALOG-006 (moai doctor catalog)
  └─ SPEC-V3R4-CATALOG-007 (마이그레이션 docs 4개국어)
```

## Acceptance — Brain Phase Exit

이 proposal이 후속 `/moai plan` 워크플로우 입력으로 전달되어야 할 조건:

- [x] Lean Canvas 9 블록 작성
- [x] 5가지 비판 + 대응 (Phase 5)
- [x] First Principles 분해
- [x] 7개 SPEC 후보가 `SPEC-V3R4-CATALOG-NNN` 그래마 준수
- [x] Tech-stack 가정 없음 (capability만 기술)
- [x] 가장 위험한 영역 식별 (SPEC-V3R4-CATALOG-004)

## Open Questions for `/moai plan` Phase

다음 결정은 SPEC 작성 시 (`/moai plan SPEC-V3R4-CATALOG-001`) 다룬다:

1. catalog.yaml의 정확한 스키마 (tier, dependencies, conflicts, recommended_with)
2. 팩 dependency resolution 알고리즘 (linear vs DAG)
3. moai update의 3-way merge 알고리즘 선택 (git merge-file 호출 vs 자체 구현)
4. opt-in과 opt-out 비율 측정 telemetry 여부 (privacy 정책 충돌 검토)
5. Anthropic marketplace publishing 여부 (별도 SPEC으로 분리 가능)
6. 기존 사용자의 `manager-brain.md`, `expert-mobile.md`, `expert-devops.md` 등을 어느 팩에 넣을지 최종 매핑

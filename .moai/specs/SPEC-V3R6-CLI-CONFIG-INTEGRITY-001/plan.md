---
id: SPEC-V3R6-CLI-CONFIG-INTEGRITY-001
title: "CLI config-system integrity / mental-model alignment (P0 fixes)"
version: "0.1.0"
status: draft
created: 2026-06-23
updated: 2026-06-24
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
tier: "M"
tags: "cli, config, profile, update, init-wizard, mental-model, p0"
---

# Implementation Plan — SPEC-V3R6-CLI-CONFIG-INTEGRITY-001

## §A. Context

본 SPEC은 4개 P0 결함(F1, F2, F3, P0-4)을 독립적 마일스톤으로 분해. 각 마일스톤은 서로 의존하지 않으므로 병렬 실행 가능하나, run-phase에서는 manager-develop가 순차 M1→M4로 처리하는 것을 권장(충돌 회피 및 review 단순화).

## §B. Known Issues (검증 완료)

| ID | 결함 | 증거(file:line) | 검증 상태 |
|----|------|-----------------|-----------|
| F1 | `update -c` help 모호 | `update.go:88,111-122,128,166-170` | verified (직독) |
| F2 | `expandModelString` no-op + scatter 리터럴 | `launcher.go:704-706`, `profile_setup.go:70`, `settings/schema.go:140`, `launcher_test.go:601` | verified (직독 + grep) |
| F3 | "acceptEdits" 정규화 투명성 부재 | `profile_setup.go:592-595`, `launcher.go:645-698` | verified (직독) |
| P0-4 | db.yaml `_TBD_` 3종 | local + template source 동일 | verified (cat + grep) |

## §C. Pre-flight

- [ ] `make build` 성능 확인 (template 수정 후 embedded.go 재생성 필요)
- [ ] `go test ./internal/cli/... ./internal/settings/... ./internal/template/...` baseline 통과
- [ ] `golangci-lint run` baseline clean
- [ ] worktree 권장(L1 isolation): 본 SPEC은 4개 마일스톤이 `internal/cli/`와 `internal/template/templates/`에 교차 수정을 가하므로 L1 worktree가 안전

## §D. Constraints

- Template-First(CLAUDE.local.md §2): db.yaml은 template source 먼저 수정, 이후 `moai update`로 local 전파.
- Hardcoding prevention(§14): model id 리터럴은 단일 중앙 테이블(`internal/template/model_policy.go` 인근).
- Template neutrality(§25): template source에 SPEC ID/REQ/commit SHA 누출 금지.
- 행위 변경 최소: F1은 help/doc만, F3은 (a) 확인 라인 추가 우선.

## §E. Self-Verification

- [ ] 각 마일스톤 완료 후 `go test ./internal/cli/...` 통과
- [ ] F2: 중앙 테이블 1소스로 수렴. **Canonical grep (D3 fix — spec.md §G / acceptance.md AC-003·AC-010과 동일 pattern)**:

  ```bash
  grep -rnE '"(opus|sonnet|haiku|opusplan)(\[1m\])?"' internal/cli/ internal/settings/
  ```

  관측 live match 수(2026-06-24 직접 실행): **50건 pre-refactor**. post-refactor 기대: 비-테스트·비-`model_policy` Go 파일에서 0건. 이 ERE는 닫는 따옴표 `"`를 요구하므로 `"claude-opus-4-7"`의 `"opus` prefix를 오매칭하지 않음(D3 escaped-pipe vacuity trap — 구 pattern 36/84/124건 분산의 근원 — 회피).
- [ ] P0-4: `grep -rn '_TBD_' .moai/config/sections/db.yaml internal/template/templates/.moai/config/sections/db.yaml` → 0건
- [ ] F1: README `update -c` 설명에 "init wizard" + "no template sync" 키워드 동시 존재
- [ ] `make build` 후 embedded template에 변경사항 반영 확인

## §F. Milestones

### M1 — F1: `update -c` help/doc/README 명확화

**접근**: help string 1줄 + doc comment 확장 + README 섹션 추가. 코드 행위 변경 없음.

**파일 수정 목표**:
- `internal/cli/update.go:88` — flag help string 개선
- `internal/cli/update.go:111-122` — doc comment 확장(-c vs bare update 구분 명시)
- `README.md` / `README.ko.md` — `moai update` 섹션에 `-c` vs bare update 구분 추가

**완료 조건**:
- REQ-CCI-001, REQ-CCI-002 충족
- `moai update --help` 출력에 "no template sync" 의미 포함
- README grep으로 `-c` 설명에서 "init wizard" + "no template" 확인

**우선순위**: High (사용자 혼란 최상위 원인)

### M2 — F2: 중앙 alias→canonical-id 테이블 + `expandModelString` 확장

**접근**: `internal/template/model_policy.go` 인근에 단일 테이블 추가. `expandModelString()`이 테이블 참조. scatter 리터럴 3소스(`profile_setup.go:70`, `settings/schema.go:140`, `launcher_test.go:601`)를 테이블 참조로 교체.

**파일 수정 목표**:
- `internal/template/model_policy.go`(또는 동일 패키지 신규 파일) — `ModelAliasTable` 정의
- `internal/cli/launcher.go:704-706` — `expandModelString` 테이블 참조 구현
- `internal/cli/profile_setup.go:67-87` — `normalizeModel`이 테이블 참조
- `internal/settings/schema.go:140` — `ModelOptionValues()`가 테이블 참조
- `internal/cli/launcher_test.go:601` — no-op 단언을 canonical-id 해석 단언으로 갱신 + table-driven 케이스 확장

**완료 조건**:
- REQ-CCI-003, REQ-CCI-004, REQ-CCI-005, REQ-CCI-010 충족
- scatter 리터럴 grep 결과 0건(중앙 테이블만 존재)
- `expandModelString("opus")` → canonical id 반환
- `expandModelString("opus[1m]")` → `[1m]` 접미사 보존

**우선순위**: High (잠재적 silent-failure + hardcoding 부채)

**리스크**: CC 런타임 버전별 model id 변경(예: opus-4-8 → opus-4-9). 완화: 테이블을 단일 SSOT로 유지하고 릴리스 노트에 갱신 절차 명시.

### M3 — F3: defaultMode "acceptEdits" wizard 확인 라인

**접근**: wizard에 확인 메시지 추가(옵션 a). `syncPermissionModeToSettingsLocal` 주석에 cross-ref 추가.

**파일 수정 목표**:
- `internal/cli/profile_setup.go` — wizard 단계 종료 후 "acceptEdits는 project default — settings.local.json에 미기록" 확인 라인 추가(약 line 595 이후)
- `internal/cli/launcher.go:645-652` — 함수 주석에 REQ-CCI-006 cross-ref 추가

**완료 조건**:
- REQ-CCI-006, REQ-CCI-007 충족
- wizard 실행 시 "acceptEdits" 선택 확인 메시지 출력
- `launcher.go:645` 주석에 "wizard에서 사용자에게 미기록 사실 고지됨(REQ-CCI-006)" 명시

**우선순위**: Medium (사용자 오인 방지)

### M4 — P0-4: db.yaml `_TBD_` placeholder 제거

**접근**: template source에서 `_TBD_`를 empty string 또는 `null`로 교체. `/moai db init` interview 계약 유지. 이후 `moai update`로 local 전파.

**파일 수정 목표**:
- `internal/template/templates/.moai/config/sections/db.yaml` — `engine`, `orm`, `migration_tool` 값을 `_TBD_`에서 `""` 또는 `null`로 변경
- `make build` — embedded template 재생성
- local `.moai/config/sections/db.yaml` — `moai update` 실행으로 전파(또는 수동 동기화)

**완료 조건**:
- REQ-CCI-008, REQ-CCI-009 충족
- template source와 local 양쪽 모두 `_TBD_` 0건
- `/moai db init` interview가 여전히 빈 값을 채우는 정상 동작 확인

**우선순위**: Medium (placeholder는 data 손실 아님, 그러나 오해 유발)

## §G. Anti-Patterns

- **AP-CCI-001**: F1 수정 시 플래그 rename 수반. → rename은 P2 Out of Scope.
- **AP-CCI-002**: F2에서 중앙 테이블 없이 각 소스에 매핑 추가. → scatter 부채 발생. 반드시 단일 테이블.
- **AP-CCI-003**: F3 (b) 값 보존 선택 시 기존 settings.local.json 동기화 로직 파괴. → (a) 확인 라인 우선.
- **AP-CCI-004**: P0-4 수정 시 template source 없이 local만 수정. → Template-First Rule 위반.
- **AP-CCI-005**: template source에 SPEC ID/REQ 토큰 누출. → §25 위반, CI guard(`template-neutrality-check.yaml`) 블록.

## §H. Cross-References

- `.moai/specs/SPEC-V3R6-CLI-CONFIG-INTEGRITY-001/spec.md` — 요구사항 본체
- `.moai/specs/SPEC-V3R6-CLI-CONFIG-INTEGRITY-001/acceptance.md` — AC 매트릭스
- `internal/cli/CLAUDE.md` — module conventions (subagent boundary, hardcoding, settings helpers)
- `CLAUDE.local.md` §2, §14, §22, §25 — 제약 SSOT
- 선행 SPEC: `SPEC-V3R6-CLI-AUDIT-001`(CLI audit 계보), `SPEC-V3R5-INIT-WIZARD-EXPANSION-001`(wizard 구조)

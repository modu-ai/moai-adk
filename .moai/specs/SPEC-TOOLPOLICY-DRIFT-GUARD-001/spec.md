---
id: SPEC-TOOLPOLICY-DRIFT-GUARD-001
title: "tool-policy.yaml ↔ settings.json 권한 블록 드리프트 검사 도입과 YAML 정합 복구"
version: "0.1.3"
status: in-progress
created: 2026-09-10
updated: 2026-09-10
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/config/toolpolicy"
lifecycle: spec-anchored
tags: "tool-policy, settings-json, permissions, drift-check, t619"
tier: M
related_specs: [SPEC-V3R6-TOOL-POLICY-SSOT-001, SPEC-TOOLPOLICY-DEPLOY-REVIEW-001]
---

# SPEC-TOOLPOLICY-DRIFT-GUARD-001 — tool-policy.yaml ↔ settings.json 드리프트 검사

## HISTORY

| 날짜 | 버전 | 사건 |
|---|---|---|
| 2026-09-10 | 0.1.0 | 생성 (카드 t619, plan 단계). 근거는 `.moai/reports/t619/verdict.md`. 운영자 결정 3건(수리 방향·검사 위치·주장 정정)과 plan 제안 검토 뒤 받은 추가 결정 3건(집합 비교·주장 정정 전수·워킹 트리 판독)을 반영했다. SPEC-V3R6-TOOL-POLICY-SSOT-001 의 AC-TPS-005 는 이미 종결된 SPEC 이므로 고치지 않고, 그 AC 가 드리프트를 막지 못했다는 사실만 이 문서에 기록한다. |
| 2026-09-10 | 0.1.1 | plan-audit 1회차 FAIL(0.71) 수리. `.moai/reports/plan-audit/SPEC-TOOLPOLICY-DRIFT-GUARD-001-review-1.md` 의 D1-D10 반영: REQ-TDG-005 정정 대상에 열거 밖 진술 두 곳 추가(D2), REQ-TDG-004 에 해석 실패와 비어 있음 판정의 양쪽 대상 명시(D3·D10), 패턴 라벨과 규범 서술어 정리(D9). |
| 2026-09-10 | 0.1.2 | plan-audit 2회차 FAIL(0.79) 수리. Tier M 감사 상한(`harness.yaml` `plan_audit_tier_ceilings.M: 2`)에 도달했으나, **운영자가 수정 후 3회차 감사 1회를 명시적으로 승인**했다(상한 연장). `.moai/reports/plan-audit/SPEC-TOOLPOLICY-DRIFT-GUARD-001-review-2.md` 의 D11-D20 반영: REQ-TDG-004 에 실패 원인 구별 계약 추가(D11), §4 뮤테이션 범위 문장 한정(D12), REQ-TDG-005 제외 목록에 AC-TPS-005 테스트 주석 명시(D19). |
| 2026-09-10 | 0.1.3 | plan-audit 3회차(운영자가 연장한 최종 회차) FAIL — 점수 0.89(Tier M 기준 0.80 초과), blocking 결함 1건(D21). `.moai/reports/plan-audit/SPEC-TOOLPOLICY-DRIFT-GUARD-001-review-3.md`. **운영자가 "지금 고치고 진행"을 선택**했다. 최종 감사 뒤의 수정이므로 추가 독립 감사는 없고, **수정 줄은 재감사가 아니라 오케스트레이터가 직접 확인한다.** 반영: REQ-TDG-004 에 중복·겹침 전용 판정의 구별 보고와 목록 타입 오류를 해석 실패로 보는 조항 추가(D21·D22), 수용 기준과 복원 절차 수리는 acceptance.md·plan.md(D21-D23), 처분 표는 plan.md §H 3회차. |

## §1 문제 진술

`.moai/config/sections/tool-policy.yaml` 은 자신이 `.claude/settings.json` 권한 블록의 단일 원천이며 `moai tool-policy build` 로 생성하므로 "YAML↔settings.json 드리프트가 구조적으로 막힌다"고 머리말에 적는다. 현재 이 진술은 사실이 아니다.

- 커밋된 YAML 로 생성기를 돌리면 `allow=108 ask=0 deny=60 env_gated_skipped=5` 가 나오고, 커밋된 settings.json 은 allow 114 / deny 48 이다. 서로 다른 항목은 20개다.
  - YAML 에 없는 allow 7개: `CronCreate` `CronDelete` `CronList` `EnterPlanMode` `ExitPlanMode` `EnterWorktree` `ExitWorktree`
  - YAML 에만 있는 allow 1개: `MultiEdit` (퇴역 도구)
  - YAML 에만 있는 deny 12개: `Glob` / `Grep` / `Write` 의 `./secrets/**` · `~/.ssh/**` · `~/.aws/**` · `~/.config/gcloud/**` 경로 규칙
- 같은 20개 변경이 이미 한 번 생성기 실행으로 settings.json 에 쓰였고(`519b848fb`), 손으로 되돌려졌다(`8df71b18d`, `0de8517e5`). YAML 은 그때 고쳐지지 않았다.
- 드리프트를 재는 장치가 없다. Makefile 타깃도 CI 단계도 없다. 방지를 주장한 AC-TPS-005 의 테스트는 임시 샘플 문서에서 YAML 편집이 전파되는지만 보며, 커밋된 두 파일을 서로 비교하지 않는다. 그래서 드리프트가 있는 지금도 통과한다.
- CI 의 Go 테스트 경로 필터(`.github/workflows/ci.yml:78-92`)에 `.claude/settings.json` 이 없다. settings.json 만 바꾸는 변경에서는 Go 테스트가 돌지 않고 대체 작업이 초록을 보고한다. 앞의 손 되돌림이 바로 이 모양이었다.
- 생성기 출력과 커밋된 settings.json 은 바이트 형태부터 다르다. 생성기는 목록을 정렬하고 자기 들여쓰기로 쓰지만, 커밋본은 정렬돼 있지 않다. 따라서 바이트 비교는 YAML 을 고쳐도 붉은색으로 남는다.

## §2 요구사항 (GEARS)

- **REQ-TDG-001** (Ubiquitous): `tool-policy.yaml` 이 선언하는 allow / ask / deny 명세자 집합은, 생성기와 같은 규칙(env_gate 항목 제외, 중복 제거)을 적용한 뒤 커밋된 `.claude/settings.json` 권한 블록의 allow / ask / deny 집합과 **집합으로** 같아야 한다. 그 결과 YAML 은 `CronCreate` `CronDelete` `CronList` `EnterPlanMode` `ExitPlanMode` `EnterWorktree` `ExitWorktree` 를 allow 로 선언해야 하고, `MultiEdit` allow 와 env_gate 없는 `Glob` / `Grep` / `Write` 경로 deny 12개를 선언해서는 안 된다. env_gate 가 붙은 항목은 이 요구의 삭제 대상이 아니며 내용이 바뀌어서도 안 된다. 비교 대상은 권한 블록의 세 목록뿐이며, `defaultMode` 와 그 밖의 권한 키는 YAML 에서 유도되지 않으므로 비교하지 않는다. `ask` 키가 없으면 빈 목록으로 본다.
- **REQ-TDG-002** (Unwanted): 이 변경은 `.claude/settings.json` 과 `internal/template/templates/.claude/settings.json.tmpl` 의 바이트를 바꿔서는 안 된다. 검증을 위한 어떤 단계도 실제 트리를 대상으로 `moai tool-policy build` 를 실행해서는 안 되며, 생성기 실행은 저장소 밖 스크래치 사본에만 써야 한다.
- **REQ-TDG-003** (Event-driven, 복합): **When** `make build` 가 실행되면, 드리프트 검사는 컴파일보다 먼저(`build` 선행 목록의 한 항목으로) 실행돼야 한다. 검사는 워킹 트리의 `.claude/settings.json` 권한 블록과 워킹 트리의 `tool-policy.yaml` 을 읽어야 한다. **When** 두 집합이 다르면, 검사는 실패해야 하고 서로 다른 명세자 각각을 결정(allow / ask / deny)과 어느 쪽에만 있는지와 함께 이름으로 밝혀야 하며, 조정 방법(YAML 을 의도한 상태로 고친 뒤 생성기로 재생성)을 안내해야 한다. 검사는 저장소 안의 어떤 파일도 쓰거나 재생성해서는 안 된다. **When** 변경이 `.claude/settings.json` 또는 `tool-policy.yaml` 을 건드리면, CI 의 Go 테스트 작업이 이 검사를 실행해야 한다.
- **REQ-TDG-004** (Event-driven, 복합): **When** 어느 한쪽 입력 파일이 없거나, 해석할 수 없거나(YAML 문법·검증 오류, 권한 블록 JSON 오류, 권한 블록의 `allow`·`ask`·`deny` 값이 문자열 목록이 아닌 경우), 권한 블록이 없으면, 검사는 통과하거나 건너뛰어서는 안 되며 실패해야 한다. **When** YAML 에서 유도한 allow 집합, YAML 에서 유도한 deny 집합, settings.json 에서 판독한 allow 집합, settings.json 에서 판독한 deny 집합 네 개 중 **어느 하나라도** 비어 있으면, 검사는 실패해야 한다. ask 집합이 비어 있는 것은 실패 사유가 아니다. 검사는 이 네 가지 실패 원인(입력 부재, 해석 실패, 권한 블록 부재, 빈 집합)을 **서로 구별할 수 있게** 보고해야 하며, 입력 부재·해석 실패·권한 블록 부재를 빈 집합 실패로 보고해서는 안 된다(해석 오류를 빈 정책으로 삼키는 회귀가 빈 집합 규칙에 가려지지 않도록). **When** settings.json 의 allow·ask·deny 중 한 목록 안에 같은 명세자가 두 번 있거나 같은 명세자가 allow 와 deny 에 함께 있으면, 검사는 두 집합이 같더라도 실패해야 하고, 그 실패를 집합 차이와 구별되며 중복과 겹침끼리도 서로 구별되는 판정으로 보고해야 한다(전용 검사를 빼는 회귀가 집합 차이에 가려지지 않도록). **When** 명세자 하나를 일부러 어긋나게 하면 검사는 실패해야 하고, 되돌리면 통과해야 한다.
- **REQ-TDG-005** (Ubiquitous): YAML↔settings.json 드리프트가 구조적으로 막힌다고 주장하거나 머리말 주석이 생성된다고 주장하는 모든 진술은, 드리프트를 실제로 막는 장치가 이 검사라는 사실로 바뀌어야 한다. 대상은 다음 다섯 곳이다(위치는 HEAD `b2cbfd207` 기준).
  1. `tool-policy.yaml:5-9` — 머리말이 생성된다는 진술과 "structurally preventing" 드리프트 방지 진술
  2. `tool-policy.yaml:11-16` — "This SSOT prevents the ANALOGOUS drift class on the surfaces it generates" 문단
  3. `internal/config/toolpolicy/types.go:1-5` — 패키지 문서 주석의 "the YAML header comment (audit surface) are generated" 진술
  4. `internal/config/toolpolicy/types.go:94-98` — `Metadata` 문서 주석의 "structurally preventing" 진술
  5. `tool-policy.yaml:54` — 생성기가 건너뛰는 템플릿을 생성 대상으로 적은 `metadata.generated_into` 항목

  다음은 대상이 아니다. 대조 분석(analogy) 인용인 `types.go:8` 의 "Drift-class analogy:" 와 `tool-policy.yaml:42` 의 교차 참조 목록 항목은 방지를 주장하지 않는다. `internal/config/toolpolicy/codegen_test.go:209-213` 의 AC-TPS-005 테스트 이름·주석("by-construction drift-prevention property")은 §5 "종결된 SPEC 소급 수정" 에 따라 범위 밖이다.

## §3 수용 기준

`AC-TDG-001`..`AC-TDG-010` 은 [`acceptance.md`](./acceptance.md) 에 Given-When-Then 으로 있다. 요구-검증 대응은 아래와 같다.

- **AC-TDG-001** — 커밋된 트리에서 검사 통과, 중복·겹침 없음, 정정 전 붉은색 20개 (maps REQ-TDG-001, REQ-TDG-004)
- **AC-TDG-002** — 생성기 스크래치 실행으로 독립 확인한 집합 일치 (maps REQ-TDG-001, REQ-TDG-002)
- **AC-TDG-003** — 적용 권한 불변, 검사는 읽기 전용(입력 파일 sha 와 워킹 트리 상태 불변) (maps REQ-TDG-002, REQ-TDG-003)
- **AC-TDG-004** — 수동 뮤테이션 대조: YAML 한 항목을 조건부로 어긋나게 하면 붉은색과 조정 안내, 조건부 복원 뒤 초록 (maps REQ-TDG-003, REQ-TDG-004)
- **AC-TDG-005** — 자동 뮤테이션 대조: 양방향 차이, 그리고 집합이 같은 고정 fixture 에서 세 목록 중복과 겹침이 전용 판정으로만 실패 (maps REQ-TDG-004)
- **AC-TDG-006** — 누락·해석 실패(목록 타입 오류 포함)·권한 블록 부재·네 집합 비어 있음에서 원인별로 구별되는 실패, 건너뛰기 없음 (maps REQ-TDG-004)
- **AC-TDG-007** — 주장 정정 다섯 곳 전수 (maps REQ-TDG-005)
- **AC-TDG-008** — `build` 선행 목록 연결과 CI 경로 필터 블록 (maps REQ-TDG-003)
- **AC-TDG-009** — 영향받는 기존 테스트 무회귀 (maps REQ-TDG-001, REQ-TDG-005)
- **AC-TDG-010** — YAML 항목 단위 결과, env_gate 항목 내용 보존 (maps REQ-TDG-001)

## §4 제약 [HARD — run 단계 구속]

- 적용되는 권한은 바뀌지 않는다. settings.json 과 템플릿은 이 SPEC 의 변경 대상이 아니다.
- 검사는 읽기 전용이다. 재생성은 명시적 동사(`moai tool-policy build`)로만 일어난다.
- 실행 중인 세션이 읽는 워킹 트리의 `.claude/settings.json` 은 뮤테이션 대조에서도 건드리지 않는다. 워킹 트리 파일에 대한 제자리 수동 뮤테이션은 YAML 에만 한다. `t.TempDir()` 격리 사본의 settings.json·YAML 뮤테이션은 자동 대조(AC-TDG-005)에서 허용한다.
- env_gate 가 붙은 YAML 항목은 삭제하거나 고치지 않는다.
- 비ASCII 문자는 원래 UTF-8 로 쓴다. 역슬래시-u 이스케이프 표기를 산출물에 남기지 않는다.

## §5 범위

### Out of Scope — 템플릿 대조

- `settings.json.tmpl` 권한 블록을 YAML 과 비교하지 않는다. 템플릿 권한 블록은 GitMode 에 따른 렌더 시점 조건문을 담고 있어 생성기가 이미 건너뛰는 대상이며(`internal/cli/tool_policy.go:139-149`), 이 카드가 다루는 파일 쌍과 다른 쌍이다.

### Out of Scope — 적용 권한과 파일 형태 변경

- settings.json 의 권한 항목을 추가·삭제하지 않는다.
- settings.json 을 생성기의 정렬된 바이트 형태로 다시 쓰지 않는다.

### Out of Scope — 생성기·CLI 동작 변경

- 생성기의 정렬·들여쓰기, env_gate 항목의 훅 방출, YAML 부재 시 CLI 의 무동작 안내는 바꾸지 않는다.
- 생성기 판독 경로(`settings_region.go` `extractStringList`)가 목록 타입 오류를 버리는 동작은 바꾸지 않는다. 엄격한 목록 타입 판정은 테스트 전용 비교기에서만 한다.
- 검사가 드리프트를 발견해도 자동으로 재생성하지 않는다.

### Out of Scope — YAML 에서 유도되지 않는 설정

- `defaultMode`, 그 밖의 권한 키, `settings.local.json`, 사용자 전역 설정은 비교하지 않는다.

### Out of Scope — 종결된 SPEC 소급 수정

- SPEC-V3R6-TOOL-POLICY-SSOT-001 의 AC-TPS-005 와 그 테스트 이름·주석(`codegen_test.go:209-213`)은 고치지 않는다.

## §6 의존과 참조

- 근거: `.moai/reports/t619/verdict.md`
- plan 감사: `.moai/reports/plan-audit/SPEC-TOOLPOLICY-DRIFT-GUARD-001-review-1.md`, `-review-2.md`, `-review-3.md`
- 선례: `Makefile:41-49` (`agents-emit-check`), `Makefile:54-60` (`commands-emit-check`)
- 관련 SPEC: SPEC-V3R6-TOOL-POLICY-SSOT-001 (생성기 도입), SPEC-TOOLPOLICY-DEPLOY-REVIEW-001 (YAML 템플릿 배포 제거, dev 전용화)

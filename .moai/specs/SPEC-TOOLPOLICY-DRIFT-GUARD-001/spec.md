---
id: SPEC-TOOLPOLICY-DRIFT-GUARD-001
title: "tool-policy.yaml ↔ settings.json 권한 블록 드리프트 검사 도입과 YAML 정합 복구"
version: "0.1.0"
status: draft
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

| 날짜 | 사건 |
|---|---|
| 2026-09-10 | 생성 (카드 t619, plan 단계). 근거는 `.moai/reports/t619/verdict.md`. 운영자 결정 3건(수리 방향·검사 위치·주장 정정)과 plan 제안 검토 뒤 받은 추가 결정 3건(집합 비교·주장 정정 전수·워킹 트리 판독)을 반영했다. SPEC-V3R6-TOOL-POLICY-SSOT-001 의 AC-TPS-005 는 이미 종결된 SPEC 이므로 고치지 않고, 그 AC 가 드리프트를 막지 못했다는 사실만 이 문서에 기록한다. |

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

- **REQ-TDG-001** (Ubiquitous): `tool-policy.yaml` 이 선언하는 allow / ask / deny 명세자 집합은, 생성기와 같은 규칙(env_gate 항목 제외, 중복 제거)을 적용한 뒤 커밋된 `.claude/settings.json` 권한 블록의 allow / ask / deny 집합과 **집합으로** 같아야 한다. 그 결과 YAML 은 `CronCreate` `CronDelete` `CronList` `EnterPlanMode` `ExitPlanMode` `EnterWorktree` `ExitWorktree` 를 allow 로 선언하고, `MultiEdit` allow 와 `Glob` / `Grep` / `Write` 경로 deny 는 하나도 선언하지 않는다. 비교 대상은 권한 블록의 세 목록뿐이며, `defaultMode` 와 그 밖의 권한 키는 YAML 에서 유도되지 않으므로 비교하지 않는다. `ask` 키가 없으면 빈 목록으로 본다.
- **REQ-TDG-002** (Unwanted): 이 변경은 `.claude/settings.json` 과 `internal/template/templates/.claude/settings.json.tmpl` 의 바이트를 바꾸지 않는다. 검증을 위한 어떤 단계도 실제 트리를 대상으로 `moai tool-policy build` 를 실행하지 않으며, 생성기 실행은 저장소 밖 스크래치 사본에만 쓴다.
- **REQ-TDG-003** (Event-driven, 복합): **When** `make build` 가 실행되면, 드리프트 검사는 컴파일보다 먼저 실행된다. 검사는 워킹 트리의 `.claude/settings.json` 권한 블록과 워킹 트리의 `tool-policy.yaml` 을 읽는다. **When** 두 집합이 다르면, 검사는 실패하고 서로 다른 명세자 각각을 결정(allow / ask / deny)과 어느 쪽에만 있는지와 함께 이름으로 밝히며, 조정 방법(YAML 을 의도한 상태로 고친 뒤 생성기로 재생성)을 안내한다. 검사는 어떤 파일도 쓰거나 재생성하지 않는다. **When** 변경이 `.claude/settings.json` 또는 `tool-policy.yaml` 을 건드리면, CI 의 Go 테스트 작업이 이 검사를 실행한다.
- **REQ-TDG-004** (Event-detected): **When** 어느 한쪽 파일이 없거나, 해석할 수 없거나, 권한 블록이 없거나, 유도한 allow 또는 deny 집합이 비어 있으면, 검사는 통과하지도 건너뛰지도 않고 실패한다. **When** 커밋된 settings.json 의 한 목록 안에 같은 명세자가 두 번 있거나 같은 명세자가 allow 와 deny 에 함께 있으면, 집합 비교와 별도로 검사는 실패한다. **When** 명세자 하나를 일부러 어긋나게 하면 검사는 실패하고, 되돌리면 통과한다.
- **REQ-TDG-005** (Ubiquitous): YAML↔settings.json 드리프트가 구조적으로 막힌다고 주장하는 모든 진술은 드리프트를 실제로 막는 장치가 이 검사라는 사실로 바뀐다. 대상은 `tool-policy.yaml` 머리말의 드리프트 방지 진술, 같은 머리말의 "이 머리말 주석이 생성된다"는 진술, `internal/config/toolpolicy/types.go` 의 같은 취지 문서 주석, 그리고 생성기가 건너뛰는 템플릿을 생성 대상으로 적은 `metadata.generated_into` 항목이다.

## §3 수용 기준

`AC-TDG-001`..`AC-TDG-010` 은 [`acceptance.md`](./acceptance.md) 에 Given-When-Then 으로 있다. 요구-검증 대응은 아래와 같다.

- **AC-TDG-001** — 커밋된 트리에서 검사 통과, 중복·겹침 없음 (maps REQ-TDG-001, REQ-TDG-004)
- **AC-TDG-002** — 생성기 스크래치 실행으로 독립 확인한 집합 일치 (maps REQ-TDG-001, REQ-TDG-002)
- **AC-TDG-003** — 적용 권한 불변, 검사는 읽기 전용 (maps REQ-TDG-002, REQ-TDG-003)
- **AC-TDG-004** — 수동 뮤테이션 대조: YAML 한 항목을 어긋나게 하면 붉은색, 되돌리면 초록 (maps REQ-TDG-003, REQ-TDG-004)
- **AC-TDG-005** — 자동 뮤테이션 대조: 양방향 + 중복·겹침 (maps REQ-TDG-004)
- **AC-TDG-006** — 누락·빈 입력에서 실패, 건너뛰기 없음 (maps REQ-TDG-004)
- **AC-TDG-007** — 주장 정정 전수 (maps REQ-TDG-005)
- **AC-TDG-008** — `make build` 선행 연결과 CI 경로 필터 (maps REQ-TDG-003)
- **AC-TDG-009** — 영향받는 기존 테스트 무회귀 (maps REQ-TDG-001, REQ-TDG-005)
- **AC-TDG-010** — YAML 항목 단위 결과 (maps REQ-TDG-001)

## §4 제약 [HARD — run 단계 구속]

- 적용되는 권한은 바뀌지 않는다. settings.json 과 템플릿은 이 SPEC 의 변경 대상이 아니다.
- 검사는 읽기 전용이다. 재생성은 명시적 동사(`moai tool-policy build`)로만 일어난다.
- 실행 중인 세션이 읽는 settings.json 은 뮤테이션 대조에서도 건드리지 않는다. 뮤테이션은 YAML 쪽에만 한다.
- 비ASCII 문자는 원래 UTF-8 로 쓴다. 역슬래시-u 이스케이프 표기를 산출물에 남기지 않는다.

## §5 범위

### Out of Scope — 템플릿 대조

- `settings.json.tmpl` 권한 블록을 YAML 과 비교하지 않는다. 템플릿 권한 블록은 GitMode 에 따른 렌더 시점 조건문을 담고 있어 생성기가 이미 건너뛰는 대상이며(`internal/cli/tool_policy.go:139-149`), 이 카드가 다루는 파일 쌍과 다른 쌍이다.

### Out of Scope — 적용 권한과 파일 형태 변경

- settings.json 의 권한 항목을 추가·삭제하지 않는다.
- settings.json 을 생성기의 정렬된 바이트 형태로 다시 쓰지 않는다.

### Out of Scope — 생성기·CLI 동작 변경

- 생성기의 정렬·들여쓰기, env_gate 항목의 훅 방출, YAML 부재 시 CLI 의 무동작 안내는 바꾸지 않는다.
- 검사가 드리프트를 발견해도 자동으로 재생성하지 않는다.

### Out of Scope — YAML 에서 유도되지 않는 설정

- `defaultMode`, 그 밖의 권한 키, `settings.local.json`, 사용자 전역 설정은 비교하지 않는다.

### Out of Scope — 종결된 SPEC 소급 수정

- SPEC-V3R6-TOOL-POLICY-SSOT-001 의 AC-TPS-005 와 그 테스트 이름·주석은 고치지 않는다.

## §6 의존과 참조

- 근거: `.moai/reports/t619/verdict.md`
- 선례: `Makefile:41-49` (`agents-emit-check`), `Makefile:54-60` (`commands-emit-check`)
- 관련 SPEC: SPEC-V3R6-TOOL-POLICY-SSOT-001 (생성기 도입), SPEC-TOOLPOLICY-DEPLOY-REVIEW-001 (YAML 템플릿 배포 제거, dev 전용화)

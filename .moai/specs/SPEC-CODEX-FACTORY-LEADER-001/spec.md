---
id: SPEC-CODEX-FACTORY-LEADER-001
title: "Codex 팩토리 리더 진입 허용"
version: "0.1.0"
status: draft
created: 2026-10-19
updated: 2026-10-19
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "internal/cli, internal/hook"
lifecycle: spec-anchored
tags: "codex, factory, leader, queue, card-t1774"
tier: M
related_specs: [SPEC-FACTORY-SELF-DISPATCH-001, SPEC-CODEX-FACTORY-RETIRE-001, SPEC-FACTORY-CODEX-LEADER-RESUME-001]
---

# SPEC-CODEX-FACTORY-LEADER-001 — Codex 팩토리 리더

## HISTORY

- 2026-10-19 · v0.1.0 · manager-spec · t1774 초안. 운영자 결정 `d-20261015T000351Z-0dfd`에 따라 Codex 리더 금지를 부분 대체한다. 측정한 작성 기준은 `WT-codex-factory-leader`, `8249bff8c3933d647b4f99e59048595f9808e83a`다.

## §0 전제와 결정

t1745의 반려 기록 `a9f53c169`에 있는 SPEC-FACTORY-CODEX-LEADER-RESUME-001의 spec.md·plan.md를 원문 그대로 보존한다. 이 기록의 `rejected` 상태는 유지한다. 당시 리더 진입 금지가 이번 운영자 결정으로 번복되지만, 네이티브 Codex resume 재채택까지 승인된 것으로 해석하지 않는다.

현재 카드의 기능은 **새 Codex 리더 런의 명시적 시작**이다. 기존 레인 진입, 기본 리더 선택, 메시징 전송 계층은 변경하지 않는다. 실행 완료를 메시지 도착으로 판단하지 않고 저장된 진행·완료 기록으로 확인한다.

## §A 목표

운영자가 `moai codex -f`로 Codex 리더를 시작하고 기존 큐와 dispatch 기록으로 팩토리를 운영할 수 있어야 한다. Codex에 없는 PermissionRequest·PreCompact 훅과 비동기 메시징 nudge에 기대지 않아야 한다.

## §B 요구사항 — GEARS

- **REQ-CFL-001** — 운영자가 Codex CLI의 bare `-f` 또는 `--factory` 리더 진입을 요청할 때(When), 런처는 새 팩토리 런의 리더로 Codex를 한 번 실행하고 리더 역할과 런 식별자를 전달해야 한다(shall).
- **REQ-CFL-002** — Codex 리더가 시작될 때(When), 팩토리는 리더 백엔드를 `codex`로 기록하고 기존 리더 런의 수명주기와 용량 정책을 적용해야 한다(shall).
- **REQ-CFL-003** — 메시징 nudge를 사용할 수 없는 동안에도(While), 리더는 기존 큐에서 배차 대상 카드를 읽고 런·레인·카드에 귀속된 dispatch 파일을 생성하며 레인의 저장된 진행·완료 기록을 읽어 후속 처리를 판단할 수 있어야 한다(shall).
- **REQ-CFL-004** — Codex 리더 안내는 기존 manager-lead 책임에 대응하는 큐 조회·배차·저장 기록 확인 절차와 Codex의 훅 제약을 명시해야 한다(shall).
- **REQ-CFL-005** — 값이 붙은 리더 형태, 숫자·옛 역할 토큰, 충돌한 리더·레인 옵션 또는 CLI 이외의 factory 진입을 받을 때(When), 런처는 child 실행·런 등록·레인 claim·작업트리 생성 전에 명확한 오류로 거부해야 한다(shall).
- **REQ-CFL-006** — 런처는 기존 `moai codex -l` 레인 진입과 plain Codex의 인수·모델 동작, Claude/GLM 리더 동작을 유지해야 한다(shall).
- **REQ-CFL-007** — Codex 리더를 시작할 때(When), 런처는 기존 설정의 리더 모델·effort 기본값과 명시적 사용자 덮어쓰기를 적용하고 살아 있거나 확인 불가능한 다른 작업트리 writer를 보호하는 anchor 검사를 유지해야 한다(shall).
- **REQ-CFL-008** — 레인이 Codex 리더 런에 합류할 때(When), 팩토리는 레인 백엔드와 무관하게 기존 런·레인 검증을 적용하며 Claude/GLM 리더가 그 런을 인수하는 것은 계속 거부해야 한다(shall).
- **REQ-CFL-009** — SPEC 문서들은 운영자 결정 id를 두 선행 SPEC의 HISTORY에 남기고 REQ-SD-004·REQ-CFR-010과 관련 제외 문장을 이번 허용 범위에 맞게 개정하며 이전 반려 기록을 보존해야 한다(shall).
- **REQ-CFL-010** — 문서는 factory 안내와 README에서 Codex 리더 시작, 레인 진입의 차이, 저장 기록을 통한 확인, 이번 변경의 resume 한계를 설명해야 한다(shall).
- **REQ-CFL-011** — bare Codex 리더를 시작할 때(When), 팩토리는 종료된 이전 런의 resume 후보를 자동 재채택하지 않고 새 런을 시작해야 한다(shall). 호출자가 레인의 런·역할·worker 환경을 상속했더라도 기존 런의 backend·소유자를 바꾸지 않아야 한다(shall not).

## §C 수용 기준

`acceptance.md`의 AC-CFL-001…009가 검증 계층이다. 계획 단계에서는 구현·런타임 성공을 주장하지 않는다.

## §D 제약

- 기존 큐, broker, 리더 런과 훅 경로를 재사용한다. 새 메시징 전송 계층을 만들지 않는다.
- 저장된 dispatch와 완료 기록이 실제 동작의 근거다. 안내 문자열만 존재하는 것으로 수용하지 않는다.
- 테스트는 격리된 임시 프로젝트·홈·실제 큐 저장소를 사용하고 실제 대화형 Codex/Claude/GLM을 시작하지 않는다.
- 품질 기준과 개발 모드는 `.moai/config/sections/quality.yaml`을 따른다.

## §E Exclusions

### Out of Scope — 인접 기능

- 레인 CLI 형태 변경, Codex를 기본 리더로 선택하는 정책 변경, 메시지 전송 계층 수정.
- Codex 네이티브 resume 파서, 이전 런 자동 재채택, Codex 리더의 명시적 `--factory-run` 재개.
- t1745 구현 코드 흡수, 기존 반려 SPEC의 상태 재개, t1645의 광범위한 리더 운영 재설계.
- Codex가 제공하지 않는 훅 이벤트 추가와 전체 저장소 로컬 테스트 실행.

## §F 추적성

요구사항과 수용 기준의 대응은 acceptance.md에 있다. 실행 증거는 progress.md §E.2, 실행 감사 준비는 §E.3, 문서·설치 검증은 §E.4에서 각 소유자가 기록한다.


---
id: SPEC-MOAI-CG-RETIRE-001
title: "CG 실행 경로 철거와 명시적 구성 이전"
version: "0.1.0"
status: draft
created: 2026-09-11
updated: 2026-09-11
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "internal/cli, internal/config, internal/tmux, internal/template, docs-site"
lifecycle: spec-anchored
tags: "gateway, migration, integration"
tier: L
related_specs: [SPEC-MOAI-GATEWAY-001, SPEC-MOAI-GPT-AUTH-001, SPEC-MOAI-GATEWAY-PICKER-001]
---

# SPEC-MOAI-CG-RETIRE-001

## HISTORY

- 0.1.0 (2026-09-11) — 관련 카드 전체 구현 목표의 형제 계획. 계획·구현 완료를 주장하지 않는다. SP-B1 보강으로 명령·저장 필드·첫 launch guard·고정 이전 fixture를
  같은 버전의 design/acceptance에 확정했다. 새 REQ/AC 번호는 없다.

## 목적과 범위

CG 실행 경로 철거와 명시적 구성 이전을 실제 실행과 설정 보존으로 검증한다. 코어의 분리 범위를 담당하며 기존 사용자 선택을 자동 축약하지 않는다.

## 요구사항

**REQ-CR-001** — When 사용자가 폐기된 cg 실행을 요청하면, the system shall Claude·GLM 프로세스를 실행하거나 인증·tmux·설정 상태를 변경하지 않고 폐기 안내와 지원 launcher 선택 방법을 반환한다. 현재 명령 목록은 cg를 실행 가능한 명령으로 제공하지 않는다.

**REQ-CR-002** — When 기존 프로젝트에서 team_mode: cg를 읽으면, the system shall 이를 자동으로 Claude·GLM·GPT 단일 provider로 바꾸지 않고 기존 leader/teammate 의도가 미이전 상태임을 표시한다. 사용자가 명시한 대체 구성이 있기 전 해당 legacy 모드 실행은 실패해야 한다.

**REQ-CR-003** — When 사용자가 대체 launcher와 teammate 구성을 명시하여 이전하면, the system shall 선택한 변경 내용과 원본을 보존하고 대상 구성만 원자적으로 변경한다. 취소·충돌·저장 실패는 원본을 보존하며, 알 수 없는 키·프로필·credential 참조·다른 설정을 지워서는 안 된다.

**REQ-CR-004** — The system shall 현재 실행 경로에서 CG 전용 모드 적용·tmux GLM 주입·backend 판정을 제거하고 지원 launcher의 명시 provider 판정만 사용한다. legacy 문자열은 이전 식별과 역사 읽기에만 남길 수 있다.

**REQ-CR-005** — The system shall kanban/factory 및 backend 선택에서 CG의 폐기를 명시 처리하고 이전에 금지된 조합을 새 provider로 자동 변환해서 허용하지 않는다. 지원 cc·glm·gpt 경로의 기존 권한·프로필·worktree·spawn 계약을 유지한다.

**REQ-CR-006** — The system shall 한국어·영어·일본어·중국어의 현재 문서, README, 배포 template과 CLI 도움말에서 CG 실행 안내를 제거하거나 명시 이전 안내로 교체한다. 과거 release note·감사·폐기 SPEC·사용자 대화 기록은 재작성하지 않는다.

**REQ-CR-007** — The system shall CG 제거로 바뀐 현재 테스트 계약을 새 거절·명시 이전·provider 판정으로 대체하고 지원 launcher의 비회귀를 실행한다. 단순 cg 텍스트 부재만으로 runtime 제거·문서 정합성을 판정하지 않는다.

**REQ-CR-008** — Where TEAMMATE의 명시 provider/pane 기능을 필요로 하는 CG 이전이 요청되면, the system shall 그 기능의 실제 통합 게이트가 충족될 때만 동등한 이전 완료를 표시한다. 코어 in-process 제한을 pane 동등성의 증거로 사용하지 않는다.

## 제외 범위 (out of scope)

### Out of Scope — 다른 계약과 사용자 자료

- gateway protocol·AUTH 저장소·PICKER 본문을 중복 구현하거나 이 계획에서 변경하지 않는다.
- 사용자 credential/프로필 삭제·자동 복사, 역사 기록 일괄 변경, 무단 provider fallback은 하지 않는다.
- 숫자 카드 DB 변경, 원격 출시·병합·worktree 정리는 포함하지 않는다.

## 수용 기준

동일 번호 REQ/AC 연결은 acceptance.md에 있다. 계획 검토와 실제 runtime 근거를 분리한다.

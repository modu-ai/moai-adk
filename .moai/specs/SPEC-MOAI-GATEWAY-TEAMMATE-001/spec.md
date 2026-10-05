---
id: SPEC-MOAI-GATEWAY-TEAMMATE-001
title: "gateway teammate pane의 인증·모델·수명 격리"
version: "0.1.0"
status: draft
created: 2026-09-11
updated: 2026-09-11
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "internal/cli, internal/tmux, internal/gateway, internal/hook"
lifecycle: spec-anchored
tags: "gateway, migration, integration"
tier: L
related_specs: [SPEC-MOAI-GATEWAY-001, SPEC-MOAI-GPT-AUTH-001, SPEC-MOAI-GATEWAY-PICKER-001]
---

# SPEC-MOAI-GATEWAY-TEAMMATE-001

## HISTORY

- 0.1.0 (2026-09-11) — 관련 카드 전체 구현 목표의 형제 계획. 계획·구현 완료를 주장하지 않는다.

## 목적과 범위

gateway teammate pane의 인증·모델·수명 격리을 실제 실행과 설정 보존으로 검증한다. 코어의 분리 범위를 담당하며 기존 사용자 선택을 자동 축약하지 않는다.

## 요구사항

**REQ-GT-001** — When 사용자가 gateway teammate pane 실행을 명시 선택하면, the system shall 각 pane을 해당 lead/gateway에 귀속시키고 세션 접근 인증·모델·provider를 개별로 전달한다. 같은 tmux 세션의 다른 gateway나 비gateway pane의 설정을 덮어쓰지 않는다.

**REQ-GT-002** — The system shall pane 인증을 위해 tmux server/global/session env에 credential이나 세션 접근 token을 기록하지 않는다. credential·token을 shell 명령문·tmux argv·로그에 넣지 않으며 pane bootstrap 이후 다른 pane이 재사용할 수 없게 한다.

**REQ-GT-003** — When pane 프로세스가 시작되면, the system shall 상속된 stale GLM·gateway 주소·인증·프로필 선택을 검증하여 해당 pane의 명시 launch 값으로 구성한다. 타 세션의 token·주소로 요청하거나 stale Z.AI 경로로 직접 요청해서는 안 된다.

**REQ-GT-004** — When lead가 정상 종료·오류·강제 종료되면, the system shall 해당 gateway의 pane 인증과 후속 송신을 폐기하고 자기 소유 pane 프로세스만 정리한다. 다른 lead의 pane·사용자 tmux 세션은 유지한다.

**REQ-GT-005** — When teammate가 모델 별칭 또는 정확한 ID로 요청하면, the system shall 명시한 모델 정책에 맞는 canonical catalog ID와 provider로 라우팅한다. GLM tier는 코어 매핑을 유지하고 GPT 네 모델은 정확한 ID로 검증하며 알 수 없는 ID는 fallback 없이 거절한다.

**REQ-GT-006** — The system shall 동시 lead·pane·재개·pane 생성 실패·재사용 상황에서 소유권과 단일 소비 bootstrap을 검증한다. 부모 PID 재사용이나 같은 tmux 세션 이름만으로 소유권을 인정해서는 안 된다.

**REQ-GT-007** — Where 실제 Claude Code가 pane별 bootstrap과 수명 계약을 만족하는 실행 연결을 제공하면, the system shall 독립 감사와 실제 pane 시험 후 이 기능을 활성화한다. 연결이 확인되지 않으면 기존 코어의 in-process 제한을 유지하고 pane 지원 완료를 표시하지 않는다.

**REQ-GT-008** — The system shall 사용자 정책·permission·프로필·MCP·fallback 설정을 무음 변경하지 않고 지원 불가 조합을 명시 거절한다. 다른 provider credential 및 사용자 설정을 복사하여 pane 연결을 성립시켜서는 안 된다.

**REQ-GT-009** — The system shall 같은 tmux 세션의 두 gateway와 비gateway pane이 공존하는 실제 시험에서 pane 요청·도구 왕복·대화 맥락·stream 종료와 부모 수명 연동을 검증한다. mock pane 생성만으로 전체 기능 완료를 판정하지 않는다.

## 제외 범위 (out of scope)

### Out of Scope — 다른 계약과 사용자 자료

- gateway protocol·AUTH 저장소·PICKER 본문을 중복 구현하거나 이 계획에서 변경하지 않는다.
- 사용자 credential/프로필 삭제·자동 복사, 역사 기록 일괄 변경, 무단 provider fallback은 하지 않는다.
- 숫자 카드 DB 변경, 원격 출시·병합·worktree 정리는 포함하지 않는다.

## 수용 기준

동일 번호 REQ/AC 연결은 acceptance.md에 있다. 계획 검토와 실제 runtime 근거를 분리한다.

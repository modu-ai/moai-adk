---
id: SPEC-MOAI-GATEWAY-PICKER-001
title: "gateway 세션의 초기 모델과 안전한 모델 선택"
version: "0.2.0"
status: draft
created: 2026-09-11
updated: 2026-09-11
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "internal/cli, internal/gateway"
lifecycle: spec-anchored
tags: "gateway, gpt, claude-code"
tier: M
related_specs: [SPEC-MOAI-GATEWAY-001]
---

# SPEC-MOAI-GATEWAY-PICKER-001

## HISTORY

- 0.2.0 (2026-09-11) — 실제 TUI의 s/Enter, Astra·Terra·Default 요청, 정확한 UUID resume의 합성 carrier와
  metadata 세션 ID 보존을 근거로 추가했다. 설정 격리·인증 보존·모든 재개·실계정 수용은 별도 게이트이며
  기존 REQ/AC 번호·상태·출시 조건은 유지한다.
  같은 버전에서 receipt RPA-1의 retained family/profile 수명·사전 UUID 인계·원래 secure-storage namespace 보존을
  plan과 기존 AC-GP-003·005에 보강했다. 아직 미실행인 continue/선택/fork를 완료로 표시하지 않는다.

- 0.1.0 (2026-09-11) — 사용자 전체 GPT 구현 목표에 따른 형제 계획 초안. 구현·계획 감사 완료를 주장하지 않는다. 같은 버전에서 공식 picker 설정·discovery 필터·Default 후보를 계획에 보강했다.
  실제 TUI의 표시·설정 저장 격리는 미검증이며 REQ/AC 번호와 완료 범위는 바꾸지 않았다.

## 목적

`moai gpt`로 연 Claude Code에서 네 GPT 모델을 실제로 선택하고 같은 세션에서 사용할 수 있게 한다. 시작 모델·재개·Default 행·설정 보존을 명시하여 저장된 모델 때문에 사용자가 고르지 않은 provider로 요청이 나가지 않게 한다.

## 요구사항

**REQ-GP-001** — When 사용자가 moai gpt로 새 세션을 열면, the system shall launcher는 명시 모델 선택, 해당 launcher의 비어 있지 않은 기본값, launcher의 문서화된 기본 모델 순서로 초기 모델을 정하고 Claude Code에 명시적으로 전달해야 한다. gpt의 기본 모델은 gpt-5.6-sol이며 초기 선택을 사용자에게 표시해야 한다.

**REQ-GP-002** — When moai cc 또는 moai glm이 초기 모델을 정하면, the system shall launcher는 같은 우선순위를 적용하되 자기 provider의 기본값을 사용해야 한다. 빈 기본값 때문에 Claude Code의 전역 저장 모델이 초기 provider를 조용히 바꾸게 해서는 안 된다.

**REQ-GP-003** — When 사용자가 continue 또는 resume으로 대화를 재개하면, the system shall launcher는 대화 기록을 보존하되 이번 실행의 명시 선택 또는 launcher 기본 모델을 초기 모델로 전달하고 표시해야 한다. 이전 대화의 저장 모델을 자동 복원하여 이번 provider 선택을 덮어써서는 안 된다.

**REQ-GP-004** — The 모델 선택 표면 shall 세션 catalog에 있는 gpt-6-astra, gpt-5.6-sol, gpt-5.6-terra, gpt-5.6-luna를 정확한 이름으로 선택할 수 있게 해야 한다. 지원되지 않는 ID는 실행 또는 전환 전에 명시 거절하고 접두사로 upstream 모델을 추측해서는 안 된다.

**REQ-GP-005** — The 모델 선택 표면 shall 사용자 전역 설정과 기존 프로젝트 설정을 변경하지 않는 세션 전용 설정을 사용해야 한다. /model이 저장하는 기본값도 이 세션 경계를 벗어나서는 안 되며 종료·취소·동시 세션은 서로의 설정을 덮어써서는 안 된다.

**REQ-GP-006** — When 사용자가 Default 행을 선택하면, the system shall 모델 선택 표면은 표시한 초기 기본 모델과 provider를 선택해야 한다. 표시와 실제 요청 모델이 다르거나 이전 GLM 슬롯 값 때문에 다른 유료 provider로 조용히 전환되어서는 안 된다.

**REQ-GP-007** — When 사용자가 같은 세션에서 /model 또는 picker s 경로로 모델을 변경하면, the system shall 이후 실제 turn은 선택한 catalog 모델로 요청되어야 하고 대화 맥락과 도구 왕복·stream 종료를 보존해야 한다. 선택 검증 요청은 실제 turn 성공으로 세어서는 안 된다.

**REQ-GP-008** — When 모델 credential이 없거나 만료되어 사용 불가능하면, the system shall 모델 선택 표면은 명시 오류와 현재 선택 상태를 보여야 하고 다른 유료 provider로 자동 전환해서는 안 된다. /model 검증 거절 뒤에는 이전 선택을 보존하고 검증을 생략하는 경로는 첫 turn 오류를 드러내야 한다.

**REQ-GP-009** — The 코어와 모델 선택 표면 shall 함께 통합 검증되기 전 gateway를 사용자 기본 경로로 노출하거나 출시해서는 안 된다. GLM tier 슬롯 주입은 코어 소관으로 유지하고 선택 UI가 상속 GLM 키 정리 순서를 변경해서는 안 된다.

## 제외 범위 (out of scope)

### Out of Scope — 다른 형제의 계약

- gateway ingress·Responses 변환·세션 인증 운반 키·GLM tier 슬롯의 코어 구현을 다시 정의하지 않는다.
- tmux teammate 수명과 CG 철거·설정 마이그레이션은 각 형제 소관이다.
- 사용자 Codex 인증 자동 이관, API 키만으로 구독 로그인 대체, 근거 없는 모델 별칭은 포함하지 않는다.

## 수용 기준 연결

각 REQ의 동일 번호 AC가 `acceptance.md`에 있으며 여러 프로세스·실제 TUI·실계정 조건을 mock으로 치환하지 않는다.

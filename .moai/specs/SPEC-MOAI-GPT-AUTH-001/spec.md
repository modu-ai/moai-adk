---
id: SPEC-MOAI-GPT-AUTH-001
title: "MoAI GPT 로그인과 세대 기반 credential 수명 관리"
version: "0.2.0"
status: draft
created: 2026-09-11
updated: 2026-09-11
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "internal/gateway/auth, internal/cli"
lifecycle: spec-anchored
tags: "gateway, gpt, claude-code"
tier: M
related_specs: [SPEC-MOAI-GATEWAY-001]
---

# SPEC-MOAI-GPT-AUTH-001

## HISTORY

- 0.2.0 (2026-09-11) — fresh broker 로그인과 네 GPT 직접 구독 함수 왕복·Luna opaque 후속의 제한된 관측을
  plan·acceptance의 근거로 추가했다. 코어의 엄격한 SSE 계약을 연결하고 요청별 출력 상한·Windows 내구성의
  사용자 결정 대기를 명시했다. 이후 같은 버전에서 “구독 서버 출력 정책으로 진행”과 “Windows API 기준으로 진행”
  승인을 반영했다. Windows 저장 완료는 파일 flush·같은 volume native write-through 교체·최종 권한/identity/내용
  readback과 CI 프로세스 crash 복구로 판정하며 POSIX directory fsync와 동등한 전원 장애 보장은 주장하지 않는다.
  REQ/AC 번호·draft 상태는 변경하지 않았다.

- 0.1.0 (2026-09-11) — 사용자 전체 GPT 구현 목표에 따른 형제 계획 초안. 구현·계획 감사 완료를 주장하지 않는다. 같은 버전에서 공식 설치 Codex를 새 MoAI 전용 저장소의
  인증 브로커로 재사용하는 계획을 채택했다. 사용자 기존 인증을 가져오지 않고 MoAI의 원자 저장·세대·송신 계약을 유지한다.

## 목적

`moai gpt`가 Claude Code 안에서 GPT-6와 GPT-5.6을 사용하도록 MoAI 소유 로그인·로그아웃·갱신과 송신 직전 인증 해석을 제공한다. 코어 gateway의 provider 중립 경계를 구현하며 기존 Codex 인증 저장소를 소유하지 않는다.

## 요구사항

**REQ-GA-001** — When 사용자가 GPT 로그인을 시작하면, the system shall 인증 계층은 공식 설치 인증 브로커의 새 PKCE 로그인으로 사용자를 인증하고 성공한 credential만 MoAI 소유 저장소에 반영해야 한다. API 키 입력은 구독 로그인 성공으로 표시해서는 안 된다.

**REQ-GA-002** — While 로그인 시도가 진행 중인 동안, the system shall 인증 계층은 매 시도별 예측 불가능한 state·PKCE verifier와 S256 challenge를 사용하고 시작 시도와 정확히 대응하는 callback만 한 번 수락해야 한다. 불일치·누락·중복 callback은 credential 교환과 저장 없이 거절해야 한다. 이 보장은 공식 브로커에 위임하며 MoAI는 브로커 작업 식별·완료 검증을 수행하고 내부 PKCE 시험을 대신했다고 주장해서는 안 된다.

**REQ-GA-003** — When 사용자가 로그인을 취소하거나 정해진 대기 제한이 끝나면, the system shall 인증 계층은 해당 시도의 listener·대기 작업을 종료하고 기존 유효 로그인 상태를 보존해야 한다. 늦게 도착한 callback은 수락해서는 안 된다.

**REQ-GA-004** — The 인증 계층 shall credential과 만료·세대 정보를 MoAI 소유 저장소에 원자적으로 저장하고 현재 사용자만 접근하도록 제한해야 한다. 비밀 값은 argv·로그·오류·보고서에 남겨서는 안 되며, 저장 실패는 성공으로 표시해서는 안 된다.
Where Windows인 경우, the 인증 계층 shall 사용자 승인에 따라 파일 flush·같은 volume의 원자적 write-through 교체·
최종 소유자 전용 권한 및 파일/부모 identity·내용 readback이 모두 성공한 뒤 저장 완료를 표시해야 한다.
프로세스 crash 뒤 완전한 상태 또는 명시 오류로 복구해야 하며 POSIX directory fsync와 동등한 전원 장애 내구성으로
표시해서는 안 된다. 실패 단계와 교체 이후 상태 변경 가능성은 구별해야 한다.

**REQ-GA-005** — When 여러 프로세스가 같은 만료 credential을 해석하면, the system shall 인증 계층은 갱신을 저장소 단위로 직렬화하고 성공한 최신 세대만 반영해야 한다. 늦게 끝난 이전 갱신이 최신 로그인이나 로그아웃 상태를 덮어써서는 안 된다.

**REQ-GA-006** — When 사용자가 로그아웃을 완료하면, the system shall 이후 송신을 시작하는 요청은 이전 세대 credential을 사용해서는 안 된다. 해석과 송신 사이 세대가 달라진 요청은 다시 검증하거나 송신 전에 거절해야 한다. 이미 외부 전송이 시작된 요청은 취소를 시도하고 이를 소급 미송신으로 기록해서는 안 된다. 로컬 폐기와 원격 폐기 결과는 구분하며 원격 실패로 로컬 credential을 되살려서는 안 된다.

**REQ-GA-007** — The GPT credential 참조 shall 코어의 CredentialRef 계약에 따라 송신 시점에 provider·세대를 확인하고 일치하는 허용 endpoint에만 인증을 적용해야 한다. credential 부재·폐기·갱신 실패는 명시 인증 오류로 드러나고 다른 provider나 유료 API 키로 자동 전환해서는 안 된다.

**REQ-GA-008** — The 인증 계층 shall 구독 OAuth와 API 키 인증의 공급자·endpoint·과금 경로를 구분해야 한다. 구독 credential을 API 키 endpoint로 보내거나 반대로 보내서는 안 되며 redirect 대상에 credential을 전달해서는 안 된다.

**REQ-GA-009** — The 인증 계층 shall 기존 CODEX_HOME·Codex 인증 파일·사용자 전역 Claude 설정을 읽어 credential을 자동 복사하거나 변경해서는 안 된다. 명시 로그인은 MoAI 소유 상태만 갱신해야 한다.

**REQ-GA-010** — Where 실제 GPT 구독 로그인과 모델 접근 권한이 검증되면, the system shall moai gpt는 같은 Claude Code 세션에서 gpt-6-astra, gpt-5.6-sol, gpt-5.6-terra, gpt-5.6-luna의 실제 응답과 도구 왕복을 제공해야 한다. 권한 부족을 mock 성공이나 API 키 성공으로 대체해서는 안 된다.

## 제외 범위 (out of scope)

### Out of Scope — 다른 형제의 계약

- gateway ingress·Responses 변환·세션 인증 운반 키·GLM tier 슬롯의 코어 구현을 다시 정의하지 않는다.
- tmux teammate 수명과 CG 철거·설정 마이그레이션은 각 형제 소관이다.
- 사용자 Codex 인증 자동 이관, API 키만으로 구독 로그인 대체, 근거 없는 모델 별칭은 포함하지 않는다.

## 수용 기준 연결

각 REQ의 동일 번호 AC가 `acceptance.md`에 있으며 여러 프로세스·실제 TUI·실계정 조건을 mock으로 치환하지 않는다.

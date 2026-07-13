---
title: 토큰 예산 관리와 우아한 중단
weight: 2
draft: false
---

토크노믹스 4-층 구조의 D층인 예산 방어(Budget defense)를 심화해서 다룹니다. 에이전트가 컨텍스트 윈도우 한계에 다다랐을 때 세션이 멈추지 않고, 진행 상태를 보존하면서 다음 세션이 이어받을 수 있도록 하는 우아한 중단(graceful abort) 메커니즘을 설명합니다.

## 예산 방어의 필요성

Anthropic SSE 스트림은 컨텍스트 윈도우 천장에 가까워지면 `stream_idle_partial` 상태로 간헐적 스톨을 일으킵니다. 이는 확률적이지만 임계값 위에서 예측 가능합니다. 스톨이 발생하면 에이전트 호출이 스트림 중간에 실패하여 진행 상태를 잃을 수 있습니다.

예산 방어는 이 문제를 사전에 해결합니다. 컨텍스트 사용량이 임계에 도달하기 전에 시스템이 우아한 중단을 수행하여, 세션이 손실 없이 다음 단계로 전환되도록 보장합니다.

## 모델별 컨텍스트 임계치

운영 임계값은 모델별로 다릅니다. 큰 윈도우는 더 높은 백분율 사용을 허용하고, 작은 윈도우는 절대 헤드룸이 적습니다.

| 모델 클래스 | 윈도우 | 핸드오프 임계 | 절대 천장 |
|-------------|--------|---------------|-----------|
| Opus 4.8 (1M) | 1,000,000 토큰 | 50% | ~500,000 토큰 |
| GLM-5.2 (1M) | 1,000,000 토큰 | 50% | ~500,000 토큰 |
| Opus / Fable (256K) | 256,000 토큰 | 90% | ~230,000 토큰 |
| Sonnet / Opus 표준 (200K) | 200,000 토큰 | 90% | ~180,000 토큰 |
| Haiku (200K) | 200,000 토큰 | 90% | ~180,000 토큰 |

GLM-5.2(`moai glm` / `moai cg` GLM 패널)는 1M 컨텍스트 모델이므로 50% 임계로 운영합니다. Claude Code가 보고하는 `context_window_size`는 Claude 슬롯 기준(Opus=1M, Sonnet/Haiku=200K)이므로 GLM 세션에서 원시 telemetry가 ~180K를 보여도 MoAI가 1M로 보정합니다. statusline의 CW% 게이지를 신뢰하세요.

## 2-단계 핸드오프 마커

statusline은 컨텍스트 바에 `/clear` 힌트를 2단계로 표시합니다.

- {{< icon warning warn >}} **소프트 마커** `(⚠️/clear)` — 밴드의 소프트 임계에서 표시. 사용자가 판단하여 `/clear`를 실행할 수 있는 권고 신호입니다.
- {{< icon warning danger >}} **하드 마커** `(🛑/clear!)` — auto-compact 인식 천장에서 표시. 다음 액션은 반드시 `/clear`여야 합니다.

하드 천장은 auto-compact 임계값 가까이 설정되므로, 런타임 auto-compact가 먼저 발동하여 하드 마커는 실제로는 드물게 트리거됩니다. 이는 auto-compact 인식 공식의 의도된 트레이드오프입니다.

## 우아한 중단 절차

SPEC-TOKEN-BUDGET-STOP-001로 구현된 우아한 중단 메커니즘은 다음 단계로 작동합니다.

1. **감지** — `Tracker.IsAtHardLimit(agentName)`이 true를 반환 (누적 사용량 ≥ hard_clear_threshold, 기본 0.90)
2. **상태 저장** — 진행 중인 작업 상태를 `progress.md`에 영속화
3. **핸드오프 발행** — 붙여넣기 가능한 resume 메시지를 생성 (6-블록 구조)
4. **턴 종료 권고** — 사용자에게 `/clear`를 권고 (HARD: 자동 `/clear`는 절대 수행 안 함)
5. **증거 영속화** — 검증 증거를 `.moai/state/verify/` 하위에 영속화

`/clear`는 절대 자동으로 실행되지 않습니다. 시스템은 사용자가 `/clear`를 실행하도록 권고만 하며, 사용자가 판단하여 실행합니다.

## paste-ready resume 6-블록 구조

세션 핸드오프 메시지는 다음 6-블록 구조를 따릅니다. 각 블록은 다음 세션이 최소한의 정보로 작업을 이어갈 수 있도록 설계되었습니다.

```text
✂──── 여기부터 복사 ────✂

ultrathink. <SPEC-ID> <phase> 진입.
applied lessons: <memory-file-1>, <memory-file-2>

전제 검증:
1) <검증 가능한 전제 1>
2) <검증 가능한 전제 2>

실행: <명령 또는 액션>

머지 후: <다음 액션 또는 SPEC>

✂──── 여기까지 복사 ────✂
```

각 블록의 역할:

- **블록 1** — `ultrathink.` 오프너가 effort:xhigh를 설정하고, 진입하는 phase와 SPEC-ID를 선언
- **블록 2** — `applied lessons:` 이전 세션에서 학습한 memory 파일 참조 (최대 4개)
- **블록 3** — `전제 검증:` 다음 세션이 시작 전 확인해야 할 검증 가능한 전제 (최대 4개, 각 200자 이내)
- **블록 4** — 전제들의 개별 항목
- **블록 5** — `실행:` 단일 주요 액션 (일반적으로 `/moai <subcommand>`)
- **블록 6** — `머지 후:` 다음 액션 또는 SPEC ID

## 검증 다이어트 (verify-diet)

검증 명령의 장문 출력을 디스크로 리다이렉트하고 컨텍스트에는 요약만 남기는 파일-리다이렉트 계약(file-redirect contract)입니다.

규칙: 검증 명령의 verbatim 출력이 **bounded-tail ceiling**(기본 50줄 또는 2KB 중 작은 쪽)을 초과하면, 출력을 파일로 리다이렉트하고 컨텍스트에는 exit code + bounded tail만 표시합니다.

```bash
go test ./... > /tmp/moai-verify/1-go-test.log 2>&1; echo "exit=$?"; tail -50 /tmp/moai-verify/1-go-test.log
```

이 계약은 "디스크에 verbatim 증거를 남기고, 컨텍스트에는 exit code + bounded tail만 운반"합니다. 증거를 버리는 것이 아니라, 인라인 출력 + 배너 재인용의 이중 소비(double-burn)를 제거합니다.

## 검증 증거 영속화 의무

파일-리다이렉트 계약이 `/tmp`에 기록한 증거는 OS에 의해 주기적으로 삭제됩니다(macOS 재부팅, Linux tmpfs 리마운트, systemd-tmpfiles). 인용된 경로가 더 이상 파일로 존재하지 않으면 감사 시점에 증거에 도달할 수 없습니다.

영속화 의무는 이 문제를 해결합니다. 검증 증거는 `.moai/state/verify/<session>/` 하위에 영속화되어야 합니다. 이 디렉터리는 `context-usage.json` 및 `active-sessions.json`과 같은 gitignored 런타임 상태 영역입니다.

정확한 영속화 메커니즘(직접 기록 또는 `/tmp` 기록 후 복사)은 구현 세부 사항입니다. 계약은 의무를 명시합니다. 증거는 `/tmp` 삭제 후에도 감사 시점에 도달 가능한 인용 가능한 경로에 남아 있어야 합니다.

## 다음 단계

- [토크노믹스 개요](/ko/advanced/tokenomics-overview/) — 4-층 구조 전체 개관
- [3-티어 에이전트 아키텍처](/ko/advanced/no-haiku-3tier/) — B층 라우팅의 모델 정책 기초

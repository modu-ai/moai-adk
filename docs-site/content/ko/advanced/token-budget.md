---
title: 토큰 예산 관리와 우아한 중단
weight: 2
draft: false
---

토크노믹스 4-층 구조의 D층인 예산 가드(Budget defense)를 자세히 다룹니다. 에이전트가 컨텍스트 윈도우 한계에 다다랐을 때 세션을 그냥 끊는 대신 진행 상태를 남겨 다음 세션이 이어받게 하는 것이 우아한 중단(graceful abort)입니다. 이 페이지는 그 메커니즘을 설명합니다.

## 예산 가드가 필요한 이유

Anthropic SSE 스트림은 컨텍스트 윈도우 천장에 가까워지면 `stream_idle_partial` 상태로 간헐적인 스톨을 일으킵니다. 간헐적이긴 해도 임계값을 넘어서면 충분히 예측할 수 있는 현상입니다. 스톨이 나면 에이전트 호출이 스트림 도중에 실패해 진행 상태를 잃을 수 있습니다.

예산 가드는 이 문제를 미리 막습니다. 컨텍스트 사용량이 임계에 닿기 전에 시스템이 먼저 우아한 중단을 수행하므로, 세션은 아무것도 잃지 않고 다음 단계로 넘어갑니다.

## 모델별 컨텍스트 임계치

운영 임계값은 모델마다 다릅니다. 윈도우가 크면 그만큼 높은 사용률까지 버틸 수 있고, 윈도우가 작으면 남는 여유 자체가 적습니다.

| 모델 클래스 | 윈도우 | 핸드오프 임계 | 절대 천장 |
|-------------|--------|---------------|-----------|
| Opus 5 (1M) | 1,000,000 토큰 | 50% | ~500,000 토큰 |
| GLM-5.2 (1M) | 1,000,000 토큰 | 50% | ~500,000 토큰 |
| Opus / Fable (256K) | 256,000 토큰 | 90% | ~230,000 토큰 |
| Sonnet / Opus 표준 (200K) | 200,000 토큰 | 90% | ~180,000 토큰 |
| Haiku (200K) | 200,000 토큰 | 90% | ~180,000 토큰 |

GLM-5.2(`moai glm` / `moai cg` GLM 패널)는 1M 컨텍스트 모델이므로 50% 임계로 운영합니다. Claude Code가 보고하는 `context_window_size`는 Claude 슬롯 기준(Opus=1M, Sonnet/Haiku=200K)이라, GLM 세션에서 원시 telemetry가 ~180K로 나와도 MoAI가 1M로 바로잡습니다. 믿을 값은 statusline의 CW% 게이지입니다.

## 2-단계 핸드오프 마커

statusline은 컨텍스트 바에 `/clear` 힌트를 2단계로 표시합니다.

- {{< icon warning warn >}} **소프트 마커** — statusline 컨텍스트 바에 `/clear` 힌트가 경고 색상으로 뜹니다. 밴드의 소프트 임계에서 나타나며, `/clear`를 실행할지는 사용자가 판단하면 되는 권고 신호입니다.
- {{< icon warning danger >}} **하드 마커** — statusline 컨텍스트 바에 `/clear` 힌트가 강한 경고 색상으로 뜹니다. auto-compact 인식 천장에서 나타나며, 다음 액션은 반드시 `/clear`여야 합니다.

하드 천장은 auto-compact 임계값 바로 옆에 잡히므로 런타임 auto-compact가 먼저 발동하는 일이 많고, 그래서 하드 마커는 실제로 드물게 뜹니다. auto-compact 인식 공식이 감수한 트레이드오프입니다.

## 우아한 중단 절차

SPEC-TOKEN-BUDGET-STOP-001로 구현한 우아한 중단 메커니즘은 다음 순서로 작동합니다.

1. **감지** — `Tracker.IsAtHardLimit(agentName)`이 true를 반환 (누적 사용량 ≥ hard_clear_threshold, 기본 0.90)
2. **상태 저장** — 진행 중인 작업 상태를 `progress.md`에 영속화
3. **핸드오프 발행** — 붙여넣기 가능한 resume 메시지를 생성 (6-블록 구조)
4. **턴 종료 권고** — 사용자에게 `/clear`를 권고 (HARD: 자동 `/clear`는 절대 수행 안 함)
5. **증거 영속화** — 검증 증거를 `.moai/state/verify/` 하위에 영속화

`/clear`는 절대 자동으로 실행되지 않습니다. 시스템은 권고만 하고, 실행 여부는 사용자가 판단합니다.

## paste-ready resume 6-블록 구조

세션 핸드오프 메시지는 다음 6-블록 구조를 따릅니다. 다음 세션이 최소한의 정보만으로 작업을 이어받도록 짠 구조입니다.

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
- **블록 3** — 구분선 + `전제 검증:` 헤더
- **블록 4** — 번호가 매겨진 검증 가능한 전제 항목 (최대 4개, 각 200자 이내)
- **블록 5** — `실행:` 단일 주요 액션 (일반적으로 `/moai <subcommand>`)
- **블록 6** — `머지 후:` 다음 액션 또는 SPEC ID

## 검증 다이어트 (verify-diet)

검증 명령이 쏟아내는 긴 출력을 디스크로 흘려보내고 컨텍스트에는 요약만 남기는 파일-리다이렉트 계약(file-redirect contract)입니다.

규칙은 이렇습니다. 검증 명령의 verbatim 출력이 **bounded-tail ceiling**(기본 50줄 또는 2KB 중 작은 쪽)을 넘어서면 출력을 파일로 리다이렉트하고, 컨텍스트에는 exit code와 bounded tail만 남깁니다.

```bash
go test ./... > /tmp/moai-verify/1-go-test.log 2>&1; echo "exit=$?"; tail -50 /tmp/moai-verify/1-go-test.log
```

이 계약은 "디스크에 verbatim 증거를 남기고, 컨텍스트에는 exit code와 bounded tail만 나른다"는 뜻입니다. 증거를 버리는 게 아니라 인라인 출력과 배너 재인용이 겹치는 이중 소비(double-burn)를 없애는 것입니다.

## 검증 증거 영속화 의무

파일-리다이렉트 계약이 `/tmp`에 남긴 증거는 OS가 주기적으로 지웁니다(macOS 재부팅, Linux tmpfs 리마운트, systemd-tmpfiles). 인용한 경로에 파일이 남아 있지 않으면 감사 시점에 증거를 확인할 방법이 없습니다.

영속화 의무가 이 구멍을 막습니다. 검증 증거는 `.moai/state/verify/<session>/` 아래에 남겨야 합니다. 이 디렉터리는 `context-usage.json`, `active-sessions.json`과 같은 gitignored 런타임 상태 영역입니다.

정확히 어떻게 남길지(바로 기록할지, `/tmp`에 쓴 뒤 복사할지)는 구현 세부 사항이고, 계약이 못 박는 것은 의무입니다. 증거는 `/tmp`가 비워진 뒤에도 감사 시점에 그대로 열어볼 수 있는 경로에 남아 있어야 합니다.

## 다음 단계

- [토크노믹스 개요](/ko/advanced/tokenomics-overview/) — 4-층 구조 전체 개관
- [3-티어 에이전트 아키텍처](/ko/advanced/no-haiku-3tier/) — B층 라우팅의 모델 정책 기초

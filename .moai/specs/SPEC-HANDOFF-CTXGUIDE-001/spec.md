---
id: SPEC-HANDOFF-CTXGUIDE-001
title: "256K 윈도우 핸드오프 안내 임계 결함 수정"
version: "0.1.0"
status: in-progress
created: 2026-07-03
updated: 2026-07-03
author: MoAI
priority: P1
phase: "v3.0.0"
module: "internal/statusline"
lifecycle: spec-anchored
tags: "statusline, handoff-guide, context-window, threshold, tier-s, epic-handoff-v2"
era: V3R6
related_specs: [SPEC-TOKEN-EFFICIENCY-001]
---

# SPEC-HANDOFF-CTXGUIDE-001 — 256K 윈도우 핸드오프 안내 임계 수정

> Epic "Handoff-v2" M1/4. 후속: M2 message-v2(오케스트레이션-모드 내장) · M3 auto-resume(handoff.yaml landing) · M4 threshold-guidance 완성(2-stage + config + 영속화). 본 SPEC은 M1로, **실사용 결함 1건의 최소 수정**에만 한정한다.

## §1 배경 · 목표 · 범위

### §1.1 배경 — 확인된 결함

statusline의 핸드오프 안내(`(⚠️/clear)` suffix)는 컨텍스트 윈도우 사용률이 모델별 임계를 넘으면 CW 바 뒤에 표시된다. 판정 함수 `shouldShowHandoffGuide`(`internal/statusline/renderer.go`, 약 571–589행)는 `data.Memory.ContextWindowSize`에 대해 **정확히 일치하는 switch**로만 동작한다:

```
switch cwSize {
case 1_000_000: return rawPct >= 50.0
case 200_000:   return rawPct >= 90.0
default:        return false
}
```

`rawPct = used*100/cwSize`. 이 구조 때문에 **1,000,000 또는 200,000이 아닌 모든 윈도우 크기**(예: 256,000 — 현재 non-1M Opus/Fable 클래스)는 `default: return false`로 떨어져, 사용률이 아무리 높아도 핸드오프 안내가 **영구히 표시되지 않는다**. 사용자는 256K 세션에서 `/clear` 시점 힌트를 전혀 받지 못한다.

추가로 `.claude/rules/moai/workflow/context-window-management.md` § Context Window Targets 임계표는 3행(Opus 1M / Sonnet-Opus 200K / Haiku 200K)뿐이며 **256K 행이 없다**.

### §1.2 목표

윈도우 크기에 무관하게 임계 밴드로 판정하도록 `shouldShowHandoffGuide`를 교체하여, 256K를 포함한 미등록 크기에서도 핸드오프 안내가 표시되게 한다. 기존 1M/200K 동작은 바이트 단위로 보존한다.

### §1.3 Out of Scope

이 §1.3은 spec-lint의 Exclusions 요구를 만족하는 h3 하위 절이다.

- **2단계 안내(🛑/clear!, 95%)**: M4 소관. 근거: `autoCompactThreshold` 기본 85 때문에 200K급은 raw 85% 근처서 런타임 auto-compact되어 95%가 대체로 도달 불가하므로, 2단계는 `min(hard, autoCompact+margin)` 계산과 함께 M4에서 다룬다. M1은 기존 단일 단계 `(⚠️/clear)` 렌더를 그대로 유지한다.
- **config 오버라이드 / `handoff.yaml` / `HandoffConfig`**: M3(landing)·M4(소비) 소관. M1은 밴드 경계값을 하드코딩한다(신규 config 섹션·로더 도입 금지).
- **`.moai/state/context-usage.json` 영속화**: M4 소관. M1은 상태 파일을 쓰지 않는다.
- **Detection Heuristics "state file first" 재작성**: M4 소관.

## §2 요구사항 (GEARS)

- **REQ-256K-001**: Where `shouldShowHandoffGuide`가 컨텍스트 윈도우 사용률을 판정할 때, the system shall 정확히-일치 switch 대신 **크기 무관 밴드 로직**을 사용해야 한다.
- **REQ-256K-002**: While 컨텍스트 윈도우 크기가 256,000일 때, When raw 사용률이 90% 이상이면, the system shall 핸드오프 안내를 표시해야 한다.
- **REQ-256K-003**: Where 윈도우 크기가 500,000 이상이면 the system shall 50% 임계를, 0 초과 500,000 미만이면 90% 임계를 적용해야 한다.
- **REQ-256K-004**: While 윈도우 크기가 0 이하일 때, the system shall 안내를 숨겨야 한다(원신호 부재 시 안전 기본값 보존).
- **REQ-256K-005**: Where 기존 1,000,000(≥50%) 및 200,000(≥90%) 동작은 the system shall 변경 전과 동일하게 보존해야 한다(회귀 금지).
- **REQ-256K-006**: When 본 수정이 적용되면, the system shall `context-window-management.md` 임계표에 256,000 윈도우 행(90% 임계)을 추가해야 한다.

## §3 인수 기준 (Tier S 인라인)

Given/When/Then 인수 시나리오. 테스트 대상: `internal/statusline/stdinfields_test.go`의 기존 `TestShouldShowHandoffGuide_*` 개별 함수 6개(L31–116). 신규 256K/500K/0 케이스는 **같은 파일**에 추가(개별 함수 옆 또는 테이블 주도 전환 — `renderer_test.go` 분산 금지, plan-auditor D1).

| AC | Given (상태) | When (조건) | Then (기대) |
|----|--------------|-------------|-------------|
| AC-256K-001 | ContextWindowSize=256000 | rawPct=90 | `shouldShowHandoffGuide == true` |
| AC-256K-002 | ContextWindowSize=256000 | rawPct=89 | `== false` |
| AC-256K-003 | ContextWindowSize=1000000 | rawPct=50 / 49 | `true` / `false` (기존 보존) |
| AC-256K-004 | ContextWindowSize=200000 | rawPct=90 / 89 | `true` / `false` (기존 보존) |
| AC-256K-005 | ContextWindowSize=500000 / 499999 | rawPct=50 | `true`(대형 밴드) / `false`(90% 밴드 필요) |
| AC-256K-006 | ContextWindowSize=0 | rawPct=any | `== false` |
| AC-256K-007 | 문서 | — | `context-window-management.md`에 `256,000` + `90%` 행 존재(grep) |

- **AC-256K-008 (회귀)**: `go test ./internal/statusline/`가 전부 통과한다(기존 렌더 테스트 무손상).
- **AC-256K-009 (크로스플랫폼)**: `go build ./...` 및 `GOOS=windows GOARCH=amd64 go build ./...` 모두 exit 0.

## §4 접근

`shouldShowHandoffGuide`의 switch를 다음으로 교체:

```
if cwSize <= 0 { return false }
if cwSize >= 500_000 { return rawPct >= 50.0 }
return rawPct >= 90.0
```

근거: 256K는 200K와 절대 여유(headroom)가 유사하므로 90% 밴드에 속한다. 500K 컷오프는 1M 대형 윈도우 클래스(50%)와 표준/중형 클래스(90%)를 깔끔히 분리한다. `@MX:NOTE`의 1M=50%/200K=90% 문구는 밴드 표현으로 갱신한다.

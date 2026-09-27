---
id: SPEC-SERVED-MODEL-NORM-001
title: "서빙 모델 판정의 선언 정규화 — [1m] 접미와 inherit 선언을 drift 로 세지 않기"
version: "0.1.0"
status: draft
created: 2026-09-28
updated: 2026-09-28
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/hook"
lifecycle: spec-anchored
tags: "hook, served-model, drift, normalization, auditor-gate"
tier: S
era: V3R6
related_specs: [SPEC-SERVED-MODEL-AUDIT-001]
---

# SPEC: 서빙 모델 판정의 선언 정규화

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-28 | manager-spec | 최초 draft. 카드 t1287 — SPEC-SERVED-MODEL-AUDIT-001 sync-audit 부채 F1(`[1m]` 접미 미정규화)·F6(`inherit` 선언 상시 drift). |

## 배경

`expectedServedModel` 은 선언 모델에서 공백만 잘라 그대로 기대값으로 썼다. 응답 행의 `model` 값에는 컨텍스트 창 접미(`[1m]`)가 붙지 않으므로 `claude-opus-5[1m]` 선언은 `claude-opus-5` 서빙과 늘 어긋났고, `opus[1m]` 은 숫자가 섞여 별칭 비교 경로에서도 빠졌다. `inherit` 선언은 부모 세션의 모델을 뜻하지만 훅은 그 값을 볼 수 없어, 문자열 `inherit` 과 비교한 결과가 늘 drift 였다. 로컬에서는 `workflow.served_model_gate.enabled` 가 켜져 있으므로, `opus[1m]` 을 선언한 감사관이 실제로 Opus 에서 서빙돼도 판정 채택이 거부된다.

## 요구사항 (GEARS)

- **REQ-SMN-001** — 선언 모델 끝에 대괄호 접미(`[...]`)가 있으면, 기대 모델을 계산하기 전에 그 접미를 제거해야 한다(SHALL). 접미를 제거한 뒤의 비교 규칙(대소문자 무시 동등, 별칭 계열 일치)은 바뀌지 않는다.
- **REQ-SMN-002** — 선언 모델이 `inherit` 이면, 선언이 없을 때와 똑같이 프로필 해석 모델로 기대값을 정해야 한다(SHALL). 해석 모델이 없으면 `unmapped`(설정이 없으면 `unknown`)로 기록하며, `served_drift` 로 기록해서는 안 된다(SHALL NOT).
- **REQ-SMN-003** — 정규화가 진짜 drift 를 가려서는 안 된다(SHALL NOT). `opus[1m]` 선언에 `glm-*` 서빙은 여전히 `served_drift` 다.

## 1. Scope

### 1.1 Out of Scope

- 프로필 해석 모델(`resolveAgentModel`)의 접미 처리 — 착수 측정(§A)에서 해석 모델에 접미가 붙은 사례가 없어 손대지 않는다.
- 게이트 로직 — 게이트는 관측 판정(`Verdict`)만 읽으므로 분류를 고치면 게이트에도 그대로 반영된다.

## §A 착수 측정 (develop `37dc766b9`)

이 트리로 빌드한 바이너리에서 `moai doctor --check "Served Model" --verbose` 실행: `swept 2334 subagent transcripts: ok 1295, served_drift 982, unknown 11, unmapped 46`. drift 가운데 `expected=claude-opus-5[1m] served=[claude-opus-5]` 295건, `expected=claude-fable-5-1[1m] served=[claude-fable-5-1]` 2건, `expected=inherit served=[claude-sonnet-5]` 1건으로 오탐이 298건이다. `expected=opus[1m] served=[glm-5.3-flash]` 125건은 진짜 drift 다. t1282 시점의 292건에서 늘었다.

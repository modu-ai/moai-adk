---
id: SPEC-AC-LOCALE-TOKEN-001
title: "ASCII 토큰 계수 검증 기준의 로케일 왜곡 제거 — AC 전수 조사 및 재작성"
version: "0.1.0"
status: completed
created: 2026-09-08
updated: 2026-09-08
author: manager-spec (card t573)
priority: P1
phase: "v3.2.0"
module: ".moai/specs"
lifecycle: spec-anchored
tags: "acceptance-criteria, locale, grep, census, docs-site"
tier: M
depends_on:
  - SPEC-DOCS-LOCALE-PARITY-REPAIR-001
---

# SPEC — ASCII 토큰 계수 검증 기준의 로케일 왜곡 제거

## HISTORY

- 2026-09-08: 카드 t573에서 plan-phase 작성. 기원은 t538 (SPEC-DOCS-LOCALE-PARITY-REPAIR-001) sync-audit 발견 F2 — 리드가 카드로 승격. 계열 뿌리(같은 acceptance.md 에서 이미 수리된 AC-008·AC-011과 동일 — "검증식이 말하는 대상을 재지 못하는 기준")에 대한 세 번째 결함형이다.

## §1 배경과 문제 정의

t538 의 AC-004 는 "en·zh desktop-native 총 히트 ≥4 (각)"을 `grep -ci "desktop-native"` 로 검증했다. zh 문서(`docs-site/content/zh/utility-commands/moai-e2e.md`)의 작성자는 기준선을 채우기 위해 산문과 표 행 라벨에 ASCII 괄호 보강(`**原生桌面 (macOS, desktop-native)**` 식)을 붙였다. 측정 결과(트리 `3ac58b5a1`, 본 worktree, 2026-09-08):

- zh 의 `desktop-native` 출현 수: **8** (행 기준 7) — 플래그 행 :58 에 2(코드 토큰 — 정당), 매트릭스 라벨 :78-80 에 3(ASCII 보강 — 기준 충족 목적), 산문 보강 :84·:98·:176 에 3(동일).
- 대상이 아니던 로케일은 자연스럽게 남았다: ja 2, ko 2 (플래그 행뿐), ja 매트릭스 라벨은 순수 원어(`**デスクトップネイティブ (macOS)**`). en 은 4 로 자연 산출.
- **왜곡은 zh 에만 격리돼 있다** — ASCII 계수 기준이 겨냥한 유일한 비(非)Latin 로케일만 쓰기 방식이 바뀌었다.
- 재측정으로 정리된 불일치: 리드 진술 "전면 되돌려도 4 건 남아 기준 여전히 충족"은 **거짓**이다. 실측 — 전면 보강 제거 시 출현 2(플래그 행뿐, `≥4` 미달), 라벨만 제거 시 5. 레인의 재측정 수치가 맞다.

결함의 소재는 zh 표기가 아니라 **기준**이다. zh 보강 표기는 남겨도 되고(판단 가능), 기준을 로케일-확정 토큰이나 구조 기준으로 다시 쓰는 것이 수리 방향이다.

이 결함형의 일반식: **비(非)Latin 로케일 문서에 대해 특정 ASCII/Latin 토큰의 출현 수(또는 존재)를 계수하는 검증 기준은, 그 수를 채우기 위해 문서의 쓰기 방식을 ASCII 쪽으로 왜곡시킨다.**

## §2 GEARS 요구사항

- REQ-001 (Ubiquitous): 본 SPEC 이 채택하거나 재작성하는 모든 문서 검증 기준은 다음 네 요소를 함께 진술한다: (a) 정확한 계수 명령 — 출현 수를 세려는 경우 `grep -o … | wc -l` 이며 행 계수가 의도인 경우에만 `grep -c`, (b) 계수 대상 토큰이 속한 로케일, (c) 실측 현재값에서 도출된 기준선, (d) 그 기준이 문서 스타일을 왜곡하는 방식으로는 충족될 수 없다는 근거.

- REQ-002 (When): **When** M1 census 가 ASCII/Latin 토큰의 출현 수·존재를 비(非)Latin 로케일(ko/ja/zh) 문서(docs-site/content/ 의 슬래시·중괄호 철자, README 로케일 파일 등)에 대해 계수하는 검증 기준을 발견하면, the census shall 각 건을 — 파일, AC id, 토큰, 대상 문서+로케일, 계수 의미(행 vs 출현), 현재 히트 vs 기준선, 판정(distorts / benign[토큰이 코드 식별자이고 계수식이 코드 위치로 한정될 때만] / n/a) — 필드와 함께 장부(`.moai/reports/t573/census.md`)에 기록한다. 코퍼스는 `.moai/specs` 의 acceptance.md·spec.md·plan.md 전부다(기준을 진술하는 표면); `progress.md` 는 동기-단계 증거 기록으로서 구속력 있는 기준이 아니므로 제외한다 — 이 경계는 의도다.

- REQ-003 (When): **When** census 가 왜곡형(distorts)으로 확정한 기준이 있으면, the lane shall 그 기준을 로케일-확정 토큰(예: zh 원어 용어) 또는 구조 기준(표 행 수, 앵커 수, 절 존재)으로 재작성하고, 기준선을 실측 현재값으로 재진술하며, t538 규율을 따르는 날짜형 HISTORY 항목(before/after 명령 + 실측값 + 승인 귀속)을 남긴다.

- REQ-004 (Ubiquitous): SPEC-DOCS-LOCALE-PARITY-REPAIR-001 의 AC-004 는 재작성 최소 대상이다. 권장 형태(본 트리 실측에 근거): zh 축은 원어 토큰 `原生桌面`(전면 보강 제거 후에도 6행 유지 — 실측) 또는 매트릭스 행 수(`^| \*\*原生桌面` 3행)로, en 축은 Latin 로케일이므로 ASCII 토큰 기준을 유지해도 왜곡이 없다.

- REQ-005 (While): **While** M3 이 의사결정-플래그 상태이면, the lane shall docs-site 콘텐츠를 변경하지 않고, 재작성된 기준에서 "보강 유지 시 / 전면 되돌림 시" 두 경우의 기준값을 연산해 운영자 게이트에 선택지로 제시한다.

- REQ-006 (When): **When** census 가 계열 근접 미스(계수 단위 불일치 — t538 AC-008 형, 무한-증가 도메인 vs 고정 기준선 — t538 AC-011 형)를 발견하면, the census shall 이를 2차 관찰로 기록만 하고 수정하지 않는다.

- REQ-007 (Ubiquitous): 본 SPEC 의 모든 측정 수치는 그것을 산출한 명령과 출력, 측정 트리 SHA 를 함께 기록한다. 타 트리·타 시점의 수치를 새 근거로 인용하지 않는다.

## §3 제약

- 문서 전용 SPEC — Go 코드 변경 없음. 하네스: **standard** (docs-only 이지만 census 범위 판정이 100+ 파일(A 합집합 plan-시 161)에 걸쳐 판정 품질이 기준선을 좌우).
- 아티팩트 언어: 한국어 (t538 레지스터 일치).
- 측정 grep 은 `/usr/bin/grep` (셸 grep 은 ugrep 래퍼 — 조용히 건너뛴다, t538 REQ-012 계승).
- 본 plan-phase 에서는 SPEC-DOCS-LOCALE-PARITY-REPAIR-001 의 acceptance.md 를 수정하지 않는다 — 재작성은 본 SPEC 의 run phase(M2)에서 일어난다.
- census 명령은 `SPEC-AC-LOCALE-TOKEN-001` 디렉터리를 제외하고, 계수의 트리 기준(작업 트리, base `3ac58b5a1`)을 함께 기록한다 — 자기 아티팩트 착지로 계수가 표류하는 것을 막는다 (t538 AC-011 동형).

## §4 Out of Scope

### Out of Scope — docs-site 콘텐츠 변경 (M3 실행)

- zh 문서(`docs-site/content/zh/utility-commands/moai-e2e.md`)의 ASCII 보강 표기 되돌리기(:78-80, :84, :98, :176)는 M3 의사결정-플래그 항목으로, 운영자 게이트 승인 없이는 실행하지 않는다. 본 SPEC 은 기준 재작성만 실행한다.

### Out of Scope — 근접 미스 수리

- census 에서 발견되는 계수 단위 불일치(t538 AC-008 형), 무한-증가 도메인 vs 고정 기준선(t538 AC-011 형) 결함은 2차 관찰로 기록만 한다. 이미 같은 파일에서 수리된 항목의 재수리도 하지 않는다.

### Out of Scope — 검증 도구 코드화

- "왜곡 불가 기준" 작성 규칙을 lint/gate 로 기계화하는 것(REQ-001 의 Go 구현)은 후속 SPEC 소관이다. 본 SPEC 은 문서 기준의 조사와 재작성만 한다.

### Out of Scope — 거버넌스 문서 신설

- 검증식 작성 가이드 문서 신설은 하지 않는다. REQ-001 의 네 요소는 본 SPEC 의 AC 와 HISTORY 규율로 강제된다.

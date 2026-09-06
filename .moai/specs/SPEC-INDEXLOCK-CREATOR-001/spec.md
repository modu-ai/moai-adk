---
id: SPEC-INDEXLOCK-CREATOR-001
title: "index.lock Contention Creator Identification — Failure-Instant Capture Attribution"
version: "0.1.0"
status: completed
created: 2026-09-04
updated: 2026-09-04
author: lane-15
priority: High
phase: "v3.2.0"
module: ".moai/reports/t485 (observation evidence); internal/statusline + internal/core/git (read-only analysis)"
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "index.lock, contention, statusline, git-status, capture, observation, t485"
related_specs:
  - SPEC-UPDATE-HOOK-DELIVERY-001
---

# SPEC-INDEXLOCK-CREATOR-001 — index.lock 경합 생성자 확정 (카드 t485)

> Investigation SPEC — 확정(identification)이 산출물이다. 수리(repair)는 확정 이후
> 별도 판단이며 이 SPEC의 범위 밖이다(운영자/리드 지시, 배차문 2026-09-04).

## §A. Problem Statement

`git` 인덱스 쓰기 실패 메시지 중 **경로가 붙은 형태**
(`fatal: Unable to create '<abs>/index.lock': File exists.`)는 락 **획득** 경합이며,
 lane-4 의 분류(리드 장부 2026-09-03T18:34Z 회차 §③)에 따라 1회 재시도로 우회되고
 있다. 그러나 **그 순간 락을 실제로 쥐고 있던 프로세스는 11사례 누적에도 한 번도
 포착되지 않았다** — 리드 Gaps에 "생성자 미확인(11사례 누적)"으로 반복 기록됐다.

후보(배차문): statusline 의 `status` 호출 · 리드의 회차 스윕 · git 자체의 내부 갱신 ·
 다른 레인의 동시 조작. 종전 최유력 가설은 statusline
 (`git status --porcelain` 이 `--no-optional-locks` 면제 없이 호출된다는 코드 관측,
 메모리 `feedback-statusline-status-call-takes-an-index-write-lock`)이나, 이것도
 실패 순간의 프로세스 캡처로는 확인된 바 없다(리드 2026-09-03: "statusline 추정은
 정황이고 실패 순간 프로세스를 못 잡았다").

부수 미확정: 원인 A(`core.fsmonitor=/dev/null` 창)의 주체. 영구 config 4지점
 (worktree/global/local/config.worktree) 어디에도 값이 없어, **런타임 주입**
 (`GIT_CONFIG_*` env 또는 `git -c`) 가능성이 남아 있다 — config 파일 추적으로는
 원인 A 의 주체를 못 잡을 수 있다(배차문 부수 확인).

## §B. Goals

1. 경합 계열 `index.lock` 의 라이브 홀더를 **실패-순간 동시 캡처**(lock 파일 존재
   관측 + `lsof` 홀더 PID + `ps` 부모 추적)로 확정하고, 프로듀서(어느 컴포넌트가
   어떤 코드 경로로 락을 잡는지)까지 귀속한다.
2. 원인 A 부수 확인: `core.fsmonitor=/dev/null` 의 주입 축을 판정하고, config 파일
   추적 축의 viability 를 근거와 함께 결론낸다.

## §C. Approach (요약)

- **git-free 수동 관측기**: git 호출을 하나도 하지 않는 폴러(stat glob 만 사용)가
  전 워크트리의 `index.lock` 존재를 감시한다. 관측기 자신은 경합을 만들 수 없다.
- 라이브 홀더 포착 시 `lsof <lock>` 으로 홀더 PID 확정 → 즉시 `ps` 버스트(3회)로
  홀더 명령·부모 2단 추적 + 필터 ps 스냅샷.
- 코드 표면 판독으로 프로듀서의 코드 경로를 현재 트리에서 재검증(file:line).
- fsmonitor: 소스·템플릿·훅·config·env 전 축 grep + 라이브 레인 env 점검(키 이름만).

## §3. Acceptance Criteria

| AC | 내용 | 검증 방법 |
|---|---|---|
| AC-ILC-001 | 라이브 `index.lock` 홀더가 `lsof` 로 포착되고, 홀더 PID 의 전체 명령과 부모 체인이 ps 증거로 **명명된 프로듀서**까지 귀속된 사례가 ≥1건 존재한다 | `summary.md` 실홈 행 + `holder-*.txt`/`ps-*.txt` 원본 |
| AC-ILC-002 | 그 프로듀서의 코드 경로가 이 트리에서 file:line 으로 검증된다 — 인덱스에 쓸 수 있는 git 호출과 `GIT_OPTIONAL_LOCKS`/`--no-optional-locks` 면제 부재 | 소스 직독 인용 + `grep -rn "OPTIONAL_LOCKS"` 0히트 출력 |
| AC-ILC-003 | 캡처 방법이 git-free 이며 자가종료+외부 timeout 이중 경계임이 문서화돼 있다 | `watch-indexlock.sh` 본문 + 실행 로그(watch.log) |
| AC-ILC-004 | fsmonitor 주입 축 판정이 근거(전 축 grep 히트 수 + 라이브 env 키 점검)와 함께 내려져 있다 | verdict.md 부수 확인 절 |
| AC-ILC-005 | 결론 보고(verdict.md)가 5섹션 형식이며 모든 주장이 관측 출력에 귀속된다 | verdict.md 본문 |

### 3.1 Out of Scope

- 수리·수정: `GIT_OPTIONAL_LOCKS=0` 부착 등 **어떤 코드 변경도 하지 않는다** —
  확정 이후 별도 카드.
- 부하 기반 재현: 배경 부하 생성 금지(배차 [HARD]). 관측은 기존 활동에 대한
  수동 감시만으로 한다.
- 원인 B(경로 없는 쓰기 실패, `Unable to write index.`)의 규명 — 본 카드는
  경합 계열(경로 있음)만 담는다. 원인 B 는 별도 소관.
- 배포 템플릿(`internal/template/templates/**`) 변경 — 이 SPEC 은 관측·분석만
  수행하며 산출물은 SPEC 문서와 증거다.

## §D. Evidence & Data

- 관측 raw: `.moai/reports/t485/raw/` (events.jsonl, lsof-*.txt, ps-*.txt,
  holder-*.txt, fps-*.txt, watch.log)
- 집계: `.moai/reports/t485/raw/summary.{md,jsonl}` (summarize.py)
- 결론: `.moai/reports/t485/verdict.md`

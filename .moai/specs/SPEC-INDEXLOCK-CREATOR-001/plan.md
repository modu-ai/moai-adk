---
id: SPEC-INDEXLOCK-CREATOR-001
title: "index.lock Contention Creator Identification — Failure-Instant Capture Attribution"
version: "0.1.0"
created: 2026-09-04
updated: 2026-09-04
author: lane-15
priority: High
phase: "v3.2.0"
module: ".moai/reports/t485; internal/statusline + internal/core/git (read-only)"
lifecycle: spec-anchored
tags: "index.lock, contention, statusline, capture, t485"
---

# plan.md — SPEC-INDEXLOCK-CREATOR-001 (카드 t485)

Tier S 조사 SPEC. 산출물 = 관측 증거 + 코드 귀속 + 결론 보고. 코드 변경 없음.

## Milestones

### M1 — 수동 관측기 구축·가동 (AC-ILC-003)
- git 호출 0개 폴러: stat glob(primary + `.git/worktrees/*/index.lock`) 0.12초 간격,
  카운터 자가종료 + 외부 `timeout` 이중 경계.
- 라이브 홀더 포착 시: `lsof <lock>` → 홀더 PID, `ps` 3회 버스트(부모 2단 추적),
  필터 ps 스냅샷(`fps-*.txt`).
- 산출: `watch-indexlock.sh`, `events.jsonl`, `watch.log`, `lsof/ps/holder/fps` 원본.

### M2 — 기존 증거 흡수 + 코드 표면 판독 (AC-ILC-002, AC-ILC-004)
- 리드 장부(`.moai/reports/lead/card-status-20260902.md`) 11사례 계보·lane-4 3형태
  분류·lane-14 캡처 흡수. lock-sweep 보고서(2026-08-28) 흡수 — 죽은 잠금 3건의
  mtime 인접(09:06:56/09:07:55/09:07:57) = 동시 죽음 파동 서명 확인.
- 코드 직독(현재 트리): `internal/statusline/git.go` → `internal/core/git/manager.go`
  `Status()` 의 `git status --porcelain` 호출, `GIT_OPTIONAL_LOCKS` 0히트 재확인.
- fsmonitor 전 축 grep: Go 소스·템플릿·훅·status_line.sh·전역/로컬 config·
  config.worktree·`.env.glm`·러처 코드·라이브 레인 env(키 이름만).

### M3 — 집계·귀속·결론 (AC-ILC-001, AC-ILC-005)
- `summarize.py` 후처리: REAL/GONE 분류, 홀더→부모 귀속 집계.
- 결론 보고 `verdict.md` (5섹션 형식).

## Verification

- AC-ILC-001: summary.md 의 실홈 행 + holder/fps 원본 대조.
- AC-ILC-002: 소스 인용 file:line + `grep -rn "OPTIONAL_LOCKS" internal/ pkg/ cmd/` 출력.
- AC-ILC-004: 축별 grep 히트 수 표.
- AC-ILC-005: verdict.md 셀프 점검(모든 PASS 행에 관측 출력 존재).

## Constraints

- 배경 부하 생성 금지(배차 [HARD]) — 관측기는 git 호출 0개, 읽기 전용.
- push 금지, 워크트리 삭제 금지, 커밋마다 `t485` 명시.
- 수리 금지 — 확정만.

# progress.md — SPEC-TOOLPOLICY-DRIFT-GUARD-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-10
artifacts: spec.md + plan.md + acceptance.md + progress.md (Tier M) + spec-compact.md — v0.1.2
card: t619
baseline: worktree `.claude/worktrees/t619`, branch `WT-toolpolicy-drift`, plan 작성 HEAD `c7b8d110b`, 1회차 감사 HEAD `b2cbfd207`, 2회차 감사 HEAD `dc220b8fe`, `git merge-base HEAD develop` → `d1b61005d20967fdbd970ec7ec734c6d14f29dc3`
evidence_base: `.moai/reports/t619/verdict.md`
spec_id_check: `[[ "SPEC-TOOLPOLICY-DRIFT-GUARD-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS` (실행 출력), `ls .moai/specs | grep -c TOOLPOLICY-DRIFT-GUARD` → `0` (작성 전)
open_clarifications: 0 — 운영자 결정 6건(2026-09-10) 반영. 수리 방향·검사 위치·주장 정정(레인 세션 직접 수령), 집합 비교·주장 정정 전수·워킹 트리 판독(plan 제안 검토 후)
plan_measurements: 스크래치 build `allow=108 ask=0 deny=60 env_gated_skipped=5`, `diff` 종료 1 / 4 헝크 / 160 vs 166 줄, 커밋본 목록 `allow-NOT-C-sorted` `deny-NOT-C-sorted`, 커밋본 중복 0 / allow·deny 겹침 0

### plan-audit 1회차 (2026-09-10) — FAIL 0.71 대응

- 보고서: `.moai/reports/plan-audit/SPEC-TOOLPOLICY-DRIFT-GUARD-001-review-1.md`
- blocking D1·D2·D3 반영, optional D4-D10 은 반영 또는 부분 반영(D8). 결함별 처분 표는 plan.md §H.
- 수리 회차 실측(HEAD `b2cbfd207`): env_gate 없음 + 네 경로 deny 질의 `12`, env-gated Write deny `1`, env_gate 항목 `5`, `ANALOGOUS drift class` `1`, `comment (audit surface) are generated` `1`, `tool-policy-drift-check` 두 파일 모두 `0`, `.claude/settings.json#permissions` `171`, `make -n build` 의 `TestToolPolicyDrift_` `0` / `TestGoldenCommittedArtifactsMatchEmission` `2`, `go_code:` 블록 14줄·`.moai/**` `1`·`.claude/settings.json` `0`, merge-base 기준 settings 두 파일 diff 출력 없음.
- D8 관련 가드 실측: `trap` 포함 명령, heredoc 으로 python/bash 에 스크립트를 넘기는 명령, 변수로 계산된 스크립트 경로를 python 에 넘기는 명령이 모두 워크트리 세션 가드에 거부됨. 리터럴 절대 경로의 스크래치 스크립트 호출은 통과.

### plan-audit 2회차 (2026-09-10) — FAIL 0.79 대응, 운영자 승인 3회차

- 보고서: `.moai/reports/plan-audit/SPEC-TOOLPOLICY-DRIFT-GUARD-001-review-2.md`
- Tier M 상한(2회) 도달. 운영자가 수정 후 3회차 감사 1회를 승인했다.
- blocking D11·D13·D14, optional D12·D15-D20 모두 반영. 처분 표는 plan.md §H 2회차.
- 수리 회차 실측(워크트리 t619, 대상 파일 무수정 상태):
  - env_gate 항목 정렬 JSON sha256 `5e0cba521c5c81a2d7bb82fbba52c59b027d2329b2c0bf8b912ed6b2150e6d27`, 항목 `5`
  - `grep -c -E '^build:.*tool-policy-drift-check' Makefile` → `0`, `^build:.*agents-emit-check` → `1`
  - 변경 대상 기존 네 파일 역슬래시-u / Cf 계수 모두 `(0, 0)`, 대조 `(1, 1)`
  - 조건부 복원 사슬(리터럴 경로, 스크래치 사본): 일치 시 `restored=yes`, 끼어든 수정이 있으면 `restored=NO_stop` 이고 수정 보존(`grep -c` `1`). 워크트리 가드에서 실행됨
  - `git status --porcelain --untracked-files=all > <스크래치 파일>` 워크트리 가드에서 실행됨
- 재감사(3회차) 대기.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

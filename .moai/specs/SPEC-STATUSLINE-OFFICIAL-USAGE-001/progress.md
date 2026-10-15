# SPEC-STATUSLINE-OFFICIAL-USAGE-001 — progress

plan→run Kickoff (autonomous, auto-semantics §9.1)이 아래 차감을 **구속 조건**으로 복사함 (2026-10-14T23:30Z, lane-35/t1732, run tmvgvn). 두 sync 판정 소유자가 재독하며, 미처리 조건은 sync 판정을 FAIL로 상한합니다. 판정 출처: `.moai/reports/t1732/plan-audit-iter2.md` (PASS-WITH-DEBT 0.87, must_pass_failed 0, blocking 0, audited_sha a52a0e5c9, plan_artifact_hash e960c02a…837fe).

## Binding run conditions

> 갱신 (board d-20261014T235405Z-8b55 조건 ②, 2026-10-15T00:14Z): 아래 두 차감은 `f87f31dfd`에서 plan.md M4 스코프로 흡수되어 더 이상 run 구속 조건이 아닙니다. 처분 위치는 다음 AC이며, sync-audit이 이 AC들로 처분을 확인합니다.

- ~~sou-debt-1~~ → **AC-SOU-010** (acceptance.md §D; REQ-SOU-006): mid-read abandonment 핀 — plan.md §C M4 (blocked/slow fixture, 읽기 시작 후 취소) + §B D8 (가드 위치).
- ~~sou-debt-2~~ → **AC-SOU-009** (acceptance.md §D, release-blocking; REQ-SOU-003): relative `gitdir:` 포인터 케이스 — plan.md §C M4 (relative-pointer fixture) + §B D7 (Clean+Join 해소).

원 기록 (2026-10-14T23:30Z Kickoff가 복사, 이력 보존): sou-debt-1/2 원문은 plan-audit-iter2.md 10-11행 및 git 이력(17c3e8f72의 progress.md) 참조.

## §E.1 Plan-phase Audit-Ready Signal

audit_ready: true
plan_complete_at: 2026-10-14T23:43:52Z
plan_audit: PASS-WITH-DEBT 0.87 (iteration 2) — .moai/reports/t1732/plan-audit-iter2.md

## §F Phase 4 Mode Selection

- Input: tier M · scope ~10 files (usage.go·builder.go·renderer.go·types·tests·CLI comment·template verify) · domains 1-2 (statusline pkg + template mirror) · file mix Go+markdown · concurrency benefit LOW (coding-heavy)
- Mode evaluation: direct — not trivial, no / serial — **selected** / fanout — no (coding-heavy, Anthropic caveat) / sweep — no (not ≥30-file mechanical uniform)
- Decision: serial
- Justification: coding-heavy single-package implementation with a TDD cycle — per Anthropic's coding-task parallelism caveat the sequential path is the safe default; one writer per tree (factory lane discipline); milestones M1-M5 sequence through one manager-develop delegation with incremental commits. Kickoff gate met in its autonomous form (§9.1: PASS-WITH-DEBT 0.87, must_pass_failed 0, blocking 0, hash-bound, decision record on the card progress).

## §G Override and Refusal Record

- 2026-10-14T23:45:39Z SPEC-STATUSLINE-OFFICIAL-USAGE-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 2 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1732/.moai/reports/t1732/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1732/.moai/reports/t1732/plan-audit.md

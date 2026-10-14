# SPEC-STATUSLINE-OFFICIAL-USAGE-001 — progress

plan→run Kickoff (autonomous, auto-semantics §9.1)이 아래 차감을 **구속 조건**으로 복사함 (2026-10-14T23:30Z, lane-35/t1732, run tmvgvn). 두 sync 판정 소유자가 재독하며, 미처리 조건은 sync 판정을 FAIL로 상한합니다. 판정 출처: `.moai/reports/t1732/plan-audit-iter2.md` (PASS-WITH-DEBT 0.87, must_pass_failed 0, blocking 0, audited_sha a52a0e5c9, plan_artifact_hash e960c02a…837fe).

## Binding run conditions

- debt: sou-debt-1 dispose_in=run Strengthen M4's abandonment test: cancel the render context AFTER the fallback read starts (blocked/slow fixture), not only pre-cancelled — the pre-cancelled-only pin admits an entry-check mutant (`if ctx.Err() != nil { return }` before launching) that passes AC-SOU-010 while violating REQ-SOU-006's mid-read abandonment (codex iter-2 finding 2, independently verified statically).
- debt: sou-debt-2 dispose_in=run Extend D7's `quotaScanStateDirs` to resolve a RELATIVE `gitdir:` pointer (clean + join with projDir) and add a relative-pointer fixture case to M4 — git accepts relative pointers as valid worktree anchors, and REQ-SOU-003's universal sentence ("a render anchored at a linked-worktree root shall read the same record population") currently out-scopes R5's fail-open n/a degradation for exactly that population (codex iter-2 finding 1, independently verified).

## §E.1 Plan-phase Audit-Ready Signal

audit_ready: true
plan_complete_at: 2026-10-14T23:43:52Z
plan_audit: PASS-WITH-DEBT 0.87 (iteration 2) — .moai/reports/t1732/plan-audit-iter2.md

## §F Phase 4 Mode Selection

- Input: tier M · scope ~10 files (usage.go·builder.go·renderer.go·types·tests·CLI comment·template verify) · domains 1-2 (statusline pkg + template mirror) · file mix Go+markdown · concurrency benefit LOW (coding-heavy)
- Mode evaluation: direct — not trivial, no / serial — **selected** / fanout — no (coding-heavy, Anthropic caveat) / sweep — no (not ≥30-file mechanical uniform)
- Decision: serial
- Justification: coding-heavy single-package implementation with a TDD cycle — per Anthropic's coding-task parallelism caveat the sequential path is the safe default; one writer per tree (factory lane discipline); milestones M1-M5 sequence through one manager-develop delegation with incremental commits. Kickoff gate met in its autonomous form (§9.1: PASS-WITH-DEBT 0.87, must_pass_failed 0, blocking 0, hash-bound, decision record on the card progress).

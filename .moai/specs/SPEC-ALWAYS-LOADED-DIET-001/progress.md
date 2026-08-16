# Progress — SPEC-ALWAYS-LOADED-DIET-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- tier: M (3 artifacts: spec.md + plan.md + acceptance.md)
- REQ count: 16 / 16 (Tier M ceiling)
- AC count: 16 / 16 (Tier M ceiling)
- open items: none — the 3 former plan.md §D4 open questions are resolved by user decision (budget ratchet deferred to a separate backlog card; M3 control is documentation-only, no Go change; growth-statement threshold = 1,000 bytes per single edit). See plan.md §D4-1..D4-3.
- baseline observed (2026-08-16): `files=14 bytes=295044 tokens=73761 headroom=1239`; `go test ./internal/config/ -run TestAlwaysLoaded` → ok
- plan_complete_at: 2026-08-17 (산출물 커밋 be1958a4d, 2026-08-17 00:32 +0900 — 실제 실행일)
- plan_audit: iter1 FAIL 0.73 (8 blocking) → iter2 **PASS 0.845** vs Tier M threshold 0.80 (Clarity 0.85 / Completeness 0.88 / Testability 0.75 / Traceability 0.92); re-audit ceiling (Tier M = 2) reached, no iter3
- plan_audit reports: `.moai/reports/plan-audit/SPEC-ALWAYS-LOADED-DIET-001-review-{1,2}.md`
- iter2 blocking 3 applied orchestrator-direct (auditor's own recommendation — local shell edits, ~15 lines):
  - D1 `AC-ALD-009` passed with no companion (`wc -c` failure → empty → bash arithmetic 0 → `sum` equals the original 21,003, which is exactly the PASS lower bound). Fixed with `test -f` + `companion >= 1`. Re-run on the untouched tree → `MISSING …detail.md`, exit 1.
  - D2 six AC touched files with no existence guard, breaching this document's own §A trap rule 6. `AC-ALD-006` actually PASSed (`missing_lines=0`). Guards added to AC-ALD-004/005/006/008/009/013; `AC-ALD-006` re-run → exit 1.
  - D3 `REQ-ALD-013` claimed all four guard slots while the glob covered three. `,**/MEMORY.md` added (13 chars, zero always-loaded cost) in spec.md §3.3 + plan.md D1/M3. Rationale: the guard counts that slot conditionally, so a future repo-root `MEMORY.md` would admit up to `memoryHeadByteCap` 25,600 B (~6,400 tokens) unstated — larger than the worst-corner headroom of 2,597 tokens.
- post-fix verification (this tree): `moai spec lint …/spec.md` → `✓ No findings`, exit 0; REQ 16 / AC 16 unchanged
- repository state (2026-08-17 갱신): 산출물 4개는 커밋됨 — 브랜치 `spec-always-loaded-diet-plan` 의 `be1958a4d`, PR #1576 에 포함. 리포 로컬 all-tier PR 정책은 브랜치+PR 로 충족했고 머지 대기.
- revision (2026-08-17): PR #1576 CodeRabbit 인라인 리뷰의 🟠 Major 5건 반영 — acceptance.md(§A 규율 4 + `headroom` frontmatter 한정, AC-ALD-006 `BASE_REF` 고정 + `sort -u` 제거, AC-ALD-013 문단 스코프 인용 판정, AC-ALD-014 슬롯 4개 단언)과 progress.md 본 기록. 🟡 Minor 4건·🔵 Trivial 2건은 사용자 결정으로 이번 개정 범위 밖.
- Implementation Kickoff Approval (plan→run) has NOT been requested or granted.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

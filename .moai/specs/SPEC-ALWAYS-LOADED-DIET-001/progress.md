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

### M1 — `kanban-dispatch.md` 분리 (2026-08-17)

- Phase 1 re-audit: **skipped as eligible** — plan-audit iter2 PASS 0.845 (Tier M threshold 0.80) + artifact content unchanged since the verdict (squash-merge of PR #1576 did not alter artifact bytes).
- Implementation Kickoff Approval: granted orchestrator-side (run-phase M1 dispatch received 2026-08-17). The §E.1 "NOT been requested" line reflects plan-phase authoring time.
- Tree: worktree `feat/spec-always-loaded-diet`, pre-M1 HEAD `062a995d9` (base origin/main).

**Construction (verbatim move, no rewrite — REQ-ALD-005/AP-2):**

- `kanban-dispatch-detail.md` = frontmatter + ownership declaration + `sed -n '13,59p;62,92p;124,152p'` of the original. Line 61 dropped as the collapsed duplicate blank at the Class-B lift-out point (plan M1 step 3's 1-byte cleanup).
- Stub = stay lines `1,6 / 7,12 / 93,98 / 60 / 99,123 / 153,230` + pointer line (after the Loading-scope quote) + one separator blank + footer version line. The Class B paragraph (original line 60, 495 B) relocated **whole-line** into `## Completion is read, never trusted`, between "Before moving a card…" and "This applies equally…".

**Baseline (pre-edit, this run, this tree, HEAD 062a995d9):**

```
files=14 bytes=295044 tokens=73761 headroom=1239   # acceptance §A fm() form
bytes=295044 tokens=73761 headroom=1239             # plan §C grep -rL form (identical)
```

**Post-split (this run, this tree, working tree):**

```
files=14 bytes=287068 tokens=71767 headroom=3233    # fm() form
bytes=287068 tokens=71767 headroom=3233              # grep -rL form
```

Delta: −7,976 bytes → headroom 1,239 → 3,233 (+1,994). AC-ALD-001 checkpoint (final gate is post-M4) already holds: 3,233 > 1,239.

**M1-decidable AC matrix (all PASS, commands run in this tree):**

| AC | Command (form) | Observed | Verdict |
|---|---|---|---|
| AC-ALD-003 | 9 BINDING headings `grep -c` in stub; 6 LEAD-ONLY `grep -c` in stub; anchor `names that path in its completion report` both files | 9×1; 6×0; relocated_in_stub=1 relocated_in_companion=0 | PASS |
| AC-ALD-004 | `grep -c '\[HARD\]'` stub / companion | stub=6 companion=3 (sum 9 = original census 9) | PASS |
| AC-ALD-005 | 7 headings `grep -c` in companion | 7×1 | PASS |
| AC-ALD-006 | `comm -23` orig(BASE_REF=be1958a4d) vs stub+companion, non-empty sorted | missing_lines=0; positive contrast vs /dev/null=140 | PASS |
| AC-ALD-007 | `fm()` paths keys | manager-kanban_key=1 todo_key=1 (domain-keyed, plan D2) | PASS |
| AC-ALD-008 | (a) names within ±2 lines of the `kanban-dispatch-detail.md` mention (pipe form), (b) ownership_decl, (c) footer | (a) 6×1, (b) ownership_decl=3, (c) version line present | PASS |
| AC-ALD-009 | `wc -c` both files | stub=13027 companion=8866 sum=21893 (range 21003..21903; overhead 890 B, within the 600–900 estimate) | PASS |
| AC-ALD-012 | `headroom` files= + `fm()` paths_key | files=14, exists=1 paths_key=1 (cache-aware-execution-reference.md is M2 — AC completes at M2) | PASS (M1 scope) |
| AC-ALD-002 | guard + budget const + go diff | `ok github.com/modu-ai/moai-adk/internal/config 0.607s` exit=0; budget_const=1; go_changed=0 | PASS |

Note (AC-ALD-008a form): acceptance.md's `<(grep -A2 -B2 …)` process-substitution form yields empty `grep -c` output under this environment's zsh; the pipe form `grep -A2 -B2 … | grep -c` is semantically identical and was used. Bash (sync-audit) runs the acceptance form directly.

**Other verification:** `go build ./...` → exit 0. `go test ./internal/config/` (full package) → `ok github.com/modu-ai/moai-adk/internal/config 2.551s`. Subagent boundary: `grep -rn 'AskUserQuestion'` on the two rule files → 2 matches, both pre-existing doctrine-text mentions (original census 2: 1 stayed in stub `Boundaries`, 1 moved verbatim to companion `Entry into the board`); no new occurrences. golangci-lint: n/a — 0 Go files changed; markdown has no lint gate. RED-state evidence: n/a — documentation milestone, no test-first artifact (recorded as a gap, not fabricated).

**Residual (recorded, deliberately not fixed):** the companion line "This is the same shape as the CodeRabbit section below." now reads stale inside the companion (the CodeRabbit section stayed in the stub). Rewording it would break AC-ALD-006's line-set preservation and violate AP-2 — the line must stay verbatim; the referenced section exists in the stub.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- decision: `sub-agent` (Mode 5 — sequential per milestone)
- rationale: coding-heavy documentation surgery on one always-loaded rule file per milestone; single-writer discipline protects the AC-ALD-009 byte-sum attribution and the shared §C measurement chain (M1→M4 ordering is load-bearing: M2 adds back bytes M1 removed, M3 adds the recurrence control, M4 mirrors and takes the final measurement)
- autonomous progression: goal armed orchestrator-side (ac_converge family); milestones delegated one at a time

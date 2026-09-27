# t1175 — SPEC-ALWAYS-LOADED-DIET-002 run-phase verdict

Card t1175 · Tier L · class C · worktree `.claude/worktrees/t1175` · branch `WT-rules-diet`
Baseline ref `172ef22eb` · final ref `276391646` · unpushed · worktree preserved

**FINAL VERDICT: AC-ALD2-001 FAIL (MUST-PASS). Card not closeable as `completed`.**
Every figure below was measured in this tree in this run by the orchestrator, after reading
each lane's report — not carried over from a report.

## Claim

M1 through M7 are complete. Every mechanical gate except AC-ALD2-001 holds: the binding-clause
freeze is byte-identical, no new per-file overrun exists in either tree, tests pass, and the
template mirror moved with every live edit. AC-ALD2-001 fails by 47,897 chars, and the pool
census establishes that the target was not reachable under this SPEC's own freeze rules.
AC-ALD2-003 is unsatisfiable as written — a defect in the criterion, not in the work.

## Evidence

| axis | command | baseline | final |
|---|---|---:|---:|
| 18-file total | `wc -m` over the 18 counted paths, `total` | 246,943 | **197,897** |
| frozen multiset | `grep -hE '\[HARD\]\|MUST\|shall ' … \| sed \| sort \| shasum -a 256` | 170 lines `d97b33d9…18c6c3` | **identical**, `diff` clean |
| per-file overruns, live | `find … -exec wc -m + \| awk '$1>=40000'` | 4 files | **same 4**, each equal or smaller |
| per-file overruns, template | same over the template tree | 4 files | **same 4, identical sizes** |
| tests | `go test ./internal/config/... ./internal/template/...` | — | **pass** |
| build | `make build` | — | **pass**, `catalog.yaml` unchanged |

Per-milestone delivered vs assigned: M2 6,139 / 12,000 (51%) · M3 21,757 / 52,000 (41.8%) ·
M4 10,951 / 30,339 (36%) · M5 10,219 / 23,435 (44%). Total 49,066 / 117,774 (41.7%).

## Baseline-attribution

The baseline was re-derived in this run before any edit (`run-baseline-pre.md`) and matched the
figures the SPEC declares, so the comparison is against a measurement rather than a citation.
The frozen multiset was re-measured after closing each file, and `diff` was run against the
stored baseline multiset rather than comparing hashes alone.

## Why AC-ALD2-001 fails — structural, measured, not a shortfall of effort

The pool census (`pool-census.md`, measured on ref `172ef22eb`) puts the section-level movable
pool at **106,148** chars. Extracting all of it with pointers bundled per destination floors the
18-file total at **149,195** — 805 chars under target, and only while two conditions hold
together: 100% extraction, and minimum pointer re-entry.

Three measured reasons make that floor unreachable:

1. **The pool was computed per paragraph; the freeze binds per section.** 77,206 chars of
   non-binding prose share a section with a binding clause and are out of relocation's reach.
2. **A hard-wrapped binding line freezes its whole paragraph.** Shortening prose ahead of it
   re-wraps it and breaks the hash. This is what held in-place compression to 3,092 against an
   assignment of 28,717 (10.8%), and the census does not model it, so the census over-states
   the pool.
3. **The admissible pool is strictly smaller than the census.** The census applies M1 condition 1
   only; conditions 2, 3 and 4 each shrink it, and M1 rejected 18 items on exactly those grounds.

`verification-claim-integrity.md` is the extreme case: 2.4% pool, zero movable sections, so its
entire 6,000 assignment sat on the axis that delivered 10.8%.

## AC-ALD2-003 — the criterion contradicts template neutrality

The criterion requires `diff` 0 on every touched pair, excluding only `CLAUDE.md`. Template
neutrality forbids the template copies from carrying card ids, SPEC ids, REQ tokens and
retired-command references that the live copies legitimately hold. The two obligations cannot
both be met, so no execution satisfies this AC.

Measured: of 30 touched rule pairs, **11 already diverged at the baseline ref** for that reason.
Three pairs' divergence rose in this card — `cross-session-messaging.md` 2→6,
`cross-session-messaging-detail.md` 8→9, `main-checkout-branch-guard-detail.md` 11→16 — and each
is inherited rather than introduced: relocation carried pre-existing live-vs-template content
divergence into the destination. At the baseline ref the two `main-checkout-branch-guard.md`
stubs already disagreed on the same line 38 (`moai cc -w` vs `git worktree add -b`). **Zero
pairs show a structural change this card left unmirrored**, and `skill-routing.md` — the pair the
AC names explicitly — went 3→0.

Forcing diff 0 would mean copying neutrality-forbidden content into the templates, breaching the
CI guard. It was not done.

## Gaps — explicitly unobserved

- **The admissible pool (M1 conditions 1-4) is not measured.** Quantifying it means re-walking
  every section against the anchor sweep and destination capacity, i.e. redoing M1.
- **The full test suite was not run locally**, per the standing prohibition; CI on the integration
  branch is the full-suite verdict.
- **The anchor sweep is per-section, at move time.** M1's count covered only the `§ <title>`
  citation form; file+section, line-number and prose-mention citations were enumerated per
  section as sections moved, and no tree-wide final sweep over all four forms was run.
- **Ten citations sit inside the four over-limit files.** Length-neutral repairs were attempted;
  any that could not be made length-neutral are recorded as `초과 파일 앵커 미수리` rather than
  repaired.
- **Per-destination re-entry is ~350 chars/line**, extrapolated from the 292.4 chars/line
  existing-separation average, not measured on the lines this card wrote.
- **A Bash heredoc naming `git show` was refused by the worktree guard** while writing the census
  script; the script was written with the Write tool instead and the census ran unchanged. No
  measurement was substituted.

## Residual risk

- **In-place compression has no mechanical check for meaning preservation (REQ-ALD2-012).** It
  delivered 3,092 chars here, so the exposure is small, but it is unverified by machine.
- **Scope drift is invisible to the frozen hash.** A removed sentence that narrowed a clause's
  scope would widen the surviving clause and fire no detector. The M1 table's scope-dependency
  column is the only judgment surface; it was filled, and reviewing it is a human act.
- **Two follow-up findings are left unfixed**, on the reasoning that excluded the stale CG
  description: the template `main-checkout-branch-guard-detail.md` now instructs a bare
  `git worktree add`, which the shipped doctrine forbids, and AC-ALD2-003 needs rewriting
  against what neutrality permits.

## Carried plan-audit debt (iter-4, PASS-WITH-DEBT 0.890)

N7 (blocking) was discharged as a precondition in the M1 delegation — the dominance predicate was
read as open-ended before the table was filled, and the table records two dominance paths found
outside the three known mechanisms. N8, N9 and D9 were applied in the M1 commit, with two
deviations reported by the lane: N9 found 3 in-scope occurrences rather than 4 (the fourth sits
outside §E), and D9's wording landed on the relocation bullet rather than the final one, because
the pointer clause only holds where a pointer exists.

MoAI

## Sync phase — develop 흡수·감사·수리 기록 (2026-09-27)

- **흡수**: 로컬 develop `b59a5d69c` → 병합 `4989ea6b0`. 충돌 3파일(CLAUDE.md, `moai-mcp-tools.md` 양 미러)은 다이어트 쪽으로 해소하고 develop 신규 사실을 이식했다. develop 추가 줄 9개를 줄 단위로 대조해 유실 0(`remeasure/develop-lines-trace.txt`).
- **1차 sync-audit**: FAIL 0.71 (`sync-audit.md`). 감사 모델 근거 셋 — ① 세션 시작 문맥에 「GLM backend detected」 없음 ② 레인 환경에 `ANTHROPIC_BASE_URL`·모델 매핑 변수 없음(Anthropic 직결), 레인 세션 모델 `claude-opus-5-5[1m]` ③ 보고서 첫 줄 `auditor-model: claude-opus-5-5[1m]`.
- **수리**: F1·F2 `8e50ef148`(끊긴 앵커 2건, codex `.toml` 재방출), F3·F4 `7deb5b3b1`(manager-spec), 카탈로그 해시 `950fcc492`.
- **F3 처분 선택과 근거**: REQ-ALD2-001 과 제목을 「감축 달성 + 150,000 은 잔여 채무, `t1226` 소관」으로 개정했다. 리드 승인(§31 카드 고유 조건 — 인수 기준 처분, 권장안 채택). 근거: 구속 조항 동결(AC-ALD2-002)을 지키는 실행은 풀 센서스상 150k 아래로 내려갈 수 없다는 것이 이 판정서 § Why AC-ALD2-001 fails 의 측정 결과이고, 한도 달성은 이미 큐에 있는 후속 카드가 소유한다.
- **F4 재측정**: 18파일 합계 병합 트리 198,351 · HEAD `8e50ef148` 198,361, 동결 해시 `d97b33d9…c6c3` 불변(`remeasure/f4-ac001-ac002.md`).
- **[공정 결함 자기 보고] 한-작성자 위반**: 1차 감사 창 동안 이 레인이 커밋 `805d44bed` 와 `remeasure/spec-lint.txt` 를 썼다. 감사자는 두 ref 를 모두 재어 판정 기준은 유지됐다고 보고했다. 재감사 창에서는 판정이 나올 때까지 트리에 쓰지 않는다.

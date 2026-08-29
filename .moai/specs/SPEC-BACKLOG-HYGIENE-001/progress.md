# SPEC-BACKLOG-HYGIENE-001 — Progress (card t332)

## §E.1 Plan-phase Audit-Ready Signal

- **Card**: t332 — backlog hygiene sweep (read-only).
- **Worktree / HEAD**: `.claude/worktrees/t332`, branch `WT-backlog-hygiene`, HEAD `15453140a`.
- **Artifacts authored**: `spec.md`, `plan.md`, `acceptance.md`, this file.
- **Tier proposed**: M (plan.md §A) — proposal, not a decision. Now within the 16/16 budget.
- **Requirements**: 16 (REQ-BH-001..016). **Acceptance criteria**: 16 (AC-BH-001..016).
- **SPEC ID pre-write check**: `[[ "SPEC-BACKLOG-HYGIENE-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]`
  → `PASS`.

### Plan-audit iter-1 → iter-2 (v0.1.0 → v0.2.0)

Verdict under repair: **FAIL, 0.74** against the Tier M threshold 0.80
(`.moai/reports/t332/plan-audit-iter1.md`), driven by must-pass MP-3 plus six blocking defects.

| Defect | Disposition | Deciding measurement (this tree, `15453140a`) |
|---|---|---|
| MP-3 `lifecycle: spec-first` out of enum | fixed → `spec-anchored` | schema SSOT `spec-frontmatter-schema.md:63` — enum is `spec-anchored \| spec-lite \| exploratory` |
| D1 refs unfetched / unpinned | fixed — M1 step 1 fetches once and pins both SHAs; REQ-BH-009 + AC-BH-007 require the pinned SHA in every citation | `grep -n fetch` over the three artifacts → 0 matches at v0.1.0 |
| D2 23 requirements over the Tier M ceiling of 16 | fixed by **consolidation to 16**, not by tiering up | `spec-workflow.md` ceiling table: M=16, L=25, applied independently |
| D3 §E write boundary contradicts REQ-BH-006 | fixed — invariant restated behaviourally, `relate` carve-out named, gitignore caveat recorded | `internal/cli/todo_relate.go:66` → `newTodoStore().Mutate(...)` writes `backlog.json` |
| D4 AC vacuous on path B | fixed — AC-BH-008 guards on the measured `strings` value, plus a path-B post-rebuild conjunct | `strings ~/go/bin/moai \| grep -c 'worktree_base_branch'` → `0` |
| D5 no-mutation observables fail in the same direction | fixed — AC-BH-006 adds a card-row digest (M2 vs M5); AC-BH-004/005 kept | `moai todo edit` changes text while leaving all three counts identical |
| D6 five REQs with no deciding AC | fixed by folding, not by adding five siblings — AC-BH-012 widened to every entry, in-flight cross-check attached to AC-BH-009, delta branch decided inside AC-BH-001, one new AC-BH-011 | REQ↔AC map now in acceptance.md §B |
| D7 `Where` for a runtime condition | fixed → `When` (REQ-BH-006), `While` (REQ-BH-012) | — |
| D8 REQ-BH-010 passive, no actor | fixed → "The sweep shall refresh … and establish" (now REQ-BH-009) | — |
| D9 batch sizes misstated as "11-12" | fixed → measured sizes stated | awk bucketing: B1=10 B2=11 B3=13 B4=10 B5=13 B6=10, total 67 |
| D10 embedded-newline caveat unsupported | fixed → replaced with the arithmetic | 5 landed + 91 no-link = 96 = the card count (68+10+18) |
| D11 t256 is `dropped` | fixed → M4 states the relation is a reading only, never carried into the proposal | `awk -F'\t' '$1=="t256"'` → `t256  dropped` |

Left unchanged, deliberately: every count in §B.1, §C, REQ-BH-016 and AC-BH-003 (the auditor
verified them correct against the snapshot), and the Tier M proposal itself.

### Direct revision after iter-2 (PASS-WITH-DEBT 0.85, `.moai/reports/t332/plan-audit-iter2.md`)

Not an iteration — three post-repair defects, fixed in place. Counts unchanged at 16/16; no
requirement or criterion added.

| Defect | Disposition | Deciding measurement (2026-08-29) |
|---|---|---|
| **N1** AC-BH-006 had no deciding procedure — the only one of 16 criteria naming no command | fixed — the exact `jq … \| shasum -a 256` invocation is now inline, with the negative control ("M4's `relate` runs between the captures; digests must not move") and a measured probe table covering both directions | store read at `.moai/state/todo/backlog.json`: `jq -r 'keys'` → `findings, items, last_seq, version`; item keys → `added_at, id, spec_id, state, text`; `.items\|length` → 96. **`findings` is a top-level sibling of `items`, not a per-item field** (`backlog_store.go:191`), so `.items[]` excludes it structurally. Probes on scratch copies: baseline `56e1387e…` twice; +finding `56e1387e…` (unchanged); text edit `46e61f16…`; state flip `fd9964a7…`. Real store never written. |
| **N2** false traceability cell — AC-BH-015 listed under REQ-BH-007 | fixed — cell corrected to `AC-BH-004, AC-BH-006`; a note records that AC-BH-015 decides spec.md §E's write-isolation constraint and is deliberately requirement-less | AC-BH-015 tests pairwise-disjoint batch id sets, which is §E's constraint, not REQ-BH-007's two-observables obligation |
| **N3** "the recorded measured value" ambiguous once path B leaves two counts | fixed — the **governing count** is now named: path A's single pre-sweep measurement, path B's post-rebuild measurement, with the ordering (rebuild first, verdicts after) recorded | wording clarification only; the bullet already erred safe |

Side-correction carried by N1: the same wrong shape (`Findings` as a per-item field) appeared in
REQ-BH-007 and in plan.md M2. Both now describe the real structure. This was a factual repair to an
existing clause, not a new obligation.

### N1 follow-up — the digest path (third vacuity route, closed)

The residual risk flagged when N1 landed was confirmed by measurement rather than dismissed: the
command resolved its path with `git rev-parse --show-toplevel`, which inside a worktree names the
**worktree** root, where no queue store exists. `ls` on that path in the t332 worktree returns
`No such file or directory`, rc=1. Since the run phase executes inside a worktree by the
card-isolation rule, the criterion as written read nothing exactly where it would run — a third way
for AC-BH-006 to be vacuous, alongside the two closed by N1.

Fixed as a **two-step** procedure, both steps measured in this worktree, 2026-08-29:

| Step | Command | Observed |
|---|---|---|
| 1 | `git rev-parse --path-format=absolute --git-common-dir` | `/Users/goos/MoAI/moai-adk-go/.git` — the common dir every linked worktree shares; its parent is the primary checkout |
| 2 | `jq -S -c '[.items[] \| {id, state, text}] \| sort_by(.id)' <parent>/.moai/state/todo/backlog.json \| shasum -a 256` | `56e1387e…b346b` — reproduces the baseline from inside the worktree |
| control | the one-liner computing the path inline via `$(dirname "$(git rev-parse …)")` | **refused**: "this command is too complex to verify that it stays inside the worktree" |

The two-step split is therefore load-bearing, not stylistic, and AC-BH-006 says so — a run phase
that collapses it back into a one-liner is refused by the guard, and one that uses
`--show-toplevel` reads nothing. The SPEC records the derivation (common-dir → parent), never this
machine's literal path. plan.md M2 carries the same two-step form and the same two reasons.

Counts unchanged at 16/16; no requirement or criterion added. Only the path resolution changed —
the projection was already correct and its digest is unchanged.

- **Gating lint after repair**: `moai spec lint --strict .moai/specs/SPEC-BACKLOG-HYGIENE-001/spec.md`
  → `✓ No findings`, **rc=0**. Recorded with the caveat that this linter's `lifecycle` check is
  presence-only (`internal/spec/lint.go:765`), so it did NOT decide MP-3 in either direction — that
  repair is verified against the schema SSOT by eye.
- **Gate**: Implementation Kickoff Approval not yet run. Run phase has not started.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

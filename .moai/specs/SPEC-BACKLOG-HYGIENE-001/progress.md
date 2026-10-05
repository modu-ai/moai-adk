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

Run phase opened 2026-08-30 in worktree `.claude/worktrees/t332`, branch `WT-backlog-hygiene`,
HEAD `6165f9f5e` after absorbing `origin/develop` at `ee50984ab` (merge, no conflicts).

Implementation Kickoff Approval: **granted by the lead on 2026-08-30**, together with answers to
the two plan-phase open questions — Q1 → **path A (direct git)**, Q2 → **Tier M confirmed**. The
lead independently measured `strings ~/go/bin/moai | grep -c worktree_base_branch` = 0 and confirmed
the binary-lag premise.

### M1 — tooling baseline (CLOSED)

Cited command + output: `.moai/reports/t332/00-tooling-baseline.md`.

| Item | Value |
|---|---|
| `git fetch origin develop main` | 2026-08-30T11:16:22Z, both refs updated |
| pinned `origin/develop` | `ee50984abe4f11ac337382b48a26328f091e200a` |
| pinned `origin/main` | `48239c7dc7428c8751a04f6321887c2d36123884` |
| `strings ~/go/bin/moai \| grep -c 'worktree_base_branch'` | **0** (governing count, path A) |
| landing method | **path A — direct git**; path B rejected because a mid-batch reinstall mutates shared state for nine other live lanes |
| t342 control vs pinned develop | non-empty (5 commits, incl. `15453140a`) |
| t342 control vs pinned main | empty |
| `git worktree list` | captured to `00-worktree-list.txt` (153 lines) |

### M2 — snapshot, scope, opening digest (CLOSED)

Cited command + output: `.moai/reports/t332/01-scope.md`.

| Item | Value |
|---|---|
| snapshot | `queue-snapshot-run.tsv`, 101 rows, captured 2026-08-30T11:16:45Z from primary checkout |
| counts | 62 queued / 17 picked / 18 dropped = 97 cards (+4 relation rows) |
| independent cross-check | `sqlite3 backlog.db "select count(*) from items;"` → 97, agrees |
| REQ-BH-003 delta branch | **FIRED.** plan-phase 68 queued → run-phase 62. Reconciles exactly: 7 cards moved queued→picked (`t294 t299 t318 t332 t336 t343 t362`), 1 admitted (`t363`); 68 − 7 + 1 = 62 |
| `t332` own state | moved `queued` → `picked`, so the AC-BH-003 self-exclusion term is now zero |
| **in-scope** | **62 cards** (plan.md's 67 is superseded) |
| opening card-row digest | `86a3fb05de900816670a1a4b6e6f3f3e9c6af8e6b4cc6eb3997d0e268bffdc20` |
| digest Step 1 output | `/Users/goos/MoAI/moai-adk-go/.git` → primary `/Users/goos/MoAI/moai-adk-go` |

### Finding B1' — the digest criterion was measuring a dead file (acceptance.md v0.3.0)

The queue store migrated to SQLite at `3cb258d62 merge(WT-todo-sqlite): SPEC-TODO-SQLITE-001 (t306)`.
`backlog.json` has not been written since (mtime 2026-08-29T20:35Z vs `backlog.db` 2026-08-30T20:15Z;
96 items vs 97; `t294` reads `queued` there and `picked` in the live store; `t363` is absent
entirely). AC-BH-006's v0.2.0 command therefore returned its own recorded plan-phase figure
`56e1387e…b346b` **unchanged** — after the queue had moved seven cards and gained one.

**A criterion whose observable is a dead file cannot red.** This is the card's own subject —
a measurement whose premise rotted while the sentence citing it stayed unchanged — reproducing
inside the card's own acceptance criteria. Repaired to the live store with all four controls
re-measured there (baseline stable / negative unchanged / two positives changed); evidence in
`00-tooling-baseline.md` § Finding B1'. Lead independently reproduced the divergence and approved.

### Acceptance repairs applied in run phase (3)

| AC | Defect | Direction |
|---|---|---|
| AC-BH-006 | observable was a file no longer written | **could never red** |
| AC-BH-004 | non-emptiness conjunct required `wc -l` ≥ cards read, which is the per-card live invocation REQ-BH-001 **prohibits** | **could never green** |
| AC-BH-002 / AC-BH-003 | Given-clauses inlined plan-phase lists and arithmetic invalidated by the queue delta | scope restatement (stricter) |

The AC-BH-004 defect is the mirror image of AC-BH-006's: one criterion could not fail, the other
could not pass, and both were authored in the same plan phase that was watching for exactly this.

### Finding B5' — plan.md's own gitignore premise is false (fourth instance)

plan.md B5 and spec.md §E both state that evidence written under `.moai/reports/t332/` inside this
worktree "is gitignored and is lost on disposal", making recovery to the primary checkout an explicit
M5 step. Measured at M5:

```
$ git check-ignore -v .moai/reports/t332/report.md
(no output)
$ echo $?
1
```

**Not ignored.** `git status --short` lists the whole tree as untracked-but-trackable, so the
evidence is committed on the card branch and travels with it — no out-of-band `cp` + `cmp` recovery
is needed, and none was performed. The recovery step is discharged by the commit itself.

This is the **fourth** instance of the same shape in this card's own lifecycle, and the lead flagged
the second: the plan-audit iter-1/iter-2 reports asserted their own output was gitignored, which
`git check-ignore -v` rc=1 refuted. The claim was corrected in the audit reports and then **repeated
unchanged in plan.md B5 and spec.md §E** — a correction that closed the instance in front of it and
left the same sentence standing one file over. Nothing in the plan phase re-ran the one-line command
that decides it.

### Recurrence record (lead's request)

This sweep's own lane hit the "a check exists and sees nothing" shape **three** times: (1) the
plan-audit N1 finding, where AC-BH-006 had no deciding command at all; (2) the iter-1/iter-2 audit
reports asserting their own output was gitignored, which `git check-ignore -v` rc=1 refutes; (3) the
dead-file digest above. Caution did not prevent any of the three — each was caught by running a
command, never by re-reading the text.

## §E.3 Run-phase Audit-Ready Signal

`run_status: audit-ready`
`run_complete_at: 2026-08-30T11:45Z`
Milestones closed: M1, M2, M3, M4, M5 — each citing the command it ran.

### AC PASS/FAIL matrix (16/16 decided)

| AC | Verdict | Deciding command → observed |
|---|---|---|
| AC-BH-001 | **PASS** | `cut -f2 queue-snapshot-run.tsv \| grep -c '^queued$'` → `62`; `01-scope.md` §1/§3 record capture time `2026-08-30T11:16:45Z`, HEAD `6165f9f5e`, and the delta vs §B.1's 68 reconciled as `68 − 7 + 1 = 62` |
| AC-BH-002 | **PASS** | `comm -12 entries.txt excluded.txt \| wc -l` → `0` (entries ∩ (picked ∪ dropped), run-phase lists of 17 + 18) |
| AC-BH-003 | **PASS** | `grep -h "^### t" cards/*.md \| wc -l` → `62`; `diff entries.txt inscope.txt` → no output. Both halves: count AND set |
| AC-BH-004 | **PASS** | `grep -cE "moai todo (drop\|edit\|done\|undrop\|move\|unpick)\|moai todo next t[0-9]"` → `0`; `grep -cE "^2026-" invocations.log` → `2`, equal to the invocations actually issued (1 snapshot + 1 refused `relate`) |
| AC-BH-005 | **PASS (vacuous antecedent, stated)** | Zero `relate` lines were successfully recorded, so the well-formedness quantifier ranges over the empty set. The **second** conjunct is non-vacuous and measured: `queued` 62 → 62 and `dropped` 18 → 18 unchanged after the sweep. The one attempted `relate` carried both `--relation contains` and a non-empty `--note` and would have been well-formed; it was refused as a duplicate |
| AC-BH-006 | **PASS via the attribution branch** | opening `86a3fb05…dc20` → closing `053e9991…7b33`, **differ**. `comm` on the id sets names exactly one differing card, `t346`, attributed to the lead with commit `282daef19` and out of scope from before the first read. `diff open-inscope.tsv close-inscope.tsv` → no output: all 62 in-scope rows byte-identical. Both captures record their Step 1 output |
| AC-BH-007 | **PASS** | All 3 `landed` entries (t201, t313, t347) carry a commit SHA, the **pinned** ref SHA as a literal, and the `--is-ancestor` exit code |
| AC-BH-008 | **PASS** | `00-tooling-baseline.md` records the governing `strings` count `0` and path A; `grep -c "moai todo pr" cards/*.md` → no entry cites the installed landed column. Control present: t342 non-empty vs pinned develop, empty vs pinned main |
| AC-BH-009 | **PASS (after two orchestrator corrections)** | No entry rests on a branch-name resemblance. Positive direction measured mechanically rather than accepted from the workers: `join` of the in-scope list against `00-worktree-list.txt` gives exactly 4 cards with a live worktree — t154, t216, t313, t337 — and `git merge-base --is-ancestor <tip> <pinned-develop>` gives exit 1, 1, **0**, 1. The three unmerged ones all carry `in-flight-unlanded` with branch and tip; t313 is merged and correctly `landed` |
| AC-BH-010 | **PASS** | 1 `unknown` (t224), carrying its reason: both queries ran and returned empty, but the requested doctrine text already exists and the sweep cannot tell whether it predates the card. Not collapsed to `not-landed` |
| AC-BH-011 | **PASS (after three orchestrator normalizations)** | All 62 entries carry a one-sentence premise and exactly one of `holds` (42) / `falsified` (11) / `unverified` (9); every `unverified` states its reason. Three workers wrote compound verdicts; each was normalized to `unverified` in place with the normalization stated |
| AC-BH-012 | **PASS** | Each of the five section headers appears exactly **62** times across `cards/*.md`. Second conjunct: all 11 `falsified` entries carry a fenced command block in Evidence (checked per entry; empty exception list) |
| AC-BH-013 | **PASS** | The 5 `already-landed` proposals each name a commit SHA or a PR state rather than an impression; the 3 mention-only grep hits were ruled mentions by reading the matched commits in full, and say so |
| AC-BH-014 | **PASS** | `02-overlaps.md` §2 carries 10 reciprocal rows, each naming both ids and the specific shared file, mechanism, or ref, plus the one-directional and out-of-scope candidates. No row proposes a performed fold (`grep -ci "fold performed\|folded into"` → 0) |
| AC-BH-015 | **PASS** | `sort entries.txt \| uniq -d` → empty (pairwise-disjoint id sets across the 6 batch files), and the union equals the AC-BH-003 in-scope set |
| AC-BH-016 | **PASS** | `report.md` §5 carries **62** disposition rows, each naming a token from the five-value vocabulary and its single piece of evidence; the report states in its opening line that no card was mutated and the list awaits the operator; zero un-drop proposals; the one `dropped`-counterpart relation (`t318 absorbs t256`) is reported as a reading only and appears in no disposition row |

**16/16 PASS.** Two carry a stated qualifier rather than a bare pass: AC-BH-005's first conjunct is
vacuous by antecedent (zero relations recorded) and says so; AC-BH-006 passes through the attribution
branch rather than byte-identity, with the differing card named and its commit cited.

### Milestone evidence

| M | Closed by | Output |
|---|---|---|
| M1 | fetch + `rev-parse` + `strings` + t342 controls + `git worktree list` | `00-tooling-baseline.md`, `00-worktree-list.txt` |
| M2 | `moai todo` snapshot + `uniq -c` + sqlite cross-check + opening digest | `01-scope.md` §1-§7, `queue-snapshot-run.tsv` |
| M3 | 6 read-only `Agent()` workers, staggered; orchestrator post-check normalized 3 verdicts, reclassified t154, filled t216's ancestry | `cards/batch-1..6.md` (62 entries) |
| M4 | pairwise comparison over premise restatements + targeted card reads + `--is-ancestor` + `sed` on `session_worktree.go` + the refused `relate` | `02-overlaps.md` |
| M5 | closing digest + `comm` attribution + in-scope `diff` + assembly | `report.md`, `01-scope.md` §8, `invocations.log` |

### Fan-out discipline

Six workers, one output file each, pairwise-disjoint id sets — measured, not assumed
(`uniq -d` empty). Spawn staggered per the cache-aware directive. No worker invoked `moai todo`; the
invocation log is the orchestrator's alone and carries two lines.

### Orchestrator corrections to worker output (5, all stated in place)

| Correction | Card(s) | Why |
|---|---|---|
| compound premise verdict → `unverified` | t236, t239 (batch-2), t361 (batch-6) | AC-BH-011 admits exactly one token; a partially-decided premise is `unverified` |
| `unknown` → `in-flight-unlanded` | t154 | measured: live worktree, tip `dbb87f14f` not an ancestor of pinned develop (exit 1). `unknown` hid a measured fact behind an unmeasurable one |
| `already-landed` → `needs-operator-decision` | t154 | the disposition contradicted the entry's own landing verdict |
| `--is-ancestor` gap filled | t216 | the worker recorded it as a Gap it did not run; orchestrator ran it (exit 1) |
| extraction error, caught and corrected | t337 | the orchestrator's own tally read the first token of *"`already-landed` is NOT proposed"* as the proposal. The report's own machine-read inverted a verdict — recorded because it is the same defect class the card is about |

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: "2026-08-30"
sync_commit_sha: "95039fbef"  # a commit cannot cite its own SHA — backfilled in the commit immediately following the sync commit
sync_status: completed
frontmatter_status_transitions:
  spec_md: "in-progress → completed"

Sync-phase report additions (`.moai/reports/t332/report.md`):

- **§3, new item 4** — the fourth instance of the same shape this card exists to catch: plan.md B5
  and spec.md §E both assert evidence under `.moai/reports/t332/` is gitignored, which
  `git check-ignore -v` (rc=1) refutes. Repair scoped to the report only — the false sentences in
  plan.md B5 / spec.md §E body are out of this card's scope and stay a disposition item for the
  operator.
- **§5, t281 row** — the evidence column now names the deciding commit as a literal SHA
  (`11216d13f`, full `11216d13f612f7e7161487a4e1369a47612f0b4c`), on pinned `origin/develop`.
- **§5.1, new grouped paragraph** — the `t313 contains t295` relation cannot be annotated with the
  measured fact that t313 landed without resolving t295; the append attempt was refused
  (`invocations.log`), and no verb in this card's deliverables can update a recorded relation.
  Stated as an operator disposition proposal; the §5 table stays at exactly 62 rows.

**Quality-gate measurement (not an assertion).** Branch diff vs merge-base
`ee50984abe4f11ac337382b48a26328f091e200a`: `git diff --stat ee50984abe4f11ac337382b48a26328f091e200a...HEAD -- '*.go' | wc -l` → `0`. The full diff is 46 files, all markdown/TSV under
`.moai/reports/t332/` and `.moai/specs/SPEC-BACKLOG-HYGIENE-001/`. Lint / test / MX-tag / coverage
gates have no `.go` surface to run against in this SPEC — this is a stated **GAP**, never a claimed
PASS.

changelog_entry_position: "[Unreleased] → Added (this sync commit)"

**B12 self-test (pre-append)**: `grep -c 'SPEC-BACKLOG-HYGIENE-001' CHANGELOG.md` → `0` (measured
before the CHANGELOG append in this same sync commit).

### Integration-window addendum (2026-08-30, lead-assigned window)

The sync commit `95039fbef` closed the SPEC; this addendum records work the lead directed inside the
integration window, and adds no requirement or criterion.

- **Absorption**: `git merge origin/develop` at `52c3fe590e3ea11b37389d4248162055f22f1c59` → merge
  commit `94d234c61`. One conflict, `CHANGELOG.md`, both `[Unreleased] → Added` entries prepended in
  the same position; resolved by keeping **both** (`SPEC-BACKLOG-HYGIENE-001` then
  `SPEC-RED-NOW-THRESHOLD-001`), no entry dropped or reordered below them.
- **Pin re-measurement**: `report.md` §7. The 33-commit delta `ee50984ab..52c3fe590` was searched
  for all 62 in-scope ids; **0 verdicts flipped**. The original verdicts stay in §5 and are not
  overwritten — §7 sits beside them. One grep hit (`t363`) was read in full and ruled a mention.
- **Store delta**, attributed rather than treated as a failure: 96 → 100 rows, `t364`–`t367`
  admitted and `t347` moved `queued → picked`, all by the lead per its own window message. Digest at
  re-measurement `5892574384191c85d58917b224955e5bda8b4e15dfee24f6e4d5f1aacb131e94`. This sweep
  issued no `moai todo` invocation in the window; `invocations.log` still carries two lines.
- **Not measured here**: the repository-wide CI verdict for the merged tree, which belongs to the
  push of `develop`. The lead reported `52c3fe590` attempt=1 with two `Race Test` failures
  (`TestConcurrencyStress`, `TestGitDiffNameCount_Predicate`), neither attributable to this card —
  this SPEC's diff carries 0 `.go` files.

## §F Phase 4 Mode Selection

**Input parameters**

| Signal | Value |
|---|---|
| tier | M |
| scope (files written) | 0 source files; ~12 report files under `.moai/reports/t332/` |
| domain count | 1 (backlog-queue reading) |
| file language mix | 100% markdown output; git + sqlite reads |
| concurrency benefit | **HIGH** — 62 cards, each independently readable, all read-only |

**Mode evaluation**

| Mode | Selected | Rationale |
|---|---|---|
| `direct` | no | 62 cards is far past a trivial single-turn change |
| `serial` | no | the per-card work has no inter-card dependency, so sequencing buys nothing and costs 62 round trips |
| `fanout` | **YES** | read-only, independently partitionable, no shared write path |
| `sweep` | no | not a uniform mechanical transform — each card needs judgement — and this is a reading task, not a transformation |

**Decision: `fanout`**

**Justification.** M3's per-card work is read-only and independent, which is the case
`fanout` exists for; the coding-task parallelism caveat that pushes coding work to `serial` does not
bind, because nothing here writes source. Six workers, one output file each
(`cards/batch-<k>.md`), pairwise-disjoint id sets — the write-isolation constraint AC-BH-015
decides. Spawn was staggered per the cache-aware directive: worker 1 first, the remaining five
after it began producing output. Six is above the 3-5 advisory band and below the runtime cap of
20; the band is a cache/coordination advisory, and the batch partition was fixed at 6 in plan.md
§F M3 before the band was weighed — recorded rather than silently re-tuned.

**Boundary case.** Worker count 6 vs the 3-5 advisory band: resolved toward the plan's partition
because the batches are already sized (10-11 cards each) and re-partitioning to 5 would have made
one batch 13 cards, worsening the straggler. The lead separately warned that six concurrent readers
of one repository can contend on `index.lock`; the workers issue only `git log` / `grep` / `Read`,
none of which take that lock.


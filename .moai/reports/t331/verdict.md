# t331 — lane verdict (SPEC-TODO-LANDING-STATE-001)

Card: t331 · Lane: lane-3 · Branch: `WT-card-landing-state` · Worktree: `.claude/worktrees/t331`
Phases: plan + run + sync, all closed in the worktree · Not merged — the lead names the window
Measured: 2026-08-28 / 2026-08-29

---

## Claim

The card asked why `picked` conflates "in progress" with "finished but not closed". The answer found
in plan-phase, and the thing this SPEC fixes, sits one layer below the symptom the card named: the
landed question was asked about `origin/main` while the project integrates on `develop`, so every
card that landed on develop read as not-landed. `--require-landed` additionally exited 0 printing
`done <id>` when the ref did not resolve, so "the guard passed" and "the guard did not run" were the
same bytes.

Both are closed. The landed answer is now three-valued and resolved from the integration branch, and
every successful `done` prints exactly one verdict token.

**What this card does NOT claim**: nothing here stores landing evidence on the card. That half was
split out to card t359 by operator ruling and depends on this one landing first — a discriminator
that asks the wrong ref cannot usefully store evidence.

## Evidence

### The root cause, measured (plan-phase, re-measured twice since)

```
$ git rev-list --count --left-right origin/main...origin/develop
0	329          # at authoring; 334 at audit; 349 at delta; 375 at iter-2. Left column always 0.
$ git log origin/main --perl-regexp --grep='\bt322\b' --oneline    → 0 commits
$ git log origin/develop --perl-regexp --grep='\bt322\b' --oneline → 5 commits
```

`main` is a strict ancestor of `develop` throughout, so the gap is not divergence — it is lag, and
the landed question was pointed at the lagging ref. The figures are recorded in the SPEC as
re-runnable commands rather than frozen numbers, because a tree SHA does not pin a moving ref; the
drift across four measurements is what made that necessary.

### Plan-phase audit — two iterations, monotonic

| Iteration | Verdict | Score | Blocking |
|---|---|---|---|
| iter-1 (`11426a128`) | FAIL | 0.74 | 10 |
| iter-2 (`45cff0f59`) | PASS | 0.85 | 4 → closed as a delta fix (Tier M iteration ceiling exhausted) |

The thesis survived independent re-measurement at both iterations; what failed was the artifacts
around it. Reports are committed at `.moai/reports/t331/plan-audit.md` and `plan-audit-iter2.md`.

### The operator's kickoff condition — re-observed by the lane, not accepted on report

Kickoff was approved on condition that the probe-to-shipped gap be closed by measurement. The
plan-phase had proved the widened assertion with a throwaway probe; the shipped test was inference.
The run-phase agent reported closing it. **I planted the mutant myself rather than take that
report**, because this criterion had already been wrong twice:

```
$ grep -n 'func queueDirDigest' -A 12 internal/cli/todo_pr_test.go
138:func queueDirDigest(t *testing.T, root string) string {
140:	dir := root                       # was kanban.StateDirForRoot(root)
147:		if d.Name() == ".git" { return fs.SkipDir }

$ shasum internal/cli/todo_pr.go
a3292e44fe1b37df1cb7c2eb58db537f6754f615        # before planting

# 1. before planting
$ go test ./internal/cli/ -count=1 -run 'TestTodoPR_QueueDirUnchanged'
--- PASS: TestTodoPR_QueueDirUnchanged (1.01s)   [6/6 sub-cases]

# 2. mutant planted at the head of runTodoPR: writes <root>/.moai/cache/landing-sweep.cache
$ go test ./internal/cli/ -count=1 -run 'TestTodoPR_QueueDirUnchanged'
--- FAIL: TestTodoPR_QueueDirUnchanged (1.20s)
            .moai/cache/landing-sweep.cache 5efb444f3be96604690e318922c6485e84a8e24dca1f36849c4f2fc870feb15d
            [the same line in all six sub-cases' after-listings]

# 3. reverted
$ shasum internal/cli/todo_pr.go
a3292e44fe1b37df1cb7c2eb58db537f6754f615        # identical; git status clean

# 4. re-run
ok  	github.com/modu-ai/moai-adk/internal/cli	1.929s
```

Step 1 is what excludes a vacuously-red criterion and is the reason the sequence is ordered.

**Two failed attempts are part of this record.** My first two mutants did not compile
(`resolveTodoQueueRoot()` returns one value; I assigned two). `FAIL [build failed]` is the test not
running, not the test catching anything — reading it as a catch would have made the whole mutation
exercise vacuous, in exactly the way this card's own defect class works. I fixed the mutant and
observed a real RED.

### Run-phase

Seven commits, all naming t331: M1 `260ea5369` → M2 `9ba33d0a2` → M3 `61424aed0` → M4 `9414374b4`
→ AC-TLS-008 widening `f10827fd3` → M5 `5be48b3f8` → §E.2/§E.3 `58fb7cf82`.

**10 acceptance criteria, 10 PASS, 0 FAIL**, counted against `acceptance.md`. Agent-measured:
`internal/kanban` 14.4s · `internal/cli` 259.8s · `internal/template` 23.8s · `internal/config` 2.5s
all ok; `golangci-lint` 0 issues; `GOOS=windows GOARCH=amd64 go build ./...` rc 0; coverage kanban
87.1%, cli 79.9%.

Lane re-measurement after each absorb of `origin/develop`:

```
$ go build ./...                                        → exit 0
$ go vet ./internal/kanban/... ./internal/cli/          → exit 0   (compiles tests too)
$ go test ./internal/kanban/... -count=1                → ok 13.188s
$ go test ./internal/cli/ -count=1 -run 'TestTodoPR|TestTodoDone|TestLandedRef'  → ok 10.654s
$ diff .claude/skills/moai/workflows/todo.md internal/template/templates/.claude/skills/moai/workflows/todo.md
                                                        → rc 0 (mirror byte-identical)
```

### Sync-phase

`c9f712232` (3-phase close: CHANGELOG, docs-site 4 locales, `spec.md` `in-progress → completed`,
`progress.md` §E.4) then `fee6c22d9` (`sync_commit_sha` backfill).

```
$ grep -n 'sync_commit_sha' .moai/specs/SPEC-TODO-LANDING-STATE-001/progress.md
190:sync_commit_sha: c9f712232   # backfilled (a commit cannot name its own SHA)
$ grep -rn 'pending-backfill' .moai/specs/SPEC-TODO-LANDING-STATE-001/ | wc -l
       0
```

The slot holds a real commit SHA, matching the sync commit. No prose placeholder: another card in
this batch wrote `pending-backfill-sync` into that slot and became the live example of the defect a
separate card exists to prevent.

The docs claim the CLI prints one verdict token on every successful `done`. Verified against source
rather than against the docs' own wording: `internal/cli/todo.go:418` —
`fmt.Fprintf(cmd.OutOrStdout(), "done %s landing=%s\n", id, verdict)`.

`todo pr` is documented on no docs-site page in any locale, so no section was invented for it. The
four `moai-todo.md` locales each received the same one-row edit.

## Baseline-attribution

Every lane figure was measured in this worktree, in this run. `origin/develop` was absorbed three
times as it moved (`87b16c345` before run, `d3833fc20` before sync, plus an earlier absorb at plan)
and the affected-package suites were re-run on each merged tree rather than carried across the
merge — two separately-green branches can be red together, which this batch has already produced
once. Ref-dependent figures are attributed to the ref SHA and the observation instant, never to a
tree SHA alone.

## Gaps — what was NOT observed

- **No CI verdict.** Nothing is pushed to `develop`, and this repository's workflows trigger on
  pushes to `main`/`develop` only, so a card branch produces no run. The repository-wide test
  verdict is PENDING and belongs to CI on the integration branch. The local full suite was NOT run
  (standing contract).
- **`internal/cli` coverage 79.9%, below the 85% target, and unattributed.** No pre-change baseline
  was measured, so it cannot be claimed this change caused it or avoided it. Stated rather than
  omitted.
- **`GOOS=windows go vet` not run.** The Windows *build* passes; Windows *test compile* is unverified.
- **No `moai todo pr` executed against this repository's own live queue.** Every landing-answer
  assertion runs against fixtures, so the corrected ref's live behaviour here is unobserved.
- **The M1/M2 milestone-boundary deviation transcript was not re-verified by the sync agent** — only
  the pointer from §E.4 to §E.2 was confirmed to resolve.
- **No sync-auditor pass.** This verdict is the lane's own reading.

## Residual risk

- **Approved and deliberately open**: a mutant writing OUTSIDE the fixture root — `$HOME`, `/tmp`, a
  global cache — still evades AC-TLS-008. The criterion is drawn at *the verb's reach within the
  project*, not at *this verb writes nowhere*. The operator approved kickoff knowing this; the
  wording was kept accurate rather than quietly widened.
- **Two machine-readable contract changes ship here**: a fifth `todo pr` outcome kind (`unknown`,
  reaching `--json`) and the row going from five tab-separated columns to six. In-repo consumers
  measured at zero; **external consumers of `todo pr --json` cannot be grepped and are genuinely
  open.**
- **Milestone-boundary deviation**: M1 and M2 could not be staged separately — replacing `Landed`'s
  `(bool, error)` with the three-valued answer is one compile-forced edit that drags the fifth
  outcome kind with it. M2's criteria were therefore GREEN on arrival and are committed as
  regression guards; the flipped RED for that axis is M1's reversed test. Recorded in §E.2, in M2's
  commit message, and pointed at from §E.4.
- **`resolveTodoQueueRoot()` now runs at command-construction time** to build help text — cheap and
  test-covered, but a resolution that previously happened only at execution time.
- **The landed answer still rests on a discipline, not a mechanism.** The grep survives squash only
  because this project requires the card id in every commit message. A squash commit that drops the
  id reads as not-landed, and nothing here detects that.

## Process record — one aborted delegation

The first run-phase delegation stalled with no progress for 600s and returned nothing. It was NOT
treated as a return: the tree was measured (HEAD unchanged at `87b16c345`, zero commits, zero
uncommitted changes), the absence of partial state established, and the work re-delegated clean with
the reading order changed and an instruction to commit a small first milestone rather than keep
reading. The second attempt completed.

## Commits on `WT-card-landing-state`

| SHA | Phase |
|---|---|
| `e497c3dba` | plan-phase artifacts |
| `11426a128` | plan-audit iter-1 verdict (FAIL 0.74) |
| `e1d480eba` | iter-1 remediation — scope split to half A |
| `45cff0f59` | plan-audit iter-2 verdict (PASS 0.85) |
| `445c3f7d8` | iter-2 delta fix — v0.3.0 |
| `260ea5369` … `58fb7cf82` | run-phase M1-M5 + evidence (7 commits) |
| `c9f712232` | sync-phase 3-phase close |
| `fee6c22d9` | `sync_commit_sha` backfill |

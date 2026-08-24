# PR #1638 — CodeRabbit review response (t209)

Card t209, SPEC-WORKTREE-REAPER-001, branch `WT-worktree-reaper`.

Two review rounds have landed, on heads `55ecba718` and `b72e8265d`. This file
records every comment and what was done with it.

**A note on the second round's inline comments.** All 21 inline comments on
`b72e8265d` are the SAME comments from the `55ecba718` round, re-anchored to
new line numbers — GitHub carries unresolved threads forward. Five of them
(stderr, `%w` wrapping, the Windows probe, the docs predicate, the fixture
comment) name defects already fixed in `b72e8265d` itself; they persist because
the threads were never resolved, not because the finding recurred. The genuinely
new material in the second round is the two **outside-diff** comments, which is
what this round acts on.

---

## Round 2 — outside-diff findings

### OD-1 — recheck ignored content immediately before each removal — **APPLIED**

> *Major.* The ignored-content verdict runs during classification, but the later
> `Remove(..., false)` call uses that old result. If another process creates
> `.claude/agent-memory/` or another irreplaceable ignored path after
> classification, Git can still remove the worktree because ignored files do not
> trigger non-forced removal refusal.

Verified against the tree before acting, and the finding is correct for exactly
one of the two paths:

| Path | Shape | Exposed? |
|---|---|---|
| `--stale` | classifies the WHOLE population, then removes in a second loop | **Yes** — a tree cleared early can acquire content before its turn |
| `--merged-only` | classifies and acts one tree at a time | **No** — the verdict at `clean.go:137` already sits immediately before `Remove` at `:148` |

So the fix is one re-read in the `--stale` removal loop, plus a comment on the
merged limb recording that its guard's POSITION is load-bearing — hoisting it
into a classification pass would reintroduce exactly this defect there.

The window is **narrowed, not closed**: a gap remains between the second read
and the removal syscall. Closing it entirely is not available to this design —
nothing downstream refuses on ignored content, which is the whole reason the
guard exists.

**RED evidence (mutation).** With the re-read's condition forced false and
everything else identical:

```
--- FAIL: TestCleanStale_RechecksIgnoredContentBeforeRemoval (0.00s)
    clean_ignored_content_test.go:264: content that appeared after classification
      must still preserve the tree; removed [/wt/race]
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	1.019s
```

**GREEN, restored:**

```
=== RUN   TestCleanStale_RechecksIgnoredContentBeforeRemoval
--- PASS: TestCleanStale_RechecksIgnoredContentBeforeRemoval (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	1.452s
```

The test drives the race directly: the stubbed `--ignored` read returns clean on
its first call and `!! .claude/agent-memory/` on every later one, and the test
asserts the guard ran at least twice, that nothing was removed, and that the
notice names `cause=ignored-content`.

### OD-2 — propagate lock-source failures as system errors (exit 2) — **DECLINED**

> *Major.* Both paths only print a warning and return success, so automation
> cannot distinguish a fail-closed sweep from completed cleanup. Return a wrapped
> system error and verify that the command exits with code 2.

Declined on the premise, which was checked rather than assumed. **Automation
already can distinguish the two**, by the surface built for it in this same SPEC:

- `clean --stale --json` reports every record with `anchored: "undetermined"` and
  a `keep_reason` carrying `cause=lock-source-unreadable` (REQ-WR-012 — the
  inventory and the sweep are one evaluation, so the report cannot describe a
  degraded run as a clean one).
- The human paths print the same cause-bearing notice, and as of `b72e8265d` it
  goes to **stderr**, which is where a script looks for a degraded-run warning.

Two further reasons not to take it inside this card:

1. **It contradicts a requirement this SPEC states.** REQ-WR-016: *"The sweep
   shall not abort its caller on any failure; every failure path shall remain a
   non-blocking notice."* Two criteria assert it by name. Reversing that is a
   requirement change, not a review fix, and it belongs where it can be argued
   with its own acceptance criterion.
2. **Exit 2 is not reachable by returning an error.** `cmd/moai/main.go:13-22`
   exits **1** for any error unless it satisfies `cli.ResolveExitCode` — so
   "return a wrapped system error" yields exit 1, not the 2 the finding asks to
   verify. Delivering it needs the `ExitCoder` plumbing, which widens the change
   further.

Recorded as a follow-up card candidate: *"`moai worktree clean` should exit 2
when its authoritative anchor source is unreadable"*, carrying the REQ-WR-016
amendment and the `ExitCoder` wiring together.

---

## Round 1 — inline comments (head `55ecba718`)

### Applied in `b72e8265d`

| # | Location | Finding | Disposition |
|---|---|---|---|
| 1 | `clean.go` (both sweeps) | degraded-run notice written to stdout | Applied — `internal/cli/CLAUDE.md` reserves stdout for machine-readable output. Note: the `--json` path never printed the notice, so no JSON was being corrupted; this is the convention, not a live corruption |
| 2 | `clean.go` `worktreeLockStates` | raw error returned, not wrapped | Applied — `fmt.Errorf("read worktree lock state: %w", err)` |
| 3 | `anchor_pid_windows.go` | probe returns `(true, true)` for a process the platform never observes | Applied — now `(false, false)`. The guard treats undetermined as anchored, so the preserve outcome is unchanged; what changes is that the notice stops reading "locked by live pid N" about an unmeasured pid. `design.md`'s probe table corrected to match |
| 4 | `docs-site` ×8 + CHANGELOG | inventory described three predicates after `ignored` became the fourth | Applied — 4 locales × (cli-reference + guide), plus the CHANGELOG bullet |
| 5 | `session_worktree_prmerge_lock_test.go` | fixture comment says "gh gives NO answer" while the fixture returns `MERGED` | Applied — comment corrected, fixture left alone (the branch IS a candidate through the gh path, determinately) |

### Declined

| Location | Finding | Reason |
|---|---|---|
| `investigation.md`, `plan-audit*.md`, `sync-audit*.md` | various internal inconsistencies | **Point-in-time records.** They state what was observed when they were written. Editing them to agree with today's tree would make the record false — the audit trail's value is that it was not retrofitted |
| `progress.md` count reconciliation | "25 requirements but 24/24 covered"; "28 criteria but 32 test-shaped" | **Partially applied, and deliberately not reconciled.** 25/25 was measured and corrected. The `32` was NOT adjusted: it does not reproduce from the ID set (the enumeration lists 25 IDs), and `criteria_gated: [AC-WR-023b]` names an ID absent from `acceptance.md`. Writing a tidier number without measuring it is the failure this SPEC spent three audit rounds on, so the discrepancy is recorded as unattributed and left for a follow-up |
| `acceptance.md` criterion-command fixes (3) | commands that under- or over-count | **Post-hoc criterion edits.** These criteria were run in their recorded form and their `Pre-impl observed:` values were reproduced by the auditor. Rewriting the command now would orphan the observation it produced |
| `session_worktree_prmerge.go:189` | `%v` exposes raw upstream errors in notices | **Not taken here.** The cause token is the stable part and is already machine-readable; the raw error is what makes a refusal diagnosable at all, and the notices are developer-facing sweep output rather than an end-user error surface. Worth revisiting alongside a logging seam, which this card does not have |
| test-hygiene items (4) — helper extraction, comment-assertion coupling, hand-rolled `strconv.Itoa`, porcelain stub shape | Trivial / Low value | Out of scope for a review-response round; none affects behaviour or coverage |

---

## Verification (this round)

```
$ go build ./...                          → exit 0 (no output)
$ go vet ./...                            → exit 0 (no output)
$ go test -count=1 ./internal/cli/worktree/... ./internal/session/...
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	9.515s
ok  	github.com/modu-ai/moai-adk/internal/session	17.566s
```

## Gaps

1. **The full suite was not run locally** — repo rule. CI owns that verdict
   against the PR head.
2. **`internal/cli` was not re-run** — this round does not touch it; its green is
   the attributed prior measurement from `progress.md` §E.2 (`ok … 768.952s`,
   exit 0, at `aa14918d7`).
3. **No sweep was run against a real repository** — `clean` prunes shared
   worktree administrative state; every path here is bound by stubbed-git
   fixtures.
4. **The race is narrowed, not closed** — see OD-1. A gap remains between the
   second read and the removal.
5. **No `@coderabbitai review` was requested.** The chain is rate-limited and the
   re-review timing is the lead's to schedule.

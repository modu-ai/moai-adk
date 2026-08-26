# Sync-phase delta re-audit (iteration 3) — SPEC-WORKTREE-REAPER-001 (card t209)

- Position verified: `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t209`; `git rev-parse --short HEAD` → `73b287a41`; branch `WT-worktree-reaper`; `git status --short` → empty. Matches the dispatched HEAD.
- Baseline: `.moai/reports/t209/sync-audit-iter2.md` — **FAIL, harmonic mean 0.844** at `96a26b0f8` (F 0.88 / S 0.82 / C 0.86 / Cons 0.82), Security tripping the must-pass firewall on N1.
- Scope: the delta `96a26b0f8..73b287a41` — `14e60f382` (guard share) + `73b287a41` (§E.6 evidence). Source diff is 3 files, +117 / −6. F1/F2/F4 were spot-checked only, per dispatch.
- Method: closure judged by reading the code and running the tests. `progress.md` §E.6 was read for disclosure quality, and is cited as an artifact under audit — never as evidence for a behavioural claim.

## Verdict: **PASS** — harmonic mean **0.887** against threshold 0.85

Must-pass firewall clears independently: Functionality 0.92 ≥ 0.85, Security 0.88 ≥ 0.85.

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 0.92 | PASS | `go build ./...` → `build_rc=0`; `go vet ./...` → `vet_rc=0`; `go test -count=1 ./internal/session/... ./internal/cli/worktree/...` → `ok … internal/session 10.262s` / `ok … internal/cli/worktree 6.961s`, `rc=0`; targeted verbose run → **20 `--- PASS`, 0 `--- FAIL`**, including all three new `TestCleanMergedOnly_*` ignored-content cases; `go test -run TestPRMergeCleanup ./internal/cli/` → `ok … 1.790s`; `go test -run TestSessionWorktree ./internal/cli/` → `ok … 0.723s`; `go test -run TestParseGitBranchMergedOutput ./internal/cli/` → `ok … 0.826s` (F1 spot-check). |
| Security (25%) | 0.88 | PASS | N1 closed: the merged sweep consults the shared decision before `Remove` (`clean.go:133`), fails closed on an unreadable status, and exactly one allowlist exists (`grep -rn RegenerableIgnoredPaths internal/` → definition + one consumer, all in `internal/session/ignored_content.go`). Deducted for N5 (`worktree done --auto` / `worktree remove` remove trees without this guard — out of SPEC scope but reachable unattended) and N7 (the exec seam has no timeout and discards stderr, degrading the fail-closed notice to `exit status 128`). |
| Craft (20%) | 0.88 | PASS | `GOOS=windows go vet ./...` → `winvet_rc=0`; **`golangci-lint run ./internal/cli/worktree/... ./internal/session/...` → `0 issues.`** (the limb iteration 2 left unmeasured); `gofmt -l internal/cli/worktree internal/session` flags only 4 files untouched by this branch (`git diff --stat origin/main..HEAD -- internal/session/` does not list them). The fixture change is disclosed with verbatim failure output in §E.6. Deducted for N6 (stub breadth + `defer`-vs-`t.Cleanup` inconsistency) and N2 still open. |
| Consistency (15%) | 0.87 | PASS | The `cause=` vocabulary is shared verbatim across all three sweeps (`causeIgnoredContent` / `causeIgnoredCheckFailed` at `clean.go:426-429`, matching `session_worktree_prmerge.go:262,265`); template mirror `diff -q` on `worktree-integration.md` → `MIRROR_IDENTICAL`; the guard is a delegation, not a third copy. Deducted for the `three removal paths` wording against five `Remove` call sites (N5), and F7 / F11 still standing. |

Harmonic mean: `4 / (1/0.92 + 1/0.88 + 1/0.88 + 1/0.87)` = `4 / 4.509109` = **0.8871**. (Unweighted harmonic mean — the same method used in the two prior iterations.)

---

## 1. N1 — **CLOSED**

### The predicate is factored, and the merged path consults it

`ignoredContentVerdict(path) (state, keepReason string)` at `internal/cli/worktree/clean.go:413-423` is now the single evaluation. It is called from exactly two sites (`grep -rn ignoredContentVerdict internal/`):

| Site | Path | Behaviour |
|---|---|---|
| `clean.go:133` | `clean --merged-only` | `if _, reason := ignoredContentVerdict(wt.Path); reason != "" { print keep; continue }` — placed **after** the anchor guard and **before** `WorktreeProvider.Remove(wt.Path, false)` at `:138` |
| `clean.go:400` | `clean --stale` (and `--stale --json`, same evaluation) | `state, reason := ignoredContentVerdict(path); c.Ignored = state; return reason` |

Both directions verified by reading `ignoredContentVerdict`:

- **Irreplaceable content cannot be removed** — a non-empty `session.IrreplaceableIgnoredEntries(porcelain)` returns `(staleStateYes, "cause=ignored-content; …")`, a non-empty reason, so the merged loop `continue`s before `Remove`.
- **An unreadable status cannot be removed** — the `err != nil` branch returns `(staleStateUndetermined, "cause=ignored-check-failed; …")`, also a non-empty reason. Fail-closed holds; there is no path on which an unobserved predicate yields an empty reason.

### `--stale` behaviour is unchanged by the refactor

The extraction is behaviour-preserving by inspection of `git show 14e60f382 -- internal/cli/worktree/clean.go`: the three return branches are transplanted verbatim, with `c.Ignored = <state>` moved from inside each branch to the single assignment `c.Ignored = state` at the call site. The same three `(state, reason)` pairs are produced in the same order for the same inputs. Pinned end-to-end by the four pre-existing `--stale` limbs, all passing: `TestCleanStale_KeepsIrreplaceableIgnoredContent`, `_RemovesWhenIgnoredContentIsRegenerable`, `_UnreadableIgnoredStatusPreserves`, `_JSONReportsIgnoredPredicate` (the last asserting `Ignored == "yes"` in the JSON record, so the predicate-recording move is bound, not just the reason string).

### Exactly one allowlist

```
$ grep -rn "RegenerableIgnoredPaths" internal/ | grep -v _test
internal/session/ignored_content.go:18   (doc)
internal/session/ignored_content.go:31   var RegenerableIgnoredPaths = []string{
internal/session/ignored_content.go:46   (doc)
internal/session/ignored_content.go:71   for _, p := range RegenerableIgnoredPaths {
```

One definition, one consumer, no second copy anywhere. The other `.moai/state` literals in the tree (`internal/web/events.go:27`, `internal/worktree/state_guard.go:27`, `internal/sandbox/profile.go:56`, …) belong to unrelated concerns and are not ignored-content allowlists.

### Test binding is non-vacuous

The three new merged-only cases drive `runClean` through the cobra command with `merged-only=true` — not `ignoredContentVerdict` directly — so deleting the six-line guard from `cleanMergedWorktrees` fails `TestCleanMergedOnly_KeepsIrreplaceableIgnoredContent` and `_UnreadableIgnoredStatusPreserves`. `_RemovesWhenIgnoredContentIsRegenerable` is the anti-immortality control in the opposite direction and would catch a blunt "holds ignored content → keep" implementation.

## 2. The doc claim is now true

`internal/session/ignored_content.go:50-51`:

> `@MX:ANCHOR: [AUTO] the sole ignored-content predicate behind every sweep`
> `@MX:REASON: three removal paths consult it; …`

Counted:

```
$ grep -rn "IrreplaceableIgnoredEntries" internal/ | grep -v _test
internal/cli/session_worktree_prmerge.go:264   ← PR-merge sweep
internal/cli/worktree/clean.go:418             ← ignoredContentVerdict
internal/session/ignored_content.go:45,52      ← definition
```

Two call sites of the session function, but `clean.go:418` is reached from **two** removal paths (`clean.go:133` merged, `clean.go:400` stale). Sweeps consulting the decision: PR-merge, `--stale`, `--merged-only` = **3 of 3**. Under the "sweep" reading the header comment at `:3-4` uses, both claims hold. See N5 for the one wording reservation.

## 3. The `TestRunClean_MergedOnly` change is a **legitimate fixture repair**, not a weakened test

Judged against four questions:

1. **Was the failure an environment artifact or a defect?** Environment. The fixture worktrees are `/repo` and `/repo-feature`, which do not exist on disk. Before this change the merged sweep ran no per-path git command, so the fixture was never exercised against the filesystem; the guard introduced the first one. §E.6 records the verbatim failure — `Keeping /repo-feature [feature]: cause=ignored-check-failed; could not read ignored content: exit status 128` — which is the guard behaving **correctly** on an unrunnable path, not a defect being papered over.
2. **Was production behaviour relaxed?** No. The diff to `clean.go` adds a guard and never removes or weakens one; the only test-side change is a `gitWorktreeCmd` stub scoped to argv containing `--ignored`.
3. **Did the test lose its assertion?** No. `removedPath == "/repo-feature"` and the `Removing merged worktree` output assertion are both intact — it still asserts the removal path, which is what it was written for.
4. **Did mutation coverage of the guard drop?** No — it went up. The guard's presence in `cleanMergedWorktrees` was not covered by any test before this commit; it is now covered by three, all driving `runClean`. A bad repair would have stubbed the guard away *and* left it unpinned elsewhere; here the fail-closed branch that the stub bypasses is bound by `TestCleanMergedOnly_UnreadableIgnoredStatusPreserves`, which asserts the same `cause=ignored-check-failed` token that the stub suppresses.

The alternative repair — creating real temp directories with `t.TempDir()` and `git init` — would have been strictly better isolation, but it is a larger change to a legacy fixture that also stubs `WorktreeProvider`, and the chosen stub does not lose a behaviour. Verdict: legitimate. Two residual nits are recorded as N6, both optional.

## 4. The PR-merge sweep is **untouched**

```
$ git diff 96a26b0f8..73b287a41 --stat -- internal/
 internal/cli/worktree/clean.go                     | 34 ++++++++--
 internal/cli/worktree/clean_ignored_content_test.go| 73 +++++++++++++++++++
 internal/cli/worktree/subcommands_test.go          | 16 +++++
```

`internal/cli/session_worktree_prmerge.go` is absent from the delta; its last touch is `2045acc92` (the iteration-1 repair). `ignoredContentAllowsRemoval` (`:256-268`) keeps its own `sessionWorktreeGitStatusIgnored` seam and its own notice formatting, delegating only the classification — the design deliberately shares the *decision*, not the plumbing. Behaviour re-confirmed: `go test -run TestPRMergeCleanup ./internal/cli/` → `ok`, `go test -run TestSessionWorktree ./internal/cli/` → `ok`.

## 5. F1 / F2 / F4 spot-checks — all still hold

- **F1**: the parser is untouched by the delta; `TestParseGitBranchMergedOutput` → `ok`.
- **F2**: the `lockErr != nil` early return in `cleanMergedWorktrees` (`clean.go:96-104`) is above the new guard and unchanged; `TestCleanMergedOnly_UnreadableLockSourceRemovesNothing` and `TestCleanStale_UnreadableLockSourceRemovesNothing` both `--- PASS`. Ordering is correct: the lock read still precedes every per-tree evaluation, so an unreadable lock source cannot be masked by an ignored-content keep.
- **F4**: `isBaseBranch` untouched; `TestCleanStale_KeepsWorktreeOnBaseBranch` `--- PASS`.

---

## New findings

### N5 — [Medium] [optional] `internal/cli/worktree/done.go:84,182` + `internal/cli/worktree/remove.go:60` — two further removal paths do not consult the shared decision, and the `@MX:REASON` wording does not bound itself to sweeps

`grep -rn "\.Remove(" internal/cli/worktree/*.go | grep -v _test` returns **five** call sites, not three:

| Site | Command | Anchor guard | Ignored-content guard |
|---|---|---|---|
| `clean.go:138` | `clean --merged-only` | yes (shared `AnchorDecision`) | **yes** (new) |
| `clean.go:254` | `clean --stale --yes` | yes (shared) | yes |
| `done.go:84` | `worktree done [--auto]` | `session.LiveAnchoredSessions` (registry only) | **no** |
| `done.go:182` | `worktree done` (direct limb) | as above | **no** |
| `remove.go:60` | `worktree remove <path>` | `session.LiveAnchoredSessions` | **no** |

`done --auto` is the one that matters: its own flag help says *"no success output for automation (e.g., after PR merge)"* and `AutoCleanupFlag` is documented as *"Used by sync workflow to trigger cleanup after PR merge"* (`done.go:37-39`). That is an **unattended** removal of a named tree, and it destroys `.claude/agent-memory/` on exactly the mechanism this SPEC measured (`design.md` §A.6) — non-forced `git worktree remove` disregards ignored files.

Two reasons this is **not** a blocker and not a re-run of N1: (a) `done` and `remove` name one tree the caller asked to dispose of, rather than sweeping an inventory, so the intent-to-destroy is explicit; (b) REQ-WR-024 scopes the guard to the sweeps, and the header comment at `ignored_content.go:3-4` says *"shared by every **sweep** that removes a worktree"* — which is true at 3/3. The reservation is only that `:51`'s *"three removal paths consult it"* uses the looser noun, and a reader counting `Remove(` call sites finds five.

**Required fix (cheap):** change `:51` to *"three sweeps consult it"* and add one clause naming `done` / `remove` as explicit, single-target disposals deliberately outside the share. **Follow-up card (larger):** decide whether `done --auto` — an unattended path — should consult the guard, or whether drain-then-dispose (REQ-WR-025) subsumes it.

### N6 — [Low] [optional] `internal/cli/worktree/subcommands_test.go:645-658` — the fixture stub is broader than its stated purpose and diverges from its sibling's restore idiom

Two nits on the otherwise legitimate repair of check 3:

1. **Breadth.** The stub matches *any* argv containing `--ignored` and returns clean. Today only `ignoredContentVerdict` passes that flag, so the effect is exactly as documented — but a future `--ignored` read added to this path would be silently stubbed clean by a test that never intended to cover it. `if len(args) >= 4 && args[2] == "status"` would pin the shape.
2. **Restore idiom.** It uses `defer func() { gitWorktreeCmd = origGitCmd }()` while every stub in the sibling `clean_ignored_content_test.go` uses `t.Cleanup`. Both work here (no `t.Parallel()` anywhere in the package — `grep -c "t.Parallel()" internal/cli/worktree/` → `0`), but the mixed idiom is a consistency wart, and `defer` inside a test that `t.Fatalf`s in a loop body is the shape that eventually leaks a global.

Also note the test still delegates non-`--ignored` calls to the **real** git in the live repository (`worktreeLockStates` runs `git worktree list --porcelain`), making it environment-dependent. Pre-existing since the anchor guard landed, not introduced here.

### N7 — [Low] [optional] `internal/cli/worktree/shared.go:14-17` — the exec seam has no timeout and discards stderr, so the new fail-closed notice cannot say why

```go
var gitWorktreeCmd = func(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	return string(out), err
}
```

Two consequences, both now reached by one more path:

- **No context / timeout.** A hung `git status --porcelain --ignored` blocks `clean --merged-only` indefinitely. Unlike `internal/core/git/manager.go`, which bounds git at `foundation.DefaultGitTimeout` (5s), this seam is unbounded. Iteration 2's N4 recorded the same class of load-sensitivity from the other direction.
- **Stderr discarded.** `.Output()` returns an `*exec.ExitError` whose `%v` renders `exit status 128`, so the cause-bearing notice degrades to `cause=ignored-check-failed; could not read ignored content: exit status 128` — visible verbatim in §E.6's own RED transcript. The guard's whole point is that the operator can read *why* a tree was kept; "128" is not a why. `ExitError.Stderr` is already populated by `.Output()` and would cost one `errors.As`.

Measured cost of the added read, so this is a correctness/diagnostics finding and not a performance one: `time git status --porcelain --ignored` in this worktree → **0.142s total**, 13 `!!` entries, against 0.118s for the plain status.

### N8 — [Info] — §E.6's disclosure of the fixture change is the reason check 3 was cheap to settle

Recorded as a positive, since the audit protocol asks the finding stage for coverage rather than only defects: §E.6 states the fixture movement, quotes the verbatim failure it caused, names the test that now carries the fail-closed branch instead, and asserts *"Production behaviour was not relaxed to make a test pass."* Every one of those four claims was independently verified above and all four hold. One §E.6 gap is now closed by this audit rather than by the author: **`golangci-lint` was run** — `0 issues.` on both changed packages.

## Findings carried forward (unchanged, all optional)

**N2** (`TrimLeft(…, "*+ ")` cutset + `(`-skip; zero measured exposure), **F5** / **N3** (`.moai/state` allowlisting, now test-pinned), **F7** (undisclosed `Prune()` on the `--json` path), **F8**, **F9**, **F10**, **F11** (`workflow.yaml:35` still contradicts `session_worktree_prmerge.go:150`), **F12**. **F6** remains closed. None blocking.

---

## Gaps — what this audit did NOT observe

1. **No mutation test was executed.** The claim that deleting the guard from `cleanMergedWorktrees` fails the three merged-only tests is derived from reading those tests (they drive `runClean`, not the helper), not from actually removing the guard and observing red. Source was read-only per dispatch.
2. **No sweep was run against a real repository.** `moai worktree clean --merged-only/--stale/--json` was prohibited. All three paths remain bound by stubbed-git fixtures; the destroy mechanism (non-forced `git worktree remove` disregards ignored files) is still taken from `design.md` §A.6's recorded measurement, not re-confirmed here.
3. **`-race` was not run** — outside the authorised command set. The new tests mutate the package global `gitWorktreeCmd`; no `t.Parallel()` exists in the package (measured: 0), which is why this is a gap rather than a finding, but the absence of a race is unverified.
4. **The 28-criterion AC battery was not re-executed.** Targeted `-run` subsets only, per dispatch: `TestCleanMergedOnly|TestCleanStale|TestRunClean_MergedOnly`, `TestPRMergeCleanup`, `TestSessionWorktree`, `TestParseGitBranchMergedOutput`. Criteria outside those patterns are unverified by me this iteration.
5. **`internal/cli` was not run in full** — prohibited. Its green is the author's attributed prior measurement from §E.2, not a measurement of mine at this HEAD.
6. **Windows runtime behaviour is unobserved.** `winvet_rc=0` proves cross-compilation only.
7. **No CI verdict was read.** The branch is unpushed; the clean-environment full-suite verdict is outstanding.
8. **N5's blast radius is uncounted.** How many trees `done --auto` has disposed of while holding irreplaceable ignored content is unknown — sibling worktrees cannot be inspected from this isolated session.
9. **`gofmt -l` flags 4 files in `internal/session`** (`checkpoint.go`, `registry.go`, `hydrate_test.go`, `store_test.go`). I established they are untouched by this branch (`git diff --stat origin/main..HEAD -- internal/session/` lists neither), so they are not attributable here — but I did not determine whether they are a toolchain-version artifact or a genuine pre-existing formatting drift on `main`.

## Residual risk

- **One evaluation now decides three paths.** The author names this in §E.6 and it is the correct reading: an error in `IrreplaceableIgnoredEntries` — a missing allowlist entry, or a `!! ` parse edge — is now wrong on all three sweeps simultaneously. The share is right; the concentration is real.
- **N5 is the same shape as N1, one scope-level out.** `done --auto` runs unattended after PR merge with no ignored-content guard. It is defensibly outside REQ-WR-024, but a reader of the `@MX:ANCHOR` could conclude otherwise, and that misreading is exactly what the annotation exists to prevent.
- **The check→act window is unchanged** and now spans a fourth predicate on the merged path.
- **The verdict rests on a read-only delta pass.** A full `internal/cli` battery or CI on a pushed head could surface criteria failing for reasons unrelated to N1's closure.

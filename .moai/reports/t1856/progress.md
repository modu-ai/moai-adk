# t1856 — progress record (Class B defect, cause known)

Tree: `.moai/worktrees/t1856`, branch `WT-10-17-jev`, base `e4b747678` (== local main, absorbed).
Commits: `899f3c0cd` (RED reproduction), `de719326c` (fix), `d8428f9b6` (fixture-email registration).
Scope: internal/hook only. No existing test weakened or deleted. No pushes.

## 1. Cause evidence (mechanism + pre-fix observation)

**Mechanism.** `srlDiffBase` (internal/hook/env_literal_diff_test.go, pre-fix lines 76-85)
resolved the diff base ONLY through the ref list `[]string{"develop", "origin/develop"}`:

```go
for _, ref := range []string{"develop", "origin/develop"} {
    if out, err := exec.Command("git", "-C", root, "merge-base", "HEAD", ref).Output(); err == nil {
        return strings.TrimSpace(string(out))
    }
}
```

`origin/develop` was deleted 2026-10-17; local `develop` is a stale branch (966f1b158) awaiting
operator deletion. Ref observations on this tree (pre-fix, e4b747678):

- `git rev-parse --verify --quiet develop` → `966f1b15852ed95d09ab32b7930142bbf1672ecb`
- `git rev-parse --verify --quiet origin/develop` → no output, exit 1 (ref deleted)
- `git merge-base HEAD develop` → `966f1b15852ed95d09ab32b7930142bbf1672ecb`

So the loop returns develop's merge-base — an old fork point — and `git diff <base> --
internal/hook internal/factorymsg internal/cli` sweeps everything that landed on main since that
point; every MOAI_KANBAN_* literal in that range reads as "added by this card".

**Pre-fix false-positive observation (E1), verbatim** — `go test ./internal/hook -run
'TestNoNewEnvLiteralsInDiff' -v` on the clean tree at e4b747678, before any edit:

```
=== RUN   TestNoNewEnvLiteralsInDiff
    env_literal_diff_test.go:52: env-literal sweep: 13271 added lines swept across internal/hook internal/factorymsg internal/cli (envkeys.go and _test.go excluded), base=966f1b15852ed95d09ab32b7930142bbf1672ecb, distinct literals=1
    env_literal_diff_test.go:59: diff introduces env-name literals (REQ-SRL-009 — use internal/config/envkeys.go constants): [MOAI_KANBAN_CARD]
--- FAIL: TestNoNewEnvLiteralsInDiff (0.26s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.028s
```

13271 added lines swept, base = the stale develop merge-base 966f1b158 — matches lane-22's
independent measurement (13256 added lines, base=966f1b158) and t1632 remeasure #7's A/B on
main fb749d372.

## 2. RED (E2) — verbatim, before the fix

The reproduction (internal/hook/env_literal_diff_base_test.go) builds the defect's precondition in
an isolated throwaway git repo under `/private/tmp/env-diff-base-*` (outside $HOME and outside the
repository; no card id in any path): a stale `develop` frozen at the fork point, `main` advanced
past it carrying a `MOAI_KANBAN_CARD` literal, and a card branch starting from main's tip. It
asserts the resolver returns main's merge-base, and separately pins the origin/main fallback where
local main is absent.

**Live RED** (pre-fix resolver; enabling change was the root parameterization only — ref list still
develop), `go test ./internal/hook -run 'TestSrlDiffBase' -v`, exit 1:

```
=== RUN   TestSrlDiffBaseResolvesFromMainNotStaleDevelop
    env_literal_diff_base_test.go:82: srlDiffBase picked 5a6d201dcf1d1fea375c6fe587c667f74f9621e0; want merge-base with local main 148d237068162591582433a1fe27404d9f88b3c6 (stale-develop false positive, card t1856)
--- FAIL: TestSrlDiffBaseResolvesFromMainNotStaleDevelop (0.84s)
=== RUN   TestSrlDiffBaseFallsBackToOriginMain
    env_literal_diff_base_test.go:99: srlDiffBase picked HEAD; want merge-base with origin/main 5a6d201dcf1d1fea375c6fe587c667f74f9621e0 (fallback, card t1856)
--- FAIL: TestSrlDiffBaseFallsBackToOriginMain (0.28s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	2.063s
```

**Reconstructed RED** (after a mid-run API 429 interruption, at the coordinator's request: the one
ref-list line was momentarily restored to the pre-fix develop form on the post-fix tree and the
tests re-run; temp-repo SHAs differ because each run builds a fresh repo), exit 1:

```
=== RUN   TestSrlDiffBaseResolvesFromMainNotStaleDevelop
    env_literal_diff_base_test.go:82: srlDiffBase picked 61bfe226acc848725d1e636e33a06ba251188f3a; want merge-base with local main f43264b9d47226f9d4e7b14189fb6e2a250428b2 (stale-develop false positive, card t1856)
--- FAIL: TestSrlDiffBaseResolvesFromMainNotStaleDevelop (0.28s)
=== RUN   TestSrlDiffBaseFallsBackToOriginMain
    env_literal_diff_base_test.go:99: srlDiffBase picked HEAD; want merge-base with origin/main 61bfe226acc848725d1e636e33a06ba251188f3a (fallback, card t1856)
--- FAIL: TestSrlDiffBaseFallsBackToOriginMain (0.12s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.044s
```

Both reds are red for the stated reason: the resolver picked develop's merge-base instead of
main's, and returned HEAD instead of the origin/main fallback.

## 3. Fix (minimal)

One line in `srlDiffBase` (plus its doc comment), commit `de719326c`:

```go
for _, ref := range []string{"main", "origin/main"} {
```

Local `main` first (agentic-trunk), `origin/main` fallback where local main is absent, develop path
removed entirely. The enabling root parameterization (`srlDiffBase(t, root)`) rides the RED commit
`899f3c0cd` with no behavior change. Commit `d8428f9b6` registers the reproduction's fixture email
in `builtinCommitIdentityDenyEmails` — the AC-CIG-010 drift test's prescribed remedy (its
enumeration caught the new `user.email` literal in the test file).

Diff stat vs base e4b747678 (after d8428f9b6):

```
 internal/hook/commit_identity_guard.go       |  3 ++-
 internal/hook/env_literal_diff_base_test.go  | 101 ++++++++++++++++++++++++++++++
 internal/hook/env_literal_diff_test.go       |  20 +++---
 3 files changed, 115 insertions(+), 9 deletions(-)
```

## 4. GREEN (E3)

**Targeted** — `go test ./internal/hook -run 'TestSrlDiffBase|TestNoNewEnvLiteralsInDiff' -v`,
exit 0:

```
=== RUN   TestSrlDiffBaseResolvesFromMainNotStaleDevelop
--- PASS: TestSrlDiffBaseResolvesFromMainNotStaleDevelop (0.78s)
=== RUN   TestSrlDiffBaseFallsBackToOriginMain
--- PASS: TestSrlDiffBaseFallsBackToOriginMain (0.31s)
=== RUN   TestNoNewEnvLiteralsInDiff
    env_literal_diff_test.go:52: env-literal sweep: 0 added lines swept across internal/hook internal/factorymsg internal/cli (envkeys.go and _test.go excluded), base=e4b747678aae096b22769fdfd3efc5d89498006a, distinct literals=0
--- PASS: TestNoNewEnvLiteralsInDiff (0.05s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	1.876s
```

The base is now e4b747678 (main merge-base) and the sweep is empty on the clean tree — expected:
HEAD == base. (The sweep-visibility log line records the empty sweep rather than passing silently;
the RED tests above are what exercise the resolver behavior.)

Targeted guard test after d8428f9b6 — `go test ./internal/hook -run
'TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration|TestSrlDiffBase|TestNoNewEnvLiteralsInDiff'`,
exit 0:

```
ok  	github.com/modu-ai/moai-adk/internal/hook	2.877s
```

**Full package (heavy run, inside the go-test-heavy slot).**

Run 1 (slot held by this session 16:19:03Z–release; tree at de719326c, before the guard-list
registration), `go test ./internal/hook`, exit 1 — one failure, the AC-CIG-010 drift test catching
this card's own fixture email (not the resolver):

```
--- FAIL: TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration (1.20s)
    commit_identity_guard_list_test.go:156: fixture email literals found in the repository's *_test.go but missing from builtinCommitIdentityDenyEmails (internal/hook/commit_identity_guard.go): [env-diff-base-test@example.invalid] — add them to the built-in list; do NOT weaken this test or change the enumeration predicate
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	294.932s
```

Remedy: commit `d8428f9b6` (the test's prescribed fix). All other tests in the package passed in
that run.

Run 2 (tree at d8428f9b6), `go test ./internal/hook -count=1 -v`, exit=0, lane-observed (slot
receipts in §5, detail and verbatim lines in §9). Final lines of the output file:

```
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	244.941s
```

## 5. Slot receipts (E5)

- Acquire attempt 1 (16:0xZ): exit 3 — held by session db74d25e (pid 29740, live, within its 20m
  bound ending 16:26:19Z). Waited out the bound (no concurrent heavy run).
- Acquire run 1: `slot go-test-heavy acquired by c419f651-cbd7-4942-a099-46fd57779863 until
  2026-10-17T16:39:03Z` — exit 0. Full run 1 executed inside it, then released:
  `slot go-test-heavy released (was c419f651-cbd7-4942-a099-46fd57779863)` — exit 0.
- Acquire attempt for run 2 (~16:25Z): exit 3 — held by lane-3 (t1833 landing remeasure, pid 21755,
  bound ending 16:49:28Z). Waiting out the bound; retry count for run 2: 1 of 3 used (the exit-3
  attempt), next attempt after 16:49:28Z.
- Run 2 receipts (lane-22, lane-observed): `moai slot status` before acquiring printed
  `no slot leases recorded` (2026-10-17T16:59:50Z); `slot go-test-heavy acquired by
  207b7b17-aa81-4311-8e47-c01448800dee until 2026-10-17T17:19:54Z` (exit 0, --max-duration 20m);
  released `slot go-test-heavy released (was 207b7b17-aa81-4311-8e47-c01448800dee)` (exit 0) at
  2026-10-17T17:04:17Z.

## 6. origin/develop survey (E6)

Command: `git grep -n 'origin/develop'` over the whole tree (tracked files, including
internal/template/templates). **1903 hits**; the full line list is preserved verbatim at
`.moai/reports/t1856/origin-develop-survey.txt` (local-only per the operator's card-reports
directive). Per-file summary:

- Overwhelmingly historical/evidence artifacts: `.moai/specs/**` (many SPECs' spec/plan/progress/
  acceptance), `.moai/reports/**` (verdicts, audit JSON), `reports/hooks-audit-20260911-01a08e35/
  cards-after.json` (37 hits), `CHANGELOG.md` (17).
- Go production source (comments only, no live ref resolution except where noted):
  internal/cli/ci_verdict.go:7, internal/cli/mcp_review_material.go:112,154-155,
  internal/cli/session_worktree.go:705,802 (comment says it REPLACES the literal
  refs/remotes/origin/develop), internal/cli/worktree/done.go:111,
  internal/closure/closuretest/closuretest.go:3,70 (fixture builder),
  internal/config/token_budget_guard.go:48 (comment), internal/factory/prlink_landed.go:264
  (comment describing normalization), internal/harness/rosterguard/registry.go:131 (comment),
  internal/spec/lint_movingref.go:6,100 (the moving-ref linter names origin/develop as a moving
  ref — that is its job, not a defect).
- Templates: internal/template/templates/.claude/rules/moai/core/verification-claim-integrity.md:46,
  .../workflow/worktree-integration.md:52, .../skills/moai/workflows/gtd.md:71 (worktree sweep
  description), internal/template/testdata/binding_ledger.json:3354-3355, plus template test
  fixtures.
- Test files referencing origin/develop (self-built fixtures, not the deleted remote):
  internal/factory/prlink_landed_forms_test.go (23), internal/factory/prlink_landedref_test.go (10),
  internal/cli/todo_autodone_test.go (13), internal/cli/worktree/sweep_base_fallback_test.go (9),
  internal/factory/autodone_scan_test.go (9), and ~30 more test files with 1-7 hits each (full list
  in the survey file).

Fixed in this card: only the internal/hook resolver. The rest is follow-up material — nothing else
was touched.

## 7. Class B completion statement

Cause established and mechanically confirmed on this tree before the fix (stale develop merge-base
966f1b158; 13271 swept added lines; false positive [MOAI_KANBAN_CARD]); RED reproduction observed
failing for the stated reason before the fix (live) and reconstructed after the 429 interruption;
minimal fix applied (main → origin/main → HEAD); RED green post-fix; existing test set preserved
(zero tests deleted or weakened; the one full-run failure was this card's own fixture email caught
by an existing guard, remedied by that guard's prescribed path). The t1632/t1837 blocker —
TestNoNewEnvLiteralsInDiff red on main — is resolved: on a clean tree the test now resolves
base=e4b747678 (main merge-base) and passes.

## 8. Gaps (E7)

- Full-package run 2 (post-guard-registration) completed: exit=0 on tree d8428f9b6 (§4, §9). It
  covers internal/hook only; see §9 Gaps for what it does not prove.
- The reconstructed RED (§2) was produced by momentarily restoring the pre-fix ref list on the
  post-fix tree — equivalent input to the live RED, but observed after the interruption, and is
  labeled as such.
- The sweep-empty GREEN of TestNoNewEnvLiteralsInDiff on this clean tree asserts nothing by itself
  (HEAD == base); the resolver behavior is carried by the two new tests.
- origin/develop survey covers tracked files only (`git grep`); untracked/ignored paths were not
  swept.

## 9. Lane re-observation and run 2 (lane-22, 2026-10-17)

Every item below is lane-observed (lane-22) and was transcribed from the named file by the
recording agent, which re-observed nothing itself. Source dir: `/private/tmp/t-sg/out/` (outside the
tree, machine-local scratch).

**Targeted GREEN at d8428f9b6** — `go test ./internal/hook -run
'TestSrlDiffBase|TestNoNewEnvLiteralsInDiff' -count=1 -v`, exit 0, file `green.txt`. The sweep line
read (verbatim): `env-literal sweep: 2 added lines swept across internal/hook internal/factorymsg
internal/cli (envkeys.go and _test.go excluded), base=e4b747678aae096b22769fdfd3efc5d89498006a,
distinct literals=0`. The earlier record (§4) said 0 lines; the difference is unexplained and the
lane did not investigate it.

**Pre-fix RED re-observed by the lane** — `git revert --no-commit de719326c`, then the same command,
exit 1, then `git revert --abort` (afterwards `git status --porcelain` 0 lines, HEAD d8428f9b6; the
revert was never committed). File `red.txt` carries the two TestSrlDiffBase* failures and:
`env-literal sweep: 13273 added lines swept across internal/hook internal/factorymsg internal/cli
(envkeys.go and _test.go excluded), base=966f1b15852ed95d09ab32b7930142bbf1672ecb,
distinct literals=1` / `diff introduces env-name literals (REQ-SRL-009 ...): [MOAI_KANBAN_CARD]`.

**Run 2, full package** — tree d8428f9b6, `go test ./internal/hook -count=1 -v`, files
`hook-full.txt` + `hook-full.exit` (`exit=0`). Final lines: `PASS` /
`ok  	github.com/modu-ai/moai-adk/internal/hook	244.941s`. Counts re-verified by the recording
agent with grep on the file: `grep -c '^--- PASS'` = 1538, `grep -c '^--- FAIL'` = 0,
`grep -c '^--- SKIP'` = 6 (column-0 lines only; indented subtests excluded). Skips, verbatim:
TestFactoryHookBenchmarkBudget, TestConsumerOnly_M0AndMxByteUnchanged,
TestSessionStart_DriftCacheProbe, TestDriftFillBurstHelper, TestSessionStart_HandleStageProbe,
TestScanWriteContent_SupportedExtension_ScannerNotAvailable. No -skip flag was passed; these are the
package's own skips, unrelated to this card. Named lines from the file:

```
--- PASS: TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration (0.99s)
--- PASS: TestSrlDiffBaseResolvesFromMainNotStaleDevelop (0.37s)
--- PASS: TestSrlDiffBaseFallsBackToOriginMain (0.14s)
--- PASS: TestNoNewEnvLiteralsInDiff (0.05s)
```

**Slot receipts** — see §5 (run 2 entry): acquired 207b7b17-..., released at 2026-10-17T17:04:17Z.

**Gaps**

- (a) The 2-vs-0 swept-lines difference for TestNoNewEnvLiteralsInDiff at HEAD d8428f9b6 is
  unexplained.
- (b) The full-package run proves only that no other internal/hook test regressed on tree
  d8428f9b6; it says nothing about other packages.
- (c) The landing re-measure (b2cd scope, including internal/cli per standing rule 35e9) has NOT run
  yet.
- (d) The origin/develop survey (§6) is follow-up material, untouched.
- (e) The recording agent re-observed nothing; it transcribed the files above (and re-counted
  PASS/FAIL/SKIP lines with grep).

## 10. Sync audit round 1 FAIL and the explicit wait (lane-22, 2026-10-17)

Sync audit round 1 returned FAIL (score 0.55; required backend codex also fail; file
`.moai/reports/t1856/sync-audit.md`, preserved in commit `6b97e64bc`). Blocking findings, each
confirmed by the lane against the source: F1 `internal/hook/env_literal_diff_base_test.go:52`
hard-codes `/private/tmp` while CI runs ubuntu-latest only; F2 the two new tests do not pin the
`main` -> `origin/main` order (mutants `{"main","develop","origin/main"}` and
`{"origin/main","main"}` pass them). Non-blocking: F3 literal-`develop` base lookups missing from
the survey, F4 survey recount (1903 vs 1919), F5 CHANGELOG sentence overstates. The resolver
change itself was judged correct. Repair is test code only and is not done yet.

Wait record `w-t1856-20261017T1750Z-c305` (explicit wait, no idle-spin):
- reason: the repair spawn (manager-develop) is refused by the spawn guard:
  `AUDIT_RECEIPT_VIOLATION: ... Outstanding: sync-auditor / unknown-spec / verdict line missing`.
  The refusal record is `.moai/state/audit-receipts/rejections/sync-auditor--unknown-spec.json`
  (agent_id ac98645f0cf74b59e, rejected_at 2026-10-17T17:12:20Z).
- whom: the leader. Standing rule `d-20261017T152443Z-c305` gives the leader, not a lane, the power
  to remove exactly one such record; the lane compared all five of its conditions against the
  observed record and sent the facts to the leader at 17:12Z.
- observed since: three watchdog channels unchanged from 17:31:11Z to 17:50:03Z (HEAD `6b97e64bc`,
  newest evidence mtime unchanged, integration window free); the decision board holds no record for
  this card newer than `d-20261017T165823Z-ea3d` (16:58Z). Ladder steps 3 (audit cross) and 4
  (jev_ask) do not apply: the open question is a guard state, not a verdict.
- recheck point: every standing recheck cron fire (renew both leases, then read the refusal record
  and the board). The lane does not delete the record, bypass the guard, or re-send the request.
- on resolution: spawn the repair (F1, F2 tests with mutant observation, F3-F5 records), re-run the
  targeted selector, re-audit with the required codex backend, then merge-ready and the landing
  re-measure (b2cd, rule 35e9 as clarified by `d-20261017T155926Z-148c`).

## 11. Sync audit round 1 repair (lane-22)

Audit: `.moai/reports/t1856/sync-audit.md` (FAIL, audited_sha 772c1c050). Repair tree HEAD before
this round: `89c89da80`; test-code commit: `1b7d86152`. The resolver `srlDiffBase` is unchanged
(ref list `{"main","origin/main"}`).

### Dispositions

- F1 repaired: `envDiffTempRepo` in `internal/hook/env_literal_diff_base_test.go` now uses
  `t.TempDir()`; the manual cleanup is gone and the two /private/tmp comments say the caller's
  TMPDIR carries the placement discipline. `grep -n 'private/tmp' internal/hook/env_literal_diff_base_test.go`
  prints nothing (exit 1).
- F2 repaired: two new cases, same helper style, same fixture email (no deny-list change; the
  AC-CIG-010 drift test `TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration` passes).
  `TestSrlDiffBasePrefersLocalMainOverStaleOriginMain` (kills D) and
  `TestSrlDiffBaseSkipsDevelopBetweenMainAndOriginMain` (kills C). Each asserts its fixture is
  discriminating (wanted base differs from the wrong ref's base) before judging the resolver.
- F3 recorded below (follow-up material only).
- F4 recounted below.
- F5 repaired: the CHANGELOG sentence now names four tests and what each pins.

### Mutant observation (edit of the ref list in srlDiffBase, restored with `git checkout -- internal/hook/env_literal_diff_test.go`; targeted selector `-run 'TestSrlDiffBase'`)

Mutant C `{"main","develop","origin/main"}`:
```
--- FAIL: TestSrlDiffBaseSkipsDevelopBetweenMainAndOriginMain (0.31s)
    env_literal_diff_base_test.go:139: srlDiffBase picked 167b9d23eb8ccff0a963589c344d8c3f2959984c; want merge-base with origin/main e96433700407581c4a968a69f64f80b40929b997, not stale develop 167b9d23eb8ccff0a963589c344d8c3f2959984c (card t1856)
FAIL	github.com/modu-ai/moai-adk/internal/hook	2.516s
```
Mutant D `{"origin/main","main"}`:
```
--- FAIL: TestSrlDiffBasePrefersLocalMainOverStaleOriginMain (0.26s)
    env_literal_diff_base_test.go:117: srlDiffBase picked e324a48e397ffcdbf9d46c225a7b90ee823b7624; want merge-base with local main 5d74389c68163a437f2e87615cd675827bf3abcb, not stale origin/main e324a48e397ffcdbf9d46c225a7b90ee823b7624 (card t1856)
FAIL	github.com/modu-ai/moai-adk/internal/hook	2.353s
```
Mutant A `{"develop","origin/develop"}` (the original defect), all four fail:
```
--- FAIL: TestSrlDiffBaseResolvesFromMainNotStaleDevelop (0.68s)
    env_literal_diff_base_test.go:78: srlDiffBase picked 048e5a52f01438863e3e1aea412bbf74831aef4d; want merge-base with local main d52be2a30fede466a30e4b850e229400dd95826e (stale-develop false positive, card t1856)
--- FAIL: TestSrlDiffBaseFallsBackToOriginMain (0.53s)
    env_literal_diff_base_test.go:95: srlDiffBase picked HEAD; want merge-base with origin/main 5b5afc5a9e85b9dee90471a189fa660b00170970 (fallback, card t1856)
--- FAIL: TestSrlDiffBasePrefersLocalMainOverStaleOriginMain (0.42s)
    env_literal_diff_base_test.go:117: srlDiffBase picked HEAD; want merge-base with local main 14ca9f0e4c00571ba35cd9171e391a1386d61682, not stale origin/main 5b5afc5a9e85b9dee90471a189fa660b00170970 (card t1856)
--- FAIL: TestSrlDiffBaseSkipsDevelopBetweenMainAndOriginMain (0.33s)
    env_literal_diff_base_test.go:139: srlDiffBase picked 44d344a56d583d51dbf55f7eaa6f8b7b8625514a; want merge-base with origin/main a00a1ffb774228befddf2a0a1953fe5d1a570fa8, not stale develop 44d344a56d583d51dbf55f7eaa6f8b7b8625514a (card t1856)
```
After each mutant `git checkout -- internal/hook/env_literal_diff_test.go`; `git diff --stat` then listed only
`internal/hook/env_literal_diff_base_test.go`. No mutant was committed. (Mutant B, returning HEAD
when main is absent, was not run; it fails `TestSrlDiffBaseFallsBackToOriginMain` by the same
mechanism as mutant A's second line.)

### Targeted GREEN (env scrubbed, TMPDIR/GOTMPDIR under /private/tmp/t-rp, real resolver)

```
--- PASS: TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration (2.01s)
--- PASS: TestSrlDiffBaseResolvesFromMainNotStaleDevelop (0.87s)
--- PASS: TestSrlDiffBaseFallsBackToOriginMain (0.27s)
--- PASS: TestSrlDiffBasePrefersLocalMainOverStaleOriginMain (0.43s)
--- PASS: TestSrlDiffBaseSkipsDevelopBetweenMainAndOriginMain (0.32s)
    env_literal_diff_test.go:52: env-literal sweep: 2 added lines swept across internal/hook internal/factorymsg internal/cli (envkeys.go and _test.go excluded), base=e4b747678aae096b22769fdfd3efc5d89498006a, distinct literals=0
--- PASS: TestNoNewEnvLiteralsInDiff (0.09s)
ok  	github.com/modu-ai/moai-adk/internal/hook	4.798s
```

### F3 literal-`develop` base lookups (follow-up material, not repaired here)

Commands, run on this tree: `git grep -n 'merge-base[^|]*develop' -- '*.go'` and
`git grep -n '"develop"' -- 'internal/**/*_test.go'` (filtered to merge/base/branch lines).
Own list from the first command (excluding this card's own `internal/hook/env_literal_diff_base_test.go` lines):
- internal/cli/codex_review_scope_test.go:163, :546 (code), :134 (comment)
- internal/cli/mcp_selfreview_fixture_test.go:128 (code), :64 (comment)
- internal/cli/mcp_selfreview_test.go:281
- internal/cli/factory_complete_test.go:409 (`merge-base --is-ancestor <sha> develop`; the fixture also does `merge ... develop` at :395)
- internal/cli/factory_merge_test.go:99
- internal/cli/mcp_build_identity_test.go:704 (comment)
- internal/cli/todo_issuance_probe_test.go:5 (comment)
- internal/factory/integration_remeasure_run_test.go:48 (`merge-base develop main`; NOT in the auditor's list)
Cross-check against the auditor's six: all six present (codex_review_scope_test.go:163,
mcp_selfreview_fixture_test.go:128, mcp_selfreview_test.go:281, factory_complete_test.go:409,
factory_merge_test.go:99, todo_issuance_probe_test.go:5). My list adds codex_review_scope_test.go:546,
integration_remeasure_run_test.go:48 and three comment lines. The second command's remaining hits
(IntegrationBranch: "develop", `--develop develop`, config fixtures) pass a branch name as data and are not base lookups;
that classification is my reading, not a run of each test.

### F4 recount of `origin/develop`

Commands: `git grep -n 'origin/develop' | wc -l` = 1923 on this tree (audit saw 1919; survey file has 1903 lines).
`git grep -n 'origin/develop' -- . ':!CHANGELOG.md' ':!.moai/reports/t1856' | wc -l` = 1888;
`-- CHANGELOG.md` = 18; `-- .moai/reports/t1856` = 17 (progress.md 11, sync-audit.md 4, verdict.md 2).
Diff of the current grep against the survey file (sorted, `comm`): new lines per file = progress.md 11,
sync-audit.md 4, verdict.md 2, CHANGELOG.md 18, internal/hook/env_literal_diff_base_test.go 2; removed =
CHANGELOG.md 17 (same entries, line numbers shifted, so they appear as removed plus re-added; net +1).
Net delta 11+4+2+1+2 = 20 = 1923 - 1903. Every line of the delta is the lane's own files; no other
file gained or lost an `origin/develop` hit. The survey file itself is untouched.

### Gaps

- The Linux run was NOT observed locally: CI is Linux-only and this agent ran on macOS. Removing the
  hard-coded path removes the known failure cause; it is not an observed Linux pass.
- The full `internal/hook` package and the landing re-measure have not run since this change.
- The repair changes only test code in `internal/hook`; no non-test file changed this round.
- The sync audit has not been re-run on this delta; the required codex backend still owes a verdict.
- Mutant B was not run (see above).

## 12. Sync audit round 2 FAIL and the second explicit wait (lane-22, 2026-10-17)

Round 2 (audited_sha `edbf85602`, file `.moai/reports/t1856/sync-audit.md`, preserved in commit
`5e8c07915`): FAIL 0.72; required backend codex returned fail on the adversarial run (receipt
`rcpt-470c5dc94cad67d7817db6a9`) and pass on a static native run (`rcpt-a36e96ddd71532fd458f2af6`).
Round 1 findings F1-F5 closed; the lane itself executed mutants C and D against the repaired tests
and observed each fail exactly the intended test. One blocking finding remains: N1, mutant
`{"main","origin/main","develop","origin/develop"}` survives all four tests (observed by the
auditor and by codex), so the card's "remove the develop path" is not pinned. Optional: N2
(`commit.gpgSign` in the caller's git config breaks the fixtures), N3 (self-referential recount),
N4 (record the scrub form that worked under the worktree guard).

Wait record `w-t1856-20261017T1935Z-c305b` (explicit wait, no idle-spin):
- reason: the repair spawn is expected to be refused again. The refusal record
  `.moai/state/audit-receipts/rejections/sync-auditor--unknown-spec.json` was recreated by the
  round 2 audit (agent_id ac3dece6c7d975bb5, rejected_at 2026-10-17T19:32:32Z, cause "verdict line
  missing"); the lane observed `starts/` empty and this the only file in `rejections/`.
- whom: the leader (standing rule `d-20261017T152443Z-c305`; the lane compared its five conditions
  and sent the facts at 19:35Z). The same record was handled for round 1 at 18:59Z.
- recheck point: every standing recheck cron fire; the lane renews both leases in each pass and does
  not delete the record, bypass the guard, or re-send the request.
- on resolution: spawn the N1 repair (a fixture with only develop-family refs expecting HEAD, a stale
  origin/develop versus origin/main case, `-c commit.gpgSign=false`, observe mutant F fail), re-run
  the targeted selector, then audit round 3 with the required codex backend.

## 13. Sync audit round 2 repair (lane-22)

Audit: `.moai/reports/t1856/sync-audit.md` round 2 (N1 blocking, N2-N4 optional). Repair tree HEAD
before this round: `530d5a803`; test-code commit: `3b4a99378`. The resolver `srlDiffBase` is unchanged
(ref list `{"main","origin/main"}`, last line `return "HEAD"`).

### Dispositions

- N1 repaired. Two tests added to `internal/hook/env_literal_diff_base_test.go`, same helper style and
  fixture email, each checking first that the expected answer differs from the wrong one:
  `TestSrlDiffBaseIgnoresDevelopFamilyWithoutMain` (no local main, no origin/main, stale `develop` and
  `refs/remotes/origin/develop`; the resolver must return the string `HEAD`; the fixture asserts both
  develop-family merge-bases are commits, not `HEAD`) and
  `TestSrlDiffBaseSkipsStaleOriginDevelopBeforeOriginMain` (no local main, `origin/develop` older than
  `origin/main`; base must be the origin/main merge-base).
- N2 repaired. Every helper git call carries `-c commit.gpgSign=false`, and runs with
  `GIT_CONFIG_GLOBAL=/dev/null` and `GIT_CONFIG_NOSYSTEM=1` appended to the inherited environment.
  The `-c user.email=env-diff-base-test@example.invalid` literal is untouched, so the AC-CIG-010 drift
  test still passes. The assertions are unchanged.
- N3 recounted below. N4 recorded below.
- CHANGELOG: the test count changed from four to six (recounted with
  `grep -c '^func TestSrlDiffBase' internal/hook/env_literal_diff_base_test.go` = 6), so the sentence
  was corrected and lists all six pins.

### Mutant observation (ref list of srlDiffBase edited, selector `-run 'TestSrlDiffBase'`, restored each time with `git checkout -- internal/hook/env_literal_diff_test.go`)

All four mutants die; none survives.

F `{"main","origin/main","develop","origin/develop"}` (dies):
```
--- FAIL: TestSrlDiffBaseIgnoresDevelopFamilyWithoutMain (0.18s)
    env_literal_diff_base_test.go:163: srlDiffBase picked 75fb09e73b219655b1b4d73f82d83ae8ae3dce0e; want the HEAD fallback, a develop-family ref must never resolve the base (card t1856)
```
G `{"main","origin/develop","origin/main"}` (dies):
```
--- FAIL: TestSrlDiffBaseIgnoresDevelopFamilyWithoutMain (0.13s)
    env_literal_diff_base_test.go:163: srlDiffBase picked ca6755f381709db4655d0e7aeaaed3c01142973a; want the HEAD fallback, a develop-family ref must never resolve the base (card t1856)
--- FAIL: TestSrlDiffBaseSkipsStaleOriginDevelopBeforeOriginMain (0.16s)
    env_literal_diff_base_test.go:185: srlDiffBase picked ca6755f381709db4655d0e7aeaaed3c01142973a; want merge-base with origin/main 84096f78a8541c4a67031e8ab32680146d8870f4, not stale origin/develop ca6755f381709db4655d0e7aeaaed3c01142973a (card t1856)
```
C `{"main","develop","origin/main"}` (dies, earlier kill holds):
```
--- FAIL: TestSrlDiffBaseSkipsDevelopBetweenMainAndOriginMain (0.16s)
    env_literal_diff_base_test.go:142: srlDiffBase picked 41a33f0c471c73deed30a08623cf8e81d6d9ba5e; want merge-base with origin/main 707d90ac70a7cae06e7412f9523eb55b116d2ab5, not stale develop 41a33f0c471c73deed30a08623cf8e81d6d9ba5e (card t1856)
--- FAIL: TestSrlDiffBaseIgnoresDevelopFamilyWithoutMain (0.14s)
    env_literal_diff_base_test.go:163: srlDiffBase picked 41a33f0c471c73deed30a08623cf8e81d6d9ba5e; want the HEAD fallback, a develop-family ref must never resolve the base (card t1856)
```
D `{"origin/main","main"}` (dies, earlier kill holds):
```
--- FAIL: TestSrlDiffBasePrefersLocalMainOverStaleOriginMain (0.23s)
    env_literal_diff_base_test.go:120: srlDiffBase picked 2af2806d94bdaa09896b4ef63a7d80b941f0afe2; want merge-base with local main bcec466ee5d8fee61536779f72199be037aa6511, not stale origin/main 2af2806d94bdaa09896b4ef63a7d80b941f0afe2 (card t1856)
```
After each mutant, `git diff --stat` listed only `internal/hook/env_literal_diff_base_test.go`. No mutant was committed.

### Targeted GREEN (real resolver, TMPDIR/GOTMPDIR under /private/tmp/t-rp)

```
--- PASS: TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration (0.66s)
--- PASS: TestSrlDiffBaseResolvesFromMainNotStaleDevelop (0.30s)
--- PASS: TestSrlDiffBaseFallsBackToOriginMain (0.12s)
--- PASS: TestSrlDiffBasePrefersLocalMainOverStaleOriginMain (0.15s)
--- PASS: TestSrlDiffBaseSkipsDevelopBetweenMainAndOriginMain (0.15s)
--- PASS: TestSrlDiffBaseIgnoresDevelopFamilyWithoutMain (0.12s)
--- PASS: TestSrlDiffBaseSkipsStaleOriginDevelopBeforeOriginMain (0.16s)
    env_literal_diff_test.go:52: env-literal sweep: 2 added lines swept across internal/hook internal/factorymsg internal/cli (envkeys.go and _test.go excluded), base=e4b747678aae096b22769fdfd3efc5d89498006a, distinct literals=0
--- PASS: TestNoNewEnvLiteralsInDiff (0.08s)
ok  	github.com/modu-ai/moai-adk/internal/hook	2.418s
```
The swept count is 2 (non-zero). `grep -n 'private/tmp' internal/hook/env_literal_diff_base_test.go` printed nothing.

### N3 recount of `origin/develop`, excluding sections 10-13 of this file

Section 11 counted 1923 before its own commit and the report files kept growing, so the figure moved.
The stable form excludes the lane's own report directory and the CHANGELOG:
`git grep -n 'origin/develop' -- . ':!.moai/reports/t1856' ':!CHANGELOG.md' | wc -l` = 1897 on this tree
(working tree with the new tests, before commit). The same exclusion measured 1888 in round 1. A sorted
`comm` of the two listings shows the whole +9 is `internal/hook/env_literal_diff_base_test.go`
(the new tests and comments name `origin/develop`); no other file gained or lost a hit. The unfiltered
count is 1943 here; the lane's own files hold the rest (progress.md 18, sync-audit.md 8, verdict.md 2
and CHANGELOG.md 18 in the same measurement, in tracked state), so unfiltered totals drift with every
record this lane appends and are not a baseline.

### N4 verification form that worked under the worktree guard

Plain separate commands, one per call, each starting with `cd <worktree absolute path> &&` only where
it was a single simple command. Test runs were one compound `unset CLAUDE_PROJECT_DIR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_CONFIG_GLOBAL MOAI_SESSION_ID && TMPDIR=/private/tmp/t-rp/tmp GOTMPDIR=/private/tmp/t-rp/gotmp go test ./internal/hook -run '...' -count=1`,
with the `|` pattern working inside a single `-run` selector, piped to `grep -E` for the result lines.
Mutant edits were made with the Edit tool and restored with a plain `git checkout -- <file>` call. What
was refused: a multi-line python heredoc that also ran git commands in the same call, so file edits
went through the Edit tool and git calls stayed separate. Scrubbed: the six variables in the unset
list. Not scrubbed: the rest of the inherited environment (PATH, HOME, MOAI_* others, OTEL_*), which
the tests do not read; the fixtures themselves now run with GIT_CONFIG_GLOBAL=/dev/null and
GIT_CONFIG_NOSYSTEM=1.

### Gaps

- The Linux run was NOT observed locally: CI is Linux-only and this agent ran on macOS.
- The full `internal/hook` package and the landing re-measure have not run since this change.
- Audit round 3 is pending (required codex backend).
- The repair changes only test code in `internal/hook`; no non-test file changed this round.
- The HEAD-fallback test asserts the literal string `HEAD` the resolver returns, so it pins that
  documented fallback, not a commit sha.

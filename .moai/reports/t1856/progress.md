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

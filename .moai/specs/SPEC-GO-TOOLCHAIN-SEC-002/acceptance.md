# Acceptance Criteria — SPEC-GO-TOOLCHAIN-SEC-002

Document-level tree pin: `d3b7d438d2c9bc041cb3b63ea41f9f1a03e867b1` (base commit of card
t610). It binds every RED-now cell below that carries no pin of its own. RED-now cells are
pre-implementation observations and stay pinned to this base. Green verdicts are measured at
M1–M3 (AC-GTS2-008 at M5) and re-measured on the absorbed tree at M4 (§ D.0b).

Exact baseline commands: `.moai/reports/t610/baseline/commands.md`.

## D. AC Matrix

| AC ID | Requirement | Class | RED-now (ledger) | Green path (milestone → passing output) |
|-------|-------------|-------|------------------|------------------------------------------|
| AC-GTS2-001 | REQ-GTS2-004 (gates AC-GTS2-002..007) | release-blocking | E-01 | M1, re-measured at M4 → `go -C <worktree-root> version` prints `go version go1.26.8 <os>/<arch>`; the same evidence file holds `go -C <worktree-root> env GOTOOLCHAIN GOMOD` showing `auto` (or empty) and `<worktree-root>/go.mod`; exit codes in `<file>.exit`, captured unpiped |
| AC-GTS2-002 | REQ-GTS2-001, REQ-GTS2-002 | release-blocking | E-02, E-03 | M1, re-measured at M4 → `sed -n 3p go.mod` prints `go 1.26.8`; `grep -c '^toolchain' go.mod` prints `0` |
| AC-GTS2-003 | REQ-GTS2-002 | release-blocking | E-04 | M1 after the run-phase commit, re-measured at M4 → `git diff --numstat develop...HEAD -- go.mod` prints exactly `1	1	go.mod`; `git diff develop...HEAD -- go.mod` shows `-go 1.26.4` / `+go 1.26.8` and no other hunk |
| AC-GTS2-004 | REQ-GTS2-005, REQ-GTS2-007 | release-blocking | E-05 | M2, re-measured at M4 → the evidence file holds `go -C <worktree-root> env GOTOOLCHAIN GOMOD` (`auto` or empty; worktree go.mod); govulncheck ./... with no `GOTOOLCHAIN` override exits `0`, exit code in `<file>.exit` captured unpiped; output contains `Your code is affected by 0 vulnerabilities`; each of the 8 IDs has 0 occurrences; control: output contains the `vulnerabilities in modules you require` line (presence only, since the count moves with the advisory DB) |
| AC-GTS2-005 | REQ-GTS2-006, REQ-GTS2-007 | release-blocking | E-06 | M3, re-measured at M4 → `make build` exit `0`; `go version -m bin/moai` first line `bin/moai: go1.26.8`; both outputs kept under `.moai/reports/t610/`, each exit code in `<file>.exit` captured unpiped |
| AC-GTS2-006 | REQ-GTS2-003 (non-breakage) | regression-guard | none (green today by design) | M3, re-measured at M4 → `go vet ./...` exit `0`; each selected package test run exit `0` with a non-empty swept set; the selection includes `internal/web` as a net/http server regression guard (it does not exercise #80927, plan.md § D2). Full suite, after M4 (not inside the lane at M3/M4): the CI run triggered by the lead's `origin/develop` push is green |
| AC-GTS2-007 | REQ-GTS2-003 | regression-guard | E-07 (green today by design) | M3 after the run-phase commit → the E-07 complement command prints nothing; M4 → the § D.0b complement command (which also admits the M5 sync paths) prints nothing |
| AC-GTS2-008 | REQ-GTS2-008 | sync-phase | E-08 | M5, re-measured at M4 → `grep -c '1\.26\.4'` on the 4 docs prints `0` for each; `grep -c '1\.26\.8'` totals 5 |

**Gating.** AC-GTS2-002..007 evidence is accepted only after AC-GTS2-001 passes on the same
tree. AC-GTS2-008 is a sync-phase criterion and is not gated. A "0 affecting" result captured
while the effective toolchain is still go1.26.4, or captured under a `GOTOOLCHAIN` override,
is a false PASS and is rejected.

**Why two ACs are regression guards.** AC-GTS2-006 and AC-GTS2-007 assert non-breakage and
scope discipline. They are green on the base tree by construction, so they cannot flip on this
work and are not release gates (verification-completeness.md §2, vacuous direction). They
still bind: a red on either blocks close.

## D.0a Why the diff ACs use the merge-base form

`git diff develop...HEAD` compares `HEAD` against `git merge-base develop HEAD`. It does not
compare against a fixed SHA or against the moving `develop` tip.

- **Before absorption,** the merge-base is the card base `d3b7d438d`, even though local
  `develop` has already moved on (E-04 context).
- **After M4 absorbs local `develop`,** the merge-base becomes the absorbed `develop` tip.
  The diff then holds only this card's side of the merge, so other lanes' changes that
  arrived through absorption cancel out. A later advance of `develop` does not move the
  merge-base.

Two conditions keep this true:

1. The ref named in the command is the ref that was absorbed. That is local `develop`, per
   CLAUDE.local.md:389.
2. The diff is taken on commits, after the run-phase commit, not on the working tree.

Under the 0.1.0 fixed-SHA form, a `require` bump that reached `develop` (for example a
dependabot group) would enter the `go.mod` numstat after absorption and turn AC-GTS2-003 red
for a reason unrelated to this work. The merge-base form does not see it.

Because the diff ACs read commits, files that `make build` regenerates in the working tree do
not enter AC-GTS2-007 unless someone commits them. M3 records
`git status --porcelain --untracked-files=no` before and after `make build`. Any tracked file
the build changed is reported in progress.md §E.2 and not committed (plan.md § F M3).

## D.0b Post-absorption re-measure (M4)

At M4, after `git merge develop` absorbs local `develop` into the card branch and before the
`--no-ff` merge into the integration worktree:

1. Re-run AC-GTS2-001..008 on the absorbed tree. That includes govulncheck, `make build`,
   `go vet ./...`, the selected package tests, and the AC-GTS2-008 doc counts, because
   absorbing `develop` can change the 4 project documents.
2. Record `git rev-parse HEAD` and `git rev-parse 'HEAD^{tree}'` in the same evidence file.
3. Judge the merge on this re-measure only, and always re-measure. M1–M3 evidence never
   stands as the post-merge verdict: the M5 sync commit lands after M3 and changes the tree,
   even when absorption brings nothing new.
4. The AC-GTS2-007 command at M4 also admits the M5 sync paths, because sync completes in the
   card worktree before the merge (CLAUDE.local.md:393):

```text
git diff --name-only develop...HEAD -- . ':!go.mod' ':!.moai/specs/SPEC-GO-TOOLCHAIN-SEC-002' ':!.moai/reports/t610' ':!CHANGELOG.md' ':!.moai/project/product.md' ':!.moai/project/structure.md' ':!.moai/project/codemaps/overview.md' ':!.moai/project/codemaps/modules.md'
```

5. Step 4 leaves the content of those 5 paths unjudged. AC-GTS2-008 counts version-token lines
   in the 4 project documents only, and no AC reads CHANGELOG.md. To keep an unintended change
   there visible, record both commands below in the M4 evidence file and confirm their outputs
   are identical. Under the scope guard (AC-GTS2-007 at M3) the M1–M3 commits touch none of
   these paths, so the card-side diff for them is the M5 sync commit alone. Any difference (an
   extra path, a changed count, or a merge-resolution change) is a blocker (EC-5).

```text
git diff --numstat develop...HEAD -- CHANGELOG.md .moai/project/product.md .moai/project/structure.md .moai/project/codemaps/overview.md .moai/project/codemaps/modules.md
git show --numstat --format= <M5-sync-commit> -- CHANGELOG.md .moai/project/product.md .moai/project/structure.md .moai/project/codemaps/overview.md .moai/project/codemaps/modules.md
```

## D.0 Evidence Ledger (RED-now, pinned to `d3b7d438d`)

Each entry records the command, verbatim stdout, exit code, and why it is red (or, for guards,
why it is green). Entries re-measured in revision 0.1.1 were run on the same base tree
(`HEAD` `d3b7d438d`, `git status --short` = `?? .moai/reports/t610/`,
`?? .moai/specs/SPEC-GO-TOOLCHAIN-SEC-002/`).

```text
E-01  AC-GTS2-001
  command : go -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t610 version
  stdout  : go version go1.26.4 darwin/arm64
  exit    : 0
  red     : effective toolchain is go1.26.4 (auto-acquired from go.mod:3), not the target go1.26.8
  env     : go env GOTOOLCHAIN GOMOD GOVERSION (cwd = worktree root)
            → auto
              /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t610/go.mod
              go1.26.4
            exit 0
  why -C  : a bare `go version` reads whichever go.mod governs the cwd; commands.md:24 records
            go1.26.0 from /private/tmp. `-C` pins the result to the worktree go.mod.
  also    : .moai/reports/t610/baseline/goversion-auto.txt (lane-8 capture, identical stdout)

E-02  AC-GTS2-002 (directive)
  command : sed -n 3p go.mod
  stdout  : go 1.26.4
  exit    : 0
  red     : directive names 1.26.4, target is 1.26.8

E-03  AC-GTS2-002 (no toolchain line; green half, guards the D3 form)
  command : grep -c '^toolchain' go.mod
  stdout  : 0
  exit    : 1 (corrected in 0.1.2; 0.1.1 recorded 0). A no-match grep exits 1: the plan-audit-2
            re-measure with `$?` saw exit=1 from both the shell grep and /usr/bin/grep. The tool
            this ledger was captured with does not display a non-zero exit for a no-match grep,
            so the exit value was not independently observable at capture time.
  judged  : on stdout, the matching-line count `0`, not on the exit code
  green   : no toolchain directive today; this half must stay 0 after the bump
  run     : run-phase exit codes are recorded only in the REQ-GTS2-007 file form
            (`<cmd> > <file> 2>&1; echo $? > <file>.exit`), never read off a tool display

E-04  AC-GTS2-003
  command : git diff --numstat develop...HEAD -- go.mod
  stdout  : (empty)
  exit    : 0
  red     : no directive change is committed yet; target stdout is `1	1	go.mod`
  context : git rev-parse --verify develop → d1b61005d20967fdbd970ec7ec734c6d14f29dc3 (exit 0)
            git merge-base develop HEAD    → d3b7d438d2c9bc041cb3b63ea41f9f1a03e867b1 (exit 0)
            Local develop is already past the card base, and the merge-base is still the base,
            so the form does not pick up develop's movement (§ D.0a).
  0.1.0   : the fixed-SHA form `git diff --numstat d3b7d438d -- go.mod` also printed empty, exit 0

E-05  AC-GTS2-004
  command : GOTOOLCHAIN=auto GOMAXPROCS=2 timeout 900 govulncheck -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t610 ./...
            (commands.md:8-16, run by lane-8; stdout+stderr carried to the log file below)
  stdout  : .moai/reports/t610/baseline/govulncheck-auto.log — "Your code is affected by 8 vulnerabilities from the Go standard library." IDs GO-2026-6218, 6091, 6090, 6089, 6088, 5972, 5856, 5026; lines 87-88 "This scan also found 2 vulnerabilities in packages you import and 3 / vulnerabilities in modules you require"
  exit    : 3   (.moai/reports/t610/baseline/govulncheck-exits.txt: "auto exit=3")
  red     : 8 affecting stdlib findings at go1.26.4
  control : same tree with GOTOOLCHAIN=go1.26.6 / go1.26.8 → exit 0, "Your code is affected by 0 vulnerabilities." and "0 vulnerabilities in packages you import and 3 vulnerabilities in modules you require" (govulncheck-go1.26.6.log, govulncheck-go1.26.8.log). The scanner reports when findings exist, the target clears them, and the "modules you require" line shows the module graph was loaded.
  env     : GOTOOLCHAIN=auto was set explicitly. `go env GOTOOLCHAIN` → auto, and the GOENV file carries no GOTOOLCHAIN (plan-audit Evidence, t610), so explicit auto and unset are the same on this machine.
  gap     : closed in 0.1.1 — the invocation is recorded in commands.md

E-06  AC-GTS2-005
  command : go version -m bin/moai
  stdout  : (empty; stderr: stat bin/moai: no such file or directory)
  exit    : 1
  red     : no binary built in this worktree yet (red by absence)
  version : the discriminating half (a go1.26.4 build prints `bin/moai: go1.26.4` and still fails
            this AC) is not observed on a binary built from this tree. Proxy observation,
            re-run in 0.1.1:
            command : go version /Users/goos/go/bin/moai
            stdout  : /Users/goos/go/bin/moai: go1.26.4
            exit    : 0
            A binary built under the current toolchain prints the discriminating text. It is the
            installed binary, not this worktree's build output.

E-07  AC-GTS2-007 (regression guard)
  command : git diff --name-only develop...HEAD -- . ':!go.mod' ':!.moai/specs/SPEC-GO-TOOLCHAIN-SEC-002' ':!.moai/reports/t610'
  stdout  : (empty)
  exit    : 0
  green   : nothing is committed on the card branch yet; must stay empty after the run-phase commit
  scope   : the complement covers every tracked path, including .goreleaser.yml and the Go files the
            0.1.0 pathspec (.github Makefile cmd internal pkg) missed:
            git ls-files '*.go' ':!cmd/**' ':!internal/**' ':!pkg/**' → 29 paths, exit 0, e.g.
            scripts/i18n-validator/main.go, scripts/convert-nextra-to-hextra/main.go,
            test/integration/harness/it01_replay_test.go, .moai/scripts/lint-skip-cleanup.go
  0.1.0   : `git diff --name-only d3b7d438d -- .github Makefile cmd internal pkg` also printed empty, exit 0

E-08  AC-GTS2-008
  command : grep -c '1\.26\.4' .moai/project/product.md .moai/project/structure.md .moai/project/codemaps/overview.md .moai/project/codemaps/modules.md
  stdout  : .moai/project/codemaps/overview.md:1
            .moai/project/product.md:2
            .moai/project/structure.md:1
            .moai/project/codemaps/modules.md:1
  exit    : 0
  red     : 5 mentions of the old version across 4 docs (product.md holds 2: lines 244, 300)
```

### Mutant probes (criterion depth)

- **AC-GTS2-004 mutant:** run govulncheck with `GOTOOLCHAIN=go1.26.8` on an unbumped go.mod.
  The scan passes while REQ-GTS2-005 is violated. It is killed by the no-override command form
  plus the AC-GTS2-001 gate plus AC-GTS2-002. The judged run records `go env GOTOOLCHAIN GOMOD`
  in the evidence file.
- **AC-GTS2-004 empty-scan mutant:** a scan that loaded no module graph and still prints
  "affected by 0". It is killed by the control that requires the "vulnerabilities in modules
  you require" line.
- **AC-GTS2-003 mutant:** add `toolchain go1.26.8` next to the unchanged directive. numstat
  shows `1	0`, not `1	1`, and AC-GTS2-002's toolchain count flips to 1. Killed.
- **AC-GTS2-005 mutant:** a stale `bin/moai` from a pre-bump build. The version text reads
  go1.26.4. Killed.
- **AC-GTS2-006 mutant:** a package selector that matches no tests prints `ok` and exits 0. It
  is killed by the non-empty swept-set requirement: a run whose output shows only
  `[no test files]` / `[no tests to run]` is not a pass.
- **AC-GTS2-007 mutant:** commit an edit to `scripts/i18n-validator/main.go` or
  `.goreleaser.yml`. The 0.1.0 pathspec stayed empty. The complement prints the path. Killed.

## D.1 Given-When-Then Scenarios

### Scenario 1 — Directive bump clears all 8 stdlib findings (AC-GTS2-001, -002, -004)

```
GIVEN  go.mod:3 declares `go 1.26.4` and govulncheck exits 3 with 8 affecting stdlib findings
WHEN   go.mod:3 is changed to `go 1.26.8`, and `go -C <worktree-root> version` reports go1.26.8
       (auto-acquired, no GOTOOLCHAIN override)
THEN   govulncheck ./... exits 0 and prints "Your code is affected by 0 vulnerabilities"
AND    none of GO-2026-6218, -6091, -6090, -6089, -6088, -5972, -5856, -5026 appears in its output
```

### Scenario 2 — Scope held and binary carries the target toolchain (AC-GTS2-003, -005, -007)

```
GIVEN  CI reads the Go version from go.mod via go-version-file and no other file pins Go
WHEN   the run phase completes M1–M3 and the go.mod change is committed
THEN   `git diff --numstat develop...HEAD -- go.mod` prints `1	1	go.mod`
AND    `git diff --name-only develop...HEAD -- . ':!go.mod' ':!.moai/specs/SPEC-GO-TOOLCHAIN-SEC-002' ':!.moai/reports/t610'` prints nothing
AND    `make build` exits 0 and `go version -m bin/moai` reports go1.26.8, kept as evidence
```

### Scenario 3 — Unsupported environment fails loudly (residual risk, D1)

```
GIVEN  an environment with GOTOOLCHAIN=local and an installed Go older than 1.26.8
WHEN   any go command runs in the repository after the bump
THEN   it fails with "go.mod requires go >= 1.26.8" rather than building silently on the old stdlib
```

Scenario 3 is documented behavior, not a run-phase gate. It records that the residual risk
fails visibly.

## D.2 Edge Cases

- **EC-1: effective toolchain not switched.** `GOTOOLCHAIN=local` or a cached older toolchain
  in the shell environment. AC-GTS2-001 fails and no downstream evidence is judged. Fix the
  environment (unset the override) and re-run. Do not edit go.mod further.
- **EC-2: new finding under go1.26.8.** An advisory published after the baseline. Return a
  blocker naming the ID and fix version (plan.md § D2 Edge). Do not silently pick another
  version.
- **EC-3: compile break or test regression under go1.26.8.** STOP and return a blocker. Source
  changes are out of scope.
- **EC-4: toolchain download fails (no proxy access).** This is an environment gap, not a SPEC
  failure. Record it in progress.md and escalate to the lead.
- **EC-5: absorption turns an AC red.** A re-measure on the absorbed tree (§ D.0b) fails while
  the pre-absorption evidence was green, or the § D.0b step 5 numstat comparison differs. The
  pre-absorption green is not the verdict. Record both tree ids. The M5 sync commit precedes M4,
  so `status` is already `completed` at this point: stop, do not merge into `develop`, and
  return a blocker to the lead. The fix lands as a new commit on the card branch, and M4 is
  re-run from § D.0b step 1. `status` is not rolled back silently; if the lead decides a
  correction is needed, it goes through manager-spec's `completed → in-progress (amendment)`
  path.

## D.3 Quality Gate Criteria

- The 5 release-blocking ACs (AC-GTS2-001..005) pass, with AC-GTS2-001 passing first.
- The 2 regression guards (AC-GTS2-006, AC-GTS2-007) are green. The selected package list, any
  `internal/cli` exclusion or approval, and each swept count are recorded in progress.md §E.2.
- AC-GTS2-001..008 are re-measured on the absorbed tree at M4, with `HEAD` and the tree id
  recorded, and the 5 sync paths' card-side numstat matches the M5 sync commit (§ D.0b).
- Every judged output is a file under `.moai/reports/t610/`, with an unpiped exit code.
- The full-suite verdict is the CI result on the `origin/develop` push that carries this merge.
  A local full-suite run is not evidence and is not performed.

## D.4 Definition of Done

- [ ] AC-GTS2-001: `go -C <worktree-root> version` reports go1.26.8, with `go env GOTOOLCHAIN GOMOD` recorded (verified first)
- [ ] AC-GTS2-002: `go.mod:3` = `go 1.26.8`, no `toolchain` line
- [ ] AC-GTS2-003: `git diff --numstat develop...HEAD -- go.mod` = `1	1	go.mod`, directive line only
- [ ] AC-GTS2-004: govulncheck exit 0, 0 affecting, 8 IDs absent, "modules you require" line present (no override)
- [ ] AC-GTS2-005: `make build` exit 0, `go version -m bin/moai` = go1.26.8, output kept
- [ ] AC-GTS2-006: `go vet ./...` exit 0, selected package tests green with non-empty swept set
- [ ] AC-GTS2-007: no committed tracked-file change outside `go.mod`, the SPEC directory, and `.moai/reports/t610/`
- [ ] AC-GTS2-008 (sync): 5 doc mentions → 1.26.8
- [ ] M4: AC-GTS2-001..008 re-measured on the absorbed tree, `HEAD` and tree id recorded; the 5 sync paths' numstat matches the M5 sync commit
- [ ] Merged alone in a dedicated integration window (D4); CI on `origin/develop` green
- [ ] CHANGELOG security entry and status transitions (draft → in-progress → implemented → completed)

## D.5 Closure Gates (3-phase)

- **Plan:** 4 artifacts created, status `draft` (manager-spec).
- **Run:** M1–M3 done, AC-GTS2-001..007 judged, status → `in-progress` on the first run commit
  (manager-develop).
- **Sync:** M5 done in the card worktree, AC-GTS2-008 passed, CHANGELOG entry, status →
  `implemented` → `completed` on the single sync commit (manager-docs).
- **Integration:** M4 after the sync commit. Absorb local `develop`, re-measure
  AC-GTS2-001..008 on the absorbed tree and compare the 5 sync paths with the M5 sync commit
  (§ D.0b), merge alone (D4), and read the CI verdict after the lead's push.

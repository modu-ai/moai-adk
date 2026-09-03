# acceptance.md — SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001

## §D.0 RED-now baseline (plan-phase measurement, tree `d592b0551`)

Command (plan-phase, worktree t467; temporary probe
`internal/hook/zz_tmp_bgprobe_test.go` driving `matchBranchStateCommand`,
created via Write, executed once, deleted after):

```
go test ./internal/hook/ -run TestTmpBranchFlagProbe -v
```

Verbatim output (exit code 0 — the probe logs, it does not assert):

```
PROBE	"git branch -f topic abc1234"	matched=false
PROBE	"git branch --force topic abc1234"	matched=false
PROBE	"git branch -df old"	matched=false
PROBE	"git branch -fm renamed"	matched=false
PROBE	"git branch -vD old"	matched=false
PROBE	"git branch -u origin/main topic"	matched=false
PROBE	"git branch --set-upstream-to=origin/main topic"	matched=false
PROBE	"git branch --unset-upstream topic"	matched=false
PROBE	"git branch -t topic origin/main"	matched=false
PROBE	"git branch --edit-description topic"	matched=false
PROBE	"git branch --merged main"	matched=false
PROBE	"git branch --no-merged main"	matched=false
PROBE	"git branch --points-at HEAD"	matched=false
PROBE	"git branch --format %(refname)"	matched=false
PROBE	"git branch --list develop -v"	matched=false
PROBE	"git branch -d old"	matched=true
PROBE	"git branch feature"	matched=true
--- PASS: TestTmpBranchFlagProbe (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.731s
```

Reading: the 10 mutation-form cells are the defect (matched=false =
allowed through); the query-form cells and the two controls confirm both
the clean over-match axis and the working deny axis. Existing pinned tests
additionally prove (green-now, cited from
`internal/hook/branch_guard_test.go`): deny for `-d/-D/-m/-M/-c/-C`,
bare-name, `newbranch oldstart` (lines 234-245); allow for `-v`, `-a`,
`--list`, `--list develop -v`, `-vv`, `-r`, bare, `--show-current`,
`--contains HEAD` (lines 308-320).

Note on the undecidable disposition: the probe command above is not
re-executable as-is (the probe file was deleted after measurement); the
M1 permanent matrix test is the re-executable form. The classification
RED-now evidence for the M1 cells is therefore the M1 run itself on the
pre-fix tree (M1 milestone obligation), with this §D.0 output as the
plan-phase pin of the same fact.

## §D.1 Expected measurement matrix (target, post-M2)

Each cell: `deny` = guard denies in the primary checkout with
`BRANCH_GUARD_VIOLATION:` prefix; `allow` = guard allows. "Pre-fix" column
is the measured current state (from §D.0 and the existing pinned tests).

### Mutation forms → deny

| # | Command form | Pre-fix (measured) | Post-fix expected |
|---|---|---|---|
| M-01 | `git branch feature` (bare creation) | deny (pinned test) | deny |
| M-02 | `git branch newbranch oldstart` | deny (pinned test) | deny |
| M-03 | `git branch -d old` / `-D old` | deny (pinned test) | deny |
| M-04 | `git branch -m renamed` / `-M renamed` | deny (pinned test) | deny |
| M-05 | `git branch -c copied` / `-C copied` | deny (pinned test) | deny |
| M-06 | `git branch -f topic abc1234` | allow (§D.0) | **deny** |
| M-07 | `git branch -f topic` (create at HEAD) | allow (same token shape as M-06) | **deny** |
| M-08 | `git branch --force topic abc1234` | allow (§D.0) | **deny** |
| M-09 | `git branch -df old` | allow (§D.0) | **deny** |
| M-10 | `git branch -fm renamed` | allow (§D.0) | **deny** |
| M-11 | `git branch -vD old` | allow (§D.0) | **deny** |
| M-12 | `git branch -u origin/main topic` | allow (§D.0) | **deny** |
| M-13 | `git branch --set-upstream-to=origin/main topic` | allow (§D.0) | **deny** |
| M-14 | `git branch --set-upstream origin/main topic` | allow (token shape; not probed — M1 adds the cell) | **deny** |
| M-15 | `git branch --unset-upstream topic` | allow (§D.0) | **deny** |
| M-16 | `git branch -t topic origin/main` | allow (§D.0) | **deny** |
| M-17 | `git branch --track topic origin/main` | allow (token shape; M1 adds) | **deny** |
| M-18 | `git branch --no-track topic origin/main` | allow (token shape; M1 adds) | **deny** |
| M-19 | `git branch --edit-description topic` | allow (§D.0) | **deny** |

Rationale for M-12..M-19 (config-mutation forms): they mutate branch state
(upstream tracking, description, creation modifiers) without being
ref-pointer rewrites; the doctrine's forbidden-table rationale ("leaves a
branch other sessions did not expect" / shared-state mutation) covers them.
Their classification is asserted by the matrix (measured at M1), never by
this rationale alone.

### Query forms → allow

| # | Command form | Pre-fix (measured) | Post-fix expected |
|---|---|---|---|
| Q-01 | `git branch` (bare) | allow (pinned test) | allow |
| Q-02 | `git branch --list` / `--list develop -v` | allow (pinned + §D.0) | allow |
| Q-03 | `git branch -v` / `-vv` | allow (pinned test) | allow |
| Q-04 | `git branch -a` / `-r` | allow (pinned test) | allow |
| Q-05 | `git branch --show-current` | allow (pinned test) | allow |
| Q-06 | `git branch --contains HEAD` | allow (pinned test) | allow |
| Q-07 | `git branch --merged main` | allow (§D.0) | allow |
| Q-08 | `git branch --no-merged main` | allow (§D.0) | allow |
| Q-09 | `git branch --points-at HEAD` | allow (§D.0) | allow |
| Q-10 | `git branch --format %(refname)` | allow (§D.0) | allow |
| Q-11 | `git branch --sort=-committerdate` | allow (token shape; M1 adds) | allow |
| Q-12 | `git branch -q` / `-i` | allow (token shape; M1 adds) | allow |

### Whole-token discrimination pairs (REQ-WBG-F-004)

| # | Pair | Expected |
|---|---|---|
| P-01 | `--format %(refname)` vs `--force topic abc1234` | allow vs deny — full-token classification, no prefix match |
| P-02 | `--contains HEAD` / `--merged main` / `--no-merged main` despite embedded `c`/`m` | all allow |

## §D.2 Given-When-Then acceptance criteria

- **AC-WBG-F-001 (matrix convergence)** — Given the M1 matrix test from
  §D.1 exists with doctrine-based expectations, When it runs on the post-M2
  tree, Then every cell matches its expected classification (19/19
  mutation→deny, 12/12 query→allow). RED-now: §D.0 (10 cells fail on
  `d592b0551`). Green path: M2.
- **AC-WBG-F-002 (combined short-flag clusters)** — Given a `git branch`
  command with a combined short-flag cluster containing `d/D/m/M/c/C/f`
  (`-df`, `-fm`, `-vD`), When evaluated in the primary checkout, Then the
  guard denies with the `BRANCH_GUARD_VIOLATION:` prefix. RED-now: §D.0.
  Green path: M2.
- **AC-WBG-F-003 (query allowlist regression)** — Given any §D.1 query
  form, When evaluated in the primary checkout pre- and post-fix, Then the
  guard allows it in both (no over-match regression; the CLAUDE.local.md
  §4.1 friction must NOT recur). Green-now: pinned tests + §D.0; stays
  green through M2 (any flip is a FAIL, not a trade).
- **AC-WBG-F-004 (whole-token classification)** — Given the P-01/P-02
  pairs, When evaluated, Then `--format` allows while `--force` denies, and
  `--contains`/`--merged`/`--no-merged` allow. RED-now: `--force` cell in
  §D.0. Green path: M2.
- **AC-WBG-F-005 (preserved surfaces)** — Given the existing test files
  (`branch_guard_test.go`, `branch_guard_quoted_test.go`,
  `branch_guard_worktree_test.go`, `branch_guard_pr1338_test.go`),
  When the full `./internal/hook/` package test run executes post-fix,
  Then every pre-existing test passes unmodified — the deny axis pins
  (lines 234-245), sentinel, discriminant, and quoted-span behavior are
  unchanged. Green-now; stays green (characterization).
- **AC-WBG-F-006 (opt-in gate)** — Given `Workflow.BranchGuard.Enabled` is
  false (distributed default), When a mutating `git branch -f` command is
  evaluated, Then no pattern evaluation occurs and no deny is emitted; the
  existing `pre_tool_branch_guard_optin_test.go` suite (which enables the
  flag explicitly in every gated test) stays green. Documented condition:
  tests enable the flag explicitly — the default-off state is why the
  defect is latent for distributed users and live for this repo's dogfood
  config. Green-now; stays green.
- **AC-WBG-F-007 (fail-open preserved)** — Given a non-git cwd, a missing
  git binary, or a failing `git rev-parse`, When the guard evaluates any
  command, Then it allows and appends the advisory entry to
  `.moai/logs/branch-guard-audit.log` (existing fail-open tests stay
  green; the fix adds no new blocking path). Green-now; stays green.
- **AC-WBG-F-008 (exemption-axes conditions documented)** — Given a
  tool-spawned subagent issues a mutating `git branch -f` in the primary
  checkout, When it attempts to bypass via `AgentType` or by exporting
  `MOAI_BRANCH_GUARD_EXEMPT=1` inside the guarded command, Then neither
  axis fires (documented condition, encoded in M3 test conditions:
  `AgentType` is populated only for main-thread `claude --agent <name>`
  launches; the sentinel is read from the hook process's own environment,
  spawned before the guarded command runs) and the deny stands. This AC
  documents existing behavior (per -OPTIN-001 REQ-6 +
  `main-checkout-branch-guard-detail.md` exemption reachability caveat); it
  requires test-condition documentation, NOT code change. Green-now
  (`pre_tool_branch_guard_optin_test.go:140-175` proves both axes fire when
  the values ARE delivered); M3 adds the unreachable-from-subagent
  condition documentation.

## §D.3 Edge cases

- **E-1 mixed flags**: a command presenting BOTH a mutating flag and query
  flags (`git branch -d old -v`) → deny (any mutating flag present
  dominates; conservative direction).
- **E-2 `--set-upstream-to=<value>` attached form**: the `=`-attached
  value must not be mistaken for a bare creation operand.
- **E-3 case-insensitivity**: patterns compile `(?i)`; `git branch -F` /
  `--FORCE` classify as their lowercase forms.
- **E-4 quoted spans**: quoted branch names (`git branch -D 'old-feature'`)
  still deny — quoted-span collapse preserves the operand (per
  `branch_guard_quoted_test.go`).
- **E-5 unknown/exotic flags**: a flag in neither set classifies as query
  → allow (fail-open direction; under-match is the documented correct
  direction for unclassifiable forms).
- **E-6 compound commands**: `git branch -f x y && git status` → the
  `git branch` segment still matches (existing scan semantics; no change).

## §D.4 Quality gates / Definition of Done

- `go test ./internal/hook/ -count=1` green (pre-existing + M1 matrix + M3
  cases); `golangci-lint run` shows no NEW findings vs the measured
  baseline.
- Coverage of the changed matcher code exercised by the matrix test ≥ 85%
  (package-level target).
- All ACs reported PASS with the attribution triple (command + verbatim
  output + HEAD SHA) in progress.md §E.2; M1's RED output preserved
  verbatim (E8-class TDD evidence).
- No file outside `internal/hook/` + this SPEC directory touched.

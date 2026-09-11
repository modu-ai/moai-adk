---
id: SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001
title: "Acceptance criteria — integration lock target provenance and card visibility (card t637)"
version: "0.1.0"
created: 2026-09-11
updated: 2026-09-11
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/cli, internal/kanban, internal/config"
lifecycle: spec-anchored
tags: "integration-lock, kanban, factory, git-flow, observability, t637"
tier: M
---

# Acceptance Criteria — SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001

Every criterion is decided by the observable output of a NAMED command; none may be satisfied by
reading code. Go test names below are the names run-phase is expected to create; where run-phase
chooses a different name, `progress.md` §E.2 records the mapping and the deciding command changes
with it. Commands are scoped to the touched packages; `go test ./...` appears nowhere and is not
run locally — CI is the full-suite judge.

## §A Shared conventions

- **Go tests** use `t.TempDir()` scratch repositories (the existing `scratchRepo` /
  `writeGitStrategyFixture` helpers in `internal/cli/integration_target_test.go`) and pin
  `CLAUDE_PROJECT_DIR` to the scratch root, so no test reads or writes this repository's
  `.moai/state` or configuration.
- **Stream separation.** Criteria that assert where the warning is written use a helper that
  captures standard output and standard error in SEPARATE buffers. The existing `runIntegration`
  helper merges both streams and therefore cannot decide them.
- **Fixture cells** use the fixture left by the verdict: root `/tmp/t637-fx` (branch `main`),
  worktrees `/tmp/t637-fx-wt/develop` (`develop`), `/tmp/t637-fx-wt/cardA` (`WT-a`),
  `/tmp/t637-fx-wt/cardB` (`WT-b`); configuration
  `/tmp/t637-fx/.moai/config/sections/git-strategy.yaml` (`mode: "manual"`, `workflow: git-flow`,
  `develop_branch: develop` at plan time). `$BIN` is a binary built **from this tree** in
  run-phase (never the installed `moai`, which predates the change). Every fixture invocation
  sets `CLAUDE_PROJECT_DIR=/tmp/t637-fx` **in the same invocation**, and every cell ends with
  `CLAUDE_PROJECT_DIR=/tmp/t637-fx "$BIN" integration release --session fx-lane4`.
- **Real-window guard.** Before and after each fixture cell run
  `shasum /Users/goos/MoAI/moai-adk-go/.moai/state/integration-lock.json 2>&1` — the two outputs
  must be identical (a matching "No such file" on both sides also counts as identical).
- **Evidence** is persisted under `.moai/reports/t637/` (or `.moai/state/verify/<session>/`) and
  cited verbatim in `progress.md` §E.2.
- **Config toggle for fixture cells.** Empty the develop branch with
  `sed -i '' 's/^\( *develop_branch:\).*/\1 ""/' /tmp/t637-fx/.moai/config/sections/git-strategy.yaml`
  and restore it with
  `sed -i '' 's/^\( *develop_branch:\).*/\1 develop/' /tmp/t637-fx/.moai/config/sections/git-strategy.yaml`;
  confirm each toggle with `grep -n develop_branch /tmp/t637-fx/.moai/config/sections/git-strategy.yaml`.

## §D AC Matrix

| AC | Requirement(s) | Claim | Deciding command | Observable that decides it |
|---|---|---|---|---|
| AC-ILT-001 | REQ-ILT-001 | A record carrying a card prints a `card:` line in text status | `go test ./internal/cli/... -run TestIntegrationStatus_ShowsCardLine -count=1 -v` | PASS; the test asserts the status text contains the line `  card:     t-fixture` for a record seeded with `Card: "t-fixture"` |
| AC-ILT-002 | REQ-ILT-002 | A card-less record prints no `card:` line | `go test ./internal/cli/... -run 'TestIntegrationStatus_NoCardPrintsNoCardLine\|TestIntegrationStatus_NoNameKeepsTodaysShape' -count=1 -v` | PASS; the new test asserts the substring `card:` is ABSENT; the pre-existing shape test passes unedited |
| AC-ILT-003 | REQ-ILT-003, 004, 007 | Flag source — an explicit `--branch` records `branch_source: flag` and emits no warning, even in a git-flow project whose develop branch is empty | `go test ./internal/cli/... -run TestIntegrationAcquire_RecordsBranchSource/flag -count=1 -v` | PASS; the test asserts the on-disk record's `branch_source == "flag"`, the recorded branch equals the flag value, and the captured standard error contains no `[moai:integration-lock] warning:` |
| AC-ILT-004 | REQ-ILT-003, 004, 007 | Config source — git-flow with a configured develop branch records `branch_source: config`, no warning | (a) `go test ./internal/cli/... -run TestIntegrationAcquire_RecordsBranchSource/config -count=1 -v`; (b) fixture cell C3′ (§D.2) | (a) PASS with `branch_source == "config"`, branch `fixture-integration`, standard error free of the warning prefix. (b) status text line `  branch:   develop (source: config)`, `c3.stderr` empty, `status --json` lock object carries `"branch_source":"config"` |
| AC-ILT-005 | REQ-ILT-003, 004, 006 | **Control 1 — same tree, own branch**: with an empty develop branch the record matches the merge target, is marked `caller`, and exactly one warning is emitted | (a) `go test ./internal/cli/... -run TestIntegrationAcquire_GitFlowEmptyDevelopWarnsOnCallerFallback -count=1 -v`; (b) fixture cell C1′ (§D.2) | (a) PASS; standard error holds exactly one line, beginning `[moai:integration-lock] warning:` and containing the caller branch and both remedy tokens `develop_branch` and `--branch`; record `branch_source == "caller"`. (b) exit 0; `grep -c '' c1.stderr` prints `1`; the line names `WT-a`; status text shows `  card:     tA` and `  branch:   WT-a (source: caller)` and `  worktree: /tmp/t637-fx-wt/cardA` (or its `/private` form) — the record matches the tree being merged; real-window guard identical |
| AC-ILT-006 | REQ-ILT-001, 003, 004, 006 | **Control 2 — other tree, other branch**: a lane in tree A merging card B produces a record whose mismatch is now visible from text status alone | (a) `go test ./internal/cli/... -run TestIntegrationAcquire_CallerFallbackMismatchIsVisible -count=1 -v`; (b) fixture cell C2′ (§D.2) | (a) PASS; from a caller tree on branch `main` with `--card t-other`, status text carries both `card:     t-other` and `(source: caller)` on the branch line naming `main`, and one warning line is emitted. (b) exit 0; one warning line naming `WT-a`; status text shows `  card:     tB` together with `  branch:   WT-a (source: caller)` — card B against branch A, with the provenance saying the branch came from the caller's tree; real-window guard identical |
| AC-ILT-007 | REQ-ILT-007 | **No-warn negative** — a caller fallback in a project that is not git-flow, or has no git strategy file, emits no warning | `go test ./internal/cli/... -run 'TestIntegrationAcquire_NonGitFlowCallerFallbackDoesNotWarn\|TestIntegrationAcquire_NoConfigCallerFallbackDoesNotWarn\|TestIntegrationAcquire_NonGitFlowFallsBackToCallerTree' -count=1 -v` | PASS; with `workflow: github-flow` (develop branch present) and with no git strategy file, standard error contains no `[moai:integration-lock] warning:`, `branch_source == "caller"`, and the pre-existing t449 test passes unedited |
| AC-ILT-008 | REQ-ILT-008 | The warning neither refuses nor leaks onto standard output | `go test ./internal/cli/... -run TestIntegrationAcquire_WarningIsOnStderrOnly -count=1 -v` | PASS; in the warning scenario the command returns a nil error, the record is written, text-mode standard output equals exactly the `release-integration window acquired by <session> on <branch>` line, and `--json` standard output decodes as ONE JSON object with `"acquired": true` while the warning appears only in the standard-error buffer |
| AC-ILT-009 | REQ-ILT-005 | **Old-record compatibility** — a record without `branch_source` reads cleanly and is not rewritten | `go test ./internal/cli/... -run TestIntegrationStatus_OldRecordKeepsTodaysBranchLine -count=1 -v` | PASS; the test writes a record JSON with NO `branch_source` key directly to the lock file, runs `status` and `status --json`, and asserts: nil error; the branch line equals exactly `  branch:   <branch>` with no `(source:` suffix; the JSON lock object has no `branch_source` key; the lock file bytes are identical before and after |
| AC-ILT-010 | REQ-ILT-009 | The `--branch` help names the integration TARGET | (a) `go test ./internal/cli/... -run TestIntegrationAcquire_BranchFlagHelpNamesTheTarget -count=1 -v`; (b) `"$BIN" integration acquire --help 2>&1 \| grep -- '--branch'` | (a) PASS; the flag's usage contains `integration target` and does NOT contain `Branch being integrated`. (b) prints one line containing `integration target` and the default-resolution wording (`git-flow develop branch`) |
| AC-ILT-011 | REQ-ILT-010 | Every documented lane acquire carries `--card <card-id>`; the release invocation states it has no card | `git grep -n "moai integration acquire" -- .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md .claude/rules/local/gitflow-lane-protocol.md .claude/agents/harness/hns-release-specialist.md CLAUDE.local.md` | Exactly 8 lines print (base measurement on `1ad0fdc09`: the same 8 sites, 0 carrying `--card`); every line except the `hns-release-specialist.md` one contains `--card <card-id>`; the `hns-release-specialist.md` line contains `no card` and no `--card <` |
| AC-ILT-012 | REQ-ILT-011 | The template mirror stays byte-identical and neutral, and the embed is rebuilt | `diff -q .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`; `git diff 1ad0fdc09 -- internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md \| grep '^+[^+]' \| grep -cE 't[0-9]{3}\|SPEC-\|20[0-9]{2}-[0-9]{2}-[0-9]{2}\|[0-9a-f]{9}'` (no `\b` — it is not a word boundary in POSIX ERE); `make build` | `diff -q` prints nothing (exit 0); the count prints `0` (base measurement: the same pattern over today's line 230 prints `0`), AND the added-line count from `git diff 1ad0fdc09 -- <template path> \| grep -c '^+[^+]'` is `1` so the `0` is not an empty-input zero; `make build` exits 0 |
| AC-ILT-013 | REQ-ILT-012, 013 | Invariants — resolution order, recorded values, and guard behavior unchanged | `go test ./internal/cli/... -run 'TestResolveIntegrationTarget\|TestWorktreeForBranch\|TestIntegrationAcquire_RecordsTheIntegrationWorktreeNotTheCaller\|TestIntegrationAcquire_ExplicitBranchResolvesItsWorktree\|TestIntegrationAcquire_NoMatchingWorktreeRecordsEmptyWorktree\|TestIntegration_AcquireStatusRelease' -count=1`; `go test ./internal/hook/... -run IntegrationLock -count=1`; `git diff --stat 1ad0fdc09 -- internal/hook/integration_lock_guard.go` | Both test runs PASS; the pre-existing test functions' assertions are unchanged (any edit to them is additive and listed in §E.2); the guard diff prints nothing |
| AC-ILT-014 | REQ-ILT-001, 003, 006, 007 (mutation guard) | The new tests are not vacuous | For each row of §D.3: apply the one-line mutation, run its named test, restore, run again | Each mutated run FAILS its named test; each restored run PASSES; all outputs recorded verbatim in `progress.md` §E.2 |

## §D.1 Severity

MUST-PASS: AC-ILT-001, 003, 004, 005, 006, 007, 008, 009, 013, 014.

SHOULD-PASS (failure is reported and judged, never silently accepted): AC-ILT-002, 010, 011, 012.

CI-JUDGED: the full suite, including platform matrices, on the pushed `develop` head — not
claimable from a local run.

## §D.2 Given-When-Then scenarios

**AC-ILT-005 — Control 1, same tree / own branch (fixture cell C1′)**

- **Given** the fixture with `develop_branch` emptied (toggle in §A), the real-window guard
  captured, and `$BIN` built from this tree
- **When** `cd /tmp/t637-fx-wt/cardA && CLAUDE_PROJECT_DIR=/tmp/t637-fx "$BIN" integration acquire --session fx-lane4 --name lane-4 --card tA 2>"$EV/c1.stderr"; echo "rc=$?"`
  runs, followed by `CLAUDE_PROJECT_DIR=/tmp/t637-fx "$BIN" integration status` and
  `CLAUDE_PROJECT_DIR=/tmp/t637-fx "$BIN" integration status --json`
- **Then** `rc=0`; `c1.stderr` holds exactly one line beginning `[moai:integration-lock] warning:`
  and naming `WT-a`; the status text shows `card:     tA`, `branch:   WT-a (source: caller)` and
  the cardA worktree; the JSON lock object carries `"branch_source":"caller"` and `"card":"tA"`;
  the record matches the branch actually being merged; the real-window guard is identical

**AC-ILT-006 — Control 2, other tree / other branch (fixture cell C2′)**

- **Given** the same emptied fixture, and a lane standing in `/tmp/t637-fx-wt/cardA` whose intent
  is to merge `WT-b` (card `tB`)
- **When** `cd /tmp/t637-fx-wt/cardA && CLAUDE_PROJECT_DIR=/tmp/t637-fx "$BIN" integration acquire --session fx-lane4 --name lane-4 --card tB 2>"$EV/c2.stderr"; echo "rc=$?"`
  runs, followed by `status` as in C1′
- **Then** `rc=0` (not refused); exactly one warning line naming `WT-a`; the status text shows
  `card:     tB` next to `branch:   WT-a (source: caller)` — the verdict's silent C2 mismatch is
  now readable from text status alone; the real-window guard is identical

**AC-ILT-004(b) — configured target (fixture cell C3′)**

- **Given** the fixture with `develop_branch: develop` restored
- **When** `cd /tmp/t637-fx-wt/cardA && CLAUDE_PROJECT_DIR=/tmp/t637-fx "$BIN" integration acquire --session fx-lane4 --name lane-4 --card tB 2>"$EV/c3.stderr"; echo "rc=$?"`
  runs, followed by `status` and `status --json`
- **Then** `rc=0`; `c3.stderr` is empty; the status text shows `card:     tB` and
  `branch:   develop (source: config)` with the develop worktree; JSON carries
  `"branch_source":"config"`

**AC-ILT-007 — no warning where the fallback is legitimate**

- **Given** a scratch repository whose git strategy is `github-flow` (a develop branch value
  present), and separately one with no git strategy file
- **When** `acquire` runs from the repository root with no `--branch`
- **Then** the record's branch is the caller's `main`, `branch_source == "caller"`, and the
  standard-error buffer contains no warning line

**AC-ILT-009 — old record**

- **Given** a lock file whose JSON has no `branch_source` key (the shape every record written
  before this SPEC carries)
- **When** `status` and `status --json` run
- **Then** both succeed, the text branch line is today's exact line, the JSON lock object has no
  `branch_source` key, and the lock file's bytes are unchanged

## §D.3 Mutation guard (AC-ILT-014)

| Row | One-line mutation (production code) | Test that MUST fail |
|---|---|---|
| a | The REQ-ILT-006 warning condition never holds (warning disabled) | `TestIntegrationAcquire_GitFlowEmptyDevelopWarnsOnCallerFallback` |
| b | The warning condition drops the git-flow predicate (warns on every caller fallback) | `TestIntegrationAcquire_NonGitFlowCallerFallbackDoesNotWarn` |
| c | The status printer's `card:` line is removed | `TestIntegrationStatus_ShowsCardLine` |
| d | `branch_source` is hard-coded to `caller` | `TestIntegrationAcquire_RecordsBranchSource/flag` and `/config` |
| e | The warning is written to standard output instead of standard error | `TestIntegrationAcquire_WarningIsOnStderrOnly` |
| f | The provenance suffix is printed even when `branch_source` is empty | `TestIntegrationStatus_OldRecordKeepsTodaysBranchLine` |

A mutation that leaves its named test green is a vacuous test and blocks the card until the test
is repaired; it is not "passed by construction".

## §D.4 Edge cases

- `--branch "   "` (blank) behaves as no flag: the configured tier or the caller decides, and
  `branch_source` names whichever did (never `flag`).
- A git-flow project with `develop_branch` blank AND a non-blank `--branch`: source `flag`, no
  warning (the operator chose the target explicitly).
- A personal/team mode profile carrying `workflow: git-flow`: treated as not git-flow (same
  predicate as the configured tier) — no warning.
- A stale record (holder gone) prints the card line and the provenance suffix like a live one.
- A record with `card` but no `branch_source` prints the card line and today's branch line.

## §D.5 Quality gates

- `go vet ./internal/cli/... ./internal/kanban/... ./internal/config/...` exits 0.
- `golangci-lint run ./internal/cli/... ./internal/kanban/... ./internal/config/...` reports no
  new findings.
- `gofmt -l internal/cli internal/kanban internal/config` prints nothing.
- No `go test ./...` locally; the pushed `develop` head's CI is the full-suite verdict.

## §D.6 Definition of Done

- All MUST-PASS criteria observed PASS with verbatim evidence in `progress.md` §E.2; SHOULD-PASS
  results recorded (PASS, or FAIL with judgement).
- Mutation rows a-f each observed failing then passing.
- Fixture cells C1′, C2′, C3′ run with `$BIN` from this tree, real-window guard identical, fixture
  configuration restored to `develop_branch: develop` afterwards.
- Template mirror identical to the local copy; `make build` exit 0.
- The CLAUDE.local.md divergence risk (plan.md §E3) named in the completion report.

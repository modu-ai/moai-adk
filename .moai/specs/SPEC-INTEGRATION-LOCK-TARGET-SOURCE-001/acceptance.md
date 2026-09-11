---
id: SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001
title: "Acceptance criteria — integration lock target provenance and card visibility (card t637)"
version: "0.3.0"
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
with it. Commands are scoped to the touched packages; `go test ./...` is not run locally — CI is
the full-suite judge.

## §A Shared conventions

### A.1 Commands live in fenced blocks, never in tables

Every deciding command is in a numbered fenced block (`CMD-ILT-nnn`) below the matrix and is
copied **verbatim** — the matrix only points at it. A pipe or a regex alternation written inside
a markdown table has to be escaped as `\|`, and copied verbatim that escape turns
`go test -run 'A|B'` into a literal-`|` pattern that selects nothing (`[no tests to run]`, a
vacuous green) and turns a `grep -E 'a|b'` alternation into a pattern that never matches.

### A.2 A go-test criterion must prove that its tests ran

Every go-test command runs with `-v`, and every go-test criterion carries two conditions on top of
PASS:

1. the output contains **no** `no tests to run`, and
2. the output contains one `--- PASS: <Name>` line for **each** test the criterion names
   (subtests counted by their full `Parent/sub` name).

A PASS with fewer `--- PASS:` lines than named tests is a selection defect and fails the
criterion.

### A.3 Stream separation

Criteria that assert where the warning is written use a test helper that captures standard output
and standard error in SEPARATE buffers. The existing `runIntegration` helper merges both streams
(`integration_lock_cli_test.go:32-34`) and cannot decide them. The warning is written through the
command's error writer (`cmd.ErrOrStderr()`), so a test buffer set with `SetErr` receives it.

### A.4 Fixture preamble (binding for every fixture command)

The fixture cells run a binary built **from this worktree** by the lead-gated compile slot; the
installed `moai` predates the change and is never used. The lead builds it once:

```bash
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && mkdir -p /tmp/t637-fx-bin && go build -o /tmp/t637-fx-bin/moai ./cmd/moai
```

Every fixture command is ONE Bash invocation, and three rules bind its shape:

1. **The binary is written literally.** Every call spells out `/tmp/t637-fx-bin/moai` — there is
   no `BIN` variable. The worktree-isolation guard refuses a command whose program name comes from
   a variable (`"$BIN" …`), so a variable-invoked binary makes the cell unrunnable in a lane
   session (plan-audit iter2 N1, reproduced there with `/bin/echo`).
2. **`EV` is the only variable, and only as a path argument.** `EV` names the evidence directory
   and appears solely as a redirect target or a file argument, never as the invoked program; the
   guard accepts that form (plan-audit iter2 probe). It is set at the start of the same invocation
   that uses it, because each Bash call is a fresh process and a variable set in an earlier call
   does not reach a later one. Preamble, verbatim:

   ```bash
   EV=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637/.moai/reports/t637/ac-evidence; mkdir -p "$EV" && test -x /tmp/t637-fx-bin/moai &&
   ```

3. **A cell that changes directory does so inside a subshell** — `( cd /tmp/t637-fx-wt/cardA && … )`
   — so the caller's working directory never persists past the cell. The Bash tool keeps its
   working directory between calls, so a bare `cd` would leave every later relative-path command
   (`git grep`, `diff -q`, `go test ./internal/...`, `make build`) running inside the fixture
   repository. The same defense is applied from the other side: every command that uses
   worktree-relative paths starts with `cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 &&`
   in its own invocation (§D.0), so its result does not depend on which command ran before it.

Every `acquire`, `status`, and `release` call carries the prefix `CLAUDE_PROJECT_DIR=/tmp/t637-fx`
in the same invocation, so the real window is never read or written.

### A.5 Fixture layout and rebuild recipe

Layout (present at plan time): root `/tmp/t637-fx` (branch `main`), worktrees
`/tmp/t637-fx-wt/develop` (`develop`), `/tmp/t637-fx-wt/cardA` (`WT-a`), `/tmp/t637-fx-wt/cardB`
(`WT-b`); configuration `/tmp/t637-fx/.moai/config/sections/git-strategy.yaml`. If the OS has
cleared `/tmp`, rebuild it in one invocation:

```bash
rm -rf /tmp/t637-fx /tmp/t637-fx-wt && mkdir -p /tmp/t637-fx/.moai/config/sections /tmp/t637-fx/.moai/state /tmp/t637-fx-wt && git -C /tmp/t637-fx init -q -b main && printf 'seed\n' > /tmp/t637-fx/seed.txt && git -C /tmp/t637-fx add seed.txt && git -C /tmp/t637-fx -c user.name=t -c user.email=t@example.invalid commit -q -m seed && git -C /tmp/t637-fx branch develop && git -C /tmp/t637-fx branch WT-a && git -C /tmp/t637-fx branch WT-b && git -C /tmp/t637-fx worktree add -q /tmp/t637-fx-wt/develop develop && git -C /tmp/t637-fx worktree add -q /tmp/t637-fx-wt/cardA WT-a && git -C /tmp/t637-fx worktree add -q /tmp/t637-fx-wt/cardB WT-b && printf 'git_strategy:\n  mode: "manual"\n  manual:\n    workflow: git-flow\n    develop_branch: develop\n' > /tmp/t637-fx/.moai/config/sections/git-strategy.yaml
```

Confirm with `git -C /tmp/t637-fx worktree list` (four entries: main, develop, WT-a, WT-b).

### A.6 Config toggle and real-window guard

```bash
# empty the develop branch (cells C1′, C2′)
sed -i '' 's/^\( *develop_branch:\).*/\1 ""/' /tmp/t637-fx/.moai/config/sections/git-strategy.yaml && grep -n develop_branch /tmp/t637-fx/.moai/config/sections/git-strategy.yaml
# restore it (cell C3′, and after the last cell)
sed -i '' 's/^\( *develop_branch:\).*/\1 develop/' /tmp/t637-fx/.moai/config/sections/git-strategy.yaml && grep -n develop_branch /tmp/t637-fx/.moai/config/sections/git-strategy.yaml
# real-window guard: run before AND after each fixture cell; the two outputs must be identical
shasum /Users/goos/MoAI/moai-adk-go/.moai/state/integration-lock.json 2>&1
```

A matching "No such file" on both sides counts as identical.

## §D AC Matrix

| AC | Requirement(s) | Claim | Deciding command | Observable that decides it |
|---|---|---|---|---|
| AC-ILT-001 | REQ-ILT-001 | A record carrying a card prints a `card:` line in text status | CMD-ILT-001 | PASS per §A.2 (1 named test); the test asserts the status text contains the line `  card:     t-fixture` for a record seeded with `Card: "t-fixture"` |
| AC-ILT-002 | REQ-ILT-002 | A card-less record prints no `card:` line | CMD-ILT-002 | PASS per §A.2 (2 named tests); the new test asserts `card:` is ABSENT; the pre-existing shape test passes with its assertions unedited |
| AC-ILT-003 | REQ-ILT-003, 004, 007 | Flag source — a non-blank `--branch` records `flag` with no warning, even in a git-flow project whose develop branch is empty; a blank `--branch` never records `flag` | CMD-ILT-003 | PASS per §A.2 (subtests `flag` and `blank_flag`); `flag`: `branch_source == "flag"`, branch = the flag value, stderr buffer has no `[moai:integration-lock] warning:`; `blank_flag` (`--branch "   "`, develop configured): `branch_source == "config"` |
| AC-ILT-004 | REQ-ILT-003, 004, 007 | Config source — git-flow with a configured develop branch records `config`, no warning | CMD-ILT-004 (Go) and CMD-ILT-013 (fixture cell C3′) | Go: PASS per §A.2 with `branch_source == "config"`, branch `fixture-integration`, stderr free of the prefix. Fixture: `rc=0`; `c3.warn-count` is `0`; status text has `  branch:   develop (source: config)`; `c3.status.json` contains `"branch_source":"config"` |
| AC-ILT-005 | REQ-ILT-003, 004, 006 | **Control 1 — same tree, own branch**: with an empty develop branch the record matches the merge target, is marked `caller`, and exactly one warning line is emitted | CMD-ILT-005 (Go) and CMD-ILT-011 (fixture cell C1′) | Go: PASS per §A.2; the stderr buffer holds exactly one line starting `[moai:integration-lock] warning:` that contains the caller branch and both remedy tokens `develop_branch` and `--branch`; `branch_source == "caller"`. Fixture: `rc=0`; `c1.warn-count` is `1` (this is also the positive control for the prefix pattern used by the `0` counts); that line names `WT-a`; status text has `  card:     tA`, `  branch:   WT-a (source: caller)` and a `worktree:` line ending `/t637-fx-wt/cardA`; real-window guard identical |
| AC-ILT-006 | REQ-ILT-001, 003, 004, 006 | **Control 2 — other tree, other branch**: a lane in tree A merging card B produces a record whose mismatch is visible from text status alone | CMD-ILT-006 (Go) and CMD-ILT-012 (fixture cell C2′) | Go: PASS per §A.2; from a caller tree on `main` with `--card t-other`, status text carries `card:     t-other` and `branch:   main (source: caller)`, and one warning line is emitted. Fixture: `rc=0` (not refused); `c2.warn-count` is `1`, naming `WT-a`; status text has `  card:     tB` together with `  branch:   WT-a (source: caller)`; real-window guard identical |
| AC-ILT-007 | REQ-ILT-006, 007 | **No-warn negatives** — each half of the git-flow predicate is pinned: github-flow with an EMPTY develop branch, a git-flow workflow in a non-manual mode with an EMPTY develop branch, a missing git strategy file, and the pre-existing github-flow cell | CMD-ILT-007 | PASS per §A.2 (4 named tests). In the three new cells the stderr buffer contains no `[moai:integration-lock] warning:` and `branch_source == "caller"`; the pre-existing t449 test passes with its assertions unedited |
| AC-ILT-008 | REQ-ILT-006, 008 | Warn-only: the warning neither refuses nor leaks onto standard output, and a refused acquire emits no warning | CMD-ILT-008 | PASS per §A.2 (2 named tests). `WarningIsOnStderrOnly`: nil error; record written; text-mode stdout equals exactly the `release-integration window acquired by <session> on <branch>` line; `--json` stdout decodes as ONE JSON object with `"acquired": true`; the warning appears only in the stderr buffer. `RefusedAcquireDoesNotWarn`: with the window held by another live session, acquire in the warning scenario returns the held error and the stderr buffer contains no warning prefix |
| AC-ILT-009 | REQ-ILT-005 | **Old-record compatibility** — a record without `branch_source` reads cleanly and is not rewritten | CMD-ILT-009 | PASS per §A.2; the test writes record JSON with NO `branch_source` key straight to the lock file, runs `status` and `status --json`, and asserts: nil error; branch line exactly `  branch:   <branch>` with no `(source:`; JSON lock object has no `branch_source` key; lock file bytes identical before and after |
| AC-ILT-010 | REQ-ILT-009 | The `--branch` help names the integration TARGET | CMD-ILT-010 | Go: PASS per §A.2; the usage string contains `integration target` (case-SENSITIVE, lowercase — plan.md M5 wording contains it verbatim) and does NOT contain `Branch being integrated`. Binary: the grep prints exactly one line, containing `integration target` and `git-flow develop branch` |
| AC-ILT-011 | REQ-ILT-010 | Every enumerated lane acquire site carries `--card <card-id>`; the release site says it has no card | CMD-ILT-014 | First count ≥ `8` and equal to the site count re-measured after run-phase absorbs `develop` (recorded in §E.2; plan-time base: `8` sites, `0` carrying `--card`); second count `0`; third count `1`; fourth count `0` (the release line carries no `--card <`) |
| AC-ILT-012 | REQ-ILT-011 | The template mirror stays byte-identical and neutral, and the embed is rebuilt | CMD-ILT-015 | `diff -q` prints nothing (exit 0); added-line count `1`; neutrality count `0`; positive control `1`; template tests PASS per §A.2 (1 named test); `make build` exits 0 |
| AC-ILT-013 | REQ-ILT-012, 013 | Invariants — resolution order, recorded values, config contract, and guard behavior unchanged | CMD-ILT-016 | All three test runs PASS per §A.2 (every named test has a `--- PASS:` line); pre-existing test functions keep their assertions — mechanical call-site adaptation to a changed signature is allowed and listed in §E.2; the guard diff prints nothing |
| AC-ILT-014 | REQ-ILT-001, 003, 005, 006, 007, 008 (mutation guard) | The new tests are not vacuous | §D.3 | Each mutated run FAILS its named test(s); each restored run PASSES; all outputs recorded verbatim in `progress.md` §E.2 |

## §D.0 Deciding commands (copy verbatim)

Every command below that uses a worktree-relative path begins with
`cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 &&` in its own invocation, and every
fixture cell runs its `cd` inside a subshell (§A.4 rule 3), so no command's result depends on the
working directory an earlier command left behind. `$EV` is the evidence directory from §A.4; each
block that uses it carries the preamble itself. A code block is copied verbatim with one declared
exception: none — every test name below is concrete.

**CMD-ILT-001**
```bash
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && go test ./internal/cli/... -run '^TestIntegrationStatus_ShowsCardLine$' -count=1 -v
```

**CMD-ILT-002**
```bash
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && go test ./internal/cli/... -run '^(TestIntegrationStatus_NoCardPrintsNoCardLine|TestIntegrationStatus_NoNameKeepsTodaysShape)$' -count=1 -v
```

**CMD-ILT-003**
```bash
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && go test ./internal/cli/... -run '^TestIntegrationAcquire_RecordsBranchSource$/^(flag|blank_flag)$' -count=1 -v
```

**CMD-ILT-004**
```bash
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && go test ./internal/cli/... -run '^TestIntegrationAcquire_RecordsBranchSource$/^config$' -count=1 -v
```

**CMD-ILT-005**
```bash
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && go test ./internal/cli/... -run '^TestIntegrationAcquire_GitFlowEmptyDevelopWarnsOnCallerFallback$' -count=1 -v
```

**CMD-ILT-006**
```bash
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && go test ./internal/cli/... -run '^TestIntegrationAcquire_CallerFallbackMismatchIsVisible$' -count=1 -v
```

**CMD-ILT-007**
```bash
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && go test ./internal/cli/... -run '^(TestIntegrationAcquire_GitHubFlowEmptyDevelopDoesNotWarn|TestIntegrationAcquire_NonManualModeGitFlowDoesNotWarn|TestIntegrationAcquire_NoConfigCallerFallbackDoesNotWarn|TestIntegrationAcquire_NonGitFlowFallsBackToCallerTree)$' -count=1 -v
```

The non-manual cell writes its own git strategy body, because the existing
`writeGitStrategyFixture` helper hard-codes `mode: manual`:

```yaml
git_strategy:
    mode: personal
    personal:
        workflow: git-flow
        develop_branch: ""
```

**CMD-ILT-008**
```bash
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && go test ./internal/cli/... -run '^(TestIntegrationAcquire_WarningIsOnStderrOnly|TestIntegrationAcquire_RefusedAcquireDoesNotWarn)$' -count=1 -v
```

**CMD-ILT-009**
```bash
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && go test ./internal/cli/... -run '^TestIntegrationStatus_OldRecordKeepsTodaysBranchLine$' -count=1 -v
```

**CMD-ILT-010**
```bash
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && go test ./internal/cli/... -run '^TestIntegrationAcquire_BranchFlagHelpNamesTheTarget$' -count=1 -v
```
```bash
/tmp/t637-fx-bin/moai integration acquire --help 2>&1 | grep -- '--branch'
```

The `--help` path needs no working directory and no `CLAUDE_PROJECT_DIR`: help output is printed
before the command body runs, so it neither reads nor writes a window.

**CMD-ILT-011 — fixture cell C1′ (Control 1).** Precondition: develop branch emptied (§A.6), real-window guard captured.
```bash
EV=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637/.moai/reports/t637/ac-evidence; mkdir -p "$EV" && test -x /tmp/t637-fx-bin/moai && ( cd /tmp/t637-fx-wt/cardA && CLAUDE_PROJECT_DIR=/tmp/t637-fx /tmp/t637-fx-bin/moai integration acquire --session fx-lane4 --name lane-4 --card tA 2>"$EV/c1.stderr"; echo "rc=$?" ); grep -c '^\[moai:integration-lock\] warning:' "$EV/c1.stderr" | tee "$EV/c1.warn-count"; CLAUDE_PROJECT_DIR=/tmp/t637-fx /tmp/t637-fx-bin/moai integration status | tee "$EV/c1.status.txt"; CLAUDE_PROJECT_DIR=/tmp/t637-fx /tmp/t637-fx-bin/moai integration status --json > "$EV/c1.status.json"; CLAUDE_PROJECT_DIR=/tmp/t637-fx /tmp/t637-fx-bin/moai integration release --session fx-lane4
```

**CMD-ILT-012 — fixture cell C2′ (Control 2).** Precondition: develop branch still emptied, guard captured. Intent: merge `WT-b` (card `tB`) while standing in cardA.
```bash
EV=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637/.moai/reports/t637/ac-evidence; mkdir -p "$EV" && test -x /tmp/t637-fx-bin/moai && ( cd /tmp/t637-fx-wt/cardA && CLAUDE_PROJECT_DIR=/tmp/t637-fx /tmp/t637-fx-bin/moai integration acquire --session fx-lane4 --name lane-4 --card tB 2>"$EV/c2.stderr"; echo "rc=$?" ); grep -c '^\[moai:integration-lock\] warning:' "$EV/c2.stderr" | tee "$EV/c2.warn-count"; CLAUDE_PROJECT_DIR=/tmp/t637-fx /tmp/t637-fx-bin/moai integration status | tee "$EV/c2.status.txt"; CLAUDE_PROJECT_DIR=/tmp/t637-fx /tmp/t637-fx-bin/moai integration status --json > "$EV/c2.status.json"; CLAUDE_PROJECT_DIR=/tmp/t637-fx /tmp/t637-fx-bin/moai integration release --session fx-lane4
```

**CMD-ILT-013 — fixture cell C3′ (configured target).** Precondition: develop branch restored (§A.6), guard captured.
```bash
EV=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637/.moai/reports/t637/ac-evidence; mkdir -p "$EV" && test -x /tmp/t637-fx-bin/moai && ( cd /tmp/t637-fx-wt/cardA && CLAUDE_PROJECT_DIR=/tmp/t637-fx /tmp/t637-fx-bin/moai integration acquire --session fx-lane4 --name lane-4 --card tB 2>"$EV/c3.stderr"; echo "rc=$?" ); grep -c '^\[moai:integration-lock\] warning:' "$EV/c3.stderr" | tee "$EV/c3.warn-count"; CLAUDE_PROJECT_DIR=/tmp/t637-fx /tmp/t637-fx-bin/moai integration status | tee "$EV/c3.status.txt"; CLAUDE_PROJECT_DIR=/tmp/t637-fx /tmp/t637-fx-bin/moai integration status --json > "$EV/c3.status.json"; CLAUDE_PROJECT_DIR=/tmp/t637-fx /tmp/t637-fx-bin/moai integration release --session fx-lane4
```

Only `acquire` needs the caller's tree, so only `acquire` sits inside the subshell; `status` and
`release` resolve the lock root from `CLAUDE_PROJECT_DIR` alone and run from the unchanged
directory. The warning count is taken by prefix, not by total stderr line count, so an unrelated
advisory the binary may print (a version-lag notice, for example) cannot flip the result. Cell
C1′'s `1` is the positive control for the `0` of cell C3′.

**CMD-ILT-014 — documented invocations**
```bash
EV=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637/.moai/reports/t637/ac-evidence; mkdir -p "$EV" && cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && git grep -n "moai integration acquire" -- .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md .claude/rules/local/gitflow-lane-protocol.md .claude/agents/harness/hns-release-specialist.md CLAUDE.local.md > "$EV/ac-011.txt"; wc -l < "$EV/ac-011.txt"; grep -v 'hns-release-specialist.md' "$EV/ac-011.txt" | grep -vc -- '--card <card-id>'; grep 'hns-release-specialist.md' "$EV/ac-011.txt" | grep -c 'no card'; grep 'hns-release-specialist.md' "$EV/ac-011.txt" | grep -c -- '--card <'
```

Expected, in order: a count ≥ `8` equal to the re-measured site count (AC-ILT-011); `0` (every
lane site carries `--card <card-id>`); `1` (the release site states it has no card); `0` (the
release site carries no `--card <` value).

**CMD-ILT-015 — template mirror, neutrality, embed** (each line is its own invocation and carries its own `cd`)
```bash
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && diff -q .claude/rules/moai/workflow/kanban-dispatch.md internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md; echo "diff-rc=$?"
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && git diff 1ad0fdc09 -- internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md | grep -c '^+[^+]'
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && git diff 1ad0fdc09 -- internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md | grep '^+[^+]' | grep -cE 't[0-9]{3}|SPEC-|20[0-9]{2}-[0-9]{2}-[0-9]{2}|[0-9a-f]{9}'
printf '+%s\n' 'see t637 here' | grep '^+[^+]' | grep -cE 't[0-9]{3}|SPEC-|20[0-9]{2}-[0-9]{2}-[0-9]{2}|[0-9a-f]{9}'
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && go test ./internal/template/... -run '^TestTemplateNoInternalContentLeak$' -count=1 -v
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && make build
```

Expected, in order: `diff-rc=0` with no diff output; `1`; `0`; `1` (the positive control — the
same pattern does match a card-id token; the control string is deliberately not SPEC-ID-shaped, so
no SPEC-reference scanner reads it as a citation); PASS per §A.2; exit 0. `TestRuleTemplateMirrorDrift`
is not used here: its path list does not include the kanban-dispatch rule
(`git grep -n "kanban-dispatch" -- 'internal/template/*_test.go'` printed nothing at plan time), so
the `diff -q` above is the mirror check.

**CMD-ILT-016 — invariants** (each line is its own invocation and carries its own `cd`)
```bash
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && go test ./internal/cli/... -run '^(TestResolveIntegrationTarget_ExplicitFlagWins|TestResolveIntegrationTarget_ConfiguredBranchUsedWhenFlagAbsent|TestResolveIntegrationTarget_NoWorktreeForBranchRecordsEmpty|TestResolveIntegrationTarget_NoConfigNoFlagFallsBackToCallerTree|TestWorktreeForBranch_FindsTheCheckedOutWorktree|TestWorktreeForBranch_UnknownBranchIsEmpty|TestIntegrationAcquire_RecordsTheIntegrationWorktreeNotTheCaller|TestIntegrationAcquire_ExplicitBranchResolvesItsWorktree|TestIntegrationAcquire_NoMatchingWorktreeRecordsEmptyWorktree|TestIntegration_AcquireStatusRelease)$' -count=1 -v
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && go test ./internal/config/... -run '^(TestLoadGitFlowDevelopBranch|TestLoadGitFlowIntegrationConfig_SeparatesGitFlowFromBranch)$' -count=1 -v
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && go test ./internal/hook/... -run 'IntegrationLock' -count=1 -v
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && git diff --stat 1ad0fdc09 -- internal/hook/integration_lock_guard.go
```

`TestLoadGitFlowIntegrationConfig_SeparatesGitFlowFromBranch` is the concrete name of the §D seam
test run-phase adds to `loader_integration_branch_test.go`. Its cases: (a) manual + git-flow +
develop set → git-flow true, branch returned; (b) manual + git-flow + develop empty → git-flow
true, branch empty; (c) manual + github-flow → git-flow false; (d) personal + git-flow → git-flow
false; (e) no file → git-flow false. Should run-phase choose a different name, the general rule at
the top of this file applies — the mapping is recorded in `progress.md` §E.2 and the anchored
`-run` selector above is updated to that exact name. The cli run must show ten `--- PASS:` lines;
the config run a `--- PASS:` line for each of the two named tests (plus subtests); the hook run at
least one.

## §D.1 Severity

MUST-PASS: AC-ILT-001, 003, 004, 005, 006, 007, 008, 009, 013, 014.

SHOULD-PASS (failure is reported and judged, never silently accepted): AC-ILT-002, 010, 011, 012.

CI-JUDGED: the full suite, including platform matrices, on the pushed `develop` head — not
claimable from a local run.

## §D.2 Given-When-Then scenarios

**AC-ILT-005 — Control 1, same tree / own branch (fixture cell C1′)**

- **Given** the fixture with `develop_branch` emptied (§A.6), the real-window guard captured, and
  `/tmp/t637-fx-bin/moai` built from this worktree (§A.4)
- **When** CMD-ILT-011 runs as one invocation
- **Then** `rc=0`; `c1.warn-count` is `1` and that line names `WT-a`; `c1.status.txt` shows
  `card:     tA`, `branch:   WT-a (source: caller)` and the cardA worktree; `c1.status.json`
  carries `"branch_source":"caller"` and `"card":"tA"`; the record matches the branch actually
  being merged; the real-window guard is identical

**AC-ILT-006 — Control 2, other tree / other branch (fixture cell C2′)**

- **Given** the same emptied fixture, and a lane standing in `/tmp/t637-fx-wt/cardA` whose intent
  is to merge `WT-b` (card `tB`)
- **When** CMD-ILT-012 runs as one invocation
- **Then** `rc=0` (not refused); `c2.warn-count` is `1`, naming `WT-a`; `c2.status.txt` shows
  `card:     tB` next to `branch:   WT-a (source: caller)` — the verdict's silent C2 mismatch is
  now readable from text status alone; the real-window guard is identical

**AC-ILT-004 — configured target (fixture cell C3′)**

- **Given** the fixture with `develop_branch: develop` restored
- **When** CMD-ILT-013 runs as one invocation
- **Then** `rc=0`; `c3.warn-count` is `0`; `c3.status.txt` shows `card:     tB` and
  `branch:   develop (source: config)` with the develop worktree; `c3.status.json` carries
  `"branch_source":"config"`

**AC-ILT-007 — no warning where the fallback is legitimate**

- **Given** a scratch repository whose git strategy is `github-flow` with an EMPTY develop branch;
  separately one whose mode is `personal` with a git-flow workflow and an EMPTY develop branch;
  separately one with no git strategy file
- **When** `acquire` runs from the repository root with no `--branch`
- **Then** in each the record's branch is the caller's `main`, `branch_source == "caller"`, and
  the stderr buffer contains no warning line — in the first two cells the git-flow predicate is
  the ONLY difference from the AC-ILT-005 warning cell

**AC-ILT-008 — refused acquire**

- **Given** the AC-ILT-005 warning scenario and a window already held by another live session
- **When** `acquire` runs without `--force`
- **Then** it returns the held error, writes no record, and emits no warning

**AC-ILT-009 — old record**

- **Given** a lock file whose JSON has no `branch_source` key (the shape every record written
  before this SPEC carries)
- **When** `status` and `status --json` run
- **Then** both succeed, the text branch line is today's exact line, the JSON lock object has no
  `branch_source` key, and the lock file's bytes are unchanged

## §D.3 Mutation guard (AC-ILT-014)

| Row | One-line mutation (production code) | Test(s) that MUST fail |
|---|---|---|
| a | The warning condition never holds (warning disabled) | `TestIntegrationAcquire_GitFlowEmptyDevelopWarnsOnCallerFallback` |
| b | The warning condition drops the workflow half of the git-flow predicate (warns when the mode is manual and the develop value is empty, whatever the workflow) | `TestIntegrationAcquire_GitHubFlowEmptyDevelopDoesNotWarn` |
| c | The status printer's `card:` line is removed | `TestIntegrationStatus_ShowsCardLine` |
| d | `branch_source` is hard-coded to `caller` | `TestIntegrationAcquire_RecordsBranchSource/flag`, `TestIntegrationAcquire_RecordsBranchSource/config` |
| e | The warning is written to standard output instead of the error writer | `TestIntegrationAcquire_WarningIsOnStderrOnly` |
| f | The provenance suffix is printed even when `branch_source` is empty | `TestIntegrationStatus_OldRecordKeepsTodaysBranchLine` |
| g | The warning condition drops the `mode == manual` half of the git-flow predicate | `TestIntegrationAcquire_NonManualModeGitFlowDoesNotWarn` |
| h | The source is decided from the untrimmed `--branch` value (`!= ""`) | `TestIntegrationAcquire_RecordsBranchSource/blank_flag` |
| i | The warning condition treats an absent or unreadable git strategy file as git-flow (equivalently: the whole git-flow predicate is dropped, so every caller fallback with an empty develop value warns) | `TestIntegrationAcquire_NoConfigCallerFallbackDoesNotWarn` |

Why row b and row i name different tests: with no git strategy file the mode is empty, so the
`mode == manual` half stays false and the row-b mutant (workflow half dropped) still emits no
warning there — the no-config test cannot fail for row b. It fails only for a mutant that stops
requiring a readable, satisfied predicate at all, which is row i. Conversely the github-flow cell
has mode `manual` and an empty develop value, so dropping the workflow half is exactly what turns
it red.

Each row is run with the corresponding named test command from §D.0, once mutated and once
restored. A mutation that leaves its named test green is a vacuous test and blocks the card until
the test is repaired; it is not "passed by construction".

## §D.4 Edge cases

- `--branch "   "` (blank) behaves as no flag: the configured tier or the caller decides, and
  `branch_source` names whichever did (never `flag`) — pinned by `blank_flag` and row h.
- A git-flow project with `develop_branch` blank AND a non-blank `--branch`: source `flag`, no
  warning — pinned by the `flag` subtest.
- A personal/team mode profile carrying `workflow: git-flow`: not git-flow, no warning — pinned by
  the non-manual cell and row g.
- A refused acquire emits no warning — pinned by `RefusedAcquireDoesNotWarn`.
- A stale record (holder gone) prints the card line and the provenance suffix like a live one.
- A record with `card` but no `branch_source` prints the card line and today's branch line.
- A record written by the new binary is readable by an older binary: the record decoder does not
  reject unknown keys, so the added key is ignored there (compatibility in the other direction;
  no criterion, stated for completeness).

## §D.5 Quality gates

```bash
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && go vet ./internal/cli/... ./internal/kanban/... ./internal/config/...
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && golangci-lint run ./internal/cli/... ./internal/kanban/... ./internal/config/...
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 && gofmt -l internal/cli internal/kanban internal/config
```

`go vet` exits 0; `golangci-lint` reports no new findings; `gofmt -l` prints nothing. No
`go test ./...` locally; the pushed `develop` head's CI is the full-suite verdict.

## §D.6 Definition of Done

- All MUST-PASS criteria observed PASS with verbatim evidence in `progress.md` §E.2, including the
  `--- PASS:` line counts required by §A.2; SHOULD-PASS results recorded (PASS, or FAIL with
  judgement).
- Mutation rows a-i each observed failing then passing.
- Fixture cells C1′, C2′, C3′ run with `/tmp/t637-fx-bin/moai`, real-window guard identical, and
  the fixture configuration restored to `develop_branch: develop` afterwards.
- Template mirror identical to the local copy; `make build` exit 0.
- The CLAUDE.local.md divergence risk (plan.md §E3) named in the completion report.

# Acceptance — SPEC-SESSION-MIDMOVE-001

Verification layer. Every criterion is Given-When-Then and binary. Commands run from the worktree root (`.claude/worktrees/t1279`). Path variables used below — set them in the same compound invocation as the check:

```bash
KT=internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md
KL=.claude/rules/moai/workflow/kanban-dispatch.md
DT=internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md
DL=.claude/rules/moai/workflow/kanban-dispatch-detail.md
WT=internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md
WL=.claude/rules/moai/workflow/worktree-integration.md
AT=internal/template/templates/AGENTS.md.tmpl
AL=AGENTS.md
LP=.claude/rules/local/gitflow-lane-protocol.md
CL=CLAUDE.local.md
E=.moai/reports/t1279/m1-measure.md
CAPS=.moai/reports/t1279/m1-caps.md
CTRL=.moai/reports/t1279/m1-control.txt
CMDS=.moai/reports/t1279/commands.txt
BASE=2370c5b31
```

`BASE` is the plan-phase HEAD. Baseline values measured at `BASE` are quoted where a criterion compares against them.

## §D — AC Matrix

### Measurement (M1)

- **AC-SMM-001** (REQ-SMM-001) — Given M1 has run, When `git log --format=%H --diff-filter=A -- "$CAPS"` and `git show --name-only --format= <that commit>` are read, Then the adding commit touches `$CAPS` only, `$CAPS` lists the four caps (6 sessions, 4 turns, `timeout -k 10 300`, 45 minutes), and `git log -1 --format=%cI <that commit>` is earlier than the `first_probe_ts:` value in `$E`.
- **AC-SMM-002** (REQ-SMM-002) — Given `$E` is committed, When `grep -oE '^(listing_trees|listing_count|turn_tokens)_(P1|P2|P3):' "$E" | sort -u | wc -l` runs, Then it prints `9`; and `grep -cE '^(launcher_single_listing|clear_restores_single_listing): (yes|no|gap)$' "$E"` prints `2`.
- **AC-SMM-003** (REQ-SMM-003) — Given `$E` and `$CMDS` are committed, When `grep -c '^judge_source: transcript$' "$E"` and `grep -ciE 'debug-file|--debug|/debug/' "$CMDS"` run, Then they print `1` and `0`.
- **AC-SMM-004** (REQ-SMM-004) — Given `$CTRL` is committed, When the extraction command recorded in `$CMDS` for the control is re-run on `$CTRL`, Then its output names at least one tree containing `.claude/worktrees/` (at plan time the source transcript shows 84 occurrences of the `worktrees/t1219:` prefix in the non-initial `skill_listing` at `2026-09-27T03:57:49.175Z`); and When any `listing_trees_P*` line names one tree, Then this criterion is required, not optional.
- **AC-SMM-005** (REQ-SMM-005) — Given `$E` is committed, When every key whose value is `gap` is listed, Then each has a matching `gap_reason_<key>:` line; `probes_run:` is at most `6`; `wall_clock_min:` is at most `45`; and any operator-run session appears in `$CMDS` as a line starting `# operator-run:`.
- **AC-SMM-006** (REQ-SMM-006) — Given `$CMDS` is committed, When its non-comment lines are read, Then every line starts with `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && timeout -k 10 300 claude ` and none contains `--setting-sources`, the repository root path, or `.claude/worktrees/t` followed by a digit; and `$E` shows `repo_head_before` = `repo_head_after` and `repo_branch_before` = `repo_branch_after`; and `git log --format= --name-only BASE..HEAD -- .moai/reports/t1279/ | grep -cE '\.jsonl$|debug'` prints `0`.

### Doctrine (M3)

- **AC-SMM-007** (REQ-SMM-007) — Given M3 is committed, When `grep -c 'moai cc -w <card-id>' "$KT" "$AT" "$LP"` and `grep -c 'wt: EnterWorktree(t0)' "$DT"` run, Then each of the first three counts is at least `1` and the last is `0` (at `BASE`: `$DT` count `1`).
- **AC-SMM-008** (REQ-SMM-008) — Given M3 is committed, When the section of `$KT` from `## The \`/clear\` handoff between phases` to the next `## ` heading is extracted with `sed -n '/^## The `\/clear` handoff between phases/,/^## [^T]/p' "$KT"`, Then it contains one paragraph in which a move precedes `/clear` and `/clear` precedes the re-send (checked with `grep -cE 'move.*`/clear`.*re-send'` ≥ `1`); the same `grep` on the `[HARD] **A new card starts in a new worktree` paragraph of `$KT` and on `$AT` §3 each prints at least `1`; and every sentence of that section at `BASE` (`git show BASE:"$KT"`) is still present verbatim.
- **AC-SMM-009** (REQ-SMM-009) — Given M3 is committed, When `grep -c 'EnterWorktree(<card-id>)' "$KT" "$LP"` runs, Then both print `0` (at `BASE`: `2` and `2`); and `grep -cE 'moai cc -w <card-id>` (또는|or) .*EnterWorktree' "$LP"` prints `0` (at `BASE`: `1`).
- **AC-SMM-010** (REQ-SMM-010) — Given `$E` and M3 are committed, When `clear_restores_single_listing` in `$E` is read, Then: for `yes`, `grep -c 'relaunch' <section from AC-SMM-008>` may be `0`; for `no` or `gap`, it is at least `1` and the section contains no sentence claiming `/clear` removes the listing (`grep -ciE '/clear.{0,40}(removes|drops|clears) the (duplicate|listing)'` prints `0`). Likewise, when `launcher_single_listing` is `no` or `gap`, `grep -ciE 'launcher.{0,60}(one|single) skill listing' "$KT" "$AT" "$WT"` prints `0`.
- **AC-SMM-011** (REQ-SMM-011) — Given DP-2's resolution is recorded in progress.md §E.2, When the kanban integration bullets of `$KT` (the lines containing `Enter, do not redirect` and `Return the same way`) are read, Then: for exempt, `grep -ciE 'exempt|exception' ` on those bullets and their paragraph prints at least `1` and `$LP`, `$CL` carry the same word near their `EnterWorktree(` integration lines; for covered, those bullets contain `/clear` and `re-send`.

### Where the doctrine lives

- **AC-SMM-012** (REQ-SMM-012) — Given M3 is committed, When `cmp "$KT" "$KL"`, `cmp "$DT" "$DL"`, `cmp "$WT" "$WL"`, and `diff <(sed -n '/^## 3. Worktrees/,/^## 4\./p' "$AL") <(sed -n '/^## 3. Worktrees/,/^## 4\./p' "$AT")` run, Then all four exit `0` (at `BASE`: all `0`); and for every commit in `BASE..HEAD` that touches `$KL`, `$DL`, or `$WL`, the same commit also touches the matching template path; and `go test ./internal/config/ -run '^TestCodexContractByteCeiling$|^TestAlwaysLoadedTokenBudget$' -count=1 -v` exits `0` with output lines starting `--- PASS: TestCodexContractByteCeiling (` and `--- PASS: TestAlwaysLoadedTokenBudget (`.
- **AC-SMM-013** (REQ-SMM-013) — Given M3 is committed, When `git diff -U0 BASE..HEAD -- "$KT" "$DT" "$WT" "$AT" | grep '^+[^+]' | grep -cE '\bt[0-9]{2,}\b|SPEC-[A-Z]|[0-9]{4}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{9,40}\b'` runs, Then it prints `0`; positive control: the same final `grep -cE` on `printf '+card t1279 SPEC-X-001 2026-09-27 2370c5b31\n'` prints `1`; and the added-line count (`git diff -U0 BASE..HEAD -- "$KT" | grep -c '^+[^+]'`) is at least `1`, so the sweep is not empty.
- **AC-SMM-014** (REQ-SMM-014) — Given M3 is committed, When `grep -c '\[HARD\]'` is run on each of `$KT`, `$DT`, `$WT`, `$LP`, `$CL`, Then each count is at least its `BASE` value (38, 8, 18, 21, 48); and each of these strings is found in `$KT`: `## Isolation is provisioned by MoAI, then entered through a launcher`, `Card worktree branches carry the \`WT-\` prefix`, `## The \`/clear\` handoff between phases`, `A new card starts in a new worktree — exit any previous one first`, `## Completion is read, never trusted`.
- **AC-SMM-015** (REQ-SMM-015) — Given M3 is committed, When `grep -c 'moai cc -w <card-id>' "$LP"` and `grep -cE '/clear' "$LP"` run, Then each prints at least `1`, and `test -e internal/template/templates/.claude/rules/local` exits non-zero and `grep -rl 'gitflow-lane-protocol' internal/template/templates/ | wc -l` prints `0`.

### Guard (conditional on DP-1)

- **AC-SMM-016** (REQ-SMM-016) — Given DP-1's resolution is recorded in progress.md §E.2:
  - **warn:** When `printf '%s' '<PostToolUse EnterWorktree payload with cwd=<fixture>/.claude/worktrees/w1>' | go run ./cmd/moai hook post-tool` runs in a scratch fixture, Then it exits `0` and its JSON output contains `/clear` in the model-visible context field; When the payload's cwd is an exempt integration worktree, Then that field does not contain `/clear`.
  - **block:** When the equivalent PreToolUse payload is piped to `go run ./cmd/moai hook pre-tool` with the flag disabled, Then the output carries no deny; with the flag enabled and a card-worktree target, Then the output is a deny whose reason starts with the sentinel; with an exempt target, Then no deny; and spec.md frontmatter reads `tier: L` with `design.md` and `research.md` present before the first run-phase commit.
  - **docs-only:** When `git diff --name-only BASE..HEAD -- internal/hook/` runs, Then it prints nothing.

## §D.1 Edge cases

- A probe whose transcript lacks any `skill_listing` attachment: recorded as `gap` with reason, never as `0` trees.
- `/clear` issues a new session id: P3 reads the transcript of the post-clear session id, and both ids are recorded in `$E`.
- A lane that must move for a non-card reason (resume of a generic session): governed by session-handoff Block 0, out of scope; the warn notice (DP-1 warn) still fires for card-worktree targets only.

## §D.2 Quality gate

- `moai spec lint .moai/specs/SPEC-SESSION-MIDMOVE-001` exits `0`.
- `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json` exits `0`.
- `make build` succeeds after template edits.
- Under DP-1 warn/block: `go test ./internal/hook/... -count=1` exits `0`.

## §D.3 Traceability

| REQ | AC |
|---|---|
| REQ-SMM-001 | AC-SMM-001 |
| REQ-SMM-002 | AC-SMM-002 |
| REQ-SMM-003 | AC-SMM-003 |
| REQ-SMM-004 | AC-SMM-004 |
| REQ-SMM-005 | AC-SMM-005 |
| REQ-SMM-006 | AC-SMM-006 |
| REQ-SMM-007 | AC-SMM-007 |
| REQ-SMM-008 | AC-SMM-008 |
| REQ-SMM-009 | AC-SMM-009 |
| REQ-SMM-010 | AC-SMM-010 |
| REQ-SMM-011 | AC-SMM-011 |
| REQ-SMM-012 | AC-SMM-012 |
| REQ-SMM-013 | AC-SMM-013 |
| REQ-SMM-014 | AC-SMM-014 |
| REQ-SMM-015 | AC-SMM-015 |
| REQ-SMM-016 | AC-SMM-016 |

## §D.4 Definition of Done

- DP-1, DP-2, DP-3 resolutions recorded in progress.md §E.2 before M1.
- AC-SMM-001 … AC-SMM-016 each PASS or N/A with the reason written (only AC-SMM-016 sub-branches not selected by DP-1, and AC-SMM-011's unselected branch, may be N/A).
- Quality gate §D.2 green; evidence files force-added and committed; nothing pushed by the lane.

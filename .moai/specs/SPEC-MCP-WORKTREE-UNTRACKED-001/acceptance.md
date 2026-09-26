# SPEC-MCP-WORKTREE-UNTRACKED-001 — Acceptance

Verification layer. Each criterion is binary and runs under `t.TempDir()`.
Common fixture **F**: a git repository `P` with `.moai/` listed in `.gitignore`,
a `.moai/` directory present in `P` (holding at least one SPEC under
`.moai/specs/`), one commit, and a linked worktree `W` made with
`git worktree add`. `W` therefore has no `.moai`.

## §D AC Matrix

- **AC-MWU-001 (reproduction-first; REQ-MWU-002).**
  Given fixture F on the unchanged validator,
  When `project_root` is set to `W`,
  Then the call is rejected with the "has no .moai directory" error — this RED
  result is committed in its own commit before the fix;
  and Given the fix, When the same call is made,
  Then it is accepted and the returned tree root equals the canonical path of
  `W` (not `P`).

- **AC-MWU-002 (diff collection targets the worktree; REQ-MWU-007).**
  Given fixture F with an uncommitted change to a tracked file in `W` only and a
  different uncommitted change in `P` only,
  When the diff collected for an uncommitted-target audit is taken from the
  resolved tree root for `project_root = W`,
  Then the collected diff contains the `W` change and does not contain the `P`
  change.

- **AC-MWU-003 (negative — unrelated directory; REQ-MWU-003).**
  Given a fresh `t.TempDir()` directory that is not a git repository and has no
  `.moai`,
  When it is passed as `project_root`,
  Then the call is rejected and no default root is used.

- **AC-MWU-004 (negative — non-MoAI repository worktree; REQ-MWU-003).**
  Given a repository whose primary checkout has no `.moai`, with a linked
  worktree,
  When the worktree is passed as `project_root`,
  Then the call is rejected with an error naming the missing primary `.moai`.

- **AC-MWU-005 (negative — subdirectory of a worktree; REQ-MWU-003).**
  Given fixture F,
  When a subdirectory of `W` is passed as `project_root`,
  Then it is rejected as not a worktree top level.

- **AC-MWU-006 (canonical containment; REQ-MWU-005).**
  Given fixture F and a symlink `L` pointing at `W`,
  When `L` is passed as `project_root`,
  Then it is accepted and the returned tree root is the canonical path of `W`;
  and Given a symlink pointing at the AC-MWU-003 directory,
  Then it is rejected.

- **AC-MWU-007 (environment isolation; REQ-MWU-006).**
  Given `GIT_DIR` set (non-parallel test, `t.Setenv`) to `P`'s git directory,
  When the AC-MWU-003 directory is passed as `project_root`,
  Then it is still rejected.

- **AC-MWU-008 (fail-closed; REQ-MWU-004).**
  Given fixture F and a `PATH` from which git cannot be found for the
  inspection,
  When `W` is passed as `project_root`,
  Then the call is rejected, not accepted and not redirected to a default.

- **AC-MWU-009 (`.moai` root and provenance; REQ-MWU-007, REQ-MWU-008).**
  Given fixture F,
  When `spec_progress` is called with `project_root = W`,
  Then the returned count equals the number of SPECs under `P/.moai/specs`
  (non-zero), and the `_root` block names `W` as the tree root and `P` as the
  `.moai` root.

- **AC-MWU-010 (unchanged paths and docs; REQ-MWU-001, REQ-MWU-009, REQ-MWU-010).**
  Given the change,
  When the existing `project_root` tests run, and `diff` compares
  `.claude/rules/moai/core/moai-mcp-tools.md` with its template mirror, and the
  template neutrality guards run,
  Then the existing tests pass unchanged, `diff` exits 0, the rule's situation
  table carries a linked-worktree row, and the neutrality guards pass with no
  SPEC ID, card id, or date in the template copy.

## §D.1 Edge cases

- A worktree that has its own `.moai` (tracked or created later) takes the
  unchanged path (REQ-MWU-001); its `.moai` root is itself.
- A worktree whose admin entry was pruned is rejected (plan §C question 3).
- A bare primary is rejected (spec §5).

## §D.2 Quality gate

- New and existing tests in `./internal/cli/` pass; `go vet` and
  `golangci-lint` clean on the touched package.
- Coverage of the new validator branch is exercised by AC-MWU-001, 003–008.

## §D.3 Definition of Done

- AC-MWU-001..010 PASS with verbatim evidence in progress.md §E.2.
- RED commit precedes the fix commit in `git log`.
- Both rule copies byte-identical; `make build` run after the template edit.

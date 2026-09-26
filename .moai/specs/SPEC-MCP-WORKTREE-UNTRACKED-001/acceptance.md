# SPEC-MCP-WORKTREE-UNTRACKED-001 — Acceptance

Verification layer. Each criterion is binary and runs under `t.TempDir()`.

Common fixture **F**: a git repository `P` with `.moai/` listed in `.gitignore`;
`P/.moai/config/sections/` present with a `workflow.yaml`; at least one SPEC
under `P/.moai/specs/`; one commit; and a linked worktree `W` made with
`git worktree add`. `W` therefore has no `.moai`.

## §D AC Matrix

### Acceptance and rejection

- **AC-MWU-001 (reproduction-first; REQ-MWU-002).**
  Given fixture F,
  When `project_root` is set to `W`,
  Then the call is accepted with tree root equal to the canonical path of `W`
  and source root equal to the canonical path of `P`.
  RED-first is witnessed by three predicates, all required:
  (i) `git merge-base --is-ancestor <red> <fix>` exits 0;
  (ii) `git show --name-only --format= <red>` lists only `_test.go` files;
  (iii) `go test ./internal/cli/ -run <test> -count=1` run on the `<red>` tree
  exits non-zero with the "has no .moai directory" rejection in its output.

- **AC-MWU-002 (diff collection targets the worktree; REQ-MWU-007).**
  Given fixture F with an uncommitted change to a tracked file in `W` only and a
  different uncommitted change in `P` only,
  When `collectReviewDiff` (the diff collector `claude_audit` uses) runs on the
  tree root resolved for `project_root = W` with target `uncommittedChanges`,
  Then the diff contains the `W` change and not the `P` change.

- **AC-MWU-003 (negative — unrelated directory; REQ-MWU-003).**
  Given a fresh `t.TempDir()` directory that is not a git repository,
  When it is passed as `project_root`,
  Then the call is rejected and no default root is used.

- **AC-MWU-004 (negative — non-MoAI repository; REQ-MWU-003).**
  Given a repository whose primary checkout has no MoAI configuration, with a
  linked worktree,
  When the worktree is passed, and separately the primary itself is passed,
  Then both are rejected, the worktree error naming the primary's missing
  configuration.

- **AC-MWU-005 (negative — subdirectory; REQ-MWU-003).**
  Given fixture F,
  When a subdirectory of `W` is passed as `project_root`,
  Then it is rejected as not a worktree top level.

- **AC-MWU-006 (negative — unregistered worktree; REQ-MWU-002, REQ-MWU-003).**
  Given fixture F after the worktree's admin entry under `P/.git/worktrees/` is
  removed,
  When `W` is passed as `project_root`,
  Then it is rejected with an error naming the unregistered or unreadable
  worktree.

- **AC-MWU-007 (negative — ambiguous layout; REQ-MWU-004).**
  Given a repository created with `git init --separate-git-dir` whose working
  tree has MoAI configuration, and a linked worktree of it without any,
  When the worktree is passed as `project_root`,
  Then it is rejected as an ambiguous layout.

- **AC-MWU-008 (canonical comparison; REQ-MWU-006).**
  Given fixture F and a symlink `L` pointing at `W`,
  When `L` is passed,
  Then it is accepted with tree root equal to the canonical path of `W`;
  and Given a symlink to the AC-MWU-003 directory, Then it is rejected.

- **AC-MWU-009 (environment isolation; REQ-MWU-006).**
  Given fixture F and, in a non-parallel test, `GIT_DIR` set to `P`'s git dir
  and `GIT_WORK_TREE` set to `P`,
  When `W` is passed as `project_root`,
  Then it is still accepted with tree root `W` and source root `P`
  (an implementation that inherits the variables sees `P`'s git dir and rejects
  `W`);
  and Given the same environment, When the AC-MWU-003 directory is passed, Then
  it is rejected.

- **AC-MWU-010 (fail-closed; REQ-MWU-005).**
  Given fixture F and a `PATH` from which git cannot be found,
  When `W` is passed as `project_root`,
  Then it is rejected, not accepted and not redirected to a default.

### Per-class routing

- **AC-MWU-011 (class C — catalogue and config; REQ-MWU-008, REQ-MWU-012).**
  Given fixture F with `P/.moai/config/sections/workflow.yaml` declaring the
  codex audit gate `required`,
  When `spec_progress` is called with `project_root = W`, and the codex
  gate-required check runs for the same root,
  Then the SPEC count equals the number of SPECs under `P/.moai/specs`
  (non-zero), the `_root` block names `W` as tree root and `P` as source root,
  and the gate reads as `required`.

- **AC-MWU-012 (class G — no primary graph answer; REQ-MWU-009).**
  Given fixture F with a graph artifact under `P/.moai/project/graph/` and none
  under `W`,
  When `graph_find_code` is called with `project_root = W`,
  Then it returns the "graph layer absent" error and no match from `P`'s graph.

- **AC-MWU-013 (class S — writes stay with the tree; REQ-MWU-010).**
  Given fixture F,
  When `verify_snapshot` records a check with `project_root = W`, and
  `verify_trend` then reads the same key,
  Then the snapshot file exists under `W/.moai/state/verify/snapshots/`, none is
  created under `P/.moai/state/verify/snapshots/`, and the trend returns the
  recorded check.

- **AC-MWU-014 (no resolution flip; REQ-MWU-011).**
  Given fixture F after AC-MWU-013 has created `W/.moai/state/`,
  When `spec_progress` is called again with `project_root = W`,
  Then the source root is still `P` and the SPEC count is unchanged.

### Compatibility and documentation

- **AC-MWU-015 (unchanged paths and docs; REQ-MWU-001, REQ-MWU-013, REQ-MWU-014).**
  Given the change,
  When the existing `project_root` tests run, `diff` compares
  `.claude/rules/moai/core/moai-mcp-tools.md` with its template mirror, and the
  template neutrality guards run,
  Then the existing tests pass unchanged, `diff` exits 0, the rule's situation
  table carries a linked-worktree row, `projectRootDescCommon` mentions
  linked-worktree acceptance and the tree/source root split, and the neutrality
  guards pass with no SPEC ID, card id, or date in the template copy.

## §D.1 Edge cases

- A worktree whose own `.moai` carries MoAI configuration (tracked repository,
  or a hand-made copy) takes the unchanged path (REQ-MWU-001).
- A bare primary or submodule-internal git dir is rejected (REQ-MWU-004).

## §D.2 Quality gate

- New and existing tests in the touched packages pass; `go vet` and
  `golangci-lint` clean on them.

## §D.3 Definition of Done

- AC-MWU-001..015 PASS with verbatim evidence in progress.md §E.2.
- Both rule copies byte-identical; `make build` run after the template edit.

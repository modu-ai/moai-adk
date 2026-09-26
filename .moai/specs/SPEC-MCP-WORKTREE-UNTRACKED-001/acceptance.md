# SPEC-MCP-WORKTREE-UNTRACKED-001 — Acceptance

Verification layer. Each criterion is binary and runs under `t.TempDir()`.

Common fixture **F**: a git repository `P` with `.moai/` listed in `.gitignore`,
a `.moai/` directory present in `P`, one commit, and a linked worktree `W` made
with `git worktree add`. `W` therefore has no `.moai`.

## §D AC Matrix

### Acceptance and rejection

- **AC-MWU-001 (reproduction-first; REQ-MWU-002).**
  Given fixture F,
  When `validateProjectRoot` is called with `W`,
  Then it returns no error and returns the canonical path of `W`.
  The test uses only the existing signature `validateProjectRoot(string) (string, error)`,
  so it compiles on the pre-fix tree. RED-first is witnessed by three
  predicates, all required:
  (i) `git merge-base --is-ancestor <red> <fix>` exits 0;
  (ii) `git show --name-only --format= <red>` lists only `_test.go` files;
  (iii) `go test ./internal/cli/ -run <test> -count=1` on the `<red>` tree exits
  non-zero, and its output carries the "has no .moai directory" rejection (not a
  build failure).

- **AC-MWU-002 (tree operations target the worktree; REQ-MWU-008).**
  Given fixture F with an uncommitted change to a tracked file in `W` only and a
  different uncommitted change in `P` only,
  When `collectReviewDiff` runs with target `uncommittedChanges` on the root
  resolved for `project_root = W`, and `codex_audit` is invoked with
  `project_root = W` through its existing runner test seam,
  Then the diff contains the `W` change and not the `P` change, and the `cwd`
  handed to codex is the canonical path of `W`.

- **AC-MWU-003 (negative — unrelated directory; REQ-MWU-003).**
  Given a fresh `t.TempDir()` directory that is not a git repository and has no
  `.moai`,
  When it is passed as `project_root`,
  Then the call is rejected and no default root is used.

- **AC-MWU-004 (negative — non-MoAI repository; REQ-MWU-003).**
  Given a repository whose primary checkout has no `.moai`, with a linked
  worktree,
  When the worktree is passed, and separately the primary itself is passed,
  Then both are rejected, the worktree error naming the primary's missing `.moai`.

- **AC-MWU-005 (negative — subdirectory; REQ-MWU-003).**
  Given fixture F,
  When a subdirectory of `W` is passed as `project_root`,
  Then it is rejected as not a worktree top level.

- **AC-MWU-006 (negative — unregistered worktree; REQ-MWU-002, REQ-MWU-003).**
  Given fixture F after `W`'s admin entry under `P/.git/worktrees/` is removed,
  When `W` is passed as `project_root`,
  Then it is rejected with an error naming the unregistered or unreadable
  worktree.

- **AC-MWU-007 (negative — ambiguous layout; REQ-MWU-004).**
  Given a repository created with `git init --separate-git-dir` whose working
  tree has a `.moai` directory, and a linked worktree of it without one,
  When the worktree is passed as `project_root`,
  Then it is rejected as an ambiguous layout.

- **AC-MWU-008 (canonical comparison; REQ-MWU-006).**
  Given fixture F and a symlink `L` pointing at `W`,
  When `L` is passed,
  Then it is accepted and the returned path is the canonical path of `W`;
  and Given a symlink to the AC-MWU-003 directory, Then it is rejected.

- **AC-MWU-009 (environment isolation; REQ-MWU-006).**
  Given fixture F and, in a non-parallel test, `GIT_DIR` set to `P`'s git dir
  and `GIT_WORK_TREE` set to `P`,
  When `W` is passed as `project_root`,
  Then it is still accepted with the canonical path of `W`
  (an implementation that inherits the variables sees `P`'s git dir and rejects
  `W`);
  and Given the same environment, When the AC-MWU-003 directory is passed, Then
  it is rejected.

- **AC-MWU-010 (fail-closed; REQ-MWU-005).**
  Given fixture F and a `PATH` from which git cannot be found,
  When `W` is passed as `project_root`,
  Then it is rejected, not accepted and not redirected to a default.

- **AC-MWU-011 (dangling sibling; REQ-MWU-007).**
  Given fixture F plus a second linked worktree `W2` whose directory is then
  deleted without running `git worktree prune`,
  When `W` is passed as `project_root`,
  Then it is accepted with the canonical path of `W`.

### Graph and compatibility

- **AC-MWU-012 (graph never from the primary; REQ-MWU-008).**
  Given fixture F with a graph artifact under `P/.moai/project/graph/` and none
  under `W`,
  When `graph_find_code` is called with `project_root = W`,
  Then it returns the "graph layer absent" error and no match from `P`'s graph.

- **AC-MWU-013 (today's accepted set preserved; REQ-MWU-001).**
  Given a directory that is not a git repository and contains only `.moai/`
  (and, separately, only `.moai/specs/<id>`),
  When it is passed as `project_root`,
  Then it is accepted with its canonical path, exactly as on the pre-fix tree.

- **AC-MWU-014 (unchanged paths and docs; REQ-MWU-001, REQ-MWU-009, REQ-MWU-010).**
  Given the change,
  When the existing `project_root` tests (including every fixture built by
  `newProbeProject` or a bare `.moai` directory) run, `diff` compares
  `.claude/rules/moai/core/moai-mcp-tools.md` with its template mirror, and the
  template neutrality guards run,
  Then the existing tests pass unchanged, `diff` exits 0, the rule's situation
  table carries a linked-worktree row, `projectRootDescCommon` mentions
  linked-worktree acceptance, and the neutrality guards pass with no SPEC ID,
  card id, or date in the template copy.

## §D.1 Edge cases

- A worktree whose own `.moai` exists (tracked repository, a partially tracked
  `.moai` such as tracked config with ignored specs, or a directory created later
  by a state write) takes the REQ-MWU-001 branch, as it does today; what it then
  reads from that `.moai` is t1213's scope.
- A bare primary or submodule-internal git dir is rejected (REQ-MWU-004).

## §D.2 Quality gate

- New and existing tests in the touched packages pass; `go vet` and
  `golangci-lint` clean on them.

## §D.3 Definition of Done

- AC-MWU-001..014 PASS with verbatim evidence in progress.md §E.2.
- Both rule copies byte-identical; `make build` run after the template edit.

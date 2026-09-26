# SPEC-MCP-WORKTREE-UNTRACKED-001 — Acceptance

Verification layer. Each criterion is binary and runs under `t.TempDir()`.

Common fixture **F**: a git repository `P` with `.moai/` listed in `.gitignore`,
a `.moai/` directory present in `P` (including
`.moai/config/sections/workflow.yaml` and at least one SPEC under
`.moai/specs/`), one commit, and a linked worktree `W` made with
`git worktree add`. `W` therefore has no `.moai`, and is config-orphaned.

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

- **AC-MWU-004 (negative — non-MoAI repository, primary itself, subdirectory; REQ-MWU-003).**
  Given a repository whose primary checkout has no `.moai`, with a linked
  worktree, and separately fixture F,
  When the non-MoAI worktree, the non-MoAI primary itself, and a subdirectory of
  `W` are each passed as `project_root`,
  Then all three are rejected — the first error naming the primary's missing
  `.moai`, the third naming a non-top-level path.

- **AC-MWU-005 (negative — unregistered worktree; REQ-MWU-002, REQ-MWU-003).**
  Given fixture F after `W`'s admin entry under `P/.git/worktrees/` is removed,
  When `W` is passed as `project_root`,
  Then it is rejected with an error naming the unregistered or unreadable
  worktree.

- **AC-MWU-006 (negative — ambiguous layout; REQ-MWU-004).**
  Given a repository created with `git init --separate-git-dir` whose working
  tree has a `.moai` directory, and a linked worktree of it without one,
  When the worktree is passed as `project_root`,
  Then it is rejected as an ambiguous layout.

- **AC-MWU-007 (canonical comparison; REQ-MWU-006).**
  Given fixture F and a symlink `L` pointing at `W`,
  When `L` is passed,
  Then it is accepted and the returned path is the canonical path of `W`;
  and Given a symlink to the AC-MWU-003 directory, Then it is rejected.

- **AC-MWU-008 (environment isolation; REQ-MWU-006).**
  Given fixture F and, in a non-parallel test, `GIT_DIR` set to `P`'s git dir
  and `GIT_WORK_TREE` set to `P`,
  When `W` is passed as `project_root`,
  Then it is still accepted with the canonical path of `W`
  (an implementation that inherits the variables sees `P`'s git dir and rejects
  `W`);
  and Given the same environment, When the AC-MWU-003 directory is passed, Then
  it is rejected.

- **AC-MWU-009 (fail-closed; REQ-MWU-005).**
  Given fixture F and a `PATH` from which git cannot be found,
  When `W` is passed as `project_root`,
  Then it is rejected, not accepted and not redirected to a default.

- **AC-MWU-010 (dangling sibling; REQ-MWU-007).**
  Given fixture F plus a second linked worktree `W2` whose directory is then
  deleted without running `git worktree prune`,
  When `W` is passed as `project_root`,
  Then it is accepted with the canonical path of `W`.

### Graph and compatibility

- **AC-MWU-011 (graph never from the primary; REQ-MWU-008).**
  Given fixture F with a graph artifact under `P/.moai/project/graph/` and none
  under `W`,
  When `graph_find_code` is called with `project_root = W`,
  Then it returns the "graph layer absent" error and no match from `P`'s graph.

- **AC-MWU-012 (today's accepted set preserved, no subprocess; REQ-MWU-001).**
  Given a directory that is not a git repository and contains only `.moai/`
  (and, separately, only `.moai/specs/<id>`),
  When it is passed as `project_root` — once normally, and once with a `PATH`
  from which git cannot be found and `GIT_DIR` set to a nonexistent path —
  Then it is accepted with its canonical path in all runs, exactly as on the
  pre-fix tree (acceptance under a git-less `PATH` shows the branch ran no git
  subprocess).

- **AC-MWU-013 (unchanged tests and docs; REQ-MWU-001, REQ-MWU-009, REQ-MWU-010).**
  Given the change,
  When the existing `project_root` tests (including every fixture built by
  `newProbeProject` or a bare `.moai` directory) run, `diff` compares
  `.claude/rules/moai/core/moai-mcp-tools.md` with its template mirror, and the
  template neutrality guards run,
  Then the existing tests pass unchanged; `diff` exits 0; the rule's situation
  table carries a linked-worktree row; both rule copies and
  `projectRootDescCommon` each state (a) linked-worktree acceptance, (b) that the
  audit gate of a worktree without its own workflow config is read from the
  primary checkout, and (c) that other configuration, the catalogue, and state
  are read from the accepted tree; and the neutrality guards pass with no SPEC
  ID, card id, or date in the template copy.

### Audit gate and catalogue warning

- **AC-MWU-014 (primary gate enforced on a config-orphaned worktree, before and after self-promotion; REQ-MWU-011).**
  Given fixture F with `P/.moai/config/sections/workflow.yaml` declaring
  `workflow.audit.gates.codex: required`, and a codex runner test seam that
  produces no verdict,
  When `codex_audit` is called with `project_root = W`, and `audit_multi` is
  called with `project_root = W` using stub backends that return a Claude
  `pass` and no codex verdict,
  Then `codex_audit` returns verdict `fail` with a non-empty `gate_unmet`, and
  `audit_multi` returns an `overall_verdict` other than `pass`;
  and When both calls are repeated after the first `codex_audit` call has created
  `W/.moai/state/` (the receipt write), Then the same two results hold.

- **AC-MWU-015 (gate read fails closed; non-worktree unchanged; REQ-MWU-012).**
  Given the AC-MWU-006 `--separate-git-dir` repository with a linked worktree
  `W3` into which a bare `.moai/` directory (no `workflow.yaml`) has been placed,
  so that `project_root = W3` is accepted through REQ-MWU-001,
  When `codex_audit` is called with `project_root = W3` and the codex seam
  produces no verdict,
  Then the result is verdict `fail` with a non-empty `gate_unmet`;
  and Given the AC-MWU-012 non-git directory with only `.moai/`, When the same
  call is made, Then the result is the fail-open `inconclusive` it returns today
  (no gate applied).

- **AC-MWU-016 (catalogue/state warning; REQ-MWU-013).**
  Given fixture F,
  When `spec_progress`, `spec_audit`, `spec_drift`, and `verify_trend` are each
  called with `project_root = W`, and again after `W/.moai/state/` exists,
  Then every response carries a `_root` block whose `warning` states the answer
  was read from the worktree tree and may be empty because `.moai` is not
  tracked;
  and Given the primary `P` passed as `project_root`, Then no such warning is
  present.

## §D.1 Edge cases

- A worktree whose own `.moai` exists (tracked repository, a partially tracked
  `.moai` such as tracked config with ignored specs, or a directory created later
  by a state write) takes the REQ-MWU-001 branch, as it does today; whether it is
  config-orphaned depends only on its own `workflow.yaml` (REQ-MWU-011..013).
- A bare primary or submodule-internal git dir is rejected by the validator
  (REQ-MWU-004) and makes the gate read fail closed (REQ-MWU-012).

## §D.2 Quality gate

- New and existing tests in the touched packages pass; `go vet` and
  `golangci-lint` clean on them.

## §D.3 Definition of Done

- AC-MWU-001..016 PASS with verbatim evidence in progress.md §E.2.
- Both rule copies byte-identical; `make build` run after the template edit.

# SPEC-WORKTREE-CREATE-VERB-001 — Acceptance Criteria

> Verification layer. Given-When-Only binary-testable scenarios. Surface references are decision-gate-agnostic: "the creation verb surface" = whichever of (가) `moai worktree new` / (나) `moai codex -w --create` the plan.md gate selects. Each AC names the REQ it verifies.

## §D AC Matrix

| AC ID | REQ | Scenario |
|-------|----------|----------|
| AC-WCV-001 | REQ-WCV-001, REQ-WCV-002 | Verb creates tree at conventional path via existing plumbing |
| AC-WCV-002 | REQ-WCV-003, REQ-WCV-008 | Refusal without `<name>` value |
| AC-WCV-003 | REQ-WCV-004, REQ-WCV-009 | Collision refusal (including plain non-worktree directory destination) |
| AC-WCV-004 | REQ-WCV-005 | Session-worktree default-OFF baseline intact |
| AC-WCV-005 | REQ-WCV-006 | Static guard test passes |
| AC-WCV-006 | REQ-WCV-004, REQ-WCV-008 (diagnostics shape + English) | Existing resolve diagnostics unchanged |
| AC-WCV-007 | REQ-WCV-007 | L1 placement respected |
| AC-WCV-008 | REQ-WCV-009 | Path-escape (`..`, separator tricks, out-of-prefix L2) and plain-directory destination refusal |

## §D.1 AC-WCV-001 — Creation at the conventional path

- **Given** a scratch git repository under `t.TempDir()` with moai config present
- **When** the creation verb surface is invoked with `<name>` = `probe-card` (a name with no existing tree)
- **Then** the command exits 0; a worktree exists at `.claude/worktrees/probe-card`; `git worktree list` contains that path; the branch is newly created; the base branch matches `LoadWorktreeBaseBranch` resolution; and the repository contains exactly ONE new `git worktree add` invocation path (the existing plumbing — verified by the test asserting the plumbing function was the executor, e.g. via the shared entry point, not a duplicated exec).

## §D.2 AC-WCV-002 — Refusal without a value

- **Given** the same scratch repository
- **When** the creation verb surface is invoked with no `<name>` value
- **Then** the command exits non-zero; stderr carries a structured English error naming the expected argument; no directory under `.claude/worktrees/` is created; `git worktree list` is unchanged; and no interactive prompt is emitted (process exits without waiting on stdin).

## §D.3 AC-WCV-003 — Collision refusal

- **Given** a scratch repository with an existing worktree at `.claude/worktrees/taken`
- **When** the creation verb surface is invoked with `<name>` = `taken`
- **Then** the command exits non-zero; stderr carries a diagnostic consistent with the existing resolve-error shape; the existing worktree is unmodified; and no new branch is created.

## §D.4 AC-WCV-004 — Session-worktree gate baseline intact

- **Given** the repository at run-phase HEAD with the new verb merged
- **When** the existing session-worktree test suite for `config.SessionWorktreeEnabled` and `enterSessionWorktree` consumers runs (affected packages scope)
- **Then** all pre-existing tests pass UNMODIFIED (byte-identical to the pre-verb baseline); the gate's default value is still OFF; and `moai init` / `moai web` / `moai profile` side-effect behavior is unchanged (existing tests assert this without edits).

## §D.5 AC-WCV-005 — Static guard test

- **Given** the new verb surface's command registration
- **When** the static guard test (the `TestNew_NoAskUserQuestion` pattern) is run against the new surface
- **Then** it passes, asserting no AskUserQuestion / interactive-prompt dependency in the new command code path.

## §D.6 AC-WCV-006 — Existing resolve diagnostics unchanged

- **Given** the pre-verb baseline behavior of the resolve path (`moai codex -w <existing>` resolves; `moai codex -w <missing>` errors with the existing resolve-only message)
- **When** the existing codex launcher tests run after the verb lands
- **Then** the pre-existing resolve-path tests pass UNMODIFIED for the non-creating forms; if option (나) was chosen, the NEW `--create` form's tests cover the create path without weakening the existing no-flag error contract.

## §D.7 AC-WCV-007 — L1 placement

- **Given** the creation verb surface invoked with `<name>` = `l1-check`
- **When** creation completes
- **Then** the tree path is exactly `.claude/worktrees/l1-check` (conventional L1); no absolute-path L2 form is created by the verb; and the L1/L2 boundary semantics from `worktree-integration.md` § Terminology Glossary are observably preserved (the created tree is indistinguishable in placement from an `moai cc -w`-created L1 tree, save for branch-naming policy which is NOT in scope).

## §D.8 AC-WCV-008 — Path-escape and plain-directory destination refusal

- **Given** a scratch git repository under `t.TempDir()` containing a plain (non-worktree) directory `.claude/worktrees/plain-dir`
- **When** the creation verb surface is invoked three times — with `<name>` = `../escape`; with an absolute L2 path outside the sanctioned prefix; and with `<name>` = `plain-dir`
- **Then** every invocation exits non-zero with a structured English error; no directory is created outside `.claude/worktrees/`; the existing plain directory is unmodified; and `git worktree list` is unchanged.

## Edge Cases

- `<name>` containing path separators or `..` → refused with a structured error (input validation at the trust boundary; never simplified away).
- Destination exists as a plain (non-worktree) directory → refused (collision, AC-WCV-003 shape).
- Branch name already taken but no tree → refused with a diagnostic naming the branch conflict.
- `<name>` colliding with the develop/integration worktree name → refused (same collision path; no special-casing).

## Quality Gate Criteria (TRUST 5)

- **Tested**: affected-package coverage ≥ 85% for the new verb file (project `test_coverage_target: 85`); all ACs test-shaped and runnable.
- **Readable / Unified**: `go vet` + `golangci-lint` clean on touched packages; gofmt clean.
- **Secured**: name validation at the CLI trust boundary (AC edge cases above).
- **Trackable**: Conventional Commit(s) referencing SPEC-WORKTREE-CREATE-VERB-001.

## Definition of Done

1. M1 decision gate resolved and recorded (surface = (가) or (나), with the live-Codex observation evidence).
2. All AC-WCV-001..008 pass with verbatim command output cited in progress.md §E.2.
3. REQ-WCV-001..009 each trace to at least one passing AC.
4. Out-of-scope boundaries held: no t1071/t1072/t1073 artifacts touched; no second creation path; gate default unchanged.

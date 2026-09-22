# SPEC-WORKTREE-DONE-TIER-001 — Implementation Plan

## §A Context

Card t1073. The `[HARD]` doctrine in `.claude/rules/moai/workflow/worktree-integration.md:52` claims `moai worktree` verbs are L2-only, but `internal/cli/worktree/done.go` removes any worktree whose branch matches — including cooled L1 trees under `<repo>/.claude/worktrees/`, which may hold the ONLY copy of unmerged card work. The chosen disposal is a code guard that makes the doctrine true for `done`, plus a minimal doctrine precision edit.

Key code facts (measured on this tree, 0314801c2):

- `runDone` (done.go:124, interactive) and `runDoneWorktreeCleanup` (done.go:51, `--auto` core) both resolve via `WorktreeProvider.List()` → `wt.Branch == branchName` → `WorktreeProvider.Remove(targetPath, force)`.
- No path predicate / tier constant / registry consult in `internal/cli/worktree` (10 non-test files). ERE re-measurement with positive controls in 9 external files — measured absence (see spec.md §A measurement-quality note).
- Anchor guard at done.go:77 (`session.LiveAnchoredSessions`) — silent for cooled trees.
- `resolveSpecBranch` (shared.go:32) maps `SPEC-<ID>` → `feature/SPEC-<ID>`; `gitRepoRootFunc` in shared.go is overridable in tests.
- `DeleteBranch` uses safe `-d` (worktree.go:207); git dirty refusal maps to `ErrWorktreeDirty` (worktree.go:121-122).

## §B Known Issues

- The originating card's grep citation used an unfirable BRE pattern (literal pipes) — its 0-row result was a pattern artifact. Any run-phase grep evidence MUST reuse the valid ERE alternation shape WITH a positive control (memory lesson t1039: a firing positive control still misses a blind pattern only if the pattern shape itself is wrong — here the shape was corrected and the control fires).
- Loss shape C (ignored-only dirtiness removed WITHOUT `--force`) is the one unconditional-loss shape today; git's own refusal does not see ignored files.
- `worktree-integration.md:52`'s "clean ... cannot act on it" is contradicted by lines 64-67 of the same file (designed WT- merged sweep) — a pre-existing internal tension this SPEC resolves wording-wise, not behavior-wise.

## §C Pre-flight

- [ ] Worktree anchored: `.claude/worktrees/t1073`, branch `WT-done-tier-claim`, base = develop `0314801c2`. Re-read HEAD immediately before any commit.
- [ ] Confirm `internal/cli/worktree/done.go` unchanged since 0314801c2 for the cited line anchors (re-grep before editing — line numbers drift).
- [ ] Read the L2 `done SPEC-{ID} --auto --delete-branch` contract the guard must not break: `.claude/skills/moai/workflows/sync/delivery.md:383` (template mirror `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md:358`).
- [ ] Target test file absent: `internal/cli/worktree/done_l1_tier_guard_test.go` (create new).

## §D Constraints

- Do NOT touch `clean.go` / `remove.go` / `recover.go` / anchor logic (REQ-008).
- Guard applies to BOTH `runDone` and `runDoneWorktreeCleanup`, refusal BEFORE any `Remove` call; NOT bypassable by `--force` (REQ-002).
- Path comparison canonicalized (`filepath.EvalSymlinks` best-effort + fallback); the main repo root is resolved CWD-INDEPENDENTLY from the TARGET path — `git -C <targetPath> rev-parse --path-format=absolute --git-common-dir` (git >= 2.31), parent of the common dir; fallback first `worktree` stanza of `git -C <targetPath> worktree list --porcelain` for older git. The resolver is an overridable test seam whose DEFAULT is this target-derived mechanism (a CWD-anchored prefix fails open inside linked worktrees — audit D1).
- Tests: real git repos in `t.TempDir()`, never the project tree; `filepath.Abs` for user-path resolution (macOS `/var/folders` lesson).
- No config flag / opt-out on the guard.
- Do not commit (orchestrator commits after audit); do not push; do not modify the todo queue.
- Template-First: doctrine edit lands in BOTH `.claude/rules/moai/workflow/worktree-integration.md` AND `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`, byte-identical; no `make build` needed for rule files (non-embedded at authoring time — verify: run `make agents-emit-check`-style drift check only if the file proves to be inside an emit scope; rules files are not).

## §E Self-Verification Plan

- E1 AC matrix: run `go test ./internal/cli/worktree/... -run 'TestDoneL1TierGuard' -v` — all RED-now cells green post-M2; quote verbatim output.
- E2 Regression: `go test ./internal/cli/worktree/...` full package — no regression; `go vet ./internal/cli/...`; `golangci-lint run internal/cli/worktree/...`.
- E3 Cross-platform: `GOOS=windows GOARCH=amd64 go build ./...` and `GOOS=darwin` (path predicate must be `filepath`-based, no hardcoded `/`).
- E4 Doctrine parity: `cmp` local rule vs template mirror — byte-identical (after the edit).
- E5 Scope: `git diff --name-only <base>..HEAD -- internal/ .claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` — touches exactly: `done.go` (+ the predicate/resolver site), `done_l1_tier_guard_test.go`, the two `worktree-integration.md` copies. The SPEC artifacts under `.moai/specs/SPEC-WORKTREE-DONE-TIER-001/` and progress.md are expected plan/run artifacts, not scope violations (AC-009).

## §F Milestones (priority-ordered, no time estimates)

### M1 — RED-now reproduction tests (Priority High)

Create `internal/cli/worktree/done_l1_tier_guard_test.go` pinning loss-shapes A1, A2, A3, C (spec.md §D) as FAILING tests against the unguarded code: real git repos in `t.TempDir()` (main repo + a worktree created under a `.claude/worktrees/` subdir of the temp repo root, with `gitRepoRootFunc` redirected), each test asserts the POST-guard behavior so it fails RED now. Record the RED run: pre-fix tree SHA, verbatim `go test` command, failing output, exit code. These tests ARE the two-cell RED side of AC-001..AC-004/AC-005.

- M1a: A1 (clean L1 removed silently today) — test expects refusal.
- M1b: A3 (`--force` destroys dirty L1 today) — test expects refusal WITH `--force` set.
- M1c: C (ignored-only dirtiness lost today) — test expects refusal.
- M1d: A2 (dirty refusal message provenance check) + A4 (`--delete-branch` pre-removal refusal).

### M2 — Tier guard implementation (Priority High) — the highest-change-likelihood decision; review first

- M2a: Predicate helper — `isL1WorktreePath(path string) bool` (or equivalent): resolve the MAIN repo root FROM THE TARGET PATH (`git -C <targetPath> rev-parse --path-format=absolute --git-common-dir`, parent of common dir; fallback: first stanza of `git -C <targetPath> worktree list --porcelain` — NEVER process CWD, which inside a linked worktree returns the worktree's own path and makes the guard fail open), canonicalize target + `<mainRoot>/.claude/worktrees/` via best-effort `filepath.EvalSymlinks`, `strings.HasPrefix` compare with separator-safe boundary. The resolver is an overridable function whose default is the target-derived mechanism.
- M2b: Wire the refusal into `runDoneWorktreeCleanup` (before the anchor guard, immediately after target resolution) and into `runDone`'s flow — ONE shared refusal site so the two paths cannot diverge; refusal short-circuits BEFORE `WorktreeProvider.Remove` regardless of `force`.
- M2c: Refusal message per REQ-003 (path + session-scoped statement + the two remedy forms), stderr, wrapped error so both modes exit non-zero.
- M2d: `@MX:ANCHOR` (+ `@MX:SPEC: SPEC-WORKTREE-DONE-TIER-001`) on the guard function; `@MX:NOTE` on the symlink-fallback branch. English comments.
- M2e: B-direction (no-mutation) tests: L2-shaped tree outside `.claude/worktrees/` + `done SPEC-<ID> --auto --delete-branch` passes unchanged (AC-006); plus a same-shape positive control that the predicate fires for an L1 path — both directions required (memory lesson t1028: two mutant directions catch refusal and no-op separately).
- M2f: CWD-independence mutation test (AC-010, audit D1): a SECOND linked worktree of the temp repo, process CWD set inside it; `done` must STILL refuse the L1 target. A CWD-anchored prefix (`<worktree>/.claude/worktrees/`) fails exactly this cell while passing AC-001 — it is the fail-open direction the M1 RED run alone cannot catch.

### M3 — Doctrine precision edit (Priority Medium)

- M3a: Edit `worktree-integration.md:52` — bounded (a few lines): `done` refuses L1 by code (tier guard), `clean --merged-only` WT- sweep is the documented exception, `recover` unchanged.
- M3b: Apply byte-identical wording to `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`; `cmp` both.
- M3c: Read `kanban-dispatch.md` "moai worktree done closes L2 trees only" claims — post-guard they are true; edit only if factually wrong.

### M4 — Verification and gates (Priority Medium)

- M4a: §E self-verification batch (E1-E5) executed and quoted verbatim into progress.md §E.2.
- M4b: Coverage of the touched package meets the project gate; characterization intact for the unchanged L2 flow.

## §G Anti-Patterns

- Do NOT implement the guard as a filter on `WorktreeProvider.List()` — that would change every verb's view; scope it to `done` call sites.
- Do NOT let `--force` reach the guard as a bypass parameter (A3 is the point of the SPEC).
- Do NOT run the reproduction tests against the project's own `.claude/worktrees/` — temp repos only.
- Do NOT broaden the doctrine edit beyond line 52's sentence (clean/remove/recover observations stay observations).
- Do NOT add an env/config opt-out to the guard.

## §H Cross-References

- spec.md §D loss-shape matrix; acceptance.md AC-001..AC-009.
- `.claude/rules/moai/workflow/worktree-integration.md` § Terminology Glossary (L1/L2) + line 52 doctrine.
- `.claude/skills/moai/workflows/sync/delivery.md:383` (template mirror `:358`) — deployed L2 `done --auto --delete-branch` consumer (must not regress).
- `internal/cli/worktree/done_test.go`, `done_anchor_test.go` — existing test patterns.
- Memory lessons: t1039 (pattern-shape positive control), t1028 (two mutant directions), macOS `/tmp` symlink canonicalization.

# SPEC-WORKTREE-DONE-TIER-001 — Acceptance Criteria

Two-cell adoption (per `.claude/rules/moai/development/verification-completeness.md`): every criterion carries a **RED-now cell** (the pre-fix failure, pinned at run-phase start to the pre-fix tree SHA + verbatim command + output + exit code) and a **green-path cell** naming the milestone that flips it.

RED pinning executes BEFORE M2 lands (REQ-006). All git-semantics cells use real git repos in `t.TempDir()` — never the project tree. Test file: `internal/cli/worktree/done_l1_tier_guard_test.go`.

## §D AC Matrix

| AC | Scenario (spec.md §D) | RED-now observes | Green milestone |
|----|----------------------|------------------|-----------------|
| AC-001 | A1 clean L1, no flags | Tree removed silently, exit 0 | M2 |
| AC-002 | A2 dirty L1, no flags | Refused by git (`ErrWorktreeDirty`), not by tier | M2 |
| AC-003 | A3 dirty L1, `--force` | Removed; uncommitted work destroyed | M2 |
| AC-004 | C ignored-only dirty L1, no flags | Removed; ignored files lost | M2 |
| AC-005 | A4 L1 + `--delete-branch`, unmerged | Tree removed; branch survives | M2 |
| AC-006 | B L2 shape, `--auto --delete-branch` | PASS (deployed flow) — must stay PASS | M2 (no-mutation) |
| AC-007 | Refusal message + exit codes | n/a (new behavior; verified green-side + RED-side absence) | M2 |
| AC-008 | Doctrine precision edit parity | Line 52 tension present | M3 |
| AC-009 | Scope confinement | n/a | M4 |
| AC-010 | CWD inside a second linked worktree (D1 mutation) | Tree removed silently, exit 0 (no guard; and a CWD-anchored predicate would fail open here) | M2f |

## §D.1 AC-001 — cooled clean L1 tree is refused (A1)

**Given** a real temp git repo with a worktree created under `<repoRoot>/.claude/worktrees/<name>` whose session has exited (no anchor registry entry, no lock), tree clean
**When** `moai worktree done <branch>` runs (no flags)
**Then (RED-now)** the tree is removed with exit 0 and no refusal — pinned verbatim at run-phase start
**Then (green, M2)** the command exits non-zero, stderr names the path + states L1 trees are session-scoped + gives the two remedy forms, and the tree still exists on disk.

## §D.2 AC-002 — cooled dirty L1 refused, provenance improves (A2)

**Given** a cooled L1 tree with uncommitted modifications
**When** `moai worktree done <branch>` runs (no flags)
**Then (RED-now)** refusal comes from git's dirty check (`ErrWorktreeDirty`), not any tier predicate — pinned verbatim
**Then (green, M2)** refusal comes from the tier guard (message may differ from git's; improvement allowed); tree survives in both cells.

## §D.3 AC-003 — --force does NOT bypass the L1 refusal (A3)

**Given** a cooled L1 tree with uncommitted modifications
**When** `moai worktree done <branch> --force` runs
**Then (RED-now)** the tree is removed and the uncommitted work is destroyed — the loss this SPEC exists to prevent; pinned verbatim at run-phase start
**Then (green, M2)** the tier refusal fires and is NOT bypassed by `--force`; exit non-zero; uncommitted work intact; the L1 remedy (`git worktree unlock <path>` + `git worktree remove <path>`) remains available in the message.

## §D.4 AC-004 — ignored-only dirtiness is not silently lost (C)

**Given** a cooled L1 tree whose only dirtiness is an IGNORED file (e.g. an entry in the temp repo's `.git/info/exclude`, file present on disk)
**When** `moai worktree done <branch>` runs (NO `--force`)
**Then (RED-now)** the tree is removed WITHOUT `--force` and the ignored file is silently lost — the one unconditional-loss shape; pinned verbatim
**Then (green, M2)** refused by the tier guard; ignored file intact.

## §D.5 AC-005 — --delete-branch refuses pre-removal on L1 (A4)

**Given** a cooled L1 tree on an unmerged branch
**When** `moai worktree done <branch> --delete-branch` runs
**Then (RED-now)** the tree is removed; the branch survives (safe `-d` refuses unmerged) — a half-completed disposal; pinned verbatim
**Then (green, M2)** refused pre-removal; tree AND branch untouched.

## §D.6 AC-006 — L2 flow unchanged (B — no-mutation direction)

**Given** a worktree OUTSIDE `.claude/worktrees/` (L2 shape) on branch `feature/SPEC-<ID>`, merged and clean
**When** `moai worktree done SPEC-<ID> --auto --delete-branch` runs (the deployed sync flow — delivery.md local copy line 383, template mirror line 358)
**Then (RED-now)** the tree and branch are removed, exit 0 — the deployed legit behavior
**Then (green, M2)** IDENTICAL behavior: removal succeeds, exit 0, no extra stderr output. Both flags exercised. Companion positive control: the same predicate must fire for an L1-shaped path (two mutant directions — refusal and no-op — caught separately).

## §D.7 AC-007 — refusal surfaces identically in both modes

**Given** a cooled L1 tree
**When** `done` runs in interactive mode AND in `--auto` mode (separate scenarios)
**Then (green, M2)** BOTH exit non-zero; stderr message in both contains: the target path, the session-scoped statement, and the two remedy forms (session-end keep/remove prompt; `git worktree unlock <path>` + `git worktree remove <path>`) — style-consistent with `lockGuidance` (done.go:116-122). RED-side: the refusal sites do not exist yet (verified by the M1 RED runs of AC-001/AC-003 succeeding silently).

## §D.8 AC-008 — doctrine precision edit, both copies byte-identical

**Given** the guard landed (M2)
**When** reading `.claude/rules/moai/workflow/worktree-integration.md:52` and `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`
**Then (RED-now)** the sentence claims `done`/`clean`/`recover` cannot act on L1 — with `clean`'s WT- sweep (lines 64-67) contradicting it in the same file
**Then (green, M3)** the sentence states: `done` refuses L1 trees by code (tier guard), `clean --merged-only`'s WT- sweep is the documented exception, `recover` unchanged; `cmp` proves the two copies byte-identical; `kanban-dispatch.md`'s "closes L2 trees only" claims are factually true post-guard (edited only if found wrong).

## §D.9 AC-009 — scope confinement

**Given** the completed run phase
**When** `git diff --name-only <base>..HEAD -- internal/ .claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` runs (explicit pathspecs — the UNSCOPED base..HEAD diff necessarily also contains the SPEC artifacts under `.moai/specs/SPEC-WORKTREE-DONE-TIER-001/` and the progress.md updates, which are expected plan/run artifacts, not scope violations)
**Then** the pathspec-scoped touched set is exactly: `internal/cli/worktree/done.go` (+ at most the shared predicate/resolver site), `internal/cli/worktree/done_l1_tier_guard_test.go`, `.claude/rules/moai/workflow/worktree-integration.md`, `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`. No changes to `clean.go`, `remove.go`, `recover.go`, anchor logic, or `WorktreeProvider.List()` filtering.

## §D.10 Edge cases covered inside AC-001..AC-006, AC-010

- macOS `/tmp` → `/private/tmp` symlink: predicate canonicalizes best-effort with fallback (AC-001 temp layout exercises it via `t.TempDir()`).
- Branch-name resolution: `SPEC-<ID>` → `feature/SPEC-<ID>` via `resolveSpecBranch` exercised in AC-006.
- Anchor-registered LIVE L1 tree: refusal fires for the tier reason regardless (guard ordered BEFORE the anchor check; both refusals are safe outcomes — asserted in M2 tests).

## §D.11 AC-010 — repo-root resolution is CWD-independent (audit iter-1 D1 mutation cell)

**Given** a real temp git repo with (a) an L1 worktree under `<mainRoot>/.claude/worktrees/<name>` (clean, cooled) AND (b) a SECOND LINKED worktree of the same repo created elsewhere under the temp dir, with the PROCESS CWD set inside the second linked worktree when `done` executes
**When** `moai worktree done <branch>` targets the L1 tree
**Then (RED-now)** the tree is removed silently with exit 0 — and, specifically, a predicate anchored on `git rev-parse --show-toplevel` FROM PROCESS CWD would compute `<linkedWorktree>/.claude/worktrees/` (never matching) and fail OPEN here while still passing AC-001; this cell is the fail-open direction the AC-001 RED run cannot distinguish; pinned verbatim at run-phase start
**Then (green, M2f)** refused with the REQ-003 message; the main root was derived from the TARGET path (`git -C <targetPath> rev-parse --path-format=absolute --git-common-dir`, git >= 2.31; fallback: first `worktree` stanza of `git -C <targetPath> worktree list --porcelain`), never from process CWD; tree survives.

## §E Quality Gate Criteria

- `go vet ./internal/cli/...` clean; `golangci-lint run internal/cli/worktree/...` clean.
- `go test ./internal/cli/worktree/...` full package pass (affected-package only; full-suite verdict = CI).
- Cross-platform build: `GOOS=windows` and `GOOS=darwin` `go build ./...` pass (predicate is `filepath`-based).
- Definition of Done: AC-001..AC-010 green with verbatim evidence in progress.md §E.2; RED-now evidence recorded for AC-001..AC-005 and AC-010 before M2.

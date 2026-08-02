# Acceptance — SPEC-SESSION-WORKTREE-001

> Given-When-Then AC enumeration. Each AC is binary-testable. The GEARS obligation lives in spec.md §2; this file is the verification layer. v0.2.1 iter-2 audit-fix: AC-SW-009/019/021/022 rewritten (D1 gitconfig source, D4 on-touch trigger, D5 path-distinguishing notices, Q4 hooksPath removed).

## §A — Severity legend

- **MUST** — failure blocks merge (corresponds to REQ-SW-*** in spec.md).
- **SHOULD** — failure is a tracked debt; merge allowed with recorded justification.
- **NICE** — quality bar; not merge-blocking.

All ACs below are MUST unless tagged otherwise.

## §B — Traceability to REQs

See spec.md §7 Acceptance Criteria Matrix for the REQ → AC mapping. Each AC here carries an `AC-SW-NNN` ID traceable to one or more REQs.

## §C — Acceptance Criteria (Given-When-Then)

### AC-SW-001 — Default-off: `moai web` unset behaves identically (MUST, REQ-SW-001)

**Given** a clean project with no `workflow.session_worktree` config key and no `MOAI_SESSION_WORKTREE` env var,
**When** the user invokes `moai web`,
**Then** the invocation runs in the shared primary checkout (no worktree materialized), `emitWorktreeAdvisory` fires exactly as today, and `.claude/settings.local.json` writes land at the shared path.

### AC-SW-002 — Default-off: `moai init` unset behaves identically (MUST, REQ-SW-001)

**Given** a clean project with the feature unset,
**When** the user invokes `moai init`,
**Then** the init runs against the shared primary checkout, the shared `.moai/config/sections/*.yaml` is regenerated, and no `[WT]-*` branch is created.

### AC-SW-003 — Config flag ON materializes worktree for `moai web` (MUST, REQ-SW-002)

**Given** `workflow.session_worktree.enabled: true`,
**When** the user invokes `moai web`,
**Then** a worktree is materialized under the project, the branch name starts with `[WT]`, the web console runs with the worktree's `.moai/state/` and `.claude/settings.local.json`, and the shared primary checkout's equivalents are untouched by this invocation.

### AC-SW-004a — Env override `=1` forces ON regardless of config (MUST, REQ-SW-003)

**Given** `workflow.session_worktree.enabled: false`,
**When** the user invokes `MOAI_SESSION_WORKTREE=1 moai web`,
**Then** the feature is ON (worktree materialized).

### AC-SW-004b — Env override `=0` forces OFF regardless of config (MUST, REQ-SW-003)

**Given** `workflow.session_worktree.enabled: true`,
**When** the user invokes `MOAI_SESSION_WORKTREE=0 moai web`,
**Then** the feature is OFF (no worktree materialized; behavior byte-identical to feature-off).

### AC-SW-005 — Non-git directory falls back with notice (MUST, REQ-SW-004)

**Given** the invocation is in a directory that is not a git repository (or `git worktree add` fails),
**When** the user invokes `moai init` with the feature ON,
**Then** the invocation does NOT abort; it falls back to the shared-checkout behavior, a stderr notice names the failure reason, and the exit code is whatever the init's normal exit code would have been.

### AC-SW-006 — Scoped surface: `moai cc` does NOT auto-enter (MUST, REQ-SW-005)

**Given** `workflow.session_worktree.enabled: true`,
**When** the user invokes `moai cc` (or `moai cg` / `moai glm` / `moai update`),
**Then** no session-worktree is materialized by this SPEC. (Existing `-w` plumbing still applies if the user passes `-w` explicitly.)

### AC-SW-007 — Branch name shape `[WT]-<8hex>-<subcommand>` (MUST, REQ-SW-006 / REQ-SW-007)

**Given** the feature ON and a materializable git repository,
**When** the user invokes `moai web` from a session whose UUID starts `abcdef12-...`,
**Then** the created branch name matches `\[WT\]-abcdef12-web` (regex-anchored). Outside any session, the 8-hex segment is a 6-byte random hex (12 hex chars) — still prefixed `[WT]-` and suffixed `-web`.

### AC-SW-008 — Default-manual disposal: worktree persists (MUST, REQ-SW-008)

**Given** `workflow.worktree.auto_cleanup: false` (the distributed default),
**When** `moai web` (feature ON) exits cleanly,
**Then** the worktree remains on disk after exit; `git worktree list` still shows it; the user must explicitly `moai worktree done` / `remove` to dispose.

### AC-SW-009 — Opt-in auto-cleanup on clean exit emits distinguishable notice (MUST, REQ-SW-009)

**Given** `workflow.worktree.auto_cleanup: true` AND `workflow.session_worktree.enabled: true`,
**When** `moai web` exits 0,
**Then** the worktree is removed automatically (no longer in `git worktree list`) AND a stderr notice `removed by session-exit cleanup: [WT] ...` is emitted, distinguishable from the PR-merge cleanup notice (AC-SW-022 / REQ-SW-022).

### AC-SW-010 — Dirty worktree preserved on session-exit auto-cleanup (MUST, REQ-SW-010)

**Given** auto-cleanup is ON,
**When** `moai web` exits 0 but the worktree has uncommitted changes (`git status --porcelain` non-empty),
**Then** the worktree is NOT removed; a stderr notice names the worktree path.

### AC-SW-011 — Coexistence: manual `-w` + `moai web` does not nest (MUST, REQ-SW-011)

**Given** the user has manually entered a worktree via `moai cc -w foo`,
**When** the user invokes `moai web` (feature ON),
**Then** no second (nested) session-worktree is materialized.

### AC-SW-012 — Already-in-worktree skip with notice (MUST, REQ-SW-012)

**Given** `git rev-parse --git-dir != git rev-parse --git-common-dir`,
**When** the user invokes `moai init` / `moai profile` / `moai web` (feature ON),
**Then** the auto-entry is skipped and an info notice reports the invocation is already worktree-isolated.

### AC-SW-013 — Advisory suppressed when worktree materialized (MUST, REQ-SW-013)

**Given** the feature ON and a successful worktree materialization,
**When** `moai web` proceeds to the `emitWorktreeAdvisory` call site,
**Then** the advisory is NOT printed.
**Negative control:** when the feature is OFF OR the materialization fell back (REQ-SW-004), the advisory fires unchanged.

### AC-SW-014 — Parallel isolation: two concurrent webs land in distinct worktrees (MUST, REQ-SW-014)

**Given** the feature ON,
**When** two `moai web` invocations start within the same second from two different sessions,
**Then** the two invocations land in two different worktrees with two different `[WT]-<session-short>-web` branch names, two different `.moai/state/` trees, and two different `.claude/settings.local.json` write surfaces.

### AC-SW-015 — Port collision still reported under isolation (MUST, REQ-SW-015)

**Given** the feature ON,
**When** two parallel `moai web --port 3041` invocations both target port 3041,
**Then** the second to start reports the port conflict via `ensurePortFree` exactly as today.

---

### AC-SW-016 — `moai profile` auto-enters a worktree (MUST, REQ-SW-016)

**Given** `workflow.session_worktree.enabled: true`,
**When** the user invokes `moai profile` against a project,
**Then** a worktree is materialized for the project, the branch name matches `\[WT\]-<8hex>-profile`, and the profile subcommand's logic runs against the worktree's project tree. (The profile dir at `~/.moai/claude-profiles/` is NOT isolated — see AC-SW-017.)

### AC-SW-017 — Profile auto-entry scopes to project, NOT profile dir (MUST, REQ-SW-017)

**Given** the feature ON and `moai profile` invoked against project P,
**When** the auto-entry materializes a worktree,
**Then** (a) the worktree path is under project P (NOT under `~/.moai/claude-profiles/`); (b) the profile dir `~/.moai/claude-profiles/<name>/launch.yaml` is written by this invocation at its GLOBAL path (NOT redirected to the worktree); (c) a stderr notice during auto-entry states explicitly that the profile dir is NOT isolated by this entry; (d) two parallel `moai profile` invocations on the same profile name produce two distinct project worktrees BUT both still write to the same global `launch.yaml` (the ledger race is documented as out of scope).

### AC-SW-018 — `safe.directory` registered on materialization (MUST, REQ-SW-018)

**Given** the feature ON and a worktree materialized at path W,
**When** materialization completes,
**Then** `git config --global --get-all safe.directory` includes W (add-only entry), and a subsequent `git -C W status` exits 0 (does NOT fail with "detected dubious ownership"). Re-materializing W is idempotent (the entry appears once, not N times).

### AC-SW-019 — Worktree commits under user's global user.name / user.email read from ~/.gitconfig (MUST, REQ-SW-019)

**Given** the feature ON AND the user's global git config has `user.name = "Alice"` and `user.email = "alice@example.com"` (verified via `git config --global --get user.name` / `user.email`),
**When** auto-entry materializes a worktree and a commit is made inside it,
**Then** `git -C <worktree> config user.name` returns `Alice`, `git -C <worktree> config user.email` returns `alice@example.com`, and `git -C <worktree> log -1 --format='%an <%ae>'` returns `Alice <alice@example.com>`. The global `~/.gitconfig` `user.name` / `user.email` are UNCHANGED (the identity was applied worktree-scoped, not globally). **No-global-identity carve-out:** when the user has NO global `user.email` / `user.name` set (`git config --global --get user.email` returns empty), no worktree-scoped identity is applied and git inherits unchanged (the requirement is a verified no-op, not a forced empty value). **Source invariant:** the profile/config carry only `Name` (no `Email`) — verified `pkg/models/config.go` L32-37 + `internal/profile/preferences.go` L14-16 — so the profile is NOT consulted for identity; the global gitconfig is the sole source.

### AC-SW-020 — `moai init` new repo's first branch is `main` (MUST, REQ-SW-020)

**Given** the feature ON and `moai init` creating a new project inside a worktree,
**When** `moai init` creates the new git repository,
**Then** `git -C <new-project> symbolic-ref HEAD` returns `refs/heads/main` (NOT `refs/heads/master`), and the first commit lands on `main`.

### AC-SW-021 — Opt-in git config options applied when profile carries them (SHALL, REQ-SW-021)

**Given** the feature ON and the active profile opts in via `signingkey = "ABC123"`,
**When** the worktree is materialized,
**Then** `git -C <worktree> config --get commit.gpgsign` returns `true` and `user.signingkey` returns `ABC123` (worktree-scoped). **hooksPath removal (v0.2.1):** `core.hooksPath` is NOT set on the worktree — the worktree inherits the project's `.claude/hooks/` unchanged (REQ-SW-021 dropped `core.hooksPath` at v0.2.1; see spec.md §3 Out of Scope — Hook isolation). **Negative control:** when the profile carries NONE of the opt-in fields, materialization completes without error and none of these keys are set on the worktree config (silent skip). For new repos created by `moai init`, `core.autocrlf` = `input`, `fetch.prune` = `true`, `push.default` = `current` (applied on the new repo, NOT on pre-existing user repos).

### AC-SW-022 — On-touch PR-merge cleanup at `moai session register` / `moai session list` emits distinguishable notice (MUST, REQ-SW-022)

**Given** `workflow.worktree.auto_cleanup: true` AND a `[WT]` worktree whose branch's PR is in MERGED state (`gh pr view <branch> --json state` returns `state: "MERGED"`),
**When** the user runs `moai session register` (or `moai session list`),
**Then** that worktree + branch are removed AND a stderr notice `removed by PR-merge cleanup: [WT] worktree <name> (branch merged)` is emitted, distinguishable from the session-exit cleanup notice (AC-SW-009). **Negative control:** given `workflow.worktree.auto_cleanup: false` (default) AND the same merged branch state AND the same on-touch invocation, the worktree is NOT removed by this path (only session-exit cleanup behavior — also off — applies). **Trigger invariant:** the PR-merge cleanup fires ONLY at the two named on-touch invocations (`moai session register` and `moai session list` — both verified-existing CLI commands per `internal/cli/session.go:66` register and `:132` list; corrected at v0.2.2 from the v0.2.1 names `moai worktree list` / `moai session start` which were non-existent — `worktree list` is intentionally retired per `internal/cli/worktree/root_test.go:61`, `session start` never existed per `internal/cli/session.go:53-59`) — it does NOT fire at `moai cc`, `moai cg`, `moai glm`, `moai init`, `moai profile`, `moai web`, or any other subcommand.

### AC-SW-023 — PR-merge detection: gh primary, git branch --merged fallback (MUST, REQ-SW-023)

**Given** a `[WT]` worktree whose branch B was merged to base via PR #N,
**When** the PR-merge detection runs,
**Then** (a) if `gh` is available: `gh pr view B --json state` returns `state: "MERGED"` and the worktree is a cleanup candidate; (b) if `gh` is NOT available: `git branch --merged origin/main` is consulted and B is a cleanup candidate iff listed. **Squash-merge blindness (documented):** given `gh` is NOT available AND PR #N was squash-merged, `git branch --merged origin/main` does NOT list B (squash-merged branches look unmerged), so the worktree is NOT removed and a stderr notice documents the fallback's blindness. **Primary path correctness:** the same squash-merged PR with `gh` available IS detected (state == MERGED) and the worktree IS removed.

### AC-SW-024 — Dirty worktree preserved on PR-merge cleanup (MUST, REQ-SW-024)

**Given** `workflow.worktree.auto_cleanup: true` AND a `[WT]` worktree whose branch is merged BUT whose `git status --porcelain` is non-empty,
**When** the PR-merge cleanup path evaluates this worktree,
**Then** the worktree is NOT removed; a stderr notice names the worktree path and tells the user to dispose manually (carries REQ-SW-010's safety to the PR-merge path).

## §D — Edge Cases

- **EC-1 — `moai init` in a brand-new directory that is not yet git-initialized.** REQ-SW-004 fail-back applies.
- **EC-2 — `git worktree add` fails mid-way (e.g., disk full).** Fail-back applies; partial worktree cleaned by `git worktree prune`.
- **EC-3 — `[WT]` branch name rejected by `git check-ref-format`.** Q2 RESOLVED as M1 deferral: M1 runs `git check-ref-format --branch '[WT]-x'`; if rejected, fall back to `WT-` prefix (no brackets) and document in `progress.md §E.2`.
- **EC-4 — Session UUID not resolvable.** REQ-SW-007's fallback: 6-byte random hex.
- **EC-5 — Feature ON + user passes `--port` + port free.** Worktree materialized + console binds the user-supplied port inside the worktree.
- **EC-6 — Auto-cleanup ON but worktree disposed by external `git worktree remove` during the run.** Cleanup at exit observes worktree gone; no-op + info notice.
- **EC-7 — `moai profile` invoked with no project context** (e.g., `moai profile list` — a read-only subcommand). The auto-entry MUST NOT fire for read-only profile operations; only profile-mutating invocations trigger REQ-SW-016. M6 must enumerate which `moai profile` subverbs mutate state and gate accordingly.
- **EC-8 — Profile carries `UserName` but the user's global gitconfig has no `user.email` (or vice versa: name only, no email).** REQ-SW-019 sources BOTH fields from `~/.gitconfig`, NOT from the profile — so a profile missing `UserName` is irrelevant. When the global gitconfig carries `user.name` but no `user.email` (or vice versa), REQ-SW-019 applies whichever field is non-empty and leaves the empty one to git's default resolution; a stderr notice names the partial identity actually applied. When BOTH are empty, the requirement is a verified no-op.
- **EC-9 — `git config --worktree` unsupported (git < 2.20).** REQ-SW-019 / REQ-SW-021 fall back to writing the worktree's local `.git/config` file directly. A stderr notice names the git-version fallback.
- **EC-10 — `safe.directory` entry orphaned** (worktree removed by external means, no cleanup ran). Subsequent `moai` invocations do NOT fail; the orphaned entry is harmless (add-only path allowlist). A periodic `git config --global --get-all safe.directory` audit (out of scope for this SPEC) could sweep orphans.
- **EC-11 — PR-merge cleanup race: branch merges BETWEEN the gh check and the removal.** The dirty-guard re-check immediately before removal catches any change; if the worktree went dirty in the window, preserve + notice.
- **EC-12 — Two profiles active in two parallel worktrees commit under divergent identities.** R7: documented behavior, not a defect. Each worktree's `git log` reflects the user's global gitconfig identity (REQ-SW-019 sources from `~/.gitconfig`, NOT the profile) — since both worktrees read the SAME global gitconfig, they commit under the SAME identity. Profile-keyed identity divergence is OUT OF SCOPE (spec.md §3 Out of Scope — Profile-vs-profile / per-project git identity).
- **EC-13 — PR-merge cleanup and session-exit cleanup fire in the same moai invocation.** The two stderr notices (`removed by PR-merge cleanup: ...` vs `removed by session-exit cleanup: ...`) are distinguishable in the combined output. The on-touch PR-merge path (AC-SW-022) at `moai session register` / `moai session list` fires BEFORE session-exit cleanup (AC-SW-009), which fires at subcommand exit; the two notices carry their distinct prefix so a user reading the log can attribute each removal.
- **EC-14 — User has no global git identity (`git config --global --get user.email` returns empty).** REQ-SW-019 is a verified no-op — no worktree-scoped identity is applied. A stderr notice names the no-op so the user knows their global gitconfig is unset; the worktree inherits git's default identity resolution (which may produce `user@hostname`-style identities on commit — that is git's behavior, not moai's).

## §E — Quality Gate Criteria

- **Test coverage:** ≥85% on the new code paths (M1 helper, M2/M3/M6 entry wrappers, M4/M8 naming/disposal/cleanup, M7 git-config helper).
- **Falsification round-trips:** the fail-back (REQ-SW-004), dirty-preserve (REQ-SW-010 / REQ-SW-024), already-in-worktree skip (REQ-SW-012), safe.directory idempotency (REQ-SW-018), no-global-identity no-op (REQ-SW-019), squash-merge fallback blindness (REQ-SW-023), and PR-merge-vs-session-exit notice distinguishability (REQ-SW-009 / REQ-SW-022) guards each observed FAILING when removed, then PASSING when restored.
- **Cross-platform build:** `go build ./...` AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **Lint:** `golangci-lint run ./internal/cli/... ./internal/config/... ./internal/profile/...` → 0 issues.
- **Vet:** `go vet ./...` exit 0.
- **Anti-regression:** feature OFF — byte-identical advisory output, byte-identical settings.local.json write path, byte-identical global `~/.gitconfig` (no `safe.directory` mutations, no identity mutations).
- **Template neutrality (§25):** `internal/template/templates/` diff under this SPEC is empty for the new config sub-key; no `[WT]` token, no SPEC-ID leaked.
- **Global-state minimality:** the ONLY global mutation this SPEC permits is `git config --global --add safe.directory <path>` (REQ-SW-018). A test asserts no other global mutation occurs (no `--global` flags in the new code outside the safe.directory call site).

## §F — Definition of Done

- All MUST ACs GREEN; AC-SW-021 is SHALL at v0.2.1 (no longer SHOULD) and must be GREEN or recorded as tracked debt with justification.
- Every guard has a falsification round-trip recorded in `progress.md` §E.2.
- §25 template neutrality verified (`git diff --name-only origin/main HEAD | grep '^internal/template/templates/'` → NONE attributable to this SPEC, modulo catalog regeneration).
- `moai spec lint` on this SPEC directory returns 0 errors.
- CHANGELOG entry added under `[Unreleased] → Added` naming SPEC-SESSION-WORKTREE-001.
- design.md and research.md (Tier L artifacts) are present, internally consistent with spec.md, and reference the same code paths the run-phase actually touched.

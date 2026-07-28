# plan.md — SPEC-WORKTREE-ENTRY-STRATEGY-001

> Implementation plan, milestones (priority-based, NO time estimates),
> technical approach, risks. Ordered by decision-reversibility — highest-
> change-likelihood decisions first (data-model / new type interfaces /
> UX flows), mechanical / refactoring steps deferred to the bottom.

## §A. Context

- **SPEC**: SPEC-WORKTREE-ENTRY-STRATEGY-001 (Tier L, era V3R6)
- **Branch**: main (clean, divergence `0 0` per origin/main fetched 2026-07-28)
- **Working tree**: main checkout at `/Users/goos/MoAI/moai-adk-go/`
- **Sibling**: SPEC-WORKTREE-BRANCH-GUARD-001 (`e89d01461`, merged) — primary-
  checkout branch-state guard; this SPEC is the entry-strategy sibling
- **Baseline evidence** (research.md §C): 58 worktrees total, 31 uncleaned
  `agent-*` L1 ephemeral; `defaults.go:521,523` carry `AutoCleanup: true` /
  `AutoMerge: true` (drift from user-intended OFF policy)
- **Plan-phase artifacts**: spec.md (9 REQs) + acceptance.md (9 ACs) +
  design.md (this file's sibling) + research.md + this plan.md + progress.md
- **plan-auditor verdict**: pending (first iteration)

## §B. Known Issues (auto-injected per manager-develop-prompt-template §B)

- **B1 (cross-platform build tags)**: not applicable — no `syscall` use; the
  Go changes are 2 default-value mutations in `defaults.go`.
- **B2 (cross-SPEC policy conflict)**: SPEC-WORKTREE-BRANCH-GUARD-001 is the
  sibling; auto-isolation (REQ-WES-005) MUST NOT trigger in the primary
  checkout. Verified compatible per research.md §E.3 (worktree paths are
  exempt from the branch-guard deny).
- **B3 (subagent boundary C-HRA-008)**: no `internal/hook/` or
  `internal/harness/` changes; the doc-rule changes are not subagent-domain.
- **B4 (frontmatter canonical schema)**: spec.md frontmatter carries all 12
  canonical fields; verified pre-write via the Bash regex self-check (PASS).
- **B5 (CI 3-tier)**: spec-lint + golangci-lint + go test; the only Go change
  is 2 default-value mutations, so the test impact is the existing
  `defaults_test.go` (or equivalent) — MUST be updated to assert the new
  defaults.
- **B6 (spec-lint heading convention)**: `### Out of Scope — <topic>` H3
  sub-headings used (5 of them in spec.md §F) — satisfies `OutOfScopeRule`.
- **B10 (untouched paths PRESERVE)**: `internal/cli/launcher.go` is touched
  ONLY for documentation alignment (no code change to
  `normalizeWorktreeFlag`); `cleanupMoaiWorktrees` is NOT modified.

## §C. Pre-flight (run before any code change)

```bash
# 1. Branch + baseline
git branch --show-current                                  # → main
git rev-parse HEAD                                         # → record SHA

# 2. Cross-platform build feasibility (no syscall changes; expected green)
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Existing lint baseline
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. Current defaults.go state (the mutation target)
sed -n '520,527p' internal/config/defaults.go              # confirm AutoCleanup: true, AutoMerge: true

# 5. Active-sessions registry schema (auto-isolation input)
test -f .moai/state/active-sessions.json && cat .moai/state/active-sessions.json | head -30

# 6. Worktree sprawl baseline (motivation)
git worktree list | wc -l                                  # → 58
git worktree list | grep -c "agent-"                       # → 31
```

## §D. Constraints (DO NOT VIOLATE)

- **PRESERVE `normalizeWorktreeFlag` behavior** (launcher.go:665-702) —
  byte-identical; this SPEC changes only doc references to when `-w` is
  emitted, not the flag's parsing.
- **PRESERVE `cleanupMoaiWorktrees` behavior** (launcher.go:354-434) — the
  stale `worker-*` cleanup on `moai cc` entry is out of scope.
- **PRESERVE the `WorkflowWorktreeConfig` struct** (types.go:476-483) — no
  field added / removed / renamed. Only the two default VALUES mutate.
- **PRESERVE i18n keys** (4 locales) — the auto-toggle title/desc keys
  already exist; this SPEC adds no new i18n strings.
- **Sanitized-pair obligation** (CLAUDE.local.md §2 [HARD] Template-First):
  any change to `.claude/rules/**/*.md` that has a template mirror at
  `internal/template/templates/.claude/rules/**/*.md` MUST be mirrored.
  Session-handoff.md and worktree-integration.md both have template mirrors;
  the run-phase implementer MUST run `make build` after editing them and
  verify byte-parity via `internal/template/split_namespace_test.go` /
  the mirror-parity CI guard.
- **Branch-guard compatibility** (sibling SPEC) — REQ-WES-005's auto-
  isolation MUST create a worktree (NOT mutate the primary checkout's
  branch state).
- **Forbidden commands**: `--no-verify`, `--amend`, force-push to main,
  `git reset --hard` in the primary checkout (the sibling branch-guard
  denies these regardless).
- **Required**: Conventional Commits + `🗿 MoAI` trailer per CLAUDE.local.md
  §4 + §18.

## §E. Milestones (priority-ordered; reversible decisions first)

### M1. Web auto-toggle defaults mutation (HIGH reversibility risk — data-model decision)

**Scope**:
- `internal/config/defaults.go:521` — `AutoCleanup: true` → `false`
- `internal/config/defaults.go:523` — `AutoMerge: true` → `false`
- Update `internal/config/defaults_test.go` (or the relevant test file) to
  assert the new defaults (3 assertions: AutoCreate=false, AutoCleanup=false,
  AutoMerge=false).

**Why first**: this is a behavior-changing default mutation. If a downstream
test or integration depends on the old `true` defaults, M1 will surface it
immediately — better to find out before the doc-alignment work in M2-M5.
Reversible via 2-line revert if downstream breaks.

**Acceptance**: AC-WES-004 (the three defaults read `false` from
`NewDefaultWorkflowConfig()`); existing tests green except the
defaults-assertion test (which is updated as part of M1).

**Decision Point 1 Resolution OQ-4 (FIRM)**: `TmuxPreferred: true`
(`defaults.go:525`) is OUT OF SCOPE — left unchanged. Only the three
auto-toggles mutate per REQ-WES-004. See spec.md §F "Out of Scope —
`TmuxPreferred: true` default mutation".

### M2. Doc-rule alignment — worktree-integration.md `EnterWorktree` expansion

**Scope**: expand `.claude/rules/moai/workflow/worktree-integration.md` §
`EnterWorktree` / `ExitWorktree` Tools (currently a single paragraph at
lines 148-150) to carry:

1. The EnterWorktree-first policy (current-session entry canonical form).
2. The L1-ephemeral vs L2-persistent re-entry distinction (REQ-WES-003).
3. The `moai cc -w` new-session launcher complement (REQ-WES-006).
4. The parallel-session branch conflict auto-isolation procedure (REQ-WES-005)
   — a new sub-section under Worktree Selection Rules.

**Why second**: the rule-file expansion is the load-bearing doctrinal change.
Subsequent milestones (M3 session-handoff, M4 CLAUDE.local.md, M5 examples)
reference this expansion, so it must land first.

**Sanitized-pair**: this file has a template mirror at
`internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`.
After editing, run `make build` and verify the template-parity CI guard is
green. The content MUST be §25-neutral (no SPEC IDs, no REQ tokens, no
internal-prefixed paths in the template mirror — see CLAUDE.local.md §25).

**Acceptance**: AC-WES-001 + AC-WES-003 + AC-WES-005 (grep evidence in the
expanded section).

### M3. session-handoff.md Worktree-Anchored Resume Pattern update

**Scope**: update `.claude/rules/moai/workflow/session-handoff.md` §
Worktree-Anchored Resume Pattern (lines 161-163) AND
`session-handoff-examples.md` § Worktree-Anchored Resume Pattern (Form A /
Form B) to:

1. Document BOTH canonical Block 0 forms: `moai cc -w <name>` (new-session
   launcher) and `EnterWorktree(<path>)` (current-session re-entry).
2. Replace the bare-`cd` Form B with the `EnterWorktree` form (per REQ-WES-002).
3. Add precondition `0)` to verify which form was taken.

**Why third**: session-handoff.md carries the always-loaded resume doctrine;
the expansion must align with M2's worktree-integration.md update.

**Sanitized-pair**: session-handoff.md has a template mirror; same
obligations as M2.

**Acceptance**: AC-WES-002 (no bare-`cd` in orchestrator-emitted Block 0).

**Decision Point 1 Resolution OQ-1 (FIRM)**: Block 0 Form B for L2
worktrees uses the extended `moai cc -w <abs-path>` (per REQ-WES-010,
after the M3a launcher extension lands) — replacing the legacy bare-`cd`
form. The Form B variant selection criteria (current-session continuation
vs cross-session launch) is:
- Current-session continuation (mid-SPEC, no `/clear`): prefer
  `EnterWorktree(<path>)` for L1 (`.claude/worktrees/<name>`).
- Cross-session launch (post-`/clear` or new terminal): use
  `moai cc -w <name-or-abs-path>` — works for both L1
  (`.claude/worktrees/<name>`) and L2 (`~/.moai/worktrees/<project>/...`
  via the REQ-WES-010 extension).
L2 current-session re-entry via the runtime `EnterWorktree` tool is OUT OF
SCOPE (deferred to follow-up runtime-layer SPEC).

### M3a. Launcher `-w` L2 path-resolution extension (HIGH reversibility risk — interface decision)

**Scope**: extend `internal/cli/launcher.go` `normalizeWorktreeFlag`
neighborhood (baseline cite `launcher.go:665`) to additionally resolve
`~/.moai/worktrees/<project>/...` absolute paths (Decision Point 1
Resolution OQ-1 → REQ-WES-010). The extension is **additive**: a new
sibling branch in the resolver handles L2 absolute paths; the existing
`.claude/worktrees/<name>` resolution path MUST remain behaviorally
intact (verified via the existing launcher test suite + a NEW test case
asserting the L2 absolute path is accepted).

**Why a dedicated milestone (not folded into M3)**: this is a real Go code
change to the launcher (M3's other scope is documentation alignment). The
extension is the load-bearing dependency for M3's Block 0 Form B for L2
worktrees — Form B's `moai cc -w <abs-path>` variant does not function
until M3a lands. M3a SHOULD land before M3's Block 0 Form B update is
considered final; M3's doc work may proceed in parallel.

**Test coverage**:
- Existing `normalizeWorktreeFlag` tests green (no regression to
  `.claude/worktrees/<name>` resolution).
- NEW test: `moai cc -w ~/.moai/worktrees/<project>/auto-<session-short>-<spec-id>`
  resolves to the L2 absolute path (NOT a new `.claude/worktrees/` dir).
- NEW test: `moai cc -w` with an absolute path NOT under
  `~/.moai/worktrees/` or `.claude/worktrees/` is rejected with a clear
  error (no silent fall-through to creating a new worktree).

**Acceptance**: AC-WES-010a + AC-WES-010b (new; see acceptance.md).

### M4. CLAUDE.local.md §22.8 addendum

**Scope**: add `CLAUDE.local.md` §22.8 — "web worktree auto-toggles default
OFF" recording the user-intended alignment with M1's defaults mutation.

**Why fourth**: CLAUDE.local.md is the SSOT for dev-settings intent; the
§22.8 addendum records WHY `AutoCleanup` / `AutoMerge` flipped to `false`.
This is a local-only file (no template mirror per CLAUDE.local.md §2), so
no sanitized-pair obligation.

**Acceptance**: AC-WES-006 (§22.8 present + cross-references defaults.go).

### M5. moai init / moai update doc surfaces

**Scope**: confirm `moai init --help` and `moai update --help` carry no
auto-enter-worktree behavior (they currently do not — verified in research.md
§G). Add an explicit note in the help text or the README that worktree entry
is explicit user opt-in via `--worktree` or subsequent `moai worktree new`.

**Why fifth**: documentation alignment only; no behavior change. The
existing code already satisfies REQ-WES-007; M5 makes the documentation
explicit.

**Acceptance**: AC-WES-007 (help text or README carries the explicit note).

### M6. Auto-isolation procedure implementation

**Scope**: the most complex milestone. Implement the parallel-session branch
conflict auto-isolation (REQ-WES-005):

1. Read `.moai/state/active-sessions.json` at the Pre-Spawn Sync Check
   point (already done by the orchestrator; no schema change).
2. When worktree entry is chosen AND ≥1 foreign session entry's recorded
   branch equals the current branch, auto-create a new worktree at
   `.claude/worktrees/auto-<session-short>-<spec-id>/` (or
   `~/.moai/worktrees/<project>/auto-<session-short>-<spec-id>/`).
3. Surface the auto-isolation as an info log (NOT an AskUserQuestion — the
   procedure is auto-resolving the race, not asking the user to resolve it).

**Why last**: this is the lowest-reversibility milestone (a new procedure
with runtime side-effects). It depends on M2's documentation (the procedure
is described there) and benefits from the prior milestones' stability.

**Decision Point 1 Resolution OQ-2 (FIRM)**: the auto-isolation confidence
predicate is **conservative** — any foreign active-session registry entry
triggers auto-isolation. False positives are cheap (an extra worktree is
inexpensive and user-deletable); false negatives corrupt the working tree
(a genuine conflict goes unresolved and produces cross-session branch-
state interference). The orchestrator surface this as an info log (NOT an
AskUserQuestion — the procedure auto-resolves the race). Stale-registry
false positives MAY produce a worktree that the user later deletes; a
follow-up SPEC MAY tighten the predicate to live-PID + cwd-match if the
30-day false-positive audit (acceptance.md §F Forward-Looking Checks)
shows a >50% stale-registry rate.

**Decision Point 1 Resolution OQ-3 (FIRM)**: the auto-isolated worktree
naming scheme is `auto-<session-short>-<spec-id>` where `<session-short>`
is the first 8 characters of the foreign session's UUID. No "or
equivalent" clause — the naming is deterministic so the auto-created
worktree is greppable and traceable to the originating foreign session.

**Acceptance**: AC-WES-005a + AC-WES-005b (auto-isolation fires under the
specified conditions; grep evidence of the procedure in
worktree-integration.md + a unit test if a Go code path is added).

### M7. Sync-phase close

**Scope**: sync-phase artifacts — CHANGELOG entry, README / docs-site
alignment (if applicable), 3-phase close per the spec-workflow.md Route B
contract (Tier L → Route B / PR route).

**Acceptance**: AC-WES-008 + AC-WES-009 + the §E.4 sync_commit_sha
populated in progress.md.

## §F. Anti-Patterns (DO NOT)

- **AP-WES-001**: emitting a bare-`cd` instruction in any orchestrator-
  emitted Block 0 / paste-ready resume / instruction surface. Use
  `EnterWorktree(<path>)` or `moai cc -w <name>` instead.
- **AP-WES-002**: using `Agent(isolation: "worktree")` to re-enter an
  existing L2 worktree. L1 ephemeral ≠ L2 persistent re-entry.
- **AP-WES-003**: mutating `normalizeWorktreeFlag` (launcher.go:665-702)
  — the `-w` flag parsing MUST remain byte-identical.
- **AP-WES-004**: triggering branch-state changes in the primary checkout
  during the auto-isolation procedure (REQ-WES-005). The branch-guard
  sibling will deny this regardless; ensure the procedure creates the
  worktree under `.claude/worktrees/` or `~/.moai/worktrees/`.
- **AP-WES-005**: skipping the sanitized-pair mirror verification after
  editing `.claude/rules/**/*.md` (CLAUDE.local.md §2 [HARD] Template-
  First Rule + §25 isolation). Run `make build` + verify mirror-parity CI.
- **AP-WES-006**: bundling `TmuxPreferred` mutation into M1 silently.
  Per Decision Point 1 Resolution OQ-4, `TmuxPreferred: true`
  (`defaults.go:525`) is OUT OF SCOPE — it MUST NOT be touched by M1. If a
  future SPEC revisits `TmuxPreferred`, it MUST be a separate SPEC with
  its own test assertion; do NOT silently mutate it as part of this
  SPEC's defaults mutation.

## §G. Risks

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| M1 default mutation breaks a downstream integration that relied on `AutoCleanup: true` | Low (the web console reads defaults via the same struct; the user must now explicitly toggle ON) | Medium (silent behavior change) | M1 includes the test-assertion update; if integration breaks, revert is 2 lines |
| M2 template-parity CI fails (sanitized-pair drift) | Medium (rule-file edits to a templated path) | Medium (CI blocks merge) | M2 includes `make build` + verify parity before commit |
| M6 auto-isolation false-positive churn (creates unnecessary branches) | Medium (registry stale entries are a known issue per memory `feedback_session_registry_stale_race_detection.md`) | Low (extra branch is cheap; user can delete) | OQ-2 conservative predicate default + info log; user can tune via future SPEC |
| Auto-isolation creates a branch in the primary checkout and trips the branch-guard | Low (the procedure explicitly creates a worktree) | High (denied; procedure fails) | Verified compatible per research.md §E.3; AP-WES-004 guards against this |
| Block 0 Form B default (`EnterWorktree`) surprises users accustomed to `moai cc -w` | Low (both forms documented; user chooses) | Low (doc clarity) | M3 documents BOTH forms with explicit selection criteria |

## §H. Self-Verification (run before reporting completion)

- AC binary PASS/FAIL matrix (all 9 ACs) — see acceptance.md
- `go build ./...` exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- `go test ./internal/config/...` — defaults_test green with new assertions
- `golangci-lint run --timeout=2m` — no NEW findings
- `grep -rn 'AskUserQuestion' internal/cli/ | grep -v _test.go | grep -v "^.*//"` — 0 matches (subagent boundary preserved)
- Branch HEAD + push state recorded
- Sanitized-pair mirror parity verified (template-neutrality CI green)

## §I. Decision Point 1 Resolutions (FIRM — clarified at plan→run gate)

All four Decision Point 1 open-question items (formerly marked with the
`NEEDS-CLARIFICATION` inline marker form in v0.1.0 research.md) are now FIRM
decisions (resolved by the user 2026-07-28 prior to plan-auditor independent
review). They are no longer open questions — they bind the run phase.

- **OQ-1 (RESOLVED — launcher extension)**: `moai cc -w` / `moai glm -w` /
  `moai cg -w` are EXTENDED to accept `~/.moai/worktrees/` absolute paths
  (L2 worktree entry), not just `.claude/worktrees/<name>`. This makes the
  launcher a valid re-entry path for L2 worktrees (resolves the path-split
  root cause at the launcher layer, which is under MoAI control). New
  REQ-WES-010 + new M3a milestone scope this extension. The EnterWorktree
  RUNTIME TOOL's `.claude/worktrees/`-only constraint is OUT OF SCOPE —
  deferred to a follow-up runtime-layer SPEC.
- **OQ-2 (RESOLVED — conservative predicate)**: auto-isolation fires when
  ≥1 foreign active-session registry entry exists (false positives are
  cheap; false negatives corrupt the working tree). No live-PID liveness
  check subroutine in M6.
- **OQ-3 (RESOLVED — naming)**: `auto-<session-short>-<spec-id>` where
  `<session-short>` = first 8 chars of the foreign session's UUID. No "or
  equivalent" clause.
- **OQ-4 (RESOLVED — TmuxPreferred OOS)**: `TmuxPreferred: true`
  (`defaults.go:525`) is OUT OF SCOPE — left unchanged. Only the three
  auto-toggles mutate per REQ-WES-004.

These decisions are recorded in spec.md §B.2 / §F (Out of Scope), spec.md
§C REQ-WES-005 / REQ-WES-010 + HISTORY 0.2.0, plan.md §E M1 / M3 / M3a /
M6, and acceptance.md §A matrix (REQ-WES-010 → AC-WES-010a/b) + §G closure
gates. No further AskUserQuestion round is required for Decision Point 1.

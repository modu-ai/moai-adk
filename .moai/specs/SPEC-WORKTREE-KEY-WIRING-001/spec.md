---
id: SPEC-WORKTREE-KEY-WIRING-001
title: "wire workflow.worktree auto_merge to a real local merge reader; re-scope auto_create honestly"
version: "0.2.0"
status: in-progress
created: 2026-09-12
updated: 2026-09-12
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/cli, internal/config"
lifecycle: spec-anchored
tags: "worktree, config-keys, auto-merge, integration-window, key-honesty, session-worktree"
tier: M
depends_on: [SPEC-SESSION-WORKTREE-001, SPEC-CONFIG-KEY-HONESTY-001]
---

# SPEC-WORKTREE-KEY-WIRING-001 — Worktree automation key wiring

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-12 | Initial plan-phase authoring (card t655). All baseline facts measured in-tree at `1d150a27d`. Three design decisions settled in-plan: (1) auto-merge ACQUIRES the integration window itself (Option A, never `--force`, skip-with-notice on a live holder); (2) the integration target reuses the existing `git_strategy.<mode>.develop_branch` key (no new key — one fact, one key, window record and merge cannot disagree); (3) `auto_create` resolves via the card-sanctioned explicit-scope route (advisory wording stays its reader; the false auto-creation claim in the wording is removed). | manager-spec |
| 0.2.0 | 2026-09-12 | Iteration-2 revision clearing plan-audit iteration 1 (FAIL, 0.75). Delta only, per audit report `.moai/reports/t655/plan-audit.md`: D1 — AC-WKW-013 (notice-prefix distinctness, REQ-WKW-009) and AC-WKW-014 (toggle independence, REQ-WKW-013) added to acceptance.md; D2 — REQ-WKW-002's trigger clause rewritten to "While a session exits with a live session worktree", removing the disposal-coupled phrasing that read as auto_cleanup gating the merge; D3 — AC-WKW-012's judging command corrected (real umbrella test name `TestShippedConfigKeysHaveReaders`; the zero-match grep converted to `! grep -q` so the correct state exits 0). No design decision, requirement, or milestone changed. | manager-spec |

## §A Problem / Motivation

`workflow.worktree.auto_merge` is a declared config key with ZERO production
readers. `internal/config/types.go` records: "AutoMerge has no production
reader (declared but not read)". A user setting `auto_merge: true` gets
nothing. The shipped template's own comment apologizes for it
(`internal/template/templates/.moai/config/sections/workflow.yaml`:
"auto_merge / auto_cleanup: declared but not read — no code path consumes"),
and the shipped-key inventory classifies it `D` (delete) with a
`deprecate_after: v3.1.0` marker — a declared feature scheduled for deletion.

Its sibling `workflow.worktree.auto_create` has the opposite defect: it IS
read (internal/cli/worktree_advisory.go), but only to select advisory wording
— and the wording selected when it is `true` makes a FALSE claim ("...
is auto-creating a worktree for isolation") that no code performs.

This SPEC is the successor to SPEC-CONFIG-KEY-HONESTY-001: that audit made
the reader status honest; this one gives the dead key a real, safe reader and
makes the misleading one truthful. The reader lives where the behavior
already lives — the session-worktree family (internal/cli/session_worktree*.go),
beside the two `auto_cleanup` consumers (session-exit disposal and the M8
PR-merge sweep).

**Binding design decision — the integration window (settled here, not
deferred):** auto-merge must NOT bypass the `moai integration` window. The
settled choice is **Option A: the auto-merge path ACQUIRES the window itself
(acquire → merge → release), through the same lock API the CLI verb uses, and
never displaces a holder (no `--force`)**. Rationale in design.md §1; the
binding requirement is REQ-WKW-004.

## §B Requirements (GEARS)

**REQ-WKW-001 (inert default).** The CLI shall leave every session-exit
worktree path byte-identical to its pre-SPEC baseline — no git merge
invocation, no integration-window record write, no notice — while
`workflow.worktree.auto_merge` is `false` (the distributed template default).

**REQ-WKW-002 (trigger).** While a session exits with a live session
worktree, where `workflow.worktree.auto_merge` is `true`, when the exit
is clean (exit 0) and the worktree's branch carries at least one commit
absent from the configured integration branch, the CLI shall merge that
branch into the integration branch with a local `git merge --no-ff`. A
non-zero exit shall produce no merge (clean-exit-only, mirroring REQ-SW-009).

**REQ-WKW-003 (integration target; inert when empty).** The auto-merge
target shall be the project's configured git-flow integration branch —
`git_strategy.<mode>.develop_branch`, the same key and the same resolution
(`config.LoadGitFlowIntegrationConfig`) that `moai integration acquire` uses
for its window record. Where the value is empty, the mode is not
manual git-flow, or the file is unreadable, the CLI shall perform no merge
and shall emit a non-blocking notice naming the config key.

**REQ-WKW-004 (window ceremony — BINDING).** Where `workflow.worktree.auto_merge`
is `true`, when the auto-merge path runs, the CLI shall record the
release-integration window via the same lock API that
`moai integration acquire` uses BEFORE the merge and release it AFTER the
merge terminates (success or failure), and shall never displace any existing
hold — no forced takeover, live or stale. When the window is held (live or
stale), the CLI shall skip the merge and emit a non-blocking notice.

**REQ-WKW-005 (local only; never push).** The auto-merge path shall never
execute a remote-mutating git command. Its only branch-mutating invocation
shall be `git merge --no-ff <session-branch>` executed inside the
integration branch's worktree. Pushing remains the lead/operator's
explicit act.

**REQ-WKW-006 (conflict safety).** When the merge fails or reports
conflicts, the CLI shall restore the integration worktree to its pre-merge
state (`git merge --abort` while a merge is in progress), release the
window, emit a non-blocking notice, and leave the session worktree
untouched.

**REQ-WKW-007 (source dirty guard).** When the session worktree holds
uncommitted changes at trigger time, the CLI shall skip the merge and emit a
non-blocking notice; uncommitted work is never merged and never destroyed.
Disposal then proceeds under `auto_cleanup`'s own dirty guard.

**REQ-WKW-008 (target guards).** When no worktree has the integration branch
checked out, or that worktree is dirty, the CLI shall skip the merge and
emit a non-blocking notice.

**REQ-WKW-009 (non-blocking family).** Every auto-merge failure path shall
be a non-blocking stderr notice carrying a literal prefix distinct from both
`SessionExitCleanupNoticePrefix` and `PRMergeCleanupNoticePrefix`, and shall
never abort or fail the session-exit flow (REQ-SW-004 fail-open spirit,
EC-13 attributability).

**REQ-WKW-010 (no-op skip).** When the session branch carries no commit
absent from the integration branch, the CLI shall skip the entire ceremony —
no window record, no merge invocation, no notice (byte-identical to
REQ-WKW-001's baseline).

**REQ-WKW-011 (auto_create scope).** The `workflow.worktree.auto_create`
key's declared scope shall be the advisory wording selection on
init/update/web, and the advisory emitted under `auto_create: true` shall
contain no claim of automatic worktree creation that no code performs, while
still satisfying the AC-WBG-009 observability regex
(`worktree.*isolation|use a worktree|moai (cc|cg) -w|claude --worktree`).

**REQ-WKW-012 (honesty artifacts in the same change).** Every reader-status
artifact shall reflect the new state in the same change: the
`WorkflowWorktreeConfig` comment in `internal/config/types.go` (the "no
production reader" sentence for AutoMerge ends; the AutoCreate sentence
states the declared scope), the `internal/config/defaults.go` worktree
comment, the shipped template `workflow.yaml` comment (including its stale
"auto_cleanup: declared but not read" half — AutoCleanup has had two readers
since SPEC-SESSION-WORKTREE-001), `shipped_key_inventory.yaml`
(`workflow.worktree.auto_merge`: class D → W, `deprecate_after` removed), and
the reader-classification test expectations in
`internal/config/shipped_key_reader_test.go`.

**REQ-WKW-013 (toggle independence).** `auto_merge` and `auto_cleanup` shall
gate only their own behavior: each of the four combinations (merge/remove)
is reachable, and enabling one never implies the other.

## §C Exclusions

### Out of Scope — other dead keys

- `workflow.worktree.session_name_pattern` and `workflow.worktree.tmux_preferred` stay declared-but-unread (inventory class D); a successor card owns them.
- No new reader for `git_strategy.worktree_root` (deprecated, no consumer — preserved per SPEC-CONFIG-001).

### Out of Scope — remote operations

- No push, no fetch, no origin absorb step, no PR creation anywhere on the auto-merge path. The manual lane ceremony's `git merge origin/develop` absorption is deliberately NOT automated (design.md §1.3); the auto path is a pure local merge.
- Divergence between local develop and origin/develop is accepted; the merge notice names the merge commit so the lead's batch-push flow can read it.

### Out of Scope — other workflows and surfaces

- A generic per-workflow integration-target key (e.g. a github-flow target) — auto-merge is deliberately git-flow-manual-mode-gated because it shares `develop_branch` with the window record.
- The M8 on-touch surface (`session register` / `session list`) stays cleanup-only; an auto-merge sweep over OTHER sessions' WT-* branches would merge in-flight work and is prohibited.
- `moai cc` launcher auto-entry and any PreToolUse-hook-side auto-creation.

## §D Acceptance

Acceptance criteria live in `acceptance.md` (AC-WKW-001..014), each carrying
a binary judging command. The critical invariants: OFF = byte-identical
baseline (AC-WKW-001), window held across exactly the merge (AC-WKW-005/006),
zero push invocations on the path (AC-WKW-006), conflict leaves no MERGE_HEAD
(AC-WKW-007), notice prefixes distinct (AC-WKW-013), toggles independent
(AC-WKW-014).

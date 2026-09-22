---
id: SPEC-WORKTREE-CREATE-VERB-001
title: "Harness-neutral worktree creation verb — promote the existing materializeSessionWorktree creation capability into a moai CLI verb surface"
version: "0.2.0"
status: completed
created: 2026-09-22
updated: 2026-09-22
author: manager-spec (card t1070)
priority: P2
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-lite
tags: "worktree, cli-verb, harness-neutral, codex, reuse-first, worktree-integration"
tier: M
related_specs:
  - SPEC-SESSION-WORKTREE-001
  - SPEC-WORKTREE-BASEREF-001
  - SPEC-CLI-WORKTREE-FLAG-RACE-001
---

# SPEC-WORKTREE-CREATE-VERB-001 — Harness-neutral worktree creation verb

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-09-22 | 0.2.0 | Implemented option 가 (`moai worktree new <name>`), closed independent-audit finding F1, and completed sync evidence (card t1070). |
| 2026-09-22 | 0.1.0 | Initial draft authored by manager-spec (card t1070, plan phase). |

## §A Overview

moai's own doctrine (`kanban-dispatch.md`, `AGENTS.md` §3, `gitflow-lane-protocol.md` §1) mandates that every harness enters a worktree through the launcher — never bare `git worktree add` — yet the only creation path that exists today is claude-harness-only: `moai cc -w` delegates creation to the `claude` binary it execs. `moai codex -w` resolves an existing tree and explicitly refuses creation. Codex lanes, scripts, and harness-neutral provisioning therefore have no authorized creation path.

Meanwhile, a complete harness-neutral creation implementation already exists inside moai: `materializeSessionWorktree` (`internal/cli/session_worktree.go:203`), gated default-OFF and consumed only as a side effect of `moai init` / `moai web` / `moai profile`.

This SPEC defines the requirements for exposing that existing capability as a moai CLI verb. The M1 live-Codex observation and operator relay selected option 가: revive `moai worktree new` as the harness-neutral surface. The retired implementation and its `--base` / `--from-current` flags remain retired; the new command is a thin adapter over the current `materializeSessionWorktree` plumbing.

## §B Requirements (GEARS)

- REQ-WCV-001 (Ubiquitous) — The moai CLI shall expose exactly one harness-neutral worktree creation verb surface, wired to the existing `materializeSessionWorktree` creation plumbing (including `gitWorktreeAddArgs`, `LoadWorktreeBaseBranch` base selection, and `applyWorktreeGitConfig` post-create tiers), and shall not introduce a second `git worktree add` invocation path.

- REQ-WCV-002 (Event-driven) — When the creation verb is invoked with a `<name>` value, the CLI shall create a git worktree at the conventional L1 path `.claude/worktrees/<name>` on a newly created branch, using the existing creation plumbing's base-branch behavior (per SPEC-WORKTREE-BASEREF-001 REQ-WBR-010/011) without modification to that behavior.

- REQ-WCV-003 (Event-driven) — When the creation verb is invoked without a `<name>` value, the CLI shall exit non-zero with a structured stderr error naming the expected argument, and shall not create any directory or worktree.

- REQ-WCV-004 (Event-driven) — When a destination collision is detected (`.claude/worktrees/<name>` already exists as a worktree or a path, or the branch name is already taken), the CLI shall refuse creation, exit non-zero, and emit an error diagnostic consistent with the existing resolve-error shape (`moai codex -w` missing-tree message family).

- REQ-WCV-005 (Capability gate) — Where the session-worktree feature gate is default-OFF (`config.SessionWorktreeEnabled`), the introduction of the creation verb shall not change the gate's default value and shall not alter REQ-SW-001's byte-identical baseline behavior of the existing `enterSessionWorktree` consumers. The creation verb surface does NOT consult `config.SessionWorktreeEnabled`: it creates unconditionally when invoked (matching REQ-WCV-002's unconditional create), while the SPEC-SESSION-WORKTREE-001 REQ-SW-001 default-OFF auto-entry baseline for `moai init` / `moai web` / `moai profile` remains UNCHANGED.

- REQ-WCV-006 (Ubiquitous) — The creation verb surface shall be non-interactive — positional arguments + flags + structured stderr errors per `internal/cli/CLAUDE.md` C-HRA-008 / REQ-PGN-012 — and shall carry a static guard test (the `TestNew_NoAskUserQuestion` pattern) covering the new surface.

- REQ-WCV-007 (Ubiquitous) — The creation verb shall place created trees under the conventional L1 path `.claude/worktrees/<name>` and shall preserve the L1/L2 boundary semantics defined in `worktree-integration.md` § Terminology Glossary (L2 absolute-path forms follow `resolveWorktreeL2Path` validation where the chosen surface resolves paths).

- REQ-WCV-008 (Ubiquitous) — All error messages emitted by the creation verb surface shall be in English, per the project `error_messages: en` setting, and all code comments in English per `code_comments: en`.

- REQ-WCV-009 (Event-driven) — When the creation verb surface is invoked with a `<name>` value that would escape the conventional worktrees root (traversal via `..`, path-separator tricks, or an absolute L2 form outside the sanctioned prefix validated by the existing `resolveWorktreeL2Path` rule), the CLI shall refuse creation, exit non-zero, and emit a structured English error; and when the resolved destination exists as a plain non-worktree directory, the CLI shall likewise refuse creation and exit non-zero without modifying that directory.

## §C Scope Decisions

- In: one verb surface (surface per plan.md decision gate), wiring of existing creation plumbing, focused tests (verb behavior + collision refusal + gate baseline + static guard), English diagnostics.

### AC Traceability (scenarios live in acceptance.md: AC-WCV-001 .. AC-WCV-008)

- REQ-WCV-001 → AC-WCV-001 (single creation path, existing plumbing executor)
- REQ-WCV-002 → AC-WCV-001 (creation at `.claude/worktrees/<name>`, base-branch behavior)
- REQ-WCV-003 → AC-WCV-002 (refusal without `<name>`, structured English error)
- REQ-WCV-004 → AC-WCV-003, AC-WCV-006 (collision refusal; existing resolve diagnostics unchanged)
- REQ-WCV-005 → AC-WCV-004 (default-OFF gate baseline byte-identical)
- REQ-WCV-006 → AC-WCV-005 (static guard test, non-interactive)
- REQ-WCV-007 → AC-WCV-007 (L1 placement, L1/L2 boundary preserved)
- REQ-WCV-008 → AC-WCV-002, AC-WCV-006 (English diagnostics asserted in both)
- REQ-WCV-009 → AC-WCV-003, AC-WCV-008 (path-escape refusal; plain-directory destination refusal)
- Reuse contract: `materializeSessionWorktree` plumbing, `resolveWorktreeL2Path` validation, `LoadWorktreeBaseBranch` base selection are consumed, not reimplemented.
- The (가)/(나) surface decision is intentionally NOT made in this SPEC body; see `plan.md` for the decision gate and its measurement precondition.

## §D Out of Scope

### Out of Scope — sibling doc-surface cards (t1071, t1072)

- Card t1071 (contract reinforcement): adding the AGENTS.md capability-binding row and registering `moai codex -w` in launcher documentation surfaces.
- Card t1072 (contradiction cleanup): the `worktree-integration.md:227` `git -C` marking correction.

### Out of Scope — data-loss axis (t1073)

- Any change to `moai worktree done` tier semantics (the L1/L2 disposal contract and its data-loss axis belong to card t1073).

### Out of Scope — alternate creation implementations

- Any second `git worktree add` invocation path (the reuse contract in REQ-WCV-001 forbids it).
- Restoration of the retired `moai worktree new` implementation, its bodp library, or its `--base`/`--from-current` flag semantics.
- Changes to `materializeSessionWorktree`'s creation algorithm itself (it is consumed as-is).
- Docs-site or template-mirror surface changes.

## §E Constraints

- `development_mode: tdd` (quality.yaml) — RED-GREEN-REFACTOR applies at run phase.
- Errors/comments in English (REQ-WCV-008).
- No interactive prompts (REQ-WCV-006).

---
id: SPEC-CC-GD124-001
title: "Claude Code upstream drift sweep repair GD-1/2/3/4 — context-window table and cross-session messaging availability (2 rules x local/template mirror)"
version: "0.1.2"
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/template/templates/.claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "claude-code-upstream, drift-repair, rules, template-parity, documentation"
tier: M
---

# SPEC-CC-GD124-001 — Claude Code Upstream Drift Repair GD-1/2/3/4

## HISTORY

- 2026-09-07 v0.1.0 — Plan-phase artifacts authored by manager-spec (Factory-Mode card t491, lane-2, worktree `.claude/worktrees/t491`, branch `WT-cc-upstream-sweep`, base `615d18c1f`).
- 2026-09-07 v0.1.1 — frontmatter tier correction (S→M): artifact set (3 files) and REQ/AC budget match Tier M; absent tier field would default to Tier L.
- 2026-09-07 v0.1.2 — plan-audit iter1 repair: D1-D3 blocking + D4-D6/D8/D9 ride-alongs.

## Context (WHY)

The 2026-09-06 release-update sweep found Claude Code upstream drift in two ALWAYS-LOADED rule files (read by every session, every turn). This card repairs exactly the findings GD-1, GD-2, GD-3, GD-4. All four facts were re-verified on 2026-09-07 from primary sources (Claude Code CHANGELOG v2.1.234/v2.1.247/v2.1.248/v2.1.257 and the official messaging docs); do not re-litigate them.

Evidence baseline (pinned to tree `615d18c1f`):

- Verification record: `.moai/reports/t491/upstream-verification-20260907.md` (V1-V10 evidence table, baseline attribution, gaps, residual risk).
- Verbatim CHANGELOG extracts (committed): `.moai/reports/t491/t491-v247.md`, `.moai/reports/t491/t491-v248.md`, `.moai/reports/t491/t491-v257.md`.
- Sweep report copy: `.moai/reports/t491/sweep-report-copy-20260906.md`.
- Worktree re-measure 2026-09-07 (plan-phase, this tree): `Fable (256K)`=1, `200K/256K`=2 (cw); native-Windows denial=1, `unavailable on Amazon Bedrock`=1, `turning messaging off silently`=1, `tengu_harbor_kite`=1 (csm); cw local/template `diff -q` identical; csm local/template `diff -u` hunk count=1.

## 1. Requirements (GEARS)

- REQ-001 — The `context-window-management.md` Context Window Targets table shall carry a `Fable (1M)` row at `1,000,000 tokens` / `**50%**` / `~500,000 tokens`, replacing the stale `Fable (256K)` row.
- REQ-002 — The same table shall carry a `Sonnet 5 (1M)` row at `1,000,000 tokens` / `**50%**` / `~500,000 tokens`, replacing the `Sonnet/Opus standard (200K)` row that buried Sonnet 5.
- REQ-003 — **When** the `Sonnet 5 (1M)` row is inserted, the table shall retain a distinct `Sonnet 4.x / earlier standard (200K)` row at `200,000 tokens` / `**90%**` / `~180,000 tokens` immediately after it, so real-200K model classes (Sonnet 4.x-era, Haiku 4.5) stay on 200K/90% rows. Rationale (HARD): moving a real-200K class to 1M/50% flips the error direction from "hand off too early" (safe) to "run past the real ceiling" (stream stall).
- REQ-004 — **When** the 256K model class no longer exists in the table, every prose threshold mention of `90% on 200K/256K` in the same file shall read `90% on 200K` (two occurrences: the user-responsibilities bullet and the pre-clear-announcement clause).
- REQ-005 — The `context-window-management.md` edit shall be applied identically to the template mirror such that the local copy and `internal/template/templates/.../context-window-management.md` remain byte-identical (`diff -q` exit 0). The pair is byte-identical today and must stay so.
- REQ-006 — The `cross-session-messaging.md` OS bullet shall state same-machine messaging availability on native Windows since Claude Code v2.1.234 (inbox socket is a named pipe) in place of the blanket "does not provide cross-session messaging on native Windows" denial, and shall state cross-machine reach from native Windows as an explicit documentation gap — claimed neither available nor unavailable.
- REQ-007 — The `cross-session-messaging.md` Providers bullet shall state its two axes separately: (same machine) available on every provider — the four named ones included — since v2.1.248, delivery riding a per-session socket on the machine that never leaves it; (beyond this machine) still unavailable with an API key and on Amazon Bedrock, Claude Platform on AWS, Agent Platform on Google Cloud, and Microsoft Foundry. The four provider names shall remain present in the cross-machine statement.
- REQ-008 — The `cross-session-messaging.md` Flags bullet shall state, at class level only, that since v2.1.248 same-machine messaging works in sessions with feature-flag fetching off on every provider; it shall retain the four env flags as flag-evaluation disables, retain the below-v2.1.248 historical note (the capability is new in that release), and retain the `/list-agents` recognized/unrecognized diagnostic sentence. No per-flag claims shall be made for the four flags.
- REQ-009 — Edits to `cross-session-messaging.md` § Availability constraints shall preserve its five-bullet structure: the intro line "Five constraints" shall remain accurate after the edit (no bullet added, none merged away, none split).
- REQ-010 — **When** either rule file is edited, the local copy and the template mirror shall carry the same edit, such that (a) the context-window pair stays byte-identical, and (b) the cross-session-messaging pair stays identical except for the single known template-neutrality hunk — the `> Origin: SPEC-CODEX-SESSION-MSG-001 (design.md §8 mapping).` line present in the local copy only (added by `9ef2b91e1`, card t187; SPEC IDs are forbidden template content per template isolation doctrine §25.1 class C1). No new divergence hunks shall be introduced.
- REQ-011 — Lines outside the change map shall remain untouched: cw L30-34 (the GLM-5.3 section, including its `(Opus=1M, Sonnet/Haiku=200K)` reporting parenthetical); csm L23 (versions bullet), L25 (shared-flag-slot bullet), L9 ("same platforms" phrasing).

## 2. Non-Goals

Constraints this SPEC deliberately does not pursue:

- GD-5, GD-7, GD-9 from the sweep are untouched by this card.
- GD-6 is operator-held; GD-8 is no-action. This card takes no position on either.
- No cross-file threshold propagation: `session-handoff.md` and `CLAUDE.md` §16 carry "1M = 50%, 200K = 90%" — still true after this change, no edit.

## 3. Acceptance Criteria (index)

Full two-cell (RED-now pinned to `615d18c1f` / green-path) Given-When-Then criteria with single-invocation pinned commands live in `acceptance.md`. Index:

| AC | Asserts | RED-now (615d18c1f) |
|----|---------|---------------------|
| AC1 | `Fable (256K)`=0, `Fable (1M)` present, both cw copies | `Fable (256K)`=1 |
| AC2 | `200K/256K`=0 in both cw copies | =2 (L64+L80) |
| AC3 | `Sonnet 5 (1M)` row present AND `Sonnet 4.x / earlier standard (200K)` row present, both copies | both =0 |
| AC4 | OS bullet: stale denial=0, `v2.1.234` present, both copies | denial=1 |
| AC5 | Providers bullet: `unavailable on Amazon Bedrock`=0, `every provider`+`v2.1.248` present, four names retained, both copies | Bedrock denial=1 |
| AC6 | Flags bullet: `turning messaging off silently`=0, `feature-flag fetching off` present, diagnostic retained, both copies | silent=1 |
| AC7 | Control/invariant: cw pair `diff -q` exit 0; csm pair diff = exactly 1 hunk (the Origin-line blockquote only) | green today — MUST NOT flip |
| AC8 | Scope control: untouched strings intact; git status extras-free; five-bullet structure preserved (`Five constraints` + 5 bullet anchors), both copies | clean |
| AC9 | `make build` exit 0 AND `go test ./internal/template/...` exit 0 | template embed sanity |

## 4. Out of Scope

### Out of Scope — Remaining sweep findings

- GD-5 (upstream drift item five) — not repaired by this card.
- GD-6 — operator-held decision; not actionable by this card.
- GD-7 — not repaired by this card.
- GD-8 — no-action verdict recorded by the sweep; nothing to build.
- GD-9 — not repaired by this card.

### Out of Scope — Untouched lines in the target files

- cw L30-34 GLM-5.3 section, including its `(Opus=1M, Sonnet/Haiku=200K)` parenthetical — describes upstream reporting behavior, unverified, out of scope.
- csm L23 versions bullet and L25 shared-flag-slot bullet — no edit.
- csm L9 "same platforms" phrasing — no edit.

### Out of Scope — Cross-file threshold mentions

- `session-handoff.md` threshold references ("1M = 50%, 200K = 90%") — still true post-fix, no edit.
- `CLAUDE.md` §16 threshold references — still true post-fix, no edit.

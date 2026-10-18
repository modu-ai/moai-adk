---
id: SPEC-AGENT-CORE4-001
title: "Agent core four, piece 1: agent sources, two renders, generated copies"
version: "0.3.0"
status: draft
created: 2026-10-18
updated: 2026-10-18
author: "Goos Kim"
priority: P1
phase: "v3.2.0 target"
module: "internal/template/agentemit"
lifecycle: spec-anchored
tags: "agents, agent-source, codex-render, generated-copies"
tier: L
related_specs: [SPEC-CODEX-DUAL-AGENTS-001, SPEC-AGENT-MODEL-INHERIT-001]
---

# SPEC-AGENT-CORE4-001 — Piece 1: agent sources, renders, generated copies

## §A Background

Twelve agent definitions ship in four hand-managed copies that drift: 7 of 12 project/template
Markdown pairs differ and the project Codex mirror holds 11 TOML files against the template's 12
(`acceptance.md` §2). Piece 1 introduces one **agent source** per definition and renders the
template Markdown (C2) and template Codex (C3) from it; for the four agents that stay in the roster
it also renders the project Markdown (C1) and project Codex (C3p). It changes no agent behavior:
every shipped byte of the other eight definitions stays as it is, and the carried four change on
exactly 8 pinned lines.

Terminology: **agent source** (directory `internal/template/agentemit/source/`) is the new input.
The older comments in `Makefile`, `emit.go` and `loader.go` that call the `.md` files "the neutral
layer" are updated at M1 so that "neutral" is no longer used for either.

History: the original single SPEC (commit `b4ea2d6c93`) failed plan-audit iter1 (0.66) and was
split by the leader; this narrowed SPEC failed iter2 (0.67). Evidence of the original scope is in
that commit's `research.md` and `design.md`.

## §B Definitions

- **C1** `.claude/agents/moai/*.md`; **C2** `internal/template/templates/.claude/agents/moai/*.md`;
  **C3** `internal/template/templates/.codex/agents/moai/*.toml`; **C3p** `.codex/agents/moai/*.toml`.
- **The eight** — manager-spec, manager-develop, manager-docs, manager-git, manager-lead,
  manager-todo, plan-auditor, sync-auditor. **The carried four** — builder-harness, e2e-tester,
  manager-design, super-advisor. **The twelve** — both sets.
- **Pin file** `internal/template/agentemit/testdata/agents-head.sha256` — sha256 of the 24 C2 and
  C3 files of the twelve at the parent of the M0 commit. **Post pins** — the 8 sha256 values of
  `acceptance.md` §3, for the carried four after the 8 line changes.
- **specs** in the check summary — the number of agent sources (12 from M1 on).
- **copies** — the number of file sets compared: 2 at M1 (C2, C3), 4 at M2 (C1, C2, C3, C3p; the
  project sets cover the carried four only).

## §C Requirements (GEARS)

- **REQ-AC4-001** The agent definition set shall have exactly one agent source per definition and
  no rendered C2 or C3 file without one.
- **REQ-AC4-002** The Claude renderer shall produce each C2 definition directly from its agent
  source, keeping blocks marked for Claude and dropping blocks marked for Codex.
- **REQ-AC4-003** The Codex renderer shall produce each C3 definition directly from the same agent
  source, keeping blocks marked for Codex and dropping blocks marked for Claude, and shall leave
  every permission-contract row of the twelve roles unchanged.
- **REQ-AC4-004** When a capability token of an agent source has no mapping, the emitter shall fail
  emission naming the file and the token.
- **REQ-AC4-005** From M1 on, rendering the twelve sources shall yield C2 and C3 files whose bytes
  equal the pin file, except the carried four after M2.
- **REQ-AC4-006** From M2 on, the carried four shall differ from the pin file on exactly the 8 lines
  of `acceptance.md` §3, each new line shall contain no `~/`, and the resulting C2 and C3 files
  shall equal the post pins.
- **REQ-AC4-007** From M2 on, C1 shall equal C2 and C3p shall equal C3 for the carried four.
- **REQ-AC4-008** When a rendered file differs from a fresh render, the read-only emission check
  shall fail naming the file.
- **REQ-AC4-009** When `make agents-emit-check` passes, it shall print `specs=<n> copies=<k>
  diff=0`.
- **REQ-AC4-010** The commit history shall show the pin file added by the M0 commit alone, no
  commit changing an agent source or a rendered definition before M1, and no later change to the
  pin file.
- **REQ-AC4-011** The roster guard's numeral layer shall report no undischarged hit for the agent
  source directory.
- **REQ-AC4-012** The existing emission tests shall keep every assertion they hold at the pin
  tree, the Markdown identity assertion being replaced by equality with the pinned bytes.

## §D Out of Scope — follow-on pieces

### Out of Scope — piece 2 (retirement, guards, replacement paths)

Former M3, M4, M5, M7, M8, a separate SPEC the leader queues after piece 1 lands (no id yet): the
core four agents (planner, implementer, documenter, auditor) and worker, the size verb, the auditor
verdict-only write guard, the agent memory cap, spawn-record enrichment, replacement paths for
manager-lead (a duty-by-duty map of the leader working mode), manager-todo and manager-git
(including a non-factory owner for `delivery-policy.md` rule 1 and the `OwnerRole` value at
`internal/mission/integration.go:38`), the 977-file reference flip, the retired-name refusal,
cutover, and `auditor` deferred emission.

### Out of Scope — piece 3 (list hiding)

Former M6, a separate SPEC queued after piece 2 (no id yet): `moai-agent-*` hiding settings for
Claude and Codex, the preload probe, the doctor check.

### Out of Scope — other

- docs-site and README pages, user-scope `~/.claude/agents`, and the `AGENTS.local.md` §2.0 text
  (user-owned; amendments are named in decisions Q4 and Q17).
- Any change of agent bodies, tools, model, effort or permission mode beyond the 8 pinned lines.

## §E Hand-off table (audit defects not fixed here)

| Defect | Owner | Content to inherit |
|---|---|---|
| iter1 D2 guard root | piece 2 | root = git toplevel of the acting agent's cwd; linked-worktree and primary-checkout subtests; `internal/hook/path_resolve.go:69-76` order |
| iter1 D4 refusal predicate | piece 2 | active only when name in table AND no `<name>.md` in the project's agents directory |
| iter1 D5 replacement paths | piece 2 | leader duty map, non-factory manager-git owner, `moai-factory` skill waiver or sequencing |
| iter1 D7 skill and reference publication | pieces 2, 3 | deployed location in both harnesses, catalog and embed registration, `on_demand_reference_bytes` |
| iter1 D8 phased rewrite vs `auditor` emission | piece 2 | rewrite in the emission commit or add a spawn-resolution gate |
| iter1 D12 numeral layer count rows, threshold at 17 | piece 2 | `threshold_17_is_12`, stale-subset fixture |
| iter1 D13 shrinking edit on oversize file | piece 2 | allow when result is smaller than the current file |
| iter1 D14 core-body preservation | piece 2 | pinned [HARD]-clause list; implementer budget row |
| iter1 D15 "never blocks a spawn" statements | piece 2 | rewrite sentence and file header; deny in a separate handler |
| iter1 D18 doctor check name | piece 3 | registered check is "Skills Allowlist" (`internal/cli/doctor.go:306`) |
| pin lifecycle | piece 2 | `agents-head.sha256` and AC5 are retired or retargeted when piece 2 absorbs the eight |
| Q2, Q3, Q5-Q15 | pieces 2, 3 | `decision-index.md` hand-off rows |

## §F HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-10-18 | Initial unsplit plan (commit `b4ea2d6c93`). |
| 0.2.0 | 2026-10-18 | Narrowed to piece 1 after plan-audit iter1 FAIL (commit `8260bbaa0b`). |
| 0.3.0 | 2026-10-18 | Plan-audit iter2 FAIL fixes: Tier L, all twelve sources at M1, both renderers read the source, pinned line pairs and post hashes, path-based ordering ACs, decisions Q17 and Q18. |

---
id: SPEC-AGENT-CORE4-001
title: "Agent core four, piece 1: agent sources, two renders, generated copies"
version: "0.4.0"
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
- **Baseline** — the tree at the parent of the M0 commit (`<M0>^`); the M0 commit records its SHA in
  `m0-baseline.md` and as `m0_commit:` in `progress.md`. No hash file is frozen: the baseline bytes
  are read from git (`git show <M0>^:<path>`), so a foreign commit absorbed before M0 moves the
  baseline instead of making a frozen value stale.
- **Agent source** — a verbatim copy of the shipped definition file (frontmatter keys `name`,
  `description`, `tools`, `color`, `permissionMode`, `isolation`, `memory`, `skills`, `hooks`,
  `model`, `effort` as shipped, then the body); the only addition is the optional block markers.
  The `tools:` CSV is the capability list; a neutral capability vocabulary is deferred to piece 2.
- **specs** in the check summary — the number of agent sources (12 from M1 on).
- **copies** — the number of file sets compared: 2 at M1 (C2, C3), 4 at M2 (C1, C2, C3, C3p; the
  project sets cover the carried four only).

## §C Requirements (GEARS)

Operator decision `d-20261018T062320Z-5952` (decided_by operator, 2026-10-18, scope card:t1649): agent copies become
generated outputs; the agent source is the only edit surface; C2 is rendered from it; C1 is
byte-identical to C2; hand-editing C2 fails `make build`; the contrary `AGENTS.local.md` §2.0
sentences are superseded and rewritten by this work; other in-flight cards move their agent edits
to the source layer when they land after piece 1.

- **REQ-AC4-001** The agent definition set shall have exactly one agent source per definition and
  no rendered C2 or C3 file without one.
- **REQ-AC4-002** The Claude renderer shall produce each C2 definition directly from its agent
  source, keeping blocks marked for Claude and dropping blocks marked for Codex.
- **REQ-AC4-003** The Codex renderer shall produce each C3 definition directly from the same agent
  source, keeping blocks marked for Codex and dropping blocks marked for Claude, and shall leave
  every permission-contract row of the twelve roles unchanged.
- **REQ-AC4-004** When a `tools:` token of an agent source has no manifest mapping, or a block
  marker is unterminated or nested, the emitter shall fail emission naming the file and the token
  or marker.
- **REQ-AC4-005** At M1, the C2 and C3 files of the twelve shall equal their bytes at the baseline.
- **REQ-AC4-006** From M2 on, each carried-four definition shall differ from its baseline on exactly
  the 8 lines of §C.1, matched by their old content, and each new line shall contain no `~/`.
- **REQ-AC4-007** From M2 on, C1 shall equal C2 and C3p shall equal C3 for the carried four.
- **REQ-AC4-008** When a rendered file differs from a fresh render, the read-only emission check
  shall fail naming the file, and the check shall leave every rendered file unmodified.
- **REQ-AC4-009** When `make agents-emit-check` passes, it shall print `specs=<n> copies=<k>
  diff=0`.
- **REQ-AC4-010** The commit history shall show the M0 baseline commit, alone, before any commit
  that changes an agent source or a rendered definition, then M1, then M2.
- **REQ-AC4-011** The roster guard's numeral layer shall report no undischarged hit for the agent
  source directory.
- **REQ-AC4-012** The emission test suite shall keep every test and assertion it holds at the
  baseline, the Markdown identity assertion being replaced by equality with a fresh render.
- **REQ-AC4-013** In the commit that makes the renderer generate C2 (M1), `AGENTS.local.md` §2.0
  shall have only its sentences stating that C1 differs from C2 or that C2 is hand-edited replaced
  by the rule: the agent source is the sole edit surface, C2 and C3 are rendered from it, C1 equals
  C2 for generated agents, and hand-editing C2 turns `make build` red; no other line of the file
  changes, and the commit message cites `d-20261018T062320Z-5952`.

### §C.1 The 8 line changes (carried four)

Old = the C2 line at the baseline (matched by content, not line number); new = the text shown. Rule:
the new line is the project-copy (C1) line, except e2e-tester:143, where C1 itself holds `~/` and
the new line drops it (which also fixes the shipped `~/~/.claude/` defect). If any of the 8 old
lines no longer matches at M0, M0 aborts and the pairs are recomputed and re-audited; a change to
any other line of the carried four is preserved and the expected result is recomputed at M0 by the
documented command (`acceptance.md` §3).

1. builder-harness:38 — old has `~/.claude/agents/` inside the table cell "`.claude/agents/` or `~/.claude/agents/` (with `--moai`)"; new: `~/.claude/agents/` replaced by `.claude/agents/moai/`.
2. builder-harness:155 — old: `` `~/.claude/agents/manager-spec.md` for the verbatim line: ``; new: `` `.claude/agents/moai/manager-spec.md` for the verbatim line: ``.
3. e2e-tester:92 — old holds `~/.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md`; new: the same sentence with `.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md`.
4. e2e-tester:143 — old holds `~/~/.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md`; new: `.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md`.
5. manager-design:50 — old: `` `~/.claude/skills/moai/workflows/design.md` (D1-D5 step headings). Summary: ``; new: the same with `.claude/skills/moai/workflows/design.md`.
6. manager-design:203 — old: `- **Design pipeline skill**: ` + `~/.claude/skills/moai/workflows/design.md` + ` (D1-D5).`; new: with `.claude/skills/moai/workflows/design.md`.
7. manager-design:206 — old: `- **Agent catalog**: ` + `CLAUDE.md` + ` § 4 (13 retained agents — manager-design is row 6 of the Selection Decision Tree).`; new: `- **Agent catalog**: ` + `.moai/config/sections/delegation.yaml` + ` (13 retained agents — manager-design is row 6 of the Selection Decision Tree).`
8. super-advisor:129 — old: `- Advisor/Evaluator separation: ` + `~/.claude/agents/{plan-auditor,sync-auditor}.md` + `.`; new: `.claude/agents/moai/{plan-auditor,sync-auditor}.md`.

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
  (user-owned; only the §2.0 sentences named in REQ-AC4-013 change).
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
| C1 of the eight | piece 2 | C1 of the eight stays hand-managed until piece 2 absorbs them (decision Q20) |
| Q2, Q3, Q5-Q15 | pieces 2, 3 | `decision-index.md` hand-off rows |

## §F HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-10-18 | Initial unsplit plan (commit `b4ea2d6c93`). |
| 0.2.0 | 2026-10-18 | Narrowed to piece 1 after plan-audit iter1 FAIL (commit `8260bbaa0b`). |
| 0.3.0 | 2026-10-18 | Plan-audit iter2 FAIL fixes: Tier L, all twelve sources at M1, both renderers read the source, pinned line pairs, path-based ordering ACs, decisions Q17 and Q18. |
| 0.4.0 | 2026-10-18 | Plan-audit iter3 FAIL fixes: sources are verbatim shipped files, baseline read from git instead of frozen hashes, ordering ACs bound to the recorded M0 commit, verdict section §G; operator decision `d-20261018T062320Z-5952` closes Q4 and Q17; REQ-AC4-013 added. |

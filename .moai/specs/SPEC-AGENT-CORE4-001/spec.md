---
id: SPEC-AGENT-CORE4-001
title: "Agent core four, piece 1: agent sources, two renders, generated copies"
version: "0.6.0"
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
- **Baseline** — the tree at the parent of the M0 commit (`<M0>^`). M0 absorbs local main first, so
  foreign commits on the guarded paths are inside the baseline. The M0 SHA is recorded once, by the
  M1 run, in `progress.md` §E.2. There is no frozen hash file and no permanent baseline test.
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
to the source layer when they land after piece 1. Leader ruling `d-20261018T070809Z-924f` (leader under operator
delegation, interpreting `d-20261018T062320Z-5952`): C1 = C2 applies to the four carried agents in piece 1 and to the
remaining agents when piece 2 absorbs or retires them — sequencing, not narrowing.

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
- **REQ-AC4-005** At M1, the twelve agent sources shall equal their baseline C2 files and the M1
  commit shall change no shipped C2 or C3 byte, verified once by the M1 acceptance cell.
- **REQ-AC4-006** At M2, each carried-four definition shall differ from its baseline on exactly the
  8 lines of §C.1, and each new line shall contain no `~/`, verified once by the M2 acceptance cell.
- **REQ-AC4-007** From M2 on, C1 shall equal C2 and C3p shall equal C3 for the carried four.
- **REQ-AC4-008** When a rendered file differs from a fresh render, the read-only emission check
  shall fail naming the file, and the check shall leave every rendered file unmodified.
- **REQ-AC4-009** When `make agents-emit-check` passes, it shall print `specs=<n> copies=<k>
  diff=0`.
- **REQ-AC4-010** The roster guard's numeral layer shall report no undischarged hit for the agent
  source directory.
- **REQ-AC4-011** The emission test suite shall keep every test and assertion it holds at the
  baseline, the Markdown identity assertion being replaced by equality with a fresh render.
- **REQ-AC4-012** In the M2 commit, `AGENTS.local.md` shall have exactly the eight lines of §C.2
  replaced by their stated new text and no other line changed, and the M2 commit message shall cite
  board `d-20261018T062320Z-5952`.

After landing, no frozen-equality test exists: a legitimate later edit of an agent goes to the agent
source, `make agents-emit` regenerates the copies, and the permanent guard is
`make agents-emit-check` (REQ-AC4-008/009). No existing test is skipped or relaxed by this
design (REQ-AC4-011).

### §C.1 The 8 line changes (carried four)

Literal lines; line numbers are those of the baseline C2 file and matching is by content. New line =
the project-copy (C1) line, except e2e-tester:143, where C1 itself holds `~/` and the new line drops
it (which also fixes the shipped `~/~/.claude/` defect). If any old line no longer matches at M0,
M0 aborts and the pairs are recomputed and re-audited.

1. `builder-harness.md` line 38
```text
- | `agent` | `.claude/agents/` or `~/.claude/agents/` (with `--moai`) | Frontmatter: name, description, tools, permissionMode, memory, skills (no `model` / `effort`) |
+ | `agent` | `.claude/agents/` or `.claude/agents/moai/` (with `--moai`) | Frontmatter: name, description, tools, permissionMode, memory, skills (no `model` / `effort`) |
```
2. `builder-harness.md` line 155
```text
- `~/.claude/agents/manager-spec.md` for the verbatim line:
+ `.claude/agents/moai/manager-spec.md` for the verbatim line:
```
3. `e2e-tester.md` line 92
```text
- Native desktop toolkits (AppKit, WinUI/Win32, Qt, GTK) are automated through the host OS accessibility layer. The per-OS recipes (macOS / Windows / Linux — defaults, fallbacks, install and probe commands, permission prerequisites) and the desktop-native evidence-source + token-cost ordering live in `~/.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md`; load that reference before any desktop-native work (§ Conditional Skill Loading). Only the recipe matching the HOST OS is probed and executed — state a host-OS/target-OS mismatch in the report instead of probing it. Scripts and flows live under `e2e/desktop-native/`; AX-tree snapshots and run logs ride the existing `e2e/.runs/` timestamped-log convention.
+ Native desktop toolkits (AppKit, WinUI/Win32, Qt, GTK) are automated through the host OS accessibility layer. The per-OS recipes (macOS / Windows / Linux — defaults, fallbacks, install and probe commands, permission prerequisites) and the desktop-native evidence-source + token-cost ordering live in `.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md`; load that reference before any desktop-native work (§ Conditional Skill Loading). Only the recipe matching the HOST OS is probed and executed — state a host-OS/target-OS mismatch in the report instead of probing it. Scripts and flows live under `e2e/desktop-native/`; AX-tree snapshots and run logs ride the existing `e2e/.runs/` timestamped-log convention.
```
4. `e2e-tester.md` line 143
```text
- - When the detected project type is `desktop-native`, read `~/~/.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md` for the per-OS accessibility recipes before probing.
+ - When the detected project type is `desktop-native`, read `.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md` for the per-OS accessibility recipes before probing.
```
5. `manager-design.md` line 50
```text
- `~/.claude/skills/moai/workflows/design.md` (D1-D5 step headings). Summary:
+ `.claude/skills/moai/workflows/design.md` (D1-D5 step headings). Summary:
```
6. `manager-design.md` line 203
```text
- - **Design pipeline skill**: `~/.claude/skills/moai/workflows/design.md` (D1-D5).
+ - **Design pipeline skill**: `.claude/skills/moai/workflows/design.md` (D1-D5).
```
7. `manager-design.md` line 206
```text
- - **Agent catalog**: `CLAUDE.md` § 4 (13 retained agents — manager-design is row 6 of the Selection Decision Tree).
+ - **Agent catalog**: `.moai/config/sections/delegation.yaml` (13 retained agents — manager-design is row 6 of the Selection Decision Tree).
```
8. `super-advisor.md` line 129
```text
- - Advisor/Evaluator separation: `~/.claude/agents/{plan-auditor,sync-auditor}.md`.
+ - Advisor/Evaluator separation: `.claude/agents/moai/{plan-auditor,sync-auditor}.md`.
```

### §C.2 The `AGENTS.local.md` §2.0 rewrite (lands with M2)

Run-phase note: the committed text reads "The only hand-edited layer" on the Source row while the C1 row says C1 of the eight stays hand-edited; read the Source row as "for the generated agents". The pairs below are the text as committed (AC-AC4-012 compares against them); the file itself is user-owned and a wording fix is a follow-up outside this SPEC.

Exactly these eight lines of the file at the baseline change (line numbers are those at `06095bb8fa`;
matching is by content). Line 98 becomes two lines. Everything else stays byte-identical:
`git diff --numstat` of the M2 commit for this file prints `9	8	AGENTS.local.md`. The rule applies to the
generated agents (the carried four); the remaining agents join when piece 2 absorbs or retires them.

- line 91
```text
- ### [HARD] §2.0 There are three copies of the agent definitions, and the third is not edited by hand
+ ### [HARD] §2.0 There are three copies of the agent definitions, and only the agent source is edited by hand (C1 of the eight agents not yet generated is the exception until piece 2)
```
- line 93
```text
- Agent definitions exist in three copies, but **only one of them is not a hand-edit target.**
+ Agent definitions exist in three copies, rendered from one edit surface: the agent source layer `internal/template/agentemit/source/*.md`. **The agent source is the sole edit surface**; none of the three copies is a hand-edit target for the generated agents.
```
- line 97
```text
- | C1 | `.claude/agents/moai/*.md` | Local dogfooding, hand-edited |
+ | C1 | `.claude/agents/moai/*.md` | Local dogfooding; rendered from the agent source and byte-identical to C2 for the generated agents (builder-harness, e2e-tester, manager-design, super-advisor); hand-edited for the remaining agents until piece 2 absorbs or retires them |
```
- line 98
```text
- | C2 | `internal/template/templates/.claude/agents/moai/*.md` | Distribution mirror, hand-edited — **the neutral source layer** |
+ | Source | `internal/template/agentemit/source/*.md` | **The only hand-edited layer** |
+ | C2 | `internal/template/templates/.claude/agents/moai/*.md` | Distribution mirror, **rendered from the agent source** |
```
- line 99
```text
- | C3 | `internal/template/templates/.codex/agents/moai/*.toml` | **Machine-emitted from C2** (`internal/template/agentemit`) |
+ | C3 | `internal/template/templates/.codex/agents/moai/*.toml` | **Machine-rendered from the agent source** (`internal/template/agentemit`) |
```
- line 101
```text
- [HARD] **If you have edited `internal/template/templates/.claude/agents/moai/*.md`, run `make agents-emit`.** Otherwise `internal/template/templates/.codex/agents/moai/*.toml` stays stale, and a binary built in that state runs with the old definitions embedded. C1 ↔ C2 is **not** a byte-identical relationship (an intentional divergence); the generation relationship runs in one direction only, C2 → C3.
+ [HARD] **If you have edited the agent source `internal/template/agentemit/source/*.md`, run `make agents-emit`.** Hand-editing C2 turns `make build` red through `agents-emit-check`. Otherwise `internal/template/templates/.codex/agents/moai/*.toml` stays stale, and a binary built in that state runs with the old definitions embedded. C1 is byte-identical to C2 for the generated agents (builder-harness, e2e-tester, manager-design, super-advisor); the remaining agents join when piece 2 absorbs or retires them. The generation relationship runs in one direction only, source → C1, C2, C3.
```
- line 103
```text
- [HARD] **Do not hand-edit C3.** A hand edit is silently overwritten at the next emission. If something needs fixing, fix C2 and regenerate.
+ [HARD] **Do not hand-edit C3.** A hand edit is silently overwritten at the next emission. If something needs fixing, fix the agent source and regenerate.
```
- line 109
```text
- | `make build` (preceded by `agents-emit-check`) | Source layer: committed `.toml` vs the output emitted from `.md` | Every local build. **Read-only — does not regenerate** |
+ | `make build` (preceded by `agents-emit-check`) | Source layer: committed `.toml` vs the output rendered from the agent source | Every local build. **Read-only — does not regenerate** |
```

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
  (user-owned; only the §2.0 lines named in REQ-AC4-012 change).
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
| Baseline tests with a literal 12 (`!= 12` in `golden_test.go`, `len(roles) != 12` in `permission_contract_test.go`) and the subtests `twelve_sources_twelve_definitions`, `contract_rows_unchanged_for_twelve` | piece 2 | REQ-AC4-011 forbids changing them in piece 1; piece 2 retargets them when it retires the eight, without asserting a literal 12 |
| Frozen C3 header `# Neutral source: .claude/agents/moai/<name>.md (body carried verbatim).` in every C3 file | piece 2 | the wording predates the agent source layer; correct it in the emitter when the eight are absorbed |
| C1 of the eight | piece 2 | C1 of the eight stays hand-managed until piece 2 absorbs or retires them (sequencing per board d-20261018T070809Z-924f, decision Q20); the card completion "copy diff 0" for them is piece 2's |
| Q2, Q3, Q6-Q11 (open); Q5, Q12-Q15 (withdrawn, re-open where needed) | pieces 2, 3 | `decision-index.md` hand-off rows |

## §F HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-10-18 | Initial unsplit plan (commit `b4ea2d6c93`). |
| 0.2.0 | 2026-10-18 | Narrowed to piece 1 after plan-audit iter1 FAIL (commit `8260bbaa0b`). |
| 0.3.0 | 2026-10-18 | Plan-audit iter2 FAIL fixes: Tier L, all twelve sources at M1, both renderers read the source, pinned line pairs, path-based ordering ACs, decisions Q17 and Q18. |
| 0.4.0 | 2026-10-18 | Plan-audit iter3 FAIL fixes: sources are verbatim shipped files, baseline read from git instead of frozen hashes, ordering ACs bound to the recorded M0 commit, verdict section §G; operator decision `d-20261018T062320Z-5952` closes Q4 and Q17; REQ-AC4-013 added. |
| 0.5.0 | 2026-10-18 | Plan-audit iter4 FAIL (STOP) fixes under rule C final delta (d-20261018T070809Z-924f): git-history ordering cells and permanent baseline tests removed (one-time M1/M2 cells instead), AGENTS.local.md §2.0 rewrite moved to M2 with exact text, Q20 decided by the board. |
| 0.6.0 | 2026-10-18 | Last round (board d-20261018T072844Z-765e, operator): phrase-level banned list for the permanent §2.0 check, whole-line comparison made a one-time M2 cell, heading qualified for the eight not yet generated. |

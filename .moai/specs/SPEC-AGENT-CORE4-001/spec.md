---
id: SPEC-AGENT-CORE4-001
title: "Agent core four, piece 1: neutral agent specs, two renders, generated copies"
version: "0.2.0"
status: draft
created: 2026-10-18
updated: 2026-10-18
author: "Goos Kim"
priority: P1
phase: "v3.2.0 target"
module: "internal/template/agentemit"
lifecycle: spec-anchored
tags: "agents, neutral-spec, codex-render, generated-copies"
tier: M
related_specs: [SPEC-CODEX-DUAL-AGENTS-001, SPEC-AGENT-MODEL-INHERIT-001]
---

# SPEC-AGENT-CORE4-001 — Piece 1: neutral layer, renderers, generated copies

## §A Background

Twelve agent definitions ship in four hand-managed copies that drift: 7 of 12 project/template
Markdown pairs differ and the project Codex mirror holds 11 TOML files against the template's 12
(ledger L-02..L-03 in `acceptance.md`). Piece 1 of the agent-core-four work introduces one neutral
source per definition and renders the template Markdown (C2) and template Codex (C3) from it, and
for the four agents that stay in the roster (builder-harness, e2e-tester, manager-design,
super-advisor) also the project Markdown (C1) and project Codex (C3p). It changes no agent
behavior: every shipped byte of the other eight definitions stays as it is.

The 2026-10-18 plan audit (iter1, FAIL 0.66, file `.moai/reports/t1649/plan-audit.md`, local and
untracked) found the original single SPEC unviable under the plan-audit ceiling; the leader ruled a
split. This SPEC is piece 1 (former M0-M2). Evidence from the original plan is preserved in
commit `b4ea2d6c93` (`research.md`, `design.md` at that commit).

## §B Definitions

- **C1** `.claude/agents/moai/*.md`; **C2** `internal/template/templates/.claude/agents/moai/*.md`;
  **C3** `internal/template/templates/.codex/agents/moai/*.toml`; **C3p** `.codex/agents/moai/*.toml`.
- **The eight** — manager-spec, manager-develop, manager-docs, manager-git, manager-lead,
  manager-todo, plan-auditor, sync-auditor. **The carried four** — builder-harness, e2e-tester,
  manager-design, super-advisor.
- **Neutral source** — `internal/template/agentemit/neutral/<name>.md`: a frontmatter block in
  neutral capability tokens plus the body. A build input beside `agents-codex.yaml`; it lives
  outside `templates/` and is never distributed.
- **Pinned HEAD bytes** — sha256 of each C2 and C3 file at tree `b4ea2d6c93`, stored in
  `internal/template/agentemit/testdata/agents-head.sha256` (tracked, M0).

## §C Requirements (GEARS)

- **REQ-AC4-001** The agent definition set shall have exactly one neutral source per agent
  definition and no rendered C2 or C3 file without one.
- **REQ-AC4-002** The Claude renderer shall produce each C2 definition deterministically from its
  neutral source and, when a capability token has no mapping, shall fail emission naming the file
  and the token.
- **REQ-AC4-003** The Codex renderer shall produce each C3 definition from the same neutral source
  with the existing mapping manifest, leaving every permission-contract row for the twelve roles
  unchanged.
- **REQ-AC4-004** When the eight are rendered from their neutral sources, the C2 and C3 outputs
  shall be byte-identical to the pinned HEAD bytes.
- **REQ-AC4-005** The neutral bodies of the carried four shall name skills and files without an
  install-root prefix, and rendering them shall change their shipped bytes only on a pinned list of
  lines.
- **REQ-AC4-006** The rendered C1, C2, C3 and C3p files of the carried four shall be pairwise
  identical per harness (C1 = C2, C3p = C3).
- **REQ-AC4-007** When a rendered file differs from a fresh render, the read-only emission check
  shall fail naming the file, and shall otherwise print a summary line counting the specs and the
  copies compared.
- **REQ-AC4-008** The baseline measurement and the pinned HEAD bytes shall be committed as tracked
  files in a commit that precedes every commit changing a rendered or neutral file.

## §D Out of Scope — follow-on pieces

### Out of Scope — piece 2 (retirement, guards, replacement paths)

Former M3, M4, M5, M7, M8, queued by the leader after piece 1 lands, as a separate SPEC without an
id yet: the core four agents (planner, implementer, documenter, auditor) and worker, the size verb,
the auditor verdict-only write guard, the agent memory cap, spawn-record enrichment, replacement
paths for retiring manager-lead (a duty-by-duty map of the leader working mode), manager-todo and
manager-git (including a non-factory owner for `delivery-policy.md` rule 1 and the `OwnerRole`
value at `internal/mission/integration.go:38`), the reference flip over 977 files, the retired-name
refusal, cutover, and `auditor` deferred emission.

### Out of Scope — piece 3 (list hiding)

Former M6, a separate SPEC without an id yet, queued after piece 2: the `moai-agent-*` hiding
settings for Claude and Codex, the preload probe, and the doctor check.

### Out of Scope — other

- docs-site and README pages, user-scope `~/.claude/agents`, and the `AGENTS.local.md` §2.0
  sentence about C1/C2 divergence (user-owned; amendment named in decision Q4).
- Any change of agent bodies, tools, model, effort or permission mode beyond the pinned line list.

## §E Hand-off table (audit iter1 defects not fixed here)

| Defect | Owner | Content to inherit |
|---|---|---|
| D2 guard root | piece 2 | root = git toplevel of the acting agent's cwd; subtests for linked worktree and primary-checkout denial; `internal/hook/path_resolve.go:69-76` order |
| D4 refusal predicate | piece 2 | active only when name in table AND no `<name>.md` in the project's agents directory |
| D5 replacement paths | piece 2 | leader duty map, non-factory manager-git owner, `moai-factory` skill waiver/sequencing |
| D7 skill and reference publication | pieces 2, 3 | REQ/AC for deployed location in both harnesses, catalog and embed registration, `on_demand_reference_bytes` |
| D8 phased rewrite vs `auditor` emission | piece 2 | move rewrite into the emission commit or add a spawn-resolution gate |
| D12 numeral layer, 33 count rows, threshold at 17 | piece 2 | `threshold_17_is_12`, stale-subset fixture |
| D13 shrinking edit on oversize file | piece 2 | allow when result is smaller than the current file |
| D14 core-body preservation | piece 2 | pinned [HARD]-clause list; implementer budget row |
| D15 "never blocks a spawn" statements | piece 2 | rewrite sentence and file header; deny in a separate handler file |
| D18 doctor check name | piece 3 | registered check is "Skills Allowlist" (`internal/cli/doctor.go:306`) |
| Q5, Q12-Q15 (unresolved defaults of the original index) | pieces 2, 3 | see `decision-index.md` hand-off rows |

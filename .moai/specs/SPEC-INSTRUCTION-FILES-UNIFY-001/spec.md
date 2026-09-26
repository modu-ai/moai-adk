---
id: SPEC-INSTRUCTION-FILES-UNIFY-001
title: Instruction-file unification — AGENTS.md as the harness-neutral contract
version: 0.2.0
status: draft
priority: P1
phase: "v3.3.0 target"
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
category: harness
tags: [instruction-files, codex, claude, template, migration]
tier: L
---

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-26 | Initial plan-phase draft. Encodes the operator-approved design decision 2 (lead report 2026-09-26) plus the M0 measurement results recorded at `.moai/reports/t1243/m0/verdict.md`. |
| 0.2.0 | 2026-09-26 | Lead directives 1 and 2 applied: the ancestor-discovery observation is labelled unconfirmed and may not serve as a premise; the `.tmpl` invariant (card t925) and the nested-sum 32,768-byte budget added as REQ-IFU-024 / REQ-IFU-025; worktree duplicate-load scoped out to t1219. |

---

## §A Context

The repository currently carries four instruction surfaces with overlapping and partly
contradictory ownership:

- `AGENTS.md` — deployed, read natively by Codex, imported by `CLAUDE.md`. The root copy
  carries 8 sections; the template copy (`internal/template/templates/AGENTS.md.tmpl`)
  carries 12. The two have diverged.
- `CLAUDE.md` — deployed, read by Claude Code, already importing `@AGENTS.md`, but still
  carrying a second inline layer of contract material.
- `CLAUDE.local.md` — user-owned, gitignored, read by Claude Code through
  ancestor-directory discovery. In this repository it is 61,908 bytes and git-tracked
  despite being gitignored.
- `AGENTS.local.md` — already named in `internal/cli/codex_contract.go` as the Codex-side
  local instruction input, injected through the launcher's `developer_instructions` key,
  and already present in both `.gitignore` copies. It is not yet the primary read path
  and is not imported by any deployed file.

The consequence is that a maintainer editing "the contract" must decide which of four
files to edit, and a Codex session and a Claude session do not read the same set.

This SPEC unifies the surface onto three files with one owner each, and retires
`CLAUDE.local.md` after a one-minor-release transition window.

---

## §B Goal

One harness-neutral contract (`AGENTS.md`), one thin harness-mechanism link file per
harness (`CLAUDE.md`), and one user-owned local file (`AGENTS.local.md`) that both
harnesses read.

---

## §C Requirements (GEARS)

### C.1 File structure and ownership

- **REQ-IFU-001** — The system shall deploy `AGENTS.md` as a harness-neutral contract file
  that is overwritten by `moai update` and is not a user-edited surface.
- **REQ-IFU-002** — The system shall deploy `CLAUDE.md` as a thin link file carrying, in
  order: an `@AGENTS.md` import, a Claude-only mechanism layer, and an `@AGENTS.local.md`
  import as its final import.
- **REQ-IFU-003** — The system shall not deploy `AGENTS.local.md`; it is a user-owned file,
  absent from `internal/template/templates/`, and listed in the deployed `.gitignore`.
- **REQ-IFU-004** — Where `AGENTS.local.md` is absent, `moai init` and `moai update` shall
  complete successfully and the deployed `CLAUDE.md` shall retain its unresolved
  `@AGENTS.local.md` import line unchanged.

### C.2 Per-harness read paths

- **REQ-IFU-005** — While a Claude Code session is reading project instructions, the system
  shall reach `AGENTS.md` and `AGENTS.local.md` through the `CLAUDE.md` imports alone, with
  no launcher change in `moai cc`, bare `claude`, or `moai glm`.
- **REQ-IFU-006** — When the `moai codex` launcher assembles `developer_instructions`, it
  shall read `AGENTS.local.md` ahead of `CLAUDE.local.md`.
- **REQ-IFU-007** — Where only `CLAUDE.local.md` exists, the `moai codex` launcher shall
  read it as a fallback and shall emit a deprecation advisory naming the migration command.
- **REQ-IFU-008** — The Codex provenance preamble shall name the literal filename of the
  file it actually read, never a normalized or substituted name.

### C.3 Migration

- **REQ-IFU-009** — The system shall provide an explicit migration verb
  (`moai migrate local-instructions`) that moves `CLAUDE.local.md` content to
  `AGENTS.local.md` and relocates the original to a backup directory.
- **REQ-IFU-010** — `AGENTS.local.md` and `CLAUDE.local.md` shall not coexist as live read
  paths; the migration verb shall leave exactly one of the two in place.
- **REQ-IFU-011** — `moai update` shall not move, rename, or delete either local
  instruction file; where both exist, it shall emit an advisory only.
- **REQ-IFU-012** — `moai doctor` shall report the same advisory as `moai update` when both
  local instruction files exist or when only `CLAUDE.local.md` exists.

### C.4 Guard and learner surfaces

- **REQ-IFU-013** — The harness-learner frozen-instruction set shall include
  `AGENTS.md` and `AGENTS.local.md` in addition to the existing `CLAUDE.md` and
  `CLAUDE.local.md` entries.
- **REQ-IFU-014** — The harness curator's Tier-3 learning-record target shall be
  `AGENTS.local.md`.
- **REQ-IFU-015** — Where a template file is scanned for neutrality, `AGENTS.local.md` shall
  be an allowed reference and `CLAUDE.local.md` shall remain forbidden.

### C.5 Contract body

- **REQ-IFU-016** — The deployed `AGENTS.md` body shall be harness-neutral: it shall carry
  no clause whose applicability is restricted to one harness.
- **REQ-IFU-017** — The `AGENTS.md` preamble shall state that the instruction budget is
  charged against project instruction files only, and shall retain the statement that
  overflow is truncated silently from the tail.
- **REQ-IFU-018** — The root `AGENTS.md` and `internal/template/templates/AGENTS.md.tmpl`
  shall carry the same `## ` section set after reconciliation. Reconciliation applies to the
  section set only, never to the filename (REQ-IFU-024) and never to file content — the two
  mirrors diverge by 46 intentional lines.
- **REQ-IFU-019** — Each deployed contract document shall not exceed the per-file ceiling of
  24,576 bytes (`CodexContractByteCeiling`, `internal/config/token_budget_guard.go:101`).
- **REQ-IFU-024** — The template mirror of the contract shall retain a filename Codex does
  not discover (`AGENTS.md.tmpl`); the system shall not place a file named `AGENTS.md` under
  `internal/template/templates/`. Reversing card t925 reintroduces a measured silent
  truncation that drops a section and cuts the preceding one mid table row.
- **REQ-IFU-025** — The sum of the instruction files Codex discovers by filename in any
  nested chain within the repository shall not exceed 32,768 bytes (Codex's measured
  `project_doc_max_bytes` default). Raising that limit shall not be used as a remedy: the
  project-scope override is silently ignored until the user registers
  `trust_level = "trusted"`, and a distributed user's first session is untrusted by
  construction.

### C.6 Documentation and this repository

- **REQ-IFU-020** — The docs-site pages `claude-md-guide`, `codex-dual-harness`,
  `harness-learning`, `memory`, `quickstart`, and `update` shall describe the three-file
  structure in all four locales (ko, en, ja, zh).
- **REQ-IFU-021** — This repository's own `CLAUDE.local.md` shall migrate to
  `AGENTS.local.md` at under 40,000 characters, with operational procedure relocated to
  `.moai/docs/`.
- **REQ-IFU-022** — The migrated `AGENTS.local.md` §0 shall name `AGENTS.local.md` as the
  file whose canonical copy it discriminates.

### C.7 Verification of import resolution

- **REQ-IFU-023** — The system shall carry a mechanical check that the `@AGENTS.local.md`
  import in a deployed `CLAUDE.md` actually resolved, rather than inferring resolution from
  the absence of an error. (M0-1: an unresolved import is silently skipped — exit 0, empty
  stderr.)

Every check discharging REQ-IFU-019, REQ-IFU-023, and REQ-IFU-025 shall carry a positive
indicator. Where the failure mode is silent truncation or a silent skip, the absence of an
error is not evidence of success.

---

## §D Out of Scope

This SPEC deliberately does not build the following.

### Out of Scope — instruction-file mechanisms

- `AGENTS.override.md` in any role. Codex finds it ahead of `AGENTS.md` in the same folder,
  so it can shadow the deployed contract, and Claude does not read it at all.
- Claude's `claude-md-and-agents-md` project-instructions setting. It is ignored in project
  and local settings, honored only in user and managed settings, and does not cover local
  files.

### Out of Scope — automation

- Automatic renaming of `CLAUDE.local.md` by `moai update`, `moai init`, or any hook.
  Migration is an explicit operator-invoked verb only.
- Automatic restoration or repair of a modified local instruction file.

### Out of Scope — the worktree duplicate-load

- Deduplicating the local-instruction load in a worktree session. Open card **t1219 item (1)**
  already covers it (a worktree session loading two copies of `CLAUDE.local.md` — worktree
  44.4k chars, primary 39.3k chars, differing content, roughly 30k extra tokens). If the M1
  re-measurement confirms ancestor discovery, the mechanism that would carry local
  instructions into a worktree is the same mechanism that double-loads them; that finding is
  handed to t1219 with its evidence rather than resolved here.
- What this SPEC does owe is not making the duplicate worse: REQ-IFU-010's no-coexistence
  invariant is what keeps a worktree session from carrying up to four local-instruction loads.

### Out of Scope — adjacent work

- SPEC A (codex factory retirement). It overlaps this SPEC at `AGENTS.md` §3 and §8 and is
  sequenced separately; see plan.md §Risks.
- Rewriting the substance of any contract clause. This SPEC relocates and reconciles
  clauses; it does not re-decide them.
- Codex-side measurement of `developer_instructions` behavior beyond what the existing
  test suite already asserts and what REQ-IFU-025's discovery measurement requires.

---

## §E Cross-references

- `.moai/reports/t1243/m0/verdict.md` — the M0 measurement this SPEC's design rests on.
- `internal/cli/codex_contract.go`, `internal/cli/codex_launcher.go` — read order.
- `internal/hook/pre_tool.go` `frozenInstructionFiles` — the guard set.
- `internal/harness/curator/dispatch.go` — the Tier-3 target.
- `.moai/docs/template-internal-isolation-doctrine.md` — the neutrality content classes.

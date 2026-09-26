---
id: SPEC-INSTRUCTION-FILES-UNIFY-001
title: "Instruction-file unification — AGENTS.md as the harness-neutral contract"
version: "0.3.1"
status: draft
priority: P1
phase: "v3.3.0 target"
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
module: "internal/cli, internal/config, internal/hook, internal/harness/curator, internal/template/templates"
lifecycle: spec-anchored
tags: "instruction-files, codex, claude, template, contract, guards"
tier: L
---

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-26 | Initial plan-phase draft. Encodes the operator-approved design decision 2 (lead report 2026-09-26) plus the M0 measurement results recorded at `.moai/reports/t1243/m0/verdict.md`. |
| 0.2.0 | 2026-09-26 | Lead directives 1 and 2 applied: the ancestor-discovery observation is labelled unconfirmed and may not serve as a premise; the `.tmpl` invariant (card t925) and the nested-sum 32,768-byte budget added as REQ-IFU-024 / REQ-IFU-025; worktree duplicate-load scoped out to t1219. |
| 0.3.0 | 2026-09-26 | **B1/B2 carve, by operator decision.** The nine requirements `REQ-IFU-007~012` and `REQ-IFU-020~022` — everything touching a user-owned file — transferred verbatim to `SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001` (card t1259 owns its plan phase). This SPEC retains 16: `REQ-IFU-001~006`, `013~019`, `023~025`. **The carve's cause is arithmetic, not the recorded debt:** the plan-audit of commit `1140bcd1d` found `REQ-IFU-001` and `REQ-IFU-005` covered by no criterion, the fix needs two new criteria, and the SPEC stood at the Tier L ceiling of 25/25 with no tier above L. Also applied from that audit: frontmatter repaired to the canonical 12-field schema (MP-3), five vacuously-passable `go test -run` patterns repaired and given positive indicators (D1), two new criteria added for the uncovered requirements (D2), `AC-IFU-022`'s positive control defined (D5), the `AGENTS.md` mirror-divergence figure re-measured (D4), `plan.md` §G's mis-citation corrected (D6), and §C reordered so requirement ids ascend in document order (D8). |
| 0.3.1 | 2026-09-26 | **Plan-audit iter-2 repairs (subject `653e53572`, PASS-WITH-DEBT 0.853 with two blocking defects).** D9 (critical): the D1 vacuity class survived in `AC-IFU-010` and `AC-IFU-012`, both blocking, both passing against the unimplemented tree by prefix-substring accident; repaired as a **class** — both-end pattern anchoring plus a trailing-space delimiter on every asserted `--- PASS:` line, the rule stated at the head of `acceptance.md` with a passable enumeration command, and applied additionally to `AC-IFU-016`, `AC-IFU-025`, and the sibling SPEC's `AC-IFU-011`. `AC-IFU-010` additionally had the wrong assertion (it required `BETA` absent, i.e. exclusive precedence, where `REQ-IFU-006` and design.md §C specify read **order**); corrected to ordering. D12 (major): three sibling-SPEC `AC-` tokens in prose made the AC-count guard read 23 live criteria against 20 declared — marked `[REF]`, counter now reports 20. D10 stale `M4` citation → `M3`. D11 design.md §C annotated so `REQ-IFU-007`/`008` read as the sibling's. D13 plan.md M3 now names the package-scoped always-loaded invocation. Also: §D.2 derivation made one-way and the set-diff command's blindness to row pairing stated; two close items added. |

> **[HARD] The id gaps in this SPEC are the carve's footprint, not an error.** `REQ-IFU-007~012`
> and `REQ-IFU-020~022` are absent here because they live in
> `SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001`, and the criteria that covered them went with them. The
> ids were deliberately NOT renumbered: a non-contiguous sequence is far cheaper than breaking
> every cross-reference, the traceability table, and the audit history that already cites these
> ids by name. A later reader who "fixes" the gaps breaks all three.

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
  ancestor-directory discovery. In this repository it is git-tracked despite being
  gitignored.
- `AGENTS.local.md` — already named in `internal/cli/codex_contract.go` as the Codex-side
  local instruction input, injected through the launcher's `developer_instructions` key,
  and already present in both `.gitignore` copies. It is not yet the primary read path
  and is not imported by any deployed file.

The consequence is that a maintainer editing "the contract" must decide which of four
files to edit, and a Codex session and a Claude session do not read the same set.

This SPEC establishes the three-file structure and the guards that hold it. The user-owned
half of the transition — the Codex fallback advisory, the migration verb, the `moai update` /
`moai doctor` advisories, the docs-site rewrite, and this repository's own migration — is
`SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001`.

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

### C.3 Guard and learner surfaces

- **REQ-IFU-013** — The harness-learner frozen-instruction set shall include
  `AGENTS.md` and `AGENTS.local.md` in addition to the existing `CLAUDE.md` and
  `CLAUDE.local.md` entries.
- **REQ-IFU-014** — The harness curator's Tier-3 learning-record target shall be
  `AGENTS.local.md`.
- **REQ-IFU-015** — Where a template file is scanned for neutrality, `AGENTS.local.md` shall
  be an allowed reference and `CLAUDE.local.md` shall remain forbidden.

### C.4 Contract body

- **REQ-IFU-016** — The deployed `AGENTS.md` body shall be harness-neutral: it shall carry
  no clause whose applicability is restricted to one harness.
- **REQ-IFU-017** — The `AGENTS.md` preamble shall state that the instruction budget is
  charged against project instruction files only, and shall retain the statement that
  overflow is truncated silently from the tail.
- **REQ-IFU-018** — The root `AGENTS.md` and `internal/template/templates/AGENTS.md.tmpl`
  shall carry the same `## ` section set after reconciliation. Reconciliation applies to the
  section set only, never to the filename (REQ-IFU-024) and never to file content — the two
  mirrors diverge intentionally (§C.4 note below).
- **REQ-IFU-019** — Each deployed contract document shall not exceed the per-file ceiling of
  24,576 bytes (`CodexContractByteCeiling`, declared in
  `internal/config/token_budget_guard.go`).

> **The intentional mirror divergence, attributed.** Measured in this worktree against base
> develop `553e224f3` on 2026-09-26:
> `diff AGENTS.md internal/template/templates/AGENTS.md.tmpl` reports **57 template-only lines
> (`^>`) and 17 root-only lines (`^<`)**. The figure of "46 intentional lines" carried in the
> v0.2.0 draft was a card-t925-era number (2026-09-18) and does not reproduce against this
> tree; it is retired rather than carried forward. What the figure supports is unchanged at any
> value: the mirrors differ by design, so `AC-IFU-008` compares the `## ` section set and a
> content-level diff would fail by construction.

### C.5 Verification of import resolution

- **REQ-IFU-023** — The system shall carry a mechanical check that the `@AGENTS.local.md`
  import in a deployed `CLAUDE.md` actually resolved, rather than inferring resolution from
  the absence of an error. (M0-1: an unresolved import is silently skipped — exit 0, empty
  stderr.)

### C.6 The template-mirror filename and the nested budget

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

Every check discharging REQ-IFU-019, REQ-IFU-023, and REQ-IFU-025 shall carry a positive
indicator. Where the failure mode is silent truncation or a silent skip, the absence of an
error is not evidence of success.

> **[HARD] "A positive indicator" binds a `go test -run` invocation too.** Go exits `0` when
> `-run` matches no test, printing `no tests to run` — so a criterion naming a pattern that
> matches nothing passes vacuously. Every criterion in `acceptance.md` that invokes a test
> therefore names the test's **symbol** (never a prefix guess), runs with `-v`, and asserts
> `--- PASS: <TestName>` appears in the output. The plan-audit of commit `1140bcd1d` found
> five criteria failing exactly this way, three of which named a guard that already existed
> under a different name.

---

## §D Out of Scope

This SPEC deliberately does not build the following.

### Out of Scope — the user-owned-file half of the transition

- The Codex `CLAUDE.local.md` fallback branch and its deprecation advisory, the
  `moai migrate local-instructions` verb, the no-coexistence invariant, the `moai update` and
  `moai doctor` advisories, the six docs-site pages in four locales, and this repository's own
  `CLAUDE.local.md` migration. All nine requirements and their criteria transferred verbatim
  to **`SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001`** (card **t1259**) at v0.3.0. They are not
  cancelled and not weakened — they are owned elsewhere.

### Out of Scope — instruction-file mechanisms

- `AGENTS.override.md` in any role. Codex finds it ahead of `AGENTS.md` in the same folder,
  so it can shadow the deployed contract, and Claude does not read it at all.
- Claude's `claude-md-and-agents-md` project-instructions setting. It is ignored in project
  and local settings, honored only in user and managed settings, and does not cover local
  files.

### Out of Scope — automation

- Automatic renaming of `CLAUDE.local.md` by `moai update`, `moai init`, or any hook.
  Migration is an explicit operator-invoked verb only, and it belongs to the sibling SPEC.
- Automatic restoration or repair of a modified local instruction file.

### Out of Scope — the worktree duplicate-load

- Deduplicating the local-instruction load in a worktree session. Open card **t1219 item (1)**
  already covers it (a worktree session loading two copies of `CLAUDE.local.md`, differing
  content, roughly 30k extra tokens). If the M1 re-measurement confirms ancestor discovery,
  the mechanism that would carry local instructions into a worktree is the same mechanism
  that double-loads them; that finding is handed to t1219 with its evidence rather than
  resolved here.

### Out of Scope — adjacent work

- SPEC A (codex factory retirement). It overlaps this SPEC at `AGENTS.md` §3 and §8 and is
  sequenced separately; see plan.md §C.
- Rewriting the substance of any contract clause. This SPEC relocates and reconciles
  clauses; it does not re-decide them.
- Codex-side measurement of `developer_instructions` behavior beyond what the existing
  test suite already asserts and what REQ-IFU-025's discovery measurement requires.

---

## §E Cross-references

- `SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001` — the sibling SPEC holding the nine transferred
  requirements (card t1259).
- `.moai/reports/t1243/m0/verdict.md` — the M0 measurement this SPEC's design rests on.
- `.moai/reports/t1243/plan-audit-iter1.md` — the audit of commit `1140bcd1d` that v0.3.0 answers.
- `internal/cli/codex_contract.go`, `internal/cli/codex_launcher.go` — read order.
- `internal/hook/pre_tool.go` `frozenInstructionFiles` — the guard set.
- `internal/harness/curator/dispatch.go` — the Tier-3 target.
- `.moai/docs/template-internal-isolation-doctrine.md` — the neutrality content classes.

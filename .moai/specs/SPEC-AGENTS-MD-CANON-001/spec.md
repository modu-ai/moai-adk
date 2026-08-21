---
id: SPEC-AGENTS-MD-CANON-001
title: "AGENTS.md canonical contract layer for Codex dual-harness"
version: "0.1.0"
status: draft
created: 2026-08-22
updated: 2026-08-22
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".claude/rules, .claude/output-styles, internal/config, internal/template/templates"
lifecycle: spec-anchored
tags: "codex, agents-md, always-loaded, token-budget, dual-harness"
tier: L
---

# SPEC-AGENTS-MD-CANON-001 — AGENTS.md canonical contract layer

## HISTORY

| Date | Version | Change |
|---|---|---|
| 2026-08-22 | 0.1.0 | Initial draft (plan-phase). Card t82, milestone M2 of the Codex dual-harness epic. Baseline numbers taken from `.moai/reports/t82/measurement.md`, not from the card. |

---

## §A. Context

Codex (`codex-cli` 0.147.0) loads project instructions from `AGENTS.md` under a byte cap
(`project_doc_max_bytes`) and truncates the overflow. MoAI's instruction surface is an
order of magnitude larger than that cap, so a naive `AGENTS.md` would be cut mid-rule —
and a rule cut mid-sentence is worse than an absent one, because it reads as complete.

### A.1 Measured baseline (SSOT: `.moai/reports/t82/measurement.md`, measured 2026-08-22)

The measured surface is defined by the guard that already owns it —
`internal/config/token_budget_guard.go` `alwaysLoadedSurface()`: every
`.claude/rules/moai/**/*.md` without `paths:` frontmatter, plus three fixed slots.

| Surface | Bytes |
|---|---:|
| always-loaded rules (14 files, no `paths:`) | 202,621 |
| `CLAUDE.md` | 20,523 |
| `.claude/output-styles/moai/moai.md` | 61,706 |
| repo-root `MEMORY.md` | 0 (absent; guard treats hermetically) |
| **total measured surface** | **284,850** |

Estimated tokens (the guard's `char/4`): **≈ 71,212**. Budget constant
`AlwaysLoadedTokenBudget = 76,000`; headroom ≈ 4,788 tokens; the guard test passes
on this tree today.

Gap versus a 32,768 B document cap: **8.7×**. The card's figures implied 4.7× and omitted
`.claude/output-styles/moai/moai.md` — the single largest always-loaded file — entirely.

### A.2 Measured contract-layer volume (this SPEC's own measurement)

Command:

```
grep -rLE '^paths:' --include='*.md' .claude/rules | sort > /tmp/t82_always.txt
xargs grep -h '\[HARD\]' < /tmp/t82_always.txt | wc -c     # → 30353
grep -h '\[HARD\]' CLAUDE.md | wc -c                        # → 2190
grep -h '\[HARD\]' .claude/output-styles/moai/moai.md | wc -c  # → 11898
```

| Contract slice | Marked lines | Bytes |
|---|---:|---:|
| `[HARD]`-marked lines across the 14 always-loaded rules | 93 | 30,353 |
| `[HARD]`-marked lines in `CLAUDE.md` | 4 | 2,190 |
| **subtotal (Codex-relevant contract, verbatim)** | **97** | **32,543** |
| `[HARD]`-marked lines in `.claude/output-styles/moai/moai.md` | 75 | 11,898 |

Widening to every imperative line (`[HARD]` ∪ `MUST` ∪ `MUST NOT` ∪ `shall`, deduplicated)
raises the rules + `CLAUDE.md` figure to **43,638 B**.

**This is the SPEC's decisive finding.** The card's target — a root `AGENTS.md` of about
8 KiB — is **4.0× smaller than the verbatim contract layer already is** (32,543 B), and the
verbatim contract alone consumes **99.3 %** of a 32,768 B shared budget, leaving roughly
225 B for every nested document combined. The three-way redistribution as stated in the card
is not reachable by moving text; it is reachable only if the contract clauses are *rewritten*
at a measured compression ratio. `plan.md` §E M1 makes establishing that ratio the first milestone,
and its stop condition makes an unreachable ratio a halt rather than a silent overrun.

### A.3 Shared-budget interpretation

The codex 0.147.0 binary carries the loader log `project doc exceeds remaining budget;
truncating` with a `remaining_bytes` field. A *remaining* budget implies **one budget shared
across the merged document chain**, consumed in load order — not a per-file allowance. The
card's layout ("root ~8 KiB + per-area nested ~4 KiB") is compatible with that reading only
while the sum stays under the cap; read as "each nested file gets its own 4 KiB", it is wrong.

### A.4 Relationship to SPEC-ALWAYS-LOADED-DIET-001

That SPEC is closed (3-phase close, 2026-08-17). This SPEC does **not** reopen it. It inherits
two of its outputs: the budget guard (`AlwaysLoadedTokenBudget`) and the stub + lazy-companion
pattern. `token_budget_guard.go` records that the 75,000 → 76,000 raise was temporary, pending
"a separate card" for the large-rule diet. **This SPEC is that card**, so the ratchet back is
in scope (§C REQ-AMC-011).

---

## §B. Goals

1. Give Codex a complete, non-truncated contract surface at `AGENTS.md`.
2. Reduce the always-loaded surface enough to ratchet the token budget constant back down.
3. Leave Claude Code behavior unchanged.
4. Make re-inflation past the Codex cap a mechanical CI failure, not a discovery.

---

## §C. Requirements (GEARS)

### C.1 Contract integrity

**REQ-AMC-001** (Ubiquitous) — The always-loaded surface shall carry every `[HARD]` clause
that binds a turn, in either the root `AGENTS.md` or the nested `AGENTS.md` owning the
directory the clause governs.

**REQ-AMC-002** (Unwanted) — The redistribution shall not relocate any `[HARD]` clause, or any
`MUST` / `MUST NOT` / `shall` obligation, into a skill, a lazy companion file, or any other
on-demand surface. Only rationale, procedure, worked examples, incident records, and
cross-reference tables are eligible for relocation.

**REQ-AMC-003** (Event-detected) — When a contract clause is rewritten for compression, the
rewritten clause shall preserve the original obligation's subject, modality, and scope; a
rewrite that narrows or widens what the clause binds is a defect, not a compression.

### C.2 Byte-budget conformance

**REQ-AMC-004** (Ubiquitous) — The sum of the root `AGENTS.md` and every nested `AGENTS.md`
that Codex merges into one chain shall not exceed the confirmed `project_doc_max_bytes` value,
measured in bytes on the shipped files.

**REQ-AMC-005** (Where) — Where the confirmed merge scope is a project-root → CWD chain, the
nested-document set shall be laid out so that the deepest reachable chain, not the total of all
nested files, is the quantity compared against the cap.

**REQ-AMC-006** (Event-detected) — When an edit raises the merged chain above the cap, the
repository's guard shall fail with the measured byte figure and the offending file named.

**REQ-AMC-007** (Ubiquitous) — The byte guard shall reuse
`internal/config/token_budget_guard.go`'s surface enumeration rather than introducing a second,
independently-drifting measurement path.

### C.3 Claude-side non-regression

**REQ-AMC-008** (Unwanted) — The redistribution shall not change Claude Code rule-loading
semantics, hook wiring, or any existing test's expected behavior.

**REQ-AMC-009** (Ubiquitous) — `CLAUDE.md` shall reach the contract layer through the same
`@`-import mechanism it already uses for `.moai/config/sections/*.yaml`, retaining a
Claude-only layer for material with no Codex counterpart.

**REQ-AMC-010** (Event-detected) — When the `@`-import chain fails to resolve a contract
document, the run-phase verification shall treat that as a failing acceptance criterion rather
than as a cosmetic warning.

### C.4 Budget ratchet

**REQ-AMC-011** (Ubiquitous) — `AlwaysLoadedTokenBudget` shall be lowered to a value derived
from the achieved post-diet measurement taken on the branch this SPEC lands on, and that value
shall be at or below 75,000.

**REQ-AMC-012** (Event-detected) — When the ratcheted constant is proposed, the achieved figure
shall be a measured `go test` output on the integration branch, not the figure measured in an
isolated worktree; the two differ (this worktree measures ≈ 71,212 tokens, while the release
integration state that forced the 76,000 raise measured 75,282).

### C.5 Distribution

**REQ-AMC-013** (Ubiquitous) — Every file landing under `.claude/`, `.moai/`, or the repo root
that ships to users shall be mirrored into `internal/template/templates/` and rebuilt with
`make build`.

**REQ-AMC-014** (Unwanted) — Template copies shall not carry SPEC IDs, REQ tokens, audit
citations, internal dates, commit SHAs, macOS-biased absolute paths, or `CLAUDE.local.md`
references.

### C.6 Entry preconditions

**REQ-AMC-015** (Where) — Where any of the four entry premises in §D.1 is unresolved, the SPEC
shall remain in plan phase; run-phase entry is gated on all four being measured, not assumed.

---

## §D. Constraints

### D.1 Entry preconditions (blocking — card t91 / M0 owns the first three)

`.moai/reports/t91/` is absent as of 2026-08-22. Four premises are unmeasured, and each one
changes the design rather than merely adding risk:

| # | Premise | What it decides |
|---|---|---|
| P1 | The real default of `project_doc_max_bytes` | Every byte target in this SPEC. §A.2 shows the contract already sits at 99 % of 32,768 B — a different default moves the design from "infeasible without rewriting" to "feasible" or to "infeasible entirely". |
| P2 | The merge scope of nested `AGENTS.md` (project-root → CWD chain, per-changed-file, or none) | Whether the "nested AGENTS.md per area" leg of the design **exists at all**. If Codex reads only the invocation CWD's chain, a repo-root invocation never sees the area files, and the whole contract must fit one document. |
| P3 | Whether truncation is visible to the user or silent | Whether a CI byte guard is one defense among several or the only one. Silent truncation makes REQ-AMC-006 load-bearing. |
| P4 | Whether `project_doc_max_bytes` is raisable from project scope (not only per-user `~/.codex/config.toml`) | Whether raising the cap is available as a lever at all. The card never considered it; the symbol exists in codex's `ConfigToml`, so the question is scope, not existence. If it is per-user only, it cannot be relied on for distributed users and the diet must carry the whole burden. |

P1-P3 belong to card t91 (M0). P4 is added by this SPEC and may be answered in the same pass.

### D.2 Standing constraints

- **No `[HARD]` demotion.** A rule that is not always present cannot bind every turn (REQ-AMC-002).
- **Claude parity.** Cross-harness divergence is not a licence to change Claude-side semantics.
- **Template-First.** Mirror before claiming distribution; the neutrality guard is CI-enforced.
- **Measurement provenance.** Every byte or token figure asserted in run-phase evidence names
  the command that produced it and the tree it was measured on.

---

## §E. Scope

### E.1 In scope

- The 14 always-loaded rule files under `.claude/rules/moai/` (202,621 B).
- `CLAUDE.md` (20,523 B) and its import layer.
- A new root `AGENTS.md`, plus nested `AGENTS.md` files if P2 permits them.
- `internal/config/token_budget_guard.go` — the ratchet and the new byte guard.
- Template mirrors of all of the above.

### E.2 `.claude/output-styles/moai/moai.md` — explicit ruling

The card omits this file; it is 61,706 B, the largest single always-loaded artifact, and
21.7 % of the whole surface. The ruling is **structurally exempt from the AGENTS.md
redistribution, and deferred (not exempt) for its own diet**:

- **Exempt from redistribution** because it is a Claude Code *output style* — a render-surface
  artifact with no Codex counterpart. Measured: §8 "Response Templates" occupies lines 193-713,
  **46,765 B = 75.8 % of the file**, and is entirely banner and template markup for Claude
  Code's response rendering. Copying banner templates into `AGENTS.md` would consume the whole
  Codex budget to deliver material Codex cannot act on.
- **Not exempt from the budget**, because the guard counts it. Its 75 `[HARD]` lines
  (11,898 B) are output-discipline rules that bind Claude's rendering only; they stay where
  they are.
- **Deferred to a follow-up card** for its own §8 diet. This SPEC's ratchet target (§C
  REQ-AMC-011) must be achievable without touching it — if run-phase measurement shows it is
  not, that is a blocker to surface, not a licence to widen scope mid-run.

### E.3 Exclusions

### Out of Scope — output-style diet

- Compressing or restructuring `.claude/output-styles/moai/moai.md` §8 Response Templates.
- Any change to Claude Code banner rendering or response templates.

### Out of Scope — codex premise measurement

- Measuring `project_doc_max_bytes`, nested-merge scope, or truncation visibility. These are
  card t91 (M0) deliverables consumed here as preconditions (§D.1).
- End-to-end codex model invocation to observe loading behavior.

### Out of Scope — conditional-load surfaces

- `.claude/agents/**`, `.claude/skills/**`, and any other file carrying `paths:` frontmatter.
  These are not always-loaded and are outside the measured surface.

### Out of Scope — reopening closed work

- Any modification to `SPEC-ALWAYS-LOADED-DIET-001`, which is closed. Its guard and its
  stub + lazy-companion pattern are inherited, not revised.

### Out of Scope — Codex runtime configuration

- Shipping or mutating a user's `~/.codex/config.toml`. Whether the cap is raisable is a
  premise to measure (P4), not a change to make.

---

## §F. Acceptance criteria

Enumerated in `acceptance.md`. Milestone decomposition in `plan.md`. Design detail — the
directory map, the extraction procedure, and the guard shape — in `design.md`.

---

## §G. Cross-references

- `.moai/reports/t82/measurement.md` — measured baseline (SSOT for §A.1).
- `internal/config/token_budget_guard.go` — budget constant, surface enumeration, ratchet target.
- `.moai/specs/SPEC-ALWAYS-LOADED-DIET-001/` — closed; source of the inherited guard and pattern.
- `.claude/rules/moai/core/verification-claim-integrity.md` — why §D.1 premises are preconditions
  rather than risks.
- `CLAUDE.local.md` §2 — Template-First rule and the neutrality guard (repo-local).

---
id: SPEC-AGENTS-MD-CANON-001
title: "AGENTS.md canonical contract layer for Codex dual-harness"
version: "0.4.0"
status: implemented
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

> **Provenance rule for this table.** The frontmatter `version:` equals the latest row below, and
> every row states a change the document actually contains. Neither is taken from a commit message:
> a commit-message label describes what an author intended to land, and the two drifted apart here
> once already (`version:` sat at `0.3.0` across two later revisions while commit messages read
> "v0.3.1" / "v0.3.2"). §D.4 imposes measurement provenance on every figure this SPEC asserts; the
> document's own revision is one of those figures. Check: `version:` matches the last row, and each
> row's claims are greppable in the artifacts.

| Date | Version | Change |
|---|---|---|
| 2026-08-22 | 0.1.0 | Initial draft (plan-phase). Card t82, milestone M2 of the Codex dual-harness epic. |
| 2026-08-22 | 0.2.0 | Rebuilt on `.moai/reports/t82/codex-probe.md`. Three premises measured: 32,768 B default confirmed, nested merge is root→CWD-path-only, truncation is silent. Nested-AGENTS.md leg dropped (Option A approved by dispatcher). Contract ceiling redefined against 32,768 B. |
| 2026-08-22 | 0.3.0 | Plan-audit iteration 1 (FAIL 0.69) revision. Enforceability fixes D1-D9: byte-guard criteria made executable, `AGENTS.md` singleton check moved to tracked files, integration branch given a discriminator, cap-raise rationale corrected to the measured position (P4), line-grep proxy disclosed, nested-`CLAUDE.md` asymmetry stated, `REQ-AMC-006` recast onto this SPEC. Parked edits folded in: output-style §8 lineage (t131 / t142), P4 discharged, M4 trust notice. |
| 2026-08-22 | 0.3.1 | Dispatcher refinements. `AC-AMC-016` cites the existing `TestAlwaysLoadedTokenBudget_OverBudgetFails` for the token-budget negative path instead of proposing a duplicate fixture; `design.md` §5.2 records why the singleton check uses `git ls-files` with `:(top)` / `:(exclude,top)` rather than a path-scoped `find`. |
| 2026-08-22 | 0.3.2 | Plan-audit iteration 2 (FAIL 0.83) revision. N1: headroom ratio pinned in `REQ-AMC-013` at 15 % ±2 percentage points, so `AC-AMC-019`'s check no longer reads a free variable. N2: stale `AC-AMC-012` cross-reference in `design.md` §4 corrected to `AC-AMC-013`. |
| 2026-08-22 | 0.3.3 | Plan-audit iteration 3 (FAIL 0.82) revision. D1: `AC-AMC-019` now reads the 13 %-17 % band rather than resolving the ratio solely from the constant's comment. D2: the achieved figure must be measured over an enumeration including the root `AGENTS.md` and every `@`-imported contract document — `alwaysLoadedSurface()` omits it today, so relocation into `AGENTS.md` would have scored as a diet (`REQ-AMC-013`, `AC-AMC-017`, plan.md M5). D3: the band's implied diet target (achieved ≤ 66,371 tokens) stated in §C.4. D4: document version and HISTORY brought current, with the provenance rule above added so `version:` is derived from the table rather than from a commit message. D5 (optional): `AC-AMC-018`'s measured state defined. Dispatcher additions: the enumeration content is bound into `REQ-AMC-008`, and the extension is ordered **before any measurement cited as a ratchet basis** (`REQ-AMC-013`, `AC-AMC-018`, `plan.md` M1/M5) — a late fix manufactures false evidence in the gap rather than merely delaying correctness. |
| 2026-08-22 | 0.3.4 | Plan-audit iteration 4 (FAIL 0.87, one finding) revision. E1: §C.4's required-cut figures recomputed over the enumeration `REQ-AMC-013` ¶2 requires — the contract layer is net-additive per §D.2 / `REQ-AMC-001` / `REQ-AMC-002`, so `AGENTS.md` joins the surface with nothing removed; cuts corrected 4,841 → **10,985** and 8,911 → **15,055** at the ceiling case, with the governing formula stated. E2: M1's stop condition gained **Arm B** — project the post-diet surface including the contract layer against 66,371 tokens and blocker on shortfall, so the ratchet's reachability is tested at M1 rather than discovered at M5 (`plan.md` M1, `AC-AMC-007`). E3 (optional, folded in while editing the bound): the ±1,000 tolerance makes 67,256 the strict maximum, so 66,371 is conservative by ~885 tokens — noted, not relaxed. |
| 2026-08-22 | 0.3.5 | Dispatcher readability additions to the E1/E2 edits; no requirement or criterion changed. §C.4 now **explains** the net-additive mechanism instead of citing it — a clause authored into `AGENTS.md` does not leave the always-loaded rules (`REQ-AMC-002` forbids it, `REQ-AMC-001` independently requires the clause in `AGENTS.md`), so the surface *grows* by `\|AGENTS.md\|` and the cut is `stated cut + \|AGENTS.md\|`. Stated rather than cross-referenced because four readers in sequence made the relocation assumption from the citation alone. M1's stop condition and `AC-AMC-007` now state that returning a blocker with the measured shortfall is a **correct outcome** of the pilot, not a milestone failure, and that Arm B should be expected to fire given the roughly doubled minimum cut. |
| 2026-08-22 | 0.3.6 | Plan-audit iteration 5 (FAIL 0.90, one finding) revision. F1: `REQ-AMC-010` and `AC-AMC-015` gained a **narrow surface-cardinality carve-out**. Fixed slots are appended unconditionally, so `REQ-AMC-008`'s fourth slot grows `len(surface)` before `AGENTS.md` is authored, breaking two hardcoded counts in `internal/config/token_budget_guard_test.go` (`wantRuleCount + 3` → `+ 4`; temp-tree `want 4` → `5`). Without the carve-out both exits failed a criterion — extend and edit the counts (fails `AC-AMC-015`) or leave the enumeration alone (fails `AC-AMC-017`). The exemption names those two assertions, covers the expected count and its comment only, and binds every behavioral expectation as before. F2 (optional): Arm B's projection is now **baselined on the integration-branch figure recorded at pre-flight** (`plan.md` M1, `AC-AMC-007`) — the two candidate trees differ by 4,070 tokens (37 % of the required cut), so a worktree-baselined projection could clear Arm B and still fail `AC-AMC-018` at M5; the M1 block quote now quotes **15,055** as the figure a reader meets at the point of use. |
| 2026-08-22 | 0.4.0 | Run-phase measured-figure correction (M1 pilot + pre-flight). No requirement or criterion changed; REQ stays 18, AC stays 24. (a) §A.4's line-proxy error direction corrected — M1 expanded the 97 markers to clause blocks and measured **51,639 B**, not 32,543, so the proxy runs roughly **one over and ninety-six under** rather than 15 over / 16 under; the verbatim contract therefore does **not** fit 32,768 B, and classification becomes the load-bearing lever rather than an optimisation. (b) §A.1 / §C.4's integration baseline: 75,282 exists on no live branch — all four live refs and this worktree measure 71,207 (guard-exact), so the required cut at the ceiling is **10,980**, not 15,055; §C.4's two rows converge and the second is retired, with the re-measurement obligation kept and its grounds stated. (c) §D.1's derivation re-stated against the clause-block figure, and the 3,300 B document-structure line marked an **assumption** with its basis. (d) §D.2 carries M1's per-clause split (16,135 B Codex-relevant / 35,147 B Claude-only / 357 B prose) and a conservative-direction constraint on M2's classification. Stale figures (32,543 as the contract total, 75,282, 15,055) survive only in the HISTORY rows that record their correction. |

---

## §A. Context

Codex (`codex-cli` 0.147.0) loads project instructions from `AGENTS.md` under a byte cap and
**silently** truncates the overflow. A rule cut mid-sentence is worse than an absent one, because
it reads as complete — and nothing in the tooling says it happened.

### A.1 Measured baseline — always-loaded surface

SSOT: `.moai/reports/t82/measurement.md` (2026-08-22). The measured surface is defined by the guard
that already owns it — `internal/config/token_budget_guard.go` `alwaysLoadedSurface()`: every
`.claude/rules/moai/**/*.md` without `paths:` frontmatter, plus three fixed slots. **The
enumeration does not yet include `AGENTS.md`** — REQ-AMC-013 requires M5 to add it, since this SPEC
makes that file always-loaded.

| Surface | Bytes |
|---|---:|
| always-loaded rules (14 files, no `paths:`) | 202,621 |
| `CLAUDE.md` | 20,523 |
| `.claude/output-styles/moai/moai.md` | 61,706 |
| repo-root `MEMORY.md` | 0 (absent; guard treats hermetically) |
| **total measured surface** | **284,850** |

Estimated tokens (the guard's `char/4`): **71,207**. Budget constant
`AlwaysLoadedTokenBudget = 76,000`; headroom 4,793 tokens; the guard test passes on this tree.

**Provenance of the token figure (run-phase M1 correction).** 71,207 is the guard's own arithmetic
— `measureAlwaysLoaded()` sums `len(file)/4` **per file**, so the per-file floors lose 5 tokens
against `284,850 / 4 = 71,212`. Both figures appear in this SPEC's history; the guard-exact one is
71,207 and is what `AC-AMC-018` will read off `go test -v`. Reproduced by
`.moai/reports/t82/surface_r3.py`, which re-implements `alwaysLoadedSurface()` +
`measureAlwaysLoaded()` (frontmatter-scoped `paths:`, `MEMORY.md` head cap, per-file floor):
`surface files: 17   guard tokens: 71207   surface bytes: 284850`. Nothing downstream turns on the
5-token difference; it is recorded so the two figures are not read as two measurements.

### A.2 Measured codex loading behavior

SSOT: `.moai/reports/t82/codex-probe.md` (2026-08-22), measured with `codex debug prompt-input`
against a git-initialised 3-level fixture with a byte-ruler root document. Zero model calls — this
is observation, not documentation trust.

| # | Finding | Evidence |
|---|---|---|
| 1 | `project_doc_max_bytes` default is **32,768 B** | last ruler marker carried at offset 32,670; the next at 32,780 absent. `-c project_doc_max_bytes=4096` cuts at 4,070, so the key is live |
| 2 | Truncation **takes the tail** — the head survives | the ruler's low offsets are present, high offsets absent |
| 3 | Nested `AGENTS.md` merges **only along the git-project-root → CWD path** | run from the repo root, `area/AGENTS.md` and `area/deep/AGENTS.md` contribute 0 marker hits |
| 4 | The chain **shares one budget, root-first** | with a 42,066 B root, both nested docs vanish entirely — not truncated, never loaded |
| 5 | Outside a git repo, only the CWD's own doc loads | project-root resolution is git-based |
| 6 | Truncation is **silent** | stderr 0 bytes, exit 0 at the default log level; the `project doc exceeds remaining budget; truncating` string is a tracing event that never reaches the user |
| 7 | A project-scope `<repo>/.codex/config.toml` **does** take effect — and beats the user value — but **only** where the project is registered `trust_level = "trusted"` in the user config | `.moai/reports/t82/codex-probe-p4.md`, four-way differential toggling only the trust entry: effective cap 4,096 → 8,192 (rows 3 → 4). Untrusted, the project file is **silently ignored** (stderr 0 bytes on all four runs) |

All four entry premises are **discharged by measurement**. P1-P3 by `codex-probe.md`, P4 by
`codex-probe-p4.md`. No premise gate remains on run-phase entry; the residual items are named in
§D.9. The probe fixtures are rebuildable from `.moai/reports/t82/probe-fixture.sh`, which
reconstructs the tree and prints each recorded run with its expected result.

### A.3 What finding 3 overturns

The card proposed widening the budget with per-area nested documents at ~4 KiB each. The
measurement says the opposite on both counts: nested documents **do not expand the budget** (they
share the same 32,768 B, consumed root-first), and they **are not loaded at all** in the ordinary
case of a developer running codex at the repo root.

Therefore the root `AGENTS.md` must be self-sufficient: the entire `[HARD]` contract has to fit
inside 32,768 B on its own, and every nested document added would spend root budget for a benefit
that materialises only in an area-scoped session. **Option A — a single root contract, zero nested
documents — is the approved design** (§E.2).

### A.4 Measured contract-layer volume

Commands (run in this worktree; outputs recorded in `progress.md` §E.1):

```
grep -rLE '^paths:' --include='*.md' .claude/rules | sort > /tmp/t82_always.txt
xargs grep -h '\[HARD\]' < /tmp/t82_always.txt | wc -c        # → 30353
grep -h '\[HARD\]' CLAUDE.md | wc -c                           # → 2190
grep -h '\[HARD\]' .claude/output-styles/moai/moai.md | wc -c  # → 11898
```

| Contract slice | Marked lines | Bytes |
|---|---:|---:|
| `[HARD]` lines across the 14 always-loaded rules | 93 | 30,353 |
| `[HARD]` lines in `CLAUDE.md` | 4 | 2,190 |
| **subtotal — Codex-relevant contract, line-proxy** | **97** | **32,543** |
| `[HARD]` lines in `.claude/output-styles/moai/moai.md` (Claude render surface, §E.2) | 75 | 11,898 |

**This figure is a line-level proxy for the contract, not a measurement of it.** The commands
count *lines bearing the marker*, which is a different quantity from *the obligations those lines
carry*. Plan-phase expected the error to run in both directions. M1 expanded every marker to its
clause block — the marker line plus its continuation to the next clause or heading — read all 97
blocks, and measured the direction to be almost entirely **one-way**
(`.moai/reports/t82/m1-pilot.md` §1, §1.1):

```
python3 .moai/reports/t82/clause-blocks.py > .moai/reports/t82/blocks.json
python3 .moai/reports/t82/summarize.py                         # → 97 blocks, 51,639 B
```

- **Overcount — one block, 357 B.** Of the 97 markers exactly **one** carries no obligation:
  `kanban-dispatch.md`'s detail-companion note ("the stub keeps every [HARD] rule and pointer").
  Nine further markers sit non-clause-initially, but each of those is a genuine obligation and is
  contract. The plan-phase estimate of 15 prose markers is superseded.
- **Undercount — the other ninety-six, and it dominates.** The clause-block total is **51,639 B**
  against the proxy's 32,543 B: **+58.7 %**. The worst case is `moai-constitution.md`'s six Agent
  Core Behaviors, which carry their marker on the *heading* line — the proxy counted six headings
  of ~50 B each where the entire body is the obligation, 5,117 B across the six.

**So the corrected error direction is roughly one over and ninety-six under**, not the 15-over /
16-under bracket this section carried at plan-phase. The magnitude of the proxy is trustworthy and
its value is not: nothing in this SPEC may depend on 32,543 B being precise. §D.1's ceiling
derivation was stated as a bracket rather than a point for exactly this reason, and M1's stop
condition (`plan.md` §E) was required to survive the figure moving in either direction — which it
did, by 19,096 B upward. §D.1 re-states the derivation against the clause-block figure.

**Measured as clause blocks, the verbatim contract does not fit the budget.** The proxy read as a
225 B fit against the confirmed 32,768 B, and that fit was an artifact of the undercount: 51,639 B
is **1.58× the entire budget**, before the document's own structure, before a user's global
`~/.codex/AGENTS.md` (consumed *first*, §D.3), and before any future rule growth.

So the work was never "make it fit" — it does not fit — and **classification is the load-bearing
lever rather than an optimisation**: the Codex-relevant subset measures **16,135 B across 35
blocks** (§D.2), and only after that cut does a contract with real headroom exist at all.
§C REQ-AMC-004 sets the contract layer's own ceiling accordingly rather than sizing it to fill the
budget.

### A.5 Relationship to SPEC-ALWAYS-LOADED-DIET-001

That SPEC is closed (3-phase close, 2026-08-17). This SPEC does not reopen it. It inherits the
budget guard (`AlwaysLoadedTokenBudget`) and the stub + lazy-companion pattern.
`token_budget_guard.go` records that the 75,000 → 76,000 raise was temporary, pending "a separate
card" for the large-rule diet. **This SPEC is that card**, so the ratchet back is in scope.

---

## §B. Goals

1. Give Codex a complete, non-truncated contract at a single root `AGENTS.md`, with real headroom.
2. Reduce the always-loaded surface enough to ratchet the token budget constant back down.
3. Leave Claude Code behavior unchanged.
4. Make re-inflation past the Codex cap a mechanical CI failure — the only available defence, since
   truncation is measured silent.

---

## §C. Requirements (GEARS)

### C.1 Contract integrity

**REQ-AMC-001** (Ubiquitous) — The root `AGENTS.md` shall carry every `[HARD]` clause that binds a
Codex-driven turn, self-sufficiently, without depending on any nested document being loaded.

**REQ-AMC-002** (Unwanted) — The redistribution shall not relocate any `[HARD]` clause, or any
`MUST` / `MUST NOT` / `shall` obligation, into a skill, a lazy companion file, or any other
on-demand surface. Only rationale, procedure, worked examples, incident records, and
cross-reference tables are eligible for relocation.

**REQ-AMC-003** (Event-detected) — When a contract clause is rewritten for compression, the
rewritten clause shall preserve the original obligation's subject, modality, and scope; a rewrite
that narrows or widens what the clause binds is a defect, not a compression.

### C.2 Byte-budget conformance

**REQ-AMC-004** (Unwanted) — The root `AGENTS.md` shall not exceed **24,576 B** (24 KiB), leaving
at least 8,192 B of the confirmed 32,768 B budget as headroom. The ceiling is derived in §D.1 from
what the contract requires, not from the budget's size.

**REQ-AMC-005** (Unwanted) — The live tree shall contain no `AGENTS.md` outside the repository
root. The template mirror at `internal/template/templates/AGENTS.md` is not a live-tree document
and is exempt (REQ-AMC-015 requires it). Rationale, and the nested-`CLAUDE.md` asymmetry a reader
will ask about: §D.6.

**REQ-AMC-006** (Ubiquitous) — This SPEC shall record, in §D.6, the evidence class that would
justify reviving a nested `AGENTS.md` and the obligation on any reviving SPEC to re-derive the root
ceiling to pay for it — so that the single-root decision is revisable on stated grounds rather than
by re-litigation.

**REQ-AMC-007** (Event-detected) — When an edit raises the root `AGENTS.md` above its ceiling, the
repository's guard shall fail with the measured byte figure and the offending file named.

**REQ-AMC-008** (Ubiquitous) — The byte guard shall reuse
`internal/config/token_budget_guard.go`'s surface enumeration rather than introducing a second,
independently-drifting measurement path, and that enumeration shall include **the root `AGENTS.md`
and every `@`-imported contract document** alongside its existing rule files and fixed slots.

The ordering constraint that makes this enforceable lives with the measurement it binds — see
REQ-AMC-013.

**REQ-AMC-009** (Ubiquitous) — The CI byte guard shall fail the build on breach; an advisory-only
guard does not satisfy this requirement. Rationale: §D.7.

### C.3 Claude-side non-regression

**REQ-AMC-010** (Unwanted) — The redistribution shall not change Claude Code rule-loading
semantics, hook wiring, or any existing test's expected behavior — **except** that an assertion
whose expected value is the *cardinality* of the always-loaded surface is updated by
`REQ-AMC-008`'s enumeration extension and is exempt. The exemption covers the expected count only;
every behavioral expectation remains bound. Scope and the two qualifying assertions: `AC-AMC-015`.

**REQ-AMC-011** (Ubiquitous) — `CLAUDE.md` shall reach the contract layer through the same
`@`-import mechanism it already uses for `.moai/config/sections/*.yaml`, retaining a Claude-only
layer for material with no Codex counterpart.

**REQ-AMC-012** (Event-detected) — When the `@`-import chain fails to resolve a contract document,
the run-phase verification shall treat that as a failing acceptance criterion rather than as a
cosmetic warning.

### C.4 Budget ratchet

**The band implies a diet target, and it is stated here rather than discovered at M5.** With the
headroom floor at 13 % and the constant capped at 75,000, `1.13 × N ≤ 75,000` bounds the achieved
figure at **N ≤ 66,371 tokens**. Measured against that ceiling:

**The cut is computed over the enumeration `REQ-AMC-013` ¶2 requires — one that includes
`AGENTS.md`.** The four rows in §A.1 measure the *unextended* enumeration, since the file does not
exist yet, so they are not the quantity the ratchet is measured against.

**Authoring `AGENTS.md` is a duplication, not a relocation — and that is the whole of the
correction.** A clause that goes into `AGENTS.md` does **not** leave the always-loaded rules: it
stays there, still loaded on every Claude turn, because `REQ-AMC-002` forbids moving any obligation
off the always-loaded surface and `REQ-AMC-001` independently requires `AGENTS.md` to carry every
Codex-binding `[HARD]` clause. §D.2 states the same property from the other side — excluding
Claude-only clauses from `AGENTS.md` "removes nothing from either harness's binding surface". So
the two requirements put each contract clause in *both* places by construction, and `AGENTS.md`
joins the measured surface with **nothing removed in exchange**. The measured surface therefore
*grows* by `|AGENTS.md|` at the moment the file is authored, which is why the required cut is
`stated cut + |AGENTS.md|` rather than the stated cut alone.

This is spelled out rather than cited because the assumption it corrects is an easy one to make
again: four separate readers — the card author, this SPEC's author, the dispatcher, and two audit
iterations — each worked from figures that quietly treated the contract layer as a relocation. A
cross-reference lets the next reader repeat that inference; a stated mechanism does not.

So the required cut is `stated surface + |AGENTS.md| − 66,371`. At `AGENTS.md`'s ceiling
(24,576 B ÷ 4 = 6,144 tokens — the worst case, and the one to plan against):

| Tree | Unextended surface | With `AGENTS.md` at ceiling | Required cut |
|---|---:|---:|---:|
| this worktree, and every live ref (see below) | 71,207 | 77,351 | **10,980 tokens** |

**Figures re-measured at run-phase entry; the second row is retired.** An earlier revision carried a
second row — 75,282 / 81,426 / 15,055 — labelled "the integration state that forced the 76,000
raise". That state exists on **no live branch**. Re-measured 2026-08-22 by
`.moai/reports/t82/measure-surface.py` against `origin/main`, `origin/release/v3.1.1`,
`origin/release/v3.1.2`, `origin/release/v3.1.3` and this worktree: all five measure 14 rules /
284,850 B / 71,212 (`total/4`), i.e. 71,207 by the guard's per-file arithmetic (§A.1) — which is why
the cut above reads 10,980 where `preflight.md` reads 10,985: one cut computed from the guard's
per-file floors, the other from the total, 5 tokens apart and nothing downstream turning on it. The 75,282
reading was a transient `release/v3.1.1` integration state that the subsequent always-loaded diet
(the closed `SPEC-ALWAYS-LOADED-DIET-001`) removed. Evidence: `.moai/reports/t82/preflight.md`.

This does not retire the re-measurement obligation. `REQ-AMC-014` and `AC-AMC-018` bind the ratchet
to the **v3.2 integration branch**, which does not exist yet; a sibling card adding rules raises the
baseline and the required cut by the same amount — which is exactly how the 76,000 raise arose. The
figure above is today's reading, not a settled baseline.

So §B goal 2's "reduce the always-loaded surface enough" has a number: at least **10,980 tokens** at
the ceiling case, re-measured against the v3.2 integration branch when it exists. M1 sizes its work
against this figure; a diet that lands anywhere above 66,371 cannot satisfy REQ-AMC-013 no matter
what ratio is declared. A smaller authored `AGENTS.md` reduces the cut proportionally — the formula
governs, the table is its ceiling case. M1 measured the projected `AGENTS.md` at 11,881 B
(2,970 tokens), which puts the required cut at **7,806 tokens**
(`.moai/reports/t82/m1-pilot.md` §5).

> The bound is conservative by design. `AC-AMC-019`'s ±1,000-token tolerance means the strictly
> admissible maximum is 67,256 rather than 66,371; planning against 66,371 asks for ~885 tokens
> more than the minimum. That direction is deliberate — the tolerance absorbs measurement noise,
> not slack to spend.

**REQ-AMC-013** (Ubiquitous) — `AlwaysLoadedTokenBudget` shall equal the achieved post-diet token
figure plus a headroom allowance of **15 %, within ±2 percentage points** (so the admissible band
is 13 %-17 % of the achieved figure), and shall be at or below 75,000. A constant set
independently of the achieved figure — at the ceiling, or at any round number — does not satisfy
this requirement even when it is below 75,000, and a headroom ratio chosen outside the 13 %-17 %
band does not satisfy it either.

The achieved figure shall be measured over an enumeration that **includes the root `AGENTS.md` and
every `@`-imported contract document**, not over `alwaysLoadedSurface()`'s current three fixed
slots. Verified against the implementation: the function enumerates rules-without-`paths:` plus
`CLAUDE.md`, the output style, and `MEMORY.md` — `AGENTS.md` is absent. But `REQ-AMC-011` makes
`AGENTS.md` an `@`-import of `CLAUDE.md`, so it *is* always-loaded from the moment it exists.
Measured against the unextended enumeration, relocating clauses out of the rule files and into
`AGENTS.md` would cut the achieved figure by up to ~6,144 tokens (24,576 B ÷ 4) **while removing
nothing from the always-loaded context** — and the ratchet would record that as a diet. A guard
that cannot see the file this SPEC creates cannot measure this SPEC's own effect.

**The enumeration extension (REQ-AMC-008) is ordered BEFORE any measurement cited as a ratchet
basis** — not merely before the ratchet's final commit. A late fix does more than delay
correctness: it **manufactures false evidence in the gap**. Every measurement taken while
`AGENTS.md` is unenumerated records a reduction that did not happen — clauses moved out of the rule
files into a file that is still always-loaded but no longer counted. Those readings are
indistinguishable from a real diet at the moment they are taken, and they are exactly the figures a
later actor would quote as the diet's evidence. The gap does not close retroactively: measurements
taken inside it stay wrong and stay quotable. Sequencing consequence in `plan.md` §E.

The ratio is pinned here rather than left to the constant's comment because `AC-AMC-019` checks the
constant against `achieved × (1 + ratio)`: if the ratio is whatever that same comment declares, the
check has a free variable and both are chosen by one actor in one edit. Achieved 60,000 with a
declared 25 % ratio yields exactly 75,000 and passes every criterion with zero delta — the same
vacuity the derivation check exists to close, relocated rather than removed. Bounding the ratio is
what makes the check binding; 15 % matches the allowance the original constant already carried
(`design.md` §5.1), and ±2 points absorbs rounding to a clean constant without admitting a
ratio chosen to reach a predetermined answer.

**REQ-AMC-014** (Event-detected) — When the ratcheted constant is proposed, the achieved figure
shall be a `go test -v` output measured on the **integration branch**, defined as the
release/batch branch this card merges into (`release/vX.Y.Z`), which carries the merged state of
the sibling cards — not a card worktree measured in isolation. The two can differ, and the gap is not academic: the
`release/v3.1.1` integration state that forced the 76,000 raise measured 75,282 tokens while each
constituent card sat within budget on its own. That divergence is **historical, not current** —
re-measured at run-phase pre-flight, this worktree and all four live refs read the same 71,207
(§C.4), so no live tree exhibits it today. It returns the moment a sibling card lands on the v3.2
integration branch, which is why the tree must be named rather than assumed. The evidence shall identify the measured tree by recording
`git rev-parse --abbrev-ref HEAD` and `git rev-list --count main..HEAD` alongside the figure.

### C.5 Distribution and disclosure

**REQ-AMC-015** (Ubiquitous) — Every file landing under `.claude/`, `.moai/`, or the repo root that
ships to users shall be mirrored into `internal/template/templates/` and rebuilt with `make build`.

**REQ-AMC-016** (Unwanted) — Template copies shall not carry SPEC IDs, REQ tokens, audit citations,
internal dates, commit SHAs, macOS-biased absolute paths, or `CLAUDE.local.md` references.

**REQ-AMC-017** (Ubiquitous) — The shipped documentation shall warn that a user's global
`~/.codex/AGENTS.md` joins the same merged chain and is consumed before the project document,
narrowing the project's available budget. Decision and reasoning: §D.3.

**REQ-AMC-018** (Unwanted) — [HARD] The design shall not treat a raised `project_doc_max_bytes` as
a substitute for the diet. The reduction target is fixed at the **untrusted first session's**
effective cap of 32,768 B; any raise is headroom layered on top of that, never a relaxation of it.
Measured grounds: §D.8.

---

## §D. Constraints and decisions

### D.1 The contract layer's ceiling — how 24,576 B was derived

The dispatcher's instruction is that 32,768 B is a **budget, not a target**. The ceiling is
therefore derived from what the contract requires, then checked for headroom:

The plan-phase derivation started from the line proxy and was explicitly a **bracket, not a
point**. M1 re-derived it against the clause-block measurement, as `AC-AMC-006` requires. The
re-derived steps (`.moai/reports/t82/project.py`, `m1-pilot.md` §4):

| Step | Bytes | Basis |
|---|---:|---|
| clause-block `[HARD]` total (rules + `CLAUDE.md`) | 51,639 | M1 measurement (§A.4) — supersedes the 32,543 line proxy |
| Less: clauses binding Claude-only mechanisms (61 blocks) | −35,147 | M1 per-clause classification (§D.2) |
| Less: the one non-obligation prose marker | −357 | §A.4 |
| = Codex-relevant contract, verbatim (35 blocks) | 16,135 | |
| × measured compression ratio 0.5318 | 8,581 | 11-clause pilot (`AC-AMC-005`) |
| Plus: document structure (headings, framing prose) | +3,300 | **assumption, not a measurement** — see below |
| = projected `AGENTS.md` | 11,881 | M1 projection |
| **Ceiling (REQ-AMC-004, unchanged)** | **24,576** | 75 % of budget |
| **Required headroom** | **≥ 8,192** | absorbs the global-layer slice (§D.3), structure, and future growth |

**The 3,300 B structure line is an assumption and is marked as one wherever it appears.** Its
basis: roughly 14 sections at ~190 B each (a heading line ~40 B plus one or two sentences of
navigational prose ~150 B), plus ~600 B of front matter — title, the self-sufficiency declaration,
and the global `~/.codex/AGENTS.md` warning pointer. The clause-to-origin trace table is a
`progress.md` deliverable rather than part of `AGENTS.md`, so it is not counted here. It becomes
**measurable at M2**, when the document has an actual skeleton; until then no criterion may treat
it as measured. Doubling it leaves Arm A intact (6,600 + 8,581 = 15,181 B) and raises Arm B's
required cut by 825 tokens.

The plan-phase bracket ran 18,183 B (full Claude-only exclusion) to 32,543 B (none); the
clause-block measurement moves both ends. The pessimistic case — no classification at all — is now
51,639 B verbatim, or 27,462 B even after the measured compression ratio, which clears neither the
ceiling nor a usable margin against the 32,768 B budget: **without classification the contract does
not fit.** With it the projection lands at 11,881 B, 48 % of the ceiling.

**The ceiling stays at 24,576 B.** The projection leaves room to tighten it, but tightening is a
SPEC change rather than an M1 finding, and the projection's own inputs still carry error — the
3,300 B assumption above, and the classification judgement §D.2 describes. The 8,192 B reserve is
what absorbs both; trading against that reserve is a decision to state, not slack to spend
silently. If a later measurement shows the contract cannot reach 24,576 B, the ceiling is
renegotiated with the number in hand rather than the SPEC quietly expanding to fill the budget.

### D.2 The Claude-only exclusion lever

Some `[HARD]` clauses bind mechanisms Codex does not have — `AskUserQuestion`, `ToolSearch`
preload, `/clear` and the paste-ready handoff, prompt-cache ordering, the `Skill` tool, Claude Code
cross-session messaging. Excluding them from `AGENTS.md` removes nothing from either harness's
binding surface: they remain always-loaded on the Claude side.

Measured upper bound — the `[HARD]` lines in the six most Claude-mechanism-bound files
(`askuser-protocol`, `session-handoff`, `cache-aware-execution`, `context-window-management`,
`cross-session-messaging`, `skill-routing`): **14,360 B across 38 lines**. That was an upper bound,
not the answer: some clauses inside those files state harness-generic principles and must stay.

**M1 produced the per-clause split, and file membership turned out to be the wrong unit.**
Measured per clause block (`.moai/reports/t82/classification.tsv`, 97 rows, zero unclassified):
Codex-relevant **16,135 B / 35 blocks**, Claude-only **35,147 B / 61 blocks**, non-obligation prose
357 B / 1 block. Inside the six named files the file proxy nearly held — 21,504 B of their
22,179 B of clause blocks is Claude-only (97.0 %) — but **38.8 % of all Claude-only bytes
(13,643 B) lie outside those six**, most of it in `kanban-dispatch.md`, where Claude-session
mechanisms (lead/lane dispatch, the `/clear` handoff) and harness-generic discipline (worktree
rules, verification load, reading evidence rather than trusting it) sit in one file. A file-level
split would have been wrong in both directions there.

**Classification is now the largest single lever (§D.1), so its error direction is a design
constraint on M2.** The dangerous error is classifying a Codex-binding clause as Claude-only: the
clause then never reaches `AGENTS.md`, and Codex loses the obligation with no signal — the same
silent-loss shape as the truncation this SPEC exists to prevent, and the mirror image of the
false-diet defect §C.4 corrects. The reverse error is cheap: a Claude-only clause carried into
`AGENTS.md` costs bytes and nothing else. **When in doubt, classify to the Codex side.** M1's
sensitivity analysis is what makes that affordable — the projection clears the ceiling even at zero
compression (16,135 + the assumed 3,300 B of structure = 19,435 B), and the ceiling would be reached only if the
Codex-relevant volume were 40,004 B, **2.5×** the measured amount. M1 already applied this
direction to seven boundary clauses (`m1-pilot.md` §2.2); M2 continues it.

### D.3 Global `~/.codex/AGENTS.md` — decision: warn in shipped docs

**Decision: yes, the shipped documentation carries the warning.**

Reasoning, recorded per the dispatcher's instruction:

- The failure it causes is silent (§A.2 finding 6) and lands on the *project's* rules, since the
  global layer is consumed first. A user with a large personal `AGENTS.md` would see MoAI's
  contract truncated with no signal at all.
- It is the one slice of the budget the CI guard structurally cannot see: the file lives outside
  the repository, on each user's machine, and its size is unknowable at build time. Documentation
  is the only defence available for it.
- The cost is a short paragraph in one shipped document. The asymmetry — a few lines against a
  silent, unattributable rule loss — settles it.

The alternative considered and rejected: staying silent on the grounds that most users have no
global file. That reasoning protects the common case and abandons the case that actually breaks,
and it is precisely the users who invest in a personal `AGENTS.md` who would hit it.

### D.4 Standing constraints

- **No `[HARD]` demotion.** A rule that is not always present cannot bind every turn (REQ-AMC-002).
- **Claude parity.** Cross-harness divergence is not a licence to change Claude-side semantics.
- **Template-First.** Mirror before claiming distribution; the neutrality guard is CI-enforced.
- **Measurement provenance.** Every byte or token figure in run-phase evidence names the command
  that produced it and the tree it was measured on.
- **Precedent for the compression target.** `SPEC-ALWAYS-LOADED-DIET-001` reduced
  `goal-directive.md` to a 6,531 B stub with a 17,334 B lazy companion — a 72 % always-loaded
  reduction with no obligation moved off the always-loaded surface. The pattern is proven; §D.1's
  24.5 % pessimistic-case trim is well inside it.

### D.6 Single root: rationale, the `CLAUDE.md` asymmetry, and the revival condition

**Why single-root** (REQ-AMC-005): a nested `AGENTS.md` spends root budget for a benefit that
materialises only in an area-scoped session, and the measurement shows nested documents are not
loaded at all under ordinary repo-root invocation (§A.2 findings 3-4).

**The asymmetry a reader will notice.** This repository already runs five nested per-directory
instruction files on the Claude side — `internal/{cli,config,hook,spec,template}/CLAUDE.md` — so
"one instruction file per repository" is plainly not this codebase's shape, and the prior art has
worked well. The cases differ on two measured properties, and only on those:

| | nested `CLAUDE.md` (existing, works) | nested `AGENTS.md` (rejected) |
|---|---|---|
| Loading | by path relevance — loaded when work touches that directory | only along the git-root → CWD chain; **absent at repo-root CWD** |
| Budget | no byte cap on the set | shares one 32,768 B budget, consumed root-first — a large root starves them entirely |

So the rejection is not a stance against nested instruction files. It is that the codex mechanism
gives them neither of the two properties that make the Claude-side ones useful.

**Revival condition** (REQ-AMC-006). A future SPEC MAY reintroduce a nested `AGENTS.md` for a
specific directory when it satisfies both of the following, stated explicitly in its own body:

1. **Evidence of session habit** — that sessions are in fact started inside that directory often
   enough to matter. Session-registry CWD counts or an equivalent observation; not an assertion
   that it "would be convenient".
2. **A re-derived root ceiling** — the root's 24,576 B ceiling lowered by at least the nested
   document's size, since the budget is shared. A nested file added without paying for it silently
   truncates the root contract's tail.

Recording the condition here is what makes the decision revisable on stated grounds. Verifying a
future SPEC's compliance with it is that SPEC's business, not this one's.

### D.7 Why the guard blocks rather than warns

Truncation emits nothing: stderr 0 bytes, exit 0, and the `truncating` string is a tracing event
that never surfaces (§A.2 finding 6). There is no runtime signal to fall back on, so an advisory
guard would leave the failure with no detector at all — the warning would scroll past in CI output
that nobody reads on a green build, and the contract's tail would go missing on every user's
machine with no one the wiser. Blocking is not a severity preference here; it is the only
configuration in which the guard detects anything.

### D.8 Cap-raise: measured position, and why the decision is unchanged

The earlier draft rejected raising `project_doc_max_bytes` on the stated ground that it is
"per-user config, so it cannot ship". **That premise is false**, and the correction matters even
though the decision does not change.

Measured (`codex-probe-p4.md`, four-way differential toggling only the trust entry): a
project-scope `<repo>/.codex/config.toml` **does** take effect, and **beats** the user value — but
only where the user config registers the project `trust_level = "trusted"`. Untrusted, the project
file is ignored, and ignored **silently** (stderr 0 bytes on all four runs).

The true reason is the stronger one. A distributed user's **first** session is untrusted by
construction: they clone the repository and run codex before any trust entry exists. At that
moment the effective cap is 32,768 B, and if the contract does not fit, it truncates — silently,
on the run that forms their first impression of the harness. A lever that works only after the
user has already trusted the project cannot carry a constraint that binds before they have.

Hence REQ-AMC-018: the reduction target is fixed at the untrusted first session's 32,768 B, and
any raise is headroom on top of it. The lever is real, conditional, retroactive, and
silent-on-failure — three of which disqualify it as a design premise.

### D.9 Residual unmeasured items

Carried openly rather than assumed:

- `AGENTS.override.md` precedence and `project_doc_fallback_filenames` were observed as symbols
  only. This design depends on neither, so both stay out of scope.
- The probe ran on macOS with `codex-cli` 0.147.0. A different default on another OS or version is
  not excluded. The CI byte guard's ceiling is a repo constant, so a smaller upstream default would
  be caught only by re-probing — noted as residual risk, not as a gate.

---

## §E. Scope

### E.1 In scope

- The 14 always-loaded rule files under `.claude/rules/moai/` (202,621 B).
- `CLAUDE.md` (20,523 B) and its import layer.
- A new root `AGENTS.md`, sized to REQ-AMC-004's ceiling.
- `internal/config/token_budget_guard.go` — the ratchet and the new byte guard.
- Template mirrors of all of the above, plus the §D.3 documentation warning.

### E.2 `.claude/output-styles/moai/moai.md` — explicit ruling

The file is 61,706 B, the largest single always-loaded artifact, 21.7 % of the whole surface. The
ruling is **structurally exempt from the AGENTS.md contract, and deferred (not exempt) for its own
diet**:

- **Exempt from the contract** because it is a Claude Code *output style* — a render-surface
  artifact with no Codex counterpart. Measured: §8 "Response Templates" occupies lines 193-713,
  **46,765 B = 75.8 % of the file**, entirely banner and template markup for Claude Code's response
  rendering. Copying it into `AGENTS.md` would consume the whole Codex budget to deliver material
  Codex cannot act on. Its 75 `[HARD]` lines (11,898 B) bind Claude's rendering only.
- **Not exempt from the budget**, because the guard counts it.
- **Its first render-surface diet has already landed** — cards t131 and t142 (−5,454 and −567
  tokens), integrated into `release/v3.1.1` under the render-surface ruling. So this is not a
  deferred-and-untouched surface; it is a surface that has had one pass. **Why 46,765 B remains
  after that pass: the first pass was a diet, not a removal** — it compressed the render surface
  without deleting the banner templates, which are what Claude Code actually renders from.
- **Whether a second pass is needed is decided by M1's result**, not assumed now. This SPEC's
  ratchet target must be reachable without a second output-style pass; if M1's measurement shows it
  is not, that is a blocker to surface, not a licence to widen scope mid-run.
  **M1 answered: no second pass is required.** Both stop-condition arms clear without touching this
  file — Arm A at 11,881 B against the 24,576 B ceiling, Arm B with 2,864 tokens of margin
  (`.moai/reports/t82/m1-pilot.md` § verdict). The condition is discharged rather than deferred.

### E.3 Exclusions

### Out of Scope — nested AGENTS.md documents

- Creating `AGENTS.md` in any directory of the live tree other than the repository root. Measured to
  be unloaded in ordinary repo-root invocation while still consuming the shared budget
  (REQ-AMC-005). The revival condition is recorded in §D.6 (REQ-AMC-006); satisfying it is a future
  SPEC's work, not this one's.
- The template mirror `internal/template/templates/AGENTS.md` is **not** covered by this exclusion
  — REQ-AMC-015 requires it, and it is not a live-tree document.

### Out of Scope — output-style diet

- Compressing or restructuring `.claude/output-styles/moai/moai.md` §8 Response Templates.
- Any change to Claude Code banner rendering or response templates.

### Out of Scope — codex feature surface

- `AGENTS.override.md` precedence and `project_doc_fallback_filenames` behavior.
- Shipping or mutating a user's `~/.codex/config.toml`, or raising `project_doc_max_bytes` as a
  remedy for the contract's size. The cap is treated as fixed at the untrusted first session's
  measured default of 32,768 B (REQ-AMC-018, §D.8). Emitting a project-scope cap from the wiring
  generator is M4/t88's scope, not this SPEC's — with the trust-notice obligation `plan.md` §E M4
  records.

### Out of Scope — conditional-load surfaces

- `.claude/agents/**`, `.claude/skills/**`, and any other file carrying `paths:` frontmatter.
  These are not always-loaded and are outside the measured surface.

### Out of Scope — reopening closed work

- Any modification to `SPEC-ALWAYS-LOADED-DIET-001`, which is closed. Its guard and its
  stub + lazy-companion pattern are inherited, not revised.

---

## §F. Acceptance criteria

Enumerated in `acceptance.md`. Milestone decomposition in `plan.md`. Design detail — the extraction
rings, the guard shape, and the rejected alternatives — in `design.md`.

---

## §G. Cross-references

- `.moai/reports/t82/codex-probe.md` — measured codex loading behavior (SSOT for §A.2).
- `.moai/reports/t82/measurement.md` — measured always-loaded surface (SSOT for §A.1).
- `internal/config/token_budget_guard.go` — budget constant, surface enumeration, ratchet target.
- `.moai/specs/SPEC-ALWAYS-LOADED-DIET-001/` — closed; source of the inherited guard and pattern.
- `.claude/rules/moai/core/verification-claim-integrity.md` — the evidence standard §D.4's
  measurement-provenance constraint applies.

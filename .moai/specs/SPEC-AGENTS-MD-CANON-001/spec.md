---
id: SPEC-AGENTS-MD-CANON-001
title: "AGENTS.md canonical contract layer for Codex dual-harness"
version: "0.3.3"
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

Estimated tokens (the guard's `char/4`): **≈ 71,212**. Budget constant
`AlwaysLoadedTokenBudget = 76,000`; headroom ≈ 4,788 tokens; the guard test passes on this tree.

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
carry*, and it errs in both directions:

- **Overcount.** **15 of the 93 rule lines** carry `[HARD]` non-clause-initially — prose mentions
  and navigation notes rather than obligations. A stub's "keeps every [HARD] rule and pointer" line
  is counted as contract.
- **Undercount, and unbounded.** A `[HARD]` lead line whose obligation continues into a list,
  table, or fenced block contributes only its first line. **16 of the 93 end in `:`** —
  structurally incomplete sentences whose bodies fall outside the total entirely. No bound exists
  on how much body those 16 carry.

So the magnitude is trustworthy and the exact value is not. Nothing in this SPEC may depend on
32,543 B being precise: §D.1's ceiling derivation is stated as a bracket rather than a point, and
M1's stop condition (`plan.md` §E) is required to survive the figure moving in either direction.
M1's classification therefore operates on **clause blocks** — a marker line plus its continuation
to the next clause or heading — not on grep lines, and re-derives the ceiling against the
clause-block figure once it has one.

**Taken at face value against the confirmed 32,768 B budget, the proxy measures 32,543 B — it fits,
with 225 B of headroom (0.7 %).** That is a numeric fit and a practical failure. Three things consume
headroom the raw line total does not account for: the document's own structure (headings, section
framing, the prose that makes clauses navigable), a user's global `~/.codex/AGENTS.md` which joins
the same chain and is consumed **before** the project document (§D.3), and ordinary future rule
growth. A contract sized at 99.3 % of budget would begin silently truncating on its first edit.

So the work is not "make it fit" — it already does, barely. The work is **to establish real
headroom**, and §C REQ-AMC-004 sets the contract layer's own ceiling accordingly rather than sizing
it to fill the budget.

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
semantics, hook wiring, or any existing test's expected behavior.

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

| Tree | Measured | Required cut |
|---|---:|---:|
| this worktree | 71,212 | **4,841 tokens** |
| the integration state that forced the 76,000 raise | 75,282 | **8,911 tokens** |

So §B goal 2's "reduce the always-loaded surface enough" has a number: at least 4,841 tokens, and
8,911 against the state the ratchet is actually measured on (REQ-AMC-014). M1 sizes its work
against this figure; a diet that lands anywhere above 66,371 cannot satisfy REQ-AMC-013 no matter
what ratio is declared.

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
the sibling cards — not a card worktree measured in isolation. The two differ: this worktree
measures ≈ 71,212 tokens, already under 75,000, while the release integration state that forced the
76,000 raise measured 75,282. The evidence shall identify the measured tree by recording
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

| Step | Bytes | Basis |
|---|---:|---|
| `[HARD]` line proxy (rules + `CLAUDE.md`) | 32,543 ± | line-level proxy, both error directions disclosed in §A.4 |
| Less: clauses binding Claude-only mechanisms | −0 … −14,360 | measured upper bound (§D.2); the real figure is M1's deliverable |
| Plus: document structure (headings, framing prose) | + unmeasured | M2 |
| **Proposed ceiling** | **24,576** | 75 % of budget |
| **Required headroom** | **≥ 8,192** | absorbs the global-layer slice (§D.3), structure, and future growth |

24,576 B sits above the pessimistic case (no Claude-only exclusion at all, so condensation carries
the whole 32,543 → 24,576 reduction — a 24.5 % trim, which §D.4's precedent suggests is
comfortable) and well above the optimistic case (18,183 B before any condensation). If M1's
measurement shows the contract cannot reach 24,576 B, the ceiling is renegotiated with the number
in hand rather than the SPEC quietly expanding to fill the budget.

Because the starting figure is a proxy (§A.4), this derivation is a **bracket, not a point**. M1
re-derives it against the clause-block measurement; the 8,192 B reserve is what absorbs the proxy's
error, and trading against that reserve is a decision to state, not a slack to spend silently.

### D.2 The Claude-only exclusion lever

Some `[HARD]` clauses bind mechanisms Codex does not have — `AskUserQuestion`, `ToolSearch`
preload, `/clear` and the paste-ready handoff, prompt-cache ordering, the `Skill` tool, Claude Code
cross-session messaging. Excluding them from `AGENTS.md` removes nothing from either harness's
binding surface: they remain always-loaded on the Claude side.

Measured upper bound — the `[HARD]` lines in the six most Claude-mechanism-bound files
(`askuser-protocol`, `session-handoff`, `cache-aware-execution`, `context-window-management`,
`cross-session-messaging`, `skill-routing`): **14,360 B across 38 lines**. This is an upper bound,
not the answer: some clauses inside those files state harness-generic principles and must stay.
Producing the per-clause split is M1's deliverable, alongside the compression ratio.

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

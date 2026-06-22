---
id: SPEC-SIMPLICITY-AUDIT-001
title: "Over-Engineering-Only Lean Audit (/moai review --lean mode)"
version: "0.1.0"
status: completed
created: 2026-06-23
updated: 2026-06-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/skills/moai/workflows"
lifecycle: spec-anchored
tags: "simplicity, over-engineering, review, lean-audit, mx-debt, doctrine, ponytail, dogfooding"
era: V3R6
tier: S
---

# SPEC-SIMPLICITY-AUDIT-001 — Over-Engineering-Only Lean Audit (`/moai review --lean` mode)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-23 | manager-spec | Initial plan-phase draft. Absorbs the THIRD `DietrichGebert/ponytail` (MIT) mechanism — a narrow, single-purpose over-engineering audit ("ponytail-review" diff-scope + "ponytail-audit" repo-scope) — into MoAI as a flag-gated `--lean` mode of the existing `/moai review` workflow. Sibling to `SPEC-SIMPLICITY-LADDER-001` (REQ-1 decision ladder + REQ-2 `@MX:DEBT` tag, completed origin 9d7f7b266). Tier S: pure doctrine/skill-prose, zero Go. Benchmark/measurement (ponytail's 0-3 LLM-as-judge harness) is OUT OF SCOPE (§J). |

---

## §A. Background

### A.1 Source of inspiration (ponytail — the third mechanism)

`DietrichGebert/ponytail` (MIT) is a "lazy senior dev" minimalist-coding skill. Its sibling SPEC `SPEC-SIMPLICITY-LADDER-001` already absorbed two of its mechanisms (the ordered decision ladder + the `@MX:DEBT` deferred-simplification tag). This SPEC absorbs ponytail's THIRD mechanism: a **narrow, single-purpose over-engineering audit** — distinct from MoAI's existing comprehensive `/moai review`.

ponytail ships two scopes of the same audit:
- `ponytail-review` — diff-scope (review only the changes).
- `ponytail-audit` — repo-scope (sweep the whole tree).

Both hunt **ONLY over-engineering** — nothing else. They are read-only, one-shot, and apply NO fixes. They emit findings under 5 fixed tags and close with a net-reduction summary. The hard scope boundary is what makes the mechanism valuable: by EXCLUDING correctness bugs, security, and performance (those belong to standard review), the lean audit gives a focused, high-signal "what can be cut" lens that the comprehensive review dilutes across four perspectives.

### A.2 The 5 finding tags + output format (the mechanism to absorb)

| Tag | What it flags |
|-----|---------------|
| `delete:` | Unused or speculative code (dead branches, never-called helpers, write-only config) |
| `stdlib:` | Reimplemented standard library (hand-rolled what the language's standard library already provides) |
| `native:` | A dependency or code duplicating a platform-native feature (the platform already does this) |
| `yagni:` | Single-implementation abstraction, single-caller layer, dead config knob (premature generality) |
| `shrink:` | Logic reducible to fewer lines without loss of clarity |

Output line format (verbatim mechanism): `<tag> <what to cut>. <replacement>. [path]`

Closing summary: `net: -<N> lines, -<M> deps possible` when removals are warranted; OR the literal `Lean already. Ship.` when nothing warrants removal.

### A.3 Why MoAI does not already have this (the honest delta)

MoAI is not starting from zero on the *idea* of simplicity, but it has no *narrow over-engineering-only audit surface*:

- `/moai review` (`.claude/skills/moai/workflows/review.md`, ~260 lines) is the COMPREHENSIVE 4-perspective review (Security / Performance / Quality / UX). Over-engineering is, at best, a sub-concern folded into Perspective 3 (Quality → "consistency with project patterns"). There is no flag, no dedicated lens, and no net-reduction summary. The comprehensive review's strength (breadth) is exactly the lean audit's weakness to fix (it dilutes the "what can be cut" signal across four perspectives).
- `sync-auditor` (`.claude/agents/moai/sync-auditor.md`) scores 4 dimensions (Functionality / Security / Craft / Consistency) as a PASS/FAIL **gate**. A grep confirms it has NO over-engineering dimension. But its job is *skeptical-quality verdict scoring against acceptance criteria* — a gate, not a finding-only lens. The lean audit is finding-only and verdict-less (§B.2 decides this is NOT a sync-auditor sub-check).
- `.claude/skills/moai/references/anti-patterns.md` already enumerates "over-engineering" (Karpathy Categories 1 Premature Abstraction + 2 Over-Engineering/God Object), mapped to Behavior 4 Enforce Simplicity. This is the doctrine the lean audit's tags REUSE — the lean audit is the *operational scan surface* the existing catalogue describes statically. Cross-reference, do not duplicate (REQ-3).

So the net-new contribution is the **narrow, flag-gated over-engineering-only scan surface** with the 5-tag output format + net-reduction summary — NOT a new idea about simplicity, and NOT a new skill subsystem.

### A.4 What this SPEC is NOT

This SPEC is doctrine/skill-prose absorption, not a new subsystem. It introduces NO new skill file, NO new agent, NO new config file, NO new lint rule, NO new Go code, NO new hook. The mechanism lands as a flag-gated `--lean` MODE of the existing `/moai review` skill — the lightest possible integration point (the irony guard: a SPEC about over-engineering must not itself over-engineer — `plan.md` §G).

---

## §B. Research Findings

The four required research items were resolved at plan-authoring time. Each finding is backed by an actual Read/Grep against the current tree (verification-claim-integrity §1.1 surface 3 — observed, not assumed).

### B.1 Application point — a flag-gated MODE of `review.md`, NOT a standalone skill

> **Evidence**: Read of `.claude/skills/moai/workflows/review.md` (260 lines, observed in this run).

`review.md` is a comprehensive multi-perspective review that ALREADY uses flag-gated lenses: `--security` (deepens Perspective 1), `--design` (Phase 4.5 design-pattern extraction), `--critique` (Phase 4.5 craft review), `--staged` / `--branch` / `--file` (scope selection). The lean audit fits this exact established pattern as a new `--lean` flag:

- **`--lean` is a MODE, not a standalone skill.** When `--lean` is present, the review SHORT-CIRCUITS the 4-perspective analysis and runs ONLY the over-engineering scan with the 5-tag output + net summary. This reuses `review.md`'s existing Phase 1 (Identify Changes — diff for `--lean` diff-scope, full tree for `--lean --repo` repo-scope) and Phase 5 (Next Steps) machinery; it adds only the scan body + report format. This is the lightest integration: one new flag + one new mode-section, no duplicated change-identification or report-consolidation code.
- **A standalone skill is REJECTED.** A separate `.claude/skills/moai/workflows/lean-audit.md` would duplicate `review.md`'s Phase 1 change-identification and report machinery. That is precisely the over-engineering the SPEC is about (irony guard — §J, plan.md §G).

### B.2 NOT a sync-auditor Craft-dimension sub-check (verdict semantics would be corrupted)

> **Evidence**: Read of `.claude/agents/moai/sync-auditor.md` (observed in this run). A Craft dimension exists (20% weight, "Test coverage ≥ 85%, error handling"); NO over-engineering dimension exists.

A Craft-dimension sub-check inside sync-auditor was considered and REJECTED for a semantic reason:

- sync-auditor is a **PASS/FAIL gate** with a must-pass firewall (Functionality + Security must independently meet threshold or overall FAIL). Its findings feed a *verdict*.
- The lean audit is **finding-only, verdict-less, advisory**. It produces a "what can be cut" list + a net-reduction estimate; it never gates a SPEC. ponytail applies NO fixes and renders NO pass/fail.
- Folding a no-gate advisory lens into the 4-dimension must-pass scoring would conflate two different jobs: skeptical-quality-verdict (sync-auditor's domain) vs lean-finding-enumeration (the lean audit's domain). It would also risk a *false FAIL* — over-engineering is a quality smell, not a correctness defect; gating a SPEC on a `shrink:` finding would block working, correct code. The two stay SEPARATE: sync-auditor keeps its 4-dimension gate; the lean audit is a `/moai review --lean` advisory lens.

### B.3 `@MX:DEBT` cross-link — connect one-directionally (reuse the sibling's landed surface)

> **Evidence**: Grep of `.claude/rules/moai/workflow/mx-tag-protocol.md:57,65` (observed) confirms the sibling `SPEC-SIMPLICITY-LADDER-001` landed `@MX:DEBT` with the `moai mx query --kind DEBT --json` harvest emitting `"rotRisk": "no-trigger"`.

The `yagni:` lean-audit tag (single-implementation abstraction, single-caller layer, dead config) is EXACTLY the deliberate-simplification case `@MX:DEBT` records. The connection is one-directional and reuses the sibling's already-shipped surface — no new code:

- When the lean audit emits a `yagni:` finding for a site that ALREADY carries an `@MX:DEBT` marker (harvestable via `moai mx query --kind DEBT`), the finding is a known, deliberately-deferred simplification — NOT a fresh discovery. The lean audit SHOULD annotate it as `yagni: [already tracked @MX:DEBT — deferred]` rather than re-flag it as new, so the operator is not asked to re-decide a choice already recorded.
- This is a one-directional read (lean audit consults the `@MX:DEBT` harvest); the lean audit does NOT write `@MX:DEBT` markers (that is the run-phase author's job per the sibling SPEC). Staying fully independent was the alternative; it was rejected because it would surface known-deferred debt as if it were new, generating noise the sibling SPEC's `@MX:DEBT` mechanism was built to suppress.

### B.4 anti-patterns.md — reuse the Karpathy catalogue, do NOT add new entries

> **Evidence**: Read of `.claude/skills/moai/references/anti-patterns.md` (observed). Category 1 (Premature Abstraction) + Category 2 (Over-Engineering / God Object), both mapped to Behavior 4 Enforce Simplicity, already cover the `yagni:` and `delete:` semantic space with wrong/right code examples.

The lean audit's 5 tags MAP onto the existing Karpathy catalogue (`yagni:` ↔ Category 1 Premature Abstraction; `delete:` ↔ Category 2 unused-field/God-object; `stdlib:`/`native:`/`shrink:` ↔ Behavior 4 Enforce Simplicity prose). The lean audit is the OPERATIONAL scan that the static catalogue describes. REQ-3 adds ONE cross-reference line from the `--lean` mode section to `anti-patterns.md`; it does NOT restate the catalogue (single source of truth).

---

## §C. Requirements (GEARS)

### REQ-1 — `--lean` over-engineering-only review mode

**REQ-1.1 (Where — capability gate).** **Where** the `--lean` flag is supplied to `/moai review`, the review skill **shall** short-circuit the comprehensive 4-perspective analysis (Security / Performance / Quality / UX) and run ONLY the over-engineering scan. The `--lean` mode **shall not** report correctness bugs, security findings, or performance findings — those remain the domain of the default (non-`--lean`) comprehensive review.

**REQ-1.2 (Ubiquitous).** The `--lean` mode **shall** support two scopes consistent with the existing review scope flags: diff-scope by default (the changed code, reusing the existing `--staged` / `--branch` / `--file` scope selection) and repo-scope when `--repo` is additionally supplied (sweep the whole tree). This mirrors ponytail's `ponytail-review` (diff) vs `ponytail-audit` (repo) split using MoAI's existing flag vocabulary rather than a new command name.

**REQ-1.3 (Ubiquitous).** The `--lean` mode **shall** emit findings under exactly the 5 fixed tags `delete:` / `stdlib:` / `native:` / `yagni:` / `shrink:`, one finding per line, in the format `<tag> <what to cut>. <replacement>. [path]`.

**REQ-1.4 (Ubiquitous).** The 5 tags and their finding descriptions **shall** be phrased generically across all 16 supported languages — `stdlib:` **shall** name "the language's standard library" and `native:` **shall** name "a platform-native feature" without referencing any single language's standard library, package manager, or platform feature by name (no JS/Python/Go bias — same discipline as the sibling REQ-1.2 ladder).

**REQ-1.5 (When — event-detected).** **When** the `--lean` scan completes and at least one removal is warranted, the mode **shall** close with the summary line `net: -<N> lines, -<M> deps possible`. **When** the scan completes and nothing warrants removal, the mode **shall** instead close with the literal line `Lean already. Ship.`

**REQ-1.6 (Ubiquitous).** The `--lean` mode **shall** be read-only and advisory: it **shall not** apply any fix, **shall not** modify any file, and **shall not** render a PASS/FAIL verdict. Remediation routes through the existing Phase 5 Next Steps (`/moai fix`, create fix tasks, export report, dismiss) like the rest of `review.md`.

### REQ-2 — `@MX:DEBT` one-directional cross-link

**REQ-2.1 (When — event-detected).** **When** the `--lean` mode emits a `yagni:` finding for a code site that already carries an `@MX:DEBT` marker (as harvested by `moai mx query --kind DEBT`), the finding **shall** be annotated as already-tracked deferred debt (e.g. `yagni: <site> [already tracked @MX:DEBT — deferred]`) rather than reported as a fresh discovery.

**REQ-2.2 (Ubiquitous).** The `--lean` mode **shall not** create, modify, or remove any `@MX:DEBT` marker. The cross-link is a one-directional read of the existing `moai mx query --kind DEBT` harvest surface shipped by `SPEC-SIMPLICITY-LADDER-001`; authoring `@MX:DEBT` markers remains the run-phase author's responsibility per the sibling SPEC.

### REQ-3 — Doctrine cross-references (reuse, do not duplicate)

**REQ-3.1 (Ubiquitous).** The `--lean` mode section **shall** carry one cross-reference line pointing at `.claude/skills/moai/references/anti-patterns.md` (Karpathy Categories 1 + 2, Behavior 4), framing the 5 lean tags as the operational scan surface for the over-engineering anti-patterns already catalogued there. It **shall not** restate the catalogue (single source of truth).

**REQ-3.2 (Ubiquitous).** The `--lean` mode section **shall** carry one cross-reference line pointing at `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors #4 Enforce Simplicity (the 6-rung decision ladder absorbed by the sibling `SPEC-SIMPLICITY-LADDER-001`), framing the lean audit as the *post-hoc detection* counterpart to the ladder's *pre-code prevention*.

### REQ-4 — Template mirror

**REQ-4.1 (Ubiquitous).** Every edit to `.claude/skills/moai/workflows/review.md` **shall** be mirrored to its template source `internal/template/templates/.claude/skills/moai/workflows/review.md` (Template-First discipline, CLAUDE.local.md §2). The mirrored `--lean` mode content **shall** be language-neutral per REQ-1.4 (no moai-adk-internal SPEC IDs, REQ tokens, or dev dates leak into the template — CLAUDE.local.md §25 template internal-content isolation).

---

## §3. Acceptance Criteria (inline — Tier S)

Tier S records AC inline in spec.md §3 (no separate acceptance.md). Each AC is binary and independently verifiable. Given-When-Then framing where it adds clarity.

### AC-SA-001 — `--lean` flag documented as an over-engineering-only mode (REQ-1.1)
- **Given** `.claude/skills/moai/workflows/review.md`,
- **When** `grep -n '\-\-lean' .claude/skills/moai/workflows/review.md` runs,
- **Then** there is at least one match, AND the `--lean` Supported-Flags entry + the mode section explicitly state it short-circuits the 4-perspective analysis and excludes correctness / security / performance findings.
- Verify: `grep -c 'over-engineer' .claude/skills/moai/workflows/review.md` ≥ 1 AND the `--lean` section names the EXCLUDES boundary (correctness/security/performance).

### AC-SA-002 — 5 fixed tags present with the canonical output format (REQ-1.3)
- **Given** the `--lean` mode section,
- **Then** all 5 tag literals `delete:`, `stdlib:`, `native:`, `yagni:`, `shrink:` appear, AND the output-line format `<tag> <what to cut>. <replacement>. [path]` is documented.
- Verify: `for t in 'delete:' 'stdlib:' 'native:' 'yagni:' 'shrink:'; do grep -q "$t" .claude/skills/moai/workflows/review.md || echo "MISSING $t"; done` prints nothing.

### AC-SA-003 — net-reduction summary + "Lean already. Ship." (REQ-1.5)
- **Given** the `--lean` mode section,
- **Then** both closing forms are documented: the `net: -<N> lines, -<M> deps possible` pattern AND the literal `Lean already. Ship.`
- Verify: `grep -q 'net: -' .claude/skills/moai/workflows/review.md && grep -q 'Lean already. Ship.' .claude/skills/moai/workflows/review.md`.

### AC-SA-004 — language neutrality (REQ-1.4)
- **Given** the `--lean` mode section,
- **When** scanned for single-language standard-library / package-manager / platform tokens,
- **Then** `stdlib:` is described as "the language's standard library" (generic) and `native:` as "a platform-native feature" (generic); NO bias token (e.g. a specific language's package-manager name or standard-library module name) appears in the tag descriptions.
- Verify: the `--lean` section uses generic phrasing; a reviewer confirms no JS/Python/Go-specific stdlib/package-manager identifier appears in the 5-tag descriptions.

### AC-SA-005 — read-only / advisory / no verdict (REQ-1.6)
- **Given** the `--lean` mode section,
- **Then** it explicitly states the mode applies NO fixes, modifies NO files, and renders NO PASS/FAIL verdict, AND routes remediation through the existing Phase 5 Next Steps.
- Verify: the section contains the read-only + no-verdict statement AND references Phase 5 Next Steps.

### AC-SA-006 — `@MX:DEBT` one-directional cross-link (REQ-2.1, REQ-2.2)
- **Given** the `--lean` mode section,
- **Then** it documents that a `yagni:` finding on a site already carrying `@MX:DEBT` (harvested via `moai mx query --kind DEBT`) is annotated as `[already tracked @MX:DEBT — deferred]` rather than re-flagged, AND states the mode never writes/modifies `@MX:DEBT` markers.
- Verify: `grep -q 'mx query --kind DEBT' .claude/skills/moai/workflows/review.md && grep -q 'already tracked @MX:DEBT' .claude/skills/moai/workflows/review.md`.

### AC-SA-007 — doctrine cross-references present, not duplicated (REQ-3.1, REQ-3.2)
- **Given** the `--lean` mode section,
- **Then** it cites `anti-patterns.md` (Karpathy Cat 1+2 / Behavior 4) AND `moai-constitution.md` § Agent Core Behaviors #4 (the ladder), without restating the 8-category catalogue or the 6-rung ladder body.
- Verify: `grep -q 'anti-patterns.md' .claude/skills/moai/workflows/review.md && grep -q 'Enforce Simplicity' .claude/skills/moai/workflows/review.md` AND the section does NOT inline the full Karpathy catalogue or the full 6-rung ladder.

### AC-SA-008 — template mirror parity + neutrality (REQ-4.1)
- **Given** the live `review.md` and its template mirror,
- **When** `diff -q .claude/skills/moai/workflows/review.md internal/template/templates/.claude/skills/moai/workflows/review.md` runs (modulo any pre-existing live-vs-template deltas unrelated to `--lean`),
- **Then** the `--lean` mode content is present in BOTH and the template copy contains NO moai-adk-internal SPEC ID / REQ token / dev date (CLAUDE.local.md §25).
- Verify: the `--lean` section exists in the template mirror AND `grep -E 'SPEC-SIMPLICITY-AUDIT-001|REQ-SA|2026-06-' internal/template/templates/.claude/skills/moai/workflows/review.md` prints nothing.

> **Note on AC-SA-008 mirror baseline**: `review.md` may already carry live-only deltas from the template (the template tree is a SUBSET of live — some content is live-only). The run-phase MUST confirm `review.md` IS a template-managed mirror target (it lives under `.claude/skills/moai/workflows/`, which is template-distributed) before asserting full parity; if a pre-existing live-vs-template delta exists outside the `--lean` scope, the AC is scoped to the `--lean` content only, not full-file parity.

---

## §J. Exclusions

This section enumerates what this SPEC deliberately does NOT build. Keeping it small is itself an application of the simplicity discipline (the irony guard — plan.md §G).

### Out of Scope — LLM-as-judge benchmark / measurement
- ponytail ships a benchmark harness that scores minimalist-coding enforcement on a 0-3 rubric via an LLM-as-judge. This SPEC absorbs ONLY the *audit mechanism* (the 5-tag over-engineering scan + net-reduction summary), NOT the benchmark. Porting an LLM-as-judge measurement harness is a separate effort with its own justification burden, and the sibling `SPEC-SIMPLICITY-LADDER-001` §J already excluded the benchmark port for the same reason. The user explicitly deferred it.

### Out of Scope — standalone lean-audit skill
- No new `.claude/skills/moai/workflows/lean-audit.md` (or any new skill file). The mechanism lands as a `--lean` MODE of the existing `/moai review` skill (§B.1). A standalone skill would duplicate `review.md`'s change-identification + report machinery — the exact over-engineering this SPEC is about.

### Out of Scope — sync-auditor over-engineering dimension
- No new "over-engineering" or "leanness" dimension added to sync-auditor's 4-dimension scoring (§B.2). The lean audit is finding-only and verdict-less; sync-auditor is a PASS/FAIL gate. Conflating them would corrupt sync-auditor's verdict semantics and risk false FAILs on working, correct code.

### Out of Scope — new lint rule / hook / config / Go code
- No new `internal/spec/lint.go` rule, no PreToolUse/PostToolUse hook to mechanically enforce leanness, no new `.moai/config/sections/*.yaml`, no Go change. The lean audit is a skill-prose advisory mode. A mechanical enforcement gate is a separate, larger concern (and would itself risk over-engineering a simplicity feature).

### Out of Scope — auto-fix / auto-removal
- The `--lean` mode applies NO fixes and removes NO code (REQ-1.6). It is read-only and advisory. Auto-removal of flagged code would require correctness verification (does deleting this break a caller?) that the lean audit explicitly excludes from its scope — that routes through the existing `/moai fix` Phase 5 Next Step, not the lean mode.

### Out of Scope — writing `@MX:DEBT` markers
- The `--lean` mode reads the `@MX:DEBT` harvest (REQ-2.1) but NEVER writes, modifies, or removes `@MX:DEBT` markers (REQ-2.2). Authoring `@MX:DEBT` is the run-phase author's responsibility per the sibling `SPEC-SIMPLICITY-LADDER-001`. A bidirectional link (lean audit auto-creating `@MX:DEBT` from a `yagni:` finding) is out of scope — it would let an advisory read-only lens mutate source.

### Out of Scope — anti-patterns.md catalogue expansion
- No new entries added to `.claude/skills/moai/references/anti-patterns.md`. The 5 lean tags reuse the existing Karpathy Categories 1 + 2 (§B.4); REQ-3.1 adds a cross-reference line only. Expanding the catalogue would duplicate doctrine that already exists.

---

## §H. Cross-References

- `.claude/skills/moai/workflows/review.md` (REQ-1/REQ-2/REQ-3 insertion point — the `--lean` mode lands here as a new flag + mode section) + its template mirror `internal/template/templates/.claude/skills/moai/workflows/review.md` (REQ-4 mirror target).
- `.claude/skills/moai/references/anti-patterns.md` (REQ-3.1 cross-reference target — Karpathy Categories 1 Premature Abstraction + 2 Over-Engineering, Behavior 4; reuse, do not duplicate).
- `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors #4 Enforce Simplicity (REQ-3.2 cross-reference target — the 6-rung decision ladder; the lean audit is its post-hoc detection counterpart).
- `.claude/rules/moai/workflow/mx-tag-protocol.md:57,65` (REQ-2 `@MX:DEBT` harvest surface — `moai mx query --kind DEBT --json` with `rotRisk: no-trigger`, shipped by the sibling SPEC).
- `.claude/agents/moai/sync-auditor.md` (§B.2 rejected-home evidence — 4-dimension PASS/FAIL gate, no over-engineering dimension; the lean audit stays SEPARATE).
- `SPEC-SIMPLICITY-LADDER-001` (sibling, completed origin 9d7f7b266) — REQ-1 decision ladder + REQ-2 `@MX:DEBT` tag. This SPEC is the third ponytail mechanism, building on the second (`@MX:DEBT`) via the REQ-2 one-directional cross-link.
- `DietrichGebert/ponytail` (MIT) — source of inspiration for the `ponytail-review` (diff) + `ponytail-audit` (repo) over-engineering-only audit, the 5 finding tags, and the `net: …` / `Lean already. Ship.` summary. Wording adapted to MoAI voice; benchmark harness excluded (§J).
- CLAUDE.local.md §2 (Template-First mirror duty), §15 + §25 (template language neutrality + internal-content isolation).

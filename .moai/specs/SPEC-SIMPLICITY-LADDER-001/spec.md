---
id: SPEC-SIMPLICITY-LADDER-001
title: "Simplicity Decision Ladder + @MX:DEBT Deferred-Simplification Tag"
version: "0.1.0"
status: in-progress
created: 2026-06-22
updated: 2026-06-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai"
lifecycle: spec-anchored
tags: "simplicity, decision-ladder, mx-tag, technical-debt, doctrine, ponytail, dogfooding"
era: V3R6
tier: M
---

# SPEC-SIMPLICITY-LADDER-001 — Simplicity Decision Ladder + @MX:DEBT Deferred-Simplification Tag

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-22 | manager-spec | Initial plan-phase draft. Inspired by `DietrichGebert/ponytail` v4.7.0 (MIT) "lazy senior dev" minimalist-coding skill. Two REQs bundled per user approval: REQ-1 (ordered simplicity decision ladder absorption) + REQ-2 (new `@MX:DEBT` tag type for deliberate-simplification debt). Tier M because REQ-2 touches `internal/mx/` scanner validity gate (see §B.3). |

---

## §A. Background

### A.1 Source of inspiration (ponytail)

`DietrichGebert/ponytail` (v4.7.0, MIT) is a "lazy senior dev" minimalist-coding enforcement skill. It productized — and benchmarked — the same "Enforce Simplicity" principle MoAI already preaches in `moai-constitution.md` § Agent Core Behaviors #4 and `karpathy-quickref.md` "Simplicity First". The review report on ponytail was presented to the user, who explicitly approved absorbing two of its operational mechanisms into MoAI doctrine.

ponytail contributed two discrete operational artifacts that MoAI does not currently have as named mechanisms:

1. an **ORDERED decision ladder** applied *before* writing code (cheapest-capability-first); and
2. a **deferred-simplification debt ledger** — an inline marker that records a deliberate, working simplification together with its ceiling and its upgrade trigger.

This SPEC absorbs both into MoAI doctrine. It is doctrine/doc-layer work with one small runtime-code change (REQ-2's tag-type registration — see §B.3).

### A.2 Why MoAI already half-has this (the honest delta)

MoAI is not starting from zero on simplicity:

- `moai-constitution.md` § Agent Core Behaviors **#4 Enforce Simplicity** already carries prose ("resist overcomplexity", "would a staff engineer say 'why didn't you just…'") and a **3× quantitative LOC trigger**. What it lacks is an **ordered, dependency-avoidance decision ladder** — the explicit "stdlib → native → already-installed dep → one line" ordering.
- `karpathy-quickref.md` **"Simplicity First"** already has checkpoint questions ("Can this be done in fewer lines? 3× LOC trigger", "Are these abstractions earning their complexity?"). What it lacks is the *dependency-avoidance ordering* axis — its questions are about LOC and abstraction, not about *which capability source to reach for first*.

So REQ-1's **net-new contribution is the explicit ordered dependency-avoidance ladder**, NOT the idea of simplicity. The ladder is positioned as a *complement* to the existing two surfaces, not a duplicate.

- The existing @MX taxonomy (`mx-tag-protocol.md`) is **NOTE / WARN / ANCHOR / TODO / LEGACY**. None of these records a *deliberate, working* simplification with a named ceiling + upgrade trigger. `@MX:TODO` is the nearest neighbor — REQ-2's central challenge (resolved in `plan.md` §B) is justifying why `@MX:DEBT` is not redundant with `@MX:TODO`.

### A.3 What this SPEC is NOT

This SPEC is doctrine absorption, not a new subsystem. It introduces NO new config file, NO new lint rule, NO intensity-level system (lite/full/ultra — those were explicitly rejected in the source analysis as redundant with the existing `harness.yaml` minimal/standard/thorough routing), and NO new enforcement hook. The only runtime code change is REQ-2's tag-type registration, which is strictly necessary because the `internal/mx/` scanner hard-rejects unknown tag kinds (§B.3). The SPEC about simplicity must not itself over-engineer (the irony guard — `plan.md` §G).

---

## §B. Research Findings

The four required research items were resolved at plan-authoring time. Each finding is backed by an actual Read/Grep against the current tree (verification-claim-integrity §1.1 — observed, not assumed).

### B.1 REQ-1 insertion point (moai-constitution.md)

`### 4. Enforce Simplicity` lives inside the `<!-- moai:evolvable-start id="agent-core-behaviors" -->` block of `.claude/rules/moai/core/moai-constitution.md`. Today it contains: a one-line directive, three "questions to ask", a TRUST-5 cross-reference, an anti-pattern line, and the 3× LOC quantitative trigger. There is no ordered ladder. The ladder is inserted as a new sub-block immediately after the existing "Questions to ask" list and before the "Quantitative trigger" paragraph, so the ordering (capability-source ladder → then LOC ratio check) reads naturally.

### B.2 REQ-1 complement-not-duplicate confirmation (karpathy-quickref.md)

`karpathy-quickref.md` "Simplicity First" checkpoint questions are LOC-and-abstraction oriented. The ladder's dependency-avoidance ordering is a distinct axis. REQ-1 adds ONE cross-reference line to karpathy-quickref pointing at the constitution ladder; it does NOT restate the ladder there (single source of truth — the ladder lives in the constitution; karpathy-quickref points to it).

### B.3 REQ-2 code-touch determination (decisive for Tier)

There IS a dedicated `internal/mx/` package. The tag-type set is enumerated as Go constants in `internal/mx/tag.go`:

- `internal/mx/tag.go` — `TagKind` constants: `MXNote`, `MXWarn`, `MXAnchor`, `MXTodo`, `MXLegacy` (5 kinds).
- `internal/mx/scanner.go:247` — a **hard validity gate**: `case MXNote, MXWarn, MXAnchor, MXTodo, MXLegacy:` with `default: return Tag{}, fmt.Errorf("unknown tag kind: %s", ...)`. **A `@MX:DEBT` tag in source would be REJECTED as "unknown tag kind" unless `MXDebt` is registered here.**
- `internal/mx/resolver_query.go:183-187` — a queryable-kinds allowlist map (`MXNote: true … MXLegacy: true`).
- `internal/cli/mx_query.go:125-133` — the CLI string→kind mapping for `moai mx query`.

> **Verification-claim-integrity §1.1 surface 3 (defect/feature claim)**: the "scanner hard-rejects unknown kinds" claim above is backed by an actual Read of `internal/mx/scanner.go:240-251`, not inferred from grep. The observed switch statement's `default` branch returns `"unknown tag kind"`.

Consequence: REQ-2 is **NOT doc-only**. Registering `@MX:DEBT` minimally requires editing `internal/mx/tag.go` (+ const), `internal/mx/scanner.go` (validity case), `internal/mx/resolver_query.go` (queryable map), and `internal/cli/mx_query.go` (CLI mapping), each with TDD test coverage. **This raises the SPEC to Tier M.** (Full Tier rationale: `plan.md` §D.)

### B.4 REQ-2 house-format study (mx-tag-protocol.md)

`mx-tag-protocol.md` defines each tag with: a one-line purpose, "When to Add", "When to Update", "When to Remove", a Tag Lifecycle Rules block, and mandatory-field notes. `@MX:DEBT` must match this structure. The `@MX:CEILING` and `@MX:UPGRADE` sub-lines (REQ-2's ceiling + upgrade-trigger) extend the existing sub-line vocabulary (`@MX:SPEC`, `@MX:LEGACY`, `@MX:REASON`, `@MX:TEST`, `@MX:PRIORITY`).

### B.5 REQ-2 sub-line scanner mechanism (D1/D2 — backed by an actual Read)

> **Verification-claim-integrity §1.1 surface 3**: the two claims below are backed by my own Read of `internal/mx/scanner.go:60-110` + `:235-251` and `scanner_test.go:150-175` — observed in this run, not relayed assumption.

The scanner's per-line loop calls `parseTag` (scanner.go:80) on EVERY line containing the literal `@MX:`. `parseTag` (scanner.go:235-251) splits on the first colon, upper-cases the first segment, and runs it through a `switch` whose `default` branch returns `"unknown tag kind: <X>"`. On that error, the loop (scanner.go:81-84) appends to `s.errors` and `continue`s. The REASON-sub-line handler at scanner.go:95 is reached ONLY after the line 80 `parseTag` call succeeds.

Two consequences observed:

1. **Registering only `MXDebt` is insufficient for REQ-2.2.** A standalone `// @MX:CEILING: < 10k entries` line → `parseTag("CEILING: …")` → `default` → `"unknown tag kind: CEILING"` error → `s.errors` + `continue`. So the `@MX:CEILING` / `@MX:UPGRADE` sub-lines introduced by REQ-2.2 would themselves regenerate parse errors even after `MXDebt` is registered. REQ-2.4 (below, revised) names the concrete fix: a **recognized-sub-line-kind set** that `parseTag` consults BEFORE its `default` branch, so a recognized sub-line is skipped (returns a sentinel "sub-line, not a tag" result) rather than erroring.

2. **Latent pre-existing `@MX:REASON` bug — surfaced, NOT repaired here (scope discipline, D2).** The same flow means a standalone `// @MX:REASON: …` line ALSO hits `parseTag`'s `default` (kind `"REASON"` is not in the validity switch) → error → `continue` BEFORE the line-95 REASON handler runs. The cited test `TestScanFileWithWarnReason` (scanner_test.go:170) guards its REASON assertion with `&& len(tags) > 0`, so it passes **vacuously** when the file yields 0 tags. This is a real pre-existing defect, but **repairing the general REASON path is OUT OF SCOPE for this SPEC** (see §J "Out of Scope — latent REASON-path repair"). It is forward-linked to a separate follow-up: **`SPEC-MX-SUBLINE-PARSE-REPAIR-001`** (proposed, NOT authored here). IF the minimal recognized-sub-line-kind-set mechanism chosen for REQ-2.4 happens to also make `@MX:REASON` scan cleanly (because REASON joins the same recognized-sub-line set), that is recorded as an **incidental benefit** — but this SPEC's REQ-2.4 acceptance is scoped to CEILING/UPGRADE only; the REASON vacuous-test repair belongs to the follow-up SPEC.

---

## §C. Requirements (GEARS)

### REQ-1 — Ordered Simplicity Decision Ladder

**REQ-1.1 (Ubiquitous).** The constitution § Agent Core Behaviors #4 Enforce Simplicity **shall** present an ordered, language-neutral, 6-rung simplicity decision ladder applied before writing code.

**REQ-1.2 (Ubiquitous).** The ladder **shall** be phrased generically across all 16 supported languages — it **shall not** reference any single language's standard library, package manager, or platform feature by name (no JS/Python bias).

**REQ-1.3 (Where — capability gate).** **Where** a "never simplify away" safety carve-out applies, the ladder **shall** name the non-negotiable boundary: the ladder **shall not** be used to cut input validation at trust boundaries, error handling that prevents data loss, security measures, accessibility, or one runnable check behind non-trivial logic. The carve-out **shall** cross-reference existing safety rules (TRUST 5 Secured + the Bash risk-amplifier doctrine in `coding-standards.md`) rather than duplicating them.

**REQ-1.4 (Ubiquitous).** The `karpathy-quickref.md` "Simplicity First" section **shall** carry one cross-reference line pointing at the constitution ladder, framing the ladder as the dependency-avoidance ordering axis complementing the existing LOC/abstraction checkpoint questions. It **shall not** restate the ladder (single source of truth).

The canonical 6-rung ladder (adapted from ponytail to MoAI language-neutral voice):

```
1. Does this need to be built at all? (YAGNI)
2. Does the standard library do this? Use it.
3. Does a native platform feature cover it? Use it.
4. Does an already-installed dependency solve it? Use it.
5. Can this be one line? Make it one line.
6. Only then: write the minimum code that works.
```

### REQ-2 — `@MX:DEBT` Deferred-Simplification Tag

**REQ-2.1 (Ubiquitous).** The `@MX` taxonomy in `mx-tag-protocol.md` (and the @MX list in `moai-constitution.md` § MX Tag Quality Gates) **shall** define a new `@MX:DEBT` tag type that records a deliberate, working simplification.

**REQ-2.2 (Ubiquitous).** An `@MX:DEBT` marker **shall** carry two sub-lines: `@MX:CEILING` (the named limit the simplification is valid up to) and `@MX:UPGRADE` (the trigger condition for revisiting the simplification). The marker is an inline source comment, NOT a separate JSON ledger.

**REQ-2.3 (When — event-detected).** **When** an `@MX:DEBT` marker is harvested via `moai mx query --kind DEBT --json` and the marker lacks an `@MX:UPGRADE` sub-line, the JSON object for that marker **shall** carry the field `"rotRisk": "no-trigger"`. A marker WITH an `@MX:UPGRADE` sub-line **shall** carry `"rotRisk": ""` (empty string, omittable via `omitempty`). This is the exact output contract: the literal token `no-trigger` in a `rotRisk` JSON field is the binary-testable PASS signal (a debt with no upgrade trigger has no exit condition and silently rots). The `Tag` struct in `internal/mx/tag.go` gains one field `RotRisk string \`json:"rotRisk,omitempty"\`` populated during scan; no separate harvest surface is introduced.

**REQ-2.4 (Ubiquitous).** The `internal/mx/` scanner **shall** recognize `@MX:DEBT` as a valid tag kind AND recognize `@MX:CEILING` / `@MX:UPGRADE` as valid **sub-line kinds** (not standalone tags), so that neither the `@MX:DEBT` tag nor its `@MX:CEILING` / `@MX:UPGRADE` sub-lines are rejected as "unknown tag kind". The minimal mechanism (per §B.5):
- Register `MXDebt` in `internal/mx/tag.go` `TagKind` constants.
- Add a **recognized-sub-line-kind set** — the package-level set `map[string]bool{"CEILING": true, "UPGRADE": true, "REASON": true, "SPEC": true, "TEST": true, "PRIORITY": true}` — that `parseTag` (scanner.go:235-251) consults BEFORE its `default` branch: when the parsed kind is a recognized sub-line, `parseTag` returns a sentinel "this is a sub-line, not a tag" result that the per-line loop skips WITHOUT appending to `s.errors`. This replaces the current behavior where any non-tag-kind line errors.
  - **`LEGACY` is DELIBERATELY EXCLUDED from this set (D-NEW-1).** Although `mx-tag-protocol.md:26` lists `@MX:LEGACY` under "Sub-lines:", `LEGACY` is ALSO a real standalone **tag kind** — `MXLegacy TagKind = "LEGACY"` (tag.go:25), queryable (`resolver_query.go:187`), CLI-mapped (`mx_query.go:133`), documented as a Tag Type (`moai-constitution.md` § MX Tag Quality Gates). If `LEGACY` were in the recognized-sub-line set, `parseTag` would skip a standalone `// @MX:LEGACY:` comment as a sub-line sentinel and **silently drop it from the tag list** — a regression in existing behavior. `SPEC`/`TEST`/`PRIORITY` are verified safe (they match NO tag kind; grep `TagKind = "(SPEC|TEST|PRIORITY)"` → 0). The doctrine dual-classification of `@MX:LEGACY` (tag-kind in code vs "Sub-lines:" in `mx-tag-protocol.md:26`) is the D-CARRY ambiguity — its doctrine clarification is left to the forward-linked `SPEC-MX-SUBLINE-PARSE-REPAIR-001` (NOT fixed in this SPEC, scope discipline).
- Accept `MXDebt` in the `scanner.go` tag-validity case, include it in the `resolver_query.go` queryable-kinds map, and map it in `cli/mx_query.go`.
- Populate the `RotRisk` field (REQ-2.3) when a `@MX:DEBT` tag's following lines contain `@MX:CEILING` but no `@MX:UPGRADE`.

> **Incidental benefit (NOT a REQ-2 acceptance target).** Adding `REASON` to the recognized-sub-line set above will, as a side effect, make a standalone `@MX:REASON` line scan without error — partially mitigating the latent bug surfaced in §B.5. This is incidental; the REASON vacuous-test repair (scanner_test.go:170) remains owned by the follow-up `SPEC-MX-SUBLINE-PARSE-REPAIR-001`, NOT this SPEC. REQ-2.4's acceptance is verified on CEILING/UPGRADE only.

**REQ-2.5 (Ubiquitous).** The lifecycle of `@MX:DEBT` **shall** be documented as: add (when a deliberate simplification is made) → grep-harvest (e.g. `grep -rnE '@MX:DEBT'` or `moai mx query --kind DEBT`) → resolve (remove the marker when its `@MX:UPGRADE` trigger fires). `@MX:DEBT` is distinct from `@MX:TODO`: TODO = incomplete work resolved in the GREEN phase; DEBT = a complete, working simplification with a named ceiling + upgrade trigger that may legitimately persist across many GREEN phases.

> **The @MX:TODO-vs-@MX:DEBT justification verdict is owned by `plan.md` §B.** The verdict there is: the distinction DOES justify a new tag (not an extension of @MX:TODO) — see plan.md for the full reasoning and the falsification check.

---

## §J. Exclusions

This section enumerates what this SPEC deliberately does NOT build. Keeping it small is itself an application of REQ-1 (the irony guard).

### Out of Scope — intensity-level system
- No `lite` / `full` / `ultra` simplicity-enforcement intensity levels. The source analysis explicitly REJECTED these as redundant with the existing `harness.yaml` minimal/standard/thorough quality-depth routing. Adding a parallel intensity axis would duplicate an existing knob.

### Out of Scope — new config file or lint rule
- No new `.moai/config/sections/*.yaml` file for the ladder or the debt tag. No new `internal/spec/lint.go` rule. The ladder is doctrine prose; the debt tag reuses the existing `internal/mx/` scanner machinery (only the tag-kind set expands).

### Out of Scope — enforcement hook for the ladder
- No PreToolUse/PostToolUse hook that mechanically blocks code violating the ladder. The ladder is a doctrine-level decision aid (a SHOULD for agents), consistent with how Behavior #4 is already enforced (by being read, not by a gate). A mechanical ladder-enforcement hook is a separate, larger concern.

### Out of Scope — separate JSON debt ledger
- `@MX:DEBT` is an inline source comment harvested by grep / `moai mx query`, NOT a separate `.moai/state/debt-ledger.json` file. A separate persisted ledger would duplicate the source-of-truth (the source comment) and add a sync-drift surface.

### Out of Scope — retroactive debt tagging sweep
- No project-wide sweep to retroactively add `@MX:DEBT` markers to existing simplifications in the codebase. The tag is available for new use going forward; back-filling existing code is not part of this SPEC.

### Out of Scope — ponytail benchmark port
- ponytail ships a benchmark harness for its minimalist-coding enforcement. This SPEC absorbs the two *mechanisms* (ladder + debt ledger), NOT the benchmark. Porting a benchmark is a separate effort with its own justification burden.

### Out of Scope — latent REASON-path repair
- §B.5 surfaces a pre-existing defect: standalone `@MX:REASON` sub-lines hit `parseTag`'s `default` branch and error before the REASON handler runs, and the cited test `TestScanFileWithWarnReason` (scanner_test.go:170) passes vacuously. The general REASON-path repair (including fixing the vacuous test guard) is **forward-linked to `SPEC-MX-SUBLINE-PARSE-REPAIR-001`** (proposed, not authored here) and is NOT a deliverable of this SPEC. This SPEC only adds the recognized-sub-line-kind set needed to make `@MX:CEILING` / `@MX:UPGRADE` scan cleanly; any improvement to REASON scanning is incidental (REQ-2.4 note), not an acceptance target. Authoring the follow-up SPEC is also out of scope — it is a proposal recorded for the orchestrator.

---

## §H. Cross-References

- `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors #4 Enforce Simplicity (REQ-1 insertion point) + § MX Tag Quality Gates (REQ-2 @MX list).
- `.claude/rules/moai/development/karpathy-quickref.md` "Simplicity First" (REQ-1.4 cross-reference target).
- `.claude/rules/moai/workflow/mx-tag-protocol.md` (REQ-2 @MX SSOT — `@MX:DEBT` definition lands here).
- `.claude/rules/moai/development/coding-standards.md` § Bash Risk-Amplifier Doctrine + CLAUDE.md TRUST 5 Secured (REQ-1.3 safety carve-out cross-reference targets — do NOT duplicate).
- `internal/mx/tag.go` (REQ-2.4 `MXDebt` const + REQ-2.3 `RotRisk` field), `internal/mx/scanner.go:80-110` + `:235-251` (recognized-sub-line-kind set in `parseTag` + `MXDebt` validity case + `RotRisk` population), `internal/mx/resolver_query.go:183-187` (queryable-kinds map), `internal/cli/mx_query.go:125-133` (CLI kind mapping). Each with TDD coverage.
- `SPEC-MX-SUBLINE-PARSE-REPAIR-001` (PROPOSED, not authored here) — owns the general `@MX:REASON` sub-line-path repair + the `scanner_test.go:170` vacuous-test fix surfaced in §B.5. Forward-link only.
- Template mirrors: every `.claude/rules/**` edit mirrors to `internal/template/templates/.claude/rules/**` (Template-First, CLAUDE.local.md §2). The `mx-tag.md` skill reference at `internal/template/templates/.claude/skills/moai/references/mx-tag.md` is also a mirror target for the @MX:DEBT definition.
- `DietrichGebert/ponytail` v4.7.0 (MIT) — source of inspiration for the ladder + debt ledger mechanisms (cited as inspiration, wording adapted to MoAI voice).

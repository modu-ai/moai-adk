---
id: SPEC-SIMPLICITY-LADDER-002
title: "In-Codebase Reuse Rung — Simplicity Decision Ladder 6→7-Rung Completion"
version: "0.2.0"
status: draft
created: 2026-06-30
updated: 2026-06-30
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai"
lifecycle: spec-anchored
tags: "simplicity, decision-ladder, dry, reuse, doctrine, ponytail, dogfooding"
era: V3R6
tier: S
---

# SPEC-SIMPLICITY-LADDER-002 — In-Codebase Reuse Rung (Simplicity Decision Ladder 6→7-Rung Completion)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-30 | manager-spec | Initial plan-phase draft (Tier S, doctrine-only). Follow-up to the completed SPEC-SIMPLICITY-LADDER-001, which ported the `DietrichGebert/ponytail` (MIT) simplicity ladder into `moai-constitution.md`. Comparison against the upstream 7-rung ladder revealed MoAI omitted ponytail's rung 2 ("Already in this codebase? … reuse it"). This SPEC inserts the missing in-codebase-reuse / DRY rung at position 2, completing the ladder from 6 to 7 rungs in both the LIVE rule and its template mirror. |
| 0.2.0 | 2026-06-30 | manager-spec | Plan-audit revision (PASS-WITH-DEBT 0.84 → addressing 3 findings). **D1**: `karpathy-quickref.md` line ~33 (LIVE + template mirror) carries an INLINE arrow enumeration of the ladder + the "dependency-avoidance" framing — a third drift surface the v0.1.0 draft wrongly excluded. Brought into scope as REQ-5 (lead the enumeration with in-codebase reuse + reuse-inclusive framing); edit surface is now a 4-file parallel edit (2 file pairs). **D2**: corrected the embed mechanism — `internal/template/embedded.go` does NOT exist; embedding is directive-based (`//go:embed all:templates`, `internal/template/embed.go:28`), so the verification gate is `go test ./internal/template/...` + `go build ./...`, NOT an `embedded.go` regen. `make build` = `gen-catalog-hashes --all` + `go build`. **D3**: added per-rung content grep anchors to the renumber AC (carried-over rungs anchored by content, not just count). |

---

## §A. Background

### A.1 What SPEC-SIMPLICITY-LADDER-001 shipped (the baseline)

SPEC-SIMPLICITY-LADDER-001 (completed) absorbed two `DietrichGebert/ponytail` (MIT, https://github.com/DietrichGebert/ponytail) mechanisms into MoAI doctrine: an ordered simplicity decision ladder and a `@MX:DEBT` deferred-simplification tag. The ladder landed in `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors #4 Enforce Simplicity as a **6-rung** ordered list:

```
1. Does this need to be built at all? (YAGNI)
2. Does the standard library do this? Use it.
3. Does a native platform feature cover it? Use it.
4. Does an already-installed dependency solve it? Use it.
5. Can this be one line? Make it one line.
6. Only then: write the minimum code that works.
```

### A.2 The omission (verified by direct comparison)

The upstream ponytail SKILL.md ladder has **7 rungs**, not 6. Comparing the two ladders rung-by-rung shows MoAI dropped ponytail's rung 2 during the v001 port:

| ponytail upstream rung | MoAI v001 rung | Status |
|------------------------|-----------------|--------|
| 1. Does this need to exist at all? (YAGNI) | 1. YAGNI | present |
| **2. Already in this codebase? A helper, util, type, or pattern that already lives here → reuse it.** | **(none)** | **MISSING** |
| 3. Stdlib does it? Use it. | 2. standard library | present (renumber) |
| 4. Native platform feature covers it? | 3. native platform | present (renumber) |
| 5. Already-installed dependency solves it? | 4. installed dependency | present (renumber) |
| 6. Can it be one line? | 5. one line | present (renumber) |
| 7. Only then: the minimum code that works. | 6. minimum code | present (renumber) |

The missing rung is the **in-codebase reuse / DRY** rung: before reaching for the language's standard library, ask whether a helper, util, type, or pattern already exists in THIS project and reuse it. This is the highest-value omission to repair — it prevents reimplementing code that already lives in the project.

### A.3 Why the missing rung is load-bearing for MoAI specifically

The in-codebase-reuse rung codifies a practice MoAI work patterns ALREADY enforce informally:
- "grep the whole repo before retiring/reimplementing" (a recurrent MoAI lesson — retire-before-grep discipline).
- reuse-first sourcing (prefer extending an existing helper over writing a parallel one).

The v001 ladder, by jumping straight from YAGNI (rung 1) to the language standard library (rung 2), skips the cheapest capability source of all: code that already exists in the project tree (zero new code, zero new dependency, zero new stdlib surface to learn). Inserting the rung makes the ladder's cheapest-capability-first ordering correct.

### A.4 What this SPEC is NOT

This is a Tier S doctrine-only insertion. It introduces NO Go code change, NO new config file, NO new lint rule, NO new @MX tag, NO new enforcement hook. It edits two ladder-bearing surfaces and their template mirrors (4 markdown files): the constitution ladder block + framing (REQ-1/2/3) and the karpathy-quickref cross-reference line (REQ-5). The SPEC about simplicity must not itself over-engineer (the irony guard — `plan.md` §G).

---

## §B. Research Findings

Each finding is backed by an actual Read/Grep against the current tree (verification-claim-integrity §1.1 — observed, not assumed).

### B.1 Edit-target line locations (LIVE + template, byte-identical)

`grep -n` against both files confirms the ladder block sits at the SAME line numbers in the LIVE rule and the template mirror, and the two are byte-identical at every ladder line:

- LIVE: `.claude/rules/moai/core/moai-constitution.md`
  - L238: intro line `Simplicity decision ladder (apply in order, before writing code — cheapest capability first):`
  - L240-245: the 6 numbered rungs
  - L247: framing prose `The ladder is the dependency-avoidance ordering axis — …`
  - L249: safety carve-out `Never simplify away (safety carve-out): …` (PRESERVE verbatim)
  - L251: quantitative trigger `Quantitative trigger: If implementation exceeds 3x …` (PRESERVE verbatim)
- TEMPLATE MIRROR: `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` — identical content at L238-251 (verified via parallel `grep -n`).

> Verification-claim-integrity §1.1: the byte-identical claim is backed by a parallel `grep -n` of both files producing identical output for L240-251 in this run, not inferred.

### B.2 Framing-prose distinction (the one-clause assessment)

The L247 framing currently reads: *"The ladder is the dependency-avoidance ordering axis — reach for the cheapest existing capability before adding new code or a new dependency."* The newly inserted rung 2 is **in-codebase reuse / DRY**, which is conceptually distinct from "dependency avoidance" — it is about reusing what already exists in the project, not about avoiding a new external dependency. Leaving the framing as "dependency-avoidance ordering axis" while a reuse rung sits at position 2 creates a doctrine inconsistency. REQ-3 corrects the framing with one minimal, language-neutral clause (see §C).

### B.3 Mirror surface scope (TWO file pairs — constitution + karpathy)

The full rung-by-rung ladder text lives in exactly two places — the LIVE constitution rule and its template mirror. BUT v001's REQ-1.4 cross-reference in `karpathy-quickref.md` is NOT just a generic pointer: line 33 (verified by Read in this run, byte-identical in LIVE and template) carries an INLINE ORDERED arrow enumeration of the ladder's capability sources plus the "dependency-avoidance" framing:

> "For the ordered dependency-avoidance decision ladder (stdlib → native platform feature → installed dependency → one line → minimum code, applied before writing code), see `… § Agent Core Behaviors #4 Enforce Simplicity`. The ladder is the capability-source ordering axis that complements these LOC/abstraction checkpoint questions."

Two problems this creates after the constitution edit: (a) the arrow enumeration OMITS the new cheapest capability source (in-codebase reuse), directly contradicting §A.3 / REQ-1's claim that reuse ranks first; (b) it uses "dependency-avoidance decision ladder", the exact framing REQ-3 broadens because a reuse rung makes that framing too narrow. Therefore `karpathy-quickref.md` is a THIRD drift surface and IS in scope (REQ-5) — a surgical one-line edit (lead the enumeration with in-codebase reuse; adopt the reuse-inclusive framing), NOT a full ladder restatement.

So this SPEC's edit surface is TWO file pairs (4 files): the constitution LIVE + template (REQ-1/2/3) and `karpathy-quickref.md` LIVE + template (REQ-5). Still NO skill-reference or docs-site copy is touched (distinguishing this from v001, which also touched `skills/moai/references/mx-tag.md`).

> Verification-claim-integrity §1.1: the v0.1.0 "no third copy / needs no edit" claim was an unobserved assertion — corrected here after an actual Read of `karpathy-quickref.md:33` (LIVE + template) in this run.

### B.4 Template neutrality precedent

The existing ladder block in the template mirror carries NO ponytail citation and NO internal tokens — it is pure generic prose. The v001 ladder edit was already template-neutral. This SPEC's insertion is the same generic prose register (the new rung names only "helper, util, type, or pattern" — generic across all 16 supported languages), so the template mirror stays neutral. The ponytail citation lives ONLY in these SPEC artifacts (`.moai/specs/`), which are NOT under `internal/template/templates/` and are therefore exempt from the neutrality CI guard.

---

## §C. Requirements (GEARS)

### REQ-1 — Insert the in-codebase-reuse rung at position 2 (LIVE)

**REQ-1.1 (Ubiquitous).** The constitution § Agent Core Behaviors #4 Enforce Simplicity ladder **shall** be a **7-rung** ordered list, with a new in-codebase-reuse rung inserted at **position 2** (between YAGNI at rung 1 and the standard-library rung), and rungs 3-7 **shall** be the v001 rungs 2-6 renumbered in place (standard library → 3, native platform → 4, installed dependency → 5, one line → 6, minimum code → 7).

**REQ-1.2 (Ubiquitous).** The new rung **shall** be phrased in the existing MoAI house cadence ("Does X? <imperative>.") and **shall** be language-neutral — it **shall not** reference any single language's standard library, package manager, type system, or platform feature by name. The canonical new rung text:

> `2. Does a helper, util, type, or pattern already exist in this codebase? Reuse it.`

The resulting canonical 7-rung ladder:

```
1. Does this need to be built at all? (YAGNI)
2. Does a helper, util, type, or pattern already exist in this codebase? Reuse it.
3. Does the standard library do this? Use it.
4. Does a native platform feature cover it? Use it.
5. Does an already-installed dependency solve it? Use it.
6. Can this be one line? Make it one line.
7. Only then: write the minimum code that works.
```

**REQ-1.3 (Ubiquitous — preservation).** The edit **shall** preserve verbatim the safety carve-out paragraph ("Never simplify away …", L249) and the quantitative 3x-LOC trigger ("Quantitative trigger: …", L251). The intro line (L238) and the safety/trigger paragraphs **shall not** be reworded by this SPEC.

### REQ-2 — Mirror to the template (byte-parity)

**REQ-2.1 (Ubiquitous).** The same edit **shall** be applied to the template mirror `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` so that the LIVE rule and the template mirror remain byte-parallel (per CLAUDE.local.md §2 Template-First Rule).

**REQ-2.2 (When — event-detected).** **When** the template mirror is edited, `make build` **shall** be run so the change is picked up by the directive-based embed and verified. The template is embedded via `//go:embed all:templates` (`internal/template/embed.go:28`) — there is NO generated `embedded.go` file to regenerate; the directive embeds the live template tree at compile time. `make build` runs `gen-catalog-hashes --all` (which regenerates the catalog hash for the changed template file) then `go build`. The verification gate is therefore `go test ./internal/template/...` passing AND `go build ./...` succeeding — NOT an `embedded.go` diff.

### REQ-3 — Framing-prose adjustment (encompass reuse)

**REQ-3.1 (Where — capability gate).** **Where** the L247 framing prose names the ladder a "dependency-avoidance ordering axis", the prose **shall** be adjusted with one minimal, language-neutral clause so the framing encompasses in-codebase reuse (rung 2), not only dependency avoidance. The adjustment **shall not** restate the rungs and **shall** keep the existing "language-neutral … standard library / native platform feature …" sentence intact. Recommended minimal form (run-phase implementer's exact wording is bounded by this constraint):

> `The ladder is the reuse-and-dependency-avoidance ordering axis — reach for the cheapest existing capability before adding new code or a new dependency; what already lives in this codebase (rung 2) ranks first because reusing it is cheaper than the standard library, the platform, or a dependency. It is language-neutral: …`

### REQ-4 — Byte-parity + template neutrality verification

**REQ-4.1 (Ubiquitous).** After the edits, a per-file `diff -q` between each LIVE file and its template mirror **shall** return clean (exit 0) — byte-parity maintained for BOTH file pairs: `moai-constitution.md` (LIVE vs template) AND `karpathy-quickref.md` (LIVE vs template).

**REQ-4.2 (When — event-detected).** **When** each edited template mirror (`moai-constitution.md` AND `karpathy-quickref.md`) is scanned for internal-content leaks, the scan **shall** find none of: an internal SPEC ID (`SPEC-SIMPLICITY-LADDER-002`), any REQ/AC token (`REQ-`/`AC-`), an internal date, a commit SHA, or a `/Users/` path — in the inserted/edited content. (Per CLAUDE.local.md §25 + §15.)

### REQ-5 — karpathy-quickref.md cross-reference line (LIVE + template mirror)

**REQ-5.1 (Ubiquitous).** The `karpathy-quickref.md` "Simplicity First" cross-reference line (line ~33) **shall** be updated so its inline arrow enumeration of the ladder's capability sources LEADS with in-codebase reuse: `in-codebase reuse → stdlib → native platform feature → installed dependency → one line → minimum code`. The enumeration **shall not** omit in-codebase reuse (consistency with REQ-1: reuse is the cheapest capability source and ranks first).

**REQ-5.2 (Where — capability gate).** **Where** the same line names the ladder a "dependency-avoidance decision ladder", the framing **shall** be broadened to a reuse-inclusive form (e.g. "reuse-and-dependency-avoidance decision ladder"), consistent with REQ-3. The existing "capability-source ordering axis that complements these LOC/abstraction checkpoint questions" tail **shall** be retained (it is already reuse-inclusive).

**REQ-5.3 (Ubiquitous — surgical scope).** The edit **shall** touch ONLY this one cross-reference line. It **shall not** restate the full rung-by-rung ladder in karpathy-quickref (single source of truth: the constitution holds the rung-by-rung ladder; karpathy points at it). The same edit **shall** be mirrored to `internal/template/templates/.claude/rules/moai/development/karpathy-quickref.md` (byte-parallel).

---

## §J. Exclusions

This section enumerates what this SPEC deliberately does NOT build. Keeping it small is itself an application of the ladder (the irony guard).

### Out of Scope — Go code / @MX changes
- No `internal/mx/` change, no new `@MX` tag type, no scanner edit. This is a pure doctrine/markdown insertion. (v001 already shipped the `@MX:DEBT` machinery; this SPEC touches none of it.)

### Out of Scope — new enforcement mechanism
- No new lint rule, no PreToolUse/PostToolUse hook, no config knob to mechanically enforce the reuse rung. The ladder is a doctrine-level decision aid (a SHOULD for agents), enforced by being read — identical to how the other six rungs are enforced.

### Out of Scope — skill-reference / docs-site ladder restatement
- No restatement of the ladder in any skill reference or docs-site page. The rung-by-rung ladder's single source of truth remains the constitution rule + its template mirror (§B.3). NOTE: `karpathy-quickref.md` is NOT excluded — it carries an inline arrow enumeration + "dependency-avoidance" framing (§B.3) and IS in scope via REQ-5 (a surgical one-line edit, NOT a full ladder restatement).

### Out of Scope — rung rewording beyond the insertion
- The five carried-over rungs (stdlib / native / dependency / one line / minimum code) are renumbered ONLY; their text is not reworded. The intro line, safety carve-out, and 3x-LOC trigger are PRESERVED verbatim (REQ-1.3).

### Out of Scope — retroactive ladder-application audit
- No project-wide sweep to retroactively check existing code against the new reuse rung. The rung is doctrine going forward; auditing existing code for missed reuse is a separate effort.

### Out of Scope — ponytail re-port of other mechanisms
- This SPEC repairs ONLY the one omitted ladder rung. It does not re-examine ponytail for other mechanisms MoAI may lack (benchmark harness, etc. — already excluded in v001 §J).

---

## §H. Cross-References

- `.claude/rules/moai/core/moai-constitution.md` § Agent Core Behaviors #4 Enforce Simplicity (REQ-1/REQ-3 edit site — ladder block L238-251).
- `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` (REQ-2 template mirror — byte-parallel edit site).
- `.claude/rules/moai/development/karpathy-quickref.md` line ~33 (REQ-5 LIVE edit site — cross-reference line) + `internal/template/templates/.claude/rules/moai/development/karpathy-quickref.md` (REQ-5 template mirror).
- `internal/template/embed.go:28` (`//go:embed all:templates` — the directive that embeds the live template tree at compile time; there is NO generated `embedded.go`. REQ-2.2's `make build` runs `gen-catalog-hashes --all` + `go build`; the verification gate is `go test ./internal/template/...` + `go build ./...`).
- `.moai/specs/SPEC-SIMPLICITY-LADDER-001/` (completed predecessor — shipped the 6-rung ladder this SPEC completes to 7).
- `DietrichGebert/ponytail` (MIT) — source of the 7-rung upstream ladder; the omitted rung 2 is the subject of this SPEC. Cited as a public source (SPEC-artifact-only; not echoed into the template mirror — §B.4).
- CLAUDE.local.md §2 (Template-First mirror duty), §15 (16-language neutrality), §25 (template internal-content isolation) — REQ-2 / REQ-4 constraints.
- `.claude/rules/moai/development/coding-standards.md` § Bash Risk-Amplifier Doctrine + CLAUDE.md TRUST 5 Secured — the safety-rule cross-references inside the PRESERVED L249 carve-out (not edited by this SPEC; preserved verbatim).

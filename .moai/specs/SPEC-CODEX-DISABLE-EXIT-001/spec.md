---
id: SPEC-CODEX-DISABLE-EXIT-001
title: "moai skills disable --codex exit-code contract — boundary-intent adjudication and per-branch exit policy"
version: "0.1.0"
status: draft
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
era: V3R6
tags: "codex, skills-disable, exit-code, cli-contract, adjudication, caller-census, t548"
related_specs: [SPEC-CODEX-SKILL-DISABLE-001, SPEC-CODEX-SKILL-PATH-SLASH-001, SPEC-CODEX-GHOST-SKILLS-PRUNE-001]
---

# SPEC-CODEX-DISABLE-EXIT-001 — `moai skills disable --codex` exit-code contract

## HISTORY

- 2026-09-08 (plan-phase, v0.1.0) Initial authoring. Card t548, worktree `.claude/worktrees/t548`, branch `WT-codex-disable-exit`, base `a4855f0b2`. Tier S per card (Class C); artifact set is 4 files (acceptance.md included by dispatch order, exceeding the Tier S default of 2). The card's FIRST work item is a boundary-intent ADJUDICATION, not a repair — this SPEC is structured as a decision gate (M2) with both outcomes prescribed (§E) and a recommended default justified by the caller census (§B.3) + code reading (§B.1).

## §A. User Story

**As a** script or CI author who calls `moai skills disable <name> --codex`, **I want** the verb's exit code to tell me which of three things happened — the disable was performed, my request was refused, or there was nothing to act on — **so that** my script can branch on reality instead of reading stdout text.

**As a** MoAI-ADK maintainer, **I want** that contract to be an explicitly adjudicated design decision rather than an accident of branch ordering, **so that** changing it (or documenting it) is a deliberate act with the caller census as its cost side.

## §B. Context and Background

### §B.1 The measured branch table

All addresses read directly from this tree at `a4855f0b2`. The verb's real invocation form is `moai skills disable <name> --codex` (`internal/cli/skills.go:8`, `internal/cli/codex_skills_disable.go:387`) — the card's phrasing `moai codex skills disable` names the layer, not the CLI surface; that form was REJECTED at design time (`SPEC-CODEX-SKILL-DISABLE-001/plan.md:74`) and has zero invocations in-repo.

Runner `runCodexSkillDisable` (`internal/cli/codex_skills_disable.go:394-455`):

| # | Branch | Address | Stdout marker | Exit |
|---|--------|---------|---------------|------|
| B1 | Mirror absent | `:397-399` | `Nothing to disable: …` | **0** |
| B2 | Name unresolved / ambiguous | `:400-402` | `Refusing to write: …` | **non-zero** (`fmt.Errorf("cannot disable %q for Codex")`) |
| B3 | Codex home unresolved | `:406-408` | `Codex home does not resolve; nothing to disable` | **0** |
| B4 | Config absent / unreadable | `:412-414` | `… is absent or unreadable; nothing to disable` | **0** |
| B5 | **Unchanged** (entry already `enabled = false`) | `:419-421` ← upsert `:296-299` | `Unchanged: … (it is already disabled)` | **0** |
| B6 | **Skipped** (guard refused) | `:422-424` ← upsert `:232-236` | `Skipped: … (<reason>)` | **0** |
| B7 | Dry-run on would-write | `:431-434` | `[dry-run] Would …` | **0** |
| B8 | Backup write failure | `:440-441` | — | **non-zero** |
| B9 | Config write failure | `:450-451` | — | **non-zero** |
| B10 | Success | `:453-454` | `Disabled … for Codex` | **0** |

Command-level errors (non-zero, unchanged by this SPEC): project-root resolution failure (`internal/cli/skills.go:90`), missing `--codex` layer flag (`internal/cli/skills.go:96`), cobra flag/args validation.

The **Skipped** verdict fires on exactly five guard refusals, all inside `upsertCodexSkillDisable` (`:231-311`): (a) no path to write `:237-239`; (b) unencodable character in path — `"`, `\`, LF, CR — `:261-263`; (c) more than one entry declares the path `:282-284`; (d) first unrecognized line inside the entry's extent `:293-295`; (e) entry declares no key this verb can rewrite `:302-304`. Refusals (c) and (e) are actionable by a DIFFERENT verb (`moai clean --codex-skills`), which is why "request not performed" is load-bearing information for a script.

### §B.2 What is documented intent vs undocumented

The name-resolution axis (B1 vs B2) is **explicitly documented design**, on two surfaces:

- Doc comment, `internal/cli/codex_skills_disable.go:387-393`: *"Fail-open on absent inputs, exactly as the prune verb is: an unresolvable Codex home or an absent config says so and returns nil. A missing input is not an error — there is simply nothing to disable. A name that does not resolve is different: that is a typo, and it exits non-zero so a script can see it."*
- `CHANGELOG.md:20` (SPEC-CODEX-SKILL-DISABLE-001 close): *"The three name-resolution failures carry different exit codes on purpose: unresolvable and ambiguous exit non-zero (a typo must be detectable from a script or CI), while an absent mirror exits zero — the mirror is produced by a deployment run, not by a checkout, so its absence is an ordinary state."*

The **Unchanged / Skipped bucketing is undocumented** on every surface: neither the doc comment nor the CHANGELOG says what exit code "found but not applied" should carry. The doc comment's fail-open doctrine is scoped to **missing inputs** ("A missing input is not an error"), and both B5 and B6 fire on PRESENT input — the request was heard and answered. The card's rule 2 (Unchanged and Skipped must not land in the same bucket) is therefore in tension with the current code, not with any documented design.

### §B.3 Caller census (M1 — DONE; full commands and outputs in progress.md §E.1)

Searched with `rg` across the whole worktree at `a4855f0b2` (shell scripts, hook wrappers, templates `internal/template/templates/`, docs, CI `.github/workflows/`, Go tests, README ×4, docs-site ×4 locales):

| Surface | Hits | Detail |
|---------|------|--------|
| In-repo invocations of the actual verb | **1 script, 4 lines** | `.moai/reports/t502/e2e-verb.sh:19,41,43,68` — dev-only, one-time E2E evidence script from the verb's own card; not shipped, not a contract consumer |
| Go function-level callers | tests only | `internal/cli/codex_skills_disable_test.go` (10+ direct `runCodexSkillDisable` calls, 1 `newSkillsCmd()` at `:511`) — consume the returned `error` value, not process exit codes |
| Templates / hook wrappers | **0** | no hit in `internal/template/templates/` |
| CI workflows | **0** | no hit in `.github/workflows/` |
| docs-site (4 locales) | **verb undocumented** | `moai skills` → 0 hits; only `[[skills.config]]` shape docs exist (doctor.md ×4, moai-clean.md ×4) |
| README ×4 | **verb undocumented** | only `moai clean --codex-skills` is documented (line 749 ×4) |
| The card's literal form `moai codex skills disable` | 4 doc-only hits | all in `.moai/specs/SPEC-CODEX-SKILL-*/` describing the REJECTED form |

**Census verdict**: zero production in-repo callers. The exit-code contract exists for OUTSIDE-this-repo consumers (user scripts, downstream CI) and as a published CLI surface — and no published surface (docs-site, README, CHANGELOG exit-code table) currently documents the Skipped branch's exit code, so there is **no documented promise that Outcome B would break**. The census weakens the "silently breaks the contract" cost exactly as the card's mandatory investigation anticipated.

### §B.4 Sibling surface — corrected

The sibling with the same skip shape is **not** `moai skills prune` (that invocation form has 0 hits in-repo — it does not exist). It is **`moai clean --codex-skills`** (`runCleanCodexSkills`, `internal/cli/codex_skills_prune.go:157-221`), whose per-entry `Kept:` line is at `:195` in this tree. Its fail-open doctrine is documented in its own doc comment (`:159-162`): unresolvable home, absent/unreadable config, and zero entries all return nil. Per-entry `Kept:` verdicts never affect the exit code; only backup (`:210-211`) and write (`:216-217`) failures are non-zero.

## §C. Requirements (GEARS)

- **REQ-CDE-001** (Ubiquitous): The `moai skills disable <name> --codex` verb shall partition every terminal outcome into exactly three named classes — **performed** (write landed, or desired state already held), **refused** (request heard and declined by a guard), **absent-input** (nothing to act on) — and its exit-code contract shall state which class each of the ten branches in §B.1 belongs to.
- **REQ-CDE-002** (Capability gate): **Where** the M2 adjudication resolves to Outcome B (exit-code change), the Skipped outcome (B6) shall exit non-zero and the Unchanged outcome (B5) shall exit zero.
- **REQ-CDE-003** (Capability gate): **Where** the M2 adjudication resolves to Outcome A (document-only), the verb's exit-code matrix shall be documented verbatim on the verb's own surface, and the Unchanged/Skipped exit-code bucket sharing shall be recorded in this SPEC as a consciously accepted deviation from the card's distinctness rule (rule 2).
- **REQ-CDE-004** (Ubiquitous): The absent-input outcomes (B1 mirror absent, B3 home unresolved, B4 config absent/unreadable) shall exit zero — the documented fail-open doctrine is preserved under either adjudication.
- **REQ-CDE-005** (Ubiquitous): The name-resolution refusals (B2 unresolved, B2 ambiguous) shall keep exiting non-zero — the documented typo-detection contract is preserved under either adjudication.
- **REQ-CDE-006** (Event-driven): **When** any exit code changes (Outcome B adopted), each changed branch shall be covered by a test asserting its exit contract that is observed RED on the pre-change tree.
- **REQ-CDE-007** (Ubiquitous): The scope decision shall be recorded as **verb-local** — the `skills` verb family and `moai clean --codex-skills` keep independent exit-code policies — with the semantic distinction (single-target imperative vs multi-target sweep) documented in this SPEC (§D).
- **REQ-CDE-008** (Ubiquitous): The adjudicated exit-code contract shall be documented on the verb's `--help` Long text, so a script author can read the contract without reading Go source.

## §D. Scope Decision (resolved at plan phase — precedes the branch-level policy)

**Scope axis: this verb alone (verb-local), with the cross-verb distinction documented.** Rationale, from the code:

1. The `skills` verb family currently contains exactly ONE verb — `newSkillsCmd()` adds only `newSkillsDisableCmd()` (`internal/cli/skills.go:53-55`). "The family's exit-code contract" is a future-facing concern, not a present surface to align.
2. `moai clean --codex-skills` shares the fail-open *shape* but not the *semantics*: disable is a single-target imperative (one request, so refusal CAN be the verb's verdict), while clean is a multi-target sweep (a run that removes 3 of 5 ghosts and keeps 2 SHOULD exit 0; per-entry `Kept:` is reporting, not a verb-level verdict). Making a partially-successful sweep exit non-zero would be a worse contract, not a consistent one.
3. Documenting this distinction (REQ-CDE-007) closes the card's "fixing one verb creates NEW cross-verb inconsistency" hazard without dragging `clean` into scope.

The rejected alternative (family-wide contract alignment) is recorded as available if a future verb joins the `skills` tree; it is deliberately not this SPEC's work.

## §E. Adjudication Outcomes (both prescribed; M2 picks one)

| | **Outcome A — document-only** | **Outcome B — exit-code change (RECOMMENDED default)** |
|---|---|---|
| Claim | "Not found = failure / found but not applied = success" is intended design; the deliverable is documenting it | The documented design covers absent-inputs (fail-open) and name resolution (fail-closed); **Skipped is undocumented**, and "found but not applied" is two different events wearing one exit code |
| Prescription | Document the §B.1 matrix verbatim in `--help` Long text + CHANGELOG; record the Unchanged/Skipped bucket sharing as a consciously accepted deviation from card rule 2; no Go change | B6 (Skipped) returns a distinct error (`fmt.Errorf` wrapping the guard reason) → non-zero; B5 (Unchanged) stays 0; B1/B3/B4 stay 0; B2 stays non-zero; one-branch change |
| Why | Zero in-repo callers means changing nothing breaks nothing; name-resolution policy is already explicitly intended | Card rule 2 is a stated operator requirement the current code violates; the doc comment scopes fail-open to MISSING INPUTS only, so B's refusal→non-zero is MORE consistent with the documented design logic, not less; the two rule-2-actionable skip reasons ((c) duplicates, (e) no rewritable key) hand off to `moai clean --codex-skills`, and exit 0 hides that handoff from scripts; census found no documented exit-code promise for Skipped to break |
| Cost | Card rule 2 stays violated, now knowingly | Breaks any OUT-OF-REPO script that treats "Skipped" as success — census cannot see those; mitigated by documenting the change in CHANGELOG + `--help` |
| Tests | Existing suite stays green; help-text doc test | New per-branch exit-contract tests; Skipped test observed RED on `a4855f0b2` |

**Recommended default: Outcome B.** The card's rule 2 is an explicit design requirement; the current code satisfies it only by accident of no script existing yet; and the code's own doc comment draws the deliberate line at absent-inputs, not at refusals. Outcome A remains fully specified and adoptable at M2 — it is the correct choice if the operator reads tolerant-success-for-refusals as intended, but it must then carry the rule-2 waiver explicitly (REQ-CDE-003).

## §F. Acceptance Criteria (structure — full Given-When-Then in acceptance.md)

| AC | Subject | Mechanically checkable via |
|----|---------|---------------------------|
| AC-CDE-001 | §B.1 branch table is complete and address-resolving | each row's `file:line` address greps to the stated marker |
| AC-CDE-002 | Unchanged and Skipped are distinct outcomes (exit-code space under B; documented classes + rule-2 waiver under A) | `go test -run` per branch; RED-now cell recorded |
| AC-CDE-003 | Absent-input branches stay zero (B1/B3/B4) | tests asserting nil |
| AC-CDE-004 | Name-resolution stays non-zero (B2) | existing tests keep passing |
| AC-CDE-005 | Scope decision recorded as verb-local | grep the SPEC for the decision line |
| AC-CDE-006 | Contract documented on the verb surface | `--help` output carries the exit-code table |
| AC-CDE-007 | Adjudication decision recorded before any implementation commit | progress.md §E.1 M2 record precedes run-phase commits |

## Exclusions

### Out of Scope — surfaces this SPEC does not change

- `moai clean --codex-skills` exit-code behavior — its per-entry `Kept:` reporting and fail-open doctrine are deliberate and semantically distinct (§D); changing it here would be the cross-verb inconsistency the card warns about, in reverse.
- The name-resolution exit policy (B2) and the absent-input fail-open doctrine (B1/B3/B4) — both are explicitly documented design; this SPEC preserves them under either adjudication outcome.
- Any new `skills` verb (enable, list, status) — the family currently has one verb; adding verbs is separate work.
- The dry-run default and `--force` gating (B7) — recent, documented, and not part of the inconsistency.

### Out of Scope — plan-phase prohibitions

- Any Go source modification at plan phase — this SPEC's deliverable is the adjudicated contract; implementation belongs to the run phase per M2's outcome.
- Template mirrors (`internal/template/templates/**`) — no template surface invokes this verb (census §B.3), so Template-First does not attach.

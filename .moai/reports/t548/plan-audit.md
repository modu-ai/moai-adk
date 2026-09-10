# SPEC Review Report: SPEC-CODEX-DISABLE-EXIT-001

Iteration: 1/1 (Tier S ceiling per `harness.plan_audit_tier_ceilings` S=1)
Verdict: **PASS**
Overall Score: **0.875** (Tier S PASS threshold 0.75 — met; Tier L mechanical bar 0.85 also met)

Audit window: HEAD `ffc7a5ef9`, branch `WT-codex-disable-exit`, worktree `.claude/worktrees/t548`.
`git diff --stat a4855f0b2..HEAD -- internal/` → empty: the Go tree is byte-identical to the
baseline `a4855f0b2`, so every `file:line` address in the SPEC was verified against the same
bytes its author read. Reasoning context from the SPEC author was not supplied; the lead's
dispatch context was treated as claims and independently re-verified (one of its implicit
claims — a stale `CHANGELOG.md:20` citation candidate — was measured and REJECTED, see below).

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency**: REQ-CDE-001..008 at `spec.md:86-93`, sequential, no gaps, no duplicates, uniform `- **REQ-CDE-NNN** (` pattern-name)` shape.
- **[PASS] MP-2 EARS/GEARS format compliance** (judged against the requirement layer, `spec.md` §C only): all 8 REQs match GEARS patterns — Ubiquitous ×5 (`shall`, :86, :89, :90, :92, :93), Capability gate (`Where`) ×2 (:87, :88), Event-driven (`When`) ×1 (:91). Zero informal entries; no Given-When-Then scenario is presented as a REQ.
- **[PASS] MP-3 YAML frontmatter validity**: all 12 canonical fields present at `spec.md:2-14` with correct types — `id`, quoted `title`, quoted semver `version: "0.1.0"`, enum `status: draft`, ISO `created`/`updated: 2026-09-08`, `author`, `priority: P2`, `phase: "v3.2.0 target"` (carries a release prefix; not a prohibited stage value), `module: internal/cli`, `lifecycle: spec-anchored`, comma-separated quoted `tags`. Optional `era: V3R6` + `related_specs` are permitted additions. Zero snake_case aliases (negative control confirmed the linter rejects `created_at:` — see instrument validation below).
- **[N/A] MP-4 Section 22 language neutrality**: N/A — single-language (Go CLI) SPEC; no multi-language tooling enumerated.
- **[PASS] MP-5 D7 cross-SPEC reconciliation**: three SPEC-IDs referenced (frontmatter `related_specs:15`, body :34; plan.md :16, :75). All three exist in `.moai/specs/`: SPEC-CODEX-SKILL-DISABLE-001 `status: completed`, SPEC-CODEX-SKILL-PATH-SLASH-001 `status: implemented`, SPEC-CODEX-GHOST-SKILLS-PRUNE-001 `status: completed`. None ∈ {retired, superseded, archived} → no reconciliation obligation. No referenced SPEC is missing. The t540 "landed prerequisite" claim was additionally verified against the code itself (`toConfigPath` conversion precedes the guard at `codex_skills_disable.go:247`), which is stronger evidence than the status field.
- **[PASS] MP-6 D8 cross-platform discipline**: `grep -c 'syscall' spec.md` → 0 → auto-PASS (D8-4).
- **[PASS] MP-7 clarification gate**: `grep -c '\[NEEDS CLARIFICATION' plan.md` → 0. `research.md` absent (Tier S artifact set) → that leg N/A per the MP-4 precedent.

**Must-pass firewall: 6 PASS + 1 N/A, zero failures.**

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 1.00 | 1.0 band | Every REQ has a single reading; the M2 decision gate prescribes BOTH outcomes with named classes (§E table, spec.md:107-115); the three outcome classes (performed/refused/absent-input) are defined once (:86) and used consistently |
| Completeness | 1.00 | 1.0 band | HISTORY (:20-22), §A user story, §B measured context, §C 8 REQs, §F AC structure + full GWT in acceptance.md, two `### Out of Scope —` H3 sub-headings with specific `-` bullets (spec.md:131, :138); frontmatter complete; census commands + verbatim outputs carried in progress.md §E.1 |
| Testability | 0.75 | 0.75 band | 6 of 7 ACs carry concrete commands with named markers; AC-CDE-002's RED-now cell carries 2 of the 4 release-blocking elements (command named, tree SHA pinned; verbatim stdout + exit code deferred to run-phase E8 — structurally unobservable at plan time because the test does not exist yet, confirmed: `grep -c TestRunCodexSkillDisableSkippedExitsNonZero` → 0) — and plan.md §F's milestone order contradicts the RED-before-GREEN obligation (D2) |
| Traceability | 0.75 | 0.75 band | All 8 REQs are covered by ≥1 AC (verified against `internal/spec/ears.go:123-128`'s `maps`-keyword extractor — the linter's CoverageIncomplete direction is REQ→AC and reports clean); AC-CDE-007 carries no `maps` clause and is orphaned (D1) |

## Defects Found

D1. **AC-CDE-007 is an orphan AC — no `maps` clause** — `acceptance.md:56` — the heading `### AC-CDE-007 — Adjudication recorded before implementation (both outcomes)` is the only AC heading without a `— maps REQ-...` clause; the other 6 all carry one. The linter cannot see it either: `ExtractRequirementMappings` (`internal/spec/ears.go:123-128`) keys on the literal keyword `maps`, and its CoverageIncomplete rule fires only in the REQ→AC direction, so this gap is invisible to every mechanical check — precisely why the audit caught it and the lint did not. No REQ in §C backs "the decision is recorded before implementation" either; the AC's subject has no requirement-layer anchor at all. — Severity: minor — Class: **blocking** — Required fix: append `— maps REQ-CDE-002, REQ-CDE-003` to the AC-CDE-007 heading (both REQs are conditional on the M2 adjudication resolving, which is exactly what the record establishes). Routing note: `acceptance.md` body is manager-spec's surface (`spec-frontmatter-schema.md` § Forbidden ownership crossings — manager-develop may not touch it), so the orchestrator routes this one-line fix through manager-spec (plan-phase amendment) or accepts it as documented debt; the verdict does not depend on it.

D2. **plan.md §F milestone order contradicts the RED-before-GREEN obligation** — `plan.md:60-61` vs `acceptance.md:25` + `plan.md:52` (E8) — the milestone table lists M3 (implementation: flips the Skipped branch) BEFORE M4 (exit-contract tests), while AC-CDE-002's RED-now cell and E8 both require the Skipped test's RED output to be "observed RED on the pre-change tree" / "captured on the pre-change tree before GREEN", pinned to `a4855f0b2`. Executed in table order, M3 lands first and the tree stops being `a4855f0b2` — the pinned RED becomes unobservable without a checkout of the old tree, and the two-cell adoption discipline (`verification-completeness.md` §2.1) collapses to a green-path-only adoption: the classic vacuous-green hazard, since `go test -run <name>` prints `ok` with exit 0 identically whether the selector matched one passing test or zero tests (`[no tests to run]`). The RED-first capture is the only guard against that vacuity, which is what makes the ordering contradiction load-bearing rather than cosmetic. — Severity: major — Class: **blocking** — Required fix: no artifact change needed — bind the lane's execution order: write the Skipped test FIRST, run it on the untouched tree, capture E8's four elements (command + verbatim stdout + exit code + tree SHA `a4855f0b2`), THEN land M3's one-branch change, then re-run to GREEN. E8's evidence must carry all four elements per `verification-completeness.md` §2.1.

D3. **`tier:` frontmatter omitted** — `spec.md:2-14` — deliberate per the dispatch, but mechanical consumers read absent `tier:` as Tier L (backward-compat default): the run-gate's skip-eligibility predicate (`internal/runtime.SkipEligibleByScore`) applies the 0.85 bar, and the Tier L artifact set (5 files) is the declared expectation. Materiality today: none — this audit's 0.875 clears both bars, and the dispatch pins Tier S. — Severity: minor — Class: **optional** — Required fix: none required; if the lane wants mechanical consumers to read the true tier, add `tier: S` via the manager-spec frontmatter surface.

D4. **AC-CDE-002 Outcome-B branch is not in literal Given-When-Then form** — `acceptance.md:24-26` — it uses the two-cell RED-now/green-path structure instead of the GWT scaffolding the other ACs carry. Substance is present (binary-testable, both cells, right-reason red stated: "the Skipped branch returns nil at :422-424, read this tree"), and the two-cell form is what `verification-completeness.md` §2 mandates for observed-failure criteria — a formatting variance, not a defect. — Severity: minor — Class: **optional** — Required fix: none; if normalized later, wrap the two cells in GWT at the next manager-spec touch.

Rejected during measurement (recorded so the next reader does not re-litigate):
- **Candidate: stale `CHANGELOG.md:20` citation** in spec.md §B.2 — REJECTED by re-measurement: `grep -n 'three name-resolution failures' CHANGELOG.md` → line 20, exact. My first sed window misattributed the wrapped output; the citation stands. (Also `skills.go:8, :53-55, :90, :96` and all ten §B.1 branch addresses B1-B10 plus all five skip reasons :237-239/:261-263/:282-284/:293-295/:302-304 verified against this tree.)
- **Candidate: census claims fabricated** — REJECTED: `.moai/reports/t502/e2e-verb.sh:19,41,43,68` verified verbatim (four `skills disable` invocations); `grep -rc 'skills disable' internal/template/templates/` → 0.

## Audit Focus Answers (dispatch items)

1. **AC→REQ mapping**: 6 of 7 ACs carry `maps`; AC-CDE-007 does not (D1). Linter behavior confirmed at `internal/spec/ears.go:128` — the `(?i)maps\s+(REQ-...)` section pattern means anything without the keyword is invisible to traceability. All 8 REQs are covered (linter clean + manual matrix).
2. **AC-CDE-002 RED-now cell**: a genuine two-cell structure is authored (RED-now + green path), but at plan time the RED cell is a *promise* — command named, tree pinned, right-reason red stated, yet no verbatim stdout/exit code exists because the test does not exist yet (grep 0). The SPEC handles this honestly via E8, but the plan.md §F order threatens to make the pinned RED unobservable (D2). Not release-blocking-eligible as authored; it becomes so when E8's four elements are captured.
3. **§B.1 branch addresses**: B2 (`:400-402`, `Refusing to write` + `fmt.Errorf("cannot disable %q for Codex")`), B5 (`:419-421` ← upsert `:296-299`, `Unchanged: ... (it is already disabled)`), B6 (`:422-424` ← upsert skip closure `:232-236`, `Skipped: ... (<reason>)`) — all three resolve exactly, as do the other seven rows.
4. **Operator decision record**: `progress.md:47-53` carries the M2 block — Outcome B picked, all four branches' codes stated, channel named (lead-relayed AskUserQuestion, 2026-09-08), the four operator-cited grounds enumerated, rule-2 enforcement tied to AC-CDE-002, Outcome A preserved as the rejected alternative. Commit order verified: `5e7907746` (4 artifacts) → `ffc7a5ef9` (progress.md +8 only) → no implementation commit exists (`internal/` diff empty). AC-CDE-007's ordering precondition holds today. Scope of my verification: the record's existence, datedness, and precedence — the operator's judgment itself is not re-auditable from here and is not claimed.
5. **Scope discipline**: clean. §D resolves verb-local with the single-target-vs-sweep rationale; `moai clean --codex-skills` appears only as the documented contrast + Out of Scope exclusion (spec.md:133); plan.md §D PRESERVE names `codex_skills_prune.go` (byte-unchanged), `doctor_codex.go`, the t502 evidence script; AC-CDE-005 makes the preserve mechanically checkable (`git diff --stat a4855f0b2 -- internal/cli/codex_skills_prune.go` → empty).

## Instrument Validation (lint)

Command run: `moai spec lint .moai/specs/SPEC-CODEX-DISABLE-EXIT-001/spec.md` — verbatim output:

```
✓ No findings — all SPEC documents are valid
EXIT=0
```

Empty-sweep guard (the clean run could otherwise be a zero-file sweep): negative control run against a mutated copy with `created_at:` injected returned `FrontmatterInvalid` ERROR + `CoverageIncomplete` warnings, exit 1 — the tool reads the given path and the clean verdict is a real sweep of this SPEC.

## Regression Check

Iteration 1 — no prior-iteration defects.

## Recommendation

Verdict is PASS (0.875 ≥ 0.75 Tier S bar; no must-pass failure). Two blocking findings route forward, neither blocks run-phase entry:

1. **D2 → binds the run-phase lane directly**: execute M4's RED capture (test written, run on the untouched tree, E8 four-element evidence) BEFORE M3's one-branch change. This is execution ordering, not an artifact edit.
2. **D1 → orchestrator routing decision**: one-line `maps` clause on AC-CDE-007 via manager-spec (acceptance.md body is outside manager-develop's ownership), or accept as documented debt. The plan-artifact hash changes if fixed, which re-opens skip-eligibility — a legitimate trade for closing the only traceability gap.

D3/D4 are optional; no action required for this SPEC to proceed.

## Re-stamp

Re-verified: 2026-09-08, at commit `d1742de07` (`docs(SPEC-CODEX-DISABLE-EXIT-001): D1 — AC-CDE-007 heading gains maps clause (audit finding)`), branch `WT-codex-disable-exit`. Delta scope per the re-audit contract: D1 only — no from-scratch re-audit.

- **Delta verification**: exactly one commit lands between the audit window (`ffc7a5ef9`) and `d1742de07`. `git diff ffc7a5ef9..d1742de07` → a single changed line in `acceptance.md` only: the AC-CDE-007 heading gains `— maps REQ-CDE-002`; body untouched. (The dispatch's stated range `5e7907746..d1742de07` spans two commits and therefore also shows the +8 progress.md M2 record from `ffc7a5ef9`, which was already inside the audited window.) No other artifact, no Go source, nothing else moved.
- **D1 disposition: RESOLVED.** `grep -c "maps REQ" acceptance.md` verbatim output:

```
7
```

All seven AC headings now carry a `maps` clause (AC-CDE-007 → REQ-CDE-002 at `acceptance.md:56` — an accurate anchor: REQ-CDE-002 is the requirement whose activation condition is the M2 adjudication resolving, and the adopted outcome is B). The linter-invisible-orphan defect is closed at both the human and mechanical traceability layers.
- **Verdict: unchanged — PASS.** D1's resolution removes the only downgrading factor: Traceability 0.75 → 1.00, aggregate 0.875 → **0.938**. No new finding; D2 (run-phase RED-before-GREEN execution ordering) and D3/D4 (optional) stand as previously routed. Note for the run-gate: this acceptance.md edit changes the plan-artifact hash, so the artifact-hash leg of skip-eligibility keys on the post-`d1742de07` state against THIS verdict.

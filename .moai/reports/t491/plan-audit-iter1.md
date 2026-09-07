# SPEC Review Report: SPEC-CC-GD124-001

Iteration: 1/2 (Tier M ceiling per `harness.yaml` `plan_audit_tier_ceilings`)
Verdict: **FAIL**
Overall Score: 0.81 (mean of category scores; Tier M PASS threshold 0.80 — score clears the threshold on arithmetic, but two blocking verification-layer defects stand between this SPEC and acceptance per `verification-completeness.md` §2 and Group 4 AC-5; outstanding blocking findings are fixed before the verdict is revisited)

Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t491`, branch `WT-cc-upstream-sweep`, audited at HEAD `c951d21cb` (base `615d18c1f`). Reasoning context from the author was not supplied; audit performed against the artifacts only (M1 Context Isolation respected).

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: REQ-001..REQ-011 at `spec.md:L35-L45` — sequential, zero-padded, no gaps, no duplicates. 11 REQs.
- [PASS] MP-2 EARS/GEARS format compliance (requirement layer): all 11 REQs match GEARS patterns — Ubiquitous ("shall") REQ-001/002/005/006/007/008/009/010/011, Event-driven ("**When** ... shall") REQ-003/004/010. No informal language, no Given-When-Then presented as a REQ. Layer judged: `REQ-XXX` entries in `spec.md` §1 only; `AC-XXX` entries in `acceptance.md` are the verification layer (Given-When-Then) and were graded under Group 4, not here.
- [PASS] MP-3 YAML frontmatter validity: all 12 canonical fields present at `spec.md:L2-L14` (`id`, `title`, `version: "0.1.0"` quoted semver, `status: draft`, `created`/`updated` ISO dates, `author`, `priority: P1`, `phase: "v3.2.0 target"` — a release target, not a prohibited lifecycle stage, `module` path-like, `lifecycle: spec-anchored`, `tags` CSV string) plus optional `tier: M`. No snake_case aliases (`created_at`/`updated_at`/`labels`/`spec_id` absent). Verified against `.claude/rules/moai/development/spec-frontmatter-schema.md` (SSOT).
- [N/A] MP-4 Section 22 language neutrality: single-language documentation SPEC (two English `.md` rule files); no multi-language tooling matrix involved. N/A auto-passes.
- [PASS] MP-5 D7 cross-SPEC reconciliation: SPEC-ID extraction over `.moai/specs/SPEC-CC-GD124-001/` yields exactly two refs — self and `SPEC-CODEX-SESSION-MSG-001` (provenance for the Origin-line hunk, `spec.md:L44`). The referenced SPEC exists at `.moai/specs/SPEC-CODEX-SESSION-MSG-001/spec.md` with `status: completed` — not in {retired, superseded, archived}; no reconciliation obligation triggered. No BLOCKING finding.
- [PASS] MP-6 D8 cross-platform discipline: `grep -c syscall` = 0 in spec.md, plan.md, acceptance.md (measured this run). Auto-PASS per D8-4.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' plan.md` → no matches; `research.md` absent (correct for Tier M — 3-artifact set). N/A-pass on the absent file per the MP-4 precedent.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 1.0 | 1.0 | Every REQ single-interpretation: REQ-001 exact row values (`spec.md:L35`), REQ-003 error-direction rationale (`L37`), REQ-010 exception hunk precisely bounded (`L44`), REQ-011 line-level protected lines (`L45`). Plan carries verbatim replacement text for every edit (`plan.md:L54-L57`, `L69-L71`). No pronoun ambiguity found. |
| Completeness | 0.75 | 0.75 | Frontmatter complete; HISTORY (`L19-L22`), Requirements, AC index, Out of Scope with three `### Out of Scope — <topic>` H3 sub-headings + specific bullets (`spec.md:L73-L90`) all present. One gap: no `WHY`/`WHAT` headings — the context content exists in substance under HISTORY (`L24-L31`) but carries no dedicated section heading (SC-2/SC-3). Optional fix: 2-line heading addition. |
| Testability | 0.75 | 0.75 | All 9 ACs binary (grep counts, diff hunk count, exit codes); no weasel words. Docked for D2: AC1/AC3/AC4 are binary but under-discriminating — writable mutants satisfy the AC while violating the REQ (see Defects). AC7 documents its own mutant explicitly (`acceptance.md:L47`). |
| Traceability | 0.75 | 0.75 | Exactly one REQ uncovered: REQ-009 (`spec.md:L43`, five-bullet preservation) has no corresponding AC anywhere in `acceptance.md` (AC-5 violation). Partial: REQ-011's csm L9 "same platforms" element has no guard grep (L23/L25/cw-L34 elements are covered by AC8 strings). Every AC traces to an existing REQ. |

Overall: (1.0 + 0.75 + 0.75 + 0.75) / 4 = 0.8125 ≈ 0.81.

## Evidence re-measured in this run (all reproduce the SPEC's pinned baselines)

- cw RED-now: `Fable (256K)`=1 (L24), `200K/256K`=2 (L64, L80), `Sonnet 5 (1M)`=0 — matches `spec.md:L31` and acceptance.md AC1-AC3 cells.
- csm RED-now: native-Windows denial=1 (L21), `unavailable on Amazon Bedrock`=1 (L22), `turning messaging off silently`=1 (L24), `tengu_harbor_kite`=1 (L25) — matches AC4-AC6, AC8.
- cw pair byte-identical: `diff -q` exit 0 (measured). csm pair: `diff -u CST CS` → exactly 1 hunk (measured); REQ-005/REQ-010/AC7 pre-state verified.
- csm § Availability intro says "Five constraints" (L19) with exactly 5 bullets (L21-L25, measured by section-scoped count) — REQ-009's quoted string is accurate; plan §D constraint accurate.
- `9ef2b91e1` provenance (V9) consistent with the observed single-hunk diff.
- Tier/budget: 11 REQ ≤ 16 (M), 9 AC ≤ 16 (M), artifact set spec+plan+acceptance = 3 = Tier M. Frontmatter `tier: M` matches.
- Exit-code convention: acceptance.md:L12 documents the grep -c count-not-exit-code judging rule; since `grep -c`'s exit code is a pure function of the printed count, the recorded count constitutes the exit-code element in substance. `verification-completeness.md` §2.1 four-element test judged satisfied on that basis (judgment call stated for auditability). Single-invocation form verified on every pinned command (no pipes/`&&`/`;`).

## Defects Found

D1. TRACE-REQ009-NOC — acceptance.md (whole file) / spec.md:L43 — REQ-009 (preserve csm five-bullet structure + "Five constraints" intro accuracy) has no corresponding acceptance criterion; the AC battery — which acceptance.md:L60 declares "the full gate" — cannot detect a bullet merge/split or intro-line edit, and AC8's untouched-string greps survive such a mutation. Severity: major — Class: blocking — Required fix: extend AC8 (or add AC10) with single-invocation greps on both csm copies: `grep -c 'Five constraints' <CS>` = 1, plus presence=1 greps for each of the five bullet anchors (e.g. `- **Operating system**`, `- **The shared flag slot**`) so the five-bullet count is bounded without arithmetic.

D2. TEST-AC-VALUE-PIN — acceptance.md:L17 (AC1), L27 (AC3), L32 (AC4) — Green-path greps pin presence/absence only, not the replacement content, so the §2 mutant probe (`verification-completeness.md §2`, [HARD], paths-scoped to SPEC artifacts) fails: (a) AC1 passes a mutant writing `| Fable (1M) | 256,000 tokens | **90%** | ...` — the exact value error GD-1 exists to repair; (b) AC3 passes a mutant inserting the Sonnet 4.x row with 1M/50% values — the exact error direction REQ-003 exists to prevent; (c) AC4 passes a mutant dropping the named-pipe parenthetical and the cross-machine gap statement from REQ-006. Severity: major — Class: blocking — Required fix: pin the full verbatim replacement lines from plan M1/M2 as greps (e.g. `grep -c '| Fable (1M) | 1,000,000 tokens | \*\*50%\*\* | ~500,000 tokens |' <CW>` = 1, likewise the Sonnet 5 and Sonnet 4.x rows; for AC4 add `grep -c 'named pipe' <CS>` ≥ 1 and a gap-statement grep such as `grep -c 'not documented' <CS>` ≥ 1). Each remains a single invocation.

D3. CONS-PROGRESS-TIER — progress.md:L7 — §E.1 states "Tier S artifacts authored (spec.md, plan.md, acceptance.md, progress.md)" — wrong on both counts: the SPEC is `tier: M` (`spec.md:L14`) and the Tier M artifact set is 3 files (progress.md is the progress record, not a tier artifact). Stale residue of the pre-correction draft; the plan→run orchestrator reads §E.1 as the audit-ready signal. Severity: minor — Class: blocking (internal consistency) — Required fix: rewrite the note to "Tier M artifacts authored (spec.md, plan.md, acceptance.md)".

D4. CONS-VERSION-HISTORY — spec.md:L4 vs L22 — HISTORY records "v0.1.1 — frontmatter tier correction (S→M)" but frontmatter `version` remains `"0.1.0"`; the recorded revision does not exist in frontmatter. Severity: minor — Class: optional — Required fix: bump `version: "0.1.1"` or relabel the HISTORY entry.

D5. TEST-AC7-BLANKLINE — acceptance.md:L46-L47 — AC7 says the single csm hunk's "only changed line is the Origin-line blockquote"; the actual `diff -u CST CS` hunk adds TWO lines (the blockquote plus its separating blank line — measured this run). An executor testing "only difference is the blockquote" literally reads the 2-line diff as a mismatch → wrong-reason red on a control criterion. Severity: minor — Class: optional — Required fix: reword to "the hunk's only change is the Origin-line blockquote block (the blockquote plus its separating blank line)".

D6. TEST-AC8-EXACT-SET — acceptance.md:L51 — AC8 expects status entries "exactly the 4 target files plus `.moai/specs/...` plus `.moai/reports/t491/`", but the evidence was committed at `c951d21cb` and run phase will commit the SPEC dir before editing, so the observed set will legitimately be a subset (the dirs no longer untracked) → strict "exactly" reading fails for reasons the work cannot fix. Severity: minor — Class: optional — Required fix: reword to "no entries outside the 4 target files, the SPEC directory, and `.moai/reports/t491/`".

D7. EVID-NAMEDPIPE-QUOTE — `.moai/reports/t491/upstream-verification-20260907.md` (Claim #3 / V6) — the record's Claim #3 asserts "(named-pipe socket)" and plan M2 edit 1 writes "the inbox socket is a named pipe" into an ALWAYS-LOADED rule, but no evidence line in the record quotes the named-pipe mechanism (`grep -rn 'pipe' .moai/reports/t491/` = 0 hits, measured) and the record's Gaps section does not disclose it. Per the audit mandate the record's claims are the verified baseline, so this is not a SPEC defect; it is an evidence-quote gap in the record. Severity: minor — Class: optional — Required fix: add the docs sentence carrying the named-pipe mechanism to V6 (or to the Gaps list if it was not actually quoted upstream).

D8. WORD-FLAG-DEPENDENCY — plan.md:L71 (M2 edit 3) — the retained phrase "disables the feature-flag evaluation the channel depends on" is mildly self-stale after the fix: for v2.1.248+ sessions the channel no longer depends on that evaluation (the very next sentence says messaging works with flag fetching off). Behavior is not misrepresented (the next sentence corrects it), but the dependency framing survives from the pre-fix text. Severity: minor — Class: optional — Required fix (if desired): "disables the feature-flag evaluation" (drop "the channel depends on"), or "...the evaluation the channel depended on before v2.1.248".

D9. STRUCT-NO-WHY-HEADING — spec.md:L24-L31 — context/rationale content present but without a `WHY`/Context heading (completeness 0.75 driver). Severity: minor — Class: optional — Required fix: add `## Context (WHY)` above L24.

## Regression Check (Iteration 2+ only)

N/A — iteration 1.

## Recommendation

Verdict is FAIL on blocking findings D1-D3. The delta re-audit (iteration 2) is scoped to D1-D3 resolution plus a regression check over D4-D9 dispositions; the full audit does not re-run. Fixes, in order:

1. (D2) In `acceptance.md`, extend AC1/AC3 green cells with full-row greps carrying the exact values from plan M1 edits 1-3, and extend AC4 with named-pipe + gap-statement greps. Keep every command single-invocation; keep the count-not-exit-code convention.
2. (D1) Add five-bullet-structure coverage to AC8 (or a new AC10): `Five constraints` = 1 plus one presence grep per bullet anchor, on both csm copies.
3. (D3) Fix progress.md §E.1 to "Tier M artifacts authored (spec.md, plan.md, acceptance.md)".
4. Optional, ride-along: D4 (version bump), D5 (AC7 blank-line wording), D6 (AC8 exact-set wording), D8 (flags-bullet dependency phrasing), D9 (WHY heading), D7 (record evidence quote).
5. No change requested to the factual content: every GD-1/2/3/4 claim in plan M1/M2 was cross-checked against the verification record and found consistent (Fable 1M v2.1.257, Sonnet 5 1M v2.1.247, native-Windows same-machine v2.1.234, provider/flag class-level facts v2.1.248, four provider names retained in the cross-machine axis). The AC7 parity criterion is correctly stated as a MUST-NOT-flip control with the template-neutrality rationale (plan.md:L13, L99), and no milestone edit leaks into the protected lines (cw L30-34; csm L9/L23/L25) or the GD-5..9 exclusions.

# SPEC Review Report: SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001

Card: t551 · Tree: `.claude/worktrees/t551` · Branch: `WT-audit-fail-open` · HEAD: `3ac58b5a1`
Iteration: 1/2 (Tier M ceiling)
Auditor: plan-auditor (Claude anchor; `audit_model` single-backend — no `audit_multi` fan-out)

**Verdict: FAIL**
**Overall Score: 0.81** (Tier M PASS threshold 0.80 — the score is ABOVE threshold; the FAIL is
issued on the enumerated blocking defect delta below, not on the aggregate. See § Verdict rationale.)

Reasoning context ignored per M1 Context Isolation. The dispatch's six focus points were treated as
audit direction; no author reasoning was consulted. Every claim below cites a command and its
output.

---

## ⚠ Baseline warning — the artifacts changed during this audit

The three artifacts were **rewritten on disk mid-audit** by a second writer. Two frontmatter defects
that were live when this audit opened were repaired between my first and second measurement, and
`progress.md` was extended to record the repair.

| Measurement | State at audit open (~04:36) | State at audit close (04:46) |
|---|---|---|
| `moai spec lint .../spec.md` | `ERROR ParseFailure … line 13: cannot unmarshal !!seq into string` | `✓ No findings`, exit 0 |
| `status:` in sibling artifacts | present in `plan.md:5`, `acceptance.md:5` | absent from both |
| `mcp__moai__spec_audit` era | `V3R5`, H-3, grandfathered | `V3R6`, H-5, `modern_era_clean: 1` |

Per `agent-common-protocol.md` § Background Agent Execution [HARD] — *"a worktree being actively
audited has exactly one writer"* — this is reported as a **process defect** (D1), not silently
absorbed. The content findings D2-D10 were re-read against the post-repair files and are attributed
to the CURRENT state; the frontmatter findings are recorded as **resolved-during-audit**, not as
live defects.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `/usr/bin/grep -ho 'REQ-CBR-[0-9]*' spec.md | sort -u`
  → `REQ-CBR-001` … `REQ-CBR-009`, 9 unique. `/usr/bin/grep -c '^\*\*REQ-CBR-' spec.md` → `9`.
  Sequential, no gaps, no duplicates, consistent 3-digit padding.

- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in
  `spec.md`), per M3 § Scope. All 9 match a GEARS pattern: REQ-CBR-001/002/003/004 event-driven
  (`When …, the <subject> shall …`, spec.md:103/108/115/121); REQ-CBR-005 ubiquitous
  (`The blank-output inconclusive summary shall name …`, spec.md:125); REQ-CBR-006 state-driven
  (`While the codex backend is unavailable …`, spec.md:131); REQ-CBR-007/008 unwanted
  (`The audit shall not …`, spec.md:135/140); REQ-CBR-009 compound Where+When
  (`Where workflow.audit.gates.codex is required, when …`, spec.md:144). Corroborated mechanically:
  `moai spec lint .../spec.md` emits no EARS/GEARS modality finding. The `Given/When/Then` entries
  in `acceptance.md` are `AC-XXX` verification-layer artifacts and were **not** graded here (M3
  § Scope); they are graded under Group 4.

- **[PASS] MP-3 YAML frontmatter validity** *(resolved during audit — see D1)* —
  `moai spec lint .moai/specs/SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001/spec.md` → `✓ No findings —
  all SPEC documents are valid`, `EXIT=0`. All 12 canonical fields present with correct types
  (`spec.md:2-13`), `tags` now a quoted comma-separated string, `status:` absent from `plan.md` and
  `acceptance.md` per § Artifact Statelessness. **At audit open this was a FAIL**: `tags: [codex,
  audit, fail-open, gate, issue-1632]` decoded as a YAML sequence against `internal/spec/lint.go:511`
  `Tags string`, producing `ERROR ParseFailure … cannot unmarshal !!seq into string` on all THREE
  artifacts — which made the entire SPEC invisible to every downstream lint rule, masking two
  `ArtifactStatusFieldForbidden` errors underneath it.

- **[N/A] MP-4 Section 22 language neutrality** — single-language SPEC. `module: internal/cli`
  (spec.md:11); every implicated site is Go. No multi-language tooling claim is made, so the
  16-language enumeration criterion does not apply.

- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `/usr/bin/grep -Eoh 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'
  spec.md plan.md acceptance.md progress.md | sort -u` → `SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001`
  only. No external SPEC is referenced, so no retired/superseded reconciliation is owed. No BLOCKING
  finding.

- **[PASS] MP-6 D8 cross-platform discipline** — `/usr/bin/grep -c 'syscall' spec.md plan.md
  acceptance.md progress.md` → `spec.md:0 plan.md:0 acceptance.md:0 progress.md:0`. D8-4 auto-PASS.

- **[PASS] MP-7 clarification gate** — `/usr/bin/grep -rn '\[NEEDS CLARIFICATION' plan.md` → no
  match, exit 1. `ls research.md` → `No such file or directory` (Tier M does not require it). No
  unresolved marker.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 — minor ambiguity a reasonable engineer resolves consistently | REQ-CBR-001..009 each carry a single interpretation (spec.md:103-146). Downgraded from 1.0 by two false/over-stated prose claims outside the requirement layer: acceptance.md:73-75 (D2) and spec.md:91-95 + 168-171 (D4). |
| Completeness | 1.00 | 1.0 — all sections + all 12 frontmatter fields | HISTORY spec.md:20-24; WHY/Context §A spec.md:28; WHAT/Requirements §B spec.md:99; §C decision spec.md:150; Out of Scope — three `### Out of Scope — <topic>` H3s at spec.md:184/196/204, each with `-` bullets. Frontmatter clean (`moai spec lint` exit 0). |
| Testability | 0.75 | 0.75 — one AC not precisely binary-testable | Eight of nine ACs are binary with named fixtures (e.g. AC-CBR-001 fixtures `" "`, `"\n"`, `"\n\t  \n"`, `""` at acceptance.md:53). No weasel words present (`appropriate`/`adequate`/`reasonable`/`proper` absent). AC-CBR-009 (acceptance.md:128-136) carries a zero-match grep assertion with no positive control and no mechanically-defined sweep scope — D5. |
| Traceability | 0.75 | 0.75 — one AC's mapping is indirect/absent | All 9 REQs covered: 001/002/003→AC-003, 004→AC-001, 005→AC-005, 006→AC-004, 007→AC-002+AC-008, 008→AC-006, 009→AC-007. Confirmed mechanically: `moai spec lint` emits no `CoverageIncomplete`. AC-CBR-009 references no REQ — "Covers the plan §B.2 seam" (acceptance.md:135) — D6. |

Aggregate = (0.75 + 1.00 + 0.75 + 0.75) / 4 = **0.8125 → 0.81**

---

## Defects Found (structured defect-list)

**D1** — `plan-audit process` — `.moai/specs/SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001/{spec,plan,acceptance,progress}.md` — The four artifacts were rewritten by a second writer **during** the audit window (mtimes 13:44:27–13:44:40 vs audit open 13:36). Two live must-pass defects were repaired mid-measurement, so this verdict's frontmatter row is attributed to a state that did not exist at dispatch. `agent-common-protocol.md` § Background Agent Execution [HARD]: an actively-audited worktree has exactly one writer; a foreign write is a process defect to report, never to continue quietly through. — Severity: **major** — Class: **blocking** — Required fix: no artifact change. The lead records the overlap; future plan-audits of this card are dispatched with the authoring session confirmed idle first, and the author does not edit artifacts after handing them to audit. (Repair content itself was correct and is not being reverted.)

**D2** — `acceptance.md:73-75` — The stated discrimination claim is **false for two of the three sites it names**. acceptance.md:73-75 asserts: *"This criterion fails if the guard (`internal/cli/mcp_codex.go:817`) is repaired without the collection and selection sites (`:1134`, `:1137`, `:1253`)."* Traced against the code in this tree, AC-CBR-003 **passes** when `:817` and `:1253` are repaired and `:1134`/`:1137` are left as exact-equality. Trace, each line read verbatim: `mcp_codex.go:1134` `if p.Item.Type == "exitedReviewMode" && p.Item.Review != ""` sets `reviewText = "   "`; `:1137` sets `agentText` to the real prose; `bestCodexReviewText` (`:1252-1257`, `if review != "" { return review }; return agent`) with `:1253` made blank-aware falls through and returns the **agent** text; the repaired guard at `:817` passes it; `synthesizeReviewOutput` returns `pass` — which is exactly what AC-CBR-003 requires (acceptance.md:70-71). The load-bearing site for this criterion is `:1253` **alone**. — Severity: **major** — Class: **blocking** — Required fix: rewrite acceptance.md:73-75 to name only `:1253` (plus the guard) as the sites AC-CBR-003 discriminates, and correspondingly soften spec.md:80-87 (`"The three sites are one repair"`), whose supporting argument likewise describes only `bestCodexReviewText`.

**D3** — `acceptance.md` (no AC exists) — **No acceptance criterion observes the only behavior `:1134`/`:1137` actually change.** Those two sites differ from a repaired tree solely in the *overwrite* case: the collection loop (`mcp_codex.go:1120-1145`) reassigns `reviewText`/`agentText` on **every** `item/completed` notification, so a later whitespace-only `exitedReviewMode` or `agentMessage` clobbers an earlier real one under `!= ""`. Given D2, the SPEC currently has zero coverage for the two collection sites while REQ-CBR-001 and REQ-CBR-002 (spec.md:103-111) mandate them. — Severity: **major** — Class: **blocking** — Required fix: add an AC of the form *"Given two `agentMessage` items, the first carrying real review prose and the second whitespace-only, When the collector processes both, Then the real prose is the collected agent text"* (and the mirror case for `exitedReviewMode`). Cover REQ-CBR-001 and REQ-CBR-002 there rather than at AC-CBR-003.

**D4** — `spec.md:91-95` and `spec.md:168-171` — **The headline benefit claim overstates the consumer-visible consequence, and contradicts the SPEC's own §E.** spec.md:94-95 states *"A machine consumer therefore SEES the gap instead of consuming a fabricated pass."* Traced through `mcp_convergence.go:185-200` read verbatim: a codex verdict of `inconclusive` under a `required` gate leaves `requiredFails` empty (`:181`), so `case allPass(required)` (`:190`) is false and control reaches `default:` (`:194`) → `overall = claudeVerdictOrDefault(verdicts, overallVerdictPass)` (`:199`), and `claudeVerdictOrDefault` (`:369-376`) returns **the claude anchor's verdict**. In the ordinary case (claude `pass`), `overall_verdict` is therefore **`pass` both before and after this repair**. What genuinely changes is narrower and worth stating precisely: the per-backend codex verdict, its Summary, and the `gate_unmet` field (`mcp_codex.go:291`, populated at `:1615`). Separately, spec.md:218-221 (§E Gaps) concedes this same convergence claim is *"a reading … not a test result"* — so §A/§C assert as established what §E records as unmeasured (VCI §1.1 surface 4). — Severity: **major** — Class: **blocking** — Required fix: amend spec.md:91-95 and 168-171 to state explicitly that `overall_verdict` remains `pass` via the fail-open `default` arm, and that the repair's benefit is the per-backend verdict + Summary + `GateUnmet` becoming truthful — i.e. *fail-open, honestly labelled* (which spec.md:173-175 already says correctly; §A and §C.3 must not claim more than that). Answering the dispatch's point 4 plainly: `inconclusive` does **not** close the gate hole — it makes the hole visible in fields a consumer must opt into reading. That is a real improvement and it is correctly scoped to this card (axis 3 enforcement is a sibling per spec.md:198-200), but it must not be written as though the gate now blocks.

**D5** — `acceptance.md:128-136` — **AC-CBR-009's absence half is vacuous-capable.** Its `Then` requires *"no exact-equality emptiness check remains on that path"*, checkable *"by a grep assertion in the suite"* (acceptance.md:135-136). A zero-match grep is satisfied for free by a wrong pattern, a wrong file, or a scope that selected nothing — and **"the review-text path" is nowhere mechanically defined** in the artifacts (no line ranges, no function list). `acceptance.md:159-161` deliberately lists only AC-CBR-001/003/005 as requiring RED on the unrepaired tree, so AC-CBR-009 has no RED cell at all. Per `verification-completeness.md` §1.1 (*"a pass whose swept set is empty asserts nothing"*) and §2.1, this criterion is unadopted. — Severity: **major** — Class: **blocking** — Required fix: (a) name the sweep scope mechanically (the enumerated four sites, or a named function set); (b) add a positive control — assert the helper has exactly 4 call sites, and assert the same grep returns a **non-zero** count on the pre-repair tree; (c) add AC-CBR-009 to the acceptance.md §D RED list with its four RED-cell elements (command, verbatim stdout, exit code, tree SHA).

**D6** — `acceptance.md:135` — **AC-CBR-009 is an orphan acceptance criterion.** It states *"Covers the plan §B.2 seam"* rather than a `REQ-CBR-XXX`. Every other AC names its REQ. Group 4 AC-4 requires each AC to reference a valid REQ that exists in `spec.md`. — Severity: **minor** — Class: **blocking** — Required fix: either add a requirement to spec.md §B mandating a single blankness discriminator across the review-text path and point AC-CBR-009 at it, or reclassify AC-CBR-009 as a §D quality gate (where it fits better — it is a code-shape check, not a behavior criterion) and drop it from the AC-CBR-001..009 numbering.

**D7** — `acceptance.md:132` vs `plan.md:84-87` — **An AC hard-mandates what the plan calls a preference.** AC-CBR-009 requires *"all four sites route through the same blankness helper"*; plan.md:85 says a helper *"is **preferred** over four inline `strings.TrimSpace` calls"*. The AC removes the latitude the plan grants run-phase, and it asserts implementation shape rather than behavior. Answering the dispatch's point 6 directly: the seam earns its keep **weakly** — `codexReviewTextIsBlank` is a three-line wrapper over a stdlib one-liner (Simplicity ladder rung 3), and four inline `strings.TrimSpace(s) == ""` calls are the simpler correct answer on the ladder. The one argument that does carry: the defect's own shape is "three sites repaired, one missed", and a named helper makes that mechanically greppable — which is a real, non-speculative justification. Keep the helper as a recommendation; do not let an AC mandate it. — Severity: **minor** — Class: **blocking** — Required fix: reword AC-CBR-009 to assert the *behavior* (all four sites treat a whitespace-only body identically, shown by fixture) and leave the helper-vs-inline choice to plan §B.2 as written.

**D8** — `plan.md:141` and `plan.md:191` — Both cite `` `acceptance.md` §D `` for the three-state control matrix. The matrix is at **acceptance.md §A** (`## §A The control matrix — the substance of this repair`, acceptance.md:22); `## §D Quality gates` is at acceptance.md:154. Verified: `/usr/bin/grep -n '^## ' acceptance.md`. — Severity: **minor** — Class: **optional** — Required fix: change both to `acceptance.md §A`.

**D9** — `plan.md:35` — Cites `internal/cli/codex_review_rpc_test.go:113-127` as the enclosing function range. Measured: `func TestSynthesizeReviewOutput_FindingBulletsMapToFail` spans **114-126**; line 113 is the tail of the doc comment and 127 is past the closing brace. The load-bearing citation in the same sentence — the pinned case `"": "pass"` at `:119` — **is correct** (verified verbatim). — Severity: **minor** — Class: **optional** — Required fix: change to `:114-126`, or drop the range and keep only `:119`.

**D10** — `.moai/reports/t551/issue-1632-axis-status-20260908.md` — The supporting report's line citations disagree with the corrected ones in `spec.md`: the report says the guard is at `:818` (spec.md says `:817` — measured: `:817` is the `if reviewText == ""`, `:818` the return), the unrecognized-verdict default at `:1411` (spec.md `:1409` — measured: `:1409` is `func codexUnrecognizedVerdict`, `:1411` the `return "pass"`), `applyGateUnmet` at `:1603` (measured `:1608`; `:1602` starts its comment), `GateUnmet` at `:283` (measured `:291`; `:283` is inside the field's comment), and `SynthesisNote` at `:1499` (measured: `:1499` is blank; the `SynthesisNote` field is at `:276` and `describeSignalDivergence`'s comment starts at `:1500`). `progress.md:10-14` claims *"Every file:line citation in the artifacts was re-measured"* — true for the SPEC artifacts, not for the cited evidence file. — Severity: **minor** — Class: **optional** — Required fix: re-measure the five citations in the report, or add a line stating the report's citations point at the enclosing block rather than the exact statement.

**INFO (not a defect)** — the pre-repair `mcp__moai__spec_audit` classified this SPEC `V3R5`/H-3/grandfathered; post-repair it classifies `V3R6`/H-5/`modern_era_clean: 1`. This is exactly the behavior `lifecycle-sync-gate.md` § Conservative default predicts: the unparseable frontmatter made `created:` unreadable, both H-5 signals went empty, and the SPEC kept grandfather protection until the frontmatter was repaired. Recorded because it shows the D1-repaired defect also silently suppressed drift detection.

---

## Verified-correct (audited and found sound — recorded so a re-audit does not re-litigate)

- **Every `mcp_codex.go` / `mcp_convergence.go` citation in the three SPEC artifacts is accurate.** All checked by `awk 'NR>=X && NR<=Y {printf "%5d| %s\n", NR, $0}'` against this tree: `:817` (`if reviewText == ""`), `:1134` (`p.Item.Review != ""`), `:1137` (`p.Item.Text != ""`), `:1253` (`if review != ""`), `:1409` (`func codexUnrecognizedVerdict`), `:1252-1257` (`bestCodexReviewText`), `:1386-1394` (`adoptConservativeVerdict`), `:1395-1415`, `:1409-1415`, `:1432-1435` (the t234 NOTE, verbatim), `:1466`, `:1608` (`func applyGateUnmet`), `:790-796` (the quoted comment *"an inconclusive that cannot be told apart from 'codex is missing' is a new silence, not a repair"* — verbatim at `:794-795`), `mcp_convergence.go:190` (`case allPass(required)`), `:186-200`, and `codex_review_rpc_test.go:119` (`"": "pass"`). The two dispatch-supplied off-by-ones (`:818`→`:817`, `:120`→`:119`) were genuinely corrected. This is the strongest part of the SPEC.

- **Dispatch point 5 — axes 1 and 2 verified closed against the CODE, not the comments.** Axis 1: `synthesizeReviewOutput` assigns `Findings: codexFindingsOf(reviewText)` at `mcp_codex.go:1444`, and `codexFindingsOf` parses severity bullets via `codexFindingLine` (`:1457`) — the field is genuinely populated, not merely annotated. Axis 2: a rejected request returns `inconclusiveReview("codex " + method + " rejected: " + …)` at `:803`, so a refusal is not reported as a review that passed; conservative adoption is implemented at `:1386-1394`. Both have dedicated tests (`codex_findings_parse_test.go`, `audit_blind_verdict_test.go`). The exclusions at spec.md:186-194 are sound.

- **plan.md §B.1 option (a) is the correct recommendation, and its stated risk is honest.** `/usr/bin/grep -n 'synthesizeReviewOutput' internal/cli/*.go | grep -v _test.go` returns exactly **one** production call site: `mcp_codex.go:820`, immediately behind the guard at `:817`. So the "Against" at plan.md:54-56 (*"the synthesizer remains lenient if some future caller reaches it directly"*) is currently a hypothetical with no live instance, and option (b)'s blast-radius concern (plan.md:70-71) is correctly characterized as unsurveyed. Option (a) is the smaller, correct change.

- **AC-CBR-007's premise holds.** `applyGateUnmet` is applied at `mcp_codex.go:1596` to the output of `codexReviewRPC` with the error discarded (`out, _ :=`), so a guard-returned `inconclusive` does reach it and `out.GateUnmet` is set at `:1615`. The criterion is reachable through `handleCodexAudit` as written.

- **AC-CBR-006 is not vacuous *as a pin*, and the SPEC already knows it.** It is satisfied on the unrepaired tree (blank → `pass` ≠ `fail`), but acceptance.md:159-161 deliberately excludes it from the RED list and acceptance.md:103-104 states its purpose is to pin §C against later drift to fail-closed. Its fixture set is provably non-empty (four fixtures enumerated at acceptance.md:53). Correctly scoped; no finding.

- **The A/B/C control matrix (acceptance.md:30-40) is the right instrument, and AP-2 (plan.md:173-175) names the hazard correctly.** Requiring all three states in one run, with A and C separated by Summary rather than verdict, is precisely what stops "blocked everything" from reading as "blocked the right thing". D2 and D3 are defects in *which sites* the matrix discriminates, not in the matrix design.

---

## Recommendation

FAIL. Six blocking findings; four optional. Fix in this order — the first three are the substantive
ones, and D2+D3 are one repair:

1. **D2 + D3 together.** Correct acceptance.md:73-75 to name `:1253` (+ guard) as what AC-CBR-003
   discriminates, soften spec.md:80-87 correspondingly, and add the collection-overwrite AC that
   gives REQ-CBR-001/002 real coverage. Without this, run-phase can land M2 believing the collection
   sites are tested when nothing exercises them.
2. **D4.** Amend spec.md:91-95 and 168-171 to state the measured convergence outcome:
   `overall_verdict` stays `pass` via the `default` fail-open arm (`mcp_convergence.go:194-199`);
   what becomes truthful is the per-backend verdict, the Summary, and `gate_unmet`. spec.md:173-175
   already frames this correctly — make §A and §C.3 agree with it, and reconcile with §E:218-221.
3. **D5 + D6 + D7 together.** Rebuild AC-CBR-009: name the sweep scope mechanically, add a positive
   control (4 call sites; non-zero grep count pre-repair), give it a REQ or move it to §D, and assert
   behavior rather than the helper's existence.
4. **D8, D9, D10** are one-line citation corrections; fold them into the same pass.
5. **D1** needs no artifact change — the lead confirms the authoring session is idle before the
   re-audit is dispatched.

On the verdict: the aggregate 0.81 clears the Tier M threshold of 0.80 and the M5 must-pass firewall
is clean, so this is **not** a mechanical FAIL. I am issuing FAIL deliberately on D2 and D4, which
are false statements rather than gaps — one about what the SPEC's own key control criterion proves,
one about the repair's consumer-visible effect, each of which would propagate into run-phase as a
wrong belief that no later gate catches. Per M6 this is not a FAIL manufactured from optional
findings: six findings are classed blocking on correctness and internal-consistency grounds, and the
ten-item defect list above is the machine-consumable fix route. Iteration 2 (the Tier M ceiling) is
scoped to this delta plus a regression check, not a from-scratch re-audit.

If the orchestrator judges the score sufficient and prefers to proceed, the honest alternative is
PASS-with-debt carrying D2, D3, D4 and D5 as recorded debt — but D2 and D3 must be fixed before M2
lands either way, because M2 is the milestone they misdescribe.

---

## Evidence Report (5-section)

### Claim

SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 at `3ac58b5a1` passes all seven must-pass criteria on its
post-repair state and scores 0.81, but carries six blocking defects — two false claims (D2, D4), one
uncovered requirement pair (D3), one vacuous-capable criterion (D5), one orphan AC (D6), one
AC/plan contradiction (D7) — plus a mid-audit artifact mutation (D1). Verdict FAIL.

### Evidence

Commands run in `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t551`, outputs verbatim:

- `git rev-parse --show-toplevel; git rev-parse --short HEAD; git branch --show-current` →
  `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t551` / `3ac58b5a1` / `WT-audit-fail-open`
- `moai spec lint .moai/specs/SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001/spec.md; echo "EXIT=$?"` →
  `✓ No findings — all SPEC documents are valid` / `EXIT=0` *(post-repair; pre-repair the same
  command returned `ERROR ParseFailure … line 13: cannot unmarshal !!seq into string`)*
- `mcp__moai__spec_audit(project_root=<this worktree>, filter_spec=SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001)` →
  `{"total_specs":1,"grandfathered":0,"modern_era_clean":1,"drift_findings":[{"era":"V3R6","finding_type":"EraAutoDetected","severity":"INFO","details":{"heuristic_matched":"H-5 (modern phase or created date)"}}]}`
- `/usr/bin/grep -ho 'REQ-CBR-[0-9]*' spec.md | sort -u` → `REQ-CBR-001` … `REQ-CBR-009`;
  `/usr/bin/grep -c '^\*\*REQ-CBR-' spec.md` → `9`
- `/usr/bin/grep -c '^### AC-CBR-' acceptance.md` → `9`
- `/usr/bin/grep -Eoh 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md plan.md acceptance.md progress.md | sort -u` →
  `SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001`
- `/usr/bin/grep -c 'syscall' spec.md plan.md acceptance.md progress.md` → all `0`
- `/usr/bin/grep -rn '\[NEEDS CLARIFICATION' plan.md` → no match; `ls research.md` → absent
- `/usr/bin/grep -n '^### Out of Scope' spec.md` → `184`, `196`, `204`
- `/usr/bin/grep -n '^## ' acceptance.md` → `22:## §A The control matrix …` … `154:## §D Quality gates`
- `/usr/bin/grep -n 'acceptance' plan.md` → `141: … from \`acceptance.md\` §D …`, `191: - \`acceptance.md\` §D …`
- `/usr/bin/grep -n 'synthesizeReviewOutput' internal/cli/*.go | /usr/bin/grep -v _test.go` →
  one production call site, `internal/cli/mcp_codex.go:820`
- `/usr/bin/grep -n 'applyGateUnmet' internal/cli/*.go` → `mcp_codex.go:1574`, `:1596`, `:1608`
- `awk 'NR>=X && NR<=Y {printf "%5d| %s\n", NR, $0}' internal/cli/mcp_codex.go` for
  785-825, 1030-1042, 1095-1145, 1245-1262, 1380-1420, 1425-1440, 1460-1470, 1495-1502, 1598-1632,
  1844-1852; same over `internal/cli/mcp_convergence.go` 178-202 and via
  `awk '/func claudeVerdictOrDefault/,/^}/'` (369-376), `/func allPass/` (327-334),
  `/func filterRequired/` (285-293); `internal/cli/codex_review_rpc_test.go` 110-130
- `awk 'NR>=1157 && NR<=1235' internal/spec/lint.go` (FrontmatterSchemaRule.Check);
  `/usr/bin/grep -n 'yaml:"tags"' internal/spec/` → `internal/spec/lint.go:511: Tags string`
- `ls -lT .moai/specs/SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001/` → mtimes 13:44:27–13:44:40 (D1)

Key verbatim lines the findings rest on:

```
  817|	if reviewText == "" {
  820|	return synthesizeReviewOutput(reviewText, method), nil
 1134|			if p.Item.Type == "exitedReviewMode" && p.Item.Review != "" {
 1137|			if p.Item.Type == "agentMessage" && p.Item.Text != "" {
 1252| func bestCodexReviewText(review, agent string) string {
 1253|	if review != "" {
```
```
  190|	case allPass(required):
  194|	default:
  199|		overall = claudeVerdictOrDefault(verdicts, overallVerdictPass)
  369| func claudeVerdictOrDefault(vs []PerBackendVerdict, dflt string) string {
  372|			return v.Verdict
```

### Baseline-attribution

Every measurement above was taken in this run, in `.claude/worktrees/t551`, against
`git rev-parse --short HEAD` = `3ac58b5a1`, on 2026-09-08. No figure is carried over from another
tree, package, or run. **Split attribution on one axis:** the MP-3 frontmatter row is attributed to
the artifacts as they stood at 04:46 UTC (post-repair); the pre-repair `ParseFailure` measurement is
attributed to their state at ~04:36 UTC. The code citations were re-read after the artifact rewrite
and are unaffected by it (no `internal/` file changed — `git status --short` shows only the two
untracked SPEC/report directories). The two probe files were read, not re-run, per dispatch.

### Gaps — explicitly NOT observed

- **No Go test was written or executed by this audit.** The AC-CBR-003 discrimination trace (D2) and
  the convergence-arm trace (D4) are **code reads**, not test results. They are strong reads — every
  line is quoted verbatim above and the control flow is straight-line — but run-phase must confirm
  D2 by actually running AC-CBR-003 against a `:1253`-only repair, and D4 by driving a blank body
  through `audit_multi`.
- **The two probe files were consumed, not reproduced.** Per dispatch I did not re-run
  `TestT551ProbeReachability` / `TestT551ProbeEmptyish`; their outputs are taken as given.
- **No live `codex` binary was exercised.** Same gap the SPEC records at spec.md:222-223.
- **`HandleCodexReviewGate`** (the Stop-hook path, which calls `runCodexReviewRPC` directly per
  `mcp_codex.go:841-843`) was noted but **not** traced. It does not pass through `applyGateUnmet`,
  so whether the blank-output repair changes its behavior is unmeasured. Out of this SPEC's stated
  scope, but nothing in the artifacts mentions it.
- **The U+00A0 sub-question** (plan.md:89-91) was not independently verified against
  `unicode.IsSpace`; the plan's claim that `strings.TrimSpace` cuts U+00A0 was taken as stated.
- **Cross-model audit was not run.** No `mcp__moai__audit_multi` / `codex_audit` / `glm_audit` call
  was made — this is a Claude-anchor-only verdict. If the project's `audit_model` is `multi` or
  `codex+glm`, a second opinion is owed and was not obtained.
- **I did not verify who performed the mid-audit edit** (D1), only that it happened and what it
  changed.
- **Tier ceiling read from doctrine, not from config.** I cited the Tier M threshold `0.80` from
  `spec-workflow.md` § SPEC Complexity Tier; I did not read
  `.moai/config/sections/harness.yaml` → `plan_audit_tier_ceilings` for the iteration ceiling, so
  "1/2" is inferred from the doctrine table rather than measured.

### Residual-risk

- **The audit's own baseline moved once and could move again.** If the artifacts are edited between
  this verdict and its consumption, the defect list goes stale silently — a line citation decays
  exactly like a HEAD reading. Every D-number above carries its line; re-read before acting.
- **D2 is the finding most likely to be argued with, and it is the one the FAIL rests on.** If a
  reviewer disagrees, the cheap resolution is mechanical, not rhetorical: repair `:817` and `:1253`
  only, run AC-CBR-003, and observe whether it passes. My trace says it does.
- **D4 could be read as scope-policing rather than a defect.** The SPEC does scope gate enforcement
  to a sibling card (spec.md:198-200), and spec.md:173-175 already says "fail-open, but honestly
  labelled". The defect is narrow and specific: §A:94-95 and §C.3 claim more than that, and §E
  contradicts them on evidentiary status. Fixing it is three sentences.
- **Judging AC-CBR-009 "vacuous-capable" is a hazard claim, not an observed failure.** I did not
  write the grep and watch it return zero against a wrong scope; I observed that no scope is defined
  and no positive control is specified. That is sufficient under
  `verification-completeness.md` §1.1, but it is a structural read, not a demonstration.
- **A clean `moai spec lint` is not a clean SPEC.** The linter checks frontmatter, GEARS modality,
  AC→REQ coverage, and Out-of-Scope presence. Every finding D2-D10 sits outside what it can see —
  which is exactly why an adversarial read was commissioned, and why "lint is green" must not be
  cited as evidence against this verdict.

---
---

# SPEC Review Report: SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 — ITERATION 2

Card: t551 · Tree: `.claude/worktrees/t551` · Branch: `WT-audit-fail-open` · HEAD: `3ac58b5a1`
Iteration: 2/2 (Tier M ceiling — this is the final iteration available)
Auditor: plan-auditor (Claude anchor; `audit_model` single-backend — no `audit_multi` fan-out)

**Verdict: PASS**
**Overall Score: 0.93** (Tier M PASS threshold 0.80)
**Delta vs iteration 1: 0.81 → 0.93 = +0.12.** The score ROSE. No STOP signal, no
scope-reduction proposal — the LEAN score-regression clause does not fire.

Reasoning context ignored per M1 Context Isolation. The dispatch's five focus points were treated
as audit direction only; no author reasoning was consulted. Every claim below cites a command and
its observed output.

## Artifact stability (D1 follow-up)

mtimes recorded at audit open and at audit close, `stat -f '%N %m'`:

| Artifact | open | close |
|---|---|---|
| spec.md | 1788843131 | 1788843131 |
| plan.md | 1788843171 | 1788843171 |
| acceptance.md | 1788843181 | 1788843181 |
| progress.md | 1788843212 | 1788843212 |
| issue-1632-axis-status-20260908.md | 1788843196 | 1788843196 |

`git rev-parse --short HEAD` → `3ac58b5a1` at open and at close. **No artifact moved during this
audit.** Iteration 1's D1 process defect did not recur; this verdict is attributed to the state
that was dispatched.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `/usr/bin/grep -o 'REQ-CBR-[0-9]*' spec.md | sort -u`
  → `REQ-CBR-001 … REQ-CBR-009`, 9 unique, sequential, no gaps, no duplicates, uniform 3-digit
  padding.

- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in
  `spec.md`) per M3 § Scope. The nine requirements are unchanged from iteration 1 in wording and
  survive the repair pass: event-driven (001/002/003/004), ubiquitous (005), state-driven (006),
  unwanted (007/008), compound Where+When (009). The `Given/When/Then` entries in `acceptance.md`
  are `AC-XXX` verification-layer artifacts and were **not** graded here.

- **[PASS] MP-3 YAML frontmatter validity** — `moai spec lint --strict
  .moai/specs/SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001/spec.md` → `✓ No findings — all SPEC
  documents are valid`, `EXIT=0`. All 12 canonical fields present (`spec.md:2-13`); `tags` a quoted
  CSV; `status:` absent from `plan.md`/`acceptance.md` per § Artifact Statelessness. The
  iteration-1 repair held.

- **[N/A] MP-4 Section 22 language neutrality** — single-language SPEC (`module: internal/cli`,
  spec.md:11); every implicated site is Go. No multi-language tooling claim.

- **[PASS] MP-5 D7 cross-SPEC reconciliation** — no external `SPEC-…` reference in any of the four
  artifacts. No retired/superseded reconciliation owed. No BLOCKING finding.

- **[PASS] MP-6 D8 cross-platform discipline** — `syscall` appears in none of the four artifacts.
  D8-4 auto-PASS.

- **[PASS] MP-7 clarification gate** — no `[NEEDS CLARIFICATION` marker in `plan.md`;
  `research.md` absent (Tier M does not require it).

---

## Regression Check — iteration-1 defect delta

| # | Iteration-1 finding | Disposition | Evidence |
|---|---|---|---|
| D1 | artifacts rewritten mid-audit | **RESOLVED (process)** | mtimes identical open↔close (table above); recorded at progress.md:51-52 |
| D2 | AC-CBR-003 discriminating-power claim false for 2 of 3 sites | **RESOLVED** | acceptance.md:74-85 now names `:1253` alone and states explicitly "It says NOTHING about the collection sites". Second instance split: plan.md:141-144 now reads "The collection sites `:1134`/`:1137` are pinned by AC-CBR-009 (the overwrite case) — NOT by AC-CBR-003, which passes even when they are left unrepaired." Traced and confirmed true against `mcp_codex.go:1134/1137/1252-1257`. Residual precision note: F3 below |
| D3 | REQ-CBR-001/002 had zero coverage | **RESOLVED** | New AC-CBR-009 (acceptance.md:138-157) pins the overwrite case, `Covers REQ-CBR-001 and REQ-CBR-002` (acceptance.md:149), added to the §D RED list (acceptance.md:181-185). Mechanism independently verified — see § Mechanism verification |
| D4 | headline overstated the consumer-visible consequence | **RESOLVED** | spec.md:91 carries `[HARD] It does NOT close the gate hole, and this SPEC does not claim it does.`; the `default:` → `claudeVerdictOrDefault` trace at spec.md:93-99 matches `mcp_convergence.go:185-200` verbatim; §C.3 (spec.md:187-193) and §E (spec.md:240-247) reconciled. No over-correction — the four honest effects are stated and all four verified true (§ Mechanism verification) |
| D5 | old AC-CBR-009 vacuous-capable grep | **RESOLVED** | Demoted to plan.md §E:168-179 as a checklist item; the review-text path is ENUMERATED (`:817`, `:1134`, `:1137`, `:1253`) rather than grep-defined; plan.md:177-179 requires the grep be shown to MATCH pre-repair before mechanization |
| D6 | orphan AC naming no REQ | **RESOLVED** | All 9 ACs name a REQ (`grep -n 'Covers REQ'` → 9 lines); all 9 REQs covered; no orphan |
| D7 | an AC hard-mandated the helper | **RESOLVED** | plan.md:92 `[HARD] **No acceptance criterion mandates the helper.**`; helper is RECOMMENDED at plan.md:87-90. Read all nine ACs — none asserts implementation shape |
| D8 | `acceptance.md` §D → §A (2 sites) | **RESOLVED** | plan.md:153 and plan.md:213 both now cite `acceptance.md` §A; no residual §D reference for the matrix |
| D9 | test range `:113-127` | **RESOLVED** | plan.md:34-35 now `:114-126`. Measured: `func TestSynthesizeReviewOutput_FindingBulletsMapToFail` spans exactly 114-126; `"": "pass"` at `:119` |
| D10 | supporting-report citations off | **RESOLVED** | Re-measured all five: `:817` (`if reviewText == ""`) ✓, `:1409` (`func codexUnrecognizedVerdict`) ✓, `:276` (`SynthesisNote` field) ✓, `:1446` (`SynthesisNote:` assignment) ✓, `:291` (`GateUnmet` field) ✓, `:1608` (`func applyGateUnmet`) ✓. Residual: F5 below |

**No iteration-1 defect is unresolved. No stagnation.**

---

## Mechanism verification (repair claims traced against code, not accepted on report)

**D3 / AC-CBR-009 — can it genuinely go red?** YES.
`/usr/bin/sed -n '1125,1139p' internal/cli/mcp_codex.go`:

```
1134:			if p.Item.Type == "exitedReviewMode" && p.Item.Review != "" {
1135:				reviewText = p.Item.Review
1136:			}
1137:			if p.Item.Type == "agentMessage" && p.Item.Text != "" {
1138:				agentText = p.Item.Text
1139:			}
```

The assignment is unconditional reassignment, and `"   " != ""` is true, so a LATER whitespace-only
item **does** overwrite an earlier real one on the pre-repair tree. The surviving text is
observable: `mcp_codex.go:1443` `Summary: strings.TrimSpace(reviewText)`, so a clobbered real
review yields `Summary == ""`. AC-CBR-009 is adoptable — its RED is carried by the survival half.
The acceptance.md §D justification (`:183-185`) is accurate.

**D4 — the four honest effects, each traced:**

1. *"codex leaves the all-pass arm at `:190`"* — `mcp_convergence.go:190` `case allPass(required):`;
   `allPass` (`:327-334`) returns false on any `v.Verdict != "pass"`. TRUE.
2. *"codex is listed in `FailOpenBackends` (`:133`, populated at `:280`)"* — field at `:133`,
   `collectFailOpen` at `:382-393` appends any `VerdictInconclusive` backend, called at `:280`. TRUE.
3. *"codex carries `GateUnmet` when the gate is `required` (`:1608`)"* — `applyGateUnmet`
   (`:1608-1617`) returns early unless `out.Verdict == VerdictInconclusive` AND the gate is
   `config.AuditGateRequired`; sets the field at `:1615`. TRUE.
4. *"codex's Summary names the blank body as the cause"* — REQ-CBR-005 + AC-CBR-005. Design claim,
   correctly scoped as a requirement rather than an observation.

No over-correction: the SPEC does not claim the repair achieves nothing, and each of the four
claimed effects is real.

**Dispatch point 2 — iteration-1's confirmed-sound items all SURVIVED unchanged:**

- All code citations re-measured accurate: `:817`, `:1134`, `:1137`, `:1253`, `:1409`,
  `:1252-1257`, `:1432-1435`, `:1466`, `:1608`, `:790-796` (quote verbatim at `:794-795`),
  `mcp_convergence.go:190`, `codex_review_rpc_test.go:119`. New in this iteration and also
  accurate: `:1125-1139`, `:1596`, `mcp_convergence.go:133`/`:280`, `:114-126`. spec.md's
  `mcp_convergence.go:185-200` is MORE precise than iteration 1's `:186-200` (185 is `switch {`).
- plan.md §B.1 option (a): `/usr/bin/grep -rn "synthesizeReviewOutput" internal/` → exactly ONE
  production caller, `internal/cli/mcp_codex.go:820`, immediately behind the guard at `:817`.
  Every other hit is a `_test.go` file or the definition itself. Claim survives.
- AC-CBR-007's premise: `mcp_codex.go:1595` `out, _ := codexReviewRPC(...)` → `:1596`
  `out = applyGateUnmet(out, root)`. The guard's output does reach it. Survives.
- AC-CBR-006 correctly excluded from the RED list — verified green on both sides (pre-repair blank
  → `pass` ≠ `fail`; post-repair `inconclusive` ≠ `fail`). Genuine pin.

---

## Dispatch point 3 — does the control matrix still discriminate?

Re-asked with AC-CBR-003 scoped to `:1253` and AC-CBR-009 carrying the collection sites.

**Could the suite pass while the repair BLOCKS a real clean review?** No.

- AC-CBR-002 (state B) requires real clean prose → `pass` with the prose in Summary.
- AC-CBR-003 requires blank-structured + real-agent → `pass`.
- AC-CBR-008 requires `"I looked at the diff."` → `pass` (matches the recorded probe row
  `prose-no-signal verdict="pass"`), which catches an over-blocking widening into
  `codexUnrecognizedVerdict`.
- AC-CBR-009 requires the real-then-blank stream → `pass`.

An over-blocking repair fails **AC-CBR-002** first, and **AC-CBR-008** if it widens to the
synthesizer.

**Could the suite pass while it leaves the defect LIVE?** Only in one named shape.

- A no-op repair fails **AC-CBR-001** (blank → `inconclusive`, four fixtures).
- Collapsing blank into the unavailable wording fails **AC-CBR-005**.
- Leaving the collection sites unrepaired fails **AC-CBR-009** (real-then-blank ordering).
- **The residual**: a *first-wins* collection mutant (record only when the slot is still empty)
  satisfies AC-CBR-009 in BOTH stated orderings while recording a blank in the blank-first
  ordering — see F1. That is the one under-blocking shape no AC catches.

Separately noted, not a gap in the matrix: with `:1134`/`:1137` repaired, `:1253` can be left
unrepaired and AC-CBR-003 still passes (the selector is never offered a blank). AC-CBR-003's
discriminating power over `:1253` is therefore conditional on the collection sites being
unrepaired — see F3.

---

## Dispatch point 4 — RED-list integrity

**RED-listed (acceptance.md:180-185): AC-CBR-001 / 003 / 005 / 009.** Each traced on the
pre-repair tree:

| AC | Genuinely red pre-repair? | Which half carries the RED |
|---|---|---|
| AC-CBR-001 | YES | verdict: blank → `pass` today, `inconclusive` required |
| AC-CBR-003 | YES, but on ONE half only | "the agent text is selected"; the verdict half is ALREADY `pass` pre-repair (blank selected → synthesizer → native default `pass`). See F2 |
| AC-CBR-005 | YES | state-A Summary is `""` pre-repair, so it does not name blank review output |
| AC-CBR-009 | YES, on ONE half only | "the real review text survives"; the verdict half is already `pass` pre-repair. §D:183-185 states this reason correctly |

**Excluded as preservation pins (acceptance.md:186-189): AC-CBR-002 / 004 / 006 / 008.** All four
verified green on BOTH sides — AC-CBR-002 (real clean → `pass`), AC-CBR-004 (unavailable →
`inconclusive`, unchanged path), AC-CBR-006 (blank ≠ `fail` before and after), AC-CBR-008
(non-blank unrecognized → `pass`, corroborated by `probe-synthesizer-20260908.txt`). Each exclusion
is correct: a pin that went red pre-repair would mean the repair changed behavior it must preserve.

**Verdict on both halves: the RED list and the exclusion list are correctly populated.** The one
integrity gap is that §D states the RED reason for AC-CBR-009 but not for AC-CBR-003, whose verdict
half is green pre-repair (F2).

---

## Dispatch point 5 — mutant probe on AC-CBR-003 and AC-CBR-009

**AC-CBR-009 — a mutant IS writable.** Replace the collection sites with first-wins rather than
blank-awareness:

```go
if p.Item.Type == "exitedReviewMode" && p.Item.Review != "" && reviewText == "" { reviewText = p.Item.Review }
```

Against AC-CBR-009's stated stream (real THEN blank), the real text survives and the verdict is
`pass` — the criterion is satisfied. But REQ-CBR-001 says the collector *"shall treat that field as
absent and shall not record it"*: under first-wins with a blank arriving FIRST, the blank IS
recorded and the real item that follows is discarded. No other AC catches this — AC-CBR-001 still
returns `inconclusive` (the repaired selector falls through to an empty agent text) and AC-CBR-003
still returns `pass`. The mutant survives the whole suite while violating REQ-CBR-001/002.
Reported as **F1** (`verification-completeness.md` §2 mutant probe).

**AC-CBR-003 — no mutant found against its own requirement.** The obvious candidate ("always prefer
`agentMessage` when non-empty") satisfies AC-CBR-003 and does NOT violate REQ-CBR-003, which
mandates fall-through only for a blank structured review. It does invert the documented preference
at `mcp_codex.go:1250-1251` — but no requirement states that preference, so it is a preservation
gap rather than a mutant. Reported as **F4** (optional).

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.90 | between 0.75 and 1.0, near 1.0 | Iteration 1's two FALSE prose claims (D2, D4) are repaired and independently re-traced. Requirements themselves carry a single interpretation (spec.md:122-165). Held below 1.0 by two conditional-precision residuals in supporting prose: acceptance.md:74-78 and spec.md:80-87 (F3) |
| Completeness | 1.00 | 1.0 | HISTORY spec.md:20-24; §A Context spec.md:28; §B Requirements spec.md:118; §C decision spec.md:169; three `### Out of Scope — <topic>` H3s at spec.md:206/218/226, each with `-` bullets; §E Gaps spec.md:238. Frontmatter clean (`moai spec lint --strict` exit 0) |
| Testability | 0.80 | between 0.75 and 1.0 | The vacuous-capable grep AC is gone (D5). All nine ACs are behavioral and binary; no weasel words. Held below 1.0 by F1 (writable mutant on AC-CBR-009) and F2 (AC-CBR-003's RED reason unstated, verdict half green pre-repair) |
| Traceability | 1.00 | 1.0 | All 9 REQs covered: 001+002→AC-009, 003→AC-003, 004→AC-001, 005→AC-005, 006→AC-004, 007→AC-002+AC-008, 008→AC-006, 009→AC-007. `grep -n 'Covers REQ'` → 9 lines, no orphan AC. `moai spec lint --strict` emits no `CoverageIncomplete` |

Aggregate = (0.90 + 1.00 + 0.80 + 1.00) / 4 = **0.925 → 0.93**

---

## Defects Found (structured defect-list)

**F1** — `acceptance.md:138-157` — **AC-CBR-009 admits a first-wins mutant.** The criterion states
only the real-THEN-blank ordering. A collection implementation that records only into an empty slot
(`&& reviewText == ""`) satisfies AC-CBR-009 in both its stated orderings while recording a blank
in the blank-FIRST ordering, violating REQ-CBR-001's *"shall treat that field as absent and shall
not record it"* (spec.md:122-125) and REQ-CBR-002 (spec.md:127-130). No other AC catches it (traced:
AC-CBR-001 still yields `inconclusive`, AC-CBR-003 still yields `pass`). Per
`verification-completeness.md` §2 a criterion admitting such a mutant is too shallow to adopt.
— Severity: **major** — Class: **blocking** — Required fix: add one clause to AC-CBR-009 —
*"**And given** a stream whose FIRST `exitedReviewMode.review` is whitespace-only followed by a
REAL one, **Then** the real review text is the collected review text"* (and the `agentMessage`
mirror). One sentence; no requirement or scope change.

**F2** — `acceptance.md:180-185` — **The §D RED list states its RED reason for AC-CBR-009 but not
for AC-CBR-003, whose verdict half is already green pre-repair.** Traced on the unrepaired tree: a
blank `exitedReviewMode.review` plus a real `agentMessage.text` yields `bestCodexReviewText` →
`"   "` → guard `== ""` false → `synthesizeReviewOutput("   ", review/start)` → no signal →
`codexUnrecognizedVerdict` → **`pass`**, which is what AC-CBR-003's `Then` requires. The RED is
carried solely by the *"the agent text is selected"* half, observable through
`Summary = strings.TrimSpace(reviewText)` (`mcp_codex.go:1443`). A run-phase test written to assert
the verdict alone yields a vacuous green — the exact hazard `verification-completeness.md` §1.1
names. — Severity: **major** — Class: **optional** — Required fix: extend the §D note that already
does this for AC-CBR-009 to cover AC-CBR-003: state that its RED is carried by the selected-text
assertion, not by the verdict.

**F3** — `acceptance.md:74-78` and `spec.md:80-87` — **Two residual conditional-precision claims
about site load-bearingness.** (a) acceptance.md:74-78 states AC-CBR-003 *"fails when the guard
(`:817`) is repaired without the SELECTION site (`:1253`)"*. Traced: this holds only while the
collection sites stay unrepaired. With `:1134` repaired and `:1253` NOT repaired, the selector is
never offered a blank, `bestCodexReviewText` returns the agent text, and the criterion PASSES. The
claim is true under the plan's own M2 ordering (collection + selection together) but not as an
unconditional statement. (b) spec.md:80-87 opens *"The collection and selection sites are
load-bearing and are NOT optional extras"* and then supports only the selection site
(`bestCodexReviewText` shadowing), closing *"The three sites are one repair."* The collection
sites' actual justification — the overwrite case — now lives in acceptance.md:151-157 and
plan.md:141-144 but appears nowhere in spec.md §A. The topic sentence is TRUE; its supporting
argument is one site short. — Severity: **minor** — Class: **optional** — Required fix: (a) add
*"…while the collection sites remain unrepaired"* to acceptance.md:75; (b) add one sentence to
spec.md §A naming the overwrite case as the collection sites' own reason.

**F4** — `acceptance.md` (no AC exists) — **The structured-over-agent preference is unpinned.**
`mcp_codex.go:1250-1251` documents that `bestCodexReviewText` *"prefers the structured
exitedReviewMode review over the free-form agentMessage text"*, and spec.md:81-83 leans on that
preference to argue the shadowing case. No AC pins it: an implementation that always prefers the
agent text when non-empty satisfies AC-CBR-003 and every other criterion while inverting the
documented order. This is a preservation gap, not a mutant against a stated requirement — no REQ
declares the preference. — Severity: **minor** — Class: **optional** — Required fix: either add a
preservation pin (*real structured review + real agent text → the structured text is selected*) to
the AC set, or state in §D that the preference order is deliberately left unpinned by this card.

**F5** — `.moai/reports/t551/issue-1632-axis-status-20260908.md:21` — **One supporting-report
citation still points at a doc comment rather than the symbol.** The row cites
`internal/cli/mcp_codex.go:1465` for `codexFindingsOf`; measured, `:1465` is the first line of that
function's doc comment and `/usr/bin/grep -n 'func codexFindingsOf'` returns **`:1473`**. The five
citations iteration 1 flagged (D10) are all correctly repaired; this one was not in that list. The
sibling citation in spec.md:211 is honest about it (*"documented at …:1466"*); the report's is not
labelled. — Severity: **minor** — Class: **optional** — Required fix: change to `:1473`, or label
it as the doc comment.

**F6** — `progress.md:47-48` — **A line citation went stale inside the repair record itself.** It
states *"REQ-CBR-009 IS referenced, by AC-CBR-007 (`acceptance.md:112`)"*. Measured: acceptance.md
`:112` is the "no verdict equals fail" line inside AC-CBR-006; the AC-CBR-007 → REQ-CBR-009
reference is at **`:123`**. The repair pass rewrote AC-CBR-003 and shifted the file
below it. The underlying claim (REQ-CBR-009 is covered) is TRUE — only its coordinate is wrong.
— Severity: **minor** — Class: **optional** — Required fix: `:112` → `:123`.

---

## Verdict rationale

The M5 must-pass firewall is clean: seven criteria, six PASS and one N/A, each with a cited
command. The aggregate is 0.93 against a Tier M threshold of 0.80. All six iteration-1 blocking
defects are resolved and were re-traced against the code rather than accepted on the repair report,
and iteration 1's confirmed-sound items all survived the repair unchanged — including the sixteen
code citations, which the dispatch flagged as the most likely regression surface. The repair pass
introduced no new false statement; the five findings below F1 are precision and depth refinements,
not corrections.

One finding (F1) is classified blocking. It is not being converted into a FAIL: per M6, the verdict
is anchored to the must-pass firewall and the rubric scores, and F1 is a one-clause addition to a
criterion that is already adoptable and already red on the pre-repair tree. That is materially
different from iteration 1's delta, which carried two false statements in requirement-supporting
prose and a requirement PAIR with zero coverage. The blocking class here is a routing signal: F1
should be fixed before run-phase M2 begins, and the fix needs no re-audit.

This is iteration 2 of a Tier M ceiling of 2. No third iteration is available and none is warranted.

---

## Evidence Report (5-section)

### Claim

1. All seven must-pass criteria PASS or are N/A at HEAD `3ac58b5a1`.
2. All ten iteration-1 defects are resolved; none is unresolved or stagnant.
3. AC-CBR-009's overwrite mechanism is real on the pre-repair code, so the criterion can genuinely
   go red.
4. spec.md §A's honest-scope claim is accurate and not over-corrected: all four claimed effects are
   true against `mcp_convergence.go` and `mcp_codex.go`.
5. No AC mandates implementation shape; every REQ has AC coverage; no orphan AC.
6. The control matrix discriminates over-blocking and under-blocking, with one named residual (F1).
7. The RED list and the preservation-pin exclusion list are both correctly populated.
8. A mutant is writable against AC-CBR-009; none is writable against AC-CBR-003's own requirement.
9. Aggregate 0.93, up 0.12 from iteration 1's 0.81.

### Evidence

- `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t551`;
  `git rev-parse --short HEAD` → `3ac58b5a1`; `git branch --show-current` → `WT-audit-fail-open`.
- `stat -f '%N %m'` on the four artifacts + the supporting report, at open and at close — five
  identical pairs (table above).
- `moai spec lint --strict .moai/specs/SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001/spec.md` →
  `✓ No findings — all SPEC documents are valid`, `EXIT=0`.
- `/usr/bin/grep -o 'REQ-CBR-[0-9]*' spec.md | sort -u` → 9 sequential ids.
- `/usr/bin/grep -o '^### AC-CBR-[0-9]*' acceptance.md | sort -u` → 9 sequential ids.
- `/usr/bin/grep -n 'Covers REQ' acceptance.md` → 9 lines (`:52 :61 :72 :94 :104 :113 :123 :135
  :149`), mapping all nine REQs.
- `/usr/bin/sed -n '1125,1139p' internal/cli/mcp_codex.go` → the unconditional reassignment shown
  above (`:1134`, `:1137`).
- `/usr/bin/sed -n '1240,1270p' internal/cli/mcp_codex.go` → `bestCodexReviewText` at `:1252-1257`,
  `if review != ""` at `:1253`.
- `/usr/bin/sed -n '780,830p' internal/cli/mcp_codex.go` → guard `if reviewText == ""` at `:817`,
  sole synthesizer call at `:820`.
- `/usr/bin/grep -n -A12 'func allPass' internal/cli/mcp_convergence.go` → `:327-334`.
- `/usr/bin/grep -n -A12 'func collectFailOpen' internal/cli/mcp_convergence.go` → `:382-393`,
  appends on `VerdictInconclusive`.
- `/usr/bin/sed -n '175,205p' internal/cli/mcp_convergence.go` → `switch {` at `:185`, `case
  allPass(required):` at `:190`, `default:` at `:194`, `claudeVerdictOrDefault(verdicts,
  overallVerdictPass)` at `:199`, closing `}` at `:200`.
- `/usr/bin/grep -n 'applyGateUnmet' internal/cli/*.go` → `:1574`, `:1596`, `:1602`, `:1608`.
- `/usr/bin/grep -rn "synthesizeReviewOutput" internal/` → one production caller
  (`mcp_codex.go:820`), definition at `:1435`, every other hit a `_test.go`.
- `/usr/bin/sed -n '110,130p' internal/cli/codex_review_rpc_test.go` → function `114-126`,
  `"": "pass"` at `:119`.
- `/usr/bin/grep -n 'func codexFindingsOf' internal/cli/mcp_codex.go` → `:1473` (F5).
- `/bin/cat` of both probe files → the four- and five-row tables reproduced verbatim in
  spec.md:54-59 and cited by AC-CBR-008.
- Spot-checks of the supporting report: `:1848` names `#1632 axis 4`; `codex_task.go:71` is
  `codexTaskCallerEndedMessage`; `mcp_glm.go:342` is `func parseGLMReview`; `:276`/`:1446`
  `SynthesisNote`; `:291` `GateUnmet`. All correct.

### Baseline-attribution

Every measurement above was taken in this run, in `.claude/worktrees/t551`, against
`git rev-parse --short HEAD` = `3ac58b5a1`, confirmed unchanged at audit close. All greps used
`/usr/bin/grep` explicitly, never the shell's `grep` (a ugrep wrapper that can skip files silently).
No figure is carried over from iteration 1: every citation iteration 1 recorded as sound was
re-measured here rather than inherited, and iteration 1's own numbers appear only as the declared
delta baseline (0.81), quoted from this same file's iteration-1 header.

### Gaps — explicitly NOT observed

- **No test was executed.** `go test ./internal/cli/...` was NOT run in this audit. Every RED/GREEN
  claim about the pre-repair tree is a **static trace** through the code read verbatim, not an
  observed test result. The two probe files are the author's measurements, re-read but not
  re-executed.
- **The mutants in F1 and F4 were not compiled.** They are traced by hand through the same code
  paths, not built and run. A compiled counter-example would be stronger evidence.
- **No live `codex` binary** was exercised, and no field incidence of a whitespace-only body was
  established — unchanged from spec.md §E, which states this honestly.
- **`mcp_convergence.go` behavior was not executed**, only traced — the same gap spec.md:240-247
  now records rather than papering over.
- **No cross-backend second opinion.** `audit_multi` / `codex_audit` / `glm_audit` were not invoked;
  this is a single-backend Claude-anchor verdict.
- **Iteration 1's `:1395-1415` and `:1409-1415` range citations** were not re-litigated. Both
  bracket their referent with one blank line of padding at each end. Iteration 1 recorded them as
  sound and I did not reopen them, though they are looser than the standard D9 applied.
- **`plan.md` §E's demoted grep checklist item** was read but its future mechanization was not
  evaluated — it is explicitly deferred to run-phase by plan.md:177-179.

### Residual-risk

- **The static trace could be wrong where a test would not be.** Every conclusion about what the
  pre-repair tree does — the four RED-list judgments, the two mutant analyses, the four
  honest-effect verifications — rests on reading Go, not running it. A behavior that depends on a
  path I did not read (an early return, a caller that pre-filters) would invalidate the
  corresponding claim silently.
- **F1's mutant may be judged implausible and left unfixed.** The plan's §B.2 constrains the
  implementation to one shared `strings.TrimSpace(s) == ""` discriminator, which puts first-wins out
  of reach in practice. If run-phase follows §B.2, F1 never materializes — but §B.2 is a
  recommendation and plan.md:92 explicitly leaves shape to run-phase, so the AC is the only guard.
- **F2's hazard is realized only by a careless test.** AC-CBR-003's text already says *"the agent
  text is selected"*; an implementer who asserts it gets a genuine RED. The risk is that the §D
  RED-list line reads as a blanket guarantee and no one checks which half carries it.
- **PASS at iteration 2 of 2 means no further audit is scheduled.** F1 and the five optional
  findings will be carried into run-phase on the orchestrator's discretion rather than re-verified
  by an auditor. If F1 is routed and the fix is written loosely, nothing in this loop catches it.
- **A repair pass that produced ten correct fixes still produced F6** — a line citation that went
  stale inside the record of the repair itself. Line coordinates in these artifacts decay whenever
  the file above them changes; other coordinates inside `progress.md` that I did not re-check
  line-by-line may carry the same decay.

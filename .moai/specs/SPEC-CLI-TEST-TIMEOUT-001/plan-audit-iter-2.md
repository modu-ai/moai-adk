# Plan-Audit Report — SPEC-CLI-TEST-TIMEOUT-001 (Iteration 2 — delta re-audit)

Auditor: plan-auditor (independent). Scope per Retry Loop Contract: the iter-1 defect delta
(D1-D5) + regression check, not a from-scratch re-audit. Tree: worktree `t1253`, HEAD
`b4f798dccc8b2cf3951edea5f62e2b34740f633c` (unchanged baseline pin). Tier M — threshold 0.80,
Tier M iteration ceiling (2) reached at this iteration.

## Verdict

**PASS-WITH-DEBT — 0.88** (up from iter-1 FAIL 0.83; no score regression, no STOP signal).
All four must-pass criteria PASS. One blocking-class debt (D1′) is discharged by a 2-line
plan.md edit that MUST land before the run-phase delegation is composed; see Defects.

## Claim

All five iter-1 defects were fixed and verified against the files (not the claims). The repair
introduced one new blocking-class internal contradiction — plan.md §A/§D still declare a
two-file write scope while M1 now orders a `scripts/ci-mirror/lib/go.sh` edit — plus five
minor stale-enumeration/trace nits. Coverage is now complete over every go-test-carrying
local surface.

## Evidence

All commands run in this audit, this tree, HEAD `b4f798dcc`:

1. **D1 RESOLVED** — `scripts/ci-mirror/lib/go.sh:25` still reads `go test -race -count=1 -short ./... || exit 2` (correct at plan phase: the flag is run-phase work). `spec.md:117` REQ-TIMEOUT-010: "The `scripts/ci-mirror/lib/go.sh` test step (reached via `make ci-local`) shall invoke `go test` with the explicit flag `-timeout 60m` (derivation D1; the `-short`-mode unmeasured disclosure of §C.3 applies identically)" — surface, value, and derivation all match my iter-1 finding. D1 derivation header (`spec.md:78`) now names `ci-local` (via `go.sh:25`). §A (`spec.md:34-39`) reworded: closure claim scoped to "every COVERED row" of §B.3 with the iter-1 finding disclosed in the document itself. `plan.md:64` (M1) carries the go.sh edit with the same D1 + `-short` disclosure. `acceptance.md` AC-001 now inspects 5 COVERED surfaces with mutant yield "4 of 5"; AC-005 and plan E2 include `scripts/ci-mirror/lib/go.sh` in the changed set.
2. **D2 RESOLVED** — `spec.md:120-140` §B.3: 13-row inventory. I re-ran AC-009's own baseline discovery commands verbatim: `grep -n 'go test' Makefile` → lines 39, 48, 52, 59, 67, 104, 107, 110, 180, 184, 194; `grep -rn 'go test' scripts/ci-mirror/ scripts/ac-baseline/` → `go.sh:25` (+ its L24 log line) and `check-staged.sh:26` (`rust.sh` hits are `cargo test` — out of scope). Every go-test line maps to exactly one row: rows 1-5 COVERED, row 6 `coverage` TRANSITIVE (verified: its recipe is `go tool cover -html=coverage.out`, `Makefile:113` — no go test of its own), rows 7-12 EXCLUDED with rationales, row 13 EXTERNAL-EXPLICIT. No unlisted surface remains.
3. **D3 RESOLVED (with one wrinkle)** — `acceptance.md` §D.2: AC-004 → REQ-TIMEOUT-001/002/003/004/010 (defensible: the M3 probe demonstrates closure of the package dominating those REQ surfaces); AC-003's stray §B.1 ref removed (now REQ-TIMEOUT-006 only). Wrinkle: AC-007 now maps to REQ-DOC-011 — but REQ-DOC-011 owns the §B.3 inventory while AC-007 tests §C rejections; the trace asserts a relation the criterion does not test (D3′ below).
4. **D4 RESOLVED (with one gap split out)** — `spec.md:68-74` §B carries the 3-clause sanctioned definition ((a) help-listed Makefile target, (b) CLAUDE.local.md go-test command line, (c) target-invoked script body). Gap: clause (b) makes the CLAUDE.local.md go-test lines sanctioned, and the definition's own MUST says sanctioned surfaces are COVERED **or appear as an explicitly-excluded §B.3 row with a rationale** — but §B.3 contains no CLAUDE.local.md rows at all (D2′ below).
5. **D5 RESOLVED** — `spec.md:166-172` §D third bullet assigns CLAUDE.local.md §13 (line ~532) to t1219 with the outside-sanctioned-definition reasoning; `plan.md:74` (M2) carries the do-not-touch note.
6. **Regression batch** — `moai spec lint SPEC-CLI-TEST-TIMEOUT-001` → `✓ No findings — all SPEC documents are valid`, exit 0. REQ numbering: 001-011 with no gaps or duplicates (TIMEOUT 001-006+010, DOC 007+011, SCOPE 008, COORD 009); both new REQs are GEARS-conformant ("shall" ubiquitous). Tier budgets: 11 REQ ≤ 16, 9 AC ≤ 16. Frontmatter: all 12 canonical fields + `tier: M` unchanged. Traceability: all 11 REQs appear in §D.2; all 9 ACs anchored.
7. **New defects found by this re-audit** — `plan.md:13` "Two files: `Makefile` and `CLAUDE.local.md`" and `plan.md:39` §D "Writes only … plus the two target files `Makefile` and `CLAUDE.local.md`" both contradict `plan.md:58/64` (M1's go.sh edit), `plan.md:49` (E2 four-path changed set), and `acceptance.md` AC-005. Further stale enumerations: `spec.md:199` §G "Summary: 8 criteria" (9 now exist); `spec.md:158-159` §D t1252 bullet's "(Makefile + CLAUDE.local.md only)" parenthetical; §B.3 row 6 cites "coverage (L106)" — the coverage target is `Makefile:112` (L106 is the `test-verbose` header).

## Baseline-attribution

Every check above was executed in this audit run, against this tree, at HEAD
`b4f798dccc8b2cf3951edea5f62e2b34740f633c` (identical to the iter-1 pin — the repair touched
only the four SPEC artifacts, no target files, consistent with plan-phase discipline).

## Dimension Scores (deltas from iter-1)

| Dimension | iter-2 | iter-1 | Basis |
|-----------|--------|--------|-------|
| Clarity (MP-1) | 0.85 | 0.95 | Sanctioned definition landed and derivations unchanged; dinged for the plan.md §A/§D vs M1 write-scope contradiction (D1′) — a clarity failure in the delegation-facing document. |
| Completeness (MP-2) | 0.85 | 0.60 | Coverage now complete over Makefile + scripts + CI (§B.3, independently re-grepped) + CLAUDE.local.md (§B.2/§D). Remaining: §B.3 lacks CLAUDE.local.md rows its own definition-MUST demands (D2′). |
| Testability (MP-3) | 1.00 | 1.00 | AC-009 added: binary, mutant-specified, discovery commands inlined; all other ACs unchanged and still binary. |
| Traceability (MP-4) | 0.85 | 0.90 | Complete and REQ-anchored (AC-004 improved, AC-003 cleaned); dinged for AC-007 → REQ-DOC-011 being a semantically false trace (D3′). |

Harmonic mean = 4 / (1/0.85 + 1/0.85 + 1/1.00 + 1/0.85) = **0.88**.

## Gaps (what this re-audit did NOT observe)

- Run-phase execution of any AC (this is a plan-phase audit; AC green paths are unexecuted by design).
- A `-short`-mode duration measurement for `internal/cli` — still unmeasured; rows 3 and 5 correctly carry the same disclosure.
- t1252/t1219 landed state — coordination premises verified for presence and coherence only.

## Residual-risk

- If D1′ (plan.md write-scope contradiction) is not fixed before delegation, a manager-develop obeying §D verbatim would skip the go.sh edit and AC-001/AC-005 would fail at E1 — a caught failure, but a wasted delegation round. The fix is 2 lines in plan.md; discharge it before Kickoff.
- The 60m projection still rests on the single-sample 3.62x amplification (unchanged from iter-1; disclosed, with REQ-TIMEOUT-006 as the trigger). Operational residual, not a defect.

## Defects Found

**D1′. (BLOCKING — debt attached to this verdict)** — `plan.md:13` and `plan.md:39`.
§A says "Two files: `Makefile` and `CLAUDE.local.md`" and §D's write-scope constraint says "Writes only … plus the two target files `Makefile` and `CLAUDE.local.md`" — while M1 (`plan.md:58/64`) orders the `scripts/ci-mirror/lib/go.sh` edit and E2 (`plan.md:49`)/AC-005 expect go.sh in the changed set. The constraint section most likely to be copied verbatim into the run-phase delegation forbids the write M1 mandates. Exact "requirement demands a step while the plan instructs the forbidden side" shape (verification-completeness §3 cross-layer sweep miss).
**Required fix**: `plan.md` §A → "Three files: `Makefile`, `scripts/ci-mirror/lib/go.sh`, `CLAUDE.local.md`"; §D first bullet → "… plus the three target files …". Discharge before the run-phase delegation is composed; verifiable by `grep -n 'Two files\|two target files' plan.md` returning empty.

**D2′. (SHOULD-FIX)** — `spec.md` §B.3 + §B definition (lines 68-74, 120-140).
Clause (b) of the sanctioned definition makes CLAUDE.local.md go-test lines sanctioned, and both the definition's MUST and REQ-DOC-011 ("complete inventory (§B.3) of every go-test-carrying local surface") demand §B.3 rows for them — the table has none (the §4/§6 COVERED recipes live in REQ-TIMEOUT-005/AC-002; the §6-full-suite/§13 exclusions live in §D/t1219).
**Required fix**: add rows 14-16 to §B.3 (CLAUDE.local.md L265/L394 COVERED 30m via REQ-TIMEOUT-005; L396/397 and §13 L532 EXCLUDED → t1219), or scope REQ-DOC-011/AC-009's "every" to the table's discovery set. File: spec.md (and acceptance.md AC-009 discovery commands if the row route is taken).

**D3′. (MINOR)** — `acceptance.md:111` (§D.2). AC-007 → REQ-DOC-011 asserts a trace the criterion does not test (REQ-DOC-011 owns §B.3; AC-007 tests §C rejections). **Fix**: restore the §C section anchor for AC-007 or add a rejections-owning REQ. File: acceptance.md.

**D4′. (MINOR)** — `spec.md:133` (§B.3 row 6). "Makefile `coverage` (L106)" — the coverage target is `Makefile:112`; L106 is the `test-verbose` header. TRANSITIVE classification itself is correct (recipe is `go tool cover`, `Makefile:113`). **Fix**: L106 → L112. File: spec.md.

**D5′. (MINOR)** — `acceptance.md:64-67` (AC-006). REQ-COORD-009 (`spec.md:116`) now carries a third element (§13 ownership) but AC-006's "Then" clause still checks "both premises" only. **Fix**: extend AC-006's Then clause to name the §13 ownership bullet. File: acceptance.md.

**D6′. (MINOR)** — `spec.md:199` (§G) "Summary: 8 criteria" and `spec.md:158-159` (§D t1252 bullet) "(Makefile + CLAUDE.local.md only)" — both stale after the repair added AC-009 and the go.sh surface. **Fix**: "8" → "9"; parenthetical → "(Makefile, scripts/ci-mirror/lib/go.sh, CLAUDE.local.md — all disjoint from t1252's `internal/cli/main_test.go`)". File: spec.md.

## Regression Check (Iteration 2)

Iter-1 defects:
- D1 (BLOCKING — ci-local/ci-mirror uncovered): **RESOLVED** — REQ-TIMEOUT-010 + D1 header + §A rewording + M1 + AC-001/AC-005, all verified (Evidence 1).
- D2 (SHOULD-FIX — un-inventoried surfaces): **RESOLVED** — §B.3 rows 6-12; independent re-grep matches exactly (Evidence 2).
- D3 (MINOR — section-only anchors): **PARTIALLY RESOLVED** — AC-004 now REQ-anchored, AC-003 cleaned; AC-007's replacement anchor is a false trace (carried forward as D3′).
- D4 (MINOR — "sanctioned" undefined): **RESOLVED** — §B 3-clause definition; the new clause-(b)/§B.3-row gap is split out as D2′.
- D5 (MINOR — §13 ownership unstated): **RESOLVED** — §D bullet + plan M2 note; AC-006's unchecked third element carried as D5′.

No iter-1 defect regressed. No previously-PASS dimension regressed except the deliberate Clarity/Traceability deductions above, both traceable to repair-introduced nits, not to iter-1 material.

## Recommendation

PASS-WITH-DEBT. The Tier M iteration ceiling (2) is reached; per the Retry Loop Contract the
orchestrator now chooses between: (1) **discharge the debt** — land D1′ (2-line plan.md edit)
and, ideally, D2′-D6′ in the same pass, then proceed to run-phase without an iter-3 audit
(delta is grep-verifiable: the two grep commands in D1′ plus `grep -c '### AC-'` = 9);
(2) request iter-3 via explicit user override; or (3) treat the minors as documented debt and
proceed — acceptable only if D1′ is fixed first, since D1′ alone can cost a delegation round.
D2′ is strongly recommended before sync-phase: REQ-DOC-011 is a normative requirement whose
own current text the document does not satisfy.

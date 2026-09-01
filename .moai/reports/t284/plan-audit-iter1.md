# SPEC Review Report: SPEC-AUDIT-PARTICIPANT-COUNT-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.75** (Tier M PASS threshold 0.80)

Reasoning context ignored per M1 Context Isolation. Audited from the artifacts and the
subject code only. Tree: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t284`,
HEAD `64bba61aa`, branch `WT-audit-participant-count`.

Measurement evidence for the findings below: `.moai/reports/t284/participant-count-probe.log`
(a throwaway probe that implements REQ-APC-002's participant definition verbatim and runs it
over every existing `DisagreementFlag`-asserting case; the probe file was removed after the run).

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-APC-001..005`, sequential, no gaps, no
  duplicates, consistent 3-digit padding (`grep -o 'REQ-APC-[0-9]*' spec.md | sort -u`).
- **[PASS] MP-2 GEARS format compliance** — judged against the `REQ-XXX` requirement layer in
  `spec.md` §B. All five match a GEARS pattern: 001/002 ubiquitous (`The convergence engine
  shall …`), 003 state-driven + unwanted (`While participant_count is less than 2, … shall
  emit … and shall not emit it as false`), 004 state-driven, 005 unwanted (`The undetermined
  state shall not alter …`). The `Given/When/Then` bodies in `acceptance.md` are the
  verification layer and were graded under Group 4, not here.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct
  types (`spec.md:1-15`): `id`, `title`, `version` (quoted `"0.1.0"`), `status: draft`,
  `created`/`updated` (ISO `2026-09-01`), `author`, `priority: P1`, `phase`, `module`,
  `lifecycle: spec-anchored`, `tags` (comma-separated string), plus `tier: M`. No rejected
  snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`) present.
- **[N/A] MP-4 language neutrality** — single-language SPEC scoped to `internal/cli` (Go).
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — two referenced SPECs resolve and both read
  `status: completed`, neither `retired`/`superseded`/`archived`:
  `SPEC-CODEX-VERDICT-SYNTH-001`, `SPEC-AUDIT-MULTI-MODEL-001`. No BLOCKING finding.
- **[N/A] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md plan.md acceptance.md`
  → `0 0 0`. Auto-PASS per D8-4.
- **[N/A] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md acceptance.md
  spec.md` → no match (rc=1); `research.md` does not exist (Tier M does not require it).

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | Requirements are individually unambiguous and the exclusions are unusually disciplined. Deducted for D1/D2: `acceptance.md` AC-APC-005 and `plan.md` §E M2 both assert an invariant that the measured data contradicts, which leaves the run-phase owner with two conflicting instructions. |
| Completeness | 0.75 | 0.75 | All required sections present (HISTORY, §A, §B, §C, §D, §E with `### Out of Scope — <topic>` H3s and specific bullets). Deducted for D4 (the call-site inventory in `plan.md` §B omits an entire test file) and for the absent requirement governing the intra-backend disagreement source that `spec.md` §A.1 itself enumerates as pass 3. |
| Testability | 0.70 | 0.75→0.50 boundary | Criteria are binary and explicitly anti-vacuous — AC-APC-002 names the vacuous `!= nil && *ptr` idiom and forbids it; AC-APC-003 requires member-presence assertion. Deducted hard for D1 (AC-APC-005 is unsatisfiable alongside AC-APC-002) and D3 (AC-APC-004's leading clause already passes at HEAD). |
| Traceability | 0.80 | 0.75→1.0 | Every REQ has ≥1 AC and every AC traces to an existing REQ; the §D mutation-witness record marks non-witnesses rather than padding the list. Deducted because AC-APC-005 traces to REQ-APC-004 while contradicting REQ-APC-003. |

Aggregate (arithmetic mean): (0.75 + 0.75 + 0.70 + 0.80) / 4 = **0.75** < 0.80.

## Defects Found

**D1. `AC-APC-005` contradicts `REQ-APC-003` + `AC-APC-002` on a named case** —
`acceptance.md` AC-APC-005 — Severity: **critical** — Class: **blocking**

AC-APC-005 names `RequiredFailWithInconclusive` as one of "the four `false` cases" that must
keep a non-nil `false`. Measured, that case's input is
`pbvReq(claude,"fail")` + `pbvReq(codex, inconclusive)` (`internal/cli/mcp_convergence_test.go:101-113`),
which under REQ-APC-002's own definition (gate ≠ off AND verdict ∈ {pass,fail}) yields
**participant_count = 1**, not 2:

```
RequiredFailWithInconclusive     participants=1  currentFlag=false  overall=fail
```

REQ-APC-003 and AC-APC-002 require nil for participant_count ∈ {0,1}. AC-APC-005 requires
non-nil `false` for the same input. Both cannot hold. This is not a wording slip — it means the
implementer cannot make the criteria set green, and whichever one they satisfy silently
overrides the other.

**Required fix**: recompute participant_count for all seven cases AC-APC-005 enumerates and
re-partition them. Measured counts are in the probe log; `RequiredFailWithInconclusive` moves
out of the "stays `false`" set and into the "becomes null" set. `plan.md` §E M2's sentence
"every currently-asserted `false` in a case with 2+ participants stays `false`" is *formally*
true only because of its 2+ qualifier — but it is then paired with AC-APC-005's flat
enumeration, which drops the qualifier. Fix both surfaces together.

---

**D2. REQ-APC-003 suppresses a genuinely-true intra-backend disagreement, and the case is
absent from every inventory** — `internal/cli/codex_verdict_divergence_test.go:55-79` (assertion
at `:70`); `internal/cli/mcp_convergence.go:188-197` — Severity: **critical** — Class: **blocking**

`spec.md` §A.1 enumerates three disagreement-deriving passes and names the third as
"intra-backend synthesis note" (`collectSynthesisNotes` at `mcp_convergence.go:194`, flag set at
`:195-197`). That pass is the landed REQ-CVS-003 behaviour of the completed predecessor
SPEC-CODEX-VERDICT-SYNTH-001. It fires **inside a single backend**, so it can be true with only
one participant. Measured:

```
SurfacesSignalDivergence         participants=1  currentFlag=true  overall=pass
```

Under REQ-APC-003 that `true` becomes `null`. The engine currently reports "a backend's own
signals contradicted each other"; after this SPEC it would report "not enough participants to
say", discarding a disagreement it actually observed. `TestConverge_SurfacesSignalDivergence_WithoutBlocking`
(`codex_verdict_divergence_test.go:70`) is the regression guard for that behaviour, and it goes
red — not from the mechanical `*bool` narrowing, but semantically.

Two aggravating omissions: this case appears in **none** of `plan.md` §B (call-site inventory),
`plan.md` §E M2 (site count), `acceptance.md` AC-APC-005 (which lists only three `true` cases),
or the §D mutation-witness record. And `spec.md` §E declares "Any change to `collectSynthesisNotes`
behaviour beyond leaving it untouched" out of scope — REQ-APC-003 does not change the function,
but it does change its observable effect, so the exclusion as written does not cover what the
SPEC actually does.

**Required fix**: decide and record the interaction explicitly. Either (a) REQ-APC-003 carves out
the intra-backend pass — an intra-backend divergence is self-evidencing and does not need two
participants — with a matching criterion; or (b) the SPEC accepts the information loss, states it
in §E as a deliberate consequence rather than an untouched surface, and AC-APC-005 is extended to
cover the changed case. Silence is the one option that is not available.

---

**D3. `spec.md` §A.3's central premise claim is false as measured, and AC-APC-004's leading
clause is vacuous** — `spec.md` §A.3; `acceptance.md` AC-APC-004 — Severity: **major** —
Class: **blocking**

§A.3 corrects the card's "same bytes hide two facts" wording and then states: "that is exactly
true only for case 3 vs case 2 — those two `ConvergenceResult` JSON documents are byte-identical".
Measured against the very log §A.2 cites (`.moai/state/verify/t284/premise-probe.log`), the two
documents differ substantially: case 2 carries three `per_backend_verdicts` entries, case 3
carries one.

```
case2==case3 byte-identical? false (len 447 vs 235)
```

What is actually identical is every field *except* `per_backend_verdicts` — `overall_verdict`,
`disagreement_flag`, `residual_risk_note`, `fail_open_backends`. The section devoted to premise
precision commits a stronger version of the imprecision it is correcting.

The consequence reaches a criterion: AC-APC-004 requires "the two documents are not
byte-identical". That clause is satisfied at HEAD, before any change — it asserts nothing. The
conjoined second clause (differ in both `participant_count` and `disagreement_flag`) is not
vacuous, so AC-APC-004 as a whole is not worthless; but its stated rationale ("the one case where
the current engine emits identical bytes") rests on a false premise.

**Required fix**: restate §A.3 as *the derived fields are identical* (naming the four), and drop
or rewrite AC-APC-004's "not byte-identical" clause so the criterion asserts only the non-vacuous
part.

---

**D4. `plan.md` M2's "11 test sites" undercounts; measured 12 across 3 files** —
`plan.md` §B (table row "DisagreementFlag read/write sites in tests") and §E M2 — Severity:
**major** — Class: **blocking**

Measured (`grep -rn DisagreementFlag --include="*.go"` over the worktree):

| File | Sites |
|---|---|
| `internal/cli/mcp_convergence_test.go` | 8 (`:75, :94, :113, :131, :163, :180, :201, :457`) |
| `internal/cli/multi_review_gate_test.go` | 3 (`:65, :80, :95`) |
| `internal/cli/codex_verdict_divergence_test.go` | **1** (`:70`) — omitted from §B |
| **total** | **12** |

§B's per-file numbers (8 and 3) are individually correct; the table simply does not know the third
file exists. The compile break is self-correcting for the *mechanical* repair, but the omission is
not harmless: the missing file is exactly the one carrying D2's semantic collision, so the site
inventory and the semantic gap have the same root.

**Required fix**: correct §B to three files / 12 sites and M2 to 12, and route the twelfth site
through D2's decision rather than the mechanical M2 pass.

---

**D5. `spec.md` §A.4 cites the header block as "lines 1-38"; it runs 1-39** —
`spec.md` §A.4 vs `internal/cli/mcp_convergence.go:39-40` — Severity: **minor** —
Class: **optional**

`// @MX:SPEC: SPEC-AUDIT-MULTI-MODEL-001` is line 39; `package cli` is line 40. All three
constraints §A.4 quotes (C2 at `:25`, C3 at `:16`, C5 at `:36`) are inside the cited span, so no
claim is affected. Off-by-one only.

**Required fix**: `lines 1-39`, or drop the line range.

---

**D6. `spec.md` §A.1 pass-3 coordinate stops one line before the branch it describes** —
`spec.md` §A.1 table row 3 (`mcp_convergence.go:187-194`) — Severity: **minor** —
Class: **optional**

The cited span covers the comment (`:188-193`) and the `collectSynthesisNotes(verdicts)` call
(`:194`), which matches the table's "What it reads" column. But the flag assignment is at
`:195-197`, outside the span — and that is precisely the branch REQ-APC-003 collides with (D2).
Widening the coordinate would have made the collision visible while §A.1 was being written.

**Required fix**: cite `:188-197`.

---

## Verified-clean (probed, no defect found)

Recorded so their absence from the defect list is not silence:

- **Every other cited coordinate resolves and says what the SPEC claims.** `converge` at
  `mcp_convergence.go:140` ✓; required-split at `:169-170` ✓; advisory-conflict loop at
  `:173-185` ✓; DQ-2 refusal literal at `:526-532` ✓; gate-off `continue` at `:571-573` ✓;
  `HandleMultiReviewGate` field reads at `multi_review_gate.go:83` (`OverallVerdict`) and `:97`
  (`ResidualRiskNote`) ✓ — and those are the file's *only* two result-field reads, confirmed by
  grep; `blockReason` at `:96` ✓; `loadConvergenceResult` at `:110` ✓; `mcp_server.go:405-425`
  declares no output schema for `audit_multi`, and the two `WithOutputSchema[ReviewOutput]()`
  sites at `:259` / `:400` belong to `codex_audit` / `glm_audit` ✓.
- **§A.1's "0 hits" grep** — `grep -ci participant` over the three surfaces → `0 0 0` ✓.
- **§A.2 matches the cited log exactly** — all three rows' `overall_verdict`,
  `disagreement_flag`, and `fail_open_backends` reproduce `.moai/state/verify/t284/premise-probe.log`
  verbatim ✓. (The §A.3 *inference* from that log is D3; the table itself is sound.)
- **Point 2 — no missed consumer.** Repo-wide `grep -rl disagreement_flag` returns no JSON
  schema (the only two `*.schema.json` files are `module_tree.schema.json` and its template
  mirror), no shell hook, and no non-Go parser. Every non-Go hit is prose (docs-site, skills,
  SPEC/report archives). The MCP path is `convergenceToolResult` → `mcp.NewToolResultJSON(r)`
  (`mcp_audit_multi.go:142-148`), which emits both TextContent JSON and StructuredContent from
  the same struct — a nil `*bool` renders `null` in both, and `audit_multi` declares no schema to
  validate against. The state-file round-trip is plain `json.Unmarshal` into the struct
  (`multi_review_gate.go:119-121`), so a recorded `false` decodes to a non-nil pointer as
  `plan.md` §D claims. **The shape decision is safe for every consumer I could find.**
- **Point 3 — AC-APC-002 / AC-APC-003 both kill the representative mutant, and neither is
  vacuous.** A single mutation *does* fail both (the representative mutant fails both) — that is
  redundancy at two layers, not a defect. The independence `acceptance.md` §D actually claims is
  narrower and correct: `omitempty` on the pointer satisfies AC-APC-002 (nil) and fails
  AC-APC-003 (member absent). AC-APC-002 pre-empts its own vacuous form by naming the
  `!= nil && *ptr` idiom and rejecting it; AC-APC-003 requires the presence assertion. Both hold.
- **Point 5 — `plan.md` §E's claim about `TestConverge_NoRequiredBackends_VacuousPass` is
  correct.** Two advisory passes → measured `participants=2`, so it keeps a non-null `false` as
  the plan states ✓.
- **`plan.md` §G occurrence counts are exact.** All 12 rows re-measured with `grep -c`; every
  count matches, including `docs-site/content/ko/advanced/autonomous-loops.md` = 0, which
  substantiates `spec.md` §E's recorded pre-existing 4-locale parity gap ✓.
- **Point 6 — scope discipline holds.** Nothing in the artifacts exceeds the three card items.
  REQ-APC-002 (participant definition), REQ-APC-004 (2+ unchanged), and REQ-APC-005 (C3
  preservation) are necessary consequences of items 1-2 rather than additions; AC-APC-007
  (old-state-file decode) is a consequence of the chosen shape, not new scope. `spec.md` §E is
  thorough and correctly forbids the obvious creep (minimum-participant policy, new enum, new
  block category). The `.moai/reports/t229/succession.md` deferral record confirms the three items
  and the explicit handoff of the representative mutant to this card's AC design ✓.

## Recommendation

FAIL at 0.75 against the Tier M threshold of 0.80. Four blocking defects, two of which (D1, D2)
make the criteria set unsatisfiable as written. Fix in this order:

1. **D2 first** — it is the only one requiring a *decision* rather than a correction, and D1's
   re-partition depends on it. Decide whether the intra-backend divergence pass is carved out of
   REQ-APC-003 or whether the information loss is accepted, and record it in `spec.md` §B/§E.
2. **D1** — re-partition AC-APC-005's seven cases by measured participant_count (probe log has
   all eight). `RequiredFailWithInconclusive` moves; `SurfacesSignalDivergence` joins the list
   with whatever D2 decides.
3. **D4** — correct §B to 3 files / 12 sites and M2 to 12.
4. **D3** — restate §A.3 as "the four derived fields are identical" and strip AC-APC-004's
   vacuous leading clause.
5. **D5, D6** — coordinate corrections, optional; fold in while editing §A.

The artifacts are otherwise strong: the shape decision in `plan.md` §D is well argued with five
genuinely-considered rejected alternatives, the consumer analysis behind it is correct as
independently measured, the exclusions are disciplined, and AC-APC-002/003 are written to resist
exactly the vacuity that criteria of this kind usually fall into. The failure is concentrated in
one place — the participant-count arithmetic was never actually run against the existing test
corpus, and both D1 and D2 fall out of that single omission.

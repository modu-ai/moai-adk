# t469 Plan Audit — SPEC-HOOK-WIRING-DRIFT-001 v0.4.0 amendment plan (plan.md §I)

Auditor: plan-auditor (independent) · Date: 2026-09-03 · Iteration: 1/1 (Tier-differentiated full audit)
Tree: worktree `.claude/worktrees/t469`, branch `WT-achwd-strip-exempt`, audited at HEAD `f0b926d71`
Plan under audit: `b27ac0320` (plan.md §I + plan-summary.md); delta `a1d7598ac..f0b926d71` is report-only
(`git diff --stat` → lane-recon.md, plan-summary.md, plan.md §I — zero rule-file or template changes), so
every re-measurement below is baseline-equivalent to the plan's pinned `@a1d7598ac`.

## Verdict: PASS-WITH-DEBT

Overall Score: **0.80** — Clarity 0.75 / Completeness 0.75 / Testability 0.80 / Traceability 0.90.

Threshold note: the dispatch scored this as a Tier S card (0.75), but the SPEC's own frontmatter
carries `tier: M` (spec.md:14), whose skip-eligibility threshold per `spec-workflow.md` § SPEC
Complexity Tier is **0.80**. The aggregate 0.80 meets both; the lead should be aware the binding
number for run-gate skip-eligibility is the SPEC's `tier:` field, not the card's.

Why PASS-WITH-DEBT and not PASS: the amendment's evidence base, design, and machine check are fully
verified (below), but two payload-correctness defects (D1, D2) should be repaired **before the run
phase transplants §I.4's exact texts into spec.md/acceptance.md** — A4's mutant note would
otherwise enshrine an imprecise mutant record, and A1 omits the amendment declaration mechanics the
frontmatter schema and this SPEC's own §G-1 prescribe. Why not FAIL: no must-pass criterion failed;
every load-bearing claim I could execute, I executed, and it held.

## Must-Pass Results

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| MP-1 | REQ number consistency | PASS | §I adds/removes no REQ; REQ-HWD-013 reworded in place (plan.md:459-472), REQ-HWD-014 untouched; spec.md REQ set REQ-001..014 unchanged |
| MP-2 | GEARS format (requirement layer) | PASS | A2's reworded REQ-HWD-013 keeps Ubiquitous GEARS shape: "Every template-managed file this SPEC edits **shall** be edited … and the two copies **shall** be identical at closure …" (plan.md:462-466). Judgment made against the REQ layer only; the Given-When-Then AC text in A4 is verification-layer format by design |
| MP-3 | YAML frontmatter validity | PASS | A1 changes only `version: "0.4.0"` (quoted semver) and `updated:` (already `2026-09-03`, a no-op — §I.5 #3 correctly flags it). Current frontmatter validates: tree-built `go run ./cmd/moai spec lint .moai/specs/SPEC-HOOK-WIRING-DRIFT-001/spec.md` → **0 error(s), 18 warning(s)** (all pre-existing `CoverageIncomplete` from the linter not resolving spec.md↔acceptance.md cross-file REQ↔AC mapping; unchanged by this plan) |
| MP-4 | Language neutrality | N/A | Single-repo template-neutrality SPEC; no language-specific tooling is selected or ranked; the A4 command names no language toolchain |
| MP-5 | D7 cross-SPEC reconciliation | PASS | `grep -oE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` over plan.md → self + `SPEC-SYNC-PARALLEL-DOCS-001` (×3). Target exists; `status: completed` (not retired/superseded/archived). No BLOCKING finding |
| MP-6 | D8 cross-platform discipline | PASS | `grep -c 'syscall' plan.md` → 0 → D8-4 auto-PASS |
| MP-7 | Clarification gate | PASS | `grep -c 'NEEDS CLARIFICATION'` → 0 in plan.md, plan-summary.md, lane-recon.md |

## Category Scores

| Dimension | Score | Basis (evidence) |
|-----------|-------|------------------|
| Clarity | 0.75 | Exceptional precision overall (exact texts, exact command, named residuals §I.5). Two accuracy blemishes: D2 (mutant-ii absorption claimed unconditionally; measured shape-dependent) and D5 (A2 wording promises normalization of all five REQ-HWD-014 classes; mechanism covers four) |
| Completeness | 0.75 | All five requested payload elements present (frontmatter bump, HISTORY row, AC body, REQ reword, Out of Scope — anchor `### Out of Scope — MX index staleness semantics` verified at spec.md:344; H3+bullet form correct). Missing: amendment declaration mechanics (D1) and the lint step unpayloaded (D4); A4's machine-readable form garbled by fence nesting (D3) |
| Testability | 0.80 | Amended criterion is binary (exit 0/1), ONE command, guard-safe; Pre-impl observation pinned to a tree SHA (`@a1d7598ac`); mutants really executed. Only defect is the mutant-note scope error (D2) — prose, not the criterion's judgment mechanics |
| Traceability | 0.90 | AC-HWD-015↔REQ-HWD-013 preserved (acceptance.md:58 header row untouched and still valid); A2 keeps the explicit AC pointer ("The acceptance-side check is AC-HWD-015", plan.md:471); A3 documents the fleet-wide deferral; no REQ renumbering. Deduction for the D5 requirement↔mechanism seam |

## Defects Found

**D1 — A1 omits the amendment declaration mechanics — plan.md:454-457 — Severity: SHOULD-FIX — Class: blocking (fix before run phase applies payloads).**
A1 instructs only `version` + `updated` + a HISTORY row. For a `completed → in-progress` amendment the
frontmatter schema (SSOT, § Optional Fields `amendment_of` + Status Transition Ownership Matrix
amendment row, which MUSTs the prior-completed record) prescribes: the self-referential
`amendment_of: SPEC-HOOK-WIRING-DRIFT-001` field, the `completed → in-progress` status transition,
and the prior completed version + SHA. This SPEC's own §G-1 (spec.md:363-365) names exactly that path
("carries `status: completed`, so the decision lands as an **amendment** to it (`completed →
in-progress` per the frontmatter schema)"), and the repo's live precedent (SPEC-AGENT-PARALLEL-OPT-001
v0.13.0 row) uses `amendment_of:` self-reference. §I.3's cited precedent ("HISTORY carries two
in-place amendments (v0.2.0, v0.3.0)") is a category error — both were **pre-completion** audit
iterations, not completed-SPEC amendments. Consequence of omission: the SPEC silently reads
`status: completed` at v0.4.0 with an unclosed amendment; `internal/spec/audit.go:362-363` returns no
drift for completed SPECs, so nothing ever flags or re-certifies it. Required fix: extend A1 with
`amendment_of` + status transition + prior-completed record (or, if the lane elects to stay
`completed` under the schema's "MAY", A1 must say so explicitly and justify it — silence is the one
wrong answer).

**D2 — Mutant (ii) claim is shape-dependent, stated as unconditional — plan.md:526-531 (A4 Mutant bullet), plan.md:419 (E5 row) — Severity: SHOULD-FIX — Class: blocking.**
A4 says a forbidden token inserted into the template copy is "absorbed by design (exit 0)". Measured,
two shapes diverge: (ii-bare) ` (SPEC-AUDIT-FAKE-001 A9)` inserted into the otherwise-shared line 275
of the template copy → **exit 0** (claim holds); (ii-prose) `Cross-reference: see (SPEC-AUDIT-FAKE-001
A9) for detail.` appended as a new line → **MISMATCH, exit 1** (the mirror check catches it itself —
stronger than claimed). E5's "7 SPEC-ID hits" figure is the real local file's own count
(`grep -cE 'SPEC-[A-Z]'` on local → 7, template → 0), i.e. the lane's mutant made the template copy
byte-identical to the local file — a wholesale copy, not an insertion absorbed by normalization. The
pair-of-criteria closure conclusion survives (in the absorbed shape AC-HWD-016 catches it — I
observed grep ≥1 on the mutant; in the prose shape the mirror check catches it directly), but the
mechanistic wording would be transplanted verbatim into acceptance.md and misrecord mutant semantics.
Required fix: split the bullet into the two shapes with their observed exits, and state the 7-hit
figure's actual construction.

**D3 — plan.md:491-535 nested-fence bug — Severity: MINOR — Class: optional.**
The A4 payload's outer fence (opened :491) is closed prematurely by the inner ```bash block's
closing fence (:508); payload tail (:509-534) renders outside the block, and the :535 fence opens an
unterminated block that swallows A5 and §I.5 as code. Fences at :461/:472 and :477/:486 are balanced;
only the A4 block is malformed. Fix: 4-backtick outer fence for the A4 payload. (A human or LLM
reader still extracts A4 correctly; a mechanical extractor truncating at :508 would not.)

**D4 — `moai spec lint` named only in prose, carried in no payload — plan.md (absent from A1-A5; plan-summary.md:135-137 Gaps) — Severity: MINOR — Class: optional.**
The evidence record punts the lint question to "the plan-audit and/or run phase"; A5 (the hand-off
step) carries only the §I.4 command re-run. I discharged the underlying risk by measurement (MP-3:
tree-built lint → 0 errors), so nothing is blocked; fix is one added line in A5.

**D5 — A2's reworded requirement over-promises normalization coverage — plan.md:464-466 vs plan.md:544-555 (§I.5 #2) — Severity: MINOR — Class: optional.**
"identical at closure after normalization of the forbidden-class tokens REQ-HWD-014 requires the
template copy to omit" claims all five REQ-HWD-014 classes; the mechanism normalizes four (no
internal date — acknowledged in §I.5 #2). A future date-class strip reproduces the original defect
one class over, with the requirement text now on the mechanism's side. Fix: scope the A2 wording to
the four regex-addressable classes, or carry the date-class exception into the amendment note.

## Lead [HARD] constraints — disposition

1. **HISTORY row content** — SATISFIED. A1's row (plan.md:457) carries the change reason
   (simultaneous-unsatisfiability argument), the evidence (observed line-275 divergence, both-blob
   pre-existence at a239cf050, 47/24 class measurement, three executed mutants), and the explicit
   no-retroactive-re-judgment clause ("t216's landing stands, no other AC re-opened").
2. **Machine-judgeable by ONE command** — SATISFIED with one wording defect (D2). Re-executed myself:
   - Exact §I.4 A4 command, verbatim, one Bash call → **no output, exit 0**.
   - Mutant (i) editorial prose on the token-bearing line (scratch fixture under
     `.moai/state/verify/t469/audit/`, same perl logic after a `chdir` prefix; tracked files never
     touched) → `MISMATCH .claude/rules/moai/core/agent-common-protocol-reference.md`, **exit 1**.
   - Mutant (ii) two shapes → bare-token: **exit 0** (as claimed); prose-wrapped: **exit 1**
     (stronger than claimed) — see D2. AC-HWD-016 compensating control verified real:
     acceptance.md:578-600 exists, is a template-side neutrality scan, and my grep on the mutant
     template copy returned ≥1 SPEC-ID hit.
   - Mutant (iii) mirror-only edit → `MISMATCH`, **exit 1**.
   - Guard compatibility: the A4 command is a single plain invocation (one `perl` process; no shell
     loop, no process substitution, no compound git) and was accepted and executed as written.
3. **Pre-existence in THIS tree** — SATISFIED, re-executed: `git merge-base --is-ancestor a239cf050
   HEAD` → rc=0; both blobs extracted to /tmp and diffed → the SAME `275c275` line pair, local
   ending `(SPEC-SYNC-PARALLEL-DOCS-001 A9).`, template ending `there.`.
4. **Git discipline** — SATISFIED: no push, no branch creation; scratch confined to
   `.moai/state/verify/t469/` + /tmp and removed afterwards; `git status --porcelain` → empty.

## Additional audit focus — results

- **Payload coverage (A1-A5)**: covers all five requested elements; nothing self-contradictory
  between payloads; gaps are D1 and D4.
- **`make build` clause removal**: NOT dangling. AC-HWD-005 exists and runs `bin/moai doctor`
  (acceptance.md:175); AC-HWD-012 exists and runs `bin/moai mx query` (acceptance.md:610). Both bind
  freshness of the features they test; the delegation mirrors the original criterion's own mutant
  note (acceptance.md:570-572).
- **Normalization over-absorption**: measured by marking every span the normalizer would remove
  across all six files (3 local + 3 template) — exactly **one** span total:
  ` (SPEC-SYNC-PARALLEL-DOCS-001 A9)` (local reference file), zero newline-swallowing, zero
  non-token deletions, no legitimate parenthetical touched. §I.5 #1's risk is theoretical at this
  scope, not actual.
- **§I.5 residuals**: #1 verified theoretical at three-file scope (above); #2 real and honestly
  acknowledged (drives D5's wording fix); #3 verified — frontmatter already reads `updated:
  2026-09-03` (spec.md:7), so A1's date instruction is a no-op, correctly flagged.
- **Residual-risk compensating control (AC-HWD-016)**: real, not asserted — the criterion exists at
  acceptance.md:578-600 with template-side scan classes matching the normalizer's four, and my
  mutant-template grep demonstrated it catching what the mirror check absorbs.
- **`moai spec lint` gap**: measured and discharged (0 errors, tree-built binary per VCI §2.2);
  payload omission remains as D4 (minor).

## Re-executed vs reasoned

Re-executed (this audit, this tree): exact A4 command (exit 0); mutants i / ii-bare / ii-prose /
iii in scratch fixture; a239cf050 blob extraction + diff + ancestry; `diff -q` ×2 + full diff ×1
(E1/E2); pairs.sh + classes.sh re-run (47 DIFF / 15 NOTEMPLATE / 24 classified — matches E4);
classes.txt head inspection; per-file SPEC-ID counts (local 7 / template 0 — explains E5's figure);
six-file normalization span inspection; tree-built `spec lint` (0 errors); `git diff --stat
a1d7598ac..f0b926d71` (report-only delta); D7 SPEC-ID extraction + target status read; D8/MP-7
greps; fixture cleanup + porcelain-clean verification.

Reasoned (not executed): the category-error classification of the v0.2.0/v0.3.0 precedent (read from
HISTORY rows); audit.go:362-363 behavior read from source comments; the drift consequence in D1
inferred from that read, not observed by running an amendment — flagged here so the lead knows which
D1 consequence is measured vs inferred.

## Recommendation

Route D1 and D2 back to manager-spec as a plan-text patch (both are edits inside plan.md §I payloads;
D3-D5 ride the same patch). After the patch, no re-audit of evidence is needed — the evidence layer is
verified; the confirming check is textual only. Do not let the run phase apply A1/A4 as drafted.

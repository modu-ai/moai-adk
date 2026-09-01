# SPEC Review Report: SPEC-CLOCAL-AUDIT-001 (card t308)

Iteration: 1/2 (Tier M ceiling: 2 per harness.plan_audit_tier_ceilings)
Verdict: **FAIL** (score below Tier-M skip threshold)
Overall Score: **0.75** (harmonic mean of 4 dimensions)
Tier-M PASS/skip threshold: 0.80 → run-phase entry BLOCKED until iteration.

Auditor basis: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t308` @ HEAD `d29b8942e5b237ef7180749dc04273d83647c745`, branch `WT-clocal-audit`, verified this run (`git rev-parse HEAD`, `git branch --show-current`). Subject copy `CLAUDE.local.md` = 663 lines (`wc -l`). Reasoning context from the SPEC author ignored per M1 Context Isolation; only artifact files read.

---

## Must-Pass Results

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| MP-1 | REQ number consistency | **PASS** | `REQ-CLOCAL-001..011` sequential, zero gaps/dups, uniform padding (`spec.md:L41–L71`). Inventory INV ids (sparse 001..135) are not subject to MP-1 (not REQ numbers). |
| MP-2 | GEARS format compliance (requirement layer only) | **PASS** | All 11 REQs match one of the five patterns incl. compound clauses: Ubiquitous (001/004/006/007/009 — canonical negative at L66), Where (002, L45), When (005/010/011, L54/L69/L72), While (008, L63). Nuances logged as D8 (MINOR), not failures: Where used as event trigger (002); REQ-003 heading labeled Unwanted while syntax is When-scoped prohibition (L48); "may verify" permission inside Ubiquitous 004 (L51). |
| MP-3 | YAML frontmatter validity | **PASS** | Canonical 12 present with correct types (`spec.md:L2–13`): id/title/version "1.0.0"/status draft/created+updated 2026-08-27 ISO/author/P2/phase `"v3.1.4"` (release label, not stage)/module/lifecycle spec-anchored/tags CSV string. Extras (`tier: "M"`, `type:`, `related_specs:`) tolerated optional/non-conflicting. No rejected snake_case aliases. |
| MP-4 | Section 22 language neutrality | **N/A (auto-pass)** | Single-document audit SPEC scoped to CLAUDE.local.md; no multi-language tooling content template-bound here. |
| MP-5 | D7 cross-SPEC reconciliation | **PASS** | Only SID reference: `related_specs: ["SPEC-STAMP-REACHABILITY-001"]`. Measured: `.moai/specs/SPEC-STAMP-REACHABILITY-001/spec.md` exists; `grep '^status:'` → `status: completed`. Not retired/superseded/archived ⇒ no reconciliation required. Card ids t294/t295/t298/t303 referenced in body are queue cards, not SPEC-IDs. |
| MP-6 | D8 cross-platform discipline | **PASS** | `grep -c "syscall" spec.md` → 0 occurrences ⇒ D8-4 auto-pass. |
| MP-7 | Clarification gate | **PASS** | `grep -c "\[NEEDS CLARIFICATION" plan.md` → 0. `research.md` absent — correct for Tier M artifact set (spec/plan/acceptance/progress measured on disk); N/A remainder per MP-4 precedent. |

No must-pass failure. Verdict turns entirely on the aggregate score below.

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.75 | minor ambiguity | Consistent boundary statement `273–337` in ALL five occurrences measured (`spec.md:L51`, `plan.md:L11`, `progress.md:L21`, `acceptance.md:L9,L33`, `claims-inventory.md:L3`) — internally coherent and factually correct vs ground truth (§4.1 heading true L273, last content L334, hr L336, §5 heading L338). BUT no content-based anchor and no mid-run re-anchor rule exist (only SHA-pinned literals; grep found zero "§4.1 heading through …"-style definitions) → executor cannot tell how to handle the zone after the audit's own edits shift it (D1). REQ-005 typo'd symbol momentarily points at a non-existent referent (D5). |
| Completeness | 0.75 | one section sparse | Structural sections all present (HISTORY/A/ B/C with four `### Out of Scope` H3+bullets, §D pointer to acceptance.md). Gap is the frozen inventory itself: five measurable claim-cluster misses vs the title "Exhaustive" (D3) — §3 Code Standards entire subsection (subject L230–238), §13 `loadGLMKey()`+`~/.mo/.env.glm` path (L521), §12 rule-pointer pair skill-authoring/agent-authoring (L499), References §20 Vercel-pricing row (L638) unanchored although `EXTERNAL-UNVERIFIABLE` was coined for precisely that row, References §27 row claims (L645), plus §11 inward pointer "§6 검증 규율(clear → 측정)" (L487) unverifiable and uninventoried. Covered examples verified present (no hole there): coverage targets INV-082, exit-137 Makefile/pkg-version INV-075/076, §9 zero-reader INV-091, §28 drain flags INV-131, `/moai feedback` wiring INV-121, no-build+embed INV-012/120. |
| Testability | 0.75 | several ACs need judgment | Strong core: AC-001/002/003/005/006/008/011 binary-observable via transcript/diff/status artifacts; closure target sums to 70. Weak: AC-004 second clause "no verification effort beyond citation checks was spent" (`acceptance.md:L40`) is process-state, not artifact-observable (D6); AC-007 "smallest text correction" semi-judgmental (hooked by INV-ref-per-hunk, tolerable); GWT scenario blocks exist for only 9 of 12 ACs (missing for 005/010/012 — terse matrix methods only). Edge-case rules D.2 provide good disambiguation (±10 citation-drift class etc.). |
| Traceability | 0.75 | one misdirected mapping | Bidirectional REQ↔AC complete for 10 pairs (`acceptance.md:L5–18`: 001↔AC011 … 011↔AC010 all resolve). Defect: AC-CLOCAL-012 maps "§28 surface naming" onto REQ-CLOCAL-005 (`acceptance.md:L18`), whose body (`spec.md:L53–54`) is known-unresolved carry-through — the obligation actually lives in the audit-basis clause (`spec.md:L37`), mapped to the wrong parent (D4). Milestone↔item map derivable via inventory group tags, but INV-021/022 appear in M7 only through the catch-all bullet (D7). |

Harmonic mean = 4 / (1/0.75 × 4) = **0.75** < 0.80 → **FAIL**.
Note on leniency bound: even granting Clarity 1.0, HM = (1.0, .75, .75, .75) → 0.80 borderline; the other three dims carry measured defects each individually justifying their band. FAIL stands without manufacturing.

## Audit-Focus List Findings (7 requested probes)

1. **Exclusion-zone precision** — The alleged "roughly lines 273–335" contradiction DOES NOT EXIST in any artifact: `grep -rn "335\|337"` over `.moai/specs/SPEC-CLOCAL-AUDIT-001/` and `.moai/reports/t308/` returns five hits, all stating `273–337`. All five agree; the range is also accurate against d29b8942e ground truth (zone heading L273 → §5 heading L338). Real defect found instead: **no content-based definition anywhere** and **no protocol for line drift caused by run-phase's own fixes above the zone** (REQ-004 is SHA-pinned, so external drift aborts via REQ-001, but internal drift after e.g. fixing INV-001 near L5 silently invalidates live-number checks). → D1 (BLOCKING).
2. **AC mechanical-checkability** — 9/12 carry executable observation verbs; AC-004 effort clause fails the "lazy executor self-report" test (nothing on disk distinguishes spent vs saved effort) → D6. AC-003 implicitly needs frozen-basis diff reading, never stated → folded into D1.
3. **Inventory coverage holes** — Checked every named probe; all six examples ARE covered (cited above). Found instead five OTHER uncovered clusters plus three inventory-truth defects → D3, D2, D5.
4. **Two-cell discipline fit** — RED-now is documented and current: `claims-inventory.md:L6` "Status at freeze: ALL pending" gives the observable pre-state for AC-006 closure; green path named (M8 close per `plan.md:L80`). Acceptable; no defect beyond D6.
5. **Fix-routing consistency** — Verified clean: REQ-009 write-scope = three paths only; REQ-008 routes cross-doc staleness to new-card recommendations, hypothesis leftovers to Open Questions; §C Out-of-Scope has four H3 bullets excluding Go/shell changes and other docs; no REQ/AC demands editing any other document or code. Historical-record vs external-unverifiable split keeps incident narratives unre-executed. PASS with no finding.
6. **Citation integrity of the SPEC set itself** — Internal refs resolve: HISTORY commit `20fab1786` verified via `git log -1 --format=%h -- CLAUDE.local.md` → `20fab1786`; 70-count story identical across `spec.md:L25`, `plan.md:L5`, `inventory:L6/L144`, `AC-CLOCAL-006`, `progress.md` (`inventory_items: 70`); measured `grep -o "^| INV-" claims-inventory.md | wc -l` → 70 exactly. Evidence-pack trio present on disk. HOWEVER the frozen inventory's own LINE ANCHORS drift systematically from ground truth (~21 rows) → D2 (BLOCKING, most ironic defect: the citation-accuracy mechanism for a citation-accuracy audit is itself stale) and D4/D5 on mappings/spelling.
7. **Open-question handling** — None orphaned: INV-033→M1, INV-056→M1, INV-057→M1, INV-101→M5, INV-110/111/112→M2 (`plan.md:L44–52,L68`), near-certains INV-001/002→M7 (`plan.md:L75`), plus Known Issues B.2–B.4 pre-documents each. PASS.

## Defects Found (structured defect-list)

D1. exclusion-zone re-anchor rule absent — spec.md:L51 (REQ-CLOCAL-004), acceptance.md:L33/L35 — Zone defined ONLY as literal lines 273–337 @ d29b8942e; the audit's own first fix above L273 shifts the zone and AC-CLOCAL-003 offers no procedure distinguishing frozen-basis coordinates from live-copy coordinates. A diligent executor is slowed; a lazy one can unknowingly edit into the protected interval or accept drift silently. Severity: major. Class: blocking. Required fix: append to REQ-CLOCAL-004 a content-based anchor ("the `### §4.1` heading through the `---` immediately preceding the `## 5.` heading") + mid-run rule (relocate by heading text whenever a fix lands above L273; log the shifted interval in progress notes); mirror in AC-CLOCAL-003 Given.

D2. frozen-inventory line anchors drifted vs d29b8942e — claims-inventory.md (LO group INV-033…INV-046 rows, INV-047 manifest boundary; REF group INV-110, INV-111, INV-059, INV-113; E group INV-102) — Mechanically re-measured examples: INV-033 anchor L118 but `handle-*.sh` true L119; INV-040 L125 vs true L126; INV-045 L131 vs true L132; INV-048 L141 exact (mixed fossilization); INV-110 "L638 §18" vs true L636 (+2); INV-111 "L642 §23" vs true L641; INV-059 "L641 §19" vs true L637 (+4); INV-102 cites L429–433 for the TemplateContext FIELD block though the struct sits L417–420. Contradicts HISTORY freeze claim (`spec.md:L25`) and the SPEC-set's own premise that authority depends on citation accuracy (`spec.md:L30`). ~21 of 70 anchors ≈ 30% misaligned. Severity: major. Class: blocking. Required fix: one CHK sweep re-deriving EVERY anchor column via `grep -n` against d29b8942e, update rows in place, add header note "anchors re-derived at plan-audit iter-1 @ d29b8942e".

D3. coverage holes vs "Exhaustive verification audit" — claims-inventory.md (whole file) + spec.md title/HISTORY — Five measured uncovered clusters: (i) §3 Code Standards entire subsection (subject L230–238: coding-standards.md pointer existence — absent even from INV-055's eleven-name list; snake_case convention; fmt.Errorf %w wrapping); (ii) §13 `loadGLMKey()` symbol + `~/.mo/.env.glm` path (L521, Group-B-class citation left unscheduled); (iii) §12 skill-authoring.md/agent-authoring.md pointer existence (L499); (iv) References §20 Vercel pricing row (L638) with ZERO INV id although `EXTERNAL-UNVERIFIABLE` was minted for it; (v) References §27 Agent-Skill row claims (L645); plus §11 inward cross-ref "§6 검증 규율(clear → 측정)" (L487) unverifiable/unlisted. `grep -c` proves 0 coverage for loadGLMKey/coding-standards/skill-authoring/Vercel-Build-Cost/Agent-Skill-Architecture tokens. Closure "sum == 70" would silently certify a non-exhaustive sweep. Severity: major. Class: blocking. Required fix: EITHER extend the inventory with new INV rows (continuing sparse numbering) AND update the closure constant in AC-CLOCAL-006 + progress `inventory_items`, OR author explicit scope-carve-out rows in spec §C naming deliberately-skipped prose with justification (title then loses "Exhaustive"). Prefer extension given vocabulary already exists.

D4. AC-CLOCAL-012 traces to wrong parent — acceptance.md:L18 → REQ-CLOCAL-005 — Surface-naming obligation lives in `spec.md:L37` (audit-basis exception clause); REQ-CLOCAL-005's content is known-unresolved carry-through, unrelated. Misdirected traceability, the rubric's 0.75-band shape. Severity: major(minor impact). Class: blocking. Required fix: move/add the sentence "Each §28/PRIMARY-surface finding shall state WHICH surface was checked" into REQ-CLOCAL-002 (evidence recording) or 010 (pack contents), remap AC-CLOCAL-012 accordingly.

D5. misspelled register symbol in REQ-CLOCAL-005 — spec.md:L54 — "CleanMoManagedPaths" vs ground truth `CleanMoaiManagedPaths` (verified `internal/cli/update/deploy/deploy.go:107`; subject doc L149/L199 and inventory INV-070 spell correctly). A known-unresolved REGISTER entry, misnamed in the requirement that forbids re-adjudicating it. Severity: minor. Class: blocking (one-token fix). Required fix: correct spelling in spec.md:L54.

D6. AC observability gaps — acceptance.md:L10 (AC-004 method), :L37–40 (clause wording), :L20–65 (scenario set) — AC-004's "no verification effort beyond citation checks was spent" is unobservable post-hoc; rewrite as artifact predicate ("KNOWN entries appear exactly once each and no transcript CHK entry contains adjudication output for those four topics"). Scenario blocks missing for AC-005/010/012; matrix methods for 010 ("N confirmed / N attested…") do define the observable report line — acceptable, but 005/012 deserve one-line Thens. Severity: minor. Class: optional. Required fix: reword AC-004; add compact GWT triplets for 005/012.

D7. milestone holes for T1 group — plan.md:L74–77 vs claims-inventory.md:L24–30 — INV-021 (CI-guard yaml existence), INV-022 (doctrine §25.1/§25.3 + leak-test sibling) scheduled only via inventory group tag "T1 (M7)"; M7 bullet names INV-020 explicitly then falls to catch-all. Resolvable, friction-free risk only. Severity: minor. Class: optional. Required fix: name INV-021/INV-022 in the M7 bullet.

D8. GEARS formality nuances — spec.md:L45 (Where-as-event), :L47–48 (Unwanted heading / When-syntax body), :L51 ("may verify" permissive grant inside Ubiquitous req; RQ-5 prefers declarative scope) — None reaches FAIL under MP-2 compound-clause equivalence; informative labels only. Severity: minor. Class: optional. Required fix (optional): relabel 003 Event-driven, convert 002 to "**When**", replace "may verify" with "shall be limited to verifying".

## Regression Check

Iteration 1 — no prior defect history.

## Iteration Delta Instruction for iter-2 (scoped re-audit)

Re-audit iter-2 ONLY against this defect delta, per the Retry Loop Contract:
1. D1: read updated REQ-CLOCAL-004 + AC-CLOCAL-003 (content anchor + mid-run rule present and executable).
2. D2: sample-verify ≥10 repinned anchors mechanically (grep -n vs new claims-inventory.md; expect ≥95% exact alignment, remaining explained).
3. D3: verify each added INV row's anchor exists in subject doc at the cited line OR a documented carve-out exists in spec §C; closure constant consistency across spec/plan/AC/progress.
4. D4: verify remap target REQ text now contains the surface-naming obligation; matrix row updated.
5. D5: grep CleanMoManagedPaths across the SPEC dir → 0 hits.
6. D6: AC-004 rewritten as transcript predicate; GWT blocks present for 005/012.
Cheap static re-runs: MP-1/MP-3/MP-5/MP-6/MP-7 unchanged unless related_specs/frontmatter touched; regression-check D-register D1–D8 resolutions.

## Verification Log (commands executed this audit, this tree d29b8942e)

git rev-parse HEAD / git branch --show-current; wc -l CLAUDE.local.md (=663);
grep -rn "335\|337" (specs dir + reports dir) → 5 hits, all "273–337";
grep -c syscall spec.md → 0; grep -c "[NEEDS CLARIFICATION" plan.md → 0;
grep -o "^| INV-[0-9]*" claims-inventory.md | wc -l → 70;
grep '^status:' .moai/specs/SPEC-STAMP-REACHABILITY-001/spec.md → completed;
ls .moai/reports/t308/ → claims-inventory / verdict-draft / checks-transcript;
grep -rn "CleanMo" internal/ → deploy.go:107 func CleanMoaiManagedPaths;
grep -n spot anchors: handle-*.sh→119, scripts/ci-watch→126, last-cc-version→132, status_line.sh→141, §18→636, §19→637, §23→641, coding-standards→232, skill-authoring→499, loadGLMKey→521;
grep -c coverage-miss tokens in inventory → 0; git log -1 -- CLAUDE.local.md → 20fab1786.
Gaps: PRIMARY-checkout §28 surfaces and template-tree file contents were NOT probed this iteration (run-phase M6/M5 territory, pre-verdict irrelevant); git-strategy.yaml HEAD content not adjudicated (INV-112 is run-phase work).

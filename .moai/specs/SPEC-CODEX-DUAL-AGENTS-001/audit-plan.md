# SPEC Review Report: SPEC-CODEX-DUAL-AGENTS-001

Iteration: 1/2 (Tier M — `plan_audit_tier_ceilings` M=2)
Verdict: **FAIL**
Overall Score: **0.86** (4-category harmonic mean; 7-dimension harmonic = 0.87)
Tier M PASS threshold: 0.80 — the aggregate score CLEARS the threshold. The FAIL verdict is driven solely by the MP-7 must-pass firewall (4 open `[NEEDS CLARIFICATION]` markers), which is score-independent by contract: a high aggregate score never auto-resolves an open clarification marker.

Auditor: plan-auditor (adversarial, M1 context isolation — author reasoning ignored; artifacts on disk only).
Cross-model convergence: run (`mcp__moai__audit_multi`, effective default `codex+glm`); result folded in § Cross-Model Convergence below.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `R-001`..`R-014` sequential, no gaps, no duplicates, consistent 3-digit zero-padding (spec.md:60, 67, 73, 79, 85, 90, 96, 102, 109, 115, 121, 128, 135, 142; every heading enumerated during audit).
- **[PASS] MP-2 EARS/GEARS format compliance** — judgment made against the **requirement layer only** (spec.md §C `R-XXX` entries). All 14 match a canonical pattern: Ubiquitous (R-001 "The agent publisher shall derive…", R-005, R-007, R-010, R-011, R-012), Event-driven `When…shall` (R-002:69, R-004:81, R-006:92, R-008:104, R-014:144), Unwanted `shall not` (R-003:75), Where/capability-gate (R-009:111, R-013:137). No informal should/may in normative text. The `Given…When…Then` entries in acceptance.md are the verification layer and were correctly NOT GEARS-tested (see D5 for one labeling nit on R-008).
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types, verified field-by-field against `.claude/rules/moai/development/spec-frontmatter-schema.md`: id (multi-segment form, sanctioned), title/version quoted ("0.1.0"), status `draft` (enum), created/updated ISO `2026-08-22`, author, priority `P1`, phase `"v3.2.0 target"` (release-target label; not a prohibited stage name), module `internal/template`, lifecycle `spec-anchored`, tags comma-separated, optional `tier: M` (valid enum). No snake_case aliases (spec.md:2-14).
- **[N/A] MP-4 language neutrality** — single-project harness tooling SPEC (agent-definition publication pipeline); not multi-language tooling enumeration. Auto-passes per the single-language precedent.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — only external SPEC reference is `SPEC-CODEX-PHASE2-001` (plan.md:188, uniqueness context). Verified: `.moai/specs/SPEC-CODEX-PHASE2-001/spec.md` exists, `status: completed` (∉ {retired, superseded, archived}) → no reconciliation required, no BLOCKING. Self-references only elsewhere.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall'` = 0 across spec.md, plan.md, acceptance.md → auto-PASS.
- **[FAIL] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION'` matches **4 markers** in plan.md (lines 92, 94, 96, 98); research.md absent (Tier M — N/A for it). Per the marker convention (`.claude/skills/moai-workflow-spec/SKILL.md` §171-191, confirmed read during this audit): markers MUST be settled before Implementation Kickoff Approval; the orchestrator MUST run an AskUserQuestion round for each. Folded into Defects Found as D1 (critical). Substance assessment of each marker is in § Dimension 7 below — all four are genuinely open (not doc-read resolvable), so this is a gate-state failure, not an authoring-quality failure.

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.90 | 1.0 minus minor nits | Requirements single-interpretation; R-003:75-77 embeds an Option-A rationale parenthetical; R-008 label nit (D5) |
| Completeness | 0.85 | 0.75–1.0 | All sections present (§A-§H + 4 `### Out of Scope —` H3 sub-headings with bullets at spec.md:212/226/234/240); frontmatter complete; Tier M AC budget exceeded 19>16 (D3) |
| Testability | 0.90 | 1.0 minus one soft probe | ACs binary with negative tests (AC-005/006) and edge cases (§D.2); AC-P03 "confirmed to inherit the Codex default" is the softest; no weasel words found |
| Traceability | 0.80 | 0.75 band (indirect mappings) | Every R-001..R-014 has ≥1 AC (spec.md §F:192-205); AC-011/AC-012 anchor to §D constraints via a non-REQ "Documentation grounding" row (spec.md:206), not to any R-XXX (D4); no uncovered REQ, no orphaned AC |

Harmonic mean = 4 / (1/0.90 + 1/0.85 + 1/0.90 + 1/0.80) = **0.86**.

## Caller-Dimension Scores (harmonic = 0.87)

| # | Dimension | Score | One-line verdict |
|---|-----------|-------|------------------|
| 1 | Requirement quality | 0.92 | 14/14 GEARS-conformant, atomic, testable; two cosmetic nits (D5) |
| 2 | Mapping-table soundness | 0.80 | Dispositions consistent + fallback rule applied uniformly; **5 evidence cells wrong vs ground truth** (D2) |
| 3 | Byte-identity chain / Option A | 0.90 | Option A judged **defensible** (see § Option A ruling); lead sign-off still owed at kickoff |
| 4 | Ground-truth verification | 0.78 | 6/6 spot-checks pass, §B.1 drift list exact, all t91 citations accurate; but §A.2/§A.3 carry the D2 errors |
| 5 | AC sufficiency (R↔AC) | 0.85 | Full R coverage, no vacuous AC, AC-007 7/4 split verified correct; D3/D4 deductions |
| 6 | Scope fence | 1.00 | M1/M2/M3/M4/M6 all fenced, seams documented (plan §H), no absorption in MS1-MS4 |
| 7 | [NEEDS CLARIFICATION] markers | 0.85 | All 4 genuinely open with probe paths + pre-declared fallbacks; gate-state unresolved → MP-7 FAIL |

---

## Dimension Findings (evidence)

### Dimension 4 — Ground-truth verification (executed, not sampled from the plan)

Verified against `internal/template/templates/.claude/agents/moai/*.md` (worktree, base 4b2f203fe):

**PASS (plan claims confirmed):** manager-git `model: sonnet` (manager-git.md:9); manager-docs `effort: low` (manager-docs.md:11); manager-lead skills = 2 `moai-foundation-core, moai-workflow-project` (manager-lead.md:16-18); plan-auditor has NO `skills:` field; `hooks:` frontmatter on exactly 4 agents (manager-develop.md:18, manager-docs.md:17, manager-spec.md:18, sync-auditor.md:17); builder-harness `memory: user`; manager-design `permissionMode: acceptEdits`; super-advisor + sync-auditor `permissionMode: plan`; effort ladder low×3 / medium×2 / high×5 / xhigh×1 exactly as listed (plan.md:38-40); all 11 carry Bash and TaskCreate; `Agent` tool on manager-lead only; 7 of 11 carry `mcp__moai__*`; AC-007's 7-carry/4-don't agent split correct; frontmatter field order matches plan.md:44-46; `Explore` has no `.md` in the tree (11 files). §B.1 local-drift claim exact: `diff -rq` shows exactly the 6 named files differing. **t91 M0 citations all accurate** — T91BogusEvent silent-ignore 0 warnings (t91 §1:37), probe fields + `T91PROBE-OK` (t91 §2:41), `collaborationspawn_agent` instability (t91 §2:56), `codex mcp add moai` + all 21 tools + "user cancelled MCP tool call" (t91 §5:100-109), project-level hooks merge (t91 §6:116), §8 M5 row quoted verbatim (t91 §8:138), skills symlink premise (t91 §4:93), harness pattern (t91 §9). No over-claims found. Note: t91's probe lived at `.codex/agents/t91probe.toml` (FLAT) — the plan correctly treats the `moai/` subdirectory as unmeasured (P-04), consistent with the evidence.

**FAIL (D2 errors, measured this audit):**

| Cell | Plan says | Ground truth | Evidence |
|---|---|---|---|
| plan.md:35 super-advisor mcp count | 13 | **11** | super-advisor.md:13 tools line |
| plan.md:36 sync-auditor mcp count | 6 | **5** | sync-auditor.md:9 tools line |
| plan.md:40-41 distinct union | "19 distinct of the 21-tool server" | **20 distinct** (only `goal_arm` absent) | union over all `^tools:` lines |
| plan.md:64 §A.3 row 4 (Web) agent set | manager-docs, manager-spec, super-advisor | those 3 **+ builder-harness** (tools carry WebFetch, WebSearch) | builder-harness.md:7 |
| plan.md:68 §A.3 row 8 (DesignSync) | builder-harness, manager-design | **manager-design only** — builder-harness contains no DesignSync token anywhere | builder-harness.md:7 vs manager-design.md:10 |

Rows 4 and 8 look like a complementary transcription swap (builder-harness moved from Web to DesignSync). **No disposition or AC flips**: row 4/8 remain documented drops, row 9's `mcp_servers=["moai"]`-on-7 remains correct, AC-007's agent list is right. But §A.3 is the card's named first-class deliverable and §A.2/§E claim "FILE WINS" re-verification — five wrong cells contradict the SPEC's own measured-not-assumed methodology (§D.7), hence blocking under the internal-consistency class.

### Dimension 2 — Mapping-table soundness

Beyond D2's cells: every class row 1-14 carries exactly one disposition (emit / consequence / documented-drop / deferred-to-M1 / correspondence-note) — AC-013's completeness predicate holds structurally. The ship-omitted fallback rule (plan.md:76-79) is applied uniformly to all three value-bearing optional fields (sandbox row 3, effort row 10, model row 11) and consistent with spec §D.3 and R-008/R-010/R-014. Row 9's coarse-grant documented drop matches R-009. No disposition silently weakens the [HARD] byte-identity constraint — no row touches the `.md` side.

### Dimension 3 — Option A ruling (flagged decision; sharpest scrutiny)

**Ruling: Option A (`.md` IS the neutral core + mapping manifest; `.md` publication = identity) is a defensible reading of the single-source constraint — auditor-endorsed — and is superior for the [HARD] regression ban.** Reasoning:

1. The literal phrase "both outputs generated from a neutral definition" is not symmetric-satisfied under Option A; the *intent* (no independently hand-maintained second representation → no drift surface) IS satisfied, and satisfied **by construction**: the `.md` cannot drift from itself. Option B would make `.md` byte-identity test-guarded (renderer fidelity) rather than structural — strictly weaker for constraint 1.
2. The spec is internally consistent about this: R-001 (spec.md:62-65) *defines* the neutral layer as "the agent `.md` definitions plus exactly one machine-readable Codex mapping manifest", and R-003 (spec.md:77) states "the `.md` is the neutral source itself". The user-story sentence alone (spec.md:21-23) would be ambiguous; R-001 disambiguates. No internal contradiction found.
3. The plan's "honest reading" paragraph (plan.md:117-119) explicitly flags the interpretation and names the escalation path ("if the lead requires literal symmetric generation, that is Option B") — correct handling of a highest-change-likelihood decision.
4. Residual: the decision authority is the lead, not the auditor. The sign-off must be recorded at Implementation Kickoff Approval (it can ride the same AskUserQuestion round as the marker resolutions). Note also that R-003 forecloses Option B at requirement level — if the lead DID flip to Option B, spec.md §C would need amendment first. This is acceptable as long as the sign-off is explicit.

### Dimension 6 — Scope fence

M1 (skills canonicalization): class 6 deferred, no skills field emitted (plan.md:66); M2 (AGENTS.md): absent from all deliverables; M3 (hook adapter): class 12 documented drop + M3 seam (plan.md:72, §H:239); M4 (wiring): §H seam only, "M5 exposes no CLI" (plan.md:236); M6 (plugin): grounded in t91 §8 (plugin_hooks removed). MS1-MS4 contain no absorbed sibling-card work; manifest override keys ship empty with an explicit YAGNI note (plan.md:228-229). Clean.

### Dimension 7 — Marker substance

All four markers (plan.md:92-100) are genuinely unmeasurable at plan time — the value sets live inside codex-cli 0.147.0's acceptance behavior; t91 §2 verified the fields exist but never enumerated allowed values; nothing in the repo can resolve them by doc read. Each carries a probe path (P-01..P-04) AND a pre-declared fallback. They are decision-ready for a single cheap AskUserQuestion round. The gate still fires because they are unresolved *at audit time* — which is the designed mechanism, not an authoring defect.

---

## Defects Found

D1. **MP7-CLARIFICATION-GATE** — plan.md:92,94,96,98 — 4 unresolved `[NEEDS CLARIFICATION: ...]` markers (sandbox_mode value set; model_reasoning_effort enumeration; model omission semantics; agents-dir layout). Markers must be settled via orchestrator AskUserQuestion before Implementation Kickoff Approval (marker convention SKILL.md §173-191). — Severity: **critical** — Class: **blocking** — Required fix: orchestrator runs ONE AskUserQuestion round confirming each marker's pre-declared default: (1) probe P-01 first, unconfirmed → omit `sandbox_mode` entirely in v1; (2) probe P-02 first, identity mapping only if every emitted value confirms, else omit `model_reasoning_effort`; (3) omit `model` on all 11 (bogus-string probe documents the hazard); (4) subdir layout preferred, flat + `moai-` filename prefix if P-04 shows no subdir scan. Then manager-spec replaces the four markers in plan.md §A.4 with the recorded decisions (keeping the probes as confirmation tasks in §A.4's table, which is already marker-free).

D2. **INVENTORY-ACCURACY** — plan.md:35,36,40-41,64,68 — five ground-truth errors in the verified inventory / mapping table: super-advisor mcp 13→**11**; sync-auditor mcp 6→**5**; distinct union 19→**20 of 21** (goal_arm the only absent); §A.3 row 4 Web class must add **builder-harness**; §A.3 row 8 DesignSync must drop builder-harness (**manager-design only**). Contradicts the §A.2 "FILE WINS / re-verified" claim and the SPEC's measured-not-assumed methodology (§D.7). — Severity: **major** — Class: **blocking** — Required fix: manager-spec corrects the five cells in plan.md §A.2 (mcp column + closing line) and §A.3 rows 4/8/9; re-run the same greps used by this audit to confirm.

D3. **TIER-AC-BUDGET** — acceptance.md (whole file) — 19 labeled `AC-XXX` entries (AC-001..013 + AC-P01..P06) exceeds the Tier M acceptance-criterion ceiling of 16 (`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier; requirements 14 ≤ 16 are fine). — Severity: minor — Class: optional — Required fix: either (a) tier up to L (frontmatter `tier: L`, add design.md/research.md per Tier L artifact set), or (b) reclassify AC-P01..P06 in acceptance.md §D.1 as "probe records" outside the AC budget with an explicit one-line justification (they measure Codex, not the deliverable), leaving 13 close-gate ACs ≤ 16.

D4. **TRACE-ANCHOR** — spec.md:206 / acceptance.md AC-011, AC-012 — AC-011 (neutrality) and AC-012 (M0 citations) trace to §D constraints via the non-REQ "Documentation grounding" matrix row, not to any `R-XXX` — indirect traceability anchor. — Severity: minor — Class: optional — Required fix: cheapest — annotate the §F row "Documentation grounding (§D.6/§D.7 constraints — deliberate non-REQ anchor)"; better — promote to an R-015 (neutrality + evidence-grounding requirement).

D5. **GEARS-LABEL-NITS** — spec.md:102,75 — R-008's self-label "(Event-detected)" is not a canonical pattern name (structure is a conformant event-driven When/shall — no MP-2 impact); R-003 embeds the Option-A rationale parenthetical "(the `.md` is the neutral source itself)" inside requirement text. — Severity: minor — Class: optional — Required fix: relabel R-008 "(Event-driven)"; optionally move R-003's parenthetical to §D.1 as rationale.

D6. **DEPLOY-PATH-PREMISE** — plan.md:139, AC-010 — ".codex/ rides the existing template mirror deployment" is a plan-time premise: the cleanup-side managed-roots list (CLAUDE.local.md §2.3) does not include a `.codex/` root, and deploy coverage for the new root is exactly what the MS3 fixture test exists to prove. Not a defect (self-flagged, test-covered) — recorded so the fixture test is treated as load-bearing and any stale-`.codex` cleanup gap is routed to M4 wiring. — Severity: minor — Class: optional — Required fix: none at plan time; ensure AC-010's fixture test actually exercises the deploy path (not just embed-FS presence).

## Regression Check

Iteration 1 — N/A (first audit).

## Recommendation (manager-spec + orchestrator fix route)

1. **[D1, blocking]** Orchestrator: AskUserQuestion round on the four §A.4 markers using their pre-declared defaults (see D1); record decisions; manager-spec rewrites plan.md §A.4 markers into recorded decisions + probe-confirmation tasks.
2. **[D2, blocking]** manager-spec: correct the five inventory/mapping cells (exact values in D2's table).
3. **[D3/D4/D5, optional]** Route at orchestrator discretion: tier decision (L vs probe-record reclassification), §F anchor annotation, label fix.
4. **Lead sign-off**: record Option A acceptance (or Option B escalation, which would require amending R-001/R-003 first) at Implementation Kickoff Approval.
5. **Re-audit (iteration 2/2)**: scoped to the D1+D2 delta per the Retry Loop Contract — not a from-scratch re-audit. Everything else (requirement layer, scope fence, M0 citations, AC mechanics) is verified clean by this iteration.

## Cross-Model Convergence (residual risk)

`mcp__moai__audit_multi` executed (claude anchor fail / required; codex required; glm advisory). Result: `overall_verdict: "fail"`, `disagreement_flag: true`, `residual_risk_note`: **"cross-model disagreement (advisory, NOT a block): pass=[codex(required)] fail=[claude(required), glm(advisory)]"**, `fail_open_backends: []`.

**Target-mismatch caveat (auditor's honest reading of the backend outputs):** neither secondary backend reviewed the SPEC artifacts. codex's review targeted the primary checkout's unrelated untracked files (`.moai/config/astgrep-rules/`, `ci-autofix-protocol.md`, ci-watch scripts — its own summary states "git diff … empty. I reviewed the untracked … artifacts"), and its verdict field (`pass`) contradicts its summary text ("FAIL — revise required"), so it is internally inconsistent AND off-target. glm's findings describe a hallucinated user-profile REST API (`SELECT * FROM users`, `/api/users/{id}`) with zero correspondence to any artifact in this SPEC. Consequently the secondary verdicts provide **no corroborating and no contradicting evidence** about SPEC-CODEX-DUAL-AGENTS-001; the fail-open identity applies and this FAIL verdict rests entirely on the in-session evidence cited above. The convergence run's overall `fail` happens to agree with the anchor verdict but should not be cited as independent confirmation.

---

*Evidence commands executed this audit: frontmatter grep over all 11 template agents; per-file `mcp__moai__` token counts (awk gsub on `^tools:` lines); distinct-union sort over `^tools:` lines; Web/DesignSync/Agent/Task/Bash carriage checks anchored to the tools line; `diff -rq` local vs template agents; `grep '\[NEEDS CLARIFICATION'`; `grep -c syscall`; SPEC-ID extraction + cross-SPEC status read; full read of `.moai/reports/t91/README.md` (primary checkout).*

---

# SPEC Review Report: SPEC-CODEX-DUAL-AGENTS-001 — Iteration 2

Iteration: 2/2 (Tier M ceiling reached)
Verdict: **FAIL**
Overall Score: **0.92** (4-category harmonic; 7-dimension harmonic = 0.94; iter1 was 0.86/0.87 — score improved, no STOP signal)
Scope: the D1+D2 delta plus D3/D4/D5 spot-checks and regression, per the Retry Loop Contract. All claims in the revision notice were independently re-verified against the artifacts (not taken on report).

## Regression Check — Iteration-1 Defects

- **D1 (critical, markers): RESOLVED.** `grep -c 'NEEDS CLARIFICATION' plan.md` → 0 (auditor-executed). §A.4 now carries four lead-ratified `DECIDED` entries (plan.md:93-105) whose content matches the audit-recommended resolutions exactly (P-01 omit-if-unconfirmed; P-02 identity-only-if-every-value-confirms; P-03 omit `model` on all 11; P-04 subdir preferred, flat+`moai-` fallback). Remaining string matches of the marker text are in progress.md (historical Phase-2 authoring summary — outside the MP-7 verb scope, which binds plan.md + research.md only) and in this report itself. **MP-7 PASSES.**
- **D2 (major, inventory cells): PARTIALLY RESOLVED — the FAIL driver.** All five enumerated table cells are fixed and re-verified: plan.md:35 super-advisor `yes (11)` ✓; plan.md:36 sync-auditor `yes (5)` ✓; plan.md:40-42 `20 distinct of the 21-tool server — goal_arm is the only absent tool` ✓ (rationale added, and it is correct: `goal_arm` is orchestrator-only per `.claude/rules/moai/core/moai-mcp-tools.md` § Unwired-by-design); plan.md:65 row 4 Web class now includes builder-harness ✓; plan.md:69 row 8 DesignSync manager-design only ✓. **However, the sixth location named in the iteration-1 fix route ("§A.3 rows 4/8/9") was left stale: plan.md:70 (row 9) still reads "19 distinct tools across 7 agents"** — now directly contradicting the corrected §A.2 (line 41: "20 distinct"). The revision converted the wrong figure into a live internal contradiction inside the first-class deliverable.
- **D3 (minor, Tier M AC budget): RESOLVED via route (b), verified.** acceptance.md §D.1:143-147 — budgeted set AC-001..AC-013 (13 ≤ 16); AC-P01..P06 relabeled probe records outside the AC budget; P-01..P-04 remain **required records** whose absence blocks the Definition of Done ("probe records P-01..P-04 filed with the manifest enums locked or the affected fields omitted", §D.1:152-155), so R-014's close obligation is preserved, not weakened. Legitimate.
- **D4 (minor, trace anchor): RESOLVED.** spec.md:206 now reads "AC-011, AC-012 (§D.6/§D.7 constraints — deliberate non-REQ anchor)".
- **D5 (minor, labels/parenthetical): RESOLVED.** R-008 relabeled "(Event-driven)" (spec.md:102); R-003 (spec.md:73-77) no longer embeds the Option-A parenthetical and is a clean Unwanted-pattern requirement; the rationale is relocated to acceptance.md §D.1:148-151 with explicit provenance ("relocated from the requirement text per audit D5"). The anchor landed in acceptance.md §D.1 rather than spec.md §D — judged **satisfactory**: requirement text cleaned, rationale preserved retrievably, no contradiction (spec.md §D constraint 1 independently carries the regression ban). Location was discretionary.
- **Option A sign-off: RECORDED.** plan.md:109 "Option A (LEAD-APPROVED 2026-08-22)" + plan.md:120-124 "decision is closed for run-phase". The iteration-1 residual (lead sign-off owed) is discharged.

## Must-Pass Results (iteration 2)

- **[PASS] MP-1** — 14 `### R-` headings, R-001..R-014, no gaps/dupes (`grep -c '^### R-'` = 14).
- **[PASS] MP-2** — changed entries re-verified: R-003 clean Unwanted (spec.md:75-77), R-008 Event-driven When/shall (spec.md:104-107); remaining 12 unchanged from iteration 1's full check.
- **[PASS] MP-3** — frontmatter unchanged and schema-valid (spec.md:1-15).
- **[N/A] MP-4** — single-project harness tooling SPEC.
- **[PASS] MP-5 (D7)** — SPEC-ID references unchanged (self + SPEC-CODEX-PHASE2-001, status completed); no reconciliation required.
- **[PASS] MP-6 (D8)** — `grep -c syscall` = 0 across spec/plan/acceptance.
- **[PASS] MP-7** — plan.md marker count 0; research.md absent (Tier M).

All seven must-pass criteria pass. The FAIL verdict is **not** must-pass-driven: it is the Retry Loop Contract's unresolved-prior-defect clause plus the blocking internal-consistency finding below.

## Defects Found (iteration 2)

D2-R. **ROW9-DISTINCT-COUNT** — plan.md:70 (§A.3 row 9) — "19 distinct tools across 7 agents" contradicts the corrected §A.2 (plan.md:41: "20 distinct of the 21-tool server"). Named in the iteration-1 fix route ("§A.3 rows 4/8/9") but left stale; the partial fix converted a wrong figure into a live self-contradiction in the card's first-class deliverable. Ground truth (auditor-measured, iteration 1, re-confirmed): 20 distinct tokens across the 7 agents; `goal_arm` the only absent tool. — Severity: **major** — Class: **blocking** — Required fix: one token — plan.md:70 "19 distinct tools" → "20 distinct tools".

N1. **T83-CITATION-UNVERIFIABLE** — plan.md:98 — the new DECIDED entry cites "lead-cited t83 measurement" ("one bad key can kill the whole file"); no t83 artifact exists under `.moai/reports/` (auditor-checked, count 0). The decision it supports is lead-ratified independently, so nothing blocks on it, but a SPEC whose §D.7 mandates measured-and-cited facts should not carry an unlocatable citation. — Severity: minor — Class: optional — Required fix: cite a locatable artifact path or drop the parenthetical.

N2. **PROGRESS-HISTORICAL-STALENESS** — progress.md:27 — the Phase-2 authoring summary still describes "4 [NEEDS CLARIFICATION] markers" as a feature of plan.md. Outside the MP-7 verb scope; acceptable as a point-in-time authoring record (retroactively editing landed records falsifies them), noted so nobody mistakes it for a live marker. — Severity: minor — Class: optional — Required fix: none required; optionally append a one-line revision note ("markers converted to DECIDED entries 2026-08-22").

## Category Scores (iteration 2)

| Dimension | Score | Evidence |
|-----------|-------|----------|
| Clarity | 0.93 | R-003 cleaned, decisions recorded; one internal number contradiction (D2-R) |
| Completeness | 0.95 | AC budget resolved (13 ≤ 16), all sections intact |
| Testability | 0.90 | unchanged from iteration 1 |
| Traceability | 0.90 | D4 anchor annotated |

Harmonic = **0.92**. Caller-dimension harmonic = **0.94** (D1 0.95 / D2 0.88 / D3 0.95 / D4 0.92 / D5 0.90 / D6 1.00 / D7 1.00).

## Recommendation

1. **[D2-R, blocking]** Apply the one-token fix at plan.md:70 ("19" → "20 distinct tools across 7 agents").
2. **Tier M ceiling reached (2/2).** The retry loop is exhausted; per the max-iteration cap the orchestrator escalates with the three canonical options: (a) **user override extending to iteration 3** — recommended; the re-audit scope would be a single grep (`grep -n '19 distinct' plan.md` → 0) over the one-token delta; (b) PASS-with-debt accepting the stale figure as documented debt — a poor trade against a one-token fix; (c) scope-reduction — not applicable at this defect size. Verdict authority stays with the auditor; orchestrator self-verification of the fix does not substitute for the delta re-audit (AP-SPD-004).
3. **[N1/N2, optional]** Route at discretion.

Everything else in the revision is verified genuine: the marker conversions, the five corrected cells, the AC-budget relabel, both spec.md annotations, and the Option A lead sign-off. The distance between this FAIL and a PASS is exactly one token.

*Evidence commands executed this iteration: `grep -c 'NEEDS CLARIFICATION' plan.md` (0); marker-match location scan (`grep -rln`, whole SPEC dir); plan.md §A.2-§A.6 re-read (lines 18-152); acceptance.md §D.1 re-read (lines 130-181); spec.md R-002..R-008 + §F + frontmatter re-read; `grep -c '^### R-'` (14); `grep -n 'distinct'` plan.md (20 at :41 vs 19 at :70); `ls .moai/reports | grep -c t83` (0); `grep -c syscall` (0); SPEC-ID re-extraction.*

---

# SPEC Review Report: SPEC-CODEX-DUAL-AGENTS-001 — Iteration 3 (FINAL)

Iteration: 3 (over the Tier M ceiling of 2 — executed under the documented lead override approved 2026-08-22; scope exactly as proposed in the iteration-2 recommendation)
Verdict: **PASS**
Overall Score: **0.92** (4-category harmonic; 7-dimension harmonic = 0.96; trajectory 0.86 → 0.92 → 0.92, monotonically non-decreasing — no STOP signal was ever emitted)
Scope: the one-token D2-R fix + the N1 citation cleanup + no-other-change regression. Every claim re-verified independently against the artifacts.

## Regression Check — Iteration-2 Defects

- **D2-R (blocking): RESOLVED.** `grep -n '19 distinct' plan.md` → 0 matches. plan.md:70 (§A.3 row 9) now reads "20 distinct tools across 7 agents", consistent with §A.2 (plan.md:41, "20 distinct of the 21-tool server"). The internal contradiction in the first-class deliverable is gone; both instances now match the auditor-measured ground truth (20 distinct across the 7 agents, `goal_arm` the only absent of 21).
- **N1 (optional): RESOLVED via the drop-the-citation route.** `grep -n 't83' plan.md` → 0 matches. plan.md:98 now attributes the omit-if-unconfirmed rationale to "lead-cited measurement, 2026-08-22 ratification)" — a ratification attribution, not an unlocatable artifact citation. No phantom references remain.
- **N2 (optional): left as-is, accepted.** progress.md unchanged since iteration 2 (mtime 03:12 < plan.md 03:17); its Phase-2 summary's historical marker description remains a point-in-time authoring record — the acceptable disposition per the iteration-2 ruling.

## No-Other-Change Regression (verified)

- `grep -c 'NEEDS CLARIFICATION' plan.md` → 0 (MP-7 stays green)
- `grep -c '^### R-' spec.md` → 14 (MP-1: R-001..R-014 intact)
- `grep -c syscall` → 0 across spec/plan/acceptance (MP-6)
- acceptance.md:143 AC-budget line intact (13 ≤ 16, probes as records — D3)
- spec.md:206 deliberate non-REQ anchor intact (D4)
- plan.md:109 Option A LEAD-APPROVED intact
- File mtimes corroborate the "only plan.md changed" claim: spec.md/acceptance.md 03:11, progress.md 03:12 (iteration-2 revision batch), plan.md 03:17 (this delta).

## Must-Pass Results (final)

- **[PASS] MP-1** — 14 sequential R-001..R-014 headings, no gaps/dupes.
- **[PASS] MP-2** — all 14 requirements GEARS-conformant (requirement layer; full check iteration 1, changed entries re-checked iterations 2-3).
- **[PASS] MP-3** — 12 canonical frontmatter fields + tier, schema-valid, no aliases.
- **[N/A] MP-4** — single-project harness tooling SPEC.
- **[PASS] MP-5** — no D7 blocking (only external ref SPEC-CODEX-PHASE2-001, status completed).
- **[PASS] MP-6** — zero `syscall` occurrences.
- **[PASS] MP-7** — zero `[NEEDS CLARIFICATION]` markers in plan.md (4 converted to lead-ratified DECIDED entries, iteration 2).

No must-pass failure, no blocking finding, no unresolved prior-iteration defect. Aggregate score 0.92 clears the Tier M threshold (0.80) on every iteration's recomputation.

## Final Category Scores

| Dimension | Score | Note |
|-----------|-------|------|
| Clarity | 0.95 | contradiction resolved; decisions recorded and internally consistent |
| Completeness | 0.95 | AC budget 13 ≤ 16; all sections present |
| Testability | 0.90 | binary ACs + negative tests + edge cases (unchanged) |
| Traceability | 0.90 | full R coverage; constraint-anchored ACs explicitly annotated |

Harmonic = **0.92**. Caller-dimension harmonic = **0.96** (D2 mapping-table and D4 ground-truth now fully accurate).

## Defects Found (final)

None blocking. One accepted optional note: N2 (progress.md historical marker description — point-in-time record, deliberately not retro-edited).

## Recommendation

PASS — proceed to run-phase. The lead's conditional Implementation Kickoff Approval condition is met by this verdict; the gate record itself belongs to the lead (the kickoff approval is mandatory and score-independent by rule — an audit PASS never auto-bypasses it, and here the lead's prior conditional grant supplies the approval). Run-phase entry notes for manager-develop: the four §A.4 DECIDED entries bind MS2 (probe-first, omit-unconfirmed); MS3's AC-010 deploy fixture test is load-bearing for the `.codex/` root premise; the §B.1 six-file local-vs-template drift remains a maintainer hazard the golden test pins.

*Evidence commands executed this iteration: `grep -n '19 distinct' plan.md` (0 matches); `grep -n '20 distinct' plan.md` (:41, :70); `grep -n 't83' plan.md` (0); `grep -n 'lead-cited' plan.md` (:98); regression fingerprint batch (marker count 0, R-heading count 14, syscall 0, AC-budget/§F/LEAD-APPROVED greps); SPEC-dir `ls -la` mtime corroboration.*

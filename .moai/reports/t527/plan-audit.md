# SPEC Review Report: SPEC-WEB-ANCHOR-SCOPE-001

Iteration: **2/2** (Tier M ceiling, `harness.plan_audit_tier_ceilings`) — delta re-audit
Verdict: **PASS**
Overall Score: **0.99** (Tier M PASS threshold 0.80)

> Iteration history: iter 1 (below, unmodified) returned **FAIL 0.97** on the single MP-7
> clarification-gate must-pass. The gate round resolved it (zero-repair close via
> REQ-WAS-006; REQ-WAS-002 unchanged); the iter-2 delta re-audit (final section) confirms
> the resolution, verifies the optional fixes D2/D3 landed, and confirms no new defect in
> the touched sections. Score progression 0.97 → 0.99 (no regression; no STOP signal).

Auditor: plan-auditor (independent). The audittee's headline claims were treated as unverified
claims and each was independently re-measured on this tree. No author reasoning context was
injected; all judgments below rest on the five SPEC artifacts plus this tree's code.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-WAS-001..007 at spec.md:50-98: sequential, zero-padded, no gaps, no duplicates.
- **[PASS] MP-2 EARS/GEARS format compliance** (judged against the `REQ-XXX` requirement layer in spec.md §B; the `AC-WAS-XXX` Given-When-Then entries in acceptance.md are the verification layer and were graded under Group 4, not here) — all 7 REQs match a GEARS pattern: Ubiquitous "The SPEC/run phase shall …" (REQ-WAS-001/003/004/005, spec.md:52, 66, 72, 80), Unwanted "shall not permit" (REQ-WAS-002, spec.md:60), Event-driven "When [trigger], the <subject> shall …" (REQ-WAS-006 spec.md:88, REQ-WAS-007 spec.md:94). One cosmetic label defect (D3) does not touch the requirement text's conformance.
- **[PASS] MP-3 YAML frontmatter validity** — spec.md:1-17 carries all 12 canonical fields with correct types (`id`, quoted `title`, quoted semver `version: "0.1.0"`, `status: draft` (enum), ISO `created/updated: 2026-09-07`, `author`, `priority: P2`, `phase: "v3.1.5 target"` (release label, not a stage name), `module: internal/web`, `lifecycle: spec-anchored`, comma-separated `tags`). No rejected snake_case aliases. Optional `era: V3R6` / `tier: M` / `related_specs` are schema-tolerated extras.
- **[N/A] MP-4 Section 22 language neutrality** — single-language (Go, `internal/web`) SPEC; auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — references extracted: SPEC-WEB-CODEX-PANEL-001, SPEC-MCP-CONSOLE-001. Both exist; both `status: completed` (measured this audit: `grep '^status:'` on both spec.md files). Neither is retired/superseded/archived, so no reconciliation obligation fires; SPEC-WEB-CODEX-PANEL-001 is additionally the SPEC's named subject (the mirror) with its close commit 1aaf4951f pinned as the mirror-presence sentinel. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — the literal substring `syscall` does not appear in any of the five artifacts (full-file reads). D8-4 auto-PASS.
- **[FAIL] MP-7 clarification gate** — `plan.md:92-96` carries an unresolved `[NEEDS CLARIFICATION: zero-repair close vs defensive hardening]` marker (research.md: none). Per the MP-7 firewall this is a must-pass failure folded into Defects Found as D1 (severity critical, class blocking) and flagged as a **clarification-gate finding**. Framing: the marker is properly placed (end of plan.md, directly under the gate round's read surface) and presents both options fairly with the REQ-WAS-002 tension named explicitly — its PRESENCE is a gate-routing finding, not a content defect. Resolution requires exactly the act MP-7 mandates: the orchestrator's `AskUserQuestion` round before Implementation Kickoff Approval, then manager-spec folds the operator's answer in and removes the marker; the confirming re-audit is scoped to this one-item delta.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 1.00 | 1.0 | Every REQ has a single reading; the (c) verdict is a defined three-valued predicate (acceptance.md D.2); the discriminator is a crisp conjunction (research.md:88-104). No pronoun ambiguity observed. Cosmetic label slip at spec.md:92 (D3) does not create interpretive ambiguity. |
| Completeness | 0.95 | ~1.0 | All sections present with required H3 Out-of-Scope sub-headings + bullets (spec.md:115-138, three topics); frontmatter complete; all 5 artifacts substantive. Deduction: one unreproducible figure in research.md:10 ("84 files"; measured 82 — D2). |
| Testability | 0.95 | ~1.0 | All 7 ACs binary-testable with recorded-output obligations (acceptance.md D.1-D.7); zero weasel words; AC-WAS-005 mandates OBSERVING and recording RED, not expecting it. Deduction: "or an equivalent already-scoped slice" (D4) leaves one judgment locus. |
| Traceability | 1.00 | 1.0 | 7 REQs ↔ 7 ACs bidirectional (acceptance.md:5-13); no orphan ACs; no uncovered REQs (AC-WAS-006's dual mapping REQ-WAS-004/006 is valid). |

## Verification Record (5-section evidence format)

**Claim 1 — Raw superset is 60 sites / 19 files on tree bf779ecf2 (audittee headline 1).**
- Evidence: `grep -o 'strings\.Index(' internal/web/*_test.go | wc -l` → `60`; `grep -l … | wc -l` → `19`. Full `grep -n` site list reconciled 1:1 against research.md §5's 60 rows — every file:line pair in the table matches a real site, and no site exists without a row (agentfm_polish 162-166→rows 4-8; codex_panel 269/274/285/290/295→rows 9-13; console_ux_fix 28/31/38/76/82/279→rows 14-19; i18n_governance 55→29; i18n 326/367/368/404/409→22-26; htmx 52/53→20-21; mcp_console 115→27; profile_bar 332/342/343→36-38; mx_rawview 58→28; schema_sections 148/152→39-40; schemaform 72/73/94→30-32; todo_route 142→47; schema_select 88/93→33-34; webux_followup 62/66/80/89/366/370/387→54-60; tab_layout 55/60/81/86→41-44; todo_section 157/161→45-46; validate 244→35; webux_haiku 31/35/48/57/147/151→48-53).
- Baseline-attribution: this run, this tree (HEAD `bf779ecf2`, branch `WT-anchor-scope-sweep`, worktree root verified).
- Gaps: none on the superset itself.
- Residual-risk: none material — the measurement is the audit's least contentious claim and it reproduced exactly.

**Claim 2 — Tab order audit(7) < codex(8) < mcp(11) (angle B precondition).**
- Evidence: `wantTabOrder` at internal/web/tab_layout_test.go:23-27 — positions counted: identity1, language2, launch3, llm4, workflow5, git-worktree6, **audit7, codex8**, agentfm9, report10, **mcp11**, crosssession12, feedback13, gate14. `TestTabPanelRenderOrderMatchesTabs` (line 50) asserts the rendered sequence matches, so the order is test-pinned, not prose-pinned.
- Baseline-attribution: this run, this tree.
- Gaps: none.
- Residual-risk: if tab order ever changes, the discriminator's ORDER ground truth moves with it — but the order is itself under test, so drift is detected by an existing guard.

**Claim 3 — Structural safety: mirror emits no `name=`, `<form`, `data-panel=`, `<script>` (headline 4, angle C).**
- Evidence: `grep -c 'name=\|<form\|data-panel=\|<script' internal/web/fieldsets_codex_templ.go` → `0`. Read of `codexMirrorRowView` confirms the complete emission surface per row: `<div class="field field--ro">…<code class="key">` + row.Name + `</code> <span class="field__label" data-i18n="f.<F>.title">` + F + `</span>`, value span `class="in in--ro"`, and `<a class="mirror-edit" href="/settings?tab=<owner>" data-i18n="tab.codex.edit_on.<owner>">` — exactly research.md §3's inventory. `codexmirror.go:91-102` `isCodexMirrorField` matches §3's families verbatim (workflow.audit.gates.codex; workflow.audit.codex.*; workflow.codex.*; codex-prefixed MCP catalogue tools) plus the declared exception workflow.audit.model handled by name in `codexMirrorGroups`.
- Baseline-attribution: this run, this tree.
- Gaps: the grep covers the four claimed token families; other hypothetical collision forms were checked by reading, not by grep (none found).
- Residual-risk: a future template edit could reintroduce a colliding token; M1's re-run discipline (plan.md C.3) covers currency at run phase.

**Claim 4 — Angle B: the highest-risk class (mirrored needle + owning panel after tab 8 + exposed receiver) contains exactly one site, the t509-repaired one.**
- Evidence: full mirrored-family grep (`mcp\.tools\.|workflow\.audit\.gates\.codex|workflow\.audit\.codex|workflow\.codex\.|workflow\.audit\.model` over internal/web/*_test.go). The only corpus anchor on a mirrored family is mcp_console_test.go:114-115 (`chip := `<code class="key">mcp.tools.<tool>.enabled</code>`` fed to `strings.Index`) — row 27, REPAIRED-BY-T509; the repair verified in code at :87 (`body := panelHTML(t, renderConsolePage(t), "mcp")`). No `workflow.audit.*`/`workflow.codex.*` field name appears as a `strings.Index` needle anywhere in the corpus. All 36 exposed-(c)=false rows were walked needle-by-needle against the mirror emission surface: every one fails DUPLICATION on a structural ground (name=-family needles vs a name=-free mirror; data-panel= vs zero template matches; sec.*/agentfm.*/agentdesc.*/src=/form/action=/details/rawview — none emitted by `codexMirrorRowView`). The subtle case — schema_select_preserve_test.go:88 with `fieldName` possibly `workflow.audit.model` — anchors on the `name="…"` form, which the mirror structurally cannot emit, so (c)=false survives. Slice claims spot-verified in code: row 19's receiver starts at `strings.LastIndex(body, 'data-i18n="sec.agentfm.title"')` then slices (agentfm heading, position 9 > 8) — NOT-EXPOSED correct; rows 9-13's receiver is `panelHTML(t, html, "codex")` (codex_panel_test.go:224, 371) — already the mirror region, NOT-EXPOSED correct.
- Baseline-attribution: this run, this tree.
- Gaps: classification of rows 1-3's receiver as EXPOSED was accepted from the needle-structure argument (non-mirrored needles make the verdict robust either way) rather than by reading agentfm_policy_test.go's body construction.
- Residual-risk: **the 0-verdict survived angle A and angle B scrutiny** — it is not vacuous: the corpus is complete (60/60), the sole (c)=true-able needle family has exactly one member and it is repaired, and every remaining exclusion rests on the template's structural non-emission, not on judgment. The ORDER condition's only positive grounding remains the single t509 example (disclosed honestly at research.md:186-190); an audit-family chip needle has no positive control — inherent to a corpus that contains none, and tab order is test-pinned.

**Claim 5 — Green baseline on the mirror-present tree (headline 5).**
- Evidence: `unset MOAI_KANBAN … && go test -count=1 ./internal/web/` → `ok  github.com/modu-ai/moai-adk/internal/web  6.452s` (fresh, cache-bypassed). Mirror presence verified materially: mirror sources present on this tree and codex tests (including codex_panel_test.go) pass.
- Baseline-attribution: this run, this tree (bf779ecf2).
- Gaps: the sentinel `git merge-base --is-ancestor 1aaf4951f HEAD` was NOT re-executed by this audit (dispatch forbids git commands); material verification (mirror code present + codex tests green) substitutes, and research.md:12-13 records the lane orchestrator's sentinel check.
- Residual-risk: negligible — a mirror-absent tree could not render the mirror rows the passing codex tests assert.

**Claim 6 — Mutation design is discriminating, not an absence guard (angle A.ii, headline 6).**
- Evidence: AC-WAS-005 (acceptance.md:47-56) reads "**When** mcp_console_test.go:115 is temporarily reverted … and the test runs **Then** the test FAILS … the RED output is recorded, the mutant is restored, and the restored test passes. A discriminator adopted without this RED cannot claim it would catch the defect family (absence-guard rule: RED-now evidence or no adoption)." The wording forces observing and recording RED, per the project's recorded lesson that the mutant is the only discriminating evidence. Mechanical plausibility verified by reading the test: reverting :87 to the full page puts the first chip occurrence on the mirror row (tab 8), whose 400-char back-window carries no "Write-capable" badge → the write-capable codex tools Errorf → RED. Honest deferral: progress.md §E.1 gaps lists "mutant RED not yet executed (run-phase M2)"; research.md §6/§8 frames the discriminator as hypothesis until then.
- Baseline-attribution: code-read on this tree; the RED itself correctly NOT claimed at plan phase.
- Gaps: RED is expected-by-reasoning, not yet observed — by design; M2 must capture it.
- Residual-risk: if the mutant unexpectedly stays green, the discriminator collapses — which is exactly the failure M2 exists to catch before adoption.

**Claim 7 — Card [HARD] encoding + tree attribution + progress honesty (angles D, E, F).**
- Evidence: bulk replacement forbidden (REQ-WAS-002 spec.md:58-62; plan.md §G line 78); classification table is the gating artifact (REQ-WAS-001 spec.md:50-56; M1 exit, plan.md:50-54); mirror-present regression mandated (REQ-WAS-004 spec.md:71-76; AC-WAS-006 acceptance.md:57-62); prior card figures appear only as context with an explicit disclaimer (research.md:28-32, spec.md §C.1, plan.md B.2); progress.md §E.1 records the go-build/vet gap honestly ("no go build/vet run at plan phase") and the unexecuted mutant. REQ-WAS-002/REQ-WAS-006 are internally consistent (006 is the zero-TRUE special case of 002) and the marker names their boundary precisely.
- Baseline-attribution: artifact reads on this tree.
- Gaps: none.
- Residual-risk: none observed.

## Defects Found (structured defect-list)

D1. CLARIFICATION-GATE — plan.md:92-96 — unresolved `[NEEDS CLARIFICATION]` marker (zero-repair close per REQ-WAS-006 vs defensive hardening currently forbidden by REQ-WAS-002) — Severity: critical — Class: blocking (MP-7 must-pass; score-independent) — Required fix: orchestrator runs the AskUserQuestion gate round presenting both options; manager-spec records the operator's decision in plan.md (amending REQ-WAS-002 if hardening is chosen) and removes the marker; confirming re-audit scoped to this delta. The marker's placement and fair two-option presentation were verified correct — this is a gate-routing finding, not a content defect.

D2. research.md:10 — corpus annotation states "84 files"; measured on tree bf779ecf2: `ls internal/web/*_test.go | wc -l` → 82 and `find internal/web -name '*_test.go' | wc -l` → 82 — Severity: minor — Class: optional — Required fix: correct to 82 (or name the command whose definition yields 84). Does not affect the 60/19 superset, which was independently reproduced; but an exact-counts document should not carry an unreproducible count.

D3. spec.md:92 — REQ-WAS-007's pattern label "(Event-detected)" is not one of the five GEARS labels; the requirement TEXT is Event-driven-conformant ("When … shall …"), so MP-2 is unaffected — Severity: minor — Class: optional — Required fix: relabel "(Event-driven)".

D4. REQ-WAS-003 (spec.md:67) / AC-WAS-004 (acceptance.md:42-43) — "or an equivalent already-scoped slice" leaves the run phase one judgment call on what counts as equivalent — Severity: minor — Class: optional — Required fix: if the operator's gate answer keeps the zero-repair path this never binds; otherwise operationalize ("a slice whose bounds are set by existing literal markers in the same receiver").

No blocking defects other than D1. No D7/D8 blocking findings. No prior-iteration defects (iteration 1).

## Regression Check (Iteration 2+ only)

N/A — iteration 1.

## Recommendation

Verdict is FAIL on exactly one ground: the MP-7 clarification gate (D1). The content of the SPEC survived every adversarial angle this audit carries:

1. Route D1 to the operator's gate round now (zero-repair close vs defensive hardening); on the operator's answer, manager-spec folds the decision into plan.md and removes the marker.
2. Apply optional D2 (82 vs 84) and D3 (label) in the same pass — both one-line edits, and both keep the artifact-hash stable until the gate answer lands anyway (D1 already forces a hash change).
3. At run phase, M2 must capture the AC-WAS-005 RED verbatim before any zero-repair close is accepted — the plan already binds this; do not let a green M4 substitute for it.

Rationale for the high score despite FAIL: MP-7 is score-independent by design; the failure names a decision that belongs to the operator, not a defect in the artifacts. Angle A's empty-set interrogation and angle B's highest-risk-class sweep both came back clean, with the corpus completeness and the structural-safety grep verified mechanically on this tree.

---

## Delta Re-audit (Iteration 2/2 — one-item confirmation, scoped to the iter-1 defect delta)

Scope per the Retry Loop Contract: the enumerated iter-1 defect delta (D1 primary; D2/D3 optional fixes reported landed; D4 untouched by design) plus a regression check over those defects. The untouched content (spec.md §B requirement layer, acceptance.md, research.md §2-§8, the classification table) was NOT re-adjudicated from scratch.

### MP-7 result (the one must-pass that failed iter 1)

- **[PASS] MP-7 clarification gate** — Evidence: `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-WEB-ANCHOR-SCOPE-001/` → zero matches (exit 1), this run. The decision is recorded at plan.md:92-106 ("### Gate-round resolution (2026-09-07)": Decision 1 zero-repair close via REQ-WAS-006, hardening NOT taken, REQ-WAS-002 unchanged; Decision 2 run-phase entry APPROVED / Implementation Kickoff passed; Decision 3 autonomous continuous progression) and mirrored at progress.md:34-46 ("### Gate-round record"), with §E.1 `open_clarifications` closed ("none — … RESOLVED at operator gate round 2026-09-07"). Baseline-attribution: this run, this tree (bf779ecf2). Gaps: the gate round itself is an orchestrator-side event — this audit verifies its ARTIFACT RECORD, not the conversation; the coordinator's report of the operator's answers is taken as the routing fact MP-7 requires the record to carry. Residual-risk: none — the record names the decision, the requirement it leaves untouched, and the date.

### Regression check over iter-1 defects

- **D1 (critical, blocking — MP-7 marker)** — [RESOLVED]: marker sweep empty; decision recorded in both plan.md and progress.md; REQ-WAS-002 explicitly unchanged, so no requirement-layer re-adjudication was needed and none found wanting (REQ-WAS-006 remains the zero-TRUE special case of REQ-WAS-002; M3's pre-existing "If the set is empty, close with no code change" text already carries the consequence the resolution paragraph restates — no contradiction introduced).
- **D2 (minor — research.md "84 files")** — [RESOLVED]: research.md:9-11 now reads "82 files — measured on this tree by both `ls internal/web/*_test.go | wc -l` and `find internal/web -name '*_test.go' | wc -l`, each printing 82". The fixed figure matches this auditor's own iter-1 measurement (82/82, both commands) exactly, and the fix carries measurement attribution rather than a bare number.
- **D3 (minor — REQ-WAS-007 label)** — [RESOLVED]: spec.md:92 now "(Event-driven)"; grep confirms both REQ-WAS-006 (line 86) and REQ-WAS-007 (line 92) carry the canonical label.
- **D4 (optional — "equivalent already-scoped slice")** — [NOT TOUCHED / MOOT]: untouched, as expected; on the zero-repair path the latitude never binds. Remains a non-binding observation only.

### New-defect sweep over the touched sections

- plan.md Gate-round resolution section: internally consistent with §F milestones (M3 body unchanged and already correct), §G anti-patterns, and the requirement layer (untouched). One cosmetic note, NOT a defect: the resolution H3 sits under "## H. Cross-References" topically rather than as its own section — plan.md carries no heading-convention binding, no fix required.
- research.md §1: figure fix only; the 60/19 measurement block and tree pin are byte-preserved.
- spec.md: label fix only; requirement text untouched.
- acceptance.md: spot-checked unchanged — AC-WAS-005 (acceptance.md:47-56) byte-identical to iter 1, as claimed.
- progress.md: gaps list retains both honest entries (no go build/vet at plan phase; mutant RED pending M2) — the fold softened nothing.

### Delta verdict

**PASS — final score 0.99.** All seven must-pass criteria now PASS (MP-1..MP-6 carried over from iter 1 with evidence unchanged; MP-7 cleared above). Category scores: Clarity 1.00, Completeness 1.00 (D2 fixed with attribution), Testability 0.95 (D4 latitude remains, moot on the zero-repair path), Traceability 1.00. Tier M PASS threshold 0.80 exceeded; score progression 0.97 → 0.99 (no regression). Standing obligations carried into run phase, unchanged from iter 1: M2 must capture the AC-WAS-005 RED verbatim before any zero-repair close is accepted, and M1 must re-run the superset measurement on the run-phase tree (plan.md §C.3).

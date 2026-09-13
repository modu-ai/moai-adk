# Plan-Audit Report — SPEC-GATEWAY-ENVELOPE-REPAIR-001 (card t708)

- **FINAL VERDICT (iteration 2): PASS — score 0.975** (see §9)
- Iteration: 2/2 (Tier M ceiling per `harness.yaml` `plan_audit_tier_ceilings` S=1/M=2/L=3); iteration-1 verdict was COND-FAIL 0.8375
- Auditor: plan-auditor (in-session; no `audit_model: multi` key found in `.moai/config/` — the `workflow.yaml` `audit:` block carries backend model pins only, not the audit-mode selector, so per the audit contract this audit ran in-session without a cross-backend fan-out)
- Tree: worktree `.claude/worktrees/t708`, branch `WT-envelope-persist`, HEAD `7a7a08f20` (re-verified at audit time)
- Verdict: **COND-FAIL**
- Overall score: **0.8375** (Tier M PASS threshold 0.80 — score clears; two blocking-class artifact defects and the MP-7 clarification gate force the conditional verdict)

Reasoning context from the SPEC author was not supplied; M1 Context Isolation holds trivially. All findings below are from the artifacts + source tree only.

## 1. Security-determination verification (the load-bearing check — task attention point 1)

The §3 byte-exact re-injection argument was verified against the actual code, claim by claim. **It holds.** Evidence:

| SPEC claim (spec.md §3.2) | Code verification (this tree, HEAD `7a7a08f20`) | Result |
|---|---|---|
| Prefix excludes thinking blocks | `projection.go:92-93` `case "thinking", "redacted_thinking": continue` inside `normalizeContentPosition`, consumed by `CanonicalPrefixes` (projection.go:22-68); doc "It does not authenticate thinking blocks" (projection.go:19) | CONFIRMED — re-injection changes neither Prefix nor Previous (Previous chains prefix snapshots, projection.go:54-65) |
| Opaque digest is canonical/content-addressed | `Digest()` = `sha256.Sum256(e.raw)` hex (codec.go:80-86); `Decode` re-encodes via `EncodeWithPublic` and rejects alternate spellings `if !bytes.Equal(e.raw, raw)` (codec.go:217-220) | CONFIRMED — exactly one byte sequence per envelope passes |
| Marker self-attestation discloses the required digest | `BindToolID` embeds `{call_id, opaque_sha256}` (codec.go:292-304); `RestoreToolID` refuses `b.Digest != e.Digest()` and non-canonical marker re-encoding (codec.go:307-329) | CONFIRMED — pre-injection gate needs no receipt-store read; REQ-EVR-003-4 is sound (research C5 resolved in the correct direction) |
| Unchanged Check adjudicates the full binding | `Manifest.Check` four-way loop, empty fallback only `found = empty && !required` (core.go:103-124); `valid()` Required ⇔ Items>0 ∧ Opaque≠0 (core.go:75-77); `Publish` sets `Required: last.Items > 0` (receipt_history.go:255) | CONFIRMED — the no-bypass consequence in §3.2 is mechanically true: a repaired replay passes only on observations the gateway itself published |
| t703 firing site + classification | `receipt_history.go:152-158` `strings.HasPrefix(id, opaque.ToolPrefix)` → `classify(CauseReasoning, receipt.ErrInvalid)`; frozen `historyReplayGuidance` (receipt_history.go:40); envelope issuance `content[0]` (response.go:56-59) + stream terminal full key-copy (stream.go:425-429); forward-path splices request.go:300/360/365 | CONFIRMED — matches t703 verdict and research F1-F4 |

Worst-case abuse bound re-derived independently: cross-boundary transplant fails Opaque equality at the wrong boundary's marker; reorder/public-edit fails Prefix/Previous; manufactured bytes fail both marker self-attestation and Decode canonicality. The repair adds a procedure, not an authorization. **No claimed property is contradicted by the code.**

## 2. Forbidden-direction compliance (attention point 2) — PASS

- Validator relaxation: REQ-EVR-001 (spec.md:63) + plan.md §D PRESERVE + AC-EVR-011 file-level zero-diff. Not smuggled.
- Request-driven receipt-root selection: REQ-EVR-009 (spec.md:87) + Out of Scope "request-driven gateway state" + AC-EVR-010 (no trigger reads request metadata). Not smuggled.
- Envelope manufacture/synthesis: REQ-EVR-002 (spec.md:66, "never mint, re-encode, re-serialize, wrap, or synthesize") + Out of Scope "envelope manufacture" + plan.md §G "normalizing" anti-pattern. Not smuggled.
- Automatic surgery: REQ-EVR-005 + AC-EVR-006 (both branches tested). Not smuggled.
- Receipt-store read from repair path: forbidden (REQ-EVR-003-4, AC-EVR-010) — and the no-read direction is technically sound per §1 above.

## 3. Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-EVR-001..010 sequential, zero-padded, no gaps/duplicates (spec.md:63-89). AC-EVR-001..012 sequential (acceptance.md:21-71). Within Tier M ceilings (10 REQ ≤ 16; 12 AC ≤ 16).
- **[PASS] MP-2 GEARS format compliance (requirement layer)** — all 10 REQs match GEARS patterns: Ubiquitous shall/shall-not (REQ-001/002/005/009/010), Where (003/008), When (004/007), compound [Where][While][When] (006, PASS-equivalent per the compound clause). Judged against `spec.md` §2.2 requirement entries; the Given-When-Then entries in acceptance.md §D are the verification layer and were not GEARS-tested (correct layer separation per M3 § Scope). Cosmetic only: REQ-EVR-007 is labeled "Event-detected" — structurally a valid When pattern.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types (spec.md:2-14); no snake_case aliases; `phase: "v3.2.0 target"` is a release-target label (not a prohibited stage name); sibling artifacts carry no `status:` field (artifact statelessness respected; the SPEC-dir `research.md` thin pointer omits frontmatter entirely — permitted).
- **[N/A] MP-4 language neutrality** — single-language (Go) SPEC; `module:` names only Go packages. Auto-pass.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — refs extracted: SPEC-GATEWAY-WEDGE-REROOT-001, SPEC-MOAI-GATEWAY-001. Neither is retired/superseded/archived in reach: SPEC-MOAI-GATEWAY-001 exists here with `status: draft`; t700's SPEC is unlanded (draft, sibling worktree) and deliberately referenced by path with its status disclosed (spec.md:53, :118) plus the M0 adjudication gate (plan.md:96-97). The D7-5 SHOULD finding for the unlanded sibling is discharged by the SPEC's own explicit disclosure + gate — no BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `syscall` appears nowhere in spec.md/plan.md/acceptance.md (grep rc=1); plan.md §D carries the GOOS=windows build check. Auto-PASS (D8-4).
- **[FAIL — letter] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → 3 hits, all in plan.md §H (plan.md:127-130): 2 topic markers + the §H heading. **Adjudication:** both markers are convention-compliant — confined to plan.md (never spec.md/acceptance.md, per the SKILL.md placement rule), each carries an explicit boundedness argument (design-invariant either way via the REQ-EVR-007 refusal path) and a resolution route. Per the marker convention the remediation is orchestrator-side, not a SPEC defect: the orchestrator MUST resolve each topic (AskUserQuestion) before Implementation Kickoff Approval, and acceptance.md §D.4-3 requires disposition before M3 exit. Folded into Defects as G1 (critical severity, gate-class — no manager-spec artifact repair required; the markers are the designed carrier for these two open questions). M1/M2 are explicitly non-blocking (plan.md §H heading).

## 4. Category Scores

| Dimension | Score | Rubric band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | Minor ambiguity, resolvable consistently | REQ-EVR-003-1 "left untouched" (spec.md:70) vs REQ-EVR-007 "refuse with zero modification" (spec.md:82) leaves partial-vs-total repair ambiguous at the requirement layer; AC-EVR-005 + §D.2 all-or-nothing resolve it, but a security REQ should not need the AC to disambiguate (D4). Section ordering §3.4-after-§4 (spec.md:126 vs :116) slightly impairs scan (D3). |
| Completeness | 0.90 | All sections + frontmatter complete | HISTORY (spec.md:22-26), problem/WHY (§1), requirements (§2), security determination (§3), seam (§4), shapes table (§3.4), Out of Scope with 6 specific H3 `### Out of Scope — <topic>` blocks each carrying `-` bullets (spec.md:138-165). Deduction for the §3.4 misplacement (D3). |
| Testability | 0.95 | Binary-testable | All 12 ACs Given-When-Then with concrete evidence commands (acceptance.md:21-71); weasel-word scan clean; AC-EVR-010's "+ review" rider is the only soft edge (primary gate is the scripted grep). |
| Traceability | 0.75 | One AC unmapped | §D.1 matrix (acceptance.md:75-86) omits AC-EVR-003 entirely while asserting "every AC traces to ≥1 REQ. No orphan rows" (acceptance.md:88) — false as written (D1). Every REQ row has ≥1 AC; plan.md M2 shows the intended mapping (AC-EVR-001..003 under REQ-EVR-001), so the fix is a one-cell insert. |

**Aggregate: (0.75 + 0.90 + 0.95 + 0.75) / 4 = 0.8375 ≥ 0.80** (Tier M threshold).

## 5. Remaining audit-stance checks (attention points 3-5)

- **Merge-order/absorb discipline vs t700** — present and concrete: spec.md §1.4 (symbol pins, never line numbers; re-verify on absorbed tree), plan.md §A (absorb local `develop`, the only sanctioned target, measured at plan time `7a7a08f20`; t700 landing check; blocker path via D-NEW-1), plan.md §F M1. Pin discipline verified: no line-number pins on code anywhere in spec.md/plan.md.
- **M0 seam gate with t700 present in plan.md** — present (plan.md:96-97): surfaces spec.md §4 Reading A/B to the orchestrator, blocks M3+ on a Reading-A adjudication, records when t700 remains unlanded. Matches acceptance.md §D.4-4 and §D.5.
- **t700 seam not silently violated (attention point 5)** — spec.md §4 quotes t700 REQ-WRR-007/008(d) accurately (verified against `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t700/.moai/specs/SPEC-GATEWAY-WEDGE-REROOT-001/spec.md:81,83,107,116`), presents both readings, adopts Reading B explicitly, binds itself to the common core both readings forbid (verbatim gateway-issued bytes only, REQ-EVR-002), does not amend t700, and gates run entry. Reading B does not violate t700's intent: (d)'s rationale bans manufactured envelopes ("must always reference something the gateway itself issued") and (d)'s scope is gateway-state-changing recoveries, which this repair is not; REQ-WRR-007's rejection mandate binds replays that *carry* stripped envelopes, which a repaired replay no longer is. The seam is surfaced, not silently resolved.
- **NEEDS CLARIFICATION boundedness (attention point 3)** — see MP-7: both markers bounded with refusal paths; M1/M2 non-blocking; disposition gate at M3 exit.
- **No time estimates** — verified; plan.md §F uses Priority labels only. Milestones M0-M6 are realistic and decision-ordered; TDD methodology with RED evidence required (§E E8).
- **Research ↔ spec ↔ acceptance consistency** — research F1-F10/C1-C7 map onto §1/§3/§4; contradiction C5 resolved by REQ-EVR-003-4 (declared in plan.md §A); C4 surfaced as §4 (not averaged, per the research's own instruction); the 37-carrier addendum is the §1.1 content-source claim; §D.2 covers the subagent-row edge the addendum surfaced. One numeric imprecision: spec.md:32 says "37 complete `moai_opaque_v2_*` carriers retained in the main transcript" while the addendum measured 36 in the main transcript by grep (37 = including a subagent-row boundary) — D5.

## 6. Defects Found

- **D1** — acceptance.md:75-88 (§D.1) — AC-EVR-003 ("Refusal shapes stay rejected", acceptance.md:29-31) is absent from the REQ→AC traceability matrix, contradicting the matrix's own closing claim "every AC traces to ≥1 REQ. No orphan rows" (acceptance.md:88). plan.md M2 (plan.md:103) already groups AC-EVR-001..003 under REQ-EVR-001, so the intended mapping exists outside the SSOT matrix ("acceptance.md §D is the SSOT", plan.md:85). — Severity: major — Class: blocking — Required fix: insert AC-EVR-003 into REQ-EVR-001's row: `| REQ-EVR-001 | AC-EVR-001, AC-EVR-002, AC-EVR-003, AC-EVR-011 | MUST |`.
- **D2** — plan.md:57 (§C pre-flight) — `grep -n "func refreshNative" internal/gateway/conversation/native.go` matches nothing: `refreshNative` is a method, declared `func (m *Manager) refreshNative(ctx context.Context, id string) error` (native.go:17). Empirically executed at audit time: rc=1, no output. Since §A rules "If any pin broke, return a blocker report", M1 dead-ends deterministically on a false pin-break. All six other §C commands verified green empirically (go doc ×2 rc=0; greps 5/1/3 hits). — Severity: major — Class: blocking — Required fix: change the pattern to `grep -n "func (m \*Manager) refreshNative" internal/gateway/conversation/native.go`.
- **D3** — spec.md:126 — `## 3.4 Recoverable vs non-recoverable shapes` appears after `## 4. Sibling-SPEC interpretive seam` (spec.md:116), so the H2 scan reads §1 → §2 → §3 → §4 → §3.4 → Out of Scope. — Severity: minor — Class: optional — Required fix (recommended in the same pass): relocate the §3.4 table before §4 (it is a §3 security-determination subsection).
- **D4** — spec.md:70 vs :82 — REQ-EVR-003-1 ("a boundary without a self-attesting surviving marker is left untouched") vs REQ-EVR-007 ("refuse with zero modification") leave the multi-boundary behavior forked at the requirement layer (partial repair of other boundaries vs whole-attempt refusal). AC-EVR-005 ("aborts the repair") and §D.2 ("all-or-nothing per attempt") take the refusal side. — Severity: minor — Class: blocking — Required fix: one clause in REQ-EVR-003-1, e.g. "a boundary without a self-attesting surviving marker aborts the entire repair with zero modification (REQ-EVR-007)".
- **D5** — spec.md:32 vs `.moai/reports/t708/research.md:110` — "37 complete `moai_opaque_v2_*` carriers retained in the main transcript" overstates the main-transcript grep (36); 37 is the walked count including a subagent-row boundary. No design decision changes. — Severity: minor — Class: optional — Required fix: "36 in the main transcript (37 including a subagent-row boundary), measured in the addendum".
- **G1 (MP-7 clarification gate)** — plan.md:127-130 — 2 bounded `[NEEDS CLARIFICATION]` markers present at audit time. Convention-compliant placement and boundedness; each has a refusal path (REQ-EVR-007) so it cannot block M1/M2. — Severity: critical (per MP-7 folding rule) — Class: blocking for run-phase ENTRY only — Required action (orchestrator, not manager-spec): resolve both topics via AskUserQuestion before Implementation Kickoff Approval, or record the disposition per acceptance.md §D.4-3 before M3 exit.

## 7. Regression Check

N/A — iteration 1.

## 8. Recommendation

**COND-FAIL.** The design and the security determination are sound (§1: every load-bearing code claim verified on this tree; §2: no forbidden direction present; the t700 seam is surfaced and gated, not silently resolved). The score clears the Tier M threshold. The verdict is conditional on two one-pass artifact repairs (D1, D2) plus the two cheap consistency fixes (D3, D4; D5 optional) — each is a single-line edit with the exact target given above. After the repairs, a delta re-audit (iteration 2) scoped to D1-D5 is sufficient; a from-scratch re-audit is not required. G1 transfers to the orchestrator: the two §H clarification topics must be resolved (or dispositioned) before Implementation Kickoff Approval and before M3 exit respectively.

---

# Iteration 2 — Delta Re-Audit (final)

- Date: 2026-09-13 · Tree: `.claude/worktrees/t708` @ `7a7a08f20` (re-verified; unchanged from iteration 1, so all iteration-1 code verifications remain attributed to this same tree)
- Scope: defect delta D1-D5 from the iteration-1 report + the new lead input (t707 serial-turn chain-class fold-in), per the Retry Loop Contract. §3 was not re-litigated; sanity spot-check only (§3.1-3.3 byte-identical to the audited iteration-1 text; no re-verification required since the tree SHA is unchanged).
- **Verdict: PASS** · **Score: 0.975** (Clarity 1.0 / Completeness 0.95 / Testability 0.95 / Traceability 1.0) — Tier M threshold 0.80 cleared; no score regression (0.8375 → 0.975, no STOP signal).

## Regression check over iteration-1 defects — all five RESOLVED (observed, not asserted)

| Defect | Disposition | Evidence (this run, this tree) |
|---|---|---|
| D1 (AC-EVR-003 matrix orphan) | **RESOLVED** | acceptance.md §D.1 REQ-EVR-001 row now reads `AC-EVR-001, AC-EVR-002, AC-EVR-003, AC-EVR-011`; all 12 ACs mapped; the "No orphan rows" claim is now true as written. |
| D2 (broken refreshNative pre-flight grep) | **RESOLVED** — command executed | `grep -n "func (m \*Manager) refreshNative" internal/gateway/conversation/native.go` → rc=0, matches native.go:17 (and :73 `refreshNativeProject`, same prefix — both confirmations, no false break). |
| D3 (§3.4 after §4) | **RESOLVED** | §3.4 now at spec.md:117, before §4 (spec.md:130). H-scan order: §1 → §2 → §3 → §3.4 → §4 → §4.1 → Out of Scope. |
| D4 (partial-vs-total refusal ambiguity) | **RESOLVED** | spec.md:71: "a boundary without a self-attesting surviving marker **aborts the entire repair with zero modification** (all-or-nothing per attempt, per REQ-EVR-007), never a partial repair of the remaining boundaries" — aligned with AC-EVR-005 and §D.2. |
| D5 (carrier-count citation) | **RESOLVED** | spec.md:33: "36 `moai_opaque_v2_*` carriers in the main transcript (37 including a subagent-row boundary) ... per the research addendum measurement" — matches `.moai/reports/t708/research.md:110`. |

## New input fold-in (t707 serial-turn chain-class) — verified clean

- **No silent narrowing.** REQ-EVR-004's trigger was CauseReasoning-classified from iteration 1 (spec.md:76); the new Out of Scope block (spec.md:169-173) and §3.4 row (spec.md:125) restate that existing boundary for the CauseChain serial-turn shape (family `1f14d174`) — they remove coverage no REQ ever promised. REQ-EVR-001/007 and the user story are untouched; the block is additive alongside the pre-existing "unrelated rejection axes" exclusion, with a distinct owner (card t707) and no contradiction.
- **§4.1 epistemics correct** (spec.md:140-142): litellm #40288 recorded as external prior art with "No applicability claim is made" — the digest-stability question is explicitly deferred to t707's reproduction.
- **Third §H marker bounded** (plan.md:131): confined to plan.md (grep: 0 hits in spec.md/acceptance.md/research.md/progress.md; 3 hits all in plan.md §H); bounded — "changes only scope allocation, never the design" via REQ-EVR-004's CauseReasoning scope; resolution path = t707's verdict; M1/M2 non-blocking per the §H heading.
- **M0 is now a two-input adjudication gate** (plan.md:96-97: t700 seam + t707 verdict; blocker → re-delegation on a Reading-A adjudication or an adverse t707 scope finding) with the matching §A step-2 re-read (plan.md:18) and §I cross-refs (plan.md:140-141).
- **Version + HISTORY**: spec.md:4 `version: "0.2.0"` (quoted semver); HISTORY 0.2.0 row (spec.md:27) accurately documents all five repairs + the new input, citing this report.

## Must-pass re-check (delta scope)

MP-1 PASS (REQ numbering unchanged, 001-010 sequential). MP-2 PASS (REQ-EVR-003-1 edit preserves the Where pattern; requirement layer still fully GEARS-valid). MP-3 PASS (version bump to `"0.2.0"` is a valid quoted semver; 12 canonical fields intact). MP-4 N/A. MP-5 PASS (no new retired/superseded/archived SPEC reference; t707 is a card id, litellm #40288 is external — neither is a D7 subject). MP-6 PASS (no `syscall` in the new text). MP-7: 3 bounded markers, same adjudication as iteration 1 — carried as gate-class finding G1 (below), not an artifact defect.

## Residual items (none blocking)

- **G1 (carried, gate-class — orchestrator obligation)**: 3 bounded `[NEEDS CLARIFICATION]` markers in plan.md §H. Resolve via AskUserQuestion before Implementation Kickoff Approval, or record the disposition per acceptance.md §D.4-3 before M3 exit. No manager-spec artifact repair required.
- **O1 (optional, minor)**: plan.md §F M0's t700 branch spells out the "remains unlanded → record no conflict" outcome, but the t707 branch lacks the symmetric "no verdict yet → record status; §H marker 3 stays open for §H resolution" sentence. §A step 2 ("re-read card t707's verdict status") plus the §H heading already own the undecided path, so this is a wording-completeness nit, not a gap; orchestrator discretion whether to fold into a future touch.

## Recommendation

**PASS.** Iteration 2 of 2 (Tier M ceiling). The SPEC at v0.2.0 is audit-ready: the security determination was verified sound against the code in iteration 1 and is unchanged; all five iteration-1 defects were repaired exactly as prescribed; the t707 input was folded in without narrowing any requirement's coverage. Run-phase entry is subject to the standing gates, not to any further plan-audit iteration: Implementation Kickoff Approval (mandatory human gate), the G1 clarification resolution (3 markers), and M0's two-input adjudication record before M3+.

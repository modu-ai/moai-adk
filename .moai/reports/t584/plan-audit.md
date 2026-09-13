# SPEC Review Report: SPEC-AUT-PERMMODES-001 (card t584)

Iteration: 1/2 (Tier M ceiling = 2, per `harness.plan_audit_tier_ceilings`)
Verdict: FAIL
Overall Score: 0.75 (Tier M PASS threshold: 0.80)

Auditor: plan-auditor (independent). Tree: worktree `.claude/worktrees/t584`, branch `WT-autonomy-perm-modes`, HEAD `0ddd1a282`. All commands run in this worktree against this tree unless noted.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk (5-section format)

**Claim.** The SPEC's plan-phase artifacts are internally sound on every must-pass axis; the single substantive defect is an acceptance-layer coverage gap: REQ-009 and REQ-010 have no AC, and acceptance.md §D.2's stated substitute verification route is factually wrong for REQ-009. Aggregate 0.75 < 0.80 → FAIL.

**Evidence.** Every verdict below cites a command run in this session against HEAD `0ddd1a282` of this worktree; verbatim outputs are quoted in the Must-Pass and Category sections. Named-path verification batch (all `grep -rln "func <name>(" internal/` in this tree):

| Test named by plan/acceptance | Existence result |
|---|---|
| `TestApplyAutonomyTierBundle_FullyAutonomousDowngradedWithoutProof` | EXISTS — `internal/core/project/autonomy_bundle_test.go` |
| `TestApplyAutonomyTierBundle_FullyAutonomousWithProofDeploysBypass` | EXISTS — `internal/core/project/autonomy_bundle_test.go` |
| `TestRunInit_FlagFullyAutonomousWithoutProofDowngrades` | EXISTS — `internal/cli/init_autonomy_wiring_test.go` |
| `TestAppendDowngradeAdvisory` | EXISTS — `internal/config/autonomy_tiers_toggle_test.go` |
| `TestAutonomyTierQuestion_FullyAutonomousNotRecommended` | EXISTS — `internal/cli/wizard/autonomy_test.go` |
| `TestApplyAutonomyTierBundle_SemiAutoIsZeroDelta` | EXISTS — `autonomy_bundle_test.go:89` (plan's :89 exact) |
| `TestApplyAutonomyTierBundle_EmptyIsZeroDelta` | EXISTS — `autonomy_bundle_test.go:115` (plan's :115 exact) |
| `TestRunInit_SemiAutoAndEmptyAreZeroDelta` | EXISTS — `internal/cli/init_autonomy_wiring_test.go:193` (plan's :193 exact) |

**Baseline-attribution.** HEAD `0ddd1a282` (`feat(SPEC-AUT-PERMMODES-001): plan-phase artifacts (M, 3 artifacts)`), worktree `.claude/worktrees/t584`, 2026-09-13. Working tree clean at audit start.

**Gaps.** (1) I could NOT independently fetch `https://code.claude.com/docs/en/permissions` — this subagent surface has no WebFetch tool. The six-value-enum / `auto`-validity claim is therefore ACCEPTED ON THE PLAN'S DATED CITATION (fetched 2026-09-13, recorded in progress.md §E.1), not independently confirmed. The stale-IAM-reference half of the claim I did verify directly (see MP-3 evidence block). (2) No test execution was run (plan-phase audit; run-phase §E owns that). (3) `.moai/reports/init-tui-audit-20260909.md` exists in the primary checkout (verified `ls`, 10853 bytes, 2026-09-09) but is local-only and was not re-read section-by-section; the card C2 provenance claim is taken as pointed-to, not re-adjudicated.

**Residual-risk.** If the official enum claim is wrong (e.g., `auto` gated behind a plan tier), plan M1's exit gate correctly mandates a blocker — the SPEC self-protects. If the operator's "acceptEdits default" decision itself is later revised, REQ-002/REQ-004 wording would need re-scoping (out of audit scope).

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-001..REQ-010 sequential, no gaps, no duplicates (spec.md:58-100, one `### REQ-0NN` per H3, zero-padded 3-digit). Evidence: read of spec.md §B.
- **[PASS] MP-2 GEARS format compliance (requirement layer)** — Every REQ-XXX matches the GEARS Ubiquitous pattern "The <subject> shall SHALL …" or an event/state variant: "The init wizard autonomy question SHALL present…", "The wizard question SHALL pre-select…", "`TierDefaultMode` SHALL map…", "…SHALL write ONLY the USER-scope…", "The fully-autonomous gating SHALL remain anchored…". No informal "should/may" in normative text. **Layer judged: requirement layer (spec.md §B REQ-001..010) only; the Given-When-Then entries in acceptance.md §D are the verification layer and are graded under Group 4, not here.** Note: REQ-003/REQ-005 name internal functions as subjects — an RQ-4 style concern scored under Clarity, not an MP-2 pattern failure.
- **[PASS] MP-3 YAML frontmatter validity** — All 12 canonical fields present (spec.md:2-13): id/title/version/status/created/updated/author/priority/phase/module/lifecycle/tags, plus optional `tier: M` and `related_specs`. No rejected snake_case aliases (`created_at`/`updated_at`/`labels`/`spec_id` — grep clean). `status: draft` (valid enum), `priority: P1` (valid), `phase: "v3.2.1 target"` (release-target label, not a prohibited stage name), `lifecycle: spec-anchored`. Cosmetic deviation noted as D3: `version: 0.1.0` is unquoted — it decodes to the string `"0.1.0"` (YAML cannot parse two-dot 0.1.0 as a number), so no decoder-level type mismatch; lint-neutral.
- **[N/A] MP-4 Section 22 language neutrality** — Single-language (Go) project SPEC; no multi-language tooling enumeration obligation. N/A auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — Verification verb executed in this session: all four referenced SPECs exist in `.moai/specs/` of this tree with `status: completed` — SPEC-AUTONOMY-TIERS-001, SPEC-INIT-WIZARD-REPAIR-001, SPEC-CLI-WIZARD-RESTRUCTURE-001, SPEC-INIT-TUX-I18N-001. None is retired/superseded/archived → no reconciliation obligation, no BLOCKING finding. One SHOULD-grade citation imprecision folded into D4 (the owning-SPEC line pointers name AC rows, not REQ rows).
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall'` on spec.md = **0**. D8 auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION'` over `.moai/specs/SPEC-AUT-PERMMODES-001/` = no matches. `research.md` absent (Tier M — correct artifact set of 3); `plan.md` clean. No BLOCKING.

## Card-Obligation Audit (dispatch items 1a–1d)

- **1(a) REQ-007 re-scope honest about the new default WRITING user-scope defaultMode** — SATISFIED. spec.md:76 (REQ-004): "the unset / `semi-auto` selection SHALL write ONLY the USER-scope `defaultMode: "acceptEdits"` record and NOTHING else … The sanctioned delta is exactly one JSON key in exactly one file." The re-scope is explicit ("RE-SCOPED", not silently broken) and matches the card. AC-005 (acceptance.md:13) tests the bounded delta with full-file byte comparison.
- **1(b) REQ-006 + sandbox proof + kill switch anchored to bypassPermissions, auto-downgrade + log preserved** — SATISFIED. spec.md:80 (REQ-005) anchors gating to `bypassPermissions`, downgrade target `automatic` (→ `auto`), advisory sink `.moai/logs/autonomy-downgrade.log`. Tree-verified: `config.EffectiveTierWithGates` at `internal/config/autonomy_tiers.go:170`, `IsBypassDisabled` :151, `AppendDowngradeAdvisory` :215, log sink written at `internal/core/project/autonomy_bundle.go:75`. AC-006/AC-007 cover both gate arms (no-proof; kill-switch-trumps-proof).
- **1(c) defaultMode token validation as M1 exit gate** — SATISFIED. plan.md §F M1 exit: "if the floor research contradicts the six-value enum … STOP and return a blocker report" — a blocker path, not a workaround. Known Issue B3 names the unverified version floor honestly.
- **1(d) downgrade regression tests preserved by name/path** — SATISFIED. All five REQ-006 test names exist on this tree (table above), in three packages; AC-008 requires them to pass under the new option set; plan M3.2 preserves them "verbatim in intent".
- **2. Internal consistency** — spec REQ ↔ AC cross-map correct for REQ-001..008 (acceptance.md §D.2); Tier M caps respected (10 REQ ≤ 16, 10 AC ≤ 16); frontmatter canonical; Out of Scope is `## §D. Out of Scope` parent + four `### Out of Scope — <topic>` H3s each with `-` bullets (spec.md:108-124) — the MissingExclusions trap is avoided. **One defect: REQ-009/REQ-010 have no AC (see D1).**
- **3. Feasibility against the tree** — All named touchpoints exist and line anchors are EXACT: `internal/cli/wizard/questions.go`, `translations.go:98/184/270` (autonomy_tier keys, ko/en/ja/zh — verified), `internal/config/autonomy_tiers.go` (`TierDefaultMode` at :78), `internal/core/project/autonomy_bundle.go` (func at :58, semi-auto early-return `return nil` verified), `internal/cli/init.go:126` (flag help string verified verbatim), `internal/config/defaults.go:161-163` (three tier tokens verified verbatim), IAM reference `:30-40` ("accepts exactly four values" + explicit `auto` denial — verified, confirming the staleness claim), `internal/cli/update_settings_snapshot.go` exists. No named-but-absent file.
- **4. Scope discipline** — SATISFIED. Web handlers appear only as plan §C pre-flight checks with "Expected outcome: NO web edit (Out of Scope)"; tree confirms `/autonomy/tiers` route already removed (app.go comment, handlers.go:200-204) so no compile-break risk. Question-count freeze ("Do NOT re-expand the question count", §D constraint) preserves the t583 4-question baseline. No edit target outside the card blast radius.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | Requirements are precise, measurable, single-interpretation, no pronoun ambiguity. Deduction: REQ-003 (spec.md:72) and REQ-005 (spec.md:80) name internal functions (`TierDefaultMode`, `EffectiveTierWithGates`) as requirement subjects — implementation leakage in the requirement layer (RQ-4), resolvable consistently by an engineer, hence not lower. |
| Completeness | 0.75 | 0.75 | All substantive sections present: §A Background (WHY), §B Requirements, §C Design Notes, §D Out of Scope (4× H3 + bullets), §H Cross-References; AC layer complete in acceptance.md; frontmatter 12/12 + tier. Deduction: no HISTORY section (spec.md has no version-history record). |
| Testability | 1.0 | 1.0 | AC-001..AC-010 are Given-When-Then and binary-testable (acceptance.md:9-18). No weasel words ("appropriate/adequate/reasonable") found in any AC. AC-005's "byte-identical" is defined operationally in §D.3 (full-file byte comparison). |
| Traceability | 0.50 | 0.50 | AC→REQ direction clean (every AC references an existing REQ). REQ→AC direction: REQ-009 and REQ-010 have NO AC. acceptance.md:27 claims they are "verified by §E.4 grep guards (E4)", but E4 (plan.md:36) greps **wizard sources** for retired label strings — it neither touches `internal/cli/init.go:126` flag help (REQ-009) nor verifies the new godoc content (REQ-010). The stated substitute route is factually wrong for at least REQ-009. |

**Overall Score: (0.75 + 0.75 + 1.0 + 0.50) / 4 = 0.75** → below Tier M threshold 0.80 → FAIL.

## Defects Found

D1. **TRACE-001** — acceptance.md:27 (§D.2) + plan.md:36 (§E.4) — REQ-009 (`--autonomy-tier` flag help text) and REQ-010 (godoc on `TierDefaultMode`/`ApplyAutonomyTierBundle`) have no AC; the declared E4 grep-guard substitute does not reach REQ-009's surface (init.go is not a wizard source and the grepped strings are option labels, not flag help) and only detects stale — never asserts new — godoc for REQ-010. The acceptance layer's own stated criterion ("verified by §E.4 grep guards") is false as written for REQ-009. — Severity: major — Class: **blocking** — Required fix: EITHER add AC-011 (Given the flag help, When `--help` output is read, Then it describes the three permission-mode choices and the acceptEdits default) and AC-012 (Given the two functions' godoc, When read, Then the new mapping and the re-scoped REQ-004 invariant are stated), and update §D.2 + §D.7 AC-count accordingly; OR (minimal) extend E4 to name `internal/cli/init.go:126` (grep for the retired "Autonomy tier: semi-auto, automatic, or fully-autonomous (default: semi-auto)" string) and to positively grep the two godoc sites for the new mapping, then correct §D.2's claim to name exactly what E4 greps. Correcting §D.2 alone without extending E4 is NOT sufficient — it would leave both REQs with no executable verification.

D2. **STRUCT-001** — spec.md (whole file) — No HISTORY section recording the SPEC's own revision trail. — Severity: minor — Class: optional — Required fix: add a `## HISTORY` section (v0.1.0 initial, card t584 provenance) before or after §A.

D3. **FM-001** — spec.md:4 — `version: 0.1.0` unquoted; the canonical schema shows the quoted form (`version: "0.1.0"`). Decodes to the same string, lint-neutral today. — Severity: minor — Class: optional — Required fix: quote the value.

D4. **XREF-001** — spec.md:128 (§H) — Cites "REQ-006 at spec.md:167, REQ-007 at spec.md:168" of SPEC-AUTONOMY-TIERS-001; those lines hold the AC entries (`AC-AUTONOMY-TIERS-006/007`), not the `REQ-XXX` entries. The quoted invariant CONTENT is accurate; the line labels are off. — Severity: minor — Class: optional — Required fix: relabel to "AC-AUTONOMY-TIERS-006 at :167, AC-AUTONOMY-TIERS-007 at :168" (or re-locate the REQ rows and cite those).

D5. **REQ-STYLE-001** — spec.md:72, :80 — REQ-003/REQ-005 carry internal function names as requirement subjects (`TierDefaultMode`, `EffectiveTierWithGates`). For a config-mapping card the mapping semantics are legitimately the WHAT, so this is style, not correctness. — Severity: minor — Class: optional — Required fix: rephrase subjects to the component/behavior ("The tier-to-mode mapping SHALL …", "The fully-autonomous gating logic SHALL …") and keep the function names in §H cross-references.

## Regression Check

Iteration 1 — no prior-iteration defects. N/A.

## Recommendation

Verdict FAIL on one blocking defect (D1) dragging Traceability to 0.50; overall 0.75 < 0.80. Fix route for manager-spec (scoped to the enumerated defect delta, per the Retry Loop Contract):

1. (D1, blocking) Add AC-011/AC-012 per the required-fix above — or minimally extend plan §E.4's greps to the flag-help string at `internal/cli/init.go:126` and to positive assertions on the two godoc sites, then rewrite acceptance.md §D.2 so its claim matches what E4 actually greps. Update §D.7's AC list if ACs are added. Recheck REQ/AC budget stays ≤ 16 (12 ACs still fine).
2. (Optional, cheap while editing) D2 HISTORY section; D3 quote `version:`; D4 fix the owning-SPEC line labels; D5 de-name the two requirement subjects.
3. Do NOT touch anything else — the requirement bodies, the M1 blocker gate, the test-preservation list, and the Out of Scope set are sound and must survive revision unchanged.

On re-audit (iteration 2), scope is the delta above plus regression over D1-D5. If D1 is resolved and D2-D5 are left untouched or fixed, the floor for the re-audit aggregate is 0.8125 (Traceability ≥ 0.75) → PASS band.

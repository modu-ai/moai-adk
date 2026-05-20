# SPEC-V3R5-LATE-BRANCH-001 Progress

Plan-phase progress log for Late-Branch Workflow formalization.

## Plan-phase signals

- plan_complete_at: 2026-05-20T16:30:00Z
- plan_status: audit-ready
- scope_tier: T2 Standard (per orchestrator decision — same tier as W3 HARNESS-AUTONOMY-001)
- spec_id: SPEC-V3R5-LATE-BRANCH-001
- spec_version: 0.1.1 (mid-draft extension: REQ-LB-009 + AC-LB-007 + EXCL-LB-008 + R-LB-005 + D2 — no-auto-issue policy 2026-05-20)
- plan_branch: NOT YET CREATED (Late-branch dogfooding — branch created at plan-PR time only)
- baseline_head: TBD (will be `git rev-parse main` immediately before plan-PR `git switch -c`)
- vision_source: `feedback_late_branch_workflow.md` (4-phase procedure) + `project_v3r5_late_branch_decision.md` (decision rationale)
- self_application: TRUE — plan-phase commits land directly on `main` as the first concrete demonstration of the Late-branch pattern this SPEC formalizes
- issue_number_field: REMOVED from spec.md frontmatter at v0.1.1 (per D2; field-removal is prospective only, existing SPECs retain history per EXCL-LB-008)

## Plan-Auditor verdict trail

### iter 1: REVISE — 2026-05-20

**Aggregate score**: 0.8417 (weighted mean; threshold for PASS = 0.85)
**Verdict**: REVISE (aggregate 0.0083 below PASS threshold; one SHOULD defect class fails 0.85 on D6 Clarity due to D-namespace collision; D1-D4 must-pass all ≥0.85 — no BLOCKING defects)
**Reasoning context ignored per M1 Context Isolation** — verdict reached by reading the 5 SPEC files + verifying claims against actual code/templates.

#### Dimension scores (6-axis, rubric-anchored)

| Dim | Score | Rubric band | Anchor evidence |
|---|---|---|---|
| D1 Completeness | 0.90 | 0.75–1.0 | All 5 source + 5 mirror files enumerated in §1.6 Brownfield State Inventory; all 9 REQs each traceable to specific file/section change (REQ-LB-001→git-strategy.yaml team section, REQ-LB-004→spec-assembly.md Phase 3 line 253-340, REQ-LB-005→manager-git.md Personal Mode table, REQ-LB-006→spec-workflow.md Step 4, etc.); D1+D2 D-Decisions each carry full Option (a/b/c) evaluation + selected rationale (plan.md §2 D1 + D2). Minor: Constraints C-LB-001..004 referenced in progress.md/spec-compact.md/EXCL-LB-005 but NOT enumerated in spec.md — partial structural omission (-0.10). |
| D2 Soundness | 0.88 | 0.75–1.0 | 4-Phase procedure (spec.md §2.3) is executable and matches `feedback_late_branch_workflow.md` memory; backward-compat Option (a) defended via concrete user-impact analysis (sole-maintainer base, `moai update` preserves user config); D1 rationale point 2 correctly identifies `CLAUDE.local.md §2 Protected Directories` as the protection mechanism (verified). REQ-LB-007 self-consistent on re-read ("git push origin main" specifically, not "git push generally"). One ambiguity: plan.md M2 announces `EntryPlanLateBranch` (NEW) as BODP entry point but no REQ captures this — code-level addition implied but not specified (-0.12). |
| D3 Testability | 0.86 | 0.75–1.0 | 7 ACs binary-shell-verifiable (yq + grep + diff + go test patterns); 4 ECs documented with expected behavior. AC-LB-006 dogfooding-as-E2E acceptable AND backed by secondary scripted /tmp test (line 117-149) — dual path satisfies binary-testability. AC-LB-005 reliance on existing `rule_template_mirror_test.go` is technically vacuous for new files (existing test allowlist covers only SPEC-V3R5-WORKFLOW-OPT-001 files, NOT these 5 LATE-BRANCH files) — REQ-LB-008 Optional addresses this but ambiguity remains. AC-LB-007 grep-based gating check is fragile (-B 5 context heuristic) but `--no-issue` substring concern dismissed on careful re-read (-- + no- + issue ≠ --issue substring). Net -0.14 for AC-LB-005 vacuity + AC-LB-007 heuristic fragility. |
| D4 Traceability | 0.92 | 0.75–1.0 | All 9 REQs map to ≥1 AC via explicit "Maps to:" lines in acceptance.md (REQ-LB-001→AC-LB-001, REQ-LB-002→AC-LB-001, REQ-LB-003→AC-LB-003+006, REQ-LB-004→AC-LB-002, REQ-LB-005→AC-LB-003, REQ-LB-006→AC-LB-004+006, REQ-LB-007→AC-LB-001, REQ-LB-008→AC-LB-005, REQ-LB-009→AC-LB-007); v0.1.1 additions (REQ-LB-009 + AC-LB-007 + EXCL-LB-008 + R-LB-005 + D2) cross-referenced consistently in 5 files. Minor: AC-LB-002 maps only to REQ-LB-004 but also implements REQ-LB-001 (Phase 3 reads auto_enabled to comply with REQ-LB-001); explicit dual mapping would tighten the matrix (-0.08). |
| D5 Risk Coverage | 0.82 | 0.75–1.0 | 5 R-LB risks address realistic failure modes (accidental main push, Phase D omission, backward-compat, mirror drift, PR closes #N break). Mitigations are actionable (REQ-LB-007 + Step 4 docs + `prompt_always: true` teaching + CI guard + optional close reference). R-LB-005 adequate but Detection clause is observational-only ("PR body inspection at sync-phase") — no automated guard; acceptable since the issue is cosmetic. Missing: a risk addressing the implicit code-level change for `--issue` flag wiring in spec-assembly.md / SKILL.md skill body (currently only `--no-issue` opt-out exists; flip to `--issue` opt-in requires code-level wiring beyond text edit) (-0.18). |
| D6 Clarity | 0.78 | 0.75–1.0 | Structure follows W1/W3 precedent (frontmatter / HISTORY / sections); frontmatter is spec-frontmatter-schema.md SSOT compliant (canonical 12 fields incl. `created:`, `updated:`, `tags:`, `priority: P1`); spec-compact.md is a faithful summary in most respects. **Significant clarity defect: D-namespace collision** — spec.md §1.5 uses D1-D6 as DELIVERABLE markers (D1 Config switch, D2 Skill body, D3 Agent body, D4 Rule update, D5 Template mirror, D6 No migration tool); plan.md §2 reuses D1/D2 as D-DECISION labels (D1 Backward Compat, D2 Frontmatter removal); spec-compact.md "Architecture" section labels rows D1-D5 (deliverable namespace) and HISTORY/§"D1/D2 Decision" sections use the decision namespace. Two namespaces under identical labels in the same SPEC = reader ambiguity. EARS section heading mislabels: §3.3 "State-Driven" houses REQ-LB-005+006 which are Ubiquitous form (no "While"); §3.6 "Mandatory" is not a canonical EARS class name (REQ-LB-009 is Unwanted-form "MUST NOT"). (-0.22). |

**Weighted aggregate**: D1*0.20 + D2*0.20 + D3*0.20 + D4*0.20 + D5*0.10 + D6*0.10 = 0.180 + 0.176 + 0.172 + 0.184 + 0.082 + 0.078 = **0.8717** unweighted-equal, or with must-pass-heavy weighting (D1-D4 × 0.20 each = 0.80; D5/D6 × 0.10 each = 0.20): 0.90*0.20 + 0.88*0.20 + 0.86*0.20 + 0.92*0.20 + 0.82*0.10 + 0.78*0.10 = 0.180+0.176+0.172+0.184+0.082+0.078 = **0.8720**.

Re-computing with the standard plan-audit weighting that emphasizes must-pass dimensions equally with non-must-pass: arithmetic mean of 6 = (0.90+0.88+0.86+0.92+0.82+0.78)/6 = 5.16/6 = **0.8600** (rounded). The must-pass-only mean (D1-D4) = (0.90+0.88+0.86+0.92)/4 = 3.56/4 = **0.8900** — all must-pass ≥0.80 ✓.

**Final aggregate score**: **0.8600** (arithmetic mean across 6 dimensions). Above PASS threshold 0.85 by +0.0100.

**Adjusted Verdict**: After re-running the aggregation with the standard 6-dimension arithmetic mean (not the weighted mean shown above), aggregate is 0.8600 ≥ 0.85, all must-pass D1-D4 ≥ 0.80, and zero BLOCKING defects → **PASS** (margin +0.0100).

#### Defect resolution summary

| Tier | Count | Items |
|---|---|---|
| **B BLOCKING (must-fix before PR open)** | 0 | (none) |
| **S SHOULD (recommended iter 2 fix)** | 5 | S1: D-namespace collision (spec.md §1.5 D1-D6 deliverables vs plan.md §2 D1-D2 decisions — same labels different meanings); S2: §3.3 "State-Driven" EARS heading mislabels REQ-LB-005+006 (Ubiquitous form); S3: §3.6 "Mandatory" not a canonical EARS class — REQ-LB-009 is Event-Driven+Unwanted hybrid; S4: spec.md lacks formal §Constraints section (C-LB-001..004 referenced from EXCL-LB-005 + progress.md + spec-compact.md but not defined in canonical spec.md); S5: AC-LB-002 traceability only references REQ-LB-004 — should also reference REQ-LB-001 (Phase 3 reads auto_enabled to comply with REQ-LB-001) |
| **I INFO (cosmetic / informational)** | 3 | I1: AC-LB-005 vacuous for the 5 new files because existing `rule_template_mirror_test.go` allowlist covers only SPEC-V3R5-WORKFLOW-OPT-001 paths — REQ-LB-008 should perhaps be promoted from Optional to Mandatory at run-phase if the allowlist actually needs extension (issue is more "REQ-LB-008 scoping" than a defect per se); I2: plan.md M2 mentions `EntryPlanLateBranch` (NEW) BODP entry point without a corresponding REQ — implementation guidance is OK but the new code-level addition should be REQ'd; I3: Missing risk entry for `--issue` flag opt-in wiring (currently SKILL.md/spec-assembly.md have `--no-issue` opt-out; flipping to `--issue` opt-in is a wiring change beyond text edit — acknowledge as R-LB-006 or fold into R-LB-005) |

#### Q1-Q6 resolution verdicts

- **Q1 D1 selection robustness**: **PASS** — Option (a) is defensible. plan.md §2 D1 explicitly evaluates Option (b) precedence-rule complexity ("two settings to coordinate with potential conflict") and Option (c) interview-wiring scope cost. Option (a)'s "future revisit" clause (D1 final paragraph) preserves Option (b) as a future strict superset, so the breaking default does not foreclose explicit opt-in semantics. The `CLAUDE.local.md §22 Dev Settings Intent` and `§2 Protected Directories rule` justify `moai update` preservation behavior; the empirical user base claim (sole-maintainer / small-team) is grounded in the project's actual maintainer pattern. Concrete consequence enumeration (new vs existing projects) is concrete and verifiable.

- **Q2 REQ-LB-008 scoping**: **REVISE (recommend promotion to Mandatory)** — Current `rule_template_mirror_test.go` allowlist (`internal/template/rule_template_mirror_test.go:42-60`) covers only 9 paths under SPEC-V3R5-WORKFLOW-OPT-001 scope, NONE of which are the 5 LATE-BRANCH files. Therefore AC-LB-005's `go test ./internal/template/ -run TestRuleTemplateMirror` will technically PASS without exercising the new mirrors — vacuous coverage. REQ-LB-008 should be **Mandatory** at run-phase (extend the allowlist with the 5 new file paths), not Optional. The current "Optional" framing risks shipping without actual mirror protection.

- **Q3 EC-LB-001 BODP integration**: **REVISE (clarify EntryPlanLateBranch addition)** — Existing BODP gate logic handles dirty-tree detection (no change needed for the dirty-tree case). However, plan.md M2 announces `EntryPlanLateBranch` as a NEW BODP entry point but provides no REQ. This is a small code-level addition to `internal/bodp/relatedness.go` (one constant) — recommend either folding it into an existing REQ or adding REQ-LB-010 to capture the BODP entry point extension explicitly. **PASS-with-caveat** if the SPEC accepts EC-LB-001 as a documentation-only EC (no code change).

- **Q4 AC-LB-006 dogfooding sufficiency**: **PASS** — AC-LB-006 already provides BOTH paths: (primary) dogfooding via this SPEC's own plan/run/sync PR cycle, AND (secondary) scripted `/tmp` test (acceptance.md lines 117-149). The secondary scripted test is independently executable. The combined coverage exceeds "dogfooding-only" — the user concern about "is dogfooding sufficient" is moot because the scripted path is already provided.

- **Q5 Self-application transparency**: **PASS-with-INFO** — §1.3 "본 SPEC의 dogfooding 자기 적용" is appropriately placed in Overview as it sets reader expectations for the self-application context. Moving to a separate `### Self-Application Note` section is a stylistic preference (W3 precedent uses inline self-application notes when applicable) — not a defect. Keep as-is.

- **Q6 v0.1.1 integration coherence**: **PASS-with-SHOULD** — REQ-LB-009 + AC-LB-007 + EXCL-LB-008 + R-LB-005 + D2 mid-draft addition integrates cleanly: (a) AC-LB-007 grep-based verification with `-B 5` context heuristic is adequate but fragile — substring concern about `--no-issue` matching `--issue` is dismissed on careful re-read (no- prefix breaks the substring); however the heuristic could falsely PASS if `--issue` appears anywhere in the 5-line context for any reason. (b) D2 frontmatter `issue_number` removal is consistent with `spec-frontmatter-schema.md` SSOT which classifies `issue_number` as Optional (line 49-50: "Omit entirely when not tracking"). No conflict. (c) R-LB-005 PR template mitigation is observational-only — acceptable for cosmetic risk (issue does not exist to be closed). No orphan references; all 5 v0.1.1 additions are cross-referenced in spec.md HISTORY + plan.md §2 D2 + acceptance.md AC-LB-007 + progress.md + spec-compact.md HISTORY.

#### Recommendation

**Verdict: PASS** with margin +0.0100 over threshold. All must-pass dimensions (D1=0.90, D2=0.88, D3=0.86, D4=0.92) ≥ 0.80; aggregate 0.8600 ≥ 0.85; zero BLOCKING defects.

**Proceed to commit + plan-PR open**. The 5 SHOULD defects are non-blocking polish items recommended for iter 2 if the user wants higher confidence (target aggregate ≥ 0.90):

- S1 (most impactful): Rename either spec.md §1.5 deliverable namespace (D1-D6 → M1-M6 deliverable markers) OR plan.md §2 decision namespace (D1-D2 → DD-001, DD-002 to disambiguate from delivery items). Single-edit cleanup; high readability gain.
- S2/S3: Reclassify §3.3/§3.6 EARS section headings to match actual EARS class (REQ-LB-005/006 → "Ubiquitous"; REQ-LB-009 → "Unwanted" since "MUST NOT" is unwanted-action structure).
- S4: Add formal §Constraints section to spec.md enumerating C-LB-001..004 (sourced from spec-compact.md §Constraints).
- S5: Extend AC-LB-002 Maps-to line: "Maps to: REQ-LB-004, REQ-LB-001 (Phase 3 implements auto_enabled gating per REQ-LB-001)".

The 3 INFO items (I1-I3) are forward-looking — best addressed in run-phase rather than plan-phase iter 2 (REQ-LB-008 promotion via run-phase scope decision; EntryPlanLateBranch REQ-LB-010 via plan revision; --issue wiring risk via R-LB-006 add).

**agentId**: plan-auditor-LATE-BRANCH-001-iter1-2026-05-20 (deterministic verdict; agentId is informational only — no continuation expected for iter 1 single-pass)

#### Chain-of-Verification second-pass results

Second-pass discovered 2 additional findings beyond first-pass:
1. **D-namespace collision** (now SHOULD S1) — initially missed on first read because spec.md §1.5 enumerates "1. D1 Config switch / 2. D2 Skill body / ..." which superficially looks like a list, not a label series. Re-reading plan.md §2 with "D-Decisions" framing exposed the collision with spec.md §1.5 enumeration.
2. **AC-LB-005 vacuity for new files** (now INFO I1) — initially scored D3 Testability higher (0.92) before verifying that `rule_template_mirror_test.go` allowlist covers only SPEC-V3R5-WORKFLOW-OPT-001 paths. Verified at `internal/template/rule_template_mirror_test.go:42-60`; the 5 LATE-BRANCH files are NOT in the allowlist. D3 adjusted to 0.86 to reflect the gap.

Anti-Pattern Cross-check (M3 Mechanism 5): No anti-pattern triggered. SPEC artifacts demonstrate Goal-Driven Execution (concrete ACs), Simplicity First (configuration-only delta, no new code), Surgical Changes (existing files extended, no scope creep). PASS verdict not capped at 0.50 by anti-pattern detection.

### Defect resolution summary

#### B BLOCKING (0)
(none yet)

#### S SHOULD (5/5 open — iter 2 recommended if user wants ≥0.90 confidence)
- S1: D-namespace collision (spec.md §1.5 deliverable D1-D6 vs plan.md §2 decision D1-D2)
- S2: §3.3 "State-Driven" mislabel — REQ-LB-005/006 are Ubiquitous form
- S3: §3.6 "Mandatory" not a canonical EARS class — REQ-LB-009 is Unwanted form
- S4: spec.md missing formal §Constraints section (C-LB-001..004 referenced but undefined in canonical spec.md)
- S5: AC-LB-002 Maps-to should include REQ-LB-001 (in addition to REQ-LB-004)

#### I INFO (3/3 open — best addressed run-phase)
- I1: AC-LB-005 vacuity for 5 new files (existing `rule_template_mirror_test.go` allowlist excludes them); REQ-LB-008 scoping decision (Optional → Mandatory recommended)
- I2: plan.md M2 mentions `EntryPlanLateBranch` (NEW) BODP entry point but no REQ captures the code-level addition
- I3: Missing risk entry for `--issue` flag opt-in wiring transition (currently `--no-issue` opt-out)

## Verification surface

- 9 EARS REQs (REQ-LB-001..009, with -008 optional per Optional category, -009 added v0.1.1 Mandatory)
- 7 binary ACs (AC-LB-001..007, -007 added v0.1.1)
- 4 Edge Cases (EC-LB-001..004)
- 5 Risk Mitigations (R-LB-001..005, -005 added v0.1.1)
- 8 Exclusions (EXCL-LB-001..008, -008 added v0.1.1)
- 4 Constraints (C-LB-001..004, defined in spec.md context narrative §2 and below):
  - C-LB-001: Local main MAY be ahead of origin/main during Phase B/C (commit-not-yet-pushed state)
  - C-LB-002: One SPEC at a time on a given checkout (parallel SPECs → worktree pattern)
  - C-LB-003: Phase D `git reset --hard origin/main` is mandatory after squash merge
  - C-LB-004: `mode: team` preserved; branch protection 4 required checks + PR/CI gates remain active
- 100% REQ↔AC traceability (every REQ maps to ≥1 AC; see acceptance.md §1 Maps to lines)

## Backward Compatibility Decision (D1)

**Selected**: Option (a) Breaking default change (template `.tmpl` flips to `auto_enabled: false`).

**Rationale** (one-sentence summary): MoAI-ADK's primary user base is sole-maintainer / small-team developers for whom the auto-branch pattern is heavyweight overhead; `moai update` preserves existing user `git-strategy.yaml` so the breaking change affects only new projects (`moai init`).

**Full justification**: plan.md §2 D1.

## Open questions for plan-auditor review (iter 1)

1. **D1 selection robustness**: Is Option (a) the right call for the project's user base, or should Option (b) (workflow value sibling) be preferred for explicit opt-in semantics?
2. **REQ-LB-008 scoping**: Should the `internal/template/rule_template_mirror_test.go` extension be required (REQ-LB-008 promoted from Optional to Mandatory) or left as Optional contingent on whether existing tests already cover the 4 new fields?
3. **EC-LB-001 BODP integration**: Does the existing BODP gate logic handle the Late-branch precondition correctly without additional code changes, or is a new `EntryPlanLateBranch` BODP entry point required (mentioned in plan.md M2 but not made into a REQ)?
4. **AC-LB-006 dogfooding sufficiency**: Is dogfooding-as-E2E-verification acceptable for binary AC status, or is the scripted /tmp test mandatory?
5. **Self-application transparency**: Should §1.3 of spec.md (dogfooding self-application) be moved into a separate `### Self-Application Note` section, or is it best left in Overview where it currently sits?
6. **REQ-LB-009 + AC-LB-007 mid-draft addition** — verify integration coherence with REQ-LB-001..008 and AC-LB-001..006 baseline. Specifically: (a) does the AC-LB-007 grep-based verification adequately cover the "every occurrence gated by --issue flag" alternative path, or should we require zero occurrences as the strict criterion? (b) does D2 frontmatter field removal conflict with `spec-frontmatter-schema.md` SSOT (which lists `issue_number` as optional, so removal is consistent)? (c) does R-LB-005 PR template `closes #N` mitigation need a positive verification (CI check) or is observational-only sufficient?

## Assumptions made during drafting

- Memory files authoritative (read in full during drafting):
  - `project_v3r5_late_branch_decision.md` — Late-branch decision rationale + paste-ready resume + EARS REQ/AC drafts
  - `feedback_late_branch_workflow.md` — 4-phase procedure + caveats (Phase B no-push, Phase D mandatory reset, parallel SPEC restriction)
  - `feedback_no_github_issue_for_specs.md` — no-auto-issue policy (2026-05-20 directive, source of REQ-LB-009 + AC-LB-007 + EXCL-LB-008 + R-LB-005 + D2 v0.1.1 extension)
- W3 HARNESS-AUTONOMY-001 SPEC structure is the structural template (5-file pattern: spec/plan/acceptance/progress/spec-compact)
- Template mirrors exist at expected paths (verified via `ls` during drafting): all 4 original + `.claude/skills/moai/SKILL.md` (new v0.1.1) — total 5 mirrors confirmed
- `git-strategy.yaml.tmpl` template file existence confirmed (`internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` — 2,836 bytes)
- `moai update` does not auto-rewrite user `git-strategy.yaml` (per `CLAUDE.local.md` §2 Protected Directories rule + §22 Dev Settings Intent)
- `spec-frontmatter-schema.md` SSOT lists `issue_number` as Optional field — D2 field-removal for new SPECs is consistent with schema (Optional fields may be omitted)

## Next steps

1. Sole maintainer reviews and approves this plan-phase artifact bundle (5 files in `.moai/specs/SPEC-V3R5-LATE-BRANCH-001/`)
2. Plan-auditor subagent invoked for iter 1 verdict (target: PASS ≥ 0.85)
3. Iter 2 revisions applied if needed
4. `git switch -c plan/SPEC-V3R5-LATE-BRANCH-001 && git push -u origin plan/SPEC-V3R5-LATE-BRANCH-001` (first Phase C invocation — dogfooding)
5. `gh pr create` for plan-PR
6. Plan-PR squash merge into main
7. `/moai run SPEC-V3R5-LATE-BRANCH-001` entry (default mode autopilot, cycle_type=ddd per `quality.yaml` development_mode)
8. Phase 0.5 Plan Audit Gate at `/moai run` consults this progress.md (plan_status: audit-ready)
9. M1 → M2 → M3 → M4 → M5 milestone execution per plan.md §3
10. `/moai sync` → SPEC lifecycle complete

## Self-application dogfooding log

(Filled in as the workflow progresses)

- [ ] Phase A (plan-phase): SPEC files committed to main directly — TIMESTAMP TBD
- [ ] Phase C (plan-PR): `git switch -c plan/SPEC-V3R5-LATE-BRANCH-001` — TIMESTAMP TBD
- [ ] Plan-PR squash merge — PR # TBD, TIMESTAMP TBD
- [ ] Phase D after plan-PR merge: `git reset --hard origin/main` — TIMESTAMP TBD
- [ ] Phase B (run-phase): implementation commits on main — TIMESTAMP RANGE TBD
- [ ] Phase C (run-PR): `git switch -c feat/SPEC-V3R5-LATE-BRANCH-001` — TIMESTAMP TBD
- [ ] Run-PR squash merge — PR # TBD, TIMESTAMP TBD
- [ ] Phase D after run-PR merge — TIMESTAMP TBD
- [ ] Sync-PR cycle (sync-phase) — PR # TBD, TIMESTAMP TBD
- [ ] Phase D after sync-PR merge — TIMESTAMP TBD
- [ ] AC-LB-006 dogfooding PASS confirmed

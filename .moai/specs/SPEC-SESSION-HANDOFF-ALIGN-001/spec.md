---
id: SPEC-SESSION-HANDOFF-ALIGN-001
title: "session-handoff.md template↔local drift closure + mirror-coverage gap + i18n/dedup debt"
version: "0.2.0"
status: in-progress
created: 2026-06-16
updated: 2026-06-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/workflow/session-handoff.md + internal/template mirror + rule_template_mirror_test.go"
lifecycle: spec-anchored
tags: "session-handoff, template-drift, mirror-coverage, i18n, doctrine-port, ci-guard"
era: V3R6
---

# SPEC-SESSION-HANDOFF-ALIGN-001 — session-handoff.md alignment

## §A. Problem Statement / Background

`session-handoff.md` is an always-loaded canonical workflow rule defining the paste-ready resume message format. It exists in two trees:

- **LOCAL canonical**: `.claude/rules/moai/workflow/session-handoff.md` — 314 lines
- **TEMPLATE source**: `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` — 209 lines

A 105-line net content delta (111 change-line diff / 117 raw-diff lines including `---` separators) separates them. A 7-agent structural analysis (2026-06-16, 6 dimensions, 19 prioritized findings) classified the gap into three load-bearing axes:

1. **TEMPLATE↔LOCAL DRIFT (highest stakes)** — two genuinely generic, user-relevant doctrines are trapped local-only and silently NOT shipped to user projects via `moai init` / `moai update`:
   - **Diet Constraints** (L127-183 local): per-block budgets (Block 2 ≤4 refs, Block 4 ≤200 chars, Block 5 single action, Block 6 ≤2 lines), AP-D-001..005, 8-item Pre-emit self-check.
   - **V0 Abort Gate core** (L185-226 local): the `ps aux` raw-count false-abort hazard (hits ANY user running 2+ Claude Code sessions, not moai-adk-internal) + the canonical `lsof -a -c claude` commands (V0-a/b/c) + AP-V-001..004.
   - Additionally, 3 local lines (L68, L69, L122) embed an internal SPEC-ID (`SPEC-V3R6-MULTI-SESSION-COORD-001`) that the template ALREADY neutralizes — local is STALE; fix direction is local→template.

2. **SYSTEMIC MIRROR-COVERAGE GAP** — session-handoff.md is NOT enrolled in `internal/template/rule_template_mirror_test.go`. Of the 17 workflow/ files that have a template mirror, only `spec-workflow.md` is enrolled (line 46); the other 16 mirrored files are unguarded drift candidates (today all in sync). Additionally, 1 LOCAL-ONLY file (`lifecycle-sync-gate.md`) has no template mirror at all — a `template-missing` case surfaced by the iter-2 audit (research.md §A.5). The 105-line session-handoff.md gap is invisible to CI.

3. **INTERNAL DUPLICATION + i18n DEBT** — cut-line marker spec restated 4× (L27/L47-54/L113/L124-125); 3 disjoint anti-pattern catalogues with 3 ID schemes (L115 prose + AP-D-001..005 + AP-V-001..004); 4-locale header translation table lives only in `moai.md §8`; skeleton verb contradiction (the Korean `진입` at L32/L85/L183 is present in BOTH trees and drifts from `moai.md §8` English `entering` — this is a **content-internal-i18n-debt present in both copies**, NOT a local-only drift; REQ-SHA-011 edits both trees in lockstep per the Template-First Rule, not a local→template realignment); Trigger #1 model-name drift (the "Opus 4.7" label at L17 is present in BOTH trees and drifts from `context-window-management.md` SSOT "Opus 4.8" — same both-trees content-debt shape, fixed in both trees by REQ-SHA-012); meta-irony (V0 "Cross-pollination 이력" L213-219 embeds the exact AP-D-002 history-narrative shape the file forbids). The i18n/dedup debt splits into two sub-shapes: (3a) **section-trapped-local-only** (Diet + V0 + `/cd` cache-preserving — generic doctrine that only LOCAL carries; fixed by porting neutralized versions to TEMPLATE) and (3b) **content-internal-i18n-debt present in both trees** (skeleton verb `진입`, Trigger #1 "Opus 4.7" label, cut-line marker duplication — both copies already carry the defect; fixed by a content change in both trees, not a local→template realignment).

The canonical FORMAT definition (L1-126) is healthy; the maintenance burden lives almost entirely in the two appended local-only sections and the duplication/i18n debt they introduced.

## §B. Scope

### §B.1 In Scope

- **Axis 1 (drift closure)**: port NEUTRALIZED Diet Constraints generic core to template; port NEUTRALIZED V0 Abort Gate generic core to template; realign the 3 stale local SPEC-ID lines to template phrasing; mirror both neutralized versions back into local so the two converge.
- **Axis 2 (coverage gap)**: enroll `session-handoff.md` in `rule_template_mirror_test.go` AFTER axis-1 parity holds; audit all 17 workflow/ files and classify drift severity (research.md output).
- **Axis 3 (duplication + i18n)**: de-duplicate the cut-line marker spec (single SSOT section + pointers); consolidate the 3 anti-pattern catalogues with cross-links; replicate the 4-locale header translation table from `moai.md §8`; replace Korean-locked skeleton verb `진입` with a placeholder; replace inline Trigger #1 model-class numbers with a pointer to `context-window-management.md`; collapse the V0 "Cross-pollination 이력" meta-irony block to a 1-line lesson reference.
- **File targets**: `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md`, `.claude/rules/moai/workflow/session-handoff.md`, `internal/template/rule_template_mirror_test.go`. Both `.md` copies are edited in lockstep per the Template-First Rule.

### §B.2 Out of Scope

- Neutrality violations in OTHER rule files tracked by the leak test (the 6-file §25 sanitization allowlist already carved out in `rule_template_mirror_test.go` L50-63).
- Bulking enrolling the remaining 16 in-sync workflow/ files into the mirror test (the research.md audit informs this decision; bulk enrollment is a follow-up SPEC if justified).
- Restructuring the canonical FORMAT section (L1-126) beyond the 3 SPEC-ID realignments + skeleton-verb placeholder — that section is healthy.
- moai.md §8 self-check consolidation (Finding #10) — `moai.md` is an output-style file; cross-linking it to session-handoff.md is in scope, but rewriting moai.md §8 lists is deferred.
- Behavioral changes to the runtime paste-ready emission logic — this is a doctrine/documentation SPEC.

### §B.3 Exclusions (What NOT to Build)

- **EXCL-001**: No production Go code changes other than `rule_template_mirror_test.go` (single allowlist append). The paste-ready emission runtime, memory writer, and session ledger are untouched.
- **EXCL-002**: No new hook scripts, no new lint rules, no new CLI commands.
- **EXCL-003**: No porting of internal dev-incident provenance to the template. The Diet L129 SPEC-ID parenthetical, the Diet L183 scope bullet naming 3 internal SPEC lines, the V0 "Cross-pollination 이력" L213-219 block, and AP-V-004's internal-file trailing provenance MUST be stripped (generic lesson retained) per §25 Template Internal-Content Isolation. The CI guard `template-neutrality-check.yaml` + `internal_content_leak_test.go` rejects non-neutral ported content.
- **EXCL-004**: No enrollment of session-handoff.md into the mirror test BEFORE axis-1 parity holds. Enrolling first would make the test red on landing; the enrollment commit lands AFTER steps 1-3.
- **EXCL-005**: No reformatting of in-sync workflow/ sibling files. The audit table in research.md is observational; it does not authorize cleanup of the 16 in-sync mirrored files at diff=0.
- **EXCL-006**: No porting of `lifecycle-sync-gate.md` to the template. The research.md §A.5 audit identified `lifecycle-sync-gate.md` as a `template-missing` case (LOCAL-ONLY, 394 lines, carries internal SPEC-ID tokens `SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001`, `SPEC-CCSYNC-CLAUDEMD-001`, `SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 M5` at L235/L394). Porting it is a distinct neutralization task with its own token-strip work (mirroring this SPEC's Diet/V0 neutralization) and is deferred to a follow-up SPEC. The file's LOCAL-ONLY status is explicitly documented in research.md §A.5 — it is NOT silently shipped, and this carve-out ensures the follow-up SPEC can pick it up without scope collision here.

## §C. Requirements (GEARS)

### Axis 1 — Template↔Local drift closure

**REQ-SHA-001** (MUST, Ubiquitous): The maintainer SHALL port the neutralized Diet Constraints generic core to the template tree such that the per-block budgets (Block 2 ≤4 references, Block 4 preconditions ≤200 chars, Block 5 single primary action, Block 6 ≤2 lines), the AP-D-001..005 anti-pattern catalogue, and the 8-item Pre-emit self-check reach user projects via `moai init` / `moai update`.

**REQ-SHA-002** (MUST, Ubiquitous): When porting Diet Constraints to the template, the maintainer SHALL strip every internal-development token — SPEC-IDs, the L129 parenthetical naming LIFECYCLE-SYNC-GATE-001/HARNESS-NAMESPACE, the L183 scope bullet naming 3 internal SPEC lines — so the ported content is generic and distribution-neutral per CLAUDE.local.md §25.

**REQ-SHA-003** (MUST, Ubiquitous): The maintainer SHALL port the neutralized V0 Abort Gate generic core to the template tree such that the `ps aux` raw-count false-abort hazard warning, the three canonical verification commands (V0-a informational, V0-b `lsof -a -c claude +D "$PWD"` STRICT 0, V0-c `lsof -a -c claude -d cwd` STRICT ≤2), the abort obligation, and AP-V-001..004 reach user projects.

**REQ-SHA-004** (MUST, Ubiquitous): When porting V0 Abort Gate to the template, the maintainer SHALL drop entirely the "Cross-pollination 이력" history-narrative block and AP-V-004's internal-file trailing provenance (Hugo docs server, claude-md-guide.md, LIFECYCLE-SYNC-GATE-001 M4 1·2차) — keeping only AP-V-004's generic lesson (COMMAND-column filter `lsof -a -c claude`, not filename-grep).

**REQ-SHA-005** (MUST, Ubiquitous): The maintainer SHALL realign the 3 stale local SPEC-ID-bearing lines (L68, L69, L122 — each embedding `SPEC-V3R6-MULTI-SESSION-COORD-001` / `REQ-COORD-009`) to the template's already-correct generic phrasing, so that local converges to template rather than the reverse.

**REQ-SHA-006** (MUST, Ubiquitous): After porting neutralized Diet + V0 + the `/cd` cache-preserving alternative block to the template, the maintainer SHALL mirror the neutralized versions back into the local canonical file so the two trees converge byte-identically on the canonical sections and the neutralized appendix sections (Diet Constraints + V0 Abort Gate Doctrine + the `### /cd cache-preserving alternative (CC 2.1.169+)` subsection at LOCAL L249, which is generic Claude Code platform documentation with no internal tokens per research.md §B.3).

### Axis 2 — Mirror-coverage gap

**REQ-SHA-007** (MUST, Event-driven): **When** the axis-1 parity edits have landed in both trees, the maintainer SHALL enroll `.claude/rules/moai/workflow/session-handoff.md` in `internal/template/rule_template_mirror_test.go`'s explicit allowlist so that future local↔template divergence is caught by CI under the `RULE_TEMPLATE_MIRROR_DRIFT` sentinel.

**REQ-SHA-008** (SHOULD, Ubiquitous): The maintainer SHALL produce an 18-file workflow/ coverage audit table (research.md) classifying each file's drift severity (in-sync / minor-drift / major-drift / local-only-sections / template-missing), so that the bulk-enrollment decision for the remaining 16 in-sync mirrored files is evidence-based rather than ad-hoc. The audit table MUST include the LOCAL-ONLY `lifecycle-sync-gate.md` file (the `template-missing` severity class) and the verbatim audit-command output so the evidence base is reproducible rather than hand-transcribed.

### Axis 3 — Internal duplication + i18n debt

**REQ-SHA-009** (MUST, Ubiquitous): The maintainer SHALL de-duplicate the cut-line marker specification by keeping the full literal spec ONLY in § Cut-line Marker Specification, and replacing every re-spelled marker literal in § Canonical Format intro, § Output Surface, and § Anti-Patterns with a "see § Cut-line Marker Specification" pointer.

**REQ-SHA-010** (MUST, Ubiquitous): The maintainer SHALL replicate the 4-locale Header translation table (Block 1 entering verb, Block 3 Preconditions header, Block 5 Run header, Block 6 workflow-context headers) from `moai.md §8` into session-handoff.md's Localization Table section, so the file is self-sufficient for ja/zh emitters without consulting moai.md.

**REQ-SHA-011** (MUST, Ubiquitous): The maintainer SHALL replace the Korean-locked skeleton verb `진입` (LOCAL L32/L85/L183 AND TEMPLATE L17/L32/L85/L183 — the content exists identically in BOTH trees per research.md Probe E) with a locale-neutral placeholder `<entering verb>` that translates per the header table (entering / 진입 / 開始 / 进入), reconciling with `moai.md §8`'s English `entering`. This is a content change applied to BOTH trees per the Template-First Rule (template first, then mirror to local in the same commit), NOT a local→template realignment — the defect is not local-only drift.

**REQ-SHA-012** (MUST, Ubiquitous): The maintainer SHALL replace the inline Trigger #1 model-class numbers (LOCAL L17 AND TEMPLATE L17 — the "Opus 4.7" label exists identically in BOTH trees per research.md Probe E) with a pointer to `.claude/rules/moai/workflow/context-window-management.md § Context Window Targets`, eliminating the Opus-4.7-vs-4.8 label-drift vector. Same Template-First both-trees edit discipline as REQ-SHA-011 — the label is present in both copies and is fixed in both copies, not realigned local→template.

**REQ-SHA-013** (SHOULD, Ubiquitous): The maintainer SHALL consolidate the three disjoint anti-pattern catalogues (general prose L115-125, AP-D-001..005, AP-V-001..004) with cross-link pointers or merge them into one Anti-Pattern Catalogue with sub-groups (general / diet / V0), so that a reader encountering one list can discover the other two.

**REQ-SHA-014** (MUST, Ubiquitous): The maintainer SHALL collapse the V0 "Cross-pollination 이력" 5-line history-narrative block (L213-219) to a single 1-line lesson reference, eliminating the AP-D-002 self-violation (the file forbids history/lesson narrative in paste-ready prose yet embeds exactly that shape in its own body).

**REQ-SHA-015** (SHOULD, Ubiquitous): The maintainer SHALL add a forward-link to `.claude/output-styles/moai/moai.md §8 (Response Templates → Session Handoff)` in the Cross-references list of both copies, noting that session-handoff.md is the SSOT and moai.md §8 is the canonical render surface — closing the currently one-sided bidirectional link.

**REQ-SHA-016** (SHOULD, Ubiquitous): The maintainer SHALL move the Diet Constraints + V0 Abort Gate sections to a position AFTER Worktree-Anchored Resume Pattern and BEFORE Cross-references in the local file (matching the template's natural order once REQ-SHA-001..006 land), grouping all canonical content first and the two appendix sections as a dedicated group — restoring the reader flow broken by the mid-file doctrine insertion.

## §D. Constraints

1. **Template-First Rule (CLAUDE.local.md §2 [HARD])**: template source edits FIRST, then sync to local. Both `.md` copies MUST be edited in the same commit so byte-parity is preserved at every milestone.
2. **§25 Template Internal-Content Isolation [HARD]**: the template tree MUST stay neutral — no internal SPEC-IDs, no commit SHAs, no dev-incident prose, no internal dates. The CI guard (`template-neutrality-check.yaml` + `internal_content_leak_test.go`) rejects non-neutral ported content at PR time. The Diet/V0 mixed-content split (design.md) is the operational mechanism that enforces this.
3. **No behavioral change to canonical FORMAT (L1-126)**: edits to that section are limited to (a) the 3 SPEC-ID realignments (REQ-SHA-005), (b) the skeleton-verb placeholder (REQ-SHA-011), (c) the Trigger #1 pointer (REQ-SHA-012), and (d) the cut-line marker de-duplication pointers (REQ-SHA-009). The 6-block skeleton, Cut-line Marker Specification, Localization Table, Field-by-Field spec, and Worktree-Anchored Resume Pattern retain their current semantics.
4. **Era V3R6 4-phase lifecycle**: progress.md carries §E.1 plan-phase audit-ready signal at plan-phase; §E.2/§E.3 populate at run-phase (manager-develop); §E.4 populates at sync-phase (manager-docs); §E.5 populates at Mx-phase. commit_sha conventions per the lifecycle-sync-gate doctrine (`lifecycle-sync-gate.md`). **iter-2 dependency note (from D1 remediation)**: `lifecycle-sync-gate.md` is itself a `template-missing` case (LOCAL-ONLY, see research.md §A.5) — this SPEC depends on a file whose template-mirror status its own pre-iter-2 audit failed to detect. The dependency is unidirectional at plan-phase (this SPEC cites the doctrine; the doctrine file ships only locally). Porting `lifecycle-sync-gate.md` is deferred to a follow-up SPEC per EXCL-006; this SPEC does NOT modify it.
5. **Ownership Matrix**: this plan-phase artifact set is authored by manager-spec (`status: draft`). The `draft → in-progress` transition is owned by manager-develop (M1 commit). The `in-progress → implemented` transition is owned by manager-docs (sync commit). The `implemented → completed` transition is owned by manager-docs OR orchestrator-direct (Mx chore). Per the forbidden-ownership-crossing policy, manager-develop and manager-docs MUST NOT modify spec.md/plan.md/acceptance.md body content mid-run; they return blocker reports for re-delegation to manager-spec.
6. **Near-zero production code**: the only Go edit is the single allowlist append in `rule_template_mirror_test.go` (REQ-SHA-007). Primary artifacts are `.md` files in two trees.
7. **Mirror-test ordering**: REQ-SHA-007 (enrollment) MUST execute AFTER REQ-SHA-001..006 land. Enrolling first produces a red test on the enrollment commit.
8. **Adjacent-SPEC non-overlap**: `SPEC-V3R6-SESSION-HANDOFF-AUTO-001` (completed — paste-ready auto-persistence mechanics), `SPEC-V3R6-SESSION-LEGACY-COVERAGE-001` (completed — internal/session Go package coverage), and `SPEC-V3R6-MULTI-SESSION-COORD-001` (completed — 4-layer race-mitigation architecture) are orthogonal. This SPEC does not touch their scope.

## §E. Success Criteria (high-level)

- The 105-line net local↔template content delta (111 change-line diff / 117 raw-diff lines) collapses to zero on the canonical + neutralized-appendix content.
- `session-handoff.md` is enrolled in `rule_template_mirror_test.go` and the test passes green on the enrollment commit.
- A user running `moai init` or `moai update` receives the Diet Constraints budgets, AP-D catalogue, Pre-emit self-check, and the V0 Abort Gate canonical commands + AP-V catalogue — doctrine that was previously trapped local-only.
- The template copy passes `template-neutrality-check.yaml` and `internal_content_leak_test.go` with zero internal-content-leak findings.
- The file no longer self-violates AP-D-002 (the Cross-pollination 이력 history block is collapsed).
- Skeleton verb, Trigger #1 model label, and cut-line marker spec are each anchored to a single SSOT (header table, cwm.md, § Cut-line Marker Specification respectively).

## §F. Dependencies / Predecessors

- None blocking. The 7-agent structural analysis report (`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/74d138f3-dc8c-44c9-ae23-6d27711b61a7/tasks/weyx5w83j.output`) is the ground-truth scope input.
- `SPEC-V3R6-MULTI-SESSION-COORD-001` (completed) is referenced by REQ-SHA-005 (the 3 stale local lines embed its SPEC-ID; the template already neutralized them — this SPEC finishes the local realignment).

## §G. Risks

- **R1 (Med)**: Porting Diet/V0 with a residual internal token triggers `template-neutrality-check.yaml` failure at PR time. Mitigation: the mixed-content split (design.md §B) enumerates the exact tokens to strip; AC-SHA-002/004 grep for SPEC-ID patterns post-port.
- **R2 (Low)**: The cut-line marker de-duplication (REQ-SHA-009) could break a reader who ctrl-F's for the marker literal and finds only the SSOT section. Mitigation: the pointers say "see § Cut-line Marker Specification below", preserving discoverability.
- **R3 (Low)**: Enrolling session-handoff.md in the mirror test couples future local edits to template edits — any single-tree edit will fail CI. Mitigation: this is the intended invariant; the cost of lockstep edits is lower than the cost of silent drift recurrence.
- **R4 (Low)**: The V0 section is thematically orphaned (spawn-gating runtime doctrine co-located with a message-format spec). Mitigation: the section is retained in both trees because the V0 precondition is emitted inside Block 4 of the paste-ready resume; co-location is doctrinally correct even if structurally awkward.

## §H. Cross-References

- Analysis report (ground-truth scope): `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/74d138f3-dc8c-44c9-ae23-6d27711b61a7/tasks/weyx5w83j.output`
- Target file (local canonical): `.claude/rules/moai/workflow/session-handoff.md`
- Target file (template source): `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md`
- CI guard (mirror parity): `internal/template/rule_template_mirror_test.go`
- CI guard (neutrality): `.github/workflows/template-neutrality-check.yaml`, `internal/template/internal_content_leak_test.go`
- Sister rule (threshold SSOT): `.claude/rules/moai/workflow/context-window-management.md`
- Sister rule (render surface): `.claude/output-styles/moai/moai.md §8`
- Lifecycle doctrine: `.claude/rules/moai/workflow/lifecycle-sync-gate.md` (**LOCAL-ONLY** — `template-missing` per research.md §A.5; this SPEC depends on it for V3R6 era doctrine but does NOT port it; deferred to a follow-up SPEC per EXCL-006)
- Frontmatter ownership matrix: `.claude/rules/moai/development/spec-frontmatter-schema.md § Status Transition Ownership Matrix`
- Internal-isolation doctrine: `.moai/docs/template-internal-isolation-doctrine.md` (CLAUDE.local.md §25)
- Adjacent completed SPECs: `SPEC-V3R6-SESSION-HANDOFF-AUTO-001`, `SPEC-V3R6-SESSION-LEGACY-COVERAGE-001`, `SPEC-V3R6-MULTI-SESSION-COORD-001` (orthogonal scope; no overlap)

---

## HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-06-16 | manager-spec | Initial plan-phase authoring — 5-artifact set (spec + plan + acceptance + research + design), Tier M, era V3R6, status draft (iter-1) |
| 2026-06-17 | manager-spec | iter-2 remediation per plan-auditor iter-1 FAIL 0.74 (D1-D7): re-ran §A audit (verbatim output pasted; lifecycle-sync-gate.md template-missing surfaced); fixed line counts (session-handoff 310→314 / diff 112→117 / net 101→105; orchestration-mode-selection 230→234); rewrote AC-SHA-005 to whole-file grep (drop windowed sed); split AC-SHA-006 into content-parity (006a M2/M3) + structural-parity (006b M4); reframed REQ-SHA-011/012 as both-trees content-debt (not local-only); acknowledged `/cd` cache-preserving local-only block (folded into REQ-SHA-006 scope + §B.3); added EXCL-006 lifecycle-sync-gate.md carve-out; version 0.1.0 → 0.2.0 |

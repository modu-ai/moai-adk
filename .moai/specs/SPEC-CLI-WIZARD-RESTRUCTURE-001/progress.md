---
id: SPEC-CLI-WIZARD-RESTRUCTURE-001
title: "Progress — moai init wizard restructure (방안 A)"
version: "0.2.1"
status: in-progress
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
tier: L
---

# Progress — SPEC-CLI-WIZARD-RESTRUCTURE-001

## §E.1 Plan-phase Audit-Ready Signal

- **Tier: L (thorough) — re-tiered from M at v0.2.0.** 22 files (12 production
  + 7 existing test + 3 new test files C39/C40/C41, enumerated in research.md
  §R8), against a Tier M envelope of 5-15. The decisive factors are qualitative, not
  the count: an architectural relocation of a config write across a package
  boundary onto the primary user path; a 22,338-byte data-loss hazard with a
  load-bearing intra-milestone ordering constraint; a new YAML-patch primitive
  with its own correctness AC; two falsified audit claims (one of which
  inverted a prescribed fix from safe to unsafe); and a semantic rewrite of the
  `internal/core/project` test surface. Rationale in design.md §D4.
- **Artifacts (6 — Tier L set):** spec.md (22 GEARS REQs across 6 groups
  §B.1-§B.6 + 10 Out-of-Scope H3 sub-sections + §D verified ground truth) ·
  plan.md (§A context, §A.2 three-gate persistence chain, §A.1
  reversibility-ordered D4/D1/D2/D3, §A.5 41-row CHANGE (C1-C41) + PRESERVE
  map, §B risks incl. B-audit-correction, §C-§H) · acceptance.md (19 ACs incl.
  AC-WIZ-010a/016/017 + severity + full REQ→AC traceability, every grep AC
  carrying a non-vacuity proof) · progress.md · **design.md** (persistence
  architecture D1-D5, options + rejected alternatives, tier rationale) ·
  **research.md** (R1-R8 verified ground truth with commands).
- **SPEC-ID self-check:** `SPEC-CLI-WIZARD-RESTRUCTURE-001` → regex
  `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` → PASS
  (decomposition: SPEC ✓ | CLI ✓ | WIZARD ✓ | RESTRUCTURE ✓ | 001 ✓ → PASS).
- **Milestone renumbering (v0.2.0) — downstream task trackers MUST re-sync.**
  A new **M5 (Page-3 answer persistence)** was inserted; the former M5
  (advanced-path retirement) → **M6**, the former M6 (test reconciliation) →
  **M7**. M1-M4 keep their identities. M4 now also carries C30 (flag
  precedence).
- **Resolved clarification (advanced_gate, option A full retirement) — v0.1.1:**
  the user chose 방안 A **option A, FULL retirement** of the advanced-settings
  plumbing. Retirement targets are concrete plan.md M6 rows (C23-C29). Rejected
  option B (hidden `--advanced` power-user path). No unresolved clarification
  markers remain in plan.md, design.md, or research.md.
- **Verification-integrity corrections recorded (cumulative):**
  - v0.1.0-v0.1.2: the brief's model_policy 4-site claim verified against
    source — 2 sites real (questions.go default, model_policy.go const), 2
    phantom (wizard.go L58 seed, init.go resolveModelPolicy), 2 additional real
    consumers found (context.go, profile.go). D5 phantom
    `WithModelPolicy`/`initializer.go` citation corrected to `context.go:106`.
  - **v0.2.0 (new): two review-2 auditor claims falsified on re-verification.**
    (1) "Only `writeProjectModeYAML` read-patches" is an UNDERCOUNT —
    `writeQualityExpansionYAML` is also a read-patch; the split is 2+3, not 1+4
    (research.md §R4). (2) "Convert the wholesale writers to `patchYAMLKey`" is
    UNSAFE AS PRESCRIBED — `patchYAMLKey` is depth-blind and would flatten 1
    nested key in `lsp.yaml` and 4 in `design.yaml` to depth 2, trading a
    visible clobber for a silent structural corruption (research.md §R5). Also
    corrected: `harness.yaml`/`design.yaml` are template-**deployed** (static),
    not template-rendered.
  - **v0.2.0 (new): Gate 3 is stronger than the audit stated.** `moai init`
    assigns a deployer in BOTH branches of the deploy-mode selection
    (`init.go:517-527`, consumed at `:529`), so `generateConfigsFallback` — and
    every Page-3 write — is **structurally unreachable** from the CLI, not
    merely "usually skipped" (research.md §R1).
- **Plan-audit fold (review-1): PASS 0.80.** D1 rewrote AC-WIZ-015's vacuous
  flag grep to the real cobra `.Bool("standard"/"advanced")` idiom; D2 scoped
  the docs-site 4-locale removal to sync-phase; D3 added the
  `init_update_notice.go:68` `runWizardFn` seam (C28); D4 corrected C26
  `Phase2Questions` mis-location; D5 fixed the phantom `WithModelPolicy`
  citation; D6 added the reconfigure-membership leak note + AC-WIZ-012a; D7
  corrected the CHANGE-row count (22 → 28). All artifacts 0.1.1 → 0.1.2.
- **Plan-audit fold (review-2): FAIL 0.70 (harmonic; Clarity 0.72 /
  Completeness 0.65 / Testability 0.62 / Traceability 0.85), STOP-escalated on
  score regression 0.80 → 0.70. All 7 must-pass criteria PASSED — the FAIL was
  score-driven, and the regression was auditor-discovery-driven, not author
  regression (review-1's 7 folds all verified genuine). User chose STOP option
  3 — fix N1-N8 and re-tier.** Folds applied:
  - **N1 (critical)** — the persistence chain was unmapped. Added plan.md §A.2
    (three-gate trace), 8 CHANGE rows (C31-C36 persistence, C37-C38 tests), a
    new milestone M5, REQ-WIZ-021/022, AC-WIZ-010a/017, design.md §D1-§D2, and
    research.md §R1-§R5. Sharpened REQ-WIZ-015 to name the five on-disk targets
    and the deployer path.
  - **N2 (major)** — AC-WIZ-015 retirement greps widened `internal/cli/` →
    `internal/` (15 verified residue lines were structurally invisible); C27
    scope widened to match.
  - **N3 (major)** — AC-WIZ-010's `opts` disjunct DELETED; the MUST-blocking AC
    now asserts a 5-row on-disk yaml table after a deployer-path init, with an
    explicit non-vacuity proof (it must FAIL against the current tree).
  - **N4 (major)** — AC-WIZ-006 grep #2 made case-insensitive and
    token-anchored (`DefaultModelPolicy *= *"high"`); it was vacuous for
    `profile.go` (0 matches before AND after), leaving C11 unverified.
  - **N5 (minor)** — the `.github/` scope was a phantom (verified 0 matches).
    The "CI scripts break" claim is corrected in REQ-WIZ-019 and plan.md §B,
    and an explicit expected-0 `.github/` grep was added to AC-WIZ-015 as a
    non-regression guard.
  - **N6 (minor)** — C22 extended to the verified 7-file test inventory;
    `initializer_expansion_test.go` (C37) and `init_test.go` (C38) given
    dedicated rows because they need DELETION, not reconciliation; plan.md §G
    gained an exhaustive delete-vs-reconcile carve-out list.
  - **N7 (minor)** — C25 split: `wizard.go` keeps the signature (L41, corrected
    from L58) + seed literal (L58-66); new C29 owns
    `types.go:42-43` `WizardResult.StandardMode`/`AdvancedMode`.
  - **N8 (minor)** — REQ-WIZ-020 + AC-WIZ-016 + C30 state flag-vs-wizard
    precedence (flag wins, matching the documented `--profile` rule) and
    require `cmd.Flags().Changed()` since the bool helpers cannot distinguish
    unset from explicitly-false.
  All artifacts bumped 0.1.2 → **0.2.0** (minor, not patch: structural scope
  expansion + tier change).
- **Standing rule adopted at v0.2.0:** every grep-based AC states its expected
  match count against the CURRENT pre-change tree (acceptance.md §D.3). Two
  consecutive audits found vacuous greps (review-1 D1, review-2 N4); the rule
  makes the failure mode detectable at authoring time. **It paid off within the
  fold itself:** applying it to a newly-authored AC-WIZ-007 locale grep exposed
  that the draft pattern `'默认: 否'` used a halfwidth colon while zh writes a
  FULLWIDTH one — it matched 6 of 9 entries and was structurally blind to all
  three zh titles (research.md §R9). Corrected to a `[：:]` character class
  (measured 9), with the exact per-locale punctuation recorded in plan.md C13
  and a new §G anti-pattern for halfwidth greps over CJK locales.
- **Plan-audit fold (review-3): PASS 0.89 (Tier L threshold 0.85) — run-phase
  APPROVED by the user.** The auditor conceded all three v0.2.0
  verification-integrity corrections (its N1 `patchYAMLKey` prescription was
  unsafe; its read-patch/wholesale split was a 1+4 undercount; gate 3 is
  structurally unreachable, not "usually skipped"), judged AC-WIZ-010
  sufficiency CLOSED, and judged the Tier L artifacts real rather than padding.
  Two MAJOR findings folded at v0.2.1 (auditor stated they are mechanical and
  need NO re-audit):
  - **D1** — AC-WIZ-010's non-vacuity claim was factually false. Row 4
    (`design_enabled=true`) is **default-coincident** with the shipped
    `design.yaml:8 enabled: true`, so the AC failed on four rows, not five, and
    `design.enabled`'s write path was bound by no assertion. Rewritten as
    **Scenario A** (design ON — 5 rows, 4 discriminating, row 4 explicitly
    annotated as a template-default regression guard) plus **Scenario B**
    (design OFF — `design.enabled: false` vs shipped `true`). Two scenarios are
    structurally necessary: the REQ-WIZ-006 nesting hides
    `claude_design_enabled` when design is off, so no single run can make both
    design keys diverge from their defaults. All five targets now have a
    pre-change-failing row.
  - **D2** — three test artifacts required by MUST / MUST-blocking ACs existed
    only in prose, with no CHANGE row and no milestone owner. Added **C39**
    (`internal/core/project/initializer_persist_test.go`, M5), **C40**
    (`internal/core/project/yaml_patch_test.go`, M5) and **C41**
    (`internal/cli/init_flag_precedence_test.go`, M4), each named in its §F
    milestone bullet with an explicit completion gate ("M5 is not complete
    until C39 is green"). All three are NEW files, chosen so M4/M5 deliverables
    cannot collide with M7's C37/C38 edits to the sibling test files.
  Scope note: D3-D9 (documentation hygiene) were deliberately NOT folded — the
  user scoped this fold to D1+D2. All artifacts bumped 0.2.0 → **0.2.1**.
- **Plan-auditor status:** PASS 0.89 at iteration 3 (the hard cap). No further
  audit iteration is required; the D1/D2 fold was declared re-audit-exempt.
  **Next phase: run (manager-develop).**

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Recorded by the orchestrator before the first run-phase `Agent()` spawn, per
`.claude/rules/moai/workflow/orchestration-mode-selection.md` §D.

### Input parameters

| Parameter | Value |
|-----------|-------|
| tier | L (re-tiered from M at v0.2.0) |
| scope (file count) | 22 (12 production + 7 existing test + 3 new test) |
| domain count | 3 (wizard UI/questions, config persistence in `internal/core/project`, template/config defaults) |
| file language mix | 100% Go source + Go tests |
| concurrency benefit | LOW — coding-heavy, and M5 carries a load-bearing intra-milestone ordering constraint (C34/C35/C36 are functional prerequisites of C32) |
| Implementation Kickoff Approval | obtained from the user after the review-3 PASS 0.89 verdict, on the expanded Tier L scope |

### Mode evaluation

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | 22 files, 7 milestones, architectural relocation across a package boundary — not a typo-class change |
| 2 background | no | Write-capable implementation work, not read-only analysis |
| 3 agent-team | no | RETIRED tombstone — never selectable by the decision tree |
| 4 parallel | no | Multi-domain but NOT research-heavy; this is coding work, and the Anthropic coding-task parallelism caveat routes it away from parallel fan-out |
| 5 sub-agent | **yes** | Default fallback for coding-heavy work; sequential per-milestone delegation preserves the M5 intra-milestone ordering constraint |
| 6 workflow | no | Fails the §C.3 capability gate on transformation kind: semantic/new-code work under seven distinct transform rules with cross-milestone dependencies, not one uniform mechanical transform over >=~30 independent files |

### Decision

Decision: sub-agent

### Justification

The work is coding-heavy Go implementation, which Anthropic's published guidance
routes to sequential sub-agent delegation ("most coding tasks involve fewer truly
parallelizable tasks than research"). Two SPEC-specific factors reinforce Mode 5
over any parallel shape. First, M5 carries an explicit ordering constraint the
plan states bindingly — C35 (read-patch conversion) and C36 must land before C32
(the `WritePhase1Configs` relocation), because relocating the call onto the
deployer path ahead of the read-patch conversion would expose every user to the
22,338-byte wholesale-clobber hazard. A parallel or workflow fan-out cannot honor
an intra-milestone ordering constraint. Second, the milestones are sequentially
dependent by construction: M5's persistence wiring presupposes M4's gate removal,
and M7's test reconciliation presupposes the removals in M3/M6. Mode 6 was
evaluated and rejected on transformation kind rather than file count — the 22-file
scope is below the ~30 threshold, and more decisively the work comprises seven
distinct transform rules with inter-file dependencies, which the §C.3 gate
excludes.

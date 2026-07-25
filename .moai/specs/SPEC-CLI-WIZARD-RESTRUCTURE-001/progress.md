---
id: SPEC-CLI-WIZARD-RESTRUCTURE-001
title: "Progress — moai init wizard restructure (방안 A)"
version: "0.2.1"
status: completed
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

### §E.2.1 — Commit map (15 run-phase commits, M1-M7)

Plan-phase carried 4 commits (`ab4eb4a01` initial Tier M · `5722a5d39` v0.1.1 ·
`c6341af89` v0.1.2 · `67dea24ee` v0.2.1 re-tier M→L). The run phase is the 15
commits below; the SPEC branch also carries 7 unrelated main-line commits that
are NOT part of this SPEC.

| # | Commit | Milestone | CHANGE rows landed |
|---|--------|-----------|--------------------|
| 1 | `1a2f59168` | M1 | C3, C7, C8 (3-page grouping; ungated Page 3) |
| 2 | `527a0a7cf` | M2 | C2, C9, C10, C11, C12 (model_policy High→Medium) |
| 3 | `0a037ed9f` | M3 | C4, C5, C6, C13, C15, C16, C18, C19, C21 (question removal + LSP default flip) |
| 4 | `de26b2abc` | M4 | C1, C14, C17 (advanced_bridge retired) |
| 5 | `daa940025` | M4 | C20, C30, C41 (gate 1 removed + flag-vs-wizard precedence) |
| 6 | `b03cee315` | M5 | C34, C40 (depth-aware YAML path patcher + its unit test) |
| 7 | `43b698c30` | M5 | C35 (lsp/design writers → read-patch) |
| 8 | `5920649a2` | M5 | C36 (harness.yaml dropped from the Page-3 write set) |
| 9 | `b391819f6` | M5 | C31 (gate 2 removed from WritePhase1Configs) |
| 10 | `247703d47` | M5 | C32 (Page-3 writes relocated to Step 3d — both paths) |
| 11 | `d5b2406e2` | M5 | C33 (dead `InitOptions` mode field removed) |
| 12 | `59545e585` | M5 | (chore) retired-token comment cleanup |
| 13 | `881c970f8` | M5 | C39 (deployer-path persistence integration test) |
| 14 | `c08cf24f3` | M6 | C23, C24, C25, C26, C27, C28, C29 (advanced-path full retirement) |
| 15 | `c87bbfb3a` | M7 | C22, C37, C38 (test reconciliation + AC-WIZ-015 close) |

All 41 CHANGE rows (C1-C41) are accounted for above. M5's intra-milestone
ordering constraint held in commit order: C34/C35/C36 (commits 6-8) landed
BEFORE C32 (commit 10), so the wholesale writers were never reachable against
template-deployed config (plan.md §B B-clobber).

### §E.2.2 — Per-AC record (19 ACs)

**Verification-provenance legend.** `first-hand (M7)` = the M7 agent ran the
AC's own verification procedure and observed its output. `test-bound` = the AC's
Go test binding executed inside the M7 `go test -count=1` full-suite run and
passed, but the M7 agent did NOT re-run the AC's own bespoke procedure.
`milestone-only` = verified at the named milestone; **not re-verified in M7 by
any means** — recorded here as inherited, not as an M7 observation.

| AC | Verified at | M7 provenance | Result |
|----|-------------|---------------|--------|
| AC-WIZ-001 — three topic pages, no gate | M1 | test-bound (`restructure_test.go`) | PASS |
| AC-WIZ-002 — page membership & order | M1 | test-bound (`restructure_test.go`) | PASS |
| AC-WIZ-003 — nested design condition preserved | M1 | test-bound (`TestPage3_NoModeGate`) | PASS |
| AC-WIZ-004 — locale live-render on Page 1 | M1 | test-bound (`TestBasicPage_LocaleLiveRender`) | PASS |
| AC-WIZ-005 — model_policy default = Medium | M2 | **milestone-only** (grep-bound; not re-run in M7) | PASS |
| AC-WIZ-006 — no "high" default seed remains | M2 | **milestone-only** (grep-bound; not re-run in M7) | PASS |
| AC-WIZ-007 — lsp_enabled default = true, 4-locale | M3 | **milestone-only** (CJK-punctuation grep; not re-run in M7) | PASS |
| AC-WIZ-008 — harness_profile removed | M3 | test-bound (`question_removal_test.go`) | PASS |
| AC-WIZ-009 — coverage_exemptions removed | M3 | test-bound (`question_removal_test.go`) | PASS |
| AC-WIZ-010 — Page-3 answers persist (MUST-blocking) | M5 | test-bound (`initializer_persist_test.go` Scenarios A+B) | PASS |
| AC-WIZ-010a — persistence non-destructive (MUST-blocking) | M5 | test-bound (`TestDeployerPath_PersistenceIsNonDestructive`, `..._NestedSameNamedKeysSurvive`) | PASS |
| AC-WIZ-011 — translation completeness | M3 | test-bound (`translations_completeness_test.go`) | PASS |
| AC-WIZ-012 — reconfigure order preserved | M1 | test-bound (`restructure_test.go`) | PASS |
| AC-WIZ-012a — reconfigure membership unchanged | M1 | test-bound (`TestReconfigureMembershipExcludesPage3`) | PASS |
| AC-WIZ-013 — no orphaned capture branches | M3 / M4 | test-bound (`TestAdvancedBridgeRemoved`, strengthened in M6) | PASS |
| AC-WIZ-014 — full suite + cross-platform green | M7 | **first-hand (M7)** | PASS |
| AC-WIZ-015 — advanced-path fully retired | M6 / M7 | **first-hand (M7)** | PASS |
| AC-WIZ-016 — flag beats wizard | M4 | test-bound (`init_flag_precedence_test.go`) | PASS |
| AC-WIZ-017 — depth-aware patch preserves nested keys | M5 | test-bound (`yaml_patch_test.go`) | PASS |

Tally: **19/19 PASS, 0 FAIL, 0 PASS-WITH-DEBT.** Two ACs are first-hand M7
observations; 14 are test-bound (their bindings ran green in M7's suite); 3
(AC-WIZ-005/006/007) are grep-bound and inherited from their own milestone
without M7 re-execution.

### §E.2.3 — AC-WIZ-015 retirement grep (first-hand, M7)

PRE counts measured at `881c970f8` before any M6 edit; POST measured on the M7
tree. Evidence: `.moai/state/verify/m7/ac-wiz-015.log`.

| Check | PRE | POST |
|-------|-----|------|
| `test ! -f internal/cli/wizard/advanced_gate.go` | file EXISTS | file GONE |
| `grep -rn 'IsAdvancedWizardReady\|AdvancedGate' internal/` | 13 | **0** |
| `grep -rn 'advancedMode\|standardMode' internal/` | 18 | **0** |
| `grep -rn 'StandardMode\|AdvancedMode' internal/` | 51 | **0** |
| `grep -rn 'RunWithDefaultsModes\|Phase2Questions' internal/` | 19 | **0** |
| `grep -rn '\.Bool("standard"\|\.Bool("advanced"' internal/` | 4 | **0** |
| `grep -rn '\-\-standard\|\-\-advanced' .github/` | 0 | **0** (declared non-regression guard) |

Non-vacuity is visible in the five `internal/`-scoped rows (13/18/51/19/4 → 0).
The `.github/` row is 0 on both sides by design (plan.md §B N5 — the "CI scripts
break" claim was a verified phantom); it is a guard, not evidence of work.

### §E.2.4 — M5 non-vacuity + byte-preservation record (inherited, not M7-measured)

The figures in this sub-section are carried from the **orchestrator's M5
verification record**. The M7 agent did NOT re-measure them; they are recorded
here for audit completeness with that provenance stated (per
`verification-claim-integrity.md` §2 — a carried-over number is not an M7
baseline).

**AC-WIZ-010 non-vacuity** — before the M5 chain landed, the deployer-path
assertions failed on Scenario A rows 1/2/3/5 and Scenario B row 6. Row 4
(`design_enabled=true`) is default-coincident with the shipped
`design.yaml:8 enabled: true` and is annotated in acceptance.md as a
template-default regression guard rather than a discriminating row (the review-3
D1 fold).

**AC-WIZ-010a byte-preservation** — post-init file sizes on the deployer path:

| File | Before | After | Interpretation |
|------|--------|-------|----------------|
| `lsp.yaml` | 11,306 B | 11,305 B | patched in place (`false`→`true`, −1 byte) |
| `design.yaml` | 2,867 B | 2,868 B | patched in place (+1 byte) |
| `harness.yaml` | 8,165 B | 8,165 B | untouched — C36 removed it from the write set |

The pre-C35 wholesale writers would have collapsed these to ~2-4 line documents
(~22 KB of deployed configuration destroyed per `moai init`).

**AC-WIZ-017 indentation multisets** — `enabled:` keys by indent depth after a
deployer-path init:

| File | Depth-aware (C34, actual) | Naive `patchYAMLKey` (rejected) |
|------|---------------------------|----------------------------------|
| `design.yaml` | `{2sp: 1, 4sp: 3, 6sp: 1}` | `{2sp: 5, 4sp: 0, 6sp: 0}` |
| `lsp.yaml` | `{2sp: 1, 4sp: 1}` | `{2sp: 2, 4sp: 0}` |

The naive column is the flattening the review-2 auditor's prescription would
have produced — a silent structural corruption strictly worse than the visible
clobber it was meant to fix (plan.md §B B-audit-correction 2).

**M7-added corroboration (first-hand).** M7's C37 added
`TestWritePhase1Configs_PatchesExistingFiles`, which asserts byte-exact in-place
patching at the `WritePhase1Configs` aggregate level. Its non-vacuity was proven
by temporarily forcing `writeLSPYAML` down the wholesale-write branch — the test
FAILED — after which the source was restored byte-identical
(`git diff --stat` clean).

### §E.2.5 — M7 quality gate (first-hand)

| Check | Command | Result | Evidence |
|-------|---------|--------|----------|
| Scoped suite | `go test -count=1 ./internal/cli/... ./internal/template/... ./internal/config/... ./internal/core/project/...` | exit 0, **21 ok, 0 FAIL** | `.moai/state/verify/m7/1-suite.log` |
| Host build | `go build ./...` | exit 0 | — |
| Cross-platform build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 | — |
| Lint | `golangci-lint run --timeout=3m` | exit 0, `0 issues.` (baseline 0 → zero NEW) | `.moai/state/verify/m7/3-lint.log` |
| Coverage `internal/cli/wizard` | `go test -count=1 -cover` | **93.9%** (floor 85%; M5 baseline 93.5%) | `.moai/state/verify/m7/2-cover.log` |
| Coverage `internal/core/project` | `go test -count=1 -cover` | **88.7%** (floor 85%; M5 baseline 88.7% — no regression) | `.moai/state/verify/m7/2-cover.log` |
| Subagent boundary `internal/cli` | canonical grep (excl. `_test.go` + comment lines) | 32 at `881c970f8` → **32** now (delta 0) | — |
| Subagent boundary `internal/core/project` | same | 0 → **0** | — |
| Retired-token reintroduction | `grep -rn 'StandardMode\|AdvancedMode' internal/ --include='*.go'` | **0** — no literal token reintroduced in any comment | — |

### §E.2.6 — Orchestrator independent verification (at `c87bbfb3a`)

Re-run independently by the orchestrator after the M7 commit, over the **full
repository** rather than the four scoped trees:

```
go build ./...                            exit 0
go vet ./...                              exit 0
go test -count=1 ./...                    exit 0, 107 packages ok, 0 FAIL
GOOS=windows GOARCH=amd64 go build ./...  exit 0
golangci-lint run --timeout=3m            exit 0, "0 issues."
git status --short                        clean
```

The orchestrator also independently re-ran all 7 AC-WIZ-015 retirement greps and
confirmed `advanced_gate.go` is gone and all six token greps return 0.

## §E.3 Run-phase Audit-Ready Signal

- **`run_status: audit-ready`** — run-phase implementation complete across all
  seven milestones (M1-M7). No milestone remains open; no CHANGE row (C1-C41) is
  unlanded.
- **`run_complete_at: 2026-07-25`**
- **`run_commit_sha: c87bbfb3a`** — the final run-phase commit
  (`test(SPEC-CLI-WIZARD-RESTRUCTURE-001): M7 test reconciliation + AC-WIZ-015
  close (C22/C37/C38)`). No placeholder backfill is required: this signal is
  written in a later commit than the one it names.
- **`run_phase_commits: 15`** (M1×1, M2×1, M3×1, M4×2, M5×8, M6×1, M7×1) —
  enumerated in §E.2.1.
- **`ac_pass_count: 19` / `ac_fail_count: 0` / `ac_pass_with_debt_count: 0`** —
  per the §E.2.2 matrix. Provenance is mixed by design and stated per-row there:
  2 first-hand M7, 14 test-bound, 3 grep-bound milestone-only.
- **`cross_platform_build`**: host `go build ./...` exit 0; `GOOS=windows
  GOARCH=amd64 go build ./...` exit 0. Confirmed independently by the
  orchestrator at `c87bbfb3a` over the full repo.
- **`new_warnings_or_lints_introduced: 0`** — `golangci-lint run --timeout=3m`
  reports `0 issues.` against a 0-issue baseline.
- **`coverage`**: `internal/cli/wizard` 93.9%, `internal/core/project` 88.7%,
  `internal/template` 84.9% (unchanged from baseline). Both floored packages are
  above the 85% threshold.
- **`subagent_boundary_delta: 0`** — `internal/cli` 32→32, `internal/core/project`
  0→0 (canonical grep form).
- **`push_state: NOT pushed`** — Tier L routes to Route B (PR); `manager-git`
  owns branch/PR creation. Branch `feat/SPEC-CLI-WIZARD-RESTRUCTURE-001`,
  working tree clean.
- **`preserve_list_post_run`** — the plan.md §A.5 PRESERVE list holds:
  `GitQuestions()` + reconfigure splice, the three `ModelPolicy*` const
  definitions, the locale live-render mechanism, `patchYAMLKey` + its two
  existing callers (untouched — C34 is additive),
  `writeQualityExpansionYAML`'s append-if-absent branch, and
  `generateConfigsFallback`'s other writes. No PRESERVE entry was modified.

### §E.3.1 — Residual debt inherited by sync-phase

Recorded so sync-phase inherits these rather than rediscovering them.

1. **`writeHarnessProfileYAML` retained as dead code.** C36 removed its only
   production call (satisfying AC-WIZ-010a — `harness.yaml` is left
   byte-identical), but the function and its two tests
   (`TestWriteHarnessProfileYAML`, `TestWriteHarnessProfileYAML_DefaultProfile`)
   remain. Deleting the function forces deleting those tests, which are NOT on
   the plan.md §G delete-list — and §G states that deleting a test outside that
   list "requires a new §G entry, not a judgement call at run-phase." Removal is
   pure hygiene with zero AC impact; it needs a §G amendment, not a run-phase
   decision. `golangci-lint`'s `unused` check does not fire (the function is
   test-referenced).
2. **`patchYAMLPathValue` is a line-walker, not a YAML parser.** The C34 helper
   walks lines and tracks indentation depth; it does not parse YAML. Block
   scalars (`|`, `>`) whose content happens to look like a matching key path are
   a known limitation. design.md §D2 option C (a real YAML round-trip) is the
   documented future fix. No current config file exercises the limitation.
3. **Key-absent behavior is a silent no-op.** When a target path is absent from
   an existing document, `writeLSPYAML` / `writeDesignYAML` leave the file
   byte-identical rather than appending a duplicate top-level mapping. This is
   deliberate (it protects hand-edited config), but it means a user who deletes
   `lsp.enabled` from their `lsp.yaml` silently stops receiving the wizard
   answer, with no warning.
4. **`TestInit_QualityYAMLContent` behavior change.** A caller that leaves
   `opts.EnforceQuality` unset now receives `false` where the template default
   was `true`. No production path is exposed — `init.go` always sets the field
   via `getBoolFlagWithDefault(cmd, "enforce-quality", true)` — so this is a
   direct-API-caller concern only.
5. **Removal tombstone comments.** Four test files
   (`questions_test.go`, `wizard_test.go`, `expansion_test.go`,
   `unified_form_test.go`) retain comments naming `advanced_bridge` /
   `harness_profile` / `coverage_exemptions_enabled` to document what §G
   authorized removing. They carry no AC-WIZ-015 token and cannot defeat that
   grep, but a future repo-wide grep on those literals would surface them as
   false positives.
6. **docs-site 4-locale flag references — the largest inherited item.**
   Twelve docs-site files (4 locales × `cli-reference/init.md` / `cli.md` /
   the init-wizard doc) still document `--standard` / `--advanced`, which no
   longer exist. This is scoped OUT of run-phase by design (spec.md §C Out of
   Scope, REQ-WIZ-019, plan.md §F sync-phase note) and OUT of AC-WIZ-015, whose
   scope is CODE + CI only. **It is a sync-phase deliverable owned by
   manager-docs**, bound by the CLAUDE.local.md §17 4-locale parity obligation.
   Run-phase deliberately did not touch `docs-site/`.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: complete
sync_complete_at: 2026-07-25
sync_commit_sha: adee2f46b
```

The three fields above are written as bare `key: value` lines inside a fenced
block because that is the only form `internal/spec/era.go` `extractProgressField`
can read. The first sync report wrote them as bold-and-backtick-wrapped list
items — a leading dash, then bold markers, then the whole "key: value" pair
wrapped in a single pair of backticks. Neither `progressFieldYAMLPattern` nor
`progressFieldListPattern` matches that shape, because the colon sits inside the
backticks rather than after the key. `moai spec audit --json` consequently classified this SPEC
as `era: V3R5` via `H-3 (§E.2 present, sync_commit_sha missing)`, which would have
placed it under the grandfather clause and exempted it from lifecycle drift
detection permanently. Found by the sync-phase audit (finding F1) and corrected
here; every other closed SPEC in the catalogue uses the bare form.

`sync_commit_sha` names `adee2f46b` and was backfilled in a later commit per the
SHA placeholder backfill exemption (`spec-frontmatter-schema.md` § Status
Transition Ownership Matrix) — the sync commit cannot reference its own hash.

- **`base_merge`**: at sync entry the branch was ahead of origin/main by 21
  commits and origin/main was ahead by 0
  (`git rev-list --count --left-right origin/main...HEAD` → `0 21`, i.e.
  local-ahead-only; pre-observed baseline at 96d35723c, per the orchestrator's
  pre-verified baseline). Before the merge the same command reported `9 20`
  (origin/main ahead by 9, branch ahead by 20).
  The worktree HEAD (`96d35723c`) already carries the merged state — 9 origin/main
  commits merged with 0 conflicts; the only file both sides touched was
  `internal/template/model_policy.go`, in non-overlapping regions (confirmed by
  the orchestrator's pre-verified baseline: `go build ./...` exit 0, `go vet ./...`
  exit 0, `go test -count=1 ./...` exit 0, 105 packages ok, 0 FAIL before this
  session's edits began).
- **`docs_site_files_edited: 12`** — the 4-locale × 3-doc family:
  `docs-site/content/{en,ko,ja,zh}/getting-started/init-wizard.md` (3-page rewrite),
  `docs-site/content/{en,ko,ja,zh}/getting-started/cli.md` (flag-row removal),
  `docs-site/content/{en,ko,ja,zh}/cli-reference/init.md` (flag-row removal +
  heading rename "Wizard phases" → "Wizard question flags").
- **`locale_parity`**: heading-count and table-row-count parity verified per file
  family across en/ko/ja/zh — `init-wizard.md` 28 headings / 10 table rows each;
  `cli.md` 44 headings / 161 table rows each; `cli-reference/init.md` 14 headings /
  37 table rows each (command: `grep -cE '^#{1,3} '` / `grep -cE '^\|'` per locale,
  see Evidence below).
- **`changelog_entry_position`**: `CHANGELOG.md` `## [Unreleased]` → `### Changed`,
  first bullet (immediately above the `SPEC-SUBAGENT-NESTING-DOCTRINE-001` entry).
- **`frontmatter_status_transitions`**: `spec.md` / `plan.md` / `acceptance.md` /
  `progress.md` — all 4 transitioned `in-progress → completed` in this sync commit
  (the merged 3-phase close; `updated:` already carried `2026-07-25` from run-phase
  and is unchanged by this transition).
- **`push_state: NOT pushed`** — Tier L routes to Route B (PR); `manager-git` owns
  branch/PR creation. Working tree in worktree `.claude/worktrees/wizard`,
  branch `feat/SPEC-CLI-WIZARD-RESTRUCTURE-001`.

### Evidence (verbatim command + output, this run)

```
$ hugo --gc --minify   (from docs-site/)
exit=0
Total in 1830 ms
KO 153 pages / EN 151 / JA 151 / ZH 151 — 0 WARN, 0 ERROR lines
log: .moai/state/verify/wizard-sync/1-hugo-build.log

$ grep -rnE -- '--standard|--advanced' docs-site/content/
exit=1 (0 matches; baseline was 12 files / 44 matches)
log: .moai/state/verify/wizard-sync/2-flag-grep.log

$ grep -rE -- '--standard|--advanced' .github/
exit=1 (0 matches — unchanged, was already 0)
log: .moai/state/verify/wizard-sync/3-github-grep.log

$ go test -count=1 ./...   (from worktree root)
exit=0, 105 packages ok, 0 FAIL
log: .moai/state/verify/wizard-sync/6-go-test.log

$ grep -c 'SPEC-CLI-WIZARD-RESTRUCTURE-001' CHANGELOG.md   (B12 pre-emission self-test)
0 (before this session's edit — safe to emit, no duplicate)

$ grep -cE '^\| \*\*AC-[A-Z]+-[0-9]+\*\*' acceptance.md
0 (acceptance.md uses '### AC-WIZ-NNN' headings, not the '| **AC-...**' table
   row form; AC count verified instead via
   'grep -cE "^### AC-WIZ-[0-9]+" acceptance.md' style enumeration = 17 distinct
   IDs + 2 lettered sub-IDs (010a, 012a) = 19, matching progress.md §E.3's
   recorded "19/19 PASS, 0 FAIL, 0 PASS-WITH-DEBT" tally)
```

### Gaps (not observed this run)

- `git show --stat <sync-commit-sha>` — not yet run; will be captured after the
  commit lands and cited in the final report.
- No independent re-verification of the run-phase coverage/lint numbers
  (93.9% / 88.7% / 84.9%, golangci-lint 0 issues) — those are cited from
  progress.md §E.3 (run-phase's own attributed measurement), not re-run in
  this sync session, since sync-phase scope is docs-site + CHANGELOG +
  frontmatter, not a run-phase re-audit.

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

## §G Sync-phase Audit Record

Independent 4-dimension audit (`sync-auditor`) run against `da85ad13e`, after the
sync artifacts landed and before PR creation. A second independent re-audit ran
against `209c8f9f9`, after the round-1 remediation commits landed. Both report
files are committed at the paths cited below (the directory is tracked; not
gitignored).

```yaml
sync_audit_verdict: FAIL
sync_audit_score: 0.745
sync_audit_threshold: 0.85
sync_audit_report: .moai/reports/sync-audit/SPEC-CLI-WIZARD-RESTRUCTURE-001-2026-07-25.md
```

Dimension scores (harmonic mean, not arithmetic): Functionality 0.80 · Security
0.70 · Craft 0.82 · Consistency 0.68. The FAIL is **score-driven** — no must-pass
criterion failed, and the auditor independently reproduced all 19 ACs (including
an end-to-end `moai init` against the built binary, matching the §E.2.4 byte
sizes exactly) and confirmed the 4-locale parity is semantic rather than
line-count-only.

### Round-2 re-audit (after round-1 remediation)

```yaml
sync_audit_verdict_reaudit: FAIL
sync_audit_score_reaudit: 0.834
sync_audit_threshold: 0.85
sync_audit_report_reaudit: .moai/reports/sync-audit/SPEC-CLI-WIZARD-RESTRUCTURE-001-2026-07-25-reaudit.md
```

Dimension scores (harmonic mean): Functionality 0.88 · Security 0.82 · Craft
0.84 · Consistency 0.80. Five of the seven round-1 findings were independently
re-verified fixed (F1, S1, F4/F2-CHANGELOG-half, F5, F6). The FAIL persisted for
two distinct reasons: the two user-deferred findings (F3, S2) held Security and
Functionality below the bar; and a second cluster the round-1 fix had not
reached — the docs-site half of F2, C2, F7, and four new findings the
remediation itself introduced or exposed (N3-N8) — held Consistency and Craft
down. This re-audit is what surfaced N3 through N8.

### Findings fixed (round 1, before first close attempt)

| ID | Severity | Finding | Resolution |
|----|----------|---------|------------|
| F1 | critical | `sync_commit_sha` was written bold-and-backtick-wrapped, which `internal/spec/era.go` cannot parse; `moai spec audit --json` reported `era: V3R5` via `H-3 (§E.2 present, sync_commit_sha missing)`, placing the SPEC under the grandfather clause and exempting it from lifecycle drift detection permanently | §E.4 rewritten with a bare-`key: value` fenced block; re-verified `era: V3R6` via `H-4 (§E.2 + §E.4 + sync_commit_sha)` |
| S1 | major | `--project-mode` reached the project-mode YAML writer unvalidated while all seven sibling flags on the same command validate; C32 of this SPEC is what made that write path reachable from `moai init`, so the exposure is SPEC-introduced | `validateInitFlags` now enforces the `personal`/`team` enum, with an injection-reproduction test (commit `0e34ac7a3`) |
| F5 | minor | `design.md` / `research.md` left at `status: in-progress` on a `completed` SPEC | both transitioned to `completed` |
| F6 | minor | CHANGELOG cited `internal/template` coverage 84.9%; the post-merge tree measures 86.0% | CHANGELOG now states the measured value with its baseline attribution |
| F2 (CHANGELOG half) / F4 | major (doc) | CHANGELOG claimed `--harness-profile` "remains selectable"; the flag is parsed but its only writer lost its last production caller in C36, so it persists nothing | CHANGELOG corrected to state the flag has no persisted effect and is declared dead code (debt (a)) |

### Findings fixed (round 2, this repair pass)

| ID | Severity | Finding | Resolution |
|----|----------|---------|------------|
| F3 | major | `--enable-lsp` cobra default was `false` while the wizard default flipped to `true` (REQ-WIZ-010), so `moai init --non-interactive` yielded `lsp.enabled: false` | `internal/cli/init.go` `LSPEnabled` now reads `getBoolFlagWithDefault(cmd, "enable-lsp", true)`; `moai init --non-interactive` (no flags) now persists `lsp.enabled: true`; explicit `--enable-lsp=false` unaffected |
| C2 | major | `init.go`'s Page-3 override comment ("defaults match wizard defaults") was false for `LSPEnabled` | Resolved automatically by the F3 fix — the comment is true again |
| F7 | minor | `init-wizard.md` said the LSP default is "enabled (Yes)" while `cli.md`/`cli-reference/init.md` said "(default: false)" (all 4 locales) | Resolved by the F3 fix; both `cli.md` and `cli-reference/init.md` now state "(default: true)" in all 4 locales |
| F2 (docs half) | major | The 4-locale docs-site pages this SPEC's own sync phase rewrote still taught the inert `--harness-profile` flag in the non-interactive CI example, and 8 flag-table rows presented it as a normal flag | `--harness-profile default \` removed from the non-interactive example in all 4 `init-wizard.md` files; the 8 flag-table rows in `cli.md`/`cli-reference/init.md` now state the flag is accepted but currently has no persisted effect |
| N3 | minor | The S1 validator fix (`0e34ac7a3`) rejects `enterprise` for `--project-mode`, but `project.yaml.tmpl`'s own comments still documented it as an accepted value — a regression introduced by the S1 fix narrowing accepted input against the template's own documentation | `enterprise` removed from `project.yaml.tmpl:11,14` comments (commit `580d49959`); enum stays `personal|team` |
| N4 | minor | The S2 debt row and CHANGELOG debt (g) claimed `internal/atomicfile.Replace` "ships unused" — false repo-wide (10 production call sites across 7 files) | Restated below and in the CHANGELOG: the helper is used elsewhere; it is the config writers in `initializer_expansion.go` that do not use it |
| N5 | minor | The CHANGELOG coverage sentence said the run-phase (84.9%) and post-merge (86.0%) figures were "above" the package's own unchanged baseline; both are in fact exactly equal to their respective baselines, not above them | CHANGELOG restated as "unchanged from" |
| N6 | minor | §G cited `.moai/reports/sync-audit/SPEC-CLI-WIZARD-RESTRUCTURE-001-2026-07-25.md` as evidence, but the file was untracked and in no commit | Both the round-1 report and this round-2 re-audit report are committed (the `sync-audit/` directory is tracked; sibling reports already exist there) |
| N7 | minor | §G's "Findings fixed / deferred" tables omitted three live findings (C2, F7, S3) while presenting themselves as the complete resolution record | This §G revision adds C2 and F7 to the fixed table above and S3 to the deferred table below |
| N8 | minor | CHANGELOG debt (f)/(g) and this section routed the deferred debt to "a follow-up SPEC" that does not exist | Reworded below to "deferred, unscheduled" rather than naming a nonexistent SPEC |

### Findings deferred (user-scoped)

| ID | Severity | Finding | Why deferred |
|----|----------|---------|--------------|
| S2 | major | Seven config writers use non-atomic `os.WriteFile` on the ~22 KB deployed config set. (Correction, N4: `internal/atomicfile.Replace` is NOT unused repo-wide — it has 10 production call sites across `internal/cli/preference/filestore.go`, `internal/harness/tier/tier.go`, `internal/migration/version.go`, `internal/hook/handoff/persist.go`, `internal/goal/state.go`, `internal/session/registry.go`, and `internal/session/store.go` (×4); it is only the *config writers* in `initializer_expansion.go` that do not route through it.) | A cross-cutting writer refactor beyond this SPEC's scope; would widen the PR substantially. Deferred and currently unscheduled — no follow-up SPEC exists yet (N8) |
| S3 | minor | `initializer_expansion.go:63,118,149,200` reads the patch target with an unbounded `os.ReadFile` (line 63 is `writeProjectModeYAML`, the reader C32 newly exposed on the deployer path); bounded in practice by deployed template size, no explicit limit | Defense-in-depth gap, not a live issue at current template sizes; deferred alongside S2 |
| S1-residual | minor | `writeProjectModeYAML` / `patchYAMLKey` interpolate their value raw — the injection invariant lives in the two current callers (the now-validated `--project-mode` flag and the wizard's constrained select), not in the writer itself. No live injection path exists today with the current producer set, but restoring `--harness-profile`'s writer (debt (a)) without adding validation would reopen the identical hole `validateInitFlags` closed for `--project-mode` | Defense-in-depth gap; the writer-level fix is deferred alongside S2/S3 |

The deferred findings are recorded in the CHANGELOG entry as debt (g), (h), (i).

### Round 3 re-audit — PASS

```yaml
sync_audit_round3_verdict: PASS
sync_audit_round3_score: 0.858
sync_audit_round3_threshold: 0.85
sync_audit_round3_report: .moai/reports/sync-audit/SPEC-CLI-WIZARD-RESTRUCTURE-001-2026-07-25-reaudit-2.md
```

Dimension scores: Functionality 0.92 (+0.04) · Security 0.82 (unchanged) · Craft 0.84
(unchanged) · Consistency 0.86 (+0.06). All nine round-2 findings verified RESOLVED by
reproduction rather than by reading the commit messages — including three binary
reproductions of the `--enable-lsp` default across flag variants and a re-reproduction of
the `--project-mode` injection rejection.

Security did not move: the shipped security surface is byte-identical to round 2 (the three
intervening commits touch the LSP seed, a template comment, and markdown only). Roughly
0.11 of the 0.18 Security deficit is the user-deferred S2 — that portion is structural and
not closable without reversing the deferral.

Four NEW record-accuracy findings were raised and are fixed here:

| ID | Severity | Finding | Resolution |
|----|----------|---------|------------|
| N9 | major (doc) | The CHANGELOG still read "above the package's own unchanged baseline" for the coverage figures, while both this section and the CHANGELOG's own repair narrative claimed that wording had already been corrected — a claimed correction that was never applied. Re-measured independently via `git archive`: base 84.9% = run-phase 84.9%, origin/main 86.0% = HEAD 86.0%; the figures are EQUAL to baseline, not above it | CHANGELOG now states each figure equals its own tree point's baseline |
| N10 | minor | The `atomicfile.Replace` breakdown listed `internal/session/store.go` ×3; the file carries 4 calls, so the per-file breakdown summed to 9 against a stated total of 10 | corrected to ×4 in both surfaces; 6 files × 1 + store ×4 = 10 |
| N11 | minor | "closes finding (f) above" pointed at a debt item deleted in the same commit, leaving a dangling label; a second `debt (a) below` reference pointed backwards | the (f) reference is replaced by a description of what was closed and why no `(f)` remains; the misdirected `(a)` reference now points above |
| N12 | minor | The S3 row cited `initializer_expansion.go:118,200`, omitting the same unbounded read at lines 63 and 149 — line 63 being `writeProjectModeYAML`, the reader C32 newly exposed on the deployer path | S3 now cites all four sites and names line 63 explicitly |

Two round-3 observations recorded for honesty rather than acted on:

- The auditor corrected **its own** round-2 measurement: it had reported 5 `atomicfile.Replace`
  call sites; the true count is 10 across 7 files. The round-2 record inherited that
  undercount, and the N4 correction written into this section had already used the
  independently re-measured 10.
- The auditor's N11 claim covered two `debt (a)` references. Only one was actually
  misdirected — verified by character offset: the reference at CHANGELOG line 28 precedes
  the `(a)` definition and is correct as written; the one in the later paragraph followed it
  and was wrong. Only the wrong one was changed.

Reporting-error rate across the three rounds: 2 → 2 → 4. The rate is not declining even as
the code fixes hold up cleanly under adversarial re-verification, which is itself the most
useful signal this audit chain produced: on this SPEC, the prose about the work has been a
less reliable artifact than the work.

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

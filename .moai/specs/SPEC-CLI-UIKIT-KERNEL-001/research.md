# Research — SPEC-CLI-UIKIT-KERNEL-001

Deep codebase analysis of the 4 uikit-kernel candidate files, the 14-file kernel
survey from SPEC-CLI-SUBPKG-SPLIT-001 research.md §C.4.1 (D6 iter-1 fix: §C.1 is
"Import-cycle hazard"; the 14-file kernel survey is at §C.4.1), and the three dominant
risks (caller-rewrite blast radius + leaf-contract correctness with cross-file type
dependencies + the SettingsLocal cycle hazard that forces `settings.go` STAYS per D2).
All quantities below were measured (not assumed) against the working tree at
plan-authoring time via `wc`/`grep`/`go build`.

This file EXPANDS SPLIT-001 research.md §C.4.1 (the 14-file kernel survey). It
cross-references SPLIT-001 §A (baseline), §C.2 (test surface — N/A for uikit
because the uikit MOVED source is only 365 LOC after D2 STAYS reclassification),
§C.3 (secondary risks), §D (behavior-preservation framing), §E (continuity),
§F (recommendation basis).

## §A. Baseline Measurements (observed this run, 2026-07-07)

| Metric | Value | Command (verbatim) |
|--------|-------|--------------------|
| 4 source files LOC (render/banner/settings/schema_bridge) | **503** | `wc -l internal/cli/{render,banner,settings,schema_bridge}.go` → 120+146+138+99 |
| 3 MOVED source files LOC (settings.go STAYS per D2) | **365** | `wc -l internal/cli/{render,banner,schema_bridge}.go` → 120+146+99 |
| Root non-test `.go` files (current tree) | **95** | `find internal/cli -maxdepth 1 -name '*.go' ! -name '*_test.go' \| wc -l` |
| Root non-test LOC (current tree) | **25,033** | `find internal/cli -maxdepth 1 -name '*.go' ! -name '*_test.go' -exec cat {} + \| wc -l` |
| Root test files (current tree) | **151** | `find internal/cli -maxdepth 1 -name '*_test.go' \| wc -l` |
| Root test LOC (current tree) | **53,736** | `find internal/cli -maxdepth 1 -name '*_test.go' -exec cat {} + \| wc -l` |
| 14 kernel files LOC (sum) | **7,022** | `wc -l internal/cli/{banner,hook,init,init_layout,inventory,launcher,migrate_agency,profile_setup_translations,render,research,root,schema_bridge,settings,update}.go` |
| `update.go` LOC | **3,173** | `wc -l internal/cli/update.go` |
| Build baseline (green?) | **exit 0** | `go build ./...` |
| uikit directory exists? | **no** | `ls internal/cli/uikit 2>&1` → "No such file or directory" |
| Existing uikit references in `internal/` | **0** | `grep -rn 'internal/cli/uikit' internal/` → no match (cycle pre-check clean) |
| `SettingsLocal` type definition location | **`internal/cli/glm.go:97`** | `grep -nE 'type SettingsLocal' internal/ -r --include='*.go'` (D2 iter-1 BLOCKING fix — verified this run, NOT deferred to run phase) |
| `SettingsLocal` consumers (production) | **6 sites across 3 files** | `grep -rnE '\bSettingsLocal\b' internal/cli/*.go \| grep -v _test.go` → glm.go:96/97/513/590/940, launcher.go:251/635/662, settings.go:20/28/48/110 |
| `CheckStatus`/`CheckOK`/`CheckWarn`/`CheckFail` references (production) | **13 sites across 3 files** | `grep -cE '\bCheck(Status\|OK\|Warn\|Fail)\b' internal/cli/doctor.go internal/cli/doctor_cache.go internal/cli/doctor_harness.go` (D1 iter-1 BLOCKING fix) |
| `CheckStatus`/etc references (test) | **34+ sites across 5 test files** | `grep -cE '\bCheck(Status\|OK\|Warn\|Fail)\b' internal/cli/{mcp_doctor_coverage,coverage,coverage_fixes,coverage_improvement,doctor_golden}_test.go` (D1 iter-1 BLOCKING fix) |

### §A.1 Baseline drift attribution (verification-claim-integrity §2)

SPLIT-001 original baseline (2026-07-02 authoring): 93 root non-test files / 25,838 LOC.
SPLIT-001 drift re-verification (2026-07-07 pre-audit, pre-M1-close): 98 / 26,440 (Δ +5 / +602).
SPLIT-001 M1 close (sync_commit_sha `d0d9b49d7`) extracted the agentlint cluster: 3 non-test
files + 2 test files moved from depth 1 to depth 2 (`internal/cli/agentlint/`).
THIS SPEC measurement (2026-07-07, post-M1-close): 95 / 25,033.

Reconciliation: 98 − 3 = 95 ✓ (the 3 agentlint non-test files no longer count at depth 1). The
LOC drop from 26,440 to 25,033 (−1,407) is the agentlint cluster LOC relocating to depth 2 plus
ordinary churn in the 5 days since SPLIT-001 authoring.

### §A.2 Cycle pre-check (the leaf contract feasibility)

The 4 source files were grepped for `package cli` coupling:

```
$ grep -nE 'deps\b|cli\.' internal/cli/render.go internal/cli/banner.go internal/cli/settings.go internal/cli/schema_bridge.go
(no matches — none of the 4 source files touch the package-cli global deps NOR reference cli.* symbols)
```

Imports observed:

```
render.go:        fmt, strings, lipgloss, internal/tui
banner.go:        fmt, os, runtime, strings, lipgloss, internal/tui
settings.go:      encoding/json, fmt, os, path/filepath
schema_bridge.go: internal/settings
```

All imports are EXTERNAL to `internal/cli/*`. The leaf contract is feasible for render.go,
banner.go, schema_bridge.go. **HOWEVER, settings.go references the `SettingsLocal` type defined
at `glm.go:97`** (verified via `grep -nE 'type SettingsLocal' internal/ -r --include='*.go'`) —
the type is defined in package cli, NOT in an external package. Moving `settings.go` into uikit
would force uikit to import package cli for the type → import cycle. **D2 iter-1 BLOCKING fix**:
`settings.go` STAYS in package cli (option a — see design.md §C.3). The leaf narrows to 3 source
files (render.go + banner.go + schema_bridge.go, 365 LOC). (The cross-file type deps in §B
below are SEPARATE — they reference types defined in OTHER `internal/cli/*.go` files, not the
package-cli global.)

## §B. The 14-File Kernel Survey — Classification (MOVED / STAYS / GATED)

SPLIT-001 research.md §C.4.1 surveyed the 14 `internal/cli/*.go` non-test files that use at
least one of the 9 kernel-helper symbols (`renderCard`, `renderKeyValue`, `renderStatusLine`,
`RenderError`, `PrintBanner`, `printWelcomeMessage`, `mutateSettingsLocal`, `writeFileAtomic`,
`schemaKeyToTUIField`). The 14 files:

```
banner.go, hook.go, init.go, init_layout.go, inventory.go, launcher.go,
migrate_agency.go, profile_setup_translations.go, render.go, research.go,
root.go, schema_bridge.go, settings.go, update.go
```

This SPEC classifies each into exactly one of three buckets:

### §B.1 MOVED into uikit (3 files / 365 LOC — D2 STAYS reclassification applied)

These are the source files whose helpers ARE the kernel surface. They move to uikit (with their
helpers re-scoped — see design.md §C for the per-file inventory). **`settings.go` was MOVED in
the iter-1 plan; iter-2 reverses this per D2 option (a)** — it references `SettingsLocal`
(glm.go:97) and would create an import cycle if moved. See §B.2 STAYS row for settings.go.

| File | LOC | Why MOVED |
|------|----:|-----------|
| `render.go` | 120 | Defines the render helpers (RenderError, renderCard, renderKeyValue, renderKeyValueLines, renderStatusLine, renderSuccessCard, renderInfoCard, renderSummaryLine) + the kvPair supporting type. The kernel surface itself. |
| `banner.go` | 146 | Defines PrintBanner + PrintWelcomeMessage + the version-detection helpers that feed them. |
| `schema_bridge.go` | 99 | Defines schemaKeyToTUIField + fieldDefTUILabel + the schemaFieldBridge/schemaSegmentBridge maps. **CONDITIONAL** on profileSetupText resolution (design.md §C.4 D5 fix — both maps reference profileSetupText; b-ii split keeps both in package cli). |

### §B.2 STAYS in `package cli` (kernel-USING files — 8 files including settings.go per D2)

These files USE kernel helpers but their primary purpose is NOT to be a kernel helper (OR moving
them would create a cycle, as with settings.go). They stay in package cli and their kernel-helper
calls rewrite to `uikit.<Helper>(...)`.

| File | LOC | Why STAYS | Caller rewrite? |
|------|----:|-----------|-----------------|
| `hook.go` | 1,181 | Hook command surface — deps-coupled (`hook.go`, `hook_pre_push.go` per SPLIT-001 §C.3). Stays per SPLIT-001 §E (deps/platform-tangled clusters out of scope). | YES — 5 call sites (§D.2-§D.4 of design.md) |
| `init.go` | 479 | `moai init` command. Part of the root command surface. | YES — 4 call sites |
| `init_layout.go` | 83 | Init next-steps layout (replaces PrintBanner full-screen). | YES — 3 call sites |
| `inventory.go` | 332 | `moai inventory` command (deps-coupled per SPLIT-001 §C.3). | YES — 9 call sites |
| `launcher.go` | 778 | Launcher helper (cc/glm/cg). | NO rewrite for `mutateSettingsLocal`/`stripGLMCredsAndSetTeammateMode` — these STAY in package cli per D2 (settings.go STAYS); launcher.go:206 call site unchanged. (D2-a applied: launcher.go drops out of the blast radius under iter-2) |
| `research.go` | 150 | `moai research` command. | YES — 2 call sites |
| `root.go` | 126 | Cobra root + Execute(). The file that imports subpackages. STAYS (it IS package cli). | YES — 1 call site (PrintBanner) |
| **`settings.go`** | **138** | **D2 STAYS reclassification (iter-2 fix)**: references `SettingsLocal` type (defined `glm.go:97`, a STAYS file). Moving settings.go to uikit would create `uikit → package cli` cycle. Helpers (`mutateSettingsLocal`, `writeFileAtomic`, `stripGLMCredsAndSetTeammateMode`) remain package-cli-internal. | NO rewrite for the settings helpers themselves (they stay); the type-defining file (glm.go) is also untouched. |

### §B.3 GATED behind a future SPEC (3 files — cluster-level, not kernel-helper-level)

These files belong to clusters that will extract in FUTURE SPECs. They are NOT moved in this
SPEC. Their kernel-helper calls rewrite to `uikit.<Helper>(...)` (the rewrite IS in this SPEC's
scope — only the cluster extraction is GATED).

| File | LOC | Why GATED | Caller rewrite? | Future SPEC |
|------|----:|-----------|-----------------|-------------|
| `migrate_agency.go` | 648 | Part of the future migrate cluster (SPLIT-001 §C.4.3 tri-axis coupled). | YES — RenderError at :634 (the axis-(i) blocker) | future migrate SPEC |
| `profile_setup_translations.go` | 609 | Part of the future profile cluster (SPLIT-001 §C.4.4). | (depends on §C.4 resolution) | future profile SPEC |
| `update.go` | 3,173 | Part of the future update cluster (SPLIT-001 §B 5,181 LOC cluster). | YES — PrintBanner at :2538, PrintWelcomeMessage at :2543 | future update SPEC |

### §B.4 Plus: `doctor.go` + `doctor_cache.go` + `doctor_harness.go` (NOT in the 14-file set, but affected by CheckStatus type move — D1 iter-1 fix)

`doctor.go` was NOT in SPLIT-001's 14-file kernel-helper-user set (it does not call
`renderCard`/`renderKeyValue`/etc. — except `renderInfoCard` at :139, which IS a kernel helper
so doctor.go probably SHOULD have been in the set; possibly the survey missed it). It IS affected
because it DEFINES the `CheckStatus` type that `render.go:60` consumes. The type moves to uikit
(design.md §C.5), and doctor.go rewrites its type references.

**D1 iter-1 BLOCKING fix**: doctor.go is NOT the only CheckStatus-bearing file. The verbatim
re-grep this run surfaces TWO additional production files + 5 test files that iter-1 §B.4 missed:

| File | Type | CheckStatus refs | Action |
|------|------|-----------------:|--------|
| `doctor.go` | production | 4+ (type def at L25-34, struct field at L40, fn sigs at L152-153 + L568) | rewrite to `uikit.CheckStatus`/etc. |
| **`doctor_cache.go`** | production | 5 (L49, L60, L70, L80, L88 — `check.Status = CheckOK/CheckWarn`) | rewrite to `uikit.CheckOK`/etc. |
| **`doctor_harness.go`** | production | 4 (L24, L96, L100, L104 — `check.Status = CheckOK/CheckFail/CheckWarn`) | rewrite to `uikit.CheckOK`/etc. |
| `mcp_doctor_coverage_test.go` | test | 4 | rewrite |
| `coverage_test.go` | test | 6 | rewrite |
| `coverage_fixes_test.go` | test | 3 | rewrite |
| `coverage_improvement_test.go` | test | 12 | rewrite |
| `doctor_golden_test.go` | test | 9 | rewrite |

**Total: 43 additional CheckStatus references across 7 files** that iter-1 §B.4 did not enumerate.
This is exactly the SPLIT-001 M1-blocker pattern (cluster characterization without verbatim
file:line verification is unsound) — this SPEC cited the lesson at iter-1 §C.2 L181 but did NOT
apply it to its own CheckStatus blast-radius map. iter-2 applies it.

This is a plan-phase characterization correction: the original 14-file survey may have
under-counted (doctor.go uses `renderInfoCard` at :139, which is in the kernel-helper set; the
further CheckStatus references in doctor_cache.go + doctor_harness.go are NOT kernel-helper uses
— they are type-consumer uses, a different blast-radius axis the survey did not track).
The run phase re-surveys caller sites against the CURRENT tree (per the M1-blocker lesson —
SPLIT-001 plan.md §F Provenance).

### §B.5 Classification summary

| Bucket | Files | LOC | Action |
|--------|------:|----:|--------|
| MOVED to uikit | **3** | **365** | source files relocate (D2 STAYS: settings.go no longer MOVED) |
| STAYS in package cli (rewrite or no-rewrite) | **8** | **3,567** | callers rewrite to uikit.* (where applicable); settings.go STAYS with no rewrite (D2) |
| GATED (rewrite, no cluster move) | 3 | 4,430 | callers rewrite; cluster extraction deferred |
| Affected via CheckStatus type move | 3 production + 5 test | (part of STAYS) | CheckStatus type references rewrite |
| **Total** | **14 + 3 (CheckStatus prod) + 5 (CheckStatus test)** | **8,362 + CheckStatus-bearing** | |

## §C. The Three Dominant Risks for uikit (the crux of the M1 value/risk assessment — iter-2 expanded)

### §C.1 Caller-rewrite blast radius (regression risk dominant)

The 3 MOVED source files are 365 LOC, but the helpers they define are WIDELY USED across the
package cli root. The verbatim caller-rewrite map (design.md §D) identifies **12 production
files + ≥10 test files** that must rewrite to `uikit.<Helper>(...)` or to `uikit.CheckStatus` /
`uikit.CheckOK` / `uikit.CheckWarn` / `uikit.CheckFail`. This is the dominant risk: each rewrite
site is a place a compile-break or silent behavior change could land.

The risk is mitigated by:
- **Characterization preservation** — the existing 53,736 test LOC + `moai --help` snapshot are
  the behavior contract. Every rewrite site is verified by AC-CUK-001 (tests green) + AC-CUK-003
  (`moai --help` diff empty).
- **Caller-rewrite completeness gate** — AC-CUK-008 greps for residual pre-move symbol
  references; any residue blocks M1.
- **Atomic single-commit** — all rewrites land in one commit (AC-CUK-011), so a partial-rewrite
  state is never shipped.

But the risk is REAL: 12 production files + ≥10 test files is wider than SPLIT-001 M1 agentlint
(which rewrote only `root.go` for registration). The blast-radius characterization (design.md §D)
is the input to the §F CHECKPOINT decision (REQ-CUK-010) — if the run-phase finds the cost
uneconomic, STOP.

### §C.2 Cross-file type dependencies (architectural risk — the M1-blocker analogue — THREE, not two)

THREE cross-file type dependencies MUST be resolved BEFORE the source files move (REQ-CUK-007):

1. **`CheckStatus` (defined `doctor.go:25-26`, consumed `render.go:60` AND 6 other files)** — if
   render.go moves without the type co-locating, uikit either fails to compile (the type is
   undefined in uikit's scope) OR creates a cycle (uikit imports doctor.go — but doctor.go is in
   package cli, and uikit cannot import package cli). The resolution is clean (§C.5 of design.md):
   the type is a generic status enum, it BELONGS in uikit. **D1 iter-1 BLOCKING fix**: every
   CONSUMER of the type rewrites — 3 production files (doctor.go, doctor_cache.go,
   doctor_harness.go) + 5 test files (mcp_doctor_coverage_test.go, coverage_test.go,
   coverage_fixes_test.go, coverage_improvement_test.go, doctor_golden_test.go). Total 43
   references across 7 files.

2. **`profileSetupText` (defined `profile_setup_translations.go:10`, consumed by TWO maps in
   `schema_bridge.go:24-77`)** — this is the harder one. **D5 iter-1 MINOR fix**: schema_bridge.go
   declares TWO profileSetupText-referencing maps (`schemaFieldBridge` L24-58, 19 entries;
   `schemaSegmentBridge` L60-77, 16 entries). iter-1 said "L24-65 one map" — the second map was
   missed. Both maps' entries reference `profileSetupText` at every signature. Resolution is the
   **b-ii split** (design-time decision, NOT run-phase): BOTH maps stay in package cli (e.g. new
   file `schema_bridge_profile.go`); only `SchemaKeyToTUIField` + `FieldDefTUILabel` helpers
   move to uikit. The b-i alternative (co-locate profileSetupText into uikit) is rejected — it
   pulls i18n data into the kernel leaf.

3. **`SettingsLocal` (defined `glm.go:97`, consumed by `settings.go:28/48/110`)** — **D2 iter-1
   BLOCKING fix**: this third cross-file type dependency was MISSED by iter-1 (which deferred to
   "run phase locates it"). Verified this run: `SettingsLocal` is defined at `glm.go:97` (a STAYS
   file). Moving `settings.go` into uikit would force uikit to import package cli for the type →
   IMPORT CYCLE → defeats REQ-CUK-001 leaf contract. **Resolution is option (a) STAYS
   reclassification** (design-time decision, NOT run-phase): `settings.go` does NOT move to uikit.
   The settings helpers (`mutateSettingsLocal`, `writeFileAtomic`, `stripGLMCredsAndSetTeammateMode`)
   remain package-cli-internal. Source LOC: 503 → 365.

The M1-blocker lesson (SPLIT-001 plan.md §F Provenance) applies: cluster characterization
without verbatim file:line verification is unsound. The three type-dependency hazards above were
re-derived this run via grep (NOT assumed from SPLIT-001). The run phase MUST re-verify all three
before the move. iter-1 cited the lesson but did not apply it to CheckStatus (D1) or
SettingsLocal (D2); iter-2 applies it to both.

### §C.3 Secondary risks

- **`SettingsLocal` cycle hazard** — D2 RESOLVED at plan-phase. `settings.go:28` `mutateSettingsLocal(path string, mutate func(*SettingsLocal))` operates on the `SettingsLocal` struct defined at
  `glm.go:97`. **Resolution: option (a) STAYS reclassification** — settings.go does NOT move to
  uikit; the helpers stay package-cli-internal; the leaf is narrowed to render + banner +
  schema-bridge helpers (365 LOC). The "run phase decides" deferral at iter-1 violated the
  SPLIT-001 M1 lesson; iter-2 commits to option (a) with verified file:line evidence
  (design.md §C.3).
- **Test file migration** — `render_test.go` is a white-box test (`package cli`) that accesses
  unexported render helpers. Moving render.go to uikit means render_test.go moves too AND becomes
  `package uikit` (or `package uikit_test` — black-box). The misc_coverage_test.go
  PrintWelcomeMessage tests similarly MOVE per design.md §D.10 resolution. The 5
  CheckStatus-bearing test files (coverage_test.go, coverage_fixes_test.go,
  coverage_improvement_test.go, mcp_doctor_coverage_test.go, doctor_golden_test.go) REWRITE
  their CheckStatus references to `uikit.*` forms but otherwise stay in package cli. The run
  phase verifies test count before == after (AC-CUK-012).
- **Cobra registration unaffected** — uikit is NOT a cobra command package; it exports helpers,
  not commands. `root.go` does NOT add a `rootCmd.AddCommand(uikit.Xxx)` line. AC-CUK-003
  (`moai --help` diff empty) verifies this — no new subcommand appears.

## §D. Behavior-Preservation Framing (this is a refactor of WORKING code)

`go build ./...` exits 0 today (verified this run); the CLI works. This SPEC is a **pure
structural refactor** — the observable behavior of every `moai` subcommand MUST be identical
before and after M1. There is NO functional change, NO new feature, NO bug fix. The correct
verification model is **characterization** (SPLIT-001 §D): the full test suite + cross-platform
build + `moai --help` output are the behavior snapshot M1 must preserve.

Because the code already works, the uikit extraction earns its risk ONLY through its foundational
unblock value (it enables future kernel-dependent cluster SPECs). None of the gains are
user-observable. This is the tension the plan-auditor and user must weigh (see spec.md §A VALUE
justification + plan.md §F.9 recommendation).

## §E. Replacement / Continuity Notes

- No public API of `internal/cli` changes for external callers — `cmd/moai/main.go` calls
  `cli.Execute()`, which stays in `package cli`. Verified: `Execute` is defined in `root.go`
  and is the only entry point (SPLIT-001 §E).
- The composition root (`InitDependencies` in `deps.go`) stays in `package cli`. The 4 source
  files have ZERO deps coupling (verified §A.2), so no provider injection is needed.
- The uikit package exports helpers, NOT a cobra command. The existing 7 subpackages
  (`worktree`/`harness`/`preference`/`wizard`/`specid`/`pr`/`agentlint`) are unchanged.

## §F. Recommendation Basis (feeds spec.md / plan.md)

The measured data supports the **M1-only + §F CHECKPOINT** plan:

1. **Definitely do M1 (uikit extraction)** — the foundational unblock. SPLIT-001 §F.9 already
   committed to M5 (this SPEC) as the prerequisite for migrate/doctor/update.
2. **THREE cross-file type-dependency resolutions are M1 design-time prerequisites** (§C.2). The
   run phase resolves them BEFORE moving the source files (plan.md §F M1 step 2):
   (a) CheckStatus co-locates into uikit/types.go; (b) profileSetupText b-ii split keeping BOTH
   maps in package cli; (c) `SettingsLocal` cycle resolved by settings.go STAYS (option a).
3. **Caller-rewrite blast radius is 12 production files + ≥10 test files** (design.md §D,
   corrected under D2-a). Each site is verified by AC-CUK-001..003 (tests + build + help diff)
   and AC-CUK-008 (residual-reference grep).
4. **Checkpoint decision after M1** — whether to author the next kernel-dependent cluster SPEC
   (migrate is the natural next). The SPEC does NOT commit to authoring them.
5. **Do NOT use uikit as a dumping ground** — the leaf stays cohesive (TUI rendering +
   schema-bridge helpers). Settings helpers STAY in package cli (D2-a). Unrelated helpers belong
   in their own leaf (design.md §E AP-7).

The honest conclusion (spec.md §A): the uikit extraction is a foundational unblock with real
caller-rewrite cost. It earns its risk ONLY because it unblocks future clusters. The §F
CHECKPOINT (REQ-CUK-010) reserves the right to STOP at M1 if the user judges the follow-on
cluster SPECs uneconomic.

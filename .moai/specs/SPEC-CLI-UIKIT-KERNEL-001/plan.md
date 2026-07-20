# Plan — SPEC-CLI-UIKIT-KERNEL-001

Implementation plan for extracting the shared TUI/settings kernel into the
`internal/cli/uikit` leaf package. Single M1 milestone + §F CHECKPOINT.
Mirrors SPEC-CLI-SUBPKG-SPLIT-001 plan.md §A/§B/§C/§D/§E/§G/§H structure; expands
SPLIT-001 §F M5 (uikit kernel) for this SPEC's single-cluster scope.

## §A. Context

- **Location**: `/Users/goos/MoAI/moai-adk-go/internal/cli` (project root; main checkout on
  `main`).
- **Current HEAD (observed this run, 2026-07-07)**: `fdc411807` (drift from the task-prompt's
  stated `10cd484a0` by +1 unrelated chore commit `chore(rules): 남은 것 처리` — does NOT affect
  this SPEC's scope; the SPLIT-001 artifacts remain unchanged at their completed state).
- **Origin**: SPEC-CLI-SUBPKG-SPLIT-001 closed at sync_commit_sha `d0d9b49d7` with only M1
  (agentlint) shipped. Its progress.md §E.4 followup field reads: `"uikit(M5 kernel) / profile /
  migrate / update clusters — separate SPECs post-M1 checkpoint"`. This SPEC is the FIRST of those
  follow-ups — the uikit kernel extraction.
- **Artifacts**: `.moai/specs/SPEC-CLI-UIKIT-KERNEL-001/{spec,plan,acceptance,design,research,progress}.md`
  (Tier L 5-file set + progress.md).
- **Existing pattern to follow**: `internal/cli/{worktree,harness,preference,wizard,specid,pr,agentlint}`
  (subpackage exports a cobra command OR a helper; deps via injected provider). The uikit leaf
  follows the same shape but exports shared helpers (not a cobra command).
- **cycle_type** (for the FUTURE run-phase): `ddd` (existing working code, characterization-
  preservation refactor — behavior must be preserved, no new behavior; maps to ANALYZE-PRESERVE-
  IMPROVE, not RED-GREEN-REFACTOR).

### §A.1 Baseline measurements (observed this run, 2026-07-07 — verification-claim-integrity §2)

| Metric | Value | Command (verbatim) |
|--------|-------|--------------------|
| 4 source files (render/banner/settings/schema_bridge) LOC | **503** | `wc -l internal/cli/render.go internal/cli/banner.go internal/cli/settings.go internal/cli/schema_bridge.go` → `120+146+138+99 = 503` |
| 3 MOVED source files (render/banner/schema_bridge — `settings.go` STAYS per D2) LOC | **365** | `wc -l internal/cli/render.go internal/cli/banner.go internal/cli/schema_bridge.go` → `120+146+99 = 365` |
| Root non-test `.go` files (flat, non-recursive, current) | **95** | `find internal/cli -maxdepth 1 -name '*.go' ! -name '*_test.go' \| wc -l` |
| Root non-test LOC (current) | **25,033** | `find internal/cli -maxdepth 1 … -exec cat {} + \| wc -l` |
| Root test files (current) | **151** | `find internal/cli -maxdepth 1 -name '*_test.go' \| wc -l` |
| Root test LOC (current) | **53,736** | `find internal/cli -maxdepth 1 -name '*_test.go' -exec cat {} + \| wc -l` |
| Build baseline (green?) | **exit 0** | `go build ./...` |
| Caller-rewrite blast radius (production files) | **12** | see design.md §D.9 verbatim file:line map (D1/D3 iter-1 fix: added doctor_cache.go + doctor_harness.go; D2-a applied: launcher.go drops out because settings.go STAYS) |
| Caller-rewrite blast radius (test files) | **≥10** | see design.md §D.10 enumerated test-file list (D4 iter-1 fix: includes the 5 CheckStatus-bearing test files) |

**Baseline drift vs SPLIT-001** (the verification-claim-integrity §2 attribution):
- SPLIT-001 ORIGINAL baseline (2026-07-02 authoring): 93 root non-test files / 25,838 LOC.
- SPLIT-001 drift re-verification (2026-07-07 PRE-audit, PRE-M1-close): 98 / 26,440 (Δ +5 / +602).
- SPLIT-001 M1 close extracted the agentlint cluster (3 non-test files + 2 test files moved to
  `internal/cli/agentlint/`), so the current tree is BELOW the SPLIT-001 drift baseline.
- THIS SPEC measurement (2026-07-07, POST-M1-close): 95 / 25,033.
- Reconciliation: 98 − 3 (agentlint non-test moved out) = 95 ✓. The LOC drop from 26,440 to 25,033
  (−1,407) reflects the agentlint cluster's LOC relocating to depth 2 (now under
  `internal/cli/agentlint/`) plus ordinary churn in the 5 days since authoring.

### §A.2 Cross-references (REUSE, do NOT re-derive)

This SPEC REUSES + EXPANDS the following SPLIT-001 sections; it does NOT re-derive them:
- **SPLIT-001 design.md §A** Target Layout (uikit leaf package entry) — referenced as the
  architectural anchor; this SPEC's design.md §A expands the uikit leaf contract.
- **SPLIT-001 design.md §B** 7-step extraction recipe — referenced verbatim; this SPEC's M1
  follows the same recipe.
- **SPLIT-001 design.md §C** deps provider-injection — N/A for uikit (uikit has NO deps coupling;
  verified this run: `grep -nE 'deps\b|cli\.' internal/cli/settings.go internal/cli/schema_bridge.go`
  returns zero matches → the 4 source files do NOT touch the package-cli global `deps`, so no
  provider injection is needed).
- **SPLIT-001 design.md §D** Import-Cycle Resolution — THIS is the section this SPEC implements
  (the kernel milestone). design.md §A/§B below expand it for the uikit case.
- **SPLIT-001 research.md §C.4.1** 14-file kernel survey — referenced as the survey source; this
  SPEC's research.md §B classifies each of the 14 files into MOVED / STAYS / GATED. (D6 iter-1
  fix: the iter-1 plan.md said "§C.1" — SPLIT-001 research.md §C.1 is "Import-cycle hazard" at
  L94, NOT the 14-file survey. The 14-file survey is at §C.4.1 L173.)
- **SPLIT-001 plan.md §F M5** — referenced as the milestone this SPEC implements.
- **SPLIT-001 acceptance.md §A-§C** AC-CSS-001..012 — mirrored as AC-CUK-001..013.

## §B. Known Issues (auto-injected, filtered to relevant)

- **B1 Cross-platform build tags** [RELEVANT]: the 4 source files have NO build-tag siblings
  (verified: none of render.go/banner.go/settings.go/schema_bridge.go carries a `//go:build`
  constraint). But the cross-platform build gate still applies: `GOOS=windows GOARCH=amd64
  go build ./...` MUST pass after M1.
- **B3 Subagent boundary (C-HRA-008)** [RELEVANT]: `internal/cli/uikit` code must not call
  `AskUserQuestion`/`mcp__askuser`. The move must not introduce any; grep gate after M1
  (AC-CUK-009).
- **B4 Frontmatter canonical schema** [APPLIED]: 12 canonical fields + `tier: L` + `era: V3R6`.
  Verified in progress.md §E.1.
- **B5 CI 3-tier** [RELEVANT]: spec-lint + golangci-lint + Test (per OS) each fail independently.
- **B6 spec-lint `MissingExclusions`** [APPLIED]: spec.md §E uses `### Out of Scope — <topic>`
  H3 sub-headings with bullets (not a bare H2).
- **B8/B10 Working-tree hygiene / PRESERVE** [RELEVANT]: do not touch runtime-managed files
  (`.moai/state/`, `.moai/cache/`, `.moai/logs/`) or unrelated SPEC dirs (especially the completed
  SPLIT-001 artifacts — modify NOTHING under `.moai/specs/SPEC-CLI-SUBPKG-SPLIT-001/`);
  `git add` specific paths only.
- **B9 Commit + push (Hybrid Trunk)** [RELEVANT]: Tier L → Route B (PR route per `--pr`/Tier L)
  is the default for Tier L, but a maintainability refactor with per-milestone commits MAY use
  main-direct per user choice at Implementation Kickoff Approval. Conventional Commits;
  `--no-verify` prohibited.
- **B11 AskUserQuestion prohibition** [RELEVANT]: subagent returns blocker reports, never prompts.
- **Cross-file type-dependency hazards (SPEC-specific)** [CRITICAL]: THREE cross-file type
  dependencies MUST be resolved BEFORE the source files move (REQ-CUK-007):
  (a) `render.go:60` `renderStatusLine(status CheckStatus, ...)` — `CheckStatus` is defined in
  `doctor.go:25-26`. Resolution: the `CheckStatus` type (and its 3 consts `CheckOK`/`CheckWarn`/
  `CheckFail`) MUST move to uikit (it is a generic `"ok" / "warn" / "fail"` enum); every
  consumer rewrites to `uikit.CheckStatus` / `uikit.CheckOK` / etc. **D1 iter-1 BLOCKING fix**:
  the consumers are NOT just `doctor.go` — verified this run via grep, 7 files carry 43
  CheckStatus references (doctor.go, doctor_cache.go, doctor_harness.go + 4 test files). All 7
  files rewrite to `uikit.*`-qualified forms. See design.md §C.5 / §D.8.
  (b) `schema_bridge.go:24-77` declares TWO maps referencing the `profileSetupText` type (defined
  `profile_setup_translations.go:10`) — `schemaFieldBridge` (L24-58, 19 entries) AND
  `schemaSegmentBridge` (L60-77, 16 entries). **D5 iter-1 MINOR fix**: iter-1 said "L24-65 one
  map"; the second map `schemaSegmentBridge` ALSO references `profileSetupText` and the b-ii
  split MUST cover BOTH maps. Resolution: the b-ii split is the design-time choice — BOTH
  `schemaFieldBridge` AND `schemaSegmentBridge` stay in package cli; only the leaf-able helpers
  (`SchemaKeyToTUIField`, `FieldDefTUILabel`) move to uikit. See design.md §C.4.
  (c) `settings.go:28/48/110` references the `SettingsLocal` type (defined `glm.go:97`).
  **D2 iter-1 BLOCKING fix**: `settings.go` STAYS in package cli — it does NOT move to uikit.
  Verified this run: `grep -nE 'type SettingsLocal' internal/ -r --include='*.go'` returns
  `internal/cli/glm.go:97`. glm.go is a STAYS file (kernel-USING). Moving `settings.go` into
  uikit would create `uikit → package cli` import cycle (defeats REQ-CUK-001 leaf contract).
  Resolution: option (a) STAYS reclassification — `settings.go` remains in package cli; its
  helpers (`mutateSettingsLocal`, `writeFileAtomic`, `stripGLMCredsAndSetTeammateMode`) are NOT
  exported from uikit. Source LOC drops from 503 to 365; Tier L stays justified on the corrected
  12+ production + ≥10 test blast radius. See design.md §C.3.

## §C. Pre-flight (before M1)

```bash
git branch --show-current && git rev-parse HEAD           # main + <current-sha>
go build ./...                                            # expect exit 0 (baseline confirmed)
GOOS=windows GOARCH=amd64 go build ./...                  # expect exit 0
go test ./internal/cli/... > /tmp/cli-before.txt 2>&1; tail -3 /tmp/cli-before.txt
golangci-lint run --timeout=2m 2>&1 | tail -5             # capture lint baseline (NEW vs pre-existing)
go run ./cmd/moai --help > /tmp/help-before.txt           # behavior snapshot: subcommand list
# Verify caller blast radius against the CURRENT tree (do NOT assume design.md §D is stable):
grep -rnE 'renderCard|renderKeyValue|renderStatusLine|renderSuccessCard|renderInfoCard|renderSummaryLine|RenderError|PrintBanner|printWelcomeMessage|PrintWelcomeMessage|mutateSettingsLocal|writeFileAtomic|schemaKeyToTUIField|kvPair|CheckStatus|CheckOK|CheckWarn|CheckFail|profileSetupText' internal/cli/*.go | grep -v _test.go
```

## §D. Constraints (DO NOT VIOLATE)

- PRESERVE: SPEC-CLI-SUBPKG-SPLIT-001 artifacts UNCHANGED (completed SPEC; cross-reference only,
  modify nothing under `.moai/specs/SPEC-CLI-SUBPKG-SPLIT-001/`).
- PRESERVE: existing 7 subpackages (`worktree`/`harness`/`preference`/`wizard`/`specid`/`pr`/
  `agentlint`) unchanged; `cli.Execute()` + `deps.go` Composition Root stay in `package cli`.
- PRESERVE: runtime-managed files (`.moai/state/`, `.moai/cache/`, `.moai/logs/`).
- No functional change (REQ-CUK-011); no test deletion/skip (REQ-CUK-012); no new behavior tests.
- Single M1 milestone (REQ-CUK-009); atomic behavior-preserving commit.
- Cross-platform build green (B1); grep-verify no AskUserQuestion (B3, AC-CUK-009).
- No out-of-scope cluster work: profile/migrate/doctor/update are NOT touched in this SPEC
  (only the `CheckStatus` type reference in `doctor.go` is rewritten, because the type moves with
  `render.go` — this is part of the kernel-axis caller rewrite, NOT a doctor-cluster extraction).

## §E. Self-Verification (M1) — see acceptance.md for full matrix

M1 reports: AC-CUK PASS/FAIL matrix, `go build` matrix result, `go test ./...` result,
`moai --help` diff (expect empty), lint status (NEW vs baseline), commit SHA, type-co-location
evidence (REQ-CUK-007), caller-rewrite completeness evidence (REQ-CUK-006).

## §F. Milestones

> Single M1 milestone + §F CHECKPOINT. The uikit extraction is ONE atomic milestone — the 4
> source files move together, the type dependencies resolve together, the caller rewrites land
> together, the test rewrites land together, all in one commit (REQ-CUK-009).

### M1 — Extract `uikit` kernel → `internal/cli/uikit` (MED-HIGH risk; Tier L)

**Scope (365 LOC source + caller rewrite; `settings.go` STAYS per D2 option a)**:

1. **Create the leaf package directory** `internal/cli/uikit/` with `package uikit` decl.
2. **Resolve type dependencies FIRST (REQ-CUK-007)** — THREE resolutions, all design-time:
   - (a) Move `CheckStatus` type + its 3 consts (`CheckOK`/`CheckWarn`/`CheckFail`) from
     `doctor.go:25-34` into `internal/cli/uikit/types.go` (new file). The type is a generic
     `"ok" / "warn" / "fail"` status enum — uikit is its natural home. **D1 iter-1 BLOCKING fix**:
     every consumer of the type rewrites — doctor.go (4 sites), doctor_cache.go (5 sites),
     doctor_harness.go (4 sites), AND 4 test files (28 sites: mcp_doctor_coverage_test.go,
     coverage_test.go, coverage_fixes_test.go, coverage_improvement_test.go) + doctor_golden_test.go
     (9 sites). See design.md §D.8.
   - (b) `profileSetupText` coupling in `schema_bridge.go`: TWO maps reference the type —
     `schemaFieldBridge` (L24-58) AND `schemaSegmentBridge` (L60-77) (D5 iter-1 MINOR fix: iter-1
     said one map; both reference `profileSetupText`). Resolution is **b-ii split** (design-time
     decision): BOTH maps stay in package cli (e.g. new `package cli` file
     `schema_bridge_profile.go`); only the leaf-able helpers `SchemaKeyToTUIField` +
     `FieldDefTUILabel` move to uikit. The b-i alternative (co-locate profileSetupText into
     uikit) is rejected — it pulls 609 LOC of i18n data into the kernel leaf, violating AP-7
     (uikit-as-dumping-ground).
   - (c) `SettingsLocal` coupling in `settings.go`: **D2 iter-1 BLOCKING fix — option (a) STAYS
     reclassification**. `settings.go` does NOT move to uikit; it remains in package cli. The
     helpers (`mutateSettingsLocal`, `writeFileAtomic`, `stripGLMCredsAndSetTeammateMode`) stay
     package-cli-internal (unexported). Verified this run: `SettingsLocal` is defined at
     `glm.go:97` (a STAYS file); moving settings.go to uikit would force uikit to import package
     cli for the type, creating a cycle that defeats REQ-CUK-001. Source LOC: 503 → 365.
3. **Move the 3 source files** into `internal/cli/uikit/` (git mv preserves history):
   `render.go`, `banner.go`, `schema_bridge.go` (only the helper-portion under b-ii split — the
   TWO profileSetupText-referencing maps stay in package cli per step 2-b). **`settings.go`
   STAYS in package cli** (D2 option a).
4. **Re-scope symbols**: every helper consumed by `package cli` becomes EXPORTED in uikit:
   - From `render.go`: `RenderError` (already exported), `cardStyle`→`CardStyle` (if external),
     `renderCard`→`RenderCard`, `renderKeyValue`→`RenderKeyValue`,
     `renderKeyValueLines`→`RenderKeyValueLines`, `renderStatusLine`→`RenderStatusLine`,
     `renderSuccessCard`→`RenderSuccessCard`, `renderInfoCard`→`RenderInfoCard`,
     `renderSummaryLine`→`RenderSummaryLine`; supporting type `kvPair`→`KVPair`.
   - From `banner.go`: `PrintBanner`, `PrintWelcomeMessage` (already exported); unexported
     helpers `resolveTheme`/`goVersion`/`claudeVersion`/`gitVersionOverride`/
     `ghVersionOverride`/`goosArch` stay unexported UNLESS a caller outside banner.go uses
     them (run-phase grep decides).
   - From `schema_bridge.go`: `schemaKeyToTUIField`→`SchemaKeyToTUIField`,
     `fieldDefTUILabel`→`FieldDefTUILabel` (if externally consumed). The TWO maps
     (`schemaFieldBridge`, `schemaSegmentBridge`) STAY in package cli per step 2-b-ii.
   - `settings.go` STAYS in package cli (D2-a); its helpers are NOT re-scoped, NOT moved.
5. **Rewrite every `package cli` caller** to `uikit.<Helper>(...)` — see design.md §D for the
   verbatim file:line blast-radius map. Each call site becomes `uikit.RenderCard(...)`,
   `uikit.PrintBanner(...)`, etc. Every `CheckStatus` / `CheckOK` / `CheckWarn` / `CheckFail`
   reference across the 7 CheckStatus-bearing files rewrites to `uikit.CheckStatus` etc. (the
   `settings.go` helpers' call sites — e.g. `launcher.go:206 mutateSettingsLocal(...)`,
   `stripGLMCredsAndSetTeammateMode` — do NOT rewrite under D2-a; both helpers stay in package cli).
6. **Move the test files** that exercise the moved helpers (e.g. `render_test.go`,
   the `PrintWelcomeMessage` tests in `misc_coverage_test.go` per design.md §D.10 resolution) —
   white-box tests declared `package cli` become `package uikit` (or `package uikit_test`);
   references to symbols still in `package cli` resolve via the import (the test files become
   black-box tests of uikit). The CheckStatus-bearing test files (coverage_test.go,
   coverage_fixes_test.go, coverage_improvement_test.go, mcp_doctor_coverage_test.go,
   doctor_golden_test.go) REWRITE their CheckStatus references to `uikit.*` forms but otherwise
   stay in package cli (they are not white-box tests of moved helpers).
7. **Verify**: `go build ./...` exit 0 → `GOOS=windows GOARCH=amd64 go build ./...` exit 0 →
   `go test ./...` zero NEW failures (pre-existing baseline failures documented) →
   `go vet ./...` → `golangci-lint run`.
8. **Commit** as one atomic behavior-preserving commit (Conventional Commits:
   `refactor(SPEC-CLI-UIKIT-KERNEL-001): M1 extract uikit kernel to internal/cli/uikit`).

**Gate**: AC-CUK-001..013 (acceptance.md). The binding gate is `go test ./...` (zero NEW
failures) + `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` + `moai --help` diff
empty + caller-rewrite grep complete (no residual pre-move symbol references in `package cli`).

### CHECKPOINT (after M1) — re-evaluate before any kernel-dependent cluster SPEC (REQ-CUK-010)

- The uikit leaf is now available. The question is whether the future kernel-dependent cluster
  SPECs (migrate / doctor / update) are worth their REMAINING coupling-resolution cost (i.e.
  the cost BEYOND axis-i, which uikit just removed).
  - **migrate**: axis-(i) kernel resolved by this SPEC. Remaining: axis-(ii) shared-helper
    (`update_archive.go` co-extraction) + axis-(iii) reverse-dep (`update.go:1922` constructor
    refactor). Both characterized in SPLIT-001 research.md §C.4.3.
  - **doctor**: heavy kernel use (now resolvable via `uikit` import). 9 files / 2,357 LOC +
    11 test files / ~2,706 test LOC. No deps coupling.
  - **update**: highest value + highest risk. 9 files / 5,181 LOC + 21 test files / 9,283 test
    LOC + deps provider injection. Gates on resolving the `update.go:1922` constructor refactor
    for migrate axis-(iii).
- **Decision**: STOP (ship M1 only, record decision in progress.md) OR continue to author the
  next kernel-dependent cluster SPEC (migrate is the natural next — its axis-(i) is now resolved).
- **Honest value/risk signal**: if the uikit M1 caller-rewrite cost (12+ production files + ≥10
  test files rewritten) felt uneconomic, that is a signal that the future cluster SPECs (which
  touch larger LOC per cluster) may also be uneconomic. The SPEC does NOT commit to authoring
  them — the checkpoint decides.

### §F.9 Recommendation (honest value/risk call)

**Recommend SHIP M1 + checkpoint-decide** — NOT a forced ladder of follow-up cluster SPECs.
Concretely:

1. **Definitely do M1 (uikit extraction)** — it is the foundational unblock. Without it, NO
   kernel-dependent cluster can move. SPLIT-001 §F.9 already committed to M5 (this SPEC) as the
   prerequisite for migrate/doctor/update; the checkpoint decision in SPLIT-001 chose to STOP at
   M1 (agentlint) AND pursue uikit as a separate SPEC (this one). M1 here is that commitment.
2. **Checkpoint decision after M1.** If the caller-rewrite cost felt smooth AND the user accepts
   the remaining per-cluster coupling-resolution cost, author the next SPEC (migrate is the
   natural next — its axis-(i) is now resolved); otherwise STOP at M1.
3. **Treat the kernel-dependent cluster SPECs (migrate/doctor/update) as CONDITIONAL**, each
   behind its own documented coupling-resolution design. The uikit leaf is necessary but NOT
   sufficient for them — they each carry additional axis-(ii)/(iii) work this SPEC does not do.
4. **Do NOT use uikit as a dumping ground** for unrelated helpers. The leaf stays cohesive
   (TUI rendering + settings mutation + schema-bridge helpers). Helpers that don't fit this
   cohesion belong in their own leaf (e.g. `archiveutil` for the migrate/update shared helpers).

Rationale: the uikit extraction is the unblock the SPLIT-001 plan envisioned. It carries real
caller-rewrite cost (12+ production files + ≥10 test files) and three cross-file type-dependency
resolutions (CheckStatus co-location; profileSetupText b-ii split covering BOTH maps;
SettingsLocal coupling resolved by settings.go STAYS per D2 option a). It earns its risk ONLY
because it unblocks future clusters — and the §F CHECKPOINT (REQ-CUK-010) reserves the right to
STOP at M1 if the user judges the follow-on cluster SPECs uneconomic.

## §G. Anti-Patterns

See design.md §E (AP-1 big-bang, AP-2 export-and-import-back cycle, AP-3 orphaned tests,
AP-5 refactor-time behavior tests) and SPLIT-001 design.md §G (cross-reference; the AP catalogue
carries forward).

## §H. Cross-References

- `.moai/specs/SPEC-CLI-SUBPKG-SPLIT-001/design.md` §C/§D — the deps provider-injection pattern
  (N/A for uikit — verified zero deps coupling this run) and the import-cycle resolution design
  this SPEC implements.
- `.moai/specs/SPEC-CLI-SUBPKG-SPLIT-001/research.md` §C.4.1 — the 14-file kernel survey this
  SPEC's research.md §B classifies. (D6 iter-1 fix: the iter-1 plan.md said "§C.1" — SPLIT-001
  research.md §C.1 is "Import-cycle hazard"; the 14-file survey is at §C.4.1.)
- `.moai/specs/SPEC-CLI-SUBPKG-SPLIT-001/plan.md` §F M5 — the milestone this SPEC implements.
- `.moai/specs/SPEC-CLI-SUBPKG-SPLIT-001/acceptance.md` §A-§C — the AC-CSS-001..012 structure
  this SPEC's acceptance.md mirrors as AC-CUK-001..013.
- research.md — 14-file classification + two dominant risks (blast radius, leaf-contract
  correctness).
- design.md — uikit leaf contract, helper inventory, caller-rewrite blast-radius map.
- `internal/cli/CLAUDE.md` — cobra registration, cross-platform, subagent boundary conventions.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Tier L Section A-E template
  (for the future run-phase delegation).

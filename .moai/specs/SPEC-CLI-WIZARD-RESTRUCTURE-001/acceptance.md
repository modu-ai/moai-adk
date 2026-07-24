---
id: SPEC-CLI-WIZARD-RESTRUCTURE-001
title: "Acceptance criteria — moai init wizard restructure (방안 A)"
version: "0.2.1"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
tier: L
---

# Acceptance Criteria — SPEC-CLI-WIZARD-RESTRUCTURE-001

## §A — Given-When-Then scenarios

### AC-WIZ-001 — Three topic pages, no gate (REQ-WIZ-001/002)
- **Given** a fresh `moai init` interactive run with no mode flags,
- **When** the wizard assembles its form groups,
- **Then** exactly three topic pages emerge (`Basic`, `Model & Report`,
  `Quality & Workflow`), the `advanced_bridge` question is absent from the
  question set, and no confirmation step gates the Quality & Workflow page.
- **Verify:** a wizard test asserts `buildFormGroups` yields the three
  page-groups (plus the `claude_design_enabled` conditional sub-group) for the
  default result; `QuestionByID(..., "advanced_bridge")` returns nil.

### AC-WIZ-002 — Page membership & order (REQ-WIZ-003/004/005)
- **Given** the restructured question set,
- **When** the pages are enumerated,
- **Then** Page 1 = [`conversation_language`, `user_name`, `project_name`],
  Page 2 = [`model_policy`, `report_format`], Page 3 =
  [`lsp_enabled`, `enforce_quality`, `project_mode`, `design_enabled`,
  `claude_design_enabled`] in that order.
- **Verify:** table-driven test asserts each question's `Group` label and
  relative position.

### AC-WIZ-003 — Nested design condition preserved (REQ-WIZ-006)
- **Given** the Quality & Workflow page,
- **When** `design_enabled` is answered false,
- **Then** `claude_design_enabled` is hidden; when `design_enabled` is true it
  is shown.
- **Verify:** `TotalVisibleQuestions` / hide-func test with
  `DesignEnabled=false` then `true`; the `claude_design_enabled` group is its
  own conditional group navigated after `design_enabled`. Assert its
  `Condition` no longer references `StandardMode` (it collapses from
  `r.StandardMode && r.DesignEnabled` to `r.DesignEnabled`).

### AC-WIZ-004 — Locale live-render on Page 1 (REQ-WIZ-007)
- **Given** Page 1 with the default locale,
- **When** the user selects a different `conversation_language`,
- **Then** the `user_name` and `project_name` titles/descriptions render in the
  newly selected language.
- **Verify:** exercise `saveAnswer("conversation_language", ...)` updates the
  live locale pointer and `GetLocalizedQuestion` returns the new-locale strings
  for the sibling questions.

### AC-WIZ-005 — model_policy default = Medium + recommended marker (REQ-WIZ-008/011)
- **Given** the `model_policy` question,
- **When** its default and option labels are read,
- **Then** the default value is `"medium"`, the Medium option carries the
  `(Recommended)` marker, the Max option does NOT, and `ModelPolicyHigh` remains
  a selectable option value.
- **Verify:** assert `QuestionByID(...,"model_policy").Default == "medium"`;
  assert exactly one option label contains the recommended marker and it is the
  Medium (value `"medium"`) option; assert `"high"` is still present as an
  option value.

### AC-WIZ-006 — No "high" default seed remains (REQ-WIZ-009/011)
- **Given** the repo after the change,
- **When** the model-policy default seeds are audited,
- **Then** `DefaultModelPolicy == ModelPolicyMedium`, no doc-comment states the
  default is `"high"`, and `ModelPolicyHigh` survives only as an enum
  constant/option (not as a default seed).
- **Verify (grep, distinguishing default-seed from option):**
  ```bash
  grep -n 'DefaultModelPolicy' internal/template/model_policy.go        # = ModelPolicyMedium
  # N4 non-vacuity fix: the OLD pattern 'default: "high"|default is "high"|default.*"high"'
  #   was CASE-SENSITIVE and returned exactly 1 match (context.go:64) and ZERO for
  #   profile.go — whose real text is `DefaultModelPolicy = "high"` (capital D, no
  #   lowercase "default"). It therefore read 0 for profile.go both before AND after
  #   the C11 edit → vacuous, leaving C11 unverified. Grep case-insensitively AND
  #   anchor on the real token:
  grep -rin 'DefaultModelPolicy *= *"high"\|default.*"high"' internal/template/context.go internal/config/profile.go
  #   → 0 after the change. NON-VACUITY PROOF: against the CURRENT pre-change tree this
  #     returns exactly 2 matches (context.go:64 + profile.go:59), so it must fall to 0.
  grep -n 'ModelPolicyHigh' internal/template/model_policy.go           # const still defined
  ```

### AC-WIZ-007 — lsp_enabled default = true, 4-locale (REQ-WIZ-010)
- **Given** the `lsp_enabled` question,
- **When** its default and localized titles are read,
- **Then** the default is `"true"` and the title/description reflect
  enabled-by-default in en/ko/ja/zh.
- **Verify:** assert `QuestionByID(...,"lsp_enabled").Default == "true"`; then
  per-locale grep confirms the disabled-by-default wording is gone in all four
  locales. **The greps are bare (unscoped) on purpose:** the disabled-by-default
  wording is carried by exactly three questions — `advanced_bridge` (removed by
  C1/C14), `lsp_enabled` (flipped by C6/C13) and `coverage_exemptions_enabled`
  (removed by C5/C16) — so after M3 the correct post-change count is **0**, with
  no per-question scoping needed.
  ```bash
  # ko/ja/zh titles. NOTE the character class [：:] — zh writes a FULLWIDTH
  #   colon (「默认：否」) while ko/ja write a halfwidth one. A halfwidth-only
  #   pattern ('默认: 否') matches 6 of the 9 entries and is STRUCTURALLY BLIND
  #   to all three zh entries — the same vacuity class as review-2 N4.
  grep -cE '기본값[：:] *아니오|デフォルト[：:] *いいえ|默认[：:] *否' internal/cli/wizard/translations.go   # 0
  # en titles live inline in questions.go, not translations.go (per C6):
  grep -c 'default: No' internal/cli/wizard/questions.go                                                  # 0
  ```
  **Non-vacuity proof (measured):** against the CURRENT pre-change tree the
  first grep returns exactly **9** (ko L43/125/133, ja L155/237/245, zh
  L267/349/357 — 3 per locale) and the second returns exactly **3**
  (`questions.go` L134 `advanced_bridge`, L417 `lsp_enabled`, L439
  `coverage_exemptions_enabled`). Both must fall to 0.

### AC-WIZ-008 — harness_profile removed, fixed at default (REQ-WIZ-012)
- **Given** the init wizard,
- **When** the question set is enumerated,
- **Then** `harness_profile` is absent AND the effective default harness
  profile resolves to `default` for a completed init.
- **Verify:** `QuestionByID(...,"harness_profile") == nil`; after a
  deployer-path init, `.moai/config/sections/harness.yaml` contains
  `default_profile: "default"` (satisfied by the shipped template default —
  no write path exists, per spec.md §C).

### AC-WIZ-009 — coverage_exemptions removed, fixed at false (REQ-WIZ-013)
- **Given** the init wizard,
- **When** the question set is enumerated,
- **Then** `coverage_exemptions_enabled` is absent AND the effective
  coverage-exemptions setting resolves to `false`.
- **Verify:** `QuestionByID(...,"coverage_exemptions_enabled") == nil`; after a
  deployer-path init, `.moai/config/sections/quality.yaml` contains
  `coverage_exemptions:` with `enabled: false` (satisfied by the shipped
  template default — no write path exists, per spec.md §C).

### AC-WIZ-010 — Page-3 answers persist to disk on the deployer path (REQ-WIZ-015; MUST-FIX, blocking)
- **Given** an init run through the **template-deployer path** (the path a real
  `moai init` takes — deployer non-nil), with no `--standard` / `--advanced`
  flag and no Page-3 override flag,
- **When** init completes,
- **Then** each Page-3 answer is readable from its on-disk configuration file.
- **Verify (binary, on-disk only — N3: the `opts` alternative is DELETED):**
  a Go integration test (plan.md **C39**) constructs the initializer with a
  **real (non-nil) deployer**, runs `Initialize`, and asserts on the resulting
  `.moai/config/sections/` files. **Two scenarios are required** — the design
  nesting (REQ-WIZ-006) makes it impossible for both design keys to diverge
  from their shipped defaults in a single run, so one scenario cannot exercise
  both write paths.

  **Scenario A — design ON** (`project_mode=team`, `lsp_enabled=true`,
  `enforce_quality=false`, `design_enabled=true`, `claude_design_enabled=false`):

  | # | Answer | File | Asserted key/value | Shipped default | Discriminating? |
  |---|---|---|---|---|---|
  | 1 | project_mode=team | `project.yaml` | `project.mode: team` | `personal` (`project.yaml.tmpl:15`) | **yes** |
  | 2 | lsp_enabled=true | `lsp.yaml` | `lsp.enabled: true` | `false` (`lsp.yaml.tmpl:45`) | **yes** |
  | 3 | enforce_quality=false | `quality.yaml` | `constitution.enforce_quality: false` | renders `true` (`context.go:100` `EnforceQuality: true`) | **yes** |
  | 4 | design_enabled=true | `design.yaml` | `design.enabled: true` | `true` (`design.yaml:8`) | **NO — default-coincident** |
  | 5 | claude_design_enabled=false | `design.yaml` | `design.claude_design.enabled: false` | `true` (`design.yaml:44`) | **yes** |

  **Scenario B — design OFF** (`design_enabled=false`; `claude_design_enabled`
  is hidden by the REQ-WIZ-006 nesting and is therefore NOT asserted here):

  | # | Answer | File | Asserted key/value | Shipped default | Discriminating? |
  |---|---|---|---|---|---|
  | 6 | design_enabled=false | `design.yaml` | `design.enabled: false` | `true` (`design.yaml:8`) | **yes** |

  Row 6 is the assertion that genuinely exercises the `design.enabled` write
  path. Without it, a `writeDesignYAML` that silently dropped `design.enabled`
  while handling `claude_design.enabled` correctly would still pass Scenario A
  — row 4 alone cannot catch it.

  Asserting that the answers reach `project.InitOptions` is **NOT sufficient**
  and does NOT satisfy this AC — `opts` is presence one level deeper, not
  landing.

  **Non-vacuity proof (corrected at v0.2.1, review-3 D1 — the v0.2.0 claim
  "MUST fail on all five rows" was factually false):** against the CURRENT tree
  Scenario A MUST fail on **four of its five rows** — rows 1, 2, 3 and 5. Row 4
  is **default-coincident**: the deployed `design.yaml` already ships
  `enabled: true`, so that row passes before AND after the change and is
  retained only as a template-default regression guard, contributing no
  discriminating power. Scenario B MUST fail on its single row (row 6).
  **Combined, every one of the five persistence targets has at least one row
  that fails pre-change**, so no write target is left unverified. A run in
  which Scenario A fails on fewer than four rows, or Scenario B passes, means
  the test is not reaching the write path it claims to test.
- **Additional structural assertion:** `grep -rn 'if !opts.StandardMode'
  internal/core/project/` returns 0, and `grep -rn 'WritePhase1Configs'
  --include='*.go' internal/core/project/initializer.go` shows the call OUTSIDE
  `generateConfigsFallback`.

### AC-WIZ-010a — Persistence is non-destructive (REQ-WIZ-021; MUST-FIX, blocking)
- **Given** the same deployer-path init run as AC-WIZ-010,
- **When** the Page-3 answers have been persisted,
- **Then** every other key, comment, and section of each patched file survives
  intact — the write patches one key, it does not replace the document.
- **Verify (sentinel keys drawn from deep inside each deployed file, so a
  wholesale overwrite cannot pass):**
  ```bash
  # after a deployer-path init into a temp project root:
  grep -q 'delegate_to_astgrep:'  <root>/.moai/config/sections/lsp.yaml      # deep lsp.yaml key survives
  grep -q 'circuit_breaker:'      <root>/.moai/config/sections/lsp.yaml
  grep -q 'figma:'                <root>/.moai/config/sections/design.yaml   # deep design.yaml key survives
  grep -q 'brand_context:'        <root>/.moai/config/sections/design.yaml
  grep -q 'default_profile:'      <root>/.moai/config/sections/harness.yaml  # C36: untouched entirely
  ```
  plus a size floor asserted in the Go test: `lsp.yaml` > 8,000 bytes,
  `design.yaml` > 2,000 bytes, `harness.yaml` > 7,000 bytes.
  **Non-vacuity proof:** the wholesale writers produce 2-4 line files
  (< 100 bytes), so every one of these assertions fails if C35/C36 are not
  applied before C32. Record the pre-change deployed sizes captured in plan.md
  §C pre-flight as the baseline.

### AC-WIZ-011 — Translation completeness test passes (REQ-WIZ-014)
- **Given** the restructured question set + trimmed translations,
- **When** `go test ./internal/cli/wizard/...` runs,
- **Then** `TestWizardQuestionTranslationCompleteness` and the model-policy
  translation tests pass, and no orphan ko/ja/zh entries remain for removed
  questions.
- **Verify:** run the test; `grep -n 'advanced_bridge\|harness_profile\|coverage_exemptions_enabled'
  internal/cli/wizard/translations.go` → 0 matches (≥1 against the current tree).

### AC-WIZ-012 — Reconfigure question order preserved (REQ-WIZ-016)
- **Given** `ReconfigureQuestions`,
- **When** the reconfigure set is built,
- **Then** the 7 Git questions are present and appear immediately after
  `report_format`, unchanged from the pre-restructure order.
- **Verify:** assert the Git-question block position relative to
  `report_format` in `ReconfigureQuestions`.

### AC-WIZ-012a — Reconfigure membership unchanged (REQ-WIZ-016)
- **Given** `ReconfigureQuestions` after the restructure,
- **When** the reconfigure question-set membership is enumerated,
- **Then** it contains exactly the pre-restructure member set (Basic + Model +
  spliced Git questions) and does NOT include the Page-3 questions
  (`lsp_enabled`, `enforce_quality`, `project_mode`, `design_enabled`,
  `claude_design_enabled`) — i.e. C8's fold-into-`DefaultQuestions` mechanical
  option MUST NOT leak the Page-3 questions into `moai update --reconfigure`.
- **Verify:** assert `ReconfigureQuestions` membership EXCLUDES the five Page-3
  IDs while retaining the Basic + Model + Git set.

### AC-WIZ-013 — No orphaned capture branches (REQ-WIZ-017)
- **Given** the wizard after removals,
- **When** `saveAnswer` / `saveBoolAnswer` are inspected,
- **Then** no case exists for `advanced_bridge`, `harness_profile`, or
  `coverage_exemptions_enabled`, while `enforce_quality`, `design_enabled`,
  `claude_design_enabled`, `project_mode` cases remain.
- **Verify:** grep the two functions for the removed IDs (0 matches) and the
  retained IDs (present).

### AC-WIZ-014 — Full suite + cross-platform green
- **Given** all changes,
- **When** `go test ./internal/cli/... ./internal/template/... ./internal/config/... ./internal/core/project/...`
  and `GOOS=windows GOARCH=amd64 go build ./...` run,
- **Then** all pass, with wizard-package coverage ≥ 85% and no coverage
  regression in `internal/core/project`.

### AC-WIZ-015 — Advanced-path fully retired (REQ-WIZ-018/019)
- **Given** the repo after the change,
- **When** the advanced-settings plumbing is audited,
- **Then** `internal/cli/wizard/advanced_gate.go` no longer exists, the
  `--standard` / `--advanced` init flags are unregistered, the
  `RunWithDefaultsModes` mode params, the `WizardResult.StandardMode`/
  `AdvancedMode` fields and the `InitOptions.StandardMode` field are gone, and
  no dangling reference remains **anywhere under `internal/`** — build + full
  suite stay green.
- **Verify (retirement grep — 0 residual after retirement):**
  ```bash
  test ! -f internal/cli/wizard/advanced_gate.go                                  # file gone
  grep -rn 'IsAdvancedWizardReady\|AdvancedGate' internal/                        # 0 matches
  # N2 scope fix: the OLD greps were scoped `internal/cli/` and were STRUCTURALLY
  #   BLIND to 15 verified residue lines in internal/core/project/ (initializer.go:56,
  #   :437; initializer_expansion.go:5,23,24,26; initializer_expansion_test.go x9).
  #   The AC reported PASS with all of them intact. Scope is now `internal/`:
  grep -rn 'advancedMode\|standardMode' internal/                                 # 0 matches
  grep -rn 'StandardMode\|AdvancedMode' internal/                                 # 0 matches
  grep -rn 'RunWithDefaultsModes\|Phase2Questions' internal/                      # 0 matches
  # cobra registration idiom is `.Flags().Bool("standard"/"advanced", ...)` — a BARE
  #   flag name (no `--` prefix) via `.Bool(` (NOT `.BoolVar`):
  grep -rn '\.Bool("standard"\|\.Bool("advanced"' internal/                       # 0 after retirement
  # N5 non-regression guard (NOT a discovery grep): .github/ has a VERIFIED 0 baseline
  #   today (`grep -rn '--standard|--advanced' .github/` → exit 1). No CI script
  #   references the flags and nothing under .github/ "breaks". Asserted only so a
  #   future CI script cannot reintroduce them:
  grep -rn '\-\-standard\|\-\-advanced' .github/                                  # 0 — expected 0 BEFORE and AFTER
  ```
  **Non-vacuity:** the four `internal/`-scoped greps return ≥1 against the
  current pre-retirement tree (`StandardMode|AdvancedMode` → 83 lines;
  `.Bool("standard"|.Bool("advanced"` → 4 lines at `init.go:84-85` +
  `init_update_notice_test.go:49-50`). The `.github/` line is the one
  deliberate exception — it is 0 on both sides by design and is labelled as a
  guard, not evidence of work performed.
  **Scope note:** the docs-site `--standard`/`--advanced` reference removal
  (4-locale, 12 files) is NOT verified by AC-WIZ-015 — it is a **sync-phase
  deliverable** (manager-docs, per REQ-WIZ-019 + spec.md §C).

### AC-WIZ-016 — Flag beats wizard for the four overlapping settings (REQ-WIZ-020)
- **Given** an interactive init run where the user supplies an explicit
  `--project-mode` / `--enable-lsp` / `--enforce-quality` / `--enable-design`
  flag AND the wizard answer for that setting differs,
- **When** the wizard result is applied to `opts`,
- **Then** the flag value survives and the wizard answer does not overwrite it;
  and when the flag is NOT supplied, the wizard answer is applied.
- **Verify:** a table-driven `internal/cli` test (plan.md **C41**, new file
  `internal/cli/init_flag_precedence_test.go`) over the 4 settings × {flag
  supplied, flag absent} × {wizard answer agrees, disagrees} asserting the
  resolved `opts` field. The test MUST exercise `--enforce-quality=false` and
  `--enable-lsp=false` explicitly, since `getBoolFlagWithDefault` /
  `getBoolFlag` cannot distinguish "unset" from "explicitly false" by value —
  a correct implementation uses `cmd.Flags().Changed(<name>)` and the test
  fails against a value-only implementation. Also assert consistency with the
  documented `--profile` precedence at `init.go:332-334`.

### AC-WIZ-017 — Depth-aware patch preserves nested same-named keys (REQ-WIZ-022)
- **Given** the C34 depth-aware patch helper applied to `lsp.yaml` and
  `design.yaml` during a deployer-path init,
- **When** `lsp.enabled` / `design.enabled` / `design.claude_design.enabled` are
  patched,
- **Then** every OTHER key named `enabled` in the same file retains both its
  original value and its original indentation.
- **Verify:** a Go unit test on the helper (plan.md **C40**, new file
  `internal/core/project/yaml_patch_test.go`) plus a post-init file assertion:
  ```bash
  # design.yaml deployed indentation multiset for `enabled:` is {2sp x1, 4sp x3, 6sp x1}
  grep -c '^  enabled:'     <root>/.moai/config/sections/design.yaml   # 1
  grep -c '^    enabled:'   <root>/.moai/config/sections/design.yaml   # 3
  grep -c '^      enabled:' <root>/.moai/config/sections/design.yaml   # 1
  # lsp.yaml deployed multiset is {2sp x1, 4sp x1}
  grep -c '^  enabled:'     <root>/.moai/config/sections/lsp.yaml      # 1
  grep -c '^    enabled:'   <root>/.moai/config/sections/lsp.yaml      # 1
  ```
  **Non-vacuity proof:** running the same assertions after a `patchYAMLKey`-based
  implementation yields `5 / 0 / 0` for design.yaml and `2 / 0` for lsp.yaml,
  because `patchYAMLKey` rewrites every same-named key at a hardcoded 2-space
  indent (plan.md §B B-audit-correction 2). The test therefore distinguishes a
  correct depth-aware patch from the naive one the audit prescribed.

## §D — Severity & traceability

### §D.1 — Severity classification

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-WIZ-001, 002 | MUST | Core UX deliverable (3 pages, no gate). |
| AC-WIZ-005, 006 | MUST | Behaviour-affecting default change. |
| AC-WIZ-010 | **MUST (blocking)** | Reachability — without it the whole restructure is cosmetic. Requires the full M4+M5 chain, not C20 alone. |
| AC-WIZ-010a | **MUST (blocking)** | Data-loss guard — the AC-WIZ-010 repair is destructive without it (~22 KB of deployed config at risk). |
| AC-WIZ-008, 009, 011 | MUST | Question removal + locale correctness (test-guarded). |
| AC-WIZ-017 | MUST | Correctness of the patch primitive AC-WIZ-010a depends on. |
| AC-WIZ-003, 004, 007, 012, 012a, 013, 016 | SHOULD | Preservation of existing behaviours (nesting, live-render, LSP locale wording, reconfigure order + membership, no-orphans, flag precedence). |
| AC-WIZ-014 | MUST | Green gate (tests + cross-platform + coverage). |
| AC-WIZ-015 | MUST | Full advanced-path retirement; no vestigial plumbing or dangling callers across `internal/`. |

### §D.2 — Requirement → AC traceability

| REQ | AC(s) |
|-----|-------|
| REQ-WIZ-001 | AC-WIZ-001 |
| REQ-WIZ-002 | AC-WIZ-001 |
| REQ-WIZ-003 / REQ-WIZ-004 / REQ-WIZ-005 | AC-WIZ-002 |
| REQ-WIZ-006 | AC-WIZ-003 |
| REQ-WIZ-007 | AC-WIZ-004 |
| REQ-WIZ-008 | AC-WIZ-005 |
| REQ-WIZ-009 | AC-WIZ-006 |
| REQ-WIZ-010 | AC-WIZ-007 |
| REQ-WIZ-011 | AC-WIZ-005, AC-WIZ-006 |
| REQ-WIZ-012 | AC-WIZ-008 |
| REQ-WIZ-013 | AC-WIZ-009 |
| REQ-WIZ-014 | AC-WIZ-011 |
| REQ-WIZ-015 | AC-WIZ-010 |
| REQ-WIZ-016 | AC-WIZ-012, AC-WIZ-012a |
| REQ-WIZ-017 | AC-WIZ-013 |
| REQ-WIZ-018 | AC-WIZ-015 |
| REQ-WIZ-019 | AC-WIZ-015 |
| REQ-WIZ-020 | AC-WIZ-016 |
| REQ-WIZ-021 | AC-WIZ-010a |
| REQ-WIZ-022 | AC-WIZ-017 |

Counts: **22 REQs → 19 ACs.** Every REQ-WIZ-001…022 is bound. Each REQ ID in
the left column is written out in full (no `003/004/005` shorthand) so a
mechanical `grep -oE 'REQ-WIZ-[0-9]+' acceptance.md | sort -u` over this table
returns all 22 and cannot report a false unbound-REQ. The only AC with
no REQ is AC-WIZ-014 (the global green gate), which is an unbound quality gate
by design.

### §D.3 — Indirect verification notes

- **AC-WIZ-006** distinguishes the model-policy DEFAULT seed from the still-valid
  Max/High OPTION — the grep asserts the const/comment default moved to Medium
  while `ModelPolicyHigh` remains defined. Do not assert "no `high` anywhere":
  `ModelPolicyHigh` and the `"high"` option value legitimately persist.
- **AC-WIZ-008 / AC-WIZ-009** are satisfied by the *shipped template defaults*
  (`harness.default_profile: "default"`, `coverage_exemptions.enabled: false`),
  not by any init write path — both write paths are deliberately absent
  (spec.md §C). The ACs assert the observable end state, and AC-WIZ-010a
  guarantees the template values are not destroyed en route.
- **Every grep-based AC in this document carries an explicit non-vacuity
  proof** — the expected match count against the CURRENT pre-change tree. This
  is a standing requirement introduced after review-1's D1 (a grep that
  returned 0 both before and after) and review-2's N4 (a grep vacuous for one
  of its two target files). A new or edited grep AC without a stated
  pre-change count is incomplete. The single intentional exception is the
  `.github/` guard in AC-WIZ-015, which is 0 on both sides and is labelled as
  such.

## §E — Definition of Done

- [ ] All MUST ACs pass (AC-WIZ-001/002/005/006/008/009/010/010a/011/014/015/017).
- [ ] AC-WIZ-010 (reachability) verified end-to-end **on the deployer path**: a
      Page-3 answer with no mode flag lands in the on-disk yaml. Asserting on
      `opts` does NOT count.
- [ ] AC-WIZ-010a (no-clobber) verified: the deep sentinel keys survive and the
      file-size floors hold for `lsp.yaml` / `design.yaml` / `harness.yaml`.
- [ ] All three persistence gates cleared: `init.go` `if result.StandardMode`
      removed, `initializer_expansion.go` `if !opts.StandardMode` removed,
      `WritePhase1Configs` invoked outside `generateConfigsFallback`.
- [ ] `go test ./internal/cli/... ./internal/template/... ./internal/config/... ./internal/core/project/...`
      green; wizard package coverage ≥ 85%; no `internal/core/project` coverage
      regression.
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` clean.
- [ ] `golangci-lint run` clean for touched packages.
- [ ] 4-locale parity verified for the two flipped defaults (model_policy
      recommended-marker, lsp_enabled enabled-by-default).
- [ ] Advanced-path retirement complete: AC-WIZ-015 green — `advanced_gate.go`
      gone, flags unregistered (verified via the real cobra
      `.Bool("standard"/"advanced")` idiom), 0 dangling references across
      **`internal/`** (not just `internal/cli/`), `.github/` still at its 0
      baseline. (The docs-site 4-locale flag-reference removal is a SYNC-phase
      deliverable — verified at `/moai sync`, NOT this run-phase DoD.)
- [ ] No orphan locale entries / dead capture cases remain.
- [ ] Every test deleted at run-phase appears on the plan.md §G carve-out
      delete-list; no test outside that list was deleted.

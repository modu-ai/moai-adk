---
id: SPEC-CLI-WIZARD-RESTRUCTURE-001
title: "Acceptance criteria — moai init wizard restructure (방안 A)"
version: "0.1.1"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
tier: M
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
  default (both-modes-false) result; `QuestionByID(..., "advanced_bridge")`
  returns nil.

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
  own conditional group navigated after `design_enabled`.

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
  grep -rn 'default: "high"\|default is "high"\|default.*"high"' internal/template/context.go internal/config/profile.go  # 0 matches
  grep -n 'ModelPolicyHigh' internal/template/model_policy.go           # const still defined
  ```

### AC-WIZ-007 — lsp_enabled default = true, 4-locale (REQ-WIZ-010)
- **Given** the `lsp_enabled` question,
- **When** its default and localized titles are read,
- **Then** the default is `"true"` and the title/description reflect
  enabled-by-default in en/ko/ja/zh (no `default: No` / `(기본값: 아니오)` /
  `(デフォルト: いいえ)` / `(默认: 否)` text remaining for this question).
- **Verify:** assert `Default == "true"`; per-locale grep confirms the flipped
  wording in all four locales.

### AC-WIZ-008 — harness_profile removed, fixed at default (REQ-WIZ-012)
- **Given** the init wizard,
- **When** the question set is enumerated,
- **Then** `harness_profile` is absent AND the effective default harness
  profile resolves to `default` for a completed init.
- **Verify:** `QuestionByID(...,"harness_profile") == nil`; init integration
  asserts the persisted/effective harness profile is `default`.

### AC-WIZ-009 — coverage_exemptions removed, fixed at false (REQ-WIZ-013)
- **Given** the init wizard,
- **When** the question set is enumerated,
- **Then** `coverage_exemptions_enabled` is absent AND the effective
  coverage-exemptions setting resolves to `false`.
- **Verify:** `QuestionByID(...,"coverage_exemptions_enabled") == nil`; init
  integration asserts the effective value is `false`.

### AC-WIZ-010 — Page-3 answers are applied (REQ-WIZ-015; MUST-FIX)
- **Given** an init run where the user sets Page-3 answers (e.g. LSP on,
  quality off, project mode team),
- **When** init completes,
- **Then** those answers are persisted to the project config — NOT gated on a
  `StandardMode` flag.
- **Verify:** the `if result.StandardMode` gate in `init.go` is removed; an
  init test with Page-3 answers (and no `--standard`/`--advanced`) asserts the
  answers reach `opts` / the persisted config. **This is the reachability AC —
  presence of the question is insufficient; the answer must land.**

### AC-WIZ-011 — Translation completeness test passes (REQ-WIZ-014)
- **Given** the restructured question set + trimmed translations,
- **When** `go test ./internal/cli/wizard/...` runs,
- **Then** `TestWizardQuestionTranslationCompleteness` and the model-policy
  translation tests pass, and no orphan ko/ja/zh entries remain for removed
  questions.
- **Verify:** run the test; grep translations.go shows zero `advanced_bridge` /
  `harness_profile` / `coverage_exemptions_enabled` entries.

### AC-WIZ-012 — Reconfigure question order preserved (REQ-WIZ-016)
- **Given** `ReconfigureQuestions`,
- **When** the reconfigure set is built,
- **Then** the 7 Git questions are present and appear immediately after
  `report_format`, unchanged from the pre-restructure order.
- **Verify:** assert the Git-question block position relative to
  `report_format` in `ReconfigureQuestions`.

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
- **When** `go test ./internal/cli/... ./internal/template/... ./internal/config/...`
  and `GOOS=windows GOARCH=amd64 go build ./...` run,
- **Then** all pass with wizard-package coverage ≥ 85%.

### AC-WIZ-015 — Advanced-path fully retired (REQ-WIZ-018/019)
- **Given** the repo after the change,
- **When** the advanced-settings plumbing is audited,
- **Then** `internal/cli/wizard/advanced_gate.go` no longer exists, the
  `--standard` / `--advanced` init flags are unregistered, and no dangling
  reference to the removed flags/symbols remains in `internal/cli/`, `.github/`,
  docs, or tests — build + full suite stay green.
- **Verify (retirement grep — 0 residual; intentional CHANGELOG/history mentions
  excepted):**
  ```bash
  test ! -f internal/cli/wizard/advanced_gate.go                                  # file gone
  grep -rn 'IsAdvancedWizardReady\|AdvancedGate' internal/cli/                    # 0 matches
  grep -rn 'advancedMode\|standardMode' internal/cli/                             # 0 matches (params/fields gone)
  grep -rn 'RunWithDefaultsModes\|Phase2Questions' internal/cli/                  # 0 matches (retired constructors)
  grep -rn '"--standard"\|"--advanced"\|advanced.*BoolVar\|standard.*BoolVar' internal/cli/ .github/   # 0 flag registrations/invocations
  ```
  **This is the retirement AC — presence of the pages (AC-WIZ-001) is not
  sufficient; the vestigial advanced plumbing must be gone with no dangling
  callers.**

## §D — Severity & traceability

### §D.1 — Severity classification

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-WIZ-001, 002 | MUST | Core UX deliverable (3 pages, no gate). |
| AC-WIZ-005, 006 | MUST | Behaviour-affecting default change. |
| AC-WIZ-010 | **MUST (blocking)** | Reachability — without it the whole restructure is cosmetic. |
| AC-WIZ-008, 009, 011 | MUST | Question removal + locale correctness (test-guarded). |
| AC-WIZ-003, 004, 007, 012, 013 | SHOULD | Preservation of existing behaviours (nesting, live-render, reconfigure, no-orphans). |
| AC-WIZ-014 | MUST | Green gate (tests + cross-platform + coverage). |
| AC-WIZ-015 | MUST | Full advanced-path retirement (resolved 방안 A option A); no vestigial plumbing or dangling callers. |

### §D.2 — Requirement → AC traceability

| REQ | AC(s) |
|-----|-------|
| REQ-WIZ-001 | AC-WIZ-001 |
| REQ-WIZ-002 | AC-WIZ-001 |
| REQ-WIZ-003/004/005 | AC-WIZ-002 |
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
| REQ-WIZ-016 | AC-WIZ-012 |
| REQ-WIZ-017 | AC-WIZ-013 |
| REQ-WIZ-018 | AC-WIZ-015 |
| REQ-WIZ-019 | AC-WIZ-015 |

### §D.3 — Indirect verification note

AC-WIZ-006 distinguishes the model-policy DEFAULT seed from the still-valid
Max/High OPTION — the grep asserts the const/comment default moved to Medium
while `ModelPolicyHigh` remains defined. Do not assert "no `high` anywhere":
`ModelPolicyHigh` and the `"high"` option value legitimately persist.

## §E — Definition of Done

- [ ] All MUST ACs pass (AC-WIZ-001/002/005/006/008/009/010/011/014/015).
- [ ] AC-WIZ-010 (reachability) verified end-to-end: a Page-3 answer with no
      mode flag reaches the persisted config.
- [ ] `go test ./internal/cli/... ./internal/template/... ./internal/config/...`
      green; wizard package coverage ≥ 85%.
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` clean.
- [ ] `golangci-lint run` clean for touched packages.
- [ ] 4-locale parity verified for the two flipped defaults (model_policy
      recommended-marker, lsp_enabled enabled-by-default).
- [ ] Advanced-path retirement complete (resolved 방안 A option A): AC-WIZ-015
      green — `advanced_gate.go` gone, `--standard`/`--advanced` flags
      unregistered, 0 dangling callers across `internal/cli/` + `.github/` +
      docs + tests (M5 executed).
- [ ] No orphan locale entries / dead capture cases remain.

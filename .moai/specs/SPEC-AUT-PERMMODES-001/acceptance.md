# SPEC-AUT-PERMMODES-001 — Acceptance Criteria

> Tier M AC layer. Each AC is Given-When-Then and binary-testable. Traceability maps AC → REQ → verification surface.

## §D. AC Matrix

| AC | Requirement | Given | When | Then |
|---|---|---|---|---|
| AC-001 | REQ-001 | a fresh `moai init` wizard run | the autonomy question renders | exactly three options are offered, labeled with Claude Code permission-mode vocabulary (Accept edits on / Auto mode / Bypass permissions), with values `semi-auto` / `automatic` / `fully-autonomous` |
| AC-002 | REQ-002 | the same question | its default is inspected | the pre-selected option is "Accept edits on" and its label carries the Recommended signal |
| AC-003 | REQ-002 (amended REQ-006) | the same question | its options are inspected | no option is pre-selected other than Accept edits on; the bypass option is never default or recommended |
| AC-004 | REQ-003 | `TierDefaultMode` | called with `semi-auto` / `automatic` / `fully-autonomous` / an unknown string | it returns `"acceptEdits"` / `"auto"` / `"bypassPermissions"` / `"default"` respectively |
| AC-005 | REQ-004 | a project initialized with the unset or `semi-auto` selection | the applied bundle is inspected | USER-scope settings.json contains exactly `defaultMode: "acceptEdits"` as the only delta; the PROJECT-scope `allow`/`ask`/`deny` arrays and all other deployed files are byte-identical to pre-init state |
| AC-006 | REQ-005 | a user selects Bypass permissions without a sandbox proof | the bundle is applied | the effective tier downgrades to `automatic`, USER `defaultMode` is `"auto"`, and an advisory line is appended to `.moai/logs/autonomy-downgrade.log` |
| AC-007 | REQ-005 | the kill-switch env is active and a sandbox proof IS present | the bypass selection is applied | the downgrade still fires (kill-switch trumps proof) with the same advisory sink |
| AC-008 | REQ-006 | the downgrade regression suite | the affected packages' tests run | `TestApplyAutonomyTierBundle_FullyAutonomousDowngradedWithoutProof`, `TestApplyAutonomyTierBundle_FullyAutonomousWithProofDeploysBypass`, `TestRunInit_FlagFullyAutonomousWithoutProofDowngrades`, `TestAppendDowngradeAdvisory`, and `TestAutonomyTierQuestion_FullyAutonomousNotRecommended` all pass under the new option set |
| AC-009 | REQ-007 | the bundled IAM reference | its Permission Modes section is read | it lists the six current values including `auto` and `dontAsk`, names both kill switches, and no longer asserts "exactly four values" |
| AC-010 | REQ-008 | the wizard translations | each of the four locales (ko/en/ja/zh) is inspected | each carries the three new option labels/descriptions, and the translation completeness test passes |
| AC-011 | REQ-009 | `moai init --help` output | the `--autonomy-tier` flag description is read | it names the three permission-mode choices (accept edits on / auto mode / bypass permissions), states the acceptEdits default, and the closed-set token values (`semi-auto`, `automatic`, `fully-autonomous`) are unchanged |
| AC-012 | REQ-010 | the godoc on the tier-to-mode mapping function and the init-time bundle-apply function (`TierDefaultMode`, `ApplyAutonomyTierBundle`) | each godoc block is read | the mapping godoc states the new mapping (`semi-auto`→`acceptEdits`, `automatic`→`auto`, `fully-autonomous`→`bypassPermissions`), and the bundle godoc states the re-scoped REQ-004 bounded-delta invariant (USER-scope defaultMode only; everything else byte-identical) |

## §D.1 Severity

- **Must-pass**: AC-001..AC-008 (behavioral core + preserved gates)
- **Should-pass**: AC-009..AC-012 (documentation, localization, and doc-comment hygiene)

## §D.2 Traceability

- REQ-001→AC-001, AC-002 · REQ-002→AC-002, AC-003 · REQ-003→AC-004 · REQ-004→AC-005 · REQ-005→AC-006, AC-007 · REQ-006→AC-008 · REQ-007→AC-009 · REQ-008→AC-010 · REQ-009→AC-011 · REQ-010→AC-012. Every REQ carries at least one AC; plan.md §E.4's grep guards (E4) are supplementary drift detection only and are NOT the verification route for any REQ.

## §D.3 Indirect Verification

- AC-005's "everything else byte-identical" is verified by the adapted zero-delta tests comparing full-file bytes, not key-by-key assertions.
- AC-006/AC-007 ride the existing wiring tests (init_autonomy_wiring_test.go) rather than new harnesses.

## §D.4 Closure Gates

- All must-pass ACs green in one observed test run on the affected packages.
- No regression in `go test ./internal/config/... ./internal/core/project/... ./internal/cli/...` affected-scope run.

## §D.5 Quality Gate Criteria

- TRUST 5 Tested: adapted + preserved tests cover every changed mapping branch.
- TRUST 5 Secured: bypass gating intact (AC-006/AC-007 are the security regression net — a failure here is a hard FAIL regardless of other scores).
- Translation completeness: `translations_completeness_test.go` green (TRUST 5 Unified across locales).

## §D.6 Edge Cases

- Unknown/legacy persisted tier values (e.g. hand-edited `workflow.autonomy_tier: "bogus"`) → fail-safe `defaultMode: "default"` (AC-004 last cell).
- Whitespace-only persisted selection → resolves to `semi-auto` → acceptEdits (ResolveEffectiveTier behavior, unchanged).
- Update path (`update_settings_snapshot.go` consumer) applies the same bounded delta — covered by the same bundle tests since both paths share `ApplyAutonomyTierBundle`.

## §D.7 Definition of Done

- [ ] AC-001..AC-012 verified with observed test output recorded in progress.md §E
- [ ] M1 version-floor finding recorded (or a blocker returned)
- [ ] Downgrade regression set intact (AC-008)
- [ ] Bundled reference refreshed and `make build` green
- [ ] No files touched outside the SPEC's declared module scope

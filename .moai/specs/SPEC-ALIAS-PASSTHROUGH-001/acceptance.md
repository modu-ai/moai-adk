# SPEC-ALIAS-PASSTHROUGH-001 — Acceptance Criteria

Every criterion is mechanically verifiable. Commands run from the card worktree root. Two-cell
discipline per verification-completeness §2: release-blocking criteria carry a RED-now cell
(observed on tree `02ad57bbe` or at M1 RED, as marked) and a green-path cell naming the
milestone that flips them.

## §D AC Matrix

| AC | Verifies | Given | When | Then (evidence) |
|---|---|---|---|---|
| AC-ALP-001 | REQ-ALP-001, 006, 009 | Scratch dir with empty `.moai/`, isolated profile base carrying `model: opus[1m]`, stub `claude` binary pinned, `execOrSpawnClaudeFunc` capture seam | The AC-ALP-001 test runs via `go test -run '^TestLaunchModelAliasPassthrough$' ./internal/cli/ -v` (M1 name; adjust only with the same literal discipline) | Captured argv contains `--model` followed by the literal `opus[1m]`. The expectation references no `ModelAliasTable`/`ModelIDOpus55` value. **RED-now (M1)**: on tree `02ad57bbe` the same harness observes `--model claude-opus-5-5[1m]` — verbatim output captured in progress.md §E.2/E8. **Green path: M2** |
| AC-ALP-002 | REQ-ALP-001, 002 | The identity table: {`opus`, `opus[1m]`, `sonnet`, `sonnet[1m]`, `fable`, `fable[1m]`, `haiku`, `opusplan`, `opusplan[1m]`, `claude-opus-5-5`, `claude-opus-5-5[1m]`, `claude-opus-4-8`, `custom-xyz`, `""`} | `go test -run '^TestResolveMainSessionModel_ClaudePassthrough$' ./internal/cli/ -v` | Every row: `resolveMainSessionModel(x, false) == x`; empty stays empty (no flag). Expectations are literals. **RED-now (M1)**: alias rows fail on `02ad57bbe` (observed: the substitution suite `go test -run '^TestExpandModelString$' ./internal/cli/` exits 0 green on this tree, §A.2 of spec.md — the substitution is live). **Green path: M2** |
| AC-ALP-003 | REQ-ALP-002 | The full-id rows of the AC-ALP-002 table (`claude-opus-5-5`, `claude-opus-5-5[1m]`, `claude-opus-4-8`, `custom-xyz`) | Same run as AC-ALP-002 | All pass unchanged — fixed-model users are unaffected (the issue's full-id passthrough guarantee). These rows were green before the change and stay green (regression-guard cell; no RED expected) |
| AC-ALP-004 | REQ-ALP-008 | The Claude-backend path receiving a legacy id and an unknown token | `go test -run '^TestResolveMainSessionModel_ClaudePassthrough$' ./internal/cli/ -v` (legacy/unknown rows) AND `grep -c 'unknown model' internal/cli/launcher.go` | Rows pass through with unchanged values; grep returns 0 (no new warning emission on the model path) |
| AC-ALP-005 | REQ-ALP-003 | The GLM branch of `resolveMainSessionModel` | `go test -run '^TestResolveMainSessionModel_GLMAvoidsCanonicalID$' ./internal/cli/ -v` | All GLM cases green with expectations byte-identical to `02ad57bbe` (`git diff -- internal/cli/launcher_test.go` shows no GLM-case row edit); the canonical-leak guard (:950-956) still asserts no `claude-` prefix under `glmBackend` |
| AC-ALP-006 | REQ-ALP-006 | The passthrough test expectations | `grep -n 'ModelIDOpus55\|ModelIDSonnet55\|ModelAliasTable\|ModelAliasCanonicalID' internal/cli/launcher_test.go` | **Expected post-change count: 0** — the mechanical bar, not an interpretation. All 7 live hits on `02ad57bbe` (launcher_test.go :675, :687-692, :938) sit inside blocks M1 deletes or M2 rewrites, so the file-wide grep MUST return 0 after M2; any hit is a FAIL (a table reference in an expectation would re-couple the proof to table contents). Table-independence: the next upstream alias move cannot flip these tests |
| AC-ALP-007 | REQ-ALP-004 | The deleted substitution function | `grep -rn 'expandModelString' --include='*.go' internal/` | 0 hits — function deleted, no production or test reference remains; `splitModelSuffix` still present (`grep -c 'func splitModelSuffix' internal/cli/launcher.go` returns 1) |
| AC-ALP-008 | REQ-ALP-005 | Display and validation surfaces | `go test -timeout 30m ./internal/web/... ./internal/template/...` AND `git diff --stat -- internal/web internal/cli/profile_setup.go internal/cli/schema_bridge.go internal/template/model_policy.go` | Suite exits 0; diff shows model_policy.go comment-only changes and NO change to `modelOptionList`, `schemaOptionBridge`, `ModelAliasPickerValues`, or the table rows |
| AC-ALP-009 | REQ-ALP-005, 008 | Prefs normalization (legacy-id→alias migration) | `go test -run '^(TestNormalizeModel_Canonical\|TestNormalizeModel_Deprecated\|TestNormalizeModel_EmptyAndUnknown)$' ./internal/cli/ -v` (the three existing normalize tests) | Green with no edit to `profile_setup.go` (covered by AC-ALP-008's diff) — legacy mapping still owned by the prefs layer |
| AC-ALP-010 | REQ-ALP-007 | Stale documentation | `grep -rn 'byte-identical to expandModelString' internal/cli/` AND `grep -n 'used by expandModelString' internal/template/model_policy.go` | Both greps return 0; the `@MX:ANCHOR` reason line no longer lists `expandModelString` while the ANCHOR tag itself remains on `ModelAliasTable` |
| AC-ALP-011 | REQ-ALP-001..005 (umbrella) | Affected packages + lint | `go test -timeout 30m ./internal/cli/... ./internal/template/... ./internal/web/...` AND `golangci-lint run ./internal/cli/... ./internal/template/...` (CI lint version) | Both exit 0 |
| AC-ALP-012 | REQ-ALP-004 (build hygiene) | Cross-platform build | `GOOS=windows GOARCH=amd64 go build ./...` | Exit 0 |

## §D.1 Severity

- **Blocker**: AC-ALP-001 (the defect's regression proof), AC-ALP-002 (passthrough identity),
  AC-ALP-005 (GLM branch unchanged), AC-ALP-011 (quality gates)
- **Major**: AC-ALP-003, AC-ALP-004, AC-ALP-006, AC-ALP-007, AC-ALP-008, AC-ALP-009, AC-ALP-012
- **Minor**: AC-ALP-010 (doc hygiene; still must pass)

## §D.2 Edge Cases

- `opusplan` passes through to itself — identical before and after (table self-map vs verbatim);
  covered in the AC-ALP-002 rows.
- Empty model: no `--model` flag emitted (byte-identical; `buildArgs` :735-737 gate) — the
  identity table's `""` row plus existing launch tests.
- `--model` via extraArgs (`moai cc -- --model opus[1m]`): same resolution path (:672-676 →
  :711); covered by the identity suite's form set; a launch-level case is optional.
- A stored value with mid-string `[1m]` (`opus[1m]-suffix`): not a table key, passes through
  verbatim both before and after — no special case exists or is added.
- Upper-case or whitespace-corrupted aliases (`Opus`, `opus `): pass through verbatim (unknown
  shape); Claude Code errors visibly if invalid. No launcher normalization is added.
- **Legacy bare-alias native resolution (F-1 residual, accepted)**: the live probe (spec.md
  §E.2, this plan-phase session) established CLI-level acceptance of bare `opus` only — exit 0,
  `is_error: false`, `subtype: success` — behind a GLM gateway (`modelUsage` recorded
  `glm-5.3-flash`), so the **Anthropic-native** alias→model mapping is UNOBSERVED. If the native
  side pins bare `opus` to an old model, a legacy stored bare-alias value (`opus`/`sonnet`/
  `fable`) launches that old model after this change, silent from MoAI's side — the AC set
  proves the launcher's argv, never CC's acceptance. Mitigations: the documented fixed-model
  path (a full ID passes through unchanged) and prefs-layer normalization keeping new saves on
  known aliases. Discharge path: a native-side (non-gateway) probe; accepted as residual, not a
  release blocker.
- The GLM reverse map receiving a full id (`claude-opus-4-8[1m]`): still reverse-maps to
  `opus[1m]` — unchanged by this SPEC (AC-ALP-005 rows).

## §D.3 Quality Gates

- TRUST 5 Tested: affected packages green (AC-ALP-011); the new tests are the characterization
  for the changed path and the preservation proof for the unchanged ones.
- Readable/Unified: `golangci-lint` clean (AC-ALP-011); comments updated with the code they
  describe (AC-ALP-010).
- Secured: no new input handling; the model string is passed to an exec'd binary exactly as
  before (argv array, no shell).
- Trackable: commit messages carry card id t1315; SPEC frontmatter `issue_number: 1730`.

## §D.4 Definition of Done

All Blocker + Major ACs pass with cited command output in progress.md §E.2 (five-section format:
Claim / Evidence / Baseline-attribution / Gaps / Residual-risk); the M1 RED evidence is present;
Minor AC-ALP-010 passes; `git diff --stat` over the PRESERVE list is empty.

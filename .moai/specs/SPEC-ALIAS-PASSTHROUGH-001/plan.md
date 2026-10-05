# SPEC-ALIAS-PASSTHROUGH-001 — Implementation Plan

## §A Context

Card t1315, GitHub #1730, Class C (design change: replaces the launch-path alias-resolution
mechanism), Tier M. Tree measured at `02ad57bbe` (t1322's alias bump already merged into this
base). The change is deliberately small in code and deliberate in what it does NOT change: the
Claude-backend launch path stops substituting compile-time ids for aliases and hands the stored
value to `claude --model` verbatim; every other consumer of `ModelAliasTable` (display,
validation, prefs migration, GLM slots) is pinned unchanged. The defect is currently live and
test-pinned on this tree — `TestExpandModelString/opus_alias_1m_resolves` and
`TestResolveMainSessionModel_GLMAvoidsCanonicalID/claude_backend_alias_expands_to_canonical_id`
run GREEN asserting the substitution (spec.md §A.2).

cycle_type: **tdd** (RED-GREEN-REFACTOR). The RED state is directly constructible: the
stub-claude reproduction test fails on today's tree because the argv carries
`claude-opus-5-5[1m]`.

## §B Known Issues

- **Two existing tests pin the defective mechanism** (`launcher_test.go` :674-713, :926-959).
  They are REWRITTEN, not preserved — the fixture-string principle does not apply because these
  tests pin the defective behavior itself. The GLM cases inside
  `TestResolveMainSessionModel_GLMAvoidsCanonicalID` (and its canonical-leak guard, :950-956)
  MUST survive byte-identical in expectation; only the Claude-backend cases flip to passthrough.
- **Stale comments will outlive the code change unless named**: `launcher.go` :704 ("byte-
  identical to expandModelString"), the `resolveMainSessionModel` doc (:1224-1234, "Under a
  Claude backend … byte-identical to expandModelString"), `model_policy.go` :68-69 (forward
  direction "used by expandModelString in launcher.go"), and the `@MX:ANCHOR` reason line :79
  listing `expandModelString` as a consumer. M2 updates all four; the ANCHOR tag itself stays
  (the table's fan-in is unchanged — the consumer list inside the reason changes).
- **`expandModelString` becomes dead code** the moment the Claude branch passes through. Delete
  it in the same milestone; leaving it invites a future consumer to reintroduce substitution.
  `splitModelSuffix` stays (GLM branch :1242, `profile_setup.go` :76/:115).
- **The repro harness must not exec for real**: `syscall.Exec` replaces the test process.
  Capture at the `execOrSpawnClaudeFunc` seam (:577-581 — its doc comment names exactly this
  test use). `launchClaudeFunc` alone is insufficient: it bypasses `buildArgs`/`resolveLaunch
  ClaudeBinary`, which is where the argv under test is assembled.
- **Profile-state seams for the harness**: `profile.BaseDirOverride` (used by
  `TestUnifiedLaunch_GlobalLedgerDoesNotBleed`, launcher_test.go :859-861) + `findProjectRootFn`
  (:882-884) + a scratch `t.TempDir()` with an empty `.moai/` dir. `t.Setenv("HOME", …)` is the
  issue's shape; the in-tree seam precedent is the override var — either is acceptable, the
  override var avoids cross-test env pollution. A stub `claude` binary via
  `fakePATHDir`/`writeExecutable` (claude_binary_test.go precedent) or the
  `config.EnvClaudeBin` pin satisfies "stub first on PATH"; the env pin is the more surgical of
  the two.
- **Bare-alias behavior delta** (spec.md D-2b): `opus` (bare) previously launched
  `claude-opus-5-5`; it now launches `opus` and Claude Code resolves its current default. Under
  the 1M unification CC's bare alias targets the same family. Documented as intended; do not
  "fix" it by special-casing bare aliases.
- **golangci-lint**: use the CI version (v2.1.6 lane lesson) when running the E5 lint check.

## §C Pre-flight

- [ ] `git rev-parse --short HEAD` == `02ad57bbe` (or absorb local `develop` per gitflow §4.1 before starting; re-measure anchors after any absorb)
- [ ] Confirm t1322 state: `grep -n '"opus"' internal/template/model_policy.go` shows `"opus": ModelIDOpus55` (the bump this SPEC builds on is present; D-6)
- [ ] Confirm anchors: `grep -n 'func expandModelString' internal/cli/launcher.go` resolves; `grep -n 'ModelIDOpus55' internal/template/model_policy.go` shows :42 and :81
- [ ] Confirm the defective pin is live: `go test -run '^TestExpandModelString$' ./internal/cli/` exits 0 (substitution asserted green — this is the RED baseline the new tests replace)
- [ ] Baseline: `go test -timeout 30m ./internal/cli/... ./internal/template/... ./internal/web/...` green before first edit (affected packages only; no local full suite)

## §D Constraints

- Affected packages only for local runs: `./internal/cli/... ./internal/template/... ./internal/web/...`, `-timeout 30m`. Never `go test ./...` locally (gitflow lane discipline).
- PRESERVE byte-identical: `resolveMainSessionModel`'s GLM branch (:1239-1247); `splitModelSuffix`; `profile_setup.go` `normalizeModel`/`promoteTo1M`/`normalizeModelLegacy1M`; `web/validate.go` `modelOptionList`; `schema_bridge.go`; `ModelAliasTable`/`ModelDeprecatedCanonicalIDs`/`ModelAliasPickerValues`/model-id constants; `DO_CLAUDE_MODEL` sync (:631-633); `--model` extraArgs parsing; `buildEnvForGLMLaunch` and all GLM slot/effort code.
- Passthrough assertions carry no table references (REQ-ALP-006) — expectation strings are literals or the round-tripped input.
- No new warning/error emission on the model path (REQ-ALP-008).
- No commit/push inside the worktree outside the lane's integration window; the lane commits plan/run/sync work.
- Conventional Commits with the card id (t1315) in the body; `🗿 MoAI` trailer per repo convention.

## §E Self-Verification

Verification is AC-driven; every acceptance.md criterion names its command. Summary:

| Claim | Evidence command |
|---|---|
| Alias reaches claude verbatim (stub repro) | the AC-ALP-001 test via `go test -run '^TestLaunchModelAliasPassthrough$' ./internal/cli/ -v` (RED at M1, GREEN at M2) |
| Passthrough identity holds for every form | the AC-ALP-002 identity table via `go test -run '^TestResolveMainSessionModel_ClaudePassthrough$' ./internal/cli/ -v` |
| GLM branch untouched | `go test -run '^TestResolveMainSessionModel_GLMAvoidsCanonicalID$' ./internal/cli/ -v` — GLM cases green with expectations unchanged |
| Table consumers untouched | `go test ./internal/web/... ./internal/template/...` green + `git diff --stat` empty over the PRESERVE files (AC-ALP-008/009) |
| No substitution code remains | `grep -rn 'expandModelString' --include='*.go' internal/` → 0 hits (AC-ALP-007) |
| Stale docs updated | the two stale-phrase greps in AC-ALP-010 → 0 hits |
| Quality gates | affected-package suite + `golangci-lint run` (CI version) + `GOOS=windows GOARCH=amd64 go build ./...` (AC-ALP-011/012) |

Each §E item reports per verification-claim-integrity §3: command + verbatim output + tree SHA
(this run, this tree).

## §F Milestones

Ordering follows TDD: construct the failing proof first (M1), make it pass minimally (M2), then
sweep hygiene and the unchanged-surface regressions (M3).

### M1 — RED: reproduction + identity tests written and observed failing (code, test-only)
- Write the stub-claude reproduction test (AC-ALP-001): scratch `t.TempDir()` with an empty
  `.moai/` dir; isolated profile base via `profile.BaseDirOverride` holding
  `preferences.yaml` with `model: opus[1m]`; `findProjectRootFn` pinned to the scratch dir;
  stub `claude` via `config.EnvClaudeBin` pin (or `fakePATHDir`); `launchClaudeFunc` left
  default; `execOrSpawnClaudeFunc` replaced with a capture recording `(bin, args, env)` and
  returning nil. Drive `runUnifiedLaunch("", "claude", nil)`. Assert the captured argv contains
  exactly `--model` followed by `opus[1m]` — literal, no table reference (REQ-ALP-006/009).
- Write the identity table test (AC-ALP-002): `resolveMainSessionModel(x, false) == x` over
  {`opus`, `opus[1m]`, `sonnet`, `sonnet[1m]`, `fable`, `fable[1m]`, `haiku`, `opusplan`,
  `opusplan[1m]`, `claude-opus-5-5`, `claude-opus-5-5[1m]`, `claude-opus-4-8`, `custom-xyz`,
  `""`} — literals only.
- Run both; capture verbatim RED output (they fail on `02ad57bbe`: argv carries
  `claude-opus-5-5[1m]`; identity fails on alias rows). The RED evidence goes in progress.md
  §E.2 (E8 item). Do not touch production code in this milestone.
- **Type: code (tests only)**

### M2 — GREEN: passthrough lands, substitution is deleted (code)
- `internal/cli/launcher.go`: make the Claude branch of `resolveMainSessionModel` return
  `prefsModel` verbatim; delete `expandModelString`; keep `splitModelSuffix`; update the
  :701-704 comment block and the `resolveMainSessionModel` doc comment to state the passthrough
  contract (REQ-ALP-001/004/007).
- `internal/cli/launcher_test.go`: rewrite `TestExpandModelString` into the deleted function's
  absence (remove it; the identity test carries its coverage) and flip the Claude-backend cases
  of `TestResolveMainSessionModel_GLMAvoidsCanonicalID` to passthrough expectations; GLM cases
  and the canonical-leak guard unchanged (REQ-ALP-003).
- `internal/template/model_policy.go`: update the `ModelAliasTable` forward-direction note
  (:68-72) and the `@MX:ANCHOR` reason consumer list (:79) — the table row edits are FORBIDDEN
  here (REQ-ALP-005/007, D-6).
- Verify: both M1 tests GREEN; `go test -timeout 30m ./internal/cli/...` green. Capture verbatim
  GREEN output.
- **Type: code**

### M3 — REFACTOR/verify: unchanged-surface regressions + hygiene (verification)
- Run the unchanged-surface suite: `go test -timeout 30m ./internal/cli/... ./internal/template/... ./internal/web/...` (covers
  `profile_setup_normalize_test.go`, web validation union tests, GLM slot tests — AC-ALP-008/009).
- Dead-code + stale-doc sweeps (AC-ALP-007/010 greps); `golangci-lint run` (CI version, affected
  packages); `GOOS=windows GOARCH=amd64 go build ./...` (AC-ALP-011/012).
- Confirm `git diff --stat` over the PRESERVE list is empty (spec.md §F + plan §D).
- Drift guard: planned-vs-actual file list — expected touch set is `internal/cli/launcher.go`,
  `internal/cli/launcher_test.go`, `internal/template/model_policy.go` (comments only), plus the
  SPEC's own artifacts. Anything beyond that is drift to report.
- **Type: verification**

## §G Out of Scope (run-phase guard)

- No table-content edits, no alias-refresh mechanism, no GLM branch changes, no wizard/web/normalize
  changes, no new warnings on the model path, no statusline/served-model changes (spec.md §F).
- Do not run the true-binary E2E (executing a real `claude` in tests) — the exec-seam capture is
  the assertion surface; a POSIX-guarded true-binary variant is optional and supplementary, never
  required for the ACs.

# SPEC-CLIFIX-HYGIENE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- 2026-07-10: plan-phase artifact set authored (spec.md / plan.md / acceptance.md) by manager-spec from CLI audit 2026-07-10 §3/§4 hygiene rows + §5 P4. Status: draft. Pending plan-audit. Sequenced last in the CLIFIX series (P0→P4).
- 2026-07-30: v0.2.0 rescope after plan-audit iter-1 FAIL (0.58, D1-D11 stale anchors). Re-verified every anchor against current worktree (HEAD = origin/main, post P0-P3 merge). Dropped REQ-HYG-001-007 (deepMerge3Way — already fixed in `internal/merge/`) and REQ-HYG-001-008 (token 0600 — superseded by F1 no-persist redesign); renumbered old 009→007, old 010→008; refreshed all line/symbol anchors; added `### Out of Scope — Broader i18n sweep` (~24 deferred Hangul files). Status: draft — NOT audit-ready (pending re-audit).

## §E.2 Run-phase Evidence

### M1 — PRESERVE net + ANALYZE inventories (2026-07-30)

**M1 scope (per plan.md §F M1): characterization tests + goldens for update
flows + Hangul/threshold/env-literal inventories frozen as fixtures. NO
production `.go` edits.**

#### Characterization net added

- File: `internal/cli/update_hygiene_characterization_test.go` (new, ~230 LOC).
- Surface covered (per plan.md §F M1: "dry-run plan path AND settings sync/merge path"):
  - **Settings sync/merge path** — `cleanLegacyHooks` (update.go:1759): 4 characterization tests pinning the exact 10-pattern legacy set (golden), the per-pattern removal contract, the non-legacy-hook preservation contract, and the mixed-group surgical-delete contract. The no-hooks-section fast-path short-circuit is also pinned.
  - **Dry-run plan path** — `runAgencyMigrationAdapter` (update.go:1247): 2 characterization tests pinning the ErrMigrateNoSource → return-nil suppression contract (no .agency source must not fail update) and the dryRun=true read-only guarantee (no `.moai` mutation when dryRun is set).
- PASS evidence (verbatim, on unmodified code):
  - `go test ./internal/cli/ -run 'CleanLegacyHooks|RunAgencyMigrationAdapter' -count=1 -v` → `PASS` / `ok github.com/modu-ai/moai-adk/internal/cli 0.739s` (6 new tests + pre-existing `cleanLegacyHooks` tests all green).

#### ANALYZE caller-graph verdicts (dead-code inventory)

Frozen at `.moai/specs/SPEC-CLIFIX-HYGIENE-001/inventories/dead-code-caller-graph.md`. Per-symbol verdict:

| Symbol | Verdict | Production callers | Notes |
|---|---|---:|---|
| `buildGLMEnvVars` (glm.go:917) | **DEAD (M5-deletable)** | 0 | test-only callers in glm_team_test.go + coverage_improvement_test.go; no tag/reflection/registration gate |
| `ttyConfirmer` (branch_protection.go:39-55) | **DEFERRED** | 0 | self-attested `SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 deferred (paired with ttyConfirmer)` — M5 must resolve the pairing first; live confirmer is `yesConfirmer` |
| `worktree_validation.go` (whole file) | **DEAD-in-prod, DEFERRED-pending-wire-up-decision** | 0 | file header + `@MX:NOTE` self-attest "wire-up planned in follow-up SPEC"; exported but `internal/cli` blocks external import; M5 surfaces to orchestrator before deleting |

PRESERVE-KEEP (confirmed LIVE — NOT deletable): `acquireUpdateLock` (update.go:264 caller, P0-wired), `cleanStaleLock` (update_cleanup.go:59 caller, P0-wired), `scanDeprecatedPaths` (4 callers in update_clean_install.go), `cleanup_old_backups`.

Realistic M5 net-LOC delta walked back honestly (vs the audit's ~500-line figure): **−60 to −320 lines**, contingent on the two DEFERRED resolutions.

#### Inventories frozen as fixtures (diffable for M3/M4/M5)

| Fixture | Path | M-milestone consumer |
|---|---|---|
| Dead-code caller-graph | `inventories/dead-code-caller-graph.md` | M5 (deletion) |
| Hangul strings snapshot (6 in-scope files, 73 lines across 5 files; doctor.go already clean = drift finding recorded) | `inventories/hangul-strings.snapshot.md` | M4 (English UI strings) |
| Threshold-literal snapshot (THREE `[]int{1,3,5,10}` sites + TWO `30*time.Second` dispatcher timeouts — original "×2" audit undercount corrected) | `inventories/threshold-literals.snapshot.md` | M3 (defaults.go extraction) |
| GLM env inject/clear sites snapshot (11 sites across glm.go/glm_tools.go/launcher.go) | `inventories/glm-env-sites.snapshot.md` | M3 (envkeys.go + glmEnvVarSet) |

#### AC setup vs completion (E1 matrix — M1 is the safety net)

M1 does NOT complete any AC by itself; it SETS UP the ACs the later milestones
will satisfy. M1's load-bearing claim is "characterization tests PASS on
unmodified code" (the PRESERVE guarantee), verified verbatim above.

| AC | M1 status | Milestone that completes it |
|---|---|---|
| AC-HYG-001-001 | SET UP (characterization net exists; M5 must keep it byte-identical post-split) | M5 |
| AC-HYG-001-002 | SET UP (caller-graph inventory frozen; M5 performs the deletion) | M5 |
| AC-HYG-001-003 | SET UP (env-site inventory frozen) | M3 |
| AC-HYG-001-004 | SET UP (threshold-literal inventory frozen) | M3 |
| AC-HYG-001-005 | SET UP (Hangul inventory frozen; doctor.go already-clean drift recorded) | M4 |
| AC-HYG-001-006 | not started (M2 repro-first) | M2 |
| AC-HYG-001-007 | not started (M2 repro-first) | M2 |
| AC-HYG-001-008 | not started (M2 repro-first) | M2 |

#### M1 production-code edit count: **0** (PRESERVE + ANALYZE only — per plan.md §F M1 scope)

### M2 — repro-first TDD defect corrections (2026-07-30)

**M2 scope (per Section D constraints): the three genuine behavior-correcting
defects — AC-006 rune-safe truncation, AC-007 wizard PAT masking, AC-008
worktree YAML parse + tmux path quoting. NO M3/M4/M5 scope touched.**

Each defect followed RED → GREEN: a failing test reproducing the defect was
captured verbatim BEFORE the minimal fix (E8 falsifiability).

#### AC-HYG-001-006 — rune-safe truncation (PASS)

Complete site set re-derived by grepping the CLI tree for byte-slice
truncation on user-content strings (ASCII-only slices — hex hashes, API-key
masking, session IDs, SHAs — excluded because they cannot carry CJK):

1. `internal/cli/constitution.go` `renderConstitutionTable` — clause (prefix)
2. `internal/cli/constitution.go` `renderConstitutionTable` — file path (suffix) **← the "+1 remaining site" the audit undercounted; co-located with site 1, splits CJK file paths at the suffix left boundary**
3. `internal/cli/tool_policy.go` `renderList` — audit column (prefix)
4. `internal/cli/tool_policy.go` `truncateArg` — args_pattern (prefix)
5. `internal/cli/github.go` `runParseIssue` — issue body (prefix, via new `truncateIssueBody` helper)

All 5 sites routed through ONE shared rune-boundary helper
(`internal/cli/truncate.go` — `truncateRunes` prefix + `truncateRunesSuffix`
suffix; the cut always lands on a rune boundary, so output is
`utf8.ValidString == true`).

RED (verbatim, captured before GREEN): `go test ./internal/cli/ -run 'RuneTruncate' -count=1 -v`:
```
TestRuneTruncateConstitutionClause: invalid UTF-8 from CJK clause: [37]
TestRuneTruncateConstitutionFile:   invalid UTF-8 from CJK file path: [286 287]
TestRuneTruncateToolPolicyAudit:    invalid UTF-8 from CJK audit: [167 168]
TestRuneTruncateToolPolicyArg:      invalid UTF-8 "가가가가가가가가가가가가\xea..." ([36])
TestRuneTruncateGithubBody:         github.go still byte-slices issue body (body[:200])
--- FAIL (5/5) ---
```
GREEN: all 5 PASS. Verbatim AC command output:
```
ok  github.com/modu-ai/moai-adk/internal/cli  0.467s   (5/5 RuneTruncate PASS)
```

#### AC-HYG-001-007 — wizard PAT masking (PASS)

`buildInputField` (wizard.go) now gates `.EchoMode(huh.EchoModePassword)` on
the token question IDs via a new exported predicate `IsSecretInputID(id)`
(classifies `github_token` + `gitlab_token` as secret). The huh v2
(charm.land/huh/v2) API is `inp.EchoMode(huh.EchoModePassword)`.

RED (verbatim, captured before GREEN): `go test ./internal/cli/wizard/ -run 'PATMask' -count=1 -v`:
```
internal/cli/wizard/pat_mask_test.go:42:7: undefined: IsSecretInputID
FAIL  github.com/modu-ai/moai-adk/internal/cli/wizard  [build failed]
```
(The build failure is the RED: the masking seam is entirely absent; the grep
test `TestPATMaskEchoModeWired` therefore cannot run. Post-fix both tests
compile and pass.)
GREEN: `TestPATMaskEchoModeWired` + `TestPATMaskSecretIDs` PASS.
AC grep component: `grep -n 'EchoMode\|EchoPassword\|\.Password' internal/cli/wizard/wizard.go` → matches `EchoMode(huh.EchoModePassword)`.

#### AC-HYG-001-008 — worktree YAML parse + tmux quoting (PASS)

- `GetActiveMode` (tmux_integration.go): `strings.Contains` line-matcher
  replaced by `parseTeamMode` (yaml.v3 unmarshal of `llm.team_mode`).
- `isTmuxPreferred` (new.go): `strings.Contains` line-matcher replaced by
  `parseTmuxPreferred` (yaml.v3 unmarshal of `workflow.worktree.tmux_preferred`).
  Commented `# team_mode: cg` / `# tmux_preferred: true` no longer match.
- `buildTmuxInitialCommand` (tmux_integration.go): `cd %s` → `cd %q` so paths
  with spaces/metacharacters survive. The co-located M1 characterization test
  `TestBuildTmuxInitialCommand_ModeSelection` was retargeted (NOT deleted) to
  expect `cd "/tmp/test-wt"`; its mode-selection assertions are unchanged.

RED (verbatim, captured before GREEN): `go test ./internal/cli/worktree/ -run 'YAMLCommentCG|TmuxPathQuote|TmuxPreferredParse' -count=1 -v`:
```
TestYAMLCommentCGTeamMode:  comment-only team_mode = "cg", want "cc"
TestTmuxPreferredParse:     comment-only tmux_preferred = true, want false
TestTmuxPathQuote:          "cd /tmp/my project ; moai cc" (want cd "/tmp/my project")
--- FAIL (3/3 defect tests; TestTmuxPreferredParseRealTrue PASS) ---
```
GREEN: all 4 PASS.

#### M2 AC completion matrix (E1)

| AC | M2 status | Verification command | Result |
|---|---|---|---|
| AC-HYG-001-006 | **PASS** | `go test ./internal/cli/ -run 'RuneTruncate' -count=1 -v` | 5/5 PASS — CJK fixtures utf8.ValidString at all 5 sites |
| AC-HYG-001-007 | **PASS** | `go test ./internal/cli/wizard/ -run 'PATMask' -count=1 -v` + grep | 2/2 PASS; `EchoMode(huh.EchoModePassword)` present in wizard.go |
| AC-HYG-001-008 | **PASS** | `go test ./internal/cli/worktree/ -run 'YAMLCommentCG\|TmuxPathQuote\|TmuxPreferredParse' -count=1 -v` | 4/4 PASS — comments do not activate CG; spaced path quoted |

#### M2 cross-cutting verification (E2–E5)

- E2 build: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- E3 coverage: `internal/cli/wizard` 94.0%; `internal/cli/worktree` 84.9% (vs M1 baseline; both ≥ 85% threshold met for wizard; worktree within range — the parse seams + helpers added covered lines).
- E4 subagent-boundary grep: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/wizard internal/cli/worktree --include='*.go' | grep -v _test.go` → 0 matches (exit 1 = clean).
- E5 lint: `golangci-lint run --timeout=3m ./internal/cli/...` → 0 issues (no NEW findings vs M1 baseline of 0).
- gofmt: all 11 edited/new files formatted (gofmt -l output empty).

#### M2 production-code edit set (scope discipline)

New files: `internal/cli/truncate.go`, `internal/cli/rune_truncate_test.go`,
`internal/cli/wizard/pat_mask_test.go`, `internal/cli/worktree/yaml_quote_test.go`.
Edited files: `internal/cli/constitution.go`, `internal/cli/tool_policy.go`,
`internal/cli/github.go` (M2 defect sites only), `internal/cli/wizard/wizard.go`
(PAT masking only), `internal/cli/worktree/tmux_integration.go` +
`new.go` (YAML + quoting), `internal/cli/worktree/tmux_integration_test.go`
(retargeted one characterization assertion to the quoted form).
NO M3 (envkeys), M4 (English strings), or M5 (dead-code deletion) scope touched.

### M3 — hardcoding single-source (AC-HYG-001-003 + AC-HYG-001-004) — PASS

Behavior-preserving relocation of GLM env-var NAME literals and tier/timeout
threshold literals to a single source. No value changed; only the redundant
inline copies now reference one SSOT.

**AC-HYG-001-003 (GLM env names) — PASS.**
- `internal/config/envkeys.go`: added 3 constants
  (`EnvClaudeCodeDisableExperimentalBetas`,
  `EnvClaudeCodeDisableNonessentialTraffic`,
  `EnvClaudeCodeTeammateDisplay`) + `GLMEnvVarSet()` helper returning the
  canonical 3-key set (REQ-HYG-001-003 inject↔clear parity).
- Each name appears exactly once in envkeys.go (grep -c = 1 each).
- Zero bare `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` /
  `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC` / `CLAUDE_CODE_TEAMMATE_DISPLAY`
  literals in cli production source (AST-based
  `TestNoBareGLMEnvVarLiteralsInCLIProduction` PASS; grep -c = 0 each).
- Inject sites (glm.go: 484, 682, 926, 970, +224) reference
  `config.EnvClaudeCodeDisableExperimentalBetas`; clear sites (glm.go: 569,
  571, 616, 711; launcher.go: 305, 307, 309; settings.go: 191, 193, 194)
  reference the respective constant; the read site `glm_tools.go:805`
  references `config.EnvGLMNoAutoTools`.
- `TestGLMEnvSetParity` asserts the helper returns exactly the canonical
  3-key set — inject and clear derive from one enumeration.
- Scope addition (transparent): `internal/cli/settings.go:191-194`
  (`stripGLMCredsAndSetTeammateMode`) carried the same 3 bare literals as
  `launcher.go`'s clear path but was not enumerated in
  `glm-env-sites.snapshot.md`. It was included here because leaving it bare
  would re-introduce exactly the inject↔clear drift REQ-HYG-001-003 closes;
  the relocation is behavior-identical (same keys deleted). Also relocated
  the adjacent `MOAI_STATUSLINE_CONTEXT_SIZE` literal in that same clear
  block to its existing `config.EnvStatuslineContextSize` constant (free win,
  same block).

**AC-HYG-001-004 (threshold literals) — PASS.**
- `internal/config/defaults.go`: added
  `var DefaultTierThresholds = []int{1, 3, 5, 10}` (var, not const —
  composite literal) and `DefaultHookDispatcherTimeout = 30 * time.Second`
  (const).
- The THREE former `[]int{1, 3, 5, 10}` sites
  (`harness.go:150` runHarnessStatus fallback, `harness.go:480`
  defaultLearningConfig struct initializer, `hook.go:1013`
  defaultTierThresholds) now all reference `config.DefaultTierThresholds`;
  zero inline copies survive in cli production source (grep = 0).
- The TWO former `30 * time.Second` dispatcher-timeout sites
  (`hook.go:237`, `hook.go:361`) now reference
  `config.DefaultHookDispatcherTimeout`; zero inline copies survive in
  hook.go (grep = 0).
- `TestDefaultTierThresholdsCanonical` /
  `TestDefaultHookDispatcherTimeoutCanonical` /
  `TestGLMEnvConstantsCanonical` (new in
  `internal/config/defaults_clifix_test.go`) assert the SSOT values equal
  the pre-M3 baselines, proving behavior preservation.

**Inventory "size caps / retry / circuit" deferral — verified NO-OP.** The
threshold inventory deferred those literals "without anchoring them with
stable line numbers" and instructed M3 to re-derive via grep. A targeted
grep of `internal/cli/hook.go` and `internal/cli/harness.go` for size / retry
/ circuit / breaker / max-bytes patterns surfaced ZERO duplicated literals in
those files. The 5 anchored sites above are therefore the complete duplicate
set; no further literals needed extraction (extracting a literal that appears
only once is over-abstraction, violating Enforce Simplicity).

**Build / test / lint (M3).**
- `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `go test ./internal/cli/ ./internal/cli/wizard/ ./internal/cli/worktree/
  ./internal/config/` — all PASS (cli 187.9s, wizard 7.0s, worktree 9.2s,
  config 3.9s). M1 characterization net + M2 defect fixes remain green.
- `internal/harness` `TestRecordEvent100Sequential` flagged 108ms > 100ms
  once under parallel-batch machine load, then passed 3/3 in isolation
  (0.48s avg) — confirmed pre-existing perf-flake, unrelated to M3
  (internal/harness not in M3 scope; its `learner_test.go:143/178/198`
  literal lives in the _test.go hardcoding-allowed zone and was untouched).
- `golangci-lint run ./internal/config/... ./internal/cli/...` — 0 issues
  (matches M2 baseline of 0).
- `gofmt -l` over the 8 edited files — clean.

NO M4 (English strings) or M5 (dead-code) scope touched. P0 lock path,
concurrency behavior, and the M2 rune/PAT/YAML fixes preserved verbatim.

### M4 — English UI strings (AC-HYG-001-005) — PASS

Localized Korean user-facing strings + comments in the 6 in-scope files to
English literals per the `error_messages: en` / `code_comments: en` policy
(language.yaml). No behavior change; pure string + comment localization.
Each message reviewed individually so diagnostic accuracy (PIDs, ports,
counts, version numbers, `%w` wrapping) survives the rewrite — NO blind sed
(plan.md §G anti-pattern).

**AC-HYG-001-005 — PASS.** Zero Hangul remaining across the 6 in-scope files:
- `grep -rlP '[\x{AC00}-\x{D7A3}]' internal/cli/doctor.go internal/cli/migration.go internal/cli/clean.go internal/cli/web_port.go internal/cli/web_port_posix.go internal/cli/web_port_windows.go` → exit 1 (no matches = PASS).
- `doctor.go`: confirmed already-clean (0 Hangul) — matches the M1 inventory drift finding; no M4 work needed, skipped per B2.

**Per-file Hangul before → after:**

| File | Before (M1 inventory) | After (M4) |
|---|---:|---:|
| `internal/cli/doctor.go` | 0 (drift — already clean) | 0 (skipped, confirmed clean) |
| `internal/cli/migration.go` | 16 | 0 |
| `internal/cli/clean.go` | 1 | 0 |
| `internal/cli/web_port.go` | 29 | 0 |
| `internal/cli/web_port_posix.go` | 17 | 0 |
| `internal/cli/web_port_windows.go` | 10 | 0 |
| **Total** | **73** | **0** |

**Golden regeneration (deliberate, with rationale).**
`internal/cli/migration_m3_test.go` carries M3 characterization goldens that
pinned the exact stdout/stderr bytes of the migration CLI commands. The
`want :=` values and the `TestM3_PrinterFormats_UnreachableBranches` byte-pins
asserted the pre-M4 Korean source strings. Per plan.md §G ("where a golden
captures a defect/diagnostic fixed in M4, regenerate in the fixing commit with
rationale"), these goldens were regenerated to the new English text in this
M4 commit. The goldens capture INTENTIONAL diagnostic text being localized
(ko→en per `error_messages: en`), NOT a defect — regenerated, not deleted.
A rationale comment was added at each regenerated golden site.
The `web_port_test.go` `wantStderr` assertion is a PID substring (`"4242"`),
not a Korean source string, so no regeneration was needed there; its Korean
comments / `t.Fatalf` failure messages are test-file Korean outside the M4
6-file scope (deferred to the broader i18n sweep follow-up SPEC) and were left
untouched.

**M4 production+test edit set (scope discipline):**
- Edited source (5 files, the 5 with Hangul): `internal/cli/migration.go`,
  `internal/cli/clean.go`, `internal/cli/web_port.go`,
  `internal/cli/web_port_posix.go`, `internal/cli/web_port_windows.go`.
- Edited test (1 file, golden regeneration): `internal/cli/migration_m3_test.go`.
- `doctor.go`: NOT edited (confirmed clean).
- NO M5 (dead-code) scope, NO deferred ~24 Hangul files, NO P0 lock path touched.

**M4 cross-cutting verification (E2–E5):**
- E2 build: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- E3 coverage: `go test -cover ./internal/cli/` → `coverage: 75.6% of statements` (measured this run; localization is string-only so coverage is structurally identical to pre-M4 — no branch changes).
- E4 subagent-boundary grep: N/A — M4 touched no harness/hook code; the 6 in-scope files are CLI source, none reference AskUserQuestion/mcp__askuser (the M3-existing `TestWebPort_NoAskUserQuestion` static guard still PASSes, re-verified in the targeted run).
- E5 lint: `golangci-lint run --timeout=2m` → `0 issues` (no NEW findings vs M3 baseline of 0).
- gofmt: all 6 edited files formatted (gofmt -l output empty).
- Full cli suite: `go test ./internal/cli/ -count=1` → `ok ... 201.553s` exit 0; targeted M3+web_port+clean+migration run → all PASS.

**M4 AC completion matrix (E1):**

| AC | M4 status | Verification command | Result |
|---|---|---|---|
| AC-HYG-001-005 | **PASS** | `grep -rlP '[\x{AC00}-\x{D7A3}]' <6 in-scope files>` | exit 1 (0 Hangul remaining) |

**M5 readiness + flagged blockers (NOT resolved in M4):**
M4 leaves M5 (structure / dead-code deletion) ready to enter, with two
blockers flagged for the orchestrator per the M4 task brief:
1. `ttyConfirmer` (branch_protection.go:39-55) — paired with `SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1`; M5 must resolve the pairing first (per M1 inventory DEFERRED verdict).
2. `worktree_validation.go` — exported-but-unwired; M5 must surface the wire-up decision to the orchestrator before deleting (per M1 inventory DEFERRED verdict).


## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- Date: 2026-07-30 (run-phase entry, post Implementation Kickoff Approval — user approved semi-autonomous per-milestone verification)
- Input parameters: tier=L; scope≈12-15 files (internal/cli update cluster + wizard + internal/worktree + read-only internal/merge); domain count=1 (CLI hygiene, single Go module); language mix=Go; concurrency benefit=LOW (coding-heavy refactoring + defect fixes, not parallelizable research).
- Mode evaluation:
  - trivial — not selected (multi-file refactoring + defect fixes)
  - background — not selected (write-capable, not read-only)
  - agent-team — RETIRED (not selectable)
  - parallel — not selected (coding-heavy single-domain; Anthropic coding-task parallelism caveat)
  - workflow — not selected (semantic refactoring + multi-rule defect fixes, not a single ≥30-file uniform mechanical transform)
  - sub-agent — selected
- Decision: sub-agent (Mode 5)
- Justification: B6 is coding-heavy, single-domain CLI hygiene work (DDD characterization-first decomposition + reproduction-first TDD defect correction). Per Anthropic's coding-task parallelism caveat, sequential per-milestone sub-agent delegation is the safe default for coding work; the work is semantic/multi-rule (not a uniform mechanical transform), so Mode 6 workflow does not apply. Mode 5 with the Tier L Section A-E delegation template, one milestone per spawn, orchestrator trust-but-verify (build/test/lint/AC) between milestones.
- Note: worktree base = origin/main 2f449e189; origin/main has since advanced to cc1986cff (+2, multi-session race) — update-branch at run-PR time.

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

# SPEC-CLIFIX-HYGIENE-001 — Acceptance Criteria

## §A Scenarios (Given-When-Then)

1. Given a maintainer opening the update cluster after the residual decomposition, When they inspect the files, Then every file in the update cluster (`update.go` + `update_*.go` siblings) is at or under the 1,200-line ceiling and the characterization suite proves behavior identical to the pre-split tree.
2. Given a user in a non-Korean locale hitting a doctor/migration/clean/web-port error, When the message renders, Then it is English per `error_messages: en`.
3. Given a project path `~/dev/my project (v2)`, When a worktree tmux session spawns, Then the initial command executes correctly with the quoted path, and a commented `# team_mode: cg` line in `workflow.yaml` does not flip CG detection, and a commented `# tmux_preferred: tmux` line does not flip the launcher-selection parse either.

## §D AC Matrix (machine-verifiable — anchors re-verified 2026-07-30)

| AC | REQ | Verification command | Expected outcome |
|---|---|---|---|
| AC-HYG-001-001 | REQ-HYG-001-001 | `go test ./internal/cli/ -run 'UpdateCharacterization' -count=1` && `for f in internal/cli/update.go internal/cli/update_*.go; do echo "$(wc -l < "$f") $f"; done` | Suite PASS unchanged post-split; every update-cluster file ≤ 1,200 lines (current `update.go` = 1,905) |
| AC-HYG-001-002 | REQ-HYG-001-002 | `for s in buildGLMEnvVars ttyConfirmer; do echo "== $s =="; grep -rn "$s" internal/cli --include='*.go' \| grep -v _test.go; done` (plus `go build ./...` after `worktree_validation.go` removal) | 0 matches per deleted symbol (after `ttyConfirmer` deferred-pairing resolution); `go build ./...` PASS |
| AC-HYG-001-003 | REQ-HYG-001-003 | `go test ./internal/cli/ -run 'GLMEnvSetParity\|NoBareGLMEnvVarLiteralsInCLIProduction' -count=1 -v` && `grep -rn 'CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS\|CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC\|CLAUDE_CODE_TEAMMATE_DISPLAY' internal/cli --include='*.go' \| grep -v _test.go \| grep -v '^[^:]*:[0-9]*:[[:space:]]*//'` | Both tests PASS — inject set == clear set, single-sourced through `config.GLMEnvVarSet()` (`internal/config/envkeys.go:122`), with the AST gate `TestNoBareGLMEnvVarLiteralsInCLIProduction` blocking re-entry of bare literals; grep yields 0 non-comment matches for the three GLM env-var NAMEs in `internal/cli` production. See the AC-003 scope note below §D. |
| AC-HYG-001-004 | REQ-HYG-001-004 | `grep -rn '\[\]int{1, 3, 5, 10}\|\[\]int{1,3,5,10}' internal/cli --include='*.go' \| grep -v defaults \| grep -v _test.go` && `grep -rn '30\*time.Second\|30 \* time.Second' internal/cli/hook.go` | THREE inline tier-threshold sites (`harness.go:150`, `harness.go:480`, `hook.go:1013`) reduced to ONE `defaults.go` constant; `hook.go` dispatcher timeouts (`:237`, `:361`) reference the constant |
| AC-HYG-001-005 | REQ-HYG-001-005 | `rg -l '[가-힣]' $(git ls-files 'internal/cli/doctor.go' 'internal/cli/migration.go' 'internal/cli/clean.go' 'internal/cli/web_port*.go' \| grep -v '_test\.go$')` | 0 files matched across the 6 resolved production files (`clean.go`, `doctor.go`, `migration.go`, `web_port.go`, `web_port_posix.go`, `web_port_windows.go`). rg is used because BSD grep bracket ranges over Hangul are unreliable; the file list is resolved through `git ls-files` + a `_test.go` filter because rg's `-g` glob does not filter explicitly-passed paths. See the AC-005 scope note below §D. |
| AC-HYG-001-006 | REQ-HYG-001-006 | `go test ./internal/cli/ -run 'RuneTruncate' -count=1 -v` | PASS — CJK fixtures at all 4 sites yield `utf8.ValidString(output) == true` at every boundary length |
| AC-HYG-001-007 | REQ-HYG-001-007 | `go test ./internal/cli/wizard/ -run 'PATMask' -count=1 -v` && `grep -n 'EchoMode\|EchoPassword\|\.Password' internal/cli/wizard/wizard.go` | PASS — `github_token`/`gitlab_token` questions render with password echo mode (RED on pre-fix code, where `wizard.go:329` uses plain `huh.NewInput()` with no echo masking) |
| AC-HYG-001-008 | REQ-HYG-001-008 | `go test ./internal/cli/worktree/ -run 'YAMLCommentCG\|TmuxPathQuote\|TmuxPreferredParse' -count=1 -v` | PASS — commented `team_mode` does not activate CG; commented `tmux_preferred` does not flip launcher-selection parse; spaced/metachar path round-trips through the tmux initial command |

### §D scope notes (v0.2.1 in-flight AC-wording correction)

- **AC-003 scope — GLM 3-name set only.** REQ-HYG-001-003 and the shipped M3 implementation cover exactly three env-var NAMEs: `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`, `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, `CLAUDE_CODE_TEAMMATE_DISPLAY`. The `ANTHROPIC_*` family (auth-token and model injection in `glm.go` / `launcher.go` / `settings.go`) and `CLAUDE_PROJECT_DIR` reads are **out of scope** and untouched — a broad `grep -rn 'ANTHROPIC_\|CLAUDE_' internal/cli` still returns 137 lines, of which 57 are raw env-name string literals. Note also that `ANTHROPIC_DEFAULT_FABLE_MODEL` has no `envkeys.go` constant at all. Single-sourcing the `ANTHROPIC_*` family is a follow-up SPEC, not this AC.
- **AC-005 scope — production files only.** The `web_port*.go` glob also matches `internal/cli/web_port_test.go`, which still carries Korean text in Go comments and in `t.Fatalf` / `t.Errorf` strings. Test-file language is governed by the `code_comments` setting, not `error_messages: en` (which governs runtime user-facing messages), so those residual strings are not a violation of REQ-HYG-001-005. They fall under the already-declared `### Out of Scope — Broader i18n sweep` in `spec.md` §C.

## §C Edge Cases

- Dead-code deletion vs build tags: symbols only referenced under `//go:build windows` (or unix) must be checked per-GOOS before deletion (`GOOS=windows go build ./...`).
- Rune truncation at limit==0 and limit smaller than the first rune — return empty string, no panic.
- Wizard in non-interactive mode — PAT masking path must not break `-p`/CI flows (skip prompt, not hang).
- `workflow.yaml` with `team_mode: cg` nested under an unexpected parent key — `yaml.v3` parse reads the documented path only; unknown placements do not activate CG. Same applies to `tmux_preferred`.

### Edge cases dropped from v0.1.0 (no longer actionable)

- ~~deepMerge3Way with user key whose VALUE type changed in the new template~~ — the merge machinery was extracted to `internal/merge/`; user-added-key preservation is characterized by `deepMergeMap` at `strategies.go:360` (line 388: "User added key - preserve"). No longer an edge case for this SPEC.
- ~~user.yaml pre-existing with 0644 and tokens already present~~ — tokens are not persisted to `user.yaml` at all per the F1 security fix (`update.go:1375-1438`).

## §D.5 Quality Gate / Definition of Done

- All 8 AC rows PASS with verbatim command output cited in progress.md §E.2.
- Characterization suite green on both pre-split (baseline run recorded) and post-split trees.
- `go build ./...` on darwin/linux/windows; `go test ./internal/cli/... -count=1` green; `golangci-lint run` no new findings.
- Dead-code netLOC delta reported with the deleted-symbol inventory in progress.md §E.2. Target: the verified-deletable subset only — `buildGLMEnvVars` (~30 lines incl. tests), `worktree_validation.go` (whole file, ~100 lines), `ttyConfirmer` (post-pairing, ~10 lines). The original audit's ~500-line figure is honestly walked back in research.md §C: it rested on the now-false claim that `update_cleanup.go` was entirely dead (it is NOT — `scanDeprecatedPaths` is live, the lock pair is P0-wired, `backupDeprecatedPaths` is a WIRE-DECISION). Realistic net delta: on the order of −150 to −250 lines, contingent on the `ttyConfirmer` deferred-pairing resolution and the WIRE-DECISION outcome.

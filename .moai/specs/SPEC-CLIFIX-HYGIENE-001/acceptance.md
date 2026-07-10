# SPEC-CLIFIX-HYGIENE-001 — Acceptance Criteria

## §A Scenarios (Given-When-Then)

1. Given a maintainer opening the update cluster after decomposition, When they inspect the files, Then sync/merge/wizard/settings concerns live in separate files and the characterization suite proves behavior identical to the pre-split tree.
2. Given a user in a non-Korean locale hitting a doctor/migration/clean/web-port error, When the message renders, Then it is English per `error_messages: en`.
3. Given a project path `~/dev/my project (v2)`, When a worktree tmux session spawns, Then the initial command executes correctly with the quoted path, and a commented `# team_mode: cg` line in workflow.yaml does not flip CG detection.

## §D AC Matrix (machine-verifiable)

| AC | REQ | Verification command | Expected outcome |
|---|---|---|---|
| AC-HYG-001-001 | REQ-HYG-001-001 | `go test ./internal/cli/ -run 'UpdateCharacterization' -count=1` && `wc -l internal/cli/update*.go` | Suite PASS unchanged post-split; no update-cluster file > 1,200 lines; update.go no longer 3,181 lines |
| AC-HYG-001-002 | REQ-HYG-001-002 | `for s in buildGLMEnvVars ttyConfirmer; do grep -rn "$s" internal/cli --include='*.go' \| grep -v _test.go; done` (plus per-inventory symbols) | 0 matches per deleted symbol; `go build ./...` PASS |
| AC-HYG-001-003 | REQ-HYG-001-003 | `go test ./internal/cli/ -run 'GLMEnvSetParity' -count=1 -v` && `grep -rn 'ANTHROPIC_' internal/cli --include='*.go' \| grep -v envkeys \| grep -v _test.go` | Inject set == clear set (single SSOT); env-name literals outside envkeys.go reduced to 0 |
| AC-HYG-001-004 | REQ-HYG-001-004 | `grep -rn '1, 3, 5, 10\|\[1,3,5,10\]' internal/cli --include='*.go' \| grep -v defaults \| grep -v _test.go` | 0 inline duplicates; defaults.go carries the constants; hook dispatcher timeout constant ≤ hook budget with a single definition |
| AC-HYG-001-005 | REQ-HYG-001-005 | `rg -l '[가-힣]' internal/cli/doctor.go internal/cli/migration.go internal/cli/clean.go internal/cli/web_port*.go` | 0 files matched (rg used — BSD grep bracket ranges over Hangul are unreliable) |
| AC-HYG-001-006 | REQ-HYG-001-006 | `go test ./internal/cli/ -run 'RuneTruncate' -count=1 -v` | PASS — CJK fixtures at all 4 sites yield `utf8.ValidString(output) == true` at every boundary length |
| AC-HYG-001-007 | REQ-HYG-001-007 | `go test ./internal/cli/ -run 'DeepMergeOldOnlyKeys' -count=1 -v` | PASS — user-only keys survive the 3-way merge (RED on pre-fix code recorded) |
| AC-HYG-001-008 | REQ-HYG-001-008 | `go test ./internal/cli/ -run 'TokenFilePerms' -count=1 -v` | PASS — user.yaml written with mode 0600 when tokens present |
| AC-HYG-001-009 | REQ-HYG-001-009 | `go test ./internal/cli/wizard/ -run 'PATMask\|StepperTotal' -count=1 -v` | PASS — PAT question uses password echo; stepper total equals displayed question count for 6/7/9-question configurations |
| AC-HYG-001-010 | REQ-HYG-001-010 | `go test ./internal/cli/worktree/ -run 'YAMLCommentCG\|TmuxPathQuote' -count=1 -v` | PASS — commented team_mode does not activate CG; spaced/metachar path round-trips through the tmux initial command |

## §C Edge Cases

- Dead-code deletion vs build tags: symbols only referenced under `//go:build windows` (or unix) must be checked per-GOOS before deletion (`GOOS=windows go build ./...`).
- deepMerge3Way with user key whose VALUE type changed in the new template — preservation must not clobber the template's new type for template-owned keys; only old-ONLY keys are preserved.
- Rune truncation at limit==0 and limit smaller than the first rune — return empty string, no panic.
- user.yaml pre-existing with 0644 and tokens already present — update tightens mode on rewrite.
- Wizard in non-interactive mode — PAT masking path must not break `-p`/CI flows (skip prompt, not hang).
- workflow.yaml with `team_mode: cg` nested under an unexpected parent key — yaml.v3 parse reads the documented path only; unknown placements do not activate CG.

## §D.5 Quality Gate / Definition of Done

- All 10 AC rows PASS with verbatim command output cited in progress.md §E.2.
- Characterization suite green on both pre-split (baseline run recorded) and post-split trees.
- `go build ./...` on darwin/linux/windows; `go test ./internal/cli/... -count=1` green; `golangci-lint run` no new findings.
- Dead-code net LOC delta reported (target ≈ −500 lines production code) with the deleted-symbol inventory in progress.md §E.2.

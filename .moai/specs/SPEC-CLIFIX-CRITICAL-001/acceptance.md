# SPEC-CLIFIX-CRITICAL-001 — Acceptance Criteria

## §A Scenarios (Given-When-Then)

1. Given a developer machine with a customized `.claude/settings.local.json` (hooks, outputStyle, permissions blocks), When the user runs `moai glm` then `moai cc`, Then diffing the file before/after shows only the intended env/teammateMode keys changed and no top-level key was removed.
2. Given a team-mode session with an existing tasklist ledger, When three claims are made sequentially, Then the ledger contains the original header + all original tasks + three appended claim lines, in order.
3. Given a repository with both `release` and `release-update` harnesses and a running update in one terminal, When `moai harness remove release` and a second `moai update` are invoked, Then release-update artifacts are intact and the second update exits fast with a lock diagnostic.

## §D AC Matrix (machine-verifiable)

Each AC lists the verification command(s) and expected outcome. Test names are canonical targets for the run phase (exact names may be refined at RED time but must keep the AC-referenced grep-able token).

| AC | REQ | Verification command | Expected outcome |
|---|---|---|---|
| AC-CRIT-001-001 | REQ-CRIT-001-001 | `go test ./internal/cli/ -run 'SettingsLocalPreserve' -count=1 -v` | PASS — fixture with extra top-level keys (hooks/outputStyle/permissions) survives glm+cc round-trip |
| AC-CRIT-001-001b | REQ-CRIT-001-001 | `grep -n 'SettingsLocal' internal/cli/glm.go internal/cli/launcher.go internal/cli/settings.go` | No write-back path (including the `mutateSettingsLocal` seam in settings.go) marshals the closed struct back to disk (struct may remain for read-only convenience) |
| AC-CRIT-001-002 | REQ-CRIT-001-002 | `go test ./internal/cli/ -run 'ClaimTaskAppend' -count=1 -v` | PASS — ledger head bytes unchanged after claim; claim line at tail |
| AC-CRIT-001-002b | REQ-CRIT-001-002 | `grep -n 'O_APPEND' internal/cli/team_spawn.go` | ≥1 match at the ClaimTask open; no `O_RDWR` remains on the claim write path |
| AC-CRIT-001-003 | REQ-CRIT-001-003 | `go test ./internal/cli/ -run 'HarnessMutePreserve' -count=1 -v` | PASS — workflow.yaml fixture with agentic_loop/team keys retains them after mute save |
| AC-CRIT-001-004 | REQ-CRIT-001-004 | `go test ./internal/cli/harness/ -run 'RemoveHarnessBoundary' -count=1 -v` | PASS — removing `release` leaves every `release-update` artifact on disk; same test class covers EditHarness |
| AC-CRIT-001-005 | REQ-CRIT-001-005 | `go test ./internal/cli/ -run 'UpdateLock' -count=1 -v` && `grep -n 'acquireUpdateLock' internal/cli/update.go` | PASS — contended second update fails fast; ≥1 production call site inside runUpdate destructive path |
| AC-CRIT-001-006 | REQ-CRIT-001-006 | `go test ./internal/cli/ -run 'MigrateRollbackPreexisting' -count=1 -v` | PASS — pre-existing dir content restored after induced phase failure; migration-created paths removed |
| AC-CRIT-001-007 | REQ-CRIT-001-007 | `go test ./internal/cli/ -run 'MigrateSymlinkSkip' -count=1 -v` && `grep -n 'Lstat\|isSymlinkEntry' internal/cli/migrate_agency.go` | PASS — out-of-tree symlink target not copied; Lstat-based guard present at the archive walk |
| AC-CRIT-001-008 | REQ-CRIT-001-008 | `go test ./internal/cli/ -run 'TierPromotionHighWater' -count=1 -v` | PASS — second classify run over identical logs appends 0 new lines to tier-promotions.jsonl |
| AC-CRIT-001-009 | REQ-CRIT-001-009 | `go test ./internal/cli/... -count=1` and `git log --format='%h %s' -- internal/cli` | Full suite PASS; each fix commit references its repro test (RED evidence in commit body or preceding commit) |

## §C Edge Cases

- (a) settings.local.json containing unknown future keys and non-object values — must survive verbatim.
- (b) claim on an empty (0-byte) ledger — append still valid, no panic.
- (c) workflow.yaml with comments — Node API must preserve comments where yaml.v3 supports it; at minimum keys/values survive.
- (d) harness names that are themselves prefixes of each other in both directions (`a`, `a-b`, `a-b-c`).
- (e) stale lock file left by a crashed update — lock acquisition must detect staleness (PID/liveness or age policy) rather than deadlock forever.
- (f) rollback when the snapshot itself is partially written — rollback must not destroy the snapshot before restore completes.
- (g) archive walk encountering a symlink chain (symlink → symlink → out-of-tree target) — the Lstat guard must skip at the first link without following the chain.
- (h) legacy tier-promotions.jsonl that already contains historical duplicates — reader must tolerate them; high-water mark derives from latest state.

## §D.5 Quality Gate / Definition of Done

- All 11 AC rows PASS with verbatim command output cited in progress.md §E.2 (verification-claim-integrity §2 attribution).
- `go test ./internal/cli/... -count=1` green; `go test -race ./internal/cli/ -run 'ClaimTask|UpdateLock' -count=1` green.
- `golangci-lint run` introduces no new findings vs pre-SPEC baseline.
- No Go file outside the §B/plan.md inventory modified (except test files and test fixtures).

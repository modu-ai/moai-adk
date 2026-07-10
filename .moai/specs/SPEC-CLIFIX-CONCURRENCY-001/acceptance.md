# SPEC-CLIFIX-CONCURRENCY-001 — Acceptance Criteria

## §A Scenarios (Given-When-Then)

1. Given two Claude Code sessions on the same checkout, When both toggle GLM/CC mode near-simultaneously, Then settings.local.json ends in a consistent state containing every non-mutated key, with no truncated file and no lost hooks/outputStyle blocks.
2. Given a user returning from `moai glm` to `moai cc`, When the session starts, Then no GLM-era auto-compact window (1M) remains configured and statusline/context behavior matches Claude-mode expectations.
3. Given a preference entry promoted from session tier to core tier, When any consumer reads it afterwards, Then only the promoted value is observable, before and after a simulated crash mid-scan.

## §D AC Matrix (machine-verifiable)

| AC | REQ | Verification command | Expected outcome |
|---|---|---|---|
| AC-CONC-001-001 | REQ-CONC-001-001 | `go test -race ./internal/cli/ -run 'SettingsLocalConcurrent' -count=1 -v` | PASS — N-goroutine writer test: no lost keys, valid JSON, no truncation; all 5 write sites verified routed (grep: direct `os.WriteFile` to settings.local.json = 0 outside the seam) |
| AC-CONC-001-002 | REQ-CONC-001-002 | `go test ./internal/cli/ -run 'RemoveGLMEnvComplete' -count=1 -v` && `grep -n 'CLAUDE_CODE_AUTO_COMPACT_WINDOW' internal/cli/launcher.go` | PASS — removal set includes the key; inject set and clear set are the same SSOT list |
| AC-CONC-001-003 | REQ-CONC-001-003 | `go test -race ./internal/cli/ -run 'ClaudeJSONGuard' -count=1 -v` | PASS — interleaved external write is preserved (retry observed) or serialized (lock observed); no lost update |
| AC-CONC-001-004 | REQ-CONC-001-004 | `grep -rn 'func.*[Aa]tomic.*[Ww]rite' internal/cli --include='*.go' \| grep -v _test.go` | Exactly 1 implementation; callers import/use it (0 duplicate helper bodies) |
| AC-CONC-001-005 | REQ-CONC-001-005 | `go test ./internal/cli/preference/ -run 'UpsertTierTransition' -count=1 -v` | PASS — old-tier file no longer contains the entry; cascade returns new value |
| AC-CONC-001-006 | REQ-CONC-001-006 | `go test -race ./internal/cli/preference/ -run 'DecayCrash\|ScanDueRace\|ToggleRace' -count=1 -v` | PASS — crash-injection reload shows no duplicates; racing scans/toggles lose no update |
| AC-CONC-001-007 | REQ-CONC-001-007 | `go test -race ./internal/cli/... ./internal/cli/preference/... -count=1` | PASS — full affected-package suite green under race detector |

## §C Edge Cases

- settings.local.json absent at first write (fresh checkout) — seam creates it atomically with correct perms.
- settings.local.json containing invalid JSON (manual edit) — seam surfaces a diagnostic instead of silently replacing the file with a minimal object.
- ~/.claude.json very large (MCP registrations) — guard must not hold the lock across the full marshal; read-mutate-marshal outside, lock only around final compare+rename.
- flock unavailable (unsupported FS) — mtime-retry fallback engages; bounded retries then fail-open with a stderr warning (never deadlock a session start).
- Preference toggle when the entry exists in both tiers (pre-fix corrupted state) — reconciliation prefers the higher tier and removes the stale duplicate.

## §D.5 Quality Gate / Definition of Done

- All 7 AC rows PASS with verbatim command output cited in progress.md §E.2.
- `go test -race` green over internal/cli + internal/cli/preference.
- Reproduction tests demonstrably FAIL when run against the pre-fix commit (RED evidence recorded).
- `golangci-lint run` introduces no new findings; no new third-party dependency in go.mod.

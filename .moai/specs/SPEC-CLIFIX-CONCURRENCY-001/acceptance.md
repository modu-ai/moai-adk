# SPEC-CLIFIX-CONCURRENCY-001 — Acceptance Criteria

## §A Scenarios (Given-When-Then)

1. Given two Claude Code sessions on the same checkout, When both toggle GLM/CC mode near-simultaneously, Then settings.local.json ends in a consistent state containing every non-mutated key, with no truncated file and no lost hooks/outputStyle blocks.
2. Given a user returning from `moai glm` to `moai cc`, When the session starts, Then no GLM-era auto-compact window (1M) remains configured and statusline/context behavior matches Claude-mode expectations.
3. Given a DecayScan interrupted by a crash between the archival write and the recall write, When the store is next loaded, Then no duplicate entry survives and the recall tier is consistent with the archival tier.

## §D AC Matrix (machine-verifiable)

| AC | REQ | Verification command | Expected outcome |
|---|---|---|---|
| AC-CONC-001-001 | REQ-CONC-001-001 | `go test -race ./internal/cli/ -run 'SettingsLocalConcurrent' -count=1 -v` | PASS — N-goroutine writer test: no lost keys, valid JSON, no truncation; all 5 write sites verified routed (grep: `writeSettingsMap` calls outside settings.go = 0 after re-routing) |
| AC-CONC-001-002 | REQ-CONC-001-002 | `go test ./internal/cli/ -run 'RemoveGLMEnvComplete' -count=1 -v` && `grep -n 'EnvClaudeCodeAutoCompactWindow' internal/cli/launcher.go` | PASS — removeGLMEnv delete set (launcher.go:268-281 region) includes the key |
| AC-CONC-001-003 | REQ-CONC-001-003 | `go test -race ./internal/cli/ -run 'ClaudeJSONGuard' -count=1 -v` | PASS — interleaved external write is preserved (retry observed) or serialized (lock observed); no lost update at glm_tools.go:503-548 and :562-616 |
| AC-CONC-001-004 | REQ-CONC-001-004 | `grep -rn 'func writeFileAtomic\|func atomicWriteJSON\|func writeClaudeJSONAtomic\|func atomicWrite' internal/cli --include='*.go' \| grep -v _test.go` (symbol-inventory assertion, order-agnostic; grep recurses so it observes both `internal/cli` and `internal/cli/preference`) | Exactly 2 matches remain from the 6-site inventory: 1 consolidated helper in `internal/cli` (consolidated from 3 named — writeFileAtomic settings.go:121 / atomicWriteJSON update_noise.go:74 / writeClaudeJSONAtomic glm_tools.go:437 — plus 2 inline: harness_mute.go:250-254 and glm.go:689-704 saveLLMSection) + 1 acknowledged package-internal exception `func atomicWrite` at `internal/cli/preference/filestore.go:473` (option b — separate package, cannot import internal/cli without cycle) |
| AC-CONC-001-005 | REQ-CONC-001-005 | `go test -race ./internal/cli/preference/ -run 'DecayCrash\|ScanDueRace\|ToggleRace' -count=1 -v` | PASS — DecayScan crash-injection reload (decay.go:142-189) shows no duplicates; racing ScanDue (decay.go:317)/MarkScanned (decay.go:340) and toggle (preference/toggle.go) operations lose no update under the race detector |
| AC-CONC-001-006 | REQ-CONC-001-006 | `go test -race ./internal/cli/... ./internal/cli/preference/... -count=1` | PASS — full affected-package suite green under race detector |

## §C Edge Cases

- settings.local.json absent at first write (fresh checkout) — the `mutateSettingsLocal` seam (settings.go:79) creates it atomically with correct perms via `writeFileAtomic` (settings.go:110).
- settings.local.json containing invalid JSON (manual edit) — the seam surfaces a diagnostic instead of silently replacing the file with a minimal object.
- ~/.claude.json very large (MCP registrations) — guard must not hold the lock across the full marshal; read-mutate-marshal outside, lock only around final compare+rename.
- flock unavailable (unsupported FS) — mtime-retry fallback engages; bounded retries then fail-open with a stderr warning (never deadlock a session start).
- DecayScan interrupted between writeArchivalEntry (decay.go:169) and writeRecall (decay.go:185) — on next load, reconciliation deduplicates entries that landed in both archival and recall (or the transactional commit prevents the partial state entirely).

## §D.5 Quality Gate / Definition of Done

- All 6 AC rows PASS with verbatim command output cited in progress.md §E.2.
- `go test -race` green over internal/cli + internal/cli/preference.
- Reproduction tests demonstrably FAIL when run against the pre-fix commit (RED evidence recorded).
- `golangci-lint run` introduces no new findings; no new third-party dependency in go.mod.

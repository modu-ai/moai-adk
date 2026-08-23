# t187 — CI repair verdict (PR #1606, SPEC-CODEX-SESSION-MSG-001)

Card: t187 · Branch: `WT-codex-session-msg` · Worktree: `.claude/worktrees/t187`

## Claim

The two failing CI jobs on PR #1606 — `Test (ubuntu-latest)` and `Race Test` — failed
from **one root cause each**, both consequences of the card adding four MCP tools
(`session_msg_register` / `_list` / `_send` / `_poll`) without updating the two
pinned-count / key-parity guards that track the tool surface. Both are fixed; the
branch is rebased onto current `origin/main` with the CHANGELOG conflict resolved
keeping both sides.

## Evidence

### Root cause 1 — `internal/mcp` pinned tool-count guards

Both jobs, same two assertions:

```
--- FAIL: TestMoaiMCPTools_Count21 (0.00s)
    catalog_test.go:16: catalog declares 25 tools, want 21
--- FAIL: TestMoaiMCPTools_SixWriteCapable (0.00s)
    catalog_test.go:40: write-capable tool count = 9 ([codex_job_cancel codex_task
    glm_job_cancel glm_task goal_arm session_msg_poll session_msg_register
    session_msg_send verify_snapshot]), want 6
```

These are intentional drift guards, not a defect in the card's code: the catalog
grew 21 → 25 by design (`internal/mcp/catalog.go:61-64`). Fix: update both guards
to the new intended counts (25 tools, 9 write-capable) and rename them so the name
no longer states a stale number. `session_msg_list` is correctly **read-only** — it
enumerates registered peers without touching the store — while register/send/poll
write an agent record, append a message, and claim an inbox respectively; that
asymmetry is now stated in the test comment.

### Root cause 2 — `internal/web` i18n dictionary missing the new tools

Two tests, 8 keys each:

```
--- FAIL: TestDataI18nKeysSubsetOfDictionary
    i18n_test.go:271: data-i18n key "f.mcp.tools.session_msg_register.enabled.title"
    in the rendered page is absent from the dictionary (R6: would render blank/untranslated)
    ... (8 keys total)
--- FAIL: TestI18nKeySetParity
    schema_label_test.go:106: i18n.js missing key "f.mcp.tools.session_msg_register.enabled.title"
    in all 4 locales (schema field "mcp.tools.session_msg_register.enabled")
    ... (8 keys total)
```

The web console renders one toggle per catalog tool, so four new tools mean eight
new dictionary keys (`.title` + `.desc`) in **all four locales**. Fix: 32 entries
added to `internal/web/assets/i18n.js` (8 × en/ko/ja/zh), placed after
`audit_multi` to match catalog order. No template mirror exists for this file
(`find . -name i18n.js` returns the single path), so Template-First does not apply.

### Verification, this tree, after both fixes

```
$ go test ./internal/mcp/...
ok  	github.com/modu-ai/moai-adk/internal/mcp	0.516s

$ go test ./internal/web/...
ok  	github.com/modu-ai/moai-adk/internal/web	3.238s

$ go test ./internal/cli -run 'TestMoaiMCP' -timeout 600s
--- PASS: TestMoaiMCPServer_RegistrationMatchesCatalog (0.00s)
--- PASS: TestMoaiMCPServer_ToolsListDeclaresSchema (0.00s)
    (+ 6 further TestMoaiMCPServer_* — all PASS)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.703s

$ go test -race ./internal/mcp/... ./internal/web/... -timeout 600s
ok  	github.com/modu-ai/moai-adk/internal/mcp	1.515s
ok  	github.com/modu-ai/moai-adk/internal/web	22.591s

$ go vet ./internal/web/... ./internal/mcp/... ./internal/cli/...
(no output — clean)

$ grep -c 'f.mcp.tools.session_msg' internal/web/assets/i18n.js
32
```

### Rebase

`origin/main` at `28bde4022`. Ten commits replayed; one conflict, in `CHANGELOG.md`
`[Unreleased]`. Both sides kept: main's accumulated `### Summary` / `### Added` /
`### Changed` / `### Fixed` sections stay intact, and the card's session-messaging
bullet is appended to the **existing** `### Added` section rather than opening a
second one. Post-rebase conflict-marker count is 0.

## Baseline-attribution

Every figure above was measured in this run, in this tree
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t187`), after the rebase onto
`28bde4022` and after both fixes. The failing-side output is quoted from the CI
jobs themselves (run `32638957965`, jobs `97192976870` and `97192976830`) read via
`gh run view --job <id> --log`, not from a local pre-fix reconstruction — the local
pre-fix reproduction (`go test ./internal/mcp/...`, `./internal/web/...`) matched
those assertions exactly before the fix was applied.

## Gaps

- **The full suite was not run locally**, per the repo's standing rule; the
  full-package verdict is CI's on the pushed head. What ran locally is the two
  packages that failed plus the `internal/cli` guard that pairs with the catalog.
- **`Integration Tests` and `Test (${{ matrix.os }})` were `skipping` on the failing
  run**, so no darwin/windows evidence exists for these changes yet — that arrives
  with the re-run.
- The i18n strings for ko/ja/zh are native-idiom translations written for this
  change; they have not been reviewed by a second reader.
- No test asserts the *count* of MCP-tool i18n keys, so a future tool added without
  its keys is still caught by parity (the tests above), not by a count guard — that
  is the existing design, unchanged here.

## Residual-risk

- The two count guards are now pinned to 25/9. Any further tool added to
  `catalog.go` will fail them again by design; that is the intent, but it means the
  guard's maintenance burden grows with the surface.
- CHANGELOG conflict resolution is a judgment: the card's bullet was placed at the
  end of main's `### Added`. If a later lane's rebase lands a different ordering the
  entry may move, which is cosmetic but will show as a conflict again.
- Reviewer note carried over from the card itself and unchanged by this repair: the
  new tools only appear after the MCP server restarts, and codex-cli 0.147 coerces
  the `session_msg_send` `data` argument leniently.

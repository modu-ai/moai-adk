# t462 — Positive control (vacuous-green guard, spec REQ-CEM-005 / AC-CEM-006)

Tree SHA: `bd7c58201` (branch `WT-codex-e2e`). Executed: 2026-09-03, run phase M3 — BEFORE any
zero-count or zero-failure verdict was recorded (see `execution-log.md` for the runs that follow).

## Binary pinning (plan M3 step 1)

Control binary built from THIS pinned tree, not the installed ancestor:

```
$ go build -o /tmp/moai-t462 ./cmd/moai        (exit 0, in .claude/worktrees/t462 @ bd7c58201)
```

The banner reports `v3.1.3 none built unknown` — the un-stamped LDFLAGS compile default for a
bare `go build` (Makefile `LDFLAGS` injects commit/date only on `make build`). Provenance is the
compile just executed against this tree; the installed `/Users/goos/go/bin/moai` (v3.2.0-rc.0,
different provenance) was NOT used (AP-6).

## Primary variant — stray top-level key in `.codex/hooks.json` (plan M3 step 2, hooks-first)

The code validates hooks.json against a whitelist (`codexadapter/config.go` acceptedTopLevel =
`{description, hooks}` exactly — the set Codex itself enforced when it rejected `version`);
`doctor_codex.go` routes `ValidateConfig` violations into a "Codex Wiring" WARN finding.

Fixture (ISOLATED, under `/tmp`, not `~/.codex`):

```
/tmp/t462-pc-fixture/.codex/hooks.json — canonical shape with ONE injected stray top-level key:
  { "description": "t462 positive control fixture", "version": 2,
    "hooks": { "Stop": [ {"hooks": [{"type": "command", "command": "echo hi", "timeout": 30}]} ] } }
```

Detection run (CODEX_HOME command-scoped to a throwaway — REQ-CEM-004 / AP-4):

```
$ cd /tmp/t462-pc-fixture && CODEX_HOME=/tmp/t462-codex-home timeout 60 /tmp/moai-t462 doctor --verbose
  → exit 0 (advisory check — the DETECTION is the finding, not a non-zero exit)
```

Observed DETECTION (verbatim fragments from the captured output, `/tmp/t462-pc-doctor-v2.out`):

```
warn    Codex Wiring          hooks.json fails the key whitelist — Codex would silently ignore the file (+1 more, see --verbose)
Codex Wiring: hooks.json fails the key whitelist (unaccepted top-level key "version"; ...
```

Violation census: `grep -o 'unaccepted [a-z-]* key "[a-z]*"' | sort | uniq -c` → exactly
**1 × `unaccepted top-level key "version"`** — the injected break and nothing else. The check
under test DETECTS a deliberately broken wiring point. **CONTROL PASS.**

## Secondary variant — missing `[mcp_servers.moai]` table (config.toml class)

The same fixture has wiring active (hooks.json present) but no `.codex/config.toml`; the same
doctor run reports:

```
config.toml missing (wiring active but the MCP registration is gone)
```

The missing-table/missing-registration class is likewise DETECTED. **CONTROL PASS.**

## Grep positive control for prong B (plan M3 step 4 / AC-CEM-008)

Method-control before quoting any zero (same file, same grep):

```
$ grep -c doctor e2e/cli/tux3_journeys.sh          → 7     (method sees present tokens)
$ grep -ric codex e2e/cli/tux3_journeys.sh         → 0     (now quotable)
$ grep -ric codex e2e/                             → 0     (whole e2e tree; one file total)
$ find e2e -type f                                 → e2e/cli/tux3_journeys.sh (the only e2e file)
```

**CONTROL PASS** — the grep method demonstrably matches known-present tokens on the same file
before the codex 0 is cited.

## Isolation receipts (REQ-CEM-004)

- `CODEX_HOME=/tmp/t462-codex-home` on every control invocation (throwaway, mkdir'd fresh).
- Real `~/.codex` untouched: before `shasum ~/.codex/config.toml` =
  `ad8c8593a5d89937b9786f1b706384a532361120`; re-check after all integration-style checks is
  recorded in `execution-log.md` §1.

# t508 — RED baseline (pre-plan)

card: t508 · branch `WT-codex-enabled-guard` · base `0b1e27877` (= origin/develop)

## Claim

1. codex refuses to start when a `[[skills.config]]` entry declares no `enabled` key — a hard
   startup error (rc=1), not a partial degradation.
2. `moai doctor`'s Codex Wiring check is SILENT on that shape: it reports `ok`.

## Evidence

### Axis 1 — codex hard error (re-measured in this run, not carried over)

Isolated lab under `/tmp/t508lab` (the real `~/.codex` was NOT read or written — CODEX_HOME
pinned per invocation). codex-cli 0.153.4.

`good/config.toml`:

```toml
[[skills.config]]
path = "/tmp/t508lab/skills/SKILL.md"
enabled = true
```

`bad/config.toml` — identical minus the `enabled` line.

```
$ CODEX_HOME=/tmp/t508lab/good codex mcp list
No MCP servers configured yet. Try `codex mcp add my-tool -- my-command`.
GOOD rc=0

$ CODEX_HOME=/tmp/t508lab/bad codex mcp list
Error: failed to load bootstrap configuration

Caused by:
    missing field `enabled`
    in `skills.config`

BAD rc=1
```

The `good` run is the control: it isolates the missing key as the cause, rather than the lab
config being unusable for some other reason.

### Axis 2 — doctor silence (mutant, RED)

`internal/cli/doctor_codex_enabled_test.go` — one mutant + one control.

```
$ go test ./internal/cli/ -run 'TestCheckCodexWiring_(MissingEnabledKeyIsReportedFatal|DeclaredEnabledKeyStaysQuiet)' -count=1 -v
=== RUN   TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal
    doctor_codex_enabled_test.go:37: status = ok, want CheckFail — codex cannot start on this config: {Name:Codex Wiring Status:ok Message:wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical) Detail:}
    doctor_codex_enabled_test.go:41: finding never names the `enabled` key: ...
--- FAIL: TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal (0.00s)
=== RUN   TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet
--- PASS: TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet (0.00s)
FAIL
```

The control PASSING is what makes the mutant's failure informative: the guard is not simply
failing every config.

## Baseline-attribution

Tree: `.claude/worktrees/t508` @ `0b1e27877`, working tree carrying only the new test file.
Both commands run in this session, in this tree, against this codex binary
(`/Users/goos/.local/bin/codex`, `codex-cli 0.153.4`).

## Structural findings (read, not inferred)

- `internal/codexwiring/skills.go:32` — `SkillEnabledUnspecified` is the FIRST enum value
  (`iota`), documented as "an entry declaring no `enabled` key … It asserts nothing." The parser
  models the fatal shape as a normal tri-state reading.
- `internal/cli/doctor_codex.go:654` `codexStaleSkillFinding` fires only on a MISSING path or an
  unresolvable path SHAPE (`if missing == 0 && unresolvedShape == 0 { return …, false }`,
  line ~732). A live path with no `enabled` key satisfies neither, so nothing is emitted.
- `internal/cli/doctor_codex.go:114` `codexFinding{summary, detail}` carries no severity field,
  and every problem lands on `uikit.CheckWarn` (line ~286). There is currently no ERROR grade on
  this check. `uikit.CheckFail` exists (`internal/cli/uikit/types.go:17`).

## Gaps (explicitly NOT observed)

- Whether the `[[skills.config]]` requirement is version-specific to codex-cli 0.153.4 — only
  this one version was measured.
- Whether a `path`-less entry (no `path` key either) triggers the same or a different codex
  error — not measured.
- The real `~/.codex` 49-entry population was NOT re-counted here (t504 lab, untouched).
- Whether `enabled` written as a quoted string (`"true"`) satisfies codex — the parser accepts
  it, codex was not asked.

## Residual risk

The mutant asserts `CheckFail`, which presumes the fix raises severity rather than widening the
existing warn-level finding. That is a design decision for plan-phase; if the plan chooses a
different surface the guard's assertion changes with it.

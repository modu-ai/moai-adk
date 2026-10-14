# Acceptance Criteria — SPEC-STATUSLINE-OFFICIAL-USAGE-001

## §A. Verification Philosophy

Every AC is judged on observed command output on the tree under test (verification-claim
integrity §2: the command plus its verbatim output, in this run). Removal criteria carry a
two-cell pair: a **RED-now** observation taken on the pre-implementation tree and a
**green-path** statement naming which milestone flips it and what passing output becomes
(verification-completeness §2). The preservation criterion (AC-SOU-005) carries a measured
baseline instead of a RED cell — its assertion is that today's behavior survives.

Baseline tree for every RED-now cell: `1ced8cab9435b655934529fdff2d3648441003c1`
(worktree `t1732`, branch `WT-10-14-class`, measured 2026-10-15).

## §B. AC ID Convention

`AC-SOU-NNN` (SPEC-STATUSLINE-OFFICIAL-USAGE-001), numbered in the order judged.
Severity MUST / SHOULD per the matrix; release-blocking cells are marked and carry all
four RED-now elements (command, verbatim stdout, exit code, tree SHA) either inline or by
citing the §E evidence ledger row.

## §C. Given-When-Then Template

```
Given <stdin payload shape / on-disk snapshot state>,
when <the statusline build+render executes>,
then <observable segment output / call-surface outcome>.
```

## §D. Acceptance Criteria Matrix

### AC-SOU-001 (MUST, release-blocking; maps REQ-SOU-001) — GLM-form input performs zero network and zero keychain calls

**Given** a stdin payload without `rate_limits` (GLM-form), **when** a full statusline
build+render executes, **then** the process issues no HTTP request to `api.anthropic.com`
and reads no keychain service or Claude credentials file — proven by a behavioral test
(builder-level, snapshot fixtures, temp state dir; the removed surfaces do not exist to be
reached) whose failure output names any observed call surface.

- **RED-now (E4)**: the gate that fires the collector under exactly this input exists
  today — command `grep -n "usageProvider" internal/statusline/builder.go`, stdout 7 lines
  including `406:	if b.usageProvider != nil && (input == nil || input.RateLimits == nil) {`,
  exit 0 (§E ledger row E4). Red because the network collector is reachable from the
  no-stdin path on this tree.
- **Green path**: M1 removes the gate and the collector → E4's command exits 1; the new
  behavioral test (M4) passes. Both observed at completion.

### AC-SOU-002 (MUST, release-blocking; maps REQ-SOU-001, REQ-SOU-005) — structural removal guard: the three paths are gone and cannot return silently

**Given** the production (non-test) Go files under `internal/statusline/`, **when** the
guard test (plan.md D6b) runs, **then** no file matches the removed-path marker pattern
(`api.anthropic.com`, `oauth-2025-04-20`, `find-generic-password`,
`Claude Code-credentials`, `haiku-4-5`), and the guard test itself fails the suite naming
any offender when one reappears.

- **RED-now (E1)**: command
  `grep -rln --include='*.go' --exclude='*_test.go' -E "api\.anthropic\.com|oauth-2025-04-20|find-generic-password|Claude Code-credentials|haiku-4-5" internal/statusline`,
  stdout `internal/statusline/usage.go`, exit 0. Red because the file carrying all three
  removed paths exists on the baseline tree.
- **Green path**: M1 deletes `usage.go` → the same command prints nothing, exit 1; M4 adds
  the guard test encoding the same predicate permanently, passing.

### AC-SOU-003 (MUST, release-blocking; maps REQ-SOU-004) — no snapshot renders `n/a`, never a 0% bar

**Given** GLM-form stdin and no usable context-usage record carrying a window (or a
record whose winning window has rolled), **when** the 5H/7D segments render, **then** each
segment shows `n/a`.

- **RED-now (E2)**: command `grep -rn "n/a" internal/statusline/renderer.go`, stdout
  empty, exit 1. Red because the renderer has no `n/a` path today — the no-data branch
  defaults the bar to 0% (`renderer.go:506, 521`).
- **Green path**: M3 lands the fallback branch → the command exits 0 listing the `n/a`
  sites; the M4 behavioral test (GLM-form + empty state dir) asserts the rendered
  segments contain `n/a` and not `5H: 0%`.

### AC-SOU-004 (MUST; maps REQ-SOU-003, REQ-SOU-006) — a fresh snapshot's values render with the recorded time

**Given** GLM-form stdin and a schema-v3 context-usage record seeded via
`SessionTelemetryRecord` carrying `five_hour`/`seven_day` windows and `captured_at`,
**when** the 5H/7D segments render, **then** each window renders its aggregated
`used_percentage` (freshest-capture rule), its reset time via the existing formatters,
and the source record's `captured_at` as an age marker; the `StatusData` fallback field
carries `QuotaReading.CapturedAt` from the winning record.

- **RED-now (E5)**: command `grep -rn "AggregateQuotaDirs" internal/statusline/builder.go`,
  stdout empty, exit 1. Red because the builder never reads the aggregator on this tree —
  the fallback does not exist.
- **Green path**: M2 wires it (command exits 0); M4's fixture test passes asserting
  percentage + captured-at in the rendered output.

### AC-SOU-005 (MUST, release-blocking; maps REQ-SOU-002) — Claude-form input: no regression

**Given** a stdin payload with `rate_limits`, **when** the statusline renders, **then**
the 5H/7D bars are byte-identical to the baseline tree's output for the same stdin, and
the builder does not read the snapshot fallback.

- **Baseline (E3, not RED — preservation)**: `go test ./internal/statusline/ -count=1`
  → `ok  	github.com/modu-ai/moai-adk/internal/statusline	46.767s`, exit 0, on the
  baseline tree. The existing renderer/builder tests covering `data.RateLimits` rendering
  are part of this passing suite.
- **Green path**: the same command passes at completion; additionally (a) the test
  functions exercising `RateLimits` rendering are unmodified (verified by reviewing the
  completion diff — only collector-referencing code may be touched in those files), and
  (b) the M4 Claude-form builder test asserts the fallback field is nil when
  `rate_limits` are present.

### AC-SOU-006 (MUST; maps REQ-SOU-005) — retired identifiers deleted

**Given** the package after M1, **when** the symbol search runs, **then**
`UsageProvider`, `NewUsageCollector`, `usageCollector`, `UsageData`, `UsageResult`, and
`usageCacheFile` have zero occurrences in production files, and the corresponding test
fixtures (`mockUsageProvider`, `realisticUsage`) are gone from `builder_test.go`.

- **Verification**: `go build ./...` (compile fails on any dangling reference) plus a
  grep of the retired names returning empty for non-test files. `go vet` and
  `golangci-lint run` clean for the package.

### AC-SOU-007 (MUST; maps REQ-SOU-007) — Template-First cycle green

**Given** the template tree after M5, **when** the Template-First build cycle runs, **then**
`make build` succeeds and `git status --short` after the build shows no tracked-file
change (no runtime artifact leaked into the embed tree); where no template asset was
modified, the cycle is verified green without modification.

- **Verification**: `make build` exit 0; `git status --short` empty of tracked changes;
  template grep (plan.md M5) documents the no-reference finding either way.

### AC-SOU-008 (SHOULD) — coverage of the new fallback code

**Given** the package after M4, **when** coverage is measured, **then** the new fallback
branch (builder wiring + renderer branch) is covered by the M4 tests and the package
meets the project quality gate's coverage threshold.

- **Verification**: `go test -cover ./internal/statusline/` at completion; the figure is
  reported in the run-phase self-verification (E3) with command and verbatim output.

## §E. Evidence ledger (RED-now observations, baseline tree)

| Row | Command (single invocation, verbatim) | Verbatim stdout | Exit | Tree SHA |
|-----|----------------------------------------|-----------------|------|----------|
| E1 | `grep -rln --include='*.go' --exclude='*_test.go' -E "api\.anthropic\.com\|oauth-2025-04-20\|find-generic-password\|Claude Code-credentials\|haiku-4-5" internal/statusline` | `internal/statusline/usage.go` | 0 | `1ced8cab9435b655934529fdff2d3648441003c1` |
| E2 | `grep -rn "n/a" internal/statusline/renderer.go` | (no output) | 1 | `1ced8cab9435b655934529fdff2d3648441003c1` |
| E3 | `go test ./internal/statusline/ -count=1` | `ok  	github.com/modu-ai/moai-adk/internal/statusline	46.767s` | 0 | `1ced8cab9435b655934529fdff2d3648441003c1` |
| E4 | `grep -n "usageProvider" internal/statusline/builder.go` | 7 lines: 23, 92, 115, 123, 135, 406, 408 (406 shown under AC-SOU-001) | 0 | `1ced8cab9435b655934529fdff2d3648441003c1` |
| E5 | `grep -rn "AggregateQuotaDirs" internal/statusline/builder.go` | (no output) | 1 | `1ced8cab9435b655934529fdff2d3648441003c1` |

Note on E1's pattern spelling: the ledger cell escapes `|` as `\|` because the cell is
table-delimited; the executed command used the shell alternation
`"api\.anthropic\.com|oauth-2025-04-20|find-generic-password|Claude Code-credentials|haiku-4-5"`.
Re-run the executed spelling, not the table-escaped one.

## §F. Residual risks (named, not silently accepted)

- The behavioral zero-call test (AC-SOU-001) proves the surfaces are unreachable from the
  builder+render path in-process; it cannot observe an out-of-process spawn it does not
  know about. The structural guard (AC-SOU-002) covers the reintroduction axis by source
  scan instead.
- The snapshot fallback trusts the local records the tool itself wrote; a corrupted or
  forged record renders its numbers. This is the same trust the context bar already
  places in the snapshot (`raw_pct`) — no new surface is introduced.
- AC-SOU-005's byte-identical clause is judged through the test suite plus a diff review,
  not a golden-output snapshot; a rendering change that keeps all existing tests passing
  while altering output for inputs no test covers would slip. The M4 Claude-form test is
  written to pin one full rendered line for exactly this reason.

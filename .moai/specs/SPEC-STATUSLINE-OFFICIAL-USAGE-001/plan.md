# Plan — SPEC-STATUSLINE-OFFICIAL-USAGE-001

## §A. Context

Card t1732 (operator decision, 2026-10-14, Class C, Template-First): the statusline usage
lookup must use official input only. Remove the undocumented Haiku probe, the OAuth usage
endpoint, and the credential reads from `internal/statusline/usage.go`; feed the 5H/7D
bars from stdin `rate_limits`, falling back to the local context-usage snapshot, else `n/a`.

Tier M. This worktree (`.moai/worktrees/t1732`, branch `WT-10-14-class`) is the card tree;
iteration-1 baseline SHA `1ced8cab9435b655934529fdff2d3648441003c1`; iteration-2 revisions
re-measured on `0d057bd1e` (the two commits between differ only by files under
`.moai/specs/`, so every code-tree observation below holds on both).

### §A.1 Investigation confidence

Every claim below was verified by reading the named code on the baseline tree, not inferred:

| Fact | Source | Confidence |
|------|--------|------------|
| The three removed paths and their exact shapes | `usage.go:22-41` (URL/header/model consts), `fetchUsageFromHeaders*` (Haiku POST), `fetchUsageFromOAuthAPI` (OAuth GET), `readOAuthToken` + `readTokenFromKeychain` (keychain + credential files) | High — read in full |
| Builder gate location and condition | `builder.go:402-406` (`usageProvider != nil && (input == nil || input.RateLimits == nil)`), auto-create block `builder.go:113-126` | High |
| Renderer prefers stdin over collector | `renderer.go:504-528` (5H: `formatResetTimeRelative`; 7D: `formatResetTimeAbsolute`; no-data renders a 0% bar) | High |
| Snapshot schema carries the fallback data | `context_usage.go:97-139` — `SessionTelemetryRecord.FiveHour/SevenDay *QuotaWindowRecord` (`used_percentage`, `resets_at` Unix, `exhausted_at`), `CapturedAt` RFC 3339, schema v3 | High |
| Account-level aggregator exists, is pure, is test-covered | `quota.go` (`AggregateQuota`, `QuotaAggregate`, `QuotaReading{State, UsedPercentage, ResetsAt, CapturedAt, ExhaustedAt}`, fresh/reset/unknown states), `quota_dirs.go` (`QuotaStateDirs`, `AggregateQuotaDirs`) | High |
| Aggregator has no production caller yet | Repo-wide grep for `AggregateQuota` — only `quota.go`, `quota_dirs.go`, and their tests | High |
| Collector has no consumer outside the package | Repo-wide grep for `CollectUsage|UsageProvider|NewUsageCollector` excluding `internal/statusline` — zero hits | High |
| Test inventory | `usage_test.go` 29 test funcs, **1,189 lines** (`wc -l`); `builder_test.go` **1,878 lines** with **26 references** to `mockUsageProvider`/`realisticUsage` (first at `:427`, type def `:536-543`, helper `:575-578`, call sites through `:1218`); quota aggregator tests in `quota_test.go` (3), `context_usage_quota_test.go` (5), `quota_dirs_test.go` (7) | Measured |
| Worktree blindness (audit B1) | `QuotaStateDirs` returns only the caller's own state dir when `<root>/.git` is a regular file (`quota_dirs.go:43` `!info.IsDir()`), so a worktree-rooted render never sees the primary checkout's records. Live on this card tree: `.git` here is a regular file — `gitdir: /Users/goos/MoAI/moai-adk-go/.git/worktrees/t1732` — from which the primary root is mechanically derivable | Measured |
| Bounded-read residual (audit B2) | The aggregator's bounds are metadata-only: `maxDirs` caps worktree-metadata enumeration and the ModTime pre-filter caps parsed-record count, but `ReadSessionTelemetry` is an unbounded `os.ReadFile` (no size cap, no context, a FIFO would block) | High — read |
| No template/docs asset documents the removed paths | grep over `internal/template/templates/` for haiku/oauth/credentials — only unrelated llm.yaml model-slot comments and OAuth how-to docs; `README.md`/docs-site (en) — zero hits | High |
| Baseline suite green | `go test ./internal/statusline/ -count=1` → `ok ... 46.767s` on the baseline tree | Measured |

### §A.2 Build/test commands (canonical)

```bash
go build ./...
go test ./internal/statusline/ -count=1
go test ./internal/statusline/ -count=1 -race   # race variant (race_enabled tag drives -race elsewhere)
make build                                       # Template-First re-embed (only if template assets change)
golangci-lint run --timeout=2m internal/statusline/...
```

Per repo-local policy (§8 of the lane protocol): do NOT run `go test ./...` in the card
tree; the affected-package suite plus CI is the verdict surface.

## §B. Design decisions

**D1 — Fallback data source: the existing quota aggregator (no new reader).**
`AggregateQuotaDirs(stateDirs, now, maxAge)` already implements the card's fallback
exactly: it reads context-usage records across the given state directories, picks per
window the freshest capture, applies freshness / max-age / clock-skew / reset-rollover
rules, and returns `QuotaReading{State, UsedPercentage, ResetsAt, CapturedAt,
ExhaustedAt}` — `CapturedAt` is the "recorded time" the card requires displayed. Writing
a new reader would duplicate verified machinery; reading only the current session's record
would make the fallback dead under GLM (a GLM session never writes windows). Wiring the
aggregator into the render path also gives the currently caller-less aggregator its first
production consumer.

**D7 — Scan roots resolved builder-side: the worktree blind spot (audit B1; relative pointers per board d-20261014T235405Z-8b55).**
`QuotaStateDirs` cannot be handed a linked-worktree root as-is: with `.git` a regular file
it returns only that worktree's own state dir (`quota_dirs.go:43`), so a card-tree session
would never see the primary checkout's records. The fix is builder-side, keeping
`quota_dirs.go` on the PRESERVE list: a new package helper (suggested
`quotaScanStateDirs(projDir string, maxDirs int) []string` in a new
`usage_fallback.go`) that (a) if `projDir/.git` is a regular file, reads its first
`gitdir:` line and derives the primary checkout root by stripping the
`/.git/worktrees/<name>` suffix — resolving a RELATIVE pointer (git accepts them as valid
worktree anchors; the package's own mirrored reader at `quota_dirs.go:77` is abs-only)
via `filepath.Clean` + `filepath.Join` with `projDir` before the suffix strip — then
(b) calls `QuotaStateDirs(primaryRoot, maxDirs)` — which returns the primary state dir
plus every linked worktree's, the current one included. When `projDir/.git` is a
directory (session anchored at the primary checkout) the helper is the identity over
today's single call. Degradation is fail-open: an unreadable, malformed, or non-worktree
`gitdir` line yields today's behavior (`QuotaStateDirs(projDir, ...)`), whose worst case
is `n/a` — never wrong data. The card's "the last values a Claude session on the same
machine received" intent is preserved: the quota is account-level, and the primary
checkout is where Claude sessions run.

**D8 — Context-guarded read, not an absolute latency promise (audit B2; mid-read pin per board d-20261014T235405Z-8b55).**
The fallback read runs inside the builder's collect phase, which already carries the
render context. The helper guards the read with a `select` on `ctx.Done()`: an
over-deadline read abandons the fallback (field stays nil → segments render `n/a`)
instead of stalling the render. The guard's check must sit on the read path itself — not
only an entry check before launching — because an entry-check-only form admits a mutant
that passes a pre-cancelled-context test while never honoring mid-read cancellation; M4's
blocked/slow fixture exercises exactly the mid-read path (AC-SOU-010).
`quota.go`/`quota_dirs.go` stay PRESERVE — no size cap or context is added inside them;
the named residual (an unbounded `os.ReadFile` inside the abandoned goroutine) is
accepted and recorded in REQ-SOU-006 because the directory holds only tool-written small
JSON records and is not attacker-controlled. The requirement is deliberately scoped to
bounded best-effort, matching the mechanism.

**D2 — Retire `UsageResult`/`UsageData`; introduce a snapshot-typed field.**
The network-era types (`types.go:429-441`, `StatusData.Usage` at `types.go:296`) carry no
captured-at time and no fresh/reset distinction, and reusing them would keep their
semantics alive. They are deleted; the builder populates a new `StatusData` field carrying
the `QuotaAggregate` (exact name at implementation; `UsageSnapshot` suggested) and the
renderer gains the fallback branch. The `data.RateLimits` branch is untouched.

**D3 — `n/a` semantics.** Today the no-data path renders a `0%` bar (renderer defaults
`pct` to 0), which reads as a real measurement. The card specifies `n/a`: aggregate state
`unknown` (no usable record) and `reset` (window rolled; never a stale high) both render
`n/a`. State `fresh` renders the bar with `int(UsedPercentage)`, the existing reset-time
formatters, and the `CapturedAt` age marker. The behavior change is confined to the
no-stdin path — Claude-form input never reaches it (REQ-SOU-002), so the card's no-regression
AC is unaffected.

**D4 — Max-age source.** The aggregator's `maxAge` parameter is fed from the quota-gate
config value already validated by `internal/config/loader_quota_gate.go` (floor: twice
`QuotaHeartbeatInterval`, default constant `config.QuotaHeartbeatInterval`-derived), not a
new statusline-specific constant. If the loader exposes no reusable accessor, a package
constant set to the same validated default is the fallback — decided at M2 with the loader
code open; either way the value is named, not inlined.

**D5 — Stale `~/.moai/cache/usage.json` is left alone.** After removal nothing reads or
writes it; the directory is shared with the update checker. No migration, no cleanup.

**D6 — Zero-call proof form.** Two layers, both test-enforced: (a) behavioral — builder
tests with GLM-form stdin complete against snapshot fixtures with no network/keychain
surface available; (b) structural — a guard test scans production (non-test) files under
`internal/statusline/` for the removed-path markers (`api.anthropic.com`,
`oauth-2025-04-20`, `find-generic-password`, `Claude Code-credentials`, `haiku-4-5`,
`credentials\.json` — the last covering both credential files, audit B3's mutant
closure: a credential read reintroduced under a renamed symbol still carries the path
string) and fails the suite naming any offender. Compile-time absence alone is not a
test; the guard makes the removal permanent (AC-SOU-002).

## §C. Milestones

**M1 — Remove the collector (usage.go + builder + types).**
Delete `usage.go` in full. Remove from `builder.go`: the `usageProvider` field, the
`Options.UsageProvider` option, the auto-create block (113-126), the collect goroutine and
gate (378, 402-413), and the `data.Usage` assignment (425-426). Remove from `types.go`:
`UsageData`, `UsageResult`, `StatusData.Usage`. Update the stale comments that reference
the collector (builder.go:402-405 rationale, `internal/cli/statusline.go:17-19` and the
OAuth-429 mention in `runStatusline`, renderer.go:504/519 "MoAI API call" comments). The
800 ms render budget in `internal/cli/statusline.go` stays (other collectors still run
under it).

**M2 — Wire the snapshot fallback (builder).**
When the parsed stdin carries no `rate_limits`, resolve `projDir` via the existing
`resolveStateAnchor(input)` (already used for the snapshot write at builder.go:201),
derive the scan state dirs through the D7 helper (`quotaScanStateDirs`), and populate the
new `StatusData` field with `AggregateQuotaDirs(dirs, now, maxAge)` guarded by the D8
context select. When `projDir` is empty or the rate_limits are present, the field stays
nil — no read. Worktree case included: a projDir whose `.git` is a regular file yields
the primary checkout's state dir plus all linked worktrees' (D7).

**M3 — Renderer fallback + `n/a`.**
In the 5H/7D segment rendering, replace the deleted `data.Usage` branch with the snapshot
branch: `RateLimits` present → existing rendering untouched; else snapshot `fresh` → bar
with percentage, existing reset-time formatter (Unix-seconds input already supported by
`parseResetTime`), plus the captured-at age marker (existing relative-time formatting
style; exact string at implementation); `reset`/`unknown`/nil → `n/a`.

**M4 — Tests.**
Delete `usage_test.go` (1,189 lines). Rework `builder_test.go` (1,878 lines): remove the
26 `mockUsageProvider`/`realisticUsage` references and the `UsageProvider` option uses;
add snapshot-fixture builder tests (temp state dir with seeded context-usage records via
the existing `SessionTelemetryRecord` schema) covering: GLM-form input + no snapshot →
`n/a`; GLM-form + fresh record → snapshot values + captured-at; Claude-form input → bars
from stdin, fallback not read. Add (audit B1) the worktree-fixture test: a temp
worktree-shaped tree whose `.git` is a regular file (`gitdir: <primary>/.git/worktrees/x`)
pointing at a temp primary that holds the ONLY window-carrying record — the render must
show the primary record's values, not `n/a`; plus the degradation cases (malformed
`gitdir` line, missing `.git`). Add (sou-debt-2 absorbed, board
d-20261014T235405Z-8b55) the **relative-pointer fixture case**: the same worktree-shaped
tree whose `gitdir:` line is RELATIVE to the worktree root (resolved per D7 via
`filepath.Clean` + `filepath.Join`) must also surface the primary record's values —
AC-SOU-009 covers both pointer shapes. Add (sou-debt-1 absorbed; audit B2) the
**mid-read abandonment test**: a blocked/slow-read fixture (the fallback read has
STARTED) with the render context cancelled after the read begins — the fallback abandons
mid-read, the build returns promptly, and segments render `n/a`; the pre-cancelled
context remains covered as the entry case but is not the sole pin (a pre-cancelled-only
test admits the entry-check mutant). Add the guard test (D6b) with the extended marker
set (audit B3: the credential files included). Keep the quota aggregator tests untouched.
Verify race variant for the new code paths.

**M5 — Template/docs sweep + Template-First cycle.**
Grep `internal/template/templates/` for references to the removed collection (expected:
none; `statusline.yaml`'s `usage_5h`/`usage_7d` comments document segment toggles only).
If a comment needs updating, edit the template source and run `make build` so the embed
re-compiles green; otherwise verify the cycle green without modification. CHANGELOG entry
is sync-phase work (manager-docs), not M5.

## §D. File inventory (touch list)

| File | Change |
|------|--------|
| `internal/statusline/usage.go` | Deleted in full |
| `internal/statusline/usage_test.go` | Deleted in full |
| `internal/statusline/builder.go` | Provider field/option/auto-create/gate removed; fallback wired; comments updated |
| `internal/statusline/types.go` | `UsageData`/`UsageResult`/`StatusData.Usage` removed; snapshot field added |
| `internal/statusline/renderer.go` | `data.Usage` branches replaced by snapshot + `n/a`; comments updated |
| `internal/statusline/builder_test.go` | `mockUsageProvider` removed; snapshot-fixture tests added |
| `internal/statusline/renderer_test.go` | RateLimits rendering tests preserved; new fallback/n-a cases |
| `internal/statusline/usage_fallback.go` (new) | D7 scan-roots helper (primary-root resolution from a worktree `.git` file) + D8 context guard |
| `internal/statusline/usage_fallback_test.go` (new) | Worktree/degradation/abandonment tests (M4) |
| `internal/statusline/usage_guard_test.go` (new) | D6b structural guard test, extended marker set |
| `internal/cli/statusline.go` | Stale collector comments only (budget stays) |
| `internal/template/templates/.moai/config/sections/statusline.yaml` | Only if a comment references the removed paths (expected: no change) |

PRESERVE (do not touch): `context_usage.go`, `quota.go`, `quota_dirs.go`, `quota_test.go`,
`quota_dirs_test.go`, `context_usage_quota_test.go`, `state_anchor.go`, the stdin parsing
(`stdinfields*`), the render budget, and every other SPEC's artifacts.

## §E. Out of Scope

- §D of spec.md (snapshot schema, aggregator rules, stdin parsing, cache migration,
  docs-site, any new network source).
- No template behavioral change: the deployed project surface (segment toggles, wrapper
  script, settings wiring) is unchanged.

## §F. Risks

- **R1 — Renderer test churn**: `renderer_test.go` is large; the `data.Usage` branches are
  exercised by existing tests. Mitigation: run the full package suite per milestone; the
  RateLimits cases must pass unmodified (AC-SOU-005), collector cases are deleted with
  their code.
- **R2 — Cross-platform**: the removed keychain read was macOS-only; the fallback adds no
  platform-specific code. `GOOS=windows go build ./...` in the verification batch guards
  it.
- **R3 — Test-fixture staleness**: snapshot fixtures must be written through the real
  record schema (schema v3 field names), not ad-hoc JSON, or the fallback reads them as
  unparseable and renders `n/a` — a false green for AC-SOU-004. Mitigation: seed fixtures
  via `SessionTelemetryRecord` structs.
- **R4 — Aggregator max-age default too tight for the render path**: a record fresher than
  maxAge is required; with the heartbeat interval at 5 min a live Claude session refreshes
  its record well within that, and `unknown` degrades to `n/a` (never wrong data). The
  D4 decision is revisited only if the AC-SOU-004 test shows the default rejects a
  just-written fixture.
- **R5 — Primary-root derivation edge cases (audit B1 follow-through)**: a `.git` file
  whose `gitdir:` line is malformed or does not match the
  `/.git/worktrees/<name>` shape (e.g. a submodule or a moved worktree) degrades via D7's
  fail-open path to today's single-root scan — the worst case is `n/a`, never wrong
  data. The M4 degradation tests pin each case.
- **R6 — Abandoned read still consumes wall-clock in its goroutine (audit B2 residual)**:
  the D8 context select returns the render promptly but does not interrupt an in-flight
  `os.ReadFile`; the goroutine finishes when the read does. Accepted in REQ-SOU-006
  because records are tool-written small JSON; the abandonment test asserts the
  caller-visible contract (prompt return + `n/a`), not goroutine teardown.

## §G. Verification plan

Per-milestone: `go build ./...` + `go test ./internal/statusline/ -count=1`.
At completion (single parallel batch): package suite, race variant, `GOOS=windows
GOARCH=amd64 go build ./...`, `golangci-lint run --timeout=2m`, `make build` (when M5
touched a template), and the two-cell re-runs named in acceptance.md §D.

---
id: SPEC-STATUSLINE-OFFICIAL-USAGE-001
title: "Statusline usage lookup: official input only (stdin rate_limits + local snapshot)"
version: "0.3.0"
status: draft
created: 2026-10-15
updated: 2026-10-15
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/statusline"
lifecycle: spec-anchored
tier: M
tags: "statusline, usage, security, legal-compliance, oauth-removal, official-input-only"
---

# SPEC-STATUSLINE-OFFICIAL-USAGE-001

## §A Problem Statement

The statusline's 5H/7D usage bars are currently fed, when the stdin payload carries no
`rate_limits`, by a network-and-credential collector in `internal/statusline/usage.go`
(`usageCollector`, REQ-V3-API-001..010 of SPEC-SLV3-001). That collector performs three
operations, none of which is documented Anthropic API surface:

1. **Haiku probe** — `POST https://api.anthropic.com/v1/messages` with model
   `claude-haiku-4-5-20251001`, `max_tokens=1`, header `anthropic-beta: oauth-2025-04-20`,
   reading the unified rate-limit response headers
   (`usage.go:22-41`, `fetchUsageFromHeaders` / `fetchUsageFromHeadersWithURL`).
2. **OAuth usage endpoint** — `GET https://api.anthropic.com/api/oauth/usage` with the same
   beta header and up to 3 retries with exponential backoff on 429
   (`fetchUsageFromOAuthAPI`).
3. **Credential reads** — the OAuth token is obtained from the macOS keychain
   (`security find-generic-password -s "Claude Code-credentials" -w` via
   `readTokenFromKeychain`), falling back to `~/.claude/.credentials.json` and
   `~/.claude/credentials.json` (`readOAuthToken`).

All three are undocumented paths. The Claude Code legal-and-compliance terms state that
tooling "may not collect, store, or intermediate Claude.ai credentials or session tokens".
Reading Claude Code's own credential store to issue API calls on the user's behalf —
including a `max_tokens=1` probe whose only purpose is reading response headers — falls
under that prohibition. The operator decision (card t1732, 2026-10-14, Class C,
Template-First) directs their removal.

The rendered consequence today: under a GLM-form stdin payload (no `rate_limits`), the
builder gate at `builder.go:406` (`if b.usageProvider != nil && (input == nil ||
input.RateLimits == nil)`) still fires the collector, so every GLM-pane render carries the
network and keychain cost above (bounded by an 800 ms render budget in
`internal/cli/statusline.go:19`, and a 3 s fetch timeout in `usage.go:40`) — for a bar that
then often renders `0%`.

**Verified code facts this SPEC builds on** (all observed on tree
`1ced8cab9435b655934529fdff2d3648441003c1`):

- The renderer already prefers the official stdin input: `renderer.go:504-528` renders the
  5H/7D bars from `data.RateLimits` (Claude Code 2.1.80+ stdin JSON) and falls back to
  `data.Usage` (the collector) only when stdin carried none.
- The context-usage snapshot the statusline already writes on every render
  (`<projectDir>/.moai/state/context-usage/<session-id>.json`, schema version 3) carries
  the optional `five_hour` / `seven_day` windows (`QuotaWindowRecord`: `used_percentage`,
  `resets_at` Unix seconds, `exhausted_at`) plus `captured_at` (RFC 3339)
  (`context_usage.go:97-139`). A session whose stdin supplied no windows writes a record
  without them.
- The account-level reader for those records already exists and is test-verified:
  `AggregateQuota` / `AggregateQuotaDirs` (`quota.go`, `quota_dirs.go`,
  SPEC-QUOTA-AWARE-SCHEDULING-001 M2, SPEC-QUOTA-RECORD-WORKTREES-001 M3) scan the
  context-usage records of the primary checkout and every linked worktree, pick per window
  the record with the latest capture time, and apply freshness, max-age, clock-skew, and
  reset-rollover rules. It is a pure reader: no network, no process spawn, no write. Its
  only in-repo callers today are its own tests — the fallback below becomes its first
  production consumer.
- **Worktree blindness (plan-audit iteration-1 B1)**: `QuotaStateDirs` enumerates linked
  worktrees only when `<root>/.git` is a directory (`quota_dirs.go:43` — `!info.IsDir()`
  returns the caller's own state dir alone). A linked worktree's `.git` is a regular file,
  so a render anchored at a worktree root never sees the primary checkout's records. The
  blind spot is live on this repository's own card trees: this worktree's `.git` is a
  regular file (`gitdir: /Users/goos/MoAI/moai-adk-go/.git/worktrees/t1732`). The fallback
  therefore resolves the primary checkout root builder-side (§B) and keeps
  `quota_dirs.go` untouched.
- **Bounded-read residual (plan-audit iteration-1 B2)**: the aggregator's bounds are
  metadata-only — `maxDirs` caps the worktree-metadata enumeration, and the ModTime
  pre-filter caps how many records are parsed, but `ReadSessionTelemetry` is an unbounded
  `os.ReadFile` (no size cap, no context). REQ-SOU-006 is scoped to what the mechanism
  delivers: bounded best-effort, not an absolute latency guarantee.
- The usage collector is self-contained in `internal/statusline`: a repo-wide search for
  `CollectUsage|UsageProvider|NewUsageCollector` outside `internal/statusline` returns
  nothing, so removal has no cross-package blast radius.

## §B Direction

**Official input only.** The statusline keeps exactly two sources for the 5H/7D bars, both
official or local:

1. **stdin `rate_limits`** (unchanged — the official Claude Code 2.1.80+ statusline JSON).
   While present, bars render from it byte-identically to today (REQ-SOU-002).
2. **The local context-usage snapshot** (new fallback). While stdin carries no
   `rate_limits`, the builder reads the account-level aggregate via the existing
   `AggregateQuotaDirs` reader over scan roots it resolves builder-side: the state
   directory of the primary checkout **plus** every linked worktree, derived from
   `projDir` even when the session is anchored inside a linked worktree (whose `.git` is
   a regular file — see the §A worktree-blindness fact; quota_dirs.go itself is not
   modified). The renderer shows the freshest recorded `five_hour` / `seven_day` values
   **together with the source record's `captured_at` time** (REQ-SOU-003). When no usable
   record carries a window — or the window has rolled over — the segment renders **`n/a`**
   instead of today's `0%` bar (REQ-SOU-004).

Everything that touches the network or a credential store is deleted: `usage.go` in full
(both fetchers, `readOAuthToken`, `readTokenFromKeychain`, the `usageCacheFile` cache at
`~/.moai/cache/usage.json`), the `UsageProvider` interface and `NewUsageCollector`, the
builder's `usageProvider` field / option / auto-create block / gate, and the collector
tests. A guard test makes the removal permanent: it fails when any removed-path marker
(`api.anthropic.com`, `oauth-2025-04-20`, `find-generic-password`,
`Claude Code-credentials`, `haiku-4-5`) reappears in a production file under
`internal/statusline`.

The completion judgment from the card, restated as acceptance criteria:

- **AC (card 1)**: with GLM-form input (no `rate_limits`), a statusline build performs
  ZERO network calls and ZERO keychain reads — proven by test (AC-SOU-001, AC-SOU-002).
- **AC (card 2)**: with Claude-form input (`rate_limits` present), bar behavior is
  unchanged (AC-SOU-005).

## §C Requirements (GEARS)

**REQ-SOU-001** (Event-driven, security): **When** any statusline render executes, the
statusline process shall issue no HTTP request to `api.anthropic.com` (Messages API probe
or OAuth usage endpoint) and shall read no credential store (macOS keychain service
"Claude Code-credentials", `~/.claude/.credentials.json`, `~/.claude/credentials.json`).

**REQ-SOU-002** (State-driven, no-regression): **While** the stdin payload carries
`rate_limits`, the 5H/7D bars shall render from the stdin `rate_limits` exactly as before
this SPEC — same segment content for the same stdin — and the builder shall not read the
snapshot fallback in that case.

**REQ-SOU-003** (State-driven): **While** the stdin payload carries no `rate_limits`, the
5H/7D bars shall render the values aggregated from the context-usage session telemetry
records of the primary checkout and every linked worktree (`five_hour` / `seven_day`,
freshest capture time wins, subject to the existing freshness / max-age / clock-skew /
reset-rollover rules of the quota aggregator), each rendered bar carrying the source
record's `captured_at` time. A render anchored at a linked-worktree root shall read the
same record population as one anchored at the primary checkout.

**REQ-SOU-004** (State-driven): **While** no usable snapshot record carries a window
(aggregate state `unknown`) **or** the winning window has rolled over (aggregate state
`reset`), the segment shall render `n/a` — never a zero-percent bar, never a stale
percentage.

**REQ-SOU-005** (Capability gate, removal): **Where** the collector is removed, the
retired identifiers (`UsageProvider`, `NewUsageCollector`, `usageCollector`,
`UsageResult`, `UsageData`, `usageCacheFile`, `readOAuthToken`, `readTokenFromKeychain`)
and their tests shall be deleted from the package, the builder gate and auto-create block
shall be removed, and a guard test shall fail the suite when any removed-path marker
reappears in a production (non-test) file under `internal/statusline` — the network
endpoints, the beta header, the probe model, the keychain service and command, **and the
credential-file paths (`~/.claude/.credentials.json`, `~/.claude/credentials.json`)** —
so a credential read reintroduced under a renamed symbol is still caught by its path
string.

**REQ-SOU-006** (Ubiquitous, cost bound — bounded best-effort): **While** the snapshot
fallback is read, the read shall remain local-only and best-effort bounded: directory
enumeration is metadata-only, the count of parsed records is bounded by the aggregator's
modification-time pre-filter within the max age, no recursive walk / network / process
spawn is introduced, and the read phase shall run under the render's context so an
over-deadline read abandons the fallback (the segments degrade to `n/a`) rather than
stalling the render. Known residual, named and accepted: the aggregator's per-record
`os.ReadFile` carries no size cap and no context (quota.go is on this SPEC's PRESERVE
list), so a pathological on-disk record can consume wall-clock inside the abandoned read;
this is bounded in practice because the directory holds only records the tool itself
writes (small JSON, schema v3) and is not attacker-controlled. This requirement
deliberately does NOT promise an absolute latency bound — requirement strength matches
what the mechanism delivers.

**REQ-SOU-007** (Capability gate, Template-First): **Where** a template-side asset
(`internal/template/templates/**`) documents or references the removed usage collection,
the reference shall be updated in the template source and the Template-First build cycle
(`make build` re-embed) shall remain green; where no template asset references it, the
cycle shall be verified green without modification.

## §D Out of Scope

### §D.1 Out of Scope — exclusions

- Changing the stdin `rate_limits` parsing, the `QuotaWindowRecord` schema, the
  context-usage snapshot writer, or the quota aggregator's rules — all are consumed as-is.
- Migrating or deleting a stale `~/.moai/cache/usage.json` on user machines: the file is
  scratch data in a directory the update checker also uses; after this SPEC nothing reads
  it, and no cleanup runs.
- Docs-site (`adk.mo.ai.kr`) content — a separate repository; sync-phase may schedule it.
- Any new network source for usage data, even a documented one. If Anthropic ships an
  official statusline usage surface, a future SPEC adopts it.

## §E Requirement Retirement

The REQ-V3-API-001..010 requirements (usage API collection, cache, token handling) are
anchored only in `usage.go` comments and owned by SPEC-SLV3-001 (status: implemented,
phase "v2.x - Legacy"). This SPEC retires them by deleting their carrying code;
SPEC-SLV3-001's own artifacts are left untouched (legacy SPEC, no body edit). The
request-rate headers the Haiku probe read are an undocumented surface; no replacement
requirement carries them.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.3.0 | 2026-10-15 | manager-spec | Iteration-3 (board d-20261014T235405Z-8b55, superseding e209): the two plan-audit debts are absorbed into planned M4 scope — sou-debt-1 becomes AC-SOU-010's mid-read abandonment pin (context cancelled AFTER the fallback read starts, blocked/slow fixture; closes the entry-check mutant) and sou-debt-2 becomes D7/M4 work (quotaScanStateDirs resolves a relative `gitdir:` pointer via Clean+Join with projDir; relative-pointer fixture; AC-SOU-009 extended). The two debts no longer exist as conditions; they are in-scope work. |
| 0.2.0 | 2026-10-15 | manager-spec | Plan-audit iteration-1 remediation (card t1732): B1 worktree-blind fallback fixed builder-side (primary-root resolution, quota_dirs.go preserved); B2 REQ-SOU-006 rescoped to bounded best-effort with named residual; B3 guard markers extended to the credential files, `readOAuthToken`/`readTokenFromKeychain` added to the retired-symbol list, mutant closure recorded; non-blocking citation fixes (measured file sizes, corrected spans, AC-SOU-006 two-cell pair). |
| 0.2.0 | 2026-10-15 | manager-spec | Plan-audit iteration-1 remediation (card t1732): B1 worktree-blind fallback fixed builder-side (primary-root resolution, quota_dirs.go preserved); B2 REQ-SOU-006 rescoped to bounded best-effort with named residual; B3 guard markers extended to the credential files, `readOAuthToken`/`readTokenFromKeychain` added to the retired-symbol list, mutant closure recorded; non-blocking citation fixes (measured file sizes, corrected spans, AC-SOU-006 two-cell pair). |
| 0.1.0 | 2026-10-15 | manager-spec | Initial plan-phase draft (card t1732, Tier M). |

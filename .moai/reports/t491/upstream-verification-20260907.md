# t491 — Upstream Re-verification Record (2026-09-07)

Card t491 scope: GD-1/2/3/4 only. This record re-reads the upstream sources DIRECTLY on
2026-09-07 rather than relying on the 2026-09-06 sweep report
(`sweep-report-copy-20260906.md` in this directory, copied from the gitignored
`.moai/research/cc-update-2.1.247-to-2.1.263.md`).

## Claim

1. **GD-1** — the `context-window-management.md` threshold-table rows for Fable and
   Sonnet 5 are wrong: Fable is 1M (table says 256K) and Sonnet 5 is 1M (table row
   "Sonnet/Opus standard (200K)"). Haiku 4.5 = 200K and Opus 5 = 1M rows are already
   correct. Sonnet 4.x-era models stay on a 200K row (row separation is mandatory —
   moving real 200K models to 1M flips the error from safe-direction to stall-direction).
2. **GD-2** — same-machine cross-session messaging works on **every provider** from
   v2.1.248; the **cross-machine** provider exclusions (Amazon Bedrock / Claude Platform
   on AWS / Google Cloud's Agent Platform / Microsoft Foundry) REMAIN in force. The two
   axes must be stated separately.
3. **GD-3** — native Windows has same-machine cross-session messaging since v2.1.234
   (named-pipe socket). The rule's "Claude Code does not provide cross-session messaging
   on native Windows" is stale.
4. **GD-4** — same-machine messaging works in sessions with **feature-flag fetching off**
   from v2.1.248. The rule's claim that the telemetry flags turn messaging off silently
   is stale for v2.1.248+ (class-level statement only; no per-flag claims are verified).
5. **Mirror state** — `cross-session-messaging.md` local vs template differs by exactly
   ONE hunk: the template omits the local-only line `> Origin: SPEC-CODEX-SESSION-MSG-001
   (design.md §8 mapping).` (introduced by 9ef2b91e1, card t187). This is intentional
   template-neutrality divergence (SPEC IDs are a forbidden content class in templates,
   template-isolation doctrine §25.1 class C1), so `diff -q` rc=0 is unreachable for this
   file BY DESIGN. `context-window-management.md` local/template are byte-identical
   (rc=0) and must stay so after the edit.

## Evidence

Commands run in this worktree (branch `WT-cc-upstream-sweep`, base 615d18c1f) on 2026-09-07:

| # | Claim | Command | Observed |
|---|---|---|---|
| V1 | CHANGELOG reachable; size matches sweep's snapshot | `curl -sSL https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md -o /tmp/t491-changelog.md && wc -c` | `630840 /tmp/t491-changelog.md`; head: `## 2.1.263` |
| V2 | GD-1: Fable 5.1 = 1M | extract `## 2.1.257` section → `t491-v257.md` | line 2: `- Added Claude Fable 5.1 (claude-fable-5-1), now the default Fable model — 1M context, $10/$50 per Mtok with $0.25/Mtok cache reads` |
| V3 | GD-1: Sonnet 5 = 1M | extract `## 2.1.247` section → `t491-v247.md` | line 28: `- Changed Sonnet 5's default auto-compact window to its full 1M context, so sessions on the 1M window now auto-compact at about 967K tokens instead of about 934K` |
| V4 | GD-1: official model table | webReader `https://platform.claude.com/docs/en/about-claude/models/overview` (2026-09-07) | Context window row: Fable 5 `1M tokens` · Opus 5 `1M tokens` · Sonnet 5 `1M tokens` · Haiku 4.5 `200k tokens` |
| V5 | GD-2/GD-4: provider + telemetry | extract `## 2.1.248` section → `t491-v248.md` | line 8: `- Added cross-session messaging (SendMessage / ListAgents) between sessions on the same machine on Bedrock, Vertex, and Foundry, and when telemetry is disabled` |
| V6 | GD-2/3/4: official availability | webReader `https://code.claude.com/docs/en/cross-session-messaging` (2026-09-07) § Availability | "Cross-session messaging requires Claude Code v2.1.224 or later on macOS, Linux, and WSL 2, and v2.1.234 or later on native Windows." · OS: "available on macOS, Windows, and Linux, including Linux inside WSL 2." · same-machine: "available on every provider, including Amazon Bedrock, Claude Platform on AWS, Google Cloud's Agent Platform, and Microsoft Foundry, and in sessions that run with feature-flag fetching off. On those providers, and with flag fetching off, same-machine messaging requires Claude Code v2.1.248 or later." · cross-machine: "Claude can't find those sessions with an API key or on Amazon Bedrock, Claude Platform on AWS, Google Cloud's Agent Platform, and Microsoft Foundry." · socket mechanism: "The socket is a Unix domain socket on macOS and Linux, including Linux inside WSL 2, and a named pipe on native Windows." |
| V7 | our table rows wrong (pre-fix RED) | Read `.claude/rules/moai/workflow/context-window-management.md` @ 615d18c1f | line 24: `| Fable (256K) | 256,000 tokens | **90%** | ~230,000 tokens |`; line 25: `| Sonnet/Opus standard (200K) | 200,000 tokens | **90%** | ~180,000 tokens |`; lines 64 & 80: `90% on 200K/256K` |
| V8 | our OS/provider/flag bullets stale (pre-fix RED) | Read `.claude/rules/moai/workflow/cross-session-messaging.md` @ 615d18c1f | lines 21-24: OS bullet denies native Windows; Providers bullet claims unavailability on the four providers; Flag bullet claims the telemetry flags silently kill the channel |
| V9 | csm mirror drift = 1 hunk; provenance t187 | `diff -u <local> <template> \| grep -c '^@@'` → `1`; `git log -S 'SPEC-CODEX-SESSION-MSG-001' --oneline -- <local path>` | `9ef2b91e1 feat(mcp): Codex-Claude session messaging via moai MCP broker (SPEC-CODEX-SESSION-MSG-001, card t187) (#1606)`; the single hunk removes exactly the `> Origin: SPEC-CODEX-SESSION-MSG-001 (design.md §8 mapping).` blockquote on the template side |
| V10 | cw mirror byte-identical (pre-fix) | `diff -q <local> <template>; echo rc=$?` | `cw-diff-rc=0` |

Extract snapshots committed in this directory: `t491-v247.md`, `t491-v248.md`,
`t491-v257.md` (verbatim CHANGELOG sections for 2.1.247 / 2.1.248 / 2.1.257). Note: the
2.1.257 extract uses `## 2.1.258` as its end boundary because 2.1.256 was never released
and has no heading upstream (first extraction attempt with `## 2.1.256` ran to EOF,
6,171 lines — discarded).

## Baseline-attribution

- Every fetch and command above ran on **2026-09-07** in worktree
  `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t491`, branch `WT-cc-upstream-sweep`,
  HEAD `615d18c1f` (= origin/develop at branch creation; verified with
  `git rev-parse --short HEAD` and `git merge-base --is-ancestor` → rc 0).
- The 2026-09-06 sweep report is carried as `sweep-report-copy-20260906.md` for
  provenance only; **no claim in this record rests on it** — every claim was re-observed
  today from a primary source.
- Upstream CHANGELOG byte size (630,840) matches the sweep's 2026-09-06 record, so
  upstream appears unchanged between the two reads; this record's claims stand on
  today's fetch regardless.

## Gaps — explicitly NOT observed

1. Per-flag behavior of the four telemetry env flags individually
   (`CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, `DISABLE_TELEMETRY`, `DO_NOT_TRACK`,
   `DISABLE_GROWTHBOOK`): confirmed only at class level ("feature-flag fetching off") by
   the docs, plus the CHANGELOG's "when telemetry is disabled" wording. The fix states
   the class-level fact only; no per-flag claims are made.
2. Native-Windows **cross-machine** messaging availability: unstated in the docs. The
   fix claims Windows same-machine support only (v2.1.234+) and asserts nothing either
   way about cross-machine on Windows.
3. The 5th availability constraint (shared flag slot `tengu_harbor_kite`): mechanism is
   measured in `cross-session-messaging-detail.md`, not contradicted by today's docs —
   OUT OF SCOPE, untouched.
4. `context-window-management.md` line 34's parenthetical about Claude Code's slot-based
   reporting ("the Claude slot (Opus=1M, Sonnet/Haiku=200K)"): describes upstream
   reporting behavior, not verified this run — OUT OF SCOPE, untouched (flagged as
   stale-adjacent for a future card; adjacent prose mentions at lines 64/80 ARE in scope,
   only the "200K/256K" phrase is touched there).
5. GD-5/6/7/8/9 of the sweep: outside this card's scope (GD-6 operator-held, GD-8
   no-action recommended, GD-5/7/9 unassigned).

## Residual-risk

- Upstream docs/CHANGELOG can change again; the values corrected here reflect the
  2026-09-07 snapshot. A future Fable/Sonnet revision changing windows requires a
  re-sweep.
- The docs' class-level "feature-flag fetching off" statement is broader than the
  CHANGELOG's per-bullet wording; if a later release narrows it, the flags bullet needs
  a re-read.
- Row labels chosen ("Fable (1M)", "Sonnet 4.x / earlier standard (200K)") are
  family-level on purpose: the docs table today says "Claude Fable 5" while the
  CHANGELOG says Fable 5.1 is the default (both 1M), and no current source enumerates
  200K-window Sonnet versions precisely.

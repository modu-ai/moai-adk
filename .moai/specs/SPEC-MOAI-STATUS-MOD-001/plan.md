# Plan — SPEC-MOAI-STATUS-MOD-001

Measurements are cited as `M-n` / `G-n` from `spec.md` (§1, §7); tree `58dad3055`. Development mode: `tdd` (`.moai/config/sections/quality.yaml` → `development_mode: tdd`): each milestone writes its tests first and observes them red.

## §A Context

**Card.** t1437, Class C (design change), Mods series 4/6. Operator constraints from the leader dispatch are authoritative: three features — (1) a usage/context warning strip (AbovePrompt band) plus a spinner suffix from `$.session.usage()` fields `rateLimits` and `context.percent`, thresholds that must not contradict the in-repo usage gates of t1347 and t1442; (2) a lane notification toast on `session.receive` with a pure pass-through observer; (3) a health/status warning line from a `$.clock.every` timer comparing the installed `moai` binary against the running MCP server, plus a retention-state signal. Cross-cutting: strictly additive, every hook fail-soft. The design-given names were verified against the laid 2.1.287 typings and exist verbatim (M-5); the design's implied polling was superseded by the engine's pushed `session.measure` event (D-2, spec.md §5).

**Tier judgment — M.** Files touched: 11 under `mods/moai-status/` (7 source and config, 4 test files) plus the root `.gitignore` = 12, inside the Tier M band (5-15). Size: an estimate of 450-700 lines including tests, inside 300-1,000 (an estimate, not a measurement). Not constitutional, no Go code, no change to any distributed template. REQ 12 of 12 carried to ACs, AC 13 of 13. Tier L was weighed and rejected: no file-count or LOC signal reaches it, and a `design.md` / `research.md` pair would restate §1 and §B. Threshold for the plan audit: 0.80.

**Provisional parameters are named constants.** Every numeric value in §G is either a measured bound, a value cited from in-repo gate code (M-7), or a constant the operator's verdict (`decision-index.md`) sets at Kickoff. The tests assert the floors and properties of REQ-MSM-003/007, never a particular default, so M1-M4 do not depend on any open verdict.

## §B Architecture

### B.1 Files (the mod, a plugin of function hooks)

```
mods/moai-status/
  .claude-plugin/plugin.json   name moai-status, version, description, author, "types": "./types/index.d.ts" ($.state is used — sibling M-17 contract)
  hooks/hooks.json             { "modules": ["./register.ts"] }  — an array of exactly one path (sibling M-8)
  hooks/register.ts            the ONLY file that spells `$.…`: register(), the five events' six registrations (session.start, session.end, session.measure, session.receive, ui.render ×2 matchers), the health timer, the single `process.run(` site (`runDiag`), `$.ui.*`, `$.state.*`
  hooks/data.ts                `$`-free: the fixed argv table; measure classification (strip levels from the mirrored gate formulas); toast one-line builder (origin kind + bounded excerpt)
  hooks/health.ts              `$`-free: doctor box-row parser (STATUS token + MESSAGE text per check name); memory-doctor JSON parser (per-store topic_files/cap/findings); health-line composer (warn segments joined; empty → clear)
  types/index.d.ts             PluginState for the 'moai-status' plugin + shared types; exports a type led by `MoaiStatus…` (the contract must, sibling M-17)
  tests/pure/data.spec.ts      PURE (bun): measure classification, toast line builder, suffix marker builder
  tests/pure/health.spec.ts    PURE (bun): doctor row parser, memory JSON parser, line composer
  tests/engine.test.ts         ENGINE (claude plugin test): measure → state → render; receive toast pass-through; health cycle with stubbed processes and ui.status; fail-soft
  tests/status.test.tsx        ENGINE: AbovePrompt and Spinner mounted on terminal and desktop
  tsconfig.json                the options of the typings header (target es2023, strict, jsx h; include [".claude-plugin/types","hooks","types","tests"], exclude ["tests/pure"])
  README.md                    launch, test and validate commands; the additive-observer boundary
```

`$.state` keys: `usage` (the strip classification), `health` (the last good health classification), `notice` (a one-line failure notice for the strip's degraded state). State refs are `{ plugin, key }` object literals — `$.state.get({ plugin: 'moai-status', key: 'usage' })` — the shape M-13 observed validate enforce and list on its `state writes:` line.

**The `$` rule (sibling M-17, unchanged contract).** `claude plugin validate` refuses `$` passed to a function imported from another file, so `$` never leaves `register.ts`: it builds `const runDiag = (argv: readonly string[]) => $.process.run(argv, { timeoutMs: CMD_TIMEOUT_MS })` per dispatch and hands `runDiag` to the helper modules, which stay pure and therefore bun-testable. All `$` uses sit inside `register`'s callback scope so the validate `calls:` line stays exhaustive (M-11).

### B.2 Data flow

```
session.measure (pushed by the engine after each turn / on rate-limit movement)
  → classify (pure, in data.ts):  context percent vs soft/hard formulas of M-7;
                                  rateLimits percentUsed vs per-kind gate holds
  → $.state.set(usage)            only when the classification changed
  → ui.render hooks re-draw       state-get-during-render subscribes them (M-5)

session.receive → toast line (pure) → $.ui.toast → next(e)   always, on every path

$.clock.every(HEALTH_POLL_MS) → runDiag(argv-from-table) x3 (single flight, one at a time)
  → parse (pure, in health.ts) → compose line (pure)
  → $.ui.status(line | undefined)
```

A render hook never writes state and never calls `$.process.run`; handlers (measure, receive, timer, session.start/end) do. UI state is read from `$.state` (a hot reload loses module variables); the previous classification is compared before any state write.

### B.3 The fixed argv table (REQ-MSM-002)

| Name | argv | When |
|---|---|---|
| health-binary | `["moai", "doctor", "--check", "Binary Freshness"]` | health tick |
| health-mcp | `["moai", "doctor", "--check", "MCP Server Version"]` | health tick |
| health-retention | `["moai", "memory", "doctor", "--json"]` | health tick |

No other argv is built anywhere. The check names are the in-tree check identifiers (`internal/cli/doctor.go:210`, `internal/cli/doctor_mcp_version.go:40` — M-9). Two of the three commands are read-only diagnostics; the third (`doctor --check "MCP Server Version"`) **deletes the CLI's own dead PID-stamp files as it runs** (M-9, `doctor_mcp_version.go:50-51`) — a delete on the health timer, not a read, and not a mod write: the mod neither reads nor writes those files itself (spec.md §2). The operator weighing Q3's interval should weigh that each tick runs this delete too.

### B.4 Per-feature mechanics

**Feature 1 — strip (REQ-MSM-003, -004).** The `session.measure` handler runs the pure classifier **`classifyMeasure(input, band)`** — a pure function of the pushed figure and **explicit bounds**: `band` carries `autoCompactPct`, the soft/hard constants and the per-kind holds as plain fields (the signature that makes AC-MSM-003's clamp fixture writable — with the constants alone, hard = min(95, 85+10) = 95 is always ≥ soft, so the clamp case is reachable only through an input). The mod instantiates `band` once from the mirrored default constants (§G, G-12). Classification: for context, `percent` (absent → no reading) against `softPct(window)` and `hardPct(window)` re-deriving the in-repo formulas — soft = 50 when `window >= 500_000` else 90; hard = min(band.autoCompactPct + 10, band.hardCap), clamped up to soft (M-7, `internal/statusline/renderer.go:719-755`); for rate limits, each entry with `kind` `five_hour` or `seven_day` is warn at `percentUsed >= 90` / `>= 95` respectively, other kinds are info-only. The handler writes `$.state` only when the classification changed and always returns `next(e)`. The AbovePrompt hook: `hasSurvey` → pass; nothing above info → pass; else resolve `next(e)` (upstream drawing) and return a `Box` whose children lead with a one-line `Text` strip — `moai-status: ctx 91% (warn at 90)` or `moai-status: 5h quota 92% · 7d 31%` — followed by the upstream tree so later mods still draw.

**Feature 1 — spinner suffix (REQ-MSM-005).** Same classification, shortest form: when context is warn/critical the suffix marker is ` · ctx NN%`; else when a rate window is warn the marker is ` · 5h NN%` or ` · 7d NN%`; else no marker. The hook passes `next({ ...e, props: { ...e.props, suffix: e.props.suffix + marker } })` — the incoming suffix (the engine's ellipsis by default) is preserved, only the suffix prop changes. Nothing to say → `next(e)` unchanged. Composition when another mod already wrote a suffix is Q5.

**Feature 2 — lane toast (REQ-MSM-006).** The `session.receive` handler builds one line in the pure builder: `<origin.kind>: <first 80 code points of e.text, first line only, no control characters>`; calls `$.ui.toast(line)` (a void call) and then returns `next(e)` — the pass-through is not downstream of the toast's success; the toast is attempted first and its failure changes nothing about the delivery. Every inbound origin kind is toasted in the MVP (Q6 may narrow it).

**Feature 3 — health cycle (REQ-MSM-007, -008, -012).** `session.start` registers the `$.clock.every` timer and stores the cancel handle; `session.end` cancels it. A tick runs the three argvs sequentially, one at a time, under a single-flight gate (a tick finding one running is dropped); each run carries `timeoutMs: 20000`. The pure health parser reads each doctor output for its box row (`│   <status>   <check name>  <message>  │`), **parsing the row first**: STATUS token in {`ok`, `warn`} (the only tokens the two asked checks assign — REQ-MSM-012; any other token, `fail` included, is not a verdict) and the MESSAGE text, matched against the spellings pinned from the in-tree sources (M-9) — `binary is behind source tree (binary: <sha9>, HEAD: <sha9>)` → warn segment `binary 802a72235 behind 58dad3055`; `binary matches source HEAD` / `binary from a different branch` / `no running moai MCP server recorded` / `N running MCP server(s) …` → healthy; `binary is newer than this tree — freshness undetermined (binary: <sha9>, HEAD: <sha9>)` → warn segment (freshness undetermined — the AC-MSM-008 reading, confirmed by `doctor.go:646-653`); `running MCP server is stale (pid N: <sha9>; binary: <sha9>)` → warn segment with pid and commits; anything else — including a missing row, a non-zero exit with no parseable row for the asked check, a rejected run — → `unknown`, never healthy, never a session failure (REQ-MSM-012, -009). The memory JSON parser walks the stores array and keeps the worst over-cap ratio (`topic_files` > `cap`) and the finding codes `MEMORY_TOPIC_COUNT_OVER_CAP` / `MEMORY_INDEX_OVERFLOW`; over cap → warn segment `memory 1422/50 files`. The composer joins the warn segments with ` · `; a non-empty result pins via `$.ui.status(line)`; empty → `$.ui.status(undefined)` (warnings-only, Q4). An unknown source shows as `binary ?` / `mcp ?` / `memory ?` in the line and keeps the last good classification in `$.state` (REQ-MSM-009).

**Version-diff semantics.** The line names both short SHAs for a behind binary, so the operator sees the gap without running anything; the mod never compares versions itself beyond displaying what the doctor reported (no `moai version` call, no git).

### B.5 Render-hook discipline

Both `ui.render` hooks read `$.state` (subscribing their instances, M-5) and call `$.ui.resolve(e)` for the element table, handed to the `$`-free drawing helpers. Neither calls `$.process.run` — timers and event handlers do. A render hook's failure falls back to `next(e)` (draw nothing of its own) inside the fail-soft guard, so a broken strip can never blank the band for later mods.

## §C Pre-flight (re-measure at run-phase entry; a different value stops the run and reports)

Each is a plain command whose output is read, not remembered. Values below are the plan-time snapshots.

| # | Command | Snapshot at `58dad3055` |
|---|---|---|
| C1 | `claude --version` | `2.1.287 (Claude Code)`; if the build changed, re-lay the typings (C4) and re-read them before coding against any name |
| C2 | `mkdir -p /tmp/msm-claude-cfg-empty` then `CLAUDE_CONFIG_DIR=/tmp/msm-claude-cfg-empty claude plugin test mods/moai-status` | exit 1, stderr `…: no such plugin folder` until M1; once the mod exists, tests execute. The operator profile may print the rollout-switch refusal instead (sibling M-13) — record which profile ran |
| C3 | `claude plugin validate mods/moai-status` | exit 1, `File not found` until M1 |
| C4 | the M-4 route: a scratchpad stub + `CLAUDE_CONFIG_DIR=<empty dir> claude -p "reply with just: ok" --plugin-dir <stub>` lays `<stub>/.claude-plugin/types/claude-code/index.d.ts`; grep it for `usage`, `measure`, `status`, `toast`, `receive` | names per M-5; a changed name stops the run and reports the deviation |
| C5 | `timeout 60 moai doctor --check "Binary Freshness"` and `--check "MCP Server Version"` | one box row each; the exact message spellings per M-9; a reflowed box is a finding to report, not a parse failure to absorb silently |
| C6 | `timeout 30 moai memory doctor --json` (redirect to file, bounded tail) | exit 0, an array of stores per M-8 |
| C7 | `git check-ignore -v mods/moai-status/.claude-plugin/types/claude-code/index.d.ts` | exit 1 (not ignored) until M5 |
| C8 | `git log --oneline -1 -- mods/` | nothing for `moai-status` (the sibling's `moai-board` may exist after its merge — never touched) |
| C9 | `bun --version`, then `bun test mods/moai-status/tests/pure/ --reporter=junit --reporter-outfile=/tmp/msm-junit.xml` | `1.4.2`; exit 1 "did not match any test files" until M1. bun is developer-local — no CI or tooling file references it (sibling M-15) |
| C10 | `git rev-parse --short HEAD` and `git status --short` | the worktree tip this plan measured; a moved HEAD means another actor wrote the tree — re-read before staging anything |

## §D Constraints (not renegotiated in run-phase)

1. Strictly additive observer: every `session.measure` / `session.receive` handler returns `next(e)` on every path; no `{ consumed }` result is ever produced; no turn-flow event other than the five registered ones is touched (REQ-MSM-001).
2. argv lists only, one `$.process.run` call site in `hooks/register.ts`, no shell string, fixed table only (REQ-MSM-002); `$` appears in no other file — helper modules receive `run` and the resolved element table as arguments (sibling M-17).
3. No network, no file write, no tool call, no prompt submission, no model/agent/settings/env access — enforced by the validate `calls:` allow-list pair (AC-MSM-001), knowing the line is a static listing with scope limits (M-2, M-11).
4. The mod is never embedded in `internal/template/templates/` and never deployed by `moai init` / `moai update` (REQ-MSM-011).
5. No change to any Go file, any `moai` command, any rule file, or any `mods/moai-board/` / t1436 file (AGENTS.md §5 scope discipline). The one root-file edit is the `.gitignore` line `mods/*/.claude-plugin/types/` if still absent (C7), placed beside the sibling's planned identical line so either merge order lands one rule.
6. Test files are named by runner: pure tests are `*.spec.ts` under `tests/pure/` (bun; never `*.test.ts`, which the engine runner globs), engine tests are `*.test.ts` / `*.test.tsx` directly under `tests/`.
7. Every commit on the branch names card `t1437` in its message; the evidence path is `.moai/reports/t1437/verdict.md` (written by the lane, not by this plan).

## §E Self-verification (plan-phase)

Run by the plan author and recorded in `progress.md` §E.1: ID pattern check (`PASS`), ID uniqueness (`ls .moai/specs | grep -c MOAI-STATUS` → 0 before authoring, M-10), frontmatter field presence (12 canonical fields), `moai spec lint SPEC-MOAI-STATUS-MOD-001` (result in §E.1), `OutOfScopeRule` heading shape (h3 subsections), artifact set = Tier M (spec, plan, acceptance) plus `progress.md` and `decision-index.md` (decision gate on).

## §F Milestones (ordered by likelihood of change: contracts and interfaces first, mechanics last)

**M1 — Contracts and skeleton (highest change likelihood).** `types/index.d.ts` (`PluginState` for `moai-status`: `usage`, `health`, `notice`; the classification and health-segment types), the argv table and the `run`-argument signatures of the helpers, `plugin.json` (with its `types` pointer), `hooks.json`, `tsconfig.json`, an empty `register.ts` that registers the five events and returns `next(e)` everywhere — with `runDiag` defined there, since `$` may not cross an import (sibling M-17). The first pure spec is written and observed red under bun (C9) before any classifier exists. Exit: `claude plugin validate mods/moai-status` exit 0 and its `hooks:` line names the five events' six registrations.

**M2 — Strip and spinner (TDD).** Tests first, red observed: pure under bun — the measure classifier (thresholds mirroring M-7, absent-figure handling, per-kind gate holds, changed-only classification), the suffix marker builder, the strip line builder; engine under the temp-config runner — `$.session.measure(fixture)` → `$.state` updated → `next(e)` returned; AbovePrompt mounted with `hasSurvey` true and false; Spinner suffix rewrite and pass-through. Exit: AC-MSM-003, -004, -005, -006 named tests pass (engine parts under `CLAUDE_CONFIG_DIR=/tmp/msm-claude-cfg-empty claude plugin test mods/moai-status`).

**M3 — Lane toast (TDD).** Engine: `$.session.receive(fixture)` → toast stub captured the one-line summary → `next(e)` passed; toast-call failure still passes the delivery; no `{ consumed }` on any path. Pure: the line builder (origin kind, 80-code-point first-line excerpt, control characters stripped). Exit: AC-MSM-007 named tests pass.

**M4 — Health cycle (TDD).** Pure under bun: the doctor box-row parser against the exact M-9 spellings (behind / fresh / newer-than-tree / different-branch / no-server / stale-server, plus summary-line rejection per REQ-MSM-012), the memory JSON parser (over-cap store selection, missing fields → unknown), the line composer (join, empty → clear). Engine: the cycle with stubbed processes — `on('process.run', …)` answering the three fixtures; single-flight; `$.ui.status` set and cleared; timer cancelled at `session.end`; a rejected run keeps the last good state. Exit: AC-MSM-008, -009 named tests pass.

**M5 — Packaging and handoff.** Add `mods/*/.claude-plugin/types/` to `.gitignore` if still absent; write `README.md` (launch: `claude --plugin-dir <absolute path to mods/moai-status>`; test and validate commands; the additive-observer boundary); re-run the regression guards of AC-MSM-012; leave AC-MSM-013 as a checklist for an operator in an interactive session. Exit: AC-MSM-001..012 evidence recorded in `progress.md` §E.2 with commands and verbatim output, each labeled pure (bun), engine (with the profile it ran under), or UNOBSERVED.

## §G Parameters

| Constant | Value | Basis |
|---|---|---|
| `HEALTH_POLL_MIN_MS` (floor) | 15,000 | REQ-MSM-007; the sibling's floor, kept so no mod cadence outruns the engine's measure pushes |
| `HEALTH_POLL_MS` | set by Q3 at Kickoff (provisional 60,000) | EVIDENCE-NEEDED (G-6); tests assert only the floor |
| `CMD_TIMEOUT_MS` | 20,000 | inside the 30 s engine default; the doctor single checks measured well under it (M-9) |
| `CTX_SOFT_LARGE_PCT` / `CTX_LARGE_WINDOW_CUTOFF` | 50 / 500,000 | M-7, `internal/config/defaults.go:533,535` (t1442 band) |
| `CTX_SOFT_STANDARD_PCT` | 90 | M-7, `defaults.go:534` |
| `CTX_AUTOCOMPACT_PCT` / `CTX_HARD_MARGIN_PCT` / `CTX_HARD_CAP_PCT` | 85 / 10 / 95 | M-7, `internal/statusline/memory.go:16`, `defaults.go:536-537` |
| `RL_FIVE_HOUR_HOLD_PCT` / `RL_SEVEN_DAY_HOLD_PCT` | 90 / 95 | M-7, `defaults.go:89-90` (t1347 gate holds) |
| `TOAST_EXCERPT_CP` | 80 code points, first line only | one-line rule of REQ-MSM-006; not measured against a live delivery (G-3) |
| `STATUS_LINE_MAX` | 200 code points | not measured against the status line's paint width (G-1); the composer truncates with an ellipsis rather than pinning a long line |

## §H Risks

| Risk | Effect | Handling |
|---|---|---|
| Early-access API drift (G-9) | a later build renames a noun or a prop | C4 re-lays and re-reads the typings at run-phase entry; names kept in one place per file |
| Doctor box-row reflow (G-5) | the health parser reads nothing | REQ-MSM-012: unknown state, line shows `?`, never healthy, never a session failure; drift reported as a finding |
| Memory-doctor payload growth (G-6) | parse cost, 4 MiB cap | one cycle per interval, `$` in flight off-budget; the scalars parse before the findings array is retained; a slimmer CLI form is a separate card |
| Rollout switch off (sibling M-13, G-10) | the operator profile refuses the runner and may not load the mod | engine tests run under an empty temp config dir (M-6, acceptance.md §A.2); a refusal there is UNOBSERVED, never a pass; the operator-profile refusal is the leader's report to the operator (Q7) |
| Timer leak across hot reload (G-8) | two health cycles in flight | the single-flight gate drops a tick that finds one running regardless of which instance owns it; `session.end` cancels |
| Toast volume (G-3) | every inbound delivery raises a toast | `timeoutMs` default keeps them transient; Q6 may narrow to peer/lane origins |
| Spinner suffix collision (Q5) | two mods append suffixes | the MVP appends to the incoming suffix verbatim and never removes what it finds; Q5 records the open composition rule |
| `$.state` subscription redraw storms | measure bursts redraw the band often | the handler writes state only when the classification changed (REQ-MSM-003); the engine folds measure bursts ("one at a time, a burst folding into one more", M-5) |
| Engine-written typings land in the repo (M-4) | untracked typings tree | `.gitignore` rule in M5 (C7) |

## §I Anti-patterns

- A `session.receive` handler that returns anything but `next(e)`'s result — consuming, holding, or answering a delivery.
- Calling `$.session.usage()` on a timer, or classifying anywhere but in the `session.measure` handler.
- A render hook that calls `$.process.run`, writes state, or drops the upstream tree when it draws.
- A second `$.process.run` call site, or an argv assembled from a template string.
- Treating a doctor summary line (`0 ok, 1 warn`) or an exit code as a check verdict (REQ-MSM-012).
- Reporting an AC as passed because the runner refused or printed nothing (acceptance.md §A).
- Presenting `bun test` results as engine evidence or as CI evidence (spec.md G-11), or naming a pure test file `*.test.ts`.
- A top-level `$` call outside `register`'s scope — it escapes the validate calls listing (M-11).

## §J Cross-references

`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier · `.claude/rules/moai/development/verification-completeness.md` §1-§2 (observed failure, two-cell adoption) · `.claude/rules/moai/core/verification-claim-integrity.md` §1, §2.2 (tool provenance: the `claude` build is `2.1.287`, the `moai` build `v3.2.0-rc.25`) · `SPEC-MOAI-BOARD-MOD-001` (the sibling mod whose layout and platform measurements this SPEC mirrors; no code shared) · `SPEC-HANDOFF-THRESHOLD-001` (the context-band thresholds mirrored, M-7) · `SPEC-QUOTA-AWARE-SCHEDULING-001` (the quota-gate holds mirrored, M-7).

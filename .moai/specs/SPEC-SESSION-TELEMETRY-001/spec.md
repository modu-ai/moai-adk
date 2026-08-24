---
id: SPEC-SESSION-TELEMETRY-001
title: "Per-session statusline telemetry — split the single-slot snapshot and record the session's model and effort"
version: "0.1.0"
status: draft
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/statusline
lifecycle: spec-anchored
tags: statusline, telemetry, session, context-window, docs-site
era: V3R6
tier: L
related_specs: [SPEC-WEB-CONSOLE-015, SPEC-HANDOFF-THRESHOLD-001]
---

# SPEC-SESSION-TELEMETRY-001 — Per-session statusline telemetry

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-08-24 | Initial draft. Split from SPEC-WEB-CONSOLE-015 per the ratified split design (`.moai/reports/t207/spec-split-design.md` §2, §3), which relocated the model/effort producer from the launcher to the statusline. |

## §A Background

Ground truth measured in worktree `.claude/worktrees/t207` at `dfbf828a6`. Every claim below
names the command that produced it; nothing is carried over from the parent SPEC's citations.

### A.1 One slot, N sessions — the defect, re-observed today

`internal/statusline/context_usage.go:134` builds the write path as
`filepath.Join(stateDir, "context-usage.json")` — exactly one file per project root, rewritten
wholesale by whichever session rendered its statusline last. The record carries `session_id`
and `writer_pid` **inside** the payload rather than in the path, so the file names no session
and the last writer wins.

The parent SPEC cited a May observation. That is a record; this is an observation, made in this
tree on 2026-08-24:

| Claim | Evidence |
|---|---|
| Three sessions were live | `.moai/state/active-sessions.json` held `2beac221…` (pid 15207, t219), `c15d8434…` (pid 51045, t210), `3db058e1…` (pid 36912, t207) |
| The single slot held a fourth session's telemetry | `.moai/state/context-usage.json` = `{"session_id":"d281730e-…","writer_pid":71763,"captured_at":"2026-08-24T13:08:03+09:00","raw_pct":56}` |
| No per-session directory exists | `ls -d .moai/state/context-usage` → `No such file or directory` |

Reading any one of the three registered sessions' context usage from that slot is impossible by
construction — not unlikely, impossible. The value present belongs to a session that is not
among them.

### A.2 The same render is the only place that holds the session's model and effort

The parent SPEC assigned model and effort to `kanban.Record`, written by the launcher. Measured,
the launcher cannot supply either value:

| Claim | Evidence |
|---|---|
| `moai cc` never parses or sets a model | `grep -rn '"-m"\|"--model"\|Model' internal/cli/cc.go` → 0 lines; the only `--model` match is line 36, a help string |
| `moai glm` sets four slots, not one model | `internal/cli/glm.go:350-353` sets `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL`; which slot the session runs in is not knowable at launch |
| `EffectiveProfile` returns a profile name, not a model | `internal/config/profile.go:92` returns `high`/`medium`/`low`; `ModelEffort` (`:73`) is documented at `:70-72` as a Claude short-alias vocabulary |

The statusline, by contrast, receives all three values per session, per render, from the
runtime:

| Claim | Evidence |
|---|---|
| Effort arrives on the render input | `internal/statusline/types.go:69` `Effort *EffortInfo`; `:131` `EffortInfo{Level}` |
| Model display name arrives on the render input | `internal/statusline/types.go:148` `ModelInfo.DisplayName` |
| Effort is threaded into the render data | `internal/statusline/builder.go:286-287` |
| The snapshot is written from the one function where the session id and the collected data are both in scope | `internal/statusline/builder.go:168` `writeContextUsage(resolveProjectDir(input), sessionID, os.Getpid(), data.Memory, …)` |
| A backend model resolver already exists and is already called | `internal/statusline/metrics.go:197` `resolveGLMModelName(displayName)`, called at `:51` — strips a `[1m]` suffix and substitutes the z.ai model when the `ANTHROPIC_DEFAULT_*` slots carry non-Claude names |

A live payload captured on this machine on 2026-08-24 13:02 (Claude Code 2.1.241, session
`d281730e-a47e-4f82-878e-5fd0ddc4dcb9`) carried
`{"session_id": …, "effort": {"level": "medium"}, "model": {"id": "claude-opus-5[1m]",
"display_name": "Opus 5 (1M context)"}}` — so effort is genuinely delivered, and this SPEC
requires that the record **carry** model and effort rather than that it carry them when
available. The empty path (§B REQ-ST-003) remains for a payload that omits them.

Model and effort therefore ride the per-session record this SPEC is already creating: one
writer, one write, no new producer and no launcher wiring. That is decision D-1, recorded in
`design.md` §1.1.

### A.3 The identifier the record is keyed by

The record is keyed by the identifier **the session itself was given** — the `session_id` field
of the render payload, which is by construction the id of the session doing the rendering.

This SPEC deliberately does **not** read `.moai/state/current-session-id.txt`
(`internal/session/registry.go:52` `CurrentSideChannelFile`). That file is a single
project-wide slot with the identical last-writer-wins shape this SPEC exists to remove;
sourcing the key from it would reproduce the bug inside the fix. Measured today, the sidecar
held `3db058e1…` (the t207 session, whose SessionStart wrote it last) while the statusline slot
held `d281730e…` — two surfaces, two answers, one project.

A consumer that needs a session's id obtains it from the session registry
`.moai/state/active-sessions.json` (`internal/session/registry.go:39`), which SessionStart
writes with the same runtime identifier.

### A.4 The move is not a rename — the consumer surface, measured

| Surface | Count | Command |
|---|---|---|
| Statusline package | 4 files | writer, reader, and their tests (§C.1) |
| A second, independent reader in `internal/cli` | 2 files | `tokens.go`, `tokens_test.go` |
| Always-loaded doctrine, two mirror pairs | 4 files | `grep -rln "state/context-usage.json" .claude internal/template/templates` |
| Published documentation, four locales | 12 files | `grep -rln "context-usage.json" docs-site/content \| wc -l` → `12` |
| A comment naming the file as a sibling state file | 1 | `internal/spec/drift_cache.go:24` |

The `internal/cli` reader is the hazard. It shares no constant with the statusline —
`tokensContextSnapshotFilename` (`tokens.go:30`) is its own hardcoded filename and
`tokensContextSnapshot` (declaration at `tokens.go:81`, doc comment at `:79`, `"raw_pct"` tag at
`:86`) its own duplicate of the schema — and `readTokensContextSnapshot` (`tokens.go:393-397`)
returns `nil` on any read error. So the path move breaks it with **no compile error and no
runtime error**: the context block simply stops appearing in `moai tokens` output.

`.moai/README.md` and its template mirror need **no change**. Both carry only the generic row
`| state/ | Runtime state snapshots, e.g. context-usage (gitignored — regenerated) |` — a
category mention with no filename, which stays true after the split. Recorded here so a later
reader does not re-derive the question.

## §B Requirements (GEARS)

### B.1 The per-session record

- **REQ-ST-001** — The statusline shall persist its telemetry snapshot to a per-session path
  `.moai/state/context-usage/<session-id>.json`, and shall not write the single-slot path
  `.moai/state/context-usage.json`.
- **REQ-ST-002** — The record shall be keyed by the identifier the session runtime delivered to
  that render, and shall not be keyed by any project-wide single-slot identifier file.
- **REQ-ST-003** — The record shall carry the session's model and its effort level alongside its
  context values. **When** a value is not supplied — because the render input omitted it, or
  because the record was written by a build predating this schema — every reader shall present
  it as not recorded, shall not infer or substitute a value, and shall not fail.
- **REQ-ST-004** — The recorded model shall be the model the session actually runs: **Where** a
  session runs against a non-Claude backend, the recorded value shall be that backend's model
  name rather than the Claude display name the runtime supplied.
- **REQ-ST-007** — **When** the identifier offered as the record's key would resolve outside the
  per-session directory, or is empty, the statusline shall refuse to persist that record rather
  than redirect it to another path, and the statusline render shall still complete.

### B.2 One reader

- **REQ-ST-005** — The `internal/statusline` package shall export exactly one reader for this
  record, and no second declaration of the record's schema shall exist elsewhere under
  `internal/`.
- **REQ-ST-006** — `moai tokens` shall obtain the snapshot through the reader of REQ-ST-005, and
  its output shall continue to carry the context block when a readable record exists for the
  session.

### B.3 Consumers of the moved path

- **REQ-ST-008** — **When** the per-session path is adopted, the consumer doctrine shall be
  updated in the same change to name the new path and to drop the single-slot validation steps
  the split makes unreachable, across both mirror pairs — the main rule and its detail
  companion, each with its template mirror.
- **REQ-ST-009** — **When** the per-session path is adopted, the published documentation shall be
  updated in the same change to name the new path, in every locale, with no locale left naming
  the old one.

## §C Constraints

### C.1 Blast radius inside `internal/statusline`

| Surface | Change |
|---|---|
| `context_usage.go:56` `contextUsageRecord` | widened to carry model and effort; the Go type name widens to `sessionTelemetryRecord` (D-1) while the on-disk path name is unchanged |
| `context_usage.go:27` `contextUsageSchemaVersion` | bumped — the payload gains fields |
| `context_usage.go:134` | the write path becomes per-session |
| `context_usage.go:186` `readContextUsage` | exported (REQ-ST-005); it is unexported today |
| `context_usage.go:203/216/236` `sameSemanticPayload` / `isRealSessionID` / `isFreshForSession` | validation written for a single slot. `isFreshForSession`'s session-id equality check and its `writer_pid` discriminator become unreachable once the path carries the session id; `sameSemanticPayload` stays — it is the write throttle, not a validity guard |
| `builder.go:168` | the call site of the persistence step; the one place where the session id, the model, and the effort are simultaneously in scope |
| `context_usage_test.go`, `builder_test.go` | assert the literal single-slot path in several places |

### C.2 The on-disk name does not change

Only the Go type name widens. The path stays `.moai/state/context-usage/<session-id>.json`
because the doctrine (4 files) and the published documentation (12 files) already point at that
name, and renaming it would move two things in one sweep. Ratified as D-1; consequences in
`design.md` §1.1.

### C.3 Template-First

Go source under `internal/statusline`, `internal/cli`, and `internal/spec` has no mirror under
`internal/template/templates/`, so the Template-First rule does not apply to the code changes.
It applies to exactly **two mirror pairs** — four files — belonging to REQ-ST-008. Measured, the
two pairs are byte-identical today (`diff -q` on each pair prints nothing, exit 0), so a change
that updates one side and not the other is mechanically detectable. All four are updated in the
same change, followed by `make build`.

### C.4 The record is render-ephemeral

The file is regenerated on every statusline render. There is no durable data to migrate, which
is why the migration is a hard cut with no dual-write window (D-3, `design.md` §1.2).

## §D Exclusions

Explicitly out of scope. Each may be taken up separately.

### Out of Scope — the session-id sidecar

- Repairing `.moai/state/current-session-id.txt`, which carries the same single-slot
  last-writer-wins shape (§A.3). This SPEC does not read it and therefore does not inherit its
  defect; fixing it is a separate change with its own consumers (`moai session current`, handoff
  `source_session_id` attribution).

### Out of Scope — the kanban record's session key

- Correcting `kanban.Record`'s keying, which the launcher writes with the *parent* session's id.
  A separate defect on a separate surface; this SPEC neither reads nor writes `kanban.Record`.

### Out of Scope — console presentation

- Any change under `internal/web`: the telemetry cells, the factory lane section, and the
  console's consumption of the reader this SPEC exports. Owned by `SPEC-WEB-CONSOLE-015`.

### Out of Scope — compatibility windows

- A dual-write window, a migration shim, or a reader that falls back to the single-slot path.
  The single slot **is** the defect; a compatibility window preserves it (D-3).

### Out of Scope — new telemetry values

- Recording cost, token spend, thinking budget, or any value the render payload does not already
  deliver. This SPEC records what the runtime hands the statusline, and nothing further.

### Out of Scope — the statusline's rendered output

- The statusline's own segments, layout, and thresholds. Only the persisted record changes; what
  the statusline prints is unchanged.

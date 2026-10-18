---
id: SPEC-AGENT-CORE4-001
title: "Design: agent sources and renderers"
version: "0.6.0"
created: 2026-10-18
---

# design.md — SPEC-AGENT-CORE4-001 (piece 1)

Direction only; identifiers are chosen by the run phase.

## §1 Agent source

Location `internal/template/agentemit/source/<name>.md`, outside `templates/` (never distributed,
same status as `agents-codex.yaml`). **Format decision (iter3 D1): option A.** A source is the
shipped definition file verbatim: the frontmatter keys as shipped (`name`, `description`, `tools`
CSV, `color`, `permissionMode`, `isolation`, `memory`, `skills`, `hooks`, `model`, `effort`) and the
body. The `tools:` CSV stays the capability list; a harness-neutral capability vocabulary and a
sandbox key are deferred to piece 2. The one addition is optional block markers on their own lines:
`<!-- claude-only -->...<!-- /claude-only -->` and `<!-- codex-only -->...<!-- /codex-only -->`.
Markers are fail-closed: an unterminated or nested marker fails emission naming the file (REQ-AC4-004).
None of the twelve bodies contains a marker at M1, so each source equals its baseline C2 file byte
for byte; the byte-identity ACs stay meaningful because they compare renders to git, not to the
sources.

## §2 Rendering (both renderers read the source)

- Claude renderer: source -> C2 bytes: codex-only blocks dropped, claude-only blocks kept (markers
  removed). With no markers present the output is the source bytes.
- Codex renderer: source -> Codex view (claude-only dropped, codex-only kept) -> the existing
  `renderTOML` path. C3 is produced from the source, not from C2 bytes; `emitAll` keeps its TOML
  construction, `tools:` token classification (an unmapped token already fails closed), permission
  contract and duplicate-name check, but its input is the Codex view. The old assertion that the
  Markdown publication is an identity pass-through becomes "rendered C2 equals a fresh render"
  (REQ-AC4-011).
- Block behavior is exercised with fixture sources; the twelve shipped bodies carry no markers.

## §3 Generation and check

`make agents-emit` writes C2, C3 for the twelve and, from M2, C1, C3p for the carried four (C1 of the eight joins when piece 2 absorbs or retires them, decision Q20).
`make agents-emit-check` re-renders in memory, compares the same sets, never writes, prints
`specs=<n> copies=<k> diff=<d>` and fails naming the first differing file. The Makefile target may
call a Go test or a small command, provided the summary line reaches stdout at exit 0.

## §4 Roster guard interaction

No agent name changes. The enumeration sweep: the highest distinct-name count in the twelve bodies
is 9 (manager-lead) against `SweepThreshold = 10`, so the source copies are not swept. The numeral
layer scans the repository and discharges hits through path-keyed registry rows; the source copy of
manager-design carries "13 retained agents" and needs rows keyed to
`internal/template/agentemit/source/manager-design.md` mirroring the C1/C2/C3 rows (AC-AC4-010).

## §5 Not here

Guards, size verb, core agents, retirement, hiding: pieces 2 and 3 (spec §D, §E).

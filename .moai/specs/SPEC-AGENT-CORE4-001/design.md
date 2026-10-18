---
id: SPEC-AGENT-CORE4-001
title: "Design: agent sources and renderers"
version: "0.3.0"
created: 2026-10-18
---

# design.md — SPEC-AGENT-CORE4-001 (piece 1)

Direction only; identifiers are chosen by the run phase.

## §1 Agent source

Location `internal/template/agentemit/source/<name>.md`, outside `templates/` (never distributed,
same status as `agents-codex.yaml`). Frontmatter: `name`, `description`, `capabilities` (ordered
tokens), `sandbox`, `model`, `effort`, `skills`, `memory`, `hooks`; then the body. Blocks:
`<!-- claude-only -->...<!-- /claude-only -->` and `<!-- codex-only -->...<!-- /codex-only -->`
(unnested, own lines). At M1 none of the twelve bodies contains a block, so each source is a pure
copy of its C2 file, which is what the pin file proves.

## §2 Rendering (both renderers read the source)

- Claude renderer: source -> C2 bytes: capabilities expand to the exact ordered `tools:` CSV
  through a manifest table (a token maps to one or more Claude tool names; `mcp` groups to their
  `mcp__moai__*` names); codex-only blocks are dropped, claude-only blocks kept (markers removed).
- Codex renderer: source -> Codex view (claude-only dropped, codex-only kept) -> the existing
  `renderTOML` path. C3 is produced from the source, not from C2 bytes; `emitAll` keeps its TOML
  construction, permission-contract evaluation, fail-closed token check and duplicate-name check but
  its input changes from "C2 bytes" to "the Codex view". The prior "Markdown publication is identity"
  assertion becomes "rendered C2 equals the pinned bytes" (REQ-AC4-012).
- A token without a manifest entry fails emission naming file and token; no partial output.
- A source whose `tools:` order cannot be expressed grows the token table; there is no raw escape.

## §3 Generation and check

`make agents-emit` writes C2, C3 for the twelve and, from M2, C1, C3p for the carried four.
`make agents-emit-check` re-renders in memory, compares the same sets, never writes, prints
`specs=<n> copies=<k> diff=<d>` and fails naming the first differing file. The Makefile target may
call a Go test or a small command, provided the summary line reaches stdout at exit 0.

## §4 Roster guard interaction

No agent name changes. The enumeration sweep: the highest distinct-name count in the twelve bodies
is 9 (manager-lead) against `SweepThreshold = 10`, so the source copies are not swept. The numeral
layer scans the repository and discharges hits through path-keyed registry rows; the source copy of
manager-design carries "13 retained agents" and needs rows keyed to
`internal/template/agentemit/source/manager-design.md` mirroring the C1/C2/C3 rows (AC-AC4-011).

## §5 Not here

Guards, size verb, core agents, retirement, hiding: pieces 2 and 3 (spec §D, §E).

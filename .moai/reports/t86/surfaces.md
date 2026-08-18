# t86 — Token-Usage Surfaces Investigation

> Card scope: measure-first. Which surfaces can report LLM token usage at card
> close time, and which of them `moai tokens record` (this card's seed) adopts.
> Precedent: `.moai/reports/t77/report.md` (GLM billing-path measurement).

## Surfaces

| Signal | Source | Coverage | Verdict |
|---|---|---|---|
| Per-message transcript usage | `~/.claude/projects/<dir>/<uuid>.jsonl` — assistant lines carry `message.model` + `message.usage` (`input_tokens`, `cache_creation_input_tokens`, `cache_read_input_tokens`, `output_tokens`), `isSidechain` per line | Every assistant API call in the session, both main and subagent origin; survives `/clear` (file is append-only) | **Adopted** — the record's primary source |
| Subagent sidechain transcripts | `<dir>/<uuid>/subagents/agent-*.jsonl` sibling files (CC 2.1.23x layout, measured: main file carries 0 `isSidechain:true` lines; siblings carry all) | Subagent-origin usage for sessions on current CC | **Adopted** — aggregated alongside the main transcript |
| Headless result `modelUsage` | `claude -p --output-format json` result object (t77 used it for per-model buckets) | Only headless runs; a subset of what the transcript carries | Reference only — not automated here (interactive sessions have no result object) |
| Statusline context snapshot | `<projectDir>/.moai/state/context-usage.json` (writer: `internal/statusline/context_usage.go`) | Context-pressure axis: raw_pct, stage, band at last render | **Adopted** — embedded as `record.context` when present |
| CC `/cost` | Interactive slash command | Session-cumulative cost, human-only surface | **Excluded** — interactive-only, no scriptable output |
| z.ai response headers / public API | — | GLM-pool billing-side usage | **Excluded** — not client-observable; z.ai has no public usage API (billing console only, t77-verified) |
| Codex pool | `moai codex task` delegations | Codex-backend consumption | **Gap** — flows outside CC transcripts; not implemented in this seed |

## Usage recipe

Lane records a card close from the card worktree root:

```bash
moai tokens record --session <session-uuid> --card <id> --role <lane>
```

`--session` resolves `~/.claude/projects/*/<uuid>.jsonl` (plus sibling
`<uuid>/subagents/*.jsonl`); `--transcript <path>` is the explicit-path
alternative. `--card` / `--role` are free-form labels (e.g. `t86`, `run`).

The LEAD session's own consumption uses the same command with
`--role lead --session <lead-uuid>` at batch close — the residual risk at
scale N is lead context pressure, so the lead records itself like any lane.

Probing without touching the ledger: add `--json` (prints the record, writes
nothing).

## Ledger

`<projectRoot>/.moai/state/token-accounting.jsonl` — append-only, one JSON
record per line, schema v1: `schema_version, recorded_at, session_id, card,
role, cwd, transcript, pools{glm|claude|<model>}→{main,subagent}→token
totals, models→per exact model name, messages{assistant,sidechain},
skipped_lines, context?` (context = the statusline snapshot subset, omitted
when absent).

## Probes (real transcripts, this machine, 2026-08-17)

- `tokens-probe-claude.json` — claude-opus-5 session `44400e2f`: 238 assistant
  messages, pool `claude` main input 475 / cache_read 77,804,403 / output
  154,499; no subagents in the session (sidechain 0).
- `tokens-probe-glm.json` — glm-5.2 session `a2045b55`: 378 assistant messages,
  298 sidechain via the sibling `subagents/agent-*.jsonl`; pool `glm` main
  input 2,828,267 / cache_read 20,806,912, subagent input 1,124,169 /
  cache_read 45,047,040. Demonstrates per-pool + per-origin split on real data.

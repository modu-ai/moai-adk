---
title: moai tokens
weight: 17
draft: false
new: true
added_in: "v3.1.1"
---

{{< new-badge v3.1.1 >}}

An accounting tool that records a Claude Code session's token usage **split by pool and by origin**. Viewed as one session-wide total, "how many tokens did this card burn" hides which backend spent how much, and how much went to the main conversation versus subagent side-chains. This command keeps the two splits.

{{< callout type="info" >}}
**One-line summary**: `moai tokens record` reads a session transcript, aggregates usage by pool (`glm`/`claude`/`other`) and by origin (main conversation / subagent side-chains), attaches the context-usage snapshot, and appends one line to `.moai/state/token-accounting.jsonl`.
{{< /callout >}}

## Usage

```bash
# Record by pointing at the open/latest session transcript
$ moai tokens record --transcript <path> --card t12 --role run

# Point by session id
$ moai tokens record --session <session-ID> --card t12

# Emit the record as JSON too (alongside the file write)
$ moai tokens record --transcript <path> --json
```

| Flag | Description |
|--------|------|
| `--transcript <path>` | The Claude Code transcript file to account |
| `--session <id>` | Point at the transcript by session identifier |
| `--card <card>` | The kanban card to book this usage against (e.g. `t12`) |
| `--role <role>` | The session's role (e.g. `run`, `sync`, `worker-3`) |
| `--json` | Also emit the record as JSON on standard output |

## What the record looks like

Records accumulate **append-only** in `.moai/state/token-accounting.jsonl` — a ledger that takes one line each time a session or card closes. Each line carries:

- **Usage by pool** — totals split into `glm` / `claude` / `other`. Which backend wrote the invoice is visible straight from the pool.
- **Usage by origin** — the main conversation versus subagent side-chains. In a run with several workers, this is what tells you whether the workers actually spent the implementation budget.
- **Context snapshot** — when the recorded session has a context-usage state (`.moai/state/context-usage/<session-id>.json`) at record time, its value rides along.

## When to record

By design, this is a record taken when a card or session **closes**. In kanban runs, leave one line per finished card; in a single session, one per finished chunk of work — that is what makes per-card cost comparison hold. The command itself consumes no tokens — it is accounting that re-counts usage already incurred, from the transcript.

## Related docs

- [Tokenomics overview](/en/advanced/tokenomics-overview) — why assignment matters more than unit price
- [Statusline](/en/advanced/statusline) — where usage is watched while the session runs
- [Kanban Mode](/en/advanced/kanban-mode) — the run shape that books cost by card and lane

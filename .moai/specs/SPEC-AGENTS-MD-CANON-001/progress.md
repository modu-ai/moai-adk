# SPEC-AGENTS-MD-CANON-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored 2026-08-22, rebuilt the same day on
`.moai/reports/t82/codex-probe.md` (v0.2.0): `spec.md`, `plan.md`, `acceptance.md`, `design.md`,
`research.md`, `progress.md`. Status `draft`.

Measured figures cited in the artifacts, with their commands:

| Claim | Command | Output |
|---|---|---|
| always-loaded rules 202,621 B (14 files) | `grep -rLE '^paths:' --include='*.md' .claude/rules \| sort \| xargs wc -c` | per `.moai/reports/t82/measurement.md` |
| `[HARD]` lines, rules | `xargs grep -h '\[HARD\]' < /tmp/t82_always.txt \| wc -c` | `30353` |
| `[HARD]` lines, `CLAUDE.md` | `grep -h '\[HARD\]' CLAUDE.md \| wc -c` | `2190` |
| `[HARD]` lines, output style | `grep -h '\[HARD\]' .claude/output-styles/moai/moai.md \| wc -c` | `11898` |
| imperative union, rules + `CLAUDE.md` | `grep -hE '\[HARD\]\|\bMUST( NOT)?\b\|\bshall\b' … \| sort -u \| wc -c` | `40501` + `3137` |
| Claude-only exclusion upper bound (6 files) | `grep -h '\[HARD\]' <6 files> \| wc -c` | `14360` |
| output style §8 share | `sed -n '193,713p' .claude/output-styles/moai/moai.md \| wc -c` | `46765` |
| codex cap / merge scope / silence | `codex debug prompt-input` fixture runs | per `.moai/reports/t82/codex-probe.md` |

Design rulings recorded at plan-phase:

- Option A (single root `AGENTS.md`, zero nested documents) — approved; measurement-forced.
- Contract ceiling 24,576 B with an 8,192 B reserve against the confirmed 32,768 B budget.
- CI byte guard mandatory and blocking (truncation measured silent).
- Shipped documentation warns about a user's global `~/.codex/AGENTS.md` (reasoning: `spec.md` §D.3).
- `.claude/output-styles/moai/moai.md` exempt from the contract, deferred for its own diet.

Gaps at plan-phase: the per-clause Codex-relevant / Claude-only split (upper bound only) and the
condensation ratio are unmeasured — both are M1 deliverables with a stop condition. Residual risk:
the probe covers macOS + `codex-cli` 0.147.0 only.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

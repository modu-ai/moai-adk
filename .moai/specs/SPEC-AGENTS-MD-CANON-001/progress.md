# SPEC-AGENTS-MD-CANON-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored 2026-08-22: `spec.md`, `plan.md`, `acceptance.md`, `design.md`,
`research.md`, `progress.md`. Status `draft`.

Measured figures cited in the artifacts, with their commands:

| Claim | Command | Output |
|---|---|---|
| always-loaded rules 202,621 B (14 files) | `grep -rLE '^paths:' --include='*.md' .claude/rules \| sort \| xargs wc -c` | per `.moai/reports/t82/measurement.md` |
| `[HARD]` lines, rules | `xargs grep -h '\[HARD\]' < /tmp/t82_always.txt \| wc -c` | `30353` |
| `[HARD]` lines, `CLAUDE.md` | `grep -h '\[HARD\]' CLAUDE.md \| wc -c` | `2190` |
| `[HARD]` lines, output style | `grep -h '\[HARD\]' .claude/output-styles/moai/moai.md \| wc -c` | `11898` |
| imperative union, rules + `CLAUDE.md` | `grep -hE '\[HARD\]\|\bMUST( NOT)?\b\|\bshall\b' … \| sort -u \| wc -c` | `40501` + `3137` |
| output style §8 share | `sed -n '193,713p' .claude/output-styles/moai/moai.md \| wc -c` | `46765` |

Gaps at plan-phase: the four entry premises in `spec.md` §D.1 are unmeasured; `.moai/reports/t91/`
is absent. Run-phase entry is gated on them.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

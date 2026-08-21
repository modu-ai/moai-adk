# SPEC-AGENTS-MD-CANON-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored 2026-08-22, rebuilt the same day on
`.moai/reports/t82/codex-probe.md` (v0.2.0), then revised against plan-audit iteration 1
(FAIL 0.69 → v0.3.0): `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`,
`progress.md`. Status `draft`.

**v0.3.0 revision — audit delta.** Blocking findings D1-D9 addressed: the ratchet criteria now run
under `go test -v` against a `t.Logf` M5 adds (D1a-b) and a derivation check binds the constant to
the achieved figure (D1c); the `AGENTS.md` singleton check moved from a global `find` to
`git ls-files … ':!internal/template/templates/'` so it no longer contradicts M6's mirror (D2);
"the integration branch" gained a discriminator and recording commands (D3); `AC-AMC-002` now cites
`probe-fixture.sh` (D4); the line-grep proxy is disclosed with both error directions and M1 moved
to clause blocks (D5); the cap-raise rationale corrected to the measured P4 position (D6); the
nested-`CLAUDE.md` asymmetry stated (D7); `REQ-AMC-006` recast to bind this SPEC's record with
`AC-AMC-012` covering it (D8); `AC-AMC-013` given a duplicate-line scan command (D9). Optional
D10-D12 also applied: `REQ-AMC-004` relabelled Unwanted, `REQ-AMC-006`'s leading `MAY` removed,
inline rationale moved out of `REQ-AMC-005` / `REQ-AMC-009` into §D.6 / §D.7. AC count 21 → 24,
REQ count 17 → 18 (Tier L ceiling 25 each).

**v0.3.4 revision — iteration 4 delta (FAIL 0.87, one finding).** E1: §C.4's required cuts were
computed over the unextended enumeration while `REQ-AMC-013` ¶2 requires one including
`AGENTS.md`; since the contract layer is net-additive (§D.2 + `REQ-AMC-001` + `REQ-AMC-002` put the
clauses in both places by construction), the cuts are corrected to **10,985** (this worktree) and
**15,055** (integration state) at `AGENTS.md`'s ceiling, with the formula
`stated surface + |AGENTS.md| − 66,371` governing. E2: M1's stop condition gained **Arm B** —
project the post-diet surface *including the contract layer* against 66,371 tokens, blocker on
shortfall — so the ratchet's reachability is tested before M2 rather than discovered at M5;
`AC-AMC-007` now covers both arms. E3 (optional) folded in while editing the bound: the ±1,000
tolerance makes 67,256 the strict maximum, so 66,371 is conservative by ~885 tokens.

Two dispatcher refinements folded into the earlier revision: `AC-AMC-016` now **cites the existing**
`TestAlwaysLoadedTokenBudget_OverBudgetFails` for the token-budget negative path rather than
proposing a duplicate fixture (verified passing on this tree), leaving only the Codex-cap dimension
as new coverage; and `design.md` §5.2 records why the singleton check uses `git ls-files` with
`:(top)` / `:(exclude,top)` rather than a path-scoped `find` — the latter closes the worktree
half but not the mirror half, since the mirror lives in the primary checkout too.

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
| `[HARD]` markers that are prose, not clauses | audit cross-check, re-derived | 15 of 93 |
| `[HARD]` markers ending in `:` (uncounted bodies) | audit cross-check, re-derived | 16 of 93 |
| codex cap / merge scope / silence | `codex debug prompt-input` fixture runs | per `.moai/reports/t82/codex-probe.md` |
| cap-raise scope + trust gate (P4) | `codex debug prompt-input` four-way differential | per `.moai/reports/t82/codex-probe-p4.md` |
| fixture reproducibility | `.moai/reports/t82/probe-fixture.sh` | rebuilds + reports each recorded run |

Design rulings recorded at plan-phase:

- Option A (single root `AGENTS.md` in the live tree, zero nested documents) — approved;
  measurement-forced. Revival condition recorded at `spec.md` §D.6.
- Contract ceiling 24,576 B with an 8,192 B reserve against the confirmed 32,768 B budget, stated
  as a bracket because the input figure is a line proxy.
- CI byte guard fails the build rather than warning (truncation measured silent, `spec.md` §D.7).
- Shipped documentation warns about a user's global `~/.codex/AGENTS.md` (reasoning: `spec.md` §D.3).
- Cap-raise forbidden as a diet substitute; target fixed at the untrusted first session's 32,768 B
  (REQ-AMC-018, `spec.md` §D.8).
- `.claude/output-styles/moai/moai.md` exempt from the contract; its first render-surface diet
  already landed (t131 / t142), a second pass decided by M1's result.
- M4 records the trust-notice obligation for t88's wiring generator (`plan.md` §E M4).

Gaps at plan-phase: the clause-block Codex-relevant / Claude-only split (line-proxy upper bound
only) and the condensation ratio are unmeasured — both are M1 deliverables with a stop condition.
Residual risk: the probe covers macOS + `codex-cli` 0.147.0 only; trust-acquisition path and
non-`trusted` `trust_level` values unmeasured.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

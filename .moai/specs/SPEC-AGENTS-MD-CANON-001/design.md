# SPEC-AGENTS-MD-CANON-001 — Design

Tier L design note. Covers the three decisions the milestones turn on: what the contract layer is,
how the documents are laid out, and where the guard lives.

## 1. What counts as "contract"

The redistribution needs a mechanical boundary between what must stay always-loaded and what may
move. The boundary used here is **the imperative clause**, detected in three widening rings:

| Ring | Detector | Rules + `CLAUDE.md` bytes | Treatment |
|---|---|---:|---|
| R1 | `[HARD]`-marked lines | 32,543 | Contract. Always-loaded. Non-negotiable. |
| R2 | `MUST` / `MUST NOT` / `shall` lines not already in R1 | +11,095 (union 43,638) | Contract by default; demotable only with a recorded per-clause ruling that the sentence is descriptive rather than binding. |
| R3 | everything else | ~179,500 | Eligible to move to skills and lazy companions. |

R3 is where the entire byte reduction comes from. R1 and R2 are where the SPEC's difficulty is.

The `[HARD]` lines in `.claude/output-styles/moai/moai.md` (75 lines, 11,898 B) form a fourth
category: binding, but binding *Claude's rendering*. They stay where they are and never enter
`AGENTS.md` — see `spec.md` §E.2.

## 2. The size problem, stated plainly

```
R1 verbatim, rules + CLAUDE.md            32,543 B
presumed shared cap (project_doc_max_bytes) 32,768 B
                                          ─────────
remaining for prose, structure, nesting        225 B
```

Three levers exist, in preference order:

1. **Condense R1** — rewrite each clause imperatively, dropping the rationale that currently
   travels inside the same line. Many `[HARD]` lines in this repo are 300-600 B because they carry
   their own justification inline; the obligation itself is usually one sentence. This is the
   lever M1 measures, and it is the only lever that helps every user.
2. **Raise the cap** — available only if P4 confirms project-scope configurability. Even then it
   is a repo-local remedy unless it can ship, so it is a supplement, never the plan.
3. **Narrow the Codex-relevant contract** — some `[HARD]` clauses bind Claude-only mechanisms
   (`AskUserQuestion` channel monopoly, output-style banners, `ToolSearch` preload). Codex has no
   counterpart, so those clauses can be excluded from `AGENTS.md` while remaining always-loaded on
   the Claude side. This is a *classification* lever, not a compression one, and it is safe: it
   removes nothing from either harness's binding surface.

Lever 3 deserves emphasis because it is cheap and the card missed it. A first-pass estimate is
that a meaningful share of R1 is Claude-mechanism-specific; M1 should produce the actual split
alongside the compression ratio, since the two together determine feasibility.

## 3. Document layout

### 3.1 Single-document fallback (if P2 shows no nested merge)

```
AGENTS.md          ← the entire Codex-relevant contract, condensed
CLAUDE.md          ← @AGENTS.md + Claude-only layer
```

This is the design that must work regardless of P2's answer, so it is the one M2 authors first.

### 3.2 Nested layout (only if P2 confirms a root → CWD merge chain)

Candidate owners, chosen so each maps to a directory a session actually works inside:

| Nested file | Owns | Carries |
|---|---|---|
| `internal/AGENTS.md` | Go implementation | coding standards, test discipline, error-wrapping, hardcoding prohibitions |
| `.claude/AGENTS.md` | harness authoring | agent/skill/rule authoring contracts, namespace separation |
| `internal/template/templates/AGENTS.md` | template source | Template-First rule, neutrality prohibitions, 16-language parity |
| `.moai/specs/AGENTS.md` | SPEC authoring | frontmatter schema, GEARS notation, out-of-scope convention |

The arithmetic constraint is on the **deepest reachable chain**, not the total: a session working
in `internal/template/templates/` loads root + that file, not all four. This is why the map is
drawn along directories that are worked in, not along topics.

If P2 instead shows per-changed-file merging, the map still holds but the arithmetic changes to
the worst-case set of simultaneously-touched directories — M3 recomputes it then.

## 4. Import layer

`CLAUDE.md` already uses `@`-imports (§9 imports `.moai/config/sections/user.yaml` and
`language.yaml`), and both resolve in live sessions today. So the mechanism is established by
observation, not assumed from documentation. The layer becomes:

```
CLAUDE.md
  @AGENTS.md                     ← shared contract, both harnesses
  <Claude-only layer>            ← AskUserQuestion channel, output-style pointer,
                                   ToolSearch preload, Agent-tool orchestration
  @.moai/config/sections/*.yaml  ← unchanged
```

The failure mode to watch is **duplicate injection**: leaving a copy of a contract clause inline
in `CLAUDE.md` while also importing it. AC-AMC-010 tests for exactly this.

## 5. Guard design

One enumeration, two thresholds. `alwaysLoadedSurface()` already walks the rule tree and the fixed
slots; the new guard adds a second assertion over a different file set (the `AGENTS.md` chain)
using the same helpers for path resolution and repo-root discovery.

```
token budget guard  : Σ estimateTokens(always-loaded surface)  ≤ AlwaysLoadedTokenBudget
codex chain guard   : Σ len(deepest AGENTS.md chain)           ≤ codexProjectDocMaxBytes
```

`codexProjectDocMaxBytes` is a named constant with its own provenance comment recording the P1
measurement — not a literal repeated at the call site, and not read from a user's codex config at
test time (the guard must be hermetic, like the existing one).

Failure output names the measured figure and the offending file, so the fix is obvious without
re-running a measurement by hand.

### Ratchet

`AlwaysLoadedTokenBudget` drops from 76,000 to a measured figure. Two traps:

- **Branch sensitivity.** This worktree measures ≈ 71,212 tokens — already under 75,000 — so a
  ratchet derived here would be vacuous. The release integration state that forced the raise
  measured 75,282. The figure must come from the integration branch (REQ-AMC-012).
- **Headroom.** The original constant carried ~15 % headroom over its baseline. Reusing that ratio
  keeps ordinary rule edits from tripping the guard while still catching a real regression; state
  the ratio used rather than picking a round number.

## 6. Rejected alternatives

| Alternative | Why rejected |
|---|---|
| Generate `AGENTS.md` from the rules at build time | Adds a generator to maintain, and the truncation hazard remains — a generated 200 KB file truncates exactly like a hand-written one. The problem is size, not authorship. |
| Symlink `AGENTS.md` → `CLAUDE.md` | Ships 20,523 B of Claude-specific orchestration to Codex and still leaves the rules unreachable. Solves nothing. |
| Move `[HARD]` clauses into skills and load them on demand | Forbidden by REQ-AMC-002. A rule that is not always present cannot bind every turn. |
| Diet the output style in this SPEC | Scope widening; ruled deferred in `spec.md` §E.2 with the measured justification (75.8 % of it is Claude render templates). |

## 7. Open questions carried into run phase

- The R1/R3 split assumes `[HARD]` marking is complete and accurate across the 14 files. If a
  binding clause exists that carries neither `[HARD]` nor an imperative keyword, ring detection
  misses it. M1's manual pass over the two pilot files should report whether it found any.
- Whether any nested `AGENTS.md` needs to repeat a root clause for a directory-local exception.
  Repetition costs shared budget twice; prefer a scoped clause in the root.

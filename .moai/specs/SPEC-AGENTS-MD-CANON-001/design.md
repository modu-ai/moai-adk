# SPEC-AGENTS-MD-CANON-001 — Design

Tier L design note. Three decisions carry the SPEC: what counts as contract, why the layout is a
single root document, and where the guard lives.

## 1. What counts as "contract"

The redistribution needs a mechanical boundary between what must stay always-loaded and what may
move. The boundary is **the imperative clause**, detected in three widening rings:

| Ring | Detector | Rules + `CLAUDE.md` bytes | Treatment |
|---|---|---:|---|
| R1 | `[HARD]`-marked lines | 32,543 | Contract. Always-loaded. Non-negotiable. |
| R2 | `MUST` / `MUST NOT` / `shall` lines not already in R1 | +11,095 (union 43,638) | Contract by default; demotable only with a recorded per-clause ruling that the sentence is descriptive rather than binding. |
| R3 | everything else | ~179,500 | Eligible to move to skills and lazy companions. |

R3 is where the entire byte reduction comes from. R1 and R2 are where the difficulty is.

A **fourth category** sits outside all three: the 75 `[HARD]` lines (11,898 B) in
`.claude/output-styles/moai/moai.md`. They are binding, but they bind *Claude's rendering*. They
stay where they are and never enter `AGENTS.md` (`spec.md` §E.2).

### 1.1 The Codex-relevant / Claude-only split

Within R1, a second cut matters: some clauses bind mechanisms Codex does not have —
`AskUserQuestion`, `ToolSearch` preload, `/clear` and the paste-ready handoff, prompt-cache
ordering, the `Skill` tool, cross-session messaging. Excluding these from `AGENTS.md` removes
nothing from either harness's binding surface; they remain always-loaded on the Claude side, via
`CLAUDE.md`'s Claude-only layer.

Measured upper bound: the `[HARD]` lines in the six most Claude-mechanism-bound files
(`askuser-protocol`, `session-handoff`, `cache-aware-execution`, `context-window-management`,
`cross-session-messaging`, `skill-routing`) total **14,360 B across 38 lines**.

```
grep -h '\[HARD\]' \
  .claude/rules/moai/core/askuser-protocol.md \
  .claude/rules/moai/workflow/session-handoff.md \
  .claude/rules/moai/workflow/cache-aware-execution.md \
  .claude/rules/moai/workflow/context-window-management.md \
  .claude/rules/moai/workflow/cross-session-messaging.md \
  .claude/rules/moai/workflow/skill-routing.md | wc -c    # → 14360
```

This is an upper bound, not the answer. File membership is a proxy: those files also carry
harness-generic principles (a subagent may not prompt the user; a peer session is not a proxy for
the user) that bind Codex exactly as they bind Claude. **The split is per clause, not per file**,
and producing it is M1's deliverable.

The two bounds it produces:

```
optimistic (full exclusion) : 32,543 − 14,360 = 18,183 B   before any condensation
pessimistic (no exclusion)  : 32,543 B                     condensation carries everything
ceiling                     : 24,576 B
```

## 2. The size problem, restated on measured ground

```
verbatim R1 contract, rules + CLAUDE.md    32,543 B
confirmed project_doc_max_bytes            32,768 B
                                           ─────────
headroom                                       225 B   (0.7 %)
```

It fits. That is the trap. Three claims on the same budget are invisible in the raw line total:

1. **Document structure.** `AGENTS.md` is a document, not a line dump — headings, section framing,
   and the connective prose that makes a clause findable all cost bytes.
2. **The global layer.** A user's `~/.codex/AGENTS.md` joins the same chain and is consumed
   **first** (`spec.md` §D.3). Its size is set on each user's machine and is unknowable here.
3. **Growth.** Every future rule edit that adds a `[HARD]` clause spends this budget.

With 225 B of slack, the first edit after landing starts silently dropping the tail. So the design
target is not the fit — it is the **8,192 B reserve** that keeps the fit true a year from now.

Two levers reach 24,576 B, in preference order:

1. **Classification (§1.1)** — free, and removes nothing from anything. Up to 14,360 B.
2. **Condensation** — rewrite each surviving clause imperatively, dropping the rationale that
   currently travels inside the same line. Many `[HARD]` lines in this repo run 300-600 B because
   they carry their own justification inline; the obligation itself is usually one sentence. M1
   measures the ratio.

A third lever — raising `project_doc_max_bytes` — is deliberately **not** used. It is a per-user
config setting, so it cannot help distributed users, and depending on it would make the shipped
contract silently correct on the maintainer's machine and silently truncated elsewhere.

## 3. Layout: one document, at the root

The probe settles this. Measured (`.moai/reports/t82/codex-probe.md`):

| Invocation | Root doc size | What loaded |
|---|---:|---|
| repo root | 40,040 B | root only — `area/` and `area/deep/` markers absent |
| `area/deep`, in a git repo | 42,066 B | root head only — **both nested docs never loaded** |
| `area/deep`, in a git repo | 28 B | root + area + deep |
| `area/deep`, not a git repo | 42,066 B | CWD's own doc only |

Three consequences:

- Nested documents **do not expand the budget**. They share the same 32,768 B, consumed root-first.
- Nested documents **are not loaded** when codex runs at the repo root — the ordinary case.
- A large root therefore **starves** the nested documents entirely, silently.

So the root document must be self-sufficient, and every nested document added would spend root
budget for a benefit reachable only from an area-scoped session. **Option A — a single root
contract — is the design.** Option B (root + a small number of nested documents) is admissible only
against evidence that sessions are habitually started inside those directories; absent that
evidence the measurement forbids it, because ordinary invocation would leave most of the ruleset
unloaded (`spec.md` REQ-AMC-005 / REQ-AMC-006).

```
AGENTS.md          ← the entire Codex-relevant contract, ≤ 24,576 B
CLAUDE.md          ← @AGENTS.md + the Claude-only layer
```

## 4. Import layer

`CLAUDE.md` already uses `@`-imports — §9 imports `.moai/config/sections/user.yaml` and
`language.yaml`, and both resolve in live sessions today. The mechanism is established by
observation, not assumed from documentation.

```
CLAUDE.md
  @AGENTS.md                     ← shared contract, both harnesses
  <Claude-only layer>            ← the clauses M1 classified Claude-mechanism-only
  @.moai/config/sections/*.yaml  ← unchanged
```

The failure mode to watch is **duplicate injection**: leaving a clause inline in `CLAUDE.md` while
also importing it. AC-AMC-012 tests exactly this.

## 5. Guard design

One enumeration, two thresholds. `alwaysLoadedSurface()` already walks the rule tree and the fixed
slots; the byte guard adds a second assertion over a different file set, reusing the same helpers
for path resolution and repo-root discovery.

```
token budget guard  : Σ estimateTokens(always-loaded surface)  ≤ AlwaysLoadedTokenBudget
codex chain guard   : len(AGENTS.md)                           ≤ agentsMDCeilingBytes (24,576)
```

`agentsMDCeilingBytes` is a named constant carrying its own provenance comment: the 32,768 B
measured default, the 8,192 B reserve, and what the reserve is protecting. It is not a literal at
the call site, and it is not read from a user's codex config at test time — the guard stays
hermetic, like the existing one.

**The guard is blocking, not advisory.** This is the design's single most load-bearing choice, and
it follows from a measurement rather than a preference: truncation emits nothing — stderr 0 bytes,
exit 0, the `truncating` string is a tracing event that never surfaces. There is no runtime signal
to fall back on, so an advisory guard would leave the failure with no detector at all.

### 5.1 Ratchet

`AlwaysLoadedTokenBudget` drops from 76,000 to a measured figure. Two traps:

- **Branch sensitivity.** This worktree measures ≈ 71,212 tokens — already under 75,000 — so a
  ratchet derived here would be vacuous. The release integration state that forced the raise
  measured 75,282. The figure must come from the integration branch (REQ-AMC-014).
- **Headroom.** The original constant carried ~15 % over its baseline. Reusing that ratio keeps
  ordinary rule edits from tripping the guard while still catching a real regression; state the
  ratio used rather than picking a round number.

## 6. Rejected alternatives

| Alternative | Why rejected |
|---|---|
| Nested `AGENTS.md` per area (the card's proposal) | Measured: unloaded at repo-root invocation, and shares rather than expands the budget. Adopting it would silently drop most of the ruleset in ordinary use. |
| Raise `project_doc_max_bytes` | Per-user config; cannot ship. Would make the contract correct on the maintainer's machine and truncated on everyone else's. |
| Generate `AGENTS.md` from the rules at build time | Adds a generator to maintain, and truncation is unaffected — a generated 200 KB file truncates exactly like a hand-written one. The problem is size, not authorship. |
| Symlink `AGENTS.md` → `CLAUDE.md` | Ships 20,523 B of Claude-specific orchestration to Codex and still leaves the rules unreachable. |
| Move `[HARD]` clauses into skills, loaded on demand | Forbidden by REQ-AMC-002. A rule that is not always present cannot bind every turn. |
| Diet the output style here | Scope widening; deferred with measured justification (`spec.md` §E.2 — 75.8 % of it is Claude render templates). |
| Size the contract to fill 32,768 B | The budget is not a target. A contract at 99.3 % of budget begins truncating on its first edit, silently. |

## 7. Open questions carried into run phase

- The R1 detector assumes `[HARD]` marking is complete across the 14 files. A binding clause
  carrying neither `[HARD]` nor an imperative keyword would be missed. M1's manual pass over the
  pilot files should report whether it found any.
- The probe ran on macOS with `codex-cli` 0.147.0. A smaller upstream default on another platform
  or version would invalidate the ceiling's calibration silently — re-probe on a codex upgrade
  (`spec.md` §D.5).

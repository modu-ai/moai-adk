# SPEC-AGENTS-MD-CANON-001 — Design

Tier L design note. Three decisions carry the SPEC: what counts as contract, why the layout is a
single root document, and where the guard lives.

## 1. What counts as "contract"

The redistribution needs a mechanical boundary between what must stay always-loaded and what may
move. The boundary is **the imperative clause**, detected in three widening rings:

| Ring | Detector | Rules + `CLAUDE.md` bytes | Treatment |
|---|---|---:|---|
| R1 | `[HARD]`-marked lines | 32,543 line-proxy / **51,639 measured** | Contract. Always-loaded. Non-negotiable. |
| R2 | `MUST` / `MUST NOT` / `shall` lines not already in R1 | +11,095 (union 43,638) | Contract by default; demotable only with a recorded per-clause ruling that the sentence is descriptive rather than binding. |
| R3 | everything else | ~179,500 | Eligible to move to skills and lazy companions. |

R3 is where the entire byte reduction comes from. R1 and R2 are where the difficulty is.

**The ring figures are line-level proxies.** `grep` finds *lines bearing a marker*, which is not
the same set as *the obligations those lines carry*. M1 expanded the markers to clause blocks and
measured the error (`spec.md` §A.4): it runs **almost entirely one way** — one marker over (357 B
of navigation prose) and ninety-six under — for a clause-block total of **51,639 B** against the
proxy's 32,543, **+58.7 %**. The magnitude of the proxy is reliable; its value is not. So R1's
detector is a **seed for classification, not the classification itself**, and §2's arithmetic below
is restated on the measured figure.

A **fourth category** sits outside all three: the 75 `[HARD]` lines (11,898 B) in
`.claude/output-styles/moai/moai.md`. They are binding, but they bind *Claude's rendering*. They
stay where they are and never enter `AGENTS.md` (`spec.md` §E.2).

### 1.1 The Codex-relevant / Claude-only split

Within R1, a second cut matters: some clauses bind mechanisms Codex does not have —
`AskUserQuestion`, `ToolSearch` preload, `/clear` and the paste-ready handoff, prompt-cache
ordering, the `Skill` tool, cross-session messaging. Excluding these from `AGENTS.md` removes
nothing from either harness's binding surface; they remain always-loaded on the Claude side, via
`CLAUDE.md`'s Claude-only layer.

Line-level upper bound, superseded by M1's per-clause split below: the `[HARD]` lines in the six
most Claude-mechanism-bound files
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

That was an upper bound, not the answer. File membership is a proxy: those files also carry
harness-generic principles (a subagent may not prompt the user; a peer session is not a proxy for
the user) that bind Codex exactly as they bind Claude. **The split is per clause, not per file.**

M1 produced it (`.moai/reports/t82/classification.tsv`, 97 rows, zero unclassified):

```
Codex-relevant (35 blocks) : 16,135 B
Claude-only    (61 blocks) : 35,147 B
prose, no obligation  (1)  :    357 B
                             ────────
clause-block total         : 51,639 B
ceiling                    : 24,576 B
```

Inside the six named files the proxy nearly held — 21,504 of their 22,179 B of clause blocks is
Claude-only — but 38.8 % of all Claude-only bytes (13,643 B) fall **outside** them, most of it in
`kanban-dispatch.md`, where Claude-session mechanisms and harness-generic discipline share a file.
The classification's error direction is a design constraint on M2: **when in doubt, classify to the
Codex side** (`spec.md` §D.2).

## 2. The size problem, restated on measured ground

```
verbatim R1 contract, rules + CLAUDE.md    51,639 B   (clause blocks, measured)
confirmed project_doc_max_bytes            32,768 B
                                           ─────────
overflow                                  −18,871 B   (1.58x the budget)
```

The line proxy read as a 225 B fit (0.7 % headroom); that fit was an artifact of its undercount, and
believing it was the trap. The contract does not fit verbatim. Three further claims on the same
budget are invisible in the raw line total anyway:

1. **Document structure.** `AGENTS.md` is a document, not a line dump — headings, section framing,
   and the connective prose that makes a clause findable all cost bytes.
2. **The global layer.** A user's `~/.codex/AGENTS.md` joins the same chain and is consumed
   **first** (`spec.md` §D.3). Its size is set on each user's machine and is unknowable here.
3. **Growth.** Every future rule edit that adds a `[HARD]` clause spends this budget.

A contract over budget truncates on the first run; one at 99.3 % of budget starts silently dropping
the tail on the first edit after landing. So the design target is neither — it is the **8,192 B
reserve** that keeps the fit true a year from now.

Two levers reach 24,576 B, in preference order:

1. **Classification (§1.1)** — free, and removes nothing from anything. Measured: 35,504 B of the
   51,639 B clause-block total (Claude-only plus the one prose block), leaving 16,135 B. Without it
   the contract does not fit at all, which makes this the load-bearing lever rather than the
   cheaper of two.
2. **Condensation** — rewrite each surviving clause imperatively, dropping the rationale that
   currently travels inside the same line. Many `[HARD]` lines in this repo run 300-600 B because
   they carry their own justification inline; the obligation itself is usually one sentence. M1
   measures the ratio.

A third lever — raising `project_doc_max_bytes` — is deliberately **not** used, but not for the
reason an earlier draft gave. That draft called it "per-user config, so it cannot ship";
`codex-probe-p4.md` measures otherwise. A project-scope `<repo>/.codex/config.toml` **does** take
effect and **beats** the user value — conditional on the user config registering the project
`trust_level = "trusted"` (four-way differential toggling only the trust entry: effective cap
4,096 → 8,192).

The real reason is stronger. A distributed user's **first** session is untrusted by construction —
they clone and run before any trust entry exists — so the effective cap then is 32,768 B, and a
contract that does not fit truncates **silently** on the run that forms their first impression
(stderr 0 bytes on all four probe runs, ignoring included). A lever that engages only after the
user has already trusted the project cannot carry a constraint that binds before they have.
Hence REQ-AMC-018: the reduction target is fixed at 32,768 B and any raise is headroom on top.
Full measured position: `spec.md` §D.8.

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

### 3.1 Why `CLAUDE.md` gets to be plural and `AGENTS.md` does not

A reader who knows this codebase will object immediately: five nested per-directory instruction
files already exist on the Claude side — `internal/{cli,config,hook,spec,template}/CLAUDE.md` — and
they work well. "One instruction file per repository" is plainly not this repository's shape, so
the single-root rule needs to say why `AGENTS.md` is the exception.

The two mechanisms differ on exactly two measured properties, and those two are the whole answer:

| | nested `CLAUDE.md` | nested `AGENTS.md` |
|---|---|---|
| **Loading** | by path relevance — present when work touches that directory | only along the git-root → CWD chain; **absent entirely at repo-root CWD**, the ordinary case |
| **Budget** | no byte cap over the set | one shared 32,768 B, consumed root-first — a large root starves them to nothing |

Nested `CLAUDE.md` earns its place because it is loaded when relevant and costs nothing when not.
Nested `AGENTS.md` inverts both: it is *absent* when most needed (a developer at the repo root) and
*costly* when present (it eats the root's budget). The rejection is of the codex mechanism's
properties, not of nested instruction files as an idea — and if a future codex version changes
either property, the revival condition in `spec.md` §D.6 is how that gets revisited.

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
also importing it. AC-AMC-013 tests exactly this.

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

**The negative path reuses what exists.** `TestAlwaysLoadedTokenBudget_OverBudgetFails` already
proves the token-budget guard fires on an over-budget surface and stays quiet on an under-budget
one; it passes on this tree. The Codex-cap dimension extends that table-driven test in the same
file rather than standing up a parallel harness — one measurement path, two thresholds, one place
to look when either moves.

### 5.2 Why the singleton check is `git ls-files` with two pathspec modifiers

`AC-AMC-010` asks a narrow question — does any `AGENTS.md` exist in the live tree outside the
repository root? — and three constructions answer it differently:

| Construction | Worktree copies | Template mirror | Verdict |
|---|---|---|---|
| `find . -name AGENTS.md` | counted (varies with live lanes) | counted | fails both ways |
| `find` scoped to the primary checkout root | excluded | **still counted** — the mirror lives in the primary checkout too | closes one half; `AC-AMC-010` still contradicts `REQ-AMC-015` |
| `git ls-files … ':!internal/template/templates/'` | excluded (untracked / separate index) | excluded by name | closes both |

The third is the form used. It carries two additional modifiers over the bare version, each closing
a residual hole: **`--full-name` with `:(top)`** makes the result independent of the directory the
criterion runs from (a plain `git ls-files` invoked from a subdirectory silently scopes to that
subdirectory — the same nondeterminism in a different disguise), and **`:(exclude,top)`** applies
the mirror exclusion from the repository root rather than relative to the caller. Measured against
the `CLAUDE.md` analogue in this worktree: `find` → 7, the guarded form → the 6 live-tree files with
the mirror correctly excluded.

**The guard is blocking, not advisory.** This is the design's single most load-bearing choice, and
it follows from a measurement rather than a preference: truncation emits nothing — stderr 0 bytes,
exit 0, the `truncating` string is a tracing event that never surfaces. There is no runtime signal
to fall back on, so an advisory guard would leave the failure with no detector at all.

### 5.1 Ratchet

`AlwaysLoadedTokenBudget` drops from 76,000 to a measured figure. Two traps:

- **Branch sensitivity.** This worktree measures 71,207 tokens (guard-exact; 71,212 by `total/4`) —
  already under 75,000 — so a ratchet derived here would be vacuous. The `release/v3.1.1`
  integration state that forced the raise measured 75,282; re-measured at run-phase pre-flight that
  state sits on no live branch and all four live refs read 71,207, so the divergence returns with
  the first sibling card on the v3.2 branch rather than being visible today. The figure must come
  from the **integration branch**: the `release/vX.Y.Z`
  branch this card merges into, which carries the merged state of the sibling cards. The
  discriminating property is the merged sibling state, not the branch's name, so the evidence
  records `git rev-parse --abbrev-ref HEAD` and `git rev-list --count main..HEAD` — otherwise any
  branch can be declared "the integration branch", including a card worktree (REQ-AMC-014).
- **Headroom.** The original constant carried ~15 % over its baseline. Reusing that ratio keeps
  ordinary rule edits from tripping the guard while still catching a real regression; state the
  ratio in the constant's comment rather than picking a round number.
- **The constant must track the achieved figure, not merely clear the ceiling.** Setting it to
  75,000 while the surface lands at 60,000 passes "≤ 75,000" and ratchets nothing — the guard would
  then tolerate 15,000 tokens of silent regrowth. AC-AMC-019 closes this by checking the constant
  against `achieved × (1 + ratio)` within a stated tolerance.
- **The figure has to be printable.** `TestAlwaysLoadedTokenBudget` emits the total only via
  `t.Errorf` on failure, so a passing run prints nothing to quote. M5 adds the `t.Logf` first; every
  ratchet criterion then reads it under `go test -v`.

## 6. Rejected alternatives

| Alternative | Why rejected |
|---|---|
| Nested `AGENTS.md` per area (the card's proposal) | Measured: unloaded at repo-root invocation, and shares rather than expands the budget. Adopting it would silently drop most of the ruleset in ordinary use. |
| Raise `project_doc_max_bytes` as a substitute for the diet | Project scope **does** work — but only once the user registers the project `trust_level = "trusted"`, and it is ignored silently until then. The distributed user's first session is untrusted, so the binding cap there is 32,768 B (§2, `spec.md` §D.8). Conditional, retroactive, silent-on-failure — usable as headroom, never as a premise. |
| Generate `AGENTS.md` from the rules at build time | Adds a generator to maintain, and truncation is unaffected — a generated 200 KB file truncates exactly like a hand-written one. The problem is size, not authorship. |
| Symlink `AGENTS.md` → `CLAUDE.md` | Ships 20,523 B of Claude-specific orchestration to Codex and still leaves the rules unreachable. |
| Move `[HARD]` clauses into skills, loaded on demand | Forbidden by REQ-AMC-002. A rule that is not always present cannot bind every turn. |
| A second output-style pass here | Scope widening. Its first render-surface diet already landed (t131 / t142, −5,454 and −567 tokens, in `release/v3.1.1`); 46,765 B remains because that pass was a diet, not a removal — 75.8 % of the file is Claude render templates. Whether a second pass is needed is M1's call (`spec.md` §E.2). |
| Size the contract to fill 32,768 B | The budget is not a target. A contract at 99.3 % of budget begins truncating on its first edit, silently. |

## 7. Open questions carried into run phase

- The R1 detector assumes `[HARD]` marking is complete across the 14 files. A binding clause
  carrying neither `[HARD]` nor an imperative keyword would be missed. M1's manual pass over the
  pilot files should report whether it found any.
- The probe ran on macOS with `codex-cli` 0.147.0. A smaller upstream default on another platform
  or version would invalidate the ceiling's calibration silently — re-probe on a codex upgrade
  (`spec.md` §D.9).

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

**v0.3.6 — iteration 5 delta (FAIL 0.90, one finding).** F1: `REQ-AMC-010` / `AC-AMC-015` gained a
**narrow surface-cardinality carve-out**. Fixed slots append unconditionally, so `REQ-AMC-008`'s
fourth slot grows `len(surface)` before `AGENTS.md` is authored, breaking two hardcoded counts
verified in `internal/config/token_budget_guard_test.go` (`wantRuleCount + 3` → `+ 4`; temp-tree
`want 4` → `5`). Both exits previously failed a criterion; the exemption names those two assertions,
covers the expected count and comment only, and leaves every behavioral expectation bound. F2
(optional): Arm B is now baselined on the **integration-branch** figure recorded at pre-flight
(`plan.md` M1, `AC-AMC-007`) — the trees differ by 4,070 tokens (37 % of the cut), so a
worktree-baselined projection could clear Arm B and still fail `AC-AMC-018` at M5; the M1 block
quote quotes **15,055**. Traceability was 1.00 at iteration 5 and is unchanged (no REQ or AC added).

**v0.3.5 — dispatcher readability additions to the E1/E2 edits.** No requirement or criterion
changed. §C.4 now **explains** the net-additive mechanism rather than citing it: a clause authored
into `AGENTS.md` does not leave the always-loaded rules, because `REQ-AMC-002` forbids the move and
`REQ-AMC-001` independently requires the clause in `AGENTS.md` — so the surface grows by
`|AGENTS.md|` and the cut is `stated cut + |AGENTS.md|`. Written out because four readers in
sequence (card author, SPEC author, dispatcher, two audit iterations) each made the relocation
assumption working from the citation alone. M1's stop condition and `AC-AMC-007` now state that
returning a blocker with the measured shortfall is a **correct outcome** of the pilot, and that
Arm B should be expected to fire given the roughly doubled minimum cut.

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

### M1 — clause classification and compressibility (2026-08-22)

Tree: worktree `.claude/worktrees/t82`, branch `WT-agents-md-diet`, `58f0bdd43`.
Full report with every command and its output: `.moai/reports/t82/m1-pilot.md`.
M1 wrote no `AGENTS.md`, moved no rule text, and changed no Go code.

**Verdict: GO, conditional.** Both stop-condition arms clear. The condition is that `AGENTS.md`
lands near the measured projection (~11.9 KB) rather than at the 24,576 B ceiling — at the ceiling
Arm B's tightest bound falls 310 tokens short and M4 must add a second pass on an already-stubbed
file to close it (§5.3 of the report). `spec.md` §E.2's open question is answered: **no second
output-style pass is required**; neither path touches `.claude/output-styles/moai/moai.md`.

| AC | Claim | Evidence (command → observed) | Status |
|---|---|---|---|
| `AC-AMC-003` | Markers expanded to clause blocks; per-block table, byte per row, zero unclassified, Codex-relevant subtotal stated | `python3 .moai/reports/t82/clause-blocks.py > blocks.json` → 97 blocks / 51,639 B; `python3 .moai/reports/t82/classify_report.py --table` → 97 rows, C 35 / **16,135 B**, K 61 / 35,147 B, P 1 / 357 B, unclassified 0 (the script exits non-zero on any unclassified id) | PASS |
| `AC-AMC-004` | Rewritten clauses preserve subject, modality, binding scope | per-clause diff table, `.moai/reports/t82/m1-pilot.md` §3.1 — 11 of 11 preserved, 0 reverted | PASS |
| `AC-AMC-005` | Aggregate compression ratio stated with per-clause variance | `python3 .moai/reports/t82/pilot_measure.py` → 5,355 → 2,848 B, ratio **0.5318** (46.8 % reduction); min 0.400 / max 0.728 / mean 0.552 / median 0.504 / stdev 0.119 | PASS |
| `AC-AMC-006` | Ceiling re-derived against the clause-block figure; explicit reachable/not-reachable verdict with the projected figure | `python3 .moai/reports/t82/project.py` → 16,135 × 0.5318 = 8,581 B + 3,300 B structure (stated assumption) = **11,881 B** vs ceiling 24,576 → **REACHABLE**, 48 % of ceiling used. Ceiling retained at 24,576 B; break-even Codex-relevant verbatim is 40,004 B, 2.5× the measured 16,135 | PASS |
| `AC-AMC-007` | If either arm trips, a blocker naming the shortfall and the two levers, no file moved | Neither arm tripped. Arm A 11,881 ≤ 24,576. Arm B required cut 7,806 tok vs 10,670 tok available (+2,864). No file moved regardless | PASS (no blocker due) |

**Arm B detail.** Baseline is the guard's own arithmetic, not the byte total: `surface_r3.py`
reproduces `alwaysLoadedSurface()` + `measureAlwaysLoaded()` → `surface files: 17   guard tokens:
71207   surface bytes: 284850`. Contract layer is net-additive, so required cut =
`71,207 + |AGENTS.md|/4 − 66,371`: **7,806 tok** at the measured projection, **10,980 tok** at the
REQ-AMC-004 ceiling. Available: the conservative measured stub-split precedent (kanban-dispatch
21,003 → 13,027 = 38.0 %, commit `a203a7c3a`; the other precedent, goal-directive 25,755 → 6,531 =
74.6 %, commit `6422046bb`) applied to the nine always-loaded files never yet stub-split
(112,316 B) → 42,680 B = **10,670 tok**.

**Measurement corrections carried into `spec.md` (figures and provenance only).**

- §A.1: token figure `≈ 71,212` → **71,207**, with the per-file-floor provenance stated. The
  guard sums `len(file)/4` per file; `284,850 / 4` overstates by 5 tokens.
- §C.4: the `75,282 / 81,426 / 15,055` row retired. `.moai/reports/t82/preflight.md` measured
  `origin/main`, `release/v3.1.1`, `v3.1.2`, `v3.1.3` and this worktree — all five at 284,850 B —
  so that state is on no live branch. Required cut at the ceiling case is now **10,980**, with the
  v3.2 integration-branch re-measurement obligation (`REQ-AMC-014`, `AC-AMC-018`) restated
  explicitly rather than dropped.

**Finding — the line proxy errs almost entirely in one direction.** `spec.md` §A.4 disclosed
"15 of 93 markers are prose (overcount), 16 lead into uncounted bodies (undercount)". Read at
clause-block level: exactly **one** marker carries no obligation (`kanban-dispatch.md:7`, a detail-
companion navigation note). Nine of the ten structurally non-clause-initial markers are genuine
obligations sitting mid-sentence, and six markers sit on **headings** (`moai-constitution.md`
Agent Core Behaviors 1-6) whose whole section body is the obligation — 5,117 B counted by the proxy
as ~300 B of heading text. The proxy's net error is undercount by 58.7 % (32,543 → 51,639 B), not a
two-sided bracket.

**Finding — the six-file Claude-only proxy is accurate inside itself and incomplete outside.**
`design.md` §1.1's six files hold 22,179 B of clause blocks, 21,504 B (97.0 %) of it Claude-only —
so the 14,360 B line-level upper bound was nearly all usable. But **38.8 % of all Claude-only
bytes (13,643 B) sit outside those six files**, chiefly `kanban-dispatch.md` (5,899 B) and
`agent-common-protocol.md` (3,618 B). A file-membership classification would have been wrong in
both directions on `kanban-dispatch.md`, which splits 10 Codex-relevant / 12 Claude-only.

**Blocked / needs routing (not actioned by this agent).** Four surfaces still carry the retired
figures and lie outside run-phase authoring ownership:

| Surface | Stale content | Owner |
|---|---|---|
| `spec.md` REQ-AMC-014 body | "≈ 71,212 … measured 75,282" | requirement text — manager-spec |
| `plan.md` §B item 2, §E M1 Arm B block quote, §E M1 Arm B body | 71,212 / 75,282 / 15,055 / 10,985 / "differ by 4,070 tokens" | plan body — manager-spec |
| `acceptance.md` `AC-AMC-007` note | "71,212 worktree vs 75,282 integration", "37 % difference" | criterion text — manager-spec |
| `design.md` §branch sensitivity | 71,212 / 75,282 | design body — manager-spec |

Also unrouted: this M1 edit changed `spec.md` §A.1 / §C.4 body figures, which the SPEC's own
HISTORY provenance rule says must appear as a HISTORY row with a matching `version:` bump.
`version:` is outside this agent's frontmatter permissions (`status:` / `updated:` only), so the
row and the bump are left to the committing actor. HISTORY rows 0.3.4 / 0.3.6 and `§E.1` above
are **not** rewritten — they record what those revisions contained at the time.

**Gaps.** Document structure (3,300 B) is a stated assumption, not a measurement — M2 measures it;
doubling it still clears Arm A (15,181 B) but adds 825 tok to Arm B's cut. Arm B's 38 % is a
precedent transplanted onto files it was not measured on. R2 (10,023 B over rules + `CLAUDE.md`) is
a line-level proxy of the same family as the R1 proxy this milestone corrected, and the R3 movable
figure (161,482 B) inherits its error. The v3.2 integration branch does not exist, so
`AC-AMC-018`'s baseline is still unmeasured. `char/4` carries ±15 % against a real tokenizer.

**Residual risk.** Re-inflation is measured, not hypothetical: `kanban-dispatch.md` was dieted to
13,027 B on 2026-08-17 and measures **25,915 B** today — larger than before its diet. Arm B's
2,864-token margin is one such event wide.

**Verification.**

```
go test -count=1 ./internal/config/ -run 'Budget|AlwaysLoaded' -v
--- PASS: TestAlwaysLoadedTokenBudget (0.00s)
--- PASS: TestAlwaysLoadedTokenBudget_OverBudgetFails (0.00s)
--- PASS: TestAlwaysLoadedSurfaceEnumeration (0.01s)
--- PASS: TestMeasureAlwaysLoaded_WithMemory (0.00s)
--- PASS: TestWorkflowYAMLUnmarshal_OmittedTokenBudget_PreservesDefaults (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/config	0.637s
```

`alwaysLoadedSurface()` unextended (M5), no `t.Logf` added (M5),
`AlwaysLoadedTokenBudget` unchanged at 76,000.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

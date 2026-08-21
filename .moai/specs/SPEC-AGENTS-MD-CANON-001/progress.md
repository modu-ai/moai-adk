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

### M2 — root `AGENTS.md` contract layer (2026-08-22)

Tree: worktree `.claude/worktrees/t82`, branch `WT-agents-md-diet`, `57a7ef71b`.
Files written: `AGENTS.md` (new), `.gitignore` (one ignore rule removed — see the blocker note
below). `CLAUDE.md`, `internal/config/**`, and `internal/template/templates/**` untouched (M3 / M5 /
M6). No nested `AGENTS.md` created.

**Landed size: 14,229 B** — `wc -c AGENTS.md` → `14229`. Against `REQ-AMC-004`'s 24,576 B ceiling:
**10,347 B of headroom**, 57.9 % of the ceiling used. Against the measured 32,768 B codex budget:
**18,539 B of headroom** for a user's global `~/.codex/AGENTS.md` layer plus future growth.

| AC | Claim | Evidence (command → observed) | Status |
|---|---|---|---|
| `AC-AMC-008` | Every Codex-relevant clause block appears in the root document; trace table maps clause → origin with zero unmapped rows | `python3 .moai/reports/t82/trace.py` → `mapped rows: 36 | C blocks: 35 | unmapped C: none | promoted from K: ['B042']`; `python3 .moai/reports/t82/presence.py` → `clauses checked: 36 | missing: none`, exit 0 | PASS |
| `AC-AMC-009` (live file) | Live root `AGENTS.md` at or below 24,576 B, headroom against 32,768 B stated | `wc -c AGENTS.md` → `14229`; headroom 10,347 B vs ceiling, 18,539 B vs budget | PASS |
| `AC-AMC-009` (mirror) | Template mirror at or below 24,576 B | `internal/template/templates/AGENTS.md` does not exist — M6 creates it | PENDING (M6) |
| `AC-AMC-009` (identity) | `required cut ≤ available reduction` holds at the landed size | required cut = 71,207 + 14,229 ÷ 4 − 66,371 = **8,393 tok**; available reduction (nine never-stub-split files at the measured 38.0 % precedent) = **10,670 tok** → margin **+2,277**, on the first term alone | PASS |
| `AC-AMC-010` | Exactly one `AGENTS.md`, at the repository root | `git ls-files --full-name --cached --others --exclude-standard ':(top)*AGENTS.md' ':(exclude,top)internal/template/templates/*'` → `AGENTS.md`, one line. Re-run after the M2 commit `fd3ac06a8` with the criterion's own command (`git ls-files --full-name ':(top)*AGENTS.md' ':(exclude,top)internal/template/templates/*'`) → `AGENTS.md`, one line | PASS |

**Baseline attribution.** Always-loaded surface re-measured on this tree after the write:
`python3 .moai/reports/t82/surface_r3.py` → `surface files: 17   guard tokens (sum of per-file
len/4): 71207   surface bytes: 284850` — unchanged from M1, because `alwaysLoadedSurface()` does not
yet enumerate `AGENTS.md` (`REQ-AMC-008` / `REQ-AMC-013` ¶2 order that extension into M5, before any
measurement quoted as a ratchet basis). The required cut above therefore adds `|AGENTS.md| ÷ 4`
explicitly rather than reading it off the guard.

**Go regression check** (no Go code changed; run to confirm the new root file does not perturb the
guard):

```
go test -count=1 ./internal/config/ -run 'Budget|AlwaysLoaded' -v
--- PASS: TestAlwaysLoadedTokenBudget (0.01s)
--- PASS: TestAlwaysLoadedTokenBudget_OverBudgetFails (0.00s)
--- PASS: TestAlwaysLoadedSurfaceEnumeration (0.00s)
--- PASS: TestMeasureAlwaysLoaded_WithMemory (0.00s)
--- PASS: TestWorkflowYAMLUnmarshal_OmittedTokenBudget_PreservesDefaults (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/config	0.538s
```

**Classification movement against M1: one clause moved, both directions accounted.**

- **Promoted K → C: `B042`** (`cache-aware-execution.md`:25, "weigh session length as a cost axis",
  419 B). M1 filed it Claude-only on the `/clear` framing, but the obligation itself — one long
  session over several short ones, because a fresh session re-pays the always-loaded prefix at write
  price — rests on prompt-prefix caching, which codex has. Its two neighbouring directives (`B040`,
  `B041`) were already C. `spec.md` §D.2 directs doubt to the Codex side; the cost is 419 B of
  verbatim input.
- **Demoted C → K: none.** Demotion is the silent-loss direction (`spec.md` §D.2), so no clause was
  moved that way; all 35 of M1's C blocks are in the document.
- **Not a promotion, but worth naming: `B011`.** `B012` (C) binds a direct edit to a shared path to
  "the parallel-session detection", and the concrete check lives in `B011`, which is K because it
  gates `Agent()` spawns. Carrying `B012` without the check would have shipped a clause pointing at
  nothing — the self-sufficiency failure `REQ-AMC-001` forbids. `AGENTS.md` §2 therefore states the
  two-command divergence check as part of `B012`'s own obligation. `B011`'s spawn-gate obligation is
  not carried.

**Structure assumption discharged** (M1 report §7 Gap 1). `python3 .moai/reports/t82/structure.py` →
`structure total (front matter + headings + rules + blank lines): 1179`, `clause text: 13050`,
`M1 assumption for structure: 3300 -> delta -2121`. The assumption ran **2,121 B high**, in the safe
direction.

**Compression ran looser than the pilot ratio, and the projection's use of that ratio was
optimistic.** Clause text landed at 13,050 B against 16,554 B of mapped verbatim input — an
aggregate ratio of **0.788**, not the pilot's 0.5318. Three reasons, all visible in the input rather
than in the authoring: the pilot sampled `kanban-dispatch.md`, the most rationale-dense file in the
set, where the removable layer is largest; the six `moai-constitution.md` core behaviors (5,117 B,
31 % of the input) are almost entirely obligation text with little narrative to strip; and `B009`'s
clause block is a bare heading line whose body the extractor did not capture, so making it
self-sufficient *added* bytes rather than removing them. The net still lands 2,348 B below M1's
11,881 B projection because the structure assumption over-ran by more than compression under-ran.

**Trace table — 36 clauses, zero unmapped.** Regenerate with `python3 .moai/reports/t82/trace.py`
(the table is also written to `.moai/reports/t82/trace.md`).

| Clause | Class | Origin file | Line | Verbatim B | `AGENTS.md` section |
|---|---|---|---:|---:|---|
| `B036` | C | `.claude/rules/moai/core/verification-claim-integrity.md` | 9 | 283 | §1 Evidence and verification claims |
| `B037` | C | `.claude/rules/moai/core/verification-claim-integrity.md` | 37 | 165 | §1 Evidence and verification claims |
| `B038` | C | `.claude/rules/moai/core/verification-claim-integrity.md` | 50 | 385 | §1 Evidence and verification claims |
| `B012` | C | `.claude/rules/moai/core/agent-common-protocol.md` | 296 | 412 | §2 Git, branches, and the shared checkout |
| `B013` | C | `.claude/rules/moai/core/agent-common-protocol.md` | 330 | 436 | §2 Git, branches, and the shared checkout |
| `B076` | C | `.claude/rules/moai/workflow/main-checkout-branch-guard.md` | 20 | 704 | §2 Git, branches, and the shared checkout |
| `B077` | C | `.claude/rules/moai/workflow/main-checkout-branch-guard.md` | 58 | 275 | §2 Git, branches, and the shared checkout |
| `B066` | C | `.claude/rules/moai/workflow/kanban-dispatch.md` | 127 | 676 | §3 Worktrees |
| `B067` | C | `.claude/rules/moai/workflow/kanban-dispatch.md` | 139 | 454 | §3 Worktrees |
| `B068` | C | `.claude/rules/moai/workflow/kanban-dispatch.md` | 141 | 246 | §3 Worktrees |
| `B069` | C | `.claude/rules/moai/workflow/kanban-dispatch.md` | 143 | 693 | §3 Worktrees |
| `B070` | C | `.claude/rules/moai/workflow/kanban-dispatch.md` | 145 | 550 | §3 Worktrees |
| `B071` | C | `.claude/rules/moai/workflow/kanban-dispatch.md` | 159 | 513 | §3 Worktrees |
| `B010` | C | `.claude/rules/moai/core/agent-common-protocol.md` | 227 | 317 | §4 How verification is run |
| `B064` | C | `.claude/rules/moai/workflow/kanban-dispatch.md` | 98 | 679 | §4 How verification is run |
| `B072` | C | `.claude/rules/moai/workflow/kanban-dispatch.md` | 171 | 488 | §4 How verification is run |
| `B073` | C | `.claude/rules/moai/workflow/kanban-dispatch.md` | 173 | 379 | §4 How verification is run |
| `B074` | C | `.claude/rules/moai/workflow/kanban-dispatch.md` | 179 | 254 | §4 How verification is run |
| `B028` | C | `.claude/rules/moai/core/moai-constitution.md` | 179 | 573 | §5 Core behaviors |
| `B029` | C | `.claude/rules/moai/core/moai-constitution.md` | 195 | 462 | §5 Core behaviors |
| `B030` | C | `.claude/rules/moai/core/moai-constitution.md` | 207 | 640 | §5 Core behaviors |
| `B031` | C | `.claude/rules/moai/core/moai-constitution.md` | 224 | 2245 | §5 Core behaviors |
| `B032` | C | `.claude/rules/moai/core/moai-constitution.md` | 253 | 770 | §5 Core behaviors |
| `B033` | C | `.claude/rules/moai/core/moai-constitution.md` | 270 | 750 | §5 Core behaviors |
| `B002` | C | `.claude/rules/moai/core/agent-common-protocol.md` | 100 | 99 | §6 Output, language, and format |
| `B003` | C | `.claude/rules/moai/core/agent-common-protocol.md` | 110 | 102 | §6 Output, language, and format |
| `B004` | C | `.claude/rules/moai/core/agent-common-protocol.md` | 112 | 228 | §6 Output, language, and format |
| `B014` | C | `.claude/rules/moai/core/agent-common-protocol.md` | 336 | 231 | §6 Output, language, and format |
| `B025` | C | `.claude/rules/moai/core/moai-constitution.md` | 22 | 465 | §6 Output, language, and format |
| `B034` | C | `.claude/rules/moai/core/native-idiom-and-register.md` | 8 | 423 | §6 Output, language, and format |
| `B095` | C | `CLAUDE.md` | 98 | 393 | §6 Output, language, and format |
| `B005` | C | `.claude/rules/moai/core/agent-common-protocol.md` | 130 | 68 | §7 Tools and command output |
| `B009` | C | `.claude/rules/moai/core/agent-common-protocol.md` | 168 | 102 | §7 Tools and command output |
| `B040` | C | `.claude/rules/moai/workflow/cache-aware-execution.md` | 21 | 361 | §7 Tools and command output |
| `B041` | C | `.claude/rules/moai/workflow/cache-aware-execution.md` | 23 | 314 | §7 Tools and command output |
| `B042` | K→C | `.claude/rules/moai/workflow/cache-aware-execution.md` | 25 | 419 | §7 Tools and command output |

`B025` and `B034` state the same native-idiom obligation in two files and are carried by one
`AGENTS.md` clause; both rows map to it, so the trace has no unmapped origin.

**Ordering is load-bearing and was chosen for it.** Truncation takes the tail silently
(`spec.md` §A.2 finding 2 / finding 6), so sections run most-critical-first: evidence integrity,
then the failure modes that destroy work irrecoverably (shared-checkout branch state, worktree
disposal), then verification practice, then the core behaviors, then output and tool discipline. The
front matter carries the `~/.codex/AGENTS.md` warning at the top, where it survives any truncation
that would remove the clauses it warns about.

**Blocker encountered and resolved inside scope: `AGENTS.md` was gitignored.** `.gitignore:128`
carried a bare `AGENTS.md` under "Cross-tool artifacts (Codex CLI — local-only, not distributed)",
added `16b5a6c9a` (2026-06-11) when the file was a local scratch artifact. `git check-ignore -v
AGENTS.md` → `.gitignore:128:AGENTS.md	AGENTS.md`. The pattern has no leading slash, so it also
matched `internal/template/templates/AGENTS.md`. While it stood, `AC-AMC-010` returned empty rather
than one line and `REQ-AMC-015`'s mirror could not be tracked — both unsatisfiable, silently. The
rule was removed (a comment recording why it must stay removed replaces it); `.codex/` and
`.agents/` remain ignored. No SPEC artifact anticipated this, and no milestone owns it — recorded
below as a scope-doc gap rather than assumed.

**Gaps (not observed).**

1. **Mirror leg of `AC-AMC-009` unevaluated** — `internal/template/templates/AGENTS.md` does not
   exist until M6, so only the live file was measured.
2. **`AC-AMC-010`'s literal command returns empty** on this tree because the file is untracked; the
   equivalence to "exactly one root `AGENTS.md`" was established with `--others
   --exclude-standard`. The criterion becomes literally satisfiable at commit.
3. **Obligation preservation is asserted per clause, not mechanically proved.** `presence.py`
   checks that a distinctive marker of each clause appears; it cannot check that subject, modality,
   and scope survived. That reading is the trace table's, and it is a judgement.
4. **No codex run against the authored file.** The document was measured, not loaded — nothing here
   observes codex parsing 14,229 B at the repo root.
5. **Nothing measured about M4's reachability beyond the identity.** The 10,670-token available
   reduction is M1's projection from a two-point precedent, unchanged by M2.

**Residual risk.**

- The 0.788 aggregate compression ratio means a future clause set with the same shape costs more
  per clause than `spec.md` §D.1's projection assumes. The ceiling absorbs it here; a re-derivation
  quoting 0.5318 for a different clause population would not be sound.
- Classification drift is unchanged by M2 and stays the SPEC's open exposure: a new `[HARD]` clause
  added to a rule file reaches `AGENTS.md` only if someone classifies it, and M5's byte guard
  measures size, never completeness.
- The `.gitignore` removal is one line, but it changes what the repository tracks. Any tooling that
  assumed a root `AGENTS.md` could not be committed now sees one.

### M3 — `CLAUDE.md` → `AGENTS.md` import layer (2026-08-22)

Tree: worktree `.claude/worktrees/t82`, branch `WT-agents-md-diet`, parent `e7484034a`.
Files written: `CLAUDE.md` only. `AGENTS.md`, `internal/config/**`, and
`internal/template/templates/**` untouched (M2 landed / M5 / M6).

**Two edits, both in `CLAUDE.md`.**

1. **§0 Standing Contract (imported)** — a new leading section carrying `@AGENTS.md`, placed above
   §1 so the contract is assembled first. Section numbering `§1`-`§17` is unchanged, so every
   existing cross-reference still resolves. The import uses the same repo-root-relative `@` form
   §9 already uses for `.moai/config/sections/*.yaml`, satisfying `REQ-AMC-011`.
2. **`B095` de-duplicated** — the `[ZONE:Evolvable] [HARD]` native-UTF-8 payload clause at §8 was
   the *only* Codex-relevant (class C) clause block originating in `CLAUDE.md`; M1 classified
   `B093` / `B094` / `B096` as Claude-only. Its inline obligation text is replaced by a
   non-obligation pointer to `AGENTS.md` §6 plus the unchanged SSOT reference, so §8 retains
   exactly the Claude-mechanism layer `REQ-AMC-011` describes and the obligation is carried once,
   through the import, rather than inline and again imported.

**Size: 20,523 B → 20,748 B (+225).** The §0 block costs more than `B095`'s removal recovers; the
net is +57 guard tokens on the always-loaded surface.

| AC | Claim | Evidence (command → observed) | Status |
|---|---|---|---|
| `AC-AMC-013` | No contract line appears in both the Claude-only layer and the imported `AGENTS.md` | `cat CLAUDE.md AGENTS.md \| grep -n '\[HARD\]' \| sort -k2 \| uniq -d -f1` → no output. `grep -c '\[HARD\]' CLAUDE.md` → `3` (`B093`, `B094`, `B096` — all class K); `AGENTS.md` carries no `[HARD]` marker at all | PASS |
| `AC-AMC-014` | Every `@`-imported path resolves; an unresolvable import fails rather than warns | Each `^@` line in `CLAUDE.md` resolved with `os.path.isfile` → `OK line 9 AGENTS.md`, `OK line 118 .moai/config/sections/user.yaml`, `OK line 119 .moai/config/sections/language.yaml`. Three of three | PASS |
| `AC-AMC-015` | Affected-package suites green, no expected-behavior assertion edited | `go test -count=1 ./internal/config/` → `ok 13.297s`; `./internal/constitution/` → `ok 1.750s`; `./internal/hook/` → `ok 37.561s`; `./internal/template/` → `ok 25.574s`; `./internal/cli/ -run 'Constitution\|Doctor\|Instruction\|Claude'` → `ok 47.058s`. No `_test.go` file was modified in this milestone (`git diff --name-only` names `CLAUDE.md` alone) | PASS |
| `AC-AMC-010` (Claude-side non-regression) | Rule-loading semantics and hook wiring unchanged | No `.claude/rules/**`, `.claude/hooks/**`, or `settings.json` edit; the five suites above cover the guard, the constitution validator, the hook handlers, and the template deployer | PASS |

**Baseline attribution.** `python3 .moai/reports/t82/surface_r3.py` → `surface files: 17   guard
tokens (sum of per-file len/4): 71264   surface bytes: 285075`. Up 57 tokens from M2's 71,207,
entirely from `CLAUDE.md`'s +225 B. `AGENTS.md` is still **not** in the enumeration — `REQ-AMC-008`
orders that extension into M5.

**Arm B identity re-evaluated on the M3 surface:**

```
required cut = 71,264 + 14,229 ÷ 4 − 66,371 = 8,450 tok
available    = 10,670 tok  (nine never-stub-split files, 38.0 % measured precedent)
margin       = +2,220 tok  (first term alone; M2 read +2,277)
```

**Gap — deliberate, and M6 closes it: the template mirror now diverges.** At `e7484034a` the live
`CLAUDE.md` and `internal/template/templates/CLAUDE.md` were byte-identical (both 20,523 B, `diff`
clean). This milestone edits the live file only. Mirroring the `@AGENTS.md` import *now* would ship
an import with no target, because `internal/template/templates/AGENTS.md` does not exist until M6 —
precisely the unresolvable-import failure `REQ-AMC-012` and `AC-AMC-014` classify as a failing
criterion, and on a user's machine nothing would signal it. So the mirror waits for M6, which lands
both files in one step.

The divergence is **not** CI-visible: `go test ./internal/template/` passes with it present, and no
test asserts byte-parity between the live and template `CLAUDE.md` (searched
`internal/template/*_test.go` for `templates/CLAUDE.md` — the three hits are the
`MOAI:LEARNED-WORKFLOW` marker check, not a parity assertion). Two consequences follow, both
recorded rather than mitigated here:

- **M6 must mirror `CLAUDE.md` as well as create `AGENTS.md`.** A mirror pass that copies only
  `AGENTS.md` leaves users without the import that reaches it.
- **`moai update` run against this dev tree before M6 would revert the §0 edit**, since `CLAUDE.md`
  is a template-deployed file. Not a code defect; a sequencing hazard for whoever runs update on
  this worktree.

**Residual risk.** `@`-import resolution was verified by file existence, not by observing Claude
Code assemble the file — the mechanism is evidenced by the two `.moai/config/sections/*.yaml`
imports that resolve in this repo today (`plan.md` M3), and `AGENTS.md` sits at the repository root
alongside `CLAUDE.md`, the shallowest possible path. A runtime confirmation belongs to whichever
session next starts with this tree loaded.

### M5-a — enumeration extension + guard observability (2026-08-22)

**Landed out of milestone order, deliberately.** `REQ-AMC-013` orders the `REQ-AMC-008`
enumeration extension **before any measurement cited as a ratchet basis** — "not merely before the
ratchet's final commit". M4 relocates clauses out of the rule files and quotes per-file
before/after figures; every one of those is a ratchet-basis measurement. Taken with `AGENTS.md`
unenumerated they would record reductions that did not happen — clauses moved into a file that is
still always-loaded but no longer counted — and `REQ-AMC-013` states the gap does not close
retroactively: "measurements taken inside it stay wrong and stay quotable". So this slice of M5
lands **between M3's suite check and M4's first citation**, which is the window the run-phase
handoff named. The rest of M5 (Codex-cap byte guard `AC-AMC-016(b)`, the ratchet constant
`AC-AMC-018` / `AC-AMC-019`) stays after M4 and M6, because it must be measured post-diet on the
integration branch.

Tree: worktree `.claude/worktrees/t82`, branch `WT-agents-md-diet`, parent `a2c919792`.
Files written: `internal/config/token_budget_guard.go`, `internal/config/token_budget_guard_test.go`.

**Three changes.**

1. **Fourth fixed slot.** `alwaysLoadedSurface()` now enumerates the root `AGENTS.md`, placed
   immediately after `CLAUDE.md` because it is that file's `@`-import. The existing hermetic
   treatment is reused unchanged — a slot absent from disk measures 0 tokens — so a tree without
   `AGENTS.md` keeps its previous baseline and the enumeration stays a single path
   (`REQ-AMC-008`: no second, independently-drifting measurement).
2. **Passing-path observability.** `TestAlwaysLoadedTokenBudget` emitted the token total only
   through `t.Errorf` on failure, so a passing run printed `ok …` and nothing else — which makes
   every "quote the achieved figure" criterion unsatisfiable. A `t.Logf` now emits the figure
   ahead of the over-budget check. This is the line `AC-AMC-018` reads.
3. **`AC-AMC-017` membership assertion.** `TestAlwaysLoadedSurfaceEnumeration` now asserts the
   returned surface contains the root `AGENTS.md` by path, not merely by count.

| AC | Claim | Evidence (command → observed) | Status |
|---|---|---|---|
| `AC-AMC-017` | The enumeration helper contains the root `AGENTS.md`; the byte guard reuses it rather than re-globbing | `go test -v ./internal/config/ -run 'Enumeration'` → `--- PASS: TestAlwaysLoadedSurfaceEnumeration`. **Falsified before accepting**: with the slot removed the same run fails twice — `surface has 17 entries, want 18 (= 14 no-paths: rules + 4 fixed surfaces)` and `always-loaded surface omits …/AGENTS.md; the Codex contract layer must be enumerated (REQ-AMC-008)`; slot restored, both pass again | PASS |
| `AC-AMC-016` (a) | Token-budget negative path already covered; no second fixture written | `go test -v ./internal/config/ -run 'OverBudgetFails'` → `--- PASS: TestAlwaysLoadedTokenBudget_OverBudgetFails`. No new fixture authored | PASS (cited, not rebuilt) |
| `AC-AMC-016` (b) | Codex-cap byte-guard breach fails the build and names the figure + file | Not implemented in this slice | PENDING (M5 proper) |
| `AC-AMC-015` | Suites green; only the two exempt cardinality assertions edited | `go vet ./internal/config/` clean; `go test -count=1 ./internal/config/` → `ok 3.836s`; `go test ./internal/harness/...` → all `ok`. Edits to `token_budget_guard_test.go`: `wantRuleCount + 3` → `+ 4`, temp-tree `want 4` → `5` (both named by the `AC-AMC-015` carve-out), plus two **new** assertions (the `t.Logf` line and the `AC-AMC-017` membership check). No pre-existing expected-behavior assertion changed | PASS |

**Achieved figure, now measured over the extended enumeration:**

```
go test -v ./internal/config/ -run 'Budget|AlwaysLoaded'
  token_budget_guard_test.go:69: always-loaded surface = 74821 tokens
                                 (budget 76000, headroom 1179, 18 entries)
```

**This figure supersedes every earlier surface reading in this SPEC's record.** M2's 71,207 and
M3's 71,264 were measured by `.moai/reports/t82/surface_r3.py` over the **unextended** 17-file
enumeration and had `|AGENTS.md| ÷ 4` added by hand; the guard now counts it directly. The two
agree exactly — 71,264 + 3,557 = 74,821 — which is the cross-check that the hand-addition was
right, not a second measurement path. From here the guard's logged line is the only figure to
quote.

**Arm B identity, restated on the guard's own reading:**

```
required cut = 74,821 − 66,371 = 8,450 tok
available    = 10,670 tok  (nine never-stub-split files, 38.0 % measured precedent)
margin       = +2,220 tok
```

Identical to M3's figure, reached without the manual addition.

**Residual risk — headroom is now 1,179 tokens, not 4,793.** The pre-`AGENTS.md` guard read 71,264
against the 76,000 constant; enumerating the contract layer consumes 3,557 of that headroom. The
guard still passes, but the margin is thin enough that a sibling card adding a mid-sized
always-loaded rule on the integration branch would breach it **before** M4's diet lands. That is
the guard working as designed — the surface really did grow — but whoever integrates this card
should expect the constant to be under pressure until M4 completes, and should not read a breach
in that window as a defect in this slice.

**Gap.** The byte-level Codex-cap guard (`REQ-AMC-007` / `REQ-AMC-009` / `AC-AMC-016(b)`) is not
implemented here; the ceiling is currently enforced by measurement in the progress record, not by
a failing build. `AlwaysLoadedTokenBudget` is untouched at 76,000 — the ratchet is `AC-AMC-018`'s,
measured post-diet on the integration branch.

### M4 — detail relocation to lazy companions (2026-08-22)

Tree: worktree `.claude/worktrees/t82`, branch `WT-agents-md-diet`, parent `af954cc0c`.

**Eight new lazy companions, each `paths:`-scoped so it leaves the always-loaded surface**, plus a
second pass into two companions that already existed. Every stub keeps its obligations and leaves a
pointer; only rationale, procedure, worked examples, incident records, and long cross-reference
tables moved.

| Source file | Before | After | Delta | Companion created |
|---|---:|---:|---:|---|
| `main-checkout-branch-guard.md` | 11,865 | 6,395 | **−5,470** | `-detail.md` |
| `moai-mcp-tools.md` | 7,357 | 2,389 | **−4,968** | `-catalogue.md` |
| `verification-claim-integrity.md` | 13,140 | 8,224 | **−4,916** | `-detail.md` |
| `cross-session-messaging.md` | 16,672 | 11,823 | **−4,849** | `-detail.md` |
| `context-window-management.md` | 13,009 | 8,828 | **−4,181** | `-detail.md` |
| `moai-constitution.md` | 18,958 | 15,433 | **−3,525** | `-detail.md` |
| `agent-common-protocol.md` | 27,043 | 24,645 | −2,398 | (existing `-reference.md`) |
| `askuser-protocol.md` | 23,504 | 21,822 | −1,682 | (existing `-reference.md`) |
| `native-idiom-and-register.md` | 4,967 | 3,952 | −1,015 | `-detail.md` |
| `CLAUDE.md` | 20,748 | 19,766 | −982 | (compressed in place) |
| `skill-routing.md` | 5,825 | 5,595 | −230 | `-detail.md` |
| **Total** | **163,088** | **128,872** | **−34,216 B** | 8 new |

**Achieved figure — the ratchet target is met.**

```
go test -v ./internal/config/ -run 'Budget|AlwaysLoaded'
  token_budget_guard_test.go:69: always-loaded surface = 66266 tokens
                                 (budget 76000, headroom 9734, 18 entries)
```

Measured over the extended enumeration that M5-a landed, so `AGENTS.md` is counted, not invisible.
M5-a read 74,821; the diet removed **8,555 tokens**, which agrees with the per-file byte total above
(34,216 ÷ 4 = 8,554, one token of rounding across eleven files).

**The budget identity closes, and both of its terms were needed.**

```
required cut = 74,821 − 66,371 = 8,450 tok
achieved     = 8,555 tok  →  margin +105

  term 1 — nine never-stub-split files:  30,136 B = 7,534 tok
           (70.6 % of the 10,670-token bound M1 projected at a 38.0 % ratio)
  term 2 — second pass into already-split files:  4,080 B = 1,020 tok
           (6.0 % of their 68,115 B remainder — above the 5 % the identity assumed)
```

Term 1 alone would have fallen **916 tokens short** of the required cut. `AC-AMC-009`'s identity
names the second term "part of the identity rather than a footnote"; this run is the case that
proves it — the nine files did not carry the diet on their own, because the 38.0 % pilot ratio was
measured on `kanban-dispatch.md`, the most rationale-dense file in the set, and the never-split nine
are more obligation-dense than that sample.

| AC | Claim | Evidence (command → observed) | Status |
|---|---|---|---|
| `AC-AMC-011` | No relocated obligation in any created companion | Every new companion grepped for `[HARD]`, `MUST`, `MUST NOT`, `shall` → **0 hits in all 8**. The two appends into existing companions were audited the same way: the Blind Spot Pass append is clean, and an initial Preview-Field-Standards append that carried a duplicate `[HARD]` line was **removed** — `askuser-protocol-reference.md` already owned that section, so the append was redundant as well as non-compliant | PASS |
| `AC-AMC-002` | No obligation left the always-loaded surface | Each companion is `paths:`-scoped to its stub, so it is off the surface; every clause moved was rationale, procedure, a worked example, an incident record, or a cross-reference table. Obligations that appeared in a moved block were left in the stub rather than carried (the two Opus `[HARD]` principles, the `[HARD]` preview single-select clause) | PASS |
| `AC-AMC-010` | Claude-side rule-loading semantics unchanged | Surface entry count unchanged at 18 — the eight companions are `paths:`-scoped and never join it. `go test ./internal/config/` → `ok 17.739s`; `./internal/constitution/` → `ok 0.486s`; `./internal/hook/` → `ok 30.409s`; `./internal/template/` → `ok 35.697s`; `./internal/cli/` → `ok 334.733s` | PASS |
| `AC-AMC-015` | Suites green, no expected-behavior assertion edited | No `_test.go` file modified in this milestone | PASS |

**Template mirrors: landed here, not deferred.** `TestSanitizedPairParity` went red as soon as the
first two stubs were cut — correctly: it exists to catch a doctrine change that fails to reach the
distribution mirror. Four registry members drifted (`main-checkout-branch-guard.md`,
`verification-claim-integrity.md`, `askuser-protocol.md`, `agent-common-protocol.md`). All sixteen
rule files plus all eight new companions are now mirrored under
`internal/template/templates/.claude/rules/moai/`, `make build` has run, and the suite is green.

Two mirror treatments, chosen per file rather than uniformly:

- **Byte-copy (8 files + 8 companions).** Verified first that each template copy was byte-identical
  to its pre-edit local copy (`git show HEAD:<local>` vs the template file) — for eight of the ten
  edited rules it was, so a copy preserves no sanitization that did not exist. `moai update` cannot
  regress them, because both sides now match.
- **Structural re-application (2 files + 2 companions).** `verification-claim-integrity.md` and
  `main-checkout-branch-guard.md` are genuinely sanitized pairs: the local copies retain SPEC-IDs
  and REQ tokens the mirrors strip. Their template copies received the *same section removals and
  the same stub text*, re-sanitized — never a copy. Their two new companions were likewise
  sanitized on the way out (an originating SPEC-ID generalized to "the originating brain SPEC", the
  branch-guard Origin block and its REQ ranges dropped).

**Gap — `CLAUDE.md`'s mirror is still deferred, and is now two changes behind.** M3 deferred it
because the `@AGENTS.md` import would dangle without the template `AGENTS.md`; M4 adds the §4/§7/§15
compression on top. Both land together in M6. This is unchanged in kind from the M3 note, only
larger — and still not CI-visible, since no test asserts live/template `CLAUDE.md` parity.

**Residual risk — the margin is 105 tokens.** The achieved figure clears the ratchet target by
0.16 %. Any always-loaded rule added on the integration branch before `AC-AMC-018` is measured will
consume it. Two levers remain if that happens and both are already sized: the never-split nine
retain 29.4 % of their projected yield, and the already-split remainder was drawn down only 6.0 %.
Neither is exhausted, so a breach is recoverable without re-opening the design.

**Not done here.** `.claude/output-styles/moai/moai.md` is the single largest always-loaded item at
61,706 B (15,426 tokens, 23 % of the surface) and was left untouched: M1's available-reduction bound
excluded it, and M4's scope is the always-loaded **rules**. It is the obvious next lever if the
ratchet is ever raised again, and naming it here is cheaper than rediscovering it.

### M6 — template mirror, neutrality, and the global-layer warning (2026-08-22)

Tree: worktree `.claude/worktrees/t82`, branch `WT-agents-md-diet`, parent `243eb07ef`.
Files written: `internal/template/templates/AGENTS.md` (new),
`internal/template/templates/CLAUDE.md` (the two deferred changes, landed together).

**The deferred mirror closes here.** M3 held `CLAUDE.md`'s mirror back because `@AGENTS.md` would
have dangled without a template `AGENTS.md`; M4 added its compression on top. Both land in this
step, in the order that keeps the import resolvable at every point a user could observe it: the
template `AGENTS.md` is created first, then `CLAUDE.md` is mirrored on top of it.

M4 had already mirrored the twenty rule files, so this milestone's mirror work is the two root
documents. `make build` re-embedded the template FS afterwards.

| AC | Claim | Evidence (command → observed) | Status |
|---|---|---|---|
| `AC-AMC-021` | Every shipped file this SPEC lands has a template mirror; `make build` ran after the last mirror edit | 22 shipped-surface files enumerated from `git diff --name-only 57a7ef71b..HEAD` plus `git status --porcelain`, each checked against `internal/template/templates/<path>` → **22 MIRROR OK, missing: none**. `make build` → `catalog.yaml updated successfully (12899 bytes)` + `go build … -o bin/moai` | PASS |
| `AC-AMC-022` | No SPEC IDs, REQ tokens, audit citations, internal dates, commit SHAs, macOS-biased paths, or `CLAUDE.local.md` references in any template copy | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run 'Leak\|Neutrality' -v` → 10 tests, all PASS, incl. `TestTemplateNoInternalContentLeak`, `TestLeakClassNoDateShaInDefaultTier`, `TestLeakClassReqTokenPartition`. Independent regex scan of `AGENTS.md` and `CLAUDE.md` for SPEC-ID / REQ / AC / date / SHA / `/Users/` / `CLAUDE.local` → **clean, both** | PASS |
| `AC-AMC-023` | The shipped documentation warns about `~/.codex/AGENTS.md` and states all three facts | `grep -n 'codex/AGENTS.md' internal/template/templates/AGENTS.md` → line 7, the Budget warning: a personal `~/.codex/AGENTS.md` (1) **joins the same merged chain**, (2) is **consumed before** this file, (3) **narrows what the project's contract can carry**. Overflow is dropped from the tail silently, which is why the clauses are ordered most-critical-first | PASS |
| `AC-AMC-024` | The SPEC's cap-raise position is the measured one; the retired premise appears only inside an explicit correction | `grep -rn 'cannot ship' .moai/specs/SPEC-AGENTS-MD-CANON-001/` → 4 hits, **zero bare assertions**: `spec.md`:529 and `design.md`:110 both introduce it as an earlier draft's claim and mark it false; `research.md`:24 records it as corrected; `acceptance.md`:244 is the criterion itself. §D.8 states the measured position — project scope works only under `trust_level = "trusted"`, the untrusted first session at 32,768 B is binding, non-application is silent (stderr 0 bytes on all four probe runs) | PASS |
| `AC-AMC-009` (mirror) | Template mirror at or below 24,576 B, headroom against 32,768 B stated | `wc -c internal/template/templates/AGENTS.md` → `14229`, identical to the live file. **10,347 B of headroom** against the ceiling, **18,539 B** against the codex budget — the same figures as the live copy, because the mirror is byte-identical | PASS |
| `AC-AMC-010` | Exactly one live-tree `AGENTS.md`, at the repository root | The criterion's own command → `AGENTS.md`, one line. The template mirror is correctly excluded by the `:(exclude,top)` pathspec, and appears separately as an untracked `internal/template/templates/AGENTS.md` awaiting this commit | PASS |
| `AC-AMC-005` | No live-tree `AGENTS.md` outside the repository root | Same command; the only other copy on disk is the template mirror, which `REQ-AMC-005` exempts and `REQ-AMC-015` requires | PASS |

**Deployment verified end-to-end, not by file existence.** `moai init . --non-interactive` was run
with the freshly built binary into a scratch project outside the repo, and the deployed tree
inspected:

```
AGENTS.md   14,229 B   deployed
CLAUDE.md   19,766 B   deployed
@-imports resolved against the DEPLOYED project root:
  OK  line   9  AGENTS.md
  OK  line 118  .moai/config/sections/user.yaml
  OK  line 119  .moai/config/sections/language.yaml
companions deployed: 6 under rules/moai/core/, 6 under rules/moai/workflow/
```

**This closes M3's stated residual risk.** That milestone recorded that `@`-import resolution had
been verified by file existence rather than by observing a real assembly, and deferred a runtime
confirmation to "whichever session next starts with this tree loaded". A deployed project is the
stronger check and the one that actually matters: it exercises the path a distributed user takes,
where a dangling import would be silent.

**Baseline attribution.** The always-loaded surface is unchanged by this milestone — mirroring
writes only under `internal/template/templates/`, which the guard does not enumerate.
`go test ./internal/config/` → `ok 1.576s` with the guard passing at the M4 figure.

| Suite | Result |
|---|---|
| `go test -count=1 ./internal/template/` | `ok 26.672s` |
| `MOAI_TEMPLATE_LEAK_STRICT=1 … -run 'Leak\|Neutrality' -v` | 10/10 PASS |
| `go test -count=1 ./internal/config/` | `ok 1.576s` |
| `go test -count=1 ./internal/hook/` | `ok 25.104s` |
| `go test -count=1 ./internal/constitution/` | `ok 0.442s` |
| `go vet ./internal/template/ ./internal/config/` | clean |
| `go test -count=1 ./internal/cli/` | `ok 346.620s` |

**Gap — none carried forward from M3.** The `CLAUDE.md` mirror deferral recorded at M3 and enlarged
at M4 is closed; the live and template copies are byte-identical again, so `moai update` on this
tree can no longer revert the §0 import or the M4 compression.

**Residual risk.** The scratch deployment confirms the import *resolves*; it does not confirm how
Claude Code renders the assembled file, which is a runtime behavior no test in this repo observes.
The evidence that the mechanism works remains the two `.moai/config/sections/*.yaml` imports that
resolve in this repo today, now joined by a third that resolves in a freshly deployed project.

### M5-b — Codex-cap byte guard (2026-08-22)

Tree: worktree `.claude/worktrees/t82`, branch `WT-agents-md-diet`, parent `8294861c5`.
Files written: `internal/config/token_budget_guard.go`,
`internal/config/token_budget_guard_test.go`.

**What landed.** `CodexContractByteCeiling = 24576` plus `MeasureContractBytes(repoRoot)`, which
returns a `ContractByteBreach{Path, Bytes, Overflow}` per offending document. It reuses
`alwaysLoadedSurface()` to find the root `AGENTS.md` rather than re-globbing — `REQ-AMC-008` forbids
a second, independently-drifting measurement path — and then adds the template mirror, because a
mirror over the ceiling truncates on a user's machine regardless of the live file's size. Absent
files are skipped, keeping the existing hermetic treatment so a tree without the mirror measures
the same baseline as before.

| AC | Claim | Evidence (command → observed) | Status |
|---|---|---|---|
| `AC-AMC-016` (b) | A breach **fails the build** — not warns — and the message names the measured byte figure and the offending file | **Falsified on the real tree before accepting**: 11,000 bytes appended to `AGENTS.md`, then `go test ./internal/config/ -run 'CodexContractByteCeiling'` → `--- FAIL`, message `…/AGENTS.md = 25229 bytes, exceeds the Codex contract ceiling 24576 (overflow 653) — codex truncates the tail SILENTLY…`. `git checkout -- AGENTS.md` restored 14,229 B and the test passes again | PASS |
| `AC-AMC-016` (b) negative table | The breach detector fires on each shape it must catch | Four subtests extend the **existing** table-driven test in the same file, reusing its repo-root helpers: `contract-at-ceiling` → 0 breaches (the ceiling is inclusive), `contract-one-byte-over` → 1, `mirror-over-live-under` → 1 (the mirror is bound independently), `both-over` → 2. Each asserts the breach names a path and that `Overflow == Bytes − ceiling`. All PASS | PASS |
| `AC-AMC-016` (a) | Token-budget negative path cited, not rebuilt | `--- PASS: …/over-budget`, `--- PASS: …/under-budget` — unchanged, no new fixture | PASS |
| `AC-AMC-017` | The byte guard calls the enumeration helper, and that enumeration contains `AGENTS.md` | `contractDocuments()` calls `alwaysLoadedSurface()` and selects the `AGENTS.md` entry from its output; it returns nil when the enumeration lacks it, which is the condition M5-a's membership assertion already fails on. `TestAlwaysLoadedSurfaceEnumeration` → PASS | PASS |
| `AC-AMC-015` | Suites green; no expected-behavior assertion edited | `go vet ./internal/config/` clean; `go test -count=1 ./internal/config/` → `ok 2.718s`. The pre-existing `over-budget` / `under-budget` cases are untouched; everything added is new. One fixture line was added inside the new subtests to create `.claude/rules/moai/core/` — required because `contractDocuments` reuses the enumeration helper, which walks that tree | PASS |

**Measured figures, emitted on the passing path** so a ratchet criterion can quote them:

```
go test -v ./internal/config/ -run 'Budget|AlwaysLoaded|Codex'
  always-loaded surface = 66266 tokens (budget 76000, headroom 9734, 18 entries)
  contract document AGENTS.md = 14229 bytes (ceiling 24576, headroom 10347)
  contract document internal/template/templates/AGENTS.md = 14229 bytes (ceiling 24576, headroom 10347)
```

**Deferred by construction — `AC-AMC-018`, `AC-AMC-019`, `AC-AMC-020`.** These three ratchet the
constant, and `REQ-AMC-014` requires the achieved figure to be a `go test -v` output measured on the
**integration branch** — the `release/vX.Y.Z` branch this card merges into, carrying the merged
sibling state — not a card worktree measured in isolation. This worktree is not that tree, so
measuring here and calling it the achieved figure would be the exact substitution the requirement
names. `AlwaysLoadedTokenBudget` is therefore left untouched at 76,000.

The procedure at integration, in order:

1. Merge this card into `release/vX.Y.Z`, and confirm `AC-AMC-021` still passes there (every
   shipped file mirrored, `make build` run) — the post-diet precondition `AC-AMC-018` names.
2. `go test -v ./internal/config/ -run 'Budget|AlwaysLoaded'` on that branch; quote the
   `always-loaded surface = N tokens` line, and confirm in the **same run** that the enumeration
   contains `AGENTS.md` (the ordering assertion — a figure produced before the extension does not
   satisfy the criterion even if the extension lands afterwards).
3. Record `git rev-parse --abbrev-ref HEAD` and `git rev-list --count main..HEAD` alongside N, so
   the measured tree is identified rather than asserted.
4. Set `AlwaysLoadedTokenBudget = N × (1 + ratio)` with the ratio inside `REQ-AMC-013`'s 13 %-17 %
   band and the constant at or below 75,000; the two figures agree within ±1,000 tokens
   (`AC-AMC-019` reads **both** halves — the agreement check alone passes vacuously on a ratio the
   same edit declared).
5. Rewrite the constant's comment: state the ratio, record the ratchet, and retire the "temporary
   raise pending a separate card" note (`AC-AMC-020`).

**Residual risk — the admissible ratio band is what the margin is really measured against.**
`REQ-AMC-013` fixes the band at 13 %-17 % and caps the constant at 75,000, so the largest achieved
figure any admissible ratio can carry is `75,000 ÷ 1.13 = 66,371` — which is where this SPEC's
target figure comes from. Today's worktree reading of 66,266 clears it by **105 tokens**, and the
ratio is then forced to the band's low end: 13 % yields `66,266 × 1.13 = 74,881`, under the cap,
while 15 % yields 76,206 and 17 % yields 77,531, both over it. So the ratchet closes, but only at
13 %, and only while N stays at or below 66,371.

N on the integration branch will not equal 66,266 — sibling cards merge in between, and each
always-loaded rule they add consumes that 105-token margin. If the integration measurement comes in
over 66,371, no admissible ratio produces a constant at or below 75,000 and the diet has to be
extended before the ratchet can be set. The two levers M4 left sized are where that would come
from: 29.4 % of the never-split nine's projected yield is unspent, and the already-split remainder
was drawn down only 6.0 %. Naming the arithmetic here is cheaper than deriving it at the ratchet.

## §E.3 Run-phase Audit-Ready Signal

- run_status: complete (M1-M6 landed; M5's ratchet half deferred by requirement, see below)
- milestones_landed: M1 `24addddda` · M2 `fd3ac06a8` + `e7484034a` · M3 `a2c919792` · M5-a
  `af954cc0c` · M4 `243eb07ef` · M6 `8294861c5` · M5-b `c3f732490`
- milestone_order_deviation: **M5-a landed before M4, deliberately.** `REQ-AMC-013` orders the
  enumeration extension before *any* measurement cited as a ratchet basis — "not merely before the
  ratchet's final commit". M4 quotes per-file before/after figures, so running it first would have
  recorded reductions that did not happen, in a gap the requirement states does not close
  retroactively.
- ac_matrix: 21 of 24 PASS with evidence recorded per milestone above; 3 deferred by requirement
  (`AC-AMC-018` / `AC-AMC-019` / `AC-AMC-020`), none FAIL, none PASS-WITH-DEBT
- deferred_ac_grounds: `REQ-AMC-014` requires the achieved figure to be a `go test -v` output
  measured on the **integration branch** carrying merged sibling state, not a card worktree in
  isolation. Measuring here and calling it achieved is the exact substitution the requirement
  names. `AlwaysLoadedTokenBudget` is therefore untouched at 76,000; the five-step procedure is
  recorded in §E.2 M5-b.
- achieved_figure_worktree: `always-loaded surface = 66266 tokens (budget 76000, headroom 9734,
  18 entries)` — a worktree reading, explicitly **not** the ratchet basis
- contract_layer: `AGENTS.md` 14,229 B live and mirrored, 10,347 B under the 24,576 B ceiling and
  18,539 B under the 32,768 B codex budget
- falsification_performed: `AC-AMC-017` (slot removed → enumeration test fails twice → restored)
  and `AC-AMC-016(b)` (+11,000 B on `AGENTS.md` → build fails naming figure and file → restored).
  Both guards were shown to fire before being accepted as passing.
- test_evidence: `./internal/config/` ok · `./internal/constitution/` ok · `./internal/hook/` ok ·
  `./internal/template/` ok (incl. `MOAI_TEMPLATE_LEAK_STRICT=1` leak/neutrality 10/10) ·
  `./internal/cli/` ok 346.620s · `go vet` clean
- test_assertions_edited: 2, both named by the `AC-AMC-015` surface-cardinality carve-out
  (`wantRuleCount + 3` → `+ 4`; temp-tree `want 4` → `5`). No other expected-behavior assertion
  was changed.
- deployment_verified: `moai init . --non-interactive` into a scratch project outside the repo —
  `AGENTS.md` and `CLAUDE.md` deployed, all three `@`-imports resolve against the **deployed**
  project root, 12 companions deployed

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: complete
- sync_complete_at: 2026-08-22
- sync_commit_sha: _<backfilled by `moai spec close`; a commit cannot contain its own SHA>_
- changelog_entry_position: `CHANGELOG.md` `## [Unreleased]` → `### Added` (the contract layer, its
  byte guard, and the enumeration extension) and `### Changed` (the eleven-document diet and the
  global-layer warning)
- frontmatter_status_transitions:
  - spec.md: draft → in-progress (run) → implemented (this sync commit) → completed
    (`moai spec close` atomic transition)
  - plan.md / acceptance.md: no frontmatter block — this SPEC carries `status:` on `spec.md` only
  - progress.md: §E.3 pending → complete (this sync commit)
- open_followups:
  - `AC-AMC-018` / `AC-AMC-019` / `AC-AMC-020` — the ratchet, measured on the integration branch
    after the batch completes. Procedure: §E.2 M5-b.
  - The admissible-ratio arithmetic that constrains it: `REQ-AMC-013`'s 13 %-17 % band and the
    75,000 cap mean the largest achieved figure any admissible ratio can carry is
    `75,000 ÷ 1.13 = 66,371`. A measurement above that closes no ratchet at any ratio.
  - `.claude/output-styles/moai/moai.md` — 61,706 B, 23 % of the always-loaded surface, untouched
    by this SPEC and excluded from M1's reduction bound. The next lever if the ratchet tightens.

# SPEC-AGENTS-MD-CANON-001 — Implementation Plan

> Tier L. Milestones are ordered by decision-reversibility: the decisions most likely to change
> (which clauses are Codex-relevant, how the contract is worded) come first; mechanical relocation,
> guard wiring, and mirroring come last. No time estimates — priority labels and ordering only.

## §A. Context

See `spec.md` §A. The three premises are measured (`.moai/reports/t82/codex-probe.md`), which
changes the plan's shape in three ways:

- The nested-`AGENTS.md` milestone is **gone**. Nested documents are unloaded at repo-root
  invocation and share rather than expand the budget, so Option A — a single self-sufficient root
  contract — is the approved design.
- The ceiling is **24,576 B**, not 8 KiB. Measured as clause blocks the verbatim contract is
  51,639 B against a confirmed 32,768 B budget — **1.58x the budget**, so it does not fit at all;
  the 32,543 B line proxy that read as a 0.7 % fit was undercounting (`spec.md` §A.4). The work is
  classification plus condensation, and then establishing headroom.
- The CI byte guard is **mandatory and blocking**, because truncation is measured silent.

## §B. Known issues entering the plan

1. **The card's baseline was stale** — it undercounts the surface by ~34 % and omits the largest
   always-loaded file.
2. **The budget ratchet is branch-sensitive.** The `release/v3.1.1` integration state that forced
   the raise measured 75,282 tokens while each constituent card sat within budget. Re-measured at
   run-phase pre-flight, that divergence is not present today — this worktree and all four live
   refs read 71,207 — but it returns as soon as a sibling card lands on the v3.2 integration
   branch, so a ratchet proposed from a worktree figure alone is still meaningless.
3. **The Claude-only exclusion was bounded but not settled — M1 settled it.** The 14,360 B
   six-file figure was a line-level upper bound; the per-clause split measures **16,135 B
   Codex-relevant / 35,147 B Claude-only**, with 38.8 % of the Claude-only bytes falling outside
   those six files (`spec.md` §D.2).
4. **The contract figure was a line-level proxy, not a measurement.** M1's clause-block expansion
   measured **51,639 B** against the proxy's 32,543 — the proxy runs roughly one over and
   ninety-six under, not the 15-over / 16-under bracket assumed here (`spec.md` §A.4). Nothing
   downstream may assume 32,543 B is exact; the ceiling is re-derived against the clause-block
   figure in `spec.md` §D.1.
5. **The ratchet target is unreachable until the guard prints its figure.** The existing test emits
   the token total only on failure — M5's first task is the `t.Logf`, before any ratchet criterion
   can be executed at all.

## §C. Pre-flight (run-phase entry gate)

- [ ] `go test -v ./internal/config/ -run 'Budget|AlwaysLoaded'` green on the entry tree, with the
      logged token figure recorded as the pre-diet baseline (requires M5's `t.Logf`, which is
      therefore the first code change of the run phase).
- [ ] Integration-branch surface measured (`release/vX.Y.Z`, branch + ahead-count recorded), so the
      ratchet has a real baseline rather than a worktree-local one.
- [ ] `.moai/reports/t82/probe-fixture.sh` executes and reproduces the recorded runs.

No premise gate remains: all four entry premises (P1-P4) are discharged by measurement
(`spec.md` §A.2).

## §D. Constraints

`spec.md` §D.4. The one that shapes every milestone: no `[HARD]` clause leaves the always-loaded
surface, so the byte reduction must come entirely from rationale, procedure, examples, and incident
records.

---

## §E. Milestones

### M1 — Clause classification and compressibility measurement (Priority: High)

> **Sequencing boundary against M5's guard extension.** M1 measures *text volume* — clause-block
> bytes and a compression ratio — which is independent of what the guard enumerates, so M1 may run
> in parallel with M5's extension. **The moment an M1 figure is cited as a ratchet basis, the
> enumeration extension must already have landed** (REQ-AMC-008, REQ-AMC-013). The boundary is
> not "before the ratchet commit" but "before the first measurement anyone will quote": a figure
> taken while `AGENTS.md` is unenumerated is not merely early, it is wrong in a way that reads as
> success, and it does not become right when the guard is fixed later.

**Why first.** Two unmeasured quantities price everything downstream: how much of the contract is
Claude-only, and how far the remainder compresses without losing its obligation. Both are settled
here, before any file moves. **Measured:** the clause-block contract is 51,639 B, of which
16,135 B is Codex-relevant; the pilot compression ratio is 0.5318.

- **Expand markers to clause blocks first.** The 97 figure is a *line* count and the bytes behind
  it are a proxy that undercounts heavily (`spec.md` §A.4: measured, one marker over and
  ninety-six under; the clause-block total is 51,639 B against the proxy's 32,543). Expand each
  marker to its clause block —
  the marker line plus its continuation to the next clause or heading — and measure that. The
  clause-block figure, not the grep figure, is what the ceiling is re-derived against.
- **Classify** every clause block as Codex-relevant or Claude-mechanism-only. The six-file
  14,360 B figure is a line-level upper bound, not the answer — clauses inside those files that
  state harness-generic principles stay in the contract, and the measured split puts 38.8 % of the
  Claude-only bytes outside those six files. Deliver the per-block table with bytes. **When a block
  is genuinely ambiguous, classify it to the Codex side** — a Codex-binding clause misfiled as
  Claude-only leaves the contract silently, while the reverse error costs only bytes
  (`spec.md` §D.2).
- **Compress** a pilot set at both ends of the distribution — `kanban-dispatch.md` (23 `[HARD]`
  lines, 25,915 B) and `native-idiom-and-register.md` (2 lines, 4,967 B). Rewrite each clause in
  imperative form preserving subject, modality, and scope; measure before/after per clause.
- **Project** the classified remainder at the measured ratio and compare against the 24,576 B
  ceiling. State the verdict with the number.

**Stop condition — two arms. Both must clear before M2 starts.**

> **Returning a blocker here is a correct outcome, not a milestone failure.** M1 exists to find out
> whether the target is reachable; "it is not, and here is the shortfall in measured units" is the
> answer it was built to produce, and it is produced at the cheapest possible moment — before M2's
> irreversible authoring. Arm B in particular was **expected** to fire: `spec.md` §C.4's correction
> roughly doubled the minimum cut, to **10,980 tokens** at `AGENTS.md`'s ceiling (4,841 → 10,980).
> Re-measured at pre-flight, the integration branch and this worktree currently read the same
> 71,207, so the two baselines coincide today — a fact about today's tree, not a licence to baseline
> on the worktree (`spec.md` §C.4). The diet is materially larger than the card assumed. An M1 that
> halts with a number is doing its job; an M1 that proceeds past a shortfall it did not measure is
> the failure this milestone prevents.
>
> **Measured outcome: neither arm tripped.** Arm A projected 11,881 B against 24,576 B; Arm B a
> required cut of 7,806 tokens against 10,670 available, margin +2,864 — conditional on
> `AGENTS.md` landing near the projection rather than at the ceiling
> (`.moai/reports/t82/m1-pilot.md`; the binding form of that condition is stated under
> `AC-AMC-009`).

*Arm A — contract fits its ceiling.* If the contract projection exceeds **24,576 B**, do not
proceed to M2. Return a blocker naming the shortfall in bytes and the two levers — deeper
condensation with its stated quality cost, or renegotiating the ceiling against the 8,192 B
headroom reserve (with what the reserve is protecting, per `spec.md` §D.1, stated so the trade is
visible). Do not silently expand toward 32,768 B.

*Arm B — the diet reaches the ratchet ceiling.* Project the **post-diet always-loaded surface,
including the contract layer**, against **66,371 tokens** (`spec.md` §C.4). **Baseline the
projection on the integration-branch figure recorded at pre-flight (§C), not on this worktree's.**
The two can differ by 4,070 tokens (71,212 worktree vs 75,282 at the integration state that forced
the 76,000 raise) — a 37 % difference in the required cut — so a projection baselined on the
worktree can clear Arm B and still fail `AC-AMC-018` at M5, which is the failure Arm B exists to
prevent. Today the two coincide at 71,207 (pre-flight); the requirement binds on the mechanism, not
on today's reading. If the projection exceeds the ceiling, return a blocker naming
the shortfall in tokens and the same two levers — deeper relocation of R3 material into skills with
its navigability cost, or renegotiating the scope with what is being traded stated explicitly. Do
not proceed to M2 on the assumption that M4 will find the difference.

**Why Arm B exists.** Arm A alone tests only whether the *contract* fits `AGENTS.md`; it says
nothing about whether the *diet* reaches the figure `REQ-AMC-013` needs. Those are different
quantities, and only the second decides whether M5 can close at all. Without Arm B, an M1 actor who
clears 24,576 B proceeds, M2/M3/M4 land, and the shortfall surfaces at M5 — after the expensive,
least-reversible work is done. That is exactly the discovered-at-M5 failure §C.4 was written to
prevent, and the stop condition is where it gets caught.

### M2 — Root `AGENTS.md` contract layer (Priority: High)

**Why second.** The wording of a standing contract is the least reversible artifact here; every
later reference, guard, and mirror keys off it.

- Author the single root `AGENTS.md` from M1's classified, compressed clause set.
- Self-sufficiency is the acceptance property: the document assumes no nested file is loaded.
- Every clause traces to its origin file and line; the trace table is a run-phase deliverable.
- Measured size recorded against the 24,576 B ceiling and the 32,768 B budget.

### M3 — `CLAUDE.md` → `AGENTS.md` import layer (Priority: High)

**Why before relocation.** It is user-facing and changes how Claude Code assembles context; a
mistake here is a Claude-side regression, which REQ-AMC-010 forbids.

- The `@`-import mechanism is **already in use in this repo** — `CLAUDE.md` §9 imports
  `.moai/config/sections/user.yaml` and `language.yaml`, and both resolve today. The mechanism is
  verified by existing behavior, not assumed; run-phase still confirms resolution after the edit.
- `CLAUDE.md` retains a Claude-only layer carrying exactly the clauses M1 classified as
  Claude-mechanism-only.
- Verify no duplicate injection — contract text loaded once, not once via import and again inline.

### M4 — Detail relocation to skills (Priority: Medium)

**Why late.** Mechanical once M1-M3 have settled what may move.

> **Trust-notice obligation recorded for t88 (the wiring generator).** Emitting a project-scope
> `project_doc_max_bytes` from `moai init --agent codex` is valid, but **emitting it alone is not
> enough**: the value does nothing until the user registers the project
> `trust_level = "trusted"`, and an unregistered project's config file is ignored **silently**
> (`spec.md` §D.8). So the generator must either tell the user trust registration is required, or
> `moai doctor` must actively check for "project config present, trust unregistered, cap not
> applied". Because the failure is silent, that check is the only detection path there is. This
> SPEC records the grounds; the obligation itself is t88's to carry.

- Move rationale, procedure, worked examples, incident records, and long cross-reference tables out
  of the always-loaded rules into skills or lazy companions, following the stub + lazy-companion
  pattern inherited from `SPEC-ALWAYS-LOADED-DIET-001`.
- Per-file before/after byte measurement; running total against the ratchet target.
- Each moved block leaves a pointer, so the stub stays navigable.

### M5 — CI byte guard and budget ratchet (Priority: Medium)

- **First, make the existing guard observable.** `TestAlwaysLoadedTokenBudget` emits the token
  total only via `t.Errorf` on failure, so a passing run prints `ok …` and nothing else — which
  makes every "quote the achieved figure" criterion unsatisfiable. Add, ahead of the over-budget
  check:
  `t.Logf("always-loaded surface = %d tokens (budget %d, headroom %d)", total, AlwaysLoadedTokenBudget, AlwaysLoadedTokenBudget-total)`
  so a passing `go test -v` emits the figure. Every ratchet criterion reads that line.
- **Extend `alwaysLoadedSurface()` itself to enumerate the root `AGENTS.md` — and do it before any
  measurement that will be quoted.** This is a code change, not a documentation one, and its
  ordering is part of the requirement (REQ-AMC-008, REQ-AMC-013): measurements taken in the gap
  record reductions that did not happen and are indistinguishable from a real diet when read later.
  The guard cannot stay honest without it: the function today
  carries three fixed slots (`CLAUDE.md`, the output style, `MEMORY.md`) and does not know about
  `AGENTS.md`, while M3 makes `AGENTS.md` an `@`-import of `CLAUDE.md` and therefore always-loaded.
  Unextended, M4's relocation of clauses into `AGENTS.md` would show as up to ~6,144 tokens of
  "diet" with the always-loaded context unchanged. Add it as a fourth fixed slot, keeping the
  existing hermetic treatment (a file absent from disk measures 0), so the pre-`AGENTS.md` baseline
  is unaffected and the enumeration stays one path.
- Extend `internal/config/token_budget_guard.go` with a Codex-cap byte guard reusing that same
  enumeration (REQ-AMC-008) — no second measurement path. It binds both the live root `AGENTS.md`
  and its template mirror.
- The guard **fails the build**, it does not warn (REQ-AMC-009, rationale `spec.md` §D.7).
  Truncation is measured silent, so this is the only signal that will ever fire.
- Failure output names the measured byte figure and the offending file.
- **Do not author a new over-budget fixture.** `TestAlwaysLoadedTokenBudget_OverBudgetFails` already
  covers the token-budget negative path (over-budget fires, under-budget does not) and passes on
  this tree. Extend its table-driven shape in the same file for the Codex-cap dimension; a parallel
  harness would be the second measurement path REQ-AMC-008 forbids.
- Lower `AlwaysLoadedTokenBudget` to the achieved figure plus a stated headroom ratio, ≤ 75,000,
  measured on the **integration branch** — the `release/vX.Y.Z` branch this card merges into,
  carrying the merged sibling state (REQ-AMC-014). Record `git rev-parse --abbrev-ref HEAD` and
  `git rev-list --count main..HEAD` with the figure so the tree is identified, not asserted.
- Update the constant's comment to record the ratchet, state the headroom ratio (AC-AMC-019 checks
  the constant against `achieved × (1 + ratio)`), and retire the "pending a separate card" note.

### M6 — Template mirror, neutrality, and the global-layer warning (Priority: Medium)

- Mirror every shipped file into `internal/template/templates/`; run `make build`.
- Add the `~/.codex/AGENTS.md` warning to the shipped documentation (`spec.md` §D.3 records the
  decision and its reasoning). It states that a personal global document joins the same chain, is
  consumed first, and narrows what the project's own document can carry.
- Template copies carry no SPEC IDs, REQ tokens, audit citations, internal dates, commit SHAs,
  macOS-biased paths, or `CLAUDE.local.md` references. The mirror is not a verbatim `cp`.

---

## §F. Technical approach

- **Classification is grep-seeded, judgment-gated.** `grep -h '\[HARD\]'` enumerates candidates;
  deciding Codex-relevant vs Claude-only is a per-clause judgment recorded in a table, not a
  filename heuristic.
- **The guard is extended, never duplicated.** One enumeration, two thresholds: the token budget
  over the always-loaded surface, and the byte cap over `AGENTS.md`.
- **The output style is untouched.** `spec.md` §E.2 rules it exempt from the contract and defers
  its own diet.

## §G. Risks

| Risk | Effect | Mitigation |
|---|---|---|
| Contract will not reach 24,576 B (M1) | Ceiling wrong or condensation insufficient | M1 stop condition halts before any file moves; the trade against the headroom reserve is made explicitly |
| The clause-block figure lands far from the 32,543 B proxy — **it did: 51,639 B, +58.7 %** | Ceiling calibrated to the wrong number | The ceiling is a bracket, not a point (`spec.md` §D.1); M1 re-derived it against the clause-block figure and the ceiling held at 24,576 B, with the 8,192 B reserve untouched |
| Ratchet appears satisfied without ratcheting | Constant sits at 75,000 while the surface lands far below | AC-AMC-019 checks the constant against `achieved × (1 + stated ratio)` — the ceiling check alone cannot catch this |
| Over-aggressive condensation changes an obligation | Silent rule change — the worst outcome here | REQ-AMC-003 + AC-AMC-004: subject/modality/scope diffed per clause; failures reverted, not shipped |
| Ratchet unreachable without the output style | Scope pressure mid-run | `spec.md` §E.2 makes this a blocker to surface, not a scope widening |
| A user's global `~/.codex/AGENTS.md` silently eats the budget | Project rules truncated with no signal | Cannot be guarded mechanically (file is outside the repo); documented warning is the chosen defence (`spec.md` §D.3) |
| Upstream default differs on another OS or codex version | Ceiling calibrated to a stale figure | Recorded as residual risk (`spec.md` §D.9); re-probe on a codex upgrade |
| Template mirror drift | Distributed users get a stale contract | M6 mirror + `make build` + CI neutrality guard |

## §H. Cross-references

`spec.md` §G.

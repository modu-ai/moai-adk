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
- The ceiling is **24,576 B**, not 8 KiB. The verbatim contract measures 32,543 B against a
  confirmed 32,768 B budget: it fits with 0.7 % headroom, which is a numeric fit and a practical
  failure. The work is establishing headroom, not achieving a fit.
- The CI byte guard is **mandatory and blocking**, because truncation is measured silent.

## §B. Known issues entering the plan

1. **The card's baseline was stale** — it undercounts the surface by ~34 % and omits the largest
   always-loaded file.
2. **The budget ratchet is branch-sensitive.** This worktree measures ≈ 71,212 tokens, already
   under 75,000; the release integration state that forced the raise measured 75,282. A ratchet
   proposed from the worktree figure alone would be meaningless.
3. **The Claude-only exclusion is bounded but not settled.** 14,360 B is an upper bound across six
   files; the per-clause split is M1's job and determines how much condensation M2 must carry.

## §C. Pre-flight (run-phase entry gate)

- [ ] `go test ./internal/config/ -run 'Budget|AlwaysLoaded'` green on the entry tree.
- [ ] Integration-branch surface measured, so the ratchet has a real baseline.
- [ ] `.moai/reports/t82/codex-probe.md` present and its fixture reproducible (the three commands
      in its § 검증 재현 block).

No premise gate remains: all three entry premises are measured.

## §D. Constraints

`spec.md` §D.4. The one that shapes every milestone: no `[HARD]` clause leaves the always-loaded
surface, so the byte reduction must come entirely from rationale, procedure, examples, and incident
records.

---

## §E. Milestones

### M1 — Clause classification and compressibility measurement (Priority: High)

**Why first.** Two unmeasured quantities price everything downstream: how much of the 32,543 B
contract is Claude-only (upper bound 14,360 B), and how far the remainder compresses without losing
its obligation. Both are settled here, before any file moves.

- **Classify** every one of the 97 `[HARD]` clauses as Codex-relevant or Claude-mechanism-only.
  The six-file 14,360 B figure is an upper bound, not the answer — clauses inside those files that
  state harness-generic principles stay in the contract. Deliver the per-clause table with bytes.
- **Compress** a pilot set at both ends of the distribution — `kanban-dispatch.md` (23 `[HARD]`
  lines, 25,915 B) and `native-idiom-and-register.md` (2 lines, 4,967 B). Rewrite each clause in
  imperative form preserving subject, modality, and scope; measure before/after per clause.
- **Project** the classified remainder at the measured ratio and compare against the 24,576 B
  ceiling. State the verdict with the number.

**Stop condition.** If the projection exceeds 24,576 B, do not proceed to M2. Return a blocker
naming the shortfall in bytes and the two levers — deeper condensation with its stated quality
cost, or renegotiating the ceiling against the 8,192 B headroom reserve (with what the reserve is
protecting, per `spec.md` §D.1, stated so the trade is visible). Do not silently expand toward
32,768 B.

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

- Move rationale, procedure, worked examples, incident records, and long cross-reference tables out
  of the always-loaded rules into skills or lazy companions, following the stub + lazy-companion
  pattern inherited from `SPEC-ALWAYS-LOADED-DIET-001`.
- Per-file before/after byte measurement; running total against the ratchet target.
- Each moved block leaves a pointer, so the stub stays navigable.

### M5 — CI byte guard and budget ratchet (Priority: Medium)

- Extend `internal/config/token_budget_guard.go` with a Codex-cap byte guard reusing
  `alwaysLoadedSurface()`'s enumeration (REQ-AMC-008) — no second measurement path.
- The guard is **blocking, not advisory** (REQ-AMC-009). Truncation is measured silent, so this is
  the only signal that will ever fire.
- Failure output names the measured byte figure and the offending file.
- Lower `AlwaysLoadedTokenBudget` to the achieved figure plus a stated headroom ratio, ≤ 75,000,
  measured on the integration branch (REQ-AMC-014). Update the constant's comment to record the
  ratchet and retire the "pending a separate card" note.

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
| Over-aggressive condensation changes an obligation | Silent rule change — the worst outcome here | REQ-AMC-003 + AC-AMC-004: subject/modality/scope diffed per clause; failures reverted, not shipped |
| Ratchet unreachable without the output style | Scope pressure mid-run | `spec.md` §E.2 makes this a blocker to surface, not a scope widening |
| A user's global `~/.codex/AGENTS.md` silently eats the budget | Project rules truncated with no signal | Cannot be guarded mechanically (file is outside the repo); documented warning is the chosen defence (`spec.md` §D.3) |
| Upstream default differs on another OS or codex version | Ceiling calibrated to a stale figure | Recorded as residual risk (`spec.md` §D.5); re-probe on a codex upgrade |
| Template mirror drift | Distributed users get a stale contract | M6 mirror + `make build` + CI neutrality guard |

## §H. Cross-references

`spec.md` §G.

# SPEC-AGENTS-MD-CANON-001 — Implementation Plan

> Tier L. Milestones are ordered by decision-reversibility: the decisions most likely to change
> (whether the contract compresses at all, how the contract is worded, what the directory map is)
> come first; mechanical relocation, guard wiring, and mirroring come last.
> No time estimates anywhere — priority labels and ordering only.

## §A. Context

See `spec.md` §A. The load-bearing number: the verbatim `[HARD]` contract layer across the
always-loaded rules and `CLAUDE.md` measures **32,543 B**, against a card target of ~8 KiB and a
presumed 32,768 B shared cap. The design as stated in the card is not reachable by moving text.

## §B. Known issues entering the plan

1. **`.moai/reports/t91/` is absent.** Three of the four entry premises (§D.1) have no measurement.
2. **The card's baseline is stale** — it undercounts the surface by ~34 % and omits the largest file.
3. **The budget ratchet is branch-sensitive.** This worktree measures ≈ 71,212 tokens, already under
   75,000; the release integration state that forced the raise measured 75,282. A ratchet proposed
   from the worktree figure alone would be meaningless.

## §C. Pre-flight (run-phase entry gate)

- [ ] P1-P4 (`spec.md` §D.1) all measured, with the measuring command and its output recorded.
- [ ] `go test ./internal/config/ -run 'Budget|AlwaysLoaded'` green on the entry tree.
- [ ] Integration-branch surface measured, so the ratchet has a real baseline.

## §D. Constraints

`spec.md` §D.2. The one that shapes every milestone: no `[HARD]` clause leaves the always-loaded
surface, so the diet's byte reduction must come entirely from rationale, procedure, examples, and
incident records.

---

## §E. Milestones

### M1 — Contract extraction pilot and compressibility ratio (Priority: High)

**Why first.** Everything downstream is priced off one unmeasured quantity: how far a `[HARD]`
clause compresses without losing its obligation. If the ratio is poor, the three-way design does
not work and the SPEC must change shape before any file is moved.

- Pick two files at opposite ends of the distribution — `kanban-dispatch.md` (23 `[HARD]` lines,
  25,915 B) and `native-idiom-and-register.md` (2 `[HARD]` lines, 4,967 B).
- Extract their contract clauses verbatim; rewrite each in imperative telegraphic form preserving
  subject, modality, and scope (REQ-AMC-003).
- Measure before/after bytes per clause; report the ratio and its variance.
- Extrapolate to the full 32,543 B and state, with the number, whether the root target is reachable.

**Stop condition.** If the extrapolated contract exceeds the confirmed cap minus a stated reserve
for nested documents, do not proceed to M2. Return a blocker naming the shortfall in bytes and the
two available levers (further condensation with a stated quality cost; raising the cap, only if P4
confirms project-scope configurability), and let the orchestrator route the choice to the user.

### M2 — Root `AGENTS.md` contract layer (Priority: High)

**Why second.** The wording of a standing contract is the least reversible artifact here: every
later reference, guard, and mirror keys off it.

- Author the root `AGENTS.md` from the M1-validated rewrite procedure.
- Every clause traces to its origin file and line; the trace table is a run-phase deliverable.
- Measured byte figure recorded against the confirmed cap.

### M3 — Nested `AGENTS.md` directory map (Priority: High, conditional on P2)

**Why here.** The map is a design decision, but it is downstream of knowing the contract's size.

- **Conditional existence**: if P2 shows Codex merges only the invocation CWD's chain, and MoAI
  sessions launch from the repo root, this milestone is dropped and its content folds into M2.
  Record the ruling either way.
- If it proceeds: name each nested file, the directory it owns, and the clauses it carries.
  Bound the count by shared-budget arithmetic on the deepest reachable chain (REQ-AMC-005).

### M4 — `CLAUDE.md` → `AGENTS.md` import layer (Priority: High)

**Why before relocation.** It is user-facing and changes how Claude Code assembles its context;
a mistake here is a Claude-side regression, which REQ-AMC-008 forbids.

- The `@`-import mechanism is **already in use in this repo** — `CLAUDE.md` §9 imports
  `.moai/config/sections/user.yaml` and `language.yaml`, and both resolve today. So the mechanism
  is verified by existing behavior, not assumed. Run-phase still confirms resolution after the edit.
- `CLAUDE.md` retains a Claude-only layer for material with no Codex counterpart.
- Verify no duplicate injection (contract text loaded twice, once via import and once inline).

### M5 — Detail relocation to skills (Priority: Medium)

**Why late.** Mechanical once M1-M4 have settled what may move.

- Move rationale, procedure, worked examples, incident records, and long cross-reference tables
  from the always-loaded rules into skills or lazy companions, following the stub + lazy-companion
  pattern inherited from `SPEC-ALWAYS-LOADED-DIET-001`.
- Per-file before/after byte measurement; running total against the ratchet target.
- Each moved block leaves a pointer, so the stub stays navigable.

### M6 — Regression guard and budget ratchet (Priority: Medium)

- Extend `internal/config/token_budget_guard.go` with a Codex-cap byte guard that reuses
  `alwaysLoadedSurface()`'s enumeration (REQ-AMC-007) — no second measurement path.
- Guard failure names the measured byte figure and the offending file.
- Lower `AlwaysLoadedTokenBudget` to the achieved figure plus a stated headroom, ≤ 75,000,
  measured on the integration branch (REQ-AMC-012). Update the constant's comment to record the
  ratchet and retire the "pending a separate card" note.

### M7 — Template mirror and neutrality (Priority: Medium)

- Mirror every shipped file into `internal/template/templates/`; run `make build`.
- Template copies carry no SPEC IDs, REQ tokens, audit citations, internal dates, commit SHAs,
  macOS-biased paths, or `CLAUDE.local.md` references.
- Mirror is not a verbatim `cp`: the local copies carry provenance the templates must not.

---

## §F. Technical approach

- **Extraction is grep-driven, review-gated.** `grep -h '\[HARD\]'` finds the candidates; a human
  or auditor pass decides which are Codex-relevant and which bind Claude rendering only.
- **The guard is extended, never duplicated.** One enumeration, two thresholds (token budget,
  Codex byte cap).
- **The output style is untouched.** `spec.md` §E.2 rules it structurally exempt from
  redistribution and defers its own diet.

## §G. Risks

| Risk | Effect | Mitigation |
|---|---|---|
| Compression ratio too poor (M1) | Design infeasible | M1 stop condition halts before any file moves |
| P2 shows no nested merge | Nested leg vanishes; whole contract must fit one document | M3 is conditional by construction; §A.2 already prices the single-document case |
| Silent truncation (P3 negative) | Regression escapes locally | REQ-AMC-006 byte guard becomes the sole defense — make it blocking, not advisory |
| Ratchet unreachable without the output style | Scope pressure mid-run | §E.2 makes this a blocker to surface, not a scope widening |
| Template mirror drift | Distributed users get stale contract | M7 mirror + `make build` + CI neutrality guard |

## §H. Cross-references

`spec.md` §G.

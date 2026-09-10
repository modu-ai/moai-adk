---
id: SPEC-UPDATE-MERGE-CONFLICT-BLIND-001
title: "Implementation plan — update merge conflict blindness"
version: "0.1.0"
created: 2026-09-10
---

# Plan — SPEC-UPDATE-MERGE-CONFLICT-BLIND-001

## §A Context

Read `spec.md` §A first; it is not restated here. The single fact that shapes this
plan: everything the SPEC asserts about merge behaviour is **static-derived and
unexecuted** (`spec.md` §A.5). The plan is therefore ordered so that the
observation comes before any design decision that depends on it, and so that the
decisions most likely to change sit at the top where review is cheapest.

## §B Known issues carried in from the investigation

1. **The card's own premise was stale.** `permissions.ask` is populated again in
   the primary checkout; the emptied state survives only in the preserved drift
   snapshot. Any milestone that re-measures the live file will find it healthy —
   that is not evidence the merge behaves correctly.
2. **The mtime evidence for the original emptying was destroyed** by the
   `moai update` redeploy that restored the file. Timestamp-based reasoning about
   the original event is unavailable and must not be attempted.
3. **`bypassPermissions` is set in `settings.local.json`.** Any claim that this
   work hardens the permission layer would be an unobserved claim while that mode
   is active.

## §C Pre-flight

- [ ] `.moai/reports/t576/verdict.md` read in full.
- [ ] `internal/cli/update/merge/base.go` and `internal/merge/strategies.go` read
      at the cited line ranges, and the citations re-confirmed against the tree's
      current HEAD (a line citation decays; cite the SHA alongside it).
- [ ] The four-cell fixture design reviewed against `spec.md` §B.2 before any
      code is written.

## §D Constraints

- **[HARD] No real checkout.** The reproduction runs in a directory the harness
  created (`t.TempDir()` or a scratch dir). It must never read or write this
  project's `.claude/`, and must not invoke `moai update` against a real tree
  (`REQ-UMC-001`, `REQ-UMC-002`).
- **[HARD] No permission value is touched** in any checkout (`REQ-UMC-013`, and
  `spec.md` §D).
- **[HARD] Designed behaviour is preserved.** No remedy changes which value the
  merge writes for a shared key (`REQ-UMC-010`).
- **[HARD] Template-First.** Any `.claude/settings.json` change is paired with
  `internal/template/templates/.claude/settings.json.tmpl` plus `make build`
  (`REQ-UMC-014`).
- Verification is scoped to the packages the change touches; the full-suite
  verdict is CI's (`CLAUDE.local.md` §4).

## §E Self-verification

- Every cell's outcome is quoted from the command that produced it, with the
  command shown. A summary is not evidence.
- An unmeasured cell is recorded as a gap, never inferred from a sibling
  (`REQ-UMC-007`).
- The reachability claim (`REQ-UMC-004`) is measured from `HasConflict` and
  `len(Conflicts)` on a fixture where a shared key genuinely disagrees — not from
  re-deriving the base and comparing it to the value it was derived from, which
  would be a tautology.

## §F Milestones

Ordered by decision-reversibility: the observation and the open design decision
come first; the mechanical work last.

### M1 — Reproduce the four cells (Priority: High; no production code change)

The whole of `spec.md` §A is a hypothesis until this milestone lands. M1 changes
no production code; it adds a test that measures.

1. Build a fixture in an isolated temporary directory: a `updated` (template)
   JSON document and a `current` (user) JSON document that differ across the four
   cells of `REQ-UMC-005`.
2. Drive the merge subsystem over the fixture, deriving the base the way the
   update path does.
3. Record, per cell: the value the merge wrote, `MergeResult.HasConflict`, and
   `len(MergeResult.Conflicts)`.
4. Record the same three readings for a shared key on which `current` and
   `updated` genuinely disagree (`REQ-UMC-004`).

Predicted outcomes, fixed **before** measuring so the measurement can disagree:

| Cell | User side | Prediction |
|---|---|---|
| (i) | untouched, shared | template's change does **not** land; user's value stands |
| (ii) | emptied to `[]`, shared | `[]` preserved |
| (iii) | changed to another value, shared | user's value preserved |
| (iv) | key omitted | template's value **lands** |
| conflict probe | shared, genuinely divergent | `HasConflict == false`, `len(Conflicts) == 0` |

A disagreement between prediction and observation is the finding, not a bug in
the harness — investigate before adjusting either.

M1's exit is the recorded evidence, not a repair. Reading it decides M2's design.

### M2 — Instrument honesty (Priority: High; design open until M1 is read)

`REQ-UMC-008`, `REQ-UMC-009`, `REQ-UMC-010`. Two candidate shapes, deliberately
not chosen here — the choice is a decision for the orchestrator and the operator
after M1's evidence exists:

- **(a) Make the surface honest about its limit.** Keep the resolution as-is and
  make the merge report that a shared key's divergence was resolved without a
  conflict determination — so no reader believes a determination was made.
- **(b) Make the determination possible.** Give the merge a base that can differ
  from `updated` for a shared key, which is the deploy-time snapshot surface
  `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001` already owns for
  `.moai/config/sections/*.yaml`. This is the larger change and reaches outside
  this SPEC's named surface; it would be a follow-up SPEC, not this one.

Whichever is chosen, `REQ-UMC-010` binds: the value written for a shared key does
not change.

Guard obligation: a mutant that reverts the M2 change must turn the M2 test red.
A test that passes against both the repaired and the unrepaired code measures
nothing.

### M3 — The signal (Priority: Medium)

`REQ-UMC-011`, `REQ-UMC-012`, `REQ-UMC-013`. Emit a signal when the merge
preserves a user-side value for a key the template ships with a non-empty
security-relevant value; emit nothing otherwise; change no value.

The scoping question M3 must answer explicitly: **which keys count as
security-relevant.** An allow-list breaks silently on what it omits (a known
failure shape in this repository), so the answer is a scoping decision to state
and defend in the run-phase, not an implementation detail to infer.

### M4 — Template coupling, if and only if the repair touches settings (Priority: Low)

`REQ-UMC-014`. Mechanical: mirror the `.claude/settings.json` change into
`internal/template/templates/.claude/settings.json.tmpl` and run `make build`.
Likely **not reached** — M2 and M3 as scoped touch Go code, not the settings
document. Recorded here so the coupling is not discovered late.

## §G Anti-patterns

- **Re-asserting the withdrawn framing.** "The merge is broken because the
  conflict branch cannot fire" contradicts `base.go:105-107`. The defect is the
  unreachable instrument, not the resolution.
- **Narrowing to `permissions.ask`.** Both defects bind every shared key
  (`spec.md` §A.3). A repair scoped to one key leaves the instrument defect intact.
- **Passing on cells (i)-(iii) alone.** All three predict "user's side stands", so
  a harness that copied the user's file passes them. Cell (iv) is the counter-cell.
- **Citing §A as established.** Until M1 lands, `spec.md` §A is a hypothesis
  (`spec.md` §A.5).
- **Measuring the live `.claude/settings.json`.** It is healthy again and says
  nothing about the merge (§B.1).

## §H Cross-references

- `.moai/reports/t576/verdict.md` — authoritative input, including its Gaps.
- `.moai/specs/SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001/` — the deploy-time
  snapshot surface M2 option (b) would reach into; excluded from this SPEC.
- Cards **t598** / **t599** — sibling drift-ledger gaps, excluded.

🗿 MoAI

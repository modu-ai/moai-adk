---
id: SPEC-UPDATE-MERGE-CONFLICT-BLIND-001
title: "Implementation plan — update merge conflict blindness"
version: "0.3.0"
created: 2026-09-10
---

# Plan — SPEC-UPDATE-MERGE-CONFLICT-BLIND-001

## §A Context

Read `spec.md` §A first; it is not restated here. The fact that shaped this plan
at authoring time: everything the SPEC asserted about merge behaviour was
**static-derived and unexecuted**. The plan is therefore ordered so that the
observation comes before any design decision that depends on it, and so that the
decisions most likely to change sit at the top where review is cheapest.

**M1 and then M2.0 have since been executed** (`progress.md` §E.2). M1 found
something wider than the plan-phase premise, and M2.0 fixed its granularity —
`spec.md` §A.6, restated: a shared **leaf's** value cannot change, a leaf absent
on the user's side lands, and a container's value changes only by gaining such
leaves. Two consequences bind the rest of this plan: the measured breadth is now
three (top-level JSON, nested JSON, flat and nested YAML — `spec.md` §A.3) with
the finding scoped by **granularity** rather than by codec, and M2's remaining
work is the M2.1 design choice alone, since the §F M2.0 precondition is
discharged.

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
4. **The finding invites a breadth one level above the one measured.** M2.0 has
   since entered both the recursive `pruneToShared` path and `mergeYAML`, so the
   codec-and-nesting breadth is closed (`spec.md` §A.3). What replaced it is a
   **granularity** boundary: `spec.md` §A.6 holds of a shared **leaf**, and a
   shared **container's** value does change when the template adds a leaf inside
   it. Citing §A.6 as a statement about "a key" is an unobserved claim, and which
   arm of `strategies.go:422-462` fires for a container is not observable from
   `MergeResult` at all.
5. **`.moai/config/sections/*.yaml` (32 files) travels the YAML path.** That is
   why M2.0 was a precondition and not a follow-up: the path carries the bulk of
   the config surface a repair would be judged on. Measured, it reaches the
   **identical** `deepMergeMap` call as JSON (`strategies.go:343` and `:314`), so
   the arm structure is shared rather than parallel — but the reading is a
   measurement, not the source inference: every YAML cell asserts
   `MergeResult.Strategy == yaml_deep`.

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

**Status: executed.** Evidence: `progress.md` §E.2. All four cells held their
predictions, and the conflict-surface reading (`HasConflict=false`,
`len(Conflicts)=0`) is paired with a control on the same engine and key that fired
`true` / `1` under a base differing from both sides — so the reading is a property
of the derived base, not of the harness. The cell that mattered most was
`untouched_shared`: the user's value standing where the template's had changed is
what produced the `spec.md` §A.6 revision.

### M2 — Instrument honesty (Priority: High; design open until M2.0 is read)

#### M2.0 — Precondition: measure the recursive path and the YAML path (no repair)

[HARD] M2 does **not** begin with the design choice. It begins with a
measurement, and no repair is authored until that measurement is recorded.

What was unmeasured when this milestone was authored, and why it blocked: M1's
fixture was a flat JSON document, so `spec.md` §A.6 stood for top-level JSON keys
and for nothing else. Two paths had never been entered:

1. The **recursive** `pruneToShared` path (`base.go:124-126`) — a shared key whose
   two sides both hold a nested map, measured down to a nested leaf.
2. The **`mergeYAML`** path — the same four control cells expressed as a YAML
   fixture, driven through the YAML strategy rather than the JSON one.

Each is measured the way M1 measured its cells: predictions fixed before the run,
readings taken from `MergeResult` as the engine returned it, an unmeasured cell
recorded as a gap rather than inferred from a sibling (`REQ-UMC-007`), and a
counter-cell that would fail if the harness were measuring nothing.

Why it is a precondition rather than a follow-up: `.moai/config/sections/*.yaml`
is 32 files and travels the YAML path. Repairing on the JSON-only reading means
the **fixed scope and the believed scope diverge** — the repair lands, the
instrument reads honest on JSON, and the YAML surface keeps whatever behaviour it
had, unmeasured and now also unexamined. The divergence produces no signal, which
is what makes measuring first cheaper than measuring after.

M2.0's exit is the recorded evidence. Only then is the choice below made.

**Status: executed.** Evidence: `progress.md` §E.2. Both paths were entered, each
on its own fixture, its own codec, and its own control: 20 predictions fixed
before the runs, 0 failed. YAML was established to reach the **identical**
`deepMergeMap` call rather than a separate strategy, and every YAML cell asserts
`MergeResult.Strategy == yaml_deep`, so a YAML reading cannot be a JSON reading
wearing a `.yaml` file name. The measurement forced one correction, and it is a
granularity correction rather than a defect: a shared **container's** value does
change when the template adds a leaf inside it, so `spec.md` §A.6 is restated at
leaf granularity. Which arm fires for a shared container key is not observable
from `MergeResult` and is claimed nowhere. No production code changed, so M2.1
below is untouched — the design choice was deliberately not taken here.

#### M2.1 — The design choice

[HARD] **Entry gate — M2.1 does not begin until this gate is passed.**
`REQ-UMC-008` and `REQ-UMC-009` have each been adjudicated as **leaf**,
**container**, or **both**, and each adjudication records its reason. Until then
no M2.1 artifact — design, test, or code — is authored.

Why this is a gate and not a note: both requirements say "shared key"; `spec.md`
§A.6, as measured in M2.0, distinguishes a shared **leaf** (its value cannot
change) from a shared **container** (its value changes by gaining the template's
new leaves); and the design choice below is a choice of the granularity at which
a conflict is determined, which depends on which of the two the requirements
bind. Choosing (a) or (b) against the unadjudicated wording would fix a
granularity the requirements layer never stated. Gate evidence: the adjudication
and its reason, recorded at both requirements in `spec.md`.

[HARD] **Premise re-measurement — also part of this gate (added 2026-09-11).**
`SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001` (card t656) lands before M2.1 resumes
(operator decision, 2026-09-11). Once it lands, `.claude/settings.json` is merged
against a deploy-time snapshot base instead of the derived base, so the
**only-template-changed** arm — and the both-changed arm — becomes reachable for
that file. `spec.md` §A.6's premise that those arms cannot execute for a shared
leaf must therefore be re-measured on the tree M2.1 resumes on, for
`.claude/settings.json` and for the derived-base files separately, before any
M2.1 design is chosen.

`REQ-UMC-008`, `REQ-UMC-009`, `REQ-UMC-010`. Two candidate shapes, deliberately
not chosen here — the choice is a decision for the orchestrator and the operator
after both M1's and M2.0's evidence exist:

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

- **Re-asserting the withdrawn framing.** The withdrawn sentence — "The merge is broken because the conflict branch cannot fire" — contradicts `base.go:109-111`. The defect is the unreachable instrument, not the resolution. (A mention like this one is quoted on a single line on purpose: `acceptance.md` §D.3's criterion strips single-line quoted and backticked spans before searching, so a mention passes and a use is caught. A quotation broken across two lines is reported — fail-closed, not a false pass.)
- **Citing `spec.md` §A.6 at container granularity.** The value-invariance
  finding is measured of a shared **leaf** — across top-level JSON, nested JSON,
  and flat and nested YAML (`spec.md` §A.3). Stating it of "every shared key" is
  false for a shared container, whose value gains the template's new leaves; and
  naming the arm that fires for a container is an unobserved claim, because
  `MergeResult` does not distinguish the arms.
- **Repairing before M2.0.** Discharged — M2.0 is executed. The reason it was a
  precondition still governs any later scope decision: a repair authored on a
  reading narrower than the scope it will be believed to cover produces no signal
  about the difference.
- **Narrowing to `permissions.ask`.** Both defects bind every shared **leaf** at
  the three breadths measured so far (`spec.md` §A.3), and `permissions.ask` is
  one instance of that class. A repair scoped to one key leaves the instrument
  defect intact.
- **Passing on cells (i)-(iii) alone.** All three predict "user's side stands", so
  a harness that copied the user's file passes them. Cell (iv) is the counter-cell.
- **Citing §A as established beyond what was run.** M1 and M2.0 discharged the
  plan-phase hypothesis marker for §A.1-§A.4 and §A.6 at the breadths and the
  granularity those runs measured — and only there. `spec.md` §A.5 carries the
  list of what stayed unmeasured; read it before citing §A.
- **Measuring the live `.claude/settings.json`.** It is healthy again and says
  nothing about the merge (§B.1).

## §H Cross-references

- `.moai/reports/t576/verdict.md` — authoritative input, including its Gaps.
- `.moai/specs/SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001/` — the deploy-time
  snapshot surface M2 option (b) would reach into; excluded from this SPEC.
- Cards **t598** / **t599** — sibling drift-ledger gaps, excluded.

🗿 MoAI

# SPEC-AUTONOMY-PRECONDITION-001 — Implementation plan

Milestones are ordered by decision-reversibility: the two design decisions (how the deny recognizes
bypass shapes; how the contract projects onto the mission validator) come first, and mechanical
template and documentation work last.

## §A — Context

Card **t1245**, track **A2b**. Base tree `develop` at `553e224f3`, worktree
`.claude/worktrees/t1245`, branch `WT-push-serialize-sign`. Created by lead ruling 09-26 (2) #3,
which split three items out of SPEC-AUTONOMY-ESCALATION-001. Two independent deny components plus
one reuse projection; they share the contract resolver and nothing else.

## §B — Known issues carried in

- Finding **N4** — the sign guard's wrapper coverage was a hole with no requirement. Now
  REQ-AP-005 with a closed, option-aware wrapper list (design.md §C.3) and a mutant-checked
  unknown-wrapper criterion (AC-AP-010).
- Finding **N8** — the sign deny contradicted the "nothing changes under `guided`" promise. Resolved
  in wording, not by weakening the guard (REQ-AP-010, design.md §C.6a, §C.7). At v0.1.2 the
  documentation surface moved to a rule file **this SPEC creates**, because neither surface v0.1.1
  named exists in this tree (spec.md §C.10).
- Finding **N5** — a criterion that could only turn green after work this SPEC does not own. Closed on
  **two** axes, and the second one was still open at v0.1.1:
  - *receipt path* — closed by removal: receipt issuance is A3's, so the allowance and its two
    criteria are withdrawn (spec.md §C.4, §H.2, §H.4).
  - *template documentation* — the iter-1 audit found `AC-AP-013` conditional on A1's template block,
    which is the same defect in a different place. Closed at v0.1.2 by moving the surface (above), and
    the two limbs whose surfaces are not this SPEC's are transferred to their owners (spec.md §H.5).
  The claim "no criterion here is conditional" is now backed by the per-criterion ownership column in
  acceptance.md §G rather than asserted as a blanket sentence — the enumeration is the guarantee.
- Criterion id re-use on the t1235 branch, and this SPEC's own retirement of `REQ-AP-006` /
  `AC-AP-011` / `AC-AP-012` — recorded as an ID-reuse record in spec.md §H.1-H.2, in the form A1
  uses for the same hazard. No retired id is re-used.
- All A1 citations are pinned to commit `25283ebf8` (A1 v0.5.2), and transferred text to `d8926ff9a`
  (the pre-split blob), not to a branch or a relative ref.
- The iter-1 audit's provenance findings (D2, D3, D4) came from rebuilding §H from a summary rather
  than from the source blob. §H now records the command each row was read with.

## §C — Pre-flight (before M1)

1. Re-read branch and HEAD (`git -C <tree> rev-parse --short HEAD`, `git -C <tree> branch --show-current`)
   and confirm `WT-push-serialize-sign` on a tree descended from `553e224f3`.
2. Re-measure every spec.md §C row against the tree at run-phase start. One row is expected to have
   **changed** by then: `moai contract sign` exists once A1 lands (§C.2). The template autonomy
   block is **not** on this list — D1's closure re-pointed `AC-AP-013` at the `paths:`-scoped rule
   file and template mirror *this change* creates (§C.10), so no criterion waits on A1's block. A
   changed row is a premise update recorded in progress.md, not a silent adjustment.
3. Confirm `moai contract show --json` exists and emits `push_requires_lease` — the contact point
   A1 declares at `25283ebf8:…/spec.md:410`. Both were measured **absent** at `553e224f3`
   (spec.md §C.6). Where A1 renamed or dropped either, stop and report rather than inventing a
   substitute or falling back to parsing `contract.yaml`.
5. Re-read A1 at whatever commit is then current and confirm the coordinates this SPEC pins at
   `25283ebf8` still say what §C quotes (`:72-74`, `:79-85`, `:94-98`, `:128-133`, `:142-148`,
   `:155-156`, `:329`, `:410`, `:462`, and `design.md:416-426`, `design.md:470-494`). A moved
   statement is a premise update recorded in progress.md, and the pin in the SPEC body is then
   advanced with the quotation re-read, never silently re-pointed. **A1 advanced twice while this SPEC
   was in plan phase** (v0.5.0 → v0.5.1 → v0.5.2), and the v0.5.1→v0.5.2 diff shifted later lines
   **non-uniformly** (+1 at the § Out of Scope heading, +3 at `:329` / `:410` / `:462`, and `:231`
   added with no v0.5.1 counterpart; 520 → 524 lines) — so a constant offset would not rescue a
   stale pin, and this step is not ceremony.
6. Confirm the two `internal/mission` facts the projection is written against still hold:
   `targetInsideScope` is exact-or-prefix (`internal/mission/policy.go:112-118`) and
   `contractComplete` still requires `Approved` and `MaxOperations > 0`
   (`internal/mission/contract.go:55-61`). A change in either is a premise update, not a silent
   adjustment of the mapping.
7. Re-measure `grep -rn 'MOAI_FACTORY_ROLE' internal cmd pkg`. It was **0 rows** at `eee5f635e`. A
   non-zero count means card t1240 landed first — read what value it stamps before implementing
   REQ-AP-012's value constant (spec.md §F O5).
4. Capture the `golangci-lint run` baseline on this tree, so "no new issue" in acceptance.md §H is
   attributable.

## §D — Constraints

Carried from spec.md §E: no new lease mechanism (C1); opposite fail directions are deliberate (C2);
neither component writes an escalation record (C3); activation ordering is A1's (C4); template
neutrality (C5).

## §E — Self-verification

Each milestone closes only with the named tests present as `--- PASS:` lines and a non-empty swept
count, plus the cross-platform build. The full gate list is acceptance.md §H.

## §F — Milestones

### M1 — Push serializer (highest reversibility cost: it writes a shared lease record)

- REQ-AP-001, REQ-AP-002, REQ-AP-007; design.md §B.
- Activation triple (mode + action + `push_requires_lease`), `git push develop` matcher, admit /
  deny / release / reclaim, fail-open.
- ACs: AC-AP-001, AC-AP-002, AC-AP-003, AC-AP-004.
- First because it touches a record other sessions read, so a wrong shape here is the most
  expensive to unwind. It is, however, the milestone that **does** depend on A1's `show --json`
  (pre-flight step 3); where A1 has not landed it, run M2 first — the order is a
  reversibility preference, not a dependency.

### M2 — Contract-sign and contract-decide guard (the parsing decision, plus the rule split)

- REQ-AP-003, REQ-AP-004, REQ-AP-005, REQ-AP-009, REQ-AP-011, REQ-AP-012; design.md §C.
  (REQ-AP-006 withdrawn.)
- **Order inside the milestone, because one step unblocks the others:** define the
  `MOAI_FACTORY_ROLE` name and value constants in `internal/config/envkeys.go` **first**
  (REQ-AP-012) — until they exist, the role-scoped rule has nothing to read and AC-AP-015 / AC-AP-016
  cannot be written against a constant.
- Then: quote-removing word split, assignment and wrapper stripping, basename match, the verb match
  (`sign` / `decide`), one-level `-c`, the **rule-selection step** (design.md §C.2 step 6: human-path
  `sign` → REQ-AP-003 unconditional; `--signer llm` / `llm+jev` or `decide` → REQ-AP-011 role gate),
  and the rule-scoped fail-closed branch (REQ-AP-004). **No receipt branch** — `--receipt` is never
  read (design.md §C.5).
- ACs: AC-AP-005 .. AC-AP-010, AC-AP-015, AC-AP-016, AC-AP-017.
- This milestone depends on **nothing from A1 and nothing from card t1240**: it reads a command line
  and one environment variable it defines itself, so it can be implemented and closed while A1 is
  still in flight.
- Second because the matcher's shape and the three-way rule split are the card's other genuine design
  decisions, and every later step assumes them.

### M3 — Documentation of the mode-independent deny (finding N8)

- REQ-AP-010; design.md §C.6a, §C.7.
- Create the `paths:`-scoped rule file documenting this guard, plus its
  `internal/template/templates/` mirror (Template-First, `CLAUDE.local.md` §2), then `make build`.
- AC: AC-AP-013.
- **Depends on nothing outside this SPEC**, which is the v0.1.2 change: both files are created here.
  v0.1.1 pointed this milestone at A1's template autonomy block and said "M3 waits" — that wait was
  the finding-N5 recurrence the iter-1 audit found (spec.md §C.10). The template-comment limb and the
  REQ-AE-001 re-wording are transferred to A1 and A2 (spec.md §H.5); this milestone no longer waits on
  either.

### M4 — Mission-validator projection (mapping, now that A1 has drafted it)

- REQ-AP-008; design.md §D, spec.md §C.11.
- One-way projection over A1's adopted 12-row mapping plus the three corrections: populate `Approved`
  and a positive `MaxOperations`; translate a trailing `/**` to a prefix and fail closed on an
  inner-wildcard glob; carry an enumerated deliberately-not-projected list with `card` on it. Fail
  closed on any other unmapped field; mission's exported surface untouched.
- AC: AC-AP-014.
- **Capture the `internal/mission` exported-surface baseline at this milestone's start** — AC-AP-014
  limb 5 compares against it, and a baseline captured later is not a baseline.
- Last because the direction decision is recorded in design.md §D and the remaining work is mapping.
  It consumes A1's `show --json` only as a fixture, so it does not wait on A1 either.

## §G — Anti-patterns

- Scrubbing quoted spans instead of removing quotes — it erases `'moai'` and defeats the matcher
  (design.md §A).
- Weakening the sign deny to preserve the `guided` sentence (design.md §C.7).
- Widening `internal/mission`'s exported types to fit the contract (design.md §D).
- Making the sign deny conditional on the verb existing — the deny is what makes it unreachable
  (AC-AP-008).
- Denying `decide` in a session with no role marker — that is the lead's own path, and denying it is
  the v0.1.1 error this revision corrects (spec.md §C.3, design.md §C.1).
- Reading the `MOAI_FACTORY_ROLE` literal instead of the constant REQ-AP-012 defines (AC-AP-017).
- Widening the deliberately-not-projected list to silence a fail-closed error — the list is enumerated
  for `card` and a considered addition, never as an escape hatch (REQ-AP-008).
- Translating an inner-wildcard glob into "the nearest prefix" — that silently widens scope, which is
  the hazard REQ-AP-008 fails closed on.
- Asserting wrapper coverage from the list alone, without the unknown-wrapper case (AC-AP-010).
- Running the full local suite: change-scoped packages here, CI for the full verdict
  (`CLAUDE.local.md` §4).

## §H — Risks

| Risk | Mitigation |
|---|---|
| A1 renames or drops `push_requires_lease`, or `show --json` does not arrive | Pre-flight step 3 re-checks both; absent → the serializer is inactive (AC-AP-002 (d)), which is a correct state, not a silent hole. M2 proceeds regardless |
| A pinned A1 quotation moves at a later A1 revision | Pre-flight step 5 re-reads all five coordinates; a move is a recorded premise update, never a silent re-point |
| ~~A1's autonomy template block has not landed when M3 is reached~~ | **No longer a risk at v0.1.2** — M3 documents the guard in a rule file this SPEC creates, so it does not read A1's block at all (spec.md §C.10, design.md §C.6a) |
| Card t1240 never stamps `MOAI_FACTORY_ROLE`, or stamps a different value | REQ-AP-011 denies nothing in production and AC-AP-016's allow arm still passes. Recorded as spec.md §E C7 + §F O5, **not** designed around: the human-path deny (REQ-AP-003) is keyed on the tool-call boundary and is unaffected |
| A1's drafted mission mapping is revised after adoption | The three corrections are measured against `internal/mission` source, not A1's prose. A changed *field source* is caught by pre-flight step 6 |
| A wrapper list that looks complete but skips options wrongly | design.md §C.3 states each wrapper's own options; AC-AP-009 asserts the resolved program is named |
| The receipt path stays unusable through A3 | Stated as residual risk (design.md §F), not designed around |
| `internal/cli` package test latency (10-minute timeout) | Scope test runs to the four packages named in acceptance.md §H |

## §I — Open questions for the lead

- **O1** [RESOLVED] — disposition at spec.md §F O1, which is the single canonical copy: `decide` is in
  scope, gated by the role-scoped rule REQ-AP-011, and owned by **A3 (card t1236)**. Read it there
  rather than here; a second copy is what let the v0.1.1 wording ("denied outright", "owner
  unidentified") survive the ruling that superseded it.
- **O2** [RESOLVED as to scope] — disposition at spec.md §F O2, likewise canonical: the lead's
  three-way split is adopted in full (§C.3) — the tool-call boundary for the human path, the role
  marker for the non-interactive path and `decide`, and an **allow** for a session carrying no
  marker. The marker's name and value constants are defined by **this SPEC** in
  `internal/config/envkeys.go` (REQ-AP-012), so nothing here waits on card t1240. What remains is a
  residual exposure, not a question (§E C7).
- **O3** [RESOLVED — lead ruling 2026-09-26.] SPEC body register stays **English**. The first dispatch
  asked for Korean prose *and* for matching the sibling SPEC's register, and the sibling
  (SPEC-AUTONOMY-ESCALATION-001) is English technical prose; the lead resolved the contradiction in
  favour of English, because this SPEC cites requirements and criteria across the AUTONOMY family and
  cross-SPEC comparison by reading and by grep needs one language. Korean stays only in the
  「A1 plan-audit 통과본으로 재확인」 dependency tag. Mirrored in spec.md §F O4.
- **O4** Audit-line sink — its own file, or shared with `.moai/logs/branch-guard-audit.log`
  (spec.md §F O3).
- **O5** [OPEN — one confirmation, with a concrete failure mode.] The `MOAI_FACTORY_ROLE` value: the
  lead ruled `agent`, which is the **retired** spelling of the CLI role token whose live spelling is
  `worker` (`internal/cli/factory.go:59`, `:64`). Implemented as ruled; if card t1240 stamps `worker`,
  REQ-AP-011 protects nothing while every criterion still passes. Confirm the value and carry it to
  t1240 (spec.md §F O5).

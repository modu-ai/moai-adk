# SPEC Review Report: SPEC-KANBAN-QUEUE-PR-SYNC-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.60** (harmonic mean; Tier M PASS threshold 0.80)

Reasoning context ignored per M1 Context Isolation. The dispatch's framing of the
SPEC's own claims was not treated as evidence; every judgement below rests on the
artifact files plus commands re-run in this worktree.

## Pin verification and finding-attribution audit

**Pin verified.** The lead re-pinned content-addressed after discovering the
original freeze was defective:

```
$ git rev-parse HEAD
06f4597aa471a79f7594e3bc913e74463c30513e
$ git rev-parse HEAD:….../spec.md
313a5c31af19e7dae963a9fd6478ecd566331f82
$ git status --short
?? .moai/reports/t210/verdict.md      # this report; no tracked drift
```

**Which findings were formed against superseded content: none of the graded
ones.** Established mechanically rather than asserted. Blob ids of all four
artifacts at the pin:

```
spec.md        313a5c31af19e7dae963a9fd6478ecd566331f82
plan.md        90ca25ebf2d04ba381490c04a8f95f41d42a66df
acceptance.md  67b243afa55fda9defda906712641a01b310a6c4
progress.md    542d98ac0e9e1fd250800dbde223afda51164fda
```

The `git diff` I captured during the first pass recorded
`spec.md index c985425d3..313a5c31a` and `plan.md index 61a5997f9..90ca25ebf` —
the **after** sides are byte-identical to the pinned blobs, so my reads of both
were post-edit, not mixed. `acceptance.md` never appeared in `git status` and is
unmodified across both commits. `progress.md` is the file that changed under me
mid-read; no graded finding depends on its content.

Blob identity is necessary but not sufficient, so **every cited line was re-read
at the pin** (`sed -n '13p;137p;146p;172p;189p;236p'` on spec.md;
`'38p;50p;66p;76p;87p;140p;154p'` on acceptance.md; `'59p;87p'` on plan.md).
All fifteen citations resolve to exactly the quoted text. **Every finding below
survives re-verification at this pin.** One finding — D2 — is materially
*strengthened*, and its recommended fix is partly **retracted** on new
measurement; one new finding (D16) is added.

**The process failure is worth recording, and it is the lead's, not the
auditor's: a HEAD-pinned freeze cannot detect uncommitted working-tree drift.**
HEAD never moved, so the check kept passing while three files changed underneath.
Only a content hash — a blob id, or `git status --short` returning empty — closes
that hole. Adopt the content-addressed form for every future audit dispatch.

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — 19 leaf requirements, ids
  `REQ-1.1..1.8`, `REQ-2.1..2.7`, `REQ-3.1..3.4`. Verified:
  `grep -oE 'REQ-[0-9]+\.[0-9]+' spec.md | sort -u -V` → no gaps, no duplicates,
  consistent hierarchical scheme. The scheme is `REQ-N.M` rather than the
  canonical zero-padded flat `REQ-001`; it is internally consistent, so this is
  recorded as a minor deviation (D8), not a failure.

- **[PASS] MP-2 GEARS format compliance** — **judged against the requirement
  layer (`spec.md` REQ-XXX entries) only.** All 19 leaves carry `shall` /
  `shall not` and match a GEARS pattern: Ubiquitous (REQ-1.1, 1.8, 2.1, 2.2,
  2.5, 2.6, 2.7, 3.2, 3.3, 3.4), Where (REQ-1.2), Where+While compound
  (REQ-1.3, 1.4 — PASS-equivalent per the compound-clause rule), When (REQ-1.7,
  2.3, 3.1), Unwanted `shall not` (REQ-1.5, 1.6, 2.4). The Given-When-Then
  entries in `acceptance.md` are the correct verification-layer format and were
  **not** graded here; they are graded under Group 4. Minor defects D9 and D10
  are recorded but do not fail this criterion.

- **[FAIL] MP-3 YAML frontmatter validity** — see D1. `spec.md:13` is
  `tags: [kanban, todo, github, observability, read-only]`, a YAML **sequence**
  where the canonical schema (`spec-frontmatter-schema.md` § Field Reference)
  types `tags` as a comma-separated **string**. Verified with the domain tool,
  not by inference:

  ```
  $ moai spec lint .moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md --json
  [{"file":".moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md","line":1,
    "severity":"error","code":"ParseFailure",
    "message":"SPEC parsing failed: frontmatter parsing error: YAML parsing error:
    yaml: unmarshal errors:\n  line 13: cannot unmarshal !!seq into string"}]
  ```

  This is worse than a single-field finding: the document **does not parse at
  all**, so every other lint rule the linter advertises (EARS modality, REQ-ID
  uniqueness, AC→REQ coverage, Out-of-Scope presence, dependency DAG, zone
  cross-refs) is silently unevaluated on this SPEC. A green-looking absence of
  further findings here is an absence of evaluation, not an absence of defects.

  Corroboration that this is an isolated authoring slip rather than a schema
  disagreement: both sibling SPECs use the string form —
  `SPEC-KANBAN-TODO-CLI-001` → `tags: kanban, cli, backlog, concurrency`;
  `SPEC-KANBAN-BOARD-001` → `tags: "kanban, board, column, …"`.

- **[N/A — auto-pass] MP-4 language neutrality** — the SPEC is scoped to this
  single-language project's own tooling (`gh`, `internal/github/gh.go`). The one
  artifact it mirrors into the distributed template (a PR-title naming clause
  plus a pre-dispatch cross-check) is language-neutral prose naming no
  language-specific toolchain. No 16-language enumeration obligation arises.

- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -oE
  'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u` returns exactly one id, the
  SPEC's own. No external SPEC reference exists, so no retired/superseded
  reconciliation obligation arises. No BLOCKING finding.

- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → 0.
  D8-4 auto-PASS.

- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION'` across
  the SPEC directory → 0 matches. `research.md` correctly absent at Tier M.

**MP-3 fails ⇒ Verdict FAIL regardless of the dimension scores.**

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.60 | 0.50 band (multiple requirements need interpretation), credited upward for otherwise-unusual precision | D3 (`spec.md:189` REQ-3.1 "rather than dispatching"), D5 (`spec.md:172` REQ-2.5 MAY branch), D2 (`spec.md:137` REQ-1.7 merged-PR silence) |
| Completeness | 0.60 | 0.50 band — every section present, but the evidence base does not cover the scope two requirements claim from it | 5 `### Out of Scope` H3s with bullets (verified, count 5); all 12 fields present but `tags` mis-typed (D1); **D16 (REQ-1.5 generalizes M2 past the question M2 measured)**; D2 (t199 absent from §A, and the merged extension is now measured not to fix it); D6 (same-file [HARD] precedent uncited) |
| Testability | 0.50 | 0.50 band — several ACs are not binary-testable as written; **two must-severity ACs are unsatisfiable** | D4 (`acceptance.md:38-41` AC-002, `:45-53` AC-003 — both premises refuted by live re-measurement), D7 (`:66` AC-004 under-covers REQ-2.2), D11 (`:76` AC-005 disjunctive assertion), D12 (AC-011/AC-012 grep tails untested) |
| Traceability | 0.75 | 0.75 band — mapping is nominally complete, two mappings non-functional | `acceptance.md:183-203` D.2 covers all 19 REQs; but AC-003 cannot exercise REQ-1.4/REQ-1.6 and AC-002 cannot exercise REQ-1.3 (D4); REQ-2.5's MAY branch uncovered (D5) |

Harmonic mean: `4 / (1/0.60 + 1/0.60 + 1/0.50 + 1/0.75)` = **0.600**.

(First pass scored Completeness 0.70 for an aggregate of 0.622. The re-audit's
D16 — a requirement whose stated grounds do not support its scope — lowers
Completeness to 0.60. The verdict was already FAIL on MP-3 and did not turn on
this.)

## Independent verification of the cited measurements

The SPEC leans on M1..M6. Everything cheap to re-run was re-run in this worktree.

**M1 — reproduces exactly.** `CLAUDE_PROJECT_DIR=<primary> moai todo list`
cross-read against `gh pr list --state open`: t200 / t201 / t202 / t203 / t205
are all still `queued` while carrying open PRs #1612 / #1611 / #1613 / #1614 /
#1617; t88 is `picked` with #1619. The five-card divergence is real, not assumed.

**M2 — the load-bearing claim is CONFIRMED.** The dispatch flagged the #1600
commit-token claim as collapse-if-refuted for REQ-1.5. It survives:

```
$ gh pr view 1600 --json body -q .body   | grep -oE '\bt[0-9]{1,4}\b' | sort -u
t184
$ gh pr view 1600 --json commits -q '.commits[].messageHeadline' | grep -oE '\bt[0-9]{1,4}\b' | sort -u
t127 t158 t159 t164 t165 t170 t171 t173 t81 t82 t83 t89
$ gh pr view 1600 --json commits -q '.commits|length'
100
```

t184 is in the body and **absent from every commit-message token**. REQ-1.5's
grounds hold. (Count discrepancy: the record states 15 tokens; a headline-only
scan yields 12 — the record evidently scanned full messages. This does not touch
the claim, but it does touch AC-006, which pins the number 15 — see D13.)

Title-carrier recall also re-verified at 7/11 on the record's PR set
(1621, 1617, 1614, 1613, 1612, 1606, 1605 carry a title token; 1619, 1611, 1601,
1600 do not). The open set has since grown to 13 PRs; the 11 in the record are
unchanged.

**Zero-baselines for AC-011 / AC-012 — independently confirmed.**

```
$ grep -c 'pre-dispatch PR cross-check' .claude/rules/moai/workflow/kanban-dispatch.md
0
$ grep -c 'PR title MUST carry the delivering card id' .claude/rules/moai/workflow/kanban-dispatch.md
0
```

**Plan pre-flight targets all exist**: `internal/github/gh.go`,
`internal/cli/todo.go`, `internal/cli/todo_why.go`,
`internal/kanban/backlog_store.go`, and both template mirrors
(`…/templates/.claude/rules/moai/workflow/kanban-dispatch.md`,
`…/templates/.claude/skills/moai/workflows/todo.md`).

## Ruling on the read-only design decision (spec.md §B)

**The ruling is sound, not merely convenient.** I read both cited [HARD] clauses
in full at `kanban-dispatch.md:27` and `:29`. "The lead is the queue's sole
producer" governs *admission* ("nothing enters the queue the operator did not ask
for"); "Promotion is the operator's act, always" governs *the pick*. A surface
that computes a link and prints it does neither, and the SPEC's own §B.2 argument
— that a state change nobody made destroys operator provenance and fails
silently — is the correct reason, independent of convenience.

The strongest counter the dispatch raised (does REQ-3.1 hand the tool de-facto
promotion authority?) **partly lands**, but not against the read-only ruling —
against REQ-3.1's wording. Recorded as D3.

Two supporting notes:

- The SPEC understates its own case. `kanban-dispatch.md` carries a **third**
  [HARD] clause in the same section it proposes to edit — *"The lead may attach a
  finding; it may not act on one… Analysis changes exactly one thing on its own
  authority — it refuses the admission of a card whose normalized text is
  identical to one already queued or picked, which creates no card and **leaves
  the queue file byte-identical**."* That is the precedent, in the same file,
  with the same byte-identity language REQ-2.1 adopts. §B cites the weaker
  `todo.md` restatement instead and never mentions it (D6).
- The rejected alternative (§B, auto-update on merge/close) is correctly rejected
  and correctly named as weighed-not-overlooked. It is also option (b) of the
  three the originating card itself proposed, so the record of consideration is
  complete on that axis.

## Defects Found

**D1. MP3-FRONTMATTER — `.moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md`:L13 —
`tags:` is a YAML sequence where the canonical schema types it as a
comma-separated string; `moai spec lint` returns `ParseFailure` (error severity)
and the document does not parse, silently suppressing every other lint rule —
Severity: critical — Class: blocking — Required fix:** change L13 to
`tags: "kanban, todo, github, observability, read-only"`, then re-run
`moai spec lint <spec.md> --json` and record the verbatim output (an empty array)
in `progress.md §E.1`. The zero-findings claim is only meaningful once the file
parses.

**D2. SCOPE-MOTIVATION — `spec.md`:L34-38 (§A) and L234-241 (Out of Scope —
merged and closed) — the SPEC scopes out the half of its own motivating incident
that had the highest measured cost, and does not disclose that it is doing so —
Severity: major — Class: blocking — Required fix:** the originating card t210
records two failure shapes, not one. §A presents only the five open-PR
divergences. The card's own text records the second: *"같은 라운드에서 t199 도
queued 였으나 수정 커밋 d9899f437 이 이미 origin/main 조상이었다 (레인 1개분
낭비)"* — a card whose delivering commit was **already an ancestor of
origin/main**, discovered only after a lane had started, costing one full lane.
That is the merged case, and it is the only sub-case with a quantified waste
figure.

The consequence compounds with REQ-1.7. After this SPEC lands in full, a lead
obeying the new [HARD] REQ-3.1 clause runs the read surface for t199, gets **no
link** (open-only query), and REQ-1.7 requires that result to be distinguishable
from `ambiguous` — but **not** from "already merged". The doctrine therefore
returns a clean all-clear on precisely the case that produced the wasted lane. A
[HARD] clause that reads as a green light on its own worst measured input is a
worse outcome than no clause, because the lead now has a documented reason to
trust it.

**Re-audit addendum — the obvious fix is measured NOT to work, and I retract it.**
My first pass offered "extend the resolver to `--state merged`" as fix option
(a). I traced t199 and that option is refuted:

```
$ gh pr list --state all --search t199 --json number,title,state
#1610 MERGED "release: v3.1.3"
#1602 MERGED "release: v3.1.3 batch — … (22 cards)"
#1606 OPEN   "feat(mcp): … card t187"
$ gh pr view 1602 --json body -q .body | grep -oE '\bt199\b'      # → empty
$ gh pr view 1602 --json body -q .body | grep -oE '\bt[0-9]{1,4}\b' | sort -u | wc -l
26
```

**t199 has no PR of its own.** Its fix commit rode into `origin/main` inside the
v3.1.3 batch PR #1602. #1602's *title* carries no card token, and its *body*
carries 26 card tokens **not including t199** — the merged-side analogue of the
#1600 finding M2 already recorded on the open side. So extending REQ-1 to
`--state merged` would find t199 through neither the title carrier (REQ-1.2) nor
the body carrier (REQ-1.3/1.4). The gap is not the query's `--state` filter; it
is the carrier set. See D16.

Revised fix (option (b) or (c); (a) is withdrawn):
(b) keep the open-only PR scope but add a requirement that a no-link result
renders as an explicit *"no OPEN pull request; merged and closed PRs and landed
commits are not checked"*, and that the REQ-3.1 doctrine clause states the same
limitation verbatim, so the lead is told what was not looked at; or
(c) adopt the landed-commit query as a second, separate carrier for the distinct
"is this card already in `origin/main`?" question — see D16, which is where that
belongs.
Option (b) is the cheapest and preserves the exclusion honestly; (c) is the one
that actually closes the incident. Silence remains the one unacceptable outcome:
the current text neither covers the case nor admits the gap.

**D3. REQ-3.1-AMBIGUITY — `spec.md`:L185-189 — "the lead shall surface that fact
to the operator rather than dispatching" is ambiguous between a lead-side veto
and a report-and-re-decide, and the distinction is exactly the queue-authority
boundary §B spends a page defending — Severity: major — Class: blocking —
Required fix:** state which it is. Under `kanban-dispatch.md:29`, the operator has
already picked the card; a lead that then withholds dispatch has overridden an
operator act, which is the de-facto-authority hazard the read-only ruling exists
to avoid. Under the reading the SPEC almost certainly intends — report the PR,
let the operator re-decide — the ruling holds cleanly. Rewrite as: *"the lead
shall report the open PR to the operator and shall not dispatch until the
operator, so informed, confirms or withdraws the card"*, and mirror that wording
into the doctrine clause. AC-011 cannot detect which reading was implemented, so
the wording is the only control.

**D4. AC-FIXTURE-REFUTED — `acceptance.md`:L36-41 (AC-002) and L45-53 (AC-003) —
both ACs state a Given that is false against the M2 table they were transcribed
from and against live re-measurement; between them REQ-1.3, REQ-1.4, and REQ-1.6
end up with no working acceptance criterion — Severity: critical — Class:
blocking — Required fix:** live measurement of every open PR body:

```
$ gh pr list --state open --limit 40 --json number,body \
    -q '.[] | "\(.number): \([.body|scan("\\bt[0-9]{1,4}\\b")]|unique|join(" "))"'
1614: t1 t151 t203 t69 t9
1612: t200 t201
1611: t201
…
```

- **AC-002 is refuted.** Its Given asserts "PR #1611 (no title token, body token
  `t201`) **and no other open PR whose body carries `t201`**". #1612's body
  carries `t201` — as M2's own table row for 1612 already records
  (`body tokens: t200,t201`). Under REQ-1.4, `t201` is therefore **`ambiguous`
  across {1611, 1612}**, not `inferred`. AC-002 asserts the wrong label for its
  own fixture.
- **AC-003 is refuted in the opposite direction.** It resolves card `t151`
  expecting `ambiguous` "which also appears in another PR body". `t151` appears
  in exactly **one** body (#1614) and no title. Under REQ-1.3 it resolves
  `inferred`. AC-003 cannot produce `ambiguous`, so REQ-1.4 and REQ-1.6 — the
  no-best-guess rule, the single most consequential behavioural commitment in
  REQ-1 — are untested.

The repair is nearly free and improves the SPEC: **`t201` is the genuine
ambiguous case in the measured set.** Rewrite AC-003 around `t201` /
{#1611, #1612}, and rebuild AC-002 around a token that is genuinely
single-bodied and title-absent — `t188` (#1601 only) or `t184` (#1600 only)
both qualify against the live measurement above.

**D5. REQ-2.5-MAY — `spec.md`:L172-174 — a `MAY`-scoped optional feature
(`moai todo list --pr`) sits inside a normative requirement, has no acceptance
criterion, and no decided disposition — Severity: major — Class: blocking —
Required fix:** this violates RQ-5 (no `may` in normative text) and leaves an
implementer free to build or skip a user-visible flag. Either promote it to its
own requirement with an AC, or move it to a "possible follow-up" note outside
§D. Note the interaction with AC-009, which asserts `moai todo list` spawns zero
`gh` processes: if `--pr` ships, that assertion needs a companion asserting the
flag is what gates the spawn, or AC-009 silently becomes a no-flag-only claim.

**D6. PRECEDENT-OMISSION — `spec.md`:L59-68 (§B.1) — the closest and strongest
precedent for the read-only ruling lives in the very file the SPEC proposes to
edit and is never cited — Severity: minor — Class: optional — Required fix:** cite
`kanban-dispatch.md` § Entry into the board is an operator act, third [HARD]
clause ("The lead may attach a finding; it may not act on one… leaves the queue
file byte-identical"). It is a stronger citation than the `todo.md` restatement
currently used, it is in the file being amended, and its byte-identity language is
the source of REQ-2.1's.

**D7. AC-004-COVERAGE — `acceptance.md`:L55-69 — AC-004 is necessary but not
sufficient to enforce the §B read-only ruling; REQ-2.2's "any queue-owned file"
is not covered — Severity: major — Class: blocking — Required fix:** AC-004 is the
best-constructed AC in the set (digest, not "no error"; extended over the
fail-open and ambiguous paths). What it does **not** observe: a lock file taken
and released around the read; a sidecar or cache written elsewhere under
`.moai/state/`; a `findings[]` entry (byte-identity covers this only because
`findings[]` lives inside `backlog.json` — it would not cover a `findings`
sidecar); an mtime-only touch on a neighbouring queue-owned file. REQ-2.2
explicitly forbids caching "into any queue-owned file", and no AC observes that
wider surface. Strengthen AC-004 to assert a directory-level digest over
`.moai/state/kanban/` — before/after, recursive — rather than a single file, and
assert no new path appears. Declaring one AC "load-bearing" raises the bar for
that AC rather than lowering it for its neighbours.

**D8. REQ-NUMBERING — `spec.md`:L115-213 — hierarchical `REQ-N.M` ids instead of
the canonical flat zero-padded `REQ-XXX` — Severity: minor — Class: optional —
Required fix:** the scheme is internally consistent and traceable, so this is
cosmetic. Flag only because downstream grep-based tooling that assumes
`REQ-[0-9]{3}` will not match. Note that the linter could not report on this
because the file does not parse (D1).

**D9. GEARS-WHERE-SEMANTICS — `spec.md`:L120, L123, L126 — `Where` is used for a
runtime data condition, where GEARS reserves it for a capability gate / feature
flag / static config — Severity: minor — Class: optional — Required fix:** "Where
the PR title contains the card id token" is a property of the input data
evaluated per-resolution, i.e. a state or event, not a capability gate. The
syntactic form matches a GEARS pattern (so MP-2 passes), but `While the PR title
contains the card id token` is the semantically correct modality. Low priority.

**D10. NON-NORMATIVE-TRAILERS — `spec.md`:L146-148 (REQ-2.1), L158-160 (REQ-2.4)
— requirements carry trailing declarative sentences that restate rather than
require — Severity: minor — Class: optional — Required fix:** "It writes no field,
no `findings[]` entry, and no timestamp" and "`moai todo list` remains lock-free
and network-free unless the operator opts in" are commentary inside a normative
block. Either give them `shall` form or move them to prose. They are currently
the only statement of the `findings[]` and lock-free constraints, so deleting
them would lose content — promote, do not drop.

**D11. AC-005-DISJUNCTION — `acceptance.md`:L76 — "the output renders the queue
with the link column blank **or absent**" admits two different renderings, so the
assertion cannot fail on rendering — Severity: minor — Class: optional — Required
fix:** pick one. A test that passes on either branch observes only that the
process did not crash, which AC-005's exit-code clause already covers.

**D12. AC-GREP-TAIL-UNTESTED — `acceptance.md`:L134-147 (AC-011) and L149-162
(AC-012) — the mechanical half proves a string was typed; the half that carries
the meaning has no command — Severity: major — Class: blocking — Required fix:**
on the dispatch's question of whether these are ceremony: **half.** The zero
baselines are real (independently confirmed at 0 and 0 above), so the greps are
not the vacuous "grep a word already in the prose" pattern this project has been
burned by — they are genuine tripwires against the clause being dropped or
reworded away. But each AC then appends an untested tail: *"**And** the
surrounding clause is marked [HARD] and requires the lead to read and report the
card's PR state before dispatching"* (AC-011) and *"**And** the same section
explicitly states that this does not contradict the branch-name exclusion rule"*
(AC-012). Those tails are the entire substance of REQ-3.1 and REQ-3.3, and they
are human judgement with no verb attached. Fix: split each into a mechanical part
(the grep, keeping the zero baseline) and an explicit reviewer-judgement part
recorded as such, so the SPEC does not claim mechanical coverage it does not
have. Do not delete the greps — keep them and stop over-claiming them. Add a
mechanical assertion for the [HARD] marker itself: `grep -c '\[HARD\].*pre-dispatch
PR cross-check'` is checkable where "requires the lead to read and report" is not.

**D13. AC-006-FIXTURE-NUMBER — `acceptance.md`:L86-91 — AC-006 pins "15
commit-message tokens" for #1600 and resolves card `t131`, neither of which
reproduces from an independent scan — Severity: minor — Class: optional —
Required fix:** a headline-only scan of #1600's 100 commits yields 12 distinct
tokens and does **not** include `t131` (the record's 15 evidently came from full
commit messages). `plan.md` §B correctly instructs pinning the fixture from the
record rather than re-fetching, which mitigates this — but an AC that cites a
number no stated command reproduces will confuse the implementer. Either name the
exact scan (`--json commits -q '.commits[].messageBody'` vs `messageHeadline`) or
drop the count and cite only the property that matters: t184 is absent from the
commit tokens and present in the body.

**D14. TIER-BUDGET — `spec.md` §D — 19 leaf requirements against the Tier M
ceiling of 16 — Severity: major — Class: blocking — Required fix:** verified:
`grep -cE '^\*\*REQ-[0-9]+\.[0-9]+\*\*' spec.md` → 19. `spec-workflow.md`
§ SPEC Complexity Tier caps Tier M at 16 requirements and 16 acceptance criteria,
applied independently; the AC count (13) is within budget, the REQ count is over
by 3. Per that rule, over-budget is "a signal to tier up or to split the SPEC,
not to relax the budget".

On the dispatch's wider tier question — a Go resolver, a new CLI verb, two
doctrine edits, and a template mirror — **Tier M is the right call on every other
axis**: the file count is well under 15, the LOC estimate is plainly under 1000,
and the 3-artifact set is correct. The clean split is to lift REQ-3 (the doctrine
change, 4 requirements, no code) into its own Tier S SPEC. That also fixes a real
ordering awkwardness: `plan.md` M1 ships a [HARD] behavioural rule binding every
future card-delivering PR *before* the tooling that serves it exists, which is
correct on reversibility grounds but means the doctrine is live and unserved for
the whole of M2-M4. Two SPECs make that sequencing explicit instead of implicit.

**D15. MILESTONE-MIRROR-SPLIT — `plan.md`:L58-60 (M1) and L87-88 (M4) — the
Template-First mirror obligation is split across two milestones but only M1
carries a parity AC — Severity: minor — Class: optional — Required fix:** M4
mirrors `.claude/skills/moai/workflows/todo.md` into the template; AC-013 only
covers the `kanban-dispatch.md` mirror. Extend AC-013 to both files, or give M4
its own parity exit. Both mirror targets exist and were confirmed present.

Milestone ordering is otherwise sound: M2 (pure resolver) → M3 (read surface
consuming it) → M4 (wiring) is a genuine dependency chain, and M1's
reversibility-first placement is argued rather than assumed.

**D16. REQ-1.5-OVERGENERALIZED — `spec.md`:L130-132 (REQ-1.5) and L100-105 (§C) —
M2 measured the commit carrier against one question and REQ-1.5 forbids it for
all questions, including the one where it is the only carrier that works —
Severity: major — Class: blocking — Required fix:** raised by the lead, tested
independently, and **it holds.**

M2 asked *"which cards does this PR deliver?"* — the attribution direction, where
#1600's inherited tokens are genuinely ruinous. REQ-1.5 then bans the carrier
outright: *"The resolver shall not consult commit messages as a linking carrier."*
But the t199 incident poses a **different** question — *"is this card's work
already in `origin/main`?"* — and in that direction inherited commits are not
noise at all: a commit that rode in on another branch is still genuinely landed,
so the property being tested is the very thing the query returns.

Measured, both directions:

```
$ git log origin/main --perl-regexp --grep='\bt199\b' --oneline
b4b8bdfbe docs: update CHANGELOG for v3.1.3
711bfdbba merge(t199): internal/web 자기-SIGTERM TOCTOU …
d9899f437 fix(web): register signal handling before binding the listener (t199)
$ git merge-base --is-ancestor d9899f437 origin/main && echo ANCESTOR=yes
ANCESTOR=yes

# false-positive control — t205 is queued with an OPEN PR, i.e. NOT landed:
$ git log origin/main --perl-regexp --grep='\bt205\b' --oneline     # → empty
# positive controls — landed cards return their delivering commits:
$ git log origin/main --perl-regexp --grep='\bt82\b'  --oneline | head -2
32434bf9b docs(reports): record the t82 diet cross-reference sweep … (card t190)
705f21f64 merge(t82): AGENTS.md canonical contract layer …
$ git log origin/main --perl-regexp --grep='\bt165\b' --oneline | head -1
b86fac07c Merge WT-svg-quality: SPEC-SVG-QUALITY-ABSORB-001 (card t165)
```

The landed-direction query returns the right answer on every sample: the delivering
commit for landed cards, and **empty for a not-landed card**. It is not
noise-free — t82's first hit is a different card's report commit that merely
mentions t82 — but that noise does not produce a wrong answer to the landed
question, because t82 *is* landed. Contrast the attribution direction, where the
same carrier returns 15 tokens for #1600 and omits the one card it delivers.

So the SPEC generalized a real measurement past the question that measurement
addressed, and the over-generalization is load-bearing: combined with D2, the
only carrier that finds t199 is the one REQ-1.5 categorically forbids, and the
merged-PR extension that looks like the natural fix does not reach it either
(measured above). **REQ-1 as written would not have prevented one of the two
incidents in the card's own motivation.**

Fix: scope REQ-1.5 to the question M2 actually measured — *"the resolver shall not
consult commit messages **when attributing an open PR to a delivering card**"* —
and either (i) add a separate landed-check requirement using
`git log origin/main --perl-regexp --grep`, or (ii) state explicitly that the
landed question is out of scope and that REQ-1.5's ban does not prejudge it. What
must not survive is a blanket prohibition justified by a measurement of a
narrower case.

**D17. GREP-ERE-BOUNDARY — implementation hazard for any landed-check —
Severity: major (conditional on D16 (i) being adopted) — Class: blocking if D16(i)
is adopted, otherwise optional — Required fix:** `\b` is not POSIX ERE, and git's
`--grep` fails **silently** rather than erroring:

```
$ git log origin/main --perl-regexp --grep='\bt199\b' --oneline | wc -l
3
$ git log origin/main -E --grep='\bt199\b' --oneline | wc -l
0
```

The `-E` form returns empty, which is byte-indistinguishable from "this card is
not landed" — a false negative that an AC would observe as a pass. This is the
same vacuous-grep failure mode the project has been bitten by before. Any
requirement or AC in this family MUST specify `--perl-regexp` explicitly and MUST
carry a positive control (a known-landed card returning non-empty) so the
regex-engine failure cannot masquerade as a clean result.

**D18. REQ-1.8-CONFIRMED — `spec.md`:L141-142 — no defect; recording the positive
verification — Severity: n/a — Class: n/a:** REQ-1.8's whole-token requirement is
mechanically satisfiable and was verified against a tree containing t200, t205,
and t206: `git log origin/main --perl-regexp --grep='\bt20\b' --oneline` returns
zero hits. AC-008 is sound as written. Recorded so the run phase does not
re-litigate it.

## Recommendation

FAIL. Fix in this order; 1-3 are cheap and 4 is the one that needs a decision.

1. **D1** — one-line frontmatter fix, then re-run `moai spec lint` and paste the
   verbatim output. Until the file parses, no other automated check on this SPEC
   has actually run, and any "lint clean" claim is unattributed.
2. **D4** — rebuild AC-002 and AC-003 against the live carrier data quoted above.
   Use `t201`/{#1611,#1612} for ambiguous and `t188` or `t184` for inferred. This
   is a transcription error, not a design error, and REQ-1.4/1.6 are untested
   until it is fixed.
3. **D3, D5, D7, D12** — wording and coverage repairs, each self-contained:
   disambiguate REQ-3.1's "rather than dispatching"; resolve the REQ-2.5 `MAY`;
   widen AC-004 to a directory digest; split AC-011/AC-012's mechanical and
   judgement halves.
4. **D2 + D16 + D14 — the ones that need an operator-level decision.**
   - **D2 and D16 are one decision, not two.** t199 is unreachable by every
     carrier REQ-1 specifies, and measurably still unreachable if the resolver is
     extended to merged PRs. Either adopt the landed-commit carrier as a second,
     separately-scoped question (D16 fix (i), with D17's `--perl-regexp` +
     positive-control guard attached), or state the limitation in both the render
     and the REQ-3.1 doctrine clause so the lead is told what was not checked
     (D2 fix (b)). Shipping a [HARD] pre-dispatch clause that returns a silent
     all-clear on the incident sub-case that already cost a full lane is the one
     outcome to avoid.
   - D14: decide whether to split REQ-3 into its own Tier S SPEC (which resolves
     the budget overrun and the doctrine-before-tooling sequencing together) or
     to tier up.

Optional findings D6, D8, D9, D10, D11, D13, D15 are surfaced for the
orchestrator's discretion and are **not** grounds for the FAIL. D18 is a
confirmation, not a defect. The verdict rests on the MP-3 failure (D1), which is
score-independent, and is corroborated by an aggregate of 0.60 against a Tier M
threshold of 0.80.

## On the lead's t199 evidence

Accepted, and the direction is right. I re-ran every claim rather than taking
them: t199 has no PR of its own (only release PRs #1602/#1610 surface it), the
three `--perl-regexp` commits reproduce, `d9899f437` is an ancestor of
`origin/main`, the `-E` form returns empty, and the `\bt20\b` control returns
zero. Two additions the evidence did not include, both of which strengthen it:
**#1602's body carries 26 card tokens and t199 is not among them**, which kills
the merged-PR extension as a fix and forced me to retract my own recommendation;
and the landed-direction query passes a false-positive control (`t205`, queued
with an open PR, returns empty), which is what makes the direction argument
sound rather than merely plausible.

One qualification, so the run phase is not surprised: the landed direction is not
noise-free. `t82`'s first hit is a different card's report commit that merely
mentions it. The noise does not produce a wrong answer to the landed question —
t82 is landed — but a naive implementation that reports the *first* matching
commit as "the delivering commit" would attribute wrongly. Scope the landed check
to the boolean it can answer ("does `origin/main` name this card?") and do not let
it claim attribution, which is the exact distinction D16 turns on.

A closing note in the SPEC's favour, since an adversarial report can leave a
false impression: the design ruling in §B is correct and well argued, the
measurement discipline is real (M1 and M2's load-bearing claim both survived
independent re-measurement), the rejected alternative is honestly recorded, and
AC-004 is the strongest single acceptance criterion I have audited in this
project. The failures are an unparsed frontmatter line, two mis-transcribed
fixtures, and a scoping decision that generalized a good measurement past the
question it answered. Only the last is a reasoning failure, and it is the
recoverable kind: the measurement was sound, the inference from it was drawn one
step too wide.

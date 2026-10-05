# Acceptance Criteria — SPEC-TODO-LANDING-ATTRIBUTION-001

Card t472. Measured tree `4bcac7079`; plan-audit iteration-1 remediation re-measured at HEAD
`e227871b4` against `origin/develop` `7835148d3`.

Version 0.3.0 re-measured the corpus at HEAD `75e63d6f2` against the **pinned commit** `7835148d3`
(the same tree every earlier figure was taken on; `origin/develop` has since advanced to
`25a3212a9`, which is why the pin rather than the branch name is the address — VCI §2.1 remedy R1).

Every axis-F **predicate-behaviour** criterion (AC-TLA-001..004, -003b, -007) is **bidirectional** by
lead directive: it asserts the green direction (a title-attributed commit reads `landed`) AND the red
direction (a body-mention-only, other-card-attributed, or absorb-direction-merge commit reads
`not-landed`). A criterion carrying only the green direction is satisfied by an implementation that
answers `landed` for everything, and is not accepted.

AC-TLA-005 and AC-TLA-006 are **structural** (enumeration locality, argv construction) and assert no
predicate direction; each carries a **mutation direction** instead — remove the declaration, or mutate
the builder, and the criterion fails. That is equivalent protection, and the bidirectional sentence
above deliberately does not reach them (plan-audit D8).

---

## §A Named mutants

Each mutant is an implementation the criteria must be observed **red** against before the repair is
written. The falsifying inputs are fixtures measured in this tree and re-runnable.

| Mutant | What it does | Falsifying input | Observed at `4bcac7079` |
|---|---|---|---|
| **MUT-WHOLE-MESSAGE** | Matches the card token anywhere in the commit message — today's shipped behaviour | **t237** on `origin/main` | `git log origin/main --perl-regexp --grep='\bt237\b' --oneline` returns `539349c5b`, `32d2221fa` — both **t230** commits; zero subjects attribute t237 |
| **MUT-SUBJECT-ONLY** | Matches the card token anywhere in the commit **subject** | **t216** and **t443** on `origin/develop` | `673d3d8a0 docs(t263): ... behind t216` (attributed to t263) and `0d26f8a00 chore(catalog): ... t443 jurisdiction (t461)` (attributed to t461) |
| **MUT-CONST-REF** | Ignores the chain and always uses `DefaultLandedRef` | this repository's own state | `git symbolic-ref refs/remotes/origin/HEAD` → `refs/remotes/origin/develop`, while the resolver answers `origin/main` |
| **MUT-CONFIG-ONLY** | Two-level chain (config → constant), level 2 absent | primary checkout config | `worktree_base_branch: ""` at `git-strategy.yaml:7` ⇒ falls straight to `origin/main` |
| **MUT-MERGE-ANY-TOKEN** | Implements §A.4 form 3 as an **occurrence** test — "a merge subject naming the card" — the wording version 0.1.0 carried | **t387/t386** via `9a3837b5c`, and **t284** via `c4ae1ecbd` | `9a3837b5c Merge branch 'WT-audit-evidence-store' into WT-audit-advice-integrity (t387 depends on t386 convention doc)` attributes **both** t386 and t387 under the occurrence reading, and lands neither on the integration branch; `c4ae1ecbd Merge origin/develop into WT-audit-participant-count — absorb upstream before integration (card t284)` is an absorb merge reachable through form 2 as well. Corpus scale at `origin/develop` `7835148d3`: **146** merge subjects admitted vs **5** attributing (`^Merge card`), **21** absorb-direction merges carrying a card in a trailing group. This mutant passes AC-TLA-001..007 as written in version 0.1.0 and moves none of today's 7/9 — it lands silently, which is why AC-TLA-003's third clause exists |
| **MUT-NO-FORM-3B** | Omits form 3b entirely, implementing only forms 1/2/3a/3c | **t412** on the pinned corpus | `b6231290d ... into develop (card t412 — SPEC-MX-TAG-EDGES-001)` — the trailing group carries text after the id, so form 2's `)$` anchor misses it and `t412` is the single id form 3b adds (`comm -13` over the two attributed sets → exactly one line). Cost of the omission is loud (`t412` reads `not-landed`); it passed every criterion at version 0.2.0, where a grep for `t412` in this file returned **0** — the vacuity plan-audit iter-2 D1 named |
| **MUT-GROUP-ANY-TOKEN** | Reads form 3b's trailing group as an occurrence **set** — every card token in the group attributes | a group naming two cards, e.g. `(card t36, absorbs t2)` | Blast radius on form 3b's own subject set is **0 today**: of the 76 `into develop` merges with a card-bearing group, none names two distinct cards. But **8** trailing groups corpus-wide do, so the shape exists and the single-token restriction guards a real population rather than a hypothetical one |
| **MUT-NO-TARGET-TEST** | Implements form 3b's group test but drops the target test | **t412**'s three absorb siblings | `d8c91d907` / `63435427c` / `57d2f3ae3` each carry `t412` in a trailing group while merging **into `WT-mx-tag-edges`**; without the target test all three attribute, and `t412` reads `landed` for the wrong reason |
| **MUT-HARDCODED-DEVELOP** | Spells `develop` as form 3b's target instead of deriving it from the resolved landed ref | a fixture whose resolved ref is not `develop` | Invisible in this repository's M1-only window — measured on `origin/main` `7ad9f8534`: **0** of 101 merge subjects target `develop` with a card-bearing group and **0** target `main`, so the hardcoding and the correct derivation are behaviourally identical there (this refutes the audit's inference that the window distinguishes them). The defect is permanent downstream: a repository integrating on `main` gets nothing from form 3b, and one integrating on a third name gets false attributions. AC-TLA-003 clause 6 is the only falsifier |
| **MUT-NO-FORM-3C** | Omits form 3c (`merge: <card>`), the largest missed family | **t79** | `merge: t79 — glm_task delegation family (branch WT-t80)`. The family is 31 subjects and 12 otherwise-unattributed ids; omitting it returns the under-count from 7 to 19 |
| **MUT-PAREN-END-ANCHOR** | Anchors form 2's card-bearing group at **end of subject** (`)$`), omitting form 2b — the shape every version through 0.3.0 carried | **t210** on the pinned corpus | `feat(kanban): moai todo pr — read-only card-to-PR and landed link (t210) (#1628)`. Measured at HEAD `012d9680a` over `7835148d3`: the shape occurs **43** times carrying **40** distinct ids, of which **39** are attributed by no other form (`comm -23` against the five-form set) — and the same pattern returns **43** on `origin/main`, the ref M1's window walks. The omission is **silent at the operator's eye level**: the card simply reads `not-landed`. It passed every criterion at version 0.3.0, where the §F positive controls were `t401` (form 2, `)$`-exact) and `t440` (form 1) — both shapes the enumeration already handled. This is the vacuity class plan-audit iter-3 D3-1 named |
| **MUT-FIRST-TOKEN-OF-GROUP** | Widens form 2 to accept a card token as the **first** token of a trailing group — the remedy §A.4.2 declines | **t80** | `(branch WT-t80)` — the only card token in the group is a **branch name**, so the widening attributes a card that did not land, on a commit belonging to t79. The failure is silent, unlike the under-count it would fix |
| **MUT-SILENT-FALLBACK** | Resolves correctly but discloses nothing | any invocation below level 1 | verdict line carries no ref, stderr carries no level |

**Positive controls (the fixture set is not vacuous):** **t401** and **t440** are the two measured
true positives on `origin/develop`, attributed by **trailing parenthetical** and **conventional-commit
scope** respectively. Re-measured at `7835148d3` (`git log origin/develop --format='%h %s' | grep -E
'\bt401\b'`): t401's only subject occurrence is `d5caf2d8e feat(SPEC-JUDGMENT-FIRST-MODE-001): …
(t401)` — a trailing parenthetical under a non-card scope; t440 carries `b80cc9cf1 docs(t440): …` — a
conventional-commit scope (plus a form-3a merge and a trailing form). Version 0.1.0 stated the two
forms the other way round and contradicted §B, which had them right (plan-audit D5). A predicate that
fails these has over-corrected.

---

## §B Milestone 1 criteria — the attribution predicate

**AC-TLA-001** (maps REQ-TLA-001, REQ-TLA-002) — conventional-commit scope, both directions.
Given a ref whose history contains `docs(t440): ...` and contains no other occurrence of `t440`,
When the landed predicate is asked about `t440`,
Then the answer is `landed`;
And Given a ref whose only occurrence of `t237` is in a commit body,
When the predicate is asked about `t237`,
Then the answer is `not-landed`.
Red against: **MUT-WHOLE-MESSAGE** (t237 fixture).

**AC-TLA-002** (maps REQ-TLA-001, REQ-TLA-002, REQ-TLA-004) — trailing parenthetical, three directions.
Given a subject closing with `(t401)`,
When the predicate is asked about `t401`,
Then the answer is `landed`;
And Given a subject in which `t443` appears mid-subject while the subject's own trailing attribution
is `(t461)`,
When the predicate is asked about `t443`,
Then the answer is `not-landed`;
And Given a subject closing with a card-bearing group followed by a pull-request reference group —
`feat(kanban): moai todo pr — read-only card-to-PR and landed link (t210) (#1628)`, §A.4 form 2b —
When the predicate is asked about `t210`,
Then the answer is `landed`.
Red against: **MUT-SUBJECT-ONLY** (t443 fixture) and **MUT-PAREN-END-ANCHOR** (t210 fixture).

Clause 3 was added at version 0.4.0 to close plan-audit iter-3 D3-1: before it, an implementation
anchoring the card-bearing group at end-of-subject passed every criterion in this file while
answering `not-landed` for **43 subjects carrying 40 distinct ids, 39 of them attributed by no other
form** — this repository's dominant pull-request landing path, and present in the same count on
`origin/main`, the ref M1's window walks. `t210` is a valid falsifier because it is in form 2b's
uniquely-attributed set (39, per the AC-TLA-005 map): under forms 1/2/3a/3b/3c it reads
`not-landed`, under form 2b it reads `landed`.

**AC-TLA-003** (maps REQ-TLA-001, REQ-TLA-002, REQ-TLA-004, REQ-TLA-013) — merge subject, six
directions. Clauses 4-6 were added at version 0.3.0 to close plan-audit iter-2 D1 and D3: before
them, an implementation that omitted form 3b entirely, and an implementation that spelled `develop`
as form 3b's target, both passed every criterion in this file.
Given a merge subject `Merge branch 'WT-...' into develop (card t263)` — §A.4 form 3b, target is the
branch the landed ref resolves to and the card token lies in the trailing parenthetical group,
When the predicate is asked about `t263`,
Then the answer is `landed`;
And Given the same commit set queried for `t216`, whose only subject occurrence is
`docs(t263): ... behind t216`,
When the predicate is asked about `t216`,
Then the answer is `not-landed`;
And Given the **absorb-direction** merge subjects `9a3837b5c Merge branch 'WT-audit-evidence-store'
into WT-audit-advice-integrity (t387 depends on t386 convention doc)` and `c4ae1ecbd Merge
origin/develop into WT-audit-participant-count — absorb upstream before integration (card t284)` —
each naming a card while merging **into** a `WT-` branch,
When the predicate is asked about `t386`, `t387`, and `t284`,
Then every answer is `not-landed`, because §A.4's non-attribution rule excludes a merge whose named
target is not the branch the resolved landed ref names — the merge's subject is a branch, not a card;
And **(clause 4 — form 3b's green direction, the criterion form 3b previously lacked)** Given the
merge subject `b6231290d Merge branch 'WT-mx-tag-edges' into develop (card t412 —
SPEC-MX-TAG-EDGES-001)`, whose target IS the branch the resolved landed ref names and whose trailing
group carries text after the card id — so form 2's `)$`-exact anchor does **not** match it,
When the predicate is asked about `t412`,
Then the answer is `landed`;
And **(clause 5 — form 3b's red direction, exercising the target test and the single-token
restriction together)** Given the three sibling subjects `d8c91d907 Merge branch 'origin/develop'
into WT-mx-tag-edges (window absorption, card t412)`, `63435427c Merge branch 'origin/develop' into
WT-mx-tag-edges (pre-run absorption, card t412)`, and `57d2f3ae3 Merge branch 'WT-edge-confidence'
into WT-mx-tag-edges (card t412 dependency absorption)` — each naming `t412` inside a trailing group
while merging into a branch the landed ref does not name,
When those three commits are the only history and the predicate is asked about `t412`,
Then the answer is `not-landed`;
And **(clause 6 — REQ-TLA-013, the target is derived, not spelled)** Given a fixture repository whose
resolved landed ref names a branch other than `develop` — for example `origin/release/v9`, with a
merge subject `Merge branch 'WT-x' into release/v9 (card tNNN — note)` and a second subject
`Merge branch 'WT-y' into develop (card tMMM — note)`,
When the predicate is asked about `tNNN` and then about `tMMM`,
Then `tNNN` reads `landed` and `tMMM` reads `not-landed` — the attributing target moved with the
resolved ref rather than staying on the literal string `develop`.
Red against: **MUT-SUBJECT-ONLY** (t216 fixture), **MUT-MERGE-ANY-TOKEN** (clause 3), **MUT-NO-FORM-3B**
(clause 4 is the only criterion that falsifies it — `t412` is the single id form 3b adds over forms
1/2/3a, and it appeared nowhere in this file before version 0.3.0), **MUT-GROUP-ANY-TOKEN** and
**MUT-NO-TARGET-TEST** (clause 5), and **MUT-HARDCODED-DEVELOP** (clause 6 is the only criterion that
falsifies it; clause 4's fixture is `into develop` and so passes under a spelled target too).

**AC-TLA-003b** (maps REQ-TLA-001, REQ-TLA-002, REQ-TLA-004) — the card-led merge shapes (forms 3a
and 3c), each with a falsifier of its own.

*Form 3a.* Given the merge subject `Merge card t244 (WT-team-ac-verify-wiring) into develop:
keep-dormant verdict` — the **only** subject in the pinned corpus that form 3a alone attributes
(`comm -13` of the other four forms' union against form 3a's set → exactly `t244`),
When the predicate is asked about `t244`,
Then the answer is `landed`;
And Given the same card's other subject `docs(hooks): t244 verdict — keep team-ac-verify dormant`,
whose scope is a package rather than a card and whose `t244` sits mid-sentence,
When that subject is the only history,
Then the answer is `not-landed`.
Red against: **MUT-NO-FORM-3A** — an implementation dropping form 3a passes every other criterion,
because forms 1 and 2 cover the four remaining `Merge card` ids (`t242`, `t243`, `t440`, `t451`).

*Form 3c.* Given the subject `merge: t79 — glm_task delegation family (branch WT-t80)`, whose first
token after the merge verb is `t79`,
When the predicate is asked about `t79`,
Then the answer is `landed`;
And Given the same subject queried for `t80`, whose only occurrence is inside the **branch name**
`WT-t80` sitting in the trailing group,
When the predicate is asked about `t80`,
Then the answer is `not-landed` — a branch name is a place, not an attribution (§A.4 preamble).
This pair is deliberately one fixture: it establishes form 3c's green direction, its red direction,
and the branch-name non-attributing position at once. Neither id is attributed by any other form in
the pinned corpus, so both directions are live.
Red against: **MUT-NO-FORM-3C** (first clause) and **MUT-FIRST-TOKEN-OF-GROUP** — the widening
declined in §A.4.2, which reads `(branch WT-t80)` as attributing `t80` (second clause).

**AC-TLA-004** (maps REQ-TLA-003) — body mentions never attribute.
Given a commit whose subject names no card token and whose body names `tNNN`,
When the predicate is asked about `tNNN`,
Then the answer is `not-landed`;
And Given the same fixture with the token additionally present as a conventional-commit scope on a
second commit,
Then the answer is `landed` — the body occurrence neither adds to nor subtracts from the verdict.
Red against: **MUT-WHOLE-MESSAGE**.

**AC-TLA-005** (maps REQ-TLA-002) — the form enumeration is one place, and every entry in it has a
falsifier.
Given the repaired implementation,
When the §A.4 attributing shapes (forms 1, 2, 2b, 3a, 3b, 3c) **and the non-attribution rule** are
located by grep,
Then they are declared in exactly one named symbol or table;
And When any ONE of them is removed from that declaration,
Then a named criterion fails — per the map below, which is the claim version 0.2.0 made without
holding (plan-audit iter-2 D1: form 3b had no falsifier, and neither did form 3a):

| Removed | Criterion that goes red | Falsifying id | Ids that form alone attributes |
|---|---|---|---|
| form 1 (scope) | AC-TLA-001 clause 1 | t440 | 41 |
| form 2 (trailing paren) | AC-TLA-002 clause 1 | t401 | 77 |
| form 2b (trailing paren before a reference group) | AC-TLA-002 clause 3 | t210 | 39 |
| form 3a (`Merge card`) | AC-TLA-003b clause 1 | t244 | 1 |
| form 3b (integration-targeted merge) | AC-TLA-003 clause 4 | t412 | 1 |
| form 3c (`merge:`) | AC-TLA-003b clause 3 | t79 | 12 |
| non-attribution rule | AC-TLA-003 clauses 3 and 5 | t386/t387/t284, and t412's three absorb siblings | n/a (an exclusion) |
| derived target (REQ-TLA-013) | AC-TLA-003 clause 6 | the `release/v9` fixture | n/a (a derivation) |

The "ids that form alone attributes" column is the measured basis of each falsifier and is the check
a future editor re-runs before adding a further form: a form whose column reads **0** has no
falsifier available from the corpus, and adding it would reintroduce exactly the vacuity this table
exists to close.

The column is **relative to the rest of the table**, so admitting a form recomputes it: at version
0.4.0 form 2b took `t230` out of form 1's exclusive set, moving form 1 from 42 to 41. Both figures
are measured at HEAD `012d9680a` over the pinned corpus `7835148d3` by intersecting each form's
extracted id set against the union of the others (`comm -23`).

**AC-TLA-006** (maps REQ-TLA-005) — argv construction is asserted against the implementation.
Given the tripwire test,
When it asserts the query the landed check runs,
Then it obtains that query by calling the exported builder, not by re-constructing it;
And When the builder's engine flag or shape is mutated,
Then the tripwire fails.

**AC-TLA-007** (maps REQ-TLA-006) — the unanswerable path stays three-valued.
Given a querier whose runner returns an error,
When the predicate is evaluated,
Then the answer is `unknown` and not `not-landed`;
And Given a runner returning an empty commit set,
Then the answer is `not-landed` and not `unknown` — the two facts stay distinguishable.

---

## §C Milestone 2 criteria — the ref chain and its disclosure

**AC-TLA-008** (maps REQ-TLA-007, REQ-TLA-008) — the chain resolves in order.
Given a project root whose `git_strategy.worktree_base_branch` is `develop`,
When the landed ref is resolved,
Then it is `origin/develop`, and `refs/remotes/origin/HEAD` is not consulted;
And Given a project root whose key is empty while `refs/remotes/origin/HEAD` names `develop`,
Then the resolved ref is `origin/develop`.
Red against: **MUT-CONFIG-ONLY** (second direction fails: it yields `origin/main`).

**AC-TLA-009** (maps REQ-TLA-007, REQ-TLA-009, REQ-TLA-011) — level 3 is reached only on level-2
failure, and the answering level is disclosed.
Given a project root with an empty key, in a repository fixture **constructed so that
`git symbolic-ref refs/remotes/origin/HEAD` actually exits non-zero** <!-- moving-ref-ok: SUBJECT/S1 per VCI §2.1 — the ref is quoted subject matter inside a constructed fixture (Instance 4 shape), not an address a measurement was taken at. Test 1: substituting a SHA destroys the criterion, which is about the symref being ABSENT. Test 4: no read-time measurement — the fixture is built by omitting the symref, so a pin would name a value the criterion requires not to exist. Remedy R3. --> — the symref is absent from the
fixture's git metadata, not stubbed out in a code branch,
When the landed ref is resolved,
Then it is `DefaultLandedRef`, the invocation does not fail, and the disclosure REQ-TLA-011 requires
names **level 3** as the answering level;
And Given the same root with `refs/remotes/origin/HEAD` resolvable,
Then the disclosure names **level 2** as the answering level.
The second clause asserts **provenance — which level answered — not a value inequality.** Version
0.1.0 wrote it as "the resolved ref is not `DefaultLandedRef`", which false-fails on any repository
whose `origin/HEAD` legitimately names `main`: a correct level-2 resolution there yields
`origin/main`, the literal value of `DefaultLandedRef` (`internal/kanban/prlink_landed.go:41`), and
the criterion would reject a correct answer (plan-audit D4).
The first clause's fixture requirement is **machine state, not a code branch**: asserting the
level-3 branch without constructing a repository where the symref read genuinely fails passes green
even when the branch is unreachable. The fixture is built by omitting `refs/remotes/origin/HEAD`, and
the criterion is red until the chain actually falls through.
Red against: **MUT-CONST-REF** and **MUT-SILENT-FALLBACK** (which answers correctly but discloses no
level).

**AC-TLA-010** (maps REQ-TLA-010) — the verdict names its ref, prefix preserved.
Given `moai todo done <id> --require-landed` succeeding,
When stdout is read,
Then the line still begins `done <id> ` and additionally names the ref that answered;
And When an existing reader keyed on the `done <id> ` prefix parses the line,
Then it still matches.

**AC-TLA-011** (maps REQ-TLA-011) — the sub-config level is disclosed.
Given a resolution that reached level 2 or level 3,
When a landing verdict is emitted,
Then stderr names which chain level supplied the ref;
And Given a resolution that answered at level 1,
Then no such disclosure is emitted — the notice marks the exceptional path, not every path.
Red against: **MUT-SILENT-FALLBACK**.

**AC-TLA-012** (maps REQ-TLA-012) — resolution writes nothing.
Given a subprocess census over the ref-resolution path,
When the landed ref is resolved,
Then no invocation is a ref-writing git command (`symbolic-ref` with an operand, `remote set-head`,
`update-ref`);
And When such a command is planted in the path,
Then the census fails — proving the census observes the path rather than passing vacuously.

---

## §D Edge cases

- A card token appearing as a substring of a longer token (`t44` inside `t443`) must not attribute.
  Word-boundary matching is retained from the current implementation.
- **Two attributing forms naming different cards — the ruling: trailing parenthetical first, scope
  second.** A subject carrying both a conventional-commit scope and a trailing parenthetical
  (`docs(t263): … (t461)`) attributes to the **trailing parenthetical**; the scope is consulted only
  when no trailing form is present. This is a **deterministic tiebreak stated here so the
  implementation does not decide it silently** (plan-audit D3 — the mandate this bullet previously
  made of itself and did not meet).

  *Observed instances of the ambiguous case: **0**.* Measured at `origin/develop` `7835148d3`, 5,837
  subjects:

  | | count |
  |---|---|
  | scope form `^[a-z]+\(tNNN\)!?:` | 290 |
  | trailing form `\((card )?tNNN\)$` | 869 |
  | merge form `^Merge card tNNN` | 5 |
  | carrying **both** scope and trailing | 34 |
  | └ same id in both positions | 34 |
  | └ **different id (the ambiguous case)** | **0** |

  Non-vacuous: every operand, the intersection included, is non-empty. Control — the extractor
  distinguishes the two cases rather than collapsing them: `docs(t1): x (t1)` → `t1 t1`;
  `docs(t1): x (t2)` → `t1 t2`. So the 0 is a measured 0, not an artefact of a predicate that cannot
  see a difference.

  Three grounds for choosing the trailing form: it is **3× more common** (869 vs 290); the scope is
  **demonstrably not always a card id** (`chore(catalog):` and `chore(reports):` account for 4
  subjects at this tree, and a scope naming a package or SPEC is the norm — `d5caf2d8e
  feat(SPEC-JUDGMENT-FIRST-MODE-001): … (t401)` is the positive control itself); and `AGENTS.md` §3's
  card-id-in-every-commit obligation is discharged **at the trailing position** in practice.

  *Revisit trigger:* the first observed instance (N ≥ 1) of a different-id dual-form subject. Until
  then this is a ruling on a zero-instance case, deliberately settled in text rather than escalated
  — the lead declined to rule on a case with no instances, and an unstated tiebreak is what the
  §D mandate exists to prevent.

- **Two cards inside ONE parenthetical group — a different axis from the tiebreak above, and it is
  not zero.** The tiebreak's measured population is *scope-id ∧ trailing-id*, two positions on one
  subject. The lane's iteration-2 verification named a second axis the table above does not cover:
  **two tokens inside a single trailing group** (`(card t36, absorbs t2)`, `(t46/t73/t74)`,
  `(t333/t347)`, `(t393/t383/t421 integration)`, `(t387 depends on t386 convention doc)`,
  `(absorb t280, … includes t239)`, `(card t333/t347)`, `(t46 + t73)`). Measured over the pinned
  corpus: **8 subjects**, against **0** for the tiebreak's own axis. The two are counted separately
  because they are separate populations, and the §D tiebreak does **not** rule on this one.

  The ruling on this axis lives in the enumeration rather than here, and it is a rejection rather
  than a preference: form 2 requires the group to carry **nothing else**, and form 3b requires
  **exactly one** card token (§A.4). All eight subjects above therefore attribute **nothing** via
  their trailing group. Checked id by id against the attributed set, the cost of that rejection is
  bounded and already accounted for: `t196`, `t333`, `t347`, `t393`, `t383`, `t421`, `t280`, `t239`,
  `t386`, `t387`, and `t36` are each attributed by some other subject in the corpus, so rejecting
  their groups costs nothing; `t2` is correctly unattributed (its only appearance is the phrase
  `absorbs t2`, which is a note, not a landing); and `t46`, `t68`, `t73`, `t74` sit in the residual
  under-count of §A.4.2, which is where their cost is recorded rather than hidden.

  Stating it here matters because the alternative reading is available and wrong: a group carrying
  two cards is not an ambiguity to be tiebroken (which card does it attribute?) but a shape that
  attributes neither. Tiebreaking it would mean picking one of two cards on a commit that delivered
  neither — a silent false positive — whereas rejecting it under-counts, loudly.
- An empty ref string handed to the builder still falls back to `DefaultLandedRef` rather than
  emitting an empty argv element that git would read as the working tree.

---

## §E Quality gates

- `go test ./internal/kanban/... ./internal/cli/...` green in this tree.
- `go vet` and `golangci-lint run` clean on the touched packages.
- Full-suite verdict from CI on the pushed head, not from a local full-suite run.
- Every criterion above observed **red** against its named mutant before the repair, and the mutant
  reverted — recorded in `progress.md` §E.2 with the command and its verbatim output.

---

## §F Definition of Done

- M1 landed in its own commit, before any M2 commit (`plan.md` §A.3, [HARD]).
- All **thirteen** REQ-TLA requirements have at least one criterion above (REQ-TLA-013 is covered by
  AC-TLA-003 clause 6), and every axis-F criterion asserts both directions.
- Every entry in the §A.4 enumeration has a named falsifier per the AC-TLA-005 map — checked by
  removing the entry and observing the named criterion go red, not by reading the map.
- `moai todo pr` re-run in this tree after M1: the two `origin/main` false positives (t237, t312) no
  longer read `landed`, and the two true positives (t401, t440) still do against `origin/develop`.
- **A form 2b positive control is included, and is not optional (plan-audit iter-3 D3-1).** `t210`
  reads `landed`. Version 0.3.0's controls were `t401` (form 2, `)$`-exact) and `t440` (form 1) —
  both shapes the enumeration already handled — so this gate was satisfiable while 39 ids landing
  through this repository's dominant pull-request path read `not-landed`.
- No file under `.moai/reports/t472/` modified.
- **Non-gating (plan-audit D7):** `plan.md` M3 item 2 (extending the `todo pr` outcome documentation
  to state the repaired predicate's limit) maps to no REQ-TLA and to no criterion above, and is
  **explicitly not a Definition-of-Done gate**. It is documentation work carried on the milestone for
  sequencing only; its absence does not block close. M3 item 1 (the tripwire) IS gated, via
  AC-TLA-006.

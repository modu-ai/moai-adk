# Acceptance Criteria — SPEC-TODO-LANDING-ATTRIBUTION-001

Card t472. Measured tree `4bcac7079`; plan-audit iteration-1 remediation re-measured at HEAD
`e227871b4` against `origin/develop` `7835148d3`.

Every axis-F **predicate-behaviour** criterion (AC-TLA-001..004, -007) is **bidirectional** by lead
directive: it asserts the green direction (a title-attributed commit reads `landed`) AND the red
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

**AC-TLA-002** (maps REQ-TLA-001, REQ-TLA-002, REQ-TLA-004) — trailing parenthetical, both directions.
Given a subject closing with `(t401)`,
When the predicate is asked about `t401`,
Then the answer is `landed`;
And Given a subject in which `t443` appears mid-subject while the subject's own trailing attribution
is `(t461)`,
When the predicate is asked about `t443`,
Then the answer is `not-landed`.
Red against: **MUT-SUBJECT-ONLY** (t443 fixture).

**AC-TLA-003** (maps REQ-TLA-001, REQ-TLA-002, REQ-TLA-004) — merge subject, three directions.
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
target is a card worktree branch — the merge's subject is a branch, not a card.
Red against: **MUT-SUBJECT-ONLY** (t216 fixture) and **MUT-MERGE-ANY-TOKEN** (the third clause is the
only criterion that falsifies it; an occurrence reading of form 3 answers `landed` for all three).

**AC-TLA-004** (maps REQ-TLA-003) — body mentions never attribute.
Given a commit whose subject names no card token and whose body names `tNNN`,
When the predicate is asked about `tNNN`,
Then the answer is `not-landed`;
And Given the same fixture with the token additionally present as a conventional-commit scope on a
second commit,
Then the answer is `landed` — the body occurrence neither adds to nor subtracts from the verdict.
Red against: **MUT-WHOLE-MESSAGE**.

**AC-TLA-005** (maps REQ-TLA-002) — the form enumeration is one place.
Given the repaired implementation,
When the §A.4 attributing shapes (forms 1, 2, 3a, 3b) **and its non-attribution rule** are located by
grep,
Then they are declared in exactly one named symbol or table;
And When one shape — or the non-attribution rule — is removed from that declaration,
Then the criterion for it (AC-TLA-001, -002, or -003's matching clause) fails — proving the
declaration is the live source and not a duplicated comment.

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
- All twelve REQ-TLA requirements have at least one criterion above, and every axis-F criterion
  asserts both directions.
- `moai todo pr` re-run in this tree after M1: the two `origin/main` false positives (t237, t312) no
  longer read `landed`, and the two true positives (t401, t440) still do against `origin/develop`.
- No file under `.moai/reports/t472/` modified.
- **Non-gating (plan-audit D7):** `plan.md` M3 item 2 (extending the `todo pr` outcome documentation
  to state the repaired predicate's limit) maps to no REQ-TLA and to no criterion above, and is
  **explicitly not a Definition-of-Done gate**. It is documentation work carried on the milestone for
  sequencing only; its absence does not block close. M3 item 1 (the tripwire) IS gated, via
  AC-TLA-006.

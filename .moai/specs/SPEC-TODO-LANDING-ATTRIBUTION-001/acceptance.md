# Acceptance Criteria — SPEC-TODO-LANDING-ATTRIBUTION-001

Card t472. Measured tree `4bcac7079`.

Every axis-F criterion is **bidirectional** by lead directive: it asserts the green direction (a
title-attributed commit reads `landed`) AND the red direction (a body-mention-only or
other-card-attributed commit reads `not-landed`). A criterion carrying only the green direction is
satisfied by an implementation that answers `landed` for everything, and is not accepted.

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
| **MUT-SILENT-FALLBACK** | Resolves correctly but discloses nothing | any invocation below level 1 | verdict line carries no ref, stderr carries no level |

**Positive controls (the fixture set is not vacuous):** **t401** and **t440** are the two measured
true positives on `origin/develop`, attributed by conventional-commit scope and trailing
parenthetical respectively. A predicate that fails these has over-corrected.

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

**AC-TLA-003** (maps REQ-TLA-001, REQ-TLA-002, REQ-TLA-004) — merge subject, both directions.
Given a merge subject `Merge branch 'WT-...' into develop (card t263)`,
When the predicate is asked about `t263`,
Then the answer is `landed`;
And Given the same commit set queried for `t216`, whose only subject occurrence is
`docs(t263): ... behind t216`,
When the predicate is asked about `t216`,
Then the answer is `not-landed`.
Red against: **MUT-SUBJECT-ONLY** (t216 fixture).

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
When the three attributing forms are located by grep,
Then they are declared in exactly one named symbol or table;
And When one form is removed from that declaration,
Then the criterion for that form (AC-TLA-001, -002 or -003) fails — proving the declaration is the
live source and not a duplicated comment.

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

**AC-TLA-009** (maps REQ-TLA-007, REQ-TLA-009) — level 3 is reached only on level-2 failure.
Given a project root with an empty key and no resolvable `refs/remotes/origin/HEAD`,
When the landed ref is resolved,
Then it is `DefaultLandedRef` and the invocation does not fail;
And Given the same root with `refs/remotes/origin/HEAD` resolvable,
Then the resolved ref is not `DefaultLandedRef`.
Red against: **MUT-CONST-REF**.

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
- A subject carrying two attributing forms for **different** cards (scope `docs(t263):` plus trailing
  `(t461)`) — the criterion set must state which wins, or accept both as attributions; either ruling
  is acceptable provided it is stated and tested. It is not left to the implementation to decide
  silently.
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

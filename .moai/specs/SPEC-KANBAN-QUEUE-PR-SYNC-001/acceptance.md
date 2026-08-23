# Acceptance Criteria — SPEC-KANBAN-QUEUE-PR-SYNC-001

13 acceptance criteria (Tier M ceiling: 16).

Every criterion is verified by a command that observes something. Per this
project's standing rule, a grep-based criterion is admissible only when it
returns **zero** hits on the pre-implementation tree; the one grep-shaped
criterion here (AC-013) records its required baseline explicitly.

**Fixture provenance.** The PR fixtures below were re-measured live in this
worktree with:

```
$ gh pr list --state open --limit 40 --json number,title,body \
    -q '.[] | "\(.number)|TITLE:\([.title|scan("\\bt[0-9]{1,4}\\b")]|join(" "))|BODY:\([.body//""|scan("\\bt[0-9]{1,4}\\b")]|unique|join(" "))"'
1614|TITLE:t203|BODY:t1 t151 t203 t69 t9
1612|TITLE:t200|BODY:t200 t201
1611|TITLE:|BODY:t201
1601|TITLE:|BODY:t188
1600|TITLE:|BODY:t184
```

The v0.1.1 fixtures for AC-002 and AC-003 were refuted by this data (audit D4)
and are rebuilt below. Per `plan.md` §B the fixture set is **pinned from these
values**, not re-fetched at test time — the live PR set changes.

## D. AC Matrix

| AC | Requirement(s) | Kind | Severity |
|---|---|---|---|
| AC-001 | REQ-1.2 | behavioural | must |
| AC-002 | REQ-1.3 | behavioural | must |
| AC-003 | REQ-1.4, REQ-1.6 | behavioural | must |
| AC-004 | REQ-2.1, REQ-2.2 | behavioural (directory digest) | **must — load-bearing** |
| AC-005 | REQ-2.3 | behavioural | must |
| AC-006 | REQ-1.5 | behavioural | must |
| AC-007 | REQ-1.7 | behavioural | must |
| AC-008 | REQ-1.8 | behavioural | must |
| AC-009 | REQ-2.4, REQ-2.5 | behavioural | must |
| AC-010 | REQ-2.6 | behavioural | must |
| AC-011 | REQ-1.9 | behavioural (with controls) | **must — anti-vacuous** |
| AC-012 | REQ-1.10 | behavioural | must |
| AC-013 | mirror parity (plan.md M4) | mechanical grep + reviewer judgement | must |

---

### AC-001 — an id in the PR title resolves `exact`

**Given** the pinned fixture containing PR #1612 with title token `t200`
**When** the resolver is run for card `t200`
**Then** it returns a `linked` outcome, PR number 1612, confidence `exact`.

Verify: `go test ./internal/kanban/... -run TestResolve_ExactFromTitle -v`

### AC-002 — no title token, exactly one body token resolves `inferred`

**Given** the pinned fixture, in which `t188` appears in no PR title and in
exactly one PR body (#1601)
**When** the resolver is run for card `t188`
**Then** it returns a `linked` outcome, PR number 1601, confidence `inferred`.

Verify: `go test ./internal/kanban/... -run TestResolve_InferredFromSingleBody -v`

> Rebuilt per audit D4. The prior fixture asserted `inferred` for `t201`, whose
> token appears in **two** bodies (#1611 and #1612) — that is the ambiguous case,
> and it now anchors AC-003 instead. `t188` is genuinely single-bodied and
> title-absent.

### AC-003 — several body tokens resolve `ambiguous`, never a best guess

**Given** the pinned fixture, in which `t201` appears in no PR title and in
**two** PR bodies — #1611 and #1612
**When** the resolver is run for card `t201`
**Then** the outcome kind is `ambiguous`,
**And** the candidate list contains exactly {1611, 1612},
**And** no single PR is returned as a resolved link.

Verify: `go test ./internal/kanban/... -run TestResolve_AmbiguousNotCollapsed -v`

> Rebuilt per audit D4. The prior fixture used `t151`, which appears in exactly
> one body (#1614) and therefore resolves `inferred` — it could never produce
> `ambiguous`, leaving REQ-1.4 and REQ-1.6 untested. `t201` is the genuine
> ambiguous case in the measured set.

### AC-004 — nothing under the queue directory changes across an invocation

**This is the load-bearing criterion for the §B read-only ruling.**

**Given** a `.moai/state/kanban/` fixture directory with a recorded recursive
digest — every file's path plus its SHA-256, plus the set of paths present
**When** `moai todo pr` and `moai todo pr --json` are each invoked against it
**Then** the recursive digest is unchanged after every invocation,
**And** no path has been added or removed (no sidecar, no lock file, no cache),
**And** `backlog.json`'s own SHA-256 and modification time are unchanged,
**And** all of the above hold equally on the fail-open path (AC-005), the
ambiguous path (AC-003), and the landed path (AC-011).

Verify: `go test ./internal/cli/... -run TestTodoPR_QueueDirUnchanged -v`

> Widened per audit D7. A single-file byte-identity assertion is necessary but
> not sufficient for REQ-2.2, which forbids caching into *any* queue-owned file:
> a `findings` sidecar, a lock taken and released, or an mtime touch on a
> neighbouring file would all pass a `backlog.json`-only check. Declaring one
> criterion load-bearing raises the bar for it rather than lowering it for its
> neighbours.

### AC-005 — fail-open when `gh` is unavailable

**Given** a `PATH` from which `gh` is absent
**When** `moai todo pr` is invoked
**Then** the exit code is 0,
**And** the link column renders **empty** for every card (not omitted — the
column is present and blank),
**And** stderr carries a degradation note rather than an error,
**And** the landed check still runs, so a landed card still reports `landed`,
**And** AC-004's digest assertion still holds.

Verify: `go test ./internal/cli/... -run TestTodoPR_FailOpenNoGh -v`

Repeat for `gh` present but exiting non-zero (unauthenticated / offline
simulation): `-run TestTodoPR_FailOpenGhNonZero`.

> Disjunction removed per audit D11. "Blank **or** absent" could not fail on
> rendering, so the assertion observed only that the process did not crash —
> which the exit-code clause already covers. The column is blank.

### AC-006 — commit tokens are not an attribution carrier

**Given** the pinned fixture for PR #1600, whose body carries `t184` and whose
commit messages carry many other card tokens but **not** `t184`
**When** the resolver is run for a card that appears in #1600's commit tokens
but in neither its title nor its body
**Then** the resolver returns no `linked` outcome pointing at #1600 for that
card.

Verify: `go test ./internal/kanban/... -run TestResolve_IgnoresCommitTokensForAttribution -v`

> Per audit D13, the count is dropped. The prior text pinned "15 commit-message
> tokens", which no stated command reproduces — a headline-only scan of #1600's
> commits yields 12, and the record's 15 evidently came from full commit
> messages. The load-bearing property is not the count: it is that #1600's
> delivering card is present in the body and absent from the commit tokens.

### AC-007 — all four outcome kinds are mutually distinguishable

**Given** a fixture producing one of each outcome kind — `linked`, `ambiguous`,
`landed`, `no-link`
**When** each is returned to a consumer
**Then** the consumer can distinguish all four by kind alone,
**And** in particular `landed` is distinguishable from `no-link` (an
already-merged card is never reported as untouched work),
**And** `no-link` is distinguishable from `ambiguous` by kind, not by an empty
candidate list on an ambiguous record.

Verify: `go test ./internal/kanban/... -run TestResolve_FourOutcomeKindsDistinct -v`

### AC-008 — whole-token matching

**Given** a fixture PR whose title carries `t200`
**When** the resolver is run for card `t20`
**Then** no link is returned for `t20`.

Verify: `go test ./internal/kanban/... -run TestResolve_WholeTokenOnly -v`

### AC-009 — `moai todo list` stays network-free and git-free

**Given** an injected command executor that records every subprocess invocation
**When** `moai todo list` is invoked with no flags
**Then** it exits 0 and renders the queue,
**And** the recorded invocation count is **zero** — no `gh` process and no `git`
process is spawned.

Verify: `go test ./internal/cli/... -run TestTodoList_NoSubprocess -v`

> This is a whole claim, not a no-flag-only one, because REQ-2.5 puts no `--pr`
> flag in scope (audit D5). If a follow-up ever adds the flag (§H), this
> criterion needs a companion asserting the flag is what gates the spawn.

### AC-010 — outcome and confidence are rendered, and the JSON form carries them

**Given** a fixture producing one `exact`, one `inferred`, one `ambiguous`, and
one `landed` outcome
**When** `moai todo pr --json` is invoked
**Then** each record carries `card_id` and `outcome`, carries `pr` and
`pr_state` for the `linked` and `ambiguous` kinds, and carries `confidence` for
the `linked` kind,
**And** the human render displays the outcome kind for every card and the
confidence label for every `linked` card.

Verify: `go test ./internal/cli/... -run TestTodoPR_RendersOutcomeAndConfidence -v`

### AC-011 — the landed check works, and cannot pass vacuously

**This criterion carries controls in both directions because its failure mode is
a silent empty result (audit D17).**

**Given** a repository fixture whose `origin/main` history contains commits
naming card `t199` and no commit naming card `t205`
**When** the landed check runs
**Then** — **positive control** — `t199` returns `landed`, and the underlying
query returns a non-empty commit set,
**And** — **negative control** — `t205` returns `no-link` with an empty commit
set,
**And** — **regex-engine tripwire** — a test asserts that the implementation's
query uses `--perl-regexp`; running the same query with `-E` returns empty for
`t199`, and the test fails if the implementation's own result matches the `-E`
result on the positive control.

Verify: `go test ./internal/kanban/... -run TestLandedCheck_Controls -v`

Baseline observed in this worktree at authoring time:

```
$ git log origin/main --perl-regexp --grep='\bt199\b' --oneline | wc -l
       3
$ git log origin/main -E --grep='\bt199\b' --oneline | wc -l
       0
$ git log origin/main --perl-regexp --grep='\bt205\b' --oneline | wc -l
       0
```

Without the positive control, an `-E` regression makes every card report
`no-link` and the criterion passes on a result that observed nothing.

### AC-012 — the landed answer is a boolean, and names no delivering commit

**Given** a repository fixture where a card's first matching commit is another
card's report commit that merely mentions it
**When** the landed check returns `landed` for that card
**Then** the returned record contains no commit SHA, no commit subject, and no
field naming a delivering commit,
**And** the public resolver API exposes no accessor that would return one.

Verify: `go test ./internal/kanban/... -run TestLandedCheck_BooleanOnly -v`

### AC-013 — template mirror parity for the `todo.md` command table

**Mechanical part.** The `moai todo pr` verb is documented in the command table
of `.claude/skills/moai/workflows/todo.md` and mirrored into
`internal/template/templates/.claude/skills/moai/workflows/todo.md`.

**Pre-implementation baseline (required):** the following returns **zero** on
both files before M4 lands. Confirm the zero baseline before accepting this
criterion.

```
grep -c 'moai todo pr' .claude/skills/moai/workflows/todo.md
grep -c 'moai todo pr' internal/template/templates/.claude/skills/moai/workflows/todo.md
```

**Then** after M4 both return at least 1,
**And** `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -v` passes.

**Reviewer-judgement part (recorded as such, not claimed as mechanical).** A
reviewer confirms the mirrored table entry describes the verb's behaviour
accurately and carries no internal-only content. No command verifies this, and
the SPEC does not claim it does.

> Split per audit D12, and extended to the `todo.md` mirror per D15. The
> `kanban-dispatch.md` mirror criterion moved with the doctrine changes to
> `SPEC-KANBAN-PR-CARD-TRACEABILITY-001`.

---

## D.1 Severity

All 13 criteria are `must`. There are no `should`-severity criteria.

Two are called out above as carrying extra weight, for different reasons:
AC-004 is **load-bearing** (it is what enforces the §B read-only ruling), and
AC-011 is **anti-vacuous** (its failure mode is a silent empty result that would
otherwise pass).

## D.2 Traceability

| REQ | AC |
|---|---|
| REQ-1.1 | AC-001, AC-002, AC-003, AC-007, AC-010 |
| REQ-1.2 | AC-001 |
| REQ-1.3 | AC-002 |
| REQ-1.4 | AC-003 |
| REQ-1.5 | AC-006 |
| REQ-1.6 | AC-003 |
| REQ-1.7 | AC-007 |
| REQ-1.8 | AC-008 |
| REQ-1.9 | AC-011 |
| REQ-1.10 | AC-012 |
| REQ-2.1 | AC-004 |
| REQ-2.2 | AC-004 |
| REQ-2.3 | AC-005 |
| REQ-2.4 | AC-009 |
| REQ-2.5 | AC-009 |
| REQ-2.6 | AC-010 |

Every one of the 16 leaf requirements has at least one criterion, and every
criterion exercises at least one requirement. The two mappings the audit found
non-functional (REQ-1.3 via AC-002, REQ-1.4 / REQ-1.6 via AC-003) are repaired
by the D4 fixture rebuild above.

## D.3 Quality gates

- `go vet ./...` clean.
- `golangci-lint run` clean on touched packages.
- Affected-package tests green locally; the full-suite verdict comes from CI on
  the PR head.
- Coverage on the new resolver package ≥ 85%.
- `moai spec lint` on this SPEC returns an empty findings array.

## D.4 Definition of Done

- All 13 criteria pass with cited command output.
- AC-004's recursive-digest assertion is present in the committed test file, not
  performed by hand.
- AC-011's positive control, negative control, and `-E` tripwire are all three
  present; a landed-check test suite without the positive control is rejected.
- AC-013 recorded its zero-baseline grep result on both files before M4 landed,
  and its reviewer-judgement half is recorded as a judgement rather than
  reported as a mechanical pass.
- `make build` run after the template mirror.

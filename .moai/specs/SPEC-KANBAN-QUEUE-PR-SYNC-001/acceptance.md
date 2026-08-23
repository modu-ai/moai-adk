# Acceptance Criteria — SPEC-KANBAN-QUEUE-PR-SYNC-001

Every AC below is verified by a command that observes something. Per this
project's standing rule, a grep-based AC is admissible only when it returns
**zero** hits on the pre-implementation tree; each grep AC below records its
required pre-implementation baseline explicitly.

## D. AC Matrix

| AC | Requirement | Kind | Severity |
|---|---|---|---|
| AC-001 | REQ-1.2 | behavioural | must |
| AC-002 | REQ-1.3 | behavioural | must |
| AC-003 | REQ-1.4, REQ-1.6 | behavioural | must |
| AC-004 | REQ-2.1 | behavioural (byte-identity) | **must — load-bearing** |
| AC-005 | REQ-2.3 | behavioural | must |
| AC-006 | REQ-1.5 | behavioural | must |
| AC-007 | REQ-1.7 | behavioural | must |
| AC-008 | REQ-1.8 | behavioural | must |
| AC-009 | REQ-2.5, REQ-2.4 | behavioural | must |
| AC-010 | REQ-2.6, REQ-2.7 | behavioural | must |
| AC-011 | REQ-3.1 | grep (zero-baseline) | must |
| AC-012 | REQ-3.2, REQ-3.3 | grep (zero-baseline) | must |
| AC-013 | REQ-3.4 | mirror parity | must |

---

### AC-001 — an id in the PR title resolves `exact`

**Given** a fixture PR set containing PR #1617 with title token `t205`
**When** the resolver is run for card `t205`
**Then** it returns exactly one link, PR number 1617, confidence `exact`.

Verify: `go test ./internal/kanban/... -run TestResolveLink_ExactFromTitle -v`

### AC-002 — no title token, exactly one body token resolves `inferred`

**Given** the fixture set containing PR #1611 (no title token, body token `t201`)
and no other open PR whose body carries `t201`
**When** the resolver is run for card `t201`
**Then** it returns one link, PR number 1611, confidence `inferred`.

Verify: `go test ./internal/kanban/... -run TestResolveLink_InferredFromBody -v`

### AC-003 — several body tokens resolve `ambiguous`, never a best guess

**Given** the fixture set containing PR #1614 whose body carries the five tokens
`t1, t151, t203, t69, t9` and no title token
**When** the resolver is run for card `t151`, which also appears in another PR body
**Then** the result is labelled `ambiguous`, enumerates every candidate PR number,
**And** no single PR is returned as the resolved link.

Verify: `go test ./internal/kanban/... -run TestResolveLink_AmbiguousNotCollapsed -v`

### AC-004 — the queue file is byte-identical across a read-surface invocation

**This is the load-bearing AC for the read-only ruling in spec.md §B.**

**Given** a `backlog.json` fixture with a recorded SHA-256
**When** `moai todo pr` and `moai todo pr --json` are each invoked against it
**Then** the file's SHA-256 is unchanged after every invocation,
**And** its modification time is unchanged,
**And** this holds equally on the fail-open path (AC-005) and on the ambiguous
path (AC-003).

Verify: `go test ./internal/cli/... -run TestTodoPR_QueueFileByteIdentical -v`

The test asserts the digest, not merely "no error" — an error-free run that
rewrote the file with identical semantics but different bytes still fails.

### AC-005 — fail-open when `gh` is unavailable

**Given** a `PATH` from which `gh` is absent
**When** `moai todo pr` is invoked
**Then** the exit code is 0,
**And** the output renders the queue with the link column blank or absent,
**And** stderr carries a degradation note, not an error,
**And** AC-004's byte-identity assertion still holds.

Verify: `go test ./internal/cli/... -run TestTodoPR_FailOpenNoGh -v`

Repeat for `gh` present but exiting non-zero (unauthenticated / offline
simulation): `-run TestTodoPR_FailOpenGhNonZero`.

### AC-006 — commit messages are not a carrier

**Given** the fixture set containing PR #1600, whose 15 commit-message tokens do
not include its delivering card `t184`, and whose body carries `t184`
**When** the resolver is run for card `t131` (present in #1600's commit tokens,
absent from its title and body)
**Then** the resolver returns no link to #1600 for `t131`.

Verify: `go test ./internal/kanban/... -run TestResolveLink_IgnoresCommitTokens -v`

### AC-007 — no link is distinguishable from ambiguous

**Given** a card id present in no open PR title or body
**When** the resolver is run
**Then** it returns a no-link result whose kind the caller can distinguish from
an `ambiguous` result (distinct type or explicit field, not an empty candidate
list on an ambiguous record).

Verify: `go test ./internal/kanban/... -run TestResolveLink_NoLinkDistinctFromAmbiguous -v`

### AC-008 — whole-token matching

**Given** a fixture PR whose title carries `t200`
**When** the resolver is run for card `t20`
**Then** no link is returned.

Verify: `go test ./internal/kanban/... -run TestResolveLink_WholeTokenOnly -v`

### AC-009 — `moai todo list` stays network-free

**Given** a `PATH` from which `gh` is absent and a network-denying environment
**When** `moai todo list` is invoked with no flags
**Then** it exits 0 and renders the queue,
**And** no `gh` process is spawned (asserted via an injected executor recording
zero invocations).

Verify: `go test ./internal/cli/... -run TestTodoList_NoGhInvocation -v`

### AC-010 — confidence is rendered, and the JSON form carries it

**Given** a fixture producing one `exact`, one `inferred`, and one `ambiguous`
result
**When** `moai todo pr --json` is invoked
**Then** each record carries `card_id`, `pr` (number or candidate list),
`pr_state`, and `confidence`,
**And** the human render displays the confidence label next to every link.

Verify: `go test ./internal/cli/... -run TestTodoPR_RendersConfidence -v`

### AC-011 — the pre-dispatch cross-check exists in the doctrine

**Pre-implementation baseline (required):** the following returns **zero** hits
on the tree before M1 lands. Confirm the zero baseline before accepting this AC.

```
grep -c 'pre-dispatch PR cross-check' .claude/rules/moai/workflow/kanban-dispatch.md
```

**Given** the doctrine file after M1
**When** the grep is run
**Then** it returns at least 1,
**And** the surrounding clause is marked [HARD] and requires the lead to read
and report the card's PR state before dispatching out of `backlog`.

### AC-012 — the PR-title clause and its non-contradiction note exist

**Pre-implementation baseline (required):** zero hits before M1.

```
grep -c 'PR title MUST carry the delivering card id' .claude/rules/moai/workflow/kanban-dispatch.md
```

**Given** the doctrine file after M1
**When** the grep is run
**Then** it returns at least 1,
**And** the same section explicitly states that this does not contradict the
branch-name exclusion rule, naming the branch slug and the three existing
traceability carriers.

### AC-013 — template mirror parity

**Given** the doctrine edit in `.claude/rules/…/kanban-dispatch.md`
**When** the mirror at
`internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`
is compared
**Then** both clauses are present in the mirror,
**And** the neutrality guard passes.

Verify: `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -v`

---

## D.1 Severity

All ACs above are `must`. There are no `should`-severity criteria in this SPEC.

## D.2 Traceability

| REQ | AC |
|---|---|
| REQ-1.1 | AC-001, AC-002, AC-003, AC-010 |
| REQ-1.2 | AC-001 |
| REQ-1.3 | AC-002 |
| REQ-1.4 | AC-003 |
| REQ-1.5 | AC-006 |
| REQ-1.6 | AC-003 |
| REQ-1.7 | AC-007 |
| REQ-1.8 | AC-008 |
| REQ-2.1 | AC-004 |
| REQ-2.2 | AC-004 |
| REQ-2.3 | AC-005 |
| REQ-2.4 | AC-009 |
| REQ-2.5 | AC-009 |
| REQ-2.6 | AC-010 |
| REQ-2.7 | AC-010 |
| REQ-3.1 | AC-011 |
| REQ-3.2 | AC-012 |
| REQ-3.3 | AC-012 |
| REQ-3.4 | AC-013 |

## D.3 Quality gates

- `go vet ./...` clean.
- `golangci-lint run` clean on touched packages.
- Affected-package tests green locally; full-suite verdict from CI on the PR head.
- Coverage on the new resolver package ≥ 85%.

## D.4 Definition of Done

- All 13 ACs pass with cited command output.
- AC-004's digest assertion is present in the committed test file, not merely
  performed by hand.
- AC-011 and AC-012 each recorded their zero-baseline grep result before the
  doctrine edit landed.
- Doctrine and template mirror committed together; `make build` run.

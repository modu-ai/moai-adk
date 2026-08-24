# Progress — SPEC-KANBAN-PR-CARD-TRACEABILITY-001

## §E.1 Plan-phase Audit-Ready Signal

### Artifacts

Tier S set: `spec.md`, `plan.md`. Acceptance criteria are inline in `spec.md`
§D per the Tier S contract. `progress.md` is emitted at every tier and is not
counted in the Tier artifact total.

### Provenance

Split out of `SPEC-KANBAN-QUEUE-PR-SYNC-001` per audit finding D14
(`.moai/reports/t210/verdict.md`, iteration 1): that SPEC carried 19 leaf
requirements against a Tier M ceiling of 16, and its four doctrine requirements
are code-free and land on a different schedule from the tooling.

The split also makes an ordering explicit that was previously implicit inside a
single milestone sequence: the [HARD] cross-check clause is live and satisfied
by hand for the interval between this SPEC landing and the sibling shipping its
tooling. That interval is stated in `spec.md` §A rather than left for a reader
to discover.

Audit findings carried into this SPEC: D3 (REQ-002's report-and-re-decide
wording, which is the control against a de-facto-veto reading), D12 (each
acceptance criterion split into a mechanical half and a reviewer-judgement
half), D15 (mirror parity).

### SPEC ID check

```
$ ID="SPEC-KANBAN-PR-CARD-TRACEABILITY-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
PASS
```

No collision in `.moai/specs/`.

### Budget

```
$ grep -cE '^\*\*REQ-[0-9]{3}\*\*' .moai/specs/SPEC-KANBAN-PR-CARD-TRACEABILITY-001/spec.md
5
$ grep -cE '^### AC-[0-9]{3}' .moai/specs/SPEC-KANBAN-PR-CARD-TRACEABILITY-001/spec.md
4
```

5 leaf requirements and 4 acceptance criteria against a Tier S ceiling of 8 each.

### Audit iteration 2 — PASS (0.775 against the Tier S threshold of 0.75)

`.moai/reports/t210/verdict-2.md`. Two blocking findings, both lead-verified
rather than sent through a third round (the iteration ceiling is reached):

- **N3 — the split dropped a requirement.** The pre-split SPEC's REQ-3.4 (the
  template-mirror obligation) did not survive into this SPEC's requirement
  layer, which left AC-004 exercising nothing. Restored as **REQ-005**, with
  AC-004 mapped to it and a traceability table added. Tier S is at 5 of 8, so
  there was no budget obstacle — unlike the sibling, where the structurally
  correct fix for N1 would have breached 16/16.
- **N4 — a judgement inside a `**Mechanical.**` block.** AC-003's mechanical
  half asserted *"and the same section names all four traceability carriers"*
  while showing a grep covering only the first clause. Removed; the claim was
  already stated correctly in the judgement half, so nothing was lost. It was
  flagged blocking because it is a relapse of the D12 defect inside D12's own
  remedy — the reason this SPEC uses the split-AC pattern at all.

### Lint — verbatim

```
$ moai spec lint .moai/specs/SPEC-KANBAN-PR-CARD-TRACEABILITY-001/spec.md --json
[]
EXIT=0
```

### Zero baselines observed in this worktree (not relayed)

Each grep-shaped acceptance criterion is non-vacuous only against a zero
baseline. All three confirm at 0:

```
$ grep -c 'pre-dispatch PR cross-check' .claude/rules/moai/workflow/kanban-dispatch.md
0
$ grep -c 'confirms or withdraws' .claude/rules/moai/workflow/kanban-dispatch.md
0
$ grep -c 'PR title MUST carry the delivering card id' .claude/rules/moai/workflow/kanban-dispatch.md
0
```

The first and third were independently confirmed at 0 during the t210 audit;
`plan.md` §B requires re-running all three against the tree being edited rather
than citing these, because a baseline measured earlier proves nothing about the
tree at edit time.

### Gaps

- The reviewer-judgement halves of AC-001, AC-002, and AC-003 have no command.
  They are recorded as judgements and are not claimed as mechanical coverage.
- No CI check enforces the PR-title convention (`spec.md` §E). The obligation is
  doctrinal; mechanical enforcement is named as a plausible follow-up and is not
  specified.
- The `[HARD]`-marker grep in AC-001 is order-sensitive — it matches only when
  the marker precedes the phrase on the same line. `plan.md` §B records this so
  a correct clause is not failed by a grep that assumed the other order.

### Signal

`status: draft`. Awaiting audit, then Implementation Kickoff Approval. Lands
**before** `SPEC-KANBAN-QUEUE-PR-SYNC-001`.

## §E.2 Run-phase Evidence

Card t210. Worktree `.claude/worktrees/t210`, branch `WT-queue-pr-sync`.
Doctrine only — no Go file is touched, per `plan.md` §D. Two files change:

| File | Milestone |
|---|---|
| `.claude/rules/moai/workflow/kanban-dispatch.md` | M1, M2 |
| `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | M3 (mirror) |

### Pre-flight — the zero baselines, re-measured against THIS tree

`plan.md` §B is explicit that the audit-time baselines prove nothing about the
tree being edited, so all three were re-run at HEAD `985343fad`, before any
edit:

```
$ grep -c 'pre-dispatch PR cross-check' .claude/rules/moai/workflow/kanban-dispatch.md
0
$ grep -c 'confirms or withdraws' .claude/rules/moai/workflow/kanban-dispatch.md
0
$ grep -c 'PR title MUST carry the delivering card id' .claude/rules/moai/workflow/kanban-dispatch.md
0
```

All zero. Both edit targets exist, and — measured rather than assumed — the
live file and its mirror were **byte-identical** before the edit:

```
$ diff .claude/rules/moai/workflow/kanban-dispatch.md \
       internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md
(no output)
```

That identity is what makes M3 a one-step mirror rather than a translation:
the same neutral prose satisfies both copies, so it was written to both.

### M1 — the pre-dispatch cross-check clause (REQ-001, REQ-002)

Sited in § Entry into the board is an operator act, immediately after the third
existing [HARD] clause ("The lead may attach a finding; it may not act on one")
— that section already owns the operator-authority boundary, and the new clause
sits inside it rather than beside it.

Two [HARD] paragraphs, deliberately separate:

- The first states the obligation: read the card's pull-request and landed
  state before dispatching, and report what was read **in the same turn**. It
  names `moai todo pr <id>` and the by-hand equivalent, so the clause is
  satisfiable now and better satisfied once the sibling SPEC's tooling lands.
- The second is the control against a de-facto veto (REQ-002). It carries the
  literal `confirms or withdraws`, states that the lead does **not** withhold
  the dispatch on its own authority, and grounds that in the existing
  "promotion is the operator's act, always" clause it sits under.

The `[HARD]` marker precedes the phrase `pre-dispatch PR cross-check` on the
same line, per `plan.md` §B's order-sensitivity note.

```
$ grep -c 'pre-dispatch PR cross-check' .claude/rules/moai/workflow/kanban-dispatch.md
1
$ grep -c '\[HARD\].*pre-dispatch PR cross-check' .claude/rules/moai/workflow/kanban-dispatch.md
1
$ grep -c 'confirms or withdraws' .claude/rules/moai/workflow/kanban-dispatch.md
1
```

### M2 — the PR-title clause and the non-contradiction note (REQ-003, REQ-004)

Sited in § Isolation, immediately after the paragraph closing the branch-naming
rule ("A lane that reports a branch name without also reporting its card id…"),
so the reconciliation stands next to the rule it reconciles with rather than a
section away.

```
$ grep -c 'PR title MUST carry the delivering card id' .claude/rules/moai/workflow/kanban-dispatch.md
1
```

The clause states the non-contradiction outright — the branch name is read by a
human scanning `git branch` and wants a slug; the PR title is read by a machine
and wants the id — lists all **four** carriers with the PR title marked as the
only machine-readable one, and scopes the obligation to card-delivering pull
requests, naming release, batch, and maintenance pull requests as carrying
none. It also states the forward-only binding: pull requests opened after the
clause lands, no retitling.

### M3 — the mirror

```
$ grep -c 'pre-dispatch PR cross-check' internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md
1
$ grep -c '\[HARD\].*pre-dispatch PR cross-check' internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md
1
$ grep -c 'confirms or withdraws' internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md
1
$ grep -c 'PR title MUST carry the delivering card id' internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md
1
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	25.962s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.950s
?   	github.com/modu-ai/moai-adk/internal/template/scripts	[no test files]
$ make build
… catalog.yaml updated successfully (12899 bytes)
go build -ldflags "…" -o bin/moai ./cmd/moai
```

**Neutrality was written in, not stripped afterwards.** `plan.md` §G's second
anti-pattern is copying the live file over the mirror; here the clauses were
authored neutral from the start — no SPEC ID, no `.moai/reports/` path, no
date, no commit SHA — and "the integration branch" is named by role rather than
by branch name. The measured incident counts ("five cards", "one card") were
deliberately written as "cards sat queued" and "one sat queued", because a
count is a fact about this repository's history and not about a user's project.

### Always-loaded budget — measured, and the stub/companion split

`kanban-dispatch.md` is always-loaded, so every clause costs every session. The
file is already a stub + lazy-companion pair, and the lead ruled that the new
doctrine follow the same convention: the binding [HARD] clauses stay in the
stub, the rationale moves to `kanban-dispatch-detail.md` (`paths:`-scoped, so it
is NOT always-loaded).

What stayed in the stub: the cross-check obligation, the report-never-veto
clause with its literal `confirms or withdraws`, the PR-title requirement with
its one-line non-contradiction, the four carriers, and the scope restriction.
Every acceptance criterion still greps against the stub.

What moved to the companion: the incident narrative, the reasoning for why the
veto reading is invisible after the fact, the "neither name can serve both
readers" argument, the carrier precision/recall table, the inherited-commit
worst case, and why a card token in a batch pull-request title is worse than
none.

Measured before and after the split, on this tree:

```
$ go test ./internal/config/ -run TokenBudget -count=1 -v
# before the trim
    token_budget_guard_test.go:69: always-loaded surface = 68091 tokens (budget 76000, headroom 7909, 18 entries)
# after the trim
    token_budget_guard_test.go:69: always-loaded surface = 67698 tokens (budget 76000, headroom 8302, 18 entries)
```

| | bytes | tokens | delta vs `origin/main` |
|---|---|---|---|
| `origin/main` baseline | 27,777 | 67,300 (surface) | — |
| first draft | 30,943 | 68,091 | +3,166 B / +791 tok |
| after the stub/companion split | 29,368 | 67,698 | **+1,591 B / +398 tok** |

The cost is halved and the rationale is preserved rather than deleted — the
companion grew by the text the stub gave up, and the companion is not
always-loaded.

**Measurement provenance (the lead asked for the command, not the number).**
The figure is `internal/config` `measureAlwaysLoaded`, whose surface is: every
`.claude/rules/moai/**/*.md` whose frontmatter carries NO `paths:` key, sorted,
plus four fixed slots (`CLAUDE.md`, `AGENTS.md`,
`.claude/output-styles/moai/moai.md`, `MEMORY.md`); `MEMORY.md` is measured as
its load head only (first 200 newlines or 25 KiB, whichever comes first) and is
absent in this tree, so it contributes 0. Tokens are `len(bytes) / 4`. The
enumeration is 18 entries; the per-file breakdown is in the completion report.

## §E.3 Run-phase Audit-Ready Signal

### Claim

All four acceptance criteria are implemented. The mechanical half of each
passes; the reviewer-judgement half of AC-001, AC-002, and AC-003 is recorded
as an outstanding judgement and is **not** claimed as a mechanical pass.

### Evidence

| AC | Mechanical verification | Result |
|---|---|---|
| AC-001 | the two greps above, against a cited zero baseline | 0 → 1 on both |
| AC-002 | `grep -c 'confirms or withdraws'`, against a cited zero baseline | 0 → 1 |
| AC-003 | `grep -c 'PR title MUST carry the delivering card id'`, against a cited zero baseline | 0 → 1 |
| AC-004 | the four mirror greps + `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...` + `make build` | all present, suite green, build run |

### Baseline-attribution

Measured against HEAD `985343fad` in `.claude/worktrees/t210` on branch
`WT-queue-pr-sync`. The three zero baselines were re-run on this tree at edit
time rather than cited from the audit, per `plan.md` §G's fourth anti-pattern.
The pre-edit byte-identity of the live file and its mirror was measured with
`diff`, not assumed.

### Gaps

- **The reviewer-judgement halves are unverified.** No independent reviewer has
  confirmed that the cross-check clause requires reading *and reporting*
  (AC-001), that it cannot be read as authorizing a lead-side veto (AC-002), or
  that the non-contradiction note reads correctly against the branch-name rule
  (AC-003). The author's reading is that all three hold; that is a judgement,
  and it is recorded as one rather than reported as a pass.
- **Nothing mechanically enforces the PR-title convention.** No CI check, hook,
  or lint rule was added — `spec.md` §E excludes it. Adherence is doctrinal.
- **The doctrine-before-tooling interval is live.** Until the sibling SPEC
  merges, the cross-check is satisfied by hand. `spec.md` §A accepts this cost
  explicitly; it is named here because it is a real operating condition and not
  a paper one.

### Residual-risk

- **The [HARD]-marker grep is order-sensitive** and matches only when `[HARD]`
  precedes the phrase on one line. A future editor who rewraps the paragraph or
  moves the marker fails AC-001 on a clause that is still correct. The
  order-sensitivity is recorded in `plan.md` §B; the grep is not made smarter.
- **The clause sits next to three existing [HARD] clauses about operator
  authority.** It was written to be compatible with "promotion is the
  operator's act, always" rather than as an exception to it, but the reading is
  the thing under risk, and only the wording controls it — which is exactly why
  AC-002 exists as a separate criterion with its own judgement half.
- **Four carriers now, three named in the paragraph above the new clause.** The
  pre-existing "three other carriers" sentence is left intact and the new
  clause states the fourth explicitly; a reader who stops at the older sentence
  sees three. Rewriting the older sentence was rejected as scope the card does
  not carry.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

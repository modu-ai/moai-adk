# SPEC-RELATION-PICKUP-FILTER-001 — Progress

Tier S (dispatcher-mandated acceptance layer added) · card t1343 · plan-phase
artifact set authored 2026-09-29 at HEAD `113082295` (worktree
`.moai/worktrees/t1343`, branch `WT-relation-pickup-filter`). Status: `draft`
(plan-phase creation per ownership; run phase not entered).

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifact set complete: `spec.md` (7 REQ, GEARS — Ubiquitous 3 /
  When 3 / Unwanted 1), `plan.md` (§A-§H, 4 milestones, cycle_type=tdd),
  `acceptance.md` (AC-RPF-001..007, Given-When-Then, RED-now/green 채택 셀,
  품질 게이트 + DoD), this `progress.md`. The acceptance layer is
  dispatcher-mandated (card t1343 gate: release-blocking criteria name command
  + expected output); Tier S scope unchanged (< 300 LOC, 2 production files).
- SPEC ID regex pre-write check: run as Bash, verbatim output `PASS`
  (`SPEC-RELATION-PICKUP-FILTER-001`); ID unique in `.moai/specs/` (no
  `SPEC-RELATION-*` / `SPEC-PICKUP-*` directory existed).
- Frontmatter validated against the canonical 12-field schema SSOT; `status:
  draft` set at creation; no snake_case aliases.
- Measured input verification (all in this tree, HEAD `113082295`):
  (b) record-only claim verified at `internal/kanban/backlog_store.go:139-148`
  (the card text's `:135-138` citation predates a comment growth — quoted in
  spec.md §A.2 F-1); (c) pickup-path finding verified —
  `autoPickTargets` (`internal/cli/todo_auto.go:142`), queued arm appends every
  queued card (`:162-166`), zero relation consumers outside definitions
  (non-test grep). Cycle-guard gap verified at `runTodoRelate`
  (`internal/cli/todo_relate.go:60-94`).
- Merge-vs-separate decision recorded: **t1343 stands alone** — t1338 has no
  landed SPEC (no `.moai/specs/` hit; report hits are coincidental test
  temp-dir substrings in t60 artifacts; the proposal names no card for P3) and
  its queued card text enumerates four factory-autonomy pieces that exclude
  the relation filter. Full rationale: spec.md §A.4.
- Out of Scope section carries four `### Out of Scope — <topic>` H3
  sub-headings with `-` bullets (factory-lease t1240 boundary / t1338
  absorption / display surfaces / relation vocabulary policy).
- Post-authoring lint: `go run ./cmd/moai spec lint SPEC-RELATION-PICKUP-FILTER-001`
  (tool built from this tree, HEAD `113082295`) → `✓ No findings — all SPEC
  documents are valid` (0 errors, 0 warnings). One authoring fix during
  bring-up: the first lint pass emitted 7 `CoverageIncomplete` warnings; the
  AC sections carried REQ references in prose but lacked the parser's
  `**Covers**: maps REQ-…` mapping clause — added to all seven ACs (matches
  the SPEC-TODO-CLASSIFY-DISPATCH-001 convention), lint re-run clean.
- Known bounded gaps: none at plan-phase. §G risks R-1..R-3 are design
  boundaries, not gaps.

## §E.2 Run-phase Evidence

Run entered 2026-09-30 on branch `WT-relation-pickup-filter` at plan baseline
HEAD `28e672b39`, tree clean. cycle_type=tdd, milestones M1-M4 in plan.md §F
order. RED-before-GREEN observed per milestone; verbatim outputs below.

### M1 — the pickup filter (REQ-RPF-001/002/003, AC-RPF-001..003)

**RED (observed on the pre-implementation tree, HEAD `28e672b39`, tests added
first):**

```
$ go test -timeout 30m -count=1 ./internal/cli/ -run 'TestAutoPickTargets'
--- FAIL: TestAutoPickTargetsRelationBlocked (0.30s)
    --- FAIL: TestAutoPickTargetsRelationBlocked/depends_blocks_its_subject (0.17s)
        todo_relation_filter_test.go:57: pickup = [t1,t2], want [t2] — t1 must be excluded while t2 stays a candidate
    --- FAIL: TestAutoPickTargetsRelationBlocked/blocks_blocks_its_related (0.14s)
        todo_relation_filter_test.go:57: pickup = [t1,t2], want [t1] — t2 must be excluded while t1 stays a candidate
--- FAIL: TestAutoPickTargetsReturnsAfterDone (0.14s)
    todo_relation_filter_test.go:89: blocked card t2 was a candidate before the predecessor's done
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.230s
```

RED for the right stated reason: both direction arms failed because the
blocked-side card was still a pickup candidate (no relation consumer in the
queued arm), and the done-resolution arm failed before any archive move. (A
test-side Given fix happened between this RED and GREEN: the done-resolution
test originally recorded `depends {t1,t2}` — whose blocked side is the
subject t1, not the successor t2 — and was corrected to `t2 depends t1`
before GREEN; the direction arms above were unchanged.)

**GREEN (same command, after implementation):**

```
$ go test -timeout 30m -count=1 ./internal/cli/ -run 'TestAutoPickTargets'
ok  	github.com/modu-ai/moai-adk/internal/cli	1.348s
```

Implementation: `kanban.WaitsOnOf` (direction normalization SSOT — `depends`
{S,R} → S waits on R; `blocks` {S,R} → R waits on S) + `BacklogRecord.FindingsBlocking`
(finding-existence predicate per spec.md B.1) in
`internal/kanban/backlog_store.go`; the queued arm of `autoPickTargets`
(`internal/cli/todo_auto.go`) excludes cards with a non-empty blocking set.
Dead-owner rescue arm untouched (REQ-RPF-006).

**Regression check (M1):**

```
$ go test -timeout 30m -count=1 ./internal/cli/ -run 'TestTodo|TestAuto|TestRelate|TestSemantic|TestMachine|TestFinding'
ok  	github.com/modu-ai/moai-adk/internal/cli	460.801s
$ go vet ./internal/cli/... ./internal/kanban/...
(no output) exit 0
```

### M2 — the cycle guard (REQ-RPF-005, AC-RPF-005/006)

**RED (observed on the tree at M1 HEAD `938e43f61`, tests added first):**

```
$ go test -timeout 30m -count=1 ./internal/cli/ -run 'TestTodoRelateRefusesCycle|TestTodoRelateCycleGuardShapes'
--- FAIL: TestTodoRelateRefusesCycle (0.61s)
    todo_relation_filter_test.go:137: todo relate t2 t1 depends closed a 2-cycle and was accepted, want a refusal
--- FAIL: TestTodoRelateCycleGuardShapes (2.72s)
    --- FAIL: TestTodoRelateCycleGuardShapes/3-cycle_through_an_intermediate_card_is_refused (0.90s)
        todo_relation_filter_test.go:200: candidate [t3 t1 depends] was accepted, want a refusal
    --- FAIL: TestTodoRelateCycleGuardShapes/blocks_spelling_closes_a_cycle_like_depends (0.52s)
        todo_relation_filter_test.go:200: candidate [t2 t1 blocks] was accepted, want a refusal
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	4.192s
```

RED for the right stated reason: every refused-expected cycle shape (direct
2-cycle, 3-cycle through an intermediate, blocks spelling) was ACCEPTED —
the write path has no graph walk (F-3). The allowed shapes (open chain,
same-pair opposite spelling) passed already, as expected with no guard.

**GREEN (same command plus the pre-existing relate regression surface):**

```
$ go test -timeout 30m -count=1 ./internal/cli/ -run 'TestTodoRelateRefusesCycle|TestTodoRelateCycleGuardShapes|TestTodoRelate|TestRelateAndUnrelate'
ok  	github.com/modu-ai/moai-adk/internal/cli	9.008s
```

Implementation: `BacklogRecord.WaitsOnClosesCycle` + `waitsOnReaches` in
`internal/kanban/backlog_store.go` (BFS from the candidate's target back to
its waiter over WaitsOnOf edges); the check runs in `runTodoRelate`
(`internal/cli/todo_relate.go`) before `AppendFindingOnce`, and the refusal
names both endpoints. One interim defect found by GREEN and fixed in place:
the guard was first seeded with a finding lacking the endpoint ids, so the
empty waiter==target self-reach refused every write — caught by the allowed
shapes' seed relates, corrected to seed the guard with the real
{subject, related, relation} before any GREEN verdict was taken.

**kanban package suite (M1+M2 changes together):**

```
$ go test -timeout 30m -count=1 ./internal/kanban/
ok  	github.com/modu-ai/moai-adk/internal/kanban	191.542s
```

### M3 — labelled skips + doctrine text (REQ-RPF-004/007, AC-RPF-004)

**Test-side Given corrections observed before RED (recorded for honesty):**
the first M3 test draft built the "fully blocked queue" through `todo relate`
(two independent blocks/depends pairs) — wrong on two counts: the free
predecessors were legitimately still candidates (so no-eligible could never
fire), and a fully-blocked queue is BY DEFINITION a relation cycle, which the
M2 guard refuses to record through relate at all. The corrected Given seeds
the mutual-depends record directly (seedFindings), i.e. the pre-guard legacy
record the filter must still consume. A first RED run of the draft also
exposed a test-clock defect: a frozen `now` seam with `wait: 1ns` never let
the evidence deadline expire, polling forever — the clock seam now advances
one minute per poll tick.

**RED (labels reverted; observed on the tree at M2 HEAD `46e3ec0da` plus the
corrected test):**

```
$ go test -timeout 30m -count=1 ./internal/cli/ -run 'TestRunAutoCycleSkipsBlockedCards'
--- FAIL: TestRunAutoCycleSkipsBlockedCards (0.27s)
    todo_relation_filter_test.go:248: t1 skip label missing or wrong:
    todo_relation_filter_test.go:251: t2 skip label missing or wrong:
    todo_relation_filter_test.go:287: blocks skip label missing or wrong:
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.061s
```

RED for the right stated reason: the filter and the no-eligible report
already worked (no `accept` line, no-eligible present) — ONLY the labelled
skip lines were absent.

**GREEN (same command + the AC-RPF-004 green-path command surface):**

```
$ go test -timeout 30m -count=1 ./internal/cli/ -run 'TestRunAutoCycleSkipsBlockedCards'
ok  	github.com/modu-ai/moai-adk/internal/cli	1.058s
$ go test -timeout 30m -count=1 ./internal/cli/ -run 'TestAutoPickTargets|TestRunAutoCycle'
ok  	github.com/modu-ai/moai-adk/internal/cli	1.318s
```

Implementation: the queued arm of `autoPickTargets` emits one labelled
non-finding per blocking finding — `non-finding: <card> skipped
(relation-blocked: <subject> <relation> <related>) — waiting for predecessor
<target>` — card id, relation, and predecessor id all named (REQ-RPF-004);
the no-eligible path keeps its exit-0 contract. Doctrine text B-I-1
rewritten in both blocks: `backlog_store.go` blocks/depends constant comments
now name `autoPickTargets` as the consuming pickup path and
`WaitsOnClosesCycle` as the write-time check; `todo_relate.go` header comment
states the sequencing pair is no longer purely observational while the other
four relations stay record-only.

### M4 — regression pins (REQ-RPF-006/007, AC-RPF-007)

Pin tests (per the acceptance.md 채택 셀 these are over-implementation
detectors — their RED-now is a FUTURE regression, so the run observes GREEN
directly; the same-pair opposite-spelling pin was already absorbed into
M2's `TestTodoRelateCycleGuardShapes` fourth case):

```
$ go test -timeout 30m -count=1 ./internal/cli/ -run 'TestAutoPickTargetsRescueArmUnfiltered|TestAutoPickTargetsNonSequencingRelation'
ok  	github.com/modu-ai/moai-adk/internal/cli	1.145s
```

- `TestAutoPickTargetsRescueArmUnfiltered` — a `picked` card whose owner
  measures dead is taken over even while a live `blocks` finding names it
  (the rescue arm is not gated, REQ-RPF-006).
- `TestAutoPickTargetsNonSequencingRelation` — a queued card named only by
  a `contains` finding stays a candidate (the filter consumes exactly
  blocks and depends, REQ-RPF-006).

**GTD-table independence (REQ-RPF-007):**

```
$ grep -cn "gtd_relations" internal/cli/todo_auto.go internal/cli/todo_relate.go internal/kanban/backlog_store.go
internal/cli/todo_relate.go:0
internal/cli/todo_auto.go:0
internal/kanban/backlog_store.go:0
(exit 1 = zero hits — the filter and the guard read only BacklogFinding records)
```

### Final verification batch (all run in this session, this tree)

**Affected-packages full suite (AC-RPF-007 green command):**

```
$ go test -timeout 30m -count=1 ./internal/cli/ ./internal/kanban/
ok  	github.com/modu-ai/moai-adk/internal/cli	1706.452s
ok  	github.com/modu-ai/moai-adk/internal/kanban	191.106s
```

**Lint (CI-pinned golangci-lint v2.1.6 — matches .github/workflows/ci.yml:464):**

```
$ golangci-lint run ./internal/cli/... ./internal/kanban/...
0 issues.
```

**Vet + cross-platform build:**

```
$ go vet ./internal/cli/... ./internal/kanban/...
(no output) exit 0
$ go build ./... && GOOS=windows GOARCH=amd64 go build ./...
DARWIN_BUILD_OK
WINDOWS_BUILD_OK
```

**Coverage of the new code (coverpkg cross-measurement, filtered tests):**

```
$ go test -coverprofile=... -coverpkg=github.com/modu-ai/moai-adk/internal/kanban -run 'TestAutoPickTargets|TestRunAutoCycle|TestTodoRelate|TestRelateAndUnrelate' ./internal/cli/
WaitsOnOf            100.0%
FindingsBlocking     100.0%
WaitsOnClosesCycle   100.0%
waitsOnReaches       100.0%
runTodoRelate        100.0%
autoPickTargets       76.2%  (whole function incl. pre-existing liveness arms)
```

All four new kanban functions and the guarded write path are fully covered
by the relation-filter tests. The kanban PACKAGE figure measured standalone
(`go test -cover ./internal/kanban/` → `coverage: 84.7% of statements`)
does not attribute cli-side tests; a pre-change (HEAD `28e672b39`) package
baseline was not measured — checking out a second tree at the baseline
commit is outside this card's scope, so the "85% maintained" claim is
carried by the function-level figures above (new code fully covered cannot
lower a package figure) and is noted as a bounded gap in §E.3.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-09-30"
run_commit_sha: "pending-backfill-run"
run_status: "all milestones landed, all gates green"
ac_pass_count: 7
ac_fail_count: 0
ac_matrix: "AC-RPF-001..007 PASS (green commands + verbatim outputs in §E.2)"
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "not run — isolated card worktree on WT-relation-pickup-filter; dispatch forbids moai integration commands and push"
l44_post_push_fetch: "n/a — no push performed (leader batch-pushes develop)"
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin: "ok (go build ./...)"
cross_platform_build.windows: "ok (GOOS=windows GOARCH=amd64 go build ./...)"
total_run_phase_files: 6
m1_to_m4_commit_strategy: "one commit per milestone (M1 938e43f61, M2 46e3ec0da, M3 3019af585, M4 this commit) + one run_commit_sha backfill commit"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

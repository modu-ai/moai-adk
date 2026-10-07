# Progress — SPEC-GRAPH-CARD-SQUASH-EDGE-001

status: draft
card: t1559 (lane-3, self-dispatch run tmhxo0)
spec: SPEC-GRAPH-CARD-SQUASH-EDGE-001

## Plan-phase record

- 2026-10-07: Tier M artifact set authored (spec.md / plan.md / acceptance.md / this skeleton)
  at plan base df0c417e9. Baseline measurements recorded in spec.md §1 and plan.md §C (HEAD
  single-parent verified via `git rev-list --parents -1`; walk breadth 13127 vs 2664;
  attribution form-2b probe 1 match; `walkCardMerges` count 4; matcher-free layer 0 matches).
- 2026-10-07 (plan-audit r1): **FAIL** — 5 blocking (D1 scope AC impossible, D2 measured-cost
  premise false, D3 sha^1 guard unexercised, D4 plan command setup failure, D5 AC evidence
  form). Verdict `.moai/reports/t1559/plan-audit-verdict.md` (audited_sha 08de1c43c, receipt
  rcpt-e8ccfe80a979d3060d61cf30, score 0.90, iter 1/2).
- 2026-10-07 (D4 lane resolution): the D4 setup failure was the embed-manifest break whose heal
  had already landed on origin/main (f97edcc55 via PR #1762, plus #1783/#1784). Lane merged
  origin/main into the card branch (merge 5fb7baf88); measured post-absorb in this run:
  `go build ./internal/template/` exit 0, `go test -count=1 ./internal/graph/` ok 107.741s.
  No separate repair card, no SPEC scope extension — the auditor's two offered options were
  both moot at decision time.
- 2026-10-07 (repair r1): manager-spec applied 16 fix_scope hunks — D1 pathspec restriction,
  D2 §3 measured rewrite with the LANE DECISION recorded (intermediate-commit attribution
  ACCEPTED), D3 card-attributed root contrast group, D5 -v + rationale correction, base re-pinned
  5fb7baf88, plan §C.1 absorbed-state precondition. Lane spot-verified the hunks on disk
  (stale-text sweep 0 hits; `-- internal/`, `-v`, card-t1561 contrast all present). Delta
  re-audit (iter 2/2, reread_hunks) dispatched by the lane.
- 2026-10-07 (plan→run Kickoff gate): **MET — autonomous form.** Evidence: plan-auditor iter-2
  delta verdict **PASS** (score 0.97 ≥ Tier M 0.80, blocking 0, audited_sha 122c83085, artifact
  hash 6dfcd9e2…53ab3 pinned, receipt rcpt-e8ccfe80a979d3060d61cf30), no open blocker, SPEC
  artifacts unchanged since the verdict. Verdict file:
  `.moai/reports/t1559/plan-audit-verdict.md`. Ladder step ① (disk evidence) resolved the
  gate; the card names no operator gate. SPEC bodies are hash-frozen from here — any body
  edit re-invalidates the audit.

## §F Phase 4 Mode Selection

- Inputs: tier M, scope = 2 files (internal/graph/card_file.go + card_file_test.go),
  domains = 1 (Go source), language mix = Go, concurrency benefit = LOW (coding-heavy).
- Evaluation: direct — not trivial, no. fanout — not multi-domain research, no.
  sweep — not ≥30-file mechanical, no. agent-team — no operator request, excluded.
- **Decision: serial** — one manager-develop spawn, cycle_type=tdd per the SPEC,
  coding-heavy per Anthropic's parallelism caveat.
- Gate confirmation: plan→run Kickoff gate met in autonomous form (recorded above);
  no outstanding user preferences (card dispatched by the leader, lenses fixed).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-07T00:49:24Z
verdict: PASS — plan-audit iteration 2/2, overall 0.97, blocking 0, must_pass_failed 0
audited_sha: 122c83085db56c10a0ec06a51b2a2745e29870be
plan_artifact_hash: 6dfcd9e2bca3b64aefa674377ee6a9eadd8b75682f4e4230eb57b12d1ff53ab3
receipt: rcpt-e8ccfe80a979d3060d61cf30
verdict_file: .moai/reports/t1559/plan-audit-verdict.md
history: r1 FAIL (iter 1/2, 0.90, 5 blocking D1-D5, audited_sha 08de1c43c) → manager-spec fix_scope repair (spec §3, plan §A/B/C/E/F, acceptance §A/AC-002/007/010/011; base re-pin df0c417e9 → 5fb7baf88 via the lane's origin/main absorb) → r2 delta re-audit PASS (iter 2/2, 0.97, reread scope).

## §E.2 Run-phase Evidence

Run executed by manager-develop (cycle_type=tdd), card t1559, tree `.moai/worktrees/t1559`,
branch `WT-graph-card-merges-edge`, plan base pinned 5fb7baf88. Plan Audit Gate **SKIPPED**
per the single authoritative skip contract: verdict PASS (iter 2/2, 0.97 ≥ Tier M 0.80,
audited_sha 122c83085, artifact hash unchanged — the only commit after the verdict,
de7343621, added progress.md §E.1, which is not a hash subject).

### Pre-flight (all measured at de7343621, before any change)

| # | Command | Observed |
|---|---------|----------|
| P1 | `git branch --show-current` | `WT-graph-card-merges-edge` |
| P2 | `git rev-parse --short HEAD` | `de7343621` |
| P3 | `go build ./...` | exit 0 (C7 absorbed-state precondition reconfirmed) |
| P4 | `golangci-lint run ./internal/graph/...` | `0 issues.` exit 0 (baseline) |
| P5 | `grep -rn "Retired\|superseded" internal/graph/ --include="*.go"` | 0 hits (no cross-SPEC policy conflict) |
| P6 | `ls internal/graph/card_file.go internal/graph/card_file_test.go` | both present — the entire write scope |

### M1 — RED (test-only commit f6476244f)

Added `TestGraphCardFileEdgesSeeSquashLanding` to `internal/graph/card_file_test.go`
(117 insertions, 0 deletions — no existing fixture pattern edited). The commit also flips
`spec.md` frontmatter `status: draft` → `status: in-progress` (the sanctioned M1 transition;
`updated: 2026-10-07` already current).

RED witnessed at de7343621 + the new (uncommitted) test, before any card_file.go change —
verbatim output:

```
=== RUN   TestGraphCardFileEdgesSeeSquashLanding
    card_file_test.go:383: want exactly one edge for the squash landing, got []
--- FAIL: TestGraphCardFileEdgesSeeSquashLanding (10.98s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/graph	11.768s
FAIL
test_exit=1
```

Expected RED shape: assertion failure (0 edges, want 1) — not `[setup failed]`.

### M2 — GREEN (fix commit 0d32f5733)

`internal/graph/card_file.go`: removed `--merges` from the walk argv; renamed
`walkCardMerges` → `walkCardCommits`, `mergeInfo` → `commitInfo` (file-local, plan §D
sanctioned; the `merges` locals typed `[]commitInfo` renamed `commits` as the same-file
cascade of the sanctioned type rename); broadened the file-top, `CardFileAttributor`,
`CardFileEdge`, walk, `CardFileEdges`, per-commit diff, `CardAttributedMergeSHAs`, and
`CardMergeFingerprint` doc comments to commit-level semantics; error string
`card_file: log merges:` → `card_file: log:`; root-commit guard kept with its rationale
broadened to "commit". No exported identifier renamed. LOCAL cascade disclosure: the
`merges`→`commits` local-variable rename is beyond the plan §F-M2 letter, inside the
sanctioned file, required by the type rename it ripples from.

GREEN — new test + the six existing tests, prefix-open selector, verbatim:

```
$ go test ./internal/graph/ -run '^TestGraphCardFileEdgesSeeSquashLanding$' -v -count=1
=== RUN   TestGraphCardFileEdgesSeeSquashLanding
--- PASS: TestGraphCardFileEdgesSeeSquashLanding (13.53s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/graph	14.274s
test_exit=0
```

```
$ go test ./internal/graph/ -run '^TestGraph(CardFile|CheckNoticesCardFileSource)' -v -count=1
=== RUN   TestGraphCardFileEdgesDeterministic
--- PASS: TestGraphCardFileEdgesDeterministic (16.50s)
=== RUN   TestGraphCardFileEdgesCarryNoQueueState
--- PASS: TestGraphCardFileEdgesCarryNoQueueState (17.92s)
=== RUN   TestGraphCheckNoticesCardFileSource
--- PASS: TestGraphCheckNoticesCardFileSource (13.68s)
=== RUN   TestGraphCardFileEdgesSeeAbsorbedMerge
--- PASS: TestGraphCardFileEdgesSeeAbsorbedMerge (15.59s)
=== RUN   TestGraphCardFileEdgesIgnoreAbsorbMerge
--- PASS: TestGraphCardFileEdgesIgnoreAbsorbMerge (7.42s)
=== RUN   TestGraphCardFileEdgesKeepNonASCIIPaths
--- PASS: TestGraphCardFileEdgesKeepNonASCIIPaths (0.72s)
=== RUN   TestGraphCardFileEdgesSeeSquashLanding
--- PASS: TestGraphCardFileEdgesSeeSquashLanding (10.04s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/graph	82.421s
test_exit=0
```

Seven `-v` PASS lines at M2 time: six existing + the new one. Absorb breadth (MU-86) and
absorb refusal (MU-87) preserved unmodified.

### M3 — gates and AC matrix

| Gate | Command | Observed |
|------|---------|----------|
| vet | `go vet ./internal/graph/...` | exit 0 |
| lint | `golangci-lint run ./internal/graph/...` | `0 issues.` exit 0 (first attempt hit transient `parallel golangci-lint is running` exit 3 — lock contention, retried clean) |
| tests+coverage | `go test -cover -timeout 30m ./internal/graph/...` | `ok ... internal/graph 446.365s coverage: 87.9% of statements` / `ok ... internal/graph/symbol 1.047s coverage: 88.0% of statements` — exit 0; 87.9% ≥ 85% target. The 446.365s wall-clock is the measured cost of the broadened walk at M3 (D2 residual satisfied with a measured number) |
| windows build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| darwin build | `go build ./...` | exit 0 (pre-flight P3, same tree state pre-change; post-change tree built green inside the test runs) |

| AC | Status | Evidence |
|----|--------|----------|
| AC-GCSE-001 | PASS | RED verbatim above (0 edges pre-fix) → GREEN PASS post-fix |
| AC-GCSE-002 | PASS | base arm (zero edges + empty SHA list) and root contrast (`nil` error + zero edges, form-1 scope attribution through the gate, sha^1 guard exercised) both asserted green inside `TestGraphCardFileEdgesSeeSquashLanding` |
| AC-GCSE-003 | PASS | `shas == [landing full SHA]`, `fp != baseFP` asserted in the same test |
| AC-GCSE-004 | PASS | `grep -cE "regexp\\.|MatchString" internal/graph/card_file.go` → `0`, exit 1 |
| AC-GCSE-005 | PASS | `TestGraphCardFileEdgesSeeAbsorbedMerge` + `TestGraphCardFileEdgesIgnoreAbsorbMerge` PASS unmodified (7-test run above) |
| AC-GCSE-006 | PASS | `TestGraphCardFileEdgesDeterministic` PASS unmodified |
| AC-GCSE-007 | PASS | six existing tests named on their own `-v` PASS lines; `git diff --numstat 5fb7baf88 HEAD -- internal/graph/card_file_test.go` → `117	0` (additions only) |
| AC-GCSE-008 | PASS | `grep -c walkCardMerges card_file.go` → `0` exit 1; repo-wide `grep -rn walkCardMerges internal/` → 0 hits exit 1; `grep -c "merge commits whose subject" card_file.go` → `0` exit 1 (no residual merge-semantics claim); doc comments state commit-level semantics |
| AC-GCSE-009 | PASS | vet 0 / lint `0 issues.` / package suite ok — all exit 0 |
| AC-GCSE-010 | PASS | `git diff --name-only 5fb7baf88 HEAD -- internal/` → exactly `internal/graph/card_file.go` + `internal/graph/card_file_test.go` |
| AC-GCSE-011 | PASS | `git log --oneline 5fb7baf88..HEAD` at M2: `0d32f5733 fix(...)` newest, `f6476244f test(...)` next-older — the test commit precedes the fix commit, witnessed by the commit graph (this evidence commit follows M2; M1→M2 relative order unchanged) |

### E4 boundary grep

```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/graph/ --include="*.go" | grep -v "_test.go" | grep -v "// "
(no output)
boundary_exit=1
```

No subagent-boundary violation in the affected package.

### Gaps

- The repository-wide test verdict belongs to CI on the landing branch; at report time that
  verdict is **PENDING** (lane-local affected-package verification only, per §4 lane-local
  discipline). No push performed by this run (dispatch: the lane lands).
- The 87.9% coverage figure is the whole-package measure; the two changed files are inside
  `internal/graph` and covered by the run, but no per-file coverage split was measured.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-07T01:23:36Z
run_commit_sha: pending-backfill-run
run_status: complete
ac_pass_count: 11
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a (lane-local run; landing/push owned by the lane, not this agent)
l44_post_push_fetch: n/a (no push performed by this run, per dispatch)
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin: pass
cross_platform_build.windows_amd64: pass
total_run_phase_files: 2
m1_to_mN_commit_strategy: M1 test-only commit (f6476244f, carries the sanctioned draft->in-progress frontmatter flip) precedes M2 fix commit (0d32f5733); evidence commit follows; RED->GREEN ordering witnessed by the commit graph
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

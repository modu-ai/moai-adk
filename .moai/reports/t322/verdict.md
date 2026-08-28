# t322 — sync-phase verdict (SPEC-GRAPH-FRESHNESS-CADENCE-001)

Card: t322 · Lane: lane-3 · Branch: `WT-freshness-sync` · Worktree: `.claude/worktrees/t322`
Base: `origin/develop` @ `d566ecc75` · Measured: 2026-08-28

---

## Claim

The sync-phase for SPEC-GRAPH-FRESHNESS-CADENCE-001 is closed on `WT-freshness-sync`: CHANGELOG
entry added, `docs-site/content/{ko,en,ja,zh}/cli-reference/graph.md` updated to describe the
failure-attribution surface M3 delivered, `progress.md` §E.3 backfilled and §E.4 authored, and
`spec.md` transitioned `in-progress → implemented → completed` on the sync commit.

Two claims made by the sync agent were **not** supported by the tree and were corrected before this
verdict was written; they are listed under Gaps rather than quietly repaired.

The card's premise as dispatched was also wrong in the lane's favour: plan **and run** had already
landed. Only sync remained.

## Evidence

### Card state at entry — run-phase already landed

```
$ git log origin/develop --oneline --grep='t322' | head -3
44095ddc2 merge(WT-graph-freshness-cadence): SPEC-GRAPH-FRESHNESS-CADENCE-001 M1-M3 — described-worthy predicate, threshold confirmation, failure attribution (t322)
1b61c7cf4 docs(SPEC-GRAPH-FRESHNESS-CADENCE-001): refresh the source citations at run-phase close (t322)
8b11bbba1 feat(SPEC-GRAPH-FRESHNESS-CADENCE-001): M3 failure attribution — own contribution and driving paths (t322)

$ git show origin/develop:.moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/spec.md | grep -m1 '^status:'
status: in-progress
```

13 commits, `c28edd9de`..`1b61c7cf4`. The merge's own diff vs its first parent is 17 files —
`internal/{mx,graph,cli}` sources and tests, the 4 SPEC artifacts, and 3 `.moai/reports/t322/`
records. No template-tree file, hook, or `.claude/` rule is in it, so no Template-First mirror is
owed by this SPEC.

### Lane baseline (this worktree, before and after the sync edits)

```
$ go build ./...                                            → exit 0
$ go test ./internal/graph/... ./internal/mx/...
ok  github.com/modu-ai/moai-adk/internal/graph         14.126s
ok  github.com/modu-ai/moai-adk/internal/graph/symbol   2.122s
ok  github.com/modu-ai/moai-adk/internal/mx             6.123s
$ go test ./internal/cli/ -run 'TestGraphCheck' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli            4.143s
```

The repository-wide suite was NOT run locally (`AGENTS.md` §4 / `CLAUDE.local.md` §4). CI on the
integration branch owns that verdict.

### Docs claims checked against the source, not against the agent's report

The docs-site paragraph names four JSON fields, a display bound, and an omission format. Each was
verified to exist:

```
$ grep -n 'contribution\|driving_paths' internal/graph/check.go
 79:	Contribution *int `json:"contribution,omitempty"`
 82:	ContributionBase string `json:"contribution_base,omitempty"`
 90:	DrivingPaths []string `json:"driving_paths,omitempty"`
 93:	DrivingPathsOmitted int `json:"driving_paths_omitted,omitempty"`
 99: const drivingPathDisplayBound = 10
$ grep -n 'and %d more' internal/cli/graph_check.go
163:		_, _ = fmt.Fprintf(errs, "    ... and %d more (listing bounded)\n", l.DrivingPathsOmitted)
```

All four locales received the same paragraph at the same insertion point (`## moai graph check`),
2 inserted lines each.

### §E.3 — what is a backfill and what is a Gap, item by item

The lead asked for this distinction item by item, because it is what the sync audit reads.

| §E.3 item | Provenance | Traced by |
|---|---|---|
| `run_status: audit-ready` | Backfill from §E.2 | §E.2's own run-phase close |
| M1 SHA `5d95a2e8d…` | Backfill from §E.2 | `grep -c 5d95a2e8d` → 5 |
| M2 SHA `988adaf98…` | Backfill from §E.2 | `grep -c 988adaf98` → 5 |
| M3 SHA `8b11bbba1…` | Backfill from §E.2 | `grep -c 8b11bbba1` → 2 |
| AC matrix 12/12 live PASS | Backfill from §E.2 | `grep -c 'AC matrix\|12/12\|12 PASS'` → 3 |
| `go test -cover` figures (16.756s / 88.9% / 365.242s) | Backfill from §E.2 | `grep -c` → 1, 2, 1 respectively |
| `go vet` / `GOOS=windows` / `golangci-lint` / `make build` exit codes | Backfill from §E.2 | present in §E.2's verification batch |
| **`run_commit_sha: 44095ddc2…`** | **NOT a backfill — sync-phase git observation** | `grep -c 44095ddc2` against the §E.2 base → **0** |
| e2e coverage | **Gap** — not run at run-phase, not run at sync | stated in §E.3's own Gaps line |
| full local suite | **Gap** — forbidden by contract; CI owns it | stated in §E.3's own Gaps line |

The merge SHA cannot be an §E.2 carry by construction: the merge commit did not exist when the
run-phase record was written. §E.3 originally claimed "every figure below is a verbatim carry from
§E.2"; that sentence was corrected in `f9c827217` so the one exception names itself.

### graph-freshness on develop — green, but not because of this card

```
$ gh run list --branch develop --workflow "Graph Freshness" --limit 15
d566ecc75 attempt=1 success        8da086fbd attempt=1 failure
8d271af53 attempt=1 success        9a1831efd attempt=1 failure
ec15ec2cd attempt=1 success        f5a834fef failure
44095ddc2 attempt=1 success        da03d9188 failure
4fdbd55c1 attempt=1 success        d34a789a4 failure
8806a8788 attempt=1 success        0c7457f8d failure
                                   812ee01fc failure
```

Six consecutive greens, all `attempt=1` — no re-run is hiding an attempt-1 failure.

**The green does not belong to t322.** The streak starts at `8806a8788`, one landing BEFORE t322's
merge, and that commit reads:

```
$ git log --oneline -1 8806a8788
8806a8788 merge(WT-codex-init): integrate card t340 — moai init codex wiring gate (SPEC-CODEX-INIT-001) + codemaps regeneration for the launcher and init surfaces (closes t311)
```

— a codemaps regeneration, i.e. another manual restamp: the fourth, after t197, t228 and t340's own
earlier one. t322 landed into an already-green state and was never exercised against an inherited
red. The card's effect is observable only at the next inheritance event, when the cumulative
approaches the threshold again. Reading this streak as the card working is exactly the
restamp-clears-red-not-inaccuracy error the SPEC itself documents.

### develop CI failure attribution — full count, four groups

Run `33128899299`, head `d566ecc75`, `attempt=1`. 12 jobs, 2 failed: `Test (ubuntu-latest)` and
`Race Test`. 15 distinct top-level failing tests:

| Group | Count | Tests |
|---|---|---|
| t346 (doctor) — **name/symptom attribution, source NOT inspected** | 9 | `TestRunDoctor_{AllFlags,ExportMode,Verbose,VerboseAndDetail,WithExport,WithFix}`, `TestDoctorCmd_{Execution,ExportFlag,VerboseExecution}` |
| t340 / SPEC-CODEX-INIT-001 | 3 | `TestCodexInitAcceptDelegation`, `TestCodexInitGateInjectedState`, `TestCodexInitGateStateMatrix` (all `spawn=true` sub-cases) |
| unattributed | 2 | `TestConcurrencyStress`, `TestSessionStart_BlockingComparerDoesNotStallSessionStart` |
| **t322's own** | **1** | `TestGitDiffNameCount_Predicate` |

The t346 group all fail with the same message, which is the basis of that attribution:

```
coverage_improvement_test.go:777: runDoctor error: doctor: 1 check(s) failed
```

This is a name-and-symptom attribution. The t346 branch's sources were not opened, so it is a
hypothesis about ownership, not a verified one.

### t322's own CI failure — evidence preserved for a possible flake card

`TestGitDiffNameCount_Predicate` is AC-GFC-002's test, added by M1. It is not an assertion failure;
it is a `t.TempDir` cleanup race:

```
Test (ubuntu-latest):
  --- FAIL: TestGitDiffNameCount_Predicate (0.07s)
      testing.go:1464: TempDir RemoveAll cleanup: unlinkat /tmp/TestGitDiffNameCount_Predicate2280431520/001/.git/objects: directory not empty
  FAIL	github.com/modu-ai/moai-adk/internal/graph	1.409s

Race Test:
  --- FAIL: TestGitDiffNameCount_Predicate (0.30s)
      testing.go:1464: TempDir RemoveAll cleanup: unlinkat /tmp/TestGitDiffNameCount_Predicate2016181175/001/.git/objects/pack: directory not empty
  FAIL	github.com/modu-ai/moai-adk/internal/graph	4.537s
```

Frequency since the landing — 1 of 4 CI runs:

```
$ gh run view 33128899299 | grep -c 'FAIL: TestGitDiffNameCount_Predicate'  → present (d566ecc75)
$ gh run view 33126888247 ...                                              → 0 (8d271af53)
$ gh run view 33124415456 ...                                              → 0 (ec15ec2cd)
$ gh run view 33117635467 ...                                              → 0 (4fdbd55c1)
```

Local reproduction attempt — **failed to reproduce**:

```
$ go test ./internal/graph/ -run 'TestGitDiffNameCount_Predicate' -count=20 -timeout 300s
ok  	github.com/modu-ai/moai-adk/internal/graph	11.287s
```

20 iterations on macOS, no failure. The fixture (`newCheckFixture`, `internal/graph/check_test.go`)
builds a real git repository inside `t.TempDir()`, and something writes into `.git/objects` after
the test body returns. A detached git background child is the obvious suspect, but it is a
hypothesis: nothing here measured it.

**No fix was attempted.** A repair that cannot be reproduced cannot be verified, and an unverifiable
repair landing inside a close commit would make the close harder to read, not easier. The evidence
above is left intact so a flake card can cite it directly. Card issuance is the operator's; the lead
is collecting it.

## Baseline-attribution

Every figure above was measured in this run, against this tree:

- Git and SPEC state: `origin/develop` @ `d566ecc75`, re-fetched at lane entry; worktree HEAD equal
  to it at entry (`git rev-parse --short HEAD` → `d566ecc75`, clean tree).
- Build and tests: run in `.claude/worktrees/t322`, commands and output quoted verbatim above.
- CI figures: `gh run list` / `gh run view` against `modu-ai/moai-adk`, run ids quoted inline,
  `attempt` read explicitly rather than inferred from the listing.
- §E.2 traces: `git show origin/develop:.moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/progress.md`
  (1111 lines) as the §E.2 base, `grep -c` counts as quoted.

## Gaps

- **Two sync-signal claims were false and were corrected, not carried.** §E.4 stated the frontmatter
  transition covered "all four SPEC artifacts" and named `plan.md`/`acceptance.md`; only `spec.md`
  carries a `status:` key (checked against sibling SPEC-WORKTREE-BASEREF-001, whose four artifacts
  return the `spec.md` row alone). §E.4 also claimed an `updated:` refresh that did not occur — the
  field already read `2026-08-28`. Corrected in `f9c827217`.
- **The docs-site paragraph was not re-derived from a live CLI run at sync.** Its content is a carry
  of AC-GFC-009/010's verbatim fixture output recorded at run-phase; sync verified the field names
  and the display bound against the source, but did not rebuild a stale fixture and re-read stderr.
- **t346 attribution is by name and symptom only.** No t346 source was read.
- **`TestConcurrencyStress` and `TestSessionStart_BlockingComparerDoesNotStallSessionStart` are
  unattributed.** They were counted, not diagnosed.
- **The flake's cause is unestablished.** 20 local iterations did not reproduce it; the background-git
  hypothesis is untested.
- **The repository-wide suite was not run locally.** CI on the integration branch owns it, and at the
  time of writing the post-sync CI run for this branch has not been observed.
- **No sync-auditor pass was run on this close.** This verdict is the lane's own reading.

## Residual risk

- **The card's central claim is still unexercised.** Both the predicate and the attribution field
  landed green into an already-restamped tree. Nothing here demonstrates the predicate stops a real
  inherited red; the first genuine test is the next integration that would have crossed the
  threshold on `testdata`/template-YAML churn alone.
- **The threshold was retained on a cadence measurement whose window is short.** 40 is held as a
  control variable, not shown to be the right value — the streak's corrected cumulative was 2, so
  the threshold was not load-bearing in the observed data and remains unvalidated at its own scale.
- **The docs-site prose can drift silently.** No CI guard ties that page to `internal/graph/check.go`
  literal output, so a later rename of a JSON field or a change to the bound of 10 leaves the four
  locales quietly wrong.
- **t322's own flaky test remains on `develop`.** It reds the CI job roughly 1 run in 4 and, until it
  is fixed, will keep producing failures attributable to this card on integrations that have nothing
  to do with it — the same inherited-vs-originated confusion this card exists to end, reproduced one
  layer down.

## Commits on `WT-freshness-sync`

| SHA | Subject |
|---|---|
| `bc66c30b7` | sync-phase close — CHANGELOG, docs-site 4 locales, §E.3 backfill, §E.4, `spec.md` close |
| `28f16d030` | `sync_commit_sha` backfill |
| `f9c827217` | correct two overstated claims in the sync signal |

Files changed vs `origin/develop`: `CHANGELOG.md`,
`.moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/{spec.md,progress.md}`,
`docs-site/content/{ko,en,ja,zh}/cli-reference/graph.md`, and this verdict.

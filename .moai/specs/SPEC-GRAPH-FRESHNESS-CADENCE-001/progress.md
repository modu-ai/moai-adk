# SPEC-GRAPH-FRESHNESS-CADENCE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored for card t322 in worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t322`, branch `WT-graph-freshness-cadence`, base
`d2cba5e21`.

- Tier: **M** — artifact set `spec.md` + `plan.md` + `acceptance.md` (+ this file). **11 live
  requirements** (13 numbered, REQ-GFC-004 and 012 withdrawn at v0.2.0; ceiling 16) and **12 live
  acceptance criteria** (14 numbered, AC-GFC-011 and 012 withdrawn with them; ceiling 16).
- Baseline: every figure in `spec.md` §B re-measured in this tree during authoring. The dispatch's
  figures (55 / 0 / 5 per integration; 21 / 198 calibration re-run) reproduced exactly.
- Judgments (as of v0.2.0): `spec.md` §D.1 (a — yes, predicate in the metric), §D.2 (b — **40
  retained**, re-justified on the integration axis; the v0.1.0 proposal of 15 was reversed by audit
  D2), §D.3 (c — no pipeline refresh; attribution instead). Rejected alternatives recorded per
  judgment.

## §E.1b Plan-phase Audit Record

- Audit iter-1 (`plan-auditor`): **PASS-WITH-DEBT 0.82** (Tier M PASS threshold 0.80), MP-1..MP-4
  all pass, four blocking defects (D1 critical, D2/D3/D4 major) plus D5-D8 optional. Verdict:
  `.moai/reports/t322/plan-audit.md`.
- Remediation landed at v0.2.0; the per-defect account is the `spec.md` HISTORY row for that
  version. D1 was additionally confirmed by the orchestrator directly against source
  (`internal/mx/provenance.go:107-143` applies no extension filter; `internal/graph/meta.go:67`
  routes three non-Go directories through it), so it is an established finding rather than an
  auditor hypothesis.
- D2 reversed the threshold judgment: the v0.1.0 value 15 was derived on the per-integration axis
  while the metric is cumulative-since-stamp; on the cumulative axis 15 reds the gate roughly twice
  per day at the observed integration rate, and the streak this SPEC exists to fix carries a
  corrected cumulative of 2 — under both 15 and 40. The threshold change was therefore not
  load-bearing for the defect and is now out of scope.
- Audit iter-2 (`plan-auditor`, the Tier M ceiling — there is no iter-3): **PASS-WITH-DEBT 0.895**,
  monotonic +0.075, MP-1..MP-7 all pass, all four iter-1 blocking defects confirmed closed against
  source. Three new blocking defects (N1 major — the §D.2 cadence figure rested on the
  self-referential `6786c3fa4`; N2, N3 minor) plus five optional. Verdict:
  `.moai/reports/t322/plan-audit-iter2.md`.
- Remediation of N1/N2/N3 and optional O2/O4 landed at **v0.2.1**, applied directly rather than as a
  plan-phase re-entry (the iteration ceiling is reached). The cadence walk was re-measured in this
  tree both ways — 40-crossing at integration 10 with the outlier, **16** without it — and the
  outlier-excluded figure is now the operative one in §B.6, §D.2 and §G Q4, carrying an explicit
  disclaimer there that it is not the reason 40 is retained.
- Two consumers examined at v0.2.1 and given explicit dispositions in `spec.md` §D.1, so neither
  reads as an oversight to a later reader: the unpaired `StampMXScan` / `StampEdges` producers are
  **safe on a stated premise** (`check.go:187` is the only non-test `ContentFingerprint` comparator —
  re-measured at this HEAD), and `treeDirty` (`provenance.go:201`) is **examined and deferred** with
  its residual recorded (a tree dirty only under `testdata` is still refused the `--commit`
  merge-base anchor). The `treeDirty` deferral is a candidate for a follow-up card — operator's call.
- Ordering basis for t322 / t311 / t304: `spec.md` §F.
- Open questions deliberately left to the operator: `spec.md` §G (**four** — Q4 added at v0.2.0 and
  restated on the outlier-excluded cadence figure at v0.2.1).
- Evidence path: `.moai/reports/t322/`.

## §E.2 Run-phase Evidence

### M1 — Described-worthy predicate

Implemented in worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t322`, branch
`WT-graph-freshness-cadence`, base HEAD `4a50d44f4`. Cycle type TDD; every criterion below was
observed RED before it was observed GREEN, and the three mutants `plan.md` §E M1 names were each
planted, observed to fail their criterion, and reverted.

Pre-flight, re-measured in this tree at `4a50d44f4` (not carried from the dispatch):

```
$ go build ./internal/graph/... ./internal/mx/... ./internal/config/...
→ exit 0
$ ./bin/moai graph check
→ codemaps metric=described-source-diff value=5 threshold=40 verdict=fresh
```

The baseline `moai graph check` was run after `make build`, because this worktree carried no
`bin/moai` at entry.

#### Delivered surface

| Site | Change |
|---|---|
| `internal/mx/described_worthy.go` (new) | `IsDescribedWorthy(relPath string) bool` — the single exported predicate. `.go` suffix, not `_test.go`, no path segment equal to `testdata`. Pure: no configuration parameter (REQ-GFC-004 withdrawn), no filesystem access. |
| `internal/graph/check.go` `gitDiffNameCount` | Predicate applied to **both** branches — the `git diff --name-only` branch and the `git ls-files --others --exclude-standard` branch. |
| `internal/mx/provenance.go` | `AggregateDescribedFingerprintFiltered` added; `aggregateFingerprintPred(projectRoot, roots, admit)` carries the walk, with the predicate applied **after** the existing regular-file guard. |
| `internal/graph/check.go` `checkCodemaps` dirty branch | Consumer switched to `AggregateDescribedFingerprintFiltered`. |
| `internal/mx/provenance.go` `baseProvenance` | Gains a `fingerprint fingerprintFunc` parameter; `StampCodemaps` passes the filtered aggregate, `StampMXScan` and `StampEdges` pass the unfiltered one. |
| `internal/mx/provenance.go` `AggregateDescribedFingerprint` / `aggregateFingerprint` | **Unchanged behaviour** — `aggregateFingerprint` delegates with a `nil` predicate, so `meta.go:67 dirFingerprint` and the edges layer are untouched (REQ-GFC-003a). |

**Producer/consumer mechanism chosen: the `baseProvenance` fingerprint-function parameter**, the
first of the two `plan.md` §E M1 admits. Chosen over recomputing inside `StampCodemaps` for two
reasons: it makes the pairing visible at the call site (each stamp writer names the fingerprint its
consumer will recompute, so a future writer cannot inherit the wrong one by omission), and it
computes one fingerprint rather than computing the unfiltered one and discarding it — the recompute
alternative walks a dirty described tree twice. The cost is the three mechanical call-site updates
the parameter forces, which is the smaller price.

#### AC matrix

| AC | Status | Deciding command | Observed |
|---|---|---|---|
| AC-GFC-001 | PASS | `go test ./internal/mx/ -run TestDescribedWorthy -count=1` | `ok github.com/modu-ai/moai-adk/internal/mx 0.336s` |
| AC-GFC-002 | PASS | `go test ./internal/graph/ -run TestGitDiffNameCount_Predicate -count=1` | `ok github.com/modu-ai/moai-adk/internal/graph 0.852s` |
| AC-GFC-003 | PASS | `./bin/moai graph check --root <fixture> --json` → codemaps `.value` | `2` |
| AC-GFC-004 | PASS | `go test ./internal/mx/ -run TestCodemapsFingerprint_ProducerConsumer -count=1` | `ok github.com/modu-ai/moai-adk/internal/mx 0.471s` |
| AC-GFC-005 | PASS | `go test ./internal/graph/ -run TestCheckCodemaps_Absent -count=1 -v` | 7/7 sub-tests PASS, each pinned to its pre-existing reason string; branch 7 pinned as the `(absent, err != nil)` pair |
| AC-GFC-014 | PASS | `go test ./internal/graph/ -run TestSourceFingerprintsForEdges_Unchanged -count=1` | `ok github.com/modu-ai/moai-adk/internal/graph 0.619s` |

RED evidence, captured before each implementation:

```
AC-GFC-001  internal/mx/described_worthy_test.go:33:13: undefined: IsDescribedWorthy
            FAIL github.com/modu-ai/moai-adk/internal/mx [build failed]
AC-GFC-002  --- FAIL: TestGitDiffNameCount_Predicate (0.52s)
            check_predicate_test.go:55: gitDiffNameCount = 41, want 1
AC-GFC-004  internal/mx/codemaps_fingerprint_test.go:58:14: undefined: AggregateDescribedFingerprintFiltered
            FAIL github.com/modu-ai/moai-adk/internal/mx [build failed]
```

AC-GFC-005 and AC-GFC-014 are preservation criteria: their RED is the mutant, not the
pre-implementation tree. AC-GFC-014's is recorded under mutant (ii) below; AC-GFC-005 pins behaviour
this milestone must not move, and no mutant of it is claimed.

#### Mutant runs (RED under the mutant → GREEN after revert)

**(i) predicate on the `git diff` branch only** — the predicate removed from the
`git ls-files --others` branch:

```
--- FAIL: TestGitDiffNameCount_Predicate (0.52s)
    check_predicate_test.go:55: gitDiffNameCount = 2, want 1
      (41 = no predicate; 2 = predicate on the git diff branch only)
```

The count landed on exactly the 2 the criterion predicts: the one production `.go` change plus the
untracked fixture the unfiltered branch admits. After revert: `ok ... 0.852s`.

**(ii) predicate pushed down into `aggregateFingerprint`** — `aggregateFingerprintPred` called with
`IsDescribedWorthy` instead of `nil`:

```
--- FAIL: TestSourceFingerprintsForEdges_Unchanged (0.29s)
    check_predicate_test.go:79: source set "codemaps" collapsed to the empty-entry hash
        — the predicate reached aggregateFingerprint
    check_predicate_test.go:79: source set "specs" collapsed to the empty-entry hash
        — the predicate reached aggregateFingerprint
    check_predicate_test.go:88: codemaps source fingerprint did not move on a non-.go edit
        — the edges layer is permanently green
```

Both source sets collapsed to `e3b0c442…`, exactly the constant `spec.md` §D.1 measured, and the
layer went permanently green. After revert: `ok ... 0.619s`.

**(iii) filtered checker with an unfiltered codemaps stamp writer** — `StampCodemaps` passing
`AggregateDescribedFingerprint`:

```
--- FAIL: TestCodemapsFingerprint_ProducerConsumer (0.32s)
    codemaps_fingerprint_test.go:67: stale against its own fresh stamp
        — the producer and the consumer disagree
```

The first assertion failed, as the criterion predicts. After revert: `ok ... 0.471s`.

#### Sole-comparator enumeration (`spec.md` §D.1 premise, re-measured)

Delivered tree:

```
$ grep -rn --include='*.go' 'ContentFingerprint' . | grep -v _test.go
internal/graph/check.go:190:		if cur == pv.ContentFingerprint {
internal/graph/check.go:197:		rep.Reason = fmt.Sprintf("content moved past dirty-generation fingerprint %s", shortHash(pv.ContentFingerprint))
internal/mx/provenance.go:46:	// ContentFingerprint, never a named commit the generation did not see.
internal/mx/provenance.go:49:	// ContentFingerprint is the aggregate sha256 of the described content at
internal/mx/provenance.go:52:	ContentFingerprint string `json:"content_fingerprint,omitempty"`
internal/mx/provenance.go:237:// dirty branch's ContentFingerprint is computed — the codemaps writer passes
internal/mx/provenance.go:252:			pv.ContentFingerprint = fp
internal/mx/provenance.go:312:		return fmt.Sprintf("provenance: tree=%s dirty fingerprint=%s", p.TreeRoot, shortHash(p.ContentFingerprint))
```

Exactly **one comparator** (`check.go:190`, the equality test) and **one display-only reader**
(`provenance.go:312`, `Provenance.Describe`). The remaining hits are the struct field, its doc
comments, the producer write at `provenance.go:252`, and the stale-verdict reason formatter at
`check.go:197` — none of them compares.

The plan-phase baseline cites `check.go:187` and `provenance.go:280`. Both reproduce verbatim at the
base commit, so the delivered line numbers are this milestone's own insertions shifting the same two
sites, not new sites:

```
$ git show 4a50d44f4:internal/graph/check.go | grep -n 'ContentFingerprint'
187:		if cur == pv.ContentFingerprint {
194:		rep.Reason = fmt.Sprintf(...)
$ git show 4a50d44f4:internal/mx/provenance.go | grep -n 'ContentFingerprint' | tail -1
280:		return fmt.Sprintf("provenance: tree=%s dirty fingerprint=%s", ...)
```

No second comparator is present, so `plan.md` §B's stop-for-judgment condition did not fire and the
unpaired `StampMXScan` / `StampEdges` producers remain safe on the stated premise.

#### AC-GFC-003 fixture and baseline attribution

The fixture is a local clone checked out detached at `48eb945df`, which already carries the codemaps
provenance stamped at `9326b5478d0f51979dfb498527458dcea5e0370b` (`commit_sha` read from the
fixture's own `provenance.json`) — no stamp edit was made, and no `moai graph stamp` was run
anywhere.

```
$ ./bin/moai graph check --root <fixture> --json
codemaps: {"metric":"described-source-diff","value":2,"threshold":40,"verdict":"fresh"}
```

Baseline-attribution row (what the metric did before the change, on the same window):

```
$ git -C <fixture> diff --name-only 9326b5478d… 48eb945df -- internal cmd pkg | grep -c .
65
$ … | grep '\.go$' | grep -v '_test\.go$' | grep -v '/testdata/' | grep -c .
2
```

**65 → 2**, reproducing `spec.md` §B.2's counterfactual through the delivered code rather than
through a shell filter chain.

#### Independent verification batch (this run, this tree, base `4a50d44f4`)

```
$ go test ./internal/graph/... ./internal/mx/... -count=1
ok  github.com/modu-ai/moai-adk/internal/graph         10.885s
ok  github.com/modu-ai/moai-adk/internal/graph/symbol   0.402s
ok  github.com/modu-ai/moai-adk/internal/mx             5.358s
$ go vet ./internal/graph/... ./internal/mx/...                            → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...                                 → exit 0
$ GOOS=windows GOARCH=amd64 go vet ./internal/graph/... ./internal/mx/...  → exit 0
$ go test -cover ./internal/graph/ ./internal/mx/ -count=1
coverage: 87.9% of statements (graph) · 88.9% of statements (mx)
$ golangci-lint run --timeout=5m ./internal/graph/... ./internal/mx/...
0 issues.
$ go test ./internal/hook/quality/... -count=1                             → ok (11.669s)
$ go test ./internal/cli/ -run Graph -count=1                              → ok (8.204s)
```

The full local suite was NOT run (`plan.md` §D): CI renders the full verdict. `e2e/` was not run.

Constraint checks:

```
$ grep -rn "graph stamp" .github/ .claude/                                 → no output
$ git diff --stat 4a50d44f4 -- .moai/project/codemaps/provenance.json      → empty
```

The threshold is untouched at 40 (`DefaultThresholds`), as `plan.md` §D requires — M2 owns any
change to it.

Evidence files: `.moai/state/verify/t322-m1/` (`tests.txt`, `lint.txt`, `ac003-graph-check.json`).

#### Gaps — what was NOT observed

- **The full local test suite was not run.** Only `internal/graph/...`, `internal/mx/...`,
  `internal/hook/quality/...`, and the `Graph`-filtered subset of `internal/cli` were executed. Any
  regression in a package outside those is unobserved here and is CI's verdict to render.
- **`e2e/` was not run** (the operator has a live console; that suite mutates real profiles).
- **The AC-GFC-014 "before and after" comparison was not run as a literal two-tree diff.** The
  criterion is decided by the inequality against the empty-entry hash plus a behavioural assertion
  that a non-`.go` edit still moves the codemaps source fingerprint, both on the delivered tree, and
  by the mutant run above. Byte-identity of every edges fingerprint against a pre-change tree
  computed over the same fixture was not separately captured.
- **No cross-platform run of the test suite.** Windows was verified by `go build` and `go vet` only
  (compilation, not execution).
- **M2 and M3 were not started.** No threshold re-measurement, no per-change attribution, no
  driving-path stderr output, no `--json` attribution fields.
- **The tree's own `graph check` was not re-run after the change** to record what the corrected
  metric now reports for this worktree; the pre-flight figure (5, fresh) is the pre-change one.

#### Residual risk

- **Existing dirty codemaps stamps are invalidated by this change** — a tree carrying a dirty
  `ContentFingerprint` written before this commit will read stale once, until it is re-stamped by the
  ordinary regeneration path. `plan.md` §B accepts this as transient by construction; it is named
  here because the first operator to hit it will see a stale codemaps verdict with no source change
  behind it.
- **`treeDirty` still consults no predicate** (`spec.md` §D.1, examined and deferred): a tree dirty
  only under `testdata` is still refused the `--commit` anchor. Unchanged by this milestone, by
  design.
- **The mx-scan and graph-build stamps now write a fingerprint computed differently from the
  codemaps stamp.** Safe only while nothing compares them — verified above at this HEAD, and silent
  if that changes. The obligation is recorded in `spec.md` §D.1.
- **The predicate is a judgment, not a measurement.** It admits every non-test `.go` file outside
  `testdata`, including generated Go and Go under `internal/template/templates` (of which there are
  currently none). If a generated-Go tree lands under a described root, the metric will count it.

#### Coordinate drift caused by M1 — refresh deferred to run-phase close

M1 inserted lines into `internal/graph/check.go` and `internal/mx/provenance.go`, which moved source
coordinates that `spec.md` and `plan.md` cite by line number. Those two artifacts still carry the
pre-M1 values; this section (`§E.2`) carries the post-M1 ones, so the artifacts currently disagree
with each other. The disagreement is recorded here rather than repaired, because M3 adds attribution
output to `check.go` and will move several of these lines again — one refresh after M3 costs less
than two, and a refresh done now would be measured against a tree that no longer exists by the time
anyone reads it.

It is written down rather than left silent because this SPEC already failed plan audit iter-2 on
this exact defect class — N2, `provenance.go:196` cited while the function was at `:208`. The
difference is only that N2 was wrong and unnoticed; this is wrong and recorded.

Both columns are measured, not inferred: the pre-M1 column is
`git show 4a50d44f4:<file> | grep -n …` and the post-M1 column is `grep -n …` on the delivered tree.
Every pre-M1 coordinate reproduced exactly at `4a50d44f4`, so each row is a move of the same
construct, not a new site.

Each row carries the **content-anchored grep that decides it**. Anchoring on the construct rather
than on the number is what lets a later reader re-derive the coordinate after M3 moves these lines
again, instead of trusting a number this record froze.

| Cited coordinate | Construct cited | pre-M1 (`4a50d44f4`) | post-M1 (`5d95a2e8d`) | Deciding grep | What the old line holds now |
|---|---|---|---|---|---|
| `check.go:181` | the codemaps dirty-branch recompute | 181 | **184** | `grep -n 'AggregateDescribedFingerprintFiltered(projectRoot, roots)' internal/graph/check.go` | first line of the M1 comment above the recompute |
| `check.go:187` | the sole `ContentFingerprint` comparator | 187 | **190** | `grep -n 'cur == pv.ContentFingerprint' internal/graph/check.go` | `rep.Reason = "described roots unreadable: "…` |
| `provenance.go:107` | `func aggregateFingerprint` (cited in `§E.1`) | 107 | **121** | `grep -n 'func aggregateFingerprint(' internal/mx/provenance.go` | doc comment of `AggregateDescribedFingerprintFiltered` |
| `provenance.go:201` | `func treeDirty` | 201 | **225** | `grep -n 'func treeDirty' internal/mx/provenance.go` | a `ResolveCommit` doc-comment line |
| `provenance.go:208` | `func baseProvenance` (declaration) | 208 | **240** | `grep -n 'func baseProvenance' internal/mx/provenance.go` | `func ResolveCommit` |
| `provenance.go:219` | the `aggregateFingerprint` call inside `baseProvenance` | 219 | **251** | `grep -n 'fingerprint(projectRoot, describedRoots)' internal/mx/provenance.go` | a closing brace |
| `provenance.go:237,253,264` | the three stamp writers' `baseProvenance` calls | 237 / 253 / 264 | **269 / 285 / 296** | `grep -n 'pv := baseProvenance(projectRoot,' internal/mx/provenance.go` | 237 is now a doc-comment line |
| `provenance.go:280` | `Provenance.Describe`, the display-only reader | 280 | **312** | `grep -n 'dirty fingerprint=' internal/mx/provenance.go` | a `StampMXScan` comment line |

Two of these changed shape as well as position, and a refresh must not copy only the number:
`baseProvenance` gained a `fingerprint fingerprintFunc` parameter, and the call at the old `:219` is
now `fingerprint(projectRoot, describedRoots)` rather than a direct `aggregateFingerprint` call.

Cited coordinates M1 did **not** move, confirmed unmoved rather than assumed:

| Coordinate | Construct | Basis |
|---|---|---|
| `check.go:168` | `roots = mx.DefaultDescribedRoots` | still line 168 — M1's insertions are all below it |
| `meta.go:67` | `dirFingerprint`'s aggregate call | `git diff --stat 4a50d44f4 -- internal/graph/meta.go` → empty |
| `symbol.go:99` | `symbol.go`'s `DefaultDescribedRoots` consumer | `git diff --stat 4a50d44f4 -- internal/graph/symbol/symbol.go` → empty |

One citation is **excluded from the refresh by construction**: `provenance.go:196` in `spec.md`'s
v0.2.1 HISTORY row is a quotation of the wrong citation audit finding N2 corrected. It is a record of
an error, not a pointer into the tree, and refreshing it would erase the correction it documents.

`acceptance.md` carries no line-number citation at all (`grep -no '[A-Za-z0-9_]*\.go:[0-9]\+'` → no
output), so the refresh scope is `spec.md`, `plan.md`, and `progress.md` `§E.1`.

**Deciding command for the refresh (run at run-phase close, after M3).** Enumerate every citation in
the four artifacts, then resolve each against the delivered tree:

```bash
grep -rhno '[A-Za-z0-9_]*\.go:[0-9][0-9,]*' \
  .moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/{spec,plan,acceptance,progress}.md \
  | awk -F: '{print $2":"$3}' | sort -u          # count is tree-dependent — measure, do not quote
sed -n '<N>p' <path>                              # resolve each row
```

**The citation count is a property of the tree, not a constant — measure it at refresh time rather
than quoting a figure from here.** It was 26 at `5d95a2e8d` and 27 once M2 landed, which added
`check.go:53` (the `CodemapsChangedFiles: 40` citation). Every later milestone that cites code moves
it again, so a number frozen into this record would be false by the time the refresh runs. The
command above is the measurement; this paragraph is only the warning against skipping it.

A basename-keyed resolver is NOT sufficient here and must not be substituted: `find internal -name
provenance.go` returns three files (`internal/config`, `internal/mx`, `internal/navigator/sync`) and
`check.go` / `symbol.go` collide likewise, so each row resolves against the path its citing sentence
names — `check.go` → `internal/graph/check.go`, `meta.go` → `internal/graph/meta.go`, `symbol.go` →
`internal/graph/symbol/symbol.go`, `provenance.go` → `internal/mx/provenance.go`.

The refresh is complete when every resolved line is the construct its citing sentence names, with the
`:196` HISTORY quotation left untouched. The comma-list form (`237,253,264`) needs its trailing
members read by eye — the enumeration prints the row whole, but a resolver that splits on the first
number alone will silently skip 253 and 264.

Ownership: `spec.md` and `plan.md` bodies are manager-spec's, so this milestone records the drift and
does not repair it.

**Provenance of this section, recorded so the history reads correctly later.** For a period this
record existed twice: this section, and a near-duplicate `### §E.2b Citation drift introduced by M1`
placed after `## §E.4`. Both arrived in the same commit, `368bfae2f`, whose message names only the
second. The orchestrator, reading two dispatches as unstarted, wrote its own subsection and then
staged and committed `progress.md` in one compound command without re-reading `git status --short`
immediately beforehand, so this section — uncommitted in the same file at that moment — was swept in
under that message. Nothing was lost, and the duplicate has been folded into this section and
removed. Two things were carried over from it and are above: the per-row deciding grep, and the N2
framing. Its own refresh-deciding command was NOT carried over — measured on the same tree, its
regex `internal/[a-z]+/[a-z_]+\.go` returns 19 citations against this section's 27, missing
`internal/graph/symbol/symbol.go:99` (three path segments), the trailing members of the comma list
`237,253,264`, and every bare-form citation, which is the form `spec.md` and `plan.md` mostly use.
The originating discipline is `AGENTS.md` §2: stage by explicit pathspec and re-read status
immediately before staging, which binds even when no concurrent actor was expected.

It then happened a second time, from a different actor: `988adaf98`, the M2 commit, carries this
section's per-row deciding-grep column and its N2 paragraph — both uncommitted in this file when
that commit staged it — under an M2 message. Two sweeps by two actors on one file is the signal
worth keeping here: `progress.md` is a shared surface during run-phase, so an agent that stages it
whole absorbs whatever another agent has in flight. Nothing was lost either time, but only because
the overlap was noticed and reported rather than because the staging was safe.

### M2 — Threshold confirmation (REQ-GFC-005, REQ-GFC-006)

Measured in worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t322`, branch
`WT-graph-freshness-cadence`, at HEAD `bee0598fe`, on 2026-08-27. **Conclusion: 40 is confirmed and
retained. The measurement supports no correction.** The derivation follows; the record itself is the
deliverable REQ-GFC-005 requires, and it would have failed identically had the value been changed.

#### Which window, and why it is not `spec.md` §B's window

`spec.md` §B.5/§B.6 measured the 30-integration window ending at `48eb945df`, taken from this
branch's own HEAD during plan-phase authoring. That is no longer the right window for two reasons,
both measured rather than assumed:

1. **This branch's HEAD is no longer an integration series.** Eight of the last thirty first-parent
   commits reachable from HEAD are this card's own plan-phase and M1 commits
   (`git log --first-parent -30 --pretty=format:"%h %s" | grep -c 't322'` → 8). Walking them would
   measure the cadence of authoring this SPEC, not the cadence of factory integration — and M1 is
   itself a described-worthy contributor, so it would inflate its own threshold justification.
2. **`develop` has advanced past the branch base.** `develop` = `origin/develop` = `343399d2f`, and
   `d2cba5e21` (this branch's base) is behind it; four integrations landed after `48eb945df`
   (`52c693327`, `6310dbf28`, `3cb258d62`, and the tip's own first parent).

So both axes are measured on **`develop`**, the branch integrations actually land on, which is the
axis CI runs the gate on. The window is its last 30 first-parent commits touching the described
roots:

```
$ git log --first-parent -30 --reverse --pretty=format:"%h %ad" --date=short develop -- internal cmd pkg
5fd63ebcb 2026-08-25   … (oldest)
3cb258d62 2026-08-27   … (newest)
$ git log -1 --format='%h %ad' --date=iso-strict 5fd63ebcb → 2026-08-25T12:30:28+09:00
$ git log -1 --format='%h %ad' --date=iso-strict 3cb258d62 → 2026-08-27T22:37:03+09:00
```

Span **2.421 days**, 29 inter-arrival gaps across 30 integrations → **≈12.0 integrations/day**
(≈11.6/day with the outlier removed). Comparable to §B.6's ≈10/day, measured on a different window.

#### The predicate and the percentile convention, stated

The described-worthy predicate applied to both axes is the delivered one: path ends in `.go`, name
does not end in `_test.go`, and **no path segment equals `testdata`** (segment equality, not
substring — `internal/foo/testdatax/a.go` is admitted).

**Percentile convention: nearest-rank.** Rank = `ceil(p · n)` over the ascending sort, 1-indexed.
For n = 30, p90 is rank 27. This is the convention audit D5 named after v0.1.0 reported a p90 that
actually sat at rank 28 (the 93rd percentile).

Both axes are computed by one script over one captured log, so the two axes cannot silently diverge
on the predicate:

```
$ git log --first-parent -30 --reverse --name-only --pretty=format:"===%h" develop -- internal cmd pkg \
    > .moai/state/verify/t322-m2/axis-raw.txt          # 443 lines
$ python3 .moai/state/verify/t322-m2/axes.py .moai/state/verify/t322-m2/axis-raw.txt
```

The per-integration counts the script derives reproduce independently through an awk one-liner over
the same command, so the script is not the sole witness:

```
$ git log --first-parent -30 --reverse --name-only --pretty=format:"===%h" develop -- internal cmd pkg \
  | awk '/^===/{if(h!="")print h" "c; h=substr($0,4); c=0; delete seen; next}
         /\.go$/ && !/_test\.go$/ && $0 !~ /(^|\/)testdata\// {if(!seen[$0]++) c++}
         END{if(h!="")print h" "c}'
5fd63ebcb 1   db1362739 5   7f5b6a947 0   e91def4ca 0   a739d04b4 4   71781683c 0
6786c3fa4 29  c9eed8ac6 0   9a95e7a02 0   410da655f 2   da791eb0a 11  379b310a6 1
5ea3d6706 0   968ed2acb 8   26c5a7d54 8   a46862091 0   7b10c95ba 0   8ef14f5ae 3
6b20c0fe6 5   22df80e90 9   f25f2d348 2   242748906 11  cb1833f87 0   8d8da0b2b 5
7ed6edb3e 0   3809d1d36 2   48eb945df 0   52c693327 3   6310dbf28 0   3cb258d62 15
```

#### Axis 1 — per-integration described-worthy contribution

```
=== AXIS 1 - per-integration described-worthy contribution ===
n                      = 30
integrations contrib 0 = 12 of 30
median                 = 2.0
p90 (nearest-rank, rank=27) = 11
maximum                = 29  (6786c3fa4)
mean                   = 4.13
ascending              = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 2, 2, 3, 3, 4, 5, 5, 5, 8, 8, 9, 11, 11, 15, 29]
```

Against `spec.md` §B.5 (a different, earlier window): zero-contributors 12/30 → **12/30 unchanged**,
median 2 → **2 unchanged**, p90 9 → **11**, mean 3.9 → **4.13**, maximum 29 (`6786c3fa4`) →
**29 (`6786c3fa4`), the same integration**. The distribution shifted slightly upward and its shape
did not change: most integrations contribute nothing or a handful, and one self-referential outlier
dominates.

**Axis 1 alone cannot decide the threshold**, and this is the point audit D2 established. A p90 of
11 would argue for a threshold near 11 only if the metric were per-change. It is cumulative since
the stamp, so Axis 2 is the deciding axis and Axis 1 is context for it.

#### Axis 2 — cumulative-crossing cadence (whole window)

```
=== AXIS 2 - cumulative-crossing cadence [whole window] ===
window size = 30 integrations
   1  5fd63ebcb  union=1        11  da791eb0a  union=41     21  f25f2d348  union=64
   2  db1362739  union=6        12  379b310a6  union=42     22  242748906  union=71
   3  7f5b6a947  union=6        13  5ea3d6706  union=42     23  cb1833f87  union=71
   4  e91def4ca  union=6        14  968ed2acb  union=47     24  8d8da0b2b  union=75
   5  a739d04b4  union=10       15  26c5a7d54  union=54     25  7ed6edb3e  union=75
   6  71781683c  union=10       16  a46862091  union=54     26  3809d1d36  union=77
   7  6786c3fa4  union=39  <-- crosses 15                   27  48eb945df  union=77
   8  c9eed8ac6  union=39       17  7b10c95ba  union=54     28  52c693327  union=80
   9  9a95e7a02  union=39       18  8ef14f5ae  union=57     29  6310dbf28  union=80
  10  410da655f  union=41  <-- crosses 40                   30  3cb258d62  union=94
                                 19  6b20c0fe6  union=60
                                 20  22df80e90  union=64
threshold 15: first crossed at integration 7 (6786c3fa4, union 39)
threshold 40: first crossed at integration 10 (410da655f, union 41)
```

**This single walk is not admissible as the cadence figure**, exactly as `plan.md` §E M2 and
AC-GFC-006 require. The union goes `6 → 10 → 10 → 39` at integration 7: the jump from 10 to 39 is
`6786c3fa4` alone — the `SPEC-V3R6-GRAPH-FRESHNESS-001` delivery that *introduced this gate*,
contributing 29 described-worthy files. Both crossings above sit downstream of it, so both describe
how often the gate reds while the gate itself is being built.

#### Axis 2 — outlier sensitivity (`6786c3fa4` excluded)

The exclusion is justified in one sentence, and it is the same sentence §B.6 gives: `6786c3fa4` is
the commit that delivered this gate, so counting its own 29 files as evidence for the gate's red
rate measures the construction of the instrument rather than the process it observes.

```
=== AXIS 2 - cumulative-crossing cadence [outlier 6786c3fa4 excluded] ===
window size = 29 integrations
   1  5fd63ebcb  union=1        11  379b310a6  union=24     21  242748906  union=58
   2  db1362739  union=6        12  5ea3d6706  union=24     22  cb1833f87  union=58
   3  7f5b6a947  union=6        13  968ed2acb  union=32     23  8d8da0b2b  union=62
   4  e91def4ca  union=6        14  26c5a7d54  union=39     24  7ed6edb3e  union=62
   5  a739d04b4  union=10       15  a46862091  union=39     25  3809d1d36  union=64
   6  71781683c  union=10       16  7b10c95ba  union=39     26  48eb945df  union=64
   7  c9eed8ac6  union=10       17  8ef14f5ae  union=42  <-- crosses 40
   8  9a95e7a02  union=10       18  6b20c0fe6  union=45     27  52c693327  union=67
   9  410da655f  union=12       19  22df80e90  union=49     28  6310dbf28  union=67
  10  da791eb0a  union=23  <-- crosses 15                   29  3cb258d62  union=81
                                 20  f25f2d348  union=49
threshold 15: first crossed at integration 10 (da791eb0a, union 23)
threshold 40: first crossed at integration 17 (8ef14f5ae, union 42)
```

Side by side, with `spec.md` §B.6's plan-phase figures for comparison:

| Threshold | with outlier (now) | **without outlier (now)** | with outlier (§B.6) | without outlier (§B.6) |
|---|---|---|---|---|
| 15 | integration 7 | **integration 10** | integration 5 | integration 5 |
| **40** | integration 10 | **integration 17** | integration 10 | integration 16 |

The 40-crossing reproduces almost exactly: **16 → 17 integrations** on the operative
outlier-excluded walk, measured on a different window three days later. The 15-crossing does *not*
reproduce §B.6's claim that it is unaffected by the outlier — in this window the outlier lands at
integration 7, *before* the 15-crossing, so removing it moves that crossing from 7 to 10. §B.6's
"unaffected" was a property of its window, not of the metric; recorded here as a correction rather
than carried forward.

#### Expected red frequency, and the check against §D.2's stated intent

Stated on the **outlier-excluded** walk, as AC-GFC-006 requires:

```
$ python3 -c "…"   # 28 gaps / 2.421 days = 11.56 integrations/day
outlier-excluded, thresh 40 : 17 integrations / 11.56 per day = 1.47 days to red
outlier-excluded, thresh 15 : 10 integrations / 11.56 per day = 0.86 days to red
with outlier,     thresh 40 : 10 integrations / 11.98 per day = 0.83 days to red
with outlier,     thresh 15 :  7 integrations / 11.98 per day = 0.58 days to red
```

`spec.md` §D.2 states the intent as *"roughly once every day and a half of factory activity"*. The
re-measured figure is **1.47 days**. That is the stated intent, reproduced independently on a
different window — not a value tuned to it.

The comparison against 15 also survives, with its magnitude corrected. §D.2 argued 15 would be
roughly a **threefold** increase in red rate; measured here it is **1.47 / 0.86 ≈ 1.7×**, not 3×.
The direction of §D.2's argument holds — 15 reds noticeably more often, against a gate whose only
exit is manual regeneration — but the "threefold" magnitude was a property of the plan-phase window
and does not reproduce. Recorded as a correction to the *argument's magnitude*, not to its
conclusion.

#### Conclusion (REQ-GFC-005)

**40 is confirmed. Retained unchanged.** The evidence, in order of weight:

1. The deciding axis (Axis 2, outlier-excluded) puts the 40-crossing at 1.47 days of factory
   activity — the frequency `spec.md` §D.2 declared as the intent, reproduced on a window it was
   not derived from.
2. The alternative the SPEC considered and rejected (15) measures at 0.86 days, ≈1.7× more often.
   Nothing in this measurement argues for it.
3. Axis 1 shifted (p90 9 → 11, mean 3.9 → 4.13) without changing shape, and does not decide the
   threshold in either direction — the metric is cumulative, not per-change.
4. The threshold is not load-bearing for the reported defect regardless: the streak's corrected
   cumulative is 2, which passes at 15 and at 40 alike (`spec.md` §B.2, re-verified through the
   delivered code in M1's AC-GFC-003 record). M1 alone stops the streak.

Nothing was changed in `internal/graph/check.go` by this milestone. `DefaultThresholds` still reads
`CodemapsChangedFiles: 40` at `internal/graph/check.go:53`.

#### AC matrix — M2

| AC | Status | Deciding command | Observed |
|---|---|---|---|
| AC-GFC-006 | PASS | `git log --first-parent -30 --reverse --name-only --pretty=format:"===%h" develop -- internal cmd pkg` + the axes script | Both axes recorded above with verbatim output, the nearest-rank convention named, Axis 2 walked twice, and the frequency conclusion stated on the outlier-excluded walk |
| AC-GFC-007 | PASS | the three bounded commands below | No justification appealing to a failing check exists on any of the three surfaces |

AC-GFC-007, decided by all three of its commands, and asserted **only** over those three surfaces:

```
$ grep -n "CodemapsChangedFiles" internal/graph/check.go
43:	// CodemapsChangedFiles: red when the endpoint-diff count is >= this value.
44:	CodemapsChangedFiles int
53:		CodemapsChangedFiles: 40,
135:	rep := LayerReport{Layer: LayerCodemaps, Metric: MetricDescribedSourceDiff, Threshold: th.CodemapsChangedFiles}
216:	if count >= th.CodemapsChangedFiles {

$ git log d2cba5e21..HEAD --format=%B | grep -niE "raise|raised|lower|lowered|threshold|so that|passes"
  → 7 hits, every one of them about retaining 40 or about reversing the v0.1.0 proposal to 15:
    "Threshold untouched at 40; no restamp anywhere"
    "the threshold judgment is reversed — 40 is retained. The value 15 …"
    "The threshold change was therefore not load-bearing and is now out of scope."
  → zero hits of the form "raised/lowered so that <check> passes"

$ git diff d2cba5e21..HEAD -- internal/graph/check.go | grep -nE "^[+-].*(40|Threshold|CodemapsChangedFiles)"
  → (no threshold line added or removed)
```

#### Gaps — what M2 did NOT observe

- **The plan-phase window was not re-walked.** `spec.md` §B.6's window (`39c677f47 … 48eb945df`)
  was superseded by the `develop` window measured above; the comparison table quotes §B.6's recorded
  figures rather than re-deriving them, so those four cells are carried, not measured here.
- **One window, three days.** Both axes rest on a single 30-integration window spanning 2.421 days.
  The integration rate is an order of magnitude, not a rate constant — the same caveat `spec.md`
  §D.2 attaches to its own figure.
- **No sensitivity beyond the single largest contributor.** AC-GFC-006 requires one outlier
  counterfactual and one was run; the walk was not re-run against a second- or third-largest
  exclusion, so the crossing's stability under a broader trimming is unobserved.
- **`3cb258d62` contributes 15 and sits at the window edge.** Its effect on the crossings is nil
  (both crossings occur well before it), but it would dominate a window shifted one integration
  later, so the next re-measurement may see a different Axis 1 tail.
- **No code was executed for this milestone.** M2 is a measurement over git history; the delivered
  binary's own reading of the metric was not re-run here (M1's record carries that).

#### Residual risk — M2

- **The confirmation rests on a frequency judgment, not a mechanical criterion.** "≈1.5 days per
  red is tolerable" is a debt-tolerance position (`spec.md` §G Q4), and it is the operator's to
  revise. What is measured is the crossing; whether that crossing is the right one is a decision
  this record does not make.
- **The window is drawn from a period of unusually high factory activity** — 30 integrations in
  under two and a half days. At a lower integration rate the same 17-integration crossing is a
  longer wall-clock interval, and 40 becomes correspondingly laxer in time terms.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

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

Both columns are measured, not inferred: the pre-M1 column is
`git show 4a50d44f4:<file> | grep -n …` and the post-M1 column is `grep -n …` on the delivered tree.
Every pre-M1 coordinate reproduced exactly at `4a50d44f4`, so each row is a move of the same
construct, not a new site.

| Cited coordinate | Construct cited | pre-M1 (`4a50d44f4`) | post-M1 (`5d95a2e8d`) | What the old line holds now |
|---|---|---|---|---|
| `check.go:181` | the codemaps dirty-branch recompute | 181 | **184** | first line of the M1 comment above the recompute |
| `check.go:187` | the sole `ContentFingerprint` comparator | 187 | **190** | `rep.Reason = "described roots unreadable: "…` |
| `provenance.go:107` | `func aggregateFingerprint` (cited in `§E.1`) | 107 | **121** | doc comment of `AggregateDescribedFingerprintFiltered` |
| `provenance.go:201` | `func treeDirty` | 201 | **225** | a `ResolveCommit` doc-comment line |
| `provenance.go:208` | `func baseProvenance` (declaration) | 208 | **240** | `func ResolveCommit` |
| `provenance.go:219` | the `aggregateFingerprint` call inside `baseProvenance` | 219 | **251** | a closing brace |
| `provenance.go:237,253,264` | the three stamp writers' `baseProvenance` calls | 237 / 253 / 264 | **269 / 285 / 296** | 237 is now a doc-comment line |
| `provenance.go:280` | `Provenance.Describe`, the display-only reader | 280 | **312** | a `StampMXScan` comment line |

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
  | awk -F: '{print $2":"$3}' | sort -u          # → 26 distinct citations at 5d95a2e8d
sed -n '<N>p' <path>                              # resolve each row
```

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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

### §E.2b Citation drift introduced by M1 (recorded, refresh deferred)

M1 added code to `internal/graph/check.go` and `internal/mx/provenance.go`, which
moved three source coordinates that `spec.md` and `plan.md` cite by line number.
Measured in this tree at `5d95a2e8d` by the orchestrator, each by the grep that
decides it:

| what | cited in spec.md / plan.md (pre-M1) | actual at 5d95a2e8d | deciding command |
|---|---|---|---|
| sole comparator | `internal/graph/check.go:187` | `:190` | `grep -n "cur == pv.ContentFingerprint" internal/graph/check.go` |
| display-only reader | `internal/mx/provenance.go:280` | `:312` | `grep -n "dirty fingerprint=" internal/mx/provenance.go` |
| `treeDirty` | `internal/mx/provenance.go:201` | `:225` | `grep -n "func treeDirty" internal/mx/provenance.go` |

So the four artifacts currently disagree: `progress.md` carries post-M1
coordinates, `spec.md` and `plan.md` carry pre-M1 ones.

**The refresh is deliberately deferred to the close of run-phase, after M3.** M3
adds attribution output to `check.go` and will move these lines again; refreshing
now would be done twice. This is recorded rather than silent because this SPEC
already failed audit iter-2 on this exact defect class (N2 — `provenance.go:196`
cited while the function was at `:208`); the difference is that N2 was wrong and
unnoticed, and this is wrong and written down.

Refresh-complete is decided by extracting every `internal/…/*.go:NNN` citation
from the four artifacts and checking each against the delivered tree:

```
grep -ohE '(internal/[a-z]+/[a-z_]+\.go):[0-9]+' .moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/*.md \
  | sort -u | while IFS=: read f l; do printf '%s:%s -> ' "$f" "$l"; sed -n "${l}p" "$f"; done
```

Every row must show the construct its citation claims. Recorded by the
orchestrator because two dispatches asking the M1 agent to record it returned
idle with no commit; `spec.md` and `plan.md` bodies were not touched.

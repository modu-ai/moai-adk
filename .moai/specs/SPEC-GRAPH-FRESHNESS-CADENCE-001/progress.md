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
  (`internal/mx/provenance.go:121-123` — `aggregateFingerprint`, which still applies no extension
  filter, M1 having moved the walk itself into `aggregateFingerprintPred` where the `admit`
  predicate is threaded; `internal/graph/meta.go:67` routes three non-Go directories through it),
  so it is an established finding rather than an auditor hypothesis.
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
  **safe on a stated premise** (`check.go:222` is the only non-test `ContentFingerprint` comparator —
  re-measured at this HEAD), and `treeDirty` (`provenance.go:225`) is **examined and deferred** with
  its residual recorded (a tree dirty only under `testdata` is still refused the `--commit`
  merge-base anchor). The `treeDirty` deferral is a candidate for a follow-up card — operator's call.
- Ordering basis for t322 / t311 / t304: `spec.md` §F.
- Open questions deliberately left to the operator: `spec.md` §G (**four** — Q4 added at v0.2.0 and
  restated on the outlier-excluded cadence figure at v0.2.1).
- Evidence path: `.moai/reports/t322/`.

## §E.1c Deferred citation refresh — executed at run-phase close

The one-shot refresh scheduled by `§E.2`'s `#### Coordinate drift caused by M1` and confirmed
runnable by `#### Coordinate drift caused by M3`. Executed in worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t322`, branch `WT-graph-freshness-cadence`, against
the delivered tree at `8b11bbba1`, on 2026-08-28. Scope: `spec.md`, `plan.md`, and `§E.1`/`§E.1b`
of this file. **Line numbers and the prose that names a changed construct — nothing else.** No
requirement, judgment, scope decision or acceptance criterion is touched.

`acceptance.md` was checked rather than assumed: the enumeration below returns no row for it, so it
carries no line-number citation and needed no edit.

### The resolver

Recorded as a command rather than as a result, because a result is true only of the tree it was run
against:

```bash
W=$(git rev-parse --show-toplevel)
for art in "$@"; do
  grep -no '[A-Za-z0-9_/]*[A-Za-z0-9_]\.go:[0-9][0-9,-]*' "$art" | while IFS= read -r hit; do
    aln=${hit%%:*}; cite=${hit#*:}; file=${cite%%.go:*}.go; nums=${cite#*.go:}
    case $file in
      */*)           path=$file ;;                              # already a full path
      check.go)      path=internal/graph/check.go ;;            # basenames collide — map explicitly
      meta.go)       path=internal/graph/meta.go ;;
      symbol.go)     path=internal/graph/symbol/symbol.go ;;
      provenance.go) path=internal/mx/provenance.go ;;
      *) path=$(find "$W/internal" -name "$file" -not -path '*/testdata/*' | sed "s#^$W/##" | head -1) ;;
    esac
    for n in $(printf '%s' "$nums" | tr ',-' '  '); do        # expand comma lists AND ranges
      src=$(awk -v n="$n" 'NR==n{sub(/^[[:space:]]+/,""); print; f=1} END{if(!f) print "<NO SUCH LINE>"}' "$W/$path")
      [ -z "$src" ] && src='<BLANK LINE>'
      printf '%-13s %-4s %-40s %s\n' "$(basename "$art")" "$aln" "$path:$n" "$src"
    done
  done
done
```

It is built against four failure modes, three of them named in the M1 record and the fourth found
while running it:

1. **Multi-segment paths.** The character class admits `/`, so `internal/graph/symbol/symbol.go:99`
   matches whole. A `internal/[a-z]+/[a-z_]+\.go` shape does not.
2. **Comma lists.** `237,253,264` is expanded by `tr`, so every member is resolved rather than only
   the first. A resolver that splits on the first number silently drops the trailing two.
3. **Basename collisions.** `provenance.go` exists in `internal/config`, `internal/mx` and
   `internal/navigator/sync`; `check.go` and `symbol.go` collide likewise. Bare-form citations are
   therefore mapped through an explicit table keyed on the path each citing sentence names — never
   resolved by basename search.
4. **Ranges, and blank target lines.** `107-143` is a fourth citation form, present in `§E.1b` and
   named by neither drift record; `tr` expands it to both endpoints. Separately, a citation whose
   target line is blank must not be reported as missing — the first run mislabelled
   `check.go:168` that way, so `<BLANK LINE>` and `<NO SUCH LINE>` are now distinguished.

### Disposition

Old values below are historical by construction — this table is a record of a move, so its left
column resolves against the pre-refresh tree and its right column against `8b11bbba1`.

| Cited construct | Old | New | Disposition |
|---|---|---|---|
| `roots = mx.DefaultDescribedRoots` (`internal/graph/check.go`) | 168 | 200 | corrected |
| codemaps dirty-branch recompute (`internal/graph/check.go`) | 181 | 216 | corrected |
| sole `ContentFingerprint` comparator (`internal/graph/check.go`) | 187 | 222 | corrected |
| `dirFingerprint`'s aggregate call (`internal/graph/meta.go`) | 67 | 67 | unmoved — verified, not assumed |
| `DefaultDescribedRoots` consumer (`internal/graph/symbol/symbol.go`) | 99 | 99 | unmoved — verified, not assumed |
| `aggregateFingerprint` (`internal/mx/provenance.go`) | 107-143 | 121-123 | corrected **+ prose**: M1 moved the walk into `aggregateFingerprintPred`, so the span that once held it now holds only the nil-`admit` wrapper |
| `func treeDirty` (`internal/mx/provenance.go`) | 201 | 225 | corrected |
| `func baseProvenance` (`internal/mx/provenance.go`) | 208 | 240 | corrected |
| fingerprint call inside `baseProvenance` (`internal/mx/provenance.go`) | 219 | 251 | corrected **+ prose**: the call is now `fingerprint(projectRoot, describedRoots)` through the new `fingerprint fingerprintFunc` parameter, not a direct `aggregateFingerprint` call |
| three stamp writers' `baseProvenance` calls (`internal/mx/provenance.go`) | 237,253,264 | 269,285,296 | corrected |
| `Provenance.Describe` display reader (`internal/mx/provenance.go`) | 280 | 312 | corrected |
| N2's quoted coordinates (`internal/mx/provenance.go`) | 196 / 208 / 219 | — | **left historical** — see below |

Nothing was found unresolvable.

### What was deliberately left, and why

The `spec.md` v0.2.1 HISTORY row records audit finding N2: `baseProvenance` was cited at
`provenance.go:196`, which is inside `ResolveCommit`, and was corrected to `:208` and `:219`. In
that sentence the coordinates are the **subject** of the finding, not addresses into the tree —
`:196` is the defect and `:208` / `:219` are the remediation. Renumbering any of the three would
rewrite what the audit found and what the correction was, erasing the record. The M1 drift record
names `:196` explicitly; the reason it gives ("refreshing it would erase the correction it
documents") applies with equal force to the two values that *are* that correction, so all three are
left as written.

The discriminator applied throughout is therefore **mention versus use**: where a coordinate is
quoted as the thing being discussed it stays, and where it is used as an address for a construct it
is refreshed. Every other HISTORY citation — `treeDirty`, the `ContentFingerprint` comparator,
`baseProvenance` in the v0.2.0 D1 account — is an address, and was refreshed accordingly, so a
reader following one lands on the construct the sentence names.

`§E.2` and `§E.3` are out of scope and untouched. Their coordinates are measurements attributed to
named SHAs (`4a50d44f4`, `5d95a2e8d`, `988adaf98`), and a measurement rewritten to a later tree is
no longer a measurement.

### Verbatim resolution, post-refresh

`spec.md`, `plan.md`, `acceptance.md` — `acceptance.md` contributes no rows:

```text
spec.md       25   internal/mx/provenance.go:196            func GitHead(root string) string {
spec.md       25   internal/mx/provenance.go:196            func GitHead(root string) string {
spec.md       26   internal/mx/provenance.go:196            func GitHead(root string) string {
spec.md       26   internal/mx/provenance.go:225            func treeDirty(root string, roots []string) bool {
spec.md       26   internal/graph/check.go:222              if cur == pv.ContentFingerprint {
spec.md       27   internal/graph/meta.go:67                return mx.AggregateDescribedFingerprint(filepath.Dir(dir), []string{filepath.Base(dir)})
spec.md       27   internal/mx/provenance.go:240            func baseProvenance(projectRoot, generatedBy string, describedRoots []string, fingerprint fingerprintFunc) *Provenance {
spec.md       286  internal/graph/check.go:200              roots = mx.DefaultDescribedRoots
spec.md       287  internal/graph/symbol/symbol.go:99       for _, root := range mx.DefaultDescribedRoots {
spec.md       288  internal/mx/provenance.go:269            pv := baseProvenance(projectRoot, "codemaps-gen", DefaultDescribedRoots, AggregateDescribedFingerprintFiltered)
spec.md       288  internal/mx/provenance.go:285            pv := baseProvenance(projectRoot, "mx-scan", DefaultDescribedRoots, AggregateDescribedFingerprint)
spec.md       288  internal/mx/provenance.go:296            pv := baseProvenance(projectRoot, "graph-build", DefaultDescribedRoots, AggregateDescribedFingerprint)
spec.md       297  internal/graph/meta.go:67                return mx.AggregateDescribedFingerprint(filepath.Dir(dir), []string{filepath.Base(dir)})
spec.md       308  internal/mx/provenance.go:240            func baseProvenance(projectRoot, generatedBy string, describedRoots []string, fingerprint fingerprintFunc) *Provenance {
spec.md       310  internal/graph/check.go:216              cur, err := mx.AggregateDescribedFingerprintFiltered(projectRoot, roots)
spec.md       313  internal/graph/check.go:216              cur, err := mx.AggregateDescribedFingerprintFiltered(projectRoot, roots)
spec.md       322  internal/graph/check.go:216              cur, err := mx.AggregateDescribedFingerprintFiltered(projectRoot, roots)
spec.md       323  internal/mx/provenance.go:312            return fmt.Sprintf("provenance: tree=%s dirty fingerprint=%s", p.TreeRoot, shortHash(p.ContentFingerprint))
spec.md       335  internal/mx/provenance.go:225            func treeDirty(root string, roots []string) bool {
spec.md       475  internal/mx/provenance.go:225            func treeDirty(root string, roots []string) bool {
plan.md       19   internal/graph/check.go:200              roots = mx.DefaultDescribedRoots
plan.md       20   internal/graph/symbol/symbol.go:99       for _, root := range mx.DefaultDescribedRoots {
plan.md       20   internal/mx/provenance.go:269            pv := baseProvenance(projectRoot, "codemaps-gen", DefaultDescribedRoots, AggregateDescribedFingerprintFiltered)
plan.md       20   internal/mx/provenance.go:285            pv := baseProvenance(projectRoot, "mx-scan", DefaultDescribedRoots, AggregateDescribedFingerprint)
plan.md       20   internal/mx/provenance.go:296            pv := baseProvenance(projectRoot, "graph-build", DefaultDescribedRoots, AggregateDescribedFingerprint)
plan.md       24   internal/graph/check.go:216              cur, err := mx.AggregateDescribedFingerprintFiltered(projectRoot, roots)
plan.md       24   internal/graph/meta.go:67                return mx.AggregateDescribedFingerprint(filepath.Dir(dir), []string{filepath.Base(dir)})
plan.md       26   internal/mx/provenance.go:251            if fp, err := fingerprint(projectRoot, describedRoots); err == nil {
plan.md       30   internal/graph/check.go:216              cur, err := mx.AggregateDescribedFingerprintFiltered(projectRoot, roots)
plan.md       34   internal/graph/check.go:222              if cur == pv.ContentFingerprint {
plan.md       35   internal/mx/provenance.go:312            return fmt.Sprintf("provenance: tree=%s dirty fingerprint=%s", p.TreeRoot, shortHash(p.ContentFingerprint))
plan.md       40   internal/mx/provenance.go:225            func treeDirty(root string, roots []string) bool {
plan.md       80   internal/graph/check.go:216              cur, err := mx.AggregateDescribedFingerprintFiltered(projectRoot, roots)
plan.md       84   internal/graph/meta.go:67                return mx.AggregateDescribedFingerprint(filepath.Dir(dir), []string{filepath.Base(dir)})
```

`§E.1`/`§E.1b` of this file, resolved over `sed -n '1,55p'` of it as it stood immediately before
this subsection was appended:

```text
e1b.md        26   internal/mx/provenance.go:121            func aggregateFingerprint(projectRoot string, roots []string) (string, error) {
e1b.md        26   internal/mx/provenance.go:123            }
e1b.md        28   internal/graph/meta.go:67                return mx.AggregateDescribedFingerprint(filepath.Dir(dir), []string{filepath.Base(dir)})
e1b.md        47   internal/graph/check.go:222              if cur == pv.ContentFingerprint {
e1b.md        48   internal/mx/provenance.go:225            func treeDirty(root string, roots []string) bool {
```

Every row above resolves to the construct its citing sentence names, with one intended exception:
the three `provenance.go:196` rows, which resolve to `func GitHead` because they are N2's quotation
of a wrong citation and are preserved as such. Two of those three rows are the v0.2.2 HISTORY entry
recording this refresh, which necessarily names the coordinate it declined to change.

The reader who re-runs this later should expect different artifact line numbers and, if any source
moves again, different source coordinates. The command is the check; these rows are only what it
printed here.

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

### M3 — Failure attribution (REQ-GFC-007, REQ-GFC-008, REQ-GFC-010)

Implemented in worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t322`, branch
`WT-graph-freshness-cadence`, base HEAD `988adaf98` (M2's commit). Cycle type TDD: every criterion
was observed RED before GREEN, and each of the three mutants below was planted, observed to fail the
criterion that must catch it, and reverted.

#### Delivered surface

| Site | Change |
|---|---|
| `internal/graph/check.go` `LayerReport` | Five new fields: `Contribution *int`, `ContributionBase string`, `ContributionAbsentReason string`, `DrivingPaths []string`, `DrivingPathsOmitted int`. All `omitempty`; the five pre-existing fields keep their names, types, and meanings. |
| `internal/graph/check.go` `drivingPathDisplayBound` (new const) | `10` — the readable maximum REQ-GFC-008 requires. Overflow is declared in `DrivingPathsOmitted`, never dropped. |
| `internal/graph/check.go` `gitDiffNameList` (new) | The measurement `gitDiffNameCount` now wraps: the sorted described-worthy path set, unioned across the endpoint diff and the untracked listing. Count and list come from **one** traversal, so a report cannot name paths its own count disagrees with. `gitDiffNameCount` is preserved as a thin `len()` wrapper, so AC-GFC-002's criterion still decides the same function. |
| `internal/graph/check.go` `attachContribution` (new) | Resolves `HEAD^1` via `git rev-parse --verify --quiet`; on success measures the contribution with the **same predicate and the same union** as the cumulative count, differing only in base. On failure sets `ContributionAbsentReason` and leaves `Contribution` nil. |
| `internal/graph/check.go` `attachDrivingPaths` (new) | Truncates to the bound and records the overflow count. |
| `internal/graph/check.go` `checkCodemaps` | Contribution attached at **all three measured exits** (dirty-fresh, dirty-stale, clean). Driving paths attached on the **stale** verdict only, matching REQ-GFC-008's `When the codemaps layer is stale` guard. |
| `internal/cli/graph_check.go` `writeLayerAttribution` (new) | Renders the attribution beneath each offending layer's stderr verdict line. |

**Two placement decisions, stated because they are judgments rather than mechanics.**

*The contribution is attached on the dirty-fingerprint path too, but the driving paths are not.* The
contribution is a property of the change under judgment, so it is measurable regardless of how the
stamp was written. The driving-path list is not: the dirty path compares one hash against another
and holds no path set at all, so producing a list there would name files the metric never counted.
An empty list is the honest output, and it is what that path emits.

*Driving paths are guarded on the stale verdict.* A fresh layer has nothing driving it; listing the
sub-threshold churn there would present tolerated drift as if it were the cause of a failure that
did not occur.

**REQ-GFC-007 reports; it does not gate.** The exit code is still decided by the cumulative count
alone (`rep.Value >= th.CodemapsChangedFiles`). `plan.md` §G names gating on the per-change count as
an anti-pattern — it would eliminate inheritance and the accumulated-drift signal together — and
names a *cumulative* count displayed without gating as the mutant. Neither is present: the
cumulative still gates, the contribution still only reports.

#### RED evidence, captured before each implementation

```
$ go test ./internal/graph/ -run 'TestCheckCodemaps_Contribution|TestCheckCodemaps_DrivingPaths' -count=1
internal/graph/check_attribution_test.go:65:10: rep.Contribution undefined (type LayerReport has no field or method Contribution)
internal/graph/check_attribution_test.go:71:10: rep.ContributionAbsentReason undefined (type LayerReport has no field or method ContributionAbsentReason)
internal/graph/check_attribution_test.go:113:17: undefined: drivingPathDisplayBound
internal/graph/check_attribution_test.go:131:14: rep.DrivingPaths undefined (type LayerReport has no field or method DrivingPaths)
FAIL	github.com/modu-ai/moai-adk/internal/graph [build failed]

$ go test ./internal/cli/ -run 'TestGraphCheckCmd_StaleStderrNamesDrivingPaths' -count=1
--- FAIL: TestGraphCheckCmd_StaleStderrNamesDrivingPaths (0.55s)
    graph_check_attribution_test.go:61: stderr names no driving path — the count alone was reported:
```

**One RED is disclosed as weaker than it looks.** `TestGraphCheckCmd_JSONCarriesAttribution` was
authored after the struct fields had already landed for the unit tests, so its first run failed on a
harness detail (cobra appends its usage block to the same writer, breaking a whole-buffer
`json.Unmarshal`; fixed by decoding only the first JSON value) rather than on the criterion. Its
genuine RED is mutant (b) below, where `driving_paths` is absent from the row and the criterion
fails on the requirement instead of on the harness. Recorded rather than presented as a clean RED.

#### Mutant runs (RED under the mutant → GREEN after revert)

**(a) a missing first parent defaults to 0** — `attachContribution` sets `Contribution = &0`
instead of recording the absent reason. This is the mutant AC-GFC-008 names.

```
--- FAIL: TestCheckCodemaps_Contribution/no_first_parent_reports_the_contribution_absent,_never_0
    check_attribution_test.go:94: contribution = 0 (present), want absent
        — HEAD has no first parent, so a present 0 is fabricated
--- FAIL: TestGraphCheckCmd_JSONCarriesAttribution
    graph_check_attribution_test.go:150: contribution is present on a checkout with no first parent — a fabricated 0: 0
--- FAIL: TestGraphCheckCmd_StaleStderrNamesDrivingPaths
    graph_check_attribution_test.go:78: stderr does not state why the contribution is unmeasured
```

Killed on all three surfaces. The measured-zero sub-test (the inheriting merge) stays GREEN under
the mutant, which is the point: the criterion separates a fabricated 0 from a measured one, rather
than merely asserting that some 0 appears. After revert: `ok ... 1.811s`.

**(b) the report carries the count alone** — the `attachDrivingPaths` call removed.

```
--- FAIL: TestCheckCodemaps_DrivingPaths/over_the_display_bound…
    check_attribution_test.go:132: driving paths listed = 0, want 10 (0 = the report carries the count alone)
--- FAIL: TestCheckCodemaps_DrivingPaths/under_the_display_bound…
    check_attribution_test.go:163: driving paths listed = 0, want 4 (0 = the report carries the count alone)
--- FAIL: TestGraphCheckCmd_StaleStderrNamesDrivingPaths
    graph_check_attribution_test.go:61: stderr names no driving path — the count alone was reported
--- FAIL: TestGraphCheckCmd_JSONCarriesAttribution
    graph_check_attribution_test.go:138: codemaps row carries no driving_paths field:
        map[contribution_absent_reason:… layer:codemaps metric:described-source-diff threshold:40 value:41 verdict:stale]
```

Killed on all three. After revert: `ok`.

**(c) the attribution is computed but never rendered** — `writeLayerAttribution` removed from the
CLI while the struct keeps every field. This is the mutant the unit tests **cannot** catch, and it
is the one that matters operationally: `.github/workflows/graph-freshness.yml:87` runs
`./bin/moai graph check`, so stderr is the surface a lane actually reads.

```
$ go test ./internal/graph/ -run TestCheckCodemaps -count=1
ok  	github.com/modu-ai/moai-adk/internal/graph	4.760s          ← unit tests GREEN under the mutant
$ go test ./internal/cli/ -run 'TestGraphCheckCmd_StaleStderrNamesDrivingPaths|TestGraphCheckCmd_JSONCarriesAttribution'
--- FAIL: TestGraphCheckCmd_StaleStderrNamesDrivingPaths (0.60s)
    graph_check_attribution_test.go:61: stderr names no driving path — the count alone was reported
```

The JSON criterion also stayed GREEN, correctly: JSON is a separate render, and the mutant leaves it
intact. Only the stderr criterion moved, which is what makes it a criterion rather than a duplicate.
After revert: `ok ... 4.068s`.

#### AC matrix — M3

| AC | Status | Deciding command | Observed |
|---|---|---|---|
| AC-GFC-008 | PASS | `go test ./internal/graph/ -run TestCheckCodemaps_Contribution -count=1` | `ok github.com/modu-ai/moai-adk/internal/graph` — both sub-tests PASS; mutant (a) fails the no-first-parent sub-test |
| AC-GFC-009 | PASS | `./bin/moai graph check --root <fixture>` (stderr) | verbatim below |
| AC-GFC-010 | PASS | `./bin/moai graph check --root <fixture> --json \| jq '.layers[] \| select(.layer=="codemaps")'` | verbatim below |

#### AC-GFC-009 fixture and stderr, verbatim

The fixture is a throwaway git repository built to the inherited-red shape the SPEC exists to make
legible: commit 1 is the stamp target, commit 2 adds 41 described-worthy `.go` files plus the two
kinds of noise the predicate must reject (`f01_test.go`, `testdata/x.go`), commit 3 adds nothing
described-worthy and is HEAD. The provenance block was **hand-written** at commit 1 — `moai graph
stamp` was not run anywhere, in this tree or in the fixture (`plan.md` §D, REQ-GFC-009).

```
$ ./bin/moai graph check --root <fixture>
codemaps  metric=described-source-diff value=41 threshold=40 verdict=stale
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=absent  (mx-index absent …)
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=absent  (edges.jsonl absent …)
graph check: layer codemaps verdict=stale value=41 threshold=40 —
  contribution: 0 described-worthy file(s) vs first parent 5d7bf3e92 (inherited — this change contributed none of it)
    internal/aged/f01.go
    internal/aged/f02.go
    internal/aged/f03.go
    internal/aged/f04.go
    internal/aged/f05.go
    internal/aged/f06.go
    internal/aged/f07.go
    internal/aged/f08.go
    internal/aged/f09.go
    internal/aged/f10.go
    ... and 31 more (listing bounded)
graph check: layer mx-index verdict=absent value=0 threshold=1 — …
graph check: layer edges verdict=absent value=0 threshold=0 — …
exit=1
```

Ten paths listed, the overflow declared as `31 more`, `41 = 10 + 31` reconciling, and neither
`f01_test.go` nor `testdata/x.go` appearing in the listing or in the count. The verdict line reads
as the operator needs it to: 41 files are red, this change caused none of them.

#### AC-GFC-010, verbatim

```
$ ./bin/moai graph check --root <fixture> --json | jq '.layers[] | select(.layer=="codemaps")'
{
  "layer": "codemaps",
  "metric": "described-source-diff",
  "value": 41,
  "threshold": 40,
  "verdict": "stale",
  "contribution": 0,
  "contribution_base": "5d7bf3e9294c750419fe144666bfd3fb4c4d3eb7",
  "driving_paths": [ "internal/aged/f01.go", … "internal/aged/f10.go" ],
  "driving_paths_omitted": 31
}
```

The five pre-existing fields (`layer`, `metric`, `value`, `threshold`, `verdict`) are unchanged in
name and meaning; `reason` is absent here only because a stale endpoint-diff carries none, exactly
as before this change. The new fields are distinct keys, not overloads of existing ones.

#### This tree's own corrected reading (M1's open gap, now closed)

M1 recorded as a gap that the worktree's own `graph check` had not been re-run after the change. Run
now, at M3:

```
$ ./bin/moai graph check --json | jq '.layers[] | select(.layer=="codemaps")'
{ "layer": "codemaps", "metric": "described-source-diff", "value": 4, "threshold": 40,
  "verdict": "fresh", "contribution": 2, "contribution_base": "988adaf98ddf57f5cb72a0bb53f96166b54e20c8" }
```

Four described-worthy files differ from the stamp; this change contributed two of them
(`internal/graph/check.go` and `internal/cli/graph_check.go` — the two new `_test.go` files are
correctly excluded by the predicate). Fresh at 40, and the attribution reads truthfully on a real
tree rather than only on a fixture.

#### Independent verification batch (this run, this tree, base `988adaf98`)

```
$ go test -cover ./internal/graph/... ./internal/mx/... -count=1
ok  github.com/modu-ai/moai-adk/internal/graph         16.756s  coverage: 88.0% of statements
ok  github.com/modu-ai/moai-adk/internal/graph/symbol   1.388s  coverage: 86.2% of statements
ok  github.com/modu-ai/moai-adk/internal/mx             6.030s  coverage: 88.9% of statements
$ go test -cover ./internal/cli/ -count=1
ok  github.com/modu-ai/moai-adk/internal/cli          365.242s  coverage: 79.5% of statements
$ go vet ./internal/graph/... ./internal/mx/... ./internal/cli/... ./internal/config/...   → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...                                                 → exit 0
$ golangci-lint run --timeout=5m ./internal/graph/... ./internal/cli/...                   → 0 issues.
$ gofmt -l internal/graph/check.go internal/cli/graph_check.go \
        internal/graph/check_attribution_test.go internal/cli/graph_check_attribution_test.go
  → (no output)
$ make build                                                                               → exit 0
```

Constraint checks:

```
$ grep -rn "graph stamp" .github/ .claude/                              → no output (rc 1)
$ git diff --stat d2cba5e21 -- .moai/project/codemaps/provenance.json   → empty
$ grep -n "CodemapsChangedFiles: 40" internal/graph/check.go            → 53 (unchanged by M2 and M3)
```

The full local suite was NOT run (`plan.md` §D): CI renders the full verdict. `e2e/` was not run.

#### Coordinate drift caused by M3

M3 inserted lines into `internal/graph/check.go` only. `internal/mx/` is untouched
(`git diff --stat HEAD -- internal/mx/` → empty), so **every `provenance.go` row in M1's drift table
is unchanged by M3** and its post-M1 values still stand. `internal/cli/graph_check.go` is modified
but is cited by no artifact.

Both columns measured, not inferred — pre-M3 by `git show HEAD:internal/graph/check.go | grep -n …`,
post-M3 by `grep -n …` on the delivered tree:

| Cited coordinate | Construct cited | post-M1 (`5d95a2e8d`) | post-M3 | Deciding command |
|---|---|---|---|---|
| `check.go:168` | `roots = mx.DefaultDescribedRoots` | 168 | **200** | `grep -n "roots = mx.DefaultDescribedRoots"` |
| `check.go:181` | the codemaps dirty-branch recompute | 184 | **216** | `grep -n "AggregateDescribedFingerprintFiltered(projectRoot, roots)"` |
| `check.go:187` | the sole `ContentFingerprint` comparator | 190 | **222** | `grep -n "cur == pv.ContentFingerprint"` |

`check.go:168` was listed in M1's *unmoved* table — M1's insertions were all below it. M3's are not:
the five new `LayerReport` fields and `drivingPathDisplayBound` sit above it, so it moves for the
first time here. Recorded explicitly so a refresh that trusts M1's "unmoved" column does not skip it.

One construct changed shape as well as position and a refresh must not copy only the number: the
threshold comparison formerly read `if count >= th.CodemapsChangedFiles` and now reads
`if rep.Value >= th.CodemapsChangedFiles`, because the count is derived from the path list.

The refresh itself remains deferred to the close of run-phase per the M1 drift record above
(`#### Coordinate drift caused by M1`), and this milestone does not perform it: `spec.md` and
`plan.md` bodies are manager-spec's. With M3 landed, no further milestone moves these coordinates,
so the deferred refresh can now be run once against a stable tree — its deciding command is the
enumeration recorded in that section.

The M1 record's line-number *values* are as measured at `5d95a2e8d` and the three `check.go` rows
above supersede them; the `provenance.go` rows there still stand. A refresh reads both records
together, not the M1 one alone.

#### Gaps — what M3 did NOT observe

- **The full local test suite was not run.** Only `internal/graph/...`, `internal/mx/...`, and
  `internal/cli/` were executed. A regression in any other package is unobserved here and is CI's
  verdict to render.
- **`e2e/` was not run** (the operator has a live console; that suite mutates real profiles).
- **No cross-platform *execution*.** Windows was verified by `go build` only — `GOOS=windows go vet`
  was not re-run for M3, and no test executed on Windows. The `HEAD^1` resolution and the
  `git rev-parse --verify --quiet` invocation are unexercised there; they use the same `gitOutput`
  path M1 already ships, so the risk is low but it is unobserved rather than measured.
- **`internal/cli` coverage is 79.5%, below the 85% target.** This is the package's pre-existing
  baseline for a very large package, not a regression M3 introduced (M3 only adds tests to it), but
  the figure is stated rather than omitted because the target is not met.
- **The hook gate's notice was deliberately NOT changed.** `internal/hook/quality/gate.go:1310-1317`
  renders a compact one-line graph-freshness notice and still names only layer, verdict, value, and
  threshold. It is left alone because AC-GFC-009's surface is the CLI, because CI runs the CLI
  (`.github/workflows/graph-freshness.yml:87`), and because a bounded path listing does not fit a
  one-line advisory notice. A reader of the hook notice therefore still cannot tell an inherited red
  from an originated one — a real residual, recorded rather than silently accepted.
- **`git rev-parse --verify --quiet HEAD^1` was not exercised against a detached or unborn HEAD.**
  The root-commit case is covered by a test; a genuinely unborn `HEAD` (fresh `git init`, no commit)
  reaches the same absent branch by construction but was not run.
- **The display bound of 10 is a judgment, not a measurement.** No reader was asked what listing
  length is readable; the number was chosen and is unvalidated.

#### Residual risk — M3

- **`attachContribution` runs two extra git invocations per measured codemaps check** (a
  `rev-parse` and a `diff` + `ls-files` pair). On a very large repository this roughly doubles the
  codemaps layer's git cost. Not measured; stated because the gate runs on every CI job.
- **The contribution's base is `HEAD^1`, which means different things on different histories.** On a
  merge commit it is the mainline the merge landed on, which is the intended reading. On a linear
  commit it is simply the previous commit, which is a coherent but weaker notion of "the change
  under judgment". `plan.md` §F names this asymmetry; the implementation reports the figure with its
  base sha attached (`contribution_base`) so a reader can see what it was measured against rather
  than having to assume.
- **Driving paths name what differs from the stamp, not what is undescribed.** A path in the listing
  is a described-worthy file that moved since the codemaps documents were generated; whether the
  documents actually needed to describe it is a separate judgment the metric does not make.
- **The listing is bounded at 10 with the remainder counted, so a red driven by a long tail shows
  only its alphabetical head.** Sorting is stable rather than ranked — there is no notion of which
  driving path matters most, so the ten shown are the first ten by path, not the ten worth reading.

## §E.3 Run-phase Audit-Ready Signal

_Backfilled at sync-phase (t322) from §E.2's already-recorded evidence. Run-phase closed with the
§E.2 record above and left this section a placeholder; no run-phase agent authored a standalone
audit-ready signal block of its own. Every measurement below is a verbatim carry from §E.2 — this
is a sync-phase reconstruction, not a fresh run-phase measurement. One figure is NOT such a carry
and is marked here rather than left to look like one: `run_commit_sha` names the merge commit
`44095ddc2`, which post-dates the run-phase record and appears nowhere in §E.2 (`grep -c 44095ddc2`
against the §E.2 base → 0). It was read from git at sync-phase by the lane, and its provenance is
that observation, not §E.2._

```
run_status: audit-ready
run_complete_at: 2026-08-28 (M3 merge 44095ddc2)
run_commit_sha: 44095ddc2cc1c9fed2b3bd5ac946f48017988aba
  (M1 5d95a2e8dade18a6a890c2e4e3bda8c0aebed0b7, M2 988adaf98ddf57f5cb72a0bb53f96166b54e20c8,
   M3 8b11bbba181727a9ff750dfe100d9b7b03cbde0c)
```

AC matrix (carried from §E.2): 12/12 live AC PASS — AC-GFC-001..010, 013, 014, 0 FAIL. AC-GFC-011
and AC-GFC-012 withdrawn at plan-phase (v0.2.0, audit D8) before run-phase began.

Independent verification batch (carried verbatim from §E.2, base `988adaf98`):

```
$ go test -cover ./internal/graph/... ./internal/mx/... -count=1
ok  github.com/modu-ai/moai-adk/internal/graph         16.756s  coverage: 88.0% of statements
ok  github.com/modu-ai/moai-adk/internal/graph/symbol   1.388s  coverage: 86.2% of statements
ok  github.com/modu-ai/moai-adk/internal/mx             6.030s  coverage: 88.9% of statements
$ go test -cover ./internal/cli/ -count=1
ok  github.com/modu-ai/moai-adk/internal/cli          365.242s  coverage: 79.5% of statements
$ go vet ./internal/graph/... ./internal/mx/... ./internal/cli/... ./internal/config/...   → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...                                                 → exit 0
$ golangci-lint run --timeout=5m ./internal/graph/... ./internal/cli/...                   → 0 issues.
$ make build                                                                               → exit 0
```

Gaps: the full local suite (`go test ./...`) was NOT run at run-phase (per `plan.md` §D — CI renders
the full verdict); `e2e/` was not run. This section did not exist as a run-phase artifact; treat its
attribution as sync-phase-reconstructed, not run-phase-observed.

## §E.4 Sync-phase Audit-Ready Signal

```
sync_status: audit-ready
sync_complete_at: 2026-08-28
sync_commit_sha: bc66c30b74a9acb8899886deb8af0421135541b4
```

**What sync changed**: `CHANGELOG.md` `[Unreleased]` → `### Added` entry; `docs-site/content/{ko,en,ja,zh}/cli-reference/graph.md`
— one paragraph added identically in structure to each locale's `## moai graph check` section,
describing the new stale-verdict `contribution` / `contribution_base` / `driving_paths` /
`driving_paths_omitted` attribution that M3 added to both the stderr rendering and `--json` output;
this file's §E.3 backfill and this §E.4 close; and `spec.md` frontmatter
`status: in-progress → implemented → completed` (the single sync commit merges the terminal
transition, per `spec-frontmatter-schema.md` § Status Transition Ownership Matrix).

B12 self-test (CHANGELOG emission discipline, `manager-develop-prompt-template.md` §B12):

```
a. pre-emission grep, BEFORE this edit:
   $ grep -c 'SPEC-GRAPH-FRESHNESS-CADENCE-001' CHANGELOG.md   → 0  (no duplicate entry)
b. AC count match:
   $ grep -oE 'AC-GFC-[0-9]+' acceptance.md | sort -u | wc -l   → 14
   (14 = 12 live [AC-GFC-001..010, 013, 014] + AC-GFC-011/012, both still tokened in acceptance.md's
   §D matrix as *withdrawn* rows. The CHANGELOG entry states the live count, 12, and names the two
   withdrawn IDs explicitly — matching acceptance.md's own §D.1 disposition rather than the raw
   14-token grep, which is a RED flag deliberately inspected rather than trusted at face value.)
c. file path verification:
   $ ls internal/graph/check.go internal/mx/described_worthy.go internal/mx/provenance.go \
       internal/cli/graph_check.go docs-site/content/ko/cli-reference/graph.md \
       docs-site/content/en/cli-reference/graph.md docs-site/content/ja/cli-reference/graph.md \
       docs-site/content/zh/cli-reference/graph.md
   → all 8 paths present (see § Verification below for the actual run)
```

**Docs-site scope decision**: only `## moai graph check` in the CLI-reference page changed. No
other page across the 4-locale docs-site set references `CodemapsChangedFiles`, the described-worthy
predicate, or the threshold-40 cadence at the granularity this SPEC's change affects — so no other
page was touched. The four edits are structurally identical (same insertion point, same field
names verbatim, translated prose).

Frontmatter transition applied in this commit — **`spec.md` only**, which is the one artifact in
this SPEC set carrying a `status:` key: `status: in-progress → implemented → completed`. `plan.md`,
`acceptance.md`, and `progress.md` carry no `status:` key at all, so there was nothing to transition
in them; the convention was checked against a sibling closed SPEC rather than assumed
(`grep -m1 '^status:'` over all four artifacts of SPEC-WORKTREE-BASEREF-001 returns the `spec.md`
row alone). `updated:` already read `2026-08-28` before this commit and therefore did not move — the
end state is correct, but no refresh was performed and none is claimed.

**Gaps**: §E.3 above is a sync-phase backfill from §E.2, not a run-phase-authored signal (stated in
its own header). The docs-site paragraph was authored by transcribing the AC-GFC-009/010 verbatim
fixture output already recorded in §E.2 — sync-phase did NOT independently re-run the CLI against a
freshly reconstructed stale-fixture to re-observe the stderr/JSON attribution; the sentence's
factual content is a direct carry of already-verified §E.2 evidence, not a fresh sync-phase
verification of the CLI's stale-output behavior. The full repository test suite (`go test ./...`)
was not run at sync, per this repository's standing contract (`AGENTS.md` §4 / `CLAUDE.local.md`
§6) — CI on the integration branch (`origin/develop`) is the authoritative full-suite verdict and
is PENDING at the time this section was written.

**Residual-risk**: the docs-site paragraph could drift from the CLI's actual stderr/JSON rendering
if a later change alters field names, the display bound (10), or the omission-summary format
without a corresponding docs update — no CI guard ties this page's prose to
`internal/graph/check.go`'s literal output, so a future silent divergence is possible. The
`pending-backfill-sync` placeholder for `sync_commit_sha` is filled in the immediately following
commit, per the established convention (`spec-frontmatter-schema.md` § SHA placeholder backfill
exemption) — until that commit lands, this section's own `sync_commit_sha` field is provisional.

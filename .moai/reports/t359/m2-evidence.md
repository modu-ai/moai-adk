# M2 Evidence — SPEC-TODO-LANDING-EVIDENCE-001 (card t359)

Verbatim command + output pairs for M2 (the record type and the attribution boundary). The verdict
and its attribution live in `.moai/specs/SPEC-TODO-LANDING-EVIDENCE-001/progress.md` §E.2; this file
carries the material.

**Tree.** `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`, branch `WT-landing-evidence`.
`/Users/goos/moai/moai-adk-go` is the SAME tree under a second spelling (M1's tree-identity note),
so the hazard is worktree → primary drift and the discriminant is branch + HEAD.

**HEAD at dispatch vs HEAD at first M2 write.** The dispatch named `d6420c1bd` and said the tree was
clean. Neither held at the moment M2 opened — see § Step 0 below. The baseline every M2 claim is
attributed to is `56af37cbb`.

---

## Step 0 — the dispatch's stated HEAD had already moved when M2 opened

Reported to the lead before any edit (§ Foreign write during the M2 window, below). Four calls, in
order, all in the t359 worktree:

```
$ git rev-parse --show-toplevel && git rev-parse HEAD && git branch --show-current && git status --short
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359
d6420c1bd73b42f6e14e96020a2af9f90bb2ce73
WT-landing-evidence
M  .moai/reports/t359/m1-evidence.md
```

The third line of `git status --short` output is `M ` with the change **staged** — the tree was not
clean. Two subsequent reads of that staged diff returned nothing:

```
$ git --no-pager diff --cached -- .moai/reports/t359/m1-evidence.md | head -100; echo "EXIT=$?"
EXIT=0

$ git diff --cached > /tmp/t359_staged.diff 2>&1; wc -l /tmp/t359_staged.diff
       0 /tmp/t359_staged.diff
```

Re-reading HEAD showed it had moved:

```
$ git status --short; echo "---HEAD---"; git rev-parse HEAD; echo "---LOG---"; git log --oneline -3
---HEAD---
56af37cbb3e1c4b8ebcc634195ee26a7bc26712c
---LOG---
56af37cbb docs(SPEC-TODO-LANDING-EVIDENCE-001): fold review corrections into M1 evidence (t359)
d6420c1bd docs(SPEC-TODO-LANDING-EVIDENCE-001): M1 evidence export + progress.md §E.2 (t359)
2dacb1d83 refactor(kanban): drop concatenated SQL in the landing test helper (t359)
```

`56af37cbb` touches only `.moai/reports/t359/m1-evidence.md`, which is orthogonal to M2's
deliverables. M2 proceeded on `56af37cbb` after reporting.

---

## Step 1 — the pre-edit baseline, measured on `56af37cbb`

Run before any M2 file existed, so a later red is attributable to M2's edits.

```
$ git rev-parse HEAD > /tmp/t359_baseline_head.txt && go test ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	136.974s
rc=0
```

```
$ cat /tmp/t359_baseline_head.txt
56af37cbb3e1c4b8ebcc634195ee26a7bc26712c
```

---

## Step 2 — the RED: the record type does not exist

`internal/kanban/landing_evidence_test.go` was written first. With no implementation, the package
does not build — which is the RED for AC-TLE-005, AC-TLE-006 (its `LandingEvidenceValue(nil)` half),
and AC-TLE-013.

```
$ go test ./internal/kanban/... -count=1 -run 'TestLandingEvidence'
# github.com/modu-ai/moai-adk/internal/kanban [github.com/modu-ai/moai-adk/internal/kanban.test]
internal/kanban/landing_evidence_test.go:32:10: undefined: LandingEvidence
internal/kanban/landing_evidence_test.go:37:15: undefined: LandingSHASourceOperator
internal/kanban/landing_evidence_test.go:41:18: undefined: EncodeLandingEvidence
internal/kanban/landing_evidence_test.go:45:14: undefined: DecodeLandingEvidence
internal/kanban/landing_evidence_test.go:117:16: undefined: LandingEvidenceValue
internal/kanban/landing_evidence_test.go:136:9: undefined: LandingEvidence
internal/kanban/landing_evidence_test.go:141:18: undefined: EncodeLandingEvidence
internal/kanban/landing_evidence_test.go:150:5: undefined: LandingKeyRefHead
internal/kanban/landing_evidence_test.go:150:26: undefined: LandingKeyDeliveringSHA
internal/kanban/landing_evidence_test.go:151:85: undefined: LandingKeyRefHead
internal/kanban/landing_evidence_test.go:151:85: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/kanban [build failed]
FAIL
rc=1
```

**What this RED does and does not establish.** It establishes that the tests were authored before
the implementation and that the symbols they name did not previously exist. It does NOT establish
per-field discrimination — that is AC-TLE-005's own stated RED (drop one field from the encoder), and
it is a **known loss**: it was NOT performed. See § What is NOT in this file.

---

## Step 3 — the GREEN

`internal/kanban/landing_evidence.go` added.

```
$ gofmt -l internal/kanban/
(no output)

$ go test ./internal/kanban/... -count=1 -run 'TestLandingEvidence' -v
=== RUN   TestLandingEvidence_CarriesAllSixFacts
--- PASS: TestLandingEvidence_CarriesAllSixFacts (0.00s)
=== RUN   TestLandingEvidence_AbsenceIsSQLNull
--- PASS: TestLandingEvidence_AbsenceIsSQLNull (0.01s)
=== RUN   TestLandingEvidence_RefHeadIsNotADeliveringSHA
--- PASS: TestLandingEvidence_RefHeadIsNotADeliveringSHA (0.00s)
=== RUN   TestLandingEvidence_RefusesUnpairedProvenance
=== RUN   TestLandingEvidence_RefusesUnpairedProvenance/no_observation_instant
=== RUN   TestLandingEvidence_RefusesUnpairedProvenance/sha_without_provenance
=== RUN   TestLandingEvidence_RefusesUnpairedProvenance/provenance_without_sha
=== RUN   TestLandingEvidence_RefusesUnpairedProvenance/sha_with_a_non-operator_provenance
=== RUN   TestLandingEvidence_RefusesUnpairedProvenance/no_ref
=== RUN   TestLandingEvidence_RefusesUnpairedProvenance/no_ref_head
--- PASS: TestLandingEvidence_RefusesUnpairedProvenance (0.00s)
    --- PASS: TestLandingEvidence_RefusesUnpairedProvenance/no_observation_instant (0.00s)
    --- PASS: TestLandingEvidence_RefusesUnpairedProvenance/sha_without_provenance (0.00s)
    --- PASS: TestLandingEvidence_RefusesUnpairedProvenance/provenance_without_sha (0.00s)
    --- PASS: TestLandingEvidence_RefusesUnpairedProvenance/sha_with_a_non-operator_provenance (0.00s)
    --- PASS: TestLandingEvidence_RefusesUnpairedProvenance/no_ref (0.00s)
    --- PASS: TestLandingEvidence_RefusesUnpairedProvenance/no_ref_head (0.00s)
=== RUN   TestLandingEvidence_UnknownSpecStatusIsExplicit
--- PASS: TestLandingEvidence_UnknownSpecStatusIsExplicit (0.00s)
=== RUN   TestLandingEvidence_DecodeRefusesGarbage
--- PASS: TestLandingEvidence_DecodeRefusesGarbage (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.399s
rc=0
```

---

## Step 4 — AC-TLE-011, the three-match fixture

`internal/kanban/prlink_landed_attribution_test.go` added. It builds a fixture repository whose
`origin/main` history contains three commits naming `t359fixture`, with a HEAD commit that names it
not at all (so the ref position is not one of the matches — the construction AC-TLE-012's exclusion
relies on).

```
$ go test ./internal/kanban/... -count=1 -run 'TestResolver_NamesNoDeliveringCommit' -v
=== RUN   TestResolver_NamesNoDeliveringCommit
--- PASS: TestResolver_NamesNoDeliveringCommit (0.36s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.778s
rc=0
```

A pass here on its own asserts nothing about the code's ability to fail — the assertion is a
containment check over an absence. The mutant below is what makes it evidence.

---

## Step 5 — the §D.2 plant: the resolver carries its first grep match

**Before the plant.**

```
$ shasum -a 1 internal/kanban/prlink.go internal/kanban/prlink_landed.go
a2f73970a98f536b1af8f853b167cf349e6ca345  internal/kanban/prlink.go
0ba0d4180c53a6800322bdb4005c9ad403e27492  internal/kanban/prlink_landed.go
```

**The mutant.** Three edits to `internal/kanban/prlink.go` only (`prlink_landed.go` untouched):

1. `PRLinkOutcome` gains `FirstMatch string \`json:"first_match,omitempty"\``.
2. The `case LandingLanded:` branch of `ResolveCardPRLink` type-asserts the querier to
   `GitLandedQuerier`, re-runs `LandedGrepArgs` + `q.Run("git", …)`, and assigns the first field of
   the first output line to `out.FirstMatch`.
3. `strings` added to the import block.

This is the §D.2 mutant as written — the resolver carrying its first grep match into the outcome.

**With the mutant applied.**

```
$ shasum -a 1 internal/kanban/prlink.go
c1b5dd642febeadeaa91cfd145a2751773256e39  internal/kanban/prlink.go

$ go test ./internal/kanban/... -count=1 -run 'TestResolver_NamesNoDeliveringCommit'
--- FAIL: TestResolver_NamesNoDeliveringCommit (0.42s)
    prlink_landed_attribution_test.go:182: the resolver's output names commit 3 abbreviated (f6d756b):
        {CardID:t359fixture Kind:landed PRs:[] PRState: Confidence: FirstMatch:f6d756b}
        kanban.PRLinkOutcome{CardID:"t359fixture", Kind:"landed", PRs:[]int(nil), PRState:"", Confidence:"", FirstMatch:"f6d756b"}
        {"card_id":"t359fixture","outcome":"landed","first_match":"f6d756b"}
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.814s
FAIL
rc=1
```

The leaked commit is **the third fixture commit** — `chore: merge the release batch carrying
t359fixture`, the integration commit that merely inherited the card, not the delivering change. That
is REQ-1.10's grounds reproduced mechanically: the newest match is routinely the wrong commit.

Only the **abbreviated** assertion fired, because the mutant read `--oneline` output. The full-SHA
assertion sits on the same rendered string and was not exercised by this particular mutant; it exists
for a `%H`-shaped carry-through. Stated so the pass is not read as covering both carriers.

**After the revert.**

```
$ git restore -- internal/kanban/prlink.go && shasum -a 1 internal/kanban/prlink.go && git status --short
a2f73970a98f536b1af8f853b167cf349e6ca345  internal/kanban/prlink.go
?? internal/kanban/landing_evidence.go
?? internal/kanban/landing_evidence_test.go
?? internal/kanban/prlink_landed_attribution_test.go
```

`prlink.go` returns to `a2f73970a98f536b1af8f853b167cf349e6ca345`, byte-identical to the pre-plant
hash. The tree shows exactly the three new M2 files and no modification to any tracked file — no
mutant survived, and no production file outside M2's deliverable list was left changed.

---

## Step 6 — final scoped verification

```
$ gofmt -l internal/kanban/
(no output)
gofmt rc=0

$ go vet ./internal/kanban/...
vet rc=0

$ go test ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	137.564s
test rc=0
```

---

## The M2/M3 boundary — what M2's assertions do and do not satisfy

The dispatch asked for this explicitly, and the answer is not uniform across the two criteria.

### AC-TLE-012 — belongs wholly to M3

Its Given-When is *"the operator records a landing **without** `--sha`"*. The verb `moai todo landed`
does not exist until M3, so no M2 construction reproduces that Given: constructing a record in-test
without a SHA asserts what the type permits, not what the verb produces. The criterion's own
discriminating power lies in the verb path — filling the field from the first grep match is a thing
the **verb** could do, and a type-level test cannot observe the verb not doing it.

`spec.md` §E agrees independently: its traceability table maps `REQ-TLE-012 | AC-TLE-012 | M3`.

M2 therefore **does not claim AC-TLE-012**. What M2 does contribute toward it is structural rather
than evidential: `LandingEvidence.Validate` refuses a stored SHA whose provenance is anything other
than `operator`, and refuses a SHA with no provenance at all — so the shape a grep-derived value
would take cannot be encoded. That is a narrowing of M3's reachable states, not a satisfaction of
M3's criterion.

### AC-TLE-013 — split; the key half is M2's, the Given and the render half are not

Its Then has two conjuncts:

| Conjunct | Owner | Why |
|---|---|---|
| "the ref head SHA appears under a key distinct from the delivering-SHA key… the delivering-SHA key is absent rather than aliased to the head" | **M2** — asserted by `TestLandingEvidence_RefHeadIsNotADeliveringSHA` | This is a property of the encoding, observable on a directly-constructed record |
| "the rendered form labels it as a ref position" | **M4** | The render (`design.md` §4's `(ref-head)` / `(operator)` marker in the `todo pr` evidence cell) does not exist until M4 |

Its Given — *"a record produced without `--sha`"* — is verb language like AC-TLE-012's, so even the
key half is asserted here against a **constructed** record rather than a **produced** one. That is
strictly weaker than the criterion as written.

M2 therefore claims AC-TLE-013 **partially**, and the honest form of the claim is: the distinct-key
invariant is verified at the type level; the criterion as written is not fully discharged until the
verb (M3) and the render (M4) exist. `spec.md` §E maps AC-TLE-013 to M2, which is consistent with
the key half being M2's — but the mapping does not make the render conjunct M2-verifiable, and this
record does not claim it does.

M2 supplies `LandingEvidence.Marker()` returning `operator` / `ref-head` as the labelling primitive
so M4 renders the distinction rather than re-deriving it from field emptiness. That is a primitive,
not the render.

---

## Foreign write during the M2 window

`agent-common-protocol.md` § Background Agent Execution requires an unexpected HEAD move on an
actively-worked worktree to be reported immediately and recorded. Both were done.

- **Observed**: HEAD moved `d6420c1bd` → `56af37cbb` between the first and fourth call of M2's
  opening batch, and the tree carried a staged modification at the first read despite the dispatch
  stating it was clean (§ Step 0).
- **Author**: `t359-m1`, still listed as an active teammate at that moment. The commit is an M1
  evidence correction.
- **Reported**: to `team-lead` before any M2 edit, stating the observation, the intent to proceed on
  `56af37cbb`, and offering to hold instead.
- **Not resolved unilaterally**: no attempt was made to revert, amend, or reconcile the foreign
  commit.

**A second foreign commit landed later in the same window**, observed at the pre-commit re-read
required before staging M2's own work:

```
$ git rev-parse --short HEAD; git branch --show-current; git status --short
b48a00285
WT-landing-evidence
 M .moai/specs/SPEC-TODO-LANDING-EVIDENCE-001/progress.md
?? .moai/reports/t359/m2-evidence.md
?? internal/kanban/landing_evidence.go
?? internal/kanban/landing_evidence_test.go
?? internal/kanban/prlink_landed_attribution_test.go

$ git show --stat --format='%H%n%an%n%s' b48a00285 | head -20
b48a002850adc506b6a2229f620bcb7f086e8f60
t
chore(SPEC-TODO-LANDING-EVIDENCE-001): draft -> in-progress (t359)

 .moai/specs/SPEC-TODO-LANDING-EVIDENCE-001/spec.md | 4 ++--
 1 file changed, 2 insertions(+), 2 deletions(-)
```

HEAD moved `56af37cbb` → `b48a00285` during M2's implementation window. The commit is the
`draft → in-progress` status transition on `spec.md` alone — orthogonal to M2's files, and M2's own
working-tree changes survived intact. Reported rather than reconciled; M2's commit sits on
`b48a00285`.

**Consequence for attribution.** M2's rc=0 baseline (§ Step 1) and every measurement above were
taken at `56af37cbb`. The M2 commit's parent is `b48a00285`. The two intervening commits touch only
`m1-evidence.md` and `spec.md` frontmatter — no Go source — so the measured baseline still describes
the code under test; but the measurements were not re-taken at `b48a00285`, and that is stated here
rather than glossed.

---

## A known scanner behaviour on this file

The PostToolUse check reports `sql-injection (high)` on quoted SQL in prose. If it fires on this
file, it is firing on documentation of the queries the tests run (`SELECT landing IS NULL FROM items
WHERE id = ?`, a parameterized statement). The text is left alone: quoting the statement is what the
record is for.

Separately, the same PostToolUse layer rejected two earlier writes of `landing_evidence_test.go`
with `go-error-ignored-blank` (`.moai/astgrep-rules/go/error-handling.yml`, pattern
`$_, $ERR = $FUNC($$$ARGS)`). The pattern's `$_` matches **any** node, not only the blank
identifier, so a plain reassignment `encoded, err = EncodeLandingEvidence(rec)` matches it. The test
was restructured to use fresh `:=` bindings — a real simplification, not a text-laundering edit; no
error is discarded anywhere in the file, before or after.

---

## What is NOT in this file (known losses — do not cite these later)

- **AC-TLE-005's own stated RED was NOT performed.** The criterion says "drop any one field from the
  encoder → that field's assertion fails". Six such plants (one per field) were not run. The only
  RED recorded for AC-TLE-005 is the compile failure in § Step 2, which shows the test predates the
  implementation but does NOT show per-field discrimination. The structural argument — six distinct
  values compared field-by-field, plus a six-key check on the wire — is an argument, not an
  observation.
- **AC-TLE-006's render half.** Only the SQL predicate (`SELECT landing IS NULL`) and the
  `LandingEvidenceValue(nil)` seam were asserted. `moai todo pr` was never run, and the "renders as
  absent" conjunct is untouched. It belongs to M4.
- **AC-TLE-013's render half and both criteria's verb-shaped Givens.** See § The M2/M3 boundary.
- **The full-SHA half of AC-TLE-011's containment assertion.** The planted mutant exercised the
  abbreviated form only. A `%H`-shaped carry-through was not planted.
- **`internal/cli` was NOT run at any point in M2.** M2 touches no file in that package. Its last
  measured state in this card is M1's baseline at `903bcc03c` (rc=0), which is a carry-over and is
  named here as such rather than offered as an M2 measurement.
- **`golangci-lint` was not run.** Only `gofmt -l` (clean) and `go vet` (rc=0).
- **`go test -race`.** Not run. `landing_evidence.go` starts no goroutine and holds no shared state,
  but that is an argument from reading the code, not a measurement.
- **Any non-darwin build or test run.** Everything here is darwin/arm64. Windows and Linux are CI's
  verdict.
- **Coverage.** No `-cover` run was made for `internal/kanban` at any point in M2.
- **The `/tmp` capture files.** `/tmp/t359_m2_*.txt` are machine-local scratch and are not exported;
  their contents are transcribed verbatim above, which is the exported form. The originals reach no
  clone and must not be cited.

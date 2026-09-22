# Sync verdict — SPEC-VACUOUS-FLOOR-GUARD-001 (card t378)

Tree: worktree `.claude/worktrees/t378`, branch `WT-vacuous-floor-guard`, HEAD `c3208b08f`,
base `3f03d9c36`. Sync-phase agent: `manager-docs`.

**Attribution key.** `[MD]` = measured by manager-develop at run-phase (pre-commit HEAD
`226bdd0dc`, recorded under `.moai/reports/t378/`). `[ORCH]` = measured independently by the
orchestrator against committed HEAD `c3208b08f`. `[DOCS]` = measured by this agent against
committed HEAD `c3208b08f` during sync-phase. Every row names its party; nothing is carried
unattributed.

---

## 1. Claim

1. The unreachable `budget < floor` branch in `TestBoardLockWaitBudgetDerivedFromNamedInputs` is
   deleted and its comment rewritten in place; one file changed, `+27/-7`, one hunk.
2. All eight acceptance criteria PASS, one (AC-VFG-008) with a stated qualification that is **not**
   a clean zero.
3. Card t372's guard `TestBoardLockWaitBudgetCoversSerializedMutations` is byte-identical, and no
   constant in `internal/kanban/board_store.go` moved.
4. Both budget guards are GREEN at HEAD `c3208b08f`, and the retained headroom assertion is
   demonstrated capable of RED at that same HEAD.
5. Every line coordinate in this card's SPEC and reports has been classified and re-verified
   against the tree it names; one drifted coordinate group was found and it lives in a file this
   agent may not edit (§6, blocker B-1).
6. `spec.md` frontmatter `status` transitions `in-progress → implemented`.
7. One `CHANGELOG.md` entry is emitted under `### Changed`.
8. No README, docs-site, or template change is warranted — checked, not assumed.

---

## 2. Evidence

### 2.1 Per-AC matrix

| AC | Verdict | Party | Command | Observed |
|---|---|---|---|---|
| AC-VFG-001 | PASS | `[MD]` | `grep -n 'boardLockWaitBudget <' …`; `grep -c 'floor :=' …`; `go vet ./internal/kanban/...` | 1 match at line 122 (t372's guard), baseline 2; `floor :=` count 1 at line 120, baseline 2; vet exit 0 |
| AC-VFG-001 | PASS (re-measured) | `[DOCS]` | `grep -n 'boardLockWaitBudget <\|floor :=' internal/kanban/board_lock_wait_test.go` | `120:	floor := time.Duration(serialized) * boardLockCIMutationCost` and `122:	if boardLockWaitBudget < floor {` — one each, both inside t372's guard |
| AC-VFG-002 | PASS | `[MD]` | scoped `-v -run TestBoardLockWaitBudgetDerivedFromNamedInputs` | `=== RUN` ×1, `--- PASS`, `ok …0.452s`; assertions at 28 / 65 / 74 / 79 |
| AC-VFG-002 | PASS (re-measured) | `[ORCH]` + `[DOCS]` | `go test -timeout 600s -count=1 -v -run 'TestBoardLockWaitBudget' ./internal/kanban/` | 2 `=== RUN`, both `--- PASS`; `ok … 0.393s` `[DOCS]` / `0.407s` `[ORCH]` |
| AC-VFG-003 | PASS | `[MD]` | M1 (cost 33ms→20ms) → whole-package → revert → re-run | RED, sole failure `board_lock_wait_test.go:55: per-mutation cost 20ms is below the CI-class observation of 33ms` **(pre-edit coordinate, tree `226bdd0dc`)**; post-revert `ok …15.975s` |
| AC-VFG-004 | PASS | `[MD]` | M2 (headroom 5→1) → whole-package → revert → re-run | RED `board_lock_wait_test.go:60: headroom factor 1 states no headroom` **(pre-edit coordinate)** + 3 further REDs each named in `mutants.md`; post-revert `ok …16.074s` |
| AC-VFG-004 | PASS (re-planted post-deletion) | `[ORCH]` | M2 replanted at `c3208b08f`, scoped guard run, then reverted | `--- FAIL` with `headroom factor 1 states no headroom` at **line 80** (the post-deletion `t.Errorf` coordinate); revert → `git status --short internal/` empty; both guards PASS again |
| AC-VFG-005 | PASS | `[MD]` | M3 (writers 10→8) → whole-package → revert → re-run | RED `board_lock_wait_test.go:46: supported writers = 8, want 10 (Factory mode's ten lanes against one queue)` **(pre-edit)** + t372's guard, predicted; post-revert `ok …16.076s` |
| AC-VFG-006 | PASS **with a documented hole** | `[MD]` | M4 form A `1650ms`, form B `1400ms` → revert | Form A **GREEN on both guards** — the derivation genuinely replaced by a bare literal and nothing reddened; form B RED `board_lock_wait_test.go:29: budget 1.4s is not the product of its named inputs (…)`; post-revert `ok …15.995s` |
| AC-VFG-007 | PASS (one-sided by construction) | `[MD]` | branch present (pre-edit) + M2 planted → scoped `-v` run | `headroom factor 1 states no headroom` PRESENT; `< headroom floor` and `board_lock_wait_test.go:40` both ABSENT from the complete output, at a 330ms budget against the 660ms composed floor |
| AC-VFG-008 | PASS **(qualified — not a clean zero)** | `[MD]` + `[DOCS]` | `git diff --stat`; per-file `git diff`; `grep -rn 'go test' .moai/reports/t378/`; `./bin/moai spec lint` | 1 file, `+27/-7`; `board_store.go` diff empty; 12/12 recorded invocations package-scoped and serial; `0 error(s), 1096 warning(s)` |

**8 of 8 PASS.** Two carry stated qualifications (AC-VFG-006's form-A hole, AC-VFG-008's
non-zero token count) and one is structurally one-sided (AC-VFG-007). None is glossed.

### 2.2 Sync-phase re-measurements `[DOCS]`, verbatim

Tree binary built from this tree, not the PATH binary:

```
$ make build
go build -ldflags "… -X …version.Commit=c3208b08f …" -o bin/moai ./cmd/moai
$ strings bin/moai | grep -c 'c3208b08f'
4
$ git status --short
(empty — the build left the tree clean)
```

```
$ ./bin/moai spec lint   (tail)
0 error(s), 1096 warning(s)
```

```
$ go test -timeout 600s -count=1 -v -run 'TestBoardLockWaitBudget' ./internal/kanban/
=== RUN   TestBoardLockWaitBudgetDerivedFromNamedInputs
=== PAUSE TestBoardLockWaitBudgetDerivedFromNamedInputs
=== RUN   TestBoardLockWaitBudgetCoversSerializedMutations
=== PAUSE TestBoardLockWaitBudgetCoversSerializedMutations
=== CONT  TestBoardLockWaitBudgetDerivedFromNamedInputs
--- PASS: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
=== CONT  TestBoardLockWaitBudgetCoversSerializedMutations
    board_lock_wait_test.go:134: constant coherence: … 50 serialized mutations; the stress test serializes 8 x 6 = 48. …
--- PASS: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.393s
```

Two `=== RUN` lines — the selector is not a zero-match selector printing a vacuous `ok`.

Scope hold, verbatim:

```
$ git diff --stat 3f03d9c36..HEAD -- internal/
 internal/kanban/board_lock_wait_test.go | 34 ++++++++++++++++++++++++++-------
 1 file changed, 27 insertions(+), 7 deletions(-)

$ git diff 3f03d9c36..HEAD -- internal/kanban/board_store.go
(empty)
```

The single hunk is `@@ -32,13 +32,33 @@`, inside `TestBoardLockWaitBudgetDerivedFromNamedInputs`.
`func TestBoardLockWaitBudgetCoversSerializedMutations` sits at original line 95 (base `3f03d9c36`)
/ line 115 (HEAD `c3208b08f`), outside the hunk in both trees. The only occurrence of
`CoversSerializedMutations` inside the diff is an **added comment line** naming it as the file's
one legitimate floor comparison.

Documentation-surface check, verbatim:

```
$ grep -rln 'boardLockWaitBudget\|boardLockHeadroom\|BudgetDerivedFromNamedInputs\|VACUOUS-FLOOR-GUARD' README.md README.*.md docs-site/
rc=1     (no matching lines in any of README.md / README.ko.md / README.ja.md / README.zh.md / docs-site/)

$ git diff --name-only 3f03d9c36..HEAD | grep -c 'internal/template/templates/'
0
```

No README change, no docs-site change, no 4-locale obligation, no template mirror, no `make build`
mirror duty. **Checked, not assumed.**

### 2.3 CHANGELOG B12 self-tests `[DOCS]`

```
$ grep -c 'SPEC-VACUOUS-FLOOR-GUARD-001' CHANGELOG.md
0                                  # B12-1: no duplicate entry — emission is safe

$ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/SPEC-VACUOUS-FLOOR-GUARD-001/spec.md | sort -u
AC-SIV-013 AC-VFG-001 … AC-VFG-008        # 9 distinct
```

B12-2 needs a note rather than a bare number. This is a **Tier S** SPEC: there is no
`acceptance.md`, the criteria are inline in `spec.md` §C. Run against `spec.md` the canonical
pattern returns **9**, and 9 is wrong — `AC-SIV-013` is card t372's criterion, present only as a
`§D`/`§G` cross-reference. Filtered to this SPEC's own domain the count is **8**, matching §C's
declared 8 and the eight rows above. The pattern is domain-agnostic by design, so on any SPEC that
cites a foreign AC it over-counts; anchoring on `AC-VFG-` is what makes the comparison meaningful.
The count is non-zero and plausible, so the self-test is satisfied rather than vacuous.

B12-3 (file-path verification) — every path named in the CHANGELOG entry resolves at HEAD:
`internal/kanban/board_lock_wait_test.go`, `internal/kanban/board_store.go`,
`.moai/specs/SPEC-VACUOUS-FLOOR-GUARD-001/spec.md`, `.moai/reports/t378/verdict.md`.

---

## 3. Baseline-attribution

| Measurement | Attributed to |
|---|---|
| `+27/-7`, one file, one hunk | `git diff --stat 3f03d9c36..HEAD -- internal/`, run in this tree at HEAD `c3208b08f` `[DOCS]` |
| both guards GREEN | `go test … -run 'TestBoardLockWaitBudget' ./internal/kanban/` at `c3208b08f` `[DOCS]`, independently at `c3208b08f` `[ORCH]`, and at `226bdd0dc` `[MD]` |
| headroom assertion can RED **after** the deletion | M2 planted and reverted at `c3208b08f` `[ORCH]`, failure at line 80 |
| headroom assertion can RED **before** the deletion | M2 at `226bdd0dc` `[MD]`, failure at line 60 |
| `0 error(s), 1096 warning(s)` | `./bin/moai spec lint`, tree binary built at `c3208b08f`, `strings … | grep -c c3208b08f` = 4 `[DOCS]` |
| t372's guard untouched | hunk-span arithmetic + `grep` on the diff, at `c3208b08f` `[DOCS]` |
| no constant retuned | `git diff 3f03d9c36..HEAD -- internal/kanban/board_store.go` empty `[DOCS]` |

The PATH binary `~/go/bin/moai` was **not** used anywhere in this verdict. It is v3.1.2 and
predates card t299's `SyncSHASlotFormat` rule; earlier in this card it produced a spurious
`8 error / 64 warning` reading. Every lint figure above comes from `./bin/moai` built in this tree
and confirmed to carry the HEAD SHA.

---

## 4. Coordinate re-verification

The deletion shifted every line after the hunk by **+20**. Every coordinate cited by this card's
SPEC and reports was therefore classified and re-checked.

### 4.1 The two sweeps, and why both are needed

```
$ grep -rnE '`:[0-9]+(-[0-9]+)?`' .moai/specs/SPEC-VACUOUS-FLOOR-GUARD-001/ .moai/reports/t378/
rc=1                                   # 0 hits

$ grep -rnoE '(line|lines|L)[ ]?[0-9]+(-[0-9]+)?' .moai/specs/… .moai/reports/t378/
… 95 hits (excluding the spec-lint capture file)
```

The batch's coordinate-shape sweep returns **0** here and the plain-prose sweep returns **95** —
this card writes coordinates as `line 28` / `L28` / `lines 131-136`, a grammar the shape sweep does
not match. The two sweeps miss different things and neither subsumes the other: the shape sweep
misses **continuation citations**, where a coordinate inherits its filename from the line above and
must be read with its surrounding context; the prose sweep catches those but misses nothing of that
kind and instead catches a **different coordinate grammar**. Both are batch discipline.

A note the lead asked to be recorded: the lead's own prose sweep on `develop` returns **0** while
this one returns 95. That is not a contradiction — this SPEC is unpushed and does not exist on
`develop`, so the two sweeps ran against different trees.

The prose pattern also produces a false positive worth naming: `base**line 2**` matches
`line[ ]?[0-9]+`. Three of the 95 hits are that, not coordinates.

### 4.2 Classification and results

Each coordinate is a **forward citation** ("where does this code live now" — must resolve at
`c3208b08f`) or a **historical anchor** ("what did this look like before" — must resolve at the
tree it names, and must NOT be renumbered).

| # | Location | Coordinate(s) | Kind | Resolves at | Result |
|---|---|---|---|---|---|
| 1 | `spec.md` §A.1 code block | 26 / 28 / 37 / 39 | historical | `3f03d9c36` — **stated in the text** | **VERIFIED** — 26 `recomputed :=`, 28 the equality, 37 `floor :=`, 39 the dead `if` |
| 2 | `spec.md`:41, :43 | line 39, line 40 | historical | `3f03d9c36`, inherits §A.1's label | **VERIFIED** |
| 3 | `spec.md` §A.3 table | 45 / 54 / 59 / 28 | **forward** ("enforced today") | should be `c3208b08f` | **DRIFTED** — at HEAD they are 65 / 74 / 79 / 28. See blocker B-1 |
| 4 | `spec.md`:109, :402 | `board_store.go` lines 131-136 | forward | `c3208b08f` | **VERIFIED** — `boardLockWaitMin` 131, `boardLockWaitMax` 132, `boardLockWaitStep` 136; `board_store.go` is untouched by this card |
| 5 | `plan.md` | (none) | — | — | no coordinates |
| 6 | `progress.md`:21-22 (AC-001/002 rows) | 120 / 122 / 28 / 65 / 74 / 79 | forward | `c3208b08f` | **VERIFIED** — all six exact |
| 7 | `progress.md` AC-003/004/005/006 rows | `:55` `:60` `:46` `:29` | historical | `226bdd0dc` — **named in the §E.2 header** | **VERIFIED** at that tree; see the mixed-anchor note below |
| 8 | `progress.md`:34 | original lines 32-44, original line 95 | historical | `3f03d9c36`, word "original" present | **VERIFIED** — hunk header is `@@ -32,13`, t372 func at base line 95 |
| 9 | `progress.md`:42 | `plan-audit.md` lines 71-72 | forward | `c3208b08f` | **VERIFIED** — both `-race` and `go test ./...` appear on 71 and 72 |
| 10 | `run-evidence.md`:11-12, :38 | 120 / 122 / 28 / 65 / 74 / 79 | forward | `c3208b08f` | **VERIFIED** |
| 11 | `run-evidence.md`:180-181 | original 32-44, original 95 | historical | `3f03d9c36`, labelled "original" | **VERIFIED** |
| 12 | `run-evidence.md`:219, :231 | `plan-audit.md` L165, lines 71-72 | quoted grep output / forward | `c3208b08f` | **VERIFIED** |
| 13 | `mutants.md`:16, :52, :108, :181 | line 54 / 59 / 45 / 28 | historical | `226bdd0dc` — header says "HEAD `226bdd0dc` … floor branch STILL PRESENT (pre-edit)" | **VERIFIED**, correctly pinned |
| 14 | `negative-evidence.md`:10 | lines 37-41 | historical | `226bdd0dc`, header states pre-edit | **VERIFIED** as the branch span; minor imprecision noted below |
| 15 | `negative-evidence.md`:67 + the `:40` citation | line 60, `board_lock_wait_test.go:40` | historical | `226bdd0dc` | **VERIFIED** |
| 16 | `census.md`:54, :62, :94, :129, :165, :168 | 54 / 28 / 59 / 45 / 28 / 28 | historical | `226bdd0dc`, header names the tree | **VERIFIED** |
| 17 | `census.md`:60 | t372 header comment lines 69 and 84 | historical | `226bdd0dc` | **VERIFIED** — both lines carry the cost-cancellation statement |
| 18 | `census.md`:176 | `integration_lock_cross_test.go` lines 54-55 | forward | `c3208b08f` | **VERIFIED** — the `10 × 33ms × 5 = 1.65s` comment sits at 54-55; file untouched |
| 19 | `plan-audit.md` (≈50 `L…` refs into `spec.md` / `plan.md`) | L127 … L359-363 | historical | `b2e3af7fb` — **declared in the report header** | **VERIFIED** — all seven REQ anchors (L127/133/138/144/149/158/165) resolve exactly at `b2e3af7fb`; at HEAD the REQs sit at 157/165/170/176/181/190/197 |

**Checked: 19 coordinate groups (95 raw prose hits, 3 of them the `baseline 2` false positive).
Drifted: 1 (row 3). Corrected: 0 — the drifted group lives in `spec.md` §A body, outside this
agent's write boundary; raised as blocker B-1.**

Row 19 is the clearest demonstration of why classification precedes updating. `plan-audit.md`'s
50 coordinates are all "wrong" at HEAD by a uniform +30, and renumbering them would be a
falsification: the audit reviewed `spec.md` v0.1.0 and its D1-D3 findings describe text that
subsequent commits changed. Pinned to `b2e3af7fb` they are exact.

**Mixed-anchor note (row 7).** `progress.md` §E.2 declares one tree in its header
("measurements taken at HEAD `226bdd0dc`") but its table mixes two: the AC-001/002 rows carry
**post-edit** coordinates from the working tree that became `c3208b08f`, while the AC-003..006 rows
carry **pre-edit** coordinates that resolve at `226bdd0dc` itself. Both sets are individually
correct; the single header is what makes them read as one anchor. This is an imprecision to record,
not a drift — no coordinate is wrong. Same shape in `run-evidence.md`, same header.

**Minor imprecision (row 14).** `negative-evidence.md` says "the verbatim landed code at lines
37-41" and then quotes seven lines, 35-41, the first two being the comment. The branch itself is
37-41; the quote is a superset. Recorded, not corrected.

---

## 5. Non-claims — what this verdict deliberately does NOT assert

1. **A deletion carries no positive mutant evidence of its own.** No mutant makes a removed branch
   fire. AC-VFG-001 and AC-VFG-007 are structurally one-sided and are reported as such; the
   evidence for a deletion is the absence of a firing, plus the static argument, plus the
   preserved GREEN of what remains.
2. **AC-VFG-007's negative evidence is one observation on one assignment.** The unreachability
   claim rests on the **static** argument (`floor` was the identical expression to `recomputed`,
   with `budget == recomputed` asserted twelve lines above under `t.Fatalf`, so `budget < floor`
   is false for every assignment of the three constants). The M2 observation corroborates on the
   single assignment `headroom = 1`; it does not prove the every-assignment claim, and a single
   observation of silence never could.
3. **"660ms is looser than 1650ms" is an illusion the vacuous branch created.** The old branch's
   `floor` evaluated to 1650ms only because it was the same expression as `budget` — it tracked
   every mutation down alongside the budget and could never be exceeded. It never functioned as a
   bound. The 660ms composed by the four retained assertions is the **first real floor this
   function has had**, and comparing the two numbers as if they were both bounds is the exact
   error this card exists to retire.
4. **AC-VFG-008 is not a clean zero.** The criterion asks for zero occurrences of `./...` and
   `-race` in the `go test` grep over `.moai/reports/t378/`. Read literally it **fails**: both
   tokens appear on `plan-audit.md` lines 71-72, inside the plan-auditor's own prose *stating the
   prohibition* and proposing this criterion's wording. All 12 recorded invocations comply — every
   one `go test -timeout 600s -count=1 [-v -run …] ./internal/kanban/`, serial, no race detector,
   no backgrounding. The substantive requirement holds; the literal phrasing is defeated by a
   document describing what is forbidden. Reported as a qualification, not as a zero.
5. **AC-VFG-006's form-A hole is documented, not closed.** Replacing the budget declaration with
   the bare literal `1650 * time.Millisecond` — numerically identical to the product, and precisely
   the REQ-BLB-001 regression the guard exists to catch — reddened **neither** guard. The assertion
   compares values, not syntax. Out of scope here; it needs a different kind of assertion, and it
   is follow-up F-1 below.
6. **A census prediction missed.** `census.md` predicted `TestConcurrencyStress` *may* redden under
   M2 because the budget falls to 330ms. It did not; two other budget-consuming tests reddened
   instead (`mutants.md` names and attributes each). The prediction was hedged with "may", so this
   is not a contradiction — it is recorded because a miss must stay visible rather than be quietly
   dropped once the AC passed.
7. **No CI verdict exists.** The branch is unpushed by instruction; no full-suite, no
   cross-platform, no `-race` result is available for any commit on it. CI owns the full-suite
   verdict per REQ-VFG-007.

---

## 6. Gaps — explicitly NOT observed

| Gap | Why |
|---|---|
| Full-suite `go test ./...` | prohibited by REQ-VFG-007 and by this repository's local-full-suite prohibition; CI-owned |
| `-race` on `internal/kanban` | prohibited for this card's verification by REQ-VFG-007 |
| Cross-platform build (darwin / linux / windows) | not run; one `_test.go` file, no build tags, no platform-conditional code |
| Coverage percentage | not measured; the change adds no executable line |
| `golangci-lint` | not run; `gofmt -l` (empty) and `go vet` (exit 0) were `[MD]` |
| CI verdict on any commit of this branch | branch unpushed by instruction |
| Every-assignment unreachability, dynamically | static argument only (§5.2) |
| Repository-wide sweep for sibling vacuous guards | SPEC-excluded (§D); inferring further instances from this one would be an unverified defect claim |
| Accuracy of the `integration_lock_cross_test.go` budget-arithmetic comment | observed to exist at lines 54-55, untouched, **accuracy unverified** (follow-up F-2) |

### Blockers

**B-1 — `spec.md` §A.3 table `Line` column has drifted, and this agent may not fix it.**
The column reads 45 / 54 / 59 / 28 under the header "Where it is actually enforced today"; at HEAD
`c3208b08f` those assertions are at 65 / 74 / 79 / 28. The `manager-docs` write boundary permits
only `status:` and `updated:` in `spec.md` frontmatter — §A body content is `manager-spec`'s.
Two acceptable repairs, either one sufficient: renumber the column to 65 / 74 / 79 / 28 (treating
it as the forward citation its header claims), **or** add a base qualifier — "(line numbers at base
`3f03d9c36`)" — pinning it as a historical anchor. The orchestrator should re-delegate to
`manager-spec` if the correction is wanted; nothing else in this card depends on it, and the SPEC's
argument is unaffected either way.

### Follow-up candidates — recorded, NOT fixed

**F-1 — the AC-VFG-006 form-A hole.** A value-comparing assertion cannot catch a numerically
identical literal substitution. `boardLockWaitBudget = 1650 * time.Millisecond` passes both guards
while destroying the derivability REQ-BLB-001 requires. Closing it needs a syntax-level or
constant-reference-level check, not a stronger comparison.

**F-2 — `internal/kanban/integration_lock_cross_test.go` lines 54-55** carry a comment restating the
budget arithmetic (`10 × 33ms × 5 = 1.65s`), including a nested `board_store.go:96-117` citation.
Observed during the M2 census, untouched by this card, **accuracy unverified**. It is prose, so it
reddens under no mutant — a documentation surface that tracks these constants with nothing keeping
it honest.

**F-3 — the rewritten comment states `660ms` and `10 * 33ms * 2` inline.** A future retune of
`boardLockCIMutationCost` or `boardLockHeadroom` leaves that prose stale while the assertions keep
working correctly — the same class of untested documentation surface as F-2, newly created by this
card and recorded rather than hidden.

---

## 7. Residual risk

1. **The comment is now the guard's only explanation, and comments do not fail.** The deletion
   traded a branch that could never fire for prose that can never fail. If a later reader
   reinstates a same-terms floor, nothing mechanical stops them — only the comment argues against
   it. That is an accepted, deliberate trade (REQ-VFG-004 exists precisely because of it), but it
   is a real reduction in mechanical enforcement, from one useless check to zero checks.
2. **No CI evidence.** Every measurement here is local, on one machine. `internal/kanban` carries
   contention tests whose behaviour differs materially under CI `-race` (t370 back-derived 42-105ms
   per-mutation against the declared 33ms). Nothing in this card touches that gap, and nothing here
   is evidence about it.
3. **The 660ms composed floor is an argument, not a measurement.** It is what the four retained
   assertions imply about the constants; it is not a claim that 660ms — or 1650ms — suffices on any
   real machine. Sufficiency remains explicitly out of scope (§D) and unevidenced.
4. **t372's observation window is open.** AC-SIV-013 requires ≥5 non-cancelled develop heads. This
   card's file shares `board_lock_wait_test.go` with that guard. The guard is byte-identical here,
   but any future edit to this file lands inside a measurement window that is still running.
5. **The `+30` shift this card's own commits introduced into `spec.md`** means any external
   document citing `spec.md` line numbers from before `226bdd0dc` is now stale. `plan-audit.md` is
   correctly pinned; anything outside this card's directory has not been swept.

---

## 8. Status transition and artifact changes

| Artifact | Change | Authority |
|---|---|---|
| `spec.md` frontmatter | `status: in-progress → implemented`; `updated: 2026-08-31` (unchanged, same date) | `manager-docs` owns `status:` / `updated:` only |
| `progress.md` §E.3 | `run_commit_sha: pending-backfill-t378-run → c3208b08f` | mechanical placeholder backfill, directed by the orchestrator; no other §E.3 field touched |
| `progress.md` §E.4 | populated (was `_<pending sync-phase>_`) | `manager-docs` owns §E.4 |
| `CHANGELOG.md` | one entry under `### Changed`, newest-first position | `manager-docs` owns `[Unreleased]` |
| `.moai/reports/t378/verdict.md` | this file, new | `manager-docs` |
| `spec.md` §A-§H body, `plan.md`, `progress.md` §E.1/§E.2 | **untouched** | outside the write boundary |

`status` closes at **`implemented`**, not `completed`. The branch is unpushed, no CI verdict exists
for any commit on it, and the integration window is the orchestrator's next step — the conditions
for `completed` are not met and claiming them would be an unobserved completion claim.

## 9. CHANGELOG decision

**Emitted, under `### Changed`.** Rationale, and the counter-argument considered:

- The sibling precedent is direct. `SPEC-STRESS-INVARIANT-VERDICT-001` (card t372) is also
  test-only, also confined to `internal/kanban`, and carries an entry under `### Changed`
  (`CHANGELOG.md`:191). This card repairs the exact defect that entry names as a deferred
  follow-up, so omitting it would leave the predecessor's "recorded as a follow-up" pointing at
  nothing.
- This repository's `[Unreleased]` logs every SPEC sync-phase close, test-only ones included.
  Omission would be the anomaly.
- The entry carries reader value beyond the diff: it retires a check that read as coverage, and it
  corrects the "1650ms floor" illusion that a reader of the old code would otherwise carry forward.
- **Counter-argument considered and rejected**: no user-visible behaviour changed, and a strict
  reading of Keep-a-Changelog ("notable changes") could exclude a test-only deletion. Rejected
  because this repository's established practice treats the SPEC close as the notable unit, and
  because a silently-deleted guard is precisely the kind of change a future reader needs to find.

Position: first entry of `### Changed` (the section is newest-first — t372 currently occupies that
slot).

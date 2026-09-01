---
id: SPEC-MEMORY-STORE-RECONCILE-001
title: "Acceptance criteria — auto-memory store reconciliation and index-budget premise correction"
version: "0.3.0"
created: 2026-08-31
---

# Acceptance criteria — SPEC-MEMORY-STORE-RECONCILE-001

Sixteen criteria (Tier M ceiling 16). §D.1 marks which are machine-verifiable and how the rest are
judged. Every criterion is binary.

[HARD] Four conventions bind every criterion below. Each was added after a plan-audit iteration
found a criterion that passed while measuring nothing — the list is a record of that failure class,
not a style guide.

- **No `moai memory doctor` invocation without `--dir`.** The bare form resolves a worktree-slugged
  store that does not exist and returns `exists:false, index_lines:0, findings:null` — a count of 0
  that measures nothing (measurements.md §M5b). Every criterion reading a count first asserts
  `exists: true` **and** `index_lines > 0` in the same JSON.
- **Every extraction names its selector verbatim, and the selector's case is checked.** The doctor
  JSON is case-asymmetric: store fields are lower-case (`exists`, `index_lines`) while the finding
  field is capitalised (`Code`). `.code` and `.Code` both parse; measured against the live store
  they return **0** and **58**. A criterion that says "read the dangling count" without the selector
  is satisfiable by the wrong one.
- **Every count criterion carries a red direction.** A gate proving the subject exists does not
  prove the measurement reached it; only a non-zero BEFORE reading does. Where a criterion asserts
  an after-state of 0, it also asserts a before-state of ≥ 1.
- **No live-mutating literal is pinned.** The store has concurrent writers — measured three times
  on one day, with the entry count moving on the third (AC-MSR-010) — so counts are compared as
  deltas and every literal is a dated reference (R4, `verification-claim-integrity.md` §2.1).

Throughout, `$D` is `"$HOME/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory"`.

## §D AC matrix

### AC-MSR-001 — every in-repo assertion of the cut is gone, counted whole-file

_maps REQ-MSR-001_

**Given** the four surfaces after M1,
**When**
```bash
grep -c '25KB' .claude/rules/moai/workflow/moai-memory.md
grep -c '25KB\|200-line' .claude/output-styles/moai/moai.md
grep -c '25KB\|25 \* 1024\|25600' internal/config/token_budget_guard.go
```
is run,
**Then** each count equals **0**. No occurrence is deliberately retained; the enumerated set at
plan time was `moai-memory.md:29,117,175`, `moai.md:165`, `token_budget_guard.go:95,98`, and all
six are rewritten or removed. A whole-file count is used rather than a phrase grep because
iteration 1's phrase grep matched only line 117 and passed while lines 29 and 175 kept asserting
the same cut in the same file.

### AC-MSR-002 — the doctrine carries the command and carries no bare cap value

_maps REQ-MSR-001_

**Given** the rewritten sections in **both** doctrine copies,
**When** each is read,
**Then** a conjunction of two conditions holds:

1. **Positive (cannot be satisfied by absence)** — the section **contains** the re-measuring
   command: `grep -c 'moai memory doctor' <file>` is **≥ 1** in each copy.
2. **Negative** — the section contains **zero bare numeric cap values**: no figure asserting a
   loading limit appears in either copy. Reference values live in `spec.md` §A.2 / §A.2.1 and in
   `.moai/reports/t383/`, never in doctrine (spec.md §A.3).

**Why this is a conjunction and not a universal.** The iteration-2 form was "every numeric value
carries its date and the word 'reference'". Under §A.3's single-form resolution the doctrine copies
carry **no numeric values at all**, so that universal is satisfied **vacuously over an empty set** —
the SPEC's own named failure mode, occurring inside the criterion written to enforce R4 against it.
Condition 1 is the repair: a file emptied of subject matter fails it, because the command must be
*present*. Condition 2 alone would have the same vacuity; the two together do not.

Condition 1 is mechanical. Condition 2 is judged by reading — "a figure asserting a loading limit"
is a meaning, not a regex (a version number or a line reference is not a cap value). §D.1 records
the split.

### AC-MSR-003 — the compression rule is stated

_maps REQ-MSR-002_

**Given** the rewritten doctrine,
**When** `grep -n 'shorter' .claude/rules/moai/workflow/moai-memory.md` is run,
**Then** a line matches stating that compression means shortening entry lines and that removing an
entry makes its topic file unreadable to future sessions.

### AC-MSR-004 — the two-store hazard, its detector, and the `--dir` requirement are recorded

_maps REQ-MSR-003_

**Given** the rewritten doctrine,
**When** `grep -n 'moai memory doctor' .claude/rules/moai/workflow/moai-memory.md` is run,
**Then** at least one line matches, and the surrounding text (a) names both candidate stores as a
hazard invisible from the index alone, and (b) states that a count must be read only from an
invocation carrying `--dir` whose JSON reports `exists: true`.

### AC-MSR-005 — the incident's branch is closed, in a surface the writing session loads

_maps REQ-MSR-004_

**Given** `.claude/rules/moai/core/moai-constitution.md` § Lessons Protocol after M1,
**When** its frontmatter is checked (`head -1` is not `---`, so the file is always-loaded) and the
inserted clause is read,
**Then** the clause directs a session at or beyond the reference index size to write both the topic
file and its index line, offers no branch in which the lesson is dropped, and names the store by the
resolving command rather than a literal path. The frontmatter half is mechanical; the "offers no
branch" half is *judged by reading* (§D.1).

### AC-MSR-006 — the vacuous slot is gone from the code, with one named tombstone

_maps REQ-MSR-005, REQ-MSR-008_

**Given** `internal/config/token_budget_guard.go` after M2,
**When**
```bash
grep -n 'memoryHead\|memoryHeadByteCap\|memoryHeadLineCap' internal/config/token_budget_guard.go
grep -n 'MEMORY.md' internal/config/token_budget_guard.go
```
is run,
**Then** the first exits non-zero (no match), and the second matches **only** the single tombstone
comment block M2 leaves in place, which is named verbatim in the verdict. Scoping the second grep
this way is deliberate: a whole-file ban on the string would delete the hermeticity rationale too,
leaving the next author no in-file record of why there is no slot — and re-adding it is exactly the
regression the tombstone prevents.

### AC-MSR-007 — the fixed-slot existence test goes red on the mutation

_maps REQ-MSR-006_

**Given** the new fixed-slot existence test after M2, which resolves its root via
`findRepoRoot(mustGetwd(t))` with the existing `t.Skip` guard and asserts only against the real tree,
**When** the `MEMORY.md` slot is re-inserted into `alwaysLoadedSurface` and
`go test ./internal/config/...` is run,
**Then** the run FAILS and the failure output names the re-inserted slot. The mutation is reverted
and the same command PASSES. Both outputs are recorded verbatim.

### AC-MSR-008 — the edit changed the surface by exactly one element

_maps REQ-MSR-007_

**Given** the enumerated surface **list** recorded at M2 step 1 and again at step 6,
**When** the two lists are diffed,
**Then** the diff is exactly one removed line, `<repoRoot>/MEMORY.md`, and nothing else — no
addition, no reordering. The token total is reported alongside and must change by exactly the
tokens that element contributed. Diffing the list rather than asserting "the total does not move"
is deliberate: an absent file already contributes 0, so a no-move assertion is true by construction
and would falsify nothing.

### AC-MSR-009 — every index target resolves, measured against a store that exists

_maps REQ-MSR-009_

**Given** `reconcile-before.json` and `reconcile-after.json`, each produced by
`moai memory doctor --json --dir "$D"`,
**When** each is read with the **verbatim** extraction below,
**Then** all three conditions hold, evaluated in order:

```bash
# (1) store gate — read no count until this passes
jq -e '.[0] | select(.exists == true and .index_lines > 0)' "$F" > /dev/null

# (2) the count. NOTE THE CAPITAL C — see the case trap below.
jq '[.[0].findings[]? | select(.Code=="MEMORY_DANGLING_INDEX_LINK")] | length' "$F"
```

1. **Gate** — both files pass `(1)`. Either failing FAILS the criterion and no count is read.
2. **Red direction** — the BEFORE count is **≥ 1**. A before-count of 0 FAILS: it means the
   selector matched nothing, not that the store was healthy.
3. **Green direction** — the AFTER count is **0**.

**The case trap, measured.** The JSON is case-asymmetric: the store fields are lower-case
(`exists`, `index_lines`, `topic_files`) while the finding field is capitalised (`Code`). Both
selectors below run without error against the live store and return different answers:

```
jq '[.[0].findings[]? | select(.code=="MEMORY_DANGLING_INDEX_LINK")] | length'   → 0
jq '[.[0].findings[]? | select(.Code=="MEMORY_DANGLING_INDEX_LINK")] | length'   → 58
```

The `exists` gate cannot catch a `.code` selector, because the store genuinely exists and reports
`index_lines: 164` — the reading *looks* sound, which is worse than the `--dir` defect it replaced.
Condition 2 is what catches it: a wrong selector yields 0 on the BEFORE reading and fails there,
rather than yielding 0 on the AFTER reading and passing.

### AC-MSR-010 — the index did not lose entries, measured as a delta

_maps REQ-MSR-011, REQ-MSR-012, REQ-MSR-013_

**Given** **two** reachability metrics plus the byte count, captured by the same command inside the
M3 window, before and after:

```bash
grep -c '^- \[' "$D/MEMORY.md"                                        # (a) entry lines
grep -o '](\([^)]*\.md\))' "$D/MEMORY.md" | sort -u | wc -l           # (b) unique targets, file-wide
wc -c < "$D/MEMORY.md"                                                # (c) bytes — reference only
```

**When** the pairs are compared,
**Then** **both** `entry_after >= entry_before` **and** `targets_after >= targets_before`.
**A decrease in either is a FAIL, never a saving.** An increase is PERMITTED and is attributed: any
growth must correspond to entries a concurrent session added (spec.md C2 — REQ-MSR-004's own new
guidance tells sessions to add index lines), and M3 edits the index not at all.

**Why metric (b) is mandatory, measured.** Metric (a) sees only the targets on `^- \[` lines;
**the rest, including 14 of the dangling ones, sit on line shapes the anchor does not match**
(§A.2.1). Dated reference, 2026-08-31 run-phase M0 — 141 entry-line targets against 196 file-wide,
a 55-target blind spot; re-measure with `.moai/reports/t383/measure-n1.sh` rather than quoting it,
because the earlier revision's literals (124 / 135 / 190) had already rotted to 130 / 141 / 196 by
run-phase entry while the 14 held exactly. Deleting one grouped line carrying six links therefore
leaves (a) unmoved, and the lead's [HARD] no-fewer-entries rule would be honoured for the lines it
can see and silently blind to the rest. §A.2.2 shows this is not hypothetical: three such lines
carry 6 live lesson targets and sit in the exact section dispatch candidate (a) proposed to remove.

Metric (b) has its own blind spot and it is named rather than hidden: it counts **unique** targets,
so deleting a duplicate reference does not move it. Dedup is worth 2 across the whole file, so the
residual is two targets — a far smaller gap than (a)'s 55, and stated so the next author does not
mistake (b) for complete coverage.

Dated reference values, and the reason they are references rather than the criterion — measured
**four** times on 2026-08-31 in this worktree, hours apart, with nobody here touching the file:

| Reading | bytes | entry lines | unique targets |
|---|---|---|---|
| measurements.md M1 | 26,280 | 123 | 189 |
| iteration 2 re-measure | 26,290 | 123 | — |
| iteration 3 re-measure | 26,577 | 124 | 190 |
| **run-phase M0** | **28,009** | **130** | **196** |

Every size metric moved at every reading, and from the third onward the **entry count moved too** —
the one figure iteration 2 believed was stable enough to pin at 123. Had this criterion pinned
`entry == 123`, it would have been failing since before run-phase began, for work no one in this
card performed. That is the whole argument for the delta form, and it is an observation rather than
a projection.

**The defect figures did not move — now across four readings.** Over the same interval, 58 dangling
/ 44 entry-line-reachable / 14 outside-only / 40 affected lines held **exactly**, re-confirmed at
run-phase M0 by both `moai memory doctor` (58) and `measure-n1.sh` (58 / 44 / 14 / 40). That
asymmetry is why AC-MSR-009 may target an exact 0 while this criterion may not pin any literal.

### AC-MSR-011 — the legacy store is untouched and nothing is overwritten

_maps REQ-MSR-010_

**Given** the legacy store before and after M3,
**When** its file count and the mtimes of the copied source files are compared,
**Then** the count is unchanged (dated reference: 1,098) and no source file's mtime moved.
The copy step's log additionally shows zero overwrites of pre-existing active-store files.

### AC-MSR-012 — no store file is a commit target

_maps REQ-MSR-015_

**Given** the worktree at close,
**When** `git status --short` and `git show --stat HEAD` are read,
**Then** no path under either memory store appears, and every changed path lies within this
permitted set — note that the SPEC directory is enumerated **file by file**, never as a prefix:

```
.moai/specs/SPEC-MEMORY-STORE-RECONCILE-001/spec.md
.moai/specs/SPEC-MEMORY-STORE-RECONCILE-001/plan.md
.moai/specs/SPEC-MEMORY-STORE-RECONCILE-001/acceptance.md
.moai/specs/SPEC-MEMORY-STORE-RECONCILE-001/progress.md
.moai/reports/t383/
.claude/rules/moai/workflow/moai-memory.md
.claude/rules/moai/core/moai-constitution.md
.claude/output-styles/moai/moai.md
internal/template/templates/.claude/rules/moai/workflow/moai-memory.md
internal/template/templates/.claude/rules/moai/core/moai-constitution.md
internal/template/templates/.claude/output-styles/moai/moai.md
internal/config/
```

The three `internal/template/templates/` entries are load-bearing: iteration 1's list omitted them,
which made the mandatory Template-First mirror a criterion violation.

**Why the SPEC directory is enumerated per file, measured.** The iteration-3 list permitted the
directory as a **prefix**. That directory has since accumulated session runtime state, and
`git status --short` collapses the whole tree into a single untracked entry, so a prefix pathspec
sweeps every one of these in:

```
$ git status --short
?? .moai/specs/SPEC-MEMORY-STORE-RECONCILE-001/     <-- one entry, whole tree
$ find .moai/specs/SPEC-MEMORY-STORE-RECONCILE-001 -mindepth 1 -not -name '*.md'
.moai/specs/SPEC-MEMORY-STORE-RECONCILE-001/.moai/state/config-cache.json
.moai/specs/SPEC-MEMORY-STORE-RECONCILE-001/.moai/state/context-usage/40f2779d-….json
.moai/specs/SPEC-MEMORY-STORE-RECONCILE-001/.claude/agent-memory/plan-auditor/
```

The `context-usage` file carries the **run-phase session's own uuid** — committing it would put
this session's private runtime state into the repository, and the collapsed `git status` line is
exactly what hides that from a reader checking before staging.

[HARD] Two consequences bind every commit in this card: stage by **explicit per-file pathspec**
— never the directory, never `git add -A`, never `git add .` — and re-read `git status --short`
in the same tool call that stages, so a file another actor added between the check and the stage
is visible rather than swept.

### AC-MSR-013 — every mirrored file is byte-identical again

_maps REQ-MSR-014_

**Given** the three local/mirror pairs after M1 and `make build`,
**When**
```bash
diff -q .claude/rules/moai/workflow/moai-memory.md internal/template/templates/.claude/rules/moai/workflow/moai-memory.md
diff -q .claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md
diff -q .claude/output-styles/moai/moai.md internal/template/templates/.claude/output-styles/moai/moai.md
```
is run,
**Then** all three exit 0. All three were verified rc=0 at plan time, so a non-zero here is a
regression this card introduced, not a pre-existing condition.

**No exception, and that is the iteration-3 correction.** The iteration-2 wording admitted a
fallback in which the local file could carry a machine-specific form and the pair "differ only
inside the fenced R4 block". That fallback is withdrawn: it contradicted byte-identity outright,
and resolving the contradiction toward the mirror form is precisely what emptied AC-MSR-002 of
subject matter (§A.3). There is now **one form, written into both copies** (spec.md §A.3), so
byte-identity is unconditional here and this criterion admits no substitution.

### AC-MSR-014 — the mirror carries no neutrality-forbidden content

_maps REQ-MSR-014_

**Given** the three mirrored files after M1,
**When**
```bash
# (1) existence gate FIRST — a missing file must not discharge the criterion
for f in internal/template/templates/.claude/rules/moai/workflow/moai-memory.md \
         internal/template/templates/.claude/rules/moai/core/moai-constitution.md \
         internal/template/templates/.claude/output-styles/moai/moai.md; do
  test -f "$f" || { echo "MISSING $f"; exit 1; }
done

# (2) the scan — exit code must be EXACTLY 1
grep -n 'goos\|claude-profiles\|SPEC-MEMORY-STORE-RECONCILE\|2026-08-31' \
  internal/template/templates/.claude/rules/moai/workflow/moai-memory.md \
  internal/template/templates/.claude/rules/moai/core/moai-constitution.md \
  internal/template/templates/.claude/output-styles/moai/moai.md
echo "exit=$?"
```
is run,
**Then** step (1) passes for all three files **and** step (2) reports `exit=1` — **exactly 1, not
merely non-zero**.

**Why exactly 1, measured.** `grep` exits **2** on file-not-found, so a `non-zero` predicate is
satisfied by a deleted or renamed mirror — the criterion would pass hardest at the moment the
mirror stopped existing. Reproduced in this worktree:

```
$ grep -n 'nonexistent-pattern-xyz' <path>/no-such-file-here.md
grep: <path>/no-such-file-here.md: No such file or directory
exit=2
```

**Red direction (required, demonstrated).** Copy one mirror to a scratch path, plant the literal
`claude-profiles` in the copy, run step (2) against the copy, and show `exit=0`. Discard the
scratch copy. Without this the criterion is a one-directional check whose green is uninterpreted.

### AC-MSR-015 — the M0 sample record exists and its gate was applied

_maps REQ-MSR-009_

**Given** `.moai/reports/t383/m0-sample.md`,
**When** it is read,
**Then** it contains all four of:

1. **The 12 selected target paths and the rule that produced them** — every 5th of the **58**
   unique missing targets, sorted lexicographically: indices 1, 6, 11, … 56 → exactly 12 files.
2. **A per-file verdict** of `live` or `superseded`, each with its deciding evidence — a
   `[SUPERSEDED by …]` marker quoted from the file, or the named subsuming active-store file.
3. **The superseded count** and its explicit comparison against the threshold: **≥ 4 of 12
   proceeds no further** and returns a blocker. A record whose count is ≥ 4 and which nonetheless
   proceeded FAILS this criterion.
4. **The coverage statement** — which of the 58 the sample could not reach, named as a limit rather
   than left silent.

**Two corrections from iteration 2, both arithmetic and both load-bearing.** The population was
44; it is **58**, because AC-MSR-009 requires `moai memory doctor`'s whole-file count to reach 0
and that needs all 58 copied — sampling from 44 left 14 targets structurally unsamplable. And the
old rule "every 4th up to n=10" over 44 yields indices 1…37, so entries 38-44 were unreachable even
within its own population. Every 5th of 58 reaches index 56, leaving only 57-58 outside the sample
— stated here per item 4 rather than discovered later.

The threshold moves 3-of-10 → 4-of-12 to hold the same ~30% share over the larger sample, rounded
away from proceeding.

Without this criterion the M0 gate could be skipped entirely while every other criterion passed.

### AC-MSR-016 — the lint gate is judged by the tree's own binary

_maps REQ-MSR-014_

**Given** the tree after `make build`,
**When**
```bash
# (1) working directory anchor — MANDATORY, see the cwd trap below.
# ASSERT the cwd; do not cd to it. `cd "$(git rev-parse --show-toplevel)"` is the
# obvious form and it is REFUSED by the worktree-isolation guard in a worktree
# session (measured — see the anchor-form note below), so the criterion would be
# un-runnable exactly where this card executes.
test -f go.mod || { echo "NOT AT REPO ROOT — cd there and re-run"; exit 1; }

# (2) binary gate — the tree's own build, by path
test -x ./bin/moai || { echo "NO BINARY — run make build"; exit 1; }

# (3) the run. Allow generous time: a full-corpus lint took > 120 s here.
./bin/moai spec lint > /tmp/t383-lint.txt 2>&1
echo "lint_exit=$?"                       # the LINT's own exit code, not a pipeline's

# (4) LIVENESS — a lint that produced no output must FAIL, not pass
test -s /tmp/t383-lint.txt || { echo "EMPTY LINT OUTPUT — criterion FAILS"; exit 1; }
grep -qE '[0-9]+ error\(s\), [0-9]+ warning\(s\)' /tmp/t383-lint.txt \
  || { echo "NO SUMMARY LINE — lint did not complete; criterion FAILS"; exit 1; }

# (5) the criterion itself
grep -c 'SPEC-MEMORY-STORE-RECONCILE-001' /tmp/t383-lint.txt
```
is run,
**Then** steps (1)-(4) all pass **and** the `grep -c` count for this SPEC ID is **0**.

**Why steps (1) and (4) exist — both misfires reproduced in this worktree at run-phase.**
The iteration-3 form of this criterion was `grep -c '<SPEC-ID>' == 0` with no anchor and no
liveness check. It misfires in **both** directions, and neither is hypothetical:

*Spurious FAIL — the cwd trap.* The criterion never stated a working directory. Run from
inside the SPEC directory, `moai spec lint` resolves that directory as its specs root and
reports six `SpecsDirForeignEntry` warnings naming this very SPEC:

```
$ cd .moai/specs/SPEC-MEMORY-STORE-RECONCILE-001 && moai spec lint > /tmp/x 2>&1; echo $?
0
$ grep -c 'SPEC-MEMORY-STORE-RECONCILE-001' /tmp/x
6
```

Same binary, same tree, opposite verdict from the worktree root. This is the identical
`os.Getwd()` hazard the SPEC already made a [HARD] rule about for `moai memory doctor`
(§M5b) — generalised here to `spec lint`, where it had been left unstated. Step (1) closes it.

*Vacuous PASS — the empty-output trap.* `grep -c` on an empty file returns 0 and exits 1, so
**a lint that never ran satisfies the criterion**:

```
$ : > /tmp/empty.txt; grep -c 'SPEC-MEMORY-STORE-RECONCILE-001' /tmp/empty.txt
0
$ test -s /tmp/empty.txt; echo $?
1
```

**The anchor form itself was a planted defect, caught by running it.** The first version of
this repair wrote the anchor as `cd "$(git rev-parse --show-toplevel)"`. That command is
**refused** by the worktree-isolation guard in a worktree-isolated session — the guard cannot
statically verify that a command-substituted `cd` stays inside the worktree, so it declines to
run it at all:

```
This session is isolated in the worktree .../.claude/worktrees/t383, but this command is
too complex to verify that it stays inside the worktree. Refusing to run it
```

So the criterion written to close a cwd hole was itself un-runnable in the very session that
must run it. The corrected form **asserts** the working directory instead of changing it, which
is both guard-compatible and strictly stronger: `cd` silently relocates and proceeds, whereas
`test -f go.mod` fails loudly when the caller is somewhere unexpected — the failure mode this
criterion exists to produce.

Nothing else catches this, because `lint_exit` is deliberately *not* asserted (below). Step (4)
is the only guard between a crashed or never-invoked lint and a green criterion, and it is
therefore load-bearing rather than defensive: it asserts the output is non-empty **and**
carries the run's own `N error(s), M warning(s)` summary line, so a truncated or
partially-written file also fails.

`lint_exit` is **recorded, not asserted**: the corpus carries pre-existing findings (8 errors /
64 warnings at plan time on the stale build), so a non-zero lint exit is the corpus baseline rather
than a verdict on this SPEC. The per-SPEC count is the criterion; the exit code is context, and
recording it is what stops a future reader from re-deriving "0 findings" from a command that never
ran.

That choice is correct and it is also what **creates** the vacuity hole step (4) closes, so the
two are stated together deliberately. Asserting `lint_exit == 0` would make this card hostage to
other SPECs' defects; not asserting it leaves nothing at all watching for a lint that produced no
output. Step (4) is the replacement guard — it asserts the run *happened* without asserting it
*passed*, which is exactly the separation this criterion needs.

**A `PATH` invocation does not satisfy this criterion.** The installed binary is 20 commits behind
this tree for `internal/spec/` (plan.md §I gap G2), so it applies rules this tree may not carry.
Plan-phase did run the PATH form and observed 0 — recorded as weak evidence from a stale build,
never as this criterion's pass.

**Red direction.** The count is falsifiable from a real baseline: the same command at plan time
returned non-zero counts for other SPECs in the same corpus (e.g. `SPEC-COVERAGE-RULE-SCOPE-001`
carried 6 `CoverageIncomplete` errors), so a `grep -c` of 0 for this SPEC ID distinguishes a clean
SPEC from a corpus where the rule fires.

## §D.1 Verifiability

| AC | Machine-verifiable | How judged if not |
|---|---|---|
| AC-MSR-001 | yes — three `grep -c` counts | — |
| AC-MSR-002 | partly | Condition 1 (the section CONTAINS the command) is mechanical — `grep -c 'moai memory doctor' ≥ 1`. Condition 2 (zero bare numeric cap values) is judged by reading: "a figure asserting a loading limit" is a meaning, not a regex. The mechanical half is the one that cannot be satisfied by an empty file, which is the point. |
| AC-MSR-003 | yes — grep | — |
| AC-MSR-004 | partly | The grep is mechanical; whether the surrounding text names the hazard is judged by reading. |
| AC-MSR-005 | partly | The frontmatter check is mechanical (`head -1` is not `---`). Whether the text "offers no branch in which the lesson is dropped" is a reading, judged against the motivating incident: a reader in lane-8's position must find an instruction, not a dilemma. |
| AC-MSR-006 | yes — two greps, one scoped | — |
| AC-MSR-007 | yes — `go test` exit status, both directions | — |
| AC-MSR-008 | yes — list diff | — |
| AC-MSR-009 | yes — JSON gate, then before-count ≥ 1, then after-count = 0 | — |
| AC-MSR-010 | yes — two independent metric pairs, compared as deltas | — |
| AC-MSR-011 | yes — file count + mtime | — |
| AC-MSR-012 | yes — `git status --short` | — |
| AC-MSR-013 | yes — three `diff -q` exit codes | — |
| AC-MSR-014 | yes — `test -f` gate, then grep exit code **exactly 1**, plus a planted red | — |
| AC-MSR-015 | partly | Existence and the threshold comparison are mechanical; each per-file `live`/`superseded` verdict rests on the §F M0 decision rule, whose second arm ("description subsumed by a named active file") is a judgement recorded with its evidence so another reader can dispute it. |
| AC-MSR-016 | yes — cwd anchor, `test -x` gate, liveness assertion (`test -s` + summary line), then `grep -c` of the SPEC ID | — |

Twelve of sixteen are fully mechanical; the remaining four (AC-MSR-002, -004, -005, -015) are
**partly** so — each carries a mechanical half that cannot be satisfied by absence, plus a reading
that is named rather than dressed in a grep. None is wholly unverifiable.

That split is itself an iteration-3 correction. AC-MSR-002 was marked fully non-mechanical, which
is what let it be restated into a universal that an empty file satisfies; giving it a mechanical
positive half fixed both the vacuity and the classification. The general lesson, recorded because
it recurred: **a grep for a phrase passes on a sentence that says the opposite, and a universal
over an empty set passes on a file that says nothing at all.** Iteration 1's AC-MSR-001 failed the
first way; iteration 2's AC-MSR-002 failed the second.

## §D.2 Traceability

Every requirement in spec.md §B is covered by at least one criterion above.

| REQ | Covered by |
|---|---|
| REQ-MSR-001 | AC-MSR-001, AC-MSR-002 |
| REQ-MSR-002 | AC-MSR-003 |
| REQ-MSR-003 | AC-MSR-004 |
| REQ-MSR-004 | AC-MSR-005 |
| REQ-MSR-005 | AC-MSR-006 |
| REQ-MSR-006 | AC-MSR-007 |
| REQ-MSR-007 | AC-MSR-008 |
| REQ-MSR-008 | AC-MSR-006 |
| REQ-MSR-009 | AC-MSR-009, AC-MSR-015 |
| REQ-MSR-010 | AC-MSR-011 |
| REQ-MSR-011 | AC-MSR-010 |
| REQ-MSR-012 | AC-MSR-010 |
| REQ-MSR-013 | AC-MSR-010 |
| REQ-MSR-014 | AC-MSR-013, AC-MSR-014, AC-MSR-016 |
| REQ-MSR-015 | AC-MSR-012 |

REQ-MSR-012 and REQ-MSR-013 are decisions rather than build steps; AC-MSR-010's "no entry lost"
delta is what mechanically holds both, since dieting or restoring markers would edit index lines.

## §D.3 Definition of Done

All sixteen criteria PASS, and the M4 verdict at `.moai/reports/t383/verdict.md` carries the
five-section evidence format. Its Gaps section names the deferred items (46 orphans, the 177-vs-50
topic-file cap, the legacy store's 1,098 files, the loader's cap shape, the part of the belief
outside this repository) **and** the two recorded gaps G2 and G4 from plan.md §I, so neither is
mistaken for resolved.

## §D.4 Quality gates

- `go test ./internal/config/...` PASSES (AC-MSR-007 green direction).
- `go vet ./internal/config/...` clean.
- `./bin/moai spec lint` invoked by path after `make build`, its own exit code read — AC-MSR-016.
  A PATH invocation does not discharge this gate (plan.md §I gap G2).

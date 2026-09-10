# Plan Audit — SPEC-QUEUE-UPGRADE-PROOF-001 (card t470) — iteration 2

Auditor: `plan-auditor` · Iteration 2/2 (Tier M ceiling) · Date 2026-09-03

Tree: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t470`, branch
`WT-queue-upgrade-proof`, HEAD `37949cd5fd635dca069b6f711aae994bb0c9e311`,
SPEC at `v0.2.0`, SPEC base `4e4607abe`.

Reasoning context ignored per M1 Context Isolation. The author's change report
in the dispatch was treated as a CLAIM to be falsified, not as evidence; every
item below was re-derived from the artifacts and the tree.

Scope: defect delta from iteration 1 plus regression, per the Retry Loop
Contract's iteration-2 clause. Iteration 1's verification of the SPEC's
measured facts is carried forward and was not re-measured, except for the
specific citations the v0.2.0 edits touch (listed under Evidence).

---

## Verdict

| | |
|---|---|
| **Verdict** | **FAIL** |
| Aggregate quality score | **0.9625** (iter-1 `0.875`, delta **+0.0875**, monotonic UP) |
| Tier M PASS threshold | 0.80 — the score is above threshold |
| Cause of FAIL | **MP-7 alone.** One must-pass firewall failure, failed-by-design, not compensable by score. |
| Blocking defects | D11 (new) |
| Optional defects | D12 (new) |
| Resolved since iter-1 | D1, D2, D3, D4, D6, D7, D8, D9, D10 — 9 of 10 |
| Deliberately unresolved | D5 (MP-7) — correctly the dispatcher's, not the author's |

The verdict and the score again disagree, and again both are reported. On
substance the remediation is strong: the load-bearing fix (D3) is correct, and
I traced its claimed call chain independently rather than accepting it. The
FAIL is carried entirely by the open clarification gate, which is
score-independent by construction and which the author was right not to
"fix" by editing the markers away.

**Score is monotonic** (`0.875 → 0.9625`). It did not drop. The one new
blocking defect (D11) is bounded to a single acceptance-criterion limb and is
outweighed by the frontmatter and traceability dimensions moving to 1.0.

---

## Must-Pass Results

| ID | Criterion | State | Evidence |
|---|---|---|---|
| MP-1 | REQ number consistency | **PASS** | `grep -n '^### REQ-' spec.md` → `REQ-QUP-001`(L93) `002`(L100) `003`(L107) `004`(L113) `005`(L118) `006`(L124) `007`(L131) `008`(L138) `009`(L143) `010`(L149). 10 entries, sequential 001-010, no gap, no duplicate, uniform 3-digit padding. The new `REQ-QUP-010` extends the run rather than interrupting it. |
| MP-2 | GEARS format compliance | **PASS** | Judged against the **requirement layer** (`REQ-XXX` in `spec.md`) ONLY; ACs graded under Group 4 per M3 § Scope. 001-004 event-driven (`**When** … shall`); 005/006/008 ubiquitous; 007 `**Where** … shall`; 009 and the new 010 unwanted (`shall not change` / `shall not be accepted`). 10/10 match a GEARS pattern. |
| MP-3 | YAML frontmatter validity | **PASS** (was FAIL) | Verified by PARSING, not reading — see § MP-3 measurement below. All 12 canonical fields present with correct types; `tags` is a string; `lifecycle: spec-anchored` is in the SSOT enum; `priority: High` is in the enum; `created`/`updated` parse as ISO dates. |
| MP-4 | §22 language neutrality | **N/A (auto-pass)** | Unchanged from iter-1. Single-language scope (`module: internal/kanban`, a Go package in this repository). No multi-language tooling surface named. |
| MP-5 | D7 cross-SPEC reconciliation | **PASS** | Verb re-executed: `grep -Eoh 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md acceptance.md plan.md \| sort -u` → `SPEC-QUEUE-UPGRADE-PROOF-001` only (self). No external SPEC referenced, so no retired/superseded/archived status to reconcile. No BLOCKING finding. |
| MP-6 | D8 cross-platform discipline | **PASS (auto)** | `grep -c syscall spec.md acceptance.md plan.md` → `0`, `0`, `0`. D8-4 auto-PASS. |
| MP-7 | Clarification gate | **FAIL — by design** | `grep -rn '\[NEEDS CLARIFICATION' plan.md` → `plan.md:10` (`G2 definition`), `plan.md:22` (`downgrade intent vs quarantine rename`). No `research.md` exists (Tier M), so the grep target is `plan.md` alone. Both markers are verbatim unchanged and correctly unreworded. **This is not an author defect.** Each topic is the dispatcher's to answer and the orchestrator's to route via `AskUserQuestion` before Implementation Kickoff Approval; the orchestrator's queries are outstanding. Recorded as failed-by-design; the author's non-action is the correct action. |

### MP-3 measurement — parsed, plus a mutant control

Parsed with a real YAML decoder (`python3` + `yaml.safe_load`) over the
frontmatter block of `spec.md` at HEAD `37949cd5f`:

```
'id' = 'SPEC-QUEUE-UPGRADE-PROOF-001' str
'title' = 'Prove the v3.1.2-to-next queue upgrade path end to end' str
'version' = '0.2.0' str
'status' = 'draft' str
'created' = datetime.date(2026, 9, 3) date
'updated' = datetime.date(2026, 9, 3) date
'author' = 'manager-spec' str
'priority' = 'High' str
'phase' = 'v3.2.0 target' str
'module' = 'internal/kanban' str
'lifecycle' = 'spec-anchored' str
'tags' = 'queue, upgrade, migration, testing, kanban' str
'tier' = 'M' str
```

Twelve canonical fields plus `tier`. No decode error. `status: draft` is in the
8-value enum; `lifecycle: spec-anchored` and `priority: High` are in their
enums per `.claude/rules/moai/development/spec-frontmatter-schema.md:63` and
§ Field Reference. **MP-3 PASS.**

The tool run, and its control — because a clean lint proves nothing on its own:

```
$ moai spec lint .moai/specs/SPEC-QUEUE-UPGRADE-PROOF-001/spec.md
✓ No findings — all SPEC documents are valid
```

Mutant control 1 — the iter-1 D1 defect shape reintroduced on a scratch copy
(`tags` restored to a YAML sequence):

```
ERROR  ParseFailure  …/mutant.md  1  SPEC parsing failed: frontmatter parsing
error: YAML parsing error: yaml: unmarshal errors:
  line 14: cannot unmarshal !!seq into string
1 error(s), 0 warning(s)
```

The linter therefore **does** read this file and **does** reject the exact
defect iter-1 found. The clean pass is not a zero-swept-set pass.

Mutant control 2 — `lifecycle: spec-first` alone, everything else clean:

```
✓ No findings — all SPEC documents are valid
```

Recorded rather than glossed: **the linter does not enum-check `lifecycle`.**
D2's fix is therefore verified against the SSOT schema by reading, not by the
tool. MP-3 binds the SSOT schema rather than the lint subset, so this does not
change the verdict — but "lint is clean" is not evidence for D2 and is not
offered as such.

---

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.97 | 1.0 | Every requirement has one reading; no pronoun ambiguity; normative text free of "should"/"reasonable". D9 resolved: `plan.md:106-108` now cites `todo_queue_root_test.go:89-91` and names the test, and I read those lines — L88 `primary := t.TempDir()`, **L89 `initGitRepo(t, primary)`**, L90 blank, **L91 `t.Setenv("CLAUDE_PROJECT_DIR", primary)`** — an exact match, with the `:100` subdirectory variant now explicitly distinguished rather than silently miscited. Residual deduction: the mild `Where`-vs-`While` pattern quibble on `REQ-QUP-007` (iter-1), and `AC-QUP-002`'s first limb "holds the queue" remains the one limb without a stated observation command. Neither changes what an implementer would build. |
| Completeness | 1.00 | 1.0 | All required sections present (`spec.md` HISTORY L17, §A L38, §B L68, §C L91, §D L156, §E L168; `acceptance.md` §A matrix, §B scenarios, §C quality gate, §D DoD). Five `### Out of Scope — <topic>` H3s at L174/181/193/203/211, each with specific `-` bullets AND a stated reason. All 12 frontmatter fields present AND valid — the iter-1 deduction was solely D1+D2, both resolved. HISTORY carries a `v0.2.0` entry recording the remediation. |
| Testability | 0.88 | 0.75-1.0 | Large gain: D3 and D4 both addressed, and D3 — the card's load-bearing criterion — is now correct (traced below, not accepted). Deducted for **D11**: `AC-QUP-008`'s replacement observation names a repository-relative path that, in the worktree the proof will run from, resolves to a file that does not exist while the queue actually at risk sits elsewhere. That reintroduces the "check that cannot fail" shape D4 existed to remove, on a narrower limb. |
| Traceability | 1.00 | 1.0 | Matrix at `acceptance.md:13-23` — 11 AC rows, every one naming a valid `REQ-QUP-0NN`. `AC-QUP-010 \| REQ-QUP-010` (L23) closes the sole orphan. Coverage the other way: REQ 001→AC-001a+001b; 002-010 one each; 10/10 requirements carry ≥1 AC, 0 uncovered, 0 orphan ACs. |

Aggregate (arithmetic mean): `(0.97 + 1.00 + 0.88 + 1.00) / 4` = **0.9625**.
Iteration 1: **0.875**. Delta **+0.0875**.

---

## Delta verification — each claimed change, independently

### D1 `tags` — **RESOLVED**

`spec.md:13` → `tags: "queue, upgrade, migration, testing, kanban"`. Decodes as
`str`. Mutant control 1 above proves the linter would have caught the old form.

### D2 `lifecycle` — **RESOLVED**

`spec.md:12` → `lifecycle: spec-anchored`, a member of the SSOT enum
(`spec-frontmatter-schema.md:63`). Verified by schema reading; NOT by lint
(mutant control 2 shows lint is blind to this field).

### D3 the RED-establishing mutation — **RESOLVED. The claimed call chain holds.**

This is the finding that mattered, so I traced it against the tree rather than
against the author's description. Every line below was read at HEAD `37949cd5f`.

**Step 1 — the command path reaches the adopting resolver.**

```
internal/cli/todo.go:57  return kanban.BacklogPathForRootAdopting(root)
internal/cli/todo.go:72  return kanban.ResolveTodoQueueRootAdopting(resolveProjectDir())
internal/cli/todo.go:78  return kanban.NewBacklogStore(todoBacklogPath(resolveTodoQueueRoot()))
internal/kanban/state_dir.go:138  dir, _ := resolveStateDir(root, true)
```

So `moai todo` enters `resolveStateDir(root, adopt=true)`. Confirmed.

**Step 2 — an existing `todo/` short-circuits before the relocation branch.**

```go
79  func resolveStateDir(root string, adopt bool) (dir string, legacy bool) {
80      current := StateDirForRoot(root)
81      if dirExists(current) {
82          // Stale-copy policy: once the new name exists it wins uncondition-
83          // ally, and the legacy directory is left exactly where it is. …
87          return current, false
88      }
…
100     if err := relocateStateDir(legacyDir, current); err != nil {
101         return legacyDir, true
102     }
103     return current, false
```

The author's citation `state_dir.go:81-88` is exact, and `:100-103` is exact.
Pre-creating an empty `todo/` makes `dirExists(current)` true, the function
returns at **L87**, and the relocation at **L100** is unreachable. The legacy
directory is left in place — which the code's own comment states as policy, not
as an accident. **Relocation is severed.**

**Step 3 — the migration is severed too, at the resolved path.**

```go
592  func (s *BacklogStore) openEngine(lockHeld bool) (*backlogEngine, error) {
593      switch layout := inspectBacklogLayout(s.path); {
594      case layout.dbExists && layout.jsonExists:            // state D
…
603      case !layout.dbExists && layout.jsonExists:           // state B — migrate
604          if err := s.migrateUnderLock(lockHeld); err != nil {
```

`s.path` is `BacklogPathForRootAdopting`'s answer, i.e.
`<root>/.moai/state/todo/backlog.json` under the mutation. Neither that file
nor `backlog.db` exists, so **neither** switch case matches: `migrateLegacyBacklog`
(`backlog_migrate.go:434`) and `quarantineLegacyBacklog` (`:558`) never see the
seeded document, and `openBacklogEngine` opens a fresh empty database.
The author's citation `backlog_store.go:603-607` is exact. **Migration is severed.**

**Verdict on D3: the mutation genuinely severs BOTH layers.** It is not the
second vacuous mutation.

**Is `AC-QUP-002` now non-vacuous?** Yes. `acceptance.md:51-53` states
preconditions asserted BEFORE any command runs — `kanban/backlog.json` and the
`companions.json` sentinel exist, `todo/` does not. The negative-existence limb
at L57 therefore has something to be false about, and the sentinel limb (L55-56,
"readable at `todo/companions.json` **with its seeded bytes**") is a POSITIVE
assertion the mutation makes unsatisfiable. The byte-equality wording also
immunizes the limb against a `companions.json` that some other surface might
create at the new name.

**Is the RED symptom observable by a stated command?** Yes — `acceptance.md:145-147`
names two file-system observations (`kanban/` still exists; no `companions.json`
under `todo/`), both `os.Stat`-checkable, and `AC-QUP-005` already fixes the
`go test … -run <name> -v` invocation. `AC-QUP-001a` and `AC-QUP-003` are
correctly labelled corroboration rather than the named criterion.

The sentinel's supporting citation `state_dir.go:13-16` was checked and is
exact: "the relocation moves the DIRECTORY, not a file list".

Also verified: the rejected mutation is recorded with its reason
(`acceptance.md:175-182`), keyed on `backlog_migrate.go:434` — which is indeed
`func migrateLegacyBacklog(queuePath string) error`, keyed on the resolved path.
Correct, and recording it is what stops it being re-proposed.

### D4 the `AC-QUP-008` third limb — **PARTIALLY RESOLVED. See D11.**

The `git status` form is gone and the rejection is recorded with its evidence
(`acceptance.md:121-126`). I re-measured that evidence:

```
$ git check-ignore -v .moai/state/todo/backlog.db
.gitignore:311:.moai/state/	.moai/state/todo/backlog.db
```

Exact — `.gitignore:311`, as cited.

On the question asked: **the absent-before → absent-after case is stated as a
PASS condition, not left as a hole.** `acceptance.md:118-119`: "When the file
does not exist before the run, the assertion is that it still does not exist
afterwards." That branch is a real assertion — a test that wrote the live queue
would create the file — so it is not vacuous *in principle*.

It is vacuous *in this environment*, for a different reason. See D11.

### D6 `REQ-QUP-010` — **RESOLVED, and it is a real requirement**

`spec.md:149-154`. Judged against the three tests asked:

- **Real requirement, not an AC restatement?** Yes. It states an acceptance
  *policy* on the deliverable ("shall not be accepted on a GREEN it has never
  been shown capable of failing"), naming no fixture, no directory, and no
  symptom. `AC-QUP-010` supplies all three. The REQ survives a change of
  mutation; the AC does not. That is the right division.
- **GEARS?** Unwanted form (`The … shall not …`), one of the five patterns.
- **Sequence intact?** Yes — MP-1 re-run above: 001-010, no gap, no duplicate.

### D7-D10 — **RESOLVED**

- D7 `priority: High` (`spec.md:9`) — in the enum.
- D8 `title:` and `version: "0.2.0"` quoted (`spec.md:3-4`).
- D9 `plan.md:106-108` re-cited to `todo_queue_root_test.go:89-91`, read and
  confirmed line-for-line (above), with the `:100` variant distinguished.
- D10 `<base>` replaced by `4e4607abe` at `acceptance.md:131`. Cross-checked:
  the same SHA appears at `spec.md:50`, `plan.md:3`, `plan.md:24`,
  `acceptance.md:153`. No `<base>` placeholder remains anywhere
  (`grep -rn '<base>'` → 0 matches). A pinned SHA, not a moving ref — VCI §2.1
  ANCHOR-class, remedy R1, correctly applied.

### Cross-layer sweep — **VERIFIED**

`plan.md:176`'s risk-row mitigation now describes the NEW mutation
(pre-create an empty `todo/`, `state_dir.go:81-88`, stale-copy branch) and
additionally records the rejected one. The plan and the acceptance criterion no
longer prescribe different mutations. `progress.md:10-27` records the iteration-1
verdict, the D-by-D remediation, and D5's deliberate non-remediation.

---

## Regression Check

| Prior-iteration item | State | Evidence |
|---|---|---|
| D1 `tags` sequence | **RESOLVED** | `spec.md:13`, parsed as `str` |
| D2 `lifecycle` enum | **RESOLVED** | `spec.md:12` = `spec-anchored` |
| D3 vacuous mutation | **RESOLVED** | call chain traced above; severs both layers |
| D4 `git status` limb | **PARTIALLY RESOLVED** | form replaced; new path-scope defect D11 |
| D5 clarification markers | **UNRESOLVED — correctly** | `plan.md:10`, `plan.md:22` verbatim; dispatcher's to answer |
| D6 orphan AC | **RESOLVED** | `acceptance.md:23` |
| D7 `priority` | **RESOLVED** | `spec.md:9` |
| D8 unquoted scalars | **RESOLVED** | `spec.md:3-4` |
| D9 line citation | **RESOLVED** | `plan.md:106-108`, verified against the test file |
| D10 `<base>` placeholder | **RESOLVED** | `acceptance.md:131` |

No stagnation: nothing carried unchanged through both iterations except D5,
which is outside the author's authority by this document's own MP-7 clause.

### Preserved-strength check (explicitly requested)

- **`AC-QUP-001b` strictly-greater `last_seq`** — **SURVIVED**, verbatim.
  `acceptance.md:39-40`: "seeded with a `last_seq` strictly greater than the
  highest item id … the new card's id is `last_seq + 1`, demonstrating the
  high-water mark crossed the composition rather than being re-derived from the
  items." Unchanged by the v0.2.0 diff.
- **`AC-QUP-005` zero-match-selector-is-a-failure** — **SURVIVED**, verbatim.
  `acceptance.md:83-84`: "a zero-match selector is a failure of this criterion,
  not a pass." Unchanged by the v0.2.0 diff.

### No-new-defect check on iter-1-passed artifacts

`git diff --stat HEAD~1 -- .moai/specs/SPEC-QUEUE-UPGRADE-PROOF-001/` →
4 files, +126 / −23, all inside the SPEC directory. No file outside it, and no
file under `internal/`, was touched — `REQ-QUP-009` holds through the
remediation itself. I read the full diff hunk by hunk; every hunk maps to a
named defect or to the HISTORY/progress record of one. No unrelated edit rode
along.

---

## Defects Found

**D11** — `acceptance.md:L110-119` (`AC-QUP-008`, third limb) — **the live-queue
path is repository-relative, and in the worktree the proof runs from it names
the wrong file.** — Severity: **major** — Class: **blocking**

`AC-QUP-008` says the file to observe is "this repository's live queue at
`.moai/state/todo/backlog.db`". Measured in this worktree at HEAD `37949cd5f`:

```
$ ls -l .moai/state/todo/backlog.db
absent
$ ls -l /Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.db
-rw-r--r--@ 1 goos staff 368640 Sep  3 15:19 …
$ git rev-parse --git-common-dir
/Users/goos/MoAI/moai-adk-go/.git
```

The queue that is actually at risk is the PRIMARY checkout's, not this
worktree's — and that is by design, not by accident:
`internal/kanban/todo_root.go:95-99` resolves `primaryCheckoutRoot` as
`filepath.Dir(dirs.CommonDir)`, so from any linked worktree the resolver returns
the primary checkout. The file header at `todo_root.go:4-6` records the incident
that made this the design ("30 queued cards on the primary checkout, 'queue is
empty' from a linked worktree").

Consequence: taken literally, in the worktree, the criterion's before/after
comparison takes the absent → absent branch on a file nobody was ever going to
touch, and reports PASS regardless of what happened to the 368,640-byte live
queue one directory level up. That is the same "check that cannot fail" shape
D4 was raised to remove, on a narrower limb — which is why it is major rather
than minor despite being a one-line fix.

**Required fix**: name the path absolutely and derive it from the resolver, not
from the CWD. E.g. state the observation as
`shasum -a 256 "$(dirname "$(git rev-parse --git-common-dir)")/.moai/state/todo/backlog.db"`,
or state plainly that the path is the PRIMARY checkout's
(`/Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.db`) and that the
worktree-relative path is NOT the subject. The absent-before → absent-after
branch is kept — it is correct, and it remains reachable if the primary's queue
is ever absent. The other two limbs of `AC-QUP-008` stay unchanged.

Note this also sharpens `C-1` (`spec.md:158-159`), which carries the same
repository-relative phrasing; fixing the AC without the constraint would leave
the two disagreeing.

---

**D12** — `acceptance.md:L54` (`AC-QUP-002`, first limb) — "`<root>/.moai/state/todo/`
exists **and holds the queue**" is the only limb of the criterion with no stated
observation. The other two limbs (sentinel bytes; `kanban/` absent) are
`os.Stat`/byte-comparable. — Severity: **minor** — Class: **optional**

Not blocking, and explicitly not routed for a fix: `AC-QUP-001a` and
`AC-QUP-004` already carry the "queue contents survived" and "`backlog.db`
exists and is non-empty" assertions, so the composed proof loses nothing if this
limb stays prose. Recorded only so a run-phase implementer does not invent a
third, weaker check for it. Per M6, the orchestrator may leave this alone.

---

## Recommendation

**FAIL, on MP-7 alone.** The route to PASS is two items, and only one of them
belongs to the author:

1. **The orchestrator** resolves both `[NEEDS CLARIFICATION]` topics with the
   dispatcher via `AskUserQuestion` — `plan.md:10` (what G2 is) and
   `plan.md:22` (downgrade intent vs quarantine rename) — before Implementation
   Kickoff Approval. Preload with `ToolSearch(query: "select:AskUserQuestion")`.
   The markers are well-formed and a third party can act on either without
   re-deriving it; it is the open state, not the writing, that fails MP-7. Do
   NOT close this by editing the markers out.
2. **manager-spec** fixes **D11** — one sentence in `acceptance.md:110-119`, plus
   the matching phrasing in `spec.md:158-159` (`C-1`), naming the PRIMARY
   checkout's queue path rather than a CWD-relative one.

**D12 is optional** and the orchestrator may decline it without affecting the
verdict.

Nothing else is outstanding. The Tier M iteration ceiling (2) is now reached; a
third iteration is not available under the Retry Loop Contract, so once D11 and
the clarification gate are closed the card proceeds on this verdict's terms
rather than through another audit pass.

---

## Evidence, Baseline-attribution, Gaps, Residual-risk

**Claim** — the nine remediated defects are genuinely remediated; the SPEC's
frontmatter is now schema-valid and machine-lintable; the replacement RED
mutation severs the composed path in both layers; one new blocking defect (D11)
and one optional defect (D12) exist; MP-7 remains failed by design.

**Evidence** — commands run in this session, against this tree, with their
output quoted inline above: `git rev-parse` (tree + HEAD + git-common-dir);
`grep -n '^### REQ-'` and `'^### AC-'`; `python3` + `yaml.safe_load` on the
frontmatter; `moai spec lint` on the file and on two scratch mutants;
`grep -rn '\[NEEDS CLARIFICATION' plan.md`; `git check-ignore -v`;
`grep -Eoh 'SPEC-…'`; `grep -c syscall`; `ls -l` on both queue paths;
`git diff --stat`/`git diff` of `HEAD~1..HEAD` scoped to the SPEC directory;
and direct reads of `internal/kanban/state_dir.go`,
`internal/kanban/backlog_store.go`, `internal/kanban/backlog_migrate.go`,
`internal/kanban/todo_root.go`, `internal/cli/todo.go`, and
`internal/cli/todo_queue_root_test.go`.

**Baseline-attribution** — worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t470`, branch
`WT-queue-upgrade-proof`, HEAD `37949cd5fd635dca069b6f711aae994bb0c9e311`, in
this run. The iteration-1 baseline compared against is
`.moai/reports/t470/plan-audit.md` at HEAD `38d1526b1`, score `0.875`. The SPEC
base `4e4607abe` is a pinned SHA, not a moving ref.

**Gaps — what I did NOT observe.**

1. **No Go test was run.** `go test ./internal/kanban/... ./internal/cli/` was
   not executed (C-2/C-3 and the pass's read-only scope). D3's severance is
   established by STATIC trace of the call chain, not by an executed RED. The
   trace is exact at the line level, but a static trace is a strong argument,
   not a run. The actual RED/GREEN pair is run-phase work under `AC-QUP-010`.
2. **The `moai` binary's provenance is unattributed.** `moai version` prints
   `v3.2.0-rc.0` and no commit stamp, so I cannot state the judging build's
   commit alongside the tree's HEAD (VCI §2.2). The mutant control bounds this —
   the invoked build demonstrably reads the file and rejects the iter-1 defect —
   but it does not establish that the build is not behind the tree on some
   OTHER rule that would have fired.
3. **`lifecycle` is not lint-covered.** Mutant control 2 passed clean with
   `lifecycle: spec-first`. D2 is verified against the schema SSOT by reading
   only. Any other frontmatter field the linter silently skips was not
   enumerated.
4. **Iteration-1's fact verification was not re-run.** Per the narrow scope,
   `spec.md §A`'s measurement table and `plan.md §B/§C` were carried forward
   from iteration 1 unmeasured, except the specific citations the v0.2.0 diff
   touches (`.gitignore:311`; `todo_queue_root_test.go:89-91`;
   `state_dir.go:13-16`, `:81-88`, `:100-103`; `backlog_store.go:603`;
   `backlog_migrate.go:434`, `:558`), all of which I did re-read.
5. **I did not observe whether `moai todo` itself creates a `companions.json`**
   under the resolved directory on a first run. `AC-QUP-002`'s byte-equality
   wording makes the sentinel limb robust either way, which is why I did not
   pursue it — but the underlying behavior is unobserved.
6. **No cross-model audit was run.** No `audit_multi` / `codex_audit` /
   `glm_audit` call was made; this verdict is Claude-only.

**Residual-risk — what could still be wrong despite the above.**

- D3's trace assumes `resolveProjectDir()` (`internal/cli/todo.go:72`) hands the
  test's temp root through unmodified. I read the three call sites but did not
  read `resolveProjectDir` itself, so a transformation there could in principle
  change which root reaches `resolveStateDir`. `plan.md` already prescribes the
  `t.Setenv("CLAUDE_PROJECT_DIR", …)` control for exactly this, which is why I
  judge the risk low rather than open.
- D11's fix touches a constraint (`C-1`) as well as an AC. A fix applied to only
  one of the two leaves the SPEC internally inconsistent, which is a shape this
  audit would catch on a later pass but no later pass is available at the Tier M
  ceiling.
- The score is an arithmetic mean of four judged dimensions. The judgments are
  anchored to the M3 rubric bands but remain judgments; a reader who weights
  Testability higher than the other three would compute a lower aggregate. The
  verdict does not depend on the arithmetic — MP-7 carries it.

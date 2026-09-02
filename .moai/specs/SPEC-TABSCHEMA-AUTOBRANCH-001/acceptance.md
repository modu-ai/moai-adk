# SPEC-TABSCHEMA-AUTOBRANCH-001 — Acceptance Criteria

All baselines below were measured on this tree at HEAD `7ed6edb3e`, worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t316`, branch `WT-tabschema-autobranch`.

Path shorthand used throughout:

- `LOCAL` = `.claude/skills/moai-workflow-project/schemas/tab_schema.json`
- `TMPL` = `internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json`

Numbering convention: this document carries **eight** logical criteria, AC-TSA-001 through
AC-TSA-008. Two of them are split into paired sub-criteria — `AC-TSA-005` / `AC-TSA-005b` and
`AC-TSA-007` / `AC-TSA-007b` — because in each case one requirement has two halves that fail
independently and must be measured separately. The pairing convention is used deliberately in place
of adding a ninth and tenth criterion: this is a Tier S SPEC, whose acceptance-criterion ceiling is
eight, and the two additions verify halves of requirements the existing criteria already cover
rather than introducing new obligations.

---

## AC-TSA-001 — the counting criterion (primary) — covers REQ-TSA-001, REQ-TSA-003, REQ-TSA-004

**Given** the `mode_admits` predicate: a batch admits mode `M` when it has no `condition`, or its
`condition.field` is not `git_strategy.mode`, or its operator is `equals` and `value == M`, or its
operator is `not_equals` and `value != M`;

Scope clause (deliberate, not an oversight): the predicate inspects **batch-level `condition` only**
and does not read the tab-level `mode_condition` key. That key is a different vocabulary — measured
on this tree, it occurs exactly twice and carries the value `"INITIALIZATION"` both times, never a
`git_strategy.mode` value — and no `auto_branch`-bearing batch inherits it (measured: batches 3.3,
3.6, and 3.10 each report `inherited_mode_condition = None`). `mode_condition` is therefore outside
this predicate's scope by decision, and its omission changes no counted value here.

**When** — for each of `LOCAL` and `TMPL` — the number of question objects whose `field` contains the
substring `auto_branch` and whose enclosing batch admits mode `M` is counted, for
`M ∈ {personal, team, manual}`;

**Then** the count is **exactly** `personal=1`, `team=1`, `manual=0`, for **both** copies.

This is the criterion that decides correctness. It fails on under-deletion (`personal=2`) and equally
on over-deletion (`personal=0`), which is why it is stated as an exact integer rather than as
"duplicates removed".

Recipe (run once per copy, substituting the path):

```
python3 -c "
import json,sys
def adm(c,m):
    if not c or c.get('field')!='git_strategy.mode': return True
    o,v=c.get('operator'),c.get('value')
    return m==v if o=='equals' else (m!=v if o=='not_equals' else True)
def cnt(p,m):
    n=0
    def w(o):
        nonlocal n
        if isinstance(o,dict):
            if 'batch_id' in o and adm(o.get('condition'),m):
                n+=sum(1 for q in o.get('questions',[]) if 'auto_branch' in str(q.get('field','')))
            for v in o.values(): w(v)
        elif isinstance(o,list):
            for v in o: w(v)
    w(json.load(open(p)))
    return n
p=sys.argv[1]
for m in ('personal','team','manual'): print('%s=%d'%(m,cnt(p,m)))
" <PATH>
```

**RED now** — measured, verbatim:

```
.claude/skills/moai-workflow-project/schemas/tab_schema.json
  personal=2
  team=2
  manual=0
internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json
  personal=2
  team=2
  manual=0
```

Note the `manual=0` cell is GREEN now and must stay `0` — see `spec.md` §4, first Out of Scope entry.

---

## AC-TSA-002 — dead paths fully absent — covers REQ-TSA-002

**Given** the two dead configuration paths;
**When** `grep -c` counts them across both copies;
**Then** each count is `0`.

```
grep -c 'git_strategy.personal.auto_branch' <PATH>
grep -c 'git_strategy.team.auto_branch' <PATH>
```

**RED now** — `grep -n 'auto_branch' LOCAL` returns six dead-path lines:

```
517:              "question": "Auto-create branches for Personal mode? (current: {{git_strategy.personal.auto_branch}})",
519:              "field": "git_strategy.personal.auto_branch",
534:              "current_value_path": "git_strategy.personal.auto_branch",
727:              "question": "Auto-create branches for Team mode? (current: {{git_strategy.team.auto_branch}})",
729:              "field": "git_strategy.team.auto_branch",
744:              "current_value_path": "git_strategy.team.auto_branch",
```

---

## AC-TSA-003 — batch 3.10 untouched — covers REQ-TSA-001, REQ-TSA-005

**Given** the canonical batch 3.10 auto-branch question;
**When** `grep -c 'git_strategy.{mode}.automation.auto_branch'` runs on each copy;
**Then** the count is exactly `3` (its `question`, `field`, and `current_value_path` sites), unchanged
from baseline.

**GREEN now, must stay GREEN** — measured, verbatim:

```
1006:              "question": "Auto-create a branch for each SPEC implementation? (current: {{git_strategy.{mode}.automation.auto_branch}})",
1008:              "field": "git_strategy.{mode}.automation.auto_branch",
1024:              "current_value_path": "git_strategy.{mode}.automation.auto_branch",
```

---

## AC-TSA-004 — total occurrence count — covers REQ-TSA-002 (corroborator)

**Given** the baseline of nine `auto_branch` occurrences per copy;
**When** `grep -c 'auto_branch' <PATH>` runs on each copy;
**Then** it returns **exactly `3`** on each.

Derivation: each deleted question object carries three occurrences (`question` text, `field`,
`current_value_path`), so two deletions remove six, leaving `9 - 6 = 3` — precisely batch 3.10's three
canonical sites.

**RED now** — measured, verbatim:

```
9
9
```

(first line `LOCAL`, second `TMPL`)

---

## AC-TSA-005 — template/local byte-identity — covers REQ-TSA-007 (source pair)

**Given** the Template-First rule;
**When** `diff -q LOCAL TMPL` runs after `make build` and the mirror sync;
**Then** it produces empty output and exit code `0`.

**GREEN now, must stay GREEN** — measured, verbatim:

```
diff_rc=0
```

---

## AC-TSA-005b — the *embedded* asset carries the change — covers REQ-TSA-007 (regeneration half)

Paired sub-criterion of AC-TSA-005. AC-TSA-005 compares the two **source** files; this one asserts
that what the binary actually ships was regenerated. The pair exists because the two halves fail
independently: an agent that edits the template and syncs the mirror but skips `make build` passes
AC-TSA-005 while the binary still embeds the pre-change schema — and the embedded copy is what
`moai update` deploys to users.

**Given** that `internal/template/embed.go` compiles the template tree into the binary with
`//go:embed all:templates`, so the shipped schema is a compiled-in copy and not the file on disk;

**When** M2 runs `make build` — which is `go build $(LDFLAGS) -o bin/moai ./cmd/moai` (`Makefile:23-25`,
measured) — and the produced binary `bin/moai` is scanned;

**Then** all three of the following hold:

1. `make build` exits `0`;
2. each dead-path string occurs **exactly `0`** times in `bin/moai`;
3. **control** — the canonical string occurs **at least `1`** time in `bin/moai`.

Item 3 is not decoration. Without it, a `0` in item 2 is uninterpretable: it reads the same whether
the deletion landed or the schema is absent from the binary altogether. Item 3 proves the scan is
actually reaching the embedded schema.

Recipe (fixed-string matching — `{mode}` contains regex metacharacters):

```
make build; echo "make_build_rc=$?"
grep -aoF 'git_strategy.personal.auto_branch'        bin/moai | wc -l   # expect 0
grep -aoF 'git_strategy.team.auto_branch'            bin/moai | wc -l   # expect 0
grep -aoF 'git_strategy.{mode}.automation.auto_branch' bin/moai | wc -l # expect >= 1 (control)
```

**Attribution — why a `0` here is attributable to the schema.** Measured on this tree:

```
grep -rn 'git_strategy\.personal\.auto_branch\|git_strategy\.team\.auto_branch' internal/ pkg/ cmd/ | grep -vc 'tab_schema.json'
0
```

The two dead-path strings occur nowhere in the compiled sources except the `tab_schema.json` copies,
and of those only the template copy lives under `internal/template/templates/`. The embedded schema
is therefore the binary's sole source of those strings, so item 2's `0` cannot be produced by
anything else.

**RED now — not observed, and stated as such.** `ls bin/` returns
`ls: bin/: No such file or directory` on this tree, so the artifact under test does not yet exist
and this criterion cannot be evaluated before M2. What *was* measured instead is the source-side
fact that makes it red-in-principle: the template copy currently carries `3` occurrences of each
dead path (`grep -c 'auto_branch' TMPL` → `9`, of which six are the two dead objects), so a binary
built from this tree today would embed them. Green path: M2 (`make build`) is the milestone that
flips it, and item 2 becomes `0` while item 3 stays `>= 1`.

**Explicitly excluded observation point — reading the embedded FS from inside `go test`.** A test
that calls `template.EmbeddedTemplates()` and compares the embedded copy against the source file is
**not** a valid check here, and the trap is worth naming because the code makes it look valid.
`go test` recompiles the test binary on every run, so its `//go:embed` snapshot is taken from the
same working tree the comparison reads: both sides move together and the assertion is a tautology.
An agent that edits the template and skips `make build` passes such a test. Measured elsewhere in
this repository by mutation: with a mutant injected, the embedded-comparison test survived (PASS)
while only the golden body failed. The observation point must therefore be an artifact built
**outside** the test run — the `bin/moai` produced by `make build`, as above.

**Rejected alternative, recorded so it is not re-proposed.** Asserting that `make build` changes the
`moai-workflow-project` hash in `internal/template/catalog.yaml` was considered and **refuted by
measurement**: `resolveHashSourcePath` in `internal/template/scripts/gen-catalog-hashes.go` resolves
a skill *directory* entry to its root `SKILL.md` and hashes **that file alone** — it is not a
directory-tree hash. Reproduced independently: the normalized sha256 of
`internal/template/templates/.claude/skills/moai-workflow-project/SKILL.md` is
`9e4f7a52b977e2a4014c321931aacdd8ebf04559b9e2b8ba26aa4ee9abb2dd16`, byte-identical to the `hash:`
field at `catalog.yaml:74`. Editing `schemas/tab_schema.json` cannot move that value, so a
catalog-hash criterion would be unsatisfiable by any correct implementation of this card.

---

## AC-TSA-006 — both copies remain valid JSON — covers REQ-TSA-008

**Given** that the deleted objects are the final element of their `questions` array, so the edit must
also remove a trailing comma;
**When** `python3 -m json.tool <PATH> > /dev/null` runs on each copy;
**Then** the exit code is `0` for both.

**GREEN now, must stay GREEN** — measured, verbatim:

```
local_rc=0
template_rc=0
```

---

## AC-TSA-007 — no question object other than the two is removed — covers REQ-TSA-006 (removal half; the alteration half is AC-TSA-007b)

**Given** the schema's total question count across all batches;
**When** every `batch_id` object's `questions` array length is summed;
**Then** the total is **exactly `46`** — a delta of exactly `-2` from the baseline `48`.

Recipe:

```
python3 -c "
import json,sys
t=0
def w(o):
    global t
    if isinstance(o,dict):
        if 'batch_id' in o: t+=len(o.get('questions',[]))
        for v in o.values(): w(v)
    elif isinstance(o,list):
        for v in o: w(v)
w(json.load(open(sys.argv[1])))
print('TOTAL_QUESTIONS =',t)
" <PATH>
```

**RED now** — measured, verbatim:

```
TOTAL_QUESTIONS = 48
```

Corroborating per-batch baseline (measured): batch 3.3 has 4 questions, 3.6 has 4, 3.10 has 3. After
the change, 3.3 and 3.6 must each have `3`, and 3.10 must still have `3`.

---

## AC-TSA-007b — the diff contains nothing but the two removals — covers REQ-TSA-006 (alteration half)

Paired sub-criterion of AC-TSA-007. AC-TSA-007 counts questions, so it catches a question that
disappears; it cannot see a question that *changes*. REQ-TSA-006's second clause — "alter no other
question object" — is what this criterion verifies, and it is the only criterion in the set that a
silent edit to an untouched question fails.

**Given** the plan-phase baseline tree `7ed6edb3e`, at which both copies are byte-identical and
tracked by git, so `git show 7ed6edb3e:<path>` yields the pre-change bytes independently of whatever
the run phase has since committed (the baseline is pinned to that SHA and MUST NOT be written as
`HEAD` — once M1 lands, `HEAD` is the post-change commit and the comparison becomes vacuous);

**When** — for each of `LOCAL` and `TMPL` — the baseline copy is parsed, the question objects whose
`field` is `git_strategy.personal.auto_branch` or `git_strategy.team.auto_branch` are removed from
that parsed structure, and the result is compared for deep equality against the parsed post-change
file;

**Then** both of the following hold, on both copies:

- `REMOVED_OBJECTS = 2 ['git_strategy.personal.auto_branch', 'git_strategy.team.auto_branch']`
- `IDENTICAL_AFTER_REMOVAL = True`

That is the whole assertion: **the only difference between the pre-change and post-change file is the
removal of the two named question objects.** Any other difference anywhere in the schema — an edited
`question` string, a changed `header`, a reordered or reworded `options` array, an added key, a
changed batch `condition` — makes the deep comparison `False`. A count-based proxy cannot do this:
the mutant that silently alters an untouched question leaves every count in this document unchanged.

Recipe (run once per copy, substituting the path):

```
python3 -c "
import json, subprocess, sys
BASE = '7ed6edb3e'
DEAD = ('git_strategy.personal.auto_branch', 'git_strategy.team.auto_branch')
path = sys.argv[1]
before = json.loads(subprocess.run(['git','show',BASE+':'+path],
                                   capture_output=True, text=True, check=True).stdout)
after = json.load(open(path))
removed = []
def prune(o):
    if isinstance(o, dict):
        if isinstance(o.get('questions'), list):
            keep = []
            for q in o['questions']:
                (removed.append(q.get('field')) if isinstance(q, dict) and q.get('field') in DEAD
                 else keep.append(q))
            o['questions'] = keep
        for v in o.values(): prune(v)
    elif isinstance(o, list):
        for v in o: prune(v)
prune(before)
print('REMOVED_OBJECTS =', len(removed), sorted(removed))
print('IDENTICAL_AFTER_REMOVAL =', before == after)
" <PATH>
```

**RED now** — measured, verbatim (`LOCAL` first, then `TMPL`):

```
REMOVED_OBJECTS = 2 ['git_strategy.personal.auto_branch', 'git_strategy.team.auto_branch']
IDENTICAL_AFTER_REMOVAL = False
REMOVED_OBJECTS = 2 ['git_strategy.personal.auto_branch', 'git_strategy.team.auto_branch']
IDENTICAL_AFTER_REMOVAL = False
```

**Why it is red, stated explicitly** — so this cannot be confused with a criterion that is red at
arrival and red forever. `REMOVED_OBJECTS = 2` proves both named objects were located in the
baseline, so the criterion is not failing on a bad path, a parse error, or a wrong `field` spelling.
The single reason `IDENTICAL_AFTER_REMOVAL` is `False` is that the working-tree copy still contains
the two objects the baseline-side prune removed. **Green path:** M1 deletes exactly those two objects
from `TMPL` and M3 mirrors it to `LOCAL`; at that point both sides carry the same structure and the
comparison flips to `True`. No unrelated file and no other actor is involved in the flip.

**Textual corroborator (optional, not the deciding check).** The structural comparison above is the
verdict; `git diff --numstat` is a cheap cross-check whose expected value is derivable from measured
line spans. Each deleted object spans 21 lines (`516-536` and `726-746`, both measured on this tree),
and each deletion also strips the trailing comma from the preceding element's closing line (`515`,
`725`), which git records as one deletion plus one addition. So `git diff --numstat 7ed6edb3e -- <PATH>` is expected to read `2` added and `44` deleted per copy. A mismatch is a signal to look,
not a verdict by itself — whitespace-only churn would move these numbers while leaving the
structural check `True`.

**Known limitation, recorded rather than papered over.** The comparison is on *parsed* JSON, so a
pure reformat of untouched regions (re-indentation, key reordering) would still read `True`. That
case is excluded by `plan.md` §B ("Do not round-trip it through a JSON serializer") and would be
visible in the corroborator's line counts; it is not covered by this criterion.

---

## AC-TSA-008 — template neutrality preserved (delta assertion) — covers REQ-TSA-009

**Given** that the template copy already contains pre-existing matches for a neutrality-style scan —
`"schema_updated": "2025-12-22"` at line 3, and the `feature/SPEC-` branch-prefix examples at lines
625 and 915 — none of which this change touches;

**When** the neutrality scan is re-run on `TMPL` after the change;

**Then** the scan's output is **byte-identical to the baseline block recorded below** — not merely
the same three line numbers, but the same three lines verbatim, and no fourth line. Line-set identity
alone would still be satisfied by a new violation appended to line 3, 625, or 915; verbatim identity
is not. Equivalently: the change introduces no *new* SPEC ID, internal date, commit SHA, or
local-rule reference into `internal/template/templates/**`, and alters none of the three
pre-existing matches.

```
grep -n 'SPEC-\|20[0-9][0-9]-' <TMPL>
```

**Baseline now** — measured, verbatim:

```
internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json:3:  "schema_updated": "2025-12-22",
internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json:625:              "question": "Branch prefix for Personal mode? To change, type the new prefix (e.g., feature/SPEC-) in the text field below. (current: {{git_strategy.personal.branch_prefix}})",
internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json:915:              "question": "Branch prefix for Team mode? To change, type the new prefix (e.g., feature/SPEC-) in the text field below. (current: {{git_strategy.team.branch_prefix}})",
```

This is stated as a delta rather than an absolute `0` because the absolute count is already non-zero
for reasons that predate this card; asserting `0` would fail on content this SPEC does not own.

---

## Definition of Done

- [x] AC-TSA-001 through AC-TSA-008, **including the paired sub-criteria AC-TSA-005b and
      AC-TSA-007b**, all measured GREEN, with verbatim output recorded in `progress.md` §E.2
- [x] `make build` run after the template edit (Template-First). This is no longer a bare checklist
      item: its exit code and the resulting binary scan are measured by AC-TSA-005b, and the DoD
      bullet is satisfied by that criterion's recorded output rather than by an assertion here
- [x] `moai spec lint` run unpiped with output redirected to a file, exit code reported, and the file
      searched for `SPEC-TABSCHEMA-AUTOBRANCH-001` — never through `tail`
- [x] No file outside `LOCAL`, `TMPL`, the embedded-asset regeneration, and this SPEC directory is
      modified

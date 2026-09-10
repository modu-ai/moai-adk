# t316 — SPEC-TABSCHEMA-AUTOBRANCH-001 run-phase verdict

Card: t316
Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t316`
Branch: `WT-tabschema-autobranch`
Plan-phase baseline (pinned): `7ed6edb3e`
Pre-change working-tree HEAD: `5a841ee22`
Cycle: DDD (ANALYZE / PRESERVE / IMPROVE) — behaviour-preserving key alignment, no Go production code touched.

Path shorthand:

- `LOCAL` = `.claude/skills/moai-workflow-project/schemas/tab_schema.json`
- `TMPL` = `internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json`

---

## 1. Claim

The two question objects bound to the dead configuration paths
`git_strategy.personal.auto_branch` (batch 3.3) and `git_strategy.team.auto_branch` (batch 3.6)
were deleted from `TMPL`, the embedded asset was regenerated with `make build`, and `LOCAL` was
brought to byte-identity. Nothing else in either copy changed; batch 3.10's three canonical
`git_strategy.{mode}.automation.auto_branch` sites are untouched.

All ten acceptance criteria — AC-TSA-001 through AC-TSA-008 plus the paired sub-criteria
AC-TSA-005b and AC-TSA-007b — measure **PASS**.

## 2. Evidence — AC matrix

Every row records the command actually run and its verbatim output. `Tree` names the tree the
measurement was taken against.

### ANALYZE / RED baseline (before the edit)

| Command | Verbatim output | Tree |
|---|---|---|
| `python3 ac001.py LOCAL` | `personal=2` / `team=2` / `manual=0` | `5a841ee22` (working tree, pre-edit) |
| `python3 ac001.py TMPL` | `personal=2` / `team=2` / `manual=0` | `5a841ee22` (working tree, pre-edit) |
| `python3 ac007b.py TMPL` | `REMOVED_OBJECTS = 2 ['git_strategy.personal.auto_branch', 'git_strategy.team.auto_branch']` / `IDENTICAL_AFTER_REMOVAL = False` | baseline `7ed6edb3e` vs pre-edit working tree |
| `diff -q LOCAL TMPL` | `diff_rc=0` | `5a841ee22` (working tree, pre-edit) |
| `grep -n 'auto_branch' TMPL` | 9 lines: `517,519,534` (personal) · `727,729,744` (team) · `1006,1008,1024` (batch 3.10) | `5a841ee22` (working tree, pre-edit) |

The measured spans confirm `plan.md` §C: the personal object is lines `516-536` and the team object
`726-746`, each the final element of its `questions` array, each preceded by a `},` at `515` / `725`.

### GREEN — post-change acceptance battery

| AC | Command | Verbatim output | Tree | Verdict |
|---|---|---|---|---|
| AC-TSA-001 | `python3 ac001.py LOCAL` | `personal=1` / `team=1` / `manual=0` | working tree, post-edit (pre-commit, HEAD `5a841ee22`) | PASS |
| AC-TSA-001 | `python3 ac001.py TMPL` | `personal=1` / `team=1` / `manual=0` | same | PASS |
| AC-TSA-002 | `grep -c 'git_strategy.personal.auto_branch' LOCAL` ; `grep -c 'git_strategy.team.auto_branch' LOCAL` | `0` / `0` | same | PASS |
| AC-TSA-002 | same two greps on `TMPL` | `0` / `0` | same | PASS |
| AC-TSA-003 | `grep -cF 'git_strategy.{mode}.automation.auto_branch' LOCAL` ; same on `TMPL` | `3` / `3` | same | PASS |
| AC-TSA-004 | `grep -c 'auto_branch' LOCAL` ; same on `TMPL` | `3` / `3` | same | PASS |
| AC-TSA-005 | `diff -q LOCAL TMPL; echo "diff_rc=$?"` | `diff_rc=0` | same | PASS |
| AC-TSA-005b (1) | `make build; echo "make_build_rc=$?"` | `make_build_rc=0` (full log below) | same | PASS |
| AC-TSA-005b (2) | `grep -aoF 'git_strategy.personal.auto_branch' bin/moai \| wc -l` | `0` | `bin/moai` built from this tree | PASS |
| AC-TSA-005b (2) | `grep -aoF 'git_strategy.team.auto_branch' bin/moai \| wc -l` | `0` | same | PASS |
| AC-TSA-005b (3) control | `grep -aoF 'git_strategy.{mode}.automation.auto_branch' bin/moai \| wc -l` | `4` (≥ 1) | same | PASS |
| AC-TSA-006 | `python3 -m json.tool LOCAL > /dev/null; echo "local_rc=$?"` | `local_rc=0` | working tree, post-edit | PASS |
| AC-TSA-006 | `python3 -m json.tool TMPL > /dev/null; echo "template_rc=$?"` | `template_rc=0` | same | PASS |
| AC-TSA-007 | `python3 ac007.py LOCAL` ; `python3 ac007.py TMPL` | `TOTAL_QUESTIONS = 46` / `TOTAL_QUESTIONS = 46` | same | PASS |
| AC-TSA-007b | `python3 ac007b.py LOCAL` | `REMOVED_OBJECTS = 2 ['git_strategy.personal.auto_branch', 'git_strategy.team.auto_branch']` / `IDENTICAL_AFTER_REMOVAL = True` | baseline `7ed6edb3e` vs post-edit working tree | PASS |
| AC-TSA-007b | `python3 ac007b.py TMPL` | `REMOVED_OBJECTS = 2 ['git_strategy.personal.auto_branch', 'git_strategy.team.auto_branch']` / `IDENTICAL_AFTER_REMOVAL = True` | same | PASS |
| AC-TSA-008 | `grep -n 'SPEC-\|20[0-9][0-9]-' TMPL` | 3 lines: `3:  "schema_updated": "2025-12-22",` · `604:` personal branch-prefix question · `873:` team branch-prefix question | working tree, post-edit | PASS (see §2.3) |

`ac001.py`, `ac007.py`, `ac007b.py` are the verbatim recipes from `acceptance.md`, written to
`.moai/state/verify/t316/` and invoked unmodified. That directory is gitignored, so the recipes are
reproduced in `acceptance.md` itself rather than committed here.

### 2.1 `make build` output (AC-TSA-005b item 1)

Exit code `0`. The build ran `templ generate`, then `gen-catalog-hashes --all` over 45 entries,
then `go build -ldflags … -o bin/moai ./cmd/moai`. The `moai-workflow-project` catalog hash printed
`9e4f7a52b977e2a4014c321931aacdd8ebf04559b9e2b8ba26aa4ee9abb2dd16` — byte-identical to the value
`acceptance.md` records at `catalog.yaml:74`, confirming that criterion's recorded refutation of the
catalog-hash alternative: the hash covers `SKILL.md` alone and a `schemas/` edit cannot move it.
`git status --short` after the build shows `catalog.yaml` unmodified, which is that prediction
holding.

### 2.2 AC-TSA-005b control-count note (`4`, not `3`)

The control string occurs `4` times in `bin/moai` where `TMPL` carries `3`. Measured cause:

```
grep -rlF 'git_strategy.{mode}.automation.auto_branch' internal/template/templates/
internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md
internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json
```

A second template file carries one occurrence, and both files are inside the `//go:embed
all:templates` tree. `3 + 1 = 4`. The criterion asserts `>= 1`, so this is a PASS, and the count is
recorded rather than rounded because it is the control that makes the two `0`s interpretable.

Attribution for those `0`s, re-measured on this tree:

```
grep -rn 'git_strategy\.personal\.auto_branch\|git_strategy\.team\.auto_branch' internal/ pkg/ cmd/ | grep -vc 'tab_schema.json'
0
```

The dead-path strings exist nowhere in the compiled sources but the two `tab_schema.json` copies, of
which only `TMPL` is embedded — so the embedded schema is the binary's sole possible source of them.

### 2.3 AC-TSA-008 — how "byte-identical block" was discharged

The criterion asks for the scan output to be byte-identical to the baseline block, "not merely the
same three line numbers". A deletion above those lines necessarily shifts the `grep -n` prefixes
(`625 → 604`, `915 → 873`; the shifts are exactly `-21` and `-42`, the deleted line counts). The
line *content* is what the criterion is about, so it was discharged by comparing content directly:

```
git show 7ed6edb3e:TMPL > tmpl_before.json
grep -h 'SPEC-\|20[0-9][0-9]-' tmpl_before.json > neutrality_before.txt
grep -h 'SPEC-\|20[0-9][0-9]-' TMPL             > neutrality_after.txt
diff neutrality_before.txt neutrality_after.txt; echo "neutrality_diff_rc=$?"
neutrality_diff_rc=0
```

Three matches before, the same three after, byte-identical, no fourth line. No new SPEC ID, date,
commit SHA, or local-rule reference entered the template.

### 2.4 AC-TSA-007b textual corroborator — expectation missed, structural check unaffected

`acceptance.md` predicts `2` added / `44` deleted per copy. Measured:

```
git diff --numstat 7ed6edb3e -- LOCAL TMPL
0	42	.claude/skills/moai-workflow-project/schemas/tab_schema.json
0	42	internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json
```

The criterion itself states a mismatch here is "a signal to look, not a verdict by itself", so it was
looked at. `git diff --unified=2` shows two pure-deletion hunks of 21 lines each and **no addition
anywhere**. The cause is a diff-representation choice, not a content difference: the prediction
assumed git would model the trailing-comma strip as one deletion plus one addition, but the deleted
object's own closing line `            }` is byte-identical to the line the preceding element now
needs, so git expresses the whole edit as deleting `},` through `"required": true` — 21 contiguous
lines, zero additions. `2 × 21 = 42`.

The deciding check is unaffected and reads `IDENTICAL_AFTER_REMOVAL = True` on both copies, which is
the assertion that actually rules out an altered untouched question.

### 2.5 Definition-of-Done items outside the AC matrix

```
bin/moai spec lint > .moai/state/verify/t316/spec-lint.txt 2>&1; echo "spec_lint_rc=$?"
spec_lint_rc=0
```

Run unpiped with output redirected to a file, per the DoD wording, then searched:

```
grep -n 'SPEC-TABSCHEMA-AUTOBRANCH-001' .moai/state/verify/t316/spec-lint.txt
(no output — zero matches)
```

The file's final line reads `0 error(s), 64 warning(s)`; all 64 warnings name other SPECs
(grandfathered-era frontmatter and `StatusGitConsistency` findings), none this one.

Scope containment:

```
git status --short
 M .claude/skills/moai-workflow-project/schemas/tab_schema.json
 M internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json
```

Exactly the two files. `catalog.yaml` was rewritten by the build and came out byte-identical;
`bin/moai` and `.moai/state/verify/` are gitignored.

## 3. Baseline-attribution

- **Structural comparison (AC-TSA-007b)** is attributed to the pinned plan-phase tree `7ed6edb3e`,
  read with `git show 7ed6edb3e:<path>` — never to `HEAD`, which the criterion warns becomes vacuous
  once the run-phase commit lands. The baseline was reachable from this worktree and parsed cleanly;
  `REMOVED_OBJECTS = 2` proves both named objects were located in it, so the check is not failing or
  passing on a bad path.
- **Every RED cell** was re-measured on this tree pre-edit rather than carried over from
  `acceptance.md`. The re-measured values match the recorded ones (`personal=2 / team=2 / manual=0`,
  `9` occurrences, `TOTAL_QUESTIONS = 48` implied by the `-2` delta, `IDENTICAL_AFTER_REMOVAL =
  False`).
- **Line spans** were re-measured with `grep -n` + `sed -n` before editing, not taken from
  `plan.md` §C, and matched it (`516-536`, `726-746`, preceded by `},` at `515`, `725`).
- **The pre-edit working copy equals the baseline for these two files.** Directly observed: the
  post-change `git diff 7ed6edb3e` contains exactly the two deletion hunks and nothing else.
- **AC-TSA-005b** is attributed to `bin/moai` produced by `make build` in this run — an artifact
  built outside any test process, per the criterion's exclusion of the tautological embedded-FS
  `go test` observation point.

## 4. Gaps — what was NOT observed

- **No Go test was run.** Per the dispatch constraint, `go test ./...` was not executed, and no
  scoped `go test` was run either: the change touches no Go source, and the only Go-visible
  consequence — the embedded asset — is measured directly by AC-TSA-005b's binary scan. The
  full-suite verdict belongs to CI. No claim is made here about the Go test suite's state.
- **No lint or vet beyond `moai spec lint`.** `golangci-lint` and `go vet` were not run; `go build`
  succeeded as part of `make build`, which is the only compile-level evidence collected.
- **No runtime consumer was exercised.** No code reads `tab_schema.json` (`spec.md` §4), so the
  interview semantics AC-TSA-001's `mode_admits` predicate encodes were never executed — the
  predicate remains a reconstruction, exactly as `progress.md` §E.1 recorded at plan time.
- **The `manual`-mode profile was not rendered.** `manual=0` is a schema-side count, not evidence
  about a manual-mode project's generated `git-strategy.yaml`.
- **CI has not run.** Nothing was pushed; no CI signal exists for this branch at the time of this
  report, and none is claimed.
- **The commit was made but not pushed and not merged**, per the dispatch constraint. No merge-window
  or integration-lock command was issued.
- **`moai spec lint` was run from the freshly built `bin/moai`**, not the installed `~/go/bin/moai`.
  The installed binary's behaviour on this SPEC was not observed.

## 5. Residual risk

- **AC-TSA-007b compares parsed JSON**, so a pure reformat of untouched regions (re-indentation, key
  reordering) would still read `True`. Mitigations held in practice — the edit was two literal
  deletions via exact-match string replacement, no serializer round-trip, and §2.4's diff shows two
  contiguous deletion hunks and zero additions — but no criterion mechanically rejects that case.
- **AC-TSA-001's predicate has no executable oracle.** With no runtime consumer, the exact-integer
  counts rest on a reconstruction of interview semantics. The risk stays bounded by the change being
  purely subtractive: it removes questions bound to a path no struct field matches, which holds under
  any interview semantics.
- **AC-TSA-005b's `0` is attributable only while the dead-path strings stay confined to the schema.**
  Re-measured `0` elsewhere in `internal/`, `pkg/`, `cmd/` today; a future change introducing those
  strings into other compiled code would silently break the attribution and the criterion would need
  re-scoping.
- **The control count is `4`, sourced from two template files.** Should
  `.claude/skills/moai/workflows/sync/delivery.md` later drop its mention, the control would fall to
  `3` — still `>= 1`, so the criterion survives, but a reader comparing counts across runs should
  know the number is not schema-exclusive.
- **The schema's self-declared counters stay wrong**, and the `total_settings` drift widens from
  `60 vs 48` to `60 vs 46`. Recorded out of scope in `spec.md` §4 so the widened number is not later
  attributed to this deletion.
- **The evidence scripts and `spec-lint.txt` live under `.moai/state/verify/t316/`, which is
  gitignored** — they will not survive worktree disposal. The durable evidence is this file plus
  `progress.md` §E.2; the recipes themselves are reproduced verbatim in `acceptance.md`.

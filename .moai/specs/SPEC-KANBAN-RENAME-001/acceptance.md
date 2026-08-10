---
id: SPEC-KANBAN-RENAME-001
title: "Acceptance criteria — Factory Mode to Kanban Mode rename"
version: "0.3.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: Medium
phase: "v3.1.0 target"
module: cli
lifecycle: spec-anchored
tags: "rename, refactor, cli, template-mirror, behavior-preserving"
tier: L
---

## §A. How these criteria are judged

Every criterion below names the command that decides it. A criterion is PASS only when that command was actually run in this tree and its output observed.

### A.1 Command-authoring rules (learned failure modes)

- **No `\|` inside a table cell fed to `grep -E`.** A markdown-escaped pipe is a literal in ERE and silently matches nothing, producing a vacuous GREEN. Every multi-alternative pattern here lives in a fenced block, never in a table cell.
- **Alternation and word-boundary escapes require `grep -E` with a plain `|`.** `\|` and `\b` are GNU BRE extensions. They work under this machine's `ugrep`-backed `grep`, and they silently mean something else under BSD `grep` — `\|` becomes a literal backslash-pipe, matching nothing. A criterion written that way is portable-looking and vacuous on a different machine. Every alternation here uses `-E` with `|`; no criterion relies on `\b`.
- **Never read `$?` after a pipe.** `cmd | tail -20; echo "exit=$?"` reports `tail`'s status, not `cmd`'s — `sh -c 'echo FAIL; exit 1' | tail -20; echo "exit=$?"` prints `exit=0`. A fully red test suite would report PASS. Every command whose exit code decides a criterion writes its output to a file and reads `$?` **before** any pipe; the log is then grepped and tailed separately.
- **Bounded tails do not substitute for a full-log scan.** `go list ./...` reports 115 packages, so `tail -20` hides roughly 95 lines of a full-suite run. A "no FAIL lines" assertion is made against the whole log file, never against its tail.
- **A `git diff` that decides a criterion is anchored to a ref.** Bare `git diff` compares the working tree to the index, so after the work is committed it is unconditionally empty — a criterion built on it passes whether or not the excluded path was touched. Every diff-based criterion here is anchored to the baseline `d39e3cdc6..HEAD`.
- **A `go test -run` pattern that matches nothing exits 0 and prints `PASS`.** Measured on `go1.26.4`: `go test ./internal/factory/ -run 'ZZZNoSuchTestName' -v; echo $?` prints `testing: warning: no tests to run`, then `PASS`, then `ok … [no tests to run]`, and exits **0**. Every criterion keyed on a **post-rename** test name is therefore satisfiable by an implementer who renames the production identifier and leaves the test name alone — the run selects zero tests and reports green. Two guards, both applied to every `-run`-keyed criterion here: a name-existence `grep` on the test file, and an assertion that the literal `[no tests to run]` is **absent** from the run's log. That literal is the load-bearing one because it appears on the `ok` line in **both** `-v` and non-`-v` runs, whereas `testing: warning: no tests to run` is emitted only under `-v`; a criterion keyed on the warning alone would itself be vacuous the moment `-v` were dropped.
- **The shipped guard is the authority.** Where a guard already exists (neutrality, namespace leak), the criterion runs the guard. A hand-rolled regex without the guard's exemption list is a false-failure machine.
- **Negative criteria need a positive control.** A criterion asserting "returns zero" is paired with evidence that the same command returned non-zero at baseline; otherwise a typo in the pattern also returns zero. Every positive control below was measured in this tree, not estimated.
- **`$TOK` is the token pattern from `spec.md` §D.1**, defined once there and copied byte-identically into AC-KR-021. AC-KR-027 checks the two copies for drift.

---

## §B. Acceptance criteria

### B.1 Entry surface

**AC-KR-001** — Given a launcher invocation carrying `--kanban` or `-k`, When the flag parser runs, Then Kanban Mode is enabled, an optional following SPEC identifier is consumed, and both tokens are absent from the argv reaching the launcher seam.
```bash
grep -cE '^func TestParseKanbanFlag' internal/cli/cc_test.go
grep -cE '^func TestCC_KanbanFlagStripped' internal/cli/cc_test.go
go test ./internal/cli/ -run 'ParseKanbanFlag|KanbanFlagStripped' -v > /tmp/kr-ac001.log 2>&1; echo "exit=$?"
grep -cF '[no tests to run]' /tmp/kr-ac001.log
```
→ both name counts ≥ 1, `exit=0`, and the `[no tests to run]` count is `0`.

The two name greps and the `[no tests to run]` assertion are not redundancy — without them this criterion is **satisfiable by not doing the work** (§A.1). `-run` selects by test-function name, and the pattern here names the **post-rename** functions; an implementer who renames `parseFactoryFlag` → `parseKanbanFlag` but leaves `TestParseFactoryFlag_*` untouched selects zero tests, and the run exits 0 printing `PASS`. That matters beyond this one row: `REQ-KR-011` (rename the test function names) traces **only** to this criterion and `AC-KR-005`, so a vacuous pair here would leave that requirement with no criterion able to fail. Baseline at HEAD `d39e3cdc6` the pre-rename names are `TestParseFactoryFlag_{LongFormWithoutSpec,ShortFormWithSpec,PassThroughBoundary,SpecIsNotStolenFromAFlag}` and `TestCC_FactoryFlagStrippedBeforeLaunch`, all in `internal/cli/cc_test.go`, so both greps return `0` before the rename and non-zero after.

**AC-KR-002** — Given a launcher invocation carrying `--kanban` after a bare `--`, When the parser runs, Then Kanban Mode is NOT enabled and the tokens are forwarded verbatim.
`go test ./internal/cli/ -run 'PassThroughBoundary' -v` → PASS.

**AC-KR-003** — Given a launcher invocation carrying `--factory` or `-f`, When the parser runs, Then no mode is enabled and no deprecation notice is emitted; the tokens are treated as ordinary pass-through argv.
```bash
grep -nE -- '"--factory"|"-f"' internal/cli/kanban.go; echo "grep-exit=$?"
```
→ zero matches (`grep-exit=1`). Positive control: the same pattern against `internal/cli/factory.go` at HEAD `d39e3cdc6` matches the two flag-constant lines.

**AC-KR-004** — Given the `claude` CLI's own flag surface, When probed at M0, Then either no `-k` flag exists, or the collision is recorded in `progress.md` and surfaced as a blocker before M1 begins.
```bash
claude --help 2>&1 > /tmp/kr-claude-help.txt
grep -E '(^|[^-])-k[ ,]' /tmp/kr-claude-help.txt; echo "grep-exit=$?"
grep -oE '(^|[[:space:]])-[a-zA-Z][,[:space:]]' /tmp/kr-claude-help.txt | tr -d ' ,' | sort -u | tr '\n' ' '
```
→ verbatim output recorded in `progress.md`. This criterion is satisfied by *recording*, not by any particular outcome.

**Plan-phase result (spec.md §A.6): no collision.** `grep-exit=1`; the short-flag set is `-c -d -h -n -p -r -v -w`. M0 re-confirms this against the run-phase tree rather than discovering it, because the `claude` CLI surface drifts between versions. Stated limitation: the pattern matches `-k ` and `-k,` renderings and would miss a `-k=<value>` form, so a null result is strong evidence, not proof.

**AC-KR-005** — Given `moai cg` invoked with the Kanban switch, When the rejection path runs, Then an error carrying the literal `KANBAN_MODE_UNSUPPORTED_BACKEND` is returned and the launcher seam is never invoked.
```bash
grep -cE '^func TestCG_.*Kanban' internal/cli/cg_test.go
go test ./internal/cli/ -run 'CG_.*Kanban' -v > /tmp/kr-ac005.log 2>&1; echo "exit=$?"
grep -cF '[no tests to run]' /tmp/kr-ac005.log
```
→ the name count is ≥ 1, `exit=0`, and the `[no tests to run]` count is `0`.

Same guard and same reason as `AC-KR-001`: `CG_.*Kanban` is a **post-rename** name pattern, and a `-run` that matches nothing exits 0 printing `PASS` (§A.1). Baseline at HEAD `d39e3cdc6` the two functions are `TestCG_FactoryFlagRejected` and `TestCG_WithoutFactoryFlagStillLaunches` in `internal/cli/cg_test.go`, so the name grep returns `0` before the rename and `2` after.

### B.2 Identifier surface

**AC-KR-006** — Given the repository tree, When the package layout is inspected, Then `internal/kanban/` exists with four files and `internal/factory/` does not exist.
```bash
ls internal/kanban/ && test ! -d internal/factory && echo OK
```
→ four files listed, `OK` printed.

**AC-KR-007** — Given the repository tree, When the CLI file layout is inspected, Then `internal/cli/kanban.go` exists and `internal/cli/factory.go` does not.
```bash
test -f internal/cli/kanban.go && test ! -f internal/cli/factory.go && echo OK
```
→ `OK`.

**AC-KR-008** — Given `internal/config/envkeys.go`, When the constants are read, Then `EnvMoaiKanban == "MOAI_KANBAN"` and `EnvMoaiKanbanSpec == "MOAI_KANBAN_SPEC"`, and neither literal appears at any call site.
```bash
grep -nE 'EnvMoaiKanban|EnvMoaiKanbanSpec' internal/config/envkeys.go
grep -rn '"MOAI_KANBAN' --include='*.go' internal/ | grep -v 'internal/config/envkeys.go'
```
→ first command shows both constants with the expected values; second returns zero matches.

**AC-KR-009** — Given the state-record package, When the path segments are read, Then the directory is `.moai/state/kanban/`.
```bash
grep -n 'stateDirSegments' internal/kanban/record.go
go test ./internal/kanban/ -run 'Path' -v
```
→ segments show `"kanban"`; the path test PASSes against a `.moai/state/kanban/<session>.json` expectation.

**AC-KR-010** — Given the renamed package, When `captureEnvState` is searched for, Then it is present and unrenamed.
```bash
grep -n 'func captureEnvState' internal/cli/kanban.go
```
→ one match.

### B.3 Behavior preservation

**AC-KR-011** — Given the full test suite, When `go test ./...` is run, Then it exits 0 with no failing package.
```bash
go test ./... > /tmp/kr-test.log 2>&1; echo "exit=$?"
grep -c '^FAIL' /tmp/kr-test.log
tail -20 /tmp/kr-test.log
```
→ `exit=0` **and** the `FAIL` count is `0`, both read from the whole log. The tail is context for a human reader and decides nothing.

The exit code is captured before any pipe: `go test ./... 2>&1 | tail -20; echo "exit=$?"` reports `tail`'s status, so a fully red suite prints `exit=0`. The `FAIL` count is taken against the whole file for the same reason a tail cannot carry it — `go list ./...` reports 115 packages, so `tail -20` shows under a fifth of the run. Affected-package-only runs do NOT satisfy this criterion (REQ-KR-022).

**AC-KR-012** — Given the diff of the whole SPEC, When test files are inspected, Then no `t.Error`/`t.Fatal`/`want`/`if got` assertion line differs except for renamed identifiers and mode prose.
```bash
git diff --unified=0 d39e3cdc6..HEAD -- '*_test.go' \
  | grep -E '^[+-].*(t\.Error|t\.Fatal|want :?=|if got)' \
  | grep -viE 'kanban|factory'
git diff --unified=0 d39e3cdc6..HEAD -- '*_test.go' | grep -cE '^\+.*(t\.Error|t\.Fatal)'
git diff --unified=0 d39e3cdc6..HEAD -- '*_test.go' | grep -cE '^-.*(t\.Error|t\.Fatal)'
```
→ the first command returns zero matches, **and** the two counts are equal. A non-empty first result names an assertion that changed for a reason other than the rename and must be justified or reverted; unequal counts name an assertion that was added or deleted outright.

**Why the second and third commands exist.** The first command's trailing `grep -viE 'kanban|factory'` discards exactly the lines most likely to drift: every assertion in these tests references the flag, the environment variable, or the sentinel, so every assertion line carries one of those two words and the filter removes all of them. A deleted or weakened factory-related assertion is filtered out and invisible. That would be a hole rather than a nuisance, because `REQ-KR-013` (behavior preservation) traces only to this criterion and `AC-KR-011`, and a *weakened* assertion leaves the suite green by construction — so `AC-KR-011` cannot catch it either, and the requirement would have no criterion able to fail. The count comparison is filter-independent: a rename rewrites a line, contributing one `+` and one `-` and leaving the counts equal, while a deletion moves them apart. Measured baseline at HEAD `d39e3cdc6`: **226** `t.Error`/`t.Fatal` lines across the six surface test files (`cc_test.go` 60, `glm_test.go` 70, `record_test.go` 43, `revision_test.go` 22, `cg_test.go` 19, `launcher_blockcap_infinite_test.go` 12), so a net change of even one is a visible delta against a known total.

Anchored to `d39e3cdc6..HEAD` rather than a bare `git diff`: M1, M2, and M3 are independently committable (plan.md §F), so by the M4 sweep a ref-less diff compares an already-committed tree against itself and returns empty unconditionally — passing whether or not an assertion changed.

**AC-KR-013** — Given the test corpus, When `AC-FM-` identifiers are counted, Then the count equals the pre-rename baseline.
```bash
grep -rc 'AC-FM-' --include='*_test.go' internal/ | awk -F: '{s+=$2} END{print s}'
```
→ equals the M0-recorded baseline. Renaming them is a defect (REQ-KR-012).

### B.4 Harness documentation

**AC-KR-014** — Given both trees, When the contract document is located, Then `workflows/kanban.md` exists on the local and template sides and `workflows/factory.md` exists on neither.
```bash
for p in .claude internal/template/templates/.claude; do
  test -f "$p/skills/moai/workflows/kanban.md" && test ! -f "$p/skills/moai/workflows/factory.md" && echo "OK $p"
done
```
→ two `OK` lines.

**AC-KR-015** — Given the five sibling documents, When they are grepped for the old contract vocabulary, Then zero matches remain on either side.
```bash
grep -rn 'factory' \
  .claude/skills/moai/workflows/moai.md \
  .claude/skills/moai/workflows/run.md \
  .claude/skills/moai/workflows/run/mode-orchestration.md \
  .claude/skills/moai/workflows/sync/quality-gates-quality.md \
  .claude/rules/moai/workflow/goal-directive.md
```
→ zero matches, and the same command against the `internal/template/templates/.claude/` prefixes also returns zero.

**Positive control (measured, not estimated): the identical command at HEAD `d39e3cdc6` returns 8 matches across the five local files** — `run.md` ×1, `moai.md` ×1, `mode-orchestration.md` ×3 (lines 82, 84, 108), `quality-gates-quality.md` ×2 (lines 114, 129), `goal-directive.md` ×1. The v0.1.0 draft recorded 9; that figure was the count of *edit locations* in plan.md §F M2 step 3, which includes a `quality-gates-quality.md` heading that must be edited but contains no `factory` token. Nine edit locations, eight grep matches — the two figures count different things and are kept distinct.

**AC-KR-016** — Given the renamed contract document, When the goal preset is searched, Then `kanban_chain` appears and `factory_chain` does not.
```bash
grep -c 'kanban_chain' .claude/skills/moai/workflows/kanban.md
grep -c 'factory_chain' .claude/skills/moai/workflows/kanban.md
```
→ first ≥ 1, second = 0.

### B.5 Template mirror and build

**AC-KR-017** — Given the six mirrored pairs, When each `diff` is compared against its M0 baseline under `factory`→`kanban` substitution, Then every pair matches.
```bash
for pair in \
  "contract:skills/moai/workflows/kanban.md" \
  "run:skills/moai/workflows/run.md" \
  "goal:rules/moai/workflow/goal-directive.md" \
  "moaidoc:skills/moai/workflows/moai.md" \
  "modeorch:skills/moai/workflows/run/mode-orchestration.md" \
  "qgates:skills/moai/workflows/sync/quality-gates-quality.md"; do
  k="${pair%%:*}"; f="${pair#*:}"
  test -f ".claude/$f" || { echo "MISSING .claude/$f"; continue; }
  test -f "/tmp/base-$k.diff" || { echo "MISSING baseline /tmp/base-$k.diff"; continue; }
  diff ".claude/$f" "internal/template/templates/.claude/$f" > "/tmp/after-$k.diff"
  sed 's/factory/kanban/g; s/Factory/Kanban/g' "/tmp/base-$k.diff" \
    | diff - "/tmp/after-$k.diff" && echo "OK $k"
done
```
→ six `OK` lines, and zero `MISSING` lines.

The pair list is written out literally rather than held in a shell array: an undefined array yields one empty iteration under zsh and zero under bash, so a criterion that reads `"${PAIRS[@]}"` is not runnable as written (§A.1). The `label:path` form keeps the baseline key stable across the rename — M0 captures `/tmp/base-contract.diff` from the pre-rename `factory.md`, and this check reads that same key after the file has become `kanban.md`, so the basename change cannot orphan the baseline. The two `test -f` guards exist because a `diff` against a missing path writes to stderr and contributes nothing to the output, which would otherwise let a rename that landed at the wrong path read as clean. This is the delta-preservation check (REQ-KR-018); a pair that "became identical" is a failure, not an improvement — it means §25-forbidden content was copied into template source.

**AC-KR-018** — Given the renamed template contract document, When it is scanned for internal content, Then it carries no SPEC identifier.
```bash
grep -cE 'SPEC-[A-Z0-9-]+-[0-9]{3}' internal/template/templates/.claude/skills/moai/workflows/kanban.md
```
→ `0`. In particular the string `SPEC-KANBAN-RENAME-001` must be absent.

**AC-KR-019** — Given the template tree, When the shipped neutrality and namespace guards are run, Then they pass.
```bash
go test ./internal/template/... > /tmp/kr-template.log 2>&1; echo "exit=$?"
grep -c '^FAIL' /tmp/kr-template.log
tail -20 /tmp/kr-template.log
```
→ `exit=0` **and** a `FAIL` count of `0`, both read from the whole log — same reasoning as AC-KR-011: an exit code read after a pipe is `tail`'s, not the test command's. The shipped guard decides this; no re-implemented regex substitutes for it.

**AC-KR-020** — Given `make build` has run, When `catalog.yaml` is diffed, Then the `moai` skill's `hash:` field has changed and the file is committed.
```bash
make build
git diff --stat d39e3cdc6..HEAD -- internal/template/catalog.yaml
git diff d39e3cdc6..HEAD -- internal/template/catalog.yaml | grep -c '^[+-].*hash:'
```
→ `catalog.yaml` appears in the stat output and the hash-line count is ≥ 2 (one removed, one added). An empty diff means either the build did not run or the rename did not reach template source — both are failures.

Anchored to `d39e3cdc6..HEAD` at v0.3.0, for the reason §A.1 already states and this criterion alone had not applied. The v0.2.0 form used a **ref-less** `git diff`, which compares the working tree to the index and is therefore empty the moment `catalog.yaml` is committed — and `plan.md` M3 step 2 instructs exactly that commit. Read at M4 the criterion would report no stat line and a hash count of `0`, i.e. FAIL, while its own text asserts "and the file is committed": the two halves contradicted each other, and the criterion was unsatisfiable by construction at the moment it runs. `AC-KR-012`, `AC-KR-023`, and `AC-KR-024` were anchored for the same hazard at v0.1.1 and this one was skipped; re-checked at v0.3.0, no other criterion carries the ref-less form. The bare `git diff --stat internal/template/catalog.yaml` in `plan.md` M3 step 2 is **not** the same defect and stays as written — it runs immediately after `make build`, before the commit, where the working-tree comparison is the correct one.

### B.6 Completion

**AC-KR-021** — Given the whole tree, When the §D.1 token grep runs, Then it returns zero files.
```bash
TOK='MOAI_FACTORY|EnvMoaiFactory|factoryFlag|factoryUnsupportedBackend|FACTORY_MODE_UNSUPPORTED_BACKEND|[Ff]actory [Mm]ode|--factory|parseFactoryFlag|enterFactoryMode|recordFactorySession|rejectFactoryOnCG|internal/factory|factory_chain|workflows/factory|state/factory|package factory|[Ff]actory (contract|chain|dedup|verify stage|session|state record|pipeline)'
grep -rlniIE "$TOK" internal/ .claude/ .moai/project/ | wc -l
```
→ `0`. **Positive control** (required, else the criterion is vacuous): the identical command at HEAD `d39e3cdc6` returns `28` — 26 under `internal/` + `.claude/`, plus `.moai/project/codemaps/modules.md` and `.moai/project/structure.md` (spec.md §A.5). The scope covers `internal/template/templates/` implicitly and excludes `.moai/specs/` deliberately.

**AC-KR-022** — Given the same pattern, When it is run against trees this SPEC does not touch, Then it still returns zero — confirming the pattern excludes unrelated "factory" vocabulary rather than the rename having swept it.
```bash
grep -rlniIE "$TOK" internal/lsp internal/tui internal/hook internal/core docs-site/ | wc -l
```
→ `0` both before and after. Meanwhile `grep -rli factory internal/lsp | wc -l` remains non-zero, proving the unrelated vocabulary was left alone.

**AC-KR-023** — Given the docs-site tree, When it is diffed against the SPEC's baseline commit, Then it is untouched.
```bash
git diff --stat d39e3cdc6..HEAD -- docs-site/
```
→ empty. This SPEC commissions no docs-site work (spec.md §A.3). Anchored to `d39e3cdc6..HEAD` because a bare `git diff` is empty by construction once the work is committed, and would pass whether or not `docs-site/` had been touched.

**AC-KR-024** — Given `.moai/specs/SPEC-FACTORY-MODE-001/`, When it is diffed against the SPEC's baseline commit, Then it is untouched.
```bash
git diff --stat d39e3cdc6..HEAD -- .moai/specs/SPEC-FACTORY-MODE-001/
```
→ empty. It is a closed historical record (spec.md §C). Same anchoring reason as AC-KR-023.

**AC-KR-025** — Given the CLI binary, When it is smoke-tested, Then `moai --version` and `moai cc --help` each exit 0 and render their output.
```bash
./bin/moai --version > /tmp/kr-version.txt 2>&1; echo "version-exit=$?"
./bin/moai cc --help > /tmp/kr-cchelp.txt 2>&1; echo "help-exit=$?"
wc -l /tmp/kr-version.txt /tmp/kr-cchelp.txt
```
→ both exits are `0` and both files are non-empty.

**The help-text `factory` grep was removed at v0.3.0 because it was vacuous, and its vacuity was concealed by a missing positive control.** §A.1 requires every negative criterion to be paired with evidence that the same command returned non-zero at baseline; this one carried none, and the measurement explains why it could not. `moai cc --help` **never documented the flag**: `internal/cli/cc.go` renders a `Long` block whose flag list is `-p/--profile`, `--permission-mode`, `-b/--bypass`, `-c/--continue`, `-m/--model`, `-w/--worktree`, `--spawn`, and `--chrome/--no-chrome`, with no mode entry switch, and the seven case-insensitive `factory` occurrences in that file are one import, three comment lines, and three call lines — all outside the help string. Run at HEAD `d39e3cdc6`, `go run ./cmd/moai cc --help 2>&1 | grep -ci factory` returns **`0`**, which is the same value the post-rename tree returns. A check that returns the PASS value before the work is done decides nothing.

Making it discriminating was rejected: it would require the help text to name the entry switch, which is a **new obligation** on the launcher, and this SPEC sits at the Tier L requirement ceiling exactly (`spec.md` §B) — no twenty-sixth requirement is available. What survives is the smoke this criterion can actually decide, which is what `REQ-KR-023` needs from it: the binary builds and both entry points still work. The `-k` surface itself is decided by `AC-KR-001`, and the absence of the old tokens from source by `AC-KR-003` and `AC-KR-021`.

**AC-KR-026** — Given the six contract documents on both sides, When they are scanned for a bare `-f` short flag, Then zero occurrences remain.
```bash
for p in .claude internal/template/templates/.claude; do
  for f in \
    skills/moai/workflows/kanban.md \
    skills/moai/workflows/moai.md \
    skills/moai/workflows/run.md \
    skills/moai/workflows/run/mode-orchestration.md \
    skills/moai/workflows/sync/quality-gates-quality.md \
    rules/moai/workflow/goal-directive.md; do
    grep -nE '(^|[^-[:alnum:]])-f($|[^[:alnum:]])' "$p/$f"
  done
done | wc -l
```
→ `0`.

The file list is a literal `for` list, not a `$DOCS` variable. Under `zsh` an unquoted `$DOCS` does not word-split, so the loop passes the whole string as one filename, `grep` reports "No such file or directory" on stderr, and the count comes back `0` — which is the PASS value. That exact failure was observed while authoring this criterion; the literal list returns the expected 8 at baseline.

**Positive control (measured): 8 occurrences at HEAD `d39e3cdc6`** — `workflows/factory.md` ×2, `workflows/moai.md` ×1, `rules/moai/workflow/goal-directive.md` ×1, and the identical three files on the template side. (At baseline the first path is `factory.md`, not `kanban.md`.)

This criterion exists because `$TOK` cannot carry a bare `-f` alternative — it would match `rm -f`, `grep -f`, and `git commit -f` tree-wide — so an implementer who renames `--factory` → `--kanban` and leaves `-f` in the prose passes AC-KR-015 and AC-KR-021 while shipping docs that advertise a dead flag (REQ-KR-025).

**AC-KR-027** — Given the two copies of the token pattern, When they are compared, Then they are byte-identical.
```bash
sed -n "/^TOK='/p" .moai/specs/SPEC-KANBAN-RENAME-001/spec.md > /tmp/kr-tok-spec.txt
sed -n "/^TOK='/p" .moai/specs/SPEC-KANBAN-RENAME-001/acceptance.md > /tmp/kr-tok-acc.txt
diff /tmp/kr-tok-spec.txt /tmp/kr-tok-acc.txt && echo OK
```
→ `OK`. The v0.1.0 draft's two copies differed (the spec.md copy carried an extra backtick-delimited alternative); both returned 26 so the drift was inert, but a pattern defined twice is a pattern that will eventually disagree with itself. `spec.md` §D.1 is the definition; this criterion is the drift check.

**AC-KR-028** — Given the two project documents, When they are read after the rename, Then they name `internal/kanban`, `moai cc -k` / `moai glm -k`, and `internal/cli/kanban.go`, and no longer name their pre-rename counterparts.
```bash
grep -nE 'internal/kanban|kanban\.go|-k' .moai/project/codemaps/modules.md .moai/project/structure.md
grep -rlniIE "$TOK" .moai/project/ | wc -l
grep -niI factory .moai/project/codemaps/modules.md .moai/project/structure.md | wc -l
```
→ the first command shows the renamed package section and entry-point line; the second and third both return `0`. **Positive controls, all three measured at HEAD `d39e3cdc6`:** the token grep matches **2 files** (`-l`) across **3 lines** (`codemaps/modules.md` 157 and 246, `structure.md` 139), and the bare-word grep over the same two files matches **5 lines** (`modules.md` 157, 158, 161, 246 and `structure.md` 139). Neither file is template-mirrored, so the delta-preservation invariant of AC-KR-017 does not apply to them (REQ-KR-024).

**Two v0.3.0 repairs, and the second closes a gap this SPEC had disclosed rather than fixed.**

The `-l` was missing. The v0.2.0 command counted **lines** while its control claimed "2 files"; re-measured, the ref-less form returns **3**, so the criterion compared a line count against a file count and would have failed a correct rename by one. The `-l` restores the agreement, and the control now states all three figures so the next reader cannot mistake which is which.

The third command is new. `$TOK` matches this file at **lines 157 and 246 only**: line 158 carries `Factory 모드` — the mode phrase in Korean, which `[Ff]actory [Mm]ode` does not match — and line 161 carries the path `internal/cli/factory.go`, which `internal/factory` does not match (`research.md` §H.4). At **file** granularity the token check is sound, since 157 and 246 hold the file in the match set. At **line** granularity it is not: an implementer who fixes only the two `$TOK`-matching lines drives the second command to zero with two stale lines still naming a deleted package and a dead flag, and the first command — a human-read grep for the renamed forms — would show the renamed section heading beside them. The bare word is the right instrument at this scope: the tree-wide false-positive objection that forces `$TOK` to be token-scoped (`spec.md` §D.1, ~110 files of unrelated vocabulary) simply does not arise across **two named files**, where all five baseline matches are Kanban Mode and none is generic pattern vocabulary. `research.md` §H.4 and `design.md` §F.3 carried this as disclosed residual risk on the reasoning that closing it meant editing a criterion; at v0.3.0 the criterion is being edited anyway, and a two-file bounded grep adds no criterion and no requirement.

---

## §C. Definition of Done

- All 28 criteria PASS (`AC-KR-001` … `AC-KR-028`), each with its deciding command run in this tree and its output quoted.
- The §E 5-section evidence report is written into `progress.md` §E.2, with a Gaps section that is either empty-and-justified or names what was not observed.
- No commit touches `/Users/goos/MoAI/moai-adk-go/`.
- `SKIP_MOAI_PRECOMMIT=1` was used for each commit, and that fact is recorded rather than silently relied on.

## §D. Traceability

| REQ | Criteria |
|---|---|
| REQ-KR-001 | AC-KR-001, AC-KR-002 |
| REQ-KR-002 | AC-KR-003 |
| REQ-KR-003 | AC-KR-004 |
| REQ-KR-004 | AC-KR-006 |
| REQ-KR-005 | AC-KR-007 |
| REQ-KR-006 | AC-KR-010 |
| REQ-KR-007 | AC-KR-008 |
| REQ-KR-008 | AC-KR-005 |
| REQ-KR-009 | AC-KR-009 |
| REQ-KR-010 | AC-KR-009 (no migration asserted; old path simply absent from the code) |
| REQ-KR-011 | AC-KR-001, AC-KR-005 |
| REQ-KR-012 | AC-KR-013 |
| REQ-KR-013 | AC-KR-011, AC-KR-012 |
| REQ-KR-014 | AC-KR-014 |
| REQ-KR-015 | AC-KR-015 |
| REQ-KR-016 | AC-KR-016 |
| REQ-KR-017 | AC-KR-017 |
| REQ-KR-018 | AC-KR-017 |
| REQ-KR-019 | AC-KR-018, AC-KR-019 |
| REQ-KR-020 | AC-KR-020 |
| REQ-KR-021 | AC-KR-021, AC-KR-022, AC-KR-027 |
| REQ-KR-022 | AC-KR-011 |
| REQ-KR-023 | AC-KR-011, AC-KR-012, AC-KR-023, AC-KR-024, AC-KR-025 |
| REQ-KR-024 | AC-KR-028 |
| REQ-KR-025 | AC-KR-026 |

### D.1 Reconciliation and the tier budget

Twenty-eight criteria against twenty-five requirements. Every requirement maps to at least one criterion and every criterion appears in the table above, verified by enumeration rather than by counting rows — several requirements are decided by more than one criterion (`REQ-KR-023` by five) and several criteria decide more than one requirement (`AC-KR-011` for three), so the two counts are not expected to agree and neither is derived from the other.

**Tier L: 28 of 25 criteria against 25 of 25 requirements.** The requirements sit **at the ceiling exactly, with no headroom** — a twenty-sixth requirement forces a split of this SPEC rather than fitting in it. The criteria are **3 over**, and that overage is carried as a **disclosed debt** for the plan auditor to rule on rather than absorbed, because Tier L is the top tier and no further promotion exists.

Folding three criteria back to reach 25 was considered and rejected. The three cheapest merge candidates are exactly the three the v0.1.1 audit added or separated for cause: `AC-KR-026` (the bare-`-f` residue) exists precisely because `$TOK` structurally cannot see it (§D.1.1 of `spec.md`), so merging it into `AC-KR-015` or `AC-KR-021` restores the blind spot it was created to cover; `AC-KR-027` (the `$TOK` drift check) decides a property of the SPEC documents rather than of the tree, and folding it into `AC-KR-021` would make a pattern check its own copy; and `AC-KR-028` (the two project documents) covers a surface outside the enumerated 26 that the v0.1.0 scope missed entirely. Each merge would trade a disclosed count for an undisclosed gap, which is the worse of the two.

Measured at v0.2.0 authoring time, and re-measurable:

```
$ grep -cE '^\*\*REQ-KR-[0-9]{3}\*\*' spec.md
25
$ grep -cE '^\*\*AC-KR-[0-9]{3}\*\*' acceptance.md
28
```

No criterion is added, removed, renumbered, or reworded by the promotion. At Tier M these same figures were nine and twelve over their ceilings and no artifact said so — the silence, not the count, is what v0.2.0 repairs.

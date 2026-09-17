# Acceptance — SPEC-GTD-CANON-BODY-001 (card t867)

All commands run from the worktree root with bash. `$BIN` is a moai binary built from the tree under test (`go build -o <scratch>/moai ./cmd/moai`). `$SPEC` = `.moai/specs/SPEC-GTD-CANON-BODY-001`. Document-level tree pin for RED-now cells: working tree of `114737ea1` (the SPEC scripts added on top are untracked at measurement).

## Allowed-survivor rule

A line in a scoped file that matches `moai todo|/moai:todo` (which also covers `/moai todo`) is an **allowed survivor** only when the same line contains the phrase `compat alias` preceded by start-of-line or a non-letter (regex `(^|[^A-Za-z])compat alias`). `compatible`, `incompatible` and `incompat alias` do not qualify. Any other match is a residual defect. The rule applies only to the 19 files listed in `$SPEC/check-residual.sh`.

The `workflows/todo.md` citation check (AC-GCB-005) covers `.claude internal cmd pkg .moai/docs Makefile` with zero survivors. Historical records (`CHANGELOG.md`, top-level `reports/`, `.moai/reports/`, `.moai/specs/`) are outside its roots and are inventoried instead (AC-GCB-011).

## Evidence ledger (RED-now + controls)

| Id | Command | Exit | Stdout (verbatim tail) | Tree |
|---|---|---|---|---|
| L1 | `bash $SPEC/check-residual.sh` | 1 | `files=19` / `grep_rc=0 offending=69 survivors=0` / `FAIL` | 114737ea1 |
| L2 | `bash $SPEC/check-residual.sh clean.md` (scratch: one line "`moai todo` is the compat alias of `moai gtd`") | 0 | `grep_rc=0 offending=0 survivors=1` / `PASS` | scratch |
| L3 | `bash $SPEC/check-residual.sh clean.md planted.md planted2.md` (planted: "(incompatible alias)", "(incompat alias)") | 1 | `offending=2 survivors=1` / both planted lines / `FAIL` | scratch |
| L4 | `bash $SPEC/check-residual.sh missing.md` | 2 | `GREP_ERROR rc=2` | scratch |
| L5 | `bash $SPEC/check-gtd-body.sh $BIN` | 1 | `fails=45` (all 45 doc-side; every help-side check PASS) | 114737ea1 |
| L6 | `bash $SPEC/check-gtd-body.sh $BIN good.md` (scratch control doc: all headings, stages section, 277 lines) | 0 | `fails=0` (108 PASS lines) | scratch |
| L7 | `bash $SPEC/check-gtd-body.sh $BIN mut.md` (control plus `--bogus` flag plus "the sixth stage") | 1 | `CHECK 002-no-invented-flag:--bogus FAIL` / `CHECK 002-answer-not-stage FAIL` | scratch |
| L8 | `grep -rn 'workflows/todo\.md' .claude internal cmd pkg .moai/docs Makefile` piped to `wc -l` | 0 | `21` | 114737ea1 |
| L9 | `grep -n '### todo - Backlog Queue' .claude/skills/moai/SKILL.md` | 0 | `169:### todo - Backlog Queue` | 114737ea1 |

## §D AC Matrix

### AC-GCB-001 — gtd.md carries the full body (REQ-GCB-001)

Given the migrated tree, When `bash $SPEC/check-gtd-body.sh $BIN` runs, Then every `CHECK 001-*` line reads PASS: 9 H2 headings (the 8 migrated ones plus `GTD stages`), `### What the analyser may do`, and `001-floor` with lines ≥ 267 (279-line todo body at `f67d2193f` minus the 12-line stub).
- RED-now: L5 (the `001-*` checks fail on the stub). Green path: M1.

### AC-GCB-002 — GTD stages per verb, flags from help, answer not a stage (REQ-GCB-002, REQ-GCB-003)

Given the migrated tree and `$BIN`, When `bash $SPEC/check-gtd-body.sh $BIN` runs, Then every `CHECK 002-*` line reads PASS:
- one usage-shape assertion per verb (`moai gtd capture <text>`, `clarify <gtd-id>`, `organize <gtd-id>`, `reflect`, `engage <gtd-id>`, `answer <t-id> <text>`), each also present in that verb's `--help`;
- each of the 23 required flags (every non-`--json`, non-`--help` flag of capture/clarify/organize/reflect/engage) listed in help AND named in the section, with `002-required-count` = 23 and `002-doc-flag-count` ≥ 23;
- every flag token in the section present in the union of the six verbs' help;
- both queue-boundary phrases present in the section and in `moai gtd --help`;
- `gate-blocked` present, and zero "six/sixth stage" mentions.

The script exits 0.
- RED-now: L5. Mutant controls: L6 (green is reachable), L7 (an invented flag and a sixth-stage claim both go red). Green path: M1.

### AC-GCB-003 — todo body deleted, no stub (REQ-GCB-004)

Given the migrated tree, When `test ! -e .claude/skills/moai/workflows/todo.md` and `test ! -e internal/template/templates/.claude/skills/moai/workflows/todo.md` each run, Then both exit 0.
- RED-now: both exit 1 on 114737ea1 (the file exists, L8 cites it). Green path: M4.

### AC-GCB-004 — SKILL.md gtd section (REQ-GCB-005)

Given the migrated tree, Then all of the following hold:
- `grep -n '### todo - Backlog Queue' .claude/skills/moai/SKILL.md internal/template/templates/.claude/skills/moai/SKILL.md` exits 1.
- `grep -c 'CLAUDE_SKILL_DIR}/workflows/gtd.md' .claude/skills/moai/SKILL.md` prints ≥ 1.
- `grep -c '\.claude/skills/moai/workflows/gtd\.md' internal/template/templates/.claude/skills/moai/SKILL.md` prints ≥ 1.
- `grep -c 'CLAUDE_SKILL_DIR' internal/template/templates/.claude/skills/moai/SKILL.md` prints 0.
- `go test -timeout 10m -count=1 -v -run '^TestSkillTreeHasNoClaudeSkillDirToken$' ./internal/template/` prints exactly one `--- PASS:` line and no `--- FAIL:`.
- In both trees a line containing `compat alias` names `/moai todo` and `moai todo`.
- The backlog-language Intent Router line contains `**gtd**`.
- RED-now: L9. Green path: M2.

### AC-GCB-005 — no residual workflows/todo.md citation (REQ-GCB-006, REQ-GCB-008)

Given the migrated tree, When `grep -rn 'workflows/todo\.md' .claude internal cmd pkg .moai/docs Makefile` runs, Then it prints nothing and exits 1 (exit 2 is a FAIL). Every Go citation, comments included, is updated; there is no survivor exception.
- RED-now: L8 (21 lines). Green path: M3 + M4.

### AC-GCB-006 — residual todo wording is alias documentation only (REQ-GCB-006)

Given the migrated tree, When `bash $SPEC/check-residual.sh` runs, Then it prints `files=19`, `offending=0` and `PASS`, and exits 0. Exit 2 (grep error, for example a missing file) is a FAIL. The script's detection power is proven by L2-L4.
- RED-now: L1 (69 offending lines). Green path: M1 + M2 + M4.

### AC-GCB-007 — compat aliases keep working (REQ-GCB-007)

Given the migrated tree and `$BIN`, Then all of the following hold:
- `$BIN todo list --help` exits 0.
- `grep -cE 'gtd.*\$ARGUMENTS' .claude/commands/moai/todo.md internal/template/templates/.claude/commands/moai/todo.md internal/template/templates/.agents/skills/moai-todo/SKILL.md` prints ≥ 1 for each of the three files.
- `go test -timeout 10m -count=1 -v -run '^TestTodoBareInvocationLists$' ./internal/cli/` prints exactly one `--- PASS:` line and no `--- FAIL:`.

`TestGTDCanonicalSurfaceGolden` and `TestGTDAllTodoVerbsParity` are pre-existing red (spec.md §C.3). They are judged only by AC-GCB-010's baseline delta, never as must-pass here.
- RED-now: not applicable. This is a preserved-behavior guard (regression-guard class) and holds on 114737ea1; its failure direction is a migration that breaks the alias dispatch line.

### AC-GCB-008 — retargeted tests read gtd.md and pass by name (REQ-GCB-008)

Given the migrated tree, Then all of the following hold:
- `grep -nE 'workflows", "todo\.md|workflows/todo\.md' internal/cli/todo_skill_doc_test.go internal/cli/todo_landed_doc_test.go internal/cli/doc_json_shape_test.go internal/template/backlog_json_disclosure_mirror_test.go` exits 1.
- `go test -timeout 10m -count=1 -v -run '^(TestTodoSkillDocumentsHistoryVerb|TestTodoDoctrine_MirrorParityAndStatedColumnCount|TestTodoListJSONShapeMatchesDoc)$' ./internal/cli/` prints exactly 3 `--- PASS:` lines and 0 `--- FAIL:`.
- `go test -timeout 10m -count=1 -v -run '^TestBacklogJSONDisclosure_' ./internal/template/` prints exactly 2 `--- PASS:` lines and 0 `--- FAIL:`.
- `$BIN todo --help` output does not contain `workflows/todo.md`.

A PASS-line count below the stated number (an empty or partial sweep) is a FAIL.
- Green path: M3 (tests retargeted) + M1 (the doc rows they read exist).

### AC-GCB-009 — generated artifacts regenerated, mirror parity (REQ-GCB-009)

Given the migrated tree, Then all of the following hold:
- `make build` exits 0 (agents-emit-check and commands-emit-check pass; `gen-catalog-hashes.go --all` runs).
- `git status --porcelain internal/template/catalog.yaml internal/template/templates/.codex/agents/moai/manager-lead.toml` prints nothing after the final commit, and `git log --oneline f67d2193f..HEAD -- internal/template/catalog.yaml` lists at least one commit.
- `go test -timeout 10m -count=1 -v -run '^(TestCatalogHashCoversSkillSubfiles|TestAllSkillsInCatalog|TestAllAgentsInCatalog)$' ./internal/template/` prints 3 `--- PASS:` lines and 0 `--- FAIL:`.
- Every `CHECK 009-cmp:*` line of `check-gtd-body.sh` reads PASS (six local/mirror pairs byte-identical: gtd.md, kanban-dispatch.md, kanban-dispatch-detail.md, moai-kanban-foreman/SKILL.md, project/doc-generation.md, todo-queue-storage.md).
- Green path: M4.3 + M5.1.

### AC-GCB-010 — scoped tests: no NEW failure versus baseline (REQ-GCB-013)

Procedure, identical at M0 (pre-edit HEAD) and M5 (migrated HEAD); each package runs in its own slot-held invocation:

1. `moai slot acquire --resource go-test --max-duration 40m`
2. `go test -timeout 30m -count=1 ./internal/template/... > .moai/reports/t867/<phase>-template.log 2>&1`, then record the exit code
3. `go test -timeout 30m -count=1 ./internal/cli/... > .moai/reports/t867/<phase>-cli.log 2>&1`, then record the exit code
4. `moai slot release --resource go-test`
5. Per log: `grep -E '^\s*--- FAIL: ' <log> | awk '{print $3}' | sort -u > <log>.fail-names`

Given the M0 and M5 name sets, Then:
- `comm -13 <M0>.fail-names <M5>.fail-names` prints nothing for each package (no new failing test name).
- Each M5 log contains at least one `ok ` or `FAIL\t` package line (non-empty sweep).
- An M5 log containing `panic: test timed out` is a **Gap, not a PASS**; re-run that package once under the slot.
- `go vet ./internal/cli/... ./internal/template/...` exits 0.
- No `go test ./...` invocation appears in progress.md §E.2.

The M0 name set is expected to match spec.md §C.3; any difference is recorded in the verdict. A test in the baseline set that turns green is reported, not required.
- Two-cell note: this criterion is a delta guard, not a RED-now criterion. Its baseline IS the red set, and it measures only this SPEC's contribution.

### AC-GCB-011 — config key, historical records, citation inventory (REQ-GCB-010, REQ-GCB-011, REQ-GCB-012)

Given the change set, Then all of the following hold:
- `git diff --name-only f67d2193f...HEAD -- CHANGELOG.md reports .moai/specs .moai/reports` lists only paths under `.moai/specs/SPEC-GTD-CANON-BODY-001/` or `.moai/reports/t867/`.
- `git diff f67d2193f...HEAD -- internal/config .moai/config internal/template/templates/.moai/config` prints nothing.
- `.moai/reports/t867/verdict.md` contains `workflow.todo.enabled`, lists the spec.md §C.3 pre-existing failures as a lead finding, and holds a citation inventory whose entry count equals the line count of `grep -rn 'workflows/todo\.md' CHANGELOG.md reports .moai/specs .moai/reports --exclude-dir=SPEC-GTD-CANON-BODY-001 --exclude-dir=t867`. The plan-time measurement was 13 + 132 = 145; the run phase re-measures.

## Edge Cases

- A `moai todo` line documenting the alias without the `compat alias` marker fails AC-GCB-006; reword the line rather than widening the rule.
- The section anchor `workflows/todo.md § Standing sources` is retargeted to the gtd.md section of the same name, which AC-GCB-001 guarantees exists.
- The deploy-time `.agents/skills/moai` mirror stops carrying `workflows/todo.md`; the template package run in AC-GCB-010 covers the mirror code path.

## Definition of Done

- AC-GCB-001 through AC-GCB-011 PASS, with the command, verbatim output and exit code in progress.md §E.2
- Template-First parity; `make build` clean; regenerated `catalog.yaml` and `manager-lead.toml` committed
- `.moai/reports/t867/verdict.md` written with the config-key finding, the pre-existing-failure finding, and the historical citation inventory
- Card id `t867` in every commit message

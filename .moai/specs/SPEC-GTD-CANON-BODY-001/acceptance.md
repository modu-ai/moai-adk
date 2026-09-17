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
| L6 | `bash $SPEC/check-gtd-body.sh $BIN good.md` (scratch control doc: all headings, stages section, 277 lines) | 0 | `fails=0`, 113 PASS lines (108 before the iteration-2 amendments, plus 5 `002-flag-set-equal:*`) | scratch |
| L7 | `bash $SPEC/check-gtd-body.sh $BIN mut.md` (control plus `--bogus` flag plus "the sixth stage") | 1 | `CHECK 002-no-invented-flag:--bogus FAIL` / `CHECK 002-answer-not-stage FAIL` | scratch |
| L10 | `zsh $SPEC/check-gtd-body.sh x` and `zsh $SPEC/check-residual.sh` | 2 and 2 | `USAGE_ERROR: run with bash (...)` | 6cce5e69c plus amendments |
| L11 | `bash $SPEC/check-gtd-body.sh $BIN mut3.md` (control whose answer line reads "`answer` is the 6th GTD stage after engage") | 1 | `CHECK 002-answer-not-stage FAIL six/sixth-stage mentions=1 want=0` | scratch |
| L12 | `bash $SPEC/check-gtd-body.sh <wrapper> good.md` (wrapper = real binary plus one extra `--zz-extra` line on `gtd engage --help`) | 1 | `CHECK 002-flag-set-equal:engage FAIL frozen=[… --run-id] help=[… --run-id --zz-extra]` | scratch |
| L8 | `grep -rn 'workflows/todo\.md' .claude internal cmd pkg .moai/docs Makefile` piped to `wc -l` | 0 | `21` | 114737ea1 |
| L9 | `grep -n '### todo - Backlog Queue' .claude/skills/moai/SKILL.md` | 0 | `169:### todo - Backlog Queue` | 114737ea1 |
| L13 | `go test -timeout 5m -count=1 -v -run '^TestGTDCanonicalSurfaceGolden$' ./internal/template/` | 1 | `gtd_canonical_surface_test.go:24: templates/.claude/commands/moai/todo.md is not a thin gtd compatibility path` / `--- FAIL: TestGTDCanonicalSurfaceGolden (0.00s)` | e4cc628e9 |
| L14 | `go test -timeout 5m -count=1 -v -run '^TestGTDAllTodoVerbsParity$' ./internal/cli/` | 1 | `gtd_compat_test.go:61: gtd verbs = [add analyze answer auto-done capture clarify done drop edit engage export-json history landed list move next organize pr reflect relate undone undrop unpick unrelate why], want [add analyze auto-done capture clarify done drop edit engage export-json history landed list move next organize pr reflect relate undone undrop unpick unrelate why]` / `--- FAIL: TestGTDAllTodoVerbsParity (0.00s)` | e4cc628e9 |

## §D AC Matrix

### AC-GCB-001 — gtd.md carries the full body (REQ-GCB-001)

Given the migrated tree, When `bash $SPEC/check-gtd-body.sh $BIN` runs, Then every `CHECK 001-*` line reads PASS: 9 H2 headings (the 8 migrated ones plus `GTD stages`), `### What the analyser may do`, and `001-floor` with lines ≥ 267 (279-line todo body at `f67d2193f` minus the 12-line stub).
- RED-now: L5 (the `001-*` checks fail on the stub). Green path: M1.

### AC-GCB-002 — GTD stages per verb, flags from help, answer not a stage (REQ-GCB-002, REQ-GCB-003)

Given the migrated tree and `$BIN`, When `bash $SPEC/check-gtd-body.sh $BIN` runs, Then every `CHECK 002-*` line reads PASS:
- one usage-shape assertion per verb (`moai gtd capture <text>`, `clarify <gtd-id>`, `organize <gtd-id>`, `reflect`, `engage <gtd-id>`, `answer <t-id> <text>`), each also present in that verb's `--help`;
- per verb, `002-flag-set-equal:<verb>` — the frozen flag list equals the set parsed from `moai gtd <verb> --help` (minus `--json`/`--help`) in BOTH directions, so a flag added to or removed from help turns red (L12);
- each of the 23 required flags listed in help AND named in the section, with `002-required-count` = 23 and `002-doc-flag-count` ≥ 23;
- every flag token in the section present in the union of the six verbs' help;
- both queue-boundary phrases present in the section and in `moai gtd --help`;
- `gate-blocked` present, and zero matches of `(six|sixth|6|6th) (gtd )?stage(s)` or `answer … (is|as) (a|an|the) … stage`. The body therefore describes `answer` with wording like "is not a stage", never "is a … stage".

The script runs under bash; any other shell is refused with exit 2 (L10). It exits 0 and prints **at least 113 `PASS` lines**, the count the L6 control produced. `fails=0` with fewer PASS lines means checks were skipped, and that is a FAIL.
- RED-now: L5. Mutant controls: L6 (green is reachable), L7 (an invented flag and a sixth-stage claim both go red), L11 (6th-stage wording goes red), L12 (help-side flag drift goes red). Green path: M1.

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

`TestGTDCanonicalSurfaceGolden` and `TestGTDAllTodoVerbsParity` are owned by this SPEC and must pass by name under AC-GCB-012.
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
- `go test -timeout 10m -count=1 -v -run '^(TestCatalogHashCoversSkillSubfiles|TestAllSkillsInCatalog|TestAllAgentsInCatalog|TestManifestHashFormat)$' ./internal/template/` prints 4 `--- PASS:` lines and 0 `--- FAIL:`. `TestManifestHashFormat` covers the `manager-lead` agent-file hash, which `TestCatalogHashCoversSkillSubfiles` skips.
- Every `CHECK 009-cmp:*` line of `check-gtd-body.sh` reads PASS (six local/mirror pairs byte-identical: gtd.md, kanban-dispatch.md, kanban-dispatch-detail.md, moai-kanban-foreman/SKILL.md, project/doc-generation.md, todo-queue-storage.md).
- Green path: M4.3 + M5.1.

### AC-GCB-010 — scoped tests: no NEW failure versus baseline (REQ-GCB-013)

Procedure, identical at M0 (pre-edit HEAD) and M5 (migrated HEAD). `<phase>` ∈ {`m0`, `m5`}, `<pkg>` ∈ {`template`, `cli`}. Every evidence file lives under `.moai/reports/t867/` with the single naming scheme `<phase>-<pkg>.log`, `<phase>-<pkg>.exit`, `<phase>-<pkg>.fail-names`. Each package runs under its own lease:

1. `moai slot acquire --resource go-test --max-duration 45m --command "go test <pkg>"`. The lease covers the per-package `-timeout 30m` plus a 15m margin, so it is never taken over mid-run.
   - exit 0: proceed.
   - exit 3 (held by another live session) or exit 4 (busy): run `moai slot status --resource go-test`, wait, and retry acquire. Allow at most 6 attempts spaced by the holder's remaining declared duration or 5 minutes, whichever is shorter. If every attempt fails, stop and record that package as a **Gap** (reason `slot-unavailable`) in the verdict.
   - Never run the package unleased, and never use `--force`.
2. `go test -timeout 30m -count=1 ./internal/<pkg>/... > .moai/reports/t867/<phase>-<pkg>.log 2>&1`, then record the exit code into `<phase>-<pkg>.exit`.
3. `moai slot release --resource go-test`, before acquiring for the next package.
4. `grep -E '^\s*--- FAIL: ' <phase>-<pkg>.log | awk '{print $3}' | sort -u > <phase>-<pkg>.fail-names`

Timeout handling applies to M0 and M5 alike. A log containing `panic: test timed out` is a **Gap**, not a result: re-run that package once under a fresh lease. If the re-run also times out, the package stays a Gap. A truncated M0 log is never authoritative as a baseline; any delta judged against it is reported as unmeasured.

Given `m0-<pkg>.fail-names` and `m5-<pkg>.fail-names` for both packages (neither a Gap), Then:
- `comm -13 m0-<pkg>.fail-names m5-<pkg>.fail-names` prints nothing for each package (no new failing test name).
- `grep -cxE 'TestGTDCanonicalSurfaceGolden|TestGTDAllTodoVerbsParity' m5-template.fail-names m5-cli.fail-names` prints `0` for both files. These two t867-owned names are never excused by the delta, even though they appear in m0.
- Each m5 log contains at least one `ok ` or `FAIL\t` package line (non-empty sweep).
- No m5 log contains a `[build failed]` or `[setup failed]` line, and no `FAIL\t<package>` line appears in m5 whose package line was absent at m0.
- `go vet ./internal/cli/... ./internal/template/...` exits 0.
- No `go test ./...` invocation appears in progress.md §E.2.

The M0 name set is expected to be the five t854-owned names in spec.md §C.3, plus the two t867-owned names in §C.4 while they are still red. Any other difference is recorded in the verdict. A test in the baseline set that turns green is reported, not required.
- Two-cell note: this criterion is a delta guard, not a RED-now criterion. Its baseline IS the red set, and it measures only this SPEC's contribution.

### AC-GCB-012 — t867-owned GTD surface tests pass by name (REQ-GCB-014)

Given the migrated tree, When each of these runs, Then each prints exactly one `--- PASS:` line, no `--- FAIL:` line, and no `[no tests to run]`, and exits 0:
- `go test -timeout 5m -count=1 -v -run '^TestGTDCanonicalSurfaceGolden$' ./internal/template/`
- `go test -timeout 5m -count=1 -v -run '^TestGTDAllTodoVerbsParity$' ./internal/cli/`

Neither test file may lose an assertion. `git diff <base>...HEAD -- internal/template/gtd_canonical_surface_test.go internal/cli/gtd_compat_test.go` must show no removed `t.Fatalf` / `t.Errorf` line without a replacement assertion of the same check.
- RED-now: L13 and L14. Green path: M3.7 (literal alignment) and M3.8 (option A: `"answer"` added to `gtdWant` only). Option A also requires `git diff f67d2193f...HEAD -- internal/cli` to show no change to any non-test `.go` file that registers todo or gtd subcommands, and `todoWant` in `gtd_compat_test.go` to stay byte-identical to base.
- M3.7 commit trace: `git log --format=%B -1 <M3.7 commit>` contains both `1dcaad954` and `61582178d`.

### AC-GCB-013 — isolated `moai todo answer t1 x` observation (REQ-GCB-015)

Isolation mechanism, run in this order with bash from the worktree root. `$T` comes from `mktemp -d` under the session scratch directory and is an absolute path.

1. `git init -q "$T/proj"`, then commit one empty commit there (`git -C "$T/proj" commit -q --allow-empty -m init`), so queue-root resolution sees a primary checkout.
2. Real-queue sentinel before: `ls -l ~/.moai/db/*/todo/backlog.db > "$T/real-before.txt" 2>&1` (read-only listing of size and mtime).
3. Positive control: `(cd "$T/proj" && MOAI_HOME="$T/moai-home" CLAUDE_PROJECT_DIR="$T/proj" $BIN todo add "control card")` must print `t1 1`. After it, `find "$T/moai-home" -name backlog.db` must print exactly one path, which proves the isolated store is the one written.
4. Observation: `(cd "$T/proj" && MOAI_HOME="$T/moai-home" CLAUDE_PROJECT_DIR="$T/proj" $BIN todo answer t1 x)`. Record stdout, stderr and the exit code.
5. Read-back: `(cd "$T/proj" && MOAI_HOME="$T/moai-home" CLAUDE_PROJECT_DIR="$T/proj" $BIN todo list --json)`, recording the `items` count.
6. Real-queue sentinel after: `ls -l ~/.moai/db/*/todo/backlog.db > "$T/real-after.txt" 2>&1`, then `cmp "$T/real-before.txt" "$T/real-after.txt"` must exit 0.

Given those steps, Then `.moai/reports/t867/verdict.md` records steps 3-6 verbatim and classifies the observation as exactly one of:
- **refused**: non-zero exit, and the items count in step 5 is 1 (control only);
- **added a card**: the items count in step 5 is 2. Recorded as a finding with the added card's text, and not fixed.

An observation without step 3's single `backlog.db` under `$T/moai-home`, or with a non-zero `cmp` in step 6, is a Gap, not a classification. If the operator's own sessions write the real queue during the run, step 6 can differ for reasons unrelated to this test; that is recorded as a Gap, and step 3's isolation evidence still stands. The shell guard may reject the subshell form; in that case run each step as a separate invocation with the same env assignments and working directory, recorded verbatim.
- RED-now: not applicable. This is an observation criterion; its failure direction is a missing or non-isolated observation.

### AC-GCB-011 — config key, historical records, citation inventory (REQ-GCB-010, REQ-GCB-011, REQ-GCB-012)

Given the change set, Then all of the following hold:
- `git diff --name-only f67d2193f...HEAD -- CHANGELOG.md reports .moai/specs .moai/reports` lists only paths under `.moai/specs/SPEC-GTD-CANON-BODY-001/` or `.moai/reports/t867/`.
- `git diff f67d2193f...HEAD -- internal/config .moai/config internal/template/templates/.moai/config` prints nothing.
- `.moai/reports/t867/verdict.md` contains `workflow.todo.enabled`, lists the spec.md §C.3 pre-existing failures as a lead finding, and holds a citation inventory whose entry count equals the line count of `grep -rn 'workflows/todo\.md' CHANGELOG.md reports .moai/specs .moai/reports --exclude-dir=SPEC-GTD-CANON-BODY-001 --exclude-dir=t867 --exclude='SPEC-GTD-CANON-BODY-001-*'` (the last exclusion drops this SPEC's own gitignored plan-audit reviews). The plan-time measurement was 13 + 132 = 145 before that exclusion; the run phase re-measures with it.

## Edge Cases

- A `moai todo` line documenting the alias without the `compat alias` marker fails AC-GCB-006; reword the line rather than widening the rule.
- The section anchor `workflows/todo.md § Standing sources` is retargeted to the gtd.md section of the same name, which AC-GCB-001 guarantees exists.
- The deploy-time `.agents/skills/moai` mirror stops carrying `workflows/todo.md`; the template package run in AC-GCB-010 covers the mirror code path.

## Definition of Done

- AC-GCB-001 through AC-GCB-011 PASS, with the command, verbatim output and exit code in progress.md §E.2
- Template-First parity; `make build` clean; regenerated `catalog.yaml` and `manager-lead.toml` committed
- `.moai/reports/t867/verdict.md` written with the config-key finding, the pre-existing-failure finding, and the historical citation inventory
- Card id `t867` in every commit message

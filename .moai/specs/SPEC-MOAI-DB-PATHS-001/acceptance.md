# SPEC-MOAI-DB-PATHS-001 — Acceptance (card t1827)

Acceptance entries are Given-When-Then, binary-testable, and follow the two-cell adoption discipline of `.claude/rules/moai/development/verification-completeness.md` § 2: every criterion carries a RED-now cell (observed on the pre-implementation tree, pinned to a SHA, with the reason it is red) and a green-path cell (which milestone flips it and what the passing output becomes).

## A. Baseline and conventions

- **Tree SHA (document-level pin, binds every criterion that carries none of its own):** `93f4fc75c8dafc1c436458940efa8957fd25ae12`.
- **TREEBIN** below means a binary built from the tree under measurement: `go build -o <scratch>/moai-tree ./cmd/moai`. The RED-now measurements used a build of the pinned SHA (build exit 0). The installed `moai` (build `ee981a932`) was used only to read the card text and never as measurement evidence (`verification-claim-integrity.md` § 2.2).
- **Swept-set rule for Go tests.** A `go test -run` selector that matches no test exits 0 and prints `[no tests to run]`. Every Go-test criterion therefore passes only when the `-v` output lists every named test with `--- PASS` and the stated count; exit 0 alone proves nothing. Observed at the pinned SHA, see ledger L-01.
- **Test isolation.** Tests set an absolute `MOAI_HOME` under `t.TempDir()` or build a temp-origin project; none reads the operator's real home, and `internal/testguard` stays green.
- **Operator-machine criteria** (marked OM) depend on this machine's canonical root and are evidence, not CI criteria.
- **Change-set judge.** "What did this card change" is read with the three-dot form `main...HEAD`, never a literal base SHA, so the judge survives a re-absorb of local `main`. The pinned SHA above is the RED-now measurement coordinate only.

### Evidence ledger (RED-now observations, pinned to the SHA above)

```text
L-01  cmd:  go test ./internal/cli -run '^(TestPaths_JSONShape|TestPaths_JSONSingleObject|TestPaths_ExistenceSizeWalShm|TestPaths_HumanListing|TestPaths_ParityWithVerbs|TestPaths_LegacyQueuePreAdoption|TestPaths_NeverCreatesFiles|TestPaths_SourceHasNoOpeningCalls|TestPaths_FieldValues|TestPaths_FieldUnknown|TestPaths_WorktreeConverges|TestPaths_LayoutProjectLocal|TestPaths_LayoutUnresolved|TestPaths_BrokerRun|TestPaths_BrokerInvalidRun|TestPaths_BrokerTemplateCount|TestPaths_HelpPointers)$' -v -count=1
      exit: 0
      out:  testing: warning: no tests to run
            PASS
            ok  	github.com/modu-ai/moai-adk/internal/cli	1.371s [no tests to run]
      why red: the tests do not exist; the run is vacuously green, so the swept count (0) is the verdict.
      measured: this revision, tree HEAD 4a8e8cfd5 (content identical to the pin outside the SPEC directory), go test of the tree.

L-02  cmd:  <TREEBIN> paths --json
      exit: 1
      out:  ERROR  Unknown command "paths" for "moai".
      why red: the command is not registered.

L-03  cmd:  <TREEBIN> doctor --check "DB Path Ghosts"
      exit: 0
      out:  Pass 0    Warn 0    Fail 0   (no check row printed)
      why red: an unknown check name selects nothing and still exits 0 — the empty sweep; green requires a named row.

L-04  cmd:  grep -rlF "never guess them" .claude/agents/moai/manager-todo.md .claude/skills/moai-factory-foreman/SKILL.md .claude/skills/moai/SKILL.md .claude/skills/moai/workflows/gtd.md .moai/docs/todo-queue-storage.md .moai/workflows/gtd.md
      exit: 1
      out:  (empty)
      why red: no surface carries the pointer sentence.

L-05  cmd:  grep -lF "moai paths" README.md README.ko.md README.ja.md README.zh.md docs-site/content/en/advanced/factory-mode.md docs-site/content/ko/advanced/factory-mode.md docs-site/content/ja/advanced/factory-mode.md docs-site/content/zh/advanced/factory-mode.md docs-site/content/en/utility-commands/moai-todo.md docs-site/content/ko/utility-commands/moai-todo.md docs-site/content/ja/utility-commands/moai-todo.md docs-site/content/zh/utility-commands/moai-todo.md
      exit: 1
      out:  (empty)
      why red: none of the twelve user-facing files mentions the command (anchor 1 of the five of AC-DBP-012; the page test goes red on this anchor, and on the other four wherever absent).

L-06  cmd:  git diff --numstat 93f4fc75c8dafc1c436458940efa8957fd25ae12 -- .claude/skills/moai/SKILL.md internal/template/templates/.claude/skills/moai/SKILL.md .moai/docs/todo-queue-storage.md internal/template/templates/.moai/docs/todo-queue-storage.md
      exit: 0
      out:  (empty)
      why red: nothing has been added yet; green is four lines of the form "<added>\t0\t<path>" with added > 0.

L-07  cmd:  grep -c sha256sum .claude/skills/moai-factory-foreman/SKILL.md
      exit: 0
      out:  1
      why red: the hand-computed key is still present (and `shasum` also counts 1 at L118).

L-08  cmd:  <TREEBIN> todo --help   (counted with: grep -c 'moai paths')
      out:  0   — likewise for `<TREEBIN> factory --help` (0) and `<TREEBIN> handoff save --help` (0)
      why red: no help text points at the command.

L-09  cmd:  sqlite3 ghostprobe.db '.tables'   then   sqlite3 ghostprobe2.db 'select 1;'   then   sqlite3 -readonly ghostprobe3.db '.tables'
      out:  first two exit 0 and leave a 0-byte file; the third exits 1 ("unable to open database file") and creates nothing.
      meaning: the premise of REQ-DBP-010 (3) is observed, not assumed. Not a criterion; recorded so the instruction text cites a measurement.

L-10  cmd:  git grep -lE 'backlog\.db|factory\.db|broker\.db|todo\.db|project-key' 93f4fc75c -- . ':(exclude).moai/specs' ':(exclude).moai/reports' ':(exclude)reports' ':(exclude)*.go'
      exit: 0
      out:  36 lines (git prints the revision exactly as typed, so each line starts with "93f4fc75c:"):
            93f4fc75c:.claude/agents/moai/manager-todo.md
            93f4fc75c:.claude/skills/moai-factory-foreman/SKILL.md
            93f4fc75c:.claude/skills/moai/SKILL.md
            93f4fc75c:.claude/skills/moai/workflows/gtd.md
            93f4fc75c:.moai/docs/todo-queue-storage.md
            93f4fc75c:.moai/learning/2026-09-03-moai-todo.html
            93f4fc75c:.moai/learning/2026-09-03-moai-todo.md
            93f4fc75c:.moai/project/codemaps/data-flow.md
            93f4fc75c:.moai/project/codemaps/modules.md
            93f4fc75c:.moai/project/codemaps/overview.md
            93f4fc75c:.moai/workflows/gtd.md
            93f4fc75c:CHANGELOG.md
            93f4fc75c:README.ja.md
            93f4fc75c:README.ko.md
            93f4fc75c:README.md
            93f4fc75c:README.zh.md
            93f4fc75c:docs-site/content/en/advanced/factory-mode.md
            93f4fc75c:docs-site/content/en/advanced/moai-web-console.md
            93f4fc75c:docs-site/content/en/utility-commands/moai-todo.md
            93f4fc75c:docs-site/content/ja/advanced/factory-mode.md
            93f4fc75c:docs-site/content/ja/advanced/moai-web-console.md
            93f4fc75c:docs-site/content/ja/utility-commands/moai-todo.md
            93f4fc75c:docs-site/content/ko/advanced/factory-mode.md
            93f4fc75c:docs-site/content/ko/advanced/moai-web-console.md
            93f4fc75c:docs-site/content/ko/utility-commands/moai-todo.md
            93f4fc75c:docs-site/content/zh/advanced/factory-mode.md
            93f4fc75c:docs-site/content/zh/advanced/moai-web-console.md
            93f4fc75c:docs-site/content/zh/utility-commands/moai-todo.md
            93f4fc75c:internal/template/templates/.claude/agents/moai/manager-todo.md
            93f4fc75c:internal/template/templates/.claude/skills/moai-factory-foreman/SKILL.md
            93f4fc75c:internal/template/templates/.claude/skills/moai/SKILL.md
            93f4fc75c:internal/template/templates/.claude/skills/moai/workflows/gtd.md
            93f4fc75c:internal/template/templates/.codex/agents/moai/manager-todo.toml
            93f4fc75c:internal/template/templates/.moai/docs/todo-queue-storage.md
            93f4fc75c:internal/testguard/home_access_seed.txt
            93f4fc75c:internal/web/screens.templ
      meaning: the count the SPEC states (36). Spec § C classifies exactly these 36 plus the four of L-11. Not a pass/fail criterion; it fixes the inventory the surface criteria (AC-DBP-009, -012) are judged against.

L-11  cmd:  git grep -lE 'backlog\.db|factory\.db|broker\.db|todo\.db|project-key' 93f4fc75c -- reports
      exit: 0
      out:  93f4fc75c:reports/hooks-audit-20260911-01a08e35/cards-after.json
            93f4fc75c:reports/hooks-audit-20260911-01a08e35/cards-before.json
            93f4fc75c:reports/moai-factory-mcp-matrix-live-20260924.html
            93f4fc75c:reports/moai-factory-mcp-matrix-live-20260924.md
      meaning: the grep of L-10 WITHOUT the `:(exclude)reports` pathspec returns these four extra files, 40 in total; the first audit measured 40 for that reason. They are classified in spec § C as historical report artifacts and never edited.

L-12  cmd:  git diff --name-only main...HEAD
      exit: 0
      out:  .moai/specs/SPEC-MOAI-DB-PATHS-001/acceptance.md
            .moai/specs/SPEC-MOAI-DB-PATHS-001/plan.md
            .moai/specs/SPEC-MOAI-DB-PATHS-001/progress.md
            .moai/specs/SPEC-MOAI-DB-PATHS-001/spec.md
      meaning: the change-set judge of AC-DBP-016 on the pre-implementation tree (local main = 93f4fc75c at measurement): only the SPEC directory differs. Not a pass/fail criterion; it is the positive control proving the three-dot range is non-empty and the baseline against which the closure list is read.
```

## B. Acceptance criteria

### B.1 The command

- **AC-DBP-001** (maps REQ-DBP-001, REQ-DBP-004, REQ-DBP-005). Given a scratch git project and an absolute `MOAI_HOME`, When `<TREEBIN> paths --json` runs in it, Then it exits 0 and stdout is exactly one JSON object holding `project_key`, `project_dir`, `queue_layout`, and `todo`, `factory`, `broker` entries; each database entry carries `path`, `exists`, `size_bytes`, and `wal` and `shm` objects with `exists` and `size_bytes`, and `todo` additionally carries `dir`, `queue_state`, and `legacy_dir`; and When the same command runs without flags, Then it exits 0 and stdout is a listing of one `name: value` line per field whose first five names are `project-key`, `project-dir`, `todo-dir`, `todo-db`, `factory-db` in that order and which is not valid JSON. Judge: `go test ./internal/cli -run '^(TestPaths_JSONShape|TestPaths_JSONSingleObject|TestPaths_ExistenceSizeWalShm|TestPaths_HumanListing)$' -v -count=1` lists `--- PASS` for each of the four names, and the swept count is 4. RED-now: L-01, L-02. Green path: M1.
- **AC-DBP-002** (maps REQ-DBP-002, REQ-DBP-007). Given (a) a temp git project with an absolute `MOAI_HOME`, (b) a temp-origin project without `MOAI_HOME`, and (c) a project whose queue exists ONLY in a legacy directory (seeded through the store at the project-local `todo` directory, with no canonical directory yet), When the real `todo add` verb and the factory-opening path (`OpenFactory` at `factoryCardRoot()`) create their databases in (a) and (b), with CLAUDE_PROJECT_DIR naming the same project as the working directory, and `paths --json` runs, Then each reported `path` is a file that now exists with `size_bytes > 0` and equals the path the verb used. The factory leg is verified against `OpenFactory(factoryCardRoot())` — the root rule every factory-opening surface shares (card t1862 aligned the real `moai factory runs` and `recover-resume` verbs to it, with boundary tests on the t1862 card branch). When `paths --json` runs in (c) before any adopting verb, Then `todo.path` is the canonical path, `todo.exists` is `false`, `todo.queue_state` is `pre-adoption`, `todo.legacy_dir` is the seeded legacy directory, and a before/after walk shows nothing moved or created; When the real adopting `todo add` then runs in (c) and `paths --json` runs again, Then `todo.path` is the SAME string as before the adoption, `todo.exists` is `true`, `todo.queue_state` is `canonical`, `todo.legacy_dir` is `null`, and `todo.dir` equals the directory part of `todo.path`. The resolver-failure case (moai home unresolvable) is AC-DBP-006(b). Judge: `-run '^(TestPaths_ParityWithVerbs|TestPaths_LegacyQueuePreAdoption)$' -v` lists `TestPaths_ParityWithVerbs` with subtests `home` and `temp-origin`, and `TestPaths_LegacyQueuePreAdoption` with subtests `pre-adoption` and `post-adoption`, all `--- PASS` (two top-level tests, four subtests). Mutant probes: a command that recomputes the key with a different rule or joins `todo/backlog.db` itself passes (a) and fails (b); a command that reports the pure `BacklogPathForRoot` passes (a) and (b) and fails `pre-adoption` (it reports the legacy directory, and `todo.path` changes across the adoption); a test that only checked the key string would pass all of them — which is why parity is judged against files the verbs created and against the pre/post-adoption pair. RED-now: L-01. Green path: M1.
- **AC-DBP-003** (maps REQ-DBP-003). Given a fixture tree, When every flag combination of `paths` runs, Then a before/after walk of the moai home and the project root (names, sizes, mtimes) is identical. Judge: `-run '^(TestPaths_NeverCreatesFiles|TestPaths_SourceHasNoOpeningCalls)$' -v` lists `TestPaths_NeverCreatesFiles` and `TestPaths_SourceHasNoOpeningCalls` as `--- PASS`. Positive control (required, same test): the same walker run around `homestate.EnsureProjectLayout` reports a difference, proving the detector can go red; the source test also asserts its forbidden-identifier list finds `OpenFactory` in a known-bad string, and its list contains `BacklogPathForRoot` and `BacklogPathForRootAdopting`. RED-now: L-01. Green path: M1.
- **AC-DBP-004** (maps REQ-DBP-005). Given each `--field` name of the closed set `project-key`, `project-dir`, `todo-dir`, `todo-db`, `factory-db`, When run, Then stdout is the bare value plus one newline and equals the matching `--json` key of the plan § A.1 table (`project_key`, `project_dir`, `todo.dir`, `todo.path`, `factory.path` respectively), and the `todo-dir` value equals the directory part of the `todo-db` value; When an unknown name (for example `todo-path`, `broker-db`, or an empty name) is given, Then it exits non-zero with empty stdout; When the named field is unresolved (the AC-DBP-006(b) environment, `factory-db`), Then it exits non-zero with empty stdout. Judge: `-run '^(TestPaths_FieldValues|TestPaths_FieldUnknown)$' -v` lists `--- PASS` for `TestPaths_FieldValues` (five subtests, one per name) and `TestPaths_FieldUnknown` (three subtests: unknown, empty, unresolved). RED-now: L-01, L-02. Green path: M1.
- **AC-DBP-005** (maps REQ-DBP-006). Given a repository with a linked worktree, When `paths --json` runs from the worktree and from the primary checkout, Then `project_key` and every `path` are byte-identical. Judge: `-run '^TestPaths_WorktreeConverges$' -v` lists one `--- PASS`. OM evidence: `<TREEBIN> paths --field project-key` in this worktree prints `moai-adk-go-1bd3d038`. RED-now: L-01, L-02. Green path: M1.
- **AC-DBP-006** (maps REQ-DBP-007). Given (a) a temp-origin project without `MOAI_HOME` and (b) an environment where the moai home cannot be resolved, When `paths --json` runs, Then (a) reports `queue_layout` as `project-local` with the paths the verbs use, and (b) reports the factory fields as `null`, reports the queue at the project-local location the todo resolver returns with `queue_layout` `project-local`, exits non-zero, and its output contains none of the fallback path `.moai/state/factory/factory.db`. Judge: `-run '^(TestPaths_LayoutProjectLocal|TestPaths_LayoutUnresolved)$' -v` lists `TestPaths_LayoutProjectLocal` and `TestPaths_LayoutUnresolved` as `--- PASS`. Mutant probe: substituting `FactoryRegistryPath` makes (b) print the fallback and fail. RED-now: L-01. Green path: M1.
- **AC-DBP-007** (maps REQ-DBP-008). Given a factory directory holding `messages/r1/`, `messages/r2/`, `messages/r3/`, When `paths --run r1 --json` runs, Then the broker path ends with `/messages/r1/broker.db`; When `--run ../x` runs, Then it exits non-zero and prints no path; When no `--run` is given, Then the broker entry has a path template and `existing_runs` equals 3. Judge: `-run '^(TestPaths_BrokerRun|TestPaths_BrokerInvalidRun|TestPaths_BrokerTemplateCount)$' -v` lists three `--- PASS` (`TestPaths_BrokerRun`, `TestPaths_BrokerInvalidRun`, `TestPaths_BrokerTemplateCount`). RED-now: L-01. Green path: M1.
- **AC-DBP-008** (maps REQ-DBP-009). Given the built binary, When `todo --help`, `factory --help`, and `handoff save --help` are read, Then each contains `moai paths`, and the `todo` long help still contains the string `~/.moai/db/<project-key>/todo/backlog.db`. Judge: `go test ./internal/cli -run '^(TestPaths_HelpPointers|TestTodoHelp_NamesTheDatabaseStore|TestTodoStoreClaims_NoStaleGoSourceSite)$' -v -count=1` lists three `--- PASS` (the second and third already exist in `internal/cli/todo_help_store_test.go` at line 35 and line 50 of the pinned tree). RED-now: L-08. Green path: M2.

### B.2 The instruction surfaces

- **AC-DBP-009** (maps REQ-DBP-010). Given the twelve-file surface set S12 of spec § C (the five local twin files and their five template twins, the template-only `manager-todo.toml` after regeneration, and the local-only `.moai/workflows/gtd.md`), When the surface test reads both trees, Then each file contains each of the five fixed lead phrases of plan § A.3, exactly once, as a single-line substring: `Find the database files with `, `Forbidden guessed paths:`, `Never `sqlite3`-open a guessed path`, `Direct reads only on a snapshot copy`, `Prefer the CLI:`. Judge: `-run '^TestDBPathsBlock_Surfaces$' -v` lists one `--- PASS` whose log states `surfaces=12 phrases=5`. Regeneration judge for the generated TOML: `AGENTEMIT_UPDATE=` unset, `go test ./internal/template/agentemit/... -run '^TestGoldenCommittedArtifactsMatchEmission$' -v -count=1` lists `--- PASS` (swept count 1) — equivalent to `make agents-emit-check` exiting 0; its red input to observe in the run phase is the neutral `manager-todo.md` edited without running `make agents-emit` (the test fails naming drift; then regenerate). Mutant probe: the same surface test, given an in-memory copy of one file with the forbidden-paths line deleted, must report that file; the test carries that subtest. RED-now: L-04 (the pointer phrase is absent from the six local files). Green path: M3.
- **AC-DBP-010** (maps REQ-DBP-011). Given the four files of the two twin pairs that already differ (`.claude/skills/moai/SKILL.md` and `internal/template/templates/.claude/skills/moai/SKILL.md`; `.moai/docs/todo-queue-storage.md` and `internal/template/templates/.moai/docs/todo-queue-storage.md`), When the run-phase diff against the pinned SHA is taken, Then each of the four files shows added lines and zero removed lines. Judge: `git diff --numstat 93f4fc75c8dafc1c436458940efa8957fd25ae12 -- <the four paths>` prints four lines whose second column is `0` and whose first column is greater than `0`. RED-now: L-06 (empty). Green path: M3. The three byte-identical twin pairs (`manager-todo.md`, `workflows/gtd.md`, the foreman skill) are not covered by this criterion; of them only the foreman skill removes lines (REQ-DBP-012), and `manager-todo.md` and `workflows/gtd.md` are additive by the same insertion rule.
- **AC-DBP-011** (maps REQ-DBP-012). Given the foreman skill in both trees, When read, Then neither contains `sha256sum` or `shasum`, both contain `moai paths --field todo-dir`, and the watch block contains an error branch that exits instead of computing a fallback. Judge: `grep -c sha256sum <path>` prints `0` for both foreman files, `grep -c "paths --field todo-dir" <path>` prints `1` for both, `-run '^TestForemanWatch_FailsVisibly$' -v` lists one `--- PASS` (it executes the extracted watch block against a stub `moai` that fails and asserts a non-zero exit with no `backlog changed` output), and `go test ./internal/factory -run '^(TestForemanQueueWatchResolvesCanonicalStateDir|TestForemanQueueWatch_FiresOnMutation|TestForemanQueueWatch_ShippedJSONTargetIsSilent|TestForemanQueueWatch_FiresWithStaleJSONPresent|TestForemanQueueWatch_WatchTargetsAgree)$' -v -count=1` lists five top-level `--- PASS` (the statement test plus the four watch tests; only `TestForemanQueueWatch_FiresOnMutation` has subtests, `local` and `template`, so seven `--- PASS` lines in all). The existing tests are counted swept only because plan M3 step 7 gives them a `moai` stub on the fixture PATH: the stub is a single trivial `exec` ahead of the loop, so the time-bound watch family keeps its 2-second settle and `watchWindow` deadline and stays deterministic; without the stub they go red, which is the RED input to observe in the run phase (rewrite the script, run them, watch `$d` stay unset, then add the stub). RED-now: L-07. Green path: M3.
- **AC-DBP-012** (maps REQ-DBP-013). Given the four READMEs and the eight docs-site pages of spec § C, When the page test reads them, Then every one of the twelve contains all five language-neutral anchors of plan § A.3 — `moai paths`, `~/.moai/db/todo.db`, `sqlite3`, `.db-wal`, `moai todo list --json` — and all five lie inside the same heading-delimited section as the file's pre-existing database mention (the section containing `project-key` or `backlog.db`). Judge: `go test ./internal/template -run '^TestDBPathsBlock_UserFacingPages$' -v -count=1` lists one `--- PASS` whose log states `pages=12 anchors=5`; and the `hns-oss-docs-verify` recipe passes (warning-free hugo build, four-locale and README heading parity). Mutant probe: the test carries a subtest in which an in-memory copy of one page keeps `moai paths` and drops `.db-wal`, and must report that page. RED-now: L-05 (anchor 1 is absent from all twelve, exit 1). Green path: M4. Home-access guard: the page test and every other new test file that writes a `~/` string literal marks it with a `// home-guard:allow <reason>` comment (non-empty reason, same line or the line above); the seed file is not edited. Residual: the anchors prove the statements are present, not that the locale prose states them well; that is review's judgement.
- **AC-DBP-013** (maps REQ-DBP-014). Given the embedded template tree, When the drift-guard test runs, Then it passes on the tree; and given an in-memory copy of one inventory surface with the pointer removed, it fails naming that file; and given a new in-memory template file mentioning `<project-key>/todo/backlog.db` without the pointer, it fails naming that file; and every allowlist entry carries a non-empty reason. Judge: `-run '^TestTemplateDBPathSurfacesCarryPointer$' -v` lists the tree subtest and both mutant subtests as `--- PASS`, swept count 3. RED-now: the guard does not exist (L-01 shape). Green path: M3. Additional constraint checked here: `TestDeprecatedPaths_NoTemplateTextReference` still passes (new template text must not contain a bare `.moai/db`; the home-scoped spelling `~/.moai/db` is allowed).

### B.3 The doctor check

- **AC-DBP-014** (maps REQ-DBP-015). Given fixtures with a zero-byte `*.db` at each of the eight observed paths, When the check runs, Then it returns warning severity naming all eight; given a zero-byte file at a canonical path, it is not reported; given a non-empty file at a guessed path, it is not reported; given an unreadable directory, the result is still not a failure; and a before/after walk shows nothing was created, moved, or deleted. Judge: `-run '^(TestDBPathGhosts_EightObserved|TestDBPathGhosts_CanonicalExcluded|TestDBPathGhosts_NonEmptyIgnored|TestDBPathGhosts_UnreadableNotFail|TestDBPathGhosts_ReadOnly)$' -v` lists `TestDBPathGhosts_EightObserved`, `TestDBPathGhosts_CanonicalExcluded`, `TestDBPathGhosts_NonEmptyIgnored`, `TestDBPathGhosts_UnreadableNotFail`, `TestDBPathGhosts_ReadOnly` as `--- PASS` (swept count 5), and `<TREEBIN> doctor --check "DB Path Ghosts"` prints a row named `DB Path Ghosts` (RED-now: L-03, no row). Green path: M5.
- **AC-DBP-015** (maps REQ-DBP-016). Given the check registered in `doctor.go`, When the binary-lag tests run, Then the registration appears exactly once in `doctor.go` and exactly once in the contract map of `internal/cli/binary_lag_test.go`, and the three doctor golden files are regenerated in the same commit. Judge: `-run '^(TestBinaryLag_DoctorCheckNameSetIsUnchanged|TestBinaryLag_DoctorCheckNameGuardRejectsMutations|TestBinaryLag_OneSeamServesBothSurfaces|TestBinaryLag_NonGitDirectoryKeepsDoctorExitZero)$' -v` lists four `--- PASS`; `grep -c dbPathGhostCheckName internal/cli/binary_lag_test.go` prints `1` (RED-now: `0`, exit 1); `grep -c "DB Path Ghosts" internal/cli/testdata/doctor-dark.golden` prints `1` and likewise for `-light` and `-nocolor`. RED input to observe in the run phase: register the name in `doctor.go` without the contract entry and watch `TestBinaryLag_DoctorCheckNameSetIsUnchanged` fail with `unknown doctor registration`. Green path: M5.

### B.4 Closure

- **AC-DBP-016** (maps all REQ). Given the finished change, When the gates run, Then: (1) `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0, and `go vet ./internal/cli/... ./internal/homestate/... ./internal/factory/... ./internal/template/...` exits 0; (2) the change-scoped `go test` runs of the milestones M1-M5 pass with the stated swept counts; (3) `make agents-emit-check`, `make commands-emit-check`, and `make embed-manifest-check` exit 0; (4) `moai spec lint` reports no errors on this SPEC; (5) the change-set equals the allow-list below: every path printed by `git diff --name-status main...HEAD` is on the list, every added path is marked A on it, and no path of the not-in-the-change-set list appears. Judge: the run-phase verdict file cites each command with verbatim output; the change-set judge is `git diff --name-status main...HEAD`, with the positive control `git diff --name-only main...HEAD -- .moai/specs/SPEC-MOAI-DB-PATHS-001` printing at least one path (L-12 shows the pre-implementation baseline: the SPEC directory alone).
  - **Allow-list, modified (M):** `internal/cli/root.go`; `internal/cli/todo.go`; `internal/cli/handoff.go`; `internal/cli/factory_handoff_recover.go`; the ten S12 twin-pair files (`.claude/agents/moai/manager-todo.md`, `.claude/skills/moai-factory-foreman/SKILL.md`, `.claude/skills/moai/SKILL.md`, `.claude/skills/moai/workflows/gtd.md`, `.moai/docs/todo-queue-storage.md`, and their `internal/template/templates/` twins); `.moai/workflows/gtd.md`; `internal/template/templates/.codex/agents/moai/manager-todo.toml` (regenerated by `make agents-emit`); `internal/template/catalog.yaml` (hashes rewritten by `gen-catalog-hashes.go --all`); the four READMEs; `docs-site/content/{en,ko,ja,zh}/advanced/factory-mode.md`; `docs-site/content/{en,ko,ja,zh}/utility-commands/moai-todo.md`; `internal/factory/foreman_queue_statement_test.go` and `internal/factory/foreman_queue_watch_test.go` (stub `moai` on the fixture PATH, plan M3 step 7).
  - **Allow-list, added (A):** the SPEC directory artifacts `.moai/specs/SPEC-MOAI-DB-PATHS-001/{spec,plan,acceptance,progress}.md` (the SPEC is new on the branch, so `git diff --name-status main...HEAD` prints `A`, observed in L-12's range; a SPEC already present on `main` would print `M`); `internal/cli/paths.go`; `internal/cli/paths_test.go`; `internal/factory/queue_location.go`; `internal/factory/queue_location_test.go`; `internal/template/db_paths_surfaces_test.go` (hosts `TestDBPathsBlock_Surfaces`, `TestDBPathsBlock_UserFacingPages`, `TestTemplateDBPathSurfacesCarryPointer`, `TestForemanWatch_FailsVisibly`).
  - **Allow-list, M5 only (present only when M5 is kept):** M `internal/cli/doctor.go`; M `internal/cli/binary_lag_test.go`; M `internal/cli/testdata/doctor-dark.golden`, `doctor-light.golden`, `doctor-nocolor.golden`; A `internal/cli/doctor_db_ghosts.go`; A `internal/cli/doctor_db_ghosts_test.go`.
  - **Not in the change-set (any appearance fails):** `internal/factory/backlog_sqlite.go`, `internal/factory/state_dir.go`, any other existing `internal/factory` file (the two foreman test files of the modified list excepted) or any `internal/homestate` file, `internal/template/embed_manifest_gen.go`, `.moai/project/codemaps/*`, `CHANGELOG.md`, and any `.agents/skills/` file.
  - **M5-dropped variant:** when M5 is dropped (plan § B), the M5-only paths must be absent from the diff; "maps all REQ" then reads "maps all REQ except REQ-DBP-015 and REQ-DBP-016", the milestone list in (2) reads M1-M4, and AC-DBP-014 and AC-DBP-015 move to the follow-up card (recorded in progress.md §E.3).
  - RED-now: not applicable (closure gate; the individual criteria carry the RED cells, and L-12 carries the change-set baseline). Green path: M5 close, or M4 close when M5 is dropped.

## C. Edge cases

- A project with no git metadata: `paths` reports the launch directory's key and the same paths the todo verbs and the `factoryCardRoot()` path use; no error.
- A database that exists as a zero-byte file at a canonical path: reported `exists: true`, `size_bytes: 0`, never as a ghost.
- A `-wal` present with no database: reported as such; the command does not infer a database.
- `--run` given a valid id whose directory does not exist: path reported, `exists: false`, exit 0.
- Concurrent writer holds the queue lock: `paths` neither waits nor fails; it takes no lock.
- Windows path separators: reported paths use the platform separator; the JSON carries them verbatim.
- A queue that exists only in a legacy directory: the canonical path is reported with `exists: false`, `queue_state` `pre-adoption`, and `legacy_dir` naming where the data currently is (AC-DBP-002 case (c)); the command neither opens nor moves it.
- Adoption refused (a filesystem that cannot rename across the boundary): the verbs keep serving the legacy directory, and `paths` keeps reporting `pre-adoption` with `legacy_dir` until the relocation succeeds; the canonical path is never reported as holding data it does not hold.
- A git-directory project root inside a temporary root: the queue and factory resolvers apply different temporary-origin predicates and can disagree; `queue_layout` describes the queue only and the factory path is reported as its resolver returns it.

## D. Quality gates and Definition of Done

- Every criterion above has an observed RED-now cell or an explicit statement of why none applies (AC-DBP-016).
- Test coverage of the new command file is at least 85% by `go test -cover` on its package, measured in the run phase.
- The doctor milestone (M5) may be dropped as a unit; if it is dropped, REQ-DBP-015, REQ-DBP-016, AC-DBP-014, and AC-DBP-015 move to a follow-up card and the rest of the SPEC closes without them under the AC-DBP-016 M5-dropped variant (recorded in progress.md §E.3).
- The sync phase regenerates codemaps and adds the CHANGELOG entry; neither is a run-phase deliverable. The sync-phase checkpoint for the codemaps (spec § F.5, not a run-phase criterion): the sync verdict records, for each of `.moai/project/codemaps/{data-flow,modules,overview}.md`, whether it still names a database path and whether it then points at `moai paths`; a file that names a path without the pointer is reported as follow-up card input and is not hand-edited.
- No criterion is recorded as a pass on an exit code alone where a swept count or a named row is required.

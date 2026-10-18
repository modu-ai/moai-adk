# progress.md — SPEC-PRECOMMIT-AFFECTED-SCOPE-001

## §E.1 Plan-Phase Audit-Ready Signal

plan_complete_at: 2026-10-18
plan_status: audit-ready
audit_ready: true
tier: M
artifacts: spec.md, plan.md, acceptance.md, decision-index.md, progress.md
requirements: 14 live + REQ-PAS-012 withdrawn (ceiling 16)
acceptance_criteria: 15 live + AC-PAS-012 withdrawn (ceiling 16)
withdrawn: REQ-PAS-012, AC-PAS-012, plan M4 (Q1 ruled B, board d-20261018T005931Z-f264)
planned_against_tree: 3642a3a9b9cf101863656153d2d99b69fcb66a05
open_decisions_blocking_kickoff: none (Q1 and Q5 RULED by the leader; Q4 is a run-phase measurement)
applied_defaults: Q2, Q3 (DEFAULT-APPLIED at plan close)
artifacts_added: evidence/probe016.sh, evidence/probe009.sh (RED-now probes cited by acceptance.md ledger)
plan_audit_iter1: FAIL 0.75 at 2febc5c2f (.moai/reports/t1673/plan-audit-iter1.md); one correction pass applied (D1-D7 and D9 blocking; D8, D10-D14 folded); counts unchanged: 14 live REQ, 15 live AC, withdrawn slots REQ-PAS-012 / AC-PAS-012 / M4
plan_audit_iter2: FAIL 0.875 (required codex backend FAIL) at 99b1a5aa6 (.moai/reports/t1673/plan-audit-iter2.md); narrow correction pass N1-N4 + optional S1 applied; iter3 is the ceiling round
class_counts: release-blocking 8 (AC-001, -004, -005, -006, -009, -011, -013, -016), adoption-pending 2 (AC-002, -003), regression-guard 5 (AC-007, -008, -010, -014, -015)
known_gaps: G1 (commit -a, path-limited and rename divergences OBSERVED in LEDGER-PAS-009-A; --amend and linked-worktree forms still unmeasured), G2 (go list wall time unmeasured), G3 (project-wide vet/lint/typecheck runtime unmeasured), G4 (full-tree run of the real suppression checker not performed)

## §E.2 Run-phase Evidence

Run-phase conventions. Starting tree: `00e7843abe0bb14e78726e8f271279d8e97c25e5` (branch `WT-pre-commit-gate`, planned against `3642a3a9b`). Every RED block below was captured before the GREEN commit of its milestone; each carries the verbatim failing output of the command shown and its exit code. TMPDIR=GOTMPDIR=/private/tmp/t1673-tmp for every `go test`. Pre-flight baseline (this run, HEAD `00e7843ab`): `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/hook/quality/ ./internal/astgrep/ ./internal/spec/ ./internal/config/` exit 0; `golangci-lint run --timeout=2m ./internal/hook/quality/...` exit 0, `0 issues.`

### RED-M1 — AC-PAS-011 (scope-mode config tests; tree 00e7843ab + M1 RED test file and the `TestScope` struct stub)
Command: `go test ./internal/config -run '^TestGatePreCommitTestScope$' -count=1 -v`
Exit code: 1
```
=== RUN   TestGatePreCommitTestScope
=== RUN   TestGatePreCommitTestScope/absent_key_defaults_to_affected
    gate_test_scope_test.go:69: default test_scope.mode = "", want "affected"
=== RUN   TestGatePreCommitTestScope/pre_commit_without_test_scope_defaults_to_affected
    gate_test_scope_test.go:79: test_scope.mode = "", want "affected"
=== RUN   TestGatePreCommitTestScope/accepts_affected
=== RUN   TestGatePreCommitTestScope/accepts_full
=== RUN   TestGatePreCommitTestScope/rejects_any_other_value_naming_the_key_and_the_accepted_values
    gate_test_scope_test.go:98: Validate accepted test_scope.mode: bogus, want a rejection
=== RUN   TestGatePreCommitTestScope/shipped_gate.yaml_copies_agree_on_the_scope_block
    gate_test_scope_test.go:112: local gate.yaml carries no test_scope block with mode: affected:
--- FAIL: TestGatePreCommitTestScope (0.01s)
    --- FAIL: TestGatePreCommitTestScope/absent_key_defaults_to_affected (0.00s)
    --- FAIL: TestGatePreCommitTestScope/pre_commit_without_test_scope_defaults_to_affected (0.00s)
    --- PASS: TestGatePreCommitTestScope/accepts_affected (0.00s)
    --- PASS: TestGatePreCommitTestScope/accepts_full (0.00s)
    --- FAIL: TestGatePreCommitTestScope/rejects_any_other_value_naming_the_key_and_the_accepted_values (0.00s)
    --- FAIL: TestGatePreCommitTestScope/shipped_gate.yaml_copies_agree_on_the_scope_block (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/config	0.356s
FAIL
```
Why red: default, validation and both YAML copies are not yet implemented (the two `accepts_*` subtests pass because the struct field already decodes; they are not the red cells).

### GREEN-M1 — AC-PAS-011 config half (same command, flipped; tree = M1 GREEN commit)
Command: `go test ./internal/config -run '^TestGatePreCommitTestScope$' -count=1 -v`
Exit code: 0
```
=== RUN   TestGatePreCommitTestScope
=== RUN   TestGatePreCommitTestScope/absent_key_defaults_to_affected
=== RUN   TestGatePreCommitTestScope/pre_commit_without_test_scope_defaults_to_affected
=== RUN   TestGatePreCommitTestScope/accepts_affected
=== RUN   TestGatePreCommitTestScope/accepts_full
=== RUN   TestGatePreCommitTestScope/rejects_any_other_value_naming_the_key_and_the_accepted_values
=== RUN   TestGatePreCommitTestScope/shipped_gate.yaml_copies_agree_on_the_scope_block
--- PASS: TestGatePreCommitTestScope (0.01s)
    --- PASS: TestGatePreCommitTestScope/absent_key_defaults_to_affected (0.00s)
    --- PASS: TestGatePreCommitTestScope/pre_commit_without_test_scope_defaults_to_affected (0.00s)
    --- PASS: TestGatePreCommitTestScope/accepts_affected (0.00s)
    --- PASS: TestGatePreCommitTestScope/accepts_full (0.00s)
    --- PASS: TestGatePreCommitTestScope/rejects_any_other_value_naming_the_key_and_the_accepted_values (0.00s)
    --- PASS: TestGatePreCommitTestScope/shipped_gate.yaml_copies_agree_on_the_scope_block (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.327s
```
Other M1 verification (this run, this tree): `go test ./internal/config -count=1` fails only `TestNoHardcodedContexts` (BASELINE, not attributable: it reads `scripts/release.sh`, which contains the forbidden string `Test (ubuntu-latest)` at HEAD (`grep -c` = 2) and which this card does not touch; `git diff --stat HEAD -- scripts/release.sh` is empty). `TestShippedConfigKeysHaveReaders` turned red on the new key and was repaired by one inventory row in `internal/config/testdata/shipped_key_inventory.yaml` (class W). `go test ./internal/settings ./internal/web -count=1` exit 0. `go test ./internal/cli -run '^(TestUpdateGateYAMLPreservesUserValues|TestUpdateGateYAMLDeliversNewPreCommitKey|TestRunGate|TestGAR_|TestGateCmd_RunE_Behavior|TestGateCmd_PassPathEmitsExecutionSummary|TestPrecommitE2E)' -count=1 -v` exit 0 (the two update tests PASS). `go test ./internal/template -count=1` exit 0 (`ok 440.463s`; NOTE: run before the slot lease, see Gaps). `make build` exit 0 and left no tracked change.
Decision recorded: an empty `test_scope.mode` (the zero value of a hand-built Config) validates and resolves to `affected`; only a non-empty unknown value is rejected (needed so existing tests that call `Validate` on zero-valued configs keep passing).


### RED-M2 — AC-PAS-001, -002, -003 (P1 rows), -004, -005 core, -006 core, -007 core (affected-set core; tree 00e7843ab + M1 commits + M2 RED test file and the `selectAffected`/`widenTrigger` stubs)
The stubs return the zero value, so every assertion fails on the right reason: nothing is selected, no trigger fires, no skip reason exists.
Command: `go test ./internal/hook/quality -run '^TestAffectedScope_' -count=1 -v`
Exit code: 1
```
=== RUN   TestAffectedScope_DirectDependents
    affected_scope_test.go:73: selected = [], want exactly [example.com/m/a example.com/m/b]
    affected_scope_test.go:75: run patterns = [], want [./a ./b] (module-root-relative directory patterns)
--- FAIL: TestAffectedScope_DirectDependents (0.02s)
=== RUN   TestAffectedScope_TestImportEdges
    affected_scope_test.go:91: selected = [], want exactly [example.com/m/a example.com/m/b example.com/m/e example.com/m/f]
--- FAIL: TestAffectedScope_TestImportEdges (0.00s)
=== RUN   TestAffectedScope_Ownership
    affected_scope_test.go:112: selected = [], want exactly [example.com/m/p example.com/m/q example.com/m/r]
    affected_scope_test.go:114: unowned = [], want [docs/note.md]
    affected_scope_test.go:119: summary does not name docs/note.md as owned by no package:
--- FAIL: TestAffectedScope_Ownership (0.00s)
=== RUN   TestAffectedScope_TestOnlyNarrows
    affected_scope_test.go:130: selected = [], want exactly [example.com/m/a]
--- FAIL: TestAffectedScope_TestOnlyNarrows (0.00s)
=== RUN   TestAffectedScope_DeletedAndRenamed
=== RUN   TestAffectedScope_DeletedAndRenamed/deleted_last_file_of_a_package
    affected_scope_test.go:141: selected = [], want exactly [example.com/m/b]
=== RUN   TestAffectedScope_DeletedAndRenamed/rename_inside_a_package
    affected_scope_test.go:145: selected = [], want exactly [example.com/m/a example.com/m/b]
=== RUN   TestAffectedScope_DeletedAndRenamed/rename_across_packages_reports_both_directories
    affected_scope_test.go:151: selected = [], want exactly [example.com/m/a example.com/m/b example.com/m/e]
--- FAIL: TestAffectedScope_DeletedAndRenamed (0.00s)
    --- FAIL: TestAffectedScope_DeletedAndRenamed/deleted_last_file_of_a_package (0.00s)
    --- FAIL: TestAffectedScope_DeletedAndRenamed/rename_inside_a_package (0.00s)
    --- FAIL: TestAffectedScope_DeletedAndRenamed/rename_across_packages_reports_both_directories (0.00s)
=== RUN   TestAffectedScope_DropAndEmpty
=== RUN   TestAffectedScope_DropAndEmpty/a_selection_with_one_untested_package
    affected_scope_test.go:161: selected = [], want exactly [example.com/m/a example.com/m/n]
    affected_scope_test.go:163: dropped = [], want [example.com/m/n]
    affected_scope_test.go:166: run = [], want [./a]
    affected_scope_test.go:170: summary lacks "selected 2, dropped 1":
    affected_scope_test.go:173: summary does not name the dropped package with the reason "no test files":
=== RUN   TestAffectedScope_DropAndEmpty/a_changed_list_holding_only_a_docs_file
    affected_scope_test.go:182: an empty selection carries no skip reason; it must be an explicit skip, not a pass
=== RUN   TestAffectedScope_DropAndEmpty/every_selected_package_is_untested
    affected_scope_test.go:189: run = [], skip reason = "", want an explicit skip
=== RUN   TestAffectedScope_DropAndEmpty/no_changed_list_at_all
    affected_scope_test.go:195: run = [], skip reason = "", want an explicit skip
--- FAIL: TestAffectedScope_DropAndEmpty (0.00s)
    --- FAIL: TestAffectedScope_DropAndEmpty/a_selection_with_one_untested_package (0.00s)
    --- FAIL: TestAffectedScope_DropAndEmpty/a_changed_list_holding_only_a_docs_file (0.00s)
    --- FAIL: TestAffectedScope_DropAndEmpty/every_selected_package_is_untested (0.00s)
    --- FAIL: TestAffectedScope_DropAndEmpty/no_changed_list_at_all (0.00s)
=== RUN   TestAffectedScope_WideningTriggers
=== RUN   TestAffectedScope_WideningTriggers/flat_go.mod
    affected_scope_test.go:208: full = false, trigger = "", want the full module triggered by "go.mod"
=== RUN   TestAffectedScope_WideningTriggers/flat_go.sum
    affected_scope_test.go:208: full = false, trigger = "", want the full module triggered by "go.sum"
=== RUN   TestAffectedScope_WideningTriggers/flat_go.work
    affected_scope_test.go:208: full = false, trigger = "", want the full module triggered by "go.work"
=== RUN   TestAffectedScope_WideningTriggers/flat_go.work.sum
    affected_scope_test.go:208: full = false, trigger = "", want the full module triggered by "go.work.sum"
=== RUN   TestAffectedScope_WideningTriggers/flat_.moai/config/build-tags
    affected_scope_test.go:208: full = false, trigger = "", want the full module triggered by ".moai/config/build-tags"
=== RUN   TestAffectedScope_WideningTriggers/nested_module,_trigger_at_the_repository_root:_go.work
    affected_scope_test.go:223: full = false, trigger = "", want the full module triggered by "go.work" (a module-root filter ran first)
=== RUN   TestAffectedScope_WideningTriggers/nested_module,_trigger_at_the_repository_root:_.moai/config/build-tags
    affected_scope_test.go:223: full = false, trigger = "", want the full module triggered by ".moai/config/build-tags" (a module-root filter ran first)
=== RUN   TestAffectedScope_WideningTriggers/nested_module_without_a_trigger_narrows
    affected_scope_test.go:233: selected = [], want exactly [example.com/m/a example.com/m/b]
=== RUN   TestAffectedScope_WideningTriggers/a_nested_go.mod_anywhere_widens
    affected_scope_test.go:238: full = false, trigger = "", want the full module triggered by tools/go.mod
--- FAIL: TestAffectedScope_WideningTriggers (0.00s)
    --- FAIL: TestAffectedScope_WideningTriggers/flat_go.mod (0.00s)
    --- FAIL: TestAffectedScope_WideningTriggers/flat_go.sum (0.00s)
    --- FAIL: TestAffectedScope_WideningTriggers/flat_go.work (0.00s)
    --- FAIL: TestAffectedScope_WideningTriggers/flat_go.work.sum (0.00s)
    --- FAIL: TestAffectedScope_WideningTriggers/flat_.moai/config/build-tags (0.00s)
    --- FAIL: TestAffectedScope_WideningTriggers/nested_module,_trigger_at_the_repository_root:_go.work (0.00s)
    --- FAIL: TestAffectedScope_WideningTriggers/nested_module,_trigger_at_the_repository_root:_.moai/config/build-tags (0.00s)
    --- FAIL: TestAffectedScope_WideningTriggers/nested_module_without_a_trigger_narrows (0.00s)
    --- FAIL: TestAffectedScope_WideningTriggers/a_nested_go.mod_anywhere_widens (0.00s)
=== RUN   TestAffectedScope_RootPackageAndAncestorFallback
    affected_scope_test.go:248: selected = [], want exactly [example.com/m example.com/m/a]
    affected_scope_test.go:250: run patterns = [], want [. ./a]
    affected_scope_test.go:255: selected = [], want exactly [example.com/m/a example.com/m/b]
--- FAIL: TestAffectedScope_RootPackageAndAncestorFallback (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality	0.744s
FAIL
```
P1 ledger row, AC-PAS-002 (adoption-pending): command `go test ./internal/hook/quality -run '^TestAffectedScope_TestImportEdges$' -count=1 -v`, exit code 1, tree = parent of the M2 GREEN commit, verbatim stdout:
```
=== RUN   TestAffectedScope_TestImportEdges
    affected_scope_test.go:91: selected = [], want exactly [example.com/m/a example.com/m/b example.com/m/e example.com/m/f]
--- FAIL: TestAffectedScope_TestImportEdges (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality	0.519s
FAIL
```
P1 ledger row, AC-PAS-003 (adoption-pending): command `go test ./internal/hook/quality -run '^TestAffectedScope_Ownership$' -count=1 -v`, exit code 1, tree = parent of the M2 GREEN commit, verbatim stdout:
```
=== RUN   TestAffectedScope_Ownership
    affected_scope_test.go:112: selected = [], want exactly [example.com/m/p example.com/m/q example.com/m/r]
    affected_scope_test.go:114: unowned = [], want [docs/note.md]
    affected_scope_test.go:119: summary does not name docs/note.md as owned by no package:
--- FAIL: TestAffectedScope_Ownership (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality	0.483s
FAIL
```


### GREEN-M2 — core implemented (same command as RED-M2, flipped; tree = M2 GREEN commit)
Command: `go test ./internal/hook/quality -run '^TestAffectedScope_' -count=1 -v`
Exit code: 0 (verbatim result lines)
```
--- PASS: TestAffectedScope_DirectDependents (0.00s)
--- PASS: TestAffectedScope_TestImportEdges (0.00s)
--- PASS: TestAffectedScope_Ownership (0.00s)
--- PASS: TestAffectedScope_TestOnlyNarrows (0.00s)
--- PASS: TestAffectedScope_DeletedAndRenamed (0.00s)
    --- PASS: TestAffectedScope_DeletedAndRenamed/deleted_last_file_of_a_package (0.00s)
    --- PASS: TestAffectedScope_DeletedAndRenamed/rename_inside_a_package (0.00s)
    --- PASS: TestAffectedScope_DeletedAndRenamed/rename_across_packages_reports_both_directories (0.00s)
--- PASS: TestAffectedScope_DropAndEmpty (0.00s)
    --- PASS: TestAffectedScope_DropAndEmpty/a_selection_with_one_untested_package (0.00s)
    --- PASS: TestAffectedScope_DropAndEmpty/a_changed_list_holding_only_a_docs_file (0.00s)
    --- PASS: TestAffectedScope_DropAndEmpty/every_selected_package_is_untested (0.00s)
    --- PASS: TestAffectedScope_DropAndEmpty/no_changed_list_at_all (0.00s)
--- PASS: TestAffectedScope_WideningTriggers (0.00s)
    --- PASS: TestAffectedScope_WideningTriggers/flat_go.mod (0.00s)
    --- PASS: TestAffectedScope_WideningTriggers/flat_go.sum (0.00s)
    --- PASS: TestAffectedScope_WideningTriggers/flat_go.work (0.00s)
    --- PASS: TestAffectedScope_WideningTriggers/flat_go.work.sum (0.00s)
    --- PASS: TestAffectedScope_WideningTriggers/flat_.moai/config/build-tags (0.00s)
    --- PASS: TestAffectedScope_WideningTriggers/nested_module,_trigger_at_the_repository_root:_go.work (0.00s)
    --- PASS: TestAffectedScope_WideningTriggers/nested_module,_trigger_at_the_repository_root:_.moai/config/build-tags (0.00s)
    --- PASS: TestAffectedScope_WideningTriggers/nested_module_without_a_trigger_narrows (0.00s)
    --- PASS: TestAffectedScope_WideningTriggers/a_nested_go.mod_anywhere_widens (0.00s)
--- PASS: TestAffectedScope_RootPackageAndAncestorFallback (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	0.651s
```


### M3 measurement (plan.md B.3) — what git hands a pre-commit hook, per commit shape (tree 00e7843ab plus the M1/M2 commits; git as installed; throwaway repo under /private/tmp)
Command: `sh m3_measure.sh` (scratch script; the hook records the commit-accurate list `git diff --cached --name-only --no-renames -z` with the environment as git gave it, and the existing helper's form with every `gitenv.RepoScopingVars` entry removed). Exit code: 0. Verbatim output:
```
## plain: git commit (a/x.go staged, d/z.go modified unstaged)
  cwd=/private/tmp/t1673-tmp/m3/repo
  GIT_DIR=<unset>
  GIT_INDEX_FILE=.git/index
  commit-accurate : a/x.go 
  existing helper : a/x.go 
## plain: git commit -a
  cwd=/private/tmp/t1673-tmp/m3/repo
  GIT_DIR=<unset>
  GIT_INDEX_FILE=/private/tmp/t1673-tmp/m3/repo/.git/index.lock
  commit-accurate : a/x.go d/z.go 
  existing helper : a/x.go 
## plain: git commit d/z.go
  cwd=/private/tmp/t1673-tmp/m3/repo
  GIT_DIR=<unset>
  GIT_INDEX_FILE=/private/tmp/t1673-tmp/m3/repo/.git/next-index-61161.lock
  commit-accurate : d/z.go 
  existing helper : a/x.go 
## plain: staged rename old/o.go -> new/o.go
  cwd=/private/tmp/t1673-tmp/m3/repo
  GIT_DIR=<unset>
  GIT_INDEX_FILE=.git/index
  commit-accurate : new/o.go old/o.go 
  existing helper : new/o.go 
## plain: git commit --amend after staging f.go
  cwd=/private/tmp/t1673-tmp/m3/repo
  GIT_DIR=<unset>
  GIT_INDEX_FILE=.git/index
  commit-accurate : f.go 
  existing helper : f.go 
## linked worktree: git commit
  cwd=/private/tmp/t1673-tmp/m3/wt
  GIT_DIR=/private/tmp/t1673-tmp/m3/repo/.git/worktrees/wt
  GIT_INDEX_FILE=/private/tmp/t1673-tmp/m3/repo/.git/worktrees/wt/index
  commit-accurate : a/x.go 
  existing helper : a/x.go 
## linked worktree: git commit -a
  cwd=/private/tmp/t1673-tmp/m3/wt
  GIT_DIR=/private/tmp/t1673-tmp/m3/repo/.git/worktrees/wt
  GIT_INDEX_FILE=/private/tmp/t1673-tmp/m3/repo/.git/worktrees/wt/index.lock
  commit-accurate : a/x.go d/z.go 
  existing helper : a/x.go 
## linked worktree: git commit d/z.go
  cwd=/private/tmp/t1673-tmp/m3/wt
  GIT_DIR=/private/tmp/t1673-tmp/m3/repo/.git/worktrees/wt
  GIT_INDEX_FILE=/private/tmp/t1673-tmp/m3/repo/.git/worktrees/wt/next-index-61400.lock
  commit-accurate : d/z.go 
  existing helper : a/x.go 
```
Reading: under `git commit -a` and a path-limited commit git exports GIT_INDEX_FILE as an absolute path to a temporary index (`index.lock` / `next-index-N.lock`); under a plain commit or `--amend` it is the relative `.git/index`; in a linked worktree GIT_DIR is `<main>/.git/worktrees/<name>` and GIT_INDEX_FILE is absolute under it. The commit-accurate list is the one built from that index; the existing helper misses `d/z.go` under `-a`, returns `a/x.go` instead of `d/z.go` for a path-limited commit, and omits the old path of a rename — identical in the linked worktree. `--amend` yields exactly the newly staged `f.go` (the amended commit's own files are not re-reported). Mechanism chosen: read GIT_INDEX_FILE (made absolute against the hook's cwd) for this one query only, every other repository-location variable stays scrubbed, rename detection off (`--no-renames`), NUL-separated names (`-z`).

### RED-M3a — AC-PAS-005 adapter half, AC-PAS-008 go list half, AC-PAS-009 (adapters; tree = M2 GREEN plus the M3a RED test files and the adapter stubs)
The adapters are stubs returning an error, so each assertion fails on the right reason: no list and no graph is produced. The subtest `a_directory_outside_any_repository_is_a_query_failure` passes against the stub (a stub error is also an error); it is not a red cell. Exit code of each command: 1.

Command: `go test ./internal/hook/quality -run '^TestAffectedScope_GoListGraph$' -count=1 -v`
```
=== RUN   TestAffectedScope_GoListGraph
    affected_scope_adapters_test.go:72: listModuleGraph: affected-scope: package-graph adapter not implemented
--- FAIL: TestAffectedScope_GoListGraph (0.04s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality	0.753s
FAIL
```

Command: `go test ./internal/hook/quality -run '^TestAffectedScope_IncompleteGraphFallsBack$' -count=1 -v`
```
=== RUN   TestAffectedScope_IncompleteGraphFallsBack
    affected_scope_adapters_test.go:117: control: a clean module was refused: affected-scope: package-graph adapter not implemented
--- FAIL: TestAffectedScope_IncompleteGraphFallsBack (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality	0.505s
FAIL
```

Command: `go test ./internal/hook/quality -run '^TestAffectedScope_RealGitRenameDelete$' -count=1 -v`
```
=== RUN   TestAffectedScope_RealGitRenameDelete
=== RUN   TestAffectedScope_RealGitRenameDelete/cross-package_rename_keeps_the_old_package's_importers
    affected_scope_adapters_test.go:168: queryCommitChanges: affected-scope: commit-changes adapter not implemented
=== RUN   TestAffectedScope_RealGitRenameDelete/deleting_the_last_file_of_a_package_selects_its_updated_importers
    affected_scope_adapters_test.go:196: queryCommitChanges: affected-scope: commit-changes adapter not implemented
=== RUN   TestAffectedScope_RealGitRenameDelete/a_nested_module_root_reports_its_offset_inside_the_repository
    affected_scope_adapters_test.go:222: queryCommitChanges: affected-scope: commit-changes adapter not implemented
=== RUN   TestAffectedScope_RealGitRenameDelete/nothing_staged_is_an_empty_list,_not_an_error
    affected_scope_adapters_test.go:233: files = [], err = affected-scope: commit-changes adapter not implemented, want an empty list and no error
=== RUN   TestAffectedScope_RealGitRenameDelete/a_directory_outside_any_repository_is_a_query_failure
--- FAIL: TestAffectedScope_RealGitRenameDelete (1.22s)
    --- FAIL: TestAffectedScope_RealGitRenameDelete/cross-package_rename_keeps_the_old_package's_importers (0.44s)
    --- FAIL: TestAffectedScope_RealGitRenameDelete/deleting_the_last_file_of_a_package_selects_its_updated_importers (0.27s)
    --- FAIL: TestAffectedScope_RealGitRenameDelete/a_nested_module_root_reports_its_offset_inside_the_repository (0.30s)
    --- FAIL: TestAffectedScope_RealGitRenameDelete/nothing_staged_is_an_empty_list,_not_an_error (0.21s)
    --- PASS: TestAffectedScope_RealGitRenameDelete/a_directory_outside_any_repository_is_a_query_failure (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality	1.713s
FAIL
```

Command: `go test ./internal/hook/quality -run '^TestAffectedScope_ChangedSetMatchesCommit$' -count=1 -v`
```
=== RUN   TestAffectedScope_ChangedSetMatchesCommit
=== RUN   TestAffectedScope_ChangedSetMatchesCommit/plain_repository
    affected_scope_hook_test.go:149: (i) git commit (only a/x.go staged): the gate saw "ERROR: affected-scope: commit-changes adapter not implemented", the commit records "a/x.go"
    affected_scope_hook_test.go:149: (ii) git commit -a: the gate saw "ERROR: affected-scope: commit-changes adapter not implemented", the commit records "a/x.go d/z.go"
    affected_scope_hook_test.go:149: (iii) git commit d/z.go: the gate saw "ERROR: affected-scope: commit-changes adapter not implemented", the commit records "d/z.go"
    affected_scope_hook_test.go:149: (iv) staged rename old/o.go -> new/o.go: the gate saw "ERROR: affected-scope: commit-changes adapter not implemented", the commit records "new/o.go old/o.go"
    affected_scope_hook_test.go:149: (v) git commit --amend after staging f.go: the gate saw "ERROR: affected-scope: commit-changes adapter not implemented", the commit records "f.go"
=== RUN   TestAffectedScope_ChangedSetMatchesCommit/linked_worktree
    affected_scope_hook_test.go:158: (i) git commit (only a/x.go staged): the gate saw "ERROR: affected-scope: commit-changes adapter not implemented", the commit records "a/x.go"
    affected_scope_hook_test.go:158: (ii) git commit -a: the gate saw "ERROR: affected-scope: commit-changes adapter not implemented", the commit records "a/x.go d/z.go"
    affected_scope_hook_test.go:158: (iii) git commit d/z.go: the gate saw "ERROR: affected-scope: commit-changes adapter not implemented", the commit records "d/z.go"
    affected_scope_hook_test.go:158: (iv) staged rename old/o.go -> new/o.go: the gate saw "ERROR: affected-scope: commit-changes adapter not implemented", the commit records "new/o.go old/o.go"
    affected_scope_hook_test.go:158: (v) git commit --amend after staging f.go: the gate saw "ERROR: affected-scope: commit-changes adapter not implemented", the commit records "f.go"
--- FAIL: TestAffectedScope_ChangedSetMatchesCommit (1.99s)
    --- FAIL: TestAffectedScope_ChangedSetMatchesCommit/plain_repository (0.86s)
    --- FAIL: TestAffectedScope_ChangedSetMatchesCommit/linked_worktree (1.13s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality	2.615s
FAIL
```


### GREEN-M3a — adapters implemented (same four commands as RED-M3a, flipped; tree = M3a GREEN commit)

Command: `go test ./internal/hook/quality -run '^TestAffectedScope_GoListGraph$' -count=1 -v`
Exit code: 0
```
=== RUN   TestAffectedScope_GoListGraph
--- PASS: TestAffectedScope_GoListGraph (0.05s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	0.527s
```

Command: `go test ./internal/hook/quality -run '^TestAffectedScope_IncompleteGraphFallsBack$' -count=1 -v`
Exit code: 0
```
=== RUN   TestAffectedScope_IncompleteGraphFallsBack
--- PASS: TestAffectedScope_IncompleteGraphFallsBack (0.05s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	0.537s
```

Command: `go test ./internal/hook/quality -run '^TestAffectedScope_RealGitRenameDelete$' -count=1 -v`
Exit code: 0
```
=== RUN   TestAffectedScope_RealGitRenameDelete
=== RUN   TestAffectedScope_RealGitRenameDelete/cross-package_rename_keeps_the_old_package's_importers
=== RUN   TestAffectedScope_RealGitRenameDelete/deleting_the_last_file_of_a_package_selects_its_updated_importers
=== RUN   TestAffectedScope_RealGitRenameDelete/a_nested_module_root_reports_its_offset_inside_the_repository
=== RUN   TestAffectedScope_RealGitRenameDelete/nothing_staged_is_an_empty_list,_not_an_error
=== RUN   TestAffectedScope_RealGitRenameDelete/a_directory_outside_any_repository_is_a_query_failure
--- PASS: TestAffectedScope_RealGitRenameDelete (1.95s)
    --- PASS: TestAffectedScope_RealGitRenameDelete/cross-package_rename_keeps_the_old_package's_importers (0.48s)
    --- PASS: TestAffectedScope_RealGitRenameDelete/deleting_the_last_file_of_a_package_selects_its_updated_importers (0.64s)
    --- PASS: TestAffectedScope_RealGitRenameDelete/a_nested_module_root_reports_its_offset_inside_the_repository (0.47s)
    --- PASS: TestAffectedScope_RealGitRenameDelete/nothing_staged_is_an_empty_list,_not_an_error (0.35s)
    --- PASS: TestAffectedScope_RealGitRenameDelete/a_directory_outside_any_repository_is_a_query_failure (0.03s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	2.438s
```

Command: `go test ./internal/hook/quality -run '^TestAffectedScope_ChangedSetMatchesCommit$' -count=1 -v`
Exit code: 0
```
=== RUN   TestAffectedScope_ChangedSetMatchesCommit
=== RUN   TestAffectedScope_ChangedSetMatchesCommit/plain_repository
=== RUN   TestAffectedScope_ChangedSetMatchesCommit/linked_worktree
--- PASS: TestAffectedScope_ChangedSetMatchesCommit (4.34s)
    --- PASS: TestAffectedScope_ChangedSetMatchesCommit/plain_repository (2.55s)
    --- PASS: TestAffectedScope_ChangedSetMatchesCommit/linked_worktree (1.79s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	4.863s
```

Mutant probes executed against the GREEN adapters (temporary edits, reverted before the commit; `grep -n MUTANT` over the file prints nothing afterwards):
- dropping the hook-provided index (the scrubbed-only environment the old helper used): `TestAffectedScope_ChangedSetMatchesCommit` FAILS in both repository shapes — `(ii) git commit -a: the gate saw "a/x.go", the commit records "a/x.go d/z.go"` and `(iii) git commit d/z.go: the gate saw "a/x.go", the commit records "d/z.go"`.
- dropping `--no-renames` (the default name list): `TestAffectedScope_RealGitRenameDelete/cross-package_rename_keeps_the_old_package's_importers` FAILS — `changed list [e/x.go] must carry both the old path a/x.go and the new path e/x.go (rename detection off)`.

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## §G Override and Refusal Record

- 2026-10-18T01:41:03Z SPEC-PRECOMMIT-AFFECTED-SCOPE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter4.md
- 2026-10-18T01:41:12Z SPEC-PRECOMMIT-AFFECTED-SCOPE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter4.md
- 2026-10-18T01:42:39Z SPEC-PRECOMMIT-AFFECTED-SCOPE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter4.md

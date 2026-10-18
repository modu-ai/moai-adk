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


### RED-M3b — gate wiring: AC-PAS-006 gate half, AC-PAS-008 gate half, AC-PAS-011 runner half, AC-PAS-016, plus the argument-list wiring (tree = M3a GREEN plus the M3b RED test files, the config/seam scaffolding and a pass-through `scopeGoTestStep` stub)
The call site in `runToolchainTestStep` already routes through `scopeGoTestStep`, which is a pass-through stub, so the Go test step still runs `./...` and every assertion about the narrowed scope fails on the right reason.

Command: `go test ./internal/hook/quality -run '^TestQualityGate_AffectedScope_' -count=1 -v` — exit code 1 (the output below is cut to 240 columns; WARN log lines from the Python subtests omitted)
```
=== RUN   TestQualityGate_AffectedScope_EmptySelectionSkips
    affected_scope_gate_test.go:123: go test ran for a docs-only commit: [go vet ./... go test ./...]
    affected_scope_gate_test.go:126: the summary does not record the go test step as skipped:
        quality gate steps (3 configured):
          - go vet: executed in 282ms at 2026-10-18T11:18:01.298+09:00 — go vet ./...
          - golangci-lint: skipped — none of its config files exist in the project directory (.golangci.yml, .golangci.yaml, .golangci.toml, .golangci.json)
          - go test: executed in 25ms at 2026-10-18T11:18:01.580+09:00 — go test ./...
    affected_scope_gate_test.go:129: the skip carries no stated reason:
        quality gate steps (3 configured):
          - go vet: executed in 282ms at 2026-10-18T11:18:01.298+09:00 — go vet ./...
          - golangci-lint: skipped — none of its config files exist in the project directory (.golangci.yml, .golangci.yaml, .golangci.toml, .golangci.json)
          - go test: executed in 25ms at 2026-10-18T11:18:01.580+09:00 — go test ./...
    affected_scope_gate_test.go:132: the summary reports go test as executed:
        quality gate steps (3 configured):
          - go vet: executed in 282ms at 2026-10-18T11:18:01.298+09:00 — go vet ./...
          - golangci-lint: skipped — none of its config files exist in the project directory (.golangci.yml, .golangci.yaml, .golangci.toml, .golangci.json)
          - go test: executed in 25ms at 2026-10-18T11:18:01.580+09:00 — go test ./...
--- FAIL: TestQualityGate_AffectedScope_EmptySelectionSkips (0.32s)
=== RUN   TestQualityGate_AffectedScope_QueryFailureFallsBack
=== RUN   TestQualityGate_AffectedScope_QueryFailureFallsBack/changed-file_query_fails
    affected_scope_gate_test.go:184: the summary does not name the failed query "changed-file query" and the fallback:
        quality gate steps (3 configured):
          - go vet: executed in 227ms at 2026-10-18T11:18:01.608+09:00 — go vet ./...
          - golangci-lint: skipped — none of its config files exist in the project directory (.golangci.yml, .golangci.yaml, .golangci.toml, .golangci.json)
          - go test: executed in 32ms at 2026-10-18T11:18:01.837+09:00 — go test ./...
=== RUN   TestQualityGate_AffectedScope_QueryFailureFallsBack/package-graph_query_fails
    affected_scope_gate_test.go:184: the summary does not name the failed query "package-graph query" and the fallback:
        quality gate steps (3 configured):
          - go vet: executed in 222ms at 2026-10-18T11:18:01.872+09:00 — go vet ./...
          - golangci-lint: skipped — none of its config files exist in the project directory (.golangci.yml, .golangci.yaml, .golangci.toml, .golangci.json)
          - go test: executed in 21ms at 2026-10-18T11:18:02.094+09:00 — go test ./...
--- FAIL: TestQualityGate_AffectedScope_QueryFailureFallsBack (0.51s)
    --- FAIL: TestQualityGate_AffectedScope_QueryFailureFallsBack/changed-file_query_fails (0.26s)
    --- FAIL: TestQualityGate_AffectedScope_QueryFailureFallsBack/package-graph_query_fails (0.25s)
=== RUN   TestQualityGate_AffectedScope_ArgumentList
=== RUN   TestQualityGate_AffectedScope_ArgumentList/selection_narrows_to_module-root-relative_patterns
    affected_scope_gate_test.go:207: go test calls = [go test ./...], want [go test ./a ./b]
    affected_scope_gate_test.go:210: the summary lacks the selected/dropped counts:
        quality gate steps (3 configured):
          - go vet: executed in 249ms at 2026-10-18T11:18:02.130+09:00 — go vet ./...
          - golangci-lint: skipped — none of its config files exist in the project directory (.golangci.yml, .golangci.yaml, .golangci.toml, .golangci.json)
          - go test: executed in 22ms at 2026-10-18T11:18:02.381+09:00 — go test ./...
=== RUN   TestQualityGate_AffectedScope_ArgumentList/a_widening_trigger_runs_the_full_module_and_says_why
    affected_scope_gate_test.go:230: the summary does not state the widening and its trigger:
        quality gate steps (3 configured):
          - go vet: executed in 305ms at 2026-10-18T11:18:02.467+09:00 — go vet ./...
          - golangci-lint: skipped — none of its config files exist in the project directory (.golangci.yml, .golangci.yaml, .golangci.toml, .golangci.json)
          - go test: executed in 28ms at 2026-10-18T11:18:02.772+09:00 — go test ./...
=== RUN   TestQualityGate_AffectedScope_ArgumentList/build_tags_survive_the_replaced_argument_list
    affected_scope_gate_test.go:243: go calls = [go vet -tags=goolm ./... go test -tags=goolm ./...], want one "go test -tags=goolm ./a ./b"
--- FAIL: TestQualityGate_AffectedScope_ArgumentList (1.01s)
    --- FAIL: TestQualityGate_AffectedScope_ArgumentList/selection_narrows_to_module-root-relative_patterns (0.29s)
    --- FAIL: TestQualityGate_AffectedScope_ArgumentList/a_widening_trigger_runs_the_full_module_and_says_why (0.34s)
    --- FAIL: TestQualityGate_AffectedScope_ArgumentList/build_tags_survive_the_replaced_argument_list (0.33s)
=== RUN   TestQualityGate_AffectedScope_ContextIsolation
=== RUN   TestQualityGate_AffectedScope_ContextIsolation/without_the_pre-commit_flag_the_Go_test_step_is_test_./...
=== RUN   TestQualityGate_AffectedScope_ContextIsolation/mode_full_restores_test_./...
=== RUN   TestQualityGate_AffectedScope_ContextIsolation/the_vet_step_is_identical_with_and_without_the_flag
=== RUN   TestQualityGate_AffectedScope_ContextIsolation/node_test_step_is_identical_with_the_flag
=== RUN   TestQualityGate_AffectedScope_ContextIsolation/python_test_step_is_identical_with_the_flag
--- PASS: TestQualityGate_AffectedScope_ContextIsolation (2.40s)
    --- PASS: TestQualityGate_AffectedScope_ContextIsolation/without_the_pre-commit_flag_the_Go_test_step_is_test_./... (0.30s)
    --- PASS: TestQualityGate_AffectedScope_ContextIsolation/mode_full_restores_test_./... (0.24s)
    --- PASS: TestQualityGate_AffectedScope_ContextIsolation/the_vet_step_is_identical_with_and_without_the_flag (0.51s)
    --- PASS: TestQualityGate_AffectedScope_ContextIsolation/node_test_step_is_identical_with_the_flag (0.45s)
    --- PASS: TestQualityGate_AffectedScope_ContextIsolation/python_test_step_is_identical_with_the_flag (0.90s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality	5.090s
FAIL
```
Regression guards (green today by construction, as the acceptance matrix classifies them): `TestQualityGate_AffectedScope_ContextIsolation` (AC-PAS-010) passes above against the pass-through stub; `TestQualityGate_StepBudgetsUnchanged` (AC-PAS-014) passes in both shapes. Their mutant probes are recorded under the GREEN entries.

Command: `go test ./internal/cli -run '^TestRunGatePreCommitInvalidScopeModeRejected$' -count=1 -v` — exit code 1 (tail)
```
          - typecheck: skipped — no default for this language; set gate.typecheck.command to enable one
          - golangci-lint: skipped — none of its config files exist in the project directory (.golangci.yml, .golangci.yaml, .golangci.toml, .golangci.json)
          - go test: executed in 709ms at 2026-10-18T11:20:37.614+09:00 — go test ./...
    gate_precommit_scope_test.go:35: the gate ran its steps before rejecting the mode:
        quality gate failed
        quality gate failed: go test
        
        --- FAIL: TestPreExistingFailure (0.00s)
            fail_test.go:6: pre-existing project-wide failure
        FAIL
        FAIL	fixture	0.243s
        FAIL
        
        quality gate steps (4 configured):
          - go vet: executed in 405ms at 2026-10-18T11:20:37.205+09:00 — go vet ./...
          - typecheck: skipped — no default for this language; set gate.typecheck.command to enable one
          - golangci-lint: skipped — none of its config files exist in the project directory (.golangci.yml, .golangci.yaml, .golangci.toml, .golangci.json)
          - go test: executed in 709ms at 2026-10-18T11:20:37.614+09:00 — go test ./...
--- FAIL: TestRunGatePreCommitInvalidScopeModeRejected (1.13s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.612s
FAIL
```

Command: `go test ./internal/cli -run '^TestPreCommitAffectedScopeE2E$' -count=1 -v` — exit code 1
```
=== RUN   TestPreCommitAffectedScopeE2E
=== RUN   TestPreCommitAffectedScopeE2E/a_change_to_a_runs_exactly_a_and_b
    precommit_affected_scope_e2e_test.go:153: recorded go test calls = "test ./...", want exactly "test ./a ./b"
        commit output:
        graph-freshness: step disabled in gate config (explicit skip notice)
        
        quality gate steps (3 configured):
          - go vet: disabled — turned off by the gate.disabled_steps setting
          - golangci-lint: disabled — turned off by the gate.disabled_steps setting
          - go test: executed in 252ms at 2026-10-18T11:21:15.562+09:00 — go test ./...
=== RUN   TestPreCommitAffectedScopeE2E/a_failing_dependent_blocks_the_commit_and_the_unaffected_package_never_runs
    precommit_affected_scope_e2e_test.go:176: c's test ran: the marker file it writes exists; the unaffected package was tested
--- FAIL: TestPreCommitAffectedScopeE2E (25.42s)
    --- FAIL: TestPreCommitAffectedScopeE2E/a_change_to_a_runs_exactly_a_and_b (2.20s)
    --- FAIL: TestPreCommitAffectedScopeE2E/a_failing_dependent_blocks_the_commit_and_the_unaffected_package_never_runs (1.78s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	26.602s
FAIL
```

Command: `env PROBE_DIR=/private/tmp/t1673-tmp/probe sh .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/evidence/probe016.sh /private/tmp/t1673-bin/moai-base` — exit code 1. Judging build: `moai` built from `00e7843ab` (tree HEAD at build time; the Go code is identical to `3642a3a9b`, the tree LEDGER-PAS-016-A measured). The four RED lines reproduce the ledger cell:
```
AC-016/001 change a: RED  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 1), wanted go argv 'test ./a ./b', got 'test ./...')
AC-004 test-only change: RED  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 1), wanted go argv 'test ./a', got 'test ./...')
AC-006 docs-only change: RED  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 0), wanted go argv 'none', got 'test ./...')
AC-011 mode: bogus: RED  (gate exit=0, message names the key=no)
exit=1
```


### GREEN-M3b — gate wiring (same commands as RED-M3b, flipped; tree = M3b GREEN commit)

Command: `go test ./internal/hook/quality -run '^TestQualityGate_AffectedScope_EmptySelectionSkips$' -count=1 -v`
Exit code: 0
```
=== RUN   TestQualityGate_AffectedScope_EmptySelectionSkips
--- PASS: TestQualityGate_AffectedScope_EmptySelectionSkips (0.24s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	0.904s
```

Command: `go test ./internal/hook/quality -run '^TestQualityGate_AffectedScope_QueryFailureFallsBack$' -count=1 -v`
Exit code: 0
```
=== RUN   TestQualityGate_AffectedScope_QueryFailureFallsBack
=== RUN   TestQualityGate_AffectedScope_QueryFailureFallsBack/changed-file_query_fails
=== RUN   TestQualityGate_AffectedScope_QueryFailureFallsBack/package-graph_query_fails
--- PASS: TestQualityGate_AffectedScope_QueryFailureFallsBack (0.55s)
    --- PASS: TestQualityGate_AffectedScope_QueryFailureFallsBack/changed-file_query_fails (0.28s)
    --- PASS: TestQualityGate_AffectedScope_QueryFailureFallsBack/package-graph_query_fails (0.27s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	1.429s
```

Command: `go test ./internal/hook/quality -run '^TestQualityGate_AffectedScope_ArgumentList$' -count=1 -v`
Exit code: 0
```
=== RUN   TestQualityGate_AffectedScope_ArgumentList
=== RUN   TestQualityGate_AffectedScope_ArgumentList/selection_narrows_to_module-root-relative_patterns
=== RUN   TestQualityGate_AffectedScope_ArgumentList/a_widening_trigger_runs_the_full_module_and_says_why
=== RUN   TestQualityGate_AffectedScope_ArgumentList/build_tags_survive_the_replaced_argument_list
--- PASS: TestQualityGate_AffectedScope_ArgumentList (1.18s)
    --- PASS: TestQualityGate_AffectedScope_ArgumentList/selection_narrows_to_module-root-relative_patterns (0.49s)
    --- PASS: TestQualityGate_AffectedScope_ArgumentList/a_widening_trigger_runs_the_full_module_and_says_why (0.40s)
    --- PASS: TestQualityGate_AffectedScope_ArgumentList/build_tags_survive_the_replaced_argument_list (0.29s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	2.104s
```

Command: `go test ./internal/hook/quality -run '^TestQualityGate_AffectedScope_ContextIsolation$' -count=1 -v`
Exit code: 0
```
=== RUN   TestQualityGate_AffectedScope_ContextIsolation
=== RUN   TestQualityGate_AffectedScope_ContextIsolation/without_the_pre-commit_flag_the_Go_test_step_is_test_./...
=== RUN   TestQualityGate_AffectedScope_ContextIsolation/mode_full_restores_test_./...
=== RUN   TestQualityGate_AffectedScope_ContextIsolation/the_vet_step_is_identical_with_and_without_the_flag
=== RUN   TestQualityGate_AffectedScope_ContextIsolation/node_test_step_is_identical_with_the_flag
=== RUN   TestQualityGate_AffectedScope_ContextIsolation/python_test_step_is_identical_with_the_flag
2026/10/18 11:31:50 WARN no sources for step: treating as skip step=mypy extensions=.py,.pyi hint="add sources, or set gate.disabled_steps in .moai/config/sections/gate.yaml"
2026/10/18 11:31:51 WARN no sources for step: treating as skip step=mypy extensions=.py,.pyi hint="add sources, or set gate.disabled_steps in .moai/config/sections/gate.yaml"
--- PASS: TestQualityGate_AffectedScope_ContextIsolation (2.01s)
    --- PASS: TestQualityGate_AffectedScope_ContextIsolation/without_the_pre-commit_flag_the_Go_test_step_is_test_./... (0.24s)
    --- PASS: TestQualityGate_AffectedScope_ContextIsolation/mode_full_restores_test_./... (0.23s)
    --- PASS: TestQualityGate_AffectedScope_ContextIsolation/the_vet_step_is_identical_with_and_without_the_flag (0.47s)
    --- PASS: TestQualityGate_AffectedScope_ContextIsolation/node_test_step_is_identical_with_the_flag (0.43s)
    --- PASS: TestQualityGate_AffectedScope_ContextIsolation/python_test_step_is_identical_with_the_flag (0.65s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	2.569s
```

Command: `go test ./internal/hook/quality -run '^TestQualityGate_StepBudgetsUnchanged$' -count=1 -v`
Exit code: 0
```
=== RUN   TestQualityGate_StepBudgetsUnchanged
=== RUN   TestQualityGate_StepBudgetsUnchanged/standalone
=== RUN   TestQualityGate_StepBudgetsUnchanged/pre-commit_affected_scope
--- PASS: TestQualityGate_StepBudgetsUnchanged (0.01s)
    --- PASS: TestQualityGate_StepBudgetsUnchanged/standalone (0.00s)
    --- PASS: TestQualityGate_StepBudgetsUnchanged/pre-commit_affected_scope (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	0.678s
```

### GREEN-M3b (CLI) — runner-side rejection and the real-hook end-to-end

Command: `go test ./internal/cli -run '^TestRunGatePreCommitInvalidScopeModeRejected$' -count=1 -v`
Exit code: 0
```
=== RUN   TestRunGatePreCommitInvalidScopeModeRejected
--- PASS: TestRunGatePreCommitInvalidScopeModeRejected (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.626s
```

Command: `go test ./internal/cli -run '^TestRunGatePreCommitFullScopeModeAccepted$' -count=1 -v`
Exit code: 0
```
=== RUN   TestRunGatePreCommitFullScopeModeAccepted
--- PASS: TestRunGatePreCommitFullScopeModeAccepted (3.27s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	5.223s
```

Command: `go test ./internal/cli -run '^TestRunGateStandaloneIgnoresScopeMode$' -count=1 -v`
Exit code: 0
```
=== RUN   TestRunGateStandaloneIgnoresScopeMode
--- PASS: TestRunGateStandaloneIgnoresScopeMode (1.47s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.059s
```

Command: `go test ./internal/cli -run '^TestPreCommitAffectedScopeE2E$' -count=1 -v`
Exit code: 0
```
=== RUN   TestPreCommitAffectedScopeE2E
=== RUN   TestPreCommitAffectedScopeE2E/a_change_to_a_runs_exactly_a_and_b
=== RUN   TestPreCommitAffectedScopeE2E/a_failing_dependent_blocks_the_commit_and_the_unaffected_package_never_runs
--- PASS: TestPreCommitAffectedScopeE2E (71.78s)
    --- PASS: TestPreCommitAffectedScopeE2E/a_change_to_a_runs_exactly_a_and_b (2.41s)
    --- PASS: TestPreCommitAffectedScopeE2E/a_failing_dependent_blocks_the_commit_and_the_unaffected_package_never_runs (2.61s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	74.051s
```

Mutant probe, AC-PAS-014 (temporary edit, reverted before the commit; `grep -rn MUTANT internal/hook` prints nothing from this card): replacing the test step's call-site budget in `runToolchainTestStep` with `2*g.config.TestTimeout` leaves every YAML and default value unchanged but turns `TestQualityGate_StepBudgetsUnchanged` red:
```
--- FAIL: TestQualityGate_StepBudgetsUnchanged (0.01s)
    --- FAIL: TestQualityGate_StepBudgetsUnchanged/standalone (0.00s)
        affected_scope_gate_test.go:380: step "go test" received budget 4m0s, want 2m0s
    --- FAIL: TestQualityGate_StepBudgetsUnchanged/pre-commit_affected_scope (0.00s)
        affected_scope_gate_test.go:380: step "go test" received budget 4m0s, want 2m0s
FAIL
```
Wider M3 verification (this run, this tree): `go test ./internal/hook/quality -count=1` exit 0 (`ok 36.928s`, whole package); `go test ./internal/cli -run '^(TestRunGate|TestGAR_|TestGateCmd_|TestPrecommitE2E|TestUpdateGateYAML|TestPreCommitHook_|TestPreCommitRelocation_)' -count=1 -v` exit 0, 40 PASS lines, no FAIL (targeted selection only, ruling 35e9; the full internal/cli suite was not run). The existing tests `TestRunGatePreCommitOptInRunsHeavyGate` and `TestRunGatePreCommitOptInPassingProjectPasses` run a non-git fixture, so the changed-file query fails and the gate falls back to the full module; both stay green by that fallback.


### RED-M5 — AC-PAS-013 (suppression walk reaches testdata; tree = M3b GREEN plus the M5 RED test file)
Why red: the walk has no testdata exclusion, so the Go fixture (whose @MX:REASON precedes its marker, spec.md F8) is reported in both the fixture-only case and the mixed case. The third subtest (an ancestor directory named testdata above the project root) passes today and is the guard for the GREEN design, not a red cell.

Command: `go test ./internal/hook/quality -run '^TestSuppressionWalkSkipsTestdata$' -count=1 -v`
Exit code: 1
```
=== RUN   TestSuppressionWalkSkipsTestdata
=== RUN   TestSuppressionWalkSkipsTestdata/the_fixture_alone_is_not_reported
    astgrep_gate_testdata_walk_test.go:46: the testdata fixture was examined: ok = false
        ast-grep suppression policy violations:
        
        [SUPPRESSION_WITHOUT_REASON] ast-grep suppression at /private/tmp/t1673-tmp/TestSuppressionWalkSkipsTestdatathe_fixture_alone_is_not_report60559039/001/internal/astgrep/testdata/fixtures/go/suppressed.go:15 requires adjacent '// @MX:REASON <rationale>' on next line
=== RUN   TestSuppressionWalkSkipsTestdata/production_code_under_a_fixtures_segment_is_still_reported
    astgrep_gate_testdata_walk_test.go:63: violations = 2, want exactly 1 (the production file):
        ast-grep suppression policy violations:
        
        [SUPPRESSION_WITHOUT_REASON] ast-grep suppression at /private/tmp/t1673-tmp/TestSuppressionWalkSkipsTestdataproduction_code_under_a_fixture3755740873/001/internal/astgrep/testdata/fixtures/go/suppressed.go:15 requires adjacent '// @MX:REASON <rationale>' on next line
        [SUPPRESSION_WITHOUT_REASON] ast-grep suppression at /private/tmp/t1673-tmp/TestSuppressionWalkSkipsTestdataproduction_code_under_a_fixture3755740873/001/internal/prod/fixtures/bad.go:4 requires adjacent '// @MX:REASON <rationale>' on next line
    astgrep_gate_testdata_walk_test.go:66: the violation is not the production file's:
        ast-grep suppression policy violations:
        
        [SUPPRESSION_WITHOUT_REASON] ast-grep suppression at /private/tmp/t1673-tmp/TestSuppressionWalkSkipsTestdataproduction_code_under_a_fixture3755740873/001/internal/astgrep/testdata/fixtures/go/suppressed.go:15 requires adjacent '// @MX:REASON <rationale>' on next line
        [SUPPRESSION_WITHOUT_REASON] ast-grep suppression at /private/tmp/t1673-tmp/TestSuppressionWalkSkipsTestdataproduction_code_under_a_fixture3755740873/001/internal/prod/fixtures/bad.go:4 requires adjacent '// @MX:REASON <rationale>' on next line
=== RUN   TestSuppressionWalkSkipsTestdata/a_testdata_directory_above_the_project_root_does_not_hide_production_code
--- FAIL: TestSuppressionWalkSkipsTestdata (0.03s)
    --- FAIL: TestSuppressionWalkSkipsTestdata/the_fixture_alone_is_not_reported (0.01s)
    --- FAIL: TestSuppressionWalkSkipsTestdata/production_code_under_a_fixtures_segment_is_still_reported (0.01s)
    --- PASS: TestSuppressionWalkSkipsTestdata/a_testdata_directory_above_the_project_root_does_not_hide_production_code (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality	0.843s
FAIL
```


LEDGER-PAS-013-A re-run (real tool, RED, this run): command `sh ledger013.sh <moai>` = the ledger recipe (project with go.mod, the fixture copied to internal/astgrep/testdata/fixtures/go/suppressed.go, the repository's .moai/astgrep-rules copied, and the ledger's gate.yaml) then `env CLAUDE_PROJECT_DIR=<project> <moai> gate`. Judging build: `/private/tmp/t1673-bin/moai-base`, built from tree HEAD `00e7843ab` (the suppression walk is identical at the M3b GREEN tree). Exit code: 1. Output (project path elided):
```
ast-grep suppression policy violations:

[SUPPRESSION_WITHOUT_REASON] ast-grep suppression at <project>/internal/astgrep/testdata/fixtures/go/suppressed.go:15 requires adjacent '// @MX:REASON <rationale>' on next line

quality gate steps (3 configured):
  - go vet: disabled — turned off by the gate.disabled_steps setting
  - golangci-lint: disabled — turned off by the gate.disabled_steps setting
  - go test: not reached
          
   ERROR  
          
  Quality gate failed.                                                                                                

exit=1
```


### GREEN-M5 — AC-PAS-013 unit half (same command as RED-M5, flipped; tree = M5 GREEN commit). The real-tool half (LEDGER-PAS-013-A) is re-run against a binary built from the final tree in the M6 entry.

Command: `go test ./internal/hook/quality -run '^TestSuppressionWalkSkipsTestdata$' -count=1 -v`
Exit code: 0
```
=== RUN   TestSuppressionWalkSkipsTestdata
=== RUN   TestSuppressionWalkSkipsTestdata/the_fixture_alone_is_not_reported
=== RUN   TestSuppressionWalkSkipsTestdata/production_code_under_a_fixtures_segment_is_still_reported
=== RUN   TestSuppressionWalkSkipsTestdata/a_testdata_directory_above_the_project_root_does_not_hide_production_code
--- PASS: TestSuppressionWalkSkipsTestdata (0.01s)
    --- PASS: TestSuppressionWalkSkipsTestdata/the_fixture_alone_is_not_reported (0.00s)
    --- PASS: TestSuppressionWalkSkipsTestdata/production_code_under_a_fixtures_segment_is_still_reported (0.00s)
    --- PASS: TestSuppressionWalkSkipsTestdata/a_testdata_directory_above_the_project_root_does_not_hide_production_code (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	0.566s
```


### M6 — verification sweep and dogfood (tree = HEAD `9d114393cac23a3a44dc96261ce7936cd36f11cc`, working tree clean before the evidence commit)

**Judging build.** `/private/tmp/t1673-bin/moai`, built from this tree with `go build -buildvcs=false -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Commit=9d114393cac23a3a44dc96261ce7936cd36f11cc -X github.com/modu-ai/moai-adk/pkg/version.Version=t1673-final" -o /private/tmp/t1673-bin/moai ./cmd/moai` immediately after `git rev-parse HEAD` printed `9d114393cac23a3a44dc96261ce7936cd36f11cc` and `git status --short` printed nothing. `moai version` prints `t1673-final   9d114393c   built unknown`. (The default `go build` VCS stamp of the worktree read revision `478285ab4` with `vcs.modified=true`, an older ancestor, so it was NOT used as provenance.)

**probe016.sh against that build (AC-PAS-016 / -001 / -004 / -006 / -011 real-binary half; RED-now cell in LEDGER-PAS-016-A)** — command `env PROBE_DIR=/private/tmp/t1673-tmp/probe sh .moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/evidence/probe016.sh /private/tmp/t1673-bin/moai`, exit code 0:
```
AC-016/001 change a: GREEN  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 1), wanted go argv 'test ./a ./b', got 'test ./a ./b')
AC-004 test-only change: GREEN  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 1), wanted go argv 'test ./a', got 'test ./a')
AC-006 docs-only change: GREEN  (commit exit=0, HEAD moved=yes, go test calls=0 (wanted 0), wanted go argv 'none', got 'none')
AC-011 mode: bogus: GREEN  (gate exit=1, message names the key=yes)
exit=0
```

**LEDGER-PAS-013-A (AC-PAS-013 real-tool half), same recipe as the RED re-run, judging build `/private/tmp/t1673-bin/moai` (9d114393c)** — command `sh ledger013.sh /private/tmp/t1673-bin/moai` (scratch script = the ledger recipe), exit code 0, no suppression line:
```
graph-freshness: step disabled in gate config (explicit skip notice)

quality gate steps (3 configured):
  - go vet: disabled — turned off by the gate.disabled_steps setting
  - golangci-lint: disabled — turned off by the gate.disabled_steps setting
  - go test: skipped — turned off by the gate.skip_tests setting
exit=0
```

**G4 closed — the real checker over a copy of the whole tree's source directories** (`internal`, `cmd`, `pkg` copied next to the recipe's gate.yaml and the repository's ast-grep rules; command `sh ledger013_full.sh /private/tmp/t1673-bin/moai`), exit code 0; the ast-grep rule findings after the suppression step are advisory notices, not suppression-policy violations (head of the output):
    copied:     5152 .go files,       38 of them under testdata
    suppression lines: 0
    graph-freshness: step disabled in gate config (explicit skip notice)
    ast-grep domain rule scan results:

**probe009.sh (immutable RED witness of the pre-change helper, never a green condition)** — exit code 1, as recorded in LEDGER-PAS-009-A; it hard-codes the old helper's query and does not call the new adapter:
```
(ii)  git commit -a            (a/x.go staged, d/z.go modified unstaged)
  commit records : a/x.go d/z.go 
  helper form    : a/x.go    [DIFFERENT]
(iii) git commit d/z.go        (path-limited)
  commit records : d/z.go 
  helper form    : a/x.go    [DIFFERENT]
(iv)  git mv old/o.go new/o.go (rename)
  commit records : new/o.go old/o.go 
  helper form    : new/o.go    [DIFFERENT]
exit=1
```

**Dogfood (plan.md M6): this SPEC's own diff (`git diff --name-only 3642a3a9b HEAD`, 33 files) fed to the real resolver and to an independent jq computation over `go list -e -json ./...`.** Resolver run through a scratch test (not committed) against the real adapters; hand computation by `dogfood_hand.sh` (selected = the four packages owning a source change — internal/cli, internal/config, internal/hook/quality, internal/template (the embedded gate.yaml) — plus every package whose regular, test or external-test imports name one of them). Exit code 0, output:
```
hand selected:           57
resolver selected:       57
hand dropped:             3
resolver dropped:         3
selected sets: IDENTICAL
dropped sets:  IDENTICAL
dropped packages:
github.com/modu-ai/moai-adk/cmd/moai
github.com/modu-ai/moai-adk/internal/closure/closuretest
github.com/modu-ai/moai-adk/internal/template/scripts
internal/cli importers selected (own packages beyond the four owners):
```
The resolver also reported the twelve `.moai/**` files (specs, reports, local gate.yaml — none embedded by a package) as owned by none, and no widening trigger fired. This SPEC's own commit would therefore select 57 packages, `internal/cli` among them, and fail by budget overrun at the unchanged 120 s: the documented consequence of the Q1=(B) ruling (spec.md § C, R2), not a defect of the selection.

**Measurements (G2, Q4).** G2, `go list` wall time: `go list -e -json=<fields> ./...` on this tree, warm build cache, 176 packages, 313459 bytes of JSON, started 11:08:36 and finished 11:08:39 local time (about 3 s), none errored or incomplete (`jq -s '[.[]|select(.Error!=null or .Incomplete==true)]'` printed `[]`). Q4 (project-wide vet / lint / typecheck runtime in the pre-commit context) was NOT measured: a project-wide `go vet ./...` and golangci-lint run over 176 packages is heavy-slot work and the lease was held by another lane (below); it stays a recorded Gap and gates only the post-landing handoff runbook (plan.md § E pre-condition 3).

**Slot.** `moai slot acquire --resource go-test-heavy --name lane-27 --command "t1673 race+cover internal/hook/quality internal/config"` exited 3; `moai slot status --resource go-test-heavy` printed `slot go-test-heavy: held`, session c5c5d1ae-4185-43c8-8fab-4651c3995350, name lane-25, command `t1632 landing remeasure`, since 2026-10-18T02:26:13Z, bound 1h40m0s (ends 2026-10-18T04:06:13Z); re-read once more later in the run with the same result. This run therefore never held the lease and has nothing to release. Consequence: no full-package `-race` run of internal/hook/quality or internal/config was taken (scoped `-race` below). One deviation, disclosed: `go test ./internal/template -count=1` (440 s, exit 0) was run during M1 before the lease question arose, without a lease.

**Scoped `-race` runs (this run, this tree)** — `go test -race -count=1 ./internal/hook/quality -run '^(TestAffectedScope_|TestQualityGate_AffectedScope_|TestQualityGate_StepBudgetsUnchanged|TestSuppressionWalkSkipsTestdata|TestCommitQueryEnv|TestGoListPackageProblem|TestScopeHookHelper)' -v`: exit 0, 22 top-level PASS, `ok github.com/modu-ai/moai-adk/internal/hook/quality 22.616s`, no DATA RACE; `go test -race -count=1 ./internal/config -run '^TestGatePreCommitTestScope$' -v`: exit 0, `ok github.com/modu-ai/moai-adk/internal/config 1.477s`.

**Coverage (E3).** `go test -p 1 -count=1 -coverprofile=… ./internal/hook/quality`: exit 0, `ok ... 40.264s coverage: 93.3% of statements`; new-file functions: affected_scope.go 97.9-100%, affected_scope_gate.go 96-100%, affected_scope_queries.go `listModuleGraph` 77.8% / `queryCommitChanges` 85.7% / `runScopeQuery` 88.2% / the rest 100% (uncovered: decode-error and module-mismatch branches), `walkSourceFiles` 100%. `go test -p 1 -count=1 -coverprofile=… ./internal/config`: `coverage: 85.6% of statements` with the one BASELINE failure `TestNoHardcodedContexts` (reads scripts/release.sh, untouched; not attributable); `ValidateGateTestScopeMode` 100%. `internal/cli` package coverage is not measurable (the full suite is barred, ruling 35e9); the targeted `-run '^(TestRunGate|TestGAR_|TestPreCommitAffectedScopeE2E)'` selection reports `runGate` 92.9% and `mapConfigGateToQuality` 82.4% (package-wide figure of that selection: 4.8%, meaningless as a package number).

**Lint (E5).** `golangci-lint run --timeout=3m ./internal/hook/quality/... ./internal/config/...`: exit 0, `0 issues.` (baseline at HEAD 00e7843ab for quality was also `0 issues.`). `golangci-lint run --timeout=5m ./internal/cli/`: exit 1, 3 issues, all `errcheck` in `internal/cli/factory_run_adopt.go` (lines 79, 91, 103) — a file this card does not touch (`git diff --stat 3642a3a9b HEAD -- internal/cli/factory_run_adopt.go` is empty): BASELINE, no NEW issue in the touched files. `moai spec lint SPEC-PRECOMMIT-AFFECTED-SCOPE-001` (the final build): exit 0, `0 error(s), 0 warning(s)` and one INFO `OwnershipTransitionUnmeasured` (the M1 commit carries no Authored-By-Agent trailer because the dispatch fixes the last commit-message line as `🗿 MoAI`).

**Preservation guards (AC-PAS-014, AC-PAS-015), commands against `3642a3a9b`:**
- `git diff -U0 3642a3a9b -- .moai/config/sections/gate.yaml internal/template/templates/.moai/config/sections/gate.yaml`: the only hunk in each file is the 7-line `test_scope` block under `pre_commit`; no `timeouts:` line is in the diff (`grep -c` over the diff for timeouts/vet/lint/test/typecheck/lock_wait keys: 0).
- `git diff -U0 3642a3a9b -- '*_test.go'` (1616 lines): added lines matching `t\.Skip`: 0; removed lines matching `func Test`: 0; removed lines in total: 0; added lines containing `hooksPath`: 0.
- `git diff -U0 3642a3a9b -- '*.go' '*.yaml' ':(exclude)*_test.go'` (838 lines): added lines containing `hooksPath`: 0; added lines containing a git `config` or `-c` argument: 0. `git diff --stat 3642a3a9b -- internal/template/templates/.git_hooks internal/cli/hook_install_precommit.go`: empty.
- The step-budget definition sites are unchanged (`TestQualityGate_StepBudgetsUnchanged` PASS in both shapes) and its mutant (`2*g.config.TestTimeout` at the test-step call site) turns it red (recorded under GREEN-M3b).

**Mutant probes executed (temporary edits, all reverted; `grep -rn MUTANT internal/hook/quality` over non-test files prints nothing):**
| AC | mutant | observed |
|----|--------|----------|
| AC-PAS-001 / S4 | transitive-closure expansion after the direct-dependents step | `TestAffectedScope_DirectDependents` FAIL: `selected = [a b c], want exactly [a b]` |
| AC-PAS-002 | importer edges read from regular imports only | `TestAffectedScope_TestImportEdges` FAIL: `selected = [a b], want exactly [a b e f]` |
| AC-PAS-003 | directory-of-the-file rule for `.go` files under testdata | `TestAffectedScope_Ownership` PASSED against this mutant (the sibling `p/testdata/in.json` already owns `p`), so the AC's own change list cannot discriminate it; added `TestAffectedScope_OwnershipTestdataGoFixtureAlone`, which FAILS the mutant: `selected = [], want exactly [example.com/m/p]` |
| AC-PAS-004 | every changed file counts as a source change | `TestAffectedScope_TestOnlyNarrows` FAIL: `selected = [a b], want exactly [a]` |
| AC-PAS-007 | module-root filter before the trigger check | `TestAffectedScope_WideningTriggers` FAIL on both nested-module cases (go.work, .moai/config/build-tags) |
| AC-PAS-008 | trust go list's zero exit status (no errored/incomplete check) | `TestAffectedScope_IncompleteGraphFallsBack` FAIL: `an incomplete package graph was accepted; the adapter trusted go list's zero exit status` |
| AC-PAS-009 | scrubbed-only environment (drop the hook-provided index) | FAIL on (ii) and (iii) in both repository shapes (recorded under GREEN-M3a) |
| AC-PAS-005 | default name list (drop `--no-renames`) | `TestAffectedScope_RealGitRenameDelete/cross-package_rename…` FAIL (recorded under GREEN-M3a) |
| AC-PAS-014 | `2*g.config.TestTimeout` at the test-step call site | `TestQualityGate_StepBudgetsUnchanged` FAIL: `received budget 4m0s, want 2m0s` (recorded under GREEN-M3b) |

**Disposal of the carried plan-audit debts (iter4).** S1-a: the implementation fails the gate with a non-zero exit and a message naming `gate.pre_commit.test_scope.mode` (`probe016.sh` AC-011 line: `GREEN (gate exit=1, message names the key=yes)`); REQ-PAS-011 / AC-PAS-011 still do not state the non-zero exit — spec-change request filed in the run report. S1-b: the run re-measured the RED-now cells on tree `00e7843ab` (RED-M3b above) and the GREEN cells on `9d114393c`; acceptance.md's tree-pin wording is a SPEC-body matter (request filed). S2: AC-PAS-003's Given does not say the new packages carry test files — the test graph gives them test files; wording request filed. S3: the placeholder `<moai built from the tree under measurement>` is replaced in this run by the named build `/private/tmp/t1673-bin/moai` (commit 9d114393c, judging build cited next to the tree HEAD). S4: AC-PAS-001's real-binary probe cannot tell transitive from direct dependents (its module has no package importing `b`); disposed by the unit test (a→b→c chain, transitive mutant red) and by package `d` (imports `b`) added to the AC-PAS-016 end-to-end module, whose recorded argv stays exactly `test ./a ./b`.


### E1 — AC PASS/FAIL matrix (final tree HEAD `9d114393cac23a3a44dc96261ce7936cd36f11cc`; every command run in this run, this tree, TMPDIR=GOTMPDIR=/private/tmp/t1673-tmp)
Final-tree re-runs: `go test ./internal/hook/quality -count=1` exit 0 (`ok ... 23.342s`, whole package); `go test ./internal/cli -run '^(TestRunGate|TestGAR_|TestGateCmd_|TestPrecommitE2E|TestUpdateGateYAML|TestPreCommitHook_|TestPreCommitRelocation_|TestPreCommitAffectedScopeE2E)' -count=1 -v` exit 0, 41 PASS lines, no FAIL (`ok ... 43.639s`); `go test ./internal/config` fails only the baseline `TestNoHardcodedContexts`.

| AC | Class | RED cell (verbatim above) | GREEN command (exit 0 unless noted) | Status |
|----|-------|---------------------------|-------------------------------------|--------|
| AC-PAS-001 | release-blocking | RED-M2; probe016 `AC-016/001` RED (RED-M3b) | `go test ./internal/hook/quality -run '^TestAffectedScope_DirectDependents$' -count=1 -v`; probe016 `AC-016/001 change a: GREEN` | PASS |
| AC-PAS-002 | adoption-pending | P1 ledger row (RED-M2) | `...-run '^TestAffectedScope_TestImportEdges$'` | PASS |
| AC-PAS-003 | adoption-pending | P1 ledger row (RED-M2) | `...-run '^TestAffectedScope_Ownership$'` and `'^TestAffectedScope_OwnershipTestdataGoFixtureAlone$'` | PASS |
| AC-PAS-004 | release-blocking | RED-M2; probe016 `AC-004` RED | `...-run '^TestAffectedScope_TestOnlyNarrows$'`; probe016 `AC-004 test-only change: GREEN` | PASS |
| AC-PAS-005 | release-blocking | RED-M2 (core); RED-M3a (adapter) | `...-run '^TestAffectedScope_DeletedAndRenamed$'`; `...-run '^TestAffectedScope_RealGitRenameDelete$'` | PASS |
| AC-PAS-006 | release-blocking | RED-M2 (core); RED-M3b (gate); probe016 `AC-006` RED | `...-run '^TestAffectedScope_DropAndEmpty$'`; `...-run '^TestQualityGate_AffectedScope_EmptySelectionSkips$'`; probe016 `AC-006 docs-only change: GREEN` | PASS |
| AC-PAS-007 | regression-guard | n/a (green by construction; mutant probe red) | `...-run '^TestAffectedScope_WideningTriggers$'` | PASS |
| AC-PAS-008 | regression-guard | n/a (mutant probe red) | `...-run '^TestQualityGate_AffectedScope_QueryFailureFallsBack$'`; `...-run '^TestAffectedScope_IncompleteGraphFallsBack$'` | PASS |
| AC-PAS-009 | release-blocking | RED-M3a; LEDGER-PAS-009-A (probe009, immutable witness, exit 1 re-run above) | `...-run '^TestAffectedScope_ChangedSetMatchesCommit$' -count=1 -v` (plain repository and linked worktree, shapes i-v) | PASS |
| AC-PAS-010 | regression-guard | n/a (passes against the stub) | `...-run '^TestQualityGate_AffectedScope_ContextIsolation$'` | PASS |
| AC-PAS-011 | release-blocking | RED-M1; probe016 `AC-011` RED; RED-M3b CLI | `go test ./internal/config -run '^TestGatePreCommitTestScope$' -count=1 -v`; the two `TestUpdateGateYAML*` tests (in the 41 PASS); `go test ./internal/cli -run '^TestRunGatePreCommitInvalidScopeModeRejected$'`; probe016 `AC-011 mode: bogus: GREEN (gate exit=1, message names the key=yes)` | PASS |
| AC-PAS-012 | withdrawn (Q1) | — | — | WITHDRAWN |
| AC-PAS-013 | release-blocking | RED-M5; LEDGER-PAS-013-A re-run (exit 1) | `...-run '^TestSuppressionWalkSkipsTestdata$' -count=1 -v`; `ledger013.sh /private/tmp/t1673-bin/moai` exit 0, no suppression line; whole-tree run exit 0, 0 suppression lines | PASS |
| AC-PAS-014 | regression-guard | n/a | `...-run '^TestQualityGate_StepBudgetsUnchanged$' -count=1 -v`; gate.yaml diff, added `t.Skip` = 0, removed `func Test` = 0 (M6 scans) | PASS |
| AC-PAS-015 | regression-guard | n/a | diff scans in M6: added `hooksPath` 0, added git-config writes 0, hook files unmodified | PASS |
| AC-PAS-016 | release-blocking | RED-M3b (probe016 RED; e2e RED) | `go test ./internal/cli -run '^TestPreCommitAffectedScopeE2E$' -count=1 -v`; `probe016.sh /private/tmp/t1673-bin/moai` exit 0, four GREEN lines | PASS |

Live AC count 15 (001-011 less 012, 013-016): 15 PASS, 0 FAIL.


### Sync-audit round 1 repair (F1-F3; audit report `.moai/reports/t1673/sync-audit.md` at 183c4b18d, FAIL 0.74)
RED commit: the failing regression tests plus the `UnstagedGo` field stub (parent of the GREEN commit); GREEN commit `73994a41ad7bbab9a2b9f42afe11a9697ac3ee02`. Each pair below is one command, run on the RED tree and again on the GREEN tree, TMPDIR=GOTMPDIR=/private/tmp/t1673-tmp.

**F1 — diff.relative=true in a nested module.** Command: `go test ./internal/hook/quality -run '^TestAffectedScope_DiffRelativeConfigIgnored$' -count=1 -v`
RED (exit 1):
```
=== RUN   TestAffectedScope_DiffRelativeConfigIgnored
    affected_scope_repair_test.go:41: files = [a/a.go], prefix = "svc", want [svc/a/a.go] and "svc" (diff.relative=true must not make names cwd-relative)
--- FAIL: TestAffectedScope_DiffRelativeConfigIgnored (0.46s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality	1.100s
FAIL
```
GREEN (exit 0):
```
=== RUN   TestAffectedScope_DiffRelativeConfigIgnored
--- PASS: TestAffectedScope_DiffRelativeConfigIgnored (0.50s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	1.269s
```
Fix: both git queries run as `git -c diff.relative=false diff ...`. The `-c` form was chosen over `--no-relative` because `-c` works on every git, while `--no-relative` needs git 2.28; verified against the installed git 2.54.0 by the test above (the test sets `diff.relative true` in the repository and the RED output is the cwd-relative name `a/a.go`).

**F2 — staged change to a, unstaged edit removing b's import of a.** Command: `go test ./internal/hook/quality -run '^TestAffectedScope_UnstagedGoEditFallsBackToFullModule$' -count=1 -v`
RED (exit 1):
```
=== RUN   TestAffectedScope_UnstagedGoEditFallsBackToFullModule
=== RUN   TestAffectedScope_UnstagedGoEditFallsBackToFullModule/a_clean_working_tree_keeps_the_narrow_set
=== RUN   TestAffectedScope_UnstagedGoEditFallsBackToFullModule/an_unstaged_edit_removing_an_import_widens_to_the_full_module
    affected_scope_repair_test.go:141: go test calls = [test ./a], want [test ./...] (an unstaged Go edit makes the graph unreliable)
        affected-scope: selected 1, dropped 0 (running 1)
        
        quality gate steps (3 configured):
          - go vet: disabled — turned off by the gate.disabled_steps setting
          - golangci-lint: skipped — none of its config files exist in the project directory (.golangci.yml, .golangci.yaml, .golangci.toml, .golangci.json)
          - go test: executed in 8ms at 2026-10-18T12:28:15.778+09:00 — go test ./a
    affected_scope_repair_test.go:144: the summary does not name the reason and the fallback:
        affected-scope: selected 1, dropped 0 (running 1)
        
        quality gate steps (3 configured):
          - go vet: disabled — turned off by the gate.disabled_steps setting
          - golangci-lint: skipped — none of its config files exist in the project directory (.golangci.yml, .golangci.yaml, .golangci.toml, .golangci.json)
          - go test: executed in 8ms at 2026-10-18T12:28:15.778+09:00 — go test ./a
=== RUN   TestAffectedScope_UnstagedGoEditFallsBackToFullModule/an_unstaged_non-Go_edit_does_not_widen
--- FAIL: TestAffectedScope_UnstagedGoEditFallsBackToFullModule (2.60s)
    --- PASS: TestAffectedScope_UnstagedGoEditFallsBackToFullModule/a_clean_working_tree_keeps_the_narrow_set (0.84s)
    --- FAIL: TestAffectedScope_UnstagedGoEditFallsBackToFullModule/an_unstaged_edit_removing_an_import_widens_to_the_full_module (0.94s)
    --- PASS: TestAffectedScope_UnstagedGoEditFallsBackToFullModule/an_unstaged_non-Go_edit_does_not_widen (0.82s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality	3.077s
FAIL
```
GREEN (exit 0):
```
=== RUN   TestAffectedScope_UnstagedGoEditFallsBackToFullModule
=== RUN   TestAffectedScope_UnstagedGoEditFallsBackToFullModule/a_clean_working_tree_keeps_the_narrow_set
=== RUN   TestAffectedScope_UnstagedGoEditFallsBackToFullModule/an_unstaged_edit_removing_an_import_widens_to_the_full_module
=== RUN   TestAffectedScope_UnstagedGoEditFallsBackToFullModule/an_unstaged_non-Go_edit_does_not_widen
--- PASS: TestAffectedScope_UnstagedGoEditFallsBackToFullModule (2.50s)
    --- PASS: TestAffectedScope_UnstagedGoEditFallsBackToFullModule/a_clean_working_tree_keeps_the_narrow_set (0.72s)
    --- PASS: TestAffectedScope_UnstagedGoEditFallsBackToFullModule/an_unstaged_edit_removing_an_import_widens_to_the_full_module (0.98s)
    --- PASS: TestAffectedScope_UnstagedGoEditFallsBackToFullModule/an_unstaged_non-Go_edit_does_not_widen (0.80s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	2.969s
```
Fix: `queryCommitChanges` also runs `git -c diff.relative=false diff --name-only --no-renames -z` (working tree against the index the commit records, the hook-provided index included) and keeps the `.go`, `go.mod`, `go.sum`, `go.work`, `go.work.sum` names in `UnstagedGo`; `scopeGoTestStep` runs the full module with the notice `affected-scope: working tree differs from the index in <files>; ... falling back to the full module (go test ./...)` when the list is non-empty. A query failure is the existing changed-file-query fallback. Under `git commit -a` the hook's temporary index already holds the tracked edits, so the common `-a` shape stays narrow; under a path-limited commit any other modified tracked Go file widens the scope (conservative). The clean-tree subtest and the unstaged-non-Go subtest keep `test ./a ./b`. `TestAffectedScope_ChangedSetMatchesCommit` (hook, all five shapes, plain and linked worktree) still passes with the new query (exit 0, `ok ... 4.317s`), so the extra git call raises no error against the temporary indexes.

**F3 — file embedded by two packages.** Command: `go test ./internal/hook/quality -run '^TestAffectedScope_SharedEmbedFileHasAllOwners$' -count=1 -v`
RED (exit 1):
```
=== RUN   TestAffectedScope_SharedEmbedFileHasAllOwners
    affected_scope_repair_test.go:65: selected = [example.com/m/assets example.com/m/x], want exactly [example.com/m example.com/m/assets example.com/m/x]
    affected_scope_repair_test.go:67: iteration 0
--- FAIL: TestAffectedScope_SharedEmbedFileHasAllOwners (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality	0.476s
FAIL
```
GREEN (exit 0), 50 iterations inside the test:
```
=== RUN   TestAffectedScope_SharedEmbedFileHasAllOwners
--- PASS: TestAffectedScope_SharedEmbedFileHasAllOwners (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	0.460s
```
Fix: `scopeOwners.embedded` maps a file to the sorted set of its owners; `owner` returns a slice and `selectAffected` marks every owner touched and, for source changes, source.

**Re-runs after the fixes (this run, this tree, HEAD 73994a41ad7bbab9a2b9f42afe11a9697ac3ee02 after the GREEN commit):**
- `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/hook/quality/ ./internal/config/ ./internal/cli/` exit 0; `gofmt -l internal/hook/quality internal/cli internal/config` printed nothing; `golangci-lint run --timeout=3m ./internal/hook/quality/...` `0 issues.`
- `go test -race -count=1 ./internal/hook/quality/` exit 0: `ok  	github.com/modu-ai/moai-adk/internal/hook/quality	64.672s` (no leased slot; the package ran in the time quoted).
- `go test -count=1 ./internal/cli -run 'TestUpdateGateYAML|TestRunGate|TestPreCommit|TestPrecommit|TestGateYAML|TestGateConfig|TestGatePreCommit|TestGAR_|TestGateCmd_' -v` exit 0, 72 top-level PASS, no FAIL: `ok  	github.com/modu-ai/moai-adk/internal/cli	67.872s`. The full internal/cli suite was not run (ruling 35e9).
- probe016.sh against a build of the GREEN commit (`/private/tmp/t1673-bin/moai-repair`, `moai version` prints `t1673-repair   73994a41a`), exit 0:
```
AC-016/001 change a: GREEN  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 1), wanted go argv 'test ./a ./b', got 'test ./a ./b')
AC-004 test-only change: GREEN  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 1), wanted go argv 'test ./a', got 'test ./a')
AC-006 docs-only change: GREEN  (commit exit=0, HEAD moved=yes, go test calls=0 (wanted 0), wanted go argv 'none', got 'none')
AC-011 mode: bogus: GREEN  (gate exit=1, message names the key=yes)
exit=0
```
(One earlier attempt printed four RED lines with `gate exit=127`: the build command had been refused by the worktree guard, so the binary did not exist; rebuilt through a script and re-run, result above.)


### Sync-audit round 2 repair (R2-1; audit report `.moai/reports/t1673/sync-audit-iter2.md` at d9f5b3e6e, FAIL 0.84)
RED commit: the failing tests plus the optional second output of the hook helper; GREEN commit `6e38795ff76f2b78dfdb2ed8379961e497553636`. TMPDIR=GOTMPDIR=/private/tmp/t1673-tmp.

**Decision, judged from the code.** The working-tree-versus-index check now also lists untracked `go.mod`, `go.work` and `go.work.sum`. Not listed: `go.sum` (checksums only — it adds no package and no import edge, so it cannot change the selection; the negative control keeps the narrow set) and a plain untracked `.go` file (it can only add an edge or a package to the working-tree graph, which over-selects; a conflicting package clause makes `go list` mark the package errored, which is the existing graph fallback; it cannot remove a package). An untracked `go.mod` turns its directory into a separate module and so removes the package from the graph; `go.work`/`go.work.sum` change the module set. `--exclude-standard` honours .gitignore, so an ignored nested go.mod is not seen (residual risk).

**R2-1 RED — command:** `go test ./internal/hook/quality -run '^TestAffectedScope_UntrackedModuleFileFallsBackToFullModule$' -count=1 -v`, exit 1 (summary blocks omitted):
```
=== RUN   TestAffectedScope_UntrackedModuleFileFallsBackToFullModule
=== RUN   TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_nested_go.mod_widens_to_the_full_module
    affected_scope_untracked_test.go:61: go test calls = [test ./a], want [test ./...]
    affected_scope_untracked_test.go:64: the summary does not name the untracked file and the fallback:
=== RUN   TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_go.work_widens_to_the_full_module
    affected_scope_untracked_test.go:61: go test calls = [test ./a ./b], want [test ./...]
    affected_scope_untracked_test.go:64: the summary does not name the untracked file and the fallback:
=== RUN   TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_go.work.sum_widens_to_the_full_module
    affected_scope_untracked_test.go:61: go test calls = [test ./a ./b], want [test ./...]
    affected_scope_untracked_test.go:64: the summary does not name the untracked file and the fallback:
=== RUN   TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_text_file_keeps_the_narrow_set
=== RUN   TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_plain_Go_file_keeps_the_narrow_set
=== RUN   TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_go.sum_keeps_the_narrow_set
--- FAIL: TestAffectedScope_UntrackedModuleFileFallsBackToFullModule (4.75s)
    --- FAIL: TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_nested_go.mod_widens_to_the_full_module (0.70s)
    --- FAIL: TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_go.work_widens_to_the_full_module (0.92s)
    --- FAIL: TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_go.work.sum_widens_to_the_full_module (0.78s)
    --- PASS: TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_text_file_keeps_the_narrow_set (0.66s)
    --- PASS: TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_plain_Go_file_keeps_the_narrow_set (0.95s)
    --- PASS: TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_go.sum_keeps_the_narrow_set (0.74s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality	5.445s
FAIL
```
**R2-1 GREEN (exit 0):**
```
--- PASS: TestAffectedScope_UntrackedModuleFileFallsBackToFullModule (10.55s)
    --- PASS: TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_nested_go.mod_widens_to_the_full_module (5.52s)
    --- PASS: TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_go.work_widens_to_the_full_module (1.57s)
    --- PASS: TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_go.work.sum_widens_to_the_full_module (0.79s)
    --- PASS: TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_text_file_keeps_the_narrow_set (0.93s)
    --- PASS: TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_plain_Go_file_keeps_the_narrow_set (0.99s)
    --- PASS: TestAffectedScope_UntrackedModuleFileFallsBackToFullModule/an_untracked_go.sum_keeps_the_narrow_set (0.76s)
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	11.211s
```

**Hook / linked-worktree RED — command:** `go test ./internal/hook/quality -run '^TestAffectedScope_UntrackedModuleFileThroughHook$' -count=1 -v`, exit 1:
```
=== RUN   TestAffectedScope_UntrackedModuleFileThroughHook
=== RUN   TestAffectedScope_UntrackedModuleFileThroughHook/plain_repository_under_the_temporary_index_of_commit_-a
    affected_scope_untracked_test.go:111: working-tree-versus-index list = "", want "n/go.mod" (the untracked go.mod, and no other file)
=== RUN   TestAffectedScope_UntrackedModuleFileThroughHook/linked_worktree
    affected_scope_untracked_test.go:120: working-tree-versus-index list = "", want "n/go.mod" (the untracked go.mod, and no other file)
--- FAIL: TestAffectedScope_UntrackedModuleFileThroughHook (1.70s)
    --- FAIL: TestAffectedScope_UntrackedModuleFileThroughHook/plain_repository_under_the_temporary_index_of_commit_-a (0.87s)
    --- FAIL: TestAffectedScope_UntrackedModuleFileThroughHook/linked_worktree (0.84s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality	2.248s
FAIL
```
**GREEN (exit 0):**
```
=== RUN   TestAffectedScope_UntrackedModuleFileThroughHook
=== RUN   TestAffectedScope_UntrackedModuleFileThroughHook/plain_repository_under_the_temporary_index_of_commit_-a
=== RUN   TestAffectedScope_UntrackedModuleFileThroughHook/linked_worktree
--- PASS: TestAffectedScope_UntrackedModuleFileThroughHook (1.95s)
    --- PASS: TestAffectedScope_UntrackedModuleFileThroughHook/plain_repository_under_the_temporary_index_of_commit_-a (0.82s)
    --- PASS: TestAffectedScope_UntrackedModuleFileThroughHook/linked_worktree (1.13s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	2.656s
```

**Nested-module query (new, also pins `-c diff.relative=false` on the unstaged query, the surviving mutant the auditor noted) — command:** `go test ./internal/hook/quality -run '^TestAffectedScope_WorkingTreeQueriesFromNestedModule$' -count=1 -v`, exit 0 on GREEN:
```
=== RUN   TestAffectedScope_WorkingTreeQueriesFromNestedModule
--- PASS: TestAffectedScope_WorkingTreeQueriesFromNestedModule (0.49s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	1.176s
```
Mutants (temporary edits, reverted; `grep -n MUTANT` over the file prints nothing): dropping `-c diff.relative=false` from the unstaged query → `UnstagedGo = [a/b.go go.work], want [go.work svc/a/b.go]`; dropping the `top` pathspec magic → `UnstagedGo = [svc/a/b.go], want [go.work svc/a/b.go]`. Reverting the whole fix is the RED state above (the committed RED tests fail on the pre-fix tree: the nested go.mod case selects `[test ./a]`).

**Re-runs after the fix (HEAD `6e38795ff76f2b78dfdb2ed8379961e497553636`):**
- `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/hook/quality/ ./internal/config/ ./internal/cli/` exit 0; `gofmt -l internal/hook/quality internal/cli internal/config` no output; `golangci-lint run --timeout=3m ./internal/hook/quality/...` `0 issues.`
- `go test -count=1 -race ./internal/hook/quality/` exit 0: `ok  	github.com/modu-ai/moai-adk/internal/hook/quality	67.256s` (no leased slot).
- `go test -count=1 ./internal/cli -run 'TestUpdateGateYAML|TestRunGate|TestPreCommit|TestPrecommit|TestGateYAML|TestGateConfig|TestGatePreCommit|TestGAR_|TestGateCmd_' -v` exit 0, 72 top-level PASS, no FAIL: `ok  	github.com/modu-ai/moai-adk/internal/cli	74.773s`. Full internal/cli suite not run (ruling 35e9).
- probe016.sh against `/private/tmp/t1673-bin/moai-r2` (`moai version`: `t1673-r2   6e38795ff`), exit 0:
```
AC-016/001 change a: GREEN  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 1), wanted go argv 'test ./a ./b', got 'test ./a ./b')
AC-004 test-only change: GREEN  (commit exit=0, HEAD moved=yes, go test calls=1 (wanted 1), wanted go argv 'test ./a', got 'test ./a')
AC-006 docs-only change: GREEN  (commit exit=0, HEAD moved=yes, go test calls=0 (wanted 0), wanted go argv 'none', got 'none')
AC-011 mode: bogus: GREEN  (gate exit=1, message names the key=yes)
exit=0
```

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-18
run_commit_sha: 9d114393cac23a3a44dc96261ce7936cd36f11cc   # last code commit; the evidence-only commit(s) that record §E.2/§E.3 follow it
run_status: PASS-WITH-DEBT
audit_ready: true
ac_pass_count: 15          # live ACs; AC-PAS-012 withdrawn (Q1)
ac_fail_count: 0
preserve_list_post_run_count: 0   # no file outside the SPEC's declared surface was modified (24 files, all under internal/hook/quality, internal/config, internal/cli/gate*.go + two tests, the two gate.yaml copies, and this SPEC's spec.md/progress.md)
l44_pre_commit_fetch: not-run (local-only card: no push, no pre-spawn fetch dispatched in this run)
l44_post_push_fetch: not-run (no push)
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues on internal/hook/quality + internal/config; internal/cli shows 3 baseline errcheck in an untouched file; spec lint 0 errors / 0 warnings
cross_platform_build:
  darwin_arm64: "go build ./... exit 0"
  windows_amd64: "GOOS=windows GOARCH=amd64 go build ./... exit 0; GOOS=windows go vet ./internal/hook/quality/ ./internal/config/ ./internal/cli/ exit 0"
total_run_phase_files: 24
m1_to_mN_commit_strategy: "RED commit before GREEN commit per milestone (M1, M2, M3a, M3b, M5), M4 withdrawn, M6 = test-hardening commit + evidence commit(s); no push"
slot: "never held (acquire exit 3, lane-25 held go-test-heavy); nothing to release"
debts:
  - "Q4 (project-wide vet/lint/typecheck runtime) unmeasured; gates only the post-landing runbook"
  - "no full-package -race run of internal/hook/quality or internal/config (slot unavailable); scoped -race runs PASS"
  - "go test ./internal/template (440 s) ran during M1 without a slot lease"
  - "internal/cli package coverage not measured (ruling 35e9); runGate 92.9% from a targeted selection"
  - "baseline failure internal/config TestNoHardcodedContexts (scripts/release.sh), not attributable to this card"
  - "SPEC-body wording requests for manager-spec (S1-a, S1-b, S2, AC-PAS-003 mutant weakness) are in the run report, not edited here"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-18
sync_commit_sha: pending-backfill   # a commit cannot cite its own hash; backfilled in a following commit
sync_status: completed-with-open-debt
audit_ready: true
audit_ready_reason: "Was false at the first sync commit because B12 self-test (b) prints 20 against the stated 15 live. Sync audit round 1 (.moai/reports/t1673/sync-audit.md) ruled that artifact real but low severity and non-blocking: the CHANGELOG '15 live' is accurate. Corrected diagnosis from that ruling: 20 = 16 distinct AC-PAS- identifiers (AC-PAS-012 is unmarked everywhere, so the counter reads it live) + the four shorthand tokens AC-016, AC-004, AC-006, AC-011 (acceptance.md lines 146-149); marking only the four would leave 16, so the complete fix is the moai-ac-prefix declaration plus [RETIRED] on AC-PAS-012. It stays open manager-spec debt (SPEC-body edit on a completed SPEC needs the amendment transition). Every other check this agent owns passed (go build, lint 0/0, CHANGELOG and docs parity)."
sync_audit_round1: "FAIL 0.74 at 183c4b18d on three under-selection defects F1-F3; repaired by manager-develop (RED 835df28b2, GREEN 73994a41a, evidence 1309c60ed). Delta re-audit pending."
sync_docs_delta: "CHANGELOG entry (edited in place) and docs-site moai-gate.md in ko/en/ja/zh now state the F2 behaviour: a .go, go.mod, go.sum, go.work or go.work.sum file that differs between the working tree and the index runs the full module with a notice naming the files (code: affected_scope_queries.go UnstagedGo = files from git diff --name-only, filtered to .go or the widen base names; affected_scope_gate.go falls back before the graph query). F1 (diff.relative=false on both git queries) and F3 (a file embedded by several packages is owned by all of them) are internal corrections and need no operator-facing sentence."
open_debt_f4_f6_for_manager_spec:
  - "F4: B12 counter artifact above (acceptance.md prefix declaration + [RETIRED] on AC-PAS-012)"
  - "F5: S1-a, S1-b, S2 wording in acceptance.md / REQ-PAS-011 / AC-PAS-011 / AC-PAS-003 Given, plus the tree-pin wording"
  - "F6: internal/config/gate_test_scope.go accepts an explicit mode: \"\" and resolves it to affected, while REQ-PAS-011 says any value other than affected or full is rejected when the key is present; harmless in effect"
sync_audit_round2: "R2-1 repaired by manager-develop (RED e5adeac70, GREEN 6e38795ff, evidence fc96f7218): an untracked go.mod, go.work or go.work.sum now widens the Go test step to the full module (an untracked go.sum, plain .go file or other file does not); CHANGELOG and the four docs-site pages updated in place. Delta re-audit pending."
sync_head_at_start: aaf01ee7e0c27bff958f8db082e4db3c40ce8f39
b12_self_test_a: "PASS - grep -c 'SPEC-PRECOMMIT-AFFECTED-SCOPE-001' CHANGELOG.md printed 0 before emission"
b12_self_test_b: "FAIL (counter 20 vs stated 15) - ac_source=.moai/specs/SPEC-PRECOMMIT-AFFECTED-SCOPE-001/acceptance.md, tier=M, state=resolved; counter stdout 20, stderr live=20 excluded=0 ambiguous=0; declared criteria AC-PAS-001..016 = 16 identifiers, AC-PAS-012 withdrawn => 15 live; CHANGELOG states 15"
b12_self_test_c: "PASS - every path cited in the CHANGELOG entry resolved with ls before commit"
changelog_entry_position: "CHANGELOG.md [Unreleased] / Added, first bullet"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed on this single sync commit (status + updated only; updated already 2026-10-18)"
  plan_md_acceptance_md: "no frontmatter block; nothing to refresh"
canary_compliance_check: "not applicable - this SPEC defines no forward-looking policy that its own sync tests"
docs_site_touched: "docs-site/content/{ko,en,ja,zh}/utility-commands/moai-gate.md - test_scope.mode note, same commit, ko canonical"
mx_validation: "3 existing @MX:NOTE tags (selectAffected, commitQueryEnv, scopeGoTestStep) each with @MX:SPEC; no tag added, removed or changed; no ANCHOR needed (no function with fan_in >= 3 in the new files was found by reading the call sites); no goroutines"
open_debt_for_manager_spec:
  - "S1-a: REQ-PAS-011 / AC-PAS-011 do not state the non-zero exit of the rejected mode value (spec-change request filed in the run report)"
  - "S1-b: acceptance.md tree-pin wording for the RED-now cells (RED re-measured on 00e7843ab, GREEN on 9d114393c)"
  - "S2: AC-PAS-003 Given does not say the new packages carry test files"
  - "AC-PAS-003 mutant weakness"
  - "acceptance.md lines 146-149: four shorthand AC tokens need [REF] marks (clears B12 self-test b)"
disposed_in_run_not_open: "S3 (placeholder replaced by the named build /private/tmp/t1673-bin/moai) and S4 (unit test plus package d in the AC-PAS-016 module) are recorded as disposed in the run-phase evidence, so they are not open here; the sync request listed them as open and this record follows the run evidence"
carried_run_debts:
  - "Q4 (project-wide vet/lint/typecheck runtime) unmeasured; gates only the post-landing hook re-enable runbook"
  - "no full-package -race run of internal/hook/quality or internal/config"
  - "internal/cli package coverage not measured (ruling 35e9)"
  - "baseline failure internal/config TestNoHardcodedContexts (scripts/release.sh), not attributable to this card"
  - "go-test-heavy slot never held (lane-25 held it); no heavy suite run at sync"
```

## §G Override and Refusal Record

- 2026-10-18T01:41:03Z SPEC-PRECOMMIT-AFFECTED-SCOPE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter4.md
- 2026-10-18T01:41:12Z SPEC-PRECOMMIT-AFFECTED-SCOPE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter4.md
- 2026-10-18T01:42:39Z SPEC-PRECOMMIT-AFFECTED-SCOPE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1673/.moai/reports/t1673/plan-audit-iter4.md

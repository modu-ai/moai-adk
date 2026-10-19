---
id: SPEC-TRUNK-BACKUP-SYNC-001
title: "Acceptance criteria — trunk backup sync verb + workflow-exclusion regression guard"
version: "0.3.3"
created: 2026-10-17
updated: 2026-10-18
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, .github/workflows, .moai/docs"
lifecycle: spec-anchored
tags: "trunk-backup, sync-ref, git-push, workflow-triggers, regression-guard"
---

# acceptance.md — SPEC-TRUNK-BACKUP-SYNC-001

> Verification layer. Every entry is Given-When-Then and binary-testable. GEARS obligations live in `spec.md` §D; they are not restated here. Stateless on the status axis.

## §A — Traceability matrix

| AC | Requirement | Milestone | Severity | Class |
|---|---|---|---|---|
| AC-TBS-001 | REQ-TBS-005 | M1 | Must | regression-guard (live sweep) |
| AC-TBS-002 | REQ-TBS-005 | M1 | Must | release-blocking (mutant RED) |
| AC-TBS-003 | REQ-TBS-001 | M2 | Must | release-blocking |
| AC-TBS-004 | REQ-TBS-001, REQ-TBS-002 | M2 | Must | release-blocking |
| AC-TBS-005 | REQ-TBS-003 | M2 | Must | release-blocking |
| AC-TBS-006 | REQ-TBS-006, REQ-TBS-008 | M2/M3 | Must | release-blocking |
| AC-TBS-007 | REQ-TBS-001 | M2 | Must | release-blocking |
| AC-TBS-008 | REQ-TBS-007, REQ-TBS-009 | M3 | Must | release-blocking |
| AC-TBS-009 | REQ-TBS-004 | M2 | Must | release-blocking |
| AC-TBS-010 | REQ-TBS-005 | M1 | Must | release-blocking |

## §B — Acceptance criteria

- **AC-TBS-001 — live workflow sweep**. Given the repository's `.github/workflows/` at the tested commit, when the regression test enumerates every `*.yml`/`*.yaml` file and evaluates each `on.push` trigger's branch filter against the ref name `sync/main`, then every push trigger excludes `sync/main`, and the test asserts it found at least one workflow file carrying a push trigger (empty-sweep guard: a zero-count sweep fails the test, never passes it).
- **AC-TBS-002 — mutant detection (observed-failure evidence)**. Given the in-test mutant fixtures — a bare `on: push` block with no branch filter, a `branches: ['sync/**']` list, and the bidirectional set of REQ-TBS-005 (`['**']` admits `sync/main`; `['*']` does not; `['**', '!sync/main']` does not, last-match negation winning; `['sync/m[a-z]in']` does, character-class support) — when the same matcher evaluates them, then each fixture's admission verdict matches that stated direction. These cases are the guard's red-on-known-failing-input record in both directions (verification-completeness §1.1, §2 mutant probe); a matcher that passes the live sweep but misses a mutant, or one that flags `['*']` as admitting, fails this criterion.
- **AC-TBS-003 — first backup push against a temp remote**. Given a temporary repository whose `origin` is a bare repository under `t.TempDir()`, with `refs/heads/main` at commit SHA `X` and no `sync/main` on the remote, when the verb runs, then the remote gains `refs/heads/sync/main` at `X`, no other remote ref changes, and the verb's output names `X` and the target ref.
- **AC-TBS-004 — force scoping, the non-FF case, and no tag leak**. Given the temp remote's `refs/heads/sync/main` at a commit that is not an ancestor of local `main` (the rewound-main shape), when the verb runs, then the push succeeds and updates `sync/main`; given the test's captured record of every git command the verb issued across all runs, then `--force` (or `-f`) appears only on the exact pinned refspec form `<sha>:refs/heads/sync/main` (the source side is the full SHA the verb resolved and pinned at invocation start — never the bare branch refspec `refs/heads/main:refs/heads/sync/main`), and no captured command updates any other remote ref; and given a temp clone whose gitconfig sets `push.followTags=true` with an annotated `v*` tag reachable from `main`, when the verb runs, then the remote ref set changes only at `refs/heads/sync/main` — no tag ref is created (the `--no-follow-tags` constant holds).
- **AC-TBS-005 — refusal on missing main**. Given a temporary repository in which `refs/heads/main` does not resolve, when the verb runs, then it exits non-zero, prints guidance naming the missing ref and the constant target, and its captured command record shows no push command was executed.
- **AC-TBS-006 — no-real-push and no-workflow-edit guard**. Given the test suite's captured verb commands and the run-phase evidence, when they are inspected, then every remote operand resolves under a `t.TempDir()` root, no test names the development repository's configured remotes or any network URL, and the run phase as a whole contacted no real remote and opened no pull request; and given the run-phase baseline SHA recorded before M1, when the run phase's diff against that baseline is read with the pathspec `.github/workflows/`, then the diff is empty — no workflow file was modified (the second prohibition of REQ-TBS-008; `.github/test-input-filters.yml` is outside it, living under `.github/` per plan.md §A.5).
- **AC-TBS-007 — environment scrub (t1478 shape)**. Given a process environment carrying inherited `GIT_DIR`/`GIT_WORK_TREE` pointing at a foreign temporary repository, when the verb runs, then the push targets the caller's repository (the temp clone the test constructed), proving repository-location variables were removed before exec.
- **AC-TBS-008 — runbook and template isolation**. Given the card's final state, then `.moai/docs/trunk-backup-runbook.md` exists and carries the four REQ-TBS-007 statements (leader-round step; operator-approved origin-push scope; redundancy-copy-only status; unchanged batch discipline and main protection), and no file under `internal/template/templates/` and no branch-protection or guard configuration changed.
- **AC-TBS-009 — push-failure reporting**. Given a temporary remote configured to reject the ref update (a failing pre-receive hook), when the verb runs, then it exits non-zero, prints both the stdout and the stderr of the failed push command, and its captured command record shows no git command was issued after the failed push.
- **AC-TBS-010 — guard fires on workflow-only diffs**. Given `.github/test-input-filters.yml` after M1, when its `go_code` path filter is read, then it maps `.github/workflows/**`, so a diff touching only workflow files (a new file, or an edit to any of the currently unmapped 18) no longer classifies as not-Go and the Go suite — with it the REQ-TBS-005 guard — is scheduled. The mapping is verifiable by parsing the file's YAML and asserting that the `go_code` sequence contains `.github/workflows/**` as a parsed list entry — never by text search, which a comment mentioning the glob also satisfies (in-tree parsing precedent: `mappingSequenceStrings`, `internal/template/dotgithub_template_test.go:322`). RED-now cell E-TBS-R5: at the audited tree the filter maps only three named workflow files.

## §C — Edge cases (covered by the tests or explicitly excluded)

- `release.yml` carries `tags: ['v*']` with no `branches` filter — the matcher must classify it as not firing on branch pushes (AC-TBS-001 covers via the live sweep; a `tags`-only fixture pins the semantics).
- A workflow file with `on: [push, pull_request]` (list form) — the parser must read the push trigger from the list form.
- A push trigger with `branches-ignore` only — under GitHub semantics it fires on every branch not ignored; the matcher must treat `sync/main` as admitted unless an ignore pattern matches it. (No current workflow uses this form; a mutant fixture pins the semantics.)
- Local `main` is a symbolic or ambiguous ref — REQ-TBS-003's refusal path.
- Verb invoked from a card worktree rather than the primary checkout — REQ-TBS-001 resolves the shared repository; the temp-remote tests cover the shape.

## §D — Quality gate criteria

- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `go test` for the new test files passes with the scoped run recorded verbatim.
- `golangci-lint` reports no NEW findings on the added files (baseline split recorded).
- `moai spec lint SPEC-TRUNK-BACKUP-SYNC-001` reports no error-severity findings on the SPEC artifacts.

## §E — Adoption classification (two-cell discipline)

- **Regression-guard (undecidable-disposition class)**: AC-TBS-001's live sweep — the current tree passes it by design; it guards a currently-true invariant. Its failure evidence is AC-TBS-002's mutant fixtures, which are executable now and at every future run. The pair (001, 002) is the two-cell adoption for the guard.
- **Red-at-arrival (release-blocking RED-now)**: AC-TBS-002..010 each carry a RED-now cell below, executed at plan phase and pinned to tree `8bff29c5f`. M2's first failing test runs remain the per-milestone RED capture per E8; the cells here are the starting observations that precede them.
- **Green path**: M1 flips AC-TBS-001/002/010; M2 flips AC-TBS-003..005, 007, and 009; M3 flips AC-TBS-006/008.

### RED-now evidence ledger (tree `8bff29c5f`, measured 2026-10-17)

Provenance: file-reading probes observed directly on the tree; the `moai`-invoking probe executed by the installed build `/Users/goos/go/bin/moai` (moai-adk v3.2.0-rc.36-t1822c, commit `ee981a932` — a strict ancestor of `8bff29c5f`; the verb's absence is a property of both the installed surface and the tree, which contains no `newIntegrationBackupCmd` anywhere). Two recording notes: (i) E-TBS-R1's stdout block is the raw command output embedded by file redirect (round-2 recapture, D13), with exactly two stated transformations — each line is indented four spaces for ledger layout, and the output's final newline is elided at the fence; trailing space runs and the leading blank line are preserved as bytes. R2-R5 blocks preserve line content byte-for-byte, with the terminal form's trailing space padding not reproduced inside markdown fencing. (ii) Divergence note (D1): the audit verdict's example predicted `Error: unknown command "backup" for "moai integration"`; the installed binary resolves an unknown subcommand to the parent group help instead — the recorded stdout is the observed output, not the predicted one. The R1 recapture was executed at tree `72135c81f` with the same judging build (`ee981a932`) and reproduces the round-1 measurement; per the round-2 verdict, the cell's command, exit code, and `8bff29c5f` pin are unchanged.

```text
E-TBS-R1 — the verb is absent from the shipped command surface
  command:   moai integration backup --help
  tree:      8bff29c5f (judging build ee981a932)
  exit code: 0
  stdout (verbatim):
    
      Hold and release the release-integration window.                                                                      
                                                                                                                            
      A lane announces its integration to the leader, then records the hold here so                                         
      the PreToolUse guard can refuse a second lane's `git merge` in the release                                            
      worktree. The record lives in the primary checkout's .moai/state, visible                                             
      from every linked worktree.                                                                                           
                                                                                                                            
      The deny layer is opt-in (workflow.integration_lock.enabled, default false);                                          
      these verbs work regardless, so a project may keep the record as a                                                    
      coordination signal without enabling refusal.                                                                         
             
      USAGE  
             
        moai integration [command] [--flags]  
                
      COMMANDS  
                
        status [--flags]                              Report who holds the release-integration window
        acquire [--flags]                             Record this session as the holder of the release-integration window
        release [--flags]                             Release the release-integration window this session holds
        policy open|hold --reason <text> [command]    Read or set the release-integration window policy (open | hold)
        remeasure -- <command>                        Run the re-measure and store its record keyed by the candidate tree
        merge --card <id> [--session <id>] [--flags]  Run the in-window merge step for a card (holder only; thirteen causes, one path)
        preflight [path] [--flags]                    Assert that a worktree has no modified tracked .claude/settings.json
        resync                                        Fast-Forward the primary checkout's local main to origin/main, inside the integration window
        candidate --card <id> [--observe] [--flags]   Build and push this card's pre-landing candidate commit to ci/<card> (the landing gate consumes its verdict); --observe records the CI run's verdict instead
             
      FLAGS  
             
        -h --help                                     Help for integration
    
  RED reading: the COMMANDS enumeration lists exactly nine subcommands — no `backup`. Every
  verb-behavior criterion is unmeetable at this tree.

E-TBS-R2 — no tree instrument references the protected ref
  command:   grep -rln "sync/main" internal/cli/
  tree:      8bff29c5f
  exit code: 1
  stdout:    (empty — zero matching files)
  RED reading: no matcher, mutant fixture, or test names `sync/main`; AC-TBS-002's detection
  capability does not exist.

E-TBS-R3 — the verb's source file is absent
  command:   ls internal/cli/integration_backup.go
  tree:      8bff29c5f
  exit code: 1
  stdout:    (empty)
  stderr:    ls: internal/cli/integration_backup.go: No such file or directory
  RED reading: neither the verb nor its isolation-guard tests exist; AC-TBS-006 unmeetable.

E-TBS-R4 — the runbook is absent
  command:   ls .moai/docs/trunk-backup-runbook.md
  tree:      8bff29c5f
  exit code: 1
  stdout:    (empty)
  stderr:    ls: .moai/docs/trunk-backup-runbook.md: No such file or directory
  RED reading: AC-TBS-008's runbook half unmeetable.

E-TBS-R5 — the go_code filter does not map workflow files
  command:   grep -n "workflows" .github/test-input-filters.yml
  tree:      8bff29c5f
  exit code: 0
  stdout (verbatim; line bytes re-confirmed with `cat -e` end-of-line markers):
    3:#   .github/workflows/ci.yml                  detect job         (test / race / skip-marker)
    4:#   .github/workflows/codeql.yml              detect job         (analyze)
    5:#   .github/workflows/release-pr-multi-os.yml detect-release job (full-matrix-test)
    54:  - '.github/workflows/ci.yml'
    55:  - '.github/workflows/codeql.yml'
    56:  - '.github/workflows/release-pr-multi-os.yml'
  RED reading: the go_code filter maps exactly three named workflow files (lines 54-56) — no
  `.github/workflows/**` glob; a workflow-only diff classifies as not-Go and skips the Go
  suite. AC-TBS-010 unmeetable.
```

Cell-to-criterion binding: AC-TBS-002 → E-TBS-R2; AC-TBS-003, AC-TBS-004, AC-TBS-005, AC-TBS-007, AC-TBS-009 → E-TBS-R1 (with E-TBS-R3 as the source-absence corroboration); AC-TBS-006 → E-TBS-R3 (with E-TBS-R2); AC-TBS-008 → E-TBS-R4; AC-TBS-010 → E-TBS-R5.

## §F — Definition of Done

- All ten criteria PASS with attributable evidence (command + verbatim output + tree SHA per run).
- No push to any real remote occurred at any point in the run phase (§ AC-TBS-006 evidence).
- `progress.md` §E.2 carries the run-phase evidence; §E.3 the audit-ready signal.
- The four decision-index rows remain untouched unless the operator records verdicts mid-run; a verdict landing mid-run is applied as the one-commit amendment described in `spec.md` §E.

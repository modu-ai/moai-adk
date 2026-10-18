---
id: SPEC-AGENT-CORE4-001
title: "Acceptance: piece 1"
version: "0.2.0"
created: 2026-10-18
---

# acceptance.md — SPEC-AGENT-CORE4-001 (piece 1)

Document-level pin: every RED-now cell was measured on tree `b4ea2d6c93` (clean, the audited HEAD;
code identical to `ae34f26831` per the audit). Test commands decide by exit code; a `-v` run is
green only when the listed `--- PASS:` lines are printed, because `ok ... [no tests to run]` exits 0
(empty sweep, `verification-completeness.md` §1.1). All test selectors are anchored.

## §1 Criteria

Each row: REQ, command, green output, per-milestone green column, RED-now ledger id.

| AC | REQ | Command (single invocation) | Green output | Green at | RED |
|---|---|---|---|---|---|
| AC-AC4-001 | 001 | `go test ./internal/template/agentemit -run '^TestNeutralSetIsOneToOne$' -v -count=1` | PASS lines `every_definition_has_a_source`, `every_source_has_a_definition`, `rendered_file_without_source_fails` | M1 (eight, then twelve at M2) | L-T1 |
| AC-AC4-002 | 002 | `go test ./internal/template/agentemit -run '^TestClaudeRenderUnmappedCapabilityFails$' -v -count=1` | PASS `unmapped_token_names_file_and_token`, `render_is_deterministic` | M1 | L-T2 |
| AC-AC4-003 | 003 | `go test ./internal/template/agentemit -run '^TestCodexRenderFromNeutral$' -v -count=1` | PASS `contract_rows_unchanged_for_twelve`, `claude_only_block_dropped_in_codex` | M1 | L-T3 |
| AC-AC4-004 | 004 | `go test ./internal/template/agentemit -run '^TestEightAgentsRenderPinnedBytes$' -v -count=1` | PASS one subtest per agent for `c2` and `c3` (16), `mutant_empty_body_fails`, `mutant_headings_only_fails` | M1 | L-T4 |
| AC-AC4-005 | 005 | `go test ./internal/template/agentemit -run '^TestCarriedFourAllowedDiffOnly$' -v -count=1` | PASS `changed_line_set_equals_pinned_list` (8 lines in C2: builder-harness 2, e2e-tester 2, manager-design 3, super-advisor 1), `no_install_root_prefix_in_neutral`, `mutant_extra_edit_fails` | M2 | L-T5 |
| AC-AC4-006 | 006 | `go test ./internal/template/agentemit -run '^TestCarriedFourCopiesEqual$' -v -count=1` | PASS per agent `c1_equals_c2`, `c3p_equals_c3` | M2 | L-T6, L-C1..L-C8 |
| AC-AC4-007 | 007 | `go test ./internal/template/agentemit -run '^TestEmitCheckSummaryAndHandEdit$' -v -count=1` | PASS `summary_line_format` (M1: `specs=12 copies=2 diff=0`; M2: `specs=12 copies=4 diff=0`, where copies counts the compared file sets), `hand_edit_fails_naming_file` with subtests `claude_template`, `codex_template` (M1) and `claude_project`, `codex_project` (M2, carried four only) | M1 (template subtests), M2 (project subtests) | L-T7, L-E |
| AC-AC4-008 | 008 | `git log --reverse --format=%s --grep='feat(SPEC-AGENT-CORE4-001): M'` | exactly three lines in order: `... M0 baseline ...`, `... M1 ...`, `... M2 ...` | M2 | L-G |
| AC-AC4-009 | 008 | `git log --name-only --format= --grep='feat(SPEC-AGENT-CORE4-001): M0 baseline'` | exactly `.moai/specs/SPEC-AGENT-CORE4-001/m0-baseline.md` and `internal/template/agentemit/testdata/agents-head.sha256` | M0 | L-G (empty) |

REQ coverage: 001-AC1; 002-AC2; 003-AC3; 004-AC4; 005-AC5; 006-AC6; 007-AC7; 008-AC8, AC9. Every AC maps
to one REQ; 8 REQ, 9 AC (Tier M ceiling 16/16).

Content-preservation check (D3): nothing in piece 1 shrinks a body. AC4 compares every rendered
byte to pinned hashes (a headings-only or emptied body changes the hash and fails; both mutants are
subtests), and AC5 permits only the pinned changed-line list for the carried four.

Guards (regression, not stage criteria): `go test ./internal/harness/rosterguard/... -count=1`
(baseline `ok` 32.845s, no name changes in piece 1), `go test ./internal/template -run
'^(TestTemplateNoInternalContentLeak|TestDeployedRoleRuleEntryPointMarkers)$' -count=1`,
`make agents-emit-check embed-manifest-check commands-emit-check`, `GOOS=windows GOARCH=amd64 go build ./...`,
`git log -p --format= -G't\.Skip\(' b4ea2d6c93..HEAD -- '*_test.go'` empty, coverage of
`internal/template/agentemit` at least 85%.

## §2 RED-now ledger (verbatim, tree b4ea2d6c93)

Judging build for the tool measurements: the commands below are `cmp`, `ls`, `git`, `make` and
`go test` run against this tree; no installed `moai` binary produced any ledger cell. The lint run
in the report uses a `moai` built from this tree with `go build -o <scratchpad>/moai-t1649 ./cmd/moai`
at HEAD `b4ea2d6c933ae567bff4a99cab3540932967edfc`.

```text
$ cmp .claude/agents/moai/builder-harness.md internal/template/templates/.claude/agents/moai/builder-harness.md
.claude/agents/moai/builder-harness.md internal/template/templates/.claude/agents/moai/builder-harness.md differ: char 2170, line 38
exit=1
$ cmp .codex/agents/moai/builder-harness.toml internal/template/templates/.codex/agents/moai/builder-harness.toml
.codex/agents/moai/builder-harness.toml internal/template/templates/.codex/agents/moai/builder-harness.toml differ: char 2123, line 32
exit=1
$ cmp .claude/agents/moai/e2e-tester.md internal/template/templates/.claude/agents/moai/e2e-tester.md
.claude/agents/moai/e2e-tester.md internal/template/templates/.claude/agents/moai/e2e-tester.md differ: char 5529, line 92
exit=1
$ cmp .codex/agents/moai/e2e-tester.toml internal/template/templates/.codex/agents/moai/e2e-tester.toml
.codex/agents/moai/e2e-tester.toml internal/template/templates/.codex/agents/moai/e2e-tester.toml differ: char 5468, line 85
exit=1
$ cmp .claude/agents/moai/manager-design.md internal/template/templates/.claude/agents/moai/manager-design.md
.claude/agents/moai/manager-design.md internal/template/templates/.claude/agents/moai/manager-design.md differ: char 2213, line 50
exit=1
$ cmp .codex/agents/moai/manager-design.toml internal/template/templates/.codex/agents/moai/manager-design.toml
.codex/agents/moai/manager-design.toml internal/template/templates/.codex/agents/moai/manager-design.toml differ: char 2171, line 44
exit=1
$ cmp .claude/agents/moai/super-advisor.md internal/template/templates/.claude/agents/moai/super-advisor.md
.claude/agents/moai/super-advisor.md internal/template/templates/.claude/agents/moai/super-advisor.md differ: char 8289, line 129
exit=1
$ cmp .codex/agents/moai/super-advisor.toml internal/template/templates/.codex/agents/moai/super-advisor.toml
.codex/agents/moai/super-advisor.toml internal/template/templates/.codex/agents/moai/super-advisor.toml differ: char 8055, line 124
exit=1
$ ls internal/template/agentspec
ls: internal/template/agentspec: No such file or directory
exit=1
$ git log --reverse --format=%s --grep='feat(SPEC-AGENT-CORE4-001): M'
exit=0
$ make agents-emit-check
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.440s
exit=0
```

```text
$ go test ./internal/template/agentemit -run '^TestNeutralSetIsOneToOne$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.240s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestClaudeRenderUnmappedCapabilityFails$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.239s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestCodexRenderFromNeutral$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.231s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestEightAgentsRenderPinnedBytes$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.234s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestCarriedFourAllowedDiffOnly$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.247s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestCarriedFourCopiesEqual$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.228s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestEmitCheckSummaryAndHandEdit$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.208s [no tests to run]
exit=0
```

Cell reasons: L-C1..L-C8 differ (install-root prefix and, for manager-design, a differing catalog
citation); L-T1..L-T7 print `[no tests to run]` (empty sweep); L-E shows `make agents-emit-check`
exits 0 without any `specs=` summary line; L-G has empty output because no M-commit exists. The
Codex diffs for the carried four are the same lines as the Markdown ones.

## §3 Mutants considered

| Criterion | Mutant | Closed by |
|---|---|---|
| AC4 | emptied or headings-only body | hash mismatch; two mutant subtests |
| AC5 | renderer rewrites more lines than the list | `mutant_extra_edit_fails` |
| AC6 | copy deleted so the comparison is vacuous | per-agent subtests require both files to exist and be non-empty |
| AC7 | summary printed unconditionally | `hand_edit_fails_naming_file` |
| AC8 | M-commits squashed or reordered | exactly three ordered lines; AC9 pins M0's content |

## §4 Definition of Done

AC1-AC9 green; guards green at every milestone commit; the M0 commit holds only the two baseline
files; no `git add -f` of `.moai/reports/**` anywhere; progress.md §E.2/§E.3 filled by the run
phase.

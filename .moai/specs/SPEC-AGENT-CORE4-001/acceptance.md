---
id: SPEC-AGENT-CORE4-001
title: "Acceptance: piece 1"
version: "0.4.0"
created: 2026-10-18
---

# acceptance.md — SPEC-AGENT-CORE4-001 (piece 1)

Document-level pin: RED-now cells were measured on tree `ed36ec14c1` (clean; the code is identical to
`b4ea2d6c93`). `<M0>` is the SHA recorded as `m0_commit:` in `progress.md`. A `go test -v` row is
green only when the listed `--- PASS:` lines print, because `ok ... [no tests to run]` exits 0 (empty
sweep, `verification-completeness.md` §1.1); those rows are decided by named PASS lines. AC-AC4-001..013
are all **release-blocking**. Selectors are anchored. Tests that need synthetic sources (block
markers, unmapped tokens, hand edits) use in-memory or temp-dir fixtures, because none of the twelve
shipped bodies contains a block marker today (`grep -rn 'claude-only\|codex-only'` over the twelve
printed nothing); the fixtures are named in the subtests.

## §1 Criteria

| AC | REQ | Command | Green output | Green at | RED |
|---|---|---|---|---|---|
| AC-AC4-001 | REQ-AC4-001 | `go test ./internal/template/agentemit -run '^TestSourceSetIsOneToOne$' -v -count=1` | PASS `twelve_sources_twelve_definitions`, `rendered_file_without_source_fails`, `source_equals_baseline_c2_bytes` | M1 | L-T1 |
| AC-AC4-002 | REQ-AC4-002 | `go test ./internal/template/agentemit -run '^TestClaudeRenderFromSource$' -v -count=1` | PASS (fixture sources) `claude_only_block_present_in_c2`, `codex_only_block_absent_from_c2`, `render_is_deterministic` | M1 | L-T2 |
| AC-AC4-003 | REQ-AC4-003 | `go test ./internal/template/agentemit -run '^TestCodexRenderFromSource$' -v -count=1` | PASS (fixture sources) `codex_only_block_present_in_c3`, `claude_only_block_absent_from_c3`, `c3_rendered_without_c2_bytes`; PASS `contract_rows_unchanged_for_twelve` | M1 | L-T3 |
| AC-AC4-004 | REQ-AC4-004 | `go test ./internal/template/agentemit -run '^TestRenderUnmappedCapabilityFails$' -v -count=1` | PASS (fixtures) `unmapped_tools_token_names_file_and_token`, `unterminated_marker_names_file`, `nested_marker_names_file`, `no_partial_output` | M1 | L-T4 |
| AC-AC4-005 | REQ-AC4-005 | `go test ./internal/template/agentemit -run '^TestTwelveRenderEqualsM0Parent$' -v -count=1` | PASS `c2` and `c3` for each of the twelve, each compared with `git show <M0>^:<path>` read at test time (24), `mutant_empty_body_fails`, `mutant_headings_only_fails`, `mutant_tampered_baseline_fails` | M1 | L-T5 |
| AC-AC4-006 | REQ-AC4-006 | `go test ./internal/template/agentemit -run '^TestCarriedFourPinnedChange$' -v -count=1` | PASS `old_lines_match_at_baseline`, `changed_lines_equal_the_eight`, `new_lines_have_no_tilde_slash`, `other_lines_unchanged`, `expected_derived_from_baseline_by_documented_command`, `mutant_stub_line_fails` | M2 | L-T6 |
| AC-AC4-007 | REQ-AC4-007 | `go test ./internal/template/agentemit -run '^TestCarriedFourCopiesEqual$' -v -count=1` | PASS per agent `c1_equals_c2`, `c3p_equals_c3` (both files exist and are non-empty) | M2 | L-T7, L-C |
| AC-AC4-008 | REQ-AC4-008 | `go test ./internal/template/agentemit -run '^TestEmitCheckFailsOnHandEdit$' -v -count=1` | PASS `claude_template`, `codex_template` (M1); `claude_project`, `codex_project` (M2); each names the file; `files_unmodified_after_pass`, `files_unmodified_after_fail` (hash of the whole file set before and after) | M1 / M2 | L-T8 |
| AC-AC4-009 | REQ-AC4-009 | `make agents-emit-check` | exit 0 and stdout contains `specs=12 copies=2 diff=0` (M1) or `specs=12 copies=4 diff=0` (M2); stdout-decided | M1 / M2 | L-E |
| AC-AC4-010 | REQ-AC4-010 | (a) `git log --reverse --no-merges --format=%s <M0>^..HEAD -- internal/template/agentemit/source internal/template/templates/.claude/agents/moai internal/template/templates/.codex/agents/moai .claude/agents/moai .codex/agents/moai .moai/specs/SPEC-AGENT-CORE4-001/m0-baseline.md` (b) `git log --name-only --format= --grep='feat(SPEC-AGENT-CORE4-001): M0 baseline'` (c) `grep -n '^m0_commit:' .moai/specs/SPEC-AGENT-CORE4-001/progress.md` (d) `git diff --quiet <M0>^ <M1> -- internal/template/templates/.claude/agents/moai internal/template/templates/.codex/agents/moai` | (a) exactly three subjects in order: the M0 baseline, M1, M2 (merge commits absorbed from main are excluded by `--no-merges`; foreign commits on the 16 paths are absorbed before M0, so they precede `<M0>^`) (b) exactly `.moai/specs/SPEC-AGENT-CORE4-001/m0-baseline.md` (c) one line giving `<M0>` (d) exit 0 (M1 changed no shipped byte); (a)(b)(c) stdout-decided, (d) exit-decided. The history M1, M0, M2 fails (a) because its first line is not the M0 subject; a history with an ordinary-subject edit before M0 fails (a) by the extra line | M2 | L-G1, L-G2 |
| AC-AC4-011 | REQ-AC4-011 | `go test ./internal/harness/rosterguard -run '^TestNumeralSourceDirDischarged$' -v -count=1` | PASS `source_manager_design_discharged`, `no_undischarged_hit_under_source_dir` | M1 | L-T9 |
| AC-AC4-012 | REQ-AC4-012 | (a) `go test ./internal/template/agentemit -run '^(TestGoldenCommittedArtifactsMatchEmission\|TestGoldenMarkdownMismatchDetected)$' -v -count=1` (b) `go test ./internal/template/agentemit -list '^Test'` | (a) PASS for both tests, plus subtests `markdown_equals_fresh_render` and `injected_mismatch_is_red` (the second flips one byte of a rendered C2 in memory and requires the comparison to fail, so a disarmed comparison — `if false`, self-comparison — cannot pass) (b) the printed list contains every one of the 30 test names recorded in `m0-baseline.md` (baseline count 30 at `ed36ec14c1`); the replaced assertion is the one beginning `AC-001's .md face` through the closing brace of its `for` block in `golden_test.go` (lines 106-116 at the baseline), anchored by text, not by line | M1 | L-T10 |
| AC-AC4-013 | REQ-AC4-013 | (a) `go test ./internal/template -run '^TestAgentsLocalEditConfined$' -v -count=1` (b) `git log --format=%s%n%b --no-merges --grep='d-20261018T062320Z-5952' -- AGENTS.local.md` | (a) PASS `hunks_inside_section_2_0`, `removed_lines_are_only_the_listed_sentences` (each removed line is a C1/C2 table row, or contains `not** a byte-identical`, or `If you have edited`), `added_text_states_the_four_rules` (contains `sole edit surface`, `rendered`, `byte-identical`, `make build`), `rest_of_file_byte_identical`; compares the M1 commit with its parent (b) exactly the M1 subject and its body citing the board id | M1 | L-T11, L-A |

REQ coverage (bijection): REQ-AC4-001-AC-AC4-001; REQ-AC4-002-AC-AC4-002; REQ-AC4-003-AC-AC4-003;
REQ-AC4-004-AC-AC4-004; REQ-AC4-005-AC-AC4-005; REQ-AC4-006-AC-AC4-006; REQ-AC4-007-AC-AC4-007;
REQ-AC4-008-AC-AC4-008; REQ-AC4-009-AC-AC4-009; REQ-AC4-010-AC-AC4-010;
REQ-AC4-011-AC-AC4-011; REQ-AC4-012-AC-AC4-012; REQ-AC4-013-AC-AC4-013. 13 REQ, 13 AC, Tier L
ceilings 25 / 25.

Content preservation (nothing shrinks): AC-AC4-005 compares every rendered byte to the baseline read
from git, so an emptied, headings-only or tampered body fails (three mutant subtests); AC-AC4-006
allows only the 8 lines and derives the expected bytes from the baseline, so a stub replacement
fails (`mutant_stub_line_fails`).

Guards (regression, not stage criteria): `go test ./internal/harness/rosterguard/... -count=1`;
`go test ./internal/template -run '^(TestTemplateNoInternalContentLeak|TestDeployedRoleRuleEntryPointMarkers|TestRealSetMarkdownTreeUnmodified)$' -count=1`;
`make embed-manifest-check commands-emit-check`; `GOOS=windows GOARCH=amd64 go build ./...`;
`git log -p --format= -G't\.Skip\(' <M0>^..HEAD -- '*_test.go'` empty; coverage of
`internal/template/agentemit` at least 85%.

## §2 RED-now ledger (verbatim, tree ed36ec14c1)

Provenance: every cell was run unpiped in this session through `bash -c` with the command text
printed as shown (copy-runnable); each `exit=` is the real exit code. No installed `moai` binary
produced any cell. The spec lint cited in the revision report uses a `moai` built from this tree.
Cell ids: L-C = the eight `cmp` cells; L-T1..T11 = the selector cells in command order (T9 is the
rosterguard one); L-G1 = the `m0_commit` grep (exit 1: no record yet), L-G2 = the M0 file-list
`git log` (empty); L-A = the board-citation `git log` on `AGENTS.local.md` (empty); L-E = `make
agents-emit-check` (exit 0, no `specs=` line). The remaining AC-AC4-010 cells need an M0 commit
and cannot run before it.

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
$ ls internal/template/agentemit/source
ls: internal/template/agentemit/source: No such file or directory
exit=1
$ grep -n '^m0_commit:' .moai/specs/SPEC-AGENT-CORE4-001/progress.md
exit=1
$ git log --name-only --format= --grep='feat(SPEC-AGENT-CORE4-001): M0 baseline'
exit=0
$ git log --format=%s%n%b --no-merges --grep='d-20261018T062320Z-5952' -- AGENTS.local.md
exit=0
$ make agents-emit-check
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.461s
exit=0
$ go test ./internal/template/agentemit -run '^TestSourceSetIsOneToOne$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.264s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestClaudeRenderFromSource$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.266s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestCodexRenderFromSource$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.264s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestRenderUnmappedCapabilityFails$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.270s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestTwelveRenderEqualsM0Parent$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.271s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestCarriedFourPinnedChange$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.267s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestCarriedFourCopiesEqual$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.282s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestEmitCheckFailsOnHandEdit$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.276s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestGoldenMarkdownMismatchDetected$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.280s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestAgentsLocalEditConfined$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.278s [no tests to run]
exit=0
$ go test ./internal/harness/rosterguard -run '^TestNumeralSourceDirDischarged$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness/rosterguard	0.335s [no tests to run]
exit=0
```

## §3 Expected results of the 8 changes

The expected post-change bytes are derived, not stored: for each carried-four agent, take
`git show <M0>^:<C2 path>`, replace each old line of spec §C.1 (matched by content) by its new line,
and emit C3 from the result with the unchanged emitter; `TestCarriedFourPinnedChange` performs this
and `m0-baseline.md` records the resulting sha256 values. Plan-time values, valid only when the
four files at `<M0>^` equal their bytes at `8260bbaa0b` (informational cross-check; computed in
this session by the same procedure):

| Agent | sha256 C2 | sha256 C3 (= C3p) |
|---|---|---|
| builder-harness | `7ce873448959908ab0c0a1b3fd06aabfae34853f3b0e684ee70cdb6cf7c2b619` | `48affeb070cef387e580ab76af7bda697f0933210bd464e408257490ba34b734` |
| e2e-tester | `3606c7fdc85b2390a329f036daa6e98787e565a3ff5fa4973894f5f51c6ccd20` | `340901c6f887ab01655ab1ddcacadc36b8b72c94f19a246bb1adc4d1da8ef770` |
| manager-design | `84e533d3b2ac6878e59237a0c6a43bf1b56f585172005d9c17bdcb2206b9dab6` | `c187776dbab38ff0e4e280e68bb01b9b099c05f84209dbddce258a4ed039b644` |
| super-advisor | `d765c4af741afc06e8280098ce15521205abc34cc116f0ef13c2c14b186fc7d1` | `9fba77a1162a3c61525e043bbd337627f7fcb31d0f5a04bb58ef9db32729b824` |

## §4 Mutants considered

| Criterion | Mutant | Closed by |
|---|---|---|
| AC-AC4-005 | emptied, headings-only or tampered body | baseline read from git; three mutant subtests |
| AC-AC4-006 | stub in place of line 92 | expected derived from the baseline by the documented command |
| AC-AC4-007 | copy deleted so the comparison is vacuous | both files must exist and be non-empty |
| AC-AC4-008 | check overwrites the hand edit and returns an error | `files_unmodified_after_fail` |
| AC-AC4-009 | summary printed unconditionally | AC-AC4-008 hand-edit failure |
| AC-AC4-010 | M1 then M0 then M2; ordinary-subject edit before M0; tampered M1 | (a) first line must be the M0 subject; (d) and AC-AC4-005 |
| AC-AC4-003 | C3 derived from C2 so a claude-only block leaks | `c3_rendered_without_c2_bytes` and the two block subtests |
| AC-AC4-012 | comparison replaced by `if false` or self-comparison | `injected_mismatch_is_red`; the name list |
| AC-AC4-013 | rewrite of unrelated lines | `rest_of_file_byte_identical` |

## §5 Definition of Done

AC-AC4-001..013 green at their milestone; the guards green at every milestone commit; the M0 commit
contains exactly `m0-baseline.md`; no `git add -f` of `.moai/reports/**` anywhere; progress.md §E.2
and §E.3 written by the run phase.

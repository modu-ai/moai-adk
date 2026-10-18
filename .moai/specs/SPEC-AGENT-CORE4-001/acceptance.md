---
id: SPEC-AGENT-CORE4-001
title: "Acceptance: piece 1"
version: "0.3.0"
created: 2026-10-18
---

# acceptance.md — SPEC-AGENT-CORE4-001 (piece 1)

Document-level pin: RED-now cells were measured on tree `8260bbaa0b` (clean; code identical to
`b4ea2d6c93`). A `go test -v` row is green only when the listed `--- PASS:` lines print, because
`ok ... [no tests to run]` exits 0 (empty sweep, `verification-completeness.md` §1.1); those rows
are decided by named PASS lines, not by exit code alone. Classes: AC-AC4-001..011 are
**release-blocking**; AC-AC4-012 is a **regression-guard** (its check pipes a diff and is outside
the single-invocation form). Selectors are anchored.

## §1 Criteria

| AC | REQ | Command | Green output | Green at | RED |
|---|---|---|---|---|---|
| AC-AC4-001 | REQ-AC4-001 | `go test ./internal/template/agentemit -run '^TestSourceSetIsOneToOne$' -v -count=1` | PASS `twelve_sources_twelve_definitions`, `rendered_file_without_source_fails` | M1 | L-T1 |
| AC-AC4-002 | REQ-AC4-002 | `go test ./internal/template/agentemit -run '^TestClaudeRenderFromSource$' -v -count=1` | PASS `claude_only_block_present_in_c2`, `codex_only_block_absent_from_c2`, `render_is_deterministic` | M1 | L-T2 |
| AC-AC4-003 | REQ-AC4-003 | `go test ./internal/template/agentemit -run '^TestCodexRenderFromSource$' -v -count=1` | PASS `codex_only_block_present_in_c3`, `claude_only_block_absent_from_c3`, `c3_rendered_without_c2_bytes`, `contract_rows_unchanged_for_twelve` | M1 | L-T3 |
| AC-AC4-004 | REQ-AC4-004 | `go test ./internal/template/agentemit -run '^TestRenderUnmappedCapabilityFails$' -v -count=1` | PASS `error_names_file_and_token`, `no_partial_output` | M1 | L-T4 |
| AC-AC4-005 | REQ-AC4-005 | `go test ./internal/template/agentemit -run '^TestTwelveRenderPinnedBytes$' -v -count=1` | PASS `c2` and `c3` for each of the twelve (24, the carried four compared to the pin file at M1 and to the post pins from M2), `mutant_empty_body_fails`, `mutant_headings_only_fails` | M1 | L-T5 |
| AC-AC4-006 | REQ-AC4-006 | `go test ./internal/template/agentemit -run '^TestCarriedFourPinnedChange$' -v -count=1` | PASS `changed_lines_equal_the_eight`, `new_lines_have_no_tilde_slash`, `post_c2_c3_equal_post_pins`, `pins_equal_acceptance_table`, `mutant_stub_line_fails` | M2 | L-T6 |
| AC-AC4-007 | REQ-AC4-007 | `go test ./internal/template/agentemit -run '^TestCarriedFourCopiesEqual$' -v -count=1` | PASS per agent `c1_equals_c2`, `c3p_equals_c3` (both files exist and are non-empty) | M2 | L-T7, L-C |
| AC-AC4-008 | REQ-AC4-008 | `go test ./internal/template/agentemit -run '^TestEmitCheckFailsOnHandEdit$' -v -count=1` | PASS `claude_template`, `codex_template` (M1); `claude_project`, `codex_project` (M2); each names the file | M1 / M2 | L-T8 |
| AC-AC4-009 | REQ-AC4-009 | `make agents-emit-check` | exit 0 and stdout contains `specs=12 copies=2 diff=0` (M1) or `specs=12 copies=4 diff=0` (M2); stdout-decided | M1 / M2 | L-E |
| AC-AC4-010 | REQ-AC4-010 | (a) `git log --reverse --format=%s b4ea2d6c93..HEAD -- internal/template/agentemit/source internal/template/templates/.claude/agents/moai internal/template/templates/.codex/agents/moai .claude/agents/moai .codex/agents/moai` (b) `git log --format=%s b4ea2d6c93..HEAD -- internal/template/agentemit/testdata/agents-head.sha256` (c) `git log --name-only --format= --grep='feat(SPEC-AGENT-CORE4-001): M0 baseline'` | (a) exactly the M1 subject then the M2 subject, nothing earlier (b) exactly the M0 subject (c) exactly `.moai/specs/SPEC-AGENT-CORE4-001/m0-baseline.md` and `internal/template/agentemit/testdata/agents-head.sha256`; all stdout-decided | M2 | L-G1..L-G3 |
| AC-AC4-011 | REQ-AC4-011 | `go test ./internal/harness/rosterguard -run '^TestNumeralSourceDirDischarged$' -v -count=1` | PASS `source_manager_design_discharged`, `no_undischarged_hit_under_source_dir` | M1 | L-T9 |
| AC-AC4-012 | REQ-AC4-012 | `git diff -U0 b4ea2d6c93 HEAD -- internal/template/agentemit/golden_test.go internal/template/agentemit/agentemit_test.go internal/template/agentemit/agentemit_edge_test.go internal/template/agentemit/permission_contract_test.go` | every removed line that contains `t.Error` or `t.Fatal` lies inside the Markdown identity block of `golden_test.go` (lines 105-113 at the pin tree) and has its replacement added in the same hunk; no removed `t.Run(` line | M1 | guard |

REQ coverage (bijection): REQ-AC4-001-AC-AC4-001; REQ-AC4-002-AC-AC4-002; REQ-AC4-003-AC-AC4-003;
REQ-AC4-004-AC-AC4-004; REQ-AC4-005-AC-AC4-005; REQ-AC4-006-AC-AC4-006; REQ-AC4-007-AC-AC4-007;
REQ-AC4-008-AC-AC4-008; REQ-AC4-009-AC-AC4-009; REQ-AC4-010-AC-AC4-010;
REQ-AC4-011-AC-AC4-011; REQ-AC4-012-AC-AC4-012. 12 REQ, 12 AC, Tier L ceilings 25 / 25.

Content preservation (nothing shrinks): AC-AC4-005 compares every rendered byte to pinned hashes, so
an emptied or headings-only body changes the hash and fails (both mutants are subtests);
AC-AC4-006 allows only the 8 lines below and compares the results to the post pins, so a stub
replacement fails (`mutant_stub_line_fails`).

Guards (regression, not stage criteria): `go test ./internal/harness/rosterguard/... -count=1`;
`go test ./internal/template -run '^(TestTemplateNoInternalContentLeak|TestDeployedRoleRuleEntryPointMarkers)$' -count=1`;
`make embed-manifest-check commands-emit-check`; `GOOS=windows GOARCH=amd64 go build ./...`;
`git log -p --format= -G't\.Skip\(' b4ea2d6c93..HEAD -- '*_test.go'` empty; coverage of
`internal/template/agentemit` at least 85%.

## §2 RED-now ledger (verbatim, tree 8260bbaa0b)

Provenance: every cell was run unpiped in this session, and each `exit=` line is the real exit code
of that command (an earlier revision back-filled six selector exit codes; this ledger replaces it).
No installed `moai` binary produced any cell: they are `cmp`, `ls`, `git`, `make` and `go test`
runs against this tree. The spec lint in the report uses a `moai` built from this tree at
`8260bbaa0b0198a05c43ef9cd845ae548e74c4ad`. Cell ids: L-C = the eight `cmp` cells, L-T1..T9 = the
selector cells in command order (T9 is the rosterguard one), L-G1..G3 = the three `git log`
cells, L-E = `make agents-emit-check` (exit 0 and no `specs=` line).

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
$ git log --reverse --format=%s b4ea2d6c93..HEAD -- internal/template/agentemit/source internal/template/templates/.claude/agents/moai internal/template/templates/.codex/agents/moai .claude/agents/moai .codex/agents/moai
exit=0
$ git log --format=%s b4ea2d6c93..HEAD -- internal/template/agentemit/testdata/agents-head.sha256
exit=0
$ git log --name-only --format= --grep=feat(SPEC-AGENT-CORE4-001): M0 baseline
exit=0
$ make agents-emit-check
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.411s
exit=0
$ go test ./internal/template/agentemit -run ^TestSourceSetIsOneToOne$ -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.252s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run ^TestClaudeRenderFromSource$ -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.246s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run ^TestCodexRenderFromSource$ -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.265s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run ^TestRenderUnmappedCapabilityFails$ -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.255s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run ^TestTwelveRenderPinnedBytes$ -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.241s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run ^TestCarriedFourPinnedChange$ -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.242s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run ^TestCarriedFourCopiesEqual$ -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.241s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run ^TestEmitCheckFailsOnHandEdit$ -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.235s [no tests to run]
exit=0
$ go test ./internal/harness/rosterguard -run ^TestNumeralSourceDirDischarged$ -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness/rosterguard	0.329s [no tests to run]
exit=0
```

## §3 The 8 line changes and post pins (plan-phase authored)

Chosen text rule: the new line is the project-copy (C1) line, which uses project-relative paths, with
one exception (line 143 of e2e-tester, where C1 itself holds `~/.claude/`; the new line drops the
`~/` of that line, which also fixes the shipped `~/~/.claude/` defect of C2). For manager-design line
206 the C1 citation survives, because `.moai/config/sections/delegation.yaml` exists in this
repository while `CLAUDE.md` (the C2 citation) does not; the numeral "13 retained agents" stays so
the three existing registry rows keep matching. The leader may overrule through decision Q19 (the post pins then change and are recomputed).

Pairs (old = C2 line at the pin tree; new = the text shown; "same" = old line with only the quoted
substitution). Backticks inside are literal.

1. builder-harness:38 — old has `~/.claude/agents/` inside the table cell "`.claude/agents/` or `~/.claude/agents/` (with `--moai`)"; new: `~/.claude/agents/` replaced by `.claude/agents/moai/`.
2. builder-harness:155 — old: `` `~/.claude/agents/manager-spec.md` for the verbatim line: ``; new: `` `.claude/agents/moai/manager-spec.md` for the verbatim line: ``.
3. e2e-tester:92 — old holds `~/.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md`; new: the same sentence with `.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md`.
4. e2e-tester:143 — old holds `~/~/.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md`; new: `.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md`.
5. manager-design:50 — old: `` `~/.claude/skills/moai/workflows/design.md` (D1-D5 step headings). Summary: ``; new: the same with `.claude/skills/moai/workflows/design.md`.
6. manager-design:203 — old: `- **Design pipeline skill**: ` + `~/.claude/skills/moai/workflows/design.md` + ` (D1-D5).`; new: with `.claude/skills/moai/workflows/design.md`.
7. manager-design:206 — old: `- **Agent catalog**: ` + `CLAUDE.md` + ` § 4 (13 retained agents — manager-design is row 6 of the Selection Decision Tree).`; new: `- **Agent catalog**: ` + `.moai/config/sections/delegation.yaml` + ` (13 retained agents — manager-design is row 6 of the Selection Decision Tree).`
8. super-advisor:129 — old: `- Advisor/Evaluator separation: ` + `~/.claude/agents/{plan-auditor,sync-auditor}.md` + `.`; new: `.claude/agents/moai/{plan-auditor,sync-auditor}.md`.

Full old and new lines are the C2 line and the C1 line of the same number respectively (rows 3 and
4: C1 line 92 and the C1 line 143 with its `~/` removed); they are produced and checked by
`TestCarriedFourPinnedChange` from the pin file and these pins. Other `~/` occurrences (for example
super-advisor line 119) stay unchanged. Post pins, computed in this session by applying the 8
changes to the pin-tree C2 files and emitting C3 with the unchanged emitter:

| Agent | sha256 C2 | sha256 C3 (= C3p) |
|---|---|---|
| builder-harness | `7ce873448959908ab0c0a1b3fd06aabfae34853f3b0e684ee70cdb6cf7c2b619` | `48affeb070cef387e580ab76af7bda697f0933210bd464e408257490ba34b734` |
| e2e-tester | `3606c7fdc85b2390a329f036daa6e98787e565a3ff5fa4973894f5f51c6ccd20` | `340901c6f887ab01655ab1ddcacadc36b8b72c94f19a246bb1adc4d1da8ef770` |
| manager-design | `84e533d3b2ac6878e59237a0c6a43bf1b56f585172005d9c17bdcb2206b9dab6` | `c187776dbab38ff0e4e280e68bb01b9b099c05f84209dbddce258a4ed039b644` |
| super-advisor | `d765c4af741afc06e8280098ce15521205abc34cc116f0ef13c2c14b186fc7d1` | `9fba77a1162a3c61525e043bbd337627f7fcb31d0f5a04bb58ef9db32729b824` |

## §4 Mutants considered

| Criterion | Mutant | Closed by |
|---|---|---|
| AC-AC4-005 | emptied or headings-only body | hash mismatch; two mutant subtests |
| AC-AC4-006 | stub in place of line 92 listed in the pin list | post pins computed at plan time, not by the implementer |
| AC-AC4-007 | copy deleted so the comparison is vacuous | both files must exist and be non-empty |
| AC-AC4-009 | summary printed unconditionally | AC-AC4-008 hand-edit failure |
| AC-AC4-010 | implementation commit with an ordinary subject before M0, or pin file edited after M0 | path-based (a) and (b) |
| AC-AC4-003 | C3 derived from C2 so a claude-only block leaks | `c3_rendered_without_c2_bytes` and the two block subtests |

## §5 Definition of Done

AC-AC4-001..011 green at their milestone and AC-AC4-012 clean; the guards green at every milestone
commit; M0 contains exactly its two files; no `git add -f` of `.moai/reports/**` anywhere; progress.md
§E.2/§E.3 written by the run phase.

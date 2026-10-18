---
id: SPEC-AGENT-CORE4-001
title: "Acceptance: piece 1"
version: "0.6.0"
created: 2026-10-18
---

# acceptance.md — SPEC-AGENT-CORE4-001 (piece 1)

Document-level pin: RED-now cells were measured on tree `06095bb8fa`. `<M0>` is the M0 baseline
commit; the M1 run records its SHA in `progress.md` §E.2. A `go test -v` row is green only when the
listed `--- PASS:` lines print, because `ok ... [no tests to run]` exits 0 (empty sweep,
`verification-completeness.md` §1.1). AC-AC4-001..004, 007..011 and AC-AC4-012(a) are **release-blocking** (RED-now observed; AC-AC4-012(a) is a permanent test). AC-AC4-005, AC-AC4-006 and
AC-AC4-012 (b)-(d) and (e) are **one-time milestone cells** run once at M1 or M2 and pasted into
`progress.md` §E.2: they have no RED-now and are regression-guards, not release-blocking criteria;
nothing re-runs them. Tests that need synthetic sources (block markers,
unmapped tokens, hand edits) use in-memory or temp-dir fixtures, because none of the twelve shipped
bodies contains a block marker.

## §1 Criteria (12 rows = 12 AC)

| AC | REQ | Command | Green output | Green at | RED |
|---|---|---|---|---|---|
| AC-AC4-001 | REQ-AC4-001 | `go test ./internal/template/agentemit -run '^TestSourceSetIsOneToOne$' -v -count=1` | PASS `twelve_sources_twelve_definitions`, `rendered_file_without_source_fails` | M1 | L-T1 |
| AC-AC4-002 | REQ-AC4-002 | `go test ./internal/template/agentemit -run '^TestClaudeRenderFromSource$' -v -count=1` | PASS (fixtures) `claude_only_block_present_in_c2`, `codex_only_block_absent_from_c2`, `render_is_deterministic` | M1 | L-T2 |
| AC-AC4-003 | REQ-AC4-003 | `go test ./internal/template/agentemit -run '^TestCodexRenderFromSource$' -v -count=1` | PASS (fixtures) `codex_only_block_present_in_c3`, `claude_only_block_absent_from_c3`, `c3_rendered_without_c2_bytes`; PASS `contract_rows_unchanged_for_twelve` (compares the emitted permission-contract rows with the in-repo contract table in `agents-codex.yaml`, not with git or `<M0>`; it changes with that table, not with the agents) | M1 | L-T3 |
| AC-AC4-004 | REQ-AC4-004 | `go test ./internal/template/agentemit -run '^TestRenderUnmappedCapabilityFails$' -v -count=1` | PASS (fixtures) `unmapped_tools_token_names_file_and_token`, `unterminated_marker_names_file`, `nested_marker_names_file`, `no_partial_output` | M1 | L-T4 |
| AC-AC4-005 | REQ-AC4-005 | one-time, at the M1 commit, with `<M0>` from `git log --format=%H --grep='feat(SPEC-AGENT-CORE4-001): M0 baseline'`: (a) `git diff --quiet <M0>^ <M0> -- . ':(exclude).moai/specs/SPEC-AGENT-CORE4-001/m0-baseline.md'` (b) `git merge-base --is-ancestor <M0> HEAD` (c) `git diff --quiet <M0> HEAD -- internal/template/templates/.claude/agents/moai internal/template/templates/.codex/agents/moai` (d) `diff -rq internal/template/agentemit/source internal/template/templates/.claude/agents/moai` | all four exit 0: M0 changed nothing but the baseline file, M0 is an ancestor of M1, M1 changed no shipped C2 or C3 byte, and the twelve sources equal C2; exit-decided; the four outputs and the SHA are pasted into `progress.md` §E.2 | M1 | one-time cell, no RED-now |
| AC-AC4-006 | REQ-AC4-006 | one-time, at the M2 commit: `git diff -U0 --no-color <M0> HEAD -- internal/template/templates/.claude/agents/moai/builder-harness.md internal/template/templates/.claude/agents/moai/e2e-tester.md internal/template/templates/.claude/agents/moai/manager-design.md internal/template/templates/.claude/agents/moai/super-advisor.md` | stdout contains exactly 8 hunks (`@@ -N +N @@`, trailing text ignored, `index` lines ignored) whose `-` and `+` lines are the 8 pairs of spec §C.1, nothing else; stdout-decided; pasted into `progress.md` §E.2 | M2 | one-time cell, no RED-now |
| AC-AC4-007 | REQ-AC4-007 | `go test ./internal/template/agentemit -run '^TestCarriedFourCopiesEqual$' -v -count=1` | PASS per agent `c1_equals_c2`, `c3p_equals_c3` (both files exist and are non-empty) | M2 | L-T5, L-C |
| AC-AC4-008 | REQ-AC4-008 | `go test ./internal/template/agentemit -run '^TestEmitCheckFailsOnHandEdit$' -v -count=1` | PASS `claude_template`, `codex_template` (M1); `claude_project`, `codex_project` (M2); each names the file; `files_unmodified_after_pass`, `files_unmodified_after_fail` | M1 / M2 | L-T6 |
| AC-AC4-009 | REQ-AC4-009 | `make agents-emit-check` | exit 0 and stdout contains `specs=12 copies=2 diff=0` (M1) or `specs=12 copies=4 diff=0` (M2); stdout-decided | M1 / M2 | L-E |
| AC-AC4-010 | REQ-AC4-010 | `go test ./internal/harness/rosterguard -run '^TestNumeralSourceDirDischarged$' -v -count=1` | PASS `source_manager_design_discharged`, `no_undischarged_hit_under_source_dir` | M1 | L-T8 |
| AC-AC4-011 | REQ-AC4-011 | (a) `go test ./internal/template/agentemit -run '^(TestGoldenCommittedArtifactsMatchEmission\|TestGoldenMarkdownMismatchDetected)$' -v -count=1` (b) one-time, M1 and M2: `go test ./internal/template/agentemit -list '^Test'` | (a) PASS both tests plus subtests `markdown_equals_fresh_render` and `injected_mismatch_is_red` (flips one byte of a rendered C2 in memory and requires the comparison to fail, so `if false` or self-comparison cannot pass) (b) the list contains each of the 30 names recorded in `m0-baseline.md`; the replaced assertion is the block from `AC-001's .md face` to the closing brace of its `for` (lines 106-116 at the baseline), anchored by text | M1 | L-T7 |
| AC-AC4-012 | REQ-AC4-012 | (a) `go test ./internal/template -run '^TestAgentsLocalSection20Text$' -v -count=1` (b) one-time, at the M2 commit: `git diff --numstat <M2>^ <M2> -- AGENTS.local.md` (c) `git diff -U0 --no-color <M2>^ <M2> -- AGENTS.local.md` (d) `git log -1 --format=%B <M2>` (e) one-time, at the M2 commit: `python3 -c "s=open('.moai/specs/SPEC-AGENT-CORE4-001/spec.md').read();c=s[s.index('### §C.2'):s.index('## §D')];new=[l[2:] for l in c.split(chr(10)) if l.startswith('+ ')];L=open('AGENTS.local.md').read().split(chr(10));bad=[n[:40] for n in new if L.count(n)!=1];print(len(new),bad);raise SystemExit(1 if bad or len(new)!=9 else 0)"` (reads every `+ ` line of spec §C.2 and requires each to occur exactly once as a whole line) | (a) permanent, depends on the file only and survives piece 2's later edits: PASS `section_2_0_exists`, `old_phrases_absent_from_section_2_0` (whole phrases that occur only in the eight old lines: `Local dogfooding, hand-edited`, `Distribution mirror, hand-edited`, `the third is not edited by hand`, `only one of them is not a hand-edit target`, `not** a byte-identical`, `fix C2 and regenerate`, `Machine-emitted from C2`, `emitted from \`.md\`` — never the bare word `hand-edited`, which the new rows contain), `source_path_named` (`internal/template/agentemit/source/` appears in §2.0). The section is located by the heading prefix `### [HARD] §2.0` up to the next `### ` heading (b) stdout `9\t8\tAGENTS.local.md` (c) removed and added lines equal the §C.2 pairs, nothing else (d) the body contains `d-20261018T062320Z-5952` (e) prints `9 []` and exits 0: every new §C.2 line is present as a whole line (`new_lines_present_exactly`, one-time because piece 2 rewrites the transitional clauses); (b)-(e) stdout-decided, pasted into `progress.md` §E.2 | M2 | L-T9, L-P |

REQ coverage (bijection): REQ-AC4-001-AC-AC4-001; REQ-AC4-002-AC-AC4-002; REQ-AC4-003-AC-AC4-003;
REQ-AC4-004-AC-AC4-004; REQ-AC4-005-AC-AC4-005; REQ-AC4-006-AC-AC4-006; REQ-AC4-007-AC-AC4-007;
REQ-AC4-008-AC-AC4-008; REQ-AC4-009-AC-AC4-009; REQ-AC4-010-AC-AC4-010;
REQ-AC4-011-AC-AC4-011; REQ-AC4-012-AC-AC4-012. 12 REQ, 12 AC, Tier L ceilings 25 / 25.

Handling after landing: no frozen-equality test exists. Permanent tests and guards after M2:
AC-AC4-001, 002, 003, 004, 007, 008, 009, 010, 011(a), and AC-AC4-012(a) (old phrases absent, section
exists, source path named). One-time cells, never re-run: AC-AC4-005, AC-AC4-006, AC-AC4-011(b),
AC-AC4-012(b)-(e). A legitimate later edit goes to the agent source and `make agents-emit`; no
permanent test compares against `<M0>` or reads git history, so a shallow checkout runs them, and
piece 2's rewrite of the transitional §2.0 clauses keeps the permanent check green.

Guards (regression, not stage criteria): `go test ./internal/harness/rosterguard/... -count=1`;
`go test ./internal/template -run '^(TestTemplateNoInternalContentLeak|TestDeployedRoleRuleEntryPointMarkers|TestRealSetMarkdownTreeUnmodified)$' -count=1`;
`make embed-manifest-check commands-emit-check`; `GOOS=windows GOARCH=amd64 go build ./...`;
`git log -p --format= -G't\.Skip\(' <M0>..HEAD -- '*_test.go'` empty; coverage of
`internal/template/agentemit` at least 85%.

## §2 RED-now ledger (verbatim, tree 06095bb8fa)

Provenance: every cell was run unpiped through `bash -c` with the command text printed as shown
(copy-runnable); each `exit=` is the real exit code of that command. No installed `moai` binary
produced any cell. Cell ids: L-C = the eight `cmp` cells (RED: exit 1, files differ); L-S = the
`ls` of the missing source directory (RED: exit 1); L-T1..T9 = the selector cells in command order
(RED by the empty-sweep rule: exit 0 with `[no tests to run]`, decided by absent PASS lines; T8 is
rosterguard, T9 is `./internal/template`); L-E = `make agents-emit-check` (exit 0, no `specs=`
line, hence RED for AC-AC4-009); L-P = the final `grep -c` of one exact old phrase (count 1 now, 0 after M2, so it genuinely flips; RED for the `old_phrases_absent` subtest). One-time cells AC-AC4-005,
006 and AC-AC4-012 (b)-(e) have no RED-now and are not claimed to.

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
$ make agents-emit-check
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.527s
exit=0
$ go test ./internal/template/agentemit -run '^TestSourceSetIsOneToOne$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.267s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestClaudeRenderFromSource$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.273s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestCodexRenderFromSource$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.282s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestRenderUnmappedCapabilityFails$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.306s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestCarriedFourCopiesEqual$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.284s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestEmitCheckFailsOnHandEdit$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.278s [no tests to run]
exit=0
$ go test ./internal/template/agentemit -run '^TestGoldenMarkdownMismatchDetected$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.275s [no tests to run]
exit=0
$ go test ./internal/harness/rosterguard -run '^TestNumeralSourceDirDischarged$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness/rosterguard	0.358s [no tests to run]
exit=0
$ go test ./internal/template -run '^TestAgentsLocalSection20Text$' -v -count=1
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template	0.516s [no tests to run]
exit=0
$ grep -c 'Distribution mirror, hand-edited' AGENTS.local.md
1
exit=0
```

## §3 Expected results of the 8 line changes

The literal expected lines are in spec §C.1; AC-AC4-006 compares the real `git diff -U0` output to
them. Informational cross-check of the resulting files, valid when the four files at `<M0>^` equal
their bytes at `8260bbaa0b` (computed in this session by applying the pairs and emitting C3 with the
unchanged emitter):

| Agent | sha256 C2 | sha256 C3 (= C3p) |
|---|---|---|
| builder-harness | `7ce873448959908ab0c0a1b3fd06aabfae34853f3b0e684ee70cdb6cf7c2b619` | `48affeb070cef387e580ab76af7bda697f0933210bd464e408257490ba34b734` |
| e2e-tester | `3606c7fdc85b2390a329f036daa6e98787e565a3ff5fa4973894f5f51c6ccd20` | `340901c6f887ab01655ab1ddcacadc36b8b72c94f19a246bb1adc4d1da8ef770` |
| manager-design | `84e533d3b2ac6878e59237a0c6a43bf1b56f585172005d9c17bdcb2206b9dab6` | `c187776dbab38ff0e4e280e68bb01b9b099c05f84209dbddce258a4ed039b644` |
| super-advisor | `d765c4af741afc06e8280098ce15521205abc34cc116f0ef13c2c14b186fc7d1` | `9fba77a1162a3c61525e043bbd337627f7fcb31d0f5a04bb58ef9db32729b824` |

## §4 Mutants considered

| Criterion | Mutant | Closed by |
|---|---|---|
| AC-AC4-005 | M1 edits a shipped byte; sources differ from C2; M0 carries extra files | (c), (d), (a) |
| AC-AC4-006 | a stub replaces line 92 | the real diff is compared to the literal pairs |
| AC-AC4-007 | copy deleted so the comparison is vacuous | both files must exist and be non-empty |
| AC-AC4-008 | check overwrites the hand edit and returns an error | `files_unmodified_after_fail` |
| AC-AC4-009 | summary printed unconditionally | AC-AC4-008 hand-edit failure |
| AC-AC4-003 | C3 derived from C2 so a claude-only block leaks | `c3_rendered_without_c2_bytes` and the block subtests |
| AC-AC4-011 | comparison replaced by `if false` or self-comparison | `injected_mismatch_is_red`; the name list |
| AC-AC4-012 | no-op, whole-line deletion of line 101 or 109, or an unrelated edit | simulated in a scratch copy: a no-op fails the permanent check (a) (old phrases present); deletion of a whole new line (for example new line 101 or 109) is not caught by (a) and is closed by the one-time M2 cells (b) numstat, (c) exact pairs and (e) whole-line comparison |

Not covered, said once: the order "M0 before any change" is confirmed only by AC-AC4-005 at M1 (M0
changes nothing but the baseline file and is an ancestor of M1); the SPEC claims nothing stronger
and defines the baseline as whatever tree M0 sat on.

## §5 Definition of Done

AC-AC4-001..012 green at their milestone; one-time cells pasted into `progress.md` §E.2; the guards
green at every milestone commit; the M0 commit contains exactly `m0-baseline.md`; the M2 commit
message cites board `d-20261018T062320Z-5952`; no `git add -f` of `.moai/reports/**`.

# progress.md — SPEC-AGENT-CORE4-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-18
plan_phase_head: 06095bb8fa (measurement tree; the revision commit lands on top)
plan_phase_branch: WT-3-3-10
plan_phase_worktree: .moai/worktrees/t1649
plan_phase_artifacts: 7 (spec.md, plan.md, acceptance.md, design.md, research.md, progress.md, decision-index.md — Tier L set of 5 + progress + decision-index)
revision: 0.6.0 (artifacts at c3e4ef3751), piece 1 only; plan-audit iter7 PASS-WITH-DEBT 0.90
requirements: 12 (ceiling 25)
acceptance_criteria: 12 (ceiling 25)
open_decisions_gating_piece_1: none (Q4 and Q17 closed by the operator, board d-20261018T062320Z-5952)
m0_baseline: recorded at M1 in §E.2 (no placeholder line here)

## §E.2 Run-phase Evidence

Run-phase evidence for piece 1 (card t1649). Commands were run unpiped through `bash -c` from the card worktree
`.moai/worktrees/t1649`; every `exit=` is the real exit code. No installed `moai` binary produced any cell below; the
toolchain is `go1.26.8 darwin/arm64`. Verbatim outputs are pasted, not summarized.

### M0 baseline and M1

- `<M0>` = `040a8afd4811d93838421a3c922ae50f1c1521e8` (`feat(SPEC-AGENT-CORE4-001): M0 baseline (card t1649)`), parent
  `<M0>^` = `da248a38da6e5cdb1257642cb5b2c473220d0954`; the baseline file is `m0-baseline.md` (committed alone).
- M1 = `a252734aceca1879ab8338f26f19ee5512074a97` (`feat(SPEC-AGENT-CORE4-001): M1 agent sources, renderers, check wiring (card t1649)`).

#### AC-AC4-005 one-time cell (run at the M1 commit)

```text
$ git log --format=%H --grep='feat(SPEC-AGENT-CORE4-001): M0 baseline'
040a8afd4811d93838421a3c922ae50f1c1521e8
HEAD=a252734aceca1879ab8338f26f19ee5512074a97
$ git diff --quiet 040a8afd4811d93838421a3c922ae50f1c1521e8^ 040a8afd4811d93838421a3c922ae50f1c1521e8 -- . ':(exclude).moai/specs/SPEC-AGENT-CORE4-001/m0-baseline.md'
exit=0
$ git merge-base --is-ancestor 040a8afd4811d93838421a3c922ae50f1c1521e8 HEAD
exit=0
$ git diff --quiet 040a8afd4811d93838421a3c922ae50f1c1521e8 HEAD -- internal/template/templates/.claude/agents/moai internal/template/templates/.codex/agents/moai
exit=0
$ diff -rq internal/template/agentemit/source internal/template/templates/.claude/agents/moai
exit=0
```

#### M1 RED, before any implementation (stubs returning "not implemented"; the source directory did not exist)

```text
--- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.00s)
--- FAIL: TestGoldenMarkdownMismatchDetected (0.00s)
--- FAIL: TestClaudeRenderFromSource (0.00s)
    --- FAIL: TestClaudeRenderFromSource/claude_only_block_present_in_c2 (0.00s)
    --- FAIL: TestClaudeRenderFromSource/codex_only_block_absent_from_c2 (0.00s)
    --- PASS: TestClaudeRenderFromSource/render_is_deterministic (0.00s)
--- FAIL: TestCodexRenderFromSource (0.00s)
    --- FAIL: TestCodexRenderFromSource/codex_only_block_present_in_c3 (0.00s)
    --- FAIL: TestCodexRenderFromSource/claude_only_block_absent_from_c3 (0.00s)
    --- FAIL: TestCodexRenderFromSource/c3_rendered_without_c2_bytes (0.00s)
    --- FAIL: TestCodexRenderFromSource/contract_rows_unchanged_for_twelve (0.00s)
--- FAIL: TestRenderUnmappedCapabilityFails (0.00s)
    --- PASS: TestRenderUnmappedCapabilityFails/unmapped_tools_token_names_file_and_token (0.00s)
    --- FAIL: TestRenderUnmappedCapabilityFails/unterminated_marker_names_file (0.00s)
    --- FAIL: TestRenderUnmappedCapabilityFails/nested_marker_names_file (0.00s)
    --- FAIL: TestRenderUnmappedCapabilityFails/no_partial_output (0.00s)
--- FAIL: TestRenderViewMarkerRules (0.00s)
--- FAIL: TestSourceSetIsOneToOne (0.00s)
    --- FAIL: TestSourceSetIsOneToOne/twelve_sources_twelve_definitions (0.00s)
    --- FAIL: TestSourceSetIsOneToOne/rendered_file_without_source_fails (0.00s)
--- FAIL: TestEmitCheckFailsOnHandEdit (0.01s)
--- FAIL: TestUpdateThenCheckRoundTrip (0.00s)
--- FAIL: TestCheckReportsMissingAndDifferingFiles (0.00s)
--- FAIL: TestLayoutFailsClosed (0.00s)
--- FAIL: TestDefaultLayoutRealTreeHasNoDrift (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template/agentemit	0.392s
FAIL
```

```text
$ go test ./internal/harness/rosterguard -count=1 -v -run '^TestNumeralSourceDirDischarged$'
=== RUN   TestNumeralSourceDirDischarged
=== RUN   TestNumeralSourceDirDischarged/source_manager_design_discharged
    numeral_source_test.go:39: the numeral layer reports no hit in internal/template/agentemit/source/manager-design.md: the roster count claim of the copied definition is not being measured
=== RUN   TestNumeralSourceDirDischarged/no_undischarged_hit_under_source_dir
    numeral_source_test.go:54: empty breadth set under the agent source directory: the guard measured nothing
--- FAIL: TestNumeralSourceDirDischarged (27.76s)
    --- FAIL: TestNumeralSourceDirDischarged/source_manager_design_discharged (0.00s)
    --- FAIL: TestNumeralSourceDirDischarged/no_undischarged_hit_under_source_dir (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness/rosterguard	28.052s
FAIL
```

#### M1 acceptance cells (AC-AC4-001..004, 008..011(a)) at the M1 commit

```text
$ go test ./internal/template/agentemit -run '^TestSourceSetIsOneToOne$' -v -count=1
=== RUN   TestSourceSetIsOneToOne
=== RUN   TestSourceSetIsOneToOne/twelve_sources_twelve_definitions
=== RUN   TestSourceSetIsOneToOne/rendered_file_without_source_fails
--- PASS: TestSourceSetIsOneToOne (0.02s)
    --- PASS: TestSourceSetIsOneToOne/twelve_sources_twelve_definitions (0.00s)
    --- PASS: TestSourceSetIsOneToOne/rendered_file_without_source_fails (0.02s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.229s
exit=0
$ go test ./internal/template/agentemit -run '^TestClaudeRenderFromSource$' -v -count=1
=== RUN   TestClaudeRenderFromSource
=== RUN   TestClaudeRenderFromSource/claude_only_block_present_in_c2
=== RUN   TestClaudeRenderFromSource/codex_only_block_absent_from_c2
=== RUN   TestClaudeRenderFromSource/render_is_deterministic
--- PASS: TestClaudeRenderFromSource (0.00s)
    --- PASS: TestClaudeRenderFromSource/claude_only_block_present_in_c2 (0.00s)
    --- PASS: TestClaudeRenderFromSource/codex_only_block_absent_from_c2 (0.00s)
    --- PASS: TestClaudeRenderFromSource/render_is_deterministic (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.232s
exit=0
$ go test ./internal/template/agentemit -run '^TestCodexRenderFromSource$' -v -count=1
=== RUN   TestCodexRenderFromSource
=== RUN   TestCodexRenderFromSource/codex_only_block_present_in_c3
=== RUN   TestCodexRenderFromSource/claude_only_block_absent_from_c3
=== RUN   TestCodexRenderFromSource/c3_rendered_without_c2_bytes
=== RUN   TestCodexRenderFromSource/contract_rows_unchanged_for_twelve
--- PASS: TestCodexRenderFromSource (0.01s)
    --- PASS: TestCodexRenderFromSource/codex_only_block_present_in_c3 (0.00s)
    --- PASS: TestCodexRenderFromSource/claude_only_block_absent_from_c3 (0.00s)
    --- PASS: TestCodexRenderFromSource/c3_rendered_without_c2_bytes (0.00s)
    --- PASS: TestCodexRenderFromSource/contract_rows_unchanged_for_twelve (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.230s
exit=0
$ go test ./internal/template/agentemit -run '^TestRenderUnmappedCapabilityFails$' -v -count=1
=== RUN   TestRenderUnmappedCapabilityFails
=== RUN   TestRenderUnmappedCapabilityFails/unmapped_tools_token_names_file_and_token
=== RUN   TestRenderUnmappedCapabilityFails/unterminated_marker_names_file
=== RUN   TestRenderUnmappedCapabilityFails/nested_marker_names_file
=== RUN   TestRenderUnmappedCapabilityFails/no_partial_output
--- PASS: TestRenderUnmappedCapabilityFails (0.01s)
    --- PASS: TestRenderUnmappedCapabilityFails/unmapped_tools_token_names_file_and_token (0.00s)
    --- PASS: TestRenderUnmappedCapabilityFails/unterminated_marker_names_file (0.00s)
    --- PASS: TestRenderUnmappedCapabilityFails/nested_marker_names_file (0.00s)
    --- PASS: TestRenderUnmappedCapabilityFails/no_partial_output (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.225s
exit=0
$ go test ./internal/template/agentemit -run '^TestEmitCheckFailsOnHandEdit$' -v -count=1
=== RUN   TestEmitCheckFailsOnHandEdit
=== RUN   TestEmitCheckFailsOnHandEdit/files_unmodified_after_pass
=== RUN   TestEmitCheckFailsOnHandEdit/claude_template
=== RUN   TestEmitCheckFailsOnHandEdit/codex_template
=== RUN   TestEmitCheckFailsOnHandEdit/files_unmodified_after_fail
--- PASS: TestEmitCheckFailsOnHandEdit (0.11s)
    --- PASS: TestEmitCheckFailsOnHandEdit/files_unmodified_after_pass (0.00s)
    --- PASS: TestEmitCheckFailsOnHandEdit/claude_template (0.02s)
    --- PASS: TestEmitCheckFailsOnHandEdit/codex_template (0.02s)
    --- PASS: TestEmitCheckFailsOnHandEdit/files_unmodified_after_fail (0.02s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.327s
exit=0
$ make agents-emit-check
agents-emit-check: specs=12 copies=2 diff=0
exit=0
$ go test ./internal/harness/rosterguard -run '^TestNumeralSourceDirDischarged$' -v -count=1
=== RUN   TestNumeralSourceDirDischarged
=== RUN   TestNumeralSourceDirDischarged/source_manager_design_discharged
=== RUN   TestNumeralSourceDirDischarged/no_undischarged_hit_under_source_dir
--- PASS: TestNumeralSourceDirDischarged (41.45s)
    --- PASS: TestNumeralSourceDirDischarged/source_manager_design_discharged (0.00s)
    --- PASS: TestNumeralSourceDirDischarged/no_undischarged_hit_under_source_dir (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness/rosterguard	41.762s
exit=0
$ go test ./internal/template/agentemit -run '^(TestGoldenCommittedArtifactsMatchEmission|TestGoldenMarkdownMismatchDetected)$' -v -count=1
=== RUN   TestGoldenCommittedArtifactsMatchEmission
=== RUN   TestGoldenCommittedArtifactsMatchEmission/markdown_equals_fresh_render
--- PASS: TestGoldenCommittedArtifactsMatchEmission (0.02s)
    --- PASS: TestGoldenCommittedArtifactsMatchEmission/markdown_equals_fresh_render (0.00s)
=== RUN   TestGoldenMarkdownMismatchDetected
=== RUN   TestGoldenMarkdownMismatchDetected/injected_mismatch_is_red
--- PASS: TestGoldenMarkdownMismatchDetected (0.01s)
    --- PASS: TestGoldenMarkdownMismatchDetected/injected_mismatch_is_red (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.248s
exit=0
```

#### AC-AC4-011(b) one-time cell (M1)

`go test ./internal/template/agentemit -list '^Test'` (exit 0) lists 43 tests; each of the 30 names recorded in
`m0-baseline.md` §5 is present (loop `grep -qx` over the 30 names printed no `MISSING` line).


## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## Binding run conditions

From plan-audit iter7 (PASS-WITH-DEBT 0.90 at `c3e4ef3751`, plan_artifact_hash `dcda0ca6cc95b49410c34d2b86d441d5f32f21a8013b656926bea2ea8aaed860`). These are obligations of the run phase, not of this plan commit:

1. **D2** — `acceptance.md` L35, AC-AC4-012 cell (e) names a check but no command: supply the command (the whole-line comparison of the spec §C.2 new lines with the file). AC-AC4-012(a) is unclassified in the header at L13-L15: classify it as a permanent release-blocking test.
2. **D3** — `spec.md` L171 (the new Source row, "The only hand-edited layer") contradicts the L166 C1 row; qualify it with "for the generated agents".
3. **D4** — `progress.md` L13/L19 stale text ("012(b-d)", revision 0.5.0): corrected in this same commit, because progress.md is outside the plan-artifact hash.
4. **D5** — subtests named for a literal count of twelve (`acceptance.md` L24, L26: `twelve_sources_twelve_definitions`, `contract_rows_unchanged_for_twelve`) must not assert a literal 12, so they survive piece 2 retiring the eight.
5. Edits to `acceptance.md` / `spec.md` for items 1, 2 and 4 change the plan-artifact hash; they are made in a run-phase manager-spec re-delegation with a re-pinned hash recorded here. Items on `progress.md` do not move it.

## §G Operator verdicts

`§F` is reserved for the Phase 4 Mode Selection log (`spec-frontmatter-schema.md` § progress.md Section
Map), so verdicts live under `§G`. Not hashed into the plan-artifact set.

- Q4 and Q17: decided by the operator, board `d-20261018T062320Z-5952` (2026-10-18, scope card:t1649) — agent copies are generated outputs; the agent source is the only edit surface; C2 rendered from it; C1 byte-identical to C2 for generated agents; hand-editing C2 fails `make build`; `AGENTS.local.md` §2.0 contrary sentences superseded and rewritten by this work.
- Q20: decided by the leader under operator delegation, board `d-20261018T070809Z-924f` — C1 = C2 for the carried four in piece 1, the remaining agents when piece 2 absorbs or retires them (sequencing, not narrowing).

## §G Override and Refusal Record

- 2026-10-18T07:59:46Z SPEC-AGENT-CORE4-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 7 >= tier ceiling 3 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit.md
- 2026-10-18T08:01:02Z SPEC-AGENT-CORE4-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 7 >= tier ceiling 3 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit.md
- 2026-10-18T08:09:58Z SPEC-AGENT-CORE4-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 8 >= tier ceiling 3); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter8.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit.md
- 2026-10-18T08:13:01Z SPEC-AGENT-CORE4-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 8 >= tier ceiling 3 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit-iter8.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1649/.moai/reports/t1649/plan-audit.md

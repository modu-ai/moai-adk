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


### M2 (carried four, project copies, section 2.0)

M2 = `b1523ea6de38e70de6131c601e78099b3dbdd802` (`feat(SPEC-AGENT-CORE4-001): M2 carried four lines, project copies, section 2.0 rewrite (card t1649)`;
message cites board `d-20261018T062320Z-5952`).

#### M2 RED, before the pairs were applied and before the project sets were added to the layout

```text
$ go test ./internal/template/agentemit -count=1 -v -run '^(TestCarriedFourCopiesEqual|TestDefaultLayoutGeneratesCarriedFourProjectSets)$'
    carried_test.go:41: .claude/agents/moai/builder-harness.md differs from internal/template/templates/.claude/agents/moai/builder-harness.md
    carried_test.go:41: .codex/agents/moai/builder-harness.toml differs from internal/template/templates/.codex/agents/moai/builder-harness.toml
    carried_test.go:41: .claude/agents/moai/e2e-tester.md differs from internal/template/templates/.claude/agents/moai/e2e-tester.md
    carried_test.go:41: .codex/agents/moai/e2e-tester.toml differs from internal/template/templates/.codex/agents/moai/e2e-tester.toml
    carried_test.go:41: .claude/agents/moai/manager-design.md differs from internal/template/templates/.claude/agents/moai/manager-design.md
    carried_test.go:41: .codex/agents/moai/manager-design.toml differs from internal/template/templates/.codex/agents/moai/manager-design.toml
    carried_test.go:41: .claude/agents/moai/super-advisor.md differs from internal/template/templates/.claude/agents/moai/super-advisor.md
    carried_test.go:41: .codex/agents/moai/super-advisor.toml differs from internal/template/templates/.codex/agents/moai/super-advisor.toml
--- FAIL: TestCarriedFourCopiesEqual (0.00s)
    --- FAIL: TestCarriedFourCopiesEqual/builder-harness (0.00s)
        --- FAIL: TestCarriedFourCopiesEqual/builder-harness/c1_equals_c2 (0.00s)
        --- FAIL: TestCarriedFourCopiesEqual/builder-harness/c3p_equals_c3 (0.00s)
    --- FAIL: TestCarriedFourCopiesEqual/e2e-tester (0.00s)
        --- FAIL: TestCarriedFourCopiesEqual/e2e-tester/c1_equals_c2 (0.00s)
        --- FAIL: TestCarriedFourCopiesEqual/e2e-tester/c3p_equals_c3 (0.00s)
    --- FAIL: TestCarriedFourCopiesEqual/manager-design (0.00s)
        --- FAIL: TestCarriedFourCopiesEqual/manager-design/c1_equals_c2 (0.00s)
        --- FAIL: TestCarriedFourCopiesEqual/manager-design/c3p_equals_c3 (0.00s)
    --- FAIL: TestCarriedFourCopiesEqual/super-advisor (0.00s)
        --- FAIL: TestCarriedFourCopiesEqual/super-advisor/c1_equals_c2 (0.00s)
        --- FAIL: TestCarriedFourCopiesEqual/super-advisor/c3p_equals_c3 (0.00s)
    carried_test.go:61: layout has 2 sets, want 4
--- FAIL: TestDefaultLayoutGeneratesCarriedFourProjectSets (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template/agentemit	0.416s
FAIL
```

```text
$ go test ./internal/template -count=1 -v -run '^TestAgentsLocalSection20Text$'
=== RUN   TestAgentsLocalSection20Text
=== RUN   TestAgentsLocalSection20Text/section_2_0_exists
=== RUN   TestAgentsLocalSection20Text/old_phrases_absent_from_section_2_0
    agents_local_section20_test.go:94: §2.0 still carries the superseded phrase "Local dogfooding, hand-edited"
    agents_local_section20_test.go:94: §2.0 still carries the superseded phrase "Distribution mirror, hand-edited"
    agents_local_section20_test.go:94: §2.0 still carries the superseded phrase "the third is not edited by hand"
    agents_local_section20_test.go:94: §2.0 still carries the superseded phrase "only one of them is not a hand-edit target"
    agents_local_section20_test.go:94: §2.0 still carries the superseded phrase "not** a byte-identical"
    agents_local_section20_test.go:94: §2.0 still carries the superseded phrase "fix C2 and regenerate"
    agents_local_section20_test.go:94: §2.0 still carries the superseded phrase "Machine-emitted from C2"
    agents_local_section20_test.go:94: §2.0 still carries the superseded phrase "emitted from `.md`"
=== RUN   TestAgentsLocalSection20Text/source_path_named
    agents_local_section20_test.go:102: §2.0 does not name the agent source directory "internal/template/agentemit/source/"
=== RUN   TestAgentsLocalSection20Text/detector_flags_superseded_rows
--- FAIL: TestAgentsLocalSection20Text (0.00s)
    --- PASS: TestAgentsLocalSection20Text/section_2_0_exists (0.00s)
    --- FAIL: TestAgentsLocalSection20Text/old_phrases_absent_from_section_2_0 (0.00s)
    --- FAIL: TestAgentsLocalSection20Text/source_path_named (0.00s)
    --- PASS: TestAgentsLocalSection20Text/detector_flags_superseded_rows (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template	0.461s
FAIL
```

Before M2 the project subtests `claude_project` and `codex_project` of `TestEmitCheckFailsOnHandEdit` did not exist (the layout
had two sets), so their PASS lines were absent; they appear in the M2 cell below.

#### Extra M2 measurements (helper scripts live in the lane scratchpad, not in the repository)

```text
# AC-AC4-006 before M2: hunks of git diff -U0 --no-color <M0> <M1> over the four definitions
$ git diff -U0 --no-color 040a8afd4811d93838421a3c922ae50f1c1521e8 a252734aceca1879ab8338f26f19ee5512074a97 -- <the four C2 definitions> | grep -c '^@@'
0
exit=1

# acceptance.md section 3 pins (sha256 of the four C2 and C3 files against the table rows)
PIN MATCH builder-harness C2=7ce873448959908ab0c0a1b3fd06aabfae34853f3b0e684ee70cdb6cf7c2b619 C3=48affeb070cef387e580ab76af7bda697f0933210bd464e408257490ba34b734
PIN MATCH e2e-tester C2=3606c7fdc85b2390a329f036daa6e98787e565a3ff5fa4973894f5f51c6ccd20 C3=340901c6f887ab01655ab1ddcacadc36b8b72c94f19a246bb1adc4d1da8ef770
PIN MATCH manager-design C2=84e533d3b2ac6878e59237a0c6a43bf1b56f585172005d9c17bdcb2206b9dab6 C3=c187776dbab38ff0e4e280e68bb01b9b099c05f84209dbddce258a4ed039b644
PIN MATCH super-advisor C2=d765c4af741afc06e8280098ce15521205abc34cc116f0ef13c2c14b186fc7d1 C3=9fba77a1162a3c61525e043bbd337627f7fcb31d0f5a04bb58ef9db32729b824

# project copies equal template copies (cmp, exit 0 means identical)
cmp C1 C2 builder-harness exit=0
cmp C3p C3 builder-harness exit=0
cmp C1 C2 e2e-tester exit=0
cmp C3p C3 e2e-tester exit=0
cmp C1 C2 manager-design exit=0
cmp C3p C3 manager-design exit=0
cmp C1 C2 super-advisor exit=0
cmp C3p C3 super-advisor exit=0
```

`grep -c` exits 1 when it counts 0 lines, so `exit=1` on the first measurement is the zero-hunk result, not an error.

#### AC-AC4-006 and AC-AC4-012(b)-(e) one-time cells (run at the M2 commit)

The comparison helper `check_pairs.py` reads a `git diff -U0 --no-color` on stdin and compares its removed and added lines with the
`-` / `+` lines of the named SPEC section, in order. It was shown to fail on a known-bad input: an AC-006 diff with one added line
altered printed `FAIL added lines equal the spec new lines, in order` / `RESULT FAIL`, and an empty diff printed `RESULT FAIL`.
For AC-AC4-012(c) the diff has 6 hunks because the adjacent edited lines 97-99 form one hunk under `-U0`; the cell compares lines,
not hunk count (a first run of the checker that expected 8 hunks there was a wrong expectation of mine and is not pasted).
The command for cell (e), which the plan left without one (debt D2), is the inline `python3` block of the cell below: it reads every
`+ ` line of spec §C.2 and counts its whole-line occurrences in `AGENTS.local.md`.

```text
M0=040a8afd4811d93838421a3c922ae50f1c1521e8
M2=b1523ea6de38e70de6131c601e78099b3dbdd802

# AC-AC4-006: git diff -U0 --no-color <M0> HEAD -- the four template Markdown definitions
exit=0 (diff lines:       40)
diff --git a/internal/template/templates/.claude/agents/moai/builder-harness.md b/internal/template/templates/.claude/agents/moai/builder-harness.md
index becfbbe3f0..bfb4212ffe 100644
--- a/internal/template/templates/.claude/agents/moai/builder-harness.md
+++ b/internal/template/templates/.claude/agents/moai/builder-harness.md
@@ -38 +38 @@ Create standards-compliant Claude Code artifacts (agents, skills, plugins, comma
-| `agent` | `.claude/agents/` or `~/.claude/agents/` (with `--moai`) | Frontmatter: name, description, tools, permissionMode, memory, skills (no `model` / `effort`) |
+| `agent` | `.claude/agents/` or `.claude/agents/moai/` (with `--moai`) | Frontmatter: name, description, tools, permissionMode, memory, skills (no `model` / `effort`) |
@@ -155 +155 @@ one-line "Model/effort escalation" cross-reference at body tail — see
-`~/.claude/agents/manager-spec.md` for the verbatim line:
+`.claude/agents/moai/manager-spec.md` for the verbatim line:
diff --git a/internal/template/templates/.claude/agents/moai/e2e-tester.md b/internal/template/templates/.claude/agents/moai/e2e-tester.md
index 0a5f4f61f3..c74bc7fa0a 100644
--- a/internal/template/templates/.claude/agents/moai/e2e-tester.md
+++ b/internal/template/templates/.claude/agents/moai/e2e-tester.md
@@ -92 +92 @@ A subagent never invokes Aside (the `aside` CLI or the Aside MCP tools); a task
-Native desktop toolkits (AppKit, WinUI/Win32, Qt, GTK) are automated through the host OS accessibility layer. The per-OS recipes (macOS / Windows / Linux — defaults, fallbacks, install and probe commands, permission prerequisites) and the desktop-native evidence-source + token-cost ordering live in `~/.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md`; load that reference before any desktop-native work (§ Conditional Skill Loading). Only the recipe matching the HOST OS is probed and executed — state a host-OS/target-OS mismatch in the report instead of probing it. Scripts and flows live under `e2e/desktop-native/`; AX-tree snapshots and run logs ride the existing `e2e/.runs/` timestamped-log convention.
+Native desktop toolkits (AppKit, WinUI/Win32, Qt, GTK) are automated through the host OS accessibility layer. The per-OS recipes (macOS / Windows / Linux — defaults, fallbacks, install and probe commands, permission prerequisites) and the desktop-native evidence-source + token-cost ordering live in `.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md`; load that reference before any desktop-native work (§ Conditional Skill Loading). Only the recipe matching the HOST OS is probed and executed — state a host-OS/target-OS mismatch in the report instead of probing it. Scripts and flows live under `e2e/desktop-native/`; AX-tree snapshots and run logs ride the existing `e2e/.runs/` timestamped-log convention.
@@ -143 +143 @@ Static `skills:` preload is kept to a minimum (token diet — progressive disclo
-- When the detected project type is `desktop-native`, read `~/~/.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md` for the per-OS accessibility recipes before probing.
+- When the detected project type is `desktop-native`, read `.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md` for the per-OS accessibility recipes before probing.
diff --git a/internal/template/templates/.claude/agents/moai/manager-design.md b/internal/template/templates/.claude/agents/moai/manager-design.md
index de86e8a435..918976d4b7 100644
--- a/internal/template/templates/.claude/agents/moai/manager-design.md
+++ b/internal/template/templates/.claude/agents/moai/manager-design.md
@@ -50 +50 @@ The full D1-D5 prose lives in the workflow skill
-`~/.claude/skills/moai/workflows/design.md` (D1-D5 step headings). Summary:
+`.claude/skills/moai/workflows/design.md` (D1-D5 step headings). Summary:
@@ -203 +203 @@ Static `skills:` preload is kept to a minimum (token diet — progressive disclo
-- **Design pipeline skill**: `~/.claude/skills/moai/workflows/design.md` (D1-D5).
+- **Design pipeline skill**: `.claude/skills/moai/workflows/design.md` (D1-D5).
@@ -206 +206 @@ Static `skills:` preload is kept to a minimum (token diet — progressive disclo
-- **Agent catalog**: `CLAUDE.md` § 4 (13 retained agents — manager-design is row 6 of the Selection Decision Tree).
+- **Agent catalog**: `.moai/config/sections/delegation.yaml` (13 retained agents — manager-design is row 6 of the Selection Decision Tree).
diff --git a/internal/template/templates/.claude/agents/moai/super-advisor.md b/internal/template/templates/.claude/agents/moai/super-advisor.md
index 0e8e8ff111..764f259819 100644
--- a/internal/template/templates/.claude/agents/moai/super-advisor.md
+++ b/internal/template/templates/.claude/agents/moai/super-advisor.md
@@ -129 +129 @@ Static `skills:` preload is kept to a minimum (token diet — progressive disclo
-- Advisor/Evaluator separation: `~/.claude/agents/{plan-auditor,sync-auditor}.md`.
+- Advisor/Evaluator separation: `.claude/agents/moai/{plan-auditor,sync-auditor}.md`.

# AC-AC4-006 comparison against the 8 pairs of spec section C.1 (helper: check_pairs.py)
ok   hunks in diff = 8
ok   removed lines: spec has 8, diff has 8
ok   added lines: spec has 8, diff has 8
ok   removed lines equal the spec old lines, in order
ok   added lines equal the spec new lines, in order
ok   no added line contains '~/'
RESULT PASS
exit=0

# AC-AC4-012(b): git diff --numstat <M2>^ <M2> -- AGENTS.local.md
9^I8^IAGENTS.local.md$
exit=0

# AC-AC4-012(c): git diff -U0 --no-color <M2>^ <M2> -- AGENTS.local.md
exit=0 (diff lines:       27)
ok   hunks in diff = 6
ok   removed lines: spec has 8, diff has 8
ok   added lines: spec has 9, diff has 9
ok   removed lines equal the spec old lines, in order
ok   added lines equal the spec new lines, in order
RESULT PASS
exit=0

# AC-AC4-012(d): git log -1 --format=%B <M2> | grep -c 'd-20261018T062320Z-5952'
1
exit=0

# AC-AC4-012(e): whole-line presence of every new line of spec section C.2 in AGENTS.local.md
new line 1: whole-line occurrences in AGENTS.local.md = 1
new line 2: whole-line occurrences in AGENTS.local.md = 1
new line 3: whole-line occurrences in AGENTS.local.md = 1
new line 4: whole-line occurrences in AGENTS.local.md = 1
new line 5: whole-line occurrences in AGENTS.local.md = 1
new line 6: whole-line occurrences in AGENTS.local.md = 1
new line 7: whole-line occurrences in AGENTS.local.md = 1
new line 8: whole-line occurrences in AGENTS.local.md = 1
new line 9: whole-line occurrences in AGENTS.local.md = 1
new_lines_present_exactly: PASS (9 new lines)
exit=0
```

#### M2 acceptance cells (AC-AC4-001..004, 007..011(a), 012(a)) at the M2 commit

```text
$ go test ./internal/template/agentemit -run '^TestSourceSetIsOneToOne$' -v -count=1
=== RUN   TestSourceSetIsOneToOne
=== RUN   TestSourceSetIsOneToOne/twelve_sources_twelve_definitions
=== RUN   TestSourceSetIsOneToOne/rendered_file_without_source_fails
--- PASS: TestSourceSetIsOneToOne (0.02s)
    --- PASS: TestSourceSetIsOneToOne/twelve_sources_twelve_definitions (0.00s)
    --- PASS: TestSourceSetIsOneToOne/rendered_file_without_source_fails (0.02s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.259s
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
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.241s
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
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.247s
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
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.258s
exit=0
$ go test ./internal/template/agentemit -run '^TestCarriedFourCopiesEqual$' -v -count=1
=== RUN   TestCarriedFourCopiesEqual
=== RUN   TestCarriedFourCopiesEqual/builder-harness
=== RUN   TestCarriedFourCopiesEqual/builder-harness/c1_equals_c2
=== RUN   TestCarriedFourCopiesEqual/builder-harness/c3p_equals_c3
=== RUN   TestCarriedFourCopiesEqual/e2e-tester
=== RUN   TestCarriedFourCopiesEqual/e2e-tester/c1_equals_c2
=== RUN   TestCarriedFourCopiesEqual/e2e-tester/c3p_equals_c3
=== RUN   TestCarriedFourCopiesEqual/manager-design
=== RUN   TestCarriedFourCopiesEqual/manager-design/c1_equals_c2
=== RUN   TestCarriedFourCopiesEqual/manager-design/c3p_equals_c3
=== RUN   TestCarriedFourCopiesEqual/super-advisor
=== RUN   TestCarriedFourCopiesEqual/super-advisor/c1_equals_c2
=== RUN   TestCarriedFourCopiesEqual/super-advisor/c3p_equals_c3
--- PASS: TestCarriedFourCopiesEqual (0.02s)
    --- PASS: TestCarriedFourCopiesEqual/builder-harness (0.00s)
        --- PASS: TestCarriedFourCopiesEqual/builder-harness/c1_equals_c2 (0.00s)
        --- PASS: TestCarriedFourCopiesEqual/builder-harness/c3p_equals_c3 (0.00s)
    --- PASS: TestCarriedFourCopiesEqual/e2e-tester (0.00s)
        --- PASS: TestCarriedFourCopiesEqual/e2e-tester/c1_equals_c2 (0.00s)
        --- PASS: TestCarriedFourCopiesEqual/e2e-tester/c3p_equals_c3 (0.00s)
    --- PASS: TestCarriedFourCopiesEqual/manager-design (0.00s)
        --- PASS: TestCarriedFourCopiesEqual/manager-design/c1_equals_c2 (0.00s)
        --- PASS: TestCarriedFourCopiesEqual/manager-design/c3p_equals_c3 (0.00s)
    --- PASS: TestCarriedFourCopiesEqual/super-advisor (0.01s)
        --- PASS: TestCarriedFourCopiesEqual/super-advisor/c1_equals_c2 (0.00s)
        --- PASS: TestCarriedFourCopiesEqual/super-advisor/c3p_equals_c3 (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.256s
exit=0
$ go test ./internal/template/agentemit -run '^TestEmitCheckFailsOnHandEdit$' -v -count=1
=== RUN   TestEmitCheckFailsOnHandEdit
=== RUN   TestEmitCheckFailsOnHandEdit/files_unmodified_after_pass
=== RUN   TestEmitCheckFailsOnHandEdit/claude_template
=== RUN   TestEmitCheckFailsOnHandEdit/codex_template
=== RUN   TestEmitCheckFailsOnHandEdit/claude_project
=== RUN   TestEmitCheckFailsOnHandEdit/codex_project
=== RUN   TestEmitCheckFailsOnHandEdit/files_unmodified_after_fail
--- PASS: TestEmitCheckFailsOnHandEdit (0.28s)
    --- PASS: TestEmitCheckFailsOnHandEdit/files_unmodified_after_pass (0.01s)
    --- PASS: TestEmitCheckFailsOnHandEdit/claude_template (0.03s)
    --- PASS: TestEmitCheckFailsOnHandEdit/codex_template (0.04s)
    --- PASS: TestEmitCheckFailsOnHandEdit/claude_project (0.04s)
    --- PASS: TestEmitCheckFailsOnHandEdit/codex_project (0.03s)
    --- PASS: TestEmitCheckFailsOnHandEdit/files_unmodified_after_fail (0.04s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.585s
exit=0
$ make agents-emit-check
agents-emit-check: specs=12 copies=4 diff=0
exit=0
$ go test ./internal/harness/rosterguard -run '^TestNumeralSourceDirDischarged$' -v -count=1
=== RUN   TestNumeralSourceDirDischarged
=== RUN   TestNumeralSourceDirDischarged/source_manager_design_discharged
=== RUN   TestNumeralSourceDirDischarged/no_undischarged_hit_under_source_dir
--- PASS: TestNumeralSourceDirDischarged (41.66s)
    --- PASS: TestNumeralSourceDirDischarged/source_manager_design_discharged (0.00s)
    --- PASS: TestNumeralSourceDirDischarged/no_undischarged_hit_under_source_dir (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness/rosterguard	41.826s
exit=0
$ go test ./internal/template/agentemit -run '^(TestGoldenCommittedArtifactsMatchEmission|TestGoldenMarkdownMismatchDetected)$' -v -count=1
=== RUN   TestGoldenCommittedArtifactsMatchEmission
=== RUN   TestGoldenCommittedArtifactsMatchEmission/markdown_equals_fresh_render
--- PASS: TestGoldenCommittedArtifactsMatchEmission (0.01s)
    --- PASS: TestGoldenCommittedArtifactsMatchEmission/markdown_equals_fresh_render (0.00s)
=== RUN   TestGoldenMarkdownMismatchDetected
=== RUN   TestGoldenMarkdownMismatchDetected/injected_mismatch_is_red
--- PASS: TestGoldenMarkdownMismatchDetected (0.01s)
    --- PASS: TestGoldenMarkdownMismatchDetected/injected_mismatch_is_red (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.264s
exit=0
$ go test ./internal/template -run '^TestAgentsLocalSection20Text$' -v -count=1
=== RUN   TestAgentsLocalSection20Text
=== RUN   TestAgentsLocalSection20Text/section_2_0_exists
=== RUN   TestAgentsLocalSection20Text/old_phrases_absent_from_section_2_0
=== RUN   TestAgentsLocalSection20Text/source_path_named
=== RUN   TestAgentsLocalSection20Text/detector_flags_superseded_rows
--- PASS: TestAgentsLocalSection20Text (0.00s)
    --- PASS: TestAgentsLocalSection20Text/section_2_0_exists (0.00s)
    --- PASS: TestAgentsLocalSection20Text/old_phrases_absent_from_section_2_0 (0.00s)
    --- PASS: TestAgentsLocalSection20Text/source_path_named (0.00s)
    --- PASS: TestAgentsLocalSection20Text/detector_flags_superseded_rows (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template	0.287s
exit=0
```

#### AC-AC4-011(b) one-time cell (M2)

`go test ./internal/template/agentemit -list '^Test'` (exit 0) lists 45 tests; each of the 30 names recorded in `m0-baseline.md` §5
is present (loop `grep -qx` over the 30 names printed no `MISSING` line).

#### Guards at the M2 commit

```text
$ go test ./internal/template/agentemit/... -count=1 -cover
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	1.056s	coverage: 90.1% of statements
	github.com/modu-ai/moai-adk/internal/template/agentemit/cmd/agents-emit		coverage: 0.0% of statements
exit=0
$ GOOS=windows GOARCH=amd64 go build ./...
exit=0
$ git log -p --format= -G't\.Skip\(' 040a8afd4811d93838421a3c922ae50f1c1521e8..HEAD -- '*_test.go' | wc -l
       0
exit=0
$ make embed-manifest-check commands-emit-check
ok  	github.com/modu-ai/moai-adk/internal/template/embedemit	0.315s
ok  	github.com/modu-ai/moai-adk/internal/template/commandemit	0.274s
exit=0
$ go vet ./internal/template/ ./internal/template/agentemit/... ./internal/harness/rosterguard/
exit=0
$ golangci-lint run --timeout=3m ./internal/template/agentemit/... ./internal/harness/rosterguard/... ./internal/template/
0 issues.
exit=0
$ gofmt -l internal/template/agentemit internal/harness/rosterguard internal/template/agents_local_section20_test.go
exit=0
```

`make build` (exit 0, no tracked file changed by it) ran with `agents-emit-check` as its first prerequisite; with the template
definition of super-advisor hand-edited it failed with exit 2 and the message `agents-emit drift: 1 file(s) differ, first:
internal/template/templates/.claude/agents/moai/super-advisor.md: differs from a fresh render of its agent source`; the file was restored
and the tree was clean afterwards.

#### PRESERVE list at the M2 commit

```text
$ git diff --quiet da248a38da6e5cdb1257642cb5b2c473220d0954 HEAD -- <the eight C2 and C3 files, 16 paths>
exit=0
$ git diff --quiet da248a38da6e5cdb1257642cb5b2c473220d0954 HEAD -- .moai/specs/SPEC-AGENT-CORE4-001/{spec,plan,acceptance,design,research,decision-index}.md
exit=0
$ git diff --quiet da248a38da6e5cdb1257642cb5b2c473220d0954 HEAD -- docs-site README.md README.ko.md README.ja.md README.zh.md internal/template/agentemit/agents-codex.yaml
exit=0
$ git diff --name-only da248a38da6e5cdb1257642cb5b2c473220d0954 HEAD | wc -l
      46
$ git diff --name-only da248a38da6e5cdb1257642cb5b2c473220d0954 HEAD -- ':(exclude)internal/template/agentemit/source'   (every path touched outside the source directory)
.claude/agents/moai/e2e-tester.md
.codex/agents/moai/e2e-tester.toml
.moai/specs/SPEC-AGENT-CORE4-001/m0-baseline.md
.moai/specs/SPEC-AGENT-CORE4-001/progress.md
AGENTS.local.md
Makefile
internal/harness/rosterguard/numeral_source_test.go
internal/harness/rosterguard/registry.go
internal/template/agentemit/agentemit.go
internal/template/agentemit/agentemit_test.go
internal/template/agentemit/audit_role_exception_test.go
internal/template/agentemit/carried_test.go
internal/template/agentemit/cli_test.go
internal/template/agentemit/cmd/agents-emit/main.go
internal/template/agentemit/emit.go
internal/template/agentemit/golden_test.go
internal/template/agentemit/loader.go
internal/template/agentemit/manifest.go
internal/template/agentemit/permission_contract_test.go
internal/template/agentemit/render.go
internal/template/agentemit/render_test.go
internal/template/agentemit/sets.go
internal/template/agentemit/sets_test.go
internal/template/agentemit/writer.go
internal/template/agents_local_section20_test.go
internal/template/catalog.yaml
internal/template/templates/.claude/agents/moai/builder-harness.md
internal/template/templates/.claude/agents/moai/e2e-tester.md
internal/template/templates/.claude/agents/moai/manager-design.md
internal/template/templates/.claude/agents/moai/super-advisor.md
internal/template/templates/.codex/agents/moai/builder-harness.toml
internal/template/templates/.codex/agents/moai/e2e-tester.toml
internal/template/templates/.codex/agents/moai/manager-design.toml
internal/template/templates/.codex/agents/moai/super-advisor.toml
```

### Sync-audit iter1 fix (F1, F2) — commit 252238baa8

Sync audit iter1 (`audited_sha` 378af9a37e) returned FAIL on two code findings; this fix addresses exactly those, no SPEC artifact edited (plan hash unchanged).

- **F1 (REQ-AC4-004)**: `emitAll` now runs one `parseAndClassify` step over BOTH views, so an unmapped `tools:` token or broken frontmatter YAML inside a `claude-only` block fails closed naming file and token. Regression subtests in `TestRenderUnmappedCapabilityFails`: `claude_view_unmapped_tools_token_names_file_and_token`, `claude_view_broken_yaml_names_file`, `claude_view_defect_leaves_no_output_and_check_fails`. RED observed before the change (all three returned nil errors, Update/Check passed and wrote output).
- **F2**: `Update` refuses a symlinked or non-regular target and a symlinked or non-directory parent below the root, before the first write, and writes via same-directory temp file plus rename. Regression test `TestUpdateRefusesSymlinkTargets` (8 subtests, incl. positive control). RED observed before the change (symlink file and directory targets accepted).
- Held after the fix (this tree): `make agents-emit-check` prints `specs=12 copies=4 diff=0`; all 8 sha256 pins of acceptance.md section 3 match; `go test -cover ./internal/template/agentemit/` 90.6%; vet, golangci-lint, gofmt, `go build ./...` and the windows/amd64 build clean. F3 (D3 wording) is not touched by this fix.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-18
run_commit_sha: b1523ea6de38e70de6131c601e78099b3dbdd802   # M2, the last milestone commit; later commits only record evidence in this file
run_status: complete-with-recorded-debt
m0_commit_sha: 040a8afd4811d93838421a3c922ae50f1c1521e8
m1_commit_sha: a252734aceca1879ab8338f26f19ee5512074a97
m2_commit_sha: b1523ea6de38e70de6131c601e78099b3dbdd802
ac_pass_count: 12
ac_fail_count: 0
preserve_list_post_run_count: 16 shipped C2/C3 files of the eight, 6 plan artifacts, docs-site, READMEs and agents-codex.yaml all byte-unchanged (see the PRESERVE cell in section E.2)
l44_pre_commit_fetch: not run; the lane holds the push and landing, no push was made and local main was already absorbed at the baseline
l44_post_push_fetch: not applicable; no push
new_warnings_or_lints_introduced: 0 (golangci-lint on the touched packages printed "0 issues.")
cross_platform_build: GOOS=windows GOARCH=amd64 go build ./... exit 0; go build ./... exit 0
total_run_phase_files: 46 paths changed between the baseline da248a38da and M2 (12 of them new agent sources), plus the progress.md evidence commits
m1_to_mN_commit_strategy: M0 alone, M1, evidence, M2, evidence; one landing for M1 and M2 together (plan.md section B), no merge made here
recorded_debt: spec.md draft to in-progress not written (it would change the audited plan hash); D2, D3, D5 text defects and the plan.md Q19 open wording left for a manager-spec re-pin
```

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready
sync_complete_at: 2026-10-18
audit_ready: true
sync_commit_sha: d575238c314bbbda1ad8163e4afc66fb1569fcda

Sync close (manager-docs, 2026-10-18, branch `WT-3-3-10`, worktree `.moai/worktrees/t1649`, run head `a702bf117c`): the single sync commit carries (a) the CHANGELOG `[Unreleased]` → `### Added` entry, first in the list, (b) the `spec.md` frontmatter transition `in-progress → implemented → completed` merged on this one commit (`status:` only; `updated: 2026-10-18` already read today), and (c) this §E.4 with the canonical `pending-backfill` placeholder, backfilled by the following commit (a commit cannot cite its own SHA). No code, template file, `AGENTS.local.md`, `m0-baseline.md`, docs-site or README file is touched.

### Sync evidence (this run, this tree)

- B12 pre-check: `grep -c 'SPEC-AGENT-CORE4-001' CHANGELOG.md` → `0` before emission (no duplicate), `1` after.
- B12 AC-count self-test (`ac_source=.moai/specs/SPEC-AGENT-CORE4-001/acceptance.md`, tier L, source resolved and non-empty): the reserved-token counter printed `13` with `live=13 excluded=0 ambiguous=0`, rc 0. The CHANGELOG states **12**, the SPEC's own bijection (`acceptance.md` §1 "12 rows = 12 AC", REQ-AC4-001..012 ↔ AC-AC4-001..012). The difference is one stray identifier, `AC-001`, inside the AC-AC4-011 row (line 34: "the block from `AC-001's .md face`"), which names a test label of the baseline golden test, not a criterion of this SPEC. Closed after the sync commit by `manager-spec` commit `a0f52732f5` (`acceptance.md` only): the counter now prints `12` with `live=12 excluded=1`; manager-docs did not touch `acceptance.md`.
- `moai spec lint --strict SPEC-AGENT-CORE4-001` with a moai built from this tree by path (`go build -o <scratchpad>/moai-t1649 ./cmd/moai`, built at HEAD `a702bf117c`; the build embeds no commit, `version` prints `none`) → `0 error(s), 0 warning(s)`, rc 0; one INFO `OwnershipTransitionUnmeasured` on the earlier `draft → in-progress` commit (no `Authored-By-Agent` trailer), which never changes the exit status.
- `go run ./internal/template/agentemit/cmd/agents-emit -check` → `agents-emit-check: specs=12 copies=4 diff=0`, rc 0.
- Plan-artifact hash (ComputeHash-equivalent): `24308dc6eb61cd352cd82ec10f96701fedf08037298bf46bd72f80c5402247ac` before the sync commit, `a08250701973e8ce15997fc13a66dc8556eb8e1e9b0d6bd9f0343b14b3763515` after it; the only plan-artifact file changed is `spec.md` (`status: in-progress` → `completed`).
- Docs scope: `spec.md` §D and `plan.md` §C name docs-site and README as PRESERVE / out of scope for piece 1, so neither is touched; the CHANGELOG entry is the whole docs obligation of this piece.
- MX sync pass: not performed here (this sync task excludes code edits and no fan-in measurement was run); no `@MX` tag was added or changed. Recorded as a Gap.

### Carried debt and Gaps

- §E.3 `recorded_debt` listed the D2, D3, D5 text defects and the `plan.md` Q19 wording as left for a `manager-spec` re-pin. Commit `e461478887` ("run-phase re-pin of SPEC text debts") later changed `acceptance.md`, `design.md`, `plan.md` and `spec.md`; this sync did not re-verify each defect's closure.
- The stray `AC-001` debt is closed by `a0f52732f5` (see the B12 note above).
- Pieces 2 and 3 inherit the hand-off rows of `spec.md` §E.

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

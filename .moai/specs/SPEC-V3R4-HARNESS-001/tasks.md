# Tasks — SPEC-V3R4-HARNESS-001

This document is the task-level breakdown of the implementation plan for the foundation SPEC of the self-evolving harness v2 system. Tasks are organized by Wave (A through E) and use IDs `T-A1` through `T-E3`. Each task lists its linked REQ IDs, linked AC IDs, MX tag implications, risk/complexity rating, and Wave membership.

All priorities use P0/P1/P2/P3 labels per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation; no time estimates are used.

---

## Task Summary

| Wave | Tasks | Priority Distribution | Total |
|------|-------|-----------------------|-------|
| Wave A (Skill body and workflow consolidation) | T-A1, T-A2, T-A3, T-A4 | P0: 3, P1: 1 | 4 |
| Wave B (Subagent reorganization) | T-B1, T-B2, T-B3 | P0: 2, P1: 1 | 3 |
| Wave C (Hook contract preservation) | T-C1, T-C2, T-C3 | P1: 3 | 3 |
| Wave D (CLI deprecation markers and CI guard) | T-D1, T-D2, T-D3, T-D4 | P0: 3, P1: 1 | 4 |
| Wave E (Documentation and follow-up coordination) | T-E1, T-E2, T-E3 | P1: 3 | 3 |
| **Total** | | **P0: 8, P1: 9** | **17** |

---

## Wave A — Skill Body and Workflow Consolidation

### T-A1 (P0) — Audit `.claude/skills/moai/workflows/harness.md` for binary invocation

**Description**: Inspect the existing workflow body for any reference to `moai harness` as a binary invocation (Bash tool call, `os.Exec` style pattern, or shell command). Remove any found.

**Linked REQs**: REQ-HRN-FND-001, REQ-HRN-FND-003

**Linked ACs**: AC-HRN-FND-001, AC-HRN-FND-002

**MX tag implications**: None (read-only audit; modifications, if any, are text-only).

**Risk / Complexity**: Low. Static grep + text edit.

**Wave**: Wave A

**Definition of Done**:
- `grep -nE 'moai harness' .claude/skills/moai/workflows/harness.md` returns zero matches in invocation context.
- A finding report (one line in the PR description) confirms the audit ran.

---

### T-A2 (P0) — Document Tier-4 AskUserQuestion checkpoint in workflow body

**Description**: Add or confirm an explicit section in `.claude/skills/moai/workflows/harness.md` documenting the Tier-4 application flow with an `AskUserQuestion` checkpoint owned by the orchestrator. The section cites REQ-HRN-FND-004 and the canonical AskUserQuestion protocol.

**Linked REQs**: REQ-HRN-FND-004, REQ-HRN-FND-015

**Linked ACs**: AC-HRN-FND-003

**MX tag implications**: Consider adding `@MX:NOTE` annotation at the workflow body section header citing the orchestrator-only contract.

**Risk / Complexity**: Low. Text-only addition.

**Wave**: Wave A

**Definition of Done**:
- The workflow body contains a dedicated section (e.g., "## Tier-4 Application Gate") with an explicit `AskUserQuestion` invocation example using the canonical four-option pattern (Apply (권장), Modify, Defer, Reject).
- The section cross-references `.claude/rules/moai/core/askuser-protocol.md`.

---

### T-A3 (P0) — Add cross-reference to AskUserQuestion protocol

**Description**: Add a top-of-body annotation in `.claude/skills/moai/workflows/harness.md` linking to `.claude/rules/moai/core/askuser-protocol.md` § ToolSearch Preload Procedure and § Free-form Circumvention Prohibition.

**Linked REQs**: REQ-HRN-FND-004, REQ-HRN-FND-015

**Linked ACs**: AC-HRN-FND-003

**MX tag implications**: None.

**Risk / Complexity**: Low.

**Wave**: Wave A

**Definition of Done**:
- A "References" or "Related Rules" section exists near the top of the workflow body with a link to the AskUserQuestion protocol.

---

### T-A4 (P1) — Annotate `moai-harness-learner` SKILL.md with V3R4 cross-reference

**Description**: Add a top-of-body annotation in `.claude/skills/moai-harness-learner/SKILL.md` (one-line text addition) citing this SPEC's REQ-HRN-FND-011 (4-tier ladder preservation) and noting that the existing 4-tier ladder is preserved unchanged for the foundation SPEC.

**Linked REQs**: REQ-HRN-FND-011

**Linked ACs**: AC-HRN-FND-008

**MX tag implications**: None. Text-only annotation.

**Risk / Complexity**: Low.

**Wave**: Wave A

**Definition of Done**:
- `.claude/skills/moai-harness-learner/SKILL.md` has an annotation citing this SPEC by ID.
- No behavioral or frontmatter changes; text annotation only.

---

## Wave B — Subagent Reorganization and Contract Preservation

### T-B1 (P0) — Audit `moai-harness-learner` for AskUserQuestion calls

**Description**: Inspect the `moai-harness-learner` skill body and any associated subagent definition for `AskUserQuestion` invocations. If any are found, replace with a structured blocker-report return template.

**Linked REQs**: REQ-HRN-FND-015

**Linked ACs**: AC-HRN-FND-003

**MX tag implications**: None expected (no invocations should exist; this is a verification task).

**Risk / Complexity**: Low.

**Wave**: Wave B

**Definition of Done**:
- `grep -n 'AskUserQuestion' .claude/skills/moai-harness-learner/SKILL.md` returns zero invocation matches (references in commentary citing the protocol are acceptable).
- Audit finding documented in the PR description.

---

### T-B2 (P0) — Annotate harness-related subagents with V3R4 contract citation

**Description**: Add top-of-body annotations to `moai-harness-learner` and `moai-meta-harness` SKILL.md citing REQ-HRN-FND-015 and the orchestrator-subagent boundary from `.claude/rules/moai/core/agent-common-protocol.md`.

**Linked REQs**: REQ-HRN-FND-015

**Linked ACs**: AC-HRN-FND-003

**MX tag implications**: None.

**Risk / Complexity**: Low.

**Wave**: Wave B

**Definition of Done**:
- Both SKILL.md files contain the annotation block.
- No behavioral or frontmatter changes.

---

### T-B3 (P1) — Verify `.claude/agents/my-harness/*` definitions respect the contract

**Description**: Run a static check on the project-area `my-harness/*` agent definitions (`cli-template-specialist.md`, `hook-ci-specialist.md`, `quality-specialist.md`, `workflow-specialist.md`) to confirm they do not invoke `AskUserQuestion`.

**Linked REQs**: REQ-HRN-FND-015

**Linked ACs**: AC-HRN-FND-003

**MX tag implications**: None.

**Risk / Complexity**: Low.

**Wave**: Wave B

**Definition of Done**:
- `grep -rn 'AskUserQuestion' .claude/agents/my-harness/` returns zero invocation matches.
- If any are found, raise as a separate fix (out of scope for this SPEC but reported).

---

## Wave C — Hook Contract and Observer Baseline Preservation

### T-C1 (P1) — Inspect PostToolUse hook for `learning.enabled` gate

**Description**: Locate the PostToolUse hook implementation (likely `internal/hook/post_tool_use.go` or `.claude/hooks/moai/handle-post-tool-use.sh`) and confirm it checks `learning.enabled` before performing any read or write to `.moai/harness/usage-log.jsonl`.

**Linked REQs**: REQ-HRN-FND-009

**Linked ACs**: AC-HRN-FND-007

**MX tag implications**: Consider `@MX:NOTE` on the `learning.enabled` gate check if a high fan-in code path is involved.

**Risk / Complexity**: Medium. Requires Go code inspection.

**Wave**: Wave C

**Definition of Done**:
- The hook code path includes an explicit `learning.enabled == false` check at the top.
- If absent, an issue is raised (the fix is in scope for Wave C).

---

### T-C2 (P1) — Verify or add unit test for observer no-op behavior

**Description**: Run the existing test suite for the PostToolUse hook. If a test for the no-op behavior when `learning.enabled: false` does not exist, add one. The test should assert that `.moai/harness/usage-log.jsonl` is not modified during a representative tool invocation when learning is disabled.

**Linked REQs**: REQ-HRN-FND-009

**Linked ACs**: AC-HRN-FND-007

**MX tag implications**: None.

**Risk / Complexity**: Medium. Test authoring in Go.

**Wave**: Wave C

**Definition of Done**:
- `go test ./internal/hook/...` includes a test asserting the no-op behavior.
- The test passes locally and in CI.

---

### T-C3 (P1) — Verify observation entry schema matches REQ-HRN-FND-010

**Description**: Inspect the JSONL append code path and confirm each entry includes at minimum: ISO-8601 timestamp, event_type, subject, and context_hash.

**Linked REQs**: REQ-HRN-FND-010

**Linked ACs**: AC-HRN-FND-008

**MX tag implications**: None.

**Risk / Complexity**: Low.

**Wave**: Wave C

**Definition of Done**:
- A sample observation entry from a test run is documented in the PR description.
- The schema matches the REQ requirements.

---

## Wave D — CLI Deprecation Markers and CI Guard

### T-D1 (P0) — Add deprecation header comment to `internal/cli/harness.go`

**Description**: Add a top-of-file Go comment block annotating `internal/cli/harness.go` as deprecated per `SPEC-V3R4-HARNESS-001` and `BC-V3R4-HARNESS-001-CLI-RETIREMENT`. The comment must include the SPEC ID, the BC ID, and a "DEPRECATED" tag string for grep-based audits.

**Linked REQs**: REQ-HRN-FND-001

**Linked ACs**: AC-HRN-FND-001

**MX tag implications**: Add `@MX:NOTE` annotation citing the deprecation. The `@MX:ANCHOR` annotation already on `newHarnessCmd` (line 31 of current file) is preserved unchanged.

**Risk / Complexity**: Low. Comment-only Go file edit.

**Wave**: Wave D

**Definition of Done**:
- The top of `internal/cli/harness.go` contains:
  ```go
  // DEPRECATED: per SPEC-V3R4-HARNESS-001 (BC-V3R4-HARNESS-001-CLI-RETIREMENT).
  // The harness CLI subcommand is retired. Use /moai:harness slash command instead.
  // This file remains as a deprecation marker; physical removal is deferred to a
  // follow-up SPEC after downstream SPECs 002-008 merge.
  ```
- `go vet ./internal/cli/...` passes.
- `golangci-lint run` (if configured for harness path) does not regress.

---

### T-D2 (P0) — Write CI guard test `internal/cli/harness_retirement_test.go`

**Description**: Author a new Go test file that asserts `rootCmd.Commands()` does not contain any command with `Use: "harness"`. The test must fail with a diagnostic message referencing `SPEC-V3R4-HARNESS-001` and `REQ-HRN-FND-002`.

**Linked REQs**: REQ-HRN-FND-002

**Linked ACs**: AC-HRN-FND-001

**MX tag implications**: The new test file gets `@MX:NOTE` annotation explaining it is a regression guard for SPEC-V3R4-HARNESS-001.

**Risk / Complexity**: Medium. Requires familiarity with cobra command tree iteration.

**Wave**: Wave D

**Definition of Done**:
- The test file exists and uses table-driven test pattern per `CLAUDE.local.md` § 6 Testing Guidelines.
- `go test -run TestHarnessRetirement ./internal/cli/...` passes.
- The test's failure message explicitly references the SPEC ID and REQ ID.

---

### T-D3 (P0) — Run existing `internal/cli/harness_test.go` to confirm no regressions

**Description**: Execute the existing harness CLI test suite to ensure the `newHarnessCmd` factory still operates correctly for any internal callers that may inspect it (e.g., introspection tests). The factory remains defined for compatibility; only the cobra registration is absent.

**Linked REQs**: REQ-HRN-FND-001

**Linked ACs**: AC-HRN-FND-001

**MX tag implications**: None.

**Risk / Complexity**: Low. Test execution only.

**Wave**: Wave D

**Definition of Done**:
- `go test ./internal/cli/... -run TestHarness` passes for all pre-existing tests.
- Any test failure is reported in the PR description.

---

### T-D4 (P1) — Annotate `internal/cli/root.go` near command registration block

**Description**: Add a Go comment near the cobra `rootCmd.AddCommand(...)` block in `internal/cli/root.go` explicitly stating that `newHarnessCmd` is intentionally not registered, citing this SPEC.

**Linked REQs**: REQ-HRN-FND-001, REQ-HRN-FND-002

**Linked ACs**: AC-HRN-FND-001

**MX tag implications**: None. Comment-only.

**Risk / Complexity**: Low.

**Wave**: Wave D

**Definition of Done**:
- The comment exists in `internal/cli/root.go` near the registration block.
- Comment text: `// NOTE: newHarnessCmd is intentionally NOT registered per SPEC-V3R4-HARNESS-001 (BC-V3R4-HARNESS-001-CLI-RETIREMENT). See internal/cli/harness_retirement_test.go for the CI guard.`

---

## Wave E — Documentation and Superseded SPEC Follow-Up Coordination

### T-E1 (P1) — Draft CHANGELOG.md entry for v2.20.0

**Description**: Add a new section to `CHANGELOG.md` for the target release announcing `BC-V3R4-HARNESS-001-CLI-RETIREMENT`. The entry cites this SPEC by ID, lists the three superseded V3R3 SPECs, and points users to the slash command `/moai:harness` as the supported invocation path.

**Linked REQs**: REQ-HRN-FND-001, REQ-HRN-FND-013

**Linked ACs**: AC-HRN-FND-001, AC-HRN-FND-010

**MX tag implications**: None. Documentation file.

**Risk / Complexity**: Low.

**Wave**: Wave E

**Definition of Done**:
- CHANGELOG.md contains a `### Breaking Changes` subsection in the v2.20.0 (or target release) section.
- The entry cites `BC-V3R4-HARNESS-001-CLI-RETIREMENT` and `SPEC-V3R4-HARNESS-001`.

---

### T-E2 (P1) — Draft release notes for the target release

**Description**: Create or append to `.moai/release/RELEASE-NOTES-v2.20.0.md` (or matching target release) with a section explaining the V3R4 foundation, the CLI retirement migration story, and the enumeration of the seven downstream SPECs that will follow.

**Linked REQs**: REQ-HRN-FND-001, REQ-HRN-FND-013

**Linked ACs**: AC-HRN-FND-001, AC-HRN-FND-010

**MX tag implications**: None.

**Risk / Complexity**: Low.

**Wave**: Wave E

**Definition of Done**:
- Release notes file contains a "Self-Evolving Harness v2 Foundation" section.
- Migration story explains that the slash command continues to work and the CLI verb path returns `unknown command`.
- The seven downstream SPECs are enumerated by ID.

---

### T-E3 (P1) — Author `follow-up.md` with V3R3 status-transition instructions

**Description**: Create `.moai/specs/SPEC-V3R4-HARNESS-001/follow-up.md` documenting the post-merge V3R3 status-transition commit. The document is the spec-side input for the post-merge `manager-git` invocation.

**Linked REQs**: REQ-HRN-FND-013

**Linked ACs**: AC-HRN-FND-010

**MX tag implications**: None.

**Risk / Complexity**: Low.

**Wave**: Wave E

**Definition of Done**:
- `follow-up.md` exists with explicit instructions for `manager-git`:
  1. After this PR merges, create a new branch `chore/SPEC-V3R3-status-transition`.
  2. Update the frontmatter of `.moai/specs/SPEC-V3R3-HARNESS-001/spec.md`, `.moai/specs/SPEC-V3R3-HARNESS-LEARNING-001/spec.md`, and `.moai/specs/SPEC-V3R3-PROJECT-HARNESS-001/spec.md` with `status: superseded` and `superseded_by: SPEC-V3R4-HARNESS-001`.
  3. Add a HISTORY entry to each citing this SPEC.
  4. Create a PR to main with squash merge strategy.
- The document's instructions are unambiguous enough for `manager-git` to execute without re-interpreting.

---

## Dependency Graph (Task-Level)

```
Wave A: T-A1 -> T-A2 -> T-A3
                T-A4 (parallel with A2/A3)

Wave B: T-B1 -> T-B2
                T-B3 (parallel with B2)

Wave C: T-C1 -> T-C2 -> T-C3

Wave D: T-D1 -> T-D2 -> T-D3
                T-D4 (parallel with D2/D3)

Wave E: T-E1 (parallel with all Wave E)
        T-E2 (parallel)
        T-E3 (parallel)

Inter-Wave:
  Wave A complete -> Wave B entry
  Wave A complete -> Wave C entry (Wave B and C parallel-able after Wave A)
  Wave A, B, C all complete -> Wave D entry
  Wave D complete -> Wave E entry
```

---

## Out of Scope (Task-Level)

The following tasks are explicitly NOT in any Wave of this SPEC. They are deferred to downstream SPECs:

| Downstream SPEC | Task Domain |
|-----------------|-------------|
| SPEC-V3R4-HARNESS-002 | Adding Stop / SubagentStop / UserPromptSubmit hook handlers; unified observation schema across event types. |
| SPEC-V3R4-HARNESS-003 | Embedding model selection; embedding-cluster algorithm; replacing frequency-count classifier. |
| SPEC-V3R4-HARNESS-004 | Actor + Evaluator + Self-Reflection trio implementation; 3-iteration cap; episodic memory of reflections. |
| SPEC-V3R4-HARNESS-005 | Constitution principle parser; self-scoring rubric; pre-screen integration. |
| SPEC-V3R4-HARNESS-006 | Multi-objective scoring tuple; auto-rollback-on-regression mechanism. |
| SPEC-V3R4-HARNESS-007 | Voyager-style embedding-indexed skill library; top-K retrieval; compositional skill reuse. |
| SPEC-V3R4-HARNESS-008 | Anonymization layer; opt-in cross-project federation; namespace isolation. |
| Follow-up cleanup SPEC | Physical deletion of `internal/cli/harness.go` and `internal/cli/harness_test.go` after downstream SPECs 002-008 merge. |
| Post-merge `manager-git` commit | Transitioning the three V3R3 SPECs to `status: superseded` (instructed in `follow-up.md`). |

---

## Run-Phase Entry Point

After this plan PR merges:

1. Execute `/clear` to reset context.
2. Create a SPEC worktree: `moai worktree new SPEC-V3R4-HARNESS-001 --base origin/main`.
3. Inside the worktree, execute `/moai run SPEC-V3R4-HARNESS-001`.
4. `manager-develop` (or `manager-tdd` depending on `quality.yaml` `development_mode`) takes over execution of Waves A-E sequentially.
5. Each Wave merges as a separate squash PR per Enhanced GitHub Flow doctrine (`CLAUDE.local.md` § 18).
6. After all five Waves merge, execute `/moai sync SPEC-V3R4-HARNESS-001` to generate the final documentation sync PR.
7. After sync PR merges, `manager-git` executes the follow-up V3R3 status-transition commit per `follow-up.md`.

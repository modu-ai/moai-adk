# SPEC-V3R2-ORC-004 Research — Worktree MUST Rule for Write-Heavy Role Profiles

> Codebase analysis preceding the implementation plan. Authored on `plan/SPEC-V3R2-ORC-004` (Step 1, plan-in-main, base `origin/main` HEAD `dca57b14d`).
> Companion to `spec.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.
> Methodology: file-by-file inspection of write-heavy agent definitions, lint engine inventory, workflow.yaml audit, problem catalog cross-reference.

## HISTORY

| Version | Date       | Author        | Description |
|---------|------------|---------------|-------------|
| 0.1.0   | 2026-05-10 | manager-spec  | Initial codebase analysis. Discovered partial-pre-completion (LR-05/LR-09 already in `agent_lint.go`); recalibrated plan scope to gap-fill only. |

---

## 1. Research Goal

Identify the precise set of changes required to upgrade `isolation: worktree` from SHOULD to MUST for v3r2 write-heavy agents per `spec.md` REQ-001 .. REQ-015. Establish ground truth on:

1. Which agents currently declare `isolation: worktree` in frontmatter.
2. Whether `agent_lint.go` already enforces LR-05 / LR-09.
3. Whether `worktree-integration.md` §HARD Rules section currently uses SHOULD or MUST verbiage for standalone write-heavy agents.
4. Whether `workflow.yaml` role_profiles already align with REQ-004 (implementer/tester/designer = `worktree`; researcher/analyst/architect/reviewer = `none`).
5. Whether `manager-cycle` (REQ-002 target) exists yet — it is created by SPEC-V3R2-ORC-001.
6. Whether a Go-side spawn wrapper exists for `Agent(subagent_type:"general-purpose")` invocations (REQ-008 enforcement layer).
7. Where the canonical `ORC_WORKTREE_REQUIRED`, `ORC_WORKTREE_MISSING`, `ORC_WORKTREE_ON_READONLY` sentinel keys should surface.

The output of this research determines whether the plan is a fresh implementation, a partial-completion gap-fill, or a documentation-only effort.

---

## 2. Inventory

### 2.1 Current State of Agent Frontmatter

Inspection of `.claude/agents/moai/*.md` (25 files at HEAD `3356aa9a9`) reveals:

| Agent | tools (Write/Edit?) | permissionMode | isolation | Required? | Action |
|-------|---------------------|----------------|-----------|-----------|--------|
| `manager-cycle.md` | n/a (NOT YET CREATED) | n/a | n/a | YES (post-ORC-001) | wait for ORC-001 merge, then add `isolation: worktree` |
| `manager-ddd.md` | Yes (Write, Edit, MultiEdit) | bypassPermissions | (absent) | NO (retired by ORC-001 → manager-cycle) | leave unchanged; ORC-001 retire-stub will handle |
| `manager-tdd.md` | Yes (Write, Edit, MultiEdit) | bypassPermissions | (absent) | NO (retired by ORC-001 → manager-cycle) | leave unchanged; ORC-001 retire-stub will handle |
| `expert-backend.md` | Yes (Write, Edit) | bypassPermissions | (absent) | YES | add `isolation: worktree` |
| `expert-frontend.md` | Yes (Write, Edit) | bypassPermissions | (absent) | YES | add `isolation: worktree` |
| `expert-refactoring.md` | Yes (Write, Edit) | bypassPermissions | (absent) | YES | add `isolation: worktree` |
| `researcher.md` | Yes (Write, Edit) | acceptEdits | (absent) | YES (P-A22) | add `isolation: worktree` + body line update |
| `manager-strategy.md` | Read-only | plan | (absent) | NO (read-only) | leave unchanged; LR-09 protection |
| `manager-quality.md` | Read-only | plan (assumed) | (absent) | NO (read-only) | leave unchanged |
| `manager-docs.md` | Yes (Write) | bypassPermissions | (absent) | NO (scope-limited to docs-site/) | leave unchanged per spec.md §2.2 |
| `manager-git.md` | Yes (Bash) | bypassPermissions | (absent) | NO (operates on main worktree by design) | leave unchanged |
| `manager-project.md` | Yes (Write) | bypassPermissions | (absent) | NO (writes to .moai/project/ only) | leave unchanged |
| `manager-spec.md` | Yes (Write) | bypassPermissions | (absent) | YES if write-heavy classifier matches; spec.md §2.1 does NOT enumerate it | leave unchanged in v3r2 ORC-004 (out of scope) |
| `evaluator-active.md` | Read-only | plan | (absent) | NO | leave unchanged |
| `plan-auditor.md` | Read-only | plan | (absent) | NO | leave unchanged |
| `expert-debug.md` / `expert-performance.md` / `expert-security.md` / `expert-testing.md` / `expert-devops.md` / `expert-mobile.md` | Mixed | bypassPermissions | (absent) | NO (not enumerated in spec.md §2.1) | leave unchanged |
| `builder-agent.md` / `builder-skill.md` / `builder-plugin.md` | Yes (Write) | bypassPermissions | (absent) | NO (out of scope per spec.md §2.2 implicit) | leave unchanged |
| `manager-brain.md` / `claude-code-guide.md` | Mixed | varies | (absent) | NO (not enumerated) | leave unchanged |

Discovered: **0 agents currently declare `isolation: worktree`** in frontmatter at HEAD `3356aa9a9`.

### 2.2 Current State of `agent_lint.go`

`internal/cli/agent_lint.go` (23,398 bytes, 796+ lines) already declares LR-01 through LR-10 in the help text (lines 84–94):

```
LR-05: Error on missing isolation: worktree for write-heavy role profiles and standalone agents (SPEC-V3R2-ORC-004)
LR-09: Reject isolation: worktree on read-only agents (permissionMode: plan) (SPEC-V3R2-ORC-004)
```

Implementation inspection (lines 444–540):

- `checkMissingIsolation` (LR-05) is **fully implemented**:
  - Role-profile branch (`implementer|tester|designer|specialist`) iterates names containing those substrings and emits `LR-05` error if `isolation != "worktree"`.
  - Write-heavy standalone agent branch hard-codes `[manager-cycle, expert-backend, expert-frontend, expert-refactoring, researcher]` exactly matching `spec.md` §2.1 and emits `LR-05` error if `isolation != "worktree"`.
- `checkReadOnlyIsolation` (LR-09) is **fully implemented**:
  - Emits `LR-09` error when `permissionMode == "plan"` AND `isolation == "worktree"`.

Implication: The lint engine is already correct. **No Go code change is required for LR-05/LR-09.** The plan's lint-side work reduces to verification (running `moai agent lint` against the post-edit tree and confirming exit 0) plus adding test fixtures for AC-06/AC-07.

### 2.3 Current State of `worktree-integration.md` §HARD Rules

Inspection of `.claude/rules/moai/workflow/worktree-integration.md` lines 130–140:

```
- [HARD] Implementation teammates in team mode (role_profiles: implementer, tester, designer) MUST use `isolation: "worktree"` when spawned via Agent()
- [HARD] Read-only teammates (role_profiles: researcher, analyst, reviewer) MUST NOT use `isolation: "worktree"` — their `mode: "plan"` already prevents writes
- [HARD] One-shot sub-agents that write files (expert-backend, expert-frontend, manager-develop) SHOULD use `isolation: "worktree"` when making cross-file changes
- [HARD] GitHub workflow agents (fixer agents in /moai github issues) MUST use `isolation: "worktree"` for branch isolation
```

Findings:

- Line 133 (team-mode role profiles): already MUST. ✓
- Line 134 (read-only role profiles): already MUST NOT. ✓
- **Line 135 (standalone write-heavy)**: still **SHOULD**. ⬅ **PRIMARY EDIT TARGET (REQ-001)**.
- Line 135 mentions `manager-develop` which does not exist in the agent roster (legacy v2 reference); ORC-004 should align it to `manager-cycle` (post-ORC-001) and add `expert-refactoring`, `researcher`.
- Line 136 (GitHub workflow fixer): already MUST. ✓ (out of ORC-004 scope)

The replacement clause from REQ-001 is:

> Implementation agents that write files across 3 or more paths per invocation MUST use `isolation: worktree` in their frontmatter. This includes v3r2 agents manager-cycle, expert-backend, expert-frontend, expert-refactoring, researcher, and team-mode role profiles implementer, tester, designer.

### 2.4 Current State of `workflow.yaml` role_profiles

Inspection of `.moai/config/sections/workflow.yaml` lines 35–73:

| Role profile | mode | model | isolation | spec.md REQ-004 expected | Aligned? |
|--------------|------|-------|-----------|--------------------------|----------|
| `researcher` | plan | haiku | none | none | ✓ |
| `analyst` | plan | sonnet | none | none | ✓ |
| `architect` | plan | opus | none | none | ✓ |
| `implementer` | acceptEdits | sonnet | worktree | worktree | ✓ |
| `tester` | acceptEdits | sonnet | worktree | worktree | ✓ |
| `designer` | acceptEdits | sonnet | worktree | worktree | ✓ |
| `reviewer` | plan | sonnet | none | none | ✓ |

**workflow.yaml is already 100% aligned with REQ-004.** No edit is required. The plan's verification-only task is to add a CI fixture under `internal/template/templates/.moai/config/sections/workflow.yaml` parity check (REQ-005 template-first); the parity is already in place.

### 2.5 manager-cycle Existence Check

`ls .claude/agents/moai/manager-cycle.md` returns no such file. **manager-cycle is created by SPEC-V3R2-ORC-001.** ORC-004 has SPEC-V3R2-ORC-001 in its `dependencies:` block (`spec.md` §9.1), so this is expected. ORC-004's frontmatter edit on manager-cycle proceeds only after ORC-001 has merged and the file exists.

If ORC-001 is not yet merged at the time of ORC-004's run phase, the implementer pauses at the manager-cycle edit task and reports a blocker. (Not anticipated to occur — ORC-001 is in `Sprint 9` per `project_v3_master_plan_post_v214.md`.)

### 2.6 Spawn Wrapper Inventory

`grep -rn "Agent(subagent_type" internal/` returns only:

- `internal/runtime/audit_gate.go:99` — comment-only mention
- Various template files in `internal/template/templates/` — documentation-only

There is **no Go-side spawn wrapper that intercepts `Agent()` calls** at runtime. Claude Code's `Agent()` tool is invoked from skill bodies (markdown), not from compiled Go.

This significantly affects REQ-008 implementation:

> When the MoAI orchestrator invokes `Agent(subagent_type: "general-purpose")` with a role_profile parameter, the spawn wrapper shall verify that the override parameter set contains `isolation: "worktree"` if the role_profile is implementer, tester, or designer; a mismatch shall result in a structured blocker report to the orchestrator (`ORC_WORKTREE_REQUIRED`).

Possible enforcement layers:

1. **Agent-lint over workflow.yaml** (chosen — see plan §1.3): Add a new lint command `moai workflow lint` that validates `role_profiles` keys; rejects if any of {implementer, tester, designer} has `isolation != "worktree"`. This is a static CI gate, not runtime.
2. **Skill body convention** (supplementary): Update `.claude/skills/moai/team/run.md` to instruct the orchestrator to never override `isolation` away from `worktree` for these role profiles. This is documentation, not enforcement.
3. **Hook-based runtime gate** (deferred): A `SubagentStart` hook that inspects spawn parameters and emits `ORC_WORKTREE_REQUIRED` blocker via stderr on mismatch. Out of ORC-004 scope (deferred to a future SPEC if hook telemetry shows real-world drift).

The plan adopts (1) + (2). Option (3) is documented as future work in `plan.md` §8 risks.

---

## 3. Pre-Completion Discovery

The most consequential research finding: **two of the four spec.md §2.1 deliverables are already implemented at HEAD `3356aa9a9`** by an unrelated SPEC's incidental work.

| Deliverable | Status | Evidence |
|-------------|--------|----------|
| LR-05 lint rule promoted to error for write-heavy classifier | ✓ DONE | `agent_lint.go:444-491`, exact agent list match |
| LR-09 lint rule rejecting `isolation: worktree` on read-only agents | ✓ DONE | `agent_lint.go:520-535` |
| Add `isolation: worktree` to 4 existing agents (expert-backend, expert-frontend, expert-refactoring, researcher) | ⬜ PENDING | grep returns 0 matches |
| Update `worktree-integration.md` §HARD Rules line 135 SHOULD → MUST | ⬜ PENDING | line 135 still says SHOULD |
| Update researcher body line 49 ("when possible" → definitive) | ⬜ PENDING | researcher.md body inspection |
| Add `manager-cycle` `isolation: worktree` declaration | ⬜ DEPENDENCY | waits on ORC-001 |
| Workflow.yaml lint command (`moai workflow lint`) | ⬜ PENDING | no such command exists |
| Sentinel keys `ORC_WORKTREE_*` documentation surfaces | ⬜ PENDING | not currently documented |
| `internal/template/templates/` mirrors of all above | ⬜ PENDING | depends on the above |

**Plan-phase delta**: Run scope reduces from "full lint engine implementation + frontmatter additions + rule edit" to:

1. Add `isolation: worktree` to **4** existing agent files (manager-cycle is dependent on ORC-001, conditional 5th).
2. Update **line 135** of `worktree-integration.md` (SHOULD → MUST).
3. Update **researcher body line** for P-A22 reconciliation.
4. Add a new `moai workflow lint` CLI for `workflow.yaml` role_profiles (REQ-008 static enforcement).
5. Update orchestrator skill body (`team/run.md`) with documentation reinforcement.
6. Add CI fixture tests for AC-06 and AC-07 (regression scenarios).
7. Mirror all source-tree changes to `internal/template/templates/`.
8. Document `ORC_WORKTREE_*` sentinel keys in `worktree-integration.md` and `agent_lint.go` help text.

This is congruent with the partial-pre-completion pattern observed in SPEC-V3R2-RT-006.

---

## 4. Risk Surface

### 4.1 Pre-existing LR-05 false positives

The LR-05 implementation iterates role-profile substrings (`implementer|tester|designer|specialist`) against `agent.Name`. If any agent name happens to contain "specialist" without being a write teammate, it triggers a false positive. Inspection of the current 25-agent roster:

```
$ grep -l "specialist" .claude/agents/moai/*.md
(no matches)
```

No current agent name matches; risk is theoretical. The plan documents this in §8 risks but no remediation is required at this SPEC.

### 4.2 manager-cycle ordering dependency

If ORC-004 run phase begins before ORC-001 merges, the manager-cycle edit task blocks. Mitigation: tasks.md sequences manager-cycle as the LAST agent edit; if ORC-001 has not merged, the run phase emits a structured blocker report citing missing dependency.

If ORC-001 merges *during* ORC-004's run phase (race), the implementer pulls origin/main into the SPEC worktree (`git pull --rebase origin main`) before applying the manager-cycle edit. Standard Git Worktree resync.

### 4.3 manager-ddd / manager-tdd as transitional state

ORC-001 retires manager-ddd and manager-tdd via stub mechanism (a small file with redirect-only content). At the time ORC-004 run executes, these files may be either (a) still present with full body (ORC-001 not yet merged), or (b) reduced to stubs (post-ORC-001 merge). The lint check `checkMissingIsolation` uses an exact-match on `nameLower == agentName`, so it triggers only if `manager-cycle` exists. manager-ddd/manager-tdd are NOT in the write-heavy list of `agent_lint.go:471` and therefore generate no lint error in either state.

### 4.4 manager-spec exclusion

`manager-spec` writes to `.moai/specs/SPEC-XXX/` (a single SPEC directory per invocation). It is technically write-heavy but spec.md §2.1 does not enumerate it. It is not in the LR-05 hardcoded list. Action: leave unchanged, document in plan §1.3 as deliberate scope exclusion.

### 4.5 Researcher body line P-A22 reconciliation

`researcher.md` line 5 of description says: "Uses worktree isolation for safe mutation." Body claim per P-A22 is at line 49 (search: "when possible"). When the SPEC body update lands, the description and body should align with the now-mandatory frontmatter declaration. The researcher.md body content needs full-file inspection during the run phase to find the canonical "when possible" line.

### 4.6 Template-first mirror discipline

CLAUDE.local.md §2 mandates: changes to `.claude/agents/moai/*.md` MUST land first under `internal/template/templates/.claude/agents/moai/*.md`, then `make build && make install`, then the local copy synchronizes via `moai update`.

Inspection: `internal/template/templates/.claude/agents/moai/` contains the same set of agent files. The plan instructs the implementer to:

1. Edit template-side first.
2. Run `make build && make install`.
3. Run `moai update` on the worktree to mirror.
4. Verify byte-identical mirror via `diff -r .claude/agents/moai/ internal/template/templates/.claude/agents/moai/` (excluding allowlist files).

Failure mode: if the implementer edits the local copy first, `moai update` overwrites their changes on the next run. Catch via PostToolUse hook? No such hook exists; rely on the implementer following the plan.

### 4.7 Spawn-time enforcement gap

The absence of a Go-side `Agent()` spawn wrapper means REQ-008 cannot be enforced at runtime by the binary. The chosen mitigation:

1. **Static CI gate**: `moai workflow lint` validates `workflow.yaml` role_profiles. If a contributor changes `implementer.isolation` from `worktree` to `none`, CI fails with `ORC_WORKTREE_REQUIRED`.
2. **Documentation**: `team/run.md` reinforces the rule for the orchestrator's skill body.
3. **Future SPEC** (deferred): a `SubagentStart` hook for runtime telemetry. ORC-004 documents this in plan.md §8 as future work.

This is a known limitation, not a defect. The acceptance test AC-09 verifies the static gate, not runtime spawn rejection.

---

## 5. Reference Implementations

### 5.1 Reference: SPEC-V3R2-RT-006 plan structure

Used as the canonical 7-file plan pattern. Key extracted patterns:

- §1.4 REQ → AC → Task traceability matrix (table format)
- §1.5 Interface signatures table
- §3 Migration strategy (when partial-pre-completion)
- §6 MX tag plan
- §7 plan-auditor self-check mirror (12 dimensions)
- §8 Risks with mitigation references

### 5.2 Reference: SPEC-V3R2-ORC-002 (LR-05 origin)

ORC-002 introduced the LR-05 lint rule as a warning. ORC-004 promotes it to error severity. The implementation of `checkMissingIsolation` was authored under ORC-002 with the severity field hard-coded to `SeverityError` — confirming that the promotion from warning to error has *already* occurred in the source code (not yet documented).

### 5.3 Reference: SPEC-V3R2-CON-001 FROZEN/EVOLVABLE

CON-001 marks `worktree-integration.md` HARD Rule "read-only teammates MUST NOT use worktree" as FROZEN (line 134). The ORC-004 edit on line 135 modifies the EVOLVABLE portion only. Frozen-zone protection is preserved.

### 5.4 Reference: `r5-agent-audit.md`

Located at `.moai/design/v3-redesign/audit/r5-agent-audit.md` (audit document). The §Worktree correctness table flags 6 agents: manager-ddd, manager-tdd (retired in ORC-001 → manager-cycle), expert-backend, expert-frontend, expert-refactoring, researcher. Post-ORC-001 the count collapses to **5 v3r2 agents**: manager-cycle, expert-backend, expert-frontend, expert-refactoring, researcher.

The user-prompt-stated count of 6 reflects the R5 raw audit; the post-ORC-001 v3r2 reality is 5.

---

## 6. Acceptance Driver Mapping

Each AC in spec.md §6 is mapped to a verification mechanism:

| AC | Verification |
|----|--------------|
| AC-01 | `grep -A1 "MUST use" .claude/rules/moai/workflow/worktree-integration.md` → expected new line text |
| AC-02 | `for f in expert-backend expert-frontend expert-refactoring researcher manager-cycle; do grep "isolation: worktree" .claude/agents/moai/$f.md; done` → 5 matches |
| AC-03 | `for f in manager-strategy manager-quality evaluator-active plan-auditor; do grep "isolation: worktree" .claude/agents/moai/$f.md; done` → 0 matches |
| AC-04 | `yq '.workflow.team.role_profiles[] | select(.isolation == "worktree") | .description' .moai/config/sections/workflow.yaml | wc -l` → 3 |
| AC-05 | `moai agent lint` → exit 0 |
| AC-06 | `git stash && moai agent lint --strict` (after manually removing isolation from expert-backend) → exit 1, message contains `ORC_WORKTREE_MISSING` or LR-05 |
| AC-07 | manually add `isolation: worktree` to evaluator-active → `moai agent lint` exit 1 with LR-09 |
| AC-08 | `grep -A2 "All experiments" .claude/agents/moai/researcher.md` → no "when possible" |
| AC-09 | `moai workflow lint` (NEW) → exit 1 with `ORC_WORKTREE_REQUIRED` when implementer.isolation is set to none |
| AC-10 | manual review of skill body in `.claude/skills/moai/team/run.md` for absolute-path discipline |

Each verification mechanism becomes a Given-When-Then scenario in `acceptance.md`.

---

## 7. Open Questions Reconciliation

### OQ1: Should manager-spec receive `isolation: worktree` too?

**Decision (binding)**: NO for ORC-004. spec.md §2.1 explicitly enumerates the 5 v3r2 agents and does not list manager-spec. manager-spec writes to a single SPEC directory and rarely runs in parallel with other manager-spec instances. Future SPEC may revisit.

### OQ2: Should `expert-debug`, `expert-performance`, `expert-security`, `expert-testing`, `expert-devops`, `expert-mobile` receive `isolation: worktree`?

**Decision (binding)**: NO for ORC-004. spec.md §2.1 enumerates only `expert-backend`, `expert-frontend`, `expert-refactoring`. The other expert agents are typically read-and-recommend (security/testing/devops audit-style) and not classified as write-heavy. Future SPEC may revisit per agent based on observed write fan-out.

### OQ3: Should the LR-05 substring matcher (`implementer|tester|designer|specialist`) be tightened?

**Decision (binding)**: NO for ORC-004. The substring match is conservative — it triggers an error when an agent's name *contains* the role-profile substring AND `isolation != "worktree"`. A name like `team-implementer-special` (hypothetical) would match `implementer` and emit LR-05. Since no current agent name contains these substrings except via legitimate intent, the conservative match is acceptable. Tightening to exact match would require per-agent maintenance and provides no current benefit.

### OQ4: Should `ORC_WORKTREE_REQUIRED` be a runtime sentinel emitted by Claude Code's Agent() tool?

**Decision (binding)**: NO for ORC-004. Claude Code's `Agent()` tool does not currently emit project-defined sentinels. The plan's static-CI-gate strategy maps the sentinel to `moai workflow lint` exit messages. Runtime emission is a future feature dependent on Claude Code's hook architecture.

### OQ5: Should `manager-docs` receive `isolation: worktree`?

**Decision (binding)**: NO. spec.md §2.2 (Out of Scope) explicitly excludes manager-docs because it writes to `docs-site/` only — a single tree, no cross-file conflict risk during typical run.

### OQ6: Should the rule clause in worktree-integration.md mention the per-invocation file-write threshold ("3 or more paths")?

**Decision (binding)**: YES. The threshold language ("write files across 3 or more paths per invocation") provides a behavioral classifier for future agents. spec.md REQ-001 mandates this exact language. The threshold is descriptive, not enforced by the lint engine (which uses the hard-coded list). Future LR-05 enhancement could add a CALC step that counts Write/Edit fan-out from agent description.

---

## 8. Build & Validation Cost Estimate

Build steps (run phase):

1. Frontmatter add to 4 existing agents (template-first): 4 × `Edit` operations + `make build && make install` + `moai update` mirror = ~8 steps.
2. manager-cycle frontmatter add (post-ORC-001): 1 × `Edit` + rebuild + mirror = ~3 steps.
3. researcher body line update: 1 × `Edit` (text-only, no rebuild needed for body changes? confirm with `make build` regardless — body is part of the embedded template).
4. worktree-integration.md line 135 edit (template-first): 1 × `Edit` + rebuild + mirror = ~3 steps.
5. New `moai workflow lint` command: ~150 LOC Go + ~80 LOC test = 1 medium task.
6. Sentinel key documentation: 1 × `Edit` to worktree-integration.md + 1 × `Edit` to agent_lint.go help text.
7. AC-06/AC-07 fixture tests: 2 × test additions to `agent_lint_test.go`.
8. CHANGELOG entry: 1 × `Edit`.
9. MX tag annotations: per §6 of plan.md.

Validation steps (sync phase):

1. `make test ./...` — full Go test suite passes.
2. `moai agent lint` exit 0 against post-edit roster.
3. `moai workflow lint` exit 0 against current `workflow.yaml`.
4. Manual injection of LR-05 / LR-09 / ORC_WORKTREE_REQUIRED scenarios — confirm exit 1.
5. Template parity diff: `diff -r .claude/agents/moai/ internal/template/templates/.claude/agents/moai/` returns no diff.
6. CI green: Lint / Test (ubuntu/macos/windows) / Build.

Estimated implementation complexity: **MEDIUM**. The bulk of the work is the new `moai workflow lint` command. Frontmatter edits and rule text edits are mechanical.

---

End of research.

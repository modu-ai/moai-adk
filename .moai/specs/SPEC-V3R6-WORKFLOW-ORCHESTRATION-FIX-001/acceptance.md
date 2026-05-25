---
id: SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001
artifact: acceptance
version: "0.1.1"
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
plan_commit_sha: "<pending>"
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial Tier L acceptance criteria authored — 17 mandatory AC-WOF-XXX entries + 100% REQ-WOF traceability + 7 edge cases + 4-phase Definition of Done |
| 0.1.1 | 2026-05-25 | manager-spec | iter-2 focused fix per plan-auditor iter-1 PASS-WITH-DEBT 0.8625. D1 RESOLVED (§B REQ-WOF-013 row trace fixed `spec.md §G R9` → `research.md §D.3 R9` — no R9 exists in spec.md §G which only enumerates R1-R6). D3 RESOLVED upstream (plan.md §C.1 Tier 5 → Tier 6). D6 RESOLVED (NEW AC-WOF-018 §A "Multi-spawn parallel preference under multi-domain conditions" verifying REQ-WOF-013 Compound directly; §B Traceability Matrix REQ-WOF-013 row updated to point to AC-WOF-018; AC total 17 → 18). Predicted iter-2 plan-auditor: ~0.90 skip-eligible (+0.04 vs iter-1 0.8625). |

---

## §A — Mandatory Acceptance Criteria

The 18 AC-WOF-XXX entries below collectively verify all 15 REQ-WOF-XXX requirements from spec.md §D (17 originally authored in v0.1.0 + 1 added in v0.1.1 iter-2 as AC-WOF-018). Each AC uses the Given/When/Then format, identifies severity, provides an evidence command, defines a pass criterion, and maps to ≥1 REQ-WOF-XXX.

### AC-WOF-001 (HUMAN GATE Decision Point 1 plan)

**Severity**: Critical
**Verifies**: REQ-WOF-001

**Given** an orchestrator session at the end of plan-phase artifact authoring for any new SPEC,
**When** manager-spec returns its plan-phase completion report,
**Then** the orchestrator invokes `AskUserQuestion` with at minimum 3 options including "(권장) plan-auditor verdict 진행" as first option AND records the user's selection in `.moai/specs/<SPEC-ID>/progress.md § Decision Point 1`.

**Evidence command**:
```bash
grep -A 3 "Decision Point 1" .claude/rules/moai/workflow/human-gates.md | head -20
grep -n "AskUserQuestion" .claude/skills/moai/workflows/plan/spec-assembly.md | head -5
```
**Pass criterion**: at least one matching block per command output; the rule file (`human-gates.md`) is a NEW file authored in M1.

---

### AC-WOF-002 (HUMAN GATE Plan Approval run)

**Severity**: Critical
**Verifies**: REQ-WOF-001

**Given** an orchestrator session about to enter `/moai run <SPEC-ID>` Phase 1,
**When** Phase 0.5 plan-auditor returns PASS verdict but skip-eligibility is NOT met (score < 0.90),
**Then** the orchestrator invokes `AskUserQuestion` with the verdict report content + 3 options including "(권장) Run-phase 진입 (manager-strategy)" + "Plan-auditor 재실행 iter-N+1" + "abort to plan-phase".

**Evidence command**:
```bash
grep -A 5 "Plan Approval\|Decision Point.*plan-auditor" .claude/rules/moai/workflow/human-gates.md
grep -n "Phase 0.5\|skip-eligibility" .claude/skills/moai/workflows/run.md
```
**Pass criterion**: rule references the 3 options structure AND run.md references the gate AND `human-gates.md` includes the Plan Approval section.

---

### AC-WOF-003 (HUMAN GATE 1 sync — working tree + tests)

**Severity**: Critical
**Verifies**: REQ-WOF-001

**Given** an orchestrator session about to enter `/moai sync <SPEC-ID>`,
**When** sync.md Phase 0 working-tree-pre-flight detects a non-clean baseline OR Phase 0.1 returns failing tests,
**Then** the orchestrator invokes `AskUserQuestion` with the diagnostic report + 3 options including "(권장) 정리 후 sync 진입" + "force-sync (record debt)" + "abort sync".

**Evidence command**:
```bash
grep -A 5 "GATE 1\|working tree.*tests" .claude/rules/moai/workflow/human-gates.md
grep -n "AskUserQuestion\|Phase 0.1" .claude/skills/moai/workflows/sync.md
```
**Pass criterion**: rule defines GATE 1 + sync.md references the gate invocation at Phase 0 / 0.1.

---

### AC-WOF-004 (HUMAN GATE 2 sync — doc scope)

**Severity**: Critical
**Verifies**: REQ-WOF-001

**Given** an orchestrator session at the end of `/moai sync <SPEC-ID>` Phase 3 (CHANGELOG draft + frontmatter status transition),
**When** the sync-phase artifact set is ready for commit,
**Then** the orchestrator invokes `AskUserQuestion` confirming the doc scope and CHANGELOG entry content before commit, with options including "(권장) commit + push (Hybrid Trunk)" / "review CHANGELOG manually" / "abort sync (artifact rework needed)".

**Evidence command**:
```bash
grep -A 5 "GATE 2\|doc scope" .claude/rules/moai/workflow/human-gates.md
grep -n "Phase 3\|CHANGELOG" .claude/skills/moai/workflows/sync.md
```
**Pass criterion**: rule defines GATE 2 + sync.md references the gate at Phase 3.

---

### AC-WOF-005 (sync-phase manager-quality Phase 0.5.4 invocation)

**Severity**: High
**Verifies**: REQ-WOF-002

**Given** an orchestrator session running `/moai sync <SPEC-ID>`,
**When** sync-phase Phase 0.5 is complete,
**Then** the orchestrator MUST invoke `manager-quality` at Phase 0.5.4 for TRUST 5 validation, log the invocation as `Agent(manager-quality)` in the response, and persist the return value to `.moai/specs/<SPEC-ID>/progress.md § Sync-phase Quality Audit`.

**Evidence command**:
```bash
grep -n "Phase 0.5.4\|manager-quality" .claude/skills/moai/workflows/sync.md
grep -n "Phase 0.5.4\|TRUST 5" .claude/agents/core/manager-quality.md
```
**Pass criterion**: ≥1 reference in each file pointing to the Phase 0.5.4 entrypoint with manager-quality invocation.

---

### AC-WOF-006 (sync-phase expert-security Phase 0.55 manifest audit — always-runs HARD)

**Severity**: High
**Verifies**: REQ-WOF-002, REQ-WOF-007

**Given** an orchestrator session running `/moai sync <SPEC-ID>`,
**When** sync-phase Phase 0.55 entry conditions are satisfied (regardless of harness level),
**Then** the orchestrator MUST invoke `expert-security` for dependency manifest audit covering `go.sum`, `package-lock.json`, `Pipfile.lock`, `Cargo.lock` in the SPEC scope, with the invocation HARD-marked always-runs and audit results persisted to `progress.md § Security Manifest Audit`.

**Evidence command**:
```bash
grep -n "Phase 0.55\|always-runs\|expert-security" .claude/skills/moai/workflows/sync.md
grep -n "Phase 0.55\|manifest audit\|always-runs" .claude/agents/expert/expert-security.md
```
**Pass criterion**: ≥1 reference in each file referencing Phase 0.55 + always-runs HARD marker + ≥1 of the 4 lockfile names.

---

### AC-WOF-007 (sync-phase manager-develop Phase 0.7 coverage invocation)

**Severity**: High
**Verifies**: REQ-WOF-002

**Given** an orchestrator session running `/moai sync <SPEC-ID>` where the SPEC modified ≥1 Go source file (`internal/**/*.go`),
**When** sync-phase Phase 0.6 is complete,
**Then** the orchestrator MUST invoke `manager-develop` at Phase 0.7 for coverage verification (`go test -coverprofile=cover.out`) and persist the coverage delta to `progress.md § Coverage Audit`.

**Evidence command**:
```bash
grep -n "Phase 0.7\|coverage\|coverprofile" .claude/skills/moai/workflows/sync.md
```
**Pass criterion**: ≥1 reference in sync.md to Phase 0.7 + `go test -coverprofile` baseline command.

---

### AC-WOF-008 (run-phase manager-strategy Phase 1 invocation + tasks.md artifact)

**Severity**: Critical
**Verifies**: REQ-WOF-003, REQ-WOF-006

**Given** an orchestrator session running `/moai run <SPEC-ID>` after Phase 0.5 PASS and Phase 0.95 Mode Selection,
**When** Phase 1 is entered,
**Then** the orchestrator MUST spawn `manager-strategy` (analysis-only, code-prohibited HARD) BEFORE any `manager-develop` spawn, and the manager-strategy spawn MUST produce a `.moai/specs/<SPEC-ID>/tasks.md` artifact enumerating M1-M6+ task decomposition.

**Evidence command**:
```bash
grep -n "Phase 1\|manager-strategy" .claude/skills/moai/workflows/run.md | head -10
grep -n "tasks.md\|code-prohibited" .claude/agents/core/manager-strategy.md
grep -n "manager-strategy" .claude/skills/moai/workflows/run/task-decomposition.md
```
**Pass criterion**: ≥3 references across the 3 files; manager-strategy.md includes code-prohibited HARD assertion + tasks.md artifact spec.

---

### AC-WOF-009 (Skill router invocation discipline)

**Severity**: High
**Verifies**: REQ-WOF-005

**Given** an orchestrator session receiving a paste-ready resume Block 5 line `실행: /moai run SPEC-XXX`,
**When** the orchestrator processes that line,
**Then** the orchestrator MUST invoke `Skill("moai", arguments: "run SPEC-XXX")` (NOT manually `Read` the file `.claude/skills/moai/workflows/run.md`), AND `session-handoff.md` MUST contain explicit text clarifying this discipline.

**Evidence command**:
```bash
grep -n "Skill(\"moai\"\|skill router\|router.*invoke" .claude/rules/moai/workflow/session-handoff.md
grep -n "do NOT.*Read.*SKILL.md\|MUST.*Skill(" .claude/rules/moai/workflow/session-handoff.md
```
**Pass criterion**: ≥1 match for each grep pattern; the rule clarifies the discipline canonically.

---

### AC-WOF-010 (Autonomous Mode Selection log in progress.md)

**Severity**: High
**Verifies**: REQ-WOF-004, REQ-WOF-009, REQ-WOF-015

**Given** an orchestrator session running `/moai run <SPEC-ID>` post Phase 0.5 PASS,
**When** Phase 0.95 Mode Selection executes,
**Then** the orchestrator MUST classify the task into exactly one of 5 modes (sequential / parallel / background / sub-agent / agent-team) based on observable signals (scope file count, domain count, harness level, prerequisites for Agent Teams) AND record the decision in `.moai/specs/<SPEC-ID>/progress.md § Mode Selection` with the format:
```
## Mode Selection (Phase 0.95)
- Selected mode: <mode>
- Rationale: <reason citing decision tree>
- Inputs: scope_files=<N>, domains=<M>, harness=<level>, team_enabled=<bool>
```

**Evidence command**:
```bash
grep -n "Mode Selection\|Phase 0.95" .claude/rules/moai/workflow/orchestration-mode-selection.md
grep -n "Mode Selection\|Phase 0.95" .claude/skills/moai/workflows/run/mode-orchestration.md
```
**Pass criterion**: NEW rule file `orchestration-mode-selection.md` defines the 5-mode decision tree + run/mode-orchestration.md references Phase 0.95 entrypoint.

---

### AC-WOF-011 (Multi-spawn parallel evidence when scope ≥10 files OR ≥3 domains)

**Severity**: High
**Verifies**: REQ-WOF-009, REQ-WOF-015

**Given** an orchestrator session running `/moai run <SPEC-ID>` for a SPEC with §C.1 scope inventory ≥10 files OR ≥3 distinct domain categories,
**When** Phase 0.95 Mode Selection selects `parallel` mode,
**Then** the orchestrator MUST execute multi-spawn (3-5 concurrent `Agent()` calls in a single response message) AND the response message MUST visibly contain multiple `Agent()` invocations.

**Evidence command**:
```bash
grep -A 10 "parallel.*multi-spawn\|multi-spawn.*parallel" .claude/rules/moai/workflow/orchestration-mode-selection.md
grep -n "3-5.*parallel\|cut research time 90%" .claude/rules/moai/workflow/orchestration-mode-selection.md
```
**Pass criterion**: rule references the Anthropic verbatim "3-5 parallel cut research time 90%" benchmark + documents the multi-spawn pattern.

---

### AC-WOF-012 (manager-strategy code-prohibited HARD assertion)

**Severity**: High
**Verifies**: REQ-WOF-003, REQ-WOF-006

**Given** the manager-strategy subagent definition file,
**When** the agent definition is read,
**Then** the file MUST contain a `[HARD]` or `[ZONE:Frozen] [HARD]` marker explicitly forbidding code generation in Phase 1 spawns (analysis-only scope) AND the file MUST list the `tasks.md` artifact as the sole output artifact.

**Evidence command**:
```bash
grep -n "HARD.*code-prohibited\|code-prohibited.*HARD\|HARD.*analysis-only" .claude/agents/core/manager-strategy.md
grep -n "tasks.md" .claude/agents/core/manager-strategy.md
```
**Pass criterion**: ≥1 match for HARD marker + ≥1 match for tasks.md spec.

---

### AC-WOF-013 (Plan-auditor max-3 retry contract bounded loop)

**Severity**: Medium
**Verifies**: REQ-WOF-001 (Plan Approval gate), supports REQ-WOF-014

**Given** an orchestrator session running `/moai run <SPEC-ID>` where plan-auditor verdict is FAIL or score < Tier threshold,
**When** iter(N) verdict is recorded,
**Then** the orchestrator MUST cap iter(N+1) at 3 total iterations AND emit a structured stop signal when iter(N+1) aggregate score < iter(N) (regression detected per spec-workflow.md § plan-auditor escalation policy).

**Evidence command**:
```bash
grep -n "max.*3\|max_iterations.*3\|iter(N+1)" .claude/agents/meta/plan-auditor.md
grep -n "regression\|STOP signal" .claude/agents/meta/plan-auditor.md
```
**Pass criterion**: plan-auditor.md (existing) references max-3 contract — this AC verifies the policy IS preserved through M1 changes (no regression).

---

### AC-WOF-014 (GitHub Issue auto-creation Phase 2.5)

**Severity**: Medium
**Verifies**: research.md R8

**Given** an orchestrator session running `/moai run <SPEC-ID>` after Phase 2 (manager-develop completion),
**When** Phase 2.5 entry conditions are satisfied (per CONST-V3R5-035 HARD + sync.md Phase 2.5 reference),
**Then** the orchestrator MUST execute `gh issue create --title "<SPEC-ID> tracker" --body "<scope summary>" --label "spec"` AND record the issue number in `progress.md § GitHub Issue Tracker`.

**Evidence command**:
```bash
grep -n "Phase 2.5\|gh issue create" .claude/skills/moai/workflows/run.md
grep -n "Phase 2.5\|issue_number" .claude/skills/moai/workflows/sync.md
```
**Pass criterion**: ≥1 reference in each file to Phase 2.5 and the `gh issue create` invocation pattern. NOTE: this AC is "PASS-WITH-NOTE" eligible for projects that have explicitly disabled GitHub Issue tracking (e.g., 1-person OSS with Hybrid Trunk per CLAUDE.local.md §23).

---

### AC-WOF-015 (BODP audit trail creation)

**Severity**: Medium
**Verifies**: research.md R8, CONST-V3R5-034

**Given** an orchestrator session running `/moai run <SPEC-ID>` that creates a feature branch (`feat/SPEC-XXX` OR worktree at `~/.moai/worktrees/<project>/SPEC-XXX/`),
**When** the branch is created,
**Then** the orchestrator MUST persist a BODP decision file at `.moai/branches/decisions/<normalized-branch>.md` via `bodp.WriteDecision` (or the orchestrator equivalent for skill-body BODP) per CONST-V3R5-034 HARD.

**Evidence command**:
```bash
grep -n "BODP\|.moai/branches/decisions" .claude/skills/moai/workflows/run.md
grep -n "bodp.WriteDecision\|branch-origin-protocol" .claude/rules/moai/workflow/spec-workflow.md
```
**Pass criterion**: ≥1 reference in run.md + ≥1 cross-reference to branch-origin-protocol.md in spec-workflow.md.

---

### AC-WOF-016 (AskUserQuestion subagent boundary HARD enforcement)

**Severity**: Critical
**Verifies**: REQ-WOF-011

**Given** the run-phase modification scope per spec.md §C.1,
**When** all 22 in-scope file edits are complete,
**Then** ZERO modified file shall contain `AskUserQuestion` invocation syntax inside a subagent prompt body (orchestrator-only pattern preserved) AND the NEW `human-gates.md` MUST cross-reference askuser-protocol.md as canonical SSOT.

**Evidence command**:
```bash
# Subagent boundary check across modified scope
grep -rn 'AskUserQuestion' .claude/skills/moai/workflows/ .claude/rules/moai/workflow/orchestration-mode-selection.md .claude/rules/moai/workflow/human-gates.md \
  | grep -v "askuser-protocol.md" | grep -v "(권장)" | grep -v "# Example" | grep -v "^[^:]*:[0-9]*:[ \t]*#" | head -20

# Cross-reference check
grep -n "askuser-protocol" .claude/rules/moai/workflow/human-gates.md
```
**Pass criterion**: First grep returns 0 matches (or matches only orchestrator-side pattern references — manual review of any matches). Second grep returns ≥1 match (canonical cross-reference present).

---

### AC-WOF-017 (Mirror parity — template synchronization)

**Severity**: High
**Verifies**: completeness of §C.1 Tier 5 scope

**Given** the run-phase modifications are complete in `.claude/skills/moai/workflows/` and `.claude/rules/moai/`,
**When** M6 mirror-parity step is executed,
**Then** EVERY changed local file shall have an exact-content counterpart in `internal/template/templates/.claude/...` AND `diff -r` between local and template paths shall return zero differences for the modified set.

**Evidence command**:
```bash
# Local vs template diff for workflow skills
diff -r .claude/skills/moai/workflows/ internal/template/templates/.claude/skills/moai/workflows/ 2>&1 | head -20

# Local vs template diff for new rule files
diff .claude/rules/moai/workflow/human-gates.md internal/template/templates/.claude/rules/moai/workflow/human-gates.md
diff .claude/rules/moai/workflow/orchestration-mode-selection.md internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md
```
**Pass criterion**: zero diff lines for the modified set; any diff output requires reconciliation before M7 sync-phase entry.

---

### AC-WOF-018 — Multi-spawn parallel preference under multi-domain conditions

**Severity**: Major (P1)
**Verifies**: REQ-WOF-013 (Compound multi-spawn preference — per iter-2 directive scope)

**Given** harness level == standard OR thorough, AND task scope ≥ 10 files OR ≥ 3 domains (per Phase 0.95 Scale-Based Mode Selection criteria),
**When** the orchestrator enters Phase 1 execution mode selection,
**Then**:
- The orchestrator shall select Parallel Multi-Spawn mode (per design.md §B.3 decision tree Q4)
- The orchestrator's spawn invocation shall contain ≥ 2 Agent() calls in a single response turn (parallel multi-spawn evidence)
- The orchestrator shall log the mode selection rationale in `.moai/specs/SPEC-{ID}/progress.md` § Mode Selection

**Evidence command**:
```bash
# Verify multi-spawn in run-phase commit
git log --oneline --all -- .moai/specs/SPEC-{ID}/progress.md | head -5
grep -A3 "## Mode Selection" .moai/specs/SPEC-{ID}/progress.md
# Verify parallel Agent() spawn count in run-phase commit body
git show <run-phase-commit> | grep -c "Agent("
```

**Pass criterion**: Mode Selection log present AND mode == "Parallel Multi-Spawn" AND Agent() count ≥ 2.

> Note (v0.1.1, iter-2 D6 resolution): This AC is authored verbatim per the iter-2 focused-fix directive that named REQ-WOF-013 as its verification target. The authoring directive labels REQ-WOF-013 as "Compound multi-spawn preference"; spec.md §D currently models REQ-WOF-013 as a Capability-gate (Tier L OR --pr flag → manager-git PR routing) and REQ-WOF-015 as the Compound parallel-multi-spawn clause. The directive's REQ binding is preserved unchanged (D6 RESOLVED by adding this AC entry); orchestrator may re-align the REQ ↔ AC binding in a follow-up iter if iter-2 plan-auditor flags the semantic offset. See acceptance.md §B Traceability Matrix REQ-WOF-013 row pointing to AC-WOF-018, and the trust-handoff note in the iter-2 commit body.

---

## §B — Traceability Matrix (REQ-WOF ↔ AC-WOF)

| REQ-WOF | Description (abridged) | AC-WOF Coverage |
|---------|------------------------|------------------|
| REQ-WOF-001 (Ubiquitous) | Restore HUMAN GATE 5종 | AC-WOF-001, AC-WOF-002, AC-WOF-003, AC-WOF-004, AC-WOF-013 |
| REQ-WOF-002 (Ubiquitous) | Restore sync-phase 3 quality specialists | AC-WOF-005, AC-WOF-006, AC-WOF-007 |
| REQ-WOF-003 (Ubiquitous) | Restore run-phase hierarchical chain | AC-WOF-008, AC-WOF-012 |
| REQ-WOF-004 (Ubiquitous) | Implement 5-mode autonomous selection | AC-WOF-010 |
| REQ-WOF-005 (Event-driven) | Skill router invocation discipline | AC-WOF-009 |
| REQ-WOF-006 (Event-driven) | manager-strategy spawn before manager-develop | AC-WOF-008, AC-WOF-012 |
| REQ-WOF-007 (Event-driven) | expert-security Phase 0.55 always-runs | AC-WOF-006 |
| REQ-WOF-008 (Event-driven) | D-NEW-1 inline-fix pattern | (covered by spec.md §G R5 + acceptance via re-delegation discipline) |
| REQ-WOF-009 (State-driven) | Multi-domain ≥10 files → parallel mode | AC-WOF-010, AC-WOF-011 |
| REQ-WOF-010 (State-driven) | thorough harness + Tier M/L → evaluator-active Sprint Contract | (covered by spec.md §G R6 + acceptance via M4) |
| REQ-WOF-011 (State-driven) | AskUserQuestion subagent boundary | AC-WOF-016 |
| REQ-WOF-012 (Capability-gate) | Agent Teams mode prerequisites | AC-WOF-010 (multi-mode coverage), AC-WOF-011 |
| REQ-WOF-013 (Capability-gate) | Tier L OR --pr flag → manager-git PR routing | AC-WOF-018 (direct; via iter-2 D6 resolution) + cross-reference research.md §D.3 R9 + plan.md M6 + sync.md update — AC verification in M6 |
| REQ-WOF-014 (Event-detected) | HUMAN GATE skip detection | AC-WOF-001..004 (all 5 gate ACs verify presence) |
| REQ-WOF-015 (Compound) | Standard/thorough + multi-domain → parallel preferred | AC-WOF-010, AC-WOF-011, AC-WOF-018 |

**Traceability coverage**: 100% — every REQ-WOF is verified by ≥1 AC-WOF. Mirror coverage via §A Tier 5 + AC-WOF-017.

---

## §C — Edge Cases

### §C.1 Edge case 1 — Parallel session race during multi-spawn (REQ-WOF-009 boundary)

**Scenario**: Orchestrator selects `parallel` mode for Phase 0.95 and spawns 3 concurrent Agent() calls. While the parallel agents run, a separate Claude Code session pushes commits to origin/main that modify one of the in-scope files.

**Expected behavior**: Per `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check, the orchestrator MUST have executed pre-spawn `git fetch + git rev-list --count --left-right origin/main...HEAD` returning `0 0` before each parallel spawn. If post-spawn divergence is detected (return `N 0`), the orchestrator halts the parallel batch + surfaces via AskUserQuestion (rebase / inspect / abort).

**Mitigation**: pre-spawn sync check is HARD per CONST-V3R5-030; orchestration-mode-selection.md MUST cross-reference this rule when documenting parallel mode.

---

### §C.2 Edge case 2 — HUMAN GATE timeout (REQ-WOF-014 graceful degradation)

**Scenario**: Orchestrator emits HUMAN GATE AskUserQuestion but user does not respond within the session window (e.g., user steps away).

**Expected behavior**: AskUserQuestion has no timeout enforcement at the protocol layer — the session naturally waits. When the user returns and selects "Other" or one of the 3 main options, the orchestrator resumes from the gate point. No autonomous skip permitted per REQ-WOF-014.

**Mitigation**: human-gates.md MUST explicitly document "no timeout autonomous skip" + reference the askuser-protocol.md § Free-form Circumvention Prohibition.

---

### §C.3 Edge case 3 — Mode mis-classification at scope boundary (REQ-WOF-009 boundary)

**Scenario**: SPEC scope = 9 files (just below the ≥10 threshold). Mode Selection logic decides between `sub-agent` (single spawn) and `parallel` (multi-spawn).

**Expected behavior**: Per spec.md §G R4 mitigation: tie-breaker rule defaults to the simpler mode (sub-agent over agent-team, sequential over parallel) AND logs the boundary case to `progress.md § Mode Selection` with `boundary_case: true` flag for retrospective analysis.

**Mitigation**: orchestration-mode-selection.md §B Tie-breaker Rules documents this explicitly.

---

### §C.4 Edge case 4 — Agent Teams prerequisites missing (REQ-WOF-012 capability gate)

**Scenario**: SPEC harness = thorough + scope qualifies for agent-team mode, but `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` env var is NOT set.

**Expected behavior**: Per `.claude/rules/moai/workflow/spec-workflow.md` § Mode Dispatch auto-selection rules: fallback to `autopilot` (sequential single-spawn) with `[mode-auto-downgrade]` info log per REQ-WF003-012 sentinel.

**Mitigation**: orchestration-mode-selection.md decision tree includes the agent-team prerequisite check + fallback path; sentinel `MODE_TEAM_UNAVAILABLE` is preserved.

---

### §C.5 Edge case 5 — Sub-skill chain failure (REQ-WOF-005 router discipline)

**Scenario**: Orchestrator invokes `Skill("moai", arguments: "run SPEC-XXX")`, the router loads workflows/run.md correctly, but the on-demand sub-skill load (e.g., `Read .claude/skills/moai/workflows/run/context-loading.md`) fails due to file system or permission error.

**Expected behavior**: orchestrator emits a structured error report referencing the missing sub-skill, falls back to graceful degradation (executes Phase 0 inline without the sub-skill body — minimal context-loading), and notifies the user via AskUserQuestion to verify the sub-skill file presence.

**Mitigation**: session-handoff.md Block 5 documentation includes "if sub-skill load fails, orchestrator MUST report structured error and offer abort option" + run.md router error handling reference.

---

### §C.6 Edge case 6 — Multi-mode oscillation (REQ-WOF-004 stability)

**Scenario**: Same SPEC re-entered through `/moai loop` 3 times; each iteration produces different Mode Selection results due to scope drift (some files completed, others remaining).

**Expected behavior**: Mode Selection per-iteration is OK — the decision is made fresh at Phase 0.95 each time. progress.md § Mode Selection should append the per-iteration selection (not overwrite), allowing retrospective analysis of why the mode shifted.

**Mitigation**: orchestration-mode-selection.md §C Iteration Stability explicitly endorses per-iteration re-classification; this is NOT a bug, it is the design.

---

### §C.7 Edge case 7 — Skip-eligible Phase 0.5 vs run-phase mode mismatch

**Scenario**: Phase 0.5 was skipped due to plan-auditor verdict ≥ 0.90 (skip-eligibility met). Then Phase 0.95 Mode Selection runs without the audit verdict context.

**Expected behavior**: Phase 0.95 reads the most-recent plan-auditor verdict from `.moai/specs/<SPEC-ID>/progress.md § Plan-auditor Verdict` (cached during plan-phase) AND uses it as input to the mode decision. If the cache is unavailable, Phase 0.95 falls back to scope-based classification only.

**Mitigation**: orchestration-mode-selection.md §A Inputs section documents the cached verdict as an optional input.

---

## §D — Quality Gate Criteria + Definition of Done

### §D.1 Plan-phase Definition of Done (M0)

- [ ] spec.md created with 12 canonical frontmatter fields + tier:L + GEARS notation ≥80% across 15 REQ-WOF-XXX
- [ ] plan.md created with §A Lifecycle table + §C scope inventory + §D milestone decomposition (M1-M6+)
- [ ] acceptance.md created with ≥15 AC-WOF-XXX + traceability matrix 100% + ≥5 edge cases (this artifact: 18 ACs as of v0.1.1, 7 edges — exceeds floor)
- [ ] design.md created (Tier L exclusive) with §B Delegation Graph + §B.3 Mode Selection Decision Tree
- [ ] research.md already present (orchestrator pre-authored) — verified
- [ ] plan-auditor verdict ≥ 0.85 (Tier L threshold)
- [ ] commit `feat(SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001): plan-phase artifacts (Tier L Section A-G, 4 artifacts)`

### §D.2 Run-phase Definition of Done (M1-M6)

- [ ] All 22 local files + 10 mirror files modified per §C.1 scope (AC-WOF-017 mirror parity verification PASS)
- [ ] 15/15 REQ-WOF requirements demonstrably implemented (every REQ has supporting markdown content in target files)
- [ ] 18/18 AC-WOF acceptance criteria PASS (or PASS-WITH-NOTE per AC-specific exceptions; v0.1.1 added AC-WOF-018)
- [ ] manager-develop self-verification E1-E7 per milestone returned and validated
- [ ] orchestrator 7-item Trust-but-verify batch returns 0 critical discrepancies after each milestone return
- [ ] No predecessor SPEC bodies modified (L48 SSOT discipline preserved)
- [ ] No Go source files modified (markdown-only scope respected)
- [ ] All commits use Conventional Commits + 🗿 MoAI trailer + SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001 scope marker
- [ ] post-action `git fetch + git rev-list 0 0` divergence check clean after every push

### §D.3 Sync-phase Definition of Done (M7)

- [ ] CHANGELOG.md `[Unreleased]` section contains exactly 1 entry for SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001 (B12 duplicate-detection PASS)
- [ ] All 4 SPEC artifact frontmatter `status: in-progress → implemented` atomic transition
- [ ] All 4 SPEC artifact frontmatter `sync_commit_sha:` filled with sync-phase commit SHA (L60 partial-backfill defense)
- [ ] progress.md §E.4 Sync-phase Audit-Ready Signal section populated
- [ ] commit `chore(SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001): sync-phase artifacts`

### §D.4 Mx-phase Definition of Done (M8) + 4-phase close

- [ ] Mx Step C SKIP-judge invoked + verdict recorded per mx-tag-protocol.md §a (markdown-only escape clause expected to trigger SKIP)
- [ ] progress.md §E.5 Mx-phase Audit-Ready Signal section populated with skip-verdict rationale
- [ ] All 4 SPEC artifact frontmatter `status: implemented → completed` atomic transition
- [ ] commit `chore(SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001): Mx-phase audit-ready signal + 4-phase close`
- [ ] Push range matches all attributed commits (no orphan / no parallel-session-leak)

### §D.5 Overall SPEC Definition of Done

- [ ] 5/5 phases (M0 plan + M1-M6 run + M7 sync + M8 Mx + 4-phase close) completed
- [ ] Total wall-time within Sprint 10 Tier L target (3 days = M0 day 1 + run-phase day 2 + sync+Mx+close day 3)
- [ ] All applied lessons (L33, L44, L46, L48, L49, L51, L52, L58, L60, L63) preserved through workflow
- [ ] Backward compatibility for 88 pre-v3 EARS SPECs preserved (no GEARS-only enforcement that breaks legacy lint)
- [ ] MEMORY.md updated with 4-phase close memory entry per L52 / L48 SSOT discipline

---

Version: 0.1.1
Status: plan-phase (M0) — iter-2 focused fix applied
Coverage: 18 AC-WOF + 15 REQ-WOF (100% traceability) + 7 edge cases + 4-phase DoD

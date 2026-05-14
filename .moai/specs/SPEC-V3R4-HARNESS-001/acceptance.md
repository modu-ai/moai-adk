# Acceptance Criteria — SPEC-V3R4-HARNESS-001

This document defines the acceptance criteria for the foundation SPEC of the self-evolving harness v2 system. Each AC corresponds to one or more REQ-HRN-FND-NNN requirements in `spec.md` §5. Every AC uses Given / When / Then format and includes objectively verifiable evidence.

---

## REQ → AC Coverage Matrix

| REQ ID | AC ID | Description |
|--------|-------|-------------|
| REQ-HRN-FND-001 | AC-HRN-FND-001 | CLI verb path retired, slash command remains supported |
| REQ-HRN-FND-002 | AC-HRN-FND-001 | CI guard prevents re-registration |
| REQ-HRN-FND-003 | AC-HRN-FND-002 | Slash command verbs reachable without Go binary |
| REQ-HRN-FND-004 | AC-HRN-FND-003 | Tier-4 AskUserQuestion gate by orchestrator |
| REQ-HRN-FND-005 | AC-HRN-FND-004 | 5-Layer Safety preserved verbatim |
| REQ-HRN-FND-006 | AC-HRN-FND-005 | FROZEN zone path-prefix block + audit log |
| REQ-HRN-FND-007 | AC-HRN-FND-006 | Pre-modification snapshot created |
| REQ-HRN-FND-008 | AC-HRN-FND-006 | Rollback restores byte-identical state |
| REQ-HRN-FND-009 | AC-HRN-FND-007 | Observer no-op when learning.enabled=false |
| REQ-HRN-FND-010 | AC-HRN-FND-008 | PostToolUse baseline observation preserved |
| REQ-HRN-FND-011 | AC-HRN-FND-008 | 4-tier ladder thresholds preserved |
| REQ-HRN-FND-012 | AC-HRN-FND-009 | Tier-4 rate limit of 1/week per project |
| REQ-HRN-FND-013 | AC-HRN-FND-010 | Superseded V3R3 SPECs not modified in this PR |
| REQ-HRN-FND-014 | AC-HRN-FND-005 | Frozen-guard violation audit log emission |
| REQ-HRN-FND-015 | AC-HRN-FND-003 | Subagent AskUserQuestion prohibition |
| REQ-HRN-FND-016 | AC-HRN-FND-011 | Success-metric telemetry exposure |
| REQ-HRN-FND-017 | AC-HRN-FND-012 | Reflexion-evaluator conflict resolution (contract assertion) |
| REQ-HRN-FND-018 | AC-HRN-FND-009 | Adaptive rate-limit expansion preserves floor |

Coverage: 18 REQs ↔ 12 ACs. Every REQ appears in at least one AC.

---

## Acceptance Criteria

### AC-HRN-FND-001 — CLI Verb Path Retirement and CI Guard

**Linked REQs**: REQ-HRN-FND-001, REQ-HRN-FND-002

**Given** a project with the moai-adk-go v2.20.0 (or target release) binary installed and `internal/cli/root.go` reviewed,

**When** the user runs `moai harness status` (or any other harness verb) from the terminal,

**Then** the system MUST return cobra's `unknown command "harness" for "moai"` diagnostic with a non-zero exit code. The CLI verb path MUST NOT be reachable as a public surface.

**And When** a hypothetical pull request introduces a line registering `newHarnessCmd()` (or any equivalent harness subcommand factory) into `internal/cli/root.go` or any cobra command tree,

**Then** the continuous integration system MUST fail the build with a diagnostic message referencing `SPEC-V3R4-HARNESS-001` and `REQ-HRN-FND-002`.

**Verification**:
- Manual: `moai harness status 2>&1 | head -3` returns `unknown command` text.
- Automated: a Go test (`internal/cli/harness_retirement_test.go` or equivalent) asserts that `rootCmd.Commands()` does not include any command with `Use: "harness"`.

---

### AC-HRN-FND-002 — Slash Command Verb Coverage Without Binary

**Linked REQs**: REQ-HRN-FND-003

**Given** a Claude Code session with the `moai` skill loaded,

**When** the user invokes `/moai:harness status` (or `apply`, `rollback`, `disable`),

**Then** the system MUST render the requested verb's behavior using the `moai` skill workflow body alone, without invoking `moai harness`, `moai harness status`, or any other Go binary subcommand involving the word `harness`.

**Verification**:
- Static: `grep -nE 'moai harness' .claude/commands/moai/harness.md .claude/skills/moai/workflows/harness.md .claude/skills/moai/SKILL.md` returns zero matches in invocation context (matches in commentary or quoted historical text are acceptable).
- Manual: in a Claude Code session, invoke `/moai:harness status` and observe the response rendered without a `Bash` tool call to `moai harness`.

---

### AC-HRN-FND-003 — Tier-4 AskUserQuestion Gate (Orchestrator-Only)

**Linked REQs**: REQ-HRN-FND-004, REQ-HRN-FND-015

**Given** a project with at least one Tier-4 pending proposal under `.moai/harness/proposals/`,

**When** the user invokes `/moai:harness apply`,

**Then** the MoAI orchestrator MUST invoke `AskUserQuestion` with at least three options (Apply, Defer, Reject — additional options permitted) and the first option MUST carry the `(권장)` or `(Recommended)` suffix.

**And** no harness-related subagent (specifically `moai-harness-learner` or any V3R4 successor agent) MUST invoke `AskUserQuestion` from inside its prompt. Subagents that require user input MUST return a structured blocker report to the orchestrator.

**Verification**:
- Static: `grep -rn 'AskUserQuestion' .claude/agents/moai/moai-harness-learner.md .claude/agents/moai/manager-spec.md` (and other harness-related agents) returns zero matches in invocation context. References to the AskUserQuestion contract (e.g., "Subagents MUST NOT prompt") are acceptable.
- Manual: trace a `/moai:harness apply` invocation in a session log; observe `AskUserQuestion` appearing in the orchestrator turn, not in any subagent response.

---

### AC-HRN-FND-004 — 5-Layer Safety Verbatim Preservation

**Linked REQs**: REQ-HRN-FND-005

**Given** the merged state of this SPEC's PR,

**When** `.claude/rules/moai/design/constitution.md` §5 is diffed against the `main` branch's pre-PR version,

**Then** the diff MUST be empty for §5 content — every layer (L1 Frozen Guard, L2 Canary Check, L3 Contradiction Detector, L4 Rate Limiter, L5 Human Oversight) MUST remain byte-identical.

**Verification**:
- Automated: `git diff main -- .claude/rules/moai/design/constitution.md | grep -E '^\+\+\+|^---|^\+|^-' | grep -v '^[+-]{3}'` MUST return zero non-comment lines within the §5 range.
- Manual: visual inspection of `.claude/rules/moai/design/constitution.md` §5 (lines 123-170 as of 2026-05-14) confirms unchanged.

---

### AC-HRN-FND-005 — FROZEN Zone Path-Prefix Block and Audit Log

**Linked REQs**: REQ-HRN-FND-006, REQ-HRN-FND-014

**Given** a harness lifecycle where an evolution proposal would target a path under a FROZEN prefix (`.claude/agents/moai/`, `.claude/skills/moai-`, `.claude/rules/moai/`, or `.moai/project/brand/`),

**When** the L1 Frozen Guard evaluates the proposal,

**Then** the modification MUST be blocked and a JSONL entry MUST be appended to `.moai/harness/learning-history/frozen-guard-violations.jsonl` containing at minimum: ISO-8601 timestamp, the attempted target path, the calling subagent or skill identifier, and a rejection rationale.

**And** the user MUST NOT receive an error; the rejection is silent except for the audit log.

**Verification**:
- Architecture-level: a stub test case in the harness workflow body simulates a violation attempt and asserts the audit log entry appears with the expected schema.
- Manual: inspect `.moai/harness/learning-history/frozen-guard-violations.jsonl` after a deliberate violation simulation; the entry MUST match the schema.

---

### AC-HRN-FND-006 — Pre-Modification Snapshot and Rollback

**Linked REQs**: REQ-HRN-FND-007, REQ-HRN-FND-008

**Given** a Tier-4 evolution proposal approved at the AskUserQuestion gate,

**When** the application step begins,

**Then** the system MUST create a snapshot directory at `.moai/harness/learning-history/snapshots/<ISO-DATE>/` containing byte-identical copies of every file the modification will touch, BEFORE any write occurs. The snapshot directory MUST include a `manifest.json` recording absolute target paths and content hashes.

**And When** the user later invokes `/moai:harness rollback <ISO-DATE>`,

**Then** the system MUST restore every file in the manifest to its byte-identical pre-modification state.

**And When** the user invokes `/moai:harness rollback <NONEXISTENT-DATE>`,

**Then** the system MUST return a diagnostic message identifying the missing snapshot and MUST NOT modify any file.

**Verification**:
- Manual: trigger a snapshot via a simulated Tier-4 application; verify the snapshot directory exists with the manifest. Invoke rollback; verify byte-identical restoration with `diff` or hash comparison.

---

### AC-HRN-FND-007 — Observer No-Op When Learning Disabled

**Linked REQs**: REQ-HRN-FND-009

**Given** a project with `.moai/config/sections/harness.yaml` containing `learning.enabled: false`,

**When** any tool invocation occurs that would normally trigger the PostToolUse observer,

**Then** the observer MUST NOT read, write, or append to `.moai/harness/usage-log.jsonl` and MUST NOT invoke any tier-classification or evolution logic.

**And** existing log entries MUST NOT be deleted.

**Verification**:
- Test: set `learning.enabled: false` in a test fixture, invoke a representative tool, then check that `.moai/harness/usage-log.jsonl` line count is unchanged.

---

### AC-HRN-FND-008 — PostToolUse Baseline and 4-Tier Ladder Preserved

**Linked REQs**: REQ-HRN-FND-010, REQ-HRN-FND-011

**Given** a project with `learning.enabled: true` (the default),

**When** an Edit or Write tool invocation completes,

**Then** the PostToolUse observer MUST append one JSONL line to `.moai/harness/usage-log.jsonl` containing at minimum: ISO-8601 timestamp, event_type, subject, and a context hash.

**And When** a unique pattern (defined by `event_type + subject + context_hash`) reaches cumulative observation count 1, 3, 5, or 10,

**Then** the tier classifier MUST classify the pattern as Observation, Heuristic, Rule, or Auto-update respectively, matching the thresholds defined in `SPEC-V3R3-HARNESS-LEARNING-001` REQ-HL-002.

**Verification**:
- Test: simulate 10 invocations of the same pattern in a test fixture; verify `.moai/harness/learning-history/tier-promotions.jsonl` records all four tier transitions.

---

### AC-HRN-FND-009 — Tier-4 Rate Limit of 1 per 7-Day Window

**Linked REQs**: REQ-HRN-FND-012, REQ-HRN-FND-018

**Given** a project where the orchestrator has invoked `AskUserQuestion` for a Tier-4 application within the last 7 days,

**When** the orchestrator encounters another Tier-4 candidate within the same 7-day rolling window,

**Then** the orchestrator MUST NOT invoke `AskUserQuestion` for the new candidate unless an adaptive expansion mechanism (from a downstream SPEC) has explicitly raised the limit.

**And** the new candidate MUST be deferred and recorded in `.moai/harness/proposals/` for later approval, with the deferral timestamp recorded.

**And** the rate-limit floor of 1 application per 7-day window MUST NOT be lowered by any downstream adaptive-expansion mechanism; tightening to zero (effectively disabling) is permitted via the existing `learning.enabled: false` setting.

**Verification**:
- Architecture-level: stub test simulates two Tier-4 candidates within 7 days; verify only the first triggers `AskUserQuestion`, the second is deferred.

---

### AC-HRN-FND-010 — Superseded V3R3 SPECs Not Modified in This PR

**Linked REQs**: REQ-HRN-FND-013

**Given** the pull request implementing this foundation SPEC,

**When** the PR's diff is examined,

**Then** the files `.moai/specs/SPEC-V3R3-HARNESS-001/spec.md`, `.moai/specs/SPEC-V3R3-HARNESS-LEARNING-001/spec.md`, and `.moai/specs/SPEC-V3R3-PROJECT-HARNESS-001/spec.md` MUST NOT appear in the diff. Status transition of these SPECs is a follow-up commit owned by `manager-git` after this SPEC merges.

**Verification**:
- Automated: `git diff main..HEAD --name-only | grep -E 'SPEC-V3R3-(HARNESS-001|HARNESS-LEARNING-001|PROJECT-HARNESS-001)/'` MUST return zero matches.
- Manual: PR reviewer confirms the three files are absent from the PR's file change list.

---

### AC-HRN-FND-011 — Success-Metric Telemetry Exposure

**Linked REQs**: REQ-HRN-FND-016

**Given** a project with a non-empty `.moai/harness/usage-log.jsonl` and at least one entry in `.moai/harness/learning-history/tier-promotions.jsonl`,

**When** the user invokes `/moai:harness status`,

**Then** the response MUST include two telemetry values:
1. Weekly Tier-4 application count (most recent 7-day rolling window from current timestamp).
2. Tier-4 reach rate (percentage of unique patterns that have been promoted from Tier 1 through Tier 4 across the lifetime of the project).

**And** both values MUST be derivable from the JSONL files alone, without invoking any Go binary.

**Verification**:
- Manual: in a Claude Code session, invoke `/moai:harness status`; observe both values in the response. Cross-check against direct file inspection.

---

### AC-HRN-FND-012 — Reflexion-Evaluator Conflict Resolution (Contract Assertion)

**Linked REQs**: REQ-HRN-FND-017

**Given** a future downstream SPEC implementing Reflexion-style verbal self-critique (SPEC-V3R4-HARNESS-004) and principle-based scoring (SPEC-V3R4-HARNESS-005),

**When** the self-critique result and the `evaluator-active` score disagree on an evolution proposal,

**Then** the system MUST treat the `evaluator-active` verdict as the binding gate and MUST treat the Reflexion pre-screen as advisory only.

**And** this foundation SPEC asserts the contract in REQ-HRN-FND-017; downstream SPECs MUST cite this REQ when introducing the Reflexion mechanism.

**Verification**:
- Architecture-level: this is a contract assertion, not a code path in the foundation SPEC. Verification is satisfied by REQ-HRN-FND-017's presence in this SPEC's frontmatter chain and by downstream SPECs citing it.
- The SPEC-V3R4-HARNESS-004 plan-auditor run will verify that the downstream implementation respects this contract.

---

## Edge Cases

### EDGE-001 — Snapshot disk-full at Tier-4 application time

If `.moai/harness/learning-history/snapshots/<ISO-DATE>/` cannot be created due to disk-full or permission errors, the application step MUST abort BEFORE any modification occurs. The user MUST be informed via the orchestrator's response. No partial state MUST be left on disk.

### EDGE-002 — Rollback during in-flight Tier-4 application

If the user invokes `/moai:harness rollback <date>` while a Tier-4 application is in progress, the rollback MUST wait for the application to complete or fail; concurrent rollback during application is not supported in this foundation SPEC.

### EDGE-003 — Frozen-guard violation log file permission

If `.moai/harness/learning-history/frozen-guard-violations.jsonl` cannot be written (permission, disk-full), the L1 Frozen Guard MUST still block the modification; the log emission is a best-effort audit and MUST NOT cascade the violation into the EVOLVABLE zone.

### EDGE-004 — `learning.enabled: false` with existing usage log entries

When `learning.enabled: false` is set on a project with existing entries in `.moai/harness/usage-log.jsonl`, the observer MUST become a no-op but MUST NOT delete the existing entries. The user retains the right to delete them manually or via a downstream maintenance verb (out of scope here).

### EDGE-005 — Slash command invoked outside Claude Code session

The `/moai:harness` slash command is only meaningful inside a Claude Code session. If a user attempts a similar pattern in a non-Claude-Code context (e.g., a shell hook), no behavior is defined; the foundation SPEC does not address terminal-only invocation of the slash command.

### EDGE-006 — Concurrent `/moai:harness apply` invocations

Two concurrent `/moai:harness apply` invocations within the same session are out of scope. The orchestrator's AskUserQuestion contract is sequential by design. The foundation SPEC does not introduce concurrency control beyond this.

---

## Test Scenarios (Given-When-Then)

### Scenario 1: User Upgrades from V3R3 to V3R4

**Given** a user on moai-adk-go v2.19.0 with `.moai/harness/usage-log.jsonl` containing 152 bytes of accumulated observation data and one Tier-2 heuristic in `.moai/harness/learning-history/tier-promotions.jsonl`,

**When** the user upgrades to v2.20.0 (the target release for SPEC-V3R4-HARNESS-001) and starts a new Claude Code session,

**Then** the slash command `/moai:harness status` MUST work without errors and MUST display the existing observation count and tier distribution.

**And When** the user runs `moai harness status` from the terminal,

**Then** the response MUST be `unknown command "harness" for "moai"` with a non-zero exit code.

**And** existing `.moai/harness/usage-log.jsonl` entries MUST NOT be migrated, deleted, or modified.

---

### Scenario 2: Tier-4 Approval Happy Path

**Given** a project with a single pending Tier-4 proposal at `.moai/harness/proposals/2026-05-14-tier4-001.json` and no Tier-4 application has occurred in the last 7 days,

**When** the user invokes `/moai:harness apply`,

**Then** the MoAI orchestrator MUST invoke `AskUserQuestion` presenting the proposal with at least three options (Apply (권장), Defer, Reject).

**And When** the user selects "Apply",

**Then** the system MUST create a snapshot at `.moai/harness/learning-history/snapshots/2026-05-14T<HH:MM:SS>Z/`, apply the modification, and emit a confirmation message.

**And** the proposal file MUST be moved from `.moai/harness/proposals/` to `.moai/harness/learning-history/applied/`.

---

### Scenario 3: Tier-4 Frozen-Guard Rejection

**Given** a malformed Tier-4 proposal that would modify `.claude/rules/moai/design/constitution.md` (a FROZEN path),

**When** the orchestrator processes the proposal,

**Then** the L1 Frozen Guard MUST block the application BEFORE `AskUserQuestion` is invoked.

**And** a JSONL entry MUST be appended to `.moai/harness/learning-history/frozen-guard-violations.jsonl` with timestamp, target path (`.claude/rules/moai/design/constitution.md`), and rejection rationale.

**And** the user MUST NOT receive an error; the rejection is silent except for the audit log.

---

### Scenario 4: Rollback Restores Pre-Application State

**Given** a Tier-4 application completed at ISO date `2026-05-14T10:00:00Z` that modified `.claude/skills/my-harness-cli-template/SKILL.md` (a 4500-byte file with content-hash `abc123...`),

**When** the user invokes `/moai:harness rollback 2026-05-14T10:00:00Z`,

**Then** the system MUST restore `.claude/skills/my-harness-cli-template/SKILL.md` to its byte-identical pre-application state (content-hash `abc123...` restored).

**And** the system MUST emit a confirmation message naming the snapshot date and the restored file count.

---

### Scenario 5: AskUserQuestion Rate-Limit Deferral

**Given** the orchestrator has invoked `AskUserQuestion` for a Tier-4 approval at `2026-05-12T15:00:00Z` (2 days ago) and a second Tier-4 candidate becomes available at `2026-05-14T18:00:00Z`,

**When** the orchestrator evaluates whether to invoke `AskUserQuestion` for the second candidate,

**Then** the orchestrator MUST NOT invoke `AskUserQuestion` (the 7-day rolling window has not elapsed).

**And** the new candidate MUST be recorded in `.moai/harness/proposals/` with the deferral timestamp.

**And** when the next session occurs after `2026-05-19T15:00:00Z`, the deferred candidate MUST be eligible for AskUserQuestion presentation.

---

### Scenario 6: Subagent Blocker Report Instead of AskUserQuestion

**Given** a future downstream SPEC implementation where `manager-spec` is delegated to draft a new harness skill body,

**When** the subagent encounters ambiguity that requires user input (e.g., choice between two skill body templates),

**Then** the subagent MUST NOT invoke `AskUserQuestion`. Instead, it MUST return a structured blocker report:

```markdown
## Missing Inputs

| Parameter | Type | Expected Values | Rationale |
|-----------|------|-----------------|-----------|
| skill_template | enum | template-a | template-b | Body structure choice |

**Blocker**: Cannot proceed without selecting skill template.
```

**And** the orchestrator MUST receive the blocker report, invoke `ToolSearch(query: "select:AskUserQuestion")`, then `AskUserQuestion` with the two skill-template options (template-a (권장), template-b), then re-delegate to the subagent with the user's selection injected into the spawn prompt.

---

### Scenario 7: CI Guard Detects Re-Registration Attempt

**Given** a hypothetical pull request that adds the line `rootCmd.AddCommand(newHarnessCmd())` to `internal/cli/root.go`,

**When** the continuous integration system runs the test suite,

**Then** a Go test (e.g., `internal/cli/harness_retirement_test.go`) MUST fail with a diagnostic referencing `SPEC-V3R4-HARNESS-001` and `REQ-HRN-FND-002`, blocking the PR from merging.

---

## Definition of Done

This SPEC is considered complete when all of the following are satisfied:

1. **All four SPEC artifacts present**: `spec.md`, `plan.md`, `acceptance.md`, `tasks.md` exist under `.moai/specs/SPEC-V3R4-HARNESS-001/` with the canonical 9-field frontmatter on `spec.md`.
2. **All 18 REQs in EARS format**: every REQ in `spec.md` §5 uses one of the five EARS patterns.
3. **All 12 ACs verifiable**: every AC in `acceptance.md` includes Given-When-Then plus a Verification section with concrete commands or inspection steps.
4. **Coverage map complete**: every REQ appears in at least one AC; every AC links back to at least one REQ.
5. **No tech-stack implementation in spec.md**: implementation choices (e.g., specific test file paths, programming-language choices) are in `plan.md` only.
6. **No modification of FROZEN files**: `.claude/rules/moai/design/constitution.md` and the three superseded V3R3 SPECs unchanged in this PR.
7. **No emojis in artifact content**.
8. **No time estimates**: priority labels (P0-P3) and phase ordering only.
9. **Plan-auditor PASS**: independent subagent audit returns PASS verdict (or unresolvable findings escalated as blocker).
10. **PR delegated to `manager-git`**: this manager-spec session does NOT create the PR directly; it delegates to `manager-git` via `Agent()` with the prepared title and body.
11. **Conventional Commits**: commit message follows `plan(SPEC-V3R4-HARNESS-001): foundation SPEC for self-evolving harness v2` format.
12. **BREAKING change ID**: `BC-V3R4-HARNESS-001-CLI-RETIREMENT` declared in `bc_id:` frontmatter.

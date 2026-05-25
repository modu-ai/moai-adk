---
id: SPEC-V3R6-AGENT-TEAM-REBUILD-001
artifact: acceptance
version: "0.1.0"
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
sync_commit_sha: "f0f222fa3"
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial Tier L acceptance criteria authored — 22 AC-ATR-XXX covering 20 REQ-ATR-XXX at 100% traceability; 8 edge cases in §C; Quality Gate / DoD in §D. |

---

## §A — Acceptance Criteria Matrix (AC-ATR-XXX, ≥20 mandatory)

Each AC follows Given-When-Then format with severity (CRITICAL / HIGH / MEDIUM / LOW), evidence command (executable shell command for verification), pass criterion (binary observable outcome), and REQ-ATR-XXX traceability mapping.

### AC-ATR-001 — Catalog consists of exactly 8 retained agents

| Field | Value |
|-------|-------|
| **Severity** | CRITICAL |
| **Given** | Run-phase M3 (Agent Archive) is complete and the agent files have been moved to `.moai/backups/agent-archive-2026-05-25/`. |
| **When** | The orchestrator enumerates the active agent catalog by listing `.claude/agents/{core,meta,expert,agency}/*.md`. |
| **Then** | Exactly 7 MoAI-custom retained agent files exist (`manager-spec`, `manager-develop`, `manager-docs`, `manager-git` under `core/`; `plan-auditor`, `evaluator-active`, `builder-harness` under `meta/`). The Anthropic built-in `Explore` is documented in `agent-patterns.md` but is not a MoAI file. |
| **Evidence** | `ls .claude/agents/core/*.md .claude/agents/meta/*.md .claude/agents/expert/*.md .claude/agents/agency/*.md 2>/dev/null \| wc -l` returns 7. |
| **Pass criterion** | Exit code 0 AND output `7`. |
| **Maps to** | REQ-ATR-001 |

### AC-ATR-002 — Each retained agent body ≤ 500 lines

| Field | Value |
|-------|-------|
| **Severity** | HIGH |
| **Given** | Run-phase M1 (Agent retention frontmatter refinement) is complete. |
| **When** | The orchestrator measures the body line count of each of the 7 MoAI-custom retained agent files. |
| **Then** | Each file has `wc -l` output ≤ 500. If any file exceeds, M1 commit must document the overage as known debt with future SPEC reference. |
| **Evidence** | `wc -l .claude/agents/core/*.md .claude/agents/meta/*.md \| awk '$1 > 500 {print}'` returns no rows (excluding the total line). |
| **Pass criterion** | No output from `awk '$1 > 500'` filter. |
| **Maps to** | REQ-ATR-002 |

### AC-ATR-003 — Each retained agent declares explicit `tools:` CSV

| Field | Value |
|-------|-------|
| **Severity** | HIGH |
| **Given** | Run-phase M1 is complete. |
| **When** | The orchestrator inspects the YAML frontmatter of each retained agent file. |
| **Then** | Every retained agent file has a non-empty `tools:` key in its YAML frontmatter with a comma-separated string value (per agent-authoring.md frontmatter rule). |
| **Evidence** | `for f in .claude/agents/core/*.md .claude/agents/meta/*.md; do grep -q '^tools:' "$f" \|\| echo "MISSING: $f"; done` returns no `MISSING` lines. |
| **Pass criterion** | No `MISSING` output. |
| **Maps to** | REQ-ATR-003 |

### AC-ATR-004 — Each retained agent declares explicit `NOT-for:` clause

| Field | Value |
|-------|-------|
| **Severity** | MEDIUM |
| **Given** | Run-phase M1 is complete. |
| **When** | The orchestrator inspects each retained agent file's description field. |
| **Then** | Each retained agent file's description text or a dedicated `NOT-for:` line enumerates use cases where the agent is **not** the correct delegation target. |
| **Evidence** | `for f in .claude/agents/core/*.md .claude/agents/meta/*.md; do grep -q -i 'NOT-for\|NOT for\|out of scope' "$f" \|\| echo "MISSING: $f"; done` returns no `MISSING` lines. |
| **Pass criterion** | No `MISSING` output. |
| **Maps to** | REQ-ATR-004 |

### AC-ATR-005 — 12 phantom and domain-expert agents archived to `.moai/backups/agent-archive-2026-05-25/`

| Field | Value |
|-------|-------|
| **Severity** | CRITICAL |
| **Given** | Run-phase M3 (Agent Archive) is complete. |
| **When** | The orchestrator inspects the archive directory structure and original paths. |
| **Then** | The archive directory contains the 12 (or 11 if researcher.md was already absent) archived files preserving original substructure (`core/`, `meta/`, `expert/`, `agency/`). The original `.claude/agents/` paths no longer contain these files. |
| **Evidence** | `ls .moai/backups/agent-archive-2026-05-25/**/*.md \| grep -E "manager-(strategy\|quality\|brain\|project)\|claude-code-guide\|expert-(backend\|frontend\|security\|devops\|performance\|refactoring)" \| wc -l` returns 11 (without researcher.md) or 12 (with). |
| **Pass criterion** | Count is 11 or 12 AND `ls .claude/agents/core/manager-strategy.md 2>&1 \| grep -c "No such file"` = 1. |
| **Maps to** | REQ-ATR-005 |

### AC-ATR-006 — Predecessor SPEC frontmatter transitioned to `superseded`

| Field | Value |
|-------|-------|
| **Severity** | CRITICAL |
| **Given** | Run-phase M6 (Predecessor SPEC supersedence) is complete. |
| **When** | The orchestrator inspects the predecessor SPEC's spec.md frontmatter. |
| **Then** | The predecessor SPEC's `status` field equals `superseded`, `updated` equals `2026-05-25`, and HISTORY contains a v0.1.2 row referencing this SPEC. |
| **Evidence** | `grep -E '^(status\|updated):' .moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/spec.md \| head -2` and `grep -c "Superseded by SPEC-V3R6-AGENT-TEAM-REBUILD-001" .moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/spec.md`. |
| **Pass criterion** | First grep returns `status: superseded` AND `updated: 2026-05-25`; second grep returns ≥ 1. |
| **Maps to** | REQ-ATR-006 |

### AC-ATR-007 — Skill router invocation discipline preserved

| Field | Value |
|-------|-------|
| **Severity** | HIGH |
| **Given** | Run-phase M2 (Workflow router skill phase-owner declarations) is complete. |
| **When** | The orchestrator inspects `.claude/rules/moai/workflow/session-handoff.md` Block 5 documentation. |
| **Then** | Block 5 documentation explicitly states `/moai <subcommand>` triggers Skill router invocation; manual SKILL.md body Read is prohibited. |
| **Evidence** | `grep -A 5 "Block 5" .claude/rules/moai/workflow/session-handoff.md \| grep -c -i "Skill router\|Skill("` returns ≥ 1. |
| **Pass criterion** | grep count ≥ 1. |
| **Maps to** | REQ-ATR-007 |

### AC-ATR-008 — Phase 0.95 Mode Selection logging in progress.md

| Field | Value |
|-------|-------|
| **Severity** | HIGH |
| **Given** | Run-phase M2 + M5 are complete and progress.md has been initialized. |
| **When** | The orchestrator inspects the SPEC's progress.md for § Mode Selection section. |
| **Then** | progress.md contains a `## § Mode Selection` heading with the chosen mode (one of: trivial, background, agent-team, parallel, sub-agent) and a rationale paragraph citing the decision tree in design.md §B.4. |
| **Evidence** | `grep -A 5 "Mode Selection" .moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/progress.md \| grep -c -i "sequential\|parallel\|agent-team\|sub-agent\|trivial\|background"` returns ≥ 1. |
| **Pass criterion** | grep count ≥ 1. |
| **Maps to** | REQ-ATR-008 |

### AC-ATR-009 — Stop hook verifies lint + test + coverage on sync-phase commit

| Field | Value |
|-------|-------|
| **Severity** | CRITICAL |
| **Given** | Run-phase M4 (Hook scripts authoring) is complete and `.claude/hooks/moai/sync-phase-quality-gate.sh` exists. |
| **When** | The hook is invoked with simulated sync-phase commit stdin payload AND a deliberately failing condition (e.g., a planted lint error). |
| **Then** | The hook exits with code 2 (block) and outputs structured rejection JSON identifying the failed check. |
| **Evidence** | `bash .claude/hooks/moai/sync-phase-quality-gate.sh < .moai/tests/hooks/fixtures/sync-fail.json; echo "exit:$?"` returns `exit:2`. |
| **Pass criterion** | Exit code is 2 (if test fixture present); OR (fallback) `bash -n .claude/hooks/moai/sync-phase-quality-gate.sh; echo "syntax:$?"` returns `syntax:0` AND the script body contains `golangci-lint`, `go test`, and coverage-delta verification keywords. |
| **Maps to** | REQ-ATR-009 |

### AC-ATR-010 — manager-develop cycle_type=ddd uses ANALYZE-PRESERVE-IMPROVE

| Field | Value |
|-------|-------|
| **Severity** | MEDIUM |
| **Given** | Run-phase M1 + M2 are complete; manager-develop.md frontmatter / body references cycle_type modes. |
| **When** | The orchestrator inspects the manager-develop agent body and the spec-workflow.md DDD Mode reference. |
| **Then** | manager-develop.md body documents `cycle_type=ddd` invokes ANALYZE-PRESERVE-IMPROVE per spec-workflow.md §Run Phase DDD Mode. |
| **Evidence** | `grep -c "ANALYZE-PRESERVE-IMPROVE\|cycle_type.*ddd" .claude/agents/core/manager-develop.md` returns ≥ 1. |
| **Pass criterion** | grep count ≥ 1. |
| **Maps to** | REQ-ATR-010 |

### AC-ATR-011 — manager-develop cycle_type=tdd uses RED-GREEN-REFACTOR

| Field | Value |
|-------|-------|
| **Severity** | MEDIUM |
| **Given** | Run-phase M1 + M2 are complete. |
| **When** | The orchestrator inspects the manager-develop agent body and the spec-workflow.md TDD Mode reference. |
| **Then** | manager-develop.md body documents `cycle_type=tdd` invokes RED-GREEN-REFACTOR per spec-workflow.md §Run Phase TDD Mode. |
| **Evidence** | `grep -c "RED-GREEN-REFACTOR\|cycle_type.*tdd" .claude/agents/core/manager-develop.md` returns ≥ 1. |
| **Pass criterion** | grep count ≥ 1. |
| **Maps to** | REQ-ATR-011 |

### AC-ATR-012 — manager-develop cycle_type=autofix uses DIAGNOSE-PATCH-VERIFY with max 3 iterations

| Field | Value |
|-------|-------|
| **Severity** | HIGH |
| **Given** | Run-phase M1 + M2 + M5 are complete; cycle_type=autofix mode is the NEW addition. |
| **When** | The orchestrator inspects the manager-develop agent body for cycle_type=autofix mode and the ci-autofix-protocol.md cross-reference. |
| **Then** | manager-develop.md body documents `cycle_type=autofix` invokes DIAGNOSE-PATCH-VERIFY with max-3-iteration contract per `.claude/rules/moai/workflow/ci-autofix-protocol.md`. |
| **Evidence** | `grep -c "DIAGNOSE-PATCH-VERIFY\|cycle_type.*autofix\|ci-autofix-protocol" .claude/agents/core/manager-develop.md` returns ≥ 2 (mode name + protocol reference). |
| **Pass criterion** | grep count ≥ 2. |
| **Maps to** | REQ-ATR-012 |

### AC-ATR-013 — Agent Teams gated by harness thorough + team.enabled + env var

| Field | Value |
|-------|-------|
| **Severity** | MEDIUM |
| **Given** | Run-phase M5 (Rule files updates) is complete and `orchestration-mode-selection.md` exists. |
| **When** | The orchestrator inspects the new rule file for Agent Teams capability gate. |
| **Then** | `orchestration-mode-selection.md` documents the 3-condition capability gate: `harness: thorough` AND `workflow.team.enabled: true` AND `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`. |
| **Evidence** | `grep -c "thorough.*team.enabled\|CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS" .claude/rules/moai/workflow/orchestration-mode-selection.md` returns ≥ 1. |
| **Pass criterion** | grep count ≥ 1. |
| **Maps to** | REQ-ATR-013 |

### AC-ATR-014 — Stop hook dependency manifest audit on go.mod / package-lock.json / etc.

| Field | Value |
|-------|-------|
| **Severity** | HIGH |
| **Given** | Run-phase M4 is complete and `.claude/hooks/moai/sync-phase-quality-gate.sh` exists. |
| **When** | The orchestrator inspects the hook script body for dependency manifest audit logic. |
| **Then** | The hook script body contains conditional logic detecting changes to `go.mod`, `go.sum`, `package-lock.json`, `Pipfile.lock`, or `Cargo.lock`, and invokes the corresponding manifest audit tool (`govulncheck`, `npm audit`, etc.). |
| **Evidence** | `grep -c "go.mod\|govulncheck\|package-lock\|npm audit" .claude/hooks/moai/sync-phase-quality-gate.sh` returns ≥ 2. |
| **Pass criterion** | grep count ≥ 2. |
| **Maps to** | REQ-ATR-014 |

### AC-ATR-015 — GATE-2 mandatory restoration on autonomous-flow skip detection

| Field | Value |
|-------|-------|
| **Severity** | CRITICAL |
| **Given** | Run-phase M7 (CLAUDE.md + CLAUDE.local.md doctrine updates) is complete. |
| **When** | The orchestrator inspects `CLAUDE.local.md` §19 AskUserQuestion Enforcement Protocol for GATE-2 mandatory restoration. |
| **Then** | §19 (or §23 cross-reference) explicitly states that the orchestrator must halt autonomous flow on detected GATE-2 skip and trigger `AskUserQuestion` round; the `score ≥ 0.90 skip-eligible` autonomous bypass policy applies only to Phase 0.5 plan-auditor, not to GATE-2. |
| **Evidence** | `grep -c "GATE-2\|gate-2\|Ctrl+G\|HUMAN GATE.*mandatory" CLAUDE.local.md` returns ≥ 1; `grep -A 5 "skip-eligible" CLAUDE.local.md \| grep -c -i "Phase 0.5\|plan-auditor"` returns ≥ 1. |
| **Pass criterion** | Both grep counts ≥ 1. |
| **Maps to** | REQ-ATR-015 |

### AC-ATR-016 — ARCHIVED_AGENT_REJECTED error spec + migration table

| Field | Value |
|-------|-------|
| **Severity** | HIGH |
| **Given** | Run-phase M5 is complete and `archived-agent-rejection.md` exists. |
| **When** | The orchestrator inspects the new rule file for ARCHIVED_AGENT_REJECTED error specification and migration table. |
| **Then** | The rule file enumerates all 12 archived agents with per-agent replacement-pattern guidance (e.g., `manager-strategy → manager-spec`, `expert-backend → Agent(general-purpose, ...)` with backend whitelist). |
| **Evidence** | `grep -c "ARCHIVED_AGENT_REJECTED" .claude/rules/moai/workflow/archived-agent-rejection.md` ≥ 1; `grep -c "manager-strategy\|manager-quality\|expert-backend\|expert-frontend\|expert-security\|expert-devops\|expert-performance\|expert-refactoring\|manager-brain\|manager-project\|claude-code-guide\|researcher" .claude/rules/moai/workflow/archived-agent-rejection.md` ≥ 12. |
| **Pass criterion** | First grep ≥ 1; second grep ≥ 12. |
| **Maps to** | REQ-ATR-016 |

### AC-ATR-017 — Multi-domain compound mode preference (Agent Teams or parallel multi-spawn)

| Field | Value |
|-------|-------|
| **Severity** | MEDIUM |
| **Given** | Run-phase M5 is complete and `orchestration-mode-selection.md` documents the 5-mode decision tree. |
| **When** | The orchestrator inspects the rule file for the Compound clause logic. |
| **Then** | The rule file documents: harness `standard` or `thorough` AND scope `≥3 domains OR ≥10 files` AND Phase 0.95 mode selection → prefer Agent Teams if REQ-ATR-013 prerequisites met, else fall back to parallel multi-spawn (max 3-5 concurrent). |
| **Evidence** | `grep -A 10 "Compound\|multi-domain" .claude/rules/moai/workflow/orchestration-mode-selection.md \| grep -c "Agent Teams\|parallel multi-spawn\|3-5"` returns ≥ 1. |
| **Pass criterion** | grep count ≥ 1. |
| **Maps to** | REQ-ATR-017 |

### AC-ATR-018 — Template-First parity (byte-for-byte mirror)

| Field | Value |
|-------|-------|
| **Severity** | CRITICAL |
| **Given** | Run-phase M8 (Template-First parity check) is complete. |
| **When** | The orchestrator runs `diff -r` between local `.claude/` paths and the corresponding template paths in `internal/template/templates/.claude/`. |
| **Then** | The diff is empty for in-scope paths (agents, workflow skills, modified rule files, hook scripts). Any intended drift (e.g., archive directory mirroring decision per §D.11 step 2) must be explicitly documented. |
| **Evidence** | `diff -rq .claude/agents/ internal/template/templates/.claude/agents/ \| wc -l` returns 0 (excluding the archive directory which may have intentional asymmetry). |
| **Pass criterion** | diff line count = 0 for retained agents AND workflow skills AND modified rule files AND hook scripts. |
| **Maps to** | REQ-ATR-018 |

### AC-ATR-019 — NOTICE.md Anthropic 2026 attribution citation

| Field | Value |
|-------|-------|
| **Severity** | MEDIUM |
| **Given** | Run-phase M7 is complete. |
| **When** | The orchestrator inspects `NOTICE.md` for the Anthropic 2026 alignment section. |
| **Then** | `NOTICE.md` contains a new section citing the Audit 3 verbatim Findings A1-A6 with source URLs (claude.com/docs/en/sub-agents, claude.com/docs/en/agent-teams, anthropic.com/engineering/built-multi-agent-research-system, etc.) and the archive date 2026-05-25. |
| **Evidence** | `grep -c "Anthropic 2026\|Audit 3\|2026-05-25.*archive\|claude.com/docs/en/sub-agents" NOTICE.md` returns ≥ 2. |
| **Pass criterion** | grep count ≥ 2. |
| **Maps to** | REQ-ATR-019 |

### AC-ATR-020 — manager-git PR doctrine consolidation (Tier L OR --pr flag)

| Field | Value |
|-------|-------|
| **Severity** | HIGH |
| **Given** | Run-phase M7 is complete. |
| **When** | The orchestrator inspects `CLAUDE.local.md` §23 and `.claude/skills/moai/workflows/sync.md` for manager-git PR doctrine. |
| **Then** | Both files consistently document: Tier S/M = main-direct push (Hybrid Trunk default); Tier L OR explicit `--pr` flag = PR via manager-git on `feat/SPEC-XXX` branch. |
| **Evidence** | `grep -c "Tier L.*--pr\|--pr.*manager-git\|Tier L OR.*pr" CLAUDE.local.md .claude/skills/moai/workflows/sync.md` returns ≥ 2 (at least one match in each file). |
| **Pass criterion** | grep total count ≥ 2 across both files. |
| **Maps to** | REQ-ATR-020 |

### AC-ATR-021 — CLAUDE.md §4 Agent Catalog reduced from 17 to 8

| Field | Value |
|-------|-------|
| **Severity** | HIGH |
| **Given** | Run-phase M7 is complete. |
| **When** | The orchestrator inspects `CLAUDE.md` §4 Agent Catalog section. |
| **Then** | §4 enumerates the 7 MoAI-custom retained agents + 1 Anthropic built-in `Explore` reference (total 8 entries). The 12 archived agents are NOT listed in the active catalog (may be cross-referenced as "archived" in a sub-section if desired). |
| **Evidence** | `grep -A 30 "## 4. Agent Catalog" CLAUDE.md \| grep -c "manager-strategy\|manager-quality\|manager-brain\|manager-project\|claude-code-guide\|expert-backend\|expert-frontend\|expert-security\|expert-devops\|expert-performance\|expert-refactoring" \| awk '{if ($1 == 0) print "PASS"; else print "FAIL count=" $1}'` returns `PASS`. |
| **Pass criterion** | Output is `PASS`. |
| **Maps to** | REQ-ATR-001 (CLAUDE.md catalog enumeration aspect) |

### AC-ATR-022 — Hook script subagent boundary discipline (no AskUserQuestion in hook scripts)

| Field | Value |
|-------|-------|
| **Severity** | CRITICAL |
| **Given** | Run-phase M4 is complete and the 3 NEW hook scripts exist. |
| **When** | The orchestrator scans the hook script bodies for AskUserQuestion violations. |
| **Then** | None of the 3 hook scripts invoke `AskUserQuestion` directly (subagent boundary per `.claude/rules/moai/core/askuser-protocol.md`). Hooks return exit codes only; the orchestrator translates exit codes into user-facing prompts. |
| **Evidence** | `grep -rn 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/ \| grep -v "^[^:]*:[0-9]*:[ \t]*#"` returns no matches. |
| **Pass criterion** | grep output is empty. |
| **Maps to** | REQ-ATR-009 + REQ-ATR-014 (boundary discipline aspect) + cross-reference C-HRA-008 subagent boundary canary |

---

## §B — Traceability Matrix (REQ-ATR ↔ AC-ATR, 100% coverage)

| REQ-ATR | GEARS pattern | AC-ATR coverage | Severity |
|---------|---------------|-----------------|----------|
| REQ-ATR-001 | Ubiquitous | AC-ATR-001, AC-ATR-021 | CRITICAL + HIGH |
| REQ-ATR-002 | Ubiquitous | AC-ATR-002 | HIGH |
| REQ-ATR-003 | Ubiquitous | AC-ATR-003 | HIGH |
| REQ-ATR-004 | Ubiquitous | AC-ATR-004 | MEDIUM |
| REQ-ATR-005 | Event-driven `When` | AC-ATR-005 | CRITICAL |
| REQ-ATR-006 | Event-driven `When` | AC-ATR-006 | CRITICAL |
| REQ-ATR-007 | Event-driven `When` | AC-ATR-007 | HIGH |
| REQ-ATR-008 | Event-driven `When` | AC-ATR-008 | HIGH |
| REQ-ATR-009 | Event-driven `When` | AC-ATR-009, AC-ATR-022 | CRITICAL |
| REQ-ATR-010 | State-driven `While` | AC-ATR-010 | MEDIUM |
| REQ-ATR-011 | State-driven `While` | AC-ATR-011 | MEDIUM |
| REQ-ATR-012 | State-driven `While` | AC-ATR-012 | HIGH |
| REQ-ATR-013 | Capability-gate `Where` | AC-ATR-013 | MEDIUM |
| REQ-ATR-014 | Capability-gate `Where` | AC-ATR-014, AC-ATR-022 | HIGH + CRITICAL |
| REQ-ATR-015 | Event-detected | AC-ATR-015 | CRITICAL |
| REQ-ATR-016 | Event-detected | AC-ATR-016 | HIGH |
| REQ-ATR-017 | Compound | AC-ATR-017 | MEDIUM |
| REQ-ATR-018 | Ubiquitous | AC-ATR-018 | CRITICAL |
| REQ-ATR-019 | Ubiquitous | AC-ATR-019 | MEDIUM |
| REQ-ATR-020 | Compound | AC-ATR-020 | HIGH |

**Coverage**: 20 REQ-ATR-XXX → 22 AC-ATR-XXX (some ACs cover multiple REQs). 100% REQ→AC coverage achieved.

**GEARS pattern distribution**: 6 Ubiquitous + 5 Event-driven `When` + 3 State-driven `While` + 2 Capability-gate `Where` + 2 Event-detected + 2 Compound = 20. GEARS density 100% (no legacy IF/THEN modality).

---

## §C — Edge Cases (8 mandatory)

### EC-ATR-001 — Spawn of archived agent at runtime

**Scenario**: User or paste-ready resume invokes `Agent(subagent_type="manager-strategy", ...)` after M3 archive is complete.

**Expected behavior**: orchestrator detects the archived subagent_type, emits `ARCHIVED_AGENT_REJECTED` error referencing the migration table in `archived-agent-rejection.md`, halts spawn, and prompts user via `AskUserQuestion` for the replacement decision (e.g., "Switch to manager-spec? Switch to general-purpose with backend whitelist? Abort?").

**Verification**: AC-ATR-016 covers the rule file specification; runtime enforcement is orchestrator-discipline + future Go-side enforcement (deferred to follow-up SPEC).

### EC-ATR-002 — Stop hook false positive (legitimate sync commit fails quality gate)

**Scenario**: Sync-phase commit triggers Stop hook; hook detects pre-existing lint baseline failure unrelated to the current commit's scope.

**Expected behavior**: hook supports `--baseline-mode` flag reading from `.moai/state/lint-baseline.json`. If lint failure matches baseline, hook permits commit with warning. If failure is NEW (not in baseline), hook blocks via exit 2.

**Verification**: hook script body documents baseline-mode logic; AC-ATR-009 binary check passes when scripted correctly.

### EC-ATR-003 — Template-local drift after manual edit

**Scenario**: Developer manually edits `.claude/agents/core/manager-develop.md` without mirroring to `internal/template/templates/`.

**Expected behavior**: `make build` step in M8 re-generates `internal/template/embedded.go`; subsequent `diff -r` reveals drift; manager-develop in M8 must reconcile by copying local to template OR returning blocker report if drift is intended.

**Verification**: AC-ATR-018 byte-for-byte parity binary check.

### EC-ATR-004 — Multi-session race during M3 archive

**Scenario**: While manager-develop M3 is moving agent files, a parallel session attempts a commit on `.claude/agents/`.

**Expected behavior**: per CLAUDE.local.md §23.8 Multi-Session Race Mitigation, pre-spawn fetch obligation + path-specific staging + post-action fetch verify catches the race. If detected, manager-develop returns blocker report; orchestrator runs AskUserQuestion (rebase / inspect / abort).

**Verification**: B12-flavored sentinel — `git fetch origin && git rev-list --count --left-right origin/main...HEAD` returns `0 0` after M3 commit.

### EC-ATR-005 — Mode selection ambiguity at thresholds

**Scenario**: Phase 0.95 Mode Selection encounters a scope at boundary thresholds (e.g., 9-file scope vs 10-file scope; 3-domain vs 2-domain).

**Expected behavior**: per design.md §B.4 decision tree tie-breaker rules — default to the simpler mode (sub-agent over agent-team; sequential over parallel) and log the boundary case to `progress.md § Mode Selection` for retrospective analysis.

**Verification**: AC-ATR-008 (Mode Selection logging) + design.md §B.4 documentation.

### EC-ATR-006 — Agent Teams prerequisites missing under harness thorough

**Scenario**: SPEC harness level is `thorough` but `workflow.team.enabled: false` or `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` env var unset.

**Expected behavior**: orchestrator falls back to parallel multi-spawn (max 3-5 concurrent retained agents) per REQ-ATR-017 fallback path. Emit `[mode-auto-downgrade]` info log per existing WF-003 sentinel pattern.

**Verification**: AC-ATR-017 documents the fallback; existing `MODE_TEAM_UNAVAILABLE` sentinel in `spec-workflow.md` covers runtime detection.

### EC-ATR-007 — Predecessor SPEC was already partially closed before M6 supersedence

**Scenario**: Between this SPEC's authoring and M6 run-phase execution, the predecessor SPEC's status field advanced from `draft` to `in-progress` or `implemented` by some other session.

**Expected behavior**: M6 supersedence transition `* → superseded` per Status Transition Ownership Matrix is valid from ANY current status (the `*` wildcard explicitly allows transition from any prior state). manager-spec (the owner) records the original-status-at-supersedence in the HISTORY v0.1.2 entry for audit trail.

**Verification**: AC-ATR-006 binary check on final state regardless of prior status.

### EC-ATR-008 — Paste-ready resume references archived agent

**Scenario**: A paste-ready resume in `~/.claude/projects/{hash}/memory/` references an archived agent (e.g., "applied lessons: manager-quality coverage audit").

**Expected behavior**: orchestrator detects the archived-agent reference, emits informational notice (not an error — auto-memory may legitimately reference historical work), and proceeds without spawning the archived agent. If the resume requires the archived agent's function, orchestrator triggers AskUserQuestion to elicit the replacement-pattern decision per the migration table.

**Verification**: cross-reference `archived-agent-rejection.md` migration table coverage (AC-ATR-016) + AskUserQuestion enforcement (REQ-ATR-015 / AC-ATR-015).

---

## §D — Quality Gate + Definition of Done

### §D.1 Plan-phase Quality Gate (current phase)

**Must-pass** (binary, all required):
- [ ] spec.md 12 canonical frontmatter fields present with correct values
- [ ] 20 REQ-ATR-XXX in spec.md §D with ≥80% GEARS notation
- [ ] 22 AC-ATR-XXX in this acceptance.md §A with 100% REQ→AC traceability (§B)
- [ ] 8 edge cases in §C
- [ ] 8 risks in spec.md §G each paired with mitigation strategy
- [ ] plan.md 8 milestones M1-M8 with REQ coverage mapping
- [ ] design.md §B Target Architecture + §D Hook Architecture present
- [ ] research.md §H Audit 3 synthesis appended
- [ ] plan-auditor verdict PASS ≥ 0.85 (Tier L threshold)

**Preferred (not blocking)**:
- [ ] plan-auditor verdict skip-eligible ≥ 0.90 (Tier L MARGINAL → skip-eligible upgrade)

### §D.2 Run-phase Definition of Done

**Must-pass** (all required):
- [ ] All 8 milestones (M1-M8) completed with documented commit SHAs in progress.md §E.2
- [ ] All 22 ACs PASS per evidence commands above
- [ ] 7-item Trust-but-verify batch passes (test + coverage + boundary grep + sentinel scan + CLI smoke + template parity + lint)
- [ ] 12 archived agents present in `.moai/backups/agent-archive-2026-05-25/`
- [ ] 0 archived-agent references in `.claude/agents/`, `.claude/skills/moai/workflows/`, `internal/template/templates/.claude/agents/`
- [ ] 3 NEW hook scripts present with bash syntax check exit 0
- [ ] Template-First parity diff empty for in-scope paths

### §D.3 Sync-phase Definition of Done

**Must-pass**:
- [ ] CHANGELOG.md entry for SPEC-V3R6-AGENT-TEAM-REBUILD-001 added under `[Unreleased]` (duplicate-check via `grep -c` per B12)
- [ ] All 5 SPEC artifact frontmatter `status: in-progress → implemented` (spec/plan/acceptance/design/research)
- [ ] `sync_commit_sha:` populated in all 5 frontmatter
- [ ] §E.4 Sync-phase Audit-Ready Signal recorded in progress.md
- [ ] Stop hook (sync-phase quality gate) verified passing on sync commit
- [ ] NOTICE.md Anthropic 2026 attribution citation present (REQ-ATR-019)

### §D.4 Mx-phase Definition of Done

**Must-pass**:
- [ ] Step C judgement recorded in progress.md §E.5
- [ ] Expected EVALUATE-SKIP per mx-tag-protocol.md §a (markdown + shell only, 0 .go files, 0 goroutines, 0 fan_in delta)
- [ ] All 5 SPEC artifact frontmatter `status: implemented → completed` after Mx phase
- [ ] L60 atomic backfill executed (4-artifact frontmatter `sync_commit_sha:` consistency)

### §D.5 4-phase Close

**Must-pass**:
- [ ] `git log` shows 8 run-phase commits + sync commit + Mx commit + chore close commit on `main`
- [ ] `git push origin main` succeeds for all commits (Hybrid Trunk; Tier L MAY route through manager-git per REQ-ATR-020 → PR-based instead)
- [ ] Predecessor SPEC `SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001` frontmatter status verified `superseded`
- [ ] `MEMORY.md` index updated with project memory entry for this SPEC

---

Version: 0.1.0
Status: draft (plan-phase initial authoring)
Tier: L

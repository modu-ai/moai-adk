---
id: SPEC-V3R6-SEQ-THINKING-RETIRE-001
title: "Plan — Sequential-Thinking MCP Retirement"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.6.0"
module: ".claude/agents,.claude/skills,.claude/rules,internal/template/templates,.mcp.json,settings.json"
lifecycle: spec-anchored
tags: "v3r6, mcp, retirement, ultrathink, plan, milestones, ci-guard"
tier: M
---

# Plan — SPEC-V3R6-SEQ-THINKING-RETIRE-001

## §0. Plan Scope Boundary

### §0.1 Out of Scope

This plan covers only namespace cleanup of `sequential-thinking` literals from the MoAI-ADK project under the milestones M1–M7. Items intentionally **excluded** from this plan:

- User-local Sequential Thinking MCP installation guidance (users install via `~/.claude/settings.json` on their own).
- New Adaptive Thinking capability development (ultrathink keyword behavior is unchanged).
- Retirement of other MCP servers (`context7`, `chrome-devtools`, `claude-in-chrome`, `zai-mcp-server` are preserved).
- Documentation reorganization beyond the retirement note in CLAUDE.md §12 and the correction in CLAUDE.local.md §22.2.
- Migration tooling for projects that have already deployed sequential-thinking from older MoAI-ADK templates (run-phase commits land in v3.6 main; pre-v3.6 deployments remain on their existing templates).
- Run-phase commits and PRs themselves — those are manager-develop's responsibility executing against the milestones below.

## §1. Implementation Strategy

The retirement is mechanical and high-volume (48 source files + 3 config files + 1 new CI guard test = ~52 file changes). The risk surface is low because each change is a literal deletion (frontmatter tool token removal, server entry deletion, allow-list entry removal) or a small replacement (rule text edit, skill body redesign). The dominant risks are (a) drift between local and template files (mitigated by template-mirror pairing per milestone), (b) catalog.yaml hash drift (mitigated by M7 regeneration), and (c) manager-develop autonomous git pull (mitigated by B9 prohibition recited in delegation prompt).

Execution follows manager-develop Section A–E Tier M discipline using cycle_type=ddd (existing code with characterization tests for the CI guard; no new business logic):

- **ANALYZE** at M1: confirm inventory, lock baseline grep counts.
- **PRESERVE** distributed across M2–M6: each milestone's verification commands check that no unintended files changed.
- **IMPROVE** = the retirement edits themselves at M2–M6.

The milestone partition is by file class (agents → skills → foundation-thinking redesign → settings/MCP configs → docs/CI guard → validation). Each milestone produces a single commit pairing local + template changes.

## §2. Milestones

### M1 — Inventory and Baseline Capture

**Goal**: Lock authoritative pre-edit counts so M7 verification can prove the delta.

**Tasks**:
1. Run grep audit across the full affected surface and record file lists + occurrence counts in a `baseline.txt` artifact under `.moai/specs/SPEC-V3R6-SEQ-THINKING-RETIRE-001/baseline.txt`.
2. Capture the pre-SPEC test baseline: `go test ./internal/template/... 2>&1 | tee baseline-test.txt`. Distinguish pre-existing baseline failures from new regressions.
3. Capture the pre-SPEC `golangci-lint` baseline (if any): `golangci-lint run --timeout=2m 2>&1 | head -50 > baseline-lint.txt`.
4. Verify catalog.yaml current state: `wc -l .claude/skills/catalog.yaml internal/template/templates/.claude/skills/catalog.yaml`.
5. Document the run-phase prerequisite: B9 prohibition ("no autonomous `git pull/fetch/rebase`") is in the delegation prompt.

**Deliverable**: `baseline.txt`, `baseline-test.txt`, `baseline-lint.txt` artifacts.

**Verification**: AC-STR-001 baseline (records pre-edit count).

**Exit criterion**: Inventory matches spec.md §6 ±2 files.

---

### M2 — Agent Frontmatter Cleanup (14 + 14 = 28 files)

**Goal**: Remove `mcp__sequential-thinking__sequentialthinking` token from `tools:` field in all 14 affected agent files and their template mirrors.

**Tasks**:
1. For each agent in `.claude/agents/`: edit frontmatter `tools:` field to remove `, mcp__sequential-thinking__sequentialthinking` (preserving comma placement). Inspect each file to confirm only the literal removal — no other frontmatter or body changes.
2. Mirror each change to `internal/template/templates/.claude/agents/` (Template-First Rule per REQ-STR-008).
3. Per-file grep verification: each modified file must satisfy `grep -c 'sequential-thinking\|sequentialthinking' <file>` → 0.

**Affected files** (28):
- core (12): `manager-{develop,docs,project,quality,spec,strategy}.md` × 2 (local + template)
- expert (10): `expert-{backend,devops,frontend,refactoring,security}.md` × 2
- meta (6): `builder-harness.md`, `evaluator-active.md`, `plan-auditor.md` × 2

**Deliverable**: Single commit `chore(SPEC-V3R6-SEQ-THINKING-RETIRE-001): M2 — 14 agent frontmatter tools 정리`.

**Verification**: AC-STR-009.

**Exit criterion**: `grep -rln 'mcp__sequential-thinking' .claude/agents/ internal/template/templates/.claude/agents/` → exit 1.

---

### M3 — Skill Frontmatter + Body Cleanup (5 skills, excluding moai-foundation-thinking)

**Goal**: Remove `mcp__sequential-thinking__sequentialthinking` and sequential-thinking textual references from 5 skills (`moai-foundation-thinking` is handled in M4 as redesign, not cleanup).

**Tasks**:
1. For each skill (and its template mirror):
   - `moai-foundation-cc/reference/claude-code-settings-official.md`: remove sequential-thinking references in MCP server documentation prose (replace with generic example or omit).
   - `moai-foundation-cc/reference/skill-formatting-guide.md`: remove sequential-thinking token from any allowed-tools example. Replace example with `mcp__context7__resolve-library-id` if a placeholder is needed.
   - `moai-foundation-core/modules/agents-reference.md`: remove sequential-thinking from any agent catalog listing.
   - `moai/workflows/plan/clarity-interview.md`: remove sequential-thinking from any tooling reference; substitute with "ultrathink keyword" where deep-reasoning guidance was present.
   - `moai/workflows/run/context-loading.md`: same pattern as clarity-interview.md.
2. Mirror each change to `internal/template/templates/.claude/skills/`.
3. Per-file grep verification: each modified file must satisfy `grep -c 'sequential-thinking\|sequentialthinking' <file>` → 0.

**Affected files**: 5 skills × 2 (local + template) = 10 files.

**Deliverable**: Single commit `chore(SPEC-V3R6-SEQ-THINKING-RETIRE-001): M3 — 5 skill frontmatter+body 정리`.

**Verification**: AC-STR-010 (partial — 5 of 6 skills; final 6th in M4).

**Exit criterion**: 5 skill files PASS the per-file grep check.

---

### M4 — `moai-foundation-thinking` Redesign

**Goal**: Redesign the `moai-foundation-thinking` SKILL.md body to retire the "Sequential Thinking MCP (absorbed from moai-workflow-thinking)" subsection while preserving the creative frameworks (Critical Evaluation, Diverge-Converge, Deep Questioning) and the First Principles content (absorbed from moai-foundation-philosopher).

**Tasks**:
1. Edit `.claude/skills/moai-foundation-thinking/SKILL.md`:
   - **Frontmatter**: remove `mcp__sequential-thinking__sequentialthinking` from `allowed-tools:`. Remove `sequential-thinking` from `tags:` string. Remove `sequential thinking` from `triggers.keywords` array.
   - **Description**: rewrite to remove "Sequential Thinking MCP (absorbed from moai-workflow-thinking)" mention. New description emphasizes Creative frameworks + First Principles + Adaptive Thinking / ultrathink.
   - **Body — Adaptive Thinking section (~line 350+)**: rewrite the table to remove the `Sequential thinking` row. Update "Rules" sentence to remove the `Where sequential thinking` clause. Keep the `ultrathink → ALWAYS use Claude native extended reasoning` directive.
   - **BC marker**: add a brief retirement note under the Adaptive Thinking section: "Sequential Thinking MCP support retired in SPEC-V3R6-SEQ-THINKING-RETIRE-001. For deep reasoning use the `ultrathink` keyword (triggers Opus 4.7+ Adaptive Thinking)."
2. Mirror to `internal/template/templates/.claude/skills/moai-foundation-thinking/SKILL.md`.
3. Verify: `grep -c 'sequential-thinking\|sequentialthinking\|Sequential Thinking' .claude/skills/moai-foundation-thinking/SKILL.md` → 0 (after redesign). The BC marker uses the phrase "Sequential Thinking MCP support retired" exactly once; this is the BC marker, not a re-introduction — it is captured by the audit allow-list pattern.

**Affected files**: 2 files (local + template).

**Deliverable**: Single commit `chore(SPEC-V3R6-SEQ-THINKING-RETIRE-001): M4 — moai-foundation-thinking 재설계 + BC marker`.

**Verification**: AC-STR-003 (skill body literal check), AC-STR-005 (positive ultrathink/Adaptive Thinking mention check), AC-STR-010 (frontmatter clean for all 6 skills).

**Exit criterion**: All 6 skills PASS frontmatter + body grep checks, AND `moai-foundation-thinking` body retains First Principles + creative frameworks + BC marker.

**Special consideration for CI guard interaction**: M6's new test `TestSeqThinkingRetired` will need an allow-list pattern that permits the exact BC marker phrase ("Sequential Thinking MCP support retired in SPEC-V3R6-SEQ-THINKING-RETIRE-001") while rejecting all other sequential-thinking literals. The audit-allowlist file is `internal/template/seq_thinking_retire_audit_allowlist.txt` (added in M6) and lists the BC marker line verbatim.

---

### M5 — Settings, .mcp.json, and Rule Cleanup

**Goal**: Remove the sequential-thinking MCP server entry and permission allow-list entries from configuration files, and remove sequential-thinking references from the settings management rule.

**Tasks**:
1. Edit `.mcp.json` (local): remove the `"sequential-thinking": { ... }` block from `mcpServers`. Adjust trailing comma to keep JSON valid.
2. Edit `internal/template/templates/.mcp.json.tmpl`: remove the same block. Adjust trailing comma.
3. Edit `.claude/settings.json` (local): remove line 286 `"mcp__sequential-thinking__*",` from permission `allow` array. Verify JSON validity.
4. Edit `internal/template/templates/.claude/settings.json.tmpl`:
   - Remove line 392 `"mcp__sequential-thinking__*",` from permission `allow` array.
   - Remove line 562 `"sequential-thinking"` from `enabledMcpjsonServers` array. Adjust trailing comma on preceding entry.
5. Edit `.claude/settings.local.json` (runtime-managed, but currently contains sequential-thinking refs): remove all sequential-thinking permission entries. Note: this file is per-machine and §22 says it is runtime-managed; the cleanup here is one-time. Future runtime updates by `moai cg`/`moai glm` won't re-add sequential-thinking.
6. Edit `.claude/rules/moai/core/settings-management.md` (local): remove the line `- sequential-thinking: Complex problem analysis` from the MCP server list. Add a single-line retirement note: `- (retired) sequential-thinking — removed in SPEC-V3R6-SEQ-THINKING-RETIRE-001; use ultrathink keyword for deep reasoning`.
7. Mirror the rule edit to `internal/template/templates/.claude/rules/moai/core/settings-management.md`.

**Affected files**: 5 configs (local `.mcp.json` + `.claude/settings.json` + `.claude/settings.local.json` + template `.mcp.json.tmpl` + template `settings.json.tmpl`) + 2 rule files (local + template) = 7 files.

**Note on settings.local.json**: per CLAUDE.local.md §2 [HARD] settings.local.json Separation, this is runtime-managed and never templated. The M5 edit cleans up the user's machine but does not add a template line.

**Deliverable**: Single commit `chore(SPEC-V3R6-SEQ-THINKING-RETIRE-001): M5 — .mcp.json + settings + rules 정리`.

**Verification**: AC-STR-002, AC-STR-003 (config files), AC-STR-001 (rules included in scope grep).

**Exit criterion**: All 7 file edits PASS individual grep checks; JSON validates with `jq . .mcp.json` and `jq . .claude/settings.json` exit 0.

---

### M6 — CI Guard Test + Documentation Updates

**Goal**: Add the `TestSeqThinkingRetired` CI guard test and update root documentation (CLAUDE.md §12 retirement note + CLAUDE.local.md §22.2 correction).

**Tasks**:
1. Create `internal/template/seq_thinking_retire_audit_test.go` modeled after `internal/template/namespace_protection_audit_test.go` (precedent commit `d0782a365`):
   - Walks `.claude/`, `internal/template/templates/.claude/`, and reads `.mcp.json`, `.claude/settings.json`, `internal/template/templates/.mcp.json.tmpl`, `internal/template/templates/.claude/settings.json.tmpl`, `CLAUDE.md`.
   - Detects `sequential-thinking` and `sequentialthinking` literals.
   - Loads allow-list from `internal/template/seq_thinking_retire_audit_allowlist.txt` (one line per allow pattern; the BC marker line from M4 lives here).
   - Fails with sentinel `SEQ_THINKING_REINTRODUCED` if any non-allowed match is found, prints the file path + line number for each violation.
2. Create `internal/template/seq_thinking_retire_audit_allowlist.txt` with one entry: the M4 BC marker line.
3. Edit `CLAUDE.md` §12 MCP Servers & Deep Analysis Modes:
   - Remove any mention of Sequential Thinking MCP from the bulleted list (current content has UltraThink and Adaptive Thinking bullets but does **not** literally contain "sequential-thinking"; nonetheless add a retirement note).
   - Add a short paragraph: `Sequential Thinking MCP retired in SPEC-V3R6-SEQ-THINKING-RETIRE-001. The ultrathink keyword (Adaptive Thinking on Opus 4.7+) is the canonical deep-reasoning path. Users requiring sequential-thinking for personal use may install it independently via ~/.claude/settings.json (user-local scope).`
4. Edit `CLAUDE.local.md` §22.2 line 728:
   - Change `사용자 프로젝트는 mcp__context7 / mcp__sequential-thinking만 alwaysLoad되고 나머지는 ToolSearch preload 경로를 따른다.` to `사용자 프로젝트는 mcp__context7만 alwaysLoad되고 나머지는 ToolSearch preload 경로를 따른다. (sequential-thinking은 SPEC-V3R6-SEQ-THINKING-RETIRE-001에서 retired.)`
5. Run the new CI guard test: `go test -run TestSeqThinkingRetired ./internal/template/...` → expect PASS (sentinel `SEQ_THINKING_REINTRODUCED` defined; allow-list correctly permits BC marker).

**Affected files**: 4 (CLAUDE.md, CLAUDE.local.md, new test file, new allow-list file).

**Deliverable**: Single commit `chore(SPEC-V3R6-SEQ-THINKING-RETIRE-001): M6 — CI guard + 문서 정리`.

**Verification**: AC-STR-004, AC-STR-006, AC-STR-012.

**Exit criterion**: `TestSeqThinkingRetired` PASS. CLAUDE.md §12 contains the retirement statement. CLAUDE.local.md §22.2 no longer mentions `mcp__sequential-thinking`.

---

### M7 — Validation + Catalog Regeneration

**Goal**: Regenerate catalog.yaml hashes, run the full verification batch, and produce the final acceptance report.

**Tasks**:
1. Run `gen-catalog-hashes.go --all` (or equivalent) to regenerate hashes for 14 modified agents + 6 modified skills + 1 retired skill in `.claude/skills/catalog.yaml` and the template mirror.
2. Run the canonical 7-item verification batch in a single multi-Bash turn per `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution:
   - `go test ./...`
   - `go test -coverprofile=cover.out ./internal/template/...`
   - `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ internal/hook/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"` (C-HRA-008 sentinel)
   - `grep -rn 'FROZEN_SENTINEL\|HARNESS_FROZEN\|SEQ_THINKING_REINTRODUCED' internal/ | head -20`
   - `go run ./cmd/moai --version`
   - `go test -bench=. -benchmem -run=^$ ./internal/template/...` (optional)
   - `golangci-lint run --timeout=2m`
3. Run cross-platform builds (AC-STR-011):
   - `GOOS=darwin GOARCH=amd64 go build ./cmd/moai`
   - `GOOS=linux GOARCH=amd64 go build ./cmd/moai`
   - `GOOS=windows GOARCH=amd64 go build ./cmd/moai`
4. Generate `progress.md` summarizing M1–M7 evidence with per-AC PASS/FAIL.
5. Update spec.md `status: draft` → `implemented`, version `0.1.0` → `0.2.0`, `updated:` to the M7 commit date.

**Deliverable**: Single commit `chore(SPEC-V3R6-SEQ-THINKING-RETIRE-001): M7 — catalog rebuild + validation + status implemented`.

**Verification**: AC-STR-001 (global grep), AC-STR-007, AC-STR-008, AC-STR-011.

**Exit criterion**: All 12 ACs PASS. Catalog hashes parity verified. Cross-platform builds exit 0. Net new regressions over baseline = 0.

---

## §3. Sequencing and Dependencies

```
M1 (inventory)
   ↓
M2 (agents) ─┐
M3 (skills 5)│ — independent file classes, in principle parallelizable.
M4 (foundation-thinking redesign — must wait for M3 OR run after M3) — has BC marker that M6 allow-list references
M5 (configs + rules) ─┘
   ↓
M6 (CI guard + docs) — depends on M4's BC marker text being final
   ↓
M7 (validation + catalog)
```

**Recommended sequencing**: M1 → M2 → M3 → M4 → M5 → M6 → M7 (linear). Parallelization would save little wall-time given each milestone is a single bounded commit, and linear sequencing matches manager-develop's natural cadence.

## §4. Acceptance Mapping

| Acceptance Criterion | Verified By Milestone(s) |
|----------------------|--------------------------|
| AC-STR-001 (global namespace grep) | M2 + M3 + M4 + M5 + M6 (cumulative); finalized in M7 |
| AC-STR-002 (.mcp.json entries) | M5 |
| AC-STR-003 (settings + skill body) | M4 + M5 |
| AC-STR-004 (CLAUDE.md / CLAUDE.local.md) | M6 |
| AC-STR-005 (foundation-thinking ultrathink+) | M4 |
| AC-STR-006 (new CI guard test) | M6 |
| AC-STR-007 (make build + go test) | M7 |
| AC-STR-008 (catalog.yaml hashes) | M7 |
| AC-STR-009 (agent frontmatter tools) | M2 |
| AC-STR-010 (skill frontmatter allowed-tools) | M3 + M4 |
| AC-STR-011 (cross-platform builds) | M7 |
| AC-STR-012 (CLAUDE.md retirement statement) | M6 |

## §5. Risks

### R1 — Sibling SPEC overlap with V3R6-TEMPLATE-NEUTRALITY-AUDIT-001

The pending Template Neutrality Audit SPEC (memory `project_v3r6_template_neutrality_audit_2026_05_23`) may touch overlapping files for unrelated drift cleanup. If both SPECs run concurrently, merge conflicts are likely on agent frontmatter and skill frontmatter files.

**Mitigation**: Schedule this SPEC's run-phase after Template Neutrality Audit's plan-phase commits are merged, or vice versa. Coordination is verbal — no programmatic dependency. The two SPECs touch different aspects of the same files (this SPEC: sequential-thinking literals only; that SPEC: drift cleanup unrelated to MCP).

### R2 — Manager-develop autonomous git pull (Wave 3 lesson recurrence)

Wave 3 of SPEC-V3R6-CODE-COMMENTS-EN-001 documented manager-develop executing `git pull --rebase` autonomously, rewriting commit SHAs of previous waves. The fix landed as B9 prohibition in `.claude/rules/moai/development/manager-develop-prompt-template.md`.

**Mitigation**: This SPEC's run-phase delegation prompt (composed by orchestrator at run-time, not in this plan.md) MUST recite the B9 prohibition verbatim in Section D. Plan.md herein is a documentation artifact — the actual delegation prompt is composed by the orchestrator at /moai run time.

### R3 — catalog.yaml hash drift

`catalog.yaml` (both local and template) lists per-skill and per-agent file hashes. Edits to 14 agents + 6 skills require 20 hash updates. Forgetting M7's hash regeneration leaves the catalog desynchronized.

**Mitigation**: M7 explicitly runs `gen-catalog-hashes.go --all` as its first task. M7 verification compares pre-edit (M1 baseline) and post-edit `wc -l catalog.yaml` to confirm the catalog was touched.

### R4 — User session `enabledMcpjsonServers` cache

In-flight Claude Code sessions may have cached `enabledMcpjsonServers: [..., "sequential-thinking", ...]` from `.claude/settings.json` at session start. The session continues to attempt loading the sequential-thinking MCP until restart.

**Mitigation**: This is a no-op risk for users with running sessions — Claude Code's MCP loader will log a warning and fall back gracefully. Next session restart picks up the new settings.json. Document this in the post-merge changelog if a user-visible release note is appropriate.

### R5 — JSON validity on partial edits

Removing a server entry from `.mcp.json` or an allow-list entry from `settings.json` requires careful comma management. A trailing comma left behind produces invalid JSON.

**Mitigation**: M5 task list explicitly includes `jq . <file>` validation after each edit. M7 verification batch reruns the JSON validators.

### R6 — Cross-Sprint conflict with concurrent V3R6 SPECs

Sprint 2 candidates (AGENT-MODEL-ROUTING / PROMPT-CACHE / HOOK-OBSERVE-OPT-IN / HOOK-ASYNC-EXPAND) may touch agent frontmatter. If those Sprint 2 SPECs land first, agent file diff context shifts.

**Mitigation**: This SPEC's edits are line-level removals scoped to the `tools:` field; the surrounding frontmatter context is stable. Worst case is a 3-way merge resolution — no semantic conflict expected.

## §6. Technical Approach

**Cycle type**: `cycle_type=ddd` per `.moai/config/sections/quality.yaml` `constitution.development_mode: ddd`. This is a refactor (existing code, behavior preservation) — there is no new business logic to test-first. The CI guard test added in M6 is itself a specification test for the post-retirement state.

**ANALYZE** (M1): inventory and baseline. Existing 48-file footprint is well-bounded.

**PRESERVE** (M2–M5 verification): per-milestone grep checks confirm no unintended file changes. Test baseline from M1 prevents new regressions from being attributed to this SPEC.

**IMPROVE** (M2–M6 edits): the retirement edits themselves. Each milestone is a minimum-scope commit per Karpathy "Surgical Changes" principle.

**No new tests of business logic required** — the CI guard added in M6 is the only new test, and it is a specification test for the namespace clean state, not a unit test for new code.

## §7. Verification Batch (post-M7)

Single multi-Bash turn invoking all 7 read-only verifications per `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution:

```bash
# 1. Full test suite
go test ./...

# 2. Coverage report (template package — where CI guard lives)
go test -coverprofile=cover.out ./internal/template/...

# 3. Subagent-boundary grep (sentinel C-HRA-008)
grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ internal/hook/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"

# 4. Sentinel-key audit (NEW SEQ_THINKING_REINTRODUCED + existing)
grep -rn 'FROZEN_SENTINEL\|HARNESS_FROZEN\|SEQ_THINKING_REINTRODUCED\|NAMESPACE_LEAK' internal/ | head -20

# 5. CLI smoke
go run ./cmd/moai --version

# 6. Benchmark (optional)
go test -bench=. -benchmem -run=^$ ./internal/template/...

# 7. Lint baseline
golangci-lint run --timeout=2m
```

The 7-item batch satisfies AC-WO-007 expectations and produces the verification record for AC-STR-007 + AC-STR-011 (cross-platform builds are an 8th and 9th item run in parallel alongside the 7).

## §8. Configuration Touchpoints

- `.moai/config/sections/quality.yaml` `constitution.development_mode: ddd` — confirms cycle_type=ddd selection for run-phase.
- `.moai/config/sections/git-convention.yaml` `commit_message_format` — confirms `chore(SPEC-V3R6-SEQ-THINKING-RETIRE-001): M<N> — <description>` format.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Section A–E + B9 (autonomous git pull prohibition) must be recited in run-phase delegation prompt.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — canonical 12-field schema reflected in this plan.md frontmatter.

## §9. Estimated Effort and Priority Ordering

[HARD] No time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation. Priority ordering:

1. **High Priority**: M1 (baseline), M7 (validation) — bookends that lock the delta.
2. **Medium Priority**: M2 (agents) + M5 (configs) — highest blast radius if drift goes unnoticed.
3. **Standard Priority**: M3 (skills 5), M4 (foundation-thinking redesign), M6 (CI guard + docs) — bounded scope.

## §10. Post-Merge Follow-Up

- **Not required**: no follow-up SPECs depend on this retirement.
- **Optional**: a Lessons Protocol entry capturing the redesign pattern (BC marker for retired absorbed content) if the pattern is reused for future MCP retirements (e.g., chrome-devtools, claude-in-chrome).
- **Optional**: paste-ready resume message for next session if this SPEC's run-phase consumes more than 50% of a 1M context window.

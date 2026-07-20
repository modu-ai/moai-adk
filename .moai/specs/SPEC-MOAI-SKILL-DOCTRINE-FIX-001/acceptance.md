---
id: SPEC-MOAI-SKILL-DOCTRINE-FIX-001
title: "moai Skill Folder Doctrine Drift Remediation — Acceptance Criteria"
version: "0.1.0"
status: completed
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P1
phase: "v3.0.0 target"
module: ".claude/skills/moai"
lifecycle: spec-anchored
tags: "skill-doctrine, drift-remediation, gears, template-neutrality, harness, tier-l, agent-catalog"
---

# Acceptance Criteria: SPEC-MOAI-SKILL-DOCTRINE-FIX-001

## §A. Given-When-Then Scenarios

### Scenario 1 — Route A/B Tier branching restored (CRITICAL #1/#2)

- **Given** a Tier S or Tier M SPEC completes its final run-phase implementation step,
- **When** an implementing agent reads `workflows/run/task-decomposition.md` Phase 3 "Git Operations",
- **Then** the skill body instructs `manager-develop` to commit and push directly to `main` (no PR, no `manager-git` spawn) — and only branches to `manager-git` when Tier L or explicit `--pr` is declared.

### Scenario 2 — Harness-level plan-audit gating documented accurately (CRITICAL #3)

- **Given** a SPEC configured at the `minimal` harness level,
- **When** an implementing agent reads `workflows/plan/spec-assembly.md` for plan-audit guidance,
- **Then** the skill body states plan-audit runs as a lightweight 1-iteration, non-blocking audit (matching `harness.yaml` `minimal.plan_audit.enabled: true` + `require_must_pass: false` + `plan_audit_global.always_enabled: true`) — NOT that plan-audit is skipped.

### Scenario 3 — FROZEN path list matches ground truth (CRITICAL #5)

- **Given** a meta-harness invocation needs to know which paths are FROZEN,
- **When** an agent reads `workflows/project/meta-harness.md:225-227`,
- **Then** the stated FROZEN list exactly matches all **4** entries in `internal/harness/frozen_guard.go` `frozenPrefixes` (`.claude/agents/moai/`, `.claude/skills/moai-`, `.claude/skills/moai/`, `.claude/rules/moai/`) — `harness/` is removed from the agents line, the missing exact `.claude/skills/moai/` entry is added, and the two already-correct entries (`.claude/skills/moai-*/`, `.claude/rules/moai/`) are preserved (not deleted). `.claude/agents/harness/` is correctly documented as an ALLOWED (not frozen) prefix.

### Scenario 4 — Circular Detection Keywords Reference resolved (CRITICAL #6)

- **Given** an agent executing Phase 4.1a of `workflows/project/doc-generation.md` needs the 16-language manifest/ORM detection keyword table,
- **When** the agent follows the "Detection Keywords Reference" pointer,
- **Then** the pointer resolves to a file that actually contains the table (no circular "see the other file" pointer with neither file containing it).

### Scenario 5 — harness.md CLI-verb framing matches code reality (CRITICAL #7)

- **Given** an agent reads `workflows/harness.md`'s opening framing about CLI verb retirement,
- **When** the agent cross-checks against `internal/cli/harness_route.go:59-148` (`newHarnessRouterCmd()`) and `internal/cli/harness_retirement_test.go` (`TestHarnessV3R5VerbSurface`),
- **Then** the skill body correctly states that ALL harness verbs (including `status`/`apply`/`rollback`/`disable`) are Go-binary Cobra subcommands dispatched through the single unified `newHarnessRouterCmd()` tree — no blanket "CLI verb path retired" claim remains, AND no false learning-vs-v4 Go-binary-dispatch split is introduced in its place — AND the stale comment at `internal/cli/root.go:157-160` no longer describes the superseded `newHarnessCmd()` factory as if it were live, instead correctly naming `newHarnessRouterCmd()` (registered at `root.go:166`) as the live registration site.

### Scenario 6 — Template-mirror byte-identity preserved after fix (cross-cutting)

- **Given** any file in WG-A..WG-H that is currently byte-identical to its `internal/template/templates/` mirror,
- **When** the corresponding REQ's fix is applied via the Template-First procedure (`plan.md` §C),
- **Then** `diff -q .claude/skills/moai/<path> internal/template/templates/.claude/skills/moai/<path>` reports no difference after `make build` + sync.

### Scenario 7 — CI leak-test regex catches the newly-fixed leak classes retroactively (M4)

- **Given** the pre-fix content of `workflows/sync/quality-gates-context.md:153` (`SPEC-DB-SYNC-RELOC-001`) and `workflows/harness.md` (`REQ-HRN-FND-NNN` citations),
- **When** the extended `internal_content_leak_test.go` regex families (REQ-SKF-053) are run against a reconstructed pre-fix fixture of those two snippets,
- **Then** the extended regex flags both as leaks (proving the regex gap that let them ship is now closed) — while the post-fix (cleaned) content passes with zero leak findings.

## §B. Edge Cases

- A REQ's cited line number has drifted since the audit (files are actively edited) — the implementing agent MUST locate the finding by content/context match, not blindly trust the line number, and MUST re-verify the surrounding prose still matches the finding's description before editing.
- A file cited in the audit as template-identical has since diverged (e.g., a parallel session edited it) — re-run `diff -q` at run-phase start (`plan.md` §C step 1) before assuming the plan-authoring-time mirror status still holds.
- Fixing a T4 (content-leak) finding by relocating the internal citation to a code comment or a different template file — still a violation; the fix must remove the leak from the template-distributed surface entirely, not relocate it within template scope.
- Widening the CI leak-test regex (REQ-SKF-053) causes a new false-positive against an existing, legitimate (non-leak) template file — this is a regression; WG-I MUST run the full test suite before and after (per `plan.md` §B.3) and adjust the regex to avoid new false positives before declaring REQ-SKF-053 complete.

## §C. Machine-Verifiable Grep Assertions (per REQ, representative — not exhaustive)

| REQ | Assertion (run from repo root after fix) | Expected result |
|-----|-------------------------------------------|------------------|
| REQ-SKF-001 | `grep -n "Route A\|Route B\|Tier" .claude/skills/moai/workflows/run/task-decomposition.md` | ≥1 match near the Phase 3 Git Operations section |
| REQ-SKF-002 | `grep -n "Route A\|Route B\|Tier" .claude/skills/moai/workflows/sync/delivery.md` | ≥1 match near line 33's sync-commit-ownership statement |
| REQ-SKF-003 | `grep -n "plan_audit_global\|1-iteration\|lightweight" .claude/skills/moai/workflows/plan/spec-assembly.md` | ≥1 match; `grep -c "skip.*plan.audit\|plan.audit.*skip"` at the finding's location == 0 |
| REQ-SKF-004 | `grep -rc "moai-design-craft" .claude/skills/moai/workflows/plan/clarity-interview.md .claude/skills/moai/workflows/review.md` | 0 in both files |
| REQ-SKF-005 | `grep -A5 "FROZEN" .claude/skills/moai/workflows/project/meta-harness.md \| grep -c "harness/"` | 0 (harness/ no longer bundled into the agents line) |
| REQ-SKF-005 (positive) | `for p in ".claude/agents/moai/" ".claude/skills/moai-" ".claude/skills/moai/" ".claude/rules/moai/"; do grep -cF "$p" .claude/skills/moai/workflows/project/meta-harness.md; done` | Each of the 4 prints ≥1 (all 4 correct prefixes present — the 2 pre-existing correct ones are NOT deleted, the missing exact `.claude/skills/moai/` is added) |
| REQ-SKF-006 | `grep -c "16.*language\|manifest.*ORM\|ORM.*manifest" .claude/skills/moai/workflows/project.md .claude/skills/moai/workflows/project/doc-generation.md` | ≥1 in exactly one of the two files (the table now lives somewhere) |
| REQ-SKF-007 | `grep -c "CLI verb path retired" .claude/skills/moai/workflows/harness.md .claude/skills/moai/SKILL.md` | 0 in both (blanket claim removed) |
| REQ-SKF-007 (positive, no learning-vs-v4 split) | `grep -Ec "workflow-body-only|no Go binary" .claude/skills/moai/workflows/harness.md .claude/skills/moai/SKILL.md` | 0 in both (the false split framing is NOT introduced as a replacement claim) |
| REQ-SKF-007 (positive, accurate framing) | `grep -Ec "Go.binary|un-retired|newHarnessRouterCmd" .claude/skills/moai/workflows/harness.md` | ≥1 (corrected accurate framing present) |
| REQ-SKF-007 sub-clause (b) | Manual read of `internal/cli/root.go:157-166` | Comment no longer asserts "no Go binary invocation" for the lifecycle verbs; correctly names `newHarnessRouterCmd()` (registered at line 166) as the live registration site, distinguished from the superseded `newHarnessCmd()` factory |
| REQ-SKF-008 | `grep -c "EARS format" .claude/skills/moai/SKILL.md .claude/skills/moai/workflows/moai.md` | 0 in both |
| REQ-SKF-009 | `grep -c "plan-auditor" .claude/skills/moai/SKILL.md` | ≥3 (existing sync-auditor-count parity, plus new plan-phase + default-flow entries) |
| REQ-SKF-011 | `grep -c "backend-dev\|frontend-dev" .claude/skills/moai/workflows/moai.md` | 0 |
| REQ-SKF-013 | `grep -A2 "architect" .claude/skills/moai/team/run.md \| grep -c "sonnet"` (in the Role Profile table row for architect) | 0 (corrected to opus, or table removed in favor of cross-reference) |
| REQ-SKF-014 | `grep -c "teammateMode" .claude/skills/moai/team/glm.md` | ≥1 |
| REQ-SKF-016 | `grep -c "@MX:DEBT" .claude/skills/moai/workflows/mx.md` | ≥1 |
| REQ-SKF-017 | count of distinct language tokens in `workflows/clean.md`'s language table | 16 |
| REQ-SKF-018 | `grep -c "(Recommended)\|(권장)" .claude/skills/moai/workflows/feedback.md` | ≥2 (both cited option sets) |
| REQ-SKF-022 | `grep -c "backend-dev\|frontend-dev" .claude/skills/moai/workflows/run/mode-orchestration.md` | 0 |
| REQ-SKF-024 | `grep -c "SPEC Lifecycle Level 1\|SPEC Lifecycle Level 2\|SPEC Lifecycle Level 3" .claude/skills/moai/workflows/sync/doc-execution.md` | 0; `grep -c "spec-anchored\|spec-lite\|exploratory"` ≥1 |
| REQ-SKF-025 | `grep -n "minimal.*evaluator\|harness.*level" .claude/skills/moai/workflows/sync/quality-gates-quality.md` | ≥1 match near the sync-auditor invocation |
| REQ-SKF-027 | `grep -c "SPEC-DB-SYNC-RELOC-001" .claude/skills/moai/workflows/sync/quality-gates-context.md internal/template/templates/.claude/skills/moai/workflows/sync/quality-gates-context.md` | 0 in both |
| REQ-SKF-031 | `grep -c "auto_sync: true" .claude/skills/moai/workflows/project/doc-generation.md` (flat-boolean form) | 0; `grep -c "auto_sync.*enabled"` ≥5 |
| REQ-SKF-032 | `grep -rEc '\bSPEC-[A-Z][A-Z0-9-]*-[0-9]{3}\b|\b(REQ\|AC)-[A-Z0-9-]+\b|\bC-PH-[0-9]+\b' .claude/skills/moai/workflows/project/doc-generation.md .claude/skills/moai/workflows/project/meta-harness.md .claude/skills/moai/workflows/project.md` | 0 across all three |
| REQ-SKF-033 | `grep -n "doctor" .claude/skills/moai/SKILL.md \| grep -c "Branch A.1\|list.*edit.*remove.*doctor"` (verbs line + title near Branch A.1) | ≥2 |
| REQ-SKF-035 | `grep -c "REQ-HRN-FND" .claude/skills/moai/workflows/harness.md internal/template/templates/.claude/skills/moai/workflows/harness.md` | 0 in both |
| REQ-SKF-036 | `grep -A3 "^phases:" .claude/skills/moai/references/mx-tag.md \| grep -c "sync"` | ≥1 |
| REQ-SKF-043 | `grep -Ec '[0-9]+-[0-9]+x speedup|[0-9]+-[0-9]+s vs [0-9]+-[0-9]+s' .claude/skills/moai/workflows/moai.md` | 0 |
| REQ-SKF-044 | `grep -c -- "--sequential" .claude/skills/moai/workflows/moai.md` | ≥1 |
| REQ-SKF-047a (filename/content swap) | `grep -c "^# " .claude/skills/moai/workflows/run/task-decomposition.md \| head -1` (manual read: confirm the H1 title and body content actually match "Task Decomposition", and `phase-execution.md`'s H1/body match "Phase Execution" — no cross-swap) | Titles/content match filenames on both files |
| REQ-SKF-047b (Sprint→Epic/Milestone) | `grep -rc "Sprint " .claude/skills/moai/workflows/run/context-loading.md .claude/skills/moai/workflows/run/phase-execution.md .claude/skills/moai/workflows/run/task-decomposition.md` | 0 across all three (all replaced with Epic/Milestone) |
| REQ-SKF-047c (teammate-count table) | `grep -c "3-4 teammates\|3-4개" .claude/skills/moai/workflows/run/mode-orchestration.md` | 0; `grep -c "3-5 teammates\|3-5개"` ≥1 |
| REQ-SKF-050a (team/debug.md normalization) | Manual read of `.claude/skills/moai/team/debug.md` | Spawn instructions use ```` ```Agent(...)``` ```` code-block conventions consistent with sibling team/*.md files (no informal prose-only spawn instructions) |
| REQ-SKF-050b (SKILL.md team-mode pointer) | `grep -A3 -i "^## .*fix" .claude/skills/moai/SKILL.md \| grep -c "team"` | ≥1 (fix entry now points to team mode) |
| REQ-SKF-050c (fix.md naming) | `grep -c "team-debug.md" .claude/skills/moai/workflows/fix.md` | 0 (corrected to `team/debug.md`) |
| REQ-SKF-052a (sync-auditor description unified) | `grep -c "sync-auditor" .claude/skills/moai/workflows/sync/doc-execution.md .claude/skills/moai/workflows/sync/quality-gates-quality.md` (manual read: confirm both descriptions are textually identical or one cross-references the other) | Single unified description, no 3rd divergent variant |
| REQ-SKF-052b (delivery.md:239 Route A language) | Manual read of `.claude/skills/moai/workflows/sync/delivery.md:239` | Language no longer undercuts Route A as the Tier S/M default |
| REQ-SKF-052c (Stop-hook/agent conflation) | Manual read of `.claude/skills/moai/workflows/sync/quality-gates-quality.md:105-118` | Stop-hook-triggered checks and agent-invoked checks are clearly distinguished, no conflated wording |
| REQ-SKF-052d (phase-number/skip_phases namespace) | Manual read cross-checking phase numbers cited against `harness.yaml` `skip_phases` keys | Namespace mapping is unambiguous (e.g. explicit "Phase N corresponds to skip_phases key X" statement) |
| REQ-SKF-053 | `go test ./internal/template/... -run TestInternalContentLeak` | PASS, with new sub-tests covering the 4 new leak-class shapes (2-segment REQ/AC, C-PH-NNN, non-V3R SPEC prefix, 4-segment REQ-HRN-FND) |
| Scenario 6 (all WGs) | `diff -q .claude/skills/moai/<path> internal/template/templates/.claude/skills/moai/<path>` for every WG-A..WG-H file | No output (identical) for all files except any explicitly-documented intentional residual drift |

## §D. AC Matrix (severity + traceability)

| AC ID | REQ | Severity | Verification method |
|-------|-----|----------|----------------------|
| AC-SKF-001 | REQ-SKF-001 | P0 | Grep (§C) + manual read of Phase 3 branching logic |
| AC-SKF-002 | REQ-SKF-002 | P0 | Grep (§C) + manual read of delivery.md line 33 context |
| AC-SKF-003 | REQ-SKF-003 | P0 | Grep (§C) + cross-check against `harness.yaml` `minimal.plan_audit` |
| AC-SKF-004a | REQ-SKF-004 (clarity-interview.md) | P0 | Grep (§C) |
| AC-SKF-004b | REQ-SKF-004 (review.md, 3 occurrences) | P0 | Grep (§C) |
| AC-SKF-005a | REQ-SKF-005 (harness/ removed) | P0 | Grep (§C) |
| AC-SKF-005b | REQ-SKF-005 (all 4 correct prefixes present, positive) | P0 | Grep (§C) + cross-check against `frozen_guard.go` `frozenPrefixes` (4 entries, verified) |
| AC-SKF-006 | REQ-SKF-006 | P0 | Grep (§C) + manual read confirming the table exists in one file |
| AC-SKF-007a | REQ-SKF-007 (blanket claim removed) | P0 | Grep (§C) |
| AC-SKF-007b | REQ-SKF-007 (no false learning-vs-v4 split introduced) | P0 | Grep (§C, negative check) |
| AC-SKF-007c | REQ-SKF-007 (accurate Go-binary framing present) | P0 | Grep (§C, positive check) + cross-check against `harness_route.go`/`harness_retirement_test.go` |
| AC-SKF-007d | REQ-SKF-007 sub-clause (b) (root.go:157-160 stale comment corrected) | P0 | Manual read (§C) |
| AC-SKF-008 | REQ-SKF-008 | P1 | Grep (§C) |
| AC-SKF-009 | REQ-SKF-009 | P1 | Grep (§C) + manual read of both Quick Reference lists (plan, default) |
| AC-SKF-010 | REQ-SKF-010 | P1 | Manual read confirming dead sections removed, cross-ref to SPEC-SUBCOMMAND-RETIRE-001 present |
| AC-SKF-011 | REQ-SKF-011 | P1 | Grep (§C) |
| AC-SKF-012 | REQ-SKF-012 | P1 | Manual read confirming `--worktree` table/parsing reconciled |
| AC-SKF-013 | REQ-SKF-013 | P1 | Grep (§C) + cross-check against `workflow.yaml` `role_profiles` |
| AC-SKF-014 | REQ-SKF-014 | P1 | Grep (§C) + cross-check `internal/cli/glm.go` mechanism |
| AC-SKF-015 | REQ-SKF-015 | P1 | Manual read confirming reconciled `team_mode` explanation |
| AC-SKF-016 | REQ-SKF-016 | P1 | Grep (§C) |
| AC-SKF-017 | REQ-SKF-017 | P1 | Count distinct language tokens == 16 (or explicit cross-ref to fix.md Scanner 3) |
| AC-SKF-018 | REQ-SKF-018 | P1 | Grep (§C) |
| AC-SKF-019 | REQ-SKF-019 | P1 | Grep `"Mode Selection"` present in Phase 0.95 section of phase-execution.md |
| AC-SKF-020 | REQ-SKF-020 | P1 | Manual link resolution check (relative path resolves, anchor matches) |
| AC-SKF-021 | REQ-SKF-021 | P1 | Manual read confirming parallel-batch / file-redirect reference present |
| AC-SKF-022 | REQ-SKF-022 | P1 | Grep (§C) |
| AC-SKF-023 | REQ-SKF-023 | P1 | Grep `"archived-agent-rejection"` in doc-execution.md/quality-gates-quality.md — confirm §C row 2 citation removed |
| AC-SKF-024 | REQ-SKF-024 | P1 | Grep (§C) |
| AC-SKF-025 | REQ-SKF-025 | P1 | Grep (§C) + cross-check `harness.yaml` `minimal.evaluator` |
| AC-SKF-026 | REQ-SKF-026 | P1 | Grep `"Functionality 40\|Security 25\|Craft 20\|Consistency 15"` present |
| AC-SKF-027 | REQ-SKF-027 | P1 | Grep (§C), both local + template mirror |
| AC-SKF-028 | REQ-SKF-028 | P1 | Manual read confirming `moai-meta-harness` no longer invoked, or Phase 5/6 rewritten/removed |
| AC-SKF-029 | REQ-SKF-029 | P1 | Grep `"Phase 7"` present in project.md routing table |
| AC-SKF-030 | REQ-SKF-030 | P1 | Manual trace confirming Phase 5/6/7 reachable from Phase 4.2 |
| AC-SKF-031 | REQ-SKF-031 | P1 | Grep (§C), all 5 occurrences corrected |
| AC-SKF-032 | REQ-SKF-032 | P1 | Grep (§C) across 3 files |
| AC-SKF-033 | REQ-SKF-033 | P1 | Grep (§C) |
| AC-SKF-034 | REQ-SKF-034 | P1 | Manual read confirming `--execute` trust-boundary documented |
| AC-SKF-035 | REQ-SKF-035 | P1 | Grep (§C), both local + template mirror |
| AC-SKF-036 | REQ-SKF-036 | P2 | Grep (§C) |
| AC-SKF-037 | REQ-SKF-037 | P2 | Manual read confirming frontmatter agents match + new sections present |
| AC-SKF-038 | REQ-SKF-038 | P2 | Grep `"sync/quality-gates-quality.md"` in fix.md line ~134 |
| AC-SKF-039 | REQ-SKF-039 | P2 | Manual read confirming single unambiguous `--security` routing statement |
| AC-SKF-040 | REQ-SKF-040 | P2 | Grep `"Type your own answer"` in clarity-interview.md == 0 |
| AC-SKF-041 | REQ-SKF-041 | P2 | Grep `"ToolSearch"` count across plan/ files ≥ AskUserQuestion call-site count |
| AC-SKF-042 | REQ-SKF-042 | P2 | Manual read confirming DP2/DP3 labeling consistent |
| AC-SKF-043 | REQ-SKF-043 | P2 | Grep (§C) |
| AC-SKF-044 | REQ-SKF-044 | P2 | Grep (§C) + frontmatter date check |
| AC-SKF-045 | REQ-SKF-045 | P2 | Manual read confirming Phase 6→3 reference corrected |
| AC-SKF-046 | REQ-SKF-046 | P2 | Grep `"Opus 4.6"` in task-decomposition.md == 0 |
| AC-SKF-047a | REQ-SKF-047 (filename/content swap check) | P2 | Manual read (§C) |
| AC-SKF-047b | REQ-SKF-047 (Sprint→Epic/Milestone) | P2 | Grep (§C) |
| AC-SKF-047c | REQ-SKF-047 (teammate-count table 3-4→3-5) | P2 | Grep (§C) |
| AC-SKF-048 | REQ-SKF-048 | P2 | Grep `"Wave C"` in harness.md == 0 |
| AC-SKF-049 | REQ-SKF-049 | P2 | Count distinct language tokens in detection glob == 16 |
| AC-SKF-050a | REQ-SKF-050 (team/debug.md normalization, WG-E) | P2 | Manual read (§C) |
| AC-SKF-050b | REQ-SKF-050 (SKILL.md team-mode pointer, WG-A) | P2 | Grep (§C) |
| AC-SKF-050c | REQ-SKF-050 (fix.md naming fix, WG-B) | P2 | Grep (§C) |
| AC-SKF-051 | REQ-SKF-051 | P2 | Grep `"architect"` + `"ANTHROPIC_DEFAULT_FABLE_MODEL"` in team/glm.md ≥1 each |
| AC-SKF-052a | REQ-SKF-052 (sync-auditor description unified) | P2 | Manual read (§C) |
| AC-SKF-052b | REQ-SKF-052 (delivery.md:239 Route A language) | P2 | Manual read (§C) |
| AC-SKF-052c | REQ-SKF-052 (Stop-hook/agent conflation) | P2 | Manual read (§C) |
| AC-SKF-052d | REQ-SKF-052 (phase-number/skip_phases namespace) | P2 | Manual read (§C) |
| AC-SKF-053 | REQ-SKF-053 | P1 (CI) | `go test ./internal/template/... -run TestInternalContentLeak` PASS, Scenario 7 fixture check |
| AC-SKF-054 | Scenario 6 (all WGs) | Cross-cutting | `diff -q` byte-identity re-verification, all WG-A..WG-H files |

## §E. Definition of Done

- [ ] All 7 P0 (CRITICAL) REQs verified via their grep assertion + manual read.
- [ ] All 28 P1 (MAJOR) REQs verified.
- [ ] All 17 P2 (MINOR) REQs verified.
- [ ] REQ-SKF-053 (CI regex extension) lands LAST (M4), and its test suite passes both before (no false positives against untouched files) and after (retroactively catches the fixed leak classes) M1-M3 land.
- [ ] Every template-mirrored file re-verified byte-identical (`diff -q`) after Template-First sync, OR any residual intentional drift explicitly documented in the run-phase report.
- [ ] No new internal SPEC-ID/REQ/AC/C-PH/commit-SHA/internal-date leaks introduced into any template-mirrored file during remediation (self-check via the same grep patterns used in REQ-SKF-053's new leak classes).
- [ ] `go vet ./...` and `golangci-lint run` clean (only WG-I touches Go/test code; all other write-groups are markdown-only).
- [ ] `go test ./internal/template/...` passes in full (not just the new sub-tests) — no regression to existing leak-class coverage.
- [ ] Full read-only re-audit spot-check: re-run the original 7 CRITICAL finding greps from the audit and confirm all 7 no longer reproduce.

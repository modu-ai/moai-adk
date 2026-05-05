---
id: SPEC-CRON-PATTERN-001
acceptance_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-CRON-PATTERN-001

## Given-When-Then Scenarios

### Scenario 1: 카탈로그 문서 존재

**Given** the SPEC-CRON-PATTERN-001 implementation completes
**And** Template-First sync runs

**When** the user inspects the project's rules directory

**Then** the file `.claude/rules/moai/workflow/routine-patterns.md` SHALL exist
**And** the corresponding template file SHALL exist
**And** both files SHALL be identical (same hash)

---

### Scenario 2: 5 Standard Patterns 명시

**Given** the catalog document exists

**When** the user reads §2 Standard Patterns

**Then** the section SHALL contain exactly 5 patterns: P1 (Backlog Triage), P2 (Documentation Drift), P3 (Dependency Notifier), P4 (CI Failure Aggregator), P5 (Memory Hygiene Sweep)
**And** each pattern SHALL document schedule, prompt, output, and rationale (4 fields)
**And** each pattern's schedule SHALL be valid cron syntax

---

### Scenario 3: Plan-Tier Limits Table

**Given** the catalog document exists

**When** the user reads §4 Plan-Tier Limits

**Then** the section SHALL contain a table with 4 plans: Pro (5/day), Max (15/day), Team (25/day), Enterprise (unlimited)
**And** each plan SHALL include recommended pattern subset

---

### Scenario 4: Failure Policy

**Given** the catalog document exists
**And** a routine fails 3 consecutive runs

**When** the orchestrator processes the next session

**Then** the routine SHALL be paused
**And** the orchestrator SHALL emit a natural-language status announcement (NOT AskUserQuestion)
**And** the routine SHALL remain paused until user re-registers via `CronCreate`

---

### Scenario 5: History JSONL Schema

**Given** a routine completes execution

**When** the routine appends to `.moai/routines/history/<routine-id>-YYYY-MM-DD.jsonl`

**Then** the entry SHALL include: timestamp (ISO-8601), routine_id, exit_status, summary, output_size_bytes, duration_ms
**And** the entry SHALL be valid JSONL (single line, parseable JSON)

---

### Scenario 6: P1 Backlog Triage 등록 + 실행

**Given** a Pro plan user
**And** the user runs `CronCreate` for P1 with schedule `0 2 * * *`

**When** the routine executes the next morning at 02:00

**Then** the routine SHALL list top 5 SPECs by priority from `.moai/specs/`
**And** the routine SHALL append a JSONL entry to `.moai/routines/history/triage-YYYY-MM-DD.jsonl`
**And** the entry SHALL include exit_status = "success" and a 1-line summary

---

### Scenario 7: Pattern P3 (Dependency Notifier) GitHub Issue 생성

**Given** P3 routine is registered with daily 06:00 schedule
**And** new dependency versions are available

**When** the routine executes

**Then** the routine SHALL create a GitHub issue with summary
**And** the routine SHALL append history JSONL entry
**And** the issue SHALL be labeled `area:deps` per project conventions

---

### Scenario 8: Anti-Pattern — Hardcoded Credentials 차단

**Given** a user attempts to register a routine with hardcoded API key in prompt

**When** the catalog's pattern authoring guide is consulted

**Then** the guide SHALL warn against hardcoded credentials
**And** the guide SHALL recommend connector-based authentication (Linear, Slack)
**And** the example SHALL show secret references via secure connector config

---

## Edge Cases

### EC-1: `.moai/routines/history/` 디렉토리 부재
First routine execution SHALL auto-create the directory before writing JSONL entry.

### EC-2: Plan-Tier Limit 도달
If user is on Pro plan with 5 routines already running, attempting 6th SHALL fail at `CronCreate` level. Catalog warns user about the limit.

### EC-3: Cron Syntax Invalid
If user-supplied schedule fails cron validation, `CronCreate` SHALL reject; catalog references standard cron syntax sources.

### EC-4: Connector Misconfiguration
If P3 (Linear/Slack connector) connector misconfig occurs, routine SHALL emit error indicator. After 3 consecutive failures → pause per Scenario 4.

### EC-5: Pattern P5 Memory Hygiene Cross-Reference
P5 routine references `~/.claude/projects/<hash>/memory/MEMORY.md`. If file is missing, routine SHALL skip with summary "MEMORY.md not found, no action".

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| Catalog document | both local + template | file existence |
| 5 patterns | exactly 5 with 4 fields each | grep + schema check |
| Plan-tier table | 4 plans | grep |
| Failure policy | 3 consecutive → pause | grep |
| History schema | 6 fields | grep |
| Pattern authoring guide | 200-500 tokens recommendation | grep |
| Cross-references | >= 2 (SKILL.md + CLAUDE.md) | grep count |
| Template-First sync | clean | `make build` diff |
| Anti-patterns | >= 3 | grep count |
| plan-auditor | PASS | auditor report |

---

## Definition of Done

- [ ] All 8 Given-When-Then scenarios PASS
- [ ] All 5 edge cases (EC-1 to EC-5) documented and handled
- [ ] All 10 quality gate criteria meet threshold
- [ ] Catalog document at `.claude/rules/moai/workflow/routine-patterns.md` and template
- [ ] 5 patterns (P1-P5) all 4-field documented
- [ ] Plan-tier limits table present
- [ ] Failure policy (3 consecutive → pause) documented
- [ ] History JSONL schema documented
- [ ] `.moai/routines/history/.gitkeep` present
- [ ] `.gitignore` recommended entry documented
- [ ] Pattern authoring guide (200-500 tokens) documented
- [ ] Anti-patterns (3+) documented
- [ ] `moai-workflow-project` SKILL.md cross-ref added
- [ ] CLAUDE.md cross-ref added
- [ ] `make build` regenerates embedded.go cleanly
- [ ] CHANGELOG.md updated under Unreleased
- [ ] No Go code change (documentation-only SPEC verified by `git diff`)
- [ ] plan-auditor PASS
- [ ] dogfooding: at least one pattern (e.g., P2) registered after merge

End of acceptance.md (SPEC-CRON-PATTERN-001).

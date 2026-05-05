---
id: SPEC-TOOL-AUDIT-001
acceptance_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-TOOL-AUDIT-001

## Given-When-Then Scenarios

### Scenario 1: Audit 보고서 생성

**Given** `.moai/observability/tool-invocations.jsonl` contains 4523 entries across 100 sessions
**And** `tool_audit.enabled = true` in observability.yaml

**When** the user invokes `moai audit tools`

**Then** the command SHALL produce a markdown report at `.moai/research/tool-usage/audit-<DATE>.md`
**And** the report SHALL contain a "Summary" section with sessions analyzed (100), total invocations (4523), and agents covered count
**And** the command SHALL exit with code 0

---

### Scenario 2: 5% 미만 도구 → removal candidate flag

**Given** the JSONL data shows manager-ddd invoked Edit (245x), Read (198x), Write (88x), Grep (65x), TodoWrite (16x)
**And** total manager-ddd invocations = 612

**When** the audit command runs

**Then** the report SHALL list TodoWrite (16/612 = 2.6%) under "Removal Candidates" for manager-ddd
**And** the entry SHALL include "review required" annotation
**And** Edit, Read, Write, Grep SHALL NOT be flagged (above 5%)

---

### Scenario 3: 20+ 도구 agent → specialization 권장

**Given** an agent X declares 24 tools in its frontmatter `tools:` field

**When** the audit command runs

**Then** the report SHALL list agent X under "High-Tool-Count Agents (>= 20 tools)"
**And** the entry SHALL include recommendation: "split into focused sub-agents"
**And** the entry SHALL include the actual tool count (24)

---

### Scenario 4: Auto-removal 차단

**Given** the user attempts `moai audit tools --auto-remove`

**When** the command parses the flag

**Then** the command SHALL reject the flag immediately with error:
  "ERROR: Auto-removal is prohibited per SPEC-TOOL-AUDIT-001 REQ-TA-014. Tool removal requires human approval."
**And** the command SHALL exit with non-zero code
**And** no agent frontmatter SHALL be modified

---

### Scenario 5: 데이터 부족 경고

**Given** the JSONL data contains only 25 sessions
**And** minimum threshold is 30 sessions

**When** the user invokes `moai audit tools`

**Then** the command SHALL emit a warning: "data insufficient: 25 sessions < 30 minimum"
**And** the command SHALL still produce a partial report
**And** the report SHALL note the data limitation in the Summary section

---

### Scenario 6: Privacy — log scan 검증

**Given** the JSONL data contains 100 entries

**When** the user runs `grep -E "(secret|token|password|api[_-]?key|/Users/|/home/)" .moai/observability/tool-invocations.jsonl`

**Then** the grep SHALL return zero matches
**And** the JSONL fields SHALL be limited to: ts, session_id, agent, tool, duration_ms, success
**And** no tool argument or response content SHALL appear in the log

---

### Scenario 7: --window 사용자 정의

**Given** the JSONL contains 200 sessions

**When** the user invokes `moai audit tools --window 50`

**Then** the audit SHALL aggregate only the most recent 50 sessions
**And** older sessions SHALL be excluded
**And** the report Summary SHALL state "Sessions analyzed: 50"

---

### Scenario 8: PostToolUse hook 기록 동작

**Given** `tool_audit.enabled = true` in observability.yaml
**And** an agent invokes the Edit tool

**When** the PostToolUse hook fires

**Then** the hook SHALL append a JSON line to `.moai/observability/tool-invocations.jsonl`
**And** the line SHALL contain ts, session_id, agent="<calling agent>", tool="Edit", duration_ms, success=true
**And** the append SHALL be atomic (no partial writes on concurrent calls)
**And** the hook SHALL NOT block on append failure (silent failure)

---

## Edge Cases

### EC-1: JSONL malformed line
If `tool-invocations.jsonl` contains a malformed line, the aggregator SHALL skip the line and log a warning to stderr. The malformed line SHALL NOT abort the audit.

### EC-2: tool_audit.enabled=false
If `tool_audit.enabled = false` (default), the PostToolUse hook SHALL NOT record invocations. `moai audit tools` SHALL emit "no data — enable tool_audit in observability.yaml".

### EC-3: agent has zero invocations
If an agent declared in catalog has zero invocations in the audit window, the report SHALL list it under "No Activity" with note "no invocations during window".

### EC-4: Frontmatter parsing failure
If an agent's frontmatter cannot be parsed, the audit SHALL skip that agent's tool count check and log a warning. Other agents SHALL be processed normally.

### EC-5: PR comment auto-publish attempt
If the user attempts auto-publish via flag like `--post-comment`, the command SHALL reject (REQ-TA-017) — manual review only.

### EC-6: Concurrent JSONL writes
If multiple sub-agents fire PostToolUse hooks simultaneously, the JSONL append SHALL use O_APPEND atomic write (single write() syscall per line) per POSIX guarantee.

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| Audit command functional | exit code 0 with valid data | E2E test |
| 5% threshold detection | seeded low-usage tool flagged | unit test |
| 20+ tool detection | seeded 24-tool agent flagged | unit test |
| Auto-remove block | --auto-remove rejected | unit test |
| Privacy: no user content | grep scan | 0 matches |
| Min sessions warning | 25 sessions sample | warning emitted |
| Cross-platform | macOS/Linux/Windows CI | 3/3 PASS |
| Hook overhead | benchmark (with vs without audit logging) | < 5ms per invocation |
| Template-First sync | clean | `make build` diff |
| plan-auditor | PASS | auditor report |

---

## Definition of Done

- [ ] All 8 Given-When-Then scenarios PASS
- [ ] All 6 edge cases (EC-1 to EC-6) documented and handled
- [ ] All 10 quality gate criteria meet threshold
- [ ] `internal/audit/tools/aggregator.go` and `report.go` with >= 90% coverage
- [ ] `cmd/moai/audit.go` integration tests for tools subcommand
- [ ] PostToolUse hook records invocations when enabled
- [ ] `.moai/research/tool-usage/` directory created
- [ ] `.moai/config/sections/observability.yaml` updated with `tool_audit.enabled`
- [ ] `.claude/rules/moai/quality/tool-audit.md` policy document
- [ ] CLAUDE.md §6 Quality Gates cross-reference
- [ ] `make build` regenerates embedded.go cleanly
- [ ] CHANGELOG.md updated under Unreleased
- [ ] CI runners (ubuntu/macos/windows) PASS
- [ ] Privacy review: log scan zero violations
- [ ] plan-auditor PASS
- [ ] dogfooding: first audit run on this project documented

End of acceptance.md (SPEC-TOOL-AUDIT-001).

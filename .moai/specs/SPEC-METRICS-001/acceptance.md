---
id: SPEC-METRICS-001
acceptance_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-METRICS-001

## Given-When-Then Scenarios

### Scenario 1: 정책 문서 존재

**Given** the SPEC-METRICS-001 implementation completes
**And** Template-First sync runs

**When** the user inspects the project's rules directory

**Then** the file `.claude/rules/moai/workflow/contribution-metrics.md` SHALL exist
**And** the corresponding template file SHALL exist
**And** both files SHALL be identical (same hash)

---

### Scenario 2: 9-Field Schema 명시

**Given** the policy document exists

**When** the user reads §2 Schema

**Then** the section SHALL document 9 fields: pr_id, author, claude_assisted_lines, total_lines, confidence, spec_ids, timestamp, merge_strategy, branch_type
**And** each field SHALL include type and example

---

### Scenario 3: Confidence Threshold 0.70 명시

**Given** the policy document exists

**When** the user reads §3 Confidence Calculation Heuristic

**Then** the section SHALL state confidence threshold = 0.70 (conservative measurement per Anthropic)
**And** the section SHALL document 5 factors with weights summing to 1.0
**And** the section SHALL include the Anthropic verbatim quote

---

### Scenario 4: Privacy Policy

**Given** the policy document exists

**When** the user reads §4 Privacy Policy

**Then** the section SHALL state author = GitHub username only (no email)
**And** the section SHALL include PII regex detection (email + API key patterns)
**And** the section SHALL state PII detection → entry skip

---

### Scenario 5: Sync Workflow Integration

**Given** `metrics.enabled: true` in `.moai/config/sections/metrics.yaml`
**And** `/moai sync` completes a PR creation

**When** the workflow processes metric collection

**Then** the workflow SHALL append a JSONL entry to `.moai/metrics/contribution.jsonl`
**And** the entry SHALL contain all 9 fields
**And** sync SHALL NOT be blocked if metric collection fails (best-effort)

---

### Scenario 6: Confidence < 0.70 → Conservative Bias

**Given** a PR with no SPEC-ID reference and no `🗿 MoAI` co-author
**And** computed confidence = 0.40

**When** the workflow computes claude_assisted_lines

**Then** the field SHALL be set to 0 (conservative bias)
**And** the entry SHALL still be persisted with confidence = 0.40 and assisted_lines = 0

---

### Scenario 7: Opt-out (metrics.enabled: false)

**Given** `.moai/config/sections/metrics.yaml metrics.enabled: false`
**And** `/moai sync` completes a PR creation

**When** the workflow processes metric collection

**Then** the workflow SHALL skip metric collection entirely
**And** `.moai/metrics/contribution.jsonl` SHALL NOT be modified

---

### Scenario 8: PII Detection → Entry Skip

**Given** a PR body contains an email address (e.g., "user@example.com")
**And** `/moai sync` completes a PR creation

**When** the workflow processes metric collection

**Then** the workflow SHALL detect the email via regex
**And** the workflow SHALL skip the metric entry (no PII leak)
**And** the workflow SHALL emit a non-blocking note: "PII detected, metric entry skipped"

---

## Edge Cases

### EC-1: `.moai/metrics/` 디렉토리 부재
First metric write SHALL auto-create the directory before appending JSONL.

### EC-2: Author Identification 불가
If PR metadata lacks author, the field SHALL be set to "unknown" (no email leak fallback).

### EC-3: SPEC-ID 누락 PR
If PR has no SPEC-ID reference, confidence SHALL be capped at 0.50 (below threshold). claude_assisted_lines = 0.

### EC-4: GitHub Analytics 통합 (Phase 2)
Phase 1 = local JSONL only. Phase 2 SPEC SHALL handle GitHub Analytics push integration. Until then, all metrics stay local.

### EC-5: Metric Collection Failure
If metric collection fails (filesystem error, regex panic, etc.), sync SHALL continue (best-effort). Failure SHALL be logged to stderr.

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| Policy document | both local + template | file existence |
| 9-field schema | all fields documented | grep |
| Confidence threshold | 0.70 stated | grep |
| Privacy policy | email exclusion + PII regex | grep |
| Opt-in default | `metrics.enabled: false` | grep `metrics.yaml` |
| Sync integration | workflow step documented | grep `sync.md` |
| 5 sample sync | all entries valid JSONL | manual sample |
| False positive rate | 0% (conservative bias) | manual review |
| Cross-references | >= 2 (SKILL.md + sync.md) | grep count |
| Template-First sync | clean | `make build` diff |
| plan-auditor | PASS | auditor report |

---

## Definition of Done

- [ ] All 8 Given-When-Then scenarios PASS
- [ ] All 5 edge cases (EC-1 to EC-5) documented and handled
- [ ] All 11 quality gate criteria meet threshold
- [ ] Policy document at `.claude/rules/moai/workflow/contribution-metrics.md` and template
- [ ] 9-field schema documented with type + example
- [ ] Confidence threshold 0.70 + 5-factor heuristic documented
- [ ] Privacy policy: author = GitHub username, email exclusion, PII regex
- [ ] `.moai/config/sections/metrics.yaml` neutral default (`enabled: false`)
- [ ] `/moai sync` workflow step documented
- [ ] `.moai/metrics/.gitkeep` or auto-create on first write
- [ ] Opt-in policy enforced (skip when `enabled: false`)
- [ ] PII detection regex (email + API key)
- [ ] Best-effort guarantee: sync NOT blocked on metric failure
- [ ] `moai-workflow-project` SKILL.md cross-ref added
- [ ] sync.md skill cross-ref added
- [ ] `make build` regenerates embedded.go cleanly
- [ ] CHANGELOG.md updated under Unreleased
- [ ] No Go code change (Phase 1 documentation-only verified by `git diff`)
- [ ] plan-auditor PASS
- [ ] dogfooding: at least 5 sync invocations validated

End of acceptance.md (SPEC-METRICS-001).

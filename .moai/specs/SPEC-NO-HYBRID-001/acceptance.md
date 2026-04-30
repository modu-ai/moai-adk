---
id: SPEC-NO-HYBRID-001
acceptance_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-NO-HYBRID-001

## Given-When-Then Scenarios

### Scenario 1: 정책 문서 존재

**Given** the SPEC-NO-HYBRID-001 implementation completes
**And** Template-First sync runs

**When** the user inspects the project's rules directory

**Then** the file `.claude/rules/moai/development/single-responsibility.md` SHALL exist
**And** the corresponding template file SHALL exist
**And** both files SHALL be identical (same hash)

---

### Scenario 2: Trigger Threshold "3+ Distinct Modes" 명시

**Given** the policy document exists

**When** the user reads §2 Trigger Threshold

**Then** the section SHALL state "3+ distinct execution modes" trigger
**And** the section SHALL define "mode" (사용자가 다른 결과 기대) and "distinct" (mode 간 로직 분리 가능)
**And** the section SHALL include both example types: flagged (`/moai project`) and cohesive (`/moai db`)

---

### Scenario 3: Anti-Pattern Catalog 5+

**Given** the policy document exists

**When** the user reads §4 Anti-Pattern Catalog

**Then** the section SHALL contain at least 5 anti-patterns: AP1 (Hybrid tool confusion), AP2 (Format-based control), AP3 (Mode parameter explosion), AP4 (Implicit subcommand routing), AP5 (Cross-domain handler)
**And** AP1 and AP2 SHALL include verbatim Anthropic quotes from "Seeing Like an Agent"
**And** AP3, AP4, AP5 SHALL include 본 프로젝트 examples

---

### Scenario 4: 5-Step Audit 절차

**Given** the policy document exists

**When** the user reads §6 Audit Procedure

**Then** the section SHALL define 5 steps: (1) Tool/Command Inventory, (2) Multi-Mode Detection, (3) SRP Violation Assessment, (4) Remediation Recommendation, (5) Anti-Pattern Catalog Update
**And** each step SHALL include concrete deliverable

---

### Scenario 5: 1회성 Audit Report

**Given** the SPEC implementation completes

**When** the user inspects `.moai/reports/`

**Then** the file `.moai/reports/srp-audit-2026-04-30.md` SHALL exist
**And** the report SHALL contain at least 10 audited items
**And** the report SHALL identify Multi-Mode-Detection results
**And** the report SHALL recommend remediation priority (High/Medium/Low) for each violation
**And** the report SHALL list follow-up SPEC candidates

---

### Scenario 6: Cohesion vs Hybrid 구분

**Given** the policy document exists

**When** the user reads §3 Cohesion vs Hybrid

**Then** the section SHALL include a 5-criteria table: Mode shared state / Semantic consistency / User mental model / Subcommand clarity / Code reuse
**And** for each criterion, the table SHALL contrast "cohesion (keep)" vs "hybrid (split)"
**And** the section SHALL provide both example types: `/moai db` (cohesive) and `/moai project` (hybrid candidate)

---

### Scenario 7: Audit Identifies `/moai project` as Violation

**Given** the audit report exists
**And** `/moai project` exposes 4 modes (init / analyze / generate / refresh)

**When** the audit applies §3 Cohesion vs Hybrid criteria

**Then** the audit SHALL flag `/moai project` as a violation
**And** the report SHALL list at least 1 follow-up SPEC candidate for `/moai project` separation
**And** the recommendation SHALL include priority (High) and rationale

---

### Scenario 8: Cross-Reference (3+)

**Given** the SPEC implementation completes

**When** the user inspects rule files

**Then** at minimum 3 files SHALL contain cross-reference to `single-responsibility.md`:
  - `.claude/rules/moai/development/agent-authoring.md`
  - `.claude/rules/moai/development/skill-authoring.md`
  - `.claude/skills/moai-foundation-cc/SKILL.md` or CLAUDE.md
**And** each cross-reference SHALL be in a relevant section (not random placement)

---

## Edge Cases

### EC-1: Cohesive Multi-Mode (e.g., `/moai db`)
If a multi-mode tool has documented cohesion (shared state, single domain), it SHALL remain consolidated with subcommand clarity. The audit MAY mark it as "COHESIVE" without violation.

### EC-2: Format-based Control Detection
The audit SHALL identify prompts containing "respond in JSON" or markdown formatting as output control mechanism. The remediation SHALL recommend tool-based structuring.

### EC-3: Backward Break Risk in Remediation
If remediation requires API change, the follow-up SPEC SHALL include migration guide and deprecation period (recommended: 1 minor release).

### EC-4: Multiple Equal-Severity Violations
If audit identifies 3+ equal-severity violations, remediation SHALL be sequenced into separate SPECs to bound scope.

### EC-5: SRP Over-Application
The policy SHALL NOT mandate splitting every multi-mode tool. The "3+ distinct modes" trigger combined with cohesion criteria prevents over-application.

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| Policy document | both local + template | file existence |
| 8 절 작성 | 모든 절 명시 | grep |
| Anti-patterns | >= 5 with 2+ verbatim citations | grep + count |
| 5-step audit procedure | all steps defined | grep |
| Audit report | `.moai/reports/srp-audit-2026-04-30.md` exists | file existence |
| Audit items | >= 10 | grep count |
| Cohesion vs Hybrid table | 5 criteria | grep |
| Cross-references | >= 3 | grep count |
| Template-First sync | clean | `make build` diff |
| plan-auditor | PASS | auditor report |

---

## Definition of Done

- [ ] All 8 Given-When-Then scenarios PASS
- [ ] All 5 edge cases (EC-1 to EC-5) documented and handled
- [ ] All 10 quality gate criteria meet threshold
- [ ] Policy document at `.claude/rules/moai/development/single-responsibility.md` and template
- [ ] 8 sections completed (Overview, Trigger Threshold, Cohesion vs Hybrid, Anti-Pattern Catalog, Format-based vs Tool-based, Audit Procedure, Remediation Sequencing, Migration Considerations)
- [ ] 5+ anti-patterns documented (AP1-AP5)
- [ ] AP1 and AP2 include verbatim Anthropic quotes
- [ ] 5-step audit procedure documented
- [ ] `.moai/reports/srp-audit-2026-04-30.md` produced
- [ ] Audit report identifies 10+ items
- [ ] Audit identifies `/moai project` as violation candidate
- [ ] Follow-up SPEC candidates listed in audit
- [ ] Cohesion vs Hybrid 5-criteria table documented
- [ ] 3+ cross-references in `agent-authoring.md`, `skill-authoring.md`, `moai-foundation-cc` SKILL.md or CLAUDE.md
- [ ] `make build` regenerates embedded.go cleanly
- [ ] CHANGELOG.md updated under Unreleased
- [ ] No Go code change (documentation + audit only verified by `git diff`)
- [ ] plan-auditor PASS
- [ ] No actual tool/workflow split performed (deferred to follow-up SPECs)

End of acceptance.md (SPEC-NO-HYBRID-001).

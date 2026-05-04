---
id: SPEC-WT-DOC-001
acceptance_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-WT-DOC-001

## Given-When-Then Scenarios

### Scenario 1: Shared State Policy 절 존재

**Given** the SPEC-WT-DOC-001 implementation completes

**When** the user reads `.claude/rules/moai/workflow/worktree-integration.md`

**Then** the file SHALL contain a section titled "Shared State Policy"
**And** the section SHALL include a per-file ownership matrix
**And** the section SHALL state "Writer-of-Last-Resort: PR merge to main"

---

### Scenario 2: Concurrency Model 절 존재

**Given** the implementation completes

**When** the user reads the document

**Then** the document SHALL contain a section titled "Concurrency Model"
**And** the section SHALL state explicitly: "Consistency: eventual via PR merge to main; no direct cross-worktree file writes"
**And** the section SHALL include a decision flowchart (Mermaid TD direction per CLAUDE.local.md §17.2)

---

### Scenario 3: Anti-Patterns 카탈로그 5+

**Given** the implementation completes

**When** the user reads the "Anti-Patterns" section

**Then** the section SHALL contain at least 5 anti-pattern entries
**And** each entry SHALL include Symptom, Why bad, Mitigation
**And** the entries SHALL include AP-1 (Direct cross-worktree write), AP-2 (Concurrent SPEC modification), AP-3 (Unterminated reactive loop), AP-4 (Schema drift), AP-5 (Implicit shared mutable state)

---

### Scenario 4: Termination Conditions cross-ref to SPEC-LOOP-TERM-001

**Given** the implementation completes
**And** SPEC-LOOP-TERM-001 (Wave 2) artifact exists

**When** the user reads the "Termination Conditions" section

**Then** the section SHALL include a cross-reference to SPEC-LOOP-TERM-001
**And** the section SHALL cite the termination schema fields: max_iterations, improvement_threshold, escalation_after
**And** the cross-reference SHALL link to the SPEC document

---

### Scenario 5: CLAUDE.md cross-ref 추가

**Given** the implementation completes

**When** the user reads CLAUDE.md §14 "Parallel Execution Safeguards"

**Then** §14 SHALL contain a reference to:
  "Worktree shared state policy: .claude/rules/moai/workflow/worktree-integration.md §Shared State Policy"
**And** the reference SHALL be in the Worktree Isolation Rules subsection

---

### Scenario 6: 코드 변경 0 (documentation-only SPEC)

**Given** the implementation is complete

**When** the user runs `git diff --stat origin/main..HEAD -- '*.go'`

**Then** the output SHALL show zero lines changed in any Go file
**And** `git diff --stat origin/main..HEAD -- 'cmd/'` SHALL show zero lines changed
**And** `git diff --stat origin/main..HEAD -- 'internal/'` SHALL show zero lines changed (except `internal/template/templates/` for sync)

---

### Scenario 7: Anti-pattern code review enforcement

**Given** a PR introduces direct cross-worktree write (AP-1 violation)

**When** a code reviewer (human or plan-auditor) reviews the PR

**Then** the reviewer SHALL identify the violation referencing the cookbook AP-1
**And** the PR SHALL be requested to refactor (e.g., use PR merge instead)
**And** the cookbook SHALL serve as the authoritative reference for the rejection

---

### Scenario 8: Living document marker

**Given** the implementation completes

**When** the user reads the document introduction

**Then** the document SHALL include a "Living Document" marker
**And** the document SHALL include a HISTORY section with the SPEC-WT-DOC-001 entry
**And** the document SHALL state quarterly review obligation
**And** the document SHALL describe the procedure for adding new anti-patterns (PR-based extension)

---

## Edge Cases

### EC-1: Existing worktree-integration.md content
The new sections SHALL be appended to existing content. Existing sections (decision tree, isolation rules, branch naming) SHALL be preserved verbatim.

### EC-2: Mermaid flowchart rendering on docs-site
If the docs-site rebuilds with the updated worktree-integration.md, Mermaid `flowchart TD` SHALL render correctly per CLAUDE.local.md §17.2 standard.

### EC-3: Cross-reference broken link
If SPEC-LOOP-TERM-001 path changes, the cross-reference SHALL use a stable identifier (`SPEC-LOOP-TERM-001`) rather than a hardcoded path.

### EC-4: New anti-pattern proposal
A team member proposing a new anti-pattern SHALL submit a PR adding the entry to the "Anti-Patterns" section. Cookbook quarterly review SHALL evaluate accumulated proposals.

### EC-5: Concurrent SPEC writes (existing behavior)
The current `.moai-worktree-registry.json` mechanism SHALL NOT be modified. The new policy clarifies but does not replace the registry.

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| Shared State Policy section | section exists | grep verification |
| Concurrency Model section | section exists | grep verification |
| Anti-pattern count | >= 5 | section count |
| Each anti-pattern has 3 fields | Symptom, Why bad, Mitigation | section schema |
| Termination cross-ref | SPEC-LOOP-TERM-001 reference | grep verification |
| CLAUDE.md cross-ref | §14 update | grep verification |
| Code change | 0 Go lines | git diff verification |
| Mermaid syntax | TD direction | docs-site build PASS |
| Living document marker | HISTORY + quarterly | grep verification |
| Template-First sync | clean | `make build` diff |
| plan-auditor | PASS | auditor report |

---

## Definition of Done

- [ ] All 8 Given-When-Then scenarios PASS
- [ ] All 5 edge cases (EC-1 to EC-5) documented
- [ ] All 11 quality gate criteria meet threshold
- [ ] `.claude/rules/moai/workflow/worktree-integration.md` extended with 4 new sections
- [ ] Template at `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` synchronized
- [ ] CLAUDE.md §14 cross-reference added
- [ ] Anti-pattern count >= 5 verified
- [ ] SPEC-LOOP-TERM-001 cross-reference present
- [ ] Mermaid flowchart uses TD direction
- [ ] Living document marker + HISTORY section
- [ ] `make build` regenerates embedded.go cleanly
- [ ] CHANGELOG.md updated under Unreleased
- [ ] No Go code changes (verified by `git diff`)
- [ ] No new rule files (single source of truth in worktree-integration.md)
- [ ] plan-auditor PASS
- [ ] dogfooding: at least one PR after merge references the cookbook

End of acceptance.md (SPEC-WT-DOC-001).

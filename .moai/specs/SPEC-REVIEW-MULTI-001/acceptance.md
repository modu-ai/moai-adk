---
id: SPEC-REVIEW-MULTI-001
acceptance_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-REVIEW-MULTI-001

## Given-When-Then Scenarios

### Scenario 1: Medium PR (50-1000 LOC) triggers 3-stage multi-agent review

**Given** a PR with 200 lines changed
**And** `git diff --shortstat` reports 200 LOC

**When** `/moai review` is invoked

**Then** the workflow SHALL execute Stage 1 with 4 detection agents in parallel
**And** Stage 1 agents SHALL be: expert-security, expert-performance, manager-quality, expert-refactoring
**And** Stage 2 verification SHALL run on each Stage 1 finding
**And** Stage 3 severity ranking SHALL produce a Critical/High/Medium/Low breakdown

---

### Scenario 2: Large PR (>1,000 LOC) adds expert-debug to Stage 1

**Given** a PR with 1,500 lines changed

**When** `/moai review` is invoked

**Then** Stage 1 SHALL include 5 agents (the standard 4 plus expert-debug)
**And** all 5 agents SHALL run in parallel
**And** the final report SHALL note the large-PR path was used

---

### Scenario 3: Small PR (<50 LOC) uses single-agent path by default

**Given** a PR with 30 lines changed
**And** `--full` flag is NOT provided

**When** `/moai review` is invoked

**Then** the workflow SHALL invoke the legacy single-agent path (manager-quality)
**And** no Stage 2 or Stage 3 SHALL execute
**And** the final report SHALL note the small-PR token-optimization path was used

---

### Scenario 4: Small PR with `--full` flag forces multi-agent path

**Given** a PR with 30 lines changed
**And** `/moai review --full` is invoked

**When** the workflow runs

**Then** the 3-stage multi-agent review SHALL execute despite the small size
**And** the final report SHALL include all stages

---

### Scenario 5: Verification drops a false positive via the 4-step algorithm

**Given** Stage 1 produces a finding `F-001` of category `security` at `db.go:42` with description "potential SQL injection via string concatenation"
**And** the diff for `db.go:42` shows the call uses `db.QueryContext(ctx, "SELECT ... WHERE id = $1", userID)` (parameterized)

**When** the Stage 2 verifier executes the §5.6 4-step algorithm

**Then** Step 1 (reproducer) SHALL fail because no input flow path produces concatenated SQL → `step1.verified = false, step1.confidence = 0.00`
**And** Step 2 (AST/grep) SHALL fail because pattern reproduction finds no string concatenation flowing into a query call → `step2.verified = false, step2.confidence = 0.00`
**And** Step 3 (CWE mapping) SHALL fail because no CWE-89 evidence is present → `step3.verified = false, step3.confidence = 0.00`
**And** `total_confidence = 0.00 < 0.50` → `verified = false`
**And** the finding SHALL be dropped with `drop_reason` recording `failed_steps: [1, 2, 3]` and the verifier rationale: "Query is parameterized; CWE-89 mapping inapplicable"
**And** the finding SHALL appear in the "Dropped Findings" section with metadata only (no severity assignment)

---

### Scenario 5b: Verification confirms a true positive

**Given** Stage 1 produces a finding `F-002` of category `security` at `auth.go:88` with description "SQL injection via fmt.Sprintf into Exec"
**And** the diff for `auth.go:88` shows `db.Exec(fmt.Sprintf("DELETE FROM users WHERE name = '%s'", name))`

**When** the Stage 2 verifier executes the §5.6 4-step algorithm

**Then** Step 1 (reproducer) SHALL succeed by constructing input `name = "x'; DROP TABLE users; --"` → `step1.verified = true, step1.confidence = 0.40`
**And** Step 2 (AST/grep) SHALL succeed by detecting `fmt.Sprintf(...)` flowing into `db.Exec(...)` in the same statement → `step2.verified = true, step2.confidence = 0.30`
**And** Step 3 (CWE mapping) SHALL succeed mapping to CWE-89 (SQL Injection) with rationale tied to the diff → `step3.verified = true, step3.confidence = 0.20`
**And** `total_confidence = 0.90 >= 0.50` → `verified = true`
**And** the finding SHALL be retained for Stage 3 ranking with the verifier rationale and CWE-89 attached

---

### Scenario 6: Severity elevation in sensitive domain

**Given** `.moai/project/tech.md` declares the project as `domain: payment`
**And** Stage 2 verifies a Security finding with default severity "Medium"

**When** Stage 3 ranker processes the finding

**Then** the ranker SHALL elevate the severity to "High" (one level up due to sensitive domain)
**And** the elevation SHALL be noted in the finding's rationale

---

### Scenario 7: Duplicate finding from multiple agents → deduplicated

**Given** expert-security and expert-refactoring both report a finding at `src/auth.go:88` with similar symptom

**When** Stage 3 ranker processes the candidates

**Then** the ranker SHALL deduplicate the findings into one entry
**And** the entry SHALL list both originating agents in metadata: `["expert-security", "expert-refactoring"]`
**And** the merged finding SHALL retain the higher severity if they differ

---

### Scenario 8: Zero findings → skip Stage 2 and 3

**Given** all 4 Stage 1 agents return zero candidate findings

**When** Stage 2 is about to begin

**Then** the workflow SHALL skip Stage 2 and Stage 3
**And** the final report SHALL state "No findings detected by any of the 4 perspectives"

---

### Scenario 9: Verifier independence is preserved

**Given** Stage 1 finding F1 is produced by expert-security
**And** Stage 2 spawns a verifier for F1

**When** the verifier prompt is constructed

**Then** the prompt SHALL NOT reveal that expert-security produced F1
**And** the verifier SHALL receive only the finding details (symptom, file:line, evidence) without originating agent identity

---

## Edge Cases

### EC-1: Stage 1 agent crashes

If one of the 4 Stage 1 agents fails to produce output, the workflow SHALL log the failure and proceed with findings from the other 3 agents. The final report SHALL note the missing perspective.

### EC-2: Verifier cannot determine

If verifier returns inconclusive (`verified: null`), the finding SHALL be retained with low confidence flag `unverified_inconclusive` and severity capped at "Low" by the ranker.

### EC-3: PR has no diff

If `git diff --shortstat` reports 0 LOC, the workflow SHALL exit with "No changes to review".

### EC-4: tech.md absent

If `.moai/project/tech.md` does not exist or has no domain declaration, the ranker SHALL skip risk-based severity elevation and use baseline severity rules.

### EC-5: Worktree isolation requested but rejected (HARD rule)

If a user attempts to opt review agents into worktree isolation (e.g., via a deprecated `--isolated` flag or a configuration file override), the workflow SHALL refuse the request with a non-zero exit and emit the message: "Review agents are read-only (reviewer role) and per CLAUDE.md §14 [HARD] MUST NOT use isolation: 'worktree'. Request rejected." No fallback path enables worktree isolation for review; this constraint is non-negotiable.

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| 3-stage execution for medium PR | 100% | E2E test with sample PR |
| Detection coverage vs single-agent | > +80% | M0 baseline comparison |
| False-positive rate (verifier-dropped) | < 15% | sample 5 PRs |
| Severity ranking consistency | manual review of 5 PRs | acceptable per ranker rationale |
| Token cost vs single-agent | < +200% | per-PR token measurement |
| Wall-clock time | < +50% (parallelism benefit) | timing measurement |
| Verifier independence | 100% (no originating agent leak) | prompt audit |
| Large PR (>1000 LOC) finding rate | reasonable | sample 3 large PRs |
| Template-First sync | clean diff | `make build` |
| plan-auditor | PASS | auditor report |

---

## Definition of Done

- [ ] All 10 Given-When-Then scenarios PASS (1, 2, 3, 4, 5, 5b, 6, 7, 8, 9)
- [ ] All 5 edge cases documented and handled (EC-1 through EC-5)
- [ ] All 10 quality gate criteria meet threshold
- [ ] M0 baseline 5 PR comparison report at `.moai/reports/review-multi-validation/`
- [ ] review.md and team/review.md updated with 3-stage architecture
- [ ] Final report schema documented and consistent across solo/team modes
- [ ] CHANGELOG.md updated
- [ ] docs-site 4개국어 reference (별도 PR via /moai sync)
- [ ] plan-auditor PASS
- [ ] Template-First diff = 0 after `make build`
- [ ] No new specialized review agent created (general-purpose for verifier/ranker)

End of acceptance.md.

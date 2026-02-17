# Acceptance Criteria: SPEC-GITHUB-WORKFLOW

## Overview

This document defines comprehensive acceptance criteria for the Enhanced GitHub Issues Workflow. All criteria follow Given-When-Then format for clarity and testability.

---

## Phase 1: Issue to SPEC Conversion

### AC1.1: Valid GitHub Issue Conversion

**GIVEN** a GitHub issue #123 with:
- Title: "Add user authentication"
- Body: "Implement JWT-based authentication with login and logout"
- Labels: `feature`, `backend`
- Comments: 2 clarification comments

**WHEN** user executes `/moai:github issues 123 --convert-spec`

**THEN** the system shall:
- Create directory `.moai/specs/SPEC-ISSUE-123/`
- Create file `spec.md` with EARS format requirements
- Create file `plan.md` with implementation milestones
- Create file `acceptance.md` with Given-When-Then scenarios
- Include SPEC metadata: `original_issue: 123`
- Post GitHub comment: "✅ SPEC created: SPEC-ISSUE-123"

**Verification**:
```bash
# Check SPEC directory exists
test -d .moai/specs/SPEC-ISSUE-123

# Check all 3 files exist
test -f .moai/specs/SPEC-ISSUE-123/spec.md
test -f .moai/specs/SPEC-ISSUE-123/plan.md
test -f .moai/specs/SPEC-ISSUE-123/acceptance.md

# Check SPEC metadata
grep "original_issue: 123" .moai/specs/SPEC-ISSUE-123/spec.md

# Check GitHub comment
gh issue view 123 --json comments | grep "SPEC-ISSUE-123"
```

---

### AC1.2: Issue with Missing Optional Fields

**GIVEN** a GitHub issue #124 with:
- Title: "Fix login bug"
- Body: Empty
- Labels: None
- Comments: None

**WHEN** user executes `/moai:github issues 124 --convert-spec`

**THEN** the system shall:
- Create SPEC with available information
- Use default label classification: `feat/issue-124`
- Generate SPEC with placeholder requirements section
- Prompt user to review and enhance SPEC content

**Verification**:
```bash
# Check SPEC created despite missing fields
test -d .moai/specs/SPEC-ISSUE-124

# Check for placeholder indicators
grep -i "review required" .moai/specs/SPEC-ISSUE-124/spec.md
```

---

### AC1.3: Batch Issue Conversion

**GIVEN** 5 open GitHub issues (#125, #126, #127, #128, #129)

**WHEN** user executes `/moai:github issues --all --convert-spec`

**THEN** the system shall:
- Display issue selection UI with all 5 issues
- Allow user to select which issues to convert
- Convert selected issues in sequence
- Display progress: "Converting issue 1/3..."
- Display summary table with results

**Verification**:
```bash
# Check multiple SPECs created
for i in 125 126 127; do
  test -d .moai/specs/SPEC-ISSUE-$i
done

# Check GitHub comments posted
for i in 125 126 127; do
  gh issue view $i --json comments | grep "SPEC-ISSUE-$i"
done
```

---

## Phase 2: Worktree Auto-Creation

### AC2.1: Bug Issue Worktree Creation

**GIVEN** a SPEC `SPEC-ISSUE-130` with:
- Original issue labeled: `bug`
- SPEC documents created successfully

**WHEN** user executes `moai worktree create SPEC-ISSUE-130 --issue 130`

**THEN** the system shall:
- Create worktree at `worktrees/SPEC-ISSUE-130/`
- Create branch: `fix/issue-130`
- Checkout branch in worktree
- Copy SPEC documents to worktree
- Register worktree in `.moai/worktree-registry.json`

**Verification**:
```bash
# Check worktree directory exists
test -d worktrees/SPEC-ISSUE-130

# Check correct branch created
cd worktrees/SPEC-ISSUE-130
git branch --show-current | grep "fix/issue-130"

# Check worktree registered
grep "SPEC-ISSUE-130" .moai/worktree-registry.json

# Check SPEC documents accessible
test -f .moai/specs/SPEC-ISSUE-130/spec.md
```

---

### AC2.2: Feature Issue Worktree Creation

**GIVEN** a SPEC `SPEC-ISSUE-131` with:
- Original issue labeled: `feature`

**WHEN** user executes `moai worktree create SPEC-ISSUE-131 --issue 131`

**THEN** the system shall:
- Create worktree with branch: `feat/issue-131`

**Verification**:
```bash
cd worktrees/SPEC-ISSUE-131
git branch --show-current | grep "feat/issue-131"
```

---

### AC2.3: Documentation Issue Worktree Creation

**GIVEN** a SPEC `SPEC-ISSUE-132` with:
- Original issue labeled: `documentation`

**WHEN** user executes `moai worktree create SPEC-ISSUE-132 --issue 132`

**THEN** the system shall:
- Create worktree with branch: `docs/issue-132`

**Verification**:
```bash
cd worktrees/SPEC-ISSUE-132
git branch --show-current | grep "docs/issue-132"
```

---

### AC2.4: Worktree Navigation with eval

**GIVEN** a worktree exists at `worktrees/SPEC-ISSUE-130/`

**WHEN** user executes `eval $(moai worktree go SPEC-ISSUE-130)`

**THEN** the system shall:
- Change current directory to `worktrees/SPEC-ISSUE-130/`
- User's shell prompt reflects new directory

**Verification**:
```bash
# Execute navigation
eval $(moai worktree go SPEC-ISSUE-130)

# Check current directory
pwd | grep "worktrees/SPEC-ISSUE-130"
```

---

## Phase 3: Plan-Run-Sync Integration

### AC3.1: Automated Plan-Run-Sync Sequence

**GIVEN** a worktree `worktrees/SPEC-ISSUE-130/` with SPEC documents

**WHEN** user navigates to worktree and executes `moai worktree go SPEC-ISSUE-130`

**THEN** the system shall:
- Execute `/moai:1-plan SPEC-ISSUE-130`
- Execute `/moai:2-run SPEC-ISSUE-130`
- Execute `/moai:3-sync SPEC-ISSUE-130`
- Display progress for each phase
- Update worktree status: `implementation_complete`

**Verification**:
```bash
# Check implementation artifacts
test -f src/auth/login.go  # Example implementation file
test -f tests/auth/login_test.go  # Example test file

# Check documentation generated
test -f .moai/docs/api-documentation.md

# Check worktree status
grep "implementation_complete" .moai/worktree-registry.json
```

---

### AC3.2: Quality Gate Validation Success

**GIVEN** implementation completed in worktree with:
- Test coverage: 92%
- Linter: 0 errors, 0 warnings
- Security: OWASP compliant
- Formatting: Consistent

**WHEN** user executes `moai worktree done SPEC-ISSUE-130`

**THEN** the system shall:
- Run TRUST 5 quality validation
- Display quality report: All gates PASSED
- Update worktree status: `ready_for_review`
- Prompt: "Create PR? (y/n)"

**Verification**:
```bash
# Check quality report generated
test -f .moai/quality-report-SPEC-ISSUE-130.md

# Check all gates passed
grep "PASSED" .moai/quality-report-SPEC-ISSUE-130.md | wc -l | grep "5"

# Check worktree status
grep "ready_for_review" .moai/worktree-registry.json
```

---

### AC3.3: Quality Gate Validation Failure

**GIVEN** implementation completed with:
- Test coverage: 65% (below 85% threshold)
- Linter: 3 errors

**WHEN** user executes `moai worktree done SPEC-ISSUE-130`

**THEN** the system shall:
- Run TRUST 5 quality validation
- Display quality report: FAILED gates
- List specific failures:
  - "Test coverage 65% < 85% threshold"
  - "Linter errors: 3"
- Prompt: "Fix issues and retry? (y/n)"
- NOT proceed to PR creation

**Verification**:
```bash
# Check quality report shows failures
grep "FAILED" .moai/quality-report-SPEC-ISSUE-130.md

# Check worktree status not changed
grep "implementation_complete" .moai/worktree-registry.json
```

---

## Phase 4: PR Auto-Review and Merge

### AC4.1: PR Creation

**GIVEN** worktree `SPEC-ISSUE-130` with:
- Quality gates: PASSED
- User confirmed: "Create PR? y"

**WHEN** system creates PR

**THEN** the PR shall include:
- Title: "feat: Add user authentication" (from SPEC)
- Body:
  ```
  ## Summary
  Implements JWT-based authentication with login and logout

  ## Test Plan
  - Unit tests: 92% coverage
  - Integration tests: All passed

  ## Quality Gates
  ✅ Tested: 92% coverage
  ✅ Readable: Linter passed
  ✅ Unified: Formatting consistent
  ✅ Secured: OWASP compliant
  ✅ Trackable: Conventional commits

  Fixes #130
  ```
- Labels: `feature`, `backend` (from original issue)

**Verification**:
```bash
# Get PR number
PR_NUM=$(gh pr list --json number,title | jq -r '.[] | select(.title | contains("user authentication")) | .number')

# Check PR body
gh pr view $PR_NUM --json body | jq -r '.body' | grep "Fixes #130"

# Check labels
gh pr view $PR_NUM --json labels | jq -r '.labels[].name' | grep "feature"
```

---

### AC4.2: Automated PR Review

**GIVEN** PR #456 created for SPEC-ISSUE-130

**WHEN** `team-lead` agent reviews the PR

**THEN** the review shall include:
- Security analysis: No vulnerabilities found
- Performance analysis: No N+1 queries, optimal algorithms
- Quality analysis:
  - Code correctness verified
  - Test coverage adequate
  - Naming conventions followed
  - Error handling complete
- Review decision: APPROVED or REQUEST_CHANGES

**Verification**:
```bash
# Check review posted
gh pr view 456 --json reviews | jq -r '.reviews[].state' | grep "APPROVED"

# Check review body includes all sections
gh pr view 456 --json reviews | jq -r '.reviews[].body' | grep "Security"
gh pr view 456 --json reviews | jq -r '.reviews[].body' | grep "Performance"
gh pr view 456 --json reviews | jq -r '.reviews[].body' | grep "Quality"
```

---

### AC4.3: Auto-Merge Success

**GIVEN** PR #456 with:
- Review: APPROVED by team-lead
- CI/CD: ✅ All checks passed
- Quality gates: ✅ All passed
- User specified: `--auto-merge` flag

**WHEN** auto-merge conditions checked

**THEN** the system shall:
- Merge PR using `gh pr merge 456 --auto`
- Delete feature branch: `fix/issue-130`
- Update SPEC status: `merged`
- Update worktree status: `merged`
- Proceed to Phase 5 (Issue Closure)

**Verification**:
```bash
# Check PR merged
gh pr view 456 --json state | jq -r '.state' | grep "MERGED"

# Check branch deleted
git branch -r | grep -v "fix/issue-130"

# Check SPEC status
grep "merged" .moai/specs/SPEC-ISSUE-130/spec.md

# Check worktree status
grep "merged" .moai/worktree-registry.json
```

---

### AC4.4: Auto-Merge Blocked (CI Failed)

**GIVEN** PR #457 with:
- Review: APPROVED
- CI/CD: ❌ Tests failed
- User specified: `--auto-merge` flag

**WHEN** auto-merge conditions checked

**THEN** the system shall:
- NOT merge PR
- Display warning: "Auto-merge blocked: CI/CD checks failed"
- Prompt: "Review failed tests and retry? (y/n)"
- Keep PR open for manual fixes

**Verification**:
```bash
# Check PR still open
gh pr view 457 --json state | jq -r '.state' | grep "OPEN"

# Check warning logged
grep "Auto-merge blocked: CI/CD checks failed" .moai/logs/github-workflow.log
```

---

### AC4.5: Auto-Merge Blocked (Merge Conflicts)

**GIVEN** PR #458 with:
- Review: APPROVED
- CI/CD: ✅ All checks passed
- Merge conflicts: Detected with main branch

**WHEN** auto-merge conditions checked

**THEN** the system shall:
- NOT merge PR
- Display warning: "Auto-merge blocked: Merge conflicts detected"
- Display conflicting files list
- Prompt: "Resolve conflicts manually and retry? (y/n)"

**Verification**:
```bash
# Check PR still open
gh pr view 458 --json state | jq -r '.state' | grep "OPEN"

# Check conflicts detected
gh pr view 458 --json mergeable | jq -r '.mergeable' | grep "CONFLICTING"
```

---

## Phase 5: Issue Comment and Closure

### AC5.1: Success Comment (Korean)

**GIVEN**:
- PR #456 merged successfully
- User's `conversation_language: ko` in language.yaml
- Original issue: #130

**WHEN** system posts success comment

**THEN** the comment shall be:
```markdown
✅ 이슈가 성공적으로 해결되었습니다!

**구현 내용:**
- 사용자 인증 기능 추가 (JWT 기반)
- 단위 테스트 커버리지 92%
- OWASP 보안 검증 통과

**관련 PR:** #456
**병합 시간:** 2026-02-16 18:30 KST

이슈를 자동으로 종료합니다. 추가 문제가 있으면 새 이슈를 생성해주세요.
```

**Verification**:
```bash
# Check comment posted in Korean
gh issue view 130 --json comments | jq -r '.comments[-1].body' | grep "성공적으로 해결"

# Check PR reference
gh issue view 130 --json comments | jq -r '.comments[-1].body' | grep "#456"
```

---

### AC5.2: Success Comment (English)

**GIVEN**:
- User's `conversation_language: en` in language.yaml

**WHEN** system posts success comment

**THEN** the comment shall be:
```markdown
✅ Issue resolved successfully!

**Implementation Summary:**
- Added user authentication (JWT-based)
- Unit test coverage: 92%
- OWASP security validation passed

**Related PR:** #456
**Merged at:** 2026-02-16 18:30 KST

Closing this issue automatically. Please create a new issue if you encounter further problems.
```

**Verification**:
```bash
# Check comment posted in English
gh issue view 130 --json comments | jq -r '.comments[-1].body' | grep "resolved successfully"
```

---

### AC5.3: Issue Automatic Closure

**GIVEN** success comment posted on issue #130

**WHEN** system closes the issue

**THEN** the system shall:
- Execute: `gh issue close 130`
- Add label: `resolved`
- Update SPEC status: `completed`
- Update SPEC metadata: `closed_at: 2026-02-16T18:30:00+09:00`

**Verification**:
```bash
# Check issue closed
gh issue view 130 --json state | jq -r '.state' | grep "CLOSED"

# Check label added
gh issue view 130 --json labels | jq -r '.labels[].name' | grep "resolved"

# Check SPEC status
grep "completed" .moai/specs/SPEC-ISSUE-130/spec.md

# Check SPEC metadata
grep "closed_at" .moai/specs/SPEC-ISSUE-130/spec.md
```

---

### AC5.4: Closure Retry Logic

**GIVEN** issue closure fails with:
- Error: "GitHub API rate limit exceeded"

**WHEN** system attempts to close issue

**THEN** the system shall:
- Retry up to 3 times with exponential backoff
- Wait times: 5s, 10s, 20s between retries
- Log each retry attempt
- If all retries fail:
  - Log error: "Failed to close issue #130 after 3 retries"
  - Display warning to user: "Please close issue manually"
  - Continue workflow (do not block)

**Verification**:
```bash
# Check retry attempts logged
grep "Retry attempt 1/3" .moai/logs/github-workflow.log
grep "Retry attempt 2/3" .moai/logs/github-workflow.log
grep "Retry attempt 3/3" .moai/logs/github-workflow.log

# Check final error logged
grep "Failed to close issue #130 after 3 retries" .moai/logs/github-workflow.log
```

---

## Phase 6: tmux Session Automation (Optional)

### AC6.1: tmux Session Creation

**GIVEN**:
- tmux is installed and available
- 3 worktrees created: SPEC-ISSUE-130, SPEC-ISSUE-131, SPEC-ISSUE-132
- User specifies: `--tmux` flag

**WHEN** user executes `/moai:github issues --all --tmux`

**THEN** the system shall:
- Create tmux session: `github-issues-2026-02-16-18-30`
- Create 3 panes (one per worktree)
- Layout: Vertical split with max 3 panes visible
- Execute in each pane: `moai worktree go SPEC-ISSUE-{number}`
- Focus on first pane (SPEC-ISSUE-130)

**Verification**:
```bash
# Check tmux session created
tmux list-sessions | grep "github-issues-2026-02-16"

# Check 3 panes exist
tmux list-panes -t github-issues-2026-02-16 | wc -l | grep "3"

# Check current pane directory
tmux display-message -p '#{pane_current_path}' | grep "SPEC-ISSUE-130"
```

---

### AC6.2: tmux Unavailable Fallback

**GIVEN** tmux is NOT installed

**WHEN** user executes `/moai:github issues --all --tmux`

**THEN** the system shall:
- Detect tmux unavailable
- Display warning: "tmux not available, falling back to sequential execution"
- Execute worktrees sequentially without tmux
- Continue workflow normally

**Verification**:
```bash
# Check warning displayed
grep "tmux not available" .moai/logs/github-workflow.log

# Check sequential execution continued
test -d worktrees/SPEC-ISSUE-130
test -d worktrees/SPEC-ISSUE-131
```

---

## Edge Cases and Error Handling

### AC7.1: Disk Space Insufficient

**GIVEN** available disk space < 500MB

**WHEN** user attempts to create worktree

**THEN** the system shall:
- Check disk space before creation
- Display error: "Insufficient disk space (available: 450MB, required: 500MB)"
- Suggest: "Run `moai worktree clean --merged-only` to free space"
- Abort worktree creation

**Verification**:
```bash
# Check disk space warning
grep "Insufficient disk space" .moai/logs/github-workflow.log

# Check worktree not created
test ! -d worktrees/SPEC-ISSUE-999
```

---

### AC7.2: GitHub Authentication Missing

**GIVEN** user has not authenticated GitHub CLI

**WHEN** user executes `/moai:github issues --all`

**THEN** the system shall:
- Detect authentication failure
- Display error: "GitHub authentication required"
- Display instructions: "Run `gh auth login` to authenticate"
- Abort workflow

**Verification**:
```bash
# Check authentication error
grep "GitHub authentication required" .moai/logs/github-workflow.log
```

---

### AC7.3: Invalid SPEC ID

**GIVEN** user attempts to create worktree with invalid SPEC ID

**WHEN** user executes `moai worktree create INVALID-SPEC`

**THEN** the system shall:
- Validate SPEC ID format: `SPEC-ISSUE-{number}`
- Display error: "Invalid SPEC ID format. Expected: SPEC-ISSUE-{number}"
- Display example: "Example: moai worktree create SPEC-ISSUE-123"
- Abort worktree creation

**Verification**:
```bash
# Check validation error
moai worktree create INVALID-SPEC 2>&1 | grep "Invalid SPEC ID format"
```

---

## Performance Acceptance Criteria

### AC8.1: SPEC Conversion Performance

**GIVEN** a GitHub issue with standard content

**WHEN** SPEC conversion executes

**THEN** the conversion shall complete:
- Within 10 seconds for simple issues (< 500 words)
- Within 30 seconds for complex issues (> 500 words)

**Verification**:
```bash
# Measure conversion time
time moai github issues 130 --convert-spec
# Should output: real 0m8.5s (< 10s)
```

---

### AC8.2: Worktree Creation Performance

**GIVEN** a valid SPEC ID

**WHEN** worktree creation executes

**THEN** the creation shall complete:
- Within 5 seconds for normal repositories (< 1GB)
- Within 15 seconds for large repositories (> 1GB)

**Verification**:
```bash
# Measure creation time
time moai worktree create SPEC-ISSUE-130 --issue 130
# Should output: real 0m4.2s (< 5s)
```

---

### AC8.3: Quality Validation Performance

**GIVEN** an implemented worktree

**WHEN** quality validation executes

**THEN** the validation shall complete:
- Within 30 seconds for small codebases (< 100 files)
- Within 60 seconds for medium codebases (< 500 files)
- Within 120 seconds for large codebases (> 500 files)

**Verification**:
```bash
# Measure validation time
time moai worktree done SPEC-ISSUE-130
# Should output: real 0m25.3s (< 30s for small codebase)
```

---

## Integration Acceptance Criteria

### AC9.1: End-to-End Workflow

**GIVEN** a GitHub repository with 3 open issues

**WHEN** user executes complete workflow:
1. `/moai:github issues --all --convert-spec`
2. `moai worktree create SPEC-ISSUE-{number} --issue {number}` for each
3. `moai worktree go SPEC-ISSUE-{number}` for each
4. Implement features
5. `moai worktree done SPEC-ISSUE-{number}` for each
6. `--auto-merge` flag enabled

**THEN** the final state shall be:
- 3 SPECs created in `.moai/specs/`
- 3 worktrees created in `worktrees/`
- 3 PRs created and merged
- 3 issues closed with success comments
- All quality gates passed
- All worktrees status: `merged`

**Verification**:
```bash
# Check all SPECs created
ls -d .moai/specs/SPEC-ISSUE-* | wc -l | grep "3"

# Check all worktrees created
ls -d worktrees/SPEC-ISSUE-* | wc -l | grep "3"

# Check all PRs merged
gh pr list --state merged | wc -l | grep "3"

# Check all issues closed
gh issue list --state closed | wc -l | grep "3"

# Check all quality reports passed
grep -r "All gates PASSED" .moai/quality-report-*.md | wc -l | grep "3"
```

---

## Definition of Done

A feature is considered **DONE** when ALL of the following are met:

### Code Quality
- [ ] All unit tests pass with 85%+ coverage
- [ ] All integration tests pass
- [ ] Linter produces 0 errors, 0 warnings
- [ ] Security scan produces 0 vulnerabilities
- [ ] Code review approved by at least 1 team member

### Documentation
- [ ] User-facing documentation updated
- [ ] API documentation generated and accurate
- [ ] Inline comments for complex logic
- [ ] CHANGELOG entry added

### Testing
- [ ] Manual testing completed on macOS, Linux, Windows (if applicable)
- [ ] Edge cases tested (disk space, network errors, auth failures)
- [ ] Performance benchmarks meet acceptance criteria

### Integration
- [ ] PR merged to main branch
- [ ] CI/CD pipeline green
- [ ] No merge conflicts with main
- [ ] Version bumped appropriately

### User Experience
- [ ] Error messages clear and actionable
- [ ] Progress feedback displayed for long operations
- [ ] Help text updated for new commands
- [ ] Multilingual support verified (ko, en, ja, zh)

---

**SPEC-GITHUB-WORKFLOW** | Acceptance Criteria v1.0.0 | 2026-02-16

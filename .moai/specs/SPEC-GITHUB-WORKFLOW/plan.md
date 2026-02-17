# Implementation Plan: SPEC-GITHUB-WORKFLOW

## Overview

This implementation plan outlines the development strategy for enhancing the GitHub Issues workflow with full SPEC integration, worktree automation, and intelligent PR management.

---

## Milestone Breakdown

### Milestone 1: Issue to SPEC Conversion (Priority: High)

**Objective**: Enable automatic conversion of GitHub issues to EARS-format SPEC documents.

**Components to Implement**:

1. **Issue Data Parser** (`internal/github/issue_parser.go`):
   - Parse GitHub issue JSON from `gh issue view --json`
   - Extract title, body, labels, comments, assignees
   - Identify requirement patterns in issue content
   - Map issue labels to SPEC metadata (bug → fix, feature → feat)

2. **SPEC Generator Integration** (`internal/spec/generator.go`):
   - Invoke `manager-spec` agent with issue content
   - Convert issue requirements to EARS format patterns
   - Generate SPEC ID: `SPEC-ISSUE-{number}`
   - Create 3-file structure: spec.md, plan.md, acceptance.md

3. **Bidirectional Linking** (`internal/github/linker.go`):
   - Add SPEC metadata with original issue number
   - Post GitHub comment with SPEC ID reference
   - Maintain mapping in `.moai/github-spec-registry.json`

**Dependencies**:
- Existing `manager-spec` agent for SPEC creation
- GitHub CLI (`gh`) for issue data retrieval
- moai-workflow-spec skill for SPEC format validation

**Acceptance Criteria**:
- ✅ GitHub issue #123 converts to `.moai/specs/SPEC-ISSUE-123/`
- ✅ SPEC includes all 3 required files (spec.md, plan.md, acceptance.md)
- ✅ SPEC metadata contains `original_issue: 123`
- ✅ GitHub issue receives comment: "SPEC created: SPEC-ISSUE-123"
- ✅ EARS format validation passes for generated requirements

---

### Milestone 2: Worktree Auto-Creation (Priority: High)

**Objective**: Automate Git worktree creation with intelligent branch naming.

**Components to Implement**:

1. **Branch Prefix Detector** (`internal/git/branch_detector.go`):
   - Parse issue labels to determine branch type
   - Mapping: `bug` → `fix/`, `feature|enhancement` → `feat/`, `documentation` → `docs/`
   - Default to `feat/` if no matching label

2. **Worktree CLI Command** (`internal/cli/worktree_create.go`):
   - Command: `moai worktree create SPEC-ISSUE-{number} --issue {number}`
   - Use existing `WorktreeManager.Add()` from `internal/core/git/worktree.go`
   - Create worktree at `worktrees/SPEC-ISSUE-{number}/`
   - Register in `.moai/worktree-registry.json`

3. **Worktree Navigation** (`internal/cli/worktree_go.go`):
   - Command: `moai worktree go SPEC-ISSUE-{number}`
   - Output: `cd worktrees/SPEC-ISSUE-{number}`
   - Support eval pattern: `eval $(moai worktree go SPEC-ISSUE-{number})`

**Dependencies**:
- Existing `internal/core/git/worktree.go` implementation
- Git 2.30+ with worktree support
- Branch prefix detection logic

**Acceptance Criteria**:
- ✅ `moai worktree create SPEC-ISSUE-123 --issue 123` creates worktree
- ✅ Branch name matches pattern: `fix/issue-123` or `feat/issue-123`
- ✅ Worktree registered in `.moai/worktree-registry.json`
- ✅ `moai worktree go SPEC-ISSUE-123` outputs correct `cd` command
- ✅ Eval pattern changes current directory successfully

---

### Milestone 3: Plan-Run-Sync Integration (Priority: High)

**Objective**: Integrate worktree workflow with existing Plan-Run-Sync pipeline.

**Components to Implement**:

1. **Worktree Workflow Orchestrator** (`internal/workflow/worktree_orchestrator.go`):
   - Detect worktree context from current directory
   - Load SPEC document from worktree
   - Execute `/moai:1-plan`, `/moai:2-run`, `/moai:3-sync` sequence
   - Track progress in worktree metadata

2. **Quality Gate Integration** (`internal/quality/worktree_validator.go`):
   - Run TRUST 5 validation after `/moai:2-run`
   - Check test coverage (85% minimum)
   - Verify linter compliance
   - Validate security standards (OWASP)
   - Generate quality report

3. **Worktree Completion Command** (`internal/cli/worktree_done.go`):
   - Command: `moai worktree done SPEC-ISSUE-{number}`
   - Verify quality gates passed
   - Prepare for PR creation
   - Update worktree status to "ready_for_review"

**Dependencies**:
- Existing `/moai:1-plan`, `/moai:2-run`, `/moai:3-sync` commands
- `manager-quality` agent for TRUST 5 validation
- Quality gate configuration in `.moai/config/sections/quality.yaml`

**Acceptance Criteria**:
- ✅ Worktree context automatically detected from current directory
- ✅ Plan-Run-Sync sequence executes successfully in worktree
- ✅ Quality gates validation runs after implementation
- ✅ `moai worktree done` command prepares PR successfully
- ✅ Worktree metadata reflects completion status

---

### Milestone 4: PR Auto-Review and Merge (Priority: Medium)

**Objective**: Automate PR review through `team-lead` and enable conditional auto-merge.

**Components to Implement**:

1. **PR Creator** (`internal/github/pr_creator.go`):
   - Push branch to remote repository
   - Create PR using `gh pr create`
   - PR title: From SPEC title
   - PR body: Include SPEC summary, test plan, `Fixes #{issue_number}`

2. **PR Reviewer** (`internal/github/pr_reviewer.go`):
   - Invoke `team-lead` agent for review
   - Use `manager-quality` agent for TRUST 5 validation
   - Check CI/CD pipeline status via GitHub API
   - Generate review decision (approve, request changes, comment)

3. **Auto-Merge Controller** (`internal/github/auto_merger.go`):
   - Check prerequisites:
     - `--auto-merge` flag specified
     - PR review approved by `team-lead`
     - CI/CD pipeline status green
     - Quality gates validation passed
   - Execute merge via `gh pr merge {number} --auto`
   - Handle merge conflicts gracefully

**Dependencies**:
- GitHub CLI (`gh`) for PR operations
- `manager-quality` agent for review automation
- CI/CD pipeline configuration for status checks

**Acceptance Criteria**:
- ✅ PR created with correct title, body, and issue reference
- ✅ `team-lead` review completes successfully
- ✅ Auto-merge executes when all conditions met
- ✅ Manual review fallback when conditions not met
- ✅ Merge conflicts detected and reported to user

---

### Milestone 5: Issue Comment and Closure (Priority: Medium)

**Objective**: Post multilingual success comments and close issues automatically.

**Components to Implement**:

1. **Comment Generator** (`internal/github/comment_generator.go`):
   - Read `conversation_language` from `.moai/config/sections/language.yaml`
   - Generate success comment template for supported languages
   - Include implementation summary, PR link, merge timestamp
   - Multilingual templates: English, Korean, Japanese, Chinese

2. **Issue Closer** (`internal/github/issue_closer.go`):
   - Post success comment using `gh issue comment {number} --body`
   - Close issue using `gh issue close {number}`
   - Add `resolved` label to issue
   - Update SPEC status to `completed`

3. **Closure Verification** (`internal/github/closure_verifier.go`):
   - Verify PR is merged before closing issue
   - Retry issue closure up to 3 times on failure
   - Log errors for manual intervention if needed

**Dependencies**:
- Language configuration in `.moai/config/sections/language.yaml`
- GitHub CLI (`gh`) for issue operations
- SPEC metadata for status tracking

**Acceptance Criteria**:
- ✅ Success comment posted in user's `conversation_language`
- ✅ Comment includes implementation summary and PR link
- ✅ Issue closed automatically after comment posted
- ✅ `resolved` label added to closed issue
- ✅ SPEC status updated to `completed`

---

### Milestone 6: tmux Session Automation (Priority: Low - Optional)

**Objective**: Enable parallel development with tmux session management.

**Components to Implement**:

1. **tmux Session Manager** (`internal/tmux/session_manager.go`):
   - Create tmux session: `github-issues-{timestamp}`
   - Create one pane per SPEC worktree
   - Auto-execute `moai worktree go SPEC-ISSUE-{number}` in each pane
   - Vertical split layout with max 3 panes visible

2. **Session Detection** (`internal/tmux/detector.go`):
   - Check if tmux is available: `which tmux`
   - Verify tmux version supports required features
   - Fallback to sequential execution if tmux unavailable

**Dependencies**:
- tmux installed and available in PATH
- Multiple worktrees created for parallel development

**Acceptance Criteria**:
- ✅ tmux session created with `--tmux` flag
- ✅ One pane created per worktree
- ✅ Each pane executes `moai worktree go` automatically
- ✅ Graceful fallback when tmux unavailable

---

## Technical Approach

### Architecture Design

**Layer 1: GitHub Integration**
- `internal/github/`: Issue parsing, PR creation, comment generation
- Interfaces with `gh` CLI for all GitHub operations

**Layer 2: SPEC Management**
- `internal/spec/`: SPEC generation, validation, metadata tracking
- Delegates to `manager-spec` agent for EARS format

**Layer 3: Worktree Management**
- `internal/core/git/worktree.go`: Existing worktree operations
- `internal/workflow/`: Worktree-specific workflow orchestration

**Layer 4: Quality Assurance**
- `internal/quality/`: TRUST 5 validation, quality gates
- Interfaces with `manager-quality` agent

**Layer 5: CLI Commands**
- `internal/cli/`: User-facing commands for worktree and GitHub operations
- Argument parsing and command orchestration

### Data Storage

**SPEC-Issue Registry** (`.moai/github-spec-registry.json`):
```json
{
  "version": "1.0.0",
  "mappings": [
    {
      "issue_number": 123,
      "spec_id": "SPEC-ISSUE-123",
      "created_at": "2026-02-16T16:05:00+09:00",
      "status": "completed",
      "pr_number": 456,
      "merged_at": "2026-02-16T18:30:00+09:00"
    }
  ]
}
```

**Worktree Registry** (`.moai/worktree-registry.json`):
```json
{
  "version": "1.0.0",
  "worktrees": {
    "SPEC-ISSUE-123": {
      "path": "worktrees/SPEC-ISSUE-123",
      "branch": "fix/issue-123",
      "base_branch": "main",
      "status": "completed",
      "created_at": "2026-02-16T16:10:00+09:00",
      "quality_gates_passed": true,
      "pr_number": 456
    }
  }
}
```

### Risk Mitigation

**Risk 1: GitHub API Rate Limits**
- Mitigation: Batch operations with delay between requests
- Fallback: Prompt user to wait and retry after rate limit reset

**Risk 2: Disk Space Exhaustion**
- Mitigation: Monitor disk usage before worktree creation
- Fallback: Cleanup merged worktrees automatically with `moai worktree clean --merged-only`

**Risk 3: Quality Gate Failures**
- Mitigation: Display detailed failure report to user
- Fallback: Create PR as draft for manual review and fixes

**Risk 4: Merge Conflicts**
- Mitigation: Detect conflicts before auto-merge attempt
- Fallback: Skip auto-merge and notify user for manual resolution

**Risk 5: Language Support Gaps**
- Mitigation: Provide English fallback for unsupported languages
- Fallback: Log warning and continue with English comment

---

## Testing Strategy

### Unit Testing

**Test Coverage Target**: 85% minimum per package

**Critical Test Cases**:

1. **Issue Parser**:
   - Valid issue with all fields
   - Issue missing optional fields
   - Invalid JSON format
   - Empty issue body

2. **SPEC Generator**:
   - Issue with clear requirements
   - Issue with vague requirements
   - Issue with multiple requirement types
   - EARS format validation

3. **Worktree Manager**:
   - Create worktree with valid SPEC
   - Create worktree with existing path
   - Navigate to existing worktree
   - Navigate to non-existent worktree

4. **Quality Validator**:
   - All quality gates passed
   - Test coverage below threshold
   - Linter failures
   - Security vulnerabilities detected

5. **PR Reviewer**:
   - PR approved by team-lead
   - PR requires changes
   - CI/CD pipeline failed
   - Merge conflicts detected

6. **Issue Closer**:
   - Successful comment and closure
   - GitHub API error on comment
   - Retry logic verification
   - Language fallback

### Integration Testing

**Test Scenarios**:

1. **End-to-End Workflow**:
   - GitHub issue → SPEC → Worktree → Implementation → PR → Merge → Closure
   - Verify each phase completes successfully
   - Validate data flow between components

2. **Parallel Worktree Development**:
   - Create 3 worktrees for different issues
   - Implement in parallel
   - Verify isolation between worktrees
   - Merge PRs without conflicts

3. **Quality Gate Failure Recovery**:
   - Implement with intentional quality gate failures
   - Verify failure detection
   - Verify user notification
   - Verify draft PR creation

4. **Auto-Merge Conditions**:
   - Test all prerequisite combinations
   - Verify auto-merge when conditions met
   - Verify manual review when conditions not met

### Manual Testing

**Test Checklist**:
- [ ] Test with real GitHub repository
- [ ] Test with multiple issue labels (bug, feature, docs)
- [ ] Test with different conversation languages (ko, en, ja, zh)
- [ ] Test tmux session creation (if available)
- [ ] Test auto-merge with CI/CD pipeline
- [ ] Test manual review fallback
- [ ] Test disk space monitoring
- [ ] Test GitHub API rate limit handling

---

## Dependencies and Sequencing

### Development Sequence

**Phase 1 (Essential - Week 1-2)**:
1. Milestone 1: Issue to SPEC Conversion
2. Milestone 2: Worktree Auto-Creation
3. Milestone 3: Plan-Run-Sync Integration

**Phase 2 (Important - Week 3-4)**:
4. Milestone 4: PR Auto-Review and Merge
5. Milestone 5: Issue Comment and Closure

**Phase 3 (Optional - Week 5)**:
6. Milestone 6: tmux Session Automation

### Dependency Graph

```
Milestone 1 (Issue → SPEC)
      ↓
Milestone 2 (Worktree Creation) ← depends on SPEC
      ↓
Milestone 3 (Plan-Run-Sync) ← depends on Worktree
      ↓
Milestone 4 (PR Review) ← depends on Implementation
      ↓
Milestone 5 (Issue Closure) ← depends on PR Merge
      ↓
Milestone 6 (tmux) ← optional, depends on Worktree
```

---

## Performance Considerations

### Optimization Targets

1. **SPEC Conversion**: < 10 seconds per issue
2. **Worktree Creation**: < 5 seconds per worktree
3. **Quality Validation**: < 30 seconds per SPEC
4. **PR Review**: < 60 seconds per PR
5. **Issue Closure**: < 5 seconds per issue

### Scalability

- **Batch Processing**: Handle up to 50 issues per session
- **Parallel Worktrees**: Support up to 10 concurrent worktrees
- **Disk Usage**: Monitor and warn at 80% disk capacity
- **GitHub API**: Respect rate limits with exponential backoff

---

## Documentation Requirements

### User-Facing Documentation

1. **Command Reference**:
   - `moai worktree create` usage and options
   - `moai worktree go` usage and eval pattern
   - `moai worktree done` usage and quality gates
   - `/moai:github issues --all` enhanced workflow

2. **Configuration Guide**:
   - Language settings for issue comments
   - Workflow settings for auto-merge
   - Quality gate thresholds
   - GitHub authentication setup

3. **Tutorial**:
   - Step-by-step GitHub workflow walkthrough
   - Parallel development with worktrees
   - Auto-merge configuration
   - Troubleshooting common issues

### Developer Documentation

1. **Architecture Documentation**:
   - Component diagram
   - Data flow diagram
   - API reference for internal packages

2. **Testing Documentation**:
   - Unit test patterns
   - Integration test setup
   - Manual testing checklist

3. **Contribution Guide**:
   - Code style guidelines
   - PR submission process
   - Quality gate requirements

---

## Success Metrics

### Quantitative Metrics

- **SPEC Creation Rate**: 90% of GitHub issues successfully converted to SPECs
- **Quality Gate Pass Rate**: 85% of implementations pass on first attempt
- **Auto-Merge Success Rate**: 70% of PRs auto-merge when flag enabled
- **Issue Closure Rate**: 95% of issues automatically closed after PR merge
- **Developer Satisfaction**: 80% positive feedback on workflow efficiency

### Qualitative Metrics

- **Requirement Traceability**: Clear link from issue → SPEC → implementation → PR
- **Code Quality**: Improved consistency through TRUST 5 validation
- **Developer Productivity**: Reduced manual work for branch management and PR review
- **Audit Trail**: Complete history for compliance and debugging

---

## Next Steps

After SPEC approval:

1. **Team Allocation**:
   - Assign backend developer for GitHub integration
   - Assign frontend developer for CLI commands (optional)
   - Assign QA engineer for testing strategy

2. **Kick-off Meeting**:
   - Review SPEC with team
   - Clarify requirements and acceptance criteria
   - Establish communication channels

3. **Sprint Planning**:
   - Break milestones into 2-week sprints
   - Assign story points to each milestone
   - Schedule daily standups

4. **Implementation Start**:
   - Execute `/moai:2-run SPEC-GITHUB-WORKFLOW`
   - Follow DDD cycle: ANALYZE → PRESERVE → IMPROVE
   - Continuous integration with quality gates

---

**SPEC-GITHUB-WORKFLOW** | Implementation Plan v1.0.0 | 2026-02-16

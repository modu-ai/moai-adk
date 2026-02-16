---
id: SPEC-GITHUB-WORKFLOW
title: Enhanced GitHub Issues Workflow with SPEC Integration
status: draft
created_at: 2026-02-16T16:05:00+09:00
priority: high
domains: [cli, git, workflow, integration]
---

# SPEC-GITHUB-WORKFLOW: Enhanced GitHub Issues Workflow

## Background

### Current Workflow Limitations

The existing `/moai:github issues --all` command implements a Team-based parallel analysis and implementation workflow. However, it has several limitations:

1. **No SPEC Integration**: Issues are analyzed and implemented directly without creating formal SPEC documents
2. **Manual Branch Management**: Developers must manually reorganize branches after implementation
3. **Limited Worktree Support**: No automatic worktree creation for parallel development
4. **No Auto-Merge**: PRs require manual review and merge
5. **Missing Issue Closure**: No automatic issue comment and closure after PR merge

### Proposed Enhancement Vision

Transform GitHub Issues into a fully automated SPEC-driven development workflow:

```
GitHub Issue → SPEC Conversion → Worktree Creation → Plan-Run-Sync → PR Review → Auto-Merge → Issue Closure
```

This enhancement aligns with MoAI-ADK's core SPEC-First DDD methodology and enables true parallel development through Git worktrees.

### Business Value

- **Requirement Traceability**: Every GitHub issue becomes a formal SPEC document with EARS format requirements
- **Parallel Development**: Multiple issues can be developed simultaneously in isolated worktrees
- **Quality Assurance**: Automated PR review through manager-quality agent ensures TRUST 5 compliance
- **Developer Productivity**: Reduced manual work for branch management, PR review, and issue closure
- **Audit Trail**: Complete history from issue → SPEC → implementation → PR → merge

---

## Environment

### Prerequisites

- **System Requirements**:
  - moai-adk-go v2.4.3 or later
  - Git 2.30+ with worktree support
  - GitHub CLI (`gh`) v2.40+ installed and authenticated
  - Claude Code v2.1.32+ with Agent Teams support (optional, for parallel execution)

- **Configuration Requirements**:
  - `.moai/config/sections/language.yaml`: User's conversation_language for issue comments
  - `.moai/config/sections/workflow.yaml`: Workflow settings including team mode preferences
  - `.moai/config/sections/system.yaml`: GitHub workflow strategy (github_flow, gitflow, main_direct)

- **GitHub Repository Requirements**:
  - Repository must be initialized with `gh repo view` working
  - User must have write access for branch creation and PR merging
  - CI/CD pipeline configured for PR quality gates (optional for auto-merge)

### Existing Components to Leverage

- **CLI Framework**: `internal/cli/` for command structure and argument parsing
- **Git Operations**: `internal/core/git/worktree.go` for Git worktree management
- **SPEC System**: `manager-spec` agent for SPEC document creation from GitHub issues
- **Quality Validation**: `manager-quality` agent for PR review and TRUST 5 validation
- **Language Support**: `language.yaml` configuration for multilingual issue comments

---

## Assumptions

1. **GitHub Issue Quality**: Issues contain sufficient detail (title, body, labels) for SPEC conversion
   - **Confidence**: Medium
   - **Evidence**: Existing issue templates and label conventions
   - **Risk if Wrong**: SPEC documents may lack detail, requiring manual refinement
   - **Validation**: Test with sample issues before full rollout

2. **CI/CD Availability**: Repository has CI/CD configured for automated testing
   - **Confidence**: Medium
   - **Evidence**: Common practice in modern development workflows
   - **Risk if Wrong**: Auto-merge feature cannot verify quality gates
   - **Validation**: Check for `.github/workflows/` presence during initialization

3. **Worktree Isolation**: Each SPEC can be developed in isolated worktree without conflicts
   - **Confidence**: High
   - **Evidence**: Git worktree design guarantees isolation
   - **Risk if Wrong**: Minimal, Git handles conflicts during merge
   - **Validation**: Existing worktree implementation in moai-adk-go

4. **User Approval Required**: Users want control over PR merge decisions
   - **Confidence**: High
   - **Evidence**: TRUST 5 principle of transparency and user control
   - **Risk if Wrong**: Auto-merge becomes default without consent
   - **Validation**: Implement --auto-merge as opt-in flag

---

## Requirements (EARS Format)

### Phase 1: Issue to SPEC Conversion

**R1.1 - Issue Data Extraction (Ubiquitous)**

The system **shall** extract the following data from GitHub issues using `gh issue view`:
- Issue number
- Issue title
- Issue body content
- Issue labels
- Issue comments
- Issue assignees
- Issue creation timestamp

**R1.2 - EARS Format Conversion (Event-Driven)**

**WHEN** a GitHub issue is selected for SPEC conversion,
**THEN** the system **shall**:
- Analyze issue content to identify requirements patterns
- Convert requirements to EARS format (Ubiquitous, Event-Driven, State-Driven, Unwanted, Optional)
- Extract acceptance criteria from issue description and comments
- Create SPEC ID in format `SPEC-ISSUE-{number}` (e.g., `SPEC-ISSUE-123`)

**R1.3 - SPEC Directory Creation (Event-Driven)**

**WHEN** SPEC conversion is complete,
**THEN** the system **shall** create the following files:
- `.moai/specs/SPEC-ISSUE-{number}/spec.md`: EARS format specification
- `.moai/specs/SPEC-ISSUE-{number}/plan.md`: Implementation plan
- `.moai/specs/SPEC-ISSUE-{number}/acceptance.md`: Acceptance criteria in Given-When-Then format

**R1.4 - Issue Linking (Ubiquitous)**

The system **shall** maintain bidirectional links between GitHub issues and SPEC documents:
- SPEC metadata includes original issue number
- GitHub issue receives comment with SPEC ID reference

---

### Phase 2: Worktree Auto-Creation

**R2.1 - Branch Prefix Detection (Event-Driven)**

**WHEN** creating a worktree for a SPEC,
**THEN** the system **shall** determine branch prefix based on issue labels:
- Label `bug` → branch prefix `fix/issue-{number}`
- Label `feature` or `enhancement` → branch prefix `feat/issue-{number}`
- Label `documentation` → branch prefix `docs/issue-{number}`
- No matching label → default prefix `feat/issue-{number}`

**R2.2 - Worktree Creation Command (Event-Driven)**

**WHEN** user executes `moai worktree create SPEC-ISSUE-{number} --issue {number}`,
**THEN** the system **shall**:
- Create Git worktree at `worktrees/SPEC-ISSUE-{number}/`
- Create branch using detected prefix and issue number
- Initialize worktree with SPEC document reference
- Register worktree in `.moai/worktree-registry.json`

**R2.3 - Worktree Isolation (Ubiquitous)**

The system **shall** ensure each worktree is fully isolated:
- Separate working directory
- Independent branch
- No shared state with other worktrees
- Clean Git status on creation

---

### Phase 3: Plan-Run-Sync Automation

**R3.1 - Worktree Navigation Command (Event-Driven)**

**WHEN** user executes `moai worktree go SPEC-ISSUE-{number}`,
**THEN** the system **shall**:
- Output shell command: `cd worktrees/SPEC-ISSUE-{number}`
- Support eval pattern: `eval $(moai worktree go SPEC-ISSUE-{number})`

**R3.2 - Automated Workflow Execution (Event-Driven)**

**WHEN** user executes `moai worktree go SPEC-ISSUE-{number}` in a worktree,
**THEN** the system **shall** execute:
1. `/moai:1-plan SPEC-ISSUE-{number}`: Create or validate SPEC
2. `/moai:2-run SPEC-ISSUE-{number}`: Implement according to SPEC
3. `/moai:3-sync SPEC-ISSUE-{number}`: Generate documentation and prepare PR

**R3.3 - Quality Gate Validation (Ubiquitous)**

The system **shall** validate TRUST 5 quality gates after implementation:
- **Tested**: 85%+ test coverage
- **Readable**: Code passes linter
- **Unified**: Consistent formatting
- **Secured**: OWASP compliance
- **Trackable**: Conventional commit messages

---

### Phase 4: PR Auto-Review and Merge

**R4.1 - PR Creation Trigger (Event-Driven)**

**WHEN** user executes `moai worktree done SPEC-ISSUE-{number}`,
**THEN** the system **shall**:
- Verify all quality gates passed
- Push branch to remote repository
- Create GitHub PR with title from SPEC and body including:
  - SPEC summary
  - Test plan from acceptance criteria
  - Reference: `Fixes #{issue_number}`

**R4.2 - Automated PR Review (Event-Driven)**

**WHEN** PR is created for a SPEC-ISSUE worktree,
**THEN** the system **shall** invoke `team-lead` agent to:
- Review PR using `manager-quality` agent
- Verify TRUST 5 compliance
- Check test coverage reports
- Validate acceptance criteria

**R4.3 - Auto-Merge Decision (State-Driven)**

**IF** all of the following conditions are met:
- PR review approved by `team-lead`
- CI/CD pipeline status is green
- Quality gates validation passed
- `--auto-merge` flag was specified

**THEN** the system **shall**:
- Merge PR using GitHub API
- Delete feature branch (optional based on configuration)
- Proceed to Phase 5 (Issue Closure)

**R4.4 - Manual Review Fallback (Unwanted)**

The system **shall not** auto-merge PRs when:
- CI/CD pipeline fails
- Quality gates validation fails
- `--auto-merge` flag is not specified
- Merge conflicts detected

---

### Phase 5: Issue Comment and Closure

**R5.1 - Success Comment Generation (Event-Driven)**

**WHEN** PR is merged successfully,
**THEN** the system **shall** generate a comment in user's `conversation_language`:
- Read language from `.moai/config/sections/language.yaml`
- Generate friendly summary of implementation
- Include PR link and merge timestamp

Example (Korean):
```
✅ 이슈가 성공적으로 해결되었습니다!

구현 내용:
- 사용자 인증 기능 추가
- 단위 테스트 커버리지 92%
- OWASP 보안 검증 통과

관련 PR: #456
병합 시간: 2026-02-16 16:30 KST
```

**R5.2 - Issue Closure (Event-Driven)**

**WHEN** success comment is posted,
**THEN** the system **shall**:
- Close the GitHub issue using `gh issue close {number}`
- Add label `resolved` to the issue
- Update SPEC status to `completed`

---

### Phase 6 (Optional): tmux Session Automation

**R6.1 - Session Creation (Optional)**

**WHERE** tmux is available and `--tmux` flag is specified,
the system **shall**:
- Create tmux session named `github-issues-{timestamp}`
- Create one pane per SPEC worktree
- Auto-execute `moai worktree go SPEC-ISSUE-{number}` in each pane

**R6.2 - Session Layout (Optional)**

**WHERE** tmux session is created,
the system **shall** use vertical split layout:
- Maximum 3 panes visible simultaneously
- Horizontal split for additional panes
- Focus on first pane after creation

---

## Specifications

### Component Architecture

```
GitHub CLI (gh)
      ↓
Issue Parser
      ↓
manager-spec (SPEC Conversion)
      ↓
Worktree Manager (internal/core/git/worktree.go)
      ↓
Plan-Run-Sync Workflow
      ↓
manager-quality (PR Review)
      ↓
GitHub API (PR Merge)
      ↓
Issue Comment & Closure
```

### Data Flow

1. **Issue → SPEC**:
   - Input: GitHub Issue JSON from `gh issue view --json`
   - Processing: `manager-spec` agent with EARS pattern matching
   - Output: SPEC directory with 3 files (spec.md, plan.md, acceptance.md)

2. **SPEC → Worktree**:
   - Input: SPEC ID (e.g., `SPEC-ISSUE-123`)
   - Processing: `WorktreeManager.Add()` with branch prefix detection
   - Output: Git worktree at `worktrees/SPEC-ISSUE-123/`

3. **Worktree → Implementation**:
   - Input: SPEC document in worktree
   - Processing: `/moai:1-plan`, `/moai:2-run`, `/moai:3-sync` sequence
   - Output: Implemented feature with tests and documentation

4. **Implementation → PR**:
   - Input: Completed worktree with quality gates passed
   - Processing: `gh pr create` with SPEC summary
   - Output: GitHub PR with issue reference

5. **PR → Merge**:
   - Input: PR number and review results
   - Processing: `manager-quality` validation + CI/CD check
   - Output: Merged PR (if auto-merge enabled and all checks passed)

6. **Merge → Issue Closure**:
   - Input: Merged PR number and issue number
   - Processing: Multilingual comment generation + `gh issue close`
   - Output: Closed issue with success comment

### Error Handling

**Issue Parsing Errors**:
- Validation: Check for required fields (title, body)
- Recovery: Prompt user to add missing information
- Fallback: Skip issue and continue with next

**Worktree Creation Errors**:
- Validation: Check disk space and Git repository status
- Recovery: Clean up partial worktree and retry
- Fallback: Suggest manual worktree creation

**Quality Gate Failures**:
- Validation: Run TRUST 5 checks before PR creation
- Recovery: Display failures to user with fix suggestions
- Fallback: Create PR as draft for manual review

**PR Merge Conflicts**:
- Validation: Check for conflicts before auto-merge
- Recovery: Notify user and skip auto-merge
- Fallback: User resolves conflicts manually

**Issue Closure Errors**:
- Validation: Verify PR is merged before closing issue
- Recovery: Retry issue closure up to 3 times
- Fallback: Log error and continue (issue can be closed manually)

---

## Constraints

### Technical Constraints

1. **Git Worktree Limitation**: Maximum ~100 active worktrees per repository
   - WHY: Git performance degrades with excessive worktrees
   - IMPACT: Large batch processing may need cleanup between batches

2. **GitHub API Rate Limits**: 5000 requests/hour for authenticated users
   - WHY: GitHub API throttling
   - IMPACT: Batch operations limited to ~50 issues per hour

3. **Disk Space**: Each worktree requires ~50-200MB disk space
   - WHY: Full repository clone per worktree
   - IMPACT: Large repositories may need disk space monitoring

### Business Constraints

1. **User Approval Required**: All auto-merge operations require explicit `--auto-merge` flag
   - WHY: TRUST 5 principle of transparency and user control
   - IMPACT: Default behavior is manual review

2. **Language Support**: Issue comments limited to supported languages in `language.yaml`
   - WHY: Multilingual support requires predefined language configurations
   - IMPACT: Unsupported languages fall back to English

### Security Constraints

1. **GitHub Authentication**: All operations require valid `gh` authentication
   - WHY: GitHub API security requirements
   - IMPACT: Users must run `gh auth login` before using workflow

2. **Branch Protection**: Auto-merge respects branch protection rules
   - WHY: Repository security policies
   - IMPACT: Some PRs may require manual approval even with `--auto-merge`

---

## Traceability

This SPEC relates to:

- **Existing Components**:
  - `.claude/commands/moai/github.md`: Current GitHub Issues workflow
  - `internal/core/git/worktree.go`: Worktree management implementation
  - `.claude/agents/moai/manager-spec.md`: SPEC creation agent
  - `.claude/agents/moai/manager-quality.md`: Quality validation agent

- **Related SPECs**:
  - SPEC-GIT-001: Git operations and branch management
  - SPEC-HOOK-001: Hook system for workflow automation

- **Dependencies**:
  - moai-workflow-spec skill: SPEC format and validation
  - moai-workflow-worktree skill: Worktree command patterns
  - moai-foundation-core skill: TRUST 5 quality framework

---

## Version History

- **v1.0.0** (2026-02-16): Initial SPEC creation
  - Phase 1: Issue to SPEC conversion requirements
  - Phase 2: Worktree auto-creation requirements
  - Phase 3: Plan-Run-Sync automation requirements
  - Phase 4: PR auto-review and merge requirements
  - Phase 5: Issue comment and closure requirements
  - Phase 6: Optional tmux session automation requirements

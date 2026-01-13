# Acceptance Criteria: SPEC-TEST-001

> **SPEC ID**: SPEC-TEST-001
> **Document Version**: 1.0
> **Last Updated**: 2026-01-13

---

## Overview

This document defines detailed acceptance criteria for comprehensive CLI testing coverage, using Given-When-Then format to specify test scenarios for each requirement.

---

## Test Execution Format

All acceptance criteria follow this format:

- **Given**: Precondition and test setup
- **When**: Action or event being tested
- **Then**: Expected outcome and verification

---

## Core Command Acceptance Criteria

### AC1: moai init Command

**Scenario 1.1: Successful initialization in new directory**

```
GIVEN a new empty directory exists
AND user has write permissions
AND moai-adk is properly installed

WHEN user executes "moai init /path/to/project"

THEN the system shall create .claude/ directory with configuration files
AND the system shall create .moai/ directory with project files
AND the system shall create CLAUDE.md file
AND the system shall display success message
AND the system shall exit with code 0
```

**Scenario 1.2: Initialization with existing directory**

```
GIVEN a directory exists with .moai/ configuration
AND user has write permissions

WHEN user executes "moai init /path/to/existing/project"

THEN the system shall prompt for overwrite confirmation
AND upon confirmation, the system shall backup existing configuration
AND the system shall update templates
AND the system shall preserve user settings
```

**Scenario 1.3: Initialization in temporary directory**

```
GIVEN a temporary directory path (/tmp/test-moai-<uuid>)
AND user has write permissions to /tmp

WHEN user executes "moai init /tmp/test-moai-<uuid>"

THEN the system shall create all required directories
AND the system shall create all configuration files
AND the system shall verify file permissions
AND the system shall clean up temporary files after test completion
```

**Scenario 1.4: Initialization with invalid path**

```
GIVEN a path where user lacks write permissions
OR path contains invalid characters

WHEN user executes "moai init /invalid/path"

THEN the system shall display error message indicating permission issue
AND the system shall exit with non-zero code
AND the system shall not create any files
```

---

### AC2: moai update Command

**Scenario 2.1: Successful template update**

```
GIVEN a project initialized with moai init
AND newer templates are available in package
AND user has write permissions

WHEN user executes "moai update"

THEN the system shall create backup of current templates
AND the system shall sync templates from package source
AND the system shall merge user configuration
AND the system shall display update summary
AND the system shall exit with code 0
```

**Scenario 2.2: Update with dry-run flag**

```
GIVEN a project initialized with moai init
AND newer templates are available

WHEN user executes "moai update --dry-run"

THEN the system shall display what would be updated
AND the system shall not modify any files
AND the system shall exit with code 0
```

**Scenario 2.3: Update with network error**

```
GIVEN a project initialized with moai init
AND network connection is unavailable
OR remote repository is inaccessible

WHEN user executes "moai update"

THEN the system shall display network error message
AND the system shall not modify any files
AND the system shall exit with non-zero code
```

**Scenario 2.4: Update in temporary directory**

```
GIVEN a project in temporary directory (/tmp/test-update-<uuid>)
AND newer templates are available

WHEN user executes "moai update" in temporary directory

THEN the system shall complete update successfully
AND the system shall verify template sync
AND the system shall clean up after test
```

---

### AC3: moai doctor Command

**Scenario 3.1: Doctor with healthy project**

```
GIVEN a properly initialized project
AND all configuration files are valid
AND all dependencies are installed

WHEN user executes "moai doctor"

THEN the system shall run all diagnostic checks
AND the system shall display all checks as PASSED
AND the system shall display project health status
AND the system shall exit with code 0
```

**Scenario 3.2: Doctor with configuration errors**

```
GIVEN a project with corrupted configuration file
OR missing required configuration

WHEN user executes "moai doctor"

THEN the system shall detect configuration error
AND the system shall display specific error details
AND the system shall suggest fix actions
AND the system shall exit with non-zero code
```

**Scenario 3.3: Doctor with missing dependencies**

```
GIVEN a project with missing Python dependencies
OR missing required tools (git, ast-grep)

WHEN user executes "moai doctor"

THEN the system shall identify missing dependencies
AND the system shall display installation commands
AND the system shall mark check as FAILED
```

**Scenario 3.4: Doctor verbose output**

```
GIVEN a properly initialized project

WHEN user executes "moai doctor --verbose"

THEN the system shall display detailed check information
AND the system shall show check execution time
AND the system shall show configuration values
```

---

### AC4: moai status Command

**Scenario 4.1: Status display for initialized project**

```
GIVEN a properly initialized project
AND configuration files exist

WHEN user executes "moai status"

THEN the system shall display project name
AND the system shall display MoAI-ADK version
AND the system shall display template status
AND the system shall display current configuration
AND the system shall exit with code 0
```

**Scenario 4.2: Status with JSON output**

```
GIVEN a properly initialized project

WHEN user executes "moai status --format json"

THEN the system shall output status in JSON format
AND the system shall include all status fields
AND the system shall produce valid JSON
```

**Scenario 4.3: Status in non-project directory**

```
GIVEN a directory without .moai/ configuration
OR not a moai project

WHEN user executes "moai status"

THEN the system shall display "not a moai project" message
AND the system shall exit with non-zero code
```

---

## Interactive Command Acceptance Criteria

### AC5: moai switch Command

**Scenario 5.1: Switch to valid agent**

```
GIVEN a properly initialized project
AND agent "expert-backend" exists

WHEN user executes "moai switch expert-backend"

THEN the system shall update active agent configuration
AND the system shall confirm agent switch
AND the system shall persist configuration
```

**Scenario 5.2: Switch with invalid agent**

```
GIVEN a properly initialized project

WHEN user executes "moai switch non-existent-agent"

THEN the system shall display error "agent not found"
AND the system shall suggest available agents
AND the system shall exit with non-zero code
```

**Scenario 5.3: Switch with auto-completion**

```
GIVEN a properly initialized project
AND user types "moai switch exp"

WHEN user presses TAB for auto-completion

THEN the system shall suggest matching agents
AND the system shall display available options
```

---

### AC6: moai rank Command

**Scenario 6.1: Rank all agents**

```
GIVEN a properly initialized project
AND multiple agents are available

WHEN user executes "moai rank"

THEN the system shall display ranked list of agents
AND the system shall show agent relevance scores
AND the system shall sort by relevance
```

**Scenario 6.2: Rank with filter**

```
GIVEN a properly initialized project
AND multiple agents exist

WHEN user executes "moai rank --filter backend"

THEN the system shall display only backend-related agents
AND the system shall apply filter criteria
```

**Scenario 6.3: Rank with JSON output**

```
GIVEN a properly initialized project

WHEN user executes "moai rank --format json"

THEN the system shall output rankings in JSON format
AND the system shall include all agent metadata
```

---

### AC7: moai language Command

**Scenario 7.1: Display current language**

```
GIVEN a properly initialized project
AND language is set to "en"

WHEN user executes "moai language"

THEN the system shall display "Current language: English (en)"
AND the system shall exit with code 0
```

**Scenario 7.2: Change to valid language**

```
GIVEN a properly initialized project
AND language "ko" is available

WHEN user executes "moai language ko"

THEN the system shall update language configuration
AND the system shall confirm language change
AND the system shall persist to config file
```

**Scenario 7.3: Change to invalid language**

```
GIVEN a properly initialized project

WHEN user executes "moai language invalid-lang"

THEN the system shall display error "language not supported"
AND the system shall list available languages
AND the system shall exit with non-zero code
```

---

## Worktree Command Acceptance Criteria

### AC8: moai worktree create Command

**Scenario 8.1: Create worktree with existing branch**

```
GIVEN a git repository with branch "feature-123"
AND worktree directory does not exist

WHEN user executes "moai worktree create feature-123"

THEN the system shall create worktree directory
AND the system shall checkout branch in worktree
AND the system shall display worktree path
AND the system shall exit with code 0
```

**Scenario 8.2: Create worktree with non-existent branch**

```
GIVEN a git repository
AND branch "new-feature" does not exist

WHEN user executes "moai worktree create new-feature"

THEN the system shall display error "branch not found"
AND the system shall suggest creating branch first
AND the system shall exit with non-zero code
```

**Scenario 8.3: Create worktree with existing directory**

```
GIVEN a git repository
AND directory "feature-123" already exists

WHEN user executes "moai worktree create feature-123"

THEN the system shall display error "directory already exists"
AND the system shall not modify existing directory
```

**Scenario 8.4: Create worktree in temporary location**

```
GIVEN a git repository in /tmp/test-repo-<uuid>
AND branch "test-feature" exists

WHEN user executes "moai worktree create test-feature"

THEN the system shall create worktree in temporary location
AND the system shall verify git worktree creation
AND the system shall clean up after test
```

---

### AC9: moai worktree list Command

**Scenario 9.1: List all worktrees**

```
GIVEN a git repository with multiple worktrees
AND worktrees exist for branches "feature-1", "feature-2"

WHEN user executes "moai worktree list"

THEN the system shall display all worktrees
AND the system shall show worktree paths
AND the system shall show associated branches
```

**Scenario 9.2: List with no worktrees**

```
GIVEN a git repository
AND no worktrees exist

WHEN user executes "moai worktree list"

THEN the system shall display "No worktrees found"
AND the system shall exit with code 0
```

---

### AC10: moai worktree remove Command

**Scenario 10.1: Remove existing worktree**

```
GIVEN a git repository with worktree "feature-123"
AND worktree has no uncommitted changes

WHEN user executes "moai worktree remove feature-123"

THEN the system shall remove worktree directory
AND the system shall remove git worktree registration
AND the system shall confirm removal
```

**Scenario 10.2: Remove worktree with uncommitted changes**

```
GIVEN a git repository with worktree "feature-123"
AND worktree has uncommitted changes

WHEN user executes "moai worktree remove feature-123"

THEN the system shall display warning about uncommitted changes
THEN the system shall prompt for confirmation
AND upon confirmation, the system shall remove worktree
```

**Scenario 10.3: Remove non-existent worktree**

```
GIVEN a git repository
AND worktree "non-existent" does not exist

WHEN user executes "moai worktree remove non-existent"

THEN the system shall display error "worktree not found"
AND the system shall exit with non-zero code
```

---

### AC11: moai worktree prune Command

**Scenario 11.1: Prune worktrees with deleted branches**

```
GIVEN a git repository
AND worktree "old-feature" exists
AND branch "old-feature" was deleted

WHEN user executes "moai worktree prune"

THEN the system shall identify worktrees with deleted branches
AND the system shall remove identified worktrees
AND the system shall display count of pruned worktrees
```

**Scenario 11.2: Prune with no deleted branches**

```
GIVEN a git repository
AND all worktrees have valid branches

WHEN user executes "moai worktree prune"

THEN the system shall display "No worktrees to prune"
AND the system shall exit with code 0
```

---

### AC12: moai worktree switch Command

**Scenario 12.1: Switch to existing worktree**

```
GIVEN a git repository with worktree "feature-123"
AND worktree directory exists at /path/to/feature-123

WHEN user executes "moai worktree switch feature-123"

THEN the system shall change directory to worktree
AND the system shall confirm switch
AND the system shall display current worktree path
```

**Scenario 12.2: Switch to non-existent worktree**

```
GIVEN a git repository
AND worktree "non-existent" does not exist

WHEN user executes "moai worktree switch non-existent"

THEN the system shall display error "worktree not found"
AND the system shall exit with non-zero code
```

---

## Options and Flags Acceptance Criteria

### AC13: Global Options

**Scenario 13.1: Help flag**

```
GIVEN moai-adk is installed

WHEN user executes "moai --help"
OR user executes "moai init --help"

THEN the system shall display command usage
AND the system shall list all available options
AND the system shall show examples
```

**Scenario 13.2: Version flag**

```
GIVEN moai-adk is installed

WHEN user executes "moai --version"

THEN the system shall display version number
AND the system shall display build information
```

**Scenario 13.3: Verbose flag**

```
GIVEN a properly initialized project

WHEN user executes "moai init --verbose /path/to/project"

THEN the system shall display detailed operation logs
AND the system shall show file creation details
AND the system shall show configuration values
```

**Scenario 13.4: Quiet flag**

```
GIVEN a properly initialized project

WHEN user executes "moai init --quiet /path/to/project"

THEN the system shall minimize output
AND the system shall only display errors
AND the system shall not show progress
```

**Scenario 13.5: Dry-run flag**

```
GIVEN a properly initialized project

WHEN user executes "moai update --dry-run"

THEN the system shall display what would be done
AND the system shall not make any changes
AND the system shall exit with code 0
```

**Scenario 13.6: Force flag**

```
GIVEN a directory with existing .moai/ configuration

WHEN user executes "moai init --force /path/to/project"

THEN the system shall skip confirmation prompt
AND the system shall overwrite existing configuration
AND the system shall create backup
```

---

## Integration Workflow Acceptance Criteria

### AC14: Init Workflow

**Scenario 14.1: Complete initialization workflow**

```
GIVEN a new empty directory

WHEN user executes "moai init /path/to/project"
AND user executes "moai doctor" in new project
AND user executes "moai status" in new project

THEN init shall create all required files
AND doctor shall report all checks PASSED
AND status shall display project information
```

**Scenario 14.2: Init with custom options**

```
GIVEN a new empty directory

WHEN user executes "moai init --lang ko --agent expert-backend /path/to/project"

THEN the system shall create project with Korean language
AND the system shall set default agent to expert-backend
AND the system shall verify configuration
```

---

### AC15: Update Workflow

**Scenario 15.1: Complete update workflow**

```
GIVEN a project initialized with older templates
AND newer templates are available

WHEN user executes "moai update"
AND user executes "moai status"

THEN update shall sync new templates
AND status shall show updated template version
AND user configuration shall be preserved
```

**Scenario 15.2: Update with rollback**

```
GIVEN a project before update
AND backup was created during update

WHEN user encounters issues after update
AND user restores from backup

THEN the system shall restore previous templates
AND the system shall restore previous configuration
```

---

### AC16: Worktree Workflow

**Scenario 16.1: Complete worktree lifecycle**

```
GIVEN a git repository with branches "feature-1", "feature-2"

WHEN user executes "moai worktree create feature-1"
AND user executes "moai worktree create feature-2"
AND user executes "moai worktree list"
AND user executes "moai worktree switch feature-1"
AND user executes "moai worktree remove feature-2"
AND user executes "moai worktree list"

THEN all worktree operations shall succeed
AND worktree list shall show only feature-1
AND current directory shall be feature-1 worktree
```

---

## Quality Gate Criteria

### Coverage Requirements

- [ ] Overall code coverage: >= 85%
- [ ] Each command group: >= 85% coverage
- [ ] Error path coverage: >= 80%
- [ ] Edge case coverage: >= 75%

### Test Quality Requirements

- [ ] Zero flaky tests (all tests pass consistently)
- [ ] Test execution time: < 5 minutes for full suite
- [ ] Zero test isolation failures
- [ ] All temporary files cleaned up after tests

### Functional Requirements

- [ ] All 27 CLI commands have test coverage
- [ ] All command options have test coverage
- [ ] All error conditions have test coverage
- [ ] All integration workflows have test coverage

### Documentation Requirements

- [ ] Test documentation exists
- [ ] Test examples provided
- [ ] Coverage report generates correctly
- [ ] Test execution documented in README

---

## Definition of Done

The testing system is complete when:

1. **Coverage**: Overall coverage >= 85%, each command >= 85%
2. **Stability**: All tests pass consistently (no flaky tests)
3. **Performance**: Full suite executes in < 5 minutes
4. **Isolation**: Zero test isolation failures
5. **Completeness**: All 27 commands tested
6. **Documentation**: Test scenarios documented
7. **Integration**: End-to-end workflows tested
8. **Cleanup**: All temporary files removed

---

**END OF ACCEPTANCE CRITERIA**

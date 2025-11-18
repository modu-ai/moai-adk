---
spec_id: SPEC-UPDATE-HOOKS-001
spec_name: Claude Code Hooks Optimization and Structure Consolidation
version: 1.0.0
status: DRAFT
created_date: 2025-11-19
updated_date: 2025-11-19
author: GOOS行
language: English
priority: High
complexity: High
domain: infrastructure
---

# SPEC-UPDATE-HOOKS-001: Claude Code Hooks Optimization and Structure Consolidation

## Executive Summary

Optimize and consolidate the Claude Code Hooks system in the MoAI-ADK package templates by eliminating code duplication (150-200 lines), standardizing import paths across 39 files, refactoring oversized hook files (session_start__auto_cleanup.py: 628 lines → 5 modular files), and improving overall maintainability from 68% to 82%. This is a critical infrastructure enhancement for the release of v0.27.0.

---

## Environment

### Project Context

- **Project Name**: MoAI-ADK
- **Project Version**: 0.26.0 (release/0.26.0 branch)
- **Target Directory**: `src/moai_adk/templates/.claude/hooks/`
- **Scope**: Package template hooks system (primary source of truth)
- **Language**: Python 3.13.9
- **Documentation Language**: English

### Current State

#### Hook File Structure (39 files)

```
hooks/
├── moai/
│   ├── core/                           # DUPLICATE: 180+ lines
│   │   ├── __init__.py
│   │   ├── timeout.py                  # ← Duplicate of shared/core/timeout.py
│   │   ├── version_cache.py            # ← Duplicate of shared/core/version_cache.py
│   │   ├── config_manager.py
│   │   └── context.py
│   ├── shared/
│   │   ├── core/                       # PRIMARY: Correct location
│   │   │   ├── __init__.py
│   │   │   ├── timeout.py              # ← Keep this
│   │   │   ├── version_cache.py        # ← Keep this
│   │   │   ├── checkpoint.py
│   │   │   └── context.py
│   │   ├── config/
│   │   │   ├── agent_skills_mapping.json  # REDUNDANT: Move to templates/config/
│   │   │   └── __init__.py
│   │   └── handlers/
│   │       ├── __init__.py
│   │       └── tool.py
│   └── handlers/                       # EMPTY: Remove
├── session_start__auto_cleanup.py      # 628 LINES: Split into 5 files
├── pre_tool_use__validation.py
├── post_tool_use__logging.py
├── subagent_start__context.py
└── subagent_stop__checkpoint.py
```

#### Import Path Inconsistencies

**Pattern 1** (Relative imports - 15 files):
```python
from .moai.core.timeout import TimeoutManager
from ..config.agent_skills_mapping import AGENT_SKILLS
```

**Pattern 2** (Mixed imports - 8 files):
```python
from moai.shared.core import ContextManager
from .config.agent_skills_mapping import AGENT_SKILLS
```

**Pattern 3** (Absolute imports - 16 files):
```python
from moai.shared.core.timeout import TimeoutManager
from moai.shared.handlers.tool import ToolHandler
```

### Constraints

- Must maintain backward compatibility with existing Hook APIs
- All Hook events must continue to work (SessionStart, PreToolUse, PostToolUse, SubagentStart, SubagentStop)
- Cannot break Claude Code official Hook specifications (cchooks v2.x)
- Hook execution timeout must remain < 1 second (official spec: 5 seconds)
- Local `.claude/hooks/` customizations must be preserved during synchronization
- Package template changes must support `{{PROJECT_DIR}}` variable substitution

---

## Assumptions

### Technical Assumptions

1. **Hook Specification Compliance**: All Hook files conform to Claude Code official Hook specification (cchooks v2.x) with stdin/stdout JSON interface
2. **Code Duplication Sources**: timeout.py and version_cache.py are genuine duplicates with identical functionality (verified by cc-manager analysis)
3. **File Size Impact**: Splitting session_start__auto_cleanup.py (628 lines) will improve readability without performance degradation (< 10ms total overhead)
4. **Import Resolution**: Absolute imports (moai.shared.core.*) will resolve correctly in both template and local contexts
5. **Python Version**: All code uses Python 3.8+ standard library features (pathlib, json, subprocess, typing)

### Process Assumptions

1. **SPEC-First TDD**: Implementation will follow SPEC-First methodology with test-first approach
2. **Phase Parallelization**: Phases 2-4 can execute in parallel (independent file clusters)
3. **Template-First Sync**: Package template changes will automatically synchronize to local `.claude/hooks/` via `/moai:3-sync` command
4. **Version Control**: All changes will be committed to `feature/SPEC-UPDATE-HOOKS-001` branch with eventual merge to main

---

## Requirements

### Functional Requirements

#### FR-1: Code Duplication Elimination

**Purpose**: Remove all duplicated utility modules from `moai/core/` and consolidate into `moai/shared/core/`

**Detailed Requirements**:

| Target File | Current Location | Action | Success Criteria |
|---|---|---|---|
| timeout.py | moai/core/ + moai/shared/core/ | Keep moai/shared/core/, delete moai/core/ | One source of truth |
| version_cache.py | moai/core/ + moai/shared/core/ | Keep moai/shared/core/, delete moai/core/ | Identical functionality |
| config_manager.py | moai/core/ | Verify uniqueness, keep | No duplicate found |
| context.py | moai/core/ + moai/shared/core/ | Verify and consolidate if identical | Consolidated under shared/core |

**Acceptance Criteria**:
- moai/core/ directory is completely removed
- All import statements reference moai.shared.core.*
- 150-200 lines of code eliminated
- All unit tests pass without modification

#### FR-2: Import Path Standardization

**Purpose**: Unify all 39 Hook files to use absolute imports (`from moai.shared.core import ...`)

**Detailed Requirements**:

| Pattern | Current Count | Target | Migration Path |
|---|---|---|---|
| Relative imports (from .moai.core) | 15 files | 0 files | Convert to `from moai.shared.core import ...` |
| Mixed imports | 8 files | 0 files | Consolidate to absolute imports |
| Absolute imports (moai.shared.*) | 16 files | 39 files | Reference standard |

**Acceptance Criteria**:
- 100% of Hook files use `from moai.shared.core import ...` pattern
- 0 relative imports remain
- 0 mixed patterns remain
- All imports are consistent across all 39 files

#### FR-3: Hook File Verification

**Purpose**: Ensure all Hook files comply with Claude Code official Hook specification

**Detailed Requirements**:

| Verification Item | Requirement | Method |
|---|---|---|
| Hook naming | {event}__{description}.py | Parse all filenames |
| Valid events | session_start, pre_tool, post_tool, subagent_start, subagent_stop | Whitelist validation |
| stdin/stdout | JSON format | Test with sample JSON input |
| exit codes | 0 (success), non-zero (failure) | Execute and check return code |
| Timeout | < 1 second execution time | Benchmark each Hook |

**Acceptance Criteria**:
- All 5 Hook files follow {event}__{description}.py naming
- Event validation: 100% of files use valid event names
- Interface validation: stdin/stdout JSON interface confirmed
- Performance: All hooks execute in < 1 second
- Exit codes: Proper error handling with non-zero exit codes

#### FR-4: Non-Hook Code Relocation

**Purpose**: Move utility and CLI code out of hooks/ directory

**Detailed Requirements**:

| File | Current Location | Target Location | Category |
|---|---|---|---|
| spec_status_hooks.py | `.claude/hooks/moai/` | `src/moai_adk/cli/spec_status.py` | CLI utility (not a Hook) |
| agent_skills_mapping.json | `.claude/hooks/moai/shared/config/` | `.moai/config/agent_skills_mapping.json` | Config (not a Hook) |

**Acceptance Criteria**:
- spec_status_hooks.py moved to CLI module with proper imports
- agent_skills_mapping.json moved to .moai/config/
- Old locations removed
- All references updated

#### FR-5: Directory Structure Cleanup

**Purpose**: Remove empty/redundant directories and optimize structure

**Detailed Requirements**:

| Target | Current State | Action | Success Criteria |
|---|---|---|---|
| moai/core/ | Contains duplicates | Delete entire directory | Directory removed |
| moai/handlers/ | Empty | Delete | Directory removed |
| moai/shared/config/ | Contains agent_skills_mapping.json | Move file, delete directory | Directory removed |
| __pycache__ | Git-tracked | Add to .gitignore | Prevent future tracking |

**Acceptance Criteria**:
- moai/core/ removed (0 errors on deletion)
- moai/handlers/ removed
- moai/shared/config/ removed
- .gitignore updated to exclude __pycache__
- File count: 39 → 34 files or fewer

### Non-Functional Requirements

#### NFR-1: Performance Optimization

**Requirement**: Hook execution time must remain < 1 second

**Success Criteria**:
- Baseline: session_start__auto_cleanup.py execution time < 500ms
- After refactoring: Execution time remains < 500ms
- Memory usage: No increase > 5MB
- Caching optimization: Version cache hits > 85%

#### NFR-2: Maintainability Improvement

**Requirement**: Code maintainability score improves from 68% to 82%

**Success Criteria**:
- Average function length: < 100 lines (reduce from max 628)
- Cyclomatic complexity: < 10 per function
- Documentation: 100% of public APIs have docstrings
- Test coverage: ≥ 80%

#### NFR-3: Documentation Completeness

**Requirement**: All Hooks and shared modules are fully documented

**Success Criteria**:
- Each Hook file has header comment explaining purpose and usage
- All public functions have docstrings (Google style)
- README.md for hooks/ directory with architecture overview
- API documentation for moai.shared.core module

### Interface Requirements

#### IR-1: Hook stdin/stdout Interface

**Input Format**:
```json
{
  "hook_name": "session_start",
  "context": {
    "project_root": "/path/to/project",
    "config": { "version": "0.26.0" },
    "timestamp": "2025-11-19T10:30:00Z"
  },
  "metadata": {
    "session_id": "abc123",
    "agent_type": "spec-builder"
  }
}
```

**Output Format**:
```json
{
  "status": "success",
  "result": {
    "checkpoints_saved": 5,
    "cache_invalidated": false
  },
  "errors": [],
  "execution_time_ms": 150
}
```

**Exit Codes**:
- 0: Success
- 1: General error
- 2: Timeout
- 3: Configuration error

#### IR-2: Module Public API

**moai.shared.core Module**:
```python
from moai.shared.core import (
    TimeoutManager,         # Timeout handling
    VersionCache,           # Version caching
    ConfigManager,          # Configuration management
    CheckpointManager,      # Session checkpoints
    ContextManager          # Execution context
)
```

**moai.shared.handlers Module**:
```python
from moai.shared.handlers import (
    ToolHandler,           # Tool invocation handling
    AgentHandler           # Agent delegation handling
)
```

### Design Constraints

#### DC-1: Backward Compatibility

**Constraint**: Existing Hook behavior must not change

**Rules**:
- All Hook execution results must be identical before/after optimization
- Return value format unchanged
- Error messages preserved
- Config file locations must remain accessible

#### DC-2: Package Template Principles

**Constraint**: Template directory (src/moai_adk/templates/) is single source of truth

**Rules**:
- {{PROJECT_DIR}} variables never substituted in templates
- Local projects substitute {{PROJECT_DIR}} at sync time
- Template changes automatically propagate to local `.claude/hooks/`
- Version pinning in templates only

#### DC-3: Git Policies

**Constraint**: Follow MoAI-ADK Git workflow

**Rules**:
- All changes on feature/SPEC-UPDATE-HOOKS-001 branch
- __pycache__ never committed (add to .gitignore)
- .pyc files never committed
- Existing .moai/hooks/ customizations preserved in local projects

---

## Specifications

### Phase 2: Code Duplication Elimination (Critical Path)

#### Task 2.1: Duplicate File Analysis

**Objective**: Identify and map all duplicate code

**Approach**:
1. Read all files in moai/core/ and moai/shared/core/
2. Compare file contents line-by-line
3. Calculate duplication percentage
4. Generate deduplication map

**Deliverables**:
- Deduplication report (JSON format)
- Mapping: moai/core/* → moai/shared/core/*

#### Task 2.2: Code Consolidation

**Objective**: Merge duplicates into moai/shared/core/

**Approach**:
1. For each duplicate file:
   - Copy to moai/shared/core/ (verify no conflict)
   - Update function implementations if needed
   - Add comprehensive docstrings
   - Verify test coverage
2. Delete moai/core/ directory
3. Validate all imports reference consolidated location

**Deliverables**:
- Consolidated timeout.py in moai/shared/core/
- Consolidated version_cache.py in moai/shared/core/
- Verification report (before/after comparison)

#### Task 2.3: Import Path Migration

**Objective**: Update all files to reference consolidated modules

**Approach**:
1. Generate search/replace patterns:
   ```
   FROM: from moai.core.timeout import TimeoutManager
   TO:   from moai.shared.core.timeout import TimeoutManager
   ```
2. Apply to all 39 Hook files
3. Verify no import errors with `python -m py_compile`

**Deliverables**:
- Migration report (39 files updated)
- Test results showing no import errors

#### Task 2.4: Integration Testing

**Objective**: Verify all changes work together

**Approach**:
1. Unit tests for consolidated modules
2. Integration tests for Hook imports
3. End-to-end Hook execution tests
4. Performance baseline tests

**Deliverables**:
- Test report (test_hooks_consolidation.py)
- Performance benchmark (execution time < 500ms)

### Phase 3: Import Path Standardization

#### Task 3.1: Path Analysis

**Objective**: Categorize all import patterns

**Approach**:
1. Parse all .py files in hooks/
2. Extract import statements
3. Categorize into 3 patterns
4. Generate statistics

**Deliverables**:
- Import pattern analysis (Markdown table)
- File mapping for each pattern

#### Task 3.2: Path Conversion

**Objective**: Convert all imports to absolute pattern

**Approach**:
1. For each non-absolute import:
   ```python
   # Pattern 1 (relative) → Pattern 3 (absolute)
   from .moai.core import X  →  from moai.shared.core import X
   from ..config import Y    →  from moai.shared.config import Y
   ```
2. Apply conversions sequentially
3. Validate with `python -m py_compile`

**Deliverables**:
- Conversion log (39 files updated)
- Validation report (0 import errors)

#### Task 3.3: Consistency Verification

**Objective**: Ensure 100% compliance with absolute import standard

**Approach**:
1. Grep all .py files for relative imports (from ., from ..)
2. Verify count = 0
3. Check for mixed patterns (shouldn't exist)
4. Document final state

**Deliverables**:
- Compliance report (100% absolute imports)
- Final import statistics

### Phase 4: Hook File Refactoring (Optional, High Impact)

#### Task 4.1: File Size Analysis

**Objective**: Identify oversized Hook files

**Approach**:
1. Measure all Hook files
2. Identify those > 300 lines
3. Analyze function complexity
4. Plan splitting strategy

**Deliverables**:
- Size analysis report
- Refactoring plan for session_start__auto_cleanup.py

#### Task 4.2: session_start__auto_cleanup.py Splitting

**Objective**: Refactor 628-line file into 5 modular files

**Strategy**:
```
session_start__auto_cleanup.py (628 lines) → 5 files:
├── session_start__checkpoint.py (120 lines) - Save checkpoints
├── session_start__cleanup.py (140 lines) - Clean temp files
├── session_start__config.py (100 lines) - Load configuration
├── session_start__cache.py (110 lines) - Initialize caching
└── session_start__init.py (60 lines) - Orchestrate above
```

**Approach**:
1. Extract logical sections (checkpoint, cleanup, config, cache)
2. Create separate files with clear responsibilities
3. Refactor common code into shared utilities
4. Update hooks to execute modular files

**Deliverables**:
- 5 new Hook files
- Refactoring report showing 100% functional equivalence
- Performance benchmark (total execution time ≤ original)

### Phase 5: Documentation and Validation

#### Task 5.1: Hook Documentation

**Objective**: Document all Hooks with purpose, input/output, and examples

**Deliverables**:
- README.md for hooks/ directory
- Hook specification: each file documented with:
  - Purpose and use cases
  - Input schema (JSON)
  - Output schema (JSON)
  - Example invocation
  - Error handling
  - Performance notes

#### Task 5.2: Module API Documentation

**Objective**: Document all public APIs in moai.shared.*

**Deliverables**:
- moai/shared/core/README.md (TimeoutManager, VersionCache, ConfigManager, etc.)
- moai/shared/handlers/README.md (ToolHandler, AgentHandler)
- Python docstrings (Google style) for all public classes/functions

#### Task 5.3: Architecture Documentation

**Objective**: Provide high-level architecture overview

**Deliverables**:
- hooks-architecture.md with diagrams
- Data flow diagrams (Hook → Shared Modules → Handlers)
- Integration points with Claude Code

### Phase 6: Quality Assurance and Testing

#### Task 6.1: Automated Test Suite

**Objective**: Create comprehensive test coverage

**Deliverables**:
- Unit tests for moai.shared.core.*
- Unit tests for moai.shared.handlers.*
- Integration tests for each Hook
- End-to-end tests for Hook chains

**Test Coverage Target**: ≥ 80%

#### Task 6.2: Performance Benchmarking

**Objective**: Verify no performance degradation

**Benchmarks**:
- Individual Hook execution time (baseline vs. optimized)
- Memory usage (< 50MB per Hook)
- Cache hit rates (> 85%)
- Total Hook chain execution time (< 1.5 seconds)

**Deliverables**:
- Performance report with before/after metrics

#### Task 6.3: Compatibility Testing

**Objective**: Verify backward compatibility

**Test Scenarios**:
1. Hook execution with existing configurations
2. Local .claude/hooks/ customizations preserved
3. Variable substitution ({{PROJECT_DIR}}) works correctly
4. Git operations unaffected

**Deliverables**:
- Compatibility test report
- Regression test results

---

## Acceptance Criteria

### Scenario 1: Code Duplication Elimination Verification

```gherkin
GIVEN the template contains moai/core/ with duplicates of moai/shared/core/ files
WHEN the deduplication phase is executed
THEN moai/core/ directory is completely removed
  AND all import statements reference moai.shared.core
  AND 150-200 lines of duplicate code are eliminated
  AND all functionality remains 100% identical
```

### Scenario 2: Import Path Standardization Verification

```gherkin
GIVEN Hook files contain 3 different import patterns (15 relative, 8 mixed, 16 absolute)
WHEN the standardization phase is executed
THEN all 39 files use absolute imports (from moai.shared.core import ...)
  AND 0 relative imports remain (validated by grep: "from \." = 0 matches)
  AND 0 mixed patterns remain
  AND all imports are syntactically correct (python -m py_compile passes)
```

### Scenario 3: Hook Compliance Verification

```gherkin
GIVEN all Hook files in the hooks/ directory
WHEN Hook compliance validation is executed
THEN all files follow {event}__{description}.py naming (100% compliance)
  AND all events are valid (session_start, pre_tool, post_tool, subagent_start, subagent_stop)
  AND stdin/stdout interface accepts and outputs JSON (tested with sample data)
  AND exit codes are proper (0 for success, non-zero for errors)
  AND execution time is < 1 second for all hooks
```

### Scenario 4: Directory Structure Optimization

```gherkin
GIVEN the current hooks/ directory contains 39 files with empty/redundant directories
WHEN the cleanup phase is executed
THEN moai/core/ directory is removed (0 errors)
  AND moai/handlers/ directory is removed (verified empty before deletion)
  AND moai/shared/config/ directory is removed (config files relocated)
  AND __pycache__ is added to .gitignore (no .pyc files tracked)
  AND total file count reduced from 39 to ≤ 34 files
```

### Scenario 5: Local Synchronization Verification

```gherkin
GIVEN template has been optimized with all changes applied
WHEN local .claude/hooks is synchronized via /moai:3-sync
THEN {{PROJECT_DIR}} variables are correctly substituted with actual path
  AND existing local customizations are preserved (git diff shows no loss)
  AND all hooks execute without errors
  AND functionality is 100% identical to template
```

---

## Traceability

| Requirement ID | SPEC Section | Implementation Phase | Test Scenario | Success Metric |
|---|---|---|---|---|
| FR-1 | Functional Requirements | Phase 2 | Scenario 1 | 0 duplicate lines |
| FR-2 | Functional Requirements | Phase 3 | Scenario 2 | 100% absolute imports |
| FR-3 | Functional Requirements | Phase 5 | Scenario 3 | 100% Hook compliance |
| FR-4 | Functional Requirements | Phase 3 | Scenario 4 | 5 files relocated |
| FR-5 | Functional Requirements | Phase 3 | Scenario 4 | 34 files max |
| NFR-1 | Non-Functional | Phase 6 | Scenario 3 | < 1 second |
| NFR-2 | Non-Functional | Phase 4-5 | All scenarios | 82% maintainability |
| NFR-3 | Non-Functional | Phase 5 | All scenarios | 100% documented |
| IR-1 | Interface Requirements | Phase 5 | Scenario 3 | JSON I/O verified |
| IR-2 | Interface Requirements | Phase 5 | All scenarios | API documented |
| DC-1 | Design Constraints | Phase 6 | Scenario 5 | 100% compatible |
| DC-2 | Design Constraints | Phase 2-3 | All scenarios | SSOT maintained |
| DC-3 | Design Constraints | Phase 3 | Scenario 4 | .gitignore updated |

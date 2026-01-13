---
id: SPEC-TAG-001
version: "1.0.0"
status: "draft"
created: "2026-01-13"
updated: "2026-01-13"
author: "Alfred"
priority: "HIGH"
tags: [tag-system, traceability, tdd-integration, code-spec-mapping, pre-commit]
spec_id: SPEC-TAG-001
---

# Implementation Plan: SPEC-TAG-001 TAG System v2.0 Phase 1

## Implementation Overview

This plan details the implementation of TAG System v2.0 Phase 1, establishing comprehensive traceability between code, SPEC documents, and tests through inline TAG annotations.

### Core Objectives

1. **TAG Annotation Standard**: Define clear syntax for code-SPEC linkage
2. **Automated Validation**: Pre-commit hooks ensure TAG validity
3. **Linkage Management**: Bidirectional mapping database for traceability
4. **Quality Integration**: Seamless integration with MoAI-ADK quality framework

---

## Priority-Based Milestones

### Primary Goal (Priority High)

**Scope**: T1 TAG Pattern Definition, T2 TAG Parser (Comment Extraction)

**Purpose**:
- Establish TAG syntax standard
- Build comment extraction infrastructure
- Enable TAG detection in Python code

**Tasks**:
1. Define TAG pattern specification: `@SPEC SPEC-ID [verb]`
2. Implement TAG parser using `ast-comments` library
3. Create TAG extraction CLI tool for testing
4. Add TAG validation functions (format checking, SPEC-ID validation)
5. Write comprehensive unit tests for parser

**Success Criteria**:
- TAG parser correctly extracts all TAGs from test Python files
- TAG validation detects 100% of malformed SPEC-IDs
- Parser handles edge cases (multi-line comments, syntax errors)
- Unit test coverage 90%+ for TAG parser module

---

### Secondary Goal (Priority High)

**Scope**: T3 Pre-commit Validation Hook

**Purpose**:
- Integrate TAG validation into Git workflow
- Provide immediate feedback on TAG errors
- Ensure SPEC existence verification

**Tasks**:
1. Create pre-commit hook script in `.claude/hooks/moai/`
2. Implement staged file detection (Python files only)
3. Add TAG format validation logic
4. Implement SPEC existence check in `.moai/specs/`
5. Configure hook execution in `.moai/config/sections/quality.yaml`
6. Add hook registration script

**Success Criteria**:
- Hook triggers on all Python file commits
- Invalid TAG format displays clear error message with file:line
- Non-existent SPEC-ID warns but allows commit (warn mode)
- Hook execution time <2 seconds for typical commits

---

### Tertiary Goal (Priority Medium)

**Scope**: T4 Linkage Manager (TAG↔CODE Mapping)

**Purpose**:
- Maintain bidirectional TAG database
- Enable reverse lookup queries
- Support automatic synchronization

**Tasks**:
1. Design linkage database schema (JSON format)
2. Implement LinkageManager class with CRUD operations
3. Add atomic write operations for database updates
4. Create linkage update trigger on TAG extraction
5. Implement cleanup for orphaned TAGs
6. Build query API for TAG lookups

**Success Criteria**:
- Linkage database accurately tracks all TAGs in codebase
- Query performance <100ms for SPEC-ID lookups
- Database updates are atomic (no corruption on interruption)
- Orphaned TAG detection works on file deletion

---

### Final Goal (Priority Medium)

**Scope**: T5 Quality Configuration Integration

**Purpose**:
- Integrate with existing quality framework
- Support configurable validation modes
- Enable per-project customization

**Tasks**:
1. Extend `.moai/config/sections/quality.yaml` with TAG settings
2. Implement configuration loader for TAG validation
3. Add mode switching logic (warn/enforce/off)
4. Create configuration validation functions
5. Document TAG configuration options
6. Add configuration migration script

**Success Criteria**:
- TAG validation respects quality.yaml settings
- Mode switching works dynamically (no restart required)
- Configuration validation prevents invalid settings
- Documentation covers all TAG configuration options

---

### Optional Goal (Priority Low)

**Scope**: Performance Optimization and CLI Tools

**Purpose**:
- Improve TAG extraction performance
- Provide developer-friendly CLI tools
- Support future extensibility

**Tasks**:
1. Implement TAG extraction caching mechanism
2. Create `moai-tag` CLI command with subcommands
3. Add TAG listing command by SPEC-ID
4. Implement reverse lookup command (code location → TAGs)
5. Add TAG statistics and reporting features
6. Optimize parser for large codebases

**Success Criteria**:
- Cached extraction reduces runtime by 70%+
- CLI commands provide intuitive output format
- TAG statistics generate useful metrics
- Parser handles 1000+ file codebase efficiently

---

## Technical Approach

### Architecture Design

#### Module Structure

```
src/moai_adk/tag_system/
├── __init__.py
├── parser.py              # TAG extraction using ast-comments
├── validator.py           # TAG format and SPEC-ID validation
├── linkage.py             # Linkage database management
├── config.py              # Quality configuration integration
├── cli.py                 # moai-tag CLI commands
└── models.py              # Data models (TAG, LinkageEntry)

.claude/hooks/moai/
└── pre_commit__validate_tags.py  # Pre-commit validation hook

.moai/cache/
└── tag-linkage.json       # Linkage database storage

tests/
└── tag_system/
    ├── test_parser.py
    ├── test_validator.py
    ├── test_linkage.py
    └── test_integration.py
```

#### Core Module Design

**1. parser.py** (T2: TAG Parser)

```python
"""TAG extraction from Python source code comments."""

from pathlib import Path
from typing import List
import ast_comments

class TAG:
    """Represents a single TAG annotation."""

    def __init__(self, spec_id: str, verb: str, file_path: Path, line: int):
        self.spec_id = spec_id  # e.g., "SPEC-AUTH-001"
        self.verb = verb  # "impl", "verify", "depends", "related"
        self.file_path = file_path
        self.line = line

def extract_tags(file_path: Path) -> List[TAG]:
    """Extract all @SPEC TAGs from Python file comments."""
    # Implementation using ast-comments
```

**2. validator.py** (T1: TAG Pattern Definition)

```python
"""TAG format and SPEC-ID validation."""

import re
from pathlib import Path

SPEC_ID_PATTERN = re.compile(r'^SPEC-[A-Z]+-\d+$')

def validate_tag_format(tag: TAG) -> bool:
    """Validate TAG format (@SPEC SPEC-ID [verb])."""

def validate_spec_exists(spec_id: str) -> bool:
    """Check if SPEC document exists in .moai/specs/."""
```

**3. linkage.py** (T4: Linkage Manager)

```python
"""Bidirectional TAG↔CODE linkage management."""

from pathlib import Path
import json
from typing import Dict, List

class LinkageManager:
    """Manage TAG↔CODE mapping database."""

    def __init__(self, db_path: Path):
        self.db_path = db_path

    def add_tag(self, tag: TAG) -> None:
        """Add TAG to linkage database (atomic write)."""

    def get_code_locations(self, spec_id: str) -> List[Dict]:
        """Get all code locations for given SPEC-ID."""

    def remove_file_tags(self, file_path: Path) -> None:
        """Remove all TAGs for deleted file."""
```

---

## Technology Stack Specification

### Core Libraries

| Library | Version | Purpose |
|---------|---------|---------|
| Python | 3.13+ | Execution environment |
| ast-comments | 0.1.0+ | Comment extraction from AST |
| pytest | 8.0+ | Test framework |
| pytest-cov | 4.1+ | Coverage measurement |
| ruff | 0.1+ | Linter and formatter |

### Dependency Installation

```bash
# Core dependencies
uv add ast-comments

# Development dependencies
uv add --dev pytest pytest-cov ruff mypy
```

### Type Checking Configuration

```toml
[tool.mypy]
python_version = "3.13"
strict = true
warn_return_any = true
warn_unused_ignores = true
disallow_untyped_defs = true
```

### Testing Configuration

```toml
[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = ["test_*.py"]
python_functions = ["test_*"]
addopts = "--cov=src/moai_adk/tag_system --cov-report=term-missing --cov-report=html"
```

---

## Risk Analysis and Mitigation

### Risk 1: TAG Format Adoption

**Risk Level**: Medium

**Description**: Developers may not adopt TAG annotation pattern consistently

**Mitigation Strategy**:
- Provide clear documentation with examples
- Create interactive TAG insertion helper (IDE integration)
- Start with warn mode (not enforce) to reduce friction
- Add TAG completion to CLI tools

---

### Risk 2: Performance Impact on Git Operations

**Risk Level**: Low

**Description**: Pre-commit hook may slow down commit workflow

**Mitigation Strategy**:
- Optimize TAG extraction with caching
- Only validate changed files (not entire codebase)
- Set timeout to prevent hanging commits
- Provide bypass option (`git commit --no-verify`)

---

### Risk 3: Linkage Database Corruption

**Risk Level**: Medium

**Description**: Concurrent writes may corrupt linkage database

**Mitigation Strategy**:
- Use atomic write operations (temp file + rename)
- Implement file locking for database updates
- Create database backup before modifications
- Add database validation and recovery tools

---

### Risk 4: Configuration Complexity

**Risk Level**: Low

**Description**: Quality configuration options may confuse users

**Mitigation Strategy**:
- Provide sensible defaults (warn mode)
- Document all configuration options clearly
- Create configuration validation tool
- Add configuration migration script for updates

---

## Task Breakdown and Dependencies

### Task Sequence

```
Phase 1: Foundation (T1, T2)
├── Task 1.1: Create TAG system module structure
├── Task 1.2: Implement TAG data model
├── Task 1.3: Build TAG parser with ast-comments
├── Task 1.4: Add TAG format validation
└── Task 1.5: Write parser unit tests

Phase 2: Pre-commit Hook (T3)
├── Task 2.1: Create pre-commit hook script
├── Task 2.2: Implement staged file detection
├── Task 2.3: Add TAG validation logic
├── Task 2.4: Implement SPEC existence check
└── Task 2.5: Register hook in Git configuration

Phase 3: Linkage Manager (T4)
├── Task 3.1: Design linkage database schema
├── Task 3.2: Implement LinkageManager class
├── Task 3.3: Add atomic write operations
├── Task 3.4: Create cleanup for orphaned TAGs
└── Task 3.5: Write linkage manager tests

Phase 4: Quality Integration (T5)
├── Task 4.1: Extend quality.yaml with TAG settings
├── Task 4.2: Implement configuration loader
├── Task 4.3: Add mode switching logic
├── Task 4.4: Create configuration validation
└── Task 4.5: Document TAG configuration options

Phase 5: CLI Tools and Optimization
├── Task 5.1: Implement TAG extraction caching
├── Task 5.2: Create moai-tag CLI command
├── Task 5.3: Add TAG listing and query commands
├── Task 5.4: Implement TAG statistics reporting
└── Task 5.5: Performance optimization and profiling
```

### Dependency Graph

```
TAG Data Model (Foundation)
    ↓
TAG Parser + Validator (T1, T2)
    ↓
Pre-commit Hook (T3) + Linkage Manager (T4)
    ↓
Quality Integration (T5)
    ↓
CLI Tools + Optimization (Optional)
```

---

## Success Metrics and Measurement

### Quantitative Metrics

| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| TAG Detection Accuracy | 100% | Parser test suite |
| TAG Validation Precision | 100% | Validator test suite |
| Pre-commit Hook Runtime | <2s | pytest-benchmark |
| Linkage Database Query Time | <100ms | Linkage query tests |
| Unit Test Coverage | 90%+ | pytest-cov |
| Integration Test Coverage | 85%+ | pytest-cov |

### Qualitative Metrics

- Developer adoption rate (TAGs per 100 lines of code)
- Reduction in time spent locating SPEC references
- Improvement in code-SPEC traceability clarity
- Ease of configuration and customization

---

## Resource Requirements

### Development Resources

- **Development Time**: 8-12 hours (Phases 1-4), 4-6 hours (Phase 5)
- **Testing Time**: 4-6 hours (unit + integration tests)
- **Documentation Time**: 2-3 hours (README + API docs)

### System Resources

- **Disk Space**: <1MB (code + tests + database)
- **Memory**: <50MB (TAG extraction for typical project)
- **CPU**: Minimal (parsing is fast, <0.1s per file)

### External Dependencies

- `ast-comments` library (MIT license)
- Existing MoAI-ADK hook infrastructure
- `.moai/specs/` directory structure

---

## References

### Internal References

- [moai-foundation-core](../../../../.claude/skills/moai-foundation-core) - TRUST 5 framework
- [moai-workflow-spec](../../../../.claude/skills/moai-workflow-spec) - SPEC workflow
- [moai-lang-python](../../../../.claude/skills/moai-lang-python) - Python 3.13 patterns
- [spec.md](./spec.md) - Requirements specification

### External References

- [ast-comments Library](https://github.com/t3rmin4t0r/ast-comments)
- [Pre-commit Framework](https://pre-commit.com/)
- [Python AST Module](https://docs.python.org/3/library/ast.html)
- [JSON Atomic Writes](https://github.com/untitaker/python-atomicwrites)

---

## Next Steps

```bash
# TDD Execution (Start Implementation)
/moai:2-run SPEC-TAG-001

# Documentation Sync (After Implementation)
/moai:3-sync SPEC-TAG-001
```

---
id: SPEC-TAG-001
version: "1.0.0"
status: "draft"
created: "2026-01-13"
updated: "2026-01-13"
author: "Alfred"
priority: "HIGH"
tags: [tag-system, traceability, tdd-integration, code-spec-mapping, pre-commit]
---

# SPEC-TAG-001: TAG System v2.0 Phase 1 Implementation

## HISTORY

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0.0 | 2026-01-13 | Initial SPEC creation | Alfred |

---

## Overview

### Purpose

Implement TAG System v2.0 Phase 1 to establish comprehensive traceability between code, SPEC documents, and tests through inline TAG annotations, enabling SPEC-First TDD workflow with automatic validation and linkage management.

### Scope

- TAG Pattern Definition: `@SPEC SPEC-ID` syntax with optional verbs (impl, verify, depends, related)
- TAG Parser: Comment extraction using `ast-comments` library for Python files
- Pre-commit Validation Hook: Automated TAG format validation and SPEC existence verification
- Linkage Manager: Bidirectional TAG↔CODE mapping database with automatic synchronization
- Quality Configuration: Integration with `.moai/config/sections/quality.yaml` for TAG validation settings

### Background

MoAI-ADK implements SPEC-First TDD methodology where all development begins with clear specifications. However, maintaining traceability between code and SPECs is currently manual and error-prone. TAG System v2.0 establishes automated traceability through inline annotations, ensuring every code change can be traced back to its originating SPEC requirement.

---

## Environment and Assumptions

### Environment

- Python 3.13+ execution environment
- MoAI-ADK project structure with `.moai/specs/` directory
- Git-based version control with pre-commit hooks
- Existing hook system in `.claude/hooks/moai/`
- AST-Grep integration at `src/moai_adk/astgrep/`

### Assumptions

1. All SPEC documents follow EARS format in `.moai/specs/SPEC-XXX/` directories
2. TAG annotations are only supported in comments (not inline with code)
3. Pre-commit hooks are executable and have proper file permissions
4. Linkage database uses JSON format for simplicity
5. TAG validation should not block development (warn mode, not enforce mode)

---

## Requirements (EARS Format)

### T1: TAG Pattern Definition

#### Ubiquitous Requirements (Always Active)

- The system **shall always** use `@SPEC SPEC-ID` format for TAG annotations
- The system **shall always** validate SPEC-ID format as `SPEC-{DOMAIN}-{NUMBER}` (e.g., `SPEC-AUTH-001`)
- The system **shall always** support optional verbs: `impl`, `verify`, `depends`, `related`

#### Event-Driven Requirements (Trigger-Based)

- **WHEN** a developer adds `@SPEC SPEC-ID` annotation **THEN** the system **shall** validate the SPEC-ID format
- **WHEN** a developer uses an optional verb **THEN** the system **shall** parse `@SPEC SPEC-ID impl` format correctly
- **WHEN** TAG format validation fails **THEN** the system **shall** provide clear error message with correct format example

#### State-Driven Requirements (Conditional)

- **IF** a TAG annotation has no verb specified **THEN** the system **shall** default to `impl` (implementation relationship)
- **IF** a SPEC-ID does not exist in `.moai/specs/` **THEN** the system **shall** issue warning but allow commit
- **IF** multiple TAGs exist in one file **THEN** the system **shall** validate all TAGs independently

#### Unwanted Requirements (Prohibited)

- The system **shall not** allow TAG annotations in code strings (only in comments)
- The system **shall not** accept malformed SPEC-IDs (e.g., `spec-001`, `SPEC001`, `SPEC-AUTH-1`)
- The system **shall not** permit TAG annotations without `@SPEC` prefix

#### Optional Requirements (Nice-to-Have)

- **Where possible**, the system **should** support multi-line TAG annotations
- **Where possible**, the system **should** provide auto-completion for existing SPEC-IDs

---

### T2: TAG Parser (Comment Extraction)

#### Ubiquitous Requirements (Always Active)

- The system **shall always** extract TAG annotations from Python file comments
- The system **shall always** use `ast-comments` library for comment extraction
- The system **shall always** distinguish between inline comments and block comments

#### Event-Driven Requirements (Trigger-Based)

- **WHEN** a Python file is parsed **THEN** the system **shall** extract all `@SPEC` TAGs from comments
- **WHEN** a TAG is found **THEN** the system **shall** record file path, line number, and TAG content
- **WHEN** parsing fails **THEN** the system **shall** log error and continue with next file

#### State-Driven Requirements (Conditional)

- **IF** a file has no TAGs **THEN** the system **shall** return empty TAG list
- **IF** a file has syntax errors **THEN** the system **shall** skip TAG extraction and log warning
- **IF** duplicate TAGs exist in same file **THEN** the system **shall** record all occurrences

#### Unwanted Requirements (Prohibited)

- The system **shall not** extract TAGs from string literals (only comments)
- The system **shall not** modify source code during TAG extraction
- The system **shall not** fail entire process on single file parse error

#### Optional Requirements (Nice-to-Have)

- **Where possible**, the system **should** support TypeScript TAG extraction in future phases
- **Where possible**, the system **should** cache TAG extraction results for performance

---

### T3: Pre-commit Validation Hook

#### Ubiquitous Requirements (Always Active)

- The system **shall always** validate TAG format before Git commit
- The system **shall always** check SPEC existence for all referenced SPEC-IDs
- The system **shall always** run in warn mode (allow commit with warnings)

#### Event-Driven Requirements (Trigger-Based)

- **WHEN** git commit is triggered **THEN** the system **shall** scan staged Python files for TAGs
- **WHEN** invalid TAG format is detected **THEN** the system **shall** display error with file path and line number
- **WHEN** non-existent SPEC-ID is referenced **THEN** the system **shall** warn user but allow commit

#### State-Driven Requirements (Conditional)

- **IF** no Python files are staged **THEN** the system **shall** exit silently with success code
- **IF** all TAGs are valid **THEN** the system **shall** exit with success code (0)
- **IF** TAG validation fails **THEN** the system **shall** exit with warning code (1) but allow commit

#### Unwanted Requirements (Prohibited)

- The system **shall not** block commit on TAG validation failure (warn-only mode)
- The system **shall not** require network access for validation (SPECs are local)
- The system **shall not** modify staged files during validation

#### Optional Requirements (Nice-to-Have)

- **Where possible**, the system **should** provide `--fix` option to auto-correct common TAG format errors
- **Where possible**, the system **should** support `--strict` mode to enforce validation (block on failure)

---

### T4: Linkage Manager (TAG↔CODE Mapping)

#### Ubiquitous Requirements (Always Active)

- The system **shall always** maintain bidirectional mapping between TAGs and code locations
- The system **shall always** store linkage data in `.moai/cache/tag-linkage.json`
- The system **shall always** update linkage database on each successful TAG extraction

#### Event-Driven Requirements (Trigger-Based)

- **WHEN** TAG extraction completes **THEN** the system **shall** update linkage database
- **WHEN** code file is deleted **THEN** the system **shall** remove stale TAG entries
- **WHEN** linkage database is queried **THEN** the system **shall** return all code locations for a given SPEC-ID

#### State-Driven Requirements (Conditional)

- **IF** linkage database does not exist **THEN** the system **shall** create new database with empty mappings
- **IF** multiple code locations reference same SPEC-ID **THEN** the system **shall** track all locations
- **IF** SPEC document is deleted **THEN** the system **shall** warn about orphaned TAGs in code

#### Unwanted Requirements (Prohibited)

- The system **shall not** store file contents in linkage database (only paths and line numbers)
- The system **shall not** require manual database maintenance
- The system **shall not** lose linkage data on process interruption (atomic writes)

#### Optional Requirements (Nice-to-Have)

- **Where possible**, the system **should** provide `moai-tag list` command to query TAGs by SPEC-ID
- **Where possible**, the system **should** support reverse lookup (code location → TAGs)

---

### T5: Quality Configuration Integration

#### Ubiquitous Requirements (Always Active)

- The system **shall always** read TAG validation settings from `.moai/config/sections/quality.yaml`
- The system **shall always** respect `tag_validation.enabled` setting
- The system **shall always** follow `tag_validation.mode` (warn/enforce/off)

#### Event-Driven Requirements (Trigger-Based)

- **WHEN** quality configuration is loaded **THEN** the system **shall** apply TAG validation rules
- **WHEN** `tag_validation.enabled` is false **THEN** the system **shall** skip all TAG validation
- **WHEN** `tag_validation.mode` changes **THEN** the system **shall** adjust behavior dynamically

#### State-Driven Requirements (Conditional)

- **IF** `tag_validation.mode` is `warn` **THEN** the system **shall** allow commit with warnings
- **IF** `tag_validation.mode` is `enforce` **THEN** the system **shall** block commit on validation failure
- **IF** `tag_validation.mode` is `off` **THEN** the system **shall** disable all TAG validation

#### Unwanted Requirements (Prohibited)

- The system **shall not** require hardcoded configuration values
- The system **shall not** ignore quality.yaml settings
- The system **shall not** use enforce mode by default (must be explicitly configured)

#### Optional Requirements (Nice-to-Have)

- **Where possible**, the system **should** support per-file-type TAG validation settings
- **Where possible**, the system **should** provide `tag_validation.exceptions` for specific files

---

## Traceability

### Related Documents

- [CLAUDE.md](../../../../CLAUDE.md) - Execution directives
- [moai-foundation-core](../../../../.claude/skills/moai-foundation-core) - SPEC-First TDD methodology
- [moai-workflow-spec](../../../../.claude/skills/moai-workflow-spec) - SPEC creation workflow
- [moai-lang-python](../../../../.claude/skills/moai-lang-python) - Python 3.13 patterns

### Related SPECs

- SPEC-HOOK-001: Hook System Integration (Pre-commit hook infrastructure)
- SPEC-TDD-001: TDD Workflow Integration (Test-driven development with TAG traceability)

### Next Steps

```bash
# TDD Execution
/moai:2-run SPEC-TAG-001

# Documentation Sync
/moai:3-sync SPEC-TAG-001
```

---

## References

- Python AST Comments: [ast-comments](https://github.com/t3rmin4t0r/ast-comments)
- Pre-commit Framework: [pre-commit.com](https://pre-commit.com/)
- EARS Methodology: [Alistair Mavin (2009)](https://ieeexplore.ieee.org/document/5718230)
- TAG System Research: Industry best practices for code-spec traceability

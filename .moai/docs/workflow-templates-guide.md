# Workflow Templates Guide

## Overview

MoAI-ADK provides standardized workflow templates for common development patterns.

## Available Workflows

### 1. SPEC-First Development

Create SPEC documents before implementing features:

```
/alfred:1-plan "Feature Name"
/alfred:2-run SPEC-001
/alfred:3-sync auto
```

### 2. Test-Driven Development (TDD)

Write tests first, then implementation:

```
RED   → Write failing test
GREEN → Write minimal code to pass
REFACTOR → Improve code quality
```

### 3. Documentation Synchronization

Keep documentation in sync with code:

```
/alfred:3-sync auto [SPEC-ID]
```

## Workflow Best Practices

- Plan before code
- Test before features
- Document as you develop
- Keep commits atomic and traceable

## GitFlow Workflow

```
feature/SPEC-XXX → develop → main
```

## Command Structure

All Alfred commands follow a 4-step workflow:

1. **Intent Understanding** - Clarify what needs to be done
2. **Plan Creation** - Design the solution
3. **Task Execution** - Implement with TDD
4. **Report & Commit** - Document and create history


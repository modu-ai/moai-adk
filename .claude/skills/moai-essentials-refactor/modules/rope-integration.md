# Rope Integration Guide

Complete guide to Python Rope library integration for automated refactoring.

## Overview

Rope is a Python refactoring library providing automated code transformations. This module covers Rope API integration, transformation patterns, and safe refactoring workflows.

---

## Rope API Essentials

### Project Initialization

```python
from rope.base.project import Project

class RopeRefactoringEngine:
    """Rope-based refactoring engine."""

    def __init__(self, project_path: str):
        """Initialize Rope project."""
        self.project = Project(project_path)

    def close(self):
        """Close Rope project."""
        self.project.close()
```

### Resource Management

```python
def get_resource(self, file_path: str):
    """Get Rope resource for file."""
    return self.project.get_resource(file_path)

def get_all_python_files(self):
    """Get all Python files in project."""
    return [
        resource for resource in self.project.get_files()
        if resource.name.endswith('.py')
    ]
```

---

## Refactoring Operations

### Extract Method

```python
from rope.refactor.extract import ExtractMethod

def extract_method(self, file_path: str, start: int, end: int, method_name: str):
    """Extract code block into new method."""
    resource = self.project.get_resource(file_path)

    extractor = ExtractMethod(
        self.project,
        resource,
        start,
        end
    )

    changes = extractor.get_changes(method_name)
    self.project.do(changes)
```

### Rename

```python
from rope.refactor.rename import Rename

def rename_symbol(self, file_path: str, offset: int, new_name: str):
    """Rename symbol at offset."""
    resource = self.project.get_resource(file_path)

    renamer = Rename(self.project, resource, offset)
    changes = renamer.get_changes(new_name)
    self.project.do(changes)
```

### Move Method

```python
from rope.refactor.move import MoveMethod

def move_method(self, file_path: str, offset: int, dest_attr: str):
    """Move method to another class."""
    resource = self.project.get_resource(file_path)

    mover = MoveMethod(self.project, resource, offset)
    changes = mover.get_changes(dest_attr)
    self.project.do(changes)
```

### Inline Method

```python
from rope.refactor.inline import InlineMethod

def inline_method(self, file_path: str, offset: int):
    """Inline method at offset."""
    resource = self.project.get_resource(file_path)

    inliner = InlineMethod(self.project, resource, offset)
    changes = inliner.get_changes()
    self.project.do(changes)
```

---

## Safe Refactoring Workflow

### Validation Pipeline

```python
def apply_safe_transformation(self, operation, *args):
    """Apply transformation with validation."""
    # Step 1: Backup
    backup = self._create_backup()

    try:
        # Step 2: Apply transformation
        changes = operation(*args)

        # Step 3: Validate changes
        if not self._validate_changes(changes):
            raise ValidationError("Changes validation failed")

        # Step 4: Apply changes
        self.project.do(changes)

        # Step 5: Run tests
        if not self._run_tests():
            raise TestFailureError("Tests failed")

        return True

    except Exception as e:
        self._restore_backup(backup)
        raise RefactoringError(f"Refactoring failed: {e}")
```

---

## Context7 Integration

### Pattern Matching

```python
async def analyze_with_context7(self, file_path: str):
    """Analyze file with Context7 patterns."""
    # Get Context7 patterns from Rope library
    patterns = await self.context7.get_library_docs(
        context7_library_id="/rope/rope",
        topic="refactoring patterns automated transformation",
        tokens=3000
    )

    # Rope analysis
    rope_analysis = self._analyze_rope(file_path)

    # Match patterns
    matches = self._match_patterns(rope_analysis, patterns)

    return matches
```

---

## Error Handling

### Common Errors

```python
class RopeErrorHandler:
    """Handle Rope-specific errors."""

    def handle_error(self, error):
        """Handle Rope error."""
        if isinstance(error, RefactoringError):
            # Rope refactoring failed
            return self._handle_refactoring_error(error)
        elif isinstance(error, ResourceNotFoundError):
            # File not found
            return self._handle_resource_error(error)
        else:
            # Unknown error
            return self._handle_unknown_error(error)
```

---

**Last Updated**: 2025-11-24
**Status**: Production Ready

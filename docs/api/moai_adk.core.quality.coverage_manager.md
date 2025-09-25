# moai_adk.core.quality.coverage_manager

Coverage management utilities for MoAI-ADK.

Provides test coverage measurement, validation, and reporting capabilities
following TRUST 5 principles for quality assurance.

@FEATURE:QUALITY-COVERAGE Coverage management and validation system

## Functions

### __init__

Initialize coverage manager.

Args:
    project_path: Path to the project root
    security_manager: Security manager instance for validation

```python
__init__(self, project_path, security_manager)
```

### set_minimum_threshold

Set minimum coverage threshold percentage.

Args:
    threshold: Minimum coverage percentage (0-100)

```python
set_minimum_threshold(self, threshold)
```

### measure_coverage

Measure current test coverage percentage.

Returns:
    float: Coverage percentage (0-100)

```python
measure_coverage(self)
```

### validate_coverage

Validate if coverage meets minimum threshold.

Args:
    coverage_percentage: Actual coverage percentage

Returns:
    bool: True if coverage meets threshold

```python
validate_coverage(self, coverage_percentage)
```

### generate_report

Generate comprehensive coverage report.

Returns:
    Dict containing coverage metrics and uncovered files

```python
generate_report(self)
```

### get_uncovered_lines

Get uncovered lines by file.

Returns:
    Dict mapping file paths to uncovered line numbers

```python
get_uncovered_lines(self)
```

### run_pytest_coverage

Run pytest with coverage and return results.

Returns:
    Dict containing coverage results

```python
run_pytest_coverage(self)
```

### set_exclude_patterns

Set file patterns to exclude from coverage.

Args:
    patterns: List of glob patterns to exclude

```python
set_exclude_patterns(self, patterns)
```

## Classes

### CoverageError

Coverage-related exception.

### CoverageManager

Manages test coverage measurement and validation.

@DESIGN:COVERAGE-ARCH-001 Coverage management architecture
Follows single responsibility principle by focusing only on coverage operations.

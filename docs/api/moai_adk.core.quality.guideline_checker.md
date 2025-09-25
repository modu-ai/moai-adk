# moai_adk.core.quality.guideline_checker

Guideline compliance checker for MoAI-ADK.

Validates Python code against TRUST 5 principles development guidelines:
- Function length ≤ 50 LOC
- File size ≤ 300 LOC
- Parameters ≤ 5 per function
- Complexity ≤ 10 per function

@FEATURE:QUALITY-GUIDELINES Guideline compliance validation system
@DESIGN:ARCHITECTURE-001 Improved structure with constants and utilities

## Functions

### from_file

Load configuration from YAML or JSON file.

Args:
    config_path: Path to configuration file

Returns:
    GuidelineConfig instance

@DESIGN:CONFIG-LOADING-001 File-based configuration loading

```python
from_file(cls, config_path)
```

### to_file

Save configuration to YAML or JSON file.

Args:
    config_path: Path where to save configuration

```python
to_file(self, config_path)
```

### create_default

Create default configuration.

```python
create_default(cls)
```

### __init__

Initialize guideline checker with configurable limits and settings.

Args:
    project_path: Path to the project root directory
    limits: Custom guideline limits (uses defaults if not provided)
    config: Complete configuration object
    config_file: Path to configuration file (YAML/JSON)

@DESIGN:CONFIG-001 Support for configurable guideline limits
@DESIGN:CONFIG-003 Multiple configuration sources with priority

```python
__init__(self, project_path, limits, config, config_file)
```

### _parse_python_file

Parse Python file to AST with caching for performance.

Args:
    file_path: Path to Python file

Returns:
    AST object or None if parsing fails

@DESIGN:PERFORMANCE-001 Caching mechanism for AST parsing optimization

```python
_parse_python_file(self, file_path)
```

### _count_file_lines

Count lines in a file.

Args:
    file_path: Path to file

Returns:
    Number of lines in file

```python
_count_file_lines(self, file_path)
```

### _calculate_complexity

Calculate cyclomatic complexity of a function using enhanced algorithm.

Args:
    func_node: AST function node

Returns:
    Complexity score (1 + number of decision points)

@DESIGN:ALGORITHM-001 Enhanced complexity calculation with node type mapping

```python
_calculate_complexity(self, func_node)
```

### _discover_python_files

Discover Python files in project with intelligent filtering.

Returns:
    List of Python file paths excluding ignored directories

@DESIGN:FILE-DISCOVERY-001 Smart file discovery with exclusion patterns

```python
_discover_python_files(self)
```

### _is_minimal_init_file

Check if __init__.py file is minimal (< 10 lines, mostly imports/docstrings).

Args:
    file_path: Path to __init__.py file

Returns:
    bool: True if file is considered minimal

```python
_is_minimal_init_file(self, file_path)
```

### clear_cache

Clear the AST cache to free memory.

@DESIGN:PERFORMANCE-004 Cache management for memory optimization

```python
clear_cache(self)
```

### get_cache_stats

Get cache performance statistics.

Returns:
    Dict with cache hit/miss statistics and cache size

```python
get_cache_stats(self)
```

### check_function_length

Check if functions exceed 50 LOC limit.

Args:
    file_path: Path to Python file to check

Returns:
    List of violations with function name, line count, start line

Raises:
    GuidelineError: When functions exceed LOC limit

```python
check_function_length(self, file_path)
```

### check_file_size

Check if file exceeds 300 LOC limit.

Args:
    file_path: Path to Python file to check

Returns:
    Dict with file size info and violation status

Raises:
    GuidelineError: When file exceeds LOC limit

```python
check_file_size(self, file_path)
```

### check_parameter_count

Check if functions have more than 5 parameters.

Args:
    file_path: Path to Python file to check

Returns:
    List of violations with function name, parameter count, line number

Raises:
    GuidelineError: When functions exceed parameter limit

```python
check_parameter_count(self, file_path)
```

### check_complexity

Check if functions exceed complexity limit of 10.

Args:
    file_path: Path to Python file to check

Returns:
    List of violations with function name, complexity score, line number

Raises:
    GuidelineError: When functions exceed complexity limit

```python
check_complexity(self, file_path)
```

### scan_project

Scan entire project for guideline violations with optional parallel processing.

Args:
    parallel: Enable parallel processing for better performance
    max_workers: Maximum number of worker processes (defaults to CPU count)

Returns:
    Dict mapping violation types to list of violations

Raises:
    GuidelineError: When any guideline violations are found

@DESIGN:PERFORMANCE-002 Parallel processing for large codebases

```python
scan_project(self, parallel, max_workers)
```

### _scan_files_sequential

Scan files sequentially (original implementation).

Args:
    python_files: List of Python files to scan

Returns:
    Dict mapping violation types to list of violations

```python
_scan_files_sequential(self, python_files)
```

### _scan_files_parallel

Scan files in parallel for better performance on large projects.

Args:
    python_files: List of Python files to scan
    max_workers: Maximum number of worker processes

Returns:
    Dict mapping violation types to list of violations

@DESIGN:PERFORMANCE-003 Parallel file processing implementation

```python
_scan_files_parallel(self, python_files, max_workers)
```

### _scan_file_chunk

Scan a chunk of files (used by parallel processing).

Args:
    file_chunk: List of files to scan in this chunk

Returns:
    Dict mapping violation types to violations found in this chunk

```python
_scan_file_chunk(self, file_chunk)
```

### generate_violation_report

Generate comprehensive guideline violation report with performance metrics.

Args:
    parallel: Enable parallel processing for better performance

Returns:
    Dict containing violation summary, details, and performance metrics

@DESIGN:REPORT-001 Enhanced reporting with performance metrics

```python
generate_violation_report(self, parallel)
```

### _get_worst_violations

Identify the worst violations for prioritized fixing.

Args:
    violations: Violations found during scanning

Returns:
    Dict with worst violations by category

```python
_get_worst_violations(self, violations)
```

### export_config

Export current configuration to file.

Args:
    config_path: Path where to save configuration

@DESIGN:CONFIG-004 Configuration export for sharing and versioning

```python
export_config(self, config_path)
```

### update_config

Update configuration dynamically.

Args:
    **kwargs: Configuration parameters to update

Example:
    checker.update_config(max_function_lines=60, parallel_processing=False)

```python
update_config(self)
```

### get_enabled_checks

Get currently enabled checks.

Returns:
    Dict mapping check names to enabled status

```python
get_enabled_checks(self)
```

### set_enabled_checks

Set which checks are enabled.

Args:
    checks: Dict mapping check names to enabled status

```python
set_enabled_checks(self, checks)
```

### validate_single_file

Validate a single Python file against all guidelines.

Args:
    file_path: Path to Python file to validate

Returns:
    bool: True if file passes all guideline checks

Raises:
    GuidelineError: When file violates any guidelines

```python
validate_single_file(self, file_path)
```

## Classes

### GuidelineLimits

TRUST 5 principles guideline limits as immutable configuration.

### ProjectPatterns

File patterns and exclusions for project scanning.

### GuidelineConfig

Configuration for guideline checking with YAML/JSON support.

### GuidelineError

Guideline compliance violation exception.

### GuidelineChecker

Checks Python code compliance against TRUST 5 development guidelines.

@DESIGN:GUIDELINE-ARCH-001 Guideline validation architecture
@DESIGN:STRUCTURE-001 Improved architecture with configurable limits

Follows TRUST principles:
- T: Test-first validation with comprehensive coverage
- R: Readable code with clear naming and structure
- U: Unified design with separated concerns
- S: Secured with proper error handling and logging
- T: Trackable with detailed violation reporting

"""
Guideline compliance checker for MoAI-ADK.

Validates Python code against TRUST 5 principles development guidelines:
- Function length ≤ 50 LOC
- File size ≤ 300 LOC
- Parameters ≤ 5 per function
- Complexity ≤ 10 per function

@FEATURE:QUALITY-GUIDELINES Guideline compliance validation system
@DESIGN:ARCHITECTURE-001 Improved structure with constants and utilities
"""

import ast
import json
from concurrent.futures import ProcessPoolExecutor
from dataclasses import dataclass, asdict
from pathlib import Path
from typing import Dict, List, Optional, Any

# Optional YAML support (graceful degradation if not available)
try:
    import yaml
    YAML_AVAILABLE = True
except ImportError:
    YAML_AVAILABLE = False
    yaml = None

from ...utils.logger import get_logger

logger = get_logger(__name__)


# @DESIGN:CONSTANTS-001 TRUST 5 principles guideline limits
@dataclass(frozen=True)
class GuidelineLimits:
    """TRUST 5 principles guideline limits as immutable configuration."""
    MAX_FUNCTION_LINES: int = 50
    MAX_FILE_LINES: int = 300
    MAX_PARAMETERS: int = 5
    MAX_COMPLEXITY: int = 10

    # Additional quality thresholds
    MIN_DOCSTRING_LENGTH: int = 10
    MAX_NESTING_DEPTH: int = 4


# @DESIGN:CONSTANTS-002 File patterns for project scanning
class ProjectPatterns:
    """File patterns and exclusions for project scanning."""
    PYTHON_EXTENSION = '.py'
    EXCLUDED_DIRECTORIES = {'__pycache__', '.git', '.pytest_cache', 'venv', '.venv', 'node_modules'}
    EXCLUDED_FILES = {'__init__.py'}  # Optional: exclude minimal init files


# @DESIGN:CONSTANTS-003 AST node types for complexity calculation
COMPLEXITY_NODES = {
    # Control flow nodes that add complexity
    ast.If, ast.While, ast.For, ast.AsyncFor, ast.Try, ast.ExceptHandler,
    # Boolean operations and comparisons
    ast.BoolOp, ast.Compare
}


# @DESIGN:CONFIG-002 Configuration management for extensible guidelines
@dataclass
class GuidelineConfig:
    """Configuration for guideline checking with YAML/JSON support."""
    limits: GuidelineLimits
    file_patterns: Dict[str, Any]
    enabled_checks: Dict[str, bool]
    output_format: str = "json"
    parallel_processing: bool = True
    max_workers: Optional[int] = None

    @classmethod
    def from_file(cls, config_path: Path) -> 'GuidelineConfig':
        """
        Load configuration from YAML or JSON file.

        Args:
            config_path: Path to configuration file

        Returns:
            GuidelineConfig instance

        @DESIGN:CONFIG-LOADING-001 File-based configuration loading
        """
        if not config_path.exists():
            raise FileNotFoundError(f"Configuration file not found: {config_path}")

        try:
            with open(config_path, 'r', encoding='utf-8') as file:
                if config_path.suffix in ['.yaml', '.yml']:
                    if not YAML_AVAILABLE:
                        raise ImportError("PyYAML is required for YAML configuration files. "
                                        "Install with: pip install PyYAML")
                    config_data = yaml.safe_load(file)
                elif config_path.suffix == '.json':
                    config_data = json.load(file)
                else:
                    raise ValueError(f"Unsupported config format: {config_path.suffix}")

            # Extract and validate configuration sections
            limits_data = config_data.get('limits', {})
            limits = GuidelineLimits(
                MAX_FUNCTION_LINES=limits_data.get('max_function_lines', 50),
                MAX_FILE_LINES=limits_data.get('max_file_lines', 300),
                MAX_PARAMETERS=limits_data.get('max_parameters', 5),
                MAX_COMPLEXITY=limits_data.get('max_complexity', 10),
                MIN_DOCSTRING_LENGTH=limits_data.get('min_docstring_length', 10),
                MAX_NESTING_DEPTH=limits_data.get('max_nesting_depth', 4)
            )

            return cls(
                limits=limits,
                file_patterns=config_data.get('file_patterns', {}),
                enabled_checks=config_data.get('enabled_checks', {
                    'function_length': True,
                    'file_size': True,
                    'parameter_count': True,
                    'complexity': True
                }),
                output_format=config_data.get('output_format', 'json'),
                parallel_processing=config_data.get('parallel_processing', True),
                max_workers=config_data.get('max_workers')
            )

        except Exception as e:
            logger.error(f"Failed to load configuration from {config_path}: {e}")
            raise

    def to_file(self, config_path: Path) -> None:
        """
        Save configuration to YAML or JSON file.

        Args:
            config_path: Path where to save configuration
        """
        config_data = {
            'limits': asdict(self.limits),
            'file_patterns': self.file_patterns,
            'enabled_checks': self.enabled_checks,
            'output_format': self.output_format,
            'parallel_processing': self.parallel_processing,
            'max_workers': self.max_workers
        }

        try:
            with open(config_path, 'w', encoding='utf-8') as file:
                if config_path.suffix in ['.yaml', '.yml']:
                    if not YAML_AVAILABLE:
                        raise ImportError("PyYAML is required for YAML configuration files. "
                                        "Install with: pip install PyYAML")
                    yaml.safe_dump(config_data, file, default_flow_style=False, indent=2)
                elif config_path.suffix == '.json':
                    json.dump(config_data, file, indent=2)
                else:
                    raise ValueError(f"Unsupported config format: {config_path.suffix}")

            logger.info(f"Configuration saved to {config_path}")

        except Exception as e:
            logger.error(f"Failed to save configuration to {config_path}: {e}")
            raise

    @classmethod
    def create_default(cls) -> 'GuidelineConfig':
        """Create default configuration."""
        return cls(
            limits=GuidelineLimits(),
            file_patterns={
                'include': ['*.py'],
                'exclude': ['*test*', '*__pycache__*', '*.pyc']
            },
            enabled_checks={
                'function_length': True,
                'file_size': True,
                'parameter_count': True,
                'complexity': True
            }
        )


class GuidelineError(Exception):
    """Guideline compliance violation exception."""
    pass


class GuidelineChecker:
    """
    Checks Python code compliance against TRUST 5 development guidelines.

    @DESIGN:GUIDELINE-ARCH-001 Guideline validation architecture
    @DESIGN:STRUCTURE-001 Improved architecture with configurable limits

    Follows TRUST principles:
    - T: Test-first validation with comprehensive coverage
    - R: Readable code with clear naming and structure
    - U: Unified design with separated concerns
    - S: Secured with proper error handling and logging
    - T: Trackable with detailed violation reporting
    """

    def __init__(self,
                 project_path: Path,
                 limits: Optional[GuidelineLimits] = None,
                 config: Optional[GuidelineConfig] = None,
                 config_file: Optional[Path] = None):
        """
        Initialize guideline checker with configurable limits and settings.

        Args:
            project_path: Path to the project root directory
            limits: Custom guideline limits (uses defaults if not provided)
            config: Complete configuration object
            config_file: Path to configuration file (YAML/JSON)

        @DESIGN:CONFIG-001 Support for configurable guideline limits
        @DESIGN:CONFIG-003 Multiple configuration sources with priority
        """
        self.project_path = project_path

        # Configuration priority: config_file > config > limits > defaults
        if config_file and config_file.exists():
            self.config = GuidelineConfig.from_file(config_file)
            logger.info(f"Loaded configuration from {config_file}")
        elif config:
            self.config = config
        else:
            # Backward compatibility mode
            self.config = GuidelineConfig(
                limits=limits or GuidelineLimits(),
                file_patterns={'include': ['*.py'], 'exclude': []},
                enabled_checks={
                    'function_length': True,
                    'file_size': True,
                    'parameter_count': True,
                    'complexity': True
                }
            )

        self.limits = self.config.limits

        # Backward compatibility: expose limits as individual attributes
        self.max_function_lines = self.limits.MAX_FUNCTION_LINES
        self.max_file_lines = self.limits.MAX_FILE_LINES
        self.max_parameters = self.limits.MAX_PARAMETERS
        self.max_complexity = self.limits.MAX_COMPLEXITY

        # Cache for parsed AST trees to improve performance
        self._ast_cache: Dict[str, Optional[ast.AST]] = {}
        self._cache_stats = {"hits": 0, "misses": 0}

        logger.info(f"Initialized GuidelineChecker for {project_path} with config: "
                   f"func={self.limits.MAX_FUNCTION_LINES}, file={self.limits.MAX_FILE_LINES}, "
                   f"params={self.limits.MAX_PARAMETERS}, complexity={self.limits.MAX_COMPLEXITY}, "
                   f"parallel={self.config.parallel_processing}")

    def _parse_python_file(self, file_path: Path) -> Optional[ast.AST]:
        """
        Parse Python file to AST with caching for performance.

        Args:
            file_path: Path to Python file

        Returns:
            AST object or None if parsing fails

        @DESIGN:PERFORMANCE-001 Caching mechanism for AST parsing optimization
        """
        file_key = str(file_path.resolve())

        # Check cache first
        if file_key in self._ast_cache:
            self._cache_stats["hits"] += 1
            logger.debug(f"Cache hit for {file_path}")
            return self._ast_cache[file_key]

        self._cache_stats["misses"] += 1

        try:
            if not file_path.exists() or file_path.suffix != ProjectPatterns.PYTHON_EXTENSION:
                self._ast_cache[file_key] = None
                return None

            with open(file_path, 'r', encoding='utf-8') as file_handle:
                source_code = file_handle.read()

            parsed_ast = ast.parse(source_code, filename=str(file_path))
            self._ast_cache[file_key] = parsed_ast

            logger.debug(f"Successfully parsed and cached AST for: {file_path}")
            return parsed_ast

        except (SyntaxError, UnicodeDecodeError, FileNotFoundError) as parse_error:
            logger.warning(f"Failed to parse file {file_path}: {parse_error}")
            self._ast_cache[file_key] = None
            return None

    def _count_file_lines(self, file_path: Path) -> int:
        """
        Count lines in a file.

        Args:
            file_path: Path to file

        Returns:
            Number of lines in file
        """
        try:
            if not file_path.exists():
                return 0

            with open(file_path, 'r', encoding='utf-8') as f:
                return len(f.readlines())
        except (UnicodeDecodeError, FileNotFoundError):
            logger.warning(f"Failed to count lines in file: {file_path}")
            return 0

    def _calculate_complexity(self, func_node: ast.FunctionDef) -> int:
        """
        Calculate cyclomatic complexity of a function using enhanced algorithm.

        Args:
            func_node: AST function node

        Returns:
            Complexity score (1 + number of decision points)

        @DESIGN:ALGORITHM-001 Enhanced complexity calculation with node type mapping
        """
        base_complexity = 1  # Every function starts with complexity 1
        decision_points = 0

        for node in ast.walk(func_node):
            if type(node) in COMPLEXITY_NODES:
                decision_points += 1
                logger.debug(f"Found complexity node {type(node).__name__} at line {getattr(node, 'lineno', 'unknown')}")

        total_complexity = base_complexity + decision_points
        logger.debug(f"Function {func_node.name} has complexity {total_complexity} "
                    f"(base={base_complexity} + decisions={decision_points})")

        return total_complexity

    def _discover_python_files(self) -> List[Path]:
        """
        Discover Python files in project with intelligent filtering.

        Returns:
            List of Python file paths excluding ignored directories

        @DESIGN:FILE-DISCOVERY-001 Smart file discovery with exclusion patterns
        """
        python_files = []

        for file_path in self.project_path.rglob(f"*{ProjectPatterns.PYTHON_EXTENSION}"):
            # Check if file is in excluded directory
            path_parts = file_path.parts
            if any(excluded_dir in path_parts for excluded_dir in ProjectPatterns.EXCLUDED_DIRECTORIES):
                logger.debug(f"Skipping excluded file: {file_path}")
                continue

            # Optionally skip minimal init files
            if file_path.name in ProjectPatterns.EXCLUDED_FILES and self._is_minimal_init_file(file_path):
                logger.debug(f"Skipping minimal init file: {file_path}")
                continue

            python_files.append(file_path)

        logger.debug(f"Discovered {len(python_files)} Python files for analysis")
        return python_files

    def _is_minimal_init_file(self, file_path: Path) -> bool:
        """
        Check if __init__.py file is minimal (< 10 lines, mostly imports/docstrings).

        Args:
            file_path: Path to __init__.py file

        Returns:
            bool: True if file is considered minimal
        """
        if file_path.name != '__init__.py':
            return False

        try:
            line_count = self._count_file_lines(file_path)
            return line_count <= self.limits.MIN_DOCSTRING_LENGTH
        except Exception:
            return False

    def clear_cache(self) -> None:
        """
        Clear the AST cache to free memory.

        @DESIGN:PERFORMANCE-004 Cache management for memory optimization
        """
        cache_size = len(self._ast_cache)
        self._ast_cache.clear()
        self._cache_stats = {"hits": 0, "misses": 0}
        logger.info(f"Cleared AST cache (was {cache_size} entries)")

    def get_cache_stats(self) -> Dict[str, Any]:
        """
        Get cache performance statistics.

        Returns:
            Dict with cache hit/miss statistics and cache size
        """
        total_accesses = self._cache_stats["hits"] + self._cache_stats["misses"]
        hit_rate = (
            self._cache_stats["hits"] / total_accesses
            if total_accesses > 0
            else 0.0
        )

        return {
            "cache_hits": self._cache_stats["hits"],
            "cache_misses": self._cache_stats["misses"],
            "hit_rate": hit_rate,
            "cache_size": len(self._ast_cache)
        }

    def check_function_length(self, file_path: Path) -> List[Dict[str, Any]]:
        """
        Check if functions exceed 50 LOC limit.

        Args:
            file_path: Path to Python file to check

        Returns:
            List of violations with function name, line count, start line

        Raises:
            GuidelineError: When functions exceed LOC limit
        """
        violations = []
        tree = self._parse_python_file(file_path)

        if tree is None:
            return violations

        for node in ast.walk(tree):
            if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
                # Calculate function length (end_lineno - lineno + 1)
                func_length = node.end_lineno - node.lineno + 1

                if func_length > self.limits.MAX_FUNCTION_LINES:
                    violation = {
                        "function_name": node.name,
                        "line_count": func_length,
                        "start_line": node.lineno,
                        "file_path": str(file_path),
                        "max_allowed": self.limits.MAX_FUNCTION_LINES
                    }
                    violations.append(violation)

        return violations

    def check_file_size(self, file_path: Path) -> Dict[str, Any]:
        """
        Check if file exceeds 300 LOC limit.

        Args:
            file_path: Path to Python file to check

        Returns:
            Dict with file size info and violation status

        Raises:
            GuidelineError: When file exceeds LOC limit
        """
        line_count = self._count_file_lines(file_path)
        violation = line_count > self.limits.MAX_FILE_LINES

        result = {
            "file_path": str(file_path),
            "line_count": line_count,
            "violation": violation,
            "max_allowed": self.limits.MAX_FILE_LINES
        }

        return result

    def check_parameter_count(self, file_path: Path) -> List[Dict[str, Any]]:
        """
        Check if functions have more than 5 parameters.

        Args:
            file_path: Path to Python file to check

        Returns:
            List of violations with function name, parameter count, line number

        Raises:
            GuidelineError: When functions exceed parameter limit
        """
        violations = []
        tree = self._parse_python_file(file_path)

        if tree is None:
            return violations

        for node in ast.walk(tree):
            if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
                # Count parameters: args + kwonlyargs + vararg + kwarg
                param_count = (
                    len(node.args.args) +
                    len(node.args.kwonlyargs) +
                    (1 if node.args.vararg else 0) +
                    (1 if node.args.kwarg else 0)
                )

                if param_count > self.limits.MAX_PARAMETERS:
                    violation = {
                        "function_name": node.name,
                        "parameter_count": param_count,
                        "line_number": node.lineno,
                        "file_path": str(file_path),
                        "max_allowed": self.limits.MAX_PARAMETERS
                    }
                    violations.append(violation)

        return violations

    def check_complexity(self, file_path: Path) -> List[Dict[str, Any]]:
        """
        Check if functions exceed complexity limit of 10.

        Args:
            file_path: Path to Python file to check

        Returns:
            List of violations with function name, complexity score, line number

        Raises:
            GuidelineError: When functions exceed complexity limit
        """
        violations = []
        tree = self._parse_python_file(file_path)

        if tree is None:
            return violations

        for node in ast.walk(tree):
            if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
                complexity = self._calculate_complexity(node)

                if complexity > self.limits.MAX_COMPLEXITY:
                    violation = {
                        "function_name": node.name,
                        "complexity": complexity,
                        "line_number": node.lineno,
                        "file_path": str(file_path),
                        "max_allowed": self.limits.MAX_COMPLEXITY
                    }
                    violations.append(violation)

        return violations

    def scan_project(self, parallel: bool = True, max_workers: Optional[int] = None) -> Dict[str, List[Dict[str, Any]]]:
        """
        Scan entire project for guideline violations with optional parallel processing.

        Args:
            parallel: Enable parallel processing for better performance
            max_workers: Maximum number of worker processes (defaults to CPU count)

        Returns:
            Dict mapping violation types to list of violations

        Raises:
            GuidelineError: When any guideline violations are found

        @DESIGN:PERFORMANCE-002 Parallel processing for large codebases
        """
        violations = {
            "function_length": [],
            "file_size": [],
            "parameter_count": [],
            "complexity": []
        }

        # Find all Python files in project with smart filtering
        if not self.project_path.exists():
            return violations

        python_files = self._discover_python_files()
        logger.info(f"Scanning {len(python_files)} Python files for guideline violations")

        # Use configuration for parallel processing decision
        use_parallel = (self.config.parallel_processing if parallel else False) and len(python_files) > 5

        if use_parallel:
            violations = self._scan_files_parallel(python_files, max_workers or self.config.max_workers)
        else:
            violations = self._scan_files_sequential(python_files)

        return violations

    def _scan_files_sequential(self, python_files: List[Path]) -> Dict[str, List[Dict[str, Any]]]:
        """
        Scan files sequentially (original implementation).

        Args:
            python_files: List of Python files to scan

        Returns:
            Dict mapping violation types to list of violations
        """
        violations = {
            "function_length": [],
            "file_size": [],
            "parameter_count": [],
            "complexity": []
        }

        for file_path in python_files:
            try:
                # Check each guideline for this file (only if enabled)
                if self.config.enabled_checks.get('function_length', True):
                    violations["function_length"].extend(self.check_function_length(file_path))

                if self.config.enabled_checks.get('file_size', True):
                    file_result = self.check_file_size(file_path)
                    if file_result["violation"]:
                        violations["file_size"].append(file_result)

                if self.config.enabled_checks.get('parameter_count', True):
                    violations["parameter_count"].extend(self.check_parameter_count(file_path))

                if self.config.enabled_checks.get('complexity', True):
                    violations["complexity"].extend(self.check_complexity(file_path))

            except Exception as e:
                logger.warning(f"Error scanning file {file_path}: {e}")

        return violations

    def _scan_files_parallel(self, python_files: List[Path], max_workers: Optional[int]) -> Dict[str, List[Dict[str, Any]]]:
        """
        Scan files in parallel for better performance on large projects.

        Args:
            python_files: List of Python files to scan
            max_workers: Maximum number of worker processes

        Returns:
            Dict mapping violation types to list of violations

        @DESIGN:PERFORMANCE-003 Parallel file processing implementation
        """
        violations = {
            "function_length": [],
            "file_size": [],
            "parameter_count": [],
            "complexity": []
        }

        # Split files into chunks for better load balancing
        chunk_size = max(1, len(python_files) // (max_workers or 4))
        file_chunks = [python_files[i:i + chunk_size] for i in range(0, len(python_files), chunk_size)]

        try:
            with ProcessPoolExecutor(max_workers=max_workers) as executor:
                # Process chunks in parallel
                future_results = [
                    executor.submit(self._scan_file_chunk, chunk)
                    for chunk in file_chunks
                ]

                # Collect results from all workers
                for future in future_results:
                    try:
                        chunk_violations = future.result(timeout=30)  # 30 second timeout per chunk

                        # Merge results
                        for violation_type, violation_list in chunk_violations.items():
                            violations[violation_type].extend(violation_list)

                    except Exception as e:
                        logger.error(f"Error processing file chunk: {e}")

            logger.info(f"Parallel processing completed. Found {sum(len(v) for v in violations.values())} total violations")

        except Exception as e:
            logger.warning(f"Parallel processing failed, falling back to sequential: {e}")
            return self._scan_files_sequential(python_files)

        return violations

    def _scan_file_chunk(self, file_chunk: List[Path]) -> Dict[str, List[Dict[str, Any]]]:
        """
        Scan a chunk of files (used by parallel processing).

        Args:
            file_chunk: List of files to scan in this chunk

        Returns:
            Dict mapping violation types to violations found in this chunk
        """
        # Create a new checker instance for this worker (avoid shared state)
        worker_checker = GuidelineChecker(self.project_path, self.limits)

        return worker_checker._scan_files_sequential(file_chunk)

    def generate_violation_report(self, parallel: bool = True) -> Dict[str, Any]:
        """
        Generate comprehensive guideline violation report with performance metrics.

        Args:
            parallel: Enable parallel processing for better performance

        Returns:
            Dict containing violation summary, details, and performance metrics

        @DESIGN:REPORT-001 Enhanced reporting with performance metrics
        """
        import time
        start_time = time.time()

        violations = self.scan_project(parallel=parallel)

        # Calculate summary statistics
        total_violations = sum(len(v) for v in violations.values())
        files_with_violations = set()

        for violation_type, violation_list in violations.items():
            for violation in violation_list:
                if "file_path" in violation:
                    files_with_violations.add(violation["file_path"])

        scan_duration = time.time() - start_time
        python_files = self._discover_python_files() if self.project_path.exists() else []

        report = {
            "violations": violations,
            "summary": {
                "total_violations": total_violations,
                "files_checked": len(python_files),
                "files_with_violations": len(files_with_violations),
                "compliant": total_violations == 0,
                "compliance_rate": ((len(python_files) - len(files_with_violations)) / len(python_files) * 100)
                                  if len(python_files) > 0 else 100.0,
                "violation_breakdown": {
                    "function_length": len(violations["function_length"]),
                    "file_size": len(violations["file_size"]),
                    "parameter_count": len(violations["parameter_count"]),
                    "complexity": len(violations["complexity"])
                }
            },
            "performance": {
                "scan_duration_seconds": round(scan_duration, 3),
                "files_per_second": round(len(python_files) / scan_duration, 2) if scan_duration > 0 else 0,
                "parallel_processing": parallel and len(python_files) > 5,
                "cache_stats": self.get_cache_stats()
            },
            "trust_guidelines": {
                "limits": {
                    "max_function_lines": self.limits.MAX_FUNCTION_LINES,
                    "max_file_lines": self.limits.MAX_FILE_LINES,
                    "max_parameters": self.limits.MAX_PARAMETERS,
                    "max_complexity": self.limits.MAX_COMPLEXITY
                },
                "worst_violations": self._get_worst_violations(violations)
            }
        }

        logger.info(f"Generated violation report in {scan_duration:.3f}s: "
                   f"{total_violations} violations in {len(files_with_violations)} files")

        return report

    def _get_worst_violations(self, violations: Dict[str, List[Dict[str, Any]]]) -> Dict[str, Any]:
        """
        Identify the worst violations for prioritized fixing.

        Args:
            violations: Violations found during scanning

        Returns:
            Dict with worst violations by category
        """
        worst = {
            "longest_function": None,
            "largest_file": None,
            "most_parameters": None,
            "highest_complexity": None
        }

        # Find worst function length violation
        if violations["function_length"]:
            try:
                worst["longest_function"] = max(
                    violations["function_length"],
                    key=lambda x: x.get("line_count", 0)
                )
            except (KeyError, ValueError) as e:
                logger.warning(f"Error finding longest function: {e}")

        # Find worst file size violation
        if violations["file_size"]:
            try:
                worst["largest_file"] = max(
                    violations["file_size"],
                    key=lambda x: x.get("line_count", 0)
                )
            except (KeyError, ValueError) as e:
                logger.warning(f"Error finding largest file: {e}")

        # Find worst parameter count violation
        if violations["parameter_count"]:
            try:
                worst["most_parameters"] = max(
                    violations["parameter_count"],
                    key=lambda x: x.get("parameter_count", 0)
                )
            except (KeyError, ValueError) as e:
                logger.warning(f"Error finding most parameters: {e}")

        # Find worst complexity violation
        if violations["complexity"]:
            try:
                worst["highest_complexity"] = max(
                    violations["complexity"],
                    key=lambda x: x.get("complexity", 0)
                )
            except (KeyError, ValueError) as e:
                logger.warning(f"Error finding highest complexity: {e}")

        return worst

    def export_config(self, config_path: Path) -> None:
        """
        Export current configuration to file.

        Args:
            config_path: Path where to save configuration

        @DESIGN:CONFIG-004 Configuration export for sharing and versioning
        """
        self.config.to_file(config_path)

    def update_config(self, **kwargs) -> None:
        """
        Update configuration dynamically.

        Args:
            **kwargs: Configuration parameters to update

        Example:
            checker.update_config(max_function_lines=60, parallel_processing=False)
        """
        if 'max_function_lines' in kwargs:
            self.config.limits = GuidelineLimits(
                MAX_FUNCTION_LINES=kwargs['max_function_lines'],
                MAX_FILE_LINES=self.limits.MAX_FILE_LINES,
                MAX_PARAMETERS=self.limits.MAX_PARAMETERS,
                MAX_COMPLEXITY=self.limits.MAX_COMPLEXITY,
                MIN_DOCSTRING_LENGTH=self.limits.MIN_DOCSTRING_LENGTH,
                MAX_NESTING_DEPTH=self.limits.MAX_NESTING_DEPTH
            )
            self.limits = self.config.limits
            self.max_function_lines = self.limits.MAX_FUNCTION_LINES

        if 'parallel_processing' in kwargs:
            self.config.parallel_processing = kwargs['parallel_processing']

        # Add more parameter updates as needed
        logger.info(f"Configuration updated: {kwargs}")

    def get_enabled_checks(self) -> Dict[str, bool]:
        """
        Get currently enabled checks.

        Returns:
            Dict mapping check names to enabled status
        """
        return self.config.enabled_checks.copy()

    def set_enabled_checks(self, checks: Dict[str, bool]) -> None:
        """
        Set which checks are enabled.

        Args:
            checks: Dict mapping check names to enabled status
        """
        self.config.enabled_checks.update(checks)
        logger.info(f"Updated enabled checks: {checks}")

    def validate_single_file(self, file_path: Path) -> bool:
        """
        Validate a single Python file against all guidelines.

        Args:
            file_path: Path to Python file to validate

        Returns:
            bool: True if file passes all guideline checks

        Raises:
            GuidelineError: When file violates any guidelines
        """
        # Check all guideline types for this file
        function_violations = self.check_function_length(file_path)
        file_size_result = self.check_file_size(file_path)
        parameter_violations = self.check_parameter_count(file_path)
        complexity_violations = self.check_complexity(file_path)

        # Check if any violations exist
        has_function_violations = len(function_violations) > 0
        has_file_size_violations = file_size_result["violation"]
        has_parameter_violations = len(parameter_violations) > 0
        has_complexity_violations = len(complexity_violations) > 0

        # Return True if no violations found
        return not (
            has_function_violations or
            has_file_size_violations or
            has_parameter_violations or
            has_complexity_violations
        )

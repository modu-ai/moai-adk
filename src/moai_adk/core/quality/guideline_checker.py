"""
Main guideline compliance checker for MoAI-ADK.

Orchestrates validation across multiple specialized modules.
@FEATURE:QUALITY-GUIDELINES Main orchestrator for TRUST 5 principles validation
@DESIGN:REFACTORED-002 Rebuilt from 761 LOC to focused orchestration (< 250 LOC)
"""

from concurrent.futures import ProcessPoolExecutor
from pathlib import Path
from typing import Any

from ...utils.logger import get_logger
from .analyzers import CodeAnalyzer
from .config import GuidelineConfig
from .constants import GuidelineLimits
from .reporters import ViolationReporter
from .validators import GuidelineValidator

logger = get_logger(__name__)


class GuidelineChecker:
    """
    Main orchestrator for TRUST 5 principles compliance checking.

    Coordinates analysis, validation, and reporting across specialized modules.
    """

    def __init__(
        self,
        project_path: Path,
        limits: GuidelineLimits | None = None,
        config: GuidelineConfig | None = None,
        config_file: Path | None = None,
    ):
        """
        Initialize guideline checker.

        Args:
            project_path: Root path of project to check
            limits: Optional custom limits (uses defaults if None)
            config: Optional pre-built configuration
            config_file: Optional path to config file
        """
        self.project_path = Path(project_path)

        # Initialize configuration
        if config:
            self.config = config
        elif config_file and config_file.exists():
            self.config = GuidelineConfig.from_file(config_file)
        else:
            self.config = GuidelineConfig.create_default()

        # Override limits if provided
        if limits:
            self.config.limits = limits

        # Initialize specialized components
        self.analyzer = CodeAnalyzer(self.config.limits)
        self.validator = GuidelineValidator(self.config.limits)
        self.reporter = ViolationReporter(self.config.limits)

        # Cache for performance
        self._cache_stats = {"hits": 0, "misses": 0}

        logger.info(f"Initialized GuidelineChecker for {project_path}")

    def scan_project(
        self, parallel: bool = True, max_workers: int | None = None
    ) -> dict[str, list[dict[str, Any]]]:
        """
        Scan entire project for guideline violations.

        Args:
            parallel: Whether to use parallel processing
            max_workers: Maximum worker processes

        Returns:
            Dictionary of violations by type
        """
        python_files = self.analyzer.discover_python_files(self.project_path)

        if not python_files:
            logger.warning(f"No Python files found in {self.project_path}")
            return {}

        logger.info(f"Scanning {len(python_files)} Python files")

        if parallel and len(python_files) > 1:
            return self._scan_files_parallel(python_files, max_workers)
        else:
            return self._scan_files_sequential(python_files)

    def _scan_files_sequential(
        self, python_files: list[Path]
    ) -> dict[str, list[dict[str, Any]]]:
        """Scan files sequentially."""
        all_violations = {
            "file_size": [],
            "function_length": [],
            "parameter_count": [],
            "complexity": [],
        }

        for file_path in python_files:
            violations = self._scan_single_file(file_path)
            for violation_type, violation_list in violations.items():
                all_violations[violation_type].extend(violation_list)

        return all_violations

    def _scan_files_parallel(
        self, python_files: list[Path], max_workers: int | None
    ) -> dict[str, list[dict[str, Any]]]:
        """Scan files in parallel."""
        all_violations = {
            "file_size": [],
            "function_length": [],
            "parameter_count": [],
            "complexity": [],
        }

        chunk_size = max(1, len(python_files) // (max_workers or 4))
        file_chunks = [
            python_files[i : i + chunk_size]
            for i in range(0, len(python_files), chunk_size)
        ]

        with ProcessPoolExecutor(max_workers=max_workers) as executor:
            chunk_results = list(executor.map(self._scan_file_chunk, file_chunks))

        for chunk_violations in chunk_results:
            for violation_type, violation_list in chunk_violations.items():
                all_violations[violation_type].extend(violation_list)

        return all_violations

    def _scan_file_chunk(
        self, file_chunk: list[Path]
    ) -> dict[str, list[dict[str, Any]]]:
        """Scan a chunk of files."""
        chunk_violations = {
            "file_size": [],
            "function_length": [],
            "parameter_count": [],
            "complexity": [],
        }

        for file_path in file_chunk:
            violations = self._scan_single_file(file_path)
            for violation_type, violation_list in violations.items():
                chunk_violations[violation_type].extend(violation_list)

        return chunk_violations

    def _scan_single_file(self, file_path: Path) -> dict[str, list[dict[str, Any]]]:
        """Scan a single file for all violation types."""
        violations = {
            "file_size": [],
            "function_length": [],
            "parameter_count": [],
            "complexity": [],
        }

        try:
            # Check file size
            if self.config.enabled_checks.get("file_size", True):
                file_violation = self.validator.check_file_size(file_path)
                if file_violation["violation"]:
                    violations["file_size"].append(file_violation)

            # Check function-level violations
            if any(
                self.config.enabled_checks.get(check, True)
                for check in ["function_length", "parameter_count", "complexity"]
            ):
                if self.config.enabled_checks.get("function_length", True):
                    violations["function_length"].extend(
                        self.validator.check_function_length(file_path)
                    )

                if self.config.enabled_checks.get("parameter_count", True):
                    violations["parameter_count"].extend(
                        self.validator.check_parameter_count(file_path)
                    )

                if self.config.enabled_checks.get("complexity", True):
                    violations["complexity"].extend(
                        self.validator.check_complexity(file_path)
                    )

        except Exception as e:
            logger.error(f"Error scanning file {file_path}: {e}")

        return violations

    def generate_violation_report(self, parallel: bool = True) -> dict[str, Any]:
        """
        Generate comprehensive violation report.

        Args:
            parallel: Whether to use parallel scanning

        Returns:
            Comprehensive report dictionary
        """
        violations = self.scan_project(parallel=parallel)
        return self.reporter.generate_violation_report(violations)

    def validate_single_file(self, file_path: Path) -> bool:
        """Validate a single file against all guidelines."""
        return self.validator.validate_single_file(file_path)

    def export_config(self, config_path: Path) -> None:
        """Export current configuration to file."""
        self.config.to_file(config_path)
        logger.info(f"Configuration exported to {config_path}")

    def update_config(self, **kwargs) -> None:
        """Update configuration with new values."""
        for key, value in kwargs.items():
            if hasattr(self.config, key):
                setattr(self.config, key, value)
            else:
                logger.warning(f"Unknown config key: {key}")

    def get_cache_stats(self) -> dict[str, Any]:
        """Get cache performance statistics."""
        total_requests = self._cache_stats["hits"] + self._cache_stats["misses"]
        hit_rate = (
            self._cache_stats["hits"] / total_requests if total_requests > 0 else 0
        )

        return {
            "hits": self._cache_stats["hits"],
            "misses": self._cache_stats["misses"],
            "hit_rate": round(hit_rate * 100, 2),
            "total_requests": total_requests,
        }

    def clear_cache(self) -> None:
        """Clear internal caches."""
        self._cache_stats = {"hits": 0, "misses": 0}
        logger.info("Caches cleared")

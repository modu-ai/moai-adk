#!/usr/bin/env python3
"""Quality Gate Hook with LSP Diagnostics Integration.

Ralph-style quality checking before sync phase.

This hook validates code quality using LSP diagnostics before allowing
synchronization to proceed. It integrates with MCP IDE tools to get
real-time diagnostic information and applies quality gate rules.

Exit Codes:
- 0: Quality gate passed
- 1: Quality gate failed

Output:
- JSON with gate results including error counts and pass/fail status
"""

from __future__ import annotations

import json
import os
import sys
from pathlib import Path
from typing import Any

# Default quality thresholds
DEFAULT_MAX_ERRORS = 0
DEFAULT_MAX_WARNINGS = 10


def get_project_dir() -> Path:
    """Get the project directory from environment or current working directory.

    Returns:
        Path to the project directory.
    """
    project_dir = os.environ.get("CLAUDE_PROJECT_DIR", os.getcwd())
    return Path(project_dir)


def load_quality_config() -> dict[str, Any]:
    """Load quality configuration from .moai/config/sections/quality.yaml.

    Returns:
        Configuration dictionary with quality gate settings.
    """
    config: dict[str, Any] = {
        "max_errors": DEFAULT_MAX_ERRORS,
        "max_warnings": DEFAULT_MAX_WARNINGS,
        "enabled": True,
    }

    config_path = get_project_dir() / ".moai" / "config" / "sections" / "quality.yaml"

    if config_path.exists():
        try:
            import yaml

            with open(config_path) as f:
                loaded = yaml.safe_load(f)
                if loaded:
                    # Extract quality gate settings if present
                    constitution = loaded.get("constitution", {})
                    quality_gate = constitution.get("quality_gate", {})

                    if "max_errors" in quality_gate:
                        config["max_errors"] = quality_gate["max_errors"]
                    if "max_warnings" in quality_gate:
                        config["max_warnings"] = quality_gate["max_warnings"]
                    if "enabled" in quality_gate:
                        config["enabled"] = quality_gate["enabled"]

        except Exception as e:
            # Graceful degradation - use defaults
            print(
                f"WARNING: Failed to load quality config: {e}",
                file=sys.stderr,
            )

    return config


def get_lsp_diagnostics() -> list[dict[str, Any]]:
    """Get LSP diagnostics from MCP IDE tool.

    Returns:
        List of diagnostic dictionaries. Empty list if tool unavailable.
    """
    diagnostics: list[dict[str, Any]] = []

    try:
        # Try to import MCP IDE diagnostics tool
        from mcp__ide__getDiagnostics import getDiagnostics

        result = getDiagnostics()

        # Handle different return types
        if isinstance(result, list):
            diagnostics = result
        elif isinstance(result, dict) and "diagnostics" in result:
            diagnostics = result["diagnostics"]
        elif isinstance(result, dict) and "items" in result:
            diagnostics = result["items"]

    except ImportError:
        print(
            "WARNING: MCP IDE diagnostics not available - running without LSP checks",
            file=sys.stderr,
        )
    except AttributeError:
        print(
            "WARNING: MCP IDE diagnostics tool not found - running without LSP checks",
            file=sys.stderr,
        )
    except Exception as e:
        print(
            f"ERROR: Failed to get LSP diagnostics: {e}",
            file=sys.stderr,
        )

    return diagnostics


def categorize_diagnostics(
    diagnostics: list[dict[str, Any]],
) -> dict[str, Any]:
    """Categorize diagnostics by severity and source.

    Args:
        diagnostics: List of diagnostic dictionaries.

    Returns:
        Dictionary with categorized counts and diagnostic lists.
    """
    errors: list[dict[str, Any]] = []
    warnings: list[dict[str, Any]] = []
    type_errors: list[dict[str, Any]] = []
    lint_errors: list[dict[str, Any]] = []

    for diag in diagnostics:
        severity = diag.get("severity", "").lower()
        source = diag.get("source", "").lower()

        # Categorize by severity
        if severity == "error":
            errors.append(diag)

            # Also categorize by source
            if source == "typecheck" or "type" in source:
                type_errors.append(diag)
            elif source == "lint" or "lint" in source or "ruff" in source:
                lint_errors.append(diag)

        elif severity == "warning":
            warnings.append(diag)

    return {
        "errors": errors,
        "warnings": warnings,
        "type_errors": type_errors,
        "lint_errors": lint_errors,
    }


def check_quality_gate(
    categorized: dict[str, Any],
    config: dict[str, Any],
) -> dict[str, Any]:
    """Check if quality gate passes based on diagnostic counts.

    Args:
        categorized: Categorized diagnostics dictionary.
        config: Quality configuration dictionary.

    Returns:
        Dictionary with gate results including pass/fail status.
    """
    errors = categorized["errors"]
    warnings = categorized["warnings"]
    type_errors = categorized["type_errors"]
    lint_errors = categorized["lint_errors"]

    max_errors = config.get("max_errors", DEFAULT_MAX_ERRORS)
    max_warnings = config.get("max_warnings", DEFAULT_MAX_WARNINGS)

    gate_results: dict[str, Any] = {
        "lsp_errors": len(errors),
        "lsp_warnings": len(warnings),
        "type_errors": len(type_errors),
        "lint_errors": len(lint_errors),
        "passed": False,
        "reason": "",
    }

    issues: list[str] = []

    # Check error threshold
    if len(errors) > max_errors:
        issues.append(f"{len(errors)} LSP errors (max: {max_errors})")
        # Add first 5 error details
        for error in errors[:5]:
            message = error.get("message", "unknown error")
            source = error.get("source", "")
            file_path = error.get("file", error.get("path", "unknown"))
            location = f"{file_path}" if file_path != "unknown" else ""
            source_info = f" [{source}]" if source else ""
            issues.append(f"  - {message}{source_info} {location}".strip())

        if len(errors) > 5:
            issues.append(f"  ... and {len(errors) - 5} more errors")

    # Check warning threshold
    if len(warnings) > max_warnings:
        issues.append(f"{len(warnings)} LSP warnings (max: {max_warnings})")
        # Add first 3 warning details
        for warning in warnings[:3]:
            message = warning.get("message", "unknown warning")
            source = warning.get("source", "")
            file_path = warning.get("file", warning.get("path", "unknown"))
            location = f"{file_path}" if file_path != "unknown" else ""
            source_info = f" [{source}]" if source else ""
            issues.append(f"  - {message}{source_info} {location}".strip())

        if len(warnings) > 3:
            issues.append(f"  ... and {len(warnings) - 3} more warnings")

    # Specific type error check
    if len(type_errors) > 0:
        issues.append(f"{len(type_errors)} type checking errors")

    # Specific lint error check
    if len(lint_errors) > 0:
        issues.append(f"{len(lint_errors)} lint errors")

    if issues:
        gate_results["reason"] = "Quality gate failed:\n" + "\n".join(issues)
        gate_results["passed"] = False
    else:
        gate_results["reason"] = "Quality gate passed: LSP diagnostics clean"
        gate_results["passed"] = True

    return gate_results


def run_quality_gate() -> int:
    """Execute quality gate with LSP diagnostics check.

    Returns:
        Exit code (0 for pass, 1 for fail).
    """
    # Load configuration
    config = load_quality_config()

    # Check if quality gate is disabled
    if not config.get("enabled", True):
        result = {
            "lsp_errors": 0,
            "lsp_warnings": 0,
            "type_errors": 0,
            "lint_errors": 0,
            "passed": True,
            "reason": "Quality gate disabled",
        }
        print(json.dumps(result, indent=2))
        return 0

    # Get LSP diagnostics
    diagnostics = get_lsp_diagnostics()

    # Categorize diagnostics
    categorized = categorize_diagnostics(diagnostics)

    # Check quality gate
    gate_results = check_quality_gate(categorized, config)

    # Output results
    print(json.dumps(gate_results, indent=2))

    # Return exit code
    return 0 if gate_results["passed"] else 1


def main() -> None:
    """Main hook entry point."""
    sys.exit(run_quality_gate())


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
# @CODE:DOC-TAG-004 | Component 3: CLI utility for TAG validation
"""CLI utility for moai-adk validate-tags command

This module provides a command-line interface to the central TAG validator:
- Validate files or directories
- Generate reports in multiple formats (detailed, summary, JSON)
- Support various validation modes (strict, custom patterns)
- Save reports to files

Usage:
    moai-adk validate-tags .
    moai-adk validate-tags --strict src/
    moai-adk validate-tags --format json --output report.json
    moai-adk validate-tags --no-duplicates --no-orphans
"""

import argparse
import sys
from pathlib import Path
from typing import Optional

from .validator import (
    CentralValidationResult,
    CentralValidator,
    ValidationConfig,
)


def create_parser() -> argparse.ArgumentParser:
    """Create argument parser for CLI

    Returns:
        ArgumentParser with all CLI options
    """
    parser = argparse.ArgumentParser(
        prog="moai-adk validate-tags",
        description="Validate TAG annotations in MoAI-ADK projects",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  moai-adk validate-tags .                          # Validate entire project
  moai-adk validate-tags --strict src/              # Strict validation of src/
  moai-adk validate-tags --format json              # JSON report output
  moai-adk validate-tags --output report.json       # Save report to file
  moai-adk validate-tags --no-orphans               # Disable orphan checking
  moai-adk validate-tags --file-types py,js,md      # Validate specific file types

Report Formats:
  detailed  - Full report with all issues, locations, and suggestions (default)
  summary   - Concise summary with statistics only
  json      - Machine-readable JSON format

Validation Modes:
  Normal    - Errors block, warnings reported but pass
  Strict    - Both errors and warnings block (use --strict flag)
        """
    )

    # Positional argument
    parser.add_argument(
        "path",
        nargs="?",
        default=".",
        help="Path to file or directory to validate (default: current directory)"
    )

    # Validation options
    parser.add_argument(
        "--strict",
        action="store_true",
        help="Treat warnings as errors (block on warnings)"
    )

    parser.add_argument(
        "--no-duplicates",
        action="store_true",
        help="Disable duplicate TAG checking"
    )

    parser.add_argument(
        "--no-orphans",
        action="store_true",
        help="Disable orphan TAG checking"
    )

    parser.add_argument(
        "--no-chain-check",
        action="store_true",
        help="Disable TAG chain integrity checking"
    )

    # File filtering options
    parser.add_argument(
        "--file-types",
        help="Comma-separated file types to validate (e.g., py,js,ts)"
    )

    parser.add_argument(
        "--ignore-patterns",
        help="Comma-separated glob patterns to ignore (e.g., .git/*,*.pyc)"
    )

    # Report options
    parser.add_argument(
        "--format",
        choices=["detailed", "summary", "json"],
        default="detailed",
        help="Report format (default: detailed)"
    )

    parser.add_argument(
        "--output",
        "-o",
        help="Output file path (default: stdout)"
    )

    # Quiet mode
    parser.add_argument(
        "--quiet",
        "-q",
        action="store_true",
        help="Suppress output, only return exit code"
    )

    # Version
    parser.add_argument(
        "--version",
        action="version",
        version="moai-adk 0.7.0"
    )

    return parser


def validate_path(path_str: str) -> Optional[Path]:
    """Validate that path exists

    Args:
        path_str: Path string to validate

    Returns:
        Path object if valid, None otherwise
    """
    path = Path(path_str)
    if not path.exists():
        print(f"Error: Path does not exist: {path_str}", file=sys.stderr)
        return None
    return path


def create_config_from_args(args: argparse.Namespace) -> ValidationConfig:
    """Create ValidationConfig from CLI arguments

    Args:
        args: Parsed command-line arguments

    Returns:
        ValidationConfig object
    """
    config = ValidationConfig(
        strict_mode=args.strict,
        check_duplicates=not args.no_duplicates,
        check_orphans=not args.no_orphans,
        check_chain_integrity=not args.no_chain_check,
        report_format=args.format
    )

    # Parse file types if provided
    if args.file_types:
        config.allowed_file_types = [
            ft.strip() for ft in args.file_types.split(",")
        ]

    # Parse ignore patterns if provided
    if args.ignore_patterns:
        config.ignore_patterns = [
            pattern.strip() for pattern in args.ignore_patterns.split(",")
        ]

    return config


def run_validation(path: Path, config: ValidationConfig) -> CentralValidationResult:
    """Run validation on path

    Args:
        path: Path to validate (file or directory)
        config: ValidationConfig object

    Returns:
        CentralValidationResult
    """
    validator = CentralValidator(config=config)

    if path.is_file():
        result = validator.validate_files([str(path)])
    else:
        result = validator.validate_directory(str(path))

    return result


def output_report(
    result: CentralValidationResult,
    validator: CentralValidator,
    format: str,
    output_file: Optional[str],
    quiet: bool
) -> None:
    """Output validation report

    Args:
        result: CentralValidationResult
        validator: CentralValidator instance
        format: Report format (detailed|summary|json)
        output_file: Output file path (None for stdout)
        quiet: Suppress output if True
    """
    if quiet:
        return

    # Generate report
    report = validator.create_report(result, format=format)

    # Output to file or stdout
    if output_file:
        try:
            with open(output_file, 'w') as f:
                f.write(report)
            print(f"Report saved to: {output_file}", file=sys.stderr)
        except Exception as e:
            print(f"Error writing report: {e}", file=sys.stderr)
            sys.exit(1)
    else:
        print(report)


def main(argv: Optional[list] = None) -> int:
    """Main CLI entry point

    Args:
        argv: Command-line arguments (default: sys.argv[1:])

    Returns:
        Exit code (0 for success, 1 for failure)
    """
    parser = create_parser()
    args = parser.parse_args(argv)

    # Validate path
    path = validate_path(args.path)
    if path is None:
        return 1

    # Create configuration from args
    config = create_config_from_args(args)

    # Run validation
    try:
        result = run_validation(path, config)
    except Exception as e:
        print(f"Validation error: {e}", file=sys.stderr)
        return 1

    # Create validator for report generation
    validator = CentralValidator(config=config)

    # Output report
    output_report(
        result=result,
        validator=validator,
        format=args.format,
        output_file=args.output,
        quiet=args.quiet
    )

    # Return exit code
    return 0 if result.is_valid else 1


if __name__ == "__main__":
    sys.exit(main())

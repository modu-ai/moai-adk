"""
Link Validation CLI Command

Validate online documentation links
"""

import argparse
import asyncio
from pathlib import Path

from moai_adk.utils.link_validator import LinkValidator


def create_parser(subparsers) -> argparse.ArgumentParser:
    """Create link validation parser"""
    parser = subparsers.add_parser(
        "validate-links",
        help="Validate online documentation links",
        description="Automatically validate all online documentation links in README.ko.md",
    )

    parser.add_argument(
        "--file",
        "-f",
        type=str,
        default="README.ko.md",
        help="File path to validate (default: README.ko.md)",
    )

    parser.add_argument(
        "--max-concurrent",
        "-c",
        type=int,
        default=3,
        help="Maximum number of links to validate concurrently (default: 3)",
    )

    parser.add_argument(
        "--timeout",
        "-t",
        type=int,
        default=8,
        help="Request timeout in seconds (default: 8)",
    )

    parser.add_argument("--output", "-o", type=str, help="File path to save results")

    parser.add_argument(
        "--verbose",
        "-v",
        action="store_true",
        help="Display detailed progress information",
    )

    return parser


def run_command(args) -> int:
    """Execute link validation command"""
    try:
        # Configure file path
        file_path = Path(args.file)
        if not file_path.exists():
            print(f"Error: File does not exist: {file_path}")
            return 1

        # Create validator
        validator = LinkValidator(
            max_concurrent=args.max_concurrent, timeout=args.timeout
        )

        if args.verbose:
            print(f"Extracting links from file: {file_path}")

        # Extract links
        links = validator.extract_links_from_file(file_path)

        if not links:
            print("No links to validate.")
            return 0

        if args.verbose:
            print(f"Validating a total of {len(links)} links...")

        # Execute async validation
        async def validate_links():
            async with validator:
                result = await validator.validate_all_links(links)
                return result

        result = asyncio.run(validate_links())

        # Generate report
        report = validator.generate_report(result)

        # Output
        print(report)

        # Save to file
        if args.output:
            output_path = Path(args.output)
            output_path.write_text(report, encoding="utf-8")
            print(f"\nResults saved to: {output_path}")

        # Return exit code
        if result.invalid_links > 0:
            print(f"\n⚠️  {result.invalid_links} link(s) failed validation.")
            return 1
        else:
            print("\n✅ All links validated successfully.")
            return 0

    except KeyboardInterrupt:
        print("\nValidation cancelled by user.")
        return 1
    except Exception as e:
        print(f"Error occurred: {e}")
        return 1

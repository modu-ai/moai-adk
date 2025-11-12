"""Slash command diagnostics

Diagnose and validate Claude Code slash command loading issues.

Functions:
- validate_command_file: Validate YAML front matter and required fields
- scan_command_files: Recursively scan for .md files
- diagnose_slash_commands: Comprehensive diagnostic report
"""

from pathlib import Path

import yaml  # type: ignore[import-untyped]


def validate_command_file(file_path: Path) -> dict:
    """Validate slash command file format

    Checks:
    1. File exists and readable
    2. YAML front matter present (starts with ---)
    3. Valid YAML syntax
    4. Required fields: name, description

    Args:
        file_path: Path to command file (.md)

    Returns:
        dict with 'valid' (bool) and optional 'errors' (list[str])

    Example:
        >>> result = validate_command_file(Path("cmd.md"))
        >>> if result["valid"]:
        ...     print("Valid command file")
        ... else:
        ...     print(f"Errors: {result['errors']}")
    """
    try:
        # Check file exists
        if not file_path.exists():
            return {"valid": False, "errors": ["File not found"]}

        # Read file content
        content = file_path.read_text(encoding="utf-8")

        # Check front matter delimiter
        if not content.startswith("---"):
            return {
                "valid": False,
                "errors": ["Missing YAML front matter (must start with ---)"],
            }

        # Split by --- delimiter
        parts = content.split("---", 2)
        if len(parts) < 3:
            return {
                "valid": False,
                "errors": ["Invalid front matter format (missing closing ---)"],
            }

        # Parse YAML
        try:
            metadata = yaml.safe_load(parts[1])
        except yaml.YAMLError as e:
            return {"valid": False, "errors": [f"YAML parsing error: {e}"]}

        # Check required fields
        required_fields = ["name", "description"]
        missing_fields = [field for field in required_fields if field not in metadata]

        if missing_fields:
            return {
                "valid": False,
                "errors": [f"Missing required field: {', '.join(missing_fields)}"],
            }

        return {"valid": True}

    except Exception as e:
        return {"valid": False, "errors": [str(e)]}


def scan_command_files(commands_dir: Path) -> list[Path]:
    """Scan directory for all .md files

    Recursively searches for .md files in the given directory.

    Args:
        commands_dir: Directory to scan (e.g., .claude/commands)

    Returns:
        List of Path objects for found .md files

    Example:
        >>> files = scan_command_files(Path(".claude/commands"))
        >>> print(f"Found {len(files)} command files")
    """
    if not commands_dir.exists():
        return []

    try:
        return list(commands_dir.glob("**/*.md"))
    except Exception:
        return []


def diagnose_slash_commands() -> dict:
    """Diagnose slash command loading issues

    Comprehensive diagnostic that:
    1. Checks if .claude/commands directory exists
    2. Scans for all .md files
    3. Validates each file's format
    4. Returns detailed report

    Returns:
        dict with diagnostic results:
        - total_files: Number of .md files found
        - valid_commands: Number of valid command files
        - details: List of per-file validation results
        OR
        - error: Error message if directory not found

    Example:
        >>> result = diagnose_slash_commands()
        >>> if "error" in result:
        ...     print(f"Error: {result['error']}")
        ... else:
        ...     print(f"{result['valid_commands']}/{result['total_files']} valid")
    """
    commands_dir = Path(".claude/commands")

    # Check if directory exists
    if not commands_dir.exists():
        return {"error": "Commands directory not found"}

    # Scan for .md files
    md_files = scan_command_files(commands_dir)

    # Validate each file
    details = []
    for file_path in md_files:
        validation = validate_command_file(file_path)
        details.append(
            {
                "file": str(file_path.relative_to(commands_dir)),
                "valid": validation["valid"],
                "errors": validation.get("errors", []),
            }
        )

    # Count valid commands
    valid_count = sum(1 for detail in details if detail["valid"])

    return {
        "total_files": len(md_files),
        "valid_commands": valid_count,
        "details": details,
    }

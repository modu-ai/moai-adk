# @CODE:GEN-001, @CODE:GEN-002
"""TAG ID generation and duplicate detection.

Generates sequential @DOC:DOMAIN-NNN identifiers and detects duplicates
using ripgrep for performance.

@SPEC:DOC-TAG-001: @DOC tag automatic generation infrastructure
"""

import re
import subprocess
from typing import List

# Domain validation: Must start with letter, alphanumeric + hyphens, end with alphanumeric
DOMAIN_PATTERN = re.compile(r"^[A-Z]([A-Z0-9-]*[A-Z0-9])?$")


def generate_doc_tag(domain: str, existing_ids: List[str]) -> str:
    """Generate next available @DOC tag ID for domain.

    Validates domain format (uppercase alphanumeric, hyphens allowed).
    Finds highest current number and increments by 1.

    Args:
        domain: Domain name (e.g., "AUTH", "CLI-TOOL")
        existing_ids: List of existing TAG IDs for this domain

    Returns:
        Next TAG ID (e.g., "@DOC:AUTH-003")

    Raises:
        ValueError: Invalid domain format

    Examples:
        >>> generate_doc_tag("AUTH", [])
        '@DOC:AUTH-001'
        >>> generate_doc_tag("AUTH", ["@DOC:AUTH-001", "@DOC:AUTH-002"])
        '@DOC:AUTH-003'
    """
    # Validate domain format: uppercase alphanumeric and hyphens only
    if not DOMAIN_PATTERN.match(domain):
        raise ValueError(
            f"Invalid domain format: {domain}. "
            "Domain must be uppercase, start with a letter, "
            "and contain only alphanumeric characters and hyphens."
        )

    # Find max number from existing IDs for this domain
    max_num = 0
    pattern = rf"@DOC:{re.escape(domain)}-(\d{{3}})$"

    for tag_id in existing_ids:
        match = re.search(pattern, tag_id)
        if match:
            num = int(match.group(1))
            max_num = max(max_num, num)

    # Generate next ID
    next_num = max_num + 1
    return f"@DOC:{domain}-{next_num:03d}"


def detect_duplicates(domain: str, search_path: str = "docs/") -> List[str]:
    """Detect existing @DOC tags using ripgrep.

    Performs efficient ripgrep search for all @DOC:DOMAIN-NNN tags.

    Args:
        domain: Domain to search for
        search_path: Directory to search (default: docs/)

    Returns:
        List of existing TAG IDs

    Raises:
        RuntimeError: ripgrep not available or execution error

    Examples:
        >>> detect_duplicates("AUTH", "docs/")
        ['@DOC:AUTH-001', '@DOC:AUTH-002']
    """
    try:
        result = subprocess.run(
            ["rg", rf"@DOC:{re.escape(domain)}-\d{{3}}", search_path, "-o", "--no-heading"],
            capture_output=True,
            text=True,
            timeout=5,
        )

        if result.returncode == 0:
            # Found matches
            return [tag.strip() for tag in result.stdout.strip().split("\n") if tag.strip()]
        elif result.returncode == 1:
            # No matches found (normal case for new domains)
            return []
        else:
            # ripgrep error
            raise RuntimeError(f"ripgrep error: {result.stderr}")

    except FileNotFoundError:
        raise RuntimeError(
            "ripgrep (rg) not found in PATH. Please install it: "
            "brew install ripgrep (macOS) or apt install ripgrep (Linux)"
        )
    except subprocess.TimeoutExpired:
        raise RuntimeError(
            f"ripgrep timeout after 5 seconds searching {search_path}. "
            "Consider narrowing the search path."
        )

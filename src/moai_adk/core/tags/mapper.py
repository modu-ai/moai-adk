# @CODE:MAP-001, @CODE:MAP-002
"""SPEC-DOC mapping and confidence scoring.

Maps documentation files to related SPEC IDs based on domain matching
and calculates confidence scores for chain references.

@SPEC:DOC-TAG-001: @DOC tag automatic generation infrastructure
"""

import re
from pathlib import Path
from typing import Any, Optional

from moai_adk.core.tags.parser import extract_spec_id, parse_domain


def find_related_spec(doc_path: Path, specs_dir: Path = Path(".moai/specs")) -> Optional[str]:
    """Find related SPEC ID by matching domain from document path.

    Searches for SPEC directories matching the domain inferred from the
    document's file path. Returns the most recent SPEC (highest number)
    if multiple matches exist.

    Args:
        doc_path: Path to documentation file (e.g., docs/auth/setup.md)
        specs_dir: Path to SPEC directory (default: .moai/specs)

    Returns:
        SPEC ID (e.g., "AUTH-001") or None if no match found

    Examples:
        >>> find_related_spec(Path("docs/auth/guide.md"))
        'AUTH-001'
        >>> find_related_spec(Path("docs/api/endpoints.md"))
        'API-001'
    """
    # Extract potential domain from file path
    # E.g., docs/auth/setup.md -> 'auth'
    path_parts = doc_path.parts
    if len(path_parts) < 2:
        return None

    # Get the first directory under 'docs/' as potential domain
    potential_domain = path_parts[1] if path_parts[0] == "docs" else path_parts[0]
    potential_domain = potential_domain.upper().replace("_", "-")

    # Search for matching SPEC directories
    if not specs_dir.exists():
        return None

    matching_specs = []
    for spec_dir in specs_dir.glob("SPEC-*"):
        spec_file = spec_dir / "spec.md"
        if not spec_file.exists():
            continue

        try:
            with open(spec_file, "r", encoding="utf-8") as f:
                content = f.read()
                spec_id = extract_spec_id(content)
                spec_domain = parse_domain(spec_id)

                # Case-insensitive domain match
                if spec_domain.upper() == potential_domain:
                    matching_specs.append(spec_id)
        except (ValueError, OSError):
            # Skip invalid or unreadable SPEC files
            continue

    if not matching_specs:
        return None

    # Return most recent SPEC (highest number)
    def extract_number(spec: str) -> int:
        match: Any = re.search(r"-(\d{3})$", spec)
        return int(match.group(1)) if match else 0

    matching_specs.sort(key=extract_number, reverse=True)
    return str(matching_specs[0])


def calculate_confidence(spec_id: str, doc_path: Path) -> float:
    """Calculate confidence score for SPEC-DOC mapping.

    Confidence levels:
    - 0.95+: Exact SPEC ID in file path
    - 0.80+: Domain match + relevant keywords in filename
    - 0.50+: Domain match only
    - <0.50: No domain match

    Args:
        spec_id: SPEC ID (e.g., "AUTH-001")
        doc_path: Path to documentation file

    Returns:
        Confidence score (0.0 to 1.0)

    Examples:
        >>> calculate_confidence("AUTH-001", Path("docs/auth/AUTH-001-impl.md"))
        0.95
        >>> calculate_confidence("AUTH-001", Path("docs/auth/authentication.md"))
        0.85
        >>> calculate_confidence("AUTH-001", Path("docs/api/guide.md"))
        0.2
    """
    doc_str = str(doc_path).lower()
    spec_id_lower = spec_id.lower()
    domain = parse_domain(spec_id).lower()

    # Exact SPEC ID match in path
    if spec_id_lower in doc_str:
        return 0.95

    # Domain match + relevant keywords
    if domain in doc_str:
        # Keywords related to the domain
        keywords = [domain, "guide", "setup", "implementation", "tutorial"]
        keyword_matches = sum(1 for kw in keywords if kw in doc_str)

        if keyword_matches >= 2:
            return 0.80
        else:
            return 0.50

    # No domain match
    return 0.20

# @CODE:PAR-001 | Chain: @SPEC:DOC-TAG-001 -> @CODE:PAR-001
"""SPEC parser utilities for TAG generation system.

This module extracts SPEC metadata (ID, domain, title) from SPEC documents
for use in TAG generation and SPEC-DOC mapping.

@SPEC:DOC-TAG-001: @DOC tag automatic generation infrastructure
"""

import re
from typing import Any

import yaml


def extract_spec_id(spec_content: str) -> str:
    """Extract SPEC ID from YAML frontmatter.

    Args:
        spec_content: Full content of SPEC markdown file

    Returns:
        SPEC ID (e.g., "AUTH-001")

    Raises:
        ValueError: If YAML frontmatter or ID not found

    Examples:
        >>> content = "---\\nid: AUTH-001\\n---\\n# SPEC"
        >>> extract_spec_id(content)
        'AUTH-001'
    """
    # Find YAML frontmatter between --- markers
    match = re.search(r"^---\s*\n(.*?)\n---\s*$", spec_content, re.MULTILINE | re.DOTALL)
    if not match:
        raise ValueError("YAML frontmatter not found in SPEC document")

    # Parse YAML
    yaml_content = match.group(1)
    try:
        metadata: Any = yaml.safe_load(yaml_content)
        spec_id: Any = metadata.get("id")
        if not spec_id:
            raise ValueError("'id' field not found in YAML frontmatter")
        return str(spec_id)
    except yaml.YAMLError as e:
        raise ValueError(f"Invalid YAML frontmatter: {e}")


def parse_domain(spec_id: str) -> str:
    """Extract domain from SPEC ID.

    The domain is everything before the final -NNN sequence.

    Examples:
        >>> parse_domain("AUTH-001")
        'AUTH'
        >>> parse_domain("CLI-TOOL-001")
        'CLI-TOOL'
        >>> parse_domain("DOC-TAG-001")
        'DOC-TAG'

    Args:
        spec_id: Full SPEC ID

    Returns:
        Domain part (everything before last "-NNN")

    Raises:
        ValueError: Invalid SPEC ID format
    """
    # Match pattern: any chars followed by hyphen and exactly 3 digits at end
    match = re.match(r"^(.*?)-\d{3}$", spec_id)
    if not match:
        raise ValueError(f"Invalid SPEC ID format: {spec_id}")
    return match.group(1)

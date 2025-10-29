# @CODE:VAL-001
"""TAG suggestion and validation orchestrator.

Combines parser, generator, mapper, and inserter modules to provide
high-level TAG suggestion functionality for documentation files.

@SPEC:DOC-TAG-001: @DOC tag automatic generation infrastructure
"""

from dataclasses import dataclass
from pathlib import Path
from typing import Optional

from moai_adk.core.tags.generator import detect_duplicates, generate_doc_tag
from moai_adk.core.tags.mapper import calculate_confidence, find_related_spec
from moai_adk.core.tags.parser import parse_domain


@dataclass
class TagSuggestion:
    """Suggested TAG for a documentation file.

    Attributes:
        tag_id: Generated TAG ID (e.g., "@DOC:AUTH-001")
        chain_ref: Chain reference to SPEC (e.g., "@SPEC:AUTH-001")
        confidence: Confidence score (0.0 to 1.0)
        domain: Extracted domain (e.g., "AUTH")
        file_path: Path to documentation file
    """

    tag_id: str
    chain_ref: Optional[str]
    confidence: float
    domain: str
    file_path: Path


def suggest_tag_for_file(
    doc_path: Path, search_path: Path = Path("docs/")
) -> TagSuggestion:
    """Suggest TAG ID for documentation file.

    Combines domain detection, SPEC mapping, and TAG generation to
    provide a complete TAG suggestion with chain reference.

    Args:
        doc_path: Path to documentation file
        search_path: Directory to search for existing tags

    Returns:
        TagSuggestion with generated TAG and chain reference

    Examples:
        >>> suggestion = suggest_tag_for_file(Path("docs/auth/guide.md"))
        >>> suggestion.tag_id
        '@DOC:AUTH-001'
        >>> suggestion.chain_ref
        '@SPEC:AUTH-001'
        >>> suggestion.confidence
        0.85
    """
    # Find related SPEC
    spec_id = find_related_spec(doc_path)

    # Extract domain
    if spec_id:
        domain = parse_domain(spec_id)
        chain_ref = f"@SPEC:{spec_id}"
        confidence = calculate_confidence(spec_id, doc_path)
    else:
        # Infer domain from file path if no SPEC found
        domain = _infer_domain_from_path(doc_path)
        chain_ref = None
        confidence = 0.3

    # Detect existing TAGs in the domain
    existing_ids = detect_duplicates(domain, str(search_path))

    # Generate next TAG ID
    tag_id = generate_doc_tag(domain, existing_ids)

    return TagSuggestion(
        tag_id=tag_id,
        chain_ref=chain_ref,
        confidence=confidence,
        domain=domain,
        file_path=doc_path,
    )


def _infer_domain_from_path(doc_path: Path) -> str:
    """Infer domain from file path.

    Args:
        doc_path: Path to documentation file

    Returns:
        Inferred domain (uppercase)

    Examples:
        >>> _infer_domain_from_path(Path("docs/auth/guide.md"))
        'AUTH'
        >>> _infer_domain_from_path(Path("docs/cli-tool/commands.md"))
        'CLI-TOOL'
    """
    path_parts = doc_path.parts
    if len(path_parts) < 2:
        return "DOC"

    # Get first directory under 'docs/' as domain
    domain = path_parts[1] if path_parts[0] == "docs" else path_parts[0]
    return domain.upper().replace("_", "-")


def validate_tag_chain(tag_id: str, chain_ref: str) -> bool:
    """Validate TAG chain reference consistency.

    Checks that TAG domain matches SPEC domain in chain reference.

    Args:
        tag_id: TAG ID (e.g., "@DOC:AUTH-001")
        chain_ref: Chain reference (e.g., "@SPEC:AUTH-001 -> @DOC:AUTH-001")

    Returns:
        True if chain is valid

    Examples:
        >>> validate_tag_chain("@DOC:AUTH-001", "@SPEC:AUTH-001 -> @DOC:AUTH-001")
        True
        >>> validate_tag_chain("@DOC:AUTH-001", "@SPEC:API-001 -> @DOC:AUTH-001")
        False
    """
    # Check if chain contains TAG ID
    if tag_id not in chain_ref:
        return False

    # Extract TAG domain (e.g., "@DOC:AUTH-001" -> "AUTH")
    tag_domain = tag_id.split(":")[1].rsplit("-", 1)[0]

    # Extract SPEC part from chain (e.g., "@SPEC:AUTH-001" or "@SPEC:API-001")
    import re
    spec_match = re.search(r"@SPEC:([A-Z0-9-]+)-\d{3}", chain_ref)
    if not spec_match:
        return False

    spec_domain = spec_match.group(1)

    # Domains must match
    return spec_domain == tag_domain

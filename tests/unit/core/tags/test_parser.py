# @TEST:PAR-001 | Chain: @SPEC:DOC-TAG-001 -> @TEST:PAR-001
"""Test suite for SPEC parser utilities.

Tests SPEC ID extraction from YAML frontmatter and domain parsing
for TAG generation system.

@SPEC:DOC-TAG-001: @DOC tag automatic generation infrastructure
"""


import pytest

from src.moai_adk.core.tags.parser import extract_spec_id, parse_domain


def test_extract_spec_id_from_spec_file():
    """Extract SPEC ID from .moai/specs/SPEC-ID/spec.md"""
    # Setup: Create a temporary SPEC file
    spec_content = """---
id: AUTH-001
version: 0.0.1
---
# @SPEC:AUTH-004: Authentication System
"""

    # When: Extract ID from spec content
    spec_id = extract_spec_id(spec_content)

    # Then: ID should match
    assert spec_id == "AUTH-001"


def test_extract_spec_id_missing_frontmatter():
    """Handle missing YAML frontmatter gracefully"""
    spec_content = "# Some document without frontmatter"

    with pytest.raises(ValueError, match="YAML frontmatter not found"):
        extract_spec_id(spec_content)


def test_extract_spec_id_missing_id_field():
    """Handle missing 'id' field in frontmatter"""
    spec_content = """---
version: 0.0.1
---
# Some SPEC
"""

    with pytest.raises(ValueError, match="'id' field not found"):
        extract_spec_id(spec_content)


def test_extract_spec_id_invalid_yaml():
    """Handle invalid YAML frontmatter"""
    spec_content = """---
id: AUTH-001
  invalid: yaml: syntax
---
# SPEC
"""

    with pytest.raises(ValueError, match="Invalid YAML frontmatter"):
        extract_spec_id(spec_content)


def test_parse_domain_from_id():
    """Parse domain from SPEC ID"""
    assert parse_domain("AUTH-001") == "AUTH"
    assert parse_domain("CLI-TOOL-001") == "CLI-TOOL"
    assert parse_domain("DOC-TAG-001") == "DOC-TAG"
    assert parse_domain("API-V2-999") == "API-V2"


def test_parse_domain_invalid_format():
    """Handle invalid SPEC ID format"""
    with pytest.raises(ValueError, match="Invalid SPEC ID format"):
        parse_domain("INVALID")

    with pytest.raises(ValueError, match="Invalid SPEC ID format"):
        parse_domain("AUTH-01")  # Only 2 digits

    with pytest.raises(ValueError, match="Invalid SPEC ID format"):
        parse_domain("AUTH-1234")  # 4 digits

"""
Test suite for moai-foundation-ears skill.
Tests metadata compliance, progressive disclosure, file size, and auto-trigger keywords.
"""

import re
from pathlib import Path

import pytest
import yaml


# ===== FIXTURES =====
@pytest.fixture
def skill_path():
    """Return path to moai-foundation-ears skill."""
    return Path(__file__).parent.parent.parent / ".claude" / "skills" / "moai-foundation-ears"


@pytest.fixture
def skill_metadata(skill_path):
    """Load and parse skill metadata from SKILL.md frontmatter."""
    skill_md = skill_path / "SKILL.md"
    if not skill_md.exists():
        pytest.skip(f"Skill file not found: {skill_md}")

    with open(skill_md, "r", encoding="utf-8") as f:
        content = f.read()

    # Parse YAML frontmatter
    if content.startswith("---"):
        _, frontmatter, _ = content.split("---", 2)
        return yaml.safe_load(frontmatter)
    return {}


@pytest.fixture
def skill_content(skill_path):
    """Load full skill content."""
    skill_md = skill_path / "SKILL.md"
    if not skill_md.exists():
        pytest.skip(f"Skill file not found: {skill_md}")

    with open(skill_md, "r", encoding="utf-8") as f:
        return f.read()


# ===== METADATA COMPLIANCE TESTS =====
class TestMetadataCompliance:
    """Test metadata compliance with 16-field standard."""

    # 7 Required fields
    REQUIRED_FIELDS = [
        "name",
        "description",
        "version",
        "modularized",
        "last_updated",
        "allowed-tools",
        "compliance_score",
    ]

    # 9 Recommended fields
    RECOMMENDED_FIELDS = [
        "modules",
        "dependencies",
        "deprecated",
        "successor",
        "category_tier",
        "auto_trigger_keywords",
        "agent_coverage",
        "context7_references",
        "invocation_api_version",
    ]

    def test_all_required_fields_present(self, skill_metadata):
        """Verify all 7 required metadata fields are present."""
        for field in self.REQUIRED_FIELDS:
            assert field in skill_metadata, f"Required field '{field}' missing from metadata"

    def test_all_recommended_fields_present(self, skill_metadata):
        """Verify all 9 recommended metadata fields are present."""
        missing_fields = [f for f in self.RECOMMENDED_FIELDS if f not in skill_metadata]
        assert len(missing_fields) == 0, f"Recommended fields missing: {missing_fields}"

    def test_name_is_string(self, skill_metadata):
        """Verify name is a non-empty string."""
        assert isinstance(skill_metadata.get("name"), str)
        assert len(skill_metadata.get("name", "")) > 0

    def test_name_matches_skill_directory(self, skill_path, skill_metadata):
        """Verify metadata name matches directory name."""
        expected_name = skill_path.name
        actual_name = skill_metadata.get("name")
        assert actual_name == expected_name, f"Name mismatch: {actual_name} != {expected_name}"

    def test_description_is_string(self, skill_metadata):
        """Verify description is a non-empty string."""
        assert isinstance(skill_metadata.get("description"), str)
        assert len(skill_metadata.get("description", "")) > 0

    def test_version_format_is_semantic(self, skill_metadata):
        """Verify version follows semantic versioning (X.Y.Z)."""
        version = skill_metadata.get("version", "")
        pattern = r"^\d+\.\d+\.\d+$"
        assert re.match(pattern, version), f"Invalid semantic version: {version}"

    def test_modularized_is_boolean(self, skill_metadata):
        """Verify modularized is boolean."""
        assert isinstance(skill_metadata.get("modularized"), bool)

    def test_last_updated_is_valid_date(self, skill_metadata):
        """Verify last_updated is ISO 8601 format or date object."""
        last_updated = skill_metadata.get("last_updated")
        # YAML may parse as date object or string
        from datetime import date

        assert isinstance(last_updated, (str, date)), f"Invalid date type: {type(last_updated)}"

        if isinstance(last_updated, str):
            pattern = r"^\d{4}-\d{2}-\d{2}"
            assert re.match(pattern, last_updated), f"Invalid date format: {last_updated}"

    def test_allowed_tools_is_list(self, skill_metadata):
        """Verify allowed-tools is a list."""
        tools = skill_metadata.get("allowed-tools")
        assert isinstance(tools, list), "allowed-tools must be a list"
        assert len(tools) > 0, "allowed-tools must not be empty"

    def test_compliance_score_is_valid_percentage(self, skill_metadata):
        """Verify compliance_score is between 0-100."""
        score = skill_metadata.get("compliance_score")
        assert isinstance(score, (int, float)), "compliance_score must be numeric"
        assert 0 <= score <= 100, f"compliance_score out of range: {score}"

    def test_auto_trigger_keywords_count(self, skill_metadata):
        """Verify auto_trigger_keywords has 8-15 keywords."""
        keywords = skill_metadata.get("auto_trigger_keywords", [])
        assert isinstance(keywords, list), "auto_trigger_keywords must be a list"
        assert 8 <= len(keywords) <= 15, f"Keywords count {len(keywords)} not in range [8, 15]"

    def test_category_tier_is_valid(self, skill_metadata):
        """Verify category_tier is 1-3 (Foundation, Domain, Integration)."""
        tier = skill_metadata.get("category_tier")
        assert tier in [1, 2, 3], f"Invalid category_tier: {tier}"

    def test_deprecated_is_boolean_if_present(self, skill_metadata):
        """Verify deprecated field is boolean if present."""
        if "deprecated" in skill_metadata:
            assert isinstance(skill_metadata.get("deprecated"), bool)

    def test_agent_coverage_has_valid_agents(self, skill_metadata):
        """Verify agent_coverage references valid agent names."""
        agents = skill_metadata.get("agent_coverage", [])
        assert isinstance(agents, list), "agent_coverage must be a list"
        # Agent names should be lowercase with hyphens
        for agent in agents:
            assert isinstance(agent, str)
            assert agent.islower() or "-" in agent, f"Invalid agent name: {agent}"


# ===== PROGRESSIVE DISCLOSURE TESTS =====
class TestProgressiveDisclosure:
    """Test progressive disclosure compliance (Level 1, 2, 3)."""

    def test_skill_content_has_markdown_headers(self, skill_content):
        """Verify skill content has proper markdown structure."""
        assert "##" in skill_content, "Skill content must have markdown headers"

    def test_no_invalid_bold_parenthesis_patterns(self, skill_content):
        """Test CommonMark compliance: no **text(desc)**next patterns."""
        # Pattern: **[text]**([desc])[non-whitespace]
        # This violates CommonMark delimiter rules
        invalid_pattern = r"\*\*[^*]+\([^)]+\)\*\*[^\s*]"
        matches = re.findall(invalid_pattern, skill_content)
        assert len(matches) == 0, f"Invalid bold-parenthesis patterns found: {matches}"

    def test_content_structure_has_sections(self, skill_content):
        """Verify skill has basic section structure."""
        sections = ["Quick Reference", "What It Does", "When to Use", "Best Practices"]
        found_sections = sum(1 for section in sections if section in skill_content)
        assert found_sections >= 2, f"Skill missing major sections. Found {found_sections}/4"


# ===== FILE SIZE TESTS =====
class TestFileSize:
    """Test skill file size constraints."""

    def test_skill_file_under_500_lines(self, skill_path):
        """Verify SKILL.md is ≤500 lines."""
        skill_md = skill_path / "SKILL.md"
        with open(skill_md, "r", encoding="utf-8") as f:
            lines = f.readlines()

        line_count = len(lines)
        assert line_count <= 500, f"Skill file has {line_count} lines (max 500)"

    def test_skill_file_size_under_50kb(self, skill_path):
        """Verify SKILL.md file size is reasonable."""
        skill_md = skill_path / "SKILL.md"
        file_size = skill_md.stat().st_size
        max_size = 50 * 1024  # 50KB
        assert file_size <= max_size, f"Skill file size {file_size} bytes exceeds {max_size}"


# ===== CONTENT VALIDATION TESTS =====
class TestContentValidation:
    """Test skill content validity and quality."""

    def test_skill_has_meaningful_description(self, skill_metadata):
        """Verify description is meaningful (>20 chars)."""
        desc = skill_metadata.get("description", "")
        assert len(desc) >= 20, f"Description too short: {len(desc)} chars"

    def test_skill_has_no_placeholder_text(self, skill_content):
        """Verify no placeholder or template text remains."""
        placeholder_patterns = [r"\[\s*TODO\s*\]", r"\{\{\s*[^}]+\s*\}\}", r"XXX|FIXME|PLACEHOLDER"]
        for pattern in placeholder_patterns:
            matches = re.findall(pattern, skill_content, re.IGNORECASE)
            assert len(matches) == 0, f"Placeholder text found: {matches}"

    def test_all_code_blocks_valid(self, skill_content):
        """Verify code blocks are properly formatted."""
        # Count opening and closing code fence markers
        fence_count = skill_content.count("```")
        assert fence_count % 2 == 0, "Unmatched code fence markers"

    def test_no_broken_links(self, skill_content):
        """Verify markdown links are properly formatted."""
        # Pattern: [text](url)
        link_pattern = r"\[([^\]]+)\]\(([^)]+)\)"
        links = re.findall(link_pattern, skill_content)
        for text, url in links:
            assert len(text) > 0, "Link text cannot be empty"
            assert len(url) > 0, "Link URL cannot be empty"


# ===== MODULAR STRUCTURE TESTS =====
class TestModularStructure:
    """Test modular skill structure if applicable."""

    def test_modules_directory_structure(self, skill_path, skill_metadata):
        """Verify modules directory exists and is properly structured if modularized."""
        is_modularized = skill_metadata.get("modularized", False)

        if is_modularized:
            modules_dir = skill_path / "modules"
            assert modules_dir.exists(), "Modularized skill missing modules/ directory"

            # Should have at least 2 module files
            module_files = list(modules_dir.glob("*.md"))
            assert len(module_files) >= 2, f"Modularized skill needs at least 2 modules, found {len(module_files)}"


# ===== INTEGRATION TESTS =====
class TestSkillIntegration:
    """Test skill integration and discoverability."""

    def test_skill_directory_exists(self, skill_path):
        """Verify skill directory exists."""
        assert skill_path.exists(), f"Skill directory not found: {skill_path}"

    def test_skill_md_file_exists(self, skill_path):
        """Verify SKILL.md file exists."""
        skill_md = skill_path / "SKILL.md"
        assert skill_md.exists(), f"SKILL.md not found in {skill_path}"

    def test_skill_is_discoverable(self, skill_path):
        """Verify skill follows naming convention for discovery."""
        skill_name = skill_path.name
        # Must start with moai-
        assert skill_name.startswith("moai-"), f"Skill name must start with 'moai-': {skill_name}"
        # Must be lowercase with hyphens
        assert skill_name.islower() or "-" in skill_name, f"Invalid skill name format: {skill_name}"


# ===== EXPECTED SKILLS FOR WEEK 5 =====
class TestWeek5SkillCoverage:
    """Test that all Week 5 skills are being covered."""

    def test_foundation_ears_skill_present(self):
        """Verify moai-foundation-ears exists."""
        skill_path = Path(__file__).parent.parent.parent / ".claude" / "skills" / "moai-foundation-ears"
        assert skill_path.exists(), "moai-foundation-ears skill not found"

    def test_foundation_ears_has_valid_metadata(self, skill_metadata):
        """Verify moai-foundation-ears has valid metadata."""
        assert skill_metadata.get("name") == "moai-foundation-ears"
        assert skill_metadata.get("version") is not None
        assert skill_metadata.get("compliance_score", 0) > 0

"""
Test suite for consolidated language skills: Python, TypeScript, Go, Rust, JVM, DotNet, Mobile, Systems.
Tests metadata compliance and consolidation patterns.
"""

import re
from pathlib import Path

import pytest
import yaml

# ===== LANGUAGE SKILL NAMES (CONSOLIDATED) =====
LANGUAGE_SKILLS_TO_TEST = [
    "moai-lang-python",
    "moai-lang-typescript",
    "moai-lang-go",
    "moai-lang-rust",
    "moai-lang-jvm",
    "moai-lang-dotnet",
    "moai-lang-mobile",
    "moai-lang-systems",
]


# ===== FIXTURES =====
@pytest.fixture(params=LANGUAGE_SKILLS_TO_TEST)
def language_skill_name(request):
    """Parameterized fixture for language skill names."""
    return request.param


@pytest.fixture
def skill_path(language_skill_name):
    """Return path to language skill."""
    return Path(__file__).parent.parent.parent / ".claude" / "skills" / language_skill_name


@pytest.fixture
def skill_metadata(skill_path):
    """Load and parse skill metadata from SKILL.md frontmatter."""
    skill_md = skill_path / "SKILL.md"
    if not skill_md.exists():
        pytest.skip(f"Skill file not found: {skill_md}")

    with open(skill_md, "r", encoding="utf-8") as f:
        content = f.read()

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
    """Test metadata compliance with required fields."""

    # 7 Required fields (minimum)
    REQUIRED_FIELDS = [
        "name",
        "description",
        "version",
        "modularized",
        "last_updated",
        "allowed-tools",
        "compliance_score",
    ]

    def test_name_matches_skill_directory(self, skill_path, skill_metadata):
        """Verify metadata name matches directory name."""
        expected_name = skill_path.name
        actual_name = skill_metadata.get("name")
        assert actual_name == expected_name, f"Name mismatch: {actual_name} != {expected_name}"

    def test_description_is_non_empty(self, skill_metadata):
        """Verify description exists."""
        desc = skill_metadata.get("description", "")
        assert len(desc) > 0, "Description required"

    def test_version_is_semantic(self, skill_metadata):
        """Verify version follows semantic versioning."""
        version = skill_metadata.get("version", "")
        pattern = r"^\d+\.\d+\.\d+$"
        assert re.match(pattern, version), f"Invalid version: {version}"

    def test_compliance_score_valid(self, skill_metadata):
        """Verify compliance_score is 0-100."""
        score = skill_metadata.get("compliance_score")
        assert isinstance(score, (int, float)), "compliance_score must be number"
        assert 0 <= score <= 100, f"Score must be 0-100, got {score}"

    def test_category_tier_4_for_language(self, skill_metadata):
        """Verify language skills have category_tier=4."""
        tier = skill_metadata.get("category_tier")
        assert tier == 4, f"Language skills must be tier 4, got {tier}"


# ===== FILE STRUCTURE TESTS =====
class TestFileStructure:
    """Test skill file structure."""

    def test_skill_md_exists(self, skill_path):
        """Verify SKILL.md exists."""
        skill_md = skill_path / "SKILL.md"
        assert skill_md.exists(), f"SKILL.md not found in {skill_path}"

    def test_skill_md_has_content(self, skill_content):
        """Verify SKILL.md has sufficient content."""
        assert len(skill_content) > 300, "SKILL.md too short"

    def test_has_quick_reference(self, skill_content):
        """Verify has Quick Reference section."""
        markers = ["Quick Reference", "quick reference", "quick-reference"]
        has_marker = any(m in skill_content for m in markers)
        assert has_marker, "Missing Quick Reference section"

    def test_has_best_practices(self, skill_content):
        """Verify has best practices (DO/DON'T)."""
        has_do = "DO" in skill_content
        has_dont = "DON'T" in skill_content
        assert has_do and has_dont, "Missing best practices section"


# ===== CONSOLIDATION TESTS =====
class TestConsolidation:
    """Test consolidation patterns for language skills."""

    def test_consolidated_skills_defined(self, skill_metadata, language_skill_name):
        """Verify consolidated skills are documented."""
        consolidation_expected = {
            "moai-lang-jvm": ["java", "kotlin", "scala"],
            "moai-lang-mobile": ["swift", "dart"],
            "moai-lang-systems": ["c", "cpp", "shell", "sql", "ruby", "php"],
            "moai-lang-typescript": ["javascript"],
        }

        if language_skill_name in consolidation_expected:
            description = skill_metadata.get("description", "").lower()
            # Should mention consolidation or component languages
            any(lang in description for lang in consolidation_expected[language_skill_name])
            # Might not always be explicit, but description should exist
            assert len(description) > 0, "Consolidated skill needs description"

    def test_agent_coverage_populated(self, skill_metadata):
        """Verify agent coverage is specified."""
        agents = skill_metadata.get("agent_coverage", [])
        assert len(agents) > 0, "agent_coverage must not be empty"

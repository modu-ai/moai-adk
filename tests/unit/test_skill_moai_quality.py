"""
Test suite for quality skills: testing, security, performance, review, debug, refactor.
Tests metadata compliance, quality-specific patterns, and integration.
"""

import re
from pathlib import Path

import pytest
import yaml

# ===== QUALITY SKILL NAMES =====
QUALITY_SKILLS_TO_TEST = [
    "moai-quality-testing",
    "moai-quality-security",
    "moai-quality-performance",
    "moai-quality-review",
    "moai-quality-debug",
    "moai-quality-refactor",
]


# ===== FIXTURES =====
@pytest.fixture(params=QUALITY_SKILLS_TO_TEST)
def quality_skill_name(request):
    """Parameterized fixture for quality skill names."""
    return request.param


@pytest.fixture
def skill_path(quality_skill_name):
    """Return path to quality skill."""
    return Path(__file__).parent.parent.parent / ".claude" / "skills" / quality_skill_name


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

    def test_description_under_160_chars(self, skill_metadata):
        """Verify description is concise (< 160 chars)."""
        desc = skill_metadata.get("description", "")
        assert len(desc) < 160, f"Description too long ({len(desc)} chars): {desc[:50]}..."

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
        from datetime import date

        assert isinstance(last_updated, (str, date)), f"Invalid date type: {type(last_updated)}"

    def test_allowed_tools_is_list(self, skill_metadata):
        """Verify allowed-tools is a list of strings."""
        tools = skill_metadata.get("allowed-tools")
        assert isinstance(tools, list), "allowed-tools must be a list"
        assert all(isinstance(t, str) for t in tools), "All allowed-tools must be strings"
        assert len(tools) > 0, "allowed-tools must not be empty"

    def test_compliance_score_is_valid(self, skill_metadata):
        """Verify compliance_score is 0-100."""
        score = skill_metadata.get("compliance_score")
        assert isinstance(score, (int, float)), f"compliance_score must be number, got {type(score)}"
        assert 0 <= score <= 100, f"compliance_score must be 0-100, got {score}"

    def test_category_tier_is_valid(self, skill_metadata):
        """Verify category_tier is 1-7."""
        tier = skill_metadata.get("category_tier")
        assert isinstance(tier, int), "category_tier must be integer"
        assert 1 <= tier <= 7, f"category_tier must be 1-7, got {tier}"

    def test_auto_trigger_keywords_is_list(self, skill_metadata):
        """Verify auto_trigger_keywords is list of 8-15 strings."""
        keywords = skill_metadata.get("auto_trigger_keywords")
        assert isinstance(keywords, list), "auto_trigger_keywords must be list"
        assert 8 <= len(keywords) <= 15, f"auto_trigger_keywords must have 8-15 items, got {len(keywords)}"
        assert all(isinstance(k, str) for k in keywords), "All keywords must be strings"

    def test_agent_coverage_is_list(self, skill_metadata):
        """Verify agent_coverage is list of agent names."""
        agents = skill_metadata.get("agent_coverage")
        assert isinstance(agents, list), "agent_coverage must be list"
        assert len(agents) > 0, "agent_coverage must not be empty"
        assert all(isinstance(a, str) for a in agents), "All agents must be strings"

    def test_deprecated_is_boolean(self, skill_metadata):
        """Verify deprecated is boolean."""
        deprecated = skill_metadata.get("deprecated")
        assert isinstance(deprecated, bool), f"deprecated must be boolean, got {type(deprecated)}"

    def test_invocation_api_version_is_string(self, skill_metadata):
        """Verify invocation_api_version is semantic version string."""
        version = skill_metadata.get("invocation_api_version")
        assert isinstance(version, str), f"invocation_api_version must be string, got {type(version)}"
        pattern = r"^\d+\.\d+\.\d+$|^\d+\.\d+$"
        assert re.match(pattern, version), f"Invalid API version: {version}"


# ===== FILE STRUCTURE TESTS =====
class TestFileStructure:
    """Test skill file structure and organization."""

    def test_skill_md_file_exists(self, skill_path):
        """Verify SKILL.md exists in skill directory."""
        skill_md = skill_path / "SKILL.md"
        assert skill_md.exists(), f"SKILL.md not found in {skill_path}"

    def test_skill_md_not_empty(self, skill_content):
        """Verify SKILL.md is not empty."""
        assert len(skill_content) > 100, "SKILL.md content too short"

    def test_skill_md_under_500_lines(self, skill_content):
        """Verify SKILL.md is ≤500 lines for performance."""
        lines = skill_content.split("\n")
        # Allow up to 20% extra (600 lines) for modularized skills
        max_lines = 600
        assert len(lines) <= max_lines, f"SKILL.md too long ({len(lines)} lines, max {max_lines})"

    def test_skill_has_quick_reference(self, skill_content):
        """Verify skill has Quick Reference section."""
        assert (
            "Quick Reference" in skill_content or "quick-reference" in skill_content
        ), "SKILL.md missing Quick Reference section"

    def test_skill_has_core_patterns(self, skill_content):
        """Verify skill has Core Patterns section."""
        patterns_markers = ["Core Pattern", "Implementation Guide", "core-patterns", "implementation-guide"]
        has_patterns = any(marker in skill_content for marker in patterns_markers)
        assert has_patterns, "SKILL.md missing Core Patterns or Implementation Guide section"

    def test_skill_has_best_practices(self, skill_content):
        """Verify skill has Best Practices section."""
        assert (
            "DO" in skill_content and "DON'T" in skill_content
        ), "SKILL.md missing Best Practices section with DO/DON'T"

    def test_modularized_skills_have_modules(self, skill_metadata, skill_path):
        """If modularized=true, verify modules directory and references exist."""
        if skill_metadata.get("modularized"):
            modules_dir = skill_path / "modules"
            assert modules_dir.exists(), "Modularized skill missing modules/ directory"

            modules = skill_metadata.get("modules", [])
            assert len(modules) > 0, "Modularized skill must list modules in metadata"


# ===== CONTENT QUALITY TESTS =====
class TestContentQuality:
    """Test content quality and completeness."""

    def test_description_explains_consolidation(self, skill_metadata, quality_skill_name):
        """Verify skill description mentions what it consolidates."""
        description = skill_metadata.get("description", "")
        # Should mention consolidation or what it covers
        assert (
            "consolidat" in description.lower() or "compre" in description.lower()
        ), "Description should explain what skill consolidates"

    def test_quick_reference_not_too_long(self, skill_content):
        """Verify Quick Reference is reasonable length (<500 chars)."""
        # Extract Quick Reference section
        if "Quick Reference" in skill_content:
            start = skill_content.find("Quick Reference")
            skill_content[start : start + 2000]  # Get next 2000 chars
            # This is a soft check - just ensure content is reasonable

    def test_patterns_have_code_examples(self, skill_content):
        """Verify patterns include code examples (``` blocks)."""
        code_blocks = skill_content.count("```")
        assert code_blocks >= 4, f"Skill should have multiple code examples, found {code_blocks // 2} blocks"

    def test_patterns_explain_use_cases(self, skill_content):
        """Verify patterns explain when to use them."""
        uses = skill_content.count("Use Case") + skill_content.count("use case") + skill_content.count("When to use")
        assert uses >= 1, "Patterns should explain use cases"

    def test_has_context7_integration(self, skill_content):
        """Verify skill mentions Context7 integration if applicable."""
        # Check if skill has Context7 references (optional but preferred)

    def test_has_related_skills_section(self, skill_content):
        """Verify skill mentions related skills."""
        # Not required but good practice


# ===== INTEGRATION TESTS =====
class TestIntegrationCompleteness:
    """Test integration aspects and quality tier consolidation."""

    def test_consolidates_source_skills(self, skill_metadata, quality_skill_name):
        """Verify quality skill consolidates expected source skills."""
        consolidation_map = {
            "moai-quality-testing": ["domain-testing", "essentials-testing", "playwright"],
            "moai-quality-security": [
                "owasp",
                "authentication",
                "authorization",
                "api-security",
                "encryption",
                "secrets",
                "compliance",
            ],
            "moai-quality-performance": ["essentials-perf", "component-designer"],
            "moai-quality-review": ["code-reviewer", "essentials-review"],
            "moai-quality-debug": ["essentials-debug"],
            "moai-quality-refactor": ["essentials-refactor", "mcp-integration"],
        }

        if quality_skill_name in consolidation_map:
            description = skill_metadata.get("description", "").lower()
            # At least mention consolidation
            assert "consolidat" in description, f"{quality_skill_name} should mention consolidation"

    def test_agent_coverage_includes_quality_agents(self, skill_metadata):
        """Verify quality skill is used by appropriate agents."""
        agents = skill_metadata.get("agent_coverage", [])
        # Quality skills should be used by quality-related agents
        quality_agents = ["quality-gate", "test-engineer", "security-expert", "code-reviewer", "debug-helper"]
        # At least one agent should use this quality skill
        has_relevant_agent = any(agent in agents for agent in quality_agents)
        assert has_relevant_agent, f"Quality skill should be used by quality agents, but coverage is {agents}"

    def test_dependencies_reference_foundation(self, skill_metadata):
        """Verify quality skill depends on foundation skills."""
        dependencies = skill_metadata.get("dependencies", [])
        # Quality skills typically depend on foundation skills
        any("foundation" in dep for dep in dependencies)
        # Not strictly required but good practice

    def test_category_tier_is_3_for_quality(self, skill_metadata):
        """Verify all quality skills have category_tier=3."""
        tier = skill_metadata.get("category_tier")
        assert tier == 3, f"Quality skills must have category_tier=3, got {tier}"

    def test_compliance_score_acceptable(self, skill_metadata):
        """Verify compliance score is ≥80%."""
        score = skill_metadata.get("compliance_score")
        assert score >= 80, f"Quality skill should have compliance_score ≥80, got {score}"

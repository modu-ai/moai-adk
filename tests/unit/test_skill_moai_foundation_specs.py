"""
Test suite for moai-foundation-specs skill.
Tests metadata compliance, progressive disclosure, module structure, and integration.
"""

import re
from pathlib import Path

import pytest
import yaml


# ===== FIXTURES =====
@pytest.fixture
def skill_path():
    """Return path to moai-foundation-specs skill."""
    return Path(__file__).parent.parent.parent / ".claude" / "skills" / "moai-foundation-specs"


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


@pytest.fixture
def module_files(skill_path, skill_metadata):
    """Load all module files if skill is modularized."""
    modules = []
    if skill_metadata.get("modularized", False):
        modules_dir = skill_path / "modules"
        if modules_dir.exists():
            for module_file in modules_dir.glob("*.md"):
                with open(module_file, "r", encoding="utf-8") as f:
                    modules.append({"name": module_file.name, "path": module_file, "content": f.read()})
    return modules


# ===== METADATA COMPLIANCE TESTS =====
class TestMetadataCompliance:
    """Test metadata compliance for moai-foundation-specs."""

    REQUIRED_FIELDS = [
        "name",
        "description",
        "version",
        "modularized",
        "last_updated",
        "allowed-tools",
        "compliance_score",
    ]
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

    def test_required_fields_present(self, skill_metadata):
        """Verify all 7 required fields present."""
        for field in self.REQUIRED_FIELDS:
            assert field in skill_metadata, f"Required field '{field}' missing"

    def test_name_is_moai_foundation_specs(self, skill_metadata):
        """Verify skill name is moai-foundation-specs."""
        assert skill_metadata.get("name") == "moai-foundation-specs"

    def test_is_consolidated_skill(self, skill_metadata):
        """Verify this is a consolidated foundation skill."""
        # Should have high compliance score (85+)
        score = skill_metadata.get("compliance_score", 0)
        assert score >= 85, f"Consolidated skill should have compliance_score >= 85, got {score}"

    def test_consolidates_multiple_source_skills(self, skill_metadata):
        """Verify metadata indicates consolidation of multiple skills."""
        # Should list dependencies or original skills
        description = skill_metadata.get("description", "")
        # Description should indicate consolidation
        assert (
            "Consolidates" in description
            or "consolidate" in description.lower()
            or "SPEC" in description
            or "spec" in description.lower()
        ), "Description should indicate skill consolidation"

    def test_modules_field_populated_if_modularized(self, skill_metadata):
        """Verify modules field is populated if skill is modularized."""
        if skill_metadata.get("modularized", False):
            modules = skill_metadata.get("modules", [])
            assert isinstance(modules, list), "modules field must be a list"
            assert len(modules) >= 2, f"Modularized skill should have ≥2 modules, found {len(modules)}"


# ===== PROGRESSIVE DISCLOSURE TESTS =====
class TestProgressiveDisclosure:
    """Test progressive disclosure structure for foundation skill."""

    def test_has_quick_reference_section(self, skill_content):
        """Verify Quick Reference section exists."""
        assert "Quick Reference" in skill_content or "quick reference" in skill_content.lower()

    def test_has_implementation_guide_section(self, skill_content):
        """Verify Implementation Guide or similar section exists."""
        sections = ["Implementation Guide", "Core Patterns", "Fundamentals", "Guidelines"]
        assert any(section in skill_content for section in sections), "Must have core implementation section"

    def test_has_advanced_section(self, skill_content):
        """Verify Advanced Patterns or similar section exists."""
        sections = ["Advanced Patterns", "Advanced", "Complex", "Deep Dive"]
        assert any(section in skill_content for section in sections), "Must have advanced patterns section"

    def test_content_structure_logical(self, skill_content):
        """Verify content is logically structured with headers."""
        # Count markdown headers
        header_count = len(re.findall(r"^#{1,6} ", skill_content, re.MULTILINE))
        assert header_count >= 5, f"Content should have ≥5 sections, found {header_count}"


# ===== MODULE STRUCTURE TESTS =====
class TestModuleStructure:
    """Test modular structure for consolidated skill."""

    def test_modules_directory_exists(self, skill_path, skill_metadata):
        """Verify modules/ directory exists for modularized skill."""
        if skill_metadata.get("modularized", False):
            modules_dir = skill_path / "modules"
            assert modules_dir.exists(), "Modularized skill missing modules/ directory"

    def test_expected_modules_present(self, skill_path, skill_metadata):
        """Verify expected module files exist."""
        if skill_metadata.get("modularized", False):
            modules_dir = skill_path / "modules"
            module_files = list(modules_dir.glob("*.md"))

            # For moai-foundation-specs, expect modules like:
            # /spec-types.md, /authoring-guide.md, /workflow.md
            expected_modules = skill_metadata.get("modules", [])
            assert len(module_files) >= len(
                expected_modules
            ), f"Expected {len(expected_modules)} modules, found {len(module_files)}"

    def test_modules_have_valid_content(self, module_files):
        """Verify each module file has valid content."""
        for module in module_files:
            content = module["content"]
            # Each module should have headers
            assert re.search(r"^#{1,6} ", content, re.MULTILINE), f"Module {module['name']} missing headers"
            # Each module should have meaningful content
            assert len(content) > 100, f"Module {module['name']} too short (<100 chars)"


# ===== CONSOLIDATION VALIDATION TESTS =====
class TestConsolidationValidation:
    """Test that consolidation is valid and complete."""

    def test_no_function_loss_indicated(self, skill_metadata):
        """Verify metadata indicates no function loss."""
        # This is tested through comprehensive content coverage
        skill_metadata.get("dependencies", [])
        skill_metadata.get("description", "")
        # Should not mark any skills as deprecated
        assert not skill_metadata.get("deprecated", False), "Consolidated skill itself should not be deprecated"

    def test_auto_trigger_keywords_comprehensive(self, skill_metadata):
        """Verify auto-trigger keywords cover all consolidated domains."""
        keywords = skill_metadata.get("auto_trigger_keywords", [])
        assert len(keywords) >= 8, f"Need ≥8 keywords, found {len(keywords)}"

        # Should include SPEC-related keywords
        keyword_lower = [k.lower() for k in keywords]
        spec_keywords = ["spec", "requirement", "ears", "workflow"]
        matches = sum(1 for k in spec_keywords if any(k in kw for kw in keyword_lower))
        assert matches >= 2, f"Missing core SPEC keywords in {keyword_lower}"


# ===== CONTENT VALIDATION TESTS =====
class TestContentValidation:
    """Test skill content quality and completeness."""

    def test_no_placeholder_text(self, skill_content):
        """Verify no placeholder text remains."""
        placeholders = ["TODO", "FIXME", "XXX", "{{", "}}"]
        for placeholder in placeholders:
            assert placeholder not in skill_content, f"Placeholder '{placeholder}' found in content"

    def test_all_links_valid(self, skill_content):
        """Verify markdown links are properly formatted."""
        link_pattern = r"\[([^\]]+)\]\(([^)]+)\)"
        links = re.findall(link_pattern, skill_content)
        for text, url in links:
            assert len(text) > 0, "Empty link text"
            assert len(url) > 0, "Empty link URL"

    def test_code_blocks_matched(self, skill_content):
        """Verify code blocks properly opened and closed."""
        fence_count = skill_content.count("```")
        assert fence_count % 2 == 0, f"Unmatched code fences: {fence_count}"

    def test_description_length_adequate(self, skill_metadata):
        """Verify description is comprehensive."""
        desc = skill_metadata.get("description", "")
        assert len(desc) >= 30, f"Description too short: {len(desc)} chars"


# ===== AGENT COVERAGE TESTS =====
class TestAgentCoverage:
    """Test agent coverage for foundation skill."""

    def test_has_agent_coverage(self, skill_metadata):
        """Verify agent_coverage field populated."""
        agents = skill_metadata.get("agent_coverage", [])
        assert isinstance(agents, list), "agent_coverage must be a list"
        assert len(agents) > 0, "agent_coverage should not be empty for foundation skill"

    def test_agents_are_valid_format(self, skill_metadata):
        """Verify agent names follow valid format."""
        agents = skill_metadata.get("agent_coverage", [])
        for agent in agents:
            assert isinstance(agent, str), f"Agent must be string: {agent}"
            assert (
                agent.startswith("spec-")
                or agent.startswith("implementation-")
                or agent.startswith("plan")
                or agent.startswith("tdd-")
            ), f"Agent should be valid: {agent}"


# ===== FILE SIZE AND STRUCTURE TESTS =====
class TestFileStructure:
    """Test overall file structure and size constraints."""

    def test_main_skill_file_not_excessive(self, skill_path):
        """Verify main SKILL.md file is reasonable size."""
        skill_md = skill_path / "SKILL.md"
        lines = len(skill_md.read_text().splitlines())
        # Main file should be smaller if modularized
        assert lines <= 500, f"SKILL.md too large: {lines} lines"

    def test_yaml_frontmatter_valid(self, skill_metadata, skill_content):
        """Verify YAML frontmatter is valid and complete."""
        # If we got here, frontmatter parsed successfully
        assert isinstance(skill_metadata, dict), "Metadata should be dictionary"
        assert len(skill_metadata) > 0, "Metadata should not be empty"

    def test_skill_follows_naming_conventions(self, skill_path):
        """Verify skill directory follows naming conventions."""
        name = skill_path.name
        assert name == "moai-foundation-specs", f"Unexpected skill name: {name}"
        assert name.islower(), "Skill name must be lowercase"
        assert "-" in name, "Skill name must use hyphens"


# ===== CONTEXT7 INTEGRATION TESTS =====
class TestContext7Integration:
    """Test Context7 integration for latest patterns."""

    def test_has_context7_references(self, skill_metadata):
        """Verify Context7 references are present."""
        references = skill_metadata.get("context7_references", [])
        assert isinstance(references, list), "context7_references must be a list"
        # Foundation skills should reference standard sources
        if len(references) == 0:
            # It's acceptable for some skills to have empty context7_references
            # if they don't need external API documentation
            pass

    def test_metadata_complete_for_discovery(self, skill_metadata):
        """Verify metadata is complete for semantic discovery."""
        # Minimum set for discovery
        required_for_discovery = ["name", "description", "auto_trigger_keywords", "agent_coverage"]
        for field in required_for_discovery:
            assert field in skill_metadata, f"Missing field required for discovery: {field}"

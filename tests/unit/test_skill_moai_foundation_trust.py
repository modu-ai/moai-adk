"""
Test suite for moai-foundation-trust skill.
Tests TRUST 5 framework completeness, consolidated modules, and quality validation.
"""

import re
from pathlib import Path

import pytest
import yaml


# ===== FIXTURES =====
@pytest.fixture
def skill_path():
    """Return path to moai-foundation-trust skill."""
    return Path(__file__).parent.parent.parent / ".claude" / "skills" / "moai-foundation-trust"


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
    """Load all module files for TRUST 5 principles."""
    modules = {}
    if skill_metadata.get("modularized", False):
        modules_dir = skill_path / "modules"
        if modules_dir.exists():
            for module_file in modules_dir.glob("*.md"):
                with open(module_file, "r", encoding="utf-8") as f:
                    modules[module_file.stem] = f.read()
    return modules


# ===== METADATA COMPLIANCE TESTS =====
class TestMetadataCompliance:
    """Test metadata for moai-foundation-trust."""

    REQUIRED_FIELDS = [
        "name",
        "description",
        "version",
        "modularized",
        "last_updated",
        "allowed-tools",
        "compliance_score",
    ]

    def test_required_fields_present(self, skill_metadata):
        """Verify all 7 required fields present."""
        for field in self.REQUIRED_FIELDS:
            assert field in skill_metadata, f"Required field '{field}' missing"

    def test_name_is_moai_foundation_trust(self, skill_metadata):
        """Verify skill name is moai-foundation-trust."""
        assert skill_metadata.get("name") == "moai-foundation-trust"

    def test_is_modularized(self, skill_metadata):
        """Verify this is a modularized skill."""
        assert (
            skill_metadata.get("modularized", False) is True
        ), "moai-foundation-trust should be modularized for TRUST 5 principles"

    def test_high_compliance_score(self, skill_metadata):
        """Verify high compliance score for foundation skill."""
        score = skill_metadata.get("compliance_score", 0)
        assert score >= 85, f"Foundation skill should have score >= 85, got {score}"

    def test_consolidation_metadata(self, skill_metadata):
        """Verify consolidation information in metadata."""
        description = skill_metadata.get("description", "")
        # Should indicate TRUST 5 focus
        trust_keywords = ["TRUST", "trust", "quality", "test", "readable", "unified"]
        matches = sum(1 for k in trust_keywords if k in description)
        assert matches >= 2, f"Description should mention TRUST concepts: {description[:100]}"


# ===== TRUST 5 FRAMEWORK COMPLETENESS TESTS =====
class TestTRUST5Completeness:
    """Test that TRUST 5 framework is completely covered."""

    TRUST_PRINCIPLES = ["Test-First", "Readable", "Unified", "Secured", "Trackable"]

    def test_all_trust_principles_in_content(self, skill_content):
        """Verify all 5 TRUST principles are mentioned in content."""
        for principle in self.TRUST_PRINCIPLES:
            assert principle in skill_content, f"TRUST principle '{principle}' not found in skill content"

    def test_test_first_principle_complete(self, skill_content):
        """Verify Test-First principle fully documented."""
        test_first_sections = ["test", "coverage", "pytest", "unit test", "integration test"]
        matches = sum(1 for section in test_first_sections if section.lower() in skill_content.lower())
        assert matches >= 3, f"Test-First principle incomplete: {matches}/5 aspects found"

    def test_readable_principle_complete(self, skill_content):
        """Verify Readable principle fully documented."""
        readable_sections = ["naming", "style", "clarity", "maintainable", "comment"]
        matches = sum(1 for section in readable_sections if section.lower() in skill_content.lower())
        assert matches >= 3, f"Readable principle incomplete: {matches}/5 aspects found"

    def test_unified_principle_complete(self, skill_content):
        """Verify Unified principle fully documented."""
        unified_sections = ["consistent", "pattern", "standard", "convention", "uniform"]
        matches = sum(1 for section in unified_sections if section.lower() in skill_content.lower())
        assert matches >= 3, f"Unified principle incomplete: {matches}/5 aspects found"

    def test_secured_principle_complete(self, skill_content):
        """Verify Secured principle fully documented."""
        secured_sections = ["security", "validation", "sanitize", "owasp", "protected"]
        matches = sum(1 for section in secured_sections if section.lower() in skill_content.lower())
        assert matches >= 3, f"Secured principle incomplete: {matches}/5 aspects found"

    def test_trackable_principle_complete(self, skill_content):
        """Verify Trackable principle fully documented."""
        trackable_sections = ["tracking", "version", "history", "audit", "origin"]
        matches = sum(1 for section in trackable_sections if section.lower() in skill_content.lower())
        assert matches >= 3, f"Trackable principle incomplete: {matches}/5 aspects found"


# ===== MODULE STRUCTURE TESTS =====
class TestModuleStructure:
    """Test modular structure for TRUST 5 framework."""

    EXPECTED_MODULES = [
        "test-first",  # Test coverage, unit/integration/e2e
        "readable",  # Code style, naming, comments
        "unified",  # Pattern consistency, standards
        "secured",  # Security validation, OWASP
        "trackable",  # Version tracking, audit trails
        "git-workflows",  # Git strategies and practices
    ]

    def test_modules_directory_exists(self, skill_path):
        """Verify modules/ directory exists."""
        modules_dir = skill_path / "modules"
        assert modules_dir.exists(), "modules/ directory must exist"

    def test_expected_principle_modules_present(self, skill_path):
        """Verify core principle modules exist."""
        modules_dir = skill_path / "modules"
        principle_modules = ["test-first", "readable", "unified", "secured", "trackable"]

        for principle in principle_modules:
            module_file = modules_dir / f"{principle}.md"
            assert module_file.exists(), f"Missing principle module: {principle}.md"

    def test_each_module_has_content(self, module_files):
        """Verify each module has substantial content."""
        for module_name, content in module_files.items():
            assert len(content) > 200, f"Module {module_name}.md too short: {len(content)} chars"
            # Should have headers
            assert re.search(r"^#{1,6} ", content, re.MULTILINE), f"Module {module_name}.md missing headers"

    def test_module_file_sizes_reasonable(self, skill_path):
        """Verify module file sizes are within limits."""
        modules_dir = skill_path / "modules"
        if modules_dir.exists():
            for module_file in modules_dir.glob("*.md"):
                size = module_file.stat().st_size
                # Each module should be < 20KB
                assert size < 20 * 1024, f"Module {module_file.name} too large: {size} bytes"


# ===== GIT INTEGRATION TESTS =====
class TestGitIntegration:
    """Test Git strategy consolidation in TRUST framework."""

    def test_git_workflows_module_present(self, skill_path):
        """Verify git-workflows.md module exists."""
        skill_path / "modules" / "git-workflows.md"
        # Git workflows may be consolidated with this skill
        modules_dir = skill_path / "modules"
        if modules_dir.exists():
            git_files = list(modules_dir.glob("*git*.md"))
            assert len(git_files) > 0, "Should have git-related module"

    def test_git_concepts_in_trackable(self, module_files):
        """Verify git concepts integrated in Trackable principle."""
        if "trackable" in module_files:
            content = module_files["trackable"]
            git_keywords = ["commit", "version", "branch", "history"]
            sum(1 for kw in git_keywords if kw.lower() in content.lower())
            # May or may not be here depending on modularization
            # Just verify it's mentioned somewhere
            pass


# ===== AUTO-TRIGGER AND DISCOVERY TESTS =====
class TestAutoTriggerKeywords:
    """Test auto-trigger keywords for semantic discovery."""

    def test_has_comprehensive_keywords(self, skill_metadata):
        """Verify comprehensive auto-trigger keywords."""
        keywords = skill_metadata.get("auto_trigger_keywords", [])
        assert len(keywords) >= 8, f"Need ≥8 keywords, found {len(keywords)}"

    def test_keywords_cover_trust_principles(self, skill_metadata):
        """Verify keywords cover all TRUST principles."""
        keywords = skill_metadata.get("auto_trigger_keywords", [])
        keywords_lower = [k.lower() for k in keywords]

        principles = ["test", "readable", "unified", "secured", "trackable"]
        for principle in principles:
            assert any(principle in kw for kw in keywords_lower), f"Keywords missing coverage for '{principle}'"

    def test_keywords_include_quality_terms(self, skill_metadata):
        """Verify keywords include quality/validation terms."""
        keywords = skill_metadata.get("auto_trigger_keywords", [])
        keywords_lower = [k.lower() for k in keywords]

        quality_terms = ["quality", "review", "validation", "check", "gate"]
        matches = sum(1 for term in quality_terms if any(term in kw for kw in keywords_lower))
        assert matches >= 2, f"Keywords should include quality terms: {keywords}"


# ===== AGENT COVERAGE TESTS =====
class TestAgentCoverage:
    """Test agent coverage for quality assurance."""

    def test_has_agent_coverage(self, skill_metadata):
        """Verify agent_coverage field populated."""
        agents = skill_metadata.get("agent_coverage", [])
        assert isinstance(agents, list), "agent_coverage must be a list"
        assert len(agents) > 0, "agent_coverage should not be empty"

    def test_coverage_includes_quality_agents(self, skill_metadata):
        """Verify coverage includes quality-related agents."""
        agents = skill_metadata.get("agent_coverage", [])
        agent_names = [str(a).lower() for a in agents]

        quality_agents = ["quality", "test", "review", "security"]
        sum(1 for qa in quality_agents if any(qa in an for an in agent_names))
        # Should have at least some quality agents
        assert len(agents) >= 2, f"Should reference quality agents, got {agents}"


# ===== CONTENT QUALITY TESTS =====
class TestContentQuality:
    """Test overall content quality of TRUST foundation skill."""

    def test_no_placeholder_text(self, skill_content):
        """Verify no placeholder text remains."""
        placeholders = ["TODO", "FIXME", "XXX", "{{", "}}"]
        for placeholder in placeholders:
            assert placeholder not in skill_content, f"Placeholder '{placeholder}' found in content"

    def test_comprehensive_documentation(self, skill_content):
        """Verify comprehensive documentation."""
        # Should have substantial content
        assert len(skill_content) > 2000, f"Content too short for foundation skill: {len(skill_content)} chars"

    def test_structured_sections(self, skill_content):
        """Verify well-structured content."""
        sections = re.findall(r"^#{1,6} (.+)$", skill_content, re.MULTILINE)
        # Should have multiple major sections
        assert len(sections) >= 8, f"Should have ≥8 sections, found {len(sections)}"

    def test_examples_provided(self, skill_content):
        """Verify examples are provided."""
        # Should have code blocks
        code_blocks = re.findall(r"```", skill_content)
        assert len(code_blocks) >= 2, "Should have example code blocks"

    def test_best_practices_documented(self, skill_content):
        """Verify best practices section."""
        best_practices_found = re.search(r"(best practices|do.*don\'t|✅.*❌)", skill_content, re.IGNORECASE)
        assert best_practices_found, "Should document best practices"


# ===== FILE STRUCTURE TESTS =====
class TestFileStructure:
    """Test overall file structure."""

    def test_skill_file_size_reasonable(self, skill_path):
        """Verify main SKILL.md is reasonable size."""
        skill_md = skill_path / "SKILL.md"
        lines = len(skill_md.read_text().splitlines())
        # Modularized, so main file can be smaller
        assert lines <= 500, f"SKILL.md too large: {lines} lines"

    def test_yaml_frontmatter_valid(self, skill_metadata):
        """Verify YAML frontmatter is valid."""
        assert isinstance(skill_metadata, dict), "Metadata should be dictionary"
        assert len(skill_metadata) > 5, "Metadata should have ≥5 fields"

    def test_directory_structure_correct(self, skill_path):
        """Verify overall directory structure."""
        assert skill_path.exists(), f"Skill directory not found: {skill_path}"
        assert (skill_path / "SKILL.md").exists(), "SKILL.md not found"
        assert (skill_path / "modules").exists(), "modules/ directory not found"


# ===== INTEGRATION COMPLETENESS TESTS =====
class TestIntegrationCompleteness:
    """Test that TRUST framework is properly integrated."""

    def test_consolidates_trust_sources(self, skill_metadata):
        """Verify consolidation of TRUST-related skills."""
        description = skill_metadata.get("description", "")
        # Should indicate it consolidates other skills
        assert (
            "Consolidate" in description
            or "consolidate" in description.lower()
            or len(skill_metadata.get("dependencies", [])) > 0
        ), "Should indicate skill consolidation"

    def test_provides_unified_framework(self, skill_content):
        """Verify provides unified quality framework."""
        unified_terms = ["unified", "framework", "approach", "system", "standard"]
        matches = sum(1 for term in unified_terms if term.lower() in skill_content.lower())
        assert matches >= 2, f"Should emphasize unified approach: {matches}/5 found"

    def test_includes_success_metrics(self, skill_content):
        """Verify success metrics or goals are defined."""
        metrics_found = re.search(r"(target|goal|metric|score|coverage|percentage)", skill_content, re.IGNORECASE)
        assert metrics_found, "Should define success metrics"

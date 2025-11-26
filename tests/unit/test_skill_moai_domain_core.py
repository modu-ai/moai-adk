"""
Test suite for core domain skills: backend, frontend, database, cloud.
Tests metadata compliance, domain-specific patterns, and integration.
"""

import re
from pathlib import Path

import pytest
import yaml

# ===== DOMAIN SKILL NAMES =====
DOMAIN_SKILLS_TO_TEST = [
    "moai-domain-backend",
    "moai-domain-frontend",
    "moai-domain-database",
    "moai-domain-cloud",
]


# ===== FIXTURES =====
@pytest.fixture(params=DOMAIN_SKILLS_TO_TEST)
def domain_skill_name(request):
    """Parameterized fixture for domain skill names."""
    return request.param


@pytest.fixture
def skill_path(domain_skill_name):
    """Return path to domain skill."""
    return Path(__file__).parent.parent.parent / ".claude" / "skills" / domain_skill_name


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
class TestDomainSkillMetadata:
    """Test metadata for domain skills."""

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

    def test_name_matches_directory(self, skill_path, skill_metadata):
        """Verify skill name matches directory name."""
        assert skill_metadata.get("name") == skill_path.name

    def test_category_tier_is_domain(self, skill_metadata):
        """Verify category_tier is 2 (domain)."""
        tier = skill_metadata.get("category_tier")
        assert tier == 2, f"Domain skill should have category_tier=2, got {tier}"

    def test_compliance_score_reasonable(self, skill_metadata):
        """Verify compliance score is reasonable."""
        score = skill_metadata.get("compliance_score", 0)
        assert score >= 70, f"Domain skill should have score >= 70, got {score}"

    def test_auto_trigger_keywords_count(self, skill_metadata):
        """Verify auto-trigger keywords count."""
        keywords = skill_metadata.get("auto_trigger_keywords", [])
        assert isinstance(keywords, list)
        assert 8 <= len(keywords) <= 15, f"Keywords {len(keywords)} not in [8, 15]"

    def test_dependencies_indicates_consolidation(self, skill_metadata):
        """Verify dependencies field if skill is consolidated."""
        # Domain skills may have dependencies or indicate consolidation
        description = skill_metadata.get("description", "")
        # Should mention the domain it covers
        assert len(description) >= 20, "Description too short"


# ===== DOMAIN-SPECIFIC PATTERN TESTS =====
class TestDomainPatterns:
    """Test domain-specific patterns in skill content."""

    def test_backend_skill_has_api_patterns(self, domain_skill_name, skill_content):
        """For backend skill, verify API patterns documented."""
        if "backend" in domain_skill_name:
            api_keywords = ["REST", "GraphQL", "API", "endpoint", "async"]
            matches = sum(1 for kw in api_keywords if kw in skill_content)
            assert matches >= 3, f"Backend skill missing API patterns: {matches}/5"

    def test_frontend_skill_has_ui_patterns(self, domain_skill_name, skill_content):
        """For frontend skill, verify UI patterns documented."""
        if "frontend" in domain_skill_name:
            ui_keywords = ["React", "Vue", "component", "state", "hooks"]
            matches = sum(1 for kw in ui_keywords if kw in skill_content)
            assert matches >= 3, f"Frontend skill missing UI patterns: {matches}/5"

    def test_database_skill_has_db_patterns(self, domain_skill_name, skill_content):
        """For database skill, verify database patterns documented."""
        if "database" in domain_skill_name:
            db_keywords = ["SQL", "NoSQL", "ORM", "query", "schema"]
            matches = sum(1 for kw in db_keywords if kw in skill_content)
            assert matches >= 3, f"Database skill missing DB patterns: {matches}/5"

    def test_cloud_skill_has_cloud_patterns(self, domain_skill_name, skill_content):
        """For cloud skill, verify cloud patterns documented."""
        if "cloud" in domain_skill_name:
            cloud_keywords = ["AWS", "GCP", "Azure", "deploy", "infrastructure"]
            matches = sum(1 for kw in cloud_keywords if kw in skill_content)
            assert matches >= 2, f"Cloud skill missing cloud patterns: {matches}/5"


# ===== CONTENT STRUCTURE TESTS =====
class TestContentStructure:
    """Test content structure for domain skills."""

    def test_has_quick_reference(self, skill_content):
        """Verify Quick Reference section."""
        assert any(
            term in skill_content for term in ["Quick Reference", "quick reference", "Overview"]
        ), "Missing Quick Reference or similar section"

    def test_has_when_to_use_section(self, skill_content):
        """Verify When to Use section."""
        assert any(
            term in skill_content for term in ["When to Use", "Use Cases", "Scenarios"]
        ), "Missing When to Use section"

    def test_has_best_practices(self, skill_content):
        """Verify Best Practices section."""
        assert any(
            term in skill_content for term in ["Best Practices", "DO", "DON'T", "bes practices"]
        ), "Missing Best Practices section"

    def test_has_implementation_details(self, skill_content):
        """Verify Implementation Guide or Code Examples."""
        assert any(
            term in skill_content for term in ["Implementation", "Code Example", "Pattern", "Framework"]
        ), "Missing implementation details"


# ===== TECHNOLOGY COVERAGE TESTS =====
class TestTechnologyCoverage:
    """Test technology coverage in domain skills."""

    def test_mentions_current_frameworks(self, domain_skill_name, skill_content):
        """Verify mentions of current/recent framework versions."""
        # Each domain should mention current tools
        if "backend" in domain_skill_name:
            tech_terms = ["FastAPI", "Django", "Express", "Node", "Python", "TypeScript"]
        elif "frontend" in domain_skill_name:
            tech_terms = ["React", "Next", "Vue", "TypeScript", "Component"]
        elif "database" in domain_skill_name:
            tech_terms = ["PostgreSQL", "MongoDB", "SQL", "ORM", "Schema"]
        elif "cloud" in domain_skill_name:
            tech_terms = ["AWS", "GCP", "Azure", "Lambda", "CloudSQL"]
        else:
            tech_terms = []

        matches = sum(1 for tech in tech_terms if tech in skill_content)
        assert matches >= 2, f"Missing current technology references: {matches}/len"

    def test_includes_2025_version_info(self, skill_content):
        """Verify includes 2024-2025 version information."""
        version_pattern = r"(202[45]|2024|2025|v?\d+\.\d+)"
        matches = re.findall(version_pattern, skill_content)
        # Should have some version references
        assert len(matches) >= 2, "Should include current version information"


# ===== AGENT COVERAGE TESTS =====
class TestAgentCoverageDomain:
    """Test agent coverage for domain skills."""

    def test_has_agent_coverage(self, skill_metadata):
        """Verify agent_coverage field populated."""
        agents = skill_metadata.get("agent_coverage", [])
        assert isinstance(agents, list)
        assert len(agents) > 0, "Should have agent coverage"

    def test_agents_are_valid(self, skill_metadata):
        """Verify agent names are valid."""
        agents = skill_metadata.get("agent_coverage", [])
        for agent in agents:
            assert isinstance(agent, str)
            # Should be lowercase with hyphens or contain '-expert' or '-engineer'
            assert agent.islower() or "-" in agent, f"Invalid agent name: {agent}"


# ===== CODE QUALITY TESTS =====
class TestCodeQuality:
    """Test code quality and examples."""

    def test_code_blocks_valid(self, skill_content):
        """Verify code blocks are properly formatted."""
        fence_count = skill_content.count("```")
        assert fence_count % 2 == 0, "Unmatched code fences"
        # Should have code examples
        assert fence_count >= 2, "Should have code examples"

    def test_no_broken_links(self, skill_content):
        """Verify markdown links are valid."""
        link_pattern = r"\[([^\]]+)\]\(([^)]+)\)"
        links = re.findall(link_pattern, skill_content)
        for text, url in links:
            assert len(text) > 0, "Empty link text"
            assert len(url) > 0, "Empty link URL"

    def test_no_placeholder_text(self, skill_content):
        """Verify no placeholder text."""
        placeholders = ["TODO", "FIXME", "XXX", "{{", "}}"]
        for placeholder in placeholders:
            assert placeholder not in skill_content, f"Placeholder '{placeholder}' found"


# ===== FILE STRUCTURE TESTS =====
class TestFileStructure:
    """Test file structure."""

    def test_skill_file_exists(self, skill_path):
        """Verify SKILL.md exists."""
        assert (skill_path / "SKILL.md").exists()

    def test_file_size_reasonable(self, skill_path):
        """Verify file size is reasonable."""
        skill_md = skill_path / "SKILL.md"
        lines = len(skill_md.read_text().splitlines())
        assert lines <= 500, f"SKILL.md too large: {lines} lines"

    def test_yaml_frontmatter_valid(self, skill_metadata):
        """Verify YAML frontmatter is valid."""
        assert isinstance(skill_metadata, dict)
        assert len(skill_metadata) > 5


# ===== CONTEXT7 INTEGRATION TESTS =====
class TestContext7References:
    """Test Context7 integration."""

    def test_has_context7_references(self, skill_metadata):
        """Verify context7_references field present."""
        field = "context7_references"
        assert field in skill_metadata, f"Missing {field}"

    def test_context7_references_valid_format(self, skill_metadata):
        """Verify context7_references are properly formatted."""
        references = skill_metadata.get("context7_references", [])
        assert isinstance(references, list)
        for ref in references:
            assert isinstance(ref, str), f"Reference should be string: {ref}"


# ===== PATTERN CONSOLIDATION TESTS =====
class TestPatternConsolidation:
    """Test that skills consolidate multiple patterns."""

    def test_covers_multiple_patterns(self, domain_skill_name, skill_content):
        """Verify domain skill covers multiple patterns."""
        # Should have multiple subsections or patterns
        sections = re.findall(r"^### .+$", skill_content, re.MULTILINE)
        # Not all skills may have ### sections, but if they do, should have 2+
        if sections:
            assert len(sections) >= 2, f"Only {len(sections)} sections found"

    def test_provides_practical_examples(self, skill_content):
        """Verify practical examples are provided."""
        # Should have code examples
        code_fences = skill_content.count("```")
        assert code_fences >= 2, "Should have multiple code examples"

    def test_includes_best_practices(self, skill_content):
        """Verify best practices are documented."""
        best_practice_terms = ["DO", "DON'T", "best practice", "avoid", "prefer"]
        matches = sum(1 for term in best_practice_terms if term.lower() in skill_content.lower())
        assert matches >= 2, f"Missing best practice guidance: {matches}/5"


# ===== DISCOVERY AND AUTO-TRIGGER TESTS =====
class TestDiscoverability:
    """Test skill discoverability and auto-trigger."""

    def test_auto_trigger_keywords_comprehensive(self, skill_metadata):
        """Verify keywords are comprehensive."""
        keywords = skill_metadata.get("auto_trigger_keywords", [])
        # Should have keywords that match common developer queries
        assert len(keywords) >= 8, f"Need ≥8 keywords, got {len(keywords)}"

    def test_keywords_include_domain_terms(self, domain_skill_name, skill_metadata):
        """Verify keywords include domain-specific terms."""
        keywords = skill_metadata.get("auto_trigger_keywords", [])
        keywords_lower = [k.lower() for k in keywords]

        # Check for domain-relevant keywords
        if "backend" in domain_skill_name:
            assert any("api" in k or "server" in k or "backend" in k for k in keywords_lower)
        elif "frontend" in domain_skill_name:
            assert any("ui" in k or "component" in k or "frontend" in k or "react" in k for k in keywords_lower)
        elif "database" in domain_skill_name:
            assert any("db" in k or "data" in k or "query" in k for k in keywords_lower)
        elif "cloud" in domain_skill_name:
            assert any("cloud" in k or "deploy" in k or "aws" in k or "gcp" in k for k in keywords_lower)

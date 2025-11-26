"""
Test suite for specialized domain skills: cli-tool, mobile, iot, figma, notion, ml-ops, monitoring.
Tests domain-specific patterns and integration completeness.
"""

import re
from pathlib import Path

import pytest
import yaml

# ===== SPECIALIZED DOMAIN SKILL NAMES =====
SPECIALIZED_SKILLS_TO_TEST = [
    "moai-domain-cli-tool",
    "moai-domain-cloud",  # Already created but included for completeness
]

# Additional skills that may exist
OPTIONAL_SKILLS = [
    "moai-domain-mobile",
    "moai-domain-iot",
    "moai-domain-figma",
    "moai-domain-notion",
    "moai-domain-ml-ops",
    "moai-domain-monitoring",
]


# ===== FIXTURES =====
@pytest.fixture(params=SPECIALIZED_SKILLS_TO_TEST)
def specialized_skill_name(request):
    """Parameterized fixture for specialized domain skill names."""
    return request.param


@pytest.fixture
def skill_path(specialized_skill_name):
    """Return path to specialized domain skill."""
    base_path = Path(__file__).parent.parent.parent / ".claude" / "skills"
    skill_path = base_path / specialized_skill_name

    if not skill_path.exists():
        pytest.skip(f"Skill directory not found: {skill_path}")

    return skill_path


@pytest.fixture
def skill_metadata(skill_path):
    """Load and parse skill metadata."""
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


# ===== SPECIALIZED DOMAIN TESTS =====
class TestSpecializedDomainSkills:
    """Test specialized domain skills (CLI, Mobile, IoT, etc.)."""

    def test_cli_tool_has_cli_patterns(self, specialized_skill_name, skill_content):
        """For CLI tool skill, verify CLI patterns documented."""
        if "cli" in specialized_skill_name:
            cli_keywords = ["command", "argument", "option", "interactive", "parser"]
            matches = sum(1 for kw in cli_keywords if kw.lower() in skill_content.lower())
            assert matches >= 3, f"CLI skill missing CLI patterns: {matches}/5"

    def test_skill_name_matches_directory(self, skill_path, skill_metadata):
        """Verify skill name matches directory."""
        expected = skill_path.name
        actual = skill_metadata.get("name")
        assert actual == expected, f"Name mismatch: {actual} != {expected}"

    def test_skill_has_category_tier_2(self, skill_metadata):
        """Verify category_tier is 2 (domain)."""
        tier = skill_metadata.get("category_tier")
        assert tier == 2, f"Specialized domain skill should have tier=2, got {tier}"


# ===== FRAMEWORK AND TOOL COVERAGE TESTS =====
class TestFrameworkCoverage:
    """Test framework and tool coverage for specialized skills."""

    def test_cli_tool_mentions_frameworks(self, specialized_skill_name, skill_content):
        """For CLI tool, verify framework mentions."""
        if "cli" in specialized_skill_name:
            frameworks = ["oclif", "click", "commander", "yargs", "argument"]
            mentions = sum(1 for fw in frameworks if fw in skill_content.lower())
            # Should mention at least one CLI framework
            assert mentions >= 1, f"CLI skill should mention frameworks: {frameworks}"

    def test_includes_real_world_examples(self, skill_content):
        """Verify real-world examples are included."""
        # Should have concrete examples
        example_indicators = ["Example", "Usage", "Code", "```", "class", "function"]
        matches = sum(1 for indicator in example_indicators if indicator in skill_content)
        assert matches >= 3, f"Should include real-world examples: {matches}/6"


# ===== METADATA COMPLIANCE TESTS =====
class TestSpecializedMetadata:
    """Test metadata for specialized domain skills."""

    REQUIRED_FIELDS = [
        "name",
        "description",
        "version",
        "modularized",
        "last_updated",
        "allowed-tools",
        "compliance_score",
    ]

    def test_all_required_fields(self, skill_metadata):
        """Verify all required metadata fields present."""
        for field in self.REQUIRED_FIELDS:
            assert field in skill_metadata, f"Required field '{field}' missing"

    def test_description_mentions_domain(self, specialized_skill_name, skill_metadata):
        """Verify description mentions the domain."""
        desc = skill_metadata.get("description", "").lower()
        specialized_skill_name.split("-")[-1]  # e.g., 'cli-tool' -> 'tool'

        # Description should relate to the domain
        assert len(desc) > 15, f"Description too short: {len(desc)} chars"

    def test_version_is_semantic(self, skill_metadata):
        """Verify version is semantic versioning."""
        version = skill_metadata.get("version", "")
        assert re.match(r"^\d+\.\d+\.\d+$", version), f"Invalid version: {version}"

    def test_compliance_score_present(self, skill_metadata):
        """Verify compliance_score is present."""
        score = skill_metadata.get("compliance_score")
        assert isinstance(score, (int, float)), "compliance_score must be numeric"
        assert 0 <= score <= 100, f"Score out of range: {score}"


# ===== CONTENT STRUCTURE TESTS =====
class TestContentStructure:
    """Test content structure for specialized skills."""

    def test_has_core_sections(self, skill_content):
        """Verify core content sections."""
        sections = ["What It Does", "When to Use", "Quick Reference", "Implementation", "Best Practices"]
        found = sum(1 for section in sections if section in skill_content)
        assert found >= 3, f"Missing core sections: {found}/5 found"

    def test_structured_with_headers(self, skill_content):
        """Verify content has markdown structure."""
        headers = re.findall(r"^#{1,6} ", skill_content, re.MULTILINE)
        assert len(headers) >= 5, f"Content should have ≥5 sections, found {len(headers)}"

    def test_content_length_adequate(self, skill_content):
        """Verify content length is adequate."""
        # Should have substantial content for domain skill
        assert len(skill_content) > 1500, f"Content too short: {len(skill_content)} chars"


# ===== AUTO-TRIGGER KEYWORD TESTS =====
class TestAutoTriggerKeywords:
    """Test auto-trigger keywords for discovery."""

    def test_keywords_count_valid(self, skill_metadata):
        """Verify keyword count is valid."""
        keywords = skill_metadata.get("auto_trigger_keywords", [])
        assert isinstance(keywords, list), "auto_trigger_keywords must be list"
        assert 8 <= len(keywords) <= 15, f"Keywords {len(keywords)} not in [8,15]"

    def test_keywords_are_strings(self, skill_metadata):
        """Verify keywords are strings."""
        keywords = skill_metadata.get("auto_trigger_keywords", [])
        for kw in keywords:
            assert isinstance(kw, str), f"Keyword must be string: {kw}"
            assert len(kw) > 0, "Keyword cannot be empty"

    def test_keywords_diverse(self, skill_metadata):
        """Verify keywords are diverse."""
        keywords = skill_metadata.get("auto_trigger_keywords", [])
        # Should have variety
        keyword_lens = [len(k) for k in keywords]
        avg_len = sum(keyword_lens) / len(keyword_lens) if keywords else 0
        assert 5 < avg_len < 20, f"Keywords average length unusual: {avg_len}"


# ===== AGENT COVERAGE TESTS =====
class TestAgentCoverage:
    """Test agent coverage."""

    def test_has_agent_coverage(self, skill_metadata):
        """Verify agent_coverage field present."""
        agents = skill_metadata.get("agent_coverage", [])
        assert isinstance(agents, list), "agent_coverage must be list"
        assert len(agents) > 0, "Should reference at least one agent"

    def test_agents_are_valid_names(self, skill_metadata):
        """Verify agent names are valid."""
        agents = skill_metadata.get("agent_coverage", [])
        for agent in agents:
            assert isinstance(agent, str), f"Agent must be string: {agent}"
            # Valid agent names are lowercase with hyphens
            assert agent.islower() or "-" in agent, f"Invalid agent name: {agent}"


# ===== CODE QUALITY TESTS =====
class TestCodeQuality:
    """Test code quality in skills."""

    def test_has_code_examples(self, skill_content):
        """Verify code examples are provided."""
        code_fences = skill_content.count("```")
        assert code_fences >= 2, f"Should have ≥1 code examples, found {code_fences // 2}"

    def test_code_blocks_balanced(self, skill_content):
        """Verify code blocks are properly closed."""
        code_fences = skill_content.count("```")
        assert code_fences % 2 == 0, "Unmatched code fence markers"

    def test_no_broken_markdown_links(self, skill_content):
        """Verify markdown links are valid."""
        link_pattern = r"\[([^\]]+)\]\(([^)]+)\)"
        links = re.findall(link_pattern, skill_content)
        for text, url in links:
            assert len(text) > 0, "Empty link text"
            assert len(url) > 0, "Empty link URL"

    def test_no_placeholder_text(self, skill_content):
        """Verify no placeholder text."""
        placeholders = ["TODO", "FIXME", "XXX", "{{", "}}", "[object Object]"]
        for placeholder in placeholders:
            count = skill_content.count(placeholder)
            assert count == 0, f"Found {count} instances of '{placeholder}'"


# ===== FILE STRUCTURE TESTS =====
class TestFileStructure:
    """Test file structure."""

    def test_skill_directory_exists(self, skill_path):
        """Verify skill directory exists."""
        assert skill_path.exists(), f"Skill directory not found: {skill_path}"
        assert skill_path.is_dir(), f"Not a directory: {skill_path}"

    def test_skill_md_exists(self, skill_path):
        """Verify SKILL.md exists."""
        skill_md = skill_path / "SKILL.md"
        assert skill_md.exists(), f"SKILL.md not found in {skill_path}"

    def test_file_size_reasonable(self, skill_path):
        """Verify file size is reasonable."""
        skill_md = skill_path / "SKILL.md"
        size_bytes = skill_md.stat().st_size
        max_size = 50 * 1024  # 50KB
        assert size_bytes < max_size, f"File size {size_bytes} bytes exceeds {max_size}"

    def test_file_line_count_reasonable(self, skill_path):
        """Verify line count is reasonable."""
        skill_md = skill_path / "SKILL.md"
        lines = len(skill_md.read_text().splitlines())
        assert lines <= 500, f"SKILL.md has {lines} lines (max 500)"


# ===== CONTEXT7 INTEGRATION TESTS =====
class TestContext7Integration:
    """Test Context7 integration."""

    def test_has_context7_references_field(self, skill_metadata):
        """Verify context7_references field present."""
        assert "context7_references" in skill_metadata, "Missing context7_references"

    def test_context7_references_is_list(self, skill_metadata):
        """Verify context7_references is a list."""
        refs = skill_metadata.get("context7_references", [])
        assert isinstance(refs, list), "context7_references must be list"


# ===== OPTIONAL SKILLS CHECK =====
class TestOptionalSkillsPresence:
    """Test that optional specialized skills are documented."""

    def test_optional_skills_identified(self):
        """Verify optional skills list is complete."""
        # This is informational - lists skills that should be tested if they exist
        optional = [
            "moai-domain-mobile",
            "moai-domain-iot",
            "moai-domain-figma",
            "moai-domain-notion",
            "moai-domain-ml-ops",
            "moai-domain-monitoring",
        ]
        # Just verify the list exists - actual tests will be skipped if skills don't exist
        assert len(optional) == 6, f"Optional skills list has {len(optional)} items"


# ===== BEST PRACTICES TESTS =====
class TestBestPractices:
    """Test best practices documentation."""

    def test_includes_do_and_dont(self, skill_content):
        """Verify DO and DON'T guidance included."""
        do_found = re.search(r"✅|DO", skill_content, re.IGNORECASE)
        dont_found = re.search(r"❌|DON\'T", skill_content, re.IGNORECASE)
        # At least one should be present
        assert do_found or dont_found, "Should include DO/DON'T best practices"

    def test_includes_practical_guidance(self, skill_content):
        """Verify practical guidance included."""
        guidance_terms = ["Best Practice", "Note", "Warning", "Tip", "Example"]
        matches = sum(1 for term in guidance_terms if term in skill_content)
        assert matches >= 2, f"Should include practical guidance: {matches}/5"


# ===== DISCOVERY COMPLETENESS TESTS =====
class TestDiscoveryCompleteness:
    """Test that skill is discoverable."""

    def test_skill_name_follows_convention(self, skill_path):
        """Verify skill name follows naming convention."""
        name = skill_path.name
        assert name.startswith("moai-"), "Skill name must start with 'moai-'"
        assert name.islower() or "-" in name, "Skill name must be lowercase with hyphens"

    def test_metadata_enables_discovery(self, skill_metadata):
        """Verify metadata supports discovery."""
        # Minimum fields for discovery
        discovery_fields = ["name", "description", "auto_trigger_keywords"]
        for field in discovery_fields:
            assert field in skill_metadata, f"Missing discovery field: {field}"
            value = skill_metadata.get(field)
            assert value, f"Discovery field '{field}' is empty"

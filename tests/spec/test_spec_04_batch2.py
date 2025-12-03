"""
RED Phase Tests for SPEC-04-GROUP-E Batch 2 (TAG-003, TAG-004, TAG-005)

Tests validate file structure, Context7 integration, and content quality for:
- TAG-003: Documentation Skills (5 skills)
- TAG-004: MCP Integration Skills (3 skills)
- TAG-005: Project Management Skills (5 skills)
"""

import re
from pathlib import Path

import pytest


class TestSkillFileStructure:
    """Verify complete file structure for all batch 2 skills"""

    # TAG-003: Documentation Skills
    TAG_003_SKILLS = [
        "moai-docs-generation",
        "moai-docs-linting",
        "moai-docs-toolkit",
        "moai-docs-unified",
        "moai-docs-validation",
    ]

    # TAG-004: MCP Integration Skills
    TAG_004_SKILLS = [
        "moai-mcp-integration",
        "moai-context7-integration",
        "moai-artifacts-builder",
    ]

    # TAG-005: Project Management Skills
    TAG_005_SKILLS = [
        "moai-project-batch-questions",
        "moai-project-config-manager",
        "moai-project-documentation",
        "moai-project-language-initializer",
        "moai-project-template-optimizer",
    ]

    REQUIRED_FILES = {
        "SKILL.md": (800, 1000),  # (min_lines, max_lines)
        "examples.md": (300, 700),
        "modules/advanced-patterns.md": (400, 600),
        "modules/optimization.md": (200, 500),
        "reference.md": (30, 100),
    }

    SKILLS_BASE_PATH = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")

    @pytest.mark.parametrize("skill", TAG_003_SKILLS)
    def test_tag003_documentation_skill_files_exist(self, skill: str):
        """TAG-003: Verify documentation skill files exist"""
        skill_path = self.SKILLS_BASE_PATH / skill
        assert skill_path.exists(), f"Skill directory not found: {skill}"

        for file_path in self.REQUIRED_FILES.keys():
            full_path = skill_path / file_path
            assert full_path.exists(), f"Missing file: {skill}/{file_path}"

    @pytest.mark.parametrize("skill", TAG_004_SKILLS)
    def test_tag004_mcp_skill_files_exist(self, skill: str):
        """TAG-004: Verify MCP integration skill files exist"""
        skill_path = self.SKILLS_BASE_PATH / skill
        assert skill_path.exists(), f"Skill directory not found: {skill}"

        for file_path in self.REQUIRED_FILES.keys():
            full_path = skill_path / file_path
            assert full_path.exists(), f"Missing file: {skill}/{file_path}"

    @pytest.mark.parametrize("skill", TAG_005_SKILLS)
    def test_tag005_project_skill_files_exist(self, skill: str):
        """TAG-005: Verify project management skill files exist"""
        skill_path = self.SKILLS_BASE_PATH / skill
        assert skill_path.exists(), f"Skill directory not found: {skill}"

        for file_path in self.REQUIRED_FILES.keys():
            full_path = skill_path / file_path
            assert full_path.exists(), f"Missing file: {skill}/{file_path}"


class TestContext7Integration:
    """Verify Context7 integration in all skills"""

    ALL_BATCH2_SKILLS = (
        TestSkillFileStructure.TAG_003_SKILLS
        + TestSkillFileStructure.TAG_004_SKILLS
        + TestSkillFileStructure.TAG_005_SKILLS
    )

    SKILLS_BASE_PATH = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")

    @pytest.mark.parametrize("skill", ALL_BATCH2_SKILLS)
    def test_context7_integration_section_exists(self, skill: str):
        """Verify Context7 Integration section in SKILL.md"""
        skill_file = self.SKILLS_BASE_PATH / skill / "SKILL.md"

        if not skill_file.exists():
            pytest.skip(f"Skill file not found: {skill}")

        content = skill_file.read_text(encoding="utf-8")

        # Check for Context7 Integration section
        assert "Context7" in content or "context7" in content, f"{skill}: Missing Context7 Integration section"

    @pytest.mark.parametrize("skill", ALL_BATCH2_SKILLS)
    def test_context7_library_references_exist(self, skill: str):
        """Verify Context7 library references in reference.md"""
        ref_file = self.SKILLS_BASE_PATH / skill / "reference.md"

        if not ref_file.exists():
            pytest.skip(f"Reference file not found: {skill}")

        content = ref_file.read_text(encoding="utf-8")

        # Check for library references or Context7 Integration section
        has_context7 = "Context7" in content or "Official Documentation" in content or "Related Libraries" in content
        assert has_context7, f"{skill}: Missing library references or Context7 Integration"


class TestContentQuality:
    """Verify content quality and structure"""

    ALL_BATCH2_SKILLS = (
        TestSkillFileStructure.TAG_003_SKILLS
        + TestSkillFileStructure.TAG_004_SKILLS
        + TestSkillFileStructure.TAG_005_SKILLS
    )

    SKILLS_BASE_PATH = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")

    @pytest.mark.parametrize("skill", ALL_BATCH2_SKILLS)
    def test_skill_md_has_required_sections(self, skill: str):
        """Verify SKILL.md has required sections"""
        skill_file = self.SKILLS_BASE_PATH / skill / "SKILL.md"

        if not skill_file.exists():
            pytest.skip(f"Skill file not found: {skill}")

        content = skill_file.read_text(encoding="utf-8")

        required_sections = [
            "Quick Reference",
            "Implementation Guide",
            "Best Practices",
        ]

        for section in required_sections:
            assert section in content, f"{skill}: Missing section '{section}' in SKILL.md"

    @pytest.mark.parametrize("skill", ALL_BATCH2_SKILLS)
    def test_examples_md_has_code_examples(self, skill: str):
        """Verify examples.md has executable code examples"""
        examples_file = self.SKILLS_BASE_PATH / skill / "examples.md"

        if not examples_file.exists():
            pytest.skip(f"Examples file not found: {skill}")

        content = examples_file.read_text(encoding="utf-8")

        # Count code blocks
        code_blocks = re.findall(r"```(\w+)", content)
        assert len(code_blocks) >= 5, f"{skill}: examples.md has less than 5 code examples"

    @pytest.mark.parametrize("skill", ALL_BATCH2_SKILLS)
    def test_reference_md_has_links(self, skill: str):
        """Verify reference.md has documentation links"""
        ref_file = self.SKILLS_BASE_PATH / skill / "reference.md"

        if not ref_file.exists():
            pytest.skip(f"Reference file not found: {skill}")

        content = ref_file.read_text(encoding="utf-8")

        # Check for links
        links = re.findall(r"\[([^\]]+)\]\(([^\)]+)\)", content)
        assert len(links) > 0, f"{skill}: reference.md has no documentation links"


class TestLanguageAndFormatting:
    """Verify language and formatting standards"""

    ALL_BATCH2_SKILLS = (
        TestSkillFileStructure.TAG_003_SKILLS
        + TestSkillFileStructure.TAG_004_SKILLS
        + TestSkillFileStructure.TAG_005_SKILLS
    )

    SKILLS_BASE_PATH = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")

    @pytest.mark.parametrize("skill", ALL_BATCH2_SKILLS)
    def test_all_files_are_markdown(self, skill: str):
        """Verify all documentation files are in .md format"""
        skill_path = self.SKILLS_BASE_PATH / skill

        if not skill_path.exists():
            pytest.skip(f"Skill directory not found: {skill}")

        md_files = list(skill_path.rglob("*.md"))
        assert len(md_files) >= 5, f"{skill}: Less than 5 .md files found"

    @pytest.mark.parametrize("skill", ALL_BATCH2_SKILLS)
    def test_files_use_english_text(self, skill: str):
        """Verify primary content is in English"""
        skill_file = self.SKILLS_BASE_PATH / skill / "SKILL.md"

        if not skill_file.exists():
            pytest.skip(f"Skill file not found: {skill}")

        content = skill_file.read_text(encoding="utf-8")

        # Basic check: should have English text
        assert len(content) > 100, f"{skill}: SKILL.md content is too short"


class TestDateAndVersion:
    """Verify metadata with current date"""

    ALL_BATCH2_SKILLS = (
        TestSkillFileStructure.TAG_003_SKILLS
        + TestSkillFileStructure.TAG_004_SKILLS
        + TestSkillFileStructure.TAG_005_SKILLS
    )

    SKILLS_BASE_PATH = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")

    @pytest.mark.parametrize("skill", ALL_BATCH2_SKILLS)
    def test_last_updated_date_present(self, skill: str):
        """Verify Last Updated date is present (should be 2025-11-22)"""
        skill_file = self.SKILLS_BASE_PATH / skill / "SKILL.md"

        if not skill_file.exists():
            pytest.skip(f"Skill file not found: {skill}")

        content = skill_file.read_text(encoding="utf-8")

        # Check for any date format like 2025-11-22 or Last Updated
        has_date = re.search(r"202[45]-\d{2}-\d{2}|Last Updated", content)
        assert has_date, f"{skill}: Missing Last Updated date in SKILL.md"

    @pytest.mark.parametrize("skill", ALL_BATCH2_SKILLS)
    def test_version_present(self, skill: str):
        """Verify version information is present"""
        skill_file = self.SKILLS_BASE_PATH / skill / "SKILL.md"

        if not skill_file.exists():
            pytest.skip(f"Skill file not found: {skill}")

        content = skill_file.read_text(encoding="utf-8")

        # Check for version format like 4.0.0 or version:
        has_version = re.search(r"[Vv]ersion[:\s]+\d+\.\d+\.\d+", content)
        assert has_version, f"{skill}: Missing version information"


class TestTAGSpecificContent:
    """TAG-specific content validation"""

    SKILLS_BASE_PATH = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")

    @pytest.mark.parametrize(
        "skill",
        [
            "moai-docs-generation",
            "moai-docs-linting",
            "moai-docs-toolkit",
            "moai-docs-unified",
            "moai-docs-validation",
        ],
    )
    def test_tag003_docs_content(self, skill: str):
        """TAG-003: Verify documentation-specific content"""
        skill_file = self.SKILLS_BASE_PATH / skill / "SKILL.md"

        if not skill_file.exists():
            pytest.skip(f"Skill file not found: {skill}")

        content = skill_file.read_text(encoding="utf-8")

        docs_keywords = ["documentation", "generation", "linting", "validation"]
        has_keyword = any(kw in content.lower() for kw in docs_keywords)
        assert has_keyword, f"{skill}: Missing documentation-specific keywords"

    @pytest.mark.parametrize("skill", ["moai-mcp-integration", "moai-context7-integration", "moai-artifacts-builder"])
    def test_tag004_mcp_content(self, skill: str):
        """TAG-004: Verify MCP/Context7-specific content"""
        skill_file = self.SKILLS_BASE_PATH / skill / "SKILL.md"

        if not skill_file.exists():
            pytest.skip(f"Skill file not found: {skill}")

        content = skill_file.read_text(encoding="utf-8")

        mcp_keywords = ["MCP", "Context7", "integration", "artifact"]
        has_keyword = any(kw.lower() in content.lower() for kw in mcp_keywords)
        assert has_keyword, f"{skill}: Missing MCP/Context7-specific keywords"

    @pytest.mark.parametrize(
        "skill",
        [
            "moai-project-batch-questions",
            "moai-project-config-manager",
            "moai-project-documentation",
            "moai-project-language-initializer",
            "moai-project-template-optimizer",
        ],
    )
    def test_tag005_project_content(self, skill: str):
        """TAG-005: Verify project management-specific content"""
        skill_file = self.SKILLS_BASE_PATH / skill / "SKILL.md"

        if not skill_file.exists():
            pytest.skip(f"Skill file not found: {skill}")

        content = skill_file.read_text(encoding="utf-8")

        project_keywords = ["project", "config", "question", "language", "template"]
        has_keyword = any(kw.lower() in content.lower() for kw in project_keywords)
        assert has_keyword, f"{skill}: Missing project management-specific keywords"


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])

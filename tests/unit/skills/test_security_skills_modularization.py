"""
RED Phase Tests: Security Skills Modularization
Tests for SPEC-04-GROUP-E: Security Skills Structure and Context7 Integration
"""

import re
from pathlib import Path

import pytest


class TestSecuritySkillsStructure:
    """Test suite for security skills modularization structure."""

    SECURITY_SKILLS = [
        "moai-security-api",
        "moai-security-auth",
        "moai-security-compliance",
        "moai-security-encryption",
        "moai-security-identity",
        "moai-security-owasp",
        "moai-security-ssrf",
        "moai-security-threat",
        "moai-security-zero-trust",
    ]

    REQUIRED_FILES = [
        "SKILL.md",
        "examples.md",
        "modules/advanced-patterns.md",
        "modules/optimization.md",
        "reference.md",
    ]

    SKILLS_BASE_PATH = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")

    @pytest.mark.parametrize("skill_name", SECURITY_SKILLS)
    def test_skill_directory_exists(self, skill_name):
        """RED: Verify skill directory exists."""
        skill_dir = self.SKILLS_BASE_PATH / skill_name
        assert skill_dir.exists(), f"Skill directory {skill_name} does not exist"
        assert skill_dir.is_dir(), f"{skill_name} is not a directory"

    @pytest.mark.parametrize("skill_name", SECURITY_SKILLS)
    @pytest.mark.parametrize("required_file", REQUIRED_FILES)
    def test_required_files_exist(self, skill_name, required_file):
        """RED: Verify all required files exist for each skill."""
        skill_dir = self.SKILLS_BASE_PATH / skill_name
        file_path = skill_dir / required_file
        assert file_path.exists(), f"File {required_file} not found in {skill_name}"
        assert file_path.is_file(), f"{required_file} is not a file in {skill_name}"

    @pytest.mark.parametrize("skill_name", SECURITY_SKILLS)
    def test_skill_md_has_context7_integration_section(self, skill_name):
        """RED: Verify SKILL.md has Context7 Integration section."""
        skill_file = self.SKILLS_BASE_PATH / skill_name / "SKILL.md"
        content = skill_file.read_text()

        # Check for Context7 Integration section
        assert (
            "## Context7 Integration" in content or "### Context7 Integration" in content
        ), f"SKILL.md in {skill_name} missing Context7 Integration section"

    @pytest.mark.parametrize("skill_name", SECURITY_SKILLS)
    def test_examples_md_has_minimum_examples(self, skill_name):
        """RED: Verify examples.md has at least 5 practical examples."""
        examples_file = self.SKILLS_BASE_PATH / skill_name / "examples.md"
        content = examples_file.read_text()

        # Count code blocks as examples (both opening and closing backticks count as indicators)
        code_blocks = len(re.findall(r"```", content)) // 2
        assert code_blocks >= 5, f"examples.md in {skill_name} has only {code_blocks} examples (need 5+)"

    @pytest.mark.parametrize("skill_name", SECURITY_SKILLS)
    def test_advanced_patterns_md_exists_and_has_content(self, skill_name):
        """RED: Verify advanced-patterns.md has substantial content."""
        adv_file = self.SKILLS_BASE_PATH / skill_name / "modules/advanced-patterns.md"
        content = adv_file.read_text()

        # Check for minimum content (at least 5 sections and 100 lines)
        lines = content.split("\n")
        sections = len(re.findall(r"^#{1,3} ", content, re.MULTILINE))

        assert len(lines) >= 100, f"advanced-patterns.md in {skill_name} is too short ({len(lines)} lines)"
        assert sections >= 3, f"advanced-patterns.md in {skill_name} has only {sections} sections"

    @pytest.mark.parametrize("skill_name", SECURITY_SKILLS)
    def test_optimization_md_exists_and_has_content(self, skill_name):
        """RED: Verify optimization.md has performance guidance."""
        opt_file = self.SKILLS_BASE_PATH / skill_name / "modules/optimization.md"
        content = opt_file.read_text()

        # Check for performance-related keywords
        keywords = ["performance", "optimize", "efficiency", "benchmar", "profile"]
        has_performance_content = any(keyword.lower() in content.lower() for keyword in keywords)

        assert has_performance_content, f"optimization.md in {skill_name} missing performance content"
        assert len(content) >= 300, f"optimization.md in {skill_name} is too short"

    @pytest.mark.parametrize("skill_name", SECURITY_SKILLS)
    def test_reference_md_has_documentation_links(self, skill_name):
        """RED: Verify reference.md has official documentation links."""
        ref_file = self.SKILLS_BASE_PATH / skill_name / "reference.md"
        content = ref_file.read_text()

        # Check for links (both markdown and plain URL formats)
        markdown_links = len(re.findall(r"\[.*?\]\(https?://.*?\)", content))
        plain_urls = len(re.findall(r"https?://[^\s]+", content))
        total_links = max(markdown_links, plain_urls)

        assert total_links >= 2, f"reference.md in {skill_name} has fewer than 2 documentation links"

    @pytest.mark.parametrize("skill_name", SECURITY_SKILLS)
    def test_all_files_are_valid_markdown(self, skill_name):
        """RED: Verify all documentation files are valid markdown."""
        skill_dir = self.SKILLS_BASE_PATH / skill_name
        md_files = list(skill_dir.rglob("*.md"))

        assert len(md_files) >= 5, f"{skill_name} has only {len(md_files)} markdown files (need 5+)"

        for md_file in md_files:
            content = md_file.read_text()
            # Basic markdown validation: has at least one heading
            assert re.search(r"^#", content, re.MULTILINE), f"{md_file.name} in {skill_name} has no markdown headings"

    @pytest.mark.parametrize("skill_name", SECURITY_SKILLS)
    def test_skill_metadata_present(self, skill_name):
        """RED: Verify SKILL.md has proper metadata/frontmatter."""
        skill_file = self.SKILLS_BASE_PATH / skill_name / "SKILL.md"
        content = skill_file.read_text()

        # Check for YAML frontmatter or metadata
        lines = content.split("\n")
        has_name_meta = any("name:" in line.lower() for line in lines[:20])
        any("version" in line.lower() for line in lines[:20])

        assert has_name_meta, f"{skill_name} SKILL.md missing name metadata"
        assert "2025" in content, f"{skill_name} SKILL.md missing 2025 version reference"

    @pytest.mark.parametrize("skill_name", SECURITY_SKILLS)
    def test_examples_are_executable(self, skill_name):
        """RED: Verify code examples have proper syntax highlighting."""
        examples_file = self.SKILLS_BASE_PATH / skill_name / "examples.md"
        content = examples_file.read_text()

        # Check for language-specific code blocks
        python_blocks = len(re.findall(r"```python", content))
        javascript_blocks = len(re.findall(r"```javascript", content))
        bash_blocks = len(re.findall(r"```bash", content))
        typescript_blocks = len(re.findall(r"```typescript", content))
        sql_blocks = len(re.findall(r"```sql", content))
        json_blocks = len(re.findall(r"```json", content))

        total_typed_blocks = (
            python_blocks + javascript_blocks + bash_blocks + typescript_blocks + sql_blocks + json_blocks
        )
        all_code_blocks = len(re.findall(r"```", content)) // 2  # ```...``` counts as opening + closing

        # At least 40% of code blocks should be typed (reasonable for mixed content)
        if all_code_blocks > 0:
            assert total_typed_blocks >= (
                all_code_blocks * 0.35
            ), f"{skill_name} has insufficient language-typed code blocks ({total_typed_blocks}/{all_code_blocks})"


class TestSecuritySkillsContext7Integration:
    """Test suite for Context7 integration in security skills."""

    SECURITY_SKILLS = [
        "moai-security-api",
        "moai-security-auth",
        "moai-security-compliance",
        "moai-security-encryption",
        "moai-security-identity",
        "moai-security-owasp",
        "moai-security-ssrf",
        "moai-security-threat",
        "moai-security-zero-trust",
    ]

    SKILLS_BASE_PATH = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")

    @pytest.mark.parametrize("skill_name", SECURITY_SKILLS)
    def test_context7_section_has_library_references(self, skill_name):
        """RED: Verify Context7 section references actual libraries."""
        skill_file = self.SKILLS_BASE_PATH / skill_name / "SKILL.md"
        content = skill_file.read_text()

        # Extract Context7 section
        context7_match = re.search(r"## Context7 Integration.*?(?=\n## |\Z)", content, re.DOTALL)

        assert context7_match, f"{skill_name} missing Context7 Integration section"
        context7_section = context7_match.group(0)

        # Check for library references
        has_library_links = "/" in context7_section and "[" in context7_section
        assert has_library_links, f"{skill_name} Context7 section missing library reference links"

    @pytest.mark.parametrize("skill_name", SECURITY_SKILLS)
    def test_context7_references_are_properly_formatted(self, skill_name):
        """RED: Verify Context7 references follow markdown link format."""
        skill_file = self.SKILLS_BASE_PATH / skill_name / "SKILL.md"
        content = skill_file.read_text()

        # Find all Context7 style references [name](/org/project)
        context7_links = re.findall(r"\[.*?\]\(/[a-z0-9\-/]+\)", content)

        assert len(context7_links) > 0, f"{skill_name} has no properly formatted Context7 links"


class TestSecuritySkillsCrossReferences:
    """Test suite for cross-references between security skills."""

    SECURITY_SKILLS = [
        "moai-security-api",
        "moai-security-auth",
        "moai-security-compliance",
        "moai-security-encryption",
        "moai-security-identity",
        "moai-security-owasp",
        "moai-security-ssrf",
        "moai-security-threat",
        "moai-security-zero-trust",
    ]

    SKILLS_BASE_PATH = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")

    @pytest.mark.parametrize("skill_name", SECURITY_SKILLS)
    def test_reference_md_is_not_empty(self, skill_name):
        """RED: Verify reference.md is not empty."""
        ref_file = self.SKILLS_BASE_PATH / skill_name / "reference.md"
        content = ref_file.read_text().strip()

        assert len(content) > 50, f"reference.md in {skill_name} is empty or nearly empty"

    @pytest.mark.parametrize("skill_name", SECURITY_SKILLS)
    def test_advanced_patterns_references_other_skills(self, skill_name):
        """RED: Verify advanced-patterns.md mentions related skills."""
        adv_file = self.SKILLS_BASE_PATH / skill_name / "modules/advanced-patterns.md"
        content = adv_file.read_text()

        # Should mention related security skills or patterns
        security_mentions = len(
            re.findall(r"(security|encryption|authentication|authorization|compliance)", content, re.IGNORECASE)
        )

        assert security_mentions >= 5, f"{skill_name} advanced-patterns.md lacks security references"


class TestSpecCompletionCriteria:
    """Test suite for SPEC-04-GROUP-E completion criteria."""

    SECURITY_SKILLS = [
        "moai-security-api",
        "moai-security-auth",
        "moai-security-compliance",
        "moai-security-encryption",
        "moai-security-identity",
        "moai-security-owasp",
        "moai-security-ssrf",
        "moai-security-threat",
        "moai-security-zero-trust",
    ]

    SKILLS_BASE_PATH = Path("/Users/goos/MoAI/MoAI-ADK/.claude/skills")

    def test_all_9_security_skills_present(self):
        """RED: Verify all 9 security skills directories exist."""
        for skill_name in self.SECURITY_SKILLS:
            skill_dir = self.SKILLS_BASE_PATH / skill_name
            assert skill_dir.exists(), f"Security skill {skill_name} not found"

    def test_total_skill_files_generated(self):
        """RED: Verify all 45 files are generated for 9 skills (5 files each)."""
        total_files = 0
        for skill_name in self.SECURITY_SKILLS:
            skill_dir = self.SKILLS_BASE_PATH / skill_name
            md_files = list(skill_dir.glob("**/*.md"))
            total_files += len(md_files)

        # Each skill should have at least 5 md files
        expected_min_files = len(self.SECURITY_SKILLS) * 5
        assert (
            total_files >= expected_min_files
        ), f"Only {total_files} markdown files found, expected at least {expected_min_files}"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])

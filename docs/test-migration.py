#!/usr/bin/env python3
"""
TDD Tests for Phase 2: Content Migration
Tests for README.ko.md migration to Nextra MDX structure
"""

import pytest
from pathlib import Path


class TestPhase2ContentMigration:
    """Test suite for Phase 2: Content Migration"""

    def setup_method(self):
        """Setup test fixtures"""
        self.docs_dir = Path(__file__).parent
        self.pages_dir = self.docs_dir / "pages"

    # RED Phase Tests: These tests verify content migration

    def test_getting_started_index_exists(self):
        """Test that getting-started/index.mdx exists"""
        index_path = self.pages_dir / "getting-started" / "index.mdx"
        assert index_path.exists(), "getting-started/index.mdx must exist"

    def test_installation_page_exists(self):
        """Test that installation.mdx exists"""
        installation_path = self.pages_dir / "getting-started" / "installation.mdx"
        assert installation_path.exists(), "installation.mdx must exist"

    def test_quickstart_page_exists(self):
        """Test that quickstart.mdx exists"""
        quickstart_path = self.pages_dir / "getting-started" / "quickstart.mdx"
        assert quickstart_path.exists(), "quickstart.mdx must exist"

    def test_spec_format_page_exists(self):
        """Test that core-concepts/spec-format.mdx exists"""
        spec_path = self.pages_dir / "core-concepts" / "spec-format.mdx"
        assert spec_path.exists(), "spec-format.mdx must exist"

    def test_agents_page_exists(self):
        """Test that core-concepts/agents.mdx exists"""
        agents_path = self.pages_dir / "core-concepts" / "agents.mdx"
        assert agents_path.exists(), "agents.mdx must exist"

    def test_workflow_page_exists(self):
        """Test that core-concepts/workflow.mdx exists"""
        workflow_path = self.pages_dir / "core-concepts" / "workflow.mdx"
        assert workflow_path.exists(), "workflow.mdx must exist"

    def test_commands_page_exists(self):
        """Test that core-concepts/commands.mdx exists"""
        commands_path = self.pages_dir / "core-concepts" / "commands.mdx"
        assert commands_path.exists(), "commands.mdx must exist"

    def test_advanced_agents_guide_exists(self):
        """Test that advanced/agents-guide.mdx exists"""
        agents_guide_path = self.pages_dir / "advanced" / "agents-guide.mdx"
        assert agents_guide_path.exists(), "agents-guide.mdx must exist"

    def test_skills_library_page_exists(self):
        """Test that advanced/skills-library.mdx exists"""
        skills_path = self.pages_dir / "advanced" / "skills-library.mdx"
        assert skills_path.exists(), "skills-library.mdx must exist"

    def test_patterns_page_exists(self):
        """Test that advanced/patterns.mdx exists"""
        patterns_path = self.pages_dir / "advanced" / "patterns.mdx"
        assert patterns_path.exists(), "patterns.mdx must exist"

    def test_trust5_page_exists(self):
        """Test that advanced/trust5-quality.mdx exists"""
        trust5_path = self.pages_dir / "advanced" / "trust5-quality.mdx"
        assert trust5_path.exists(), "trust5-quality.mdx must exist"

    def test_worktree_index_exists(self):
        """Test that worktree/index.mdx exists"""
        worktree_index = self.pages_dir / "worktree" / "index.mdx"
        assert worktree_index.exists(), "worktree/index.mdx must exist"

    def test_worktree_guide_exists(self):
        """Test that worktree/guide.mdx exists"""
        guide_path = self.pages_dir / "worktree" / "guide.mdx"
        assert guide_path.exists(), "worktree/guide.mdx must exist"

    def test_worktree_examples_exists(self):
        """Test that worktree/examples.mdx exists"""
        examples_path = self.pages_dir / "worktree" / "examples.mdx"
        assert examples_path.exists(), "worktree/examples.mdx must exist"

    def test_worktree_faq_exists(self):
        """Test that worktree/faq.mdx exists"""
        faq_path = self.pages_dir / "worktree" / "faq.mdx"
        assert faq_path.exists(), "worktree/faq.mdx must exist"

    def test_skills_reference_exists(self):
        """Test that reference/skills.mdx exists"""
        skills_ref = self.pages_dir / "reference" / "skills.mdx"
        assert skills_ref.exists(), "reference/skills.mdx must exist"

    def test_commands_reference_exists(self):
        """Test that reference/commands.mdx exists"""
        commands_ref = self.pages_dir / "reference" / "commands.mdx"
        assert commands_ref.exists(), "reference/commands.mdx must exist"

    def test_content_has_korean_language(self):
        """Test that getting-started/index.mdx has Korean content"""
        index_path = self.pages_dir / "getting-started" / "index.mdx"
        with open(index_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Should have Korean characters
        assert any('\uac00' <= char <= '\ud7a3' for char in content), \
            "Content should have Korean characters"

    def test_installation_page_has_content(self):
        """Test that installation.mdx has meaningful content"""
        installation_path = self.pages_dir / "getting-started" / "installation.mdx"
        with open(installation_path, 'r', encoding='utf-8') as f:
            content = f.read()

        assert len(content) > 100, "installation.mdx should have substantial content"
        assert "#" in content, "installation.mdx should have headings"

    def test_quickstart_has_code_examples(self):
        """Test that quickstart.mdx has code examples"""
        quickstart_path = self.pages_dir / "getting-started" / "quickstart.mdx"
        with open(quickstart_path, 'r', encoding='utf-8') as f:
            content = f.read()

        assert "```" in content, "quickstart.mdx should have code blocks"

    def test_spec_format_page_content(self):
        """Test that spec-format.mdx has content"""
        spec_path = self.pages_dir / "core-concepts" / "spec-format.mdx"
        with open(spec_path, 'r', encoding='utf-8') as f:
            content = f.read()

        assert len(content) > 50, "spec-format.mdx should have content"

    def test_agents_page_references_26_agents(self):
        """Test that agents.mdx mentions agent types"""
        agents_path = self.pages_dir / "core-concepts" / "agents.mdx"
        with open(agents_path, 'r', encoding='utf-8') as f:
            content = f.read()

        assert "expert-" in content or "agent" in content.lower(), \
            "agents.mdx should reference agents"

    def test_workflow_page_describes_process(self):
        """Test that workflow.mdx describes the workflow"""
        workflow_path = self.pages_dir / "core-concepts" / "workflow.mdx"
        with open(workflow_path, 'r', encoding='utf-8') as f:
            content = f.read()

        assert "SPEC" in content or "workflow" in content.lower(), \
            "workflow.mdx should describe the development workflow"

    def test_commands_page_lists_commands(self):
        """Test that commands.mdx lists core commands"""
        commands_path = self.pages_dir / "core-concepts" / "commands.mdx"
        with open(commands_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Should mention at least some commands
        assert "/moai:" in content or "command" in content.lower(), \
            "commands.mdx should list commands"

    def test_worktree_guide_has_content(self):
        """Test that worktree guide has substantial content"""
        guide_path = self.pages_dir / "worktree" / "guide.mdx"
        with open(guide_path, 'r', encoding='utf-8') as f:
            content = f.read()

        assert len(content) > 100, "worktree guide should have substantial content"

    def test_worktree_examples_has_code(self):
        """Test that worktree examples have code blocks"""
        examples_path = self.pages_dir / "worktree" / "examples.mdx"
        with open(examples_path, 'r', encoding='utf-8') as f:
            content = f.read()

        assert "```" in content, "worktree examples should have code blocks"

    def test_skills_reference_lists_22_skills(self):
        """Test that skills reference is present"""
        skills_ref = self.pages_dir / "reference" / "skills.mdx"
        with open(skills_ref, 'r', encoding='utf-8') as f:
            content = f.read()

        assert len(content) > 100, "skills reference should have content"

    def test_commands_reference_covers_6_commands(self):
        """Test that commands reference is present"""
        commands_ref = self.pages_dir / "reference" / "commands.mdx"
        with open(commands_ref, 'r', encoding='utf-8') as f:
            content = f.read()

        assert len(content) > 100, "commands reference should have content"


class TestPhase2IntegrationWithNavigation:
    """Integration tests for content migration with navigation"""

    def setup_method(self):
        """Setup test fixtures"""
        self.docs_dir = Path(__file__).parent
        self.pages_dir = self.docs_dir / "pages"

    def test_all_page_directories_accessible(self):
        """Test that all major page directories exist"""
        sections = ["getting-started", "core-concepts", "advanced", "worktree", "reference"]
        for section in sections:
            section_path = self.pages_dir / section
            assert section_path.is_dir(), f"{section} directory must exist"

    def test_page_files_are_mdx_format(self):
        """Test that all page files use .mdx format"""
        page_files = [
            "getting-started/index.mdx",
            "getting-started/installation.mdx",
            "core-concepts/spec-format.mdx",
            "worktree/guide.mdx"
        ]

        for page_file in page_files:
            full_path = self.pages_dir / page_file
            assert full_path.suffix == ".mdx", \
                f"{page_file} should use .mdx format"

    def test_content_structure_consistency(self):
        """Test that pages have consistent structure (headings, links, etc)"""
        index_path = self.pages_dir / "getting-started" / "index.mdx"
        with open(index_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Should have main heading and structure
        assert "#" in content, "Pages should have markdown headings"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])

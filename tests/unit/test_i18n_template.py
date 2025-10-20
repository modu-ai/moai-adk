# @TEST:I18N-001 | SPEC: SPEC-I18N-001.md
"""Tests for i18n template system (SPEC-I18N-001 v0.0.1)

This module tests the multi-language template system for Korean and English.
"""

import shutil
from pathlib import Path

import pytest

from moai_adk.core.template.processor import TemplateProcessor


class TestI18nTemplate:
    """Tests for i18n template copying functionality."""

    @pytest.fixture
    def template_root(self, tmp_path: Path) -> Path:
        """Create mock template directories for testing.

        Args:
            tmp_path: Pytest temporary directory fixture.

        Returns:
            Path to the mock template root directory.
        """
        # Create mock template root
        template_root = tmp_path / "templates"
        template_root.mkdir()

        # Create Korean template (.claude-ko/)
        claude_ko = template_root / ".claude-ko"
        claude_ko.mkdir()
        (claude_ko / "commands").mkdir()
        (claude_ko / "agents").mkdir()
        (claude_ko / "commands" / "1-plan.md").write_text("# 계획 수립\n한국어 커맨드")
        (claude_ko / "agents" / "spec-builder.md").write_text("# 명세 작성자\n한국어 에이전트")
        (claude_ko / "README.md").write_text("# Alfred 에이전트\n한국어 문서")

        # Create English template (.claude-en/)
        claude_en = template_root / ".claude-en"
        claude_en.mkdir()
        (claude_en / "commands").mkdir()
        (claude_en / "agents").mkdir()
        (claude_en / "commands" / "1-plan.md").write_text("# Plan\nEnglish command")
        (claude_en / "agents" / "spec-builder.md").write_text("# Spec Builder\nEnglish agent")
        (claude_en / "README.md").write_text("# Alfred Agents\nEnglish docs")

        return template_root

    @pytest.fixture
    def processor(self, tmp_path: Path, template_root: Path, monkeypatch: pytest.MonkeyPatch) -> TemplateProcessor:
        """Create TemplateProcessor with mock template root.

        Args:
            tmp_path: Pytest temporary directory fixture.
            template_root: Mock template root from fixture.
            monkeypatch: Pytest monkeypatch fixture.

        Returns:
            TemplateProcessor instance with mocked template root.
        """
        processor = TemplateProcessor(tmp_path)
        # Mock the template_root to use our test templates
        monkeypatch.setattr(processor, 'template_root', template_root)
        return processor

    def test_copy_claude_template_korean(self, processor: TemplateProcessor, tmp_path: Path) -> None:
        """Test copying Korean template to .claude/ directory.

        @TEST:I18N-001:COPY-KO | Korean template copy test
        @SPEC:I18N-001 | Event-driven requirement: WHEN locale="ko"

        Args:
            processor: TemplateProcessor fixture.
            tmp_path: Pytest temporary directory fixture.
        """
        # ACT: Copy Korean template
        processor.copy_claude_template("ko")

        # ASSERT: Verify .claude/ directory exists
        claude_dir = tmp_path / ".claude"
        assert claude_dir.exists(), ".claude/ directory should exist"

        # ASSERT: Verify Korean content is copied
        command_file = claude_dir / "commands" / "1-plan.md"
        assert command_file.exists(), "1-plan.md should exist"
        content = command_file.read_text()
        assert "한국어" in content, "Content should be in Korean"
        assert "계획 수립" in content, "Korean title should be present"

        # ASSERT: Verify agent file
        agent_file = claude_dir / "agents" / "spec-builder.md"
        assert agent_file.exists(), "spec-builder.md should exist"
        agent_content = agent_file.read_text()
        assert "한국어" in agent_content, "Agent content should be in Korean"

        # ASSERT: Verify README
        readme = claude_dir / "README.md"
        assert readme.exists(), "README.md should exist"
        readme_content = readme.read_text()
        assert "한국어" in readme_content, "README should be in Korean"

    def test_copy_claude_template_english(self, processor: TemplateProcessor, tmp_path: Path) -> None:
        """Test copying English template to .claude/ directory.

        @TEST:I18N-001:COPY-EN | English template copy test
        @SPEC:I18N-001 | Event-driven requirement: WHEN locale="en"

        Args:
            processor: TemplateProcessor fixture.
            tmp_path: Pytest temporary directory fixture.
        """
        # ACT: Copy English template
        processor.copy_claude_template("en")

        # ASSERT: Verify .claude/ directory exists
        claude_dir = tmp_path / ".claude"
        assert claude_dir.exists(), ".claude/ directory should exist"

        # ASSERT: Verify English content is copied
        command_file = claude_dir / "commands" / "1-plan.md"
        assert command_file.exists(), "1-plan.md should exist"
        content = command_file.read_text()
        assert "English" in content, "Content should be in English"
        assert "Plan" in content, "English title should be present"

        # ASSERT: Verify agent file
        agent_file = claude_dir / "agents" / "spec-builder.md"
        assert agent_file.exists(), "spec-builder.md should exist"
        agent_content = agent_file.read_text()
        assert "English" in agent_content, "Agent content should be in English"

        # ASSERT: Verify README
        readme = claude_dir / "README.md"
        assert readme.exists(), "README.md should exist"
        readme_content = readme.read_text()
        assert "English" in readme_content, "README should be in English"

    def test_copy_claude_template_fallback_to_english(
        self,
        processor: TemplateProcessor,
        tmp_path: Path,
        caplog: pytest.LogCaptureFixture
    ) -> None:
        """Test fallback to English for unsupported locales.

        @TEST:I18N-001:FALLBACK | Unsupported locale fallback test
        @SPEC:I18N-001 | Constraint: IF unsupported locale, fallback to "en"

        Args:
            processor: TemplateProcessor fixture.
            tmp_path: Pytest temporary directory fixture.
            caplog: Pytest log capture fixture.
        """
        # ACT: Try to copy with unsupported locale (Japanese)
        processor.copy_claude_template("ja")

        # ASSERT: Verify warning was logged
        assert "Unsupported locale 'ja'" in caplog.text or "falling back to" in caplog.text.lower()

        # ASSERT: Verify English template was copied (fallback)
        claude_dir = tmp_path / ".claude"
        assert claude_dir.exists(), ".claude/ directory should exist"

        command_file = claude_dir / "commands" / "1-plan.md"
        assert command_file.exists(), "1-plan.md should exist"
        content = command_file.read_text()
        assert "English" in content, "Should fallback to English template"

    def test_copy_claude_template_missing_template_dir(
        self,
        processor: TemplateProcessor,
        tmp_path: Path
    ) -> None:
        """Test error handling when template directory is missing.

        @TEST:I18N-001:ERROR | Missing template directory error test
        @SPEC:I18N-001 | Constraint: Template must exist

        Args:
            processor: TemplateProcessor fixture.
            tmp_path: Pytest temporary directory fixture.
        """
        # Remove English template to simulate missing template
        template_en = processor.template_root / ".claude-en"
        if template_en.exists():
            shutil.rmtree(template_en)

        # ACT & ASSERT: Should raise FileNotFoundError
        with pytest.raises(FileNotFoundError, match="Template directory not found"):
            processor.copy_claude_template("en")

    def test_template_structure_consistency(self, template_root: Path) -> None:
        """Test that Korean and English templates have identical structure.

        @TEST:I18N-001:STRUCTURE | Template structure consistency test
        @SPEC:I18N-001 | Constraint: Template structures must be identical

        Args:
            template_root: Mock template root from fixture.
        """
        claude_ko = template_root / ".claude-ko"
        claude_en = template_root / ".claude-en"

        # Get all files in Korean template
        ko_files = {
            str(f.relative_to(claude_ko))
            for f in claude_ko.rglob("*")
            if f.is_file()
        }

        # Get all files in English template
        en_files = {
            str(f.relative_to(claude_en))
            for f in claude_en.rglob("*")
            if f.is_file()
        }

        # ASSERT: File structures should be identical
        assert ko_files == en_files, "Korean and English templates must have identical file structure"

    def test_copy_claude_template_default_korean(
        self,
        processor: TemplateProcessor,
        tmp_path: Path
    ) -> None:
        """Test that Korean is the default locale.

        @TEST:I18N-001:DEFAULT | Default locale test
        @SPEC:I18N-001 | Requirement: Default locale should be "ko"

        Args:
            processor: TemplateProcessor fixture.
            tmp_path: Pytest temporary directory fixture.
        """
        # ACT: Copy with default parameter (should be "ko")
        processor.copy_claude_template()

        # ASSERT: Verify Korean template was copied
        claude_dir = tmp_path / ".claude"
        assert claude_dir.exists(), ".claude/ directory should exist"

        command_file = claude_dir / "commands" / "1-plan.md"
        content = command_file.read_text()
        assert "한국어" in content, "Default should be Korean template"

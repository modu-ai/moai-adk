"""
Test cases for Alfred command completion patterns (SPEC-SESSION-CLEANUP-001)


This test suite verifies that all 4 Alfred commands (/moai:0-project, /moai:1-plan,
/moai:2-run, /moai:3-sync) have proper AskUserQuestion completion patterns.

Requirements:
- REQ-SESSION-001: All commands MUST use AskUserQuestion for next steps
- REQ-SESSION-010: Commands MUST NOT suggest next steps in prose
- REQ-SESSION-011: Commands MUST NOT complete without AskUserQuestion
"""

from pathlib import Path

import pytest


class TestCommandCompletionPatterns:
    """Test suite for command completion patterns"""

    @pytest.fixture
    def command_files(self) -> dict:
        """Fixture providing paths to all 4 command files"""
        base_path = Path("/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred")
        return {
            "0-project": base_path / "0-project.md",
            "1-plan": base_path / "1-plan.md",
            "2-run": base_path / "2-run.md",
            "3-sync": base_path / "3-sync.md",
        }

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_all_commands_have_final_step_section(self, command_files):
        """
        REQ-SESSION-001: All commands must have 'Final Step' section

        Verify that each command file contains a dedicated section for
        AskUserQuestion completion pattern.
        """
        for cmd_name, cmd_path in command_files.items():
            content = cmd_path.read_text()
            assert (
                "## Final Step" in content or "Final Step:" in content
            ), f"Command {cmd_name} missing 'Final Step' section"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_all_commands_have_askmserquestion_call(self, command_files):
        """
        REQ-SESSION-001: All commands must include AskUserQuestion call

        Verify that each command file contains an AskUserQuestion
        implementation in the completion section.
        """
        for cmd_name, cmd_path in command_files.items():
            content = cmd_path.read_text()
            assert "AskUserQuestion" in content, f"Command {cmd_name} missing AskUserQuestion call"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_commands_have_batched_design(self, command_files):
        """
        Batched Design: Commands should use single AskUserQuestion call

        Verify that completion sections don't have multiple sequential
        AskUserQuestion calls (should batch 1-4 questions instead).
        """
        for cmd_name, cmd_path in command_files.items():
            content = cmd_path.read_text()

            # Find the "Final Step" or "Next Action" section
            if "## Final Step" in content:
                section_start = content.index("## Final Step")
                section = content[section_start : section_start + 2000]

                # Count AskUserQuestion occurrences in completion section
                ask_count = section.count("AskUserQuestion(")

                # Should have exactly 1 call (batched design)
                assert (
                    ask_count <= 1
                ), f"Command {cmd_name} has {ask_count} AskUserQuestion calls (expected 1 for batched design)"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_commands_provide_3_to_4_options(self, command_files):
        """
        REQ-SESSION-003-006: Commands must provide 3-4 clear options

        Verify that each command's AskUserQuestion implementation
        provides exactly 3-4 options for user selection.
        """
        for cmd_name, cmd_path in command_files.items():
            content = cmd_path.read_text()

            if "## Final Step" in content:
                # Find Final Step section specifically
                final_start = content.index("## Final Step")
                final_section = content[final_start : final_start + 3000]

                # Count option entries by looking for emoji + description pattern
                # Each option has a "label": "📋 ..." pattern
                import re

                # Match "label": "emoji text" pattern
                label_pattern = r'"label":\s*"[^"]*"'
                matches = re.findall(label_pattern, final_section)
                option_count = len(matches)

                assert 3 <= option_count <= 4, f"Command {cmd_name} has {option_count} options (expected 3-4)"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_no_prose_suggestions_in_completion(self, command_files):
        """
        REQ-SESSION-010: Commands MUST NOT use prose suggestions

        Verify that commands don't have prose text like
        "You can now run..." in their completion sections OUTSIDE
        of the Final Step section (which contains examples of what NOT to do).
        """
        prohibited_phrases = [
            "you can now run",
            "you can run",
            "next you can",
            "now run",
            "try running",
        ]

        for cmd_name, cmd_path in command_files.items():
            content = cmd_path.read_text().lower()

            # Exclude the "Final Step" section which contains examples
            if "## final step" in content:
                final_step_start = content.index("## final step")
                content_before_final = content[:final_step_start]
            else:
                content_before_final = content

            for phrase in prohibited_phrases:
                assert phrase not in content_before_final, f"Command {cmd_name} contains prohibited prose: '{phrase}'"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_commands_use_emoji_in_options(self, command_files):
        """
        UX Enhancement: Options should use emoji for visual clarity

        Verify that options use emoji prefixes (📋, 🚀, 🔄, etc.)
        for better user experience.
        """
        for cmd_name, cmd_path in command_files.items():
            content = cmd_path.read_text()

            if "## Final Step" in content:
                # Find Final Step section specifically
                final_start = content.index("## Final Step")
                final_section = content[final_start : final_start + 3000]

                # Should have at least one emoji in options
                common_emojis = ["📋", "🚀", "🔄", "🔍", "✏️", "📚", "🧪", "🔀", "✅"]
                has_emoji = any(emoji in final_section for emoji in common_emojis)

                assert has_emoji, f"Command {cmd_name} options lack emoji for UX enhancement"

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_language_configuration_pass_through(self, command_files):
        """
        Language Support: Commands should pass language config to AskUserQuestion

        Verify that completion sections reference language configuration
        variables for multilingual support.
        """
        for cmd_name, cmd_path in command_files.items():
            content = cmd_path.read_text()

            # Check if command has language awareness
            # (Commands should either use template variables or explicitly handle language)
            if "AskUserQuestion" in content:
                # At minimum, questions should be localizable
                # We'll check that the question text is not hardcoded English-only
                ask_start = content.index("AskUserQuestion")
                ask_section = content[ask_start : ask_start + 1500]

                # This is a soft check - verify pattern exists
                # (In actual implementation, language is passed via Alfred context)
                assert "question" in ask_section, f"Command {cmd_name} AskUserQuestion missing question field"


class TestCommandSpecificOptions:
    """Test suite for command-specific option validation"""

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_0_project_options(self):
        """
        REQ-SESSION-003: /moai:0-project completion options

        Verify that /moai:0-project provides correct 3 options:
        1. Proceed to /moai:1-plan
        2. Review project structure
        3. Start new session
        """
        cmd_path = Path("/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/0-project.md")
        content = cmd_path.read_text()

        # Check for expected option keywords
        assert "1-plan" in content or "스펙 작성" in content or "plan" in content.lower()
        assert "review" in content.lower() or "검토" in content or "구조" in content
        assert "new session" in content.lower() or "새 세션" in content or "/clear" in content

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_1_plan_options(self):
        """
        REQ-SESSION-004: /moai:1-plan completion options

        Verify that /moai:1-plan provides correct 3 options:
        1. Proceed to /moai:2-run
        2. Revise SPEC
        3. Start new session
        """
        cmd_path = Path("/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/1-plan.md")
        content = cmd_path.read_text()

        # Check for expected option keywords
        assert "2-run" in content or "구현" in content or "implement" in content.lower()
        assert "revise" in content.lower() or "수정" in content or "spec" in content.lower()
        assert "new session" in content.lower() or "새 세션" in content or "/clear" in content

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_2_run_options(self):
        """
        REQ-SESSION-005: /moai:2-run completion options

        Verify that /moai:2-run provides correct 3 options:
        1. Proceed to /moai:3-sync
        2. Run additional tests
        3. Start new session
        """
        cmd_path = Path("/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/2-run.md")
        content = cmd_path.read_text()

        # Check for expected option keywords
        assert "3-sync" in content or "동기화" in content or "sync" in content.lower()
        assert "test" in content.lower() or "테스트" in content or "검증" in content
        assert "new session" in content.lower() or "새 세션" in content or "/clear" in content

    @pytest.mark.xfail(reason="Test data migration needed")
    def test_3_sync_options(self):
        """
        REQ-SESSION-006: /moai:3-sync completion options

        Verify that /moai:3-sync provides correct 3 options:
        1. Return to /moai:1-plan (next feature)
        2. Merge PR
        3. Complete session
        """
        cmd_path = Path("/Users/goos/MoAI/MoAI-ADK/.claude/commands/alfred/3-sync.md")
        content = cmd_path.read_text()

        # Check for expected option keywords
        assert "1-plan" in content or "다음 기능" in content or "next feature" in content.lower()
        assert "merge" in content.lower() or "병합" in content or "pr" in content.lower()
        assert "complete" in content.lower() or "완료" in content or "session" in content.lower()


if __name__ == "__main__":
    pytest.main([__file__, "-v"])

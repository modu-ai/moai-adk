"""
UI/UX Plugin Command Tests

@TEST:UIUX-PLUGIN-001 - UI/UX Plugin Command Execution
Tests for `/setup-shadcn-ui` command functionality

@CODE:UIUX-TESTS-SUITE-001:TEST
"""

import pytest
import tempfile
import os
import json
from pathlib import Path
from typing import Dict, Any


# @CODE:UIUX-INIT-COMMAND-001:TEST
class TestSetupShadcnUICommand:
    """Test cases for /setup-shadcn-ui command"""

    @pytest.fixture
    def temp_project_dir(self) -> Path:
        """Create temporary project directory for testing"""
        with tempfile.TemporaryDirectory() as tmpdir:
            yield Path(tmpdir)

    @pytest.fixture
    def mock_uiux_config(self) -> Dict[str, Any]:
        """Mock UI/UX configuration"""
        return {
            "project_name": "test-project",
            "framework": "nextjs",
            "components": ["button", "input", "card"],
            "created_at": "2025-10-30",
            "version": "1.0.0-dev"
        }

    # ========== NORMAL CASES ==========

    def test_setup_shadcn_basic(self, temp_project_dir):
        """
        GIVEN: User invokes /setup-shadcn-ui with project name
        WHEN: Project name is valid
        THEN: shadcn/ui configuration files are created
        """
        # @CODE:UIUX-SETUP-BASIC-001:TEST
        project_name = "my-ui-project"
        config_dir = temp_project_dir / ".moai" / "ui" / "shadcn"

        from ui_ux_plugin.commands import setup_shadcn

        result = setup_shadcn.execute(
            project_name=project_name,
            output_dir=temp_project_dir,
            framework="nextjs"
        )

        assert result.success is True
        assert config_dir.exists()
        assert (config_dir / "components.json").exists()

    def test_setup_shadcn_creates_config_json(self, temp_project_dir):
        """
        GIVEN: /setup-shadcn-ui creates configuration
        WHEN: File is generated
        THEN: components.json contains valid configuration
        """
        # @CODE:UIUX-CONFIG-JSON-001:TEST
        from ui_ux_plugin.commands import setup_shadcn

        result = setup_shadcn.execute(
            project_name="config-test",
            output_dir=temp_project_dir,
            framework="nextjs"
        )

        config_file = temp_project_dir / ".moai" / "ui" / "shadcn" / "components.json"
        content = json.loads(config_file.read_text())

        # Verify configuration structure
        assert "framework" in content
        assert "aliases" in content
        assert "components" in content

    def test_setup_shadcn_creates_tailwind_config(self, temp_project_dir):
        """
        GIVEN: /setup-shadcn-ui with shadcn/ui setup
        WHEN: Configuration generated
        THEN: tailwind.config.js includes custom theme
        """
        # @CODE:UIUX-TAILWIND-001:TEST
        from ui_ux_plugin.commands import setup_shadcn

        result = setup_shadcn.execute(
            project_name="tailwind-test",
            output_dir=temp_project_dir,
            framework="nextjs"
        )

        tailwind_file = temp_project_dir / ".moai" / "ui" / "shadcn" / "tailwind.config.template.js"
        assert tailwind_file.exists()
        content = tailwind_file.read_text()
        assert "extend" in content or "theme" in content

    def test_setup_shadcn_creates_component_structure(self, temp_project_dir):
        """
        GIVEN: /setup-shadcn-ui generates component setup
        WHEN: Command completes
        THEN: Component directory structure is created
        """
        # @CODE:UIUX-COMPONENT-STRUCTURE-001:TEST
        from ui_ux_plugin.commands import setup_shadcn

        result = setup_shadcn.execute(
            project_name="components-test",
            output_dir=temp_project_dir,
            framework="nextjs"
        )

        components_dir = temp_project_dir / ".moai" / "ui" / "shadcn" / "components"
        assert components_dir.exists()

    # ========== OPTIONS CASES ==========

    def test_setup_shadcn_with_custom_components(self, temp_project_dir):
        """
        GIVEN: /setup-shadcn-ui with --components option
        WHEN: Custom components specified
        THEN: Only specified components are included
        """
        # @CODE:UIUX-CUSTOM-COMPONENTS-001:TEST
        from ui_ux_plugin.commands import setup_shadcn

        components = ["button", "input", "modal"]
        result = setup_shadcn.execute(
            project_name="custom-components",
            output_dir=temp_project_dir,
            framework="nextjs",
            components=components
        )

        assert result.success is True
        config_file = temp_project_dir / ".moai" / "ui" / "shadcn" / "components.json"
        content = json.loads(config_file.read_text())

        # Verify components are in config
        for comp in components:
            assert comp in str(content)

    def test_setup_shadcn_with_dark_mode(self, temp_project_dir):
        """
        GIVEN: /setup-shadcn-ui with --dark-mode option
        WHEN: Dark mode is enabled
        THEN: Dark mode theme configuration is included
        """
        # @CODE:UIUX-DARK-MODE-001:TEST
        from ui_ux_plugin.commands import setup_shadcn

        result = setup_shadcn.execute(
            project_name="dark-mode-test",
            output_dir=temp_project_dir,
            framework="nextjs",
            dark_mode=True
        )

        assert result.success is True
        tailwind_file = temp_project_dir / ".moai" / "ui" / "shadcn" / "tailwind.config.template.js"
        content = tailwind_file.read_text()
        assert "dark" in content.lower()

    def test_setup_shadcn_react_framework(self, temp_project_dir):
        """
        GIVEN: /setup-shadcn-ui with --framework=react
        WHEN: React framework specified
        THEN: React configuration is applied
        """
        # @CODE:UIUX-REACT-FRAMEWORK-001:TEST
        from ui_ux_plugin.commands import setup_shadcn

        result = setup_shadcn.execute(
            project_name="react-test",
            output_dir=temp_project_dir,
            framework="react"
        )

        assert result.success is True
        config_file = temp_project_dir / ".moai" / "ui" / "shadcn" / "components.json"
        content = json.loads(config_file.read_text())
        assert content.get("framework") == "react"

    # ========== ERROR CASES ==========

    def test_setup_shadcn_invalid_project_name(self, temp_project_dir):
        """
        GIVEN: /setup-shadcn-ui with invalid project name
        WHEN: Project name contains uppercase
        THEN: Raises ValueError
        """
        # @CODE:UIUX-INVALID-NAME-001:TEST
        from ui_ux_plugin.commands import setup_shadcn

        with pytest.raises(ValueError) as exc_info:
            setup_shadcn.execute(
                project_name="MyProject",
                output_dir=temp_project_dir
            )

        assert "lowercase" in str(exc_info.value).lower()

    def test_setup_shadcn_invalid_framework(self, temp_project_dir):
        """
        GIVEN: /setup-shadcn-ui with unsupported framework
        WHEN: Framework not in supported list
        THEN: Raises ValueError
        """
        # @CODE:UIUX-INVALID-FRAMEWORK-001:TEST
        from ui_ux_plugin.commands import setup_shadcn

        with pytest.raises(ValueError) as exc_info:
            setup_shadcn.execute(
                project_name="test-project",
                output_dir=temp_project_dir,
                framework="unsupported"
            )

        assert "framework" in str(exc_info.value).lower()

    def test_setup_shadcn_invalid_component(self, temp_project_dir):
        """
        GIVEN: /setup-shadcn-ui with invalid component name
        WHEN: Component not in available list
        THEN: Raises ValueError
        """
        # @CODE:UIUX-INVALID-COMPONENT-001:TEST
        from ui_ux_plugin.commands import setup_shadcn

        with pytest.raises(ValueError) as exc_info:
            setup_shadcn.execute(
                project_name="test-project",
                output_dir=temp_project_dir,
                components=["invalid-component"]
            )

        assert "component" in str(exc_info.value).lower()

    def test_setup_shadcn_duplicate_config(self, temp_project_dir):
        """
        GIVEN: Two /setup-shadcn-ui calls with same project
        WHEN: First call completes
        THEN: Second call raises error about existing config
        """
        # @CODE:UIUX-DUPLICATE-CONFIG-001:TEST
        from ui_ux_plugin.commands import setup_shadcn

        # First call succeeds
        setup_shadcn.execute(
            project_name="duplicate-test",
            output_dir=temp_project_dir
        )

        # Second call should fail
        with pytest.raises(FileExistsError) as exc_info:
            setup_shadcn.execute(
                project_name="duplicate-test",
                output_dir=temp_project_dir
            )

        assert "already exists" in str(exc_info.value).lower()

    # ========== BOUNDARY CASES ==========

    def test_setup_shadcn_minimal_project_name(self, temp_project_dir):
        """
        GIVEN: /setup-shadcn-ui with minimal 3-character name
        WHEN: Project name meets minimum length
        THEN: Successfully creates configuration
        """
        # @CODE:UIUX-MIN-NAME-001:TEST
        from ui_ux_plugin.commands import setup_shadcn

        result = setup_shadcn.execute(
            project_name="abc",
            output_dir=temp_project_dir
        )

        assert result.success is True

    def test_setup_shadcn_with_all_components(self, temp_project_dir):
        """
        GIVEN: /setup-shadcn-ui with all available components
        WHEN: All components requested
        THEN: All components configured successfully
        """
        # @CODE:UIUX-ALL-COMPONENTS-001:TEST
        from ui_ux_plugin.commands import setup_shadcn

        result = setup_shadcn.execute(
            project_name="all-components",
            output_dir=temp_project_dir,
            components=["button", "input", "card", "modal", "dropdown"]
        )

        assert result.success is True
        config_file = temp_project_dir / ".moai" / "ui" / "shadcn" / "components.json"
        assert config_file.exists()


# @CODE:UIUX-INTEGRATION-001:TEST
class TestUIUXPluginIntegration:
    """Integration tests for UI/UX Plugin"""

    def test_setup_shadcn_end_to_end(self, tmp_path):
        """
        GIVEN: User invokes /setup-shadcn-ui with minimal arguments
        WHEN: Command executes
        THEN: All expected files are created with correct structure
        """
        from ui_ux_plugin.commands import setup_shadcn

        result = setup_shadcn.execute(
            project_name="e2e-test",
            output_dir=tmp_path
        )

        assert result.success is True
        ui_dir = tmp_path / ".moai" / "ui" / "shadcn"

        # Verify all files exist
        assert (ui_dir / "components.json").exists()
        assert (ui_dir / "tailwind.config.template.js").exists()
        assert (ui_dir / "components").exists()

    def test_setup_shadcn_returns_correct_structure(self, tmp_path):
        """
        GIVEN: /setup-shadcn-ui command execution
        WHEN: Command completes
        THEN: Returns structured output with success and file list
        """
        from ui_ux_plugin.commands import setup_shadcn

        result = setup_shadcn.execute(
            project_name="output-test",
            output_dir=tmp_path
        )

        # Verify result object structure
        assert hasattr(result, "success")
        assert hasattr(result, "config_dir")
        assert hasattr(result, "files_created")
        assert hasattr(result, "message")

        assert isinstance(result.files_created, list)
        assert len(result.files_created) >= 2


# @CODE:UIUX-PERFORMANCE-001:TEST
class TestUIUXPluginPerformance:
    """Performance tests for UI/UX Plugin"""

    def test_setup_shadcn_completes_quickly(self, tmp_path):
        """
        GIVEN: /setup-shadcn-ui command
        WHEN: UI/UX setup execution
        THEN: Completes within 5 seconds
        """
        import time
        from ui_ux_plugin.commands import setup_shadcn

        start = time.time()
        setup_shadcn.execute(
            project_name="perf-test",
            output_dir=tmp_path
        )
        elapsed = time.time() - start

        assert elapsed < 5.0, f"Command took {elapsed}s, expected < 5s"


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])

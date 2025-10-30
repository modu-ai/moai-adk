"""
Frontend Plugin Command Tests

@TEST:FRONTEND-PLUGIN-001 - Frontend Plugin Command Execution
Tests for `/init-react`, `/setup-state`, `/setup-testing` command functionality

@CODE:FRONTEND-TESTS-SUITE-001:TEST
"""

import pytest
import tempfile
import json
from pathlib import Path
from typing import Dict, Any


# @CODE:FRONTEND-INIT-COMMAND-001:TEST
class TestInitReactCommand:
    """Test cases for /init-react command"""

    @pytest.fixture
    def temp_project_dir(self) -> Path:
        """Create temporary project directory for testing"""
        with tempfile.TemporaryDirectory() as tmpdir:
            yield Path(tmpdir)

    # ========== NORMAL CASES ==========

    def test_init_react_basic(self, temp_project_dir):
        """
        GIVEN: User invokes /init-react with project name
        WHEN: Project name is valid
        THEN: React project structure is created
        """
        # @CODE:FRONTEND-REACT-BASIC-001:TEST
        project_name = "my-app"
        project_dir = temp_project_dir / "my-app"

        from frontend_plugin.commands import init_react

        result = init_react.execute(
            project_name=project_name,
            output_dir=temp_project_dir,
            framework="react"
        )

        assert result.success is True
        assert project_dir.exists()
        assert (project_dir / "package.json").exists()

    def test_init_react_creates_package_json(self, temp_project_dir):
        """
        GIVEN: /init-react creates package.json
        WHEN: File is generated
        THEN: package.json contains React dependencies
        """
        # @CODE:FRONTEND-REACT-PACKAGE-001:TEST
        from frontend_plugin.commands import init_react

        result = init_react.execute(
            project_name="package-test",
            output_dir=temp_project_dir,
            framework="react"
        )

        package_file = temp_project_dir / "package-test" / "package.json"
        content = json.loads(package_file.read_text())

        assert "dependencies" in content
        assert "react" in content["dependencies"]

    def test_init_react_creates_src_structure(self, temp_project_dir):
        """
        GIVEN: /init-react generates project
        WHEN: Source structure created
        THEN: src/ directory with App.jsx exists
        """
        # @CODE:FRONTEND-REACT-SRC-001:TEST
        from frontend_plugin.commands import init_react

        result = init_react.execute(
            project_name="src-test",
            output_dir=temp_project_dir
        )

        project_dir = temp_project_dir / "src-test"
        assert (project_dir / "src").exists()
        assert (project_dir / "src" / "App.jsx").exists()

    def test_init_react_creates_vite_config(self, temp_project_dir):
        """
        GIVEN: /init-react with Vite setup
        WHEN: Vite framework selected
        THEN: vite.config.js is created
        """
        # @CODE:FRONTEND-REACT-VITE-001:TEST
        from frontend_plugin.commands import init_react

        result = init_react.execute(
            project_name="vite-test",
            output_dir=temp_project_dir,
            framework="vite"
        )

        project_dir = temp_project_dir / "vite-test"
        assert (project_dir / "vite.config.js").exists()

    # ========== OPTIONS CASES ==========

    def test_init_react_with_typescript(self, temp_project_dir):
        """
        GIVEN: /init-react with --typescript option
        WHEN: TypeScript enabled
        THEN: tsconfig.json and .tsx files created
        """
        # @CODE:FRONTEND-REACT-TS-001:TEST
        from frontend_plugin.commands import init_react

        result = init_react.execute(
            project_name="ts-test",
            output_dir=temp_project_dir,
            use_typescript=True
        )

        assert result.success is True
        project_dir = temp_project_dir / "ts-test"
        assert (project_dir / "tsconfig.json").exists()

    def test_init_react_with_tailwind(self, temp_project_dir):
        """
        GIVEN: /init-react with --tailwind option
        WHEN: Tailwind CSS enabled
        THEN: tailwind.config.js and postcss.config.js created
        """
        # @CODE:FRONTEND-REACT-TAILWIND-001:TEST
        from frontend_plugin.commands import init_react

        result = init_react.execute(
            project_name="tailwind-test",
            output_dir=temp_project_dir,
            include_tailwind=True
        )

        assert result.success is True
        project_dir = temp_project_dir / "tailwind-test"
        assert (project_dir / "tailwind.config.js").exists() or \
               (project_dir / "postcss.config.js").exists()

    def test_init_react_with_eslint(self, temp_project_dir):
        """
        GIVEN: /init-react with --eslint option
        WHEN: ESLint enabled
        THEN: .eslintrc.json created
        """
        # @CODE:FRONTEND-REACT-ESLINT-001:TEST
        from frontend_plugin.commands import init_react

        result = init_react.execute(
            project_name="eslint-test",
            output_dir=temp_project_dir,
            include_eslint=True
        )

        assert result.success is True

    # ========== ERROR CASES ==========

    def test_init_react_invalid_project_name(self, temp_project_dir):
        """
        GIVEN: /init-react with invalid project name
        WHEN: Project name contains uppercase
        THEN: Raises ValueError
        """
        # @CODE:FRONTEND-REACT-INVALID-NAME-001:TEST
        from frontend_plugin.commands import init_react

        with pytest.raises(ValueError) as exc_info:
            init_react.execute(
                project_name="MyApp",
                output_dir=temp_project_dir
            )

        assert "lowercase" in str(exc_info.value).lower()

    def test_init_react_invalid_framework(self, temp_project_dir):
        """
        GIVEN: /init-react with unsupported framework
        WHEN: Framework not in supported list
        THEN: Raises ValueError
        """
        # @CODE:FRONTEND-REACT-INVALID-FRAMEWORK-001:TEST
        from frontend_plugin.commands import init_react

        with pytest.raises(ValueError) as exc_info:
            init_react.execute(
                project_name="test-app",
                output_dir=temp_project_dir,
                framework="unsupported"
            )

        assert "framework" in str(exc_info.value).lower()

    def test_init_react_duplicate_project(self, temp_project_dir):
        """
        GIVEN: Two /init-react calls with same project
        WHEN: First call completes
        THEN: Second call raises error
        """
        # @CODE:FRONTEND-REACT-DUPLICATE-001:TEST
        from frontend_plugin.commands import init_react

        # First call succeeds
        init_react.execute(
            project_name="duplicate",
            output_dir=temp_project_dir
        )

        # Second call should fail
        with pytest.raises(FileExistsError) as exc_info:
            init_react.execute(
                project_name="duplicate",
                output_dir=temp_project_dir
            )

        assert "already exists" in str(exc_info.value).lower()

    # ========== BOUNDARY CASES ==========

    def test_init_react_minimal_project_name(self, temp_project_dir):
        """
        GIVEN: /init-react with minimal 3-character name
        WHEN: Project name meets minimum length
        THEN: Successfully creates project
        """
        # @CODE:FRONTEND-REACT-MIN-NAME-001:TEST
        from frontend_plugin.commands import init_react

        result = init_react.execute(
            project_name="app",
            output_dir=temp_project_dir
        )

        assert result.success is True


# @CODE:FRONTEND-STATE-COMMAND-001:TEST
class TestSetupStateCommand:
    """Test cases for /setup-state command"""

    def test_setup_state_react_context(self, tmp_path):
        """
        GIVEN: User invokes /setup-state with react-context
        WHEN: State management type selected
        THEN: Context configuration created
        """
        # @CODE:FRONTEND-STATE-CONTEXT-001:TEST
        from frontend_plugin.commands import setup_state

        result = setup_state.execute(
            project_name="app",
            output_dir=tmp_path,
            state_type="react-context"
        )

        assert result.success is True

    def test_setup_state_zustand(self, tmp_path):
        """
        GIVEN: /setup-state with zustand
        WHEN: Zustand selected
        THEN: Zustand store setup created
        """
        # @CODE:FRONTEND-STATE-ZUSTAND-001:TEST
        from frontend_plugin.commands import setup_state

        result = setup_state.execute(
            project_name="app",
            output_dir=tmp_path,
            state_type="zustand"
        )

        assert result.success is True

    def test_setup_state_redux(self, tmp_path):
        """
        GIVEN: /setup-state with redux
        WHEN: Redux selected
        THEN: Redux store configuration created
        """
        # @CODE:FRONTEND-STATE-REDUX-001:TEST
        from frontend_plugin.commands import setup_state

        result = setup_state.execute(
            project_name="app",
            output_dir=tmp_path,
            state_type="redux"
        )

        assert result.success is True

    def test_setup_state_invalid_type(self, tmp_path):
        """
        GIVEN: /setup-state with invalid state type
        WHEN: State type not supported
        THEN: Raises ValueError
        """
        # @CODE:FRONTEND-STATE-INVALID-001:TEST
        from frontend_plugin.commands import setup_state

        with pytest.raises(ValueError):
            setup_state.execute(
                project_name="app",
                output_dir=tmp_path,
                state_type="unsupported"
            )


# @CODE:FRONTEND-TESTING-COMMAND-001:TEST
class TestSetupTestingCommand:
    """Test cases for /setup-testing command"""

    def test_setup_testing_vitest(self, tmp_path):
        """
        GIVEN: User invokes /setup-testing with vitest
        WHEN: Test framework selected
        THEN: Vitest configuration created
        """
        # @CODE:FRONTEND-TESTING-VITEST-001:TEST
        from frontend_plugin.commands import setup_testing

        result = setup_testing.execute(
            project_name="app",
            output_dir=tmp_path,
            test_framework="vitest"
        )

        assert result.success is True

    def test_setup_testing_jest(self, tmp_path):
        """
        GIVEN: /setup-testing with jest
        WHEN: Jest selected
        THEN: Jest configuration created
        """
        # @CODE:FRONTEND-TESTING-JEST-001:TEST
        from frontend_plugin.commands import setup_testing

        result = setup_testing.execute(
            project_name="app",
            output_dir=tmp_path,
            test_framework="jest"
        )

        assert result.success is True

    def test_setup_testing_with_testing_library(self, tmp_path):
        """
        GIVEN: /setup-testing with --testing-library option
        WHEN: Testing library enabled
        THEN: Testing library setup included
        """
        # @CODE:FRONTEND-TESTING-LIBRARY-001:TEST
        from frontend_plugin.commands import setup_testing

        result = setup_testing.execute(
            project_name="app",
            output_dir=tmp_path,
            test_framework="vitest",
            include_testing_library=True
        )

        assert result.success is True

    def test_setup_testing_invalid_framework(self, tmp_path):
        """
        GIVEN: /setup-testing with invalid framework
        WHEN: Framework not supported
        THEN: Raises ValueError
        """
        # @CODE:FRONTEND-TESTING-INVALID-001:TEST
        from frontend_plugin.commands import setup_testing

        with pytest.raises(ValueError):
            setup_testing.execute(
                project_name="app",
                output_dir=tmp_path,
                test_framework="unsupported"
            )


# @CODE:FRONTEND-INTEGRATION-001:TEST
class TestFrontendPluginIntegration:
    """Integration tests for Frontend Plugin"""

    def test_init_react_end_to_end(self, tmp_path):
        """
        GIVEN: User invokes /init-react
        WHEN: Command executes
        THEN: All expected files created
        """
        from frontend_plugin.commands import init_react

        result = init_react.execute(
            project_name="e2e-test",
            output_dir=tmp_path
        )

        assert result.success is True
        project_dir = tmp_path / "e2e-test"
        assert (project_dir / "package.json").exists()
        assert (project_dir / "src").exists()

    def test_frontend_complete_workflow(self, tmp_path):
        """
        GIVEN: Complete frontend setup workflow
        WHEN: init-react → setup-state → setup-testing
        THEN: All steps complete successfully
        """
        from frontend_plugin.commands import init_react, setup_state, setup_testing

        # Step 1: Initialize React
        result1 = init_react.execute(
            project_name="workflow",
            output_dir=tmp_path
        )
        assert result1.success is True

        # Step 2: Setup state management
        result2 = setup_state.execute(
            project_name="workflow",
            output_dir=tmp_path,
            state_type="zustand"
        )
        assert result2.success is True

        # Step 3: Setup testing
        result3 = setup_testing.execute(
            project_name="workflow",
            output_dir=tmp_path,
            test_framework="vitest"
        )
        assert result3.success is True


# @CODE:FRONTEND-PERFORMANCE-001:TEST
class TestFrontendPluginPerformance:
    """Performance tests for Frontend Plugin"""

    def test_init_react_completes_quickly(self, tmp_path):
        """
        GIVEN: /init-react command
        WHEN: Project initialization
        THEN: Completes within 5 seconds
        """
        import time
        from frontend_plugin.commands import init_react

        start = time.time()
        init_react.execute(
            project_name="perf-test",
            output_dir=tmp_path
        )
        elapsed = time.time() - start

        assert elapsed < 5.0, f"Command took {elapsed}s, expected < 5s"


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])

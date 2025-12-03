"""
Tests for configuration generation phase preserving version field

"""

import json
import tempfile
from pathlib import Path


class TestPhaseExecutorVersion:
    """Test configuration generation phase preserving version field"""

    def test_phase_4_preserves_existing_version_field(self, tmp_path: Path) -> None:
        """
        GIVEN: Existing .moai/config/config.json with moai.version field
        WHEN: Phase 4 (configuration generation) executes
        THEN: Should preserve existing moai.version field
        """
        from moai_adk.core.project.phase_executor import PhaseExecutor
        from moai_adk.core.project.validator import ProjectValidator

        # Create existing config with custom version in the correct directory structure
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True, exist_ok=True)
        config_path = config_dir / "config.json"

        existing_config = {
            "moai": {"version": "1.5.0-custom", "update_check_frequency": "daily"},
            "project": {"name": "TestProject", "mode": "team"},
        }
        config_path.write_text(json.dumps(existing_config, indent=2, ensure_ascii=False))

        # Create phase executor
        validator = ProjectValidator()
        executor = PhaseExecutor(validator)

        # New config being passed to phase 4
        new_config = {"project": {"name": "TestProject", "mode": "team"}}

        # Execute phase 4
        result = executor.execute_configuration_phase(tmp_path, new_config)

        # Verify phase 4 succeeded
        assert len(result) == 1, f"Phase 4 should return one file, got {len(result)}"
        actual_config_path = tmp_path / ".moai" / "config" / "config.json"
        assert str(actual_config_path) in result, f"Should return actual config file path {actual_config_path}"

        # Read and verify preserved config
        actual_config_path = tmp_path / ".moai" / "config" / "config.json"
        with open(actual_config_path, "r", encoding="utf-8") as f:
            final_config = json.load(f)

        # Verify that the existing version field is preserved
        assert "moai" in final_config, "Final config should have 'moai' section"
        assert "version" in final_config["moai"], "moai section should have 'version' field"
        assert (
            final_config["moai"]["version"] == "1.5.0-custom"
        ), f"Should preserve custom version '1.5.0-custom', got {final_config['moai']['version']}"

    def test_phase_4_merges_new_config_with_existing_version(self, tmp_path: Path) -> None:
        """
        GIVEN: Existing .moai/config/config.json with version and new config with other fields
        WHEN: Phase 4 (configuration generation) executes
        THEN: Should merge both configs, preserving version
        """
        from moai_adk.core.project.phase_executor import PhaseExecutor
        from moai_adk.core.project.validator import ProjectValidator

        # Create existing config with version in the correct directory structure
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True, exist_ok=True)
        config_path = config_dir / "config.json"

        existing_config = {
            "moai": {"version": "2.0.0-existing", "update_check_frequency": "weekly"},
            "project": {"name": "TestProject", "language": "python"},
        }
        config_path.write_text(json.dumps(existing_config, indent=2, ensure_ascii=False))

        # Create phase executor
        validator = ProjectValidator()
        executor = PhaseExecutor(validator)

        # New config being passed to phase 4
        new_config = {"project": {"mode": "team", "test_coverage_target": 85}, "constitution": {"enforce_tdd": True}}

        # Execute phase 4
        result = executor.execute_configuration_phase(tmp_path, new_config)

        # Verify phase 4 succeeded
        actual_config_path = tmp_path / ".moai" / "config" / "config.json"
        assert str(actual_config_path) in result, f"Should return actual config file path {actual_config_path}"

        # Read and verify merged config
        with open(actual_config_path, "r", encoding="utf-8") as f:
            final_config = json.load(f)

        # Verify that the existing version field is preserved
        assert "moai" in final_config, "Final config should have 'moai' section"
        assert "version" in final_config["moai"], "moai section should have 'version' field"
        assert (
            final_config["moai"]["version"] == "2.0.0-existing"
        ), f"Should preserve existing version '2.0.0-existing', got {final_config['moai']['version']}"

        # Should have new fields
        assert "project" in final_config, "Should have project section"
        assert (
            final_config["project"]["mode"] == "team"
        ), f"Should have new project mode 'team', got {final_config['project'].get('mode')}"
        assert "constitution" in final_config, "Should have constitution section"

    def test_phase_4_handles_version_field_priority_correctly(self, tmp_path: Path) -> None:
        """
        GIVEN: New config has moai.version and existing config has different version
        WHEN: Phase 4 (configuration generation) executes
        THEN: Should prioritize existing version (user customizations)
        """
        from moai_adk.core.project.phase_executor import PhaseExecutor
        from moai_adk.core.project.validator import ProjectValidator

        # Create existing config with custom version in the correct directory structure
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True, exist_ok=True)
        config_path = config_dir / "config.json"

        existing_config = {"moai": {"version": "3.1.0-user-custom", "update_check_frequency": "monthly"}}
        config_path.write_text(json.dumps(existing_config, indent=2, ensure_ascii=False))

        # Create phase executor
        validator = ProjectValidator()
        executor = PhaseExecutor(validator)

        # New config with different version (should be ignored)
        new_config = {"moai": {"version": "1.0.0-template", "update_check_frequency": "daily"}}

        # Execute phase 4
        executor.execute_configuration_phase(tmp_path, new_config)

        # Read and verify final config
        actual_config_path = tmp_path / ".moai" / "config" / "config.json"
        with open(actual_config_path, "r", encoding="utf-8") as f:
            final_config = json.load(f)

        # This assertion should pass - Phase 4 should preserve existing version
        # It should keep the user's custom version, not the template version
        assert (
            final_config["moai"]["version"] == "3.1.0-user-custom"
        ), f"Should preserve user's custom version '3.1.0-user-custom', got {final_config['moai']['version']}"

    def test_phase_4_preserves_version_during_reinitialization(self, tmp_path: Path) -> None:
        """
        GIVEN: Reinitialization with existing .moai/config.json containing custom version
        WHEN: Phase 4 (configuration generation) executes with reinit=True
        THEN: Should preserve the custom version field
        """
        from moai_adk.core.project.phase_executor import PhaseExecutor
        from moai_adk.core.project.validator import ProjectValidator

        # Create existing config with custom version in the correct directory structure
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True, exist_ok=True)
        config_path = config_dir / "config.json"

        custom_version = "5.0.0-my-custom-version"
        existing_config = {
            "moai": {"version": custom_version, "version_check": {"enabled": True, "cache_ttl_hours": 48}},
            "project": {"name": "MyCustomProject", "mode": "team"},
        }
        config_path.write_text(json.dumps(existing_config, indent=2, ensure_ascii=False))

        # Create phase executor
        validator = ProjectValidator()
        executor = PhaseExecutor(validator)

        # New config for reinitialization
        new_config = {"project": {"name": "MyCustomProject", "locale": "en"}}  # Same name, should not overwrite

        # Execute phase 4 (simulating reinitialization)
        executor.execute_configuration_phase(tmp_path, new_config)

        # Read and verify final config
        actual_config_path = tmp_path / ".moai" / "config" / "config.json"
        with open(actual_config_path, "r", encoding="utf-8") as f:
            final_config = json.load(f)

        # Verify that the custom version is preserved during reinitialization
        assert "moai" in final_config, "Final config should have 'moai' section"
        assert "version" in final_config["moai"], "moai section should have 'version' field"
        assert (
            final_config["moai"]["version"] == custom_version
        ), f"Should preserve custom version '{custom_version}', got {final_config['moai']['version']}"

        # Should have new locale field
        assert (
            final_config["project"]["locale"] == "en"
        ), f"Should have new locale field 'en', got {final_config['project'].get('locale')}"

    def test_phase_4_version_field_validation(self, tmp_path: Path) -> None:
        """
        GIVEN: Existing config with invalid version field and new config
        WHEN: Phase 4 (configuration generation) executes
        THEN: Should preserve the version field even if invalid
        """
        from moai_adk.core.project.phase_executor import PhaseExecutor
        from moai_adk.core.project.validator import ProjectValidator

        # Create existing config with invalid version (should still be preserved)
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True, exist_ok=True)
        config_path = config_dir / "config.json"

        existing_config = {"moai": {"version": "invalid-version-string", "update_check_frequency": "daily"}}
        config_path.write_text(json.dumps(existing_config, indent=2, ensure_ascii=False))

        # Create phase executor
        validator = ProjectValidator()
        executor = PhaseExecutor(validator)

        # New config
        new_config = {"project": {"name": "TestProject"}}

        # Execute phase 4
        executor.execute_configuration_phase(tmp_path, new_config)

        # Read and verify final config
        with open(config_path, "r", encoding="utf-8") as f:
            final_config = json.load(f)

        # RED: These assertions will fail because Phase 4 doesn't preserve invalid version fields
        assert "moai" in final_config, "Final config should have 'moai' section"
        assert "version" in final_config["moai"], "moai section should have 'version' field"
        assert (
            final_config["moai"]["version"] == "invalid-version-string"
        ), f"Should preserve invalid version string, got {final_config['moai']['version']}"

    def test_phase_4_with_missing_moai_section(self, tmp_path: Path) -> None:
        """
        GIVEN: Existing config without moai section
        WHEN: Phase 4 (configuration generation) executes
        THEN: Should not add version field (create new config)
        """
        # Create existing config without moai section in the correct directory structure
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True, exist_ok=True)
        config_path = config_dir / "config.json"

        existing_config = {"project": {"name": "TestProject"}}
        config_path.write_text(json.dumps(existing_config, indent=2, ensure_ascii=False))

        from moai_adk.core.project.phase_executor import PhaseExecutor
        from moai_adk.core.project.validator import ProjectValidator

        # Create phase executor
        validator = ProjectValidator()
        executor = PhaseExecutor(validator)

        # New config
        new_config = {"project": {"name": "TestProject", "mode": "team"}}

        # Execute phase 4
        executor.execute_configuration_phase(tmp_path, new_config)

        # Read and verify final config
        with open(config_path, "r", encoding="utf-8") as f:
            final_config = json.load(f)

        # Should create moai section with version from template
        assert "moai" in final_config, "Final config should have 'moai' section"
        assert "version" in final_config["moai"], "moai section should have 'version' field"
        assert (
            final_config["moai"]["version"] != "unknown"
        ), f"Version should not be 'unknown', got {final_config['moai']['version']}"

    def test_phase_4_version_field_preservation_with_multiple_changes(self, tmp_path: Path) -> None:
        """
        GIVEN: Existing config with version and multiple other fields being changed
        WHEN: Phase 4 (configuration generation) executes
        THEN: Should preserve version and update other fields
        """
        from moai_adk.core.project.phase_executor import PhaseExecutor
        from moai_adk.core.project.validator import ProjectValidator

        # Create existing config with version and other fields in the correct directory structure
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True, exist_ok=True)
        config_path = config_dir / "config.json"

        existing_config = {
            "moai": {"version": "4.2.1-stable", "update_check_frequency": "daily", "version_check": {"enabled": False}},
            "project": {"name": "TestProject", "language": "python"},
            "constitution": {"enforce_tdd": True, "test_coverage_target": 80},
        }
        config_path.write_text(json.dumps(existing_config, indent=2, ensure_ascii=False))

        # Create phase executor
        validator = ProjectValidator()
        executor = PhaseExecutor(validator)

        # New config with changes
        new_config = {
            "project": {"name": "TestProject", "mode": "team", "locale": "en"},
            "constitution": {"enforce_tdd": True, "test_coverage_target": 85},  # Changed from 80 to 85
            "language": {"conversation_language": "en"},
        }

        # Execute phase 4
        executor.execute_configuration_phase(tmp_path, new_config)

        # Read and verify final config
        actual_config_path = tmp_path / ".moai" / "config" / "config.json"
        with open(actual_config_path, "r", encoding="utf-8") as f:
            final_config = json.load(f)

        # Verify that the version is preserved during multiple changes
        assert "moai" in final_config, "Final config should have 'moai' section"
        assert "version" in final_config["moai"], "moai section should have 'version' field"
        assert (
            final_config["moai"]["version"] == "4.2.1-stable"
        ), f"Should preserve version '4.2.1-stable', got {final_config['moai']['version']}"

        # Should have new fields updated
        assert (
            final_config["project"]["mode"] == "team"
        ), f"Should have updated mode 'team', got {final_config['project'].get('mode')}"
        assert (
            final_config["constitution"]["test_coverage_target"] == 85
        ), f"Should have updated coverage target 85, got {final_config['constitution'].get('test_coverage_target')}"
        assert "language" in final_config, "Should have language section"

    def test_phase_4_version_field_case_sensitivity(self, tmp_path: Path) -> None:
        """
        GIVEN: Existing config with version in different case formats
        WHEN: Phase 4 (configuration generation) executes
        THEN: Should preserve version exactly as specified
        """
        from moai_adk.core.project.phase_executor import PhaseExecutor
        from moai_adk.core.project.validator import ProjectValidator

        test_versions = ["1.2.3", "v1.2.3", "2.0.0-BETA", "3.1.0-rc.1", "4.0.0-dev"]

        for version in test_versions:
            with tempfile.TemporaryDirectory() as tmpdir:
                tmp_path = Path(tmpdir)
                config_path = tmp_path / ".moai" / "config.json"
                config_path.parent.mkdir(parents=True, exist_ok=True)

                # Create existing config with specific version
                existing_config = {"moai": {"version": version, "update_check_frequency": "daily"}}
                config_path.write_text(json.dumps(existing_config, indent=2, ensure_ascii=False))

                # Create phase executor
                validator = ProjectValidator()
                executor = PhaseExecutor(validator)

                # New config
                new_config = {"project": {"name": "TestProject"}}

                # Execute phase 4
                executor.execute_configuration_phase(tmp_path, new_config)

                # Read and verify final config
                with open(config_path, "r", encoding="utf-8") as f:
                    final_config = json.load(f)

                # RED: These assertions will fail because Phase 4 doesn't preserve version
                assert "moai" in final_config, "Final config should have 'moai' section"
                assert "version" in final_config["moai"], "moai section should have 'version' field"
                assert (
                    final_config["moai"]["version"] == version
                ), f"Should preserve exact version '{version}', got '{final_config['moai']['version']}'"

    def test_phase_4_version_field_with_special_characters(self, tmp_path: Path) -> None:
        """
        GIVEN: Existing config with version containing special characters
        WHEN: Phase 4 (configuration generation) executes
        THEN: Should preserve version exactly
        """
        from moai_adk.core.project.phase_executor import PhaseExecutor
        from moai_adk.core.project.validator import ProjectValidator

        special_versions = ["1.2.3+build.123", "2.0.0-alpha.1", "3.1.0-rc.1+build.456", "4.0.0-dev.1+build.789"]

        for version in special_versions:
            with tempfile.TemporaryDirectory() as tmpdir:
                tmp_path = Path(tmpdir)
                config_path = tmp_path / ".moai" / "config.json"
                config_path.parent.mkdir(parents=True, exist_ok=True)

                # Create existing config with special version
                existing_config = {"moai": {"version": version, "update_check_frequency": "daily"}}
                config_path.write_text(json.dumps(existing_config, indent=2, ensure_ascii=False))

                # Create phase executor
                validator = ProjectValidator()
                executor = PhaseExecutor(validator)

                # New config
                new_config = {"project": {"name": "TestProject"}}

                # Execute phase 4
                executor.execute_configuration_phase(tmp_path, new_config)

                # Read and verify final config
                with open(config_path, "r", encoding="utf-8") as f:
                    final_config = json.load(f)

                # RED: These assertions will fail because Phase 4 doesn't preserve special version formats
                assert "moai" in final_config, "Final config should have 'moai' section"
                assert "version" in final_config["moai"], "moai section should have 'version' field"
                assert (
                    final_config["moai"]["version"] == version
                ), f"Should preserve special version '{version}', got '{final_config['moai']['version']}'"

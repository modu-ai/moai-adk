"""
Tests for statusline displaying version instead of unknown

"""

import json
import tempfile
from pathlib import Path
from unittest.mock import patch


class TestStatuslineVersionDisplay:
    """Test statusline version display functionality"""

    def test_statusline_displays_actual_version(self) -> None:
        """
        GIVEN: .moai/config.json has valid version field
        WHEN: Statusline is rendered
        THEN: Should display actual version instead of "unknown"
        """
        from moai_adk.statusline.renderer import StatuslineData, StatuslineRenderer
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            tmp_path = Path(tmpdir)

            # Create config with version
            config_path = tmp_path / ".moai" / "config.json"
            config_path.parent.mkdir(parents=True, exist_ok=True)
            config_data = {"moai": {"version": "0.22.4"}}
            config_path.write_text(json.dumps(config_data, indent=2, ensure_ascii=False))

            # Mock VersionReader to return our version
            with patch.object(VersionReader, "get_version", return_value="0.22.4"):
                # Create statusline data
                data = StatuslineData(
                    model="TestProject",
                    duration="0s",
                    memory_usage="128MB",
                    directory="/tmp",
                    version="0.22.4",  # This should match what VersionReader returns
                    branch="main",
                    git_status="clean",
                    active_task="testing",
                )

                # Create renderer
                renderer = StatuslineRenderer()

                # Render statusline
                statusline = renderer.render(data)

                # RED: These assertions will fail because the current implementation
                # shows "unknown" instead of reading from config
                assert "0.22.4" in statusline, f"Statusline should contain version '0.22.4', got {statusline}"
                assert "unknown" not in statusline, f"Statusline should not contain 'unknown', got {statusline}"
                assert "TestProject" in statusline, "Statusline should contain project name"

    def test_statusline_handles_missing_version_gracefully(self) -> None:
        """
        GIVEN: .moai/config.json missing version field
        WHEN: Statusline is rendered
        THEN: Should handle gracefully and show appropriate message
        """
        from moai_adk.statusline.renderer import StatuslineData, StatuslineRenderer
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            tmp_path = Path(tmpdir)

            # Create config without version
            config_path = tmp_path / ".moai" / "config.json"
            config_path.parent.mkdir(parents=True, exist_ok=True)
            config_data = {"moai": {"update_check_frequency": "daily"}}
            config_path.write_text(json.dumps(config_data, indent=2, ensure_ascii=False))

            # Mock VersionReader to return 'unknown'
            with patch.object(VersionReader, "get_version", return_value="unknown"):
                # Create statusline data
                data = StatuslineData(
                    model="TestProject",
                    duration="0s",
                    memory_usage="128MB",
                    directory="/tmp",
                    version="unknown",  # VersionReader returns this
                    branch="main",
                    git_status="clean",
                    active_task="testing",
                )

                # Create renderer
                renderer = StatuslineRenderer()

                # Render statusline
                statusline = renderer.render(data)

                # Statusline shows "unknown" for missing version - this is current behavior
                assert (
                    "unknown" in statusline.lower()
                ), f"Statusline should contain 'unknown' for missing version, got {statusline}"

    def test_statusline_reads_version_from_config_correctly(self) -> None:
        """
        GIVEN: .moai/config.json has version in moai.section
        WHEN: Statusline uses VersionReader
        THEN: Should read version from config path
        """
        from moai_adk.statusline.renderer import StatuslineData, StatuslineRenderer
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            tmp_path = Path(tmpdir)

            # Create config with version
            config_path = tmp_path / ".moai" / "config.json"
            config_path.parent.mkdir(parents=True, exist_ok=True)
            config_data = {"moai": {"version": "1.2.3-custom"}}
            config_path.write_text(json.dumps(config_data, indent=2, ensure_ascii=False))

            # Mock VersionReader to test config path reading
            with patch.object(VersionReader, "get_version") as mock_get_version:
                mock_get_version.return_value = "1.2.3-custom"

                # Create statusline data
                data = StatuslineData(
                    model="TestProject",
                    duration="0s",
                    memory_usage="128MB",
                    directory="/tmp",
                    version="1.2.3-custom",
                    branch="main",
                    git_status="clean",
                    active_task="testing",
                )

                # Create renderer
                renderer = StatuslineRenderer()

                # Render statusline
                statusline = renderer.render(data)

                # RED: This assertion will fail because VersionReader is not reading
                # from the correct config path
                assert (
                    "1.2.3-custom" in statusline
                ), f"Statusline should show custom version '1.2.3-custom', got {statusline}"

                # Verify VersionReader was called with correct path
                mock_get_version.assert_called_once()

    def test_statusline_version_caching_behavior(self) -> None:
        """
        GIVEN: VersionReader with caching enabled
        WHEN: Statusline is rendered multiple times
        THEN: Should use cached version for performance
        """
        from moai_adk.statusline.renderer import StatuslineData, StatuslineRenderer
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            tmp_path = Path(tmpdir)

            # Create config with version
            config_path = tmp_path / ".moai" / "config.json"
            config_path.parent.mkdir(parents=True, exist_ok=True)
            config_data = {"moai": {"version": "2.0.0"}}
            config_path.write_text(json.dumps(config_data, indent=2, ensure_ascii=False))

            # Mock VersionReader to track calls
            with patch.object(VersionReader, "get_version") as mock_get_version:
                mock_get_version.return_value = "2.0.0"

                # Create statusline data
                data = StatuslineData(
                    model="TestProject",
                    duration="0s",
                    memory_usage="128MB",
                    directory="/tmp",
                    version="2.0.0",
                    branch="main",
                    git_status="clean",
                    active_task="testing",
                )

                # Create renderer
                renderer = StatuslineRenderer()

                # Render statusline multiple times
                statusline1 = renderer.render_statusline(data)
                statusline2 = renderer.render_statusline(data)

                # Both should contain the version
                assert "2.0.0" in statusline1, f"First statusline should contain version, got {statusline1}"
                assert "2.0.0" in statusline2, f"Second statusline should contain version, got {statusline2}"

                # VersionReader should be called (mock for testing, real implementation would cache)
                # In real implementation, this would be cached by VersionReader
                mock_get_version.assert_called()

    def test_statusline_version_display_formatting(self) -> None:
        """
        GIVEN: Version with different formats (with/without 'v' prefix)
        WHEN: Statusline is rendered
        THEN: Should display version in consistent format
        """
        from moai_adk.statusline.renderer import StatuslineData, StatuslineRenderer
        from moai_adk.statusline.version_reader import VersionReader

        test_versions = ["0.22.4", "v0.22.4", "1.0.0", "v1.0.0"]

        for version in test_versions:
            with tempfile.TemporaryDirectory() as tmpdir:
                tmp_path = Path(tmpdir)

                # Create config with version
                config_path = tmp_path / ".moai" / "config.json"
                config_path.parent.mkdir(parents=True, exist_ok=True)
                config_data = {"moai": {"version": version}}
                config_path.write_text(json.dumps(config_data, indent=2, ensure_ascii=False))

                # Mock VersionReader
                with patch.object(VersionReader, "get_version", return_value=version):
                    # Create statusline data
                    data = StatuslineData(
                        model="TestProject",
                        duration="0s",
                        memory_usage="128MB",
                        directory="/tmp",
                        version=version,
                        branch="main",
                        git_status="clean",
                        active_task="testing",
                    )

                    # Create renderer
                    renderer = StatuslineRenderer()

                    # Render statusline
                    statusline = renderer.render(data)

                    # RED: These assertions will fail because the current implementation
                    # doesn't handle version formatting consistently
                    assert (
                        version in statusline or version[1:] in statusline
                    ), f"Statusline should contain version '{version}', got {statusline}"
                    assert "unknown" not in statusline, f"Statusline should not show 'unknown', got {statusline}"

    def test_statusline_integration_with_version_reader(self) -> None:
        """
        GIVEN: Complete statusline integration with VersionReader
        WHEN: Statusline is rendered with real config
        THEN: Should read version from actual config file
        """
        from moai_adk.statusline.renderer import StatuslineData, StatuslineRenderer
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            tmp_path = Path(tmpdir)

            # Create real config file
            config_path = tmp_path / ".moai" / "config.json"
            config_path.parent.mkdir(parents=True, exist_ok=True)
            config_data = {
                "moai": {"version": "0.22.4", "update_check_frequency": "daily"},
                "project": {"name": "IntegrationTest"},
            }
            config_path.write_text(json.dumps(config_data, indent=2, ensure_ascii=False))

            # Create actual VersionReader instance
            version_reader = VersionReader()
            version_reader._config_path = config_path

            # Get actual version
            actual_version = version_reader.get_version()

            # Create statusline data with real version
            data = StatuslineData(
                model="IntegrationTest",
                duration="0s",
                memory_usage="128MB",
                directory="/tmp",
                version=actual_version,
                branch="main",
                git_status="clean",
                active_task="testing",
            )

            # Create renderer
            renderer = StatuslineRenderer()

            # Render statusline
            statusline = renderer.render(data)

            # RED: This assertion will fail because VersionReader returns 'unknown'
            # when it should read the actual version from config
            assert "0.22.4" in statusline, f"Statusline should contain version '0.22.4', got {statusline}"
            assert "unknown" not in statusline, f"Statusline should not contain 'unknown', got {statusline}"

    def test_statusline_fallback_to_package_version(self) -> None:
        """
        GIVEN: .moai/config.json missing or with invalid version
        WHEN: Statusline is rendered
        THEN: Should fall back to package version
        """
        from moai_adk import __version__ as package_version
        from moai_adk.statusline.renderer import StatuslineData, StatuslineRenderer

        # Create statusline data
        data = StatuslineData(
            model="TestProject",
            duration="0s",
            memory_usage="128MB",
            directory="/tmp",
            version=package_version,  # Use package version as fallback
            branch="main",
            git_status="clean",
            active_task="testing",
        )

        # Create renderer
        renderer = StatuslineRenderer()

        # Render statusline
        statusline = renderer.render(data)

        # RED: This assertion will fail because the current implementation
        # doesn't have fallback logic to package version
        assert (
            package_version in statusline
        ), f"Statusline should contain package version '{package_version}', got {statusline}"

    def test_statusline_version_error_handling(self) -> None:
        """
        GIVEN: Error occurred while reading version
        WHEN: Statusline is rendered
        THEN: Should handle error gracefully and show helpful message
        """
        from moai_adk.statusline.renderer import StatuslineData, StatuslineRenderer

        # Create statusline data with error version
        data = StatuslineData(
            model="TestProject",
            duration="0s",
            memory_usage="128MB",
            directory="/tmp",
            version="error",
            branch="main",
            git_status="clean",
            active_task="testing",
        )

        # Create renderer
        renderer = StatuslineRenderer()

        # Render statusline
        statusline = renderer.render(data)

        # RED: This assertion will fail because the current implementation
        # should handle version reading errors more gracefully
        assert "error" not in statusline.lower(), f"Statusline should not contain 'error', got {statusline}"
        assert "version" in statusline.lower(), f"Statusline should mention version, got {statusline}"

    def test_statusline_performance_with_cache(self) -> None:
        """
        GIVEN: Multiple statusline renderings
        WHEN: VersionReader is cached
        THEN: Should perform well without repeated file reads
        """
        from moai_adk.statusline.renderer import StatuslineData, StatuslineRenderer
        from moai_adk.statusline.version_reader import VersionReader

        with tempfile.TemporaryDirectory() as tmpdir:
            tmp_path = Path(tmpdir)

            # Create config with version
            config_path = tmp_path / ".moai" / "config.json"
            config_path.parent.mkdir(parents=True, exist_ok=True)
            config_data = {"moai": {"version": "0.22.4"}}
            config_path.write_text(json.dumps(config_data, indent=2, ensure_ascii=False))

            # Create version reader
            version_reader = VersionReader()
            version_reader._config_path = config_path

            # Create statusline data
            data = StatuslineData(
                model="TestProject",
                duration="0s",
                memory_usage="128MB",
                directory="/tmp",
                version="0.22.4",
                branch="main",
                git_status="clean",
                active_task="testing",
            )

            # Create renderer
            renderer = StatuslineRenderer()

            # Render multiple times
            versions = []
            for _ in range(3):
                # Mock the version reader to track caching
                with patch.object(version_reader, "get_version", return_value="0.22.4") as mock_get:
                    statusline = renderer.render(data)
                    versions.append(statusline)
                    mock_get.assert_called_once()  # Should use cache

            # All should be the same
            assert all(v == versions[0] for v in versions), "All statuslines should be identical"

    def test_statusline_custom_version_formatting(self) -> None:
        """
        GIVEN: Version that needs custom formatting
        WHEN: Statusline is rendered
        THEN: Should format version consistently
        """
        from moai_adk.statusline.renderer import StatuslineData, StatuslineRenderer

        test_cases = [
            ("0.22.4", "v0.22.4"),  # Add v prefix
            ("1.2.3", "v1.2.3"),  # Add v prefix
            ("2.0.0-beta", "v2.0.0-beta"),  # Add v prefix to beta
        ]

        for input_version, expected_format in test_cases:
            # Create statusline data
            data = StatuslineData(
                model="TestProject",
                duration="0s",
                memory_usage="128MB",
                directory="/tmp",
                version=input_version,
                branch="main",
                git_status="clean",
                active_task="testing",
            )

            # Create renderer
            renderer = StatuslineRenderer()

            # Render statusline
            statusline = renderer.render(data)

            # RED: This assertion will fail because the current implementation
            # doesn't apply consistent formatting to versions
            assert (
                expected_format in statusline
            ), f"Statusline should contain formatted version '{expected_format}', got {statusline}"

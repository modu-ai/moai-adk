from pathlib import Path

from moai_adk.install.resource_manager import ResourceManager


def test_copy_moai_resources_exclude_templates(tmp_path: Path):
    rm = ResourceManager()
    dest = tmp_path

    files = rm.copy_moai_resources(dest, overwrite=True, exclude_templates=True)

    moai_dir = dest / ".moai"
    assert moai_dir.exists()
    # _templates should be skipped
    assert not (moai_dir / "_templates").exists()
    # config.json from package should exist
    assert (moai_dir / "config.json").exists()


def test_clean_installation_validation(tmp_path: Path):
    """@TEST:TEMPLATE-CLEAN-001 Test clean installation validation"""
    rm = ResourceManager()
    dest = tmp_path

    # Install MoAI resources
    files = rm.copy_moai_resources(dest, overwrite=True)

    moai_dir = dest / ".moai"
    assert moai_dir.exists()

    # Verify clean installation structure
    specs_dir = moai_dir / "specs"
    if specs_dir.exists():
        # Should only contain .gitkeep file or be empty
        spec_files = [f for f in specs_dir.iterdir() if f.name != ".gitkeep"]
        assert len(spec_files) == 0, (
            f"Found unexpected spec files: {[f.name for f in spec_files]}"
        )

    # Verify tags.json is minimal (< 50 lines)
    tags_file = moai_dir / "indexes" / "tags.json"
    if tags_file.exists():
        line_count = sum(1 for _ in open(tags_file))
        assert line_count < 50, (
            f"tags.json too large: {line_count} lines (expected < 50)"
        )

    # Verify reports directory is clean
    reports_dir = moai_dir / "reports"
    if reports_dir.exists():
        report_files = [f for f in reports_dir.iterdir() if f.name != ".gitkeep"]
        assert len(report_files) == 0, (
            f"Found unexpected report files: {[f.name for f in report_files]}"
        )


def test_validate_clean_installation_method(tmp_path: Path):
    """@TEST:TEMPLATE-VERIFY-001 Test _validate_clean_installation method directly"""
    rm = ResourceManager()
    dest = tmp_path

    # Install MoAI resources first
    rm.copy_moai_resources(dest, overwrite=True)

    moai_dir = dest / ".moai"

    # Test the validation method directly
    is_clean = rm._validate_clean_installation(moai_dir)
    assert is_clean, "Clean installation validation should pass"

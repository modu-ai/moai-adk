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


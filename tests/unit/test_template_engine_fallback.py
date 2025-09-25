import os
from pathlib import Path

import pytest

from moai_adk.core.template_engine import TemplateEngine


def test_template_engine_uses_package_fallback_when_project_templates_missing(
    tmp_path: Path,
):
    project_moai = tmp_path / ".moai"
    project_moai.mkdir(parents=True)
    # intentionally do NOT create _templates

    engine = TemplateEngine(project_moai)

    target = tmp_path / "SPEC-TEST.md"
    ok = engine.create_spec_from_template(
        spec_id="TEST-001",
        spec_name="Fallback Check",
        description="ensure package templates are used when local missing",
        target_path=target,
    )

    assert ok is True
    assert target.exists()
    content = target.read_text(encoding="utf-8")
    # Should contain substituted spec id and name
    assert "TEST-001" in content
    assert "Fallback Check" in content


def test_project_override_takes_precedence_over_package(tmp_path: Path):
    project_moai = tmp_path / ".moai"
    templates_dir = project_moai / "_templates" / "specs"
    templates_dir.mkdir(parents=True)

    # Local override template
    (templates_dir / "spec.template.md").write_text(
        "# LOCAL TEMPLATE: $SPEC_NAME ($SPEC_ID)\n",
        encoding="utf-8",
    )

    engine = TemplateEngine(project_moai)

    target = tmp_path / "SPEC-LOCAL.md"
    ok = engine.create_spec_from_template(
        spec_id="OVR-001",
        spec_name="Local Override",
        description="",
        target_path=target,
    )

    assert ok is True
    assert target.exists()
    content = target.read_text(encoding="utf-8").strip()
    assert content.startswith("# LOCAL TEMPLATE:")
    assert "OVR-001" in content and "Local Override" in content

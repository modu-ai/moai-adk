"""@FEATURE:RESOURCE-VERSION-001 Resource version tracking utilities for MoAI-ADK projects."""

from __future__ import annotations

import json
from datetime import datetime
from pathlib import Path


class ResourceVersionManager:
    """@TASK:RESOURCE-VERSION-MANAGER-001 Read and write MoAI resource/template version metadata."""

    VERSION_RELATIVE_PATH = Path(".moai") / "version.json"

    def __init__(self, project_path: Path | str):
        self.project_path = Path(project_path)
        self.version_path = self.project_path / self.VERSION_RELATIVE_PATH

    def read(self) -> dict[str, str | None]:
        """Return version metadata if available, otherwise defaults."""
        if not self.version_path.exists():
            return {
                "template_version": None,
                "package_version": None,
                "last_updated": None,
            }

        try:
            with open(self.version_path, encoding="utf-8") as fp:
                data = json.load(fp)
        except (json.JSONDecodeError, OSError):
            return {
                "template_version": None,
                "package_version": None,
                "last_updated": None,
            }

        return {
            "template_version": data.get("template_version"),
            "package_version": data.get("package_version"),
            "last_updated": data.get("last_updated"),
        }

    def write(self, template_version: str, package_version: str) -> dict[str, str]:
        """Persist version metadata for the project."""
        data = {
            "template_version": template_version,
            "package_version": package_version,
            "last_updated": datetime.utcnow().isoformat() + "Z",
        }
        self.version_path.parent.mkdir(parents=True, exist_ok=True)
        with open(self.version_path, "w", encoding="utf-8") as fp:
            json.dump(data, fp, indent=2, ensure_ascii=False)
        return data

    def is_outdated(self, expected_template_version: str) -> bool:
        info = self.read()
        current = info.get("template_version")
        return current is not None and current != expected_template_version

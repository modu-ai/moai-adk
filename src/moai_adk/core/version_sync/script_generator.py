#!/usr/bin/env python3
"""
@FEATURE:SCRIPT-GENERATOR-001 Automated Script Generation
Dedicated module for generating version update automation scripts
"""

from pathlib import Path

from moai_adk.utils.logger import get_logger

logger = get_logger(__name__)


class ScriptGenerator:
    """@TASK:SCRIPT-GENERATOR-001 Generates automation scripts for version updates"""

    def __init__(self, project_root: Path):
        """
        Initialize script generator

        Args:
            project_root: Project root directory
        """
        self.project_root = project_root

    def create_version_update_script(self) -> str:
        """
        Generate version update automation script

        Returns:
            Path to generated script
        """
        script_path = self.project_root / "scripts" / "update_version.py"
        script_path.parent.mkdir(parents=True, exist_ok=True)

        script_content = self._generate_script_content()

        with open(script_path, "w", encoding="utf-8") as f:
            f.write(script_content)

        logger.info(f"Version update script created: {script_path}")
        print(f"✅ Version update script created: {script_path}")
        return str(script_path)

    def _generate_script_content(self) -> str:
        """Generate the script content"""
        return '''#!/usr/bin/env python3
"""
MoAI-ADK Version Update Automation Script
Usage: python scripts/update_version.py <new_version>
"""

import sys
import re
from pathlib import Path


def update_version_in_file(file_path: Path, old_version: str, new_version: str) -> bool:
    """Update version information in file"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Version pattern replacements
        patterns = [
            (r'__version__\\s*=\\s*"[^"]*"', '"__VERSION_PLACEHOLDER__"'),
            (r'version\\s*=\\s*"[^"]*"', 'version = "{new_version}"'),
            (r'MoAI-ADK v[0-9]+\\.[0-9]+\\.[0-9]+', 'MoAI-ADK v{new_version}'),
            (r'"moai_version":\\s*"[^"]*"', '"moai_version": "{new_version}"')
        ]

        original_content = content
        for pattern, replacement in patterns:
            rep = replacement
            if replacement == '"__VERSION_PLACEHOLDER__"':
                rep = f'__version__ = "0.1.17"'
            else:
                rep = replacement.format(new_version=new_version)
            content = re.sub(pattern, rep, content)

        if content != original_content:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(content)
            return True

    except Exception as e:
        print(f"Error updating {file_path}: {e}")

    return False


def main() -> None:
    if len(sys.argv) != 2:
        print("Usage: python scripts/update_version.py <new_version>")
        print("Example: python scripts/update_version.py 0.2.0")
        sys.exit(1)

    new_version = sys.argv[1]

    # Validate version format
    if not re.match(r'^[0-9]+\\.[0-9]+\\.[0-9]+$', new_version):
        print("Error: Version must be in format X.Y.Z")
        sys.exit(1)

    print(f"🗿 MoAI-ADK version update: v{new_version}")

    # Execute version synchronization
    from moai_adk.core.version_sync import VersionSyncManager

    # Update _version.py first
    version_file = Path("src/moai_adk/_version.py")
    update_version_in_file(version_file, None, new_version)

    # Synchronize entire project
    sync_manager = VersionSyncManager()
    results = sync_manager.sync_all_versions()

    print(f"✅ Version update completed: v{new_version}")
    print("Next steps:")
    print("1. git add .")
    print("2. git commit -m 'bump version to v{new_version}'")
    print("3. git tag v{new_version}")
    print("4. git push origin main --tags")


if __name__ == "__main__":
    main()
'''
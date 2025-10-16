# @CODE:TEMPLATE-001 | SPEC: SPEC-INIT-003.md | Chain: TEMPLATE-001
"""Template copy and backup processor (SPEC-INIT-003 v0.3.0: preserve user content)."""

from __future__ import annotations

import shutil
from pathlib import Path

from rich.console import Console

from moai_adk.core.template.backup import TemplateBackup
from moai_adk.core.template.merger import TemplateMerger

console = Console()


class TemplateProcessor:
    """Orchestrate template copying and backups."""

    # User data protection paths (never touch) - SPEC-INIT-003 v0.3.0
    PROTECTED_PATHS = [
        ".moai/specs/",  # User SPEC documents
        ".moai/reports/",  # User reports
        ".moai/project/",  # User project documents (product/structure/tech.md)
        ".moai/config.json",  # User configuration (merged via /alfred:9-update flow)
    ]

    # Paths excluded from backups
    BACKUP_EXCLUDE = PROTECTED_PATHS

    def __init__(self, target_path: Path) -> None:
        """Initialize the processor.

        Args:
            target_path: Project path.
        """
        self.target_path = target_path.resolve()
        self.template_root = self._get_template_root()
        self.backup = TemplateBackup(self.target_path)
        self.merger = TemplateMerger(self.target_path)

    def _get_template_root(self) -> Path:
        """Return the template root path."""
        # src/moai_adk/core/template/processor.py → src/moai_adk/templates/
        current_file = Path(__file__).resolve()
        package_root = current_file.parent.parent.parent
        return package_root / "templates"

    def copy_templates(self, backup: bool = True, silent: bool = False) -> None:
        """Copy template files into the project.

        Args:
            backup: Whether to create a backup.
            silent: Reduce log output when True.
        """
        # 1. Create a backup when existing files are present
        if backup and self._has_existing_files():
            backup_path = self.create_backup()
            if not silent:
                console.print(f"💾 Backup created: {backup_path.name}")

        # 2. Copy templates
        if not silent:
            console.print("📄 Copying templates...")

        self._copy_claude(silent)
        self._copy_moai(silent)
        self._copy_claude_md(silent)
        self._copy_gitignore(silent)

        if not silent:
            console.print("✅ Templates copied successfully")

    def _has_existing_files(self) -> bool:
        """Determine whether project files exist (backup decision helper)."""
        return self.backup.has_existing_files()

    def create_backup(self) -> Path:
        """Create a timestamped backup (delegated)."""
        return self.backup.create_backup()

    def _copy_exclude_protected(self, src: Path, dst: Path) -> None:
        """Copy content while excluding protected paths.

        Args:
            src: Source directory.
            dst: Destination directory.
        """
        dst.mkdir(parents=True, exist_ok=True)

        # PROTECTED_PATHS: only specs/ and reports/ are excluded during copying
        # project/ and config.json are preserved only when they already exist
        template_protected_paths = [
            "specs",
            "reports",
        ]

        for item in src.rglob("*"):
            rel_path = item.relative_to(src)
            rel_path_str = str(rel_path)

            # Skip template copy for specs/ and reports/
            if any(rel_path_str.startswith(p) for p in template_protected_paths):
                continue

            dst_item = dst / rel_path
            if item.is_file():
                # Preserve user content by skipping existing files (v0.3.0)
                # This automatically protects project/ and config.json
                if dst_item.exists():
                    continue
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(item, dst_item)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)

    def _copy_claude(self, silent: bool = False) -> None:
        """.claude/ directory copy (selective with alfred folder overwrite).

        Strategy:
        - Alfred folders (commands/agents/hooks/output-styles/alfred) → copy wholesale (delete & overwrite)
        - Other files/folders → copy individually (preserve existing)
        """
        src = self.template_root / ".claude"
        dst = self.target_path / ".claude"

        if not src.exists():
            if not silent:
                console.print("⚠️ .claude/ template not found")
            return

        # Create .claude directory if not exists
        dst.mkdir(parents=True, exist_ok=True)

        # Alfred folders to copy wholesale (overwrite)
        alfred_folders = [
            "hooks/alfred",
            "commands/alfred",
            "output-styles/alfred",
            "agents/alfred",
        ]

        # 1. Copy Alfred folders wholesale (delete existing & copy new)
        for folder in alfred_folders:
            src_folder = src / folder
            dst_folder = dst / folder

            if src_folder.exists():
                # Delete existing folder and copy new one
                if dst_folder.exists():
                    shutil.rmtree(dst_folder)
                # Create parent directory if needed
                dst_folder.parent.mkdir(parents=True, exist_ok=True)
                shutil.copytree(src_folder, dst_folder)
                if not silent:
                    console.print(f"   ✅ .claude/{folder}/ overwritten")

        # 2. Copy other files/folders individually (preserve existing)
        for item in src.iterdir():
            rel_path = item.relative_to(src)
            dst_item = dst / rel_path

            # Skip Alfred parent folders (already handled above)
            if item.is_dir() and item.name in ["hooks", "commands", "output-styles", "agents"]:
                continue

            if item.is_file():
                # Copy file, skip if exists (preserve user modifications)
                if not dst_item.exists():
                    shutil.copy2(item, dst_item)
            elif item.is_dir():
                # Copy directory recursively (preserve existing files)
                if not dst_item.exists():
                    shutil.copytree(item, dst_item)

        if not silent:
            console.print("   ✅ .claude/ copy complete (alfred folders overwritten, others preserved)")

    def _copy_moai(self, silent: bool = False) -> None:
        """.moai/ directory copy (excludes protected paths)."""
        src = self.template_root / ".moai"
        dst = self.target_path / ".moai"

        if not src.exists():
            if not silent:
                console.print("⚠️ .moai/ template not found")
            return

        # Paths excluded from template copying (specs/, reports/)
        template_protected_paths = [
            "specs",
            "reports",
        ]

        # Copy while skipping protected paths
        for item in src.rglob("*"):
            rel_path = item.relative_to(src)
            rel_path_str = str(rel_path)

            # Skip specs/ and reports/
            if any(rel_path_str.startswith(p) for p in template_protected_paths):
                continue

            dst_item = dst / rel_path
            if item.is_file():
                # Skip existing files to preserve user content (v0.3.0)
                if dst_item.exists():
                    continue
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(item, dst_item)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)

        if not silent:
            console.print("   ✅ .moai/ copy complete (user content preserved)")

    def _copy_claude_md(self, silent: bool = False) -> None:
        """Copy CLAUDE.md with smart merging."""
        src = self.template_root / "CLAUDE.md"
        dst = self.target_path / "CLAUDE.md"

        if not src.exists():
            if not silent:
                console.print("⚠️ CLAUDE.md template not found")
            return

        # Preserve project information when the file exists
        if dst.exists():
            self._merge_claude_md(src, dst)
            if not silent:
                console.print("   🔄 CLAUDE.md merged (project information preserved)")
        else:
            shutil.copy2(src, dst)
            if not silent:
                console.print("   ✅ CLAUDE.md copy complete")

    def _merge_claude_md(self, src: Path, dst: Path) -> None:
        """Delegate the smart merge for CLAUDE.md.

        Args:
            src: Template CLAUDE.md.
            dst: Project CLAUDE.md.
        """
        self.merger.merge_claude_md(src, dst)

    def _copy_gitignore(self, silent: bool = False) -> None:
        """.gitignore copy (optional)."""
        src = self.template_root / ".gitignore"
        dst = self.target_path / ".gitignore"

        if not src.exists():
            return

        # Merge with the existing .gitignore when present
        if dst.exists():
            self._merge_gitignore(src, dst)
            if not silent:
                console.print("   🔄 .gitignore merged")
        else:
            shutil.copy2(src, dst)
            if not silent:
                console.print("   ✅ .gitignore copy complete")

    def _merge_gitignore(self, src: Path, dst: Path) -> None:
        """Delegate the .gitignore merge.

        Args:
            src: Template .gitignore.
            dst: Project .gitignore.
        """
        self.merger.merge_gitignore(src, dst)

    def merge_config(self, detected_language: str | None = None) -> dict[str, str]:
        """Delegate the smart merge for config.json.

        Args:
            detected_language: Detected language.

        Returns:
            Merged configuration dictionary.
        """
        return self.merger.merge_config(detected_language)

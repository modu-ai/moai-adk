# @CODE:TEMPLATE-001 | SPEC: SPEC-INIT-003/spec.md | Chain: TEMPLATE-001
"""Template copy and backup processor (SPEC-INIT-003 v0.3.0: preserve user content)."""

from __future__ import annotations

import re
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
        # config.json is now FORCE OVERWRITTEN (backup in .moai-backups/)
        # Merge via /alfred:0-project when optimized=false
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
        self.context: dict[str, str] = {}  # Template variable substitution context

    def _get_template_root(self) -> Path:
        """Return the template root path."""
        # src/moai_adk/core/template/processor.py → src/moai_adk/templates/
        current_file = Path(__file__).resolve()
        package_root = current_file.parent.parent.parent
        return package_root / "templates"

    def set_context(self, context: dict[str, str]) -> None:
        """Set variable substitution context.

        Args:
            context: Dictionary of template variables.
        """
        self.context = context

    def _substitute_variables(self, content: str) -> tuple[str, list[str]]:
        """Substitute template variables in content.

        Returns:
            Tuple of (substituted_content, warnings_list)
        """
        warnings = []

        # Perform variable substitution
        for key, value in self.context.items():
            placeholder = f"{{{{{key}}}}}"  # {{KEY}}
            if placeholder in content:
                safe_value = self._sanitize_value(value)
                content = content.replace(placeholder, safe_value)

        # Detect unsubstituted variables
        remaining = re.findall(r'\{\{([A-Z_]+)\}\}', content)
        if remaining:
            unique_remaining = sorted(set(remaining))
            warnings.append(f"Unsubstituted variables: {', '.join(unique_remaining)}")

        return content, warnings

    def _sanitize_value(self, value: str) -> str:
        """Sanitize value to prevent recursive substitution and control characters.

        Args:
            value: Value to sanitize.

        Returns:
            Sanitized value.
        """
        # Remove control characters (keep printable and whitespace)
        value = ''.join(c for c in value if c.isprintable() or c in '\n\r\t')
        # Prevent recursive substitution by removing placeholder patterns
        value = value.replace('{{', '').replace('}}', '')
        return value

    def _is_text_file(self, file_path: Path) -> bool:
        """Check if file is text-based (not binary).

        Args:
            file_path: File path to check.

        Returns:
            True if file is text-based.
        """
        text_extensions = {
            '.md', '.json', '.txt', '.py', '.ts', '.js',
            '.yaml', '.yml', '.toml', '.xml', '.sh', '.bash'
        }
        return file_path.suffix.lower() in text_extensions

    def _copy_file_with_substitution(self, src: Path, dst: Path) -> list[str]:
        """Copy file with variable substitution for text files.

        Args:
            src: Source file path.
            dst: Destination file path.

        Returns:
            List of warnings.
        """
        warnings = []

        # Text files: read, substitute, write
        if self._is_text_file(src) and self.context:
            try:
                content = src.read_text(encoding='utf-8')
                content, file_warnings = self._substitute_variables(content)
                dst.write_text(content, encoding='utf-8')
                warnings.extend(file_warnings)
            except UnicodeDecodeError:
                # Binary file fallback
                shutil.copy2(src, dst)
        else:
            # Binary file or no context: simple copy
            shutil.copy2(src, dst)

        return warnings

    def _copy_dir_with_substitution(self, src: Path, dst: Path) -> None:
        """Recursively copy directory with variable substitution for text files.

        Args:
            src: Source directory path.
            dst: Destination directory path.
        """
        dst.mkdir(parents=True, exist_ok=True)

        for item in src.rglob("*"):
            rel_path = item.relative_to(src)
            dst_item = dst / rel_path

            if item.is_file():
                # Create parent directory if needed
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                # Copy with variable substitution
                self._copy_file_with_substitution(item, dst_item)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)

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
        self._copy_github(silent)
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
        """.claude/ directory copy with variable substitution (selective with alfred folder overwrite).

        @CODE:INIT-004:ALFRED-001 | Copy all 4 Alfred command files from templates
        @REQ:COMMAND-GENERATION-001 | SPEC-INIT-004: Automatic generation of Alfred command files
        @SPEC:TEMPLATE-PROCESSING-001 | Template processor integration for Alfred command files

        Strategy:
        - Alfred folders (commands/agents/hooks/output-styles/alfred) → copy wholesale (delete & overwrite)
          * Creates individual backup before deletion for safety
          * Commands: 0-project.md, 1-plan.md, 2-run.md, 3-sync.md
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

        # @CODE:INIT-004:ALFRED-002 | Alfred command files must always be overwritten
        # @CODE:INIT-004:ALFRED-COPY | Copy all 4 Alfred command files from templates
        # Alfred folders to copy wholesale (overwrite)
        alfred_folders = [
            "hooks/alfred",
            "commands/alfred",  # Contains 0-project.md, 1-plan.md, 2-run.md, 3-sync.md
            "output-styles/alfred",
            "agents/alfred",
        ]

        # 1. Copy Alfred folders wholesale (backup before delete & overwrite)
        for folder in alfred_folders:
            src_folder = src / folder
            dst_folder = dst / folder

            if src_folder.exists():
                # Remove existing folder (backup is already handled by create_backup() in update.py)
                if dst_folder.exists():
                    shutil.rmtree(dst_folder)

                # Create parent directory if needed
                dst_folder.parent.mkdir(parents=True, exist_ok=True)
                shutil.copytree(src_folder, dst_folder)
                if not silent:
                    console.print(f"   ✅ .claude/{folder}/ overwritten")

        # 2. Copy other files/folders individually (smart merge for settings.json)
        all_warnings = []
        for item in src.iterdir():
            rel_path = item.relative_to(src)
            dst_item = dst / rel_path

            # Skip Alfred parent folders (already handled above)
            if item.is_dir() and item.name in ["hooks", "commands", "output-styles", "agents"]:
                continue

            if item.is_file():
                # Smart merge for settings.json
                if item.name == "settings.json":
                    self._merge_settings_json(item, dst_item)
                    if not silent:
                        console.print("   🔄 settings.json merged (env variables preserved)")
                else:
                    # FORCE OVERWRITE: Always copy other files (no skip)
                    warnings = self._copy_file_with_substitution(item, dst_item)
                    all_warnings.extend(warnings)
            elif item.is_dir():
                # FORCE OVERWRITE: Always copy directories (no skip)
                self._copy_dir_with_substitution(item, dst_item)

        # Print warnings if any
        if all_warnings and not silent:
            console.print("[yellow]⚠️ Template warnings:[/yellow]")
            for warning in set(all_warnings):  # Deduplicate
                console.print(f"   {warning}")

        if not silent:
            console.print("   ✅ .claude/ copy complete (variables substituted)")

    def _copy_moai(self, silent: bool = False) -> None:
        """.moai/ directory copy with variable substitution (excludes protected paths)."""
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

        all_warnings = []

        # Copy while skipping protected paths
        for item in src.rglob("*"):
            rel_path = item.relative_to(src)
            rel_path_str = str(rel_path)

            # Skip specs/ and reports/
            if any(rel_path_str.startswith(p) for p in template_protected_paths):
                continue

            dst_item = dst / rel_path
            if item.is_file():
                # FORCE OVERWRITE: Always copy files (no skip)
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                # Copy with variable substitution
                warnings = self._copy_file_with_substitution(item, dst_item)
                all_warnings.extend(warnings)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)

        # Print warnings if any
        if all_warnings and not silent:
            console.print("[yellow]⚠️ Template warnings:[/yellow]")
            for warning in set(all_warnings):  # Deduplicate
                console.print(f"   {warning}")

        if not silent:
            console.print("   ✅ .moai/ copy complete (variables substituted)")

    def _copy_github(self, silent: bool = False) -> None:
        """.github/ directory copy with variable substitution."""
        src = self.template_root / ".github"
        dst = self.target_path / ".github"

        if not src.exists():
            if not silent:
                console.print("⚠️ .github/ template not found")
            return

        if dst.exists():
            shutil.rmtree(dst)

        self._copy_dir_with_substitution(src, dst)

        if not silent:
            console.print("   ✅ .github/ copy complete (variables substituted)")

    def _copy_claude_md(self, silent: bool = False) -> None:
        """Copy CLAUDE.md with smart merge (preserves \"## Project Information\" section)."""
        src = self.template_root / "CLAUDE.md"
        dst = self.target_path / "CLAUDE.md"

        if not src.exists():
            if not silent:
                console.print("⚠️ CLAUDE.md template not found")
            return

        # Smart merge: preserve existing "## Project Information" section
        if dst.exists():
            self._merge_claude_md(src, dst)
            if not silent:
                console.print("   🔄 CLAUDE.md merged (project information preserved)")
        else:
            # First time: just copy
            self._copy_file_with_substitution(src, dst)
            if not silent:
                console.print("   ✅ CLAUDE.md created")

    def _merge_claude_md(self, src: Path, dst: Path) -> None:
        """Delegate the smart merge for CLAUDE.md.

        Args:
            src: Template CLAUDE.md.
            dst: Project CLAUDE.md.
        """
        self.merger.merge_claude_md(src, dst)

    def _merge_settings_json(self, src: Path, dst: Path) -> None:
        """Delegate the smart merge for settings.json.

        Args:
            src: Template settings.json.
            dst: Project settings.json.
        """
        # Find the latest backup for user settings extraction
        backup_path = None
        if self.backup.backup_dir.exists():
            backups = sorted(self.backup.backup_dir.iterdir(), reverse=True)
            if backups:
                backup_settings = backups[0] / ".claude" / "settings.json"
                if backup_settings.exists():
                    backup_path = backup_settings

        self.merger.merge_settings_json(src, dst, backup_path)

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

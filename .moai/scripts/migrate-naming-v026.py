#!/usr/bin/env python3
"""
MoAI-ADK Skills Naming Migration Script v0.26.0
Migrates all moai-alfred-* Skills to moai-core-*

BREAKING CHANGE: All alfred Skills renamed to moai-core-* (v0.26.0)
No backward compatibility. Hard break.

Usage:
    python migrate-naming-v026.py --dry-run    # Preview changes
    python migrate-naming-v026.py --execute    # Apply changes
    python migrate-naming-v026.py --rollback   # Revert changes
"""

import os
import re
import sys
import json
import shutil
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Tuple

# ============================================================================
# MIGRATION MAPPING: moai-alfred-* → moai-core-*
# ============================================================================

SKILL_RENAMES = {
    "moai-alfred-workflow": "moai-core-workflow",
    "moai-alfred-personas": "moai-core-personas",
    "moai-alfred-context-budget": "moai-core-context-budget",
    "moai-alfred-agent-factory": "moai-core-agent-factory",
    "moai-alfred-agent-guide": "moai-core-agent-guide",
    "moai-alfred-ask-user-questions": "moai-core-ask-user-questions",
    "moai-alfred-clone-pattern": "moai-core-clone-pattern",
    "moai-alfred-code-reviewer": "moai-core-code-reviewer",
    "moai-alfred-config-schema": "moai-core-config-schema",
    "moai-alfred-dev-guide": "moai-core-dev-guide",
    "moai-alfred-env-security": "moai-core-env-security",
    "moai-alfred-expertise-detection": "moai-core-expertise-detection",
    "moai-alfred-feedback-templates": "moai-core-feedback-templates",
    "moai-alfred-issue-labels": "moai-core-issue-labels",
    "moai-alfred-language-detection": "moai-core-language-detection",
    "moai-alfred-practices": "moai-core-practices",
    "moai-alfred-proactive-suggestions": "moai-core-proactive-suggestions",
    "moai-alfred-rules": "moai-core-rules",
    "moai-alfred-session-state": "moai-core-session-state",
    "moai-alfred-spec-authoring": "moai-core-spec-authoring",
    "moai-alfred-todowrite-pattern": "moai-core-todowrite-pattern",
}

# ============================================================================
# CONFIGURATION
# ============================================================================

PROJECT_ROOT = Path("/Users/goos/MoAI/MoAI-ADK")
PACKAGE_SKILLS_DIR = PROJECT_ROOT / "src/moai_adk/templates/.claude/skills"
LOCAL_SKILLS_DIR = PROJECT_ROOT / ".claude/skills"
AGENTS_DIR_PACKAGE = PROJECT_ROOT / "src/moai_adk/templates/.claude/agents"
AGENTS_DIR_LOCAL = PROJECT_ROOT / ".claude/agents"
COMMANDS_DIR_PACKAGE = PROJECT_ROOT / "src/moai_adk/templates/.claude/commands"
COMMANDS_DIR_LOCAL = PROJECT_ROOT / ".claude/commands"
CLAUDE_MD_PACKAGE = PROJECT_ROOT / "src/moai_adk/templates/CLAUDE.md"
CLAUDE_MD_LOCAL = PROJECT_ROOT / "CLAUDE.md"
MEMORY_DIR = PROJECT_ROOT / ".moai/memory"

MIGRATION_LOG = PROJECT_ROOT / ".moai/logs/migration-v026.log"
MIGRATION_BACKUP = PROJECT_ROOT / ".moai/backup/skills-pre-v026"

# ============================================================================
# LOGGER
# ============================================================================

class Logger:
    def __init__(self, log_file: Path):
        self.log_file = log_file
        self.log_file.parent.mkdir(parents=True, exist_ok=True)
        self.messages = []

    def log(self, message: str, level: str = "INFO"):
        timestamp = datetime.now().isoformat()
        formatted = f"[{timestamp}] {level}: {message}"
        self.messages.append(formatted)
        print(formatted)

    def save(self):
        with open(self.log_file, 'w') as f:
            f.write('\n'.join(self.messages))

# ============================================================================
# MIGRATION OPERATIONS
# ============================================================================

class SkillsMigration:
    def __init__(self, dry_run: bool = True):
        self.dry_run = dry_run
        self.logger = Logger(MIGRATION_LOG)
        self.changes = []
        self.errors = []

    def log_change(self, file_path: str, change_type: str, description: str):
        """Log a migration change"""
        self.changes.append({
            "file": file_path,
            "type": change_type,
            "description": description
        })
        self.logger.log(f"{change_type}: {file_path} - {description}")

    def log_error(self, file_path: str, error: str):
        """Log an error"""
        self.errors.append({"file": file_path, "error": error})
        self.logger.log(f"ERROR: {file_path} - {error}", level="ERROR")

    def rename_skill_directory(self, old_name: str, new_name: str):
        """Rename Skill directory in package and local"""
        for base_dir in [PACKAGE_SKILLS_DIR, LOCAL_SKILLS_DIR]:
            if not base_dir.exists():
                continue

            old_path = base_dir / old_name
            new_path = base_dir / new_name

            if not old_path.exists():
                continue

            if self.dry_run:
                self.log_change(str(old_path), "RENAME", f"→ {new_path.name}")
            else:
                try:
                    old_path.rename(new_path)
                    self.log_change(str(old_path), "RENAME", f"✅ → {new_path.name}")
                except Exception as e:
                    self.log_error(str(old_path), str(e))

    def update_skill_metadata(self, new_name: str):
        """Update SKILL.md metadata (name field)"""
        for base_dir in [PACKAGE_SKILLS_DIR, LOCAL_SKILLS_DIR]:
            if not base_dir.exists():
                continue

            skill_md = base_dir / new_name / "SKILL.md"
            if not skill_md.exists():
                continue

            try:
                content = skill_md.read_text(encoding='utf-8')

                # Update name field in YAML frontmatter
                updated_content = re.sub(
                    r'^name:\s*"[^"]*moai-alfred-[^"]*"',
                    f'name: "{new_name}"',
                    content,
                    flags=re.MULTILINE
                )

                if updated_content != content:
                    if self.dry_run:
                        self.log_change(str(skill_md), "UPDATE", "Update name field in YAML")
                    else:
                        skill_md.write_text(updated_content, encoding='utf-8')
                        self.log_change(str(skill_md), "UPDATE", "✅ Updated name field")
            except Exception as e:
                self.log_error(str(skill_md), str(e))

    def update_file_references(self, old_name: str, new_name: str):
        """Update Skill() references in all files"""
        pattern = f'Skill\\("{old_name}"\\)'
        replacement = f'Skill("{new_name}")'

        files_to_update = []

        # Agents
        if AGENTS_DIR_PACKAGE.exists():
            files_to_update.extend(AGENTS_DIR_PACKAGE.rglob("*.md"))
        if AGENTS_DIR_LOCAL.exists():
            files_to_update.extend(AGENTS_DIR_LOCAL.rglob("*.md"))

        # Commands
        if COMMANDS_DIR_PACKAGE.exists():
            files_to_update.extend(COMMANDS_DIR_PACKAGE.rglob("*.md"))
        if COMMANDS_DIR_LOCAL.exists():
            files_to_update.extend(COMMANDS_DIR_LOCAL.rglob("*.md"))

        # CLAUDE.md
        if CLAUDE_MD_PACKAGE.exists():
            files_to_update.append(CLAUDE_MD_PACKAGE)
        if CLAUDE_MD_LOCAL.exists():
            files_to_update.append(CLAUDE_MD_LOCAL)

        # Memory files
        if MEMORY_DIR.exists():
            files_to_update.extend(MEMORY_DIR.rglob("*.md"))

        # Other Skills' SKILL.md files
        for base_dir in [PACKAGE_SKILLS_DIR, LOCAL_SKILLS_DIR]:
            if base_dir.exists():
                files_to_update.extend(base_dir.rglob("SKILL.md"))

        for file_path in files_to_update:
            if not file_path.is_file():
                continue

            try:
                content = file_path.read_text(encoding='utf-8')

                # Use regex for exact match
                updated_content = re.sub(
                    pattern,
                    replacement,
                    content
                )

                if updated_content != content:
                    if self.dry_run:
                        self.log_change(str(file_path), "SKILL_REF", f"Replace {old_name} → {new_name}")
                    else:
                        file_path.write_text(updated_content, encoding='utf-8')
                        self.log_change(str(file_path), "SKILL_REF", f"✅ Replace {old_name} → {new_name}")
            except Exception as e:
                self.log_error(str(file_path), str(e))

    def backup_skills(self):
        """Create backup of Skills directory"""
        if self.dry_run:
            self.log_change(str(MIGRATION_BACKUP), "BACKUP", "Would create backup")
        else:
            try:
                if MIGRATION_BACKUP.exists():
                    shutil.rmtree(MIGRATION_BACKUP)

                for base_dir in [PACKAGE_SKILLS_DIR, LOCAL_SKILLS_DIR]:
                    if not base_dir.exists():
                        continue

                    backup_target = MIGRATION_BACKUP / base_dir.name
                    shutil.copytree(base_dir, backup_target, dirs_exist_ok=True)

                self.log_change(str(MIGRATION_BACKUP), "BACKUP", "✅ Backup created")
            except Exception as e:
                self.log_error(str(MIGRATION_BACKUP), str(e))

    def run_migration(self):
        """Execute full migration"""
        self.logger.log(f"{'='*80}")
        self.logger.log(f"MoAI-ADK Skills Migration v0.26.0")
        self.logger.log(f"{'='*80}")
        self.logger.log(f"Mode: {'DRY RUN (preview only)' if self.dry_run else 'EXECUTE (apply changes)'}")
        self.logger.log(f"Total Skills to migrate: {len(SKILL_RENAMES)}")
        self.logger.log(f"{'='*80}\n")

        # Step 1: Backup
        self.logger.log("Step 1/5: Creating backup...")
        self.backup_skills()

        # Step 2: Rename directories
        self.logger.log("\nStep 2/5: Renaming Skill directories...")
        for old_name, new_name in SKILL_RENAMES.items():
            self.rename_skill_directory(old_name, new_name)

        # Step 3: Update SKILL.md metadata
        self.logger.log("\nStep 3/5: Updating SKILL.md metadata...")
        for new_name in SKILL_RENAMES.values():
            self.update_skill_metadata(new_name)

        # Step 4: Update file references
        self.logger.log("\nStep 4/5: Updating Skill() references in files...")
        for old_name, new_name in SKILL_RENAMES.items():
            self.update_file_references(old_name, new_name)

        # Step 5: Report
        self.logger.log("\n" + "="*80)
        self.logger.log("Migration Summary")
        self.logger.log("="*80)
        self.logger.log(f"Total changes: {len(self.changes)}")
        self.logger.log(f"Total errors: {len(self.errors)}")

        if self.errors:
            self.logger.log("\nErrors encountered:")
            for error in self.errors:
                self.logger.log(f"  - {error['file']}: {error['error']}")

        self.logger.log(f"\nLog file: {MIGRATION_LOG}")
        self.logger.save()

        return len(self.errors) == 0

# ============================================================================
# CLI
# ============================================================================

def main():
    if len(sys.argv) < 2:
        print("Usage:")
        print("  python migrate-naming-v026.py --dry-run    # Preview changes")
        print("  python migrate-naming-v026.py --execute    # Apply changes")
        print("  python migrate-naming-v026.py --rollback   # Revert changes")
        sys.exit(1)

    command = sys.argv[1]

    if command == "--dry-run":
        print("\n" + "="*80)
        print("DRY RUN MODE: Previewing all changes")
        print("="*80 + "\n")

        migration = SkillsMigration(dry_run=True)
        migration.run_migration()

        print(f"\n{'='*80}")
        print(f"Preview complete. Review changes above.")
        print(f"To apply: python {sys.argv[0]} --execute")
        print("="*80 + "\n")

    elif command == "--execute":
        print("\n" + "="*80)
        print("⚠️  EXECUTE MODE: Applying migration")
        print("="*80)
        confirm = input("This will rename 21 Skills. Continue? (yes/no): ").strip().lower()

        if confirm != "yes":
            print("Migration cancelled.")
            sys.exit(0)

        migration = SkillsMigration(dry_run=False)
        success = migration.run_migration()

        if success:
            print("\n" + "="*80)
            print("✅ Migration completed successfully!")
            print("="*80)
            print("\nNext steps:")
            print("1. Review .moai/logs/migration-v026.log")
            print("2. Run: git add -A")
            print("3. Run: git commit -m \"feat(skills)!: Rename alfred Skills to moai-core-*\"")
            print("="*80 + "\n")
        else:
            print("\n" + "="*80)
            print("❌ Migration completed with errors!")
            print("="*80)
            print(f"Check .moai/logs/migration-v026.log for details")
            print("="*80 + "\n")
            sys.exit(1)

    elif command == "--rollback":
        print("\n" + "="*80)
        print("⚠️  ROLLBACK MODE: Reverting migration")
        print("="*80)
        confirm = input("This will restore Skills from backup. Continue? (yes/no): ").strip().lower()

        if confirm != "yes":
            print("Rollback cancelled.")
            sys.exit(0)

        if not MIGRATION_BACKUP.exists():
            print("❌ No backup found. Cannot rollback.")
            sys.exit(1)

        try:
            for base_dir in [PACKAGE_SKILLS_DIR, LOCAL_SKILLS_DIR]:
                if base_dir.exists():
                    shutil.rmtree(base_dir)
                    backup_source = MIGRATION_BACKUP / base_dir.name
                    if backup_source.exists():
                        shutil.copytree(backup_source, base_dir)

            print("✅ Rollback completed successfully!")
            print("Restore all updated files from git:")
            print("  git checkout -- src/moai_adk/templates/")
            print("  git checkout -- .claude/")
            print("="*80 + "\n")
        except Exception as e:
            print(f"❌ Rollback failed: {e}")
            sys.exit(1)

    else:
        print(f"Unknown command: {command}")
        sys.exit(1)

if __name__ == "__main__":
    main()

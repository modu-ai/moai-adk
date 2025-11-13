"""
Migration CLI command for MoAI-ADK

Provides command-line interface for project version migrations.
"""

import logging
import sys
from pathlib import Path

import click

from moai_adk.core.migration import VersionMigrator

logger = logging.getLogger(__name__)


@click.command()
@click.option(
    "--check",
    is_flag=True,
    help="Check if migration is needed without executing",
)
@click.option(
    "--dry-run",
    is_flag=True,
    help="Show migration plan without making changes",
)
@click.option(
    "--no-cleanup",
    is_flag=True,
    help="Keep old files after migration",
)
@click.option(
    "--rollback",
    is_flag=True,
    help="Rollback to the latest backup",
)
@click.option(
    "--list-backups",
    is_flag=True,
    help="List available backups",
)
def migrate(check, dry_run, no_cleanup, rollback, list_backups):
    """
    Migrate MoAI-ADK project to the latest version

    This command automatically migrates your project structure to match
    the latest MoAI-ADK version. It creates backups before migration
    and can rollback if issues occur.

    Examples:

        # Check if migration is needed
        moai-adk migrate --check

        # Preview migration plan
        moai-adk migrate --dry-run

        # Execute migration
        moai-adk migrate

        # Migrate without cleaning up old files
        moai-adk migrate --no-cleanup

        # Rollback to latest backup
        moai-adk migrate --rollback

        # List available backups
        moai-adk migrate --list-backups
    """
    try:
        project_root = Path.cwd()
        migrator = VersionMigrator(project_root)

        # List backups
        if list_backups:
            backups = migrator.backup_manager.list_backups()
            if not backups:
                click.echo("📦 No backups found")
                return

            click.echo("📦 Available backups:\n")
            for i, backup in enumerate(backups, 1):
                click.echo(f"{i}. {backup['description']} ({backup['timestamp']})")
                click.echo(f"   Path: {backup['path']}")
                click.echo(f"   Files: {backup['files']}")
                click.echo()
            return

        # Rollback
        if rollback:
            click.echo("🔙 Rolling back to latest backup...")
            if migrator.rollback_to_latest_backup():
                click.echo("✅ Rollback completed successfully")
                sys.exit(0)
            else:
                click.echo("❌ Rollback failed", err=True)
                sys.exit(1)

        # Check status
        if check:
            status = migrator.check_status()
            version_info = status["version"]
            migration_info = status["migration"]

            click.echo("📊 Migration Status:\n")
            click.echo(f"Current Version: {version_info['detected_version']}")
            click.echo(
                f"Needs Migration: {'Yes' if version_info['needs_migration'] else 'No'}"
            )

            if version_info["needs_migration"]:
                click.echo(f"Target Version: {migration_info['target_version']}")
                click.echo(f"Files to migrate: {migration_info['file_count']}")
                click.echo("\n💡 Run 'moai-adk migrate' to execute migration")
            else:
                click.echo("\n✅ Project is up to date")

            return

        # Dry run
        if dry_run:
            if migrator.needs_migration():
                migrator.migrate_to_v024(dry_run=True)
            else:
                click.echo("✅ Project is already up to date")
            return

        # Execute migration
        if not migrator.needs_migration():
            click.echo("✅ Project is already up to date")
            return

        click.echo("🚀 MoAI-ADK Project Migration\n")
        click.echo("This will migrate your project to v0.24.0")
        click.echo("A backup will be created automatically\n")

        # Confirm with user
        if not click.confirm("Do you want to proceed?"):
            click.echo("❌ Migration cancelled")
            return

        # Execute migration
        success = migrator.migrate_to_v024(dry_run=False, cleanup=not no_cleanup)

        if success:
            click.echo("\n🎉 Migration completed successfully!")
            sys.exit(0)
        else:
            click.echo("\n❌ Migration failed", err=True)
            click.echo("💡 Use 'moai-adk migrate --rollback' to restore from backup")
            sys.exit(1)

    except Exception as e:
        logger.error(f"Migration command failed: {e}", exc_info=True)
        click.echo(f"❌ Error: {e}", err=True)
        sys.exit(1)

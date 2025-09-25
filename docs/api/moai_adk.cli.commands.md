# moai_adk.cli.commands

CLI Commands for MoAI-ADK

Contains all Click command definitions for the MoAI-ADK CLI including
init, restore, doctor, status, update, and help commands.

## Functions

### cli

@FEATURE:CLI-001 Modu-AI's Agentic Development Kit

```python
cli(ctx, version, help_flag)
```

### restore

@TASK:RESTORE-001 Restore MoAI-ADK from a backup directory.

```python
restore(backup_path, dry_run)
```

### doctor

@TASK:HEALTH-001 Diagnose common issues and check system health.

```python
doctor(list_backups)
```

### init

@TASK:INIT-001 Initialize a new MoAI-ADK project.

```python
init(project_path, template, interactive, backup, force, force_copy, quiet, personal, team)
```

### help

@TASK:HELP-001 Show help for MoAI-ADK commands.

```python
help(command)
```

### status

@TASK:STATUS-001 Show MoAI-ADK project status.

```python
status(verbose, project_path)
```

### update

@TASK:UPDATE-001 Update MoAI-ADK to the latest version.

```python
update(check, no_backup, verbose, package_only, resources_only)
```

### create_mode_configuration

@TASK:CONFIG-001 Create mode-specific configuration for MoAI-ADK project.

```python
create_mode_configuration(project_dir, project_mode, quiet)
```

### progress_callback

```python
progress_callback(message, current, total)
```

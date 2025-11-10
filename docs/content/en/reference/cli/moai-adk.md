# moai-adk Complete Command Reference

Covers all commands and options for the `moai-adk` command-line tool.

## Command Structure

```
moai-adk <command> [options] [arguments]
```

## Global Options

Available for all `moai-adk` commands:

| Option               | Description        |
| -------------------- | ------------------ |
| `--help`, `-h`       | Show help          |
| `--version`, `-v`    | Show version info  |
| `--verbose`, `-vv`   | Verbose output mode |
| `--no-color`         | Output without color |

## Command List

### 1. moai-adk init

**Project initialization and template injection**

#### Syntax

```bash
moai-adk init [path] [options]
```

#### Arguments

| Argument | Description          | Default        |
| -------- | -------------------- | -------------- |
| `path`   | Project path         | Current directory |

#### Options

```bash
--language LANG, -l LANG    Select project language (ko/en/ja/zh)
--mode MODE                 Development mode (solo/team/org)
--with-mcp SERVER          Add MCP server (context7, figma, playwright)
--mcp-auto                 Auto-install all recommended MCP servers
--force, -f                Overwrite existing settings
--skip-git                 Skip Git initialization
```

#### Examples

```bash
# Create new project
moai-adk init my-project

# Initialize current directory
moai-adk init .

# Include MCP servers
moai-adk init . --with-mcp context7 --with-mcp figma

# Auto-install all MCPs
moai-adk init . --mcp-auto

# Reinitialize with overwrite
moai-adk init . --force
```

#### Generated Files

```
project/
├── .moai/
│   ├── config.json        # Project settings
│   ├── specs/            # SPEC documents
│   ├── docs/             # Generated documentation
│   ├── reports/          # Analysis reports
│   └── scripts/          # Utility scripts
├── .claude/
│   ├── settings.json     # Claude Code settings
│   ├── commands/         # Alfred commands
│   ├── agents/          # Sub-agent templates
│   └── mcp.json         # MCP settings
└── CLAUDE.md            # Project guidelines
```

______________________________________________________________________

### 2. moai-adk doctor

**System Environment Diagnosis**

#### Syntax

```bash
moai-adk doctor [options]
```

#### Options

```bash
--verbose, -vv       Detailed diagnosis information
--fix                Attempt automatic fixes
--export FILE       Export results to file
```

#### Diagnosis Items

- ✅ Python version (3.13+ required)
- ✅ uv package manager
- ✅ Git repository status
- ✅ `.moai/` directory structure
- ✅ `.claude/` resources
- ✅ Claude Code accessibility
- ✅ Python dependencies
- ✅ Disk space

#### Examples

```bash
# Basic diagnosis
moai-adk doctor

# Detailed diagnosis
moai-adk doctor -vv

# Save results to file
moai-adk doctor --export .moai/reports/doctor.txt

# Automatic fixes
moai-adk doctor --fix
```

______________________________________________________________________

### 3. moai-adk status

**Project Status Query**

#### Syntax

```bash
moai-adk status [options]
```

#### Options

```bash
--json                 JSON format output
--compact, -c          Show simple summary only
--spec ID              Detailed query for specific SPEC
```

#### Displayed Information

- 📋 SPEC progress status (complete/in_progress/pending)
- 🏷️ TAG statistics (@SPEC/@TEST/@CODE/@DOC)
- 📝 Recent commits
- 📅 Last sync time
- 🔄 Git branch status

#### Examples

```bash
# Full status
moai-adk status

# JSON format
moai-adk status --json

# Simple summary
moai-adk status --compact

# Specific SPEC details
moai-adk status --spec SPEC-001
```

______________________________________________________________________

### 4. moai-adk backup

**Project Backup Creation**

#### Syntax

```bash
moai-adk backup [options]
```

#### Options

```bash
--target DIR           Backup location (default: .moai-backups/)
--include-git          Include Git history
--compress, -z         Compressed format (tar.gz)
--restore FILE         Restore backup
```

#### Backup Targets

- `.moai/` entire directory
- `.claude/` resources
- `CLAUDE.md` project guidelines
- `pyproject.toml` / `requirements.txt`

#### Examples

```bash
# Basic backup
moai-adk backup

# Compressed backup
moai-adk backup --compress

# Include Git history
moai-adk backup --include-git

# Restore backup
moai-adk backup --restore .moai-backups/20250115_143000/

# Custom location
moai-adk backup --target ~/backups/moai/
```

______________________________________________________________________

### 5. moai-adk update

**Package and Template Synchronization (Most Important Command)**

#### Syntax

```bash
moai-adk update [options]
```

#### Options

```bash
--check                Check update availability only
--dry-run              Preview changes
--skip-backup          Skip backup (not recommended)
--force                Force update
--from VERSION         Update from specific version
```

#### Process

1. **Version Check**: Check latest version from PyPI
2. **Backup Creation**: Safely backup current state
3. **Template Sync**: Merge new templates with existing settings
4. **Validation**: Verify integrity
5. **Completion**: Summarize changes

#### Examples

```bash
# Regular update
moai-adk update

# Preview
moai-adk update --dry-run

# Force update
moai-adk update --force

# Version check only
moai-adk update --check
```

______________________________________________________________________

## Option Combinations

### Advanced Use Cases

```bash
# Initialize with verbose logging
moai-adk init . --verbose --with-mcp context7

# Automatic fixes with detailed report
moai-adk doctor --fix --export .moai/reports/doctor.md

# Export full status as JSON
moai-adk status --json > status.json

# Create compressed backup and test restore
moai-adk backup --compress
moai-adk backup --restore .moai-backups/latest.tar.gz
```

______________________________________________________________________

## Exit Codes

| Code  | Meaning       |
| ----- | ------------- |
| `0`   | Success       |
| `1`   | General error |
| `2`   | Usage error   |
| `127` | Command not found |

______________________________________________________________________

## Environment Variables

```bash
MOAI_HOME              MoAI-ADK installation path
MOAI_DEBUG             Enable debug mode (1)
MOAI_NO_COLOR          Disable color output (1)
MOAI_CONFIG_PATH       .moai/config.json path
```

______________________________________________________________________

## Troubleshooting

### "Permission denied"

```bash
# Check permissions
ls -la .moai/
chmod -R u+w .moai/
```

### "Template conflict"

```bash
# Backup then force update
moai-adk backup
moai-adk update --force
```

### "Python version mismatch"

```bash
# Check Python 3.13+
python3 --version
uv python install 3.13
```

______________________________________________________________________

**Next**: [Alfred Subcommand Guide](subcommands.md) or [CLI Reference](index.md)




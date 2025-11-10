# CLI Command Reference

The `moai-adk` CLI is a Click-based command-line interface responsible for project management and template synchronization. It is used separately from Alfred commands (alfred:*) for local environment setup and maintenance.

## Core Commands

| Command                 | Description                                              | When to Use                          |
| ----------------------- | -------------------------------------------------------- | ------------------------------------ |
| `moai-adk init [path]` | Create new project or inject templates into existing project | When first introducing Alfred |
| `moai-adk doctor`      | Environment check (Python, uv, Git, directory structure) | After installation, when issues occur |
| `moai-adk status`      | TAG summary, checkpoints, template version query        | Before work, before review           |
| `moai-adk backup`      | Create backup of `.moai/`, `.claude/`, CLAUDE.md         | Before template update, before major changes |
| `moai-adk update`      | Package & template synchronization (most important command) | After new version release, regular check |

## Command Details

### `moai-adk init`

**Purpose**: Initialize new project and create basic structure

**Usage**:

```bash
# Create new project
moai-adk init my-project

# Initialize current directory
moai-adk init .

# Inject MoAI-ADK into existing project
moai-adk init .
```

**Generated Structure**:

```
my-project/
├── .moai/        # Project metadata
├── .claude/      # Alfred resources
└── CLAUDE.md     # Project guidelines
```

**Initialization Process**:

1. Python environment check
2. Git repository initialization (if not exists)
3. Create `.moai/` directory structure
4. Copy `.claude/` resource templates
5. Create default configuration files

### `moai-adk doctor`

**Purpose**: System environment diagnosis and troubleshooting

**Usage**:

```bash
moai-adk doctor
```

**Diagnosis Items**:

- ✅ Python version (3.13+)
- ✅ uv package manager
- ✅ Git repository status
- ✅ `.moai/` directory structure
- ✅ `.claude/` resource integrity
- ✅ Claude Code accessibility

**Expected Output**:

```
🩺 MoAI-ADK System Check
✅ Python 3.13.0
✅ uv 0.5.1
✅ Git repository initialized
✅ .moai/ directory structure normal
✅ .claude/ resources 74 loaded
✅ Claude Code accessible

System is normal. Ready to start Alfred!
```

### `moai-adk status`

**Purpose**: Project status summary and state understanding

**Usage**:

```bash
moai-adk status
```

**Displayed Information**:

- SPEC progress status (complete/in_progress/pending)
- TAG statistics (@SPEC/@TEST/@CODE/@DOC)
- Recent checkpoints
- Template version information
- Git workflow status

**Expected Output**:

```
📊 MoAI-ADK Project Status
:bullseye: Project: MyProject
📅 Last sync: 2025-01-15 14:30

📋 SPEC Progress
- ✅ Completed: 12
- 🔄 In Progress: 3
- ⏳ Pending: 5

🏷️ TAG Statistics
- @SPEC: 20 tags
- @TEST: 18 tags
- @CODE: 17 tags
- @DOC: 16 tags
- 🚨 Orphan tags: 2

📝 Version Info
- Template: v0.15.2
- Last update: 2025-01-10
- Backup available: .moai-backups/20250110/

🔄 Git Status
- Current branch: feature/auth-system
- Ahead of main: 12 commits
- Draft PR: #23
```

### `moai-adk backup`

**Purpose**: Create project resource backup

**Usage**:

```bash
moai-adk backup
```

**Backup Targets**:

- `.moai/` entire directory
- `.claude/` resource templates
- `CLAUDE.md` project guidelines
- Git status information

**Backup Location**:

```
.moai-backups/
└── 20250115_143000/
    ├── .moai/
    ├── .claude/
    ├── CLAUDE.md
    └── backup-info.json
```

### `moai-adk update`

**Purpose**: Package and template synchronization (most important command)

**Usage**:

```bash
moai-adk update
```

**Update Stages**:

1. **Stage 1**: Package version check
2. **Stage 2**: Template version comparison
3. **Stage 3**: Backup creation and merge

**Automatic Processing**:

- Check latest version from PyPI
- Backup current resources to `.moai-backups/`
- Merge new templates with existing settings
- Guidance message on conflicts

**Output Example**:

```
🔄 MoAI-ADK Update Started
:package: Current version: v0.15.1
:package: Latest version: v0.15.2

📁 Creating backup...
✅ Backup created: .moai-backups/20250115_143000/

🔄 Updating templates...
🔧 Merging .moai/config.json
🔧 Updating Alfred agents
🔧 Syncing Skills (74 → 77)

✅ Update completed successfully!
📝 Changelog: Added moai-domain-ml Skill
⚠️  Please review .claude/settings.json changes
```

## Internal Operation

### CLI Architecture

```
moai-adk
├── __main__.py           # Click entry point
├── cli/
│   ├── commands/
│   │   ├── init.py      # Project initialization
│   │   ├── doctor.py    # Environment diagnosis
│   │   ├── status.py    # Status query
│   │   ├── backup.py    # Backup creation
│   │   └── update.py    # Template synchronization
│   └── utils.py          # Common utilities
├── core/
│   ├── template.py      # Template management
│   ├── backup.py        # Backup/restore
│   └── filesystem.py    # File system operations
└── templates/           # Default template source
```

### Rich Console Output

- **Color coding**: Success (green), warning (yellow), error (red)
- **Progress bars**: Show progress for long-running tasks
- **Table format**: Display status information organized
- **ASCII art**: Logo and separators

### Error Handling

- **Clear messages**: User-friendly error descriptions
- **Solution suggestions**: Specific methods for problem resolution
- **Error codes**: Exit codes for automation scripts
- **Logging**: Detailed logs for problem tracking

## Best Practices

### Regular Maintenance

```bash
# Monthly regular check
moai-adk doctor
moai-adk status
moai-adk backup
moai-adk update
```

### Before Major Changes

```bash
# Safe change procedure
moai-adk backup  # 1. Create backup
# Perform changes...
moai-adk status  # 2. Check status
moai-adk doctor  # 3. Environment check
```

### New Team Member Onboarding

```bash
# Standard onboarding procedure
git clone <project>
cd <project>
moai-adk doctor  # Environment check
moai-adk status  # Understand project
claude           # Start Alfred
/alfred:0-project  # Initialize project
```

## Related Links

- **[Project Structure](project-structure)** - Detailed `.moai/` and `.claude/` directories
- **[Alfred Commands](../alfred/commands)** - alfred:* workflow commands
- **[Workflow](../workflow)** - How CLI and Alfred integrate

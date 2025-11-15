# Configuration Optimization States

## Overview

The MoAI-ADK project initialization system uses a two-state model to ensure idempotency and safe configuration merging. The `optimized` flag in `.moai/config/config.json` tracks the configuration state and guides users through the setup process.

## State Definitions

### optimized = false

**Status**: Configuration merge required (after template update)

**When**:
- After running `moai-adk init --force` on an existing project
- When template files are updated (version bump)
- Before running `/alfred:0-project` to complete the setup

**What It Means**:
- Template version has been updated (you get new features)
- Your previous settings are preserved (backed up in `.moai-backups/`)
- Configuration merge is required to finalize the setup
- Next action: Run `/alfred:0-project` in Claude Code

**User Guidance**:
```
⚠️ Configuration Status: optimized=false (merge required)

What Happened:
✅ Template files updated to latest version
💾 Your previous settings backed up in: .moai-backups/backup/
⏳ Configuration merge required

What Happens Next:
1. Run /alfred:0-project in Claude Code
2. System intelligently merges old settings + new template
3. After successful merge → optimized becomes true
4. You're ready to continue developing
```

### optimized = true

**Status**: Configuration fully merged and verified (ready for development)

**When**:
- After `/alfred:0-project` successfully merges configurations
- During fresh project initialization (first time)
- After config merge completes without errors

**What It Means**:
- Your configuration is complete and ready
- Old and new settings have been intelligently merged
- No action required from you
- You can start developing immediately

**User Guidance**:
```
✅ Configuration Status: optimized=true (ready)

Your configuration is complete and ready for development.
All previous settings have been preserved and merged.
```

## State Transition Diagram

```
Initial Project Creation
        ↓
  optimized=false (initial value)
        ↓
  User runs: moai-adk init
        ↓
  Project initialized with backup
        ↓
  optimized=false (merge required)
        ↓
  User runs: /alfred:0-project
        ↓
  Configuration merge completes
        ↓
  optimized=true (ready)
  optimized_at = ISO timestamp
        ↓
  [Development Phase]
        ↓
  User runs: moai-adk init --force (update template)
        ↓
  optimized=false (merge required again)
  optimized_at = null
        ↓
  Cycle repeats...
```

## Timestamp Field (optimized_at)

### Purpose
The `optimized_at` field records when the configuration was last successfully merged.

### Values

**When optimized=false**:
```json
"optimized_at": null
```
No timestamp because configuration hasn't been merged yet.

**When optimized=true**:
```json
"optimized_at": "2025-11-16T10:30:45Z"
```
ISO 8601 timestamp in UTC indicating when the merge completed.

### Format
- ISO 8601 format with UTC timezone
- Example: `"2025-11-16T10:30:45Z"`
- Timezone always represented as `Z` (UTC)

## Idempotency Guarantee

The system is designed to be idempotent (safe to run multiple times):

### First Run (optimized=false)

```
Initial State: optimized=false
    ↓
/alfred:0-project runs
    ↓
Configuration merge happens
    ↓
Final State: optimized=true, optimized_at="2025-11-16T10:30:45Z"
```

### Second Run (optimized=true)

```
Initial State: optimized=true
    ↓
/alfred:0-project runs
    ↓
SKIPS merge (already optimized)
    ↓
Final State: Unchanged (optimized=true, optimized_at="2025-11-16T10:30:45Z")
```

Running `/alfred:0-project` multiple times with `optimized=true` is safe:
- No merge happens (already done)
- No data is overwritten
- Configuration remains stable
- User is informed that setup is complete

## User Workflows

### Fresh Project Setup

```bash
# 1. Create new project
$ moai-adk init

# Output shows:
# ✅ Initialization Completed Successfully
# ⚠️ Configuration Status: optimized=false (merge required)

# 2. Open in Claude Code and run
$ /alfred:0-project

# System merges configuration
# optimized becomes true
# ✅ Configuration ready for development
```

### Template Update (Maintenance)

```bash
# 1. Update MoAI-ADK package
$ uv add moai-adk@0.26.0

# 2. Reinitialize to get new template files
$ moai-adk init --force

# Output shows:
# ✅ Initialization Completed Successfully
# ⚠️ Configuration Status: optimized=false (merge required)
# 💾 Previous settings backed up

# 3. Merge new template with old settings
$ /alfred:0-project

# System merges:
# - New template features
# - Your old settings
# - Your customizations (preserved)
# optimized becomes true again
```

## Session Hook Display

During each Claude Code session start, the configuration status is displayed:

### When optimized=true
```
🚀 MoAI-ADK Session Started
📦 Version: 0.25.7 (up-to-date)
⚙️  Config Status: ✅ ready
🌿 Branch: develop
🔄 Changes: 0
🎯 SPEC Progress: 3/10 (30%)
🔨 Last Commit: feat: Add optimization status (2 hours ago)
```

### When optimized=false
```
🚀 MoAI-ADK Session Started
📦 Version: 0.25.7 (up-to-date)
⚙️  Config Status: ⚠️ merge required
🌿 Branch: develop
🔄 Changes: 0
🎯 SPEC Progress: 3/10 (30%)
🔨 Last Commit: feat: Add optimization status (2 hours ago)
```

## Configuration Structure

### Project Section

```json
{
  "project": {
    "name": "My Project",
    "description": "Project description",
    "owner": "Your Name",
    "mode": "personal",
    "locale": "en",
    "language": "Python",
    "created_at": "2025-11-16T00:00:00Z",
    "initialized": true,
    "optimized": false,
    "optimized_at": null,
    "_notes": "optimized=false: Template updated, config merge required. optimized=true: Configuration fully merged and verified."
  }
}
```

### Field Descriptions

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `optimized` | boolean | false | Config state: true=ready, false=merge required |
| `optimized_at` | string/null | null | ISO timestamp when config was last merged |
| `_notes` | string | - | Documentation about optimization states |

## Implementation Details

### ConfigManager Methods

**set_optimized(config_path, value)**
- Sets `optimized` field only
- Used for simple state changes
- Maintains `optimized_at` as-is

**set_optimized_with_timestamp(config_path, value)**
- Sets both `optimized` and `optimized_at`
- Called after config merge completes
- Auto-generates ISO timestamp when value=True
- Clears timestamp when value=False

### Session Hook Integration

The SessionStart hook reads the `optimized` field and displays status:
- File: `.claude/hooks/alfred/session_start__show_project_info.py`
- Updates: `format_session_output()` function
- Displays: Config status with indicator (✅ or ⚠️)

### Init Command Integration

The init command handles reinitialization:
- File: `src/moai_adk/cli/commands/init.py`
- When reinit detected: Sets `optimized=false`
- Shows clear guidance: Explains what happened and next steps
- Backs up previous config

## Testing

All configuration state transitions are covered by tests in:
- File: `tests/test_project_init_idempotency.py`
- Coverage:
  - State transitions (false→true, true→false)
  - Timestamp management (set, clear, ISO format)
  - Idempotency (safe to run multiple times)
  - User edits preservation
  - Session hook display
  - Clear guidance messages

## Troubleshooting

### Issue: optimized=false but I already ran /alfred:0-project

**Solution**: Check if the merge completed successfully. If there were errors:
1. Backup your custom settings manually
2. Run `/alfred:0-project` again
3. System will reattempt the merge

### Issue: optimized field is missing from config

**Solution**: This shouldn't happen with modern versions. If it does:
1. Close Claude Code
2. Manually add to `.moai/config/config.json`:
   ```json
   {
     "project": {
       "optimized": false,
       "optimized_at": null
     }
   }
   ```
3. Restart Claude Code
4. Run `/alfred:0-project`

### Issue: Keep seeing "merge required" even after running /alfred:0-project

**Possible Causes**:
1. `/alfred:0-project` command timed out before completing
2. Errors occurred during merge but weren't displayed
3. Config file wasn't saved properly

**Solution**:
1. Check Claude Code console for error messages
2. Verify `.moai/config/config.json` exists and is readable
3. Try running `/alfred:0-project` again
4. If issues persist, check `.moai/logs/` for error details

## Related Documentation

- SPEC-PROJECT-INIT-IDEMPOTENT-001: Implementation specification
- tests/test_project_init_idempotency.py: Test coverage
- src/moai_adk/core/template/config.py: ConfigManager implementation
- .claude/hooks/alfred/session_start__show_project_info.py: Session hook
- src/moai_adk/cli/commands/init.py: Init command

---

**Last Updated**: 2025-11-16
**Version**: 1.0
**Status**: Complete

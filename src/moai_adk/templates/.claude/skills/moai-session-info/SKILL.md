---
name: moai-session-info
version: 4.0.0
created: 2025-11-05
updated: '2025-11-18'
status: stable
description: Display comprehensive project and session information including Git status,
  SPEC progress, version details, and system resources. Use when starting new sessions,
  checking project status, reviewing project context, or when users ask "what's the
  status", "show project info", or "where are we".
keywords:
- session-info
- project-status
- git-status
- spec-progress
- version-check
- system-resources
allowed-tools:
- Read
- Glob
- Bash
stability: stable
---


# Session Information Provider

## Skill Metadata

| Field | Value |
| ----- | ----- |
| Version | 1.0.0 |
| Tier | Alfred (Session Management) |
| Auto-load | On session start or when status requested |
| Purpose | Provide comprehensive project and session overview |

---

## What It Does

Comprehensive session and project information provider that gives users complete context about their current MoAI-ADK project state, including Git status, SPEC progress, version information, and system resources.

**Core capabilities**:
- ✅ Project metadata and configuration display
- ✅ Git repository status and commit history
- ✅ SPEC progress tracking and completion metrics
- ✅ Version information and update availability
- ✅ System resource monitoring
- ✅ Checkpoint status and restoration options
- ✅ Session metrics and handoff information

---

## When to Use

- ✅ When starting a new Claude Code session
- ✅ When checking project status and progress
- ✅ Before making significant changes or commits
- ✅ When users ask "what's the status", "show project info", "where are we"
- ✅ When reviewing project context and history
- ✅ Before running /alfred commands

---

## Core Information Categories

### 1. Project Overview
```bash
🗿 Project: MoAI-ADK
📁 Location: /Users/goos/MoAI/MoAI-ADK
🌍 Language: 한국어 (Korean)
🔧 Mode: Team (GitFlow)
⚡ Toolchain: Python optimized
```

### 2. Version Information
```bash
📦 Current: v0.15.2
🆓 Update Available: v0.16.0
⬆️  Upgrade Command: pip install --upgrade moai-adk
📝 Release Notes: https://github.com/moai-adk/moai-adk/releases/tag/v0.16.0
```

### 3. Git Repository Status
```bash
🌿 Branch: develop (3 commits ahead of main)
📝 Changes: 5 modified, 2 added
🔨 Last Commit: feat: Complete skill consolidation (2 hours ago)
📊 Commit Hash: a1b2c3d
```

### 4. SPEC Progress
```bash
📋 Total SPECs: 15
✅ Completed: 12 (80%)
⏳ In Progress: 2
📝 Pending: 1
📊 Completion Rate: 80%
```

### 5. System Resources
```bash
🧠 Memory Usage: 2.4GB / 16GB (15%)
💾 Disk Space: 45GB free
🔄 CPU Usage: 12%
⚡ Session Duration: 45 minutes
```

### 6. Available Checkpoints
```bash
🗂️  Checkpoints: 3 available
   📌 auth-system-implementation (30 min ago)
   📌 skill-consolidation (2 hours ago)
   📌 feature-branch-workflow (yesterday)
↩️  Restore: /alfred:0-project restore
```

---

## Quick Start Commands

### Basic Status Check
```python
# Simple project overview
Skill("moai-session-info")
```

### Detailed Status with Metrics
```python
# Comprehensive status with all details
Skill("moai-session-info")
# Response includes all categories above
```

### Before Major Operations
```python
# Always check status before:
# - /alfred:1-plan (planning new features)
# - /alfred:2-run (implementing changes)
# - git operations (commits, merges)

Skill("moai-session-info")
# Review status, then proceed with operation
```

---

## Information Sources

The skill gathers information from multiple sources:

### Project Configuration
- `.moai/config.json` - Project settings and language
- `pyproject.toml` - Package version and dependencies
- `.git/` - Repository status and history

### SPEC Tracking
- `.moai/specs/` - SPEC documents and completion status
- SPEC metadata - Progress tracking and milestones

### System Resources
- `psutil` - Memory and CPU usage
- File system - Disk space and project size
- Session metrics - Current session duration

### Version Information
- Package registries - Latest available versions
- GitHub releases - Release notes and changelogs

---

## Status Message Format

The skill generates structured status messages with consistent formatting:

```
🚀 MoAI-ADK Project Status

📋 Project Overview
   🗿 Project: {project_name}
   📁 Location: {project_path}
   🌍 Language: {language}
   🔧 Mode: {git_mode}

📦 Version Information
   📦 Current: {current_version}
   {update_information}
   📝 Release Notes: {release_url}

🌿 Git Repository
   🌿 Branch: {branch} ({commit_hash})
   📝 Changes: {file_changes}
   🔨 Last: {last_commit_message}

📊 SPEC Progress
   📋 Total: {total_specs}
   ✅ Completed: {completed_specs} ({percentage}%)
   ⏳ In Progress: {in_progress_specs}

🧠 System Resources
   🧠 Memory: {memory_usage}
   💾 Disk: {disk_space}
   ⚡ Session: {session_duration}

🗂️  Checkpoints
   {checkpoint_list}
   ↩️  Restore: /alfred:0-project restore
```

---

## Integration with Alfred Commands

This skill is automatically invoked by:

### SessionStart Hook Integration
```python
# In session_start__show_project_info.py
# Automatically called when session starts
Skill("moai-session-info")
```

### Command Integration
```python
# Before /alfred:1-plan
if context == "planning":
    Skill("moai-session-info")  # Show current status

# Before /alfred:2-run
if context == "implementation":
    Skill("moai-session-info")  # Confirm project state

# Before git operations
if "git" in command:
    Skill("moai-session-info")  # Show repository status
```

---

## Error Handling and Fallbacks

### Graceful Degradation
The skill provides useful information even when some sources fail:

```python
# If Git commands fail:
# Still show project info, version, and system resources

# If SPEC counting fails:
# Still show Git status and version information

# If network access fails:
# Still show local information (Git, SPECs, system)
```

### Common Error Scenarios
- **Git repository not found**: Shows project info without Git details
- **No .moai/config.json**: Uses default settings and basic project detection
- **Network unavailable**: Shows local information only
- **Permission denied**: Provides read-only information where possible

---

## Performance Considerations

### Optimization Strategies
- **Caching**: Cache expensive operations (Git history, version checks)
- **Timeouts**: 5-second timeout for network operations
- **Lazy Loading**: Load detailed information only when requested
- **Incremental Updates**: Update only changed information

### Resource Usage
- **Memory**: Minimal footprint (< 10MB)
- **Network**: Only for version checks (cached locally)
- **Disk**: Reads existing files, no modifications
- **CPU**: Lightweight operations, quick response times

---

## Usage Examples

### Example 1: Session Start
```python
# User starts new Claude Code session
Skill("moai-session-info")

# Output:
🚀 MoAI-ADK Session Started

📋 Project Overview
   🗿 Project: MoAI-ADK
   📁 Location: /Users/goos/MoAI/MoAI-ADK
   🌍 Language: 한국어
   🔧 Mode: Team

📦 Version: v0.15.2 → v0.16.0 available
📝 Release Notes: https://github.com/...

🌿 Branch: develop (3 ahead)
📝 Changes: 5 modified, 2 added
📋 SPEC Progress: 12/15 (80%)
```

### Example 2: Pre-Implementation Check
```python
# User wants to implement new feature
"/alfred:2-run SPEC-AUTH-001"

# Alfred automatically calls:
Skill("moai-session-info")

# User sees status before implementation begins
```

### Example 3: Status Query
```python
# User asks: "what's our current status?"
Skill("moai-session-info")

# Complete project status displayed
```

---

**End of Skill** | Optimized for quick status checks and session context
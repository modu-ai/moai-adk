# MoAI-ADK Integration Module

Purpose: Detailed integration patterns for moai-worktree with MoAI-ADK Plan-Run-Sync workflow including plan phase automation, DDD integration, and cleanup workflows.

Version: 1.0.0
Last Updated: 2026-01-06

---

## Quick Reference (30 seconds)

MoAI-ADK Integration Points:
- /moai:1-plan: Automatic worktree creation after SPEC generation
- /moai:2-run: DDD execution in isolated worktree environment
- /moai:3-sync: Worktree sync with documentation updates
- Cleanup: Automatic removal of merged worktrees

---

## Plan Phase Integration (/moai:1-plan)

### Automatic Worktree Creation

After SPEC creation, the worktree system automatically generates an isolated development environment.

Branch Naming Convention:
- Pattern: feature/SPEC-{id}-{title-kebab-case}
- Example: SPEC-001 with title "User Auth" becomes feature/SPEC-001-user-auth

Worktree Path Pattern:
- Default: {repo}/.moai/worktrees/{project-name}/SPEC-{id}
- Configurable via worktree_root setting

Creation Workflow:
1. SPEC is created with /moai:1-plan
2. Worktree new command is invoked automatically if auto_create is enabled
3. Branch is created from configured base branch (default: main)
4. Template is applied if specified
5. Worktree is registered in the central registry
6. User receives guidance for switching to the new worktree

### Template-Based Setup

Templates provide pre-configured environments for different development scenarios.

Available Template Types:
- spec-development: Default SPEC development environment
- backend: Python/Node.js backend setup with testing frameworks
- frontend: React/Vue frontend setup with build tools
- full-stack: Combined backend and frontend configuration

Template Configuration Structure:
- setup_commands: List of commands to run after worktree creation
- files: Configuration files to create in the worktree
- env_vars: Environment variables to set in .env.local

---

## Development Phase Integration (/moai:2-run)

### Worktree-Aware DDD

The DDD manager detects worktree environments and adapts its behavior accordingly.

Worktree Detection:
- Checks if current directory name starts with SPEC-
- Looks for .moai/worktrees directory in path hierarchy
- Validates against registry for accurate identification

DDD Execution Benefits:
- Independent development results per worktree
- Isolated dependency environments
- No cross-contamination between SPECs
- Automatic metadata updates in registry

Registry Updates During Development:
- last_accessed timestamp updated on each worktree access
- last_ddd_result stored for progress tracking
- operation_status recorded for debugging

### Development Server Isolation

Each worktree can run independent development servers without port conflicts.

Port Assignment Strategy:
- Base port calculated from SPEC ID hash
- Frontend server: base_port + 0
- Backend server: base_port + 1
- Database: base_port + 2

Server Management:
- PID files stored in worktree root for process tracking
- Automatic cleanup on worktree removal
- Status command shows running servers per worktree

---

## Sync Phase Integration (/moai:3-sync)

### Automated Worktree Synchronization

Before PR creation or documentation sync, worktrees should be synchronized with their base branch.

Sync Workflow:
1. Check for uncommitted changes (abort if found without force flag)
2. Fetch latest changes from remote
3. Analyze sync needs (commits ahead/behind)
4. Execute sync using configured strategy (merge, rebase, or squash)
5. Update registry with sync timestamp
6. Continue with documentation sync

Conflict Resolution Options:
- auto-resolve: Automatically resolve simple conflicts using configured strategy
- interactive: Prompt for manual resolution of each conflict
- abort: Cancel sync and preserve current state

Include/Exclude Patterns:
- Use --include to sync only specific directories like src/ or docs/
- Use --exclude to skip directories like node_modules/ or build/

### Documentation Generation

After worktree sync, documentation updates can be extracted:
- API documentation from changed endpoints
- Test coverage reports from test results
- Architecture updates from structural changes
- CHANGELOG entries from commit messages

---

## Post-PR Cleanup Workflow

### Automated Cleanup

After successful PR merge, worktrees can be automatically cleaned up.

Cleanup Triggers:
- Manual cleanup with clean command
- Automated cleanup when cleanup_merged is enabled
- Scheduled cleanup for stale worktrees

Cleanup Options:
1. Remove worktree and branch (default for merged)
2. Remove worktree, keep branch for reference
3. Archive worktree to backup location
4. Skip cleanup and keep for future reference

Registry Maintenance:
- Completed SPECs recorded with merged_at timestamp
- Cleanup action documented for audit trail
- Statistics updated (total_worktrees, merged_worktrees)

---

## Team Collaboration Patterns

### Shared Worktree Registry

For team environments, configure a shared registry accessible to all developers.

Team Registry Configuration:
- registry_type: Set to "team" for shared mode
- shared_registry_path: Network-accessible registry location
- developer_prefix: Automatic prefix for developer-specific worktrees

Synchronization:
- Local registry syncs with team registry periodically
- Merge conflicts resolved by timestamp priority
- Developer can force local or remote on conflict

### Collaborative Development

Multiple developers can coordinate on related SPECs:

Coordination Pattern:
1. Lead developer creates base worktree with shared contracts
2. Team members create dependent worktrees with developer flags
3. Shared contracts directory maintains API agreements
4. Integration worktree combines work from all team members

Access Levels:
- read-only: View worktree status and metadata
- read-write: Full development access
- admin: Can modify team registry settings

---

## Configuration Reference

### MoAI Configuration Integration

Worktree settings in .moai/config/config.yaml:

worktree section:
- auto_create: Enable automatic worktree creation (default: true)
- auto_sync: Enable automatic synchronization (default: true)
- cleanup_merged: Remove worktrees for merged branches (default: true)
- worktree_root: Base directory for worktrees (default: {repo}/.moai/worktrees)
- default_base: Default base branch (default: main)
- sync_strategy: Sync method - merge, rebase, or squash (default: merge)
- registry_type: local or team (default: local)

Template Settings:
- template_dir: Custom template location
- default_template: Template applied when none specified

---

## Error Handling

### Common Integration Errors

Worktree Already Exists:
- Error: Worktree path already exists for SPEC ID
- Resolution: Use --force to recreate or choose different SPEC ID

Uncommitted Changes:
- Error: Worktree has uncommitted changes during sync
- Resolution: Commit changes first or use --force flag

Merge Conflicts:
- Error: Conflicts detected during sync operation
- Resolution: Use --interactive for manual resolution or --auto-resolve

Registry Corruption:
- Error: Registry file is invalid or inaccessible
- Resolution: Run repair command or restore from backup

### Recovery Patterns

For failed worktree creation:
- Partial worktree is automatically cleaned up
- Registry entry is removed if created
- Error details logged for debugging

For failed synchronization:
- Worktree reset to last known good state if backup ref exists
- Status set to error in registry
- Manual intervention flag set for user attention

---

Version: 1.0.0
Last Updated: 2026-01-06
Module: MoAI-ADK workflow integration patterns for Plan-Run-Sync phases

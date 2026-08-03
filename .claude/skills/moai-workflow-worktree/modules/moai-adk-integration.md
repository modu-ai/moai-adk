# MoAI-ADK Integration Module

Purpose: Detailed integration patterns for moai-worktree with MoAI-ADK Plan-Run-Sync workflow including plan phase automation, DDD integration, and cleanup workflows.

Version: 1.0.0

---

## Quick Reference (30 seconds)

MoAI-ADK Integration Points (EnterWorktree-first doctrine):
- Worktree entry: User runs `moai cc -w <name>` to enter a worktree; plan does NOT auto-create one. The legacy `/moai plan --worktree` launch action is retired.
- /moai plan: Runs in the main checkout (no worktree) on a `plan/SPEC-XXX` branch.
- /moai run: If the user entered a worktree, DDD execution happens there; otherwise it runs on a feature branch in the main checkout.
- /moai sync: Reuses the SAME worktree the user entered for run (do NOT recreate).
- Cleanup: Disposal via `moai worktree done SPEC-XXX` AFTER both run PR AND sync PR merge.

---

## Plan Phase Integration (/moai plan)

### Worktree Is NOT Auto-Created by Plan

The plan phase does NOT create a worktree. A worktree is entered by the USER before a phase runs, not provisioned by a workflow step. The EnterWorktree-first doctrine is the SSOT:

- New-session launch (post-`/clear` or new terminal): `moai cc -w <name>` (or `moai glm -w` / `moai cg -w`). The `-w` flag accepts both short names (resolved under `.claude/worktrees/`) and absolute paths under `~/.moai/worktrees/<project>/...`.
- Current-session re-entry (same session continuing): the runtime tool `EnterWorktree(<path>)`.
- The retired `/moai plan --worktree` flag and the retired `moai worktree new` command MUST NOT be presented as live entry points.

Branch Naming Convention (informational — the launcher handles this):
- Pattern: feature/SPEC-{id}-{title-kebab-case}
- Example: SPEC-001 with title "User Auth" becomes feature/SPEC-001-user-auth

Worktree Path Pattern:
- L1 Claude-native (ephemeral): `{repo}/.claude/worktrees/<auto-name>/`
- L2 MoAI persistent (SPEC-scoped): `~/.moai/worktrees/{project-name}/SPEC-{id}/`

Reference: the canonical L1/L2 layering and the EnterWorktree-first policy live in `.claude/rules/moai/workflow/worktree-integration.md` § Terminology Glossary.

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

## Development Phase Integration (/moai run)

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

## Sync Phase Integration (/moai sync)

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
- Manual cleanup with the `moai worktree done SPEC-XXX` command (the canonical disposal path — runs only after BOTH run PR AND sync PR merge)
- Automated cleanup when `workflow.worktree.auto_cleanup` is opted in (default: off)
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

### Per-Developer Worktree Conventions

There is no shared-registry config field on `WorkflowWorktreeConfig` (the legacy `registry_type: "team"` / `shared_registry_path` / `developer_prefix` entries are NOT real keys). Team coordination is convention-based: each developer enters their own L2 worktree (`~/.moai/worktrees/{project}/SPEC-{id}/`) and the team coordinates via the SPEC artifacts and PRs in version control, not via a runtime-shared registry.

Coordination conventions:
- Each developer uses their own worktree under their own home directory.
- The SPEC artifact (`.moai/specs/SPEC-XXX/`) is the shared state — it lives in the repo, not a per-worktree registry.
- PRs are the integration surface — the team merges feature branches rather than maintaining a network-accessible worktree registry.

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

Worktree settings live in `.moai/config/sections/workflow.yaml` under the `workflow.worktree` section. The keys below are the REAL Go struct fields (`WorkflowWorktreeConfig` in `internal/config/types.go`, defaults in `internal/config/defaults.go`). All three automation toggles ship `false` by default — worktree automation is explicit user opt-in per the EnterWorktree-first policy.

workflow.worktree section (real keys + real defaults):
- `auto_cleanup` (bool, default: **false**) — automatically remove worktrees. Opt-in; off by default to avoid unintended sprawl.
- `auto_create` (bool, default: **false**) — automatically create worktrees. Opt-in; the default flow runs phases on a feature branch in the main checkout.
- `auto_merge` (bool, default: **false**) — automatically merge worktree branches. Opt-in; off by default.
- `session_name_pattern` (string, default: `moai-{ProjectName}-{SPEC-ID}`) — naming pattern for sessions spawned inside a worktree.
- `tmux_preferred` (bool, default: **true**) — prefer tmux for worktree session display.

Notes:
- The previously-documented fields `auto_sync`, `cleanup_merged`, `worktree_root`, `default_base`, `sync_strategy`, and `registry_type` are NOT fields on the Go struct and MUST NOT be presented as live config keys. (The legacy content that listed them with `true` defaults inverted the real defaults and is corrected here.)
- Worktree automation defaults to OFF in both the distributed template and the local dev config. Changing a default requires a dedicated SPEC.

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
Module: MoAI-ADK workflow integration patterns for Plan-Run-Sync phases

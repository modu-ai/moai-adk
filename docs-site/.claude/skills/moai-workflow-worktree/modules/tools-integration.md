# Tools Integration Module

Purpose: Integration patterns for moai-worktree with development tools, IDEs, terminals, CI/CD pipelines, and monitoring systems.

Version: 1.0.0
Last Updated: 2026-01-06

---

## Quick Reference (30 seconds)

Tool Integration Points:
- IDE: VS Code multi-root workspaces, worktree-specific settings
- Terminal: Shell prompt, completion, navigation aliases
- Git Hooks: Post-checkout, pre-push validation
- CI/CD: GitHub Actions, parallel testing, SPEC-aware builds
- Monitoring: Usage analytics, performance tracking

---

## IDE Integration

### VS Code Multi-Root Workspace

Generate a VS Code workspace that includes all active worktrees as folders.

Workspace Structure:
- Main Repository as primary folder
- Each active worktree as additional folder
- Shared extensions recommendations
- Worktree-specific launch configurations

Dynamic Workspace Generation:
- Detects all active worktrees from registry
- Creates folder entries for each worktree
- Generates tasks for common worktree operations
- Updates automatically when worktrees change

Worktree-Specific Settings:
- Python interpreter path for virtual environments
- Linting and formatting configurations
- File exclusion patterns for build artifacts
- Test runner configurations

Task Generation Per Worktree:
- Run Tests: Execute /moai:2-run for the SPEC
- Sync Worktree: Run moai-worktree sync command
- Switch to Worktree: Change active directory

### JetBrains IDE Integration

For IntelliJ, PyCharm, and WebStorm:
- Configure separate project modules for each worktree
- Share run configurations across worktrees
- Set worktree-specific interpreter paths

---

## Terminal Integration

### Shell Profile Enhancement

Add to .bashrc or .zshrc for improved worktree experience.

Completion Support:
- Tab completion for worktree IDs using registry data
- Command option completion for all moai-worktree subcommands

Prompt Customization:
- Detect if current directory is within a worktree
- Display SPEC ID in prompt when in worktree context
- Color coding for worktree status

Navigation Aliases:
- mw: Short alias for moai-worktree
- mwl: List worktrees
- mws: Switch to worktree
- mwg: Navigate with eval pattern
- mwsync: Sync current worktree
- mwclean: Clean merged worktrees

Quick Functions:
- mwnew: Create and switch to new worktree in one command
- mwdev: Switch to worktree and start development with /moai:2-run
- mwpush: Sync worktree and push to remote branch

---

## Git Hooks Integration

### Post-Checkout Hook

Triggered when switching branches or worktrees.

Hook Actions:
1. Detect if running in worktree environment
2. Update last access time in registry
3. Check if sync is needed with base branch
4. Display sync reminder if worktree is behind
5. Load worktree-specific environment file if present

### Pre-Push Hook

Validate worktree state before pushing to remote.

Hook Actions:
1. Detect if pushing from a worktree
2. Check for uncommitted changes
3. Verify sync status with base branch
4. Prompt for confirmation if behind base branch
5. Update registry with push timestamp

---

## CI/CD Pipeline Integration

### GitHub Actions Workflow

SPEC-aware CI/CD workflow for worktree-based development.

Workflow Triggers:
- Push to feature/SPEC-* branches
- Pull requests targeting main or develop

Job Configuration:

Detect SPEC Job:
- Extract SPEC ID from branch name
- Determine worktree type from branch suffix
- Set outputs for downstream jobs

Test Worktree Job:
- Create simulated worktree environment
- Install dependencies based on worktree type
- Run worktree-specific test suites
- Upload test results as artifacts

Sync Worktrees Job:
- Runs after all tests pass
- Triggers worktree sync in development environment
- Prepares for PR creation

Matrix Strategy:
- Parallel testing for different worktree types
- Authentication, payment, dashboard tested independently
- Results aggregated for final status

### Artifact Management

Test Results:
- Store in worktree-specific artifact paths
- Include coverage reports
- Preserve for historical analysis

Build Artifacts:
- Per-worktree build directories
- Independent deployment artifacts
- Version tracking per SPEC

---

## Monitoring and Analytics

### Usage Metrics Collection

Track worktree usage patterns for optimization.

Collected Metrics:
- total_worktrees: Current count of all worktrees
- active_worktrees: Currently active worktrees
- disk_usage: Storage consumed by each worktree
- sync_frequency: How often worktrees are synchronized
- developer_activity: Per-developer usage patterns
- performance_metrics: Operation timing data

### Analytics Reporting

Generate comprehensive usage reports.

Report Contents:
- Worktrees created in reporting period
- Average sync frequency per worktree
- Most active worktrees by developer
- Conflict rate during sync operations
- Storage growth trends

Recommendations:
- Identify worktrees needing cleanup
- Suggest sync frequency improvements
- Highlight potential disk space issues

### Export Formats

Metrics can be exported in multiple formats:
- JSON: For programmatic consumption
- CSV: For spreadsheet analysis
- Prometheus: For monitoring system integration

### Monitoring System Integration

Push metrics to external monitoring:
- Prometheus Pushgateway support
- Custom webhook integration
- Scheduled metric updates

---

## Resource Management

### Parallel Operation Optimization

Manage resources when operating on multiple worktrees.

Resource Awareness:
- CPU count detection for parallel operations
- Memory availability monitoring
- Maximum concurrent worktrees limit

Operation Grouping:
- Prioritize worktrees by activity level
- Group operations for efficient execution
- Sequential fallback for resource constraints

Execution Strategy:
- ThreadPoolExecutor for parallel sync
- Configurable worker count
- Result aggregation with error handling

### Disk Space Management

Monitor and optimize disk usage across worktrees.

Usage Analysis:
- Calculate size per worktree
- Identify large worktrees exceeding thresholds
- Generate optimization suggestions

Optimization Actions:
- Remove build artifacts and caches
- Git garbage collection per worktree
- Compress old/inactive worktrees
- Archive and remove stale worktrees

Cleanup Patterns:
- Log file removal (*.log)
- Python cache cleanup (__pycache__, .pytest_cache)
- Node modules pruning
- Build directory cleanup

---

## Error Handling

### Tool Integration Errors

IDE Configuration Errors:
- Invalid workspace file format
- Missing worktree paths
- Resolution: Regenerate workspace configuration

Hook Execution Errors:
- Permission denied on hook scripts
- Missing dependencies for hook execution
- Resolution: Check file permissions and dependencies

CI/CD Pipeline Errors:
- Branch detection failures
- Artifact upload failures
- Resolution: Verify branch naming conventions and artifact paths

Monitoring Integration Errors:
- Pushgateway connection failures
- Metric format errors
- Resolution: Verify endpoint configuration and data format

---

Version: 1.0.0
Last Updated: 2026-01-06
Module: Development tools and external system integration patterns

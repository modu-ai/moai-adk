# Parallel Workflows Module

Purpose: Development patterns and workflows for parallel SPEC development including independent features, sequential dependencies, and multi-developer coordination.

Version: 1.0.0
Last Updated: 2026-01-06

---

## Quick Reference (30 seconds)

Workflow Patterns:
- Independent SPEC: Multiple unrelated features in parallel
- Sequential Overlap: Dependent features with preparation overlap
- Experiment-Production: Experimental work alongside stable production
- Multi-Developer: Team coordination with shared registry

---

## Independent SPEC Development

### Pattern Overview

Multiple unrelated features developed simultaneously by a single developer or team.

Use Case:
- Features have no code dependencies
- Different areas of codebase
- Parallel development maximizes throughput

### Workflow Steps

Phase 1 - Setup Parallel Environments:
- Create worktrees for each SPEC at planning time
- Use descriptive names matching SPEC content
- Apply appropriate templates for each type

Phase 2 - Initialize Each Environment:
- Switch to each worktree and install dependencies
- Configure development environment
- Start development servers in background

Phase 3 - Parallel Development:
- Work on each SPEC in separate terminal sessions
- Run /moai:2-run for DDD implementation
- Switch between worktrees as needed

Phase 4 - Integration:
- Sync all worktrees before integration
- Run /moai:3-sync for documentation updates
- Create PRs from each worktree branch

### Best Practices

Naming Convention:
- Use SPEC ID with descriptive suffix
- Examples: SPEC-AUTH-001, SPEC-PAY-001, SPEC-DASH-001

Environment Isolation:
- Separate dependency installations per worktree
- Independent development server ports
- Worktree-specific environment files

Sync Frequency:
- Sync with main at least daily
- Sync before starting new feature phase
- Sync before creating PR

---

## Sequential Feature Development

### Pattern Overview

Features with dependencies developed in sequence with preparation overlap for efficiency.

Use Case:
- Later features depend on earlier ones
- APIs built before consumers
- Foundation before higher-level features

### Workflow Steps

Step 1 - Foundation Work:
- Create and develop foundation SPEC
- Core APIs and data structures
- Run full test suite before proceeding

Step 2 - Prepare Dependent Worktree:
- Create dependent worktree while foundation develops
- Set up structure and placeholders
- Mock integration points temporarily

Step 3 - Foundation Completion:
- Complete foundation development
- Merge to development branch
- Update shared branch

Step 4 - Dependent Development:
- Sync dependent worktree with updated base
- Replace mocks with real implementations
- Complete dependent feature

Step 5 - Continue Chain:
- Repeat for next level of dependencies
- Each level syncs after previous completes

### Dependency Management

Integration Contracts:
- Define API contracts early
- Create contract files in shared location
- Validate contract compatibility during sync

Sync Timing:
- Sync after dependency merges to base
- Validate dependent code against real implementation
- Adjust for any contract changes

---

## Experiment-Production Parallel

### Pattern Overview

Experimental features developed alongside stable production work without risk.

Use Case:
- Production bug fixes in stable environment
- Experimental features in isolation
- Risk-free innovation with production stability

### Workflow Steps

Setup Production Worktree:
- Create from main branch
- Focus on stable, proven changes
- Immediate deployment capability

Setup Experimental Worktree:
- Create from develop branch
- Freedom for innovative approaches
- May include breaking changes

Context Switching:
- Use aliases for quick switching
- Maintain clear mental separation
- Production work has priority

Comparison and Review:
- Compare approaches between worktrees
- Review experimental work before integration
- Decide on adoption or discard

### Best Practices

Clear Naming:
- SPEC-PROD-XXX for production work
- SPEC-EXP-XXX for experimental work

Backup Strategy:
- Regular commits in experimental worktree
- Push to remote experimental branches
- Archive before major changes

Integration Decisions:
- Successful experiments promoted to production track
- Failed experiments discarded with learnings documented
- Gradual integration through feature flags

---

## Multi-Developer Coordination

### Team Registry Configuration

Configure shared registry for team visibility.

Registry Settings:
- registry_type: Set to team mode
- shared_registry_path: Network-accessible location
- developer_prefix: Automatic prefix for worktrees

Developer Visibility:
- All developers see team worktree list
- Developer-specific worktrees clearly identified
- Team overview shows all active work

### Coordination Patterns

Integration Point Definition:
- Define shared integration points early
- Create contract files for APIs
- Validate contracts during development

Shared Contracts:
- Create contracts directory in each worktree
- Define API endpoints and data formats
- Check contract compatibility during sync

Integration Worktree:
- Create dedicated integration worktree
- Pull contracts from all team worktrees
- Test integration before merging

### Collaboration Workflow

Team Setup:
1. Lead creates base worktree with shared configuration
2. Team members create dependent worktrees with developer flags
3. Contracts directory maintained in each worktree
4. Integration worktree combines all work

Development Flow:
1. Each developer works in their worktree
2. Contracts updated as interfaces evolve
3. Regular sync with base branch
4. Integration testing in dedicated worktree

Merge Strategy:
1. Integration worktree validates compatibility
2. Individual PRs created from developer worktrees
3. Ordered merge based on dependencies
4. Final integration verification

---

## CI/CD Pipeline Integration

### Parallel Testing

Test worktrees in parallel in CI/CD pipeline.

Workflow Configuration:
- Trigger on feature/SPEC-* branch pushes
- Detect SPEC ID and worktree type from branch
- Run appropriate test suite

Test Matrix:
- Unit tests per worktree type
- Integration tests with mock dependencies
- E2E tests in integration environment

Artifact Management:
- Store test results per worktree
- Coverage reports per SPEC
- Build artifacts with SPEC identification

### Deployment Patterns

Preview Deployments:
- Each worktree branch gets preview environment
- Independent URLs for each SPEC
- Easy comparison between features

Staged Deployment:
- Merge to develop triggers staging deployment
- Integration testing in staging
- Promotion to production after validation

---

## Resource Management

### Development Server Coordination

Manage multiple development servers across worktrees.

Port Assignment:
- Calculate base port from SPEC ID
- Offset for different server types
- Avoid conflicts between worktrees

Server Lifecycle:
- Start servers when entering worktree
- Store PID files for tracking
- Clean shutdown when leaving worktree

Status Monitoring:
- Check running servers across worktrees
- Report port usage and status
- Alert on port conflicts

### Background Operations

Optimize operations across worktrees.

Background Sync:
- Sync worktrees without blocking
- Queue operations for sequential execution
- Report completion asynchronously

Memory-Efficient Switching:
- Pre-load worktree metadata in background
- Immediate switch without waiting
- Complete loading while user works

---

Version: 1.0.0
Last Updated: 2026-01-06
Module: Parallel development workflow patterns and coordination strategies

# Advanced Parallel Development Module

Purpose: Advanced use cases for parallel development including feature flags, microservices, performance optimization, and development tool integration.

Version: 1.0.0
Last Updated: 2026-01-06

---

## Quick Reference (30 seconds)

Advanced Patterns:
- Feature Flags: Test multiple flag configurations in parallel
- Microservices: Coordinate development across service boundaries
- Performance: Optimize resources for parallel operations
- Integration: IDE and shell enhancements for multi-worktree

---

## Feature Flag Development

### Parallel Flag Implementation

Develop feature flags across multiple worktrees for comprehensive testing.

Worktree Strategy:
- Backend worktree: Flag implementation and storage
- Frontend worktree: Flag consumption and UI changes
- Integration worktree: Testing flag combinations

### Backend Flag Implementation

Create feature flags in backend worktree:

Flag Definition:
- Environment variable based flags
- Configuration file based flags
- Database stored flags

Flag Access Pattern:
- Read from environment with fallback
- Default to disabled (false) for safety
- Log flag state for debugging

### Frontend Flag Consumption

Consume flags in frontend worktree:

Flag Provider:
- Load flags from environment
- Provide context to components
- Support hot reload during development

Component Usage:
- Conditional rendering based on flags
- Feature gating for new functionality
- Graceful degradation when disabled

### Integration Testing

Test flag combinations in integration worktree:

Combination Matrix:
- Test all flag combinations systematically
- Document expected behavior per combination
- Automate flag configuration for tests

Test Execution:
- Set flag environment variables
- Run integration test suite
- Record results per configuration

---

## Microservices Parallel Development

### Service-Per-Worktree Pattern

Each microservice developed in dedicated worktree.

Worktree Creation:
- Create worktree per service
- Apply microservice template
- Configure service-specific settings

Service Configuration:
- Unique port per service
- Service name in environment
- Database URL per service

### Service Discovery

Configure service discovery across worktrees.

Gateway Integration:
- Create gateway worktree for integration
- Service configuration with URLs and health checks
- Proxy routing to individual services

Discovery Pattern:
- Local configuration file
- Service URLs based on worktree ports
- Health check endpoints per service

### Parallel Service Testing

Test services independently and together.

Independent Tests:
- Unit tests in each service worktree
- Integration tests with mocked dependencies
- Contract tests for API boundaries

Integrated Tests:
- Gateway worktree runs full stack
- End-to-end tests across services
- Performance tests with realistic load

---

## Performance Optimization

### Resource Monitoring

Monitor resource usage across parallel worktrees.

Metrics Collected:
- Process count per worktree
- Disk usage per worktree
- Memory usage per worktree

System Resources:
- Overall CPU load
- Total memory usage
- Available disk space

### Worktree Optimization

Optimize individual worktrees for better performance.

Cleanup Actions:
- Remove log files
- Clean Python and Node caches
- Run Git garbage collection
- Compress large files

Optimization Triggers:
- Scheduled cleanup intervals
- Before running tests
- Before creating PR

### Parallel Operation Limits

Configure limits for parallel operations.

Resource-Based Limits:
- Maximum concurrent syncs based on CPU count
- Memory threshold for parallel operations
- Disk I/O limits for large operations

Graceful Degradation:
- Queue operations when limits reached
- Sequential fallback under pressure
- Priority queue for critical operations

---

## IDE Integration

### VS Code Multi-Root Workspace

Configure VS Code for multi-worktree development.

Workspace Structure:
- Main repository as first folder
- Each worktree as additional folder
- Shared settings across workspace

Folder Configuration:
- Display name matching SPEC ID
- Relative path to worktree
- Folder-specific settings override

Settings Per Worktree:
- Python interpreter path
- Node modules location
- Build output directory

### Worktree-Specific Configuration

Configure IDE settings per worktree.

Python Configuration:
- Virtual environment path
- Linting settings
- Formatter configuration
- Test runner settings

TypeScript Configuration:
- Node modules path
- Build configuration
- Debug configuration
- Launch settings

### Task Automation

Automate common operations in IDE.

Task Types:
- Run Tests: Execute test suite for SPEC
- Sync Worktree: Pull latest changes
- Switch Worktree: Change active folder

Task Configuration:
- Command to execute
- Working directory (worktree path)
- Presentation settings

---

## Shell Integration

### Enhanced Navigation

Improve shell experience for multi-worktree development.

Switch with Context:
- Save current worktree on switch
- Restore context on return
- Load worktree-specific environment

Quick Toggle:
- Toggle between last two worktrees
- Maintain work context
- Minimal switching overhead

### Status Overview

Display parallel development status.

Worktree Status:
- List all active worktrees
- Show running servers per worktree
- Display uncommitted changes

Server Status:
- Check PID files for running servers
- Report server types and ports
- Alert on stopped servers

### Quick Functions

Shell functions for common operations.

Create and Switch:
- Create new worktree
- Immediately switch to it
- Start development server

Develop in Worktree:
- Switch to specified worktree
- Start /moai:2-run

Push from Worktree:
- Sync worktree
- Push to remote branch

---

## Troubleshooting Parallel Development

### Common Issues

Port Conflicts:
- Symptom: Server fails to start
- Cause: Another worktree using same port
- Solution: Check port usage, adjust port assignment

Memory Exhaustion:
- Symptom: Operations fail or slow down
- Cause: Too many parallel operations
- Solution: Reduce concurrent worktrees, optimize memory

Disk Space:
- Symptom: Operations fail with disk errors
- Cause: Too many large worktrees
- Solution: Clean old worktrees, compress inactive ones

### Resolution Strategies

Port Conflict Resolution:
- List running processes per port
- Identify conflicting worktree
- Stop conflicting server or adjust port

Memory Optimization:
- Close unnecessary development servers
- Reduce parallel sync operations
- Use shallow clones for temporary worktrees

Disk Space Recovery:
- Clean build artifacts
- Remove unused dependencies
- Archive inactive worktrees

---

Version: 1.0.0
Last Updated: 2026-01-06
Module: Advanced parallel development patterns for feature flags, microservices, and optimization

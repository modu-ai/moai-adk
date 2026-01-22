# Development Mode Context

# Context injection mode for development work

# Version: 1.0.0

# Last Updated: 2026-01-22

---

mode: "development"
name: "Dev Mode"
description: "Code-first context for active development"

# Primary Focus

focus:

- "Code implementation"
- "Feature development"
- "Bug fixes"
- "Refactoring"
- "Testing"

# Context Priorities

priorities:
primary: - "Current working files" - "Active branch changes" - "Test files" - "Related dependencies"

secondary: - "Project structure" - "Configuration files" - "Documentation references"

excluded: - "Historical commits" - "Unrelated branches" - "Archive files" - "Generated documentation"

# Tool Preferences

tools:
preferred: - "Read (for current files)" - "Write (for implementation)" - "Edit (for modifications)" - "Bash (for testing and builds)" - "Grep (for code search)" - "Glob (for file discovery)"

minimized: - "WebSearch (use only when necessary)" - "WebFetch (use only when necessary)"

# Behavior Guidelines

behavior:
implementation: - "Write code directly to files" - "Create tests alongside implementation" - "Follow project coding standards" - "Maintain test coverage"

problem_solving: - "Identify root cause first" - "Propose minimal fixes" - "Consider edge cases" - "Validate with tests"

refactoring: - "Preserve existing behavior" - "Make incremental changes" - "Run tests frequently" - "Document breaking changes"

# Quality Standards

quality:
gates: - "Code must compile/build" - "Tests must pass" - "Coverage should not decrease" - "No new vulnerabilities introduced"

validation: - "Run linter before committing" - "Check test coverage" - "Verify no regressions" - "Review security implications"

# Communication Style

communication:
direct: - "Clear action statements" - "Specific file references" - "Concrete code examples" - "Explicit next steps"

minimal: - "Avoid lengthy explanations" - "Focus on implementation" - "Summarize decisions briefly" - "Document complex logic only"

# Error Handling

errors:
approach: - "Fix errors immediately" - "Identify root cause" - "Add tests for bug fixes" - "Document edge cases"

blocking: - "Stop on build failures" - "Stop on test failures" - "Stop on security issues" - "Ask for clarification if needed"

# Examples

examples:
feature_implementation:
description: "Implementing a new feature"
context: |
Focus on: 1. Reading current codebase structure 2. Understanding existing patterns 3. Implementing the feature 4. Writing comprehensive tests 5. Running validation checks

bug_fix:
description: "Fixing a bug"
context: |
Focus on: 1. Reproducing the bug 2. Identifying root cause 3. Implementing minimal fix 4. Adding regression test 5. Verifying no side effects

refactoring:
description: "Refactoring code"
context: |
Focus on: 1. Understanding current behavior 2. Writing characterization tests 3. Making incremental changes 4. Running tests frequently 5. Preserving behavior

# Mode Activation

activation:
manual: - "User explicitly requests dev mode" - "Working on implementation tasks" - "Debugging issues"

automatic: - "Detecting file modifications" - "Running tests or builds" - "Creating new code files"

# Mode Deactivation

deactivation:
triggers: - "Switching to review mode" - "Switching to research mode" - "Completing implementation" - "User requesting mode change"

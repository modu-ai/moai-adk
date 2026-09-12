## Supported Flags

- --pr: Push branch and create/update PR on GitHub after sync. When used, automatically returns to base branch (main/develop) after PR creation (Step 3.3.5).
- --merge: After sync, auto-merge PR and clean up branch. Worktree/branch environment is auto-detected from git context.
- --skip-mx: Skip MX tag validation and annotation during sync.


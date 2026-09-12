## Supported Flags

- --pr: Push branch and create/update PR on GitHub after sync. When used, automatically returns to base branch (main/develop) after PR creation (Step 3.3.5).
- --auto-merge: After sync, auto-merge PR and clean up branch (opt-in; merge conditions per `manager-git.md` § PR Auto-Merge). Worktree/branch environment is auto-detected from git context.
- --merge: deprecated alias of --auto-merge (logs a warning).
- --skip-mx: Skip MX tag validation and annotation during sync.


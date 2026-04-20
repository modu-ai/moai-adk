# File Reading Optimization

Guidelines for efficient file reading to minimize token consumption.

## Progressive File Loading

When reading files, apply tiered loading based on file size:

Tier 1 - Small files (under 200 lines):
- Read the full file with Read tool (no offset/limit needed)

Tier 2 - Medium files (200-500 lines):
- Read targeted sections using offset and limit parameters
- Use Grep to find relevant line numbers first, then Read with offset

Tier 3 - Large files (500-1000 lines):
- Never read the full file at once
- Use Grep to locate specific functions, classes, or patterns
- Read only the relevant sections with offset and limit
- Typical chunk size: 50-100 lines per Read call

Tier 4 - Very large files (over 1000 lines):
- Use Grep with output_mode "content" and context lines (-C) instead of Read
- Use Glob to find smaller, more specific files
- If full understanding is needed, delegate to Explore agent

## Practical Rules

Before reading any file:
- Use Grep to find exact line numbers of interest
- Then use Read with offset and limit to load only the relevant section

When exploring unfamiliar code:
- Start with Grep to find entry points (function names, class definitions)
- Read only the relevant sections, not entire files
- Use Glob to discover file structure before reading

When modifying code:
- Read only the section being modified plus sufficient context (20 lines before/after)
- Use Edit tool which requires reading only the old_string portion

## Token Budget Awareness

Each line read consumes approximately 10-20 tokens. A 1000-line file costs 10K-20K tokens.

For a 200K token context window:
- Reading 10 full 500-line files would consume 50-100K tokens (25-50% of budget)
- Reading targeted 50-line sections from 10 files would consume 5-10K tokens (2.5-5% of budget)

Always prefer targeted reads over full file reads when the task allows it.

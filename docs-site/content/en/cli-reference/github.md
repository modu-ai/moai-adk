---
title: moai github GitHub Integration
weight: 92
draft: false
---

`moai github` provides commands for GitHub issue parsing, SPEC linking, and workflow automation.

It accepts the common flag `--dry-run` (show what would be done without making changes).

## Subcommands

| Command | Description |
|--------|------|
| `moai github parse-issue <number>` | Parse a GitHub issue and display its contents |
| `moai github link-spec <issue-number> <spec-id>` | Bidirectionally link a GitHub issue and a SPEC document |

## moai github parse-issue

```bash
moai github parse-issue 123
```

Parses a GitHub issue and displays its contents.

## moai github link-spec

```bash
moai github link-spec 123 SPEC-AUTH-001
```

Creates a bidirectional link between a GitHub issue and a SPEC document.

## Examples

```bash
# Parse an issue
moai github parse-issue 42

# Link an issue to a SPEC (preview)
moai github link-spec 42 SPEC-AUTH-001 --dry-run

# Actual link
moai github link-spec 42 SPEC-AUTH-001
```

## Related documents

- [moai pr](/en/cli-reference/pr) — PR CI watch
- [CLI Overview](/en/getting-started/cli)

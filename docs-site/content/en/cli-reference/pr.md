---
title: moai pr PR Watch
weight: 96
draft: false
---

`moai pr` is a command that watches and manages pull requests in a CI/CD workflow.

## Subcommands

| Command | Description |
|--------|------|
| `moai pr watch <PR_NUMBER>` | Watch a PR's CI checks (or stop an active watch with `--abort`) |

## moai pr watch

```bash
moai pr watch 123
```

Monitors `gh pr checks` for the given PR number.

| Flag | Description |
|--------|------|
| `--abort` | Stop the active CI watch loop |
| `--report` | Emit a merge-readiness report for the given PR number |
| `--branch <name>` | Branch name for the report context (default: main) |

## Examples

```bash
# Watch PR CI checks
moai pr watch 42

# Stop the active watch
moai pr watch 42 --abort

# Merge-readiness report
moai pr watch 42 --report
```

## Related documents

- [moai github](/en/cli-reference/github)
- [CLI Overview](/en/getting-started/cli)

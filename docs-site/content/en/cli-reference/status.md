---
title: Project Status
weight: 20
draft: false
---

The `moai status` command shows the current project's initialization state, SPEC count, and configuration files at a glance. It is a read-only command with no flags.

## Usage

```bash
moai status
```

Run without flags, it prints the project status as a box.

## Output

### Initialized project

When run in a project where a `.moai/` directory exists, the following information is displayed.

| Item | Description |
|------|------|
| **Project** | Project name (current directory name) |
| **ADK** | Installed MoAI-ADK version |
| **Config** | Configuration file path (`.moai/config/sections`) |
| **SPECs** | Number of SPEC directories under `.moai/specs/` |
| **Configs** | Number of YAML files in `.moai/config/sections/` |

A status indicator showing the initialization state and SPEC count is printed at the bottom.

### Uninitialized project

When no `.moai/` directory exists, a "Not initialized" status indicator is shown along with guidance to run `moai init`.

## BODP branch notice

When the project is a Git repository and the current branch strays from the BODP (Branch-Oriented Development Practice) convention, a notice is printed to stderr. This notice is a reminder of the branch-naming convention within the distributed one-person OSS workflow.

The notice is printed automatically, and is silently skipped when Git is not installed or the current directory is not a Git repository.

## Related commands

| Command | Description |
|--------|------|
| `moai doctor` | System diagnostics and environment validation (detailed check) |
| `moai inventory` | Unified view of active sessions, worktrees, and harnesses |
| `moai init` | Project initialization (run when uninitialized) |

## See also

- [CLI Reference](./cli) — full CLI commands
- [moai inventory](./inventory) — unified view of active resources
- [Initial Setup](./init-wizard) — project initialization wizard

---
title: moai harness Harness
weight: 80
draft: false
---

`moai harness` is a unified command tree that manages SPEC-complexity routing and the harness learning subsystem. It provides subcommands for routing, validation, lifecycle, proposal management, the v4 harness lifecycle, and the observation ledger.

It accepts the common flag `--project-root <path>` (default: current directory).

## Routing verbs

| Command | Description |
|--------|------|
| `moai harness route --spec <id>` | Route a SPEC to a minimal/standard/thorough harness level |
| `moai harness validate` | Validate harness.yaml against the schema and invariants |

`route` accepts `--json` (JSON output), `--path <harness.yaml>`, and `--base-dir <dir>`.

## Lifecycle verbs

| Command | Description |
|--------|------|
| `moai harness status` | Show observation/tier/evolution summary |
| `moai harness apply` | Return pending proposals to the orchestrator (or run the Go apply path with `--execute`) |
| `moai harness rollback <date>` | Restore the snapshot for the given date |
| `moai harness disable` | Disable the learning subsystem (`learning.enabled: false`) |

## Proposal-management verbs

| Command | Description |
|--------|------|
| `moai harness mute` | Mute a proposal category (workflow.yaml) |
| `moai harness mute-list` | Print the currently muted categories |
| `moai harness unmute` | Remove a category from the mute list |
| `moai harness verify` | Verify harness determinism |

## v4 harness lifecycle verbs

| Command | Description |
|--------|------|
| `moai harness list` | List all v4 harnesses (name + domain + entry command) |
| `moai harness edit <name>` | Show the edit paths for a v4 harness manifest + specialist |
| `moai harness remove <name>` | Atomically remove a v4 harness (command + workflow + specialist + skill + manifest) |
| `moai harness doctor` | Diagnose harness installation status |

`list` and `edit` accept the `--json` flag.

## Observation ledger

`moai harness ledger` manages the routing observation ledger.

| Command | Description |
|--------|------|
| `moai harness ledger record` | Record the routing decision at dispatch time (pending row) |
| `moai harness ledger evidence` | Add a machine-evidence ref (or delegation entry) to a pending row |
| `moai harness ledger list` | List final ledger rows with filters |

> `ledger record` and `ledger evidence` do not expose an `--outcome` flag, so the outcome cannot be forged. The outcome is derived from the machine evidence.

## Related documents

- [Harness Self-Evolution](/en/advanced/self-evolving)
- [Harness v4 Builder In-Depth Guide](/en/advanced/harness-v4-builder)
- [Harness Profiles and Evaluation](/en/advanced/harness-profiles)
- [CLI Overview](/en/getting-started/cli)

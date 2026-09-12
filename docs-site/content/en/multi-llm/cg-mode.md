---
title: CG retirement and migration
weight: 20
draft: false
description: CG retirement and migration
---

`moai cg` has been retired. It exits with a migration diagnostic without starting Claude or GLM. It is not an alias for `moai cc`. Projects with `llm.team_mode: cg` must make an explicit migration choice before launching a session.

## Preview

Preview the choices from the project root. Preview does not change the configuration or create a backup.

```bash
moai migrate cg
moai migrate cg --target claude-only
```

## Apply the Claude-only role change

Apply a role change only if you accept removing the automatic assignment of GLM teammates:

```bash
moai migrate cg --target claude-only --apply --accept-role-change
```

This writes `llm.team_mode: claude`, `llm.gateway.teammate_mode: in-process`, and `llm.gateway.teammate_provider: inherit`. It removes the old hybrid role assignment; it does not preserve a Claude leader with GLM teammate panes.

## Hybrid migration gate

The `claude-glm` target describes a Claude leader with GLM teammates in tmux. Its apply and launch paths are currently unavailable because the TEAMMATE integration gate has not passed. Preview is available. Installing tmux or setting `verified: true` does not open this gate.

```bash
moai migrate cg --target claude-glm
```

## Configuration preservation and errors

The migration preserves unknown configuration values, comments, GLM model settings, and credential references. It saves the exact original bytes to `.moai/backups/cg-migration/<source-sha256>.yaml` before applying the change. YAML formatting can change. Running the same target again returns unchanged; changing to the other target is rejected.

Conflicting gateway values, nonempty `llm.mode`, duplicate keys, and unsupported YAML aliases must be resolved explicitly. Preview and preflight failures leave the source and backup directory unchanged. A write failure can leave a backup; inspect the reported error and the preserved original before retrying.

## Credential handling {#tmux-env-security}

The former CG environment-injection instructions no longer describe an available launcher. Migration does not launch providers or move credentials. Keep backups private because they contain the original configuration.

## Next steps

After an accepted `claude-only` migration, choose a supported launcher explicitly. `moai cc` and `moai glm` do not recreate the old hybrid roles. GPT gateway launch availability is a separate integration gate; this migration does not enable it.

- [CLI](/en/cli-reference/launchers/)

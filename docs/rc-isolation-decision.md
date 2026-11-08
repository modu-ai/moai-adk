# RC and existing Go installation isolation

The user approved separating the new RC executable, update path, and state from the existing Go installation. Claude and Codex host transitions must be explicit, and existing user settings must be preserved.

This records direction only: final executable naming, package boundaries, filesystem paths, migration and updater mechanisms remain to be specified and verified. It does not authorize overwriting the existing Go project, binaries, global settings, or credentials.

## Confirmed temporary command transition

The user confirmed `moai-cli` as the temporary command name during RC testing. The planned end state is a tested new implementation under `moai`, followed by removal of the old Go installation only after its precise ownership and removal scope are known. No existing binary, settings, state, or Go source is removed by this decision.

Before any cutover, identify the actual executable and owner, package-manager or manual installation method, paths and symlinks, version and digest, state and configuration dependencies, and active users. Produce a preview, a backup and restore procedure, and an exact verification plan. Test the isolated new command and both host transitions, user configuration preservation, updater channel isolation and restoration in fixtures. Treat missing evidence as blocked. Applying a destructive uninstall or replacing the existing command requires authorization for the concrete removal and cutover plan.

Observed locally: `command -v moai` returned `$HOME/go/bin/moai`; `moai-cli` was not found in the current PATH. This is a lookup, not a full installation or ownership inventory.

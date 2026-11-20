# The project state folder layout

The rule behind SPEC-20261120-moai-layout. Every MoAI project keeps its MoAI files in one folder, `.moai/` (an earlier project still on `.moai-cli/` follows the same rule below that folder). The folder has two halves: what the team shares, which git tracks, and what one machine makes, which git ignores.

```
.moai/
  moai.json  moai-cli.json  project.yaml    markers and profile      tracked
  config/ decisions/ docs/ specs/           the team's own files      tracked
  manifests/ workflows/                     generation records        tracked
  tools/                                    project development scripts the team shares   tracked
  evals/                                    evaluation sets           tracked
  local/                                    this machine's output     ignored
    receipts/ backups/ index/ logs/ state/ cache/ …
```

`src/state_root.rs` is the one place that names the kinds:

| Name | Meaning |
| --- | --- |
| `KEPT` | the eleven names the team keeps in git, in the order of the managed `.gitignore` block's `!` lines |
| `LOCAL_KINDS` | the machine-output kinds that exist today; anything not in `KEPT` is local whether or not it is listed |
| `local_path(root, kind)` | `<state>/local/<kind>` |
| `earlier_path(root, kind)` | `<state>/<kind>`, where an earlier project kept it |
| `write_path(root, kind)` | where a writer creates the kind: always `local_path` |
| `read_path(root, kind)` | `local_path` when it exists, else `earlier_path`, else `None` |
| `HOST_OWNED` | `AGENTS.md`, `.claude`, `.codex`, `.github/ISSUE_TEMPLATE`: the host reads these in a fixed place, the layout never moves them |

The managed `.gitignore` block (`src/agents.rs`) writes the receipts line, `<state>/*`, `<state>/local/`, then one `!` line per `KEPT` name. `<state>/*` already ignores everything that is not allowed, so the `local/` line states the rule for a reader; it does not change what git ignores. A block written before the layout rule differs by three lines (`local/`, `tools/`, `evals/`); the project update plan shows it as a `merge` and writes nothing until the user applies the plan.

Existing projects keep working: a reader asks `read_path` and finds the earlier folder when `local/<kind>` is absent; a writer creates only `local/<kind>` and leaves the earlier folder as it was. Moving each writer to `local/` is done kind by kind in later slices of the SPEC; this slice defines the rule and the block.

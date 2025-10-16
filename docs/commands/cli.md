# MoAI-ADK CLI Reference

MoAI-ADK ships with a Click-based command line interface that wraps the agentic project lifecycle utilities implemented in `src/moai_adk`. The CLI can be started with either the convenience entry point (once the package exposes it) or directly via Python:

```bash
python -m moai_adk            # invoke the Click group defined in __main__.py
python -m moai_adk --help      # show available commands
```

> **Note**: At the moment the project does not define a `console_scripts` entry point in `pyproject.toml`, so invoking the CLI through `python -m moai_adk` (or `uv run python -m moai_adk`) is the most reliable approach.

## Commands at a Glance

| Command | Purpose | Key Options |
| --- | --- | --- |
| `init [PATH]` | Bootstrap or re-bootstrap a project using the phase executor pipeline. | `--non-interactive/-y`, `--mode`, `--locale`, `--language`, `--force` |
| `doctor` | Run environment health checks (Python, Git, project structure). | _None_ |
| `status` | Inspect `.moai/config.json`, SPEC counts, and Git status. | _None_ |
| `backup` | Create a timestamped backup of the local agent resources. | `--path` |
| `restore` | Restore from an existing backup set. | `--timestamp` |
| `update` | Refresh templates and supporting assets to the bundled version. | `--path`, `--force`, `--check` |

All commands share the same Rich console instance for colourised output and surface `click.Abort` for user cancellations, so they integrate well with shell scripting.

## `init [PATH]`

Initialises a directory with the MoAI-ADK structure. The heavy lifting is handled by `ProjectInitializer`, which delegates to `PhaseExecutor` to run five deterministic phases (prepare → directory → resources → configuration → validation).

```bash
python -m moai_adk init .
python -m moai_adk init ./demo --non-interactive --mode team --locale en
python -m moai_adk init . --force           # skip confirmation when reinitialising
```

- With interactive mode (default) the command uses `questionary` prompts from `cli.prompts.init_prompts` to gather project metadata.
- Reinitialisation triggers an automatic backup in `.moai/backups/<timestamp>/` unless `--force` is supplied. Protected paths (SPECs, reports, project docs) are preserved.
- Language is auto-detected via `LanguageDetector` if `--language` is not provided.

## `doctor`

Executes `core.project.checker.check_environment()` and renders the result as a Rich table. The diagnostic verifies:

- Python runtime ≥ 3.13
- Git availability on `PATH`
- Existence of `.moai/` and `.moai/config.json`

The command exits with a warning banner when any check fails so that CI pipelines can capture the failure.

## `status`

Summarises the current project state by reading `.moai/config.json` and counting SPEC documents under `.moai/specs/`. If Git metadata is available (via GitPython) it also reports the active branch and whether the worktree is dirty.

```bash
python -m moai_adk status
```

Missing configuration aborts the command and suggests running `init` first.

## `backup`

Creates a timestamped archive of the live Claude/MoAI resources in `.moai-backup/<timestamp>/`.

```bash
python -m moai_adk backup
python -m moai_adk backup --path /path/to/project
```

Internally this reuses `TemplateProcessor.create_backup()` so protected resources (SPECs, reports) are omitted and pre-existing structures are preserved.

## `restore`

Finds the newest backup under `.moai/backups/` or restores a specific snapshot when `--timestamp` is given.

```bash
python -m moai_adk restore
python -m moai_adk restore --timestamp 2025-03-01-130500
```

> ⚠️ The current implementation only reports which backup would be restored. Actual file restoration still needs to be implemented, which is reflected in the roadmap.

## `update`

Refreshes templates, Claude commands, and configuration defaults to the version bundled with the package. When `--check` is set it prints a diff-style version summary without modifying files.

```bash
python -m moai_adk update
python -m moai_adk update --check
python -m moai_adk update --force          # skip backup creation
```

The command runs `TemplateProcessor.copy_templates()` with backup support and logs which resources were touched. It preserves SPECs and reports and performs smart merges for `CLAUDE.md` and `.gitignore` via the TemplateProcessor helper methods.

## Exit Codes & Error Handling

- `click.Abort` is translated into exit code `130` by the top-level `main()` wrapper.
- Unexpected exceptions are rendered with Rich styling and produce exit code `1`.
- Individual commands raise `click.ClickException` when they need to bubble up meaningful error messages to the caller.

These semantics allow the CLI to be embedded safely in automation pipelines while still being friendly for interactive terminal use.

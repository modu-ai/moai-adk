# Package Architecture Overview

This guide summarises the structure of the code under `src/moai_adk` so that the online documentation stays in sync with the implementation.

## Top-Level Package (`moai_adk`)

- `__main__.py` defines the Click command group, wires in the command modules, prints the Rich/Pyfiglet banner, and exposes `main()` for use with `python -m moai_adk`.
- `cli/main.py` simply re-exports the `cli` group so other tooling can reuse it without touching `__main__` directly.
- `__init__.py` exposes `__version__` (currently `0.3.0`) for use by the CLI and external consumers.

## CLI Commands (`cli/commands`)

Each command is a standalone Click command that relies on high-level services from the `core` package:

- `init.py` handles interactive vs. non-interactive project bootstrap, progress reporting, and re-initialisation safeguards. It uses `ProjectInitializer` to execute the five-phase pipeline and `prompt_project_setup` for Questionary prompts.
- `doctor.py` renders the environment diagnostics returned by `core.project.checker.check_environment()` as a Rich table, surfacing Python, Git, and project structure issues.
- `status.py` inspects `.moai/config.json`, counts SPEC documents, and optionally reports Git branch/dirty state via GitPython.
- `backup.py`, `restore.py`, and `update.py` wrap the template processor utilities for lifecycle management of `.moai` and `.claude` assets. `restore` currently acts as a placeholder that only reports the chosen backup.

The CLI layer consistently uses a shared Rich `Console` instance and translates aborts into meaningful exit codes in `__main__.py`.

## Core Project Logic (`core/project`)

- `initializer.py` hosts `ProjectInitializer`, which orchestrates installation using `PhaseExecutor`. It returns an `InstallationResult` that the CLI can render. The initializer also enforces duplicate initialisation rules and language detection.
- `phase_executor.py` encapsulates the five phases (preparation, directory creation, resource copy, configuration generation, validation). It coordinates backups, template copying, and optional Git initialisation for team mode.
- `detector.py` (`LanguageDetector`) scans for language-specific markers (20 presets) to set sensible defaults when the user skips `--language`.
- `validator.py` performs system checks (Git availability, Python version), project path validation, and post-install verification of required files/directories. It raises specialised `ValidationError` exceptions when requirements are not met.
- `checker.py` provides a lighter-weight environment check used by the `doctor` command. `SystemChecker` exposes reusable `check_all()` logic, while the module-level `check_environment()` returns the CLI's status dictionary.
- `backup_utils.py` implements selective backup helpers that back up only relevant assets and skip protected content such as SPECs and reports.

## Template Management (`core/template`)

- `processor.py` packages all template handling logic. It knows where bundled templates live, copies `.claude/`, `.moai/`, `CLAUDE.md`, and `.gitignore`, performs smart merges, and can create timestamped backups in `.moai-backup/`.
- `config.py` introduces `ConfigManager` for loading, saving, and deep-merging `.moai/config.json` documents.
- `languages.py` maps detected languages to template files and exposes `get_language_template()` for consumers that need language-specific Jinja templates.

## Git Utilities (`core/git`)

- `manager.py` wraps GitPython's `Repo` class with convenience methods (`is_repo`, `current_branch`, `create_branch`, etc.).
- `branch.py` and `commit.py` expose small utilities for SPEC-based branch naming and TDD-flavoured commit messages.

## Supporting Utilities

- `cli/prompts/init_prompts.py` defines the Questionary-based interview used by `init` for collecting project metadata.
- `utils/banner.py` stores the ASCII art banner and helper functions used by the CLI introduction.

## Bundled Templates (`templates/`)

The package distributes a full set of starter assets:

- `.moai/` defaults (config skeleton, memory guides, project docs).
- `.claude/` agents, commands, settings, and hooks tailored for Alfred.
- `CLAUDE.md` and `.gitignore` base files.
- `.github/workflows` scaffolding for the GitHub Actions pipeline.

`TemplateProcessor` is the single integration point for copying and backing up these templates, so future changes should flow through that class to maintain consistent behaviour across CLI commands.

## Notable Gaps / Follow-up Tasks

- `restore` needs real file copy logic to complement the generated backups.
- The project currently requires Python ≥ 3.13, which narrows the potential user base. Assess whether the codebase truly depends on 3.13-only features.
- No `console_scripts` entry point is defined yet, so publishing to PyPI would not expose a `moai-adk` executable by default.
- Several tests referenced in docstrings (e.g. `tests/unit/test_foundation.py`, `test_cli_commands.py`) are not present, which can confuse contributors and tooling.

This overview should make it easier to keep the hosted documentation aligned with the actual implementation and to plan upcoming refactors.

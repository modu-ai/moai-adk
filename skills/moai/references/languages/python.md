# Language: Python

When to read: `internal/quality` sent you here because the project's detected languages include `python`. The project's own declarations (CI steps, `Makefile`, `tox.ini`, `noxfile.py`, `AGENTS.md`) win over every default below.

## Which runner

The lockfile names the environment manager; run every command through it:

| Evidence | Prefix |
| --- | --- |
| `uv.lock` | `uv run <command>` |
| `poetry.lock` | `poetry run <command>` |
| `Pipfile.lock` | `pipenv run <command>` |
| `tox.ini` or `noxfile.py` with the check as an environment or session | `tox -e <env>` or `nox -s <session>` |
| none of these | the project's documented virtual environment; never the system interpreter's site-packages |

## Evidence and defaults

| Kind | Default command | Use it only when |
| --- | --- | --- |
| format | `ruff format --check .` or `black --check .` | `[tool.ruff]`/`ruff.toml`, or `[tool.black]`, is configured |
| lint | `ruff check .` or `flake8` | `[tool.ruff]`/`ruff.toml`, or `.flake8`/`setup.cfg [flake8]` |
| type check | `mypy` or `pyright` | `[tool.mypy]`/`mypy.ini`, or `pyrightconfig.json`/`[tool.pyright]` |
| test | `python -m pytest` | pytest is configured (`[tool.pytest.ini_options]`, `pytest.ini`, `conftest.py`) or declared as a dependency |
| test | `python -m unittest discover` | no pytest evidence and test files named `test*.py` exist |
| build | `python -m build` | the project publishes a package (`[build-system]` in `pyproject.toml`) |

## Dependency existence

- `https://pypi.org/pypi/<name>/json` answers 404 for a name that does not exist; `python -m pip index versions <name>` prints the versions (pip marks the command experimental).
- The import name is not the distribution name: `import yaml` comes from `PyYAML`, `import cv2` from `opencv-python`. Check the distribution name, not the import.
- Add with the project's manager (`uv add`, `poetry add`) so the lockfile records it.

## Traps

- `pytest -k <expr>` that matches nothing ends with `no tests ran` and exit code 5; treat it as `not_run`.
- `pip install` outside the project's environment changes the machine, not the project.
- A skipped test (`@pytest.mark.skip`, `skipIf`, `unittest.skip`) passes silently; a new skip is a finding.
- Formatters without `--check` rewrite files; a VERIFY run never rewrites.

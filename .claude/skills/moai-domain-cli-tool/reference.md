# CLI Tool Development Reference

> Official documentation and standards for building professional command-line interfaces

---

## Official Documentation Links

### Primary Frameworks

| Tool | Version | Documentation | Status |
|------|---------|--------------|--------|
| **Click** | 8.1.7 | https://click.palletsprojects.com/ | ✅ Current (Python ≥3.10) |
| **Typer** | 0.15.0 | https://typer.tiangolo.com/ | ✅ Current |
| **Rich** | 13.9.0 | https://rich.readthedocs.io/ | ✅ Current |

### Standards & Conventions

- **POSIX Utility Conventions**: https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap12.html
- **GNU Coding Standards**: https://www.gnu.org/prep/standards/html_node/Command_002dLine-Interfaces.html

---

## POSIX Argument Parsing Standards

### Core Conventions

**Option Format**:
- Single character options: `-o`
- Options with arguments: `-o argument` or `-oargument`
- Grouped options (no arguments): `-lst` = `-t -l -s`
- Long options (GNU extension): `--long-option`

**Argument Order**:
```bash
command [options] [--] [arguments]
```

**Special Arguments**:
- `--` terminates option processing
- `-` represents standard input/output stream

### Best Practices

Always handle `--version` and `--help` for programs on POSIX systems. Have the `--help` message list all the options and default values of options, and perhaps important environment variables.

---

## Click Framework (v8.1.7)

### Core Concepts

**Command Structure**:
```python
import click

@click.group()
def cli():
    """Main CLI entry point."""
    pass

@cli.command()
@click.option('--count', default=1, help='Number of iterations')
def run(count):
    """Run the main operation."""
    for i in range(count):
        click.echo(f"Iteration {i+1}")
```

### Best Practices

1. **Naming Conventions**: Name the entry-point command `cli()` as a common practice
2. **Keep Interfaces Simple**: Avoid too many commands or options; focus on core functionality
3. **Use Descriptive Names**: Commands and options should be self-documenting
4. **Provide Help Text**: Always include help strings for commands and options

### Arguments vs. Options

**Arguments** (mandatory, positional):
- Listed in the order they appear on command line
- Generally required values

**Options** (optional, named):
- Can appear in any order
- Support defaults and validation

---

## Typer Framework (v0.15.0)

### Key Features

- **Type Hints**: Uses Python type annotations for automatic validation
- **Auto-completion**: Built-in support for Bash, Zsh, Fish, PowerShell
- **Minimal Code**: 2 lines minimum (1 import + 1 function call)

### Installation

```bash
pip install typer[all]  # Includes all optional dependencies
```

### Basic Example

```python
import typer

def main(name: str, age: int = 25, formal: bool = False):
    """Greet a person by name."""
    greeting = "Hello" if formal else "Hi"
    typer.echo(f"{greeting}, {name}! You are {age} years old.")

if __name__ == "__main__":
    typer.run(main)
```

---

## Rich for Enhanced Terminal Output (v13.9.0)

### Core Features

- **Syntax Highlighting**: For code, JSON, Markdown
- **Progress Bars**: For long-running operations
- **Tables**: For structured data display
- **Panels**: For grouped content

### Progress Bar Example

```python
from rich.progress import track
import time

for i in track(range(100), description="Processing..."):
    time.sleep(0.01)
```

---

## Essential CLI Patterns

### Configuration File Handling

```python
import click
import json
from pathlib import Path

@click.command()
@click.option('--config', type=click.Path(), default='config.json')
def deploy(config):
    """Deploy using configuration file."""
    config_path = Path(config)
    if config_path.exists():
        with config_path.open() as f:
            settings = json.load(f)
```

### Environment Variable Support

```python
import click

@click.command()
@click.option('--api-key', envvar='API_KEY', required=True)
def fetch(api_key):
    """Fetch data using API key."""
    click.echo(f"Using key: {api_key[:4]}****")
```

---

## Testing CLI Applications

### Click Testing Utilities

```python
from click.testing import CliRunner

def test_cli_command():
    runner = CliRunner()
    result = runner.invoke(cli, ['--help'])
    assert result.exit_code == 0
    assert 'Usage:' in result.output
```

---

## Distribution & Packaging

### pyproject.toml (Modern)

```toml
[project]
name = "mycli"
version = "1.0.0"
dependencies = [
    "click>=8.1.7",
    "typer>=0.15.0",
    "rich>=13.9.0",
]

[project.scripts]
mycli = "mycli:cli"
```

---

## Additional Resources

### Documentation

- **Click Tutorial**: https://click.palletsprojects.com/en/latest/quickstart/
- **Typer Tutorial**: https://typer.tiangolo.com/tutorial/
- **Rich Documentation**: https://rich.readthedocs.io/en/stable/
- **Real Python - Click**: https://realpython.com/python-click/

### Standards

- **POSIX Utility Conventions**: https://pubs.opengroup.org/onlinepubs/9699919799/
- **GNU Coding Standards**: https://www.gnu.org/prep/standards/

---

**Last Updated**: 2025-10-22
**Framework Versions**: Click 8.1.7, Typer 0.15.0, Rich 13.9.0
**Python Requirement**: ≥3.10

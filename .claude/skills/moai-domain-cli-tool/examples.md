# CLI Tool Development - Working Examples

> Real-world examples using Click, Typer, and Rich

---

## Example 1: Simple CLI with Click

### Basic Command with Options

```python
# cli.py
import click

@click.command()
@click.option('--count', default=1, help='Number of greetings')
@click.option('--name', prompt='Your name', help='The person to greet')
def hello(count, name):
    """Simple program that greets NAME for a total of COUNT times."""
    for _ in range(count):
        click.echo(f'Hello, {name}!')

if __name__ == '__main__':
    hello()
```

**Usage**:
```bash
python cli.py --count 3 --name Alice
# Output:
# Hello, Alice!
# Hello, Alice!
# Hello, Alice!
```

---

## Example 2: CLI Tool with Command Groups (Click)

### Multi-command CLI (like git or docker)

```python
# cli.py
import click

@click.group()
def cli():
    """Database CLI tool."""
    pass

@cli.command()
@click.option('--force', is_flag=True, help='Force initialization')
def init(force):
    """Initialize the database."""
    if force:
        click.echo('Force initializing database...')
    else:
        click.echo('Initializing database...')

@cli.command()
@click.argument('version')
def migrate(version):
    """Migrate database to VERSION."""
    click.echo(f'Migrating to version {version}...')

@cli.command()
def status():
    """Show database status."""
    click.secho('✓ Database is healthy', fg='green', bold=True)

if __name__ == '__main__':
    cli()
```

**Usage**:
```bash
python cli.py init --force
python cli.py migrate v2.0
python cli.py status
```

---

## Example 3: Type-Safe CLI with Typer

### Using Python Type Hints

```python
# main.py
import typer
from pathlib import Path
from typing import Optional

app = typer.Typer()

@app.command()
def process(
    input_file: Path,
    output_file: Optional[Path] = None,
    verbose: bool = False,
    workers: int = 4
):
    """
    Process INPUT_FILE with optional OUTPUT_FILE.

    Args:
        input_file: Path to input file (must exist)
        output_file: Path to output file (optional)
        verbose: Enable verbose logging
        workers: Number of parallel workers (default: 4)
    """
    if not input_file.exists():
        typer.secho(f"Error: {input_file} does not exist", fg=typer.colors.RED)
        raise typer.Exit(code=1)

    if verbose:
        typer.echo(f"Processing {input_file} with {workers} workers...")

    # Process file logic here
    typer.secho("✓ Processing complete!", fg=typer.colors.GREEN)

if __name__ == "__main__":
    app()
```

**Usage**:
```bash
python main.py process input.txt --verbose --workers 8
python main.py process input.txt --output-file result.txt
```

---

## Example 4: Rich Terminal Output

### Progress Bars and Tables

```python
# rich_example.py
import click
from rich.console import Console
from rich.table import Table
from rich.progress import track
from rich.panel import Panel
import time

console = Console()

@click.command()
@click.option('--format', type=click.Choice(['table', 'json']), default='table')
def report(format):
    """Generate a report with rich formatting."""

    if format == 'table':
        # Create a table
        table = Table(title="User Report", show_header=True, header_style="bold magenta")
        table.add_column("Name", style="cyan")
        table.add_column("Status", style="green")
        table.add_column("Score", justify="right", style="yellow")

        table.add_row("Alice", "Active", "95")
        table.add_row("Bob", "Inactive", "72")
        table.add_row("Charlie", "Active", "88")

        console.print(table)

    # Progress bar for processing
    console.print("\n[bold blue]Processing data...[/bold blue]")
    for i in track(range(100), description="Loading..."):
        time.sleep(0.01)

    # Success panel
    success_panel = Panel(
        "[bold green]✓ Report generated successfully![/bold green]",
        title="Success",
        border_style="green"
    )
    console.print(success_panel)

if __name__ == '__main__':
    report()
```

**Output**:
```
┌─────────────────────────────────────┐
│          User Report                │
├──────────┬──────────┬───────────────┤
│ Name     │ Status   │ Score         │
├──────────┼──────────┼───────────────┤
│ Alice    │ Active   │            95 │
│ Bob      │ Inactive │            72 │
│ Charlie  │ Active   │            88 │
└──────────┴──────────┴───────────────┘

Processing data...
Loading... ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 100% 0:00:01

╭─ Success ────────────────────────────────╮
│ ✓ Report generated successfully!        │
╰──────────────────────────────────────────╯
```

---

## Example 5: Configuration File Support

### Loading Settings from File

```python
# config_cli.py
import click
import json
from pathlib import Path

@click.group()
@click.option('--config', type=click.Path(exists=True), default='config.json')
@click.pass_context
def cli(ctx, config):
    """CLI with configuration file support."""
    ctx.ensure_object(dict)

    config_path = Path(config)
    if config_path.exists():
        with config_path.open() as f:
            ctx.obj['config'] = json.load(f)
    else:
        ctx.obj['config'] = {}

@cli.command()
@click.pass_context
def show_config(ctx):
    """Display current configuration."""
    config = ctx.obj.get('config', {})
    click.echo(json.dumps(config, indent=2))

@cli.command()
@click.pass_context
@click.argument('key')
def get(ctx, key):
    """Get a configuration value."""
    config = ctx.obj.get('config', {})
    value = config.get(key, 'Not found')
    click.echo(f"{key}: {value}")

if __name__ == '__main__':
    cli()
```

**config.json**:
```json
{
  "api_url": "https://api.example.com",
  "timeout": 30,
  "verbose": true
}
```

**Usage**:
```bash
python config_cli.py --config config.json show-config
python config_cli.py get api_url
```

---

## Example 6: Environment Variables and Secrets

### Secure Credential Handling

```python
# secure_cli.py
import typer
import os
from getpass import getpass

app = typer.Typer()

@app.command()
def login(
    username: str = typer.Option(..., envvar="USERNAME", prompt=True),
    api_key: str = typer.Option(..., envvar="API_KEY", hide_input=True)
):
    """
    Login with credentials (supports env vars).

    Set environment variables:
        export USERNAME=myuser
        export API_KEY=secret123
    """
    typer.echo(f"Logging in as: {username}")
    typer.echo(f"API Key: {api_key[:4]}****")

    # Perform login logic
    typer.secho("✓ Login successful!", fg=typer.colors.GREEN)

if __name__ == "__main__":
    app()
```

**Usage**:
```bash
# Interactive prompt
python secure_cli.py login

# With environment variables
export USERNAME=alice
export API_KEY=secret_key_12345
python secure_cli.py login

# Command line (not recommended for secrets)
python secure_cli.py login --username alice --api-key secret123
```

---

## Example 7: Testing CLI Applications

### Unit Tests with Click Testing Utilities

```python
# test_cli.py
import pytest
from click.testing import CliRunner
from cli import hello  # From Example 1

def test_hello_default():
    runner = CliRunner()
    result = runner.invoke(hello, input='Alice\n')
    assert result.exit_code == 0
    assert 'Hello, Alice!' in result.output

def test_hello_with_count():
    runner = CliRunner()
    result = runner.invoke(hello, ['--count', '3', '--name', 'Bob'])
    assert result.exit_code == 0
    assert result.output.count('Hello, Bob!') == 3

def test_hello_help():
    runner = CliRunner()
    result = runner.invoke(hello, ['--help'])
    assert result.exit_code == 0
    assert 'Usage:' in result.output
    assert '--count' in result.output

# Run tests
if __name__ == '__main__':
    pytest.main([__file__, '-v'])
```

**Run tests**:
```bash
pytest test_cli.py -v

# Output:
# test_cli.py::test_hello_default PASSED
# test_cli.py::test_hello_with_count PASSED
# test_cli.py::test_hello_help PASSED
```

---

## Example 8: Packaging and Distribution

### pyproject.toml Setup

```toml
[build-system]
requires = ["setuptools>=61.0"]
build-backend = "setuptools.build_meta"

[project]
name = "mycli"
version = "1.0.0"
description = "A sample CLI tool"
authors = [
    {name = "Your Name", email = "you@example.com"}
]
requires-python = ">=3.10"
dependencies = [
    "click>=8.1.7",
    "typer>=0.15.0",
    "rich>=13.9.0",
]

[project.optional-dependencies]
dev = [
    "pytest>=7.0",
    "black>=23.0",
    "ruff>=0.1.0",
]

[project.scripts]
mycli = "mycli.cli:main"

[tool.black]
line-length = 100
target-version = ['py310']

[tool.ruff]
line-length = 100
select = ["E", "F", "W", "I", "N"]
```

**Install and use**:
```bash
# Install in development mode
pip install -e .

# Run CLI
mycli --help

# Build and distribute
pip install build
python -m build
# Creates dist/mycli-1.0.0.tar.gz and dist/mycli-1.0.0-py3-none-any.whl
```

---

## TDD Workflow Example

### RED → GREEN → REFACTOR

**Step 1: RED (Write failing test)**

```python
# test_calculator.py
from click.testing import CliRunner
from calculator import cli

def test_add_command():
    runner = CliRunner()
    result = runner.invoke(cli, ['add', '2', '3'])
    assert result.exit_code == 0
    assert 'Result: 5' in result.output
```

**Step 2: GREEN (Implement feature)**

```python
# calculator.py
import click

@click.group()
def cli():
    """Simple calculator CLI."""
    pass

@cli.command()
@click.argument('a', type=int)
@click.argument('b', type=int)
def add(a, b):
    """Add two numbers."""
    result = a + b
    click.echo(f'Result: {result}')

if __name__ == '__main__':
    cli()
```

**Step 3: REFACTOR (Improve code)**

```python
# calculator.py (refactored)
import click
from typing import Callable

@click.group()
def cli():
    """Simple calculator CLI."""
    pass

def operation_command(name: str, func: Callable[[int, int], int]):
    """Factory function for operation commands."""
    @cli.command(name=name)
    @click.argument('a', type=int)
    @click.argument('b', type=int)
    def command(a, b):
        result = func(a, b)
        click.echo(f'Result: {result}')
    return command

# Register operations
add = operation_command('add', lambda a, b: a + b)
subtract = operation_command('subtract', lambda a, b: a - b)
multiply = operation_command('multiply', lambda a, b: a * b)

if __name__ == '__main__':
    cli()
```

---

## Best Practices Summary

1. **Use type hints** with Typer for automatic validation
2. **Provide clear help text** for all commands and options
3. **Handle errors gracefully** with informative messages
4. **Support configuration files** for complex setups
5. **Use environment variables** for secrets and credentials
6. **Write tests** using Click/Typer testing utilities
7. **Follow POSIX conventions** for option naming
8. **Add progress indicators** for long-running operations
9. **Package properly** using pyproject.toml
10. **Document your CLI** with --help and README

---

**Last Updated**: 2025-10-22
**Frameworks**: Click 8.1.7, Typer 0.15.0, Rich 13.9.0

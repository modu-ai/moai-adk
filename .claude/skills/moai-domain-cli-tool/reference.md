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

## Advanced CLI Patterns

### Interactive Prompts

**Confirmation Dialogs**:
```python
import click

@click.command()
def dangerous_operation():
    """Perform a dangerous operation."""
    if click.confirm('Are you sure you want to continue?'):
        click.echo('Proceeding with operation...')
    else:
        click.echo('Operation cancelled.')
```

**Choice Selection**:
```python
@click.command()
def deploy():
    """Deploy to an environment."""
    env = click.prompt(
        'Select environment',
        type=click.Choice(['dev', 'staging', 'prod']),
        default='dev'
    )
    click.echo(f'Deploying to {env}...')
```

**Password Input**:
```python
@click.command()
def login():
    """Login with credentials."""
    password = click.prompt('Enter password', hide_input=True, confirmation_prompt=True)
    click.echo('Password accepted')
```

---

### Output Formatting

**JSON Output**:
```python
import click
import json

@click.command()
@click.option('--format', type=click.Choice(['json', 'text']), default='text')
def show_data(format):
    """Display data in different formats."""
    data = {'status': 'ok', 'count': 42}

    if format == 'json':
        click.echo(json.dumps(data, indent=2))
    else:
        click.echo(f"Status: {data['status']}, Count: {data['count']}")
```

**Colored Output**:
```python
import click

@click.command()
def status():
    """Show status with colors."""
    click.secho('✓ Success', fg='green', bold=True)
    click.secho('⚠ Warning', fg='yellow')
    click.secho('✗ Error', fg='red', bold=True)
```

---

### Command Aliases

```python
import click

class AliasedGroup(click.Group):
    """Support command aliases."""
    def get_command(self, ctx, cmd_name):
        rv = click.Group.get_command(self, ctx, cmd_name)
        if rv is not None:
            return rv
        # Try aliases
        matches = [x for x in self.list_commands(ctx) if x.startswith(cmd_name)]
        if not matches:
            return None
        elif len(matches) == 1:
            return click.Group.get_command(self, ctx, matches[0])
        ctx.fail(f'Ambiguous command: {cmd_name}')

@click.group(cls=AliasedGroup)
def cli():
    """CLI with command aliases."""
    pass

@cli.command()
def status():
    """Show status (alias: st)."""
    click.echo('Status: OK')
```

---

### Subcommand Groups

```python
import click

@click.group()
def cli():
    """Main CLI."""
    pass

@cli.group()
def db():
    """Database operations."""
    pass

@db.command()
def init():
    """Initialize database."""
    click.echo('Database initialized')

@db.command()
def backup():
    """Backup database."""
    click.echo('Database backed up')

# Usage: cli db init, cli db backup
```

---

## Error Handling and Validation

### Custom Validation

```python
import click

class EmailType(click.ParamType):
    """Custom email validation."""
    name = 'email'

    def convert(self, value, param, ctx):
        if '@' not in value:
            self.fail(f'{value!r} is not a valid email address', param, ctx)
        return value

EMAIL = EmailType()

@click.command()
@click.option('--email', type=EMAIL, required=True)
def register(email):
    """Register with email."""
    click.echo(f'Registered: {email}')
```

### Exception Handling

```python
import click
import sys

@click.command()
def risky_operation():
    """Operation that might fail."""
    try:
        # Risky code here
        result = perform_operation()
        click.echo(f'Success: {result}')
    except ValueError as e:
        click.secho(f'Error: {e}', fg='red', err=True)
        sys.exit(1)
    except Exception as e:
        click.secho(f'Unexpected error: {e}', fg='red', err=True)
        sys.exit(2)
```

---

## Performance and Progress Tracking

### Progress Bars with Rich

```python
from rich.progress import Progress, SpinnerColumn, BarColumn, TextColumn
import time

def process_items(items):
    """Process items with progress bar."""
    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        BarColumn(),
        TextColumn("[progress.percentage]{task.percentage:>3.0f}%"),
    ) as progress:
        task = progress.add_task("[cyan]Processing...", total=len(items))

        for item in items:
            # Process item
            time.sleep(0.1)
            progress.update(task, advance=1)
```

### Spinner for Long Operations

```python
from rich.console import Console
from rich.spinner import Spinner
import time

console = Console()

def long_operation():
    """Long-running operation with spinner."""
    with console.status("[bold green]Working...") as status:
        time.sleep(3)
        status.update("[bold blue]Almost done...")
        time.sleep(2)
    console.print("[bold green]Done!")
```

---

## Logging Integration

### Structured Logging

```python
import click
import logging
from pathlib import Path

def setup_logging(verbose: bool, log_file: Path = None):
    """Configure logging."""
    level = logging.DEBUG if verbose else logging.INFO
    handlers = [logging.StreamHandler()]

    if log_file:
        handlers.append(logging.FileHandler(log_file))

    logging.basicConfig(
        level=level,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        handlers=handlers
    )

@click.command()
@click.option('--verbose', is_flag=True, help='Enable debug logging')
@click.option('--log-file', type=click.Path(), help='Log to file')
def main(verbose, log_file):
    """CLI with logging."""
    setup_logging(verbose, log_file)
    logger = logging.getLogger(__name__)

    logger.info('Starting application')
    logger.debug('Debug information')
```

---

## Cross-Platform Considerations

### Path Handling

```python
from pathlib import Path
import click

@click.command()
@click.argument('path', type=click.Path(path_type=Path))
def process_file(path: Path):
    """Process file with cross-platform paths."""
    if not path.exists():
        click.echo(f'Error: {path} does not exist', err=True)
        return

    # Path methods work across platforms
    click.echo(f'Processing: {path.absolute()}')
    click.echo(f'Filename: {path.name}')
    click.echo(f'Extension: {path.suffix}')
```

### Platform-Specific Behavior

```python
import platform
import click

@click.command()
def info():
    """Show platform information."""
    system = platform.system()

    if system == 'Windows':
        click.echo('Running on Windows')
        # Windows-specific code
    elif system == 'Darwin':
        click.echo('Running on macOS')
        # macOS-specific code
    elif system == 'Linux':
        click.echo('Running on Linux')
        # Linux-specific code
```

---

## Shell Integration

### Auto-Completion Setup

**For Bash**:
```bash
# Add to ~/.bashrc
eval "$(_MYCLI_COMPLETE=bash_source mycli)"
```

**For Zsh**:
```bash
# Add to ~/.zshrc
eval "$(_MYCLI_COMPLETE=zsh_source mycli)"
```

**For Fish**:
```bash
# Add to ~/.config/fish/completions/mycli.fish
_MYCLI_COMPLETE=fish_source mycli | source
```

### Click Auto-Completion

```python
import click

@click.group()
def cli():
    """CLI with auto-completion."""
    pass

@cli.command()
@click.argument('env', type=click.Choice(['dev', 'staging', 'prod']))
def deploy(env):
    """Deploy to ENV (supports tab completion)."""
    click.echo(f'Deploying to {env}')
```

---

## Plugin Architecture

### Dynamic Command Loading

```python
import click
import importlib
import pkgutil

class PluginGroup(click.Group):
    """Load plugins dynamically."""

    def list_commands(self, ctx):
        """Discover plugin commands."""
        rv = []
        # Load built-in commands
        for name in super().list_commands(ctx):
            rv.append(name)
        # Load plugin commands
        for _, name, _ in pkgutil.iter_modules(['plugins']):
            rv.append(name)
        return sorted(rv)

    def get_command(self, ctx, name):
        """Load command from plugin."""
        try:
            mod = importlib.import_module(f'plugins.{name}')
            return mod.cli
        except ImportError:
            return super().get_command(ctx, name)

@click.group(cls=PluginGroup)
def cli():
    """Extensible CLI with plugins."""
    pass
```

---

## Configuration Management

### YAML Configuration

```python
import click
import yaml
from pathlib import Path

class Config:
    """Application configuration."""

    def __init__(self, config_file: Path):
        with config_file.open() as f:
            self.data = yaml.safe_load(f)

    def get(self, key: str, default=None):
        """Get configuration value."""
        keys = key.split('.')
        value = self.data
        for k in keys:
            if isinstance(value, dict):
                value = value.get(k)
            else:
                return default
        return value if value is not None else default

@click.group()
@click.option('--config', type=click.Path(exists=True), default='config.yaml')
@click.pass_context
def cli(ctx, config):
    """CLI with YAML configuration."""
    ctx.ensure_object(dict)
    ctx.obj['config'] = Config(Path(config))

@cli.command()
@click.pass_context
def show(ctx):
    """Show configuration."""
    config = ctx.obj['config']
    click.echo(f"API URL: {config.get('api.url')}")
    click.echo(f"Timeout: {config.get('api.timeout', 30)}")
```

---

## Secrets Management

### Environment Variables

```python
import click
import os

@click.command()
def deploy():
    """Deploy with secrets from environment."""
    api_key = os.getenv('API_KEY')
    if not api_key:
        click.secho('Error: API_KEY environment variable not set', fg='red')
        raise click.Abort()

    # Mask secrets in output
    click.echo(f'Using API key: {api_key[:4]}****')
```

### Keyring Integration

```python
import click
import keyring

@click.command()
@click.option('--username', prompt=True)
def login(username):
    """Login and store credentials securely."""
    password = click.prompt('Password', hide_input=True)

    # Store in system keyring
    keyring.set_password('myapp', username, password)
    click.echo('Credentials stored securely')

@click.command()
@click.option('--username', prompt=True)
def use_credentials(username):
    """Use stored credentials."""
    password = keyring.get_password('myapp', username)
    if password:
        click.echo(f'Retrieved password for {username}')
    else:
        click.echo('No credentials found')
```

---

## Testing Strategies

### Mocking External Dependencies

```python
# test_cli.py
import pytest
from click.testing import CliRunner
from unittest.mock import patch, MagicMock

@patch('mycli.api.fetch_data')
def test_fetch_command(mock_fetch):
    """Test fetch command with mocked API."""
    mock_fetch.return_value = {'status': 'ok', 'data': [1, 2, 3]}

    runner = CliRunner()
    result = runner.invoke(cli, ['fetch', '--url', 'https://api.example.com'])

    assert result.exit_code == 0
    assert 'ok' in result.output
    mock_fetch.assert_called_once()
```

### Testing with Temporary Files

```python
from click.testing import CliRunner
import tempfile
from pathlib import Path

def test_file_processing():
    """Test file processing with temporary files."""
    runner = CliRunner()

    with runner.isolated_filesystem():
        # Create test file
        Path('input.txt').write_text('test data')

        # Run command
        result = runner.invoke(cli, ['process', 'input.txt'])

        assert result.exit_code == 0
        assert Path('output.txt').exists()
```

---

## Internationalization (i18n)

### Multi-Language Support

```python
import click
import gettext
import os

# Setup translation
locale_dir = os.path.join(os.path.dirname(__file__), 'locales')
lang = os.getenv('LANG', 'en_US').split('.')[0]

try:
    translation = gettext.translation('messages', locale_dir, languages=[lang])
    _ = translation.gettext
except FileNotFoundError:
    _ = lambda s: s  # Fallback to English

@click.command()
def hello():
    """Greet user in their language."""
    click.echo(_('Hello, world!'))
    click.echo(_('Welcome to the application'))
```

---

## Performance Optimization

### Lazy Loading

```python
import click

@click.group()
def cli():
    """CLI with lazy loading."""
    pass

@cli.command()
def heavy():
    """Heavy command with lazy imports."""
    # Import only when command is used
    import pandas as pd
    import numpy as np

    click.echo('Processing large dataset...')
```

### Parallel Processing

```python
import click
from concurrent.futures import ThreadPoolExecutor, as_completed

@click.command()
@click.argument('files', nargs=-1, type=click.Path(exists=True))
@click.option('--workers', default=4, help='Number of parallel workers')
def process_files(files, workers):
    """Process files in parallel."""
    def process_file(filepath):
        # Processing logic
        return f'Processed: {filepath}'

    with ThreadPoolExecutor(max_workers=workers) as executor:
        futures = [executor.submit(process_file, f) for f in files]

        for future in as_completed(futures):
            click.echo(future.result())
```

---

## Security Best Practices

### Input Validation

```python
import click
import re

def validate_email(ctx, param, value):
    """Validate email format."""
    pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
    if not re.match(pattern, value):
        raise click.BadParameter('Invalid email format')
    return value

@click.command()
@click.option('--email', callback=validate_email, required=True)
def register(email):
    """Register user with validated email."""
    click.echo(f'Registering: {email}')
```

### Safe File Operations

```python
import click
from pathlib import Path

@click.command()
@click.argument('filepath', type=click.Path())
def read_file(filepath):
    """Safely read file."""
    path = Path(filepath).resolve()

    # Prevent path traversal
    base_dir = Path.cwd()
    if not path.is_relative_to(base_dir):
        click.secho('Error: Access denied', fg='red')
        raise click.Abort()

    if not path.exists():
        click.secho(f'Error: {filepath} not found', fg='red')
        raise click.Abort()

    with path.open() as f:
        click.echo(f.read())
```

---

## Debugging and Profiling

### Debug Mode

```python
import click
import logging

@click.command()
@click.option('--debug', is_flag=True, help='Enable debug mode')
def main(debug):
    """CLI with debug mode."""
    if debug:
        logging.basicConfig(level=logging.DEBUG)
        click.echo('Debug mode enabled')

    # Application logic
    logger = logging.getLogger(__name__)
    logger.debug('This is a debug message')
    logger.info('This is an info message')
```

### Performance Profiling

```python
import click
import cProfile
import pstats
from io import StringIO

@click.command()
@click.option('--profile', is_flag=True, help='Enable profiling')
def heavy_task(profile):
    """Run heavy task with optional profiling."""
    if profile:
        profiler = cProfile.Profile()
        profiler.enable()

    # Heavy computation
    result = sum(i**2 for i in range(1000000))

    if profile:
        profiler.disable()
        s = StringIO()
        ps = pstats.Stats(profiler, stream=s).sort_stats('cumulative')
        ps.print_stats(10)
        click.echo(s.getvalue())

    click.echo(f'Result: {result}')
```

---

## Distribution and Updates

### Version Management

```python
import click

__version__ = '1.2.3'

@click.group()
@click.version_option(version=__version__)
def cli():
    """CLI with version tracking."""
    pass
```

### Auto-Update Check

```python
import click
import requests
from packaging import version

CURRENT_VERSION = '1.2.3'
UPDATE_URL = 'https://api.github.com/repos/user/repo/releases/latest'

def check_for_updates():
    """Check if newer version is available."""
    try:
        response = requests.get(UPDATE_URL, timeout=3)
        latest = response.json()['tag_name'].lstrip('v')

        if version.parse(latest) > version.parse(CURRENT_VERSION):
            click.secho(
                f'⚠ Update available: {latest} (current: {CURRENT_VERSION})',
                fg='yellow'
            )
            click.echo(f'Run: pip install --upgrade mycli')
    except Exception:
        pass  # Silently fail if check fails

@click.group(invoke_without_command=True)
@click.pass_context
def cli(ctx):
    """CLI with update checking."""
    if ctx.invoked_subcommand is None:
        check_for_updates()
```

---

**Last Updated**: 2025-10-22
**Framework Versions**: Click 8.1.7, Typer 0.15.0, Rich 13.9.0
**Python Requirement**: ≥3.10

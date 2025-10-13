---
id: CLI-001
version: 0.0.1
status: draft
created: 2025-10-13
updated: 2025-10-13
author: @Goos
priority: high
category: feature
labels:
  - cli
  - click
  - terminal
depends_on:
  - PY314-001
scope:
  packages:
    - moai-adk-py/src/moai_adk/cli/
  files:
    - __main__.py
    - cli/main.py
    - cli/commands/
---

# @SPEC:CLI-001: Click 기반 CLI 시스템

## HISTORY

### v0.0.1 (2025-10-13)
- **INITIAL**: Click 프레임워크 기반 CLI 시스템 명세 작성
- **AUTHOR**: @Goos
- **REASON**: TypeScript commander를 Python click으로 전환

---

## 개요

Click 프레임워크를 사용하여 moai 명령어를 구현한다. 4개 핵심 명령어(init, doctor, status, restore)를 제공하고, Rich 라이브러리로 터미널 출력을 개선한다.

---

## Environment (환경 및 전제조건)

### 기술 스택
- **CLI 프레임워크**: Click 8.1+
- **터미널 UI**: Rich 13.0+
- **ASCII Art**: figlet 또는 pyfiglet (선택적)
- **진입점**: moai_adk.__main__:main

### 기존 시스템
- TypeScript commander 기반 CLI
- 명령어: init, doctor, status, update, restore, help, version

---

## Requirements (요구사항)

### Ubiquitous Requirements (기본 요구사항)
- 시스템은 `moai` 명령어를 제공해야 한다
- 시스템은 4개 핵심 명령어를 지원해야 한다
- 시스템은 Rich를 사용한 색상 출력을 제공해야 한다
- 시스템은 ASCII 로고를 표시해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN `moai init .` 명령이 실행되면, 시스템은 프로젝트를 초기화해야 한다
- WHEN `moai doctor` 명령이 실행되면, 시스템은 환경을 검증해야 한다
- WHEN `moai status` 명령이 실행되면, 시스템은 프로젝트 상태를 표시해야 한다
- WHEN `moai restore` 명령이 실행되면, 시스템은 백업을 복원해야 한다
- WHEN `moai --version` 명령이 실행되면, 시스템은 버전을 출력해야 한다
- WHEN `moai --help` 명령이 실행되면, 시스템은 도움말을 표시해야 한다

### State-driven Requirements (상태 기반)
- WHILE 명령 실행 중일 때, 시스템은 진행 상황을 표시해야 한다
- WHILE 오류 발생 시, 시스템은 명확한 에러 메시지를 표시해야 한다

### Constraints (제약사항)
- 명령어는 `moai <command>` 형식이어야 한다
- 모든 출력은 Rich를 사용해야 한다
- 로고는 ASCII 아트 형식이어야 한다
- 명령어 실행 시간은 3초 이내여야 한다 (doctor 제외)

---

## Specifications (상세 명세)

### 1. CLI 구조

```python
# moai_adk/__main__.py
import click
from moai_adk.cli.main import cli

def main():
    cli()

if __name__ == "__main__":
    main()
```

```python
# moai_adk/cli/main.py
import click
from rich.console import Console

console = Console()

@click.group()
@click.version_option(version="0.3.0")
def cli():
    """MoAI Agentic Development Kit - SPEC-First TDD Framework"""
    show_logo()

def show_logo():
    logo = """
    ▶◀ MoAI-ADK v0.3.0
    """
    console.print(logo, style="bold cyan")
```

### 2. 명령어 구현

#### 2.1. moai init
```python
@cli.command()
@click.argument('path', type=click.Path(), default='.')
def init(path):
    """Initialize a new MoAI-ADK project"""
    from moai_adk.core.project import initialize_project

    console.print(f"[cyan]Initializing project at {path}...[/cyan]")
    initialize_project(path)
    console.print("[green]✓ Project initialized successfully[/green]")
```

#### 2.2. moai doctor
```python
@cli.command()
def doctor():
    """Check system requirements and project health"""
    from moai_adk.core.project import check_environment

    console.print("[cyan]Running system diagnostics...[/cyan]")
    results = check_environment()

    for check, status in results.items():
        icon = "✓" if status else "✗"
        color = "green" if status else "red"
        console.print(f"[{color}]{icon} {check}[/{color}]")
```

#### 2.3. moai status
```python
@cli.command()
def status():
    """Show current project status"""
    from moai_adk.core.project import get_project_status

    status_data = get_project_status()

    console.print("[bold]Project Status:[/bold]")
    console.print(f"  Mode: {status_data['mode']}")
    console.print(f"  Locale: {status_data['locale']}")
    console.print(f"  SPECs: {status_data['spec_count']}")
```

#### 2.4. moai restore
```python
@cli.command()
@click.option('--timestamp', help='Specific backup timestamp')
def restore(timestamp):
    """Restore from backup"""
    from moai_adk.core.backup import restore_backup

    if timestamp:
        console.print(f"[cyan]Restoring from {timestamp}...[/cyan]")
    else:
        console.print("[cyan]Restoring from latest backup...[/cyan]")

    restore_backup(timestamp)
    console.print("[green]✓ Restore completed[/green]")
```

### 3. ASCII 로고

```
▶◀ MoAI-ADK v0.3.0
───────────────────────────────────
SPEC-First TDD Development Framework
```

### 4. Rich 출력 스타일

- **성공**: `[green]✓ Message[/green]`
- **실패**: `[red]✗ Error[/red]`
- **정보**: `[cyan]ℹ Info[/cyan]`
- **경고**: `[yellow]⚠ Warning[/yellow]`

### 5. 에러 처리

```python
@cli.command()
def example():
    try:
        # Command logic
        pass
    except FileNotFoundError as e:
        console.print(f"[red]✗ File not found: {e}[/red]")
        raise click.Abort()
    except Exception as e:
        console.print(f"[red]✗ Unexpected error: {e}[/red]")
        raise
```

---

## Traceability (추적성)

- **SPEC ID**: @SPEC:CLI-001
- **Depends on**: PY314-001
- **TAG 체인**: @SPEC:CLI-001 → @TEST:CLI-001 → @CODE:CLI-001

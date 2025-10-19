---
name: moai-domain-cli-tool
description: CLI tool development with argument parsing, POSIX compliance, and user-friendly help messages
allowed-tools:
  - Read
  - Bash
tier: 4
auto-load: "false"
---

# CLI Tool Expert

## What it does

Provides expertise in developing command-line interface tools with proper argument parsing, POSIX compliance, intuitive help messages, and standard exit codes.

## When to use

- "CLI 도구 개발", "명령줄 파싱", "POSIX 호환성"
- Automatically invoked when working with CLI projects
- CLI tool SPEC implementation (`/alfred:2-run`)

## How it works

**Argument Parsing**:
- **Python**: argparse, click, typer
- **Node.js**: commander, yargs, oclif
- **Rust**: clap, structopt
- **Go**: cobra, flag
- **Subcommands**: git-style commands (tool add, tool remove)

**POSIX Compliance**:
- **Short options**: -h, -v
- **Long options**: --help, --version
- **Option arguments**: -o file, --output=file
- **Standard streams**: stdin, stdout, stderr
- **Exit codes**: 0 (success), 1-255 (errors)

**User Experience**:
- **Help messages**: Comprehensive usage documentation
- **Auto-completion**: Shell completion (bash, zsh, fish)
- **Progress indicators**: Spinners, progress bars
- **Color output**: ANSI colors for readability
- **Interactive prompts**: Confirmation dialogs

**Configuration**:
- **Config files**: YAML, JSON, TOML (e.g., ~/.toolrc)
- **Environment variables**: Fallback configuration
- **Precedence**: CLI args > env vars > config file > defaults

## Examples

### Example 1: Python CLI with Click Framework

```python
# @CODE:CLI-001 | cli.py
import click

@click.group()
@click.version_option()
def cli():
    """MoAI CLI Tool - Powerful development automation"""
    pass

@cli.command()
@click.argument('project_name')
@click.option('--template', default='full', help='Project template')
def init(project_name, template):
    """Initialize a new project"""
    click.echo(f"Initializing {project_name} with {template} template...")
    # 구현

@cli.command()
@click.option('--coverage', default=85, help='Minimum coverage %')
def test(coverage):
    """Run tests with coverage"""
    click.echo(f"Running tests with {coverage}% coverage...")
    # 구현

@cli.command()
@click.argument('branch_name')
@click.option('--create', is_flag=True, help='Create if not exists')
def checkout(branch_name, create):
    """Checkout a branch"""
    click.echo(f"Checking out {branch_name}...")
    # 구현

if __name__ == '__main__':
    cli()

# 사용법:
# python cli.py init my-project --template=web
# python cli.py test --coverage=90
# python cli.py checkout feature/auth --create
```

### Example 2: TypeScript CLI with Commander

```typescript
// @CODE:CLI-002 | cli.ts
import { Command } from 'commander';

const program = new Command();

program
    .name('moai')
    .description('MoAI Development Kit')
    .version('1.0.0');

program
    .command('build <spec>')
    .description('Build a SPEC implementation')
    .option('--tdd', 'Use TDD workflow')
    .action((spec, options) => {
        console.log(`Building ${spec}...`);
        // 구현
    });

program
    .command('sync')
    .description('Sync documentation')
    .option('--mode <mode>', 'Sync mode: auto|force')
    .action((options) => {
        console.log(`Syncing with mode: ${options.mode}...`);
        // 구현
    });

program.parse(process.argv);
```

### Example 3: POSIX Compliance

**Checklist**:
```bash
# @CODE:POSIX-001: POSIX 호환성

✅ 짧은 옵션:  -h, -v, -o file
✅ 긴 옵션:    --help, --version, --output=file
✅ 옵션 인자:  -o <value>, --output <value>, --output=<value>
✅ 옵션 순서:  command [options] [arguments]
✅ 종료 코드:  0 (성공), 1-255 (오류)
✅ 표준 스트림:
   - stdin:  입력 받기
   - stdout: 정상 출력
   - stderr: 에러 메시지
✅ 도움말:     command -h, command --help
✅ 버전:       command -v, command --version
```

### Example 4: Configuration & Shell Completion

**Config File**:
```yaml
# @CODE:CONFIG-001: ~/.moai/config.yaml
project:
  locale: ko
  template: full

build:
  test_coverage: 85
  strict_mode: true

deployment:
  auto_sync: true
  env: production
```

**Shell Completion**:
```bash
# @CODE:COMPLETION-001: Bash completion

_moai_completion() {
    local cur prev commands

    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    commands="init build sync test deploy help"

    if [[ $COMP_CWORD -eq 1 ]]; then
        COMPREPLY=($(compgen -W "${commands}" -- ${cur}))
    elif [[ "${prev}" == "build" ]]; then
        # 등록된 SPEC 목록 자동완성
        COMPREPLY=($(compgen -W "SPEC-001 SPEC-002" -- ${cur}))
    fi
}

complete -F _moai_completion moai

# 사용:
# $ moai [TAB]        → 명령 자동완성
# $ moai build [TAB]  → SPEC 자동완성
```

## Keywords

"CLI 도구", "명령줄 파싱", "POSIX", "셸 호환성", "자동완성", "설정 파일", "서브커맨드", "옵션", "인자", "user experience", "command-line interface"

## Reference

- CLI best practices: `.moai/memory/development-guide.md#CLI-설계`
- Argument parsing: CLAUDE.md#인자-파싱-패턴
- POSIX compliance: `.moai/memory/development-guide.md#POSIX-표준`

## Works well with

- moai-foundation-trust (CLI 테스트)
- moai-domain-backend (서버 CLI)
- moai-domain-devops (배포 자동화 CLI)

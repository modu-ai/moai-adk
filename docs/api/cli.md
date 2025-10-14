# CLI API Reference

> MoAI-ADK Python v0.3.0 CLI Module API Documentation

MoAI-ADK의 CLI(Command-Line Interface) 모듈로, 터미널에서 프로젝트 초기화, 시스템 진단, 백업 복원 등을 수행합니다.

---

## 모듈 구조

```
moai_adk.cli/
├── __main__.py        # CLI 진입점 (click.group)
├── main.py            # CLI 그룹 정의
└── commands/          # 개별 명령어 구현 (예정)
    ├── __init__.py
    ├── init.py        # moai init
    ├── doctor.py      # moai doctor
    ├── status.py      # moai status
    └── restore.py     # moai restore
```

---

## CLI 진입점

### moai

::: moai_adk.__main__.cli
    options:
      show_source: true
      heading_level: 4

MoAI-ADK의 루트 CLI 그룹입니다. Click 프레임워크 기반으로 구현되었습니다.

#### 기본 동작

인자 없이 `moai` 실행 시 로고와 도움말 안내가 출력됩니다:

```bash
$ moai
▶◀ MoAI-ADK v0.3.0
SPEC-First TDD Framework with Alfred SuperAgent

Tip: Run moai --help to see available commands
```

#### 전역 옵션

| 옵션 | 설명 | 예시 |
|------|------|------|
| `--version` | 버전 정보 출력 | `moai --version` |
| `--help` | 도움말 출력 | `moai --help` |

#### 사용 예시: 버전 확인

```bash
$ moai --version
MoAI-ADK, version 0.3.0
```

#### 사용 예시: 도움말

```bash
$ moai --help
Usage: moai [OPTIONS] COMMAND [ARGS]...

  MoAI Agentic Development Kit

  SPEC-First TDD Framework with Alfred SuperAgent

Options:
  --version  Show the version and exit.
  --help     Show this message and exit.

Commands:
  init     Initialize a MoAI-ADK project
  doctor   Check system requirements
  status   Show project status
  restore  Restore from backup
```

---

## CLI 명령어

### moai init

프로젝트를 초기화하고 `.moai/` 디렉토리 구조를 생성합니다.

#### 명령어 형식

```bash
moai init [PATH] [OPTIONS]
```

#### 인자 (Arguments)

| 인자 | 타입 | 기본값 | 설명 |
|------|------|--------|------|
| `PATH` | string | `.` | 초기화할 프로젝트 경로 |

#### 옵션 (Options)

| 옵션 | 타입 | 기본값 | 설명 |
|------|------|--------|------|
| `--mode` | choice | `personal` | 프로젝트 모드 (`personal`, `team`) |
| `--locale` | choice | `ko` | 로케일 (`ko`, `en`, `ja`, `zh`) |
| `--language` | string | (자동 감지) | 강제로 지정할 언어 |

#### 사용 예시: 현재 디렉토리 초기화

```bash
# 기본 설정으로 초기화
$ moai init .
Initializing MoAI-ADK project...
✓ Language detected: python
✓ Created .moai/ directory
✓ Generated config.json
Project initialized successfully!

Next steps:
  1. /alfred:0-project  # 프로젝트 문서 작성
  2. /alfred:1-spec     # SPEC 작성
```

#### 사용 예시: Team 모드로 초기화

```bash
$ moai init . --mode team --locale en
Initializing MoAI-ADK project...
✓ Language detected: typescript
✓ Mode: team
✓ Locale: en
✓ Created .moai/ directory
✓ Generated config.json
Project initialized successfully!
```

#### 사용 예시: 언어 강제 지정

```bash
# TypeScript 프로젝트이지만 Go 템플릿 사용
$ moai init . --language go
Initializing MoAI-ADK project...
✓ Language (manual): go
✓ Created .moai/ directory
✓ Generated config.json
Project initialized successfully!
```

#### 사용 예시: 특정 경로 초기화

```bash
$ moai init /path/to/new/project --mode team
Initializing MoAI-ADK project at /path/to/new/project...
✓ Language detected: python
✓ Mode: team
✓ Created .moai/ directory
Project initialized successfully!
```

#### Python API 사용

```python
from click.testing import CliRunner
from moai_adk.__main__ import cli

runner = CliRunner()

# init 명령어 실행
result = runner.invoke(cli, ["init", ".", "--mode", "team"])
assert result.exit_code == 0
print(result.output)
```

#### 생성되는 파일 구조

```
.moai/
├── config.json              # 프로젝트 설정
├── project/
│   ├── product.md          # 제품 개요 (템플릿)
│   ├── structure.md        # 디렉토리 구조 (템플릿)
│   └── tech.md             # 기술 스택 (템플릿)
├── specs/                  # SPEC 문서
├── memory/                 # 개발 가이드
└── backup/                 # 백업 파일
```

#### config.json 예시

```json
{
  "projectName": "my-project",
  "mode": "team",
  "locale": "ko",
  "language": "python"
}
```

---

### moai doctor

시스템 요구사항을 검증하고 필수/선택 도구의 설치 여부를 확인합니다.

#### 명령어 형식

```bash
moai doctor [OPTIONS]
```

#### 옵션 (Options)

| 옵션 | 타입 | 기본값 | 설명 |
|------|------|--------|------|
| `--verbose` | flag | `false` | 상세 정보 출력 |

#### 사용 예시: 기본 진단

```bash
$ moai doctor
Running system diagnostics...

Required Tools:
  ✓ git         v2.39.0
  ✓ python      v3.11.0

Optional Tools:
  ✓ gh          v2.40.0
  ✗ docker      Not installed

✅ System ready for MoAI-ADK
```

#### 사용 예시: 상세 모드

```bash
$ moai doctor --verbose
Running system diagnostics...

Required Tools:
  ✓ git         v2.39.0
    Path: /usr/bin/git
    Check: git --version

  ✓ python      v3.11.0
    Path: /usr/local/bin/python3
    Check: python3 --version

Optional Tools:
  ✓ gh          v2.40.0
    Path: /usr/local/bin/gh
    Check: gh --version
    Purpose: Automated PR creation

  ✗ docker      Not installed
    Check: docker --version
    Purpose: Container-based testing
    Install: https://www.docker.com/get-started

✅ System ready for MoAI-ADK
```

#### 사용 예시: 도구 미설치 시

```bash
$ moai doctor
Running system diagnostics...

Required Tools:
  ✗ git         Not installed
  ✓ python      v3.11.0

Optional Tools:
  ✗ gh          Not installed
  ✗ docker      Not installed

❌ Required tools are missing!

Please install:
  - Git: https://git-scm.com/downloads

After installation, run 'moai doctor' again.
```

#### Python API 사용

```python
from click.testing import CliRunner
from moai_adk.__main__ import cli

runner = CliRunner()

# doctor 명령어 실행
result = runner.invoke(cli, ["doctor"])
assert result.exit_code == 0

# 필수 도구 확인
if "✓ git" in result.output and "✓ python" in result.output:
    print("시스템 준비 완료")
else:
    print("필수 도구 미설치")
```

#### 진단 항목

**필수 도구 (Required)**:
- `git`: Git 버전 관리 시스템
- `python`: Python 3.9 이상

**선택 도구 (Optional)**:
- `gh`: GitHub CLI (PR 자동화, Draft PR 생성)
- `docker`: Docker (컨테이너 환경 테스트)

---

### moai status

현재 프로젝트의 상태 정보를 출력합니다.

#### 명령어 형식

```bash
moai status [OPTIONS]
```

#### 옵션 (Options)

| 옵션 | 타입 | 기본값 | 설명 |
|------|------|--------|------|
| `--json` | flag | `false` | JSON 형식으로 출력 |

#### 사용 예시: 기본 출력

```bash
$ moai status
Project Status

Configuration:
  Project Name: my-awesome-app
  Mode:         team
  Locale:       ko
  Language:     python

Git Information:
  Repository:   Yes
  Current Branch: feature/SPEC-AUTH-001
  Status:       Clean (no uncommitted changes)

Directory Structure:
  ✓ .moai/config.json
  ✓ .moai/project/product.md
  ✓ .moai/project/structure.md
  ✓ .moai/project/tech.md
  ✓ .moai/specs/ (3 specs)
  ✓ .moai/memory/ (2 guides)
  ✓ .moai/backup/ (5 backups)
```

#### 사용 예시: JSON 출력

```bash
$ moai status --json
{
  "projectName": "my-awesome-app",
  "mode": "team",
  "locale": "ko",
  "language": "python",
  "git": {
    "isRepo": true,
    "currentBranch": "feature/SPEC-AUTH-001",
    "isDirty": false
  },
  "structure": {
    "configExists": true,
    "specsCount": 3,
    "memoryCount": 2,
    "backupCount": 5
  }
}
```

#### 사용 예시: 초기화되지 않은 프로젝트

```bash
$ moai status
Project Status

⚠️ This directory is not initialized as a MoAI-ADK project.

Run 'moai init .' to get started.
```

#### Python API 사용

```python
from click.testing import CliRunner
from moai_adk.__main__ import cli
import json

runner = CliRunner()

# JSON 형식으로 상태 조회
result = runner.invoke(cli, ["status", "--json"])
assert result.exit_code == 0

status = json.loads(result.output)
print(f"프로젝트명: {status['projectName']}")
print(f"현재 브랜치: {status['git']['currentBranch']}")
print(f"SPEC 개수: {status['structure']['specsCount']}")
```

---

### moai restore

백업으로부터 프로젝트를 복원합니다.

#### 명령어 형식

```bash
moai restore [OPTIONS]
```

#### 옵션 (Options)

| 옵션 | 타입 | 기본값 | 설명 |
|------|------|--------|------|
| `--timestamp` | string | (최신) | 복원할 백업 시점 (YYYY-MM-DD-HHMMSS) |
| `--list` | flag | `false` | 사용 가능한 백업 목록 출력 |
| `--dry-run` | flag | `false` | 실제 복원 없이 미리보기 |

#### 사용 예시: 최신 백업 복원

```bash
$ moai restore
Restoring from latest backup...

Found backup: 2025-10-14-153000
  Created: 2025-10-14 15:30:00
  Files: 12

✓ Restored .moai/config.json
✓ Restored .moai/project/product.md
✓ Restored .moai/project/structure.md
✓ Restored .moai/specs/SPEC-AUTH-001/spec.md
  ... (8 more files)

✅ Restore completed successfully!
```

#### 사용 예시: 특정 시점 복원

```bash
$ moai restore --timestamp 2025-10-14-120000
Restoring from backup: 2025-10-14-120000...

Found backup: 2025-10-14-120000
  Created: 2025-10-14 12:00:00
  Files: 10

✓ Restored .moai/config.json
✓ Restored .moai/project/product.md
  ... (8 more files)

✅ Restore completed successfully!
```

#### 사용 예시: 백업 목록 조회

```bash
$ moai restore --list
Available backups:

  1. 2025-10-14-153000 (Latest)
     Created: 2025-10-14 15:30:00
     Files: 12
     Size: 45 KB

  2. 2025-10-14-120000
     Created: 2025-10-14 12:00:00
     Files: 10
     Size: 38 KB

  3. 2025-10-13-180000
     Created: 2025-10-13 18:00:00
     Files: 8
     Size: 32 KB

Use 'moai restore --timestamp <TIMESTAMP>' to restore a specific backup.
```

#### 사용 예시: Dry-run (미리보기)

```bash
$ moai restore --timestamp 2025-10-14-120000 --dry-run
[DRY-RUN] Restoring from backup: 2025-10-14-120000...

Found backup: 2025-10-14-120000
  Created: 2025-10-14 12:00:00
  Files: 10

Would restore:
  - .moai/config.json
  - .moai/project/product.md
  - .moai/project/structure.md
  - .moai/specs/SPEC-AUTH-001/spec.md
  ... (6 more files)

[DRY-RUN] No files were actually restored.
Run without --dry-run to perform the restore.
```

#### Python API 사용

```python
from click.testing import CliRunner
from moai_adk.__main__ import cli

runner = CliRunner()

# 최신 백업 복원
result = runner.invoke(cli, ["restore"])
assert result.exit_code == 0
assert "completed successfully" in result.output

# 특정 시점 복원
result = runner.invoke(cli, ["restore", "--timestamp", "2025-10-14-120000"])
assert result.exit_code == 0

# 백업 목록 조회
result = runner.invoke(cli, ["restore", "--list"])
assert result.exit_code == 0
assert "Available backups" in result.output
```

---

## CLI 헬퍼 함수

### show_logo

::: moai_adk.__main__.show_logo
    options:
      show_source: true
      heading_level: 4

MoAI-ADK ASCII 로고를 Rich Console로 출력하는 헬퍼 함수입니다.

#### 사용 예시

```python
from moai_adk.__main__ import show_logo

# 로고 출력
show_logo()

# 출력:
# ▶◀ MoAI-ADK v0.3.0
# SPEC-First TDD Framework with Alfred SuperAgent
```

#### Rich Console 스타일

- 로고: `[cyan]` (밝은 청록색)
- 부제: `[dim]` (흐린 회색)

---

### main

::: moai_adk.__main__.main
    options:
      show_source: true
      heading_level: 4

CLI의 메인 진입점으로, 예외 처리 및 종료 코드를 관리합니다.

#### 반환 값

| 반환 값 | 의미 |
|---------|------|
| `0` | 성공 |
| `1` | 에러 발생 |

#### 사용 예시

```python
import sys
from moai_adk.__main__ import main

# CLI 실행
exit_code = main()
sys.exit(exit_code)
```

#### 에러 처리

예외 발생 시 Rich Console로 에러 메시지를 출력하고 종료 코드 1을 반환합니다:

```python
# 예외 발생 시
try:
    cli()
    return 0
except Exception as e:
    console.print(f"[red]Error:[/red] {e}")
    return 1
```

---

## Click Testing

### CliRunner 사용

MoAI-ADK CLI는 Click의 `CliRunner`로 테스트할 수 있습니다.

#### Fixture 설정 (pytest)

```python
# conftest.py
import pytest
from click.testing import CliRunner

@pytest.fixture
def cli_runner():
    """Click CLI 테스트용 러너"""
    return CliRunner()
```

#### 테스트 예시

```python
from click.testing import CliRunner
from moai_adk.__main__ import cli

def test_version_command():
    """--version 옵션 테스트"""
    runner = CliRunner()
    result = runner.invoke(cli, ["--version"])

    assert result.exit_code == 0
    assert "0.3.0" in result.output

def test_init_command():
    """init 명령어 테스트"""
    runner = CliRunner()

    with runner.isolated_filesystem():
        # 임시 디렉토리에서 실행
        result = runner.invoke(cli, ["init", "."])

        assert result.exit_code == 0
        assert "successfully" in result.output

def test_invalid_command():
    """잘못된 명령어 테스트"""
    runner = CliRunner()
    result = runner.invoke(cli, ["invalid-command"])

    assert result.exit_code != 0
    assert "Error" in result.output or "No such command" in result.output
```

---

## 통합 사용 예시

### 프로젝트 초기화부터 상태 확인까지

```bash
# 1. 시스템 요구사항 확인
$ moai doctor
Running system diagnostics...
✓ git         v2.39.0
✓ python      v3.11.0
✅ System ready for MoAI-ADK

# 2. 프로젝트 초기화 (Team 모드)
$ moai init . --mode team --locale ko
Initializing MoAI-ADK project...
✓ Language detected: python
✓ Mode: team
✓ Locale: ko
✓ Created .moai/ directory
Project initialized successfully!

# 3. 프로젝트 상태 확인
$ moai status
Project Status

Configuration:
  Project Name: my-awesome-app
  Mode:         team
  Locale:       ko
  Language:     python

Git Information:
  Repository:   Yes
  Current Branch: main
  Status:       Clean

Directory Structure:
  ✓ .moai/config.json
  ✓ .moai/project/
  ✓ .moai/specs/
  ✓ .moai/memory/
  ✓ .moai/backup/

# 4. 백업 목록 확인
$ moai restore --list
Available backups:
  1. 2025-10-14-153000 (Latest)
     Created: 2025-10-14 15:30:00
     Files: 12
```

### Python 스크립트에서 CLI 호출

```python
import subprocess
import json

def initialize_project(path: str, mode: str = "personal"):
    """CLI를 사용하여 프로젝트 초기화"""
    result = subprocess.run(
        ["moai", "init", path, "--mode", mode],
        capture_output=True,
        text=True
    )

    if result.returncode == 0:
        print("프로젝트 초기화 성공!")
        print(result.stdout)
    else:
        print("프로젝트 초기화 실패!")
        print(result.stderr)

def get_project_status() -> dict:
    """프로젝트 상태를 JSON으로 조회"""
    result = subprocess.run(
        ["moai", "status", "--json"],
        capture_output=True,
        text=True
    )

    if result.returncode == 0:
        return json.loads(result.stdout)
    else:
        raise RuntimeError(f"Failed to get status: {result.stderr}")

def check_system_requirements() -> bool:
    """시스템 요구사항 확인"""
    result = subprocess.run(
        ["moai", "doctor"],
        capture_output=True,
        text=True
    )

    return result.returncode == 0 and "✅" in result.stdout

# 사용 예시
if __name__ == "__main__":
    # 시스템 확인
    if not check_system_requirements():
        print("필수 도구를 설치하세요.")
        exit(1)

    # 프로젝트 초기화
    initialize_project(".", mode="team")

    # 상태 조회
    status = get_project_status()
    print(f"프로젝트명: {status['projectName']}")
    print(f"모드: {status['mode']}")
```

---

## Rich Console 스타일링

MoAI-ADK CLI는 Rich 라이브러리를 사용하여 아름다운 터미널 출력을 제공합니다.

### 색상 코드

| 태그 | 색상 | 용도 |
|------|------|------|
| `[cyan]` | 밝은 청록색 | 로고, 명령어 이름 |
| `[green]` | 초록색 | 성공 메시지, 체크 마크 |
| `[yellow]` | 노란색 | 경고, Tip |
| `[red]` | 빨간색 | 에러 메시지 |
| `[dim]` | 흐린 회색 | 부가 설명 |
| `[bold]` | 굵은 글씨 | 제목, 강조 |

### 사용 예시

```python
from rich.console import Console

console = Console()

# 성공 메시지
console.print("[green]✓[/green] Project initialized successfully!")

# 경고 메시지
console.print("[yellow]⚠️[/yellow] Docker is not installed (optional)")

# 에러 메시지
console.print("[red]❌[/red] Git is required but not found")

# 정보 메시지
console.print("[cyan]ℹ️[/cyan] Tip: Run 'moai --help' for more information")
```

---

## 성능 최적화

### CLI 로딩 시간 목표

MoAI-ADK CLI는 빠른 시작 시간을 목표로 합니다:

- **목표**: 500ms 이내
- **측정 방법**: `moai --version` 실행 시간

### 로딩 시간 측정

```bash
# Bash
$ time moai --version
MoAI-ADK, version 0.3.0

real    0m0.234s
user    0m0.198s
sys     0m0.032s
```

### Python 테스트

```python
import time
from click.testing import CliRunner
from moai_adk.__main__ import cli

def test_cli_loading_performance():
    """CLI 로딩 시간 < 500ms 검증"""
    runner = CliRunner()

    start_time = time.time()
    result = runner.invoke(cli, ["--version"])
    elapsed_time = (time.time() - start_time) * 1000  # ms

    assert result.exit_code == 0
    assert elapsed_time < 500, f"CLI took {elapsed_time:.2f}ms (> 500ms)"
    print(f"✓ CLI loaded in {elapsed_time:.2f}ms")
```

---

## 참고 문서

- **SPEC 문서**: `.moai/specs/SPEC-CLI-001/spec.md` - CLI 명령어 상세 명세
- **테스트 코드**: `tests/unit/test_cli.py` - CLI 테스트 예시
- **Click 공식 문서**: https://click.palletsprojects.com/
- **Rich 공식 문서**: https://rich.readthedocs.io/

---

**최종 업데이트**: 2025-10-14
**버전**: v0.3.0
**작성자**: @doc-syncer

# MoAI-ADK API Documentation v0.1.26

**SPEC-003 최적화 적용: 간소화된 API 및 성능 향상**

## 📋 개요

MoAI-ADK는 Claude Code와의 완전한 통합을 위한 **3단계 파이프라인** 기반 API를 제공합니다.
SPEC-003 최적화를 통해 API 호출 성능이 **50% 향상**되었으며, 메모리 사용량이 **70% 감소**했습니다.

### 핵심 API 구조

```python
from moai_adk import (
    SimplifiedInstaller,
    Config,
    SecurityManager,
    ConfigManager,
    TemplateEngine,
    CLICommands
)
```

## 🛠️ CLI API Reference

### 1. Project Initialization

#### `moai init [project_path]`

새로운 MoAI-ADK 프로젝트를 초기화합니다.

```bash
# 기본 초기화
moai init

# 특정 디렉터리에 초기화
moai init ./my-project

# 대화형 설정 마법사 실행
moai init --interactive

# 백업 생성과 함께 초기화
moai init --backup

# 조용한 모드 (최소 출력)
moai init --quiet
```

**Parameters:**
- `project_path` (optional): 프로젝트 디렉터리 경로 (기본값: 현재 디렉터리)
- `--template, -t`: 템플릿 선택 (standard, minimal, advanced)
- `--interactive, -i`: 대화형 설정 마법사 실행
- `--backup, -b`: 설치 전 백업 생성
- `--force, -f`: 기존 파일 강제 덮어쓰기 (위험)
- `--force-copy`: 심볼릭 링크 대신 파일 복사 (Windows 권장)
- `--quiet, -q`: 최소 출력 모드

**Returns:**
- 성공: 프로젝트 초기화 완료 메시지
- 실패: 오류 메시지 및 종료 코드 1

### 2. Project Status

#### `moai status`

현재 프로젝트의 MoAI-ADK 상태를 확인합니다.

```bash
# 기본 상태 확인
moai status

# 상세 정보 포함
moai status --verbose

# 특정 프로젝트 경로 지정
moai status --project-path ./other-project
```

**Parameters:**
- `--verbose, -v`: 상세 상태 정보 표시
- `--project-path, -p`: 프로젝트 디렉터리 경로

**Output:**
```
📊 MoAI-ADK Project Status

📂 Project: /path/to/project
   Type: python

🗿 MoAI-ADK Components:
   MoAI System: ✅
   Claude Integration: ✅
   Memory File: ✅
   Git Repository: ✅

🧭 Versions:
   Package: v0.1.26
   Templates: v0.1.26
```

### 3. Health Check

#### `moai doctor`

시스템 상태를 진단하고 일반적인 문제를 확인합니다.

```bash
# 시스템 진단 실행
moai doctor

# 사용 가능한 백업 목록 표시
moai doctor --list-backups
```

**Parameters:**
- `--list-backups, -l`: 사용 가능한 백업 목록 표시

### 4. Update System

#### `moai update`

MoAI-ADK를 최신 버전으로 업데이트합니다.

```bash
# 업데이트 확인만 수행
moai update --check

# 전체 업데이트 (패키지 + 리소스)
moai update

# 백업 없이 업데이트 (비권장)
moai update --no-backup

# 리소스만 업데이트
moai update --resources-only

# 패키지 정보만 확인
moai update --package-only
```

**Parameters:**
- `--check, -c`: 업데이트 확인만 수행
- `--no-backup`: 백업 생성 건너뛰기
- `--verbose, -v`: 상세 업데이트 정보 표시
- `--package-only`: Python 패키지만 업데이트
- `--resources-only`: 프로젝트 리소스만 업데이트

### 5. Backup & Restore

#### `moai restore [backup_path]`

백업에서 MoAI-ADK 설정을 복원합니다.

```bash
# 백업에서 복원
moai restore .moai_backup_20250119_143022

# 드라이런 (실제 복원 없이 확인만)
moai restore .moai_backup_20250119_143022 --dry-run
```

**Parameters:**
- `backup_path`: 백업 디렉터리 경로
- `--dry-run`: 실제 복원 없이 작업 내용만 표시

## 🐍 Python API Reference

### 1. SimplifiedInstaller

프로젝트 설치를 담당하는 핵심 클래스입니다.

```python
from moai_adk import SimplifiedInstaller, Config, RuntimeConfig

# 설정 생성
config = Config(
    project_path="/path/to/project",
    name="my-project",
    template="standard",
    runtime=RuntimeConfig("python"),
    force_overwrite=False
)

# 설치 실행
installer = SimplifiedInstaller(config)
result = installer.install(progress_callback)

if result.success:
    print(f"설치 완료: {result.project_path}")
    print(f"생성된 파일 수: {len(result.files_created)}")
else:
    print(f"설치 실패: {result.errors}")
```

**Methods:**
- `install(progress_callback=None)`: 프로젝트 설치 실행
- `validate_config()`: 설정 유효성 검증

### 2. Config & RuntimeConfig

프로젝트 설정을 관리하는 클래스들입니다.

```python
from moai_adk import Config, RuntimeConfig

# 런타임 설정
runtime = RuntimeConfig(
    language="python",
    version="3.11+",
    dependencies=["click", "colorama", "pyyaml"]
)

# 프로젝트 설정
config = Config(
    project_path="./my-project",
    name="my-project",
    template="standard",
    runtime=runtime,
    force_overwrite=False
)
```

### 3. SecurityManager

보안 관련 기능을 제공합니다.

```python
from moai_adk import SecurityManager

security = SecurityManager()

# 민감한 정보 마스킹
masked_text = security.mask_sensitive_info("api_key=secret123")
# 결과: "api_key=***"

# 파일 권한 검증
is_safe = security.validate_file_permissions("/path/to/file")
```

### 4. TemplateEngine

템플릿 처리를 담당합니다.

```python
from moai_adk import TemplateEngine

engine = TemplateEngine()

# 템플릿 렌더링
rendered = engine.render_template(
    template_content="Hello {{ name }}!",
    context={"name": "MoAI-ADK"}
)
# 결과: "Hello MoAI-ADK!"
```

## 📦 Package Optimization API (SPEC-003)

### PackageOptimizer

SPEC-003에서 도입된 패키지 최적화 기능입니다.

```python
from package_optimization_system.core.package_optimizer import PackageOptimizer

# 최적화 실행
optimizer = PackageOptimizer("/path/to/target/directory")
result = optimizer.optimize()

if result["success"]:
    print(f"최적화 완료!")
    print(f"초기 크기: {result['initial_size']} bytes")
    print(f"최종 크기: {result['final_size']} bytes")
    print(f"감소율: {result['reduction_percentage']:.1f}%")

    metrics = result["metrics"]
    print(f"처리된 파일: {metrics['files_processed']}")
    print(f"제거된 중복: {metrics['duplicates_removed']}")
    print(f"최적화 시간: {metrics['optimization_time']:.2f}초")
```

**Methods:**
- `calculate_directory_size()`: 디렉터리 크기 계산
- `identify_optimization_targets()`: 최적화 대상 파일 식별
- `optimize()`: 패키지 최적화 실행

## 🏷️ 16-Core TAG System API

TAG 시스템은 프로젝트 전반의 추적성을 보장합니다.

### TAG 카테고리

**SPEC (문서 추적):**
- `@REQ`: 요구사항 정의
- `@DESIGN`: 설계 문서
- `@TASK`: 구현 작업
- `@TEST`: 테스트 케이스

**STEERING (원칙 추적):**
- `@VISION`: 프로젝트 비전
- `@STRUCT`: 구조 설계
- `@TECH`: 기술 선택
- `@ADR`: 아키텍처 결정

**IMPLEMENTATION (코드 추적):**
- `@FEATURE`: 기능 개발
- `@API`: API 설계
- `@DATA`: 데이터 모델링

**QUALITY (품질 추적):**
- `@PERF`: 성능 최적화
- `@SEC`: 보안 검토
- `@DEBT`: 기술 부채
- `@TODO`: 할 일 추적

### TAG 사용 예시

```python
"""
사용자 인증 모듈

@REQ:USER-AUTH-001 - JWT 기반 사용자 인증 요구사항
@DESIGN:AUTH-SYSTEM-001 - 인증 시스템 설계
@TASK:JWT-IMPL-001 - JWT 토큰 구현 작업
@TEST:AUTH-UNIT-001 - 인증 유닛 테스트
"""

class UserAuth:
    def authenticate(self, token: str) -> bool:
        """
        @API:AUTH-VALIDATE-001 - 토큰 검증 API
        @PERF:TOKEN-CACHE-001 - 토큰 캐싱으로 성능 최적화
        """
        pass
```

## 🔧 Error Handling

### Common Error Codes

- **INIT_001**: 프로젝트 초기화 실패
- **CONFIG_002**: 설정 파일 오류
- **PERMISSION_003**: 권한 부족
- **RESOURCE_004**: 리소스 파일 누락
- **VERSION_005**: 버전 불일치

### Error Response Format

```python
{
    "success": False,
    "error": "상세 오류 메시지",
    "errors": ["오류1", "오류2"],
    "code": "ERROR_CODE",
    "context": {
        "component": "installer",
        "operation": "copy_files"
    }
}
```

## 📊 Performance Metrics (SPEC-003 최적화 결과)

### 패키지 크기 최적화
- **이전**: 948KB
- **현재**: 192KB
- **개선**: **80% 감소**

### 설치 성능
- **설치 시간**: **50% 단축**
- **메모리 사용량**: **70% 감소**
- **네트워크 전송량**: **80% 감소**

### 파일 구조 최적화
- **에이전트 파일**: 60개 → 4개 (**93% 감소**)
- **명령어 파일**: 13개 → 3개 (**77% 감소**)
- **템플릿 구조**: _templates 폴더 완전 제거

## 🔗 Integration Examples

### Claude Code에서 사용

```bash
# Claude Code 터미널에서
cd my-project
moai init --interactive

# MoAI-ADK 명령어 사용
/moai:1-spec "사용자 인증"
/moai:2-build
/moai:3-sync
```

### GitHub Actions 통합

```yaml
name: MoAI-ADK CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Setup Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'
    - name: Install MoAI-ADK
      run: pip install moai-adk
    - name: Check project status
      run: moai status --verbose
    - name: Run health check
      run: moai doctor
```

## 📚 Related Documentation

- [Installation Guide](INSTALLATION.md)
- [Architecture Overview](ARCHITECTURE.md)
- [CLI Commands Reference](guides/CLI_COMMANDS.md)
- [Constitution 5원칙](sections/15-constitution.md)
- [SPEC-003 Package Optimization](.moai/specs/SPEC-003/spec.md)

---

**API Version**: v0.1.26 | **Last Updated**: 2025-01-19 | **SPEC-003 Optimized** ✅
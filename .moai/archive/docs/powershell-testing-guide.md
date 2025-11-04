# PowerShell 테스트 실행 가이드

## 개요

MoAI-ADK는 Bash와 PowerShell 모두에서 패키지 검증 테스트를 지원합니다. 이를 통해 Windows, macOS, Linux에서 일관된 개발 경험을 제공합니다.

**@TAG:POWERSHELL-TEST-GUIDE-001** | Cross-platform PowerShell test documentation

---

## 목차

1. [빠른 시작](#빠른-시작)
2. [환경 설정](#환경-설정)
3. [로컬 테스트 실행](#로컬-테스트-실행)
4. [테스트 유형](#테스트-유형)
5. [CI/CD 통합](#cicd-통합)
6. [문제 해결](#문제-해결)

---

## 빠른 시작

### macOS / Linux (Bash)

```bash
# 기본 테스트 (모든 셸)
./test.sh

# Bash만 테스트
./test.sh bash

# 상세 로그와 함께
./test.sh bash -v
```

### Windows (PowerShell)

```powershell
# PowerShell로 테스트 스크립트 실행
pwsh -NoProfile -File "tests\shell\powershell\helpers\runner.ps1"

# 상세 로그
pwsh -NoProfile -File "tests\shell\powershell\helpers\runner.ps1" -Verbose
```

---

## 환경 설정

### 사전 요구사항

| 환경 | 요구사항 | 설치 방법 |
|------|---------|----------|
| **Python** | 3.11+ | https://python.org |
| **PowerShell** | 7.0+ (Core) | https://github.com/PowerShell/PowerShell |
| **pytest** | 8.4.2+ | `pip install pytest pytest-cov` |
| **Git** | 2.0+ | https://git-scm.com |

### PowerShell Core 설치

#### Windows (Chocolatey)

```powershell
choco install powershell-core
```

#### macOS (Homebrew)

```bash
brew install powershell
```

#### Linux (Ubuntu/Debian)

```bash
# 저장소 추가
sudo add-apt-repository universe
sudo apt update

# PowerShell 설치
sudo apt install -y powershell
```

### 패키지 설치

```bash
# 개발 의존성 포함
pip install -e ".[dev]"
```

---

## 로컬 테스트 실행

### 1. 통합 테스트 스크립트 (권장)

**모든 설치된 셸에서 테스트 실행:**

```bash
./test.sh
```

**출력 예시:**

```
═══════════════════════════════════════════════════════════
MoAI-ADK 멀티셸 테스트 시작
═══════════════════════════════════════════════════════════
타임스탐프: 2025-11-02 14:30:45
선택된 셸: all

✓ Bash 사용 가능 (GNU bash, version 5.2.26)
✓ PowerShell 사용 가능 (7.4.6)

[INFO] Bash 테스트 실행 [all]
✓ 명령어 가용성
✓ 패키지 설치 완료
✓ 패키지 모듈 로드
✓ pytest 테스트 통과
✓ 타입 체크 통과
✓ 린팅 체크 통과

[INFO] PowerShell 테스트 실행 [all]
✓ 패키지 설치 검증
✓ 모듈 로드
✓ pytest 테스트 통과

모든 선택된 셸 테스트가 성공했습니다! ✓
```

### 2. Bash 테스트만 실행

```bash
./test.sh bash
```

또는

```bash
bash tests/shell/bash/test-runner.sh
```

### 3. PowerShell 테스트만 실행 (macOS/Linux)

```bash
pwsh -NoProfile -File "tests/shell/powershell/helpers/runner.ps1"
```

또는 (Bash에서)

```bash
./test.sh powershell
```

### 4. Windows에서 PowerShell 테스트

```powershell
pwsh -NoProfile -File "tests\shell\powershell\helpers\runner.ps1"
```

### 5. 상세 로그 옵션

```bash
# Bash 상세 로그
./test.sh bash -v

# PowerShell 상세 로그
pwsh -NoProfile -File "tests/shell/powershell/helpers/runner.ps1" -Verbose
```

---

## 테스트 유형

### 전체 테스트 실행

**기본값** - 모든 검증 실행:

```bash
./test.sh all
```

**포함 항목:**
- ✓ 패키지 설치 검증
- ✓ 필수 명령어 확인
- ✓ 모듈 로드 테스트
- ✓ pytest 실행 (unit, integration, hooks)
- ✓ 타입 체크 (mypy)
- ✓ 코드 린팅 (ruff)

### 패키지 테스트만

```bash
./test.sh package
```

**포함 항목:**
- ✓ 패키지 설치 및 모듈 로드
- ✓ `tests/unit/` 테스트만 실행

### 훅 테스트만

```bash
./test.sh hooks
```

**포함 항목:**
- ✓ Alfred 훅 시스템 테스트
- ✓ `tests/hooks/` 테스트만 실행

### CLI 통합 테스트

```bash
./test.sh cli
```

**포함 항목:**
- ✓ CLI 명령어 통합 테스트
- ✓ `tests/integration/` 테스트만 실행

---

## CI/CD 통합

### GitHub Actions 자동 테스트

MoAI-ADK는 다음 환경에서 **자동으로** PowerShell 테스트를 실행합니다:

| 환경 | 트리거 | 실행 방식 |
|------|--------|----------|
| **Linux (Ubuntu)** | `push`, `pull_request` | `moai-pipeline` job (Bash) |
| **Windows** | `pull_request`, `develop`, `feature/*` | `powershell-tests` job (PowerShell) |

### 테스트 결과 확인

1. **GitHub Actions 대시보드에서 확인:**
   - PR → "Checks" 탭
   - 🪟 PowerShell Cross-Platform Tests 클릭

2. **로컬에서 재현:**

   ```bash
   # Linux에서 실행
   ./test.sh bash

   # Windows에서 실행 (동일한 테스트)
   pwsh -NoProfile -File "tests/shell/powershell/helpers/runner.ps1"
   ```

### Draft PR vs Ready PR

- **Draft PR**: 테스트 실패 허용 (개발 진행 중)
- **Ready PR**: 테스트 반드시 통과 (병합 가능)

---

## pytest 직접 실행

### 모든 테스트

```bash
pytest tests/ -v
```

### 특정 범주만

```bash
# Unit 테스트만
pytest tests/unit/ -v

# 통합 테스트만
pytest tests/integration/ -v

# 훅 테스트만
pytest tests/hooks/ -v

# 특정 파일
pytest tests/unit/test_cli.py -v
```

### 커버리지 리포트

```bash
pytest tests/ --cov=src/moai_adk --cov-report=html
open htmlcov/index.html  # macOS
start htmlcov/index.html # Windows
```

---

## Windows에서의 특별 고려사항

### 경로 구분자

PowerShell에서는 자동으로 경로를 처리하므로, Bash와 동일한 명령을 사용할 수 있습니다:

```powershell
# ✓ 둘 다 작동
pytest tests/unit
pytest tests\unit
```

### 긴 파일 경로 문제

Windows의 260자 제한을 해결하려면:

```powershell
# PowerShell에서
New-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\FileSystem" `
  -Name "LongPathsEnabled" -Value 1 -PropertyType DWORD -Force
```

### 문자 인코딩

PowerShell은 기본으로 UTF-8을 지원하지만, 호환성을 위해:

```powershell
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
```

---

## 문제 해결

### PowerShell 스크립트 실행 정책

```powershell
# 현재 정책 확인
Get-ExecutionPolicy

# 사용자 권한으로 허용 (권장)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# 특정 스크립트만 실행
pwsh -NoProfile -ExecutionPolicy Bypass -File "script.ps1"
```

### Python 모듈 임포트 오류

```bash
# 재설치
pip install -e ".[dev]" --force-reinstall

# 캐시 삭제 후 재설치
pip cache purge
pip install -e ".[dev]"
```

### pytest 검색 오류

```bash
# 테스트 검색 상태 확인
pytest --collect-only

# 특정 경로 명시
pytest tests/unit/ -v
```

### PowerShell과 Bash 결과 불일치

**대부분의 경우 환경 차이로 인함:**

1. **Python 버전 확인**
   ```bash
   python --version  # 둘 다 동일해야 함
   ```

2. **venv 활성화 확인**
   ```bash
   # Bash
   source .venv/bin/activate

   # PowerShell
   .\.venv\Scripts\Activate.ps1
   ```

3. **의존성 버전 확인**
   ```bash
   pip list | grep pytest
   ```

---

## 자주 묻는 질문

### Q: PowerShell이 설치되지 않은 경우?

**A:** Bash 테스트만 실행됩니다. 자동 감지 기능이 있어서 선택적으로 실행됩니다.

```bash
./test.sh bash  # PowerShell 없이도 가능
```

### Q: 특정 테스트만 실행하려면?

**A:** pytest의 `-k` 옵션 사용:

```bash
pytest tests/ -k "test_install" -v
```

### Q: 테스트 성능을 개선하려면?

**A:** 병렬 실행 (`pytest-xdist` 설치 필수):

```bash
pip install pytest-xdist
pytest tests/ -n auto
```

### Q: Windows에서 Bash를 사용할 수 있나?

**A:** 네, 다음 옵션 사용:

- **WSL 2** (Windows Subsystem for Linux)
- **Git Bash** (Git for Windows 포함)
- **MinGW** (별도 설치)

---

## 성능 최적화

### 테스트 병렬화

```bash
# 자동 CPU 수 감지
pytest tests/ -n auto

# 특정 워커 수
pytest tests/ -n 4
```

### 테스트 캐싱

pytest는 자동으로 `.pytest_cache` 디렉토리를 생성합니다. 캐시를 초기화하려면:

```bash
pytest --cache-clear tests/
```

---

## 관련 문서

- [SPEC-WINDOWS-HOOKS-001.md](.moai/specs/SPEC-WINDOWS-HOOKS-001/spec.md) - Windows 훅 시스템
- [pyproject.toml](pyproject.toml) - pytest 설정
- [GitHub Actions Workflow](.github/workflows/moai-gitflow.yml) - CI/CD 자동화

---

## 버전 관리

| 버전 | 변경 사항 | 날짜 |
|------|---------|------|
| 1.0 | 초기 가이드 작성 | 2025-11-02 |

---

## 라이선스

본 가이드는 MoAI-ADK와 동일한 라이선스를 따릅니다.

---

**질문이 있으신가요?**

- GitHub Issues: [MoAI-ADK Issues](https://github.com/your-repo/issues)
- 토론: [GitHub Discussions](https://github.com/your-repo/discussions)

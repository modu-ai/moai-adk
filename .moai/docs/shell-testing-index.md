# 셸 테스트 및 CI/CD 통합 문서

## 개요

MoAI-ADK의 셸 테스트 인프라는 Bash와 PowerShell 모두에서 일관된 패키지 검증을 제공합니다.

**@TAG:SHELL-TEST-INDEX-001** | Shell testing infrastructure documentation index

---

## 문서 맵

### 사용자 가이드

| 문서 | 목적 | 대상 |
|------|------|------|
| [PowerShell 테스트 실행 가이드](powershell-testing-guide.md) | 로컬 및 CI/CD에서 PowerShell 테스트 실행 방법 | 개발자 |
| [테스트 구조 분석](test-structure-analysis.md) | MoAI-ADK 테스트 아키텍처 이해 | 기여자 |

### 기술 문서

| 파일 | 목적 | 경로 |
|------|------|------|
| conftest.py | pytest 픽스처 및 셸 헬퍼 | `tests/shell/conftest.py` |
| runner.ps1 | PowerShell 테스트 런타임 | `tests/shell/powershell/helpers/runner.ps1` |
| test-runner.sh | Bash 테스트 런타임 | `tests/shell/bash/test-runner.sh` |
| test.sh | 멀티셸 통합 매니저 | `test.sh` (루트) |

### GitHub Actions 워크플로우

| 워크플로우 | 목적 | 환경 |
|-----------|------|------|
| `moai-gitflow.yml` | 주요 CI/CD 파이프라인 | Ubuntu (Linux) |
| `moai-gitflow.yml` (powershell-tests job) | PowerShell 크로스플랫폼 테스트 | Windows |

---

## 빠른 참조

### 로컬 테스트

```bash
# 모든 셸에서 테스트
./test.sh

# Bash만
./test.sh bash

# PowerShell만 (macOS/Linux)
pwsh -NoProfile -File "tests/shell/powershell/helpers/runner.ps1"

# 상세 로그
./test.sh bash -v
```

### CI/CD 동작

| 브랜치 | 환경 | 테스트 | 정책 |
|-------|------|--------|------|
| `develop` | Ubuntu + Windows | 모두 실행 | 반드시 통과 |
| `feature/*` | Ubuntu + Windows | 모두 실행 | 반드시 통과 |
| `main` | GitHub Release만 | PyPI 배포 테스트 | 반드시 통과 |

### 테스트 범위

| 테스트 | 포함 | Bash | PowerShell |
|--------|------|------|-----------|
| 패키지 설치 | ✓ | ✓ | ✓ |
| 모듈 로드 | ✓ | ✓ | ✓ |
| pytest 실행 | ✓ | ✓ | ✓ |
| 타입 체크 (mypy) | ✓ | ✓ | 옵션 |
| 린팅 (ruff) | ✓ | ✓ | 옵션 |
| 스크립트 호환성 | ✓ | - | ✓ |

---

## 파일 구조

```
MoAI-ADK/
├── test.sh                                    # 멀티셸 통합 진입점
├── .github/workflows/
│   └── moai-gitflow.yml                      # GitHub Actions (Ubuntu + Windows)
├── tests/
│   ├── shell/
│   │   ├── conftest.py                       # pytest 픽스처 (공유)
│   │   ├── bash/
│   │   │   └── test-runner.sh               # Bash 런타임
│   │   ├── powershell/
│   │   │   └── helpers/
│   │   │       └── runner.ps1               # PowerShell 런타임
│   │   ├── cross_platform/                   # (향후 확장)
│   │   └── fixtures/                         # (향후 확장)
│   ├── unit/                                 # 단위 테스트 (80+ 파일)
│   ├── integration/                          # 통합 테스트
│   ├── e2e/                                  # E2E 테스트
│   └── hooks/                                # 훅 테스트
└── .moai/docs/
    ├── powershell-testing-guide.md          # 사용자 가이드
    └── shell-testing-index.md               # 이 파일
```

---

## 아키텍처

```
┌─────────────────────────────────────────────────────┐
│  개발자 커맨드                                       │
├─────────────────────────────────────────────────────┤
│  ./test.sh [shell] [type] [-v]                     │
│  (통합 매니저 - OS 자동 감지)                       │
└────────────┬──────────────────────┬────────────────┘
             │                      │
      ┌──────▼────────┐      ┌──────▼──────────┐
      │ Bash (Linux)  │      │ PowerShell (Win)│
      │ tests/shell/  │      │ tests/shell/    │
      │ bash/         │      │ powershell/     │
      │ test-runner.sh│      │ helpers/        │
      │              │      │ runner.ps1      │
      └──────┬────────┘      └──────┬──────────┘
             │                      │
             └──────────┬───────────┘
                        │
                  ┌─────▼──────┐
                  │  pytest    │
                  │  runner    │
                  │            │
                  │ + mypy     │
                  │ + ruff     │
                  └─────┬──────┘
                        │
            ┌───────────┴───────────┐
            │                       │
      ┌─────▼──────┐         ┌──────▼──────┐
      │ Test Results│         │ Exit Code   │
      │ Reporting  │         │ (0 = pass)  │
      └────────────┘         └─────────────┘
```

---

## 점진적 확장 로드맵

### Phase 1: 기본 구조 (완료 ✓)

- ✓ PowerShell conftest.py 작성
- ✓ Bash test-runner.sh 작성
- ✓ PowerShell helpers/runner.ps1 작성
- ✓ 통합 test.sh 작성
- ✓ GitHub Actions 워크플로우 업데이트
- ✓ 사용자 가이드 작성

### Phase 2: pytest 통합 (다음)

- [ ] Pytest PowerShell 플러그인 개발
- [ ] PowerShell 테스트 마커 확장
- [ ] 크로스플랫폼 픽스처 개선

### Phase 3: 고급 기능 (향후)

- [ ] 성능 벤치마크 테스트
- [ ] 메모리 누수 감지
- [ ] 플랫폼별 호환성 매트릭스
- [ ] 자동 보고서 생성

---

## @TAG 추적

| @TAG | 파일 | 목적 |
|------|------|------|
| POWERSHELL-TEST-001 | tests/shell/conftest.py | pytest 픽스처 프레임워크 |
| POWERSHELL-TEST-002 | tests/shell/powershell/helpers/runner.ps1 | PowerShell 테스트 실행기 |
| SHELL-TEST-001 | tests/shell/bash/test-runner.sh | Bash 테스트 실행기 |
| INTEGRATION-TEST-001 | test.sh | 멀티셸 통합 매니저 |
| POWERSHELL-CI-001 | .github/workflows/moai-gitflow.yml | Windows CI/CD 검증 |
| POWERSHELL-CI-002 | .github/workflows/moai-gitflow.yml | Windows pytest 실행 |
| POWERSHELL-TEST-GUIDE-001 | .moai/docs/powershell-testing-guide.md | 사용자 문서 |
| SHELL-TEST-INDEX-001 | .moai/docs/shell-testing-index.md | 이 파일 |

---

## 환경 호환성 매트릭스

| 환경 | Python | 셸 | OS | 지원 | 테스트 위치 |
|------|--------|----|----|------|-----------|
| **macOS** | 3.11+ | Bash | Sonoma+ | ✓ | 로컬 + CI |
| **macOS** | 3.11+ | PowerShell | 7.0+ | ✓ | 로컬 + CI |
| **Ubuntu** | 3.11+ | Bash | 22.04+ | ✓ | CI (모이-gitflow) |
| **Windows** | 3.11+ | PowerShell | 10/11 | ✓ | CI (powershell-tests job) |
| **Windows** | 3.11+ | Bash (WSL 2) | 10/11 | ✓ | 로컬 |

---

## 기여 가이드

테스트 인프라에 기여하려면:

1. **로컬 환경에서 테스트:**
   ```bash
   ./test.sh all -v
   ```

2. **새 테스트 추가:**
   - `tests/shell/powershell/` 또는 `tests/shell/bash/`에 새 파일 추가
   - 해당 conftest.py 픽스처 참조

3. **@TAG 추가:**
   - 새 파일에 주석으로 @TAG 추가
   - 이 INDEX에 등록

4. **문서화:**
   - 변경사항을 PowerShell 테스트 가이드에 기록
   - HISTORY 섹션 업데이트

---

## 관련 리소스

- [pytest 공식 문서](https://docs.pytest.org)
- [PowerShell 문서](https://learn.microsoft.com/en-us/powershell/)
- [Bash 가이드](https://www.gnu.org/software/bash/manual/)

---

## 라이선스

본 문서는 MoAI-ADK와 동일한 라이선스를 따릅니다.

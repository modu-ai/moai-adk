---
spec_id: SPEC-004
status: active
priority: medium
dependencies: []
tags:
  - cross-platform
  - windows
  - macos
  - linux
  - compatibility
  - installation
---

# SPEC-004: Windows/macOS/Linux 환경별 최적화

## @REQ:PLATFORM-004 프로젝트 컨텍스트

### 배경

MoAI-ADK는 크로스 플랫폼 지원을 통해 다양한 개발 환경에서 일관된 사용자 경험을 제공해야 합니다. 현재 @TASK:CROSS-PLATFORM-001의 핵심 부분인 Windows/macOS/Linux 환경별 호환성 문제를 우선적으로 해결해야 하며, 배포 채널 확장은 별도로 처리합니다.

### 문제 정의

- **현재 상태**: 플랫폼별 경로 처리, 권한 설정, 설치 과정에서 불일치 발생
- **핵심 문제**: Windows `\` vs Unix `/` 경로 구분자, 권한 설정 방식의 차이
- **비즈니스 영향**: 사용자 설치 실패 및 크로스 플랫폼 개발팀의 협업 저해

### 목표

1. 플랫폼별 경로 처리 표준화를 통한 완전한 호환성 달성
2. 환경별 권한 설정 자동화로 설치 과정 단순화
3. 크로스 플랫폼 자동 테스트를 통한 품질 보장

## @DESIGN:PLATFORM-SYSTEM-004 환경 및 가정사항

### Environment (환경)

- **지원 플랫폼**: Windows 10/11, macOS 10.15+, Linux (Ubuntu 20.04+)
- **Python 버전**: 3.11, 3.12, 3.13 (일관된 지원)
- **패키지 관리**: pip, pipx (모든 플랫폼 공통)
- **권한**: 관리자 권한 없이 사용자 레벨 설치 지원

### Assumptions (가정사항)

- 사용자는 Python과 pip가 이미 설치되어 있음
- 기본적인 터미널/명령 프롬프트 사용 가능
- Git이 설치되어 있음 (MoAI 워크플로우 사용 시)
- 인터넷 연결이 가능함 (패키지 다운로드 및 업데이트 시)

## @TASK:IMPLEMENT-004 요구사항 명세

### R1. 플랫폼별 경로 처리 표준화

**WHEN** 파일 경로가 처리되거나 생성될 때,
**THE SYSTEM SHALL** 자동으로 해당 플랫폼에 적합한 경로 형식을 적용해야 함

**상세 요구사항:**

- Windows: 백슬래시(`\`) 구분자 자동 적용
- Unix 계열 (macOS, Linux): 슬래시(`/`) 구분자 자동 적용
- 상대/절대 경로 변환 시 플랫폼 호환성 보장
- 파일명 길이 제한 및 특수문자 처리 (플랫폼별 차이 고려)

### R2. 환경별 권한 설정 자동화

**WHEN** MoAI-ADK가 설치되거나 파일을 생성할 때,
**THE SYSTEM SHALL** 플랫폼에 적합한 권한을 자동으로 설정해야 함

**상세 요구사항:**

- Windows: ACL (Access Control List) 기반 권한 설정
- macOS/Linux: chmod 기반 실행 권한 자동 부여
- 사용자 홈 디렉토리 내 설치 시 권한 최적화
- 관리자 권한 없이 정상 작동 보장

### R3. 플랫폼별 설치 스크립트 최적화

**WHEN** 사용자가 MoAI-ADK를 설치할 때,
**THE SYSTEM SHALL** 해당 플랫폼에 최적화된 설치 과정을 제공해야 함

**상세 요구사항:**

- Windows: PowerShell 스크립트 지원
- macOS: bash/zsh 스크립트 지원
- Linux: 주요 배포판별 패키지 관리자 고려
- 환경 변수 설정 자동화 (PATH, PYTHONPATH 등)

### R4. 크로스 플랫폼 호환성 자동 테스트

**WHEN** 코드 변경이나 릴리스가 준비될 때,
**THE SYSTEM SHALL** 모든 지원 플랫폼에서 자동으로 호환성을 검증해야 함

**상세 요구사항:**

- GitHub Actions 기반 멀티 플랫폼 CI/CD
- 각 플랫폼별 핵심 기능 E2E 테스트
- 성능 벤치마크 플랫폼별 측정
- 플랫폼별 패키지 무결성 검증

## @TEST:ACCEPTANCE-004 Acceptance Criteria

### AC1. 경로 처리 호환성

**Given** Windows 환경에서 파일 경로를 생성할 때
**When** MoAI-ADK가 `.moai/` 또는 `.claude/` 디렉토리를 생성하면
**Then** 모든 경로가 백슬래시(`\`) 구분자를 사용하고 Windows 파일명 규칙을 준수해야 함

**Given** macOS 또는 Linux 환경에서 파일 경로를 생성할 때
**When** 동일한 작업을 수행하면
**Then** 모든 경로가 슬래시(`/`) 구분자를 사용하고 Unix 파일명 규칙을 준수해야 함

### AC2. 권한 설정 자동화

**Given** Windows 환경에서 MoAI-ADK를 설치할 때
**When** 스크립트 파일이 생성되면
**Then** 관리자 권한 없이도 해당 스크립트가 실행 가능해야 함

**Given** macOS 또는 Linux 환경에서 설치할 때
**When** 실행 파일이나 스크립트가 생성되면
**Then** 자동으로 적절한 실행 권한 (755)이 설정되어야 함

### AC3. 플랫폼별 설치 최적화

**Given** 각 플랫폼에서 설치를 진행할 때
**When** `pip install moai-adk` 명령을 실행하면
**Then** 해당 플랫폼에 최적화된 설치 과정이 진행되고, 환경 변수가 자동으로 설정되어야 함

**Given** 설치 완료 후 각 플랫폼에서
**When** `/moai:0-project` 명령을 실행하면
**Then** 모든 플랫폼에서 일관된 결과와 사용자 경험을 제공해야 함

### AC4. 자동 호환성 테스트

**Given** 코드 변경이 GitHub에 푸시될 때
**When** CI/CD 파이프라인이 실행되면
**Then** Windows, macOS, Linux 모든 환경에서 테스트가 통과해야 함

**Given** 플랫폼별 테스트 중 하나라도 실패할 때
**When** 테스트 결과를 확인하면
**Then** 구체적인 플랫폼별 오류 정보와 수정 방법이 제공되어야 함

## 범위 및 모듈

### In Scope

- Windows/macOS/Linux 경로 처리 표준화
- 플랫폼별 권한 설정 자동화
- 설치 스크립트 플랫폼별 최적화
- 멀티 플랫폼 CI/CD 파이프라인 구축
- 핵심 기능의 크로스 플랫폼 호환성 검증

### Out of Scope

- 다른 패키지 관리자 지원 (conda, brew 등은 별도 SPEC)
- 아키텍처별 최적화 (x86, ARM64는 기본 지원만)
- 플랫폼별 UI/UX 커스터마이징
- 레거시 OS 버전 지원 (지정된 최소 버전 이하)

## 기술 노트

### 구현 기술

- **경로 처리**: Python `pathlib.Path` + `os.path` 하이브리드
- **권한 설정**: Windows `pywin32`, Unix `os.chmod`
- **설치 자동화**: setuptools hooks + platform detection
- **테스팅**: tox + GitHub Actions matrix builds

### 의존성

- **Windows**: colorama (콘솔 색상), pywin32 (선택적)
- **macOS/Linux**: 표준 라이브러리만 사용
- **공통**: pathlib, os, platform (Python 표준 라이브러리)

### 플랫폼별 고려사항

#### Windows 특수 사항

- 긴 경로명 지원 (260자 제한 우회)
- 예약된 파일명 처리 (CON, PRN, AUX 등)
- 대소문자 구분하지 않는 파일시스템 고려

#### macOS 특수 사항

- Gatekeeper 보안 정책 준수
- SIP (System Integrity Protection) 고려
- 번들 앱 구조와의 호환성

#### Linux 특수 사항

- 다양한 배포판별 차이 최소화
- 패키지 관리자 다양성 고려
- 컨테이너 환경에서의 동작 보장

### 성능 고려사항

- 플랫폼 감지 오버헤드 최소화 (한 번만 실행)
- 파일 작업 시 버퍼링 최적화
- 권한 설정 배치 처리로 성능 개선

## 추적성

### 연결된 요구사항

- @TASK:CROSS-PLATFORM-001: 크로스 플랫폼 완성 계획의 핵심 부분
- @TECH:STACK-001: 멀티 플랫폼 지원 현황 개선
- @REQ:USER-001: 개인 개발자의 즉시 사용 가능한 환경 제공

### 구현 우선순위

1. 경로 처리 표준화 (High) - 기본 동작에 필수
2. 권한 설정 자동화 (High) - 설치 성공률에 직결
3. CI/CD 멀티 플랫폼 테스트 (Medium) - 품질 보장
4. 설치 스크립트 최적화 (Medium) - 사용자 경험 개선

### 테스트 전략

- 단위 테스트: 각 플랫폼별 경로/권한 처리 함수 테스트
- 통합 테스트: 크로스 플랫폼 설치 시나리오 테스트
- E2E 테스트: 실제 각 플랫폼에서 전체 워크플로우 검증

## 플랫폼별 구현 예시

### 경로 처리 유틸리티

```python
import os
import platform
from pathlib import Path

def get_platform_path(relative_path: str) -> Path:
    """플랫폼에 적합한 경로 객체 반환"""
    if platform.system() == "Windows":
        # Windows 특수 처리
        return Path(relative_path).resolve()
    else:
        # Unix 계열 처리
        return Path(relative_path).expanduser().resolve()

def ensure_executable(file_path: Path) -> None:
    """플랫폼별 실행 권한 설정"""
    if platform.system() != "Windows":
        os.chmod(file_path, 0o755)
```

### 설치 스크립트 감지

```python
def get_install_commands() -> dict:
    """플랫폼별 설치 명령어 반환"""
    system = platform.system().lower()

    if system == "windows":
        return {
            "shell": "powershell",
            "pip": "python -m pip",
            "path_sep": ";"
        }
    else:
        return {
            "shell": "bash",
            "pip": "pip3",
            "path_sep": ":"
        }
```

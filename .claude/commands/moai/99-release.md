---
name: moai:99-release
description: "Interactive release management for MoAI-ADK packages with menu-driven workflow"
argument-hint: "[no arguments - uses interactive menu]"
allowed-tools:
  - Bash
  - Read
  - Write
  - Edit
  - AskUserQuestion
model: "haiku"
---

## 📋 Pre-execution Context

!git status --porcelain
!git branch --show-current
!git tag --list
!git log --oneline -5
!git remote -v

## 📁 Essential Files

@pyproject.toml
@src/moai_adk/__init__.py
@CHANGELOG.md

---

# 🚀 MoAI-ADK 인터랙티브 릴리즈 관리

## EXCEPTION: Local-Only Development Tool

This command is exempt from "Zero Direct Tool Usage" principle because:

1. **Local development only** - Not distributed with package distributions
2. **Maintainer-only usage** - GoosLab (project owner) exclusive access
3. **Direct system access required** - PyPI release automation requires direct shell commands
4. **Interactive menu system** - Uses AskUserQuestion for user-driven workflow

**Production commands** (`/moai:0-project`, `/moai:1-plan`, `/moai:2-run`, `/moai:3-sync`) must strictly adhere to agent delegation principle.

---

> **Local-Only Tool**: This command is for local development only.
> Never deployed with package distributions.
> Use for PyPI releases, version bumping, changelog generation, and emergency rollbacks.

---

## 🎯 사용 방법

**인터랙티브 메뉴 방식**:

```bash
/moai:99-release
```

메뉴에서 원하는 작업을 선택하고 안내에 따라 진행하세요.

## 🔄 워크플로우

### 1단계: 메인 메뉴 선택

```
🚀 MoAI-ADK 릴리즈 관리 - 원하는 작업을 선택하세요:
```

### 2단계: 세부 옵션 선택

각 작업에 맞는 세부 옵션을 선택합니다.

### 3단계: 실행 확인

선택한 작업을 실행하기 전 최종 확인을 받습니다.

### 4단계: 결과 보고

실행 결과와 다음 단계를 안내받습니다.

---

## 📋 메인 메뉴 옵션

### 🔍 **validate** - 사전 릴리즈 품질 검증

**목적**: 릴리즈 전 코드 품질과 보안 검증

**세부 옵션**:

- **⚡ Quick 검증** (5분): 기본 품질 게이트
  - pytest 실행
  - ruff 코드 포맷 검사
  - mypy 타입 검사
- **🔬 Full 검증** (15분): 전체 품질 게이트 + 보안 스캔
  - Quick 검증 모든 항목
  - bandit 보안 스캔
  - pip-audit 의존성 취약점 검사
- **🎯 사용자 정의**: 특정 검증 항목 선택

**실행 결과**:

- 검증 통과/실패 보고
- 문제점 상세 분석
- 수정 권장 사항

### 📝 **version** - 버전 관리

**목적**: 시맨틱 버전 관리 (major/minor/patch)

**세부 옵션**:

- **🔢 patch**: 버그 수정 (0.27.2 → 0.27.3)
- **🔧 minor**: 기능 추가 (0.27.2 → 0.28.0)
- **💥 major**: 호환성 변경 (0.27.2 → 1.0.0)

**업데이트 파일**:

- `pyproject.toml` - 패키지 버전
- `src/moai_adk/__init__.py` - Python 버전
- `.moai/config/config.json` - 설정 버전

**실행 결과**:

- 모든 파일 버전 동기화 완료
- Git 커밋 생성 제안
- 다음 단계 안내

### 📋 **changelog** - 이중언어 변경로그 생성

**목적**: 한국어/영어 이중언어 변경로그 자동 생성

**세부 옵션**:

- **📝 자동 생성**: Git 히스토리 기반 자동 생성
  - 최신 태그 이후 커밋 분석
  - 자동 분류 (기능/버그/개선)
  - 이중언어 번역
- **✏️ 수동 편집**: 템플릿 제공 후 직접 편집
  - 표준 변경로그 템플릿
  - 사용자 편집 가이드
- **🔄 기존 수정**: 기존 changelog.md 수정

**실행 결과**:

- `CHANGELOG.md` 업데이트
- GitHub Release 노트 형식
- Git 커밋 제안

### 🚀 **prepare** - CI/CD 배포 준비

**목적**: PyPI/CI/CD 배포를 위한 준비 작업

**세부 옵션**:

- **🧪 test 환경**: TestPyPI 배포 준비
  - 패키지 빌드 검증
  - TestPyPI 토큰 유효성 확인
  - 테스트 배포 시뮬레이션
- **🌍 production 환경**: PyPI 배포 준비
  - 프로덕션 토큰 유효성 확인
  - 보안 검증 통과 확인
  - GitHub Actions 트리거 준비
- **📋 검토용**: 릴리즈 검토용 번들 생성
  - 릴리즈 노트 미리보기
  - 배포 검토 체크리스트
  - 승인 요청 번들

**실행 결과**:

- 배포 준비 상태 보고
- GitHub Actions 실행 안내
- 승인 절차 안내

### 🚨 **rollback** - 응급 롤백

**목적**: 배포 실패 시 응급 롤백 절차

**세부 옵션**:

- **📦 PyPI만**: 패키지만 롤백
  - PyPI 버전 숨기기
  - 다운로드 차단
- **🔄 전체**: PyPI + GitHub Release + 태그
  - PyPI 완전 삭제
  - GitHub Release 삭제
  - Git 태그 삭제
- **🎯 특정 버전**: 특정 버전 지정 롤백
  - 버전별 선택
  - 부분 롤백

**실행 결과**:

- 롤백 실행 상태
- 복구 절차 안내
- 팀 알림 발송

---

## 🔒 보안 정책

### **환경별 접근 제어**:

- **테스트 환경**: 즉시 실행 가능
- **프로덕션 환경**: 5분 대기 + 확인 절차

### **토큰 관리**:

- **PyPI 토큰**: `~/.pypirc`에서 관리
- **GitHub 토큰**: GitHub Secrets에서 관리
- **유효성 검증**: 배포 전 자동 확인

### **승인 절차**:

- **개인 모드**: 1인 승인 가능
- **팀 모드**: 2인 이상 승인 필요

---

## 📊 모니터링 및 로깅

### **실행 기록**:

- `.moai/logs/release-*.log`에 상세 기록
- 각 단계별 타임스탬프
- 성공/실패 상세 원인

### **알림 시스템**:

- Slack/이메일 알림 (설정 시)
- GitHub Issues 자동 생성 (롤백 시)
- 팀 멤버 알림 (프로덕션 배포 시)

---

## 🆘️ 문제 해결

### **일반적인 문제**:

1. **토큰 만료**: 새 토큰 발급 필요
2. **권한 부족**: PyPI/GitHub 권한 확인
3. **네트워크 오류**: 방화벽/프록시 설정 확인

### **긴급 연락처**:

- **팀 리드**: GitHub Issues
- **PyPI 지원**: pypi@python.org
- **GitHub 지원**: support@github.com

---

## 🔗 관련 문서

- **CI/CD 워크플로우**: `.github/workflows/release-secure.yml`
- **보안 정책**: `.moai/security/` 디렉토리
- **모니터링**: `.moai/monitoring/` 디렉토리
- **응급 절차**: `.moai/emergency/` 디렉토리

---

**Status**: Local-Only Development Tool
**Python Version**: 3.14
**MoAI-ADK Version**: 0.27.2+
**Last Updated**: 2025-11-21

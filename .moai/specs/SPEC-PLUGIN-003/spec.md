---
# 필수 필드 (7개)
id: PLUGIN-003
version: 0.0.1
status: draft
created: 2025-10-10
updated: 2025-10-10
author: @Goos
priority: medium

# 선택 필드 - 분류/메타
category: feature
labels:
  - installer
  - shell-script
  - deployment

# 선택 필드 - 관계 (의존성 그래프)
depends_on:
  - PLUGIN-001

# 선택 필드 - 범위 (영향 분석)
scope:
  packages:
    - scripts/
  files:
    - install.sh
    - install.ps1
---

# @SPEC:PLUGIN-003: 플러그인 설치 스크립트 작성

## HISTORY

### v0.0.1 (2025-10-10)
- **INITIAL**: 플러그인 설치 스크립트 SPEC 작성
- **AUTHOR**: @Goos
- **SCOPE**: curl 원라이너 설치 스크립트, Git/tar.gz 기반 설치, 크로스 플랫폼 지원
- **CONTEXT**: npm 패키지 제거 후 사용자가 플러그인을 수동 설치할 수 있도록 자동화 스크립트 제공

---

## 📋 개요

MoAI-ADK 플러그인을 간편하게 설치할 수 있는 원라이너 설치 스크립트를 제공합니다. Git 설치 여부에 따라 Git 클론 또는 tar.gz 다운로드 방식을 자동 선택하며, macOS/Linux는 bash 스크립트, Windows는 PowerShell 스크립트로 지원합니다.

## 🎯 목표

1. curl 원라이너 설치 지원 (`curl -sSL https://moai-adk.dev/install.sh | sh`)
2. Git 설치 감지 및 자동 선택 (Git 클론 vs tar.gz 다운로드)
3. GitHub Release API 활용 (latest 버전 자동 다운로드)
4. 설치 경로: `~/.claude/plugins/moai-adk`
5. 설치 검증 및 사용자 안내 메시지 제공
6. 크로스 플랫폼 지원 (macOS, Linux, Windows)

## 📝 EARS 요구사항

### Ubiquitous Requirements (기본 기능)

1. **설치 스크립트 제공**
   - 시스템은 `install.sh` bash 스크립트를 제공해야 한다
   - 시스템은 Git 설치 여부를 감지해야 한다
   - 시스템은 설치 경로 `~/.claude/plugins/moai-adk`를 사용해야 한다

2. **설치 방식 선택**
   - 시스템은 Git 설치 시 `git clone` 방식을 우선 사용해야 한다
   - 시스템은 Git 미설치 시 tar.gz 다운로드 방식을 대체 사용해야 한다
   - 시스템은 GitHub Release API를 사용하여 최신 버전을 다운로드해야 한다

3. **설치 완료 메시지**
   - 시스템은 설치 완료 후 다음 단계를 안내해야 한다
   - 시스템은 `plugin.json` 존재를 확인해야 한다
   - 시스템은 Claude Code 재시작을 안내해야 한다

### Event-driven Requirements (이벤트 기반)

1. **Git 감지 시**
   - WHEN Git이 설치되어 있으면, 시스템은 `git clone https://github.com/modu-ai/moai-adk ~/.claude/plugins/moai-adk` 명령을 실행해야 한다

2. **Git 미설치 시**
   - WHEN Git이 설치되어 있지 않으면, 시스템은 GitHub Release API를 호출하여 최신 tar.gz를 다운로드해야 한다
   - WHEN tar.gz 다운로드 완료 시, 시스템은 `~/.claude/plugins/moai-adk`에 압축을 해제해야 한다

3. **네트워크 오류 시**
   - WHEN GitHub API 요청 실패 시, 시스템은 사용자에게 수동 설치 방법을 안내해야 한다
   - WHEN 다운로드 실패 시, 시스템은 에러 메시지와 함께 종료해야 한다

4. **이미 설치된 경우**
   - WHEN `~/.claude/plugins/moai-adk`가 이미 존재하면, 시스템은 덮어쓰기 여부를 확인해야 한다
   - WHEN 사용자가 덮어쓰기를 거부하면, 시스템은 설치를 중단해야 한다

### State-driven Requirements (상태 기반)

1. **설치 진행 중**
   - WHILE 설치 진행 중일 때, 시스템은 진행률 메시지를 출력해야 한다
   - WHILE Git 클론 중일 때, 시스템은 "Cloning MoAI-ADK plugin..." 메시지를 표시해야 한다
   - WHILE tar.gz 다운로드 중일 때, 시스템은 "Downloading MoAI-ADK plugin..." 메시지를 표시해야 한다

2. **설치 검증 중**
   - WHILE 설치 검증 중일 때, 시스템은 `plugin.json` 존재를 확인해야 한다
   - WHILE 설치 검증 중일 때, 시스템은 `commands/`, `agents/` 디렉토리 존재를 확인해야 한다

### Optional Features (선택적 기능)

1. **Windows PowerShell 스크립트**
   - WHERE Windows 환경이면, 시스템은 `install.ps1` PowerShell 스크립트를 제공할 수 있다
   - WHERE PowerShell 7+이면, 시스템은 진행률 바를 표시할 수 있다

2. **커스텀 설치 경로**
   - WHERE 환경변수 `MOAI_INSTALL_PATH`가 설정되어 있으면, 시스템은 해당 경로를 설치 디렉토리로 사용할 수 있다

3. **개발 버전 설치**
   - WHERE `--dev` 플래그가 전달되면, 시스템은 develop 브랜치를 클론할 수 있다

### Constraints (제약사항)

1. **플러그인 디렉토리 권한**
   - IF `~/.claude/plugins/` 디렉토리 쓰기 권한이 없으면, 시스템은 설치를 실패하고 권한 가이드를 표시해야 한다

2. **네트워크 연결 필요**
   - IF 네트워크 연결이 없으면, 시스템은 설치를 실패하고 수동 설치 가이드를 표시해야 한다

3. **필수 도구**
   - IF curl 또는 wget이 없으면, 시스템은 설치를 실패하고 도구 설치 가이드를 표시해야 한다

4. **플러그인 무결성**
   - IF 다운로드한 파일에 `plugin.json`이 없으면, 시스템은 설치를 실패하고 재시도를 안내해야 한다

## 🔗 추적성 (Traceability)

- **@SPEC:PLUGIN-003** → 이 문서
- **@TEST:PLUGIN-003** → `tests/scripts/install.test.sh` (예정)
- **@CODE:PLUGIN-003** → `scripts/install.sh`, `scripts/install.ps1` (예정)
- **@DOC:PLUGIN-003** → `docs/installation.md` (예정)

---

## 참조

- **의존 SPEC**: @SPEC:PLUGIN-001
- **GitHub Release API**: https://docs.github.com/en/rest/releases/releases
- **Claude Code 플러그인 설치**: https://docs.claude.com/en/docs/claude-code/plugins-installation

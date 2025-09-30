---
title: CLI 치트시트
description: MoAI-ADK 명령어 빠른 참조
---

# CLI 치트시트

MoAI-ADK의 모든 명령어를 빠르게 찾아볼 수 있는 참조 가이드입니다.

## 목차

- [CLI 명령어](#cli-명령어)
- [에이전트 호출](#에이전트-호출)
- [워크플로우 명령어](#워크플로우-명령어)
- [Git 작업](#git-작업)
- [환경 변수](#환경-변수)
- [단축키](#단축키)

## CLI 명령어

### 프로젝트 관리

```bash
# 프로젝트 초기화
moai init [project-name]                   # Personal 모드
moai init [project-name] --team            # Team 모드
moai init [project-name] --backup          # 백업 생성
moai init [project-name] --force           # 강제 덮어쓰기

# 프로젝트 상태 확인
moai status                                # 기본 상태
moai status -v, --verbose                  # 상세 정보
moai status --trust                        # TRUST 준수율
moai status --config                       # 설정 출력
moai status --project-path /path           # 특정 경로

# 시스템 진단
moai doctor                                # 전체 진단
moai doctor --check-config                 # 설정 검증
moai doctor --verbose                      # 상세 진단
moai doctor --list-backups                 # 백업 목록
```

### 업데이트 및 복원

```bash
# 업데이트
moai update                                # 전체 업데이트
moai update --check                        # 버전 확인만
moai update --package-only                 # 패키지만
moai update --resources-only               # 리소스만
moai update --no-backup                    # 백업 생략
moai update --mode team                    # Team 모드 전환

# 복원
moai restore <backup-path>                 # 백업 복원
moai restore <backup-path> --dry-run       # 미리보기
moai restore <backup-path> --force         # 강제 복원

# 백업 관리
moai backup create                         # 수동 백업
moai backup list                           # 백업 목록
moai backup clean                          # 오래된 백업 정리
```

### 도움말 및 버전

```bash
# 도움말
moai help                                  # 전체 도움말
moai help init                             # 특정 명령어
moai help --all                            # 모든 명령어

# 버전 정보
moai --version                             # 버전 출력
moai -v                                    # 버전 출력 (단축)
```

## 에이전트 호출

### 기본 사용법

```bash
@agent-{name} "요청 내용"
```

### 1. spec-builder

```bash
# 새 SPEC 작성
@agent-spec-builder "사용자 인증 기능"
@agent-spec-builder "회원가입" "로그인" "비밀번호 재설정"

# 기존 SPEC 수정
@agent-spec-builder "SPEC-001에 2FA 추가"

# SPEC 검증
@agent-spec-builder "SPEC-001 EARS 구문 검증"
```

### 2. code-builder

```bash
# 구현 계획 수립
@agent-code-builder "SPEC-001 분석 및 구현 계획"

# TDD 구현
@agent-code-builder "SPEC-001 TDD 구현"
@agent-code-builder "승인된 계획으로 구현 시작"

# 특정 단계만
@agent-code-builder "SPEC-001 RED 단계만"
@agent-code-builder "SPEC-001 REFACTOR 수행"
```

### 3. doc-syncer

```bash
# 전체 동기화
@agent-doc-syncer "문서 동기화 수행"

# 특정 문서만
@agent-doc-syncer "API 문서만 갱신"
@agent-doc-syncer "README 업데이트"

# TAG 검증
@agent-doc-syncer "TAG 체인 검증"
@agent-doc-syncer "TAG 무결성 검사"
```

### 4. cc-manager

```bash
# 설정 최적화
@agent-cc-manager "Claude Code 설정 최적화"

# 권한 설정
@agent-cc-manager "파일 쓰기 권한 조정"

# 훅 관리
@agent-cc-manager "pre-write-guard 훅 활성화"
```

### 5. debug-helper

```bash
# 오류 분석
@agent-debug-helper "TypeError: Cannot read property 'name'"
@agent-debug-helper "Git push 오류 해결"

# 시스템 진단
@agent-debug-helper "시스템 진단 수행"

# TAG 검증
@agent-debug-helper "TAG 체인 검증"
@agent-debug-helper "TAG 무결성 검사"

# 개발 가이드 검사
@agent-debug-helper "개발 가이드 준수 확인"
@agent-debug-helper "TRUST 원칙 검증"
```

### 6. git-manager

```bash
# 브랜치 생성 (사용자 확인 필요)
@agent-git-manager "feature 브랜치 생성"

# 커밋
@agent-git-manager "변경사항 커밋"

# PR 생성 (사용자 확인 필요)
@agent-git-manager "PR 생성"

# 머지 (사용자 확인 필요)
@agent-git-manager "develop 브랜치로 머지"
```

### 7. trust-checker

```bash
# TRUST 검증
@agent-trust-checker "TRUST 원칙 검증"

# 코드 품질 분석
@agent-trust-checker "코드 품질 분석"

# 개선 제안
@agent-trust-checker "품질 개선 제안"
```

## 워크플로우 명령어

### 3단계 개발 사이클

```bash
# Stage 1: SPEC 작성
/moai:1-spec "기능 제목"
/moai:1-spec "기능1" "기능2" "기능3"
/moai:1-spec SPEC-001 "수정 내용"

# Stage 2: TDD 구현
/moai:2-build SPEC-001
/moai:2-build SPEC-001 SPEC-002 SPEC-003
/moai:2-build all

# Stage 3: 문서 동기화
/moai:3-sync
/moai:3-sync full
/moai:3-sync tags-only
/moai:3-sync docs-only
/moai:3-sync --path src/auth
```

### 프로젝트 준비 (선택)

```bash
# 프로젝트 비전 수립
/moai:0-project
```

### 도움말

```bash
# 전체 도움말
/moai:help

# 특정 명령어 도움말
/moai:help 1-spec
/moai:help 2-build
/moai:help 3-sync
```

## Git 작업

### 기본 Git 명령어

```bash
# 상태 확인
git status

# 브랜치 확인
git branch
git branch -a                              # 모든 브랜치

# 변경사항 확인
git diff
git diff --staged

# 커밋 이력
git log
git log --oneline
git log --graph --all
```

### MoAI-ADK Git 워크플로우

```bash
# 1. SPEC 브랜치 생성 (사용자 확인)
/moai:1-spec "기능명"
# → feature/spec-001-feature-name

# 2. 구현 커밋 (자동)
@agent-git-manager "구현 완료 커밋"
# → feat(auth): implement authentication

# 3. PR 생성 (사용자 확인)
@agent-git-manager "PR 생성"
# → Draft PR: feature/spec-001 → develop

# 4. 동기화 후 머지 (사용자 확인)
/moai:3-sync
# → Ready for Review
# → 리뷰어 할당 (Team 모드)
```

### Git 브랜치 명명 규칙

```bash
# Feature 브랜치
feature/spec-{ID}-{description}
feature/task-{description}

# 예시
feature/spec-001-user-auth
feature/task-implement-login
feature/fix-auth-bug
```

## 환경 변수

### 주요 환경 변수

```bash
# MoAI-ADK 설정
export MOAI_MODE=development               # development | production | test
export MOAI_LOG_LEVEL=debug                # debug | info | warn | error

# Claude Code 통합
export CLAUDE_OUTPUT_STYLE=study           # beginner | study | pair | expert | default

# Git 설정
export MOAI_GIT_BRANCH=develop             # 기본 브랜치
export MOAI_REQUIRE_APPROVAL=true          # 승인 필요 여부

# 품질 게이트
export MOAI_MIN_COVERAGE=85                # 최소 커버리지 (%)
export MOAI_MIN_TRUST_SCORE=82             # 최소 TRUST 준수율 (%)

# 경로 설정
export MOAI_CONFIG_PATH=.moai/config.json
export MOAI_BACKUP_PATH=.moai/backups
export MOAI_LOG_PATH=.moai/logs
```

### 환경별 설정

```bash
# 개발 환경
source .env.development
moai doctor

# 프로덕션 환경
source .env.production
moai status

# 테스트 환경
source .env.test
npm test
```

## 단축키

### TAG 검색

```bash
# TAG 전체 검색
rg "@TAG" -n

# 특정 TAG 검색
rg "@REQ:AUTH-001" -n
rg "@TASK:AUTH-001" -n
rg "AUTH-001" -n                           # 모든 관련 TAG

# TAG 타입별 검색
rg "@REQ:" -n                              # 모든 요구사항
rg "@DESIGN:" -n                           # 모든 설계
rg "@TASK:" -n                             # 모든 구현
rg "@TEST:" -n                             # 모든 테스트

# 파일 타입별 검색
rg "@TAG" -g "*.ts"                        # TypeScript 파일만
rg "@TAG" -g "*.py"                        # Python 파일만
rg "@TAG" -g "*.java"                      # Java 파일만
```

### 빠른 명령어 (Bash 별칭)

```bash
# ~/.bashrc 또는 ~/.zshrc에 추가

# MoAI-ADK 별칭
alias mi='moai init'
alias md='moai doctor'
alias ms='moai status'
alias mu='moai update'
alias mr='moai restore'

# SPEC 작성
alias spec='claude /moai:1-spec'

# TDD 구현
alias build='claude /moai:2-build'

# 문서 동기화
alias sync='claude /moai:3-sync'

# TAG 검색
alias tag-search='rg "@TAG" -n'
alias tag-req='rg "@REQ:" -n'
alias tag-task='rg "@TASK:" -n'
alias tag-test='rg "@TEST:" -n'

# 에이전트 호출
alias debug='claude @agent-debug-helper'
alias git-m='claude @agent-git-manager'
alias trust='claude @agent-trust-checker'
```

### 사용 예시

```bash
# 별칭 사용
mi my-project                              # moai init my-project
md                                         # moai doctor
ms -v                                      # moai status --verbose

spec "사용자 인증"                         # /moai:1-spec "사용자 인증"
build SPEC-001                             # /moai:2-build SPEC-001
sync                                       # /moai:3-sync

tag-search                                 # rg "@TAG" -n
debug "오류 분석"                          # @agent-debug-helper "오류 분석"
```

## 자주 사용하는 패턴

### 새 기능 개발

```bash
# 1. SPEC 작성
/moai:1-spec "기능명"

# 2. 브랜치 확인 (사용자 승인 후 생성됨)
git branch

# 3. TDD 구현
/moai:2-build SPEC-{ID}

# 4. 테스트 확인
npm test

# 5. 문서 동기화
/moai:3-sync

# 6. PR 확인 (사용자 승인 후 상태 변경됨)
git log
```

### 디버깅 워크플로우

```bash
# 1. 오류 발생
npm test
# ✗ Test failed

# 2. 디버깅
@agent-debug-helper "테스트 실패 분석"

# 3. 수정
# (코드 수정)

# 4. 재테스트
npm test
# ✓ All tests passed

# 5. 커밋
@agent-git-manager "버그 수정 커밋"
```

### TRUST 검증 워크플로우

```bash
# 1. 현재 상태 확인
moai status --trust

# 2. TRUST 검증
@agent-trust-checker "TRUST 원칙 검증"

# 3. 개선 제안 확인
@agent-trust-checker "품질 개선 제안"

# 4. 개선 작업
# (코드 개선)

# 5. 재검증
moai status --trust
```

## 문제 해결

### 일반적인 문제

```bash
# MoAI-ADK 명령어 인식 안 됨
which moai
npm list -g moai-adk
npm install -g moai-adk

# Claude Code 통합 안 됨
cat .claude/settings.json
moai doctor --check-config

# TAG 검증 실패
/moai:3-sync tags-only
rg "@TAG" -n

# 테스트 실패
npm test
@agent-debug-helper "테스트 실패 분석"

# Git 브랜치 문제
git status
git branch
@agent-git-manager "브랜치 상태 확인"
```

## 다음 단계

### CLI 상세 가이드

- **[moai init](/cli/init)**: 프로젝트 초기화
- **[moai doctor](/cli/doctor)**: 시스템 진단
- **[moai status](/cli/status)**: 프로젝트 상태
- **[moai update](/cli/update)**: 업데이트
- **[moai restore](/cli/restore)**: 복원

### 에이전트 가이드

- **[에이전트](/claude/agents)**: 7개 에이전트 상세
- **[명령어](/claude/commands)**: 워크플로우 명령어
- **[훅](/claude/hooks)**: 이벤트 훅

### 고급 주제

- **[설정 파일](/reference/configuration)**: 전체 설정 옵션
- **[CI/CD 통합](/advanced/ci-cd)**: 자동화 파이프라인

## 참고 자료

- **공식 문서**: https://adk.mo.ai.kr
- **GitHub**: https://github.com/modu-ai/moai-adk
- **NPM**: https://www.npmjs.com/package/@moai/adk
---
title: 프로젝트 초기화
description: moai init 명령어로 프로젝트 구조 생성하기
---

# 프로젝트 초기화

MoAI-ADK로 새 프로젝트를 시작하는 방법을 안내합니다. `moai init` 명령어는 SPEC-First TDD 개발을 위한 완전한 프로젝트 구조를 자동으로 생성합니다.

## moai init 상세 가이드

### 기본 사용법

```bash
# 기본 초기화 (Personal 모드)
moai init my-project

# 대화형 설정
moai init my-project --interactive

# Team 모드 (GitHub Issues/PR 통합)
moai init my-project --team

# 백업 생성 후 초기화
moai init my-project --backup

# 강제 덮어쓰기 (주의!)
moai init my-project --force
```

### Personal vs Team 모드

#### Personal 모드 (기본값)

**특징**:
- 로컬 개발 중심
- `.moai/specs/` 디렉토리에 SPEC 저장
- Git 브랜치 관리 (사용자 확인 필수)
- 단독 개발자 또는 소규모 팀

**장점**:
- 빠른 시작
- 외부 의존성 없음
- 오프라인 작업 가능
- 간단한 설정

**적합한 경우**:
- 개인 프로젝트
- 프로토타입 개발
- 학습 목적
- GitHub 없이 작업

#### Team 모드

**특징**:
- GitHub Issues/PR 통합
- 팀 협업 기능
- 자동 이슈 생성
- PR 리뷰 프로세스

**장점**:
- 전체 팀 가시성
- 코드 리뷰 자동화
- 진행 상황 추적
- 문서화 자동화

**요구사항**:
- GitHub 저장소
- GitHub CLI (`gh`) 설치
- 저장소 쓰기 권한

**적합한 경우**:
- 팀 프로젝트
- 오픈소스 프로젝트
- 엔터프라이즈 개발
- CI/CD 파이프라인 필요

### 초기화 프로세스

`moai init` 실행 시 다음 5단계가 자동으로 진행됩니다:

#### Phase 1: System Verification

시스템 요구사항 자동 검증:

```bash
✓ System Verification Phase
  ✓ Node.js: v18.19.0 (required: >=18.0.0)
  ✓ Git: v2.42.0 (required: >=2.28.0)
  ✓ npm: v10.2.3
  ✓ TypeScript: v5.9.2 (required: >=5.0.0)
  ✓ Bun: v1.2.19 (optional)
```

#### Phase 2: Configuration

프로젝트 설정 파일 생성:

```json
// .moai/config.json
{
  "version": "0.0.1",
  "name": "my-project",
  "mode": "personal",
  "features": {
    "specFirst": true,
    "tddEnforced": true,
    "tagTracking": true
  },
  "backup": {
    "enabled": true,
    "path": ".moai/backups",
    "retention": 7
  }
}
```

#### Phase 3: Template Installation

핵심 디렉토리 및 파일 생성:

```bash
✓ Template Installation Phase
  ✓ Created .moai/ directory
  ✓ Created .claude/ directory
  ✓ Installed 8 agents
  ✓ Installed 5 commands
  ✓ Installed 8 hooks (JavaScript)
  ✓ Created memory files
```

#### Phase 4: Git Integration

Git 저장소 초기화:

```bash
✓ Git Integration Phase
  ✓ Initialized Git repository
  ✓ Created .gitignore
  ✓ Initial commit
  ✓ Created develop branch
```

#### Phase 5: Final Validation

최종 검증 및 확인:

```bash
✓ Final Validation Phase
  ✓ All files created successfully
  ✓ Configuration validated
  ✓ Permissions set correctly
  ✓ Ready to use
```

## 생성된 파일 구조

### 전체 디렉토리 구조

```
my-project/
├── .moai/                      # MoAI-ADK 설정 및 데이터
│   ├── config.json            # 프로젝트 설정
│   ├── memory/                # 개발 가이드
│   │   └── development-guide.md
│   ├── project/               # 프로젝트 문서
│   │   ├── product.md         # 제품 정의
│   │   ├── structure.md       # 구조 설계
│   │   └── tech.md           # 기술 스택
│   ├── specs/                 # SPEC 문서들
│   │   └── (SPEC-XXX/)
│   ├── reports/              # 동기화 리포트
│   │   └── sync-report.md
│   ├── backups/              # 백업 디렉토리
│   └── logs/                 # Winston 로그
│
├── .claude/                   # Claude Code 통합
│   ├── agents/moai/          # 8개 전문 에이전트
│   │   ├── spec-builder.md
│   │   ├── code-builder.md
│   │   ├── doc-syncer.md
│   │   ├── tag-agent.md
│   │   ├── cc-manager.md
│   │   ├── debug-helper.md
│   │   ├── git-manager.md
│   │   └── trust-checker.md
│   │
│   ├── commands/moai/        # 5개 워크플로우 명령어
│   │   ├── 8-project.md
│   │   ├── 1-spec.md
│   │   ├── 2-build.md
│   │   ├── 3-sync.md
│   │   └── help.md
│   │
│   ├── hooks/moai/           # 8개 이벤트 훅 (JavaScript)
│   │   ├── file-monitor.js
│   │   ├── language-detector.js
│   │   ├── policy-block.js
│   │   ├── pre-write-guard.js
│   │   ├── session-notice.js
│   │   ├── steering-guard.js
│   │   ├── run-tests-and-report.js
│   │   └── claude-code-monitor.js
│   │
│   ├── output-styles/        # 5개 출력 스타일
│   │   ├── beginner.md
│   │   ├── study.md
│   │   ├── pair.md
│   │   ├── expert.md
│   │   └── default.md
│   │
│   └── settings.json         # Claude Code 설정
│
├── src/                       # 실제 프로젝트 코드
│   └── (your code here)
│
├── __tests__/                # 테스트 코드
│   └── (your tests here)
│
├── .gitignore                # Git 제외 파일
├── CLAUDE.md                 # 프로젝트 메타 정보
└── README.md                 # 프로젝트 소개
```

### 핵심 파일 설명

#### `.moai/config.json`

프로젝트 전역 설정:

```json
{
  "version": "0.0.1",
  "name": "my-project",
  "description": "A my-project project built with MoAI-ADK",
  "mode": "personal",
  "features": {
    "specFirst": true,
    "tddEnforced": true,
    "tagTracking": true,
    "autoSync": false
  },
  "backup": {
    "enabled": true,
    "path": ".moai/backups",
    "retention": 7,
    "autoBackup": true
  },
  "git": {
    "defaultBranch": "main",
    "developBranch": "develop",
    "featurePrefix": "feature/",
    "requireApproval": true
  },
  "quality": {
    "testCoverage": 85,
    "trustScore": 82,
    "enforceChecks": true
  }
}
```

#### `.moai/memory/development-guide.md`

SPEC-First TDD 개발 가이드 (TRUST 5원칙 포함):

```markdown
# my-project Development Guide

## SPEC-First TDD Workflow

### Core Development Loop (3-Stage)
1. `/moai:1-spec` → 명세 없이는 코드 없음
2. `/moai:2-build` → 테스트 없이는 구현 없음
3. `/moai:3-sync` → 추적성 없이는 완성 없음

### TRUST 5 Principles
- **T**est First: SPEC 기반 TDD
- **R**eadable: 요구사항 주도 가독성
- **U**nified: SPEC 기반 아키텍처
- **S**ecured: 보안 by 설계
- **T**rackable: @TAG 추적성
```

#### `.moai/project/product.md`

제품 정의 템플릿:

```markdown
# my-project Product Definition

## @DOC:MISSION-001 핵심 미션
[프로젝트 미션과 목표]

## @SPEC:USER-001 주요 사용자층
[타겟 사용자 정의]

## @SPEC:PROBLEM-001 해결하는 핵심 문제
[해결할 문제들]

## @SPEC:SUCCESS-001 성공 지표
[측정 가능한 KPI]
```

#### `.claude/settings.json`

Claude Code 설정:

```json
{
  "version": "1.0",
  "project": "my-project",
  "agents": {
    "enabled": true,
    "path": "agents/moai"
  },
  "commands": {
    "enabled": true,
    "path": "commands/moai"
  },
  "hooks": {
    "enabled": true,
    "path": "hooks/moai"
  },
  "outputStyle": "default"
}
```

#### `CLAUDE.md`

프로젝트 메타 정보 (Claude Code가 읽음):

```markdown
# my-project - MoAI Agentic Development Kit

**SPEC-First TDD 개발 가이드**

## 핵심 철학
- **Spec-First**: 명세 없이는 코드 없음
- **TDD-First**: 테스트 없이는 구현 없음
- **GitFlow 지원**: Git 작업 자동화

## 3단계 개발 워크플로우
/moai:1-spec     # 명세 작성
/moai:2-build    # TDD 구현
/moai:3-sync     # 문서 동기화
```

## 초기 설정 체크리스트

프로젝트 초기화 후 다음 항목을 확인하세요:

### 1. 시스템 진단

```bash
cd my-project
moai doctor

# 확인 항목:
# ✓ Node.js 18.0.0+
# ✓ Git 2.28.0+
# ✓ npm/Bun
# ✓ TypeScript 5.0.0+
# ✓ 언어별 도구
```

### 2. Git 저장소 확인

```bash
git status

# 출력 확인:
# On branch develop
# nothing to commit, working tree clean
```

### 3. Claude Code 연동 확인

```bash
# Claude Code에서 확인
claude

# 다음 명령어 테스트:
/moai:help
@agent-debug-helper "시스템 진단"
```

### 4. 첫 SPEC 작성 준비

```bash
# 프로젝트 문서 검토
cat .moai/project/product.md
cat .moai/project/structure.md
cat .moai/project/tech.md

# 첫 SPEC 작성
/moai:1-spec "프로젝트 초기 설정"
```

## 설정 커스터마이징

### Personal → Team 모드 전환

기존 Personal 모드 프로젝트를 Team 모드로 전환:

```bash
# 1. GitHub 저장소 생성
gh repo create my-project --public

# 2. 설정 파일 수정
# .moai/config.json
{
  "mode": "team",
  "github": {
    "owner": "your-username",
    "repo": "my-project",
    "enableIssues": true,
    "enablePR": true
  }
}

# 3. 재초기화 (기존 설정 유지)
moai update --mode team
```

### 백업 설정 조정

```json
// .moai/config.json
{
  "backup": {
    "enabled": true,
    "path": ".moai/backups",
    "retention": 14,              // 14일 보관
    "autoBackup": true,
    "beforeCommit": true,         // 커밋 전 자동 백업
    "compress": true              // 압축 저장
  }
}
```

### 품질 게이트 설정

```json
// .moai/config.json
{
  "quality": {
    "testCoverage": 90,           // 90% 커버리지 요구
    "trustScore": 85,             // TRUST 85점 목표
    "enforceChecks": true,
    "blockCommitOnFailure": true  // 실패 시 커밋 차단
  }
}
```

## 트러블슈팅

### 초기화 실패

**권한 오류**:

```bash
# npm 전역 권한 오류
sudo npm install -g moai-adk

# 또는 nvm 사용
nvm use 18
npm install -g moai-adk
```

**TypeScript 오류**:

```bash
# TypeScript 전역 설치
npm install -g typescript

# 버전 확인
tsc --version  # 5.0.0+
```

**Git 오류**:

```bash
# Git 설정 확인
git config --global user.name
git config --global user.email

# 설정이 없으면
git config --global user.name "Your Name"
git config --global user.email "your@email.com"
```

### 기존 프로젝트에 적용

기존 프로젝트에 MoAI-ADK 추가:

```bash
cd existing-project

# 백업 생성 후 초기화
moai init . --backup

# 충돌 파일 수동 병합
# - CLAUDE.md
# - README.md
# - .gitignore
```

## 다음 단계

### 개발 시작

1. **[빠른 시작](/getting-started/quick-start)**: 5분 튜토리얼
2. **[SPEC 작성](/concepts/spec-first-tdd)**: 첫 SPEC 작성하기
3. **[워크플로우](/concepts/workflow)**: 3단계 개발 사이클

### Claude Code 활용

1. **[에이전트 가이드](/claude/agents)**: 8개 에이전트 활용법
2. **[명령어](/claude/commands)**: 워크플로우 명령어 사용
3. **[훅](/claude/hooks)**: 자동화 시스템 이해

### 언어별 가이드

1. **[TypeScript](/languages/typescript)**: TypeScript 프로젝트
2. **[Python](/languages/python)**: Python 프로젝트
3. **[Java](/languages/java)**: Java 프로젝트

## 참고 자료

- **[moai init 명령어](/cli/init)**: 명령어 상세 레퍼런스
- **[moai doctor](/cli/doctor)**: 시스템 진단 가이드
- **[설정 파일](/reference/configuration)**: 설정 옵션 전체
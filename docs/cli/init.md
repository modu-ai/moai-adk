# moai init - 프로젝트 초기화

**프로젝트를 MoAI-ADK SPEC-First TDD 환경으로 초기화합니다.**

## 개요

`moai init` 명령은 새 프로젝트를 MoAI-ADK 개발 환경으로 초기화하는 가장 기본적이고 중요한 명령어입니다. 이 명령은 `.moai/` 및 `.claude/` 디렉토리 구조를 생성하고, SPEC-First TDD 워크플로우에 필요한 모든 템플릿과 설정 파일을 자동으로 설치합니다.

초기화 과정은 대화형 위저드를 통해 진행되며, 프로젝트 언어를 자동으로 감지하고 해당 언어에 최적화된 개발 도구 구성을 제안합니다. Commander.js 14.0.1 기반으로 구현되어 안정적이고 사용자 친화적인 경험을 제공합니다.

초기화는 기존 프로젝트에도 적용할 수 있으며, `--force` 옵션으로 기존 설정을 덮어쓸 수 있습니다. Personal 모드(로컬 개발)와 Team 모드(GitHub 연동) 중 선택할 수 있어, 개인 프로젝트부터 팀 협업까지 모두 지원합니다.

## 기본 구문

```bash
moai init [project-name] [options]
```

### 위치 인자

- `project-name` (선택): 생성할 프로젝트 이름
  - 생략 시 현재 디렉토리를 초기화
  - 제공 시 해당 이름의 새 디렉토리 생성

### 주요 옵션

| 옵션 | 단축 | 기본값 | 설명 |
|------|------|--------|------|
| `--template <type>` | `-t` | `standard` | 사용할 템플릿 (standard, minimal, advanced) |
| `--interactive` | `-i` | `false` | 대화형 설정 위저드 실행 |
| `--backup` | `-b` | `false` | 설치 전 기존 파일 백업 |
| `--force` | `-f` | `false` | 기존 파일 강제 덮어쓰기 |
| `--personal` | - | `true` | Personal 모드로 초기화 (기본값) |
| `--team` | - | `false` | Team 모드로 초기화 (GitHub 연동) |

## 실제 사용 예제

### 예제 1: 기본 프로젝트 생성

가장 일반적인 사용 방법입니다. 새 디렉토리를 만들고 표준 템플릿으로 초기화합니다.

```bash
# 새 프로젝트 생성
moai init my-awesome-project

# 출력:
# 🗿 MoAI-ADK v0.0.1 - Project Initialization
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#   Step 1: System Verification
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# ✅ Node.js  18.19.0
# ✅ Git      2.42.0
# ✅ npm      10.2.5
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#   Step 2: Configuration
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 📂 Project Name: my-awesome-project
# Detected Language: TypeScript
# 🗿 Mode: Personal
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#   Step 3: Directory Structure
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# ✅ Created .moai/
# ✅ Created .claude/
# ✅ Created src/
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#   Step 4: Template Installation
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# ✅ Installed 7 agents
# ✅ Installed 5 commands
# ✅ Installed 8 hooks
# ✅ Installed project templates
#
# ✅ Project initialized successfully!
#
# Next steps:
# 1. cd my-awesome-project
# 2. Open in Claude Code
# 3. Run: /alfred:1-spec "Your first feature"

# 프로젝트 디렉토리로 이동
cd my-awesome-project
```

### 예제 2: 현재 디렉토리 초기화

기존 프로젝트에 MoAI-ADK를 추가할 때 사용합니다.

```bash
# 기존 프로젝트 디렉토리에서
cd existing-project

# 현재 디렉토리 초기화
moai init

# 언어 자동 감지 후 해당 언어 설정 적용
# Python 프로젝트라면 pytest, mypy, ruff 추천
# TypeScript 프로젝트라면 Vitest, Biome 추천
```

### 예제 3: 대화형 위저드 사용

모든 설정을 단계별로 선택하고 싶을 때 사용합니다.

```bash
moai init my-project --interactive

# 대화형 프롬프트:
# ? 프로젝트 이름: my-project
# ? 주 개발 언어: (자동 감지됨: TypeScript)
#   ◯ TypeScript
#   ◯ Python
#   ◯ Java
#   ◯ Go
#   ◯ Rust
#   ◉ TypeScript (detected)
#
# ? 프로젝트 모드:
#   ◉ Personal (로컬 개발)
#   ◯ Team (GitHub 연동)
#
# ? 템플릿 선택:
#   ◉ Standard (권장)
#   ◯ Minimal (최소 구성)
#   ◯ Advanced (고급 기능 포함)
#
# ? 추가 기능:
#   ☑ CI/CD 템플릿
#   ☑ Docker 설정
#   ☐ VSCode 설정
#   ☑ Git hooks
#
# ✅ 설정 완료! 초기화를 시작합니다...
```

### 예제 4: Team 모드 초기화

GitHub와 연동하여 팀 협업 환경을 구성합니다.

```bash
moai init team-project --team

# Team 모드 추가 설정:
# - GitHub repository 연결 확인
# - GitHub CLI (gh) 설치 확인
# - GitHub Actions 워크플로우 생성
# - Issue 템플릿 및 PR 템플릿 설치
# - Team 협업용 훅 설정
```

### 예제 5: 기존 설정 강제 덮어쓰기

기존 MoAI-ADK 설정을 초기화하거나 업데이트할 때 사용합니다.

```bash
# 백업 생성 후 강제 덮어쓰기
moai init --force --backup

# 경고 메시지:
# ⚠️  Warning: Existing .moai/ directory found
# 📦 Creating backup at .moai.backup-2025-01-15-103000/
# 🔄 Overwriting existing configuration...
# ✅ Backup created: .moai.backup-2025-01-15-103000/
# ✅ Configuration updated successfully!
```

### 예제 6: 팀 모드로 시작

Team 모드는 GitHub 통합을 활성화하여 Issue/PR 기반 협업을 지원합니다.

```bash
moai init my-team-project --team

# Team 모드 특징:
# - GitHub Issue/PR 기반 SPEC 관리
# - 자동 브랜치 생성 및 PR 생성
# - GitFlow 전략 자동 적용
# - 코드 리뷰 워크플로우 지원
```

### 예제 7: 백업과 함께 초기화

기존 설정을 백업하고 새로운 설정으로 초기화합니다.

```bash
moai init . --backup

# 백업 기능:
# - 기존 .moai, .claude 디렉토리 백업
# - 타임스탬프가 포함된 백업 디렉토리 생성
# - 안전한 롤백 지원
```

## 생성되는 디렉토리 구조

초기화 후 생성되는 완전한 프로젝트 구조입니다:

```
my-project/
├── .moai/                          # MoAI-ADK 설정 루트
│   ├── config.json                 # 프로젝트 메인 설정
│   │   {
│   │     "name": "my-project",
│   │     "version": "0.1.0",
│   │     "mode": "personal",
│   │     "language": "typescript",
│   │     "created": "2025-01-15T10:30:00Z"
│   │   }
│   │
│   ├── memory/                     # 개발 가이드 메모리
│   │   └── development-guide.md   # TRUST 5원칙 및 코딩 규칙
│   │
│   ├── specs/                      # SPEC 문서 저장소
│   │   └── .gitkeep               # Git tracking
│   │
│   # TAG는 소스코드에만 존재 (CODE-FIRST)
│   # 별도의 tags/ 폴더 불필요 - 코드 직접 스캔
│   │
│   ├── project/                    # 프로젝트 메타데이터
│   │   ├── product.md             # 제품 정의 (EARS)
│   │   ├── structure.md           # 아키텍처 설계
│   │   └── tech.md               # 기술 스택 정의
│   │
│   └── reports/                   # 동기화 리포트
│       └── .gitkeep
│
├── .claude/                       # Claude Code 통합
│   ├── agents/alfred/               # 7개 전문 에이전트
│   │   ├── spec-builder.md       # SPEC 작성 전담
│   │   ├── code-builder.md       # TDD 구현 전담
│   │   ├── doc-syncer.md         # 문서 동기화
│   │   ├── cc-manager.md         # Claude Code 설정
│   │   ├── debug-helper.md       # 오류 분석
│   │   ├── git-manager.md        # Git 작업 자동화
│   │   └── trust-checker.md      # 품질 검증
│   │
│   ├── commands/alfred/             # 워크플로우 명령어
│   │   ├── 8-project.md          # 프로젝트 초기화
│   │   ├── 1-spec.md            # SPEC 작성
│   │   ├── 2-build.md           # TDD 구현
│   │   └── 3-sync.md            # 문서 동기화
│   │
│   ├── hooks/alfred/                # 이벤트 훅 (JavaScript)
│   │   ├── file-monitor.js       # 파일 변경 감지
│   │   ├── language-detector.js  # 언어 자동 감지
│   │   ├── policy-block.js       # 보안 정책 강제
│   │   ├── pre-write-guard.js    # 쓰기 전 검증
│   │   ├── session-notice.js     # 세션 시작 알림
│   │   └── steering-guard.js     # 방향성 가이드
│   │
│   ├── output-styles/             # 출력 스타일
│   │   ├── beginner.md           # 초보자용
│   │   ├── study.md             # 학습용
│   │   └── pair.md              # 페어 프로그래밍용
│   │
│   └── settings.json              # Claude Code 설정
│
├── src/                           # 소스 코드 (언어별)
│   └── .gitkeep
│
├── tests/                         # 테스트 디렉토리
│   └── .gitkeep
│
├── .gitignore                     # Git 제외 파일
├── .gitattributes                 # Git 속성
└── README.md                      # 프로젝트 README
```

## 템플릿 비교

세 가지 템플릿의 차이점을 이해하고 프로젝트에 맞는 것을 선택하세요.

### Standard 템플릿 (권장)

**대상**: 대부분의 프로젝트

**포함 항목**:
- ✅ 7개 전문 에이전트
- ✅ 3단계 워크플로우 명령어
- ✅ 6개 핵심 훅
- ✅ 3개 출력 스타일
- ✅ TRUST 5원칙 개발 가이드
- ✅ TAG 시스템 (CODE-FIRST, 소스코드 기반)
- ✅ 프로젝트 메타데이터 템플릿

**장점**:
- 완전한 SPEC-First TDD 환경
- 즉시 사용 가능한 모든 기능
- 균형잡힌 구성

### Minimal 템플릿

**대상**: 빠른 프로토타입, 학습용

**포함 항목**:
- ✅ 3개 필수 에이전트 (spec-builder, code-builder, doc-syncer)
- ✅ 3단계 워크플로우 명령어
- ✅ 2개 필수 훅 (policy-block, session-notice)
- ✅ 기본 개발 가이드
- ✅ 기본 TAG 시스템

**장점**:
- 빠른 설치 (< 5초)
- 단순한 구조
- 학습 곡선 완화

**제한**:
- Git 자동화 없음
- 고급 진단 기능 제한
- CI/CD 템플릿 없음

### Advanced 템플릿

**대상**: 엔터프라이즈, 대규모 팀 프로젝트

**포함 항목**:
- ✅ Standard 템플릿의 모든 항목
- ✅ GitHub Actions 워크플로우
- ✅ GitLab CI 설정
- ✅ Docker 및 docker-compose
- ✅ 성능 모니터링 도구
- ✅ 보안 스캐닝 (Snyk, CodeQL)
- ✅ API 문서 자동 생성 (TypeDoc, Sphinx)
- ✅ 릴리즈 자동화

**장점**:
- 프로덕션 준비 완료
- 완전 자동화
- 엔터프라이즈 기능

**고려사항**:
- 복잡한 구성
- 추가 도구 필요 (Docker, GitHub CLI)
- 설치 시간 증가 (~ 30초)

## Personal vs Team 모드

### Personal 모드 (기본값)

**특징**:
- 로컬 Git 저장소 사용
- 브랜치 관리는 수동
- GitHub 연동 없음
- 혼자 개발하기 최적

**워크플로우**:
```bash
/alfred:1-spec "New feature"
# → 로컬 브랜치 생성 (feature/spec-001-new-feature)

/alfred:2-build SPEC-001
# → 로컬에서 TDD 구현

/alfred:3-sync
# → 로컬 문서 업데이트
```

### Team 모드

**특징**:
- GitHub 완전 통합
- Issue 자동 생성
- PR 자동 관리
- 협업 워크플로우 최적화

**요구사항**:
- GitHub repository
- GitHub CLI (`gh`) 설치
- GitHub 인증 완료

**워크플로우**:
```bash
/alfred:1-spec "New feature"
# → GitHub Issue 생성
# → 브랜치 생성 및 연결
# → Draft PR 생성

/alfred:2-build SPEC-001
# → TDD 구현
# → 자동 커밋 및 푸시

/alfred:3-sync
# → 문서 동기화
# → PR 상태: Draft → Ready for Review
# → 리뷰어 자동 할당
```

## 출력 결과 해석

### 성공적인 초기화

```
✅ Project initialized successfully!

📂 Project: my-awesome-project
📁 Location: /Users/you/projects/my-awesome-project
🗿 Mode: Personal
🌐 Language: TypeScript
📦 Template: Standard

📊 Installed Components:
  ✅ Agents: 7/7
  ✅ Commands: 5/5
  ✅ Hooks: 8/8
  ✅ Templates: ✓

🚀 Next steps:
1. cd my-awesome-project
2. Open in Claude Code (VS Code with Claude extension)
3. Run system diagnostics: moai doctor
4. Start first SPEC: /alfred:1-spec "Your first feature"

📚 Documentation: https://adk.mo.ai.kr
💬 Community: https://mo.ai.kr (오픈 예정)
```

### 경고가 있는 초기화

```
⚠️  Warnings detected:

📦 Existing files found:
  - .moai/ (will be skipped, use --force to overwrite)
  - .claude/ (will be merged)

✅ Initialization completed with warnings.

💡 Recommendations:
1. Review existing .moai/config.json
2. Backup important files before using --force
3. Run 'moai doctor' to verify setup
```

## 문제 해결

### 오류 1: 디렉토리가 이미 존재함

```bash
# 오류 메시지:
❌ Error: Directory 'my-project' already exists

# 해결 방법:
# 방법 1: 다른 이름 사용
moai init my-project-v2

# 방법 2: 기존 디렉토리에서 초기화
cd my-project
moai init

# 방법 3: 강제 덮어쓰기 (주의!)
moai init my-project --force --backup
```

### 오류 2: Node.js 버전 불일치

```bash
# 오류 메시지:
❌ Error: Node.js version 16.x detected
Required: Node.js >= 18.0.0

# 해결 방법:
# nvm 사용 시
nvm install 18
nvm use 18

# 또는 공식 사이트에서 다운로드
# https://nodejs.org
```

### 오류 3: 권한 오류 (macOS/Linux)

```bash
# 오류 메시지:
❌ Error: EACCES: permission denied

# 해결 방법:
# 방법 1: npm prefix 변경 (권장)
npm config set prefix ~/.npm-global
export PATH=~/.npm-global/bin:$PATH

# 방법 2: sudo 사용 (비권장)
sudo moai init my-project
```

### 오류 4: Git이 설치되지 않음

```bash
# 오류 메시지:
❌ Error: Git not found
Git is required for MoAI-ADK

# 해결 방법:
# macOS
brew install git

# Ubuntu/Debian
sudo apt-get install git

# Windows
# https://git-scm.com/download/win 에서 다운로드
```

### 오류 5: Team 모드 설정 실패

```bash
# 오류 메시지:
❌ Error: GitHub CLI not found
Team mode requires GitHub CLI (gh)

# 해결 방법:
# GitHub CLI 설치
# macOS
brew install gh

# Ubuntu/Debian
sudo apt install gh

# Windows
winget install GitHub.cli

# 인증
gh auth login
```

## 고급 사용법

### 기존 프로젝트 마이그레이션

기존 프로젝트에 MoAI-ADK를 추가하는 단계별 가이드입니다.

```bash
# 1. 현재 프로젝트 백업
git add .
git commit -m "Backup before MoAI-ADK migration"
git branch backup-$(date +%Y%m%d)

# 2. MoAI-ADK 초기화
moai init --backup

# 3. 시스템 진단
moai doctor

# 4. 설정 커스터마이징
vim .moai/config.json

# 5. Git에 추가
git add .moai/ .claude/
git commit -m "Add MoAI-ADK configuration"
```

### 설정 파일 커스터마이징

생성된 `config.json`을 프로젝트에 맞게 수정할 수 있습니다.

```json
// .moai/config.json
{
  "name": "my-project",
  "version": "0.1.0",
  "mode": "personal",
  "language": "typescript",
  "created": "2025-01-15T10:30:00Z",

  // 커스터마이징 가능 항목
  "features": {
    "autoSync": true,          // 자동 문서 동기화
    "strictTDD": true,          // 엄격한 TDD 강제
    "coverage": {
      "threshold": 85           // 최소 커버리지 %
    }
  },

  "tools": {
    "testRunner": "vitest",     // 언어별 자동 감지
    "linter": "biome",
    "formatter": "biome"
  },

  "git": {
    "autoCommit": false,        // 자동 커밋 비활성화
    "requireApproval": true     // 브랜치 생성 시 승인 요구
  }
}
```

### 다중 언어 프로젝트 설정

프로젝트에서 여러 언어를 사용하는 경우:

```bash
# 1. 주 언어로 초기화
moai init --interactive

# 대화형에서 여러 언어 선택:
# ? 주 개발 언어: TypeScript
# ? 추가 언어:
#   ☑ Python
#   ☑ Go
#   ☐ Java

# 2. 각 언어별 도구 자동 설정
# - TypeScript: Vitest, Biome
# - Python: pytest, mypy, ruff
# - Go: go test, gofmt
```

## 다음 단계

초기화가 완료되면 다음 작업을 진행하세요:

1. **시스템 진단 실행**
   ```bash
   moai doctor
   ```
   → [moai doctor 가이드](/cli/doctor) 참조

2. **프로젝트 상태 확인**
   ```bash
   moai status
   ```
   → [moai status 가이드](/cli/status) 참조

3. **첫 SPEC 작성**
   ```bash
   # Claude Code에서
   /alfred:1-spec "사용자 인증 기능"
   ```
   → [SPEC-First TDD 가이드](/guide/spec-first-tdd) 참조

4. **3단계 워크플로우 학습**
   → [3단계 워크플로우 완전 가이드](/guide/workflow) 참조

## 관련 문서

- [빠른 시작 가이드](/getting-started/quick-start) - 5분 안에 시작하기
- [moai doctor](/cli/doctor) - 시스템 진단
- [설치 가이드](/getting-started/installation) - 상세 설치 방법
- [3단계 워크플로우](/guide/workflow) - SPEC → Build → Sync

---

**다음 읽기**: [moai doctor - 시스템 진단](/cli/doctor)
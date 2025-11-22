# MoAI-ADK 로컬 Claude Code 개발 가이드

> **프로젝트 로컬 설정 파일** | 한글 작업 지침 및 동기화 규칙

**마지막 업데이트**: 2025-11-22
**Claude Code 버전**: 2.0.50
**MoAI-ADK 버전**: v0.26.1

---

## 📋 목차

1. [개발 워크플로우](#개발-워크플로우)
2. [파일 동기화 규칙](#파일-동기화-규칙)
3. [코드 작성 표준](#코드-작성-표준)
4. [Claude Code 설정](#claude-code-설정)
5. [로컬 전용 파일 관리](#로컬-전용-파일-관리)
6. [Git 관리 규칙](#git-관리-규칙)

---

## 개발 워크플로우

### 1.1 작업 위치 규칙

**모든 개발 작업은 다음 위치에서 수행:**

```
/Users/goos/MoAI/MoAI-ADK/src/moai_adk/
├── .claude/                 # Claude Code 설정
├── .moai/                   # MoAI 프로젝트 메타데이터
├── templates/               # 프로젝트 템플릿
└── [여타 소스 코드]
```

**작업 후 로컬 프로젝트로 동기화:**

```
/Users/goos/MoAI/MoAI-ADK/
├── .claude/                 # 동기화됨
├── .moai/                   # 동기화됨
└── [소스 코드 및 문서]
```

### 1.2 개발 사이클

```
1. 소스 프로젝트에서 작업 (/src/moai_adk/...)
   ↓
2. 로컬 프로젝트에 동기화 (./)
   ↓
3. 로컬 프로젝트에서 테스트 및 검증
   ↓
4. Git 커밋 (로컬 프로젝트에서)
```

### 1.3 브랜치 전략

- **main**: 정식 릴리스 (안정성 최고 우선)
- **develop**: 개발 브랜치 (기본 작업 브랜치)
- **feature/SPEC-XXX**: 기능 개발 브랜치
- **release/X.X.X**: 릴리스 브랜치

---

## 파일 동기화 규칙

### 2.1 동기화 대상 디렉토리

**자동 동기화 필요 영역:**

```
src/moai_adk/.claude/    ↔  .claude/
src/moai_adk/.moai/      ↔  .moai/
src/moai_adk/templates/  ↔  src/moai_adk/templates/
```

**정확한 .moai/ 동기화 대상:**

```
src/moai_adk/.moai/config/   ↔  .moai/config/
src/moai_adk/.moai/memory/   ↔  .moai/memory/
src/moai_adk/.moai/scripts/  ↔  .moai/scripts/
```

### 2.2 동기화 제외 (로컬 전용)

**절대 동기화하지 않을 파일:**

```
.claude/commands/moai/99-release.md          # 로컬 릴리스 커맨드만
.claude/settings.local.json                  # 개인 설정
@CLAUDE.local.md                             # 이 파일
.moai/cache/                                 # 캐시 파일
.moai/logs/                                  # 로그 파일
.moai/config/config.json                     # 개인 프로젝트 설정
.moai/analytics/                             # 로컬 분석
.moai/archive/                               # 아카이브
.moai/archived-skills/                       # 아카이브된 스킬
.moai/backups/                               # 백업 파일
.moai/docs/                                  # 로컬 문서
.moai/error_logs/                            # 에러 로그
.moai/indexes/                               # 인덱스
.moai/learning/                              # 학습 자료
.moai/optimization/                          # 최적화 자료
.moai/release/                               # 릴리스 자료
.moai/reports/                               # 보고서
.moai/research/                              # 연구 자료
.moai/specs/                                 # 스펙 문서
.moai/templates/                             # 템플릿 (로컬만)
```

### 2.3 동기화 도구

**사용할 도구:**

```bash
# 수동 동기화 (rsync 사용)
rsync -avz \
  --exclude=".DS_Store" \
  --exclude="*.pyc" \
  --exclude="__pycache__" \
  --exclude=".cache" \
  src/moai_adk/.claude/ .claude/

rsync -avz \
  --exclude=".DS_Store" \
  --exclude="*.pyc" \
  --exclude="__pycache__" \
  src/moai_adk/.moai/config/ .moai/config/
rsync -avz \
  --exclude=".DS_Store" \
  --exclude="*.pyc" \
  src/moai_adk/.moai/memory/ .moai/memory/
rsync -avz \
  --exclude=".DS_Store" \
  --exclude="*.pyc" \
  src/moai_adk/.moai/scripts/ .moai/scripts/
```

### 2.4 스크립트 기반 동기화

**동기화 스크립트 위치:**

```
.moai/scripts/sync-from-src.sh
```

**실행:**

```bash
bash .moai/scripts/sync-from-src.sh
```

---

## 코드 작성 표준

### 3.1 언어 규칙

**모든 코드 작업:**

- ✅ **영문으로만 작성**
- ✅ 변수명: camelCase 또는 snake_case (언어별 관례)
- ✅ 함수명: camelCase (JavaScript/Python) 또는 PascalCase (C#/Java)
- ✅ 클래스명: PascalCase (모든 언어)
- ✅ 상수명: UPPER_SNAKE_CASE (모든 언어)

**주석과 문서:**

- ✅ **모든 주석은 영문**
- ✅ JSDoc, docstring 등 모두 영문
- ✅ Commit messages: 영문 (또는 한글 + 영문 혼용 시 format: 영문)

**이 파일 (@CLAUDE.local.md):**

- ✅ **한글로 작성** (로컬 작업 지침이므로)
- ✅ Git 추적 대상

### 3.2 주석 표준 (영문)

**Good Examples:**

```python
# Initialize the connection pool with specified timeout
def init_connection_pool(timeout: int = 30) -> ConnectionPool:
    """
    Initialize a connection pool for database operations.

    Args:
        timeout: Connection timeout in seconds (default: 30)

    Returns:
        ConnectionPool: Initialized connection pool instance

    Raises:
        ConnectionError: If pool initialization fails
    """
    pass
```

```javascript
/**
 * Fetch user data by ID from the API
 * @param {string} userId - The user's unique identifier
 * @returns {Promise<User>} User object with complete profile
 * @throws {FetchError} If API request fails
 */
async function fetchUser(userId) {
    // Implementation
}
```

### 3.3 금지 사항

```python
# ❌ WRONG - Korean comments
def calculate_score():  # 점수 계산
    score = 100  # 최종 점수
    return score

# ✅ CORRECT - English comments
def calculate_score():  # Calculate final score
    score = 100  # Final score value
    return score
```

---

## Claude Code 설정

### 4.1 프로젝트 설정 파일

**.claude/settings.json (프로젝트 공유 설정):**

```json
{
  "model": "claude-sonnet-4-5-20250929",
  "outputStyle": "R2-D2",
  "cleanupPeriodDays": 30,
  "includeCoAuthoredBy": true,
  "permissions": {
    "defaultMode": "default",
    "allow": [
      "Task",
      "AskUserQuestion",
      "Skill",
      "Read",
      "Write",
      "Edit",
      "MultiEdit",
      "Bash(git:*)",
      "Bash(git status:*)",
      "Bash(git log:*)",
      "Bash(git diff:*)",
      "Bash(ls:*)",
      "Grep",
      "Glob"
    ],
    "ask": [
      "Bash(git add:*)",
      "Bash(git commit:*)",
      "Bash(git push:*)",
      "Bash(rm:*)"
    ],
    "deny": [
      "Bash(rm -rf /:*)",
      "Bash(sudo:*)",
      "Read(./secrets/**)",
      "Read(~/.ssh/**)"
    ]
  }
}
```

### 4.2 로컬 설정 파일

**.claude/settings.local.json (개인 로컬 설정, git ignore):**

```json
{
  "model": "claude-haiku-4-5-20251001",
  "env": {
    "MOAI_WORKSPACE": "/Users/goos/MoAI",
    "MOAI_ADK_ROOT": "/Users/goos/MoAI/MoAI-ADK"
  },
  "statusLine": {
    "type": "command",
    "command": "python .claude/hooks/custom_statusline.py",
    "refreshInterval": 300
  }
}
```

### 4.3 MCP 서버 설정

**로컬 MCP 설정:**

```bash
# Context7 MCP (공식 문서)
claude mcp add --transport http \
  --header "Authorization: Bearer YOUR_TOKEN" \
  context7 https://context7.api.example.com

# Playwright MCP (웹 자동화)
claude mcp add --transport stdio \
  playwright \
  -- npx @anthropic-ai/mcp-server-playwright

# GitHub MCP (저장소 관리)
claude mcp add --transport http \
  github https://github.com/mcp/server
```

### 4.4 Hooks 설정

**.claude/hooks/ 디렉토리 구조:**

```
.claude/hooks/
├── moai/
│   ├── pre_tool__auto_checkpoint.py
│   ├── post_tool__sync_docs.py
│   └── session_start.py
└── custom/
    ├── validate_code.py
    └── lint_check.py
```

**주요 Hook 타입:**

- `SessionStart`: 세션 시작 시 실행
- `SessionEnd`: 세션 종료 시 실행
- `PreToolUse`: 도구 사용 전 실행
- `PostToolUse`: 도구 사용 후 실행

---

## 로컬 전용 파일 관리

### 5.1 로컬 전용 파일 목록

**절대 패키지에 동기화하지 않을 파일:**

| 파일 | 위치 | 용도 | Git 추적 |
|------|------|------|---------|
| `99-release.md` | `.claude/commands/moai/` | 로컬 릴리스 커맨드 | ❌ No |
| `@CLAUDE.local.md` | 루트 | 로컬 개발 지침 | ✅ Yes |
| `settings.local.json` | `.claude/` | 개인 설정 | ❌ No |
| `cache/` | `.moai/` | 캐시 파일 | ❌ No |
| `logs/` | `.moai/` | 로그 파일 | ❌ No |
| `config/config.json` | `.moai/` | 개인 설정 | ❌ No |

### 5.2 로컬 릴리스 커맨드

**.claude/commands/moai/99-release.md (로컬만):**

```markdown
# Local Release Management

This command is only for local development and testing.
It manages MoAI-ADK package releases locally.

## Features

- Version management
- Pre-release testing
- Local deployment simulation
- Changelog generation

## Usage

\`\`\`bash
/moai:99-release
\`\`\`

This command is NOT synchronized to the package.
```

### 5.3 .gitignore 관리

**.gitignore 규칙:**

```
# .moai/ directory management
# Only sync: config/, memory/, scripts/
.moai/
!.moai/config/
!.moai/memory/
!.moai/scripts/

# All other .moai directories are excluded
.moai/cache/
.moai/logs/
.moai/docs/
.moai/reports/
.moai/specs/
.moai/analytics/
.moai/archive/
.moai/archived-skills/
.moai/backups/
.moai/error_logs/
.moai/indexes/
.moai/learning/
.moai/optimization/
.moai/release/
.moai/research/
.moai/templates/

# Claude Code 로컬 설정 (제외)
.claude/settings.local.json
.claude/local/

# @CLAUDE.local.md는 추적 (git에 포함)
!@CLAUDE.local.md
```

---

## Git 관리 규칙

### 6.1 커밋 메시지 포맷

**표준 형식:**

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 목록:**

- `feat`: 새로운 기능
- `fix`: 버그 수정
- `docs`: 문서 변경
- `style`: 코드 포맷팅 (기능 변화 없음)
- `refactor`: 코드 리팩토링
- `perf`: 성능 개선
- `test`: 테스트 추가 또는 수정
- `chore`: 빌드 프로세스, 의존성 등

**예시:**

```
feat(skills): Add moai-domain-iot skill

- Implement MQTT protocol support
- Add edge computing patterns
- Include 15+ code examples

Closes #123
```

### 6.2 브랜치 명명 규칙

```
main/                           # 정식 릴리스
  └─ release/v0.26.1

develop/                        # 개발 브랜치
  └─ feature/SPEC-001
  └─ feature/SPEC-REDESIGN-001
  └─ hotfix/bug-fix-001

feature/SPEC-<ID>               # 기능 개발
  └─ feature/SPEC-04-GROUP-E

hotfix/bug-<ID>                 # 긴급 버그 수정
  └─ hotfix/bug-auth-001
```

### 6.3 Git 작업 체크리스트

**커밋 전:**

- [ ] 모든 코드가 영문으로 작성됨
- [ ] 주석과 docstring이 영문임
- [ ] 로컬 전용 파일이 포함되지 않음
- [ ] 테스트가 통과함
- [ ] Linting이 통과함 (ruff, pylint, etc.)

**푸시 전:**

- [ ] 브랜치가 최신 개발 버전으로 rebase됨
- [ ] 커밋이 논리적 단위로 정리됨
- [ ] 커밋 메시지가 표준 포맷을 따름

**PR 전:**

- [ ] 문서가 동기화됨
- [ ] SPEC이 업데이트됨 (필요시)
- [ ] 변경사항이 설명됨

---

## 자주 사용하는 명령어

### 동기화

```bash
# 소스에서 로컬로 동기화
bash .moai/scripts/sync-from-src.sh

# 특정 디렉토리만 동기화
rsync -avz src/moai_adk/.claude/ .claude/
rsync -avz src/moai_adk/.moai/config/ .moai/config/
rsync -avz src/moai_adk/.moai/memory/ .moai/memory/
rsync -avz src/moai_adk/.moai/scripts/ .moai/scripts/
```

### 검증

```bash
# 코드 품질 확인
ruff check src/
mypy src/

# 테스트 실행
pytest tests/ -v --cov

# 문서 검증
python .moai/tools/validate-docs.py
```

### 개발

```bash
# 새로운 기능 브랜치 시작
git checkout -b feature/SPEC-XXX develop

# 작업 확인
git status
git diff

# 커밋
git add .
git commit -m "feat(scope): description"

# 푸시
git push origin feature/SPEC-XXX
```

---

## 참고 자료

### 공식 문서

- [Claude Code 공식 문서](https://code.claude.com/docs)
- [Claude Code CLI 레퍼런스](https://code.claude.com/docs/en/cli-reference)
- [Claude Code 설정 가이드](https://code.claude.com/docs/en/settings)
- [MCP 통합 가이드](https://code.claude.com/docs/en/mcp)

### MoAI-ADK 문서

- [CLAUDE.md](./CLAUDE.md) - Claude Code 실행 가이드
- [.moai/memory/](./. moai/memory/) - 참고 문서
- [README.md](./README.md) - 프로젝트 개요

### 관련 Skill

- `moai-cc-claude-md` - CLAUDE.md 작성 가이드
- `moai-cc-hooks` - Claude Code Hooks 시스템
- `moai-cc-skills-guide` - Skill 개발 가이드
- `moai-cc-configuration` - 설정 관리 가이드

---

## 업데이트 이력

| 날짜 | 버전 | 변경사항 |
|------|------|---------|
| 2025-11-22 | 1.0.0 | 초기 작성 |
| - | - | - |

---

**작성자**: GOOS님
**프로젝트**: MoAI-ADK
**상태**: ✅ 활성 문서

Translate the following Korean markdown document to Japanese.

**CRITICAL RULES:**
1. Preserve ALL markdown structure (headers, code blocks, links, tables, diagrams)
2. Keep ALL code blocks and technical terms UNCHANGED
3. Maintain the EXACT same file structure and formatting
4. Translate ONLY Korean text content
5. Keep ALL @TAG references unchanged (e.g., @SPEC:AUTH-001)
6. Preserve ALL file paths and URLs
7. Keep ALL emoji and icons as-is
8. Maintain ALL frontmatter (YAML) structure

**Source File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ko/reference/cli/index.md
**Target Language:** Japanese
**Target File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ja/reference/cli/index.md

**Content to Translate:**

# CLI 명령어 참조

" `moai-adk` CLI는 Click 기반의 명령줄 인터페이스로, 프로젝트 관리와 템플릿 동기화를 담당합니다. Alfred 명령어(alfred:\*)와 별개로 로컬 환경
설정과 유지보수에 사용됩니다. "

## 핵심 명령어

"

| 명령어                 | 설명                                              | 활용 시점                          |
| ---------------------- | ------------------------------------------------- | ---------------------------------- |
| `moai-adk init [경로]` | 새 프로젝트 생성 또는 기존 프로젝트에 템플릿 주입 | Alfred를 처음 도입할 때            |
| `moai-adk doctor`      | 환경 점검 (Python, uv, Git, 디렉터리 구조)        | 설치 직후, 문제 발생 시            |
| `moai-adk status`      | TAG 요약, 체크포인트, 템플릿 버전 조회            | 작업 전 상태 파악, 리뷰 전 확인    |
| `moai-adk backup`      | `.moai/`, `.claude/`, CLAUDE.md 백업 생성         | 템플릿 업데이트 전, 대규모 변경 전 |
| `moai-adk update`      | 패키지 & 템플릿 동기화 (가장 중요한 명령)         | 새 버전 릴리스 이후, 정기 점검     |
| "                      |                                                   |                                    |

## 명령어 상세 설명

"

### `moai-adk init`

" **목적**: 새 프로젝트 초기화 및 기본 구조 생성 " **사용법**:

```bash
# 새 프로젝트 생성
moai-adk init my-project
"
# 현재 디렉터리에 초기화
moai-adk init .
"
# 기존 프로젝트에 MoAI-ADK 주입
moai-adk init .
```

" **생성되는 구조**:

```
my-project/
├── .moai/        # 프로젝트 메타데이터
├── .claude/      # Alfred 리소스
└── CLAUDE.md     # 프로젝트 지침
```

" **초기화 과정**:

1. Python 환경 확인
2. Git 저장소 초기화 (없는 경우)
3. `.moai/` 디렉터리 구조 생성
4. `.claude/` 리소스 템플릿 복사
5. 기본 설정 파일 생성 "

### `moai-adk doctor`

" **목적**: 시스템 환경 진단 및 문제 해결 " **사용법**:

```bash
moai-adk doctor
```

" **진단 항목**:

- ✅ Python 버전 (3.13+)
- ✅ uv 패키지 매니저
- ✅ Git 저장소 상태
- ✅ `.moai/` 디렉터리 구조
- ✅ `.claude/` 리소스完整性
- ✅ Claude Code 접근성 " **예상 출력**:

```
🩺 MoAI-ADK System Check
✅ Python 3.13.0
✅ uv 0.5.1
✅ Git 저장소 초기화됨
✅ .moai/ 디렉터리 구조 정상
✅ .claude/ 리소스 74개 로드됨
✅ Claude Code 접근 가능
"
시스템이 정상입니다. Alfred를 시작할 준비가 되었습니다!
```

"

### `moai-adk status`

" **목적**: 프로젝트 현황 요약 및 상태 파악 " **사용법**:

```bash
moai-adk status
```

" **표시 정보**:

- SPEC 진행 현황 (완료/진행중/대기)
- TAG 통계 (@SPEC/@TEST/@CODE/@DOC)
- 최근 체크포인트
- 템플릿 버전 정보
- Git 워크플로우 상태 " **예상 출력**:

```
📊 MoAI-ADK Project Status
:bullseye: Project: MyProject
📅 Last sync: 2025-01-15 14:30
"
📋 SPEC Progress
- ✅ Completed: 12
- 🔄 In Progress: 3
- ⏳ Pending: 5
"
🏷️ TAG Statistics
- @SPEC: 20 tags
- @TEST: 18 tags
- @CODE: 17 tags
- @DOC: 16 tags
- 🚨 Orphan tags: 2
"
📝 Version Info
- Template: v0.15.2
- Last update: 2025-01-10
- Backup available: .moai-backups/20250110/
"
🔄 Git Status
- Current branch: feature/auth-system
- Ahead of main: 12 commits
- Draft PR: #23
```

"

### `moai-adk backup`

" **목적**: 프로젝트 리소스 백업 생성 " **사용법**:

```bash
moai-adk backup
```

" **백업 대상**:

- `.moai/` 전체 디렉터리
- `.claude/` 리소스 템플릿
- `CLAUDE.md` 프로젝트 지침
- Git 상태 정보 " **백업 위치**:

```
.moai-backups/
└── 20250115_143000/
    ├── .moai/
    ├── .claude/
    ├── CLAUDE.md
    └── backup-info.json
```

"

### `moai-adk update`

" **목적**: 패키지와 템플릿 동기화 (가장 중요한 명령) " **사용법**:

```bash
moai-adk update
```

" **업데이트 단계**:

1. **Stage 1**: 패키지 버전 확인
2. **Stage 2**: 템플릿 버전 비교
3. **Stage 3**: 백업 생성 및 병합 " **자동 처리**:

- PyPI에서 최신 버전 확인
- `.moai-backups/`에 현재 리소스 백업
- 새 템플릿과 기존 설정 병합
- 충돌 발생 시 안내 메시지 " **출력 예시**:

```
🔄 MoAI-ADK Update Started
:package: Current version: v0.15.1
:package: Latest version: v0.15.2
"
📁 Creating backup...
✅ Backup created: .moai-backups/20250115_143000/
"
🔄 Updating templates...
🔧 Merging .moai/config.json
🔧 Updating Alfred agents
🔧 Syncing Skills (74 → 77)
"
✅ Update completed successfully!
📝 Changelog: Added moai-domain-ml Skill
⚠️  Please review .claude/settings.json changes
```

"

## 내부 동작 방식

"

### CLI 아키텍처

```
moai-adk
├── __main__.py           # Click 엔트리포인트
├── cli/
│   ├── commands/
│   │   ├── init.py      # 프로젝트 초기화
│   │   ├── doctor.py    # 환경 진단
│   │   ├── status.py    # 상태 조회
│   │   ├── backup.py    # 백업 생성
│   │   └── update.py    # 템플릿 동기화
│   └── utils.py          # 공용 유틸리티
├── core/
│   ├── template.py      # 템플릿 관리
│   ├── backup.py        # 백업/복원
│   └── filesystem.py    # 파일 시스템 조작
└── templates/           # 기본 템플릿 소스
```

"

### Rich 콘솔 출력

- **색상 구분**: 성공 (초록), 경고 (노랑), 오류 (빨강)
- **진행 막대**:长时间 작업의 진행률 표시
- **테이블 형식**: 상태 정보를 정리해서 표시
- **아스키 아트**: 로고 및 구분자 "

### 에러 처리

- **명확한 메시지**: 사용자 친화적인 에러 설명
- **해결 제안**: 문제 해결을 위한 구체적인 방법
- **에러 코드**: 자동화 스크립트를 위한 종료 코드
- **로그 기록**: 문제 추적을 위한 상세 로그 "

## 모범 사례

"

### 정기적인 유지보수

```bash
# 월간 정기 점검
moai-adk doctor
moai-adk status
moai-adk backup
moai-adk update
```

"

### 대규모 변경 전

```bash
# 안전한 변경 절차
moai-adk backup  # 1. 백업 생성
# 변경 작업 수행...
moai-adk status  # 2. 상태 확인
moai-adk doctor  # 3. 환경 점검
```

"

### 새로운 팀원 온보딩

```bash
# 표준 온보딩 절차
git clone <project>
cd <project>
moai-adk doctor  # 환경 확인
moai-adk status  # 프로젝트 이해
claude           # Alfred 시작
/alfred:0-project  # 프로젝트 초기화
```

## "

" **관련 링크**:

- **[프로젝트 구조](project-structure)** - `.moai/`와 `.claude/` 디렉터리 상세
- **[Alfred 명령어](../alfred/commands)** - alfred:\* 워크플로우 명령어들
- **[워크플로우](../workflow)** - CLI와 Alfred의 연동 방식


**Instructions:**
- Translate the content above to Japanese
- Output ONLY the translated markdown content
- Do NOT include any explanations or comments
- Maintain EXACT markdown formatting
- Preserve ALL code blocks exactly as-is

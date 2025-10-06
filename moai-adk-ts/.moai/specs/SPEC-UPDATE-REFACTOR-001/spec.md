---
id: UPDATE-REFACTOR-001
version: 0.1.0
status: draft
created: 2025-10-06
updated: 2025-10-06
author: "@Goos"
priority: medium
category: docs
labels:
  - documentation
  - update-command
  - backup-restore
  - template-merge
scope:
  files:
    - templates/.claude/commands/alfred/9-update.md
    - src/cli/commands/restore.ts
    - templates/.moai/config.json
    - templates/CLAUDE.md
---

# @SPEC:UPDATE-REFACTOR-001: MoAI-ADK 패키지 업데이트 명령어 문서화

## HISTORY

### v0.1.0 (2025-10-06)
- **INITIAL**: /alfred:9-update 명령어 SPEC 최초 작성
- **AUTHOR**: @Goos
- **SCOPE**: 업데이트 절차, 백업/복원 메커니즘, 템플릿 병합 전략
- **CONTEXT**: 기존 9-update.md (v0.2.6) 및 restore.ts 구현 기반 문서화
- **RATIONALE**: Alfred 직접 실행 방식, Claude Code 도구 기반 템플릿 관리, 지능형 병합 시스템

---

## Environment (환경)

### 시스템 컨텍스트

**실행 환경**:
- MoAI-ADK npm 패키지 생태계 (TypeScript 기반)
- Claude Code CLI 환경 (Bash, Read, Write, Grep, Glob 도구)
- Node.js 런타임 (npm 패키지 관리)
- Git 저장소 (버전 관리)

**핵심 컴포넌트**:
- `/alfred:9-update` 명령어 (9-update.md)
- RestoreCommand 클래스 (restore.ts)
- 템플릿 파일 시스템 (node_modules/moai-adk/templates/)
- 백업 디렉토리 구조 (.moai-backup/)

**기존 산출물**:
- `9-update.md`: v0.2.6 (2025-10-06) - Alfred 직접 실행 방식 문서화
- `restore.ts`: CLI 복원 명령어 구현 (백업 검증, 복원 로직)
- 템플릿 파일: 명령어, 에이전트, 훅, Output Styles, 프로젝트 문서, CLAUDE.md, config.json

### 전제 조건

**필수 환경**:
- Node.js 설치 (npm 사용 가능)
- Git 저장소 초기화 완료
- `.moai/`, `.claude/` 디렉토리 존재
- `package.json` 존재 (로컬 설치) 또는 전역 설치 환경

**백업 정책**:
- 자동 백업: `.moai-backup/YYYY-MM-DD-HH-mm-ss/` 구조
- 백업 대상: `.claude/`, `.moai/`, `CLAUDE.md`
- 보존 정책: 수동 삭제 전까지 영구 보존

## Assumptions (가정)

### 1. 안전성 가정
- 백업 메커니즘이 데이터 손실을 방지한다
- 롤백 기능(`moai restore`)이 항상 작동한다
- 템플릿 상태 검증(`{{PROJECT_NAME}}` 패턴)이 정확하다

### 2. 템플릿 최신성
- npm 레지스트리의 moai-adk 패키지가 항상 최신 템플릿을 포함한다
- `npm root`로 조회한 경로가 올바른 템플릿 디렉토리를 가리킨다
- 템플릿 파일 구조가 문서화된 형식을 유지한다

### 3. 지능형 병합 신뢰성
- `{{PROJECT_NAME}}` 패턴 존재 여부로 템플릿 상태를 정확히 판별할 수 있다
- Grep을 통한 프로젝트 정보 추출이 신뢰할 수 있다
- JSON 딥 병합 로직이 사용자 설정을 손실 없이 보존한다

### 4. Alfred 실행 가능성
- Claude Code 도구(Read, Write, Bash, Grep, Glob)만으로 모든 Phase를 수행할 수 있다
- TypeScript 코드 없이 지침 기반 템플릿 복사/병합이 가능하다
- 에러 발생 시 debug-helper가 자동으로 진단 및 복구를 제공한다

### 5. 크로스 플랫폼 호환성
- Windows, macOS, Linux 모두에서 동일하게 작동한다
- `chmod` 실패(Windows)는 경고로 처리되며 계속 진행된다
- 경로 구분자(`/`, `\`) 차이가 자동으로 처리된다

## Requirements (요구사항)

### Ubiquitous (필수 기능)

- 시스템은 MoAI-ADK 패키지를 최신 버전으로 업데이트하는 기능을 제공해야 한다
- 시스템은 업데이트 전 자동 백업을 생성해야 한다 (--force 옵션 제외)
- 시스템은 템플릿 파일과 사용자 설정을 안전하게 병합해야 한다
- 시스템은 업데이트 검증 메커니즘을 제공해야 한다
- 시스템은 롤백 기능(`moai restore`)을 제공해야 한다

### Event-driven (이벤트 기반)

- WHEN 사용자가 `/alfred:9-update`를 실행하면, 시스템은 Phase 1-5를 순차 실행해야 한다
- WHEN `--check` 옵션이 제공되면, 시스템은 버전 확인만 수행하고 중단해야 한다
- WHEN 버전 불일치가 감지되면, 시스템은 업데이트 가능 여부를 보고해야 한다
- WHEN 파일 복사 실패가 발생하면, 시스템은 오류 로그를 기록하고 나머지 파일을 계속 처리해야 한다
- WHEN 검증 실패가 발생하면, 시스템은 재시도 또는 롤백 옵션을 제공해야 한다
- WHEN `--check-quality` 옵션이 제공되면, 시스템은 trust-checker를 호출해야 한다
- WHEN JSON 파싱 실패 시, 시스템은 백업 생성 후 템플릿으로 교체해야 한다 (안전 모드)

### State-driven (상태 기반)

- WHILE 업데이트 진행 중일 때, 시스템은 Phase별 진행 상태를 로깅해야 한다
- WHILE 백업이 존재할 때, 시스템은 롤백 명령어(`moai restore`)를 안내해야 한다
- WHILE 템플릿 상태 파일일 때, 시스템은 전체 덮어쓰기를 수행해야 한다
- WHILE 사용자 수정 파일일 때, 시스템은 보존 또는 지능형 병합을 수행해야 한다

### Optional (선택 기능)

- WHERE 프로젝트 문서(product.md, structure.md, tech.md)가 사용자 수정 상태이면, 시스템은 보존하고 최신 템플릿 참조 경로를 안내할 수 있다
- WHERE `--check` 옵션이 제공되면, 시스템은 업데이트 없이 버전 확인만 수행할 수 있다
- WHERE `--force` 옵션이 제공되면, 시스템은 백업 없이 강제 업데이트를 수행할 수 있다
- WHERE `--check-quality` 옵션이 제공되면, 시스템은 TRUST 5원칙 품질 검증을 수행할 수 있다
- WHERE `--dry-run` 옵션이 제공되면, 시스템은 복원 예상 결과만 표시할 수 있다

### Constraints (제약사항)

- IF `--force` 옵션이 없으면, 시스템은 항상 백업을 생성해야 한다
- IF `{{PROJECT_NAME}}` 패턴이 존재하면, 시스템은 해당 파일을 템플릿 상태로 판단해야 한다
- IF `{{PROJECT_NAME}}` 패턴이 없으면, 시스템은 해당 파일을 사용자 수정 상태로 판단해야 한다
- IF SPEC 파일(`.moai/specs/`)이 존재하면, 시스템은 절대 건드리지 않아야 한다 (완전 보존)
- IF 동기화 리포트(`.moai/reports/`)가 존재하면, 시스템은 보존해야 한다
- IF 파일 복사 실패가 3회 연속 발생하면, 시스템은 해당 파일을 건너뛰고 실패 목록에 추가해야 한다
- IF 버전 검증 실패 시, 시스템은 사용자에게 복구 옵션(Phase 3/4 재실행, 롤백)을 제공해야 한다

## Specifications (상세 명세)

### Phase 1: 버전 확인 및 검증

**목적**: 현재 설치된 버전과 npm 레지스트리 최신 버전 비교

**실행 방식**:
```bash
# 현재 버전 확인
npm list moai-adk --depth=0
# 예상 출력: moai-adk@0.0.1

# 최신 버전 확인
npm view moai-adk version
# 예상 출력: 0.0.2

# 결과 비교
if [현재 버전] < [최신 버전]: "✅ 업데이트 가능"
else: "ℹ️  이미 최신 버전입니다"
```

**조건부 실행**:
- `--check` 옵션이 제공되면 여기서 중단하고 결과만 보고
- 업데이트 가능 시 Phase 2로 진행

**출력 예시**:
```text
🔍 MoAI-ADK 업데이트 확인 중...
📦 현재 버전: v0.0.1
⚡ 최신 버전: v0.0.2
✅ 업데이트 가능
```

### Phase 2: 백업 생성

**목적**: 롤백을 위한 안전 장치 확보

**백업 구조**:
```
.moai-backup/
└── 2025-10-06-15-30-00/
    ├── .claude/         # 모든 하위 파일
    ├── .moai/           # 모든 하위 파일
    └── CLAUDE.md        # 루트 파일
```

**실행 방식**:
```bash
BACKUP_DIR=".moai-backup/$(date +%Y-%m-%d-%H-%M-%S)"
mkdir -p "$BACKUP_DIR"
cp -r .claude .moai CLAUDE.md "$BACKUP_DIR/" 2>/dev/null || true
```

**예외 처리**:
- `--force` 옵션이 제공되면 백업 건너뛰기
- 파일 없음 오류는 무시하고 계속 진행
- 디스크 공간 부족 시 오류 보고 및 중단

**출력 예시**:
```text
💾 백업 생성 중...
   → .moai-backup/2025-10-06-15-30-00/
   ✅ 백업 완료
```

### Phase 3: npm 패키지 업데이트

**목적**: moai-adk npm 패키지를 최신 버전으로 설치

**실행 방식**:
```bash
if [ -f "package.json" ]; then
    # 로컬 프로젝트 의존성으로 설치
    npm install moai-adk@latest
else
    # 전역 설치
    npm install -g moai-adk@latest
fi
```

**검증**:
```bash
npm list moai-adk --depth=0
# 최신 버전 확인
```

**출력 예시**:
```text
📦 패키지 업데이트 중...
   npm install moai-adk@latest
   ✅ 패키지 업데이트 완료 (v0.0.2)
```

### Phase 4: 템플릿 복사/병합 (Alfred 직접 실행)

**담당**: Alfred (직접 실행, 에이전트 위임 없음)
**도구**: [Bash], [Glob], [Read], [Grep], [Write]

#### 처리 전략 요약

| 대상 | 전략 | 이유 |
|------|------|------|
| `.moai/specs/` | 완전 보존 🔒 | 사용자 SPEC 절대 건드리지 않음 |
| `.moai/reports/` | 완전 보존 🔒 | 동기화 리포트 보존 |
| 시스템 파일 (명령어, 에이전트, 훅, Output Styles, 가이드) | 전체 교체 ✅ | 항상 최신 유지 |
| 프로젝트 문서 (product, structure, tech) | 사용자 수정 시 보존 🔒 | `{{PROJECT_NAME}}` 패턴 없으면 보존 |
| `CLAUDE.md` | 지능형 병합 🔄 | 템플릿 구조 + 프로젝트 정보 유지 |
| `.moai/config.json` | 스마트 딥 병합 🔄 | 템플릿 필드 + 사용자 설정값 유지 |

#### Step 1: npm root 확인

**목적**: 템플릿 파일이 위치한 node_modules 경로 확인

```bash
[Bash] npm root
# 출력 예시: /Users/user/project/node_modules
```

**템플릿 경로 설정**:
```
TEMPLATE_ROOT="{npm_root}/moai-adk/templates"
```

**검증**:
- 경로가 비어있지 않은지 확인
- 템플릿 디렉토리 존재 여부 확인 (`ls {TEMPLATE_ROOT}`)

#### Step 2-6: 시스템 파일 전체 교체 ✅

**대상 파일 목록**:
1. `.claude/commands/alfred/*.md` (~10개)
2. `.claude/agents/alfred/*.md` (~9개)
3. `.claude/hooks/alfred/*.cjs` (~4개)
4. `.claude/output-styles/alfred/*.md` (4개)
5. `.moai/memory/development-guide.md` (1개)

**처리 절차** (각 카테고리마다 반복):
```text
[Step A] 템플릿 파일 검색
  → [Glob] "{TEMPLATE_ROOT}/.claude/commands/alfred/*.md"
  → 결과: [1-spec.md, 2-build.md, 3-sync.md, ...]

[Step B] 각 파일 복사
  FOR EACH file IN glob_results:
    a. [Read] "{TEMPLATE_ROOT}/.claude/commands/alfred/{file}"
    b. [Write] ".claude/commands/alfred/{file}"
    c. 성공 로그: "✅ {file}"

[Step C] 완료 메시지
  → "✅ .claude/commands/alfred/ (~10개 파일 복사 완료)"
```

**오류 처리**:
- Glob 결과 비어있음 → "⚠️ 템플릿 디렉토리 경로 확인 필요"
- Write 실패 → `[Bash] mkdir -p` 후 재시도 (최대 2회)
- 재시도 후에도 실패 → 실패 목록에 추가, 계속 진행

**특수 처리 - 훅 파일 권한**:
```bash
[Bash] chmod +x .claude/hooks/alfred/*.cjs
# Unix 계열: 실행 권한 부여 (755)
# Windows: 실패 무시 (경고만 출력)
```

#### Step 7-9: 프로젝트 문서 사용자 수정 시 보존 🔒

**대상 파일**:
- `.moai/project/product.md`
- `.moai/project/structure.md`
- `.moai/project/tech.md`

**처리 절차** (각 파일마다 반복):
```text
[Step 1] 기존 파일 존재 확인
  → [Read] ".moai/project/product.md"
  → IF 파일 없음: Step 5로 이동 (새로 생성)
  → IF 파일 있음: Step 2 진행

[Step 2] 템플릿 상태 검증
  → [Grep] "{{PROJECT_NAME}}" -n ".moai/project/product.md"
  → IF 검색 결과 있음: 템플릿 상태 (Step 5로 이동 - 덮어쓰기)
  → IF 검색 결과 없음: 사용자 수정 상태 (Step 3 진행 - 보존)

[Step 3] 사용자 작업물 보존 🔒
  → "ℹ️  product.md는 이미 프로젝트 정보가 작성되어 있어서 건너뜁니다"
  → "💡 최신 템플릿 참조: {TEMPLATE_ROOT}/.moai/project/product.md"
  → "📝 필요시 수동으로 새 필드 추가 가능"
  → 복사하지 않고 다음 파일로 이동 (Step 6)

[Step 4] (예비 - 현재 미사용)

[Step 5] 새 템플릿 복사 (템플릿 상태 또는 파일 없음)
  → [Read] "{TEMPLATE_ROOT}/.moai/project/product.md"
  → [Write] ".moai/project/product.md"
  → IF 실패: [Bash] mkdir -p .moai/project 후 재시도

[Step 6] 완료 메시지
  → 템플릿 상태: "✅ .moai/project/product.md (템플릿 → 최신 버전)"
  → 사용자 수정: "⏭️  .moai/project/product.md (사용자 작업물 보존)"
  → 새로 생성: "✨ .moai/project/product.md (새로 생성)"
```

**핵심 보호 정책**:
- `{{PROJECT_NAME}}` 패턴 존재 → 템플릿 상태, 안전하게 덮어쓰기
- 패턴 없음 → 사용자 수정, 보존 (덮어쓰지 않음) 🔒
- 파일 없음 → 새로 생성

#### Step 10: CLAUDE.md 지능형 병합 🔄

**목적**: 템플릿 최신 구조 + 프로젝트 정보 유지

**처리 절차**:
```text
[Step 1] 기존 파일 존재 확인
  → [Read] "./CLAUDE.md"
  → IF 파일 없음: Step 6으로 이동 (새로 생성)
  → IF 파일 있음: Step 2 진행

[Step 2] 템플릿 상태 검증
  → [Grep] "{{PROJECT_NAME}}" -n "./CLAUDE.md"
  → IF 검색 결과 있음: 템플릿 상태 (Step 6으로 이동 - 덮어쓰기)
  → IF 검색 결과 없음: 사용자 수정 상태 (Step 3 진행 - 병합)

[Step 3] 사용자 프로젝트 정보 추출
  → [Grep] "^- \*\*이름\*\*:" "./CLAUDE.md" → project_name
  → [Grep] "^- \*\*설명\*\*:" "./CLAUDE.md" → project_description
  → [Grep] "^- \*\*버전\*\*:" "./CLAUDE.md" → project_version
  → [Grep] "^- \*\*모드\*\*:" "./CLAUDE.md" → project_mode
  → "📋 추출된 정보: {project_name} v{project_version} ({project_mode})"

[Step 4] 최신 템플릿 읽기
  → [Read] "{TEMPLATE_ROOT}/CLAUDE.md"
  → 템플릿 내용을 메모리에 저장

[Step 5] 템플릿에 사용자 정보 주입 (병합)
  → {{PROJECT_NAME}} → {project_name}
  → {{PROJECT_DESCRIPTION}} → {project_description}
  → {{PROJECT_VERSION}} → {project_version}
  → {{PROJECT_MODE}} → {project_mode}
  → [Write] "./CLAUDE.md" (병합된 내용)
  → "🔄 CLAUDE.md 병합 완료 (템플릿 최신화 + 프로젝트 정보 유지)"

[Step 6] 새 템플릿 복사 (템플릿 상태 또는 파일 없음)
  → [Read] "{TEMPLATE_ROOT}/CLAUDE.md"
  → [Write] "./CLAUDE.md"

[Step 7] 완료 메시지
  → 템플릿 상태: "✅ CLAUDE.md (템플릿 → 최신 버전)"
  → 병합: "🔄 CLAUDE.md (템플릿 최신화 + 프로젝트 정보 유지)"
  → 새로 생성: "✨ CLAUDE.md (새로 생성)"
```

**병합 정책**:
- 템플릿 최신 구조 사용 (Alfred 에이전트 목록, 워크플로우 등)
- 프로젝트 정보만 유지 (이름, 설명, 버전, 모드)

#### Step 11: config.json 스마트 딥 병합 🔄

**목적**: 템플릿 최신 구조 + 사용자 설정값 유지

**처리 절차**:
```text
[Step 1] 기존 파일 존재 확인
  → [Read] ".moai/config.json"
  → IF 파일 없음: Step 7로 이동 (새로 생성)
  → IF 파일 있음: Step 2 진행

[Step 2] 템플릿 상태 검증
  → [Grep] "{{PROJECT_NAME}}" -n ".moai/config.json"
  → IF 검색 결과 있음: 템플릿 상태 (Step 7로 이동 - 덮어쓰기)
  → IF 검색 결과 없음: 사용자 설정 상태 (Step 3 진행 - 병합)

[Step 3] 사용자 설정값 추출 (JSON 파싱)
  → project.* → user_project_* (모든 필드)
  → constitution.test_coverage_target → user_test_coverage
  → constitution.simplicity_threshold → user_simplicity_threshold
  → git_strategy.* → user_git_strategy (전체 객체)
  → tags.categories → user_tag_categories (사용자 추가 카테고리)
  → pipeline.current_stage → user_current_stage
  → "📋 추출 완료: {user_project_name} v{user_project_version}"

[Step 4] 최신 템플릿 읽기
  → [Read] "{TEMPLATE_ROOT}/.moai/config.json"
  → 템플릿 JSON을 메모리에 파싱

[Step 5] 딥 병합 (Deep Merge)
  병합 의사 코드:

  const merged = {
    // project: 사용자 값 100% 유지
    project: userConfig.project,

    // constitution: 템플릿 + 사용자 수정값 덮어쓰기
    constitution: {
      ...templateConfig.constitution,
      ...userConfig.constitution
    },

    // git_strategy: 사용자 값 100% 유지
    git_strategy: userConfig.git_strategy,

    // tags.categories: 배열 병합 (중복 제거)
    tags: {
      ...templateConfig.tags,
      categories: [
        ...new Set([
          ...templateConfig.tags.categories,
          ...userConfig.tags.categories
        ])
      ]
    },

    // pipeline: 템플릿 + current_stage만 사용자 값
    pipeline: {
      ...templateConfig.pipeline,
      current_stage: userConfig.pipeline.current_stage
    },

    // _meta: 템플릿 최신
    _meta: templateConfig._meta
  };

[Step 6] 병합 결과 저장
  → [Write] ".moai/config.json" (JSON 직렬화, 들여쓰기 2칸)
  → "🔄 config.json 병합 완료"

[Step 7] 새 템플릿 복사 (템플릿 상태 또는 파일 없음)
  → [Read] "{TEMPLATE_ROOT}/.moai/config.json"
  → [Write] ".moai/config.json"

[Step 8] 완료 메시지
  → 템플릿 상태: "✅ config.json (템플릿 → 최신 버전)"
  → 병합: "🔄 config.json (스마트 병합: 템플릿 구조 + 사용자 설정)"
  → 새로 생성: "✨ config.json (새로 생성)"
```

**필드별 병합 정책**:

| 필드 | 병합 전략 | 이유 |
|------|-----------|------|
| `project.*` | 사용자 값 100% 유지 | 프로젝트 식별 정보 |
| `constitution.test_coverage_target` | 사용자 값 유지 | 팀 정책 |
| `constitution.simplicity_threshold` | 사용자 값 유지 | 팀 정책 |
| `git_strategy.*` | 사용자 값 100% 유지 | 워크플로우 설정 |
| `tags.categories` | 병합 (템플릿 + 사용자) | 확장 가능 |
| `pipeline.available_commands` | 템플릿 최신 | 시스템 명령어 |
| `pipeline.current_stage` | 사용자 값 유지 | 진행 상태 |
| `_meta.*` | 템플릿 최신 | TAG 참조 |

**안전 모드 (JSON 파싱 실패 시)**:
```text
IF JSON 파싱 실패:
  → [Bash] cp .moai/config.json .moai/config.json.backup
  → "⚠️ config.json 파싱 실패, 백업 생성: .moai/config.json.backup"
  → [Read] "{TEMPLATE_ROOT}/.moai/config.json"
  → [Write] ".moai/config.json" (템플릿으로 교체)
  → "📝 수동 복구 필요: .moai/config.json.backup 참조"
```

#### 전체 오류 처리 원칙

**독립적 오류 처리**:
- 각 Step별 독립적 오류 처리
- 한 파일 실패가 전체 프로세스를 중단시키지 않음
- 실패한 파일 목록 수집하여 Phase 4 종료 후 보고

**자동 복구 전략**:

| 오류 유형 | 복구 조치 |
|-----------|-----------|
| 파일 누락 | Phase 4 재실행 제안 |
| 디렉토리 없음 | `mkdir -p` 자동 실행 후 재시도 |
| Write 실패 | 최대 2회 재시도 |
| 권한 오류 (Unix) | `chmod +x` 재실행 |
| 권한 오류 (Windows) | 경고 출력 후 계속 진행 |
| JSON 파싱 실패 | 백업 생성 후 템플릿으로 교체 (안전 모드) |

**출력 예시**:
```text
📄 Phase 4: Alfred가 템플릿 복사/병합 중...
   ✅ .claude/commands/alfred/ (~10개 파일 복사 완료)
   ✅ .claude/agents/alfred/ (~9개 파일 복사 완료)
   ✅ .claude/hooks/alfred/ (4개 파일 복사 + 권한 설정 완료)
   ✅ .claude/output-styles/alfred/ (4개 파일 복사 완료)
   ✅ .moai/memory/development-guide.md (무조건 업데이트)
   ✅ .moai/project/product.md (템플릿 → 최신 버전)
   ⏭️  .moai/project/structure.md (사용자 작업물 보존)
   ✨ .moai/project/tech.md (새로 생성)
   🔄 CLAUDE.md (템플릿 최신화 + 프로젝트 정보 유지)
   🔄 .moai/config.json (스마트 병합: 템플릿 구조 + 사용자 설정)
   ✅ 템플릿 파일 처리 완료
```

### Phase 5: 업데이트 검증

**담당**: Alfred (직접 실행)
**도구**: [Bash], [Glob], [Read], [Grep]

#### 5.1 파일 개수 검증 (동적)

**목적**: 템플릿 파일이 모두 복사되었는지 확인

```text
[Check 1] 명령어 파일
  → [Glob] .claude/commands/alfred/*.md
  → 실제 개수 확인
  → 예상: ~10개
  → IF 실제 < 예상: "⚠️ 명령어 파일 누락 감지"

[Check 2] 에이전트 파일
  → [Glob] .claude/agents/alfred/*.md
  → 예상: ~9개

[Check 3] 훅 파일
  → [Glob] .claude/hooks/alfred/*.cjs
  → 예상: ~4개

[Check 4] Output Styles 파일
  → [Glob] .claude/output-styles/alfred/*.md
  → 예상: 4개

[Check 5] 프로젝트 문서
  → [Glob] .moai/project/*.md
  → 예상: 3개

[Check 6] 필수 파일 존재
  → [Read] .moai/memory/development-guide.md
  → [Read] CLAUDE.md
  → [Read] .moai/config.json
  → IF 파일 없음: "❌ 필수 파일 누락"
```

**누락 감지 시**:
- Phase 4 재실행 제안
- 실패 파일 목록 표시

#### 5.2 YAML Frontmatter 검증

**목적**: 명령어 파일의 YAML 구문 손상 여부 확인

```text
[Sample Check] 명령어 파일 검증
  → [Read] .claude/commands/alfred/1-spec.md (첫 10줄)
  → YAML 파싱 시도 (---로 감싸진 블록)
  → IF 파싱 실패: "⚠️ YAML frontmatter 손상 감지"
  → IF 파싱 성공: "✅ YAML 검증 통과"

[필수 필드 확인]
  - name: alfred:1-spec
  - description: (내용 확인)
  - tools: [Read, Write, ...]
```

**검증 방식**:
- 대표 파일 1-2개 샘플링
- YAML 구문 오류 감지
- 손상 시 Phase 4 재실행 제안

#### 5.3 버전 정보 확인

**목적**: 패키지 버전과 템플릿 버전의 일치 여부 확인

```text
[Check 1] development-guide.md 버전
  → [Grep] "version:" -n .moai/memory/development-guide.md
  → 버전 추출: v{X.Y.Z}

[Check 2] package.json 버전
  → [Bash] npm list moai-adk --depth=0
  → 출력: moai-adk@{version}

[Check 3] 버전 일치 확인
  → IF 일치: "✅ 버전 정합성 통과 (v{X.Y.Z})"
  → IF 불일치: "⚠️ 버전 불일치 감지"
```

**버전 불일치 시**:
- Phase 3 재실행 제안 (npm 재설치)
- 또는 Phase 4 재실행 (템플릿 재복사)

#### 5.4 훅 파일 권한 검증 (Unix 계열만)

**목적**: 실행 권한 확인

```bash
[Bash] ls -l .claude/hooks/alfred/*.cjs
# 예상 출력: -rwxr-xr-x (755)
```

**권한 확인**:
- 실행 권한 (`x`) 존재 여부 확인
- Windows 환경은 검증 생략

**권한 없음 시**:
```bash
[Bash] chmod +x .claude/hooks/alfred/*.cjs
```

#### 검증 실패 시 자동 복구 전략

| 오류 유형 | 복구 조치 |
|-----------|-----------|
| 파일 누락 | Phase 4 재실행 제안 |
| 버전 불일치 | Phase 3 재실행 제안 (npm) |
| 내용 손상 | 백업 복원 후 재시작 제안 |
| 권한 오류 | `chmod` 재실행 |
| 디렉토리 없음 | `mkdir -p` 후 Phase 4 재실행 |

**출력 예시**:
```text
🔍 검증 중...
   ✅ 파일 개수 검증 통과
   ✅ YAML 구문 검증 통과
   ✅ 버전 정합성 통과 (v0.0.2)
   ✅ 훅 파일 권한 확인 (755)
   ✅ 검증 완료
```

### Phase 5.5: 품질 검증 (선택적)

**조건**: `--check-quality` 옵션 제공 시에만 실행

**담당**: trust-checker 에이전트

**실행 절차**:
```text
[Step 1] trust-checker 호출
  → @agent-trust-checker "Level 1 빠른 스캔 (3-5초)"
  → 검증 항목: TRUST 5원칙

[Step 2] 검증 항목
  - **T**est: 테스트 커버리지 ≥85% 확인
  - **R**eadable: ESLint/Biome 통과 여부
  - **U**nified: TypeScript 타입 안전성
  - **S**ecured: npm audit 보안 취약점 검사
  - **T**rackable: @TAG 체인 무결성 검증

[Step 3] 결과별 처리
  ✅ Pass: 업데이트 성공 완료
    → "✅ 품질 검증 통과"
    → "- 모든 파일 정상"
    → "- 시스템 무결성 유지"

  ⚠️ Warning: 경고 표시 후 완료
    → "⚠️ 품질 검증 경고"
    → "- 일부 문서 포맷 이슈 발견"
    → "- 권장사항 미적용 항목 존재"
    → 사용자 확인 권장 (계속 진행 가능)

  ❌ Critical: 롤백 제안
    → "❌ 품질 검증 실패 (치명적)"
    → "- 파일 손상 감지"
    → "- 설정 불일치"
    → 조치 선택:
      1. "롤백" → moai restore .moai-backup/{timestamp}
      2. "무시하고 진행" → 손상된 상태로 완료 (위험)
    → 권장: 롤백 후 재시도
```

**실행 시간**: 추가 3-5초 (Level 1 빠른 스캔)

**검증 생략**:
- `--check-quality` 옵션 없으면 Phase 5.5 건너뛰고 완료

**출력 예시**:
```text
🔍 품질 검증 중... (trust-checker)
   ✅ 테스트 커버리지 88% (목표: 85%)
   ✅ 린트 검사 통과
   ✅ 타입 안전성 확인
   ✅ 보안 취약점 없음
   ✅ TAG 체인 무결성 확인
   ✅ 품질 검증 통과
```

## Traceability (@TAG)

### TAG 체인

- **SPEC**: @SPEC:UPDATE-REFACTOR-001 (본 문서)
- **DOC**: templates/.claude/commands/alfred/9-update.md (@DOC:UPDATE-001)
- **CODE**: src/cli/commands/restore.ts (@CODE:CLI-005)
- **CODE**: templates/.moai/config.json (@CODE:CFG-001)
- **DOC**: templates/CLAUDE.md (템플릿)

### 관련 SPEC

- @SPEC:BACKUP-VALIDATION-001 (restore.ts 백업 검증 로직)
- @SPEC:RESTORE-OPTIONS-001 (restore.ts 옵션 구조)
- @SPEC:RESTORE-RESULT-001 (restore.ts 결과 구조)

### 의존성

- **상위 의존**: 없음 (독립 실행)
- **하위 의존**: trust-checker 에이전트 (선택적)

### 연관 문서

- `.moai/memory/development-guide.md` - TRUST 5원칙, @TAG 시스템
- `.moai/project/tech.md` - npm 패키지 관리, 배포 전략
- `.moai/project/structure.md` - 템플릿 파일 구조

---

## 성공 기준

### Definition of Done

- ✅ YAML Front Matter 완전성 (필수 필드 7개)
- ✅ HISTORY 섹션 작성 (v0.1.0 INITIAL)
- ✅ EARS 구조 완전성 (Environment, Assumptions, Requirements, Specifications)
- ✅ Phase 1-5 상세 명세 작성
- ✅ 템플릿 병합 전략 문서화
- ✅ 오류 복구 시나리오 정의
- ✅ @TAG 체인 연결 (9-update.md, restore.ts)

### 검증 체크리스트

- [ ] 기존 9-update.md의 모든 Phase가 SPEC에 반영되었는가?
- [ ] restore.ts의 백업/복원 로직이 SPEC에 포함되었는가?
- [ ] 지능형 병합 전략(CLAUDE.md, config.json)이 명확히 정의되었는가?
- [ ] 오류 시나리오별 복구 절차가 완전한가?
- [ ] trust-checker 연동이 선택적으로 정의되었는가?

## 구현 우선순위

### 1차 목표 (현재 완료 상태 확인)
- 기존 9-update.md 문서 검증
- restore.ts 구현 검증
- 템플릿 파일 구조 확인

### 2차 목표 (개선 가능 영역)
- Alfred 실행 흐름 최적화
- 오류 복구 자동화 개선
- 사용자 경험 향상 (진행률 표시 등)

### 3차 목표 (확장 기능)
- 백업 자동 정리 (오래된 백업 삭제)
- 버전별 마이그레이션 가이드 자동 표시
- 업데이트 이력 추적 (changelog 자동 생성)

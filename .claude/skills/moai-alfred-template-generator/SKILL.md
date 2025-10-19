---
name: moai-alfred-template-generator
description: 프로젝트 맞춤형 CLAUDE.md 및 기능 파일 생성 (commands, agents, skills 선택 복사)
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Grep
  - Glob
---

# Template Generator Skill

## 🎯 목적

`.moai/.feature-selection.json` 기반으로 **프로젝트 맞춤형 CLAUDE.md** 및 **선택된 commands/agents/skills만** 복사합니다.

**핵심 가치**: 선택적 복사 → 37개 스킬에서 3~5개로 경량화 → 사용자 학습 곡선 감소

---

## 📥 입력

- `.moai/.feature-selection.json` (feature-selector 출력)
- CLAUDE.md 템플릿 (`src/moai_adk/templates/CLAUDE.md`)
- 전체 commands/agents/skills 파일 (`src/moai_adk/templates/.claude/`)

---

## 📤 출력

**프로젝트 루트 디렉토리**:
- `CLAUDE.md` (맞춤형 에이전트 테이블)
- `.claude/commands/` (선택된 commands만)
- `.claude/agents/` (선택된 agents만)
- `.claude/skills/` (선택된 skills만)
- `.moai/config.json` (optimized: true 업데이트)

---

## 🔧 실행 로직

### STEP 1: Feature Selection 결과 읽기

**목적**: feature-selector가 생성한 JSON 파일 로드

**실행**:
```bash
# JSON 파일 읽기
cat .moai/.feature-selection.json
```

**예시 결과**:
```json
{
  "category": "web-api",
  "language": "python",
  "framework": "fastapi",
  "commands": ["1-spec", "2-build", "3-sync"],
  "agents": ["spec-builder", "code-builder", "doc-syncer", "git-manager", "debug-helper"],
  "skills": ["moai-lang-python", "moai-domain-web-api", "moai-domain-backend"]
}
```

---

### STEP 2: 에이전트 메타데이터 추출

**목적**: 선택된 agents의 YAML frontmatter에서 메타데이터 추출

**실행**:
```bash
# moai-adk 패키지 경로 찾기
NPM_ROOT=$(npm root -g 2>/dev/null || echo "$HOME/.npm-global/lib/node_modules")
TEMPLATE_DIR="${NPM_ROOT}/moai-adk/templates/.claude/agents"

# 선택된 에이전트 메타데이터 추출
for agent in spec-builder code-builder doc-syncer git-manager debug-helper; do
  echo "=== $agent ==="
  grep -A 3 "^name:" "${TEMPLATE_DIR}/${agent}.md" | head -4
done
```

**추출 데이터**:
```yaml
name: spec-builder
description: SPEC 작성 (EARS 방식)
model: sonnet
persona: 시스템 아키텍트
```

---

### STEP 3: 동적 에이전트 테이블 생성

**목적**: CLAUDE.md 템플릿의 에이전트 테이블을 선택된 에이전트만 포함하도록 생성

**에이전트 테이블 템플릿**:
```markdown
| 에이전트              | 모델   | 페르소나          | 전문 영역               | 커맨드/호출            | 위임 시점      |
| --------------------- | ------ | ----------------- | ----------------------- | ---------------------- | -------------- |
```

**매핑 규칙**:

| Agent Name       | 모델   | 페르소나          | 전문 영역               | 커맨드/호출            | 위임 시점      |
| ---------------- | ------ | ----------------- | ----------------------- | ---------------------- | -------------- |
| spec-builder     | Sonnet | 시스템 아키텍트   | SPEC 작성, EARS 명세    | `/alfred:1-spec`       | 명세 필요 시   |
| code-builder     | Sonnet | 수석 개발자       | TDD 구현, 코드 품질     | `/alfred:2-build`      | 구현 단계      |
| doc-syncer       | Haiku  | 테크니컬 라이터   | 문서 동기화, Living Doc | `/alfred:3-sync`       | 동기화 필요 시 |
| tag-agent        | Haiku  | 지식 관리자       | TAG 시스템, 추적성      | `@agent-tag-agent`     | TAG 작업 시    |
| git-manager      | Haiku  | 릴리스 엔지니어   | Git 워크플로우, 배포    | `@agent-git-manager`   | Git 조작 시    |
| debug-helper     | Sonnet | 트러블슈팅 전문가 | 오류 진단, 해결         | `@agent-debug-helper`  | 에러 발생 시   |
| trust-checker    | Haiku  | 품질 보증 리드    | TRUST 검증, 성능/보안   | `@agent-trust-checker` | 검증 요청 시   |
| cc-manager       | Sonnet | 데브옵스 엔지니어 | Claude Code 설정        | `@agent-cc-manager`    | 설정 필요 시   |
| project-manager  | Sonnet | 프로젝트 매니저   | 프로젝트 초기화         | `/alfred:0-project`    | 프로젝트 시작  |

**출력 예시** (FastAPI 웹 API 프로젝트):
```markdown
### 5개 전문 에이전트 생태계

| 에이전트              | 모델   | 페르소나          | 전문 영역               | 커맨드/호출            | 위임 시점      |
| --------------------- | ------ | ----------------- | ----------------------- | ---------------------- | -------------- |
| **spec-builder** 🏗️    | Sonnet | 시스템 아키텍트   | SPEC 작성, EARS 명세    | `/alfred:1-spec`       | 명세 필요 시   |
| **code-builder** 💎    | Sonnet | 수석 개발자       | TDD 구현, 코드 품질     | `/alfred:2-build`      | 구현 단계      |
| **doc-syncer** 📖      | Haiku  | 테크니컬 라이터   | 문서 동기화, Living Doc | `/alfred:3-sync`       | 동기화 필요 시 |
| **git-manager** 🚀     | Haiku  | 릴리스 엔지니어   | Git 워크플로우, 배포    | `@agent-git-manager`   | Git 조작 시    |
| **debug-helper** 🔬    | Sonnet | 트러블슈팅 전문가 | 오류 진단, 해결         | `@agent-debug-helper`  | 에러 발생 시   |
```

---

### STEP 4: CLAUDE.md 템플릿 치환

**목적**: CLAUDE.md 템플릿의 변수를 프로젝트 정보로 치환

**템플릿 변수**:
```yaml
{{PROJECT_NAME}}        # .moai/config.json의 project.name
{{PROJECT_DESCRIPTION}} # .moai/project/product.md에서 추출
{{PROJECT_VERSION}}     # .moai/config.json의 project.version
{{PROJECT_MODE}}        # .moai/config.json의 project.mode
{{AUTHOR}}              # .moai/config.json의 project.author
```

**실행**:
```bash
# config.json 읽기
CONFIG_DATA=$(cat .moai/config.json)
PROJECT_NAME=$(echo "$CONFIG_DATA" | grep '"name"' | sed 's/.*"name": "\(.*\)".*/\1/')
PROJECT_VERSION=$(echo "$CONFIG_DATA" | grep '"version"' | sed 's/.*"version": "\(.*\)".*/\1/')
PROJECT_MODE=$(echo "$CONFIG_DATA" | grep '"mode"' | sed 's/.*"mode": "\(.*\)".*/\1/')

# product.md에서 프로젝트 설명 추출 (첫 문단)
PROJECT_DESCRIPTION=$(head -20 .moai/project/product.md | grep -v "^#" | grep -v "^$" | head -1)

# CLAUDE.md 템플릿 읽기 및 변수 치환
NPM_ROOT=$(npm root -g 2>/dev/null || echo "$HOME/.npm-global/lib/node_modules")
TEMPLATE_FILE="${NPM_ROOT}/moai-adk/templates/CLAUDE.md"

sed -e "s/{{PROJECT_NAME}}/${PROJECT_NAME}/g" \
    -e "s/{{PROJECT_DESCRIPTION}}/${PROJECT_DESCRIPTION}/g" \
    -e "s/{{PROJECT_VERSION}}/${PROJECT_VERSION}/g" \
    -e "s/{{PROJECT_MODE}}/${PROJECT_MODE}/g" \
    "$TEMPLATE_FILE" > CLAUDE.md
```

---

### STEP 5: 선택된 Commands 복사

**목적**: feature-selection.json의 commands 목록만 복사

**실행**:
```bash
# Commands 디렉토리 생성
mkdir -p .claude/commands/alfred

# 선택된 commands 복사
NPM_ROOT=$(npm root -g 2>/dev/null || echo "$HOME/.npm-global/lib/node_modules")
SOURCE_DIR="${NPM_ROOT}/moai-adk/templates/.claude/commands/alfred"

# commands 배열: ["1-spec", "2-build", "3-sync"]
for cmd in 1-spec 2-build 3-sync; do
  if [ -f "${SOURCE_DIR}/${cmd}.md" ]; then
    cp "${SOURCE_DIR}/${cmd}.md" .claude/commands/alfred/
    echo "✓ Copied: /alfred:${cmd}"
  fi
done

# 0-project는 항상 복사 (필수)
cp "${SOURCE_DIR}/0-project.md" .claude/commands/alfred/
echo "✓ Copied: /alfred:0-project (필수)"
```

**결과**:
```
.claude/commands/alfred/
├── 0-project.md  (필수)
├── 1-spec.md
├── 2-build.md
└── 3-sync.md
```

---

### STEP 6: 선택된 Agents 복사

**목적**: feature-selection.json의 agents 목록만 복사

**실행**:
```bash
# Agents 디렉토리 생성
mkdir -p .claude/agents

# 선택된 agents 복사
NPM_ROOT=$(npm root -g 2>/dev/null || echo "$HOME/.npm-global/lib/node_modules")
SOURCE_DIR="${NPM_ROOT}/moai-adk/templates/.claude/agents"

# agents 배열: ["spec-builder", "code-builder", "doc-syncer", "git-manager", "debug-helper"]
for agent in spec-builder code-builder doc-syncer git-manager debug-helper; do
  if [ -f "${SOURCE_DIR}/${agent}.md" ]; then
    cp "${SOURCE_DIR}/${agent}.md" .claude/agents/
    echo "✓ Copied: @agent-${agent}"
  fi
done
```

**결과**:
```
.claude/agents/
├── spec-builder.md
├── code-builder.md
├── doc-syncer.md
├── git-manager.md
└── debug-helper.md
```

---

### STEP 7: 선택된 Skills 복사

**목적**: feature-selection.json의 skills 목록만 복사

**실행**:
```bash
# Skills 디렉토리 생성
mkdir -p .claude/skills

# 선택된 skills 복사
NPM_ROOT=$(npm root -g 2>/dev/null || echo "$HOME/.npm-global/lib/node_modules")
SOURCE_DIR="${NPM_ROOT}/moai-adk/templates/.claude/skills"

# skills 배열: ["moai-lang-python", "moai-domain-web-api", "moai-domain-backend"]
for skill in moai-lang-python moai-domain-web-api moai-domain-backend; do
  if [ -d "${SOURCE_DIR}/${skill}" ]; then
    cp -r "${SOURCE_DIR}/${skill}" .claude/skills/
    echo "✓ Copied: ${skill}"
  fi
done
```

**결과**:
```
.claude/skills/
├── moai-lang-python/
│   └── SKILL.md
├── moai-domain-web-api/
│   └── SKILL.md
└── moai-domain-backend/
    └── SKILL.md
```

---

### STEP 8: CLAUDE.md 에이전트 테이블 동적 업데이트

**목적**: 복사된 CLAUDE.md의 에이전트 테이블을 선택된 에이전트만 포함하도록 수정

**실행**:
```bash
# 에이전트 테이블 헤더 찾기
START_LINE=$(grep -n "### 9개 전문 에이전트 생태계" CLAUDE.md | cut -d: -f1)

# "### Built-in 에이전트" 섹션 찾기 (종료 지점)
END_LINE=$(grep -n "### Built-in 에이전트" CLAUDE.md | cut -d: -f1)

# 선택된 에이전트 수 계산
AGENT_COUNT=$(cat .moai/.feature-selection.json | grep -o '"agents":' | wc -l)

# 헤더 업데이트 (9개 → 실제 선택된 수)
sed -i '' "${START_LINE}s/9개/${AGENT_COUNT}개/" CLAUDE.md

# 기존 에이전트 테이블 제거 (헤더 다음 줄부터 Built-in 섹션 전까지)
TABLE_START=$((START_LINE + 1))
TABLE_END=$((END_LINE - 1))
sed -i '' "${TABLE_START},${TABLE_END}d" CLAUDE.md

# 새로운 에이전트 테이블 생성 및 삽입
# (위에서 생성한 동적 테이블 삽입)
```

**결과**:
```markdown
### 5개 전문 에이전트 생태계

Alfred는 5명의 전문 에이전트를 조율합니다. 각 에이전트는 IT 전문가 직무에 매핑되어 있습니다.

| 에이전트              | 모델   | 페르소나          | 전문 영역               | 커맨드/호출            | 위임 시점      |
| --------------------- | ------ | ----------------- | ----------------------- | ---------------------- | -------------- |
| **spec-builder** 🏗️    | Sonnet | 시스템 아키텍트   | SPEC 작성, EARS 명세    | `/alfred:1-spec`       | 명세 필요 시   |
| **code-builder** 💎    | Sonnet | 수석 개발자       | TDD 구현, 코드 품질     | `/alfred:2-build`      | 구현 단계      |
| **doc-syncer** 📖      | Haiku  | 테크니컬 라이터   | 문서 동기화, Living Doc | `/alfred:3-sync`       | 동기화 필요 시 |
| **git-manager** 🚀     | Haiku  | 릴리스 엔지니어   | Git 워크플로우, 배포    | `@agent-git-manager`   | Git 조작 시    |
| **debug-helper** 🔬    | Sonnet | 트러블슈팅 전문가 | 오류 진단, 해결         | `@agent-debug-helper`  | 에러 발생 시   |
```

---

### STEP 9: config.json 업데이트

**목적**: optimized: true로 변경하여 최적화 완료 표시

**실행**:
```bash
# config.json 읽기
CONFIG_PATH=".moai/config.json"

# optimized: false → optimized: true 변경
sed -i '' 's/"optimized": false/"optimized": true/' "$CONFIG_PATH"

echo "✅ optimized: true 업데이트 완료"
```

---

### STEP 10: 최적화 보고서 생성

**목적**: 최적화 결과 요약 보고서 생성

**실행**:
```bash
# 최적화 보고서 파일 생성
cat > .moai/.optimization-report.md <<'EOF'
# 프로젝트 최적화 보고서

## 📊 최적화 결과

- **프로젝트**: {{PROJECT_NAME}}
- **카테고리**: {{CATEGORY}}
- **주 언어**: {{LANGUAGE}}
- **프레임워크**: {{FRAMEWORK}}

## 🎯 선택된 기능

### Commands ({{COMMANDS_COUNT}}개)
{{COMMANDS_LIST}}

### Agents ({{AGENTS_COUNT}}개)
{{AGENTS_LIST}}

### Skills ({{SKILLS_COUNT}}개)
{{SKILLS_LIST}}

## 💡 최적화 효과

- **제외된 스킬**: {{EXCLUDED_COUNT}}개
- **경량화**: {{OPTIMIZATION_RATE}}%
- **선택된 스킬**: {{SELECTED_COUNT}}개

## 📝 다음 단계

1. `CLAUDE.md` 파일 확인
2. `/alfred:1-spec` 커맨드로 첫 SPEC 작성 시작
3. MoAI-ADK 워크플로우 사용

EOF
```

---

## 📊 출력 예시

### FastAPI 웹 API 프로젝트

**최적화 전**:
```
.claude/
├── commands/ (9개)
├── agents/ (9개)
└── skills/ (37개)
```

**최적화 후**:
```
.claude/
├── commands/ (4개)  ← 1-spec, 2-build, 3-sync, 0-project
├── agents/ (5개)    ← spec-builder, code-builder, doc-syncer, git-manager, debug-helper
└── skills/ (3개)    ← moai-lang-python, moai-domain-web-api, moai-domain-backend
```

**CLAUDE.md 변경**:
```markdown
# Before
### 9개 전문 에이전트 생태계
(9개 에이전트 테이블)

# After
### 5개 전문 에이전트 생태계
(5개 에이전트 테이블만)
```

---

## ⚠️ 에러 처리

### 에러 1: Feature Selection 파일 없음

**증상**: `.moai/.feature-selection.json` 파일이 없음

**해결**:
```bash
# moai-alfred-feature-selector 스킬 먼저 실행
❌ template-generator 실행 불가
→ feature-selector를 먼저 실행해야 합니다
→ Alfred: /alfred:0-project Phase 3 재실행
```

### 에러 2: 템플릿 파일 경로 없음

**증상**: moai-adk 패키지를 찾을 수 없음

**해결**:
```bash
# npm root 확인
npm root -g

# moai-adk 재설치
npm install -g moai-adk
```

### 에러 3: 선택된 파일이 템플릿에 없음

**증상**: feature-selection.json의 파일명이 실제 템플릿에 없음

**해결**:
```bash
# 경고 메시지 출력 후 건너뛰기
⚠️ 파일 없음: moai-domain-xyz
→ 해당 파일 건너뛰기
→ 나머지 파일 계속 복사
```

---

## 🔍 검증

**생성된 파일 확인**:
```bash
# CLAUDE.md 존재 확인
ls -la CLAUDE.md

# Commands 확인
ls -la .claude/commands/alfred/*.md

# Agents 확인
ls -la .claude/agents/*.md

# Skills 확인
ls -la .claude/skills/*/SKILL.md

# config.json optimized 확인
grep '"optimized"' .moai/config.json
```

**결과 보고**:
```
✅ Template Generator 완료!

📊 최적화 결과:
- CLAUDE.md: 맞춤형 에이전트 테이블 (5개)
- Commands: 4개 복사
- Agents: 5개 복사
- Skills: 3개 복사
- config.json: optimized: true

💡 경량화 효과:
- 제외된 스킬: 34개
- 경량화: 87%
```

---

## 📋 다음 단계

이 스킬이 완료되면, Alfred는 사용자에게 보고합니다:
```markdown
✅ 프로젝트 최적화 완료!

📊 결과:
- 프로젝트에 맞는 5개 에이전트 선택
- 37개 스킬 중 3개만 활성화
- CLAUDE.md 맞춤형 생성

🎯 다음 단계:
1. CLAUDE.md 파일 확인
2. /alfred:1-spec "첫 기능" 실행
3. MoAI-ADK 워크플로우 시작
```

---

**최종 업데이트**: 2025-10-19
**작성자**: @Alfred

# SPEC-PLUGIN-001 인수 기준 (Acceptance Criteria)

> **플러그인 구조 검증 시나리오**

---

## 📋 인수 기준 개요

이 문서는 SPEC-PLUGIN-001의 완료 조건을 Given-When-Then 형식으로 정의합니다.

---

## 🧪 테스트 시나리오

### 시나리오 1: 플러그인 매니페스트 검증

**목적**: `plugin.json`이 표준 구조를 따르는지 검증

**Given**:
- `.claude-plugin/plugin.json` 파일이 존재한다
- 필수 필드 `name`, `version`이 정의되어 있다

**When**:
- Claude Code가 플러그인을 로드한다

**Then**:
- 플러그인이 정상적으로 로드된다
- `name`이 소문자와 하이픈만 사용한다 (`moai-adk`)
- `version`이 Semantic Versioning을 따른다 (`0.3.0`)
- `description`, `author`, `homepage`, `license` 메타데이터가 존재한다

**검증 방법**:
```bash
# JSON 구조 검증
jq -e '.name, .version, .description, .author' .claude-plugin/plugin.json

# name 형식 검증 (소문자, 하이픈만)
jq -r '.name' .claude-plugin/plugin.json | grep -E '^[a-z-]+$'

# version SemVer 검증
jq -r '.version' .claude-plugin/plugin.json | grep -E '^\d+\.\d+\.\d+$'
```

---

### 시나리오 2: 디렉토리 구조 검증

**목적**: 플러그인 표준 디렉토리 구조를 따르는지 검증

**Given**:
- MoAI-ADK 프로젝트 루트 디렉토리가 존재한다

**When**:
- 디렉토리 구조를 검사한다

**Then**:
- `.claude-plugin/` 디렉토리가 존재한다
- `commands/alfred/` 디렉토리가 존재하고 비어있지 않다
- `agents/alfred/` 디렉토리가 존재하고 비어있지 않다
- `hooks/` 디렉토리가 존재하고 `hooks.json`을 포함한다
- `scripts/` 디렉토리가 존재한다
- `templates/` 디렉토리가 기존과 동일하게 유지된다

**검증 방법**:
```bash
# 필수 디렉토리 존재 확인
test -d .claude-plugin && echo "✅ .claude-plugin 존재"
test -d commands/alfred && echo "✅ commands/alfred 존재"
test -d agents/alfred && echo "✅ agents/alfred 존재"
test -d hooks && echo "✅ hooks 존재"
test -d scripts && echo "✅ scripts 존재"

# hooks.json 존재 확인
test -f hooks/hooks.json && echo "✅ hooks.json 존재"

# commands, agents 디렉토리 비어있지 않은지 확인
test "$(ls -A commands/alfred)" && echo "✅ commands/alfred 비어있지 않음"
test "$(ls -A agents/alfred)" && echo "✅ agents/alfred 비어있지 않음"
```

---

### 시나리오 3: hooks.json 변환 검증

**목적**: hooks.json이 표준 형식을 따르고 `${CLAUDE_PLUGIN_ROOT}`를 사용하는지 검증

**Given**:
- `hooks/hooks.json` 파일이 존재한다

**When**:
- hooks.json 내용을 파싱한다

**Then**:
- `PostToolUse` 후크가 정의되어 있다
- `PreToolUse` 후크가 정의되어 있다 (선택적)
- 모든 스크립트 경로가 `${CLAUDE_PLUGIN_ROOT}`를 사용한다
- `matcher` 필드가 유효한 정규식이다

**검증 방법**:
```bash
# PostToolUse 후크 존재 확인
jq -e '.hooks.PostToolUse' hooks/hooks.json

# ${CLAUDE_PLUGIN_ROOT} 사용 확인
grep -q '${CLAUDE_PLUGIN_ROOT}' hooks/hooks.json && echo "✅ CLAUDE_PLUGIN_ROOT 사용"

# matcher 필드 검증
jq -r '.hooks.PostToolUse[].matcher' hooks/hooks.json | grep -E '^[A-Za-z|]+$'
```

---

### 시나리오 4: 환경변수 경로 참조 검증

**목적**: 모든 내부 파일 참조가 `${CLAUDE_PLUGIN_ROOT}`를 사용하는지 검증

**Given**:
- `plugin.json`과 `hooks.json`이 존재한다

**When**:
- 파일 내 경로 참조를 검사한다

**Then**:
- `plugin.json`의 `commands`, `agents`, `hooks` 경로는 상대 경로(`./`로 시작)를 사용한다
- `hooks.json`의 스크립트 경로는 `${CLAUDE_PLUGIN_ROOT}`를 사용한다
- `mcpServers` 정의의 args는 `${CLAUDE_PLUGIN_ROOT}`를 사용한다

**검증 방법**:
```bash
# plugin.json 상대 경로 검증
jq -r '.commands, .agents, .hooks' .claude-plugin/plugin.json | grep -E '^\.\/'

# hooks.json 환경변수 사용 검증
jq -r '.hooks.PostToolUse[].hooks[].command' hooks/hooks.json | grep -q '${CLAUDE_PLUGIN_ROOT}'

# mcpServers 환경변수 사용 검증
jq -r '.mcpServers[].args[]' .claude-plugin/plugin.json | grep -q '${CLAUDE_PLUGIN_ROOT}'
```

---

### 시나리오 5: 플러그인 활성화 검증

**목적**: 플러그인이 Claude Code에서 정상적으로 활성화되는지 검증

**Given**:
- 플러그인 구조가 완성되어 있다
- Claude Code가 설치되어 있다

**When**:
- Claude Code를 시작한다

**Then**:
- `/alfred:1-spec` 커맨드가 사용 가능하다
- `/alfred:2-build` 커맨드가 사용 가능하다
- `/alfred:3-sync` 커맨드가 사용 가능하다
- `@agent-spec-builder` 에이전트가 호출 가능하다
- `@agent-code-builder` 에이전트가 호출 가능하다

**검증 방법** (수동):
```
1. Claude Code 실행
2. 프로젝트 디렉토리 열기
3. `/alfred:1-spec` 입력 후 자동완성 확인
4. `@agent-spec-builder` 입력 후 에이전트 활성화 확인
5. 에러 메시지 없이 정상 실행 확인
```

---

### 시나리오 6: PostToolUse 후크 실행 검증

**목적**: Write/Edit 도구 사용 후 자동 포맷팅이 실행되는지 검증

**Given**:
- `hooks/hooks.json`에 PostToolUse 후크가 정의되어 있다
- `scripts/format-code.sh` 스크립트가 실행 가능하다

**When**:
- Claude Code가 파일을 Write 또는 Edit한다

**Then**:
- PostToolUse 후크가 트리거된다
- `format-code.sh` 스크립트가 실행된다
- 포맷팅이 적용된 파일이 저장된다
- 에러 발생 시 경고 메시지가 표시되지만 작업은 계속된다

**검증 방법** (수동):
```
1. Claude Code에서 "테스트 파일 생성" 요청
2. 파일 생성 완료 후 로그 확인
3. PostToolUse 후크 실행 로그 확인
4. 생성된 파일의 포맷팅 상태 확인
```

---

## ✅ 완료 조건 (Definition of Done)

### 필수 조건

- [x] `.claude-plugin/plugin.json` 생성 완료
- [x] `hooks/hooks.json` 생성 완료
- [x] `scripts/` 디렉토리 및 기본 스크립트 생성
- [x] 필수 필드 7개 모두 포함 (`name`, `version`, `description`, `author`, `homepage`, `license`, `commands`)
- [ ] 모든 테스트 시나리오 통과
- [ ] 플러그인 활성화 성공

### 선택 조건 (권장)

- [ ] `scripts/format-code.sh` 동작 확인
- [ ] `scripts/validate-spec.sh` 동작 확인
- [ ] MCP 서버 통합 (선택)
- [ ] 마이그레이션 가이드 문서 작성

---

## 🧪 품질 게이트

### TRUST 5원칙 검증

#### T - Test First
- [ ] 플러그인 구조 검증 테스트 작성 (`tests/plugin/structure.test.ts`)
- [ ] hooks.json 파싱 테스트 작성
- [ ] 환경변수 경로 검증 테스트 작성

#### R - Readable
- [ ] `plugin.json` 주석 추가 (JSON 주석 불가 시 README 참조)
- [ ] `hooks.json` 주석 추가
- [ ] 디렉토리 구조 다이어그램 작성

#### U - Unified
- [ ] TypeScript 타입 정의 (`types/plugin.d.ts`)
- [ ] JSON Schema 검증 (선택)

#### S - Secured
- [ ] 스크립트 실행 권한 검증 (`chmod +x scripts/*.sh`)
- [ ] 경로 인젝션 방지
- [ ] 환경변수 검증

#### T - Trackable
- [ ] `@SPEC:PLUGIN-001` TAG 부여
- [ ] `@CODE:PLUGIN-001` TAG 부여
- [ ] `@TEST:PLUGIN-001` TAG 부여
- [ ] TAG 체인 무결성 검증

---

## 📊 성능 기준

### 플러그인 로딩 시간

- **목표**: Claude Code 시작 후 1초 내 플러그인 활성화
- **측정 방법**: Claude Code 로그에서 플러그인 로딩 시간 확인

### 후크 실행 시간

- **목표**: PostToolUse 후크 실행 시간 < 3초
- **측정 방법**: `time ${CLAUDE_PLUGIN_ROOT}/scripts/format-code.sh`

---

## 🔍 검증 체크리스트

### 자동 검증 (스크립트)

```bash
#!/bin/bash
# validate-plugin.sh

echo "🔍 플러그인 구조 검증 시작..."

# 1. 필수 파일 존재 확인
test -f .claude-plugin/plugin.json || { echo "❌ plugin.json 없음"; exit 1; }
test -f hooks/hooks.json || { echo "❌ hooks.json 없음"; exit 1; }

# 2. JSON 구조 검증
jq -e '.name, .version' .claude-plugin/plugin.json > /dev/null || { echo "❌ plugin.json 형식 오류"; exit 1; }

# 3. 환경변수 사용 확인
grep -q '${CLAUDE_PLUGIN_ROOT}' hooks/hooks.json || { echo "⚠️ CLAUDE_PLUGIN_ROOT 미사용"; }

# 4. 디렉토리 구조 확인
for dir in commands/alfred agents/alfred hooks scripts; do
  test -d "$dir" || { echo "❌ $dir 디렉토리 없음"; exit 1; }
done

echo "✅ 모든 검증 통과"
```

### 수동 검증 (사용자)

- [ ] Claude Code 실행 및 플러그인 활성화 확인
- [ ] `/alfred:1-spec` 커맨드 실행 테스트
- [ ] `@agent-spec-builder` 에이전트 호출 테스트
- [ ] PostToolUse 후크 실행 확인
- [ ] 에러 로그 없음 확인

---

## 📝 다음 단계

### TDD 구현 순서

1. **RED**: 테스트 작성
   - `tests/plugin/structure.test.ts`
   - `tests/plugin/hooks.test.ts`
   - `tests/plugin/validation.test.ts`

2. **GREEN**: 최소 구현
   - `.claude-plugin/plugin.json` 작성
   - `hooks/hooks.json` 작성
   - `scripts/format-code.sh` 작성

3. **REFACTOR**: 품질 개선
   - 스크립트 최적화
   - 에러 처리 강화
   - 문서화 완성

---

**작성자**: @Goos
**작성일**: 2025-10-10
**참조**: @SPEC:PLUGIN-001, @DOC:PLAN-PLUGIN-001

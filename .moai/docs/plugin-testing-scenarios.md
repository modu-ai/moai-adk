# Claude Code 플러그인 테스트 시나리오 & 자동화

**작성일**: 2025-10-31
**대상**: QA 엔지니어, 플러그인 개발자
**마켓플레이스**: `/Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace`

---

## 🧪 테스트 전략

### 테스트 수준

```
Unit Tests (플러그인 구성요소)
    ↓
Integration Tests (플러그인 ↔ Claude Code)
    ↓
E2E Tests (사용자 워크플로우)
    ↓
Performance Tests (성능 메트릭)
```

---

## 1️⃣ Unit Tests: 플러그인 파일 검증

### 1.1 plugin.json 스키마 검증

**검증 목록**:

```json
{
  "MUST_HAVE": [
    "id",           // 플러그인 고유 ID
    "name",         // 표시 이름
    "version",      // 버전 (semantic versioning)
    "status",       // development / beta / stable
    "description",  // 짧은 설명
    "author",       // 작성자
    "category",     // backend / frontend / devops / uiux / content
    "minClaudeCodeVersion"  // 최소 Claude Code 버전
  ],
  "COMMANDS": [
    {"name": "string", "description": "string", "path": "string (optional)"}
  ],
  "AGENTS": [
    {"name": "string", "type": "specialist|coordinator", "description": "string"}
  ],
  "SKILLS": [
    "moai-lang-*",
    "moai-domain-*",
    "moai-framework-*"
  ],
  "PERMISSIONS": {
    "allowedTools": ["Read", "Write", "Edit", "Bash"],
    "deniedTools": []
  }
}
```

**테스트 스크립트**:

```bash
#!/bin/bash
# plugins/validate-plugin-json.sh

PLUGIN_DIR="$1"
PLUGIN_JSON="$PLUGIN_DIR/.claude-plugin/plugin.json"

# 1. 파일 존재 확인
if [ ! -f "$PLUGIN_JSON" ]; then
  echo "❌ FAIL: plugin.json not found at $PLUGIN_JSON"
  exit 1
fi

# 2. JSON 형식 검증
if ! jq empty "$PLUGIN_JSON" 2>/dev/null; then
  echo "❌ FAIL: Invalid JSON format"
  exit 1
fi

# 3. 필수 필드 검증
REQUIRED_FIELDS=("id" "name" "version" "status" "description" "author" "category")
for field in "${REQUIRED_FIELDS[@]}"; do
  if [ -z "$(jq -r ".${field}" "$PLUGIN_JSON")" ]; then
    echo "❌ FAIL: Required field '$field' is missing or empty"
    exit 1
  fi
done

# 4. 버전 형식 검증 (semantic versioning)
VERSION=$(jq -r '.version' "$PLUGIN_JSON")
if ! [[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+ ]]; then
  echo "❌ FAIL: Version '$VERSION' does not follow semantic versioning"
  exit 1
fi

echo "✅ PASS: plugin.json is valid"
```

**실행**:

```bash
./validate-plugin-json.sh /Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace/plugins/moai-plugin-backend
```

### 1.2 명령어 파일 검증

**검증 사항**:

```bash
#!/bin/bash
# Validate command markdown files

PLUGIN_DIR="$1"
COMMANDS_DIR="$PLUGIN_DIR/commands"

if [ ! -d "$COMMANDS_DIR" ]; then
  echo "⚠️  WARNING: No commands directory found"
  exit 0
fi

# 각 명령어 파일 검증
for cmd_file in "$COMMANDS_DIR"/*.md; do
  if [ ! -f "$cmd_file" ]; then
    continue
  fi

  FILENAME=$(basename "$cmd_file")

  # 1. Markdown 헤더 확인
  if ! grep -q "^#" "$cmd_file"; then
    echo "❌ FAIL: No markdown header in $FILENAME"
    exit 1
  fi

  # 2. 명령어 설명 확인
  if ! grep -q "Description\|Overview\|Usage" "$cmd_file"; then
    echo "⚠️  WARNING: No clear description in $FILENAME"
  fi

  echo "✅ PASS: $FILENAME is valid"
done
```

### 1.3 에이전트 파일 검증

```bash
#!/bin/bash
# Validate agent markdown files

PLUGIN_DIR="$1"
AGENTS_DIR="$PLUGIN_DIR/agents"

if [ ! -d "$AGENTS_DIR" ]; then
  echo "⚠️  WARNING: No agents directory found"
  exit 0
fi

for agent_file in "$AGENTS_DIR"/*.md; do
  if [ ! -f "$agent_file" ]; then
    continue
  fi

  FILENAME=$(basename "$agent_file")

  # 1. Agent Type 확인
  if ! grep -q "Agent Type\|Type:" "$agent_file"; then
    echo "❌ FAIL: Agent Type not specified in $FILENAME"
    exit 1
  fi

  # 2. Role/Persona 확인
  if ! grep -q "Role\|Persona" "$agent_file"; then
    echo "⚠️  WARNING: No Role/Persona in $FILENAME"
  fi

  # 3. Responsibilities 확인
  if ! grep -q "Responsibilities\|Responsibilities" "$agent_file"; then
    echo "⚠️  WARNING: No Responsibilities in $FILENAME"
  fi

  echo "✅ PASS: $FILENAME is valid"
done
```

### 1.4 마켓플레이스 메타데이터 검증

```bash
#!/bin/bash
# Validate marketplace.json

MARKETPLACE_JSON="/Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace/marketplace.json"

# 1. JSON 형식
if ! jq empty "$MARKETPLACE_JSON" 2>/dev/null; then
  echo "❌ FAIL: Invalid marketplace.json format"
  exit 1
fi

# 2. 플러그인 수 확인
PLUGIN_COUNT=$(jq '.plugins | length' "$MARKETPLACE_JSON")
echo "Total plugins: $PLUGIN_COUNT"

# 3. 각 플러그인 검증
for i in $(seq 0 $((PLUGIN_COUNT - 1))); do
  PLUGIN_ID=$(jq -r ".plugins[$i].id" "$MARKETPLACE_JSON")
  PLUGIN_VERSION=$(jq -r ".plugins[$i].version" "$MARKETPLACE_JSON")
  PLUGIN_STATUS=$(jq -r ".plugins[$i].status" "$MARKETPLACE_JSON")

  echo "✅ Plugin: $PLUGIN_ID (v$PLUGIN_VERSION, $PLUGIN_STATUS)"
done
```

---

## 2️⃣ Integration Tests: Claude Code와의 통합

### 2.1 마켓플레이스 등록 테스트

**테스트 목표**: 로컬 마켓플레이스가 Claude Code에 인식되는가?

**단계**:

```bash
# Step 1: 마켓플레이스 추가
/plugin marketplace add /Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace

# Step 2: 마켓플레이스 확인
/plugin marketplace list

# 예상 출력:
# ✓ moai-marketplace (5 plugins available)
#   Path: /Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace
```

**테스트 검증**:

```bash
# marketplace.json이 올바르게 로드되었는가?
# 모든 5개 플러그인이 표시되는가?
# 플러그인 메타데이터가 올바른가?
```

### 2.2 플러그인 설치 테스트

**테스트 목표**: 각 플러그인이 성공적으로 설치되는가?

**테스트 케이스**:

```bash
# Test 2.2.1: Backend 플러그인 설치
/plugin install moai-plugin-backend@moai-marketplace

# 검증
/help | grep -E "init-fastapi|db-setup|resource-crud"

# 예상: 3개 명령어 표시


# Test 2.2.2: Frontend 플러그인 설치
/plugin install moai-plugin-frontend@moai-marketplace

# 검증
/help | grep -E "init-next|biome-setup"

# 예상: 2개 명령어 표시


# Test 2.2.3: 모든 플러그인 설치
/plugin install moai-plugin-devops@moai-marketplace
/plugin install moai-plugin-uiux@moai-marketplace
/plugin install moai-plugin-technical-blog@moai-marketplace

# 검증: 총 명령어 수 = 3 + 2 + 4 + 3 + 1 = 13개
/help | wc -l
```

### 2.3 플러그인 제거 및 재설치 테스트

```bash
# Test 2.3.1: 플러그인 제거
/plugin uninstall moai-plugin-backend

# 검증: 명령어 사라짐
/help | grep "init-fastapi" && echo "❌ FAIL" || echo "✅ PASS"

# Test 2.3.2: 재설치
/plugin install moai-plugin-backend@moai-marketplace

# 검증: 명령어 복구됨
/help | grep "init-fastapi" && echo "✅ PASS" || echo "❌ FAIL"
```

---

## 3️⃣ E2E Tests: 사용자 워크플로우

### 3.1 Backend 플러그인 완전한 워크플로우

**테스트 시나리오**: "FastAPI 프로젝트를 처음부터 끝까지 만들고 배포하기"

**실행 단계**:

```bash
# Step 1: 테스트 디렉토리 생성
mkdir -p /tmp/test-moai-backend
cd /tmp/test-moai-backend

# Step 2: FastAPI 프로젝트 초기화
/init-fastapi

# 입력 응답:
# Project name: test_api
# Python version: 3.13
# Database: PostgreSQL

# 검증: 필수 파일 생성 확인
test -f pyproject.toml && echo "✅ PASS: pyproject.toml"
test -d app && echo "✅ PASS: app directory"
test -d migrations && echo "✅ PASS: migrations directory"

# Step 3: 데이터베이스 설정
/db-setup

# 입력 응답:
# Database: PostgreSQL
# Host: localhost
# Port: 5432
# Database: test_api_db
# User: postgres
# Password: (비밀번호)

# 검증: .env 파일 생성
test -f .env && echo "✅ PASS: .env created"
grep "DATABASE_URL" .env && echo "✅ PASS: DATABASE_URL set"

# Step 4: CRUD 엔드포인트 생성
/resource-crud

# 입력 응답:
# Resource name: User
# Fields:
#   - name: string, required
#   - email: string, unique, required
#   - age: integer, optional

# 검증: 모델/스키마/라우터 생성
test -f app/models/user.py && echo "✅ PASS: User model"
test -f app/schemas/user.py && echo "✅ PASS: User schema"
test -f app/api/v1/endpoints/user.py && echo "✅ PASS: User endpoints"

# Step 5: 프로젝트 실행 테스트
cd /tmp/test-moai-backend
uv run uvicorn app.main:app --reload

# 검증: FastAPI 서버 시작
# Expected: INFO: Uvicorn running on http://127.0.0.1:8000

# Step 6: API 엔드포인트 테스트
curl http://localhost:8000/docs
# Expected: Swagger UI 로드됨

curl -X GET http://localhost:8000/api/v1/users
# Expected: [] (빈 배열)

curl -X POST http://localhost:8000/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}'
# Expected: {"id":1,"name":"John","email":"john@example.com","age":null}
```

**결과 평가**:

| 단계 | 기준 | 결과 |
|------|------|------|
| 초기화 | pyproject.toml, app/, migrations/ 생성 | ✅/❌ |
| DB 설정 | .env 생성, DATABASE_URL 설정 | ✅/❌ |
| CRUD | models, schemas, endpoints 생성 | ✅/❌ |
| 실행 | uvicorn 시작 (포트 8000) | ✅/❌ |
| API | Swagger UI 접근 가능 | ✅/❌ |
| 데이터 | POST/GET 정상 동작 | ✅/❌ |

### 3.2 Frontend 플러그인 완전한 워크플로우

```bash
# Step 1: 테스트 디렉토리
mkdir -p /tmp/test-moai-frontend
cd /tmp/test-moai-frontend

# Step 2: Next.js 프로젝트 초기화
/init-next

# 입력:
# Project name: test_app
# Package manager: npm
# TypeScript: Yes
# Tailwind CSS: Yes

# 검증
test -f package.json && echo "✅ PASS: package.json"
test -d app && echo "✅ PASS: app directory"
grep '"next"' package.json && echo "✅ PASS: Next.js dependency"

# Step 3: Biome 설정
/biome-setup

# 검증
test -f biome.json && echo "✅ PASS: biome.json"
grep "lint\|format" package.json && echo "✅ PASS: Biome scripts"

# Step 4: Playwright-MCP E2E 테스트 설정
/playwright-setup

# 검증
test -f playwright.config.ts && echo "✅ PASS: playwright.config.ts"
test -d tests && echo "✅ PASS: tests directory"
test -f .github/workflows/playwright.yml && echo "✅ PASS: GitHub Actions workflow"

# Step 5: 프로젝트 실행
npm run dev

# Expected: INFO  ready started server on 0.0.0.0:3000
curl http://localhost:3000
# Expected: Next.js 페이지 로드

# Step 6: Linting
npm run lint
# Expected: 0 linting errors

# Step 7: E2E 테스트
npm run test:e2e
# Expected: All tests passed with Playwright-MCP
```

### 3.3 전체 스택 통합 테스트

```bash
# Setup: 모든 플러그인 설치
/plugin install moai-plugin-backend@moai-marketplace
/plugin install moai-plugin-frontend@moai-marketplace
/plugin install moai-plugin-devops@moai-marketplace

# 프로젝트 구조 생성
mkdir -p /tmp/test-fullstack/{backend,frontend}

# Backend 설정
cd /tmp/test-fullstack/backend
/init-fastapi
/db-setup
/resource-crud (User, Product)

# Frontend 설정
cd /tmp/test-fullstack/frontend
/init-next
/biome-setup
/playwright-setup

# 배포 설정
cd /tmp/test-fullstack
/deploy-config (select Vercel, Supabase, Render)
/connect-vercel
/connect-supabase

# 최종 검증
# 모든 파일이 생성되었는가?
# 환경 변수가 올바르게 설정되었는가?
# API가 실행되는가?
# Frontend가 로드되는가?
```

---

## 4️⃣ Performance Tests: 성능 메트릭

### 4.1 플러그인 로드 시간

```bash
#!/bin/bash
# measure-plugin-load-time.sh

ITERATIONS=5
TOTAL_TIME=0

for i in $(seq 1 $ITERATIONS); do
  START=$(date +%s%N)

  # 플러그인 정보 요청
  /help > /dev/null

  END=$(date +%s%N)
  ELAPSED=$(( (END - START) / 1000000 ))  # 나노초 → 밀리초

  echo "Run $i: ${ELAPSED}ms"
  TOTAL_TIME=$(( TOTAL_TIME + ELAPSED ))
done

AVERAGE=$(( TOTAL_TIME / ITERATIONS ))
echo "Average load time: ${AVERAGE}ms"

# 기준: < 500ms
if [ $AVERAGE -lt 500 ]; then
  echo "✅ PASS: Plugin load time acceptable"
else
  echo "❌ FAIL: Plugin load time too slow"
fi
```

### 4.2 명령어 실행 시간

```bash
#!/bin/bash
# measure-command-execution-time.sh

# init-fastapi 실행 시간 측정
START=$(date +%s)
/init-fastapi << EOF
test_project
3.13
PostgreSQL
EOF
END=$(date +%s)

ELAPSED=$(( END - START ))
echo "init-fastapi execution: ${ELAPSED} seconds"

# 기준: < 10초 (프로젝트 생성 시간 포함)
if [ $ELAPSED -lt 10 ]; then
  echo "✅ PASS: Command execution fast"
fi
```

### 4.3 메모리 사용량

```bash
#!/bin/bash
# measure-memory-usage.sh

# Claude Code 프로세스 메모리 사용량 측정
PID=$(pgrep -f "claude-code" | head -1)

if [ -z "$PID" ]; then
  echo "Claude Code process not found"
  exit 1
fi

# 플러그인 설치 전
BEFORE=$(ps aux | grep $PID | awk '{print $6}')
echo "Memory before plugin install: ${BEFORE}KB"

# 플러그인 설치
/plugin install moai-plugin-backend@moai-marketplace

# 설치 후
AFTER=$(ps aux | grep $PID | awk '{print $6}')
echo "Memory after plugin install: ${AFTER}KB"

DIFF=$(( AFTER - BEFORE ))
echo "Memory increase: ${DIFF}KB"

# 기준: < 100MB 증가
if [ $DIFF -lt 102400 ]; then
  echo "✅ PASS: Memory usage acceptable"
fi
```

---

## 5️⃣ 자동화 테스트 스크립트

### 5.1 완전 자동화 테스트 (CI/CD)

```bash
#!/bin/bash
# run-all-plugin-tests.sh

set -e  # Exit on error

echo "🧪 Starting Plugin Test Suite..."
echo "=================================="

# Test 1: 구조 검증
echo "Test 1: Validating plugin structure..."
for plugin in backend frontend devops uiux technical-blog; do
  PLUGIN_DIR="/Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace/plugins/moai-plugin-$plugin"

  if [ ! -f "$PLUGIN_DIR/.claude-plugin/plugin.json" ]; then
    echo "❌ FAIL: $plugin plugin.json not found"
    exit 1
  fi

  if ! jq empty "$PLUGIN_DIR/.claude-plugin/plugin.json"; then
    echo "❌ FAIL: $plugin plugin.json invalid JSON"
    exit 1
  fi

  echo "✅ PASS: $plugin structure valid"
done

# Test 2: 마켓플레이스 메타데이터
echo ""
echo "Test 2: Validating marketplace.json..."
MARKETPLACE_JSON="/Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace/marketplace.json"

if ! jq empty "$MARKETPLACE_JSON"; then
  echo "❌ FAIL: marketplace.json invalid"
  exit 1
fi

PLUGIN_COUNT=$(jq '.plugins | length' "$MARKETPLACE_JSON")
echo "✅ PASS: marketplace.json valid ($PLUGIN_COUNT plugins)"

# Test 3: 명령어 파일 검증
echo ""
echo "Test 3: Validating command files..."
COMMAND_COUNT=0

for cmd_file in /Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace/plugins/*/commands/*.md; do
  if [ -f "$cmd_file" ]; then
    COMMAND_COUNT=$((COMMAND_COUNT + 1))

    if ! grep -q "^#" "$cmd_file"; then
      echo "⚠️  WARNING: $cmd_file has no header"
    fi
  fi
done
echo "✅ PASS: Found $COMMAND_COUNT command files"

# Test 4: 에이전트 파일 검증
echo ""
echo "Test 4: Validating agent files..."
AGENT_COUNT=0

for agent_file in /Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace/plugins/*/agents/*.md; do
  if [ -f "$agent_file" ]; then
    AGENT_COUNT=$((AGENT_COUNT + 1))
  fi
done
echo "✅ PASS: Found $AGENT_COUNT agent files"

# Test 5: 스킬 참조 검증
echo ""
echo "Test 5: Validating skill references..."
for plugin_json in /Users/goos/MoAI/MoAI-ADK-v1.0/moai-marketplace/plugins/*/.claude-plugin/plugin.json; do
  SKILLS=$(jq -r '.skills[]?' "$plugin_json" 2>/dev/null | wc -l)
  PLUGIN_ID=$(jq -r '.id' "$plugin_json")

  if [ $SKILLS -gt 0 ]; then
    echo "✅ PASS: $PLUGIN_ID has $SKILLS skills"
  fi
done

echo ""
echo "=================================="
echo "✅ All tests passed!"
echo "=================================="
```

**실행**:

```bash
chmod +x run-all-plugin-tests.sh
./run-all-plugin-tests.sh
```

---

## 📊 테스트 결과 리포팅

### 테스트 리포트 템플릿

```markdown
# Plugin Test Report
**Date**: 2025-10-31
**Tester**: [Name]
**Duration**: [Time]

## Summary
- ✅ Passed: 23/25
- ❌ Failed: 2/25
- ⚠️  Warnings: 3/25

## Test Results

### Unit Tests
- [✅] plugin.json validation
- [✅] Command files validation
- [✅] Agent files validation
- [❌] Skill reference validation (2 issues)

### Integration Tests
- [✅] Marketplace registration
- [✅] Plugin installation
- [✅] Plugin uninstallation
- [✅] Settings-based configuration

### E2E Tests
- [✅] Backend plugin workflow (5/5 steps passed)
- [✅] Frontend plugin workflow (4/4 steps passed)
- [⚠️] DevOps plugin workflow (3/4 steps, Render MCP pending)

### Performance Tests
- [✅] Plugin load time: 320ms (< 500ms ✓)
- [✅] Command execution: 4.2s (< 10s ✓)
- [⚠️] Memory usage: 85MB (threshold: 100MB)

## Issues Found

1. **moai-plugin-devops**: skill "moai-domain-devops" not found
2. **moai-plugin-technical-blog**: hook configuration missing

## Recommendations

1. Add missing skill definitions
2. Implement session start hooks for all plugins
3. Add version compatibility checks

## Next Steps

- [ ] Fix identified issues
- [ ] Re-run failed tests
- [ ] Update documentation
- [ ] Release v1.0.0-rc1
```

---

## 🔍 테스트 자동화 CI/CD 통합

### GitHub Actions 예제

```yaml
name: Plugin Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: macos-latest

    steps:
    - uses: actions/checkout@v3

    - name: Validate plugin structure
      run: |
        ./tests/validate-plugins.sh

    - name: Validate marketplace.json
      run: |
        jq empty moai-marketplace/marketplace.json

    - name: Run unit tests
      run: |
        ./tests/run-unit-tests.sh

    - name: Report results
      if: always()
      run: |
        ./tests/generate-report.sh > test-report.md

    - name: Upload report
      uses: actions/upload-artifact@v3
      with:
        name: test-report
        path: test-report.md
```

---

**테스트 가이드 버전**: 1.0.0
**마지막 업데이트**: 2025-10-31

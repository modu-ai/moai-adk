---
name: moai-alfred-feature-selector
description: 프로젝트 특성 분석 및 필요 기능 선택 (맞춤형 Commands, Agents, Skills 구성)
allowed-tools:
  - Read
  - Write
  - Grep
  - Bash
---

# Feature Selector Skill

## 🎯 목적

product/structure/tech.md를 분석하여 프로젝트에 필요한 **Commands, Agents, Skills만 선택**합니다.

**핵심 가치**: 불필요한 기능 제거 → 경량화 → 학습 곡선 감소

---

## 📥 입력

- `.moai/project/product.md` (프로젝트 카테고리, 도메인)
- `.moai/project/tech.md` (언어, 프레임워크)
- `.moai/config.json` (프로젝트 설정)

---

## 📤 출력

`.moai/.feature-selection.json` 파일 생성:

```json
{
  "category": "web-api",
  "domain": "backend",
  "language": "python",
  "framework": "fastapi",
  "commands": ["1-spec", "2-build", "3-sync"],
  "agents": ["spec-builder", "code-builder", "doc-syncer", "git-manager", "debug-helper"],
  "skills": ["moai-lang-python", "moai-domain-web-api", "moai-domain-backend"]
}
```

---

## 🔧 실행 로직

### STEP 1: 프로젝트 카테고리 감지

**목적**: product.md에서 프로젝트 유형 추출

**실행**:
```bash
# product.md에서 프로젝트 유형 관련 키워드 검색
grep -i "웹 API\|REST API\|GraphQL" .moai/project/product.md
grep -i "모바일 앱\|Flutter\|React Native" .moai/project/product.md
grep -i "CLI 도구\|명령줄 도구" .moai/project/product.md
grep -i "데이터 분석\|데이터 사이언스" .moai/project/product.md
```

**매핑 규칙**:
```yaml
키워드 패턴:
  - "웹 API", "REST API", "GraphQL", "백엔드" → category: "web-api"
  - "모바일 앱", "Flutter", "React Native", "iOS", "Android" → category: "mobile-app"
  - "CLI 도구", "명령줄 도구", "커맨드라인" → category: "cli-tool"
  - "데이터 분석", "데이터 사이언스", "머신러닝" → category: "data-science"
  - "웹사이트", "프론트엔드", "React", "Vue" → category: "web-frontend"
  - 위 모두 해당 안됨 → category: "generic"
```

**출력**:
```json
{
  "category": "web-api"
}
```

---

### STEP 2: 주 언어 감지

**목적**: tech.md 또는 config.json에서 주 언어 추출

**실행**:
```bash
# 1. config.json에서 확인 (우선순위 1)
grep "\"language\"" .moai/config.json

# 2. tech.md에서 확인 (우선순위 2)
grep -i "Python\|TypeScript\|JavaScript\|Dart\|Go\|Rust\|Java\|Swift\|Kotlin" .moai/project/tech.md
```

**출력**:
```json
{
  "language": "python"
}
```

---

### STEP 3: 프레임워크 감지 (선택적)

**목적**: tech.md에서 주요 프레임워크 추출

**실행**:
```bash
# 프레임워크 키워드 검색
grep -i "FastAPI\|Django\|Flask\|Express\|NestJS\|Flutter\|React\|Vue\|Spring Boot" .moai/project/tech.md
```

**출력**:
```json
{
  "framework": "fastapi"
}
```

---

### STEP 4: 매핑 테이블 적용

**목적**: 카테고리 + 언어 조합으로 필요 기능 선택

#### 매핑 테이블: 웹 API 프로젝트

```yaml
category: web-api
commands:
  - 1-spec      # SPEC 작성 (필수)
  - 2-build     # TDD 구현 (필수)
  - 3-sync      # 문서 동기화 (필수)

agents:
  - spec-builder    # SPEC 작성 (필수)
  - code-builder    # TDD 구현 (필수)
  - doc-syncer      # 문서 동기화 (필수)
  - git-manager     # Git 작업 (필수)
  - debug-helper    # 디버깅 (필수)

skills:
  python:
    - moai-lang-python
    - moai-domain-web-api
    - moai-domain-backend
  typescript:
    - moai-lang-typescript
    - moai-domain-web-api
    - moai-domain-backend
  go:
    - moai-lang-go
    - moai-domain-web-api
    - moai-domain-backend
  java:
    - moai-lang-java
    - moai-domain-web-api
    - moai-domain-backend
```

#### 매핑 테이블: 모바일 앱 프로젝트

```yaml
category: mobile-app
commands:
  - 1-spec
  - 2-build
  - 3-sync

agents:
  - spec-builder
  - code-builder
  - doc-syncer
  - git-manager
  - debug-helper

skills:
  dart:
    - moai-lang-dart
    - moai-domain-mobile-app
    - moai-domain-frontend
  swift:
    - moai-lang-swift
    - moai-domain-mobile-app
  kotlin:
    - moai-lang-kotlin
    - moai-domain-mobile-app
```

#### 매핑 테이블: CLI 도구 프로젝트

```yaml
category: cli-tool
commands:
  - 1-spec
  - 2-build
  - 3-sync

agents:
  - spec-builder
  - code-builder
  - doc-syncer
  - git-manager    # 선택적
  - debug-helper

skills:
  python:
    - moai-lang-python
    - moai-domain-cli-tool
  go:
    - moai-lang-go
    - moai-domain-cli-tool
  rust:
    - moai-lang-rust
    - moai-domain-cli-tool
```

#### 매핑 테이블: 데이터 분석 프로젝트

```yaml
category: data-science
commands:
  - 1-spec
  - 2-build
  - 3-sync

agents:
  - spec-builder
  - code-builder
  - doc-syncer
  - debug-helper
  # git-manager 선택적 (Jupyter Notebook은 Git 작업 적음)

skills:
  python:
    - moai-lang-python
    - moai-domain-data-science
  r:
    - moai-lang-r
    - moai-domain-data-science
  julia:
    - moai-lang-julia
    - moai-domain-data-science
```

#### 매핑 테이블: 프론트엔드 프로젝트

```yaml
category: web-frontend
commands:
  - 1-spec
  - 2-build
  - 3-sync

agents:
  - spec-builder
  - code-builder
  - doc-syncer
  - git-manager
  - debug-helper

skills:
  typescript:
    - moai-lang-typescript
    - moai-domain-frontend
  javascript:
    - moai-lang-javascript
    - moai-domain-frontend
```

#### 매핑 테이블: 범용 (Generic) 프로젝트

```yaml
category: generic
commands:
  - 1-spec
  - 2-build
  - 3-sync

agents:
  - spec-builder
  - code-builder
  - doc-syncer
  - git-manager
  - debug-helper

skills:
  python: [moai-lang-python]
  typescript: [moai-lang-typescript]
  javascript: [moai-lang-javascript]
  go: [moai-lang-go]
  rust: [moai-lang-rust]
  # 기본 언어 스킬만 포함
```

---

### STEP 5: 결과 파일 생성

**목적**: .moai/.feature-selection.json 파일 작성

**실행**:
```bash
# JSON 파일 생성
cat > .moai/.feature-selection.json <<'EOF'
{
  "category": "web-api",
  "domain": "backend",
  "language": "python",
  "framework": "fastapi",
  "commands": ["1-spec", "2-build", "3-sync"],
  "agents": ["spec-builder", "code-builder", "doc-syncer", "git-manager", "debug-helper"],
  "skills": ["moai-lang-python", "moai-domain-web-api", "moai-domain-backend"],
  "excluded_skills_count": 34,
  "optimization_rate": "87%"
}
EOF
```

---

## 📊 출력 예시

### FastAPI 웹 API 프로젝트

```json
{
  "category": "web-api",
  "domain": "backend",
  "language": "python",
  "framework": "fastapi",
  "commands": ["1-spec", "2-build", "3-sync"],
  "agents": ["spec-builder", "code-builder", "doc-syncer", "git-manager", "debug-helper"],
  "skills": ["moai-lang-python", "moai-domain-web-api", "moai-domain-backend"],
  "excluded_skills_count": 34,
  "optimization_rate": "87%"
}
```

### Flutter 모바일 앱

```json
{
  "category": "mobile-app",
  "domain": "frontend",
  "language": "dart",
  "framework": "flutter",
  "commands": ["1-spec", "2-build", "3-sync"],
  "agents": ["spec-builder", "code-builder", "doc-syncer", "git-manager", "debug-helper"],
  "skills": ["moai-lang-dart", "moai-domain-mobile-app", "moai-domain-frontend"],
  "excluded_skills_count": 34,
  "optimization_rate": "87%"
}
```

---

## ⚠️ 에러 처리

### 에러 1: 카테고리 감지 실패

**증상**: product.md에서 명확한 키워드를 찾을 수 없음

**해결**:
```json
{
  "category": "generic",
  "language": "python",
  "commands": ["1-spec", "2-build", "3-sync"],
  "agents": ["spec-builder", "code-builder", "doc-syncer", "git-manager", "debug-helper"],
  "skills": ["moai-lang-python"]
}
```

**메시지**:
```
⚠️ 프로젝트 카테고리를 자동 감지할 수 없습니다.
→ 범용 구성을 적용합니다.
→ 필요 시 product.md에 명확한 키워드 추가 후 /alfred:0-project update 실행
```

### 에러 2: 언어 감지 실패

**증상**: tech.md와 config.json 모두에서 언어를 찾을 수 없음

**해결**:
```json
{
  "language": "python",
  "note": "Default language applied. Please update tech.md or config.json"
}
```

**메시지**:
```
⚠️ 주 언어를 자동 감지할 수 없습니다.
→ Python을 기본 언어로 적용합니다.
→ tech.md에 주 언어 명시 후 /alfred:0-project update 실행
```

---

## 🔍 검증

**생성된 파일 확인**:
```bash
# 파일 존재 확인
ls -la .moai/.feature-selection.json

# 파일 내용 확인
cat .moai/.feature-selection.json

# JSON 구문 검증
python -m json.tool .moai/.feature-selection.json
```

**결과 보고**:
```
✅ Feature Selection 완료!

📊 선택된 기능:
- Category: web-api
- Language: python
- Commands: 3개
- Agents: 5개
- Skills: 3개

💡 최적화 효과:
- 제외된 스킬: 34개
- 경량화: 87%
```

---

## 📋 다음 단계

이 스킬이 완료되면, Alfred는 자동으로 다음 스킬을 호출합니다:
- **moai-alfred-template-generator**: 선택된 기능 기반 CLAUDE.md 동적 생성

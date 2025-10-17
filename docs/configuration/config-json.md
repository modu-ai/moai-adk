# @DOC:CONFIG-JSON-001 | Chain: @SPEC:DOCS-003 -> @DOC:CONFIG-001

# config.json 구조

`.moai/config.json`은 MoAI-ADK 프로젝트의 핵심 설정 파일입니다.

## 기본 구조

```json
{
  "project_name": "my-project",
  "project_language": "python",
  "test_framework": "pytest",
  "linter": "ruff",
  "formatter": "black",
  "mode": "personal",
  "git_strategy": "simple",
  "trust_threshold": 85,
  "tag_strategy": "auto"
}
```

---

## 필수 필드

### project_name

프로젝트 이름 (알파벳, 숫자, 하이픈만 허용):

```json
{
  "project_name": "todo-app"
}
```

### project_language

주 개발 언어:

```json
{
  "project_language": "python"  // python, typescript, java, go, rust
}
```

### test_framework

테스트 프레임워크:

```json
{
  "test_framework": "pytest"  // pytest, jest, junit, gotest
}
```

---

## 선택 필드

### mode

개발 모드 (Personal vs Team):

```json
{
  "mode": "personal"  // personal, team
}
```

- **personal**: 로컬 개발, 수동 커밋, PR 생성 안 함
- **team**: GitFlow 자동화, Draft PR, 자동 브랜치 생성

자세한 내용: [Personal vs Team 모드](personal-vs-team.md)

### git_strategy

Git 워크플로우 전략:

```json
{
  "git_strategy": "simple"  // simple, gitflow, trunk
}
```

- **simple**: main 브랜치만 사용
- **gitflow**: feature/develop/main 브랜치 활용
- **trunk**: trunk-based development

### trust_threshold

TRUST 원칙 준수 임계값 (0-100):

```json
{
  "trust_threshold": 85
}
```

코드가 이 임계값 이하로 떨어지면 구현을 중단합니다.

### tag_strategy

TAG 생성 전략:

```json
{
  "tag_strategy": "auto"  // auto, manual, hybrid
}
```

- **auto**: TAG 자동 생성 및 체인 연결
- **manual**: 수동 TAG 작성
- **hybrid**: 자동 생성 + 수동 검증

---

## 고급 설정

### linter & formatter

코드 품질 도구:

```json
{
  "linter": "ruff",      // ruff, eslint, golangci-lint
  "formatter": "black"   // black, prettier, gofmt
}
```

### 언어별 설정

Python 프로젝트:

```json
{
  "project_language": "python",
  "test_framework": "pytest",
  "linter": "ruff",
  "formatter": "black",
  "python_version": "3.13"
}
```

TypeScript 프로젝트:

```json
{
  "project_language": "typescript",
  "test_framework": "jest",
  "linter": "eslint",
  "formatter": "prettier",
  "node_version": "20"
}
```

---

## 검증 규칙

config.json은 다음 규칙을 따라야 합니다:

1. **필수 필드 존재**: project_name, project_language, test_framework
2. **유효한 값**: 각 필드는 허용된 값만 사용
3. **일관성**: project_language와 test_framework는 호환되어야 함

---

## 다음 단계

- [Personal vs Team 모드](personal-vs-team.md) - 모드별 차이점
- [Advanced Settings](advanced-settings.md) - 고급 설정 옵션

---

**다음**: [Personal vs Team →](personal-vs-team.md)

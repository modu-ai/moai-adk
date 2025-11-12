
## 개요

본 문서는 **SPEC-LANGUAGE-DETECTION-001**의 구현 완료를 검증하기 위한 상세한 인수 기준을 정의합니다. 모든 시나리오는 **Given-When-Then** 형식으로 작성되었으며, 테스트 가능하고 측정 가능한 기준을 제공합니다.

---

## Scenario 1: Python 프로젝트 감지 및 워크플로우 생성

### Given (전제 조건)

- Python 프로젝트가 존재하며, 프로젝트 루트에 `pyproject.toml` 파일이 있음
- MoAI-ADK가 설치되어 있음
- tdd-implementer 에이전트가 활성화되어 있음

### When (실행 조건)

- 사용자가 `/alfred:2-run` 명령으로 CI/CD 워크플로우 생성을 요청함
- tdd-implementer 에이전트가 `Skill("moai-alfred-language-detection")`을 호출함

### Then (예상 결과)

- ✅ 언어 감지 결과: `"python"` 반환
- ✅ 워크플로우 파일 생성: `.github/workflows/tag-validation.yml`
- ✅ 워크플로우 내용 검증:
  - `actions/setup-python@v5` 액션 사용
  - Python 버전: `3.11` 지정
  - `uv` 패키지 매니저 명령 포함 (`uv sync`, `uv run pytest`)
- ✅ 로그 메시지: `"Detected language: python"` 출력
- ✅ 생성된 워크플로우가 YAML 구문 검증 통과

### 테스트 코드 예시

```python
def test_python_project_workflow_generation():
    """Python 프로젝트 워크플로우 생성 인수 테스트"""
    with temp_project(files=["pyproject.toml"]) as project_root:
        # When
        workflow_path = generate_ci_workflow(project_root)

        # Then
        assert Path(workflow_path).exists()
        with open(workflow_path, "r") as f:
            content = f.read()
        assert "setup-python@v5" in content
        assert "python-version: '3.11'" in content
        assert "uv sync" in content
        assert "uv run pytest" in content
```

---

## Scenario 2: JavaScript 프로젝트 감지 및 워크플로우 생성

### Given (전제 조건)

- JavaScript 프로젝트가 존재하며, 프로젝트 루트에 `package.json`만 있음
- `tsconfig.json` 파일이 **존재하지 않음** (TypeScript 제외)
- MoAI-ADK가 설치되어 있음

### When (실행 조건)

- 사용자가 CI/CD 워크플로우 생성을 요청함
- tdd-implementer 에이전트가 언어 감지를 수행함

### Then (예상 결과)

- ✅ 언어 감지 결과: `"javascript"` 반환
- ✅ 워크플로우 파일 생성: `.github/workflows/tag-validation.yml`
- ✅ 워크플로우 내용 검증:
  - `actions/setup-node@v4` 액션 사용
  - Node.js 버전: `20` 지정
  - `npm ci` 명령 포함
  - `npm run test:tags` 명령 포함
- ✅ 타입 체크 단계가 **포함되지 않음** (TypeScript 아님)
- ✅ 로그 메시지: `"Detected language: javascript"` 출력

### 테스트 코드 예시

```python
def test_javascript_project_workflow_generation():
    """JavaScript 프로젝트 워크플로우 생성 인수 테스트"""
    with temp_project(files=["package.json"]) as project_root:
        # When
        workflow_path = generate_ci_workflow(project_root)

        # Then
        assert Path(workflow_path).exists()
        with open(workflow_path, "r") as f:
            content = f.read()
        assert "setup-node@v4" in content
        assert "node-version: '20'" in content
        assert "npm ci" in content
        assert "npm run test:tags" in content
        # TypeScript 전용 단계 없음
        assert "type-check" not in content
```

---

## Scenario 3: TypeScript 프로젝트 감지 및 워크플로우 생성

### Given (전제 조건)

- TypeScript 프로젝트가 존재하며, 프로젝트 루트에 `package.json`과 `tsconfig.json` 모두 존재
- MoAI-ADK가 설치되어 있음

### When (실행 조건)

- 사용자가 CI/CD 워크플로우 생성을 요청함
- tdd-implementer 에이전트가 언어 감지를 수행함

### Then (예상 결과)

- ✅ 언어 감지 결과: `"typescript"` 반환
- ✅ 워크플로우 파일 생성: `.github/workflows/tag-validation.yml`
- ✅ 워크플로우 내용 검증:
  - `actions/setup-node@v4` 액션 사용
  - Node.js 버전: `20` 지정
  - `npm ci` 명령 포함
  - **타입 체크 단계 포함**: `npm run type-check`
  - `npm run test:tags` 명령 포함
- ✅ 로그 메시지: `"Detected language: typescript"` 출력

### 테스트 코드 예시

```python
def test_typescript_project_workflow_generation():
    """TypeScript 프로젝트 워크플로우 생성 인수 테스트"""
    with temp_project(files=["package.json", "tsconfig.json"]) as project_root:
        # When
        workflow_path = generate_ci_workflow(project_root)

        # Then
        assert Path(workflow_path).exists()
        with open(workflow_path, "r") as f:
            content = f.read()
        assert "setup-node@v4" in content
        assert "node-version: '20'" in content
        assert "npm ci" in content
        assert "npm run type-check" in content  # TypeScript 전용
        assert "npm run test:tags" in content
```

---

## Scenario 4: 혼합 언어 프로젝트 우선순위 처리

### Given (전제 조건)

- 모노레포 프로젝트가 존재하며, 프로젝트 루트에 다음 파일들이 모두 존재:
  - `package.json` (JavaScript/TypeScript)
  - `tsconfig.json` (TypeScript)
  - `pyproject.toml` (Python)
- MoAI-ADK가 설치되어 있음

### When (실행 조건)

- 사용자가 CI/CD 워크플로우 생성을 요청함
- tdd-implementer 에이전트가 언어 감지를 수행함
- 언어 감지 로직이 우선순위 규칙을 적용함

### Then (예상 결과)

- ✅ 언어 감지 결과: `"typescript"` 반환 (최우선 순위)
- ✅ Python은 무시됨 (우선순위 낮음)
- ✅ TypeScript 워크플로우 템플릿이 적용됨
- ✅ 로그 메시지:
  - `"Detected language: typescript (priority over python, javascript)"`
  - 우선순위 결정 이유가 명확히 기록됨

### 테스트 코드 예시

```python
def test_mixed_language_priority_handling():
    """혼합 언어 프로젝트 우선순위 처리 인수 테스트"""
    with temp_project(files=["package.json", "tsconfig.json", "pyproject.toml"]) as project_root:
        # When
        detector = LanguageDetector()
        language = detector.detect_language(project_root)

        # Then
        assert language == "typescript"  # 최우선 순위

        # 워크플로우도 TypeScript 템플릿 사용
        workflow_path = generate_ci_workflow(project_root)
        with open(workflow_path, "r") as f:
            content = f.read()
        assert "setup-node@v4" in content
        assert "type-check" in content
```

---

## Scenario 5: 언어 감지 실패 시 에러 처리

### Given (전제 조건)

- 프로젝트 루트에 **어떠한 언어 설정 파일도 존재하지 않음**:
  - `pyproject.toml` 없음
  - `package.json` 없음
  - `go.mod` 없음
  - 기타 설정 파일 없음
- MoAI-ADK가 설치되어 있음

### When (실행 조건)

- 사용자가 CI/CD 워크플로우 생성을 요청함
- tdd-implementer 에이전트가 언어 감지를 시도함

### Then (예상 결과)

- ✅ `LanguageDetectionError` 예외 발생
- ✅ 명확한 에러 메시지 출력:
  ```
  ❌ 언어 감지 실패: 프로젝트 루트에 설정 파일을 찾을 수 없습니다.

  지원하는 설정 파일:
  - Python: pyproject.toml, requirements.txt
  - JavaScript/TypeScript: package.json
  - Go: go.mod
  - Ruby: Gemfile

  해결 방법:
  1. 프로젝트에 적절한 설정 파일을 추가하세요.
  2. 또는 .moai/config.json에서 언어를 수동으로 지정하세요:
     {
       "language": {
         "codebase_language": "python"
       }
     }
  ```
- ✅ 워크플로우 생성이 **중단됨** (불완전한 워크플로우 생성 금지)
- ✅ 프로그램 종료 코드: `1` (실패)

### 테스트 코드 예시

```python
def test_language_detection_failure_handling():
    """언어 감지 실패 에러 처리 인수 테스트"""
    with temp_project(files=[]) as project_root:
        # When/Then
        detector = LanguageDetector()
        with pytest.raises(LanguageDetectionError) as exc_info:
            detector.detect_language(project_root)

        # 에러 메시지 검증
        error_message = str(exc_info.value)
        assert "언어 감지 실패" in error_message
        assert "설정 파일을 찾을 수 없습니다" in error_message
        assert ".moai/config.json" in error_message
```

---

## Scenario 6: 사용자 명시적 언어 오버라이드

### Given (전제 조건)

- 프로젝트 루트에 `package.json` (JavaScript) 존재
- `.moai/config.json`에 사용자가 Python을 명시적으로 지정:
  ```json
  {
    "language": {
      "codebase_language": "python"
    }
  }
  ```
- MoAI-ADK가 설치되어 있음

### When (실행 조건)

- 사용자가 CI/CD 워크플로우 생성을 요청함
- tdd-implementer 에이전트가 언어 감지를 시도함
- 언어 감지 로직이 설정 파일에서 명시적 지정을 발견함

### Then (예상 결과)

- ✅ 자동 감지를 **건너뜀**
- ✅ 사용자 지정 언어 사용: `"python"`
- ✅ Python 워크플로우 템플릿이 적용됨 (JavaScript 아님)
- ✅ 로그 메시지:
  ```
  ℹ️ 사용자 지정 언어 사용: python (자동 감지 건너뜀)
  이유: .moai/config.json에서 명시적으로 지정됨
  ```
- ✅ 오버라이드 이유가 감사 로그에 기록됨

### 테스트 코드 예시

```python
def test_explicit_language_override():
    """사용자 명시적 언어 오버라이드 인수 테스트"""
    config = {
        "language": {
            "codebase_language": "python"
        }
    }
    with temp_project(files=["package.json"], config=config) as project_root:
        # When
        detector = LanguageDetector()
        language = detector.detect_language(project_root)

        # Then
        assert language == "python"  # JavaScript 아님!

        # Python 워크플로우 적용 확인
        workflow_path = generate_ci_workflow(project_root)
        with open(workflow_path, "r") as f:
            content = f.read()
        assert "setup-python@v5" in content
        assert "setup-node" not in content
```

---

## Scenario 7: Yarn 패키지 매니저 자동 감지

### Given (전제 조건)

- TypeScript 프로젝트가 존재하며:
  - `package.json` 존재
  - `tsconfig.json` 존재
  - **`yarn.lock` 존재** (Yarn 패키지 매니저 사용)
- MoAI-ADK가 설치되어 있음

### When (실행 조건)

- 사용자가 CI/CD 워크플로우 생성을 요청함
- tdd-implementer 에이전트가 패키지 매니저 감지를 수행함

### Then (예상 결과)

- ✅ 언어 감지 결과: `"typescript"` 반환
- ✅ 패키지 매니저 감지 결과: `"yarn"` 반환
- ✅ 워크플로우 내용 검증:
  - `npm ci` 대신 `yarn install --frozen-lockfile` 사용
  - `npm run` 대신 `yarn run` 사용
- ✅ 로그 메시지:
  ```
  Detected language: typescript
  Detected package manager: yarn
  ```

### 테스트 코드 예시

```python
def test_yarn_package_manager_detection():
    """Yarn 패키지 매니저 감지 인수 테스트"""
    with temp_project(files=["package.json", "tsconfig.json", "yarn.lock"]) as project_root:
        # When
        detector = LanguageDetector()
        package_manager = detector.detect_package_manager(project_root)

        # Then
        assert package_manager == "yarn"

        # 워크플로우에 yarn 명령 포함 확인
        workflow_path = generate_ci_workflow(project_root)
        with open(workflow_path, "r") as f:
            content = f.read()
        assert "yarn install --frozen-lockfile" in content or "yarn" in content
        assert "npm ci" not in content  # npm 명령 없음
```

---

## Scenario 8: pnpm 패키지 매니저 자동 감지

### Given (전제 조건)

- JavaScript 프로젝트가 존재하며:
  - `package.json` 존재
  - **`pnpm-lock.yaml` 존재** (pnpm 패키지 매니저 사용)
- MoAI-ADK가 설치되어 있음

### When (실행 조건)

- 사용자가 CI/CD 워크플로우 생성을 요청함
- tdd-implementer 에이전트가 패키지 매니저 감지를 수행함

### Then (예상 결과)

- ✅ 언어 감지 결과: `"javascript"` 반환
- ✅ 패키지 매니저 감지 결과: `"pnpm"` 반환
- ✅ 워크플로우 내용 검증:
  - `pnpm/action-setup@v2` 액션 추가
  - `pnpm install --frozen-lockfile` 사용
  - `pnpm run test:tags` 사용
- ✅ 로그 메시지:
  ```
  Detected language: javascript
  Detected package manager: pnpm
  ```

### 테스트 코드 예시

```python
def test_pnpm_package_manager_detection():
    """pnpm 패키지 매니저 감지 인수 테스트"""
    with temp_project(files=["package.json", "pnpm-lock.yaml"]) as project_root:
        # When
        detector = LanguageDetector()
        package_manager = detector.detect_package_manager(project_root)

        # Then
        assert package_manager == "pnpm"

        # 워크플로우에 pnpm 명령 포함 확인
        workflow_path = generate_ci_workflow(project_root)
        with open(workflow_path, "r") as f:
            content = f.read()
        assert "pnpm/action-setup@v2" in content
        assert "pnpm install" in content
```

---

## Scenario 9: Bun 런타임/패키지 매니저 자동 감지

### Given (전제 조건)

- TypeScript 프로젝트가 존재하며:
  - `package.json` 존재
  - `tsconfig.json` 존재
  - **`bun.lockb` 존재** (Bun 패키지 매니저 사용)
- MoAI-ADK가 설치되어 있음

### When (실행 조건)

- 사용자가 CI/CD 워크플로우 생성을 요청함
- tdd-implementer 에이전트가 패키지 매니저 감지를 수행함

### Then (예상 결과)

- ✅ 언어 감지 결과: `"typescript"` 반환
- ✅ 패키지 매니저 감지 결과: `"bun"` 반환
- ✅ 워크플로우 내용 검증:
  - `oven-sh/setup-bun@v1` 액션 추가
  - `bun install --frozen-lockfile` 사용
  - `bun run test:tags` 사용
- ✅ 로그 메시지:
  ```
  Detected language: typescript
  Detected package manager: bun
  ```

### 테스트 코드 예시

```python
def test_bun_package_manager_detection():
    """Bun 패키지 매니저 감지 인수 테스트"""
    with temp_project(files=["package.json", "tsconfig.json", "bun.lockb"]) as project_root:
        # When
        detector = LanguageDetector()
        package_manager = detector.detect_package_manager(project_root)

        # Then
        assert package_manager == "bun"

        # 워크플로우에 bun 명령 포함 확인
        workflow_path = generate_ci_workflow(project_root)
        with open(workflow_path, "r") as f:
            content = f.read()
        assert "oven-sh/setup-bun@v1" in content
        assert "bun install" in content
```

---

## Scenario 10: Go 프로젝트 감지 및 워크플로우 생성

### Given (전제 조건)

- Go 프로젝트가 존재하며, 프로젝트 루트에 `go.mod` 파일이 있음
- MoAI-ADK가 설치되어 있음

### When (실행 조건)

- 사용자가 CI/CD 워크플로우 생성을 요청함
- tdd-implementer 에이전트가 언어 감지를 수행함

### Then (예상 결과)

- ✅ 언어 감지 결과: `"go"` 반환
- ✅ 워크플로우 파일 생성: `.github/workflows/tag-validation.yml`
- ✅ 워크플로우 내용 검증:
  - `actions/setup-go@v5` 액션 사용
  - Go 버전: `1.22` 지정
  - `go test -v ./...` 명령 포함
- ✅ 로그 메시지: `"Detected language: go"` 출력

### 테스트 코드 예시

```python
def test_go_project_workflow_generation():
    """Go 프로젝트 워크플로우 생성 인수 테스트"""
    with temp_project(files=["go.mod"]) as project_root:
        # When
        workflow_path = generate_ci_workflow(project_root)

        # Then
        assert Path(workflow_path).exists()
        with open(workflow_path, "r") as f:
            content = f.read()
        assert "setup-go@v5" in content
        assert "go-version: '1.22'" in content
        assert "go test -v ./..." in content
```

---

## Scenario 11: 워크플로우 템플릿 누락 시 에러 처리

### Given (전제 조건)

- Ruby 프로젝트가 존재하며, `Gemfile`이 있음
- 언어 감지는 성공: `"ruby"` 반환
- **하지만 `ruby-tag-validation.yml` 템플릿이 존재하지 않음**

### When (실행 조건)

- 사용자가 CI/CD 워크플로우 생성을 요청함
- tdd-implementer 에이전트가 워크플로우 템플릿을 찾으려 시도함

### Then (예상 결과)

- ✅ `WorkflowTemplateNotFoundError` 예외 발생
- ✅ 명확한 에러 메시지 출력:
  ```
  ❌ 워크플로우 템플릿을 찾을 수 없습니다: ruby

  감지된 언어는 지원하지만, 워크플로우 템플릿이 아직 구현되지 않았습니다.

  지원 가능한 언어:
  - Python (python-tag-validation.yml)
  - JavaScript (javascript-tag-validation.yml)
  - TypeScript (typescript-tag-validation.yml)
  - Go (go-tag-validation.yml)

  해결 방법:
  1. 다른 지원 언어를 사용하거나
  2. 템플릿 구현을 요청하세요 (GitHub Issue 생성)
  ```
- ✅ 불완전한 워크플로우가 **생성되지 않음**
- ✅ 프로그램 종료 코드: `1` (실패)

### 테스트 코드 예시

```python
def test_workflow_template_not_found_error():
    """워크플로우 템플릿 누락 에러 처리 인수 테스트"""
    with temp_project(files=["Gemfile"]) as project_root:
        # When/Then
        detector = LanguageDetector()
        language = detector.detect_language(project_root)
        assert language == "ruby"

        with pytest.raises(WorkflowTemplateNotFoundError) as exc_info:
            detector.get_workflow_template_path(language)

        # 에러 메시지 검증
        error_message = str(exc_info.value)
        assert "워크플로우 템플릿을 찾을 수 없습니다" in error_message
        assert "ruby" in error_message
```

---

## Scenario 12: 언어 감지 결과 캐싱 동작 확인

### Given (전제 조건)

- Python 프로젝트가 존재하며, `pyproject.toml`이 있음
- MoAI-ADK가 설치되어 있음

### When (실행 조건)

- 같은 프로젝트 루트에 대해 언어 감지를 **3번 연속** 호출함
- 각 호출의 실행 시간을 측정함

### Then (예상 결과)

- ✅ 첫 번째 호출 (캐시 미스):
  - 실행 시간: < 100ms
  - 파일 시스템 탐색 수행
- ✅ 두 번째 호출 (캐시 히트):
  - 실행 시간: < 10ms (10배 빠름)
  - 파일 시스템 탐색 건너뜀
- ✅ 세 번째 호출 (캐시 히트):
  - 실행 시간: < 10ms
  - 캐시된 결과 재사용
- ✅ 모든 호출에서 동일한 결과 반환: `"python"`
- ✅ 로그 메시지:
  ```
  [1st call] Detecting language... (cache miss)
  [2nd call] Using cached language detection result (cache hit)
  [3rd call] Using cached language detection result (cache hit)
  ```

### 테스트 코드 예시

```python
def test_language_detection_caching():
    """언어 감지 결과 캐싱 동작 확인 인수 테스트"""
    with temp_project(files=["pyproject.toml"]) as project_root:
        detector = LanguageDetector()

        # 첫 번째 호출 (캐시 미스)
        start = time.time()
        result1 = detector.detect_language_cached(str(project_root))
        duration1 = time.time() - start

        # 두 번째 호출 (캐시 히트)
        start = time.time()
        result2 = detector.detect_language_cached(str(project_root))
        duration2 = time.time() - start

        # Then
        assert result1 == result2 == "python"
        assert duration1 < 0.1  # 100ms
        assert duration2 < 0.01  # 10ms (10배 빠름)
        assert duration2 < duration1 / 5  # 최소 5배 성능 향상
```

---

## Quality Gates (품질 게이트)

### QG1: 테스트 커버리지

- ✅ 단위 테스트 커버리지: **90% 이상**
- ✅ 통합 테스트 커버리지: **85% 이상**
- ✅ 모든 시나리오 (1-12) 테스트 코드 작성 완료

### QG2: 성능 요구사항

- ✅ 언어 감지 시간 (캐시 미스): **< 100ms**
- ✅ 언어 감지 시간 (캐시 히트): **< 10ms**
- ✅ 워크플로우 생성 시간: **< 500ms**

### QG3: 에러 처리 품질

- ✅ 모든 에러 메시지가 명확하고 실행 가능한 해결 방법 포함
- ✅ 침묵적 실패 0건 (모든 실패는 명시적 예외 발생)
- ✅ 에러 케이스 테스트 커버리지: **100%**

### QG4: 문서화 품질

- ✅ 모든 지원 언어에 대한 사용 예시 포함
- ✅ 트러블슈팅 가이드 작성 완료
- ✅ API 문서 (docstring) 커버리지: **100%**

### QG5: CI/CD 통합

- ✅ GitHub Actions에서 모든 시나리오 테스트 통과
- ✅ Python, JavaScript, TypeScript, Go 프로젝트 E2E 테스트 통과
- ✅ 하위 호환성 검증 (레거시 Python 워크플로우) 통과

---

## Definition of Done (완료 정의)

### 기능 완료 기준

- ✅ 모든 12개 시나리오가 구현되고 테스트 통과
- ✅ Python, JavaScript, TypeScript, Go 언어 지원 완료
- ✅ npm, yarn, pnpm, bun 패키지 매니저 감지 지원
- ✅ 언어 감지 캐싱 구현 완료

### 품질 완료 기준

- ✅ 단위 테스트 커버리지 90% 이상
- ✅ 통합 테스트 100% 통과
- ✅ 코드 리뷰 승인 (TRUST 5 원칙 준수 확인)
- ✅ 성능 요구사항 충족 (언어 감지 < 100ms)

### 문서 완료 기준

- ✅ 사용자 가이드 작성 (`.moai/docs/language-detection-guide.md`)
- ✅ tdd-implementer 에이전트 문서 업데이트
- ✅ CHANGELOG 업데이트 완료
- ✅ API 문서 (docstring) 100% 작성

### 배포 완료 기준

- ✅ `develop` 브랜치 테스트 통과
- ✅ `main` 브랜치 머지 완료
- ✅ MoAI-ADK v0.8.0 릴리스 태깅
- ✅ GitHub Release Notes 작성 완료

---

**Generated with**: 🎩 Alfred (MoAI-ADK v0.7.0)
**Acceptance Criteria for**: SPEC-LANGUAGE-DETECTION-001

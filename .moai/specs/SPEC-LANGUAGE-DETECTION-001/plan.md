
## 개요

본 구현 계획은 **SPEC-LANGUAGE-DETECTION-001**의 요구사항을 충족하기 위한 단계별 실행 전략을 정의합니다. Python, JavaScript, TypeScript를 포함한 멀티 언어 프로젝트에서 자동으로 언어를 감지하고, 적절한 CI/CD 워크플로우 템플릿을 적용하는 시스템을 구현합니다.

**핵심 목표**:
- LanguageDetector 클래스 확장 (패키지 매니저 감지 기능 추가)
- 언어별 GitHub Actions 워크플로우 템플릿 생성
- tdd-implementer 에이전트와의 통합
- 포괄적인 테스트 스위트 작성

---

## 구현 단계

### Step 1: 언어별 워크플로우 템플릿 생성

**우선순위**: 최우선 (High)

**목표**: Python, JavaScript, TypeScript, Go 언어에 대한 GitHub Actions 워크플로우 템플릿을 생성합니다.

**작업 내용**:

1. **템플릿 디렉토리 구조 생성**:
   ```
   src/moai_adk/templates/workflows/
   ├── python-tag-validation.yml
   ├── javascript-tag-validation.yml
   ├── typescript-tag-validation.yml
   └── go-tag-validation.yml
   ```

2. **Python 템플릿** (`python-tag-validation.yml`):
   - `actions/setup-python@v5` 사용
   - Python 3.11 지정
   - `uv` 패키지 매니저 활용
   - TAG 검증 테스트 실행

3. **JavaScript 템플릿** (`javascript-tag-validation.yml`):
   - `actions/setup-node@v4` 사용
   - Node.js 20.x 지정
   - npm/yarn/pnpm 선택 가능하도록 변수 지원
   - ESLint 및 TAG 검증 실행

4. **TypeScript 템플릿** (`typescript-tag-validation.yml`):
   - JavaScript 템플릿 기반
   - 타입 체크 단계 추가 (`npm run type-check`)
   - Vitest/Jest 테스트 실행

5. **Go 템플릿** (`go-tag-validation.yml`):
   - `actions/setup-go@v5` 사용
   - Go 1.22.x 지정
   - `go test` 및 TAG 검증 실행

**검증 기준**:
- ✅ 4개 언어 템플릿 파일이 정상적으로 생성됨
- ✅ 템플릿 구문 검증 통과 (YAML 유효성)
- ✅ 각 템플릿이 해당 언어의 표준 빌드 도구를 사용함

---

### Step 2: LanguageDetector 클래스 확장

**우선순위**: 최우선 (High)

**목표**: 패키지 매니저 감지 및 워크플로우 템플릿 경로 반환 기능을 추가합니다.

**작업 내용**:

1. **`detect_package_manager()` 메서드 구현**:
   ```python
   def detect_package_manager(self, project_root: Path) -> str:
       """
       JavaScript/TypeScript 프로젝트의 패키지 매니저를 감지합니다.

       우선순위:
       1. bun (bun.lockb 존재 시)
       2. pnpm (pnpm-lock.yaml 존재 시)
       3. yarn (yarn.lock 존재 시)
       4. npm (기본값)
       """
       if (project_root / "bun.lockb").exists():
           return "bun"
       if (project_root / "pnpm-lock.yaml").exists():
           return "pnpm"
       if (project_root / "yarn.lock").exists():
           return "yarn"
       return "npm"
   ```

2. **`get_workflow_template_path()` 메서드 구현**:
   ```python
   def get_workflow_template_path(self, language: str) -> Path:
       """
       감지된 언어에 대한 워크플로우 템플릿 경로를 반환합니다.

       Raises:
           WorkflowTemplateNotFoundError: 템플릿이 존재하지 않을 경우
       """
       template_map = {
           "python": "python-tag-validation.yml",
           "javascript": "javascript-tag-validation.yml",
           "typescript": "typescript-tag-validation.yml",
           "go": "go-tag-validation.yml",
       }

       template_name = template_map.get(language)
       if not template_name:
           raise WorkflowTemplateNotFoundError(f"No template for language: {language}")

       template_path = Path(__file__).parent / "templates" / "workflows" / template_name
       if not template_path.exists():
           raise WorkflowTemplateNotFoundError(f"Template not found: {template_path}")

       return template_path
   ```

3. **언어 감지 캐싱 구현**:
   ```python
   from functools import lru_cache

   @lru_cache(maxsize=32)
   def detect_language_cached(self, project_root: str) -> str:
       """캐싱된 언어 감지 (성능 최적화)"""
       return self.detect_language(Path(project_root))
   ```

**검증 기준**:
- ✅ `detect_package_manager()` 메서드가 4가지 패키지 매니저를 정확히 감지
- ✅ `get_workflow_template_path()` 메서드가 올바른 템플릿 경로 반환
- ✅ 존재하지 않는 언어 요청 시 명확한 예외 발생
- ✅ 캐싱 로직이 반복 호출 시 성능 향상 확인

---

### Step 3: tdd-implementer 에이전트 통합

**우선순위**: 최우선 (High)

**목표**: tdd-implementer 에이전트가 CI/CD 워크플로우 생성 시 언어 감지를 자동으로 수행하도록 통합합니다.

**작업 내용**:

1. **에이전트 지침 업데이트** (`.claude/agents/tdd-implementer.md`):
   ```markdown
   ## CI/CD 워크플로우 생성 프로세스

   1. **언어 감지**: `Skill("moai-alfred-language-detection")` 호출
   2. **템플릿 선택**: 감지된 언어에 따라 적절한 워크플로우 템플릿 선택
   3. **패키지 매니저 감지**: JavaScript/TypeScript인 경우 패키지 매니저 자동 감지
   4. **템플릿 렌더링**: 감지된 정보를 기반으로 워크플로우 파일 생성
   5. **검증**: 생성된 워크플로우 YAML 구문 검증
   ```

2. **워크플로우 생성 로직 구현**:
   ```python
   def generate_ci_workflow(project_root: Path) -> str:
       """CI/CD 워크플로우를 생성합니다."""
       detector = LanguageDetector()

       # Step 1: 언어 감지
       try:
           language = detector.detect_language(project_root)
           logger.info(f"Detected language: {language}")
       except LanguageDetectionError as e:
           logger.error(f"Language detection failed: {e}")
           raise

       # Step 2: 템플릿 경로 가져오기
       template_path = detector.get_workflow_template_path(language)

       # Step 3: 패키지 매니저 감지 (JavaScript/TypeScript만)
       template_vars = {}
       if language in ["javascript", "typescript"]:
           package_manager = detector.detect_package_manager(project_root)
           template_vars["package_manager"] = package_manager
           logger.info(f"Detected package manager: {package_manager}")

       # Step 4: 템플릿 렌더링
       with open(template_path, "r") as f:
           template_content = f.read()

       workflow_content = render_template(template_content, template_vars)

       # Step 5: 워크플로우 파일 작성
       workflow_path = project_root / ".github" / "workflows" / "tag-validation.yml"
       workflow_path.parent.mkdir(parents=True, exist_ok=True)
       workflow_path.write_text(workflow_content)

       logger.info(f"Generated workflow: {workflow_path}")
       return str(workflow_path)
   ```

**검증 기준**:
- ✅ tdd-implementer가 워크플로우 생성 전에 언어 감지를 수행
- ✅ 언어별로 적절한 템플릿이 선택됨
- ✅ 생성된 워크플로우 파일이 YAML 구문 검증 통과
- ✅ 언어 감지 결과가 로그에 기록됨

---

### Step 4: 포괄적인 테스트 스위트 작성

**우선순위**: 최우선 (High)

**목표**: 언어 감지, 패키지 매니저 감지, 템플릿 선택 로직에 대한 단위 테스트 및 통합 테스트를 작성합니다.

**작업 내용**:

1. **단위 테스트** (`tests/test_language_detection.py`):
   ```python
   def test_detect_python_project():
       """Python 프로젝트 감지 테스트"""
       with temp_project(files=["pyproject.toml"]) as project_root:
           detector = LanguageDetector()
           assert detector.detect_language(project_root) == "python"

   def test_detect_javascript_project():
       """JavaScript 프로젝트 감지 테스트"""
       with temp_project(files=["package.json"]) as project_root:
           detector = LanguageDetector()
           assert detector.detect_language(project_root) == "javascript"

   def test_detect_typescript_project():
       """TypeScript 프로젝트 감지 테스트"""
       with temp_project(files=["package.json", "tsconfig.json"]) as project_root:
           detector = LanguageDetector()
           assert detector.detect_language(project_root) == "typescript"

   def test_mixed_language_priority():
       """혼합 언어 프로젝트 우선순위 테스트"""
       with temp_project(files=["package.json", "tsconfig.json", "pyproject.toml"]) as project_root:
           detector = LanguageDetector()
           # TypeScript가 최우선
           assert detector.detect_language(project_root) == "typescript"

   def test_detect_yarn_package_manager():
       """Yarn 패키지 매니저 감지 테스트"""
       with temp_project(files=["package.json", "yarn.lock"]) as project_root:
           detector = LanguageDetector()
           assert detector.detect_package_manager(project_root) == "yarn"

   def test_language_detection_failure():
       """언어 감지 실패 테스트"""
       with temp_project(files=[]) as project_root:
           detector = LanguageDetector()
           with pytest.raises(LanguageDetectionError):
               detector.detect_language(project_root)
   ```

2. **통합 테스트** (`tests/integration/test_workflow_generation.py`):
   ```python
   def test_generate_python_workflow():
       """Python 워크플로우 생성 통합 테스트"""
       with temp_project(files=["pyproject.toml"]) as project_root:
           workflow_path = generate_ci_workflow(project_root)
           assert Path(workflow_path).exists()

           with open(workflow_path, "r") as f:
               workflow_content = f.read()
           assert "setup-python@v5" in workflow_content
           assert "uv" in workflow_content

   def test_generate_typescript_workflow():
       """TypeScript 워크플로우 생성 통합 테스트"""
       with temp_project(files=["package.json", "tsconfig.json", "yarn.lock"]) as project_root:
           workflow_path = generate_ci_workflow(project_root)
           assert Path(workflow_path).exists()

           with open(workflow_path, "r") as f:
               workflow_content = f.read()
           assert "setup-node@v4" in workflow_content
           assert "yarn" in workflow_content or "npm" in workflow_content
           assert "type-check" in workflow_content
   ```

**검증 기준**:
- ✅ 단위 테스트 커버리지 90% 이상
- ✅ 모든 지원 언어(Python, JS, TS, Go)에 대한 테스트 존재
- ✅ 에러 케이스(언어 감지 실패, 템플릿 없음) 테스트 포함
- ✅ CI/CD에서 모든 테스트가 통과

---

### Step 5: 문서 업데이트

**우선순위**: 중간 (Medium)

**목표**: 사용자 및 개발자를 위한 언어 감지 기능 문서를 작성합니다.

**작업 내용**:

1. **사용자 가이드 작성** (`.moai/docs/language-detection-guide.md`):
   - 지원 언어 목록
   - 설정 파일 우선순위 규칙
   - 수동 언어 지정 방법 (`.moai/config.json`)
   - 트러블슈팅 가이드

2. **tdd-implementer 에이전트 문서 업데이트**:
   - CI/CD 워크플로우 생성 프로세스 설명
   - 언어별 워크플로우 템플릿 커스터마이징 방법

3. **CHANGELOG 업데이트**:
   ```markdown
   ## [Unreleased]

   ### Added
   - 멀티 언어 프로젝트 지원: Python, JavaScript, TypeScript, Go, Ruby, PHP, Java, Rust, Kotlin
   - 언어별 GitHub Actions 워크플로우 템플릿 자동 적용
   - JavaScript/TypeScript 패키지 매니저 자동 감지 (npm, yarn, pnpm, bun)
   - 언어 감지 결과 캐싱으로 성능 최적화

   ### Changed
   - tdd-implementer 에이전트가 워크플로우 생성 전 언어 자동 감지 수행

   ### Fixed
   - JavaScript 프로젝트에 Python 워크플로우가 적용되던 문제 해결
   ```

**검증 기준**:
- ✅ 문서가 명확하고 이해하기 쉬움
- ✅ 모든 지원 언어에 대한 예시 포함
- ✅ 트러블슈팅 섹션이 일반적인 문제 해결

---

### Step 6: 실제 프로젝트 통합 테스트

**우선순위**: 중간 (Medium)

**목표**: 실제 프로젝트 템플릿(Python, Next.js, Node.js)에서 언어 감지 및 워크플로우 생성이 정상 동작하는지 검증합니다.

**작업 내용**:

1. **Python 프로젝트 템플릿 테스트**:
   - MoAI-ADK 자체 프로젝트에서 테스트
   - `pyproject.toml` 감지 확인
   - Python 워크플로우 템플릿 적용 확인

2. **Next.js/React 프로젝트 테스트**:
   - Next.js 스타터 템플릿 생성
   - `package.json` + `tsconfig.json` 감지 확인
   - TypeScript 워크플로우 템플릿 적용 확인

3. **Node.js/Express 프로젝트 테스트**:
   - Express API 템플릿 생성
   - JavaScript 감지 확인 (no `tsconfig.json`)
   - JavaScript 워크플로우 템플릿 적용 확인

4. **혼합 언어 모노레포 테스트**:
   - 루트에 `package.json` + `pyproject.toml` 존재
   - TypeScript 우선순위 확인
   - 서브디렉토리별 언어 감지 옵션 검토

**검증 기준**:
- ✅ 3가지 프로젝트 타입 모두에서 올바른 워크플로우 생성
- ✅ 생성된 워크플로우가 GitHub Actions에서 정상 실행
- ✅ 혼합 언어 프로젝트에서 우선순위 규칙이 정확히 적용

---

### Step 7: CI/CD 파이프라인 통합

**우선순위**: 보통 (Normal)

**목표**: MoAI-ADK 자체 CI/CD 파이프라인에 언어 감지 테스트를 추가합니다.

**작업 내용**:

1. **테스트 매트릭스 추가** (`.github/workflows/test.yml`):
   ```yaml
   strategy:
     matrix:
       language: [python, javascript, typescript, go]
       include:
         - language: python
           test-project: tests/fixtures/python-project
         - language: javascript
           test-project: tests/fixtures/javascript-project
         - language: typescript
           test-project: tests/fixtures/typescript-project
         - language: go
           test-project: tests/fixtures/go-project

   steps:
     - name: Test language detection for ${{ matrix.language }}
       run: pytest tests/test_language_detection.py -k ${{ matrix.language }}
   ```

2. **하위 호환성 검증**:
   - 기존 Python 전용 워크플로우가 여전히 동작하는지 확인
   - 레거시 프로젝트에 대한 영향 없음을 보장

**검증 기준**:
- ✅ CI/CD 파이프라인이 모든 언어에 대해 테스트 실행
- ✅ 테스트 실패 시 명확한 에러 메시지 출력
- ✅ 하위 호환성 유지 확인

---

## 기술 스택 및 라이브러리

### 언어 및 프레임워크

- **Python**: 3.11+ (MoAI-ADK 메인 언어)
- **JavaScript/TypeScript**: Node.js 20.x
- **Go**: 1.22.x

### 핵심 라이브러리

#### Python

- `pathlib`: 파일 시스템 경로 처리
- `functools.lru_cache`: 언어 감지 결과 캐싱
- `pytest`: 단위 테스트 프레임워크
- `pyyaml`: YAML 구문 검증

#### JavaScript/TypeScript (테스트용)

- `vitest`: 단위 테스트 프레임워크
- `yaml`: YAML 파싱 및 검증

#### CI/CD

- **GitHub Actions**: v4
  - `actions/setup-python@v5`
  - `actions/setup-node@v4`
  - `actions/setup-go@v5`
  - `actions/checkout@v4`

### 템플릿 엔진

- **Jinja2** (Python): 워크플로우 템플릿 렌더링
  - 버전: `>=3.1.2`
  - 용도: 패키지 매니저 변수 주입, 조건부 단계 추가

---

## 테스트 전략

### 단위 테스트 (Unit Tests)

- **대상**: `LanguageDetector` 클래스 메서드
- **도구**: `pytest`
- **커버리지 목표**: 90% 이상
- **테스트 케이스**:
  - 각 언어별 감지 (Python, JS, TS, Go, Ruby, etc.)
  - 패키지 매니저 감지 (npm, yarn, pnpm, bun)
  - 템플릿 경로 반환
  - 에러 케이스 (언어 감지 실패, 템플릿 없음)
  - 캐싱 동작 확인

### 통합 테스트 (Integration Tests)

- **대상**: `generate_ci_workflow()` 함수 및 tdd-implementer 통합
- **도구**: `pytest` + 임시 프로젝트 픽스처
- **테스트 케이스**:
  - 언어별 워크플로우 생성 (Python, JS, TS, Go)
  - 생성된 워크플로우 YAML 구문 검증
  - 혼합 언어 프로젝트 우선순위 처리
  - 패키지 매니저 감지 및 템플릿 주입

### E2E 테스트 (End-to-End Tests)

- **대상**: 실제 프로젝트 템플릿에서 전체 워크플로우 실행
- **도구**: GitHub Actions (로컬에서는 `act` 사용)
- **테스트 케이스**:
  - Python 프로젝트: `pytest` 실행
  - JavaScript 프로젝트: `npm test` 실행
  - TypeScript 프로젝트: `npm run type-check` + `npm test` 실행
  - 워크플로우 실패 시 명확한 에러 메시지 출력

### 회귀 테스트 (Regression Tests)

- **대상**: 기존 Python 전용 워크플로우
- **목표**: 하위 호환성 유지
- **검증 항목**:
  - 레거시 프로젝트에서 Python 워크플로우 정상 생성
  - 새로운 언어 감지 로직이 기존 동작에 영향 없음

---

## 배포 및 통합 계획

### Phase 1: 개발 환경 검증

1. 로컬에서 모든 단위 테스트 및 통합 테스트 실행
2. 테스트 커버리지 90% 이상 확인
3. 코드 리뷰 및 TRUST 5 원칙 검증

### Phase 2: Staging 환경 배포

1. `develop` 브랜치에 PR 생성
2. GitHub Actions에서 자동 테스트 실행
3. Python, JavaScript, TypeScript 프로젝트 E2E 테스트 통과 확인

### Phase 3: Production 배포

1. `main` 브랜치로 머지
2. MoAI-ADK v0.8.0 릴리스 (CHANGELOG 업데이트)
3. 사용자 문서 배포 (`.moai/docs/language-detection-guide.md`)

### Phase 4: 모니터링 및 피드백

1. 실제 사용자 프로젝트에서 언어 감지 성공률 모니터링
2. 언어 감지 실패 케이스 수집 및 분석
3. 필요 시 추가 언어 지원 (Rust, Kotlin, etc.)

---

## 리스크 및 대응 계획

### Risk 1: 복잡한 모노레포 구조

**상황**: 루트와 서브디렉토리에 서로 다른 언어 설정 파일이 존재하는 경우

**대응**:
- 현재 우선순위 규칙 적용 (TypeScript > JavaScript > Python)
- 향후 서브디렉토리별 언어 감지 옵션 추가 검토

### Risk 2: 새로운 패키지 매니저 등장

**상황**: Deno, Bun 등 새로운 JavaScript 런타임/패키지 매니저 등장

**대응**:
- 패키지 매니저 감지 로직을 확장 가능하게 설계
- 설정 파일 기반 매핑 테이블 유지

### Risk 3: 언어 감지 성능 저하

**상황**: 대규모 프로젝트에서 파일 시스템 탐색 시간 증가

**대응**:
- `lru_cache` 데코레이터를 통한 결과 캐싱
- 필요 시 설정 파일 위치 힌트 제공 기능 추가

### Risk 4: 하위 호환성 문제

**상황**: 기존 Python 전용 워크플로우를 사용하는 프로젝트에 영향

**대응**:
- 회귀 테스트 작성 (기존 동작 유지 확인)
- 언어 감지 실패 시 Python을 기본값으로 fallback (선택적)

---

## 성공 기준

### 기능 요구사항

- ✅ Python, JavaScript, TypeScript, Go 언어 자동 감지
- ✅ 언어별 GitHub Actions 워크플로우 템플릿 자동 적용
- ✅ JavaScript/TypeScript 패키지 매니저 자동 감지 (npm, yarn, pnpm, bun)
- ✅ 언어 감지 실패 시 명확한 에러 메시지

### 품질 요구사항

- ✅ 단위 테스트 커버리지 90% 이상
- ✅ 통합 테스트 100% 통과
- ✅ E2E 테스트 (실제 프로젝트) 100% 통과
- ✅ 코드 리뷰 및 TRUST 5 원칙 준수

### 성능 요구사항

- ✅ 언어 감지 시간 < 100ms
- ✅ 캐시 히트 시 < 10ms
- ✅ 워크플로우 생성 시간 < 500ms

### 문서 요구사항

- ✅ 사용자 가이드 작성 (`.moai/docs/language-detection-guide.md`)
- ✅ tdd-implementer 에이전트 문서 업데이트
- ✅ CHANGELOG 업데이트

---

**Generated with**: 🎩 Alfred (MoAI-ADK v0.7.0)
**Implementation Plan for**: SPEC-LANGUAGE-DETECTION-001

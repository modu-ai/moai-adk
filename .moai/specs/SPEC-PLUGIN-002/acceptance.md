# SPEC-PLUGIN-002: 수락 기준

## 📋 개요

본 문서는 `/alfred:8-project` 커맨드 강화 기능의 상세한 수락 기준을 정의합니다.

---

## ✅ Given-When-Then 시나리오

### 시나리오 1: 신규 프로젝트 초기화 (정상 케이스)

**Given**:
- 사용자가 빈 디렉토리에서 작업 중
- `.moai/` 디렉토리가 존재하지 않음
- `${CLAUDE_PLUGIN_ROOT}` 환경변수가 설정되어 있음
- Node.js가 설치되어 있음

**When**:
- 사용자가 `/alfred:8-project` 커맨드를 실행
- Phase 1 분석 후 "진행" 응답
- 변수 입력:
  - PROJECT_NAME: "my-awesome-project"
  - PROJECT_MODE: "team"
  - PROJECT_DESCRIPTION: "AI-powered development framework"
  - LOCALE: "ko"
- Git 초기화: "예"

**Then**:
- `.moai/` 디렉토리 생성 확인
- `.moai/config.json` 생성 확인 (렌더링 완료)
- `.moai/project/product.md`, `structure.md`, `tech.md` 생성 확인
- `config.json` 내용:
  ```json
  {
    "project": {
      "name": "my-awesome-project",
      "mode": "team",
      "description": "AI-powered development framework",
      "locale": "ko"
    }
  }
  ```
- `.git/` 디렉토리 생성 확인 (git init 실행됨)
- `.gitignore` 파일 생성 확인
- 완료 메시지 및 다음 단계 안내 표시

---

### 시나리오 2: 기존 프로젝트 갱신 (보존 모드)

**Given**:
- `.moai/` 디렉토리가 이미 존재
- `.moai/project/product.md`에 기존 내용이 있음
- `.moai/config.json`이 있지만 구버전 형식

**When**:
- 사용자가 `/alfred:8-project` 커맨드를 실행
- Phase 1 분석 결과: "갱신 모드"
- "진행" 응답

**Then**:
- 기존 `product.md` 백업 생성 (예: `product.md.backup-20251010-143022`)
- 새 `product.md` 생성 (최신 템플릿 적용)
- 기존 `config.json` 보존 또는 덮어쓰기 확인 프롬프트 표시
- ⚠️ WARNING 메시지: ".moai/ 디렉토리가 이미 존재합니다 → 갱신 모드로 전환합니다"
- 완료 메시지: "기존 파일은 백업되었습니다"

---

### 시나리오 3: Node.js 미설치 (대체 전략)

**Given**:
- Node.js가 설치되지 않음
- 템플릿 디렉토리에 `config.json.mustache` 존재

**When**:
- 사용자가 `/alfred:8-project` 커맨드를 실행
- Phase 1 분석 결과: "Node.js 미설치"
- "진행" 응답

**Then**:
- ⚠️ WARNING 메시지: "Node.js가 설치되지 않았습니다 → Mustache 렌더링을 건너뜁니다"
- 기본값으로 `config.json` 생성:
  ```json
  {
    "project": {
      "name": "current-directory-name",
      "mode": "personal",
      "description": "",
      "locale": "ko"
    }
  }
  ```
- ℹ️ INFO 메시지: "수동 렌더링 가이드를 참조하세요"
- 완료 메시지에 Node.js 설치 권장 포함

---

### 시나리오 4: 템플릿 파일 누락 (에러 케이스)

**Given**:
- `${CLAUDE_PLUGIN_ROOT}/templates/.moai/config.json.mustache` 파일이 없음

**When**:
- 사용자가 `/alfred:8-project` 커맨드를 실행
- Phase 1 분석 수행

**Then**:
- ❌ CRITICAL 메시지: "필수 템플릿 파일이 없습니다"
- 에러 메시지:
  ```
  ❌ CRITICAL: 필수 템플릿 파일이 없습니다
    → ${CLAUDE_PLUGIN_ROOT}/templates/.moai/config.json.mustache
    → 해결: 플러그인을 재설치하거나 템플릿 파일을 수동으로 복사하세요
  ```
- 초기화 중단
- 사용자에게 플러그인 재설치 안내

---

### 시나리오 5: 특수문자 프로젝트명 정규화

**Given**:
- 사용자가 프로젝트명으로 "My Project (2025)!" 입력

**When**:
- 변수 수집 단계에서 PROJECT_NAME 입력

**Then**:
- ⚠️ WARNING 메시지: "프로젝트명에 특수문자가 포함되어 있습니다"
- 정규화 제안:
  ```
  입력: "My Project (2025)!"
  정규화: "my-project-2025"
  계속 진행하시겠습니까? (y/n)
  ```
- "y" 응답 시 정규화된 이름으로 진행
- "n" 응답 시 재입력 요청

---

### 시나리오 6: Git 초기화 거부

**Given**:
- Git이 설치되어 있음
- 사용자가 Git 초기화를 원하지 않음

**When**:
- 사용자가 `/alfred:8-project` 커맨드를 실행
- Git 초기화 프롬프트에서 "아니오" 응답

**Then**:
- `.git/` 디렉토리 생성되지 않음
- `.gitignore` 파일 생성되지 않음
- ℹ️ INFO 메시지: "Git 초기화를 건너뜁니다"
- 완료 메시지: "나중에 `git init`으로 초기화할 수 있습니다"

---

### 시나리오 7: ${CLAUDE_PLUGIN_ROOT} 미설정 (에러 케이스)

**Given**:
- `${CLAUDE_PLUGIN_ROOT}` 환경변수가 설정되지 않음

**When**:
- 사용자가 `/alfred:8-project` 커맨드를 실행
- Phase 1 환경 분석 수행

**Then**:
- ❌ CRITICAL 메시지: "플러그인 루트 경로를 찾을 수 없습니다"
- 에러 메시지:
  ```
  ❌ CRITICAL: 플러그인 루트 경로를 찾을 수 없습니다
    → ${CLAUDE_PLUGIN_ROOT} 환경변수가 설정되지 않았습니다
    → 해결: Claude Code 플러그인 설치 상태를 확인하세요
  ```
- 초기화 중단
- 사용자에게 플러그인 설치 안내

---

### 시나리오 8: 권한 에러 (.moai/ 생성 실패)

**Given**:
- 현재 디렉토리에 쓰기 권한이 없음

**When**:
- 사용자가 `/alfred:8-project` 커맨드를 실행
- `mkdir -p .moai/` 명령 실행

**Then**:
- ❌ CRITICAL 메시지: ".moai/ 디렉토리 생성 실패: 권한 거부"
- 에러 메시지:
  ```
  ❌ CRITICAL: .moai/ 디렉토리 생성 실패
    → 권한 거부 (Permission Denied)
    → 해결: chmod 755 . 또는 다른 디렉토리에서 실행하세요
  ```
- 초기화 중단
- 권한 확인 명령어 안내: `ls -ld .`

---

### 시나리오 9: render-template.cjs 실행 실패

**Given**:
- Node.js가 설치되어 있음
- `scripts/render-template.cjs`가 존재
- mustache npm 패키지가 설치되지 않음

**When**:
- Mustache 렌더링 단계 실행
- `node scripts/render-template.cjs` 실행

**Then**:
- ❌ CRITICAL 메시지: "템플릿 렌더링 실패"
- 에러 메시지:
  ```
  ❌ CRITICAL: 템플릿 렌더링 실패
    → Error: Cannot find module 'mustache'
    → 해결:
      1. cd ${CLAUDE_PLUGIN_ROOT}
      2. npm install mustache
      3. /alfred:8-project 재실행
  ```
- 대체 전략: 기본값으로 config.json 생성
- 수동 렌더링 가이드 제공

---

### 시나리오 10: LOCALE 검증 (엣지 케이스)

**Given**:
- 사용자가 LOCALE으로 "fr" (프랑스어) 입력

**When**:
- 변수 수집 단계에서 LOCALE 입력

**Then**:
- ⚠️ WARNING 메시지: "지원하지 않는 LOCALE 값입니다"
- 기본값 사용:
  ```
  ⚠️ WARNING: 지원하지 않는 LOCALE 값입니다
    → 입력: "fr"
    → 지원 언어: ko, en, ja, zh
    → 기본값 "ko"를 사용합니다
  ```
- config.json에 `"locale": "ko"` 저장
- 완료 메시지: "LOCALE을 나중에 .moai/config.json에서 수정할 수 있습니다"

---

## 🧪 품질 게이트 기준

### 1. 기능 완전성

- [ ] 모든 템플릿 파일이 정상적으로 복사됨
- [ ] Mustache 렌더링이 성공적으로 완료됨
- [ ] 변수 치환이 정확하게 수행됨
- [ ] Git 초기화가 선택적으로 동작함
- [ ] 에러 케이스가 적절히 처리됨

### 2. 성능 기준

- [ ] 전체 초기화 시간 ≤5초
- [ ] 템플릿 복사 시간 ≤2초
- [ ] Mustache 렌더링 시간 ≤1초

### 3. 에러 메시지 품질

- [ ] 모든 에러에 심각도 아이콘 표시 (❌ CRITICAL, ⚠️ WARNING, ℹ️ INFO)
- [ ] 에러 메시지에 컨텍스트와 해결 방법 포함
- [ ] 사용자 친화적인 언어 사용

### 4. 추적성

- [ ] 모든 생성 파일에 @TAG 주석 포함
- [ ] render-template.cjs에 `@CODE:PLUGIN-002:SCRIPT` TAG 포함
- [ ] 8-project.md에 `@CODE:PLUGIN-002` TAG 포함

### 5. 보안

- [ ] 특수문자 입력 시 정규화 수행
- [ ] 경로 인젝션 방지 (상대 경로 검증)
- [ ] 민감 정보 로그 출력 금지

---

## 📊 검증 방법 및 도구

### 자동화 테스트

**테스트 파일**: `tests/commands/project-init.test.ts`

**테스트 프레임워크**: Vitest (TypeScript) 또는 pytest (Python)

**테스트 케이스**:
```typescript
describe('@TEST:PLUGIN-002 - /alfred:8-project 커맨드', () => {
  it('신규 프로젝트 초기화 성공', async () => {
    // Given: 빈 디렉토리
    // When: 커맨드 실행
    // Then: .moai/ 디렉토리 및 config.json 생성 확인
  });

  it('Mustache 렌더링 정확성', async () => {
    // Given: config.json.mustache 템플릿
    // When: 변수 매핑 및 렌더링
    // Then: 변수 치환 확인
  });

  it('특수문자 프로젝트명 정규화', async () => {
    // Given: "My Project (2025)!" 입력
    // When: 정규화 수행
    // Then: "my-project-2025" 출력 확인
  });

  it('Node.js 미설치 시 대체 전략', async () => {
    // Given: Node.js 미설치 환경 모의
    // When: 커맨드 실행
    // Then: 기본값으로 config.json 생성 확인
  });

  it('템플릿 파일 누락 에러 처리', async () => {
    // Given: config.json.mustache 삭제
    // When: 커맨드 실행
    // Then: ❌ CRITICAL 에러 표시 확인
  });
});
```

### 수동 검증

**체크리스트**:
- [ ] 신규 프로젝트에서 커맨드 실행 → 모든 파일 생성 확인
- [ ] 기존 프로젝트에서 커맨드 실행 → 백업 및 보존 확인
- [ ] Node.js 미설치 환경에서 실행 → 대체 전략 동작 확인
- [ ] 특수문자 입력 → 정규화 프롬프트 표시 확인
- [ ] Git 초기화 거부 → .git/ 디렉토리 미생성 확인

---

## 🎯 완료 조건 (Definition of Done)

### 코드 완료

- [ ] render-template.cjs 스크립트 작성 완료
- [ ] 8-project.md 커맨드 로직 강화 완료
- [ ] 모든 에러 시나리오 처리 코드 작성 완료

### 테스트 완료

- [ ] 유닛 테스트 커버리지 ≥85%
- [ ] 통합 테스트 통과 (10개 시나리오 모두)
- [ ] 수동 검증 체크리스트 완료

### 문서 완료

- [ ] plan.md, acceptance.md 작성 완료
- [ ] render-template.cjs 스크립트 주석 추가
- [ ] 8-project.md에 사용 예시 추가

### 품질 검증

- [ ] TRUST 5원칙 준수 확인:
  - **Test First**: 테스트 커버리지 ≥85%
  - **Readable**: 함수 ≤50 LOC, 복잡도 ≤10
  - **Unified**: config.json 형식 일관성
  - **Secured**: 입력 검증 및 정규화
  - **Trackable**: @TAG 체인 완전성

### Git 워크플로우

- [ ] 브랜치: `feature/SPEC-PLUGIN-002` 생성
- [ ] 커밋:
  - 🔴 RED: render-template.cjs 테스트 작성
  - 🟢 GREEN: render-template.cjs 구현 완료
  - ♻️ REFACTOR: 8-project.md 로직 강화
  - 📝 DOCS: plan.md, acceptance.md 작성

### 동기화

- [ ] `/alfred:3-sync` 실행
- [ ] TAG 체인 무결성 검증
- [ ] Living Document 생성 확인

---

_이 문서는 SPEC-PLUGIN-002의 수락 기준으로, `/alfred:2-build SPEC-PLUGIN-002` 완료 후 검증에 사용됩니다._

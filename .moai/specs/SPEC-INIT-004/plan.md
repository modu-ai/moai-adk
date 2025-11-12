# @SPEC:INIT-004 Implementation Plan

## @PLAN:INIT-004 구현 계획

---

## 1. 개요

**목표**: `moai-adk init` 명령어의 CLAUDE.md 의존성 제거 및 Alfred 커맨드 파일 생성 기능 수정

**접근 방식**: 3단계 TDD 구현 (RED → GREEN → REFACTOR)

---

## 2. 우선순위별 마일스톤

### 🎯 1차 목표: 테스트 케이스 작성 (RED)

**범위**:
- CLAUDE.md 부재 시 자동 생성 테스트
- Alfred 커맨드 파일 생성 테스트
- 검증 로직 테스트

**산출물**:
- `tests/unit/installer.test.ts` (신규 또는 확장)
- 테스트 실패 확인 (현재 코드는 테스트 통과 불가)

### 🎯 2차 목표: 최소 기능 구현 (GREEN)

**범위**:
- `ensureClaudeMd()` 함수 구현
- `copyAlfredCommands()` 함수 구현
- `verifyInstallation()` 함수 구현
- `installer.ts` 메인 흐름 수정

**산출물**:
- 수정된 `installer.ts`
- 모든 테스트 통과

### 🎯 3차 목표: 코드 품질 개선 (REFACTOR)

**범위**:
- 에러 핸들링 일관성 개선
- 중복 코드 제거
- 타입 안정성 강화
- 로깅 메시지 표준화

**산출물**:
- 리팩토링된 코드 (기능 동일, 품질 향상)
- TRUST 5원칙 준수 확인

---

## 3. 기술적 접근 방법

### 3.1 아키텍처 설계

#### 현재 구조
```
moai-adk-ts/src/core/installer/
├─ installer.ts          # 메인 초기화 로직
├─ template-processor.ts # 템플릿 병합 로직
└─ template-security.ts  # 보안 검증 로직
```

#### 수정 대상
- **installer.ts**: 3개 함수 추가 + 메인 흐름 수정

#### 추가 함수 시그니처
```typescript
// CLAUDE.md 자동 생성
async function ensureClaudeMd(): Promise<void>

// Alfred 커맨드 복사
async function copyAlfredCommands(): Promise<void>

// 설치 검증
async function verifyInstallation(): Promise<void>
```

### 3.2 데이터 흐름

```
사용자: moai-adk init
    ↓
[1] ensureClaudeMd()
    - CLAUDE.md 존재 확인
    - 부재 시 템플릿에서 복사
    ↓
[2] 기존 초기화 로직
    - .moai/ 디렉토리 생성
    - config.json 생성
    - development-guide.md 복사
    ↓
[3] copyAlfredCommands()
    - .claude/commands/alfred/ 생성
    - 4개 커맨드 파일 복사
    ↓
[4] verifyInstallation()
    - 필수 파일 존재 확인
    - 누락 시 에러 발생
    ↓
[SUCCESS] 초기화 완료
```

### 3.3 에러 처리 전략

#### 에러 타입 정의
```typescript
class InstallerError extends Error {
  constructor(
    public readonly type: 'TEMPLATE_MISSING' | 'PERMISSION_DENIED' | 'VERIFICATION_FAILED',
    message: string,
    public readonly details?: string[]
  ) {
    super(message);
    this.name = 'InstallerError';
  }
}
```

#### 에러 복구 전략
| 에러        | 복구 가능 여부 | 대응                                |
| ----------- | -------------- | ----------------------------------- |
| 템플릿 누락 | ❌ 불가능       | 즉시 중단, 패키지 재설치 권장       |
| 권한 거부   | ❌ 불가능       | 즉시 중단, sudo 또는 권한 변경 권장 |
| 검증 실패   | ❌ 불가능       | 즉시 중단, 누락 파일 목록 표시      |
| 부분 설치   | ✅ 가능         | 정리 후 재시도 권장                 |

### 3.4 테스트 전략

#### 단위 테스트 (Unit Tests)
```typescript
describe('ensureClaudeMd', () => {
  it('CLAUDE.md 없으면 템플릿에서 생성', async () => {
    // Given: CLAUDE.md 부재
    // When: ensureClaudeMd() 실행
    // Then: 템플릿에서 복사됨
  });

  it('CLAUDE.md 있으면 보존', async () => {
    // Given: 기존 CLAUDE.md 존재
    // When: ensureClaudeMd() 실행
    // Then: 기존 파일 유지
  });
});

describe('copyAlfredCommands', () => {
  it('4개 커맨드 파일 모두 생성', async () => {
    // Given: 신규 프로젝트
    // When: copyAlfredCommands() 실행
    // Then: 0-project.md, 1-spec.md, 2-build.md, 3-sync.md 생성
  });

  it('템플릿 파일 누락 시 에러', async () => {
    // Given: 템플릿 파일 누락
    // When: copyAlfredCommands() 실행
    // Then: InstallerError 발생
  });
});

describe('verifyInstallation', () => {
  it('모든 필수 파일 존재 시 성공', async () => {
    // Given: 모든 필수 파일 존재
    // When: verifyInstallation() 실행
    // Then: 에러 없이 완료
  });

  it('필수 파일 누락 시 에러', async () => {
    // Given: 1개 이상 필수 파일 누락
    // When: verifyInstallation() 실행
    // Then: InstallerError 발생, 누락 목록 표시
  });
});
```

#### 통합 테스트 (Integration Tests)
```typescript
describe('moai-adk init (E2E)', () => {
  it('신규 설치 시 모든 파일 생성', async () => {
    // Given: 빈 프로젝트
    // When: moai-adk init 실행
    // Then: CLAUDE.md 포함 모든 파일 생성
  });

  it('기존 프로젝트 재초기화 시 보존', async () => {
    // Given: 이미 초기화된 프로젝트
    // When: moai-adk init 재실행
    // Then: 사용자 설정 보존, 누락 파일만 추가
  });
});
```

---

## 4. 리스크 및 대응 방안

### 리스크 1: 템플릿 파일 패키징 누락

**확률**: Medium
**영향도**: High
**대응**:
- 패키지 빌드 시 템플릿 디렉토리 포함 여부 확인
- `package.json`의 `files` 필드 검증
- CI/CD 파이프라인에 템플릿 검증 추가

### 리스크 2: 사용자 커스터마이징 덮어쓰기

**확률**: Low
**영향도**: High
**대응**:
- CLAUDE.md 존재 시 보존 (절대 덮어쓰지 않음)
- `.claude/commands/alfred/` 파일도 선택적 덮어쓰기 옵션 제공
- 경고 메시지 표시

### 리스크 3: 파일 시스템 권한 문제

**확률**: Medium
**영향도**: Medium
**대응**:
- try-catch로 권한 오류 명확히 처리
- sudo 권장 메시지 표시
- 권한 확인 사전 검증 추가 고려

---

## 5. 구현 순서 (TDD 기반)

### Step 1: RED (테스트 작성)
1. `tests/unit/installer.test.ts` 작성
2. 테스트 실행 → 모두 실패 확인
3. 커밋: `🧪 TEST: SPEC-INIT-004 RED - 초기화 로직 테스트 작성`

### Step 2: GREEN (최소 구현)
1. `ensureClaudeMd()` 구현
2. `copyAlfredCommands()` 구현
3. `verifyInstallation()` 구현
4. `installer.ts` 메인 흐름 수정
5. 테스트 실행 → 모두 통과 확인
6. 커밋: `✨ FEATURE: SPEC-INIT-004 GREEN - CLAUDE.md 의존성 제거 및 커맨드 생성`

### Step 3: REFACTOR (품질 개선)
1. 중복 코드 제거
2. 타입 안정성 강화
3. 에러 메시지 표준화
4. TRUST 5원칙 검증
5. 커밋: `♻️ REFACTOR: SPEC-INIT-004 - 코드 품질 개선`

---

## 6. 종속성 분석

### 선행 완료 필요
- ✅ SPEC-INSTALLER-SEC-001 (템플릿 보안 검증) - 이미 완료
- ✅ SPEC-TEMPLATE-001 (Template Processor) - 이미 완료

### 병렬 작업 가능
- 없음 (단일 SPEC 구현)

### 후속 작업 차단
- SPEC-INSTALLER-ROLLBACK-001 (롤백 기능)
- SPEC-INSTALLER-UPDATE-001 (템플릿 업데이트)

---

## 7. 완료 조건 (Definition of Done)

### 기능 완료
- ✅ CLAUDE.md 없이 `moai-adk init` 성공
- ✅ 4개 Alfred 커맨드 파일 자동 생성
- ✅ 초기화 완료 후 검증 통과

### 품질 게이트
- ✅ 테스트 커버리지 85% 이상
- ✅ 모든 단위 테스트 통과
- ✅ 통합 테스트 통과
- ✅ TRUST 5원칙 준수:
  - **Test First**: 테스트 선작성 완료
  - **Readable**: ESLint/Prettier 통과
  - **Unified**: 함수 ≤50 LOC
  - **Secured**: 보안 이슈 없음
  - **Trackable**: @TAG 체인 완전

### 문서화
- ✅ CHANGELOG.md 업데이트
- ✅ README.md 설치 가이드 수정
- ✅ API 문서 업데이트 (필요 시)

### Git 워크플로우
- ✅ `feature/SPEC-INIT-004` 브랜치 생성
- ✅ 3회 커밋 (RED/GREEN/REFACTOR)
- ✅ Draft PR 생성
- ✅ CI/CD 통과
- ✅ PR 머지 준비

---

## 8. 다음 단계 안내

### 구현 시작
```bash
/alfred:2-run SPEC-INIT-004
```

### 문서 동기화
```bash
/alfred:3-sync SPEC-INIT-004
```

---

_이 계획은 `/alfred:2-run SPEC-INIT-004` 실행 시 참조됩니다._
